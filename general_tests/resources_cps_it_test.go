//go:build performance

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
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestStressResourceProcessEvent(t *testing.T) {
	var dbConfig engine.DBCfg
	switch *utils.DBType {
	case utils.MetaInternal:
		dbConfig = engine.DBCfg{
			DataDB: &engine.DBParams{
				Type: utils.StringPointer(utils.MetaInternal),
			},
			StorDB: &engine.DBParams{
				Type: utils.StringPointer(utils.MetaInternal),
			},
		}
	case utils.MetaMySQL:
	case utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}
	content := `{

	"general": {
		"log_level": 7,
	},
    "data_db": {								 
    	"db_type": "redis",						
    	"db_port": 6379, 						
    	"db_name": "10", 						
    },
    "stor_db": {
    	"db_password": "CGRateS.org",
    },
	"apiers": {
		"enabled": true
	},
	"stats": {
		"enabled": true,
		"store_interval": "-1",
	},
    	"resources": {
    	"enabled": true,
    	"store_interval": "-1",
    	"thresholds_conns": ["*internal"]
    },
  	"thresholds": {
  		"enabled": true,
		"store_interval": "-1",
       },
	}`

	ng := engine.TestEngine{
		ConfigJSON: content,
		DBCfg:      dbConfig,
		LogBuffer:  bytes.NewBuffer(nil),
	}
	client, _ := ng.Run(t)
	t.Run("SetResourceProfile", func(t *testing.T) {
		var result string
		for i := 1; i <= 10; i++ {
			rls := &engine.ResourceProfileWithAPIOpts{
				ResourceProfile: &engine.ResourceProfile{
					Tenant:            "cgrates.org",
					ID:                fmt.Sprintf("RES_%d", i),
					FilterIDs:         []string{fmt.Sprintf("*string:~*req.Account:100%d", i)},
					UsageTTL:          -1,
					Limit:             -1,
					AllocationMessage: "Account1Channels",
					Weight:            20,
					ThresholdIDs:      []string{utils.MetaNone},
				},
			}
			if err := client.Call(context.Background(), utils.APIerSv1SetResourceProfile, rls, &result); err != nil {
				t.Error(err)
			} else if result != utils.OK {
				t.Error("Unexpected reply returned", result)
			}
		}
	})

	t.Run("StatExportEvent", func(t *testing.T) {
		var wg sync.WaitGroup
		start := time.Now()
		var sucessEvent atomic.Int32
		errCH := make(chan error, *count)
		for i := 1; i <= *count; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				args := &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     utils.UUIDSha1Prefix(),
					Event: map[string]any{
						utils.AccountField: fmt.Sprintf("100%d", ((i-1)%10)+1),
						utils.Destination:  "3420340",
					},
					APIOpts: map[string]any{
						utils.OptsResourcesUsageID: "651a8db2-4f67-4cf8-b622-169e8a482e45",
						utils.OptsResourcesUnits:   6,
					},
				}
				var reply string
				if err := client.Call(context.Background(), utils.ResourceSv1AllocateResources,
					args, &reply); err != nil {
					errCH <- fmt.Errorf("no loop %d event id %s failed with err: %v", i, args.Event[utils.AccountField], err)
					return
				} else if reply != "Account1Channels" {
					t.Error("Unexpected reply returned", reply)
				}
				sucessEvent.Add(1)
			}(i)
		}
		doneCH := make(chan struct{})
		go func() {
			wg.Wait()
			close(doneCH)
		}()
		select {
		case <-doneCH:
		case <-time.After(10 * time.Minute):
			t.Error("timeout")
		}
		close(errCH)
		for err := range errCH {
			t.Error(err)
		}
		t.Logf("Processed %v events in %v (%.2f events/sec)", sucessEvent.Load(), time.Since(start), float64(*count)/time.Since(start).Seconds())
	})

}
