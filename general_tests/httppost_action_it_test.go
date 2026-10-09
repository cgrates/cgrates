//go:build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"fmt"
	"io"
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
				"Content-Length":  []string{"52399"},
				"Content-Type":    []string{"application/json"},
				"Hdr1":            []string{"val2", "val3"},
				"Hdr3":            []string{"4"},
				"Hdrdynamic":      []string{"Startheader1End", "Start2header2End2"},
				"User-Agent":      []string{"Go-http-client/1.1"},
			}
			if fmt.Sprintln(r.Header) != fmt.Sprintln(exp) {
				t.Errorf("expected <%+v>, received <%+v>", exp, r.Header)
			} // ignore the httptest related panic if this error fails, fixing the error will fix the panic
			sqBody, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
			}
			fmt.Println(string(sqBody))
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
									"*url": "<~*req.url>",
									// "*hdrs"
									utils.MetaHdrs: map[string]any{
										"hdr1":       []string{"val2", "val3"},
										"hdr3":       "4",
										"hdrDynamic": []string{"Start<~*req.InjectHeader1>End", "Start2<~*req.InjectHeader2>End2"},
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
					"InjectHeader1":    "header1",
					"InjectHeader2":    "header2",
					"url":              rsSrv.URL,
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

func TestHTTPActionFromConfigs(t *testing.T) {
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
		"enabled": true,
		"opts":{
			"*body": [				// static body to be inserted in *http request
				{
					"tenant": "*any",
					"filterIDs": [],
					"value": {
						"Tenant": "cgrates.org",
						"Account": "9999",
						"Units": 30,
					}
				}
			],
			"*hdrs": [				// static headers to be appended in *http request
				{
					"tenant": "*any",
					"filterIDs": [],
					"value": {
						"hdr1": ["val2", "val3"],
						"hdr3": "4",
						"hdrDynamic": ["Start<~*req.InjectHeader1>End", "Start2<~*req.InjectHeader2>End2"],
					}
				}
			],
			"*method": [				// method to be used in *http request
				{
					"tenant": "*any",
					"filterIDs": [],
					"value": "POST"
				}
			],
			"*url": [				// url where *http request will be sent
				{
					"tenant": "*any",
					"filterIDs": [],
					"value": "<~*req.url>"
				}
			]
		}
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
				"Content-Length":  []string{"52"},
				"Content-Type":    []string{"application/json"},
				"Hdr1":            []string{"val2", "val3"},
				"Hdr3":            []string{"4"},
				"Hdrdynamic":      []string{"Startheader1End", "Start2header2End2"},
				"User-Agent":      []string{"Go-http-client/1.1"},
			}
			if fmt.Sprintln(r.Header) != fmt.Sprintln(exp) {
				t.Errorf("expected <%+v>, received <%+v>", exp, r.Header)
			} // ignore the httptest related panic if this error fails, fixing the error will fix the panic
			sqBody, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
			}
			expBody := `{"Account":"9999","Tenant":"cgrates.org","Units":30}`
			if string(sqBody) != expBody {
				t.Errorf("expected <%v>, \nreceived <%v>", expBody, string(sqBody))
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
						Type: utils.MetaHTTP,
						Diktats: []*utils.APDiktat{
							{
								ID: "HttpsPost",
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
					"InjectHeader1":    "header1",
					"InjectHeader2":    "header2",
					"url":              rsSrv.URL,
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

func TestHTTPActionWithOptsFromAction(t *testing.T) {
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
		"enabled": true,
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
				"Content-Length":  []string{"52"},
				"Content-Type":    []string{"application/json"},
				"Hdr1":            []string{"val2", "val3"},
				"Hdr3":            []string{"4"},
				"Hdrdynamic":      []string{"Startheader1End", "Start2header2End2"},
				"User-Agent":      []string{"Go-http-client/1.1"},
			}
			if fmt.Sprintln(r.Header) != fmt.Sprintln(exp) {
				t.Errorf("expected <%+v>, received <%+v>", exp, r.Header)
			} // ignore the httptest related panic if this error fails, fixing the error will fix the panic
			sqBody, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
			}
			expBody := `{"Account":"9999","Tenant":"cgrates.org","Units":30}`
			if string(sqBody) != expBody {
				t.Errorf("expected <%v>, \nreceived <%v>", expBody, string(sqBody))
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
						Type: utils.MetaHTTP,
						Diktats: []*utils.APDiktat{
							{
								ID: "HttpsPost",
								Opts: map[string]any{
									"*body": map[string]any{
										utils.Tenant:       "cgrates.org",
										utils.AccountField: "9999",
										utils.Units:        30,
									},
									"*url": "<~*req.url>",
									// "*hdrs"
									utils.MetaHdrs: map[string]any{
										"hdr1":       []string{"val2", "val3"},
										"hdr3":       "4",
										"hdrDynamic": []string{"Start<~*req.InjectHeader1>End", "Start2<~*req.InjectHeader2>End2"},
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
					"InjectHeader1":    "header1",
					"InjectHeader2":    "header2",
					"url":              rsSrv.URL,
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
