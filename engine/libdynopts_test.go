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
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/ericlagergren/decimal"
)

func TestLibFiltersGetFloat64OptsReturnConfigOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{},
	}
	dynOpts := []*config.DynamicFloat64Opt{
		// tenant will not be recognized, will ignore this opt
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.net",
			Value:     3,
		},
		// filter will not pass, will ignore this opt
		{
			FilterIDs: []string{"*string:~*req.Account:1002"},
			Tenant:    "cgrates.org",
			Value:     4,
		},
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.org",
			Value:     5,
		},
	}

	expected := 5.
	if rcv, err := GetFloat64Opts(context.Background(), "cgrates.org", ev, fS, dynOpts,
		config.ResourcesUnitsDftOpt, utils.OptsResourcesUnits); err != nil {
		t.Error(err)
	} else if rcv != expected {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetFloat64OptsFilterCheckErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{},
	}
	dynOpts := []*config.DynamicFloat64Opt{
		// function will return error after trying to parse the filter
		{
			FilterIDs: []string{"*string.invalid:filter"},
			Tenant:    "cgrates.org",
			Value:     4,
		},
	}

	experr := `inline parse error for string: <*string.invalid:filter>`
	if _, err := GetFloat64Opts(context.Background(), "cgrates.org", ev, fS, dynOpts,
		config.ResourcesUnitsDftOpt, utils.OptsResourcesUnits); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestLibFiltersGetFloat64OptsReturnDefaultOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{},
	}
	dynOpts := []*config.DynamicFloat64Opt{
		// filter will not pass, will ignore this opt
		{
			FilterIDs: []string{"*string:~*req.Account:1002"},
			Tenant:    "cgrates.org",
			Value:     4,
		},
	}

	if rcv, err := GetFloat64Opts(context.Background(), "cgrates.org", ev, fS, dynOpts,
		config.ResourcesUnitsDftOpt, utils.OptsResourcesUnits); err != nil {
		t.Error(err)
	} else if rcv != config.ResourcesUnitsDftOpt {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", config.ResourcesUnitsDftOpt, rcv)
	}
}

func TestLibFiltersGetFloat64OptsReturnOptFromAPIOpts(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{
			utils.OptsResourcesUnits: 6,
		},
	}
	dynOpts := []*config.DynamicFloat64Opt{
		// will never get to this opt because it will return once it
		// finds the one set in APIOpts
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.org",
			Value:     5,
		},
	}

	expected := 6.
	if rcv, err := GetFloat64Opts(context.Background(), "cgrates.org", ev, fS, dynOpts,
		config.ResourcesUnitsDftOpt, "nonExistingAPIOpt", utils.OptsResourcesUnits); err != nil {
		t.Error(err)
	} else if rcv != expected {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetStringOptsReturnConfigOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{},
	}
	dynOpts := []*config.DynamicStringOpt{
		// tenant will not be recognized, will ignore this opt
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.net",
			Value:     "value1",
		},
		// filter will not pass, will ignore this opt
		{
			FilterIDs: []string{"*string:~*req.Account:1002"},
			Tenant:    "cgrates.org",
			Value:     "value2",
		},
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.org",
			Value:     "value3",
		},
	}

	expected := "value3"
	if rcv, err := GetStringOpts(context.Background(), "cgrates.org", ev, fS, dynOpts,
		config.ResourcesUsageIDDftOpt, utils.OptsResourcesUsageID); err != nil {
		t.Error(err)
	} else if rcv != expected {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetStringOptsFilterCheckErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{},
	}
	dynOpts := []*config.DynamicStringOpt{
		// function will return error after trying to parse the filter
		{
			FilterIDs: []string{"*string.invalid:filter"},
			Tenant:    "cgrates.org",
			Value:     "value2",
		},
	}

	experr := `inline parse error for string: <*string.invalid:filter>`
	if _, err := GetStringOpts(context.Background(), "cgrates.org", ev, fS, dynOpts,
		config.ResourcesUsageIDDftOpt, utils.OptsResourcesUsageID); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestLibFiltersGetStringOptsReturnDefaultOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{},
	}
	dynOpts := []*config.DynamicStringOpt{
		// filter will not pass, will ignore this opt
		{
			FilterIDs: []string{"*string:~*req.Account:1002"},
			Tenant:    "cgrates.org",
			Value:     "value2",
		},
	}

	if rcv, err := GetStringOpts(context.Background(), "cgrates.org", ev, fS, dynOpts,
		config.ResourcesUsageIDDftOpt, utils.OptsResourcesUsageID); err != nil {
		t.Error(err)
	} else if rcv != config.ResourcesUsageIDDftOpt {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", config.ResourcesUsageIDDftOpt, rcv)
	}
}

