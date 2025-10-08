/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package ees

import (
	"bytes"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/attributes"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
	"github.com/cgrates/rpcclient"
)

type testMockEvent struct {
	calls map[string]func(_ *context.Context, _, _ any) error
}

func (sT *testMockEvent) Call(ctx *context.Context, method string, arg, rply any) error {
	if call, has := sT.calls[method]; !has {
		return rpcclient.ErrUnsupporteServiceMethod
	} else {
		return call(ctx, arg, rply)
	}
}
func TestAttrSProcessEvent(t *testing.T) {
	testMock := &testMockEvent{
		calls: map[string]func(_ *context.Context, _, _ any) error{
			utils.AttributeSv1ProcessEvent: func(_ *context.Context, args, reply any) error {
				rplyEv := &attributes.AttrSProcessEventReply{
					AlteredFields: []*attributes.FieldsAltered{{
						Fields: []string{"testcase"},
					}},
				}
				*reply.(*attributes.AttrSProcessEventReply) = *rplyEv
				return nil
			},
		},
	}
	cgrEv := &utils.CGREvent{
		APIOpts: map[string]any{
			utils.OptsAttributesProcessRuns: "10",
		},
	}
	cfg := config.NewDefaultCGRConfig()
	cfg.EEsNoLksCfg().AttributeSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}
	newIDb, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	newDM := engine.NewDataManager(newIDb, cfg, nil)
	filterS := engine.NewFilterS(cfg, nil, newDM)
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- testMock
	connMgr := engine.NewConnManager(cfg)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), utils.AttributeSv1, clientConn)
	eeS, err := NewEventExporterS(cfg, filterS, connMgr)
	if err != nil {
		t.Fatal(err)
	}
	if err := eeS.attrSProcessEvent(context.TODO(), cgrEv, []string{}, utils.EmptyString); err != nil {
		t.Error(err)
	}
}

func TestAttrSProcessEvent2(t *testing.T) {
	engine.Cache.Clear(nil)
	testMock := &testMockEvent{
		calls: map[string]func(_ *context.Context, _, _ any) error{
			utils.AttributeSv1ProcessEvent: func(_ *context.Context, _, _ any) error {
				return utils.ErrNotFound
			},
		},
	}
	cfg := config.NewDefaultCGRConfig()
	cfg.EEsNoLksCfg().AttributeSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}
	newIDb, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	newDM := engine.NewDataManager(newIDb, cfg, nil)
	filterS := engine.NewFilterS(cfg, nil, newDM)
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- testMock
	connMgr := engine.NewConnManager(cfg)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), utils.AttributeSv1, clientConn)
	eeS, err := NewEventExporterS(cfg, filterS, connMgr)
	if err != nil {
		t.Fatal(err)
	}
	cgrEv := &utils.CGREvent{}
	if err := eeS.attrSProcessEvent(context.TODO(), cgrEv, []string{}, utils.EmptyString); err != nil {
		t.Error(err)
	}
}

func TestV1ProcessEvent(t *testing.T) {
	filePath := "/tmp/TestV1ProcessEvent"
	if err := os.MkdirAll(filePath, 0777); err != nil {
		t.Error(err)
	}
	cfg := config.NewDefaultCGRConfig()
	cfg.EEsCfg().Exporters[0].Type = "*fileCSV"
	cfg.EEsCfg().Exporters[0].ID = "SQLExporterFull"
	cfg.EEsCfg().Exporters[0].ExportPath = filePath
	newIDb, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	newDM := engine.NewDataManager(newIDb, cfg, nil)
	filterS := engine.NewFilterS(cfg, nil, newDM)
	eeS, err := NewEventExporterS(cfg, filterS, nil)
	if err != nil {
		t.Fatal(err)
	}
	cgrEv := &utils.CGREventWithEeIDs{
		EeIDs: []string{"SQLExporterFull"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "voiceEvent",
			Event: map[string]any{

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
				utils.MetaRunID:    utils.MetaDefault,
				utils.Cost:         1.01,
				"ExtraFields": map[string]string{"extra1": "val_extra1",
					"extra2": "val_extra2", "extra3": "val_extra3"},
			},
			APIOpts: map[string]any{
				utils.MetaOriginID: utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()),
				utils.MetaRunID:    utils.MetaDefault,
			},
		},
	}
	var rply map[string]map[string]any
	rplyExpect := map[string]map[string]any{
		"SQLExporterFull": {},
	}
	if err := eeS.V1ProcessEvent(context.TODO(), cgrEv, &rply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rply, rplyExpect) {
		t.Errorf("Expected %q but received %q", rplyExpect, rply)
	}
	if err := os.RemoveAll(filePath); err != nil {
		t.Error(err)
	}
}

