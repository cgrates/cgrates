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

package sessions

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func TestSetSTerminator(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().SessionTTL = time.Second
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	sessions := NewSessionS(cfg, dm, nil)

	ss := new(Session)

	opts := engine.MapEvent{
		utils.OptsDebitInterval: "0s",
	}

	terminator := &sTerminator{
		ttl: cfg.SessionSCfg().SessionTTL,
	}
	sessions.setSTerminator(ss, opts)
	ss.sTerminator.timer = nil
	ss.sTerminator.endChan = nil
	if !reflect.DeepEqual(terminator, ss.sTerminator) {
		t.Errorf("Expected %+v, received %+v", terminator, ss.sTerminator)
	}

	opts = engine.MapEvent{
		utils.OptsSessionsTTL:          "1s",
		utils.OptsSessionsTTLMaxDelay:  "1s",
		utils.OptsSessionsTTLLastUsed:  "2s",
		utils.OptsSessionsTTLLastUsage: "0s",
		utils.OptsSessionsTTLUsage:     "5s",
	}
	ss.sTerminator = &sTerminator{
		timer: time.NewTimer(cfg.SessionSCfg().SessionTTL),
	}

	terminator = &sTerminator{
		ttl:          0,
		ttlLastUsed:  utils.DurationPointer(2 * time.Second),
		ttlUsage:     utils.DurationPointer(5 * time.Second),
		ttlLastUsage: utils.DurationPointer(0),
	}
	sessions.setSTerminator(ss, opts)
	ss.sTerminator.timer = nil
	ss.sTerminator.endChan = nil
	ss.sTerminator.ttl = 0
	if !reflect.DeepEqual(terminator, ss.sTerminator) {
		t.Errorf("Expected %+v, received %+v", terminator, ss.sTerminator)
	}

	ss.OptsStart = engine.MapEvent{
		utils.OptsSessionsTTL:          "1s",
		utils.OptsSessionsTTLMaxDelay:  "1s",
		utils.OptsSessionsTTLLastUsed:  "2s",
		utils.OptsSessionsTTLLastUsage: "0s",
		utils.OptsSessionsTTLUsage:     "5s",
	}
	ss.sTerminator = &sTerminator{
		timer: time.NewTimer(cfg.SessionSCfg().SessionTTL),
	}
	opts = engine.MapEvent{}
	sessions.setSTerminator(ss, opts)
	ss.sTerminator.timer = nil
	ss.sTerminator.endChan = nil
	ss.sTerminator.ttl = 0
	if !reflect.DeepEqual(terminator, ss.sTerminator) {
		t.Errorf("Expected %+v, received %+v", terminator, ss.sTerminator)
	}
}

func TestSetSTerminatorError(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().SessionTTL = time.Second
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	sessions := NewSessionS(cfg, dm, nil)

	ss := &Session{}

	var err error
	utils.Logger, err = utils.Newlogger(utils.MetaStdLog, utils.EmptyString)
	if err != nil {
		t.Error(err)
	}
	utils.Logger.SetLogLevel(7)

	buff := new(bytes.Buffer)
	log.SetOutput(buff)

	//Cannot set a terminate when ttl is 0
	ss.OptsStart = engine.MapEvent{
		utils.OptsSessionsTTL: "0s",
	}
	opts := engine.MapEvent{}
	sessions.setSTerminator(ss, opts)

	//Invalid format types for time duration
	ss.OptsStart = engine.MapEvent{
		utils.OptsSessionsTTL: "invalid_time_format",
	}
	sessions.setSTerminator(ss, opts)
	expected := "cannot extract <*sessionsTTL> for session:<>, from it's options: <{}>, err: <time: invalid duration \"invalid_time_format\">"
	if rcv := buff.String(); !strings.Contains(rcv, expected) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}

	buff = new(bytes.Buffer)
	log.SetOutput(buff)

	ss.OptsStart = engine.MapEvent{
		utils.OptsSessionsTTL:         "1s",
		utils.OptsSessionsTTLMaxDelay: "invalid_time_format",
	}
	sessions.setSTerminator(ss, opts)
	expected = "cannot extract <*sessionsTTLMaxDelay> for session:<>, from it's options: <{}>, err: <time: invalid duration \"invalid_time_format\">"
	if rcv := buff.String(); !strings.Contains(rcv, expected) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}

	buff = new(bytes.Buffer)
	log.SetOutput(buff)

	ss.OptsStart = engine.MapEvent{
		utils.OptsSessionsTTL:         "1s",
		utils.OptsSessionsTTLMaxDelay: "2s",
		utils.OptsSessionsTTLLastUsed: "invalid_time_format",
	}
	sessions.setSTerminator(ss, opts)
	expected = "cannot extract <*sessionsTTLLastUsed> for session:<>, from it's options: <{}>, err: <time: invalid duration \"invalid_time_format\">"
	if rcv := buff.String(); !strings.Contains(rcv, expected) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}

	buff = new(bytes.Buffer)
	log.SetOutput(buff)

	ss.OptsStart = engine.MapEvent{
		utils.OptsSessionsTTL:          "1s",
		utils.OptsSessionsTTLMaxDelay:  "2s",
		utils.OptsSessionsTTLLastUsed:  "1s",
		utils.OptsSessionsTTLLastUsage: "invalid_time_format",
	}
	sessions.setSTerminator(ss, opts)
	expected = "cannot extract <*sessionsTTLLastUsage> for session:<>, from it's options: <{}>, err: <time: invalid duration \"invalid_time_format\">"
	if rcv := buff.String(); !strings.Contains(rcv, expected) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}

	buff = new(bytes.Buffer)
	log.SetOutput(buff)

	cfg.SessionSCfg().SessionTTLMaxDelay = utils.DurationPointer(time.Second)
	sessions = NewSessionS(cfg, dm, nil)
	ss.OptsStart = engine.MapEvent{
		utils.OptsSessionsTTLLastUsed:  "1s",
		utils.OptsSessionsTTLLastUsage: "5s",
		utils.OptsSessionsTTLUsage:     "invalid_time_format",
	}
	sessions.setSTerminator(ss, opts)
	expected = "cannot extract <*sessionsTTLUsage> for session:<>, from it's options: <{}>, err: <time: invalid duration \"invalid_time_format\">"
	if rcv := buff.String(); !strings.Contains(rcv, expected) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}

	log.SetOutput(os.Stderr)
}

func TestSetSTerminatorAutomaticTermination(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	ss := &Session{}

	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	sessions := NewSessionS(cfg, dm, nil)

	opts := engine.MapEvent{
		utils.OptsSessionsTTL:          "1s",
		utils.OptsSessionsTTLLastUsage: "0s",
	}

	sessions.setSTerminator(ss, opts)
	select {
	case <-time.After(3 * time.Second):
		t.Fatal("timeout")
	case <-ss.sTerminator.endChan:
	}
}

func TestSetSTerminatorManualTermination(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	ss := &Session{}

	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	sessions := NewSessionS(cfg, dm, nil)

	opts := engine.MapEvent{
		utils.OptsSessionsTTL: "1s",
	}

	sessions.setSTerminator(ss, opts)
	ss.sTerminator.endChan <- struct{}{}
}

func TestForceSTerminatorManualTermination(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	ss := &Session{
		CGRID:         "CGRID",
		Tenant:        "cgrates.org",
		ResourceID:    "resourceID",
		ClientConnID:  "ClientConnID",
		EventStart:    engine.NewMapEvent(nil),
		DebitInterval: 18,
		SRuns: []*SRun{
			{Event: engine.NewMapEvent(nil),
				CD:            &engine.CallDescriptor{Category: "test"},
				EventCost:     &engine.EventCost{CGRID: "testCGRID"},
				ExtraDuration: 1,
				LastUsage:     2,
				TotalUsage:    3,
				NextAutoDebit: utils.TimePointer(time.Date(2020, time.April, 18, 23, 0, 0, 0, time.UTC)),
			},
		},
		chargeable: true,
	}

	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	sessions := NewSessionS(cfg, dm, nil)

	expected := "MANDATORY_IE_MISSING: [connIDs]"
	if err := sessions.forceSTerminate(ss, time.Second, nil, nil); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, receive %+v", expected, err)
	}
}

func TestForceSTerminatorPostCDRs(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().CDRsConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs)}
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs): nil,
	})
	sessions := NewSessionS(cfg, dm, connMgr)

	ss := &Session{
		CGRID:      "CGRID",
		Tenant:     "cgrates.org",
		EventStart: engine.NewMapEvent(nil),
		SRuns: []*SRun{
			{
				Event: map[string]interface{}{
					utils.RequestType: utils.MetaPostpaid,
				},
				CD:            &engine.CallDescriptor{Category: "test"},
				EventCost:     &engine.EventCost{CGRID: "testCGRID"},
				ExtraDuration: 1,
				LastUsage:     2,
				TotalUsage:    3,
				NextAutoDebit: utils.TimePointer(time.Date(2020, time.April, 18, 23, 0, 0, 0, time.UTC)),
			},
		},
		chargeable: true,
	}

	expected := "INTERNALLY_DISCONNECTED"
	if err := sessions.forceSTerminate(ss, time.Second, nil, nil); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, receiveD %+v", expected, err)
	}
}

func TestForceSTerminatorReleaseSession(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().ResSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)}
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources): nil,
	})
	sessions := NewSessionS(cfg, dm, connMgr)

	ss := &Session{
		CGRID:      "CGRID",
		Tenant:     "cgrates.org",
		EventStart: engine.NewMapEvent(nil),
		ResourceID: "resourceID",
		SRuns: []*SRun{
			{
				Event: map[string]interface{}{
					utils.RequestType: utils.MetaPostpaid,
				},
				CD:            &engine.CallDescriptor{Category: "test"},
				EventCost:     &engine.EventCost{CGRID: "testCGRID"},
				ExtraDuration: 1,
				LastUsage:     2,
				TotalUsage:    3,
				NextAutoDebit: utils.TimePointer(time.Date(2020, time.April, 18, 23, 0, 0, 0, time.UTC)),
			},
		},
		chargeable: true,
	}

	expected := "MANDATORY_IE_MISSING: [connIDs]"
	if err := sessions.forceSTerminate(ss, time.Second, nil, nil); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, receiveD %+v", expected, err)
	}
}

type testMockClientConn struct {
	*testRPCClientConnection
}

func (sT *testMockClientConn) Call(method string, arg interface{}, rply interface{}) error {
	return utils.ErrNoActiveSession
}

func TestForceSTerminatorClientCall(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	sTestMock := &testMockClientConn{}

	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().NodeID = "ClientConnID"
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources): nil,
	})
	sessions := NewSessionS(cfg, dm, connMgr)
	sessions.RegisterIntBiJConn(sTestMock, utils.EmptyString)

	ss := &Session{
		CGRID:        "CGRID",
		Tenant:       "cgrates.org",
		EventStart:   engine.NewMapEvent(nil),
		ResourceID:   "resourceID",
		ClientConnID: "ClientConnID",
		SRuns: []*SRun{
			{
				Event: map[string]interface{}{
					utils.RequestType: utils.MetaPostpaid,
				},
				CD:            &engine.CallDescriptor{Category: "test"},
				EventCost:     &engine.EventCost{CGRID: "testCGRID"},
				ExtraDuration: 1,
				LastUsage:     2,
				TotalUsage:    3,
				NextAutoDebit: utils.TimePointer(time.Date(2020, time.April, 18, 23, 0, 0, 0, time.UTC)),
			},
		},
		chargeable: true,
	}

	expected := "MANDATORY_IE_MISSING: [connIDs]"
	if err := sessions.forceSTerminate(ss, time.Second, nil, nil); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	time.Sleep(10 * time.Millisecond)
}

func TestDebitSession(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	sessions := NewSessionS(cfg, dm, nil)

	ss := &Session{
		CGRID:      "CGRID",
		Tenant:     "cgrates.org",
		EventStart: engine.NewMapEvent(nil),
		SRuns: []*SRun{
			{
				Event: map[string]interface{}{
					utils.RequestType: utils.MetaPostpaid,
				},
				CD:            &engine.CallDescriptor{Category: "test"},
				EventCost:     &engine.EventCost{CGRID: "testCGRID"},
				ExtraDuration: 1,
				LastUsage:     2,
				TotalUsage:    3,
				NextAutoDebit: utils.TimePointer(time.Date(2020, time.April, 18, 23, 0, 0, 0, time.UTC)),
			},
		},
		chargeable: true,
	}

	//RunIdx cannot be higher than the length of sessions runs
	expectedErr := "sRunIdx out of range"
	if _, err := sessions.debitSession(ss, 2, 0,
		utils.DurationPointer(time.Second)); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}

	//debitReserve must be higher than 0
	if maxDur, err := sessions.debitSession(ss, 0, 0,
		utils.DurationPointer(0)); err != nil {
		t.Error(err)
	} else if maxDur != 0 {
		t.Errorf("Expected %+v, received %+v", 0, maxDur)
	}

	expectedErr = "MANDATORY_IE_MISSING: [connIDs]"
	if _, err := sessions.debitSession(ss, 0, 0,
		utils.DurationPointer(5*time.Second)); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

//mocking for
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

func TestDebitSessionResponderMaxDebit(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	engine.Cache.Clear(nil)
	testMock1 := &testMockClients{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ResponderMaxDebit: func(args interface{}, reply interface{}) error {
				callCost := new(engine.CallCost)
				callCost.Timespans = []*engine.TimeSpan{
					{
						TimeStart: time.Date(2020, 07, 21, 5, 0, 0, 0, time.UTC),
						TimeEnd:   time.Date(2020, 07, 21, 10, 0, 0, 0, time.UTC),
					},
				}
				*(reply.(*engine.CallCost)) = *callCost
				return nil
			},
		},
	}

	sMock := make(chan rpcclient.ClientConnector, 1)
	sMock <- testMock1
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().RALsConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs)}
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs): sMock,
	})

	sessions := NewSessionS(cfg, dm, connMgr)

	ss := &Session{
		CGRID:      "CGRID",
		Tenant:     "cgrates.org",
		EventStart: engine.NewMapEvent(nil),
		SRuns: []*SRun{
			{
				Event: map[string]interface{}{
					utils.RequestType: utils.MetaPostpaid,
				},
				CD: &engine.CallDescriptor{
					Category:  "test",
					LoopIndex: 12,
				},
				EventCost:     &engine.EventCost{CGRID: "testCGRID"},
				ExtraDuration: time.Minute,
				LastUsage:     time.Minute,
				TotalUsage:    3 * time.Minute,
				NextAutoDebit: utils.TimePointer(time.Date(2020, time.April, 18, 23, 0, 0, 0, time.UTC)),
			},
		},
		chargeable: true,
	}

	if maxDur, err := sessions.debitSession(ss, 0, 5*time.Second,
		utils.DurationPointer(time.Second)); err != nil {
		t.Error(err)
	} else if maxDur != 5*time.Second {
		t.Errorf("Expected %+v, received %+v", time.Minute, maxDur)
	}

	ss.SRuns[0].EventCost = nil
	if _, err := sessions.debitSession(ss, 0, 5*time.Minute,
		utils.DurationPointer(time.Minute)); err != nil {
		t.Error(err)
	}

	if _, err := sessions.debitSession(ss, 0, 10*time.Hour,
		utils.DurationPointer(time.Hour)); err != nil {
		t.Error(err)
	}
}

func TestDebitSessionResponderMaxDebitError(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	engine.Cache.Clear(nil)
	sMock := &testMockClients{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ResponderMaxDebit: func(args interface{}, reply interface{}) error {
				return utils.ErrAccountNotFound
			},
			utils.SchedulerSv1ExecuteActionPlans: func(args interface{}, reply interface{}) error {
				return nil
			},
		},
	}

	internalRpcChan := make(chan rpcclient.ClientConnector, 1)
	internalRpcChan <- sMock
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().RALsConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs)}
	cfg.SessionSCfg().SchedulerConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaScheduler)}
	data := engine.NewInternalDB(nil, nil, true)
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs):      internalRpcChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaScheduler): internalRpcChan})
	dm := engine.NewDataManager(data, cfg.CacheCfg(), connMgr)
	sessions := NewSessionS(cfg, dm, connMgr)

	ss := &Session{
		CGRID:      "CGRID",
		Tenant:     "cgrates.org",
		EventStart: engine.NewMapEvent(nil),
		SRuns: []*SRun{
			{
				Event: map[string]interface{}{
					utils.RequestType: utils.MetaDynaprepaid,
				},
				CD:            &engine.CallDescriptor{Category: "test"},
				EventCost:     &engine.EventCost{CGRID: "testCGRID"},
				ExtraDuration: 1,
				LastUsage:     time.Minute,
				TotalUsage:    3 * time.Minute,
				NextAutoDebit: utils.TimePointer(time.Date(2020, time.April, 18, 23, 0, 0, 0, time.UTC)),
			},
		},
		chargeable: true,
	}

	if maxDur, err := sessions.debitSession(ss, 0, 5*time.Minute,
		utils.DurationPointer(time.Second)); err == nil || err != utils.ErrAccountNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrAccountNotFound, err)
	} else if maxDur != 0 {
		t.Errorf("Expected %+v, received %+v", 0, maxDur)
	}

	engine.Cache.Clear(nil)
	sMock.calls[utils.SchedulerSv1ExecuteActionPlans] = func(args interface{}, reply interface{}) error {
		return utils.ErrNotImplemented
	}
	newInternalRpcChan := make(chan rpcclient.ClientConnector, 1)
	newInternalRpcChan <- sMock
	connMgr = engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs):      internalRpcChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaScheduler): internalRpcChan})
	dm = engine.NewDataManager(data, cfg.CacheCfg(), connMgr)
	sessions = NewSessionS(cfg, dm, connMgr)

	if maxDur, err := sessions.debitSession(ss, 0, 5*time.Minute,
		utils.DurationPointer(time.Second)); err == nil || err != utils.ErrNotImplemented {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotImplemented, err)
	} else if maxDur != 0 {
		t.Errorf("Expected %+v, received %+v", 0, maxDur)
	}
}

