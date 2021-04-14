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

package ees

import (
	"bytes"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
	"github.com/cgrates/rpcclient"
)

func TestListenAndServe(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Cache = make(map[string]*config.CacheParamCfg)
	cgrCfg.EEsCfg().Cache = map[string]*config.CacheParamCfg{
		utils.MetaFileCSV: {
			Limit: -1,
			TTL:   5 * time.Second,
		},
		utils.MetaNone: {
			Limit: 0,
		},
	}
	newIDb := engine.NewInternalDB(nil, nil, true)
	newDM := engine.NewDataManager(newIDb, cgrCfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cgrCfg, nil, newDM)
	eeS := NewEventExporterS(cgrCfg, filterS, nil)
	stopChan := make(chan struct{}, 1)
	cfgRld := make(chan struct{}, 1)
	cfgRld <- struct{}{}
	go func() {
		time.Sleep(10 * time.Nanosecond)
		stopChan <- struct{}{}
	}()
	var err error
	utils.Logger, err = utils.Newlogger(utils.MetaStdLog, utils.EmptyString)
	if err != nil {
		t.Error(err)
	}
	utils.Logger.SetLogLevel(6)
	logBuf := new(bytes.Buffer)
	log.SetOutput(logBuf)
	eeS.ListenAndServe(stopChan, cfgRld)
	logExpect := "[INFO] <CoreS> starting <EventExporterS>"
	if rcv := logBuf.String(); !strings.Contains(rcv, logExpect) {
		t.Errorf("Expected %q but received %q", logExpect, rcv)
	}
	logBuf.Reset()
}

type testMockEvent struct {
	calls map[string]func(_ *context.Context, _, _ interface{}) error
}

func (sT *testMockEvent) Call(ctx *context.Context, method string, arg, rply interface{}) error {
	if call, has := sT.calls[method]; !has {
		return rpcclient.ErrUnsupporteServiceMethod
	} else {
		return call(ctx, arg, rply)
	}
}
func TestAttrSProcessEvent(t *testing.T) {
	testMock := &testMockEvent{
		calls: map[string]func(_ *context.Context, _, _ interface{}) error{
			utils.AttributeSv1ProcessEvent: func(_ *context.Context, args, reply interface{}) error {
				rplyEv := &engine.AttrSProcessEventReply{
					AlteredFields: []string{"testcase"},
				}
				*reply.(*engine.AttrSProcessEventReply) = *rplyEv
				return nil
			},
		},
	}
	cgrEv := &utils.CGREvent{
		APIOpts: map[string]interface{}{
			utils.OptsAttributesProcessRuns: "10",
		},
	}
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsNoLksCfg().AttributeSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}
	newIDb := engine.NewInternalDB(nil, nil, true)
	newDM := engine.NewDataManager(newIDb, cgrCfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cgrCfg, nil, newDM)
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- testMock
	connMgr := engine.NewConnManager(cgrCfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes): clientConn,
	})
	eeS := NewEventExporterS(cgrCfg, filterS, connMgr)
	// cgrEv := &utils.CGREvent{}
	if err := eeS.attrSProcessEvent(cgrEv, []string{}, utils.EmptyString); err != nil {
		t.Error(err)
	}
}

func TestAttrSProcessEvent2(t *testing.T) {
	engine.Cache.Clear(nil)
	testMock := &testMockEvent{
		calls: map[string]func(_ *context.Context, _, _ interface{}) error{
			utils.AttributeSv1ProcessEvent: func(_ *context.Context, _, _ interface{}) error {
				return utils.ErrNotFound
			},
		},
	}
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsNoLksCfg().AttributeSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}
	newIDb := engine.NewInternalDB(nil, nil, true)
	newDM := engine.NewDataManager(newIDb, cgrCfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cgrCfg, nil, newDM)
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- testMock
	connMgr := engine.NewConnManager(cgrCfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes): clientConn,
	})
	eeS := NewEventExporterS(cgrCfg, filterS, connMgr)
	cgrEv := &utils.CGREvent{}
	if err := eeS.attrSProcessEvent(cgrEv, []string{}, utils.EmptyString); err != nil {
		t.Error(err)
	}
}

