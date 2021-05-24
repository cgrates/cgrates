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

package apis

import (
	"reflect"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestCallCacheForFilter(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, true), cfg.CacheCfg(), nil)
	tnt := "cgrates.org"
	flt := &engine.Filter{
		Tenant: tnt,
		ID:     "FLTR1",
		Rules: []*engine.FilterRule{{
			Type:    utils.MetaString,
			Element: "~*req.Account",
			Values:  []string{"1001"},
		}},
	}
	if err := flt.Compile(); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetFilter(context.TODO(), flt, true); err != nil {
		t.Fatal(err)
	}
	th := &engine.ThresholdProfile{
		Tenant:    tnt,
		ID:        "TH1",
		FilterIDs: []string{flt.ID},
	}
	if err := dm.SetThresholdProfile(context.TODO(), th, true); err != nil {
		t.Fatal(err)
	}
	dsp := &engine.DispatcherProfile{
		Tenant:     tnt,
		Subsystems: []string{utils.MetaAny},
		ID:         "Dsp1",
		FilterIDs:  []string{flt.ID},
	}
	if err := dm.SetDispatcherProfile(context.TODO(), dsp, true); err != nil {
		t.Fatal(err)
	}

	exp := map[string][]string{
		utils.FilterIDs:                {"cgrates.org:FLTR1"},
		utils.DispatcherFilterIndexIDs: {"cgrates.org:*any:*string:*req.Account:1001"},
		utils.ThresholdFilterIndexIDs:  {"cgrates.org:*string:*req.Account:1001"},
	}
	rpl, err := composeCacheArgsForFilter(dm, context.TODO(), flt, tnt, flt.TenantID(), map[string][]string{utils.FilterIDs: {"cgrates.org:FLTR1"}})
	if err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(rpl, exp) {
		t.Errorf("Expected %s ,received: %s", utils.ToJSON(exp), utils.ToJSON(rpl))
	}
	flt = &engine.Filter{
		Tenant: tnt,
		ID:     "FLTR1",
		Rules: []*engine.FilterRule{{
			Type:    utils.MetaString,
			Element: "~*req.Account",
			Values:  []string{"1002"},
		}},
	}
	if err := flt.Compile(); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetFilter(context.TODO(), flt, true); err != nil {
		t.Fatal(err)
	}
	exp = map[string][]string{
		utils.FilterIDs:                {"cgrates.org:FLTR1"},
		utils.DispatcherFilterIndexIDs: {"cgrates.org:*any:*string:*req.Account:1001", "cgrates.org:*any:*string:*req.Account:1002"},
		utils.ThresholdFilterIndexIDs:  {"cgrates.org:*string:*req.Account:1001", "cgrates.org:*string:*req.Account:1002"},
	}
	rpl, err = composeCacheArgsForFilter(dm, context.TODO(), flt, tnt, flt.TenantID(), rpl)
	if err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(rpl, exp) {
		t.Errorf("Expected %s ,received: %s", utils.ToJSON(exp), utils.ToJSON(rpl))
	}
}