func TestInitSessionDebitLoops(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	sessions := NewSessionS(cfg, dm, nil)

	ss := &Session{
		CGRID:         "CGRID",
		Tenant:        "cgrates.org",
		EventStart:    engine.NewMapEvent(nil),
		DebitInterval: time.Minute,
		debitStop:     make(chan struct{}, 1),
		SRuns: []*SRun{
			{
				Event: map[string]interface{}{
					utils.RequestType: utils.MetaPrepaid,
				},
				CD:            &engine.CallDescriptor{Category: "test"},
				EventCost:     &engine.EventCost{CGRID: "testCGRID"},
				ExtraDuration: 1,
				LastUsage:     time.Minute,
				TotalUsage:    3 * time.Minute,
				NextAutoDebit: utils.TimePointer(time.Date(2020, time.April, 18, 23, 0, 0, 0, time.UTC)),
			},
		},
		chargeable: true,
	}

	sessions.initSessionDebitLoops(ss)

	ss.debitStop = nil
	sessions.initSessionDebitLoops(ss)
}

type testMockClientConnDiscSess struct {
	*testRPCClientConnection
}

func (sT *testMockClientConnDiscSess) Call(method string, arg interface{}, rply interface{}) error {
	return nil
}

func TestDebitLoopSessionErrorDebiting(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().NodeID = "ClientConnIdtest"
	cfg.SessionSCfg().TerminateAttempts = 1
	cfg.SessionSCfg().RALsConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs)}
	cfg.SessionSCfg().SchedulerConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaScheduler)}
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	sessions := NewSessionS(cfg, dm, nil)

	ss := &Session{
		CGRID:         "CGRID",
		Tenant:        "cgrates.org",
		ClientConnID:  "ClientConnIdtest",
		EventStart:    engine.NewMapEvent(nil),
		DebitInterval: time.Minute,
		SRuns: []*SRun{
			{
				Event: map[string]interface{}{
					utils.RequestType: utils.MetaDynaprepaid,
				},
				CD:            &engine.CallDescriptor{Category: "test"},
				EventCost:     &engine.EventCost{CGRID: "testCGRID"},
				ExtraDuration: 1,
				LastUsage:     time.Minute,
				TotalUsage:    3 * time.Minute,
				NextAutoDebit: utils.TimePointer(time.Date(2020, time.April, 18, 23, 0, 0, 0, time.UTC)),
			},
		},
		chargeable: true,
	}

	// session already closed
	_, err := sessions.debitLoopSession(ss, 0, time.Hour)
	if err != nil {
		t.Error(err)
	}

	ss.debitStop = make(chan struct{})
	engine.Cache.Clear(nil)
	sMock := &testMockClients{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ResponderMaxDebit: func(args interface{}, reply interface{}) error {
				return utils.ErrAccountNotFound
			},
			utils.SchedulerSv1ExecuteActionPlans: func(args interface{}, reply interface{}) error {
				return utils.ErrUnauthorizedDestination
			},
		},
	}
	internalRpcChan := make(chan rpcclient.ClientConnector, 1)
	internalRpcChan <- sMock
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs):      internalRpcChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaScheduler): internalRpcChan})
	dm = engine.NewDataManager(data, cfg.CacheCfg(), connMgr)
	sessions = NewSessionS(cfg, dm, connMgr)

	sTestMock := &testMockClientConnDiscSess{}
	sessions.RegisterIntBiJConn(sTestMock, utils.EmptyString)

	if _, err = sessions.debitLoopSession(ss, 0, time.Hour); err != nil {
		t.Error(err)
	}
}

func TestDebitLoopSession(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	engine.Cache.Clear(nil)
	testMock1 := &testMockClients{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ResponderMaxDebit: func(args interface{}, reply interface{}) error {
				callCost := new(engine.CallCost)
				callCost.Timespans = []*engine.TimeSpan{
					{
						TimeStart: time.Date(2020, 07, 21, 5, 0, 0, 0, time.UTC),
						TimeEnd:   time.Date(2020, 07, 21, 10, 0, 0, 0, time.UTC),
					},
				}
				*(reply.(*engine.CallCost)) = *callCost
				return nil
			},
		},
	}

	sMock := make(chan rpcclient.ClientConnector, 1)
	sMock <- testMock1
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().RALsConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs)}
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs): sMock,
	})

	sessions := NewSessionS(cfg, dm, connMgr)

	ss := &Session{
		CGRID:      "CGRID",
		Tenant:     "cgrates.org",
		EventStart: engine.NewMapEvent(nil),
		debitStop:  make(chan struct{}),
		SRuns: []*SRun{
			{
				Event: map[string]interface{}{
					utils.RequestType: utils.MetaPostpaid,
				},
				CD: &engine.CallDescriptor{
					Category:  "test",
					LoopIndex: 12,
				},
				EventCost:     &engine.EventCost{CGRID: "testCGRID"},
				ExtraDuration: time.Minute,
				LastUsage:     10 * time.Second,
				TotalUsage:    3 * time.Minute,
				NextAutoDebit: utils.TimePointer(time.Date(2020, time.April, 18, 23, 0, 0, 0, time.UTC)),
			},
		},
		chargeable: true,
	}
	go func() {
		if _, err := sessions.debitLoopSession(ss, 0, time.Second); err != nil {
			t.Error(err)
		}
	}()
	time.Sleep(2 * time.Second)
}

func TestDebitLoopSessionFrcDiscLowerDbtInterval(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	engine.Cache.Clear(nil)
	testMock1 := &testMockClients{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ResponderMaxDebit: func(args interface{}, reply interface{}) error {
				callCost := new(engine.CallCost)
				callCost.Timespans = []*engine.TimeSpan{
					{
						TimeStart: time.Date(2020, 07, 21, 5, 0, 0, 0, time.UTC),
						TimeEnd:   time.Date(2020, 07, 21, 10, 0, 0, 0, time.UTC),
					},
				}
				*(reply.(*engine.CallCost)) = *callCost
				return nil
			},
		},
	}

	sMock := make(chan rpcclient.ClientConnector, 1)
	sMock <- testMock1
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().RALsConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs)}
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs): sMock,
	})

	sessions := NewSessionS(cfg, dm, connMgr)

	ss := &Session{
		CGRID:      "CGRID",
		Tenant:     "cgrates.org",
		EventStart: engine.NewMapEvent(nil),
		debitStop:  make(chan struct{}),
		SRuns: []*SRun{
			{
				Event: map[string]interface{}{
					utils.RequestType: utils.MetaPostpaid,
				},
				CD: &engine.CallDescriptor{
					Category:  "test",
					LoopIndex: 12,
				},
				EventCost:     &engine.EventCost{CGRID: "testCGRID"},
				ExtraDuration: time.Minute,
				LastUsage:     10 * time.Second,
				TotalUsage:    3 * time.Minute,
				NextAutoDebit: utils.TimePointer(time.Date(2020, time.April, 18, 23, 0, 0, 0, time.UTC)),
			},
		},
		chargeable: true,
	}
	go func() {
		if _, err := sessions.debitLoopSession(ss, 0, time.Second); err != nil {
			t.Error(err)
		}
	}()
	ss.debitStop <- struct{}{}
}

func TestDebitLoopSessionLowBalance(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	engine.Cache.Clear(nil)
	testMock1 := &testMockClients{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ResponderMaxDebit: func(args interface{}, reply interface{}) error {
				return nil
			},
		},
	}

	sMock := make(chan rpcclient.ClientConnector, 1)
	sMock <- testMock1
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().RALsConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs)}
	cfg.SessionSCfg().MinDurLowBalance = 1 * time.Second
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs): sMock,
	})

	sessions := NewSessionS(cfg, dm, connMgr)

	ss := &Session{
		CGRID:      "CGRID",
		Tenant:     "cgrates.org",
		EventStart: engine.NewMapEvent(nil),
		debitStop:  make(chan struct{}),
		SRuns: []*SRun{
			{
				Event: map[string]interface{}{
					utils.RequestType: utils.MetaPostpaid,
				},
				CD: &engine.CallDescriptor{
					Category:  "test",
					LoopIndex: 12,
				},
				EventCost:     nil, //without an EventCost
				ExtraDuration: 30 * time.Millisecond,
				LastUsage:     10 * time.Second,
				TotalUsage:    3 * time.Minute,
				NextAutoDebit: utils.TimePointer(time.Date(2020, time.April, 18, 23, 0, 0, 0, time.UTC)),
			},
		},
		chargeable: true,
	}

	sessions.cgrCfg.SessionSCfg().MinDurLowBalance = 10 * time.Second
	// will disconnect faster, MinDurLowBalance higher than the debit interval
	go func() {
		if _, err := sessions.debitLoopSession(ss, 0, 50*time.Millisecond); err != nil {
			t.Error(err)
		}
	}()
	time.Sleep(1 * time.Second)
}

func TestDebitLoopSessionWarningSessions(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	engine.Cache.Clear(nil)
	testMock1 := &testMockClients{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ResponderMaxDebit: func(args interface{}, reply interface{}) error {
				return nil
			},
			utils.ResourceSv1ReleaseResources: func(args interface{}, reply interface{}) error {
				return utils.ErrNotImplemented
			},
		},
	}

	sMock := make(chan rpcclient.ClientConnector, 1)
	sMock <- testMock1
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().RALsConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs)}
	cfg.SessionSCfg().ResSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)}
	cfg.SessionSCfg().MinDurLowBalance = 1 * time.Second
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs):      sMock,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources): sMock})

	sessions := NewSessionS(cfg, dm, connMgr)

	ss := &Session{
		CGRID:         "CGRID",
		Tenant:        "cgrates.org",
		ResourceID:    "resourceID",
		ClientConnID:  "ClientConnID",
		debitStop:     make(chan struct{}),
		EventStart:    engine.NewMapEvent(nil),
		DebitInterval: 18,
		SRuns: []*SRun{
			{Event: engine.NewMapEvent(nil),
				CD:            &engine.CallDescriptor{Category: "test"},
				EventCost:     &engine.EventCost{CGRID: "testCGRID"},
				ExtraDuration: 1,
				LastUsage:     2,
				TotalUsage:    3,
				NextAutoDebit: utils.TimePointer(time.Date(2020, time.April, 18, 23, 0, 0, 0, time.UTC)),
			},
		},
		chargeable: true,
	}

	// will disconnect faster, MinDurLowBalance higher than the debit interval
	expected := "UNSUPPORTED_SERVICE_METHOD"
	if _, err := sessions.debitLoopSession(ss, 0, 2*time.Second); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestDebitLoopSessionDisconnectSession(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	engine.Cache.Clear(nil)
	testMock1 := &testMockClients{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ResponderMaxDebit: func(args interface{}, reply interface{}) error {
				return nil
			},
			utils.ResourceSv1ReleaseResources: func(args interface{}, reply interface{}) error {
				return utils.ErrNotImplemented
			},
		},
	}

	sMock := make(chan rpcclient.ClientConnector, 1)
	sMock <- testMock1
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().NodeID = "ClientConnID"
	cfg.SessionSCfg().RALsConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs)}
	cfg.SessionSCfg().ResSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)}
	cfg.SessionSCfg().MinDurLowBalance = 1 * time.Second
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs):      sMock,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources): sMock})

	sessions := NewSessionS(cfg, dm, connMgr)

	sTestMock := &testMockClientConnDiscSess{}
	sessions.RegisterIntBiJConn(sTestMock, utils.EmptyString)

	ss := &Session{
		CGRID:         "CGRID",
		Tenant:        "cgrates.org",
		ResourceID:    "resourceID",
		ClientConnID:  "ClientConnID",
		debitStop:     make(chan struct{}),
		EventStart:    engine.NewMapEvent(nil),
		DebitInterval: 18,
		SRuns: []*SRun{
			{
				Event: engine.NewMapEvent(nil),
				CD:    &engine.CallDescriptor{},
				EventCost: &engine.EventCost{
					Usage: utils.DurationPointer(5 * time.Hour),
				},
				ExtraDuration: 1,
				LastUsage:     2,
				TotalUsage:    3,
				NextAutoDebit: utils.TimePointer(time.Date(2020, time.April, 18, 23, 0, 0, 0, time.UTC)),
			},
		},
		chargeable: true,
	}

	// will disconnect faster
	if _, err := sessions.debitLoopSession(ss, 0, 2*time.Second); err != nil {
		t.Error(err)
	}

	//force disconnect
	go func() {
		if _, err := sessions.debitLoopSession(ss, 0, 2*time.Second); err != nil {
			t.Error(err)
		}
	}()
	ss.debitStop <- struct{}{}
}

func TestStoreSCost(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	engine.Cache.Clear(nil)
	testMock1 := &testMockClients{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.CDRsV1StoreSessionCost: func(args interface{}, reply interface{}) error {
				return utils.ErrExists
			},
		},
	}

	sMock := make(chan rpcclient.ClientConnector, 1)
	sMock <- testMock1
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().CDRsConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs)}
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs): sMock})

	sessions := NewSessionS(cfg, dm, connMgr)

	ss := &Session{
		CGRID:  "CGRID",
		Tenant: "cgrates.org",
		SRuns: []*SRun{
			{
				Event: engine.NewMapEvent(nil),
				CD: &engine.CallDescriptor{
					TimeStart: time.Date(2020, 07, 21, 10, 0, 0, 0, time.UTC),
					TimeEnd:   time.Date(2020, 07, 21, 12, 0, 0, 0, time.UTC),
				},
				EventCost: &engine.EventCost{
					Usage:          utils.DurationPointer(5 * time.Hour),
					Charges:        []*engine.ChargingInterval{},
					AccountSummary: &engine.AccountSummary{},
				},
			},
		},
		chargeable: true,
	}

	expected := "cannot find last active ChargingInterval"
	if err := sessions.storeSCost(ss, 0); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestRefundSession(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	engine.Cache.Clear(nil)
	testMock1 := &testMockClients{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ResponderRefundIncrements: func(args interface{}, reply interface{}) error {
				if args.(*engine.CallDescriptorWithAPIOpts).APIOpts != nil {
					return utils.ErrNotImplemented
				}
				acnt := &engine.Account{
					ID: "cgrates_test",
				}
				*reply.(*engine.Account) = *acnt
				return nil
			},
		},
	}

	sMock := make(chan rpcclient.ClientConnector, 1)
	sMock <- testMock1
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().RALsConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs)}
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs): sMock})

	sessions := NewSessionS(cfg, dm, connMgr)

	ss := &Session{
		CGRID:  "CGRID",
		Tenant: "cgrates.org",
		SRuns: []*SRun{
			{
				Event: engine.NewMapEvent(nil),
				CD: &engine.CallDescriptor{
					TimeStart: time.Date(2020, 07, 21, 11, 0, 0, 0, time.UTC),
					TimeEnd:   time.Date(2020, 07, 21, 11, 0, 30, 0, time.UTC),
				},
			},
		},
		chargeable: true,
	}

	expectedErr := "no event cost"
	//event Cost is empty
	if err := sessions.refundSession(ss, 0, 0); err == nil || err.Error() != expectedErr {
		t.Error(err)
	}

	//index run cannot be higher than the runs in sessions
	expectedErr = "sRunIdx out of range"
	if err := sessions.refundSession(ss, 1, 0); err == nil || err.Error() != expectedErr {
		t.Error(err)
	}

	ss.SRuns[0].EventCost = &engine.EventCost{
		AccountSummary: &engine.AccountSummary{},
		Usage:          utils.DurationPointer(30 * time.Second),
		Charges: []*engine.ChargingInterval{
			{
				RatingID: "21a5ab9",
				Increments: []*engine.ChargingIncrement{
					{
						Usage:          time.Duration(1 * time.Second),
						Cost:           0.005,
						AccountingID:   "44d6c02",
						CompressFactor: 30,
					},
				},
				CompressFactor: 1,
			},
		},
		Rating: map[string]*engine.RatingUnit{
			"21a5ab9": {},
		},
		Accounting: map[string]*engine.BalanceCharge{
			"44d6c02": {},
		},
	}

	//new EventCost will be empty
	if err := sessions.refundSession(ss, 0, 0); err != nil {
		t.Error(err)
	}

	expectedErr = "failed detecting last active ChargingInterval"
	if err := sessions.refundSession(ss, 0, 5*time.Minute); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}

	if err := sessions.refundSession(ss, 0, time.Second); err != nil {
		t.Error(err)
	}

	//mocking an error for calling
	ss.OptsStart = engine.MapEvent{}
	if err := sessions.refundSession(ss, 0, 2); err == nil || err != utils.ErrNotImplemented {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotImplemented, err)
	}

	//are are no increments to refund
	ss.OptsStart = nil
	ss.SRuns[0].EventCost.Charges[0].Increments[0] = &engine.ChargingIncrement{}
	ss.SRuns[0].CD.TimeStart = time.Date(2020, 07, 21, 11, 0, 30, 0, time.UTC)
	ss.SRuns[0].CD.TimeEnd = time.Date(2020, 07, 21, 11, 0, 30, 0, time.UTC)
	ss.SRuns[0].EventCost.Usage = utils.DurationPointer(2)
	if err := sessions.refundSession(ss, 0, 2); err != nil {
		t.Error(err)
	}
}

