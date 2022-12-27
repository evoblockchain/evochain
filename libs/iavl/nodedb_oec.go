package iavl

import (
	"bytes"
	"container/list"
	"encoding/binary"
	"fmt"

	cmap "github.com/orcaman/concurrent-map"

	"time"

	"github.com/go-errors/errors"
	"github.com/evoblockchain/evochain/libs/system/trace"
	dbm "github.com/evoblockchain/evochain/libs/tm-db"
	"github.com/tendermint/go-amino"
)

const (
	FlagIavlCacheInitRatio = "iavl-cache-init-ratio"
)

var (
	IavlCacheInitRatio float64 = 0
)

type tppItem struct {
	nodeMap  map[string]*Node
	listItem *list.Element
}

func (ndb *nodeDB) SaveOrphans(batch dbm.Batch, version int64, orphans []*Node) {
	ndb.mtx.Lock()
	defer ndb.mtx.Unlock()

	toVersion := ndb.getPreviousVersion(version)
	for _, node := range orphans {
		ndb.log(IavlDebug, "SAVEORPHAN", "version", node.version, "toVersion", toVersion, "hash", amino.BytesHexStringer(node.hash))
		ndb.saveOrphan(batch, node.hash, node.version, toVersion)
	}
}

func (ndb *nodeDB) dbGet(k []byte) ([]byte, error) {
	ndb.addDBReadCount()
	ts := time.Now()
	defer func() {
		ndb.addDBReadTime(time.Now().Sub(ts).Nanoseconds())
	}()
	return ndb.db.Get(k)
}

func (ndb *nodeDB) makeNodeFromDbByHash(hash []byte) *Node {
	k := ndb.nodeKey(hash)
	ndb.addDBReadCount()
	ts := time.Now()
	defer func() {
		ndb.addDBReadTime(time.Now().Sub(ts).Nanoseconds())
	}()

	v, err := ndb.db.GetUnsafeValue(k, func(buf []byte) (interface{}, error) {
		if buf == nil {
			panic(fmt.Sprintf("Value missing for hash %x corresponding to nodeKey %x", hash, ndb.nodeKey(hash)))
		}

		node, err := MakeNode(buf)
		if err != nil {
			panic(fmt.Sprintf("Error reading Node. bytes: %x, error: %v", buf, err))
		}
		return node, nil
	})

	if err != nil {
		panic(fmt.Sprintf("can't get node %X: %v", hash, err))
	}

	return v.(*Node)
}

func (ndb *nodeDB) saveNodeToPrePersistCache(node *Node) {
	if node.hash == nil {
		panic("Expected to find node.hash, but none found.")
	}
	if node.persisted || node.prePersisted {
		panic("Shouldn't be calling save on an already persisted node.")
	}

	node.prePersisted = true
	ndb.mtx.Lock()
	ndb.prePersistNodeCache[string(node.hash)] = node
	ndb.mtx.Unlock()
}

func (ndb *nodeDB) persistTpp(event *commitEvent, trc *trace.Tracer) {
	batch := event.batch
	tpp := event.tpp

	trc.Pin("batchSet")
	for _, node := range tpp {
		ndb.batchSet(node, batch)
	}
	ndb.state.increasePersistedCount(len(tpp))
	ndb.addDBWriteCount(int64(len(tpp)))

	if err := ndb.saveFastNodeVersion(batch, event.fnc, event.version); err != nil {
		panic(err)
	}

	trc.Pin("batchCommit")
	if err := ndb.Commit(batch); err != nil {
		panic(err)
	}

	ndb.asyncPersistTppFinised(event, trc)
	ndb.tpfv.remove(event.version)
}

func (ndb *nodeDB) asyncPersistTppStart(version int64) (map[string]*Node, *fastNodeChanges) {
	ndb.log(IavlDebug, "moving prePersistCache to tempPrePersistCache", "size", len(ndb.prePersistNodeCache))

	ndb.mtx.Lock()

	tpp := ndb.prePersistNodeCache
	ndb.prePersistNodeCache = make(map[string]*Node, len(tpp))

	ndb.tpp.pushToTpp(version, tpp)

	var tempPersistFastNode *fastNodeChanges
	if GetEnableFastStorage() {
		tempPersistFastNode = ndb.prePersistFastNode
		ndb.tpfv.add(version, tempPersistFastNode)
		ndb.prePersistFastNode = newFastNodeChanges()
	}
	ndb.mtx.Unlock()

	for _, node := range tpp {
		if node.persisted || !node.prePersisted {
			panic("unexpected node state")
		}
		node.persisted = true
	}

	return tpp, tempPersistFastNode
}

func (ndb *nodeDB) asyncPersistTppFinised(event *commitEvent, trc *trace.Tracer) {
	version := event.version
	iavlHeight := event.iavlHeight

	nodeNum := ndb.tpp.getTppNodesNum()

	ndb.tpp.removeFromTpp(version)

	ndb.log(IavlInfo, "CommitSchedule", "Height", version,
		"Tree", ndb.name,
		"IavlHeight", iavlHeight,
		"NodeNum", nodeNum,
		"trc", trc.Format())
}

