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

	"github.com/cgrates/rpcclient"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	sTests = []func(t *testing.T){
		testDebitLoopSessionLowBalance,
		testSetSTerminator,
		testSetSTerminatorError,
		testSetSTerminatorAutomaticTermination,
		testSetSTerminatorManualTermination,
		testForceSTerminatorManualTermination,
		testForceSTerminatorPostCDRs,
		testForceSTerminatorReleaseSession,
		testForceSTerminatorClientCall,
		testDebitSession,
		testDebitSessionResponderMaxDebit,
		testDebitSessionResponderMaxDebitError,
		testInitSessionDebitLoops,
		testDebitLoopSessionFrcDiscLowerDbtInterval,
		testDebitLoopSessionErrorDebiting,
		testDebitLoopSession,
		testDebitLoopSessionWarningSessions,
		testDebitLoopSessionDisconnectSession,
		testStoreSCost,
		testRefundSession,
		testRoundCost,
		testDisconnectSession,
		testReplicateSessions,
		testNewSession,
		testProcessChargerS,
		testTransitSState,
		testRelocateSession,
		testGetRelocateSession,
		testLibsessionsSetMockErrors,
		testSyncSessions,
		testAuthEvent,
		testAuthEventMockCall,
		testChargeEvent,
		testUpdateSession,
		testEndSession,
		testCallBiRPC,
		testBiRPCv1GetActivePassiveSessions,
		testBiRPCv1SetPassiveSession,
		testBiRPCv1ReplicateSessions,
		testBiRPCv1AuthorizeEvent,
	}
)

func TestSessionsIT(t *testing.T) {
	for _, test := range sTests {
		log.SetOutput(ioutil.Discard)
		t.Run("Running Sessions tests", test)
	}
}

func testSetSTerminator(t *testing.T) {
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
		utils.OptsSessionTTL:          "1s",
		utils.OptsSessionTTLMaxDelay:  "1s",
		utils.OptsSessionTTLLastUsed:  "2s",
		utils.OptsSessionTTLLastUsage: "0s",
		utils.OptsSessionTTLUsage:     "5s",
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
		utils.OptsSessionTTL:          "1s",
		utils.OptsSessionTTLMaxDelay:  "1s",
		utils.OptsSessionTTLLastUsed:  "2s",
		utils.OptsSessionTTLLastUsage: "0s",
		utils.OptsSessionTTLUsage:     "5s",
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

func testSetSTerminatorError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().SessionTTL = time.Second
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	sessions := NewSessionS(cfg, dm, nil)

	ss := &Session{}

	utils.Logger, err = utils.Newlogger(utils.MetaStdLog, utils.EmptyString)
	if err != nil {
		t.Error(err)
	}
	utils.Logger.SetLogLevel(7)

	buff := new(bytes.Buffer)
	log.SetOutput(buff)

	//Cannot set a terminate when ttl is 0
	ss.OptsStart = engine.MapEvent{
		utils.OptsSessionTTL: "0s",
	}
	opts := engine.MapEvent{}
	sessions.setSTerminator(ss, opts)

	//Invalid format types for time duration
	ss.OptsStart = engine.MapEvent{
		utils.OptsSessionTTL: "invalid_time_format",
	}
	sessions.setSTerminator(ss, opts)
	expected := "cannot extract <*sessionTTL> for session:<>, from it's options: <{}>, err: <time: invalid duration \"invalid_time_format\">"
	if rcv := buff.String(); !strings.Contains(rcv, expected) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}

	buff = new(bytes.Buffer)
	log.SetOutput(buff)

	ss.OptsStart = engine.MapEvent{
		utils.OptsSessionTTL:         "1s",
		utils.OptsSessionTTLMaxDelay: "invalid_time_format",
	}
	sessions.setSTerminator(ss, opts)
	expected = "cannot extract <*sessionTTLMaxDelay> for session:<>, from it's options: <{}>, err: <time: invalid duration \"invalid_time_format\">"
	if rcv := buff.String(); !strings.Contains(rcv, expected) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}

	buff = new(bytes.Buffer)
	log.SetOutput(buff)

	ss.OptsStart = engine.MapEvent{
		utils.OptsSessionTTL:         "1s",
		utils.OptsSessionTTLMaxDelay: "2s",
		utils.OptsSessionTTLLastUsed: "invalid_time_format",
	}
	sessions.setSTerminator(ss, opts)
	expected = "cannot extract <*sessionTTLLastUsed> for session:<>, from it's options: <{}>, err: <time: invalid duration \"invalid_time_format\">"
	if rcv := buff.String(); !strings.Contains(rcv, expected) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}

	buff = new(bytes.Buffer)
	log.SetOutput(buff)

	ss.OptsStart = engine.MapEvent{
		utils.OptsSessionTTL:          "1s",
		utils.OptsSessionTTLMaxDelay:  "2s",
		utils.OptsSessionTTLLastUsed:  "1s",
		utils.OptsSessionTTLLastUsage: "invalid_time_format",
	}
	sessions.setSTerminator(ss, opts)
	expected = "cannot extract <*sessionTTLLastUsage> for session:<>, from it's options: <{}>, err: <time: invalid duration \"invalid_time_format\">"
	if rcv := buff.String(); !strings.Contains(rcv, expected) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}

	buff = new(bytes.Buffer)
	log.SetOutput(buff)

	cfg.SessionSCfg().SessionTTLMaxDelay = utils.DurationPointer(time.Second)
	sessions = NewSessionS(cfg, dm, nil)
	ss.OptsStart = engine.MapEvent{
		utils.OptsSessionTTLLastUsed:  "1s",
		utils.OptsSessionTTLLastUsage: "5s",
		utils.OptsSessionTTLUsage:     "invalid_time_format",
	}
	sessions.setSTerminator(ss, opts)
	expected = "cannot extract <*sessionTTLUsage> for session:<>, from it's options: <{}>, err: <time: invalid duration \"invalid_time_format\">"
	if rcv := buff.String(); !strings.Contains(rcv, expected) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}

	log.SetOutput(os.Stderr)
}

