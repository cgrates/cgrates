// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
)

// indexPatchKey marks context patches for merging on retry and safe failed filenames.
func indexPatchKey(key string) string {
	return "\x00" + key
}

// failedReplicationFileName preserves existing key names. It hashes marked
// index keys into names of a fixed length that are safe to use in paths.
func failedReplicationFileName(key string) string {
	if key == "" || key[0] != 0 {
		return key
	}
	return fmt.Sprintf("index_%x", sha256.Sum256([]byte(key)))
}

// replicationData holds the information about a pending replication task.
type replicationData struct {
	objType string
	objID   string
	method  string
	args    any
}

// replicator manages replication tasks to synchronize data across instances.
// It can perform immediate replication or batch tasks to replicate on intervals.
type replicator struct {
	mu sync.Mutex

	cm    *ConnManager
	conns []string // ids of connections to replicate to

	// pending stores the latest version of the object, named by the key, that
	// is to be replicated.
	pending map[string]*replicationData

	interval  time.Duration  // replication frequency
	failedDir string         // where failed replications are stored (one per id)
	filtered  bool           // whether to replicate only objects coming from remote
	stop      chan struct{}  // stop replication loop
	wg        sync.WaitGroup // wait for any pending replications before closing
}

// newReplicator creates a replication manager that either performs immediate
// or batched replications based on configuration.
// When interval > 0, replications are queued and processed in batches at that interval.
// When interval = 0, each replication is performed immediately when requested.
func newReplicator(cm *ConnManager) *replicator {
	cfg := config.CgrConfig().DataDbCfg()
	r := &replicator{
		cm:        cm,
		pending:   make(map[string]*replicationData),
		interval:  cfg.RplInterval,
		failedDir: cfg.RplFailedDir,
		conns:     cfg.RplConns,
		filtered:  cfg.RplFiltered,
		stop:      make(chan struct{}),
	}
	if r.interval > 0 {
		r.wg.Add(1)
		go r.replicationLoop()
	}
	return r

}

// replicationKey gives Set and Remove the same key only when the latest
// operation makes the earlier one unnecessary.
func replicationKey(objType, objID, method string) string {
	switch method {
	case utils.ReplicatorSv1SetAccount, utils.ReplicatorSv1RemoveAccount,
		utils.ReplicatorSv1SetDestination, utils.ReplicatorSv1RemoveDestination,
		utils.ReplicatorSv1SetThresholdProfile, utils.ReplicatorSv1RemoveThresholdProfile,
		utils.ReplicatorSv1SetThreshold, utils.ReplicatorSv1RemoveThreshold,
		utils.ReplicatorSv1SetStatQueueProfile, utils.ReplicatorSv1RemoveStatQueueProfile,
		utils.ReplicatorSv1SetStatQueue, utils.ReplicatorSv1RemoveStatQueue,
		utils.ReplicatorSv1SetFilter, utils.ReplicatorSv1RemoveFilter,
		utils.ReplicatorSv1SetRankingProfile, utils.ReplicatorSv1RemoveRankingProfile,
		utils.ReplicatorSv1SetTrendProfile, utils.ReplicatorSv1RemoveTrendProfile,
		utils.ReplicatorSv1SetTrend, utils.ReplicatorSv1RemoveTrend,
		utils.ReplicatorSv1SetTiming, utils.ReplicatorSv1RemoveTiming,
		utils.ReplicatorSv1SetResourceProfile, utils.ReplicatorSv1RemoveResourceProfile,
		utils.ReplicatorSv1SetResource, utils.ReplicatorSv1RemoveResource,
		utils.ReplicatorSv1SetIPProfile, utils.ReplicatorSv1RemoveIPProfile,
		utils.ReplicatorSv1SetIPAllocations, utils.ReplicatorSv1RemoveIPAllocations,
		utils.ReplicatorSv1SetActionTriggers, utils.ReplicatorSv1RemoveActionTriggers,
		utils.ReplicatorSv1SetSharedGroup, utils.ReplicatorSv1RemoveSharedGroup,
		utils.ReplicatorSv1SetActions, utils.ReplicatorSv1RemoveActions,
		utils.ReplicatorSv1SetActionPlan, utils.ReplicatorSv1RemoveActionPlan,
		utils.ReplicatorSv1SetAccountActionPlans, utils.ReplicatorSv1RemAccountActionPlans,
		utils.ReplicatorSv1SetRouteProfile, utils.ReplicatorSv1RemoveRouteProfile,
		utils.ReplicatorSv1SetAttributeProfile, utils.ReplicatorSv1RemoveAttributeProfile,
		utils.ReplicatorSv1SetChargerProfile, utils.ReplicatorSv1RemoveChargerProfile,
		utils.ReplicatorSv1SetDispatcherProfile, utils.ReplicatorSv1RemoveDispatcherProfile,
		utils.ReplicatorSv1SetDispatcherHost, utils.ReplicatorSv1RemoveDispatcherHost:
		return objType + objID
	default:
		_, methodName, _ := strings.Cut(method, utils.NestingSep)
		return methodName + "_" + objType + objID
	}
}

