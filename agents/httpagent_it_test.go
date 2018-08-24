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

package agents

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	haCfgPath string
	haCfg     *config.CGRConfig
	haRPC     *rpc.Client
	httpC     *http.Client // so we can cache the connection
)

func TestHAitInitCfg(t *testing.T) {
	haCfgPath = path.Join(*dataDir, "conf", "samples", "httpagent")
	// Init config first
	var err error
	haCfg, err = config.NewCGRConfigFromFolder(haCfgPath)
	if err != nil {
		t.Error(err)
	}
	haCfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(haCfg)
	httpC = new(http.Client)
}

// Remove data in both rating and accounting db
func TestHAitResetDB(t *testing.T) {
	if err := engine.InitDataDb(haCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDb(haCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func TestHAitStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(haCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestHAitApierRpcConn(t *testing.T) {
	var err error
	haRPC, err = jsonrpc.Dial("tcp", haCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

// Load the tariff plan, creating accounts and their balances
func TestHAitTPFromFolder(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "oldtutorial")}
	var loadInst utils.LoadInstance
	if err := haRPC.Call("ApierV2.LoadTariffPlanFromFolder", attrs, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

func TestHAitAuthDryRun(t *testing.T) {
	reqUrl := fmt.Sprintf("http://%s%s?request_type=OutboundAUTH&CallID=123456&Msisdn=497700056231&Imsi=2343000000000123&Destination=491239440004&MSRN=0102220233444488999&ProfileID=1&AgentID=176&GlobalMSISDN=497700056129&GlobalIMSI=214180000175129&ICCID=8923418450000089629&MCC=234&MNC=10&calltype=callback",
		haCfg.HTTPListen, haCfg.HttpAgentCfg()[0].Url)
	rply, err := httpC.Get(reqUrl)
	if err != nil {
		t.Error(err)
	}
	eXml := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<response>
  <Allow>1</Allow>
  <MaxDuration>1200</MaxDuration>
</response>`)
	if body, err := ioutil.ReadAll(rply.Body); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eXml, body) {
		t.Errorf("expecting: <%s>, received: <%s>", string(eXml), string(body))
	}
	rply.Body.Close()
	time.Sleep(time.Millisecond)
}

func TestHAitAuth1001(t *testing.T) {
	reqUrl := fmt.Sprintf("http://%s%s?request_type=OutboundAUTH&CallID=123456&Msisdn=1001&Imsi=2343000000000123&Destination=1002&MSRN=0102220233444488999&ProfileID=1&AgentID=176&GlobalMSISDN=497700056129&GlobalIMSI=214180000175129&ICCID=8923418450000089629&MCC=234&MNC=10&calltype=callback",
		haCfg.HTTPListen, haCfg.HttpAgentCfg()[0].Url)
	rply, err := httpC.Get(reqUrl)
	if err != nil {
		t.Error(err)
	}
	eXml := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<response>
  <Allow>1</Allow>
  <MaxDuration>6042</MaxDuration>
</response>`)
	if body, err := ioutil.ReadAll(rply.Body); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eXml, body) {
		t.Errorf("expecting: <%s>, received: <%s>", string(eXml), string(body))
	}
	rply.Body.Close()
	time.Sleep(time.Millisecond)
}

func TestHAitCDRmtcall(t *testing.T) {
	reqUrl := fmt.Sprintf("http://%s%s?request_type=MTCALL_CDR&timestamp=2018-08-14%%2012:03:22&call_date=2018-0814%%2012:00:49&transactionid=10000&CDR_ID=123456&carrierid=1&mcc=0&mnc=0&imsi=434180000000000&msisdn=1001&destination=1002&leg=C&leg_duration=185&reseller_charge=11.1605&client_charge=0.0000&user_charge=22.0000&IOT=0&user_balance=10.00&cli=%%2B498702190000&polo=0.0100&ddi_map=N",
		haCfg.HTTPListen, haCfg.HttpAgentCfg()[0].Url)
	rply, err := httpC.Get(reqUrl)
	if err != nil {
		t.Error(err)
	}
	eXml := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<CDR_RESPONSE>
  <CDR_ID>123456</CDR_ID>
  <CDR_STATUS>1</CDR_STATUS>
</CDR_RESPONSE>`)
	if body, err := ioutil.ReadAll(rply.Body); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eXml, body) {
		t.Errorf("expecting: <%s>, received: <%s>", string(eXml), string(body))
	}
	rply.Body.Close()
	time.Sleep(50 * time.Millisecond)
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{RunIDs: []string{utils.META_DEFAULT}}
	if err := haRPC.Call("ApierV2.GetCdrs", req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Usage != "3m5s" { // should be 1 but maxUsage returns rounded version
			t.Errorf("Unexpected CDR Usage received, cdr: %s ", utils.ToJSON(cdrs[0]))
		}
		if cdrs[0].Cost != 0.2188 {
			t.Errorf("Unexpected CDR Cost received, cdr: %+v ", cdrs[0].Cost)
		}
	}
}

func TestHAitStopEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
