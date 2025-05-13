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
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
		config.NewDynamicFloat64Opt([]string{"*string:~*req.Account:1001"}, "cgrates.net", 3, nil),
		// filter will not pass, will ignore this opt
		config.NewDynamicFloat64Opt([]string{"*string:~*req.Account:1002"}, "cgrates.org", 4, nil),
		config.NewDynamicFloat64Opt([]string{"*string:~*req.Account:1001"}, "cgrates.org", 5, nil),
	}

	expected := 5.
	if rcv, err := GetFloat64Opts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		utils.OptsResourcesUnits); err != nil {
		t.Error(err)
	} else if rcv != expected {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetFloat64OptsFilterCheckErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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

		config.NewDynamicFloat64Opt([]string{"*string.invalid:filter"}, "cgrates.org", 4, nil),
	}

	experr := `inline parse error for string: <*string.invalid:filter>`
	if _, err := GetFloat64Opts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		utils.OptsResourcesUnits); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestLibFiltersGetFloat64OptsReturnDefaultOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
		config.NewDynamicFloat64Opt([]string{"*string:~*req.Account:1002"}, "cgrates.org", 4, nil),
		config.NewDynamicFloat64Opt(nil, utils.EmptyString, config.ResourcesUnitsDftOpt, nil),
	}

	if rcv, err := GetFloat64Opts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		utils.OptsResourcesUnits); err != nil {
		t.Error(err)
	} else if rcv != config.ResourcesUnitsDftOpt {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", config.ResourcesUnitsDftOpt, rcv)
	}
}

func TestLibFiltersGetFloat64OptsReturnOptFromAPIOpts(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
		config.NewDynamicFloat64Opt([]string{"*string:~*req.Account:1001"}, "cgrates.org", 5, nil),
	}

	expected := 6.
	if rcv, err := GetFloat64Opts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		"nonExistingAPIOpt", utils.OptsResourcesUnits); err != nil {
		t.Error(err)
	} else if rcv != expected {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetStringOptsReturnConfigOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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

		config.NewDynamicStringOpt([]string{"*string:~*req.Account:1001"}, "cgrates.net", "value1", nil),
		// filter will not pass, will ignore this opt
		config.NewDynamicStringOpt([]string{"*string:~*req.Account:1002"}, "cgrates.net", "value2", nil),
		config.NewDynamicStringOpt([]string{"*string:~*req.Account:1001"}, "cgrates.org", "value3", nil),
	}

	expected := "value3"
	if rcv, err := GetStringOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		config.ResourcesUsageIDDftOpt, utils.OptsResourcesUsageID); err != nil {
		t.Error(err)
	} else if rcv != expected {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetStringOptsFilterCheckErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
		config.NewDynamicStringOpt([]string{"*string.invalid:filter"}, "cgrates.org", "value2", nil),
	}

	experr := `inline parse error for string: <*string.invalid:filter>`
	if _, err := GetStringOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		config.ResourcesUsageIDDftOpt, utils.OptsResourcesUsageID); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestLibFiltersGetStringOptsReturnDefaultOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
		config.NewDynamicStringOpt([]string{"*string:~*req.Account:1002"}, "cgrates.org", "value2", nil),
	}

	if rcv, err := GetStringOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		config.ResourcesUsageIDDftOpt, utils.OptsResourcesUsageID); err != nil {
		t.Error(err)
	} else if rcv != config.ResourcesUsageIDDftOpt {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", config.ResourcesUsageIDDftOpt, rcv)
	}
}

func TestLibFiltersGetStringOptsReturnOptFromAPIOpts(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
		config.NewDynamicStringOpt([]string{"*string:~*req.Account:1001"}, "cgrates.org", "value3", nil),
	}

	expected := "value4"
	if rcv, err := GetStringOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		config.ResourcesUsageIDDftOpt, "nonExistingAPIOpt", utils.OptsResourcesUsageID); err != nil {
		t.Error(err)
	} else if rcv != expected {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetDurationOptsReturnConfigOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
		config.NewDynamicDurationOpt([]string{"*string:~*req.Account:1001"}, "cgrates.net", time.Millisecond, nil),
		// filter will not pass, will ignore this opt
		config.NewDynamicDurationOpt([]string{"*string:~*req.Account:1002"}, "cgrates.net", time.Second, nil),
		config.NewDynamicDurationOpt([]string{"*string:~*req.Account:1001"}, "cgrates.org", time.Minute, nil),
	}

	expected := time.Minute
	if rcv, err := GetDurationOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		utils.OptsResourcesUsageTTL); err != nil {
		t.Error(err)
	} else if rcv != expected {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetDurationOptsFilterCheckErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
		config.NewDynamicDurationOpt([]string{"*string.invalid:filter"}, "cgrates.org", time.Second, nil),
	}

	experr := `inline parse error for string: <*string.invalid:filter>`
	if _, err := GetDurationOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		utils.OptsResourcesUsageTTL); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestLibFiltersGetDurationOptsReturnDefaultOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
		config.NewDynamicDurationOpt([]string{"*string:~*req.Account:1002"}, "cgrates.org", time.Second, nil),
		config.NewDynamicDurationOpt(nil, "", config.ResourcesUsageTTLDftOpt, nil),
	}

	if rcv, err := GetDurationOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		utils.OptsResourcesUsageTTL); err != nil {
		t.Error(err)
	} else if rcv != config.ResourcesUsageTTLDftOpt {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", config.ResourcesUsageTTLDftOpt, rcv)
	}
}

func TestLibFiltersGetDurationPointerOptsReturnOptFromAPIOpts(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
	dynOpts := []*config.DynamicDurationPointerOpt{
		// will never get to this opt because it will return once it
		// finds the one set in APIOpts
		config.NewDynamicDurationPointerOpt([]string{"*string:~*req.Account:1001"}, "cgrates.org", utils.DurationPointer(time.Minute), nil),
	}

	expected := time.Hour
	if rcv, err := GetDurationPointerOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		"nonExistingAPIOpt", utils.OptsResourcesUsageTTL); err != nil {
		t.Error(err)
	} else if *rcv != expected {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}
func TestLibFiltersGetDurationPointerOptsReturnConfigOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{},
	}
	dynOpts := []*config.DynamicDurationPointerOpt{
		// tenant will not be recognized, will ignore this opt
		config.NewDynamicDurationPointerOpt([]string{"*string:~*req.Account:1001"}, "cgrates.net", utils.DurationPointer(time.Millisecond), nil),
		// filter will not pass, will ignore this opt
		config.NewDynamicDurationPointerOpt([]string{"*string:~*req.Account:1002"}, "cgrates.net", utils.DurationPointer(time.Second), nil),
		config.NewDynamicDurationPointerOpt([]string{"*string:~*req.Account:1001"}, "cgrates.org", utils.DurationPointer(time.Minute), nil),
	}

	expected := time.Minute
	if rcv, err := GetDurationPointerOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		utils.OptsResourcesUsageTTL); err != nil {
		t.Error(err)
	} else if *rcv != expected {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetDurationPointerOptsFilterCheckErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{},
	}
	dynOpts := []*config.DynamicDurationPointerOpt{
		// function will return error after trying to parse the filter
		config.NewDynamicDurationPointerOpt([]string{"*string.invalid:filter"}, "cgrates.org", utils.DurationPointer(time.Second), nil),
	}

	experr := `inline parse error for string: <*string.invalid:filter>`
	if _, err := GetDurationPointerOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		utils.OptsResourcesUsageTTL); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestLibFiltersGetDurationPointerOptsReturnDefaultOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{},
	}
	dynOpts := []*config.DynamicDurationPointerOpt{
		// filter will not pass, will ignore this opt
		config.NewDynamicDurationPointerOpt([]string{"*string:~*req.Account:1002"}, "cgrates.org", utils.DurationPointer(time.Second), nil),
		config.NewDynamicDurationPointerOpt(nil, "", utils.DurationPointer(config.ResourcesUsageTTLDftOpt), nil),
	}

	if rcv, err := GetDurationPointerOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		utils.OptsResourcesUsageTTL); err != nil {
		t.Error(err)
	} else if *rcv != config.ResourcesUsageTTLDftOpt {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", config.ResourcesUsageTTLDftOpt, rcv)
	}
}

