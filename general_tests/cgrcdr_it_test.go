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
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestCgrCdrEventExporter(t *testing.T) {
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
"ees": {
	"enabled": true,
	"exporters": [{
			"id": "cdr_exporter",
			"type": "*cgrcdr",
			"export_path": "` + exportPath + `",
			"opts": {
				"sqlDBName": "` + *conn.Name + `",
				"sqlTableName": "cdrs"
			},
			"synchronous": true,
			"blocker": false,
			"attempts": 1,
			"failedPostsDir": "*none"
		},
	]
},
"admins": {
	"enabled": true
}
}`

	ng := engine.TestEngine{
		ConfigJSON: content,
		DBCfg:      dbCfg,
		Encoding:   *utils.Encoding,
	}

	client, _ := ng.Run(t)
	time.Sleep(100 * time.Millisecond)

	t.Run("CDRExportEvent", func(t *testing.T) {
		cgrEvID := &utils.CGREventWithEeIDs{
			EeIDs: []string{"cdr_exporter"},
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
	t.Run("GetCDR", func(t *testing.T) {
		var cdrs []*utils.CDR
		if err := client.Call(context.Background(), utils.AdminSv1GetCDRs,
			&utils.CDRFilters{Tenant: "cgrates.org"}, &cdrs); err != nil {
			t.Fatalf("retrieving CDRs failed: %v", err)
		}
		if len(cdrs) != 1 {
			t.Fatalf("expected 1 CDR, got %d", len(cdrs))
		}
		if got := cdrs[0].Event[utils.OriginID]; got != "origin1" {
			t.Errorf("unexpected OriginID: got %v, want origin1", got)
		}
	})
}
