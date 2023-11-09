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
package general_tests

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/radigo"
)

func TestRadiusAllRound(t *testing.T) {
	var cfgDir string
	switch *utils.DBType {
	case utils.MetaInternal:
		cfgDir = "radagent_internal"
	case utils.MetaMySQL:
		cfgDir = "radagent_mysql"
	case utils.MetaMongo:
		cfgDir = "radagent_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	testEnv := engine.TestEngine{
		ConfigPath: path.Join(*utils.DataDir, "conf", "samples", cfgDir),
		TpPath:     path.Join(*utils.DataDir, "tariffplans", "oldtutorial"),
	}
	client, _ := testEnv.Run(t)
	var raAuthClnt, raAcctClnt *radigo.Client
	t.Run("CheckListenersPorts", func(t *testing.T) {
		ports := []int{1812, 1813}
		for _, port := range ports {
			if !checkPortOpenTCP(port) {
				t.Errorf("TCP Port %d is not open", port)
			}
			if !checkPortOpenUDP(port) {
				t.Errorf("UDP port %d is not open", port)
			}
		}
	})

	t.Run("ReloadConfigOK", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.ConfigSv1ReloadConfig, &config.ReloadArgs{
			Tenant:  "cgrates.org",
			Path:    path.Join(*utils.DataDir, "conf", "samples", cfgDir),
			Section: config.RA_JSN,
			DryRun:  true,
		}, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("Expected OK received: %s", reply)
		}
	})
	t.Run("Request", func(t *testing.T) {
		dictRad, err := radigo.NewDictionaryFromFoldersWithRFC2865([]string{"/usr/share/cgrates/radius/dict"})
		if err != nil {
			t.Error(err)
		}
		if raAuthClnt, err = radigo.NewClient(utils.UDP, "127.0.0.1:1812", "CGRateS.org", dictRad, 1, nil, utils.Logger); err != nil {
			t.Fatal(err)
		}
		authReq := raAuthClnt.NewRequest(radigo.AccessRequest, 1) // emulates Kamailio packet out of radius_load_caller_avps()
		if err := authReq.AddAVPWithName("User-Name", "1001", ""); err != nil {
			t.Error(err)
		}
		if err := authReq.AddAVPWithName("User-Password", "CGRateSPassword1", ""); err != nil {
			t.Error(err)
		}
		// encode the password as required so we can decode it properly
		authReq.AVPs[1].RawValue = radigo.EncodeUserPassword([]byte("CGRateSPassword1"), []byte("CGRateS.org"), authReq.Authenticator[:])
		if err := authReq.AddAVPWithName("Service-Type", "SIP-Caller-AVPs", ""); err != nil {
			t.Error(err)
		}
		if err := authReq.AddAVPWithName("Called-Station-Id", "1002", ""); err != nil {
			t.Error(err)
		}
		if err := authReq.AddAVPWithName("Acct-Session-Id", "e4921177ab0e3586c37f6a185864b71a@0:0:0:0:0:0:0:0", ""); err != nil {
			t.Error(err)
		}
		if err := authReq.AddAVPWithName("Sip-From-Tag", "51585361", ""); err != nil {
			t.Error(err)
		}
		if err := authReq.AddAVPWithName("NAS-IP-Address", "127.0.0.1", ""); err != nil {
			t.Error(err)
		}
		if err := authReq.AddAVPWithName("Event-Timestamp", "1497106115", ""); err != nil {
			t.Error(err)
		}
		reply, err := raAuthClnt.SendRequest(authReq)
		if err != nil {
			t.Fatal(err)
		}
		if reply.Code != radigo.AccessAccept {
			t.Errorf("Received reply: %+v", utils.ToJSON(reply))
		}
		if len(reply.AVPs) != 1 { // make sure max duration is received
			t.Errorf("Received AVPs: %+v", utils.ToJSON(reply.AVPs))
		} else if !bytes.Equal([]byte("session_max_time#10800"), reply.AVPs[0].RawValue) {
			t.Errorf("Received: %s", string(reply.AVPs[0].RawValue))
		}

		if raAuthClnt, err = radigo.NewClient(utils.TCP, "127.0.0.1:1812", "CGRateS.org", dictRad, 1, nil, utils.Logger); err != nil {
			t.Fatal(err)
		}
		authReq = raAuthClnt.NewRequest(radigo.AccessRequest, 1) // emulates Kamailio packet out of radius_load_caller_avps()
		if err := authReq.AddAVPWithName("User-Name", "1001", ""); err != nil {
			t.Error(err)
		}
		if err := authReq.AddAVPWithName("User-Password", "CGRateSPassword1", ""); err != nil {
			t.Error(err)
		}
		// encode the password as required so we can decode it properly
		authReq.AVPs[1].RawValue = radigo.EncodeUserPassword([]byte("CGRateSPassword1"), []byte("CGRateS.org"), authReq.Authenticator[:])
		if err := authReq.AddAVPWithName("Service-Type", "SIP-Caller-AVPs", ""); err != nil {
			t.Error(err)
		}
		if err := authReq.AddAVPWithName("Called-Station-Id", "1002", ""); err != nil {
			t.Error(err)
		}
		if err := authReq.AddAVPWithName("Acct-Session-Id", "e4921177ab0e3586c37f6a185864b71a@0:0:0:0:0:0:0:0", ""); err != nil {
			t.Error(err)
		}
		if err := authReq.AddAVPWithName("Sip-From-Tag", "51585361", ""); err != nil {
			t.Error(err)
		}
		if err := authReq.AddAVPWithName("NAS-IP-Address", "127.0.0.1", ""); err != nil {
			t.Error(err)
		}
		if err := authReq.AddAVPWithName("Event-Timestamp", "1497106115", ""); err != nil {
			t.Error(err)
		}
		reply, err = raAuthClnt.SendRequest(authReq)
		if err != nil {
			t.Fatal(err)
		}
		if reply.Code != radigo.AccessAccept {
			t.Errorf("Received reply: %+v", utils.ToJSON(reply))
		}
		if len(reply.AVPs) != 1 { // make sure max duration is received
			t.Errorf("Received AVPs: %+v", utils.ToJSON(reply.AVPs))
		} else if !bytes.Equal([]byte("session_max_time#10800"), reply.AVPs[0].RawValue) {
			t.Errorf("Received: %s", string(reply.AVPs[0].RawValue))
		}

		if raAcctClnt, err = radigo.NewClient(utils.UDP, "127.0.0.1:1813", "CGRateS.org", dictRad, 1, nil, utils.Logger); err != nil {
			t.Fatal(err)
		}
		req := raAcctClnt.NewRequest(radigo.AccountingRequest, 2) // emulates Kamailio packet for accounting start
		if err := req.AddAVPWithName("Acct-Status-Type", "Start", ""); err != nil {
			t.Error(err)
		}
		if err := req.AddAVPWithName("Service-Type", "Sip-Session", ""); err != nil {
			t.Error(err)
		}
		if err := req.AddAVPWithName("Sip-Response-Code", "200", ""); err != nil {
			t.Error(err)
		}
		if err := req.AddAVPWithName("Sip-Method", "Invite", ""); err != nil {
			t.Error(err)
		}
		if err := req.AddAVPWithName("Event-Timestamp", "1497106115", ""); err != nil {
			t.Error(err)
		}
		if err := req.AddAVPWithName("Sip-From-Tag", "51585361", ""); err != nil {
			t.Error(err)
		}
		if err := req.AddAVPWithName("Sip-To-Tag", "75c2f57b", ""); err != nil {
			t.Error(err)
		}
		if err := req.AddAVPWithName("Acct-Session-Id", "e4921177ab0e3586c37f6a185864b71a@0:0:0:0:0:0:0:0", ""); err != nil {
			t.Error(err)
		}
		if err := req.AddAVPWithName("User-Name", "1001", ""); err != nil {
			t.Error(err)
		}
		if err := req.AddAVPWithName("Called-Station-Id", "1002", ""); err != nil {
			t.Error(err)
		}
		if err := req.AddAVPWithName("Ascend-User-Acct-Time", "1497106115", ""); err != nil {
			t.Error(err)
		}
		if err := req.AddAVPWithName("NAS-Port", "5060", ""); err != nil {
			t.Error(err)
		}
		if err := req.AddAVPWithName("Acct-Delay-Time", "0", ""); err != nil {
			t.Error(err)
		}
		if err := req.AddAVPWithName("NAS-IP-Address", "127.0.0.1", ""); err != nil {
			t.Error(err)
		}
		reply, err = raAcctClnt.SendRequest(req)
		if err != nil {
			t.Error(err)
		}
		if reply.Code != radigo.AccountingResponse {
			t.Errorf("Received reply: %+v", reply)
		}
		if len(reply.AVPs) != 0 { // we don't expect AVPs to be populated
			t.Errorf("Received AVPs: %+v", reply.AVPs)
		}
		// Make sure the sessin is managed by SMG
		time.Sleep(10 * time.Millisecond)
		expUsage := 10 * time.Second
		var aSessions []*sessions.ExternalSession
		if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions,
			utils.SessionFilter{
				Filters: []string{
					fmt.Sprintf("*string:~*req.%s:%s", utils.RunID, utils.MetaDefault),
					fmt.Sprintf("*string:~*req.%s:%s", utils.OriginID, "e4921177ab0e3586c37f6a185864b71a@0:0:0:0:0:0:0:0-51585361-75c2f57b"),
				},
			}, &aSessions); err != nil {
			t.Error(err)
		} else if len(aSessions) != 1 {
			t.Errorf("Unexpected number of sessions received: %+v", aSessions)
		} else if aSessions[0].Usage != expUsage {
			t.Errorf("Expecting %v, received usage: %v\nAnd Session: %s ", expUsage, aSessions[0].Usage, utils.ToJSON(aSessions))
		}

		if raAcctClnt, err = radigo.NewClient(utils.TCP, "127.0.0.1:1813", "CGRateS.org", dictRad, 1, nil, utils.Logger); err != nil {
			t.Fatal(err)
		}
		req = raAcctClnt.NewRequest(radigo.AccountingRequest, 2) // emulates Kamailio packet for accounting start
		if err := req.AddAVPWithName("Acct-Status-Type", "Start", ""); err != nil {
			t.Error(err)
		}
		if err := req.AddAVPWithName("Service-Type", "Sip-Session", ""); err != nil {
			t.Error(err)
		}
		if err := req.AddAVPWithName("Sip-Response-Code", "200", ""); err != nil {
			t.Error(err)
		}
		if err := req.AddAVPWithName("Sip-Method", "Invite", ""); err != nil {
			t.Error(err)
		}
		if err := req.AddAVPWithName("Event-Timestamp", "1497106115", ""); err != nil {
			t.Error(err)
		}
		if err := req.AddAVPWithName("Sip-From-Tag", "51585361", ""); err != nil {
			t.Error(err)
		}
		if err := req.AddAVPWithName("Sip-To-Tag", "75c2f57b", ""); err != nil {
			t.Error(err)
		}
		if err := req.AddAVPWithName("Acct-Session-Id", "e4921177ab0e3586c37f6a185864b71a@0:0:0:0:0:0:0:0", ""); err != nil {
			t.Error(err)
		}
		if err := req.AddAVPWithName("User-Name", "1001", ""); err != nil {
			t.Error(err)
		}
		if err := req.AddAVPWithName("Called-Station-Id", "1002", ""); err != nil {
			t.Error(err)
		}
		if err := req.AddAVPWithName("Ascend-User-Acct-Time", "1497106115", ""); err != nil {
			t.Error(err)
		}
		if err := req.AddAVPWithName("NAS-Port", "5060", ""); err != nil {
			t.Error(err)
		}
		if err := req.AddAVPWithName("Acct-Delay-Time", "0", ""); err != nil {
			t.Error(err)
		}
		if err := req.AddAVPWithName("NAS-IP-Address", "127.0.0.1", ""); err != nil {
			t.Error(err)
		}
		reply, err = raAcctClnt.SendRequest(req)
		if err != nil {
			t.Error(err)
		}
		if reply.Code != radigo.AccountingResponse {
			t.Errorf("Received reply: %+v", reply)
		}
		if len(reply.AVPs) != 0 { // we don't expect AVPs to be populated
			t.Errorf("Received AVPs: %+v", reply.AVPs)
		}
		// Make sure the sessin is managed by SMG
		time.Sleep(10 * time.Millisecond)
		expUsage = 10 * time.Second

		if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions,
			utils.SessionFilter{
				Filters: []string{
					fmt.Sprintf("*string:~*req.%s:%s", utils.RunID, utils.MetaDefault),
					fmt.Sprintf("*string:~*req.%s:%s", utils.OriginID, "e4921177ab0e3586c37f6a185864b71a@0:0:0:0:0:0:0:0-51585361-75c2f57b"),
				},
			}, &aSessions); err != nil {
			t.Error(err)
		} else if len(aSessions) != 1 {
			t.Errorf("Unexpected number of sessions received: %+v", aSessions)
		} else if aSessions[0].Usage != expUsage {
			t.Errorf("Expecting %v, received usage: %v\nAnd Session: %s ", expUsage, aSessions[0].Usage, utils.ToJSON(aSessions))
		}
	})

	t.Run("StopRadiusService", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.ServiceManagerV1StopService, &servmanager.ArgStartService{
			ServiceID: utils.MetaRadius,
		}, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("Expected OK received: %s", reply)
		}
	})
	t.Run("CheckListenersPortsOff", func(t *testing.T) {
		ports := []int{1812, 1813}
		for _, port := range ports {
			if checkPortOpenTCP(port) {
				t.Errorf("TCP Port %d is open when it should'nt be", port)
			}
			if checkPortOpenUDP(port) {
				t.Errorf("UDP port %d is open when it should'nt be", port)
			}
		}
	})

	t.Run("CreateTmpBadConfig", func(t *testing.T) {
		tmpFolderPath := filepath.Join("/tmp", "test_folder")
		err := os.Mkdir(tmpFolderPath, os.ModePerm)
		if err != nil {
			t.Fatal("Failed to create temporary folder:", err)
		}

		filePath := filepath.Join(tmpFolderPath, "cgrates.json")
		jsonConfig := `{
// CGRateS Configuration file
//

"general": {
	"log_level": 6,
},


"listen": {
	"rpc_json": ":2012",				// RPC JSON listening address
	"rpc_gob": ":2013",					// RPC GOB listening address
	"http": ":2080",					// HTTP listening address
},


"data_db": {
	"db_type": "*internal",	
},


"stor_db": {
	"db_type": "*internal",	
},

"rals": {
	"enabled": true,
},

"schedulers": {
	"enabled": true,
},

"cdrs": {
	"enabled": true,
	"rals_conns": ["*internal"],
},

"resources": {
	"enabled": true,
	"store_interval": "-1",
},

"attributes": {
	"enabled": true,
},

"routes": {
	"enabled": true,
},

"chargers": {
	"enabled": true,
},

"sessions": {
	"enabled": true,
	"attributes_conns": ["*localhost"],
	"cdrs_conns": ["*localhost"],
	"rals_conns": ["*localhost"],
	"resources_conns": ["*localhost"],
	"chargers_conns": ["*internal"],
	"debit_interval": "10s",
},

"radius_agent": {
	"enabled": true,
	"sessions_conns": ["*localhost"],
	"listeners":[
		{
			"network": "udp",
			"auth_address": "127.0.0.1:1812",
			"acct_address": "127.0.0.1:1813"
		},
		{
			"network": "tcp",
			"auth_address": "badAddress",
			"acct_address": "127.0.0.1:1813"
		},
	],	
},



"apiers": {
	"enabled": true,
	"scheduler_conns": ["*internal"],
},


}
`
		err = writeToFile(filePath, jsonConfig)
		if err != nil {
			t.Fatal("Failed to write JSON content to file:", err)
		}
		time.Sleep(100 * time.Millisecond)
	})
	t.Run("StartRadiusService", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.ServiceManagerV1StartService, &servmanager.ArgStartService{
			ServiceID: utils.MetaRadius,
		}, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("Expected OK received: %s", reply)
		}
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("CheckListenersPorts", func(t *testing.T) {
		ports := []int{1812, 1813}
		for _, port := range ports {
			if !checkPortOpenTCP(port) {
				t.Errorf("TCP Port %d is not open", port)
			}
			if !checkPortOpenUDP(port) {
				t.Errorf("UDP port %d is not open", port)
			}
		}
	})

	t.Run("ReloadConfigOK", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.ConfigSv1ReloadConfig, &config.ReloadArgs{
			Tenant:  "cgrates.org",
			Path:    "/tmp/test_folder",
			Section: config.RA_JSN,
		}, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("Expected OK received: %s", reply)
		}
		deleteTmpFolder("/tmp/test_folder", t)
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("CheckListenersPortsOff", func(t *testing.T) {
		ports := []int{1812, 1813}
		for _, port := range ports {
			if checkPortOpenTCP(port) {
				t.Errorf("TCP Port %d is open when it should'nt be", port)
			}
			if checkPortOpenUDP(port) {
				t.Errorf("UDP port %d is open when it should'nt be", port)
			}
		}
	})

}

func checkPortOpenTCP(port int) bool {
	conn, err := net.Dial("tcp", net.JoinHostPort("127.0.0.1", fmt.Sprintf("%d", port)))
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

func checkPortOpenUDP(port int) bool {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return true
	}
	defer conn.Close()
	return false
}