func TestLibFiltersGetDurationOptsReturnOptFromAPIOpts(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
		config.NewDynamicDurationOpt([]string{"*string:~*req.Account:1001"}, "cgrates.org", time.Minute, nil),
	}

	expected := time.Hour
	if rcv, err := GetDurationOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		"nonExistingAPIOpt", utils.OptsResourcesUsageTTL); err != nil {
		t.Error(err)
	} else if rcv != expected {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetIntOptsReturnConfigOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
		config.NewDynamicIntOpt([]string{"*string:~*req.Account:1002"}, "cgrates.net", 3, nil),
		// filter will not pass, will ignore this opt
		config.NewDynamicIntOpt([]string{"*string:~*req.Account:1002"}, "cgrates.org", 4, nil),
		config.NewDynamicIntOpt([]string{"*string:~*req.Account:1001"}, "cgrates.org", 5, nil),
	}

	expected := 5
	if rcv, err := GetIntOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		utils.OptsAttributesProcessRuns); err != nil {
		t.Error(err)
	} else if rcv != expected {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetIntOptsFilterCheckErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
		config.NewDynamicIntOpt([]string{"*string.invalid:filter"}, "cgrates.org", 4, nil),
	}

	experr := `inline parse error for string: <*string.invalid:filter>`
	if _, err := GetIntOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		utils.OptsAttributesProcessRuns); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestLibFiltersGetIntOptsReturnDefaultOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
		config.NewDynamicIntOpt([]string{"*string:~*req.Account:1002"}, "cgrates.org", 4, nil),
		config.NewDynamicIntOpt(nil, utils.EmptyString, config.AttributesProcessRunsDftOpt, nil),
	}

	if rcv, err := GetIntOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		utils.OptsAttributesProcessRuns); err != nil {
		t.Error(err)
	} else if rcv != config.AttributesProcessRunsDftOpt {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", config.AttributesProcessRunsDftOpt, rcv)
	}
}

func TestLibFiltersGetIntOptsReturnOptFromAPIOptsOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
		config.NewDynamicIntOpt([]string{"*string:~*req.Account:1001"}, "cgrates.org", 5, nil),
	}

	expected := 6
	if rcv, err := GetIntOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		"nonExistingAPIOpt", utils.OptsAttributesProcessRuns); err != nil {
		t.Error(err)
	} else if rcv != expected {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetIntOptsReturnOptFromAPIOptsErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
		config.NewDynamicIntOpt([]string{"*string:~*req.Account:1001"}, "cgrates.org", 5, nil),
	}

	experr := `cannot convert field<bool>: true to int`
	if _, err := GetIntOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		"nonExistingAPIOpt", utils.OptsAttributesProcessRuns); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestLibFiltersGetTimeOptsReturnConfigOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
		config.NewDynamicStringOpt([]string{"*string:~*req.Account:1001"}, "cgrates.net", "2022-03-07T15:04:05", nil),
		// filter will not pass, will ignore this opt
		config.NewDynamicStringOpt([]string{"*string:~*req.Account:1002"}, "cgrates.org", "*daily", nil),
		config.NewDynamicStringOpt([]string{"*string:~*req.Account:1001"}, "cgrates.org", "*monthly", nil),
	}

	expected := time.Now().AddDate(0, 1, 0)
	if rcv, err := GetTimeOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		cfg.GeneralCfg().DefaultTimezone, config.RatesStartTimeDftOpt, utils.OptsRatesStartTime); err != nil {
		t.Error(err)
	} else if !dateEqual(rcv, expected) {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetTimeOptsFilterCheckErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
		config.NewDynamicStringOpt([]string{"*string.invalid:filter"}, "cgrates.org", "*daily", nil),
	}

	experr := `inline parse error for string: <*string.invalid:filter>`
	if _, err := GetTimeOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		cfg.GeneralCfg().DefaultTimezone, config.RatesStartTimeDftOpt, utils.OptsRatesStartTime); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestLibFiltersGetTimeOptsReturnDefaultOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{},
	}
	expected, err := utils.ParseTimeDetectLayout(config.RatesStartTimeDftOpt, cfg.GeneralCfg().DefaultTimezone)
	if err != nil {
		t.Error(err)
	}
	dynOpts := []*config.DynamicStringOpt{
		// filter will not pass, will ignore this opt
		config.NewDynamicStringOpt([]string{"*string:~*req.Account:1002"}, "cgrates.org", "*daily", nil),
		config.NewDynamicStringOpt(nil, "", config.RatesStartTimeDftOpt, nil),
	}

	if rcv, err := GetTimeOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		cfg.GeneralCfg().DefaultTimezone, config.RatesStartTimeDftOpt, utils.OptsRatesStartTime); err != nil {
		t.Error(err)
	} else if !dateEqual(rcv, expected) {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetTimeOptsReturnOptFromAPIOpts(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
		config.NewDynamicStringOpt([]string{"*string:~*req.Account:1001"}, "cgrates.org", "*monthly", nil),
	}

	expected := time.Now().AddDate(1, 0, 0)
	if rcv, err := GetTimeOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
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
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
		config.NewDynamicBoolOpt([]string{"*string:~*req.Account:1001"}, "cgrates.net", false, nil),
		// filter will not pass, will ignore this opt
		config.NewDynamicBoolOpt([]string{"*string:~*req.Account:1002"}, "cgrates.org", false, nil),
		config.NewDynamicBoolOpt([]string{"*string:~*req.Account:1001"}, "cgrates.org", true, nil),
	}

	expected := true
	if rcv, err := GetBoolOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		utils.MetaProfileIgnoreFilters); err != nil {
		t.Error(err)
	} else if rcv != expected {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetBoolOptsFilterCheckErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
		config.NewDynamicBoolOpt([]string{"*string.invalid:filter"}, "cgrates.org", false, nil),
	}

	experr := `inline parse error for string: <*string.invalid:filter>`
	if _, err := GetBoolOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		utils.MetaProfileIgnoreFilters); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestLibFiltersGetBoolOptsReturnDefaultOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
		config.NewDynamicBoolOpt([]string{"*string:~*req.Account:1002"}, "cgrates.org", true, nil),
	}

	if rcv, err := GetBoolOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		utils.MetaProfileIgnoreFilters); err != nil {
		t.Error(err)
	} else if rcv != config.ThresholdsProfileIgnoreFiltersDftOpt {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", config.ThresholdsProfileIgnoreFiltersDftOpt, rcv)
	}
}

func TestLibFiltersGetBoolOptsReturnOptFromAPIOpts(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
		config.NewDynamicBoolOpt([]string{"*string:~*req.Account:1001"}, "cgrates.org", false, nil),
	}

	expected := true
	if rcv, err := GetBoolOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		"nonExistingAPIOpt", utils.MetaProfileIgnoreFilters); err != nil {
		t.Error(err)
	} else if rcv != expected {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetInterfaceOptsReturnConfigOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
	if rcv, err := GetInterfaceOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		config.RoutesMaxCostDftOpt, utils.OptsRoutesMaxCost); err != nil {
		t.Error(err)
	} else if rcv != expected {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetInterfaceOptsFilterCheckErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
	if _, err := GetInterfaceOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		config.RoutesMaxCostDftOpt, utils.OptsRoutesMaxCost); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestLibFiltersGetInterfaceOptsReturnDefaultOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
		{
			Value: config.RoutesMaxCostDftOpt,
		},
	}

	if rcv, err := GetInterfaceOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		utils.OptsRoutesMaxCost); err != nil {
		t.Error(err)
	} else if rcv != config.RoutesMaxCostDftOpt {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", utils.ToJSON(config.RoutesMaxCostDftOpt), rcv)
	}
}

