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
package migrator

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestV2ActionTriggerAsThreshold(t *testing.T) {
	var filters []*engine.FilterRule
	v2ATR := &v2ActionTrigger{
		ID:             "test2",              // original csv tag
		UniqueID:       "testUUID",           // individual id
		ThresholdType:  "*min_event_counter", //*min_event_counter, *max_event_counter, *min_balance_counter, *max_balance_counter, *min_balance, *max_balance, *balance_expired
		ThresholdValue: 5.32,
		Recurrent:      false,                          // reset excuted flag each run
		MinSleep:       time.Duration(5) * time.Second, // Minimum duration between two executions in case of recurrent triggers
		ExpirationDate: time.Now(),
		ActivationDate: time.Now(),
		Balance: &engine.BalanceFilter{
			ID:             utils.StringPointer("test_balance"),
			DestinationIDs: &utils.StringMap{"1002": true},
			RatingSubject:  utils.StringPointer("1001"),
			Categories:     &utils.StringMap{utils.CALL: true},
			SharedGroups:   &utils.StringMap{"shared": true},
			TimingIDs:      &utils.StringMap{utils.ANY: true},
		},
		Weight:            0,
		ActionsID:         "Action1",
		MinQueuedItems:    10, // Trigger actions only if this number is hit (stats only)
		Executed:          false,
		LastExecutionTime: time.Now(),
	}
	x, _ := engine.NewFilterRule(utils.MetaDestinations, "DestinationIDs", v2ATR.Balance.DestinationIDs.Slice())
	filters = append(filters, x)
	x, _ = engine.NewFilterRule(utils.MetaPrefix, "RatingSubject", []string{*v2ATR.Balance.RatingSubject})
	filters = append(filters, x)
	x, _ = engine.NewFilterRule(utils.MetaPrefix, "Categories", v2ATR.Balance.Categories.Slice())
	filters = append(filters, x)
	x, _ = engine.NewFilterRule(utils.MetaPrefix, "SharedGroups", v2ATR.Balance.SharedGroups.Slice())
	filters = append(filters, x)
	x, _ = engine.NewFilterRule(utils.MetaPrefix, "TimingIDs", v2ATR.Balance.TimingIDs.Slice())
	filters = append(filters, x)

	filter := &engine.Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     *v2ATR.Balance.ID,
		Rules:  filters}

	thp := &engine.ThresholdProfile{
		ID:        v2ATR.ID,
		FilterIDs: []string{},
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		Blocker:   false,
		Weight:    v2ATR.Weight,
		ActivationInterval: &utils.ActivationInterval{
			ExpiryTime:     v2ATR.ExpirationDate,
			ActivationTime: v2ATR.ActivationDate,
		},
		MinSleep: v2ATR.MinSleep,
	}
	th := &engine.Threshold{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     v2ATR.ID,
	}

	newthp, newth, fltr, err := v2ATR.AsThreshold()
	if err != nil {
		t.Errorf("err")
	}
	if !reflect.DeepEqual(thp, newthp) {
		t.Errorf("expected: %+v,\nreceived: %+v", utils.ToJSON(thp), utils.ToJSON(newthp))
	}
	if !reflect.DeepEqual(th, newth) {
		t.Errorf("expected: %+v,\nreceived: %+v", th, newth)
	}
	if !reflect.DeepEqual(filter, fltr) {
		t.Errorf("expected: %+v,\nreceived: %+v", filter, fltr)
	}
}

func TestV2toV3Threshold(t *testing.T) {
	activationInterval := &utils.ActivationInterval{}
	v2ThresholdInstance := v2Threshold{
		Tenant:             "cgrates.org",
		ID:                 "id1",
		FilterIDs:          []string{"filter1", "filter2"},
		ActivationInterval: activationInterval,
		Recurrent:          true,
		MinHits:            10,
		MinSleep:           5 * time.Minute,
		Blocker:            true,
		Weight:             1.5,
		ActionIDs:          []string{"action1", "action2"},
		Async:              false,
	}
	expectedThresholdProfile := &engine.ThresholdProfile{
		Tenant:             "cgrates.org",
		ID:                 "id1",
		FilterIDs:          []string{"filter1", "filter2"},
		ActivationInterval: activationInterval,
		MinHits:            10,
		MinSleep:           5 * time.Minute,
		Blocker:            true,
		Weight:             1.5,
		ActionIDs:          []string{"action1", "action2"},
		Async:              false,
		MaxHits:            -1,
	}
	result := v2ThresholdInstance.V2toV3Threshold()

	if !reflect.DeepEqual(result, expectedThresholdProfile) {
		t.Errorf("V2toV3Threshold() = %v, want %v", result, expectedThresholdProfile)
	}
	v2ThresholdInstance.Recurrent = false
	expectedThresholdProfile.MaxHits = 1
	result = v2ThresholdInstance.V2toV3Threshold()
	if !reflect.DeepEqual(result, expectedThresholdProfile) {
		t.Errorf("V2toV3Threshold() = %v, want %v", result, expectedThresholdProfile)
	}
}

