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

package ees

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/google/go-cmp/cmp"
)

func TestElasticsearchIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaRedis, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	ng := engine.TestEngine{
		ConfigPath: filepath.Join(*utils.DataDir, "conf", "samples", "ees_elastic"),
		DBCfg:      engine.InternalDBCfg,
		Encoding:   *utils.Encoding,
		// LogBuffer:  &bytes.Buffer{},
	}
	// defer fmt.Println(ng.LogBuffer)
	client, cfg := ng.Run(t)

	addrBasic := initElsIndex(t, cfg, "basic")
	addrFields := initElsIndex(t, cfg, "fields")

	n := 2 // number of events to export
	var wg sync.WaitGroup
	for i := range n {
		wg.Add(2)
		go func() {
			defer wg.Done()
			exportElsEvent(t, client, "basic", i+1)
		}()
		go func() {
			defer wg.Done()
			exportElsEvent(t, client, "fields", i+1)
		}()
	}
	wg.Wait()
	verifyElsExports(t, addrBasic, "basic", n, map[string]any{
		utils.AccountField: "1001",
		utils.ToR:          utils.MetaData,
		utils.RequestType:  utils.MetaPostpaid,
	})
	verifyElsExports(t, addrFields, "fields", n, map[string]any{
		utils.AccountField: "1001",
		utils.Source:       "test",
	})
}

func exportElsEvent(t *testing.T, client *birpc.Client, exporterSuffix string, i int) {
	t.Helper()
	var reply map[string]map[string]any
	if err := client.Call(context.Background(), utils.EeSv1ProcessEvent,
		&utils.CGREventWithEeIDs{
			EeIDs: []string{fmt.Sprintf("els_%s", exporterSuffix)},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.ToR:          utils.MetaData,
					utils.RequestType:  utils.MetaPostpaid,
					utils.Usage:        i,
				},
				APIOpts: map[string]any{
					utils.MetaOriginID: fmt.Sprintf("%s%03d", exporterSuffix, i),
				},
			},
		}, &reply); err != nil {
		t.Fatal(err)
	}
}

// To check via CLI:
//
// Get document count
// curl localhost:9200/cdrs_basic/_count
//
// Read all documents (default limit is 10)
// curl localhost:9200/cdrs_basic/_search
func verifyElsExports(t *testing.T, addr, exporterType string, n int, expSource map[string]any) {
	t.Helper()
	index := fmt.Sprintf("cdrs_%s", exporterType)
	resp, err := http.Get(fmt.Sprintf("%s/%s/_search?size=%d", addr, index, n))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var result struct {
		Hits struct {
			Hits []elsHit `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	hits := result.Hits.Hits
	if hc := len(hits); hc != n {
		t.Fatalf("len(hits)=%d, want %d", hc, n)
	}
	slices.SortFunc(hits, func(a, b elsHit) int {
		return strings.Compare(a.ID, b.ID)
	})
	for i, hit := range hits {
		wantUsage := i + 1
		wantOriginID := fmt.Sprintf("%s%03d", exporterType, wantUsage)

		if strings.HasPrefix(hit.ID, "basic") {
			expSource[utils.Usage] = float64(wantUsage)
		} else {
			expSource[utils.Usage] = strconv.Itoa(wantUsage)

			// OriginID can only be passed via templates, as it's part of
			// APIOpts. If none are configured, only the Event would be
			// exported.
			expSource[utils.OriginID] = wantOriginID
		}
		wantDocID := wantOriginID + ":*default"
		if hit.ID != wantDocID {
			t.Errorf("hit.ID = %s, want %s", hit.ID, wantDocID)
		}
		var got map[string]any
		if err := json.Unmarshal(hit.Source, &got); err != nil {
			t.Error(err)
		}

		if strings.HasPrefix(hit.ID, "fields") {
			// Check if @timestamp field exists and has the correct format.
			// No need to test the exact value.
			timestamp, has := got["@timestamp"]
			if !has {
				t.Fatalf("timestamp missing in document with ID %s", hit.ID)
			}
			if _, err := time.Parse(time.RFC3339, utils.IfaceAsString(timestamp)); err != nil {
				t.Fatalf("failed to parse @timestamp field in document with ID %s", hit.ID)
			}
			expSource["@timestamp"] = timestamp
		}

		if diff := cmp.Diff(expSource, got); diff != "" {
			t.Errorf("search(index=%q) returned unexpected result (-want +got): \n%s", index, diff)
		}
	}
}

type elsHit struct {
	ID     string          `json:"_id"`
	Source json.RawMessage `json:"_source"`
}

func initElsIndex(t *testing.T, cfg *config.CGRConfig, exporterType string) string {
	eeCfg := cfg.EEsCfg().ExporterCfg(fmt.Sprintf("els_%s", exporterType))
	if eeCfg.Opts.ElsIndex == nil {
		t.Fatal("elsIndex opt cannot be nil")
	}
	addr := strings.Split(eeCfg.ExportPath, utils.InfieldSep)[0]
	index := *eeCfg.Opts.ElsIndex

	// The index is created automatically on first export. Remove it at the end.
	t.Cleanup(func() {
		req, err := http.NewRequest(http.MethodDelete, addr+"/"+index, nil)
		if err != nil {
			t.Errorf("failed to delete index %s: %v", index, err)
			return
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Errorf("failed to delete index %s: %v", index, err)
			return
		}
		resp.Body.Close()
		if resp.StatusCode >= http.StatusMultipleChoices {
			t.Errorf("failed to delete index %s: %s", index, resp.Status)
		}
	})
	return addr
}
