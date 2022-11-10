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
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestNewEventExporter(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaFileCSV
	cgrCfg.EEsCfg().Exporters[0].ConcurrentRequests = 0
	filterS := engine.NewFilterS(cgrCfg, nil, nil)
	ee, err := NewEventExporter(cgrCfg.EEsCfg().Exporters[0], cgrCfg, filterS, nil)
	errExpect := "open /var/spool/cgrates/ees/*default_"
	if strings.Contains(errExpect, err.Error()) {
		t.Errorf("Expected %+v but got %+v", errExpect, err)
	}
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	if err != nil {
		t.Error(err)
	}
	eeExpect, err := NewFileCSVee(cgrCfg.EEsCfg().Exporters[0], cgrCfg, filterS, dc)
	if strings.Contains(errExpect, err.Error()) {
		t.Errorf("Expected %+v but got %+v", errExpect, err)
	}
	err = eeExpect.init()
	newEE := ee.(*FileCSVee)
	newEE.dc.MapStorage[utils.TimeNow] = nil
	newEE.dc.MapStorage[utils.ExportPath] = nil
	eeExpect.csvWriter = nil
	eeExpect.dc.MapStorage[utils.TimeNow] = nil
	eeExpect.dc.MapStorage[utils.ExportPath] = nil
	if !reflect.DeepEqual(eeExpect, newEE) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(eeExpect), utils.ToJSON(newEE))
	}
}

func TestNewEventExporterCase2(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaFileFWV
	cgrCfg.EEsCfg().Exporters[0].ConcurrentRequests = 0
	filterS := engine.NewFilterS(cgrCfg, nil, nil)
	ee, err := NewEventExporter(cgrCfg.EEsCfg().Exporters[0], cgrCfg, filterS, nil)
	errExpect := "open /var/spool/cgrates/ees/*default_"
	if strings.Contains(errExpect, err.Error()) {
		t.Errorf("Expected %+v but got %+v", errExpect, err)
	}

	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	eeExpect, err := NewFileFWVee(cgrCfg.EEsCfg().Exporters[0], cgrCfg, filterS, dc)
	if strings.Contains(errExpect, err.Error()) {
		t.Errorf("Expected %+v but got %+v", errExpect, err)
	}
	err = eeExpect.init()
	newEE := ee.(*FileFWVee)
	newEE.dc.MapStorage[utils.TimeNow] = nil
	newEE.dc.MapStorage[utils.ExportPath] = nil
	eeExpect.dc.MapStorage[utils.TimeNow] = nil
	eeExpect.dc.MapStorage[utils.ExportPath] = nil
	if !reflect.DeepEqual(eeExpect, newEE) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(eeExpect), utils.ToJSON(newEE))
	}
}

func TestNewEventExporterCase3(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaHTTPPost
	cgrCfg.EEsCfg().Exporters[0].ConcurrentRequests = 0
	filterS := engine.NewFilterS(cgrCfg, nil, nil)
	ee, err := NewEventExporter(cgrCfg.EEsCfg().Exporters[0], cgrCfg, filterS, nil)
	if err != nil {
		t.Error(err)
	}
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	eeExpect, err := NewHTTPPostEE(cgrCfg.EEsCfg().Exporters[0], cgrCfg, filterS, dc)
	if err != nil {
		t.Error(err)
	}
	newEE := ee.(*HTTPPostEE)
	newEE.dc.MapStorage[utils.TimeNow] = nil
	eeExpect.dc.MapStorage[utils.TimeNow] = nil
	if !reflect.DeepEqual(eeExpect, newEE) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(eeExpect), utils.ToJSON(newEE))
	}
}

func TestNewEventExporterCase4(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaHTTPjsonMap
	cgrCfg.EEsCfg().Exporters[0].ConcurrentRequests = 0
	filterS := engine.NewFilterS(cgrCfg, nil, nil)
	ee, err := NewEventExporter(cgrCfg.EEsCfg().Exporters[0], cgrCfg, filterS, nil)
	if err != nil {
		t.Error(err)
	}
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	eeExpect, err := NewHTTPjsonMapEE(cgrCfg.EEsCfg().Exporters[0], cgrCfg, filterS, dc)
	if err != nil {
		t.Error(err)
	}
	newEE := ee.(*HTTPjsonMapEE)
	newEE.dc.MapStorage[utils.TimeNow] = nil
	eeExpect.dc.MapStorage[utils.TimeNow] = nil
	if !reflect.DeepEqual(eeExpect, newEE) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(eeExpect), utils.ToJSON(newEE))
	}
}

