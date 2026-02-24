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
	"bytes"
	"fmt"
	"path/filepath"
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

func TestDiamSy(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	ng := engine.TestEngine{
		ConfigPath: filepath.Join(*utils.DataDir, "conf", "samples", "diam_sy"),
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
		DBCfg:     engine.InternalDBCfg,
		LogBuffer: &bytes.Buffer{},
	}
	t.Cleanup(func() { fmt.Println(ng.LogBuffer) })
	client, cfg := ng.Run(t)

	time.Sleep(100 * time.Millisecond) // wait for DiameterAgent service to start

	// Start monitoring SL after CCA successful
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
	if len(replyActSess) != 2 {
		t.Errorf("expected 2 active sessions, received <%v>", replyActSess)
	}
	// t.Log(utils.ToIJSON(replyActSess))

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
}