func TestLibFiltersGetStringOptsReturnOptFromAPIOpts(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID: "value4",
		},
	}
	dynOpts := []*config.DynamicStringOpt{
		// will never get to this opt because it will return once it
		// finds the one set in APIOpts
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.org",
			Value:     "value3",
		},
	}

	expected := "value4"
	if rcv, err := GetStringOpts(context.Background(), "cgrates.org", ev, fS, dynOpts,
		config.ResourcesUsageIDDftOpt, "nonExistingAPIOpt", utils.OptsResourcesUsageID); err != nil {
		t.Error(err)
	} else if rcv != expected {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetDurationOptsReturnConfigOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{},
	}
	dynOpts := []*config.DynamicDurationOpt{
		// tenant will not be recognized, will ignore this opt
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.net",
			Value:     time.Millisecond,
		},
		// filter will not pass, will ignore this opt
		{
			FilterIDs: []string{"*string:~*req.Account:1002"},
			Tenant:    "cgrates.org",
			Value:     time.Second,
		},
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.org",
			Value:     time.Minute,
		},
	}

	expected := time.Minute
	if rcv, err := GetDurationOpts(context.Background(), "cgrates.org", ev, fS, dynOpts,
		config.ResourcesUsageTTLDftOpt, utils.OptsResourcesUsageTTL); err != nil {
		t.Error(err)
	} else if rcv != expected {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetDurationOptsFilterCheckErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{},
	}
	dynOpts := []*config.DynamicDurationOpt{
		// function will return error after trying to parse the filter
		{
			FilterIDs: []string{"*string.invalid:filter"},
			Tenant:    "cgrates.org",
			Value:     time.Second,
		},
	}

	experr := `inline parse error for string: <*string.invalid:filter>`
	if _, err := GetDurationOpts(context.Background(), "cgrates.org", ev, fS, dynOpts,
		config.ResourcesUsageTTLDftOpt, utils.OptsResourcesUsageTTL); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestLibFiltersGetDurationOptsReturnDefaultOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{},
	}
	dynOpts := []*config.DynamicDurationOpt{
		// filter will not pass, will ignore this opt
		{
			FilterIDs: []string{"*string:~*req.Account:1002"},
			Tenant:    "cgrates.org",
			Value:     time.Second,
		},
	}

	if rcv, err := GetDurationOpts(context.Background(), "cgrates.org", ev, fS, dynOpts,
		config.ResourcesUsageTTLDftOpt, utils.OptsResourcesUsageTTL); err != nil {
		t.Error(err)
	} else if rcv != config.ResourcesUsageTTLDftOpt {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", config.ResourcesUsageTTLDftOpt, rcv)
	}
}

func TestLibFiltersGetDurationOptsReturnOptFromAPIOpts(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageTTL: time.Hour,
		},
	}
	dynOpts := []*config.DynamicDurationOpt{
		// will never get to this opt because it will return once it
		// finds the one set in APIOpts
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.org",
			Value:     time.Minute,
		},
	}

	expected := time.Hour
	if rcv, err := GetDurationOpts(context.Background(), "cgrates.org", ev, fS, dynOpts,
		config.ResourcesUsageTTLDftOpt, "nonExistingAPIOpt", utils.OptsResourcesUsageTTL); err != nil {
		t.Error(err)
	} else if rcv != expected {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetIntOptsReturnConfigOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{},
	}
	dynOpts := []*config.DynamicIntOpt{
		// tenant will not be recognized, will ignore this opt
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.net",
			Value:     3,
		},
		// filter will not pass, will ignore this opt
		{
			FilterIDs: []string{"*string:~*req.Account:1002"},
			Tenant:    "cgrates.org",
			Value:     4,
		},
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.org",
			Value:     5,
		},
	}

	expected := 5
	if rcv, err := GetIntOpts(context.Background(), "cgrates.org", ev, fS, dynOpts,
		config.AttributesProcessRunsDftOpt, utils.OptsAttributesProcessRuns); err != nil {
		t.Error(err)
	} else if rcv != expected {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetIntOptsFilterCheckErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{},
	}
	dynOpts := []*config.DynamicIntOpt{
		// function will return error after trying to parse the filter
		{
			FilterIDs: []string{"*string.invalid:filter"},
			Tenant:    "cgrates.org",
			Value:     4,
		},
	}

	experr := `inline parse error for string: <*string.invalid:filter>`
	if _, err := GetIntOpts(context.Background(), "cgrates.org", ev, fS, dynOpts,
		config.AttributesProcessRunsDftOpt, utils.OptsAttributesProcessRuns); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestLibFiltersGetIntOptsReturnDefaultOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{},
	}
	dynOpts := []*config.DynamicIntOpt{
		// filter will not pass, will ignore this opt
		{
			FilterIDs: []string{"*string:~*req.Account:1002"},
			Tenant:    "cgrates.org",
			Value:     4,
		},
	}

	if rcv, err := GetIntOpts(context.Background(), "cgrates.org", ev, fS, dynOpts,
		config.AttributesProcessRunsDftOpt, utils.OptsAttributesProcessRuns); err != nil {
		t.Error(err)
	} else if rcv != config.AttributesProcessRunsDftOpt {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", config.AttributesProcessRunsDftOpt, rcv)
	}
}

