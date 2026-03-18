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

package general_tests

import (
	"bytes"
	"fmt"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

func TestGOCSIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaMySQL:
	case utils.MetaInternal, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	usNG := engine.TestEngine{
		ConfigPath: path.Join(*utils.DataDir, "conf", "samples", "gocs", "us_site"),
	}
	usClient, _ := usNG.Run(t)
	auNG := engine.TestEngine{
		ConfigPath: path.Join(*utils.DataDir, "conf", "samples", "gocs", "au_site"),
	}
	auClient, _ := auNG.Run(t)
	dspNG := engine.TestEngine{
		ConfigPath: path.Join(*utils.DataDir, "conf", "samples", "gocs", "dsp_site"),
		PreStartHook: func(t testing.TB, c *config.CGRConfig) {
			buf := &bytes.Buffer{}
			defer fmt.Println(buf)
			engine.LoadCSVsWithCGRLoader(t, c.ConfigPath, path.Join(*utils.DataDir, "tariffplans", "gocs", "dsp_site"), buf, nil, "-caches_address=")
		},
	}
	dspClient, _ := dspNG.Run(t)
	time.Sleep(100 * time.Millisecond)

	t.Run("load data", func(t *testing.T) {
		chargerProfile := &v1.ChargerWithAPIOpts{
			ChargerProfile: &engine.ChargerProfile{
				Tenant:       "cgrates.org",
				ID:           "DEFAULT",
				RunID:        utils.MetaDefault,
				AttributeIDs: []string{utils.MetaNone},
				Weight:       10,
			},
		}
		var result string
		if err := usClient.Call(context.Background(), utils.APIerSv1SetChargerProfile, chargerProfile, &result); err != nil {
			t.Error(err)
		} else if result != utils.OK {
			t.Error("Unexpected reply returned", result)
		}
		var rpl *engine.ChargerProfile
		if err := usClient.Call(context.Background(), utils.APIerSv1GetChargerProfile,
			&utils.TenantID{Tenant: "cgrates.org", ID: "DEFAULT"}, &rpl); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(chargerProfile.ChargerProfile, rpl) {
			t.Errorf("Expecting : %+v, received: %+v", chargerProfile.ChargerProfile, rpl)
		}
		if err := usClient.Call(context.Background(), utils.APIerSv1SetChargerProfile, chargerProfile, &result); err != nil {
			t.Error(err)
		} else if result != utils.OK {
			t.Error("Unexpected reply returned", result)
		}
		if err := usClient.Call(context.Background(), utils.APIerSv1GetChargerProfile,
			&utils.TenantID{Tenant: "cgrates.org", ID: "DEFAULT"}, &rpl); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(chargerProfile.ChargerProfile, rpl) {
			t.Errorf("Expecting : %+v, received: %+v", chargerProfile.ChargerProfile, rpl)
		}

		var acnt *engine.Account
		acntAttrs := &utils.AttrGetAccount{
			Tenant:  "cgrates.org",
			Account: "1001"}
		attrSetBalance := utils.AttrSetBalance{
			Tenant:      acntAttrs.Tenant,
			Account:     acntAttrs.Account,
			BalanceType: utils.MetaVoice,
			Value:       3540000000000,
			Balance: map[string]any{
				utils.ID:     "BALANCE1",
				utils.Weight: 20,
			},
		}
		// add a voice balance of 59 minutes
		var reply string
		if err := usClient.Call(context.Background(), utils.APIerSv1SetBalance, attrSetBalance, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("received: %s", reply)
		}
		expectedVoice := 3540000000000.0
		if err := usClient.Call(context.Background(), utils.APIerSv2GetAccount, acntAttrs, &acnt); err != nil {
			t.Error(err)
		} else if rply := acnt.BalanceMap[utils.MetaVoice].GetTotalValue(); rply != expectedVoice {
			t.Errorf("Expecting: %v, received: %v", expectedVoice, rply)
		}
		time.Sleep(time.Duration(*utils.WaitRater) * time.Millisecond) // Give time for scheduler to execute topups on au_site
	})

	t.Run("auth session", func(t *testing.T) {
		authUsage := 5 * time.Minute
		args := &sessions.V1AuthorizeArgs{
			GetMaxUsage: true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestSSv1ItAuth",
				Event: map[string]any{
					utils.Tenant:       "cgrates.org",
					utils.ToR:          utils.MetaVoice,
					utils.OriginID:     "testGOCS",
					utils.Category:     "call",
					utils.RequestType:  utils.MetaPrepaid,
					utils.AccountField: "1001",
					utils.Subject:      "1001",
					utils.Destination:  "1002",
					utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
					utils.Usage:        authUsage,
				},
			},
		}
		var rply sessions.V1AuthorizeReply
		if err := dspClient.Call(context.Background(), utils.SessionSv1AuthorizeEvent, args, &rply); err != nil {
			t.Fatal(err)
		}
		if rply.MaxUsage == nil || *rply.MaxUsage != authUsage {
			t.Errorf("Unexpected MaxUsage: %v", rply.MaxUsage)
		}
	})

	t.Run("init session", func(t *testing.T) {
		initUsage := 5 * time.Minute
		args := &sessions.V1InitSessionArgs{
			InitSession: true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestSSv1ItInitiateSession",
				Event: map[string]any{
					utils.Tenant:       "cgrates.org",
					utils.ToR:          utils.MetaVoice,
					utils.OriginID:     "testGOCS",
					utils.Category:     "call",
					utils.RequestType:  utils.MetaPrepaid,
					utils.AccountField: "1001",
					utils.Subject:      "1001",
					utils.Destination:  "1002",
					utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
					utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
					utils.Usage:        initUsage,
				},
			},
		}
		var rply sessions.V1InitSessionReply
		if err := dspClient.Call(context.Background(), utils.SessionSv1InitiateSession,
			args, &rply); err != nil {
			t.Fatal(err)
		}
		if rply.MaxUsage == nil || *rply.MaxUsage != initUsage {
			t.Errorf("Unexpected MaxUsage: %v", rply.MaxUsage)
		}
		// give a bit of time to session to be replicate
		time.Sleep(10 * time.Millisecond)

		aSessions := make([]*sessions.ExternalSession, 0)
		if err := usClient.Call(context.Background(), utils.SessionSv1GetActiveSessions, new(utils.SessionFilter), &aSessions); err != nil {
			t.Error(err)
		} else if len(aSessions) != 1 {
			t.Errorf("wrong active sessions: %s \n , and len(aSessions) %+v", utils.ToJSON(aSessions), len(aSessions))
		} else if aSessions[0].NodeID != "US_SITE" {
			t.Errorf("Expecting : %+v, received: %+v", "US_SITE", aSessions[0].NodeID)
		} else if aSessions[0].Usage != 5*time.Minute {
			t.Errorf("Expecting : %+v, received: %+v", 5*time.Minute, aSessions[0].Usage)
		}

		var acnt *engine.Account
		attrAcc := &utils.AttrGetAccount{
			Tenant:  "cgrates.org",
			Account: "1001",
		}

		// 59 mins - 5 mins = 54 mins
		if err := auClient.Call(context.Background(), utils.APIerSv2GetAccount, attrAcc, &acnt); err != nil {
			t.Error(err)
		} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != 3240000000000.0 {
			t.Errorf("Expecting : %+v, received: %+v", 3240000000000.0, acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
		}

		if err := usClient.Call(context.Background(), utils.APIerSv2GetAccount, attrAcc, &acnt); err != nil {
			t.Error(err)
		} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != 3240000000000.0 {
			t.Errorf("Expecting : %+v, received: %+v", 3240000000000.0, acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
		}
	})

	t.Run("update session", func(t *testing.T) {
		reqUsage := 5 * time.Minute
		args := &sessions.V1UpdateSessionArgs{
			UpdateSession: true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestSSv1ItUpdateSession",
				Event: map[string]any{
					utils.Tenant:       "cgrates.org",
					utils.ToR:          utils.MetaVoice,
					utils.OriginID:     "testGOCS",
					utils.Category:     "call",
					utils.RequestType:  utils.MetaPrepaid,
					utils.AccountField: "1001",
					utils.Subject:      "1001",
					utils.Destination:  "1002",
					utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
					utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
					utils.Usage:        reqUsage,
				},
			},
		}
		var rply sessions.V1UpdateSessionReply

		// right now dispatcher receive utils.ErrPartiallyExecuted
		// in case of of engines fails
		if err := auClient.Call(context.Background(), utils.SessionSv1UpdateSession, args, &rply); err != nil {
			t.Errorf("Expecting : %+v, received: %+v", utils.ErrPartiallyExecuted, err)
		}

		aSessions := make([]*sessions.ExternalSession, 0)
		if err := auClient.Call(context.Background(), utils.SessionSv1GetActiveSessions, new(utils.SessionFilter), &aSessions); err != nil {
			t.Error(err)
		} else if len(aSessions) != 1 {
			t.Errorf("wrong active sessions: %s", utils.ToJSON(aSessions))
		} else if aSessions[0].NodeID != "AU_SITE" {
			t.Errorf("Expecting : %+v, received: %+v", "AU_SITE", aSessions[0].NodeID)
		} else if aSessions[0].Usage != 5*time.Minute {
			t.Errorf("Expecting : %+v, received: %+v", 5*time.Minute, aSessions[0].Usage)
		}

		var acnt *engine.Account
		attrAcc := &utils.AttrGetAccount{
			Tenant:  "cgrates.org",
			Account: "1001",
		}

		// balanced changed in AU_SITE
		// 54 min - 5 mins = 49 min
		if err := auClient.Call(context.Background(), utils.APIerSv2GetAccount, attrAcc, &acnt); err != nil {
			t.Error(err)
		} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != 2940000000000.0 {
			t.Errorf("Expecting : %+v, received: %+v", 2940000000000.0, acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
		}
	})

	t.Run("verify account after start", func(t *testing.T) {
		var acnt *engine.Account
		attrAcc := &utils.AttrGetAccount{
			Tenant:  "cgrates.org",
			Account: "1001",
		}
		// because US_SITE was down we should notice a difference between balance from accounts from US_SITE and AU_SITE
		if err := auClient.Call(context.Background(), utils.APIerSv2GetAccount, attrAcc, &acnt); err != nil {
			t.Error(err)
		} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != 2940000000000.0 {
			t.Errorf("Expecting : %+v, received: %+v", 2940000000000.0, acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
		}

		if err := usClient.Call(context.Background(), utils.APIerSv2GetAccount, attrAcc, &acnt); err != nil {
			t.Error(err)
		} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != 3240000000000.0 {
			t.Errorf("Expecting : %+v, received: %+v", 3240000000000.0, acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
		}
	})

	t.Run("update session 2", func(t *testing.T) {
		reqUsage := 5 * time.Minute
		args := &sessions.V1UpdateSessionArgs{
			UpdateSession: true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestSSv1ItUpdateSession2",
				Event: map[string]any{
					utils.Tenant:       "cgrates.org",
					utils.ToR:          utils.MetaVoice,
					utils.OriginID:     "testGOCS",
					utils.Category:     "call",
					utils.RequestType:  utils.MetaPrepaid,
					utils.AccountField: "1001",
					utils.Subject:      "1001",
					utils.Destination:  "1002",
					utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
					utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
					utils.Usage:        reqUsage,
				},
			},
		}
		var rply sessions.V1UpdateSessionReply
		// Update the session on both US_SITE and AU_SITE
		// With this update the account should be replicate from US_SITE to AU_SITE and forgot about the update than happens on AU_SITE
		if err := dspClient.Call(context.Background(), utils.SessionSv1UpdateSession, args, &rply); err != nil {
			t.Errorf("Expecting : %+v, received: %+v", nil, err)
		} else if rply.MaxUsage == nil || *rply.MaxUsage != reqUsage {
			t.Errorf("Unexpected MaxUsage: %v", rply.MaxUsage)
		}

		aSessions := make([]*sessions.ExternalSession, 0)
		if err := auClient.Call(context.Background(), utils.SessionSv1GetActiveSessions, new(utils.SessionFilter), &aSessions); err != nil {
			t.Error(err)
		} else if len(aSessions) != 1 {
			t.Errorf("wrong active sessions: %s", utils.ToJSON(aSessions))
		} else if aSessions[0].NodeID != "AU_SITE" {
			t.Errorf("Expecting : %+v, received: %+v", "AU_SITE", aSessions[0].NodeID)
		} else if aSessions[0].Usage != 5*time.Minute {
			t.Errorf("Expecting : %+v, received: %+v", 5*time.Minute, aSessions[0].Usage)
		}

		aSessions = make([]*sessions.ExternalSession, 0)
		if err := usClient.Call(context.Background(), utils.SessionSv1GetActiveSessions, new(utils.SessionFilter), &aSessions); err != nil {
			t.Error(err)
		} else if len(aSessions) != 1 {
			t.Errorf("wrong active sessions: %s \n , and len(aSessions) %+v", utils.ToJSON(aSessions), len(aSessions))
		} else if aSessions[0].NodeID != "US_SITE" {
			t.Errorf("Expecting : %+v, received: %+v", "US_SITE", aSessions[0].NodeID)
		} else if aSessions[0].Usage != 10*time.Minute {
			t.Errorf("Expecting : %+v, received: %+v", 5*time.Minute, aSessions[0].Usage)
		}

		var acnt *engine.Account
		attrAcc := &utils.AttrGetAccount{
			Tenant:  "cgrates.org",
			Account: "1001",
		}

		if err := auClient.Call(context.Background(), utils.APIerSv2GetAccount, attrAcc, &acnt); err != nil {
			t.Error(err)
		} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != 2940000000000.0 {
			t.Errorf("Expecting : %+v, received: %+v", 2940000000000.0, acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
		}

		if err := usClient.Call(context.Background(), utils.APIerSv2GetAccount, attrAcc, &acnt); err != nil {
			t.Error(err)
		} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != 2940000000000.0 {
			t.Errorf("Expecting : %+v, received: %+v", 2940000000000.0, acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
		}
	})

	t.Run("terminate session", func(t *testing.T) {
		args := &sessions.V1TerminateSessionArgs{
			TerminateSession: true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "testGOCSTerminateSession",
				Event: map[string]any{
					utils.Tenant:       "cgrates.org",
					utils.ToR:          utils.MetaVoice,
					utils.OriginID:     "testGOCS",
					utils.Category:     "call",
					utils.RequestType:  utils.MetaPrepaid,
					utils.AccountField: "1001",
					utils.Subject:      "1001",
					utils.Destination:  "1002",
					utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
					utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
					utils.Usage:        15 * time.Minute,
				},
			},
		}
		var rply string
		// we send terminate session with the correct usage, but because the US_SITE was down
		// this lost the previous session operations and will debit more
		if err := dspClient.Call(context.Background(), utils.SessionSv1TerminateSession,
			args, &rply); err != nil {
			t.Error(err)
		}
		if rply != utils.OK {
			t.Errorf("Unexpected reply: %s", rply)
		}
		aSessions := make([]*sessions.ExternalSession, 0)
		if err := usClient.Call(context.Background(), utils.SessionSv1GetActiveSessions, new(utils.SessionFilter), &aSessions); err == nil ||
			err.Error() != utils.ErrNotFound.Error() {
			t.Errorf("Expected error %s received error %v and reply %s", utils.ErrNotFound, err, utils.ToJSON(aSessions))
		}

		var acnt *engine.Account
		attrAcc := &utils.AttrGetAccount{
			Tenant:  "cgrates.org",
			Account: "1001",
		}

		if err := auClient.Call(context.Background(), utils.APIerSv2GetAccount, attrAcc, &acnt); err != nil {
			t.Error(err)
		} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != 2640000000000.0 {
			t.Errorf("Expecting : %+v, received: %+v", 2640000000000.0, acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
		}

		if err := usClient.Call(context.Background(), utils.APIerSv2GetAccount, attrAcc, &acnt); err != nil {
			t.Error(err)
		} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != 2640000000000.0 {
			t.Errorf("Expecting : %+v, received: %+v", 2640000000000.0, acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
		}
	})

	t.Run("process cdr", func(t *testing.T) {
		args := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSSv1ItProcessCDR",
			Event: map[string]any{
				utils.Tenant:       "cgrates.org",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "testGOCS",
				utils.Category:     "call",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:        15 * time.Minute,
			},
		}
		var rply string
		// process cdr should apply the correction because terminate was debited to much
		// 59 - 15 = 44 minutes
		if err := usClient.Call(context.Background(), utils.SessionSv1ProcessCDR,
			args, &rply); err != nil {
			t.Error(err)
		}
		if rply != utils.OK {
			t.Errorf("Unexpected reply: %s", rply)
		}
		time.Sleep(100 * time.Millisecond)
		var acnt *engine.Account
		attrAcc := &utils.AttrGetAccount{
			Tenant:  "cgrates.org",
			Account: "1001",
		}

		if err := auClient.Call(context.Background(), utils.APIerSv2GetAccount, attrAcc, &acnt); err != nil {
			t.Error(err)
		} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != 2640000000000.0 {
			t.Errorf("Expecting : %+v, received: %+v", 2640000000000.0, acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
		}

		if err := usClient.Call(context.Background(), utils.APIerSv2GetAccount, attrAcc, &acnt); err != nil {
			t.Error(err)
		} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != 2640000000000.0 {
			t.Errorf("Expecting : %+v, received: %+v", 2640000000000.0, acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
		}
	})
}
