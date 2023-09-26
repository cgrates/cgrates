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
	"time"

	"github.com/cgrates/cgrates/utils"
)

// ERsCfg the config for ERs
type ERsCfg struct {
	Enabled         bool
	SessionSConns   []string
	Readers         []*EventReaderCfg
	PartialCacheTTL time.Duration
}

func (erS *ERsCfg) loadFromJSONCfg(jsnCfg *ERsJsonCfg, msgTemplates map[string][]*FCTemplate, sep string, dfltRdrCfg *EventReaderCfg, separator string) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		erS.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Sessions_conns != nil {
		erS.SessionSConns = make([]string, len(*jsnCfg.Sessions_conns))
		for i, fID := range *jsnCfg.Sessions_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			erS.SessionSConns[i] = fID
			if fID == utils.MetaInternal {
				erS.SessionSConns[i] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)
			}
		}
	}
	if jsnCfg.Partial_cache_ttl != nil {
		if erS.PartialCacheTTL, err = utils.ParseDurationWithNanosecs(*jsnCfg.Partial_cache_ttl); err != nil {
			return
		}
	}
	return erS.appendERsReaders(jsnCfg.Readers, msgTemplates, sep, dfltRdrCfg)
}

func (erS *ERsCfg) appendERsReaders(jsnReaders *[]*EventReaderJsonCfg, msgTemplates map[string][]*FCTemplate, sep string,
	dfltRdrCfg *EventReaderCfg) (err error) {
	if jsnReaders == nil {
		return
	}
	for _, jsnReader := range *jsnReaders {
		var rdr *EventReaderCfg
		if jsnReader.Id != nil {
			for _, reader := range erS.Readers {
				if reader.ID == *jsnReader.Id {
					rdr = reader
					break
				}
			}
		}
		if rdr == nil {
			if dfltRdrCfg != nil {
				rdr = dfltRdrCfg.Clone()
			} else {
				rdr = new(EventReaderCfg)
				rdr.Opts = &EventReaderOpts{}
			}
			erS.Readers = append(erS.Readers, rdr)
		}
		if err := rdr.loadFromJSONCfg(jsnReader, msgTemplates, sep); err != nil {
			return err
		}

	}
	return nil
}

// Clone returns a deep copy of ERsCfg
func (erS *ERsCfg) Clone() (cln *ERsCfg) {
	cln = &ERsCfg{
		Enabled:         erS.Enabled,
		SessionSConns:   make([]string, len(erS.SessionSConns)),
		Readers:         make([]*EventReaderCfg, len(erS.Readers)),
		PartialCacheTTL: erS.PartialCacheTTL,
	}
	for idx, sConn := range erS.SessionSConns {
		cln.SessionSConns[idx] = sConn
	}
	for idx, rdr := range erS.Readers {
		cln.Readers[idx] = rdr.Clone()
	}
	return
}

// AsMapInterface returns the config as a map[string]any
func (erS *ERsCfg) AsMapInterface(separator string) (initialMP map[string]any) {
	initialMP = map[string]any{
		utils.EnabledCfg:         erS.Enabled,
		utils.PartialCacheTTLCfg: "0",
	}
	if erS.PartialCacheTTL != 0 {
		initialMP[utils.PartialCacheTTLCfg] = erS.PartialCacheTTL.String()
	}
	if erS.SessionSConns != nil {
		sessionSConns := make([]string, len(erS.SessionSConns))
		for i, item := range erS.SessionSConns {
			sessionSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS) {
				sessionSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.SessionSConnsCfg] = sessionSConns
	}
	if erS.Readers != nil {
		readers := make([]map[string]any, len(erS.Readers))
		for i, item := range erS.Readers {
			readers[i] = item.AsMapInterface(separator)
		}
		initialMP[utils.ReadersCfg] = readers
	}
	return
}

type AMQPROpts struct {
	AMQPQueueID               *string
	AMQPQueueIDProcessed      *string
	AMQPUsername              *string
	AMQPPassword              *string
	AMQPUsernameProcessed     *string
	AMQPPasswordProcessed     *string
	AMQPConsumerTag           *string
	AMQPExchange              *string
	AMQPExchangeType          *string
	AMQPRoutingKey            *string
	AMQPExchangeProcessed     *string
	AMQPExchangeTypeProcessed *string
	AMQPRoutingKeyProcessed   *string
}

func (amqpr *AMQPROpts) loadFromJSONCfg(jsnCfg *EventReaderOptsJson) (err error) {
	if jsnCfg.AMQPQueueID != nil {
		amqpr.AMQPQueueID = jsnCfg.AMQPQueueID
	}
	if jsnCfg.AMQPQueueIDProcessed != nil {
		amqpr.AMQPQueueIDProcessed = jsnCfg.AMQPQueueIDProcessed
	}
	if jsnCfg.AMQPUsername != nil {
		amqpr.AMQPUsername = jsnCfg.AMQPUsername
	}
	if jsnCfg.AMQPPassword != nil {
		amqpr.AMQPPassword = jsnCfg.AMQPPassword
	}
	if jsnCfg.AMQPUsernameProcessed != nil {
		amqpr.AMQPUsernameProcessed = jsnCfg.AMQPUsernameProcessed
	}
	if jsnCfg.AMQPPasswordProcessed != nil {
		amqpr.AMQPPasswordProcessed = jsnCfg.AMQPPasswordProcessed
	}
	if jsnCfg.AMQPConsumerTag != nil {
		amqpr.AMQPConsumerTag = jsnCfg.AMQPConsumerTag
	}
	if jsnCfg.AMQPExchange != nil {
		amqpr.AMQPExchange = jsnCfg.AMQPExchange
	}
	if jsnCfg.AMQPExchangeType != nil {
		amqpr.AMQPExchangeType = jsnCfg.AMQPExchangeType
	}
	if jsnCfg.AMQPRoutingKey != nil {
		amqpr.AMQPRoutingKey = jsnCfg.AMQPRoutingKey
	}
	if jsnCfg.AMQPExchangeProcessed != nil {
		amqpr.AMQPExchangeProcessed = jsnCfg.AMQPExchangeProcessed
	}
	if jsnCfg.AMQPExchangeTypeProcessed != nil {
		amqpr.AMQPExchangeTypeProcessed = jsnCfg.AMQPExchangeTypeProcessed
	}
	if jsnCfg.AMQPRoutingKeyProcessed != nil {
		amqpr.AMQPRoutingKeyProcessed = jsnCfg.AMQPRoutingKeyProcessed
	}
	return
}