func TestLibFiltersGetIntOptsReturnOptFromAPIOptsOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProcessRuns: 6,
		},
	}
	dynOpts := []*config.DynamicIntOpt{
		// will never get to this opt because it will return once it
		// finds the one set in APIOpts
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.org",
			Value:     5,
		},
	}

	expected := 6
	if rcv, err := GetIntOpts(context.Background(), "cgrates.org", ev, fS, dynOpts,
		config.AttributesProcessRunsDftOpt, "nonExistingAPIOpt", utils.OptsAttributesProcessRuns); err != nil {
		t.Error(err)
	} else if rcv != expected {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetIntOptsReturnOptFromAPIOptsErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProcessRuns: true,
		},
	}
	dynOpts := []*config.DynamicIntOpt{
		// will never get to this opt because it will return once it
		// finds the one set in APIOpts
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.org",
			Value:     5,
		},
	}

	experr := `cannot convert field<bool>: true to int`
	if _, err := GetIntOpts(context.Background(), "cgrates.org", ev, fS, dynOpts, config.AttributesProcessRunsDftOpt,
		"nonExistingAPIOpt", utils.OptsAttributesProcessRuns); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestLibFiltersGetTimeOptsReturnConfigOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{},
	}
	dynOpts := []*config.DynamicStringOpt{
		// tenant will not be recognized, will ignore this opt
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.net",
			Value:     "2022-03-07T15:04:05",
		},
		// filter will not pass, will ignore this opt
		{
			FilterIDs: []string{"*string:~*req.Account:1002"},
			Tenant:    "cgrates.org",
			Value:     "*daily",
		},
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.org",
			Value:     "*monthly",
		},
	}

	expected := time.Now().AddDate(0, 1, 0)
	if rcv, err := GetTimeOpts(context.Background(), "cgrates.org", ev, fS, dynOpts,
		cfg.GeneralCfg().DefaultTimezone, config.RatesStartTimeDftOpt, utils.OptsRatesStartTime); err != nil {
		t.Error(err)
	} else if !dateEqual(rcv, expected) {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetTimeOptsFilterCheckErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{},
	}
	dynOpts := []*config.DynamicStringOpt{
		// function will return error after trying to parse the filter
		{
			FilterIDs: []string{"*string.invalid:filter"},
			Tenant:    "cgrates.org",
			Value:     "*daily",
		},
	}

	experr := `inline parse error for string: <*string.invalid:filter>`
	if _, err := GetTimeOpts(context.Background(), "cgrates.org", ev, fS, dynOpts,
		cfg.GeneralCfg().DefaultTimezone, config.RatesStartTimeDftOpt, utils.OptsRatesStartTime); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestLibFiltersGetTimeOptsReturnDefaultOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{},
	}
	dynOpts := []*config.DynamicStringOpt{
		// filter will not pass, will ignore this opt
		{
			FilterIDs: []string{"*string:~*req.Account:1002"},
			Tenant:    "cgrates.org",
			Value:     "*daily",
		},
	}

	expected, err := utils.ParseTimeDetectLayout(config.RatesStartTimeDftOpt, cfg.GeneralCfg().DefaultTimezone)
	if err != nil {
		t.Error(err)
	}
	if rcv, err := GetTimeOpts(context.Background(), "cgrates.org", ev, fS, dynOpts,
		cfg.GeneralCfg().DefaultTimezone, config.RatesStartTimeDftOpt, utils.OptsRatesStartTime); err != nil {
		t.Error(err)
	} else if !dateEqual(rcv, expected) {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetTimeOptsReturnOptFromAPIOpts(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{
			utils.OptsRatesStartTime: "*yearly",
		},
	}
	dynOpts := []*config.DynamicStringOpt{
		// will never get to this opt because it will return once it
		// finds the one set in APIOpts
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.org",
			Value:     "*monthly",
		},
	}

	expected := time.Now().AddDate(1, 0, 0)
	if rcv, err := GetTimeOpts(context.Background(), "cgrates.org", ev, fS, dynOpts,
		cfg.GeneralCfg().DefaultTimezone, config.RatesStartTimeDftOpt, "nonExistingAPIOpt", utils.OptsRatesStartTime); err != nil {
		t.Error(err)
	} else if !dateEqual(rcv, expected) {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func dateEqual(date1, date2 time.Time) bool {
	y1, m1, d1 := date1.Date()
	y2, m2, d2 := date2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func TestLibFiltersGetBoolOptsReturnConfigOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{},
	}
	dynOpts := []*config.DynamicBoolOpt{
		// tenant will not be recognized, will ignore this opt
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.net",
			Value:     false,
		},
		// filter will not pass, will ignore this opt
		{
			FilterIDs: []string{"*string:~*req.Account:1002"},
			Tenant:    "cgrates.org",
			Value:     false,
		},
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.org",
			Value:     true,
		},
	}

	expected := true
	if rcv, err := GetBoolOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), fS, dynOpts,
		config.ThresholdsProfileIgnoreFiltersDftOpt, utils.MetaProfileIgnoreFilters); err != nil {
		t.Error(err)
	} else if rcv != expected {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetBoolOptsFilterCheckErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{},
	}
	dynOpts := []*config.DynamicBoolOpt{
		// function will return error after trying to parse the filter
		{
			FilterIDs: []string{"*string.invalid:filter"},
			Tenant:    "cgrates.org",
			Value:     false,
		},
	}

	experr := `inline parse error for string: <*string.invalid:filter>`
	if _, err := GetBoolOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), fS, dynOpts,
		config.ThresholdsProfileIgnoreFiltersDftOpt, utils.MetaProfileIgnoreFilters); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestLibFiltersGetBoolOptsReturnDefaultOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{},
	}
	dynOpts := []*config.DynamicBoolOpt{
		// filter will not pass, will ignore this opt
		{
			FilterIDs: []string{"*string:~*req.Account:1002"},
			Tenant:    "cgrates.org",
			Value:     true,
		},
	}

	if rcv, err := GetBoolOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), fS, dynOpts,
		config.ThresholdsProfileIgnoreFiltersDftOpt, utils.MetaProfileIgnoreFilters); err != nil {
		t.Error(err)
	} else if rcv != config.ThresholdsProfileIgnoreFiltersDftOpt {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", config.ThresholdsProfileIgnoreFiltersDftOpt, rcv)
	}
}

