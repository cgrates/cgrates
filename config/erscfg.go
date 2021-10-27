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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// ERsCfg the config for ERs
type ERsCfg struct {
	Enabled         bool
	SessionSConns   []string
	Readers         []*EventReaderCfg
	PartialCacheTTL time.Duration
}

// loadErsCfg loads the Ers section of the configuration
func (erS *ERsCfg) Load(ctx *context.Context, jsnCfg ConfigDB, cfg *CGRConfig) (err error) {
	jsnERsCfg := new(ERsJsonCfg)
	if err = jsnCfg.GetSection(ctx, ERsJSON, jsnERsCfg); err != nil {
		return
	}
	return erS.loadFromJSONCfg(jsnERsCfg, cfg.templates, cfg.generalCfg.RSRSep)
}

func (erS *ERsCfg) loadFromJSONCfg(jsnCfg *ERsJsonCfg, msgTemplates map[string][]*FCTemplate, sep string) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		erS.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Sessions_conns != nil {
		erS.SessionSConns = updateInternalConns(*jsnCfg.Sessions_conns, utils.MetaSessionS)
	}
	if jsnCfg.Partial_cache_ttl != nil {
		if erS.PartialCacheTTL, err = utils.ParseDurationWithNanosecs(*jsnCfg.Partial_cache_ttl); err != nil {
			return
		}
	}
	return erS.appendERsReaders(jsnCfg.Readers, msgTemplates, sep)
}

func (erS *ERsCfg) appendERsReaders(jsnReaders *[]*EventReaderJsonCfg, msgTemplates map[string][]*FCTemplate, sep string) (err error) {
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
			rdr = getDftEvRdrCfg()
			erS.Readers = append(erS.Readers, rdr)
		}
		if err := rdr.loadFromJSONCfg(jsnReader, msgTemplates, sep); err != nil {
			return err
		}

	}
	return nil
}
func (ERsCfg) SName() string             { return ERsJSON }
func (erS ERsCfg) CloneSection() Section { return erS.Clone() }

