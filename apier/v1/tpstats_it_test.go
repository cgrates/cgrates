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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"reflect"
	"testing"
)

var tpCfgPath string
var tpCfg *config.CGRConfig
var tpRPC *rpc.Client

func TestTPStatInitCfg(t *testing.T) {
	var err error
	tpCfgPath = path.Join(*dataDir, "conf", "samples", "tutmysql")
	tpCfg, err = config.NewCGRConfigFromFolder(tpCfgPath)
	if err != nil {
		t.Error(err)
	}
	tpCfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(tpCfg)
}

// Wipe out the cdr database
func TestTPStatResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tpCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine

func TestTPStatStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tpCfgPath, 1000); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestTPStatRpcConn(t *testing.T) {
	var err error
	tpRPC, err = jsonrpc.Dial("tcp", tpCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

var tpStat = &utils.TPStats{
	TPid: "TPS1",
	ID:   "Stat1",
	Filters: []*utils.TPRequestFilter{
		&utils.TPRequestFilter{
			Type:      "*string",
			FieldName: "Account",
			Values:    []string{"1001", "1002"},
		},
		&utils.TPRequestFilter{
			Type:      "*string_prefix",
			FieldName: "Destination",
			Values:    []string{"10", "20"},
		},
	},
	ActivationInterval: &utils.TPActivationInterval{
		ActivationTime: "2014-07-29T15:00:00Z",
		ExpiryTime:     "",
	},
	TTL:        "1",
	Metrics:    []string{"MetricValue", "MetricValueTwo"},
	Blocker:    true,
	Stored:     true,
	Weight:     20,
	Thresholds: nil,
}

func TestTPStatGetTPStatIDs(t *testing.T) {
	var reply []string
	if err := tpRPC.Call("ApierV1.GetTPStatIDs", AttrGetTPStatIds{TPid: "TPS1"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func TestTPStatSetTPStat(t *testing.T) {
	var result string
	if err := tpRPC.Call("ApierV1.SetTPStat", tpStat, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func TestTPStatGetTPStat(t *testing.T) {
	var respond *utils.TPStats
	if err := tpRPC.Call("ApierV1.GetTPStat", &AttrGetTPStat{TPid: tpStat.TPid, ID: tpStat.ID}, &respond); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpStat, respond) {
		t.Errorf("Expecting: %+v, received: %+v", tpStat.TPid, respond.TPid)
	}
}

func TestTPStatUpdateTPStat(t *testing.T) {
	var result string
	tpStat.Weight = 21
	tpStat.Filters = []*utils.TPRequestFilter{
		&utils.TPRequestFilter{
			Type:      "*string",
			FieldName: "Account",
			Values:    []string{"1001", "1002"},
		},
		&utils.TPRequestFilter{
			Type:      "*string_prefix",
			FieldName: "Destination",
			Values:    []string{"10", "20"},
		},
		&utils.TPRequestFilter{
			Type:      "*rsr_fields",
			FieldName: "",
			Values:    []string{"Subject(~^1.*1$)", "Destination(1002)"},
		},
	}
	if err := tpRPC.Call("ApierV1.SetTPStat", tpStat, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var expectedTPS *utils.TPStats
	if err := tpRPC.Call("ApierV1.GetTPStat", &AttrGetTPStat{TPid: tpStat.TPid, ID: tpStat.ID}, &expectedTPS); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpStat, expectedTPS) {
		t.Errorf("Expecting: %+v, received: %+v", tpStat, expectedTPS)
	}
}

func TestTPStatRemTPStat(t *testing.T) {
	var resp string
	if err := tpRPC.Call("ApierV1.RemTPStat", &AttrGetTPStat{TPid: tpStat.TPid, ID: tpStat.ID}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
}

func TestTPStatCheckDelete(t *testing.T) {
	var respond *utils.TPStats
	if err := tpRPC.Call("ApierV1.GetTPStat", &AttrGetTPStat{TPid: "TPS1", ID: "Stat1"}, &respond); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func TestTPStatKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