func TestRoundCost(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	engine.Cache.Clear(nil)
	testMock1 := &testMockClients{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ResponderRefundRounding: func(args interface{}, reply interface{}) error {
				return utils.ErrNotImplemented
			},
		},
	}

	sMock := make(chan rpcclient.ClientConnector, 1)
	sMock <- testMock1
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().RALsConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs)}
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs): sMock})

	sessions := NewSessionS(cfg, dm, connMgr)

	ss := &Session{
		CGRID:  "CGRID",
		Tenant: "cgrates.org",
		SRuns: []*SRun{
			{
				EventCost: &engine.EventCost{
					AccountSummary: &engine.AccountSummary{},
					Usage:          utils.DurationPointer(30 * time.Second),
					Charges: []*engine.ChargingInterval{
						{
							RatingID: "21a5ab9",
							Increments: []*engine.ChargingIncrement{
								{
									Usage:          time.Duration(1 * time.Second),
									Cost:           0.005,
									AccountingID:   "44d6c02",
									CompressFactor: 30,
								},
								{
									Usage:          time.Duration(1 * time.Second),
									Cost:           0.010,
									AccountingID:   "7hslkif",
									CompressFactor: 50,
								},
							},
							CompressFactor: 1,
						},
					},
					Rating: map[string]*engine.RatingUnit{
						"21a5ab9":          {},
						utils.MetaRounding: {},
					},
					Accounting: map[string]*engine.BalanceCharge{
						"44d6c02": {
							RatingID: utils.MetaRounding,
						},
						"7hslkif": {
							RatingID: utils.MetaRounding,
						},
					},
				},
			},
		},
		chargeable: true,
	}

	//mocking an error API Call
	if err := sessions.roundCost(ss, 0); err != utils.ErrNotImplemented || err == nil {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotImplemented, err)
	}
}

func TestDisconnectSession(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	sessions := NewSessionS(cfg, dm, nil)

	ss := &Session{
		ClientConnID: "test",
		EventStart:   make(map[string]interface{}),
		SRuns: []*SRun{
			{
				TotalUsage: time.Minute,
			},
		},
		chargeable: true,
	}

	sTestMock := &testMockClientConn{}
	sessions.RegisterIntBiJConn(sTestMock, utils.EmptyString)
	sessions.biJIDs["test"] = &biJClient{
		conn: sTestMock,
	}

	if err := sessions.disconnectSession(ss, utils.EmptyString); err == nil || err != utils.ErrNoActiveSession {
		t.Errorf("Expected %+v, received %+v", utils.ErrNoActiveSession, err)
	}

	sTestMock1 := &mockConnWarnDisconnect1{}
	sessions.RegisterIntBiJConn(sTestMock1, utils.EmptyString)
	sessions.biJIDs["test"] = &biJClient{
		conn: sTestMock1,
	}
	if err := sessions.disconnectSession(ss, utils.EmptyString); err != nil {
		t.Error(err)
	}
}

func TestReplicateSessions(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	engine.Cache.Clear(nil)
	testMock1 := &testMockClients{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.SessionSv1SetPassiveSession: func(args interface{}, reply interface{}) error {
				return utils.ErrNotImplemented
			},
		},
	}

	sMock := make(chan rpcclient.ClientConnector, 1)
	sMock <- testMock1
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator): sMock})

	sessions := NewSessionS(cfg, dm, connMgr)

	sessions.replicateSessions("test_session", false,
		[]string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)})
}

func TestNewSession(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	engine.Cache.Clear(nil)
	testMock1 := &testMockClients{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ChargerSv1ProcessEvent: func(args interface{}, reply interface{}) error {
				if args.(*utils.CGREvent).ID == utils.EmptyString {
					return utils.ErrNotImplemented
				}
				chrgrs := []*engine.ChrgSProcessEventReply{
					{ChargerSProfile: "TEST_PROFILE1",
						CGREvent: &utils.CGREvent{
							Tenant: "cgrates.org",
							ID:     "TEST_ID",
							Event: map[string]interface{}{
								utils.Destination: "10",
							},
						}},
				}
				*reply.(*[]*engine.ChrgSProcessEventReply) = chrgrs
				return nil
			},
		},
	}
	sMock := make(chan rpcclient.ClientConnector, 1)
	sMock <- testMock1
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers): sMock})
	dm := engine.NewDataManager(data, cfg.CacheCfg(), connMgr)

	sessions := NewSessionS(cfg, dm, connMgr)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TEST_ID",
		Event: map[string]interface{}{
			utils.Destination: "10",
		},
	}

	expectedErr := "ChargerS is disabled"
	if _, err := sessions.newSession(cgrEv, "resourceID", "clientConnID",
		time.Second, false, false); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}

	cfg.SessionSCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)}

	expectedSess := &Session{
		CGRID:        "da39a3ee5e6b4b0d3255bfef95601890afd80709",
		Tenant:       "cgrates.org",
		ResourceID:   "resourceID",
		ClientConnID: "clientConnID",
		EventStart: map[string]interface{}{
			utils.CGRID:       "da39a3ee5e6b4b0d3255bfef95601890afd80709",
			utils.Destination: "10",
		},
		DebitInterval: time.Second,
		SRuns: []*SRun{
			{
				Event: map[string]interface{}{
					utils.Destination: "10",
				},
				CD: &engine.CallDescriptor{
					CgrID:       "da39a3ee5e6b4b0d3255bfef95601890afd80709",
					Tenant:      "cgrates.org",
					Category:    "call",
					Destination: "10",
					ExtraFields: map[string]string{},
				},
			},
		},
		chargeable: true,
	}
	if rcv, err := sessions.newSession(cgrEv, "resourceID", "clientConnID",
		time.Second, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, expectedSess) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedSess), utils.ToJSON(rcv))
	}

	//error in mocking the call from connMgr
	cgrEv.ID = utils.EmptyString
	if _, err := sessions.newSession(cgrEv, "resourceID", "clientConnID",
		time.Second, false, false); err == nil || err.Error() != utils.NewErrChargerS(utils.ErrNotImplemented).Error() {
		t.Errorf("Expected %+v, received %+v", utils.NewErrChargerS(utils.ErrNotImplemented), err)
	}

	sessions.aSessions = map[string]*Session{
		"da39a3ee5e6b4b0d3255bfef95601890afd80709": {},
	}
	//sessions already exists
	if _, err := sessions.newSession(cgrEv, "resourceID", "clientConnID",
		time.Second, false, false); err == nil || err.Error() != utils.ErrExists.Error() {
		t.Errorf("Expected %+v, received %+v", utils.ErrExists, err)
	}
}

func TestProcessChargerS(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	tmpCache := engine.Cache

	engine.Cache.Clear(nil)
	testMock1 := &testMockClients{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ChargerSv1ProcessEvent: func(args interface{}, reply interface{}) error {
				return utils.ErrExists
			},
			utils.CacheSv1ReplicateSet: func(args interface{}, reply interface{}) error {
				return utils.ErrNotImplemented
			},
		},
	}
	sMock := make(chan rpcclient.ClientConnector, 1)
	sMock <- testMock1
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers): sMock})
	dm := engine.NewDataManager(data, cfg.CacheCfg(), connMgr)

	sessions := NewSessionS(cfg, dm, connMgr)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TEST_ID",
		Event: map[string]interface{}{
			utils.Destination: "10",
		},
	}

	expected := "CHARGERS_ERROR:MANDATORY_IE_MISSING: [connIDs]"
	if _, err := sessions.processChargerS(cgrEv); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	if _, err := sessions.processChargerS(cgrEv); err != nil {
		t.Error(err)
	}

	engine.Cache.Clear(nil)
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.CacheCfg().Partitions[utils.CacheEventCharges] = &config.CacheParamCfg{
		Replicate: true,
	}
	cacheS := engine.NewCacheS(cfg, nil, nil)
	engine.Cache = cacheS
	connMgr = engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator): sMock})
	engine.SetConnManager(connMgr)

	if _, err := sessions.processChargerS(cgrEv); err == nil || err != utils.ErrNotImplemented {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotImplemented, err)
	}

	engine.Cache = tmpCache
}

func TestTransitSState(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)

	sessions := NewSessionS(cfg, dm, nil)

	rcv := sessions.transitSState("test", true)
	if rcv != nil {
		t.Error("Expected to be nil")
	}

	sessions.pSessions = map[string]*Session{
		"test": {
			CGRID: "TEST_CGRID",
		},
	}
	expected := &Session{
		CGRID: "TEST_CGRID",
	}

	rcv = sessions.getActivateSession("test")
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestRelocateSession(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)

	sessions := NewSessionS(cfg, dm, nil)

	if rcv := sessions.relocateSession(utils.EmptyString, "222", "127.0.0.1"); rcv != nil {
		t.Errorf("Expected to be nil")
	}

	sessions.cgrCfg.SessionSCfg().SessionIndexes = map[string]struct{}{}
	sessions.aSessions = map[string]*Session{
		"0d0fe8779b54c88f121e26c5d83abee5935127e5": {
			CGRID:      "TEST_CGRID",
			EventStart: map[string]interface{}{},
			SRuns: []*SRun{
				{
					Event: map[string]interface{}{},
				},
			},
		},
	}
	expected := &Session{
		CGRID: "dfa2adaa5ab49349777c1ab3bcf3455df0259880",
		EventStart: map[string]interface{}{
			utils.CGRID:    "dfa2adaa5ab49349777c1ab3bcf3455df0259880",
			utils.OriginID: "222",
		},
		SRuns: []*SRun{
			{
				Event: map[string]interface{}{
					utils.CGRID:    "dfa2adaa5ab49349777c1ab3bcf3455df0259880",
					utils.OriginID: "222",
				},
			},
		},
	}
	if rcv := sessions.relocateSession("111", "222", "127.0.0.1"); rcv == nil {
		t.Errorf("Expected to not be nil")
	} else if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	sessions.pSessions = map[string]*Session{
		"0d0fe8779b54c88f121e26c5d83abee5935127e5": nil,
	}

	rcv := sessions.relocateSession("111", "222", utils.EmptyString)
	if rcv != nil {
		t.Errorf("Expected to be nil")
	}
}

func TestGetRelocateSession(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)

	sessions := NewSessionS(cfg, dm, nil)

	rcv := sessions.getRelocateSession("test", "111", "222", "127.0.0.1")
	if rcv != nil {
		t.Errorf("Expected to be nil")
	}

	sessions.pSessions = map[string]*Session{
		"test": {
			CGRID: "TEST_CGRID",
		},
	}

	expected := &Session{
		CGRID: "TEST_CGRID",
	}
	if rcv = sessions.getRelocateSession("test", utils.EmptyString, "222", "127.0.0.1"); rcv == nil {
		t.Errorf("Expected to be nil")
	} else if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestLibsessionsSetMockErrors(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	tmp := engine.Cache

	engine.Cache.Clear(nil)
	sTestMock := &testMockClients{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.CacheSv1ReplicateSet: func(args interface{}, reply interface{}) error {
				return utils.ErrNotImplemented
			},
		},
	}
	chanInternal := make(chan rpcclient.ClientConnector, 1)
	chanInternal <- sTestMock
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.CacheCfg().Partitions[utils.CacheSTIR] = &config.CacheParamCfg{
		Replicate: true,
	}
	cacheS := engine.NewCacheS(cfg, nil, nil)
	engine.Cache = cacheS
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator): chanInternal})
	engine.SetConnManager(connMgr)

	procIndt, err := NewProcessedIdentity("eyJhbGciOiJFUzI1NiIsInBwdCI6InNoYWtlbiIsInR5cCI6InBhc3Nwb3J0IiwieDV1IjoiaHR0cHM6Ly93d3cuZXhhbXBsZS5vcmcvY2VydC5jZXIifQ.eyJhdHRlc3QiOiJBIiwiZGVzdCI6eyJ0biI6WyIxMDAyIl19LCJpYXQiOjE1ODcwMTk4MjIsIm9yaWciOnsidG4iOiIxMDAxIn0sIm9yaWdpZCI6IjEyMzQ1NiJ9.4ybtWmgqdkNyJLS9Iv3PuJV8ZxR7yZ_NEBhCpKCEu2WBiTchqwoqoWpI17Q_ALm38tbnpay32t95ZY_LhSgwJg;info=<https://www.example.org/cert.cer>;ppt=shaken")
	if err != nil {
		t.Error(err)
	}
	if err := procIndt.VerifySignature(cfg.GeneralCfg().ReplyTimeout); err == nil || err != utils.ErrNotImplemented {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotImplemented, err)
	}

	procIndt.Header.X5u = "https://raw.githubusercontent.com/cgrates/cgrates/master/data/stir/stir_pubkey.pem"
	if err := procIndt.VerifySignature(cfg.GeneralCfg().ReplyTimeout); err == nil || err != utils.ErrNotImplemented {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotImplemented, err)
	}

	if _, err := NewSTIRIdentity(procIndt.Header, procIndt.Payload, "https://raw.githubusercontent.com/cgrates/cgrates/master/data/stir/stir_privatekey.pem",
		-1); err == nil || err != utils.ErrNotImplemented {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotImplemented, err)
	}

	if _, err := NewSTIRIdentity(procIndt.Header, procIndt.Payload, "https://raw.githubusercontent.com/cgrates/cgrates/master/data/stir/stir_privatekey.pe",
		-1); err == nil || err != utils.ErrNotImplemented {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotImplemented, err)
	}

	engine.Cache = tmp
}

type testMockClientSyncSessions struct {
	*testRPCClientConnection
}

func (sT *testMockClientSyncSessions) Call(method string, arg interface{}, rply interface{}) error {
	queriedSessionIDs := []*SessionID{
		{
			OriginID:   "ORIGIN_ID",
			OriginHost: "ORIGIN_HOST",
		},
	}
	*rply.(*[]*SessionID) = queriedSessionIDs
	time.Sleep(20)
	return utils.ErrNoActiveSession
}

func TestSyncSessions(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	tmp := engine.Cache
	engine.Cache.Clear(nil)

	sTestMock := &testMockClients{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ResourceSv1ReleaseResources: func(args interface{}, reply interface{}) error {
				return utils.ErrNotImplemented
			},
			utils.CacheSv1ReplicateSet: func(args interface{}, reply interface{}) error {
				return utils.ErrNotImplemented
			},
		},
	}
	chanInternal := make(chan rpcclient.ClientConnector, 1)
	chanInternal <- sTestMock
	cfg := config.NewDefaultCGRConfig()
	//cfg.GeneralCfg().ReplyTimeout = 1
	cfg.SessionSCfg().ResSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)}
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.CacheCfg().Partitions[utils.CacheClosedSessions] = &config.CacheParamCfg{
		Replicate: true,
	}
	data := engine.NewInternalDB(nil, nil, true)
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources): chanInternal})
	dm := engine.NewDataManager(data, cfg.CacheCfg(), connMgr)
	sessions := NewSessionS(cfg, dm, connMgr)

	sTestMock1 := &testMockClientSyncSessions{}
	sessions.RegisterIntBiJConn(sTestMock1, utils.EmptyString)

	sessions.aSessions = map[string]*Session{
		"SESS1": {
			CGRID: "TEST_CGRID",
		},
	}

	sessions.syncSessions()

	sessions.cgrCfg.GeneralCfg().ReplyTimeout = 1
	cacheS := engine.NewCacheS(cfg, nil, nil)
	engine.Cache = cacheS
	connMgr = engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator): chanInternal})
	engine.SetConnManager(connMgr)
	sessions.aSessions = map[string]*Session{
		"ORIGIN_ID": {},
	}

	var reply string
	if err := sessions.BiRPCv1SyncSessions(nil, nil, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected to be OK")
	}

	engine.Cache = tmp

	//There are no sessions to be removed
	sessions.terminateSyncSessions([]string{"no_sesssion"})
}

func TestAuthEvent(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	engine.Cache.Clear(nil)

	sTestMock := &testMockClients{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ResponderGetMaxSessionTime: func(args interface{}, reply interface{}) error {
				return utils.ErrNotImplemented
			},
			utils.ChargerSv1ProcessEvent: func(args interface{}, reply interface{}) error {
				chrgrs := []*engine.ChrgSProcessEventReply{
					{ChargerSProfile: "TEST_PROFILE1",
						CGREvent: &utils.CGREvent{
							Tenant: "cgrates.org",
							ID:     "TEST_ID",
							Event: map[string]interface{}{
								utils.Destination: "10",
							},
						}},
				}

				*reply.(*[]*engine.ChrgSProcessEventReply) = chrgrs
				return nil
			},
		},
	}
	chanInternal := make(chan rpcclient.ClientConnector, 1)
	chanInternal <- sTestMock
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().RALsConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs)}
	data := engine.NewInternalDB(nil, nil, true)
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs):     chanInternal,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers): chanInternal})
	dm := engine.NewDataManager(data, cfg.CacheCfg(), connMgr)
	sessions := NewSessionS(cfg, dm, connMgr)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TEST_ID",
		Event: map[string]interface{}{
			utils.Destination: "10",
			utils.Usage:       "invalid_time",
		},
	}

	expected := "time: invalid duration \"invalid_time\""
	if _, err := sessions.authEvent(cgrEv, true); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	delete(cgrEv.Event, utils.Usage)
	expected = "ChargerS is disabled"
	if _, err := sessions.authEvent(cgrEv, true); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	sessions.cgrCfg.SessionSCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)}
	expectedTime := map[string]time.Duration{
		utils.EmptyString: 3 * time.Hour,
	}
	if usage, err := sessions.authEvent(cgrEv, true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(usage, expectedTime) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expectedTime), utils.ToJSON(usage))
	}

}

