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

package agents

import (
	"bytes"
	"encoding/base64"
	"path"
	"sync"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/radigo"
)

/*
TestRadiusCoADisconnect scenario:
1. Configure a radius_agent with:
  - a bidirectional connection to sessions
  - dmr_template field pointing to the predefined *dmr template
  - coa_template field pointing to the predefined *coa template
  - localhost:3799 inside client_da_addresses
  - an auth request processor
  - an accounting request processor

2. Set up a 'client' (acting as a server) that will handle incoming CoA/Disconnect Requests.

3. Send an AccessRequest to cgr-engine's RADIUS server in order to register the packet.

4. Send an AccountingRequest to initialize a session.

5. Send a SessionSv1AlterSessions request, that will send a CoA request to the client. The
client will then verify that the packet was populated correctly.

6. Send a SessionSv1ForceDisconnect request, that will attempt to remotely disconnect the
session created previously and verify the request packet fields.
*/
func TestRadiusCoADisconnect(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	// Set up test environment.
	cfgPath := path.Join(*utils.DataDir, "conf", "samples", "radius_coa_disconnect")
	raDiscCfg, err := config.NewCGRConfigFromPath(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := engine.InitDataDB(raDiscCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDb(raDiscCfg); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.StartEngine(cfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
	defer engine.KillEngine(100)
	raDiscRPC, err := newRPCClient(raDiscCfg.ListenCfg())
	if err != nil {
		t.Fatal(err)
	}
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "oldtutorial")}
	var loadInst utils.LoadInstance
	if err := raDiscRPC.Call(context.Background(), utils.APIerSv2LoadTariffPlanFromFolder, attrs, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*utils.WaitRater) * time.Millisecond)

	var reply string

	// Set Action which will be called by Threshold when account gets debitted
	/*
		actRadCoaAcnt1001 := &utils.AttrSetActions{
			ActionsId: "ACT_RAD_COA_ACNT_1001",
			Actions: []*utils.TPAction{{
				Identifier:      utils.MetaAlterSessions,
				ExtraParameters: "cgrates.org;*string:~*req.Account:1001;1;*radCoATemplate:mycoa;CustomFilter:custom_filter",
			}}}
		if err := raDiscRPC.Call(context.Background(), utils.APIerSv2SetActions,
			actRadCoaAcnt1001, &reply); err != nil {
			t.Error("Got error on APIerSv2.SetActions: ", err.Error())
		} else if reply != utils.OK {
			t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
		}
	*/

	// Set the Threshold profile which will call the action when account will be modified

	tPrfl := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant: "cgrates.org",
			ID:     "THD_ACNT_1001",
			FilterIDs: []string{
				"*string:~*opts.*eventType:AccountUpdate",
				"*string:~*asm.ID:1001",
			},
			//MinHits:   1,
			MaxHits:   1,
			ActionIDs: []string{"LOG_WARNING", "ACT_RAD_COA_ACNT_1001"},
			Async:     true,
		},
	}
	if err := raDiscRPC.Call(context.Background(), utils.APIerSv1SetThresholdProfile,
		tPrfl, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	// Testing the functionality itself starts here.
	var wg sync.WaitGroup
	done := make(chan struct{}) // signal to end the test when the handlers have finished processing
	go func() {
		wg.Wait()
		close(done)
	}()
	wg.Add(2)
	handleDisconnect := func(request *radigo.Packet) (*radigo.Packet, error) {
		defer wg.Done()
		encodedNasIPAddr := "fwAAAQ=="
		decodedNasIPAddr, err := base64.StdEncoding.DecodeString(encodedNasIPAddr)
		if err != nil {
			t.Error("error decoding base64 NAS-IP-Address:", err)
		}
		reply := request.Reply()
		if string(request.AVPs[0].RawValue) != "1001" ||
			!bytes.Equal(request.AVPs[1].RawValue, decodedNasIPAddr) ||
			string(request.AVPs[2].RawValue) != "e4921177ab0e3586c37f6a185864b71a@0:0:0:0:0:0:0:0" ||
			string(request.AVPs[3].RawValue) != "NORMAL_DISCONNECT" {
			t.Errorf("unexpected request received: %v", utils.ToJSON(request))
			reply.Code = radigo.DisconnectNAK
		} else {
			reply.Code = radigo.DisconnectACK
		}
		return reply, nil
	}
	handleCoA := func(request *radigo.Packet) (*radigo.Packet, error) {
		defer wg.Done()
		encodedNasIPAddr := "fwAAAQ=="
		decodedNasIPAddr, err := base64.StdEncoding.DecodeString(encodedNasIPAddr)
		if err != nil {
			t.Error("error decoding base64 NAS-IP-Address:", err)
		}
		reply := request.Reply()
		if string(request.AVPs[0].RawValue) != "1001" ||
			!bytes.Equal(request.AVPs[1].RawValue, decodedNasIPAddr) ||
			string(request.AVPs[2].RawValue) != "e4921177ab0e3586c37f6a185864b71a@0:0:0:0:0:0:0:0" ||
			string(request.AVPs[3].RawValue) != "mycustomvalue" {
			t.Errorf("unexpected request received: %v", utils.ToJSON(request))
			reply.Code = radigo.CoANAK
		} else {
			reply.Code = radigo.CoAACK
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
	testRadClient.server = radigo.NewServer(utils.UDP, "127.0.0.1:3799", secrets, dicts,
		map[radigo.PacketCode]func(*radigo.Packet) (*radigo.Packet, error){
			radigo.DisconnectRequest: handleDisconnect,
			radigo.CoARequest:        handleCoA,
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

	time.Sleep(1 * time.Second) // Give time for the ThresholdS to execute the *alter_sessions API
	if err = raDiscRPC.Call(context.Background(), utils.SessionSv1ForceDisconnect,
		utils.SessionFilterWithEvent{
			Event: map[string]any{
				utils.DisconnectCause: "NORMAL_DISCONNECT",
			},
		}, &reply); err != nil {
		t.Error(err)
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Error("client did not receive the expected requests in time")
	}
}
