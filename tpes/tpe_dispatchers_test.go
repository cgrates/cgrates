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

package tpes

import (
	"bytes"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestTPEnewTPDispatchers(t *testing.T) {
	// dataDB := &engine.DataDBM
	// dm := &engine.NewDataManager()
	cfg := config.NewDefaultCGRConfig()
	connMng := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(&engine.DataDBMock{
		GetDispatcherProfileDrvF: func(ctx *context.Context, str1 string, str2 string) (*engine.DispatcherProfile, error) {
			dsp := &engine.DispatcherProfile{
				Tenant:    "cgrates.org",
				ID:        "Dsp1",
				FilterIDs: []string{"*string:~*req.Account:1001", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
				Strategy:  utils.MetaFirst,
				StrategyParams: map[string]interface{}{
					utils.MetaDefaultRatio: "false",
				},
				Weight: 20,
				Hosts: engine.DispatcherHostProfiles{
					{
						ID:        "C1",
						FilterIDs: []string{},
						Weight:    10,
						Params:    map[string]interface{}{"0": "192.168.54.203"},
						Blocker:   false,
					},
				},
			}
			return dsp, nil
		},
	}, nil, connMng)
	exp := &TPDispatchers{
		dm: dm,
	}
	rcv := newTPDispatchers(dm)
	if rcv.dm != exp.dm {
		t.Errorf("Expected %v \nbut received %v", exp, rcv)
	}
}

func TestTPEExportItemsDispatchers(t *testing.T) {
	wrtr := new(bytes.Buffer)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	tpDsp := TPDispatchers{
		dm: dm,
	}
	dsp := &engine.DispatcherProfile{
		Tenant:    "cgrates.org",
		ID:        "Dsp1",
		FilterIDs: []string{"*string:~*req.Account:1001", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
		Strategy:  utils.MetaFirst,
		StrategyParams: map[string]interface{}{
			utils.MetaDefaultRatio: "false",
		},
		Weight: 20,
		Hosts: engine.DispatcherHostProfiles{
			{
				ID:        "C1",
				FilterIDs: []string{},
				Weight:    10,
				Params:    map[string]interface{}{"0": "192.168.54.203"},
				Blocker:   false,
			},
		},
	}
	tpDsp.dm.SetDispatcherProfile(context.Background(), dsp, false)
	err := tpDsp.exportItems(context.Background(), wrtr, "cgrates.org", []string{"Dsp1"})
	if err != nil {
		t.Errorf("Expected nil\n but received %v", err)
	}
}

func TestTPEExportItemsDispatchersNoDbConn(t *testing.T) {
	engine.Cache.Clear(nil)
	wrtr := new(bytes.Buffer)
	tpDsp := TPDispatchers{
		dm: nil,
	}
	dsp := &engine.DispatcherProfile{
		Tenant:    "cgrates.org",
		ID:        "Dsp1",
		FilterIDs: []string{"*string:~*req.Account:1001", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
		Strategy:  utils.MetaFirst,
		StrategyParams: map[string]interface{}{
			utils.MetaDefaultRatio: "false",
		},
		Weight: 20,
		Hosts: engine.DispatcherHostProfiles{
			{
				ID:        "C1",
				FilterIDs: []string{},
				Weight:    10,
				Params:    map[string]interface{}{"0": "192.168.54.203"},
				Blocker:   false,
			},
		},
	}
	tpDsp.dm.SetDispatcherProfile(context.Background(), dsp, false)
	err := tpDsp.exportItems(context.Background(), wrtr, "cgrates.org", []string{"Dsp1"})
	if err != utils.ErrNoDatabaseConn {
		t.Errorf("Expected %v\n but received %v", utils.ErrNoDatabaseConn, err)
	}
}

func TestTPEExportItemsDispatchersIDNotFound(t *testing.T) {
	wrtr := new(bytes.Buffer)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	tpDsp := TPDispatchers{
		dm: dm,
	}
	dsp := &engine.DispatcherProfile{
		Tenant:    "cgrates.org",
		ID:        "Dsp1",
		FilterIDs: []string{"*string:~*req.Account:1001", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
		Strategy:  utils.MetaFirst,
		StrategyParams: map[string]interface{}{
			utils.MetaDefaultRatio: "false",
		},
		Weight: 20,
		Hosts: engine.DispatcherHostProfiles{
			{
				ID:        "C1",
				FilterIDs: []string{},
				Weight:    10,
				Params:    map[string]interface{}{"0": "192.168.54.203"},
				Blocker:   false,
			},
		},
	}
	tpDsp.dm.SetDispatcherProfile(context.Background(), dsp, false)
	err := tpDsp.exportItems(context.Background(), wrtr, "cgrates.org", []string{"Dsp2"})
	errExpect := "<DSP_PROFILE_NOT_FOUND> cannot find DispatcherProfile with id: <Dsp2>"
	if err.Error() != errExpect {
		t.Errorf("Expected %v\n but received %v", errExpect, err)
	}
}
