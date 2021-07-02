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
	for key, val := range map[string]string{"LOGGER": "*syslog", "LOG_LEVEL": "6", "ROUND_DEC": "5",
		"DB_ENCODING": "*msgpack", "TP_EXPORT_DIR": "/var/spool/cgrates/tpe", "FAILED_POSTS_DIR": "/var/spool/cgrates/failed_posts",
		"DF_TENANT": "cgrates.org", "TIMEZONE": "Local"} {
		os.Setenv(key, val)
	}
	var err error
	if mfCgrCfg, err = NewCGRConfigFromPath("/usr/share/cgrates/conf/samples/multifiles"); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

func TestMfGeneralItems(t *testing.T) {
	if mfCgrCfg.GeneralCfg().DefaultReqType != utils.MetaPseudoPrepaid { // Twice reconfigured
		t.Error("DefaultReqType: ", mfCgrCfg.GeneralCfg().DefaultReqType)
	}
	if mfCgrCfg.GeneralCfg().DefaultCategory != "call" { // Not configred, should be inherited from default
		t.Error("DefaultCategory: ", mfCgrCfg.GeneralCfg().DefaultCategory)
	}
}

func TestMfEnvReaderITRead(t *testing.T) {
	expected := GeneralCfg{
		NodeID:           "d80fac5",
		Logger:           "*syslog",
		LogLevel:         6,
		RoundingDecimals: 5,
		DBDataEncoding:   "msgpack",
		TpExportPath:     "/var/spool/cgrates/tpe",
		PosterAttempts:   3,
		FailedPostsDir:   "/var/spool/cgrates/failed_posts",
		DefaultReqType:   utils.MetaPseudoPrepaid,
		DefaultCategory:  "call",
		DefaultTenant:    "cgrates.org",
		DefaultCaching:   utils.MetaReload,
		DefaultTimezone:  "Local",
		ConnectAttempts:  3,
		Reconnects:       -1,
		ConnectTimeout:   time.Second,
		ReplyTimeout:     2 * time.Second,
		LockingTimeout:   0,
		DigestSeparator:  ",",
		DigestEqual:      ":",
		RSRSep:           ";",
		MaxParallelConns: 100,
		FailedPostsTTL:   5 * time.Second,
	}
	if !reflect.DeepEqual(expected, *mfCgrCfg.generalCfg) {
		t.Errorf("Expected: %+v\n, received: %+v", utils.ToJSON(expected), utils.ToJSON(*mfCgrCfg.generalCfg))
	}
}

func TestMfHttpAgentMultipleFields(t *testing.T) {
	if len(mfCgrCfg.HTTPAgentCfg()) != 2 {
		t.Errorf("Expected: 2, received: %+v", len(mfCgrCfg.HTTPAgentCfg()))
	}
	expected := HTTPAgentCfgs{
		{
			ID:             "conecto1",
			URL:            "/newConecto",
			SessionSConns:  []string{utils.MetaLocalHost},
			RequestPayload: "*url",
			ReplyPayload:   "*xml",
			RequestProcessors: []*RequestProcessor{
				{
					ID:            "OutboundAUTHDryRun",
					Filters:       []string{},
					Tenant:        NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep),
					Flags:         utils.FlagsWithParams{"*dryRun": {}},
					RequestFields: []*FCTemplate{},
					ReplyFields: []*FCTemplate{{
						Tag:       "Allow",
						Path:      "response.Allow",
						Type:      "*constant",
						Value:     NewRSRParsersMustCompile("1", utils.InfieldSep),
						Mandatory: true,
						Layout:    time.RFC3339,
					}},
				},
				{
					ID:      "OutboundAUTH",
					Filters: []string{"*string:~*req.request_type:OutboundAUTH"},
					Tenant:  NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep),
					Flags: utils.FlagsWithParams{"*accounts": {},
						"*attributes": {}, "*authorize": {}},
					RequestFields: []*FCTemplate{
						{
							Tag:       "RequestType",
							Path:      "RequestType",
							Type:      "*constant",
							Value:     NewRSRParsersMustCompile("*pseudoprepaid", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339,
						},
					},
					ReplyFields: []*FCTemplate{
						{
							Tag:       "Allow",
							Path:      "response.Allow",
							Type:      "*constant",
							Value:     NewRSRParsersMustCompile("1", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339,
						},
					},
				},
				{
					ID:      "mtcall_cdr",
					Filters: []string{"*string:~*req.request_type:MTCALL_CDR"},
					Tenant:  NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep),
					Flags:   utils.FlagsWithParams{"*cdrs": {}},
					RequestFields: []*FCTemplate{{
						Tag:       "RequestType",
						Path:      "RequestType",
						Type:      "*constant",
						Value:     NewRSRParsersMustCompile("*pseudoprepaid", utils.InfieldSep),
						Mandatory: true,
						Layout:    time.RFC3339,
					}},
					ReplyFields: []*FCTemplate{{
						Tag:       "CDR_ID",
						Path:      "CDR_RESPONSE.CDR_ID",
						Type:      "*variable",
						Value:     NewRSRParsersMustCompile("~*req.CDR_ID", utils.InfieldSep),
						Mandatory: true,
						Layout:    time.RFC3339,
					}},
				},
			},
		},
		{
			ID:             "conecto_xml",
			URL:            "/conecto_xml",
			SessionSConns:  []string{utils.MetaLocalHost},
			RequestPayload: "*xml",
			ReplyPayload:   "*xml",
			RequestProcessors: []*RequestProcessor{{
				ID:     "cdr_from_xml",
				Tenant: NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep),
				Flags:  utils.FlagsWithParams{"*cdrs": {}},
				RequestFields: []*FCTemplate{
					{
						Tag:       "ToR",
						Path:      "ToR",
						Type:      "*constant",
						Value:     NewRSRParsersMustCompile("*data", utils.InfieldSep),
						Mandatory: true,
						Layout:    time.RFC3339,
					},
				},
				ReplyFields: []*FCTemplate{},
			}}},
	}
	for _, profile := range expected {
		for _, rp := range profile.RequestProcessors {
			for _, v := range rp.ReplyFields {
				v.ComputePath()
			}
			for _, v := range rp.RequestFields {
				v.ComputePath()
			}
		}
	}

	if !reflect.DeepEqual(mfCgrCfg.HTTPAgentCfg(), expected) {
		t.Errorf("Expected: %+v\n, received: %+v", utils.ToJSON(expected), utils.ToJSON(mfCgrCfg.HTTPAgentCfg()))
	}
}
