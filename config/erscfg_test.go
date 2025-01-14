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

func TestERSClone(t *testing.T) {
	cfgJSONStr := `{
"ers": {									
	"enabled": true,						
	"sessions_conns":["*internal"],			
	"readers": [
         {
            "id": "file_reader1",
			"run_delay": "-1",
			"type": "*file_csv",
			"flags": ["*dryrun"],
			"source_path": "/tmp/ers/in",
			"processed_path": "/tmp/ers/out",
			"opts": {},
			"tenant": "~*req.Destination1",											
			"timezone": "",										
			"filters": ["randomFiletrs"],										
			"flags": [],										
			"fields":[											
				{"tag": "ToR", "path": "*cgreq.ToR", "type": "*variable", "value": "~*req.2", "mandatory": true},
				{"tag": "OriginID", "path": "*cgreq.OriginID", "type": "*variable", "value": "~*req.3", "mandatory": true},
			],
			"cache_dump_fields": [
               {"tag": "ToR", "path": "*cgreq.ToR", "type": "*variable", "value": "~*req.2", "mandatory": true},
            ],
            "failed_calls_prefix": "randomPrefix",
            "partial_record_cache": "1s",
            "partial_cache_expiry_action": "randomAction",
			"reconnects": 5,
			"max_reconnect_interval": "3m"
         }
	]
}
}`
	expectedERsCfg := &ERsCfg{
		Enabled:          true,
		SessionSConns:    []string{"*internal:*sessions"},
		EEsConns:         []string{},
		ConcurrentEvents: 1,
		Readers: []*EventReaderCfg{
			{
				ID:                   utils.MetaDefault,
				Type:                 utils.MetaNone,
				RunDelay:             0,
				ConcurrentReqs:       1024,
				SourcePath:           "/var/spool/cgrates/ers/in",
				ProcessedPath:        "/var/spool/cgrates/ers/out",
				Tenant:               nil,
				Timezone:             utils.EmptyString,
				Filters:              []string{},
				Flags:                utils.FlagsWithParams{},
				Reconnects:           -1,
				MaxReconnectInterval: 5 * time.Minute,
				Fields: []*FCTemplate{
					{Tag: utils.ToR, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.2", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.OriginID, Path: utils.MetaCgreq + utils.NestingSep + utils.OriginID, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.3", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.RequestType, Path: utils.MetaCgreq + utils.NestingSep + utils.RequestType, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.4", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Tenant, Path: utils.MetaCgreq + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.6", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Category, Path: utils.MetaCgreq + utils.NestingSep + utils.Category, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.7", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AccountField, Path: utils.MetaCgreq + utils.NestingSep + utils.AccountField, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.8", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Subject, Path: utils.MetaCgreq + utils.NestingSep + utils.Subject, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.9", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Destination, Path: utils.MetaCgreq + utils.NestingSep + utils.Destination, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.10", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.SetupTime, Path: utils.MetaCgreq + utils.NestingSep + utils.SetupTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.11", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AnswerTime, Path: utils.MetaCgreq + utils.NestingSep + utils.AnswerTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.12", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Usage, Path: utils.MetaCgreq + utils.NestingSep + utils.Usage, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.13", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
				},
				CacheDumpFields:     make([]*FCTemplate, 0),
				PartialCommitFields: make([]*FCTemplate, 0),
				EEsIDs:              []string{},
				EEsSuccessIDs:       []string{},
				EEsFailedIDs:        []string{},
				Opts: &EventReaderOpts{
					CSV: &CSVROpts{
						HeaderDefineChar: utils.StringPointer(utils.InInFieldSep),
						RowLength:        utils.IntPointer(0),
						FieldSeparator:   utils.StringPointer(utils.FieldsSep),
					},
					AMQP:               &AMQPROpts{},
					AWS:                &AWSROpts{},
					SQL:                &SQLROpts{},
					Kafka:              &KafkaROpts{},
					PartialOrderField:  utils.StringPointer("~*req.AnswerTime"),
					PartialCacheAction: utils.StringPointer(utils.MetaNone),
					NATS: &NATSROpts{
						Subject: utils.StringPointer("cgrates_cdrs"),
					},
				},
			},
			{
				ID:                   "file_reader1",
				Type:                 "*file_csv",
				RunDelay:             -1,
				ConcurrentReqs:       1024,
				SourcePath:           "/tmp/ers/in",
				ProcessedPath:        "/tmp/ers/out",
				Tenant:               NewRSRParsersMustCompile("~*req.Destination1", utils.InfieldSep),
				Timezone:             utils.EmptyString,
				Filters:              []string{"randomFiletrs"},
				Flags:                utils.FlagsWithParams{},
				Reconnects:           5,
				MaxReconnectInterval: 3 * time.Minute,
				Fields: []*FCTemplate{
					{Tag: utils.ToR, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.2", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.OriginID, Path: utils.MetaCgreq + utils.NestingSep + utils.OriginID, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.3", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
				},
				CacheDumpFields: []*FCTemplate{
					{Tag: utils.ToR, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.2", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
				},
				PartialCommitFields: make([]*FCTemplate, 0),
				EEsIDs:              []string{},
				EEsSuccessIDs:       []string{},
				EEsFailedIDs:        []string{},
				Opts: &EventReaderOpts{
					CSV: &CSVROpts{
						FieldSeparator:   utils.StringPointer(utils.FieldsSep),
						HeaderDefineChar: utils.StringPointer(utils.InInFieldSep),
						RowLength:        utils.IntPointer(0),
					},
					AMQP:               &AMQPROpts{},
					AWS:                &AWSROpts{},
					SQL:                &SQLROpts{},
					Kafka:              &KafkaROpts{},
					PartialOrderField:  utils.StringPointer("~*req.AnswerTime"),
					PartialCacheAction: utils.StringPointer(utils.MetaNone),
					NATS: &NATSROpts{
						Subject: utils.StringPointer("cgrates_cdrs"),
					},
				},
			},
		},
		PartialCacheTTL: time.Second,
	}
	for _, profile := range expectedERsCfg.Readers {
		for _, v := range profile.Fields {
			v.ComputePath()
		}
		for _, v := range profile.CacheDumpFields {
			v.ComputePath()
		}
	}
	if jsonCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else {
		clonedErs := jsonCfg.ersCfg.Clone()
		if !reflect.DeepEqual(clonedErs, expectedERsCfg) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedERsCfg), utils.ToJSON(clonedErs))
		}
	}
}

func TestEventReaderloadFromJsonCfg(t *testing.T) {
	jsonCfg := NewDefaultCGRConfig()

	eventReader := new(EventReaderCfg)
	if err := eventReader.loadFromJSONCfg(nil, jsonCfg.templates, jsonCfg.generalCfg.RSRSep); err != nil {
		t.Error(err)
	}
}

func TestEventReaderloadFromJsonCase1(t *testing.T) {
	cfgJSON := &ERsJsonCfg{
		Readers: &[]*EventReaderJsonCfg{
			{
				Run_delay: utils.StringPointer("1ss"),
			},
		},
	}
	expected := "time: unknown unit \"ss\" in duration \"1ss\""
	jsoncfg := NewDefaultCGRConfig()
	if err := jsoncfg.ersCfg.loadFromJSONCfg(cfgJSON, jsoncfg.templates, jsoncfg.generalCfg.RSRSep, jsoncfg.dfltEvRdr); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	cfgJson := &ERsJsonCfg{

		Partial_cache_ttl: utils.StringPointer("test"),
	}

	if err := jsoncfg.ersCfg.loadFromJSONCfg(cfgJson, jsoncfg.templates, jsoncfg.generalCfg.RSRSep, jsoncfg.dfltEvRdr); err == nil {
		t.Error(err)
	}
}

