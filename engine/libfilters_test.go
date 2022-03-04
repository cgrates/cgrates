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
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestLibFiltersGetFloat64OptsReturnConfigOpt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]interface{}{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]interface{}{},
	}
	dynOpts := []*utils.DynamicFloat64Opt{
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
		Event: map[string]interface{}{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]interface{}{},
	}
	dynOpts := []*utils.DynamicFloat64Opt{
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
		Event: map[string]interface{}{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]interface{}{},
	}
	dynOpts := []*utils.DynamicFloat64Opt{
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
		Event: map[string]interface{}{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]interface{}{
			utils.OptsResourcesUnits: 6,
		},
	}
	dynOpts := []*utils.DynamicFloat64Opt{
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
		Event: map[string]interface{}{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]interface{}{},
	}
	dynOpts := []*utils.DynamicStringOpt{
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
		Event: map[string]interface{}{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]interface{}{},
	}
	dynOpts := []*utils.DynamicStringOpt{
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
		Event: map[string]interface{}{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]interface{}{},
	}
	dynOpts := []*utils.DynamicStringOpt{
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
		Event: map[string]interface{}{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]interface{}{
			utils.OptsResourcesUsageID: "value4",
		},
	}
	dynOpts := []*utils.DynamicStringOpt{
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
		Event: map[string]interface{}{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]interface{}{},
	}
	dynOpts := []*utils.DynamicDurationOpt{
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
		Event: map[string]interface{}{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]interface{}{},
	}
	dynOpts := []*utils.DynamicDurationOpt{
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
		Event: map[string]interface{}{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]interface{}{},
	}
	dynOpts := []*utils.DynamicDurationOpt{
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
		Event: map[string]interface{}{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]interface{}{
			utils.OptsResourcesUsageTTL: time.Hour,
		},
	}
	dynOpts := []*utils.DynamicDurationOpt{
		// will never get to this opt because it will return once it
		// finds the one set in APIOpts
		{
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Tenant:    "cgrates.org",
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
		Event: map[string]interface{}{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]interface{}{},
	}
	dynOpts := []*utils.DynamicIntOpt{
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
		Event: map[string]interface{}{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]interface{}{},
	}
	dynOpts := []*utils.DynamicIntOpt{
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
		Event: map[string]interface{}{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]interface{}{},
	}
	dynOpts := []*utils.DynamicIntOpt{
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

func TestLibFiltersGetIntOptsReturnOptFromAPIOpts(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]interface{}{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]interface{}{
			utils.OptsAttributesProcessRuns: 6,
		},
	}
	dynOpts := []*utils.DynamicIntOpt{
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
