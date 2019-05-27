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

package dispatchers

import (
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var sTestsDspTh = []func(t *testing.T){
	testDspThPingFailover,
	testDspThProcessEventFailover,

	testDspThPing,
	testDspThTestAuthKey,
	testDspThTestAuthKey2,
	testDspThTestAuthKey3,
}

//Test start here
func TestDspThresholdSTMySQL(t *testing.T) {
	testDsp(t, sTestsDspTh, "TestDspThresholdS", "all", "all2", "dispatchers", "tutorial", "oldtutorial", "dispatchers")
}

func TestDspThresholdSMongo(t *testing.T) {
	testDsp(t, sTestsDspTh, "TestDspThresholdS", "all", "all2", "dispatchers_mongo", "tutorial", "oldtutorial", "dispatchers")
}

func testDspThPingFailover(t *testing.T) {
	var reply string
	if err := allEngine.RCP.Call(utils.ThresholdSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	ev := utils.CGREventWithArgDispatcher{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("thr12345"),
		},
	}
	if err := dispEngine.RCP.Call(utils.ThresholdSv1Ping, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	allEngine.stopEngine(t)
	if err := dispEngine.RCP.Call(utils.ThresholdSv1Ping, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	allEngine2.stopEngine(t)
	if err := dispEngine.RCP.Call(utils.ThresholdSv1Ping, &ev, &reply); err == nil {
		t.Errorf("Expected error but recived %v and reply %v\n", err, reply)
	}
	allEngine.startEngine(t)
	allEngine2.startEngine(t)
}

func testDspThProcessEventFailover(t *testing.T) {
	var ids []string
	eIDs := []string{"THD_ACNT_1001"}
	nowTime := time.Now()
	args := &engine.ArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Time:   &nowTime,
			Event: map[string]interface{}{
				utils.EVENT_NAME: "Event1",
				utils.Account:    "1001"},
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("thr12345"),
		},
	}

	if err := dispEngine.RCP.Call(utils.ThresholdSv1ProcessEvent, args,
		&ids); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected error NOT_FOUND but recived %v and reply %v\n", err, ids)
	}
	allEngine2.stopEngine(t)
	if err := dispEngine.RCP.Call(utils.ThresholdSv1ProcessEvent, args, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIDs, ids) {
		t.Errorf("expecting: %+v, received: %+v", eIDs, ids)
	}
	allEngine2.startEngine(t)
}

func testDspThPing(t *testing.T) {
	var reply string
	if err := allEngine.RCP.Call(utils.ThresholdSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	if err := dispEngine.RCP.Call(utils.ThresholdSv1Ping, &utils.CGREventWithArgDispatcher{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("thr12345"),
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
}

func testDspThTestAuthKey(t *testing.T) {
	var ids []string
	nowTime := time.Now()
	args := &engine.ArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Time:   &nowTime,
			Event: map[string]interface{}{
				utils.Account: "1002"},
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("12345"),
		},
	}

	if err := dispEngine.RCP.Call(utils.ThresholdSv1ProcessEvent,
		args, &ids); err == nil || err.Error() != utils.ErrUnauthorizedApi.Error() {
		t.Error(err)
	}
	var th *engine.Thresholds
	if err := dispEngine.RCP.Call(utils.ThresholdSv1GetThresholdsForEvent, args,
		&th); err == nil || err.Error() != utils.ErrUnauthorizedApi.Error() {
		t.Error(err)
	}
}

func testDspThTestAuthKey2(t *testing.T) {
	var ids []string
	eIDs := []string{"THD_ACNT_1002"}
	nowTime := time.Now()
	args := &engine.ArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Time:   &nowTime,
			Event: map[string]interface{}{
				utils.Account: "1002"},
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("thr12345"),
		},
	}

	if err := dispEngine.RCP.Call(utils.ThresholdSv1ProcessEvent, args, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIDs, ids) {
		t.Errorf("expecting: %+v, received: %+v", eIDs, ids)
	}
	var th *engine.Thresholds
	eTh := &engine.Thresholds{
		&engine.Threshold{
			Tenant: "cgrates.org",
			ID:     "THD_ACNT_1002",
			Hits:   1,
		},
	}
	if err := dispEngine.RCP.Call(utils.ThresholdSv1GetThresholdsForEvent, args, &th); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual((*eTh)[0].Tenant, (*th)[0].Tenant) {
		t.Errorf("expecting: %+v, received: %+v", (*eTh)[0].Tenant, (*th)[0].Tenant)
	} else if !reflect.DeepEqual((*eTh)[0].ID, (*th)[0].ID) {
		t.Errorf("expecting: %+v, received: %+v", (*eTh)[0].ID, (*th)[0].ID)
	} else if !reflect.DeepEqual((*eTh)[0].Hits, (*th)[0].Hits) {
		t.Errorf("expecting: %+v, received: %+v", (*eTh)[0].Hits, (*th)[0].Hits)
	}
}

func testDspThTestAuthKey3(t *testing.T) {
	var th *engine.Threshold
	eTh := &engine.Threshold{
		Tenant: "cgrates.org",
		ID:     "THD_ACNT_1002",
		Hits:   1,
	}
	if err := dispEngine.RCP.Call(utils.ThresholdSv1GetThreshold, &utils.TenantIDWithArgDispatcher{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "THD_ACNT_1002",
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("thr12345"),
		},
	}, &th); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual((*eTh).Tenant, (*th).Tenant) {
		t.Errorf("expecting: %+v, received: %+v", (*eTh).Tenant, (*th).Tenant)
	} else if !reflect.DeepEqual((*eTh).ID, (*th).ID) {
		t.Errorf("expecting: %+v, received: %+v", (*eTh).ID, (*th).ID)
	} else if !reflect.DeepEqual((*eTh).Hits, (*th).Hits) {
		t.Errorf("expecting: %+v, received: %+v", (*eTh).Hits, (*th).Hits)
	}

	var ids []string
	eIDs := []string{"THD_ACNT_1002"}

	if err := dispEngine.RCP.Call(utils.ThresholdSv1GetThresholdIDs, &utils.TenantWithArgDispatcher{
		TenantArg: &utils.TenantArg{
			Tenant: "cgrates.org",
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("thr12345"),
		},
	}, &ids); err != nil {
		t.Fatal(err)
	}
	sort.Strings(ids)
	if !reflect.DeepEqual(eIDs, ids) {
		t.Errorf("expecting: %+v, received: %+v", eIDs, ids)
	}
}
