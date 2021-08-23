//go:build integration
// +build integration

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

package v1

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/rpc"
	"path"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cenkalti/rpc2"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	concReqsCfgPath   string
	concReqsCfg       *config.CGRConfig
	concReqsRPC       *rpc.Client
	concReqsBiRPC     *rpc2.Client
	concReqsConfigDIR string //run tests for specific configuration

	sTestsConcReqs = []func(t *testing.T){
		testConcReqsInitCfg,
		testConcReqsStartEngine,
		testConcReqsRPCConn,
		testConcReqsBusyAPIs,
		testConcReqsQueueAPIs,
		testConcReqsOnHTTPBusy,
		testConcReqsOnHTTPQueue,
		testConcReqsOnBiJSONBusy,
		testConcReqsOnBiJSONQueue,
		testConcReqsKillEngine,
	}
)

//Test start here
func TestConcReqsBusyJSON(t *testing.T) {
	concReqsConfigDIR = "conc_reqs_busy"
	for _, stest := range sTestsConcReqs {
		t.Run(concReqsConfigDIR, stest)
	}
}

func TestConcReqsQueueJSON(t *testing.T) {
	concReqsConfigDIR = "conc_reqs_queue"
	for _, stest := range sTestsConcReqs {
		t.Run(concReqsConfigDIR, stest)
	}
}

func TestConcReqsBusyGOB(t *testing.T) {
	concReqsConfigDIR = "conc_reqs_busy"
	encoding = utils.StringPointer(utils.MetaGOB)
	for _, stest := range sTestsConcReqs {
		t.Run(concReqsConfigDIR, stest)
	}
}

func TestConcReqsQueueGOB(t *testing.T) {
	concReqsConfigDIR = "conc_reqs_queue"
	encoding = utils.StringPointer(utils.MetaGOB)
	for _, stest := range sTestsConcReqs {
		t.Run(concReqsConfigDIR, stest)
	}
}

func testConcReqsInitCfg(t *testing.T) {
	var err error
	concReqsCfgPath = path.Join(*dataDir, "conf", "samples", concReqsConfigDIR)
	concReqsCfg, err = config.NewCGRConfigFromPath(concReqsCfgPath)
	if err != nil {
		t.Error(err)
	}
	concReqsCfg.DataFolderPath = *dataDir
	config.SetCgrConfig(concReqsCfg)
}

// Start CGR Engine
func testConcReqsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(concReqsCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func handlePing(clnt *rpc2.Client, arg *DurationArgs, reply *string) error {
	time.Sleep(arg.DurationTime)
	*reply = utils.OK
	return nil
}

// Connect rpc client to rater
func testConcReqsRPCConn(t *testing.T) {
	var err error
	concReqsRPC, err = newRPCClient(concReqsCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
	if concReqsBiRPC, err = utils.NewBiJSONrpcClient(concReqsCfg.SessionSCfg().ListenBijson,
		nil); err != nil {
		t.Fatal(err)
	}
}

func testConcReqsBusyAPIs(t *testing.T) {
	if concReqsConfigDIR != "conc_reqs_busy" {
		t.SkipNow()
	}
	var failedAPIs int
	wg := new(sync.WaitGroup)
	lock := new(sync.Mutex)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			var resp string
			if err := concReqsRPC.Call(utils.CoreSv1Sleep,
				&DurationArgs{DurationTime: time.Duration(10 * time.Millisecond)},
				&resp); err != nil {
				lock.Lock()
				failedAPIs++
				lock.Unlock()
				wg.Done()
				return
			}
			wg.Done()
		}()
	}
	wg.Wait()
	if failedAPIs < 2 {
		t.Errorf("Expected at leat 2 APIs to wait")
	}
}

func testConcReqsQueueAPIs(t *testing.T) {
	if concReqsConfigDIR != "conc_reqs_queue" {
		t.SkipNow()
	}
	wg := new(sync.WaitGroup)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			var resp string
			if err := concReqsRPC.Call(utils.CoreSv1Sleep,
				&DurationArgs{DurationTime: time.Duration(10 * time.Millisecond)},
				&resp); err != nil {
				wg.Done()
				t.Error(err)
				return
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func testConcReqsOnHTTPBusy(t *testing.T) {
	if concReqsConfigDIR != "conc_reqs_busy" {
		t.SkipNow()
	}
	var fldAPIs int64
	wg := new(sync.WaitGroup)
	lock := new(sync.Mutex)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(index int) {
			resp, err := http.Post("http://localhost:2080/jsonrpc", "application/json", bytes.NewBuffer([]byte(fmt.Sprintf(`{"method": "CoreSv1.Sleep", "params": [{"DurationTime":10000000}], "id":%d}`, index))))
			if err != nil {
				wg.Done()
				t.Error(err)
				return
			}
			contents, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				wg.Done()
				t.Error(err)
				return
			}
			resp.Body.Close()
			if strings.Contains(string(contents), "denying request due to maximum active requests reached") {
				lock.Lock()
				fldAPIs++
				lock.Unlock()
			}
			wg.Done()
			return
		}(i)
	}
	wg.Wait()
	if fldAPIs < 2 {
		t.Errorf("Expected at leat 2 APIs to wait")
	}
}

func testConcReqsOnHTTPQueue(t *testing.T) {
	if concReqsConfigDIR != "conc_reqs_queue" {
		t.SkipNow()
	}
	wg := new(sync.WaitGroup)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(index int) {
			_, err := http.Post("http://localhost:2080/jsonrpc", "application/json", bytes.NewBuffer([]byte(fmt.Sprintf(`{"method": "CoreSv1.Sleep", "params": [{"DurationTime":10000000}], "id":%d}`, index))))
			if err != nil {
				wg.Done()
				t.Error(err)
				return
			}
			wg.Done()
			return
		}(i)
	}
	wg.Wait()
}

func testConcReqsOnBiJSONBusy(t *testing.T) {
	if concReqsConfigDIR != "conc_reqs_busy" {
		t.SkipNow()
	}
	var failedAPIs int
	wg := new(sync.WaitGroup)
	lock := new(sync.Mutex)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			var resp string
			if err := concReqsBiRPC.Call(utils.SessionSv1Sleep,
				&DurationArgs{DurationTime: time.Duration(10 * time.Millisecond)},
				&resp); err != nil {
				lock.Lock()
				failedAPIs++
				lock.Unlock()
				wg.Done()
				return
			}
			wg.Done()
		}()
	}
	wg.Wait()
	if failedAPIs < 2 {
		t.Errorf("Expected at leat 2 APIs to wait")
	}
}

func testConcReqsOnBiJSONQueue(t *testing.T) {
	if concReqsConfigDIR != "conc_reqs_queue" {
		t.SkipNow()
	}
	wg := new(sync.WaitGroup)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			var resp string
			if err := concReqsBiRPC.Call(utils.SessionSv1Sleep,
				&DurationArgs{DurationTime: time.Duration(10 * time.Millisecond)},
				&resp); err != nil {
				wg.Done()
				t.Error(err)
				return
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func testConcReqsKillEngine(t *testing.T) {
	time.Sleep(100 * time.Millisecond)
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