func TestV1ProcessEvent(t *testing.T) {
	filePath := "/tmp/TestV1ProcessEvent"
	if err := os.MkdirAll(filePath, 0777); err != nil {
		t.Error(err)
	}
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Type = "*file_csv"
	cgrCfg.EEsCfg().Exporters[0].ID = "SQLExporterFull"
	cgrCfg.EEsCfg().Exporters[0].ExportPath = filePath
	newIDb := engine.NewInternalDB(nil, nil, true)
	newDM := engine.NewDataManager(newIDb, cgrCfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cgrCfg, nil, newDM)
	eeS := NewEventExporterS(cgrCfg, filterS, nil)
	cgrEv := &utils.CGREventWithEeIDs{
		EeIDs: []string{"SQLExporterFull"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "voiceEvent",
			Time:   utils.TimePointer(time.Now()),
			Event: map[string]interface{}{
				utils.CGRID:        utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()),
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "dsafdsaf",
				utils.OriginHost:   "192.168.1.1",
				utils.RequestType:  utils.MetaRated,
				utils.Tenant:       "cgrates.org",
				utils.Category:     "call",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Unix(1383813745, 0).UTC(),
				utils.AnswerTime:   time.Unix(1383813746, 0).UTC(),
				utils.Usage:        10 * time.Second,
				utils.RunID:        utils.MetaDefault,
				utils.Cost:         1.01,
				"ExtraFields": map[string]string{"extra1": "val_extra1",
					"extra2": "val_extra2", "extra3": "val_extra3"},
			},
		},
	}
	var rply map[string]map[string]interface{}
	rplyExpect := map[string]map[string]interface{}{
		"SQLExporterFull": {},
	}
	if err := eeS.V1ProcessEvent(cgrEv, &rply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rply, rplyExpect) {
		t.Errorf("Expected %q but received %q", rplyExpect, rply)
	}
	if err := os.RemoveAll(filePath); err != nil {
		t.Error(err)
	}
}

func TestV1ProcessEvent2(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Type = "*file_csv"
	cgrCfg.EEsCfg().Exporters[0].ID = "SQLExporterFull"
	cgrCfg.EEsCfg().Exporters[0].Filters = []string{"*prefix:~*req.Subject:20"}
	cgrCfg.EEsCfg().Exporters[0].Tenant = config.NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep)
	newIDb := engine.NewInternalDB(nil, nil, true)
	newDM := engine.NewDataManager(newIDb, cgrCfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cgrCfg, nil, newDM)
	eeS := NewEventExporterS(cgrCfg, filterS, nil)
	cgrEv := &utils.CGREventWithEeIDs{
		EeIDs: []string{"SQLExporterFull"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "voiceEvent",
			Time:   utils.TimePointer(time.Now()),
			Event: map[string]interface{}{
				utils.Subject: "1001",
				"ExtraFields": map[string]string{"extra1": "val_extra1",
					"extra2": "val_extra2", "extra3": "val_extra3"},
			},
			APIOpts: map[string]interface{}{
				utils.OptsAttributesProcessRuns: "10",
			},
		},
	}
	var rply map[string]map[string]interface{}
	errExpect := "NOT_FOUND"
	if err := eeS.V1ProcessEvent(cgrEv, &rply); err == nil || err.Error() != errExpect {
		t.Errorf("Expecting %q but received %q", errExpect, err)
	}

	errExpect = "NOT_FOUND:test"
	eeS.cfg.EEsCfg().Exporters[0].Filters = []string{"test"}
	if err := eeS.V1ProcessEvent(cgrEv, &rply); err == nil || err.Error() != errExpect {
		t.Errorf("Expecting %q but received %q", errExpect, err)
	}
}