func TestLibFiltersGetInterfaceOptsReturnOptFromAPIOpts(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
	if rcv, err := GetInterfaceOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		config.RoutesMaxCostDftOpt, "nonExistingAPIOpt", utils.OptsRoutesMaxCost); err != nil {
		t.Error(err)
	} else if rcv != expected {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetStringSliceOptsReturnConfigOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
	if rcv, err := GetStringSliceOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		config.AttributesProfileIDsDftOpt, utils.OptsAttributesProfileIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetStringSliceOptsFilterCheckErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
	if _, err := GetStringSliceOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		config.AttributesProfileIDsDftOpt, utils.OptsAttributesProfileIDs); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestLibFiltersGetStringSliceOptsReturnDefaultOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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

	if rcv, err := GetStringSliceOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		config.AttributesProfileIDsDftOpt, utils.OptsAttributesProfileIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, config.AttributesProfileIDsDftOpt) {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", config.AttributesProfileIDsDftOpt, rcv)
	}
}

func TestLibFiltersGetStringSliceOptsReturnOptFromAPIOpts(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
	if rcv, err := GetStringSliceOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		config.AttributesProfileIDsDftOpt, "nonExistingAPIOpt", utils.OptsAttributesProfileIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetDecimalBigOptsReturnConfigOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{},
	}
	strOpts := []*config.DynamicInterfaceOpt{
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
	dynOpts, _ := config.IfaceToDecimalBigDynamicOpts(strOpts)

	expected := decimal.New(1234, 0)
	if rcv, err := GetDecimalBigOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		utils.OptsRatesUsage); err != nil {
		t.Error(err)
	} else if rcv.Cmp(expected) != 0 {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestDynamicDecimalBigOptsDynVal(t *testing.T) {
	tests := []struct {
		name    string
		dynOpts []*config.DynamicInterfaceOpt
		expVal  int64
		expErr  error
	}{
		{
			name: "DynOptsVal",
			dynOpts: []*config.DynamicInterfaceOpt{
				{
					Tenant: "cgrates.org",
					Value:  "~*opts.*usage",
				},
			},
			expVal: int64(time.Second * 27),
		},
		{
			name: "DynReqVal",
			dynOpts: []*config.DynamicInterfaceOpt{
				{
					Tenant: "cgrates.org",
					Value:  "~*req.Usage",
				},
			},
			expVal: int64(time.Second * 12),
		},
		{
			name: "StaticVal",
			dynOpts: []*config.DynamicInterfaceOpt{
				{
					Tenant: "cgrates.org",
					Value:  4334,
				},
			},
			expVal: 4334,
		},
		{
			name: "NotFound",
			dynOpts: []*config.DynamicInterfaceOpt{
				{
					Tenant: "cgrates.org",
					Value:  "~*req.RandomField",
				},
			},
			expErr: utils.ErrNotFound,
		},
		{
			name: "ValueNotConvertedCorrectly",
			dynOpts: []*config.DynamicInterfaceOpt{
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
			dynOpts, err := config.IfaceToDecimalBigDynamicOpts(tt.dynOpts)
			if err != nil {
				t.Error(err)
				return
			}
			out, err := GetDecimalBigOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fs, dynOpts, utils.OptsRatesUsage)
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
func TestDynamicIntOptsDynVal(t *testing.T) {
	tests := []struct {
		name    string
		dynOpts []*config.DynamicInterfaceOpt
		expVal  int
		expErr  error
	}{
		{
			name: "DynOptsVal",
			dynOpts: []*config.DynamicInterfaceOpt{
				{
					Tenant: "cgrates.org",
					Value:  "~*opts.*destination",
				},
			},
			expVal: 6548454,
		},
		{
			name: "DynReqVal",
			dynOpts: []*config.DynamicInterfaceOpt{
				{
					Tenant: "cgrates.org",
					Value:  "~*req.Destination",
				},
			},
			expVal: 34534353,
		},
		{
			name: "StaticVal",
			dynOpts: []*config.DynamicInterfaceOpt{
				{
					Tenant: "cgrates.org",
					Value:  5346633,
				},
			},
			expVal: 5346633,
		},
		{
			name: "NotFound",
			dynOpts: []*config.DynamicInterfaceOpt{
				{
					Tenant: "cgrates.org",
					Value:  "~*req.RandomField",
				},
			},
			expErr: utils.ErrNotFound,
		},
		{
			name: "ValueNotConvertedCorrectly",
			dynOpts: []*config.DynamicInterfaceOpt{
				{
					Tenant: "cgrates.org",
					Value:  "~*req.Usage2",
				},
			},
			expErr: fmt.Errorf("strconv.Atoi: parsing \"twenty-five\": invalid syntax"),
		},
	}

	ev := &utils.CGREvent{
		Tenant: utils.CGRateSorg,
		ID:     "testIDEvent",
		Event: map[string]any{
			utils.AccountField: "1001",
			utils.Destination:  34534353,
			"Usage2":           "twenty-five",
		},
		APIOpts: map[string]any{
			utils.MetaDestination: 6548454,
		},
	}
	fs := NewFilterS(config.CgrConfig(), nil, nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dynOpts, err := config.IfaceToIntDynamicOpts(tt.dynOpts)
			if err != nil {
				t.Error(err)
				return
			}
			out, err := GetIntOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fs, dynOpts, utils.OptsRatesUsage)
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
			if tt.expVal != out {
				t.Errorf("expected %d,received %d", tt.expVal, out)
			}
		})
	}
}

func TestDynamicFloat64OptsDynVal(t *testing.T) {
	tests := []struct {
		name    string
		dynOpts []*config.DynamicInterfaceOpt
		expVal  float64
		expErr  error
	}{
		{
			name: "DynOptsVal",
			dynOpts: []*config.DynamicInterfaceOpt{
				{
					Tenant: "cgrates.org",
					Value:  "~*opts.*acc",
				},
			},
			expVal: 3434,
		},
		{
			name: "DynReqVal",
			dynOpts: []*config.DynamicInterfaceOpt{
				{
					Tenant: "cgrates.org",
					Value:  "~*req.*acd",
				},
			},
			expVal: 2213.,
		},
		{
			name: "StaticVal",
			dynOpts: []*config.DynamicInterfaceOpt{
				{
					Tenant: "cgrates.org",
					Value:  23.1,
				},
			},
			expVal: 23.1,
		},
		{
			name: "NotFound",
			dynOpts: []*config.DynamicInterfaceOpt{
				{
					Tenant: "cgrates.org",
					Value:  "~*req.RandomField",
				},
			},
			expErr: utils.ErrNotFound,
		},
		{
			name: "ValueNotConvertedCorrectly",
			dynOpts: []*config.DynamicInterfaceOpt{
				{
					Tenant: "cgrates.org",
					Value:  "~*req.Usage2",
				},
			},
			expErr: fmt.Errorf("strconv.ParseFloat: parsing \"twenty-five\": invalid syntax"),
		},
	}

	ev := &utils.CGREvent{
		Tenant: utils.CGRateSorg,
		ID:     "testIDEvent",
		Event: map[string]any{
			utils.AccountField: "1001",
			"*acd":             2213,
			"Usage2":           "twenty-five",
		},
		APIOpts: map[string]any{
			"*acc": 3434,
		},
	}
	fs := NewFilterS(config.CgrConfig(), nil, nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dynOpts, err := config.InterfaceToFloat64DynamicOpts(tt.dynOpts)
			if err != nil {
				t.Error(err)
				return
			}
			out, err := GetFloat64Opts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fs, dynOpts, utils.OptsRatesUsage)
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
			if tt.expVal != out {
				t.Errorf("expected %f,received %f", tt.expVal, out)
			}
		})
	}
}

