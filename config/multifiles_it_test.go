// +build integration

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
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

var mfCgrCfg *CGRConfig

func TestMfInitConfig(t *testing.T) {
	for key, val := range map[string]string{"LOGGER": "*syslog", "LOG_LEVEL": "6", "TLS_VERIFY": "false", "ROUND_DEC": "5",
		"DB_ENCODING": "*msgpack", "TP_EXPORT_DIR": "/var/spool/cgrates/tpe", "FAILED_POSTS_DIR": "/var/spool/cgrates/failed_posts",
		"DF_TENANT": "cgrates.org", "TIMEZONE": "Local"} {
		os.Setenv(key, val)
	}
	var err error
	if mfCgrCfg, err = NewCGRConfigFromFolder("/usr/share/cgrates/conf/samples/multifiles"); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

func TestMfGeneralItems(t *testing.T) {
	if mfCgrCfg.GeneralCfg().DefaultReqType != utils.META_PSEUDOPREPAID { // Twice reconfigured
		t.Error("DefaultReqType: ", mfCgrCfg.GeneralCfg().DefaultReqType)
	}
	if mfCgrCfg.GeneralCfg().DefaultCategory != "call" { // Not configred, should be inherited from default
		t.Error("DefaultCategory: ", mfCgrCfg.GeneralCfg().DefaultCategory)
	}
}

func TestMfCdreDefaultInstance(t *testing.T) {
	for _, prflName := range []string{"*default", "export1"} {
		if _, hasIt := mfCgrCfg.CdreProfiles[prflName]; !hasIt {
			t.Error("Cdre does not contain profile ", prflName)
		}
	}
	prfl := "*default"
	if mfCgrCfg.CdreProfiles[prfl].ExportFormat != utils.MetaFileCSV {
		t.Error("Default instance has cdrFormat: ", mfCgrCfg.CdreProfiles[prfl].ExportFormat)
	}
	if mfCgrCfg.CdreProfiles[prfl].CostMultiplyFactor != 1024.0 {
		t.Error("Default instance has cdrFormat: ", mfCgrCfg.CdreProfiles[prfl].CostMultiplyFactor)
	}
	if len(mfCgrCfg.CdreProfiles[prfl].HeaderFields) != 0 {
		t.Error("Default instance has number of header fields: ", len(mfCgrCfg.CdreProfiles[prfl].HeaderFields))
	}
	if len(mfCgrCfg.CdreProfiles[prfl].ContentFields) != 12 {
		t.Error("Default instance has number of content fields: ", len(mfCgrCfg.CdreProfiles[prfl].ContentFields))
	}
	if mfCgrCfg.CdreProfiles[prfl].ContentFields[2].Tag != "Direction" {
		t.Error("Unexpected headerField value: ", mfCgrCfg.CdreProfiles[prfl].ContentFields[2].Tag)
	}
}

func TestMfCdreExport1Instance(t *testing.T) {
	prfl := "export1"
	if mfCgrCfg.CdreProfiles[prfl].ExportFormat != utils.MetaFileCSV {
		t.Error("Export1 instance has cdrFormat: ", mfCgrCfg.CdreProfiles[prfl].ExportFormat)
	}
	if mfCgrCfg.CdreProfiles[prfl].CostMultiplyFactor != 1.0 {
		t.Error("Export1 instance has DataUsageMultiplyFormat: ", mfCgrCfg.CdreProfiles[prfl].CostMultiplyFactor)
	}
	if len(mfCgrCfg.CdreProfiles[prfl].HeaderFields) != 2 {
		t.Error("Export1 instance has number of header fields: ", len(mfCgrCfg.CdreProfiles[prfl].HeaderFields))
	}
	if mfCgrCfg.CdreProfiles[prfl].HeaderFields[1].Tag != "RunId" {
		t.Error("Unexpected headerField value: ", mfCgrCfg.CdreProfiles[prfl].HeaderFields[1].Tag)
	}
	if len(mfCgrCfg.CdreProfiles[prfl].ContentFields) != 9 {
		t.Error("Export1 instance has number of content fields: ", len(mfCgrCfg.CdreProfiles[prfl].ContentFields))
	}
	if mfCgrCfg.CdreProfiles[prfl].ContentFields[2].Tag != "Account" {
		t.Error("Unexpected headerField value: ", mfCgrCfg.CdreProfiles[prfl].ContentFields[2].Tag)
	}
}

func TestMfEnvReaderITRead(t *testing.T) {
	expected := GeneralCfg{
		NodeID:            "d80fac5",
		Logger:            "*syslog",
		LogLevel:          6,
		HttpSkipTlsVerify: false,
		RoundingDecimals:  5,
		DBDataEncoding:    "msgpack",
		TpExportPath:      "/var/spool/cgrates/tpe",
		PosterAttempts:    3,
		FailedPostsDir:    "/var/spool/cgrates/failed_posts",
		DefaultReqType:    utils.META_PSEUDOPREPAID,
		DefaultCategory:   "call",
		DefaultTenant:     "cgrates.org",
		DefaultTimezone:   "Local",
		ConnectAttempts:   3,
		Reconnects:        -1,
		ConnectTimeout:    time.Duration(1 * time.Second),
		ReplyTimeout:      time.Duration(2 * time.Second),
		ResponseCacheTTL:  time.Duration(0),
		InternalTtl:       time.Duration(2 * time.Minute),
		LockingTimeout:    time.Duration(0),
		DigestSeparator:   ",",
		DigestEqual:       ":",
		RsrSepatarot:      ";",
	}
	if !reflect.DeepEqual(expected, *mfCgrCfg.generalCfg) {
		t.Errorf("Expected: %+v\n, recived: %+v", utils.ToJSON(expected), utils.ToJSON(*mfCgrCfg.generalCfg))
	}
}

func TestMfHttpAgentMultipleFields(t *testing.T) {
	if len(mfCgrCfg.HttpAgentCfg()) != 2 {
		t.Errorf("Expected: 2, recived: %+v", len(mfCgrCfg.HttpAgentCfg()))
	}
	expected := []*HttpAgentCfg{
		&HttpAgentCfg{
			ID:             "conecto1",
			Url:            "/newConecto",
			SessionSConns:  []*HaPoolConfig{{Address: "127.0.0.2:2012", Transport: "*json"}},
			RequestPayload: "*url",
			ReplyPayload:   "*xml",
			RequestProcessors: []*HttpAgntProcCfg{
				{
					Id:            "OutboundAUTHDryRun",
					Filters:       []string{},
					Tenant:        NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP),
					Flags:         utils.StringMap{"*dryrun": true},
					RequestFields: []*FCTemplate{},
					ReplyFields: []*FCTemplate{{
						Tag:       "Allow",
						FieldId:   "response.Allow",
						Type:      "*constant",
						Value:     NewRSRParsersMustCompile("1", true, utils.INFIELD_SEP),
						Mandatory: true,
					}},
				},
				{
					Id:      "OutboundAUTH",
					Filters: []string{"*string:*req.request_type:OutboundAUTH"},
					Tenant:  NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP),
					Flags: utils.StringMap{"*accounts": true,
						"*attributes": true, "*auth": true},
					RequestFields: []*FCTemplate{
						{
							Tag:       "RequestType",
							FieldId:   "RequestType",
							Type:      "*constant",
							Value:     NewRSRParsersMustCompile("*pseudoprepaid", true, utils.INFIELD_SEP),
							Mandatory: true,
						},
					},
					ReplyFields: []*FCTemplate{
						{
							Tag:       "Allow",
							FieldId:   "response.Allow",
							Type:      "*constant",
							Value:     NewRSRParsersMustCompile("1", true, utils.INFIELD_SEP),
							Mandatory: true,
						},
					},
				},
				{
					Id:      "mtcall_cdr",
					Filters: []string{"*string:*req.request_type:MTCALL_CDR"},
					Tenant:  NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP),
					Flags:   utils.StringMap{"*cdrs": true},
					RequestFields: []*FCTemplate{{
						Tag:       "RequestType",
						FieldId:   "RequestType",
						Type:      "*constant",
						Value:     NewRSRParsersMustCompile("*pseudoprepaid", true, utils.INFIELD_SEP),
						Mandatory: true,
					}},
					ReplyFields: []*FCTemplate{{
						Tag:       "CDR_ID",
						FieldId:   "CDR_RESPONSE.CDR_ID",
						Type:      "*composed",
						Value:     NewRSRParsersMustCompile("~*req.CDR_ID", true, utils.INFIELD_SEP),
						Mandatory: true,
					}},
				},
			},
		},
		&HttpAgentCfg{
			ID:             "conecto_xml",
			Url:            "/conecto_xml",
			SessionSConns:  []*HaPoolConfig{{Address: "127.0.0.1:2012", Transport: "*json"}},
			RequestPayload: "*xml",
			ReplyPayload:   "*xml",
			RequestProcessors: []*HttpAgntProcCfg{{
				Id:     "cdr_from_xml",
				Tenant: NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP),
				Flags:  utils.StringMap{"*cdrs": true},
				RequestFields: []*FCTemplate{
					{
						Tag:       "TOR",
						FieldId:   "ToR",
						Type:      "*constant",
						Value:     NewRSRParsersMustCompile("*data", true, utils.INFIELD_SEP),
						Mandatory: true,
					},
				},
				ReplyFields: []*FCTemplate{},
			}}},
	}

	if !reflect.DeepEqual(mfCgrCfg.HttpAgentCfg(), expected) {
		t.Errorf("Expected: %+v\n, recived: %+v", utils.ToJSON(expected), utils.ToJSON(mfCgrCfg.HttpAgentCfg()))
	}
}
