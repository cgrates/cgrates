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
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

func TestSessDynaprepaidInit(t *testing.T) {
	ng := engine.TestEngine{
		ConfigPath: path.Join(*utils.DataDir, "conf", "samples", "sess_dynaprepaid"),
		TpPath:     path.Join(*utils.DataDir, "tariffplans", "testit"),
	}
	client, _ := ng.Run(t)
	time.Sleep(50 * time.Millisecond)
	t.Run("GetAccount", func(t *testing.T) {
		var acnt engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "CreatedAccount"}, &acnt); err == nil ||
			err.Error() != utils.ErrNotFound.Error() {
			t.Error(err)
		}
	})

	t.Run("InitSession", func(t *testing.T) {
		args1 := &sessions.V1InitSessionArgs{
			InitSession: true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				Event: map[string]any{
					utils.OriginID:     "sessDynaprepaid",
					utils.OriginHost:   "192.168.1.1",
					utils.Source:       "sessDynaprepaid",
					utils.ToR:          utils.MetaData,
					utils.RequestType:  utils.MetaDynaprepaid,
					utils.AccountField: "CreatedAccount",
					utils.Subject:      "NoSubject",
					utils.Destination:  "+1234567",
					utils.AnswerTime:   time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
					utils.Usage:        1024,
				},
			},
		}
		var rply1 sessions.V1InitSessionReply
		if err := client.Call(context.Background(), utils.SessionSv1InitiateSession,
			args1, &rply1); err != nil {
			t.Error(err)
			return
		} else if *rply1.MaxUsage != 1024*time.Nanosecond {
			t.Errorf("Expected <%+v>, received <%+v>", 1024*time.Nanosecond, *rply1.MaxUsage)
		}
	})

	t.Run("GetAccount2", func(t *testing.T) {
		var acnt engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "CreatedAccount"}, &acnt); err != nil {
			t.Error(err)
		}
		expAcc := &engine.Account{
			ID: "cgrates.org:CreatedAccount",
			BalanceMap: map[string]engine.Balances{
				utils.MetaMonetary: {
					&engine.Balance{
						Uuid:           acnt.BalanceMap[utils.MetaMonetary][0].Uuid,
						ID:             "",
						Categories:     utils.StringMap{},
						SharedGroups:   utils.StringMap{},
						TimingIDs:      utils.StringMap{},
						Value:          9.99966,
						Weight:         10,
						DestinationIDs: utils.StringMap{},
					},
				},
				utils.MetaSMS: {
					&engine.Balance{
						Uuid:           acnt.BalanceMap[utils.MetaSMS][0].Uuid,
						Value:          500,
						Weight:         10,
						DestinationIDs: utils.StringMap{},
						Categories:     utils.StringMap{},
						SharedGroups:   utils.StringMap{},
						TimingIDs:      utils.StringMap{},
					},
				},
			},
			UpdateTime: acnt.UpdateTime,
		}
		if !reflect.DeepEqual(utils.ToJSON(expAcc), utils.ToJSON(expAcc)) {
			t.Errorf("Expected <%v>, \nreceived <%v>", utils.ToJSON(expAcc.BalanceMap), utils.ToJSON(expAcc.BalanceMap))
		}
	})
}

