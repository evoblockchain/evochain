package monitor

import (
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/evoblockchain/evochain/libs/tendermint/libs/log"
)

// RpcMetrics ...
type RpcMetrics struct {
	Counter   metrics.Counter
	Histogram metrics.Histogram
}

type Monitor struct {
	method   string
	logger   log.Logger
	lastTime time.Time
	metrics  map[string]*RpcMetrics
}

func GetMonitor(method string, logger log.Logger, metrics map[string]*RpcMetrics) *Monitor {
	return &Monitor{
		method:  method,
		logger:  logger,
		metrics: metrics,
	}
}

func (m *Monitor) OnBegin() *Monitor {
	m.lastTime = time.Now()

	if m.metrics == nil {
		return m
	}

	if _, ok := m.metrics[m.method]; ok {
		m.metrics[m.method].Counter.Add(1)
	}

	return m
}

func (m *Monitor) OnEnd(args ...interface{}) {
	elapsed := time.Since(m.lastTime).Seconds()
	m.logger.Debug("RPC", "Method", m.method, "Elapsed", elapsed*1e3, "Params", args)

	if m.metrics == nil {
		return
	}

	if _, ok := m.metrics[m.method]; ok {
		m.metrics[m.method].Histogram.Observe(elapsed)
	}
}
