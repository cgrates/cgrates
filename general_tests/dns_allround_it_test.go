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
	"crypto/tls"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/miekg/dns"
)

func TestDNSAllRound1(t *testing.T) {
	var cfgDir string
	switch *utils.DBType {
	case utils.MetaInternal:
		cfgDir = "dnsagent_internal"
	case utils.MetaMySQL:
		cfgDir = "dnsagent_mysql"
	case utils.MetaMongo:
		cfgDir = "dnsagent_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	testEnv := engine.TestEngine{
		ConfigPath: path.Join(*utils.DataDir, "conf", "samples", cfgDir),
		TpPath:     path.Join(*utils.DataDir, "tariffplans", "dnsagent"),
	}
	client, dnsCfg := testEnv.Run(t)
	type dnsConns struct {
		UDP    *dns.Conn
		TCP    *dns.Conn
		TCPTLS *dns.Conn
	}
	var dnsClnt *dnsConns
	t.Run("CheckListenersPorts", func(t *testing.T) {

		openUDP := checkDNSPortOpen(2053, "udp")
		if !openUDP {
			t.Errorf("Port %d is not open with UDP listener", 2053)
		}

		openTCP := checkDNSPortOpen(2053, "tcp")
		if !openTCP {
			t.Errorf("Port %d is not open with TCP listener", 2053)
		}

		openTLS := checkDNSPortOpen(2054, "tcp-tls")
		if !openTLS {
			t.Errorf("Port %d is not open with TCP-TLS listener", 2054)
		}
	})

	t.Run("ReloadConfigOK", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.ConfigSv1ReloadConfig, &config.ReloadArgs{
			Tenant:  "cgrates.org",
			Path:    path.Join(*utils.DataDir, "conf", "samples", cfgDir),
			Section: config.DNSAgentJson,
			DryRun:  true,
		}, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("Expected OK received: %s", reply)
		}
	})
	t.Run("Request", func(t *testing.T) {
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
	})
	t.Run("StopDNSService", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.ServiceManagerV1StopService, &servmanager.ArgStartService{
			ServiceID: utils.MetaDNS,
		}, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("Expected OK received: %s", reply)
		}
		time.Sleep(100 * time.Millisecond)
	})
	t.Run("CheckListenersPortsOff", func(t *testing.T) {
		openUDP := checkDNSPortOpen(2053, "udp")
		if openUDP {
			t.Errorf("Port %d is open with UDP listener when it should'nt be", 2053)
		}

		openTCP := checkDNSPortOpen(2053, "tcp")
		if openTCP {
			t.Errorf("Port %d is open with TCP listener when it should'nt be", 2053)
		}

		openTLS := checkDNSPortOpen(2054, "tcp-tls")
		if openTLS {
			t.Errorf("Port %d is open with TCP-TLS listener when it should'nt be", 2054)
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
"general": {
	"log_level": 7		
},
"data_db": {
	"db_type": "*internal"
},
"stor_db": {
	"db_type": "*internal"
},
"schedulers": {
	"enabled": true,
	"cdrs_conns": ["*localhost"]
},
"sessions": {
	"enabled": true,
	"attributes_conns": ["*localhost"],
	"rals_conns": ["*localhost"],
	"cdrs_conns": ["*localhost"],
	"chargers_conns": ["*localhost"],
	"routes_conns": ["*localhost"]
},
"rals": {
	"enabled": true
},
"cdrs": {
	"enabled": true,
	"rals_conns": ["*localhost"]
},"chargers": {
	"enabled": true
},
"attributes": {
	"enabled": true
},
"routes": {
	"enabled": true
},
"dns_agent": {
	"enabled": true,
	"listeners":[
		{
			"address":":2055",
			"network":"udp"
		},
		{
			"address":"badAddress:2055",
			"network":"tcp"
		},
		{
			"address":":2056",
			"network":"tcp-tls"
		}
	],
	"sessions_conns": ["*localhost"]
},
"apiers": {
	"enabled": true,
	"scheduler_conns": ["*localhost"]
}
}`
		err = writeToFile(filePath, jsonConfig)
		if err != nil {
			t.Fatal("Failed to write JSON content to file:", err)
		}
		time.Sleep(500 * time.Millisecond)
	})
	t.Run("StartDNSService", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.ServiceManagerV1StartService, &servmanager.ArgStartService{
			ServiceID: utils.MetaDNS,
		}, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("Expected OK received: %s", reply)
		}
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("CheckListenersPorts", func(t *testing.T) {

		openUDP := checkDNSPortOpen(2053, "udp")
		if !openUDP {
			t.Errorf("Port %d is not open with UDP listener", 2053)
		}

		openTCP := checkDNSPortOpen(2053, "tcp")
		if !openTCP {
			t.Errorf("Port %d is not open with TCP listener", 2053)
		}

		openTLS := checkDNSPortOpen(2054, "tcp-tls")
		if !openTLS {
			t.Errorf("Port %d is not open with TCP-TLS listener", 2054)
		}
	})

	t.Run("ReloadConfigOK", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.ConfigSv1ReloadConfig, &config.ReloadArgs{
			Tenant:  "cgrates.org",
			Path:    "/tmp/test_folder",
			Section: config.DNSAgentJson,
		}, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("Expected OK received: %s", reply)
		}
		deleteTmpFolder("/tmp/test_folder", t)
		time.Sleep(500 * time.Millisecond)
	})

	t.Run("CheckListenersPortsOff", func(t *testing.T) {

		for _, port := range []int{2055, 2053} {
			openUDP := checkDNSPortOpen(port, "udp")
			if openUDP {
				t.Errorf("Port %d is open with UDP listener when it should'nt be", port)
			}
		}
		for _, port := range []int{2055, 2053} {
			openTCP := checkDNSPortOpen(port, "tcp")
			if openTCP {
				t.Errorf("Port %d is open with TCP listener when it should'nt be", port)
			}
		}
		for _, port := range []int{2056, 2054} {
			openTLS := checkDNSPortOpen(port, "tcp-tls")
			if openTLS {
				t.Errorf("Port %d is open with TCP-TLS listener when it should'nt be", port)
			}
		}
	})

}

func TestDNSAllRound2(t *testing.T) {

	testEnv := engine.TestEngine{
		ConfigPath: path.Join(*utils.DataDir, "conf", "samples", "dnsagent_reload"),
		TpPath:     path.Join(*utils.DataDir, "tariffplans", "dnsagent"),
	}
	client, _ := testEnv.Run(t)
	t.Run("CheckListenersPorts", func(t *testing.T) {
		openUDP := checkDNSPortOpen(2053, "udp")
		if !openUDP {
			t.Errorf("Port %d is not open with UDP listener", 2053)
		}

		openTCP := checkDNSPortOpen(2054, "tcp")
		if !openTCP {
			t.Errorf("Port %d is not open with TCP listener", 2054)
		}
	})

	t.Run("CreateTmpBadConfigTLS", func(t *testing.T) {
		tmpFolderPath := filepath.Join("/tmp", "test_folder")
		err := os.Mkdir(tmpFolderPath, os.ModePerm)
		if err != nil {
			t.Fatal("Failed to create temporary folder:", err)
		}

		filePath := filepath.Join(tmpFolderPath, "cgrates.json")
		jsonConfig := `{
"general": {
	"log_level": 7		
},
"data_db": {
	"db_type": "*internal"
},
"stor_db": {
	"db_type": "*internal"
},
"schedulers": {
	"enabled": true,
	"cdrs_conns": ["*localhost"]
},
"sessions": {
	"enabled": true,
	"attributes_conns": ["*localhost"],
	"rals_conns": ["*localhost"],
	"cdrs_conns": ["*localhost"],
	"chargers_conns": ["*localhost"],
	"routes_conns": ["*localhost"]
},
"rals": {
	"enabled": true
},
"cdrs": {
	"enabled": true,
	"rals_conns": ["*localhost"]
},"chargers": {
	"enabled": true
},
"attributes": {
	"enabled": true
},
"routes": {
	"enabled": true
},
"dns_agent": {
	"enabled": true,
	"listeners":[
		{
			"address":":2055",
			"network":"udp"
		},
		{
			"address":":2055",
			"network":"tcp"
		},
		{
			"address":":2056",
			"network":"tcp-tls"
		}
	],
	"sessions_conns": ["*localhost"]
},
"apiers": {
	"enabled": true,
	"scheduler_conns": ["*localhost"]
}
}`
		err = writeToFile(filePath, jsonConfig)
		if err != nil {
			t.Fatal("Failed to write JSON content to file:", err)
		}
		time.Sleep(500 * time.Millisecond)
	})

	t.Run("ReloadConfigOK", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.ConfigSv1ReloadConfig, &config.ReloadArgs{
			Tenant:  "cgrates.org",
			Path:    "/tmp/test_folder",
			Section: config.DNSAgentJson,
		}, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("Expected OK received: %s", reply)
		}
		deleteTmpFolder("/tmp/test_folder", t)
		time.Sleep(500 * time.Millisecond)
	})

	t.Run("CheckListenersPortsOff", func(t *testing.T) {

		for _, port := range []int{2055, 2053} {
			openUDP := checkDNSPortOpen(port, "udp")
			if openUDP {
				t.Errorf("Port %d is open with UDP listener when it should'nt be", port)
			}
		}
		for _, port := range []int{2055, 2054} {
			openTCP := checkDNSPortOpen(port, "tcp")
			if openTCP {
				t.Errorf("Port %d is open with TCP listener when it should'nt be", port)
			}
		}
		openTLS := checkDNSPortOpen(2056, "tcp-tls")
		if openTLS {
			t.Errorf("Port %d is open with TCP-TLS listener when it should'nt be", 2056)
		}

	})

}

func checkDNSPortOpen(port int, protocol string) bool {
	client := dns.Client{Net: protocol, TLSConfig: &tls.Config{InsecureSkipVerify: true}}
	msg := new(dns.Msg)
	msg.SetQuestion("example.com.", dns.TypeA)

	_, _, err := client.Exchange(msg, fmt.Sprintf("localhost:%d", port))
	return err == nil
}

func writeToFile(filePath, content string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return err
	}
	return nil
}

func deleteTmpFolder(tmpFolderPath string, t *testing.T) error {
	err := os.RemoveAll(tmpFolderPath)
	if err != nil {
		t.Errorf("Failed to delete temporary folder: %v", err)
	}
	return err
}