func TestV1ProcessEvent3(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Type = "*file_csv"
	cgrCfg.EEsCfg().Exporters[0].ID = "SQLExporterFull"
	cgrCfg.EEsCfg().Exporters[0].Flags = utils.FlagsWithParams{
		utils.MetaAttributes: utils.FlagParams{},
	}
	newIDb := engine.NewInternalDB(nil, nil, true)
	newDM := engine.NewDataManager(newIDb, cgrCfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cgrCfg, nil, newDM)
	eeS := NewEventExporterS(cgrCfg, filterS, nil)
	cgrEv := &utils.CGREventWithEeIDs{
		EeIDs: []string{"SQLExporterFull"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "voiceEvent",
			Time:   utils.TimePointer(time.Now()),
			Event:  map[string]interface{}{},
		},
	}
	var rply map[string]map[string]interface{}
	errExpect := "MANDATORY_IE_MISSING: [connIDs]"
	if err := eeS.V1ProcessEvent(cgrEv, &rply); err == nil || err.Error() != errExpect {
		t.Errorf("Expecting %q but received %q", errExpect, err)
	}
}

func TestV1ProcessEvent4(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaHTTPPost
	cgrCfg.EEsCfg().Exporters[0].ID = "SQLExporterFull"
	cgrCfg.EEsCfg().Exporters[0].Synchronous = true
	newIDb := engine.NewInternalDB(nil, nil, true)
	newDM := engine.NewDataManager(newIDb, cgrCfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cgrCfg, nil, newDM)
	eeS := NewEventExporterS(cgrCfg, filterS, nil)
	eeS.eesChs = map[string]*ltcache.Cache{
		utils.MetaHTTPPost: ltcache.NewCache(1,
			time.Second, false, onCacheEvicted),
	}
	newEeS, err := NewEventExporter(cgrCfg, 0, filterS)
	if err != nil {
		t.Error(err)
	}
	eeS.eesChs[utils.MetaHTTPPost].Set("SQLExporterFull", newEeS, []string{"grp1"})
	cgrEv := &utils.CGREventWithEeIDs{
		EeIDs: []string{"SQLExporterFull"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "voiceEvent",
			Time:   utils.TimePointer(time.Now()),
			Event:  map[string]interface{}{},
			APIOpts: map[string]interface{}{
				utils.OptsEEsVerbose: struct{}{},
			},
		},
	}
	var rply map[string]map[string]interface{}
	errExpect := "PARTIALLY_EXECUTED"
	if err := eeS.V1ProcessEvent(cgrEv, &rply); err == nil || err.Error() != errExpect {
		t.Errorf("Expecting %q but received %q", errExpect, err)
	} else if len(rply) != 0 {
		t.Error("Unexpected reply result")
	}
}

type mockEventExporter struct{}

func (mockEventExporter) ID() string                              { return utils.EmptyString }
func (mockEventExporter) ExportEvent(cgrEv *utils.CGREvent) error { return nil }
func (mockEventExporter) OnEvicted(itdmID string, value interface{}) {
	utils.Logger.Warning("NOT IMPLEMENTED")
}
func (mockEventExporter) GetMetrics() utils.MapStorage {
	return utils.MapStorage{
		utils.NumberOfEvents:  int64(0),
		utils.PositiveExports: utils.StringSet{},
		utils.NegativeExports: 5,
	}
}

