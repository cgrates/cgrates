/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestHttpPostID(t *testing.T) {
	httpPost := &HTTPPostEE{
		id: "3",
	}
	if rcv := httpPost.ID(); !reflect.DeepEqual(rcv, "3") {
		t.Errorf("Expected %+v but got %+v", "3", rcv)
	}
}

func TestHttpPostGetMetrics(t *testing.T) {
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	if err != nil {
		t.Error(err)
	}
	httpPost := &HTTPPostEE{
		dc: dc,
	}

	if rcv := httpPost.GetMetrics(); !reflect.DeepEqual(rcv, httpPost.dc) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(rcv), utils.ToJSON(httpPost.dc))
	}
}

func TestHttpPostExportEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaHTTPPost
	cgrEv := new(utils.CGREvent)
	newIDb := engine.NewInternalDB(nil, nil, true)
	newDM := engine.NewDataManager(newIDb, cgrCfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cgrCfg, nil, newDM)
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	if err != nil {
		t.Error(err)
	}
	httpPost, err := NewHTTPPostEE(cgrCfg, 0, filterS, dc)
	if err != nil {
		t.Error(err)
	}
	cgrEv.Event = map[string]interface{}{
		"Test1": 3,
	}
	errExpect := `Post "/var/spool/cgrates/ees": unsupported protocol scheme ""`
	if err := httpPost.ExportEvent(cgrEv); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %q but received %q", errExpect, err)
	}
	dcExpect := int64(1)
	if !reflect.DeepEqual(dcExpect, httpPost.dc.MapStorage[utils.NumberOfEvents]) {
		t.Errorf("Expected %q but received %q", dcExpect, httpPost.dc.MapStorage[utils.NumberOfEvents])
	}
}

func TestHttpPostExportEvent2(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaHTTPPost
	cgrEv := new(utils.CGREvent)
	newIDb := engine.NewInternalDB(nil, nil, true)
	newDM := engine.NewDataManager(newIDb, cgrCfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cgrCfg, nil, newDM)
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	if err != nil {
		t.Error(err)
	}
	bodyExpect := "2=%2Areq.field2"
	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}
		if strBody := string(body); strBody != bodyExpect {
			t.Errorf("Expected %q but received %q", bodyExpect, strBody)
		}
		rw.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	cgrCfg.EEsCfg().Exporters[0].ExportPath = srv.URL + "/"
	httpPost, err := NewHTTPPostEE(cgrCfg, 0, filterS, dc)
	if err != nil {
		t.Error(err)
	}
	cgrEv.Event = map[string]interface{}{
		"test": "string",
	}
	cgrCfg.EEsCfg().Exporters[0].Fields = []*config.FCTemplate{
		{
			Path: "*exp.1", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*req.field1", utils.InfieldSep),
		},
		{
			Path: "*exp.2", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("*req.field2", utils.InfieldSep),
		},
	}
	for _, field := range cgrCfg.EEsCfg().Exporters[0].Fields {
		field.ComputePath()
	}
	cgrCfg.EEsCfg().Exporters[0].ComputeFields()
	if err := httpPost.ExportEvent(cgrEv); err != nil {
		t.Error(err)
	}
	dcExpect := int64(1)
	if !reflect.DeepEqual(dcExpect, httpPost.dc.MapStorage[utils.NumberOfEvents]) {
		t.Errorf("Expected %q but received %q", dcExpect, httpPost.dc.MapStorage[utils.NumberOfEvents])
	}
}

