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
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
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
		{},
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
		{},
	}
	fltrS := &engine.FilterS{}
	srv := NewERService(cfg, fltrS, nil)
	rdrCfg := &config.EventReaderCfg{}
	cgrEvent := &utils.CGREvent{}
	err := srv.processEvent(cgrEvent, rdrCfg)
	if err == nil || err.Error() != "unsupported reqType: <>" {
		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", "unsupported reqType: <>", err)
	}
}

func TestERsCloseAllRdrs(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers = []*config.EventReaderCfg{
		{},
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
			Type: utils.MetaNone,
		},
	}
	fltrS := &engine.FilterS{}
	srv := NewERService(cfg, fltrS, nil)
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
	srv := NewERService(cfg, fltrS, nil)
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
	srv := NewERService(cfg, fltrS, nil)
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
	srv := NewERService(cfg, fltrS, nil)
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
	srv := NewERService(cfg, fltrS, nil)
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
	srv := NewERService(cfg, fltrS, nil)
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
	srv := NewERService(cfg, fltrS, nil)
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
	srv := NewERService(cfg, fltrS, nil)
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
	srv := NewERService(cfg, fltrS, nil)
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
	srv := NewERService(cfg, fltrS, nil)
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
	srv := NewERService(cfg, fltrS, nil)
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
	srv := NewERService(cfg, fltrS, nil)
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
	srv := NewERService(cfg, fltrS, nil)
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
	srv := NewERService(cfg, fltrS, nil)
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
	srv := NewERService(cfg, fltrS, nil)
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
	srv := NewERService(cfg, fltrS, nil)
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
	srv := NewERService(cfg, fltrS, nil)
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
	srv := NewERService(cfg, fltrS, nil)
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
	srv := NewERService(cfg, fltrS, nil)
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
	calls map[string]func(ctx *context.Context, args, reply any) error
}

func (sT *testMockClients) Call(ctx *context.Context, method string, arg any, rply any) error {
	if call, has := sT.calls[method]; !has {
		return rpcclient.ErrUnsupporteServiceMethod
	} else {
		return call(ctx, arg, rply)
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
		calls: map[string]func(ctx *context.Context, args, reply any) error{
			utils.SessionSv1ProcessMessage: func(ctx *context.Context, args, reply any) error {
				return errors.New("RALS_ERROR")
			},
		},
	}
	clientChan := make(chan birpc.ClientConnector, 1)
	clientChan <- testMockClient
	connMng := engine.NewConnManager(cfg)
	connMng.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS), utils.SessionSv1, clientChan)
	srv := NewERService(cfg, fltrS, connMng)
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
		},
	}
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrS := engine.NewFilterS(cfg, nil, dm)
	erS := &ERService{
		cfg:       cfg,
		rdrEvents: make(chan *erEvent, 1),
		filterS:   fltrS,
	}
	erS.onEvicted("ID", value)

	// rcv, err := os.ReadFile(filepath.Join(dirPath, "ID.*.*"))
	// if err != nil {
	// 	t.Error(err)
	// }
	// fmt.Println(rcv)
}

func TestErsOnEvictedMetaDumpToFileCSVWriteErr(t *testing.T) {
	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
	}()
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 3)

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
				PartialCacheAction:       utils.StringPointer(utils.MetaDumpToFile),
				PartialPath:              utils.StringPointer(dirPath),
				PartialCSVFieldSeparator: utils.StringPointer("\""),
			},
		},
	}
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrS := engine.NewFilterS(cfg, nil, dm)
	erS := &ERService{
		cfg:       cfg,
		rdrEvents: make(chan *erEvent, 1),
		filterS:   fltrS,
	}

	erS.onEvicted("ID", value)

	rcvLog := buf.String()
	if !strings.Contains(rcvLog, "error: csv: invalid field or comment delimiter") {
		t.Errorf("expected: <%s> to be included in log message: <%s>",
			"error: csv: invalid field or comment delimiter", rcvLog)
	}
}

