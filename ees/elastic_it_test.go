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
	"bytes"
	"encoding/json"
	"flag"
	"os/exec"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
	elasticsearch "github.com/elastic/go-elasticsearch/v8"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
)

var (
	elasticConfigDir  string
	elasticCfgPath    string
	elasticCfg        *config.CGRConfig
	elasticRpc        *birpc.Client
	elasticServerPath = flag.Bool("elastic", false, "Run only if the user specify it")

	sTestsElastic = []func(t *testing.T){
		testCreateDirectory,
		testElasticLoadConfig,
		testElasticResetDBs,

		testElasticStartEngine,
		testElasticRPCConn,
		testElasticStartElasticsearch,
		testElasticExportEvents,
		testElasticVerifyExports,
		testStopCgrEngine,
		testElasticCloseElasticsearch,
		testCleanDirectory,
	}
)

// To run these tests first you need to install elasticsearch server locally as a daemon
// https://www.elastic.co/guide/en/elasticsearch/reference/current/deb.html
// and pass the elastic flag
func TestElasticExport(t *testing.T) {
	if !*elasticServerPath {
		t.SkipNow()
	}
	elasticConfigDir = "ees"
	for _, stest := range sTestsElastic {
		t.Run(elasticConfigDir, stest)
	}
}

func testElasticLoadConfig(t *testing.T) {
	var err error
	elasticCfgPath = path.Join(*utils.DataDir, "conf", "samples", elasticConfigDir)
	if elasticCfg, err = config.NewCGRConfigFromPath(context.Background(), elasticCfgPath); err != nil {
		t.Error(err)
	}
}

