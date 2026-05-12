//go:build integration
// +build integration

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

package general_tests

import (
	// 	"bytes"
	// 	"crypto/tls"
	// 	"crypto/x509"
	// 	"fmt"
	// 	"io"
	// 	"net/http"

	"net/http"
	"os"
	"path"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/loaders"
	"github.com/cgrates/cgrates/utils"
	// 	"os"
	// 	"path"
	// 	"reflect"
	// 	"testing"
	// 	"time"
	// 	"github.com/cgrates/cgrates/config"
	// 	"github.com/cgrates/cgrates/engine"
	// 	"github.com/cgrates/cgrates/utils"
)

var (
	httpSsCfgPath string
	httpSsCfgDIR  string
	httpSsCfg     *config.CGRConfig
	httpSsRPC     *birpc.Client
	httpSsClnt    *http.Client // so we can cache the connection
	err           error

	httpSsTests = []func(t *testing.T){
		testHttpSsInitCfg,
		testHttpSsHttpClnt,
		// 		testHAitResetDB,
		testHttpSsStartEngine,
		testHttpSsRPC,
		testHttpSsLoadTPFromFolder,
		// 		testHAitAuthDryRun,
		// 		testHAitAuth1001,
		// 		testHAitCDRmtcall,
		// 		testHAitCDRmtcall2,
		// 		testHAitTextPlain,
		testHttpSsStopEngine,
	}
)

func TestHttpSessionsIt(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		httpSsCfgDIR = "httpss_internal"
	case utils.MetaRedis:
		t.SkipNow()
	case utils.MetaMySQL:
		httpSsCfgDIR = "httpagent_mysql"
	case utils.MetaMongo:
		httpSsCfgDIR = "httpagent_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range httpSsTests {
		t.Run(httpSsCfgDIR, stest)
	}

}

// // Init config first
func testHttpSsInitCfg(t *testing.T) {
	var err error
	if err := os.RemoveAll("/tmp/TestHttpSessionsIt/loader/in"); err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll("/tmp/TestHttpSessionsIt/loader/in", 0755); err != nil {
		t.Fatal(err)
	}
	httpSsCfgPath = path.Join(*utils.DataDir, "conf", "samples", httpSsCfgDIR)
	httpSsCfg, err = config.NewCGRConfigFromPath(context.Background(), httpSsCfgPath)
	if err != nil {
		t.Error(err)
	}
}

// Initialize the HttpClient so we can cache it's connections
func testHttpSsHttpClnt(t *testing.T) {
	httpSsClnt = new(http.Client)
}