func TestEventReaderloadFromJsonCase3(t *testing.T) {
	cfgJSON := &ERsJsonCfg{
		Readers: &[]*EventReaderJsonCfg{
			{
				Tenant: utils.StringPointer("a{*"),
			},
		},
	}
	expected := "invalid converter terminator in rule: <a{*>"
	jsoncfg := NewDefaultCGRConfig()
	if err := jsoncfg.ersCfg.loadFromJSONCfg(cfgJSON, jsoncfg.templates, jsoncfg.generalCfg.RSRSep, jsoncfg.dfltEvRdr); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestERSLoadFromjsonCfg(t *testing.T) {
	expectedERsCfg := &ERsCfg{
		Enabled:          true,
		SessionSConns:    []string{"conn1", "conn3"},
		EEsConns:         []string{},
		ConcurrentEvents: 1,
		Readers: []*EventReaderCfg{
			{
				ID:                   utils.MetaDefault,
				Type:                 utils.MetaNone,
				RunDelay:             0,
				ConcurrentReqs:       1024,
				SourcePath:           "/var/spool/cgrates/ers/in",
				ProcessedPath:        "/var/spool/cgrates/ers/out",
				Tenant:               nil,
				Timezone:             utils.EmptyString,
				Filters:              []string{},
				Flags:                utils.FlagsWithParams{},
				Reconnects:           -1,
				MaxReconnectInterval: 5 * time.Minute,
				Fields: []*FCTemplate{
					{Tag: utils.ToR, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.2", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.OriginID, Path: utils.MetaCgreq + utils.NestingSep + utils.OriginID, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.3", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.RequestType, Path: utils.MetaCgreq + utils.NestingSep + utils.RequestType, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.4", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Tenant, Path: utils.MetaCgreq + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.6", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Category, Path: utils.MetaCgreq + utils.NestingSep + utils.Category, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.7", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AccountField, Path: utils.MetaCgreq + utils.NestingSep + utils.AccountField, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.8", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Subject, Path: utils.MetaCgreq + utils.NestingSep + utils.Subject, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.9", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Destination, Path: utils.MetaCgreq + utils.NestingSep + utils.Destination, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.10", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.SetupTime, Path: utils.MetaCgreq + utils.NestingSep + utils.SetupTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.11", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AnswerTime, Path: utils.MetaCgreq + utils.NestingSep + utils.AnswerTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.12", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Usage, Path: utils.MetaCgreq + utils.NestingSep + utils.Usage, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.13", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
				},
				CacheDumpFields:     make([]*FCTemplate, 0),
				PartialCommitFields: make([]*FCTemplate, 0),
				EEsIDs:              []string{},
				EEsSuccessIDs:       []string{},
				EEsFailedIDs:        []string{},
				Opts: &EventReaderOpts{
					CSV: &CSVROpts{
						FieldSeparator:   utils.StringPointer(utils.FieldsSep),
						HeaderDefineChar: utils.StringPointer(utils.InInFieldSep),
						RowLength:        utils.IntPointer(0),
					},
					AMQP:               &AMQPROpts{},
					AWS:                &AWSROpts{},
					Kafka:              &KafkaROpts{},
					SQL:                &SQLROpts{},
					PartialOrderField:  utils.StringPointer("~*req.AnswerTime"),
					PartialCacheAction: utils.StringPointer(utils.MetaNone),
					NATS: &NATSROpts{
						Subject: utils.StringPointer("cgrates_cdrs"),
					},
				},
			},
			{
				ID:                   "file_reader1",
				Type:                 utils.MetaFileCSV,
				RunDelay:             -1,
				ConcurrentReqs:       1024,
				SourcePath:           "/tmp/ers/in",
				ProcessedPath:        "/tmp/ers/out",
				Tenant:               nil,
				Timezone:             utils.EmptyString,
				Filters:              []string{},
				Flags:                utils.FlagsWithParams{},
				Reconnects:           5,
				MaxReconnectInterval: 3 * time.Minute,
				Fields: []*FCTemplate{
					{Tag: utils.ToR, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.2", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.OriginID, Path: utils.MetaCgreq + utils.NestingSep + utils.OriginID, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.3", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.RequestType, Path: utils.MetaCgreq + utils.NestingSep + utils.RequestType, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.4", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Tenant, Path: utils.MetaCgreq + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.6", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Category, Path: utils.MetaCgreq + utils.NestingSep + utils.Category, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.7", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AccountField, Path: utils.MetaCgreq + utils.NestingSep + utils.AccountField, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.8", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Subject, Path: utils.MetaCgreq + utils.NestingSep + utils.Subject, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.9", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Destination, Path: utils.MetaCgreq + utils.NestingSep + utils.Destination, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.10", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.SetupTime, Path: utils.MetaCgreq + utils.NestingSep + utils.SetupTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.11", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AnswerTime, Path: utils.MetaCgreq + utils.NestingSep + utils.AnswerTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.12", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Usage, Path: utils.MetaCgreq + utils.NestingSep + utils.Usage, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.13", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
				},
				CacheDumpFields:     make([]*FCTemplate, 0),
				PartialCommitFields: make([]*FCTemplate, 0),
				EEsIDs:              []string{},
				EEsSuccessIDs:       []string{},
				EEsFailedIDs:        []string{},
				Opts: &EventReaderOpts{
					CSV: &CSVROpts{
						FieldSeparator:   utils.StringPointer(utils.FieldsSep),
						HeaderDefineChar: utils.StringPointer(utils.InInFieldSep),
						RowLength:        utils.IntPointer(0),
					},
					AMQP:               &AMQPROpts{},
					AWS:                &AWSROpts{},
					SQL:                &SQLROpts{},
					Kafka:              &KafkaROpts{},
					PartialOrderField:  utils.StringPointer("~*req.AnswerTime"),
					PartialCacheAction: utils.StringPointer(utils.MetaNone),
					NATS: &NATSROpts{
						Subject: utils.StringPointer("cgrates_cdrs"),
					},
				},
			},
		},
		PartialCacheTTL: time.Second,
	}
	for _, profile := range expectedERsCfg.Readers {
		for _, v := range profile.Fields {
			v.ComputePath()
		}
	}
	cfgJSONStr := `{
"ers": {
	"enabled": true,
	"sessions_conns":["conn1","conn3"],
	"readers": [
		{
			"id": "file_reader1",
			"run_delay":  "-1",
			"type": "*file_csv",
			"source_path": "/tmp/ers/in",
			"processed_path": "/tmp/ers/out",
			"cache_dump_fields": [],
			"reconnects": 5,
			"max_reconnect_interval": "3m"
		}
	]
}
}`

	if cfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedERsCfg, cfg.ersCfg) {
		t.Errorf("Expected: %+v ,\n received: %+v", utils.ToJSON(expectedERsCfg), utils.ToJSON(cfg.ersCfg))
	}

}

func TestERSloadFromJsonCfg(t *testing.T) {
	cfgJSON := &ERsJsonCfg{
		Readers: nil,
	}
	jsonCfg := NewDefaultCGRConfig()
	if err := jsonCfg.ersCfg.loadFromJSONCfg(cfgJSON, jsonCfg.templates, jsonCfg.generalCfg.RSRSep, jsonCfg.dfltEvRdr); err != nil {
		t.Error(err)
	}
}

func TestEventReaderFieldsloadFromJsonCfg(t *testing.T) {
	cfgJSON := &ERsJsonCfg{
		Readers: &[]*EventReaderJsonCfg{
			{
				Fields: &[]*FcTemplateJsonCfg{
					{
						Value: utils.StringPointer("a{*"),
					},
				},
			},
		},
	}
	expected := "invalid converter terminator in rule: <a{*>"
	jsonCfg := NewDefaultCGRConfig()
	if err := jsonCfg.ersCfg.loadFromJSONCfg(cfgJSON, jsonCfg.templates, jsonCfg.generalCfg.RSRSep, jsonCfg.dfltEvRdr); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestERSloadFromJsonCase1(t *testing.T) {
	cfgJSON := &ERsJsonCfg{
		Readers: &[]*EventReaderJsonCfg{
			{
				Fields: &[]*FcTemplateJsonCfg{
					{

						Type: utils.StringPointer(utils.MetaTemplate),
					},
				},
			},
		},
	}
	expected := "no template with id: <>"
	jsonCfg := NewDefaultCGRConfig()
	if err := jsonCfg.ersCfg.loadFromJSONCfg(cfgJSON, jsonCfg.templates, jsonCfg.generalCfg.RSRSep, jsonCfg.dfltEvRdr); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}
func TestERSloadFromJsonCase2(t *testing.T) {
	cfgJSON := &ERsJsonCfg{
		Readers: &[]*EventReaderJsonCfg{
			{
				Cache_dump_fields: &[]*FcTemplateJsonCfg{
					{

						Type: utils.StringPointer(utils.MetaTemplate),
					},
				},
			},
		},
	}
	expected := "no template with id: <>"
	jsonCfg := NewDefaultCGRConfig()
	if err := jsonCfg.ersCfg.loadFromJSONCfg(cfgJSON, jsonCfg.templates, jsonCfg.generalCfg.RSRSep, jsonCfg.dfltEvRdr); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestERSloadFromJsonCase3(t *testing.T) {
	cfgJSON := &ERsJsonCfg{
		Enabled:           utils.BoolPointer(true),
		Sessions_conns:    &[]string{"*conn1"},
		Concurrent_events: utils.IntPointer(1),
		Readers: &[]*EventReaderJsonCfg{
			{
				Id:                     utils.StringPointer("file_reader1"),
				Type:                   utils.StringPointer(utils.MetaFileCSV),
				Run_delay:              utils.StringPointer("-1"),
				Concurrent_requests:    utils.IntPointer(1024),
				Source_path:            utils.StringPointer("/tmp/ers/in"),
				Processed_path:         utils.StringPointer("/tmp/ers/out"),
				Tenant:                 nil,
				Timezone:               utils.StringPointer(""),
				Filters:                nil,
				Flags:                  &[]string{},
				Reconnects:             utils.IntPointer(5),
				Max_reconnect_interval: utils.StringPointer("3m"),
				Fields: &[]*FcTemplateJsonCfg{
					{
						Tag:    utils.StringPointer(utils.AnswerTime),
						Path:   utils.StringPointer("*cgreq.AnswerTime"),
						Type:   utils.StringPointer(utils.MetaTemplate),
						Value:  utils.StringPointer("randomTemplate"),
						Layout: utils.StringPointer("2006-01-02T15:04:05Z07:00"),
					},
				},
			},
		},
	}
	expectedERsCfg := &ERsCfg{
		Enabled:          true,
		SessionSConns:    []string{"*conn1"},
		EEsConns:         []string{},
		ConcurrentEvents: 1,
		Readers: []*EventReaderCfg{
			{
				ID:                   utils.MetaDefault,
				Type:                 utils.MetaNone,
				RunDelay:             0,
				ConcurrentReqs:       1024,
				SourcePath:           "/var/spool/cgrates/ers/in",
				ProcessedPath:        "/var/spool/cgrates/ers/out",
				Tenant:               nil,
				Timezone:             utils.EmptyString,
				Filters:              []string{},
				Flags:                utils.FlagsWithParams{},
				Reconnects:           -1,
				MaxReconnectInterval: 5 * time.Minute,
				Fields: []*FCTemplate{
					{Tag: utils.ToR, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.2", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.OriginID, Path: utils.MetaCgreq + utils.NestingSep + utils.OriginID, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.3", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.RequestType, Path: utils.MetaCgreq + utils.NestingSep + utils.RequestType, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.4", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Tenant, Path: utils.MetaCgreq + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.6", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Category, Path: utils.MetaCgreq + utils.NestingSep + utils.Category, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.7", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AccountField, Path: utils.MetaCgreq + utils.NestingSep + utils.AccountField, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.8", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Subject, Path: utils.MetaCgreq + utils.NestingSep + utils.Subject, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.9", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Destination, Path: utils.MetaCgreq + utils.NestingSep + utils.Destination, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.10", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.SetupTime, Path: utils.MetaCgreq + utils.NestingSep + utils.SetupTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.11", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AnswerTime, Path: utils.MetaCgreq + utils.NestingSep + utils.AnswerTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.12", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Usage, Path: utils.MetaCgreq + utils.NestingSep + utils.Usage, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.13", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
				},
				CacheDumpFields:     make([]*FCTemplate, 0),
				PartialCommitFields: make([]*FCTemplate, 0),
				EEsIDs:              []string{},
				EEsSuccessIDs:       []string{},
				EEsFailedIDs:        []string{},
				Opts: &EventReaderOpts{
					CSV: &CSVROpts{
						FieldSeparator:   utils.StringPointer(utils.FieldsSep),
						HeaderDefineChar: utils.StringPointer(utils.InInFieldSep),
						RowLength:        utils.IntPointer(0),
					},
					AMQP:               &AMQPROpts{},
					AWS:                &AWSROpts{},
					Kafka:              &KafkaROpts{},
					SQL:                &SQLROpts{},
					PartialOrderField:  utils.StringPointer("~*req.AnswerTime"),
					PartialCacheAction: utils.StringPointer(utils.MetaNone),
					NATS: &NATSROpts{
						Subject: utils.StringPointer("cgrates_cdrs"),
					},
				},
			},
			{
				ID:             "file_reader1",
				Type:           utils.MetaFileCSV,
				RunDelay:       -1,
				ConcurrentReqs: 1024,
				SourcePath:     "/tmp/ers/in",
				ProcessedPath:  "/tmp/ers/out",
				Tenant:         nil,
				Timezone:       utils.EmptyString,
				Filters:        []string{},
				Flags:          utils.FlagsWithParams{},
				Fields: []*FCTemplate{
					{
						Tag:    utils.CGRID,
						Path:   "*exp.CGRID",
						Type:   utils.MetaVariable,
						Layout: time.RFC3339,
					},
				},
				CacheDumpFields:      make([]*FCTemplate, 0),
				PartialCommitFields:  make([]*FCTemplate, 0),
				Reconnects:           5,
				MaxReconnectInterval: 3 * time.Minute,
				EEsIDs:               []string{},
				EEsSuccessIDs:        []string{},
				EEsFailedIDs:         []string{},
				Opts: &EventReaderOpts{
					CSV: &CSVROpts{
						FieldSeparator:   utils.StringPointer(utils.FieldsSep),
						HeaderDefineChar: utils.StringPointer(utils.InInFieldSep),
						RowLength:        utils.IntPointer(0),
					},
					AMQP:               &AMQPROpts{},
					AWS:                &AWSROpts{},
					SQL:                &SQLROpts{},
					Kafka:              &KafkaROpts{},
					PartialOrderField:  utils.StringPointer("~*req.AnswerTime"),
					PartialCacheAction: utils.StringPointer(utils.MetaNone),
					NATS: &NATSROpts{
						Subject: utils.StringPointer("cgrates_cdrs"),
					},
				},
			},
		},

		PartialCacheTTL: time.Second,
	}
	msgTemplates := map[string][]*FCTemplate{
		"randomTemplate": {
			{
				Tag:    utils.CGRID,
				Path:   "*exp.CGRID",
				Type:   utils.MetaVariable,
				Layout: time.RFC3339,
			},
		},
	}
	for _, v := range expectedERsCfg.Readers[0].Fields {
		v.ComputePath()
	}
	jsonCfg := NewDefaultCGRConfig()
	if err := jsonCfg.ersCfg.loadFromJSONCfg(cfgJSON, msgTemplates, jsonCfg.generalCfg.RSRSep, jsonCfg.dfltEvRdr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedERsCfg, jsonCfg.ersCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedERsCfg), utils.ToJSON(jsonCfg.ersCfg))
	}
}

func TestERSloadFromJsonCase4(t *testing.T) {
	cfgJSON := &ERsJsonCfg{
		Enabled:        utils.BoolPointer(true),
		Sessions_conns: &[]string{"*conn1"},
		Readers: &[]*EventReaderJsonCfg{
			{
				Id:                     utils.StringPointer("file_reader1"),
				Type:                   utils.StringPointer(utils.MetaFileCSV),
				Run_delay:              utils.StringPointer("-1"),
				Concurrent_requests:    utils.IntPointer(1024),
				Source_path:            utils.StringPointer("/tmp/ers/in"),
				Processed_path:         utils.StringPointer("/tmp/ers/out"),
				Tenant:                 nil,
				Timezone:               utils.StringPointer(""),
				Filters:                nil,
				Flags:                  &[]string{},
				Fields:                 &[]*FcTemplateJsonCfg{},
				Reconnects:             utils.IntPointer(5),
				Max_reconnect_interval: utils.StringPointer("3m"),
				Cache_dump_fields: &[]*FcTemplateJsonCfg{
					{
						Tag:   utils.StringPointer("OriginID"),
						Path:  utils.StringPointer("*exp.OriginID"),
						Type:  utils.StringPointer(utils.MetaTemplate),
						Value: utils.StringPointer("randomTemplate"),
					},
				},
			},
		},
	}
	expectedERsCfg := &ERsCfg{
		Enabled:          true,
		SessionSConns:    []string{"*conn1"},
		EEsConns:         []string{},
		ConcurrentEvents: 1,
		Readers: []*EventReaderCfg{
			{
				ID:                   utils.MetaDefault,
				Type:                 utils.MetaNone,
				RunDelay:             0,
				ConcurrentReqs:       1024,
				SourcePath:           "/var/spool/cgrates/ers/in",
				ProcessedPath:        "/var/spool/cgrates/ers/out",
				Tenant:               nil,
				Timezone:             utils.EmptyString,
				Filters:              []string{},
				Flags:                utils.FlagsWithParams{},
				Reconnects:           -1,
				MaxReconnectInterval: 5 * time.Minute,
				Fields: []*FCTemplate{
					{Tag: utils.ToR, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.2", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.OriginID, Path: utils.MetaCgreq + utils.NestingSep + utils.OriginID, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.3", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.RequestType, Path: utils.MetaCgreq + utils.NestingSep + utils.RequestType, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.4", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Tenant, Path: utils.MetaCgreq + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.6", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Category, Path: utils.MetaCgreq + utils.NestingSep + utils.Category, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.7", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AccountField, Path: utils.MetaCgreq + utils.NestingSep + utils.AccountField, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.8", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Subject, Path: utils.MetaCgreq + utils.NestingSep + utils.Subject, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.9", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Destination, Path: utils.MetaCgreq + utils.NestingSep + utils.Destination, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.10", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.SetupTime, Path: utils.MetaCgreq + utils.NestingSep + utils.SetupTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.11", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AnswerTime, Path: utils.MetaCgreq + utils.NestingSep + utils.AnswerTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.12", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Usage, Path: utils.MetaCgreq + utils.NestingSep + utils.Usage, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.13", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
				},
				CacheDumpFields:     make([]*FCTemplate, 0),
				PartialCommitFields: make([]*FCTemplate, 0),
				EEsIDs:              []string{},
				EEsSuccessIDs:       []string{},
				EEsFailedIDs:        []string{},
				Opts: &EventReaderOpts{
					CSV: &CSVROpts{
						FieldSeparator:   utils.StringPointer(utils.FieldsSep),
						HeaderDefineChar: utils.StringPointer(utils.InInFieldSep),
						RowLength:        utils.IntPointer(0),
					},
					AMQP:               &AMQPROpts{},
					AWS:                &AWSROpts{},
					Kafka:              &KafkaROpts{},
					SQL:                &SQLROpts{},
					PartialOrderField:  utils.StringPointer("~*req.AnswerTime"),
					PartialCacheAction: utils.StringPointer(utils.MetaNone),
					NATS: &NATSROpts{
						Subject: utils.StringPointer("cgrates_cdrs"),
					},
				},
			},
			{
				ID:             "file_reader1",
				Type:           utils.MetaFileCSV,
				RunDelay:       -1,
				ConcurrentReqs: 1024,
				SourcePath:     "/tmp/ers/in",
				ProcessedPath:  "/tmp/ers/out",
				Tenant:         nil,
				Timezone:       utils.EmptyString,
				Filters:        []string{},
				Flags:          utils.FlagsWithParams{},
				Fields:         []*FCTemplate{},
				EEsIDs:         []string{},
				EEsSuccessIDs:  []string{},
				EEsFailedIDs:   []string{},
				CacheDumpFields: []*FCTemplate{
					{
						Tag:   "OrderID",
						Path:  "*exp.OrderID",
						Type:  "*variable",
						Value: NewRSRParsersMustCompile("~*req.OrderID", utils.InfieldSep),
					},
				},
				PartialCommitFields:  make([]*FCTemplate, 0),
				Reconnects:           5,
				MaxReconnectInterval: 3 * time.Minute,
				Opts: &EventReaderOpts{
					CSV: &CSVROpts{
						FieldSeparator:   utils.StringPointer(utils.FieldsSep),
						HeaderDefineChar: utils.StringPointer(utils.InInFieldSep),
						RowLength:        utils.IntPointer(0),
					},
					AWS:                &AWSROpts{},
					AMQP:               &AMQPROpts{},
					SQL:                &SQLROpts{},
					Kafka:              &KafkaROpts{},
					PartialOrderField:  utils.StringPointer("~*req.AnswerTime"),
					PartialCacheAction: utils.StringPointer(utils.MetaNone),
					NATS: &NATSROpts{
						Subject: utils.StringPointer("cgrates_cdrs"),
					},
				},
			},
		},

		PartialCacheTTL: time.Second,
	}
	msgTemplates := map[string][]*FCTemplate{
		"randomTemplate": {
			{
				Tag:   "OrderID",
				Path:  "*exp.OrderID",
				Type:  "*variable",
				Value: NewRSRParsersMustCompile("~*req.OrderID", utils.InfieldSep),
			},
		},
	}
	for _, v := range expectedERsCfg.Readers[0].CacheDumpFields {
		v.ComputePath()
	}
	for _, v := range expectedERsCfg.Readers[0].Fields {
		v.ComputePath()
	}
	jsonCfg := NewDefaultCGRConfig()
	if err := jsonCfg.ersCfg.loadFromJSONCfg(cfgJSON, msgTemplates, jsonCfg.generalCfg.RSRSep, jsonCfg.dfltEvRdr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedERsCfg, jsonCfg.ersCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedERsCfg), utils.ToJSON(jsonCfg.ersCfg))
	}
}

func TestEventReaderCacheDumpFieldsloadFromJsonCfg(t *testing.T) {
	cfgJSON := &ERsJsonCfg{
		Readers: &[]*EventReaderJsonCfg{
			{
				Cache_dump_fields: &[]*FcTemplateJsonCfg{
					{
						Value: utils.StringPointer("a{*"),
					},
				},
			},
		},
	}
	expected := "invalid converter terminator in rule: <a{*>"
	jsonCfg := NewDefaultCGRConfig()
	if err := jsonCfg.ersCfg.loadFromJSONCfg(cfgJSON, jsonCfg.templates, jsonCfg.generalCfg.RSRSep, jsonCfg.dfltEvRdr); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestEventReaderSameID(t *testing.T) {
	expectedERsCfg := &ERsCfg{
		Enabled:          true,
		SessionSConns:    []string{"conn1"},
		EEsConns:         []string{},
		ConcurrentEvents: 1,
		Readers: []*EventReaderCfg{
			{
				ID:                   utils.MetaDefault,
				Type:                 utils.MetaNone,
				RunDelay:             0,
				ConcurrentReqs:       1024,
				SourcePath:           "/var/spool/cgrates/ers/in",
				ProcessedPath:        "/var/spool/cgrates/ers/out",
				Tenant:               nil,
				Timezone:             utils.EmptyString,
				Filters:              []string{},
				Flags:                utils.FlagsWithParams{},
				Reconnects:           -1,
				MaxReconnectInterval: 5 * time.Minute,
				EEsIDs:               []string{},
				EEsSuccessIDs:        []string{},
				EEsFailedIDs:         []string{},
				Fields: []*FCTemplate{
					{Tag: utils.ToR, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.2", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.OriginID, Path: utils.MetaCgreq + utils.NestingSep + utils.OriginID, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.3", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.RequestType, Path: utils.MetaCgreq + utils.NestingSep + utils.RequestType, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.4", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Tenant, Path: utils.MetaCgreq + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.6", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Category, Path: utils.MetaCgreq + utils.NestingSep + utils.Category, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.7", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AccountField, Path: utils.MetaCgreq + utils.NestingSep + utils.AccountField, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.8", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Subject, Path: utils.MetaCgreq + utils.NestingSep + utils.Subject, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.9", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Destination, Path: utils.MetaCgreq + utils.NestingSep + utils.Destination, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.10", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.SetupTime, Path: utils.MetaCgreq + utils.NestingSep + utils.SetupTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.11", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AnswerTime, Path: utils.MetaCgreq + utils.NestingSep + utils.AnswerTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.12", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Usage, Path: utils.MetaCgreq + utils.NestingSep + utils.Usage, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.13", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
				},
				CacheDumpFields:     make([]*FCTemplate, 0),
				PartialCommitFields: make([]*FCTemplate, 0),
				Opts: &EventReaderOpts{
					CSV: &CSVROpts{
						FieldSeparator:   utils.StringPointer(utils.FieldsSep),
						HeaderDefineChar: utils.StringPointer(utils.InInFieldSep),
						RowLength:        utils.IntPointer(0),
					},
					Kafka:              &KafkaROpts{},
					AMQP:               &AMQPROpts{},
					AWS:                &AWSROpts{},
					SQL:                &SQLROpts{},
					PartialOrderField:  utils.StringPointer("~*req.AnswerTime"),
					PartialCacheAction: utils.StringPointer(utils.MetaNone),
					NATS: &NATSROpts{
						Subject: utils.StringPointer("cgrates_cdrs"),
					},
				},
			},
			{
				ID:             "file_reader1",
				Type:           utils.MetaFileCSV,
				RunDelay:       -1,
				ConcurrentReqs: 1024,
				SourcePath:     "/tmp/ers/in",
				ProcessedPath:  "/tmp/ers/out",
				Tenant:         nil,
				Timezone:       utils.EmptyString,
				Filters:        []string{},
				Flags:          utils.FlagsWithParams{},
				EEsIDs:         []string{},
				EEsSuccessIDs:  []string{},
				EEsFailedIDs:   []string{},
				Fields: []*FCTemplate{
					{Tag: "CustomTag2", Path: "CustomPath2", Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("CustomValue2", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
				},
				CacheDumpFields:      make([]*FCTemplate, 0),
				PartialCommitFields:  make([]*FCTemplate, 0),
				Reconnects:           5,
				MaxReconnectInterval: 3 * time.Minute,
				Opts: &EventReaderOpts{
					CSV: &CSVROpts{
						FieldSeparator:   utils.StringPointer(utils.FieldsSep),
						HeaderDefineChar: utils.StringPointer(utils.InInFieldSep),
						RowLength:        utils.IntPointer(0),
					},
					AMQP:               &AMQPROpts{},
					AWS:                &AWSROpts{},
					SQL:                &SQLROpts{},
					Kafka:              &KafkaROpts{},
					PartialOrderField:  utils.StringPointer("~*req.AnswerTime"),
					PartialCacheAction: utils.StringPointer(utils.MetaNone),
					NATS: &NATSROpts{
						Subject: utils.StringPointer("cgrates_cdrs"),
					},
				},
			},
		},

		PartialCacheTTL: time.Second,
	}
	for _, profile := range expectedERsCfg.Readers {
		for _, v := range profile.Fields {
			v.ComputePath()
		}
	}
	cfgJSONStr := `{
"ers": {
	"enabled": true,
	"sessions_conns":["conn1"],
	"readers": [
		{
			"id": "file_reader1",
			"run_delay":  "-1",
			"type": "*file_csv",
			"row_length" : 5,
			"source_path": "/tmp/ers/in",
			"processed_path": "/tmp/ers/out",
			"fields":[
				{"tag": "CustomTag1", "path": "CustomPath1", "type": "*variable", "value": "CustomValue1", "mandatory": true},
			],
			"cache_dump_fields": [],
		},
		{
			"id": "file_reader1",
			"run_delay":  "-1",
			"type": "*file_csv",
			"source_path": "/tmp/ers/in",
			"processed_path": "/tmp/ers/out",
			"fields":[
				{"tag": "CustomTag2", "path": "CustomPath2", "type": "*variable", "value": "CustomValue2", "mandatory": true},
			],
			"cache_dump_fields": [],
			"reconnects": 5,
			"max_reconnect_interval": "3m0s"
		},
	],
}
}`

	if cfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedERsCfg, cfg.ersCfg) {
		t.Errorf("Expected: %+v ,\n received: %+v", utils.ToJSON(expectedERsCfg), utils.ToJSON(cfg.ersCfg))
	}

}

func TestERsCfgAsMapInterfaceCase1(t *testing.T) {
	cfgJSONStr := `{
"ers": {
	"enabled": true,
	"sessions_conns":["conn1","conn3"],
	"readers": [
		{
			"id": "file_reader1",
			"run_delay":  "-1",
			"start_delay":"0",
			"tenant": "~*req.Destination1",
			"type": "*file_csv",
			"source_path": "/tmp/ers/in",
			"processed_path": "/tmp/ers/out",
			"cache_dump_fields": [],
			"reconnects": 5,
			"max_reconnect_interval": "3m"
		}
	]
}
}`
	eMap := map[string]any{
		utils.EnabledCfg:          true,
		utils.SessionSConnsCfg:    []string{"conn1", "conn3"},
		utils.EEsConnsCfg:         []string{},
		utils.ConcurrentEventsCfg: 1,
		utils.ReadersCfg: []map[string]any{
			{
				utils.FiltersCfg:              []string{},
				utils.FlagsCfg:                []string{},
				utils.IDCfg:                   "*default",
				utils.ProcessedPathCfg:        "/var/spool/cgrates/ers/out",
				utils.RunDelayCfg:             "0",
				utils.StartDelayCfg:           "0",
				utils.SourcePathCfg:           "/var/spool/cgrates/ers/in",
				utils.TenantCfg:               "",
				utils.TimezoneCfg:             "",
				utils.CacheDumpFieldsCfg:      []map[string]any{},
				utils.PartialCommitFieldsCfg:  []map[string]any{},
				utils.ConcurrentRequestsCfg:   1024,
				utils.TypeCfg:                 "*none",
				utils.ReconnectsCfg:           -1,
				utils.MaxReconnectIntervalCfg: "5m0s",
				utils.FieldsCfg: []map[string]any{
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.ToR", utils.TagCfg: "ToR", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.2"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.OriginID", utils.TagCfg: "OriginID", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.3"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.RequestType", utils.TagCfg: "RequestType", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.4"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Tenant", utils.TagCfg: "Tenant", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.6"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Category", utils.TagCfg: "Category", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.7"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Account", utils.TagCfg: "Account", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.8"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Subject", utils.TagCfg: "Subject", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.9"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Destination", utils.TagCfg: "Destination", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.10"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.SetupTime", utils.TagCfg: "SetupTime", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.11"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.AnswerTime", utils.TagCfg: "AnswerTime", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.12"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Usage", utils.TagCfg: "Usage", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.13"},
				},
				utils.OptsCfg: map[string]any{
					"csvFieldSeparator":   ",",
					"csvHeaderDefineChar": ":",
					"csvRowLength":        0,
					"partialOrderField":   "~*req.AnswerTime",
					"partialCacheAction":  utils.MetaNone,
					"natsSubject":         "cgrates_cdrs",
				},
			},
			{
				utils.CacheDumpFieldsCfg:     []map[string]any{},
				utils.PartialCommitFieldsCfg: []map[string]any{},
				utils.ConcurrentRequestsCfg:  1024,
				utils.TypeCfg:                "*file_csv",
				utils.FieldsCfg: []map[string]any{
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.ToR", utils.TagCfg: "ToR", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.2"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.OriginID", utils.TagCfg: "OriginID", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.3"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.RequestType", utils.TagCfg: "RequestType", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.4"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Tenant", utils.TagCfg: "Tenant", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.6"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Category", utils.TagCfg: "Category", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.7"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Account", utils.TagCfg: "Account", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.8"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Subject", utils.TagCfg: "Subject", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.9"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Destination", utils.TagCfg: "Destination", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.10"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.SetupTime", utils.TagCfg: "SetupTime", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.11"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.AnswerTime", utils.TagCfg: "AnswerTime", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.12"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Usage", utils.TagCfg: "Usage", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.13"},
				},
				utils.FiltersCfg:              []string{},
				utils.FlagsCfg:                []string{},
				utils.IDCfg:                   "file_reader1",
				utils.ProcessedPathCfg:        "/tmp/ers/out",
				utils.RunDelayCfg:             "-1",
				utils.StartDelayCfg:           "0",
				utils.SourcePathCfg:           "/tmp/ers/in",
				utils.TenantCfg:               "~*req.Destination1",
				utils.TimezoneCfg:             "",
				utils.ReconnectsCfg:           5,
				utils.MaxReconnectIntervalCfg: "3m0s",
				utils.OptsCfg: map[string]any{
					"csvFieldSeparator":   ",",
					"csvHeaderDefineChar": ":",
					"csvRowLength":        0,
					"partialOrderField":   "~*req.AnswerTime",
					"partialCacheAction":  utils.MetaNone,
					"natsSubject":         "cgrates_cdrs",
				},
			},
		},
		utils.PartialCacheTTLCfg: "1s",
	}
	if cfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cfg.ersCfg.AsMapInterface(utils.InfieldSep); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestERSCfgAsMapInterfaceCase2(t *testing.T) {
	cfgJSONStr := `{
	"ers": {
		"enabled": true,
		"sessions_conns":["conn1","conn3"],
		"readers": [
			{
                "id": "file_reader1",
				"run_delay":  "10s",
                "tenant": "~*req.Destination1",
				"type": "*file_csv",
                "flags": ["randomFlag"],
                "filters": ["randomFilter"],
				"source_path": "/tmp/ers/in",
                "partial_record_cache": "1s",
				"processed_path": "/tmp/ers/out",
				"cache_dump_fields": [
                           {"tag": "ToR", "path": "*cgreq.ToR", "type": "*variable", "value": "~*req.2", "mandatory": true}                
				],
				"opts":{
					"kafkaGroupID": "test",
					"csvLazyQuotes": false,
					"xmlRootPath": "root.A",
					"amqpQueueID": "id",
					"amqpConsumerTag": "tag",
					"amqpExchange":"exchange",
					"amqpExchangeType":"type",
					"amqpRoutingKey":"key",
					"partialPath":"optpath",
					"partialcsvFieldSeparator":",",
					"kafkaTopic":"topic",
					"kafkaMaxWait":"1m",
					"sqlDBName":"db",
					"sqlTableName":"table",
					"pgSSLMode":"pg",
					"awsRegion":"eu",
					"awsKey":"key",
					"awsSecret":"secret",
					"awsToken":"token",
					"sqsQueueID":"sqs",
					"s3BucketID":"s3bucket",
					"natsJetStream":true,
					"natsConsumerName": "NATConsumer",
					"natsStreamName": "NATStream",
					"natsQueueID":"NATid",
					"natsJWTFile":"jwt",
					"natsCertificateAuthority":"auth",
					"natsClientCertificate":"certificate",
					"natsSeedFile":"seed",
					"natsClientKey":"clientkey",
					"natsJetStreamMaxWait":"1m",
				},
			},
		],
	}
}`
	eMap := map[string]any{
		utils.EnabledCfg:          true,
		utils.SessionSConnsCfg:    []string{"conn1", "conn3"},
		utils.EEsConnsCfg:         []string{},
		utils.ConcurrentEventsCfg: 1,
		utils.ReadersCfg: []map[string]any{
			{
				utils.FiltersCfg:              []string{},
				utils.FlagsCfg:                []string{},
				utils.IDCfg:                   "*default",
				utils.ProcessedPathCfg:        "/var/spool/cgrates/ers/out",
				utils.RunDelayCfg:             "0",
				utils.StartDelayCfg:           "0",
				utils.SourcePathCfg:           "/var/spool/cgrates/ers/in",
				utils.TenantCfg:               "",
				utils.TimezoneCfg:             "",
				utils.CacheDumpFieldsCfg:      []map[string]any{},
				utils.PartialCommitFieldsCfg:  []map[string]any{},
				utils.ConcurrentRequestsCfg:   1024,
				utils.TypeCfg:                 "*none",
				utils.ReconnectsCfg:           -1,
				utils.MaxReconnectIntervalCfg: "5m0s",
				utils.FieldsCfg: []map[string]any{
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.ToR", utils.TagCfg: "ToR", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.2"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.OriginID", utils.TagCfg: "OriginID", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.3"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.RequestType", utils.TagCfg: "RequestType", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.4"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Tenant", utils.TagCfg: "Tenant", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.6"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Category", utils.TagCfg: "Category", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.7"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Account", utils.TagCfg: "Account", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.8"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Subject", utils.TagCfg: "Subject", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.9"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Destination", utils.TagCfg: "Destination", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.10"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.SetupTime", utils.TagCfg: "SetupTime", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.11"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.AnswerTime", utils.TagCfg: "AnswerTime", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.12"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Usage", utils.TagCfg: "Usage", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.13"},
				},
				utils.OptsCfg: map[string]any{
					"csvFieldSeparator":   ",",
					"csvHeaderDefineChar": ":",
					"csvRowLength":        0,
					"partialOrderField":   "~*req.AnswerTime",
					"partialCacheAction":  utils.MetaNone,
					"natsSubject":         "cgrates_cdrs",
				},
			},
			{
				utils.CacheDumpFieldsCfg: []map[string]any{
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.ToR", utils.TagCfg: "ToR", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.2"},
				},
				utils.PartialCommitFieldsCfg: []map[string]any{},
				utils.ConcurrentRequestsCfg:  1024,
				utils.TypeCfg:                "*file_csv",
				utils.FieldsCfg: []map[string]any{
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.ToR", utils.TagCfg: "ToR", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.2"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.OriginID", utils.TagCfg: "OriginID", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.3"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.RequestType", utils.TagCfg: "RequestType", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.4"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Tenant", utils.TagCfg: "Tenant", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.6"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Category", utils.TagCfg: "Category", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.7"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Account", utils.TagCfg: "Account", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.8"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Subject", utils.TagCfg: "Subject", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.9"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Destination", utils.TagCfg: "Destination", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.10"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.SetupTime", utils.TagCfg: "SetupTime", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.11"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.AnswerTime", utils.TagCfg: "AnswerTime", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.12"},
					{utils.MandatoryCfg: true, utils.PathCfg: "*cgreq.Usage", utils.TagCfg: "Usage", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.13"},
				},
				utils.FiltersCfg:              []string{"randomFilter"},
				utils.FlagsCfg:                []string{"randomFlag"},
				utils.IDCfg:                   "file_reader1",
				utils.ProcessedPathCfg:        "/tmp/ers/out",
				utils.RunDelayCfg:             "10s",
				utils.StartDelayCfg:           "0",
				utils.SourcePathCfg:           "/tmp/ers/in",
				utils.TenantCfg:               "~*req.Destination1",
				utils.TimezoneCfg:             "",
				utils.ReconnectsCfg:           -1,
				utils.MaxReconnectIntervalCfg: "5m0s",
				utils.OptsCfg: map[string]any{
					utils.CSVLazyQuotes:              false,
					utils.KafkaGroupID:               "test",
					utils.CSVFieldSepOpt:             ",",
					utils.XMLRootPathOpt:             "root.A",
					"csvHeaderDefineChar":            ":",
					"csvRowLength":                   0,
					"partialOrderField":              "~*req.AnswerTime",
					"partialCacheAction":             utils.MetaNone,
					"natsSubject":                    "cgrates_cdrs",
					utils.AMQPQueueID:                "id",
					utils.AMQPConsumerTag:            "tag",
					utils.AMQPExchange:               "exchange",
					utils.AMQPExchangeType:           "type",
					utils.AMQPRoutingKey:             "key",
					utils.PartialPathOpt:             "optpath",
					utils.PartialCSVFieldSepartorOpt: ",",
					utils.KafkaTopic:                 "topic",
					utils.KafkaMaxWait:               "1m0s",
					utils.SQLDBNameOpt:               "db",
					utils.SQLTableNameOpt:            "table",
					utils.PgSSLModeCfg:               "pg",
					utils.AWSRegion:                  "eu",
					utils.AWSKey:                     "key",
					utils.AWSSecret:                  "secret",
					utils.AWSToken:                   "token",
					utils.SQSQueueID:                 "sqs",
					utils.S3Bucket:                   "s3bucket",
					utils.NatsJetStream:              true,
					utils.NatsConsumerName:           "NATConsumer",
					utils.NatsStreamName:             "NATStream",
					utils.NatsQueueID:                "NATid",
					utils.NatsJWTFile:                "jwt",
					utils.NatsSeedFile:               "seed",
					utils.NatsCertificateAuthority:   "auth",
					utils.NatsClientCertificate:      "certificate",
					utils.NatsClientKey:              "clientkey",
					utils.NatsJetStreamMaxWait:       "1m0s",
				},
			},
		},
		utils.PartialCacheTTLCfg: "1s",
	}
	if cfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cfg.ersCfg.AsMapInterface(utils.InfieldSep); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestERsloadFromJsonCfg(t *testing.T) {
	cfgJSON := &ERsJsonCfg{
		Enabled:        utils.BoolPointer(true),
		Sessions_conns: &[]string{"*conn1"},
		Readers: &[]*EventReaderJsonCfg{
			{
				Id:                     utils.StringPointer("file_reader1"),
				Type:                   utils.StringPointer(utils.MetaFileCSV),
				Run_delay:              utils.StringPointer("-1"),
				Concurrent_requests:    utils.IntPointer(1024),
				Source_path:            utils.StringPointer("/tmp/ers/in"),
				Processed_path:         utils.StringPointer("/tmp/ers/out"),
				Tenant:                 nil,
				Timezone:               utils.StringPointer(""),
				Filters:                nil,
				Flags:                  &[]string{},
				Reconnects:             utils.IntPointer(-1),
				Max_reconnect_interval: utils.StringPointer("5m"),
				Fields: &[]*FcTemplateJsonCfg{
					{
						Tag:    utils.StringPointer(utils.CGRID),
						Path:   utils.StringPointer("*exp.CGRID"),
						Type:   utils.StringPointer(utils.MetaVariable),
						Layout: utils.StringPointer("2006-01-02T15:04:05Z07:00"),
					},
					{Tag: utils.StringPointer("CustomTag2"), Path: utils.StringPointer("CustomPath2"), Type: utils.StringPointer(utils.MetaVariable),
						Value: utils.StringPointer("CustomValue2"), Mandatory: utils.BoolPointer(true), Layout: utils.StringPointer("2006-01-02T15:04:05Z07:00")},
				},
			},
		},
	}
	expectedERsCfg := &ERsCfg{
		Enabled:          true,
		SessionSConns:    []string{"*conn1"},
		EEsConns:         []string{},
		ConcurrentEvents: 1,
		Readers: []*EventReaderCfg{
			{
				ID:                   utils.MetaDefault,
				Type:                 utils.MetaNone,
				RunDelay:             0,
				ConcurrentReqs:       1024,
				SourcePath:           "/var/spool/cgrates/ers/in",
				ProcessedPath:        "/var/spool/cgrates/ers/out",
				Tenant:               nil,
				Timezone:             utils.EmptyString,
				Filters:              []string{},
				Flags:                utils.FlagsWithParams{},
				Reconnects:           -1,
				MaxReconnectInterval: 5 * time.Minute,
				EEsIDs:               []string{},
				EEsSuccessIDs:        []string{},
				EEsFailedIDs:         []string{},
				Fields: []*FCTemplate{
					{Tag: utils.ToR, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.2", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.OriginID, Path: utils.MetaCgreq + utils.NestingSep + utils.OriginID, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.3", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.RequestType, Path: utils.MetaCgreq + utils.NestingSep + utils.RequestType, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.4", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Tenant, Path: utils.MetaCgreq + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.6", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Category, Path: utils.MetaCgreq + utils.NestingSep + utils.Category, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.7", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AccountField, Path: utils.MetaCgreq + utils.NestingSep + utils.AccountField, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.8", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Subject, Path: utils.MetaCgreq + utils.NestingSep + utils.Subject, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.9", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Destination, Path: utils.MetaCgreq + utils.NestingSep + utils.Destination, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.10", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.SetupTime, Path: utils.MetaCgreq + utils.NestingSep + utils.SetupTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.11", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AnswerTime, Path: utils.MetaCgreq + utils.NestingSep + utils.AnswerTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.12", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Usage, Path: utils.MetaCgreq + utils.NestingSep + utils.Usage, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.13", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
				},
				CacheDumpFields:     make([]*FCTemplate, 0),
				PartialCommitFields: make([]*FCTemplate, 0),
				Opts: &EventReaderOpts{
					CSV: &CSVROpts{
						FieldSeparator:   utils.StringPointer(utils.FieldsSep),
						HeaderDefineChar: utils.StringPointer(utils.InInFieldSep),
						RowLength:        utils.IntPointer(0),
					},
					Kafka:              &KafkaROpts{},
					AMQP:               &AMQPROpts{},
					AWS:                &AWSROpts{},
					SQL:                &SQLROpts{},
					PartialOrderField:  utils.StringPointer("~*req.AnswerTime"),
					PartialCacheAction: utils.StringPointer(utils.MetaNone),
					NATS: &NATSROpts{
						Subject: utils.StringPointer("cgrates_cdrs"),
					},
				},
			},
			{
				ID:                   "file_reader1",
				Type:                 utils.MetaFileCSV,
				RunDelay:             -1,
				ConcurrentReqs:       1024,
				SourcePath:           "/tmp/ers/in",
				ProcessedPath:        "/tmp/ers/out",
				Tenant:               nil,
				Timezone:             utils.EmptyString,
				Filters:              []string{},
				Flags:                utils.FlagsWithParams{},
				Reconnects:           -1,
				MaxReconnectInterval: 5 * time.Minute,
				EEsIDs:               []string{},
				EEsSuccessIDs:        []string{},
				EEsFailedIDs:         []string{},
				Fields: []*FCTemplate{
					{
						Tag:    utils.CGRID,
						Path:   "*exp.CGRID",
						Type:   utils.MetaVariable,
						Layout: time.RFC3339,
					},
					{Tag: "CustomTag2", Path: "CustomPath2", Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("CustomValue2", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
				},
				CacheDumpFields:     make([]*FCTemplate, 0),
				PartialCommitFields: make([]*FCTemplate, 0),
				Opts: &EventReaderOpts{
					CSV: &CSVROpts{
						FieldSeparator:   utils.StringPointer(utils.FieldsSep),
						HeaderDefineChar: utils.StringPointer(utils.InInFieldSep),
						RowLength:        utils.IntPointer(0),
					},
					AMQP:               &AMQPROpts{},
					AWS:                &AWSROpts{},
					SQL:                &SQLROpts{},
					Kafka:              &KafkaROpts{},
					PartialOrderField:  utils.StringPointer("~*req.AnswerTime"),
					PartialCacheAction: utils.StringPointer(utils.MetaNone),
					NATS: &NATSROpts{
						Subject: utils.StringPointer("cgrates_cdrs"),
					},
				},
			},
		},
		PartialCacheTTL: time.Second,
	}
	for _, profile := range expectedERsCfg.Readers {
		for _, v := range profile.Fields {
			v.ComputePath()
		}
		for _, v := range profile.CacheDumpFields {
			v.ComputePath()
		}
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.ersCfg.loadFromJSONCfg(cfgJSON, cfgCgr.templates, cfgCgr.generalCfg.RSRSep, cfgCgr.dfltEvRdr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgCgr.ersCfg, expectedERsCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedERsCfg), utils.ToJSON(cfgCgr.ersCfg))
	}
}
func TestEventReaderOptsCfg(t *testing.T) {
	erCfg := new(EventReaderCfg)
	if err := erCfg.Opts.loadFromJSONCfg(nil); err != nil {
		t.Error(err)
	}
	eventReaderOptsJson := &EventReaderOptsJson{
		PartialPath:              utils.StringPointer("path"),
		PartialCSVFieldSeparator: utils.StringPointer("/"),
		CSVLazyQuotes:            utils.BoolPointer(false),
		AMQPQueueID:              utils.StringPointer("id"),
		AMQPConsumerTag:          utils.StringPointer("tag"),
		AMQPExchange:             utils.StringPointer("exchange"),
		AMQPExchangeType:         utils.StringPointer("type"),
		AMQPRoutingKey:           utils.StringPointer("key1"),
		KafkaTopic:               utils.StringPointer("kafka"),
		KafkaMaxWait:             utils.StringPointer("1m"),
		SQLDBName:                utils.StringPointer("dbname"),
		SQLTableName:             utils.StringPointer("tablename"),
		SQLBatchSize:             utils.IntPointer(-1),
		SQLDeleteIndexedFields:   utils.SliceStringPointer([]string{"id"}),
		PgSSLMode:                utils.StringPointer("sslmode"),
		AWSRegion:                utils.StringPointer("eu"),
		AWSKey:                   utils.StringPointer("key"),
		AWSSecret:                utils.StringPointer("secret"),
		AWSToken:                 utils.StringPointer("token"),
		SQSQueueID:               utils.StringPointer("SQSQueue"),
		S3BucketID:               utils.StringPointer("S3BucketID"),
		NATSJetStream:            utils.BoolPointer(false),
		NATSConsumerName:         utils.StringPointer("user"),
		NATSQueueID:              utils.StringPointer("id"),
		NATSJWTFile:              utils.StringPointer("jwt"),
		NATSSeedFile:             utils.StringPointer("seed"),
		NATSCertificateAuthority: utils.StringPointer("authority"),
		NATSClientCertificate:    utils.StringPointer("certificate"),
		NATSClientKey:            utils.StringPointer("key5"),
		NATSJetStreamMaxWait:     utils.StringPointer("1m"),
		KafkaGroupID:             utils.StringPointer("groupId"),
	}
	eventReader := &EventReaderCfg{
		Opts: &EventReaderOpts{
			CSV:   &CSVROpts{},
			AMQP:  &AMQPROpts{},
			AWS:   &AWSROpts{},
			NATS:  &NATSROpts{},
			Kafka: &KafkaROpts{},
			SQL:   &SQLROpts{},
		},
	}
	if err := eventReader.Opts.loadFromJSONCfg(eventReaderOptsJson); err != nil {
		t.Error(err)
	}
	if err := eventReader.Opts.loadFromJSONCfg(&EventReaderOptsJson{
		KafkaMaxWait: utils.StringPointer("test"),
	}); err == nil {
		t.Error(err)
	} else if err := eventReader.Opts.loadFromJSONCfg(&EventReaderOptsJson{
		NATSJetStreamMaxWait: utils.StringPointer("nil"),
	}); err == nil {
		t.Error(err)
	}

}

func TestEventReaderCfgClone(t *testing.T) {
	ban := &EventReaderCfg{
		ID:             "2",
		Type:           "type",
		RunDelay:       1 * time.Minute,
		ConcurrentReqs: 5,
		PartialCommitFields: []*FCTemplate{
			{
				Tag:  "tag1",
				Type: "type1",
			},
			{
				Tag:  "tag2",
				Type: "type2",
			},
		},
		SourcePath:    "/",
		ProcessedPath: "/path",
		Tenant:        RSRParsers{},
		Timezone:      "time.Utc",
		Flags:         utils.FlagsWithParams{},
		Opts: &EventReaderOpts{
			PartialPath: utils.StringPointer("path"),
			CSV: &CSVROpts{
				PartialCSVFieldSeparator: utils.StringPointer("/"),
				LazyQuotes:               utils.BoolPointer(false),
			},
			AMQP: &AMQPROpts{
				QueueID:      utils.StringPointer("id"),
				ConsumerTag:  utils.StringPointer("tag"),
				Exchange:     utils.StringPointer("exchange"),
				ExchangeType: utils.StringPointer("type"),
				RoutingKey:   utils.StringPointer("key1"),
			},
			SQL: &SQLROpts{
				DBName:              utils.StringPointer("dbname"),
				TableName:           utils.StringPointer("tablename"),
				BatchSize:           utils.IntPointer(0),
				DeleteIndexedFields: utils.SliceStringPointer([]string{"id"}),
				PgSSLMode:           utils.StringPointer("sslmode"),
			},
			AWS: &AWSROpts{

				Region:     utils.StringPointer("eu"),
				Key:        utils.StringPointer("key"),
				Secret:     utils.StringPointer("secret"),
				Token:      utils.StringPointer("token"),
				SQSQueueID: utils.StringPointer("SQSQueue"),
				S3BucketID: utils.StringPointer("S3BucketID"),
			},
			NATS: &NATSROpts{
				JetStream:            utils.BoolPointer(false),
				ConsumerName:         utils.StringPointer("user"),
				StreamName:           utils.StringPointer("stream"),
				QueueID:              utils.StringPointer("id"),
				JWTFile:              utils.StringPointer("jwt"),
				SeedFile:             utils.StringPointer("seed"),
				CertificateAuthority: utils.StringPointer("authority"),
				ClientCertificate:    utils.StringPointer("certificate"),
				ClientKey:            utils.StringPointer("key5"),
				JetStreamMaxWait:     utils.DurationPointer(1 * time.Minute),
			},
			Kafka: &KafkaROpts{
				Topic:   utils.StringPointer("kafka"),
				MaxWait: utils.DurationPointer(1 * time.Minute),
				GroupID: utils.StringPointer("groupId"),
			},
		},
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(rcv, ban) {
		t.Errorf("expected %v received %v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}

}

func TestEventReaderCfgloadFromJSONCfg(t *testing.T) {
	str := "test"
	amqpr := &AMQPROpts{}
	exp := &AMQPROpts{
		Username: &str,
		Password: &str,
	}
	jsnCfg := &EventReaderOptsJson{
		AMQPUsername: &str,
		AMQPPassword: &str,
	}
	err := amqpr.loadFromJSONCfg(jsnCfg)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(amqpr, exp) {
		t.Errorf("expected %v received %v", utils.ToJSON(exp), utils.ToJSON(amqpr))
	}
}

func TestEventReaderCfgClone2(t *testing.T) {
	str := "test"
	amqpr := &AMQPROpts{
		Username: &str,
		Password: &str,
	}
	rcv := amqpr.Clone()

	if !reflect.DeepEqual(amqpr, rcv) {
		t.Errorf("expected %v received %v", utils.ToJSON(amqpr), utils.ToJSON(rcv))
	}
}
func TestEventReaderCfgAsMapInterface(t *testing.T) {
	str := "test"
	amqpr := &AMQPROpts{
		Username: &str,
		Password: &str,
	}
	er := &EventReaderCfg{
		Opts: &EventReaderOpts{
			AMQP: amqpr,
		},
	}
	opts := map[string]any{
		utils.AMQPUsername: str,
		utils.AMQPPassword: str,
	}
	exp := map[string]any{
		utils.IDCfg:                   er.ID,
		utils.TypeCfg:                 er.Type,
		utils.ConcurrentRequestsCfg:   er.ConcurrentReqs,
		utils.SourcePathCfg:           er.SourcePath,
		utils.ProcessedPathCfg:        er.ProcessedPath,
		utils.TenantCfg:               er.Tenant.GetRule(""),
		utils.TimezoneCfg:             er.Timezone,
		utils.FiltersCfg:              er.Filters,
		utils.FlagsCfg:                []string{},
		utils.RunDelayCfg:             "0",
		utils.StartDelayCfg:           "0",
		utils.ReconnectsCfg:           0,
		utils.MaxReconnectIntervalCfg: "0",
		utils.OptsCfg:                 opts,
	}
	rcv := er.AsMapInterface("")

	if !reflect.DeepEqual(exp, rcv) {
		t.Errorf("expected %v received %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestERsCfgappendERsReaders(t *testing.T) {
	tm := 1 * time.Second
	nm := 1
	str := "test"
	bl := true
	tms := "1s"
	fc := []*FCTemplate{{
		Tag:              str,
		Type:             str,
		Path:             str,
		Filters:          []string{str},
		Value:            RSRParsers{},
		Width:            nm,
		Strip:            str,
		Padding:          str,
		Mandatory:        bl,
		AttributeID:      str,
		NewBranch:        bl,
		Timezone:         str,
		Blocker:          bl,
		Layout:           str,
		CostShiftDigits:  nm,
		RoundingDecimals: &nm,
		MaskDestID:       str,
		MaskLen:          nm,
		pathSlice:        []string{},
	}}
	erS := &ERsCfg{
		Enabled:       bl,
		SessionSConns: []string{str},
		Readers: []*EventReaderCfg{
			{
				ID:                  str,
				Type:                str,
				RunDelay:            tm,
				ConcurrentReqs:      nm,
				SourcePath:          str,
				ProcessedPath:       str,
				Opts:                &EventReaderOpts{},
				Tenant:              RSRParsers{},
				Timezone:            str,
				Filters:             []string{str},
				Flags:               utils.FlagsWithParams{},
				Fields:              fc,
				PartialCommitFields: fc,
				CacheDumpFields:     fc,
			},
		},
		PartialCacheTTL: tm,
	}
	fcj := &[]*FcTemplateJsonCfg{}
	jsnReader := &EventReaderJsonCfg{
		Type:                  &str,
		Run_delay:             &tms,
		Concurrent_requests:   &nm,
		Source_path:           &str,
		Processed_path:        &str,
		Opts:                  &EventReaderOptsJson{},
		Tenant:                &str,
		Timezone:              &str,
		Filters:               &[]string{str},
		Flags:                 &[]string{str},
		Fields:                fcj,
		Partial_commit_fields: fcj,
		Cache_dump_fields:     fcj,
	}
	jsnReaders := []*EventReaderJsonCfg{jsnReader}

	exp := &EventReaderCfg{
		Type:                str,
		RunDelay:            tm,
		ConcurrentReqs:      nm,
		SourcePath:          str,
		ProcessedPath:       str,
		Opts:                &EventReaderOpts{},
		Tenant:              RSRParsers{},
		Timezone:            str,
		Filters:             []string{str},
		Flags:               utils.FlagsWithParams{},
		Fields:              fc,
		PartialCommitFields: fc,
		CacheDumpFields:     fc,
	}
	err := exp.loadFromJSONCfg(jsnReader, map[string][]*FCTemplate{}, "")
	if err != nil {
		t.Error(err)
	}

	err = erS.appendERsReaders(&jsnReaders, map[string][]*FCTemplate{}, "", nil)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(erS.Readers[1], exp) {
		t.Errorf("\nexpected %s\nreceived %s\n", utils.ToJSON(exp), utils.ToJSON(erS.Readers[1]))
	}
}

func TestErsCfgloadFromJSONCfg(t *testing.T) {
	tm := 1 * time.Second
	nm := 1
	str := "test"
	str2 := "~test)`"
	bl := true
	tms := "1s"
	fc := []*FCTemplate{{
		Tag:              str,
		Type:             str,
		Path:             str,
		Filters:          []string{str},
		Value:            RSRParsers{},
		Width:            nm,
		Strip:            str,
		Padding:          str,
		Mandatory:        bl,
		AttributeID:      str,
		NewBranch:        bl,
		Timezone:         str,
		Blocker:          bl,
		Layout:           str,
		CostShiftDigits:  nm,
		RoundingDecimals: &nm,
		MaskDestID:       str,
		MaskLen:          nm,
		pathSlice:        []string{str},
	}}
	fc2 := []*FCTemplate{{
		Tag:     str2,
		Type:    utils.MetaTemplate,
		Path:    str2,
		Filters: []string{str2},
		Value: RSRParsers{{
			Rules: str2,
			path:  str2,
		}},
		Width:            nm,
		Strip:            str2,
		Padding:          str2,
		Mandatory:        bl,
		AttributeID:      str2,
		NewBranch:        bl,
		Timezone:         str2,
		Blocker:          bl,
		Layout:           str2,
		CostShiftDigits:  nm,
		RoundingDecimals: &nm,
		MaskDestID:       str2,
		MaskLen:          nm,
		pathSlice:        []string{str2},
	}}
	er := &EventReaderCfg{
		ID:                  str,
		Type:                str,
		RunDelay:            tm,
		ConcurrentReqs:      nm,
		SourcePath:          str,
		ProcessedPath:       str,
		Opts:                &EventReaderOpts{},
		Tenant:              RSRParsers{},
		Timezone:            str,
		Filters:             []string{str},
		Flags:               utils.FlagsWithParams{},
		Fields:              fc,
		PartialCommitFields: fc2,
		CacheDumpFields:     fc,
	}
	fcj := &[]*FcTemplateJsonCfg{}
	fcj2 := &[]*FcTemplateJsonCfg{
		{
			Tag:                  &str2,
			Type:                 &str2,
			Path:                 &str2,
			Attribute_id:         &str2,
			Filters:              &[]string{str2},
			Value:                &str2,
			Width:                &nm,
			Strip:                &str2,
			Padding:              &str2,
			Mandatory:            &bl,
			New_branch:           &bl,
			Timezone:             &str2,
			Blocker:              &bl,
			Layout:               &str2,
			Cost_shift_digits:    &nm,
			Rounding_decimals:    &nm,
			Mask_destinationd_id: &str2,
			Mask_length:          &nm,
		},
	}
	jsnCfg := &EventReaderJsonCfg{
		Type:                  &str,
		Run_delay:             &tms,
		Concurrent_requests:   &nm,
		Source_path:           &str,
		Processed_path:        &str,
		Opts:                  &EventReaderOptsJson{},
		Tenant:                &str,
		Timezone:              &str,
		Filters:               &[]string{str},
		Flags:                 &[]string{str},
		Fields:                fcj,
		Partial_commit_fields: fcj2,
		Cache_dump_fields:     fcj,
	}

	err := er.loadFromJSONCfg(jsnCfg, nil, "")

	if err != nil {
		if err.Error() != "Unclosed unspilit syntax" {
			t.Error(err)
		}
	} else {
		t.Error("was expecting an error")
	}

	jsnCfg2 := &EventReaderJsonCfg{
		Type:                  &str,
		Run_delay:             &tms,
		Concurrent_requests:   &nm,
		Source_path:           &str,
		Processed_path:        &str,
		Opts:                  &EventReaderOptsJson{},
		Tenant:                &str,
		Timezone:              &str,
		Filters:               &[]string{str},
		Flags:                 &[]string{str},
		Fields:                fcj,
		Partial_commit_fields: fcj,
		Cache_dump_fields:     fcj,
	}
	msgTemplates := map[string][]*FCTemplate{
		str2: fc2,
	}
	err = er.loadFromJSONCfg(jsnCfg2, msgTemplates, "")

	if err != nil {
		if err.Error() != "Unclosed unspilit syntax" {
			t.Error(err)
		}
	}
}

func TestERsCfg_ReaderCfg(t *testing.T) {
	// test data
	readerID := "reader1"
	notFoundID := "reader2"
	readers := []*EventReaderCfg{
		{ID: readerID},
		{ID: "readerX"},
	}
	ersCfg := &ERsCfg{Readers: readers}

	// finding an existing reader by ID
	reader := ersCfg.ReaderCfg(readerID)
	if reader == nil {
		t.Errorf("Expected to find reader with ID '%s', but got nil", readerID)
	}

	// not finding a reader with a non-existent ID
	reader = ersCfg.ReaderCfg(notFoundID)
	if reader != nil {
		t.Errorf("Expected not to find reader with ID '%s', but got a reader", notFoundID)
	}
}

func TestKafkaROptsClone(t *testing.T) {

	originalOpts := &KafkaROpts{
		Topic:         utils.StringPointer("topic"),
		GroupID:       utils.StringPointer("group"),
		MaxWait:       utils.DurationPointer(10 * time.Second),
		TLS:           utils.BoolPointer(true),
		CAPath:        utils.StringPointer("/ca/path"),
		SkipTLSVerify: utils.BoolPointer(false),
	}

	clonedOpts := originalOpts.Clone()

	if *clonedOpts.Topic != *originalOpts.Topic {
		t.Errorf("Expected Topic to be copied, got %s vs %s", *clonedOpts.Topic, *originalOpts.Topic)
	}
	if *clonedOpts.GroupID != *originalOpts.GroupID {
		t.Errorf("Expected GroupID to be copied, got %s vs %s", *clonedOpts.GroupID, *originalOpts.GroupID)
	}
	if *clonedOpts.MaxWait != *originalOpts.MaxWait {
		t.Errorf("Expected MaxWait to be copied, got %v vs %v", *clonedOpts.MaxWait, *originalOpts.MaxWait)
	}
	if *clonedOpts.TLS != *originalOpts.TLS {
		t.Errorf("Expected TLS to be copied, got %v vs %v", *clonedOpts.TLS, *originalOpts.TLS)
	}
	if *clonedOpts.CAPath != *originalOpts.CAPath {
		t.Errorf("Expected CAPath to be copied, got %s vs %s", *clonedOpts.CAPath, *originalOpts.CAPath)
	}
	if *clonedOpts.SkipTLSVerify != *originalOpts.SkipTLSVerify {
		t.Errorf("Expected SkipTLSVerify to be copied, got %v vs %v", *clonedOpts.SkipTLSVerify, *originalOpts.SkipTLSVerify)
	}

	*originalOpts.CAPath = "modified/ca/path"
	if *clonedOpts.CAPath == *originalOpts.CAPath {
		t.Errorf("Expected cloned CAPath to be separate, got %s", *clonedOpts.CAPath)
	}
}
