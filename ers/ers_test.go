// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ers

import (
	"bytes"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestERsProcessPartialEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	erS := NewERService(cfg, nil, nil, nil)
	event := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EventERsProcessPartial",
		Event: map[string]any{
			utils.OriginID: "originID",
		},
	}
	rdrCfg := &config.EventReaderCfg{
		ID:             utils.MetaDefault,
		Type:           utils.MetaNone,
		RunDelay:       0,
		ConcurrentReqs: 0,
		SourcePath:     "/var/spool/cgrates/ers/in",
		ProcessedPath:  "/var/spool/cgrates/ers/out",
		Filters:        []string{},
		Opts:           &config.EventReaderOpts{},
	}

	args := &erEvent{
		cgrEvent: event,
		rdrCfg:   rdrCfg,
	}
	rcv, err := erS.processPartialEvent(args.cgrEvent, args.rdrCfg)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(rcv, args.cgrEvent) {
		rcv := <-erS.rdrEvents
		if !reflect.DeepEqual(rcv, args) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", args, rcv)
		}
	}
}

func TestErsOnEvictedNilValue(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	erS := &ERService{
		cfg:       cfg,
		rdrEvents: make(chan *erEvent, 1),
	}
	erS.onEvicted("id", nil)

	// Verification TBA
}

func TestErsOnEvictedMetaPostCDROK(t *testing.T) {
	value := &erEvents{
		events: []*utils.CGREvent{
			{
				Tenant: "cgrates.org",
				ID:     "EventErsOnEvicted",
				Event: map[string]any{
					utils.AccountField: "1001",
				},
			},
		},
		rdrCfg: &config.EventReaderCfg{
			ID:            "ER1",
			Type:          utils.MetaNone,
			ProcessedPath: "/tmp",
			Opts: &config.EventReaderOpts{
				PartialCacheAction: utils.StringPointer(utils.MetaPostCDR),
			},
		},
	}
	cfg := config.NewDefaultCGRConfig()
	erS := &ERService{
		cfg:       cfg,
		rdrEvents: make(chan *erEvent, 1),
	}
	erS.onEvicted("id", value)

	if len(erS.rdrEvents) != 1 {
		t.Fatal("Expected channel to contain a value")
	}
	select {
	case data := <-erS.rdrEvents:
		if !reflect.DeepEqual(data.rdrCfg, value.rdrCfg) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(value.rdrCfg), utils.ToJSON(data.rdrCfg))
		}
		if !reflect.DeepEqual(data.cgrEvent, value.events[0]) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(value.events[0]), utils.ToJSON(data.cgrEvent))
		}
	case <-time.After(40 * time.Millisecond):
		t.Error("Time limit exceeded")
	}
}

func TestErsOnEvictedMetaPostCDRMergeErr(t *testing.T) {
	utils.Logger.SetLogLevel(4)
	utils.Logger.SetSyslog(nil)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	value := &erEvents{
		events: []*utils.CGREvent{
			{
				Tenant: "cgrates.org",
				ID:     "EventErsOnEvicted",
				Event: map[string]any{
					utils.AnswerTime:   time.Date(2021, 6, 1, 12, 0, 0, 0, time.UTC),
					utils.AccountField: "1001",
					utils.Destination:  "1002",
				},
			},
			{
				Tenant: "cgrates.org",
				ID:     "EventErsOnEvicted",
				Event: map[string]any{
					utils.AnswerTime:   time.Date(2021, 6, 1, 13, 0, 0, 0, time.UTC),
					utils.AccountField: "1001",
					utils.Destination:  "1003",
				},
			},
		},
		rdrCfg: &config.EventReaderCfg{
			ID:            "ER1",
			Type:          utils.MetaNone,
			ProcessedPath: "/tmp",
			Opts: &config.EventReaderOpts{
				PartialCacheAction: utils.StringPointer(utils.MetaPostCDR),
			},
		},
	}
	cfg := config.NewDefaultCGRConfig()
	data, err := engine.NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrS := engine.NewFilterS(cfg, nil, dm)
	erS := &ERService{
		cfg:       cfg,
		rdrEvents: make(chan *erEvent, 1),
		filterS:   fltrS,
	}
	expLog := `[WARNING] <ERs> failed posting expired parial events <[{"Tenant":"cgrates.org","ID":"EventErsOnEvicted","Time":null,"Event":{"Account":"1001","AnswerTime":"2021-06-01T13:00:00Z","Destination":"1003"},"APIOpts":null},{"Tenant":"cgrates.org","ID":"EventErsOnEvicted","Time":null,"Event":{"Account":"1001","AnswerTime":"2021-06-01T12:00:00Z","Destination":"1002"},"APIOpts":null}]> due error <unsupported comparison type: string, kind: string>`
	erS.onEvicted("id", value)
	rcvLog := buf.String()[20:]
	if !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected: <%+v> to be included in <%+v>", expLog, rcvLog)
	}

	utils.Logger.SetLogLevel(0)
}