func TestLibFiltersGetBoolOptsReturnOptFromAPIOpts(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{
			utils.MetaProfileIgnoreFilters: true,
		},
	}
	dynOpts := []*config.DynamicBoolOpt{
		// will never get to this opt because it will return once it
		// finds the one set in APIOpts
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.org",
			Value:     false,
		},
	}

	expected := true
	if rcv, err := GetBoolOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), fS, dynOpts,
		config.ThresholdsProfileIgnoreFiltersDftOpt, "nonExistingAPIOpt", utils.MetaProfileIgnoreFilters); err != nil {
		t.Error(err)
	} else if rcv != expected {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetInterfaceOptsReturnConfigOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{},
	}
	dynOpts := []*config.DynamicInterfaceOpt{
		// tenant will not be recognized, will ignore this opt
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.net",
			Value:     1,
		},
		// filter will not pass, will ignore this opt
		{
			FilterIDs: []string{"*string:~*req.Account:1002"},
			Tenant:    "cgrates.org",
			Value:     "value2",
		},
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.org",
			Value:     "value3",
		},
	}

	expected := "value3"
	if rcv, err := GetInterfaceOpts(context.Background(), "cgrates.org", ev, fS, dynOpts,
		config.RoutesMaxCostDftOpt, utils.OptsRoutesMaxCost); err != nil {
		t.Error(err)
	} else if rcv != expected {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetInterfaceOptsFilterCheckErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{},
	}
	dynOpts := []*config.DynamicInterfaceOpt{
		// function will return error after trying to parse the filter
		{
			FilterIDs: []string{"*string.invalid:filter"},
			Tenant:    "cgrates.org",
			Value:     2,
		},
	}

	experr := `inline parse error for string: <*string.invalid:filter>`
	if _, err := GetInterfaceOpts(context.Background(), "cgrates.org", ev, fS, dynOpts,
		config.RoutesMaxCostDftOpt, utils.OptsRoutesMaxCost); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestLibFiltersGetInterfaceOptsReturnDefaultOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{},
	}
	dynOpts := []*config.DynamicInterfaceOpt{
		// filter will not pass, will ignore this opt
		{
			FilterIDs: []string{"*string:~*req.Account:1002"},
			Tenant:    "cgrates.org",
			Value:     2,
		},
	}

	if rcv, err := GetInterfaceOpts(context.Background(), "cgrates.org", ev, fS, dynOpts,
		config.RoutesMaxCostDftOpt, utils.OptsRoutesMaxCost); err != nil {
		t.Error(err)
	} else if rcv != config.RoutesMaxCostDftOpt {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", config.RoutesMaxCostDftOpt, rcv)
	}
}

func TestLibFiltersGetInterfaceOptsReturnOptFromAPIOpts(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{
			utils.OptsRoutesMaxCost: 4,
		},
	}
	dynOpts := []*config.DynamicInterfaceOpt{
		// will never get to this opt because it will return once it
		// finds the one set in APIOpts
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.org",
			Value:     "value3",
		},
	}

	expected := 4
	if rcv, err := GetInterfaceOpts(context.Background(), "cgrates.org", ev, fS, dynOpts,
		config.RoutesMaxCostDftOpt, "nonExistingAPIOpt", utils.OptsRoutesMaxCost); err != nil {
		t.Error(err)
	} else if rcv != expected {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetStringSliceOptsReturnConfigOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{},
	}
	dynOpts := []*config.DynamicStringSliceOpt{
		// tenant will not be recognized, will ignore this opt
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.net",
			Values:    []string{"value1"},
		},
		// filter will not pass, will ignore this opt
		{
			FilterIDs: []string{"*string:~*req.Account:1002"},
			Tenant:    "cgrates.org",
			Values:    []string{"value2"},
		},
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.org",
			Values:    []string{"value3"},
		},
	}

	expected := []string{"value3"}
	if rcv, err := GetStringSliceOpts(context.Background(), "cgrates.org", ev, fS, dynOpts,
		config.AttributesProfileIDsDftOpt, utils.OptsAttributesProfileIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetStringSliceOptsFilterCheckErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{},
	}
	dynOpts := []*config.DynamicStringSliceOpt{
		// function will return error after trying to parse the filter
		{
			FilterIDs: []string{"*string.invalid:filter"},
			Tenant:    "cgrates.org",
			Values:    []string{"value2"},
		},
	}

	experr := `inline parse error for string: <*string.invalid:filter>`
	if _, err := GetStringSliceOpts(context.Background(), "cgrates.org", ev, fS, dynOpts,
		config.AttributesProfileIDsDftOpt, utils.OptsAttributesProfileIDs); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestLibFiltersGetStringSliceOptsReturnDefaultOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{},
	}
	dynOpts := []*config.DynamicStringSliceOpt{
		// filter will not pass, will ignore this opt
		{
			FilterIDs: []string{"*string:~*req.Account:1002"},
			Tenant:    "cgrates.org",
			Values:    []string{"value2"},
		},
	}

	if rcv, err := GetStringSliceOpts(context.Background(), "cgrates.org", ev, fS, dynOpts,
		config.AttributesProfileIDsDftOpt, utils.OptsAttributesProfileIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, config.AttributesProfileIDsDftOpt) {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", config.AttributesProfileIDsDftOpt, rcv)
	}
}

