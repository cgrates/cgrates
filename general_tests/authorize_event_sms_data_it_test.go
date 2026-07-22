//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"path"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

func TestSSv1AuthorizeEventSMS(t *testing.T) {
	var cfgDir string
	switch *utils.DBType {
	case utils.MetaInternal:
		cfgDir = "sessions_internal"
	case utils.MetaMySQL:
		cfgDir = "sessions_mysql"
	case utils.MetaMongo:
		cfgDir = "sessions_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	ng := engine.TestEngine{
		ConfigPath: path.Join(*utils.DataDir, "conf", "samples", cfgDir),
		TpPath:     path.Join(*utils.DataDir, "tariffplans", "testit"),
	}
	client, _ := ng.Run(t)

	t.Run("AuthorizeEventSMS", func(t *testing.T) {
		args := &sessions.V1AuthorizeArgs{
			GetMaxUsage:   true,
			GetAttributes: true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestSSv1ItAuthSMS",
				Event: map[string]any{
					utils.Tenant:       "cgrates.org",
					utils.ToR:          utils.MetaSMS,
					utils.OriginID:     "TestSSv1It1SMS",
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.SetupTime:    time.Date(2023, time.November, 16, 16, 60, 0, 0, time.UTC),
					utils.Usage:        20,
				},
			},
		}
		var rply sessions.V1AuthorizeReply
		if err := client.Call(context.Background(), utils.SessionSv1AuthorizeEvent, args, &rply); err != nil {
			t.Fatal(err)
		}
		if rply.MaxUsage == nil || *rply.MaxUsage != 20 {
			t.Errorf("Unexpected MaxUsage: %v", rply.MaxUsage)
		}
	})

	t.Run("AuthorizeEventData", func(t *testing.T) {
		args := &sessions.V1AuthorizeArgs{
			GetMaxUsage: true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestSSv1ItAuthData",
				Event: map[string]any{
					utils.Tenant:       "cgrates.org",
					utils.ToR:          utils.MetaData,
					utils.OriginID:     "TestSSv1It1Data",
					utils.AccountField: "1002",
					utils.Destination:  "1001",
					utils.SetupTime:    time.Date(2023, time.November, 16, 16, 60, 0, 0, time.UTC),
					utils.Usage:        1024,
				},
			},
		}
		var rply sessions.V1AuthorizeReply
		if err := client.Call(context.Background(), utils.SessionSv1AuthorizeEvent, args, &rply); err != nil {
			t.Fatal(err)
		}
		if rply.MaxUsage == nil || *rply.MaxUsage != 1024 {
			t.Errorf("Unexpected MaxUsage: %v", rply.MaxUsage)
		}
	})

}
