/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2015 ITsysCOM GmbH

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

func TestLoadCdrcConfigMultipleFiles(t *testing.T) {
	cgrCfg, err := NewCGRConfigFromFolder(".")
	if err != nil {
		t.Error(err)
	}
	eCgrCfg, _ := NewDefaultCGRConfig()
	eCgrCfg.CdrcProfiles = make(map[string]map[string]*CdrcConfig)
	// Default instance first
	eCgrCfg.CdrcProfiles["/var/log/cgrates/cdrc/in"] = map[string]*CdrcConfig{
		"*default": &CdrcConfig{
			Enabled:                 false,
			Cdrs:                    "internal",
			CdrFormat:               "csv",
			FieldSeparator:          ',',
			DataUsageMultiplyFactor: 1024,
			RunDelay:                0,
			CdrInDir:                "/var/log/cgrates/cdrc/in",
			CdrOutDir:               "/var/log/cgrates/cdrc/out",
			CdrSourceId:             "freeswitch_csv",
			CdrFilter:               utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP),
			CdrFields: []*CfgCdrField{
				&CfgCdrField{Tag: "tor", Type: "cdrfield", CdrFieldId: "tor", Value: utils.ParseRSRFieldsMustCompile("2", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "accid", Type: "cdrfield", CdrFieldId: "accid", Value: utils.ParseRSRFieldsMustCompile("3", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "reqtype", Type: "cdrfield", CdrFieldId: "reqtype", Value: utils.ParseRSRFieldsMustCompile("4", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "direction", Type: "cdrfield", CdrFieldId: "direction", Value: utils.ParseRSRFieldsMustCompile("5", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "tenant", Type: "cdrfield", CdrFieldId: "tenant", Value: utils.ParseRSRFieldsMustCompile("6", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "category", Type: "cdrfield", CdrFieldId: "category", Value: utils.ParseRSRFieldsMustCompile("7", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "account", Type: "cdrfield", CdrFieldId: "account", Value: utils.ParseRSRFieldsMustCompile("8", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "subject", Type: "cdrfield", CdrFieldId: "subject", Value: utils.ParseRSRFieldsMustCompile("9", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "destination", Type: "cdrfield", CdrFieldId: "destination", Value: utils.ParseRSRFieldsMustCompile("10", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "setup_time", Type: "cdrfield", CdrFieldId: "setup_time", Value: utils.ParseRSRFieldsMustCompile("11", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "answer_time", Type: "cdrfield", CdrFieldId: "answer_time", Value: utils.ParseRSRFieldsMustCompile("12", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "usage", Type: "cdrfield", CdrFieldId: "usage", Value: utils.ParseRSRFieldsMustCompile("13", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
			},
		},
	}
	eCgrCfg.CdrcProfiles["/tmp/cgrates/cdrc1/in"] = map[string]*CdrcConfig{
		"CDRC-CSV1": &CdrcConfig{
			Enabled:                 true,
			Cdrs:                    "internal",
			CdrFormat:               "csv",
			FieldSeparator:          ',',
			DataUsageMultiplyFactor: 1024,
			RunDelay:                0,
			CdrInDir:                "/tmp/cgrates/cdrc1/in",
			CdrOutDir:               "/tmp/cgrates/cdrc1/out",
			CdrSourceId:             "csv1",
			CdrFilter:               utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP),
			CdrFields: []*CfgCdrField{
				&CfgCdrField{Tag: "tor", Type: "cdrfield", CdrFieldId: "tor", Value: utils.ParseRSRFieldsMustCompile("2", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "accid", Type: "cdrfield", CdrFieldId: "accid", Value: utils.ParseRSRFieldsMustCompile("3", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "reqtype", Type: "cdrfield", CdrFieldId: "reqtype", Value: utils.ParseRSRFieldsMustCompile("4", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "direction", Type: "cdrfield", CdrFieldId: "direction", Value: utils.ParseRSRFieldsMustCompile("5", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "tenant", Type: "cdrfield", CdrFieldId: "tenant", Value: utils.ParseRSRFieldsMustCompile("6", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "category", Type: "cdrfield", CdrFieldId: "category", Value: utils.ParseRSRFieldsMustCompile("7", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "account", Type: "cdrfield", CdrFieldId: "account", Value: utils.ParseRSRFieldsMustCompile("8", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "subject", Type: "cdrfield", CdrFieldId: "subject", Value: utils.ParseRSRFieldsMustCompile("9", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "destination", Type: "cdrfield", CdrFieldId: "destination", Value: utils.ParseRSRFieldsMustCompile("10", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "setup_time", Type: "cdrfield", CdrFieldId: "setup_time", Value: utils.ParseRSRFieldsMustCompile("11", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "answer_time", Type: "cdrfield", CdrFieldId: "answer_time", Value: utils.ParseRSRFieldsMustCompile("12", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "usage", Type: "cdrfield", CdrFieldId: "usage", Value: utils.ParseRSRFieldsMustCompile("13", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
			},
		},
	}
	eCgrCfg.CdrcProfiles["/tmp/cgrates/cdrc2/in"] = map[string]*CdrcConfig{
		"CDRC-CSV2": &CdrcConfig{
			Enabled:                 true,
			Cdrs:                    "internal",
			CdrFormat:               "csv",
			FieldSeparator:          ',',
			DataUsageMultiplyFactor: 0.000976563,
			RunDelay:                0,
			CdrInDir:                "/tmp/cgrates/cdrc2/in",
			CdrOutDir:               "/tmp/cgrates/cdrc2/out",
			CdrSourceId:             "csv2",
			CdrFilter:               utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP),
			CdrFields: []*CfgCdrField{
				&CfgCdrField{Tag: "", Type: "", CdrFieldId: "tor", Value: utils.ParseRSRFieldsMustCompile("~7:s/^(voice|data|sms|generic)$/*$1/", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: false},
				&CfgCdrField{Tag: "", Type: "", CdrFieldId: "answer_time", Value: utils.ParseRSRFieldsMustCompile("2", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: false},
			},
		},
	}
	eCgrCfg.CdrcProfiles["/tmp/cgrates/cdrc3/in"] = map[string]*CdrcConfig{
		"CDRC-CSV3": &CdrcConfig{
			Enabled:                 true,
			Cdrs:                    "internal",
			CdrFormat:               "csv",
			FieldSeparator:          ',',
			DataUsageMultiplyFactor: 1024,
			RunDelay:                0,
			CdrInDir:                "/tmp/cgrates/cdrc3/in",
			CdrOutDir:               "/tmp/cgrates/cdrc3/out",
			CdrSourceId:             "csv3",
			CdrFilter:               utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP),
			CdrFields: []*CfgCdrField{
				&CfgCdrField{Tag: "tor", Type: "cdrfield", CdrFieldId: "tor", Value: utils.ParseRSRFieldsMustCompile("2", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "accid", Type: "cdrfield", CdrFieldId: "accid", Value: utils.ParseRSRFieldsMustCompile("3", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "reqtype", Type: "cdrfield", CdrFieldId: "reqtype", Value: utils.ParseRSRFieldsMustCompile("4", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "direction", Type: "cdrfield", CdrFieldId: "direction", Value: utils.ParseRSRFieldsMustCompile("5", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "tenant", Type: "cdrfield", CdrFieldId: "tenant", Value: utils.ParseRSRFieldsMustCompile("6", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "category", Type: "cdrfield", CdrFieldId: "category", Value: utils.ParseRSRFieldsMustCompile("7", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "account", Type: "cdrfield", CdrFieldId: "account", Value: utils.ParseRSRFieldsMustCompile("8", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "subject", Type: "cdrfield", CdrFieldId: "subject", Value: utils.ParseRSRFieldsMustCompile("9", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "destination", Type: "cdrfield", CdrFieldId: "destination", Value: utils.ParseRSRFieldsMustCompile("10", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "setup_time", Type: "cdrfield", CdrFieldId: "setup_time", Value: utils.ParseRSRFieldsMustCompile("11", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "answer_time", Type: "cdrfield", CdrFieldId: "answer_time", Value: utils.ParseRSRFieldsMustCompile("12", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "usage", Type: "cdrfield", CdrFieldId: "usage", Value: utils.ParseRSRFieldsMustCompile("13", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
			},
		},
	}
	if !reflect.DeepEqual(eCgrCfg.CdrcProfiles, cgrCfg.CdrcProfiles) {
		t.Errorf("Expected: %+v, received: %+v", eCgrCfg.CdrcProfiles, cgrCfg.CdrcProfiles)
	}
}
