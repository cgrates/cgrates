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

package analyzers

import (
	"encoding/json"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func TestNewAnalyzerService(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.AnalyzerSCfg().DBPath = "/tmp/analyzers"
	if err := os.RemoveAll(cfg.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cfg.AnalyzerSCfg().DBPath, 0700); err != nil {
		t.Fatal(err)
	}
	anz, err := NewAnalyzerService(cfg)
	if err != nil {
		t.Fatal(err)
	}
	// no need to DeepEqual
	if err = anz.Shutdown(); err != nil {
		t.Fatal(err)
	}
	if err = anz.initDB(); err != nil {
		t.Fatal(err)
	}
	shdChan := make(chan struct{}, 1)
	shdChan <- struct{}{}
	if err := anz.ListenAndServe(shdChan); err != nil {
		t.Fatal(err)
	}
	anz.db.Close()
	if err := os.RemoveAll(cfg.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
	}
}

func TestAnalyzerSLogTraffic(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()

	cfg.AnalyzerSCfg().DBPath = "/tmp/analyzers"
	cfg.AnalyzerSCfg().TTL = 30 * time.Minute
	if err := os.RemoveAll(cfg.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cfg.AnalyzerSCfg().DBPath, 0700); err != nil {
		t.Fatal(err)
	}
	anz, err := NewAnalyzerService(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t1 := time.Now().Add(-time.Hour)
	if err = anz.logTrafic(0, utils.AnalyzerSv1Ping, "status", "result", "error",
		utils.MetaJSON, "127.0.0.1:5565", "127.0.0.1:2012", t1, t1.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if err = anz.logTrafic(0, utils.CoreSv1Status, "status", "result", "error",
		utils.MetaJSON, "127.0.0.1:5565", "127.0.0.1:2012", t1, t1.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	t1 = time.Now().Add(-10 * time.Minute)
	if err = anz.logTrafic(0, utils.CoreSv1Status, "status", "result", "error",
		utils.MetaJSON, "127.0.0.1:5565", "127.0.0.1:2012", t1, t1.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if cnt, err := anz.db.DocCount(); err != nil {
		t.Fatal(err)
	} else if cnt != 2 {
		t.Errorf("Expected only 2 documents received:%v", cnt)
	}
	if err = anz.clenaUp(); err != nil {
		t.Fatal(err)
	}
	if cnt, err := anz.db.DocCount(); err != nil {
		t.Fatal(err)
	} else if cnt != 1 {
		t.Errorf("Expected only one document received:%v", cnt)
	}

	if err = anz.db.Close(); err != nil {
		t.Fatal(err)
	}
	if err = anz.clenaUp(); err != bleve.ErrorIndexClosed {
		t.Errorf("Expected error: %v,received: %+v", bleve.ErrorIndexClosed, err)
	}
	if err := os.RemoveAll(cfg.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
	}
}

func TestAnalyzersDeleteHits(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()

	cfg.AnalyzerSCfg().DBPath = "/tmp/analyzers"
	cfg.AnalyzerSCfg().TTL = 30 * time.Minute
	if err := os.RemoveAll(cfg.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cfg.AnalyzerSCfg().DBPath, 0700); err != nil {
		t.Fatal(err)
	}
	anz, err := NewAnalyzerService(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err = anz.deleteHits(search.DocumentMatchCollection{&search.DocumentMatch{}}); err != utils.ErrPartiallyExecuted {
		t.Errorf("Expected error: %v,received: %+v", utils.ErrPartiallyExecuted, err)
	}
	if err := os.RemoveAll(cfg.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
	}
}

func TestAnalyzersListenAndServe(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()

	cfg.AnalyzerSCfg().DBPath = "/tmp/analyzers"
	cfg.AnalyzerSCfg().TTL = 30 * time.Minute
	if err := os.RemoveAll(cfg.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cfg.AnalyzerSCfg().DBPath, 0700); err != nil {
		t.Fatal(err)
	}
	anz, err := NewAnalyzerService(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := anz.db.Close(); err != nil {
		t.Fatal(err)
	}
	anz.ListenAndServe(make(chan struct{}))

	cfg.AnalyzerSCfg().CleanupInterval = 1
	anz, err = NewAnalyzerService(cfg)
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		time.Sleep(1)
		runtime.Gosched()
		anz.db.Close()
	}()
	anz.ListenAndServe(make(chan struct{}))
	if err := os.RemoveAll(cfg.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
	}
}

func TestAnalyzersV1Search(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.AnalyzerSCfg().DBPath = "/tmp/analyzers"
	cfg.AnalyzerSCfg().TTL = 30 * time.Minute
	if err := os.RemoveAll(cfg.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cfg.AnalyzerSCfg().DBPath, 0700); err != nil {
		t.Fatal(err)
	}
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, true), cfg.CacheCfg(), nil)
	anz, err := NewAnalyzerService(cfg)

	if err != nil {
		t.Fatal(err)
	}
	anz.SetFilterS(engine.NewFilterS(cfg, nil, dm))
	// generate trafic
	t1 := time.Now()
	if err = anz.logTrafic(0, utils.CoreSv1Ping,
		&utils.CGREvent{
			APIOpts: map[string]interface{}{
				utils.EventSource: utils.MetaCDRs,
			},
		}, utils.Pong, nil, utils.MetaJSON, "127.0.0.1:5565",
		"127.0.0.1:2012", t1, t1.Add(time.Second)); err != nil {
		t.Fatal(err)
	}

	if err = anz.logTrafic(1, utils.CoreSv1Ping,
		&utils.CGREvent{
			APIOpts: map[string]interface{}{
				utils.EventSource: utils.MetaAttributes,
			},
		}, utils.Pong, nil,

		utils.MetaJSON, "127.0.0.1:5565", "127.0.0.1:2012",
		t1.Add(time.Second), t1.Add(20*time.Second)); err != nil {
		t.Fatal(err)
	}

	if err = anz.logTrafic(2, utils.CoreSv1Ping,
		&utils.CGREvent{
			APIOpts: map[string]interface{}{
				utils.EventSource: utils.MetaAttributes,
			},
		}, utils.Pong, nil,

		utils.MetaJSON, "127.0.0.1:5565", "127.0.0.1:2012",
		t1.Add(2*time.Second), t1.Add(10*time.Second)); err != nil {
		t.Fatal(err)
	}

	if err = anz.logTrafic(3, utils.CoreSv1Ping,
		&utils.CGREvent{
			APIOpts: map[string]interface{}{
				utils.EventSource: utils.MetaAttributes,
			},
		}, utils.Pong, nil,

		utils.MetaGOB, "127.0.0.1:5566", "127.0.0.1:2013",
		t1.Add(-24*time.Hour), t1.Add(-23*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err = anz.logTrafic(3, utils.CoreSv1Status,
		&utils.CGREvent{
			APIOpts: map[string]interface{}{
				utils.EventSource: utils.MetaEEs,
			},
		}, utils.Pong, nil,

		rpcclient.BiRPCJSON, "127.0.0.1:5566", "127.0.0.1:2013",
		t1.Add(-11*time.Hour), t1.Add(-10*time.Hour-30*time.Minute)); err != nil {
		t.Fatal(err)
	}
	reply := []map[string]interface{}{}
	if err = anz.V1StringQuery(&QueryArgs{HeaderFilters: `"` + utils.CoreSv1Ping + `"`}, &reply); err != nil {
		t.Fatal(err)
	} else if len(reply) != 4 {
		t.Errorf("Expected 4 hits received: %v", len(reply))
	}
	reply = []map[string]interface{}{}
	if err = anz.V1StringQuery(&QueryArgs{HeaderFilters: "RequestMethod:" + `"` + utils.CoreSv1Ping + `"`}, &reply); err != nil {
		t.Fatal(err)
	} else if len(reply) != 4 {
		t.Errorf("Expected 4 hits received: %v", len(reply))
	}

	expRply := []map[string]interface{}{{
		"RequestDestination": "127.0.0.1:2013",
		"RequestDuration":    "1h0m0s",
		"RequestEncoding":    "*gob",
		"RequestID":          3.,
		"RequestMethod":      "CoreSv1.Ping",
		"RequestParams":      json.RawMessage(`{"Tenant":"","ID":"","Time":null,"Event":null,"APIOpts":{"EventSource":"*attributes"}}`),
		"Reply":              json.RawMessage(`"Pong"`),
		"RequestSource":      "127.0.0.1:5566",
		"RequestStartTime":   t1.Add(-24 * time.Hour).UTC().Format(time.RFC3339),
		"ReplyError":         nil,
	}}
	reply = []map[string]interface{}{}
	if err = anz.V1StringQuery(&QueryArgs{HeaderFilters: utils.RequestDuration + ":>=" + strconv.FormatInt(int64(time.Hour), 10)}, &reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expRply, reply) {
		t.Errorf("Expected %s received: %s", utils.ToJSON(expRply), utils.ToJSON(reply))
	}

	reply = []map[string]interface{}{}
	if err = anz.V1StringQuery(&QueryArgs{HeaderFilters: utils.RequestStartTime + ":<=\"" + t1.Add(-23*time.Hour).UTC().Format(time.RFC3339) + "\""}, &reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expRply, reply) {
		t.Errorf("Expected %s received: %s", utils.ToJSON(expRply), utils.ToJSON(reply))
	}
	reply = []map[string]interface{}{}
	if err = anz.V1StringQuery(&QueryArgs{HeaderFilters: "RequestEncoding:*gob"}, &reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expRply, reply) {
		t.Errorf("Expected %s received: %s", utils.ToJSON(expRply), utils.ToJSON(reply))
	}

	reply = []map[string]interface{}{}
	if err = anz.V1StringQuery(&QueryArgs{
		HeaderFilters:  "RequestEncoding:*gob",
		ContentFilters: []string{"*string:~*rep:Pong"},
	}, &reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expRply, reply) {
		t.Errorf("Expected %s received: %s", utils.ToJSON(expRply), utils.ToJSON(reply))
	}
	reply = []map[string]interface{}{}
	if err = anz.V1StringQuery(&QueryArgs{
		HeaderFilters:  "RequestEncoding:*gob",
		ContentFilters: []string{"*string:~*req.APIOpts.EventSource:*attributes"},
	}, &reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expRply, reply) {
		t.Errorf("Expected %s received: %s", utils.ToJSON(expRply), utils.ToJSON(reply))
	}
	reply = []map[string]interface{}{}
	if err = anz.V1StringQuery(&QueryArgs{
		HeaderFilters:  "RequestEncoding:*gob",
		ContentFilters: []string{"*gt:~*hdr.RequestDuration:1m"},
	}, &reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expRply, reply) {
		t.Errorf("Expected %s received: %s", utils.ToJSON(expRply), utils.ToJSON(reply))
	}
	reply = []map[string]interface{}{}
	if err = anz.V1StringQuery(&QueryArgs{
		HeaderFilters:  "RequestEncoding:*gob",
		ContentFilters: []string{"*string:~*req.APIOpts.EventSource:*attributes"},
	}, &reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expRply, reply) {
		t.Errorf("Expected %s received: %s", utils.ToJSON(expRply), utils.ToJSON(reply))
	}
	expRply = []map[string]interface{}{{
		"RequestDestination": "127.0.0.1:2013",
		"RequestDuration":    "30m0s",
		"RequestEncoding":    "*birpc_json",
		"RequestID":          3.,
		"RequestMethod":      "CoreSv1.Status",
		"RequestParams":      json.RawMessage(`{"Tenant":"","ID":"","Time":null,"Event":null,"APIOpts":{"EventSource":"*ees"}}`),
		"Reply":              json.RawMessage(`"Pong"`),
		"RequestSource":      "127.0.0.1:5566",
		"RequestStartTime":   t1.Add(-11 * time.Hour).UTC().Format(time.RFC3339),
		"ReplyError":         nil,
	}}
	reply = []map[string]interface{}{}
	if err = anz.V1StringQuery(&QueryArgs{
		ContentFilters: []string{"*string:~*req.APIOpts.EventSource:*ees"},
	}, &reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expRply, reply) {
		t.Errorf("Expected %s received: %s", utils.ToJSON(expRply), utils.ToJSON(reply))
	}

	expRply = []map[string]interface{}{}
	reply = []map[string]interface{}{}
	if err = anz.V1StringQuery(&QueryArgs{
		HeaderFilters:  "RequestEncoding:*gob",
		ContentFilters: []string{"*string:~*req.APIOpts.EventSource:*cdrs"},
	}, &reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expRply, reply) {
		t.Errorf("Expected %s received: %s", utils.ToJSON(expRply), utils.ToJSON(reply))
	}
	if err = anz.V1StringQuery(&QueryArgs{
		HeaderFilters:  "RequestEncoding:*gob",
		ContentFilters: []string{"*notstring:~*req.APIOpts.EventSource:*attributes"},
	}, &reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expRply, reply) {
		t.Errorf("Expected %s received: %s", utils.ToJSON(expRply), utils.ToJSON(reply))
	}
	if err = anz.V1StringQuery(&QueryArgs{
		ContentFilters: []string{"*string:~*req.APIOpts.EventSource:*sessions"},
	}, &reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expRply, reply) {
		t.Errorf("Expected %s received: %s", utils.ToJSON(expRply), utils.ToJSON(reply))
	}

	expErr := utils.ErrPrefixNotErrNotImplemented("*type")
	if err = anz.V1StringQuery(&QueryArgs{
		HeaderFilters:  "RequestEncoding:*gob",
		ContentFilters: []string{"*type:~*opts.EventSource:*cdrs"},
	}, &reply); err == nil || err.Error() != expErr.Error() {
		t.Errorf("Expected error: %s,received:%v", expErr, err)
	}

	sTime := time.Now()
	if err = anz.db.Index(utils.ConcatenatedKey(utils.AttributeSv1Ping, strconv.FormatInt(sTime.Unix(), 10)),
		&InfoRPC{
			RequestDuration:  time.Second,
			RequestStartTime: sTime,
			RequestEncoding:  utils.MetaJSON,
			RequestID:        0,
			RequestMethod:    utils.AttributeSv1Ping,
			RequestParams:    `a`,
			Reply:            `{}`,
		}); err != nil {
		t.Fatal(err)
	}

	expErr = new(json.SyntaxError)
	if err = anz.V1StringQuery(&QueryArgs{
		HeaderFilters:  "RequestMethod:" + utils.AttributeSv1Ping,
		ContentFilters: []string{"*type:~*opts.EventSource:*cdrs"},
	}, &reply); err == nil || err.Error() != expErr.Error() {
		t.Errorf("Expected error: %s,received:%v", expErr, err)
	}
	if err = anz.db.Index(utils.ConcatenatedKey(utils.AttributeSv1Ping, strconv.FormatInt(sTime.Unix(), 10)),
		&InfoRPC{
			RequestDuration:  time.Second,
			RequestStartTime: sTime,
			RequestEncoding:  utils.MetaJSON,
			RequestID:        0,
			RequestMethod:    utils.AttributeSv1Ping,
			RequestParams:    `{}`,
			Reply:            `a`,
		}); err != nil {
		t.Fatal(err)
	}
	if err = anz.V1StringQuery(&QueryArgs{
		HeaderFilters:  "RequestMethod:" + utils.AttributeSv1Ping,
		ContentFilters: []string{"*type:~*opts.EventSource:*cdrs"},
	}, &reply); err == nil || err.Error() != expErr.Error() {
		t.Errorf("Expected error: %s,received:%v", expErr, err)
	}

	if err = anz.db.Close(); err != nil {
		t.Fatal(err)
	}
	if err = anz.V1StringQuery(&QueryArgs{HeaderFilters: "RequestEncoding:*gob"}, &reply); err != bleve.ErrorIndexClosed {
		t.Errorf("Expected error: %v,received: %+v", bleve.ErrorIndexClosed, err)
	}
	if err := os.RemoveAll(cfg.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
	}
}
