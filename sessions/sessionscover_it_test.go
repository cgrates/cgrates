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
)

var (
	sTests = []func(t *testing.T){
		/*
			testSetSTerminator,
			testSetSTerminatorError,
			testSetSTerminatorAutomaticTermination,
			testSetSTerminatorManualTermination,

		*/
		testForceSTerminatorManualTermination,
	}
)

func TestSessionsIT(t *testing.T) {
	for _, test := range sTests {
		log.SetOutput(ioutil.Discard)
		t.Run("Runing Sessions tests", test)
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
