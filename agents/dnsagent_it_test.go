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
	"crypto/tls"
	"path"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/miekg/dns"
)

type dnsConns struct {
	UDP    *dns.Conn
	TCP    *dns.Conn
	TCPTLS *dns.Conn
}

var (
	dnsCfgPath string
	dnsCfgDIR  string
	dnsCfg     *config.CGRConfig
	dnsRPC     *birpc.Client
	dnsClnt    *dnsConns // so we can cache the connection

	sTestsDNS = []func(t *testing.T){
		testDNSitInitCfg,
		testDNSitResetDB,
		testDNSitStartEngine,
		testDNSitApierRpcConn,
		testDNSitTPFromFolder,
		testDNSitClntConn,
		testDNSitClntADryRun,
		testDNSitClntSRVDryRun,
		testDNSitClntNAPTRDryRun,
		testDNSitClntAAttributes,
		testDNSitClntSRVAttributes,
		testDNSitClntNAPTRAttributes,
		testDNSitClntASuppliers,
		testDNSitClntSRVSuppliers,
		testDNSitClntNAPTRSuppliers,
		testDNSitClntAOpts,
		testDNSitClntSRVOpts,
		testDNSitClntNAPTROpts,
		testDNSitClntAOptsWithAttributes,
		testDNSitClntSRVOptsWithAttributes,
		testDNSitClntNAPTROptsWithAttributes,
		testDNSitStopEngine,
	}
)