// SaveNode saves a node to disk.
func (ndb *nodeDB) batchSet(node *Node, batch dbm.Batch) {
	if node.hash == nil {
		panic("Expected to find node.hash, but none found.")
	}
	if !node.persisted {
		panic("Should set node.persisted to true before batchSet.")
	}

	if !node.prePersisted {
		panic("Should be calling save on an prePersisted node.")
	}

	// Save node bytes to db.
	var buf bytes.Buffer
	buf.Grow(node.aminoSize())

	if err := node.writeBytesToBuffer(&buf); err != nil {
		panic(err)
	}

	nodeKey := ndb.nodeKey(node.hash)
	nodeValue := buf.Bytes()
	batch.Set(nodeKey, nodeValue)
	ndb.state.increasePersistedSize(len(nodeKey) + len(nodeValue))
	ndb.log(IavlDebug, "BATCH SAVE", "hash", node.hash)
	//node.persisted = true // move to function MovePrePersistCacheToTempCache
}

func (ndb *nodeDB) NewBatch() dbm.Batch {
	return ndb.db.NewBatch()
}

// Saves orphaned nodes to disk under a special prefix.
// version: the new version being saved.
// orphans: the orphan nodes created since version-1
func (ndb *nodeDB) saveCommitOrphans(batch dbm.Batch, version int64, orphans []commitOrphan) {
	ndb.log(IavlDebug, "saving committed orphan node log to disk")
	toVersion := ndb.getPreviousVersion(version)
	for _, orphan := range orphans {
		// ndb.log(IavlDebug, "SAVEORPHAN", "from", orphan.Version, "to", toVersion, "hash", amino.BytesHexStringer(orphan.NodeHash))
		ndb.saveOrphan(batch, orphan.NodeHash, orphan.Version, toVersion)
	}
}

func (ndb *nodeDB) getRootWithCacheAndDB(version int64) ([]byte, error) {
	if EnableAsyncCommit {
		root, ok := ndb.findRootHash(version)
		if ok {
			return root, nil
		}
	}
	return ndb.getRoot(version)
}

// DeleteVersion deletes a tree version from disk.
func (ndb *nodeDB) DeleteVersion(batch dbm.Batch, version int64, checkLatestVersion bool) error {
	err := ndb.checkoutVersionReaders(version)
	if err != nil {
		return err
	}

	ndb.deleteOrphans(batch, version)
	ndb.deleteRoot(batch, version, checkLatestVersion)
	return nil
}

func (ndb *nodeDB) checkoutVersionReaders(version int64) error {
	ndb.mtx.Lock()
	defer ndb.mtx.Unlock()
	if ndb.versionReaders[version] > 0 {
		return errors.Errorf("unable to delete version %v, it has %v active readers", version, ndb.versionReaders[version])
	}
	return nil
}

func (ndb *nodeDB) syncUnCacheNode(hash []byte) {
	ndb.mtx.Lock()
	defer ndb.mtx.Unlock()
	ndb.uncacheNode(hash)
}

// orphanKeyFast generates key for an orphan node,
// result must be equal to orphanKeyFormat.Key(toVersion, fromVersion, hash)
func orphanKeyFast(fromVersion, toVersion int64, hash []byte) []byte {
	key := make([]byte, orphanKeyFormat.length)
	key[0] = orphanKeyFormat.prefix
	n := 1
	if orphanKeyFormat.layout[0] != int64Size {
		panic("unexpected layout")
	}
	binary.BigEndian.PutUint64(key[n:n+int64Size], uint64(toVersion))
	n += int64Size
	if orphanKeyFormat.layout[1] != int64Size {
		panic("unexpected layout")
	}
	binary.BigEndian.PutUint64(key[n:n+int64Size], uint64(fromVersion))
	n += int64Size
	hashLen := orphanKeyFormat.layout[2]
	if hashLen < len(hash) {
		panic("hash is too long")
	}
	copy(key[n+hashLen-len(hash):n+hashLen], hash)
	return key
}

func (ndb *nodeDB) cacheNode(node *Node) {
	ndb.nc.cache(node)
}

func (ndb *nodeDB) cacheWithKey(key string, node *Node) {
	ndb.nc.cacheWithKey(key, node)
}

func (ndb *nodeDB) uncacheNode(hash []byte) {
	ndb.nc.uncache(hash)
}

func (ndb *nodeDB) getFastNodeFromCache(key []byte) (*FastNode, bool) {
	node := ndb.fastNodeCache.get(key, true)
	return node, node != nil
}

func (ndb *nodeDB) uncacheFastNode(key []byte) {
	ndb.fastNodeCache.uncache(key)
}

// Add a node to the cache and pop the least recently used node if we've
// reached the cache size limit.
func (ndb *nodeDB) cacheFastNode(node *FastNode) {
	ndb.fastNodeCache.cache(node)
}

func (ndb *nodeDB) getNodeFromCache(hash []byte, promoteRecentNode bool) (n *Node) {
	return ndb.nc.get(hash, promoteRecentNode)
}

func (ndb *nodeDB) uncacheNodeRontine(n []*Node) {
	for _, node := range n {
		ndb.uncacheNode(node.hash)
	}
}

func (ndb *nodeDB) initPreWriteCache() {
	if ndb.preWriteNodeCache == nil {
		ndb.preWriteNodeCache = cmap.New()
	}
}

func (ndb *nodeDB) cacheNodeToPreWriteCache(n *Node) {
	ndb.preWriteNodeCache.Set(string(n.hash), n)
}

func (ndb *nodeDB) finishPreWriteCache() {
	ndb.preWriteNodeCache.IterCb(func(key string, v interface{}) {
		ndb.cacheWithKey(key, v.(*Node))
	})
	ndb.preWriteNodeCache = nil
}

func (ndb *nodeDB) prePersistNodeCacheLen() (l int) {
	ndb.mtx.Lock()
	l = len(ndb.prePersistNodeCache)
	ndb.mtx.Unlock()
	return
}
