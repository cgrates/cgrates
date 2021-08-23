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
	"net"
	"net/rpc"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/sipingo"
)

var (
	saonfigDIR string
	saCfgPath  string
	saCfg      *config.CGRConfig
	saRPC      *rpc.Client
	saConn     net.Conn

	sTestsSIP = []func(t *testing.T){
		testSAitInitCfg,
		testSAitResetDataDb,
		testSAitResetStorDb,
		testSAitStartEngine,
		testSAitApierRpcConn,
		testSAitTPFromFolder,

		testSAitSIPRegister,
		testSAitSIPInvite,

		testSAitStopCgrEngine,
	}
)

// Test start here
func TestSAit(t *testing.T) {
	// engine.KillEngine(0)
	switch *dbType {
	case utils.MetaInternal:
		saonfigDIR = "sipagent_internal"
	case utils.MetaMySQL:
		saonfigDIR = "sipagent_mysql"
	case utils.MetaMongo:
		saonfigDIR = "sipagent_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsSIP {
		t.Run(saonfigDIR, stest)
	}
}

func testSAitInitCfg(t *testing.T) {
	saCfgPath = path.Join(*dataDir, "conf", "samples", saonfigDIR)
	// Init config first
	var err error
	saCfg, err = config.NewCGRConfigFromPath(saCfgPath)
	if err != nil {
		t.Error(err)
	}
	if isDispatcherActive {
		saCfg.ListenCfg().RPCJSONListen = ":6012"
	}
}