// replicate handles the object replication based on configuration.
// When interval > 0, the replication task is queued for the next batch.
// Otherwise, it executes immediately.
func (r *replicator) replicate(objType, objID, method string, args any,
	item *config.ItemOpt) error {
	if !item.Replicate {
		return nil
	}
	key := replicationKey(objType, objID, method)
	if r.interval > 0 {
		r.mu.Lock()
		defer r.mu.Unlock()
		r.pending[key] = &replicationData{
			objType: objType,
			objID:   objID,
			method:  method,
			args:    args,
		}
		return nil
	}
	return r.replicateAndRestore(key, objType, objID, method, args)
}

func mergeIndexPatch(patch, newer *utils.SetIndexesArg) {
	if newer.Clear {
		clear(patch.Indexes)
		patch.Clear = true
	}
	if patch.Indexes == nil && len(newer.Indexes) != 0 {
		patch.Indexes = make(map[string]utils.StringSet, len(newer.Indexes))
	}
	maps.Copy(patch.Indexes, newer.Indexes)
	patch.IdxItmType = newer.IdxItmType
	patch.TntCtx = newer.TntCtx
	patch.Tenant = newer.Tenant
	patch.APIOpts = newer.APIOpts
}

func (r *replicator) replicateIndexes(objType, objID string, args *utils.SetIndexesArg) error {
	key := indexPatchKey(replicationKey(objType, objID, utils.ReplicatorSv1SetIndexes))
	if r.interval <= 0 {
		return r.replicateIndexPatchAndRestore(key, objType, objID, args)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if pending, has := r.pending[key]; has {
		mergeIndexPatch(pending.args.(*utils.SetIndexesArg), args)
		return nil
	}
	patch := *args
	patch.Indexes = maps.Clone(args.Indexes)
	r.pending[key] = &replicationData{
		objType: objType,
		objID:   objID,
		method:  utils.ReplicatorSv1SetIndexes,
		args:    &patch,
	}
	return nil
}

func (r *replicator) replicateIndexPatchAndRestore(key, objType, objID string,
	args *utils.SetIndexesArg) error {
	var failedPath string
	if r.failedDir != "" {
		failedPath = filepath.Join(r.failedDir, failedReplicationFileName(key)+utils.GOBSuffix)
		task, err := NewReplicationTaskFromFile(failedPath)
		if err == nil {
			if task == nil {
				utils.Logger.Err(fmt.Sprintf(
					"<DataManager> invalid failed index patch in %q", failedPath))
			} else if failed, ok := task.Args.(*utils.SetIndexesArg); !ok ||
				task.Method != utils.ReplicatorSv1SetIndexes ||
				task.ObjType != objType || task.ObjID != objID {
				utils.Logger.Err(fmt.Sprintf(
					"<DataManager> invalid failed index patch in %q", failedPath))
			} else {
				mergeIndexPatch(failed, args)
				args = failed
			}
		} else if !os.IsNotExist(err) {
			utils.Logger.Err(fmt.Sprintf(
				"<DataManager> failed to load index patch %q: %v", failedPath, err))
		}
	}
	err := replicate(r.cm, r.conns, r.filtered, objType, objID,
		utils.ReplicatorSv1SetIndexes, args)
	if err != nil && failedPath != "" {
		task := &ReplicationTask{
			ConnIDs:  r.conns,
			Filtered: r.filtered,
			ObjType:  objType,
			ObjID:    objID,
			Method:   utils.ReplicatorSv1SetIndexes,
			Args:     args,
		}
		if wErr := task.WriteToFile(failedPath); wErr != nil {
			utils.Logger.Err(fmt.Sprintf(
				"<DataManager> failed to save index patch %q: %v", failedPath, wErr))
		}
	}
	return err
}

// replicateAndRestore is a wrapper over replicate function and checks failedDir
// to create files for tracking unsuccessful writes
func (r *replicator) replicateAndRestore(key string, objType, objID, method string, args any) error {
	if key != "" && key[0] == 0 {
		patch, ok := args.(*utils.SetIndexesArg)
		if !ok {
			return fmt.Errorf("invalid index replication patch %T", args)
		}
		return r.replicateIndexPatchAndRestore(key, objType, objID, patch)
	}
	var failedPath string
	if r.failedDir != "" {
		failedPath = filepath.Join(r.failedDir, failedReplicationFileName(key)+utils.GOBSuffix)
		// Clean up any existing file containing failed replications.
		if err := os.Remove(failedPath); err != nil && !os.IsNotExist(err) {
			utils.Logger.Warning(fmt.Sprintf(
				"<DataManager> failed to remove file %q: %v", failedPath, err))
		}
	}
	err := replicate(r.cm, r.conns, r.filtered, objType, objID, method, args)
	if err != nil && failedPath != "" {
		task := &ReplicationTask{
			ConnIDs:  r.conns,
			Filtered: r.filtered,
			ObjType:  objType,
			ObjID:    objID,
			Method:   method,
			Args:     args,
		}
		if wErr := task.WriteToFile(failedPath); wErr != nil {
			utils.Logger.Err(fmt.Sprintf(
				"<DataManager> failed to dump replication task: %v", wErr))
		}
	}
	return err
}

// replicate performs the actual replication by calling Set/Remove APIs on ReplicatorSv1
// It either replicates to all connections or only to filtered ones based on configuration.
func replicate(connMgr *ConnManager, connIDs []string, filtered bool, objType, objID, method string, args any) (err error) {
	// the reply is string for Set/Remove APIs
	// ignored in favor of the error
	var reply string
	if !filtered {
		// is not partial so send to all defined connections
		return utils.CastRPCErr(connMgr.Call(context.TODO(), connIDs, method, args, &reply))
	}
	// is partial so get all the replicationHosts from cache based on object Type and ID
	// alp_cgrates.org:ATTR1
	rplcHostIDsIfaces := Cache.tCache.GetGroupItems(utils.CacheReplicationHosts, objType+objID)
	rplcHostIDs := make(utils.StringSet)
	for _, hostID := range rplcHostIDsIfaces {
		rplcHostIDs.Add(hostID.(string))
	}
	// using the replication hosts call the method
	return utils.CastRPCErr(connMgr.CallWithConnIDs(connIDs, rplcHostIDs,
		method, args, &reply))
}

// replicationLoop runs periodically according to the configured interval
// to flush pending replications. It stops when the Replicator is closed.
func (r *replicator) replicationLoop() {
	defer r.wg.Done()
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			r.flush()
		case <-r.stop:
			r.flush()
			return
		}
	}
}