func TestV1ProcessEvent2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.EEsCfg().Exporters[0].Type = "*fileCSV"
	cfg.EEsCfg().Exporters[0].ID = "SQLExporterFull"
	cfg.EEsCfg().Exporters[0].Filters = []string{"*prefix:~*req.Subject:20"}
	newIDb, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	newDM := engine.NewDataManager(newIDb, cfg, nil)
	filterS := engine.NewFilterS(cfg, nil, newDM)
	eeS, err := NewEventExporterS(cfg, filterS, nil)
	if err != nil {
		t.Fatal(err)
	}
	cgrEv := &utils.CGREventWithEeIDs{
		EeIDs: []string{"SQLExporterFull"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "voiceEvent",
			Event: map[string]any{
				utils.Subject: "1001",
				"ExtraFields": map[string]string{"extra1": "val_extra1",
					"extra2": "val_extra2", "extra3": "val_extra3"},
			},
			APIOpts: map[string]any{
				utils.OptsAttributesProcessRuns: "10",
			},
		},
	}
	var rply map[string]map[string]any
	errExpect := "NOT_FOUND"
	if err := eeS.V1ProcessEvent(context.TODO(), cgrEv, &rply); err == nil || err.Error() != errExpect {
		t.Errorf("Expecting %q but received %q", errExpect, err)
	}

	errExpect = "NOT_FOUND:test"
	eeS.cfg.EEsCfg().Exporters[0].Filters = []string{"test"}
	if err := eeS.V1ProcessEvent(context.TODO(), cgrEv, &rply); err == nil || err.Error() != errExpect {
		t.Errorf("Expecting %q but received %q", errExpect, err)
	}
}

func TestV1ProcessEvent3(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.EEsCfg().Exporters[0].Type = "*fileCSV"
	cfg.EEsCfg().Exporters[0].ID = "SQLExporterFull"
	cfg.EEsCfg().Exporters[0].Flags = utils.FlagsWithParams{
		utils.MetaAttributes: utils.FlagParams{},
	}
	newIDb, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	newDM := engine.NewDataManager(newIDb, cfg, nil)
	filterS := engine.NewFilterS(cfg, nil, newDM)
	eeS, err := NewEventExporterS(cfg, filterS, nil)
	if err != nil {
		t.Fatal(err)
	}
	cgrEv := &utils.CGREventWithEeIDs{
		EeIDs: []string{"SQLExporterFull"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "voiceEvent",
			Event:  map[string]any{},
		},
	}
	var rply map[string]map[string]any
	errExpect := "MANDATORY_IE_MISSING: [connIDs]"
	if err := eeS.V1ProcessEvent(context.TODO(), cgrEv, &rply); err == nil || err.Error() != errExpect {
		t.Errorf("Expecting %q but received %q", errExpect, err)
	}
}

