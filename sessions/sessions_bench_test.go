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

package sessions

import (
	"flag"
	"fmt"
	"log"
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"sync"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	sBenchCfg *config.CGRConfig
	sBenchRPC *rpc.Client
	connOnce  sync.Once
	initRuns  = flag.Int("init_runs", 25000, "number of loops to run in init")
)

func startRPC() {
	var err error
	sBenchCfg, err = config.NewCGRConfigFromPath(
		path.Join(config.CgrConfig().DataFolderPath, "conf", "samples", "tutmysql"))
	if err != nil {
		log.Fatal(err)
	}
	config.SetCgrConfig(sBenchCfg)
	if sBenchRPC, err = jsonrpc.Dial("tcp", sBenchCfg.ListenCfg().RPCJSONListen); err != nil {
		log.Fatalf("Error at dialing rcp client:%v\n", err)
	}
}

func addBalance(sBenchRPC *rpc.Client, sraccount string) {
	attrSetBalance := utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     sraccount,
		BalanceType: utils.VOICE,
		BalanceID:   utils.StringPointer("TestDynamicDebitBalance"),
		Value:       utils.Float64Pointer(5 * float64(time.Hour)),
	}
	var reply string
	if err := sBenchRPC.Call("ApierV2.SetBalance",
		attrSetBalance, &reply); err != nil {
		log.Fatal(err)
	}
}

func addAccouns() {
	var wg sync.WaitGroup
	for i := 0; i < *initRuns; i++ {
		wg.Add(1)
		go func(i int, sBenchRPC *rpc.Client) {
			addBalance(sBenchRPC, fmt.Sprintf("1001%v", i))
			addBalance(sBenchRPC, fmt.Sprintf("1002%v", i))
			wg.Done()
		}(i, sBenchRPC)
	}
	wg.Wait()
}

func sendInit() {
	initArgs := &V1InitSessionArgs{
		InitSession: true,
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "",
			Event: map[string]interface{}{
				utils.EVENT_NAME:  "TEST_EVENT",
				utils.ToR:         utils.VOICE,
				utils.Category:    "call",
				utils.Tenant:      "cgrates.org",
				utils.RequestType: utils.META_PREPAID,
				utils.AnswerTime:  time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
			},
		},
	}
	var wg sync.WaitGroup
	for i := 0; i < *initRuns; i++ {
		wg.Add(1)
		go func(i int) {
			initArgs.ID = utils.UUIDSha1Prefix()
			initArgs.Event[utils.OriginID] = utils.UUIDSha1Prefix()
			initArgs.Event[utils.Account] = fmt.Sprintf("1001%v", i)
			initArgs.Event[utils.Subject] = initArgs.Event[utils.Account]
			initArgs.Event[utils.Destination] = fmt.Sprintf("1002%v", i)

			var initRpl *V1InitSessionReply
			if err := sBenchRPC.Call(utils.SessionSv1InitiateSession,
				initArgs, &initRpl); err != nil {
				log.Fatal(err)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func getCount() int {
	var count int
	if err := sBenchRPC.Call(utils.SessionSv1GetActiveSessionsCount,
		map[string]string{}, &count); err != nil {
		log.Fatal(err)
	}
	return count
}

func BenchmarkSendInitSession(b *testing.B) {
	connOnce.Do(func() {
		startRPC()
		// addAccouns()
		sendInit()
		// time.Sleep(3 * time.Minute)
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getCount()
	}
}
