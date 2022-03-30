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

package apis

import (
	"reflect"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestCDRsProcessEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	connMgr := engine.NewConnManager(cfg)
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	cdrS := engine.NewCDRServer(cfg, dm, engine.NewFilterS(cfg, connMgr, dm), connMgr)
	cdr := NewCDRsV1(cdrS)
	var reply string
	args := &utils.CGREvent{
		ID:     "TestMatchingAccountsForEvent",
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
		},
	}

	if err := cdr.ProcessEvent(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
	if reply != utils.OK {
		t.Errorf("Expected %v\n but received %v", utils.OK, reply)
	}
}

func TestCDRsProcessEventWithGet(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	connMgr := engine.NewConnManager(cfg)
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	cdrS := engine.NewCDRServer(cfg, dm, engine.NewFilterS(cfg, connMgr, dm), connMgr)
	cdr := NewCDRsV1(cdrS)
	var reply []*utils.EventsWithOpts
	args := &utils.CGREvent{
		ID:     "TestMatchingAccountsForEvent",
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
		},
	}

	if err := cdr.ProcessEventWithGet(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
	exp := []utils.EventsWithOpts{
		{
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
			Opts: map[string]interface{}{},
		},
	}
	if !reflect.DeepEqual(exp[0].Event, reply[0].Event) {
		t.Errorf("Expected %v \n but received %v", exp, reply)
	}
}
