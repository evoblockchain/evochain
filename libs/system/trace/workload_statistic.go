package trace

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/evoblockchain/evochain/libs/tendermint/libs/log"
)

var (
	startupTime = time.Now()

	applyBlockWorkloadStatistic = newWorkloadStatistic(
		[]time.Duration{time.Hour, 2 * time.Hour, 4 * time.Hour, 8 * time.Hour}, []string{LastRun, Persist})
)

// TODO: think about a very long work which longer than a statistic period.

// WorkloadStatistic accumulate workload for specific trace tags during some specific period.
// Everytime `Add` or `end` method be called, it record workload on corresponding `summaries` fields,
// and send this workload info to `shrinkLoop`, which will subtract this workload from `summaries`
// when the workload out of statistic period. To do that, `shrinkLoop` will record the workload and it's
// out-of-date timestamp; `shrinkLoop` also has a ticker promote current time once a second.
// If current time is larger or equal than recorded timestamp, it remove that workload and subtract
// it's value from `summaries`.
//
// NOTE: CAN NOT use `WorkloadStatistic` concurrently for those reasons:
// read/write almost all fields
// workload summary may be wrong if a work is still running(latestBegin.IsZero isn't true)
//
// calling sequence:
// 1. newWorkloadStatistic
// 2. calling begin/end before and after doing some work
// 3. calling summary to get a summary statistic
type WorkloadStatistic struct {
	concernedTags map[string]struct{}
	summaries     []workloadSummary

	latestTag   string
	latestBegin time.Time
	logger      log.Logger

	workCh chan singleWorkInfo
}

type workloadSummary struct {
	period   time.Duration
	workload int64
}

type singleWorkInfo struct {
	duration int64
	endTime  time.Time
}

func GetApplyBlockWorkloadSttistic() *WorkloadStatistic {
	return applyBlockWorkloadStatistic
}

func newWorkloadStatistic(periods []time.Duration, tags []string) *WorkloadStatistic {
	concernedTags := toTagsMap(tags)

	workloads := make([]workloadSummary, 0, len(periods))
	for _, period := range periods {
		workloads = append(workloads, workloadSummary{period, 0})
	}

	wls := &WorkloadStatistic{concernedTags: concernedTags, summaries: workloads, workCh: make(chan singleWorkInfo, 1000)}
	go wls.shrinkLoop()

	return wls
}

func (ws *WorkloadStatistic) SetLogger(logger log.Logger) {
	ws.logger = logger
}

func (ws *WorkloadStatistic) Add(tag string, dur time.Duration) {
	if _, ok := ws.concernedTags[tag]; !ok {
		return
	}

	now := time.Now()
	for i := range ws.summaries {
		atomic.AddInt64(&ws.summaries[i].workload, int64(dur))
	}

	ws.workCh <- singleWorkInfo{int64(dur), now}
}

func (ws *WorkloadStatistic) Format() string {
	var sumItem []string
	for _, summary := range ws.summary() {
		sumItem = append(sumItem, fmt.Sprintf("%.2f", float64(summary.workload)/float64(summary.period)))
	}

	return strings.Join(sumItem, "|")
}

func (ws *WorkloadStatistic) begin(tag string, t time.Time) {
	if _, ok := ws.concernedTags[tag]; !ok {
		return
	}

	ws.latestTag = tag
	ws.latestBegin = t
}

func (ws *WorkloadStatistic) end(tag string, t time.Time) {
	if _, ok := ws.concernedTags[tag]; !ok {
		return
	}

	if ws.latestTag != tag {
		ws.logger.Error("WorkloadStatistic", ": begin tag", ws.latestTag, "end tag", tag)
		return
	}
	if ws.latestBegin.IsZero() {
		ws.logger.Error("WorkloadStatistic", "begin is not called before end")
		return
	}

	dur := t.Sub(ws.latestBegin)
	for i := range ws.summaries {
		atomic.AddInt64(&ws.summaries[i].workload, int64(dur))
	}

	ws.workCh <- singleWorkInfo{int64(dur), t}
	ws.latestBegin = time.Time{}
}

type summaryInfo struct {
	period   time.Duration
	workload time.Duration
}

func (ws *WorkloadStatistic) summary() []summaryInfo {
	if !ws.latestBegin.IsZero() {
		ws.logger.Error("WorkloadStatistic", ": some work is still running when calling summary")
		return nil
	}

	startupDuration := time.Now().Sub(startupTime)
	result := make([]summaryInfo, 0, len(ws.summaries))

	for _, summary := range ws.summaries {
		period := minDuration(startupDuration, summary.period)
		result = append(result, summaryInfo{period, time.Duration(atomic.LoadInt64(&summary.workload))})
	}
	return result
}

func (ws *WorkloadStatistic) shrinkLoop() {
	shrinkInfos := make([]map[int64]int64, 0, len(ws.summaries))
	for i := 0; i < len(ws.summaries); i++ {
		shrinkInfos = append(shrinkInfos, make(map[int64]int64))
	}

	var latest int64
	ticker := time.NewTicker(time.Second)

	for {
		select {
		case singleWork := <-ws.workCh:
			// `earliest` record the expired timestamp which is minimum.
			// It's just for initialize `latest`.
			earliest := int64(^uint64(0) >> 1)

			for sumIndex, summary := range ws.summaries {
				expiredTS := singleWork.endTime.Add(summary.period).Unix()
				if expiredTS < earliest {
					earliest = expiredTS
				}

				info := shrinkInfos[sumIndex]
				// TODO: it makes recoding workload larger than actual value
				//       if a work begin before this period and end during this period
				if _, ok := info[expiredTS]; !ok {
					info[expiredTS] = singleWork.duration
				} else {
					info[expiredTS] += singleWork.duration
				}
			}

			if latest == 0 {
				latest = earliest
			}
		case t := <-ticker.C:
			current := t.Unix()
			if latest == 0 {
				latest = current
			}

			// try to remove workload of every expired work.
			// `latest` make sure even if ticker is not accurately,
			// we can also remove the expired correctly.
			for index, info := range shrinkInfos {
				for i := latest; i < current+1; i++ {
					w, ok := info[i]
					if ok {
						atomic.AddInt64(&ws.summaries[index].workload, -w)
						delete(info, i)
					}
				}
			}

			latest = current
		}
	}

}

func toTagsMap(keys []string) map[string]struct{} {
	tags := make(map[string]struct{})
	for _, tag := range keys {
		tags[tag] = struct{}{}
	}
	return tags
}

func minDuration(d1 time.Duration, d2 time.Duration) time.Duration {
	if d1 < d2 {
		return d1
	}
	return d2
}
