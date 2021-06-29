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

package engine

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestHealthAccountAction(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	if err := dm.SetAccountActionPlans("1001", []string{"AP1", "AP2"}, true); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetActionPlan("AP2", &ActionPlan{
		Id:            "AP2",
		AccountIDs:    utils.NewStringMap("1002"),
		ActionTimings: []*ActionTiming{{}},
	}, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}

	exp := &AccountActionPlanIHReply{
		MissingAccountActionPlans: map[string][]string{"1002": {"AP2"}},
		BrokenReferences:          map[string][]string{"AP2": {"1001"}, "AP1": nil},
	}
	if rply, err := GetAccountActionPlansIndexHealth(dm, -1, -1, -1, -1, false, false); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}
}

func TestHealthAccountAction2(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	if err := dm.SetAccountActionPlans("1001", []string{"AP1", "AP2"}, true); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetActionPlan("AP2", &ActionPlan{
		Id:            "AP2",
		AccountIDs:    utils.NewStringMap("1001"),
		ActionTimings: []*ActionTiming{{}},
	}, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}

	exp := &AccountActionPlanIHReply{
		MissingAccountActionPlans: map[string][]string{},
		BrokenReferences:          map[string][]string{"AP1": nil},
	}
	if rply, err := GetAccountActionPlansIndexHealth(dm, -1, -1, -1, -1, false, false); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}
}

func TestHealthAccountAction3(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	if err := dm.SetAccountActionPlans("1002", []string{"AP1"}, true); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetActionPlan("AP1", &ActionPlan{
		Id:            "AP1",
		AccountIDs:    utils.NewStringMap("1002"),
		ActionTimings: []*ActionTiming{{}},
	}, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetActionPlan("AP2", &ActionPlan{
		Id:            "AP2",
		AccountIDs:    utils.NewStringMap("1002"),
		ActionTimings: []*ActionTiming{{}},
	}, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}

	exp := &AccountActionPlanIHReply{
		MissingAccountActionPlans: map[string][]string{"1002": {"AP2"}},
		BrokenReferences:          map[string][]string{},
	}
	if rply, err := GetAccountActionPlansIndexHealth(dm, -1, -1, -1, -1, false, false); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}
}

func TestHealthAccountAction4(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	if err := dm.SetAccountActionPlans("1002", []string{"AP2", "AP1"}, true); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetAccountActionPlans("1001", []string{"AP2"}, true); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetActionPlan("AP1", &ActionPlan{
		Id:            "AP1",
		AccountIDs:    utils.NewStringMap("1002"),
		ActionTimings: []*ActionTiming{{}},
	}, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetActionPlan("AP2", &ActionPlan{
		Id:            "AP2",
		AccountIDs:    utils.NewStringMap("1001"),
		ActionTimings: []*ActionTiming{{}},
	}, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}

	exp := &AccountActionPlanIHReply{
		MissingAccountActionPlans: map[string][]string{},
		BrokenReferences:          map[string][]string{"AP2": {"1002"}},
	}
	if rply, err := GetAccountActionPlansIndexHealth(dm, -1, -1, -1, -1, false, false); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}
}

func TestHealthReverseDestination(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	if err := dm.SetReverseDestination("DST1", []string{"1001", "1002"}, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetReverseDestination("DST2", []string{"1001"}, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetDestination(&Destination{
		Id:       "DST2",
		Prefixes: []string{"1002"},
	}, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}

	exp := &ReverseDestinationsIHReply{
		MissingReverseDestinations: map[string][]string{"1002": {"DST2"}},
		BrokenReferences:           map[string][]string{"DST1": nil, "DST2": {"1001"}},
	}
	if rply, err := GetReverseDestinationsIndexHealth(dm, -1, -1, -1, -1, false, false); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}
}

func TestHealthReverseDestination2(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	if err := dm.SetReverseDestination("DST1", []string{"1001"}, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetReverseDestination("DST2", []string{"1001"}, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetDestination(&Destination{
		Id:       "DST2",
		Prefixes: []string{"1001"},
	}, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}

	exp := &ReverseDestinationsIHReply{
		MissingReverseDestinations: map[string][]string{},
		BrokenReferences:           map[string][]string{"DST1": nil},
	}
	if rply, err := GetReverseDestinationsIndexHealth(dm, -1, -1, -1, -1, false, false); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}
}

func TestHealthReverseDestination3(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	if err := dm.SetReverseDestination("DST1", []string{"1002"}, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetDestination(&Destination{
		Id:       "DST1",
		Prefixes: []string{"1002"},
	}, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetDestination(&Destination{
		Id:       "DST2",
		Prefixes: []string{"1002"},
	}, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}

	exp := &ReverseDestinationsIHReply{
		MissingReverseDestinations: map[string][]string{"1002": {"DST2"}},
		BrokenReferences:           map[string][]string{},
	}
	if rply, err := GetReverseDestinationsIndexHealth(dm, -1, -1, -1, -1, false, false); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}
}

func TestHealthReverseDestination4(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	if err := dm.SetReverseDestination("DST1", []string{"1002"}, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetReverseDestination("DST2", []string{"1001", "1002"}, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetDestination(&Destination{
		Id:       "DST1",
		Prefixes: []string{"1002"},
	}, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetDestination(&Destination{
		Id:       "DST2",
		Prefixes: []string{"1001"},
	}, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}

	exp := &ReverseDestinationsIHReply{
		MissingReverseDestinations: map[string][]string{},
		BrokenReferences:           map[string][]string{"DST2": {"1002"}},
	}
	if rply, err := GetReverseDestinationsIndexHealth(dm, -1, -1, -1, -1, false, false); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}
}
