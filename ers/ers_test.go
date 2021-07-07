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
	erS := NewERService(cfg, nil, nil)
	event := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EventERsProcessPartial",
		Event: map[string]interface{}{
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
		Opts:           make(map[string]interface{}),
	}

	args := &erEvent{
		cgrEvent: event,
		rdrCfg:   rdrCfg,
	}
	if err := erS.processPartialEvent(args.cgrEvent, args.rdrCfg); err != nil {
		t.Error(err)
	} else {
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
				Event: map[string]interface{}{
					utils.AccountField: "1001",
				},
			},
		},
		rdrCfg: &config.EventReaderCfg{
			ID:            "ER1",
			Type:          utils.MetaNone,
			ProcessedPath: "/tmp",
			Opts: map[string]interface{}{
				utils.PartialCacheActionOpt: utils.MetaPostCDR,
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
				Event: map[string]interface{}{
					utils.AnswerTime:   time.Date(2021, 6, 1, 12, 0, 0, 0, time.UTC),
					utils.AccountField: "1001",
					utils.Destination:  "1002",
				},
			},
			{
				Tenant: "cgrates.org",
				ID:     "EventErsOnEvicted",
				Event: map[string]interface{}{
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
			Opts: map[string]interface{}{
				utils.PartialCacheActionOpt: utils.MetaPostCDR,
			},
		},
	}
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
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
				Event: map[string]interface{}{
					utils.AccountField: "1001",
				},
			},
		},
		rdrCfg: &config.EventReaderCfg{
			ID:   "ER1",
			Type: utils.MetaNone,
			Opts: map[string]interface{}{
				utils.PartialCacheActionOpt: utils.MetaDumpToFile,
				utils.PartialPathOpt:        dirPath,
			},
			CacheDumpFields: []*config.FCTemplate{
				{
					Tag: "cacheDump",
				},
			},
		},
	}
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
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
				Event: map[string]interface{}{
					utils.AnswerTime:   time.Date(2021, 6, 1, 12, 0, 0, 0, time.UTC),
					utils.AccountField: "1001",
					utils.Destination:  "1002",
				},
			},
			{
				Tenant: "cgrates.org",
				ID:     "EventErsOnEvicted",
				Event: map[string]interface{}{
					utils.AnswerTime:   time.Date(2021, 6, 1, 13, 0, 0, 0, time.UTC),
					utils.AccountField: "1001",
					utils.Destination:  "1003",
				},
			},
		},
		rdrCfg: &config.EventReaderCfg{
			ID:   "ER1",
			Type: utils.MetaNone,
			Opts: map[string]interface{}{
				utils.PartialCacheActionOpt: utils.MetaDumpToFile,
				utils.PartialPathOpt:        dirPath,
			},
		},
	}
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
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
				Event: map[string]interface{}{
					utils.AccountField: "1001",
				},
			},
		},
		rdrCfg: &config.EventReaderCfg{
			ID:   "ER1",
			Type: utils.MetaNone,
			Opts: map[string]interface{}{
				utils.PartialCacheActionOpt: utils.MetaDumpToFile,
			},
		},
	}
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
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