func TestDynamicBoolOptsDynVal(t *testing.T) {
	tests := []struct {
		name    string
		dynOpts []*config.DynamicInterfaceOpt
		expVal  bool
		expErr  error
	}{
		{
			name: "DynOptsVal",
			dynOpts: []*config.DynamicInterfaceOpt{
				{
					Tenant: "cgrates.org",
					Value:  "~*opts.Field2",
				},
			},
			expVal: true,
		},
		{
			name: "DynReqVal",
			dynOpts: []*config.DynamicInterfaceOpt{
				{
					Tenant: "cgrates.org",
					Value:  "~*req.Field1",
				},
			},
			expVal: false,
		},
		{
			name: "StaticVal",
			dynOpts: []*config.DynamicInterfaceOpt{
				{
					Tenant: "cgrates.org",
					Value:  true,
				},
			},
			expVal: true,
		},
		{
			name: "NotFound",
			dynOpts: []*config.DynamicInterfaceOpt{
				{
					Tenant: "cgrates.org",
					Value:  "~*req.RandomField",
				},
			},
			expErr: utils.ErrNotFound,
		},
		{
			name: "ValueNotConvertedCorrectly",
			dynOpts: []*config.DynamicInterfaceOpt{
				{
					Tenant: "cgrates.org",
					Value:  "~*req.Usage2",
				},
			},
			expErr: fmt.Errorf("strconv.ParseBool: parsing \"twenty-five\": invalid syntax"),
		},
	}

	ev := &utils.CGREvent{
		Tenant: utils.CGRateSorg,
		ID:     "testIDEvent",
		Event: map[string]any{
			utils.AccountField: "1001",
			"Field1":           false,
			"Usage2":           "twenty-five",
		},
		APIOpts: map[string]any{
			"Field2": true,
		},
	}
	fs := NewFilterS(config.CgrConfig(), nil, nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dynOpts, err := config.IfaceToBoolDynamicOpts(tt.dynOpts)
			if err != nil {
				t.Error(err)
				return
			}
			out, err := GetBoolOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fs, dynOpts, utils.OptsRatesUsage)
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
			if tt.expVal != out {
				t.Errorf("expected %v,received %v", tt.expVal, out)
			}
		})
	}
}

func TestDynamicDurationOptsDynVal(t *testing.T) {
	tests := []struct {
		name    string
		dynOpts []*config.DynamicInterfaceOpt
		expVal  time.Duration
		expErr  error
	}{
		{
			name: "DynOptsVal",
			dynOpts: []*config.DynamicInterfaceOpt{
				{
					Tenant: "cgrates.org",
					Value:  "~*opts.Usage",
				},
			},
			expVal: time.Second * 10,
		},
		{
			name: "DynReqVal",
			dynOpts: []*config.DynamicInterfaceOpt{
				{
					Tenant: "cgrates.org",
					Value:  "~*req.*acd",
				},
			},
			expVal: 3500000,
		},
		{
			name: "StaticVal",
			dynOpts: []*config.DynamicInterfaceOpt{
				{
					Tenant: "cgrates.org",
					Value:  1000000000,
				},
			},
			expVal: 1 * time.Second,
		},
		{
			name: "NotFound",
			dynOpts: []*config.DynamicInterfaceOpt{
				{
					Tenant: "cgrates.org",
					Value:  "~*req.RandomField",
				},
			},
			expErr: utils.ErrNotFound,
		},
		{
			name: "ValueNotConvertedCorrectly",
			dynOpts: []*config.DynamicInterfaceOpt{
				{
					Tenant: "cgrates.org",
					Value:  "~*req.Usage2",
				},
			},
			expErr: fmt.Errorf("time: invalid duration \"twenty-five\""),
		},
	}

	ev := &utils.CGREvent{
		Tenant: utils.CGRateSorg,
		ID:     "testIDEvent",
		Event: map[string]any{
			utils.AccountField: "1001",
			"*acd":             3500000,
			"Usage2":           "twenty-five",
		},
		APIOpts: map[string]any{
			"Usage": "10s",
		},
	}
	fs := NewFilterS(config.CgrConfig(), nil, nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dynOpts, err := config.IfaceToDurationDynamicOpts(tt.dynOpts)
			if err != nil {
				t.Error(err)
				return
			}
			out, err := GetDurationOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fs, dynOpts, utils.OptsRatesUsage)
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
			if tt.expVal != out {
				t.Errorf("expected %v,received %v", tt.expVal, out)
			}
		})
	}
}

func TestDynamicDurationPointerOptsDynVal(t *testing.T) {
	tests := []struct {
		name    string
		dynOpts []*config.DynamicInterfaceOpt
		expVal  *time.Duration
		expErr  error
	}{
		{
			name: "DynOptsVal",
			dynOpts: []*config.DynamicInterfaceOpt{
				{
					Tenant: "cgrates.org",
					Value:  "~*opts.Usage",
				},
			},
			expVal: utils.DurationPointer(time.Second * 10),
		},
		{
			name: "DynReqVal",
			dynOpts: []*config.DynamicInterfaceOpt{
				{
					Tenant: "cgrates.org",
					Value:  "~*req.*acd",
				},
			},
			expVal: utils.DurationPointer(3500000),
		},
		{
			name: "StaticVal",
			dynOpts: []*config.DynamicInterfaceOpt{
				{
					Tenant: "cgrates.org",
					Value:  1000000000,
				},
			},
			expVal: utils.DurationPointer(1 * time.Second),
		},
		{
			name: "NotFound",
			dynOpts: []*config.DynamicInterfaceOpt{
				{
					Tenant: "cgrates.org",
					Value:  "~*req.RandomField",
				},
			},
			expErr: utils.ErrNotFound,
		},
		{
			name: "ValueNotConvertedCorrectly",
			dynOpts: []*config.DynamicInterfaceOpt{
				{
					Tenant: "cgrates.org",
					Value:  "~*req.Usage2",
				},
			},
			expErr: fmt.Errorf("time: invalid duration \"twenty-five\""),
		},
	}

	ev := &utils.CGREvent{
		Tenant: utils.CGRateSorg,
		ID:     "testIDEvent",
		Event: map[string]any{
			utils.AccountField: "1001",
			"*acd":             3500000,
			"Usage2":           "twenty-five",
		},
		APIOpts: map[string]any{
			"Usage": "10s",
		},
	}
	fs := NewFilterS(config.CgrConfig(), nil, nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dynOpts, err := config.IfaceToDurationPointerDynamicOpts(tt.dynOpts)
			if err != nil {
				t.Error(err)
				return
			}
			out, err := GetDurationPointerOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fs, dynOpts, utils.OptsRatesUsage)
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
			if *tt.expVal != *out {
				t.Errorf("expected %v,received %v", tt.expVal, out)
			}
		})
	}
}

func TestDynamicStringOptsDynVal(t *testing.T) {
	tests := []struct {
		name    string
		dynOpts []*config.DynamicInterfaceOpt
		expVal  string
		expErr  error
	}{
		{
			name: "DynOptsVal",
			dynOpts: []*config.DynamicInterfaceOpt{
				{
					Tenant: "cgrates.org",
					Value:  "~*opts.*context",
				},
			},
			expVal: "chargers",
		},
		{
			name: "DynReqVal",
			dynOpts: []*config.DynamicInterfaceOpt{
				{
					Tenant: "cgrates.org",
					Value:  "~*req.Supplier",
				},
			},
			expVal: "Supplier1",
		},
		{
			name: "StaticVal",
			dynOpts: []*config.DynamicInterfaceOpt{
				{
					Tenant: "cgrates.org",
					Value:  "value",
				},
			},
			expVal: "value",
		},
	}

	ev := &utils.CGREvent{
		Tenant: utils.CGRateSorg,
		ID:     "testIDEvent",
		Event: map[string]any{
			utils.AccountField: "1001",
			"Supplier":         "Supplier1",
		},
		APIOpts: map[string]any{
			"*context": "chargers",
		},
	}
	fs := NewFilterS(config.CgrConfig(), nil, nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dynOpts, err := config.InterfaceToDynamicStringOpts(tt.dynOpts)
			if err != nil {
				t.Error(err)
				return
			}
			out, err := GetStringOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fs, dynOpts)
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
			if tt.expVal != out {
				t.Errorf("expected %v,received %v", tt.expVal, out)
			}
		})
	}
}