func testSetSTerminatorAutomaticTermination(t *testing.T) {
	ss := &Session{}

	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	sessions := NewSessionS(cfg, dm, nil)

	opts := engine.MapEvent{
		utils.OptsSessionTTL:          "1s",
		utils.OptsSessionTTLLastUsage: "0s",
	}

	sessions.setSTerminator(ss, opts)
	select {
	case <-time.After(3 * time.Second):
		t.Fatal("timeout")
	case <-ss.sTerminator.endChan:
	}
}

func testSetSTerminatorManualTermination(t *testing.T) {
	ss := &Session{}

	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	sessions := NewSessionS(cfg, dm, nil)

	opts := engine.MapEvent{
		utils.OptsSessionTTL: "1s",
	}

	sessions.setSTerminator(ss, opts)
	ss.sTerminator.endChan <- struct{}{}
}

func testForceSTerminatorManualTermination(t *testing.T) {
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

func testForceSTerminatorPostCDRs(t *testing.T) {
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

func testForceSTerminatorReleaseSession(t *testing.T) {
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

func testForceSTerminatorClientCall(t *testing.T) {
	sTestMock := &testMockClientConn{}

	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().NodeID = "ClientConnID"
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources): nil,
	})
	sessions := NewSessionS(cfg, dm, connMgr)
	sessions.RegisterIntBiJConn(sTestMock)

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

func testDebitSession(t *testing.T) {
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

func testDebitSessionResponderMaxDebit(t *testing.T) {
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

func testDebitSessionResponderMaxDebitError(t *testing.T) {
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

func testInitSessionDebitLoops(t *testing.T) {
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

func testDebitLoopSessionErrorDebiting(t *testing.T) {
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
	sessions.RegisterIntBiJConn(sTestMock)

	if _, err = sessions.debitLoopSession(ss, 0, time.Hour); err != nil {
		t.Error(err)
	}
}

func testDebitLoopSession(t *testing.T) {
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

func testDebitLoopSessionFrcDiscLowerDbtInterval(t *testing.T) {
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

func testDebitLoopSessionLowBalance(t *testing.T) {
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

func testDebitLoopSessionWarningSessions(t *testing.T) {
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

func testDebitLoopSessionDisconnectSession(t *testing.T) {
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
	sessions.RegisterIntBiJConn(sTestMock)

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

func testStoreSCost(t *testing.T) {
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

	if err := sessions.storeSCost(ss, 0); err != nil {
		t.Error(err)
	}
}

func testRefundSession(t *testing.T) {
	engine.Cache.Clear(nil)
	testMock1 := &testMockClients{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ResponderRefundIncrements: func(args interface{}, reply interface{}) error {
				if args.(*engine.CallDescriptorWithOpts).Opts != nil {
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
			"21a5ab9": &engine.RatingUnit{},
		},
		Accounting: map[string]*engine.BalanceCharge{
			"44d6c02": &engine.BalanceCharge{},
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

func testRoundCost(t *testing.T) {
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

func testDisconnectSession(t *testing.T) {
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
	sessions.RegisterIntBiJConn(sTestMock)
	sessions.biJIDs["test"] = &biJClient{
		conn: sTestMock,
	}

	if err := sessions.disconnectSession(ss, utils.EmptyString); err == nil || err != utils.ErrNoActiveSession {
		t.Errorf("Expected %+v, received %+v", utils.ErrNoActiveSession, err)
	}

	sTestMock1 := &mockConnWarnDisconnect1{}
	sessions.RegisterIntBiJConn(sTestMock1)
	sessions.biJIDs["test"] = &biJClient{
		conn: sTestMock1,
	}
	if err := sessions.disconnectSession(ss, utils.EmptyString); err != nil {
		t.Error(err)
	}
}

func testReplicateSessions(t *testing.T) {
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

	if err := sessions.replicateSessions("test_session", false,
		[]string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}); err != nil {
		t.Error(err)
	}
}

func testNewSession(t *testing.T) {
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
		"da39a3ee5e6b4b0d3255bfef95601890afd80709": &Session{},
	}
	//sessions already exists
	if _, err := sessions.newSession(cgrEv, "resourceID", "clientConnID",
		time.Second, false, false); err == nil || err.Error() != utils.ErrExists.Error() {
		t.Errorf("Expected %+v, received %+v", utils.ErrExists, err)
	}
}

func testProcessChargerS(t *testing.T) {
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
	engine.SetCache(cacheS)
	connMgr = engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator): sMock})
	engine.SetConnManager(connMgr)

	if _, err := sessions.processChargerS(cgrEv); err == nil || err != utils.ErrNotImplemented {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotImplemented, err)
	}

	engine.Cache = tmpCache
}

func testTransitSState(t *testing.T) {
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

func testRelocateSession(t *testing.T) {
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

func testGetRelocateSession(t *testing.T) {
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

func testLibsessionsSetMockErrors(t *testing.T) {
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
	engine.SetCache(cacheS)
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

func testSyncSessions(t *testing.T) {
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
	sessions.RegisterIntBiJConn(sTestMock1)

	sessions.aSessions = map[string]*Session{
		"SESS1": {
			CGRID: "TEST_CGRID",
		},
	}

	sessions.syncSessions()

	sessions.cgrCfg.GeneralCfg().ReplyTimeout = 1
	cacheS := engine.NewCacheS(cfg, nil, nil)
	engine.SetCache(cacheS)
	connMgr = engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator): chanInternal})
	engine.SetConnManager(connMgr)
	sessions.aSessions = map[string]*Session{
		"ORIGIN_ID": {},
	}
	sessions.syncSessions()

	engine.Cache = tmp

	//There are no sessions to be removed
	sessions.terminateSyncSessions([]string{"no_sesssion"})
}
func testAuthEvent(t *testing.T) {
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

func testAuthEventMockCall(t *testing.T) {
	//mocking the GetMaxSession for checking the error
	engine.Cache.Clear(nil)
	sTestMock := &testMockClients{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ResponderGetMaxSessionTime: func(args interface{}, reply interface{}) error {
				usage := args.(*engine.CallDescriptorWithOpts).Opts[utils.Usage]
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

func testChargeEvent(t *testing.T) {
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
	engine.SetCache(cacheS)
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

func testUpdateSession(t *testing.T) {
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

func testEndSession(t *testing.T) {
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

func (ss *Session) BiRPCv1TestCase(clnt rpcclient.ClientConnector, args *V1TerminateSessionArgs, authReply *string) (aux string, err error) {
	return utils.EmptyString, nil
}

func testCallBiRPC(t *testing.T) {
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
	if err := sessions.CallBiRPC(sTestMock, valid, args, reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	/*
		//inexistent client
		args.CGREvent = &utils.CGREvent{}
		valid = "BiRPCv1.TestCase"
		if err := sessions.CallBiRPC(sTestMock, valid, args, reply); err == nil || err.Error() != expected {
			t.Errorf("Expected %+v, received %+v", expected, err)
		}

	*/
}

func testBiRPCv1GetActivePassiveSessions(t *testing.T) {
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

func testBiRPCv1SetPassiveSession(t *testing.T) {
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

func testBiRPCv1ReplicateSessions(t *testing.T) {
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

func testBiRPCv1AuthorizeEvent(t *testing.T) {
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

	cgrEvent := &utils.CGREvent{
		Event: map[string]interface{}{
			utils.Usage: "10s",
		},
	}
	args := NewV1AuthorizeArgs(false, []string{},
		false, []string{}, false, []string{}, false, false,
		false, false, false, nil, utils.Paginator{}, false, "")

	rply := &V1AuthorizeReply{
		Attributes:   &engine.AttrSProcessEventReply{},
		Routes:       &engine.SortedRoutes{},
		StatQueueIDs: &[]string{},
		ThresholdIDs: &[]string{},
	}

	expected := "MANDATORY_IE_MISSING: [CGREvent]"
	if err := sessions.BiRPCv1AuthorizeEvent(nil, args, rply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	args.CGREvent = cgrEvent
	//RPC caching
	sessions.cgrCfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	expected = "MANDATORY_IE_MISSING: [subsystems]"
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
}
