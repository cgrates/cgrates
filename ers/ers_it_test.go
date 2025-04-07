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

package ers

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/rpcclient"

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
	rcv := NewERService(cfg, nil, fltrS, nil)

	if !reflect.DeepEqual(expected.cfg, rcv.cfg) {
		t.Errorf("Expecting: <%+v>, received: <%+v>", expected.cfg, rcv.cfg)
	} else if !reflect.DeepEqual(expected.filterS, rcv.filterS) {
		t.Errorf("Expecting: <%+v>, received: <%+v>", expected.filterS, rcv.filterS)
	}
}

func TestERsAddReader(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fltrS := &engine.FilterS{}
	erS := NewERService(cfg, nil, fltrS, nil)
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
		{},
	}
	fltrS := &engine.FilterS{}
	srv := NewERService(cfg, nil, fltrS, nil)
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
			ID:   "",
			Type: "",
		},
	}
	fltrS := &engine.FilterS{}
	srv := NewERService(cfg, nil, fltrS, nil)
	rdrCfg := &config.EventReaderCfg{
		ID:   "",
		Type: "",
	}
	cgrEvent := &utils.CGREvent{
		Tenant: "",
		ID:     "",
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
			ID:   "",
			Type: "",
		},
	}
	fltrS := &engine.FilterS{}
	srv := NewERService(cfg, nil, fltrS, nil)
	srv.stopLsn[""] = make(chan struct{}, 1)
	srv.closeAllRdrs()
}
func TestERsListenAndServeRdrErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers = []*config.EventReaderCfg{
		{
			Type: utils.MetaNone,
		},
	}
	fltrS := &engine.FilterS{}
	srv := NewERService(cfg, nil, fltrS, nil)
	stopChan := make(chan struct{}, 1)
	cfgRldChan := make(chan struct{}, 1)
	srv.rdrErr = make(chan error, 1)
	srv.rdrErr <- utils.ErrNotFound
	time.Sleep(10 * time.Millisecond)
	err := srv.ListenAndServe(stopChan, cfgRldChan)
	if err == nil || err != utils.ErrNotFound {
		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", utils.ErrNotFound, err)
	}
}

func TestERsListenAndServeStopchan(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers = []*config.EventReaderCfg{
		{
			Type: utils.MetaNone,
		},
	}
	fltrS := &engine.FilterS{}
	srv := NewERService(cfg, nil, fltrS, nil)
	stopChan := make(chan struct{}, 1)
	cfgRldChan := make(chan struct{}, 1)
	stopChan <- struct{}{}
	time.Sleep(10 * time.Millisecond)
	err := srv.ListenAndServe(stopChan, cfgRldChan)
	if err != nil {
		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", nil, err)
	}
}

func TestERsListenAndServeRdrEvents(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers = []*config.EventReaderCfg{
		{
			Type: utils.MetaNone,
		},
	}
	fltrS := &engine.FilterS{}
	srv := NewERService(cfg, nil, fltrS, nil)
	stopChan := make(chan struct{}, 1)
	cfgRldChan := make(chan struct{}, 1)
	srv.rdrErr = make(chan error, 1)
	srv.rdrEvents = make(chan *erEvent, 1)
	srv.rdrEvents <- &erEvent{
		cgrEvent: &utils.CGREvent{},
		rdrCfg:   &config.EventReaderCfg{},
	}
	go func() {
		time.Sleep(10 * time.Millisecond)
		srv.rdrErr <- utils.ErrNotFound
	}()
	err := srv.ListenAndServe(stopChan, cfgRldChan)
	if err == nil || err != utils.ErrNotFound {
		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", utils.ErrNotFound, err)
	}
}

func TestERsListenAndServeCfgRldChan(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers = []*config.EventReaderCfg{
		{
			Type: utils.MetaNone,
		},
	}
	fltrS := &engine.FilterS{}
	srv := NewERService(cfg, nil, fltrS, nil)
	stopChan := make(chan struct{}, 1)
	cfgRldChan := make(chan struct{}, 1)
	srv.rdrErr = make(chan error, 1)
	cfgRldChan <- struct{}{}
	go func() {
		time.Sleep(10 * time.Millisecond)
		srv.rdrErr <- utils.ErrNotFound
	}()
	err := srv.ListenAndServe(stopChan, cfgRldChan)
	if err == nil || err != utils.ErrNotFound {
		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", utils.ErrNotFound, err)
	}
}

