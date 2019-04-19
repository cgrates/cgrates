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
	SessionBenchmarkCfg *config.CGRConfig
	SessionBenchmarkRPC *rpc.Client
	ConnectOnce         sync.Once
	NoSessions          int
)

func startRPC() {
	var err error
	SessionBenchmarkCfg, err = config.NewCGRConfigFromPath(path.Join(config.CgrConfig().DataFolderPath, "conf", "samples", "tutmysql"))
	if err != nil {
		log.Fatal(err)
	}
	config.SetCgrConfig(SessionBenchmarkCfg)
	if SessionBenchmarkRPC, err = jsonrpc.Dial("tcp", SessionBenchmarkCfg.ListenCfg().RPCJSONListen); err != nil {
		log.Fatalf("Error at dialing rcp client:%v\n", err)
	}
}

func addBalance(SessionBenchmarkRPC *rpc.Client, sraccount string) {
	attrSetBalance := utils.AttrSetBalance{
		Tenant:        "cgrates.org",
		Account:       sraccount,
		BalanceType:   utils.VOICE,
		BalanceID:     utils.StringPointer("TestDynamicDebitBalance"),
		Value:         utils.Float64Pointer(5 * float64(time.Hour)),
		RatingSubject: utils.StringPointer("*zero5ms"),
	}
	var reply string
	if err := SessionBenchmarkRPC.Call("ApierV2.SetBalance", attrSetBalance, &reply); err != nil {
		log.Fatal(err)
		// } else if reply != utils.OK {
		// log.Fatalf("Received: %s", reply)
	}
}

func addAccouns() {
	var wg sync.WaitGroup
	for i := 0; i < 23000; i++ {
		wg.Add(1)
		go func(i int, SessionBenchmarkRPC *rpc.Client) {
			addBalance(SessionBenchmarkRPC, fmt.Sprintf("1001%v1002", i))
			addBalance(SessionBenchmarkRPC, fmt.Sprintf("1001%v1001", i))
			wg.Done()
		}(i, SessionBenchmarkRPC)
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
				utils.OriginID:    "123491",
				utils.Account:     "1001",
				utils.Subject:     "1001",
				utils.Destination: "1002",
				utils.Category:    "call",
				utils.Tenant:      "cgrates.org",
				utils.RequestType: utils.META_PREPAID,
				utils.SetupTime:   time.Date(2016, time.January, 5, 18, 30, 59, 0, time.UTC),
				utils.AnswerTime:  time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
				utils.Usage:       "10", // 5MB
			},
		},
	}
	// var wg sync.WaitGroup
	for i := 0; i < 23000; i++ {
		// wg.Add(1)
		// go func(i int, SessionBenchmarkRPC *rpc.Client) {
		initArgs.ID = utils.UUIDSha1Prefix()
		initArgs.Event[utils.OriginID] = utils.UUIDSha1Prefix()
		initArgs.Event[utils.Account] = fmt.Sprintf("1001%v1002", i)
		initArgs.Event[utils.Subject] = initArgs.Event[utils.Account]
		initArgs.Event[utils.Destination] = fmt.Sprintf("1001%v1001", i)

		var initRpl *V1InitSessionReply
		if err := SessionBenchmarkRPC.Call(utils.SessionSv1InitiateSession,
			initArgs, &initRpl); err != nil {
			log.Fatal(err)
		}
		// _ = getCount(SessionBenchmarkRPC)
		// if c := getCount(SessionBenchmarkRPC); i+1 != c {
		// 	log.Fatalf("Not Enough sessions %v!=%v", i+1, c)
		// }
		// wg.Done()
		// }(i, SessionBenchmarkRPC)
	}
	// wg.Wait()
}

func getCount(SessionBenchmarkRPC *rpc.Client) int {
	var count int
	if err := SessionBenchmarkRPC.Call(utils.SessionSv1GetActiveSessionsCount,
		map[string]string{}, &count); err != nil {
		log.Fatal(err)
	}
	return count
}

func BenchmarkSendInitSession(b *testing.B) {
	ConnectOnce.Do(func() {
		startRPC()
		// addAccouns()
		sendInit()
		// time.Sleep(3 * time.Minute)
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getCount(SessionBenchmarkRPC)
		// if count < 2000 {
		// 	b.Fatal("Not Enough sessions")
		// }
	}
}
