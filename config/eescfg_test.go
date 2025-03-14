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
			"elsRefresh":"true",
			"elsOpType":"test2",
			"elsPipeline":"test3",
			"elsRouting":"test4",
			"elsTimeout":"1m",
			"elsWaitForActiveShards":"test6",
			"sqlMaxIdleConns":4,
			"sqlMaxOpenConns":6,
			"sqlConnMaxLifetime":"1m",
			"sqlTableName":"table",
			"sqlDBName":"db",
			"sqlUpdateIndexedFields": ["id"],
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
			utils.MetaAMQPV1jsonMap: {
				Limit: -1,

				StaticTTL: false,
				Precache:  false,
				Replicate: false,
			},
			utils.MetaAMQPjsonMap: {
				Limit: -1,

				StaticTTL: false,
				Precache:  false,
				Replicate: false,
			},
			utils.MetaElastic: {
				Limit: -1,

				StaticTTL: false,
				Precache:  false,
				Replicate: false,
			},
			utils.MetaS3jsonMap: {
				Limit: -1,

				StaticTTL: false,
				Precache:  false,
				Replicate: false,
			},
			utils.MetaSQL: {
				Limit: -1,

				StaticTTL: false,
				Precache:  false,
				Replicate: false,
			},
			utils.MetaSQSjsonMap: {
				Limit: -1,

				StaticTTL: false,
				Precache:  false,
				Replicate: false,
			},
			utils.MetaNatsjsonMap: {
				Limit: -1,

				StaticTTL: false,
				Precache:  false,
				Replicate: false,
			},
			utils.MetaKafkajsonMap: {
				Limit: -1,

				StaticTTL: false,
				Precache:  false,
				Replicate: false,
			},
		},
		Exporters: []*EventExporterCfg{
			{
				ID:            utils.MetaDefault,
				Type:          utils.MetaNone,
				Synchronous:   false,
				ExportPath:    "/var/spool/cgrates/ees",
				Attempts:      1,
				Timezone:      utils.EmptyString,
				AttributeSCtx: utils.EmptyString,
				Filters:       []string{},
				AttributeSIDs: []string{},
				Flags:         utils.FlagsWithParams{},
				Fields:        []*FCTemplate{},
				contentFields: []*FCTemplate{},
				headerFields:  []*FCTemplate{},
				trailerFields: []*FCTemplate{},
				Opts: &EventExporterOpts{
					AMQP:  &AMQPOpts{},
					AWS:   &AWSOpts{},
					RPC:   &RPCOpts{},
					Els:   &ElsOpts{},
					Kafka: &KafkaOpts{},
					NATS:  &NATSOpts{},
					SQL:   &SQLOpts{},
				},
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
					SQL: &SQLOpts{
						MYSQLDSNParams: map[string]string{
							"allowOldPasswords":    "true",
							"allowNativePasswords": "true",
						},
						MaxIdleConns:        utils.IntPointer(4),
						ConnMaxLifetime:     utils.DurationPointer(1 * time.Minute),
						TableName:           utils.StringPointer("table"),
						DBName:              utils.StringPointer("db"),
						UpdateIndexedFields: utils.SliceStringPointer([]string{"id"}),
						PgSSLMode:           utils.StringPointer("pg"),
						MaxOpenConns:        utils.IntPointer(6),
					},
					Els: &ElsOpts{
						Index:               utils.StringPointer("test"),
						Refresh:             utils.StringPointer("true"),
						OpType:              utils.StringPointer("test2"),
						Pipeline:            utils.StringPointer("test3"),
						Routing:             utils.StringPointer("test4"),
						Timeout:             utils.DurationPointer(1 * time.Minute),
						WaitForActiveShards: utils.StringPointer("test6"),
					},
					Kafka: &KafkaOpts{
						Topic: utils.StringPointer("kafka"),
					},
					AWS: &AWSOpts{
						Token:        utils.StringPointer("token"),
						S3FolderPath: utils.StringPointer("s3"),
						Region:       utils.StringPointer("eu"),
						Key:          utils.StringPointer("key"),
						Secret:       utils.StringPointer("secretkey"),
						S3BucketID:   utils.StringPointer("s3"),
						SQSQueueID:   utils.StringPointer("sqsid"),
					},
					NATS: &NATSOpts{
						JetStream:            utils.BoolPointer(true),
						Subject:              utils.StringPointer("nat"),
						JWTFile:              utils.StringPointer("jwt"),
						SeedFile:             utils.StringPointer("seed"),
						CertificateAuthority: utils.StringPointer("NATS"),
						ClientCertificate:    utils.StringPointer("NATSClient"),
						ClientKey:            utils.StringPointer("key"),
						JetStreamMaxWait:     utils.DurationPointer(1 * time.Minute),
					},
					AMQP: &AMQPOpts{
						RoutingKey:   utils.StringPointer("key"),
						QueueID:      utils.StringPointer("id"),
						ExchangeType: utils.StringPointer("type"),
						Exchange:     utils.StringPointer("exchange"),
					},
					RPC: &RPCOpts{
						RPCCodec:        utils.StringPointer("rpc"),
						ServiceMethod:   utils.StringPointer("service"),
						KeyPath:         utils.StringPointer("path"),
						CertPath:        utils.StringPointer("certpath"),
						CAPath:          utils.StringPointer("capath"),
						TLS:             utils.BoolPointer(true),
						ConnIDs:         utils.SliceStringPointer([]string{"id1", "id2"}),
						RPCConnTimeout:  utils.DurationPointer(1 * time.Minute),
						RPCReplyTimeout: utils.DurationPointer(1 * time.Minute),
						RPCAPIOpts: map[string]any{
							"key": "val",
						},
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
	if err := jsonCfg.eesCfg.loadFromJSONCfg(eventExporterJSON, jsonCfg.templates, jsonCfg.generalCfg.RSRSep, jsonCfg.dfltEvExp); err == nil || err.Error() != expected {
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
	if err := jsonCfg.eesCfg.loadFromJSONCfg(eventExporterJSON, jsonCfg.templates, jsonCfg.generalCfg.RSRSep, jsonCfg.dfltEvExp); err == nil || err.Error() != expected {
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
		ElsRefresh:               utils.StringPointer("true"),
		ElsOpType:                utils.StringPointer("test2"),
		ElsPipeline:              utils.StringPointer("test3"),
		ElsRouting:               utils.StringPointer("test4"),
		ElsTimeout:               utils.StringPointer("1m"),
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
		Els: &ElsOpts{
			Index:               utils.StringPointer("test"),
			Refresh:             utils.StringPointer("true"),
			OpType:              utils.StringPointer("test2"),
			Pipeline:            utils.StringPointer("test3"),
			Routing:             utils.StringPointer("test4"),
			Timeout:             utils.DurationPointer(1 * time.Minute),
			WaitForActiveShards: utils.StringPointer("test6"),
		},
		SQL: &SQLOpts{
			MaxIdleConns:    utils.IntPointer(4),
			MaxOpenConns:    utils.IntPointer(6),
			ConnMaxLifetime: utils.DurationPointer(1 * time.Minute),
			TableName:       utils.StringPointer("table"),
			DBName:          utils.StringPointer("db"),
			PgSSLMode:       utils.StringPointer("pg"),
		},
		AWS: &AWSOpts{
			Token:        utils.StringPointer("token"),
			S3FolderPath: utils.StringPointer("s3"),
		},
		AMQP:  &AMQPOpts{},
		Kafka: &KafkaOpts{},
		RPC:   &RPCOpts{},
		NATS: &NATSOpts{
			JetStream:            utils.BoolPointer(true),
			Subject:              utils.StringPointer("nat"),
			JWTFile:              utils.StringPointer("jwt"),
			SeedFile:             utils.StringPointer("seed"),
			CertificateAuthority: utils.StringPointer("NATS"),
			ClientCertificate:    utils.StringPointer("NATSClient"),
			ClientKey:            utils.StringPointer("key"),
			JetStreamMaxWait:     utils.DurationPointer(1 * time.Minute),
		},
	}
	eventExporter := &EventExporterCfg{
		Opts: &EventExporterOpts{
			Els:   &ElsOpts{},
			Kafka: &KafkaOpts{},
			AMQP:  &AMQPOpts{},
			NATS:  &NATSOpts{},
			SQL:   &SQLOpts{},
			RPC:   &RPCOpts{},
			AWS:   &AWSOpts{},
		},
	}
	if err := eventExporter.Opts.loadFromJSONCfg(eventExporterOptsJSON); err != nil {
		t.Error(expected)
	} else if !reflect.DeepEqual(expected, eventExporter.Opts) {
		t.Errorf("expected %v  received %v", utils.ToJSON(expected), utils.ToJSON(eventExporter.Opts))
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
	if err := jsonCfg.eesCfg.loadFromJSONCfg(eesCfg, jsonCfg.templates, jsonCfg.generalCfg.RSRSep, jsonCfg.dfltEvExp); err == nil || err.Error() != expected {
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
				StaticTTL: false,
			},
			utils.MetaSQL: {
				Limit:     -1,
				StaticTTL: false,
			},
			utils.MetaElastic: {
				Limit:     -1,
				StaticTTL: false,
			},
			utils.MetaAMQPV1jsonMap: {
				Limit:     -1,
				StaticTTL: false,
			},
			utils.MetaAMQPjsonMap: {
				Limit:     -1,
				StaticTTL: false,
			},
			utils.MetaS3jsonMap: {
				Limit:     -1,
				StaticTTL: false,
			},
			utils.MetaSQSjsonMap: {
				Limit:     -1,
				StaticTTL: false,
			},
			utils.MetaNatsjsonMap: {
				Limit:     -1,
				StaticTTL: false,
			},
			utils.MetaKafkajsonMap: {
				Limit:     -1,
				StaticTTL: false,
			},
		},
		Exporters: []*EventExporterCfg{
			{
				ID:            utils.MetaDefault,
				Type:          utils.MetaNone,
				ExportPath:    "/var/spool/cgrates/ees",
				Attempts:      1,
				Timezone:      utils.EmptyString,
				Filters:       []string{},
				AttributeSIDs: []string{},
				Flags:         utils.FlagsWithParams{},
				Fields:        []*FCTemplate{},
				contentFields: []*FCTemplate{},
				headerFields:  []*FCTemplate{},
				trailerFields: []*FCTemplate{},
				Opts: &EventExporterOpts{
					AMQP:  &AMQPOpts{},
					AWS:   &AWSOpts{},
					RPC:   &RPCOpts{},
					Els:   &ElsOpts{},
					Kafka: &KafkaOpts{},
					NATS:  &NATSOpts{},
					SQL:   &SQLOpts{},
				},
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
				headerFields:  []*FCTemplate{},
				trailerFields: []*FCTemplate{},
				Opts: &EventExporterOpts{
					AMQP:  &AMQPOpts{},
					AWS:   &AWSOpts{},
					RPC:   &RPCOpts{},
					Els:   &ElsOpts{},
					Kafka: &KafkaOpts{},
					NATS:  &NATSOpts{},
					SQL:   &SQLOpts{},
				},
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
			utils.MetaSQL: {
				Limit:     -1,
				StaticTTL: false,
			},
			utils.MetaElastic: {
				Limit:     -1,
				StaticTTL: false,
			},
			utils.MetaAMQPV1jsonMap: {
				Limit:     -1,
				StaticTTL: false,
			},
			utils.MetaAMQPjsonMap: {
				Limit:     -1,
				StaticTTL: false,
			},
			utils.MetaS3jsonMap: {
				Limit:     -1,
				StaticTTL: false,
			},
			utils.MetaSQSjsonMap: {
				Limit:     -1,
				StaticTTL: false,
			},
			utils.MetaNatsjsonMap: {
				Limit:     -1,
				StaticTTL: false,
			},
			utils.MetaKafkajsonMap: {
				Limit:     -1,
				StaticTTL: false,
			},
		},
		Exporters: []*EventExporterCfg{
			{
				ID:            utils.MetaDefault,
				Type:          utils.MetaNone,
				ExportPath:    "/var/spool/cgrates/ees",
				Attempts:      1,
				Timezone:      utils.EmptyString,
				Filters:       []string{},
				AttributeSIDs: []string{},
				Flags:         utils.FlagsWithParams{},
				contentFields: []*FCTemplate{},
				Fields:        []*FCTemplate{},
				headerFields:  []*FCTemplate{},
				trailerFields: []*FCTemplate{},
				Opts: &EventExporterOpts{
					AMQP:  &AMQPOpts{},
					AWS:   &AWSOpts{},
					SQL:   &SQLOpts{},
					Kafka: &KafkaOpts{},
					RPC:   &RPCOpts{},
					Els:   &ElsOpts{},
					NATS:  &NATSOpts{},
				},
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
				Opts: &EventExporterOpts{
					AMQP:  &AMQPOpts{},
					AWS:   &AWSOpts{},
					SQL:   &SQLOpts{},
					Kafka: &KafkaOpts{},
					RPC:   &RPCOpts{},
					Els:   &ElsOpts{},
					NATS:  &NATSOpts{},
				},
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
			utils.MetaSQL: {
				Limit:     -1,
				StaticTTL: false,
			},
			utils.MetaElastic: {
				Limit:     -1,
				StaticTTL: false,
			},
			utils.MetaAMQPV1jsonMap: {
				Limit:     -1,
				StaticTTL: false,
			},
			utils.MetaAMQPjsonMap: {
				Limit:     -1,
				StaticTTL: false,
			},
			utils.MetaS3jsonMap: {
				Limit:     -1,
				StaticTTL: false,
			},
			utils.MetaSQSjsonMap: {
				Limit:     -1,
				StaticTTL: false,
			},
			utils.MetaNatsjsonMap: {
				Limit:     -1,
				StaticTTL: false,
			},
			utils.MetaKafkajsonMap: {
				Limit:     -1,
				StaticTTL: false,
			},
		},
		Exporters: []*EventExporterCfg{
			{
				ID:            utils.MetaDefault,
				Type:          utils.MetaNone,
				ExportPath:    "/var/spool/cgrates/ees",
				Attempts:      1,
				Timezone:      utils.EmptyString,
				Filters:       []string{},
				AttributeSIDs: []string{},
				Flags:         utils.FlagsWithParams{},
				contentFields: []*FCTemplate{},
				Fields:        []*FCTemplate{},
				headerFields:  []*FCTemplate{},
				trailerFields: []*FCTemplate{},
				Opts: &EventExporterOpts{
					Els:   &ElsOpts{},
					Kafka: &KafkaOpts{},
					AMQP:  &AMQPOpts{},
					SQL:   &SQLOpts{},
					AWS:   &AWSOpts{},
					NATS:  &NATSOpts{},
					RPC:   &RPCOpts{},
				},
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
				Opts: &EventExporterOpts{
					AMQP:  &AMQPOpts{},
					AWS:   &AWSOpts{},
					SQL:   &SQLOpts{},
					Kafka: &KafkaOpts{},
					RPC:   &RPCOpts{},
					Els:   &ElsOpts{},
					NATS:  &NATSOpts{},
				},
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
	if err := jsnCfg.eesCfg.loadFromJSONCfg(jsonCfg, msgTemplates, jsnCfg.generalCfg.RSRSep, jsnCfg.dfltEvExp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedCfg, jsnCfg.eesCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedCfg), utils.ToJSON(jsnCfg.eesCfg))
	}
}

func TestEEsCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
    "ees": {
        "enabled": true,
        "attributes_conns": [
            "*internal",
            "*conn2"
        ],
        "cache": {
            "*file_csv": {
                "limit": -2,
                "precache": false,
                "replicate": false,
                "ttl": "1s",
                "static_ttl": false
            }
        },
        "exporters": [
            {
                "id": "CSVExporter",
                "type": "*file_csv",
                "export_path": "/tmp/testCSV",
                "metrics_reset_schedule": "0 12 * * *",
                "opts": {
                    "elsIndex": "test",
                    "elsRefresh": "true",
                    "kafkaTopic": "test",
                    "elsOpType": "test2",
                    "elsPipeline": "test3",
                    "elsRouting": "test4",
                    "elsTimeout": "1m",
                    "elsWaitForActiveShards": "test6",
                    "sqlMaxIdleConns": 4,
                    "sqlMaxOpenConns": 6,
                    "sqlConnMaxLifetime": "1m",
                    "sqlTableName": "table",
                    "sqlDBName": "db",
                    "sqlUpdateIndexedFields": [
                        "id"
                    ],
                    "pgSSLMode": "pg",
                    "awsToken": "token",
                    "s3FolderPath": "s3",
                    "natsJetStream": true,
                    "natsSubject": "nat",
                    "natsJWTFile": "jwt",
                    "natsSeedFile": "seed",
                    "natsCertificateAuthority": "NATS",
                    "natsClientCertificate": "NATSClient",
                    "natsClientKey": "key",
                    "natsJetStreamMaxWait": "1m",
                    "amqpQueueID": "id",
                    "amqpRoutingKey": "key",
                    "amqpExchangeType": "type",
                    "amqpExchange": "exchange",
                    "awsRegion": "eu",
                    "awsKey": "key",
                    "awsSecret": "secretkey",
                    "sqsQueueID": "sqsid",
                    "s3BucketID": "s3",
                    "rpcCodec": "rpc",
                    "serviceMethod": "service",
                    "keyPath": "path",
                    "certPath": "certpath",
                    "caPath": "capath",
                    "tls": true,
                    "connIDs": [
                        "id1",
                        "id2"
                    ],
                    "rpcConnTimeout": "1m",
                    "rpcReplyTimeout": "1m",
                    "csvFieldSeparator": ",",
                    "mysqlDSNParams": {
                        "key": "param"
                    },
                    "rpcAPIOpts": {
                        "key": "val"
                    }
                },
                "timezone": "UTC",
                "filters": [],
                "flags": [
                    "randomFlag"
                ],
                "attribute_ids": [],
                "attribute_context": "",
                "synchronous": false,
                "attempts": 1,
                "field_separator": ",",
                "fields": [
                    {
                        "tag": "CGRID",
                        "path": "*exp.CGRID",
                        "type": "*variable",
                        "value": "~*req.CGRID"
                    }
                ]
            }
        ]
    }
}`
	eMap := map[string]any{
		utils.EnabledCfg:         true,
		utils.AttributeSConnsCfg: []string{utils.MetaInternal, "*conn2"},
		utils.CacheCfg: map[string]any{
			utils.MetaFileCSV: map[string]any{
				utils.LimitCfg:     -2,
				utils.PrecacheCfg:  false,
				utils.ReplicateCfg: false,
				utils.RemoteCfg:    false,
				utils.TTLCfg:       "1s",
				utils.StaticTTLCfg: false,
			},
			utils.MetaAMQPjsonMap: map[string]any{
				utils.LimitCfg:     -1,
				utils.PrecacheCfg:  false,
				utils.ReplicateCfg: false,
				utils.RemoteCfg:    false,
				utils.StaticTTLCfg: false,
			},
			utils.MetaAMQPV1jsonMap: map[string]any{
				utils.LimitCfg:     -1,
				utils.PrecacheCfg:  false,
				utils.ReplicateCfg: false,
				utils.RemoteCfg:    false,
				utils.StaticTTLCfg: false,
			},
			utils.MetaNatsjsonMap: map[string]any{
				utils.LimitCfg:     -1,
				utils.PrecacheCfg:  false,
				utils.ReplicateCfg: false,
				utils.RemoteCfg:    false,
				utils.StaticTTLCfg: false,
			},
			utils.MetaSQSjsonMap: map[string]any{
				utils.LimitCfg:     -1,
				utils.PrecacheCfg:  false,
				utils.ReplicateCfg: false,
				utils.RemoteCfg:    false,
				utils.StaticTTLCfg: false,
			},
			utils.MetaS3jsonMap: map[string]any{
				utils.LimitCfg:     -1,
				utils.PrecacheCfg:  false,
				utils.ReplicateCfg: false,
				utils.RemoteCfg:    false,
				utils.StaticTTLCfg: false,
			},
			utils.MetaElastic: map[string]any{
				utils.LimitCfg:     -1,
				utils.PrecacheCfg:  false,
				utils.ReplicateCfg: false,
				utils.RemoteCfg:    false,
				utils.StaticTTLCfg: false,
			},
			utils.MetaSQL: map[string]any{
				utils.LimitCfg:     -1,
				utils.PrecacheCfg:  false,
				utils.ReplicateCfg: false,
				utils.RemoteCfg:    false,
				utils.StaticTTLCfg: false,
			},
			utils.MetaKafkajsonMap: map[string]any{
				utils.LimitCfg:     -1,
				utils.PrecacheCfg:  false,
				utils.ReplicateCfg: false,
				utils.RemoteCfg:    false,
				utils.StaticTTLCfg: false,
			},
		},
		utils.ExportersCfg: []map[string]any{
			{
				utils.IDCfg:         "CSVExporter",
				utils.TypeCfg:       "*file_csv",
				utils.ExportPathCfg: "/tmp/testCSV",
				utils.OptsCfg: map[string]any{
					utils.KafkaTopic:                "test",
					utils.ElsIndex:                  "test",
					utils.ElsRefresh:                "true",
					utils.ElsOpType:                 "test2",
					utils.ElsPipeline:               "test3",
					utils.ElsRouting:                "test4",
					utils.ElsTimeout:                "1m0s",
					utils.ElsWaitForActiveShards:    "test6",
					utils.SQLMaxIdleConnsCfg:        4,
					utils.SQLMaxOpenConns:           6,
					utils.SQLConnMaxLifetime:        "1m0s",
					utils.SQLTableNameOpt:           "table",
					utils.SQLDBNameOpt:              "db",
					utils.SQLUpdateIndexedFieldsOpt: []string{"id"},
					utils.PgSSLModeCfg:              "pg",
					utils.AWSToken:                  "token",
					utils.S3FolderPath:              "s3",
					utils.NatsJetStream:             true,
					utils.NatsSubject:               "nat",
					utils.NatsJWTFile:               "jwt",
					utils.NatsSeedFile:              "seed",
					utils.NatsCertificateAuthority:  "NATS",
					utils.NatsClientCertificate:     "NATSClient",
					utils.NatsClientKey:             "key",
					utils.NatsJetStreamMaxWait:      "1m0s",
					utils.AMQPQueueID:               "id",
					utils.AMQPRoutingKey:            "key",
					utils.AMQPExchangeType:          "type",
					utils.AMQPExchange:              "exchange",
					utils.AWSRegion:                 "eu",
					utils.AWSKey:                    "key",
					utils.AWSSecret:                 "secretkey",
					utils.SQSQueueID:                "sqsid",
					utils.S3Bucket:                  "s3",
					utils.RpcCodec:                  "rpc",
					utils.ServiceMethod:             "service",
					utils.KeyPath:                   "path",
					utils.CertPath:                  "certpath",
					utils.CaPath:                    "capath",
					utils.Tls:                       true,
					utils.ConnIDs:                   []string{"id1", "id2"},
					utils.RpcConnTimeout:            "1m0s",
					utils.RpcReplyTimeout:           "1m0s",
					utils.CSVFieldSepOpt:            ",",
					utils.MYSQLDSNParams: map[string]string{
						"key": "param",
					},
					utils.RPCAPIOpts: map[string]any{
						"key": "val",
					},
				},
				utils.TimezoneCfg:             "UTC",
				utils.FiltersCfg:              []string{},
				utils.FlagsCfg:                []string{"randomFlag"},
				utils.AttributeIDsCfg:         []string{},
				utils.AttributeContextCfg:     utils.EmptyString,
				utils.SynchronousCfg:          false,
				utils.AttemptsCfg:             1,
				utils.ConcurrentRequestsCfg:   0,
				utils.MetricsResetScheduleCfg: "0 12 * * *",
				utils.FieldsCfg: []map[string]any{
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
		if len(rcv[utils.ExportersCfg].([]map[string]any)) != 2 {
			t.Errorf("Expected %+v, received %+v", 2, len(rcv[utils.ExportersCfg].([]map[string]any)))
		} else if !reflect.DeepEqual(eMap[utils.ExportersCfg].([]map[string]any)[0][utils.FieldsCfg].([]map[string]any)[0][utils.ValueCfg],
			rcv[utils.ExportersCfg].([]map[string]any)[1][utils.FieldsCfg].([]map[string]any)[0][utils.ValueCfg]) {
			t.Errorf("Expected %+v, received %+v", eMap[utils.ExportersCfg].([]map[string]any)[0][utils.FieldsCfg].([]map[string]any)[0][utils.ValueCfg],
				rcv[utils.ExportersCfg].([]map[string]any)[1][utils.FieldsCfg].([]map[string]any)[0][utils.ValueCfg])
		}
		rcv[utils.ExportersCfg].([]map[string]any)[1][utils.FieldsCfg] = nil
		eMap[utils.ExportersCfg].([]map[string]any)[0][utils.FieldsCfg] = nil
		if !reflect.DeepEqual(rcv[utils.ExportersCfg].([]map[string]any)[1],
			eMap[utils.ExportersCfg].([]map[string]any)[0]) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap[utils.ExportersCfg].([]map[string]any)[0]),
				utils.ToJSON(rcv[utils.ExportersCfg].([]map[string]any)[1]))
		}
		rcv[utils.ExportersCfg] = nil
		eMap[utils.ExportersCfg] = nil
		if !reflect.DeepEqual(rcv, eMap) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
		}
	}
}

func TestEEsCfgappendEEsExporters(t *testing.T) {
	str := ""
	eeS := &EEsCfg{
		Exporters: []*EventExporterCfg{},
	}

	err := eeS.appendEEsExporters(nil, nil, "", nil)
	if err != nil {
		t.Error(err)
	}

	exporters := &[]*EventExporterJsonCfg{{
		Id: &str,
	}}
	err = eeS.appendEEsExporters(exporters, map[string][]*FCTemplate{}, "", nil)
	if err != nil {
		t.Error(err)
	}
}

func TestEEsCfgNewEventExporterCfg(t *testing.T) {
	str := "test"
	rcv := NewEventExporterCfg(str, str, str, str, 1, nil)
	exp := &EventExporterCfg{
		ID:             str,
		Type:           str,
		ExportPath:     str,
		FailedPostsDir: str,
		Attempts:       1,
		Opts:           &EventExporterOpts{},
	}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestEEsCfgloadFromJSONCfg(t *testing.T) {
	str := "test"
	bl := true
	nm := 1
	tm := 1 * time.Second
	tms := "1s"
	elsOpts := &ElsOpts{}
	jsnCfg := &EventExporterOptsJson{
		CSVFieldSeparator:           &str,
		ElsCloud:                    &bl,
		ElsAPIKey:                   &str,
		ElsServiceToken:             &str,
		ElsCertificateFingerprint:   &str,
		ElsUsername:                 &str,
		ElsPassword:                 &str,
		ElsCAPath:                   &str,
		ElsDiscoverNodesOnStart:     &bl,
		ElsDiscoverNodesInterval:    &tms,
		ElsEnableDebugLogger:        &bl,
		ElsLogger:                   &str,
		ElsCompressRequestBody:      &bl,
		ElsCompressRequestBodyLevel: &nm,
		ElsRetryOnStatus:            &[]int{nm},
		ElsMaxRetries:               &nm,
		ElsDisableRetry:             &bl,
		ElsIndex:                    &str,
		ElsRefresh:                  &str,
		ElsOpType:                   &str,
		ElsPipeline:                 &str,
		ElsRouting:                  &str,
		ElsTimeout:                  &tms,
		ElsWaitForActiveShards:      &str,
		SQLMaxIdleConns:             &nm,
		SQLMaxOpenConns:             &nm,
		SQLConnMaxLifetime:          &str,
		MYSQLDSNParams:              map[string]string{str: str},
		SQLTableName:                &str,
		SQLUpdateIndexedFields:      &[]string{"id"},
		SQLDBName:                   &str,
		PgSSLMode:                   &str,
		KafkaTopic:                  &str,
		AMQPQueueID:                 &str,
		AMQPRoutingKey:              &str,
		AMQPExchange:                &str,
		AMQPExchangeType:            &str,
		AMQPUsername:                &str,
		AMQPPassword:                &str,
		AWSRegion:                   &str,
		AWSKey:                      &str,
		AWSSecret:                   &str,
		AWSToken:                    &str,
		SQSQueueID:                  &str,
		S3BucketID:                  &str,
		S3FolderPath:                &str,
		NATSJetStream:               &bl,
		NATSSubject:                 &str,
		NATSJWTFile:                 &str,
		NATSSeedFile:                &str,
		NATSCertificateAuthority:    &str,
		NATSClientCertificate:       &str,
		NATSClientKey:               &str,
		NATSJetStreamMaxWait:        &str,
		RPCCodec:                    &str,
		ServiceMethod:               &str,
		KeyPath:                     &str,
		CertPath:                    &str,
		CAPath:                      &str,
		ConnIDs:                     &[]string{str},
		TLS:                         &bl,
		RPCConnTimeout:              &str,
		RPCReplyTimeout:             &str,
		RPCAPIOpts:                  map[string]any{str: nm},
	}
	exp := &ElsOpts{
		Index:                    &str,
		Refresh:                  &str,
		CAPath:                   &str,
		DiscoverNodesOnStart:     &bl,
		DiscoverNodeInterval:     &tm,
		Cloud:                    &bl,
		APIKey:                   &str,
		CertificateFingerprint:   &str,
		ServiceToken:             &str,
		Username:                 &str,
		Password:                 &str,
		EnableDebugLogger:        &bl,
		Logger:                   &str,
		CompressRequestBody:      &bl,
		CompressRequestBodyLevel: &nm,
		RetryOnStatus:            &[]int{nm},
		MaxRetries:               &nm,
		DisableRetry:             &bl,
		OpType:                   &str,
		Pipeline:                 &str,
		Routing:                  &str,
		Timeout:                  &tm,
		WaitForActiveShards:      &str,
	}
	err := elsOpts.loadFromJSONCfg(jsnCfg)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(elsOpts, exp) {
		t.Errorf("\nexpected %s\nreceived  %s\n", utils.ToJSON(exp), utils.ToJSON(elsOpts))
	}

	jsnCfg2 := &EventExporterOptsJson{
		ElsDiscoverNodesInterval: &str,
	}
	err = elsOpts.loadFromJSONCfg(jsnCfg2)
	if err != nil {
		if err.Error() != `time: invalid duration "test"` {
			t.Error(err)
		}
	} else {
		t.Error("was expecting an error")
	}
}

func TestEEsCfgAMQPOptsloadFromJSONCfg(t *testing.T) {
	str := "test"
	amqpOpts := &AMQPOpts{}
	jsnCfg := &EventExporterOptsJson{
		AMQPQueueID:      &str,
		AMQPRoutingKey:   &str,
		AMQPExchange:     &str,
		AMQPExchangeType: &str,
		AMQPUsername:     &str,
		AMQPPassword:     &str,
	}
	exp := &AMQPOpts{
		RoutingKey:   &str,
		QueueID:      &str,
		Exchange:     &str,
		ExchangeType: &str,
		Username:     &str,
		Password:     &str,
	}
	err := amqpOpts.loadFromJSONCfg(jsnCfg)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(amqpOpts, exp) {
		t.Errorf("\nexpected %s\nreceived  %s\n", utils.ToJSON(exp), utils.ToJSON(amqpOpts))
	}
}

func TestEEsCfgAMQPOptsClone(t *testing.T) {
	str := "test"
	AMQPOpts := &AMQPOpts{
		RoutingKey:   &str,
		QueueID:      &str,
		Exchange:     &str,
		ExchangeType: &str,
		Username:     &str,
		Password:     &str,
	}
	rcv := AMQPOpts.Clone()

	if !reflect.DeepEqual(rcv, AMQPOpts) {
		t.Errorf("\nexpected %s\nreceived  %s\n", utils.ToJSON(AMQPOpts), utils.ToJSON(rcv))
	}
}

func TestEventExporterCfgAsMapInterface(t *testing.T) {
	str := "test"
	am := &AMQPOpts{
		RoutingKey:   &str,
		QueueID:      &str,
		Exchange:     &str,
		ExchangeType: &str,
		Username:     &str,
		Password:     &str,
	}
	eeC := &EventExporterCfg{
		Opts: &EventExporterOpts{
			AMQP: am,
		},
	}
	opts := map[string]any{
		utils.AMQPQueueID:      str,
		utils.AMQPRoutingKey:   str,
		utils.AMQPExchange:     str,
		utils.AMQPExchangeType: str,
		utils.AMQPUsername:     str,
		utils.AMQPPassword:     str,
	}
	exp := map[string]any{
		utils.IDCfg:                   eeC.ID,
		utils.TypeCfg:                 eeC.Type,
		utils.ExportPathCfg:           eeC.ExportPath,
		utils.TimezoneCfg:             eeC.Timezone,
		utils.FiltersCfg:              eeC.Filters,
		utils.FlagsCfg:                []string{},
		utils.AttributeContextCfg:     eeC.AttributeSCtx,
		utils.AttributeIDsCfg:         eeC.AttributeSIDs,
		utils.SynchronousCfg:          eeC.Synchronous,
		utils.AttemptsCfg:             eeC.Attempts,
		utils.ConcurrentRequestsCfg:   eeC.ConcurrentRequests,
		utils.MetricsResetScheduleCfg: eeC.MetricsResetSchedule,
		utils.FailedPostsDirCfg:       eeC.FailedPostsDir,
		utils.OptsCfg:                 opts,
	}
	rcv := eeC.AsMapInterface("")

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nexpected %s\nreceived  %s\n", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestExporterCfg(t *testing.T) {
	testCfg := &EEsCfg{
		Exporters: []*EventExporterCfg{
			{ID: "exporter1", Type: "type1"},
			{ID: "exporter2", Type: "type2"},
			{ID: "exporter3", Type: "type1"},
		},
	}

	testCases := []struct {
		id     string
		want   *EventExporterCfg
		errMsg string
	}{
		{"exporter1", testCfg.Exporters[0], ""},
		{"exporter2", testCfg.Exporters[1], ""},
		{"nonexistent", nil, "Exporter with ID 'nonexistent' not found"},
		{"", nil, "Exporter ID cannot be empty"},
	}

	for _, tc := range testCases {
		got := testCfg.ExporterCfg(tc.id)
		if got != tc.want {
			t.Errorf("ExporterCfg(%q): expected %v, got %v; %s", tc.id, tc.want, got, tc.errMsg)
		}
	}
}

func TestLoadFromJSONCfg(t *testing.T) {

	jsonCfg := &EventExporterOptsJson{
		KafkaTopic:         utils.StringPointer("topic"),
		KafkaBatchSize:     utils.IntPointer(10),
		KafkaTLS:           utils.BoolPointer(true),
		KafkaCAPath:        utils.StringPointer("/path/to/ca"),
		KafkaSkipTLSVerify: utils.BoolPointer(false),
	}

	kafkaOpts := &KafkaOpts{}

	err := kafkaOpts.loadFromJSONCfg(jsonCfg)
	if err != nil {
		t.Errorf("Error loading KafkaOpts from JSON: %v", err)
	}

	if *kafkaOpts.Topic != "topic" {
		t.Errorf("Expected KafkaTopic to be 'topic', got %s", *kafkaOpts.Topic)
	}
	if *kafkaOpts.BatchSize != 10 {
		t.Errorf("Expected KafkaBatchSize to be 10, got %d", *kafkaOpts.BatchSize)
	}
	if *kafkaOpts.TLS != true {
		t.Errorf("Expected KafkaTLS to be true, got %v", *kafkaOpts.TLS)
	}
	if *kafkaOpts.CAPath != "/path/to/ca" {
		t.Errorf("Expected KafkaCAPath to be '/path/to/ca', got %s", *kafkaOpts.CAPath)
	}
	if *kafkaOpts.SkipTLSVerify != false {
		t.Errorf("Expected KafkaSkipTLSVerify to be false, got %v", *kafkaOpts.SkipTLSVerify)
	}
}

func TestKafkaOptsClone(t *testing.T) {

	originalOpts := &KafkaOpts{
		Topic:         utils.StringPointer("topic"),
		BatchSize:     utils.IntPointer(10),
		TLS:           utils.BoolPointer(true),
		CAPath:        utils.StringPointer("/ca/path"),
		SkipTLSVerify: utils.BoolPointer(false),
	}

	clonedOpts := originalOpts.Clone()

	if *clonedOpts.Topic != *originalOpts.Topic {
		t.Errorf("Expected Topic to be copied, got %s vs %s", *clonedOpts.Topic, *originalOpts.Topic)
	}
	if *clonedOpts.BatchSize != *originalOpts.BatchSize {
		t.Errorf("Expected BatchSize to be copied, got %d vs %d", *clonedOpts.BatchSize, *originalOpts.BatchSize)
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
