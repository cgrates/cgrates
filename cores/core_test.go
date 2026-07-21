// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	shutdown := utils.NewSyncedChan()
	expected := &CoreS{
		cfg:       cfgDflt,
		CapsStats: sts,
		caps:      caps,
		shutdown:  shutdown,
	}
	rcv := NewCoreService(cfgDflt, caps, nil, stopchan, nil, shutdown, nil)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}
	//shut down the service
	rcv.Shutdown()
	rcv.ShutdownEngine()
	select {
	case <-shutdown.Done():
	default:
		t.Error("engine did not shut down")
	}
}