func TestAuthEventMockCall(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	//mocking the GetMaxSession for checking the error
	engine.Cache.Clear(nil)
	sTestMock := &testMockClients{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ResponderGetMaxSessionTime: func(args interface{}, reply interface{}) error {
				usage := args.(*engine.CallDescriptorWithAPIOpts).APIOpts[utils.Usage]
				if usage != 10 {
					return utils.ErrNoMoreData
				}
				*reply.(*time.Duration) = 4 * time.Hour
				return nil
			},
			utils.ChargerSv1ProcessEvent: func(args interface{}, reply interface{}) error {
				chrgrs := []*engine.ChrgSProcessEventReply{
					{ChargerSProfile: "TEST_PROFILE1",
						CGREvent: &utils.CGREvent{
							Tenant: "cgrates.org",
							ID:     "TEST_ID",
							Event: map[string]interface{}{
								utils.Destination: "10",
								utils.RequestType: utils.MetaPrepaid,
							},
						}},
				}
				*reply.(*[]*engine.ChrgSProcessEventReply) = chrgrs
				return nil
			},
		},
	}
	chanInternal := make(chan rpcclient.ClientConnector, 1)
	chanInternal <- sTestMock
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().RALsConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs)}
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs):     chanInternal,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers): chanInternal})
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), connMgr)
	sessions := NewSessionS(cfg, dm, connMgr)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TEST_ID",
		Event: map[string]interface{}{
			utils.Destination: "10",
		},
		Opts: map[string]interface{}{
			utils.Usage: 10,
		},
	}

	sessions.cgrCfg.SessionSCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)}
	expectedTime := map[string]time.Duration{
		utils.EmptyString: 3 * time.Hour,
	}
	if usage, err := sessions.authEvent(cgrEv, true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(usage, expectedTime) {
		t.Errorf("Expected %+v, received %+v", expectedTime, usage)
	}

	cgrEv.Opts = map[string]interface{}{
		utils.Usage: 20,
	}
	expected := "RALS_ERROR:NO_MORE_DATA"
	if _, err := sessions.authEvent(cgrEv, true); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestChargeEvent(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	tmp := engine.Cache

	engine.Cache.Clear(nil)
	sTestMock := &testMockClients{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.CacheSv1ReplicateSet: func(args interface{}, reply interface{}) error {
				return utils.ErrNotImplemented
			},
			utils.ChargerSv1ProcessEvent: func(args interface{}, reply interface{}) error {
				chrgrs := []*engine.ChrgSProcessEventReply{
					{ChargerSProfile: "TEST_PROFILE1",
						CGREvent: &utils.CGREvent{
							Tenant: "cgrates.org",
							ID:     "TEST_ID",
							Event: map[string]interface{}{
								utils.Destination: "10",
								utils.RequestType: utils.MetaPostpaid,
							},
						}},
				}
				if args.(*utils.CGREvent).Tenant != "cgrates.org" {
					chrgrs[0].CGREvent.Event[utils.RequestType] = utils.MetaPrepaid
				}
				*reply.(*[]*engine.ChrgSProcessEventReply) = chrgrs
				return nil
			},
			utils.ResponderMaxDebit: func(args interface{}, reply interface{}) error {
				return nil
			},
		},
	}
	chanInternal := make(chan rpcclient.ClientConnector, 1)
	chanInternal <- sTestMock
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.CacheCfg().Partitions[utils.CacheClosedSessions] = &config.CacheParamCfg{
		Replicate: true,
	}
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs):     chanInternal,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers): chanInternal,
	})
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), connMgr)
	sessions := NewSessionS(cfg, dm, connMgr)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TEST_ID",
		Event: map[string]interface{}{
			utils.Usage:       10,
			utils.RequestType: utils.MetaPostpaid,
		},
	}

	//disabled chargers
	expected := "ChargerS is disabled"
	if _, err := sessions.chargeEvent(cgrEv, true); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	cacheS := engine.NewCacheS(cfg, nil, nil)
	engine.Cache = cacheS
	connMgr = engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator): chanInternal})
	engine.SetConnManager(connMgr)

	sessions.cgrCfg.SessionSCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)}

	if usage, err := sessions.chargeEvent(cgrEv, false); err != nil {
		t.Error(err)
	} else if usage != 10*time.Nanosecond {
		t.Errorf("Expected %+v, received %+v", 10*time.Nanosecond, usage)
	}

	engine.Cache.Clear(nil)
	cgrEv.Tenant = "CHANGED_TENANT"
	cgrEv.Event[utils.RequestType] = utils.MetaPrepaid
	expected = "RALS_ERROR:MANDATORY_IE_MISSING: [connIDs]"
	if _, err := sessions.chargeEvent(cgrEv, false); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	//testing initSession with if it is message
	if _, err := sessions.initSession(cgrEv, utils.EmptyString, utils.EmptyString, time.Second,
		false, true); err != nil {
		t.Error(err)
	}

	engine.Cache = tmp
}

func TestUpdateSession(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	sessions := NewSessionS(cfg, dm, nil)

	updatedEv := map[string]interface{}{
		utils.Usage:       time.Second,
		utils.Destination: 10,
	}
	ss := &Session{
		EventStart: map[string]interface{}{},
		SRuns: []*SRun{
			{
				Event: map[string]interface{}{},
				CD: &engine.CallDescriptor{
					RunID: "RUNID_TEST",
				},
			},
		},
	}

	if _, err := sessions.updateSession(ss, updatedEv, nil, false); err != nil {
		t.Error(err)
	}

	updatedEv[utils.Usage] = "invalid_format"
	expectedErr := "time: invalid duration \"invalid_format\""
	if _, err := sessions.updateSession(ss, updatedEv, nil, false); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}

	delete(updatedEv, utils.Usage)
	ss.SRuns[0].Event[utils.RequestType] = utils.MetaNone
	if _, err := sessions.updateSession(ss, updatedEv, nil, false); err != nil {
		t.Error(err)
	}
}

func TestEndSession(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	engine.Cache.Clear(nil)
	sTestMock := &testMockClients{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ResponderDebit: func(args interface{}, reply interface{}) error {
				return nil
			},
			utils.ResponderRefundRounding: func(args interface{}, reply interface{}) error {
				return utils.ErrNotImplemented
			},
			utils.CDRsV1StoreSessionCost: func(args interface{}, reply interface{}) error {
				return utils.ErrNotImplemented
			},
		},
	}
	chanInternal := make(chan rpcclient.ClientConnector, 1)
	chanInternal <- sTestMock
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().RALsConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs)}
	cfg.SessionSCfg().CDRsConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs)}
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs): chanInternal,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs): chanInternal,
	})
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), connMgr)
	sessions := NewSessionS(cfg, dm, connMgr)

	ss := &Session{
		EventStart: map[string]interface{}{},
		SRuns: []*SRun{
			{
				TotalUsage: 50 * time.Minute,
				Event:      map[string]interface{}{},
				EventCost: &engine.EventCost{
					Usage:          utils.DurationPointer(30 * time.Minute),
					AccountSummary: &engine.AccountSummary{},
					Accounting: map[string]*engine.BalanceCharge{
						utils.MetaRounding: {
							AccountID: "ACCOUNTING_ID_TEST",
							RatingID:  "RATING_ID_TEST",
						},
						"ID_TEST2": {
							AccountID: "ACCOUNTING_ID_TEST2",
							RatingID:  utils.MetaRounding,
						},
					},
					Charges: []*engine.ChargingInterval{
						{
							RatingID: "RATING_ID",
							Increments: []*engine.ChargingIncrement{
								{
									Usage:        20 * time.Minute,
									Cost:         0.50,
									AccountingID: utils.MetaRounding,
								},
								{
									Usage:        15 * time.Minute,
									Cost:         0.466,
									AccountingID: "ID_TEST2",
								},
							},
						},
					},
					Rating: map[string]*engine.RatingUnit{
						"RATING_ID_TEST": {
							RatesID:          utils.EmptyString,
							TimingID:         utils.EmptyString,
							RoundingDecimals: 2,
						},
						utils.MetaRounding: {
							RatesID:  utils.EmptyString,
							TimingID: utils.EmptyString,
						},
						"RATING_ID": {
							RatingFiltersID:  utils.EmptyString,
							RoundingDecimals: 5,
						},
					},
				},
				CD: &engine.CallDescriptor{
					LoopIndex: 1,
				},
			},
		},
		chargeable: true,
	}

	activationTime := time.Date(2020, 21, 07, 10, 0, 0, 0, time.UTC)
	expected := "cannot find last active ChargingInterval"
	if err := sessions.endSession(ss, utils.DurationPointer(20*time.Minute), utils.DurationPointer(time.Second),
		utils.TimePointer(activationTime), false); err == nil || err.Error() != expected {
		t.Error(err)
	}

	engine.Cache.Clear(nil)
	//totalUsage will be empty
	sessions.cgrCfg.SessionSCfg().StoreSCosts = true
	if err := sessions.endSession(ss, nil, utils.DurationPointer(time.Hour),
		utils.TimePointer(activationTime), false); err != nil {
		t.Error(err)
	}
}

func TestCallBiRPC(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	sessions := NewSessionS(cfg, dm, nil)

	sTestMock := &testMockClients{}
	valid := "BiRPCv1TerminateSession"
	args := new(V1TerminateSessionArgs)
	var reply *string

	if err := sessions.CallBiRPC(sTestMock, valid, args, reply); err == nil || err != rpcclient.ErrUnsupporteServiceMethod {
		t.Errorf("Expected %+v, received %+v", rpcclient.ErrUnsupporteServiceMethod, err)
	}

	valid = "BiRPC.TerminateSession"
	if err := sessions.CallBiRPC(sTestMock, valid, args, reply); err == nil || err != rpcclient.ErrUnsupporteServiceMethod {
		t.Errorf("Expected %+v, received %+v", rpcclient.ErrUnsupporteServiceMethod, err)
	}

	valid = "BiRPCv1.TerminateSession"
	expected := "MANDATORY_IE_MISSING: [CGREvent]"
	if err := sessions.Call(valid, args, reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

}

func TestBiRPCv1GetActivePassiveSessions(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	clnt := &testMockClients{}

	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().SessionIndexes = utils.StringSet{
		"ToR": {},
	}
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	sessions := NewSessionS(cfg, dm, nil)

	var reply []*ExternalSession
	if err := sessions.BiRPCv1GetActiveSessions(clnt, nil, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}

	sEv := engine.NewMapEvent(map[string]interface{}{
		utils.EventName:       "TEST_EVENT",
		utils.ToR:             "*voice",
		utils.OriginID:        "12345",
		utils.AccountField:    "account1",
		utils.Subject:         "subject1",
		utils.Destination:     "+4986517174963",
		utils.Category:        "call",
		utils.Tenant:          "cgrates.org",
		utils.RequestType:     "*prepaid",
		utils.SetupTime:       "2015-11-09T14:21:24Z",
		utils.AnswerTime:      "2015-11-09T14:22:02Z",
		utils.Usage:           "1m23s",
		utils.LastUsed:        "21s",
		utils.PDD:             "300ms",
		utils.Route:           "supplier1",
		utils.DisconnectCause: "NORMAL_DISCONNECT",
		utils.OriginHost:      "127.0.0.1",
		"Extra1":              "Value1",
		"Extra2":              5,
		"Extra3":              "",
	})
	sr2 := sEv.Clone()
	// Index first session
	session := &Session{
		CGRID:      GetSetCGRID(sEv),
		EventStart: sEv,
		SRuns: []*SRun{
			{
				Event: sEv,
				CD: &engine.CallDescriptor{
					RunID: "RunID",
				},
			},
			{
				Event: sr2,
				CD: &engine.CallDescriptor{
					RunID: "RunID2",
				},
			},
		},
		chargeable: true,
	}
	sr2[utils.ToR] = utils.MetaSMS
	sr2[utils.Subject] = "subject2"
	sr2[utils.CGRID] = GetSetCGRID(sEv)
	sessions.registerSession(session, false)

	st, err := utils.IfaceAsTime("2015-11-09T14:21:24Z", "")
	if err != nil {
		t.Fatal(err)
	}
	at, err := utils.IfaceAsTime("2015-11-09T14:22:02Z", "")
	if err != nil {
		t.Fatal(err)
	}
	eses1 := &ExternalSession{
		CGRID:       "cade401f46f046311ed7f62df3dfbb84adb98aad",
		ToR:         "*voice",
		OriginID:    "12345",
		OriginHost:  "127.0.0.1",
		Source:      "SessionS_TEST_EVENT",
		RequestType: "*prepaid",
		Category:    "call",
		Account:     "account1",
		Subject:     "subject1",
		Destination: "+4986517174963",
		SetupTime:   st,
		AnswerTime:  at,
		ExtraFields: map[string]string{
			"DisconnectCause": "NORMAL_DISCONNECT",
			"EventName":       "TEST_EVENT",
			"Extra1":          "Value1",
			"Extra2":          "5",
			"Extra3":          "",
			"LastUsed":        "21s",
			"PDD":             "300ms",
			utils.Route:       "supplier1",
		},
		NodeID: sessions.cgrCfg.GeneralCfg().NodeID,
	}

	expSess := []*ExternalSession{
		eses1,
	}

	args := &utils.SessionFilter{Filters: []string{fmt.Sprintf("*string:~*req.ToR:%s|%s", utils.MetaVoice, utils.MetaData)}}
	if err := sessions.BiRPCv1GetActiveSessions(clnt, args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expSess, reply) {
		t.Errorf("Expected %s , received: %s", utils.ToJSON(expSess), utils.ToJSON(reply))
	}

	var newReply1 int
	//nil args, but it will be an empty SessionFilter
	if err := sessions.BiRPCv1GetActiveSessionsCount(clnt, nil, &newReply1); err != nil {
		t.Error(err)
	} else if newReply1 != 2 {
		t.Errorf("Expected %+v, received: %+v", 2, newReply1)
	}

	if err := sessions.BiRPCv1GetActiveSessionsCount(clnt, args, &newReply1); err != nil {
		t.Error(err)
	} else if newReply1 != 1 {
		t.Errorf("Expected %+v, received: %+v", 1, newReply1)
	}

	//Passive session
	reply = []*ExternalSession{}
	if err := sessions.BiRPCv1GetPassiveSessions(clnt, nil, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}

	sessions.pSessions = map[string]*Session{
		utils.EmptyString: {
			SRuns: []*SRun{
				{
					EventCost: &engine.EventCost{},
				},
			},
		},
	}
	//empty filters
	sessions.cgrCfg.GeneralCfg().NodeID = "TEST_ID"
	args = &utils.SessionFilter{}
	expSess = []*ExternalSession{
		{
			Source:      "SessionS_",
			NodeID:      "TEST_ID",
			ExtraFields: map[string]string{},
		},
	}
	if err := sessions.BiRPCv1GetPassiveSessions(clnt, args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expSess) {
		t.Errorf("Expected %+v\n, received: %+v", utils.ToJSON(expSess), utils.ToJSON(reply))
	}

	if err := sessions.BiRPCv1GetPassiveSessionsCount(clnt, nil, &newReply1); err != nil {
		t.Error(err)
	}
}

func TestBiRPCv1SetPassiveSession(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	clnt := &testMockClients{}

	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	sessions := NewSessionS(cfg, dm, nil)

	var reply string
	ss := &Session{
		Tenant: "cgrates.org",
		SRuns:  []*SRun{},
	}
	expected := "MANDATORY_IE_MISSING: [CGRID]"
	if err := sessions.BiRPCv1SetPassiveSession(clnt, ss, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v\n, received: %+v", expected, err)
	} else if reply != utils.EmptyString {
		t.Errorf("Expected %+v\n, received: %+v", utils.EmptyString, err)
	}

	ss.CGRID = "CGR_ID"
	if err := sessions.BiRPCv1SetPassiveSession(clnt, ss, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v\n, received: %+v", utils.ErrNotFound, err)
	} else if reply != utils.EmptyString {
		t.Errorf("Expected %+v\n, received: %+v", utils.EmptyString, err)
	}

	sessions.pSessions = map[string]*Session{
		"CGR_ID": ss,
	}
	if err := sessions.BiRPCv1SetPassiveSession(clnt, ss, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected %+v\n, received: %+v", utils.OK, err)
	}

	sessions.aSessions = map[string]*Session{
		"CGR_ID": ss,
	}
	ss.EventStart = engine.MapEvent{}
	if err := sessions.BiRPCv1SetPassiveSession(clnt, ss, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected %+v\n, received: %+v", utils.OK, err)
	}
}

func TestBiRPCv1ReplicateSessions(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	engine.Cache.Clear(nil)
	clnt := &testMockClients{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.SessionSv1SetPassiveSession: func(args interface{}, reply interface{}) error {
				return utils.ErrNotImplemented
			},
		},
	}
	chanInternal := make(chan rpcclient.ClientConnector, 1)
	chanInternal <- clnt
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		"conn1": chanInternal,
	})
	dm := engine.NewDataManager(data, cfg.CacheCfg(), connMgr)
	sessions := NewSessionS(cfg, dm, connMgr)

	args := ArgsReplicateSessions{
		CGRID:   "CGRID_TEST",
		Passive: false,
		ConnIDs: []string{},
	}

	var reply string
	if err := sessions.BiRPCv1ReplicateSessions(clnt, args, &reply); err != nil {
		t.Error(err)
	}

	args.ConnIDs = []string{"conn1"}
	if err := sessions.BiRPCv1ReplicateSessions(clnt, args, &reply); err != nil {
		t.Error(err)
	}
}