func TestV1ProcessEventMockMetrics(t *testing.T) {
	mEe := mockEventExporter{}
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaHTTPPost
	cgrCfg.EEsCfg().Exporters[0].ID = "SQLExporterFull"
	cgrCfg.EEsCfg().Exporters[0].Synchronous = true
	newIDb := engine.NewInternalDB(nil, nil, true)
	newDM := engine.NewDataManager(newIDb, cgrCfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cgrCfg, nil, newDM)
	eeS := NewEventExporterS(cgrCfg, filterS, nil)
	eeS.eesChs = map[string]*ltcache.Cache{
		utils.MetaHTTPPost: ltcache.NewCache(1,
			time.Second, false, onCacheEvicted),
	}
	eeS.eesChs[utils.MetaHTTPPost].Set("SQLExporterFull", mEe, []string{"grp1"})
	cgrEv := &utils.CGREventWithEeIDs{
		EeIDs: []string{"SQLExporterFull"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "voiceEvent",
			Time:   utils.TimePointer(time.Now()),
			Event:  map[string]interface{}{},
			APIOpts: map[string]interface{}{
				utils.OptsEEsVerbose: struct{}{},
			},
		},
	}
	var rply map[string]map[string]interface{}
	errExpect := "cannot cast to map[string]interface{} 5 for positive exports"
	if err := eeS.V1ProcessEvent(cgrEv, &rply); err == nil || err.Error() != errExpect {
		t.Errorf("Expecting %q but received %q", errExpect, err)
	}
}
func TestV1ProcessEvent5(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters = []*config.EventExporterCfg{
		{
			Type: utils.MetaNone,
		},
		{
			ID:   "SQLExporterFull",
			Type: "invalid_type",
		},
	}
	cgrEv := &utils.CGREventWithEeIDs{
		EeIDs: []string{"SQLExporterFull"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "voiceEvent",
			Time:   utils.TimePointer(time.Now()),
			Event:  map[string]interface{}{},
			APIOpts: map[string]interface{}{
				utils.OptsEEsVerbose: struct{}{},
			},
		},
	}
	newIDb := engine.NewInternalDB(nil, nil, true)
	newDM := engine.NewDataManager(newIDb, cgrCfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cgrCfg, nil, newDM)
	eeS := NewEventExporterS(cgrCfg, filterS, nil)
	var rply map[string]map[string]interface{}
	errExpect := "unsupported exporter type: <invalid_type>"
	if err := eeS.V1ProcessEvent(cgrEv, &rply); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
}

func TestV1ProcessEvent6(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaHTTPPost
	cgrCfg.EEsCfg().Exporters[0].ID = "SQLExporterFull"
	newIDb := engine.NewInternalDB(nil, nil, true)
	newDM := engine.NewDataManager(newIDb, cgrCfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cgrCfg, nil, newDM)
	eeS := NewEventExporterS(cgrCfg, filterS, nil)
	cgrEv := &utils.CGREventWithEeIDs{
		EeIDs: []string{"SQLExporterFull"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "voiceEvent",
			Time:   utils.TimePointer(time.Now()),
			Event:  map[string]interface{}{},
			APIOpts: map[string]interface{}{
				utils.OptsEEsVerbose: struct{}{},
			},
		},
	}
	var rply map[string]map[string]interface{}
	if err := eeS.V1ProcessEvent(cgrEv, &rply); err != nil {
		t.Error(err)
	}
}

func TestOnCacheEvicted(t *testing.T) {
	var err error
	utils.Logger, err = utils.Newlogger(utils.MetaStdLog, utils.EmptyString)
	if err != nil {
		t.Error(err)
	}
	utils.Logger.SetLogLevel(7)
	bufLog := new(bytes.Buffer)
	log.SetOutput(bufLog)
	ee := mockEventExporter{}
	onCacheEvicted(utils.EmptyString, ee)
	rcvExpect := "CGRateS <> [WARNING] NOT IMPLEMENTED"
	if rcv := bufLog.String(); !strings.Contains(rcv, rcvExpect) {
		t.Errorf("Expected %v but received %v", rcvExpect, rcv)
	}
	bufLog.Reset()
}
func TestShutdown(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	newIDb := engine.NewInternalDB(nil, nil, true)
	newDM := engine.NewDataManager(newIDb, cgrCfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cgrCfg, nil, newDM)
	eeS := NewEventExporterS(cgrCfg, filterS, nil)
	logBuf := new(bytes.Buffer)
	log.SetOutput(logBuf)
	eeS.Shutdown()
	logExpect := "[INFO] <CoreS> shutdown <EventExporterS>"
	if rcv := logBuf.String(); !strings.Contains(rcv, logExpect) {
		t.Errorf("Expected %q but received %q", logExpect, rcv)
	}
	logBuf.Reset()
}