func TestHttpPostExportEvent3(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaHTTPPost
	cgrEv := new(utils.CGREvent)
	newIDb := engine.NewInternalDB(nil, nil, true)
	newDM := engine.NewDataManager(newIDb, cgrCfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cgrCfg, nil, newDM)
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	if err != nil {
		t.Error(err)
	}
	httpPost, err := NewHTTPPostEE(cgrCfg, 0, filterS, dc)
	if err != nil {
		t.Error(err)
	}
	cgrEv.Event = map[string]interface{}{
		"Test1": 3,
	}
	cgrCfg.EEsCfg().Exporters[0].Fields = []*config.FCTemplate{
		{
			Path: "*exp.1", Type: utils.MetaVariable,
			Value:   config.NewRSRParsersMustCompile("~*req.field1", utils.InfieldSep),
			Filters: []string{"*wrong-type"},
		},
		{
			Path: "*exp.1", Type: utils.MetaVariable,
			Value:   config.NewRSRParsersMustCompile("~*req.field1", utils.InfieldSep),
			Filters: []string{"*wrong-type"},
		},
	}
	for _, field := range cgrCfg.EEsCfg().Exporters[0].Fields {
		field.ComputePath()
	}
	cgrCfg.EEsCfg().Exporters[0].ComputeFields()
	errExpect := "inline parse error for string: <*wrong-type>"
	if err := httpPost.ExportEvent(cgrEv); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %q but received %q", errExpect, err)
	}
	dcExpect := int64(1)
	if !reflect.DeepEqual(dcExpect, httpPost.dc.MapStorage[utils.NumberOfEvents]) {
		t.Errorf("Expected %q but received %q", dcExpect, httpPost.dc.MapStorage[utils.NumberOfEvents])
	}
}

func TestHttpPostExportEvent4(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaHTTPPost
	cgrEv := new(utils.CGREvent)
	newIDb := engine.NewInternalDB(nil, nil, true)
	newDM := engine.NewDataManager(newIDb, cgrCfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cgrCfg, nil, newDM)
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	if err != nil {
		t.Error(err)
	}
	httpPost, err := NewHTTPPostEE(cgrCfg, 0, filterS, dc)
	if err != nil {
		t.Error(err)
	}
	cgrEv.Event = map[string]interface{}{
		"Test1": 3,
	}
	cgrCfg.EEsCfg().Exporters[0].Fields = []*config.FCTemplate{
		{
			Path: "*exp.1", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*req.field1", utils.InfieldSep),
		},
		{
			Path: "*exp.2", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*req.field2", utils.InfieldSep),
		},
		{
			Path: "*hdr.1", Type: utils.MetaVariable,
			Value:   config.NewRSRParsersMustCompile("~*req.field2", utils.InfieldSep),
			Filters: []string{"*wrong-type"},
		},
	}
	for _, field := range cgrCfg.EEsCfg().Exporters[0].Fields {
		field.ComputePath()
	}
	cgrCfg.EEsCfg().Exporters[0].ComputeFields()
	errExpect := "inline parse error for string: <*wrong-type>"
	if err := httpPost.ExportEvent(cgrEv); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %q but received %q", errExpect, err)
	}
	dcExpect := int64(1)
	if !reflect.DeepEqual(dcExpect, httpPost.dc.MapStorage[utils.NumberOfEvents]) {
		t.Errorf("Expected %q but received %q", dcExpect, httpPost.dc.MapStorage[utils.NumberOfEvents])
	}
	httpPost.OnEvicted("test", "test")
}

