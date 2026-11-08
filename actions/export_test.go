// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package actions

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestACHTTPPostExecute(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	cfg.EFsCfg().PosterAttempts = 1
	apAction := &utils.APAction{
		ID:   "TEST_ACTION_HTTPPOST",
		Type: utils.MetaHTTPPost,
		Diktats: []*utils.APDiktat{
			{
				ID:       "first",
				Weights:  utils.DynamicWeights{{Weight: 20}},
				Blockers: utils.DynamicBlockers{{Blocker: true}},
				Opts: map[string]any{
					utils.MetaHTTPurl: "~*balance.TestBalance.Value",
				},
			},
			{
				ID:      "second",
				Weights: utils.DynamicWeights{{Weight: 10}},
				Opts: map[string]any{
					utils.MetaHTTPurl: "~*balance.TestBalance.Value",
				},
			},
		},
	}
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	http, err := newActHTTPPost(context.Background(), "cgrates.org", new(utils.CGREvent),
		cacheS, new(engine.FilterS), cfg, apAction)
	if err != nil {
		t.Fatal(err)
	}
	if len(http.pstrs) != 1 || http.pstrs[0] == nil {
		t.Fatalf("expected one non-nil poster, received %v", http.pstrs)
	}

	dataStorage := utils.MapStorage{
		utils.MetaReq: map[string]any{
			utils.AccountField: "1001",
		},
		utils.MetaOpts: map[string]any{
			utils.Usage: 10 * time.Minute,
		},
	}

	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
	}()
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 7)

	expected := `<EEs> Exporter <TEST_ACTION_HTTPPOST> could not export because err: <Post "~*balance.TestBalance.Value": unsupported protocol scheme "">`
	if err := http.execute(context.Background(), dataStorage, utils.EmptyString); err != nil {
		t.Error(err)
	} else if rcv := buf.String(); !strings.Contains(rcv, expected) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}

	// channels cannot be marshaled
	dataStorage[utils.MetaReq] = make(chan struct{}, 1)
	expected = "json: unsupported type: chan struct {}"
	if err := http.execute(context.Background(), dataStorage, utils.EmptyString); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	dataStorage = utils.MapStorage{
		utils.MetaOpts: map[string]any{
			utils.MetaAsync: 10 * time.Minute,
		},
	}
	http.aCfg.Opts = make(map[string]any)
	http.aCfg.Opts[utils.MetaAsync] = true
	http.config.EFsCfg().FailedPostsDir = utils.MetaNone
	if err := http.execute(context.Background(), dataStorage, utils.EmptyString); err != nil {
		t.Error(err)
	}
}

func TestACHTTPExecuteDiktats(t *testing.T) {
	type request struct {
		path string
		body map[string]any
	}
	received := make(chan request, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		received <- request{path: r.URL.Path, body: body}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	cfg := config.NewDefaultCGRConfig()
	filters := engine.NewFilterS(cfg, nil, nil)
	event := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event:  map[string]any{utils.AccountField: "1001"},
	}
	action := &utils.APAction{
		ID:   "http",
		Type: utils.MetaHTTP,
		Diktats: []*utils.APDiktat{
			{
				ID:      "second",
				Weights: utils.DynamicWeights{{Weight: 10}},
				Opts: map[string]any{
					utils.MetaHTTPurl:    server.URL + "/second",
					utils.MetaHTTPMethod: http.MethodPost,
					utils.MetaHTTPBody:   map[string]any{"Second": "~*req.Account"},
				},
			},
			{
				ID:      "first",
				Weights: utils.DynamicWeights{{Weight: 20}},
				Opts: map[string]any{
					utils.MetaHTTPurl:    server.URL + "/first",
					utils.MetaHTTPMethod: http.MethodPost,
					utils.MetaHTTPBody:   map[string]any{"First": "~*req.Account"},
				},
			},
		},
	}
	ctx := context.Background()
	act, err := newActHTTP(ctx, event.Tenant, event, nil, filters, cfg, action)
	if err != nil {
		t.Fatal(err)
	}
	if len(act.reqDetails) != 2 {
		t.Fatalf("expected two requests, received %d", len(act.reqDetails))
	}
	if err := act.execute(ctx, event.AsDataProvider(), ""); err != nil {
		t.Fatal(err)
	}
	for _, want := range []request{
		{path: "/first", body: map[string]any{"First": "1001"}},
		{path: "/second", body: map[string]any{"Second": "1001"}},
	} {
		if got := <-received; !reflect.DeepEqual(got, want) {
			t.Errorf("expected %v, received %v", want, got)
		}
	}
}

func TestACHTTPPostValues(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	cfg.EEsCfg().ExporterCfg(utils.MetaDefault).Attempts = 1
	cfg.EEsCfg().ExporterCfg(utils.MetaDefault).FailedPostsDir = utils.MetaNone
	apAction := &utils.APAction{
		ID:   "TEST_ACTION_HTTPPostValues",
		Type: utils.MetaHTTPPost,
		Diktats: []*utils.APDiktat{
			{
				Opts: map[string]any{
					"*balancePath": "~*balance.TestBalance.Value",
				},
			},
		},
	}
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	http, err := newActHTTPPost(context.Background(), "cgrates.org", new(utils.CGREvent),
		cacheS, new(engine.FilterS), cfg, apAction)
	if err != nil {
		t.Error(err)
	}
	dataStorage := utils.MapStorage{
		utils.MetaReq: map[string]any{
			utils.AccountField: 1003,
		},
	}

	if err := http.execute(context.Background(), dataStorage,
		utils.EmptyString); err == nil || err != utils.ErrPartiallyExecuted {
		t.Errorf("Expected %+v, received %+v", utils.ErrPartiallyExecuted, err)
	}
}

func TestACHTTPPostID(t *testing.T) {
	apAction := &utils.APAction{
		ID:   "TestACHTTPPostID",
		Type: utils.MetaHTTPPost,
	}
	http := &actHTTPPost{
		aCfg: apAction,
	}
	if rcv := http.id(); rcv != apAction.ID {
		t.Errorf("Expected %+v, received %+v", apAction.ID, rcv)
	}

	actESP := &actExport{
		aCfg: apAction,
	}
	if rcv := actESP.id(); rcv != apAction.ID {
		t.Errorf("Expected %+v, received %+v", apAction.ID, rcv)
	}
}
