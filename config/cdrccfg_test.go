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
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

var cdrcCfg = CdrcCfg{
	ID:                       "*default",
	CdrsConns:                []*HaPoolConfig{{Address: utils.MetaInternal}},
	CdrFormat:                "csv",
	FieldSeparator:           ',',
	MaxOpenFiles:             1024,
	DataUsageMultiplyFactor:  1024,
	CdrInDir:                 "/var/spool/cgrates/cdrc/in",
	CdrOutDir:                "/var/spool/cgrates/cdrc/out",
	FailedCallsPrefix:        "missed_calls",
	CDRPath:                  []string{""},
	CdrSourceId:              "freeswitch_csv",
	Filters:                  []string{},
	Tenant:                   NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP),
	PartialRecordCache:       time.Duration(10 * time.Second),
	PartialCacheExpiryAction: "*dump_to_file",
	HeaderFields:             []*FCTemplate{},
	ContentFields: []*FCTemplate{
		{
			Tag:       "TOR",
			FieldId:   "ToR",
			Type:      "*composed",
			Value:     NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP),
			Mandatory: true,
		},
	},
	TrailerFields: []*FCTemplate{},
	CacheDumpFields: []*FCTemplate{
		{
			Tag:   "CGRID",
			Type:  "*composed",
			Value: NewRSRParsersMustCompile("~CGRID", true, utils.INFIELD_SEP),
		},
	},
}

func TestCdrcCfgloadFromJsonCfg(t *testing.T) {
	var cdrccfg, expected CdrcCfg
	if err := cdrccfg.loadFromJsonCfg(nil, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cdrccfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, cdrccfg)
	}
	if err := cdrccfg.loadFromJsonCfg(new(CdrcJsonCfg), utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cdrccfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, cdrccfg)
	}
	cfgJSONStr := `{
"cdrc": [
	{
		"id": "*default",								// identifier of the CDRC runner
		"enabled": false,								// enable CDR client functionality
		"dry_run": false,								// do not send the CDRs to CDRS, just parse them
		"cdrs_conns": [
			{"address": "*internal"}					// address where to reach CDR server. <*internal|x.y.z.y:1234>
		],
		"cdr_format": "csv",							// CDR file format <csv|freeswitch_csv|fwv|opensips_flatstore|partial_csv>
		"field_separator": ",",							// separator used in case of csv files
		"timezone": "",									// timezone for timestamps where not specified <""|UTC|Local|$IANA_TZ_DB>
		"run_delay": 0,									// sleep interval in seconds between consecutive runs, 0 to use automation via inotify
		"max_open_files": 1024,							// maximum simultaneous files to process, 0 for unlimited
		"data_usage_multiply_factor": 1024,				// conversion factor for data usage
		"cdr_in_dir": "/var/spool/cgrates/cdrc/in",		// absolute path towards the directory where the CDRs are stored
		"cdr_out_dir": "/var/spool/cgrates/cdrc/out",	// absolute path towards the directory where processed CDRs will be moved
		"failed_calls_prefix": "missed_calls",			// used in case of flatstore CDRs to avoid searching for BYE records
		"cdr_path": "",									// path towards one CDR element in case of XML CDRs
		"cdr_source_id": "freeswitch_csv",				// free form field, tag identifying the source of the CDRs within CDRS database
		"filters" :[],									// new filters used in FilterS subsystem
		"tenant": "cgrates.org",						// default tenant
		"continue_on_success": false,					// continue to the next template if executed
		"partial_record_cache": "10s",					// duration to cache partial records when not pairing
		"partial_cache_expiry_action": "*dump_to_file",	// action taken when cache when records in cache are timed-out <*dump_to_file|*post_cdr>
		"header_fields": [],							// template of the import header fields
		"content_fields":[								// import content_fields template, tag will match internally CDR field, in case of .csv value will be represented by index of the field value
			{"tag": "TOR", "field_id": "ToR", "type": "*composed", "value": "~2", "mandatory": true},
		],
		"trailer_fields": [],							// template of the import trailer fields
		"cache_dump_fields": [							// template used when dumping cached CDR, eg: partial CDRs
			{"tag": "CGRID", "type": "*composed", "value": "~CGRID"},
		],
	},
],		
}`
	expected = cdrcCfg
	if jsnCfg, err := NewCgrJsonCfgFromReader(strings.NewReader(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnCdrcCfg, err := jsnCfg.CdrcJsonCfg(); err != nil {
		t.Error(err)
	} else if err = cdrccfg.loadFromJsonCfg(jsnCdrcCfg[0], utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, cdrccfg) {
		t.Errorf("Expected: %+v , recived: %+v", utils.ToJSON(expected), utils.ToJSON(cdrccfg))
	}
}

func TestCdrcCfgClone(t *testing.T) {
	clnCdrcCfg := *cdrcCfg.Clone()
	if !reflect.DeepEqual(cdrcCfg, clnCdrcCfg) {
		t.Errorf("Expected: %+v , recived: %+v", cdrcCfg, clnCdrcCfg)
	}
	cdrcCfg.ContentFields[0].Tag = "CGRID"
	if reflect.DeepEqual(cdrcCfg, clnCdrcCfg) { // MOdifying a field after clone should not affect cloned instance
		t.Errorf("Cloned result: %+v", utils.ToJSON(clnCdrcCfg))
	}
	clnCdrcCfg.ContentFields[0].FieldId = "destination"
	if cdrcCfg.ContentFields[0].FieldId != "ToR" {
		t.Error("Unexpected change of FieldId: ", cdrcCfg.ContentFields[0].FieldId)
	}
}
