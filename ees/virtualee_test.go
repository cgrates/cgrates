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
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestVirtualEeID(t *testing.T) {
	vEe := &VirtualEE{
		id: "3",
	}
	if rcv := vEe.ID(); !reflect.DeepEqual(rcv, "3") {
		t.Errorf("Expected %+v \n but got %+v", "3", rcv)
	}
}

func TestVirtualEeGetMetrics(t *testing.T) {
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	if err != nil {
		t.Error(err)
	}
	vEe := &VirtualEE{
		dc: dc,
	}

	if rcv := vEe.GetMetrics(); !reflect.DeepEqual(rcv, vEe.dc) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(rcv), utils.ToJSON(vEe.dc))
	}
}
func TestVirtualEeExportEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
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
	vEe := &VirtualEE{
		id:      "string",
		cgrCfg:  cgrCfg,
		cfgIdx:  0,
		filterS: filterS,
		dc:      dc,
		reqs:    newConcReq(0),
	}
	cgrEv.Event = map[string]interface{}{
		"test1": "value",
	}
	cgrCfg.EEsCfg().Exporters[vEe.cfgIdx].Fields = []*config.FCTemplate{
		{
			Path: "*exp.1", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*req.field1", utils.InfieldSep),
		},
		{
			Path: "*exp.2", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*req.field2", utils.InfieldSep),
		},
	}
	for _, field := range cgrCfg.EEsCfg().Exporters[vEe.cfgIdx].Fields {
		field.ComputePath()
	}
	if err := vEe.ExportEvent(cgrEv); err != nil {
		t.Error(err)
	}
	cgrCfg.EEsCfg().Exporters[vEe.cfgIdx].ComputeFields()
	if err := vEe.ExportEvent(cgrEv); err != nil {
		t.Error(err)
	}
	cgrCfg.EEsCfg().Exporters[vEe.cfgIdx].Fields = []*config.FCTemplate{
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
	for _, field := range cgrCfg.EEsCfg().Exporters[vEe.cfgIdx].Fields {
		field.ComputePath()
	}
	cgrCfg.EEsCfg().Exporters[vEe.cfgIdx].ComputeFields()
	errExpect := "inline parse error for string: <*wrong-type>"
	if err := vEe.ExportEvent(cgrEv); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %q but received %q", errExpect, err)
	}
	vEe.OnEvicted("test", "test")
}
