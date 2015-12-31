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
	"time"

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
			MaxOpenFiles:            1024,
			CdrInDir:                "/var/log/cgrates/cdrc/in",
			CdrOutDir:               "/var/log/cgrates/cdrc/out",
			FailedCallsPrefix:       "missed_calls",
			CdrSourceId:             "freeswitch_csv",
			CdrFilter:               utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP),
			PartialRecordCache:      time.Duration(10) * time.Second,
			HeaderFields:            make([]*CfgCdrField, 0),
			ContentFields: []*CfgCdrField{
				&CfgCdrField{Tag: "TOR", Type: utils.META_COMPOSED, FieldId: utils.TOR, Value: utils.ParseRSRFieldsMustCompile("2", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "OriginID", Type: utils.META_COMPOSED, FieldId: utils.ACCID, Value: utils.ParseRSRFieldsMustCompile("3", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "RequestType", Type: utils.META_COMPOSED, FieldId: utils.REQTYPE, Value: utils.ParseRSRFieldsMustCompile("4", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "Direction", Type: utils.META_COMPOSED, FieldId: utils.DIRECTION, Value: utils.ParseRSRFieldsMustCompile("5", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "Tenant", Type: utils.META_COMPOSED, FieldId: utils.TENANT, Value: utils.ParseRSRFieldsMustCompile("6", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "Category", Type: utils.META_COMPOSED, FieldId: utils.CATEGORY, Value: utils.ParseRSRFieldsMustCompile("7", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "Account", Type: utils.META_COMPOSED, FieldId: utils.ACCOUNT, Value: utils.ParseRSRFieldsMustCompile("8", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "Subject", Type: utils.META_COMPOSED, FieldId: utils.SUBJECT, Value: utils.ParseRSRFieldsMustCompile("9", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "Destination", Type: utils.META_COMPOSED, FieldId: utils.DESTINATION, Value: utils.ParseRSRFieldsMustCompile("10", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "SetupTime", Type: utils.META_COMPOSED, FieldId: utils.SETUP_TIME, Value: utils.ParseRSRFieldsMustCompile("11", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "AnswerTime", Type: utils.META_COMPOSED, FieldId: utils.ANSWER_TIME, Value: utils.ParseRSRFieldsMustCompile("12", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "Usage", Type: utils.META_COMPOSED, FieldId: utils.USAGE, Value: utils.ParseRSRFieldsMustCompile("13", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
			},
			TrailerFields: make([]*CfgCdrField, 0),
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
			MaxOpenFiles:            1024,
			CdrInDir:                "/tmp/cgrates/cdrc1/in",
			CdrOutDir:               "/tmp/cgrates/cdrc1/out",
			CdrSourceId:             "csv1",
			CdrFilter:               utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP),
			HeaderFields:            make([]*CfgCdrField, 0),
			ContentFields: []*CfgCdrField{
				&CfgCdrField{Tag: "TOR", Type: utils.META_COMPOSED, FieldId: utils.TOR, Value: utils.ParseRSRFieldsMustCompile("2", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "OriginID", Type: utils.META_COMPOSED, FieldId: utils.ACCID, Value: utils.ParseRSRFieldsMustCompile("3", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "RequestType", Type: utils.META_COMPOSED, FieldId: utils.REQTYPE, Value: utils.ParseRSRFieldsMustCompile("4", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "Direction", Type: utils.META_COMPOSED, FieldId: utils.DIRECTION, Value: utils.ParseRSRFieldsMustCompile("5", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "Tenant", Type: utils.META_COMPOSED, FieldId: utils.TENANT, Value: utils.ParseRSRFieldsMustCompile("6", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "Category", Type: utils.META_COMPOSED, FieldId: utils.CATEGORY, Value: utils.ParseRSRFieldsMustCompile("7", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "Account", Type: utils.META_COMPOSED, FieldId: utils.ACCOUNT, Value: utils.ParseRSRFieldsMustCompile("8", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "Subject", Type: utils.META_COMPOSED, FieldId: utils.SUBJECT, Value: utils.ParseRSRFieldsMustCompile("9", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "Destination", Type: utils.META_COMPOSED, FieldId: utils.DESTINATION, Value: utils.ParseRSRFieldsMustCompile("10", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "SetupTime", Type: utils.META_COMPOSED, FieldId: utils.SETUP_TIME, Value: utils.ParseRSRFieldsMustCompile("11", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "AnswerTime", Type: utils.META_COMPOSED, FieldId: utils.ANSWER_TIME, Value: utils.ParseRSRFieldsMustCompile("12", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "Usage", Type: utils.META_COMPOSED, FieldId: utils.USAGE, Value: utils.ParseRSRFieldsMustCompile("13", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
			},
			TrailerFields: make([]*CfgCdrField, 0),
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
			MaxOpenFiles:            1024,
			CdrInDir:                "/tmp/cgrates/cdrc2/in",
			CdrOutDir:               "/tmp/cgrates/cdrc2/out",
			CdrSourceId:             "csv2",
			CdrFilter:               utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP),
			HeaderFields:            make([]*CfgCdrField, 0),
			ContentFields: []*CfgCdrField{
				&CfgCdrField{Tag: "", Type: "", FieldId: utils.TOR, Value: utils.ParseRSRFieldsMustCompile("~7:s/^(voice|data|sms|generic)$/*$1/", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: false},
				&CfgCdrField{Tag: "", Type: "", FieldId: utils.ANSWER_TIME, Value: utils.ParseRSRFieldsMustCompile("2", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: false},
			},
			TrailerFields: make([]*CfgCdrField, 0),
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
			MaxOpenFiles:            1024,
			CdrInDir:                "/tmp/cgrates/cdrc3/in",
			CdrOutDir:               "/tmp/cgrates/cdrc3/out",
			CdrSourceId:             "csv3",
			CdrFilter:               utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP),
			HeaderFields:            make([]*CfgCdrField, 0),
			ContentFields: []*CfgCdrField{
				&CfgCdrField{Tag: "TOR", Type: utils.META_COMPOSED, FieldId: utils.TOR, Value: utils.ParseRSRFieldsMustCompile("2", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "OriginID", Type: utils.META_COMPOSED, FieldId: utils.ACCID, Value: utils.ParseRSRFieldsMustCompile("3", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "RequestType", Type: utils.META_COMPOSED, FieldId: utils.REQTYPE, Value: utils.ParseRSRFieldsMustCompile("4", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "Direction", Type: utils.META_COMPOSED, FieldId: utils.DIRECTION, Value: utils.ParseRSRFieldsMustCompile("5", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "Tenant", Type: utils.META_COMPOSED, FieldId: utils.TENANT, Value: utils.ParseRSRFieldsMustCompile("6", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "Category", Type: utils.META_COMPOSED, FieldId: utils.CATEGORY, Value: utils.ParseRSRFieldsMustCompile("7", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "Account", Type: utils.META_COMPOSED, FieldId: utils.ACCOUNT, Value: utils.ParseRSRFieldsMustCompile("8", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "Subject", Type: utils.META_COMPOSED, FieldId: utils.SUBJECT, Value: utils.ParseRSRFieldsMustCompile("9", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "Destination", Type: utils.META_COMPOSED, FieldId: utils.DESTINATION, Value: utils.ParseRSRFieldsMustCompile("10", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "SetupTime", Type: utils.META_COMPOSED, FieldId: utils.SETUP_TIME, Value: utils.ParseRSRFieldsMustCompile("11", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "AnswerTime", Type: utils.META_COMPOSED, FieldId: utils.ANSWER_TIME, Value: utils.ParseRSRFieldsMustCompile("12", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
				&CfgCdrField{Tag: "Usage", Type: utils.META_COMPOSED, FieldId: utils.USAGE, Value: utils.ParseRSRFieldsMustCompile("13", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: true},
			},
			TrailerFields: make([]*CfgCdrField, 0),
		},
	}
	if !reflect.DeepEqual(eCgrCfg.CdrcProfiles, cgrCfg.CdrcProfiles) {
		t.Errorf("Expected: %+v, received: %+v", eCgrCfg.CdrcProfiles, cgrCfg.CdrcProfiles)
	}
}