func TestERsListenAndServeCfgRldChan2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers = []*config.EventReaderCfg{
		{
			ID:   "test",
			Type: utils.MetaNone,
		},
	}
	fltrS := &engine.FilterS{}
	srv := NewERService(cfg, nil, fltrS, nil)
	exp := &CSVFileER{
		cgrCfg: cfg,
		cfgIdx: 0,
	}
	var expected EventReader = exp
	srv.rdrs = map[string]EventReader{
		"test": expected,
	}
	srv.stopLsn["test"] = make(chan struct{})
	stopChan := make(chan struct{}, 1)
	cfgRldChan := make(chan struct{}, 1)
	srv.rdrErr = make(chan error, 1)

	cfgRldChan <- struct{}{}
	go func() {
		time.Sleep(10 * time.Millisecond)
		srv.rdrErr <- utils.ErrNotFound
	}()
	err := srv.ListenAndServe(stopChan, cfgRldChan)
	if err == nil || err != utils.ErrNotFound {
		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", utils.ErrNotFound, err)
	}
}

func TestERsListenAndServeCfgRldChan3(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers = []*config.EventReaderCfg{
		{
			ID:   "test",
			Type: utils.MetaNone,
		},
	}
	fltrS := &engine.FilterS{}
	srv := NewERService(cfg, nil, fltrS, nil)
	exp := &CSVFileER{
		cgrCfg: cfg,
		cfgIdx: 0,
	}
	var expected EventReader = exp
	srv.rdrs = map[string]EventReader{
		"test": expected,
	}
	srv.stopLsn["test"] = make(chan struct{})
	stopChan := make(chan struct{}, 1)
	cfgRldChan := make(chan struct{}, 1)

	cfgRldChan <- struct{}{}
	go func() {
		time.Sleep(10 * time.Millisecond)
		close(stopChan)
	}()
	err := srv.ListenAndServe(stopChan, cfgRldChan)
	if err != nil {
		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", nil, err)
	}
}

func TestERsListenAndServeCfgRldChan4(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers = []*config.EventReaderCfg{
		{
			ID:   "test",
			Type: utils.MetaNone,
		},
	}
	fltrS := &engine.FilterS{}
	srv := NewERService(cfg, nil, fltrS, nil)
	exp := &CSVFileER{
		cgrCfg: cfg,
		cfgIdx: 0,
	}
	var evRdr EventReader = exp
	srv.rdrs = map[string]EventReader{
		"test": evRdr,
	}
	srv.stopLsn["test"] = make(chan struct{})
	stopChan := make(chan struct{}, 1)
	cfgRldChan := make(chan struct{}, 1)
	srv.rdrPaths = map[string]string{
		"test": "path_test",
	}
	cfgRldChan <- struct{}{}
	go func() {
		time.Sleep(10 * time.Millisecond)
		close(stopChan)
	}()
	err := srv.ListenAndServe(stopChan, cfgRldChan)
	if err != nil {
		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", nil, err)
	}
}

func TestERsListenAndServeCfgRldChan5(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers = []*config.EventReaderCfg{
		{
			ID:   "test",
			Type: utils.MetaFileCSV,
		},
	}
	fltrS := &engine.FilterS{}
	srv := NewERService(cfg, nil, fltrS, nil)
	exp := &CSVFileER{
		cgrCfg: cfg,
	}
	var evRdr EventReader = exp
	srv.rdrs = map[string]EventReader{
		"test": evRdr,
	}
	srv.stopLsn["test"] = make(chan struct{})
	stopChan := make(chan struct{}, 1)
	cfgRldChan := make(chan struct{}, 1)
	srv.rdrPaths = map[string]string{
		"test": "path_test",
	}
	cfgRldChan <- struct{}{}
	go func() {
		time.Sleep(10 * time.Millisecond)
		close(stopChan)
	}()
	err := srv.ListenAndServe(stopChan, cfgRldChan)
	if err != nil {
		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", nil, err)
	}
}

func TestERsListenAndServeCfgRldChan6(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers = []*config.EventReaderCfg{
		{
			ID:   "test",
			Type: utils.MetaFileCSV,
		},
	}
	fltrS := &engine.FilterS{}
	srv := NewERService(cfg, nil, fltrS, nil)
	exp := &CSVFileER{
		cgrCfg: cfg,
		cfgIdx: 0,
	}
	var evRdr EventReader = exp
	srv.rdrs = map[string]EventReader{
		"test": evRdr,
	}
	srv.stopLsn["test"] = make(chan struct{})
	stopChan := make(chan struct{}, 1)
	cfgRldChan := make(chan struct{}, 1)
	srv.rdrPaths = map[string]string{
		"test": "path_test",
	}
	go func() {
		time.Sleep(10 * time.Millisecond)
		cfg.ERsCfg().Readers = []*config.EventReaderCfg{
			{
				ID:   "test",
				Type: "BadType",
			},
		}
		cfgRldChan <- struct{}{}
	}()
	err := srv.ListenAndServe(stopChan, cfgRldChan)
	if err == nil || err.Error() != "unsupported reader type: <BadType>" {
		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", "unsupported reader type: <BadType>", err)
	}
}

func TestERsProcessEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers = []*config.EventReaderCfg{
		{
			ID:   "test",
			Type: utils.MetaNone,
		},
	}
	fltrS := &engine.FilterS{}
	srv := NewERService(cfg, nil, fltrS, nil)
	rdrCfg := &config.EventReaderCfg{
		Flags: map[string]utils.FlagParams{
			utils.MetaLog: map[string][]string{
				"test": {"test"},
			},
		},
	}
	cgrEvent := &utils.CGREvent{}
	err := srv.processEvent(cgrEvent, rdrCfg)
	if err == nil || err.Error() != "unsupported reqType: <>" {
		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", "unsupported reqType: <>", err)
	}
}
func TestERsProcessEvent2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers = []*config.EventReaderCfg{
		{
			ID:   "test",
			Type: utils.MetaNone,
		},
	}
	fltrS := &engine.FilterS{}
	srv := NewERService(cfg, nil, fltrS, nil)
	rdrCfg := &config.EventReaderCfg{
		Flags: map[string]utils.FlagParams{
			utils.MetaDryRun: map[string][]string{
				"test": {"test"},
			},
		},
	}
	cgrEvent := &utils.CGREvent{}
	err := srv.processEvent(cgrEvent, rdrCfg)
	if err != nil {
		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", nil, err)
	}
}
func TestERsProcessEvent3(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers = []*config.EventReaderCfg{
		{
			ID:   "test",
			Type: utils.MetaNone,
		},
	}
	cfg.ERsCfg().SessionSConns = []string{}
	fltrS := &engine.FilterS{}
	srv := NewERService(cfg, nil, fltrS, nil)
	rdrCfg := &config.EventReaderCfg{
		Flags: map[string]utils.FlagParams{
			utils.MetaEvent: map[string][]string{},
		},
	}
	cgrEvent := &utils.CGREvent{
		APIOpts: map[string]any{
			utils.OptsRoutesLimit: true,
		},
	}
	err := srv.processEvent(cgrEvent, rdrCfg)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [connIDs]" {
		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", nil, err)
	}
}

func TestERsProcessEvent4(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers = []*config.EventReaderCfg{
		{
			ID:   "test",
			Type: utils.MetaNone,
		},
	}
	cfg.ERsCfg().SessionSConns = []string{}
	fltrS := &engine.FilterS{}
	srv := NewERService(cfg, nil, fltrS, nil)
	rdrCfg := &config.EventReaderCfg{
		Flags: map[string]utils.FlagParams{
			utils.MetaAuthorize: map[string][]string{},
		},
	}
	cgrEvent := &utils.CGREvent{
		APIOpts: map[string]any{
			utils.OptsRoutesLimit: true,
		},
	}
	err := srv.processEvent(cgrEvent, rdrCfg)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [connIDs]" {
		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", nil, err)
	}
}

func TestERsProcessEvent5(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers = []*config.EventReaderCfg{
		{
			ID:   "test",
			Type: utils.MetaNone,
		},
	}
	cfg.ERsCfg().SessionSConns = []string{}
	fltrS := &engine.FilterS{}
	srv := NewERService(cfg, nil, fltrS, nil)
	rdrCfg := &config.EventReaderCfg{
		Flags: map[string]utils.FlagParams{
			utils.MetaTerminate: map[string][]string{},
		},
	}
	cgrEvent := &utils.CGREvent{
		APIOpts: map[string]any{
			utils.OptsRoutesLimit: true,
		},
	}
	err := srv.processEvent(cgrEvent, rdrCfg)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [connIDs]" {
		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", nil, err)
	}
}

