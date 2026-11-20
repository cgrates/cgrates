//go:build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package agents

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
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

func syThresholdID(originID, originHost, status string) string {
	return utils.ConcatenatedKey(
		utils.MetaSy+utils.Underline+originID+utils.Underline+originHost,
		status,
	)
}

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
		LogBuffer: &bytes.Buffer{},
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
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	engine.WaitForService(t, ctx, client, utils.SessionSv1)
	time.Sleep(100 * time.Millisecond) // wait for DiameterAgent service to start

	// Start monitoring SL
	diamClientSy, err := NewDiameterClient(cfg.DiameterAgentCfg().Listeners[0].Address, "Sy-CGR-DA",
		cfg.DiameterAgentCfg().OriginRealm, cfg.DiameterAgentCfg().VendorID,
		cfg.DiameterAgentCfg().ProductName, utils.DiameterFirmwareRevision,
		cfg.DiameterAgentCfg().DictionariesPath, cfg.DiameterAgentCfg().Listeners[0].Network)
	if err != nil {
		t.Fatal(err)
	}
	syOriginID := utils.UUIDSha1Prefix()
	slr := diam.NewRequest(diam.SpendingLimit, 16777302, nil)
	slr.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(syOriginID))
	slr.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("Sy-CGR-DA"))
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
	if len(tIDs) != 2 {
		t.Errorf("expected 2 threshold profile to be created, received <%v>", tIDs)
	}
	slices.Sort(tIDs)
	if tIDs[0] != syThresholdID(syOriginID, "Sy-CGR-DA", "THD_NORMAL") {
		t.Errorf("expected <%v>, received <%v>", syThresholdID(syOriginID, "Sy-CGR-DA", "THD_NORMAL"), tIDs[0])
	} else if tIDs[1] != syThresholdID(syOriginID, "Sy-CGR-DA", "THD_THROTTLE") {
		t.Errorf("expected <%v>, received <%v>", syThresholdID(syOriginID, "Sy-CGR-DA", "THD_THROTTLE"), tIDs[1])
	}
	expTp := &engine.ThresholdProfile{
		Tenant: "cgrates.org",
		ID:     syThresholdID(syOriginID, "Sy-CGR-DA", "THD_NORMAL"),
		FilterIDs: []string{
			"*gte:~*asm.BalanceSummaries.balance_data.Value:1024",
			"*string:~*req.AccountID:1001",
		},
		MaxHits:   1,
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
	expTp = &engine.ThresholdProfile{
		Tenant: "cgrates.org",
		ID:     syThresholdID(syOriginID, "Sy-CGR-DA", "THD_THROTTLE"),
		FilterIDs: []string{
			"*lte:~*asm.BalanceSummaries.balance_data.Value:1023",
			"*string:~*req.AccountID:1001",
		},
		MaxHits:   1,
		Async:     true,
		ActionIDs: []string{utils.MetaSyPublish},
	}
	if err := client.Call(context.Background(), utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{ID: tIDs[1]}, &tp); err != nil {
		t.Error(err)
	}
	expTp.ActivationInterval = tp.ActivationInterval
	slices.Sort(tp.FilterIDs) // sort filters received since they are processed from a map
	if !reflect.DeepEqual(tp, expTp) {
		t.Errorf("expected <%v>\nreceived\n<%v>", utils.ToJSON(expTp), utils.ToJSON(tp))
	}

	diamClientRo, err := NewDiameterClient(cfg.DiameterAgentCfg().Listeners[0].Address, cfg.DiameterAgentCfg().OriginHost,
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

	snaSent := make(chan struct{})
	// prepare to receive SNR when updating CCR
	go func() {
		// receive SNR from the first change on account even though status didnt change, any changes to the account that dont change the status after this, should not send an SNR
		reply = diamClientSy.ReceivedMessage(2 * time.Second)
		if reply == nil {
			t.Fatal("received empty reply")
		}
		expected := fmt.Sprintf(`Spending-Status-Notification-Request (SNR)
%s
	Session-Id {Code:263,Flags:0x40,Length:16,VendorId:0,Value:UTF8String{%s},Padding:1}
	Origin-Host {Code:264,Flags:0x40,Length:20,VendorId:0,Value:DiameterIdentity{Sy-CGR-DA},Padding:3}
	Origin-Realm {Code:296,Flags:0x40,Length:20,VendorId:0,Value:DiameterIdentity{cgrates.org},Padding:1}
	Destination-Realm {Code:283,Flags:0x40,Length:24,VendorId:0,Value:DiameterIdentity{dr-cgrates.org},Padding:2}
	Destination-Host {Code:293,Flags:0x40,Length:20,VendorId:0,Value:DiameterIdentity{CGR-DA-DH},Padding:3}
	Auth-Application-Id {Code:258,Flags:0x40,Length:12,VendorId:0,Value:Unsigned32{16777302}}
	Policy-Counter-Status-Report {Code:2903,Flags:0xc0,Length:56,VendorId:10415,Value:Grouped{
		Policy-Counter-Identifier {Code:2901,Flags:0xc0,Length:20,VendorId:10415,Value:UTF8String{Monthly},Padding:1},
		Policy-Counter-Status {Code:2902,Flags:0xc0,Length:24,VendorId:10415,Value:UTF8String{THD_NORMAL},Padding:2},
	}}
`, reply.Header, syOriginID)
		if expected != reply.String() {
			t.Fatalf("expected \n<%v>, \nreceived \n<%v>", expected, reply.String())
		}
		// Send SNA
		sna := diam.NewRequest(SSN, 16777302, nil)
		sna.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(syOriginID))
		sna.NewAVP(avp.ResultCode, avp.Mbit, 0, datatype.Unsigned32(2001))
		sna.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("Sy-CGR-DA"))
		sna.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
		// t.Log("sendingg msg: ", sna.PrettyDump())
		if err := diamClientSy.SendMessage(sna); err != nil {
			t.Errorf("failed to send diameter message: %v", err)
		}

		// receive SNR when policy status changes
		reply = diamClientSy.ReceivedMessage(2 * time.Second)
		expected = fmt.Sprintf(`Spending-Status-Notification-Request (SNR)
%s
	Session-Id {Code:263,Flags:0x40,Length:16,VendorId:0,Value:UTF8String{%s},Padding:1}
	Origin-Host {Code:264,Flags:0x40,Length:20,VendorId:0,Value:DiameterIdentity{Sy-CGR-DA},Padding:3}
	Origin-Realm {Code:296,Flags:0x40,Length:20,VendorId:0,Value:DiameterIdentity{cgrates.org},Padding:1}
	Destination-Realm {Code:283,Flags:0x40,Length:24,VendorId:0,Value:DiameterIdentity{dr-cgrates.org},Padding:2}
	Destination-Host {Code:293,Flags:0x40,Length:20,VendorId:0,Value:DiameterIdentity{CGR-DA-DH},Padding:3}
	Auth-Application-Id {Code:258,Flags:0x40,Length:12,VendorId:0,Value:Unsigned32{16777302}}
	Policy-Counter-Status-Report {Code:2903,Flags:0xc0,Length:56,VendorId:10415,Value:Grouped{
		Policy-Counter-Identifier {Code:2901,Flags:0xc0,Length:20,VendorId:10415,Value:UTF8String{Monthly},Padding:1},
		Policy-Counter-Status {Code:2902,Flags:0xc0,Length:24,VendorId:10415,Value:UTF8String{THD_THROTTLE},Padding:0},
	}}
`, reply.Header, syOriginID)
		if expected != reply.String() {
			t.Errorf("expected \n<%v>, \nreceived \n<%v>", expected, reply.String())
		}
		// Send SNA
		sna = diam.NewRequest(SSN, 16777302, nil)
		sna.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(syOriginID))
		sna.NewAVP(avp.ResultCode, avp.Mbit, 0, datatype.Unsigned32(2001))
		sna.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("Sy-CGR-DA"))
		sna.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
		// t.Log("sendingg msg: ", sna.PrettyDump())
		if err := diamClientSy.SendMessage(sna); err != nil {
			t.Errorf("failed to send diameter message: %v", err)
		}
		close(snaSent)
	}()

	// CCR-U
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
					diam.NewAVP(avp.CCTotalOctets, avp.Mbit, 0, datatype.Unsigned64(2050)),  // CC-Total-Octets
					diam.NewAVP(avp.CCInputOctets, avp.Mbit, 0, datatype.Unsigned64(1025)),  // CC-Input-Octets
					diam.NewAVP(avp.CCOutputOctets, avp.Mbit, 0, datatype.Unsigned64(1025)), // CC-Output-Octets
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

	select { // make sure sna is sent before continuing
	case <-snaSent:
	case <-time.After(5 * time.Second):
		t.Fatal("took too long to get reply from SNR")
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

	expBalance = float64(0) // CCR update should take 2050 units
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

	snaSent = make(chan struct{})
	// prepare to receive SNR when recharging
	go func() {
		// receive SNR
		reply = diamClientSy.ReceivedMessage(2 * time.Second)
		if reply == nil {
			t.Fatal("received nil reply")
		}
		expected := fmt.Sprintf(`Spending-Status-Notification-Request (SNR)
%s
	Session-Id {Code:263,Flags:0x40,Length:16,VendorId:0,Value:UTF8String{%s},Padding:1}
	Origin-Host {Code:264,Flags:0x40,Length:20,VendorId:0,Value:DiameterIdentity{Sy-CGR-DA},Padding:3}
	Origin-Realm {Code:296,Flags:0x40,Length:20,VendorId:0,Value:DiameterIdentity{cgrates.org},Padding:1}
	Destination-Realm {Code:283,Flags:0x40,Length:24,VendorId:0,Value:DiameterIdentity{dr-cgrates.org},Padding:2}
	Destination-Host {Code:293,Flags:0x40,Length:20,VendorId:0,Value:DiameterIdentity{CGR-DA-DH},Padding:3}
	Auth-Application-Id {Code:258,Flags:0x40,Length:12,VendorId:0,Value:Unsigned32{16777302}}
	Policy-Counter-Status-Report {Code:2903,Flags:0xc0,Length:56,VendorId:10415,Value:Grouped{
		Policy-Counter-Identifier {Code:2901,Flags:0xc0,Length:20,VendorId:10415,Value:UTF8String{Monthly},Padding:1},
		Policy-Counter-Status {Code:2902,Flags:0xc0,Length:24,VendorId:10415,Value:UTF8String{THD_NORMAL},Padding:2},
	}}
`, reply.Header, syOriginID)
		if expected != reply.String() {
			t.Errorf("expected \n<%v>, \nreceived \n<%v>", expected, reply.String())
		}
		// Send SNA
		sna := diam.NewRequest(SSN, 16777302, nil)
		sna.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(syOriginID))
		sna.NewAVP(avp.ResultCode, avp.Mbit, 0, datatype.Unsigned32(2001))
		sna.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("Sy-CGR-DA"))
		sna.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
		// t.Log("sendingg msg: ", sna.PrettyDump())
		if err := diamClientSy.SendMessage(sna); err != nil {
			t.Errorf("failed to send diameter message: %v", err)
		}
		close(snaSent)
	}()

	// topup the acc (sending another SNR to PCRF)
	var actRply string
	attrsEA := &utils.AttrExecuteAction{Tenant: "cgrates.org", Account: "1001", ActionsId: "ACT_TOPUP"}
	if err := client.Call(context.Background(), utils.APIerSv1ExecuteAction, attrsEA, &actRply); err != nil {
		t.Errorf("APIerSv1ExecuteAction failed unexpectedly: %v", err)
	}

	expBalance = float64(3072)
	if err = client.Call(context.Background(), utils.APIerSv2GetAccount, attrsAcnt, &acnt); err != nil {
		t.Errorf("APIerSv1.GetAccount unexpected err: %v", err)
	} else if rply := acnt.BalanceMap[utils.MetaData].GetTotalValue(); rply != expBalance {
		t.Errorf("APIerSv1.GetAccount: data_balance: %f, want: %f", rply, expBalance)
	}

	select { // make sure sna is sent before continuing
	case <-snaSent:
	case <-time.After(5 * time.Second):
		t.Fatal("took too long to get reply from SNR")
	}

	snaSent = make(chan struct{})
	// prepare to receive SNR when updating CCR again
	go func() {
		// receive SNR when policy status changes
		reply := diamClientSy.ReceivedMessage(2 * time.Second)
		expected := fmt.Sprintf(`Spending-Status-Notification-Request (SNR)
%s
	Session-Id {Code:263,Flags:0x40,Length:16,VendorId:0,Value:UTF8String{%s},Padding:1}
	Origin-Host {Code:264,Flags:0x40,Length:20,VendorId:0,Value:DiameterIdentity{Sy-CGR-DA},Padding:3}
	Origin-Realm {Code:296,Flags:0x40,Length:20,VendorId:0,Value:DiameterIdentity{cgrates.org},Padding:1}
	Destination-Realm {Code:283,Flags:0x40,Length:24,VendorId:0,Value:DiameterIdentity{dr-cgrates.org},Padding:2}
	Destination-Host {Code:293,Flags:0x40,Length:20,VendorId:0,Value:DiameterIdentity{CGR-DA-DH},Padding:3}
	Auth-Application-Id {Code:258,Flags:0x40,Length:12,VendorId:0,Value:Unsigned32{16777302}}
	Policy-Counter-Status-Report {Code:2903,Flags:0xc0,Length:56,VendorId:10415,Value:Grouped{
		Policy-Counter-Identifier {Code:2901,Flags:0xc0,Length:20,VendorId:10415,Value:UTF8String{Monthly},Padding:1},
		Policy-Counter-Status {Code:2902,Flags:0xc0,Length:24,VendorId:10415,Value:UTF8String{THD_THROTTLE},Padding:0},
	}}
`, reply.Header, syOriginID)
		if expected != reply.String() {
			t.Errorf("expected \n<%v>, \nreceived \n<%v>", expected, reply.String())
		}
		// Send SNA
		sna := diam.NewRequest(SSN, 16777302, nil)
		sna.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(syOriginID))
		sna.NewAVP(avp.ResultCode, avp.Mbit, 0, datatype.Unsigned32(2001))
		sna.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("Sy-CGR-DA"))
		sna.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
		// t.Log("sendingg msg: ", sna.PrettyDump())
		if err := diamClientSy.SendMessage(sna); err != nil {
			t.Errorf("failed to send diameter message: %v", err)
		}
		close(snaSent)
	}()

	// spend balance again
	// CCR-U
	// t.Log("sendingg CCR-U msg: ", ccrU.PrettyDump())
	if err := diamClientRo.SendMessage(ccrU); err != nil {
		t.Errorf("failed to send diameter message: %v", err)
	}

	select { // make sure sna is sent before continuing
	case <-snaSent:
	case <-time.After(5 * time.Second):
		t.Fatal("took too long to get reply from SNR")
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

	expBalance = float64(1022) // CCR update should take 2050 units
	if err = client.Call(context.Background(), utils.APIerSv2GetAccount, attrsAcnt, &acnt); err != nil {
		t.Errorf("APIerSv1.GetAccount unexpected err: %v", err)
	} else if rply := acnt.BalanceMap[utils.MetaData].GetTotalValue(); rply != expBalance {
		t.Errorf("APIerSv1.GetAccount: data_balance: %f, want: %f", rply, expBalance)
	}

	// Change session policy subscription
	slri := diam.NewRequest(diam.SpendingLimit, 16777302, nil)
	slri.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(syOriginID))
	slri.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("Sy-CGR-DA"))
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

	expBalance = float64(1022) // balance shouldnt be changed from SLR update
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
	} else if tIDs[0] != syThresholdID(syOriginID, "Sy-CGR-DA", "NEVER_THROTTLE") {
		t.Errorf("expected <%v>, received <%v>", syThresholdID(syOriginID, "Sy-CGR-DA", "NEVER_THROTTLE"), tIDs[0])
	}
	expTp = &engine.ThresholdProfile{
		Tenant: "cgrates.org",
		ID:     syThresholdID(syOriginID, "Sy-CGR-DA", "NEVER_THROTTLE"),
		FilterIDs: []string{
			"*lte:~*asm.BalanceSummaries.balance_data.Value:-1",
			"*string:~*req.AccountID:1001",
		},
		MaxHits:   1,
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

	// CCR-T
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
					diam.NewAVP(avp.CCTotalOctets, avp.Mbit, 0, datatype.Unsigned64(5000)),  // CC-Total-Octets
					diam.NewAVP(avp.CCInputOctets, avp.Mbit, 0, datatype.Unsigned64(2500)),  // CC-Input-Octets
					diam.NewAVP(avp.CCOutputOctets, avp.Mbit, 0, datatype.Unsigned64(2500)), // CC-Output-Octets
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

	expBalance = float64(1022) // CCR-T shoulnt take debit since no usage was used
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
	str.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("Sy-CGR-DA"))
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

	expBalance = float64(1022)
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

	// Sy opens a BiRPC connection from SessionS to ThresholdS.
	// SessionS must not call AgentV1 methods on that connection.
	t.Run("DisconnectPeer", func(t *testing.T) {
		done := make(chan error, 1)
		// The API waits for a DPA, so answer the DPR while it runs.
		go func() {
			var reply string
			// Use the Sy peer's address because the Ro peer has the same
			// OriginHost and OriginRealm.
			done <- client.Call(context.Background(), utils.SessionSv1DisconnectPeer,
				&utils.DPRArgs{RemoteAddr: diamClientSy.conn.LocalAddr().String()}, &reply)
		}()
		dpr := diamClientSy.ReceivedMessage(2 * time.Second)
		if dpr == nil {
			t.Fatal("DiameterAgent did not send a DPR to the diameter peer")
		}
		dpa := dpr.Answer(diam.Success)
		dpa.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("localhost"))
		dpa.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity(cfg.DiameterAgentCfg().OriginRealm))
		if err := diamClientSy.SendMessage(dpa); err != nil {
			t.Fatal(err)
		}
		if err := <-done; err != nil {
			t.Fatalf("SessionSv1.DisconnectPeer: %v", err)
		}
	})
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
		LogBuffer: &bytes.Buffer{},
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
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	engine.WaitForService(t, ctx, client, utils.SessionSv1)
	time.Sleep(100 * time.Millisecond) // wait for DiameterAgent service to start

	// Start monitoring SL
	diamClientSy, err := NewDiameterClient(cfg.DiameterAgentCfg().Listeners[0].Address, "Sy-CGR-DA",
		cfg.DiameterAgentCfg().OriginRealm, cfg.DiameterAgentCfg().VendorID,
		cfg.DiameterAgentCfg().ProductName, utils.DiameterFirmwareRevision,
		cfg.DiameterAgentCfg().DictionariesPath, cfg.DiameterAgentCfg().Listeners[0].Network)
	if err != nil {
		t.Fatal(err)
	}
	syOriginID := utils.UUIDSha1Prefix()
	slr := diam.NewRequest(diam.SpendingLimit, 16777302, nil)
	slr.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(syOriginID))
	slr.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("Sy-CGR-DA"))
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
	if len(tIDs) != 2 {
		t.Errorf("expected 2 threshold profile to be created, received <%v>", tIDs)
	}
	slices.Sort(tIDs)
	if tIDs[0] != syThresholdID(syOriginID, "Sy-CGR-DA", "THD_NORMAL") {
		t.Errorf("expected <%v>, received <%v>", syThresholdID(syOriginID, "Sy-CGR-DA", "THD_NORMAL"), tIDs[0])
	} else if tIDs[1] != syThresholdID(syOriginID, "Sy-CGR-DA", "THD_THROTTLE") {
		t.Errorf("expected <%v>, received <%v>", syThresholdID(syOriginID, "Sy-CGR-DA", "THD_THROTTLE"), tIDs[1])
	}
	expTp := &engine.ThresholdProfile{
		Tenant: "cgrates.org",
		ID:     syThresholdID(syOriginID, "Sy-CGR-DA", "THD_NORMAL"),
		FilterIDs: []string{
			"*gte:~*asm.BalanceSummaries.balance_data.Value:1024",
			"*string:~*req.AccountID:1001",
		},
		MaxHits:   1,
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
	expTp = &engine.ThresholdProfile{
		Tenant: "cgrates.org",
		ID:     syThresholdID(syOriginID, "Sy-CGR-DA", "THD_THROTTLE"),
		FilterIDs: []string{
			"*lte:~*asm.BalanceSummaries.balance_data.Value:1023",
			"*string:~*req.AccountID:1001",
		},
		MaxHits:   1,
		Async:     true,
		ActionIDs: []string{utils.MetaSyPublish},
	}
	if err := client.Call(context.Background(), utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{ID: tIDs[1]}, &tp); err != nil {
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
	// fmt.Println(ng.LogBuffer)
	// ng.LogBuffer = &bytes.Buffer{}
	// t.Cleanup(func() { fmt.Println(ng.LogBuffer) })
	ng.PreserveDataDB = true
	ng.TpFiles = nil // dont re-load tariffplans (if we were to re-load them, it would trigger the thresholds for THD_NORMAL before the ClientConnIDs wouldve been reestablished for it from restoreSessions)
	client, cfg = ng.Run(t)
	ctx, cancel = context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	engine.WaitForService(t, ctx, client, utils.SessionSv1)
	time.Sleep(100 * time.Millisecond) // wait for DiameterAgent service since its waiting on sessions
	// reopen connection with client after we shutdown engine

	diamClientRo, err := NewDiameterClient(cfg.DiameterAgentCfg().Listeners[0].Address, cfg.DiameterAgentCfg().OriginHost,
		cfg.DiameterAgentCfg().OriginRealm, cfg.DiameterAgentCfg().VendorID,
		cfg.DiameterAgentCfg().ProductName, utils.DiameterFirmwareRevision,
		cfg.DiameterAgentCfg().DictionariesPath, cfg.DiameterAgentCfg().Listeners[0].Network)
	if err != nil {
		t.Fatal(err)
	}
	diamClientSy, err = NewDiameterClient(cfg.DiameterAgentCfg().Listeners[0].Address, "Sy-CGR-DA",
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

	snaSent := make(chan struct{})
	// prepare to receive SNR when updating CCR
	go func() {
		// receive SNR from the first change on account even though status didnt change, any changes to the account that dont change the status after this, should not send an SNR
		reply = diamClientSy.ReceivedMessage(2 * time.Second)
		if reply == nil {
			t.Fatal("nil reply")
		}
		expected := fmt.Sprintf(`Spending-Status-Notification-Request (SNR)
%s
	Session-Id {Code:263,Flags:0x40,Length:16,VendorId:0,Value:UTF8String{%s},Padding:1}
	Origin-Host {Code:264,Flags:0x40,Length:20,VendorId:0,Value:DiameterIdentity{Sy-CGR-DA},Padding:3}
	Origin-Realm {Code:296,Flags:0x40,Length:20,VendorId:0,Value:DiameterIdentity{cgrates.org},Padding:1}
	Destination-Realm {Code:283,Flags:0x40,Length:24,VendorId:0,Value:DiameterIdentity{dr-cgrates.org},Padding:2}
	Destination-Host {Code:293,Flags:0x40,Length:20,VendorId:0,Value:DiameterIdentity{CGR-DA-DH},Padding:3}
	Auth-Application-Id {Code:258,Flags:0x40,Length:12,VendorId:0,Value:Unsigned32{16777302}}
	Policy-Counter-Status-Report {Code:2903,Flags:0xc0,Length:56,VendorId:10415,Value:Grouped{
		Policy-Counter-Identifier {Code:2901,Flags:0xc0,Length:20,VendorId:10415,Value:UTF8String{Monthly},Padding:1},
		Policy-Counter-Status {Code:2902,Flags:0xc0,Length:24,VendorId:10415,Value:UTF8String{THD_NORMAL},Padding:2},
	}}
`, reply.Header, syOriginID)
		if expected != reply.String() {
			t.Fatalf("expected \n<%v>, \nreceived \n<%v>", expected, reply.String())
		}
		// Send SNA
		sna := diam.NewRequest(SSN, 16777302, nil)
		sna.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(syOriginID))
		sna.NewAVP(avp.ResultCode, avp.Mbit, 0, datatype.Unsigned32(2001))
		sna.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("Sy-CGR-DA"))
		sna.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
		// t.Log("sendingg msg: ", sna.PrettyDump())
		if err := diamClientSy.SendMessage(sna); err != nil {
			t.Errorf("failed to send diameter message: %v", err)
		}

		// receive SNR when policy status changes
		reply = diamClientSy.ReceivedMessage(2 * time.Second)
		expected = fmt.Sprintf(`Spending-Status-Notification-Request (SNR)
%s
	Session-Id {Code:263,Flags:0x40,Length:16,VendorId:0,Value:UTF8String{%s},Padding:1}
	Origin-Host {Code:264,Flags:0x40,Length:20,VendorId:0,Value:DiameterIdentity{Sy-CGR-DA},Padding:3}
	Origin-Realm {Code:296,Flags:0x40,Length:20,VendorId:0,Value:DiameterIdentity{cgrates.org},Padding:1}
	Destination-Realm {Code:283,Flags:0x40,Length:24,VendorId:0,Value:DiameterIdentity{dr-cgrates.org},Padding:2}
	Destination-Host {Code:293,Flags:0x40,Length:20,VendorId:0,Value:DiameterIdentity{CGR-DA-DH},Padding:3}
	Auth-Application-Id {Code:258,Flags:0x40,Length:12,VendorId:0,Value:Unsigned32{16777302}}
	Policy-Counter-Status-Report {Code:2903,Flags:0xc0,Length:56,VendorId:10415,Value:Grouped{
		Policy-Counter-Identifier {Code:2901,Flags:0xc0,Length:20,VendorId:10415,Value:UTF8String{Monthly},Padding:1},
		Policy-Counter-Status {Code:2902,Flags:0xc0,Length:24,VendorId:10415,Value:UTF8String{THD_THROTTLE},Padding:0},
	}}
`, reply.Header, syOriginID)
		if expected != reply.String() {
			t.Errorf("expected \n<%v>, \nreceived \n<%v>", expected, reply.String())
		}
		// Send SNA
		sna = diam.NewRequest(SSN, 16777302, nil)
		sna.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(syOriginID))
		sna.NewAVP(avp.ResultCode, avp.Mbit, 0, datatype.Unsigned32(2001))
		sna.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("Sy-CGR-DA"))
		sna.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
		// t.Log("sendingg msg: ", sna.PrettyDump())
		if err := diamClientSy.SendMessage(sna); err != nil {
			t.Errorf("failed to send diameter message: %v", err)
		}
		close(snaSent)
	}()

	// CCR-U
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
					diam.NewAVP(avp.CCTotalOctets, avp.Mbit, 0, datatype.Unsigned64(2050)),  // CC-Total-Octets
					diam.NewAVP(avp.CCInputOctets, avp.Mbit, 0, datatype.Unsigned64(1025)),  // CC-Input-Octets
					diam.NewAVP(avp.CCOutputOctets, avp.Mbit, 0, datatype.Unsigned64(1025)), // CC-Output-Octets
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

	select { // make sure sna is sent before continuing
	case <-snaSent:
	case <-time.After(5 * time.Second):
		t.Fatal("took too long to get reply from SNR")
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

	expBalance = float64(0) // CCR update should take 2050 units
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

	snaSent = make(chan struct{})
	// prepare to receive SNR when recharging
	go func() {
		// receive SNR
		reply = diamClientSy.ReceivedMessage(2 * time.Second)
		if reply == nil {
			t.Fatal("received nil reply")
		}
		expected := fmt.Sprintf(`Spending-Status-Notification-Request (SNR)
%s
	Session-Id {Code:263,Flags:0x40,Length:16,VendorId:0,Value:UTF8String{%s},Padding:1}
	Origin-Host {Code:264,Flags:0x40,Length:20,VendorId:0,Value:DiameterIdentity{Sy-CGR-DA},Padding:3}
	Origin-Realm {Code:296,Flags:0x40,Length:20,VendorId:0,Value:DiameterIdentity{cgrates.org},Padding:1}
	Destination-Realm {Code:283,Flags:0x40,Length:24,VendorId:0,Value:DiameterIdentity{dr-cgrates.org},Padding:2}
	Destination-Host {Code:293,Flags:0x40,Length:20,VendorId:0,Value:DiameterIdentity{CGR-DA-DH},Padding:3}
	Auth-Application-Id {Code:258,Flags:0x40,Length:12,VendorId:0,Value:Unsigned32{16777302}}
	Policy-Counter-Status-Report {Code:2903,Flags:0xc0,Length:56,VendorId:10415,Value:Grouped{
		Policy-Counter-Identifier {Code:2901,Flags:0xc0,Length:20,VendorId:10415,Value:UTF8String{Monthly},Padding:1},
		Policy-Counter-Status {Code:2902,Flags:0xc0,Length:24,VendorId:10415,Value:UTF8String{THD_NORMAL},Padding:2},
	}}
`, reply.Header, syOriginID)
		if expected != reply.String() {
			t.Errorf("expected \n<%v>, \nreceived \n<%v>", expected, reply.String())
		}
		// Send SNA
		sna := diam.NewRequest(SSN, 16777302, nil)
		sna.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(syOriginID))
		sna.NewAVP(avp.ResultCode, avp.Mbit, 0, datatype.Unsigned32(2001))
		sna.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("Sy-CGR-DA"))
		sna.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
		// t.Log("sendingg msg: ", sna.PrettyDump())
		if err := diamClientSy.SendMessage(sna); err != nil {
			t.Errorf("failed to send diameter message: %v", err)
		}
		close(snaSent)
	}()

	// topup the acc (sending another SNR to PCRF)
	var actRply string
	attrsEA := &utils.AttrExecuteAction{Tenant: "cgrates.org", Account: "1001", ActionsId: "ACT_TOPUP"}
	if err := client.Call(context.Background(), utils.APIerSv1ExecuteAction, attrsEA, &actRply); err != nil {
		t.Errorf("APIerSv1ExecuteAction failed unexpectedly: %v", err)
	}

	expBalance = float64(3072)
	if err = client.Call(context.Background(), utils.APIerSv2GetAccount, attrsAcnt, &acnt); err != nil {
		t.Errorf("APIerSv1.GetAccount unexpected err: %v", err)
	} else if rply := acnt.BalanceMap[utils.MetaData].GetTotalValue(); rply != expBalance {
		t.Errorf("APIerSv1.GetAccount: data_balance: %f, want: %f", rply, expBalance)
	}

	select { // make sure sna is sent before continuing
	case <-snaSent:
	case <-time.After(5 * time.Second):
		t.Fatal("took too long to get reply from SNR")
	}

	snaSent = make(chan struct{})
	// prepare to receive SNR when updating CCR again
	go func() {
		// receive SNR when policy status changes
		reply := diamClientSy.ReceivedMessage(2 * time.Second)
		expected := fmt.Sprintf(`Spending-Status-Notification-Request (SNR)
%s
	Session-Id {Code:263,Flags:0x40,Length:16,VendorId:0,Value:UTF8String{%s},Padding:1}
	Origin-Host {Code:264,Flags:0x40,Length:20,VendorId:0,Value:DiameterIdentity{Sy-CGR-DA},Padding:3}
	Origin-Realm {Code:296,Flags:0x40,Length:20,VendorId:0,Value:DiameterIdentity{cgrates.org},Padding:1}
	Destination-Realm {Code:283,Flags:0x40,Length:24,VendorId:0,Value:DiameterIdentity{dr-cgrates.org},Padding:2}
	Destination-Host {Code:293,Flags:0x40,Length:20,VendorId:0,Value:DiameterIdentity{CGR-DA-DH},Padding:3}
	Auth-Application-Id {Code:258,Flags:0x40,Length:12,VendorId:0,Value:Unsigned32{16777302}}
	Policy-Counter-Status-Report {Code:2903,Flags:0xc0,Length:56,VendorId:10415,Value:Grouped{
		Policy-Counter-Identifier {Code:2901,Flags:0xc0,Length:20,VendorId:10415,Value:UTF8String{Monthly},Padding:1},
		Policy-Counter-Status {Code:2902,Flags:0xc0,Length:24,VendorId:10415,Value:UTF8String{THD_THROTTLE},Padding:0},
	}}
`, reply.Header, syOriginID)
		if expected != reply.String() {
			t.Errorf("expected \n<%v>, \nreceived \n<%v>", expected, reply.String())
		}
		// Send SNA
		sna := diam.NewRequest(SSN, 16777302, nil)
		sna.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(syOriginID))
		sna.NewAVP(avp.ResultCode, avp.Mbit, 0, datatype.Unsigned32(2001))
		sna.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("Sy-CGR-DA"))
		sna.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
		// t.Log("sendingg msg: ", sna.PrettyDump())
		if err := diamClientSy.SendMessage(sna); err != nil {
			t.Errorf("failed to send diameter message: %v", err)
		}
		close(snaSent)
	}()

	// spend balance again
	// CCR-U
	// t.Log("sendingg CCR-U msg: ", ccrU.PrettyDump())
	if err := diamClientRo.SendMessage(ccrU); err != nil {
		t.Errorf("failed to send diameter message: %v", err)
	}

	select { // make sure sna is sent before continuing
	case <-snaSent:
	case <-time.After(5 * time.Second):
		t.Fatal("took too long to get reply from SNR")
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

	expBalance = float64(1022) // CCR update should take 2050 units
	if err = client.Call(context.Background(), utils.APIerSv2GetAccount, attrsAcnt, &acnt); err != nil {
		t.Errorf("APIerSv1.GetAccount unexpected err: %v", err)
	} else if rply := acnt.BalanceMap[utils.MetaData].GetTotalValue(); rply != expBalance {
		t.Errorf("APIerSv1.GetAccount: data_balance: %f, want: %f", rply, expBalance)
	}

	// Change session policy subscription
	slri := diam.NewRequest(diam.SpendingLimit, 16777302, nil)
	slri.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(syOriginID))
	slri.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("Sy-CGR-DA"))
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
	// get thresholds
	if err := client.Call(context.Background(), utils.ThresholdSv1GetThresholdIDs,
		&utils.TenantWithAPIOpts{}, &tIDs); err != nil {
		t.Error(err)
	}
	if len(tIDs) != 1 {
		t.Errorf("expected 1 threshold profile to be created, received <%v>", tIDs)
	} else if tIDs[0] != syThresholdID(syOriginID, "Sy-CGR-DA", "NEVER_THROTTLE") {
		t.Errorf("expected <%v>, received <%v>", syThresholdID(syOriginID, "Sy-CGR-DA", "NEVER_THROTTLE"), tIDs[0])
	}
	expTp = &engine.ThresholdProfile{
		Tenant: "cgrates.org",
		ID:     syThresholdID(syOriginID, "Sy-CGR-DA", "NEVER_THROTTLE"),
		FilterIDs: []string{
			"*lte:~*asm.BalanceSummaries.balance_data.Value:-1",
			"*string:~*req.AccountID:1001",
		},
		MaxHits:   1,
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

	expBalance = float64(1022) // balance shouldnt be changed from SLR update
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

	// CCR-T
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
					diam.NewAVP(avp.CCTotalOctets, avp.Mbit, 0, datatype.Unsigned64(5000)),  // CC-Total-Octets
					diam.NewAVP(avp.CCInputOctets, avp.Mbit, 0, datatype.Unsigned64(2500)),  // CC-Input-Octets
					diam.NewAVP(avp.CCOutputOctets, avp.Mbit, 0, datatype.Unsigned64(2500)), // CC-Output-Octets
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

	expBalance = float64(1022) // CCR-T shoulnt take debit since no usage was used
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
	str.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("Sy-CGR-DA"))
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

	expBalance = float64(1022)
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
	diamClientSy, err := NewDiameterClient(cfg.DiameterAgentCfg().Listeners[0].Address, cfg.DiameterAgentCfg().OriginHost,
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
		t.Fatal(err)
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
	if len(tIDs) != 2 {
		t.Errorf("expected 2 threshold profiles to be created, received <%v>", tIDs)
	}
	slices.Sort(tIDs)
	if tIDs[0] != syThresholdID(syOriginID, "CGR-DA", "THD_NORMAL") {
		t.Errorf("expected <%v>, received <%v>", syThresholdID(syOriginID, "CGR-DA", "THD_NORMAL"), tIDs[0])
	} else if tIDs[1] != syThresholdID(syOriginID, "CGR-DA", "THD_THROTTLE") {
		t.Errorf("expected <%v>, received <%v>", syThresholdID(syOriginID, "CGR-DA", "THD_THROTTLE"), tIDs[1])
	}
	expTp := &engine.ThresholdProfile{
		Tenant: "cgrates.org",
		ID:     syThresholdID(syOriginID, "CGR-DA", "THD_NORMAL"),
		FilterIDs: []string{
			"*gt:~*asm.BalanceSummaries.balance_data.Value:0",
			"*string:~*req.AccountID:1001",
		},
		MaxHits:   1,
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
	expTp = &engine.ThresholdProfile{
		Tenant: "cgrates.org",
		ID:     syThresholdID(syOriginID, "CGR-DA", "THD_THROTTLE"),
		FilterIDs: []string{
			"*lte:~*asm.BalanceSummaries.balance_data.Value:0",
			"*string:~*req.AccountID:1001",
		},
		MaxHits:   1,
		Async:     true,
		ActionIDs: []string{utils.MetaSyPublish},
	}
	if err := client.Call(context.Background(), utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{ID: tIDs[1]}, &tp); err != nil {
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
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		engine.WaitForService(t, ctx, client, utils.SchedulerSv1)
		time.Sleep(100 * time.Millisecond) // wait for DiameterAgent service to start
		// Start start proper SL session
		diamClientSy, err := NewDiameterClient(cfg.DiameterAgentCfg().Listeners[0].Address, cfg.DiameterAgentCfg().OriginHost,
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
		} else if tIDs[0] != syThresholdID(syOriginID, "CGR-DA", "THD_THROTTLE") {
			t.Errorf("expected <%v>, received <%v>", syThresholdID(syOriginID, "CGR-DA", "THD_THROTTLE"), tIDs[0])
		}
		expTp := &engine.ThresholdProfile{
			Tenant: "cgrates.org",
			ID:     syThresholdID(syOriginID, "CGR-DA", "THD_THROTTLE"),
			FilterIDs: []string{
				"*lte:~*asm.BalanceSummaries.balance_data.Value:1023",
				"*string:~*req.AccountID:1001",
			},
			MaxHits:   1,
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
		} else if tIDs[0] != syThresholdID(syOriginID, "CGR-DA", "THD_THROTTLE") {
			t.Errorf("expected <%v>, received <%v>", syThresholdID(syOriginID, "CGR-DA", "THD_THROTTLE"), tIDs[0])
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
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		engine.WaitForService(t, ctx, client, utils.SchedulerSv1)
		time.Sleep(100 * time.Millisecond) // wait for DiameterAgent service to start
		// Start SL intermediate session that will error with DIAMETER_UNKNOWN_SESSION_ID
		diamClientSy, err := NewDiameterClient(cfg.DiameterAgentCfg().Listeners[0].Address, cfg.DiameterAgentCfg().OriginHost,
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
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		engine.WaitForService(t, ctx, client, utils.SchedulerSv1)
		time.Sleep(100 * time.Millisecond) // wait for DiameterAgent service to start
		// Start SL intermediate session that will error with DIAMETER_UNKNOWN_SESSION_ID
		diamClientSy, err := NewDiameterClient(cfg.DiameterAgentCfg().Listeners[0].Address, cfg.DiameterAgentCfg().OriginHost,
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
	// 	diamClientSy, err := NewDiameterClient(cfg.DiameterAgentCfg().Listeners[0].Address, cfg.DiameterAgentCfg().OriginHost,
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
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		engine.WaitForService(t, ctx, client, utils.SchedulerSv1)
		time.Sleep(100 * time.Millisecond) // wait for DiameterAgent service to start
		// Start start proper SL session
		diamClientSy, err := NewDiameterClient(cfg.DiameterAgentCfg().Listeners[0].Address, cfg.DiameterAgentCfg().OriginHost,
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

// test is used to make sure engine doesnt crash on restart when no sessions are backed up
func TestDiamSyNoSessionsBackupRestart(t *testing.T) {
	ng := engine.TestEngine{
		ConfigJSON: `{
  "general": {
    "log_level": 7
  },
  "schedulers": {
    "enabled": true
  },
  "data_db": {
    "db_type": "*internal",
    "opts": {
      "internalDBDumpPath": "/tmp/internal_db",
      "internalDBDumpInterval": "-1"
    }
  },
  "stor_db": {
    "db_type": "*internal"
  },
  "apiers": {
    "enabled": true,
    "scheduler_conns": ["*localhost"]
  },
  "filters": {
    "apiers_conns": ["*localhost"]
  },
  "thresholds": {
    "enabled": true
  },

  "rpc_conns": {
    "*bijson_localhost_thresholds": {
      "conns": [
        {
          "address": "127.0.0.1:2014",
          "transport": "*birpc_json"
        }
      ]
    }
  },
  "sessions": {
    "enabled": true,
    "apiers_conns": ["*localhost"],
    // "thresholds_conns": ["*birpc_internal"],
    "thresholds_conns": ["*bijson_localhost_thresholds"],
    "chargers_conns": ["*localhost"],
    "rals_conns": ["*localhost"],
  },
  "chargers": {
    "enabled": true,
    "string_indexed_fields": ["*req.Account"]
  },
  "rals": {
    "enabled": true,
    "thresholds_conns": ["*localhost"],
    "max_computed_usage": {
      "*data": "10737418"
    }
  },
  "diameter_agent": {
    "enabled": true,
    "listeners": [
      {
        "address": "0.0.0.0:3868"
      }
    ],
    "sessions_conns": ["*bijson_localhost"],
    "snr_template": "*snr",
    "request_processors": [
      {
        "id": "data_init",
        "filters": [
          "*string:~*vars.*cmd:CCR","*string:~*req.CC-Request-Type:1",
          "*string:~*req.Auth-Application-Id:4"
        ],
        "flags": ["*initiate", "*accounts", "*log"],
        "request_fields": [
          {"tag": "ToR","path": "*cgreq.ToR","type": "*constant","value": "*data"},
          {"tag": "Tenant","path": "*cgreq.Tenant","type": "*constant","value": "cgrates.org"},
          {"tag": "OriginID","path": "*cgreq.OriginID","type": "*variable","value": "~*req.Session-Id","mandatory": true},
          {"tag": "OriginHost","path": "*cgreq.OriginHost","type": "*variable","value": "~*req.Origin-Host"},
          {"tag": "SetupTime","path": "*cgreq.SetupTime","type": "*variable","value": "~*req.Event-Timestamp","mandatory": true},
          {"tag": "AnswerTime","path": "*cgreq.AnswerTime","type": "*variable","value": "~*req.Event-Timestamp","mandatory": true},
          {"tag": "RequestType","path": "*cgreq.RequestType","type": "*constant","value": "*prepaid"},
          {"tag": "Category","path": "*cgreq.Category","type": "*constant","value": "data"},
          {"tag": "Account","path": "*cgreq.Account","type": "*variable","value": "~*req.Subscription-Id.Subscription-Id-Data[~Subscription-Id-Type(0)]","mandatory": true},
          {"tag": "Destination","path": "*cgreq.Destination","type": "*variable","value": "~*req.Service-Information.PS-Information.3GPP-SGSN-MCC-MNC","mandatory": true},
          {"tag": "Subject","path": "*cgreq.Subject","type": "*variable","value": "~*req.Service-Information.PS-Information.Called-Station-Id"},
          {"tag": "Usage","path": "*cgreq.Usage","type": "*constant","value": "2048"}
        ],
        "reply_fields": [
          {"tag": "CCATemplate","type": "*template","value": "*cca"},
          {"tag": "RatingGroup","path": "*rep.Multiple-Services-Credit-Control.Rating-Group","type": "*constant","value": "1"},
          {"tag": "CCTotalOctets1","path": "*rep.Multiple-Services-Credit-Control.Granted-Service-Unit.CC-Total-Octets","type": "*constant","value": "9999999"},
          {"tag": "GrantedUsage","path": "*rep.Multiple-Services-Credit-Control.Validity-Time","type": "*constant","value": "600"}
        ]
      },

      {
        "id": "data_update",
        "filters": ["*string:~*vars.*cmd:CCR","*string:~*req.CC-Request-Type:2"],
        "flags": ["*update", "*accounts", "*log"],
        "request_fields": [
          {"tag": "ToR","path": "*cgreq.ToR","type": "*constant","value": "*data"},
          {"tag": "Tenant","path": "*cgreq.Tenant","type": "*constant","value": "cgrates.org"},
          {"tag": "OriginID","path": "*cgreq.OriginID","type": "*variable","value": "~*req.Session-Id","mandatory": true},
          {"tag": "Destination","path": "*cgreq.Destination","type": "*variable","value": "~*req.Service-Information.PS-Information.3GPP-SGSN-MCC-MNC","mandatory": true},
          {"tag": "OriginHost","path": "*cgreq.OriginHost","type": "*variable","value": "~*req.Origin-Host"},
          {"tag": "RequestType","path": "*cgreq.RequestType","type": "*constant","value": "*prepaid"},
          {"tag": "Category","path": "*cgreq.Category","type": "*constant","value": "data"},
          {"tag": "Account","path": "*cgreq.Account","type": "*variable","value": "~*req.Subscription-Id.Subscription-Id-Data[~Subscription-Id-Type(0)]","mandatory": true},
          {"tag": "SetupTime","path": "*cgreq.SetupTime","type": "*variable","value": "~*req.Event-Timestamp","mandatory": true},
          {"tag": "AnswerTime","path": "*cgreq.AnswerTime","type": "*variable","value": "~*req.Event-Timestamp","mandatory": true},
          {"tag": "Usage","path": "*cgreq.Usage","type": "*sum","value": "~*req.Multiple-Services-Credit-Control.Used-Service-Unit.CC-Input-Octets;~*req.Multiple-Services-Credit-Control.Used-Service-Unit.CC-Output-Octets"}
        ],
        "reply_fields": [
          {"tag": "CCATemplate","type": "*template","value": "*cca"},
          {"tag": "RatingGroup","path": "*rep.Multiple-Services-Credit-Control.Rating-Group","type": "*constant","value": "1"},
          {"tag": "CCTotalOctets1","path": "*rep.Multiple-Services-Credit-Control.Granted-Service-Unit.CC-Total-Octets","type": "*constant","value": "9999999"},
          {"tag": "GrantedUsage","path": "*rep.Multiple-Services-Credit-Control.Validity-Time","type": "*constant","value": "600"},
        ]
      },

      {
        "id": "data_terminate",
        "filters": ["*string:~*vars.*cmd:CCR","*string:~*req.CC-Request-Type:3"],
        "flags": ["*terminate", "*accounts", "*log"],
        "request_fields": [
          {"tag": "ToR","path": "*cgreq.ToR","type": "*constant","value": "*data"},
          {"tag": "Tenant","path": "*cgreq.Tenant","type": "*constant","value": "cgrates.org"},
          {"tag": "OriginID","path": "*cgreq.OriginID","type": "*variable","value": "~*req.Session-Id","mandat ory": true},
          {"tag": "OriginHost","path": "*cgreq.OriginHost","type": "*variable","value": "~*req.Origin-Host"},
          {"tag": "RequestType","path": "*cgreq.RequestType","type": "*constant","value": "*prepaid"},
          {"tag": "Category","path": "*cgreq.Category","type": "*constant","value": "data"},
          {"tag": "Account","path": "*cgreq.Account","type": "*variable","value": "~*req.Subscription-Id.Subscription-Id-Data[~Subscription-Id-Type(0)]","mandatory": true},
          {"tag": "Destination","path": "*cgreq.Destination","type": "*variable","value": "~*req.Service-Information.PS-Information.3GPP-SGSN-MCC-MNC","mandatory": true}
        ],
        "reply_fields": [
          {"tag": "CCATemplate","type": "*template","value": "*cca"},
          {"tag": "ResultCode","path": "*rep.Result-Code","type": "*constant","value": "2001"}
        ]
      },
      {
        "id": "Sy_Initial_SLR",
        "filters": ["*string:~*vars.*cmd:SLR","*string:~*req.SL-Request-Type:0" // initial request
        ],
        "flags": ["*log", "*initiate"],
        "request_fields": [
          {"tag": "OriginID","path": "*cgreq.OriginID","type": "*variable","value": "~*req.Session-Id","mandatory": true},
          {"tag": "OriginHost","path": "*cgreq.OriginHost","type": "*variable","value": "~*req.Origin-Host","mandatory": true},
          {"tag": "OriginRealm","path": "*cgreq.OriginRealm","type": "*variable","value": "~*req.Origin-Realm","mandatory": true},
          {"tag": "Account","path": "*cgreq.Account","type": "*variable","mandatory": true,"value": "~*req.Subscription-Id.Subscription-Id-Data[~Subscription-Id-Type(0)]"},
          {"tag": "RequestType","path": "*cgreq.RequestType","type": "*constant","value": "*sy"},
		  // The balance Value which when under it, it will trigger the SNR
          {"tag": "BalanceValuePolicyFilter","path": "*opts.*syPolicyFilters.THD_THROTTLE","type": "*group","value": "*lte:~*asm.BalanceSummaries.balance_data.Value:1023"}
        ],
        "reply_fields": [
		  // error if session exists
          {"tag": "FailedAVP","path": "*rep.Failed-AVP","filters": ["*string:~*cgrep.Error:EXISTS"],"type": "*variable","value": "~*req.SL-Request-Type"},
		  // error if session exists
          {"tag": "ResultCode", "path": "*rep.Result-Code","filters": ["*string:~*cgrep.Error:EXISTS"],"type": "*constant","value": "5004","blocker": true},  // DIAMETER_INVALID_AVP_VALUE
          {"tag": "Session-Id","path": "*rep.Session-Id","type": "*variable","value": "~*req.Session-Id","mandatory": true},
          {"tag": "Auth-Application-Id","path": "*rep.Auth-Application-Id","type": "*variable","value": "~*req.Auth-Application-Id","mandatory": true},
          {"tag": "Origin-Host","path": "*rep.Origin-Host","type": "*variable","value": "~*req.Origin-Host","mandatory": true},
          {"tag": "Origin-Realm","path": "*rep.Origin-Realm","type": "*variable","value": "~*req.Origin-Realm","mandatory": true},
          {"tag": "ResultCode","path": "*rep.Result-Code","type": "*constant","value": "2001"},
          {"tag": "Policy-Counter-Identifier","path": "*rep.Policy-Counter-Status-Report.Policy-Counter-Identifier","type": "*group","value": "Monthly","new_branch": true},
          {"tag": "Policy-Counter-Status","path": "*rep.Policy-Counter-Status-Report.Policy-Counter-Status","type": "*group","value": "30GB"},
          {"tag": "Pending-Policy-Counter-Information-Status","path": "*rep.Policy-Counter-Status-Report.Pending-Policy-Counter-Information.Policy-Counter-Status","type": "*group","value": "30GB"},
          {"tag": "Pending-Policy-Counter-Information-Status-Change-Time","path": "*rep.Policy-Counter-Status-Report.Pending-Policy-Counter-Information.Pending-Policy-Counter-Change-Time","type": "*datetime","value": "*now"}
        ]
      },
      {
        "id": "Sy_Intermediate_SLR",
        "filters": [
          "*string:~*vars.*cmd:SLR","*string:~*req.SL-Request-Type:1" // intermediate request
        ],
        "flags": ["*log", "*update"],
        "request_fields": [
          {"tag": "OriginID","path": "*cgreq.OriginID","type": "*variable","value": "~*req.Session-Id","mandatory": true},
          {"tag": "OriginHost","path": "*cgreq.OriginHost","type": "*variable","value": "~*req.Origin-Host","mandatory": true},
          {"tag": "OriginRealm","path": "*cgreq.OriginRealm","type": "*variable","value": "~*req.Origin-Realm","mandatory": true},
          {"tag": "Account","path": "*cgreq.Account","type": "*variable","mandatory": true,"value": "~*req.Subscription-Id.Subscription-Id-Data[~Subscription-Id-Type(0)]"},
          {"tag": "RequestType","path": "*cgreq.RequestType","type": "*constant","value": "*sy"},
		  // The balance Value which when under it, it will trigger the SNR
          {"tag": "BalanceValuePolicyFilter", "path": "*opts.*syPolicyFilters.THD_THROTTLE","type": "*group","value": "*lte:~*asm.BalanceSummaries.balance_data.Value:-1"}
        ],
        "reply_fields": [
          {"tag": "Session-Id","path": "*rep.Session-Id","type": "*variable","value": "~*req.Session-Id"},
          {"tag": "Auth-Application-Id","path": "*rep.Auth-Application-Id","type": "*variable","value": "~*req.Auth-Application-Id"},
          {"tag": "Origin-Host","path": "*rep.Origin-Host","type": "*variable","value": "~*req.Origin-Host"},
          {"tag": "Origin-Realm","path": "*rep.Origin-Realm","type": "*variable","value": "~*req.Origin-Realm"},
          {"tag": "ResultCode","path": "*rep.Result-Code","type": "*constant","value": "2001"},
          {"tag": "Policy-Counter-Identifier","path": "*rep.Policy-Counter-Status-Report.Policy-Counter-Identifier","type": "*group","value": "Monthly","new_branch": true},
          {"tag": "Policy-Counter-Status","path": "*rep.Policy-Counter-Status-Report.Policy-Counter-Status","type": "*group","value": "30GB"},
          {"tag": "Pending-Policy-Counter-Information-Status","path": "*rep.Policy-Counter-Status-Report.Pending-Policy-Counter-Information.Policy-Counter-Status","type": "*group","value": "30GB"},
          {"tag": "Pending-Policy-Counter-Information-Status-Change-Time","path": "*rep.Policy-Counter-Status-Report.Pending-Policy-Counter-Information.Pending-Policy-Counter-Change-Time","type": "*datetime","value": "*now"}
        ]
      },
      {
        "id": "Sy_Terminate_STR",
        "filters": ["*string:~*vars.*cmd:STR"],
        "flags": ["*log", "*terminate"],
        "request_fields": [
          {"tag": "OriginID","path": "*cgreq.OriginID","type": "*variable","value": "~*req.Session-Id","mandatory": true},
          {"tag": "OriginHost","path": "*cgreq.OriginHost","type": "*variable","value": "~*req.Origin-Host","mandatory": true},
          {"tag": "OriginRealm","path": "*cgreq.OriginRealm","type": "*variable","value": "~*req.Origin-Realm","mandatory": true},
          {"tag": "RequestType","path": "*cgreq.RequestType","type": "*constant","value": "*sy"}
        ],
        "reply_fields": [
          {"tag": "Session-Id","path": "*rep.Session-Id","type": "*variable","value": "~*req.Session-Id"},
          {"tag": "Origin-Host","path": "*rep.Origin-Host","type": "*variable","value": "~*req.Origin-Host"},
          {"tag": "Origin-Realm","path": "*rep.Origin-Realm","type": "*variable","value": "~*req.Origin-Realm"},
          {"tag": "ResultCode","path": "*rep.Result-Code","type": "*constant","value": "2001"}
        ]
      }
    ]
  },
  "templates": {
    "*cca": [
      {"tag": "SessionId","path": "*rep.Session-Id","type": "*variable","value": "~*req.Session-Id","mandatory": true},
      {"tag": "ResultCode","path": "*rep.Result-Code","type": "*constant","value": "2001"},
      {"tag": "OriginHost","path": "*rep.Origin-Host","type": "*variable","value": "~*vars.OriginHost","mandatory": true},
      {"tag": "OriginRealm","path": "*rep.Origin-Realm","type": "*variable","value": "~*vars.OriginRealm","mandatory": true},
      {"tag": "AuthApplicationId","path": "*rep.Auth-Application-Id","type": "*variable","value": "~*vars.*appid","mandatory": true},
      {"tag": "CCRequestType","path": "*rep.CC-Request-Type","type": "*variable","value": "~*req.CC-Request-Type","mandatory": true},
      {"tag": "CCRequestNumber","path": "*rep.CC-Request-Number","type": "*variable","value": "~*req.CC-Request-Number","mandatory": true}
    ]
  }
}
`,
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
		LogBuffer: &bytes.Buffer{},
	}
	switch *utils.DBType {
	case utils.MetaInternal:
		if err := os.MkdirAll("/tmp/internal_db", 0755); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := os.RemoveAll("/tmp/internal_db"); err != nil {
				t.Error(err)
			}
		})
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}
	client, cfg := ng.Run(t)

	time.Sleep(100 * time.Millisecond) // wait for DiameterAgent service to start

	// Start monitoring SL
	diamClientSy, err := NewDiameterClient(cfg.DiameterAgentCfg().Listeners[0].Address, cfg.DiameterAgentCfg().OriginHost,
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
	} else if tIDs[0] != syThresholdID(syOriginID, "CGR-DA", "THD_THROTTLE") {
		t.Errorf("expected <%v>, received <%v>", syThresholdID(syOriginID, "CGR-DA", "THD_THROTTLE"), tIDs[0])
	}
	thresholdID := syThresholdID(syOriginID, "CGR-DA", "THD_THROTTLE")
	expTp := &engine.ThresholdProfile{
		Tenant: "cgrates.org",
		ID:     thresholdID,
		FilterIDs: []string{
			"*lte:~*asm.BalanceSummaries.balance_data.Value:1023",
			"*string:~*req.AccountID:1001",
		},
		MaxHits:   1,
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
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	engine.WaitForService(t, ctx, client, utils.SessionSv1)

	diamClientRo, err := NewDiameterClient(cfg.DiameterAgentCfg().Listeners[0].Address, cfg.DiameterAgentCfg().OriginHost,
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
					diam.NewAVP(avp.CCTotalOctets, avp.Mbit, 0, datatype.Unsigned64(5000)),  // CC-Total-Octets
					diam.NewAVP(avp.CCInputOctets, avp.Mbit, 0, datatype.Unsigned64(2500)),  // CC-Input-Octets
					diam.NewAVP(avp.CCOutputOctets, avp.Mbit, 0, datatype.Unsigned64(2500)), // CC-Output-Octets
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
	expLog := fmt.Sprintf("<ThresholdS> processEvent failed to find ClientConnID for ThresholdProfileID <%v>", thresholdID)
	if !strings.Contains(fmt.Sprintln(ng.LogBuffer), expLog) {
		t.Errorf("expected log warning <%v>,\nLogs \n<%v>", expLog, ng.LogBuffer)
	}

}

// multiple listeners will be used to open multiple SLR and CCR on the same account, with a restart in between the CCR-I
func TestDiamSyMultipleListenersRestart(t *testing.T) {

	ng := engine.TestEngine{
		TpFiles: map[string]string{
			utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,PACKAGE_1001,,,`,
			utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
PACKAGE_1001,ACT_TOPUP,*asap,10`,
			utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP,*topup_reset,,,balance_data,*data,,,,,*unlimited,,100000,,,,`,
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
		LogBuffer: &bytes.Buffer{},
	}
	switch *utils.DBType {
	case utils.MetaInternal:
		ng.ConfigPath = filepath.Join(*utils.DataDir, "conf", "samples", "diam_sy_multiple_listeners")
		if err := os.MkdirAll("/tmp/internal_db", 0755); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := os.RemoveAll("/tmp/internal_db"); err != nil {
				t.Error(err)
			}
		})
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	client, cfg := ng.Run(t)
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	engine.WaitForService(t, ctx, client, utils.SessionSv1)
	time.Sleep(100 * time.Millisecond) // wait for DiameterAgent service to start

	// Start monitoring SL
	syOriginIDs := []string{utils.UUIDSha1Prefix(), utils.UUIDSha1Prefix(), utils.UUIDSha1Prefix()}
	slices.Sort(syOriginIDs)
	diamClientSyList := make([]*DiameterClient, len(cfg.DiameterAgentCfg().Listeners))
	diamClientRoList := make([]*DiameterClient, len(cfg.DiameterAgentCfg().Listeners))
	wg := sync.WaitGroup{} // wait for responses before starting to debit from CCR
	for i := range cfg.DiameterAgentCfg().Listeners {
		// start monitoring SL concurrently
		wg.Go(func() {
			diamClientRoList[i], err = NewDiameterClient(cfg.DiameterAgentCfg().Listeners[i].Address, cfg.DiameterAgentCfg().OriginHost,
				cfg.DiameterAgentCfg().OriginRealm, cfg.DiameterAgentCfg().VendorID,
				cfg.DiameterAgentCfg().ProductName, utils.DiameterFirmwareRevision,
				cfg.DiameterAgentCfg().DictionariesPath, cfg.DiameterAgentCfg().Listeners[i].Network)
			if err != nil {
				t.Fatal(err)
			}
			diamClientSyList[i], err = NewDiameterClient(cfg.DiameterAgentCfg().Listeners[i].Address, "Sy-CGR-DA"+strconv.Itoa(i),
				cfg.DiameterAgentCfg().OriginRealm, cfg.DiameterAgentCfg().VendorID,
				cfg.DiameterAgentCfg().ProductName, utils.DiameterFirmwareRevision,
				cfg.DiameterAgentCfg().DictionariesPath, cfg.DiameterAgentCfg().Listeners[i].Network)
			if err != nil {
				t.Fatal(err)
			}
			slr := diam.NewRequest(diam.SpendingLimit, 16777302, nil)
			slr.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(syOriginIDs[i]))
			slr.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("Sy-CGR-DA"+strconv.Itoa(i)))
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
			if err := diamClientSyList[i].SendMessage(slr); err != nil {
				t.Errorf("failed to send diameter message: %v", err)
			}

			reply := diamClientSyList[i].ReceivedMessage(2 * time.Second)
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

		})
	}
	wg.Wait()
	expBalance := float64(100000)
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
	if len(replyActSess) != 3 {
		t.Errorf("expected 3 active sessions, received <%v>", utils.ToJSON(replyActSess))
	}
	// t.Log(utils.ToIJSON(replyActSess))

	// get thresholds
	for i, syOriginID := range syOriginIDs {
		var tIDs []string
		if err := client.Call(context.Background(), utils.ThresholdSv1GetThresholdIDs,
			&utils.TenantWithAPIOpts{}, &tIDs); err != nil {
			t.Error(err)
		}
		if len(tIDs) != 3 {
			t.Errorf("expected 3 threshold profiles to be created, received <%v>", tIDs)
		}
		slices.Sort(tIDs)
		if tIDs[i] != syThresholdID(syOriginID, "Sy-CGR-DA"+strconv.Itoa(i), "THD_THROTTLE") {
			t.Errorf("expected <%v>, received <%v>", syThresholdID(syOriginID, "Sy-CGR-DA"+strconv.Itoa(i), "THD_THROTTLE"), tIDs[i])
		}
		expTp := &engine.ThresholdProfile{
			Tenant: "cgrates.org",
			ID:     syThresholdID(syOriginID, "Sy-CGR-DA"+strconv.Itoa(i), "THD_THROTTLE"),
			FilterIDs: []string{
				utils.ConcatenatedKey(utils.MetaLessOrEqual, utils.DynamicDataPrefix+utils.MetaAsm+utils.NestingSep+utils.BalanceSummaries+utils.NestingSep+"balance_data"+utils.NestingSep+utils.Value, "85000"),
				utils.ConcatenatedKey(utils.MetaString, utils.MetaDynReq+utils.NestingSep+utils.AccountID, "1001"),
			},
			MaxHits:   1,
			Async:     true,
			ActionIDs: []string{utils.MetaSyPublish},
		}
		tp := &engine.ThresholdProfile{}
		if err := client.Call(context.Background(), utils.APIerSv1GetThresholdProfile,
			&utils.TenantID{ID: tIDs[i]}, &tp); err != nil {
			t.Error(err)
		}
		expTp.ActivationInterval = tp.ActivationInterval
		slices.Sort(tp.FilterIDs) // sort filters received since they are processed from a map
		if !reflect.DeepEqual(tp, expTp) {
			t.Errorf("expected <%v>\nreceived\n<%v>", utils.ToJSON(expTp), utils.ToJSON(tp))
		}
	}

	// restart engine
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
	// fmt.Println(ng.LogBuffer)
	// ng.LogBuffer = &bytes.Buffer{}
	// t.Cleanup(func() { fmt.Println(ng.LogBuffer) })
	ng.PreserveDataDB = true
	client, cfg = ng.Run(t)
	ctx, cancel = context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	engine.WaitForService(t, ctx, client, utils.SessionSv1)
	time.Sleep(100 * time.Millisecond) // wait for DiameterAgent service since its waiting on sessions
	for i := range diamClientRoList {
		diamClientRoList[i], err = NewDiameterClient(cfg.DiameterAgentCfg().Listeners[i].Address, cfg.DiameterAgentCfg().OriginHost,
			cfg.DiameterAgentCfg().OriginRealm, cfg.DiameterAgentCfg().VendorID,
			cfg.DiameterAgentCfg().ProductName, utils.DiameterFirmwareRevision,
			cfg.DiameterAgentCfg().DictionariesPath, cfg.DiameterAgentCfg().Listeners[i].Network)
		if err != nil {
			t.Fatal(err)
		}
		diamClientSyList[i], err = NewDiameterClient(cfg.DiameterAgentCfg().Listeners[i].Address, "Sy-CGR-DA"+strconv.Itoa(i),
			cfg.DiameterAgentCfg().OriginRealm, cfg.DiameterAgentCfg().VendorID,
			cfg.DiameterAgentCfg().ProductName, utils.DiameterFirmwareRevision,
			cfg.DiameterAgentCfg().DictionariesPath, cfg.DiameterAgentCfg().Listeners[i].Network)
		if err != nil {
			t.Fatal(err)
		}
	}
	ccrOriginIDs := []string{utils.UUIDSha1Prefix(), utils.UUIDSha1Prefix(), utils.UUIDSha1Prefix()}
	for i := range diamClientSyList {
		wg.Go(func() {

			ccr := diam.NewRequest(diam.CreditControl, 4, nil)
			ccr.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(ccrOriginIDs[i]))
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
			if err := diamClientRoList[i].SendMessage(ccr); err != nil {
				t.Errorf("failed to send diameter message: %v", err)
			}

			reply := diamClientRoList[i].ReceivedMessage(2 * time.Second)
			if reply == nil {
				t.Error("received empty reply")
			}
			// t.Log(reply.PrettyDump())

			avps, err := reply.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID)
			if err != nil {
				t.Error(err)
			}
			if len(avps) == 0 {
				t.Error("missing AVPs in reply")
			}

			resultCode, err := diamAVPAsString(avps[0])
			if err != nil {
				t.Error(err)
			}
			if resultCode != "2001" {
				t.Errorf("Result-Code=%s, want %s", resultCode, "2001")
			}
		})
	}
	wg.Wait()
	expBalance = float64(94000) // CCR init should take 2000 units constant from templates, 3 times since we have 3 clients. initial balance is 100000, 100000 - 6000 = 94000
	if err = client.Call(context.Background(), utils.APIerSv2GetAccount, attrsAcnt, &acnt); err != nil {
		t.Errorf("APIerSv1.GetAccount unexpected err: %v", err)
	} else if rply := acnt.BalanceMap[utils.MetaData].GetTotalValue(); rply != expBalance {
		t.Errorf("APIerSv1.GetAccount: data_balance: %f, want: %f", rply, expBalance)
	}
	// find indexed sy sessions
	if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions, utils.SessionFilter{Filters: []string{"*string:~*req.RequestType:*sy"}}, &replyActSess); err != nil {
		t.Fatal(err)
	}
	if len(replyActSess) != 3 {
		t.Errorf("expected 3 active sessions, received <%v>", utils.ToJSON(replyActSess))
	}
	// t.Log(utils.ToIJSON(replyActSess))
	// find all active sessions
	if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions, utils.SessionFilter{}, &replyActSess); err != nil {
		t.Error(err)
	}
	if len(replyActSess) != 6 {
		t.Errorf("expected 6 active sessions, received <%v>", utils.ToJSON(replyActSess))
	}
	// t.Log(utils.ToIJSON(replyActSess))

	for i := range diamClientSyList {
		wg.Go(func() {
			ccrU := diam.NewRequest(diam.CreditControl, 4, nil)
			ccrU.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(ccrOriginIDs[i]))
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
							diam.NewAVP(avp.CCTotalOctets, avp.Mbit, 0, datatype.Unsigned64(5000)),  // CC-Total-Octets
							diam.NewAVP(avp.CCInputOctets, avp.Mbit, 0, datatype.Unsigned64(2500)),  // CC-Input-Octets
							diam.NewAVP(avp.CCOutputOctets, avp.Mbit, 0, datatype.Unsigned64(2500)), // CC-Output-Octets
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
			if err := diamClientRoList[i].SendMessage(ccrU); err != nil {
				t.Errorf("failed to send diameter message: %v", err)
			}

			reply := diamClientRoList[i].ReceivedMessage(5 * time.Second)
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
				t.Errorf("Result-Code=%s, want %s", resultCode, "2001")
			}
		})
	}
	wg.Wait()
	expBalance = float64(94000 - 5000 - 5000 - 5000) // CCR update should take 5000 units from CCTotalOctets, 3 times since we have 3 clients. leftover balance is 94000, after debits 79000 is left
	if err = client.Call(context.Background(), utils.APIerSv2GetAccount, attrsAcnt, &acnt); err != nil {
		t.Errorf("APIerSv1.GetAccount unexpected err: %v", err)
	} else if rply := acnt.BalanceMap[utils.MetaData].GetTotalValue(); rply != expBalance {
		t.Errorf("APIerSv1.GetAccount: data_balance: %f, want: %f", rply, expBalance)
	}

	// find indexed sy sessions
	if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions, utils.SessionFilter{Filters: []string{"*string:~*req.RequestType:*sy"}}, &replyActSess); err != nil {
		t.Fatal(err)
	}
	if len(replyActSess) != 3 {
		t.Errorf("expected 3 active sessions, received <%v>", utils.ToJSON(replyActSess))
	}
	if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions, utils.SessionFilter{}, &replyActSess); err != nil {
		t.Error(err)
	}
	if len(replyActSess) != 6 {
		t.Errorf("expected 6 active sessions, received <%v>", utils.ToJSON(replyActSess))
	}
	// t.Log(utils.ToIJSON(replyActSess))

	// receive SNR
	for i := range diamClientSyList {
		wg.Go(func() {
			// fmt.Println("waiting for SNR msg from listener ", cfg.DiameterAgentCfg().Listeners[i].Address)
			reply := diamClientSyList[i].ReceivedMessage(2 * time.Second)
			if reply == nil {
				t.Fatal("Received empty reply")
				return
			}
			// fmt.Println("got SNR msg from listener ", cfg.DiameterAgentCfg().Listeners[i].Address)
			rplSID, err := reply.FindAVP(avp.SessionID, 0)
			if err != nil {
				t.Error(err)
			}
			rplSIDStr, err := diamAVPAsString(rplSID)
			if err != nil {
				t.Error(err)
			}
			// check if Session-Id in reply is contained in the list of syOriginIDs
			if !slices.Contains(syOriginIDs, rplSIDStr) {
				t.Errorf("Expected Session-Id to be one of <%v> received <%v>", syOriginIDs, rplSIDStr)
			}

			expected := fmt.Sprintf(`Spending-Status-Notification-Request (SNR)
%s
	Session-Id {Code:263,Flags:0x40,Length:16,VendorId:0,Value:UTF8String{%s},Padding:1}
	Origin-Host {Code:264,Flags:0x40,Length:20,VendorId:0,Value:DiameterIdentity{Sy-CGR-DA%s},Padding:2}
	Origin-Realm {Code:296,Flags:0x40,Length:20,VendorId:0,Value:DiameterIdentity{cgrates.org},Padding:1}
	Destination-Realm {Code:283,Flags:0x40,Length:24,VendorId:0,Value:DiameterIdentity{dr-cgrates.org},Padding:2}
	Destination-Host {Code:293,Flags:0x40,Length:20,VendorId:0,Value:DiameterIdentity{CGR-DA-DH},Padding:3}
	Auth-Application-Id {Code:258,Flags:0x40,Length:12,VendorId:0,Value:Unsigned32{16777302}}
	Policy-Counter-Status-Report {Code:2903,Flags:0xc0,Length:56,VendorId:10415,Value:Grouped{
		Policy-Counter-Identifier {Code:2901,Flags:0xc0,Length:20,VendorId:10415,Value:UTF8String{Monthly},Padding:1},
		Policy-Counter-Status {Code:2902,Flags:0xc0,Length:24,VendorId:10415,Value:UTF8String{THD_THROTTLE},Padding:0},
	}}
`, reply.Header, rplSIDStr, strconv.Itoa(i))
			if expected != reply.String() {
				t.Errorf("expected \n<%v>, \nreceived \n<%v>", expected, reply.String())
			}
			sna := diam.NewRequest(SSN, 16777302, nil)
			sna.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(rplSIDStr))
			sna.NewAVP(avp.ResultCode, avp.Mbit, 0, datatype.Unsigned32(2001))
			sna.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("Sy-CGR-DA"+strconv.Itoa(i)))
			sna.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
			// fmt.Println("sendingg msg: ", sna.PrettyDump())
			if err := diamClientSyList[i].SendMessage(sna); err != nil {
				t.Errorf("failed to send diameter message: %v", err)
			}

		})
	}
	wg.Wait()
	// Change session policy subscription
	for i := range diamClientSyList {
		wg.Go(func() {
			slri := diam.NewRequest(diam.SpendingLimit, 16777302, nil)
			slri.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(syOriginIDs[i]))
			slri.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("Sy-CGR-DA"+strconv.Itoa(i)))
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

			if err := diamClientSyList[i].SendMessage(slri); err != nil {
				t.Errorf("failed to send diameter message: %v", err)
			}

			reply := diamClientSyList[i].ReceivedMessage(2 * time.Second)
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
		})
	}
	wg.Wait()
	expBalance = float64(79000) // no debits shouldve been done
	if err = client.Call(context.Background(), utils.APIerSv2GetAccount, attrsAcnt, &acnt); err != nil {
		t.Errorf("APIerSv1.GetAccount unexpected err: %v", err)
	} else if rply := acnt.BalanceMap[utils.MetaData].GetTotalValue(); rply != expBalance {
		t.Errorf("APIerSv1.GetAccount: data_balance: %f, want: %f", rply, expBalance)
	}

	if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions, utils.SessionFilter{Filters: []string{"*string:~*req.RequestType:*sy"}}, &replyActSess); err != nil {
		t.Error(err)
	}
	if len(replyActSess) != 3 {
		t.Errorf("expected 3 active sessions, received <%v>", utils.ToJSON(replyActSess))
	}
	// t.Log(utils.ToIJSON(replyActSess))

	// get thresholds
	for i, syOriginID := range syOriginIDs {
		var tIDs []string
		if err := client.Call(context.Background(), utils.ThresholdSv1GetThresholdIDs,
			&utils.TenantWithAPIOpts{}, &tIDs); err != nil {
			t.Error(err)
		}
		if len(tIDs) != 3 {
			t.Errorf("expected 3 threshold profiles to be created, received <%v>", tIDs)
		}
		slices.Sort(tIDs)
		if tIDs[i] != syThresholdID(syOriginID, "Sy-CGR-DA"+strconv.Itoa(i), "NEVER_THROTTLE") {
			t.Errorf("expected <%v>, received <%v>", syThresholdID(syOriginID, "Sy-CGR-DA"+strconv.Itoa(i), "NEVER_THROTTLE"), tIDs[i])
		}
		expTp := &engine.ThresholdProfile{
			Tenant: "cgrates.org",
			ID:     syThresholdID(syOriginID, "Sy-CGR-DA"+strconv.Itoa(i), "NEVER_THROTTLE"),
			FilterIDs: []string{
				utils.ConcatenatedKey(utils.MetaLessOrEqual, utils.DynamicDataPrefix+utils.MetaAsm+utils.NestingSep+utils.BalanceSummaries+utils.NestingSep+"balance_data"+utils.NestingSep+utils.Value, "-1"),
				utils.ConcatenatedKey(utils.MetaString, utils.MetaDynReq+utils.NestingSep+utils.AccountID, "1001"),
			},
			MaxHits:   1,
			Async:     true,
			ActionIDs: []string{utils.MetaSyPublish},
		}
		tp := &engine.ThresholdProfile{}
		if err := client.Call(context.Background(), utils.APIerSv1GetThresholdProfile,
			&utils.TenantID{ID: tIDs[i]}, &tp); err != nil {
			t.Error(err)
		}
		expTp.ActivationInterval = tp.ActivationInterval
		slices.Sort(tp.FilterIDs) // sort filters received since they are processed from a map
		if !reflect.DeepEqual(tp, expTp) {
			t.Errorf("expected <%v>\nreceived\n<%v>", utils.ToJSON(expTp), utils.ToJSON(tp))
		}
	}

	// Send CCT
	for i := range diamClientSyList {
		wg.Go(func() {
			ccrT := diam.NewRequest(diam.CreditControl, 4, nil)
			ccrT.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(ccrOriginIDs[i]))
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
							diam.NewAVP(avp.CCTotalOctets, avp.Mbit, 0, datatype.Unsigned64(5000)),  // CC-Total-Octets
							diam.NewAVP(avp.CCInputOctets, avp.Mbit, 0, datatype.Unsigned64(2500)),  // CC-Input-Octets
							diam.NewAVP(avp.CCOutputOctets, avp.Mbit, 0, datatype.Unsigned64(2500)), // CC-Output-Octets
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
			if err := diamClientRoList[i].SendMessage(ccrT); err != nil {
				t.Errorf("failed to send diameter message: %v", err)
			}

			reply := diamClientRoList[i].ReceivedMessage(2 * time.Second)
			if reply == nil {
				t.Fatal("received empty reply")
			}
			// t.Log(reply.PrettyDump())

			avps, err := reply.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID)
			if err != nil {
				t.Error(err)
			}
			if len(avps) == 0 {
				t.Error("missing AVPs in reply")
			}

			resultCode, err := diamAVPAsString(avps[0])
			if err != nil {
				t.Error(err)
			}
			if resultCode != "2001" {
				t.Errorf("Result-Code=%s, want %s", resultCode, "2001")
			}
		})
	}
	wg.Wait()
	expBalance = float64(79000)
	if err = client.Call(context.Background(), utils.APIerSv2GetAccount, attrsAcnt, &acnt); err != nil {
		t.Errorf("APIerSv1.GetAccount unexpected err: %v", err)
	} else if rply := acnt.BalanceMap[utils.MetaData].GetTotalValue(); rply != expBalance {
		t.Errorf("APIerSv1.GetAccount: data_balance: %f, want: %f", rply, expBalance)
	}

	// find indexed sy sessions
	if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions, utils.SessionFilter{Filters: []string{"*string:~*req.RequestType:*sy"}}, &replyActSess); err != nil {
		t.Error(err)
	}
	if len(replyActSess) != 3 {
		t.Errorf("expected 3 active sessions, received <%v>", utils.ToJSON(replyActSess))
	}
	if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions, utils.SessionFilter{}, &replyActSess); err != nil {
		t.Error(err)
	}
	if len(replyActSess) != 3 {
		t.Errorf("expected 3 active sessions, received <%v>", utils.ToJSON(replyActSess))
	}
	// t.Log(utils.ToIJSON(replyActSess))
	for i := range diamClientSyList {
		wg.Go(func() {
			str := diam.NewRequest(diam.SessionTermination, 16777302, nil)
			str.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(syOriginIDs[i]))
			str.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("Sy-CGR-DA"+strconv.Itoa(i)))
			str.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
			str.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA-DH"))
			str.NewAVP(avp.DestinationRealm, avp.Mbit, 0, datatype.DiameterIdentity("dr-cgrates.org"))
			str.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(16777302))
			str.NewAVP(avp.TerminationCause, avp.Mbit, 0, datatype.Enumerated(1))
			// t.Log("sendingg STR msg: ", str.PrettyDump())
			if err := diamClientSyList[i].SendMessage(str); err != nil {
				t.Errorf("failed to send diameter message: %v", err)
			}

			reply := diamClientSyList[i].ReceivedMessage(2 * time.Second)
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
		})
	}
	wg.Wait()
	expBalance = float64(79000)
	if err = client.Call(context.Background(), utils.APIerSv2GetAccount, attrsAcnt, &acnt); err != nil {
		t.Errorf("APIerSv1.GetAccount unexpected err: %v", err)
	} else if rply := acnt.BalanceMap[utils.MetaData].GetTotalValue(); rply != expBalance {
		t.Errorf("APIerSv1.GetAccount: data_balance: %f, want: %f", rply, expBalance)
	}
	if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions, utils.SessionFilter{}, &replyActSess); err == nil || err.Error() != "NOT_FOUND" {
		t.Errorf("expected error <NOT_FOUND>, received <%v>", err)
	}
	// thresholds profile should be removed on STR
	var tIDs []string
	if err := client.Call(context.Background(), utils.ThresholdSv1GetThresholdIDs,
		&utils.TenantWithAPIOpts{}, &tIDs); err != nil {
		t.Error(err)
	} else if len(tIDs) != 0 {
		t.Errorf("expected no Threshold profiles, received <%v>", tIDs)
	}
}
