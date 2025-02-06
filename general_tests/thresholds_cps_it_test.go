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

func TestStressThresholdsProcessEvent(t *testing.T) {
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
	defer fmt.Println(ng.LogBuffer)
	client, _ := ng.Run(t)
	t.Run("SetThresholdProfile", func(t *testing.T) {
		var reply string
		for i := 1; i <= 10; i++ {
			thresholdPrf := &engine.ThresholdProfileWithAPIOpts{
				ThresholdProfile: &engine.ThresholdProfile{
					Tenant:    "cgrates.org",
					ID:        fmt.Sprintf("THD_%d", i),
					FilterIDs: []string{fmt.Sprintf("*string:~*req.Account:100%d", i)},
					ActivationInterval: &utils.ActivationInterval{
						ActivationTime: time.Date(2024, 7, 14, 14, 35, 0, 0, time.UTC),
					},
					MaxHits: -1,
					Blocker: false,
					Weight:  20.0,
					Async:   true,
				},
			}
			if err := client.Call(context.Background(), utils.APIerSv1SetThresholdProfile, thresholdPrf, &reply); err != nil {
				t.Error(err)
			}
		}
	})

	t.Run("ThresholdExportEvent", func(t *testing.T) {
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
					ID:     utils.GenUUID(),
					Time:   utils.TimePointer(time.Now()),
					Event: map[string]any{
						utils.AccountField: fmt.Sprintf("100%d", ((i-1)%10)+1),
						utils.AnswerTime:   utils.TimePointer(time.Now()),
						utils.Usage:        45,
						utils.Cost:         12.1,
						utils.Tenant:       "cgrates.org",
						utils.Category:     "call",
					},
				}

				var reply []string
				if err := client.Call(context.Background(), utils.ThresholdSv1ProcessEvent, args, &reply); err != nil {
					errCH <- fmt.Errorf("no loop %d event id %s failed with err: %v", i, args.Event[utils.AccountField], err)
					return
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