type KafkaROpts struct {
	KafkaTopic          *string
	KafkaGroupID        *string
	KafkaMaxWait        *time.Duration
	KafkaTopicProcessed *string
}

func (kafkaROpts *KafkaROpts) loadFromJSONCfg(jsnCfg *EventReaderOptsJson) (err error) {
	if jsnCfg.KafkaTopic != nil {
		kafkaROpts.KafkaTopic = jsnCfg.KafkaTopic
	}
	if jsnCfg.KafkaGroupID != nil {
		kafkaROpts.KafkaGroupID = jsnCfg.KafkaGroupID
	}
	if jsnCfg.KafkaMaxWait != nil {
		var kafkaMaxWait time.Duration
		if kafkaMaxWait, err = utils.ParseDurationWithNanosecs(*jsnCfg.KafkaMaxWait); err != nil {
			return
		}
		kafkaROpts.KafkaMaxWait = utils.DurationPointer(kafkaMaxWait)
	}
	if jsnCfg.KafkaTopicProcessed != nil {
		kafkaROpts.KafkaTopicProcessed = jsnCfg.KafkaTopicProcessed
	}
	return
}

type SQLROpts struct {
	SQLDBName             *string
	SQLTableName          *string
	PgSSLMode             *string
	SQLDBNameProcessed    *string
	SQLTableNameProcessed *string
	PgSSLModeProcessed    *string
}

func (sqlOpts *SQLROpts) loadFromJSONCfg(jsnCfg *EventReaderOptsJson) (err error) {
	if jsnCfg.SQLDBName != nil {
		sqlOpts.SQLDBName = jsnCfg.SQLDBName
	}
	if jsnCfg.SQLTableName != nil {
		sqlOpts.SQLTableName = jsnCfg.SQLTableName
	}
	if jsnCfg.PgSSLMode != nil {
		sqlOpts.PgSSLMode = jsnCfg.PgSSLMode
	}
	if jsnCfg.SQLDBNameProcessed != nil {
		sqlOpts.SQLDBNameProcessed = jsnCfg.SQLDBNameProcessed
	}
	if jsnCfg.SQLTableNameProcessed != nil {
		sqlOpts.SQLTableNameProcessed = jsnCfg.SQLTableNameProcessed
	}
	if jsnCfg.PgSSLModeProcessed != nil {
		sqlOpts.PgSSLModeProcessed = jsnCfg.PgSSLModeProcessed
	}
	return
}

type AWSROpts struct {
	AWSRegion             *string
	AWSKey                *string
	AWSSecret             *string
	AWSToken              *string
	AWSRegionProcessed    *string
	AWSKeyProcessed       *string
	AWSSecretProcessed    *string
	AWSTokenProcessed     *string
	SQSQueueID            *string
	SQSQueueIDProcessed   *string
	S3BucketID            *string
	S3FolderPathProcessed *string
	S3BucketIDProcessed   *string
}

func (awsROpts *AWSROpts) loadFromJSONCfg(jsnCfg *EventReaderOptsJson) (err error) {

	if jsnCfg.AWSRegion != nil {
		awsROpts.AWSRegion = jsnCfg.AWSRegion
	}
	if jsnCfg.AWSKey != nil {
		awsROpts.AWSKey = jsnCfg.AWSKey
	}
	if jsnCfg.AWSSecret != nil {
		awsROpts.AWSSecret = jsnCfg.AWSSecret
	}
	if jsnCfg.AWSToken != nil {
		awsROpts.AWSToken = jsnCfg.AWSToken
	}
	if jsnCfg.AWSRegionProcessed != nil {
		awsROpts.AWSRegionProcessed = jsnCfg.AWSRegionProcessed
	}
	if jsnCfg.AWSKeyProcessed != nil {
		awsROpts.AWSKeyProcessed = jsnCfg.AWSKeyProcessed
	}
	if jsnCfg.AWSSecretProcessed != nil {
		awsROpts.AWSSecretProcessed = jsnCfg.AWSSecretProcessed
	}
	if jsnCfg.AWSTokenProcessed != nil {
		awsROpts.AWSTokenProcessed = jsnCfg.AWSTokenProcessed
	}
	if jsnCfg.SQSQueueID != nil {
		awsROpts.SQSQueueID = jsnCfg.SQSQueueID
	}
	if jsnCfg.SQSQueueIDProcessed != nil {
		awsROpts.SQSQueueIDProcessed = jsnCfg.SQSQueueIDProcessed
	}
	if jsnCfg.S3BucketID != nil {
		awsROpts.S3BucketID = jsnCfg.S3BucketID
	}
	if jsnCfg.S3FolderPathProcessed != nil {
		awsROpts.S3FolderPathProcessed = jsnCfg.S3FolderPathProcessed
	}
	if jsnCfg.S3BucketIDProcessed != nil {
		awsROpts.S3BucketIDProcessed = jsnCfg.S3BucketIDProcessed
	}
	return
}

type NATSROpts struct {
	JetStream                     *bool
	ConsumerName                  *string
	StreamName                    *string
	Subject                       *string
	QueueID                       *string
	JWTFile                       *string
	SeedFile                      *string
	CertificateAuthority          *string
	ClientCertificate             *string
	ClientKey                     *string
	JetStreamMaxWait              *time.Duration
	JetStreamProcessed            *bool
	SubjectProcessed              *string
	JWTFileProcessed              *string
	SeedFileProcessed             *string
	CertificateAuthorityProcessed *string
	ClientCertificateProcessed    *string
	ClientKeyProcessed            *string
	JetStreamMaxWaitProcessed     *time.Duration
}