func TestERsProcessEvent6(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers = []*config.EventReaderCfg{
		{
			ID:   "test",
			Type: utils.MetaNone,
		},
	}
	cfg.ERsCfg().SessionSConns = []string{}
	fltrS := &engine.FilterS{}
	srv := NewERService(cfg, nil, fltrS, nil)
	rdrCfg := &config.EventReaderCfg{
		Flags: map[string]utils.FlagParams{
			utils.MetaInitiate: map[string][]string{},
		},
	}
	cgrEvent := &utils.CGREvent{
		APIOpts: map[string]any{
			utils.OptsRoutesLimit: true,
		},
	}
	err := srv.processEvent(cgrEvent, rdrCfg)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [connIDs]" {
		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", nil, err)
	}
}
func TestERsProcessEvent7(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers = []*config.EventReaderCfg{
		{
			ID:   "test",
			Type: utils.MetaNone,
		},
	}
	cfg.ERsCfg().SessionSConns = []string{}
	fltrS := &engine.FilterS{}
	srv := NewERService(cfg, nil, fltrS, nil)
	rdrCfg := &config.EventReaderCfg{
		Flags: map[string]utils.FlagParams{
			utils.MetaUpdate: map[string][]string{},
		},
	}
	cgrEvent := &utils.CGREvent{
		APIOpts: map[string]any{
			utils.OptsRoutesLimit: true,
		},
	}
	err := srv.processEvent(cgrEvent, rdrCfg)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [connIDs]" {
		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", nil, err)
	}
}
func TestERsProcessEvent8(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers = []*config.EventReaderCfg{
		{
			ID:   "test",
			Type: utils.MetaNone,
		},
	}
	cfg.ERsCfg().SessionSConns = []string{}
	fltrS := &engine.FilterS{}
	srv := NewERService(cfg, nil, fltrS, nil)
	rdrCfg := &config.EventReaderCfg{
		Flags: map[string]utils.FlagParams{
			utils.MetaMessage: map[string][]string{},
		},
	}
	cgrEvent := &utils.CGREvent{
		APIOpts: map[string]any{
			utils.OptsRoutesLimit: true,
		},
	}
	err := srv.processEvent(cgrEvent, rdrCfg)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [connIDs]" {
		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", nil, err)
	}
}

func TestERsProcessEvent9(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers = []*config.EventReaderCfg{
		{
			ID:   "test",
			Type: utils.MetaNone,
		},
	}
	cfg.ERsCfg().SessionSConns = []string{}
	fltrS := &engine.FilterS{}
	srv := NewERService(cfg, nil, fltrS, nil)
	rdrCfg := &config.EventReaderCfg{
		Flags: map[string]utils.FlagParams{
			utils.MetaCDRs: map[string][]string{},
		},
	}
	cgrEvent := &utils.CGREvent{
		APIOpts: map[string]any{
			utils.OptsRoutesLimit: true,
		},
	}
	err := srv.processEvent(cgrEvent, rdrCfg)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [connIDs]" {
		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", "MANDATORY_IE_MISSING: [connIDs]", err)
	}
}

func TestERsProcessEvent10(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers = []*config.EventReaderCfg{
		{
			ID:   "test",
			Type: utils.MetaNone,
		},
	}
	cfg.ERsCfg().SessionSConns = []string{}
	fltrS := &engine.FilterS{}
	srv := NewERService(cfg, nil, fltrS, nil)
	rdrCfg := &config.EventReaderCfg{
		Flags: map[string]utils.FlagParams{
			utils.MetaMessage:  map[string][]string{},
			utils.MetaAccounts: map[string][]string{},
		},
	}
	cgrEvent := &utils.CGREvent{
		Event: map[string]any{
			utils.Usage: time.Second,
		},
		APIOpts: map[string]any{
			utils.OptsRoutesLimit: true,
		},
	}
	err := srv.processEvent(cgrEvent, rdrCfg)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [connIDs]" {
		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", "MANDATORY_IE_MISSING: [connIDs]", err)
	}
}

type testMockClients struct {
	calls map[string]func(args any, reply any) error
}

func (sT *testMockClients) Call(ctx *context.Context, method string, arg any, rply any) error {
	if call, has := sT.calls[method]; !has {
		return rpcclient.ErrUnsupporteServiceMethod
	} else {
		return call(arg, rply)
	}
}

func TestERsProcessEvent11(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers = []*config.EventReaderCfg{
		{
			ID:   "test",
			Type: utils.MetaNone,
		},
	}
	cfg.ERsCfg().SessionSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)}
	fltrS := &engine.FilterS{}
	testMockClient := &testMockClients{
		calls: map[string]func(args any, reply any) error{
			utils.SessionSv1ProcessMessage: func(args any, reply any) error {
				return errors.New("RALS_ERROR")
			},
		},
	}
	clientChan := make(chan birpc.ClientConnector, 1)
	clientChan <- testMockClient
	connMng := engine.NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS): clientChan,
	})
	srv := NewERService(cfg, nil, fltrS, connMng)
	rdrCfg := &config.EventReaderCfg{
		Flags: map[string]utils.FlagParams{
			utils.MetaMessage: map[string][]string{},
		},
	}
	cgrEvent := &utils.CGREvent{
		Event: map[string]any{
			utils.Usage: 0,
		},
		APIOpts: map[string]any{
			utils.OptsRoutesLimit: true,
		},
	}
	err := srv.processEvent(cgrEvent, rdrCfg)
	if err == nil || err.Error() != "RALS_ERROR" {
		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", "RALS_ERROR", err)
	}
}