// // Start CGR Engine
func testHttpSsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(httpSsCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

// // Connect RPC client to rater
func testHttpSsRPC(t *testing.T) {
	httpSsRPC = engine.NewRPCClient(t, httpSsCfg.ListenCfg(), *utils.Encoding) // We connect over JSON so we can also troubleshoot if needed
}

// Load the data from offline TP
func testHttpSsLoadTPFromFolder(t *testing.T) {
	var reply string
	if err := httpSsRPC.Call(context.Background(), utils.LoaderSv1Run,
		&loaders.ArgsProcessFolder{
			Path: path.Join(*utils.DataDir, "tariffplans", "sessions")}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}
	//time.Sleep(time.Duration(*utils.WaitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

// func testHAitAuthDryRun(t *testing.T) {
// 	httpSsClntonst := "http"
// 	addr := httpSsCfg.ListenCfg().HTTPListen
// 	if isTls {
// 		addr = httpSsCfg.ListenCfg().HTTPTLSListen
// 		httpSsClntonst = "https"
// 	}
// 	reqUrl := fmt.Sprintf("%s://%s%s?request_type=OutboundAUTH&CallID=123456&Msisdn=497700056231&Imsi=2343000000000123&Destination=491239440004&MSRN=0102220233444488999&ProfileID=1&AgentID=176&GlobalMSISDN=497700056129&GlobalIMSI=214180000175129&ICCID=8923418450000089629&MCC=234&MNC=10&calltype=callback",
// 		httpSsClntonst, addr, httpSsCfg.HTTPAgentCfg()[0].URL)
// 	rply, err := httpSsClnt.Get(reqUrl)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	eXml := []byte(`<?xml version="1.0" encoding="UTF-8"?>
// <response>
//   <Allow>1</Allow>
//   <Concatenated>234/Val1</Concatenated>
//   <MaxDuration>1200</MaxDuration>
// </response>`)
// 	if body, err := io.ReadAll(rply.Body); err != nil {
// 		t.Error(err)
// 	} else if !reflect.DeepEqual(eXml, body) {
// 		t.Errorf("expecting: <%s>, received: <%s>", string(eXml), string(body))
// 	}
// 	rply.Body.Close()
// 	time.Sleep(time.Millisecond)
// }

// func testHAitAuth1001(t *testing.T) {
// 	acnt := "20002"
// 	maxDuration := 60

// 	attrSetBalance := utils.AttrSetBalance{
// 		Tenant:      "cgrates.org",
// 		Account:     acnt,
// 		BalanceType: utils.MetaVoice,
// 		Value:       float64(maxDuration) * float64(time.Second),
// 		Balance: map[string]any{
// 			utils.ID:            "TestDynamicDebitBalance",
// 			utils.RatingSubject: "*zero5ms",
// 		},
// 	}
// 	var reply string
// 	if err := httpSsRPC.Call(utils.APIerSv2SetBalance, attrSetBalance, &reply); err != nil {
// 		t.Error(err)
// 	} else if reply != utils.OK {
// 		t.Errorf("Received: %s", reply)
// 	}
// 	time.Sleep(10 * time.Millisecond)

// 	httpSsClntonst := "http"
// 	addr := httpSsCfg.ListenCfg().HTTPListen
// 	if isTls {
// 		addr = httpSsCfg.ListenCfg().HTTPTLSListen
// 		httpSsClntonst = "https"
// 	}

// 	reqUrl := fmt.Sprintf("%s://%s%s?request_type=OutboundAUTH&CallID=123456&Msisdn=%s&Imsi=2343000000000123&Destination=1002&MSRN=0102220233444488999&ProfileID=1&AgentID=176&GlobalMSISDN=497700056129&GlobalIMSI=214180000175129&ICCID=8923418450000089629&MCC=234&MNC=10&calltype=callback",
// 		httpSsClntonst, addr, httpSsCfg.HTTPAgentCfg()[0].URL, acnt)
// 	rply, err := httpSsClnt.Get(reqUrl)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	eXml := []byte(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
// <response>
//   <Allow>1</Allow>
//   <MaxDuration>%v</MaxDuration>
// </response>`, maxDuration))
// 	if body, err := io.ReadAll(rply.Body); err != nil {
// 		t.Error(err)
// 	} else if !reflect.DeepEqual(eXml, body) {
// 		t.Errorf("expecting: %s, received: %s", string(eXml), string(body))
// 	}
// 	rply.Body.Close()
// 	time.Sleep(time.Millisecond)
// }

// func testHAitCDRmtcall(t *testing.T) {
// 	httpSsClntonst := "http"
// 	addr := httpSsCfg.ListenCfg().HTTPListen
// 	if isTls {
// 		addr = httpSsCfg.ListenCfg().HTTPTLSListen
// 		httpSsClntonst = "https"
// 	}
// 	reqUrl := fmt.Sprintf("%s://%s%s?request_type=MTCALL_CDR&timestamp=2018-08-14%%2012:03:22&call_date=2018-0814%%2012:00:49&transactionid=10000&CDR_ID=123456&carrierid=1&mcc=0&mnc=0&imsi=434180000000000&msisdn=1001&destination=1002&leg=C&leg_duration=185&reseller_charge=11.1605&client_charge=0.0000&user_charge=22.0000&IOT=0&user_balance=10.00&cli=%%2B498702190000&polo=0.0100&ddi_map=N",
// 		httpSsClntonst, addr, httpSsCfg.HTTPAgentCfg()[0].URL)
// 	rply, err := httpSsClnt.Get(reqUrl)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	eXml := []byte(`<?xml version="1.0" encoding="UTF-8"?>
// <CDR_RESPONSE>
//   <CDR_ID>123456</CDR_ID>
//   <CDR_STATUS>1</CDR_STATUS>
// </CDR_RESPONSE>`)
// 	if body, err := io.ReadAll(rply.Body); err != nil {
// 		t.Error(err)
// 	} else if !reflect.DeepEqual(eXml, body) {
// 		t.Errorf("expecting: <%s>, received: <%s>", string(eXml), string(body))
// 	}
// 	rply.Body.Close()
// 	time.Sleep(50 * time.Millisecond)
// 	var cdrs []*engine.ExternalCDR
// 	req := utils.RPCCDRsFilter{RunIDs: []string{utils.MetaDefault}}
// 	if err := httpSsRPC.Call(utils.APIerSv2GetCDRs, &req, &cdrs); err != nil {
// 		t.Error("Unexpected error: ", err.Error())
// 	} else if len(cdrs) != 1 {
// 		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
// 	} else {
// 		if cdrs[0].Usage != "3m5s" { // should be 1 but maxUsage returns rounded version
// 			t.Errorf("Unexpected CDR Usage received, cdr: %s ", utils.ToJSON(cdrs[0]))
// 		}
// 		if utils.Round(cdrs[0].Cost, 4, utils.MetaRoundingMiddle) != 0.2188 { // sql have only 4 digits after decimal point
// 			t.Errorf("Unexpected CDR Cost received, cdr: %+v ", cdrs[0].Cost)
// 		}
// 		if cdrs[0].OriginHost != "127.0.0.1" {
// 			t.Errorf("Unexpected CDR OriginHost received, cdr: %+v ", cdrs[0].OriginHost)
// 		}
// 	}
// }

// func testHAitCDRmtcall2(t *testing.T) {
// 	xmlBody := `<?xml version="1.0" encoding="utf-8"?><complete-datasession-notification callid="48981764"><createtime>2005-08-26T14:17:34</createtime><reference>Data</reference><userid>528594</userid><username>447700086788</username><customerid>510163</customerid><companyname>Silliname</companyname><totalcost amount="0.1400" currency="USD">0.1400</totalcost><agenttotalcost amount="0.1400" currency="USD">0.1400</agenttotalcost><agentid>234</agentid><callleg calllegid="89357336"><number>447700086788</number><description>China, Peoples Republic of - China Unicom (CU-GSM)</description><mcc>460</mcc><mnc>001</mnc><seconds>32</seconds><bytes>4558</bytes><permegabyterate  currency="USD">1.3330</permegabyterate><cost amount="0.1400" currency="USD">0.1400</cost><agentpermegabyterate currency="USD">1.3330</agentpermegabyterate><agentcost amount="0.1400"currency="USD">0.1400</agentcost></callleg></complete-datasession-notification>`
// 	httpSsClntonst := "http"
// 	addr := httpSsCfg.ListenCfg().HTTPListen
// 	if isTls {
// 		addr = httpSsCfg.ListenCfg().HTTPTLSListen
// 		httpSsClntonst = "https"
// 	}
// 	url := fmt.Sprintf("%s://%s%s", httpSsClntonst, addr, httpSsCfg.HTTPAgentCfg()[1].URL)

// 	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(xmlBody)))
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	req.Header.Add("Content-Type", "application/xml; charset=utf-8")
// 	resp, err := httpSsClnt.Do(req)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	resp.Body.Close()

// 	time.Sleep(50 * time.Millisecond)
// 	var cdrs []*engine.ExternalCDR
// 	fltr := utils.RPCCDRsFilter{RunIDs: []string{utils.MetaDefault}, Accounts: []string{"447700086788"}}
// 	if err := httpSsRPC.Call(utils.APIerSv2GetCDRs, &fltr, &cdrs); err != nil {
// 		t.Error("Unexpected error: ", err.Error())
// 	} else if len(cdrs) != 1 {
// 		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
// 	} else {
// 		if cdrs[0].Usage != "4558" { // should be 1 but maxUsage returns rounded version
// 			t.Errorf("Unexpected CDR Usage received, cdr: %s ", utils.ToJSON(cdrs[0]))
// 		}
// 	}
// }

// func testHAitTextPlain(t *testing.T) {
// 	httpSsClntonst := "http"
// 	addr := httpSsCfg.ListenCfg().HTTPListen
// 	if isTls {
// 		addr = httpSsCfg.ListenCfg().HTTPTLSListen
// 		httpSsClntonst = "https"
// 	}
// 	reqUrl := fmt.Sprintf("%s://%s%s?request_type=TextPlainDryRun&CallID=123456&Msisdn=497700056231&Imsi=2343000000000123&Destination=491239440004",
// 		httpSsClntonst, addr, httpSsCfg.HTTPAgentCfg()[2].URL)
// 	rply, err := httpSsClnt.Get(reqUrl)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	response := []byte(`Variable1=Hola1
// Variable2=Hola2
// ComposedVar=TestComposed
// Item1.1=Val2
// Item1.1=Val1
// `)
// 	if body, err := io.ReadAll(rply.Body); err != nil {
// 		t.Error(err)
// 	} else if !reflect.DeepEqual(len(response), len(body)) {
// 		t.Errorf("expecting: \n<%s>\n, received: \n<%s>\n", string(response), string(body))
// 	}
// 	rply.Body.Close()
// 	time.Sleep(time.Millisecond)
// }

func testHttpSsStopEngine(t *testing.T) {
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}