func TestV1ProcessEvent4(t *testing.T) {
	testMock := &testMockEvent{
		calls: map[string]func(_ *context.Context, _, _ any) error{
			utils.EfSv1ProcessEvent: func(_ *context.Context, args, reply any) error {
				*reply.(*string) = utils.EmptyString
				return utils.ErrUnsupportedFormat
			},
		},
	}
	cfg := config.NewDefaultCGRConfig()
	cfg.EFsCfg().Enabled = true
	cfg.EEsCfg().Exporters[0].Type = utils.MetaHTTPPost
	cfg.EEsCfg().Exporters[0].ID = "SQLExporterFull"
	cfg.EEsCfg().Exporters[0].Synchronous = true
	newIDb, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	newDM := engine.NewDataManager(newIDb, cfg, nil)
	filterS := engine.NewFilterS(cfg, nil, newDM)
	connMngr := engine.NewConnManager(cfg)
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- testMock
	connMngr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEFs), utils.EfSv1, clientConn)
	eeS, err := NewEventExporterS(cfg, filterS, connMngr)
	if err != nil {
		t.Fatal(err)
	}
	eeS.exporterCache = map[string]*ltcache.Cache{
		utils.MetaHTTPPost: ltcache.NewCache(1,
			time.Second, false, false, []func(itmID string, value any){
				onCacheEvicted,
			}),
	}
	newEeS, err := NewEventExporter(cfg.EEsCfg().Exporters[0], cfg, filterS, connMngr)
	if err != nil {
		t.Error(err)
	}
	eeS.exporterCache[utils.MetaHTTPPost].Set("SQLExporterFull", newEeS, []string{"grp1"})
	cgrEv := &utils.CGREventWithEeIDs{
		EeIDs: []string{"SQLExporterFull"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "voiceEvent",
			Event:  map[string]any{},
			APIOpts: map[string]any{
				utils.OptsEEsVerbose: struct{}{},
			},
		},
	}
	var rply map[string]map[string]any
	errExpect := "PARTIALLY_EXECUTED"
	if err := eeS.V1ProcessEvent(context.TODO(), cgrEv, &rply); err == nil || err.Error() != errExpect {
		t.Errorf("Expecting %q but received %q", errExpect, err)
	} else if len(rply) != 0 {
		t.Error("Unexpected reply result")
	}
}

func newMockEventExporter() *mockEventExporter {
	return &mockEventExporter{dc: &utils.SafeMapStorage{
		MapStorage: utils.MapStorage{
			utils.NumberOfEvents:  int64(0),
			utils.PositiveExports: utils.StringSet{},
			utils.NegativeExports: 5,
		}}}
}

type mockEventExporter struct {
	dc *utils.SafeMapStorage
	bytePreparing
}

func (m mockEventExporter) GetMetrics() *utils.SafeMapStorage {
	return m.dc
}

func (mockEventExporter) Cfg() *config.EventExporterCfg                { return new(config.EventExporterCfg) }
func (mockEventExporter) Connect() error                               { return nil }
func (mockEventExporter) ExportEvent(*context.Context, any, any) error { return nil }
func (mockEventExporter) ExtraData(*utils.CGREvent) any                { return nil }
func (mockEventExporter) Close() error {
	utils.Logger.Warning("NOT IMPLEMENTED")
	return nil
}