// Remove data in both rating and accounting db
func testSAitResetDataDb(t *testing.T) {
	if err := engine.InitDataDb(saCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testSAitResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(saCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testSAitStartEngine(t *testing.T) {
	if _, err := engine.StartEngine(saCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testSAitApierRpcConn(t *testing.T) {
	var err error
	saRPC, err = newRPCClient(saCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
	if saConn, err = net.Dial(utils.TCP, "127.0.0.1:5060"); err != nil {
		t.Fatal(err)
	}
}

// Load the tariff plan, creating accounts and their balances
func testSAitTPFromFolder(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tut_sip_redirect")}
	var loadInst utils.LoadInstance
	if err := saRPC.Call(utils.APIerSv2LoadTariffPlanFromFolder, attrs, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

func testSAitStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}

func testSAitSIPRegister(t *testing.T) {
	registerMessage := "REGISTER sip:192.168.58.203 SIP/2.0\r\nCall-ID: d72a4ed6feb4167b5adb208525879db5@0:0:0:0:0:0:0:0\r\nCSeq: 1 REGISTER\r\nFrom: \"1002\" <sip:1002@192.168.58.203>;tag=d28739b9\r\nTo: \"1002\" <sip:1002@192.168.58.203>\r\nVia: SIP/2.0/UDP 192.168.58.201:5060;branch=z9hG4bK-323131-311ce8716a7bf1f6094859ae516a44eb\r\nMax-Forwards: 70\r\nUser-Agent: Jitsi2.11.20200408Linux\r\nExpires: 600\r\nContact: \"1002\" <sip:1002@192.168.58.201:5060;transport=udp;registering_acc=192_168_58_203>;expires=600\r\nContent-Length: 0\r\n"
	if saConn == nil {
		t.Fatal("connection not initialized")
	}
	var err error
	if _, err = saConn.Write([]byte(registerMessage)); err != nil {
		t.Fatal(err)
	}
	buffer := make([]byte, bufferSize)
	if _, err = saConn.Read(buffer); err != nil {
		t.Fatal(err)
	}
	var received sipingo.Message
	if received, err = sipingo.NewMessage(string(buffer)); err != nil {
		t.Fatal(err)
	}

	if expected := "SIP/2.0 405 Method Not Allowed"; received["Request"] != expected {
		t.Errorf("Expected %q, received: %q", expected, received["Request"])
	}
}

func testSAitSIPInvite(t *testing.T) {
	inviteMessage := "INVITE sip:1002@192.168.58.203 SIP/2.0\r\nCall-ID: 4d4d84b0cc83fc90aca41e295cd8ff43@0:0:0:0:0:0:0:0\r\nCSeq: 2 INVITE\r\nFrom: \"1001\" <sip:1001@192.168.58.203>;tag=99f35805\r\nTo: <sip:1002@192.168.58.203>\r\nMax-Forwards: 70\r\nContact: \"1001\" <sip:1001@192.168.58.201:5060;transport=udp;registering_acc=192_168_58_203>\r\nUser-Agent: Jitsi2.11.20200408Linux\r\nContent-Type: application/sdp\r\nVia: SIP/2.0/UDP 192.168.58.201:5060;branch=z9hG4bK-393139-939e89686023b86822cb942ede452b62\r\nProxy-Authorization: Digest username=\"1001\",realm=\"192.168.58.203\",nonce=\"XruO2167ja8uRODnSv8aXqv+/hqPJiXh\",uri=\"sip:1002@192.168.58.203\",response=\"5b814c709d1541d72ea778599c2e48a4\"\r\nContent-Length: 897\r\n\r\nv=0\r\no=1001-jitsi.org 0 0 IN IP4 192.168.58.201\r\ns=-\r\nc=IN IP4 192.168.58.201\r\nt=0 0\r\nm=audio 5000 RTP/AVP 96 97 98 9 100 102 0 8 103 3 104 101\r\na=rtpmap:96 opus/48000/2\r\na=fmtp:96 usedtx=1\r\na=ptime:20\r\na=rtpmap:97 SILK/24000\r\na=rtpmap:98 SILK/16000\r\na=rtpmap:9 G722/8000\r\na=rtpmap:100 speex/32000\r\na=rtpmap:102 speex/16000\r\na=rtpmap:0 PCMU/8000\r\na=rtpmap:8 PCMA/8000\r\na=rtpmap:103 iLBC/8000\r\na=rtpmap:3 GSM/8000\r\na=rtpmap:104 speex/8000\r\na=rtpmap:101 telephone-event/8000\r\na=extmap:1 urn:ietf:params:rtp-hdrext:csrc-audio-level\r\na=extmap:2 urn:ietf:params:rtp-hdrext:ssrc-audio-level\r\na=rtcp-xr:voip-metrics\r\nm=video 5002 RTP/AVP 105 99\r\na=recvonly\r\na=rtpmap:105 h264/90000\r\na=fmtp:105 profile-level-id=42E01f;packetization-mode=1\r\na=imageattr:105 send * recv [x=[1:1920],y=[1:1080]]\r\na=rtpmap:\r\n"
	ack := "ACK sip:1001@192.168.56.203:6060 SIP/2.0\r\nVia: SIP/2.0/UDP 192.168.56.203;rport;branch=z9hG4bKQeB89BamX86UD\r\nMax-Forwards: 69\r\nFrom: \"1001\" <sip:1001@192.168.58.203>;tag=99f35805\r\nTo: <sip:1001@192.168.56.203:6060>\r\nCall-ID: 4d4d84b0cc83fc90aca41e295cd8ff43@0:0:0:0:0:0:0:0\r\nCSeq: 21984733 ACK\r\nContent-Length: 0\r\n"
	if saConn == nil {
		t.Fatal("connection not initialized")
	}
	var err error
	if _, err = saConn.Write([]byte(inviteMessage)); err != nil {
		t.Fatal(err)
	}
	buffer := make([]byte, bufferSize)
	if _, err = saConn.Read(buffer); err != nil {
		t.Fatal(err)
	}
	var received sipingo.Message
	if received, err = sipingo.NewMessage(string(buffer)); err != nil {
		t.Fatal(err)
	}

	if expected := "SIP/2.0 302 Moved Temporarily"; received["Request"] != expected {
		t.Errorf("Expected %q, received: %q", expected, received["Request"])
	}
	if expected := "\"1002\" <sip:1002@cgrates.org>;q=0.7; expires=3600;cgr_cost=0.3;cgr_maxusage=30000000000,\"1002\" <sip:1002@cgrates.net>;q=0.2;cgr_cost=0.6;cgr_maxusage=30000000000,\"1002\" <sip:1002@cgrates.com>;q=0.1;cgr_cost=0.01;cgr_maxusage=30000000000"; received["Contact"] != expected {
		t.Errorf("Expected %q, received: %q", expected, received["Contact"])
	}

	time.Sleep(time.Second)
	buffer = make([]byte, bufferSize)
	if _, err = saConn.Read(buffer); err != nil {
		t.Fatal(err)
	}
	if received, err = sipingo.NewMessage(string(buffer)); err != nil {
		t.Fatal(err)
	}

	if expected := "SIP/2.0 302 Moved Temporarily"; received["Request"] != expected {
		t.Errorf("Expected %q, received: %q", expected, received["Request"])
	}
	if expected := "\"1002\" <sip:1002@cgrates.org>;q=0.7; expires=3600;cgr_cost=0.3;cgr_maxusage=30000000000,\"1002\" <sip:1002@cgrates.net>;q=0.2;cgr_cost=0.6;cgr_maxusage=30000000000,\"1002\" <sip:1002@cgrates.com>;q=0.1;cgr_cost=0.01;cgr_maxusage=30000000000"; received["Contact"] != expected {
		t.Errorf("Expected %q, received: %q", expected, received["Contact"])
	}

	if expected := "<sip:1002@route1.com>"; received["P-Charge-Info"] != expected {
		t.Errorf("Expected %q, received: %q", expected, received["P-Charge-Info"])
	}

	if _, err = saConn.Write([]byte(ack)); err != nil {
		t.Fatal(err)
	}
	buffer = make([]byte, bufferSize)
	saConn.SetDeadline(time.Now().Add(time.Second))
	if _, err = saConn.Read(buffer); err == nil {
		t.Error("Expected error received nil")
	} else if nerr, ok := err.(net.Error); !ok {
		t.Errorf("Expected net.Error received:%v", err)
	} else if !nerr.Timeout() {
		t.Errorf("Expected a timeout error received:%v", err)
	}
}
