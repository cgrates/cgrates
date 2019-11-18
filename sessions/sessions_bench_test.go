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
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	sBenchCfg *config.CGRConfig
	sBenchRPC *rpc.Client
	connOnce  sync.Once
	initRuns  = flag.Int("runs", 25000, "number of loops to run in init")
	cps       = flag.Int("cps", 2000, "number of loops to run in init")
	maxCps    = make(chan struct{}, *cps)
)

func startRPC() {
	var err error
	sBenchCfg, err = config.NewCGRConfigFromPath(
		path.Join(config.CgrConfig().DataFolderPath, "conf", "samples", "tutmongo"))
	if err != nil {
		log.Fatal(err)
	}
	config.SetCgrConfig(sBenchCfg)
	if sBenchRPC, err = jsonrpc.Dial("tcp", sBenchCfg.ListenCfg().RPCJSONListen); err != nil {
		log.Fatalf("Error at dialing rcp client:%v\n", err)
	}
}

func loadTP() {
	for i := 0; i < *cps; i++ { // init CPS limitation
		maxCps <- struct{}{}
	}
	if err := engine.InitDataDb(sBenchCfg); err != nil {
		log.Fatal(err)
	}
	attrs := &utils.AttrLoadTpFromFolder{
		FolderPath: path.Join(config.CgrConfig().DataFolderPath, "tariffplans", "tutorial")}
	var tpLoadInst utils.LoadInstance
	if err := sBenchRPC.Call("ApierV2.LoadTariffPlanFromFolder",
		attrs, &tpLoadInst); err != nil {
		log.Fatal(err)
	}
	time.Sleep(time.Duration(100) * time.Millisecond) // Give time for scheduler to execute topups
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
	if err := sBenchRPC.Call(utils.ApierV2SetBalance,
		attrSetBalance, &reply); err != nil {
		log.Fatal(err)
	}
}

func addAccouns() {
	var wg sync.WaitGroup
	for i := 0; i < *initRuns; i++ {
		wg.Add(1)
		go func(i int) {
			oneCps := <-maxCps // queue here for maxCps
			defer func() { maxCps <- oneCps }()
			addBalance(sBenchRPC, fmt.Sprintf("1001%v", i))
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func initSession(i int) {
	oneCps := <-maxCps // queue here for maxCps
	defer func() { maxCps <- oneCps }()
	initArgs := &V1InitSessionArgs{
		InitSession: true,
		CGREvent: &utils.CGREvent{
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
	initArgs.ID = utils.UUIDSha1Prefix()
	initArgs.Event[utils.OriginID] = utils.UUIDSha1Prefix()
	initArgs.Event[utils.Account] = fmt.Sprintf("1001%v", i)
	initArgs.Event[utils.Subject] = "1001" //initArgs.Event[utils.Account]
	initArgs.Event[utils.Destination] = fmt.Sprintf("1002%v", i)

	var initRpl *V1InitSessionReply
	if err := sBenchRPC.Call(utils.SessionSv1InitiateSession,
		initArgs, &initRpl); err != nil {
		// log.Fatal(err)
	}
}

func sendInitx10(r int) {
	var wg sync.WaitGroup
	for i := r; i < r+10; i++ {
		wg.Add(1)
		go func(i int) {
			initSession(i)
			wg.Done()

		}(i)
	}
	wg.Wait()
}

func sendInit() {
	var wg sync.WaitGroup
	for i := 0; i < *initRuns; i++ {
		wg.Add(1)
		go func(i int) {
			initSession(i)
			wg.Done()

		}(i)
	}
	wg.Wait()
}

func getCount() int {
	var count int
	if err := sBenchRPC.Call(utils.SessionSv1GetActiveSessionsCount, utils.SessionFilter{
		Filters: []string{"*string:~ToR:*voice"},
	}, &count); err != nil {
		log.Fatal(err)
	}
	return count
}

func BenchmarkSendInitSession(b *testing.B) {
	connOnce.Do(func() {
		startRPC()
		loadTP()
		addAccouns()
		sendInit()
		// time.Sleep(3 * time.Minute)
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getCount()
	}
}

func BenchmarkEncodingJSON(b *testing.B) {
	maxCps = make(chan struct{}, *cps)
	for i := 0; i < *cps; i++ { // init CPS limitation
		maxCps <- struct{}{}
	}
	var err error
	sBenchCfg, err = config.NewCGRConfigFromPath(
		path.Join(config.CgrConfig().DataFolderPath, "conf", "samples", "tutmongo"))
	if err != nil {
		log.Fatal(err)
	}

	if sBenchRPC, err = jsonrpc.Dial("tcp", sBenchCfg.ListenCfg().RPCJSONListen); err != nil {
		log.Fatalf("Error at dialing rcp client:%v\n", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		initSession(i)
	}

}

func BenchmarkEncodingGOB(b *testing.B) {
	maxCps = make(chan struct{}, *cps)
	for i := 0; i < *cps; i++ { // init CPS limitation
		maxCps <- struct{}{}
	}
	var err error
	sBenchCfg, err = config.NewCGRConfigFromPath(
		path.Join(config.CgrConfig().DataFolderPath, "conf", "samples", "tutmongo"))
	if err != nil {
		log.Fatal(err)
	}

	if sBenchRPC, err = rpc.Dial("tcp", sBenchCfg.ListenCfg().RPCGOBListen); err != nil {
		log.Fatalf("Error at dialing rcp client:%v\n", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		initSession(i)
	}

}

func benchmarkSendInitSessionx10(b *testing.B) {
	connOnce.Do(func() {
		startRPC()
		loadTP()
		addAccouns()
		// time.Sleep(3 * time.Minute)
	})
	b.ResetTimer()
	for i := 0; i < *initRuns/10; i++ {
		sendInitx10(i * 10)
		tStart := time.Now()
		_ = getCount()
		if tDur := time.Now().Sub(tStart); tDur > 100*time.Millisecond && tDur < time.Second {
			fmt.Printf("Expected answer in less than %v receved answer after %v for %v sessions\n", 100*time.Millisecond, tDur, i*10+10)
		} else if tDur >= time.Second {
			b.Fatalf("Fatal:Expected answer in less than %v receved answer after %v for %v sessions", time.Second, tDur, i*10+10)
		}
	}
}
