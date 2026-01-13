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
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestReplayFailedActions(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := fmt.Sprintf(`
{
 "general": {
	"log_level": 7,
	"poster_attempts": 1
 },
 "listen": {
	"rpc_json": ":2012",
	"rpc_gob": ":2013",
	"http": ":2080",
 },
 "data_db": {
	"db_type": "redis",
	"db_port": 6379,
	"db_name": "10",
 },
 "stor_db": {
	"db_password": "CGRateS.org",
 },
"schedulers": {
	"enabled": true,
},
 "attributes": {
	"enabled": true,
 },
 "ees": {
	"enabled": true,
	"attributes_conns":["*internal"],
	"failed_posts": {
	    "dir": "%s",
		"ttl": "50ms",
	},
 },
 "apiers": {
	"enabled": true,
	"scheduler_conns": ["*internal"],
	"ees_conns": ["*localhost"],	
 },
}
`, tmpDir)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal("Could not open a listener:", err)
	}
	unreachableAddr := listener.Addr().String()
	listener.Close()
	unreachableURL := fmt.Sprintf("http://%s", unreachableAddr)
	buf := bytes.NewBuffer(nil)
	ng := engine.TestEngine{
		ConfigJSON: cfg,
		LogBuffer:  buf,
		TpFiles: map[string]string{
			utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,PKG_1,,,`,
			utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
PKG_1,Act_Top,*asap,10`,
			utils.ActionsCsv: fmt.Sprintf(`#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
Act_Top,*topup_reset,,,main_balance,*sms,,,,,*unlimited,,10,,,,
Act_Top,*http_post,%s,,,,,,,,,,,,,,`, unreachableURL),
		},
	}
	client, _ := ng.Run(t)
	time.Sleep(500 * time.Millisecond)
	var files []os.DirEntry
	t.Run("CheckIfGobFileExists", func(t *testing.T) {
		files, err = os.ReadDir(tmpDir)
		if err != nil {
			t.Fatalf("Could not read failed posts directory: %v", err)
		}
		if len(files) == 0 || !strings.HasSuffix(files[0].Name(), ".gob") {
			t.Error("expected a .gob file in failed_post directory")
		}
	})

	t.Run("CallReplayFailedPosts", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.APIerSv1ReplayFailedPosts, v1.ReplayFailedPostsParams{SourcePath: tmpDir}, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Error("expected to replay failed post")
		}
	})
	t.Run("CheckIfGobFileExistsAfter", func(t *testing.T) {
		files, err = os.ReadDir(tmpDir)
		if err != nil {
			t.Fatalf("Could not read failed posts directory: %v", err)
		}
		if len(files) == 0 || !strings.HasSuffix(files[0].Name(), ".gob") {
			t.Error("expected a .gob file in failed_post directory")
		}
	})
}