func TestLibFiltersGetStringSliceOptsReturnOptFromAPIOpts(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProfileIDs: []string{"value4"},
		},
	}
	dynOpts := []*config.DynamicStringSliceOpt{
		// will never get to this opt because it will return once it
		// finds the one set in APIOpts
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.org",
			Values:    []string{"value3"},
		},
	}

	expected := []string{"value4"}
	if rcv, err := GetStringSliceOpts(context.Background(), "cgrates.org", ev, fS, dynOpts,
		config.AttributesProfileIDsDftOpt, "nonExistingAPIOpt", utils.OptsAttributesProfileIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetDecimalBigOptsReturnConfigOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{},
	}
	strOpts := []*config.DynamicStringOpt{
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.net",
			Value:     "42",
		},
		// filter will not pass, will ignore this opt
		{
			FilterIDs: []string{"*string:~*req.Account:1002"},
			Tenant:    "cgrates.org",
			Value:     "-1",
		},
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.org",
			Value:     "1234",
		},
	}
	dynOpts, _ := config.StringToDecimalBigDynamicOpts(strOpts)

	expected := decimal.New(1234, 0)
	if rcv, err := GetDecimalBigOpts(context.Background(), "cgrates.org", ev, fS, dynOpts,
		config.RatesUsageDftOpt, utils.OptsRatesUsage); err != nil {
		t.Error(err)
	} else if rcv.Cmp(expected) != 0 {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestDynamicDecimalBigOptsDynVal(t *testing.T) {
	tests := []struct {
		name    string
		dynOpts []*config.DynamicStringOpt
		expVal  int64
		expErr  error
	}{
		{
			name: "DynOptsVal",
			dynOpts: []*config.DynamicStringOpt{
				{
					Tenant: "cgrates.org",
					Value:  "~*opts.*usage",
				},
			},
			expVal: int64(time.Second * 27),
		},
		{
			name: "DynReqVal",
			dynOpts: []*config.DynamicStringOpt{
				{
					Tenant: "cgrates.org",
					Value:  "~*req.Usage",
				},
			},
			expVal: int64(time.Second * 12),
		},
		{
			name: "StaticVal",
			dynOpts: []*config.DynamicStringOpt{
				{
					Tenant: "cgrates.org",
					Value:  "4334",
				},
			},
			expVal: 4334,
		},
		{
			name: "NotFound",
			dynOpts: []*config.DynamicStringOpt{
				{
					Tenant: "cgrates.org",
					Value:  "~*req.RandomField",
				},
			},
			expErr: utils.ErrNotFound,
		},
		{
			name: "ValueNotConvertedCorrectly",
			dynOpts: []*config.DynamicStringOpt{
				{
					Tenant: "cgrates.org",
					Value:  "~*req.Usage2",
				},
			},
			expErr: fmt.Errorf("can't convert <twenty-five> to decimal"),
		},
	}

	ev := &utils.CGREvent{
		Tenant: utils.CGRateSorg,
		ID:     "testIDEvent",
		Event: map[string]any{
			utils.AccountField: "1001",
			utils.Usage:        "12s",
			"Usage2":           "twenty-five",
		},
		APIOpts: map[string]any{
			utils.MetaUsage: "27s",
		},
	}
	fs := NewFilterS(config.CgrConfig(), nil, nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dynOpts, err := config.StringToDecimalBigDynamicOpts(tt.dynOpts)
			if err != nil {
				t.Error(err)
				return
			}
			out, err := GetDecimalBigOpts(context.Background(), "cgrates.org", ev, fs, dynOpts, config.RatesUsageDftOpt, utils.OptsRatesUsage)
			if tt.expErr != nil {
				if err == nil {
					t.Error("expected err,received nil")
				}
				if err.Error() != tt.expErr.Error() {
					t.Errorf("expected error %v,received %v", tt.expErr, err)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected err %v", err)
			}
			val, _ := out.Int64()
			if tt.expVal != val {
				t.Errorf("expected %d,received %d", tt.expVal, val)
			}
		})
	}
}

func TestLibFiltersGetDecimalBigOptsFilterCheckErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{},
	}
	strOpts := []*config.DynamicStringOpt{
		// function will return error after trying to parse the filter
		{
			FilterIDs: []string{"*string.invalid:filter"},
			Tenant:    "cgrates.org",
			Value:     "-1",
		},
	}
	dynOpts, _ := config.StringToDecimalBigDynamicOpts(strOpts)
	experr := `inline parse error for string: <*string.invalid:filter>`
	if _, err := GetDecimalBigOpts(context.Background(), "cgrates.org", ev, fS, dynOpts,
		config.RatesUsageDftOpt, utils.OptsRatesUsage); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestLibFiltersGetDecimalBigOptsReturnDefaultOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{},
	}
	strOpts := []*config.DynamicStringOpt{
		{
			FilterIDs: []string{"*string:~*req.Account:1002"},
			Tenant:    "cgrates.org",
			Value:     "-1",
		},
	}
	dynOpts, _ := config.StringToDecimalBigDynamicOpts(strOpts)
	if rcv, err := GetDecimalBigOpts(context.Background(), "cgrates.org", ev, fS, dynOpts,
		config.RatesUsageDftOpt, utils.OptsRatesUsage); err != nil {
		t.Error(err)
	} else if rcv.Cmp(config.RatesUsageDftOpt) != 0 {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", config.RatesUsageDftOpt, rcv)
	}
}

