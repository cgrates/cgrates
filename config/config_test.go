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

var cfg *CGRConfig

func TestConfigSharing(t *testing.T) {
	cfg, _ = NewDefaultCGRConfig()
	SetCgrConfig(cfg)
	cfgReturn := CgrConfig()
	if !reflect.DeepEqual(cfgReturn, cfg) {
		t.Errorf("Retrieved %v, Expected %v", cfgReturn, cfg)
	}
}

func TestLoadCgrCfgWithDefaults(t *testing.T) {
	JSN_CFG := `
{
"sm_freeswitch": {
	"enabled": true,				// starts SessionManager service: <true|false>
	"event_socket_conns":[					// instantiate connections to multiple FreeSWITCH servers
		{"address": "1.2.3.4:8021", "password": "ClueCon", "reconnects": 3},
		{"address": "1.2.3.5:8021", "password": "ClueCon", "reconnects": 5}
	],
},

}`
	eCgrCfg, _ := NewDefaultCGRConfig()
	eCgrCfg.SmFsConfig.Enabled = true
	eCgrCfg.SmFsConfig.EventSocketConns = []*FsConnConfig{
		&FsConnConfig{Address: "1.2.3.4:8021", Password: "ClueCon", Reconnects: 3},
		&FsConnConfig{Address: "1.2.3.5:8021", Password: "ClueCon", Reconnects: 5},
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(JSN_CFG); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCgrCfg.SmFsConfig, cgrCfg.SmFsConfig) {
		t.Errorf("Expected: %+v, received: %+v", eCgrCfg.SmFsConfig, cgrCfg.SmFsConfig)
	}
}

func TestCgrCfgCDRC(t *testing.T) {
	JSN_RAW_CFG := `
{
"cdrc": [
	{
		"id": "*default",
		"enabled": true,							// enable CDR client functionality
		"content_fields":[							// import template, tag will match internally CDR field, in case of .csv value will be represented by index of the field value
			{"field_id": "ToR", "type": "*composed", "value": "~7:s/^(voice|data|sms|mms|generic)$/*$1/"},
			{"field_id": "AnswerTime", "type": "*composed", "value": "1"},
			{"field_id": "Usage", "type": "*composed", "value": "~9:s/^(\\d+)$/${1}s/"},
		],
	},
],
}`
	eCgrCfg, _ := NewDefaultCGRConfig()
	eCgrCfg.CdrcProfiles["/var/spool/cgrates/cdrc/in"] = []*CdrcConfig{
		&CdrcConfig{
			ID:                       utils.META_DEFAULT,
			Enabled:                  true,
			DryRun:                   false,
			CdrsConns:                []*HaPoolConfig{&HaPoolConfig{Address: utils.MetaInternal}},
			CdrFormat:                "csv",
			FieldSeparator:           rune(','),
			DataUsageMultiplyFactor:  1024,
			Timezone:                 "",
			RunDelay:                 0,
			MaxOpenFiles:             1024,
			CdrInDir:                 "/var/spool/cgrates/cdrc/in",
			CdrOutDir:                "/var/spool/cgrates/cdrc/out",
			FailedCallsPrefix:        "missed_calls",
			CDRPath:                  utils.HierarchyPath([]string{""}),
			CdrSourceId:              "freeswitch_csv",
			ContinueOnSuccess:        false,
			PartialRecordCache:       time.Duration(10 * time.Second),
			PartialCacheExpiryAction: "*dump_to_file",
			HeaderFields:             make([]*CfgCdrField, 0),
			ContentFields: []*CfgCdrField{
				&CfgCdrField{FieldId: "ToR", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile("~7:s/^(voice|data|sms|mms|generic)$/*$1/", utils.INFIELD_SEP)},
				&CfgCdrField{FieldId: "AnswerTime", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile("1", utils.INFIELD_SEP)},
				&CfgCdrField{FieldId: "Usage", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile("~9:s/^(\\d+)$/${1}s/", utils.INFIELD_SEP)},
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
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(JSN_RAW_CFG); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCgrCfg.CdrcProfiles, cgrCfg.CdrcProfiles) {
		t.Errorf("Expected: %+v, received: %+v", eCgrCfg.CdrcProfiles["/var/spool/cgrates/cdrc/in"][0], cgrCfg.CdrcProfiles["/var/spool/cgrates/cdrc/in"][0])
	}
}
