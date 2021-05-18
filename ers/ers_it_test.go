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
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc"
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
		{
			ID:   "",
			Type: "",
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
			ID:   "",
			Type: "",
		},
	}
	fltrS := &engine.FilterS{}
	srv := NewERService(cfg, fltrS, nil)
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
	srv := NewERService(cfg, fltrS, nil)
	srv.stopLsn[""] = make(chan struct{}, 1)
	srv.closeAllRdrs()
}
func TestERsListenAndServeRdrErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers = []*config.EventReaderCfg{
		{
			ID:   "",
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
			ID:   "",
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
			ID:   "",
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
		cgrEvent: &utils.CGREvent{
			Tenant:  "",
			ID:      "",
			Event:   nil,
			APIOpts: nil,
		},
		rdrCfg: &config.EventReaderCfg{
			ID: "",
		},
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
			ID:   "",
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
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     nil,
		rdrDir:    "",
		rdrEvents: nil,
		rdrError:  nil,
		rdrExit:   nil,
		conReqs:   nil,
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
	cgrEvent := &utils.CGREvent{
		Tenant:  "",
		ID:      "",
		Event:   nil,
		APIOpts: nil,
	}
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
	cgrEvent := &utils.CGREvent{
		Tenant:  "",
		ID:      "",
		Event:   nil,
		APIOpts: nil,
	}
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
		APIOpts: map[string]interface{}{
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
		APIOpts: map[string]interface{}{
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
		Tenant: "",
		ID:     "",
		Event:  nil,
		APIOpts: map[string]interface{}{
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
		APIOpts: map[string]interface{}{
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
		APIOpts: map[string]interface{}{
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
		Tenant: "",
		ID:     "",
		Event:  nil,
		APIOpts: map[string]interface{}{
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
		Tenant: "",
		ID:     "",
		Event:  nil,
		APIOpts: map[string]interface{}{
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
		Tenant: "",
		ID:     "",
		Event: map[string]interface{}{
			utils.Usage: time.Second,
		},
		APIOpts: map[string]interface{}{
			utils.OptsRoutesLimit: true,
		},
	}
	err := srv.processEvent(cgrEvent, rdrCfg)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [connIDs]" {
		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", "MANDATORY_IE_MISSING: [connIDs]", err)
	}
}

type testMockClients struct {
	calls map[string]func(args interface{}, reply interface{}) error
}

func (sT *testMockClients) Call(method string, arg interface{}, rply interface{}) error {
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
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.SessionSv1ProcessMessage: func(args interface{}, reply interface{}) error {
				return errors.New("RALS_ERROR")
			},
		},
	}
	clientChan := make(chan birpc.ClientConnector, 1)
	clientChan <- testMockClient
	connMng := engine.NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS): clientChan,
	})
	srv := NewERService(cfg, fltrS, connMng)
	rdrCfg := &config.EventReaderCfg{
		Flags: map[string]utils.FlagParams{
			utils.MetaMessage: map[string][]string{},
		},
	}
	cgrEvent := &utils.CGREvent{
		Tenant: "",
		ID:     "",
		Event: map[string]interface{}{
			utils.Usage: 0,
		},
		APIOpts: map[string]interface{}{
			utils.OptsRoutesLimit: true,
		},
	}
	err := srv.processEvent(cgrEvent, rdrCfg)
	if err == nil || err.Error() != "RALS_ERROR" {
		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", "RALS_ERROR", err)
	}
}