func TestDNSitSimple(t *testing.T) {
	switch *utils.DBType {
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
	dnsCfgPath = path.Join(*utils.DataDir, "conf", "samples", dnsCfgDIR)
	dnsCfg, err = config.NewCGRConfigFromPath(dnsCfgPath)
	if err != nil {
		t.Error(err)
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
	if _, err := engine.StopStartEngine(dnsCfgPath, *utils.WaitRater); err != nil {
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
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "dnsagent")}
	var loadInst utils.LoadInstance
	if err := dnsRPC.Call(context.Background(), utils.APIerSv2LoadTariffPlanFromFolder,
		attrs, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*utils.WaitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

// Connect DNS client to server
func testDNSitClntConn(t *testing.T) {
	c := new(dns.Client)
	c.TLSConfig = &tls.Config{}
	var err error
	dnsClnt = &dnsConns{}
	c.Net = dnsCfg.DNSAgentCfg().Listeners[0].Network
	if dnsClnt.UDP, err = c.Dial(dnsCfg.DNSAgentCfg().Listeners[0].Address); err != nil { // just testing the connection, not saving it
		t.Fatal(err)
	} else if dnsClnt.UDP == nil {
		t.Fatalf("conn is nil")
	}

	c.Net = dnsCfg.DNSAgentCfg().Listeners[1].Network

	if dnsClnt.TCP, err = c.Dial(dnsCfg.DNSAgentCfg().Listeners[1].Address); err != nil { // tcp has the same address as udp in this case so we can use the same here
		t.Fatal(err)
	} else if dnsClnt.TCP == nil {
		t.Fatalf("conn is nil")
	}

	c.Net = dnsCfg.DNSAgentCfg().Listeners[2].Network
	c.TLSConfig.InsecureSkipVerify = true
	if dnsClnt.TCPTLS, err = c.Dial(dnsCfg.DNSAgentCfg().Listeners[2].Address); err != nil { // tcp and tcp-tls cannot be on the same address, otherwise tcp-tls and udp is allowed
		t.Fatal(err)
	} else if dnsClnt.TCPTLS == nil {
		t.Fatalf("conn is nil")
	}
}

func testDNSitClntADryRun(t *testing.T) {
	m := new(dns.Msg)
	m.SetQuestion("cgrates.org.", dns.TypeA)
	if err := dnsClnt.UDP.WriteMsg(m); err != nil {
		t.Error(err)
	}
	if rply, err := dnsClnt.UDP.ReadMsg(); err != nil {
		t.Error(err)
	} else if len(rply.Answer) != 1 {
		t.Fatalf("wrong number of records: %s", utils.ToIJSON(rply.Answer))
	} else {
		if rply.Rcode != dns.RcodeSuccess {
			t.Errorf("failed to get an valid answer\n%v", rply)
		}
		answr0 := rply.Answer[0].(*dns.A)
		if answr0.A.String() != "51.38.77.188" {
			t.Errorf("Expected :<%q> , received: <%q>", "51.38.77.188", answr0.A)
		}
	}
	if err := dnsClnt.TCP.WriteMsg(m); err != nil {
		t.Error(err)
	}
	if rply, err := dnsClnt.TCP.ReadMsg(); err != nil {
		t.Error(err)
	} else if len(rply.Answer) != 1 {
		t.Fatalf("wrong number of records: %s", utils.ToIJSON(rply.Answer))
	} else {
		if rply.Rcode != dns.RcodeSuccess {
			t.Errorf("failed to get an valid answer\n%v", rply)
		}
		answr0 := rply.Answer[0].(*dns.A)
		if answr0.A.String() != "51.38.77.188" {
			t.Errorf("Expected :<%q> , received: <%q>", "51.38.77.188", answr0.A)
		}
	}
	if err := dnsClnt.TCPTLS.WriteMsg(m); err != nil {
		t.Error(err)
	}
	if rply, err := dnsClnt.TCPTLS.ReadMsg(); err != nil {
		t.Error(err)
	} else if len(rply.Answer) != 1 {
		t.Fatalf("wrong number of records: %s", utils.ToIJSON(rply.Answer))
	} else {
		if rply.Rcode != dns.RcodeSuccess {
			t.Errorf("failed to get an valid answer\n%v", rply)
		}
		answr0 := rply.Answer[0].(*dns.A)
		if answr0.A.String() != "51.38.77.188" {
			t.Errorf("Expected :<%q> , received: <%q>", "51.38.77.188", answr0.A)
		}
	}
}
func testDNSitClntSRVDryRun(t *testing.T) {
	m := new(dns.Msg)
	m.SetQuestion("_sip._tcp.opensips.org.", dns.TypeSRV)
	if err := dnsClnt.UDP.WriteMsg(m); err != nil {
		t.Error(err)
	}
	if rply, err := dnsClnt.UDP.ReadMsg(); err != nil {
		t.Error(err)
	} else if len(rply.Answer) != 1 {
		t.Fatalf("wrong number of records: %s", utils.ToIJSON(rply.Answer))
	} else {
		if rply.Rcode != dns.RcodeSuccess {
			t.Errorf("failed to get an valid answer\n%v", rply)
		}
		answr := rply.Answer[0].(*dns.SRV)
		if answr.Priority != uint16(0) {
			t.Errorf("Expected :<%q> , received: <%q>", uint16(0), answr.Priority)
		}
		if answr.Weight != uint16(50) {
			t.Errorf("Expected :<%q> , received: <%q>", uint16(50), answr.Weight)
		}
		if answr.Port != uint16(5060) {
			t.Errorf("Expected :<%q> , received: <%q>", uint16(5060), answr.Port)
		}
		if answr.Target != "opensips.org." {
			t.Errorf("Expected :<%q> , received: <%q>", "opensips.org.", answr.Target)
		}
	}
	if err := dnsClnt.TCP.WriteMsg(m); err != nil {
		t.Error(err)
	}
	if rply, err := dnsClnt.TCP.ReadMsg(); err != nil {
		t.Error(err)
	} else if len(rply.Answer) != 1 {
		t.Fatalf("wrong number of records: %s", utils.ToIJSON(rply.Answer))
	} else {
		if rply.Rcode != dns.RcodeSuccess {
			t.Errorf("failed to get an valid answer\n%v", rply)
		}
		answr := rply.Answer[0].(*dns.SRV)
		if answr.Priority != uint16(0) {
			t.Errorf("Expected :<%q> , received: <%q>", uint16(0), answr.Priority)
		}
		if answr.Weight != uint16(50) {
			t.Errorf("Expected :<%q> , received: <%q>", uint16(50), answr.Weight)
		}
		if answr.Port != uint16(5060) {
			t.Errorf("Expected :<%q> , received: <%q>", uint16(5060), answr.Port)
		}
		if answr.Target != "opensips.org." {
			t.Errorf("Expected :<%q> , received: <%q>", "opensips.org.", answr.Target)
		}
	}
	if err := dnsClnt.TCPTLS.WriteMsg(m); err != nil {
		t.Error(err)
	}
	if rply, err := dnsClnt.TCPTLS.ReadMsg(); err != nil {
		t.Error(err)
	} else if len(rply.Answer) != 1 {
		t.Fatalf("wrong number of records: %s", utils.ToIJSON(rply.Answer))
	} else {
		if rply.Rcode != dns.RcodeSuccess {
			t.Errorf("failed to get an valid answer\n%v", rply)
		}
		answr := rply.Answer[0].(*dns.SRV)
		if answr.Priority != uint16(0) {
			t.Errorf("Expected :<%q> , received: <%q>", uint16(0), answr.Priority)
		}
		if answr.Weight != uint16(50) {
			t.Errorf("Expected :<%q> , received: <%q>", uint16(50), answr.Weight)
		}
		if answr.Port != uint16(5060) {
			t.Errorf("Expected :<%q> , received: <%q>", uint16(5060), answr.Port)
		}
		if answr.Target != "opensips.org." {
			t.Errorf("Expected :<%q> , received: <%q>", "opensips.org.", answr.Target)
		}
	}
}

func testDNSitClntNAPTRDryRun(t *testing.T) {
	m := new(dns.Msg)
	m.SetQuestion("4.6.9.4.7.1.7.1.5.6.8.9.4.e164.arpa.", dns.TypeNAPTR)
	if err := dnsClnt.UDP.WriteMsg(m); err != nil {
		t.Error(err)
	}
	if rply, err := dnsClnt.UDP.ReadMsg(); err != nil {
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
	if err := dnsClnt.TCP.WriteMsg(m); err != nil {
		t.Error(err)
	}
	if rply, err := dnsClnt.TCP.ReadMsg(); err != nil {
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
	if err := dnsClnt.TCPTLS.WriteMsg(m); err != nil {
		t.Error(err)
	}
	if rply, err := dnsClnt.TCPTLS.ReadMsg(); err != nil {
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

func testDNSitClntAAttributes(t *testing.T) {
	m := new(dns.Msg)
	m.SetQuestion("dns.google.", dns.TypeA)
	if err := dnsClnt.UDP.WriteMsg(m); err != nil {
		t.Error(err)
	}
	if rply, err := dnsClnt.UDP.ReadMsg(); err != nil {
		t.Error(err)
	} else if len(rply.Answer) != 2 {
		t.Fatalf("wrong number of records: %s", utils.ToIJSON(rply.Answer))
	} else {
		if rply.Rcode != dns.RcodeSuccess {
			t.Errorf("failed to get an valid answer\n%v", rply)
		}
		answr0 := rply.Answer[0].(*dns.A)
		if answr0.A.String() != "8.8.8.8" {
			t.Errorf("Expected :<%q> , received: <%q>", "8.8.8.8", answr0.A)
		}
		answr1 := rply.Answer[1].(*dns.A)
		if answr1.A.String() != "8.8.4.4" {
			t.Errorf("Expected :<%q> , received: <%q>", "8.8.4.4", answr1.A)
		}
	}
	if err := dnsClnt.TCP.WriteMsg(m); err != nil {
		t.Error(err)
	}
	if rply, err := dnsClnt.TCP.ReadMsg(); err != nil {
		t.Error(err)
	} else if len(rply.Answer) != 2 {
		t.Fatalf("wrong number of records: %s", utils.ToIJSON(rply.Answer))
	} else {
		if rply.Rcode != dns.RcodeSuccess {
			t.Errorf("failed to get an valid answer\n%v", rply)
		}
		answr0 := rply.Answer[0].(*dns.A)
		if answr0.A.String() != "8.8.8.8" {
			t.Errorf("Expected :<%q> , received: <%q>", "8.8.8.8", answr0.A)
		}
		answr1 := rply.Answer[1].(*dns.A)
		if answr1.A.String() != "8.8.4.4" {
			t.Errorf("Expected :<%q> , received: <%q>", "8.8.4.4", answr1.A)
		}
	}
	if err := dnsClnt.TCPTLS.WriteMsg(m); err != nil {
		t.Error(err)
	}
	if rply, err := dnsClnt.TCPTLS.ReadMsg(); err != nil {
		t.Error(err)
	} else if len(rply.Answer) != 2 {
		t.Fatalf("wrong number of records: %s", utils.ToIJSON(rply.Answer))
	} else {
		if rply.Rcode != dns.RcodeSuccess {
			t.Errorf("failed to get an valid answer\n%v", rply)
		}
		answr0 := rply.Answer[0].(*dns.A)
		if answr0.A.String() != "8.8.8.8" {
			t.Errorf("Expected :<%q> , received: <%q>", "8.8.8.8", answr0.A)
		}
		answr1 := rply.Answer[1].(*dns.A)
		if answr1.A.String() != "8.8.4.4" {
			t.Errorf("Expected :<%q> , received: <%q>", "8.8.4.4", answr1.A)
		}
	}
}

func testDNSitClntSRVAttributes(t *testing.T) {
	m := new(dns.Msg)
	m.SetQuestion("_ldap._tcp.google.com.", dns.TypeSRV)
	if err := dnsClnt.UDP.WriteMsg(m); err != nil {
		t.Error(err)
	}
	if rply, err := dnsClnt.UDP.ReadMsg(); err != nil {
		t.Error(err)
	} else {
		if rply.Rcode != dns.RcodeSuccess {
			t.Errorf("failed to get an valid answer\n%v", rply)
		}
		answr := rply.Answer[0].(*dns.SRV)
		if answr.Priority != uint16(5) {
			t.Errorf("Expected :<%q> , received: <%q>", uint16(5), answr.Priority)
		}
		if answr.Weight != uint16(0) {
			t.Errorf("Expected :<%q> , received: <%q>", uint16(0), answr.Weight)
		}
		if answr.Port != uint16(389) {
			t.Errorf("Expected :<%q> , received: <%q>", uint16(389), answr.Port)
		}
		if answr.Target != "ldap.google.com." {
			t.Errorf("Expected :<%q> , received: <%q>", "ldap.google.com.", answr.Target)
		}
	}
	if err := dnsClnt.TCP.WriteMsg(m); err != nil {
		t.Error(err)
	}
	if rply, err := dnsClnt.TCP.ReadMsg(); err != nil {
		t.Error(err)
	} else {
		if rply.Rcode != dns.RcodeSuccess {
			t.Errorf("failed to get an valid answer\n%v", rply)
		}
		answr := rply.Answer[0].(*dns.SRV)
		if answr.Priority != uint16(5) {
			t.Errorf("Expected :<%q> , received: <%q>", uint16(5), answr.Priority)
		}
		if answr.Weight != uint16(0) {
			t.Errorf("Expected :<%q> , received: <%q>", uint16(0), answr.Weight)
		}
		if answr.Port != uint16(389) {
			t.Errorf("Expected :<%q> , received: <%q>", uint16(389), answr.Port)
		}
		if answr.Target != "ldap.google.com." {
			t.Errorf("Expected :<%q> , received: <%q>", "ldap.google.com.", answr.Target)
		}
	}
	if err := dnsClnt.TCPTLS.WriteMsg(m); err != nil {
		t.Error(err)
	}
	if rply, err := dnsClnt.TCPTLS.ReadMsg(); err != nil {
		t.Error(err)
	} else {
		if rply.Rcode != dns.RcodeSuccess {
			t.Errorf("failed to get an valid answer\n%v", rply)
		}
		answr := rply.Answer[0].(*dns.SRV)
		if answr.Priority != uint16(5) {
			t.Errorf("Expected :<%q> , received: <%q>", uint16(5), answr.Priority)
		}
		if answr.Weight != uint16(0) {
			t.Errorf("Expected :<%q> , received: <%q>", uint16(0), answr.Weight)
		}
		if answr.Port != uint16(389) {
			t.Errorf("Expected :<%q> , received: <%q>", uint16(389), answr.Port)
		}
		if answr.Target != "ldap.google.com." {
			t.Errorf("Expected :<%q> , received: <%q>", "ldap.google.com.", answr.Target)
		}
	}
}

func testDNSitClntNAPTRAttributes(t *testing.T) {
	m := new(dns.Msg)
	m.SetQuestion("4.6.9.4.7.1.7.1.5.6.8.9.4.e164.arpa.", dns.TypeNAPTR)
	if err := dnsClnt.UDP.WriteMsg(m); err != nil {
		t.Error(err)
	}
	if rply, err := dnsClnt.UDP.ReadMsg(); err != nil {
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
	if err := dnsClnt.TCP.WriteMsg(m); err != nil {
		t.Error(err)
	}
	if rply, err := dnsClnt.TCP.ReadMsg(); err != nil {
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
	if err := dnsClnt.TCPTLS.WriteMsg(m); err != nil {
		t.Error(err)
	}
	if rply, err := dnsClnt.TCPTLS.ReadMsg(); err != nil {
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

func testDNSitClntASuppliers(t *testing.T) {
	m := new(dns.Msg)
	m.SetQuestion("go.dev.", dns.TypeA)
	if err := dnsClnt.UDP.WriteMsg(m); err != nil {
		t.Error(err)
	}
	rply, err := dnsClnt.UDP.ReadMsg()
	if err != nil {
		t.Error(err)
	} else if len(rply.Answer) != 2 {
		t.Fatalf("wrong number of records: %s", utils.ToIJSON(rply.Answer))
	}
	if rply.Rcode != dns.RcodeSuccess {
		t.Errorf("failed to get an valid answer\n%v", rply)
	}

	answr0 := rply.Answer[0].(*dns.A)
	if answr0.A.String() != "216.239.32.21" {
		t.Errorf("Expected :<%q> , received: <%q>", "216.239.32.21", answr0.A)
	}
	answr1 := rply.Answer[1].(*dns.A)
	if answr1.A.String() != "216.239.34.21" {
		t.Errorf("Expected :<%q> , received: <%q>", "216.239.34.21", answr1.A)
	}
	if err := dnsClnt.TCP.WriteMsg(m); err != nil {
		t.Error(err)
	}
	rply, err = dnsClnt.TCP.ReadMsg()
	if err != nil {
		t.Error(err)
	} else if len(rply.Answer) != 2 {
		t.Fatalf("wrong number of records: %s", utils.ToIJSON(rply.Answer))
	}
	if rply.Rcode != dns.RcodeSuccess {
		t.Errorf("failed to get an valid answer\n%v", rply)
	}

	answr0 = rply.Answer[0].(*dns.A)
	if answr0.A.String() != "216.239.32.21" {
		t.Errorf("Expected :<%q> , received: <%q>", "216.239.32.21", answr0.A)
	}
	answr1 = rply.Answer[1].(*dns.A)
	if answr1.A.String() != "216.239.34.21" {
		t.Errorf("Expected :<%q> , received: <%q>", "216.239.34.21", answr1.A)
	}
	if err := dnsClnt.TCPTLS.WriteMsg(m); err != nil {
		t.Error(err)
	}
	rply, err = dnsClnt.TCPTLS.ReadMsg()
	if err != nil {
		t.Error(err)
	} else if len(rply.Answer) != 2 {
		t.Fatalf("wrong number of records: %s", utils.ToIJSON(rply.Answer))
	}
	if rply.Rcode != dns.RcodeSuccess {
		t.Errorf("failed to get an valid answer\n%v", rply)
	}

	answr0 = rply.Answer[0].(*dns.A)
	if answr0.A.String() != "216.239.32.21" {
		t.Errorf("Expected :<%q> , received: <%q>", "216.239.32.21", answr0.A)
	}
	answr1 = rply.Answer[1].(*dns.A)
	if answr1.A.String() != "216.239.34.21" {
		t.Errorf("Expected :<%q> , received: <%q>", "216.239.34.21", answr1.A)
	}

}

func testDNSitClntSRVSuppliers(t *testing.T) {
	m := new(dns.Msg)
	m.SetQuestion("_xmpp-client._tcp.xmpp.org.", dns.TypeSRV)
	if err := dnsClnt.UDP.WriteMsg(m); err != nil {
		t.Error(err)
	}
	rply, err := dnsClnt.UDP.ReadMsg()
	if err != nil {
		t.Error(err)
	} else if len(rply.Answer) != 2 {
		t.Fatalf("wrong number of records: %s", utils.ToIJSON(rply.Answer))
	}
	if rply.Rcode != dns.RcodeSuccess {
		t.Errorf("failed to get an valid answer\n%v", rply)
	}
	answr := rply.Answer[0].(*dns.SRV)
	if answr.Priority != uint16(1) {
		t.Errorf("Expected :<%q> , received: <%q>", uint16(1), answr.Priority)
	}
	if answr.Weight != uint16(1) {
		t.Errorf("Expected :<%q> , received: <%q>", uint16(1), answr.Weight)
	}
	if answr.Port != uint16(9222) {
		t.Errorf("Expected :<%q> , received: <%q>", uint16(9222), answr.Port)
	}
	if answr.Target != "xmpp.xmpp.org." {
		t.Errorf("Expected :<%q> , received: <%q>", "xmpp.xmpp.org.", answr.Target)
	}
	answr2 := rply.Answer[1].(*dns.SRV)
	if answr2.Priority != uint16(1) {
		t.Errorf("Expected :<%q> , received: <%q>", uint16(1), answr2.Priority)
	}
	if answr2.Weight != uint16(1) {
		t.Errorf("Expected :<%q> , received: <%q>", uint16(1), answr2.Weight)
	}
	if answr2.Port != uint16(9222) {
		t.Errorf("Expected :<%q> , received: <%q>", uint16(9222), answr2.Port)
	}
	if answr2.Target != "xmpp.xmpp.com." {
		t.Errorf("Expected :<%q> , received: <%q>", "xmpp.xmpp.com.", answr2.Target)
	}
	if err := dnsClnt.TCP.WriteMsg(m); err != nil {
		t.Error(err)
	}
	rply, err = dnsClnt.TCP.ReadMsg()
	if err != nil {
		t.Error(err)
	} else if len(rply.Answer) != 2 {
		t.Fatalf("wrong number of records: %s", utils.ToIJSON(rply.Answer))
	}
	if rply.Rcode != dns.RcodeSuccess {
		t.Errorf("failed to get an valid answer\n%v", rply)
	}
	answr = rply.Answer[0].(*dns.SRV)
	if answr.Priority != uint16(1) {
		t.Errorf("Expected :<%q> , received: <%q>", uint16(1), answr.Priority)
	}
	if answr.Weight != uint16(1) {
		t.Errorf("Expected :<%q> , received: <%q>", uint16(1), answr.Weight)
	}
	if answr.Port != uint16(9222) {
		t.Errorf("Expected :<%q> , received: <%q>", uint16(9222), answr.Port)
	}
	if answr.Target != "xmpp.xmpp.org." {
		t.Errorf("Expected :<%q> , received: <%q>", "xmpp.xmpp.org.", answr.Target)
	}
	answr2 = rply.Answer[1].(*dns.SRV)
	if answr2.Priority != uint16(1) {
		t.Errorf("Expected :<%q> , received: <%q>", uint16(1), answr2.Priority)
	}
	if answr2.Weight != uint16(1) {
		t.Errorf("Expected :<%q> , received: <%q>", uint16(1), answr2.Weight)
	}
	if answr2.Port != uint16(9222) {
		t.Errorf("Expected :<%q> , received: <%q>", uint16(9222), answr2.Port)
	}
	if answr2.Target != "xmpp.xmpp.com." {
		t.Errorf("Expected :<%q> , received: <%q>", "xmpp.xmpp.com.", answr2.Target)
	}
	if err := dnsClnt.TCPTLS.WriteMsg(m); err != nil {
		t.Error(err)
	}
	rply, err = dnsClnt.TCPTLS.ReadMsg()
	if err != nil {
		t.Error(err)
	} else if len(rply.Answer) != 2 {
		t.Fatalf("wrong number of records: %s", utils.ToIJSON(rply.Answer))
	}
	if rply.Rcode != dns.RcodeSuccess {
		t.Errorf("failed to get an valid answer\n%v", rply)
	}
	answr = rply.Answer[0].(*dns.SRV)
	if answr.Priority != uint16(1) {
		t.Errorf("Expected :<%q> , received: <%q>", uint16(1), answr.Priority)
	}
	if answr.Weight != uint16(1) {
		t.Errorf("Expected :<%q> , received: <%q>", uint16(1), answr.Weight)
	}
	if answr.Port != uint16(9222) {
		t.Errorf("Expected :<%q> , received: <%q>", uint16(9222), answr.Port)
	}
	if answr.Target != "xmpp.xmpp.org." {
		t.Errorf("Expected :<%q> , received: <%q>", "xmpp.xmpp.org.", answr.Target)
	}
	answr2 = rply.Answer[1].(*dns.SRV)
	if answr2.Priority != uint16(1) {
		t.Errorf("Expected :<%q> , received: <%q>", uint16(1), answr2.Priority)
	}
	if answr2.Weight != uint16(1) {
		t.Errorf("Expected :<%q> , received: <%q>", uint16(1), answr2.Weight)
	}
	if answr2.Port != uint16(9222) {
		t.Errorf("Expected :<%q> , received: <%q>", uint16(9222), answr2.Port)
	}
	if answr2.Target != "xmpp.xmpp.com." {
		t.Errorf("Expected :<%q> , received: <%q>", "xmpp.xmpp.com.", answr2.Target)
	}

}

func testDNSitClntNAPTRSuppliers(t *testing.T) {
	m := new(dns.Msg)
	m.SetQuestion("5.6.9.4.7.1.7.1.5.6.8.9.4.e164.arpa.", dns.TypeNAPTR)
	if err := dnsClnt.UDP.WriteMsg(m); err != nil {
		t.Error(err)
	}
	rply, err := dnsClnt.UDP.ReadMsg()
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
	if err := dnsClnt.TCP.WriteMsg(m); err != nil {
		t.Error(err)
	}
	rply, err = dnsClnt.TCP.ReadMsg()
	if err != nil {
		t.Error(err)
	} else if len(rply.Answer) != 2 {
		t.Fatalf("wrong number of records: %s", utils.ToIJSON(rply.Answer))
	}
	if rply.Rcode != dns.RcodeSuccess {
		t.Errorf("failed to get an valid answer\n%v", rply)
	}
	answr = rply.Answer[0].(*dns.NAPTR)
	if answr.Order != 100 {
		t.Errorf("received: <%v>", answr.Order)
	}
	if answr.Regexp != "!^(.*)$!sip:1@172.16.1.11!" {
		t.Errorf("received: <%q>", answr.Regexp)
	}
	answr2 = rply.Answer[1].(*dns.NAPTR)
	if answr2.Regexp != "!^(.*)$!sip:1@172.16.1.12!" {
		t.Errorf("received: <%q>", answr2.Regexp)
	}
	if err := dnsClnt.TCPTLS.WriteMsg(m); err != nil {
		t.Error(err)
	}
	rply, err = dnsClnt.TCPTLS.ReadMsg()
	if err != nil {
		t.Error(err)
	} else if len(rply.Answer) != 2 {
		t.Fatalf("wrong number of records: %s", utils.ToIJSON(rply.Answer))
	}
	if rply.Rcode != dns.RcodeSuccess {
		t.Errorf("failed to get an valid answer\n%v", rply)
	}
	answr = rply.Answer[0].(*dns.NAPTR)
	if answr.Order != 100 {
		t.Errorf("received: <%v>", answr.Order)
	}
	if answr.Regexp != "!^(.*)$!sip:1@172.16.1.11!" {
		t.Errorf("received: <%q>", answr.Regexp)
	}
	answr2 = rply.Answer[1].(*dns.NAPTR)
	if answr2.Regexp != "!^(.*)$!sip:1@172.16.1.12!" {
		t.Errorf("received: <%q>", answr2.Regexp)
	}
}

func testDNSitClntAOpts(t *testing.T) {
	m := new(dns.Msg)
	m.SetQuestion("example.com.", dns.TypeA)
	m.SetEdns0(4096, false)
	m.IsEdns0().Option = append(m.IsEdns0().Option, &dns.EDNS0_ESU{Uri: "sip:cgrates@cgrates.org"})
	if err := dnsClnt.UDP.WriteMsg(m); err != nil {
		t.Error(err)
	}
	if rply, err := dnsClnt.UDP.ReadMsg(); err != nil {
		t.Error(err)
	} else if len(rply.Answer) != 1 {
		t.Fatalf("wrong number of records: %s", utils.ToIJSON(rply.Answer))
	} else {
		if rply.Rcode != dns.RcodeSuccess {
			t.Errorf("failed to get an valid answer\n%v", rply)
		}
		answr0 := rply.Answer[0].(*dns.A)
		if answr0.A.String() != "93.184.216.34" {
			t.Errorf("Expected :<%q> , received: <%q>", "93.184.216.34", answr0.A)
		}
		if opts := rply.IsEdns0(); opts == nil {
			t.Error("recieved nil options")
		} else if len(opts.Option) != 2 {
			t.Errorf("recieved wrong number of options: %v", len(opts.Option))
		} else if ov, can := opts.Option[0].(*dns.EDNS0_ESU); !can {
			t.Errorf("recieved wrong option type: %T", opts.Option[0])
		} else if expected := "sip:cgrates@cgrates.com"; ov.Uri != expected {
			t.Errorf("Expected :<%q> , received: <%q>", expected, ov.Uri)
		} else if ov, can := opts.Option[1].(*dns.EDNS0_ESU); !can {
			t.Errorf("recieved wrong option type: %T", opts.Option[1])
		} else if expected := "sip:cgrates@cgrates.net"; ov.Uri != expected {
			t.Errorf("Expected :<%q> , received: <%q>", expected, ov.Uri)
		}
	}

	if err := dnsClnt.TCP.WriteMsg(m); err != nil {
		t.Error(err)
	}
	if rply, err := dnsClnt.TCP.ReadMsg(); err != nil {
		t.Error(err)
	} else if len(rply.Answer) != 1 {
		t.Fatalf("wrong number of records: %s", utils.ToIJSON(rply.Answer))
	} else {
		if rply.Rcode != dns.RcodeSuccess {
			t.Errorf("failed to get an valid answer\n%v", rply)
		}
		answr0 := rply.Answer[0].(*dns.A)
		if answr0.A.String() != "93.184.216.34" {
			t.Errorf("Expected :<%q> , received: <%q>", "93.184.216.34", answr0.A)
		}
		if opts := rply.IsEdns0(); opts == nil {
			t.Error("recieved nil options")
		} else if len(opts.Option) != 2 {
			t.Errorf("recieved wrong number of options: %v", len(opts.Option))
		} else if ov, can := opts.Option[0].(*dns.EDNS0_ESU); !can {
			t.Errorf("recieved wrong option type: %T", opts.Option[0])
		} else if expected := "sip:cgrates@cgrates.com"; ov.Uri != expected {
			t.Errorf("Expected :<%q> , received: <%q>", expected, ov.Uri)
		} else if ov, can := opts.Option[1].(*dns.EDNS0_ESU); !can {
			t.Errorf("recieved wrong option type: %T", opts.Option[1])
		} else if expected := "sip:cgrates@cgrates.net"; ov.Uri != expected {
			t.Errorf("Expected :<%q> , received: <%q>", expected, ov.Uri)
		}
	}
	if err := dnsClnt.TCPTLS.WriteMsg(m); err != nil {
		t.Error(err)
	}
	if rply, err := dnsClnt.TCPTLS.ReadMsg(); err != nil {
		t.Error(err)
	} else if len(rply.Answer) != 1 {
		t.Fatalf("wrong number of records: %s", utils.ToIJSON(rply.Answer))
	} else {
		if rply.Rcode != dns.RcodeSuccess {
			t.Errorf("failed to get an valid answer\n%v", rply)
		}
		answr0 := rply.Answer[0].(*dns.A)
		if answr0.A.String() != "93.184.216.34" {
			t.Errorf("Expected :<%q> , received: <%q>", "93.184.216.34", answr0.A)
		}
		if opts := rply.IsEdns0(); opts == nil {
			t.Error("recieved nil options")
		} else if len(opts.Option) != 2 {
			t.Errorf("recieved wrong number of options: %v", len(opts.Option))
		} else if ov, can := opts.Option[0].(*dns.EDNS0_ESU); !can {
			t.Errorf("recieved wrong option type: %T", opts.Option[0])
		} else if expected := "sip:cgrates@cgrates.com"; ov.Uri != expected {
			t.Errorf("Expected :<%q> , received: <%q>", expected, ov.Uri)
		} else if ov, can := opts.Option[1].(*dns.EDNS0_ESU); !can {
			t.Errorf("recieved wrong option type: %T", opts.Option[1])
		} else if expected := "sip:cgrates@cgrates.net"; ov.Uri != expected {
			t.Errorf("Expected :<%q> , received: <%q>", expected, ov.Uri)
		}
	}
}
func testDNSitClntSRVOpts(t *testing.T) {
	m := new(dns.Msg)
	m.SetQuestion("_matrix._tcp.matrix.org.", dns.TypeSRV)
	m.SetEdns0(4096, false)
	m.IsEdns0().Option = append(m.IsEdns0().Option,
		&dns.EDNS0_ESU{Uri: "sip:cgrates@cgrates.org"})
	if err := dnsClnt.UDP.WriteMsg(m); err != nil {
		t.Error(err)
	}
	rply, err := dnsClnt.UDP.ReadMsg()
	if err != nil {
		t.Error(err)
	} else if len(rply.Answer) != 1 {
		t.Fatalf("wrong number of records: %s", utils.ToIJSON(rply.Answer))
	}
	if rply.Rcode != dns.RcodeSuccess {
		t.Errorf("failed to get an valid answer\n%v", rply)
	}
	answr := rply.Answer[0].(*dns.SRV)
	if answr.Priority != uint16(10) {
		t.Errorf("Expected :<%q> , received: <%q>", uint16(10), answr.Priority)
	}
	if answr.Weight != uint16(5) {
		t.Errorf("Expected :<%q> , received: <%q>", uint16(5), answr.Weight)
	}
	if answr.Port != uint16(8443) {
		t.Errorf("Expected :<%q> , received: <%q>", uint16(8443), answr.Port)
	}
	if answr.Target != "matrix-federation.matrix.org.cdn.cloudflare.net." {
		t.Errorf("Expected :<%q> , received: <%q>",
			"matrix-federation.matrix.org.cdn.cloudflare.net.", answr.Target)
	}
	if opts := rply.IsEdns0(); opts == nil {
		t.Error("recieved nil options")
	} else if len(opts.Option) != 2 {
		t.Errorf("recieved wrong number of options: %v", len(opts.Option))
	} else if ov, can := opts.Option[0].(*dns.EDNS0_ESU); !can {
		t.Errorf("recieved wrong option type: %T", opts.Option[0])
	} else if expected := "sip:cgrates@cgrates.com"; ov.Uri != expected {
		t.Errorf("Expected :<%q> , received: <%q>", expected, ov.Uri)
	} else if ov, can := opts.Option[1].(*dns.EDNS0_ESU); !can {
		t.Errorf("recieved wrong option type: %T", opts.Option[1])
	} else if expected := "sip:cgrates@cgrates.net"; ov.Uri != expected {
		t.Errorf("Expected :<%q> , received: <%q>", expected, ov.Uri)
	}
	if err := dnsClnt.TCP.WriteMsg(m); err != nil {
		t.Error(err)
	}
	rply, err = dnsClnt.TCP.ReadMsg()
	if err != nil {
		t.Error(err)
	} else if len(rply.Answer) != 1 {
		t.Fatalf("wrong number of records: %s", utils.ToIJSON(rply.Answer))
	}
	if rply.Rcode != dns.RcodeSuccess {
		t.Errorf("failed to get an valid answer\n%v", rply)
	}
	answr = rply.Answer[0].(*dns.SRV)
	if answr.Priority != uint16(10) {
		t.Errorf("Expected :<%q> , received: <%q>", uint16(10), answr.Priority)
	}
	if answr.Weight != uint16(5) {
		t.Errorf("Expected :<%q> , received: <%q>", uint16(5), answr.Weight)
	}
	if answr.Port != uint16(8443) {
		t.Errorf("Expected :<%q> , received: <%q>", uint16(8443), answr.Port)
	}
	if answr.Target != "matrix-federation.matrix.org.cdn.cloudflare.net." {
		t.Errorf("Expected :<%q> , received: <%q>",
			"matrix-federation.matrix.org.cdn.cloudflare.net.", answr.Target)
	}
	if opts := rply.IsEdns0(); opts == nil {
		t.Error("recieved nil options")
	} else if len(opts.Option) != 2 {
		t.Errorf("recieved wrong number of options: %v", len(opts.Option))
	} else if ov, can := opts.Option[0].(*dns.EDNS0_ESU); !can {
		t.Errorf("recieved wrong option type: %T", opts.Option[0])
	} else if expected := "sip:cgrates@cgrates.com"; ov.Uri != expected {
		t.Errorf("Expected :<%q> , received: <%q>", expected, ov.Uri)
	} else if ov, can := opts.Option[1].(*dns.EDNS0_ESU); !can {
		t.Errorf("recieved wrong option type: %T", opts.Option[1])
	} else if expected := "sip:cgrates@cgrates.net"; ov.Uri != expected {
		t.Errorf("Expected :<%q> , received: <%q>", expected, ov.Uri)
	}
	if err := dnsClnt.TCPTLS.WriteMsg(m); err != nil {
		t.Error(err)
	}
	rply, err = dnsClnt.TCPTLS.ReadMsg()
	if err != nil {
		t.Error(err)
	} else if len(rply.Answer) != 1 {
		t.Fatalf("wrong number of records: %s", utils.ToIJSON(rply.Answer))
	}
	if rply.Rcode != dns.RcodeSuccess {
		t.Errorf("failed to get an valid answer\n%v", rply)
	}
	answr = rply.Answer[0].(*dns.SRV)
	if answr.Priority != uint16(10) {
		t.Errorf("Expected :<%q> , received: <%q>", uint16(10), answr.Priority)
	}
	if answr.Weight != uint16(5) {
		t.Errorf("Expected :<%q> , received: <%q>", uint16(5), answr.Weight)
	}
	if answr.Port != uint16(8443) {
		t.Errorf("Expected :<%q> , received: <%q>", uint16(8443), answr.Port)
	}
	if answr.Target != "matrix-federation.matrix.org.cdn.cloudflare.net." {
		t.Errorf("Expected :<%q> , received: <%q>",
			"matrix-federation.matrix.org.cdn.cloudflare.net.", answr.Target)
	}
	if opts := rply.IsEdns0(); opts == nil {
		t.Error("recieved nil options")
	} else if len(opts.Option) != 2 {
		t.Errorf("recieved wrong number of options: %v", len(opts.Option))
	} else if ov, can := opts.Option[0].(*dns.EDNS0_ESU); !can {
		t.Errorf("recieved wrong option type: %T", opts.Option[0])
	} else if expected := "sip:cgrates@cgrates.com"; ov.Uri != expected {
		t.Errorf("Expected :<%q> , received: <%q>", expected, ov.Uri)
	} else if ov, can := opts.Option[1].(*dns.EDNS0_ESU); !can {
		t.Errorf("recieved wrong option type: %T", opts.Option[1])
	} else if expected := "sip:cgrates@cgrates.net"; ov.Uri != expected {
		t.Errorf("Expected :<%q> , received: <%q>", expected, ov.Uri)
	}
}

func testDNSitClntNAPTROpts(t *testing.T) {
	m := new(dns.Msg)
	m.SetQuestion("5.6.9.4.7.1.7.1.5.6.8.9.5.e164.arpa.", dns.TypeNAPTR)
	m.SetEdns0(4096, false)
	m.IsEdns0().Option = append(m.IsEdns0().Option, &dns.EDNS0_ESU{Uri: "sip:cgrates@cgrates.org"})
	if err := dnsClnt.UDP.WriteMsg(m); err != nil {
		t.Error(err)
	}
	rply, err := dnsClnt.UDP.ReadMsg()
	if err != nil {
		t.Error(err)
	} else if len(rply.Answer) != 1 {
		t.Fatalf("wrong number of records: %s", utils.ToIJSON(rply.Answer))
	}
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
	if opts := rply.IsEdns0(); opts == nil {
		t.Error("recieved nil options")
	} else if len(opts.Option) != 2 {
		t.Errorf("recieved wrong number of options: %v", len(opts.Option))
	} else if ov, can := opts.Option[0].(*dns.EDNS0_ESU); !can {
		t.Errorf("recieved wrong option type: %T", opts.Option[0])
	} else if expected := "sip:cgrates@cgrates.com"; ov.Uri != expected {
		t.Errorf("Expected :<%q> , received: <%q>", expected, ov.Uri)
	} else if ov, can := opts.Option[1].(*dns.EDNS0_ESU); !can {
		t.Errorf("recieved wrong option type: %T", opts.Option[1])
	} else if expected := "sip:cgrates@cgrates.net"; ov.Uri != expected {
		t.Errorf("Expected :<%q> , received: <%q>", expected, ov.Uri)
	}

	if err := dnsClnt.TCP.WriteMsg(m); err != nil {
		t.Error(err)
	}
	rply, err = dnsClnt.TCP.ReadMsg()
	if err != nil {
		t.Error(err)
	} else if len(rply.Answer) != 1 {
		t.Fatalf("wrong number of records: %s", utils.ToIJSON(rply.Answer))
	}
	if rply.Rcode != dns.RcodeSuccess {
		t.Errorf("failed to get an valid answer\n%v", rply)
	}
	answr = rply.Answer[0].(*dns.NAPTR)
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
	if opts := rply.IsEdns0(); opts == nil {
		t.Error("recieved nil options")
	} else if len(opts.Option) != 2 {
		t.Errorf("recieved wrong number of options: %v", len(opts.Option))
	} else if ov, can := opts.Option[0].(*dns.EDNS0_ESU); !can {
		t.Errorf("recieved wrong option type: %T", opts.Option[0])
	} else if expected := "sip:cgrates@cgrates.com"; ov.Uri != expected {
		t.Errorf("Expected :<%q> , received: <%q>", expected, ov.Uri)
	} else if ov, can := opts.Option[1].(*dns.EDNS0_ESU); !can {
		t.Errorf("recieved wrong option type: %T", opts.Option[1])
	} else if expected := "sip:cgrates@cgrates.net"; ov.Uri != expected {
		t.Errorf("Expected :<%q> , received: <%q>", expected, ov.Uri)
	}

	if err := dnsClnt.TCPTLS.WriteMsg(m); err != nil {
		t.Error(err)
	}
	rply, err = dnsClnt.TCPTLS.ReadMsg()
	if err != nil {
		t.Error(err)
	} else if len(rply.Answer) != 1 {
		t.Fatalf("wrong number of records: %s", utils.ToIJSON(rply.Answer))
	}
	if rply.Rcode != dns.RcodeSuccess {
		t.Errorf("failed to get an valid answer\n%v", rply)
	}
	answr = rply.Answer[0].(*dns.NAPTR)
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
	if opts := rply.IsEdns0(); opts == nil {
		t.Error("recieved nil options")
	} else if len(opts.Option) != 2 {
		t.Errorf("recieved wrong number of options: %v", len(opts.Option))
	} else if ov, can := opts.Option[0].(*dns.EDNS0_ESU); !can {
		t.Errorf("recieved wrong option type: %T", opts.Option[0])
	} else if expected := "sip:cgrates@cgrates.com"; ov.Uri != expected {
		t.Errorf("Expected :<%q> , received: <%q>", expected, ov.Uri)
	} else if ov, can := opts.Option[1].(*dns.EDNS0_ESU); !can {
		t.Errorf("recieved wrong option type: %T", opts.Option[1])
	} else if expected := "sip:cgrates@cgrates.net"; ov.Uri != expected {
		t.Errorf("Expected :<%q> , received: <%q>", expected, ov.Uri)
	}
}

func testDNSitClntAOptsWithAttributes(t *testing.T) {
	m := new(dns.Msg)
	m.SetQuestion("opendns.com.", dns.TypeA)
	m.SetEdns0(4096, false)
	m.IsEdns0().Option = append(m.IsEdns0().Option, &dns.EDNS0_ESU{Uri: "sip:cgrates@cgrates.org"})
	if err := dnsClnt.UDP.WriteMsg(m); err != nil {
		t.Error(err)
	}
	if rply, err := dnsClnt.UDP.ReadMsg(); err != nil {
		t.Error(err)
	} else if len(rply.Answer) != 1 {
		t.Fatalf("wrong number of records: %s", utils.ToIJSON(rply.Answer))
	} else {
		if rply.Rcode != dns.RcodeSuccess {
			t.Errorf("failed to get an valid answer\n%v", rply)
		}
		answr0 := rply.Answer[0].(*dns.A)
		if answr0.A.String() != "146.112.62.105" {
			t.Errorf("Expected :<%q> , received: <%q>", "146.112.62.105", answr0.A)
		}
		if opts := rply.IsEdns0(); opts == nil {
			t.Error("recieved nil options")
		} else if len(opts.Option) != 1 {
			t.Errorf("recieved wrong number of options: %v", len(opts.Option))
		} else if ov, can := opts.Option[0].(*dns.EDNS0_ESU); !can {
			t.Errorf("recieved wrong option type: %T", opts.Option[0])
		} else if expected := "sip:cgrates@opendns.com."; ov.Uri != expected {
			t.Errorf("Expected :<%q> , received: <%q>", expected, ov.Uri)
		}
	}

	if err := dnsClnt.TCP.WriteMsg(m); err != nil {
		t.Error(err)
	}
	if rply, err := dnsClnt.TCP.ReadMsg(); err != nil {
		t.Error(err)
	} else if len(rply.Answer) != 1 {
		t.Fatalf("wrong number of records: %s", utils.ToIJSON(rply.Answer))
	} else {
		if rply.Rcode != dns.RcodeSuccess {
			t.Errorf("failed to get an valid answer\n%v", rply)
		}
		answr0 := rply.Answer[0].(*dns.A)
		if answr0.A.String() != "146.112.62.105" {
			t.Errorf("Expected :<%q> , received: <%q>", "146.112.62.105", answr0.A)
		}
		if opts := rply.IsEdns0(); opts == nil {
			t.Error("recieved nil options")
		} else if len(opts.Option) != 1 {
			t.Errorf("recieved wrong number of options: %v", len(opts.Option))
		} else if ov, can := opts.Option[0].(*dns.EDNS0_ESU); !can {
			t.Errorf("recieved wrong option type: %T", opts.Option[0])
		} else if expected := "sip:cgrates@opendns.com."; ov.Uri != expected {
			t.Errorf("Expected :<%q> , received: <%q>", expected, ov.Uri)
		}
	}
	if err := dnsClnt.TCPTLS.WriteMsg(m); err != nil {
		t.Error(err)
	}
	if rply, err := dnsClnt.TCPTLS.ReadMsg(); err != nil {
		t.Error(err)
	} else if len(rply.Answer) != 1 {
		t.Fatalf("wrong number of records: %s", utils.ToIJSON(rply.Answer))
	} else {
		if rply.Rcode != dns.RcodeSuccess {
			t.Errorf("failed to get an valid answer\n%v", rply)
		}
		answr0 := rply.Answer[0].(*dns.A)
		if answr0.A.String() != "146.112.62.105" {
			t.Errorf("Expected :<%q> , received: <%q>", "146.112.62.105", answr0.A)
		}
		if opts := rply.IsEdns0(); opts == nil {
			t.Error("recieved nil options")
		} else if len(opts.Option) != 1 {
			t.Errorf("recieved wrong number of options: %v", len(opts.Option))
		} else if ov, can := opts.Option[0].(*dns.EDNS0_ESU); !can {
			t.Errorf("recieved wrong option type: %T", opts.Option[0])
		} else if expected := "sip:cgrates@opendns.com."; ov.Uri != expected {
			t.Errorf("Expected :<%q> , received: <%q>", expected, ov.Uri)
		}
	}
}

func testDNSitClntSRVOptsWithAttributes(t *testing.T) {
	m := new(dns.Msg)
	m.SetQuestion("_sip._udp.opensips.org.", dns.TypeSRV)
	m.SetEdns0(4096, false)
	m.IsEdns0().Option = append(m.IsEdns0().Option,
		&dns.EDNS0_ESU{Uri: "sip:cgrates@cgrates.org"})
	if err := dnsClnt.UDP.WriteMsg(m); err != nil {
		t.Error(err)
	}
	rply, err := dnsClnt.UDP.ReadMsg()
	if err != nil {
		t.Error(err)
	} else if len(rply.Answer) != 1 {
		t.Fatalf("wrong number of records: %s", utils.ToIJSON(rply.Answer))
	}
	if rply.Rcode != dns.RcodeSuccess {
		t.Errorf("failed to get an valid answer\n%v", rply)
	}
	answr := rply.Answer[0].(*dns.SRV)
	if answr.Priority != uint16(0) {
		t.Errorf("Expected :<%q> , received: <%q>", uint16(0), answr.Priority)
	}
	if answr.Weight != uint16(50) {
		t.Errorf("Expected :<%q> , received: <%q>", uint16(50), answr.Weight)
	}
	if answr.Port != uint16(5060) {
		t.Errorf("Expected :<%q> , received: <%q>", uint16(5060), answr.Port)
	}
	if answr.Target != "opensips.org." {
		t.Errorf("Expected :<%q> , received: <%q>",
			"opensips.org.", answr.Target)
	}
	if opts := rply.IsEdns0(); opts == nil {
		t.Error("recieved nil options")
	} else if len(opts.Option) != 1 {
		t.Errorf("recieved wrong number of options: %v", len(opts.Option))
	} else if ov, can := opts.Option[0].(*dns.EDNS0_ESU); !can {
		t.Errorf("recieved wrong option type: %T", opts.Option[0])
	} else if expected := "sip:cgrates@_sip._udp.opensips.org."; ov.Uri != expected {
		t.Errorf("Expected :<%q> , received: <%q>", expected, ov.Uri)
	}
	if err := dnsClnt.TCP.WriteMsg(m); err != nil {
		t.Error(err)
	}
	rply, err = dnsClnt.TCP.ReadMsg()
	if err != nil {
		t.Error(err)
	} else if len(rply.Answer) != 1 {
		t.Fatalf("wrong number of records: %s", utils.ToIJSON(rply.Answer))
	}
	if rply.Rcode != dns.RcodeSuccess {
		t.Errorf("failed to get an valid answer\n%v", rply)
	}
	answr = rply.Answer[0].(*dns.SRV)
	if answr.Priority != uint16(0) {
		t.Errorf("Expected :<%q> , received: <%q>", uint16(0), answr.Priority)
	}
	if answr.Weight != uint16(50) {
		t.Errorf("Expected :<%q> , received: <%q>", uint16(50), answr.Weight)
	}
	if answr.Port != uint16(5060) {
		t.Errorf("Expected :<%q> , received: <%q>", uint16(5060), answr.Port)
	}
	if answr.Target != "opensips.org." {
		t.Errorf("Expected :<%q> , received: <%q>",
			"opensips.org.", answr.Target)
	}
	if opts := rply.IsEdns0(); opts == nil {
		t.Error("recieved nil options")
	} else if len(opts.Option) != 1 {
		t.Errorf("recieved wrong number of options: %v", len(opts.Option))
	} else if ov, can := opts.Option[0].(*dns.EDNS0_ESU); !can {
		t.Errorf("recieved wrong option type: %T", opts.Option[0])
	} else if expected := "sip:cgrates@_sip._udp.opensips.org."; ov.Uri != expected {
		t.Errorf("Expected :<%q> , received: <%q>", expected, ov.Uri)
	}
	if err := dnsClnt.TCPTLS.WriteMsg(m); err != nil {
		t.Error(err)
	}
	rply, err = dnsClnt.TCPTLS.ReadMsg()
	if err != nil {
		t.Error(err)
	} else if len(rply.Answer) != 1 {
		t.Fatalf("wrong number of records: %s", utils.ToIJSON(rply.Answer))
	}
	if rply.Rcode != dns.RcodeSuccess {
		t.Errorf("failed to get an valid answer\n%v", rply)
	}
	answr = rply.Answer[0].(*dns.SRV)
	if answr.Priority != uint16(0) {
		t.Errorf("Expected :<%q> , received: <%q>", uint16(0), answr.Priority)
	}
	if answr.Weight != uint16(50) {
		t.Errorf("Expected :<%q> , received: <%q>", uint16(50), answr.Weight)
	}
	if answr.Port != uint16(5060) {
		t.Errorf("Expected :<%q> , received: <%q>", uint16(5060), answr.Port)
	}
	if answr.Target != "opensips.org." {
		t.Errorf("Expected :<%q> , received: <%q>",
			"opensips.org.", answr.Target)
	}
	if opts := rply.IsEdns0(); opts == nil {
		t.Error("recieved nil options")
	} else if len(opts.Option) != 1 {
		t.Errorf("recieved wrong number of options: %v", len(opts.Option))
	} else if ov, can := opts.Option[0].(*dns.EDNS0_ESU); !can {
		t.Errorf("recieved wrong option type: %T", opts.Option[0])
	} else if expected := "sip:cgrates@_sip._udp.opensips.org."; ov.Uri != expected {
		t.Errorf("Expected :<%q> , received: <%q>", expected, ov.Uri)
	}

}

func testDNSitClntNAPTROptsWithAttributes(t *testing.T) {
	m := new(dns.Msg)
	m.SetQuestion("7.6.9.4.7.1.7.1.5.6.8.9.5.e164.arpa.", dns.TypeNAPTR)
	m.SetEdns0(4096, false)
	m.IsEdns0().Option = append(m.IsEdns0().Option, &dns.EDNS0_ESU{Uri: "sip:cgrates@cgrates.org"})
	if err := dnsClnt.UDP.WriteMsg(m); err != nil {
		t.Error(err)
	}
	rply, err := dnsClnt.UDP.ReadMsg()
	if err != nil {
		t.Error(err)
	} else if len(rply.Answer) != 1 {
		t.Fatalf("wrong number of records: %s", utils.ToIJSON(rply.Answer))
	}
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
	if opts := rply.IsEdns0(); opts == nil {
		t.Error("recieved nil options")
	} else if len(opts.Option) != 1 {
		t.Errorf("recieved wrong number of options: %v", len(opts.Option))
	} else if ov, can := opts.Option[0].(*dns.EDNS0_ESU); !can {
		t.Errorf("recieved wrong option type: %T", opts.Option[0])
	} else if expected := "sip:cgrates@e164.arpa"; ov.Uri != expected {
		t.Errorf("Expected :<%q> , received: <%q>", expected, ov.Uri)
	}

	if err := dnsClnt.TCP.WriteMsg(m); err != nil {
		t.Error(err)
	}
	rply, err = dnsClnt.TCP.ReadMsg()
	if err != nil {
		t.Error(err)
	} else if len(rply.Answer) != 1 {
		t.Fatalf("wrong number of records: %s", utils.ToIJSON(rply.Answer))
	}
	if rply.Rcode != dns.RcodeSuccess {
		t.Errorf("failed to get an valid answer\n%v", rply)
	}
	answr = rply.Answer[0].(*dns.NAPTR)
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
	if opts := rply.IsEdns0(); opts == nil {
		t.Error("recieved nil options")
	} else if len(opts.Option) != 1 {
		t.Errorf("recieved wrong number of options: %v", len(opts.Option))
	} else if ov, can := opts.Option[0].(*dns.EDNS0_ESU); !can {
		t.Errorf("recieved wrong option type: %T", opts.Option[0])
	} else if expected := "sip:cgrates@e164.arpa"; ov.Uri != expected {
		t.Errorf("Expected :<%q> , received: <%q>", expected, ov.Uri)
	}
	if err := dnsClnt.TCPTLS.WriteMsg(m); err != nil {
		t.Error(err)
	}
	rply, err = dnsClnt.TCPTLS.ReadMsg()
	if err != nil {
		t.Error(err)
	} else if len(rply.Answer) != 1 {
		t.Fatalf("wrong number of records: %s", utils.ToIJSON(rply.Answer))
	}
	if rply.Rcode != dns.RcodeSuccess {
		t.Errorf("failed to get an valid answer\n%v", rply)
	}
	answr = rply.Answer[0].(*dns.NAPTR)
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
	if opts := rply.IsEdns0(); opts == nil {
		t.Error("recieved nil options")
	} else if len(opts.Option) != 1 {
		t.Errorf("recieved wrong number of options: %v", len(opts.Option))
	} else if ov, can := opts.Option[0].(*dns.EDNS0_ESU); !can {
		t.Errorf("recieved wrong option type: %T", opts.Option[0])
	} else if expected := "sip:cgrates@e164.arpa"; ov.Uri != expected {
		t.Errorf("Expected :<%q> , received: <%q>", expected, ov.Uri)
	}

}

func testDNSitStopEngine(t *testing.T) {
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}