func TestAttrDynamicOptsFromJson(t *testing.T) {
	cfgJSONStr := `{
		"attributes": {								
			"enabled": true,	
			"stats_conns": ["*internal"],			
			"resources_conns": ["*internal"],		
			"accounts_conns": ["*internal"],			
			"prefix_indexed_fields": ["*req.index1","*req.index2"],		
			"string_indexed_fields": ["*req.index1"],
			"exists_indexed_fields": ["*req.index1","*req.index2"],		
			"notexists_indexed_fields": ["*req.index1"],
			"opts": {
				"*processRuns": [
						{
							"Value": "~*req.ProcessRuns",
							"FilterIDs": ["*string:~*req.Account:1001"]
						},
						{
							"Value": 11,
							"FilterIDs": ["*string:~*req.Account:1003"]
						},
					],
		        "*profileRuns": [			
		        	{
		        		"FilterIDs": ["*string:~*req.Account:1001"],
		        		"Value": "~*opts.ProfileRuns"
		        	}
		        ],
		        "*profileIgnoreFilters": [ 		
		        	{
		        		"FilterIDs": ["*string:~*req.Account:1001"],
		        		"Value": "~*req.IgnoreFilters"
		        	}
		         ] 
			},					
			},		
		}`

	cgrCfg, err := config.NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr)
	if err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{
		Tenant: utils.CGRateSorg,
		ID:     "testIDEvent",
		Event: map[string]any{
			utils.AccountField: "1001",
			"ProcessRuns":      4,
			"IgnoreFilters":    true,
		},
		APIOpts: map[string]any{
			"ProfileRuns": 5,
		},
	}
	ev2 := &utils.CGREvent{
		Tenant: utils.CGRateSorg,
		ID:     "testIDEvent",
		Event: map[string]any{
			utils.AccountField: "1002",
			"ProcessRuns":      4,
			"IgnoreFilters":    true,
		},
		APIOpts: map[string]any{
			"ProfileRuns": 5,
		},
	}
	ev3 := &utils.CGREvent{
		Tenant: utils.CGRateSorg,
		ID:     "testIDEvent",
		Event: map[string]any{
			utils.AccountField: "1003",
		},
		APIOpts: map[string]any{},
	}
	fltrs := NewFilterS(cgrCfg, nil, nil)
	if rcv, err := GetIntOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fltrs, cgrCfg.AttributeSCfg().Opts.ProcessRuns); err != nil {
		t.Error(err)
	} else if rcv != 4 {
		t.Errorf("expected %d,received %d", 4, rcv)
	}
	if rcv, err := GetIntOpts(context.Background(), "cgrates.org", ev3.AsDataProvider(), nil, fltrs, cgrCfg.AttributeSCfg().Opts.ProcessRuns); err != nil {
		t.Error(err)
	} else if rcv != 11 {
		t.Errorf("expected %d,received %d", 11, rcv)
	}
	if rcv, err := GetIntOpts(context.Background(), "cgrates.org", ev2.AsDataProvider(), nil, fltrs, cgrCfg.AttributeSCfg().Opts.ProcessRuns); err != nil {
		t.Error(err)
	} else if rcv != config.AttributesProcessRunsDftOpt {
		t.Errorf("expected %d,received %d", config.AttributesProcessRunsDftOpt, rcv)
	}

	if rcv, err := GetIntOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fltrs, cgrCfg.AttributeSCfg().Opts.ProfileRuns); err != nil {
		t.Error(err)
	} else if rcv != 5 {
		t.Errorf("expected %d,received %d", 5, rcv)
	}
	if rcv, err := GetIntOpts(context.Background(), "cgrates.org", ev2.AsDataProvider(), nil, fltrs, cgrCfg.AttributeSCfg().Opts.ProfileRuns); err != nil {
		t.Error(err)
	} else if rcv != config.AttributesProfileRunsDftOpt {
		t.Errorf("expected %d,received %d", config.AttributesProcessRunsDftOpt, rcv)
	}

	if rcv, err := GetBoolOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fltrs, cgrCfg.AttributeSCfg().Opts.ProfileIgnoreFilters); err != nil {
		t.Error(err)
	} else if !rcv {
		t.Errorf("expected true,received %v", rcv)
	}
	if rcv, err := GetBoolOpts(context.Background(), "cgrates.org", ev2.AsDataProvider(), nil, fltrs, cgrCfg.AttributeSCfg().Opts.ProfileIgnoreFilters); err != nil {
		t.Error(err)
	} else if rcv != config.AttributesProfileIgnoreFiltersDftOpt {
		t.Errorf("expected %v,received %v", config.AttributesProcessRunsDftOpt, rcv)
	}
}

func TestSessionDynamicOptsFromJson(t *testing.T) {
	cfgJSONStr := `{
		"sessions": {
			"enabled": true,
			"listen_bijson": "127.0.0.1:2018",
			"replication_conns": ["*localhost"],
			"store_session_costs": true,
            "min_dur_low_balance": "1s",
			"client_protocol": 2.0,
			"terminate_attempts": 10,
			"opts": {
			  "*accounts": [
		      	{
		      		"FilterIDs": ["*string:~*req.Account:1001"],
		      		"Value": "~*opts.*accounts"
		      	}
		      ],
		      "*attributes": [
		      	{
		      		"FilterIDs": ["*string:~*req.Account:1001"],
		      		"Value": "~*opts.*attributes"
		      	}
		      ],
				"*ttl": [
					{
				        "FilterIDs": ["*string:~*req.Account:1001"],
						"Value": "~*req.TTL",
					},
				],
				"*debitInterval": [
					{
				        "FilterIDs": ["*string:~*req.Account:1001"],
						"Value": "~*req.Usage",
					},
				],
			},
		},
	}`

	cgrCfg, err := config.NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr)
	if err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{
		Tenant: utils.CGRateSorg,
		ID:     "testIDEvent",
		Event: map[string]any{
			utils.AccountField: "1001",
			utils.Usage:        8 * time.Second,
			utils.TTL:          "1s",
		},
		APIOpts: map[string]any{
			utils.MetaAttributes: true,
			utils.MetaAccounts:   true,
		},
	}
	ev2 := &utils.CGREvent{
		Tenant: utils.CGRateSorg,
		ID:     "testIDEvent",
		Event: map[string]any{
			utils.AccountField: "1002",
			utils.Usage:        8 * time.Second,
			utils.TTL:          "1s",
		},
		APIOpts: map[string]any{
			utils.MetaAttributes: true,
			utils.MetaAccounts:   true,
		},
	}
	fltrs := NewFilterS(cgrCfg, nil, nil)
	if rcv, err := GetBoolOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fltrs, cgrCfg.SessionSCfg().Opts.Accounts); err != nil {
		t.Error(err)
	} else if !rcv {
		t.Errorf("expected true,received %v", rcv)
	}
	if rcv, err := GetBoolOpts(context.Background(), "cgrates.org", ev2.AsDataProvider(), nil, fltrs, cgrCfg.SessionSCfg().Opts.Accounts); err != nil {
		t.Error(err)
	} else if rcv != config.SessionsAccountsDftOpt {
		t.Errorf("expected %v,received %v", config.SessionsAccountsDftOpt, rcv)
	}

	if rcv, err := GetBoolOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fltrs, cgrCfg.SessionSCfg().Opts.Attributes); err != nil {
		t.Error(err)
	} else if !rcv {
		t.Errorf("expected true,received %v", rcv)
	}
	if rcv, err := GetBoolOpts(context.Background(), "cgrates.org", ev2.AsDataProvider(), nil, fltrs, cgrCfg.SessionSCfg().Opts.Attributes); err != nil {
		t.Error(err)
	} else if rcv != config.SessionsAttributesDftOpt {
		t.Errorf("expected %v,received %v", config.SessionsAttributesDftOpt, rcv)
	}

	if rcv, err := GetDurationOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fltrs, cgrCfg.SessionSCfg().Opts.TTL); err != nil {
		t.Error(err)
	} else if rcv != time.Second {
		t.Errorf("expected %v,received %v", time.Second, rcv)
	}
	if rcv, err := GetDurationOpts(context.Background(), "cgrates.org", ev2.AsDataProvider(), nil, fltrs, cgrCfg.SessionSCfg().Opts.TTL); err != nil {
		t.Error(err)
	} else if rcv != config.SessionsTTLDftOpt {
		t.Errorf("expected %v,received %v", config.SessionsTTLDftOpt, rcv)
	}

	if rcv, err := GetDurationOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fltrs, cgrCfg.SessionSCfg().Opts.DebitInterval); err != nil {
		t.Error(err)
	} else if rcv != (8 * time.Second) {
		t.Errorf("expected %v,received %v", 8*time.Second, rcv)
	}
	if rcv, err := GetDurationOpts(context.Background(), "cgrates.org", ev2.AsDataProvider(), nil, fltrs, cgrCfg.SessionSCfg().Opts.DebitInterval); err != nil {
		t.Error(err)
	} else if rcv != config.SessionsDebitIntervalDftOpt {
		t.Errorf("expected %v,received %v", config.SessionsDebitIntervalDftOpt, rcv)
	}
}

