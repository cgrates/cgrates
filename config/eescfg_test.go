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

func TestEESClone(t *testing.T) {
	cfgJSONStr := `{
  "ees": {
     "enabled": true,						
	"attributes_conns":["*internal", "*conn1"],					
	"cache": {
		"*file_csv": {"limit": -2, "ttl": "3s", "static_ttl": true},
	},
	"exporters": [
		{
			"id": "cgrates",									
			"type": "*none",									
			"export_path": "/var/spool/cgrates/ees",			
			"opts": {
				"csvFieldSeparator": ";",					// separator used when reading the fields
				"mysqlDSNParams": {
					"allowOldPasswords": "true",
					"allowNativePasswords": "true",
				},
			"elsIndex":"test",
			"elsIfPrimaryTerm":0,
			"elsIfSeqNo":0,
			"elsOpType":"test2",
			"elsPipeline":"test3",
			"elsRouting":"test4",
			"elsTimeout":"1m",
			"elsVersion":2,
			"elsVersionType":"test5",
			"elsWaitForActiveShards":"test6",
			"sqlMaxIdleConns":4,
			"sqlMaxOpenConns":6,
			"sqlConnMaxLifetime":"1m",
			"sqlTableName":"table",
			"sqlDBName":"db",
			"pgSSLMode":"pg",
			"awsToken":"token",
			"s3FolderPath":"s3",
			"natsJetStream":true,
			"natsSubject":"nat",
			"natsJWTFile":"jwt",
			"natsSeedFile":"seed",
			"natsCertificateAuthority":"NATS",
			"natsClientCertificate":"NATSClient",
			"natsClientKey":"key",
			"natsJetStreamMaxWait":"1m",
			"kafkaTopic":"kafka",
			"amqpQueueID":"id",
			"amqpRoutingKey":"key",
			"amqpExchangeType":"type",
			"amqpExchange":"exchange",
			"awsRegion":"eu",
			"awsKey":"key",
			"awsSecret":"secretkey",
			"sqsQueueID":"sqsid",
			"s3BucketID":"s3",
			"rpcCodec":"rpc",
			"serviceMethod":"service",
			"keyPath":"path",
			"certPath":"certpath",
			"caPath":"capath",
			"tls":true,
			"connIDs":["id1","id2"],
			"rpcConnTimeout":"1m",
			"rpcReplyTimeout":"1m",
			"rpcAPIOpts":{
				"key":"val",
			}
			},											
			"timezone": "local",										
			"filters": ["randomFiletrs"],										
			"flags": [],										
			"attribute_ids": ["randomID"],								
			"attribute_context": "",							
			"synchronous": false,								
			"attempts": 2,										
			"field_separator": ",",								
			"fields":[											
				{"tag": "CGRID", "path": "*exp.CGRID", "type": "*variable", "value": "~*req.CGRID"},
                {"tag": "CGRID", "path": "*hdr.CGRID", "type": "*variable", "value": "~*req.CGRID"},
                {"tag": "CGRID", "path": "*trl.CGRID", "type": "*variable", "value": "~*req.CGRID"},
                {"tag": "CGRID", "path": "*uch.CGRID", "type": "*variable", "value": "~*req.CGRID"},
			],
		},
	],
},
}`

	expected := &EEsCfg{
		Enabled:         true,
		AttributeSConns: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), "*conn1"},
		Cache: map[string]*CacheParamCfg{
			utils.MetaFileCSV: {
				Limit:     -2,
				TTL:       3 * time.Second,
				StaticTTL: true,
				Precache:  false,
				Replicate: false,
			},
		},
		Exporters: []*EventExporterCfg{
			{
				ID:             utils.MetaDefault,
				Type:           utils.MetaNone,
				Synchronous:    false,
				ExportPath:     "/var/spool/cgrates/ees",
				Attempts:       1,
				Timezone:       utils.EmptyString,
				AttributeSCtx:  utils.EmptyString,
				Filters:        []string{},
				AttributeSIDs:  []string{},
				Flags:          utils.FlagsWithParams{},
				Fields:         []*FCTemplate{},
				contentFields:  []*FCTemplate{},
				headerFields:   []*FCTemplate{},
				trailerFields:  []*FCTemplate{},
				Opts:           &EventExporterOpts{},
				FailedPostsDir: "/var/spool/cgrates/failed_posts",
			},
			{
				ID:             utils.CGRateSLwr,
				Type:           utils.MetaNone,
				Synchronous:    false,
				ExportPath:     "/var/spool/cgrates/ees",
				Attempts:       2,
				Timezone:       "local",
				Filters:        []string{"randomFiletrs"},
				AttributeSIDs:  []string{"randomID"},
				Flags:          utils.FlagsWithParams{},
				FailedPostsDir: "/var/spool/cgrates/failed_posts",
				Fields: []*FCTemplate{
					{
						Tag:    utils.CGRID,
						Path:   "*exp.CGRID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.CGRID", utils.InfieldSep),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.CGRID,
						Path:   "*hdr.CGRID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.CGRID", utils.InfieldSep),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.CGRID,
						Path:   "*trl.CGRID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.CGRID", utils.InfieldSep),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.CGRID,
						Path:   "*uch.CGRID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.CGRID", utils.InfieldSep),
						Layout: time.RFC3339,
					},
				},
				contentFields: []*FCTemplate{
					{
						Tag:    utils.CGRID,
						Path:   "*exp.CGRID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.CGRID", utils.InfieldSep),
						Layout: time.RFC3339,
					},
					{
						Tag:    utils.CGRID,
						Path:   "*uch.CGRID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.CGRID", utils.InfieldSep),
						Layout: time.RFC3339,
					},
				},
				headerFields: []*FCTemplate{
					{
						Tag:    utils.CGRID,
						Path:   "*hdr.CGRID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.CGRID", utils.InfieldSep),
						Layout: time.RFC3339,
					},
				},
				trailerFields: []*FCTemplate{
					{
						Tag:    utils.CGRID,
						Path:   "*trl.CGRID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.CGRID", utils.InfieldSep),
						Layout: time.RFC3339,
					},
				},
				Opts: &EventExporterOpts{

					CSVFieldSeparator: utils.StringPointer(utils.InfieldSep),
					MYSQLDSNParams: map[string]string{
						"allowOldPasswords":    "true",
						"allowNativePasswords": "true",
					},
					ElsIndex:                 utils.StringPointer("test"),
					ElsIfPrimaryTerm:         utils.IntPointer(0),
					ElsIfSeqNo:               utils.IntPointer(0),
					ElsOpType:                utils.StringPointer("test2"),
					ElsPipeline:              utils.StringPointer("test3"),
					ElsRouting:               utils.StringPointer("test4"),
					ElsTimeout:               utils.DurationPointer(1 * time.Minute),
					ElsVersion:               utils.IntPointer(2),
					ElsVersionType:           utils.StringPointer("test5"),
					ElsWaitForActiveShards:   utils.StringPointer("test6"),
					SQLMaxIdleConns:          utils.IntPointer(4),
					SQLConnMaxLifetime:       utils.DurationPointer(1 * time.Minute),
					SQLTableName:             utils.StringPointer("table"),
					SQLDBName:                utils.StringPointer("db"),
					PgSSLMode:                utils.StringPointer("pg"),
					KafkaTopic:               utils.StringPointer("kafka"),
					SQLMaxOpenConns:          utils.IntPointer(6),
					AWSToken:                 utils.StringPointer("token"),
					S3FolderPath:             utils.StringPointer("s3"),
					NATSJetStream:            utils.BoolPointer(true),
					NATSSubject:              utils.StringPointer("nat"),
					NATSJWTFile:              utils.StringPointer("jwt"),
					NATSSeedFile:             utils.StringPointer("seed"),
					NATSCertificateAuthority: utils.StringPointer("NATS"),
					NATSClientCertificate:    utils.StringPointer("NATSClient"),
					NATSClientKey:            utils.StringPointer("key"),
					NATSJetStreamMaxWait:     utils.DurationPointer(1 * time.Minute),
					AMQPRoutingKey:           utils.StringPointer("key"),
					AMQPQueueID:              utils.StringPointer("id"),
					AMQPExchangeType:         utils.StringPointer("type"),
					AMQPExchange:             utils.StringPointer("exchange"),
					AWSRegion:                utils.StringPointer("eu"),
					AWSKey:                   utils.StringPointer("key"),
					AWSSecret:                utils.StringPointer("secretkey"),
					S3BucketID:               utils.StringPointer("s3"),
					SQSQueueID:               utils.StringPointer("sqsid"),
					RPCCodec:                 utils.StringPointer("rpc"),
					ServiceMethod:            utils.StringPointer("service"),
					KeyPath:                  utils.StringPointer("path"),
					CertPath:                 utils.StringPointer("certpath"),
					CAPath:                   utils.StringPointer("capath"),
					TLS:                      utils.BoolPointer(true),
					ConnIDs:                  utils.SliceStringPointer([]string{"id1", "id2"}),
					RPCConnTimeout:           utils.DurationPointer(1 * time.Minute),
					RPCReplyTimeout:          utils.DurationPointer(1 * time.Minute),
					RPCAPIOpts: map[string]interface{}{
						"key": "val",
					},
				},
			},
		},
	}
	for _, profile := range expected.Exporters {
		for _, v := range profile.Fields {
			v.ComputePath()
		}
		for _, v := range profile.ContentFields() {
			v.ComputePath()
		}
		for _, v := range profile.HeaderFields() {
			v.ComputePath()
		}
		for _, v := range profile.TrailerFields() {
			v.ComputePath()
		}
	}
	if jsonCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else {
		cloneCfg := jsonCfg.eesCfg.Clone()
		if !reflect.DeepEqual(cloneCfg, expected) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(cloneCfg))
		}
	}
}