func TestAsSessionsCostSql(t *testing.T) {
	v2Cost := &v2SessionsCost{
		CGRID:       "cgrid1",
		RunID:       "runid1",
		OriginHost:  "originhost1",
		OriginID:    "originid1",
		CostSource:  "costsource1",
		CostDetails: nil,
		Usage:       5 * time.Second,
	}

	smSql := v2Cost.AsSessionsCostSql()
	if smSql.Cgrid != v2Cost.CGRID {
		t.Errorf("expected Cgrid %v, got %v", v2Cost.CGRID, smSql.Cgrid)
	}

	if smSql.RunID != v2Cost.RunID {
		t.Errorf("expected RunID %v, got %v", v2Cost.RunID, smSql.RunID)
	}

	if smSql.OriginHost != v2Cost.OriginHost {
		t.Errorf("expected OriginHost %v, got %v", v2Cost.OriginHost, smSql.OriginHost)
	}

	if smSql.OriginID != v2Cost.OriginID {
		t.Errorf("expected OriginID %v, got %v", v2Cost.OriginID, smSql.OriginID)
	}

	if smSql.CostSource != v2Cost.CostSource {
		t.Errorf("expected CostSource %v, got %v", v2Cost.CostSource, smSql.CostSource)
	}

	expectedCostDetails := utils.ToJSON(v2Cost.CostDetails)
	if smSql.CostDetails != expectedCostDetails {
		t.Errorf("expected CostDetails %v, got %v", expectedCostDetails, smSql.CostDetails)
	}

	expectedUsage := v2Cost.Usage.Nanoseconds()
	if smSql.Usage != expectedUsage {
		t.Errorf("expected Usage %v, got %v", expectedUsage, smSql.Usage)
	}

	if smSql.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set, got zero value")
	}
}

func TestNewV2SessionsCostFromSessionsCostSql(t *testing.T) {
	smSql := &engine.SessionCostsSQL{
		Cgrid:       "cgrid1",
		RunID:       "runid1",
		OriginHost:  "originhost1",
		OriginID:    "originid1",
		CostSource:  "costsource1",
		CostDetails: `{"detail1": "value1", "detail2": "value2"}`,
		Usage:       5 * time.Second.Nanoseconds(),
		CreatedAt:   time.Now(),
	}

	smV2, err := NewV2SessionsCostFromSessionsCostSql(smSql)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if smV2.CGRID != smSql.Cgrid {
		t.Errorf("expected CGRID %v, got %v", smSql.Cgrid, smV2.CGRID)
	}

	if smV2.RunID != smSql.RunID {
		t.Errorf("expected RunID %v, got %v", smSql.RunID, smV2.RunID)
	}

	if smV2.OriginHost != smSql.OriginHost {
		t.Errorf("expected OriginHost %v, got %v", smSql.OriginHost, smV2.OriginHost)
	}

	if smV2.OriginID != smSql.OriginID {
		t.Errorf("expected OriginID %v, got %v", smSql.OriginID, smV2.OriginID)
	}

	if smV2.CostSource != smSql.CostSource {
		t.Errorf("expected CostSource %v, got %v", smSql.CostSource, smV2.CostSource)
	}

	expectedUsage := time.Duration(smSql.Usage)
	if smV2.Usage != expectedUsage {
		t.Errorf("expected Usage %v, got %v", expectedUsage, smV2.Usage)
	}

	var expectedCostDetails map[string]interface{}
	if err := json.Unmarshal([]byte(smSql.CostDetails), &expectedCostDetails); err != nil {
		t.Fatalf("failed to unmarshal CostDetails: %v", err)
	}

}
