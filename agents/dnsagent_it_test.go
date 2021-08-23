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

package agents

import (
	"net/rpc"
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
	dnsCfgDIR  string
	dnsCfg     *config.CGRConfig
	dnsRPC     *rpc.Client
	dnsClnt    *dns.Conn // so we can cache the connection

	sTestsDNS = []func(t *testing.T){
		testDNSitInitCfg,
		testDNSitResetDB,
		testDNSitStartEngine,
		testDNSitApierRpcConn,
		testDNSitTPFromFolder,
		testDNSitClntConn,
		testDNSitClntNAPTRDryRun,
		testDNSitClntNAPTRAttributes,
		testDNSitClntNAPTRSuppliers,
		testDNSitStopEngine,
	}
)

func TestDNSitSimple(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		dnsCfgDIR = "dnsagent_internal"
	case utils.MetaMySQL:
		dnsCfgDIR = "dnsagent_mysql"
	case utils.MetaMongo:
		dnsCfgDIR = "dnsagent_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsDNS {
		t.Run(dnsCfgDIR, stest)
	}
}

// Init config
func testDNSitInitCfg(t *testing.T) {
	var err error
	dnsCfgPath = path.Join(*dataDir, "conf", "samples", dnsCfgDIR)
	dnsCfg, err = config.NewCGRConfigFromPath(dnsCfgPath)
	if err != nil {
		t.Error(err)
	}
}

// Remove data in both rating and accounting db
func testDNSitResetDB(t *testing.T) {
	if err := engine.InitDataDB(dnsCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDB(dnsCfg); err != nil {
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
	dnsRPC, err = newRPCClient(dnsCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

// Load the tariff plan, creating accounts and their balances
func testDNSitTPFromFolder(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "dnsagent")}
	var loadInst utils.LoadInstance
	if err := dnsRPC.Call(utils.APIerSv2LoadTariffPlanFromFolder,
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
		if answr.Regexp != "!^(.*)$!sip:1@172.16.1.10.!" {
			t.Errorf("received: <%q>", answr.Regexp)
		}
		if answr.Replacement != "." {
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
		if answr.Regexp != "sip:1@172.16.1.1." {
			t.Errorf("Expected :<%q> , received: <%q>", "sip:1\\@172.16.1.1.", answr.Regexp)
		}
	}
}

func testDNSitClntNAPTRSuppliers(t *testing.T) {
	m := new(dns.Msg)
	m.SetQuestion("5.6.9.4.7.1.7.1.5.6.8.9.4.e164.arpa.", dns.TypeNAPTR)
	if err := dnsClnt.WriteMsg(m); err != nil {
		t.Error(err)
	}
	rply, err := dnsClnt.ReadMsg()
	if err != nil {
		t.Error(err)
	} else if len(rply.Answer) != 2 {
		t.Fatalf("wrong number of records: %s", utils.ToIJSON(rply.Answer))
	}
	if rply.Rcode != dns.RcodeSuccess {
		t.Errorf("failed to get an valid answer\n%v", rply)
	}
	answr := rply.Answer[0].(*dns.NAPTR)
	if answr.Order != 100 {
		t.Errorf("received: <%v>", answr.Order)
	}
	if answr.Regexp != "!^(.*)$!sip:1@172.16.1.11!" {
		t.Errorf("received: <%q>", answr.Regexp)
	}
	answr2 := rply.Answer[1].(*dns.NAPTR)
	if answr2.Regexp != "!^(.*)$!sip:1@172.16.1.12!" {
		t.Errorf("received: <%q>", answr2.Regexp)
	}
}

func testDNSitStopEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
