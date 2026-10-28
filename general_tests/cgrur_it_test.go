//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestCgrUrEventExporter(t *testing.T) {
	var dbCfg engine.DBCfg
	switch *utils.DBType {
	case utils.MetaMySQL:
		dbCfg = engine.MySQLDBCfg
	case utils.MetaInternal, utils.MetaRedis, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}
	conn := dbCfg.DB.DBConns[utils.MetaDefault]
	exportPath := fmt.Sprintf("%s://%s:%s@%s:%d",
		strings.TrimPrefix(*conn.Type, utils.Meta),
		*conn.User, *conn.Password, *conn.Host, *conn.Port)
	time.Sleep(100 * time.Millisecond)

	content := `{
"logger": {
	"level": 7
},
"sessions": {
	"enabled": true
},
"ers": {
	"enabled": true,
	"conns": {
		"*sessions": [{"connIDs": ["*localhost"]}],
	},
    "readers": [{
        "id": "cgrur",
		"runDelay": "0", // manual reading only
		"flags": ["*sessions"],
	   	"type": "*cgrur",
	   	"sourcePath": "` + exportPath + `",
	  	"opts": {
			"sqlDBName":"` + *conn.Name + `",
			"sqlTableName":"urs",
		},
    }
  ]
},
"ees": {
	"enabled": true,
	"exporters": [{
			"id": "ur_exporter",
			"type": "*cgrur",
			"exportPath": "` + exportPath + `",
			"opts": {
				"sqlDBName": "` + *conn.Name + `",
				"sqlTableName": "urs"
			},
			"synchronous": true,
			"blocker": false,
			"attempts": 1,
			"failedPostsDir": "*none"
		},
	]
},
"admins": {
	"enabled": true,
	"conns": {
		"*ers": [{"connIDs": ["*localhost"]}],
	}
}
}`

	ng := engine.TestEngine{
		ConfigJSON: content,
		DBCfg:      dbCfg,
		Encoding:   *utils.Encoding,
	}

	client, _ := ng.Run(t)
	time.Sleep(100 * time.Millisecond)

	t.Run("URExportEvent", func(t *testing.T) {
		cgrEvID := &utils.CGREventWithEeIDs{
			EeIDs: []string{"ur_exporter"},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "voiceEvent",
				Event: map[string]any{

					utils.ToR:          utils.MetaVoice,
					utils.OriginID:     "origin1",
					utils.OriginHost:   "192.168.1.1",
					utils.RequestType:  utils.MetaRated,
					utils.Tenant:       "cgrates.org",
					utils.Category:     "call",
					utils.AccountField: "1001",
					utils.Subject:      "1001",
					utils.Destination:  "1002",
					utils.SetupTime:    time.Unix(1383813745, 0).UTC(),
					utils.AnswerTime:   time.Unix(1383813746, 0).UTC(),
					utils.Usage:        10 * time.Second,
					utils.Cost:         1.01,
				},
				APIOpts: map[string]any{
					utils.MetaOriginID: utils.Sha1("origin1", time.Unix(1383813745, 0).UTC().String()),
					utils.RunID:        utils.MetaDefault,
				},
			},
		}
		var reply map[string]map[string]any
		if err := client.Call(context.Background(), utils.EeSv1ProcessEvent,
			cgrEvID, &reply); err != nil {
			t.Error(err)
		}
	})

	t.Run("GetUR", func(t *testing.T) {
		var urs []map[string]*utils.CGREvent
		if err := client.Call(context.Background(), utils.AdminSv1GetURs,
			&apis.URFiltersFromReader{Tenant: "cgrates.org", ReaderID: "cgrur"}, &urs); err != nil {
			t.Fatalf("retrieving URs failed: %v", err)
		}

		if len(urs) != 1 {
			t.Fatalf("expected 1 UR, got %d", len(urs))
		}
		expUrs := []map[string]*utils.CGREvent{
			{
				utils.MetaPrimary: &utils.CGREvent{
					Tenant: "cgrates.org",
					APIOpts: map[string]any{
						utils.MetaExporterID: "ur_exporter",
						utils.MetaOriginID:   utils.Sha1("origin1", time.Unix(1383813745, 0).UTC().String()),
						utils.MetaURID:       utils.Sha1("origin1", time.Unix(1383813745, 0).UTC().String()),
						utils.MetaRunID:      utils.MetaDefault,
					},
					Event: map[string]any{
						utils.ToR:          utils.MetaVoice,
						utils.OriginID:     "origin1",
						utils.OriginHost:   "192.168.1.1",
						utils.RequestType:  utils.MetaRated,
						utils.Tenant:       "cgrates.org",
						utils.Category:     "call",
						utils.AccountField: "1001",
						utils.Subject:      "1001",
						utils.Destination:  "1002",
						utils.SetupTime:    time.Unix(1383813745, 0).UTC(),
						utils.AnswerTime:   time.Unix(1383813746, 0).UTC(),
						utils.Usage:        10 * time.Second,
						utils.Cost:         1.01,
					},
				},
			},
		}
		if reflect.DeepEqual(expUrs, urs) {
			t.Errorf("expected <%v>,\n received <%v>", utils.ToJSON(expUrs), utils.ToJSON(urs))
		}
	})
}
