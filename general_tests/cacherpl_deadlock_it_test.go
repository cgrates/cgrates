//go:build flaky

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package general_tests

import (
	"errors"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestCacheRplDeadlock(t *testing.T) {
	// NOTE: this cannot work with *internal because cache needs to be enabled
	switch *utils.DBType {
	case utils.MetaMySQL:
	case utils.MetaInternal, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	ng := engine.TestEngine{
		ConfigJSON: `{
"resources": {
	"enabled": true
},
"apiers": {
	"enabled": true
},
"caches": {
	"partitions": {
		"*event_resources": {
			"limit": -1,
			"replicate": true
		}
	},

	// replicate to self for convenience
	"replication_conns": ["rpl_conn"]
},
"rpc_conns": {
	"rpl_conn": {
		"conns": [
			{
				"address": "127.0.0.1:2013",
				"transport":"*gob",
			}
		]
	}
}
}`,
	}
	client, _ := ng.Run(t)

	// Set a resource profile to have an ID to cache in *event_resources.
	var reply string
	if err := client.Call(context.Background(), utils.APIerSv1SetResourceProfile,
		&engine.ResourceProfile{
			ID: "RES_1",
		},
		&reply); err != nil {
		t.Fatalf("APIerSv1.SetResourceProfile unexpected err: %v", err)
	}

	args := &utils.CGREvent{
		ID: "ev1",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID: "test",
		},
	}

	var rs *engine.Resources
	if err := client.Call(context.Background(), utils.ResourceSv1GetResourcesForEvent, args, &rs); err != nil {
		t.Fatalf("ResourceSv1.GetResourcesForEvent unexpected err: %v", err)
	}

	// Remove the profile to force a NOT_FOUND error, making sure that Cache.Remove
	// from the defer statement is called (matchingResourcesForEvent).
	if err := client.Call(context.Background(), utils.APIerSv1RemoveResourceProfile,
		&utils.TenantID{
			ID: "RES_1",
		}, &reply); err != nil {
		t.Fatalf("APIerSv1.RemoveResourceProfile unexpected err: %v", err)
	}

	// Define a context with Timeout since the following request will deadlock.
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	if err := client.Call(ctx, utils.ResourceSv1GetResourcesForEvent, args, &rs); errors.Is(err, context.DeadlineExceeded) {
		// we don't care about the error as long as it's not of type context.DeadlineExceeded
		t.Errorf("ResourceSv1.GetResourcesForEvent unexpected err: %v", err)
	}
}

/*
The second call to ResourceSv1.GetResourcesForEvent causes a deadlock:

1. Cache.Remove is called with a defer statement, locking the transcache.
2. The onEvicted function is triggered, sending a CacheSv1.ReplicateRemove
request to all partitions with replicate set to true.
3. While processing the request, connManager looks for cached connections and
attempts to lock the transcache again, causing a deadlock.

Note: Assuming the deadlock is fixed, the Remove method in CacheS calls
ReplicateRemove explicitly after removing an item. This means ReplicateRemove
is called twice: once by the onEvicted function and once in the Remove method
itself.

Below is the complete goroutine stack trace for reference:

goroutine 109 [sync.RWMutex.RLock]:
sync.runtime_SemacquireRWMutexR(0xf31000078616d20?, 0x88?, 0x66?)
        /usr/local/go1.24.2.linux-amd64/src/runtime/sema.go:100 +0x25
sync.(*RWMutex).RLock(...)
        /usr/local/go1.24.2.linux-amd64/src/sync/rwmutex.go:74
github.com/cgrates/ltcache.(*TransCache).Get(0xc000996200, {0x312f735, 0x10}, {0xc000951440, 0x8})
        /home/ionut/go/pkg/mod/github.com/cgrates/ltcache@v0.0.0-20250409175814-a90b4db74697/transcache.go:141 +0x67
github.com/cgrates/cgrates/engine.(*CacheS).Get(...)
        /home/ionut/work/cgrates/engine/caches.go:256
github.com/cgrates/cgrates/engine.(*ConnManager).getConn(0xc000acfe20, 0x51656a0, {0xc000951440, 0x8})
        /home/ionut/work/cgrates/engine/connmanager.go:75 +0x5c
github.com/cgrates/cgrates/engine.(*ConnManager).Call(0xc000acfe20, 0x51656a0, {0xc00082ed50?, 0x1?, 0x51ae900?}, {0x314b2e1, 0x18}, {0x2a6e880, 0xc000a59980}, {0x2a689c0, ...})
        /home/ionut/work/cgrates/engine/connmanager.go:181 +0x17a
github.com/cgrates/cgrates/engine.NewCacheS.func1({0xc00098bb20, 0x4}, {0xc00098bb20?, 0x4?})
        /home/ionut/work/cgrates/engine/caches.go:168 +0x118
github.com/cgrates/ltcache.(*Cache).remove(0xc00058a870, {0xc0009b8400, 0x4})
        /home/ionut/go/pkg/mod/github.com/cgrates/ltcache@v0.0.0-20250409175814-a90b4db74697/cache.go:275 +0x387
github.com/cgrates/ltcache.(*Cache).Remove(0xc00058a870, {0xc0009b8400, 0x4})
        /home/ionut/go/pkg/mod/github.com/cgrates/ltcache@v0.0.0-20250409175814-a90b4db74697/cache.go:198 +0x39
github.com/cgrates/ltcache.(*TransCache).Remove(0xc000996200, {0x312fa15, 0x10}, {0xc0009b8400, 0x4}, 0x90?, {0x0?, 0x0})
        /home/ionut/go/pkg/mod/github.com/cgrates/ltcache@v0.0.0-20250409175814-a90b4db74697/transcache.go:172 +0x125
github.com/cgrates/cgrates/engine.(*CacheS).Remove(0xc000abb180, {0x312fa15, 0x10}, {0xc0009b8400, 0x4}, 0x40?, {0x0?, 0x41c073?})
        /home/ionut/work/cgrates/engine/caches.go:285 +0x38
github.com/cgrates/cgrates/engine.(*ResourceService).matchingResourcesForEvent.func1()
        /home/ionut/work/cgrates/engine/resources.go:650 +0x49
github.com/cgrates/cgrates/engine.(*ResourceService).matchingResourcesForEvent(0xc0009acd20, {0xc000a65fa4, 0xb}, 0xc000a59940, {0xc0009b8400, 0x4}, 0x0)
        /home/ionut/work/cgrates/engine/resources.go:733 +0xe82
github.com/cgrates/cgrates/engine.(*ResourceService).V1GetResourcesForEvent(0xc0009acd20, 0x2?, 0xc000a59940, 0xc00097ac48)
        /home/ionut/work/cgrates/engine/resources.go:791 +0x5f2
github.com/cgrates/cgrates/apier/v1.(*ResourceSv1).GetResourcesForEvent(0x0?, 0x2?, 0x0?, 0x0?)
        /home/ionut/work/cgrates/apier/v1/resourcesv1.go:41 +0x16
reflect.Value.call({0xc000a626c0?, 0xc0005327d0?, 0x13?}, {0x310f9d0, 0x4}, {0xc000096ec8, 0x4, 0x4?})
        /usr/local/go1.24.2.linux-amd64/src/reflect/value.go:584 +0xca6
reflect.Value.Call({0xc000a626c0?, 0xc0005327d0?, 0xc0009c9b90?}, {0xc0008cfec8?, 0xc0008cfe58?, 0x47c70f?})
        /usr/local/go1.24.2.linux-amd64/src/reflect/value.go:368 +0xb9
github.com/cgrates/birpc.(*Service).call(0xc000996ec0, 0xc0009ac000, 0xc00098b940, 0xc000011ef0, 0x124b3c0?, 0xc00087c770, 0xc000752580, {0x2fda640, 0xc000a59940, 0x16}, ...)
        /home/ionut/go/pkg/mod/github.com/cgrates/birpc@v1.3.1-0.20211117095917-5b0ff29f3084/service.go:96 +0x372
created by github.com/cgrates/birpc.(*Server).ServeCodec in goroutine 91
        /home/ionut/go/pkg/mod/github.com/cgrates/birpc@v1.3.1-0.20211117095917-5b0ff29f3084/server.go:226 +0x4b0
*/