func TestResDynamicOptsFromJson(t *testing.T) {
	cfgJSONStr := `{
		"resources": {								
			"enabled": true,						
			"store_interval": "7m",					
			"thresholds_conns": ["*internal:*thresholds", "*conn1"],					
			"indexed_selects":true,		
			"nested_fields": true,
				"opts":{
		          "*usageID": [
		          	{
		          		"FilterIDs": ["*string:~*req.Account:1001"],
		          		"Value": "~*req.UsageID"
		          	}
		          ],
		          "*usageTTL": [
		          	{
		          		"FilterIDs": ["*string:~*req.Account:1001"],
		          		"Value": "~*req.UsageTTL"
		          	}
		          ],
		          "*units": [
		          	{
		          		"FilterIDs": ["*string:~*req.Account:1001"],
		          		"Value": "~*opts.*units"
		          	}
		          ]
	           }				
		},
	}`

	cgrCfg, err := config.NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr)
	if err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{
		Tenant: utils.CGRateSorg,
		ID:     "testIDEvent",
		Event: map[string]any{
			utils.AccountField: "1001",
			"UsageID":          "UsgID3232",
			"UsageTTL":         3 * time.Second,
		},
		APIOpts: map[string]any{
			"*units": 23.22,
		},
	}
	ev2 := &utils.CGREvent{
		Tenant: utils.CGRateSorg,
		ID:     "testIDEvent",
		Event: map[string]any{
			utils.AccountField: "1002",
			"UsageID":          "UsgID3232",
			"UsageTTL":         3 * time.Second,
		},
		APIOpts: map[string]any{
			"*units": 23.22,
		},
	}
	fltrs := NewFilterS(cgrCfg, nil, nil)
	if rcv, err := GetStringOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fltrs, cgrCfg.ResourceSCfg().Opts.UsageID); err != nil {
		t.Error(err)
	} else if rcv != "UsgID3232" {
		t.Errorf("expected %s,received %s", "UsgID3232", rcv)
	}
	if rcv, err := GetStringOpts(context.Background(), "cgrates.org", ev2.AsDataProvider(), nil, fltrs, cgrCfg.ResourceSCfg().Opts.UsageID); err != nil {
		t.Error(err)
	} else if rcv != config.ResourcesUsageIDDftOpt {
		t.Errorf("expected %s,received %s", config.ResourcesUsageTTLDftOpt, rcv)
	}

	if rcv, err := GetDurationOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fltrs, cgrCfg.ResourceSCfg().Opts.UsageTTL); err != nil {
		t.Error(err)
	} else if rcv != 3*time.Second {
		t.Errorf("expected %v,received %d", 3*time.Second, rcv)
	}
	if rcv, err := GetDurationOpts(context.Background(), "cgrates.org", ev2.AsDataProvider(), nil, fltrs, cgrCfg.ResourceSCfg().Opts.UsageTTL); err != nil {
		t.Error(err)
	} else if rcv != config.ResourcesUsageTTLDftOpt {
		t.Errorf("expected %d,received %d", config.ResourcesUsageTTLDftOpt, rcv)
	}

	if rcv, err := GetFloat64Opts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fltrs, cgrCfg.ResourceSCfg().Opts.Units); err != nil {
		t.Error(err)
	} else if rcv != 23.22 {
		t.Errorf("expected %f,received %f", 23.22, rcv)
	}
	if rcv, err := GetFloat64Opts(context.Background(), "cgrates.org", ev2.AsDataProvider(), nil, fltrs, cgrCfg.ResourceSCfg().Opts.Units); err != nil {
		t.Error(err)
	} else if rcv != config.ResourcesUnitsDftOpt {
		t.Errorf("expected %v,received %v", config.ResourcesUnitsDftOpt, rcv)
	}
}

func TestRoutesDynamicOptsFromJson(t *testing.T) {
	routeJsnStr := `{
		"routes": {
			"enabled": true,
			"indexed_selects":false,
             "opts":{
		         "*usage": [
		         	{
		         		"FilterIDs": ["*string:~*req.Account:1001"],
		         		"Value": "~*req.Usage"
		         	},
					{
		         		"FilterIDs": ["*string:~*req.Account:1002"],
		         		"Value": 15555
		         	},
					{
		         		"FilterIDs": ["*string:~*req.Account:1003"],
		         		"Value": "5m"
		         	},
					{
		         		"FilterIDs": ["*string:~*req.Account:1004"],
		         		"Value": "~*opts.*usage"
		         	},
		         ]
	            }
		}
	}`

	cgrCfg, err := config.NewCGRConfigFromJSONStringWithDefaults(routeJsnStr)
	if err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{
		Tenant: utils.CGRateSorg,
		ID:     "testIDEvent",
		Event: map[string]any{
			utils.AccountField: "1001",
			utils.Usage:        43364.4,
		},
		APIOpts: map[string]any{
			"*usage": "12m",
		},
	}
	var dec decimal.Big
	dec.SetFloat64(43364.4)

	fltrs := NewFilterS(cgrCfg, nil, nil)
	if rcv, err := GetDecimalBigOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fltrs, cgrCfg.RouteSCfg().Opts.Usage); err != nil {
		t.Error(err)
	} else if rcv.Cmp(&dec) != 0 {
		t.Errorf("expected %v,received %v", dec.String(), rcv.String())
	}
	ev.Event[utils.AccountField] = 1002
	dec.SetUint64(15555)
	if rcv, err := GetDecimalBigOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fltrs, cgrCfg.RouteSCfg().Opts.Usage); err != nil {
		t.Error(err)
	} else if rcv.Cmp(&dec) != 0 {
		t.Errorf("expected %v,received %v", dec.String(), rcv.String())
	}
	ev.Event[utils.AccountField] = 1003
	dec = *decimal.New(int64(5*time.Minute), 0)
	if rcv, err := GetDecimalBigOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fltrs, cgrCfg.RouteSCfg().Opts.Usage); err != nil {
		t.Error(err)
	} else if rcv.Cmp(&dec) != 0 {
		t.Errorf("expected %v,received %v", dec.String(), rcv.String())
	}
	ev.Event[utils.AccountField] = 1004
	dec = *decimal.New(int64(12*time.Minute), 0)
	if rcv, err := GetDecimalBigOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fltrs, cgrCfg.RouteSCfg().Opts.Usage); err != nil {
		t.Error(err)
	} else if rcv.Cmp(&dec) != 0 {
		t.Errorf("expected %v,received %v", dec.String(), rcv.String())
	}
	ev.Event[utils.AccountField] = 1005
	if rcv, err := GetDecimalBigOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fltrs, cgrCfg.RouteSCfg().Opts.Usage); err != nil {
		t.Error(err)
	} else if rcv.Cmp(config.RatesUsageDftOpt) != 0 {
		t.Errorf("expected %v,received %v", config.RatesUsageDftOpt.String(), rcv.String())
	}
}

