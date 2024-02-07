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
	"bytes"
	"encoding/base64"
	"path"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/radigo"
)

/*
TestRadiusDisconnect scenario:
1. Configure a radius_agent with:
  - a bidirectional connection to sessions
  - dmr_template field pointing to the predefined *dmr template
  - localhost:3799 inside client_da_addresses
  - an auth request processor
  - an accounting request processor

2. Set up a 'client' (acting as a server) that will handle incoming Disconnect Requests.

3. Send an AccessRequest to cgr-engine's RADIUS server in order to register the packet.

4. Send an AccountingRequest to initialize a session.

5. Send a SessionSv1ForceDisconnect request, that will attempt to remotely disconnect the
session created previously.

6. Verify that the request fields from the 'client' handler are correctly sent.
*/
func TestRadiusDisconnect(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	// Set up test environment.
	cfgPath := path.Join(*dataDir, "conf", "samples", "radius_disconnect")
	raDiscCfg, err := config.NewCGRConfigFromPath(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := engine.InitDataDb(raDiscCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDb(raDiscCfg); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.StartEngine(cfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
	defer engine.KillEngine(100)
	raDiscRPC, err := newRPCClient(raDiscCfg.ListenCfg())
	if err != nil {
		t.Fatal(err)
	}
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "oldtutorial")}
	var loadInst utils.LoadInstance
	if err := raDiscRPC.Call(context.Background(), utils.APIerSv2LoadTariffPlanFromFolder, attrs, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond)

	// Testing the functionality itself starts here.
	done := make(chan struct{}) // signal to end the test when the handler has finished processing
	handleDisconnect := func(request *radigo.Packet) (*radigo.Packet, error) {
		defer close(done)
		encodedNasIPAddr := "fwAAAQ=="
		decodedNasIPAddr, err := base64.StdEncoding.DecodeString(encodedNasIPAddr)
		if err != nil {
			t.Error("error decoding base64 NAS-IP-Address:", err)
		}
		reply := request.Reply()
		if string(request.AVPs[0].RawValue) != "1001" ||
			!bytes.Equal(request.AVPs[1].RawValue, decodedNasIPAddr) ||
			string(request.AVPs[2].RawValue) != "e4921177ab0e3586c37f6a185864b71a@0:0:0:0:0:0:0:0" ||
			string(request.AVPs[3].RawValue) != "FORCED_DISCONNECT" {
			t.Errorf("unexpected request received: %v", utils.ToJSON(request))
			reply.Code = radigo.DisconnectNAK
		} else {
			reply.Code = radigo.DisconnectACK
		}
		return reply, nil
	}
	type testNAS struct {
		clientAuth *radigo.Client
		clientAcct *radigo.Client
		server     *radigo.Server
	}
	var testRadClient testNAS
	secrets := radigo.NewSecrets(map[string]string{utils.MetaDefault: "CGRateS.org"})
	dicts := radigo.NewDictionaries(map[string]*radigo.Dictionary{utils.MetaDefault: dictRad})
	testRadClient.server = radigo.NewServer("udp", "127.0.0.1:3799", secrets, dicts,
		map[radigo.PacketCode]func(*radigo.Packet) (*radigo.Packet, error){
			radigo.DisconnectRequest: handleDisconnect,
		}, nil, utils.Logger)
	stopChan := make(chan struct{})
	defer close(stopChan)
	go func() {
		err := testRadClient.server.ListenAndServe(stopChan)
		if err != nil {
			t.Error(err)
		}
	}()
	if testRadClient.clientAuth, err = radigo.NewClient(utils.UDP, "127.0.0.1:1812", "CGRateS.org", dictRad, 1, nil, utils.Logger); err != nil {
		t.Fatal(err)
	}
	authReqPacket := testRadClient.clientAuth.NewRequest(radigo.AccessRequest, 1)
	if err := authReqPacket.AddAVPWithName("User-Name", "1001", ""); err != nil {
		t.Error(err)
	}
	if err := authReqPacket.AddAVPWithName("User-Password", "CGRateSPassword1", ""); err != nil {
		t.Error(err)
	}
	authReqPacket.AVPs[1].RawValue = radigo.EncodeUserPassword([]byte("CGRateSPassword1"), []byte("CGRateS.org"), authReqPacket.Authenticator[:])
	if err := authReqPacket.AddAVPWithName("Service-Type", "SIP-Caller-AVPs", ""); err != nil {
		t.Error(err)
	}
	if err := authReqPacket.AddAVPWithName("Called-Station-Id", "1002", ""); err != nil {
		t.Error(err)
	}
	if err := authReqPacket.AddAVPWithName("Acct-Session-Id", "e4921177ab0e3586c37f6a185864b71a@0:0:0:0:0:0:0:0", ""); err != nil {
		t.Error(err)
	}
	if err := authReqPacket.AddAVPWithName("NAS-IP-Address", "127.0.0.1", ""); err != nil {
		t.Error(err)
	}
	if err := authReqPacket.AddAVPWithName("Event-Timestamp", "1497106115", ""); err != nil {
		t.Error(err)
	}

	replyPacket, err := testRadClient.clientAuth.SendRequest(authReqPacket)
	if err != nil {
		t.Fatal(err)
	}
	if replyPacket.Code != radigo.AccessAccept {
		t.Errorf("unexpected reply received to AccessRequest: %+v", utils.ToJSON(replyPacket))
	}

	if testRadClient.clientAcct, err = radigo.NewClient(utils.UDP, "127.0.0.1:1813", "CGRateS.org", dictRad, 1, nil, utils.Logger); err != nil {
		t.Fatal(err)
	}
	accReqPacket := testRadClient.clientAcct.NewRequest(radigo.AccountingRequest, 2)
	if err := accReqPacket.AddAVPWithName("Acct-Status-Type", "Start", ""); err != nil {
		t.Error(err)
	}
	if err := accReqPacket.AddAVPWithName("Event-Timestamp", "1706034095", ""); err != nil {
		t.Error(err)
	}
	if err := accReqPacket.AddAVPWithName("Acct-Session-Id", "e4921177ab0e3586c37f6a185864b71a@0:0:0:0:0:0:0:0", ""); err != nil {
		t.Error(err)
	}
	if err := accReqPacket.AddAVPWithName("User-Name", "1001", ""); err != nil {
		t.Error(err)
	}
	if err := accReqPacket.AddAVPWithName("Called-Station-Id", "1002", ""); err != nil {
		t.Error(err)
	}
	if err := accReqPacket.AddAVPWithName("NAS-Port", "5060", ""); err != nil {
		t.Error(err)
	}
	if err := accReqPacket.AddAVPWithName("Acct-Delay-Time", "0", ""); err != nil {
		t.Error(err)
	}
	if err := accReqPacket.AddAVPWithName("NAS-IP-Address", "127.0.0.1", ""); err != nil {
		t.Error(err)
	}
	replyPacket, err = testRadClient.clientAcct.SendRequest(accReqPacket)
	if err != nil {
		t.Error(err)
	}
	if replyPacket.Code != radigo.AccountingResponse {
		t.Errorf("unexpected reply received to AccountingRequest: %+v", replyPacket)
	}

	var replyFD string
	if err = raDiscRPC.Call(context.Background(), utils.SessionSv1ForceDisconnect, nil, &replyFD); err != nil {
		t.Error(err)
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Error("client did not receive a DisconnectRequest in time")
	}
}
