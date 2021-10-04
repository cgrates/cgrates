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
package services

import (
	"reflect"
	"sync"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/sessions"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type testMockClients struct {
	calls func(_ *context.Context, _, _ interface{}) error
}

func (sT *testMockClients) Call(ctx *context.Context, method string, arg, rply interface{}) error {
	return sT.calls(ctx, arg, rply)
}

//TestSessionSCoverage for cover testing
func TestSessionSCoverage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ChargerSCfg().Enabled = true
	cfg.CdrsCfg().Enabled = true
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	db := NewDataDBService(cfg, nil, srvDep)
	cfg.StorDbCfg().Type = utils.Internal
	anz := NewAnalyzerService(cfg, server, filterSChan, make(chan birpc.ClientConnector, 1), srvDep)
	srv := NewSessionService(cfg, db, nil, server, make(chan birpc.ClientConnector, 1), nil, anz, srvDep)
	engine.NewConnManager(cfg)
	if srv.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	srv2 := SessionService{
		cfg:      cfg,
		dm:       db,
		server:   server,
		connChan: make(chan birpc.ClientConnector, 1),
		connMgr:  nil,
		anz:      anz,
		srvDep:   srvDep,
		sm:       &sessions.SessionS{},
	}
	if !srv2.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	serviceName := srv2.ServiceName()
	if !reflect.DeepEqual(serviceName, utils.SessionS) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.SessionS, serviceName)
	}
	shouldRun := srv2.ShouldRun()
	if !reflect.DeepEqual(shouldRun, false) {
		t.Errorf("\nExpecting <false>,\n Received <%+v>", shouldRun)
	}
}
