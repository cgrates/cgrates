/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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

package actions

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestMatchingActionProfilesForEvent(t *testing.T) {
	defaultCfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filters := engine.NewFilterS(defaultCfg, nil, dm)
	acts := NewActionS(defaultCfg, filters, dm)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TEST_ACTIONS1",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
			utils.Destination:  1002,
		},
	}

	actPrf := &engine.ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "test_id1",
		FilterIDs: []string{"*string:~*req.Account:1001;1002;1003", "*prefix:~*req.Destination:10"},
		Actions: []*engine.APAction{
			{
				ID:        "TOPUP",
				FilterIDs: []string{},
				Type:      "*topup",
				Path:      "~*balance.TestBalance.Value",
				Value:     config.NewRSRParsersMustCompile("10", defaultCfg.GeneralCfg().RSRSep),
			},
		},
	}

	if err := acts.dm.SetActionProfile(actPrf, true); err != nil {
		t.Error(err)
	}

	expActionPrf := engine.ActionProfiles{actPrf}

	if rcv, err := acts.matchingActionProfilesForEvent("cgrates.org", cgrEv, []string{}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, expActionPrf) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expActionPrf), utils.ToJSON(rcv))
	}

	cgrEv = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TEST_ACTIONS1",
		Event: map[string]interface{}{
			utils.Accounts: "10",
		},
	}
	//This Event is not matching with our filter
	if _, err := acts.matchingActionProfilesForEvent("cgrates.org", cgrEv, []string{}); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}

	cgrEv.Event[utils.AccountField] = "1001"
	actPrfIDs := []string{"inexisting_id"}
	//Unable to get from database an ActionProfile if the ID won't match
	if _, err := acts.matchingActionProfilesForEvent("cgrates.org", cgrEv, actPrfIDs); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}

	actPrfIDs = []string{"test_id1"}
	if _, err := acts.matchingActionProfilesForEvent("cgrates.org", cgrEv, actPrfIDs); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}

	actPrf.ActivationInterval = &utils.ActivationInterval{
		ActivationTime: time.Date(2012, 7, 21, 0, 0, 0, 0, time.UTC),
		ExpiryTime:     time.Date(2012, 8, 21, 0, 0, 0, 0, time.UTC),
	}
	//this event is not active in this interval time
	cgrEv.Time = utils.TimePointer(time.Date(2012, 6, 21, 0, 0, 0, 0, time.UTC))
	if _, err := acts.matchingActionProfilesForEvent("cgrates.org", cgrEv, actPrfIDs); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}
	actPrf.ActivationInterval = nil

	//when dataManager is nil, it won't be able to get ActionsProfile from database
	acts.dm = nil
	if _, err := acts.matchingActionProfilesForEvent("INVALID_TENANT", cgrEv, actPrfIDs); err == nil || err != utils.ErrNoDatabaseConn {
		t.Errorf("Expected %+v, received %+v", utils.ErrNoDatabaseConn, err)
	}

	acts.dm = engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	actPrf.FilterIDs = []string{"invalid_filters"}
	//Set in database and invalid filter, so it won t pass
	if err := acts.dm.SetActionProfile(actPrf, false); err != nil {
		t.Error(err)
	}
	expected := "NOT_FOUND:invalid_filters"
	if _, err := acts.matchingActionProfilesForEvent("cgrates.org", cgrEv, actPrfIDs); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	if err := acts.dm.RemoveActionProfile(actPrf.Tenant, actPrf.ID, utils.NonTransactional, false); err != nil {
		t.Error(err)
	}
}

func TestScheduledActions(t *testing.T) {
	defaultCfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filters := engine.NewFilterS(defaultCfg, nil, dm)
	acts := NewActionS(defaultCfg, filters, dm)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TEST_ACTIONS1",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
			utils.Destination:  1002,
		},
	}

	actPrf := &engine.ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "test_id1",
		FilterIDs: []string{"*string:~*req.Account:1001;1002;1003", "*prefix:~*req.Destination:10"},
		Actions: []*engine.APAction{
			{
				ID:        "TOPUP",
				FilterIDs: []string{},
				Type:      utils.MetaLog,
				Path:      "~*balance.TestBalance.Value",
				Value:     config.NewRSRParsersMustCompile("10", defaultCfg.GeneralCfg().RSRSep),
			},
		},
	}

	if err := acts.dm.SetActionProfile(actPrf, true); err != nil {
		t.Error(err)
	}

	if rcv, err := acts.scheduledActions(cgrEv.Tenant, cgrEv, []string{}, false); err != nil {
		t.Error(err)
	} else {
		expSchedActs := newScheduledActs(cgrEv.Tenant, cgrEv.ID, utils.MetaNone, utils.EmptyString,
			utils.EmptyString, context.Background(), &ActData{cgrEv.Event, cgrEv.Opts}, rcv[0].acts)
		if reflect.DeepEqual(expSchedActs, rcv) {
			t.Errorf("Expected %+v, received %+v", expSchedActs, rcv)
		}
	}

	cgrEv = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TEST_ACTIONS1",
		Event: map[string]interface{}{
			utils.Accounts: "10",
		},
	}
	if _, err := acts.scheduledActions(cgrEv.Tenant, cgrEv, []string{}, false); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}

	cgrEv = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "test_id1",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
			utils.Destination:  1002,
		},
	}
	actPrf.Actions[0].Type = "*topup"
	if err := acts.dm.SetActionProfile(actPrf, true); err != nil {
		t.Error(err)
	}
	if _, err := acts.scheduledActions(cgrEv.Tenant, cgrEv, []string{}, false); err == nil || err != utils.ErrPartiallyExecuted {
		t.Errorf("Expected %+v, received %+v", utils.ErrPartiallyExecuted, err)
	}
}

