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
package config

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestCdreCfgClone(t *testing.T) {
	cgrIdRsrs := utils.ParseRSRFieldsMustCompile("cgrid", utils.INFIELD_SEP)
	runIdRsrs := utils.ParseRSRFieldsMustCompile("runid", utils.INFIELD_SEP)
	emptyFields := []*CfgCdrField{}
	initContentFlds := []*CfgCdrField{
		&CfgCdrField{Tag: "CgrId",
			Type:    "*composed",
			FieldId: "cgrid",
			Value:   cgrIdRsrs},
		&CfgCdrField{Tag: "RunId",
			Type:    "*composed",
			FieldId: "runid",
			Value:   runIdRsrs},
	}
	initCdreCfg := &CdreConfig{
		ExportFormat:   utils.MetaFileCSV,
		ExportPath:     "/var/spool/cgrates/cdre",
		Synchronous:    true,
		Attempts:       2,
		FieldSeparator: rune(','),
		UsageMultiplyFactor: map[string]float64{
			utils.ANY:  1.0,
			utils.DATA: 1024,
		},
		CostMultiplyFactor: 1.0,
		ContentFields:      initContentFlds,
	}
	eClnContentFlds := []*CfgCdrField{
		&CfgCdrField{Tag: "CgrId",
			Type:    "*composed",
			FieldId: "cgrid",
			Value:   cgrIdRsrs},
		&CfgCdrField{Tag: "RunId",
			Type:    "*composed",
			FieldId: "runid",
			Value:   runIdRsrs},
	}
	eClnCdreCfg := &CdreConfig{
		ExportFormat:   utils.MetaFileCSV,
		ExportPath:     "/var/spool/cgrates/cdre",
		Synchronous:    true,
		Attempts:       2,
		FieldSeparator: rune(','),
		UsageMultiplyFactor: map[string]float64{
			utils.ANY:  1.0,
			utils.DATA: 1024.0,
		},
		CostMultiplyFactor: 1.0,
		HeaderFields:       emptyFields,
		ContentFields:      eClnContentFlds,
		TrailerFields:      emptyFields,
	}
	clnCdreCfg := initCdreCfg.Clone()
	if !reflect.DeepEqual(eClnCdreCfg, clnCdreCfg) {
		t.Errorf("Cloned result: %+v", clnCdreCfg)
	}
	initCdreCfg.UsageMultiplyFactor[utils.DATA] = 2048.0
	if !reflect.DeepEqual(eClnCdreCfg, clnCdreCfg) { // MOdifying a field after clone should not affect cloned instance
		t.Errorf("Cloned result: %+v", clnCdreCfg)
	}
	initContentFlds[0].Tag = "Destination"
	if !reflect.DeepEqual(eClnCdreCfg, clnCdreCfg) { // MOdifying a field after clone should not affect cloned instance
		t.Errorf("Cloned result: %+v", clnCdreCfg)
	}
	clnCdreCfg.ContentFields[0].FieldId = "destination"
	if initCdreCfg.ContentFields[0].FieldId != "cgrid" {
		t.Error("Unexpected change of FieldId: ", initCdreCfg.ContentFields[0].FieldId)
	}

}
