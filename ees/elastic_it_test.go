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

package ees

import (
	"encoding/json"
	"fmt"
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
	elasticsearch "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/google/go-cmp/cmp"
)

func TestElasticsearchIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
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

	// Initialize separate clients for each exporter.
	esClBasic := initElsClient(t, cfg, "basic")
	esClFields := initElsClient(t, cfg, "fields")

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
	verifyElsExports(t, esClBasic, "basic", n, map[string]any{
		utils.AccountField: "1001",
		utils.ToR:          utils.MetaData,
		utils.RequestType:  utils.MetaPostpaid,
	})
	verifyElsExports(t, esClFields, "fields", n, map[string]any{
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
				Event: map[string]interface{}{
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
func verifyElsExports(t *testing.T, client *elasticsearch.TypedClient, exporterType string, n int, expSource map[string]any) {
	t.Helper()
	req := search.Request{
		Query: &types.Query{MatchAll: &types.MatchAllQuery{}},
	}
	if n > 10 && n <= 10_000 {
		// Return more than the default 10 results limit if needed.
		// Max limit is 10_000.
		req.Size = &n
	}
	index := fmt.Sprintf("cdrs_%s", exporterType)
	resp, err := client.Search().
		Index(index).
		Request(&req).
		Do(context.TODO())
	if err != nil {
		t.Fatal(err)
	}
	if hc := len(resp.Hits.Hits); hc != n {
		t.Fatalf("len(resp.Hits.Hits)=%d, want %d", hc, n)
	}
	slices.SortFunc(resp.Hits.Hits, func(a, b types.Hit) int {
		switch {
		case *a.Id_ < *b.Id_:
			return -1
		case *a.Id_ > *b.Id_:
			return 1
		}
		return 0
	})
	for i, hit := range resp.Hits.Hits {
		wantUsage := i + 1
		wantOriginID := fmt.Sprintf("%s%03d", exporterType, wantUsage)

		if strings.HasPrefix(*hit.Id_, "basic") {
			expSource[utils.Usage] = float64(wantUsage)
		} else {
			expSource[utils.Usage] = strconv.Itoa(wantUsage)
		}
		expSource[utils.OriginID] = wantOriginID
		wantDocID := wantOriginID + ":*default"
		if *hit.Id_ != wantDocID {
			t.Errorf("hit.Id_ = %s, want %s", *hit.Id_, wantDocID)
		}
		var got map[string]any
		if err := json.Unmarshal(hit.Source_, &got); err != nil {
			t.Error(err)
		}

		if strings.HasPrefix(*hit.Id_, "fields") {
			// Check if @timestamp field exists and has the correct format.
			// No need to test the exact value.
			timestamp, has := got["@timestamp"]
			if !has {
				t.Fatalf("timestamp missing in document with ID %s", *hit.Id_)
			}
			if _, err := time.Parse(time.RFC3339, utils.IfaceAsString(timestamp)); err != nil {
				t.Fatalf("failed to parse @timestamp field in document with ID %s", *hit.Id_)
			}
			expSource["@timestamp"] = timestamp
		}

		if diff := cmp.Diff(expSource, got); diff != "" {
			t.Errorf("SearchAll(index=%q) returned unexpected result (-want +got): \n%s", index, diff)
		}
	}
}

func initElsClient(t *testing.T, cfg *config.CGRConfig, exporterType string) *elasticsearch.TypedClient {
	eeCfg := cfg.EEsCfg().ExporterCfg(fmt.Sprintf("els_%s", exporterType))
	tmp := &ElasticEE{
		cfg: eeCfg,
	}
	if err := tmp.parseClientOpts(); err != nil {
		t.Fatal(err)
	}
	client, err := elasticsearch.NewTypedClient(tmp.clientCfg)
	if err != nil {
		t.Fatal(err)
	}

	// info, err := client.Info().Do(context.TODO())
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// fmt.Println(utils.ToJSON(info))

	// Ensure index is removed at the end. No need to create beforehand, as
	// it gets created automatically.
	if eeCfg.Opts.ElsIndex == nil {
		t.Fatal("elsIndex opt cannot be nil")
	}
	index := *eeCfg.Opts.ElsIndex

	// resp, err := client.Indices.Create(index).Do(context.TODO())
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// fmt.Println(utils.ToJSON(resp))

	t.Cleanup(func() {
		resp, err := client.Indices.Delete(index).Do(context.TODO())
		if err != nil || !resp.Acknowledged {
			t.Errorf("failed to delete index %s: %v", index, err)
		}

	})
	return client
}