func TestBiRPCv1AuthorizeEvent(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	tmp := engine.Cache

	engine.Cache.Clear(nil)
	clnt := &testMockClients{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.AttributeSv1ProcessEvent: func(args interface{}, reply interface{}) error {
				cgrEv := engine.AttrSProcessEventReply{
					CGREvent: &utils.CGREvent{
						ID:     "TestID",
						Tenant: "cgrates.org",
						Event:  map[string]interface{}{},
					},
				}
				*reply.(*engine.AttrSProcessEventReply) = cgrEv
				return nil
			},
		},
	}
	chanInternal := make(chan rpcclient.ClientConnector, 1)
	chanInternal <- clnt
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes): chanInternal,
	})
	dm := engine.NewDataManager(data, cfg.CacheCfg(), connMgr)
	sessions := NewSessionS(cfg, dm, connMgr)

	cgrEvent := &utils.CGREvent{
		ID: "TestID",
		Event: map[string]interface{}{
			utils.Usage: "10s",
		},
	}
	args := NewV1AuthorizeArgs(false, []string{},
		false, []string{}, false, []string{}, false, false,
		false, false, false, nil, utils.Paginator{}, false, "")

	rply := &V1AuthorizeReply{
		Attributes:   &engine.AttrSProcessEventReply{},
		Routes:       engine.SortedRoutesList{},
		StatQueueIDs: &[]string{},
		ThresholdIDs: &[]string{},
	}

	expected := "MANDATORY_IE_MISSING: [CGREvent]"
	if err := sessions.BiRPCv1AuthorizeEvent(nil, args, rply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	args.CGREvent = cgrEvent
	//RPC caching
	sessions.cgrCfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = -1
	expected = "MANDATORY_IE_MISSING: [subsystems]"

	caches := engine.NewCacheS(cfg, dm, nil)
	value := &utils.CachedRPCResponse{
		Result: &V1AuthorizeReply{
			ResourceAllocation: utils.StringPointer("ROUTE_LEASTCOST_1"),
		},
	}
	engine.Cache = caches
	caches.SetWithoutReplicate(utils.CacheRPCResponses, utils.ConcatenatedKey(utils.SessionSv1AuthorizeEvent, args.CGREvent.ID),
		value, nil, true, utils.NonTransactional)
	if err := sessions.BiRPCv1AuthorizeEvent(nil, args, rply); err != nil {
		t.Error(err)
	}
	engine.Cache = tmp

	if err := sessions.BiRPCv1AuthorizeEvent(nil, args, rply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	//Get Attributes
	sessions.cgrCfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 0
	args = NewV1AuthorizeArgs(true, []string{},
		false, []string{}, false, []string{}, false, false,
		true, false, false, cgrEvent, utils.Paginator{}, false, "")
	args.CGREvent.ID = "TestID"
	args.CGREvent.Tenant = "cgrates.org"
	expected = "ATTRIBUTES_ERROR:NOT_CONNECTED: AttributeS"

	if err := sessions.BiRPCv1AuthorizeEvent(nil, args, rply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	expected = "NOT_CONNECTED: RouteS"
	sessions.cgrCfg.SessionSCfg().AttrSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}
	if err := sessions.BiRPCv1AuthorizeEvent(nil, args, rply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestBiRPCv1AuthorizeEvent2(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	engine.Cache.Clear(nil)
	clnt := &testMockClients{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ChargerSv1ProcessEvent: func(args interface{}, reply interface{}) error {
				cghrgs := []*engine.ChrgSProcessEventReply{
					{
						CGREvent: &utils.CGREvent{
							Tenant: "cgrates.org",
							ID:     "TestID",
							Event: map[string]interface{}{
								utils.Usage: "10s",
							},
						},
					},
				}
				*reply.(*[]*engine.ChrgSProcessEventReply) = cghrgs
				return nil
			},
			utils.ResourceSv1AuthorizeResources: func(args interface{}, reply interface{}) error {
				if args.(*utils.ArgRSv1ResourceUsage).Tenant == "new_tenant" {
					return utils.ErrNotImplemented
				}
				return nil
			},
			utils.RouteSv1GetRoutes: func(args interface{}, reply interface{}) error {
				*reply.(*engine.SortedRoutesList) = engine.SortedRoutesList{{
					Routes: []*engine.SortedRoute{
						{
							RouteID: "RouteID",
						},
					},
				}}
				return nil
			},
			utils.ThresholdSv1ProcessEvent: func(args interface{}, reply interface{}) error {
				return utils.ErrNotImplemented
			},
		},
	}
	chanInternal := make(chan rpcclient.ClientConnector, 1)
	chanInternal <- clnt
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 0
	data := engine.NewInternalDB(nil, nil, true)
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers):   chanInternal,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources):  chanInternal,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRoutes):     chanInternal,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds): chanInternal,
	})
	dm := engine.NewDataManager(data, cfg.CacheCfg(), connMgr)
	sessions := NewSessionS(cfg, dm, connMgr)

	cgrEvent := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			utils.Usage: "10s",
		},
	}
	args := NewV1AuthorizeArgs(false, []string{},
		false, []string{}, false, []string{}, false, true,
		false, false, false, cgrEvent, utils.Paginator{}, false, "")

	rply := &V1AuthorizeReply{
		Attributes:   &engine.AttrSProcessEventReply{},
		Routes:       engine.SortedRoutesList{},
		StatQueueIDs: &[]string{},
		ThresholdIDs: &[]string{},
	}

	//GetMaxUsage
	expected := "ChargerS is disabled"
	if err := sessions.BiRPCv1AuthorizeEvent(nil, args, rply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	args.CGREvent.ID = "TestID"
	sessions.cgrCfg.SessionSCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)}
	if err := sessions.BiRPCv1AuthorizeEvent(nil, args, rply); err != nil {
		t.Error(err)
	}

	//AuthorizeResources
	args = NewV1AuthorizeArgs(false, []string{},
		false, []string{}, false, []string{}, true, false,
		false, false, false, cgrEvent, utils.Paginator{}, false, "")
	expected = "NOT_CONNECTED: ResourceS"
	if err := sessions.BiRPCv1AuthorizeEvent(nil, args, rply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	sessions.cgrCfg.SessionSCfg().ResSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)}
	if err := sessions.BiRPCv1AuthorizeEvent(nil, args, rply); err != nil {
		t.Error(err)
	}

	args.CGREvent.Tenant = "new_tenant"
	expected = "RESOURCES_ERROR:NOT_IMPLEMENTED"
	if err := sessions.BiRPCv1AuthorizeEvent(nil, args, rply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	//GetRoutes
	args = NewV1AuthorizeArgs(false, []string{},
		false, []string{}, false, []string{}, false, false,
		true, false, false, cgrEvent, utils.Paginator{}, false, "")
	expected = "NOT_CONNECTED: RouteS"
	if err := sessions.BiRPCv1AuthorizeEvent(nil, args, rply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	sessions.cgrCfg.SessionSCfg().RouteSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRoutes)}
	if err := sessions.BiRPCv1AuthorizeEvent(nil, args, rply); err != nil {
		t.Error(err)
	}

	//ProcessThresholds
	args = NewV1AuthorizeArgs(false, []string{},
		true, []string{"TestID"}, false, []string{}, false, false,
		true, false, false, cgrEvent, utils.Paginator{}, false, "")
	if err := sessions.BiRPCv1AuthorizeEvent(nil, args, rply); err == nil || err != utils.ErrPartiallyExecuted {
		t.Errorf("Expected %+v, received %+v", utils.ErrPartiallyExecuted, err)
	}

	//ProcessStats
	args = NewV1AuthorizeArgs(false, []string{},
		false, []string{}, true, []string{"TestID"}, false, false,
		true, false, false, cgrEvent, utils.Paginator{}, false, "")
	if err := sessions.BiRPCv1AuthorizeEvent(nil, args, rply); err == nil || err != utils.ErrPartiallyExecuted {
		t.Errorf("Expected %+v, received %+v", utils.ErrPartiallyExecuted, err)
	}
}

func TestBiRPCv1AuthorizeEventWithDigest(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	engine.Cache.Clear(nil)
	clnt := &testMockClients{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.AttributeSv1ProcessEvent: func(args interface{}, reply interface{}) error {
				cgrEv := engine.AttrSProcessEventReply{
					CGREvent: &utils.CGREvent{
						ID:     "TestID",
						Tenant: "cgrates.org",
						Event:  map[string]interface{}{},
					},
				}
				*reply.(*engine.AttrSProcessEventReply) = cgrEv
				return nil
			},
			utils.ChargerSv1ProcessEvent: func(args interface{}, reply interface{}) error {
				cghrgs := []*engine.ChrgSProcessEventReply{
					{
						CGREvent: &utils.CGREvent{
							Tenant: "cgrates.org",
							ID:     "TestID",
							Event: map[string]interface{}{
								utils.Usage: "10s",
							},
						},
					},
				}
				*reply.(*[]*engine.ChrgSProcessEventReply) = cghrgs
				return nil
			},
			utils.ResourceSv1AuthorizeResources: func(args interface{}, reply interface{}) error {
				if args.(*utils.ArgRSv1ResourceUsage).Tenant == "new_tenant" {
					return utils.ErrNotImplemented
				}
				return nil
			},
			utils.RouteSv1GetRoutes: func(args interface{}, reply interface{}) error {
				*reply.(*engine.SortedRoutesList) = engine.SortedRoutesList{{
					Routes: []*engine.SortedRoute{
						{
							RouteID: "RouteID",
						},
					},
				}}
				return nil
			},
			utils.ThresholdSv1ProcessEvent: func(args interface{}, reply interface{}) error {
				return nil
			},
			utils.StatSv1ProcessEvent: func(args interface{}, reply interface{}) error {
				return nil
			},
		},
	}
	chanInternal := make(chan rpcclient.ClientConnector, 1)
	chanInternal <- clnt
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().AttrSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}
	cfg.SessionSCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)}
	cfg.SessionSCfg().ResSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)}
	cfg.SessionSCfg().RouteSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRoutes)}
	cfg.SessionSCfg().ThreshSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)}
	cfg.SessionSCfg().StatSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)}
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 0
	data := engine.NewInternalDB(nil, nil, true)
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers):   chanInternal,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources):  chanInternal,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRoutes):     chanInternal,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds): chanInternal,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes): chanInternal,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats):      chanInternal,
	})
	dm := engine.NewDataManager(data, cfg.CacheCfg(), connMgr)
	sessions := NewSessionS(cfg, dm, connMgr)

	cgrEvent := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			utils.Usage: "10s",
		},
	}
	args := NewV1AuthorizeArgs(true, []string{},
		true, []string{}, true, []string{}, true, true,
		true, false, false, cgrEvent, utils.Paginator{}, false, "")

	authReply := new(V1AuthorizeReplyWithDigest)
	expectedRply := &V1AuthorizeReplyWithDigest{
		AttributesDigest:   utils.StringPointer(utils.EmptyString),
		ResourceAllocation: utils.StringPointer(utils.EmptyString),
		RoutesDigest:       utils.StringPointer("RouteID"),
		MaxUsage:           10800,
		Thresholds:         utils.StringPointer(utils.EmptyString),
		StatQueues:         utils.StringPointer(utils.EmptyString),
	}
	if err := sessions.BiRPCv1AuthorizeEventWithDigest(nil, args, authReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(authReply, expectedRply) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expectedRply), utils.ToJSON(authReply))
	}

	sessions.cgrCfg.SessionSCfg().ChargerSConns = nil
	expected := "ChargerS is disabled"
	if err := sessions.BiRPCv1AuthorizeEventWithDigest(nil, args, authReply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestBiRPCv1InitiateSession1(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	tmp := engine.Cache

	engine.Cache.Clear(nil)
	clnt := &testMockClients{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ChargerSv1ProcessEvent: func(args interface{}, reply interface{}) error {
				cghrgs := []*engine.ChrgSProcessEventReply{
					{
						CGREvent: &utils.CGREvent{
							Tenant: "cgrates.org",
							ID:     "TestID",
							Event: map[string]interface{}{
								utils.Usage: "10s",
							},
						},
					},
				}
				*reply.(*[]*engine.ChrgSProcessEventReply) = cghrgs
				return nil
			},
			utils.ResourceSv1AuthorizeResources: func(args interface{}, reply interface{}) error {
				if args.(*utils.ArgRSv1ResourceUsage).Tenant == "new_tenant" {
					return utils.ErrNotImplemented
				}
				return nil
			},
			utils.ResourceSv1AllocateResources: func(args interface{}, reply interface{}) error {
				if args.(*utils.ArgRSv1ResourceUsage).UsageID == "ORIGIN_ID" {
					return utils.ErrNotImplemented
				}
				return nil
			},
			utils.AttributeSv1ProcessEvent: func(args interface{}, reply interface{}) error {
				if len(args.(*engine.AttrArgsProcessEvent).AttributeIDs) != 0 {
					return utils.ErrNotImplemented
				}
				return nil
			},
		},
	}
	chanInternal := make(chan rpcclient.ClientConnector, 1)
	chanInternal <- clnt
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	cfg.SessionSCfg().AttrSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}
	cfg.SessionSCfg().ResSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)}
	cfg.SessionSCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)}
	data := engine.NewInternalDB(nil, nil, true)
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers):   chanInternal,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources):  chanInternal,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes): chanInternal,
	})
	dm := engine.NewDataManager(data, cfg.CacheCfg(), connMgr)
	sessions := NewSessionS(cfg, dm, connMgr)

	cgrEvent := &utils.CGREvent{
		Event: map[string]interface{}{
			utils.Usage:    "10s",
			utils.OriginID: "TEST_ID",
		},
	}
	args := NewV1InitSessionArgs(true, []string{},
		false, []string{}, false, []string{}, true, false,
		nil, true)

	rply := &V1InitSessionReply{}
	expected := "MANDATORY_IE_MISSING: [CGREvent]"
	if err := sessions.BiRPCv1InitiateSession(nil, args, rply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	args.CGREvent = cgrEvent
	if err := sessions.BiRPCv1InitiateSession(nil, args, rply); err != nil {
		t.Error(err)
	}

	//get from cache error
	cgrEvent.ID = "INITIATE_SESSION_ACTIVE"
	args = NewV1InitSessionArgs(true, []string{},
		false, []string{}, false, []string{}, true, false,
		cgrEvent, true)
	caches := engine.NewCacheS(cfg, dm, nil)
	//value's error will be nil, so the error of the initiate sessions will be the same
	value := &utils.CachedRPCResponse{
		Result: &V1InitSessionReply{
			ResourceAllocation: utils.StringPointer("ROUTE_LEASTCOST_1"),
		},
	}
	engine.Cache = caches
	engine.Cache.SetWithoutReplicate(utils.CacheRPCResponses, utils.ConcatenatedKey(utils.SessionSv1InitiateSession, args.CGREvent.ID),
		value, nil, true, utils.NonTransactional)
	if err := sessions.BiRPCv1InitiateSession(nil, args, rply); err != nil {
		t.Error(err)
	}
	engine.Cache = tmp

	args.CGREvent.Tenant = utils.EmptyString
	args.AttributeIDs = []string{"attr1"}
	expected = "ATTRIBUTES_ERROR:NOT_IMPLEMENTED"
	if err := sessions.BiRPCv1InitiateSession(nil, args, rply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	args.AllocateResources = true
	args.AttributeIDs = []string{}
	sessions.cgrCfg.SessionSCfg().ResSConns = []string{}
	expected = "NOT_CONNECTED: ResourceS"
	if err := sessions.BiRPCv1InitiateSession(nil, args, rply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	sessions.cgrCfg.SessionSCfg().ResSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)}

	args = NewV1InitSessionArgs(true, []string{},
		false, []string{}, false, []string{}, true, false,
		cgrEvent, true)
	delete(args.CGREvent.Event, utils.OriginID)
	expected = "MANDATORY_IE_MISSING: [OriginID]"
	if err := sessions.BiRPCv1InitiateSession(nil, args, rply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	cgrEvent = &utils.CGREvent{
		ID:     "Test_id",
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			utils.Usage:    "10s",
			utils.OriginID: "ORIGIN_ID",
		},
	}
	args = NewV1InitSessionArgs(true, []string{},
		false, []string{}, false, []string{}, true, false,
		cgrEvent, true)
	expected = "RESOURCES_ERROR:NOT_IMPLEMENTED"
	if err := sessions.BiRPCv1InitiateSession(nil, args, rply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	//missing subsystems
	args = NewV1InitSessionArgs(false, []string{},
		false, []string{}, false, []string{}, false, false,
		cgrEvent, true)
	expected = "MANDATORY_IE_MISSING: [subsystems]"
	if err := sessions.BiRPCv1InitiateSession(nil, args, rply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestBiRPCv1InitiateSession2(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	engine.Cache.Clear(nil)
	clnt := &testMockClients{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ChargerSv1ProcessEvent: func(args interface{}, reply interface{}) error {
				cghrgs := []*engine.ChrgSProcessEventReply{
					{
						CGREvent: &utils.CGREvent{
							Tenant: "cgrates.org",
							ID:     "TestID",
							Event: map[string]interface{}{
								utils.Usage: "10s",
							},
						},
					},
				}
				if args.(*utils.CGREvent).ID == "PREPAID" {
					cghrgs[0].CGREvent.Event[utils.RequestType] = utils.MetaPrepaid
				}
				*reply.(*[]*engine.ChrgSProcessEventReply) = cghrgs
				return nil
			},
			utils.ResourceSv1AuthorizeResources: func(args interface{}, reply interface{}) error {
				if args.(*utils.ArgRSv1ResourceUsage).Tenant == "new_tenant" {
					return utils.ErrNotImplemented
				}
				return nil
			},
			utils.ThresholdSv1ProcessEvent: func(args interface{}, reply interface{}) error {
				return utils.ErrNotImplemented
			},
		},
	}
	chanInternal := make(chan rpcclient.ClientConnector, 1)
	chanInternal <- clnt
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	cfg.SessionSCfg().ThreshSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)}
	cfg.SessionSCfg().ResSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)}
	data := engine.NewInternalDB(nil, nil, true)
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers):   chanInternal,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources):  chanInternal,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds): chanInternal,
	})
	dm := engine.NewDataManager(data, cfg.CacheCfg(), connMgr)
	sessions := NewSessionS(cfg, dm, connMgr)

	cgrEvent := &utils.CGREvent{
		ID:     "Test_id",
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			utils.Usage: "invalid_usage",
		},
		Opts: map[string]interface{}{
			utils.OptsDebitInterval: "invalid_DUR_FORMAT",
		},
	}

	args := NewV1InitSessionArgs(false, []string{},
		false, []string{}, false, []string{}, false, true,
		cgrEvent, true)

	rply := &V1InitSessionReply{}
	expected := "RALS_ERROR:time: invalid duration \"invalid_DUR_FORMAT\""
	if err := sessions.BiRPCv1InitiateSession(nil, args, rply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	args.CGREvent.Opts[utils.OptsDebitInterval] = "10s"

	expected = "ChargerS is disabled"
	if err := sessions.BiRPCv1InitiateSession(nil, args, rply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	sessions.cgrCfg.SessionSCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)}

	expected = "RALS_ERROR:time: invalid duration \"invalid_usage\""
	if err := sessions.BiRPCv1InitiateSession(nil, args, rply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	sessions = NewSessionS(cfg, dm, connMgr)
	args.CGREvent.Event[utils.Usage] = "10s"
	if err := sessions.BiRPCv1InitiateSession(nil, args, rply); err != nil {
		t.Error(err)
	}

	//here we process the thresholds
	args = NewV1InitSessionArgs(false, []string{},
		true, []string{}, true, []string{}, false, true,
		cgrEvent, true)
	sessions = NewSessionS(cfg, dm, connMgr)
	if err := sessions.BiRPCv1InitiateSession(nil, args, rply); err == nil || err != utils.ErrPartiallyExecuted {
		t.Errorf("Expected %+v, received %+v", utils.ErrPartiallyExecuted, err)
	}

	//is prepaid
	cgrEvent = &utils.CGREvent{
		ID: "PREPAID",
		Event: map[string]interface{}{
			utils.Usage: "1s",
		},
		Opts: map[string]interface{}{
			utils.OptsDebitInterval: "10s",
		},
	}
	sessions = NewSessionS(cfg, dm, connMgr)
	args = NewV1InitSessionArgs(false, []string{},
		true, []string{}, true, []string{}, false, true,
		cgrEvent, true)
	expected = "EXISTS"
	if err := sessions.BiRPCv1InitiateSession(nil, args, rply); err == nil || err != utils.ErrPartiallyExecuted {
		t.Errorf("Expected %+v, received %+v", utils.ErrPartiallyExecuted, err)
	}

}
func TestBiRPCv1InitiateSessionWithDigest(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	engine.Cache.Clear(nil)
	clnt := &testMockClients{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.AttributeSv1ProcessEvent: func(args interface{}, reply interface{}) error {
				cgrEv := engine.AttrSProcessEventReply{
					CGREvent: &utils.CGREvent{
						ID:     "TestID",
						Tenant: "cgrates.org",
						Event:  map[string]interface{}{},
					},
				}
				*reply.(*engine.AttrSProcessEventReply) = cgrEv
				return nil
			},
			utils.ChargerSv1ProcessEvent: func(args interface{}, reply interface{}) error {
				cghrgs := []*engine.ChrgSProcessEventReply{
					{
						CGREvent: &utils.CGREvent{
							Tenant: "cgrates.org",
							ID:     "TestID",
							Event: map[string]interface{}{
								utils.Usage: "10s",
							},
						},
					},
				}
				*reply.(*[]*engine.ChrgSProcessEventReply) = cghrgs
				return nil
			},
			utils.ResourceSv1AllocateResources: func(args interface{}, reply interface{}) error {
				return nil
			},
			utils.RouteSv1GetRoutes: func(args interface{}, reply interface{}) error {
				*reply.(*engine.SortedRoutesList) = engine.SortedRoutesList{{
					Routes: []*engine.SortedRoute{
						{
							RouteID: "RouteID",
						},
					},
				}}
				return nil
			},
			utils.ThresholdSv1ProcessEvent: func(args interface{}, reply interface{}) error {
				return nil
			},
			utils.StatSv1ProcessEvent: func(args interface{}, reply interface{}) error {
				return nil
			},
		},
	}
	chanInternal := make(chan rpcclient.ClientConnector, 1)
	chanInternal <- clnt
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().AttrSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}
	cfg.SessionSCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)}
	cfg.SessionSCfg().ResSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)}
	cfg.SessionSCfg().RouteSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRoutes)}
	cfg.SessionSCfg().ThreshSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)}
	cfg.SessionSCfg().StatSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)}
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 0
	data := engine.NewInternalDB(nil, nil, true)
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers):   chanInternal,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources):  chanInternal,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRoutes):     chanInternal,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds): chanInternal,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes): chanInternal,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats):      chanInternal,
	})
	dm := engine.NewDataManager(data, cfg.CacheCfg(), connMgr)
	sessions := NewSessionS(cfg, dm, connMgr)

	cgrEvent := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			utils.Usage:    "10s",
			utils.OriginID: "ORIGIND_ID",
		},
	}

	args := NewV1InitSessionArgs(true, []string{},
		true, []string{}, true, []string{}, true, true,
		cgrEvent, true)

	authReply := new(V1InitReplyWithDigest)
	expectedRply := &V1InitReplyWithDigest{
		AttributesDigest:   utils.StringPointer(utils.EmptyString),
		ResourceAllocation: utils.StringPointer(utils.EmptyString),
		MaxUsage:           10800,
		Thresholds:         utils.StringPointer(utils.EmptyString),
		StatQueues:         utils.StringPointer(utils.EmptyString),
	}
	if err := sessions.BiRPCv1InitiateSessionWithDigest(nil, args, authReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(authReply, expectedRply) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expectedRply), utils.ToJSON(authReply))
	}

	sessions.cgrCfg.SessionSCfg().ChargerSConns = nil
	expected := "MANDATORY_IE_MISSING: [OriginID]"
	if err := sessions.BiRPCv1InitiateSessionWithDigest(nil, args, authReply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestBiRPCv1UpdateSession1(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	tmp := engine.Cache

	engine.Cache.Clear(nil)
	clnt := &testMockClients{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.AttributeSv1ProcessEvent: func(args interface{}, reply interface{}) error {
				if len(args.(*engine.AttrArgsProcessEvent).AttributeIDs) == 1 {
					return utils.ErrNotImplemented
				}
				return nil
			},
		},
	}
	chanInternal := make(chan rpcclient.ClientConnector, 1)
	chanInternal <- clnt
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	cfg.SessionSCfg().AttrSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}
	data := engine.NewInternalDB(nil, nil, true)
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes): chanInternal,
	})
	dm := engine.NewDataManager(data, cfg.CacheCfg(), connMgr)
	sessions := NewSessionS(cfg, dm, connMgr)

	cgrEvent := &utils.CGREvent{
		Event: map[string]interface{}{
			utils.Usage:    "10s",
			utils.OriginID: "TEST_ID",
		},
	}
	args := NewV1UpdateSessionArgs(true, []string{}, false,
		nil, true)
	rply := new(V1UpdateSessionReply)

	expected := "MANDATORY_IE_MISSING: [CGREvent]"
	if err := sessions.BiRPCv1UpdateSession(nil, args, rply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}

	caches := engine.NewCacheS(cfg, dm, nil)
	//value's error will be nil, so the error of the initiate sessions will be the same
	value := &utils.CachedRPCResponse{
		Result: &V1UpdateSessionReply{
			MaxUsage: utils.DurationPointer(time.Minute),
		},
	}
	cgrEvent.ID = "test_id"
	args = NewV1UpdateSessionArgs(true, []string{}, false,
		cgrEvent, true)
	engine.Cache = caches
	engine.Cache.SetWithoutReplicate(utils.CacheRPCResponses, utils.ConcatenatedKey(utils.SessionSv1UpdateSession, args.CGREvent.ID),
		value, nil, true, utils.NonTransactional)
	if err := sessions.BiRPCv1UpdateSession(nil, args, rply); err != nil {
		t.Error(err)
	}
	engine.Cache = tmp

	cgrEvent.ID = utils.EmptyString
	args = NewV1UpdateSessionArgs(true, []string{"attrr1"}, false,
		cgrEvent, true)
	expected = "ATTRIBUTES_ERROR:NOT_IMPLEMENTED"
	if err := sessions.BiRPCv1UpdateSession(nil, args, rply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}

	args = NewV1UpdateSessionArgs(true, []string{}, false,
		cgrEvent, true)
	if err := sessions.BiRPCv1UpdateSession(nil, args, rply); err != nil {
		t.Error(err)
	}

	args = NewV1UpdateSessionArgs(false, []string{}, false,
		cgrEvent, true)
	expected = "MANDATORY_IE_MISSING: [subsystems]"
	if err := sessions.BiRPCv1UpdateSession(nil, args, rply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}
}