func TestLibFiltersGetDecimalBigOptsFilterCheckErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{},
	}
	strOpts := []*config.DynamicInterfaceOpt{
		// function will return error after trying to parse the filter
		{
			FilterIDs: []string{"*string.invalid:filter"},
			Tenant:    "cgrates.org",
			Value:     -1,
		},
	}
	dynOpts, _ := config.IfaceToDecimalBigDynamicOpts(strOpts)
	experr := `inline parse error for string: <*string.invalid:filter>`
	if _, err := GetDecimalBigOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		utils.OptsRatesUsage); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestLibFiltersGetDecimalBigOptsReturnDefaultOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{},
	}
	strOpts := []*config.DynamicInterfaceOpt{
		{
			FilterIDs: []string{"*string:~*req.Account:1002"},
			Tenant:    "cgrates.org",
			Value:     -1,
		},
		{Value: config.RatesUsageDftOpt},
	}
	dynOpts, err := config.IfaceToDecimalBigDynamicOpts(strOpts)
	if err != nil {
		t.Fatal(err)
	}
	if rcv, err := GetDecimalBigOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		utils.OptsRatesUsage); err != nil {
		t.Error(err)
	} else if rcv.Cmp(config.RatesUsageDftOpt) != 0 {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", config.RatesUsageDftOpt, rcv)
	}
}

func TestLibFiltersGetDecimalBigOptsReturnOptFromAPIOpts(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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

	strOpts := []*config.DynamicInterfaceOpt{
		// will never get to this opt because it will return once it
		// finds the one set in APIOpts
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.org",
			Value:     decimal.New(1234, 5),
		},
	}
	dynOpts, _ := config.IfaceToDecimalBigDynamicOpts(strOpts)

	expected := decimal.New(4321, 5)
	if rcv, err := GetDecimalBigOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		"nonExistingAPIOpt", utils.OptsRatesUsage); err != nil {
		t.Error(err)
	} else if rcv.Cmp(expected) != 0 {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetIntPointerOptsReturnConfigOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
		config.NewDynamicIntPointerOpt([]string{"*string:~*req.Account:1001"}, "cgrates.net", utils.IntPointer(3), nil),
		// filter will not pass, will ignore this opt
		config.NewDynamicIntPointerOpt([]string{"*string:~*req.Account:1002"}, "cgrates.org", utils.IntPointer(4), nil),
		config.NewDynamicIntPointerOpt([]string{"*string:~*req.Account:1001"}, "cgrates.org", utils.IntPointer(5), nil),
	}

	expected := 5
	if rcv, err := GetIntPointerOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		utils.OptsRoutesProfilesCount); err != nil {
		t.Error(err)
	} else if *rcv != expected {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, *rcv)
	}
}

func TestLibFiltersGetIntPointerOptsFilterCheckErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
		config.NewDynamicIntPointerOpt([]string{"*string.invalid:filter"}, "cgrates.org", utils.IntPointer(4), nil),
	}

	experr := `inline parse error for string: <*string.invalid:filter>`
	if _, err := GetIntPointerOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		utils.OptsRoutesProfilesCount); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestLibFiltersGetIntPointerOptsReturnDft(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
		config.NewDynamicIntPointerOpt([]string{"*string:~*req.Account:1002"}, "cgrates.org", utils.IntPointer(4), nil),
	}

	if rcv, err := GetIntPointerOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		utils.OptsRoutesProfilesCount); err != nil {
		t.Error(err)
	} else if rcv != nil {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", nil, rcv)
	}
}

func TestLibFiltersGetIntPointerOptsReturnOptFromAPIOptsOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
		config.NewDynamicIntPointerOpt([]string{"*string:~*req.Account:1001"}, "cgrates.org", utils.IntPointer(5), nil),
	}

	expected := 6
	if rcv, err := GetIntPointerOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		"nonExistingAPIOpt", utils.OptsRoutesProfilesCount); err != nil {
		t.Error(err)
	} else if *rcv != expected {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", expected, rcv)
	}
}