func TestErsOnEvictedMetaDumpToFileOK(t *testing.T) {
	dirPath := "/tmp/TestErsOnEvictedMetaDumpToFile"
	err := os.Mkdir(dirPath, 0755)
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(dirPath)

	val1 := config.NewRSRParsersMustCompile("TestTenant", ",")
	val2 := config.NewRSRParsersMustCompile("1001", ",")
	val3 := config.NewRSRParsersMustCompile("1002", ",")
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
				CSV:                &config.CSVROpts{},
			},
			CacheDumpFields: []*config.FCTemplate{
				{
					Tag:   "Tenant",
					Type:  utils.MetaConstant,
					Path:  "*exp.Tenant",
					Value: val1,
				},
				{
					Tag:   "Account",
					Type:  utils.MetaConstant,
					Path:  "*exp.Account",
					Value: val2,
				},
				{
					Tag:   "Destination",
					Type:  utils.MetaConstant,
					Path:  "*exp.Destination",
					Value: val3,
				},
			},
		},
	}
	for _, field := range value.rdrCfg.CacheDumpFields {
		field.ComputePath()
	}
	cfg := config.NewDefaultCGRConfig()
	data, err := engine.NewInternalDB(nil, nil, true, false, cfg.DataDbCfg().Items)
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
	expBody := "TestTenant,1001,1002\n"
	erS.onEvicted("FileID", value)

	path := filepath.Join(dirPath, "FileID.*.tmp")
	if match, err := filepath.Glob(path); err != nil {
		t.Error(err)
	} else if len(match) != 1 {
		t.Error("expected exactly one file")
	} else if body, err := os.ReadFile(match[0]); err != nil {
		t.Error(err)
	} else if expBody != string(body) {
		t.Errorf("expected: %q\nreceived: %q", expBody, string(body))
	}
}

func TestErsOnEvictedMetaDumpToFileCSVWriteErr(t *testing.T) {
	utils.Logger.SetLogLevel(3)
	utils.Logger.SetSyslog(nil)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	dirPath := "/tmp/TestErsOnEvictedMetaDumpToFile"
	err := os.Mkdir(dirPath, 0755)
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(dirPath)

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
				CSV: &config.CSVROpts{
					PartialCSVFieldSeparator: utils.StringPointer("\"")},
			},
		},
	}
	cfg := config.NewDefaultCGRConfig()
	data, err := engine.NewInternalDB(nil, nil, true, false, cfg.DataDbCfg().Items)
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

	rcvLog := buf.String()[20:]
	if !strings.Contains(rcvLog, "error: csv: invalid field or comment delimiter") {
		t.Errorf("expected: <%s> to be included in log message: <%s>",
			"error: csv: invalid field or comment delimiter", rcvLog)
	}
	utils.Logger.SetLogLevel(0)
}

func TestErsOnEvictedMetaDumpToFileCreateErr(t *testing.T) {
	utils.Logger.SetLogLevel(3)
	utils.Logger.SetSyslog(nil)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	dirPath := "/tmp/TestErsOnEvictedMetaDumpToFile"
	err := os.Mkdir(dirPath, 0755)
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(dirPath)

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
				PartialPath:        utils.StringPointer(dirPath + "/non-existent"),
			},
		},
	}
	cfg := config.NewDefaultCGRConfig()
	data, err := engine.NewInternalDB(nil, nil, true, false, cfg.DataDbCfg().Items)
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

	rcvLog := buf.String()[20:]
	if !strings.Contains(rcvLog, "CGRateS <> [ERROR] <ERs> Failed creating /tmp/TestErsOnEvictedMetaDumpToFile/non-existent/ID.") &&
		!strings.Contains(rcvLog, "error: open /tmp/TestErsOnEvictedMetaDumpToFile/non-existent/ID.") {
		t.Errorf("expected: <%s> and <%s> to be included in log message: <%s>",
			"CGRateS <> [ERROR] <ERs> Failed creating /tmp/TestErsOnEvictedMetaDumpToFile/non-existent/ID.",
			"error: open /tmp/TestErsOnEvictedMetaDumpToFile/non-existent/ID.",
			rcvLog)
	}

	utils.Logger.SetLogLevel(0)
}