func TestBiRPCv1UpdateSession2(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	engine.Cache.Clear(nil)
	clnt := &testMockClients{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ChargerSv1ProcessEvent: func(args interface{}, reply interface{}) error {
				cghrgs := []*engine.ChrgSProcessEventReply{
					{
						CGREvent: &utils.CGREvent{
							Tenant: "cgrates.org",
							ID:     "TestID",
							Event: map[string]interface{}{
								utils.Usage: "10s",
							},
						},
					},
				}
				*reply.(*[]*engine.ChrgSProcessEventReply) = cghrgs
				return nil
			},
		},
	}
	chanInternal := make(chan rpcclient.ClientConnector, 1)
	chanInternal <- clnt
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers): chanInternal,
	})
	dm := engine.NewDataManager(data, cfg.CacheCfg(), connMgr)
	sessions := NewSessionS(cfg, dm, connMgr)

	cgrEvent := &utils.CGREvent{
		ID: "test_id",
		Event: map[string]interface{}{
			utils.Usage:    "invalid_dur_format",
			utils.OriginID: "TEST_ID",
		},
		Opts: map[string]interface{}{
			utils.OptsDebitInterval: "invalid_dur_format",
		},
	}
	args := NewV1UpdateSessionArgs(false, []string{}, true,
		cgrEvent, true)
	rply := new(V1UpdateSessionReply)
	expected := "RALS_ERROR:time: invalid duration \"invalid_dur_format\""
	if err := sessions.BiRPCv1UpdateSession(nil, args, rply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}
	cgrEvent.Opts[utils.OptsDebitInterval] = "10s"

	args = NewV1UpdateSessionArgs(false, []string{}, true,
		cgrEvent, true)
	expected = "ChargerS is disabled"
	if err := sessions.BiRPCv1UpdateSession(nil, args, rply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}
	sessions.cgrCfg.SessionSCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)}

	expected = "RALS_ERROR:time: invalid duration \"invalid_dur_format\""
	if err := sessions.BiRPCv1UpdateSession(nil, args, rply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}

	cgrEvent.Event[utils.Usage] = time.Minute
	args = NewV1UpdateSessionArgs(false, []string{}, true,
		cgrEvent, true)
	if err := sessions.BiRPCv1UpdateSession(nil, args, rply); err != nil {
		t.Error(err)
	}
}

func TestBiRPCv1TerminateSession1(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	tmp := engine.Cache

	engine.Cache.Clear(nil)
	clnt := &testMockClients{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ChargerSv1ProcessEvent: func(args interface{}, reply interface{}) error {
				cghrgs := []*engine.ChrgSProcessEventReply{
					{
						CGREvent: &utils.CGREvent{
							Tenant: "cgrates.org",
							ID:     "TestID",
							Event: map[string]interface{}{
								utils.Usage: "10s",
								utils.CGRID: "TEST_ID",
							},
						},
					},
				}
				*reply.(*[]*engine.ChrgSProcessEventReply) = cghrgs
				return nil
			},
			utils.CacheSv1ReplicateSet: func(args interface{}, reply interface{}) error {
				return utils.ErrNotImplemented
			},
		},
	}
	chanInternal := make(chan rpcclient.ClientConnector, 1)
	chanInternal <- clnt
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	cfg.SessionSCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)}
	data := engine.NewInternalDB(nil, nil, true)
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers): chanInternal,
	})
	dm := engine.NewDataManager(data, cfg.CacheCfg(), connMgr)
	sessions := NewSessionS(cfg, dm, connMgr)

	cgrEvent := &utils.CGREvent{
		ID: "test_id",
		Event: map[string]interface{}{
			utils.Usage:    "10s",
			utils.OriginID: "TEST_ID",
		},
	}

	args := NewV1TerminateSessionArgs(true, false, false, nil, false, nil, nil, true)
	var reply string
	expected := "MANDATORY_IE_MISSING: [CGREvent]"
	if err := sessions.BiRPCv1TerminateSession(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}

	cgrEvent.ID = utils.EmptyString
	args = NewV1TerminateSessionArgs(false, false, false, nil, false, nil, cgrEvent, true)
	expected = "MANDATORY_IE_MISSING: [subsystems]"
	if err := sessions.BiRPCv1TerminateSession(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}
	cgrEvent.ID = "test_id"

	caches := engine.NewCacheS(cfg, dm, nil)
	//value's error will be nil, so the error of the initiate sessions will be the same
	value := &utils.CachedRPCResponse{
		Result: utils.StringPointer("ROUTE_LEASTCOST_1"),
	}
	engine.Cache = caches
	engine.Cache.SetWithoutReplicate(utils.CacheRPCResponses, utils.ConcatenatedKey(utils.SessionSv1TerminateSession, args.CGREvent.ID),
		value, nil, true, utils.NonTransactional)
	if err := sessions.BiRPCv1TerminateSession(nil, args, &reply); err != nil {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}
	engine.Cache = tmp

	cgrEvent.Event[utils.OriginID] = utils.EmptyString
	args = NewV1TerminateSessionArgs(true, false, false, nil, false, nil, cgrEvent, true)
	expected = "MANDATORY_IE_MISSING: [OriginID]"
	if err := sessions.BiRPCv1TerminateSession(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}
	cgrEvent.Event[utils.OriginID] = "ORIGIN_ID"

	cgrEvent.Opts = make(map[string]interface{})
	cgrEvent.Opts[utils.OptsDebitInterval] = "invalid_time_format"
	args = NewV1TerminateSessionArgs(true, false, false, nil, false, nil, cgrEvent, true)
	expected = "RALS_ERROR:time: invalid duration \"invalid_time_format\""
	if err := sessions.BiRPCv1TerminateSession(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}
	cgrEvent.Opts[utils.OptsDebitInterval] = "1m"

	//by this CGRID, there will be an empty session
	cgrEvent.Event[utils.CGRID] = "CGR_ID"
	sessions.aSessions = map[string]*Session{
		"CGR_ID": {},
	}
	args = NewV1TerminateSessionArgs(true, false, false, nil, false, nil, cgrEvent, true)
	if err := sessions.BiRPCv1TerminateSession(nil, args, &reply); err != nil {
		t.Error(err)
	}
	cgrEvent.Event[utils.CGRID] = "CHANGED_CGRID"

	args = NewV1TerminateSessionArgs(true, false, false, nil, false, nil, cgrEvent, true)
	sessions.cgrCfg.SessionSCfg().ChargerSConns = []string{}
	expected = "RALS_ERROR:ChargerS is disabled"
	if err := sessions.BiRPCv1TerminateSession(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}
	sessions.cgrCfg.SessionSCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)}

	//update session error
	cgrEvent.Event[utils.Usage] = "invalid_dur_time"
	args = NewV1TerminateSessionArgs(true, false, false, nil, false, nil, cgrEvent, true)
	expected = "time: invalid duration \"invalid_dur_time\""
	if err := sessions.BiRPCv1TerminateSession(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}
	cgrEvent.Event[utils.Usage] = "1m"

	cgrEvent = &utils.CGREvent{
		ID: "test_id",
		Event: map[string]interface{}{
			utils.Usage:    "10s",
			utils.OriginID: "TEST_ID",
		},
	}
	cfg = config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.CacheCfg().Partitions[utils.CacheClosedSessions] = &config.CacheParamCfg{
		Replicate: true,
	}
	cfg.SessionSCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.SessionSCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)}
	connMgr = engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers):   chanInternal,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator): chanInternal,
	})
	sessions = NewSessionS(cfg, dm, connMgr)
	caches = engine.NewCacheS(cfg, dm, nil)
	engine.Cache = caches
	args = NewV1TerminateSessionArgs(true, false, false, nil, false, nil, cgrEvent, true)
	expected = "RALS_ERROR:NOT_IMPLEMENTED"
	if err := sessions.BiRPCv1TerminateSession(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}
	engine.Cache = tmp
}

func TestBiRPCv1TerminateSession2(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	engine.Cache.Clear(nil)
	clnt := &testMockClients{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ResourceSv1ReleaseResources: func(args interface{}, reply interface{}) error {
				if args.(*utils.ArgRSv1ResourceUsage).Tenant == "CHANGED_ID" {
					return nil
				}
				return utils.ErrNotImplemented
			},
		},
	}
	chanInternal := make(chan rpcclient.ClientConnector, 1)
	chanInternal <- clnt
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().ResSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)}
	data := engine.NewInternalDB(nil, nil, true)
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources): chanInternal,
	})
	dm := engine.NewDataManager(data, cfg.CacheCfg(), connMgr)
	sessions := NewSessionS(cfg, dm, connMgr)

	cgrEvent := &utils.CGREvent{
		ID: "test_id",
		Event: map[string]interface{}{
			utils.Usage:    "10s",
			utils.OriginID: "TEST_ID",
		},
	}
	args := NewV1TerminateSessionArgs(false, true, false, nil, false, nil, cgrEvent, true)
	var reply string
	expected := "RESOURCES_ERROR:NOT_IMPLEMENTED"
	if err := sessions.BiRPCv1TerminateSession(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}

	cgrEvent.Event[utils.OriginID] = utils.EmptyString
	args = NewV1TerminateSessionArgs(false, true, false, nil, false, nil, cgrEvent, true)
	expected = "MANDATORY_IE_MISSING: [OriginID]"
	if err := sessions.BiRPCv1TerminateSession(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}
	cgrEvent.Event[utils.OriginID] = "ORIGIN_ID"

	args = NewV1TerminateSessionArgs(false, true, false, nil, false, nil, cgrEvent, true)
	expected = "NOT_CONNECTED: ResourceS"
	sessions.cgrCfg.SessionSCfg().ResSConns = []string{}
	if err := sessions.BiRPCv1TerminateSession(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}
	sessions.cgrCfg.SessionSCfg().ResSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)}

	cgrEvent.Tenant = "CHANGED_ID"
	args = NewV1TerminateSessionArgs(false, true, true, nil, true, nil, cgrEvent, true)
	if err := sessions.BiRPCv1TerminateSession(nil, args, &reply); err == nil || err != utils.ErrPartiallyExecuted {
		t.Errorf("Exepected %+v, received %+v", utils.ErrPartiallyExecuted, err)
	}
}