func testElasticResetDBs(t *testing.T) {
	if err := engine.InitDataDB(elasticCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDB(elasticCfg); err != nil {
		t.Fatal(err)
	}
}

func testElasticStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(elasticCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testElasticRPCConn(t *testing.T) {
	elasticRpc = engine.NewRPCClient(t, elasticCfg.ListenCfg(), *utils.Encoding)
}

func testElasticStartElasticsearch(t *testing.T) {
	if err := exec.Command("systemctl", "start", "elasticsearch.service").Run(); err != nil {
		t.Error(err)
	}
	// give some time to elasticsearch server to become up
	time.Sleep(5 * time.Second)
}

func testElasticExportEvents(t *testing.T) {
	eventVoice := &utils.CGREventWithEeIDs{
		EeIDs: []string{"ElasticsearchExporter"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "voiceEvent",
			Event: map[string]any{

				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "dsafdsaf",
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

				utils.Cost: 1.01,
				"ExtraFields": map[string]string{"extra1": "val_extra1",
					"extra2": "val_extra2", "extra3": "val_extra3"},
			},
			APIOpts: map[string]any{
				utils.MetaOriginID: utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()),
				utils.MetaRunID:    utils.MetaDefault,
			},
		},
	}

	eventData := &utils.CGREventWithEeIDs{
		EeIDs: []string{"ElasticsearchExporter"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "dataEvent",
			Event: map[string]any{

				utils.ToR:          utils.MetaData,
				utils.OriginID:     "abcdef",
				utils.OriginHost:   "192.168.1.1",
				utils.RequestType:  utils.MetaRated,
				utils.Tenant:       "AnotherTenant",
				utils.Category:     "call", //for data CDR use different Tenant
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Unix(1383813745, 0).UTC(),
				utils.AnswerTime:   time.Unix(1383813746, 0).UTC(),
				utils.Usage:        10 * time.Nanosecond,

				utils.Cost: 0.012,
				"ExtraFields": map[string]string{"extra1": "val_extra1",
					"extra2": "val_extra2", "extra3": "val_extra3"},
			},
			APIOpts: map[string]any{
				utils.MetaOriginID: utils.Sha1("abcdef", time.Unix(1383813745, 0).UTC().String()),
				utils.MetaRunID:    utils.MetaDefault,
			},
		},
	}

	eventSMS := &utils.CGREventWithEeIDs{
		EeIDs: []string{"ElasticsearchExporter"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "SMSEvent",
			Event: map[string]any{

				utils.ToR:          utils.MetaSMS,
				utils.OriginID:     "sdfwer",
				utils.OriginHost:   "192.168.1.1",
				utils.RequestType:  utils.MetaRated,
				utils.Tenant:       "cgrates.org",
				utils.Category:     "call",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Unix(1383813745, 0).UTC(),
				utils.AnswerTime:   time.Unix(1383813746, 0).UTC(),
				utils.Usage:        time.Duration(1),

				utils.Cost: 0.15,
				"ExtraFields": map[string]string{"extra1": "val_extra1",
					"extra2": "val_extra2", "extra3": "val_extra3"},
			},
			APIOpts: map[string]any{
				utils.MetaOriginID: utils.Sha1("sdfwer", time.Unix(1383813745, 0).UTC().String()),
				utils.MetaRunID:    utils.MetaDefault,
			},
		},
	}

	eventSMSNoFields := &utils.CGREventWithEeIDs{
		EeIDs: []string{"ElasticExporterWithNoFields"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "SMSEvent",
			Event: map[string]any{

				utils.ToR:          utils.MetaSMS,
				utils.Tenant:       "cgrates.org",
				utils.Category:     "call",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
			},
			APIOpts: map[string]any{
				utils.MetaOriginID: utils.Sha1("sms2", time.Unix(1383813745, 0).UTC().String()),
				"ExporterUsed":     "ElasticExporterWithNoFields",
				utils.MetaRunID:    utils.MetaDefault,
			},
		},
	}
	var reply map[string]utils.MapStorage
	if err := elasticRpc.Call(context.Background(), utils.EeSv1ProcessEvent, eventVoice, &reply); err != nil {
		t.Error(err)
	}
	if err := elasticRpc.Call(context.Background(), utils.EeSv1ProcessEvent, eventData, &reply); err != nil {
		t.Error(err)
	}
	if err := elasticRpc.Call(context.Background(), utils.EeSv1ProcessEvent, eventSMS, &reply); err != nil {
		t.Error(err)
	}
	if err := elasticRpc.Call(context.Background(), utils.EeSv1ProcessEvent, eventSMSNoFields, &reply); err != nil {
		t.Error(err)
	}
}

func testElasticVerifyExports(t *testing.T) {
	es, err := elasticsearch.NewDefaultClient()
	if err != nil {
		t.Error(err)
	}
	var r map[string]any
	var buf bytes.Buffer
	query := map[string]any{
		"query": map[string]any{
			"match": map[string]any{
				utils.Tenant: "cgrates.org",
			},
		},
	}
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		t.Error(err)
	}
	// Perform the search request.
	res, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex("cdrs"),
		es.Search.WithBody(&buf),
		es.Search.WithTrackTotalHits(true),
		es.Search.WithPretty(),
	)
	if err != nil {
		t.Error(err)
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]any
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			t.Error(err)
		} else {
			t.Errorf("%+v", e)
		}
	}

	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		t.Error(err)
	}
	for _, hit := range r["hits"].(map[string]any)["hits"].([]any) {
		switch hit.(map[string]any)["_id"] {
		case "2478e9f18ebcd3c684f3c14596b8bfeab2b0d6d4:*default":
			eMp := map[string]any{
				utils.AccountField: "1001",
				utils.AnswerTime:   "2013-11-07T08:42:26Z",
				utils.MetaOriginID: "2478e9f18ebcd3c684f3c14596b8bfeab2b0d6d4",
				utils.Category:     "call",
				utils.Cost:         "0.15",
				utils.Destination:  "1002",
				utils.OriginID:     "sdfwer",
				utils.RequestType:  "*rated",
				utils.MetaRunID:    "*default",
				utils.SetupTime:    "2013-11-07T08:42:25Z",
				utils.Subject:      "1001",
				utils.Tenant:       "cgrates.org",
				utils.ToR:          "*sms",
				utils.Usage:        "1",
			}
			if !reflect.DeepEqual(eMp, hit.(map[string]any)["_source"]) {
				t.Errorf("Expected %+v, received: %+v", eMp, hit.(map[string]any)["_source"])
			}
		case "dbafe9c8614c785a65aabd116dd3959c3c56f7f6:*default":
			eMp := map[string]any{
				utils.AccountField: "1001",
				utils.AnswerTime:   "2013-11-07T08:42:26Z",
				utils.MetaOriginID: "dbafe9c8614c785a65aabd116dd3959c3c56f7f6",
				utils.Category:     "call",
				utils.Cost:         "1.01",
				utils.Destination:  "1002",
				utils.OriginID:     "dsafdsaf",
				utils.RequestType:  "*rated",
				utils.MetaRunID:    "*default",
				utils.SetupTime:    "2013-11-07T08:42:25Z",
				utils.Subject:      "1001",
				utils.Tenant:       "cgrates.org",
				utils.ToR:          "*voice",
				utils.Usage:        "10000000000",
			}
			if !reflect.DeepEqual(eMp, hit.(map[string]any)["_source"]) {
				t.Errorf("Expected %+v, received: %+v", eMp, hit.(map[string]any)["_source"])
			}
		case utils.Sha1("sms2", time.Unix(1383813745, 0).UTC().String()) + ":*default":
			eMp := map[string]any{
				utils.MetaOriginID: utils.Sha1("sms2", time.Unix(1383813745, 0).UTC().String()),
				utils.ToR:          utils.MetaSMS,
				utils.Tenant:       "cgrates.org",
				utils.Category:     "call",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.MetaRunID:    utils.MetaDefault,
			}

			if !reflect.DeepEqual(eMp, hit.(map[string]any)["_source"]) {
				t.Errorf("Expected %+v, received: %+v", eMp, hit.(map[string]any)["_source"])
			}
		}
	}
}

func testElasticCloseElasticsearch(t *testing.T) {
	if err := exec.Command("systemctl", "stop", "elasticsearch.service").Run(); err != nil {
		t.Error(err)
	}
}