func TestV1ProcessEventMockMetrics(t *testing.T) {
	mEe := mockEventExporter{dc: &utils.SafeMapStorage{
		MapStorage: utils.MapStorage{
			utils.NumberOfEvents:  int64(0),
			utils.PositiveExports: utils.StringSet{},
			utils.NegativeExports: 5,
		}}}
	cfg := config.NewDefaultCGRConfig()
	cfg.EEsCfg().Exporters[0].Type = utils.MetaHTTPPost
	cfg.EEsCfg().Exporters[0].ID = "SQLExporterFull"
	cfg.EEsCfg().Exporters[0].Synchronous = true
	newIDb, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	newDM := engine.NewDataManager(newIDb, cfg, nil)
	filterS := engine.NewFilterS(cfg, nil, newDM)
	eeS, err := NewEventExporterS(cfg, filterS, nil)
	if err != nil {
		t.Fatal(err)
	}
	eeS.exporterCache = map[string]*ltcache.Cache{
		utils.MetaHTTPPost: ltcache.NewCache(1,
			time.Second, false, false, []func(itmID string, value any){
				onCacheEvicted,
			}),
	}
	eeS.exporterCache[utils.MetaHTTPPost].Set("SQLExporterFull", mEe, []string{"grp1"})
	cgrEv := &utils.CGREventWithEeIDs{
		EeIDs: []string{"SQLExporterFull"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "voiceEvent",
			Event:  map[string]any{},
			APIOpts: map[string]any{
				utils.OptsEEsVerbose: struct{}{},
			},
		},
	}
	var rply map[string]map[string]any
	errExpect := "cannot cast to map[string]any 5 for positive exports"
	if err := eeS.V1ProcessEvent(context.TODO(), cgrEv, &rply); err == nil || err.Error() != errExpect {
		t.Errorf("Expecting %q but received %q", errExpect, err)
	}
}
func TestV1ProcessEvent5(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.EEsCfg().Exporters = []*config.EventExporterCfg{
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
			Event:  map[string]any{},
			APIOpts: map[string]any{
				utils.OptsEEsVerbose: struct{}{},
			},
		},
	}
	newIDb, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	newDM := engine.NewDataManager(newIDb, cfg, nil)
	filterS := engine.NewFilterS(cfg, nil, newDM)
	eeS, err := NewEventExporterS(cfg, filterS, nil)
	if err != nil {
		t.Fatal(err)
	}
	var rply map[string]map[string]any
	errExpect := `failed to init EventExporter "SQLExporterFull": unsupported exporter type: <invalid_type>`
	if err := eeS.V1ProcessEvent(context.TODO(), cgrEv, &rply); err == nil || err.Error() != errExpect {
		t.Errorf("expected %q, received %q", errExpect, err)
	}
}

func TestV1ProcessEvent6(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.EEsCfg().Exporters[0].Type = utils.MetaHTTPPost
	cfg.EEsCfg().Exporters[0].ID = "SQLExporterFull"
	newIDb, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	newDM := engine.NewDataManager(newIDb, cfg, nil)
	filterS := engine.NewFilterS(cfg, nil, newDM)
	eeS, err := NewEventExporterS(cfg, filterS, nil)
	if err != nil {
		t.Fatal(err)
	}
	cgrEv := &utils.CGREventWithEeIDs{
		EeIDs: []string{"SQLExporterFull"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "voiceEvent",
			Event:  map[string]any{},
			APIOpts: map[string]any{
				utils.OptsEEsVerbose: struct{}{},
			},
		},
	}
	var rply map[string]map[string]any
	if err := eeS.V1ProcessEvent(context.TODO(), cgrEv, &rply); err != nil {
		t.Error(err)
	}
}

