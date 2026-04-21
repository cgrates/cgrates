//go:build integration

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
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/go-diameter/diam"
	"github.com/cgrates/go-diameter/diam/avp"
	"github.com/cgrates/go-diameter/diam/datatype"
	"github.com/cgrates/go-diameter/diam/dict"
)

func TestDiamSyGeneral(t *testing.T) {
	ng := engine.TestEngine{
		TpFiles: map[string]string{
			utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,PACKAGE_1001,,,`,
			utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
PACKAGE_1001,ACT_TOPUP,*asap,10`,
			utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP,*topup_reset,,,balance_data,*data,,,,,*unlimited,,3072,,,,`,
			utils.ChargersCsv: `#Tenant,ID,FilterIDs,ActivationInterval,RunID,AttributeIDs,Weight
cgrates.org,DEFAULT,,,*default,*none,0`,
			utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_DATA,*any,RT_DATA,*up,4,0,`,
			utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_DATA,0,0.01,1,1,0`,
			utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,data,itsyscom,,RP_1001,`,
			utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_1001,DR_DATA,*any,10`,
		},
		// LogBuffer: &bytes.Buffer{},
	}
	switch *utils.DBType {
	case utils.MetaInternal:
		ng.ConfigPath = filepath.Join(*utils.DataDir, "conf", "samples", "diam_sy_internal")
		if err := os.MkdirAll("/tmp/internal_db", 0755); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := os.RemoveAll("/tmp/internal_db"); err != nil {
				t.Error(err)
			}
		})
	case utils.MetaMySQL:
		ng.ConfigPath = filepath.Join(*utils.DataDir, "conf", "samples", "diam_sy_mysql")
	case utils.MetaMongo:
		ng.ConfigPath = filepath.Join(*utils.DataDir, "conf", "samples", "diam_sy_mongo")
	case utils.MetaPostgres:
		ng.ConfigPath = filepath.Join(*utils.DataDir, "conf", "samples", "diam_sy_postgres")
	default:
		t.Fatal("unsupported dbtype value")
	}
	// t.Cleanup(func() { fmt.Println(ng.LogBuffer) })
	client, cfg := ng.Run(t)

	time.Sleep(100 * time.Millisecond) // wait for DiameterAgent service to start

	// Start monitoring SL
	diamClientSy, err := NewDiameterClient(cfg.DiameterAgentCfg().Listeners[0].Address, "localhost",
		cfg.DiameterAgentCfg().OriginRealm, cfg.DiameterAgentCfg().VendorID,
		cfg.DiameterAgentCfg().ProductName, utils.DiameterFirmwareRevision,
		cfg.DiameterAgentCfg().DictionariesPath, cfg.DiameterAgentCfg().Listeners[0].Network)
	if err != nil {
		t.Fatal(err)
	}
	syOriginID := utils.UUIDSha1Prefix()
	slr := diam.NewRequest(diam.SpendingLimit, 16777302, nil)
	slr.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(syOriginID))
	slr.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	slr.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	slr.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA-DH"))
	slr.NewAVP(avp.DestinationRealm, avp.Mbit, 0, datatype.DiameterIdentity("dr-cgrates.org"))
	slr.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(16777302))
	slr.NewAVP(avp.SLRequestType, avp.Vbit, 10415, datatype.Enumerated(0)) //INITIAL_REQUEST (0)
	slr.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(0)),
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("1001")), // Subscription-Id-Data (MSISDN)
		}})
	slr.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(1)),
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("104502200011")), // Subscription-Id-Data (IMSI)
		}})
	// t.Log("sendingg msg: ", slr.PrettyDump())
	if err := diamClientSy.SendMessage(slr); err != nil {
		t.Errorf("failed to send diameter message: %v", err)
	}

	reply := diamClientSy.ReceivedMessage(2 * time.Second)
	if reply == nil {
		t.Fatal("received empty reply")
	}
	// t.Log(reply.PrettyDump())
	avps, err := reply.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID)
	if err != nil {
		t.Error(err)
	}
	if len(avps) == 0 {
		t.Fatal("missing AVPs in reply")
	}

	resultCode, err := diamAVPAsString(avps[0])
	if err != nil {
		t.Error(err)
	}
	if resultCode != "2001" {
		t.Errorf("Result-Code=%s, want 2001", resultCode)
	}

	expBalance := float64(3072)
	var acnt *engine.Account
	attrsAcnt := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1001",
	}
	if err = client.Call(context.Background(), utils.APIerSv2GetAccount, attrsAcnt, &acnt); err != nil {
		t.Errorf("APIerSv1.GetAccount unexpected err: %v", err)
	} else if rply := acnt.BalanceMap[utils.MetaData].GetTotalValue(); rply != expBalance {
		t.Errorf("APIerSv1.GetAccount: data_balance: %f, want: %f", rply, expBalance)
	}

	var replyActSess []*sessions.ExternalSession // find indexed Sy sessions active
	if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions, utils.SessionFilter{Filters: []string{"*string:~*req.RequestType:*sy"}}, &replyActSess); err != nil {
		t.Error(err)
	}
	if len(replyActSess) != 1 {
		t.Errorf("expected 1 active sessions, received <%v>", replyActSess)
	}
	// t.Log(utils.ToIJSON(replyActSess))

	// get thresholds
	var tIDs []string
	if err := client.Call(context.Background(), utils.ThresholdSv1GetThresholdIDs,
		&utils.TenantWithAPIOpts{}, &tIDs); err != nil {
		t.Error(err)
	}
	if len(tIDs) != 1 {
		t.Errorf("expected 1 threshold profile to be created, received <%v>", tIDs)
	} else if tIDs[0] != utils.MetaSy+utils.Underline+syOriginID {
		t.Errorf("expected <%v>, received <%v>", utils.MetaSy+utils.Underline+syOriginID, tIDs[0])
	}
	expTp := &engine.ThresholdProfile{
		Tenant: "cgrates.org",
		ID:     utils.MetaSy + utils.Underline + syOriginID,
		FilterIDs: []string{
			utils.ConcatenatedKey(utils.MetaLessOrEqual, utils.DynamicDataPrefix+utils.MetaAsm+utils.NestingSep+utils.BalanceSummaries+utils.NestingSep+"balance_data"+utils.NestingSep+utils.Value, "1023"),
			utils.ConcatenatedKey(utils.MetaString, utils.MetaDynReq+utils.NestingSep+utils.ID, "1001"),
		},
		MaxHits:   1,
		MinHits:   1,
		Async:     true,
		ActionIDs: []string{utils.MetaSyPublish},
	}
	tp := &engine.ThresholdProfile{}
	if err := client.Call(context.Background(), utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{ID: tIDs[0]}, &tp); err != nil {
		t.Error(err)
	}
	expTp.ActivationInterval = tp.ActivationInterval
	slices.Sort(tp.FilterIDs) // sort filters received since they are processed from a map
	if !reflect.DeepEqual(tp, expTp) {
		t.Errorf("expected <%v>\nreceived\n<%v>", utils.ToJSON(expTp), utils.ToJSON(tp))
	}

	diamClientRo, err := NewDiameterClient(cfg.DiameterAgentCfg().Listeners[0].Address, "localhost",
		cfg.DiameterAgentCfg().OriginRealm, cfg.DiameterAgentCfg().VendorID,
		cfg.DiameterAgentCfg().ProductName, utils.DiameterFirmwareRevision,
		cfg.DiameterAgentCfg().DictionariesPath, cfg.DiameterAgentCfg().Listeners[0].Network)
	if err != nil {
		t.Fatal(err)
	}
	ccr := diam.NewRequest(diam.CreditControl, 4, nil)
	ccrOriginID := utils.UUIDSha1Prefix()
	ccr.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(ccrOriginID))
	ccr.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	ccr.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	ccr.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	ccr.NewAVP(avp.DestinationRealm, avp.Mbit, 0, datatype.DiameterIdentity("dr-cgrates.org"))
	ccr.NewAVP(avp.ServiceContextID, avp.Mbit, 0, datatype.UTF8String("message@DiamItCCRData"))
	ccr.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(1))
	ccr.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(0))
	ccr.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA-DH"))
	ccr.NewAVP(avp.OriginStateID, avp.Mbit, 0, datatype.Unsigned32(time.Now().Unix()))
	ccr.NewAVP(avp.EventTimestamp, avp.Mbit, 0, datatype.Time(time.Date(2018, 10, 5, 11, 43, 10, 0, time.UTC)))
	ccr.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(0)),
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("1001")), // Subscription-Id-Data
		}})
	ccr.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(1)),
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("104502200011")), // Subscription-Id-Data
		}})
	ccr.NewAVP(avp.MultipleServicesIndicator, avp.Mbit, 0, datatype.Enumerated(1))
	ccr.NewAVP(avp.MultipleServicesCreditControl, avp.Mbit, 0, &diam.GroupedAVP{ // Multiple-Services-Credit-Control
		AVP: []*diam.AVP{
			diam.NewAVP(avp.RatingGroup, avp.Mbit, 0, datatype.Unsigned32(20000)), // Rating-Group
		},
	})
	ccr.NewAVP(avp.ServiceInformation, avp.Mbit, 10415,
		&diam.GroupedAVP{
			AVP: []*diam.AVP{
				diam.NewAVP(avp.PSInformation, avp.Mbit, 10415,
					&diam.GroupedAVP{
						AVP: []*diam.AVP{
							diam.NewAVP(avp.CalledStationID, avp.Mbit, 0, datatype.UTF8String("itsyscom")),
							diam.NewAVP(avp.TGPPSGSNMCCMNC, avp.Mbit, 10415, datatype.OctetString("1002")),
						},
					},
				),
			},
		},
	)

	// t.Log("sendingg CCR msg: ", ccr.PrettyDump())
	if err := diamClientRo.SendMessage(ccr); err != nil {
		t.Errorf("failed to send diameter message: %v", err)
	}

	reply = diamClientRo.ReceivedMessage(2 * time.Second)
	if reply == nil {
		t.Error("received empty reply")
	}
	// t.Log(reply.PrettyDump())

	avps, err = reply.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID)
	if err != nil {
		t.Error(err)
	}
	if len(avps) == 0 {
		t.Error("missing AVPs in reply")
	}

	resultCode, err = diamAVPAsString(avps[0])
	if err != nil {
		t.Error(err)
	}
	if resultCode != "2001" {
		t.Errorf("Result-Code=%s, want %s", resultCode, "2001")
	}

	expBalance = float64(1024) // CCR init should take 2048 units
	if err = client.Call(context.Background(), utils.APIerSv2GetAccount, attrsAcnt, &acnt); err != nil {
		t.Errorf("APIerSv1.GetAccount unexpected err: %v", err)
	} else if rply := acnt.BalanceMap[utils.MetaData].GetTotalValue(); rply != expBalance {
		t.Errorf("APIerSv1.GetAccount: data_balance: %f, want: %f", rply, expBalance)
	}
	// find indexed sy sessions
	if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions, utils.SessionFilter{Filters: []string{"*string:~*req.RequestType:*sy"}}, &replyActSess); err != nil {
		t.Error(err)
	}
	if len(replyActSess) != 1 {
		t.Errorf("expected 1 active sessions, received <%v>", replyActSess)
	}
	// t.Log(utils.ToIJSON(replyActSess))
	// find all active sessions
	if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions, utils.SessionFilter{}, &replyActSess); err != nil {
		t.Error(err)
	}
	if len(replyActSess) != 2 {
		t.Errorf("expected 2 active sessions, received <%v>", replyActSess)
	}
	// t.Log(utils.ToIJSON(replyActSess))

	var ccaLck sync.RWMutex
	ccaReceived := false
	snaSent := make(chan struct{})
	// prepare to receive SNR when updating CCR
	go func() {
		// receive SNR
		reply = diamClientSy.ReceivedMessage(2 * time.Second)
		if reply == nil {
			t.Fatal("Received empty reply")
		} else {
			switch reply.Header.CommandCode {
			case 8388636:
				expected := fmt.Sprintf(`Spending-Status-Notification-Request (SNR)
%s
	Session-Id {Code:263,Flags:0x40,Length:16,VendorId:0,Value:UTF8String{%s},Padding:1}
	Origin-Host {Code:264,Flags:0x40,Length:16,VendorId:0,Value:DiameterIdentity{CGR-DA},Padding:2}
	Origin-Realm {Code:296,Flags:0x40,Length:20,VendorId:0,Value:DiameterIdentity{cgrates.org},Padding:1}
	Destination-Realm {Code:283,Flags:0x40,Length:24,VendorId:0,Value:DiameterIdentity{dr-cgrates.org},Padding:2}
	Destination-Host {Code:293,Flags:0x40,Length:20,VendorId:0,Value:DiameterIdentity{CGR-DA-DH},Padding:3}
	Auth-Application-Id {Code:258,Flags:0x40,Length:12,VendorId:0,Value:Unsigned32{16777302}}
	Policy-Counter-Status-Report {Code:2903,Flags:0xc0,Length:96,VendorId:10415,Value:Grouped{
		Policy-Counter-Identifier {Code:2901,Flags:0xc0,Length:20,VendorId:10415,Value:UTF8String{Monthly},Padding:1},
		Policy-Counter-Status {Code:2902,Flags:0xc0,Length:20,VendorId:10415,Value:UTF8String{512KBPS},Padding:1},
		Pending-Policy-Counter-Information {Code:2905,Flags:0xc0,Length:44,VendorId:10415,Value:Grouped{
			Policy-Counter-Status {Code:2902,Flags:0xc0,Length:16,VendorId:10415,Value:UTF8String{30GB},Padding:0},
			Pending-Policy-Counter-Change-Time {Code:2906,Flags:0xc0,Length:16,VendorId:10415,Value:Time{%v}},
		}}
	}}
`, reply.Header, syOriginID, time.Time(reply.AVP[6].Data.(*diam.GroupedAVP).AVP[2].Data.(*diam.GroupedAVP).AVP[1].Data.(datatype.Time)))
				if expected != reply.String() {
					t.Errorf("expected \n<%v>, \nreceived \n<%v>", expected, reply.String())
				}
				// Send SNA
				sna := diam.NewRequest(SSN, 16777302, nil)
				sna.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(syOriginID))
				sna.NewAVP(avp.ResultCode, avp.Mbit, 0, datatype.Unsigned32(2001))
				sna.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
				sna.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
				// t.Log("sendingg msg: ", sna.PrettyDump())
				if err := diamClientSy.SendMessage(sna); err != nil {
					t.Errorf("failed to send diameter message: %v", err)
				}
			case diam.CreditControl:
				ccaLck.Lock()
				ccaReceived = true
				ccaLck.Unlock()
				if reply == nil {
					t.Fatal("received empty reply")
				}
				// t.Log(reply.PrettyDump())

				avps, err = reply.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID)
				if err != nil {
					t.Error(err)
				}
				if len(avps) == 0 {
					t.Error("missing AVPs in reply")
				}
				resultCode, err = diamAVPAsString(avps[0])
				if err != nil {
					t.Error(err)
				}
				if resultCode != "2001" {
					t.Errorf("Result-Code=%s, want %s", resultCode, "2001")
				}

				expBalance = float64(0) // CCR update should take 2048 units
				if err = client.Call(context.Background(), utils.APIerSv2GetAccount, attrsAcnt, &acnt); err != nil {
					t.Errorf("APIerSv1.GetAccount unexpected err: %v", err)
				} else if rply := acnt.BalanceMap[utils.MetaData].GetTotalValue(); rply != expBalance {
					t.Errorf("APIerSv1.GetAccount: data_balance: %f, want: %f", rply, expBalance)
				}

				// find indexed sy sessions
				if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions, utils.SessionFilter{Filters: []string{"*string:~*req.RequestType:*sy"}}, &replyActSess); err != nil {
					t.Error(err)
				}
				if len(replyActSess) != 1 {
					t.Errorf("expected 1 active sessions, received <%v>", replyActSess)
				}
				if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions, utils.SessionFilter{}, &replyActSess); err != nil {
					t.Error(err)
				}
				if len(replyActSess) != 2 {
					t.Errorf("expected 2 active sessions, received <%v>", replyActSess)
				}
				// t.Log(utils.ToIJSON(replyActSess))
				reply = diamClientSy.ReceivedMessage(2 * time.Second)
				if reply == nil {
					t.Fatal("Received empty reply")
				}
				expected := fmt.Sprintf(`Spending-Status-Notification-Request (SNR)
%s
	Session-Id {Code:263,Flags:0x40,Length:16,VendorId:0,Value:UTF8String{%s},Padding:1}
	Origin-Host {Code:264,Flags:0x40,Length:16,VendorId:0,Value:DiameterIdentity{CGR-DA},Padding:2}
	Origin-Realm {Code:296,Flags:0x40,Length:20,VendorId:0,Value:DiameterIdentity{cgrates.org},Padding:1}
	Destination-Realm {Code:283,Flags:0x40,Length:24,VendorId:0,Value:DiameterIdentity{dr-cgrates.org},Padding:2}
	Destination-Host {Code:293,Flags:0x40,Length:20,VendorId:0,Value:DiameterIdentity{CGR-DA-DH},Padding:3}
	Auth-Application-Id {Code:258,Flags:0x40,Length:12,VendorId:0,Value:Unsigned32{16777302}}
	Policy-Counter-Status-Report {Code:2903,Flags:0xc0,Length:96,VendorId:10415,Value:Grouped{
		Policy-Counter-Identifier {Code:2901,Flags:0xc0,Length:20,VendorId:10415,Value:UTF8String{Monthly},Padding:1},
		Policy-Counter-Status {Code:2902,Flags:0xc0,Length:20,VendorId:10415,Value:UTF8String{512KBPS},Padding:1},
		Pending-Policy-Counter-Information {Code:2905,Flags:0xc0,Length:44,VendorId:10415,Value:Grouped{
			Policy-Counter-Status {Code:2902,Flags:0xc0,Length:16,VendorId:10415,Value:UTF8String{30GB},Padding:0},
			Pending-Policy-Counter-Change-Time {Code:2906,Flags:0xc0,Length:16,VendorId:10415,Value:Time{%v}},
		}}
	}}
`, reply.Header, syOriginID, time.Time(reply.AVP[6].Data.(*diam.GroupedAVP).AVP[2].Data.(*diam.GroupedAVP).AVP[1].Data.(datatype.Time)))
				if expected != reply.String() {
					t.Errorf("expected \n<%v>, \nreceived \n<%v>", expected, reply.String())
				}
				// Send SNA
				sna := diam.NewRequest(SSN, 16777302, nil)
				sna.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(syOriginID))
				sna.NewAVP(avp.ResultCode, avp.Mbit, 0, datatype.Unsigned32(2001))
				sna.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
				sna.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
				// t.Log("sendingg msg: ", sna.PrettyDump())
				if err := diamClientSy.SendMessage(sna); err != nil {
					t.Errorf("failed to send diameter message: %v", err)
				}
			default:
				t.Fatal("received wrong reply: ", reply.PrettyDump())
			}
		}
		close(snaSent)
	}()

	ccrU := diam.NewRequest(diam.CreditControl, 4, nil)
	ccrU.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(ccrOriginID))
	ccrU.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	ccrU.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	ccrU.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	ccrU.NewAVP(avp.DestinationRealm, avp.Mbit, 0, datatype.DiameterIdentity("dr-cgrates.org"))
	ccrU.NewAVP(avp.ServiceContextID, avp.Mbit, 0, datatype.UTF8String("message@DiamItCCRData"))
	ccrU.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(2))
	ccrU.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(1))
	ccrU.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA-DH"))
	ccrU.NewAVP(avp.OriginStateID, avp.Mbit, 0, datatype.Unsigned32(time.Now().Unix()))
	ccrU.NewAVP(avp.EventTimestamp, avp.Mbit, 0, datatype.Time(time.Date(2018, 10, 5, 11, 47, 10, 0, time.UTC)))
	ccrU.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(0)),
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("1001")), // Subscription-Id-Data
		}})
	ccrU.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(1)),
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("104502200011")), // Subscription-Id-Data
		}})
	ccrU.NewAVP(avp.MultipleServicesIndicator, avp.Mbit, 0, datatype.Enumerated(1))
	ccrU.NewAVP(avp.MultipleServicesCreditControl, avp.Mbit, 0, &diam.GroupedAVP{ // Multiple-Services-Credit-Control
		AVP: []*diam.AVP{
			diam.NewAVP(avp.RatingGroup, avp.Mbit, 0, datatype.Unsigned32(20000)), // Rating-Group
			diam.NewAVP(avp.UsedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{ // Used-Service-Unit
				AVP: []*diam.AVP{
					diam.NewAVP(avp.CCTotalOctets, avp.Mbit, 0, datatype.Unsigned64(7640)),  // CC-Total-Octets
					diam.NewAVP(avp.CCInputOctets, avp.Mbit, 0, datatype.Unsigned64(5337)),  // CC-Input-Octets
					diam.NewAVP(avp.CCOutputOctets, avp.Mbit, 0, datatype.Unsigned64(2303)), // CC-Output-Octets
				},
			}),
		},
	})
	ccrU.NewAVP(avp.ServiceInformation, avp.Mbit, 10415,
		&diam.GroupedAVP{
			AVP: []*diam.AVP{
				diam.NewAVP(avp.PSInformation, avp.Mbit, 10415,
					&diam.GroupedAVP{
						AVP: []*diam.AVP{
							diam.NewAVP(avp.CalledStationID, avp.Mbit, 0, datatype.UTF8String("itsyscom")),
							diam.NewAVP(avp.TGPPSGSNMCCMNC, avp.Mbit, 10415, datatype.OctetString("1002")),
						},
					},
				),
			},
		},
	)

	// t.Log("sendingg CCR-U msg: ", ccrU.PrettyDump())
	if err := diamClientRo.SendMessage(ccrU); err != nil {
		t.Errorf("failed to send diameter message: %v", err)
	}

	select { // make sure sna is sent before continuing to not mix replies
	case <-snaSent:
	case <-time.After(5 * time.Second):
		t.Fatal("took too long to get reply from SNR")
	}

	ccaLck.RLock()
	if !ccaReceived {
		reply = diamClientRo.ReceivedMessage(5 * time.Second)
		if reply == nil {
			t.Fatal("received empty reply")
		}
		// t.Log(reply.PrettyDump())

		avps, err = reply.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID)
		if err != nil {
			t.Error(err)
		}
		if len(avps) == 0 {
			t.Error("missing AVPs in reply")
		}
		resultCode, err = diamAVPAsString(avps[0])
		if err != nil {
			t.Error(err)
		}
		if resultCode != "2001" {
			t.Errorf("Result-Code=%s, want %s", resultCode, "2001")
		}

		expBalance = float64(0) // CCR update should take 2048 units
		if err = client.Call(context.Background(), utils.APIerSv2GetAccount, attrsAcnt, &acnt); err != nil {
			t.Errorf("APIerSv1.GetAccount unexpected err: %v", err)
		} else if rply := acnt.BalanceMap[utils.MetaData].GetTotalValue(); rply != expBalance {
			t.Errorf("APIerSv1.GetAccount: data_balance: %f, want: %f", rply, expBalance)
		}

		// find indexed sy sessions
		if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions, utils.SessionFilter{Filters: []string{"*string:~*req.RequestType:*sy"}}, &replyActSess); err != nil {
			t.Error(err)
		}
		if len(replyActSess) != 1 {
			t.Errorf("expected 1 active sessions, received <%v>", replyActSess)
		}
		if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions, utils.SessionFilter{}, &replyActSess); err != nil {
			t.Error(err)
		}
		if len(replyActSess) != 2 {
			t.Errorf("expected 2 active sessions, received <%v>", replyActSess)
		}
		// t.Log(utils.ToIJSON(replyActSess))
	}
	ccaLck.RUnlock()

	// Change session policy subscription
	slri := diam.NewRequest(diam.SpendingLimit, 16777302, nil)
	slri.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(syOriginID))
	slri.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	slri.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	slri.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA-DH"))
	slri.NewAVP(avp.DestinationRealm, avp.Mbit, 0, datatype.DiameterIdentity("dr-cgrates.org"))
	slri.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(16777302))
	slri.NewAVP(avp.SLRequestType, avp.Vbit, 10415, datatype.Enumerated(1)) // INTERMEDIATE_REQUEST (1)
	slri.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(0)),
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("1001")), // Subscription-Id-Data (MSISDN)
		}})
	slri.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(1)),
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("104502200011")), // Subscription-Id-Data (IMSI)
		}})
	// t.Log("sendingg msg: ", slri.PrettyDump())
	if err := diamClientSy.SendMessage(slri); err != nil {
		t.Errorf("failed to send diameter message: %v", err)
	}

	reply = diamClientSy.ReceivedMessage(2 * time.Second)
	if reply == nil {
		t.Fatal("received empty reply")
	}
	// t.Log(reply.PrettyDump())
	avps, err = reply.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID)
	if err != nil {
		t.Error(err)
	}
	if len(avps) == 0 {
		t.Fatal("missing AVPs in reply: ", reply.PrettyDump())
	}

	resultCode, err = diamAVPAsString(avps[0])
	if err != nil {
		t.Error(err)
	}
	if resultCode != "2001" {
		t.Errorf("Result-Code=%s, want 2001", resultCode)
	}

	expBalance = float64(0)
	if err = client.Call(context.Background(), utils.APIerSv2GetAccount, attrsAcnt, &acnt); err != nil {
		t.Errorf("APIerSv1.GetAccount unexpected err: %v", err)
	} else if rply := acnt.BalanceMap[utils.MetaData].GetTotalValue(); rply != expBalance {
		t.Errorf("APIerSv1.GetAccount: data_balance: %f, want: %f", rply, expBalance)
	}

	if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions, utils.SessionFilter{Filters: []string{"*string:~*req.RequestType:*sy"}}, &replyActSess); err != nil {
		t.Error(err)
	}
	if len(replyActSess) != 1 {
		t.Errorf("expected 1 active sessions, received <%v>", replyActSess)
	}
	// t.Log(utils.ToIJSON(replyActSess))

	// get thresholds
	if err := client.Call(context.Background(), utils.ThresholdSv1GetThresholdIDs,
		&utils.TenantWithAPIOpts{}, &tIDs); err != nil {
		t.Error(err)
	}
	if len(tIDs) != 1 {
		t.Errorf("expected 1 threshold profile to be created, received <%v>", tIDs)
	} else if tIDs[0] != utils.MetaSy+utils.Underline+syOriginID {
		t.Errorf("expected <%v>, received <%v>", utils.MetaSy+utils.Underline+syOriginID, tIDs[0])
	}
	expTp = &engine.ThresholdProfile{
		Tenant: "cgrates.org",
		ID:     utils.MetaSy + utils.Underline + syOriginID,
		FilterIDs: []string{
			utils.ConcatenatedKey(utils.MetaLessOrEqual, utils.DynamicDataPrefix+utils.MetaAsm+utils.NestingSep+utils.BalanceSummaries+utils.NestingSep+"balance_data"+utils.NestingSep+utils.Value, "-1"),
			utils.ConcatenatedKey(utils.MetaString, utils.MetaDynReq+utils.NestingSep+utils.ID, "1001"),
		},
		MaxHits:   1,
		MinHits:   1,
		Async:     true,
		ActionIDs: []string{utils.MetaSyPublish},
	}
	tp = &engine.ThresholdProfile{}
	if err := client.Call(context.Background(), utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{ID: tIDs[0]}, &tp); err != nil {
		t.Error(err)
	}
	expTp.ActivationInterval = tp.ActivationInterval
	slices.Sort(tp.FilterIDs) // sort filters received since they are processed from a map
	if !reflect.DeepEqual(tp, expTp) {
		t.Errorf("expected <%v>\nreceived\n<%v>", utils.ToJSON(expTp), utils.ToJSON(tp))
	}

	ccrT := diam.NewRequest(diam.CreditControl, 4, nil)
	ccrT.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(ccrOriginID))
	ccrT.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	ccrT.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	ccrT.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	ccrT.NewAVP(avp.DestinationRealm, avp.Mbit, 0, datatype.DiameterIdentity("dr-cgrates.org"))
	ccrT.NewAVP(avp.ServiceContextID, avp.Mbit, 0, datatype.UTF8String("message@DiamItCCRData"))
	ccrT.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(3))
	ccrT.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(2))
	ccrT.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA-DH"))
	ccrT.NewAVP(avp.OriginStateID, avp.Mbit, 0, datatype.Unsigned32(time.Now().Unix()))
	ccrT.NewAVP(avp.EventTimestamp, avp.Mbit, 0, datatype.Time(time.Date(2018, 10, 5, 11, 50, 10, 0, time.UTC)))
	ccrT.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(0)),
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("1001")), // Subscription-Id-Data
		}})
	ccrT.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(1)),
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("104502200011")), // Subscription-Id-Data
		}})
	ccrT.NewAVP(avp.TerminationCause, avp.Mbit, 0, datatype.Enumerated(1))
	ccrT.NewAVP(avp.MultipleServicesIndicator, avp.Mbit, 0, datatype.Enumerated(1))
	ccrT.NewAVP(avp.MultipleServicesCreditControl, avp.Mbit, 0, &diam.GroupedAVP{ // Multiple-Services-Credit-Control
		AVP: []*diam.AVP{
			diam.NewAVP(avp.RatingGroup, avp.Mbit, 0, datatype.Unsigned32(20000)), // Rating-Group
			diam.NewAVP(avp.UsedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{ // Used-Service-Unit
				AVP: []*diam.AVP{
					diam.NewAVP(avp.CCTotalOctets, avp.Mbit, 0, datatype.Unsigned64(7640)),  // CC-Total-Octets
					diam.NewAVP(avp.CCInputOctets, avp.Mbit, 0, datatype.Unsigned64(5337)),  // CC-Input-Octets
					diam.NewAVP(avp.CCOutputOctets, avp.Mbit, 0, datatype.Unsigned64(2303)), // CC-Output-Octets
				},
			}),
		},
	})
	ccrT.NewAVP(avp.ServiceInformation, avp.Mbit, 10415,
		&diam.GroupedAVP{
			AVP: []*diam.AVP{
				diam.NewAVP(avp.PSInformation, avp.Mbit, 10415,
					&diam.GroupedAVP{
						AVP: []*diam.AVP{
							diam.NewAVP(avp.CalledStationID, avp.Mbit, 0, datatype.UTF8String("itsyscom")),
							diam.NewAVP(avp.TGPPSGSNMCCMNC, avp.Mbit, 10415, datatype.OctetString("1002")),
						},
					},
				),
			},
		},
	)

	// t.Log("sendingg CCR-T msg: ", ccrT.PrettyDump())
	if err := diamClientRo.SendMessage(ccrT); err != nil {
		t.Errorf("failed to send diameter message: %v", err)
	}

	reply = diamClientRo.ReceivedMessage(2 * time.Second)
	if reply == nil {
		t.Error("received empty reply")
	}
	// t.Log(reply.PrettyDump())

	avps, err = reply.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID)
	if err != nil {
		t.Error(err)
	}
	if len(avps) == 0 {
		t.Error("missing AVPs in reply")
	}

	resultCode, err = diamAVPAsString(avps[0])
	if err != nil {
		t.Error(err)
	}
	if resultCode != "2001" {
		t.Errorf("Result-Code=%s, want %s", resultCode, "2001")
	}

	expBalance = float64(0) // CCR update should take 2048 units
	if err = client.Call(context.Background(), utils.APIerSv2GetAccount, attrsAcnt, &acnt); err != nil {
		t.Errorf("APIerSv1.GetAccount unexpected err: %v", err)
	} else if rply := acnt.BalanceMap[utils.MetaData].GetTotalValue(); rply != expBalance {
		t.Errorf("APIerSv1.GetAccount: data_balance: %f, want: %f", rply, expBalance)
	}

	// find indexed sy sessions
	if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions, utils.SessionFilter{Filters: []string{"*string:~*req.RequestType:*sy"}}, &replyActSess); err != nil {
		t.Error(err)
	}
	if len(replyActSess) != 1 {
		t.Errorf("expected 1 active sessions, received <%v>", replyActSess)
	}
	if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions, utils.SessionFilter{}, &replyActSess); err != nil {
		t.Error(err)
	}
	if len(replyActSess) != 1 {
		t.Errorf("expected 1 active sessions, received <%v>", replyActSess)
	}
	// t.Log(utils.ToIJSON(replyActSess))
	str := diam.NewRequest(diam.SessionTermination, 16777302, nil)
	str.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(syOriginID))
	str.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	str.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	str.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA-DH"))
	str.NewAVP(avp.DestinationRealm, avp.Mbit, 0, datatype.DiameterIdentity("dr-cgrates.org"))
	str.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(16777302))
	str.NewAVP(avp.TerminationCause, avp.Mbit, 0, datatype.Enumerated(1))
	// t.Log("sendingg STR msg: ", str.PrettyDump())
	if err := diamClientSy.SendMessage(str); err != nil {
		t.Errorf("failed to send diameter message: %v", err)
	}

	reply = diamClientSy.ReceivedMessage(2 * time.Second)
	if reply == nil {
		t.Fatal("received empty reply")
	}
	// t.Log(reply.PrettyDump())
	avps, err = reply.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID)
	if err != nil {
		t.Error(err)
	}
	if len(avps) == 0 {
		t.Fatal("missing AVPs in reply")
	}

	resultCode, err = diamAVPAsString(avps[0])
	if err != nil {
		t.Error(err)
	}
	if resultCode != "2001" {
		t.Errorf("Result-Code=%s, want 2001", resultCode)
	}

	expBalance = float64(0)
	if err = client.Call(context.Background(), utils.APIerSv2GetAccount, attrsAcnt, &acnt); err != nil {
		t.Errorf("APIerSv1.GetAccount unexpected err: %v", err)
	} else if rply := acnt.BalanceMap[utils.MetaData].GetTotalValue(); rply != expBalance {
		t.Errorf("APIerSv1.GetAccount: data_balance: %f, want: %f", rply, expBalance)
	}
	if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions, utils.SessionFilter{}, &replyActSess); err == nil || err.Error() != "NOT_FOUND" {
		t.Errorf("expected error <NOT_FOUND>, received <%v>", err)
	}
	// thresholds profile should be removed on STR
	if err := client.Call(context.Background(), utils.ThresholdSv1GetThresholdIDs,
		&utils.TenantWithAPIOpts{}, &tIDs); err != nil {
		t.Error(err)
	} else if len(tIDs) != 0 {
		t.Errorf("expected no Threshold profiles, received <%v>", tIDs)
	}
}

func TestDiamSyRestart(t *testing.T) {

	ng := engine.TestEngine{
		TpFiles: map[string]string{
			utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,PACKAGE_1001,,,`,
			utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
PACKAGE_1001,ACT_TOPUP,*asap,10`,
			utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP,*topup_reset,,,balance_data,*data,,,,,*unlimited,,3072,,,,`,
			utils.ChargersCsv: `#Tenant,ID,FilterIDs,ActivationInterval,RunID,AttributeIDs,Weight
cgrates.org,DEFAULT,,,*default,*none,0`,
			utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_DATA,*any,RT_DATA,*up,4,0,`,
			utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_DATA,0,0.01,1,1,0`,
			utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,data,itsyscom,,RP_1001,`,
			utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_1001,DR_DATA,*any,10`,
		},
		// LogBuffer: &bytes.Buffer{},
	}
	switch *utils.DBType {
	case utils.MetaInternal:
		ng.ConfigPath = filepath.Join(*utils.DataDir, "conf", "samples", "diam_sy_internal")
		if err := os.MkdirAll("/tmp/internal_db", 0755); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := os.RemoveAll("/tmp/internal_db"); err != nil {
				t.Error(err)
			}
		})
	case utils.MetaMySQL:
		ng.ConfigPath = filepath.Join(*utils.DataDir, "conf", "samples", "diam_sy_mysql")
	case utils.MetaMongo:
		ng.ConfigPath = filepath.Join(*utils.DataDir, "conf", "samples", "diam_sy_mongo")
	case utils.MetaPostgres:
		ng.ConfigPath = filepath.Join(*utils.DataDir, "conf", "samples", "diam_sy_postgres")
	default:
		t.Fatal("unsupported dbtype value")
	}
	// t.Cleanup(func() { fmt.Println(ng.LogBuffer) })
	client, cfg := ng.Run(t)

	time.Sleep(100 * time.Millisecond) // wait for DiameterAgent service to start

	// Start monitoring SL
	diamClientSy, err := NewDiameterClient(cfg.DiameterAgentCfg().Listeners[0].Address, "localhost",
		cfg.DiameterAgentCfg().OriginRealm, cfg.DiameterAgentCfg().VendorID,
		cfg.DiameterAgentCfg().ProductName, utils.DiameterFirmwareRevision,
		cfg.DiameterAgentCfg().DictionariesPath, cfg.DiameterAgentCfg().Listeners[0].Network)
	if err != nil {
		t.Fatal(err)
	}
	syOriginID := utils.UUIDSha1Prefix()
	slr := diam.NewRequest(diam.SpendingLimit, 16777302, nil)
	slr.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(syOriginID))
	slr.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	slr.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	slr.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA-DH"))
	slr.NewAVP(avp.DestinationRealm, avp.Mbit, 0, datatype.DiameterIdentity("dr-cgrates.org"))
	slr.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(16777302))
	slr.NewAVP(avp.SLRequestType, avp.Vbit, 10415, datatype.Enumerated(0)) //INITIAL_REQUEST (0)
	slr.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(0)),
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("1001")), // Subscription-Id-Data (MSISDN)
		}})
	slr.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(1)),
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("104502200011")), // Subscription-Id-Data (IMSI)
		}})
	// t.Log("sendingg msg: ", slr.PrettyDump())
	if err := diamClientSy.SendMessage(slr); err != nil {
		t.Errorf("failed to send diameter message: %v", err)
	}

	reply := diamClientSy.ReceivedMessage(2 * time.Second)
	if reply == nil {
		t.Fatal("received empty reply")
	}
	// t.Log(reply.PrettyDump())
	avps, err := reply.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID)
	if err != nil {
		t.Error(err)
	}
	if len(avps) == 0 {
		t.Fatal("missing AVPs in reply")
	}

	resultCode, err := diamAVPAsString(avps[0])
	if err != nil {
		t.Error(err)
	}
	if resultCode != "2001" {
		t.Errorf("Result-Code=%s, want 2001", resultCode)
	}

	expBalance := float64(3072)
	var acnt *engine.Account
	attrsAcnt := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1001",
	}
	if err = client.Call(context.Background(), utils.APIerSv2GetAccount, attrsAcnt, &acnt); err != nil {
		t.Errorf("APIerSv1.GetAccount unexpected err: %v", err)
	} else if rply := acnt.BalanceMap[utils.MetaData].GetTotalValue(); rply != expBalance {
		t.Errorf("APIerSv1.GetAccount: data_balance: %f, want: %f", rply, expBalance)
	}

	var replyActSess []*sessions.ExternalSession // find indexed Sy sessions active
	if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions, utils.SessionFilter{Filters: []string{"*string:~*req.RequestType:*sy"}}, &replyActSess); err != nil {
		t.Error(err)
	}
	if len(replyActSess) != 1 {
		t.Errorf("expected 1 active sessions, received <%v>", replyActSess)
	}
	// t.Log(utils.ToIJSON(replyActSess))

	// get thresholds
	var tIDs []string
	if err := client.Call(context.Background(), utils.ThresholdSv1GetThresholdIDs,
		&utils.TenantWithAPIOpts{}, &tIDs); err != nil {
		t.Error(err)
	}
	if len(tIDs) != 1 {
		t.Errorf("expected 1 threshold profile to be created, received <%v>", tIDs)
	} else if tIDs[0] != utils.MetaSy+utils.Underline+syOriginID {
		t.Errorf("expected <%v>, received <%v>", utils.MetaSy+utils.Underline+syOriginID, tIDs[0])
	}
	expTp := &engine.ThresholdProfile{
		Tenant: "cgrates.org",
		ID:     utils.MetaSy + utils.Underline + syOriginID,
		FilterIDs: []string{
			utils.ConcatenatedKey(utils.MetaLessOrEqual, utils.DynamicDataPrefix+utils.MetaAsm+utils.NestingSep+utils.BalanceSummaries+utils.NestingSep+"balance_data"+utils.NestingSep+utils.Value, "1023"),
			utils.ConcatenatedKey(utils.MetaString, utils.MetaDynReq+utils.NestingSep+utils.ID, "1001"),
		},
		MaxHits:   1,
		MinHits:   1,
		Async:     true,
		ActionIDs: []string{utils.MetaSyPublish},
	}
	tp := &engine.ThresholdProfile{}
	if err := client.Call(context.Background(), utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{ID: tIDs[0]}, &tp); err != nil {
		t.Error(err)
	}
	expTp.ActivationInterval = tp.ActivationInterval
	slices.Sort(tp.FilterIDs) // sort filters received since they are processed from a map
	if !reflect.DeepEqual(tp, expTp) {
		t.Errorf("expected <%v>\nreceived\n<%v>", utils.ToJSON(expTp), utils.ToJSON(tp))
	}
	// restart engine
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
	ng.PreserveDataDB = true
	client, cfg = ng.Run(t)
	time.Sleep(100 * time.Millisecond) // wait for DiameterAgent service to start

	diamClientRo, err := NewDiameterClient(cfg.DiameterAgentCfg().Listeners[0].Address, "localhost",
		cfg.DiameterAgentCfg().OriginRealm, cfg.DiameterAgentCfg().VendorID,
		cfg.DiameterAgentCfg().ProductName, utils.DiameterFirmwareRevision,
		cfg.DiameterAgentCfg().DictionariesPath, cfg.DiameterAgentCfg().Listeners[0].Network)
	if err != nil {
		t.Fatal(err)
	}
	ccr := diam.NewRequest(diam.CreditControl, 4, nil)
	ccrOriginID := utils.UUIDSha1Prefix()
	ccr.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(ccrOriginID))
	ccr.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	ccr.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	ccr.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	ccr.NewAVP(avp.DestinationRealm, avp.Mbit, 0, datatype.DiameterIdentity("dr-cgrates.org"))
	ccr.NewAVP(avp.ServiceContextID, avp.Mbit, 0, datatype.UTF8String("message@DiamItCCRData"))
	ccr.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(1))
	ccr.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(0))
	ccr.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA-DH"))
	ccr.NewAVP(avp.OriginStateID, avp.Mbit, 0, datatype.Unsigned32(time.Now().Unix()))
	ccr.NewAVP(avp.EventTimestamp, avp.Mbit, 0, datatype.Time(time.Date(2018, 10, 5, 11, 43, 10, 0, time.UTC)))
	ccr.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(0)),
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("1001")), // Subscription-Id-Data
		}})
	ccr.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(1)),
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("104502200011")), // Subscription-Id-Data
		}})
	ccr.NewAVP(avp.MultipleServicesIndicator, avp.Mbit, 0, datatype.Enumerated(1))
	ccr.NewAVP(avp.MultipleServicesCreditControl, avp.Mbit, 0, &diam.GroupedAVP{ // Multiple-Services-Credit-Control
		AVP: []*diam.AVP{
			diam.NewAVP(avp.RatingGroup, avp.Mbit, 0, datatype.Unsigned32(20000)), // Rating-Group
		},
	})
	ccr.NewAVP(avp.ServiceInformation, avp.Mbit, 10415,
		&diam.GroupedAVP{
			AVP: []*diam.AVP{
				diam.NewAVP(avp.PSInformation, avp.Mbit, 10415,
					&diam.GroupedAVP{
						AVP: []*diam.AVP{
							diam.NewAVP(avp.CalledStationID, avp.Mbit, 0, datatype.UTF8String("itsyscom")),
							diam.NewAVP(avp.TGPPSGSNMCCMNC, avp.Mbit, 10415, datatype.OctetString("1002")),
						},
					},
				),
			},
		},
	)

	// t.Log("sendingg CCR msg: ", ccr.PrettyDump())
	if err := diamClientRo.SendMessage(ccr); err != nil {
		t.Errorf("failed to send diameter message: %v", err)
	}

	reply = diamClientRo.ReceivedMessage(2 * time.Second)
	if reply == nil {
		t.Error("received empty reply")
	}
	// t.Log(reply.PrettyDump())

	avps, err = reply.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID)
	if err != nil {
		t.Error(err)
	}
	if len(avps) == 0 {
		t.Error("missing AVPs in reply")
	}

	resultCode, err = diamAVPAsString(avps[0])
	if err != nil {
		t.Error(err)
	}
	if resultCode != "2001" {
		t.Errorf("Result-Code=%s, want %s", resultCode, "2001")
	}

	expBalance = float64(1024) // CCR init should take 2048 units
	if err = client.Call(context.Background(), utils.APIerSv2GetAccount, attrsAcnt, &acnt); err != nil {
		t.Errorf("APIerSv1.GetAccount unexpected err: %v", err)
	} else if rply := acnt.BalanceMap[utils.MetaData].GetTotalValue(); rply != expBalance {
		t.Errorf("APIerSv1.GetAccount: data_balance: %f, want: %f", rply, expBalance)
	}
	// find indexed sy sessions
	if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions, utils.SessionFilter{Filters: []string{"*string:~*req.RequestType:*sy"}}, &replyActSess); err != nil {
		t.Fatal(err)
	}
	if len(replyActSess) != 1 {
		t.Errorf("expected 1 active sessions, received <%v>", replyActSess)
	}
	// t.Log(utils.ToIJSON(replyActSess))
	// find all active sessions
	if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions, utils.SessionFilter{}, &replyActSess); err != nil {
		t.Error(err)
	}
	if len(replyActSess) != 2 {
		t.Errorf("expected 2 active sessions, received <%v>", replyActSess)
	}
	// t.Log(utils.ToIJSON(replyActSess))

	ccrU := diam.NewRequest(diam.CreditControl, 4, nil)
	ccrU.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(ccrOriginID))
	ccrU.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	ccrU.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	ccrU.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	ccrU.NewAVP(avp.DestinationRealm, avp.Mbit, 0, datatype.DiameterIdentity("dr-cgrates.org"))
	ccrU.NewAVP(avp.ServiceContextID, avp.Mbit, 0, datatype.UTF8String("message@DiamItCCRData"))
	ccrU.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(2))
	ccrU.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(1))
	ccrU.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA-DH"))
	ccrU.NewAVP(avp.OriginStateID, avp.Mbit, 0, datatype.Unsigned32(time.Now().Unix()))
	ccrU.NewAVP(avp.EventTimestamp, avp.Mbit, 0, datatype.Time(time.Date(2018, 10, 5, 11, 47, 10, 0, time.UTC)))
	ccrU.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(0)),
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("1001")), // Subscription-Id-Data
		}})
	ccrU.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(1)),
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("104502200011")), // Subscription-Id-Data
		}})
	ccrU.NewAVP(avp.MultipleServicesIndicator, avp.Mbit, 0, datatype.Enumerated(1))
	ccrU.NewAVP(avp.MultipleServicesCreditControl, avp.Mbit, 0, &diam.GroupedAVP{ // Multiple-Services-Credit-Control
		AVP: []*diam.AVP{
			diam.NewAVP(avp.RatingGroup, avp.Mbit, 0, datatype.Unsigned32(20000)), // Rating-Group
			diam.NewAVP(avp.UsedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{ // Used-Service-Unit
				AVP: []*diam.AVP{
					diam.NewAVP(avp.CCTotalOctets, avp.Mbit, 0, datatype.Unsigned64(7640)),  // CC-Total-Octets
					diam.NewAVP(avp.CCInputOctets, avp.Mbit, 0, datatype.Unsigned64(5337)),  // CC-Input-Octets
					diam.NewAVP(avp.CCOutputOctets, avp.Mbit, 0, datatype.Unsigned64(2303)), // CC-Output-Octets
				},
			}),
		},
	})
	ccrU.NewAVP(avp.ServiceInformation, avp.Mbit, 10415,
		&diam.GroupedAVP{
			AVP: []*diam.AVP{
				diam.NewAVP(avp.PSInformation, avp.Mbit, 10415,
					&diam.GroupedAVP{
						AVP: []*diam.AVP{
							diam.NewAVP(avp.CalledStationID, avp.Mbit, 0, datatype.UTF8String("itsyscom")),
							diam.NewAVP(avp.TGPPSGSNMCCMNC, avp.Mbit, 10415, datatype.OctetString("1002")),
						},
					},
				),
			},
		},
	)

	// t.Log("sendingg CCR-U msg: ", ccrU.PrettyDump())
	if err := diamClientRo.SendMessage(ccrU); err != nil {
		t.Errorf("failed to send diameter message: %v", err)
	}

	reply = diamClientRo.ReceivedMessage(5 * time.Second)
	if reply == nil {
		t.Fatal("received empty reply")
	}
	// t.Log(reply.PrettyDump())

	avps, err = reply.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID)
	if err != nil {
		t.Error(err)
	}
	if len(avps) == 0 {
		t.Error("missing AVPs in reply")
	}

	resultCode, err = diamAVPAsString(avps[0])
	if err != nil {
		t.Error(err)
	}
	if resultCode != "2001" {
		t.Errorf("Result-Code=%s, want %s", resultCode, "2001")
	}

	expBalance = float64(0) // CCR update should take 2048 units
	if err = client.Call(context.Background(), utils.APIerSv2GetAccount, attrsAcnt, &acnt); err != nil {
		t.Errorf("APIerSv1.GetAccount unexpected err: %v", err)
	} else if rply := acnt.BalanceMap[utils.MetaData].GetTotalValue(); rply != expBalance {
		t.Errorf("APIerSv1.GetAccount: data_balance: %f, want: %f", rply, expBalance)
	}

	// find indexed sy sessions
	if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions, utils.SessionFilter{Filters: []string{"*string:~*req.RequestType:*sy"}}, &replyActSess); err != nil {
		t.Error(err)
	}
	if len(replyActSess) != 1 {
		t.Errorf("expected 1 active sessions, received <%v>", replyActSess)
	}
	if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions, utils.SessionFilter{}, &replyActSess); err != nil {
		t.Error(err)
	}
	if len(replyActSess) != 2 {
		t.Errorf("expected 2 active sessions, received <%v>", replyActSess)
	}
	// t.Log(utils.ToIJSON(replyActSess))

	// receive SNR
	// unfinished , SNR cant recover from engine shutdown
	// 	reply = diamClientSy.ReceivedMessage(2 * time.Second)
	// 	if reply == nil {
	// 		t.Fatal("Received empty reply")
	// 	} else {
	// 		expected := fmt.Sprintf(`Spending-Status-Notification-Request (SNR)
	// %s
	// 	Session-Id {Code:263,Flags:0x40,Length:16,VendorId:0,Value:UTF8String{%s},Padding:1}
	// 	Origin-Host {Code:264,Flags:0x40,Length:16,VendorId:0,Value:DiameterIdentity{CGR-DA},Padding:2}
	// 	Origin-Realm {Code:296,Flags:0x40,Length:20,VendorId:0,Value:DiameterIdentity{cgrates.org},Padding:1}
	// 	Destination-Realm {Code:283,Flags:0x40,Length:24,VendorId:0,Value:DiameterIdentity{dr-cgrates.org},Padding:2}
	// 	Destination-Host {Code:293,Flags:0x40,Length:20,VendorId:0,Value:DiameterIdentity{CGR-DA-DH},Padding:3}
	// 	Auth-Application-Id {Code:258,Flags:0x40,Length:12,VendorId:0,Value:Unsigned32{16777302}}
	// 	Policy-Counter-Status-Report {Code:2903,Flags:0xc0,Length:96,VendorId:10415,Value:Grouped{
	// 		Policy-Counter-Identifier {Code:2901,Flags:0xc0,Length:20,VendorId:10415,Value:UTF8String{Monthly},Padding:1},
	// 		Policy-Counter-Status {Code:2902,Flags:0xc0,Length:20,VendorId:10415,Value:UTF8String{512KBPS},Padding:1},
	// 		Pending-Policy-Counter-Information {Code:2905,Flags:0xc0,Length:44,VendorId:10415,Value:Grouped{
	// 			Policy-Counter-Status {Code:2902,Flags:0xc0,Length:16,VendorId:10415,Value:UTF8String{30GB},Padding:0},
	// 			Pending-Policy-Counter-Change-Time {Code:2906,Flags:0xc0,Length:16,VendorId:10415,Value:Time{%v}},
	// 		}}
	// 	}}
	// `, reply.Header, syOriginID, time.Time(reply.AVP[6].Data.(*diam.GroupedAVP).AVP[2].Data.(*diam.GroupedAVP).AVP[1].Data.(datatype.Time)))
	// 		if expected != reply.String() {
	// 			t.Errorf("expected \n<%v>, \nreceived \n<%v>", expected, reply.String())
	// 		}
	// 	}

	// Change session policy subscription
	slri := diam.NewRequest(diam.SpendingLimit, 16777302, nil)
	slri.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(syOriginID))
	slri.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	slri.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	slri.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA-DH"))
	slri.NewAVP(avp.DestinationRealm, avp.Mbit, 0, datatype.DiameterIdentity("dr-cgrates.org"))
	slri.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(16777302))
	slri.NewAVP(avp.SLRequestType, avp.Vbit, 10415, datatype.Enumerated(1)) // INTERMEDIATE_REQUEST (1)
	slri.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(0)),
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("1001")), // Subscription-Id-Data (MSISDN)
		}})
	slri.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(1)),
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("104502200011")), // Subscription-Id-Data (IMSI)
		}})
	// t.Log("sendingg msg: ", slri.PrettyDump())
	// reopen connection with client after we shutdown engine
	diamClientSy, err = NewDiameterClient(cfg.DiameterAgentCfg().Listeners[0].Address, "localhost",
		cfg.DiameterAgentCfg().OriginRealm, cfg.DiameterAgentCfg().VendorID,
		cfg.DiameterAgentCfg().ProductName, utils.DiameterFirmwareRevision,
		cfg.DiameterAgentCfg().DictionariesPath, cfg.DiameterAgentCfg().Listeners[0].Network)
	if err != nil {
		t.Fatal(err)
	}
	if err := diamClientSy.SendMessage(slri); err != nil {
		t.Errorf("failed to send diameter message: %v", err)
	}

	reply = diamClientSy.ReceivedMessage(2 * time.Second)
	if reply == nil {
		t.Fatal("received empty reply")
	}
	// t.Log(reply.PrettyDump())
	avps, err = reply.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID)
	if err != nil {
		t.Error(err)
	}
	if len(avps) == 0 {
		t.Fatal("missing AVPs in reply")
	}

	resultCode, err = diamAVPAsString(avps[0])
	if err != nil {
		t.Error(err)
	}
	if resultCode != "2001" {
		t.Errorf("Result-Code=%s, want 2001", resultCode)
	}

	expBalance = float64(0)
	if err = client.Call(context.Background(), utils.APIerSv2GetAccount, attrsAcnt, &acnt); err != nil {
		t.Errorf("APIerSv1.GetAccount unexpected err: %v", err)
	} else if rply := acnt.BalanceMap[utils.MetaData].GetTotalValue(); rply != expBalance {
		t.Errorf("APIerSv1.GetAccount: data_balance: %f, want: %f", rply, expBalance)
	}

	if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions, utils.SessionFilter{Filters: []string{"*string:~*req.RequestType:*sy"}}, &replyActSess); err != nil {
		t.Error(err)
	}
	if len(replyActSess) != 1 {
		t.Errorf("expected 1 active sessions, received <%v>", replyActSess)
	}
	// t.Log(utils.ToIJSON(replyActSess))

	// get thresholds
	if err := client.Call(context.Background(), utils.ThresholdSv1GetThresholdIDs,
		&utils.TenantWithAPIOpts{}, &tIDs); err != nil {
		t.Error(err)
	}
	if len(tIDs) != 1 {
		t.Errorf("expected 1 threshold profile to be created, received <%v>", tIDs)
	} else if tIDs[0] != utils.MetaSy+utils.Underline+syOriginID {
		t.Errorf("expected <%v>, received <%v>", utils.MetaSy+utils.Underline+syOriginID, tIDs[0])
	}
	expTp = &engine.ThresholdProfile{
		Tenant: "cgrates.org",
		ID:     utils.MetaSy + utils.Underline + syOriginID,
		FilterIDs: []string{
			utils.ConcatenatedKey(utils.MetaLessOrEqual, utils.DynamicDataPrefix+utils.MetaAsm+utils.NestingSep+utils.BalanceSummaries+utils.NestingSep+"balance_data"+utils.NestingSep+utils.Value, "-1"),
			utils.ConcatenatedKey(utils.MetaString, utils.MetaDynReq+utils.NestingSep+utils.ID, "1001"),
		},
		MaxHits:   1,
		MinHits:   1,
		Async:     true,
		ActionIDs: []string{utils.MetaSyPublish},
	}
	tp = &engine.ThresholdProfile{}
	if err := client.Call(context.Background(), utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{ID: tIDs[0]}, &tp); err != nil {
		t.Error(err)
	}
	expTp.ActivationInterval = tp.ActivationInterval
	slices.Sort(tp.FilterIDs) // sort filters received since they are processed from a map
	if !reflect.DeepEqual(tp, expTp) {
		t.Errorf("expected <%v>\nreceived\n<%v>", utils.ToJSON(expTp), utils.ToJSON(tp))
	}

	// Send SNA
	// unfinished
	// sna := diam.NewRequest(SSN, 16777302, nil)
	// sna.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(syOriginID))
	// sna.NewAVP(avp.ResultCode, avp.Mbit, 0, datatype.Unsigned32(2001))
	// sna.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	// sna.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	// // t.Log("sendingg msg: ", sna.PrettyDump())
	// if err := diamClientSy.SendMessage(sna); err != nil {
	// 	t.Errorf("failed to send diameter message: %v", err)
	// }

	ccrT := diam.NewRequest(diam.CreditControl, 4, nil)
	ccrT.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(ccrOriginID))
	ccrT.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	ccrT.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	ccrT.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	ccrT.NewAVP(avp.DestinationRealm, avp.Mbit, 0, datatype.DiameterIdentity("dr-cgrates.org"))
	ccrT.NewAVP(avp.ServiceContextID, avp.Mbit, 0, datatype.UTF8String("message@DiamItCCRData"))
	ccrT.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(3))
	ccrT.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(2))
	ccrT.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA-DH"))
	ccrT.NewAVP(avp.OriginStateID, avp.Mbit, 0, datatype.Unsigned32(time.Now().Unix()))
	ccrT.NewAVP(avp.EventTimestamp, avp.Mbit, 0, datatype.Time(time.Date(2018, 10, 5, 11, 50, 10, 0, time.UTC)))
	ccrT.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(0)),
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("1001")), // Subscription-Id-Data
		}})
	ccrT.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(1)),
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("104502200011")), // Subscription-Id-Data
		}})
	ccrT.NewAVP(avp.TerminationCause, avp.Mbit, 0, datatype.Enumerated(1))
	ccrT.NewAVP(avp.MultipleServicesIndicator, avp.Mbit, 0, datatype.Enumerated(1))
	ccrT.NewAVP(avp.MultipleServicesCreditControl, avp.Mbit, 0, &diam.GroupedAVP{ // Multiple-Services-Credit-Control
		AVP: []*diam.AVP{
			diam.NewAVP(avp.RatingGroup, avp.Mbit, 0, datatype.Unsigned32(20000)), // Rating-Group
			diam.NewAVP(avp.UsedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{ // Used-Service-Unit
				AVP: []*diam.AVP{
					diam.NewAVP(avp.CCTotalOctets, avp.Mbit, 0, datatype.Unsigned64(7640)),  // CC-Total-Octets
					diam.NewAVP(avp.CCInputOctets, avp.Mbit, 0, datatype.Unsigned64(5337)),  // CC-Input-Octets
					diam.NewAVP(avp.CCOutputOctets, avp.Mbit, 0, datatype.Unsigned64(2303)), // CC-Output-Octets
				},
			}),
		},
	})
	ccrT.NewAVP(avp.ServiceInformation, avp.Mbit, 10415,
		&diam.GroupedAVP{
			AVP: []*diam.AVP{
				diam.NewAVP(avp.PSInformation, avp.Mbit, 10415,
					&diam.GroupedAVP{
						AVP: []*diam.AVP{
							diam.NewAVP(avp.CalledStationID, avp.Mbit, 0, datatype.UTF8String("itsyscom")),
							diam.NewAVP(avp.TGPPSGSNMCCMNC, avp.Mbit, 10415, datatype.OctetString("1002")),
						},
					},
				),
			},
		},
	)

	// t.Log("sendingg CCR-T msg: ", ccrT.PrettyDump())
	if err := diamClientRo.SendMessage(ccrT); err != nil {
		t.Errorf("failed to send diameter message: %v", err)
	}

	reply = diamClientRo.ReceivedMessage(2 * time.Second)
	if reply == nil {
		t.Fatal("received empty reply")
	}
	// t.Log(reply.PrettyDump())

	avps, err = reply.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID)
	if err != nil {
		t.Error(err)
	}
	if len(avps) == 0 {
		t.Error("missing AVPs in reply")
	}

	resultCode, err = diamAVPAsString(avps[0])
	if err != nil {
		t.Error(err)
	}
	if resultCode != "2001" {
		t.Errorf("Result-Code=%s, want %s", resultCode, "2001")
	}

	expBalance = float64(0) // CCR update should take 2048 units
	if err = client.Call(context.Background(), utils.APIerSv2GetAccount, attrsAcnt, &acnt); err != nil {
		t.Errorf("APIerSv1.GetAccount unexpected err: %v", err)
	} else if rply := acnt.BalanceMap[utils.MetaData].GetTotalValue(); rply != expBalance {
		t.Errorf("APIerSv1.GetAccount: data_balance: %f, want: %f", rply, expBalance)
	}

	// find indexed sy sessions
	if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions, utils.SessionFilter{Filters: []string{"*string:~*req.RequestType:*sy"}}, &replyActSess); err != nil {
		t.Error(err)
	}
	if len(replyActSess) != 1 {
		t.Errorf("expected 1 active sessions, received <%v>", replyActSess)
	}
	if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions, utils.SessionFilter{}, &replyActSess); err != nil {
		t.Error(err)
	}
	if len(replyActSess) != 1 {
		t.Errorf("expected 1 active sessions, received <%v>", replyActSess)
	}
	// t.Log(utils.ToIJSON(replyActSess))
	str := diam.NewRequest(diam.SessionTermination, 16777302, nil)
	str.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(syOriginID))
	str.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	str.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	str.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA-DH"))
	str.NewAVP(avp.DestinationRealm, avp.Mbit, 0, datatype.DiameterIdentity("dr-cgrates.org"))
	str.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(16777302))
	str.NewAVP(avp.TerminationCause, avp.Mbit, 0, datatype.Enumerated(1))
	// t.Log("sendingg STR msg: ", str.PrettyDump())
	if err := diamClientSy.SendMessage(str); err != nil {
		t.Errorf("failed to send diameter message: %v", err)
	}

	reply = diamClientSy.ReceivedMessage(2 * time.Second)
	if reply == nil {
		t.Fatal("received empty reply")
	}
	// t.Log(reply.PrettyDump())
	avps, err = reply.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID)
	if err != nil {
		t.Error(err)
	}
	if len(avps) == 0 {
		t.Fatal("missing AVPs in reply")
	}

	resultCode, err = diamAVPAsString(avps[0])
	if err != nil {
		t.Error(err)
	}
	if resultCode != "2001" {
		t.Errorf("Result-Code=%s, want 2001", resultCode)
	}

	expBalance = float64(0)
	if err = client.Call(context.Background(), utils.APIerSv2GetAccount, attrsAcnt, &acnt); err != nil {
		t.Errorf("APIerSv1.GetAccount unexpected err: %v", err)
	} else if rply := acnt.BalanceMap[utils.MetaData].GetTotalValue(); rply != expBalance {
		t.Errorf("APIerSv1.GetAccount: data_balance: %f, want: %f", rply, expBalance)
	}
	if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions, utils.SessionFilter{}, &replyActSess); err == nil || err.Error() != "NOT_FOUND" {
		t.Errorf("expected error <NOT_FOUND>, received <%v>", err)
	}
	// thresholds profile should be removed on STR
	if err := client.Call(context.Background(), utils.ThresholdSv1GetThresholdIDs,
		&utils.TenantWithAPIOpts{}, &tIDs); err != nil {
		t.Error(err)
	} else if len(tIDs) != 0 {
		t.Errorf("expected no Threshold profiles, received <%v>", tIDs)
	}
}