func TestErsOnEvictedMetaDumpToFileCreateErr(t *testing.T) {
	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
	}()
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 3)

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
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrS := engine.NewFilterS(cfg, nil, dm)
	erS := &ERService{
		cfg:       cfg,
		rdrEvents: make(chan *erEvent, 1),
		filterS:   fltrS,
	}

	erS.onEvicted("ID", value)

	rcvLog := buf.String()
	if !strings.Contains(rcvLog, "CGRateS <> [ERROR] <ERs> Failed creating /tmp/TestErsOnEvictedMetaDumpToFile/non-existent/ID.") &&
		!strings.Contains(rcvLog, "error: open /tmp/TestErsOnEvictedMetaDumpToFile/non-existent/ID.") {
		t.Errorf("expected: <%s> and <%s> to be included in log message: <%s>",
			"CGRateS <> [ERROR] <ERs> Failed creating /tmp/TestErsOnEvictedMetaDumpToFile/non-existent/ID.",
			"error: open /tmp/TestErsOnEvictedMetaDumpToFile/non-existent/ID.",
			rcvLog)
	}
}

func TestErsOnEvictedNoCacheDumpFields(t *testing.T) {
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
					utils.Password:     "secure_pass",
					"Additional_Field": "Additional_Value",
				},
				APIOpts: map[string]any{
					utils.MetaOriginID: "1133dc80896edf5049b46aa911cb9085eeb27f4c",
				},
			},
		},
		rdrCfg: &config.EventReaderCfg{ // CacheDumpFields will be empty
			ID:   "ER1",
			Type: utils.MetaNone,
			Opts: &config.EventReaderOpts{
				PartialCacheAction: utils.StringPointer(utils.MetaDumpToFile),
				PartialPath:        utils.StringPointer(dirPath),
				PartialOrderField:  utils.StringPointer("2"),
			},
		},
	}

	value2 := &erEvents{
		events: []*utils.CGREvent{
			{
				Tenant: "cgrates.org",
				ID:     "EventErsOnEvicted",
				Event: map[string]any{
					utils.AccountField: "1002",
					utils.Usage:        "12s",
					utils.Category:     "call",
					utils.Destination:  "1003",
					utils.OriginID:     "1234567",
					utils.ToR:          utils.MetaSMS,
					utils.Password:     "secure_password",
					"Additional_Field": "Additional_Value2",
				},
				APIOpts: map[string]any{
					utils.MetaOriginID: "1133dc80896edf5049b46aa911cb9085eeb27f4d",
				},
			},
		},
		rdrCfg: &config.EventReaderCfg{ // CacheDumpFields will be empty
			ID:   "ER1",
			Type: utils.MetaNone,
			Opts: &config.EventReaderOpts{
				PartialCacheAction: utils.StringPointer(utils.MetaDumpToFile),
				PartialPath:        utils.StringPointer(dirPath),
				PartialOrderField:  utils.StringPointer("2"),
			},
		},
	}

	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrS := engine.NewFilterS(cfg, nil, dm)
	erS := &ERService{
		cfg:       cfg,
		rdrEvents: make(chan *erEvent, 1),
		filterS:   fltrS,
	}

	erS.onEvicted("ID", value)
	erS.onEvicted("ID2", value2)

	var files []string
	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Error(err)
	}
	compare := make(map[int][]string, 2)
	for idx, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Error(err)
		}
		s := strings.Split(string(data), ",")

		compare[idx] = s
	}
	if len(compare[0]) != 9 && len(compare[1]) != 8 {
		t.Errorf("expected <%d> and <%d>, \nreceived: <%d> and <%d>",
			10, 9, len(compare[0]), len(compare[1]))
	}
	if err := os.RemoveAll(dirPath); err != nil {
		t.Error(err)
	}
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

					utils.Password:     "secure_pass",
					"Additional_Field": "Additional_Value",
				},
				APIOpts: map[string]any{
					utils.MetaOriginID: "1133dc80896edf5049b46aa911cb9085eeb27f4c",
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
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
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
	if err != nil {
		t.Error(err)
	}

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
		utils.Password:     "secure_pass",
		"Additional_Field": "Additional_Value",
	}
	if !reflect.DeepEqual(exp, compare) {
		t.Errorf("expected <%v>,\nreceived <%v>", utils.ToJSON(exp), utils.ToJSON(compare))
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

					utils.Password:     "secure_pass",
					"Additional_Field": "Additional_Value",
				},
				APIOpts: map[string]any{
					utils.MetaOriginID: "1133dc80896edf5049b46aa911cb9085eeb27f4c",
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
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
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
	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
	}()
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 4)

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

					utils.Password:     "secure_pass",
					"Additional_Field": "Additional_Value",
					utils.AnswerTime:   time.Date(2021, 6, 1, 12, 0, 0, 0, time.UTC),
				},
				APIOpts: map[string]any{
					utils.MetaOriginID: "1133dc80896edf5049b46aa911cb9085eeb27f4c",
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
					utils.Password:     "secure_password",
					"Additional_Field": "Additional_Value2",
					utils.AnswerTime:   time.Date(2021, 6, 1, 13, 0, 0, 0, time.UTC),
				},
				APIOpts: map[string]any{
					utils.MetaOriginID: "1133dc80896edf5049b46aa911cb9085eeb27f4d",
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
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrS := engine.NewFilterS(cfg, nil, dm)
	erS := &ERService{
		cfg:       cfg,
		rdrEvents: make(chan *erEvent, 1),
		filterS:   fltrS,
	}
	expLog := `[WARNING] <ERs> failed posting expired parial events <[{"Tenant":"cgrates.org","ID":"EventErsOnEvicted2","Event":{"Account":"1002","Additional_Field":"Additional_Value2","AnswerTime":"2021-06-01T13:00:00Z","Category":"call","Destination":"1003","OriginID":"1234567","ToR":"*sms","Usage":"12s","password":"secure_password"},"APIOpts":{"*originID":"1133dc80896edf5049b46aa911cb9085eeb27f4d"}},{"Tenant":"cgrates.org","ID":"EventErsOnEvicted","Event":{"Account":"1001","Additional_Field":"Additional_Value","AnswerTime":"2021-06-01T12:00:00Z","Category":"call","Destination":"1002","OriginHost":"local","OriginID":"123456","ToR":"*voice","Usage":"10s","password":"secure_pass"},"APIOpts":{"*originID":"1133dc80896edf5049b46aa911cb9085eeb27f4c"}}]>`
	erS.onEvicted("ID", value)
	rcvLog := buf.String()
	if !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected <%+v> to be included in: <%+v>", expLog, rcvLog)
	}

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

	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
	}()
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 4)

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

					utils.Password:     "secure_pass",
					"Additional_Field": "Additional_Value",
				},
				APIOpts: map[string]any{
					utils.MetaOriginID: "1133dc80896edf5049b46aa911cb9085eeb27f4c",
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
					Value:       utils.NewRSRParsersMustCompile(utils.MetaVoice, utils.InfieldSep),
					NewBranch:   false,
					AttributeID: "ATTR_FLD_1001",
				},
			},
		},
	}

	value.rdrCfg.CacheDumpFields[0].ComputePath()

	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrS := engine.NewFilterS(cfg, nil, dm)
	erS := &ERService{
		cfg:       cfg,
		rdrEvents: make(chan *erEvent, 1),
		filterS:   fltrS,
	}

	expLog := `Converting CDR with originID: <ID_JSON> to record , ignoring due to error: <unsupported field prefix: <~*req> when set field>`
	erS.onEvicted("ID_JSON", value)
	rcvLog := buf.String()
	if !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected <%+v> to be included in: <%+v>", expLog, rcvLog)
	}

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

					utils.Password:     "secure_pass",
					"Additional_Field": "Additional_Value",
				},
				APIOpts: map[string]any{
					utils.MetaOriginID: "1133dc80896edf5049b46aa911cb9085eeb27f4c",
				},
			},
		},
		rdrCfg: &config.EventReaderCfg{
			ID:   "ER1",
			Type: utils.MetaNone,
			Opts: &config.EventReaderOpts{
				PartialCacheAction: utils.StringPointer(utils.MetaDumpToJSON),
				PartialPath:        utils.StringPointer(dirPath),
				PartialOrderField:  utils.StringPointer("2"),
			},
			Fields: []*config.FCTemplate{
				{Tag: "SessionId", Path: utils.EmptyString, Type: "*variable",
					Value: utils.NewRSRParsersMustCompile("~*req.Session-Id", utils.InfieldSep), Mandatory: true},
			},
			CacheDumpFields: []*config.FCTemplate{
				{
					Tag:       "OriginID",
					Type:      utils.MetaConstant,
					Path:      "*exp.OriginID",
					Value:     utils.NewRSRParsersMustCompile("25160047719:0", utils.InfieldSep),
					Mandatory: true,
				},
			},
		},
	}

	value.rdrCfg.CacheDumpFields[0].ComputePath()

	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
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
	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
	}()
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 4)

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

					utils.Password:     "secure_pass",
					"Additional_Field": "Additional_Value",
				},
				APIOpts: map[string]any{
					utils.MetaOriginID: "1133dc80896edf5049b46aa911cb9085eeb27f4c",
				},
			},
		},
		rdrCfg: &config.EventReaderCfg{ // CacheDumpFields will be empty
			ID:   "ER1",
			Type: utils.MetaNone,
			Opts: &config.EventReaderOpts{
				PartialCacheAction: utils.StringPointer(utils.MetaDumpToJSON),
				PartialPath:        utils.StringPointer("invalid_path"),
				PartialOrderField:  utils.StringPointer("2"),
			},
		},
	}

	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
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
	if err := os.RemoveAll(dirPath); err != nil {
		t.Error(err)
	}
}