func TestScheduleAction(t *testing.T) {
	defaultCfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filters := engine.NewFilterS(defaultCfg, nil, dm)
	acts := NewActionS(defaultCfg, filters, dm)

	cgrEv := []*utils.CGREvent{
		{
			Tenant: "cgrates.org",
			ID:     "TEST_ACTIONS1",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
				utils.Destination:  1002,
			},
		},
	}

	actPrf := &engine.ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "test_id1",
		FilterIDs: []string{"*string:~*req.Account:1001;1002;1003", "*prefix:~*req.Destination:10"},
		Schedule:  "* * * * *",
		Actions: []*engine.APAction{
			{
				ID:        "TOPUP",
				FilterIDs: []string{},
				Type:      utils.MetaLog,
				Path:      "~*balance.TestBalance.Value",
				Value:     config.NewRSRParsersMustCompile("10", defaultCfg.GeneralCfg().RSRSep),
			},
		},
	}
	if err := acts.dm.SetActionProfile(actPrf, true); err != nil {
		t.Error(err)
	}

	if err := acts.scheduleActions(cgrEv, []string{}, true); err != nil {
		t.Error(err)
	}

	//Cannot schedule an action if the ID is invalid
	if err := acts.scheduleActions(cgrEv, []string{"INVALID_ID1"}, true); err == nil || err != utils.ErrPartiallyExecuted {
		t.Errorf("Expected %+v, received %+v", utils.ErrPartiallyExecuted, err)
	}

	//When schedule is "*asap", the action will execute immediately
	actPrf.Schedule = utils.MetaASAP
	if err := acts.dm.SetActionProfile(actPrf, true); err != nil {
		t.Error(err)
	}
	if err := acts.scheduleActions(cgrEv, []string{}, true); err != nil {
		t.Error(err)
	}

	//Cannot execute the action if the cron is invalid
	actPrf.Schedule = "* * * *"
	if err := acts.dm.SetActionProfile(actPrf, true); err != nil {
		t.Error(err)
	}
	if err := acts.scheduleActions(cgrEv, []string{}, true); err == nil || err != utils.ErrPartiallyExecuted {
		t.Error(err)
	}
}

func TestAsapExecuteActions(t *testing.T) {
	newData := &dataDBMockError{}
	dm := engine.NewDataManager(newData, config.CgrConfig().CacheCfg(), nil)
	defaultCfg := config.NewDefaultCGRConfig()
	filters := engine.NewFilterS(defaultCfg, nil, dm)
	acts := NewActionS(defaultCfg, filters, dm)

	cgrEv := []*utils.CGREvent{
		{
			Tenant: "cgrates.org",
			ID:     "CHANGED_ID",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
				utils.Destination:  1002,
			},
		},
	}

	expSchedActs := newScheduledActs(cgrEv[0].Tenant, cgrEv[0].ID, utils.MetaNone, utils.EmptyString,
		utils.EmptyString, context.Background(), &ActData{cgrEv[0].Event, cgrEv[0].Opts}, nil)

	if err := acts.asapExecuteActions(expSchedActs); err == nil || err != utils.ErrNoDatabaseConn {
		t.Errorf("Expected %+v, received %+v", utils.ErrNoDatabaseConn, err)
	}

	data := engine.NewInternalDB(nil, nil, true)
	acts.dm = engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	expSchedActs = newScheduledActs(cgrEv[0].Tenant, "another_id", utils.MetaNone, utils.EmptyString,
		utils.EmptyString, context.Background(), &ActData{cgrEv[0].Event, cgrEv[0].Opts}, nil)
	if err := acts.asapExecuteActions(expSchedActs); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}
}

type dataDBMockError struct {
	*engine.DataDBMock
}

func (dbM *dataDBMockError) GetActionProfileDrv(string, string) (*engine.ActionProfile, error) {
	return &engine.ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "test_id1",
		FilterIDs: []string{"*string:~*req.Account:1001;1002;1003", "*prefix:~*req.Destination:10"},
		Actions: []*engine.APAction{
			{
				ID:        "TOPUP",
				FilterIDs: []string{},
				Type:      utils.MetaLog,
				Path:      "~*balance.TestBalance.Value",
				Value:     config.NewRSRParsersMustCompile("10", ","),
			},
		},
	}, nil
}

func (dbM *dataDBMockError) SetActionProfileDrv(*engine.ActionProfile) error {
	return utils.ErrNoDatabaseConn
}