func TestHttpPostComposeHeader(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaHTTPPost
	newIDb := engine.NewInternalDB(nil, nil, true)
	newDM := engine.NewDataManager(newIDb, cgrCfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cgrCfg, nil, newDM)
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	if err != nil {
		t.Error(err)
	}
	httpPost, err := NewHTTPPostEE(cgrCfg, 0, filterS, dc)
	if err != nil {
		t.Error(err)
	}
	cgrCfg.EEsCfg().Exporters[0].Fields = []*config.FCTemplate{
		{
			Path: "*hdr.1", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("field1", utils.InfieldSep),
		},
		{
			Path: "*hdr.2", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("field2", utils.InfieldSep),
		},
	}
	for _, field := range cgrCfg.EEsCfg().Exporters[0].Fields {
		field.ComputePath()
	}
	if _, err := httpPost.composeHeader(); err != nil {
		t.Error(err)
	}
	cgrCfg.EEsCfg().Exporters[0].ComputeFields()
	if _, err := httpPost.composeHeader(); err != nil {
		t.Error(err)
	}
	cgrCfg.EEsCfg().Exporters[0].Fields = []*config.FCTemplate{
		{
			Path: "*hdr.1", Type: utils.MetaVariable,
			Value:   config.NewRSRParsersMustCompile("field1", utils.InfieldSep),
			Filters: []string{"*wrong-type"},
		},
		{
			Path: "*hdr.1", Type: utils.MetaVariable,
			Value:   config.NewRSRParsersMustCompile("field1", utils.InfieldSep),
			Filters: []string{"*wrong-type"},
		},
	}
	for _, field := range cgrCfg.EEsCfg().Exporters[0].Fields {
		field.ComputePath()
	}
	cgrCfg.EEsCfg().Exporters[0].ComputeFields()
	errExpect := "inline parse error for string: <*wrong-type>"
	if _, err := httpPost.composeHeader(); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %q but received %q", errExpect, err)
	}
}

func TestHttpPostSync(t *testing.T) {
	//Create new exporter
	cgrCfg := config.NewDefaultCGRConfig()
	var cfgIdx int
	cfgIdx = 0

	cgrCfg.EEsCfg().Exporters[cfgIdx].Type = utils.MetaHTTPPost
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		cgrCfg.EEsCfg().Exporters[cfgIdx].Timezone,
		cgrCfg.GeneralCfg().DefaultTimezone))
	if err != nil {
		t.Error(err)
	}

	//Create an event
	cgrEvent := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			"Account":     "1001",
			"Destination": "1002",
		},
	}
	var wg1 sync.WaitGroup

	wg1.Add(3)

	test := make(chan struct{})
	go func() {
		wg1.Wait()
		close(test)
	}()

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		time.Sleep(25 * time.Millisecond)
		wg1.Done()
	}))

	defer ts.Close()

	cgrCfg.EEsCfg().Exporters[cfgIdx].ExportPath = ts.URL

	exp, err := NewHTTPPostEE(cgrCfg, cfgIdx, new(engine.FilterS), dc)
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < 3; i++ {
		go exp.ExportEvent(cgrEvent)
	}

	select {
	case <-test:
		return
	case <-time.After(50 * time.Millisecond):
		t.Error("Can't asynchronously export events")
	}
}

func TestHttpPostSyncLimit(t *testing.T) {
	//Create new exporter
	cgrCfg := config.NewDefaultCGRConfig()
	var cfgIdx int
	cfgIdx = 0

	cgrCfg.EEsCfg().Exporters[cfgIdx].Type = utils.MetaHTTPPost
	cgrCfg.EEsCfg().Exporters[cfgIdx].ConcurrentRequests = 1
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		cgrCfg.EEsCfg().Exporters[cfgIdx].Timezone,
		cgrCfg.GeneralCfg().DefaultTimezone))
	if err != nil {
		t.Error(err)
	}

	//Create an event
	cgrEvent := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			"Account":     "1001",
			"Destination": "1002",
		},
	}
	var wg1 sync.WaitGroup

	wg1.Add(3)

	test := make(chan struct{})
	go func() {
		wg1.Wait()
		close(test)
	}()

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		time.Sleep(25 * time.Millisecond)
		wg1.Done()
	}))

	defer ts.Close()

	cgrCfg.EEsCfg().Exporters[cfgIdx].ExportPath = ts.URL

	exp, err := NewHTTPPostEE(cgrCfg, cfgIdx, new(engine.FilterS), dc)
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < 3; i++ {
		go exp.ExportEvent(cgrEvent)
	}

	select {
	case <-test:
		t.Error("Should not have been possible to asynchronously export events")
	case <-time.After(50 * time.Millisecond):
		return
	}
}