func TestErsOnEvictedMetaDumpToFileSetFieldsErr(t *testing.T) {
	utils.Logger.SetLogLevel(4)
	utils.Logger.SetSyslog(nil)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	dirPath := "/tmp/TestErsOnEvictedMetaDumpToFile"
	value := &erEvents{
		events: []*utils.CGREvent{
			{
				Tenant: "cgrates.org",
				ID:     "EventErsOnEvicted",
				Event: map[string]any{
					utils.AccountField: "1001",
				},
			},
		},
		rdrCfg: &config.EventReaderCfg{
			ID:   "ER1",
			Type: utils.MetaNone,
			Opts: &config.EventReaderOpts{
				PartialCacheAction: utils.StringPointer(utils.MetaDumpToFile),
				PartialPath:        utils.StringPointer(dirPath),
			},
			CacheDumpFields: []*config.FCTemplate{
				{
					Tag: "cacheDump",
				},
			},
		},
	}
	cfg := config.NewDefaultCGRConfig()
	data, err := engine.NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrS := engine.NewFilterS(cfg, nil, dm)
	erS := &ERService{
		cfg:       cfg,
		rdrEvents: make(chan *erEvent, 1),
		filterS:   fltrS,
	}
	expLog := `[WARNING] <ERs> Converting CDR with CGRID: <ID> to record , ignoring due to error: <unsupported type: <>>
`
	erS.onEvicted("ID", value)

	rcvLog := buf.String()[20:]
	if !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected <%+v> to be included in: <%+v>", expLog, rcvLog)
	}

	utils.Logger.SetLogLevel(0)
}

func TestErsOnEvictedMetaDumpToFileMergeErr(t *testing.T) {
	utils.Logger.SetLogLevel(4)
	utils.Logger.SetSyslog(nil)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	dirPath := "/tmp/TestErsOnEvictedMetaDumpToFile"
	value := &erEvents{
		events: []*utils.CGREvent{
			{
				Tenant: "cgrates.org",
				ID:     "EventErsOnEvicted",
				Event: map[string]any{
					utils.AnswerTime:   time.Date(2021, 6, 1, 12, 0, 0, 0, time.UTC),
					utils.AccountField: "1001",
					utils.Destination:  "1002",
				},
			},
			{
				Tenant: "cgrates.org",
				ID:     "EventErsOnEvicted",
				Event: map[string]any{
					utils.AnswerTime:   time.Date(2021, 6, 1, 13, 0, 0, 0, time.UTC),
					utils.AccountField: "1001",
					utils.Destination:  "1003",
				},
			},
		},
		rdrCfg: &config.EventReaderCfg{
			ID:   "ER1",
			Type: utils.MetaNone,
			Opts: &config.EventReaderOpts{
				PartialCacheAction: utils.StringPointer(utils.MetaDumpToFile),
				PartialPath:        utils.StringPointer(dirPath),
			},
		},
	}
	cfg := config.NewDefaultCGRConfig()
	data, err := engine.NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrS := engine.NewFilterS(cfg, nil, dm)
	erS := &ERService{
		cfg:       cfg,
		rdrEvents: make(chan *erEvent, 1),
		filterS:   fltrS,
	}

	expLog := `[WARNING] <ERs> failed posting expired parial events <[{"Tenant":"cgrates.org","ID":"EventErsOnEvicted","Time":null,"Event":{"Account":"1001","AnswerTime":"2021-06-01T13:00:00Z","Destination":"1003"},"APIOpts":null},{"Tenant":"cgrates.org","ID":"EventErsOnEvicted","Time":null,"Event":{"Account":"1001","AnswerTime":"2021-06-01T12:00:00Z","Destination":"1002"},"APIOpts":null}]> due error <unsupported comparison type: string, kind: string>
`
	erS.onEvicted("ID", value)

	rcvLog := buf.String()[20:]
	if !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected <%+v> to be included in: <%+v>", expLog, rcvLog)
	}

	utils.Logger.SetLogLevel(0)
}

func TestErsOnEvictedMetaDumpToFileEmptyPath(t *testing.T) {
	value := &erEvents{
		events: []*utils.CGREvent{
			{
				Tenant: "cgrates.org",
				ID:     "EventErsOnEvicted",
				Event: map[string]any{
					utils.AccountField: "1001",
				},
			},
		},
		rdrCfg: &config.EventReaderCfg{
			ID:   "ER1",
			Type: utils.MetaNone,
			Opts: &config.EventReaderOpts{
				PartialCacheAction: utils.StringPointer(utils.MetaDumpToFile),
			},
		},
	}
	cfg := config.NewDefaultCGRConfig()
	data, err := engine.NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrS := engine.NewFilterS(cfg, nil, dm)
	erS := &ERService{
		cfg:       cfg,
		rdrEvents: make(chan *erEvent, 1),
		filterS:   fltrS,
	}
	erS.onEvicted("ID", value)

	// Verification TBA
}