func (natsOpts *NATSROpts) loadFromJSONCfg(jsnCfg *EventReaderOptsJson) (err error) {
	if jsnCfg.NATSJetStream != nil {
		natsOpts.JetStream = jsnCfg.NATSJetStream
	}
	if jsnCfg.NATSConsumerName != nil {
		natsOpts.ConsumerName = jsnCfg.NATSConsumerName
	}
	if jsnCfg.NATSStreamName != nil {
		natsOpts.StreamName = jsnCfg.NATSStreamName
	}
	if jsnCfg.NATSSubject != nil {
		natsOpts.Subject = jsnCfg.NATSSubject
	}
	if jsnCfg.NATSQueueID != nil {
		natsOpts.QueueID = jsnCfg.NATSQueueID
	}
	if jsnCfg.NATSJWTFile != nil {
		natsOpts.JWTFile = jsnCfg.NATSJWTFile
	}
	if jsnCfg.NATSSeedFile != nil {
		natsOpts.SeedFile = jsnCfg.NATSSeedFile
	}
	if jsnCfg.NATSCertificateAuthority != nil {
		natsOpts.CertificateAuthority = jsnCfg.NATSCertificateAuthority
	}
	if jsnCfg.NATSClientCertificate != nil {
		natsOpts.ClientCertificate = jsnCfg.NATSClientCertificate
	}
	if jsnCfg.NATSClientKey != nil {
		natsOpts.ClientKey = jsnCfg.NATSClientKey
	}
	if jsnCfg.NATSJetStreamMaxWait != nil {
		var jetStreamMaxWait time.Duration
		if jetStreamMaxWait, err = utils.ParseDurationWithNanosecs(*jsnCfg.NATSJetStreamMaxWait); err != nil {
			return
		}
		natsOpts.JetStreamMaxWait = utils.DurationPointer(jetStreamMaxWait)
	}
	if jsnCfg.NATSJetStreamProcessed != nil {
		natsOpts.JetStreamProcessed = jsnCfg.NATSJetStreamProcessed
	}
	if jsnCfg.NATSSubjectProcessed != nil {
		natsOpts.SubjectProcessed = jsnCfg.NATSSubjectProcessed
	}
	if jsnCfg.NATSJWTFileProcessed != nil {
		natsOpts.JWTFileProcessed = jsnCfg.NATSJWTFileProcessed
	}
	if jsnCfg.NATSSeedFileProcessed != nil {
		natsOpts.SeedFileProcessed = jsnCfg.NATSSeedFileProcessed
	}
	if jsnCfg.NATSCertificateAuthorityProcessed != nil {
		natsOpts.CertificateAuthorityProcessed = jsnCfg.NATSCertificateAuthorityProcessed
	}
	if jsnCfg.NATSClientCertificateProcessed != nil {
		natsOpts.ClientCertificateProcessed = jsnCfg.NATSClientCertificateProcessed
	}
	if jsnCfg.NATSClientKeyProcessed != nil {
		natsOpts.ClientKeyProcessed = jsnCfg.NATSClientKeyProcessed
	}
	if jsnCfg.NATSJetStreamMaxWaitProcessed != nil {
		var jetStreamMaxWait time.Duration
		if jetStreamMaxWait, err = utils.ParseDurationWithNanosecs(*jsnCfg.NATSJetStreamMaxWaitProcessed); err != nil {
			return
		}
		natsOpts.JetStreamMaxWaitProcessed = utils.DurationPointer(jetStreamMaxWait)
	}
	return
}

type CSVROpts struct {
	PartialCSVFieldSeparator *string
	CSVRowLength             *int
	CSVFieldSeparator        *string
	CSVHeaderDefineChar      *string
	CSVLazyQuotes            *bool
}

func (csvROpts *CSVROpts) loadFromJSONCfg(jsnCfg *EventReaderOptsJson) (err error) {
	if jsnCfg.PartialCSVFieldSeparator != nil {
		csvROpts.PartialCSVFieldSeparator = jsnCfg.PartialCSVFieldSeparator
	}
	if jsnCfg.CSVRowLength != nil {
		csvROpts.CSVRowLength = jsnCfg.CSVRowLength
	}
	if jsnCfg.CSVFieldSeparator != nil {
		csvROpts.CSVFieldSeparator = jsnCfg.CSVFieldSeparator
	}
	if jsnCfg.CSVHeaderDefineChar != nil {
		csvROpts.CSVHeaderDefineChar = jsnCfg.CSVHeaderDefineChar
	}
	if jsnCfg.CSVLazyQuotes != nil {
		csvROpts.CSVLazyQuotes = jsnCfg.CSVLazyQuotes
	}
	return
}

type EventReaderOpts struct {
	PartialPath        *string
	PartialCacheAction *string
	PartialOrderField  *string
	XMLRootPath        *string
	CSVOpts            *CSVROpts
	AMQPOpts           *AMQPROpts
	AWSOpts            *AWSROpts
	NATSOpts           *NATSROpts
	KafkaOpts          *KafkaROpts
	SQLOpts            *SQLROpts
}

// EventReaderCfg the event for the Event Reader
type EventReaderCfg struct {
	ID                  string
	Type                string
	RunDelay            time.Duration
	ConcurrentReqs      int
	SourcePath          string
	ProcessedPath       string
	Opts                *EventReaderOpts
	Tenant              RSRParsers
	Timezone            string
	Filters             []string
	Flags               utils.FlagsWithParams
	Fields              []*FCTemplate
	PartialCommitFields []*FCTemplate
	CacheDumpFields     []*FCTemplate
}

func (erOpts *EventReaderOpts) loadFromJSONCfg(jsnCfg *EventReaderOptsJson) (err error) {
	if jsnCfg == nil {
		return
	}
	if err = erOpts.AMQPOpts.loadFromJSONCfg(jsnCfg); err != nil {
		return
	}
	if err = erOpts.AWSOpts.loadFromJSONCfg(jsnCfg); err != nil {
		return
	}
	if err = erOpts.KafkaOpts.loadFromJSONCfg(jsnCfg); err != nil {
		return
	}
	if err = erOpts.NATSOpts.loadFromJSONCfg(jsnCfg); err != nil {
		return
	}
	if err = erOpts.SQLOpts.loadFromJSONCfg(jsnCfg); err != nil {
		return
	}
	if err = erOpts.CSVOpts.loadFromJSONCfg(jsnCfg); err != nil {
		return
	}
	if jsnCfg.PartialPath != nil {
		erOpts.PartialPath = jsnCfg.PartialPath
	}
	if jsnCfg.PartialCacheAction != nil {
		erOpts.PartialCacheAction = jsnCfg.PartialCacheAction
	}
	if jsnCfg.PartialOrderField != nil {
		erOpts.PartialOrderField = jsnCfg.PartialOrderField
	}

	if jsnCfg.XMLRootPath != nil {
		erOpts.XMLRootPath = jsnCfg.XMLRootPath
	}

	return
}