func TestOnCacheEvicted(t *testing.T) {
	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
	}()
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 7)
	ee := newMockEventExporter()
	onCacheEvicted(utils.EmptyString, ee)
	rcvExpect := "CGRateS <> [WARNING] NOT IMPLEMENTED"
	if rcv := buf.String(); !strings.Contains(rcv, rcvExpect) {
		t.Errorf("Expected %v but received %v", rcvExpect, rcv)
	}
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
	exp.MapStorage[utils.FirstEventATime] = tnow
	exp.MapStorage[utils.LastEventATime] = tnow
	exp.MapStorage[utils.FirstExpOrderID] = int64(1)
	exp.MapStorage[utils.LastExpOrderID] = int64(1)
	exp.MapStorage[utils.TotalCost] = float64(5.5)
	exp.MapStorage[utils.TotalDuration] = time.Second
	exp.MapStorage[utils.TimeNow] = dc.MapStorage[utils.TimeNow]
	exp.MapStorage[utils.PositiveExports] = utils.StringSet{"": {}}
	if updateEEMetrics(dc, "", ev, false, utils.EmptyString); !reflect.DeepEqual(dc, exp) {
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
	exp.MapStorage[utils.LastEventATime] = tnow
	exp.MapStorage[utils.LastExpOrderID] = int64(2)
	exp.MapStorage[utils.TotalCost] = float64(11)
	exp.MapStorage[utils.TotalSMSUsage] = time.Second
	if updateEEMetrics(dc, "", ev, false, utils.EmptyString); !reflect.DeepEqual(dc, exp) {
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
	exp.MapStorage[utils.LastEventATime] = tnow
	exp.MapStorage[utils.LastExpOrderID] = int64(3)
	exp.MapStorage[utils.TotalCost] = float64(16.5)
	exp.MapStorage[utils.TotalMMSUsage] = time.Second
	if updateEEMetrics(dc, "", ev, false, utils.EmptyString); !reflect.DeepEqual(dc, exp) {
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
	exp.MapStorage[utils.LastEventATime] = tnow
	exp.MapStorage[utils.LastExpOrderID] = int64(4)
	exp.MapStorage[utils.TotalCost] = float64(22)
	exp.MapStorage[utils.TotalGenericUsage] = time.Second
	if updateEEMetrics(dc, "", ev, false, utils.EmptyString); !reflect.DeepEqual(dc, exp) {
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
	exp.MapStorage[utils.LastEventATime] = tnow
	exp.MapStorage[utils.LastExpOrderID] = int64(5)
	exp.MapStorage[utils.TotalCost] = float64(27.5)
	exp.MapStorage[utils.TotalDataUsage] = time.Second
	if updateEEMetrics(dc, "", ev, false, utils.EmptyString); !reflect.DeepEqual(dc, exp) {
		t.Errorf("Expected: %s,received: %s", utils.ToJSON(exp), utils.ToJSON(dc))
	}
}

func TestEeSProcessEvent(t *testing.T) {
	filePath := "/tmp/TestV1ProcessEvent"
	if err := os.MkdirAll(filePath, 0777); err != nil {
		t.Error(err)
	}
	cfg := config.NewDefaultCGRConfig()
	cfg.EEsCfg().Exporters[0].Type = "*fileCSV"
	cfg.EEsCfg().Exporters[0].ID = "SQLExporterFull"
	cfg.EEsCfg().Exporters[0].ExportPath = filePath
	newIDb, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	newDM := engine.NewDataManager(newIDb, cfg, nil)
	filterS := engine.NewFilterS(cfg, nil, newDM)
	eeS, err := NewEventExporterS(cfg, filterS, nil)
	if err != nil {
		t.Fatal(err)
	}
	cgrEv := &utils.CGREventWithEeIDs{
		EeIDs: []string{"SQLExporterFull"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "voiceEvent",
			Event: map[string]any{

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
				utils.MetaRunID:    utils.MetaDefault,
				utils.Cost:         1.01,
				"ExtraFields": map[string]string{"extra1": "val_extra1",
					"extra2": "val_extra2", "extra3": "val_extra3"},
			},
			APIOpts: map[string]any{
				utils.MetaOriginID: utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()),
			},
		},
	}
	var reply map[string]map[string]any
	replyExpect := map[string]map[string]any{
		"SQLExporterFull": {},
	}
	if err := eeS.V1ProcessEvent(context.Background(), cgrEv, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, replyExpect) {
		t.Errorf("Expected %v \n but received \n %v", replyExpect, reply)
	}

	if err := os.RemoveAll(filePath); err != nil {
		t.Error(err)
	}
}

func TestArchiveEventsInReply(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	// cfg.EEsCfg().Exporters[0].Type = "*fileCSV"
	cfg.EEsCfg().Exporters[0].ID = "SQLExporterFull"
	newIDb, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	newDM := engine.NewDataManager(newIDb, cfg, nil)
	filterS := engine.NewFilterS(cfg, nil, newDM)
	eeS, err := NewEventExporterS(cfg, filterS, nil)
	if err != nil {
		t.Fatal(err)
	}

	args := &ArchiveEventsArgs{
		Tenant: "cgrates.org",
		APIOpts: map[string]any{
			utils.MetaExporterID: "SQLExporterFull",
		},
		Events: []*utils.EventsWithOpts{
			{
				Event: map[string]any{
					"Account": "1001",
				},
			},
		},
	}

	var reply []byte
	errExp := "exporter with ID: SQLExporterFull is not synchronous"
	if err := eeS.V1ArchiveEventsInReply(context.Background(), args, &reply); err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}
}
