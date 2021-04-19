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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestChargersmatchingChargerProfilesForEventErrPass(t *testing.T) {
	Cache.Clear(nil)
	defaultCfg := config.NewDefaultCGRConfig()
	defaultCfg.ChargerSCfg().IndexedSelects = false

	dbm := &DataDBMock{
		GetChargerProfileDrvF: func(s1, s2 string) (*ChargerProfile, error) {
			return &ChargerProfile{
				Tenant:    s1,
				ID:        s2,
				RunID:     utils.MetaDefault,
				FilterIDs: []string{"fltr1"},
			}, nil
		},
		GetKeysForPrefixF: func(s string) ([]string, error) {
			return []string{s + "cgrates.org:chr1"}, nil
		},
		GetFilterDrvF: func(s1, s2 string) (*Filter, error) {
			return nil, utils.ErrNotImplemented
		},
	}
	dmFilter := NewDataManager(dbm, defaultCfg.CacheCfg(), nil)
	cS := &ChargerService{
		dm: dmFilter,
		filterS: &FilterS{
			dm:  dmFilter,
			cfg: defaultCfg,
		},
		cfg: defaultCfg,
	}
	cgrEvTm := time.Date(2021, 4, 19, 12, 0, 0, 0, time.UTC)
	cgrEv := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "cgrEvID",
		Event: map[string]interface{}{
			"Charger":        "ChargerProfile1",
			utils.AnswerTime: time.Date(2021, 4, 19, 12, 0, 0, 0, time.UTC),
			"UsageInterval":  "10s",
			utils.Weight:     "10.0",
		},
		APIOpts: map[string]interface{}{
			utils.Subsys: utils.MetaChargers,
		},
		Time: &cgrEvTm,
	}

	experr := utils.ErrNotImplemented
	rcv, err := cS.matchingChargerProfilesForEvent(cgrEv.Tenant, cgrEv)

	if err == nil || err != experr {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}
}

func TestChargersmatchingChargerProfilesForEventNotActive(t *testing.T) {
	Cache.Clear(nil)
	defaultCfg := config.NewDefaultCGRConfig()
	defaultCfg.ChargerSCfg().IndexedSelects = false

	dbm := &DataDBMock{
		GetChargerProfileDrvF: func(s1, s2 string) (*ChargerProfile, error) {
			return &ChargerProfile{
				Tenant:    s1,
				ID:        s2,
				RunID:     utils.MetaDefault,
				FilterIDs: []string{"fltr1"},
				ActivationInterval: &utils.ActivationInterval{
					ActivationTime: time.Date(2021, 4, 19, 17, 0, 0, 0, time.UTC),
				},
			}, nil
		},
		GetKeysForPrefixF: func(s string) ([]string, error) {
			return []string{s + "cgrates.org:chr1"}, nil
		},
		GetFilterDrvF: func(s1, s2 string) (*Filter, error) {
			return nil, utils.ErrNotImplemented
		},
	}
	dmFilter := NewDataManager(dbm, defaultCfg.CacheCfg(), nil)
	cS := &ChargerService{
		dm: dmFilter,
		filterS: &FilterS{
			dm:  dmFilter,
			cfg: defaultCfg,
		},
		cfg: defaultCfg,
	}
	cgrEvTm := time.Date(2021, 4, 19, 12, 0, 0, 0, time.UTC)
	cgrEv := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "cgrEvID",
		Event: map[string]interface{}{
			"Charger":        "ChargerProfile1",
			utils.AnswerTime: time.Date(2021, 4, 19, 12, 0, 0, 0, time.UTC),
			"UsageInterval":  "10s",
			utils.Weight:     "10.0",
		},
		APIOpts: map[string]interface{}{
			utils.Subsys: utils.MetaChargers,
		},
		Time: &cgrEvTm,
	}

	experr := utils.ErrNotFound
	rcv, err := cS.matchingChargerProfilesForEvent(cgrEv.Tenant, cgrEv)

	if err == nil || err != experr {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}
}
