//go:build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestHTTPPostAction(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaRedis, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	var rsSrv *httptest.Server

	cfgJSON := fmt.Sprintf(`{
	"logger": {
		"level": 7
	},
	"actions": {
		"enabled": true
	},
	"admins": {
		"enabled": true
	}
}`)

	ng := engine.TestEngine{
		ConfigJSON: cfgJSON,
		DBCfg:      engine.InternalDBCfg,
		// LogBuffer:  &bytes.Buffer{},
	}
	// t.Cleanup(func() { fmt.Println(ng.LogBuffer) })
	client, _ := ng.Run(t)

	time.Sleep(200 * time.Millisecond)

	t.Run("StartHttpServer", func(t *testing.T) {
		rsSrv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			exp := http.Header{
				"Accept-Encoding": []string{"gzip"},
				"Content-Length":  []string{"52053"},
				"Content-Type":    []string{"application/json"},
				"Hdr1":            []string{"val2", "val3"},
				"Hdr3":            []string{"4"},
				"User-Agent":      []string{"Go-http-client/1.1"},
			}
			if fmt.Sprintln(r.Header) != fmt.Sprintln(exp) {
				t.Errorf("expected <%+v>, received <%+v>", exp, r.Header)
			}
			r.Body.Close()
		}))
	})

	t.Run("SetHTTPPostAction", func(t *testing.T) {
		actPrf := &utils.ActionProfileWithAPIOpts{
			ActionProfile: &utils.ActionProfile{
				Tenant: "cgrates.org",
				ID:     "actPrfID",
				Actions: []*utils.APAction{
					{
						ID:   "actID",
						Type: utils.MetaHTTPPost,
						Diktats: []*utils.APDiktat{
							{
								ID: "HttpsPost",
								Opts: map[string]any{
									"*url": rsSrv.URL,
									// "*hdrs"
									utils.MetaHdrs: map[string]any{
										"hdr1": []string{"val2", "val3"},
										"hdr3": "4",
									},
								},
							},
						},
						TTL: time.Duration(time.Minute),
					},
				},
			},
		}

		var reply *string
		if err := client.Call(context.Background(), utils.AdminSv1SetActionProfile,
			actPrf, &reply); err != nil {
			t.Error(err)
		}
	})

	t.Run("ExecuteHTTPPostAction", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.ActionSv1ExecuteActions,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "plan_1001",
				Event: map[string]any{
					utils.Tenant:       "cgrates.org",
					utils.AccountField: "1001",
					utils.Cost:         float64(30),
				},
			}, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Error("Unexpected reply returned", reply)
		}
	})

	t.Run("StopHttpServer", func(t *testing.T) {
		rsSrv.Close()
	})

}