func TestLibFiltersGetDecimalBigOptsReturnOptFromAPIOpts(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{
			utils.OptsRatesUsage: decimal.New(4321, 5),
		},
	}

	strOpts := []*config.DynamicStringOpt{
		// will never get to this opt because it will return once it
		// finds the one set in APIOpts
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.org",
			Value:     "1234",
		},
	}
	dynOpts, _ := config.StringToDecimalBigDynamicOpts(strOpts)

	expected := decimal.New(4321, 5)
	if rcv, err := GetDecimalBigOpts(context.Background(), "cgrates.org", ev, fS, dynOpts,
		config.RatesUsageDftOpt, "nonExistingAPIOpt", utils.OptsRatesUsage); err != nil {
		t.Error(err)
	} else if rcv.Cmp(expected) != 0 {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetIntPointerOptsReturnConfigOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{},
	}
	dynOpts := []*config.DynamicIntPointerOpt{
		// tenant will not be recognized, will ignore this opt
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.net",
			Value:     utils.IntPointer(3),
		},
		// filter will not pass, will ignore this opt
		{
			FilterIDs: []string{"*string:~*req.Account:1002"},
			Tenant:    "cgrates.org",
			Value:     utils.IntPointer(4),
		},
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.org",
			Value:     utils.IntPointer(5),
		},
	}

	expected := 5
	if rcv, err := GetIntPointerOpts(context.Background(), "cgrates.org", ev, fS, dynOpts,
		utils.OptsRoutesProfilesCount); err != nil {
		t.Error(err)
	} else if *rcv != expected {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, *rcv)
	}
}

func TestLibFiltersGetIntPointerOptsFilterCheckErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{},
	}
	dynOpts := []*config.DynamicIntPointerOpt{
		// function will return error after trying to parse the filter
		{
			FilterIDs: []string{"*string.invalid:filter"},
			Tenant:    "cgrates.org",
			Value:     utils.IntPointer(4),
		},
	}

	experr := `inline parse error for string: <*string.invalid:filter>`
	if _, err := GetIntPointerOpts(context.Background(), "cgrates.org", ev, fS, dynOpts,
		utils.OptsRoutesProfilesCount); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestLibFiltersGetIntPointerOptsReturnDft(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{},
	}
	dynOpts := []*config.DynamicIntPointerOpt{
		// filter will not pass, will ignore this opt
		{
			FilterIDs: []string{"*string:~*req.Account:1002"},
			Tenant:    "cgrates.org",
			Value:     utils.IntPointer(4),
		},
	}

	if rcv, err := GetIntPointerOpts(context.Background(), "cgrates.org", ev, fS, dynOpts,
		utils.OptsRoutesProfilesCount); err != nil {
		t.Error(err)
	} else if rcv != nil {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", nil, rcv)
	}
}

