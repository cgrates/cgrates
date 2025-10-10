/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package engine

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/guardian"
)

// replicationData holds the information about a pending replication task.
type replicationData struct {
	objType string
	objID   string
	method  string
	args    any
}

// replicator manages replication tasks to synchronize data across instances.
// It can perform immediate replication or batch tasks to replicate on intervals.
//
// For failed replications, files are created with predictable names based on
// "methodName_objTypeObjID" as the key. Before each replication attempt, any existing
// file for that key is removed. A new file is created only if the replication fails.
// This ensures at most one failed replication file exists per unique item.
type replicator struct {
	mu    sync.Mutex
	ctx   *context.Context
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
func newReplicator(cm *ConnManager, cfg *config.DataDbCfg) *replicator {
	r := &replicator{
		cm:        cm,
		ctx:       context.Background(),
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

// replicate handles the object replication based on configuration.
// When interval > 0, the replication task is queued for the next batch.
// Otherwise, it executes immediately.
func (r *replicator) replicate(ctx *context.Context, objType, objID, method string, args any,
	item *config.ItemOpts) error {
	if !item.Replicate {
		return nil
	}

	if r.interval > 0 {

		// Form a unique key by joining method name with object identifiers.
		// Including the method name (Set/Remove) allows different operations
		// on the same object to have distinct keys, which also serve as
		// predictable filenames if replication fails.
		_, methodName, _ := strings.Cut(method, utils.NestingSep)
		key := methodName + "_" + objType + objID

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

	return replicate(ctx, r.cm, r.conns, r.filtered, objType, objID, method, args)
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
		var failedPath string

		if r.failedDir != "" {
			failedPath = filepath.Join(r.failedDir, key+utils.GOBSuffix)

			// Clean up any existing file containing failed replications.
			if err := os.Remove(failedPath); err != nil && !os.IsNotExist(err) {
				utils.Logger.Warning(fmt.Sprintf(
					"<DataManager> failed to remove file for %q: %v", key, err))
			}
		}

		if err := replicate(r.ctx, r.cm, r.conns, r.filtered, data.objType, data.objID,
			data.method, data.args); err != nil {
			utils.Logger.Warning(fmt.Sprintf(
				"<DataManager> failed to replicate %q for object %q: %v",
				data.method, data.objType+data.objID, err))

			if failedPath != "" {
				task := &ReplicationTask{
					ConnIDs:  r.conns,
					Filtered: r.filtered,
					ObjType:  data.objType,
					ObjID:    data.objID,
					Method:   data.method,
					Args:     data.args,
				}
				if err := task.WriteToFile(r.ctx, failedPath); err != nil {
					utils.Logger.Err(fmt.Sprintf(
						"<DataManager> failed to dump replication task: %v", err))
				}
			}
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

// UpdateReplicationFilters will set the connID in cache
func UpdateReplicationFilters(objType, objID, connID string) {
	if connID == utils.EmptyString {
		return
	}
	Cache.SetWithoutReplicate(utils.CacheReplicationHosts, objType+objID+utils.ConcatenatedKeySep+connID, connID, []string{objType + objID},
		true, utils.NonTransactional)
}

// replicate will call Set/Remove APIs on ReplicatorSv1
func replicate(ctx *context.Context, connMgr *ConnManager, connIDs []string, filtered bool, objType, objID, method string, args any) (err error) {
	// the reply is string for Set/Remove APIs
	// ignored in favor of the error
	var reply string
	if !filtered {
		// is not partial so send to all defined connections
		return utils.CastRPCErr(connMgr.Call(ctx, connIDs, method, args, &reply))
	}
	// is partial so get all the replicationHosts from cache based on object Type and ID
	// alp_cgrates.org:ATTR1
	rplcHostIDsIfaces := Cache.tCache.GetGroupItems(utils.CacheReplicationHosts, objType+objID)
	rplcHostIDs := make(utils.StringSet)
	for _, hostID := range rplcHostIDsIfaces {
		rplcHostIDs.Add(hostID.(string))
	}
	// using the replication hosts call the method
	return utils.CastRPCErr(connMgr.CallWithConnIDs(connIDs, ctx, rplcHostIDs,
		method, args, &reply))
}

// replicateMultipleIDs will do the same thing as replicate but uses multiple objectIDs
// used when setting the LoadIDs
func replicateMultipleIDs(ctx *context.Context, connMgr *ConnManager, connIDs []string, filtered bool, objType string, objIDs []string, method string, args any) (err error) {
	// the reply is string for Set/Remove APIs
	// ignored in favor of the error
	var reply string
	if !filtered {
		// is not partial so send to all defined connections
		return utils.CastRPCErr(connMgr.Call(ctx, connIDs, method, args, &reply))
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
	return utils.CastRPCErr(connMgr.CallWithConnIDs(connIDs, ctx, rplcHostIDs,
		method, args, &reply))
}

// ReplicationTask represents a replication operation that can be saved to disk
// and executed later, typically used for failed replications.
type ReplicationTask struct {
	ConnIDs   []string
	Filtered  bool
	Path      string
	ObjType   string
	ObjID     string
	Method    string
	Args      any
	failedDir string
}

// NewReplicationTaskFromFile loads a replication task from the specified file.
// The file is removed after successful loading.
func NewReplicationTaskFromFile(ctx *context.Context, path string) (*ReplicationTask, error) {
	var taskBytes []byte
	if err := guardian.Guardian.Guard(ctx, func(ctx *context.Context) error {
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
func (r *ReplicationTask) WriteToFile(ctx *context.Context, path string) error {
	return guardian.Guardian.Guard(ctx, func(ctx *context.Context) error {
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
func (r *ReplicationTask) Execute(ctx *context.Context, cm *ConnManager) error {
	return replicate(ctx, cm, r.ConnIDs, r.Filtered, r.ObjType, r.ObjID, r.Method, r.Args)
}
