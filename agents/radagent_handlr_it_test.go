//go:build flaky

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package agents

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/radigo"
)

var (
	raHCfgPath  string
	raHCfg      *config.CGRConfig
	raHAuthClnt *radigo.Client
	raHRPC      *rpc.Client

	sTestsRadiusH = []func(t *testing.T){
		testRAHitRemoveFolders,
		testRAHitInitCfg,
		testRAHitResetDataDb,
		testRAHitResetStorDb,
		testRAHitStartEngine,
		testRAHitApierRpcConn,
		testRAHitTPFromFolder,
		testRAHitEmptyValueHandling,
		testRAHitStopCgrEngine,
		testRAHitRemoveFolders,
	}
)

// Test start here
func TestRAHandlerPlaygroundit(t *testing.T) {
	for _, stest := range sTestsRadiusH {
		t.Run("raonfigDIR", stest)
	}
}

func testRAHitRemoveFolders(t *testing.T) {
	if err := os.RemoveAll("/tmp/TestRAitEmptyValueHandling"); err != nil {
		t.Error(err)
	}

}

func testRAHitInitCfg(t *testing.T) {

	content := `{
		// CGRateS Configuration file
		//
		
		"general": {
			"log_level": 7,
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
		
		"suppliers": {
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
			"request_processors": [
			{
				"id": "RadiusMandatoryFail",
				"filters": ["*string:~*vars.*radReqType:*radAuth","*string:~*req.User-Name:10011"],
				"flags": ["*log", "*auth", "*attributes"],
				"request_fields":[
				  {"tag": "UserName", "path": "*cgreq.RadUserName", "type": "*composed", 
					"value": "~*req.User-Name"},
				  {"tag": "Password", "path": "*cgreq.RadPassword", "type": "*composed", 
					"value": "~*req.User-Password"},
				  {"tag": "ReplyMessage", "path": "*cgreq.RadReplyMessage", "type": "*constant",
					"value": "*attributes"},
				],
				"reply_fields":[
				  {"tag": "Code", "path": "*rep.*radReplyCode", "filters": ["*notempty:~*cgrep.Error:"],
					"type": "*constant", "value": "AccessReject"},
			      {"tag": "ReplyMessage", "path": "*rep.Reply-Message","filters": ["*notempty:~*cgrep.Error:"],
			 	   "type": "*composed", "value": "~*cgrep.Error"},
				],
			  },
			],
		},
		
		
		
		"apiers": {
			"enabled": true,
			"scheduler_conns": ["*internal"],
		},
		
		
		}
		`

	folderNameSuffix, err := rand.Int(rand.Reader, big.NewInt(10000))
	if err != nil {
		t.Fatalf("could not generate random number for folder name suffix, err: %s", err.Error())
	}
	raHCfgPath = fmt.Sprintf("/tmp/config%d", folderNameSuffix)
	err = os.MkdirAll(raHCfgPath, 0755)
	if err != nil {
		t.Fatal(err)
	}
	filePath := filepath.Join(raHCfgPath, "cgrates.json")
	err = os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}
	raHCfg, err = config.NewCGRConfigFromPath(raHCfgPath)
	if err != nil {
		t.Error(err)
	}

	raHCfg.DataFolderPath = raHCfgPath // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(raHCfg)
}

// Remove data in both rating and accounting db
func testRAHitResetDataDb(t *testing.T) {
	if err := engine.InitDataDb(raHCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testRAHitResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(raHCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testRAHitStartEngine(t *testing.T) {
	if _, err := engine.StartEngine(raHCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testRAHitApierRpcConn(t *testing.T) {
	var err error
	switch *utils.Encoding {
	case utils.MetaJSON:
		raHRPC, err = jsonrpc.Dial(utils.TCP, raHCfg.ListenCfg().RPCJSONListen)
	case utils.MetaGOB:
		raHRPC, err = rpc.Dial(utils.TCP, raHCfg.ListenCfg().RPCGOBListen)
	default:
		err = errors.New("UNSUPPORTED_RPC")
	}
	if err != nil {
		t.Fatal(err)
	}
}

// Load the tariff plan, creating accounts and their balances
func testRAHitTPFromFolder(t *testing.T) {
	writeFile := func(fileName, data string) error {
		err := os.MkdirAll("/tmp/TestRAitEmptyValueHandling", os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}
		csvFile, err := os.Create(path.Join("/tmp/TestRAitEmptyValueHandling", fileName))
		if err != nil {
			return err
		}
		defer csvFile.Close()
		_, err = csvFile.WriteString(data)
		if err != nil {
			return err

		}
		return csvFile.Sync()
	}

	// Create and populate Attributes.csv
	if err := writeFile(utils.AttributesCsv, `
#Tenant,ID,Contexts,FilterIDs,ActivationInterval,AttributeFilterIDs,Path,Type,Value,Blocker,Weight
cgrates.org,ATTR_RAD,*any,*string:~*req.RadUserName:10011;*prefix:~*req.RadPassword:CGRateSPassword3,,,,,,false,10
`); err != nil {
		t.Fatal(err)
	}

	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: "/tmp/TestRAitEmptyValueHandling"}
	if err := raHRPC.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	time.Sleep(500 * time.Millisecond)
}

func testRAHitEmptyValueHandling(t *testing.T) {
	var err error
	if raHAuthClnt, err = radigo.NewClient("udp", "127.0.0.1:1812", "CGRateS.org", dictRad, 1, nil, nil); err != nil {
		t.Fatal(err)
	}
	authReq := raHAuthClnt.NewRequest(radigo.AccessRequest, 1) // emulates Kamailio packet out of radius_load_caller_avps()
	if err := authReq.AddAVPWithName("User-Name", "10011", ""); err != nil {
		t.Error(err)
	}
	if err := authReq.AddAVPWithName("User-Password", "CGRateSPassword3", ""); err != nil {
		t.Error(err)
	}
	// encode the password as required so we can decode it properly
	authReq.AVPs[1].RawValue = radigo.EncodeUserPassword([]byte("CGRateSPassword3"), []byte("CGRateS.org"), authReq.Authenticator[:])
	reply, err := raHAuthClnt.SendRequest(authReq)
	if err != nil {
		t.Fatal(err)
	}
	if reply.Code != radigo.AccessReject {
		t.Errorf("Received reply: %+v", reply)
	}
	exp := "ATTRIBUTES_ERROR:" + utils.MandatoryIEMissingCaps + ": [RadReplyMessage]"
	if len(reply.AVPs) != 1 { // make sure max duration is received
		t.Errorf("Received AVPs: %+v", reply.AVPs)
	} else if exp != string(reply.AVPs[0].RawValue) {
		t.Errorf("Expected <%+v>, Received: <%+v>", exp, string(reply.AVPs[0].RawValue))
	}
}

func testRAHitStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
	if err := os.RemoveAll("/tmp/" + raHCfgPath); err != nil {
		t.Error(err)
	}
}