func TestErsOnEvictedDumpToJSONEncodeErr(t *testing.T) {
	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
	}()
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 4)

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

					utils.Password:     "secure_pass",
					"Additional_Field": "Additional_Value",
					"EncodeErr": func() {

					},
				},
				APIOpts: map[string]any{
					utils.MetaOriginID: "1133dc80896edf5049b46aa911cb9085eeb27f4c",
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
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrS := engine.NewFilterS(cfg, nil, dm)
	erS := &ERService{
		cfg:       cfg,
		rdrEvents: make(chan *erEvent, 1),
		filterS:   fltrS,
	}

	expLog := "error: json: unsupported type: func()"
	erS.onEvicted("ID_JSON", value)
	rcvLog := buf.String()
	if !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected <%+v> to be included in: <%+v>", expLog, rcvLog)
	}
	if err := os.RemoveAll(dirPath); err != nil {
		t.Error(err)
	}
}

func TestERsLineNr(t *testing.T) {
	var dbCfg engine.DBCfg
	switch *utils.DBType {
	case utils.MetaInternal:
		dbCfg = engine.InternalDBCfg
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}
	csvFd, fwvFd, xmlFd, procFd := t.TempDir(), t.TempDir(), t.TempDir(), t.TempDir()
	content := fmt.Sprintf(`{
"ers": {
	"enabled": true,
	"sessions_conns": [],
	"readers": [
		{
			"id": "file_csv_reader",
			"run_delay":  "-1",
			"type": "*fileCSV",
			"source_path": "%s",
			"flags": ["*dryRun"],
			"processed_path": "%s",
			"fields":[
				{"tag": "FileName", "path": "*cgreq.FileName", "type": "*variable", "value": "~*vars.*fileName"},
				{"tag": "LineNumber", "path": "*cgreq.LineNumber", "type": "*variable", "value": "~*vars.*fileLineNumber"},
				{"tag": "Field", "path": "*cgreq.Field", "type": "*variable", "value": "~*req.0"}
			]
		},
		{
			"id": "file_fwv_reader",
			"run_delay":  "-1",
			"type": "*fileFWV",
			"source_path": "%s",
			"flags": ["*dryRun"],
			"processed_path": "%s",
			"fields":[
				{"tag": "FileName2", "path": "*cgreq.FileName", "type": "*variable", "value": "~*vars.*fileName"},
				{"tag": "LineNumber", "path": "*cgreq.LineNumber", "type": "*variable", "value": "~*vars.*fileLineNumber"},
				{"tag": "FileSeqNr", "path": "*cgreq.FileSeqNr", "type": "*variable", "value": "~*hdr.3-6", "padding":"*zeroleft"},
				{"tag": "Field", "path": "*cgreq.Field", "type": "*variable", "value": "~*req.0-5", "padding":"*right"},
				{"tag": "NrOfElements", "type": "*variable", "path":"*cgreq.NrOfElements", "value": "~*trl.3-4"},
			]
		},
		{
			"id": "file_xml_reader",
			"run_delay":  "-1",
			"type": "*fileXML",
			"source_path": "%s",
			"flags": ["*dryRun"],
			"processed_path": "%s",
			"opts": {
				"xmlRootPath": "root.field"
			},
			"fields":[
				{"tag": "FileName", "path": "*cgreq.FileName", "type": "*variable", "value": "~*vars.*fileName"},
				{"tag": "LineNumber", "path": "*cgreq.LineNumber", "type": "*variable", "value": "~*vars.*fileLineNumber"},
				{"tag": "Field", "path": "*cgreq.Field", "type": "*variable", "value": "~*req.root.field"}
			]
		}
	]
}
}`, csvFd, procFd, fwvFd, procFd, xmlFd, procFd)
	buf := &bytes.Buffer{}
	ng := engine.TestEngine{
		ConfigJSON: content,
		DBCfg:      dbCfg,
		Encoding:   *utils.Encoding,
		LogBuffer:  buf,
	}
	_, _ = ng.Run(t)
	fileIdx := 0
	createFile := func(t *testing.T, dir, ext, content string) {
		fileIdx++
		filePath := filepath.Join(dir, fmt.Sprintf("file%d%s", fileIdx, ext))
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("could not write to file %s: %v", filePath, err)
		}
	}
	verifyLogLines := func(t *testing.T, reader io.Reader) {
		fieldRegex := regexp.MustCompile(`"Field":"test(\d+)"`)
		lineNumberRegex := regexp.MustCompile(`"LineNumber":"(\d+)"`)
		records := 0
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.Contains(line, `"Field":"test`) {
				continue
			}
			records++
			fieldMatch := fieldRegex.FindStringSubmatch(line)
			lineNumberMatch := lineNumberRegex.FindStringSubmatch(line)
			if len(fieldMatch) != 2 || len(lineNumberMatch) != 2 {
				t.Fatalf("invalid log line format: %s", line)
			}
			testNumber, err := strconv.Atoi(fieldMatch[1])
			if err != nil {
				t.Fatal(err)
			}
			lineNumber, err := strconv.Atoi(lineNumberMatch[1])
			if err != nil {
				t.Fatal(err)
			}
			if testNumber != lineNumber {
				t.Errorf("mismatch in line: %s, field number: %d, line number: %d", line, testNumber, lineNumber)
			}
		}
		if err := scanner.Err(); err != nil {
			t.Errorf("error reading input: %v", err)
		}
		if records != 18 {
			t.Errorf("expected ERs to process 18 records, but it processed %d records", records)
		}
	}
	// Create the files inside the source directories of the readers.
	createFile(t, csvFd, utils.CSVSuffix, "test1\ntest2\ntest3\ntest4\ntest5\ntest6")
	createFile(t, fwvFd, utils.FWVSuffix, `HDR002
test1
test2
test3
test4
test5
test6
TRL6
`)
	createFile(t, xmlFd, utils.XMLSuffix, `<?xml version="1.0" encoding="ISO-8859-1"?>
<root>
  <field>test1</field>
  <field>test2</field>
  <field>test3</field>
  <field>test4</field>
  <field>test5</field>
  <field>test6</field>
</root>`)
	time.Sleep(5 * time.Millisecond) // wait for the files to be processed
	// Check that the suffixes of the 'test' fields match the LineNumber field.
	logData := strings.NewReader(buf.String())
	verifyLogLines(t, logData)
}