func (er *EventReaderCfg) loadFromJSONCfg(jsnCfg *EventReaderJsonCfg, msgTemplates map[string][]*FCTemplate, sep string) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Id != nil {
		er.ID = *jsnCfg.Id
	}
	if jsnCfg.Type != nil {
		er.Type = *jsnCfg.Type
	}
	if jsnCfg.Run_delay != nil {
		if er.RunDelay, err = utils.ParseDurationWithNanosecs(*jsnCfg.Run_delay); err != nil {
			return
		}
	}
	if jsnCfg.Concurrent_requests != nil {
		er.ConcurrentReqs = *jsnCfg.Concurrent_requests
	}
	if jsnCfg.Source_path != nil {
		er.SourcePath = *jsnCfg.Source_path
	}
	if jsnCfg.Processed_path != nil {
		er.ProcessedPath = *jsnCfg.Processed_path
	}
	if jsnCfg.Tenant != nil {
		if er.Tenant, err = NewRSRParsers(*jsnCfg.Tenant, sep); err != nil {
			return err
		}
	}
	if jsnCfg.Timezone != nil {
		er.Timezone = *jsnCfg.Timezone
	}
	if jsnCfg.Filters != nil {
		er.Filters = make([]string, len(*jsnCfg.Filters))
		for i, fltr := range *jsnCfg.Filters {
			er.Filters[i] = fltr
		}
	}
	if jsnCfg.Flags != nil {
		er.Flags = utils.FlagsWithParamsFromSlice(*jsnCfg.Flags)
	}
	if jsnCfg.Fields != nil {
		if er.Fields, err = FCTemplatesFromFCTemplatesJSONCfg(*jsnCfg.Fields, sep); err != nil {
			return err
		}
		if tpls, err := InflateTemplates(er.Fields, msgTemplates); err != nil {
			return err
		} else if tpls != nil {
			er.Fields = tpls
		}
	}
	if jsnCfg.Cache_dump_fields != nil {
		if er.CacheDumpFields, err = FCTemplatesFromFCTemplatesJSONCfg(*jsnCfg.Cache_dump_fields, sep); err != nil {
			return err
		}
		if tpls, err := InflateTemplates(er.CacheDumpFields, msgTemplates); err != nil {
			return err
		} else if tpls != nil {
			er.CacheDumpFields = tpls
		}
	}
	if jsnCfg.Partial_commit_fields != nil {
		if er.PartialCommitFields, err = FCTemplatesFromFCTemplatesJSONCfg(*jsnCfg.Partial_commit_fields, sep); err != nil {
			return err
		}
		if tpls, err := InflateTemplates(er.PartialCommitFields, msgTemplates); err != nil {
			return err
		} else if tpls != nil {
			er.PartialCommitFields = tpls
		}
	}
	if jsnCfg.Opts != nil {
		err = er.Opts.loadFromJSONCfg(jsnCfg.Opts)
	}
	return
}
func (amqpOpts *AMQPROpts) Clone() *AMQPROpts {
	cln := &AMQPROpts{}
	if amqpOpts.AMQPQueueID != nil {
		cln.AMQPQueueID = new(string)
		*cln.AMQPQueueID = *amqpOpts.AMQPQueueID
	}
	if amqpOpts.AMQPQueueIDProcessed != nil {
		cln.AMQPQueueIDProcessed = new(string)
		*cln.AMQPQueueIDProcessed = *amqpOpts.AMQPQueueIDProcessed
	}
	if amqpOpts.AMQPUsername != nil {
		cln.AMQPUsername = new(string)
		*cln.AMQPUsername = *amqpOpts.AMQPUsername
	}
	if amqpOpts.AMQPPassword != nil {
		cln.AMQPPassword = new(string)
		*cln.AMQPPassword = *amqpOpts.AMQPPassword
	}
	if amqpOpts.AMQPUsernameProcessed != nil {
		cln.AMQPUsernameProcessed = new(string)
		*cln.AMQPUsernameProcessed = *amqpOpts.AMQPUsernameProcessed
	}
	if amqpOpts.AMQPPasswordProcessed != nil {
		cln.AMQPPasswordProcessed = new(string)
		*cln.AMQPPasswordProcessed = *amqpOpts.AMQPPasswordProcessed
	}
	if amqpOpts.AMQPConsumerTag != nil {
		cln.AMQPConsumerTag = new(string)
		*cln.AMQPConsumerTag = *amqpOpts.AMQPConsumerTag
	}
	if amqpOpts.AMQPExchange != nil {
		cln.AMQPExchange = new(string)
		*cln.AMQPExchange = *amqpOpts.AMQPExchange
	}
	if amqpOpts.AMQPExchangeType != nil {
		cln.AMQPExchangeType = new(string)
		*cln.AMQPExchangeType = *amqpOpts.AMQPExchangeType
	}
	if amqpOpts.AMQPRoutingKey != nil {
		cln.AMQPRoutingKey = new(string)
		*cln.AMQPRoutingKey = *amqpOpts.AMQPRoutingKey
	}
	if amqpOpts.AMQPExchangeProcessed != nil {
		cln.AMQPExchangeProcessed = new(string)
		*cln.AMQPExchangeProcessed = *amqpOpts.AMQPExchangeProcessed
	}
	if amqpOpts.AMQPExchangeTypeProcessed != nil {
		cln.AMQPExchangeTypeProcessed = new(string)
		*cln.AMQPExchangeTypeProcessed = *amqpOpts.AMQPExchangeTypeProcessed
	}
	if amqpOpts.AMQPRoutingKeyProcessed != nil {
		cln.AMQPRoutingKeyProcessed = new(string)
		*cln.AMQPRoutingKeyProcessed = *amqpOpts.AMQPRoutingKeyProcessed
	}
	return cln
}