func TestUpdateEEMetrics(t *testing.T) {
	dc, _ := newEEMetrics(utils.EmptyString)
	tnow := time.Now()
	ev := engine.MapEvent{
		utils.AnswerTime: tnow,
		utils.OrderID:    1,
		utils.Cost:       5.5,
		utils.ToR:        utils.MetaVoice,
		utils.Usage:      time.Second,
	}
	exp, _ := newEEMetrics(utils.EmptyString)
	exp[utils.FirstEventATime] = tnow
	exp[utils.LastEventATime] = tnow
	exp[utils.FirstExpOrderID] = int64(1)
	exp[utils.LastExpOrderID] = int64(1)
	exp[utils.TotalCost] = float64(5.5)
	exp[utils.TotalDuration] = time.Second
	exp[utils.TimeNow] = dc[utils.TimeNow]
	if updateEEMetrics(dc, ev, utils.EmptyString); !reflect.DeepEqual(dc, exp) {
		t.Errorf("Expected: %s,received: %s", utils.ToJSON(exp), utils.ToJSON(dc))
	}

	tnow = tnow.Add(24 * time.Hour)
	ev = engine.MapEvent{
		utils.AnswerTime: tnow,
		utils.OrderID:    2,
		utils.Cost:       5.5,
		utils.ToR:        utils.MetaSMS,
		utils.Usage:      time.Second,
	}
	exp[utils.LastEventATime] = tnow
	exp[utils.LastExpOrderID] = int64(2)
	exp[utils.TotalCost] = float64(11)
	exp[utils.TotalSMSUsage] = time.Second
	if updateEEMetrics(dc, ev, utils.EmptyString); !reflect.DeepEqual(dc, exp) {
		t.Errorf("Expected: %s,received: %s", utils.ToJSON(exp), utils.ToJSON(dc))
	}

	tnow = tnow.Add(24 * time.Hour)
	ev = engine.MapEvent{
		utils.AnswerTime: tnow,
		utils.OrderID:    3,
		utils.Cost:       5.5,
		utils.ToR:        utils.MetaMMS,
		utils.Usage:      time.Second,
	}
	exp[utils.LastEventATime] = tnow
	exp[utils.LastExpOrderID] = int64(3)
	exp[utils.TotalCost] = float64(16.5)
	exp[utils.TotalMMSUsage] = time.Second
	if updateEEMetrics(dc, ev, utils.EmptyString); !reflect.DeepEqual(dc, exp) {
		t.Errorf("Expected: %s,received: %s", utils.ToJSON(exp), utils.ToJSON(dc))
	}

	tnow = tnow.Add(24 * time.Hour)
	ev = engine.MapEvent{
		utils.AnswerTime: tnow,
		utils.OrderID:    4,
		utils.Cost:       5.5,
		utils.ToR:        utils.MetaGeneric,
		utils.Usage:      time.Second,
	}
	exp[utils.LastEventATime] = tnow
	exp[utils.LastExpOrderID] = int64(4)
	exp[utils.TotalCost] = float64(22)
	exp[utils.TotalGenericUsage] = time.Second
	if updateEEMetrics(dc, ev, utils.EmptyString); !reflect.DeepEqual(dc, exp) {
		t.Errorf("Expected: %s,received: %s", utils.ToJSON(exp), utils.ToJSON(dc))
	}

	tnow = tnow.Add(24 * time.Hour)
	ev = engine.MapEvent{
		utils.AnswerTime: tnow,
		utils.OrderID:    5,
		utils.Cost:       5.5,
		utils.ToR:        utils.MetaData,
		utils.Usage:      time.Second,
	}
	exp[utils.LastEventATime] = tnow
	exp[utils.LastExpOrderID] = int64(5)
	exp[utils.TotalCost] = float64(27.5)
	exp[utils.TotalDataUsage] = time.Second
	if updateEEMetrics(dc, ev, utils.EmptyString); !reflect.DeepEqual(dc, exp) {
		t.Errorf("Expected: %s,received: %s", utils.ToJSON(exp), utils.ToJSON(dc))
	}
}