// Clone returns a deep copy of ERsCfg
func (erS ERsCfg) Clone() (cln *ERsCfg) {
	cln = &ERsCfg{
		Enabled:         erS.Enabled,
		SessionSConns:   make([]string, len(erS.SessionSConns)),
		Readers:         make([]*EventReaderCfg, len(erS.Readers)),
		PartialCacheTTL: erS.PartialCacheTTL,
	}
	if erS.SessionSConns != nil {
		cln.SessionSConns = utils.CloneStringSlice(erS.SessionSConns)
	}
	for idx, rdr := range erS.Readers {
		cln.Readers[idx] = rdr.Clone()
	}
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (erS ERsCfg) AsMapInterface(separator string) interface{} {
	mp := map[string]interface{}{
		utils.EnabledCfg:         erS.Enabled,
		utils.PartialCacheTTLCfg: "0",
	}
	if erS.PartialCacheTTL != 0 {
		mp[utils.PartialCacheTTLCfg] = erS.PartialCacheTTL.String()
	}
	if erS.SessionSConns != nil {
		mp[utils.SessionSConnsCfg] = getInternalJSONConns(erS.SessionSConns)
	}
	if erS.Readers != nil {
		readers := make([]map[string]interface{}, len(erS.Readers))
		for i, item := range erS.Readers {
			readers[i] = item.AsMapInterface(separator)
		}
		mp[utils.ReadersCfg] = readers
	}
	return mp
}

type EventReaderOpts struct {
	PartialPath                       *string
	PartialCacheAction                *string
	PartialOrderField                 *string
	PartialCSVFieldSeparator          *string
	CSVRowLength                      *int
	CSVFieldSeparator                 *string
	CSVHeaderDefineChar               *string
	CSVLazyQuotes                     *bool
	XMLRootPath                       *string
	AMQPQueueID                       *string
	AMQPQueueIDProcessed              *string
	AMQPConsumerTag                   *string
	AMQPExchange                      *string
	AMQPExchangeType                  *string
	AMQPRoutingKey                    *string
	AMQPExchangeProcessed             *string
	AMQPExchangeTypeProcessed         *string
	AMQPRoutingKeyProcessed           *string
	KafkaTopic                        *string
	KafkaGroupID                      *string
	KafkaMaxWait                      *time.Duration
	KafkaTopicProcessed               *string
	SQLDBName                         *string
	SQLTableName                      *string
	SSLMode                           *string
	SQLDBNameProcessed                *string
	SQLTableNameProcessed             *string
	SSLModeProcessed                  *string
	AWSRegion                         *string
	AWSKey                            *string
	AWSSecret                         *string
	AWSToken                          *string
	AWSRegionProcessed                *string
	AWSKeyProcessed                   *string
	AWSSecretProcessed                *string
	AWSTokenProcessed                 *string
	SQSQueueID                        *string
	SQSQueueIDProcessed               *string
	S3BucketID                        *string
	S3FolderPathProcessed             *string
	S3BucketIDProcessed               *string
	NATSJetStream                     *bool
	NATSConsumerName                  *string
	NATSSubject                       *string
	NATSQueueID                       *string
	NATSJWTFile                       *string
	NATSSeedFile                      *string
	NATSCertificateAuthority          *string
	NATSClientCertificate             *string
	NATSClientKey                     *string
	NATSJetStreamMaxWait              *time.Duration
	NATSJetStreamProcessed            *bool
	NATSSubjectProcessed              *string
	NATSJWTFileProcessed              *string
	NATSSeedFileProcessed             *string
	NATSCertificateAuthorityProcessed *string
	NATSClientCertificateProcessed    *string
	NATSClientKeyProcessed            *string
	NATSJetStreamMaxWaitProcessed     *time.Duration
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
	if jsnCfg.PartialPath != nil {
		erOpts.PartialPath = jsnCfg.PartialPath
	}
	if jsnCfg.PartialCacheAction != nil {
		erOpts.PartialCacheAction = jsnCfg.PartialCacheAction
	}
	if jsnCfg.PartialOrderField != nil {
		erOpts.PartialOrderField = jsnCfg.PartialOrderField
	}
	if jsnCfg.PartialCSVFieldSeparator != nil {
		erOpts.PartialCSVFieldSeparator = jsnCfg.PartialCSVFieldSeparator
	}
	if jsnCfg.CSVRowLength != nil {
		erOpts.CSVRowLength = jsnCfg.CSVRowLength
	}
	if jsnCfg.CSVFieldSeparator != nil {
		erOpts.CSVFieldSeparator = jsnCfg.CSVFieldSeparator
	}
	if jsnCfg.CSVHeaderDefineChar != nil {
		erOpts.CSVHeaderDefineChar = jsnCfg.CSVHeaderDefineChar
	}
	if jsnCfg.CSVLazyQuotes != nil {
		erOpts.CSVLazyQuotes = jsnCfg.CSVLazyQuotes
	}
	if jsnCfg.XMLRootPath != nil {
		erOpts.XMLRootPath = jsnCfg.XMLRootPath
	}
	if jsnCfg.AMQPQueueID != nil {
		erOpts.AMQPQueueID = jsnCfg.AMQPQueueID
	}
	if jsnCfg.AMQPQueueIDProcessed != nil {
		erOpts.AMQPQueueIDProcessed = jsnCfg.AMQPQueueIDProcessed
	}
	if jsnCfg.AMQPConsumerTag != nil {
		erOpts.AMQPConsumerTag = jsnCfg.AMQPConsumerTag
	}
	if jsnCfg.AMQPExchange != nil {
		erOpts.AMQPExchange = jsnCfg.AMQPExchange
	}
	if jsnCfg.AMQPExchangeType != nil {
		erOpts.AMQPExchangeType = jsnCfg.AMQPExchangeType
	}
	if jsnCfg.AMQPRoutingKey != nil {
		erOpts.AMQPRoutingKey = jsnCfg.AMQPRoutingKey
	}
	if jsnCfg.AMQPExchangeProcessed != nil {
		erOpts.AMQPExchangeProcessed = jsnCfg.AMQPExchangeProcessed
	}
	if jsnCfg.AMQPExchangeTypeProcessed != nil {
		erOpts.AMQPExchangeTypeProcessed = jsnCfg.AMQPExchangeTypeProcessed
	}
	if jsnCfg.AMQPRoutingKeyProcessed != nil {
		erOpts.AMQPRoutingKeyProcessed = jsnCfg.AMQPRoutingKeyProcessed
	}
	if jsnCfg.KafkaTopic != nil {
		erOpts.KafkaTopic = jsnCfg.KafkaTopic
	}
	if jsnCfg.KafkaGroupID != nil {
		erOpts.KafkaGroupID = jsnCfg.KafkaGroupID
	}
	if jsnCfg.KafkaMaxWait != nil {
		var kafkaMaxWait time.Duration
		if kafkaMaxWait, err = utils.ParseDurationWithNanosecs(*jsnCfg.KafkaMaxWait); err != nil {
			return
		}
		erOpts.KafkaMaxWait = utils.DurationPointer(kafkaMaxWait)
	}
	if jsnCfg.KafkaTopicProcessed != nil {
		erOpts.KafkaTopicProcessed = jsnCfg.KafkaTopicProcessed
	}
	if jsnCfg.SQLDBName != nil {
		erOpts.SQLDBName = jsnCfg.SQLDBName
	}
	if jsnCfg.SQLTableName != nil {
		erOpts.SQLTableName = jsnCfg.SQLTableName
	}
	if jsnCfg.SSLMode != nil {
		erOpts.SSLMode = jsnCfg.SSLMode
	}
	if jsnCfg.SQLDBNameProcessed != nil {
		erOpts.SQLDBNameProcessed = jsnCfg.SQLDBNameProcessed
	}
	if jsnCfg.SQLTableNameProcessed != nil {
		erOpts.SQLTableNameProcessed = jsnCfg.SQLTableNameProcessed
	}
	if jsnCfg.SSLModeProcessed != nil {
		erOpts.SSLModeProcessed = jsnCfg.SSLModeProcessed
	}
	if jsnCfg.AWSRegion != nil {
		erOpts.AWSRegion = jsnCfg.AWSRegion
	}
	if jsnCfg.AWSKey != nil {
		erOpts.AWSKey = jsnCfg.AWSKey
	}
	if jsnCfg.AWSSecret != nil {
		erOpts.AWSSecret = jsnCfg.AWSSecret
	}
	if jsnCfg.AWSToken != nil {
		erOpts.AWSToken = jsnCfg.AWSToken
	}
	if jsnCfg.AWSRegionProcessed != nil {
		erOpts.AWSRegionProcessed = jsnCfg.AWSRegionProcessed
	}
	if jsnCfg.AWSKeyProcessed != nil {
		erOpts.AWSKeyProcessed = jsnCfg.AWSKeyProcessed
	}
	if jsnCfg.AWSSecretProcessed != nil {
		erOpts.AWSSecretProcessed = jsnCfg.AWSSecretProcessed
	}
	if jsnCfg.AWSTokenProcessed != nil {
		erOpts.AWSTokenProcessed = jsnCfg.AWSTokenProcessed
	}
	if jsnCfg.SQSQueueID != nil {
		erOpts.SQSQueueID = jsnCfg.SQSQueueID
	}
	if jsnCfg.SQSQueueIDProcessed != nil {
		erOpts.SQSQueueIDProcessed = jsnCfg.SQSQueueIDProcessed
	}
	if jsnCfg.S3BucketID != nil {
		erOpts.S3BucketID = jsnCfg.S3BucketID
	}
	if jsnCfg.S3FolderPathProcessed != nil {
		erOpts.S3FolderPathProcessed = jsnCfg.S3FolderPathProcessed
	}
	if jsnCfg.S3BucketIDProcessed != nil {
		erOpts.S3BucketIDProcessed = jsnCfg.S3BucketIDProcessed
	}
	if jsnCfg.NATSJetStream != nil {
		erOpts.NATSJetStream = jsnCfg.NATSJetStream
	}
	if jsnCfg.NATSConsumerName != nil {
		erOpts.NATSConsumerName = jsnCfg.NATSConsumerName
	}
	if jsnCfg.NATSSubject != nil {
		erOpts.NATSSubject = jsnCfg.NATSSubject
	}
	if jsnCfg.NATSQueueID != nil {
		erOpts.NATSQueueID = jsnCfg.NATSQueueID
	}
	if jsnCfg.NATSJWTFile != nil {
		erOpts.NATSJWTFile = jsnCfg.NATSJWTFile
	}
	if jsnCfg.NATSSeedFile != nil {
		erOpts.NATSSeedFile = jsnCfg.NATSSeedFile
	}
	if jsnCfg.NATSCertificateAuthority != nil {
		erOpts.NATSCertificateAuthority = jsnCfg.NATSCertificateAuthority
	}
	if jsnCfg.NATSClientCertificate != nil {
		erOpts.NATSClientCertificate = jsnCfg.NATSClientCertificate
	}
	if jsnCfg.NATSClientKey != nil {
		erOpts.NATSClientKey = jsnCfg.NATSClientKey
	}
	if jsnCfg.NATSJetStreamMaxWait != nil {
		var jetStreamMaxWait time.Duration
		if jetStreamMaxWait, err = utils.ParseDurationWithNanosecs(*jsnCfg.NATSJetStreamMaxWait); err != nil {
			return
		}
		erOpts.NATSJetStreamMaxWait = utils.DurationPointer(jetStreamMaxWait)
	}
	if jsnCfg.NATSJetStreamProcessed != nil {
		erOpts.NATSJetStreamProcessed = jsnCfg.NATSJetStreamProcessed
	}
	if jsnCfg.NATSSubjectProcessed != nil {
		erOpts.NATSSubjectProcessed = jsnCfg.NATSSubjectProcessed
	}
	if jsnCfg.NATSJWTFileProcessed != nil {
		erOpts.NATSJWTFileProcessed = jsnCfg.NATSJWTFileProcessed
	}
	if jsnCfg.NATSSeedFileProcessed != nil {
		erOpts.NATSSeedFileProcessed = jsnCfg.NATSSeedFileProcessed
	}
	if jsnCfg.NATSCertificateAuthorityProcessed != nil {
		erOpts.NATSCertificateAuthorityProcessed = jsnCfg.NATSCertificateAuthorityProcessed
	}
	if jsnCfg.NATSClientCertificateProcessed != nil {
		erOpts.NATSClientCertificateProcessed = jsnCfg.NATSClientCertificateProcessed
	}
	if jsnCfg.NATSClientKeyProcessed != nil {
		erOpts.NATSClientKeyProcessed = jsnCfg.NATSClientKeyProcessed
	}
	if jsnCfg.NATSJetStreamMaxWaitProcessed != nil {
		var jetStreamMaxWait time.Duration
		if jetStreamMaxWait, err = utils.ParseDurationWithNanosecs(*jsnCfg.NATSJetStreamMaxWaitProcessed); err != nil {
			return
		}
		erOpts.NATSJetStreamMaxWaitProcessed = utils.DurationPointer(jetStreamMaxWait)
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
		er.Filters = utils.CloneStringSlice(*jsnCfg.Filters)
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

func (erOpts *EventReaderOpts) Clone() *EventReaderOpts {
	return &EventReaderOpts{
		PartialPath:                       erOpts.PartialPath,
		PartialCacheAction:                erOpts.PartialCacheAction,
		PartialOrderField:                 erOpts.PartialOrderField,
		PartialCSVFieldSeparator:          erOpts.PartialCSVFieldSeparator,
		CSVRowLength:                      erOpts.CSVRowLength,
		CSVFieldSeparator:                 erOpts.CSVFieldSeparator,
		CSVHeaderDefineChar:               erOpts.CSVHeaderDefineChar,
		CSVLazyQuotes:                     erOpts.CSVLazyQuotes,
		XMLRootPath:                       erOpts.XMLRootPath,
		AMQPQueueID:                       erOpts.AMQPQueueID,
		AMQPQueueIDProcessed:              erOpts.AMQPQueueIDProcessed,
		AMQPConsumerTag:                   erOpts.AMQPConsumerTag,
		AMQPExchange:                      erOpts.AMQPExchange,
		AMQPExchangeType:                  erOpts.AMQPExchangeType,
		AMQPRoutingKey:                    erOpts.AMQPRoutingKey,
		AMQPExchangeProcessed:             erOpts.AMQPExchangeProcessed,
		AMQPExchangeTypeProcessed:         erOpts.AMQPExchangeTypeProcessed,
		AMQPRoutingKeyProcessed:           erOpts.AMQPRoutingKeyProcessed,
		KafkaTopic:                        erOpts.KafkaTopic,
		KafkaGroupID:                      erOpts.KafkaGroupID,
		KafkaMaxWait:                      erOpts.KafkaMaxWait,
		KafkaTopicProcessed:               erOpts.KafkaTopicProcessed,
		SQLDBName:                         erOpts.SQLDBName,
		SQLTableName:                      erOpts.SQLTableName,
		SSLMode:                           erOpts.SSLMode,
		SQLDBNameProcessed:                erOpts.SQLDBNameProcessed,
		SQLTableNameProcessed:             erOpts.SQLTableNameProcessed,
		SSLModeProcessed:                  erOpts.SSLModeProcessed,
		AWSRegion:                         erOpts.AWSRegion,
		AWSKey:                            erOpts.AWSKey,
		AWSSecret:                         erOpts.AWSSecret,
		AWSToken:                          erOpts.AWSToken,
		AWSRegionProcessed:                erOpts.AWSRegionProcessed,
		AWSKeyProcessed:                   erOpts.AWSKeyProcessed,
		AWSSecretProcessed:                erOpts.AWSSecretProcessed,
		AWSTokenProcessed:                 erOpts.AWSTokenProcessed,
		SQSQueueID:                        erOpts.SQSQueueID,
		SQSQueueIDProcessed:               erOpts.SQSQueueIDProcessed,
		S3BucketID:                        erOpts.S3BucketID,
		S3FolderPathProcessed:             erOpts.S3FolderPathProcessed,
		S3BucketIDProcessed:               erOpts.S3BucketIDProcessed,
		NATSJetStream:                     erOpts.NATSJetStream,
		NATSConsumerName:                  erOpts.NATSConsumerName,
		NATSSubject:                       erOpts.NATSSubject,
		NATSQueueID:                       erOpts.NATSQueueID,
		NATSJWTFile:                       erOpts.NATSJWTFile,
		NATSSeedFile:                      erOpts.NATSSeedFile,
		NATSCertificateAuthority:          erOpts.NATSCertificateAuthority,
		NATSClientCertificate:             erOpts.NATSClientCertificate,
		NATSClientKey:                     erOpts.NATSClientKey,
		NATSJetStreamMaxWait:              erOpts.NATSJetStreamMaxWait,
		NATSJetStreamProcessed:            erOpts.NATSJetStreamProcessed,
		NATSSubjectProcessed:              erOpts.NATSSubjectProcessed,
		NATSJWTFileProcessed:              erOpts.NATSJWTFileProcessed,
		NATSSeedFileProcessed:             erOpts.NATSSeedFileProcessed,
		NATSCertificateAuthorityProcessed: erOpts.NATSCertificateAuthorityProcessed,
		NATSClientCertificateProcessed:    erOpts.NATSClientCertificateProcessed,
		NATSClientKeyProcessed:            erOpts.NATSClientKeyProcessed,
		NATSJetStreamMaxWaitProcessed:     erOpts.NATSJetStreamMaxWaitProcessed,
	}
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
		cln.Filters = utils.CloneStringSlice(er.Filters)
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

// AsMapInterface returns the config as a map[string]interface{}
func (er *EventReaderCfg) AsMapInterface(separator string) (initialMP map[string]interface{}) {
	opts := map[string]interface{}{
		utils.PartialPathOpt:                       er.Opts.PartialPath,
		utils.PartialCacheActionOpt:                er.Opts.PartialCacheAction,
		utils.PartialOrderFieldOpt:                 er.Opts.PartialOrderField,
		utils.PartialCSVFieldSepartorOpt:           er.Opts.PartialCSVFieldSeparator,
		utils.CSVRowLengthOpt:                      er.Opts.CSVRowLength,
		utils.CSVFieldSepOpt:                       er.Opts.CSVFieldSeparator,
		utils.HeaderDefineCharOpt:                  er.Opts.CSVHeaderDefineChar,
		utils.CSVLazyQuotes:                        er.Opts.CSVLazyQuotes,
		utils.XMLRootPathOpt:                       er.Opts.XMLRootPath,
		utils.AMQPQueueID:                          er.Opts.AMQPQueueID,
		utils.AMQPQueueIDProcessedCfg:              er.Opts.AMQPQueueIDProcessed,
		utils.AMQPConsumerTag:                      er.Opts.AMQPConsumerTag,
		utils.AMQPExchange:                         er.Opts.AMQPExchange,
		utils.AMQPExchangeType:                     er.Opts.AMQPExchangeType,
		utils.AMQPRoutingKey:                       er.Opts.AMQPRoutingKey,
		utils.AMQPExchangeProcessedCfg:             er.Opts.AMQPExchangeProcessed,
		utils.AMQPExchangeTypeProcessedCfg:         er.Opts.AMQPExchangeTypeProcessed,
		utils.AMQPRoutingKeyProcessedCfg:           er.Opts.AMQPRoutingKeyProcessed,
		utils.KafkaTopic:                           er.Opts.KafkaTopic,
		utils.KafkaGroupID:                         er.Opts.KafkaGroupID,
		utils.KafkaMaxWait:                         er.Opts.KafkaMaxWait,
		utils.KafkaTopicProcessedCfg:               er.Opts.KafkaTopicProcessed,
		utils.SQLDBNameOpt:                         er.Opts.SQLDBName,
		utils.SQLTableNameOpt:                      er.Opts.SQLTableName,
		utils.SSLModeCfg:                           er.Opts.SSLMode,
		utils.SQLDBNameProcessedCfg:                er.Opts.SQLDBNameProcessed,
		utils.SQLTableNameProcessedCfg:             er.Opts.SQLTableNameProcessed,
		utils.SSLModeProcessedCfg:                  er.Opts.SSLModeProcessed,
		utils.AWSRegion:                            er.Opts.AWSRegion,
		utils.AWSKey:                               er.Opts.AWSKey,
		utils.AWSSecret:                            er.Opts.AWSSecret,
		utils.AWSToken:                             er.Opts.AWSToken,
		utils.AWSRegionProcessedCfg:                er.Opts.AWSRegionProcessed,
		utils.AWSKeyProcessedCfg:                   er.Opts.AWSKeyProcessed,
		utils.AWSSecretProcessedCfg:                er.Opts.AWSSecretProcessed,
		utils.AWSTokenProcessedCfg:                 er.Opts.AWSTokenProcessed,
		utils.SQSQueueID:                           er.Opts.SQSQueueID,
		utils.SQSQueueIDProcessedCfg:               er.Opts.SQSQueueIDProcessed,
		utils.S3Bucket:                             er.Opts.S3BucketID,
		utils.S3FolderPathProcessedCfg:             er.Opts.S3FolderPathProcessed,
		utils.S3BucketIDProcessedCfg:               er.Opts.S3BucketIDProcessed,
		utils.NatsJetStream:                        er.Opts.NATSJetStream,
		utils.NatsConsumerName:                     er.Opts.NATSConsumerName,
		utils.NatsSubject:                          er.Opts.NATSSubject,
		utils.NatsQueueID:                          er.Opts.NATSQueueID,
		utils.NatsJWTFile:                          er.Opts.NATSJWTFile,
		utils.NatsSeedFile:                         er.Opts.NATSSeedFile,
		utils.NatsCertificateAuthority:             er.Opts.NATSCertificateAuthority,
		utils.NatsClientCertificate:                er.Opts.NATSClientCertificate,
		utils.NatsClientKey:                        er.Opts.NATSClientKey,
		utils.NatsJetStreamMaxWait:                 er.Opts.NATSJetStreamMaxWait,
		utils.NATSJetStreamProcessedCfg:            er.Opts.NATSJetStreamProcessed,
		utils.NATSSubjectProcessedCfg:              er.Opts.NATSSubjectProcessed,
		utils.NATSJWTFileProcessedCfg:              er.Opts.NATSJWTFileProcessed,
		utils.NATSSeedFileProcessedCfg:             er.Opts.NATSSeedFileProcessed,
		utils.NATSCertificateAuthorityProcessedCfg: er.Opts.NATSCertificateAuthorityProcessed,
		utils.NATSClientCertificateProcessed:       er.Opts.NATSClientCertificateProcessed,
		utils.NATSClientKeyProcessedCfg:            er.Opts.NATSClientKeyProcessed,
		utils.NATSJetStreamMaxWaitProcessedCfg:     er.Opts.NATSJetStreamMaxWaitProcessed,
	}
	initialMP = map[string]interface{}{
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

	if flags := er.Flags.SliceFlags(); flags != nil {
		initialMP[utils.FlagsCfg] = flags
	}

	if er.Fields != nil {
		fields := make([]map[string]interface{}, len(er.Fields))
		for i, item := range er.Fields {
			fields[i] = item.AsMapInterface(separator)
		}
		initialMP[utils.FieldsCfg] = fields
	}
	if er.CacheDumpFields != nil {
		cacheDumpFields := make([]map[string]interface{}, len(er.CacheDumpFields))
		for i, item := range er.CacheDumpFields {
			cacheDumpFields[i] = item.AsMapInterface(separator)
		}
		initialMP[utils.CacheDumpFieldsCfg] = cacheDumpFields
	}
	if er.PartialCommitFields != nil {
		parCFields := make([]map[string]interface{}, len(er.PartialCommitFields))
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

type EventReaderOptsJson struct {
	PartialPath                       *string
	PartialCacheAction                *string
	PartialOrderField                 *string
	PartialCSVFieldSeparator          *string
	CSVRowLength                      *int
	CSVFieldSeparator                 *string
	CSVHeaderDefineChar               *string
	CSVLazyQuotes                     *bool
	XMLRootPath                       *string
	AMQPQueueID                       *string
	AMQPQueueIDProcessed              *string
	AMQPConsumerTag                   *string
	AMQPExchange                      *string
	AMQPExchangeType                  *string
	AMQPRoutingKey                    *string
	AMQPExchangeProcessed             *string
	AMQPExchangeTypeProcessed         *string
	AMQPRoutingKeyProcessed           *string
	KafkaTopic                        *string
	KafkaGroupID                      *string
	KafkaMaxWait                      *string
	KafkaTopicProcessed               *string
	SQLDBName                         *string
	SQLTableName                      *string
	SSLMode                           *string
	SQLDBNameProcessed                *string
	SQLTableNameProcessed             *string
	SSLModeProcessed                  *string
	AWSRegion                         *string
	AWSKey                            *string
	AWSSecret                         *string
	AWSToken                          *string
	AWSRegionProcessed                *string
	AWSKeyProcessed                   *string
	AWSSecretProcessed                *string
	AWSTokenProcessed                 *string
	SQSQueueID                        *string
	SQSQueueIDProcessed               *string
	S3BucketID                        *string
	S3FolderPathProcessed             *string
	S3BucketIDProcessed               *string
	NATSJetStream                     *bool
	NATSConsumerName                  *string
	NATSSubject                       *string
	NATSQueueID                       *string
	NATSJWTFile                       *string
	NATSSeedFile                      *string
	NATSCertificateAuthority          *string
	NATSClientCertificate             *string
	NATSClientKey                     *string
	NATSJetStreamMaxWait              *string
	NATSJetStreamProcessed            *bool
	NATSSubjectProcessed              *string
	NATSJWTFileProcessed              *string
	NATSSeedFileProcessed             *string
	NATSCertificateAuthorityProcessed *string
	NATSClientCertificateProcessed    *string
	NATSClientKeyProcessed            *string
	NATSJetStreamMaxWaitProcessed     *string
}

// EventReaderSJsonCfg is the configuration of a single EventReader
type EventReaderJsonCfg struct {
	Id                    *string
	Type                  *string
	Run_delay             *string
	Concurrent_requests   *int
	Source_path           *string
	Processed_path        *string
	Opts                  *EventReaderOptsJson
	Tenant                *string
	Timezone              *string
	Filters               *[]string
	Flags                 *[]string
	Fields                *[]*FcTemplateJsonCfg
	Partial_commit_fields *[]*FcTemplateJsonCfg
	Cache_dump_fields     *[]*FcTemplateJsonCfg
}

func diffEventReaderOptsJsonCfg(d *EventReaderOptsJson, v1, v2 *EventReaderOpts) *EventReaderOptsJson {
	if d == nil {
		d = new(EventReaderOptsJson)
	}
	if *v1.PartialPath != *v2.PartialPath {
		d.PartialPath = v2.PartialPath
	}
	if *v1.PartialCacheAction != *v2.PartialCacheAction {
		d.PartialCacheAction = v2.PartialCacheAction
	}
	if *v1.PartialOrderField != *v2.PartialOrderField {
		d.PartialOrderField = v2.PartialOrderField
	}
	if *v1.PartialCSVFieldSeparator != *v2.PartialCSVFieldSeparator {
		d.PartialCSVFieldSeparator = v2.PartialCSVFieldSeparator
	}
	if *v1.CSVRowLength != *v2.CSVRowLength {
		d.CSVRowLength = v2.CSVRowLength
	}
	if *v1.CSVFieldSeparator != *v2.CSVFieldSeparator {
		d.CSVFieldSeparator = v2.CSVFieldSeparator
	}
	if *v1.CSVHeaderDefineChar != *v2.CSVHeaderDefineChar {
		d.CSVHeaderDefineChar = v2.CSVHeaderDefineChar
	}
	if *v1.CSVLazyQuotes != *v2.CSVLazyQuotes {
		d.CSVLazyQuotes = v2.CSVLazyQuotes
	}
	if *v1.XMLRootPath != *v2.XMLRootPath {
		d.XMLRootPath = v2.XMLRootPath
	}
	if *v1.AMQPQueueID != *v2.AMQPQueueID {
		d.AMQPQueueID = v2.AMQPQueueID
	}
	if *v1.AMQPQueueIDProcessed != *v2.AMQPQueueIDProcessed {
		d.AMQPQueueIDProcessed = v2.AMQPQueueIDProcessed
	}
	if *v1.AMQPConsumerTag != *v2.AMQPConsumerTag {
		d.AMQPConsumerTag = v2.AMQPConsumerTag
	}
	if *v1.AMQPExchange != *v2.AMQPExchange {
		d.AMQPExchange = v2.AMQPExchange
	}
	if *v1.AMQPExchangeType != *v2.AMQPExchangeType {
		d.AMQPExchangeType = v2.AMQPExchangeType
	}
	if *v1.AMQPRoutingKey != *v2.AMQPRoutingKey {
		d.AMQPRoutingKey = v2.AMQPRoutingKey
	}
	if *v1.AMQPExchangeProcessed != *v2.AMQPExchangeProcessed {
		d.AMQPExchangeProcessed = v2.AMQPExchangeProcessed
	}
	if *v1.AMQPExchangeTypeProcessed != *v2.AMQPExchangeTypeProcessed {
		d.AMQPExchangeTypeProcessed = v2.AMQPExchangeTypeProcessed
	}
	if *v1.AMQPRoutingKeyProcessed != *v2.AMQPRoutingKeyProcessed {
		d.AMQPRoutingKeyProcessed = v2.AMQPRoutingKeyProcessed
	}
	if *v1.KafkaTopic != *v2.KafkaTopic {
		d.KafkaTopic = v2.KafkaTopic
	}
	if *v1.KafkaGroupID != *v2.KafkaGroupID {
		d.KafkaGroupID = v2.KafkaGroupID
	}
	if *v1.KafkaMaxWait != *v2.KafkaMaxWait {
		d.KafkaMaxWait = utils.StringPointer(v2.KafkaMaxWait.String())
	}
	if *v1.KafkaTopicProcessed != *v2.KafkaTopicProcessed {
		d.KafkaTopicProcessed = v2.KafkaTopicProcessed
	}
	if *v1.SQLDBName != *v2.SQLDBName {
		d.SQLDBName = v2.SQLDBName
	}
	if *v1.SQLTableName != *v2.SQLTableName {
		d.SQLTableName = v2.SQLTableName
	}
	if *v1.SSLMode != *v2.SSLMode {
		d.SSLMode = v2.SSLMode
	}
	if *v1.SQLDBNameProcessed != *v2.SQLDBNameProcessed {
		d.SQLDBNameProcessed = v2.SQLDBNameProcessed
	}
	if *v1.SQLTableNameProcessed != *v2.SQLTableNameProcessed {
		d.SQLTableNameProcessed = v2.SQLTableNameProcessed
	}
	if *v1.SSLModeProcessed != *v2.SSLModeProcessed {
		d.SSLModeProcessed = v2.SSLModeProcessed
	}
	if *v1.AWSRegion != *v2.AWSRegion {
		d.AWSRegion = v2.AWSRegion
	}
	if *v1.AWSKey != *v2.AWSKey {
		d.AWSKey = v2.AWSKey
	}
	if *v1.AWSSecret != *v2.AWSSecret {
		d.AWSSecret = v2.AWSSecret
	}
	if *v1.AWSToken != *v2.AWSToken {
		d.AWSToken = v2.AWSToken
	}
	if *v1.AWSRegionProcessed != *v2.AWSRegionProcessed {
		d.AWSRegionProcessed = v2.AWSRegionProcessed
	}
	if *v1.AWSKeyProcessed != *v2.AWSKeyProcessed {
		d.AWSKeyProcessed = v2.AWSKeyProcessed
	}
	if *v1.AWSSecretProcessed != *v2.AWSSecretProcessed {
		d.AWSSecretProcessed = v2.AWSSecretProcessed
	}
	if *v1.AWSTokenProcessed != *v2.AWSTokenProcessed {
		d.AWSTokenProcessed = v2.AWSTokenProcessed
	}
	if *v1.SQSQueueID != *v2.SQSQueueID {
		d.SQSQueueID = v2.SQSQueueID
	}
	if *v1.SQSQueueIDProcessed != *v2.SQSQueueIDProcessed {
		d.SQSQueueIDProcessed = v2.SQSQueueIDProcessed
	}
	if *v1.S3BucketID != *v2.S3BucketID {
		d.S3BucketID = v2.S3BucketID
	}
	if *v1.S3FolderPathProcessed != *v2.S3FolderPathProcessed {
		d.S3FolderPathProcessed = v2.S3FolderPathProcessed
	}
	if *v1.S3BucketIDProcessed != *v2.S3BucketIDProcessed {
		d.S3BucketIDProcessed = v2.S3BucketIDProcessed
	}
	if *v1.NATSJetStream != *v2.NATSJetStream {
		d.NATSJetStream = v2.NATSJetStream
	}
	if *v1.NATSConsumerName != *v2.NATSConsumerName {
		d.NATSConsumerName = v2.NATSConsumerName
	}
	if *v1.NATSSubject != *v2.NATSSubject {
		d.NATSSubject = v2.NATSSubject
	}
	if *v1.NATSQueueID != *v2.NATSQueueID {
		d.NATSQueueID = v2.NATSQueueID
	}
	if *v1.NATSJWTFile != *v2.NATSJWTFile {
		d.NATSJWTFile = v2.NATSJWTFile
	}
	if *v1.NATSSeedFile != *v2.NATSSeedFile {
		d.NATSSeedFile = v2.NATSSeedFile
	}
	if *v1.NATSCertificateAuthority != *v2.NATSCertificateAuthority {
		d.NATSCertificateAuthority = v2.NATSCertificateAuthority
	}
	if *v1.NATSClientCertificate != *v2.NATSClientCertificate {
		d.NATSClientCertificate = v2.NATSClientCertificate
	}
	if *v1.NATSClientKey != *v2.NATSClientKey {
		d.NATSClientKey = v2.NATSClientKey
	}
	if *v1.NATSJetStreamMaxWait != *v2.NATSJetStreamMaxWait {
		d.NATSJetStreamMaxWait = utils.StringPointer(v2.NATSJetStreamMaxWait.String())
	}
	if *v1.NATSJetStreamProcessed != *v2.NATSJetStreamProcessed {
		d.NATSJetStreamProcessed = v2.NATSJetStreamProcessed
	}
	if *v1.NATSSubjectProcessed != *v2.NATSSubjectProcessed {
		d.NATSSubjectProcessed = v2.NATSSubjectProcessed
	}
	if *v1.NATSJWTFileProcessed != *v2.NATSJWTFileProcessed {
		d.NATSJWTFileProcessed = v2.NATSJWTFileProcessed
	}
	if *v1.NATSSeedFileProcessed != *v2.NATSSeedFileProcessed {
		d.NATSSeedFileProcessed = v2.NATSSeedFileProcessed
	}
	if *v1.NATSCertificateAuthorityProcessed != *v2.NATSCertificateAuthorityProcessed {
		d.NATSCertificateAuthorityProcessed = v2.NATSCertificateAuthorityProcessed
	}
	if *v1.NATSClientCertificateProcessed != *v2.NATSClientCertificateProcessed {
		d.NATSClientCertificateProcessed = v2.NATSClientCertificateProcessed
	}
	if *v1.NATSClientKeyProcessed != *v2.NATSClientKeyProcessed {
		d.NATSClientKeyProcessed = v2.NATSClientKeyProcessed
	}
	if *v1.NATSJetStreamMaxWaitProcessed != *v2.NATSJetStreamMaxWaitProcessed {
		d.NATSJetStreamMaxWaitProcessed = utils.StringPointer(v2.NATSJetStreamMaxWaitProcessed.String())
	}
	return d
}

func diffEventReaderJsonCfg(d *EventReaderJsonCfg, v1, v2 *EventReaderCfg, separator string) *EventReaderJsonCfg {
	if d == nil {
		d = new(EventReaderJsonCfg)
	}
	if v1.ID != v2.ID {
		d.Id = utils.StringPointer(v2.ID)
	}
	if v1.Type != v2.Type {
		d.Type = utils.StringPointer(v2.Type)
	}
	if v1.RunDelay != v2.RunDelay {
		d.Run_delay = utils.StringPointer(v2.RunDelay.String())
	}
	if v1.ConcurrentReqs != v2.ConcurrentReqs {
		d.Concurrent_requests = utils.IntPointer(v2.ConcurrentReqs)
	}
	if v1.SourcePath != v2.SourcePath {
		d.Source_path = utils.StringPointer(v2.SourcePath)
	}
	if v1.ProcessedPath != v2.ProcessedPath {
		d.Processed_path = utils.StringPointer(v2.ProcessedPath)
	}
	d.Opts = diffEventReaderOptsJsonCfg(d.Opts, v1.Opts, v2.Opts)
	tnt1 := v1.Tenant.GetRule(separator)
	tnt2 := v2.Tenant.GetRule(separator)
	if tnt1 != tnt2 {
		d.Tenant = utils.StringPointer(tnt2)
	}
	if v1.Timezone != v2.Timezone {
		d.Timezone = utils.StringPointer(v2.Timezone)
	}
	if !utils.SliceStringEqual(v1.Filters, v2.Filters) {
		d.Filters = &v2.Filters
	}
	flgs1 := v1.Flags.SliceFlags()
	flgs2 := v2.Flags.SliceFlags()
	if !utils.SliceStringEqual(flgs1, flgs2) {
		d.Flags = &flgs2
	}
	var flds []*FcTemplateJsonCfg
	if d.Fields != nil {
		flds = *d.Fields
	}
	flds = diffFcTemplateJsonCfg(flds, v1.Fields, v2.Fields, separator)
	if flds != nil {
		d.Fields = &flds
	}

	var pcf []*FcTemplateJsonCfg
	if d.Partial_commit_fields != nil {
		pcf = *d.Partial_commit_fields
	}
	pcf = diffFcTemplateJsonCfg(pcf, v1.PartialCommitFields, v2.PartialCommitFields, separator)
	if pcf != nil {
		d.Partial_commit_fields = &pcf
	}

	var cdf []*FcTemplateJsonCfg
	if d.Cache_dump_fields != nil {
		cdf = *d.Cache_dump_fields
	}
	cdf = diffFcTemplateJsonCfg(cdf, v1.CacheDumpFields, v2.CacheDumpFields, separator)
	if cdf != nil {
		d.Cache_dump_fields = &cdf
	}

	return d
}

func getEventReaderJsonCfg(d []*EventReaderJsonCfg, id string) (*EventReaderJsonCfg, int) {
	for i, v := range d {
		if v.Id != nil && *v.Id == id {
			return v, i
		}
	}
	return nil, -1
}

func getEventReaderCfg(d []*EventReaderCfg, id string) *EventReaderCfg {
	for _, v := range d {
		if v.ID == id {
			return v
		}
	}
	return new(EventReaderCfg)
}

func diffEventReadersJsonCfg(d *[]*EventReaderJsonCfg, v1, v2 []*EventReaderCfg, separator string) *[]*EventReaderJsonCfg {
	if d == nil || *d == nil {
		d = &[]*EventReaderJsonCfg{}
	}
	for _, val := range v2 {
		dv, i := getEventReaderJsonCfg(*d, val.ID)
		dv = diffEventReaderJsonCfg(dv, getEventReaderCfg(v1, val.ID), val, separator)
		if i == -1 {
			*d = append(*d, dv)
		} else {
			(*d)[i] = dv
		}
	}

	return d
}

// EventReaderSJsonCfg contains the configuration of EventReaderService
type ERsJsonCfg struct {
	Enabled           *bool
	Sessions_conns    *[]string
	Readers           *[]*EventReaderJsonCfg
	Partial_cache_ttl *string
}

func diffERsJsonCfg(d *ERsJsonCfg, v1, v2 *ERsCfg, separator string) *ERsJsonCfg {
	if d == nil {
		d = new(ERsJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if !utils.SliceStringEqual(v1.SessionSConns, v2.SessionSConns) {
		d.Sessions_conns = utils.SliceStringPointer(getInternalJSONConns(v2.SessionSConns))
	}
	if v1.PartialCacheTTL != v2.PartialCacheTTL {
		d.Partial_cache_ttl = utils.StringPointer(v2.PartialCacheTTL.String())
	}
	d.Readers = diffEventReadersJsonCfg(d.Readers, v1.Readers, v2.Readers, separator)
	return d
}