func TestBiRPCv1ProcessCDR(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	tmp := engine.Cache

	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	sessions := NewSessionS(cfg, dm, nil)

	cgrEvent := &utils.CGREvent{
		Event: map[string]interface{}{
			utils.Usage:    "10s",
			utils.OriginID: "TEST_ID",
		},
	}
	var reply string

	cgrEvent.ID = utils.EmptyString
	expected := "MANDATORY_IE_MISSING: [connIDs]"
	if err := sessions.BiRPCv1ProcessCDR(nil, cgrEvent, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}
	cgrEvent.ID = "test_id"

	caches := engine.NewCacheS(cfg, dm, nil)
	//value's error will be nil, so the error of the initiate sessions will be the same
	value := &utils.CachedRPCResponse{
		Result: utils.StringPointer("ROUTE_LEASTCOST_1"),
	}
	engine.Cache = caches
	engine.Cache.SetWithoutReplicate(utils.CacheRPCResponses, utils.ConcatenatedKey(utils.SessionSv1ProcessCDR, cgrEvent.ID),
		value, nil, true, utils.NonTransactional)
	if err := sessions.BiRPCv1ProcessCDR(nil, cgrEvent, &reply); err != nil {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}
	engine.Cache = tmp
}

func TestBiRPCv1ProcessMessage1(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	tmp := engine.Cache

	engine.Cache.Clear(nil)
	clnt := &testMockClients{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.AttributeSv1ProcessEvent: func(args interface{}, reply interface{}) error {
				if args.(*engine.AttrArgsProcessEvent).ID == "test_id" {
					return nil
				}
				return utils.ErrNotImplemented
			},
		},
	}
	chanInternal := make(chan rpcclient.ClientConnector, 1)
	chanInternal <- clnt
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	cfg.SessionSCfg().AttrSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}
	data := engine.NewInternalDB(nil, nil, true)
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes): chanInternal,
	})
	dm := engine.NewDataManager(data, cfg.CacheCfg(), connMgr)
	sessions := NewSessionS(cfg, dm, connMgr)

	cgrEvent := &utils.CGREvent{
		ID: "test_id",
		Event: map[string]interface{}{
			utils.Usage:    "10s",
			utils.OriginID: "TEST_ID",
		},
	}

	args := NewV1ProcessMessageArgs(false, []string{},
		false, []string{}, false, []string{}, false, false,
		true, false, false, nil, utils.Paginator{}, false, "1")
	reply := V1ProcessMessageReply{}
	expected := "MANDATORY_IE_MISSING: [CGREvent]"
	if err := sessions.BiRPCv1ProcessMessage(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}

	cgrEvent.ID = utils.EmptyString
	args = NewV1ProcessMessageArgs(true, []string{},
		false, []string{}, false, []string{}, true, false,
		false, false, false, cgrEvent, utils.Paginator{}, false, "1")
	expected = "ATTRIBUTES_ERROR:NOT_IMPLEMENTED"
	if err := sessions.BiRPCv1ProcessMessage(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}

	cgrEvent.ID = "test_id"
	args = NewV1ProcessMessageArgs(true, []string{},
		false, []string{}, false, []string{}, true, false,
		false, false, false, cgrEvent, utils.Paginator{}, false, "1")
	expected = "NOT_CONNECTED: ResourceS"
	if err := sessions.BiRPCv1ProcessMessage(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	caches := engine.NewCacheS(cfg, dm, nil)
	//value's error will be nil, so the error of the initiate sessions will be the same
	value := &utils.CachedRPCResponse{
		Result: &V1ProcessMessageReply{
			MaxUsage: utils.DurationPointer(time.Hour),
		},
	}
	engine.Cache = caches
	args = NewV1ProcessMessageArgs(true, []string{},
		false, []string{}, false, []string{}, true, false,
		false, false, false, cgrEvent, utils.Paginator{}, false, "1")
	engine.Cache.SetWithoutReplicate(utils.CacheRPCResponses, utils.ConcatenatedKey(utils.SessionSv1ProcessMessage, args.CGREvent.ID),
		value, nil, true, utils.NonTransactional)
	if err := sessions.BiRPCv1ProcessMessage(nil, args, &reply); err != nil {
		t.Error(err)
	}
	engine.Cache = tmp
}

func TestBiRPCv1ProcessMessage2(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	engine.Cache.Clear(nil)
	clnt := &testMockClients{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ResourceSv1AllocateResources: func(args interface{}, reply interface{}) error {
				if args.(*utils.ArgRSv1ResourceUsage).UsageID == "ORIGIN_ID" {
					return nil
				}
				return utils.ErrNotImplemented
			},
			utils.RouteSv1GetRoutes: func(args interface{}, reply interface{}) error {
				*reply.(*engine.SortedRoutesList) = engine.SortedRoutesList{{
					Routes: []*engine.SortedRoute{
						{
							RouteID: "ROUTE_ID",
						},
					},
				}}
				return nil
			},
			utils.ChargerSv1ProcessEvent: func(args interface{}, reply interface{}) error {
				chrgrs := []*engine.ChrgSProcessEventReply{
					{ChargerSProfile: "TEST_PROFILE1",
						CGREvent: &utils.CGREvent{
							Tenant: "cgrates.org",
							ID:     "TEST_ID",
							Event: map[string]interface{}{
								utils.Destination: "10",
							},
						}},
				}
				*reply.(*[]*engine.ChrgSProcessEventReply) = chrgrs
				return nil
			},
		},
	}
	chanInternal := make(chan rpcclient.ClientConnector, 1)
	chanInternal <- clnt
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	cfg.SessionSCfg().ResSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)}
	data := engine.NewInternalDB(nil, nil, true)
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources): chanInternal,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRoutes):    chanInternal,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers):  chanInternal,
	})
	dm := engine.NewDataManager(data, cfg.CacheCfg(), connMgr)
	sessions := NewSessionS(cfg, dm, connMgr)

	cgrEvent := &utils.CGREvent{
		ID: "test_id",
		Event: map[string]interface{}{
			utils.Usage: "10s",
		},
	}

	args := NewV1ProcessMessageArgs(false, []string{},
		false, []string{}, false, []string{}, true, false,
		true, false, false, cgrEvent, utils.Paginator{}, false, "1")
	reply := V1ProcessMessageReply{}
	expected := "MANDATORY_IE_MISSING: [OriginID]"
	if err := sessions.BiRPCv1ProcessMessage(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}
	cgrEvent.Event[utils.OriginID] = "ID"

	args = NewV1ProcessMessageArgs(false, []string{},
		false, []string{}, false, []string{}, true, false,
		false, false, false, cgrEvent, utils.Paginator{}, false, "1")
	reply = V1ProcessMessageReply{}
	expected = "RESOURCES_ERROR:NOT_IMPLEMENTED"
	if err := sessions.BiRPCv1ProcessMessage(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}

	cgrEvent.Event[utils.OriginID] = "ORIGIN_ID"
	args = NewV1ProcessMessageArgs(false, []string{},
		false, []string{}, false, []string{}, true, true,
		true, false, false, cgrEvent, utils.Paginator{}, false, "1")
	expected = "NOT_CONNECTED: RouteS"
	if err := sessions.BiRPCv1ProcessMessage(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}

	sessions.cgrCfg.SessionSCfg().RouteSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRoutes)}
	expected = "ChargerS is disabled"
	if err := sessions.BiRPCv1ProcessMessage(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}

	args = NewV1ProcessMessageArgs(false, []string{},
		true, []string{}, true, []string{}, true, true,
		true, false, false, cgrEvent, utils.Paginator{}, false, "1")
	sessions.cgrCfg.SessionSCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)}

	if err := sessions.BiRPCv1ProcessMessage(nil, args, &reply); err == nil || err != utils.ErrPartiallyExecuted {
		t.Errorf("Exepected %+v, received %+v", utils.ErrPartiallyExecuted, err)
	}
}

func TestBiRPCv1ProcessEvent(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	tmp := engine.Cache

	engine.Cache.Clear(nil)
	clnt := &testMockClients{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ChargerSv1ProcessEvent: func(args interface{}, reply interface{}) error {
				chrgrs := []*engine.ChrgSProcessEventReply{
					{ChargerSProfile: "TEST_PROFILE1",
						CGREvent: &utils.CGREvent{
							Tenant: "cgrates.org",
							ID:     "TEST_ID",
							Event: map[string]interface{}{
								utils.Destination: "10",
							},
						}},
				}
				*reply.(*[]*engine.ChrgSProcessEventReply) = chrgrs
				return nil
			},
			utils.AttributeSv1ProcessEvent: func(args interface{}, reply interface{}) error {
				attrs := engine.AttrSProcessEventReply{
					CGREvent: &utils.CGREvent{
						Tenant: "cgrates.org",
						ID:     "TEST_ID",
						Event: map[string]interface{}{
							utils.Destination: "10",
						},
					},
				}
				if args.(*engine.AttrArgsProcessEvent).ID == "CHANGED_ID" {
					*reply.(*engine.AttrSProcessEventReply) = attrs
					return nil
				}
				return utils.ErrNotImplemented
			},
			utils.RouteSv1GetRoutes: func(args interface{}, reply interface{}) error {
				if args.(*engine.ArgsGetRoutes).ID == "SECOND_ID" {
					*reply.(*engine.SortedRoutesList) = engine.SortedRoutesList{{
						ProfileID: "ROUTE_PRFID",
						Routes: []*engine.SortedRoute{
							{
								RouteID: "ROUTE_ID",
							},
						},
					}}
					return nil
				}
				return utils.ErrNotImplemented
			},
			utils.ThresholdSv1ProcessEvent: func(args interface{}, reply interface{}) error {
				return utils.ErrNotImplemented
			},
		},
	}
	chanInternal := make(chan rpcclient.ClientConnector, 1)
	chanInternal <- clnt
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	data := engine.NewInternalDB(nil, nil, true)
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers):   chanInternal,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes): chanInternal,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRoutes):     chanInternal,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds): chanInternal,
	})
	dm := engine.NewDataManager(data, cfg.CacheCfg(), connMgr)
	sessions := NewSessionS(cfg, dm, connMgr)

	cgrEvent := &utils.CGREvent{
		ID: "test_id",
		Event: map[string]interface{}{
			utils.Usage:    "10s",
			utils.OriginID: "TEST_ID",
		},
	}

	args := &V1ProcessEventArgs{
		Flags: []string{utils.MetaChargers},
	}
	reply := V1ProcessEventReply{}
	expected := "MANDATORY_IE_MISSING: [CGREvent]"
	if err := sessions.BiRPCv1ProcessEvent(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}
	args.CGREvent = cgrEvent
	caches := engine.NewCacheS(cfg, dm, nil)
	//value's error will be nil, so the error of the initiate sessions will be the same
	value := &utils.CachedRPCResponse{
		Result: &V1ProcessEventReply{
			MaxUsage: map[string]time.Duration{
				utils.Usage: time.Hour,
			},
		},
	}
	engine.Cache = caches
	engine.Cache.SetWithoutReplicate(utils.CacheRPCResponses, utils.ConcatenatedKey(utils.SessionSv1ProcessEvent, args.CGREvent.ID),
		value, nil, true, utils.NonTransactional)
	if err := sessions.BiRPCv1ProcessEvent(nil, args, &reply); err != nil {
		t.Error(err)
	}
	engine.Cache = tmp

	cgrEvent.ID = utils.EmptyString
	expected = "CHARGERS_ERROR:MANDATORY_IE_MISSING: [connIDs]"
	if err := sessions.BiRPCv1ProcessEvent(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}

	args.CGREvent.ID = "TEST_ID"
	args.Flags = append(args.Flags, utils.MetaAttributes)
	sessions.cgrCfg.SessionSCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)}
	sessions.cgrCfg.SessionSCfg().AttrSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}
	expected = "ATTRIBUTES_ERROR:NOT_IMPLEMENTED"
	if err := sessions.BiRPCv1ProcessEvent(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}

	args.CGREvent.ID = "CHANGED_ID"
	args.Flags = append(args.Flags, utils.MetaRoutes)
	sessions.cgrCfg.SessionSCfg().RouteSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRoutes)}
	expected = "ROUTES_ERROR:NOT_IMPLEMENTED"
	if err := sessions.BiRPCv1ProcessEvent(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}

	args.Flags = []string{utils.MetaRoutes, "*routes:*event_cost:2", utils.MetaThresholds}
	args.CGREvent.ID = "SECOND_ID"
	sessions.cgrCfg.SessionSCfg().ThreshSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)}
	expected = "PARTIALLY_EXECUTED"
	if err := sessions.BiRPCv1ProcessEvent(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}

	args.Flags = []string{utils.MetaThresholds, utils.MetaBlockerError}
	sessions.cgrCfg.SessionSCfg().ThreshSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)}
	expected = "THRESHOLDS_ERROR:NOT_IMPLEMENTED"
	if err := sessions.BiRPCv1ProcessEvent(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}

}

func TestBiRPCv1ProcessEventStats(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	engine.Cache.Clear(nil)
	clnt := &testMockClients{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ChargerSv1ProcessEvent: func(args interface{}, reply interface{}) error {
				chrgrs := []*engine.ChrgSProcessEventReply{
					{ChargerSProfile: "TEST_PROFILE1",
						CGREvent: &utils.CGREvent{
							Tenant: "cgrates.org",
							ID:     "TEST_ID",
							Event: map[string]interface{}{
								utils.Destination: "10",
							},
						}},
				}
				*reply.(*[]*engine.ChrgSProcessEventReply) = chrgrs
				return nil
			},
			utils.StatSv1ProcessEvent: func(args interface{}, reply interface{}) error {
				return utils.ErrNotImplemented
			},
		},
	}
	chanInternal := make(chan rpcclient.ClientConnector, 1)
	chanInternal <- clnt
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().StatSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)}
	cfg.SessionSCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)}
	data := engine.NewInternalDB(nil, nil, true)
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats):    chanInternal,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers): chanInternal,
	})
	dm := engine.NewDataManager(data, cfg.CacheCfg(), connMgr)
	sessions := NewSessionS(cfg, dm, connMgr)

	cgrEvent := &utils.CGREvent{
		ID: "test_id",
		Event: map[string]interface{}{
			utils.Usage:    "10s",
			utils.OriginID: "TEST_ID",
		},
	}

	args := &V1ProcessEventArgs{
		CGREvent: cgrEvent,
		Flags:    []string{utils.MetaChargers, utils.MetaStats},
	}
	reply := V1ProcessEventReply{}
	expected := "PARTIALLY_EXECUTED"
	if err := sessions.BiRPCv1ProcessEvent(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}

	args.Flags = []string{utils.MetaStats, utils.MetaBlockerError}
	expected = "STATS_ERROR:NOT_IMPLEMENTED"
	if err := sessions.BiRPCv1ProcessEvent(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}

	args.Flags = []string{utils.MetaSTIRAuthenticate}
	args.CGREvent.Opts = make(map[string]interface{})
	args.CGREvent.Opts[utils.OptsStirATest] = "stir;test;opts"
	expected = "*stir_authenticate: missing parts of the message header"
	if err := sessions.BiRPCv1ProcessEvent(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}

	args.Flags = []string{utils.MetaSTIRInitiate}
	args.CGREvent.Opts = make(map[string]interface{})
	args.CGREvent.Opts[utils.OptsStirATest] = "stir;test;opts"
	expected = "*stir_authenticate: open : no such file or directory"
	if err := sessions.BiRPCv1ProcessEvent(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}

	args.CGREvent.Opts[utils.OptsStirOriginatorURI] = "+407590336423;USER_ID"
	if err := sessions.BiRPCv1ProcessEvent(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}
}
func TestBiRPCv1ProcessEventResources(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	engine.Cache.Clear(nil)
	clnt := &testMockClients{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ChargerSv1ProcessEvent: func(args interface{}, reply interface{}) error {
				return nil
			},
			utils.StatSv1ProcessEvent: func(args interface{}, reply interface{}) error {
				return utils.ErrNotImplemented
			},
		},
	}
	chanInternal := make(chan rpcclient.ClientConnector, 1)
	chanInternal <- clnt
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)}
	data := engine.NewInternalDB(nil, nil, true)
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources): chanInternal,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers):  chanInternal,
	})
	dm := engine.NewDataManager(data, cfg.CacheCfg(), connMgr)
	sessions := NewSessionS(cfg, dm, connMgr)

	args := &V1ProcessEventArgs{
		Flags: []string{
			utils.ConcatenatedKey(utils.MetaResources, utils.MetaDerivedReply),
			utils.ConcatenatedKey(utils.MetaResources, utils.MetaAuthorize),
			utils.MetaChargers},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testBiRPCv1ProcessEventStatsResources",
			Event: map[string]interface{}{
				utils.Tenant:       "cgrates.org",
				utils.ToR:          utils.MetaVoice,
				utils.AccountField: "1001",
				utils.Subject:      "ANY2CNT",
				utils.Destination:  "1002",
			},
		},
	}

	reply := V1ProcessEventReply{}
	expected := "NOT_CONNECTED: ResourceS"
	if err := sessions.BiRPCv1ProcessEvent(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}
	sessions.cgrCfg.SessionSCfg().ResSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)}
	args.Flags = append(args.Flags, utils.MetaResources)

	expected = "MANDATORY_IE_MISSING: [OriginID]"
	if err := sessions.BiRPCv1ProcessEvent(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}

	args.CGREvent.Event[utils.OriginID] = "ORIGIN_ID"
	args.Flags = append(args.Flags, utils.MetaBlockerError)
	expected = "RESOURCES_ERROR:UNSUPPORTED_SERVICE_METHOD"
	if err := sessions.BiRPCv1ProcessEvent(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}

	args.Flags = args.Flags[:len(args.Flags)-1]
	expected = "PARTIALLY_EXECUTED"
	if err := sessions.BiRPCv1ProcessEvent(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}

	args.Flags = []string{utils.ConcatenatedKey(utils.MetaResources, utils.MetaDerivedReply),
		utils.ConcatenatedKey(utils.MetaResources, utils.MetaAllocate),
		utils.MetaChargers}
	args.Flags = append(args.Flags, utils.MetaBlockerError)
	expected = "RESOURCES_ERROR:UNSUPPORTED_SERVICE_METHOD"
	if err := sessions.BiRPCv1ProcessEvent(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}

	args.Flags = args.Flags[:len(args.Flags)-1]
	expected = "PARTIALLY_EXECUTED"
	if err := sessions.BiRPCv1ProcessEvent(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}

	args.Flags = []string{utils.ConcatenatedKey(utils.MetaResources, utils.MetaDerivedReply),
		utils.ConcatenatedKey(utils.MetaResources, utils.MetaRelease),
		utils.MetaChargers}
	args.Flags = append(args.Flags, utils.MetaBlockerError)
	expected = "RESOURCES_ERROR:UNSUPPORTED_SERVICE_METHOD"
	if err := sessions.BiRPCv1ProcessEvent(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}

	args.Flags = args.Flags[:len(args.Flags)-1]
	expected = "PARTIALLY_EXECUTED"
	if err := sessions.BiRPCv1ProcessEvent(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}
}

