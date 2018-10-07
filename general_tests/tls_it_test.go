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
package general_tests

import (
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

var (
	tlsCfgPath       string
	tlsCfg           *config.CGRConfig
	tlsRpcClientJson *rpcclient.RpcClient
	tlsRpcClientGob  *rpcclient.RpcClient
	tlsHTTPJson      *rpcclient.RpcClient
	tlsConfDIR       string //run tests for specific configuration
	tlsDelay         int
)

var sTestsTLS = []func(t *testing.T){
	testTLSLoadConfig,
	testTLSInitDataDb,
	testTLSStartEngine,
	testTLSRpcConn,
	testTLSPing,
	testTLSStopEngine,
}

// Test start here
func TestTLS(t *testing.T) {
	tlsConfDIR = "tls"
	for _, stest := range sTestsTLS {
		t.Run(tlsConfDIR, stest)
	}
}

func testTLSLoadConfig(t *testing.T) {
	var err error
	tlsCfgPath = path.Join(*dataDir, "conf", "samples", tlsConfDIR)
	if tlsCfg, err = config.NewCGRConfigFromFolder(tlsCfgPath); err != nil {
		t.Error(err)
	}
	tlsDelay = 2000
}

func testTLSInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(tlsCfg); err != nil {
		t.Fatal(err)
	}
}

func testTLSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tlsCfgPath, tlsDelay); err != nil {
		t.Fatal(err)
	}
}

func testTLSRpcConn(t *testing.T) {
	var err error
	tlsRpcClientJson, err = rpcclient.NewRpcClient("tcp", "localhost:2022", tlsCfg.TLSClientKey,
		tlsCfg.TLSClientCerificate, 3, 3,
		time.Duration(1*time.Second), time.Duration(5*time.Minute), utils.JSON, nil, false)
	if err != nil {
		t.Errorf("Error: %s when dialing", err)
	}

	tlsRpcClientGob, err = rpcclient.NewRpcClient("tcp", "localhost:2023", tlsCfg.TLSClientKey,
		tlsCfg.TLSClientCerificate, 3, 3,
		time.Duration(1*time.Second), time.Duration(5*time.Minute), utils.GOB, nil, false)
	if err != nil {
		t.Errorf("Error: %s when dialing", err)
	}

	tlsHTTPJson, err = rpcclient.NewRpcClient("tcp", "https://localhost:2280/jsonrpc", tlsCfg.TLSClientKey,
		tlsCfg.TLSClientCerificate, 3, 3,
		time.Duration(1*time.Second), time.Duration(5*time.Minute), rpcclient.JSON_HTTP, nil, false)
	if err != nil {
		t.Errorf("Error: %s when dialing", err)
	}
}

func testTLSPing(t *testing.T) {
	var reply string

	if err := tlsRpcClientJson.Call(utils.ThresholdSv1Ping, "", &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	if err := tlsRpcClientGob.Call(utils.ThresholdSv1Ping, "", &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	if err := tlsHTTPJson.Call(utils.ThresholdSv1Ping, "", &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	if err := tlsRpcClientJson.Call(utils.DispatcherSv1Ping, "", &reply); err == nil {
		t.Error(err)
	}
	if err := tlsRpcClientGob.Call(utils.DispatcherSv1Ping, "", &reply); err == nil {
		t.Error(err)
	}
	if err := tlsHTTPJson.Call(utils.DispatcherSv1Ping, "", &reply); err == nil {
		t.Error(err)
	}
}

func testTLSStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
