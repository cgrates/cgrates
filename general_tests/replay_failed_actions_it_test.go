//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