func TestDiamSyDefaultSLR(t *testing.T) {
	ng := engine.TestEngine{
		TpFiles: map[string]string{
			utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,PACKAGE_1001,,,`,
			utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
PACKAGE_1001,ACT_TOPUP,*asap,10`,
			utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP,*topup_reset,,,balance_data,*data,,,,,*unlimited,,3072,,,,`,
			utils.ChargersCsv: `#Tenant,ID,FilterIDs,ActivationInterval,RunID,AttributeIDs,Weight
cgrates.org,DEFAULT,,,*default,*none,0`,
			utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_DATA,*any,RT_DATA,*up,4,0,`,
			utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_DATA,0,0.01,1,1,0`,
			utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,data,*any,,RP_1001,`,
			utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_1001,DR_DATA,*any,10`,
		},
		// LogBuffer: &bytes.Buffer{},
	}
	switch *utils.DBType {
	case utils.MetaInternal:
		ng.ConfigPath = filepath.Join(*utils.DataDir, "conf", "samples", "diam_sy_default")
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}
	// t.Cleanup(func() { fmt.Println(ng.LogBuffer) })
	client, cfg := ng.Run(t)

	time.Sleep(100 * time.Millisecond) // wait for DiameterAgent service to start

	// Start monitoring SL
	diamClientSy, err := NewDiameterClient(cfg.DiameterAgentCfg().Listeners[0].Address, "localhost",
		cfg.DiameterAgentCfg().OriginRealm, cfg.DiameterAgentCfg().VendorID,
		cfg.DiameterAgentCfg().ProductName, utils.DiameterFirmwareRevision,
		cfg.DiameterAgentCfg().DictionariesPath, cfg.DiameterAgentCfg().Listeners[0].Network)
	if err != nil {
		t.Fatal(err)
	}
	syOriginID := utils.UUIDSha1Prefix()
	slr := diam.NewRequest(diam.SpendingLimit, 16777302, nil)
	slr.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(syOriginID))
	slr.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	slr.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	slr.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA-DH"))
	slr.NewAVP(avp.DestinationRealm, avp.Mbit, 0, datatype.DiameterIdentity("dr-cgrates.org"))
	slr.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(16777302))
	slr.NewAVP(avp.SLRequestType, avp.Vbit, 10415, datatype.Enumerated(0)) //INITIAL_REQUEST (0)
	slr.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(0)),
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("1001")), // Subscription-Id-Data (MSISDN)
		}})
	slr.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(1)),
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("104502200011")), // Subscription-Id-Data (IMSI)
		}})
	// t.Log("sendingg msg: ", slr.PrettyDump())
	if err := diamClientSy.SendMessage(slr); err != nil {
		t.Errorf("failed to send diameter message: %v", err)
	}

	reply := diamClientSy.ReceivedMessage(2 * time.Second)
	if reply == nil {
		t.Fatal("received empty reply")
	}
	// t.Log(reply.PrettyDump())
	avps, err := reply.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID)
	if err != nil {
		t.Error(err)
	}
	if len(avps) == 0 {
		t.Fatal("missing AVPs in reply")
	}

	resultCode, err := diamAVPAsString(avps[0])
	if err != nil {
		t.Error(err)
	}
	if resultCode != "2001" {
		t.Errorf("Result-Code=%s, want 2001", resultCode)
	}

	var acnt *engine.Account
	attrsAcnt := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1001",
	}
	if err = client.Call(context.Background(), utils.APIerSv2GetAccount, attrsAcnt, &acnt); err != nil {
		t.Errorf("APIerSv1.GetAccount unexpected err: %v", err)
	} else if rply := acnt.BalanceMap[utils.MetaData].GetTotalValue(); rply != 3072 {
		t.Errorf("APIerSv1.GetAccount: *default: %v, want: %v", rply, 3072)
	}

	var replyActSess []*sessions.ExternalSession // find indexed Sy sessions active
	if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions, utils.SessionFilter{Filters: []string{"*string:~*req.RequestType:*sy"}}, &replyActSess); err != nil {
		t.Error(err)
	}
	if len(replyActSess) != 1 {
		t.Errorf("expected 1 active sessions, received <%v>", replyActSess)
	}
	// t.Log(utils.ToIJSON(replyActSess))

	// get thresholds
	var tIDs []string
	if err := client.Call(context.Background(), utils.ThresholdSv1GetThresholdIDs,
		&utils.TenantWithAPIOpts{}, &tIDs); err != nil {
		t.Error(err)
	}
	if len(tIDs) != 1 {
		t.Errorf("expected 1 threshold profile to be created, received <%v>", tIDs)
	} else if tIDs[0] != utils.MetaSy+utils.Underline+syOriginID {
		t.Errorf("expected <%v>, received <%v>", utils.MetaSy+utils.Underline+syOriginID, tIDs[0])
	}
	expTp := &engine.ThresholdProfile{
		Tenant: "cgrates.org",
		ID:     utils.MetaSy + utils.Underline + syOriginID,
		FilterIDs: []string{
			"*string:~*asm.BalanceSummaries.*default.ID:balance_data",
			"*string:~*req.ID:1001",
		},
		MaxHits:   1,
		MinHits:   1,
		Async:     true,
		ActionIDs: []string{utils.MetaSyPublish},
	}
	tp := &engine.ThresholdProfile{}
	if err := client.Call(context.Background(), utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{ID: tIDs[0]}, &tp); err != nil {
		t.Error(err)
	}
	expTp.ActivationInterval = tp.ActivationInterval
	slices.Sort(tp.FilterIDs) // sort filters received since they are processed from a map
	if !reflect.DeepEqual(tp, expTp) {
		t.Errorf("expected <%v>\nreceived\n<%v>", utils.ToJSON(expTp), utils.ToJSON(tp))
	}

	// t.Log(utils.ToIJSON(replyActSess))
	str := diam.NewRequest(diam.SessionTermination, 16777302, nil)
	str.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(syOriginID))
	str.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	str.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	str.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA-DH"))
	str.NewAVP(avp.DestinationRealm, avp.Mbit, 0, datatype.DiameterIdentity("dr-cgrates.org"))
	str.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(16777302))
	str.NewAVP(avp.TerminationCause, avp.Mbit, 0, datatype.Enumerated(1))
	// t.Log("sendingg STR msg: ", str.PrettyDump())
	if err := diamClientSy.SendMessage(str); err != nil {
		t.Errorf("failed to send diameter message: %v", err)
	}

	reply = diamClientSy.ReceivedMessage(2 * time.Second)
	if reply == nil {
		t.Fatal("received empty reply")
	}
	// t.Log(reply.PrettyDump())
	avps, err = reply.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID)
	if err != nil {
		t.Error(err)
	}
	if len(avps) == 0 {
		t.Fatal("missing AVPs in reply")
	}

	resultCode, err = diamAVPAsString(avps[0])
	if err != nil {
		t.Error(err)
	}
	if resultCode != "2001" {
		t.Errorf("Result-Code=%s, want 2001", resultCode)
	}

	expBalance := float64(3072)
	if err = client.Call(context.Background(), utils.APIerSv2GetAccount, attrsAcnt, &acnt); err != nil {
		t.Errorf("APIerSv1.GetAccount unexpected err: %v", err)
	} else if rply := acnt.BalanceMap[utils.MetaData].GetTotalValue(); rply != expBalance {
		t.Errorf("APIerSv1.GetAccount: data_balance: %f, want: %f", rply, expBalance)
	}
	if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions, utils.SessionFilter{}, &replyActSess); err == nil || err.Error() != "NOT_FOUND" {
		t.Errorf("expected error <NOT_FOUND>, received <%v>", err)
	}
	// thresholds profile should be removed on STR
	if err := client.Call(context.Background(), utils.ThresholdSv1GetThresholdIDs,
		&utils.TenantWithAPIOpts{}, &tIDs); err != nil {
		t.Error(err)
	} else if len(tIDs) != 0 {
		t.Errorf("expected no Threshold profiles, received <%v>", tIDs)
	}
}

func TestDiamSyEdgeCases(t *testing.T) {
	ng := engine.TestEngine{
		TpFiles: map[string]string{
			utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,PACKAGE_1001,,,`,
			utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
PACKAGE_1001,ACT_TOPUP,*asap,10`,
			utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP,*topup_reset,,,balance_data,*data,,,,,*unlimited,,3072,,,,`,
			utils.ChargersCsv: `#Tenant,ID,FilterIDs,ActivationInterval,RunID,AttributeIDs,Weight
cgrates.org,DEFAULT,,,*default,*none,0`,
			utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_DATA,*any,RT_DATA,*up,4,0,`,
			utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_DATA,0,0.01,1,1,0`,
			utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,data,itsyscom,,RP_1001,`,
			utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_1001,DR_DATA,*any,10`,
		},
		// LogBuffer: &bytes.Buffer{},
	}
	switch *utils.DBType {
	case utils.MetaInternal:
		ng.ConfigPath = filepath.Join(*utils.DataDir, "conf", "samples", "diam_sy_edge_cases")
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	t.Run("Test DIAMETER_INVALID_AVP_VALUE", func(t *testing.T) {
		// t.Cleanup(func() { fmt.Println(ng.LogBuffer) })
		client, cfg := ng.Run(t)

		time.Sleep(100 * time.Millisecond) // wait for DiameterAgent service to start
		// Start start proper SL session
		diamClientSy, err := NewDiameterClient(cfg.DiameterAgentCfg().Listeners[0].Address, "localhost",
			cfg.DiameterAgentCfg().OriginRealm, cfg.DiameterAgentCfg().VendorID,
			cfg.DiameterAgentCfg().ProductName, utils.DiameterFirmwareRevision,
			cfg.DiameterAgentCfg().DictionariesPath, cfg.DiameterAgentCfg().Listeners[0].Network)
		if err != nil {
			t.Fatal(err)
		}
		syOriginID := utils.UUIDSha1Prefix()
		slr := diam.NewRequest(diam.SpendingLimit, 16777302, nil)
		slr.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(syOriginID))
		slr.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
		slr.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
		slr.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA-DH"))
		slr.NewAVP(avp.DestinationRealm, avp.Mbit, 0, datatype.DiameterIdentity("dr-cgrates.org"))
		slr.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(16777302))
		slr.NewAVP(avp.SLRequestType, avp.Vbit, 10415, datatype.Enumerated(0)) //INITIAL_REQUEST (0)
		slr.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
			AVP: []*diam.AVP{
				diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(0)),
				diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("1001")), // Subscription-Id-Data (MSISDN)
			}})
		slr.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
			AVP: []*diam.AVP{
				diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(1)),
				diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("104502200011")), // Subscription-Id-Data (IMSI)
			}})
		// t.Log("sendingg msg: ", slr.PrettyDump())
		if err := diamClientSy.SendMessage(slr); err != nil {
			t.Errorf("failed to send diameter message: %v", err)
		}

		reply := diamClientSy.ReceivedMessage(2 * time.Second)
		if reply == nil {
			t.Fatal("received empty reply")
		}
		// t.Log(reply.PrettyDump())
		avps, err := reply.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID)
		if err != nil {
			t.Error(err)
		}
		if len(avps) == 0 {
			t.Fatal("missing AVPs in reply")
		}

		resultCode, err := diamAVPAsString(avps[0])
		if err != nil {
			t.Error(err)
		}
		if resultCode != "2001" {
			t.Errorf("Result-Code=%s, want 2001", resultCode)
		}
		expBalance := float64(3072)
		var acnt *engine.Account
		attrsAcnt := &utils.AttrGetAccount{
			Tenant:  "cgrates.org",
			Account: "1001",
		}
		if err = client.Call(context.Background(), utils.APIerSv2GetAccount, attrsAcnt, &acnt); err != nil {
			t.Errorf("APIerSv1.GetAccount unexpected err: %v", err)
		} else if rply := acnt.BalanceMap[utils.MetaData].GetTotalValue(); rply != expBalance {
			t.Errorf("APIerSv1.GetAccount: data_balance: %f, want: %f", rply, expBalance)
		}

		var replyActSess []*sessions.ExternalSession // find indexed Sy sessions active
		if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions, utils.SessionFilter{Filters: []string{"*string:~*req.RequestType:*sy"}}, &replyActSess); err != nil {
			t.Error(err)
		}
		if len(replyActSess) != 1 {
			t.Errorf("expected 1 active sessions, received <%v>", replyActSess)
		}
		// t.Log(utils.ToIJSON(replyActSess))

		// get thresholds
		var tIDs []string
		if err := client.Call(context.Background(), utils.ThresholdSv1GetThresholdIDs,
			&utils.TenantWithAPIOpts{}, &tIDs); err != nil {
			t.Error(err)
		}
		if len(tIDs) != 1 {
			t.Errorf("expected 1 threshold profile to be created, received <%v>", tIDs)
		} else if tIDs[0] != utils.MetaSy+utils.Underline+syOriginID {
			t.Errorf("expected <%v>, received <%v>", utils.MetaSy+utils.Underline+syOriginID, tIDs[0])
		}
		expTp := &engine.ThresholdProfile{
			Tenant: "cgrates.org",
			ID:     utils.MetaSy + utils.Underline + syOriginID,
			FilterIDs: []string{
				utils.ConcatenatedKey(utils.MetaLessOrEqual, utils.DynamicDataPrefix+utils.MetaAsm+utils.NestingSep+utils.BalanceSummaries+utils.NestingSep+"balance_data"+utils.NestingSep+utils.Value, "1023"),
				utils.ConcatenatedKey(utils.MetaString, utils.MetaDynReq+utils.NestingSep+utils.ID, "1001"),
			},
			MaxHits:   1,
			MinHits:   1,
			Async:     true,
			ActionIDs: []string{utils.MetaSyPublish},
		}
		tp := &engine.ThresholdProfile{}
		if err := client.Call(context.Background(), utils.APIerSv1GetThresholdProfile,
			&utils.TenantID{ID: tIDs[0]}, &tp); err != nil {
			t.Error(err)
		}
		expTp.ActivationInterval = tp.ActivationInterval
		slices.Sort(tp.FilterIDs) // sort filters received since they are processed from a map
		if !reflect.DeepEqual(tp, expTp) {
			t.Errorf("expected <%v>\nreceived\n<%v>", utils.ToJSON(expTp), utils.ToJSON(tp))
		}

		// start sessions with the same OriginID to get expected error
		// t.Log("sendingg msg: ", slr.PrettyDump())
		if err := diamClientSy.SendMessage(slr); err != nil {
			t.Errorf("failed to send diameter message: %v", err)
		}

		reply = diamClientSy.ReceivedMessage(2 * time.Second)
		if reply == nil {
			t.Fatal("received empty reply")
		}
		// t.Log(reply.PrettyDump())
		avps, err = reply.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID)
		if err != nil {
			t.Error(err)
		}
		if len(avps) == 0 {
			t.Fatal("missing AVPs in reply")
		}

		resultCode, err = diamAVPAsString(avps[0])
		if err != nil {
			t.Error(err)
		}
		if resultCode != "5004" {
			t.Errorf("Result-Code=%s, want 5004", resultCode)
		}
		avps, err = reply.FindAVPsWithPath([]any{"Failed-AVP"}, dict.UndefinedVendorID)
		if err != nil {
			t.Error(err)
		}
		if len(avps) == 0 {
			t.Fatal("missing AVPs in reply")
		}

		resultCode, err = diamAVPAsString(avps[0])
		if err != nil {
			t.Error(err)
		}
		expRC := `{"AVP":[{"Code":2904,"Flags":192,"Length":16,"VendorID":10415,"Data":0}]}`
		if resultCode != expRC {
			t.Errorf("Failed-AVP=<%s>, want <%s>", resultCode, expRC)
		}
		if err = client.Call(context.Background(), utils.APIerSv2GetAccount, attrsAcnt, &acnt); err != nil {
			t.Errorf("APIerSv1.GetAccount unexpected err: %v", err)
		} else if rply := acnt.BalanceMap[utils.MetaData].GetTotalValue(); rply != expBalance {
			t.Errorf("APIerSv1.GetAccount: data_balance: %f, want: %f", rply, expBalance)
		}

		// previous session should still be alove and no new sessions should be created
		if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions, utils.SessionFilter{Filters: []string{"*string:~*req.RequestType:*sy"}}, &replyActSess); err != nil {
			t.Error(err)
		}
		if len(replyActSess) != 1 {
			t.Errorf("expected 1 active sessions, received <%v>", replyActSess)
		}
		// t.Log(utils.ToIJSON(replyActSess))

		// get thresholds, no new thresholds shouldve been created
		if err := client.Call(context.Background(), utils.ThresholdSv1GetThresholdIDs,
			&utils.TenantWithAPIOpts{}, &tIDs); err != nil {
			t.Error(err)
		}
		if len(tIDs) != 1 {
			t.Errorf("expected 1 threshold profile to be created, received <%v>", tIDs)
		} else if tIDs[0] != utils.MetaSy+utils.Underline+syOriginID {
			t.Errorf("expected <%v>, received <%v>", utils.MetaSy+utils.Underline+syOriginID, tIDs[0])
		}
		if err := client.Call(context.Background(), utils.APIerSv1GetThresholdProfile,
			&utils.TenantID{ID: tIDs[0]}, &tp); err != nil {
			t.Error(err)
		}
		expTp.ActivationInterval = tp.ActivationInterval
		slices.Sort(tp.FilterIDs) // sort filters received since they are processed from a map
		if !reflect.DeepEqual(tp, expTp) {
			t.Errorf("expected <%v>\nreceived\n<%v>", utils.ToJSON(expTp), utils.ToJSON(tp))
		}
		if err := engine.KillEngine(100); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Test DIAMETER_UNKNOWN_SESSION_ID SLR", func(t *testing.T) {
		// t.Cleanup(func() { fmt.Println(ng.LogBuffer) })
		client, cfg := ng.Run(t)

		time.Sleep(100 * time.Millisecond) // wait for DiameterAgent service to start
		// Start SL intermediate session that will error with DIAMETER_UNKNOWN_SESSION_ID
		diamClientSy, err := NewDiameterClient(cfg.DiameterAgentCfg().Listeners[0].Address, "localhost",
			cfg.DiameterAgentCfg().OriginRealm, cfg.DiameterAgentCfg().VendorID,
			cfg.DiameterAgentCfg().ProductName, utils.DiameterFirmwareRevision,
			cfg.DiameterAgentCfg().DictionariesPath, cfg.DiameterAgentCfg().Listeners[0].Network)
		if err != nil {
			t.Fatal(err)
		}
		syOriginID := utils.UUIDSha1Prefix()
		slri := diam.NewRequest(diam.SpendingLimit, 16777302, nil)
		slri.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(syOriginID))
		slri.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
		slri.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
		slri.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA-DH"))
		slri.NewAVP(avp.DestinationRealm, avp.Mbit, 0, datatype.DiameterIdentity("dr-cgrates.org"))
		slri.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(16777302))
		slri.NewAVP(avp.SLRequestType, avp.Vbit, 10415, datatype.Enumerated(1)) // INTERMEDIATE_REQUEST (1)
		slri.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
			AVP: []*diam.AVP{
				diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(0)),
				diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("1001")), // Subscription-Id-Data (MSISDN)
			}})
		slri.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
			AVP: []*diam.AVP{
				diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(1)),
				diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("104502200011")), // Subscription-Id-Data (IMSI)
			}})
		// t.Log("sendingg msg: ", slri.PrettyDump())
		if err := diamClientSy.SendMessage(slri); err != nil {
			t.Errorf("failed to send diameter message: %v", err)
		}

		reply := diamClientSy.ReceivedMessage(2 * time.Second)
		if reply == nil {
			t.Fatal("received empty reply")
		}
		// t.Log(reply.PrettyDump())
		avps, err := reply.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID)
		if err != nil {
			t.Error(err)
		}
		if len(avps) == 0 {
			t.Fatal("missing AVPs in reply")
		}

		resultCode, err := diamAVPAsString(avps[0])
		if err != nil {
			t.Error(err)
		}
		if resultCode != "5002" {
			t.Errorf("Result-Code=%s, want 5002", resultCode)
		}
		expBalance := float64(3072)
		var acnt *engine.Account
		attrsAcnt := &utils.AttrGetAccount{
			Tenant:  "cgrates.org",
			Account: "1001",
		}
		if err = client.Call(context.Background(), utils.APIerSv2GetAccount, attrsAcnt, &acnt); err != nil {
			t.Errorf("APIerSv1.GetAccount unexpected err: %v", err)
		} else if rply := acnt.BalanceMap[utils.MetaData].GetTotalValue(); rply != expBalance {
			t.Errorf("APIerSv1.GetAccount: data_balance: %f, want: %f", rply, expBalance)
		}

		var replyActSess []*sessions.ExternalSession
		if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions, utils.SessionFilter{}, &replyActSess); err == nil && err != utils.ErrNotFound {
			t.Error(err)
		}
		if len(replyActSess) != 0 {
			t.Errorf("expected 0 active sessions, received <%v>", replyActSess)
		}
		// t.Log(utils.ToIJSON(replyActSess))

		// get thresholds
		var tIDs []string
		if err := client.Call(context.Background(), utils.ThresholdSv1GetThresholdIDs,
			&utils.TenantWithAPIOpts{}, &tIDs); err != nil {
			t.Error(err)
		}
		if len(tIDs) != 0 {
			t.Errorf("expected 0 threshold profile to be created, received <%v>", tIDs)
		}

		if err := engine.KillEngine(100); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Test DIAMETER_UNKNOWN_SESSION_ID STR", func(t *testing.T) {
		// t.Cleanup(func() { fmt.Println(ng.LogBuffer) })
		client, cfg := ng.Run(t)

		time.Sleep(100 * time.Millisecond) // wait for DiameterAgent service to start
		// Start SL intermediate session that will error with DIAMETER_UNKNOWN_SESSION_ID
		diamClientSy, err := NewDiameterClient(cfg.DiameterAgentCfg().Listeners[0].Address, "localhost",
			cfg.DiameterAgentCfg().OriginRealm, cfg.DiameterAgentCfg().VendorID,
			cfg.DiameterAgentCfg().ProductName, utils.DiameterFirmwareRevision,
			cfg.DiameterAgentCfg().DictionariesPath, cfg.DiameterAgentCfg().Listeners[0].Network)
		if err != nil {
			t.Fatal(err)
		}
		syOriginID := utils.UUIDSha1Prefix()
		str := diam.NewRequest(diam.SessionTermination, 16777302, nil)
		str.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(syOriginID))
		str.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
		str.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
		str.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA-DH"))
		str.NewAVP(avp.DestinationRealm, avp.Mbit, 0, datatype.DiameterIdentity("dr-cgrates.org"))
		str.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(16777302))
		str.NewAVP(avp.TerminationCause, avp.Mbit, 0, datatype.Enumerated(1))
		// t.Log("sendingg STR msg: ", str.PrettyDump())
		if err := diamClientSy.SendMessage(str); err != nil {
			t.Errorf("failed to send diameter message: %v", err)
		}

		reply := diamClientSy.ReceivedMessage(2 * time.Second)
		if reply == nil {
			t.Fatal("received empty reply")
		}
		// t.Log(reply.PrettyDump())
		avps, err := reply.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID)
		if err != nil {
			t.Error(err)
		}
		if len(avps) == 0 {
			t.Fatal("missing AVPs in reply")
		}

		resultCode, err := diamAVPAsString(avps[0])
		if err != nil {
			t.Error(err)
		}
		if resultCode != "5002" {
			t.Errorf("Result-Code=%s, want 5002", resultCode)
		}

		expBalance := float64(3072)
		var acnt *engine.Account
		attrsAcnt := &utils.AttrGetAccount{
			Tenant:  "cgrates.org",
			Account: "1001",
		}
		if err = client.Call(context.Background(), utils.APIerSv2GetAccount, attrsAcnt, &acnt); err != nil {
			t.Errorf("APIerSv1.GetAccount unexpected err: %v", err)
		} else if rply := acnt.BalanceMap[utils.MetaData].GetTotalValue(); rply != expBalance {
			t.Errorf("APIerSv1.GetAccount: data_balance: %f, want: %f", rply, expBalance)
		}
		var replyActSess []*sessions.ExternalSession
		if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions, utils.SessionFilter{}, &replyActSess); err == nil || err.Error() != "NOT_FOUND" {
			t.Errorf("expected error <NOT_FOUND>, received <%v>", err)
		}
		var tIDs []string
		if err := client.Call(context.Background(), utils.ThresholdSv1GetThresholdIDs,
			&utils.TenantWithAPIOpts{}, &tIDs); err != nil {
			t.Error(err)
		} else if len(tIDs) != 0 {
			t.Errorf("expected no Threshold profiles, received <%v>", tIDs)
		}

		if err := engine.KillEngine(100); err != nil {
			t.Fatal(err)
		}
	})

	// unfinished , maybe its not necessary to comply with standards for account checking on SLR step
	// t.Run("Test DIAMETER_USER_UNKNOWN", func(t *testing.T) {
	// 	// t.Cleanup(func() { fmt.Println(ng.LogBuffer) })
	// 	client, cfg := ng.Run(t)

	// 	time.Sleep(100 * time.Millisecond) // wait for DiameterAgent service to start
	// 	// Start start proper SL session
	// 	diamClientSy, err := NewDiameterClient(cfg.DiameterAgentCfg().Listeners[0].Address, "localhost",
	// 		cfg.DiameterAgentCfg().OriginRealm, cfg.DiameterAgentCfg().VendorID,
	// 		cfg.DiameterAgentCfg().ProductName, utils.DiameterFirmwareRevision,
	// 		cfg.DiameterAgentCfg().DictionariesPath, cfg.DiameterAgentCfg().Listeners[0].Network)
	// 	if err != nil {
	// 		t.Fatal(err)
	// 	}
	// 	syOriginID := utils.UUIDSha1Prefix()
	// 	slr := diam.NewRequest(diam.SpendingLimit, 16777302, nil)
	// 	slr.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(syOriginID))
	// 	slr.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	// 	slr.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	// 	slr.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA-DH"))
	// 	slr.NewAVP(avp.DestinationRealm, avp.Mbit, 0, datatype.DiameterIdentity("dr-cgrates.org"))
	// 	slr.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(16777302))
	// 	slr.NewAVP(avp.SLRequestType, avp.Vbit, 10415, datatype.Enumerated(0)) //INITIAL_REQUEST (0)
	// 	slr.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
	// 		AVP: []*diam.AVP{
	// 			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(0)),
	// 			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("NonExistingAccount")), // Subscription-Id-Data (MSISDN)
	// 		}})
	// 	slr.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
	// 		AVP: []*diam.AVP{
	// 			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(1)),
	// 			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("104502200011")), // Subscription-Id-Data (IMSI)
	// 		}})
	// 	// t.Log("sendingg msg: ", slr.PrettyDump())
	// 	if err := diamClientSy.SendMessage(slr); err != nil {
	// 		t.Errorf("failed to send diameter message: %v", err)
	// 	}

	// 	reply := diamClientSy.ReceivedMessage(2 * time.Second)
	// 	if reply == nil {
	// 		t.Fatal("received empty reply")
	// 	}
	// 	// t.Log(reply.PrettyDump())
	// 	avps, err := reply.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID)
	// 	if err != nil {
	// 		t.Error(err)
	// 	}
	// 	if len(avps) == 0 {
	// 		t.Fatal("missing AVPs in reply")
	// 	}

	// 	resultCode, err := diamAVPAsString(avps[0])
	// 	if err != nil {
	// 		t.Error(err)
	// 	}
	// 	if resultCode != "5030" {
	// 		t.Errorf("Result-Code=%s, want 5030", resultCode)
	// 	}

	// 	var replyActSess []*sessions.ExternalSession
	// 	if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions, utils.SessionFilter{}, &replyActSess); err == nil || err.Error() != "NOT_FOUND" {
	// 		t.Errorf("expected error <NOT_FOUND>, received <%v>", err)
	// 	}
	// 	var tIDs []string
	// 	if err := client.Call(context.Background(), utils.ThresholdSv1GetThresholdIDs,
	// 		&utils.TenantWithAPIOpts{}, &tIDs); err != nil {
	// 		t.Error(err)
	// 	} else if len(tIDs) != 0 {
	// 		t.Errorf("expected no Threshold profiles, received <%v>", tIDs)
	// 	}

	// 	if err := engine.KillEngine(100); err != nil {
	// 		t.Fatal(err)
	// 	}
	// })

	t.Run("Test DIAMETER_ERROR_UNKNOWN_POLICY_COUNTERS SLR", func(t *testing.T) {
		// t.Cleanup(func() { fmt.Println(ng.LogBuffer) })
		client, cfg := ng.Run(t)

		time.Sleep(100 * time.Millisecond) // wait for DiameterAgent service to start
		// Start start proper SL session
		diamClientSy, err := NewDiameterClient(cfg.DiameterAgentCfg().Listeners[0].Address, "localhost",
			cfg.DiameterAgentCfg().OriginRealm, cfg.DiameterAgentCfg().VendorID,
			cfg.DiameterAgentCfg().ProductName, utils.DiameterFirmwareRevision,
			cfg.DiameterAgentCfg().DictionariesPath, cfg.DiameterAgentCfg().Listeners[0].Network)
		if err != nil {
			t.Fatal(err)
		}
		slr := diam.NewRequest(diam.SpendingLimit, 16777302, nil)
		slr.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String("BadPolicyFilter"))
		slr.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
		slr.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
		slr.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA-DH"))
		slr.NewAVP(avp.DestinationRealm, avp.Mbit, 0, datatype.DiameterIdentity("dr-cgrates.org"))
		slr.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(16777302))
		slr.NewAVP(avp.SLRequestType, avp.Vbit, 10415, datatype.Enumerated(0)) //INITIAL_REQUEST (0)
		slr.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
			AVP: []*diam.AVP{
				diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(0)),
				diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("1001")), // Subscription-Id-Data (MSISDN)
			}})
		slr.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
			AVP: []*diam.AVP{
				diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(1)),
				diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("104502200011")), // Subscription-Id-Data (IMSI)
			}})
		// t.Log("sendingg msg: ", slr.PrettyDump())
		if err := diamClientSy.SendMessage(slr); err != nil {
			t.Errorf("failed to send diameter message: %v", err)
		}

		reply := diamClientSy.ReceivedMessage(2 * time.Second)
		if reply == nil {
			t.Fatal("received empty reply")
		}
		// t.Log(reply.PrettyDump())
		avps, err := reply.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID)
		if err != nil {
			t.Error(err)
		}
		if len(avps) == 0 {
			t.Fatal("missing AVPs in reply")
		}

		resultCode, err := diamAVPAsString(avps[0])
		if err != nil {
			t.Error(err)
		}
		if resultCode != "5570" {
			t.Errorf("Result-Code=%s, want 5570", resultCode)
		}

		var replyActSess []*sessions.ExternalSession
		if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions, utils.SessionFilter{}, &replyActSess); err == nil || err.Error() != "NOT_FOUND" {
			t.Errorf("expected error <NOT_FOUND>, received <%v>", err)
		}
		var tIDs []string
		if err := client.Call(context.Background(), utils.ThresholdSv1GetThresholdIDs,
			&utils.TenantWithAPIOpts{}, &tIDs); err != nil {
			t.Error(err)
		} else if len(tIDs) != 0 {
			t.Errorf("expected no Threshold profiles, received <%v>", tIDs)
		}

		if err := engine.KillEngine(100); err != nil {
			t.Fatal(err)
		}
	})

}
