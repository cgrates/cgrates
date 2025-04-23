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

func TestSessDynaprepaidAuth(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}
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

	t.Run("AuthSession", func(t *testing.T) {
		args1 := &sessions.V1AuthorizeArgs{
			GetMaxUsage: true,
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
		var rply1 sessions.V1AuthorizeReply
		if err := client.Call(context.Background(), utils.SessionSv1AuthorizeEvent,
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
						Value:          10,
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
		if !reflect.DeepEqual(utils.ToJSON(expAcc), utils.ToJSON(acnt)) {
			t.Errorf("Expected <%v>, \nreceived <%v>", utils.ToJSON(expAcc.BalanceMap), utils.ToJSON(acnt.BalanceMap))
		}
	})
}

func TestSessDynaprepaidInit(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}
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
		if !reflect.DeepEqual(utils.ToJSON(expAcc), utils.ToJSON(acnt)) {
			t.Errorf("Expected <%v>, \nreceived <%v>", utils.ToJSON(expAcc.BalanceMap), utils.ToJSON(acnt.BalanceMap))
		}
	})
}

func TestSessDynaprepaidUpdate(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}
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
		if !reflect.DeepEqual(utils.ToJSON(expAcc), utils.ToJSON(acnt)) {
			t.Errorf("Expected <%v>, \nreceived <%v>", utils.ToJSON(expAcc.BalanceMap), utils.ToJSON(acnt.BalanceMap))
		}
	})
}

func TestSessDynaprepaidTerminate(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}
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
		} else if rply1 != utils.OK {
			t.Errorf("Expected OK, received <%s>", rply1)
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
		if !reflect.DeepEqual(utils.ToJSON(expAcc), utils.ToJSON(acnt)) {
			t.Errorf("Expected <%v>, \nreceived <%v>", utils.ToJSON(expAcc.BalanceMap), utils.ToJSON(acnt.BalanceMap))
		}
	})
}

func TestSessDynaprepaidProcessEventAuth(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}
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

	t.Run("ProcessEventAuthSession", func(t *testing.T) {
		args1 := &sessions.V1ProcessEventArgs{
			Flags: []string{utils.MetaRALs + ":" + utils.MetaAuthorize},
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
		var rply1 sessions.V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			args1, &rply1); err != nil {
			t.Error(err)
			return
		} else if !reflect.DeepEqual(rply1.MaxUsage, map[string]time.Duration{utils.MetaRaw: 1024 * time.Nanosecond}) {
			t.Errorf("Expected <%+v>, received <%+v>", map[string]time.Duration{utils.MetaRaw: 1024 * time.Nanosecond}, rply1.MaxUsage)
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
						Value:          10,
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
		if !reflect.DeepEqual(utils.ToJSON(expAcc), utils.ToJSON(acnt)) {
			t.Errorf("Expected <%v>, \nreceived <%v>", utils.ToJSON(expAcc.BalanceMap), utils.ToJSON(acnt.BalanceMap))
		}
	})
}

func TestSessDynaprepaidProcessEventInit(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}
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

	t.Run("ProcessEventInitSession", func(t *testing.T) {
		args1 := &sessions.V1ProcessEventArgs{
			Flags: []string{utils.MetaRALs + ":" + utils.MetaInitiate},
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
		var rply1 sessions.V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			args1, &rply1); err != nil {
			t.Error(err)
			return
		} else if !reflect.DeepEqual(rply1.MaxUsage, map[string]time.Duration{utils.MetaRaw: 1024 * time.Nanosecond}) {
			t.Errorf("Expected <%+v>, received <%+v>", map[string]time.Duration{utils.MetaRaw: 1024 * time.Nanosecond}, rply1.MaxUsage)
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
		if !reflect.DeepEqual(utils.ToJSON(expAcc), utils.ToJSON(acnt)) {
			t.Errorf("Expected <%v>, \nreceived <%v>", utils.ToJSON(expAcc.BalanceMap), utils.ToJSON(acnt.BalanceMap))
		}
	})
}

func TestSessDynaprepaidProcessEventUpdate(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}
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

	t.Run("ProcessEventUpdateSession", func(t *testing.T) {
		args1 := &sessions.V1ProcessEventArgs{
			Flags: []string{utils.MetaRALs + ":" + utils.MetaUpdate},
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
		var rply1 sessions.V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			args1, &rply1); err != nil {
			t.Error(err)
			return
		} else if !reflect.DeepEqual(rply1.MaxUsage, map[string]time.Duration{utils.MetaRaw: 1024 * time.Nanosecond}) {
			t.Errorf("Expected <%+v>, received <%+v>", map[string]time.Duration{utils.MetaRaw: 1024 * time.Nanosecond}, rply1.MaxUsage)
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
		if !reflect.DeepEqual(utils.ToJSON(expAcc), utils.ToJSON(acnt)) {
			t.Errorf("Expected <%v>, \nreceived <%v>", utils.ToJSON(expAcc.BalanceMap), utils.ToJSON(acnt.BalanceMap))
		}
	})
}

func TestSessDynaprepaidProcessEventTerminate(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}
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

	t.Run("ProcessEventTerminateSession", func(t *testing.T) {
		args1 := &sessions.V1ProcessEventArgs{
			Flags: []string{utils.MetaRALs + ":" + utils.MetaTerminate},
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
		var rply1 sessions.V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			args1, &rply1); err != nil {
			t.Error(err)
			return
		} else if rply1.MaxUsage != nil {
			t.Errorf("Expected <nil>, received <%+v>", rply1.MaxUsage)
		}
	})

	t.Run("GetAccount2", func(t *testing.T) {
		var acnt engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "CreatedAccount"}, &acnt); err != nil {
			t.Fatal(err)
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
		if !reflect.DeepEqual(utils.ToJSON(expAcc), utils.ToJSON(acnt)) {
			t.Errorf("Expected <%v>, \nreceived <%v>", utils.ToJSON(expAcc.BalanceMap), utils.ToJSON(acnt.BalanceMap))
		}
	})
}

func TestSessDynaprepaidProcessEventCDRs(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}
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

	t.Run("ProcessEventCDRsSession", func(t *testing.T) {
		args1 := &sessions.V1ProcessEventArgs{
			Flags: []string{utils.MetaCDRs},
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
		var rply1 sessions.V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			args1, &rply1); err != nil {
			t.Error(err)
			return
		} else if rply1.MaxUsage != nil {
			t.Errorf("Expected <nil>, received <%+v>", rply1.MaxUsage)
		}
	})

	t.Run("GetAccount2", func(t *testing.T) {
		var acnt engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "CreatedAccount"}, &acnt); err != nil {
			t.Fatal(err)
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
		if !reflect.DeepEqual(utils.ToJSON(expAcc), utils.ToJSON(acnt)) {
			t.Errorf("Expected <%v>, \nreceived <%v>", utils.ToJSON(expAcc.BalanceMap), utils.ToJSON(acnt.BalanceMap))
		}
	})
}

func TestSessDynaprepaidProcessMessage(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}
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

	t.Run("ProcessMessageSession", func(t *testing.T) {
		args1 := &sessions.V1ProcessMessageArgs{
			Debit: true,
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
		var rply1 sessions.V1ProcessMessageReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessMessage,
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
			t.Fatal(err)
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
		if !reflect.DeepEqual(utils.ToJSON(expAcc), utils.ToJSON(acnt)) {
			t.Errorf("Expected <%v>, \nreceived <%v>", utils.ToJSON(expAcc.BalanceMap), utils.ToJSON(acnt.BalanceMap))
		}
	})
}