func TestERsOnEvictedDumpToJSON(t *testing.T) {
	dirPath := "/tmp/TestErsOnEvictedDumpToJSON"
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		t.Error(err)
	}

	value := &erEvents{
		events: []*utils.CGREvent{
			{
				Tenant: "cgrates.org",
				ID:     "EventErsOnEvicted",
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Usage:        "10s",
					utils.Category:     "call",
					utils.Destination:  "1002",
					utils.OriginHost:   "local",
					utils.OriginID:     "123456",
					utils.ToR:          utils.MetaVoice,
					utils.CGRID:        "1133dc80896edf5049b46aa911cb9085eeb27f4c",
					utils.Password:     "secure_pass",
					"Additional_Field": "Additional_Value",
				},
			},
		},
		rdrCfg: &config.EventReaderCfg{ // CacheDumpFields will be empty
			ID:   "ER1",
			Type: utils.MetaNone,
			Opts: &config.EventReaderOpts{
				PartialCacheAction: utils.StringPointer(utils.MetaDumpToJSON),
				PartialPath:        utils.StringPointer(dirPath),
				PartialOrderField:  utils.StringPointer("2"),
			},
		},
	}

	cfg := config.NewDefaultCGRConfig()
	data, err := engine.NewInternalDB(nil, nil, true, false, cfg.DataDbCfg().Items)
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

	erS.onEvicted("ID_JSON", value)

	var files []string
	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})

	var compare map[string]any
	// compare = make(map[int][]string, 2)
	dataJSON, err := os.ReadFile(files[0])
	if err != nil {
		t.Error(err)
	}
	err = json.Unmarshal(dataJSON, &compare)
	if err != nil {
		t.Error(err)
	}

	exp := map[string]any{
		utils.AccountField: "1001",
		utils.Usage:        "10s",
		utils.Category:     "call",
		utils.Destination:  "1002",
		utils.OriginHost:   "local",
		utils.OriginID:     "123456",
		utils.ToR:          utils.MetaVoice,
		utils.CGRID:        "1133dc80896edf5049b46aa911cb9085eeb27f4c",
		utils.Password:     "secure_pass",
		"Additional_Field": "Additional_Value",
	}
	// fmt.Println(utils.ToJSON(compare))
	if !reflect.DeepEqual(exp, compare) {
		t.Errorf("Expected %v \n but received \n %v", exp, compare)
	}
	if err := os.RemoveAll(dirPath); err != nil {
		t.Error(err)
	}
}

func TestErsOnEvictedDumpToJSONNoPath(t *testing.T) {
	dirPath := ""
	value := &erEvents{
		events: []*utils.CGREvent{
			{
				Tenant: "cgrates.org",
				ID:     "EventErsOnEvicted",
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Usage:        "10s",
					utils.Category:     "call",
					utils.Destination:  "1002",
					utils.OriginHost:   "local",
					utils.OriginID:     "123456",
					utils.ToR:          utils.MetaVoice,
					utils.CGRID:        "1133dc80896edf5049b46aa911cb9085eeb27f4c",
					utils.Password:     "secure_pass",
					"Additional_Field": "Additional_Value",
				},
			},
		},
		rdrCfg: &config.EventReaderCfg{ // CacheDumpFields will be empty
			ID:   "ER1",
			Type: utils.MetaNone,
			Opts: &config.EventReaderOpts{
				PartialCacheAction: utils.StringPointer(utils.MetaDumpToJSON),
				PartialPath:        utils.StringPointer(dirPath),
				PartialOrderField:  utils.StringPointer("2"),
			},
		},
	}

	cfg := config.NewDefaultCGRConfig()
	data, err := engine.NewInternalDB(nil, nil, true, false, cfg.DataDbCfg().Items)
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

	//Should return nothing since there is no path therefore no writing implied.
	erS.onEvicted("ID_JSON", value)

}

func TestErsOnEvictedDumpToJSONMergeError(t *testing.T) {
	utils.Logger.SetLogLevel(4)
	utils.Logger.SetSyslog(nil)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	dirPath := "/tmp/TestErsOnEvictedCacheDumpfields"
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		t.Error(err)
	}

	value := &erEvents{
		events: []*utils.CGREvent{
			{
				Tenant: "cgrates.org",
				ID:     "EventErsOnEvicted",
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Usage:        "10s",
					utils.Category:     "call",
					utils.Destination:  "1002",
					utils.OriginHost:   "local",
					utils.OriginID:     "123456",
					utils.ToR:          utils.MetaVoice,
					utils.CGRID:        "1133dc80896edf5049b46aa911cb9085eeb27f4c",
					utils.Password:     "secure_pass",
					"Additional_Field": "Additional_Value",
					utils.AnswerTime:   time.Date(2021, 6, 1, 12, 0, 0, 0, time.UTC),
				},
			},
			{
				Tenant: "cgrates.org",
				ID:     "EventErsOnEvicted2",
				Event: map[string]any{
					utils.AccountField: "1002",
					utils.Usage:        "12s",
					utils.Category:     "call",
					utils.Destination:  "1003",
					utils.OriginID:     "1234567",
					utils.ToR:          utils.MetaSMS,
					utils.CGRID:        "1133dc80896edf5049b46aa911cb9085eeb27f4d",
					utils.Password:     "secure_password",
					"Additional_Field": "Additional_Value2",
					utils.AnswerTime:   time.Date(2021, 6, 1, 13, 0, 0, 0, time.UTC),
				},
			},
		},
		rdrCfg: &config.EventReaderCfg{ // CacheDumpFields will be empty
			ID:   "ER1",
			Type: utils.MetaNone,
			Opts: &config.EventReaderOpts{
				PartialCacheAction: utils.StringPointer(utils.MetaDumpToJSON),
				PartialPath:        utils.StringPointer(dirPath),
			},
		},
	}

	cfg := config.NewDefaultCGRConfig()
	data, err := engine.NewInternalDB(nil, nil, true, false, cfg.DataDbCfg().Items)
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
	expLog := `<unsupported comparison type: string, kind: string>`
	erS.onEvicted("ID", value)
	rcvLog := buf.String()[20:]
	if !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected <%+v> to be included in: <%+v>", expLog, rcvLog)
	}

	utils.Logger.SetLogLevel(0)
	if err := os.RemoveAll(dirPath); err != nil {
		t.Error(err)
	}
}