func TestEventExporterFieldloadFromJsonCfg(t *testing.T) {
	eventExporterJSON := &EEsJsonCfg{
		Exporters: &[]*EventExporterJsonCfg{
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
	if err = jsonCfg.eesCfg.loadFromJSONCfg(eventExporterJSON, jsonCfg.templates, jsonCfg.generalCfg.RSRSep, jsonCfg.dfltEvExp); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestEventExporterFieldloadFromJsonCfg1(t *testing.T) {
	eventExporterJSON := &EEsJsonCfg{
		Exporters: &[]*EventExporterJsonCfg{
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
	if err = jsonCfg.eesCfg.loadFromJSONCfg(eventExporterJSON, jsonCfg.templates, jsonCfg.generalCfg.RSRSep, jsonCfg.dfltEvExp); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestEventExporterloadFromJsonCfg(t *testing.T) {
	jsonCfg := NewDefaultCGRConfig()

	eventExporter := new(EventExporterCfg)
	if err := eventExporter.loadFromJSONCfg(nil, jsonCfg.templates, jsonCfg.generalCfg.RSRSep); err != nil {
		t.Error(err)
	}
}

func TestEventExporterOptsloadFromJsonCfg(t *testing.T) {

	eventExporterOptsJSON := &EventExporterOptsJson{

		ElsIndex:                 utils.StringPointer("test"),
		ElsIfPrimaryTerm:         utils.IntPointer(0),
		ElsIfSeqNo:               utils.IntPointer(0),
		ElsOpType:                utils.StringPointer("test2"),
		ElsPipeline:              utils.StringPointer("test3"),
		ElsRouting:               utils.StringPointer("test4"),
		ElsTimeout:               utils.StringPointer("1m"),
		ElsVersion:               utils.IntPointer(2),
		ElsVersionType:           utils.StringPointer("test5"),
		ElsWaitForActiveShards:   utils.StringPointer("test6"),
		SQLMaxIdleConns:          utils.IntPointer(4),
		SQLMaxOpenConns:          utils.IntPointer(6),
		SQLConnMaxLifetime:       utils.StringPointer("1m"),
		SQLTableName:             utils.StringPointer("table"),
		SQLDBName:                utils.StringPointer("db"),
		PgSSLMode:                utils.StringPointer("pg"),
		AWSToken:                 utils.StringPointer("token"),
		S3FolderPath:             utils.StringPointer("s3"),
		NATSJetStream:            utils.BoolPointer(true),
		NATSSubject:              utils.StringPointer("nat"),
		NATSJWTFile:              utils.StringPointer("jwt"),
		NATSSeedFile:             utils.StringPointer("seed"),
		NATSCertificateAuthority: utils.StringPointer("NATS"),
		NATSClientCertificate:    utils.StringPointer("NATSClient"),
		NATSClientKey:            utils.StringPointer("key"),
		NATSJetStreamMaxWait:     utils.StringPointer("1m"),
	}

	expected := &EventExporterOpts{

		ElsIndex:                 utils.StringPointer("test"),
		ElsIfPrimaryTerm:         utils.IntPointer(0),
		ElsIfSeqNo:               utils.IntPointer(0),
		ElsOpType:                utils.StringPointer("test2"),
		ElsPipeline:              utils.StringPointer("test3"),
		ElsRouting:               utils.StringPointer("test4"),
		ElsTimeout:               utils.DurationPointer(1 * time.Minute),
		ElsVersion:               utils.IntPointer(2),
		ElsVersionType:           utils.StringPointer("test5"),
		ElsWaitForActiveShards:   utils.StringPointer("test6"),
		SQLMaxIdleConns:          utils.IntPointer(4),
		SQLMaxOpenConns:          utils.IntPointer(6),
		SQLConnMaxLifetime:       utils.DurationPointer(1 * time.Minute),
		SQLTableName:             utils.StringPointer("table"),
		SQLDBName:                utils.StringPointer("db"),
		PgSSLMode:                utils.StringPointer("pg"),
		AWSToken:                 utils.StringPointer("token"),
		S3FolderPath:             utils.StringPointer("s3"),
		NATSJetStream:            utils.BoolPointer(true),
		NATSSubject:              utils.StringPointer("nat"),
		NATSJWTFile:              utils.StringPointer("jwt"),
		NATSSeedFile:             utils.StringPointer("seed"),
		NATSCertificateAuthority: utils.StringPointer("NATS"),
		NATSClientCertificate:    utils.StringPointer("NATSClient"),
		NATSClientKey:            utils.StringPointer("key"),
		NATSJetStreamMaxWait:     utils.DurationPointer(1 * time.Minute),
	}
	eventExporter := &EventExporterCfg{
		Opts: &EventExporterOpts{},
	}
	if err := eventExporter.Opts.loadFromJSONCfg(eventExporterOptsJSON); err != nil {
		t.Error(expected)
	} else if !reflect.DeepEqual(expected, eventExporter.Opts) {
		t.Errorf("expected %v  received %v", expected, eventExporter.Opts)
	}
	if err := eventExporter.Opts.loadFromJSONCfg(nil); err != nil {
		t.Error(err)
	} else if err := eventExporter.Opts.loadFromJSONCfg(&EventExporterOptsJson{
		ElsTimeout: utils.StringPointer("test"),
	}); err == nil {
		t.Error(err)
	} else if err := eventExporter.Opts.loadFromJSONCfg(&EventExporterOptsJson{
		SQLConnMaxLifetime: utils.StringPointer("test"),
	}); err == nil {
		t.Error(err)
	} else if err := eventExporter.Opts.loadFromJSONCfg(&EventExporterOptsJson{
		RPCConnTimeout: utils.StringPointer("test"),
	}); err == nil {
		t.Error(err)
	} else if err := eventExporter.Opts.loadFromJSONCfg(&EventExporterOptsJson{
		NATSJetStreamMaxWait: utils.StringPointer("test"),
	}); err == nil {
		t.Error(err)
	} else if err := eventExporter.Opts.loadFromJSONCfg(&EventExporterOptsJson{

		RPCReplyTimeout: utils.StringPointer("test"),
	}); err == nil {
		t.Error(err)
	}

}

func TestEESCacheloadFromJsonCfg(t *testing.T) {
	eesCfg := &EEsJsonCfg{
		Cache: &map[string]*CacheParamJsonCfg{
			utils.MetaFileCSV: {
				Ttl: utils.StringPointer("1ss"),
			},
		},
	}
	expected := "time: unknown unit \"ss\" in duration \"1ss\""
	jsonCfg := NewDefaultCGRConfig()
	if err = jsonCfg.eesCfg.loadFromJSONCfg(eesCfg, jsonCfg.templates, jsonCfg.generalCfg.RSRSep, jsonCfg.dfltEvExp); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestEventExporterSameID(t *testing.T) {
	expectedEEsCfg := &EEsCfg{
		Enabled:         true,
		AttributeSConns: []string{"conn1"},
		Cache: map[string]*CacheParamCfg{
			utils.MetaFileCSV: {
				Limit:     -1,
				TTL:       5 * time.Second,
				StaticTTL: false,
			},
		},
		Exporters: []*EventExporterCfg{
			{
				ID:             utils.MetaDefault,
				Type:           utils.MetaNone,
				ExportPath:     "/var/spool/cgrates/ees",
				Attempts:       1,
				Timezone:       utils.EmptyString,
				Filters:        []string{},
				AttributeSIDs:  []string{},
				Flags:          utils.FlagsWithParams{},
				Fields:         []*FCTemplate{},
				contentFields:  []*FCTemplate{},
				headerFields:   []*FCTemplate{},
				trailerFields:  []*FCTemplate{},
				Opts:           &EventExporterOpts{},
				FailedPostsDir: "/var/spool/cgrates/failed_posts",
			},
			{
				ID:            "file_exporter1",
				Type:          utils.MetaFileCSV,
				Timezone:      utils.EmptyString,
				Filters:       []string{},
				AttributeSIDs: []string{},
				ExportPath:    "/var/spool/cgrates/ees",
				Attempts:      1,
				Flags:         utils.FlagsWithParams{},
				Fields: []*FCTemplate{
					{Tag: "CustomTag2", Path: "*exp.CustomPath2", Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("CustomValue2", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
				},
				contentFields: []*FCTemplate{
					{Tag: "CustomTag2", Path: "*exp.CustomPath2", Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("CustomValue2", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
				},
				headerFields:   []*FCTemplate{},
				trailerFields:  []*FCTemplate{},
				Opts:           &EventExporterOpts{},
				FailedPostsDir: "/var/spool/cgrates/failed_posts",
			},
		},
	}
	for _, profile := range expectedEEsCfg.Exporters {
		for _, v := range profile.Fields {
			v.ComputePath()
		}
		for _, v := range profile.contentFields {
			v.ComputePath()
		}
	}
	cfgJSONStr := `{
"ees": {
	"enabled": true,
	"attributes_conns":["conn1"],
	"exporters": [
		{
			"id": "file_exporter1",
			"type": "*file_csv",
			"fields":[
				{"tag": "CustomTag1", "path": "*exp.CustomPath1", "type": "*variable", "value": "CustomValue1", "mandatory": true},
			],
		},
		{
			"id": "file_exporter1",
			"type": "*file_csv",
			"fields":[
				{"tag": "CustomTag2", "path": "*exp.CustomPath2", "type": "*variable", "value": "CustomValue2", "mandatory": true},
			],
		},
	],
}
}`
	if cfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedEEsCfg, cfg.eesCfg) {
		t.Errorf("Expected: %+v ,\n received: %+v", utils.ToJSON(expectedEEsCfg), utils.ToJSON(cfg.eesCfg))
	}
}

func TestEEsCfgloadFromJsonCfgCase1(t *testing.T) {
	jsonCfg := &EEsJsonCfg{
		Enabled:          utils.BoolPointer(true),
		Attributes_conns: &[]string{"*conn1", "*conn2"},
		Cache: &map[string]*CacheParamJsonCfg{
			utils.MetaFileCSV: {
				Limit:      utils.IntPointer(-2),
				Ttl:        utils.StringPointer("1s"),
				Static_ttl: utils.BoolPointer(false),
			},
		},
		Exporters: &[]*EventExporterJsonCfg{
			{
				Id:               utils.StringPointer("CSVExporter"),
				Type:             utils.StringPointer("*file_csv"),
				Filters:          &[]string{},
				Attribute_ids:    &[]string{},
				Flags:            &[]string{"*dryrun"},
				Export_path:      utils.StringPointer("/tmp/testCSV"),
				Timezone:         utils.StringPointer("UTC"),
				Synchronous:      utils.BoolPointer(true),
				Attempts:         utils.IntPointer(1),
				Failed_posts_dir: utils.StringPointer("/var/spool/cgrates/failed_posts"),
				Fields: &[]*FcTemplateJsonCfg{
					{
						Tag:   utils.StringPointer(utils.CGRID),
						Path:  utils.StringPointer("*exp.CGRID"),
						Type:  utils.StringPointer(utils.MetaVariable),
						Value: utils.StringPointer("~*req.CGRID"),
					},
				},
			},
		},
	}
	expectedCfg := &EEsCfg{
		Enabled:         true,
		AttributeSConns: []string{"*conn1", "*conn2"},
		Cache: map[string]*CacheParamCfg{
			utils.MetaFileCSV: {
				Limit:     -2,
				TTL:       time.Second,
				StaticTTL: false,
			},
		},
		Exporters: []*EventExporterCfg{
			{
				ID:             utils.MetaDefault,
				Type:           utils.MetaNone,
				ExportPath:     "/var/spool/cgrates/ees",
				Attempts:       1,
				Timezone:       utils.EmptyString,
				Filters:        []string{},
				AttributeSIDs:  []string{},
				Flags:          utils.FlagsWithParams{},
				contentFields:  []*FCTemplate{},
				Fields:         []*FCTemplate{},
				headerFields:   []*FCTemplate{},
				trailerFields:  []*FCTemplate{},
				Opts:           &EventExporterOpts{},
				FailedPostsDir: "/var/spool/cgrates/failed_posts",
			},
			{
				ID:            "CSVExporter",
				Type:          "*file_csv",
				Filters:       []string{},
				AttributeSIDs: []string{},
				Flags:         utils.FlagsWithParamsFromSlice([]string{utils.MetaDryRun}),
				ExportPath:    "/tmp/testCSV",
				Timezone:      "UTC",
				Synchronous:   true,
				Attempts:      1,
				headerFields:  []*FCTemplate{},
				trailerFields: []*FCTemplate{},
				contentFields: []*FCTemplate{
					{
						Tag:    utils.CGRID,
						Path:   "*exp.CGRID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.CGRID", utils.InfieldSep),
						Layout: time.RFC3339,
					},
				},
				Opts: &EventExporterOpts{},
				Fields: []*FCTemplate{
					{Tag: utils.CGRID, Path: "*exp.CGRID", Type: utils.MetaVariable, Value: NewRSRParsersMustCompile("~*req.CGRID", utils.InfieldSep), Layout: time.RFC3339},
				},
				FailedPostsDir: "/var/spool/cgrates/failed_posts",
			},
		},
	}
	for _, profile := range expectedCfg.Exporters {
		for _, v := range profile.Fields {
			v.ComputePath()
		}
		for _, v := range profile.contentFields {
			v.ComputePath()
		}
	}
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.eesCfg.loadFromJSONCfg(nil, cgrCfg.templates, cgrCfg.generalCfg.RSRSep, cgrCfg.dfltEvExp); err != nil {
		t.Error(err)
	} else if err := cgrCfg.eesCfg.loadFromJSONCfg(jsonCfg, cgrCfg.templates, cgrCfg.generalCfg.RSRSep, cgrCfg.dfltEvExp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedCfg, cgrCfg.eesCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedCfg), utils.ToJSON(cgrCfg.eesCfg))
	}
}

func TestEEsCfgloadFromJsonCfgCase2(t *testing.T) {
	jsonCfg := &EEsJsonCfg{
		Enabled:          utils.BoolPointer(true),
		Attributes_conns: &[]string{"*conn1", "*conn2"},
		Cache: &map[string]*CacheParamJsonCfg{
			utils.MetaFileCSV: {
				Limit:      utils.IntPointer(-2),
				Ttl:        utils.StringPointer("1s"),
				Static_ttl: utils.BoolPointer(false),
			},
		},
		Exporters: &[]*EventExporterJsonCfg{
			{
				Id:            utils.StringPointer("CSVExporter"),
				Type:          utils.StringPointer("*file_csv"),
				Filters:       &[]string{},
				Attribute_ids: &[]string{},
				Flags:         &[]string{"*dryrun"},
				Export_path:   utils.StringPointer("/tmp/testCSV"),
				Timezone:      utils.StringPointer("UTC"),
				Synchronous:   utils.BoolPointer(true),
				Attempts:      utils.IntPointer(1),
				Fields: &[]*FcTemplateJsonCfg{
					{
						Tag:    utils.StringPointer(utils.AnswerTime),
						Path:   utils.StringPointer("*exp.AnswerTime"),
						Type:   utils.StringPointer(utils.MetaTemplate),
						Value:  utils.StringPointer("randomVal"),
						Layout: utils.StringPointer(time.RFC3339),
					},
					{
						Path:  utils.StringPointer("*req.CGRID"),
						Type:  utils.StringPointer(utils.MetaVariable),
						Value: utils.StringPointer("1"),
					},
				},
			},
		},
	}
	expectedCfg := &EEsCfg{
		Enabled:         true,
		AttributeSConns: []string{"*conn1", "*conn2"},
		Cache: map[string]*CacheParamCfg{
			utils.MetaFileCSV: {
				Limit:     -2,
				TTL:       time.Second,
				StaticTTL: false,
			},
		},
		Exporters: []*EventExporterCfg{
			{
				ID:             utils.MetaDefault,
				Type:           utils.MetaNone,
				ExportPath:     "/var/spool/cgrates/ees",
				Attempts:       1,
				Timezone:       utils.EmptyString,
				Filters:        []string{},
				AttributeSIDs:  []string{},
				Flags:          utils.FlagsWithParams{},
				contentFields:  []*FCTemplate{},
				Fields:         []*FCTemplate{},
				headerFields:   []*FCTemplate{},
				trailerFields:  []*FCTemplate{},
				Opts:           &EventExporterOpts{},
				FailedPostsDir: "/var/spool/cgrates/failed_posts",
			},
			{
				ID:            "CSVExporter",
				Type:          "*file_csv",
				Filters:       []string{},
				AttributeSIDs: []string{},
				Flags:         utils.FlagsWithParamsFromSlice([]string{utils.MetaDryRun}),
				ExportPath:    "/tmp/testCSV",
				Timezone:      "UTC",
				Synchronous:   true,
				Attempts:      1,
				headerFields:  []*FCTemplate{},
				trailerFields: []*FCTemplate{},
				contentFields: []*FCTemplate{
					{
						Tag:    utils.CGRID,
						Path:   "*exp.CGRID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.CGRID", utils.InfieldSep),
						Layout: time.RFC3339,
					},
				},
				FailedPostsDir: "/var/spool/cgrates/failed_posts",
				Opts:           &EventExporterOpts{},
				Fields: []*FCTemplate{
					{
						Tag:    utils.CGRID,
						Path:   "*exp.CGRID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("~*req.CGRID", utils.InfieldSep),
						Layout: time.RFC3339,
					},
					{
						Tag:    "*req.CGRID",
						Path:   "*req.CGRID",
						Type:   utils.MetaVariable,
						Value:  NewRSRParsersMustCompile("1", utils.InfieldSep),
						Layout: time.RFC3339,
					},
				},
			},
		},
	}
	msgTemplates := map[string][]*FCTemplate{
		"randomVal": {
			{
				Tag:    utils.CGRID,
				Path:   "*exp.CGRID",
				Type:   utils.MetaVariable,
				Value:  NewRSRParsersMustCompile("~*req.CGRID", utils.InfieldSep),
				Layout: time.RFC3339,
			},
		},
	}
	for _, profile := range expectedCfg.Exporters {
		for _, v := range profile.Fields {
			v.ComputePath()
		}
		for _, v := range profile.contentFields {
			v.ComputePath()
		}
		for _, v := range msgTemplates["randomVal"] {
			v.ComputePath()
		}
	}
	jsnCfg := NewDefaultCGRConfig()
	if err = jsnCfg.eesCfg.loadFromJSONCfg(jsonCfg, msgTemplates, jsnCfg.generalCfg.RSRSep, jsnCfg.dfltEvExp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedCfg, jsnCfg.eesCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedCfg), utils.ToJSON(jsnCfg.eesCfg))
	}
}

func TestEEsCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
      "ees": {									
	        "enabled": true,						
            "attributes_conns":["*internal","*conn2"],					
            "cache": {
		          "*file_csv": {"limit": -2, "precache": false, "replicate": false, "ttl": "1s", "static_ttl": false}
            },
            "exporters": [
            {
                  "id": "CSVExporter",									
			      "type": "*file_csv",									
                  "export_path": "/tmp/testCSV",			
			      "opts": {
					"elsIndex":"test",
					"elsIfPrimaryTerm":0,
					"kafkaTopic": "test",	
					"elsIfSeqNo":0,
					"elsOpType":"test2",
					"elsPipeline":"test3",
					"elsRouting":"test4",
					"elsTimeout":"1m",
					"elsVersion":2,
					"elsVersionType":"test5",
					"elsWaitForActiveShards":"test6",
					"sqlMaxIdleConns":4,
					"sqlMaxOpenConns":6,
					"sqlConnMaxLifetime":"1m",
					"sqlTableName":"table",
					"sqlDBName":"db",
					"pgSSLMode":"pg",
					"awsToken":"token",
					"s3FolderPath":"s3",
					"natsJetStream":true,
					"natsSubject":"nat",
					"natsJWTFile":"jwt",
					"natsSeedFile":"seed",
					"natsCertificateAuthority":"NATS",
					"natsClientCertificate":"NATSClient",
					"natsClientKey":"key",
					"natsJetStreamMaxWait":"1m",				
					"amqpQueueID":"id",
					"amqpRoutingKey":"key",
					"amqpExchangeType":"type",
					"amqpExchange":"exchange",
					"awsRegion":"eu",
					"awsKey":"key",
					"awsSecret":"secretkey",
					"sqsQueueID":"sqsid",
					"s3BucketID":"s3",
					"rpcCodec":"rpc",
					"serviceMethod":"service",
					"keyPath":"path",
					"certPath":"certpath",
					"caPath":"capath",
					"tls":true,
					"connIDs":["id1","id2"],
					"rpcConnTimeout":"1m",
					"rpcReplyTimeout":"1m",
					"csvFieldSeparator":",",
					"mysqlDSNParams":{
						"key":"param",
					},	
					"rpcAPIOpts":{
						"key":"val",
					}				
				  },											
			      "timezone": "UTC",										
			      "filters": [],										
			      "flags": ["randomFlag"],										
			      "attribute_ids": [],								
			      "attribute_context": "",							
			      "synchronous": false,								
			      "attempts": 1,										
			      "field_separator": ",",								
			      "fields":[
                      {"tag": "CGRID", "path": "*exp.CGRID", "type": "*variable", "value": "~*req.CGRID"}
                  ]
            }]
	  }
    }`
	eMap := map[string]interface{}{
		utils.EnabledCfg:         true,
		utils.AttributeSConnsCfg: []string{utils.MetaInternal, "*conn2"},
		utils.CacheCfg: map[string]interface{}{
			utils.MetaFileCSV: map[string]interface{}{
				utils.LimitCfg:     -2,
				utils.PrecacheCfg:  false,
				utils.ReplicateCfg: false,
				utils.RemoteCfg:    false,
				utils.TTLCfg:       "1s",
				utils.StaticTTLCfg: false,
			},
		},
		utils.ExportersCfg: []map[string]interface{}{
			{
				utils.IDCfg:         "CSVExporter",
				utils.TypeCfg:       "*file_csv",
				utils.ExportPathCfg: "/tmp/testCSV",
				utils.OptsCfg: map[string]interface{}{
					utils.KafkaTopic:               "test",
					utils.ElsIndex:                 "test",
					utils.ElsIfPrimaryTerm:         0,
					utils.ElsIfSeqNo:               0,
					utils.ElsOpType:                "test2",
					utils.ElsPipeline:              "test3",
					utils.ElsRouting:               "test4",
					utils.ElsTimeout:               "1m0s",
					utils.ElsVersionLow:            2,
					utils.ElsVersionType:           "test5",
					utils.ElsWaitForActiveShards:   "test6",
					utils.SQLMaxIdleConnsCfg:       4,
					utils.SQLMaxOpenConns:          6,
					utils.SQLConnMaxLifetime:       "1m0s",
					utils.SQLTableNameOpt:          "table",
					utils.SQLDBNameOpt:             "db",
					utils.PgSSLModeCfg:             "pg",
					utils.AWSToken:                 "token",
					utils.S3FolderPath:             "s3",
					utils.NatsJetStream:            true,
					utils.NatsSubject:              "nat",
					utils.NatsJWTFile:              "jwt",
					utils.NatsSeedFile:             "seed",
					utils.NatsCertificateAuthority: "NATS",
					utils.NatsClientCertificate:    "NATSClient",
					utils.NatsClientKey:            "key",
					utils.NatsJetStreamMaxWait:     "1m0s",
					utils.AMQPQueueID:              "id",
					utils.AMQPRoutingKey:           "key",
					utils.AMQPExchangeType:         "type",
					utils.AMQPExchange:             "exchange",
					utils.AWSRegion:                "eu",
					utils.AWSKey:                   "key",
					utils.AWSSecret:                "secretkey",
					utils.SQSQueueID:               "sqsid",
					utils.S3Bucket:                 "s3",
					utils.RpcCodec:                 "rpc",
					utils.ServiceMethod:            "service",
					utils.KeyPath:                  "path",
					utils.CertPath:                 "certpath",
					utils.CaPath:                   "capath",
					utils.Tls:                      true,
					utils.ConnIDs:                  []string{"id1", "id2"},
					utils.RpcConnTimeout:           "1m0s",
					utils.RpcReplyTimeout:          "1m0s",
					utils.CSVFieldSepOpt:           ",",
					utils.MYSQLDSNParams: map[string]string{
						"key": "param",
					},
					utils.RPCAPIOpts: map[string]interface{}{
						"key": "val",
					},
				},
				utils.TimezoneCfg:           "UTC",
				utils.FiltersCfg:            []string{},
				utils.FlagsCfg:              []string{"randomFlag"},
				utils.AttributeIDsCfg:       []string{},
				utils.AttributeContextCfg:   utils.EmptyString,
				utils.SynchronousCfg:        false,
				utils.AttemptsCfg:           1,
				utils.ConcurrentRequestsCfg: 0,
				utils.FieldsCfg: []map[string]interface{}{
					{
						utils.TagCfg:   utils.CGRID,
						utils.PathCfg:  "*exp.CGRID",
						utils.TypeCfg:  utils.MetaVariable,
						utils.ValueCfg: "~*req.CGRID",
					},
				},
				utils.FailedPostsDirCfg: "/var/spool/cgrates/failed_posts",
			},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else {
		rcv := cgrCfg.eesCfg.AsMapInterface(cgrCfg.generalCfg.RSRSep)
		if len(rcv[utils.ExportersCfg].([]map[string]interface{})) != 2 {
			t.Errorf("Expected %+v, received %+v", 2, len(rcv[utils.ExportersCfg].([]map[string]interface{})))
		} else if !reflect.DeepEqual(eMap[utils.ExportersCfg].([]map[string]interface{})[0][utils.FieldsCfg].([]map[string]interface{})[0][utils.ValueCfg],
			rcv[utils.ExportersCfg].([]map[string]interface{})[1][utils.FieldsCfg].([]map[string]interface{})[0][utils.ValueCfg]) {
			t.Errorf("Expected %+v, received %+v", eMap[utils.ExportersCfg].([]map[string]interface{})[0][utils.FieldsCfg].([]map[string]interface{})[0][utils.ValueCfg],
				rcv[utils.ExportersCfg].([]map[string]interface{})[1][utils.FieldsCfg].([]map[string]interface{})[0][utils.ValueCfg])
		}
		rcv[utils.ExportersCfg].([]map[string]interface{})[1][utils.FieldsCfg] = nil
		eMap[utils.ExportersCfg].([]map[string]interface{})[0][utils.FieldsCfg] = nil
		if !reflect.DeepEqual(rcv[utils.ExportersCfg].([]map[string]interface{})[1],
			eMap[utils.ExportersCfg].([]map[string]interface{})[0]) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap[utils.ExportersCfg].([]map[string]interface{})[0]),
				utils.ToJSON(rcv[utils.ExportersCfg].([]map[string]interface{})[1]))
		}
		rcv[utils.ExportersCfg] = nil
		eMap[utils.ExportersCfg] = nil
		if !reflect.DeepEqual(rcv, eMap) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
		}
	}
}