func (csvOpts *CSVROpts) Clone() *CSVROpts {
	cln := &CSVROpts{}
	if csvOpts.PartialCSVFieldSeparator != nil {
		cln.PartialCSVFieldSeparator = new(string)
		*cln.PartialCSVFieldSeparator = *csvOpts.PartialCSVFieldSeparator
	}
	if csvOpts.CSVRowLength != nil {
		cln.CSVRowLength = new(int)
		*cln.CSVRowLength = *csvOpts.CSVRowLength
	}
	if csvOpts.CSVFieldSeparator != nil {
		cln.CSVFieldSeparator = new(string)
		*cln.CSVFieldSeparator = *csvOpts.CSVFieldSeparator
	}
	if csvOpts.CSVHeaderDefineChar != nil {
		cln.CSVHeaderDefineChar = new(string)
		*cln.CSVHeaderDefineChar = *csvOpts.CSVHeaderDefineChar
	}
	if csvOpts.CSVLazyQuotes != nil {
		cln.CSVLazyQuotes = new(bool)
		*cln.CSVLazyQuotes = *csvOpts.CSVLazyQuotes
	}
	return cln
}
func (kafkaOpts *KafkaROpts) Clone() *KafkaROpts {
	cln := &KafkaROpts{}
	if kafkaOpts.KafkaTopic != nil {
		cln.KafkaTopic = new(string)
		*cln.KafkaTopic = *kafkaOpts.KafkaTopic
	}
	if kafkaOpts.KafkaGroupID != nil {
		cln.KafkaGroupID = new(string)
		*cln.KafkaGroupID = *kafkaOpts.KafkaGroupID
	}
	if kafkaOpts.KafkaMaxWait != nil {
		cln.KafkaMaxWait = new(time.Duration)
		*cln.KafkaMaxWait = *kafkaOpts.KafkaMaxWait
	}
	if kafkaOpts.KafkaTopicProcessed != nil {
		cln.KafkaTopicProcessed = new(string)
		*cln.KafkaTopicProcessed = *kafkaOpts.KafkaTopicProcessed
	}
	return cln
}

func (sqlOpts *SQLROpts) Clone() *SQLROpts {
	cln := &SQLROpts{}
	if sqlOpts.SQLDBName != nil {
		cln.SQLDBName = new(string)
		*cln.SQLDBName = *sqlOpts.SQLDBName
	}
	if sqlOpts.SQLTableName != nil {
		cln.SQLTableName = new(string)
		*cln.SQLTableName = *sqlOpts.SQLTableName
	}
	if sqlOpts.PgSSLMode != nil {
		cln.PgSSLMode = new(string)
		*cln.PgSSLMode = *sqlOpts.PgSSLMode
	}
	if sqlOpts.SQLDBNameProcessed != nil {
		cln.SQLDBNameProcessed = new(string)
		*cln.SQLDBNameProcessed = *sqlOpts.SQLDBNameProcessed
	}
	if sqlOpts.SQLTableNameProcessed != nil {
		cln.SQLTableNameProcessed = new(string)
		*cln.SQLTableNameProcessed = *sqlOpts.SQLTableNameProcessed
	}
	if sqlOpts.PgSSLModeProcessed != nil {
		cln.PgSSLModeProcessed = new(string)
		*cln.PgSSLModeProcessed = *sqlOpts.PgSSLModeProcessed
	}
	return cln
}

