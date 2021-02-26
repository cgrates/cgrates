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

package ers

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestERsNewERService(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fltrS := &engine.FilterS{}
	expected := &ERService{cfg: cfg,
		filterS:   fltrS,
		rdrs:      make(map[string]EventReader),
		rdrPaths:  make(map[string]string),
		stopLsn:   make(map[string]chan struct{}),
		rdrEvents: make(chan *erEvent),
		rdrErr:    make(chan error),
	}
	rcv := NewERService(cfg, fltrS, nil)

	if !reflect.DeepEqual(expected.cfg, rcv.cfg) {
		t.Errorf("Expecting: <%+v>, received: <%+v>", expected.cfg, rcv.cfg)
	} else if !reflect.DeepEqual(expected.filterS, rcv.filterS) {
		t.Errorf("Expecting: <%+v>, received: <%+v>", expected.filterS, rcv.filterS)
	}
}

func TestERsAddReader(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fltrS := &engine.FilterS{}
	erS := NewERService(cfg, fltrS, nil)
	reader := cfg.ERsCfg().Readers[0]
	reader.Type = utils.MetaFileCSV
	reader.ID = "file_reader"
	reader.RunDelay = 0
	cfg.ERsCfg().Readers = append(cfg.ERsCfg().Readers, reader)
	if len(cfg.ERsCfg().Readers) != 2 {
		t.Errorf("Expecting: <2>, received: <%+v>", len(cfg.ERsCfg().Readers))
	}
	if err := erS.addReader("file_reader", 1); err != nil {
		t.Errorf("Expecting: <nil>, received: <%+v>", len(cfg.ERsCfg().Readers))
	} else if len(erS.rdrs) != 1 {
		t.Errorf("Expecting: <2>, received: <%+v>", len(erS.rdrs))
	} else if !reflect.DeepEqual(erS.rdrs["file_reader"].Config(), reader) {
		t.Errorf("Expecting: <%+v>, received: <%+v>", reader, erS.rdrs["file_reader"].Config())
	}
}

func TestERsListenAndServeErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers = []*config.EventReaderCfg{
		{
			ID:                       "",
			Type:                     "",
			RowLength:                0,
			FieldSep:                 "",
			HeaderDefineChar:         "",
			RunDelay:                 0,
			ConcurrentReqs:           0,
			SourcePath:               "",
			ProcessedPath:            "",
			Opts:                     nil,
			XMLRootPath:              nil,
			Tenant:                   nil,
			Timezone:                 "",
			Filters:                  nil,
			Flags:                    nil,
			FailedCallsPrefix:        "",
			PartialRecordCache:       0,
			PartialCacheExpiryAction: "",
			Fields:                   nil,
			CacheDumpFields:          nil,
		},
	}
	fltrS := &engine.FilterS{}
	srv := NewERService(cfg, fltrS, nil)
	stopChan := make(chan struct{}, 1)
	cfgRldChan := make(chan struct{}, 1)
	err := srv.ListenAndServe(stopChan, cfgRldChan)
	if err == nil || err.Error() != "unsupported reader type: <>" {
		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", "unsupported reader type: <>", err)
	}
}
func TestERsProcessEventErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers = []*config.EventReaderCfg{
		{
			ID:                       "",
			Type:                     "",
			RowLength:                0,
			FieldSep:                 "",
			HeaderDefineChar:         "",
			RunDelay:                 0,
			ConcurrentReqs:           0,
			SourcePath:               "",
			ProcessedPath:            "",
			Opts:                     nil,
			XMLRootPath:              nil,
			Tenant:                   nil,
			Timezone:                 "",
			Filters:                  nil,
			Flags:                    nil,
			FailedCallsPrefix:        "",
			PartialRecordCache:       0,
			PartialCacheExpiryAction: "",
			Fields:                   nil,
			CacheDumpFields:          nil,
		},
	}
	fltrS := &engine.FilterS{}
	srv := NewERService(cfg, fltrS, nil)
	rdrCfg := &config.EventReaderCfg{
		ID:                       "",
		Type:                     "",
		RowLength:                0,
		FieldSep:                 "",
		HeaderDefineChar:         "",
		RunDelay:                 0,
		ConcurrentReqs:           0,
		SourcePath:               "",
		ProcessedPath:            "",
		Opts:                     nil,
		XMLRootPath:              nil,
		Tenant:                   nil,
		Timezone:                 "",
		Filters:                  nil,
		Flags:                    nil,
		FailedCallsPrefix:        "",
		PartialRecordCache:       0,
		PartialCacheExpiryAction: "",
		Fields:                   nil,
		CacheDumpFields:          nil,
	}
	cgrEvent := &utils.CGREvent{
		Tenant: "",
		ID:     "",
		Time:   nil,
		Event:  nil,
		Opts:   nil,
	}
	err := srv.processEvent(cgrEvent, rdrCfg)
	if err == nil || err.Error() != "unsupported reqType: <>" {
		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", "unsupported reqType: <>", err)
	}
}