func TestSessDynaprepaidUpdate(t *testing.T) {
	ng := engine.TestEngine{
		ConfigPath: path.Join(*utils.DataDir, "conf", "samples", "sess_dynaprepaid"),
		TpPath:     path.Join(*utils.DataDir, "tariffplans", "testit"),
	}
	client, _ := ng.Run(t)
	time.Sleep(50 * time.Millisecond)
	t.Run("GetAccount", func(t *testing.T) {
		var acnt engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "CreatedAccount"}, &acnt); err == nil ||
			err.Error() != utils.ErrNotFound.Error() {
			t.Error(err)
		}
	})

	t.Run("UpdateSession", func(t *testing.T) {
		args1 := &sessions.V1UpdateSessionArgs{
			UpdateSession: true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				Event: map[string]any{
					utils.OriginID:     "sessDynaprepaid",
					utils.OriginHost:   "192.168.1.1",
					utils.Source:       "sessDynaprepaid",
					utils.ToR:          utils.MetaData,
					utils.RequestType:  utils.MetaDynaprepaid,
					utils.AccountField: "CreatedAccount",
					utils.Subject:      "NoSubject",
					utils.Destination:  "+1234567",
					utils.AnswerTime:   time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
					utils.Usage:        1024,
				},
			},
		}
		var rply1 sessions.V1UpdateSessionReply
		if err := client.Call(context.Background(), utils.SessionSv1UpdateSession,
			args1, &rply1); err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("GetAccount2", func(t *testing.T) {
		var acnt engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "CreatedAccount"}, &acnt); err != nil {
			t.Error(err)
		}
		expAcc := &engine.Account{
			ID: "cgrates.org:CreatedAccount",
			BalanceMap: map[string]engine.Balances{
				utils.MetaMonetary: {
					&engine.Balance{
						Uuid:           acnt.BalanceMap[utils.MetaMonetary][0].Uuid,
						ID:             "",
						Categories:     utils.StringMap{},
						SharedGroups:   utils.StringMap{},
						TimingIDs:      utils.StringMap{},
						Value:          9.99966,
						Weight:         10,
						DestinationIDs: utils.StringMap{},
					},
				},
				utils.MetaSMS: {
					&engine.Balance{
						Uuid:           acnt.BalanceMap[utils.MetaSMS][0].Uuid,
						Value:          500,
						Weight:         10,
						DestinationIDs: utils.StringMap{},
						Categories:     utils.StringMap{},
						SharedGroups:   utils.StringMap{},
						TimingIDs:      utils.StringMap{},
					},
				},
			},
			UpdateTime: acnt.UpdateTime,
		}
		if !reflect.DeepEqual(utils.ToJSON(expAcc), utils.ToJSON(expAcc)) {
			t.Errorf("Expected <%v>, \nreceived <%v>", utils.ToJSON(expAcc.BalanceMap), utils.ToJSON(expAcc.BalanceMap))
		}
	})
}

func TestSessDynaprepaidTerminate(t *testing.T) {
	ng := engine.TestEngine{
		ConfigPath: path.Join(*utils.DataDir, "conf", "samples", "sess_dynaprepaid"),
		TpPath:     path.Join(*utils.DataDir, "tariffplans", "testit"),
	}
	client, _ := ng.Run(t)
	time.Sleep(50 * time.Millisecond)
	t.Run("GetAccount", func(t *testing.T) {
		var acnt engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "CreatedAccount"}, &acnt); err == nil ||
			err.Error() != utils.ErrNotFound.Error() {
			t.Error(err)
		}
	})

	t.Run("TerminateSession", func(t *testing.T) {
		args1 := &sessions.V1TerminateSessionArgs{
			TerminateSession: true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				Event: map[string]any{
					utils.OriginID:     "sessDynaprepaid",
					utils.OriginHost:   "192.168.1.1",
					utils.Source:       "sessDynaprepaid",
					utils.ToR:          utils.MetaData,
					utils.RequestType:  utils.MetaDynaprepaid,
					utils.AccountField: "CreatedAccount",
					utils.Subject:      "NoSubject",
					utils.Destination:  "+1234567",
					utils.AnswerTime:   time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
					utils.Usage:        1024,
				},
			},
		}
		var rply1 string
		if err := client.Call(context.Background(), utils.SessionSv1TerminateSession,
			args1, &rply1); err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("GetAccount2", func(t *testing.T) {
		var acnt engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "CreatedAccount"}, &acnt); err != nil {
			t.Error(err)
		}
		expAcc := &engine.Account{
			ID: "cgrates.org:CreatedAccount",
			BalanceMap: map[string]engine.Balances{
				utils.MetaMonetary: {
					&engine.Balance{
						Uuid:           acnt.BalanceMap[utils.MetaMonetary][0].Uuid,
						ID:             "",
						Categories:     utils.StringMap{},
						SharedGroups:   utils.StringMap{},
						TimingIDs:      utils.StringMap{},
						Value:          9.99966,
						Weight:         10,
						DestinationIDs: utils.StringMap{},
					},
				},
				utils.MetaSMS: {
					&engine.Balance{
						Uuid:           acnt.BalanceMap[utils.MetaSMS][0].Uuid,
						Value:          500,
						Weight:         10,
						DestinationIDs: utils.StringMap{},
						Categories:     utils.StringMap{},
						SharedGroups:   utils.StringMap{},
						TimingIDs:      utils.StringMap{},
					},
				},
			},
			UpdateTime: acnt.UpdateTime,
		}
		if !reflect.DeepEqual(utils.ToJSON(expAcc), utils.ToJSON(expAcc)) {
			t.Errorf("Expected <%v>, \nreceived <%v>", utils.ToJSON(expAcc.BalanceMap), utils.ToJSON(expAcc.BalanceMap))
		}
	})
}