func TestLibFiltersGetIntPointerOptsReturnOptFromAPIOptsOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{
			utils.OptsRoutesProfilesCount: 6,
		},
	}
	dynOpts := []*config.DynamicIntPointerOpt{
		// will never get to this opt because it will return once it
		// finds the one set in APIOpts
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.org",
			Value:     utils.IntPointer(5),
		},
	}

	expected := 6
	if rcv, err := GetIntPointerOpts(context.Background(), "cgrates.org", ev, fS, dynOpts,
		"nonExistingAPIOpt", utils.OptsRoutesProfilesCount); err != nil {
		t.Error(err)
	} else if *rcv != expected {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetIntPointerOptsReturnOptFromAPIOptsErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{
			utils.OptsRoutesProfilesCount: true,
		},
	}
	dynOpts := []*config.DynamicIntPointerOpt{
		// will never get to this opt because it will return once it
		// finds the one set in APIOpts
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.org",
			Value:     utils.IntPointer(5),
		},
	}

	experr := `cannot convert field<bool>: true to int`
	if _, err := GetIntPointerOpts(context.Background(), "cgrates.org", ev, fS, dynOpts,
		"nonExistingAPIOpt", utils.OptsRoutesProfilesCount); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestLibFiltersGetDurationOptsFromMultipleMapsReturnConfigOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	eventStart := map[string]any{
		utils.AccountField: 1001,
	}
	apiOpts := map[string]any{}
	startOpts := map[string]any{}
	dynOpts := []*config.DynamicDurationOpt{
		// tenant will not be recognized, will ignore this opt
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.net",
			Value:     time.Millisecond,
		},
		// filter will not pass, will ignore this opt
		{
			FilterIDs: []string{"*string:~*req.Account:1002"},
			Tenant:    "cgrates.org",
			Value:     time.Second,
		},
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.org",
			Value:     time.Minute,
		},
	}

	expected := time.Minute
	if rcv, err := GetDurationOptsFromMultipleMaps(context.Background(), "cgrates.org", eventStart, apiOpts, startOpts, fS, dynOpts,
		config.SessionsTTLDftOpt, utils.OptsSesTTL); err != nil {
		t.Error(err)
	} else if rcv != expected {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetDurationOptsFromMultipleMapsFilterCheckErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	eventStart := map[string]any{
		utils.AccountField: 1001,
	}
	apiOpts := map[string]any{}
	startOpts := map[string]any{}
	dynOpts := []*config.DynamicDurationOpt{
		// function will return error after trying to parse the filter
		{
			FilterIDs: []string{"*string.invalid:filter"},
			Tenant:    "cgrates.org",
			Value:     time.Second,
		},
	}

	experr := `inline parse error for string: <*string.invalid:filter>`
	if _, err := GetDurationOptsFromMultipleMaps(context.Background(), "cgrates.org", eventStart, apiOpts, startOpts, fS, dynOpts,
		config.SessionsTTLDftOpt, utils.OptsSesTTL); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestLibFiltersGetDurationOptsFromMultipleMapsErrNotFound(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	eventStart := map[string]any{
		utils.AccountField: 1001,
	}
	apiOpts := map[string]any{}
	startOpts := map[string]any{}
	dynOpts := []*config.DynamicDurationOpt{
		// filter will not pass, will ignore this opt
		{
			FilterIDs: []string{"*string:~*req.Account:1002"},
			Tenant:    "cgrates.org",
			Value:     time.Second,
		},
	}

	if rcv, err := GetDurationOptsFromMultipleMaps(context.Background(), "cgrates.org", eventStart, apiOpts, startOpts, fS, dynOpts,
		config.SessionsTTLDftOpt, utils.OptsSesTTL); err != nil {
		t.Error(err)
	} else if rcv != config.SessionsTTLDftOpt {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", config.SessionsTTLDftOpt, rcv)
	}
}

func TestLibFiltersGetDurationOptsFromMultipleMapsReturnOptFromAPIOptsOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	eventStart := map[string]any{
		utils.AccountField: 1001,
	}
	apiOpts := map[string]any{
		utils.OptsSesTTL: time.Hour,
	}
	startOpts := map[string]any{}
	dynOpts := []*config.DynamicDurationOpt{
		// will never get to this opt because it will return once it
		// finds the one set in APIOpts
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.org",
			Value:     time.Minute,
		},
	}

	expected := time.Hour
	if rcv, err := GetDurationOptsFromMultipleMaps(context.Background(), "cgrates.org", eventStart, apiOpts, startOpts, fS, dynOpts,
		config.SessionsTTLDftOpt, utils.OptsSesTTL); err != nil {
		t.Error(err)
	} else if rcv != expected {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetDurationOptsFromMultipleMapsReturnOptFromAPIOptsErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	eventStart := map[string]any{
		utils.AccountField: 1001,
	}
	apiOpts := map[string]any{
		utils.OptsSesTTL: true,
	}
	startOpts := map[string]any{}
	dynOpts := []*config.DynamicDurationOpt{
		// will never get to this opt because it will return once it
		// finds the one set in APIOpts
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.org",
			Value:     time.Minute,
		},
	}

	experr := `cannot convert field: true to time.Duration`
	if _, err := GetDurationOptsFromMultipleMaps(context.Background(), "cgrates.org", eventStart, apiOpts, startOpts, fS, dynOpts,
		config.SessionsTTLDftOpt, utils.OptsSesTTL); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestLibFiltersGetDurationOptsFromMultipleMapsReturnOptFromStartOptsOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	eventStart := map[string]any{
		utils.AccountField: 1001,
	}
	apiOpts := map[string]any{}
	startOpts := map[string]any{
		utils.OptsSesTTL: time.Hour,
	}
	dynOpts := []*config.DynamicDurationOpt{
		// will never get to this opt because it will return once it
		// finds the one set in APIOpts
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.org",
			Value:     time.Minute,
		},
	}

	expected := time.Hour
	if rcv, err := GetDurationOptsFromMultipleMaps(context.Background(), "cgrates.org", eventStart, apiOpts, startOpts, fS, dynOpts,
		config.SessionsTTLDftOpt, utils.OptsSesTTL); err != nil {
		t.Error(err)
	} else if rcv != expected {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetDurationOptsFromMultipleMapsReturnOptFromStartOptsErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	eventStart := map[string]any{
		utils.AccountField: 1001,
	}
	apiOpts := map[string]any{}
	startOpts := map[string]any{
		utils.OptsSesTTL: true,
	}
	dynOpts := []*config.DynamicDurationOpt{
		// will never get to this opt because it will return once it
		// finds the one set in APIOpts
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.org",
			Value:     time.Minute,
		},
	}

	experr := `cannot convert field: true to time.Duration`
	if _, err := GetDurationOptsFromMultipleMaps(context.Background(), "cgrates.org", eventStart, apiOpts, startOpts, fS, dynOpts,
		config.SessionsTTLDftOpt, utils.OptsSesTTL); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestLibFiltersGetDurationPointerOptsFromMultipleMapsReturnConfigOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	eventStart := map[string]any{
		utils.AccountField: 1001,
	}
	apiOpts := map[string]any{}
	startOpts := map[string]any{}
	dynOpts := []*config.DynamicDurationPointerOpt{
		// tenant will not be recognized, will ignore this opt
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.net",
			Value:     utils.DurationPointer(time.Millisecond),
		},
		// filter will not pass, will ignore this opt
		{
			FilterIDs: []string{"*string:~*req.Account:1002"},
			Tenant:    "cgrates.org",
			Value:     utils.DurationPointer(time.Second),
		},
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.org",
			Value:     utils.DurationPointer(time.Minute),
		},
	}

	expected := time.Minute
	if rcv, err := GetDurationPointerOptsFromMultipleMaps(context.Background(), "cgrates.org", eventStart, apiOpts, startOpts, fS, dynOpts,
		utils.OptsSesTTLUsage); err != nil {
		t.Error(err)
	} else if *rcv != expected {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetDurationPointerOptsFromMultipleMapsFilterCheckErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	eventStart := map[string]any{
		utils.AccountField: 1001,
	}
	apiOpts := map[string]any{}
	startOpts := map[string]any{}
	dynOpts := []*config.DynamicDurationPointerOpt{
		// function will return error after trying to parse the filter
		{
			FilterIDs: []string{"*string.invalid:filter"},
			Tenant:    "cgrates.org",
			Value:     utils.DurationPointer(time.Second),
		},
	}

	experr := `inline parse error for string: <*string.invalid:filter>`
	if _, err := GetDurationPointerOptsFromMultipleMaps(context.Background(), "cgrates.org", eventStart, apiOpts, startOpts, fS, dynOpts,
		utils.OptsSesTTLUsage); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestLibFiltersGetDurationPointerOptsFromMultipleMapsReturnDft(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	eventStart := map[string]any{
		utils.AccountField: 1001,
	}
	apiOpts := map[string]any{}
	startOpts := map[string]any{}
	dynOpts := []*config.DynamicDurationPointerOpt{
		// filter will not pass, will ignore this opt
		{
			FilterIDs: []string{"*string:~*req.Account:1002"},
			Tenant:    "cgrates.org",
			Value:     utils.DurationPointer(time.Second),
		},
	}

	if _, err := GetDurationPointerOptsFromMultipleMaps(context.Background(), "cgrates.org", eventStart, apiOpts, startOpts, fS, dynOpts,
		utils.OptsSesTTLUsage); err != nil {
		t.Error(err)
	}
}

func TestLibFiltersGetDurationPointerOptsFromMultipleMapsReturnOptFromAPIOptsOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	eventStart := map[string]any{
		utils.AccountField: 1001,
	}
	apiOpts := map[string]any{
		utils.OptsSesTTLUsage: time.Hour,
	}
	startOpts := map[string]any{}
	dynOpts := []*config.DynamicDurationPointerOpt{
		// will never get to this opt because it will return once it
		// finds the one set in APIOpts
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.org",
			Value:     utils.DurationPointer(time.Minute),
		},
	}

	expected := time.Hour
	if rcv, err := GetDurationPointerOptsFromMultipleMaps(context.Background(), "cgrates.org", eventStart, apiOpts, startOpts, fS, dynOpts,
		utils.OptsSesTTLUsage); err != nil {
		t.Error(err)
	} else if *rcv != expected {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetDurationPointerOptsFromMultipleMapsReturnOptFromAPIOptsErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	eventStart := map[string]any{
		utils.AccountField: 1001,
	}
	apiOpts := map[string]any{
		utils.OptsSesTTLUsage: true,
	}
	startOpts := map[string]any{}
	dynOpts := []*config.DynamicDurationPointerOpt{
		// will never get to this opt because it will return once it
		// finds the one set in APIOpts
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.org",
			Value:     utils.DurationPointer(time.Minute),
		},
	}

	experr := `cannot convert field: true to time.Duration`
	if _, err := GetDurationPointerOptsFromMultipleMaps(context.Background(), "cgrates.org", eventStart, apiOpts, startOpts, fS, dynOpts,
		utils.OptsSesTTLUsage); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestLibFiltersGetDurationPointerOptsFromMultipleMapsReturnOptFromStartOptsOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	eventStart := map[string]any{
		utils.AccountField: 1001,
	}
	apiOpts := map[string]any{}
	startOpts := map[string]any{
		utils.OptsSesTTLUsage: time.Hour,
	}
	dynOpts := []*config.DynamicDurationPointerOpt{
		// will never get to this opt because it will return once it
		// finds the one set in APIOpts
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.org",
			Value:     utils.DurationPointer(time.Minute),
		},
	}

	expected := time.Hour
	if rcv, err := GetDurationPointerOptsFromMultipleMaps(context.Background(), "cgrates.org", eventStart, apiOpts, startOpts, fS, dynOpts,
		utils.OptsSesTTLUsage); err != nil {
		t.Error(err)
	} else if *rcv != expected {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetDurationPointerOptsFromMultipleMapsReturnOptFromStartOptsErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	eventStart := map[string]any{
		utils.AccountField: 1001,
	}
	apiOpts := map[string]any{}
	startOpts := map[string]any{
		utils.OptsSesTTLUsage: true,
	}
	dynOpts := []*config.DynamicDurationPointerOpt{
		// will never get to this opt because it will return once it
		// finds the one set in APIOpts
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.org",
			Value:     utils.DurationPointer(time.Minute),
		},
	}

	experr := `cannot convert field: true to time.Duration`
	if _, err := GetDurationPointerOptsFromMultipleMaps(context.Background(), "cgrates.org", eventStart, apiOpts, startOpts, fS, dynOpts,
		utils.OptsSesTTLUsage); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestGetBoolOptsFieldAsInterfaceErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	dynOpts := []*config.DynamicBoolOpt{
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.org",
			Value:     false,
		},
	}

	if _, err := GetBoolOpts(context.Background(), "cgrates.org", new(mockDP), fS, dynOpts,
		config.ThresholdsProfileIgnoreFiltersDftOpt, "nonExistingAPIOpt", utils.MetaProfileIgnoreFilters); err != utils.ErrAccountNotFound {
		t.Errorf("Expecting error <%+v>,\n Recevied  error <%+v>", utils.ErrAccountNotFound, err)
	}

}

func TestGetBoolOptsCantCastErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	dynOpts := []*config.DynamicBoolOpt{
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.org",
			Value:     false,
		},
	}
	if _, err := GetBoolOpts(context.Background(), "cgrates.org", utils.StringSet{utils.MetaOpts: {}}, fS, dynOpts,
		config.ThresholdsProfileIgnoreFiltersDftOpt, "nonExistingAPIOpt", utils.MetaProfileIgnoreFilters); err.Error() != "cannot convert to map[string]any" {
		t.Errorf("Expecting error <%+v>,\n Recevied  error <%+v>", utils.ErrCastFailed, err)
	}

}
