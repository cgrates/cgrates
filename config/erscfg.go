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
		PartialPath:                       utils.StringPointer(*erOpts.PartialPath),
		PartialCacheAction:                utils.StringPointer(*erOpts.PartialCacheAction),
		PartialOrderField:                 utils.StringPointer(*erOpts.PartialOrderField),
		PartialCSVFieldSeparator:          utils.StringPointer(*erOpts.PartialCSVFieldSeparator),
		CSVRowLength:                      utils.IntPointer(*erOpts.CSVRowLength),
		CSVFieldSeparator:                 utils.StringPointer(*erOpts.CSVFieldSeparator),
		CSVHeaderDefineChar:               utils.StringPointer(*erOpts.CSVHeaderDefineChar),
		CSVLazyQuotes:                     utils.BoolPointer(*erOpts.CSVLazyQuotes),
		XMLRootPath:                       utils.StringPointer(*erOpts.XMLRootPath),
		AMQPQueueID:                       utils.StringPointer(*erOpts.AMQPQueueID),
		AMQPQueueIDProcessed:              utils.StringPointer(*erOpts.AMQPQueueIDProcessed),
		AMQPConsumerTag:                   utils.StringPointer(*erOpts.AMQPConsumerTag),
		AMQPExchange:                      utils.StringPointer(*erOpts.AMQPExchange),
		AMQPExchangeType:                  utils.StringPointer(*erOpts.AMQPExchangeType),
		AMQPRoutingKey:                    utils.StringPointer(*erOpts.AMQPRoutingKey),
		AMQPExchangeProcessed:             utils.StringPointer(*erOpts.AMQPExchangeProcessed),
		AMQPExchangeTypeProcessed:         utils.StringPointer(*erOpts.AMQPExchangeTypeProcessed),
		AMQPRoutingKeyProcessed:           utils.StringPointer(*erOpts.AMQPRoutingKeyProcessed),
		KafkaTopic:                        utils.StringPointer(*erOpts.KafkaTopic),
		KafkaGroupID:                      utils.StringPointer(*erOpts.KafkaGroupID),
		KafkaMaxWait:                      utils.DurationPointer(*erOpts.KafkaMaxWait),
		KafkaTopicProcessed:               utils.StringPointer(*erOpts.KafkaTopicProcessed),
		SQLDBName:                         utils.StringPointer(*erOpts.SQLDBName),
		SQLTableName:                      utils.StringPointer(*erOpts.SQLTableName),
		SSLMode:                           utils.StringPointer(*erOpts.SSLMode),
		SQLDBNameProcessed:                utils.StringPointer(*erOpts.SQLDBNameProcessed),
		SQLTableNameProcessed:             utils.StringPointer(*erOpts.SQLTableNameProcessed),
		SSLModeProcessed:                  utils.StringPointer(*erOpts.SSLModeProcessed),
		AWSRegion:                         utils.StringPointer(*erOpts.AWSRegion),
		AWSKey:                            utils.StringPointer(*erOpts.AWSKey),
		AWSSecret:                         utils.StringPointer(*erOpts.AWSSecret),
		AWSToken:                          utils.StringPointer(*erOpts.AWSToken),
		AWSRegionProcessed:                utils.StringPointer(*erOpts.AWSRegionProcessed),
		AWSKeyProcessed:                   utils.StringPointer(*erOpts.AWSKeyProcessed),
		AWSSecretProcessed:                utils.StringPointer(*erOpts.AWSSecretProcessed),
		AWSTokenProcessed:                 utils.StringPointer(*erOpts.AWSTokenProcessed),
		SQSQueueID:                        utils.StringPointer(*erOpts.SQSQueueID),
		SQSQueueIDProcessed:               utils.StringPointer(*erOpts.SQSQueueIDProcessed),
		S3BucketID:                        utils.StringPointer(*erOpts.S3BucketID),
		S3FolderPathProcessed:             utils.StringPointer(*erOpts.S3FolderPathProcessed),
		S3BucketIDProcessed:               utils.StringPointer(*erOpts.S3BucketIDProcessed),
		NATSJetStream:                     utils.BoolPointer(*erOpts.NATSJetStream),
		NATSConsumerName:                  utils.StringPointer(*erOpts.NATSConsumerName),
		NATSSubject:                       utils.StringPointer(*erOpts.NATSSubject),
		NATSQueueID:                       utils.StringPointer(*erOpts.NATSQueueID),
		NATSJWTFile:                       utils.StringPointer(*erOpts.NATSJWTFile),
		NATSSeedFile:                      utils.StringPointer(*erOpts.NATSSeedFile),
		NATSCertificateAuthority:          utils.StringPointer(*erOpts.NATSCertificateAuthority),
		NATSClientCertificate:             utils.StringPointer(*erOpts.NATSClientCertificate),
		NATSClientKey:                     utils.StringPointer(*erOpts.NATSClientKey),
		NATSJetStreamMaxWait:              utils.DurationPointer(*erOpts.NATSJetStreamMaxWait),
		NATSJetStreamProcessed:            utils.BoolPointer(*erOpts.NATSJetStreamProcessed),
		NATSSubjectProcessed:              utils.StringPointer(*erOpts.NATSSubjectProcessed),
		NATSJWTFileProcessed:              utils.StringPointer(*erOpts.NATSJWTFileProcessed),
		NATSSeedFileProcessed:             utils.StringPointer(*erOpts.NATSSeedFileProcessed),
		NATSCertificateAuthorityProcessed: utils.StringPointer(*erOpts.NATSCertificateAuthorityProcessed),
		NATSClientCertificateProcessed:    utils.StringPointer(*erOpts.NATSClientCertificateProcessed),
		NATSClientKeyProcessed:            utils.StringPointer(*erOpts.NATSClientKeyProcessed),
		NATSJetStreamMaxWaitProcessed:     utils.DurationPointer(*erOpts.NATSJetStreamMaxWaitProcessed),
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
	opts := map[string]interface{}{}

	if er.Opts.PartialPath != nil {
		opts[utils.PartialPathOpt] = *er.Opts.PartialPath
	}
	if er.Opts.PartialCacheAction != nil {
		opts[utils.PartialCacheActionOpt] = *er.Opts.PartialCacheAction
	}
	if er.Opts.PartialOrderField != nil {
		opts[utils.PartialOrderFieldOpt] = *er.Opts.PartialOrderField
	}
	if er.Opts.PartialCSVFieldSeparator != nil {
		opts[utils.PartialCSVFieldSepartorOpt] = *er.Opts.PartialCSVFieldSeparator
	}
	if er.Opts.CSVRowLength != nil {
		opts[utils.CSVRowLengthOpt] = *er.Opts.CSVRowLength
	}
	if er.Opts.CSVFieldSeparator != nil {
		opts[utils.CSVFieldSepOpt] = *er.Opts.CSVFieldSeparator
	}
	if er.Opts.CSVHeaderDefineChar != nil {
		opts[utils.HeaderDefineCharOpt] = *er.Opts.CSVHeaderDefineChar
	}
	if er.Opts.CSVLazyQuotes != nil {
		opts[utils.CSVLazyQuotes] = *er.Opts.CSVLazyQuotes
	}
	if er.Opts.XMLRootPath != nil {
		opts[utils.XMLRootPathOpt] = *er.Opts.XMLRootPath
	}
	if er.Opts.AMQPQueueID != nil {
		opts[utils.AMQPQueueID] = *er.Opts.AMQPQueueID
	}
	if er.Opts.AMQPQueueIDProcessed != nil {
		opts[utils.AMQPQueueIDProcessedCfg] = *er.Opts.AMQPQueueIDProcessed
	}
	if er.Opts.AMQPConsumerTag != nil {
		opts[utils.AMQPConsumerTag] = *er.Opts.AMQPConsumerTag
	}
	if er.Opts.AMQPExchange != nil {
		opts[utils.AMQPExchange] = *er.Opts.AMQPExchange
	}
	if er.Opts.AMQPExchangeType != nil {
		opts[utils.AMQPExchangeType] = *er.Opts.AMQPExchangeType
	}
	if er.Opts.AMQPRoutingKey != nil {
		opts[utils.AMQPRoutingKey] = *er.Opts.AMQPRoutingKey
	}
	if er.Opts.AMQPExchangeProcessed != nil {
		opts[utils.AMQPExchangeProcessedCfg] = *er.Opts.AMQPExchangeProcessed
	}
	if er.Opts.AMQPExchangeTypeProcessed != nil {
		opts[utils.AMQPExchangeTypeProcessedCfg] = *er.Opts.AMQPExchangeTypeProcessed
	}
	if er.Opts.AMQPRoutingKeyProcessed != nil {
		opts[utils.AMQPRoutingKeyProcessedCfg] = *er.Opts.AMQPRoutingKeyProcessed
	}
	if er.Opts.KafkaTopic != nil {
		opts[utils.KafkaTopic] = *er.Opts.KafkaTopic
	}
	if er.Opts.KafkaGroupID != nil {
		opts[utils.KafkaGroupID] = *er.Opts.KafkaGroupID
	}
	if er.Opts.KafkaMaxWait != nil {
		opts[utils.KafkaMaxWait] = er.Opts.KafkaMaxWait.String()
	}
	if er.Opts.KafkaTopicProcessed != nil {
		opts[utils.KafkaTopicProcessedCfg] = *er.Opts.KafkaTopicProcessed
	}
	if er.Opts.SQLDBName != nil {
		opts[utils.SQLDBNameOpt] = *er.Opts.SQLDBName
	}
	if er.Opts.SQLTableName != nil {
		opts[utils.SQLTableNameOpt] = *er.Opts.SQLTableName
	}
	if er.Opts.SSLMode != nil {
		opts[utils.SSLModeCfg] = *er.Opts.SSLMode
	}
	if er.Opts.SQLDBNameProcessed != nil {
		opts[utils.SQLDBNameProcessedCfg] = *er.Opts.SQLDBNameProcessed
	}
	if er.Opts.SQLTableNameProcessed != nil {
		opts[utils.SQLTableNameProcessedCfg] = *er.Opts.SQLTableNameProcessed
	}
	if er.Opts.SSLModeProcessed != nil {
		opts[utils.SSLModeProcessedCfg] = *er.Opts.SSLModeProcessed
	}
	if er.Opts.AWSRegion != nil {
		opts[utils.AWSRegion] = *er.Opts.AWSRegion
	}
	if er.Opts.AWSKey != nil {
		opts[utils.AWSKey] = *er.Opts.AWSKey
	}
	if er.Opts.AWSSecret != nil {
		opts[utils.AWSSecret] = *er.Opts.AWSSecret
	}
	if er.Opts.AWSToken != nil {
		opts[utils.AWSToken] = *er.Opts.AWSToken
	}
	if er.Opts.AWSRegionProcessed != nil {
		opts[utils.AWSRegionProcessedCfg] = *er.Opts.AWSRegionProcessed
	}
	if er.Opts.AWSKeyProcessed != nil {
		opts[utils.AWSKeyProcessedCfg] = *er.Opts.AWSKeyProcessed
	}
	if er.Opts.AWSSecretProcessed != nil {
		opts[utils.AWSSecretProcessedCfg] = *er.Opts.AWSSecretProcessed
	}
	if er.Opts.AWSTokenProcessed != nil {
		opts[utils.AWSTokenProcessedCfg] = *er.Opts.AWSTokenProcessed
	}
	if er.Opts.SQSQueueID != nil {
		opts[utils.SQSQueueID] = *er.Opts.SQSQueueID
	}
	if er.Opts.SQSQueueIDProcessed != nil {
		opts[utils.SQSQueueIDProcessedCfg] = *er.Opts.SQSQueueIDProcessed
	}
	if er.Opts.S3BucketID != nil {
		opts[utils.S3Bucket] = *er.Opts.S3BucketID
	}
	if er.Opts.S3FolderPathProcessed != nil {
		opts[utils.S3FolderPathProcessedCfg] = *er.Opts.S3FolderPathProcessed
	}
	if er.Opts.S3BucketIDProcessed != nil {
		opts[utils.S3BucketIDProcessedCfg] = *er.Opts.S3BucketIDProcessed
	}
	if er.Opts.NATSJetStream != nil {
		opts[utils.NatsJetStream] = *er.Opts.NATSJetStream
	}
	if er.Opts.NATSConsumerName != nil {
		opts[utils.NatsConsumerName] = *er.Opts.NATSConsumerName
	}
	if er.Opts.NATSSubject != nil {
		opts[utils.NatsSubject] = *er.Opts.NATSSubject
	}
	if er.Opts.NATSQueueID != nil {
		opts[utils.NatsQueueID] = *er.Opts.NATSQueueID
	}
	if er.Opts.NATSJWTFile != nil {
		opts[utils.NatsJWTFile] = *er.Opts.NATSJWTFile
	}
	if er.Opts.NATSSeedFile != nil {
		opts[utils.NatsSeedFile] = *er.Opts.NATSSeedFile
	}
	if er.Opts.NATSCertificateAuthority != nil {
		opts[utils.NatsCertificateAuthority] = *er.Opts.NATSCertificateAuthority
	}
	if er.Opts.NATSClientCertificate != nil {
		opts[utils.NatsClientCertificate] = *er.Opts.NATSClientCertificate
	}
	if er.Opts.NATSClientKey != nil {
		opts[utils.NatsClientKey] = *er.Opts.NATSClientKey
	}
	if er.Opts.NATSJetStreamMaxWait != nil {
		opts[utils.NatsJetStreamMaxWait] = er.Opts.NATSJetStreamMaxWait.String()
	}
	if er.Opts.NATSJetStreamProcessed != nil {
		opts[utils.NATSJetStreamProcessedCfg] = *er.Opts.NATSJetStreamProcessed
	}
	if er.Opts.NATSSubjectProcessed != nil {
		opts[utils.NATSSubjectProcessedCfg] = *er.Opts.NATSSubjectProcessed
	}
	if er.Opts.NATSJWTFileProcessed != nil {
		opts[utils.NATSJWTFileProcessedCfg] = *er.Opts.NATSJWTFileProcessed
	}
	if er.Opts.NATSSeedFileProcessed != nil {
		opts[utils.NATSSeedFileProcessedCfg] = *er.Opts.NATSSeedFileProcessed
	}
	if er.Opts.NATSCertificateAuthorityProcessed != nil {
		opts[utils.NATSCertificateAuthorityProcessedCfg] = *er.Opts.NATSCertificateAuthorityProcessed
	}
	if er.Opts.NATSClientCertificateProcessed != nil {
		opts[utils.NATSClientCertificateProcessed] = *er.Opts.NATSClientCertificateProcessed
	}
	if er.Opts.NATSClientKeyProcessed != nil {
		opts[utils.NATSClientKeyProcessedCfg] = *er.Opts.NATSClientKeyProcessed
	}
	if er.Opts.NATSJetStreamMaxWaitProcessed != nil {
		opts[utils.NATSJetStreamMaxWaitProcessedCfg] = er.Opts.NATSJetStreamMaxWaitProcessed.String()
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
	if v2.PartialPath != nil {
		if v1.PartialPath == nil ||
			*v1.PartialPath != *v2.PartialPath {
			d.PartialPath = v2.PartialPath
		}
	} else {
		d.PartialPath = nil
	}
	if v2.PartialCacheAction != nil {
		if v1.PartialCacheAction == nil ||
			*v1.PartialCacheAction != *v2.PartialCacheAction {
			d.PartialCacheAction = v2.PartialCacheAction
		}
	} else {
		d.PartialCacheAction = nil
	}
	if v2.PartialOrderField != nil {
		if v1.PartialOrderField == nil ||
			*v1.PartialOrderField != *v2.PartialOrderField {
			d.PartialOrderField = v2.PartialOrderField
		}
	} else {
		d.PartialOrderField = nil
	}
	if v2.PartialCSVFieldSeparator != nil {
		if v1.PartialCSVFieldSeparator == nil ||
			*v1.PartialCSVFieldSeparator != *v2.PartialCSVFieldSeparator {
			d.PartialCSVFieldSeparator = v2.PartialCSVFieldSeparator
		}
	} else {
		d.PartialCSVFieldSeparator = nil
	}
	if v2.CSVRowLength != nil {
		if v1.CSVRowLength == nil ||
			*v1.CSVRowLength != *v2.CSVRowLength {
			d.CSVRowLength = v2.CSVRowLength
		}
	} else {
		d.CSVRowLength = nil
	}
	if v2.CSVFieldSeparator != nil {
		if v1.CSVFieldSeparator == nil ||
			*v1.CSVFieldSeparator != *v2.CSVFieldSeparator {
			d.CSVFieldSeparator = v2.CSVFieldSeparator
		}
	} else {
		d.CSVFieldSeparator = nil
	}
	if v2.CSVHeaderDefineChar != nil {
		if v1.CSVHeaderDefineChar == nil ||
			*v1.CSVHeaderDefineChar != *v2.CSVHeaderDefineChar {
			d.CSVHeaderDefineChar = v2.CSVHeaderDefineChar
		}
	} else {
		d.CSVHeaderDefineChar = nil
	}
	if v2.CSVLazyQuotes != nil {
		if v1.CSVLazyQuotes == nil ||
			*v1.CSVLazyQuotes != *v2.CSVLazyQuotes {
			d.CSVLazyQuotes = v2.CSVLazyQuotes
		}
	} else {
		d.CSVLazyQuotes = nil
	}
	if v2.XMLRootPath != nil {
		if v1.XMLRootPath == nil ||
			*v1.XMLRootPath != *v2.XMLRootPath {
			d.XMLRootPath = v2.XMLRootPath
		}
	} else {
		d.XMLRootPath = nil
	}
	if v2.AMQPQueueID != nil {
		if v1.AMQPQueueID == nil ||
			*v1.AMQPQueueID != *v2.AMQPQueueID {
			d.AMQPQueueID = v2.AMQPQueueID
		}
	} else {
		d.AMQPQueueID = nil
	}
	if v2.AMQPQueueIDProcessed != nil {
		if v1.AMQPQueueIDProcessed == nil ||
			*v1.AMQPQueueIDProcessed != *v2.AMQPQueueIDProcessed {
			d.AMQPQueueIDProcessed = v2.AMQPQueueIDProcessed
		}
	} else {
		d.AMQPQueueIDProcessed = nil
	}
	if v2.AMQPConsumerTag != nil {
		if v1.AMQPConsumerTag == nil ||
			*v1.AMQPConsumerTag != *v2.AMQPConsumerTag {
			d.AMQPConsumerTag = v2.AMQPConsumerTag
		}
	} else {
		d.AMQPConsumerTag = nil
	}
	if v2.AMQPExchange != nil {
		if v1.AMQPExchange == nil ||
			*v1.AMQPExchange != *v2.AMQPExchange {
			d.AMQPExchange = v2.AMQPExchange
		}
	} else {
		d.AMQPExchange = nil
	}
	if v2.AMQPExchangeType != nil {
		if v1.AMQPExchangeType == nil ||
			*v1.AMQPExchangeType != *v2.AMQPExchangeType {
			d.AMQPExchangeType = v2.AMQPExchangeType
		}
	} else {
		d.AMQPExchangeType = nil
	}
	if v2.AMQPRoutingKey != nil {
		if v1.AMQPRoutingKey == nil ||
			*v1.AMQPRoutingKey != *v2.AMQPRoutingKey {
			d.AMQPRoutingKey = v2.AMQPRoutingKey
		}
	} else {
		d.AMQPRoutingKey = nil
	}
	if v2.AMQPExchangeProcessed != nil {
		if v1.AMQPExchangeProcessed == nil ||
			*v1.AMQPExchangeProcessed != *v2.AMQPExchangeProcessed {
			d.AMQPExchangeProcessed = v2.AMQPExchangeProcessed
		}
	} else {
		d.AMQPExchangeProcessed = nil
	}
	if v2.AMQPExchangeTypeProcessed != nil {
		if v1.AMQPExchangeTypeProcessed == nil ||
			*v1.AMQPExchangeTypeProcessed != *v2.AMQPExchangeTypeProcessed {
			d.AMQPExchangeTypeProcessed = v2.AMQPExchangeTypeProcessed
		}
	} else {
		d.AMQPExchangeTypeProcessed = nil
	}
	if v2.AMQPRoutingKeyProcessed != nil {
		if v1.AMQPRoutingKeyProcessed == nil ||
			*v1.AMQPRoutingKeyProcessed != *v2.AMQPRoutingKeyProcessed {
			d.AMQPRoutingKeyProcessed = v2.AMQPRoutingKeyProcessed
		}
	} else {
		d.AMQPRoutingKeyProcessed = nil
	}
	if v2.KafkaTopic != nil {
		if v1.KafkaTopic == nil ||
			*v1.KafkaTopic != *v2.KafkaTopic {
			d.KafkaTopic = v2.KafkaTopic
		}
	} else {
		d.KafkaTopic = nil
	}
	if v2.KafkaGroupID != nil {
		if v1.KafkaGroupID == nil ||
			*v1.KafkaGroupID != *v2.KafkaGroupID {
			d.KafkaGroupID = v2.KafkaGroupID
		}
	} else {
		d.KafkaGroupID = nil
	}
	if v2.KafkaMaxWait != nil {
		if v1.KafkaMaxWait == nil ||
			*v1.KafkaMaxWait != *v2.KafkaMaxWait {
			d.KafkaMaxWait = utils.StringPointer(v2.KafkaMaxWait.String())
		}
	} else {
		d.KafkaMaxWait = nil
	}
	if v2.KafkaTopicProcessed != nil {
		if v1.KafkaTopicProcessed == nil ||
			*v1.KafkaTopicProcessed != *v2.KafkaTopicProcessed {
			d.KafkaTopicProcessed = v2.KafkaTopicProcessed
		}
	} else {
		d.KafkaTopicProcessed = nil
	}
	if v2.SQLDBName != nil {
		if v1.SQLDBName == nil ||
			*v1.SQLDBName != *v2.SQLDBName {
			d.SQLDBName = v2.SQLDBName
		}
	} else {
		d.SQLDBName = nil
	}
	if v2.SQLTableName != nil {
		if v1.SQLTableName == nil ||
			*v1.SQLTableName != *v2.SQLTableName {
			d.SQLTableName = v2.SQLTableName
		}
	} else {
		d.SQLTableName = nil
	}
	if v2.SSLMode != nil {
		if v1.SSLMode == nil ||
			*v1.SSLMode != *v2.SSLMode {
			d.SSLMode = v2.SSLMode
		}
	} else {
		d.SSLMode = nil
	}
	if v2.SQLDBNameProcessed != nil {
		if v1.SQLDBNameProcessed == nil ||
			*v1.SQLDBNameProcessed != *v2.SQLDBNameProcessed {
			d.SQLDBNameProcessed = v2.SQLDBNameProcessed
		}
	} else {
		d.SQLDBNameProcessed = nil
	}
	if v2.SQLTableNameProcessed != nil {
		if v1.SQLTableNameProcessed == nil ||
			*v1.SQLTableNameProcessed != *v2.SQLTableNameProcessed {
			d.SQLTableNameProcessed = v2.SQLTableNameProcessed
		}
	} else {
		d.SQLTableNameProcessed = nil
	}
	if v2.SSLModeProcessed != nil {
		if v1.SSLModeProcessed == nil ||
			*v1.SSLModeProcessed != *v2.SSLModeProcessed {
			d.SSLModeProcessed = v2.SSLModeProcessed
		}
	} else {
		d.SSLModeProcessed = nil
	}
	if v2.AWSRegion != nil {
		if v1.AWSRegion == nil ||
			*v1.AWSRegion != *v2.AWSRegion {
			d.AWSRegion = v2.AWSRegion
		}
	} else {
		d.AWSRegion = nil
	}
	if v2.AWSKey != nil {
		if v1.AWSKey == nil ||
			*v1.AWSKey != *v2.AWSKey {
			d.AWSKey = v2.AWSKey
		}
	} else {
		d.AWSKey = nil
	}
	if v2.AWSSecret != nil {
		if v1.AWSSecret == nil ||
			*v1.AWSSecret != *v2.AWSSecret {
			d.AWSSecret = v2.AWSSecret
		}
	} else {
		d.AWSSecret = nil
	}
	if v2.AWSToken != nil {
		if v1.AWSToken == nil ||
			*v1.AWSToken != *v2.AWSToken {
			d.AWSToken = v2.AWSToken
		}
	} else {
		d.AWSToken = nil
	}
	if v2.AWSRegionProcessed != nil {
		if v1.AWSRegionProcessed == nil ||
			*v1.AWSRegionProcessed != *v2.AWSRegionProcessed {
			d.AWSRegionProcessed = v2.AWSRegionProcessed
		}
	} else {
		d.AWSRegionProcessed = nil
	}
	if v2.AWSKeyProcessed != nil {
		if v1.AWSKeyProcessed == nil ||
			*v1.AWSKeyProcessed != *v2.AWSKeyProcessed {
			d.AWSKeyProcessed = v2.AWSKeyProcessed
		}
	} else {
		d.AWSKeyProcessed = nil
	}
	if v2.AWSSecretProcessed != nil {
		if v1.AWSSecretProcessed == nil ||
			*v1.AWSSecretProcessed != *v2.AWSSecretProcessed {
			d.AWSSecretProcessed = v2.AWSSecretProcessed
		}
	} else {
		d.AWSSecretProcessed = nil
	}
	if v2.AWSTokenProcessed != nil {
		if v1.AWSTokenProcessed == nil ||
			*v1.AWSTokenProcessed != *v2.AWSTokenProcessed {
			d.AWSTokenProcessed = v2.AWSTokenProcessed
		}
	} else {
		d.AWSTokenProcessed = nil
	}
	if v2.SQSQueueID != nil {
		if v1.SQSQueueID == nil ||
			*v1.SQSQueueID != *v2.SQSQueueID {
			d.SQSQueueID = v2.SQSQueueID
		}
	} else {
		d.SQSQueueID = nil
	}
	if v2.SQSQueueIDProcessed != nil {
		if v1.SQSQueueIDProcessed == nil ||
			*v1.SQSQueueIDProcessed != *v2.SQSQueueIDProcessed {
			d.SQSQueueIDProcessed = v2.SQSQueueIDProcessed
		}
	} else {
		d.SQSQueueIDProcessed = nil
	}
	if v2.S3BucketID != nil {
		if v1.S3BucketID == nil ||
			*v1.S3BucketID != *v2.S3BucketID {
			d.S3BucketID = v2.S3BucketID
		}
	} else {
		d.S3BucketID = nil
	}
	if v2.S3FolderPathProcessed != nil {
		if v1.S3FolderPathProcessed == nil ||
			*v1.S3FolderPathProcessed != *v2.S3FolderPathProcessed {
			d.S3FolderPathProcessed = v2.S3FolderPathProcessed
		}
	} else {
		d.S3FolderPathProcessed = nil
	}
	if v2.S3BucketIDProcessed != nil {
		if v1.S3BucketIDProcessed == nil ||
			*v1.S3BucketIDProcessed != *v2.S3BucketIDProcessed {
			d.S3BucketIDProcessed = v2.S3BucketIDProcessed
		}
	} else {
		d.S3BucketIDProcessed = nil
	}
	if v2.NATSJetStream != nil {
		if v1.NATSJetStream == nil ||
			*v1.NATSJetStream != *v2.NATSJetStream {
			d.NATSJetStream = v2.NATSJetStream
		}
	} else {
		d.NATSJetStream = nil
	}
	if v2.NATSConsumerName != nil {
		if v1.NATSConsumerName == nil ||
			*v1.NATSConsumerName != *v2.NATSConsumerName {
			d.NATSConsumerName = v2.NATSConsumerName
		}
	} else {
		d.NATSConsumerName = nil
	}
	if v2.NATSSubject != nil {
		if v1.NATSSubject == nil ||
			*v1.NATSSubject != *v2.NATSSubject {
			d.NATSSubject = v2.NATSSubject
		}
	} else {
		d.NATSSubject = nil
	}
	if v2.NATSQueueID != nil {
		if v1.NATSQueueID == nil ||
			*v1.NATSQueueID != *v2.NATSQueueID {
			d.NATSQueueID = v2.NATSQueueID
		}
	} else {
		d.NATSQueueID = nil
	}
	if v2.NATSJWTFile != nil {
		if v1.NATSJWTFile == nil ||
			*v1.NATSJWTFile != *v2.NATSJWTFile {
			d.NATSJWTFile = v2.NATSJWTFile
		}
	} else {
		d.NATSJWTFile = nil
	}
	if v2.NATSSeedFile != nil {
		if v1.NATSSeedFile == nil ||
			*v1.NATSSeedFile != *v2.NATSSeedFile {
			d.NATSSeedFile = v2.NATSSeedFile
		}
	} else {
		d.NATSSeedFile = nil
	}
	if v2.NATSCertificateAuthority != nil {
		if v1.NATSCertificateAuthority == nil ||
			*v1.NATSCertificateAuthority != *v2.NATSCertificateAuthority {
			d.NATSCertificateAuthority = v2.NATSCertificateAuthority
		}
	} else {
		d.NATSCertificateAuthority = nil
	}
	if v2.NATSClientCertificate != nil {
		if v1.NATSClientCertificate == nil ||
			*v1.NATSClientCertificate != *v2.NATSClientCertificate {
			d.NATSClientCertificate = v2.NATSClientCertificate
		}
	} else {
		d.NATSClientCertificate = nil
	}
	if v2.NATSClientKey != nil {
		if v1.NATSClientKey == nil ||
			*v1.NATSClientKey != *v2.NATSClientKey {
			d.NATSClientKey = v2.NATSClientKey
		}
	} else {
		d.NATSClientKey = nil
	}
	if v2.NATSJetStreamMaxWait != nil {
		if v1.NATSJetStreamMaxWait == nil ||
			*v1.NATSJetStreamMaxWait != *v2.NATSJetStreamMaxWait {
			d.NATSJetStreamMaxWait = utils.StringPointer(v2.NATSJetStreamMaxWait.String())
		}
	} else {
		d.NATSJetStreamMaxWait = nil
	}
	if v2.NATSJetStreamProcessed != nil {
		if v1.NATSJetStreamProcessed == nil ||
			*v1.NATSJetStreamProcessed != *v2.NATSJetStreamProcessed {
			d.NATSJetStreamProcessed = v2.NATSJetStreamProcessed
		}
	} else {
		d.NATSJetStreamProcessed = nil
	}
	if v2.NATSSubjectProcessed != nil {
		if v1.NATSSubjectProcessed == nil ||
			*v1.NATSSubjectProcessed != *v2.NATSSubjectProcessed {
			d.NATSSubjectProcessed = v2.NATSSubjectProcessed
		}
	} else {
		d.NATSSubjectProcessed = nil
	}
	if v2.NATSJWTFileProcessed != nil {
		if v1.NATSJWTFileProcessed == nil ||
			*v1.NATSJWTFileProcessed != *v2.NATSJWTFileProcessed {
			d.NATSJWTFileProcessed = v2.NATSJWTFileProcessed
		}
	} else {
		d.NATSJWTFileProcessed = nil
	}
	if v2.NATSSeedFileProcessed != nil {
		if v1.NATSSeedFileProcessed == nil ||
			*v1.NATSSeedFileProcessed != *v2.NATSSeedFileProcessed {
			d.NATSSeedFileProcessed = v2.NATSSeedFileProcessed
		}
	} else {
		d.NATSSeedFileProcessed = nil
	}
	if v2.NATSCertificateAuthorityProcessed != nil {
		if v1.NATSCertificateAuthorityProcessed == nil ||
			*v1.NATSCertificateAuthorityProcessed != *v2.NATSCertificateAuthorityProcessed {
			d.NATSCertificateAuthorityProcessed = v2.NATSCertificateAuthorityProcessed
		}
	} else {
		d.NATSCertificateAuthorityProcessed = nil
	}
	if v2.NATSClientCertificateProcessed != nil {
		if v1.NATSClientCertificateProcessed == nil ||
			*v1.NATSClientCertificateProcessed != *v2.NATSClientCertificateProcessed {
			d.NATSClientCertificateProcessed = v2.NATSClientCertificateProcessed
		}
	} else {
		d.NATSClientCertificateProcessed = nil
	}
	if v2.NATSClientKeyProcessed != nil {
		if v1.NATSClientKeyProcessed == nil ||
			*v1.NATSClientKeyProcessed != *v2.NATSClientKeyProcessed {
			d.NATSClientKeyProcessed = v2.NATSClientKeyProcessed
		}
	} else {
		d.NATSClientKeyProcessed = nil
	}
	if v2.NATSJetStreamMaxWaitProcessed != nil {
		if v1.NATSJetStreamMaxWaitProcessed == nil ||
			*v1.NATSJetStreamMaxWaitProcessed != *v2.NATSJetStreamMaxWaitProcessed {
			d.NATSJetStreamMaxWaitProcessed = utils.StringPointer(v2.NATSJetStreamMaxWaitProcessed.String())
		}
	} else {
		d.NATSJetStreamMaxWaitProcessed = nil
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
	return &EventReaderCfg{
		Opts: &EventReaderOpts{},
	}
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
