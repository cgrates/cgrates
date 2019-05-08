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
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/miekg/dns"
)

var (
	dnsCfgPath string
	dnsCfg     *config.CGRConfig
	dnsRPC     *rpc.Client
	dnsClnt    *dns.Conn // so we can cache the connection
)

var sTestsDNS = []func(t *testing.T){
	testDNSitResetDB,
	testDNSitStartEngine,
	testDNSitApierRpcConn,
	testDNSitTPFromFolder,
	testDNSitClntConn,
	//testDNSitClntNAPTRDryRun,
	testDNSitClntNAPTRAttributes,
	//testDNSitClntNAPTRSuppliers,
	testDNSitStopEngine,
}

func TestDNSitSimple(t *testing.T) {
	dnsCfgPath = path.Join(*dataDir, "conf", "samples", "dnsagent")
	// Init config first
	var err error
	dnsCfg, err = config.NewCGRConfigFromPath(dnsCfgPath)
	if err != nil {
		t.Error(err)
	}
	dnsCfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(dnsCfg)
	for _, stest := range sTestsDNS {
		t.Run("dnsAgent", stest)
	}
}

// Remove data in both rating and accounting db
func testDNSitResetDB(t *testing.T) {
	if err := engine.InitDataDb(dnsCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDb(dnsCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testDNSitStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(dnsCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testDNSitApierRpcConn(t *testing.T) {
	var err error
	dnsRPC, err = jsonrpc.Dial("tcp", dnsCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

// Load the tariff plan, creating accounts and their balances
func testDNSitTPFromFolder(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "dnsagent")}
	var loadInst utils.LoadInstance
	if err := dnsRPC.Call(utils.ApierV2LoadTariffPlanFromFolder,
		attrs, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

// Connect DNS client to server
func testDNSitClntConn(t *testing.T) {
	c := new(dns.Client)
	var err error
	if dnsClnt, err = c.Dial(dnsCfg.DNSAgentCfg().Listen); err != nil { // just testing the connection, not not saving it
		t.Fatal(err)
	} else if dnsClnt == nil {
		t.Fatalf("conn is nil")
	}
}

func testDNSitClntNAPTRDryRun(t *testing.T) {
	m := new(dns.Msg)
	m.SetQuestion("3.6.9.4.7.1.7.1.5.6.8.9.4.e164.arpa.", dns.TypeNAPTR)
	if err := dnsClnt.WriteMsg(m); err != nil {
		t.Error(err)
	}
	if rply, err := dnsClnt.ReadMsg(); err != nil {
		t.Error(err)
	} else {
		if rply.Rcode != dns.RcodeSuccess {
			t.Errorf("failed to get an valid answer\n%v", rply)
		}
		answr := rply.Answer[0].(*dns.NAPTR)
		if answr.Order != 100 {
			t.Errorf("received: <%q>", answr.Order)
		}
		if answr.Preference != 10 {
			t.Errorf("received: <%q>", answr.Preference)
		}
		if answr.Flags != "U" {
			t.Errorf("received: <%q>", answr.Flags)
		}
		if answr.Service != "E2U+SIP" {
			t.Errorf("received: <%q>", answr.Service)
		}
		if answr.Regexp != "^.*$" {
			t.Errorf("received: <%q>", answr.Regexp)
		}
		if answr.Replacement != "sip:1\\@172.16.1.10." {
			t.Errorf("received: <%q>", answr.Replacement)
		}
	}
}

func testDNSitClntNAPTRAttributes(t *testing.T) {
	m := new(dns.Msg)
	m.SetQuestion("4.6.9.4.7.1.7.1.5.6.8.9.4.e164.arpa.", dns.TypeNAPTR)
	if err := dnsClnt.WriteMsg(m); err != nil {
		t.Error(err)
	}
	if rply, err := dnsClnt.ReadMsg(); err != nil {
		t.Error(err)
	} else {
		if rply.Rcode != dns.RcodeSuccess {
			t.Errorf("failed to get an valid answer\n%v", rply)
		}
		answr := rply.Answer[0].(*dns.NAPTR)
		if answr.Order != 100 {
			t.Errorf("received: <%q>", answr.Order)
		}
		if answr.Replacement != "sip:\\1@172.16.1.1." {
			t.Errorf("received: <%q>", answr.Replacement)
		}
	}
}

func testDNSitClntNAPTRSuppliers(t *testing.T) {
	m := new(dns.Msg)
	m.SetQuestion("5.6.9.4.7.1.7.1.5.6.8.9.4.e164.arpa.", dns.TypeNAPTR)
	if err := dnsClnt.WriteMsg(m); err != nil {
		t.Error(err)
	}
	if rply, err := dnsClnt.ReadMsg(); err != nil {
		t.Error(err)
	} else {
		if rply.Rcode != dns.RcodeSuccess {
			t.Errorf("failed to get an valid answer\n%v", rply)
		}
		answr := rply.Answer[0].(*dns.NAPTR)
		if answr.Order != 100 {
			t.Errorf("received: <%q>", answr.Order)
		}
		if answr.Replacement != "sip:1\\@172.16.1.10." {
			t.Errorf("received: <%q>", answr.Replacement)
		}
	}
}

func testDNSitStopEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
