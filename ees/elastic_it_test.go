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
	"context"
	"encoding/json"
	"flag"
	"net/rpc"
	"os/exec"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
	elasticsearch "github.com/elastic/go-elasticsearch"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
)

var (
	elasticConfigDir  string
	elasticCfgPath    string
	elasticCfg        *config.CGRConfig
	elasticRpc        *rpc.Client
	elasticServerPath = flag.Bool("elastic", false, "Run only if the user specify it")

	sTestsElastic = []func(t *testing.T){
		testCreateDirectory,
		testElasticLoadConfig,
		testElasticResetDataDB,
		testElasticResetStorDb,
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
	elasticCfgPath = path.Join(*dataDir, "conf", "samples", elasticConfigDir)
	if elasticCfg, err = config.NewCGRConfigFromPath(elasticCfgPath); err != nil {
		t.Error(err)
	}
}

func testElasticResetDataDB(t *testing.T) {
	if err := engine.InitDataDB(elasticCfg); err != nil {
		t.Fatal(err)
	}
}

func testElasticResetStorDb(t *testing.T) {
	if err := engine.InitStorDB(elasticCfg); err != nil {
		t.Fatal(err)
	}
}

func testElasticStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(elasticCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testElasticRPCConn(t *testing.T) {
	var err error
	elasticRpc, err = newRPCClient(elasticCfg.ListenCfg())
	if err != nil {
		t.Fatal(err)
	}
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
			Event: map[string]interface{}{
				utils.CGRID:        utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()),
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
				utils.RunID:        utils.MetaDefault,
				utils.Cost:         1.01,
				"ExtraFields": map[string]string{"extra1": "val_extra1",
					"extra2": "val_extra2", "extra3": "val_extra3"},
			},
		},
	}

	eventData := &utils.CGREventWithEeIDs{
		EeIDs: []string{"ElasticsearchExporter"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "dataEvent",
			Event: map[string]interface{}{
				utils.CGRID:        utils.Sha1("abcdef", time.Unix(1383813745, 0).UTC().String()),
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
				utils.RunID:        utils.MetaDefault,
				utils.Cost:         0.012,
				"ExtraFields": map[string]string{"extra1": "val_extra1",
					"extra2": "val_extra2", "extra3": "val_extra3"},
			},
		},
	}

	eventSMS := &utils.CGREventWithEeIDs{
		EeIDs: []string{"ElasticsearchExporter"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "SMSEvent",
			Event: map[string]interface{}{
				utils.CGRID:        utils.Sha1("sdfwer", time.Unix(1383813745, 0).UTC().String()),
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
				utils.RunID:        utils.MetaDefault,
				utils.Cost:         0.15,
				"ExtraFields": map[string]string{"extra1": "val_extra1",
					"extra2": "val_extra2", "extra3": "val_extra3"},
			},
		},
	}

	eventSMSNoFields := &utils.CGREventWithEeIDs{
		EeIDs: []string{"ElasticExporterWithNoFields"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "SMSEvent",
			Event: map[string]interface{}{
				utils.CGRID:        utils.Sha1("sms2", time.Unix(1383813745, 0).UTC().String()),
				utils.ToR:          utils.MetaSMS,
				utils.Tenant:       "cgrates.org",
				utils.Category:     "call",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.RunID:        utils.MetaDefault,
			},
			APIOpts: map[string]interface{}{
				"ExporterUsed": "ElasticExporterWithNoFields",
			},
		},
	}
	var reply map[string]utils.MapStorage
	if err := elasticRpc.Call(utils.EeSv1ProcessEvent, eventVoice, &reply); err != nil {
		t.Error(err)
	}
	if err := elasticRpc.Call(utils.EeSv1ProcessEvent, eventData, &reply); err != nil {
		t.Error(err)
	}
	if err := elasticRpc.Call(utils.EeSv1ProcessEvent, eventSMS, &reply); err != nil {
		t.Error(err)
	}
	if err := elasticRpc.Call(utils.EeSv1ProcessEvent, eventSMSNoFields, &reply); err != nil {
		t.Error(err)
	}
}

func testElasticVerifyExports(t *testing.T) {
	es, err := elasticsearch.NewDefaultClient()
	if err != nil {
		t.Error(err)
	}
	var r map[string]interface{}
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
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
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			t.Error(err)
		} else {
			t.Errorf("%+v", e)
		}
	}

	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		t.Error(err)
	}
	for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
		switch hit.(map[string]interface{})["_id"] {
		case "2478e9f18ebcd3c684f3c14596b8bfeab2b0d6d4:*default":
			eMp := map[string]interface{}{
				utils.AccountField: "1001",
				utils.AnswerTime:   "2013-11-07T08:42:26Z",
				utils.CGRID:        "2478e9f18ebcd3c684f3c14596b8bfeab2b0d6d4",
				utils.Category:     "call",
				utils.Cost:         "0.15",
				utils.Destination:  "1002",
				utils.OriginID:     "sdfwer",
				utils.RequestType:  "*rated",
				utils.RunID:        "*default",
				utils.SetupTime:    "2013-11-07T08:42:25Z",
				utils.Subject:      "1001",
				utils.Tenant:       "cgrates.org",
				utils.ToR:          "*sms",
				utils.Usage:        "1",
			}
			if !reflect.DeepEqual(eMp, hit.(map[string]interface{})["_source"]) {
				t.Errorf("Expected %+v, received: %+v", eMp, hit.(map[string]interface{})["_source"])
			}
		case "dbafe9c8614c785a65aabd116dd3959c3c56f7f6:*default":
			eMp := map[string]interface{}{
				utils.AccountField: "1001",
				utils.AnswerTime:   "2013-11-07T08:42:26Z",
				utils.CGRID:        "dbafe9c8614c785a65aabd116dd3959c3c56f7f6",
				utils.Category:     "call",
				utils.Cost:         "1.01",
				utils.Destination:  "1002",
				utils.OriginID:     "dsafdsaf",
				utils.RequestType:  "*rated",
				utils.RunID:        "*default",
				utils.SetupTime:    "2013-11-07T08:42:25Z",
				utils.Subject:      "1001",
				utils.Tenant:       "cgrates.org",
				utils.ToR:          "*voice",
				utils.Usage:        "10000000000",
			}
			if !reflect.DeepEqual(eMp, hit.(map[string]interface{})["_source"]) {
				t.Errorf("Expected %+v, received: %+v", eMp, hit.(map[string]interface{})["_source"])
			}
		case utils.Sha1("sms2", time.Unix(1383813745, 0).UTC().String()) + ":*default":
			eMp := map[string]interface{}{
				utils.CGRID:        utils.Sha1("sms2", time.Unix(1383813745, 0).UTC().String()),
				utils.ToR:          utils.MetaSMS,
				utils.Tenant:       "cgrates.org",
				utils.Category:     "call",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.RunID:        utils.MetaDefault,
			}
			if !reflect.DeepEqual(eMp, hit.(map[string]interface{})["_source"]) {
				t.Errorf("Expected %+v, received: %+v", eMp, hit.(map[string]interface{})["_source"])
			}
		}
	}
}

func testElasticCloseElasticsearch(t *testing.T) {
	if err := exec.Command("systemctl", "stop", "elasticsearch.service").Run(); err != nil {
		t.Error(err)
	}
}