func (awsOpt *AWSROpts) Clone() *AWSROpts {
	cln := &AWSROpts{}
	if awsOpt.AWSRegion != nil {
		cln.AWSRegion = new(string)
		*cln.AWSRegion = *awsOpt.AWSRegion
	}
	if awsOpt.AWSKey != nil {
		cln.AWSKey = new(string)
		*cln.AWSKey = *awsOpt.AWSKey
	}
	if awsOpt.AWSSecret != nil {
		cln.AWSSecret = new(string)
		*cln.AWSSecret = *awsOpt.AWSSecret
	}
	if awsOpt.AWSToken != nil {
		cln.AWSToken = new(string)
		*cln.AWSToken = *awsOpt.AWSToken
	}
	if awsOpt.AWSRegionProcessed != nil {
		cln.AWSRegionProcessed = new(string)
		*cln.AWSRegionProcessed = *awsOpt.AWSRegionProcessed
	}
	if awsOpt.AWSKeyProcessed != nil {
		cln.AWSKeyProcessed = new(string)
		*cln.AWSKeyProcessed = *awsOpt.AWSKeyProcessed
	}
	if awsOpt.AWSSecretProcessed != nil {
		cln.AWSSecretProcessed = new(string)
		*cln.AWSSecretProcessed = *awsOpt.AWSSecretProcessed
	}
	if awsOpt.AWSTokenProcessed != nil {
		cln.AWSTokenProcessed = new(string)
		*cln.AWSTokenProcessed = *awsOpt.AWSTokenProcessed
	}
	if awsOpt.SQSQueueID != nil {
		cln.SQSQueueID = new(string)
		*cln.SQSQueueID = *awsOpt.SQSQueueID
	}
	if awsOpt.SQSQueueIDProcessed != nil {
		cln.SQSQueueIDProcessed = new(string)
		*cln.SQSQueueIDProcessed = *awsOpt.SQSQueueIDProcessed
	}
	if awsOpt.S3BucketID != nil {
		cln.S3BucketID = new(string)
		*cln.S3BucketID = *awsOpt.S3BucketID
	}
	if awsOpt.S3FolderPathProcessed != nil {
		cln.S3FolderPathProcessed = new(string)
		*cln.S3FolderPathProcessed = *awsOpt.S3FolderPathProcessed
	}
	if awsOpt.S3BucketIDProcessed != nil {
		cln.S3BucketIDProcessed = new(string)
		*cln.S3BucketIDProcessed = *awsOpt.S3BucketIDProcessed
	}
	return cln
}
func (natOpts *NATSROpts) Clone() *NATSROpts {
	cln := &NATSROpts{}
	if natOpts.JetStream != nil {
		cln.JetStream = new(bool)
		*cln.JetStream = *natOpts.JetStream
	}
	if natOpts.ConsumerName != nil {
		cln.ConsumerName = new(string)
		*cln.ConsumerName = *natOpts.ConsumerName
	}
	if natOpts.StreamName != nil {
		cln.StreamName = new(string)
		*cln.StreamName = *natOpts.StreamName
	}
	if natOpts.Subject != nil {
		cln.Subject = new(string)
		*cln.Subject = *natOpts.Subject
	}
	if natOpts.QueueID != nil {
		cln.QueueID = new(string)
		*cln.QueueID = *natOpts.QueueID
	}
	if natOpts.JWTFile != nil {
		cln.JWTFile = new(string)
		*cln.JWTFile = *natOpts.JWTFile
	}
	if natOpts.SeedFile != nil {
		cln.SeedFile = new(string)
		*cln.SeedFile = *natOpts.SeedFile
	}
	if natOpts.CertificateAuthority != nil {
		cln.CertificateAuthority = new(string)
		*cln.CertificateAuthority = *natOpts.CertificateAuthority
	}
	if natOpts.ClientCertificate != nil {
		cln.ClientCertificate = new(string)
		*cln.ClientCertificate = *natOpts.ClientCertificate
	}
	if natOpts.ClientKey != nil {
		cln.ClientKey = new(string)
		*cln.ClientKey = *natOpts.ClientKey
	}
	if natOpts.JetStreamMaxWait != nil {
		cln.JetStreamMaxWait = new(time.Duration)
		*cln.JetStreamMaxWait = *natOpts.JetStreamMaxWait
	}
	if natOpts.JetStreamProcessed != nil {
		cln.JetStreamProcessed = new(bool)
		*cln.JetStreamProcessed = *natOpts.JetStreamProcessed
	}
	if natOpts.SubjectProcessed != nil {
		cln.SubjectProcessed = new(string)
		*cln.SubjectProcessed = *natOpts.SubjectProcessed
	}
	if natOpts.JWTFileProcessed != nil {
		cln.JWTFileProcessed = new(string)
		*cln.JWTFileProcessed = *natOpts.JWTFileProcessed
	}
	if natOpts.SeedFileProcessed != nil {
		cln.SeedFileProcessed = new(string)
		*cln.SeedFileProcessed = *natOpts.SeedFileProcessed
	}
	if natOpts.CertificateAuthorityProcessed != nil {
		cln.CertificateAuthorityProcessed = new(string)
		*cln.CertificateAuthorityProcessed = *natOpts.CertificateAuthorityProcessed
	}
	if natOpts.ClientCertificateProcessed != nil {
		cln.ClientCertificateProcessed = new(string)
		*cln.ClientCertificateProcessed = *natOpts.ClientCertificateProcessed
	}
	if natOpts.ClientKeyProcessed != nil {
		cln.ClientKeyProcessed = new(string)
		*cln.ClientKeyProcessed = *natOpts.ClientKeyProcessed
	}
	if natOpts.JetStreamMaxWaitProcessed != nil {
		cln.JetStreamMaxWaitProcessed = new(time.Duration)
		*cln.JetStreamMaxWaitProcessed = *natOpts.JetStreamMaxWaitProcessed
	}
	return cln
}

func (erOpts *EventReaderOpts) Clone() *EventReaderOpts {
	cln := &EventReaderOpts{}
	if erOpts.PartialPath != nil {
		cln.PartialPath = new(string)
		*cln.PartialPath = *erOpts.PartialPath
	}
	if erOpts.PartialCacheAction != nil {
		cln.PartialCacheAction = new(string)
		*cln.PartialCacheAction = *erOpts.PartialCacheAction
	}
	if erOpts.PartialOrderField != nil {
		cln.PartialOrderField = new(string)
		*cln.PartialOrderField = *erOpts.PartialOrderField
	}
	if erOpts.CSVOpts != nil {
		cln.CSVOpts = erOpts.CSVOpts.Clone()
	}
	if erOpts.XMLRootPath != nil {
		cln.XMLRootPath = new(string)
		*cln.XMLRootPath = *erOpts.XMLRootPath
	}
	if erOpts.AMQPOpts != nil {
		cln.AMQPOpts = erOpts.AMQPOpts.Clone()
	}
	if erOpts.NATSOpts != nil {
		cln.NATSOpts = erOpts.NATSOpts.Clone()
	}
	if erOpts.KafkaOpts != nil {
		cln.KafkaOpts = erOpts.KafkaOpts.Clone()
	}
	if erOpts.SQLOpts != nil {
		cln.SQLOpts = erOpts.SQLOpts.Clone()
	}
	if erOpts.AWSOpts != nil {
		cln.AWSOpts = erOpts.AWSOpts.Clone()
	}

	return cln
}

// Clone returns a deep copy of EventReaderCfg
func (er EventReaderCfg) Clone() (cln *EventReaderCfg) {
	cln = &EventReaderCfg{
		ID:             er.ID,
		Type:           er.Type,
		RunDelay:       er.RunDelay,
		ConcurrentReqs: er.ConcurrentReqs,
		SourcePath:     er.SourcePath,
		ProcessedPath:  er.ProcessedPath,
		Tenant:         er.Tenant.Clone(),
		Timezone:       er.Timezone,
		Flags:          er.Flags.Clone(),
		Opts:           er.Opts.Clone(),
	}
	if er.Filters != nil {
		cln.Filters = make([]string, len(er.Filters))
		for idx, val := range er.Filters {
			cln.Filters[idx] = val
		}
	}
	if er.Fields != nil {
		cln.Fields = make([]*FCTemplate, len(er.Fields))
		for idx, fld := range er.Fields {
			cln.Fields[idx] = fld.Clone()
		}
	}
	if er.CacheDumpFields != nil {
		cln.CacheDumpFields = make([]*FCTemplate, len(er.CacheDumpFields))
		for idx, fld := range er.CacheDumpFields {
			cln.CacheDumpFields[idx] = fld.Clone()
		}
	}
	if er.PartialCommitFields != nil {
		cln.PartialCommitFields = make([]*FCTemplate, len(er.PartialCommitFields))
		for idx, fld := range er.PartialCommitFields {
			cln.PartialCommitFields[idx] = fld.Clone()
		}
	}
	return
}

