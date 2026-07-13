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

package analyzers

import (
	stdctx "context"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

type baseIndex = bleve.Index

type indexWithInsert struct {
	baseIndex
	id  string
	doc any
}

func (idx *indexWithInsert) SearchInContext(ctx stdctx.Context, req *bleve.SearchRequest) (*bleve.SearchResult, error) {
	result, err := idx.baseIndex.SearchInContext(ctx, req)
	if err == nil && idx.doc != nil {
		doc := idx.doc
		idx.doc = nil
		if err = idx.baseIndex.Index(idx.id, doc); err != nil {
			return nil, err
		}
	}
	return result, err
}

func TestNewAnalyzerService(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.AnalyzerSCfg().DBPath = "/tmp/analyzers"
	if err := os.RemoveAll(cfg.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cfg.AnalyzerSCfg().DBPath, 0700); err != nil {
		t.Fatal(err)
	}
	anz, err := NewAnalyzerS(cfg)
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
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := anz.ListenAndServe(ctx); err != nil {
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
	anz, err := NewAnalyzerS(cfg)
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
	anz, err := NewAnalyzerS(cfg)
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
	anz, err := NewAnalyzerS(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := anz.db.Close(); err != nil {
		t.Fatal(err)
	}
	anz.ListenAndServe(context.Background())

	cfg.AnalyzerSCfg().CleanupInterval = 1
	anz, err = NewAnalyzerS(cfg)
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		time.Sleep(time.Nanosecond)
		anz.db.Close()
	}()
	anz.ListenAndServe(context.Background())
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
	idb, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: idb}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	dm.SetCache(engine.NewCacheS(cfg, dm, nil, nil))
	methodRule, err := engine.NewFilterRule(utils.MetaString, "~*hdr.RequestMethod", []string{utils.CoreSv1Ping})
	if err != nil {
		t.Fatal(err)
	}
	if err := dm.SetFilter(context.Background(), &engine.Filter{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "ANALYZER_METHOD",
		Rules:  []*engine.FilterRule{methodRule},
	}, false); err != nil {
		t.Fatal(err)
	}
	anz, err := NewAnalyzerS(cfg)
	if err != nil {
		t.Fatal(err)
	}
	anz.SetFilterS(engine.NewFilterS(cfg, nil, dm))
	// generate trafic
	t1 := time.Now()
	if err = anz.logTrafic(0, utils.CoreSv1Ping,
		&utils.CGREvent{
			APIOpts: map[string]any{
				utils.EventSource: utils.MetaCDRs,
			},
		}, utils.Pong, nil, utils.MetaJSON, "127.0.0.1:5565",
		"127.0.0.1:2012", t1, t1.Add(time.Second)); err != nil {
		t.Fatal(err)
	}

	if err = anz.logTrafic(1, utils.CoreSv1Ping,
		&utils.CGREvent{
			APIOpts: map[string]any{
				utils.EventSource: utils.MetaAttributes,
			},
		}, utils.Pong, nil,

		utils.MetaJSON, "127.0.0.1:5565", "127.0.0.1:2012",
		t1.Add(time.Second), t1.Add(20*time.Second)); err != nil {
		t.Fatal(err)
	}

	if err = anz.logTrafic(2, utils.CoreSv1Ping,
		&utils.CGREvent{
			APIOpts: map[string]any{
				utils.EventSource: utils.MetaAttributes,
			},
		}, utils.Pong, nil,

		utils.MetaJSON, "127.0.0.1:5565", "127.0.0.1:2012",
		t1.Add(2*time.Second), t1.Add(10*time.Second)); err != nil {
		t.Fatal(err)
	}

	if err = anz.logTrafic(3, utils.CoreSv1Ping,
		&utils.CGREvent{
			APIOpts: map[string]any{
				utils.EventSource: utils.MetaAttributes,
			},
		}, utils.Pong, nil,

		utils.MetaGOB, "127.0.0.1:5566", "127.0.0.1:2013",
		t1.Add(-24*time.Hour), t1.Add(-23*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err = anz.logTrafic(3, utils.CoreSv1Status,
		&utils.CGREvent{
			APIOpts: map[string]any{
				utils.EventSource: utils.MetaEEs,
			},
		}, utils.Pong, nil,

		rpcclient.BiRPCJSON, "127.0.0.1:5566", "127.0.0.1:2013",
		t1.Add(-11*time.Hour), t1.Add(-10*time.Hour-30*time.Minute)); err != nil {
		t.Fatal(err)
	}
	reply := []map[string]any{}
	if err = anz.V1StringQuery(context.Background(), &QueryArgs{Filters: []string{"ANALYZER_METHOD"}}, &reply); err != nil {
		t.Fatal(err)
	} else if len(reply) != 4 {
		t.Errorf("Expected 4 hits received: %v", len(reply))
	}

	expRply := []map[string]any{{
		"RequestDestination": "127.0.0.1:2013",
		"RequestDuration":    "1h0m0s",
		"RequestEncoding":    "*gob",
		"RequestID":          3.,
		"RequestMethod":      "CoreSv1.Ping",
		"RequestParams":      json.RawMessage(`{"Tenant":"","ID":"","Event":null,"APIOpts":{"EventSource":"*attributes"}}`),
		"Reply":              json.RawMessage(`"Pong"`),
		"RequestSource":      "127.0.0.1:5566",
		"RequestStartTime":   t1.Add(-24 * time.Hour).UTC().Format(time.RFC3339Nano),
		"ReplyError":         nil,
	}}
	reply = []map[string]any{}
	if err = anz.V1StringQuery(context.Background(), &QueryArgs{Filters: []string{"*gte:~*hdr.RequestDuration:" + strconv.FormatInt(int64(time.Hour), 10)}}, &reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expRply, reply) {
		t.Errorf("Expected %s received: %s", utils.ToJSON(expRply), utils.ToJSON(reply))
	}

	reply = []map[string]any{}
	if err = anz.V1StringQuery(context.Background(), &QueryArgs{Filters: []string{"*lte:~*hdr.RequestStartTime:" + t1.Add(-23*time.Hour).UTC().Format(time.RFC3339)}}, &reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expRply, reply) {
		t.Errorf("Expected %s received: %s", utils.ToJSON(expRply), utils.ToJSON(reply))
	}
	reply = []map[string]any{}
	if err = anz.V1StringQuery(context.Background(), &QueryArgs{Filters: []string{"*string:~*hdr.RequestEncoding:*gob"}}, &reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expRply, reply) {
		t.Errorf("Expected %s received: %s", utils.ToJSON(expRply), utils.ToJSON(reply))
	}
	reply = []map[string]any{}
	if err = anz.V1StringQuery(context.Background(), &QueryArgs{Filters: []string{
		"*string:~*hdr.RequestEncoding:*gob",
		"*string:~*rep:Pong",
	}}, &reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expRply, reply) {
		t.Errorf("Expected %s received: %s", utils.ToJSON(expRply), utils.ToJSON(reply))
	}
	reply = []map[string]any{}
	if err = anz.V1StringQuery(context.Background(), &QueryArgs{Filters: []string{
		"*string:~*hdr.RequestEncoding:*gob",
		"*string:~*req.APIOpts.EventSource:*attributes",
	}}, &reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expRply, reply) {
		t.Errorf("Expected %s received: %s", utils.ToJSON(expRply), utils.ToJSON(reply))
	}
	reply = []map[string]any{}
	if err = anz.V1StringQuery(context.Background(), &QueryArgs{Filters: []string{
		"*string:~*hdr.RequestEncoding:*gob",
		"*gt:~*hdr.RequestDuration:1m",
	}}, &reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expRply, reply) {
		t.Errorf("Expected %s received: %s", utils.ToJSON(expRply), utils.ToJSON(reply))
	}
	expRply = []map[string]any{{
		"RequestDestination": "127.0.0.1:2013",
		"RequestDuration":    "30m0s",
		"RequestEncoding":    "*birpc_json",
		"RequestID":          3.,
		"RequestMethod":      "CoreSv1.Status",
		"RequestParams":      json.RawMessage(`{"Tenant":"","ID":"","Event":null,"APIOpts":{"EventSource":"*ees"}}`),
		"Reply":              json.RawMessage(`"Pong"`),
		"RequestSource":      "127.0.0.1:5566",
		"RequestStartTime":   t1.Add(-11 * time.Hour).UTC().Format(time.RFC3339Nano),
		"ReplyError":         nil,
	}}
	reply = []map[string]any{}
	if err = anz.V1StringQuery(context.Background(), &QueryArgs{Filters: []string{"*string:~*req.APIOpts.EventSource:*ees"}}, &reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expRply, reply) {
		t.Errorf("Expected %s received: %s", utils.ToJSON(expRply), utils.ToJSON(reply))
	}

	expRply = []map[string]any{}
	reply = []map[string]any{}
	if err = anz.V1StringQuery(context.Background(), &QueryArgs{Filters: []string{
		"*string:~*hdr.RequestEncoding:*gob",
		"*string:~*req.APIOpts.EventSource:*cdrs",
	}}, &reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expRply, reply) {
		t.Errorf("Expected %s received: %s", utils.ToJSON(expRply), utils.ToJSON(reply))
	}
	if err = anz.V1StringQuery(context.Background(), &QueryArgs{Filters: []string{
		"*string:~*hdr.RequestEncoding:*gob",
		"*notstring:~*req.APIOpts.EventSource:*attributes",
	}}, &reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expRply, reply) {
		t.Errorf("Expected %s received: %s", utils.ToJSON(expRply), utils.ToJSON(reply))
	}
	if err = anz.V1StringQuery(context.Background(), &QueryArgs{Filters: []string{"*string:~*req.APIOpts.EventSource:*sessions"}}, &reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expRply, reply) {
		t.Errorf("Expected %s received: %s", utils.ToJSON(expRply), utils.ToJSON(reply))
	}

	if err = anz.V1StringQuery(context.Background(), &QueryArgs{Filters: []string{
		"*string:~*hdr.RequestEncoding:*gob",
		"*type:~*opts.EventSource:*cdrs",
	}}, &reply); err == nil || !strings.Contains(err.Error(), "Unsupported filter Type") {
		t.Errorf("expected unsupported filter error, received: %v", err)
	}

	payloadNumber := uint64(9007199254740995)
	payloadMethod := "Payload.Test"
	if err = anz.db.Index("large-payload", NewInfoRPC(0, payloadMethod,
		map[string]any{"Number": payloadNumber}, nil, nil, utils.MetaJSON, "", "", t1, t1)); err != nil {
		t.Fatal(err)
	}
	reply = nil
	if err = anz.V1StringQuery(context.Background(), &QueryArgs{Filters: []string{
		"*string:~*hdr.RequestMethod:" + payloadMethod,
		"*string:~*req.Number:" + strconv.FormatUint(payloadNumber, 10),
	}}, &reply); err != nil {
		t.Fatal(err)
	} else if len(reply) != 1 {
		t.Fatalf("expected one exact numeric payload match, got %d", len(reply))
	}

	sTime := time.Now()
	if err = anz.db.Index(utils.ConcatenatedKey(utils.AttributeSv1Ping, strconv.FormatInt(sTime.Unix(), 10)),
		&InfoRPC{
			RequestDuration:  time.Second,
			RequestStartTime: sTime.UTC().Format(time.RFC3339Nano),
			RequestEncoding:  utils.MetaJSON,
			RequestID:        0,
			RequestMethod:    utils.AttributeSv1Ping,
			RequestParams:    `a`,
			Reply:            `{}`,
		}); err != nil {
		t.Fatal(err)
	}

	var syntaxErr *json.SyntaxError
	if err = anz.V1StringQuery(context.Background(), &QueryArgs{Filters: []string{
		"*string:~*hdr.RequestMethod:" + utils.AttributeSv1Ping,
		"*string:~*opts.EventSource:*cdrs",
	}}, &reply); !errors.As(err, &syntaxErr) {
		t.Errorf("Expected JSON syntax error,received:%v", err)
	}
	if err = anz.db.Index(utils.ConcatenatedKey(utils.AttributeSv1Ping, strconv.FormatInt(sTime.Unix(), 10)),
		&InfoRPC{
			RequestDuration:  time.Second,
			RequestStartTime: sTime.UTC().Format(time.RFC3339Nano),
			RequestEncoding:  utils.MetaJSON,
			RequestID:        0,
			RequestMethod:    utils.AttributeSv1Ping,
			RequestParams:    `{}`,
			Reply:            `a`,
		}); err != nil {
		t.Fatal(err)
	}
	syntaxErr = nil
	if err = anz.V1StringQuery(context.Background(), &QueryArgs{Filters: []string{
		"*string:~*hdr.RequestMethod:" + utils.AttributeSv1Ping,
		"*string:~*opts.EventSource:*cdrs",
	}}, &reply); !errors.As(err, &syntaxErr) {
		t.Errorf("Expected JSON syntax error,received:%v", err)
	}

	reply = []map[string]any{}
	if err = anz.V1StringQuery(context.Background(), &QueryArgs{
		Filters: []string{"*string:~*hdr.RequestMethod:" + utils.CoreSv1Ping},
		Limit:   2,
	}, &reply); err != nil {
		t.Fatal(err)
	} else if len(reply) != 2 {
		t.Errorf("Expected 2 hits with Limit=2, received: %v", len(reply))
	}

	reply = []map[string]any{}
	if err = anz.V1StringQuery(context.Background(), &QueryArgs{
		Filters: []string{"*string:~*hdr.RequestMethod:" + utils.CoreSv1Ping},
		Offset:  2,
	}, &reply); err != nil {
		t.Fatal(err)
	} else if len(reply) != 2 {
		t.Errorf("Expected 2 hits with Offset=2 no Limit, received: %v", len(reply))
	}

	if err = anz.db.Close(); err != nil {
		t.Fatal(err)
	}
	if err = anz.V1StringQuery(context.Background(), &QueryArgs{Filters: []string{"*string:~*hdr.RequestEncoding:*gob"}}, &reply); err != bleve.ErrorIndexClosed {
		t.Errorf("Expected error: %v,received: %+v", bleve.ErrorIndexClosed, err)
	}
	if err := os.RemoveAll(cfg.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
	}
}

func TestAnalyzerSFilteredPagination(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.AnalyzerSCfg().IndexType = utils.MetaInternal
	idb, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: idb}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	dm.SetCache(engine.NewCacheS(cfg, dm, nil, nil))
	anz, err := NewAnalyzerS(cfg)
	if err != nil {
		t.Fatal(err)
	}
	anz.SetFilterS(engine.NewFilterS(cfg, nil, dm))
	t.Cleanup(func() {
		if err := anz.Shutdown(); err != nil {
			t.Error(err)
		}
	})

	start := time.Now()
	for i := 0; i < queryBatchSize+2; i++ {
		if err := anz.db.Index("batch-"+strconv.Itoa(10000+i), NewInfoRPC(uint64(i), "Batch.Test",
			map[string]any{"Match": i >= queryBatchSize}, nil, nil,
			utils.MetaJSON, "", "", start, start)); err != nil {
			t.Fatal(err)
		}
	}
	var reply []map[string]any
	if err := anz.V1StringQuery(context.Background(), &QueryArgs{
		Filters: []string{"*string:~*req.Match:true"},
		Limit:   1,
		Offset:  1,
	}, &reply); err != nil {
		t.Fatal(err)
	} else if len(reply) != 1 {
		t.Fatalf("expected one match across batches, got %d", len(reply))
	} else if got := utils.IfaceAsString(reply[0]["RequestID"]); got != strconv.Itoa(queryBatchSize+1) {
		t.Fatalf("got request ID %s, want %d", got, queryBatchSize+1)
	}

	anz.db = &indexWithInsert{
		baseIndex: anz.db,
		id:        "batch-00000",
		doc:       NewInfoRPC(999999, "Batch.Test", nil, nil, nil, utils.MetaJSON, "", "", start, start),
	}
	reply = nil
	if err := anz.V1StringQuery(context.Background(), &QueryArgs{
		Filters: []string{"*string:~*hdr.RequestMethod:Batch.Test"},
		Limit:   queryBatchSize + 1,
	}, &reply); err != nil {
		t.Fatal(err)
	} else if len(reply) != queryBatchSize+1 {
		t.Fatalf("expected %d results, got %d", queryBatchSize+1, len(reply))
	}
	for i, fields := range reply {
		if got := utils.IfaceAsString(fields["RequestID"]); got != strconv.Itoa(i) {
			t.Fatalf("result %d has request ID %s", i, got)
		}
	}
}

func TestAnalyzerSQueryOrder(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.AnalyzerSCfg().IndexType = utils.MetaInternal
	idb, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: idb}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	dm.SetCache(engine.NewCacheS(cfg, dm, nil, nil))
	anz, err := NewAnalyzerS(cfg)
	if err != nil {
		t.Fatal(err)
	}
	anz.SetFilterS(engine.NewFilterS(cfg, nil, dm))
	t.Cleanup(func() {
		if err := anz.Shutdown(); err != nil {
			t.Error(err)
		}
	})

	base := time.Date(2025, time.January, 15, 10, 30, 0, 123_456_789, time.UTC)
	for _, doc := range []struct {
		key   string
		id    uint64
		start time.Time
	}{
		{key: "old", id: 1, start: base},
		{key: "same-a", id: 2, start: base.Add(time.Hour)},
		{key: "same-b", id: 3, start: base.Add(time.Hour)},
		{key: "newest", id: 4, start: base.Add(2 * time.Hour)},
	} {
		if err := anz.db.Index(doc.key, NewInfoRPC(doc.id, "Order.Test", nil, nil, nil,
			utils.MetaJSON, "", "", doc.start, doc.start)); err != nil {
			t.Fatal(err)
		}
	}

	for _, tt := range []struct {
		name string
		args QueryArgs
		want []string
	}{
		{name: "unfiltered", args: QueryArgs{Limit: 4}, want: []string{"4", "2", "3", "1"}},
		{
			name: "filtered page",
			args: QueryArgs{Filters: []string{"*prefix:~*hdr.RequestMethod:Order."}, Limit: 2, Offset: 1},
			want: []string{"2", "3"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var reply []map[string]any
			if err := anz.V1StringQuery(context.Background(), &tt.args, &reply); err != nil {
				t.Fatal(err)
			}
			got := make([]string, len(reply))
			for i, fields := range reply {
				got[i] = utils.IfaceAsString(fields["RequestID"])
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAnalyzerSLogTrafficInternalDB(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()

	cfg.AnalyzerSCfg().DBPath = utils.EmptyString
	cfg.AnalyzerSCfg().IndexType = utils.MetaInternal
	cfg.AnalyzerSCfg().TTL = 30 * time.Minute
	anz, err := NewAnalyzerS(cfg)
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
}
