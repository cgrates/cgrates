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
	"path/filepath"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/go-diameter/diam"
	"github.com/cgrates/go-diameter/diam/avp"
	"github.com/cgrates/go-diameter/diam/datatype"
	"github.com/cgrates/go-diameter/diam/dict"
)

func TestDiamBalanceEnquiry(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	ng := engine.TestEngine{
		ConfigPath: filepath.Join(*utils.DataDir, "conf", "samples", "diamagent_balance_enq"),
		TpFiles:    map[string]string{},
		DBCfg:      engine.InternalDBCfg,
		LogBuffer:  &bytes.Buffer{},
	}
	client, cfg := ng.Run(t)
	time.Sleep(200 * time.Millisecond)

	var reply string
	if err := client.Call(context.Background(), utils.APIerSv2SetBalance,
		utils.AttrSetBalance{
			Tenant:      "cgrates.com",
			Account:     "1001",
			Value:       50,
			BalanceType: utils.MetaData,
			Balance: map[string]any{
				utils.ID: "balance1",
			},
		}, &reply); err != nil {
		t.Fatal(err)
	}
	if err := client.Call(context.Background(), utils.APIerSv2SetBalance,
		utils.AttrSetBalance{
			Tenant:      "cgrates.com",
			Account:     "1001",
			Value:       100,
			BalanceType: utils.MetaData,
			Balance: map[string]any{
				utils.ID: "balance2",
			},
		}, &reply); err != nil {
		t.Fatal(err)
	}
	alsPrfs := []*engine.AttributeProfileWithAPIOpts{
		{
			AttributeProfile: &engine.AttributeProfile{
				Tenant:    "cgrates.org",
				ID:        "ATTR_BALANCE_ENQUIRY",
				FilterIDs: []string{"*string:~*req.Category:sms"},
				Attributes: []*engine.Attribute{
					{
						Path: utils.MetaTenant,
						Type: utils.MetaConstant,
						Value: config.RSRParsers{
							&config.RSRParser{
								Rules: "cgrates.com",
							},
						},
					},
				},
				Weight: 20,
			},
		},
		{
			AttributeProfile: &engine.AttributeProfile{
				Tenant:    "cgrates.com",
				ID:        "ATTR_GET_BALANCE",
				FilterIDs: []string{"*string:~*req.Account:1001"},
				Attributes: []*engine.Attribute{
					{
						Path: utils.MetaReq + utils.NestingSep + "MonBalance",
						Type: utils.MetaVariable,
						Value: config.RSRParsers{
							&config.RSRParser{
								Rules: "~*accounts.1001.BalanceMap.*data.GetTotalValue",
							},
						},
					},
				},
				Weight: 10,
			},
		},
	}
	for _, alsPrf := range alsPrfs {
		alsPrf.Compile()
		if err := client.Call(context.Background(), utils.APIerSv1SetAttributeProfile,
			alsPrf, &reply); err != nil {
			t.Fatal(err)
		}
	}
	time.Sleep(300 * time.Millisecond)
	diamClient, err := NewDiameterClient(cfg.DiameterAgentCfg().Listeners[0].Address, "localhost",
		cfg.DiameterAgentCfg().OriginRealm, cfg.DiameterAgentCfg().VendorID,
		cfg.DiameterAgentCfg().ProductName, utils.DiameterFirmwareRevision,
		cfg.DiameterAgentCfg().DictionariesPath, cfg.DiameterAgentCfg().Listeners[0].Network)
	if err != nil {
		t.Fatal(err)
	}

	sendBalanceEnquiryQueryCCR(t, diamClient, 5*time.Second)
}

func sendBalanceEnquiryQueryCCR(tb testing.TB, client *DiameterClient, replyTimeout time.Duration) {
	tb.Helper()
	sessionID := utils.UUIDSha1Prefix()
	ccr := diam.NewRequest(diam.CreditControl, 4, nil)
	ccr.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(sessionID))
	ccr.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	ccr.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	ccr.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	ccr.NewAVP(avp.ServiceContextID, avp.Mbit, 0, datatype.UTF8String("manager@huawei.com"))
	ccr.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(4))
	ccr.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(1))
	ccr.NewAVP(avp.EventTimestamp, avp.Mbit, 0, datatype.Time(time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)))
	ccr.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(0)),
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("1001")),
		}})
	ccr.NewAVP(avp.ServiceIdentifier, avp.Mbit, 0, datatype.Unsigned32(1))
	ccr.NewAVP(avp.RequestedAction, avp.Mbit, 0, datatype.Enumerated(0))
	ccr.NewAVP(avp.RequestedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.CCServiceSpecificUnits, avp.Mbit, 0, datatype.Unsigned64(1)),
		}})

	if err := client.SendMessage(ccr); err != nil {
		tb.Fatalf("failed to send diameter message: %v", err)
	}

	cca := client.ReceivedMessage(replyTimeout)
	if cca == nil {
		tb.Fatal("received empty reply for query CCR")
	}

	rcAVPs, err := cca.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID)
	if err != nil {
		tb.Fatal(err)
	}
	if resultCode, err := diamAVPAsString(rcAVPs[0]); err != nil {
		tb.Fatal(err)
	} else if resultCode != "2001" {
		tb.Fatalf("CCA Result-Code=%s, want 2001", resultCode)
	}
	balanceAVPs, err := cca.FindAVPsWithPath(
		[]any{"Cost-Information", "Unit-Value", "Value-Digits"}, dict.UndefinedVendorID)
	if err != nil {
		tb.Fatalf("failed to find Value-Digits: %v", err)
	}

	if monetaryVal, err := diamAVPAsString(balanceAVPs[0]); err != nil {
		tb.Fatalf("failed to read monetary balance: %v", err)
	} else if monetaryVal != "150" {
		tb.Fatalf("monetary balance=%s, want 10000", monetaryVal)
	}

}
