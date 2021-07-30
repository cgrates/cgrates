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
	"encoding/csv"
	"io"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestFileFwvID(t *testing.T) {
	fFwv := &FileFWVee{
		id: "3",
	}
	if rcv := fFwv.ID(); !reflect.DeepEqual(rcv, "3") {
		t.Errorf("Expected %+v but got %+v", "3", rcv)
	}
}

func TestFileFwvGetMetrics(t *testing.T) {
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	if err != nil {
		t.Error(err)
	}
	fFwv := &FileFWVee{
		dc: dc,
	}

	if rcv := fFwv.GetMetrics(); !reflect.DeepEqual(rcv, fFwv.dc) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(rcv), utils.ToJSON(fFwv.dc))
	}
}

// type MyError struct{}

// func (m *MyError) Error() string {
// 	return "ERR"
// }

// func (nopCloser) WriteString(w io.Writer, s string) error {
// 	return &MyError{}
// }
func TestFileFwvComposeHeader(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	newIDb := engine.NewInternalDB(nil, nil, true)
	newDM := engine.NewDataManager(newIDb, cgrCfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cgrCfg, nil, newDM)
	byteBuff := new(bytes.Buffer)
	csvNW := csv.NewWriter(byteBuff)
	fFwv := &FileFWVee{
		id:      "string",
		cgrCfg:  cgrCfg,
		cfgIdx:  0,
		filterS: filterS,
		file:    nopCloser{byteBuff},
		dc:      &utils.SafeMapStorage{},
	}
	cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].Fields = []*config.FCTemplate{
		{
			Path: "*hdr.1", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("field1", utils.InfieldSep),
		},
		{
			Path: "*hdr.2", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("field2", utils.InfieldSep),
		},
	}
	for _, field := range cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].Fields {
		field.ComputePath()
	}
	if err := fFwv.composeHeader(); err != nil {
		t.Error(err)
	}
	cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].ComputeFields()
	if err := fFwv.composeHeader(); err != nil {
		t.Error(err)
	}
	csvNW.Flush()
	expected := "field1field2\n"
	if expected != byteBuff.String() {
		t.Errorf("Expected %q but received %q", expected, byteBuff.String())
	}
	cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].Fields = []*config.FCTemplate{
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
	for _, field := range cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].Fields {
		field.ComputePath()
	}
	cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].ComputeFields()
	byteBuff.Reset()
	errExpect := "inline parse error for string: <*wrong-type>"
	if err := fFwv.composeHeader(); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %q but received %q", errExpect, err)
	}
}

func TestFileFWVSyncLimit(t *testing.T) {
	//Create new exporter
	cgrCfg := config.NewDefaultCGRConfig()
	var cfgIdx int
	cfgIdx = 0

	cgrCfg.EEsCfg().Exporters[cfgIdx].Type = utils.MetaFileFWV
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

	var wg1 = &sync.WaitGroup{}

	wg1.Add(3)

	test := make(chan struct{})
	go func() {
		wg1.Wait()
		close(test)
	}()
	mckCsv := &mockCsv{
		wg: wg1,
	}
	exp := &FileFWVee{
		id:      cgrCfg.EEsCfg().Exporters[cfgIdx].ID,
		cgrCfg:  cgrCfg,
		cfgIdx:  cfgIdx,
		filterS: new(engine.FilterS),
		file:    mckCsv,
		dc:      dc,
		reqs:    newConcReq(cgrCfg.EEsCfg().Exporters[cfgIdx].ConcurrentRequests),
	}

	for i := 0; i < 3; i++ {
		go exp.ExportEvent(cgrEvent)
	}

	select {
	case <-test:
		t.Error("Should not have been possible to asynchronously export events")
	case <-time.After(4 * time.Second):
		return
	}
}

func TestFileFwvComposeTrailer(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	newIDb := engine.NewInternalDB(nil, nil, true)
	newDM := engine.NewDataManager(newIDb, cgrCfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cgrCfg, nil, newDM)
	byteBuff := new(bytes.Buffer)
	csvNW := csv.NewWriter(byteBuff)
	fFwv := &FileFWVee{
		id:      "string",
		cgrCfg:  cgrCfg,
		cfgIdx:  0,
		filterS: filterS,
		file:    nopCloser{byteBuff},
		dc:      &utils.SafeMapStorage{},
	}
	cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].Fields = []*config.FCTemplate{
		{
			Path: "*trl.1", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("field1", utils.InfieldSep),
		},
		{
			Path: "*trl.2", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("field2", utils.InfieldSep),
		},
	}
	for _, field := range cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].Fields {
		field.ComputePath()
	}
	if err := fFwv.composeTrailer(); err != nil {
		t.Error(err)
	}
	cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].ComputeFields()
	if err := fFwv.composeTrailer(); err != nil {
		t.Error(err)
	}
	csvNW.Flush()
	expected := "field1field2\n"
	if expected != byteBuff.String() {
		t.Errorf("Expected %q but received %q", expected, byteBuff.String())
	}
	cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].Fields = []*config.FCTemplate{
		{
			Path: "*trl.1", Type: utils.MetaVariable,
			Value:   config.NewRSRParsersMustCompile("field1", utils.InfieldSep),
			Filters: []string{"*wrong-type"},
		},
		{
			Path: "*trl.1", Type: utils.MetaVariable,
			Value:   config.NewRSRParsersMustCompile("field1", utils.InfieldSep),
			Filters: []string{"*wrong-type"},
		},
	}
	for _, field := range cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].Fields {
		field.ComputePath()
	}
	cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].ComputeFields()
	byteBuff.Reset()
	errExpect := "inline parse error for string: <*wrong-type>"
	if err := fFwv.composeTrailer(); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %q but received %q", errExpect, err)
	}
}

func TestFileFwvExportEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrEv := new(utils.CGREvent)
	newIDb := engine.NewInternalDB(nil, nil, true)
	newDM := engine.NewDataManager(newIDb, cgrCfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cgrCfg, nil, newDM)
	byteBuff := new(bytes.Buffer)
	csvNW := csv.NewWriter(byteBuff)
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	if err != nil {
		t.Error(err)
	}
	fFwv := &FileFWVee{
		id:      "string",
		cgrCfg:  cgrCfg,
		cfgIdx:  0,
		filterS: filterS,
		file:    nopCloser{byteBuff},
		dc:      dc,
		reqs:    newConcReq(0),
	}
	cgrEv.Event = map[string]interface{}{
		"test1": "value",
	}
	cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].Fields = []*config.FCTemplate{
		{
			Path: "*exp.1", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*req.test1", utils.InfieldSep),
		},
		{
			Path: "*exp.2", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("3", utils.InfieldSep),
		},
	}
	for _, field := range cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].Fields {
		field.ComputePath()
	}
	if err := fFwv.ExportEvent(cgrEv); err != nil {
		t.Error(err)
	}
	csvNW.Flush()
	expected := "value\n"
	if expected != byteBuff.String() {
		t.Errorf("Expected %q but received %q", expected, byteBuff.String())
	}
	cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].ComputeFields()
	byteBuff.Reset()
	if err := fFwv.ExportEvent(cgrEv); err != nil {
		t.Error(err)
	}
	csvNW.Flush()
	expected = "value3\n"
	if expected != byteBuff.String() {
		t.Errorf("Expected %q but received %q", expected, byteBuff.String())
	}
	cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].Fields = []*config.FCTemplate{
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
	for _, field := range cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].Fields {
		field.ComputePath()
	}
	cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].ComputeFields()
	byteBuff.Reset()
	errExpect := "inline parse error for string: <*wrong-type>"
	if err := fFwv.ExportEvent(cgrEv); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %q but received %q", errExpect, err)
	}
}

type nopCloserWrite struct {
	io.Writer
}

func (nopCloserWrite) Close() error { return nil }
func (nopCloserWrite) Write(s []byte) (n int, err error) {
	return 0, utils.ErrNotImplemented
}

func TestFileFwvExportEventWriteError(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrEv := new(utils.CGREvent)
	newIDb := engine.NewInternalDB(nil, nil, true)
	newDM := engine.NewDataManager(newIDb, cgrCfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cgrCfg, nil, newDM)
	byteBuff := new(bytes.Buffer)
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	if err != nil {
		t.Error(err)
	}
	fFwv := &FileFWVee{
		id:      "string",
		cgrCfg:  cgrCfg,
		cfgIdx:  0,
		filterS: filterS,
		file:    nopCloserWrite{byteBuff},
		dc:      dc,
		reqs:    newConcReq(0),
	}
	cgrEv.Event = map[string]interface{}{
		"test1": "value",
	}
	cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].Fields = []*config.FCTemplate{{}}
	for _, field := range cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].Fields {
		field.ComputePath()
	}
	cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].ComputeFields()
	if err := fFwv.ExportEvent(cgrEv); err == nil || err != utils.ErrNotImplemented {
		t.Errorf("Expected %q but received %q", utils.ErrNotImplemented, err)
	}
}

func TestFileFwvComposeHeaderWriteError(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	newIDb := engine.NewInternalDB(nil, nil, true)
	newDM := engine.NewDataManager(newIDb, cgrCfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cgrCfg, nil, newDM)
	byteBuff := new(bytes.Buffer)
	fFwv := &FileFWVee{
		id:      "string",
		cgrCfg:  cgrCfg,
		cfgIdx:  0,
		filterS: filterS,
		file:    nopCloserWrite{byteBuff},
		dc:      &utils.SafeMapStorage{},
		reqs:    newConcReq(0),
	}
	cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].Fields = []*config.FCTemplate{
		{
			Path: "*hdr.1", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("field1", utils.InfieldSep),
		},
		{
			Path: "*hdr.2", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("field2", utils.InfieldSep),
		},
	}
	for _, field := range cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].Fields {
		field.ComputePath()
	}
	cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].ComputeFields()
	if err := fFwv.composeHeader(); err == nil || err != utils.ErrNotImplemented {
		t.Errorf("Expected %q but received %q", utils.ErrNotImplemented, err)
	}
}

