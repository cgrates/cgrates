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
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	stsV1CfgPath string
	stsV1Cfg     *config.CGRConfig
	stsV1Rpc     *rpc.Client
)

func TestStatSV1LoadConfig(t *testing.T) {
	var err error
	stsV1CfgPath = path.Join(*dataDir, "conf", "samples", "stats")
	if stsV1Cfg, err = config.NewCGRConfigFromFolder(stsV1CfgPath); err != nil {
		t.Error(err)
	}
}

func TestStatSV1InitDataDb(t *testing.T) {
	if err := engine.InitDataDb(stsV1Cfg); err != nil {
		t.Fatal(err)
	}
}

func TestStatSV1StartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(stsV1CfgPath, 1000); err != nil {
		t.Fatal(err)
	}
}

func TestStatSV1RpcConn(t *testing.T) {
	var err error
	stsV1Rpc, err = jsonrpc.Dial("tcp", stsV1Cfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func TestStatSV1TPFromFolder(t *testing.T) {
	var reply string
	time.Sleep(time.Duration(2000) * time.Millisecond)
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	if err := stsV1Rpc.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(1000) * time.Millisecond)
}

func TestStatSV1GetStats(t *testing.T) {
	var reply []string
	// first attempt should be empty since there is no queue in cache yet
	if err := stsV1Rpc.Call("StatSV1.GetQueueIDs", struct{}{}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	var metrics map[string]string
	if err := stsV1Rpc.Call("StatSV1.GetStatMetrics", "Stats1", &metrics); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	var replyStr string
	if err := stsV1Rpc.Call("StatSV1.LoadQueues", nil, &replyStr); err != nil {
		t.Error(err)
	} else if replyStr != utils.OK {
		t.Errorf("reply received: %s", replyStr)
	}
	expectedIDs := []string{"Stats1"}
	if err := stsV1Rpc.Call("StatSV1.GetQueueIDs", struct{}{}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedIDs, reply) {
		t.Errorf("expecting: %+v, received reply: %s", expectedIDs, reply)
	}
	expectedMetrics := map[string]string{}
	if err := stsV1Rpc.Call("StatSV1.GetStatMetrics", "Stats1", &metrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %s", expectedMetrics, metrics)
	}
}

func TestStatSV1ProcessEvent(t *testing.T) {
	var reply string
	if err := stsV1Rpc.Call("StatSV1.ProcessEvent",
		engine.StatsEvent{
			utils.ID:          "event1",
			utils.ANSWER_TIME: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("received reply: %s", reply)
	}
	if err := stsV1Rpc.Call("StatSV1.ProcessEvent",
		engine.StatsEvent{
			utils.ID:          "event2",
			utils.ANSWER_TIME: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("received reply: %s", reply)
	}
	if err := stsV1Rpc.Call("StatSV1.ProcessEvent",
		map[string]interface{}{
			utils.ID:         "event3",
			utils.SETUP_TIME: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("received reply: %s", reply)
	}
	expectedMetrics := map[string]string{}
	var metrics map[string]string
	if err := stsV1Rpc.Call("StatSV1.GetStatMetrics", "Stats1", &metrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %s", expectedMetrics, metrics)
	}
}

func TestStatSV1StopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