func TestERsOnEvictedDumpToJSONWithCacheDumpFieldsErrPrefix(t *testing.T) {
	dirPath := "/tmp/TestErsOnEvictedDumpToJSON"
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		t.Error(err)
	}

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
					utils.AccountField: "1001",
					utils.Usage:        "10s",
					utils.Category:     "call",
					utils.Destination:  "1002",
					utils.OriginHost:   "local",
					utils.OriginID:     "123456",
					utils.ToR:          utils.MetaVoice,
					utils.CGRID:        "1133dc80896edf5049b46aa911cb9085eeb27f4c",
					utils.Password:     "secure_pass",
					"Additional_Field": "Additional_Value",
				},
			},
		},
		rdrCfg: &config.EventReaderCfg{ // CacheDumpFields will be empty
			ID:   "ER1",
			Type: utils.MetaNone,
			Opts: &config.EventReaderOpts{
				PartialCacheAction: utils.StringPointer(utils.MetaDumpToJSON),
				PartialPath:        utils.StringPointer(dirPath),
				PartialOrderField:  utils.StringPointer("2"),
			},
			CacheDumpFields: []*config.FCTemplate{
				{
					Tag:         "*tor",
					Type:        utils.MetaComposed,
					Path:        "~*req.ToR",
					Value:       config.NewRSRParsersMustCompile(utils.MetaVoice, utils.InfieldSep),
					NewBranch:   false,
					AttributeID: "ATTR_FLD_1001",
				},
			},
		},
	}

	value.rdrCfg.CacheDumpFields[0].ComputePath()

	cfg := config.NewDefaultCGRConfig()
	data, err := engine.NewInternalDB(nil, nil, true, false, cfg.DataDbCfg().Items)
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

	expLog := `Converting CDR with CGRID: <ID_JSON> to record , ignoring due to error: <unsupported field prefix: <~*req> when set field>`
	erS.onEvicted("ID_JSON", value)
	rcvLog := buf.String()[20:]
	if !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected <%+v> to be included in: <%+v>", expLog, rcvLog)
	}

	utils.Logger.SetLogLevel(0)
	if err := os.RemoveAll(dirPath); err != nil {
		t.Error(err)
	}
}

func TestERsOnEvictedDumpToJSONWithCacheDumpFields(t *testing.T) {
	dirPath := "/tmp/TestErsOnEvictedDumpToJSON"
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		t.Error(err)
	}

	value := &erEvents{
		events: []*utils.CGREvent{
			{
				Tenant: "cgrates.org",
				ID:     "EventErsOnEvicted",
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Usage:        "10s",
					utils.Category:     "call",
					utils.Destination:  "1002",
					utils.OriginHost:   "local",
					utils.OriginID:     "123456",
					utils.ToR:          utils.MetaVoice,
					utils.CGRID:        "1133dc80896edf5049b46aa911cb9085eeb27f4c",
					utils.Password:     "secure_pass",
					"Additional_Field": "Additional_Value",
				},
			},
		},
		rdrCfg: &config.EventReaderCfg{ // CacheDumpFields will be empty
			ID:   "ER1",
			Type: utils.MetaNone,
			Opts: &config.EventReaderOpts{
				PartialCacheAction: utils.StringPointer(utils.MetaDumpToJSON),
				PartialPath:        utils.StringPointer(dirPath),
				PartialOrderField:  utils.StringPointer("2"),
			},
			Fields: []*config.FCTemplate{
				{Tag: "SessionId", Path: utils.EmptyString, Type: "*variable",
					Value: config.NewRSRParsersMustCompile("~*req.Session-Id", utils.InfieldSep), Mandatory: true},
			},
			CacheDumpFields: []*config.FCTemplate{
				{
					Tag:       "OriginID",
					Type:      utils.MetaConstant,
					Path:      "*exp.OriginID",
					Value:     config.NewRSRParsersMustCompile("25160047719:0", utils.InfieldSep),
					Mandatory: true,
				},
			},
		},
	}

	value.rdrCfg.CacheDumpFields[0].ComputePath()

	cfg := config.NewDefaultCGRConfig()
	data, err := engine.NewInternalDB(nil, nil, true, false, cfg.DataDbCfg().Items)
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

	erS.onEvicted("ID_JSON", value)

	var files []string
	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})

	var compare map[string]any
	dataJSON, err := os.ReadFile(files[0])
	if err != nil {
		t.Error(err)
	}
	err = json.Unmarshal(dataJSON, &compare)
	if err != nil {
		t.Error(err)
	}
	exp := map[string]any{
		utils.OriginID: "25160047719:0",
	}
	if !reflect.DeepEqual(exp, compare) {
		t.Errorf("Expected %v \n but received \n %v", exp, compare)
	}
	if err := os.RemoveAll(dirPath); err != nil {
		t.Error(err)
	}
}