func TestFileFwvComposeTrailerWriteError(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	newIDb := engine.NewInternalDB(nil, nil, true)
	newDM := engine.NewDataManager(newIDb, cgrCfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cgrCfg, nil, newDM)
	byteBuff := new(bytes.Buffer)
	fFwv := &FileFWVee{
		id:      "string",
		cgrCfg:  cgrCfg,
		cfgIdx:  0,
		filterS: filterS,
		file:    nopCloserWrite{byteBuff},
		dc:      &utils.SafeMapStorage{},
		reqs:    newConcReq(0),
	}
	cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].Fields = []*config.FCTemplate{
		{
			Path: "*trl.1", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("field1", utils.InfieldSep),
		},
		{
			Path: "*trl.2", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("field2", utils.InfieldSep),
		},
	}
	for _, field := range cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].Fields {
		field.ComputePath()
	}
	cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].ComputeFields()
	if err := fFwv.composeTrailer(); err == nil || err != utils.ErrNotImplemented {
		t.Errorf("Expected %q but received %q", utils.ErrNotImplemented, err)
	}
}
func TestFileFwvOnEvictedTrailer(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	newIDb := engine.NewInternalDB(nil, nil, true)
	newDM := engine.NewDataManager(newIDb, cgrCfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cgrCfg, nil, newDM)
	byteBuff := new(bytes.Buffer)
	fFwv := &FileFWVee{
		id:      "string",
		cgrCfg:  cgrCfg,
		cfgIdx:  0,
		filterS: filterS,
		file:    nopCloserWrite{byteBuff},
		dc:      &utils.SafeMapStorage{},
		reqs:    newConcReq(0),
	}
	cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].Fields = []*config.FCTemplate{
		{
			Path: "*trl.1", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("field1", utils.InfieldSep),
		},
		{
			Path: "*trl.2", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("field2", utils.InfieldSep),
		},
	}
	for _, field := range cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].Fields {
		field.ComputePath()
	}
	cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].ComputeFields()
	fFwv.OnEvicted("test", "test")
}

type nopCloserError struct {
	io.Writer
}

func (nopCloserError) Close() error { return utils.ErrNotImplemented }
func (nopCloserError) Write(s []byte) (n int, err error) {
	return 0, utils.ErrNotImplemented
}
func TestFileFwvOnEvictedClose(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	newIDb := engine.NewInternalDB(nil, nil, true)
	newDM := engine.NewDataManager(newIDb, cgrCfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cgrCfg, nil, newDM)
	byteBuff := new(bytes.Buffer)
	fFwv := &FileFWVee{
		id:      "string",
		cgrCfg:  cgrCfg,
		cfgIdx:  0,
		filterS: filterS,
		file:    nopCloserError{byteBuff},
		dc:      &utils.SafeMapStorage{},
		reqs:    newConcReq(0),
	}
	cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].Fields = []*config.FCTemplate{
		{
			Path: "*trl.1", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("field1", utils.InfieldSep),
		},
		{
			Path: "*trl.2", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("field2", utils.InfieldSep),
		},
	}
	for _, field := range cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].Fields {
		field.ComputePath()
	}
	cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].ComputeFields()
	fFwv.OnEvicted("test", "test")
}

func TestFileFWVSync(t *testing.T) {
	//Create new exporter
	cgrCfg := config.NewDefaultCGRConfig()
	var cfgIdx int
	cfgIdx = 0

	cgrCfg.EEsCfg().Exporters[cfgIdx].Type = utils.MetaFileFWV
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

	var wg1 = &sync.WaitGroup{}

	wg1.Add(3)

	test := make(chan struct{})
	go func() {
		wg1.Wait()
		close(test)
	}()
	mckCsv := &mockCsv{
		wg: wg1,
	}
	exp := &FileFWVee{
		id:      cgrCfg.EEsCfg().Exporters[cfgIdx].ID,
		cgrCfg:  cgrCfg,
		cfgIdx:  cfgIdx,
		filterS: new(engine.FilterS),
		file:    mckCsv,
		dc:      dc,
		reqs:    newConcReq(cgrCfg.EEsCfg().Exporters[cfgIdx].ConcurrentRequests),
	}

	for i := 0; i < 3; i++ {
		go exp.ExportEvent(cgrEvent)
	}

	select {
	case <-test:
		return
	case <-time.After(4 * time.Second):
		t.Error("Can't asynchronously export events")
	}
}