// flush immediately processes all pending replications.
// Failed replications are saved to disk if a failedDir is configured.
func (r *replicator) flush() {
	r.mu.Lock()
	if len(r.pending) == 0 {
		// Skip processing when there are no pending replications.
		r.mu.Unlock()
		return
	}
	pending := r.pending
	r.pending = make(map[string]*replicationData)
	r.mu.Unlock()
	for key, data := range pending {
		if err := r.replicateAndRestore(key, data.objType, data.objID, data.method, data.args); err != nil {
			utils.Logger.Warning(fmt.Sprintf(
				"<DataManager> failed to replicate %q for object %q: %v",
				data.method, data.objType+data.objID, err))
		}
	}
}

// close stops the replication loop if it's running and waits for pending
// replications to complete.
func (r *replicator) close() {
	if r.interval > 0 {
		close(r.stop)
		r.wg.Wait()
	}
}

// UpdateReplicationFilters sets the connection ID in cache for filtered replication.
// It's a no-op if connID is empty.
func UpdateReplicationFilters(objType, objID, connID string) {
	if connID == utils.EmptyString {
		return
	}
	Cache.SetWithoutReplicate(utils.CacheReplicationHosts, objType+objID+utils.ConcatenatedKeySep+connID, connID, []string{objType + objID},
		true, utils.NonTransactional)
}

