//go:build integration
// +build integration

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
	"fmt"
	"path"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

var (
	srvMngCfgPath   string
	srvMngCfg       *config.CGRConfig
	srvMngRPC       *birpc.Client
	srvMngConfigDIR string //run tests for specific configuration

	sTestsServManager = []func(t *testing.T){
		testSrvMngInitCfg,
		testSrvMngInitDataDb,
		testSrvMngResetStorDb,
		testSrvMngStartEngine,
		testSrvMngSRPCConn,

		testSrvMngPing,

		testSrvMngSKillEngine,
	}
)

func TestServManagerIT(t *testing.T) {
	t.SkipNow()
	switch *utils.DBType {
	case utils.MetaInternal:
		srvMngConfigDIR = "apis_srvmng_internal"
	case utils.MetaMongo:
		srvMngConfigDIR = "apis_srvmng_mongo"
	case utils.MetaMySQL:
		srvMngConfigDIR = "apis_srvmng_mysql"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsServManager {
		t.Run(srvMngConfigDIR, stest)
	}
}

func testSrvMngInitCfg(t *testing.T) {
	var err error
	srvMngCfgPath = path.Join(*utils.DataDir, "conf", "samples", srvMngConfigDIR)
	srvMngCfg, err = config.NewCGRConfigFromPath(context.Background(), srvMngCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testSrvMngInitDataDb(t *testing.T) {
	if err := engine.InitDataDB(srvMngCfg); err != nil {
		t.Fatal(err)
	}
}

func testSrvMngResetStorDb(t *testing.T) {
	if err := engine.InitStorDB(srvMngCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testSrvMngStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(srvMngCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testSrvMngSRPCConn(t *testing.T) {
	srvMngRPC = engine.NewRPCClient(t, srvMngCfg.ListenCfg(), *utils.Encoding)
}

// Kill the engine when it is about to be finished
func testSrvMngSKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}

func testSrvMngPing(t *testing.T) {
	var reply string
	if err := srvMngRPC.Call(context.Background(), utils.ServiceManagerV1Ping, nil,
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Unexpected reply: %s", reply)
	}

	serviceToMethod := map[string]string{
		utils.AdminS:     utils.AdminSv1Ping,
		utils.AccountS:   utils.AccountSv1Ping,
		utils.ActionS:    utils.ActionSv1Ping,
		utils.AnalyzerS:  utils.AnalyzerSv1Ping,
		utils.AttributeS: utils.AttributeSv1Ping,
		utils.CDRServer:  utils.CDRsV1Ping,
		utils.ChargerS:   utils.ChargerSv1Ping,
		// utils.DispatcherS: utils.DispatcherSv1Ping,
		utils.EEs:        utils.EeSv1Ping,
		utils.EFs:        utils.EfSv1Ping,
		utils.RateS:      utils.RateSv1Ping,
		utils.ResourceS:  utils.ResourceSv1Ping,
		utils.RouteS:     utils.RouteSv1Ping,
		utils.SessionS:   utils.SessionSv1Ping,
		utils.StatS:      utils.StatSv1Ping,
		utils.ThresholdS: utils.ThresholdSv1Ping,
		utils.TPeS:       utils.TPeSv1Ping,
	}

	/*
		Run the following tests for each service:

		- ping before enabling service (expect can't find service error)
		- query for service status (expect service to be in state "SERVICE_DOWN")
		- start the service
		- query for service status (expect service to be in state "SERVICE_UP")
		- ping (expect "Pong")
		- stop service
		- query for service status (expect service to be in state "SERVICE_DOWN" reply)
		- ping after stopping service (expect "can't find service" error)
	*/

	var statusReply map[string]string
	for serviceID, serviceMethod := range serviceToMethod {
		t.Run(fmt.Sprintf("test for service %s", serviceID), func(t *testing.T) {
			rpcErr := fmt.Sprintf("rpc: can't find service %s", serviceMethod)
			if err := srvMngRPC.Call(context.Background(), serviceMethod, nil,
				&reply); err == nil || err.Error() != rpcErr {
				t.Errorf("expected: <%+v>,\nreceived: <%+v>", rpcErr, err)
			}

			args := servmanager.ArgsServiceID{
				ServiceID: serviceID,
			}

			if err := srvMngRPC.Call(context.Background(), utils.ServiceManagerV1ServiceStatus, args,
				&statusReply); err != nil {
				t.Error(err)
			} else if statusReply[serviceID] != utils.StateServiceDOWN {
				t.Errorf("Unexpected reply: %s", utils.ToJSON(statusReply))
			}

			if err := srvMngRPC.Call(context.Background(), utils.ServiceManagerV1StartService, args,
				&reply); err != nil {
				t.Error(err)
			} else if reply != utils.OK {
				t.Errorf("Unexpected reply: %s", reply)
			}

			if err := srvMngRPC.Call(context.Background(), utils.ServiceManagerV1ServiceStatus, args,
				&statusReply); err != nil {
				t.Error(err)
			} else if statusReply[serviceID] != utils.StateServiceUP {
				t.Errorf("Unexpected reply: %s", utils.ToJSON(statusReply))
			}

			if err := srvMngRPC.Call(context.Background(), serviceMethod, nil,
				&reply); err != nil {
				t.Error(err)
			} else if reply != utils.Pong {
				t.Errorf("Unexpected reply: %s", reply)
			}

			if err := srvMngRPC.Call(context.Background(), utils.ServiceManagerV1StopService, args,
				&reply); err != nil {
				t.Error(err)
			} else if reply != utils.OK {
				t.Errorf("Unexpected reply: %s", reply)
			}

			if err := srvMngRPC.Call(context.Background(), utils.ServiceManagerV1ServiceStatus, args,
				&statusReply); err != nil {
				t.Error(err)
			} else if statusReply[serviceID] != utils.StateServiceDOWN {
				t.Errorf("Unexpected reply: %s", utils.ToJSON(statusReply))
			}

			if err := srvMngRPC.Call(context.Background(), serviceMethod, nil,
				&reply); err == nil || err.Error() != rpcErr {
				t.Errorf("expected: <%+v>,\nreceived: <%+v>", rpcErr, err)
			}
		})
	}
}