func TestERsCloseAllRdrs(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers = []*config.EventReaderCfg{
		{
			ID:                       "",
			Type:                     "",
			RowLength:                0,
			FieldSep:                 "",
			HeaderDefineChar:         "",
			RunDelay:                 0,
			ConcurrentReqs:           0,
			SourcePath:               "",
			ProcessedPath:            "",
			Opts:                     nil,
			XMLRootPath:              nil,
			Tenant:                   nil,
			Timezone:                 "",
			Filters:                  nil,
			Flags:                    nil,
			FailedCallsPrefix:        "",
			PartialRecordCache:       0,
			PartialCacheExpiryAction: "",
			Fields:                   nil,
			CacheDumpFields:          nil,
		},
	}
	fltrS := &engine.FilterS{}
	srv := NewERService(cfg, fltrS, nil)
	srv.stopLsn[""] = make(chan struct{}, 1)
	srv.closeAllRdrs()
}
func TestERsListenAndServeRdrErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers = []*config.EventReaderCfg{
		{
			ID:                       "",
			Type:                     utils.MetaNone,
			RowLength:                0,
			FieldSep:                 "",
			HeaderDefineChar:         "",
			RunDelay:                 0,
			ConcurrentReqs:           0,
			SourcePath:               "",
			ProcessedPath:            "",
			Opts:                     nil,
			XMLRootPath:              nil,
			Tenant:                   nil,
			Timezone:                 "",
			Filters:                  nil,
			Flags:                    nil,
			FailedCallsPrefix:        "",
			PartialRecordCache:       0,
			PartialCacheExpiryAction: "",
			Fields:                   nil,
			CacheDumpFields:          nil,
		},
	}
	fltrS := &engine.FilterS{}
	srv := NewERService(cfg, fltrS, nil)
	stopChan := make(chan struct{}, 1)
	cfgRldChan := make(chan struct{}, 1)
	srv.rdrErr = make(chan error, 1)
	srv.rdrErr <- utils.ErrNotFound
	time.Sleep(30 * time.Millisecond)
	err := srv.ListenAndServe(stopChan, cfgRldChan)
	if err == nil || err != utils.ErrNotFound {
		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", utils.ErrNotFound, err)
	}
}

func TestERsListenAndServeStopchan(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers = []*config.EventReaderCfg{
		{
			ID:                       "",
			Type:                     utils.MetaNone,
			RowLength:                0,
			FieldSep:                 "",
			HeaderDefineChar:         "",
			RunDelay:                 0,
			ConcurrentReqs:           0,
			SourcePath:               "",
			ProcessedPath:            "",
			Opts:                     nil,
			XMLRootPath:              nil,
			Tenant:                   nil,
			Timezone:                 "",
			Filters:                  nil,
			Flags:                    nil,
			FailedCallsPrefix:        "",
			PartialRecordCache:       0,
			PartialCacheExpiryAction: "",
			Fields:                   nil,
			CacheDumpFields:          nil,
		},
	}
	fltrS := &engine.FilterS{}
	srv := NewERService(cfg, fltrS, nil)
	stopChan := make(chan struct{}, 1)
	cfgRldChan := make(chan struct{}, 1)
	stopChan <- struct{}{}
	time.Sleep(30 * time.Millisecond)
	err := srv.ListenAndServe(stopChan, cfgRldChan)
	if err != nil {
		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", nil, err)
	}
}

/*
func TestERsListenAndServeRdrEvents(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers = []*config.EventReaderCfg{
		{
			ID:                       "",
			Type:                     utils.MetaNone,
			RowLength:                0,
			FieldSep:                 "",
			HeaderDefineChar:         "",
			RunDelay:                 0,
			ConcurrentReqs:           0,
			SourcePath:               "",
			ProcessedPath:            "",
			Opts:                     nil,
			XMLRootPath:              nil,
			Tenant:                   nil,
			Timezone:                 "",
			Filters:                  nil,
			Flags:                    nil,
			FailedCallsPrefix:        "",
			PartialRecordCache:       0,
			PartialCacheExpiryAction: "",
			Fields:                   nil,
			CacheDumpFields:          nil,
		},
	}
	fltrS := &engine.FilterS{}
	srv := NewERService(cfg, fltrS, nil)
	stopChan := make(chan struct{}, 1)
	cfgRldChan := make(chan struct{}, 1)
	srv.rdrEvents = make(chan *erEvent, 1)
	srv.rdrEvents <- &erEvent{
		cgrEvent: &utils.CGREvent{
			Tenant: "",
			ID:     "",
			Time:   nil,
			Event:  nil,
			Opts:   nil,
		},
		rdrCfg: &config.EventReaderCfg{
			ID:                       "",
			Type:                     "",
			RowLength:                0,
			FieldSep:                 "",
			HeaderDefineChar:         "",
			RunDelay:                 0,
			ConcurrentReqs:           0,
			SourcePath:               "",
			ProcessedPath:            "",
			Opts:                     nil,
			XMLRootPath:              nil,
			Tenant:                   nil,
			Timezone:                 "",
			Filters:                  nil,
			Flags:                    nil,
			FailedCallsPrefix:        "",
			PartialRecordCache:       0,
			PartialCacheExpiryAction: "",
			Fields:                   nil,
			CacheDumpFields:          nil,
		},
	}
	time.Sleep(30 * time.Millisecond)
	err := srv.ListenAndServe(stopChan, cfgRldChan)
	if err != nil {
		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", nil, err)
	}
}
*/