// replicateMultipleIDs replicates operations for multiple object IDs.
// It functions similarly to replicate but handles a collection of IDs rather than a single one.
// Used primarily for setting LoadIDs.
// TODO: merge with replicate function.
func replicateMultipleIDs(connMgr *ConnManager, connIDs []string, filtered bool, objType string, objIDs []string, method string, args any) (err error) {
	// the reply is string for Set/Remove APIs
	// ignored in favor of the error
	var reply string
	if !filtered {
		// is not partial so send to all defined connections
		return utils.CastRPCErr(connMgr.Call(context.TODO(), connIDs, method, args, &reply))
	}
	// is partial so get all the replicationHosts from cache based on object Type and ID
	// combine all hosts in a single set so if we receive a get with one ID in list
	// send all list to that hos
	rplcHostIDs := make(utils.StringSet)
	for _, objID := range objIDs {
		rplcHostIDsIfaces := Cache.tCache.GetGroupItems(utils.CacheReplicationHosts, objType+objID)
		for _, hostID := range rplcHostIDsIfaces {
			rplcHostIDs.Add(hostID.(string))
		}
	}
	// using the replication hosts call the method
	return utils.CastRPCErr(connMgr.CallWithConnIDs(connIDs, rplcHostIDs,
		method, args, &reply))
}

// ReplicationTask represents a replication operation that can be saved to disk
// and executed later, typically used for failed replications.
type ReplicationTask struct {
	ConnIDs  []string
	Filtered bool
	Path     string
	ObjType  string
	ObjID    string
	Method   string
	Args     any
}

// NewReplicationTaskFromFile loads a replication task from the specified file.
// The file is removed after successful loading.
func NewReplicationTaskFromFile(path string) (*ReplicationTask, error) {
	var taskBytes []byte
	if err := guardian.Guardian.Guard(func() error {
		var err error
		if taskBytes, err = os.ReadFile(path); err != nil {
			return err
		}
		return os.Remove(path) // file is not needed anymore
	}, config.CgrConfig().GeneralCfg().LockingTimeout, utils.FileLockPrefix+path); err != nil {
		return nil, err
	}
	dec := gob.NewDecoder(bytes.NewBuffer(taskBytes))
	var task *ReplicationTask
	if err := dec.Decode(&task); err != nil {
		return nil, err
	}
	return task, nil
}

// WriteToFile saves the replication task to the specified path.
// This allows failed tasks to be recovered and retried later.
func (r *ReplicationTask) WriteToFile(path string) error {
	return guardian.Guardian.Guard(func() error {
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()
		enc := gob.NewEncoder(f)
		return enc.Encode(r)
	}, config.CgrConfig().GeneralCfg().LockingTimeout, utils.FileLockPrefix+path)
}

// Execute performs the replication task.
func (r *ReplicationTask) Execute(cm *ConnManager) error {
	return replicate(cm, r.ConnIDs, r.Filtered, r.ObjType, r.ObjID, r.Method, r.Args)
}
