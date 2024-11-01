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

package cores

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestNewCoreService(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	cfgDflt.CoreSCfg().CapsStatsInterval = time.Second
	stopchan := make(chan struct{}, 1)
	caps := engine.NewCaps(1, utils.MetaBusy)
	sts := engine.NewCapsStats(cfgDflt.CoreSCfg().CapsStatsInterval, caps, stopchan)
	expected := &CoreS{
		cfg:       cfgDflt,
		CapsStats: sts,
		caps:      caps,
	}
	rcv := NewCoreService(cfgDflt, caps, nil, stopchan, nil, nil)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}
	//shut down the service
	rcv.shtDw = func() {}
	rcv.Shutdown()
	rcv.ShutdownEngine()
}