func TestBiRPCv1ProcessEventRals1(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	engine.Cache.Clear(nil)
	clnt := &testMockClients{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ChargerSv1ProcessEvent: func(args interface{}, reply interface{}) error {
				if args.(*utils.CGREvent).ID != "RALS_ID" {
					return nil
				}
				chrgrs := []*engine.ChrgSProcessEventReply{
					{ChargerSProfile: "TEST_PROFILE1",
						CGREvent: &utils.CGREvent{
							Tenant: "cgrates.org",
							ID:     "TEST_ID",
							Event: map[string]interface{}{
								utils.Destination: "10",
							},
						}},
				}
				*reply.(*[]*engine.ChrgSProcessEventReply) = chrgrs
				return nil
			},
			utils.ResponderGetCost: func(args interface{}, reply interface{}) error {
				if args.(*engine.CallDescriptorWithAPIOpts).Tenant == "CHANGED_ID" {
					return nil
				}
				return utils.ErrNotImplemented
			},
		},
	}
	chanInternal := make(chan rpcclient.ClientConnector, 1)
	chanInternal <- clnt
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)}
	cfg.SessionSCfg().RALsConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs)}
	data := engine.NewInternalDB(nil, nil, true)
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs):     chanInternal,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers): chanInternal,
	})
	dm := engine.NewDataManager(data, cfg.CacheCfg(), connMgr)
	sessions := NewSessionS(cfg, dm, connMgr)

	args := &V1ProcessEventArgs{
		Flags: []string{utils.MetaRALs,
			utils.ConcatenatedKey(utils.MetaRALs, utils.MetaCost),
			utils.ConcatenatedKey(utils.MetaRALs, utils.MetaDerivedReply),
			utils.MetaChargers},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testBiRPCv1ProcessEventStatsResources",
			Event: map[string]interface{}{
				utils.Tenant:       "cgrates.org",
				utils.ToR:          utils.MetaVoice,
				utils.AccountField: "1001",
				utils.Destination:  "1002",
			},
		},
	}

	reply := V1ProcessEventReply{}
	if err := sessions.BiRPCv1ProcessEvent(nil, args, &reply); err == nil || err != utils.ErrNotImplemented {
		t.Errorf("Exepected %+v, received %+v", utils.ErrNotImplemented, err)
	}

	args.CGREvent.Tenant = "CHANGED_ID"
	args.Flags = []string{utils.MetaRALs,
		utils.ConcatenatedKey(utils.MetaRALs, utils.MetaCost),
		utils.ConcatenatedKey(utils.MetaRALs, utils.MetaDerivedReply),
		utils.ConcatenatedKey(utils.MetaRALs, utils.MetaAuthorize),
		utils.MetaChargers}
	args.CGREvent.Event[utils.Usage] = "invalid_usage_format"
	expected := "time: invalid duration \"invalid_usage_format\""
	if err := sessions.BiRPCv1ProcessEvent(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}

	args.CGREvent.Event[utils.Usage] = "1m"
	if err := sessions.BiRPCv1ProcessEvent(nil, args, &reply); err != nil {
		t.Error(expected)
	}

	args.Flags = []string{utils.MetaRALs,
		utils.ConcatenatedKey(utils.MetaRALs, utils.MetaInitiate),
		utils.MetaChargers}
	args.Opts = make(map[string]interface{})
	args.Opts[utils.OptsDebitInterval] = "invalid_dbtitrvl_format"
	expected = "RALS_ERROR:time: invalid duration \"invalid_dbtitrvl_format\""
	if err := sessions.BiRPCv1ProcessEvent(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}
	args.Opts[utils.OptsDebitInterval] = "5s"

	sessions.cgrCfg.SessionSCfg().ChargerSConns = []string{}
	expected = "ChargerS is disabled"
	if err := sessions.BiRPCv1ProcessEvent(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}
	sessions.cgrCfg.SessionSCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)}

	args.CGREvent.Event[utils.Usage] = "invalid_format"
	args.CGREvent.Tenant = "cgrates.org"
	expected = "RALS_ERROR:time: invalid duration \"invalid_format\""
	if err := sessions.BiRPCv1ProcessEvent(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}
	args.CGREvent.Event[utils.Usage] = "10s"
}

func TestBiRPCv1ProcessEventRals2(t *testing.T) {
	tmp := engine.Cache
	log.SetOutput(ioutil.Discard)

	engine.Cache.Clear(nil)
	clnt := &testMockClients{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ChargerSv1ProcessEvent: func(args interface{}, reply interface{}) error {
				chrgrs := []*engine.ChrgSProcessEventReply{
					{ChargerSProfile: "TEST_PROFILE1",
						CGREvent: &utils.CGREvent{
							Tenant: "cgrates.org",
							ID:     "TEST_ID",
							Event: map[string]interface{}{
								utils.Destination: "10",
								utils.RequestType: utils.MetaPrepaid,
							},
						}},
				}
				*reply.(*[]*engine.ChrgSProcessEventReply) = chrgrs
				return nil
			},
			utils.CacheSv1ReplicateSet: func(args interface{}, reply interface{}) error {
				return utils.ErrNotImplemented
			},
			utils.ResponderMaxDebit: func(args interface{}, reply interface{}) error {
				return nil
			},
		},
	}
	chanInternal := make(chan rpcclient.ClientConnector, 1)
	chanInternal <- clnt
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)}
	cfg.SessionSCfg().RALsConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs)}
	cfg.SessionSCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	data := engine.NewInternalDB(nil, nil, true)
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs):       chanInternal,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers):   chanInternal,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator): chanInternal,
	})
	dm := engine.NewDataManager(data, cfg.CacheCfg(), connMgr)
	sessions := NewSessionS(cfg, dm, connMgr)

	args := &V1ProcessEventArgs{
		Flags: []string{utils.MetaRALs,
			utils.ConcatenatedKey(utils.MetaRALs, utils.MetaInitiate),
			utils.MetaChargers},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testBiRPCv1ProcessEventStatsResources",
			Event: map[string]interface{}{
				utils.Tenant:      "cgrates.org",
				utils.Destination: "1002",
				utils.RequestType: utils.MetaPrepaid,
			},
			Opts: map[string]interface{}{
				utils.OptsDebitInterval: "10s",
			},
		},
	}

	reply := V1ProcessEventReply{}

	if err := sessions.BiRPCv1ProcessEvent(nil, args, &reply); err != nil {
		t.Error(err)
	}

	args.Flags = []string{utils.MetaRALs,
		utils.ConcatenatedKey(utils.MetaRALs, utils.MetaUpdate),
		utils.MetaChargers}
	args.Opts[utils.OptsDebitInterval] = "invalid_format"
	expected := "RALS_ERROR:time: invalid duration \"invalid_format\""
	if err := sessions.BiRPCv1ProcessEvent(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}
	args.Opts[utils.OptsDebitInterval] = "10s"

	args.Event[utils.CGRID] = "test_id_new"
	sessions.cgrCfg.SessionSCfg().ChargerSConns = []string{}
	expected = "ChargerS is disabled"
	if err := sessions.BiRPCv1ProcessEvent(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}
	sessions.cgrCfg.SessionSCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)}
	args.Event[utils.CGRID] = utils.EmptyString

	args.CGREvent.Event[utils.Usage] = "invalid_format"
	expected = "RALS_ERROR:time: invalid duration \"invalid_format\""
	if err := sessions.BiRPCv1ProcessEvent(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}
	args.CGREvent.Event[utils.Usage] = "10"

	args.Flags = []string{utils.MetaRALs,
		utils.ConcatenatedKey(utils.MetaRALs, utils.MetaUpdate),
		utils.MetaChargers}
	delete(args.Opts, utils.OptsDebitInterval)
	if err := sessions.BiRPCv1ProcessEvent(nil, args, &reply); err != nil {
		t.Error(err)
	}

	args.Flags = []string{utils.MetaRALs,
		utils.ConcatenatedKey(utils.MetaRALs, utils.MetaTerminate),
		utils.MetaChargers}
	args.Opts[utils.OptsDebitInterval] = "invalid_format"
	expected = "RALS_ERROR:time: invalid duration \"invalid_format\""
	if err := sessions.BiRPCv1ProcessEvent(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}
	args.Opts[utils.OptsDebitInterval] = "10s"

	args.Event[utils.CGRID] = "test_id_new"
	sessions.cgrCfg.SessionSCfg().ChargerSConns = []string{}
	expected = "ChargerS is disabled"
	if err := sessions.BiRPCv1ProcessEvent(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}
	sessions.cgrCfg.SessionSCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)}
	args.Event[utils.CGRID] = utils.EmptyString

	engine.Cache.Clear(nil)
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.CacheCfg().Partitions[utils.CacheClosedSessions] = &config.CacheParamCfg{
		Replicate: true,
	}
	cacheS := engine.NewCacheS(cfg, nil, nil)
	engine.Cache = cacheS
	engine.SetConnManager(connMgr)
	expected = "RALS_ERROR:NOT_IMPLEMENTED"
	if err := sessions.BiRPCv1ProcessEvent(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}

	engine.Cache = tmp
}

func TestBiRPCv1ProcessEventCDRs(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	engine.Cache.Clear(nil)
	clnt := &testMockClients{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ChargerSv1ProcessEvent: func(args interface{}, reply interface{}) error {
				chrgrs := []*engine.ChrgSProcessEventReply{
					{ChargerSProfile: "TEST_PROFILE1",
						CGREvent: &utils.CGREvent{
							Tenant: "cgrates.org",
							ID:     "TEST_ID",
							Event: map[string]interface{}{
								utils.Destination: "10",
								utils.RequestType: utils.MetaPrepaid,
							},
						}},
				}
				*reply.(*[]*engine.ChrgSProcessEventReply) = chrgrs
				return nil
			},
			utils.CDRsV1ProcessEvent: func(args interface{}, reply interface{}) error {
				return utils.ErrNotImplemented
			},
		},
	}
	chanInternal := make(chan rpcclient.ClientConnector, 1)
	chanInternal <- clnt
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)}
	cfg.SessionSCfg().CDRsConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs)}
	data := engine.NewInternalDB(nil, nil, true)
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers): chanInternal,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs):     chanInternal,
	})
	dm := engine.NewDataManager(data, cfg.CacheCfg(), connMgr)
	sessions := NewSessionS(cfg, dm, connMgr)

	args := &V1ProcessEventArgs{
		Flags: []string{utils.MetaCDRs,
			utils.ConcatenatedKey(utils.MetaRALs, utils.MetaDerivedReply),
			utils.MetaChargers},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testBiRPCv1ProcessEventStatsResources",
			Event: map[string]interface{}{
				utils.Tenant:      "cgrates.org",
				utils.Destination: "1002",
				utils.RequestType: utils.MetaPrepaid,
			},
			Opts: map[string]interface{}{
				utils.OptsDebitInterval: "10s",
			},
		},
	}
	reply := V1ProcessEventReply{}

	expected := "PARTIALLY_EXECUTED"
	if err := sessions.BiRPCv1ProcessEvent(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}

	args.Flags = []string{utils.MetaCDRs, utils.MetaBlockerError,
		utils.ConcatenatedKey(utils.MetaRALs, utils.MetaDerivedReply),
		utils.MetaChargers}
	expected = "CDRS_ERROR:NOT_IMPLEMENTED"
	if err := sessions.BiRPCv1ProcessEvent(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}

	sessions.cgrCfg.SessionSCfg().CDRsConns = []string{}
	expected = "NOT_CONNECTED: CDRs"
	if err := sessions.BiRPCv1ProcessEvent(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}
}

func TestBiRPCv1GetCost(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	tmp := engine.Cache

	engine.Cache.Clear(nil)
	clnt := &testMockClients{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.AttributeSv1ProcessEvent: func(args interface{}, reply interface{}) error {
				attr := &engine.AttrSProcessEventReply{
					CGREvent: &utils.CGREvent{
						Tenant: "cgrates.org",
						ID:     "ATTRIBUTES",
						Event: map[string]interface{}{
							utils.Usage: "20m",
						},
					},
				}
				*reply.(*engine.AttrSProcessEventReply) = *attr
				return nil
			},
			utils.ResponderGetCost: func(args interface{}, reply interface{}) error {
				return nil
			},
		},
	}
	chanInternal := make(chan rpcclient.ClientConnector, 1)
	chanInternal <- clnt
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	data := engine.NewInternalDB(nil, nil, true)
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs):       chanInternal,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes): chanInternal,
	})
	dm := engine.NewDataManager(data, cfg.CacheCfg(), connMgr)
	sessions := NewSessionS(cfg, dm, connMgr)

	args := &V1ProcessEventArgs{
		Flags: []string{utils.MetaAttributes,
			utils.ConcatenatedKey(utils.MetaRALs, utils.MetaDerivedReply),
			utils.MetaChargers},
	}
	cgrEvent := &utils.CGREvent{
		ID: "TestBiRPCv1GetCost",
		Event: map[string]interface{}{
			utils.Tenant:      "cgrates.org",
			utils.Destination: "1002",
			utils.RequestType: utils.MetaPrepaid,
		},
		Opts: map[string]interface{}{
			utils.OptsDebitInterval: "10s",
		},
	}
	reply := V1GetCostReply{}
	expected := "MANDATORY_IE_MISSING: [CGREvent]"
	if err := sessions.BiRPCv1GetCost(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}
	args.CGREvent = cgrEvent

	caches := engine.NewCacheS(cfg, dm, nil)
	//value's error will be nil, so the error of the initiate sessions will be the same
	value := &utils.CachedRPCResponse{
		Result: &V1GetCostReply{
			EventCost: &engine.EventCost{
				Cost: utils.Float64Pointer(1.50),
			},
		},
	}
	engine.Cache = caches
	engine.Cache.SetWithoutReplicate(utils.CacheRPCResponses, utils.ConcatenatedKey(utils.SessionSv1GetCost, args.CGREvent.ID),
		value, nil, true, utils.NonTransactional)
	if err := sessions.BiRPCv1GetCost(nil, args, &reply); err != nil {
		t.Error(err)
	}
	engine.Cache = tmp

	args.CGREvent.ID = utils.EmptyString
	expected = "ATTRIBUTES_ERROR:NOT_CONNECTED: AttributeS"
	if err := sessions.BiRPCv1GetCost(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}
	sessions.cgrCfg.SessionSCfg().AttrSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}

	expected = "MANDATORY_IE_MISSING: [connIDs]"
	if err := sessions.BiRPCv1GetCost(nil, args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}
	sessions.cgrCfg.SessionSCfg().RALsConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs)}

	expectedVal := V1GetCostReply{
		Attributes: &engine.AttrSProcessEventReply{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "ATTRIBUTES",
				Event: map[string]interface{}{
					utils.Usage: "20m",
				},
			},
		},
		EventCost: &engine.EventCost{
			CGRID:         "ATTRIBUTES",
			Usage:         utils.DurationPointer(0),
			Cost:          utils.Float64Pointer(0),
			Charges:       []*engine.ChargingInterval{},
			Rating:        map[string]*engine.RatingUnit{},
			Accounting:    map[string]*engine.BalanceCharge{},
			RatingFilters: map[string]engine.RatingMatchedFilters{},
			Rates:         map[string]engine.RateGroups{},
			Timings:       map[string]*engine.ChargedTiming{},
		},
	}
	expectedVal.EventCost.Compute()
	if err := sessions.BiRPCv1GetCost(nil, args, &reply); err != nil {
		t.Errorf("Exepected %+v, received %+v", expected, err)
	}
	/*
		else if !reflect.DeepEqual(expectedVal, reply) {
			fmt.Printf("%T and %T \n", expectedVal.EventCost.Cost, reply.EventCost.Cost)
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedVal), utils.ToJSON(reply))
		}

	*/
}