func TestNewEventExporterCase6(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaVirt
	cgrCfg.EEsCfg().Exporters[0].ConcurrentRequests = 0
	filterS := engine.NewFilterS(cgrCfg, nil, nil)
	ee, err := NewEventExporter(cgrCfg.EEsCfg().Exporters[0], cgrCfg, filterS, nil)
	if err != nil {
		t.Error(err)
	}
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	if err != nil {
		t.Error(err)
	}
	eeExpect := NewVirtualEE(cgrCfg.EEsCfg().Exporters[0], dc)
	newEE := ee.(*VirtualEE)
	newEE.dc.MapStorage[utils.TimeNow] = nil
	eeExpect.dc.MapStorage[utils.TimeNow] = nil
	if !reflect.DeepEqual(eeExpect, newEE) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(eeExpect), utils.ToJSON(newEE))
	}
}

func TestNewEventExporterDefaultCase(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaNone
	cgrCfg.EEsCfg().Exporters[0].ConcurrentRequests = 0
	filterS := engine.NewFilterS(cgrCfg, nil, nil)
	_, err := NewEventExporter(cgrCfg.EEsCfg().Exporters[0], cgrCfg, filterS, nil)
	errExpect := fmt.Sprintf("unsupported exporter type: <%s>", utils.MetaNone)
	if err.Error() != errExpect {
		t.Errorf("Expected %+v \n but got %+v", errExpect, err)
	}
}

// Test for Case 7
func TestNewEventExporterCase7(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaElastic
	cgrCfg.EEsCfg().Exporters[0].ConcurrentRequests = 0
	cgrCfg.EEsCfg().Exporters[0].ExportPath = "/invalid/path"
	filterS := engine.NewFilterS(cgrCfg, nil, nil)
	ee, err := NewEventExporter(cgrCfg.EEsCfg().Exporters[0], cgrCfg, filterS, nil)
	if err != nil {
		t.Error(err)
	}
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	if err != nil {
		t.Error(err)
	}
	eeExpect, err := NewElasticEE(cgrCfg.EEsCfg().Exporters[0], dc)
	if err != nil {
		t.Error(err)
	}
	newEE := ee.(*ElasticEE)
	newEE.dc.MapStorage[utils.TimeNow] = nil
	eeExpect.dc.MapStorage[utils.TimeNow] = nil
	eeExpect.eClnt = newEE.eClnt
	if !reflect.DeepEqual(eeExpect, newEE) {
		t.Errorf("Expected %+v \n but got %+v", eeExpect, newEE)
	}
}

// Test for Case 8
func TestNewEventExporterCase8(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaSQL
	cgrCfg.EEsCfg().Exporters[0].ConcurrentRequests = 0
	filterS := engine.NewFilterS(cgrCfg, nil, nil)
	_, err := NewEventExporter(cgrCfg.EEsCfg().Exporters[0], cgrCfg, filterS, nil)
	errExpect := "MANDATORY_IE_MISSING: [sqlTableName]"
	if err == nil || err.Error() != errExpect {
		t.Errorf("Expected %+v \n but got %+v", errExpect, err)
	}
}

// Test for invalid "dc"
func TestNewEventExporterDcCase(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.GeneralCfg().DefaultTimezone = "invalid_timezone"
	_, err := NewEventExporter(cgrCfg.EEsCfg().Exporters[0], cgrCfg, nil, nil)
	errExpect := "unknown time zone invalid_timezone"
	if err == nil || err.Error() != errExpect {
		t.Errorf("Expected %+v \n but got %+v", errExpect, err)
	}
}

func TestNewConcReq(t *testing.T) {
	if reply := newConcReq(5); len(reply.reqs) != 5 {
		t.Errorf("Expected 5 \n but received \n %v", len(reply.reqs))
	}
}

func TestGet(t *testing.T) {
	c := &concReq{
		reqs:  make(chan struct{}, 2),
		limit: 2,
	}
	c.reqs <- struct{}{}
	c.reqs <- struct{}{}
	c.get()
	if len(c.reqs) != 1 {
		t.Error("Expected length of 1")
	}
}

func TestDone(t *testing.T) {
	c := &concReq{
		reqs:  make(chan struct{}, 3),
		limit: 3,
	}
	c.reqs <- struct{}{}
	c.reqs <- struct{}{}
	c.done()
	if len(c.reqs) != 3 {
		t.Error("Expected length of 3")
	}
}

func TestEEPrepareOrderMap(t *testing.T) {
	bP := new(bytePreparing)
	onm := utils.NewOrderedNavigableMap()
	fullPath := &utils.FullPath{
		PathSlice: []string{utils.MetaReq, utils.MetaTenant},
		Path:      utils.MetaTenant,
	}
	val := &utils.DataLeaf{
		Data: "value1",
	}
	onm.Append(fullPath, val)
	rcv, err := bP.PrepareOrderMap(onm)
	if err != nil {
		t.Error(err)
	}

	valMp := map[string]interface{}{
		"*req.*tenant": "value1",
	}
	body, err := json.Marshal(valMp)
	if !reflect.DeepEqual(rcv, body) {
		t.Errorf("Expected %v \n but received \n %v", utils.IfaceAsString(body), utils.IfaceAsString(rcv))
	}
}
