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
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestCoreSStatus(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(2, utils.MetaTopUp)
	coreService := cores.NewCoreService(cfg, caps, nil, utils.EmptyString, make(chan struct{}), nil, nil, nil)
	cS := NewCoreSv1(coreService)
	arg := &utils.TenantWithAPIOpts{
		Tenant:  "cgrates.org",
		APIOpts: map[string]interface{}{},
	}
	var reply map[string]interface{}
	if err := cS.Status(context.Background(), arg, &reply); err != nil {
		t.Error(err)
	}
}

func TestCoreSSleep(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(2, utils.MetaTopUp)
	coreService := cores.NewCoreService(cfg, caps, nil, utils.EmptyString, make(chan struct{}), nil, nil, nil)
	cS := NewCoreSv1(coreService)
	arg := &utils.DurationArgs{
		Duration: 1 * time.Millisecond,
	}
	var reply string
	if err := cS.Sleep(context.Background(), arg, &reply); err != nil {
		t.Error(err)
	} else if reply != "OK" {
		t.Errorf("Expected OK, received %+v", reply)
	}
}

func TestCoreSShutdown(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(2, utils.MetaTopUp)
	shdChan := utils.NewSyncedChan()
	coreService := cores.NewCoreService(cfg, caps, nil, utils.EmptyString, make(chan struct{}), nil, nil, shdChan)
	cS := NewCoreSv1(coreService)
	arg := &utils.CGREvent{}
	var reply string
	if err := cS.Shutdown(context.Background(), arg, &reply); err != nil {
		t.Error(err)
	} else if reply != "OK" {
		t.Errorf("Expected OK, received %+v", reply)
	}
}
