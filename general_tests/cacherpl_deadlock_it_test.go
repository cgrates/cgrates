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
	"bytes"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/resources"
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
"admins": {
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
		DBCfg:     engine.MySQLDBCfg,
		LogBuffer: &bytes.Buffer{},
		Encoding:  *utils.Encoding,
	}
	t.Cleanup(func() {
		fmt.Println(ng.LogBuffer)
	})
	client, _ := ng.Run(t)

	// Set a resource profile to have an ID to cache in *event_resources.
	var reply string
	if err := client.Call(context.Background(), utils.AdminSv1SetResourceProfile,
		&utils.ResourceProfile{
			ID: "RES_1",
		},
		&reply); err != nil {
		t.Fatalf("AdminSv1.SetResourceProfile unexpected err: %v", err)
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

	var rs *resources.Resources
	if err := client.Call(context.Background(), utils.ResourceSv1GetResourcesForEvent, args, &rs); err != nil {
		t.Fatalf("ResourceSv1.GetResourcesForEvent unexpected err: %v", err)
	}

	// Remove the profile to force a NOT_FOUND error, making sure that Cache.Remove
	// from the defer statement is called (matchingResourcesForEvent).
	if err := client.Call(context.Background(), utils.AdminSv1RemoveResourceProfile,
		&utils.TenantID{
			ID: "RES_1",
		}, &reply); err != nil {
		t.Fatalf("AdminSv1.RemoveResourceProfile unexpected err: %v", err)
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

goroutine 95 [sync.RWMutex.RLock]:
sync.runtime_SemacquireRWMutexR(0x0?, 0x0?, 0x0?)
        /usr/local/go1.24.2.linux-amd64/src/runtime/sema.go:100 +0x25
sync.(*RWMutex).RLock(...)
        /usr/local/go1.24.2.linux-amd64/src/sync/rwmutex.go:74
github.com/cgrates/ltcache.(*TransCache).Get(0xc00095d640, {0x311f43e, 0x10}, {0xc00030bcb0, 0x8})
        /home/ionut/go/pkg/mod/github.com/cgrates/ltcache@v0.0.0-20250409175814-a90b4db74697/transcache.go:141 +0x67
github.com/cgrates/cgrates/engine.(*CacheS).Get(...)
        /home/ionut/work/cgrates/engine/caches.go:230
github.com/cgrates/cgrates/engine.(*ConnManager).getConn(0xc00097e180, 0x513dec0, {0xc00030bcb0, 0x8})
        /home/ionut/work/cgrates/engine/connmanager.go:59 +0x5c
github.com/cgrates/cgrates/engine.(*ConnManager).Call(0xc00097e180, 0x513dec0, {0xc000b15a00?, 0x11?, 0x1?}, {0x313a797, 0x18}, {0x2a6b1a0, 0xc00090c0c0}, {0x2a665a0, ...})
        /home/ionut/work/cgrates/engine/connmanager.go:158 +0x149
github.com/cgrates/cgrates/engine.NewCacheS.func1({0xc00049234c, 0x4}, {0xc00049234c?, 0x4?})
        /home/ionut/work/cgrates/engine/caches.go:144 +0x11e
github.com/cgrates/ltcache.(*Cache).remove(0xc000c91200, {0xc0004925a8, 0x4})
        /home/ionut/go/pkg/mod/github.com/cgrates/ltcache@v0.0.0-20250409175814-a90b4db74697/cache.go:275 +0x387
github.com/cgrates/ltcache.(*Cache).Remove(0xc000c91200, {0xc0004925a8, 0x4})
        /home/ionut/go/pkg/mod/github.com/cgrates/ltcache@v0.0.0-20250409175814-a90b4db74697/cache.go:198 +0x39
github.com/cgrates/ltcache.(*TransCache).Remove(0xc00095d640, {0x311f44e, 0x10}, {0xc0004925a8, 0x4}, 0x80?, {0x0?, 0x0})
        /home/ionut/go/pkg/mod/github.com/cgrates/ltcache@v0.0.0-20250409175814-a90b4db74697/transcache.go:172 +0x125
github.com/cgrates/cgrates/engine.(*CacheS).Remove(0xc000c95410, 0xc0006de660, {0x311f44e, 0x10}, {0xc0004925a8, 0x4}, 0x80?, {0x0?, 0x41c4f3?})
        /home/ionut/work/cgrates/engine/caches.go:259 +0x56
github.com/cgrates/cgrates/resources.(*ResourceS).matchingResourcesForEvent.func1()
        /home/ionut/work/cgrates/resources/resources.go:483 +0x4e
github.com/cgrates/cgrates/resources.(*ResourceS).matchingResourcesForEvent(0xc000bc75e0, 0xc0006de660, {0xc00002b0a4, 0xb}, 0xc00090c080, {0xc0004925a8, 0x4}, 0xc0004925c0)
        /home/ionut/work/cgrates/resources/resources.go:574 +0x1078
github.com/cgrates/cgrates/resources.(*ResourceS).V1GetResourcesForEvent(0xc000bc75e0, 0xc0006de660, 0xc00090c080, 0xc0008a0378)
        /home/ionut/work/cgrates/resources/apis.go:81 +0x899
reflect.Value.call({0xc0001439e0?, 0xc000bc84e0?, 0x13?}, {0x30ffac3, 0x4}, {0xc000095ec8, 0x4, 0x4?})
        /usr/local/go1.24.2.linux-amd64/src/reflect/value.go:584 +0xca6
reflect.Value.Call({0xc0001439e0?, 0xc000bc84e0?, 0xc000bc39b0?}, {0xc000b9dec8?, 0x0?, 0x0?})
        /usr/local/go1.24.2.linux-amd64/src/reflect/value.go:368 +0xb9
github.com/cgrates/birpc.(*Service).call(0xc00090de00, 0xc000a60fa0, 0xc000ab7b68, 0xc0008c99c8, 0x1102c60?, 0xc000bc1260, 0xc00062a300, {0x2f82f60, 0xc00090c080, 0x16}, ...)
        /home/ionut/go/pkg/mod/github.com/cgrates/birpc@v1.3.1-0.20211117095917-5b0ff29f3084/service.go:96 +0x372
created by github.com/cgrates/birpc.(*Server).ServeCodec in goroutine 74
        /home/ionut/go/pkg/mod/github.com/cgrates/birpc@v1.3.1-0.20211117095917-5b0ff29f3084/server.go:226 +0x4b0
*/