// AsMapInterface returns the config as a map[string]any
func (er *EventReaderCfg) AsMapInterface(separator string) (initialMP map[string]any) {
	opts := map[string]any{}

	if er.Opts.PartialPath != nil {
		opts[utils.PartialPathOpt] = *er.Opts.PartialPath
	}
	if er.Opts.PartialCacheAction != nil {
		opts[utils.PartialCacheActionOpt] = *er.Opts.PartialCacheAction
	}
	if er.Opts.PartialOrderField != nil {
		opts[utils.PartialOrderFieldOpt] = *er.Opts.PartialOrderField
	}

	if csvOpts := er.Opts.CSVOpts; csvOpts != nil {
		if csvOpts.PartialCSVFieldSeparator != nil {
			opts[utils.PartialCSVFieldSepartorOpt] = *csvOpts.PartialCSVFieldSeparator
		}
		if csvOpts.CSVRowLength != nil {
			opts[utils.CSVRowLengthOpt] = *csvOpts.CSVRowLength
		}
		if csvOpts.CSVFieldSeparator != nil {
			opts[utils.CSVFieldSepOpt] = *csvOpts.CSVFieldSeparator
		}
		if csvOpts.CSVHeaderDefineChar != nil {
			opts[utils.HeaderDefineCharOpt] = *csvOpts.CSVHeaderDefineChar
		}
		if csvOpts.CSVLazyQuotes != nil {
			opts[utils.CSVLazyQuotes] = *csvOpts.CSVLazyQuotes
		}
	}
	if er.Opts.XMLRootPath != nil {
		opts[utils.XMLRootPathOpt] = *er.Opts.XMLRootPath
	}
	if amqpOpts := er.Opts.AMQPOpts; amqpOpts != nil {
		if amqpOpts.AMQPQueueID != nil {
			opts[utils.AMQPQueueID] = *amqpOpts.AMQPQueueID
		}
		if amqpOpts.AMQPQueueIDProcessed != nil {
			opts[utils.AMQPQueueIDProcessedCfg] = *amqpOpts.AMQPQueueIDProcessed
		}
		if amqpOpts.AMQPUsername != nil {
			opts[utils.AMQPUsername] = *amqpOpts.AMQPUsername
		}
		if amqpOpts.AMQPPassword != nil {
			opts[utils.AMQPPassword] = *amqpOpts.AMQPPassword
		}
		if amqpOpts.AMQPUsernameProcessed != nil {
			opts[utils.AMQPUsernameProcessedCfg] = *amqpOpts.AMQPUsernameProcessed
		}
		if amqpOpts.AMQPPasswordProcessed != nil {
			opts[utils.AMQPPasswordProcessedCfg] = *amqpOpts.AMQPPasswordProcessed
		}
		if amqpOpts.AMQPConsumerTag != nil {
			opts[utils.AMQPConsumerTag] = *amqpOpts.AMQPConsumerTag
		}
		if amqpOpts.AMQPExchange != nil {
			opts[utils.AMQPExchange] = *amqpOpts.AMQPExchange
		}
		if amqpOpts.AMQPExchangeType != nil {
			opts[utils.AMQPExchangeType] = *amqpOpts.AMQPExchangeType
		}
		if amqpOpts.AMQPRoutingKey != nil {
			opts[utils.AMQPRoutingKey] = *amqpOpts.AMQPRoutingKey
		}
		if amqpOpts.AMQPExchangeProcessed != nil {
			opts[utils.AMQPExchangeProcessedCfg] = *amqpOpts.AMQPExchangeProcessed
		}
		if amqpOpts.AMQPExchangeTypeProcessed != nil {
			opts[utils.AMQPExchangeTypeProcessedCfg] = *amqpOpts.AMQPExchangeTypeProcessed
		}
		if amqpOpts.AMQPRoutingKeyProcessed != nil {
			opts[utils.AMQPRoutingKeyProcessedCfg] = *amqpOpts.AMQPRoutingKeyProcessed
		}
	}

	if kafkaOpts := er.Opts.KafkaOpts; kafkaOpts != nil {
		if kafkaOpts.KafkaTopic != nil {
			opts[utils.KafkaTopic] = *kafkaOpts.KafkaTopic
		}
		if kafkaOpts.KafkaGroupID != nil {
			opts[utils.KafkaGroupID] = *kafkaOpts.KafkaGroupID
		}
		if kafkaOpts.KafkaMaxWait != nil {
			opts[utils.KafkaMaxWait] = kafkaOpts.KafkaMaxWait.String()
		}
		if kafkaOpts.KafkaTopicProcessed != nil {
			opts[utils.KafkaTopicProcessedCfg] = *kafkaOpts.KafkaTopicProcessed
		}
	}

	if sqlOpts := er.Opts.SQLOpts; sqlOpts != nil {
		if sqlOpts.SQLDBName != nil {
			opts[utils.SQLDBNameOpt] = *sqlOpts.SQLDBName
		}
		if sqlOpts.SQLTableName != nil {
			opts[utils.SQLTableNameOpt] = *sqlOpts.SQLTableName
		}
		if sqlOpts.PgSSLMode != nil {
			opts[utils.PgSSLModeCfg] = *sqlOpts.PgSSLMode
		}
		if sqlOpts.SQLDBNameProcessed != nil {
			opts[utils.SQLDBNameProcessedCfg] = *sqlOpts.SQLDBNameProcessed
		}
		if sqlOpts.SQLTableNameProcessed != nil {
			opts[utils.SQLTableNameProcessedCfg] = *sqlOpts.SQLTableNameProcessed
		}
		if sqlOpts.PgSSLModeProcessed != nil {
			opts[utils.PgSSLModeProcessedCfg] = *sqlOpts.PgSSLModeProcessed
		}
	}

	if awsOpts := er.Opts.AWSOpts; awsOpts != nil {
		if awsOpts.AWSRegion != nil {
			opts[utils.AWSRegion] = *awsOpts.AWSRegion
		}
		if awsOpts.AWSKey != nil {
			opts[utils.AWSKey] = *awsOpts.AWSKey
		}
		if awsOpts.AWSSecret != nil {
			opts[utils.AWSSecret] = *awsOpts.AWSSecret
		}
		if awsOpts.AWSToken != nil {
			opts[utils.AWSToken] = *awsOpts.AWSToken
		}
		if awsOpts.AWSRegionProcessed != nil {
			opts[utils.AWSRegionProcessedCfg] = *awsOpts.AWSRegionProcessed
		}
		if awsOpts.AWSKeyProcessed != nil {
			opts[utils.AWSKeyProcessedCfg] = *awsOpts.AWSKeyProcessed
		}
		if awsOpts.AWSSecretProcessed != nil {
			opts[utils.AWSSecretProcessedCfg] = *awsOpts.AWSSecretProcessed
		}
		if awsOpts.AWSTokenProcessed != nil {
			opts[utils.AWSTokenProcessedCfg] = *awsOpts.AWSTokenProcessed
		}
		if awsOpts.SQSQueueID != nil {
			opts[utils.SQSQueueID] = *awsOpts.SQSQueueID
		}
		if awsOpts.SQSQueueIDProcessed != nil {
			opts[utils.SQSQueueIDProcessedCfg] = *awsOpts.SQSQueueIDProcessed
		}
		if awsOpts.S3BucketID != nil {
			opts[utils.S3Bucket] = *awsOpts.S3BucketID
		}
		if awsOpts.S3FolderPathProcessed != nil {
			opts[utils.S3FolderPathProcessedCfg] = *awsOpts.S3FolderPathProcessed
		}
		if awsOpts.S3BucketIDProcessed != nil {
			opts[utils.S3BucketIDProcessedCfg] = *awsOpts.S3BucketIDProcessed
		}
	}

	if natsOpts := er.Opts.NATSOpts; natsOpts != nil {
		if natsOpts.JetStream != nil {
			opts[utils.NatsJetStream] = *natsOpts.JetStream
		}
		if natsOpts.ConsumerName != nil {
			opts[utils.NatsConsumerName] = *natsOpts.ConsumerName
		}
		if natsOpts.StreamName != nil {
			opts[utils.NatsStreamName] = *natsOpts.StreamName
		}
		if natsOpts.Subject != nil {
			opts[utils.NatsSubject] = *natsOpts.Subject
		}
		if natsOpts.QueueID != nil {
			opts[utils.NatsQueueID] = *natsOpts.QueueID
		}
		if natsOpts.JWTFile != nil {
			opts[utils.NatsJWTFile] = *natsOpts.JWTFile
		}
		if natsOpts.SeedFile != nil {
			opts[utils.NatsSeedFile] = *natsOpts.SeedFile
		}
		if natsOpts.CertificateAuthority != nil {
			opts[utils.NatsCertificateAuthority] = *natsOpts.CertificateAuthority
		}
		if natsOpts.ClientCertificate != nil {
			opts[utils.NatsClientCertificate] = *natsOpts.ClientCertificate
		}
		if natsOpts.ClientKey != nil {
			opts[utils.NatsClientKey] = *natsOpts.ClientKey
		}
		if natsOpts.JetStreamMaxWait != nil {
			opts[utils.NatsJetStreamMaxWait] = natsOpts.JetStreamMaxWait.String()
		}
		if natsOpts.JetStreamProcessed != nil {
			opts[utils.NATSJetStreamProcessedCfg] = *natsOpts.JetStreamProcessed
		}
		if natsOpts.SubjectProcessed != nil {
			opts[utils.NATSSubjectProcessedCfg] = *natsOpts.SubjectProcessed
		}
		if natsOpts.JWTFileProcessed != nil {
			opts[utils.NATSJWTFileProcessedCfg] = *natsOpts.JWTFileProcessed
		}
		if natsOpts.SeedFileProcessed != nil {
			opts[utils.NATSSeedFileProcessedCfg] = *natsOpts.SeedFileProcessed
		}
		if natsOpts.CertificateAuthorityProcessed != nil {
			opts[utils.NATSCertificateAuthorityProcessedCfg] = *natsOpts.CertificateAuthorityProcessed
		}
		if natsOpts.ClientCertificateProcessed != nil {
			opts[utils.NATSClientCertificateProcessed] = *natsOpts.ClientCertificateProcessed
		}
		if natsOpts.ClientKeyProcessed != nil {
			opts[utils.NATSClientKeyProcessedCfg] = *natsOpts.ClientKeyProcessed
		}
		if natsOpts.JetStreamMaxWaitProcessed != nil {
			opts[utils.NATSJetStreamMaxWaitProcessedCfg] = natsOpts.JetStreamMaxWaitProcessed.String()
		}
	}
	initialMP = map[string]any{
		utils.IDCfg:                 er.ID,
		utils.TypeCfg:               er.Type,
		utils.ConcurrentRequestsCfg: er.ConcurrentReqs,
		utils.SourcePathCfg:         er.SourcePath,
		utils.ProcessedPathCfg:      er.ProcessedPath,
		utils.TenantCfg:             er.Tenant.GetRule(separator),
		utils.TimezoneCfg:           er.Timezone,
		utils.FiltersCfg:            er.Filters,
		utils.FlagsCfg:              []string{},
		utils.RunDelayCfg:           "0",
		utils.OptsCfg:               opts,
	}

	initialMP[utils.OptsCfg] = opts

	if flags := er.Flags.SliceFlags(); flags != nil {
		initialMP[utils.FlagsCfg] = flags
	}

	if er.Fields != nil {
		fields := make([]map[string]any, len(er.Fields))
		for i, item := range er.Fields {
			fields[i] = item.AsMapInterface(separator)
		}
		initialMP[utils.FieldsCfg] = fields
	}
	if er.CacheDumpFields != nil {
		cacheDumpFields := make([]map[string]any, len(er.CacheDumpFields))
		for i, item := range er.CacheDumpFields {
			cacheDumpFields[i] = item.AsMapInterface(separator)
		}
		initialMP[utils.CacheDumpFieldsCfg] = cacheDumpFields
	}
	if er.PartialCommitFields != nil {
		parCFields := make([]map[string]any, len(er.PartialCommitFields))
		for i, item := range er.PartialCommitFields {
			parCFields[i] = item.AsMapInterface(separator)
		}
		initialMP[utils.PartialCommitFieldsCfg] = parCFields
	}

	if er.RunDelay > 0 {
		initialMP[utils.RunDelayCfg] = er.RunDelay.String()
	} else if er.RunDelay < 0 {
		initialMP[utils.RunDelayCfg] = "-1"
	}
	return
}
