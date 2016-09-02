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
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestLoadCdrcConfigMultipleFiles(t *testing.T) {
	cgrCfg, err := NewCGRConfigFromFolder(".")
	if err != nil {
		t.Error(err)
	}
	eCgrCfg, _ := NewDefaultCGRConfig()
	eCgrCfg.CdrcProfiles = make(map[string][]*CdrcConfig)
	// Default instance first
	eCgrCfg.CdrcProfiles["/var/spool/cgrates/cdrc/in"] = []*CdrcConfig{
		&CdrcConfig{
			ID:                       utils.META_DEFAULT,
			Enabled:                  false,
			CdrsConns:                []*HaPoolConfig{&HaPoolConfig{Address: utils.MetaInternal}},
			CdrFormat:                "csv",
			FieldSeparator:           ',',
			DataUsageMultiplyFactor:  1024,
			RunDelay:                 0,
			MaxOpenFiles:             1024,
			CdrInDir:                 "/var/spool/cgrates/cdrc/in",
			CdrOutDir:                "/var/spool/cgrates/cdrc/out",
			FailedCallsPrefix:        "missed_calls",
			CDRPath:                  utils.HierarchyPath([]string{""}),
			CdrSourceId:              "freeswitch_csv",
			CdrFilter:                utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP),
			PartialRecordCache:       time.Duration(10) * time.Second,
			PartialCacheExpiryAction: utils.MetaDumpToFile,
			HeaderFields:             make([]*CfgCdrField, 0),
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
			CacheDumpFields: []*CfgCdrField{
				&CfgCdrField{Tag: "CGRID", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.CGRID, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "RunID", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.MEDI_RUNID, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "TOR", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.TOR, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "OriginID", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.ACCID, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "RequestType", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.REQTYPE, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "Direction", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.DIRECTION, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "Tenant", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.TENANT, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "Category", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.CATEGORY, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "Account", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.ACCOUNT, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "Subject", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.SUBJECT, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "Destination", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.DESTINATION, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "SetupTime", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.SETUP_TIME, utils.INFIELD_SEP), Layout: "2006-01-02T15:04:05Z07:00"},
				&CfgCdrField{Tag: "AnswerTime", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.ANSWER_TIME, utils.INFIELD_SEP), Layout: "2006-01-02T15:04:05Z07:00"},
				&CfgCdrField{Tag: "Usage", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.USAGE, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "Cost", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.COST, utils.INFIELD_SEP)},
			},
		},
	}
	eCgrCfg.CdrcProfiles["/tmp/cgrates/cdrc1/in"] = []*CdrcConfig{
		&CdrcConfig{
			ID:                       "CDRC-CSV1",
			Enabled:                  true,
			CdrsConns:                []*HaPoolConfig{&HaPoolConfig{Address: utils.MetaInternal}},
			CdrFormat:                "csv",
			FieldSeparator:           ',',
			DataUsageMultiplyFactor:  1024,
			RunDelay:                 0,
			MaxOpenFiles:             1024,
			CdrInDir:                 "/tmp/cgrates/cdrc1/in",
			CdrOutDir:                "/tmp/cgrates/cdrc1/out",
			CDRPath:                  utils.HierarchyPath([]string{""}),
			CdrSourceId:              "csv1",
			CdrFilter:                utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP),
			PartialRecordCache:       time.Duration(10) * time.Second,
			PartialCacheExpiryAction: utils.MetaDumpToFile,
			HeaderFields:             make([]*CfgCdrField, 0),
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
			CacheDumpFields: []*CfgCdrField{
				&CfgCdrField{Tag: "CGRID", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.CGRID, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "RunID", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.MEDI_RUNID, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "TOR", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.TOR, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "OriginID", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.ACCID, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "RequestType", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.REQTYPE, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "Direction", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.DIRECTION, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "Tenant", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.TENANT, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "Category", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.CATEGORY, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "Account", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.ACCOUNT, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "Subject", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.SUBJECT, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "Destination", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.DESTINATION, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "SetupTime", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.SETUP_TIME, utils.INFIELD_SEP), Layout: "2006-01-02T15:04:05Z07:00"},
				&CfgCdrField{Tag: "AnswerTime", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.ANSWER_TIME, utils.INFIELD_SEP), Layout: "2006-01-02T15:04:05Z07:00"},
				&CfgCdrField{Tag: "Usage", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.USAGE, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "Cost", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.COST, utils.INFIELD_SEP)},
			},
		},
	}
	eCgrCfg.CdrcProfiles["/tmp/cgrates/cdrc2/in"] = []*CdrcConfig{
		&CdrcConfig{
			ID:                       "CDRC-CSV2",
			Enabled:                  true,
			CdrsConns:                []*HaPoolConfig{&HaPoolConfig{Address: utils.MetaInternal}},
			CdrFormat:                "csv",
			FieldSeparator:           ',',
			DataUsageMultiplyFactor:  0.000976563,
			RunDelay:                 1000000000,
			MaxOpenFiles:             1024,
			CdrInDir:                 "/tmp/cgrates/cdrc2/in",
			CdrOutDir:                "/tmp/cgrates/cdrc2/out",
			CDRPath:                  utils.HierarchyPath([]string{""}),
			CdrSourceId:              "csv2",
			CdrFilter:                utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP),
			PartialRecordCache:       time.Duration(10) * time.Second,
			PartialCacheExpiryAction: utils.MetaDumpToFile,
			HeaderFields:             make([]*CfgCdrField, 0),
			ContentFields: []*CfgCdrField{
				&CfgCdrField{FieldId: utils.TOR, Value: utils.ParseRSRFieldsMustCompile("~7:s/^(voice|data|sms|mms|generic)$/*$1/", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: false},
				&CfgCdrField{Tag: "", Type: "", FieldId: utils.ANSWER_TIME, Value: utils.ParseRSRFieldsMustCompile("1", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: false},
				&CfgCdrField{FieldId: utils.USAGE, Value: utils.ParseRSRFieldsMustCompile("~9:s/^(\\d+)$/${1}s/", utils.INFIELD_SEP),
					FieldFilter: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP), Width: 0, Strip: "", Padding: "", Layout: "", Mandatory: false},
			},
			TrailerFields: make([]*CfgCdrField, 0),
			CacheDumpFields: []*CfgCdrField{
				&CfgCdrField{Tag: "CGRID", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.CGRID, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "RunID", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.MEDI_RUNID, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "TOR", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.TOR, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "OriginID", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.ACCID, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "RequestType", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.REQTYPE, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "Direction", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.DIRECTION, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "Tenant", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.TENANT, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "Category", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.CATEGORY, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "Account", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.ACCOUNT, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "Subject", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.SUBJECT, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "Destination", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.DESTINATION, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "SetupTime", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.SETUP_TIME, utils.INFIELD_SEP), Layout: "2006-01-02T15:04:05Z07:00"},
				&CfgCdrField{Tag: "AnswerTime", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.ANSWER_TIME, utils.INFIELD_SEP), Layout: "2006-01-02T15:04:05Z07:00"},
				&CfgCdrField{Tag: "Usage", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.USAGE, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "Cost", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.COST, utils.INFIELD_SEP)},
			},
		},
	}
	eCgrCfg.CdrcProfiles["/tmp/cgrates/cdrc3/in"] = []*CdrcConfig{
		&CdrcConfig{
			ID:                       "CDRC-CSV3",
			Enabled:                  true,
			CdrsConns:                []*HaPoolConfig{&HaPoolConfig{Address: utils.MetaInternal}},
			CdrFormat:                "csv",
			FieldSeparator:           ',',
			DataUsageMultiplyFactor:  1024,
			RunDelay:                 0,
			MaxOpenFiles:             1024,
			CdrInDir:                 "/tmp/cgrates/cdrc3/in",
			CdrOutDir:                "/tmp/cgrates/cdrc3/out",
			CDRPath:                  utils.HierarchyPath([]string{""}),
			CdrSourceId:              "csv3",
			CdrFilter:                utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP),
			PartialRecordCache:       time.Duration(10) * time.Second,
			PartialCacheExpiryAction: utils.MetaDumpToFile,
			HeaderFields:             make([]*CfgCdrField, 0),
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
			CacheDumpFields: []*CfgCdrField{
				&CfgCdrField{Tag: "CGRID", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.CGRID, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "RunID", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.MEDI_RUNID, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "TOR", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.TOR, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "OriginID", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.ACCID, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "RequestType", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.REQTYPE, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "Direction", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.DIRECTION, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "Tenant", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.TENANT, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "Category", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.CATEGORY, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "Account", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.ACCOUNT, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "Subject", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.SUBJECT, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "Destination", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.DESTINATION, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "SetupTime", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.SETUP_TIME, utils.INFIELD_SEP), Layout: "2006-01-02T15:04:05Z07:00"},
				&CfgCdrField{Tag: "AnswerTime", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.ANSWER_TIME, utils.INFIELD_SEP), Layout: "2006-01-02T15:04:05Z07:00"},
				&CfgCdrField{Tag: "Usage", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.USAGE, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "Cost", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.COST, utils.INFIELD_SEP)},
			},
		},
	}
	if !reflect.DeepEqual(eCgrCfg.CdrcProfiles, cgrCfg.CdrcProfiles) {
		t.Errorf("Expected: \n%s\n, received: \n%s\n", utils.ToJSON(eCgrCfg.CdrcProfiles), utils.ToJSON(cgrCfg.CdrcProfiles))
	}
}