func TestLibFiltersGetIntPointerOptsReturnOptFromAPIOptsErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
		config.NewDynamicIntPointerOpt([]string{"*string:~*req.Account:1001"}, "cgrates.org", utils.IntPointer(5), nil),
	}

	experr := `cannot convert field<bool>: true to int`
	if _, err := GetIntPointerOpts(context.Background(), "cgrates.org", ev.AsDataProvider(), nil, fS, dynOpts,
		"nonExistingAPIOpt", utils.OptsRoutesProfilesCount); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestLibFiltersGetDurationOptsFromMultipleMapsReturnConfigOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
	fS := NewFilterS(cfg, nil, dm)
	eventStart := map[string]any{
		utils.AccountField: 1001,
	}
	apiOpts := map[string]any{}
	startOpts := map[string]any{}
	dynOpts := []*config.DynamicDurationOpt{
		// tenant will not be recognized, will ignore this opt
		config.NewDynamicDurationOpt([]string{"*string:~*req.Account:1001"}, "cgrates.net", time.Millisecond, nil),
		// filter will not pass, will ignore this opt
		config.NewDynamicDurationOpt([]string{"*string:~*req.Account:1002"}, "cgrates.net", time.Second, nil),
		config.NewDynamicDurationOpt([]string{"*string:~*req.Account:1001"}, "cgrates.org", time.Minute, nil),
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
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
	fS := NewFilterS(cfg, nil, dm)
	eventStart := map[string]any{
		utils.AccountField: 1001,
	}
	apiOpts := map[string]any{}
	startOpts := map[string]any{}
	dynOpts := []*config.DynamicDurationOpt{
		// function will return error after trying to parse the filter
		config.NewDynamicDurationOpt([]string{"*string.invalid:filter"}, "cgrates.org", time.Second, nil),
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
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
	fS := NewFilterS(cfg, nil, dm)
	eventStart := map[string]any{
		utils.AccountField: 1001,
	}
	apiOpts := map[string]any{}
	startOpts := map[string]any{}
	dynOpts := []*config.DynamicDurationOpt{
		// filter will not pass, will ignore this opt
		config.NewDynamicDurationOpt([]string{"*string:~*req.Account:1002"}, "cgrates.org", time.Second, nil),
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
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
		config.NewDynamicDurationOpt([]string{"*string:~*req.Account:1001"}, "cgrates.org", time.Minute, nil),
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
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
		config.NewDynamicDurationOpt([]string{"*string:~*req.Account:1001"}, "cgrates.org", time.Minute, nil),
	}

	experr := `cannot convert field: true to time.Duration`
	if _, err := GetDurationOptsFromMultipleMaps(context.Background(), "cgrates.org", eventStart, apiOpts, startOpts, fS, dynOpts,
		config.SessionsTTLDftOpt, utils.OptsSesTTL); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestLibFiltersGetDurationOptsFromMultipleMapsReturnOptFromStartOptsOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
		config.NewDynamicDurationOpt([]string{"*string:~*req.Account:1001"}, "cgrates.org", time.Minute, nil),
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
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
		config.NewDynamicDurationOpt([]string{"*string:~*req.Account:1001"}, "cgrates.org", time.Minute, nil),
	}

	experr := `cannot convert field: true to time.Duration`
	if _, err := GetDurationOptsFromMultipleMaps(context.Background(), "cgrates.org", eventStart, apiOpts, startOpts, fS, dynOpts,
		config.SessionsTTLDftOpt, utils.OptsSesTTL); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestLibFiltersGetDurationPointerOptsFromMultipleMapsReturnConfigOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
	fS := NewFilterS(cfg, nil, dm)
	eventStart := map[string]any{
		utils.AccountField: 1001,
	}
	apiOpts := map[string]any{}
	startOpts := map[string]any{}
	dynOpts := []*config.DynamicDurationPointerOpt{
		// tenant will not be recognized, will ignore this opt
		config.NewDynamicDurationPointerOpt([]string{"*string:~*req.Account:1001"}, "cgrates.net", utils.DurationPointer(time.Millisecond), nil),
		// filter will not pass, will ignore this opt
		config.NewDynamicDurationPointerOpt([]string{"*string:~*req.Account:1002"}, "cgrates.org", utils.DurationPointer(time.Second), nil),
		config.NewDynamicDurationPointerOpt([]string{"*string:~*req.Account:1001"}, "cgrates.org", utils.DurationPointer(time.Minute), nil),
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
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
	fS := NewFilterS(cfg, nil, dm)
	eventStart := map[string]any{
		utils.AccountField: 1001,
	}
	apiOpts := map[string]any{}
	startOpts := map[string]any{}
	dynOpts := []*config.DynamicDurationPointerOpt{
		// function will return error after trying to parse the filter
		config.NewDynamicDurationPointerOpt([]string{"*string.invalid:filter"}, "cgrates.org", utils.DurationPointer(time.Second), nil),
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
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
	fS := NewFilterS(cfg, nil, dm)
	eventStart := map[string]any{
		utils.AccountField: 1001,
	}
	apiOpts := map[string]any{}
	startOpts := map[string]any{}
	dynOpts := []*config.DynamicDurationPointerOpt{
		// filter will not pass, will ignore this opt
		config.NewDynamicDurationPointerOpt([]string{"*string:~*req.Account:1002"}, "cgrates.org", utils.DurationPointer(time.Second), nil),
	}

	if _, err := GetDurationPointerOptsFromMultipleMaps(context.Background(), "cgrates.org", eventStart, apiOpts, startOpts, fS, dynOpts,
		utils.OptsSesTTLUsage); err != nil {
		t.Error(err)
	}
}

func TestLibFiltersGetDurationPointerOptsFromMultipleMapsReturnOptFromAPIOptsOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
		config.NewDynamicDurationPointerOpt([]string{"*string:~*req.Account:1001"}, "cgrates.org", utils.DurationPointer(time.Minute), nil),
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
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
		config.NewDynamicDurationPointerOpt([]string{"*string:~*req.Account:1001"}, "cgrates.org", utils.DurationPointer(time.Minute), nil),
	}

	experr := `cannot convert field: true to time.Duration`
	if _, err := GetDurationPointerOptsFromMultipleMaps(context.Background(), "cgrates.org", eventStart, apiOpts, startOpts, fS, dynOpts,
		utils.OptsSesTTLUsage); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestLibFiltersGetDurationPointerOptsFromMultipleMapsReturnOptFromStartOptsOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
		config.NewDynamicDurationPointerOpt([]string{"*string:~*req.Account:1001"}, "cgrates.org", utils.DurationPointer(time.Minute), nil),
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
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
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
		config.NewDynamicDurationPointerOpt([]string{"*string:~*req.Account:1001"}, "cgrates.org", utils.DurationPointer(time.Minute), nil),
	}

	experr := `cannot convert field: true to time.Duration`
	if _, err := GetDurationPointerOptsFromMultipleMaps(context.Background(), "cgrates.org", eventStart, apiOpts, startOpts, fS, dynOpts,
		utils.OptsSesTTLUsage); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestGetBoolOptsFieldAsInterfaceErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
	fS := NewFilterS(cfg, nil, dm)
	dynOpts := []*config.DynamicBoolOpt{
		config.NewDynamicBoolOpt([]string{"*string:~*req.Account:1"}, "cgrates.org", false, nil),
	}

	if _, err := GetBoolOpts(context.Background(), "cgrates.org", new(mockDP), nil, fS, dynOpts,
		"nonExistingAPIOpt", utils.MetaProfileIgnoreFilters); err != utils.ErrAccountNotFound {
		t.Errorf("Expecting error <%+v>,\n Recevied  error <%+v>", utils.ErrAccountNotFound, err)
	}

}

func TestGetBoolOptsCantCastErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := NewInternalDB(nil, nil, nil, nil)
	dm := NewDataManager(dataDB, cfg, nil)
	fS := NewFilterS(cfg, nil, dm)
	dynOpts := []*config.DynamicBoolOpt{
		config.NewDynamicBoolOpt([]string{"*string:~*req.Account:1001"}, "cgrates.org", false, nil),
	}
	if _, err := GetBoolOpts(context.Background(), "cgrates.org", utils.StringSet{utils.MetaOpts: {}}, nil, fS, dynOpts,
		"nonExistingAPIOpt", utils.MetaProfileIgnoreFilters); err.Error() != "cannot convert to map[string]any" {
		t.Errorf("Expecting error <%+v>,\n Recevied  error <%+v>", utils.ErrCastFailed, err)
	}

}

func TestOptIfaceFromDP(t *testing.T) {
	tests := []struct {
		name   string
		values []string
		exp    any
		err    error
	}{
		{name: "TestOptBoolValue", values: []string{utils.MetaProfileIgnoreFilters}, exp: true, err: nil},
		{name: "TestOptBoolValue", values: []string{utils.OptsRoutesIgnoreErrors}, exp: true, err: nil},
		{name: "TestOptFloat64Value", values: []string{utils.OptsAttributesProcessRuns}, exp: 5, err: nil},
		{name: "TestOptIntValue", values: []string{utils.OptsResourcesUnits}, exp: float64(23.1), err: nil},
		{name: "TestOptIntValue", values: []string{utils.OptsAttributesProfileRuns}, exp: 3, err: nil},
		{name: "TestOptStringValue", values: []string{utils.OptsRatesStartTime}, exp: "2021-01-01T00:00:00Z", err: nil},
		{name: "TestOptFloat64Value", values: []string{utils.OptsResourcesUsageID}, exp: "RES1", err: nil},
		{name: "TestFieldNotExist", values: []string{"field1"}, err: utils.ErrNotFound},
	}
	cgrEv := utils.MapStorage{
		utils.MetaReq: map[string]any{},
		utils.MetaOpts: map[string]any{
			utils.MetaProfileIgnoreFilters:  true,
			utils.OptsRoutesIgnoreErrors:    true,
			utils.OptsAttributesProcessRuns: 5,
			utils.OptsAttributesProfileRuns: 3,
			utils.OptsRatesStartTime:        "2021-01-01T00:00:00Z",
			utils.OptsResourcesUsageID:      "RES1",
			utils.OptsResourcesUnits:        23.1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if rcv, err := optIfaceFromDP(cgrEv, nil, tt.values); err != tt.err {
				t.Errorf("expected: <%v>,\nreceived: <%v>", tt.err, err)
			} else if rcv != tt.exp {
				t.Errorf("expected: <%v>,\nreceived: <%v>", tt.exp, rcv)
			}
		})
	}
}