func TestErsOnEvictedDumpToJSONInvalidPath(t *testing.T) {
	utils.Logger.SetLogLevel(4)
	utils.Logger.SetSyslog(nil)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	// dirPath := "/tmp/TestErsOnEvictedDumpToJSON"
	// err := os.MkdirAll(dirPath, 0755)
	// if err != nil {
	// 	t.Error(err)
	// }
	value := &erEvents{
		events: []*utils.CGREvent{
			{
				Tenant: "cgrates.org",
				ID:     "EventErsOnEvicted",
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Usage:        "10s",
					utils.Category:     "call",
					utils.Destination:  "1002",
					utils.OriginHost:   "local",
					utils.OriginID:     "123456",
					utils.ToR:          utils.MetaVoice,
					utils.CGRID:        "1133dc80896edf5049b46aa911cb9085eeb27f4c",
					utils.Password:     "secure_pass",
					"Additional_Field": "Additional_Value",
				},
			},
		},
		rdrCfg: &config.EventReaderCfg{ // CacheDumpFields will be empty
			ID:   "ER1",
			Type: utils.MetaNone,
			Opts: &config.EventReaderOpts{
				PartialCacheAction: utils.StringPointer(utils.MetaDumpToJSON),
				PartialPath:        utils.StringPointer("invalid path"),
				PartialOrderField:  utils.StringPointer("2"),
			},
		},
	}

	cfg := config.NewDefaultCGRConfig()
	data, err := engine.NewInternalDB(nil, nil, true, false, cfg.DataDbCfg().Items)
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

	expLog := ".tmp: no such file or directory"
	erS.onEvicted("ID_JSON", value)
	rcvLog := buf.String()[20:]
	if !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected <%+v> to be included in: <%+v>", expLog, rcvLog)
	}
	utils.Logger.SetLogLevel(0)
	// if err := os.RemoveAll(dirPath); err != nil {
	// 	t.Error(err)
	// }
}

func TestErsOnEvictedDumpToJSONEncodeErr(t *testing.T) {
	utils.Logger.SetLogLevel(4)
	utils.Logger.SetSyslog(nil)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	dirPath := "/tmp/TestErsOnEvictedDumpToJSON"
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		t.Error(err)
	}
	value := &erEvents{
		events: []*utils.CGREvent{
			{
				Tenant: "cgrates.org",
				ID:     "EventErsOnEvicted",
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Usage:        "10s",
					utils.Category:     "call",
					utils.Destination:  "1002",
					utils.OriginHost:   "local",
					utils.OriginID:     "123456",
					utils.ToR:          utils.MetaVoice,
					utils.CGRID:        "1133dc80896edf5049b46aa911cb9085eeb27f4c",
					utils.Password:     "secure_pass",
					"Additional_Field": "Additional_Value",
					"EncodeErr": func() {

					},
				},
			},
		},
		rdrCfg: &config.EventReaderCfg{ // CacheDumpFields will be empty
			ID:   "ER1",
			Type: utils.MetaNone,
			Opts: &config.EventReaderOpts{
				PartialCacheAction: utils.StringPointer(utils.MetaDumpToJSON),
				PartialPath:        utils.StringPointer(dirPath),
				PartialOrderField:  utils.StringPointer("2"),
			},
		},
	}

	cfg := config.NewDefaultCGRConfig()
	data, err := engine.NewInternalDB(nil, nil, true, false, cfg.DataDbCfg().Items)
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

	expLog := "error: json: unsupported type: func()"
	erS.onEvicted("ID_JSON", value)
	rcvLog := buf.String()[20:]
	if !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected <%+v> to be included in: <%+v>", expLog, rcvLog)
	}
	utils.Logger.SetLogLevel(0)
	if err := os.RemoveAll(dirPath); err != nil {
		t.Error(err)
	}
}
