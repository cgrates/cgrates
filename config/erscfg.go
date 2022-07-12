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

// AsMapInterface returns the config as a map[string]interface{}
func (erS *ERsCfg) AsMapInterface(separator string) (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
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
		readers := make([]map[string]interface{}, len(erS.Readers))
		for i, item := range erS.Readers {
			readers[i] = item.AsMapInterface(separator)
		}
		initialMP[utils.ReadersCfg] = readers
	}
	return
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
	PgSSLMode                         *string
	SQLDBNameProcessed                *string
	SQLTableNameProcessed             *string
	PgSSLModeProcessed                *string
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
	if jsnCfg.PgSSLMode != nil {
		erOpts.PgSSLMode = jsnCfg.PgSSLMode
	}
	if jsnCfg.SQLDBNameProcessed != nil {
		erOpts.SQLDBNameProcessed = jsnCfg.SQLDBNameProcessed
	}
	if jsnCfg.SQLTableNameProcessed != nil {
		erOpts.SQLTableNameProcessed = jsnCfg.SQLTableNameProcessed
	}
	if jsnCfg.PgSSLModeProcessed != nil {
		erOpts.PgSSLModeProcessed = jsnCfg.PgSSLModeProcessed
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

func (erOpts *EventReaderOpts) Clone() *EventReaderOpts {
	cln := &EventReaderOpts{}
	if erOpts.PartialPath != nil {
		cln.PartialPath = utils.StringPointer(*erOpts.PartialPath)
	}
	if erOpts.PartialCacheAction != nil {
		cln.PartialCacheAction = utils.StringPointer(*erOpts.PartialCacheAction)
	}
	if erOpts.PartialOrderField != nil {
		cln.PartialOrderField = utils.StringPointer(*erOpts.PartialOrderField)
	}
	if erOpts.PartialCSVFieldSeparator != nil {
		cln.PartialCSVFieldSeparator = utils.StringPointer(*erOpts.PartialCSVFieldSeparator)
	}
	if erOpts.CSVRowLength != nil {
		cln.CSVRowLength = utils.IntPointer(*erOpts.CSVRowLength)
	}
	if erOpts.CSVFieldSeparator != nil {
		cln.CSVFieldSeparator = utils.StringPointer(*erOpts.CSVFieldSeparator)
	}
	if erOpts.CSVHeaderDefineChar != nil {
		cln.CSVHeaderDefineChar = utils.StringPointer(*erOpts.CSVHeaderDefineChar)
	}
	if erOpts.CSVLazyQuotes != nil {
		cln.CSVLazyQuotes = utils.BoolPointer(*erOpts.CSVLazyQuotes)
	}
	if erOpts.XMLRootPath != nil {
		cln.XMLRootPath = utils.StringPointer(*erOpts.XMLRootPath)
	}
	if erOpts.AMQPQueueID != nil {
		cln.AMQPQueueID = utils.StringPointer(*erOpts.AMQPQueueID)
	}
	if erOpts.AMQPQueueIDProcessed != nil {
		cln.AMQPQueueIDProcessed = utils.StringPointer(*erOpts.AMQPQueueIDProcessed)
	}
	if erOpts.AMQPConsumerTag != nil {
		cln.AMQPConsumerTag = utils.StringPointer(*erOpts.AMQPConsumerTag)
	}
	if erOpts.AMQPExchange != nil {
		cln.AMQPExchange = utils.StringPointer(*erOpts.AMQPExchange)
	}
	if erOpts.AMQPExchangeType != nil {
		cln.AMQPExchangeType = utils.StringPointer(*erOpts.AMQPExchangeType)
	}
	if erOpts.AMQPRoutingKey != nil {
		cln.AMQPRoutingKey = utils.StringPointer(*erOpts.AMQPRoutingKey)
	}
	if erOpts.AMQPExchangeProcessed != nil {
		cln.AMQPExchangeProcessed = utils.StringPointer(*erOpts.AMQPExchangeProcessed)
	}
	if erOpts.AMQPExchangeTypeProcessed != nil {
		cln.AMQPExchangeTypeProcessed = utils.StringPointer(*erOpts.AMQPExchangeTypeProcessed)
	}
	if erOpts.AMQPRoutingKeyProcessed != nil {
		cln.AMQPRoutingKeyProcessed = utils.StringPointer(*erOpts.AMQPRoutingKeyProcessed)
	}
	if erOpts.KafkaTopic != nil {
		cln.KafkaTopic = utils.StringPointer(*erOpts.KafkaTopic)
	}
	if erOpts.KafkaGroupID != nil {
		cln.KafkaGroupID = utils.StringPointer(*erOpts.KafkaGroupID)
	}
	if erOpts.KafkaMaxWait != nil {
		cln.KafkaMaxWait = utils.DurationPointer(*erOpts.KafkaMaxWait)
	}
	if erOpts.KafkaTopicProcessed != nil {
		cln.KafkaTopicProcessed = utils.StringPointer(*erOpts.KafkaTopicProcessed)
	}
	if erOpts.SQLDBName != nil {
		cln.SQLDBName = utils.StringPointer(*erOpts.SQLDBName)
	}
	if erOpts.SQLTableName != nil {
		cln.SQLTableName = utils.StringPointer(*erOpts.SQLTableName)
	}
	if erOpts.PgSSLMode != nil {
		cln.PgSSLMode = utils.StringPointer(*erOpts.PgSSLMode)
	}
	if erOpts.SQLDBNameProcessed != nil {
		cln.SQLDBNameProcessed = utils.StringPointer(*erOpts.SQLDBNameProcessed)
	}
	if erOpts.SQLTableNameProcessed != nil {
		cln.SQLTableNameProcessed = utils.StringPointer(*erOpts.SQLTableNameProcessed)
	}
	if erOpts.PgSSLModeProcessed != nil {
		cln.PgSSLModeProcessed = utils.StringPointer(*erOpts.PgSSLModeProcessed)
	}
	if erOpts.AWSRegion != nil {
		cln.AWSRegion = utils.StringPointer(*erOpts.AWSRegion)
	}
	if erOpts.AWSKey != nil {
		cln.AWSKey = utils.StringPointer(*erOpts.AWSKey)
	}
	if erOpts.AWSSecret != nil {
		cln.AWSSecret = utils.StringPointer(*erOpts.AWSSecret)
	}
	if erOpts.AWSToken != nil {
		cln.AWSToken = utils.StringPointer(*erOpts.AWSToken)
	}
	if erOpts.AWSRegionProcessed != nil {
		cln.AWSRegionProcessed = utils.StringPointer(*erOpts.AWSRegionProcessed)
	}
	if erOpts.AWSKeyProcessed != nil {
		cln.AWSKeyProcessed = utils.StringPointer(*erOpts.AWSKeyProcessed)
	}
	if erOpts.AWSSecretProcessed != nil {
		cln.AWSSecretProcessed = utils.StringPointer(*erOpts.AWSSecretProcessed)
	}
	if erOpts.AWSTokenProcessed != nil {
		cln.AWSTokenProcessed = utils.StringPointer(*erOpts.AWSTokenProcessed)
	}
	if erOpts.SQSQueueID != nil {
		cln.SQSQueueID = utils.StringPointer(*erOpts.SQSQueueID)
	}
	if erOpts.SQSQueueIDProcessed != nil {
		cln.SQSQueueIDProcessed = utils.StringPointer(*erOpts.SQSQueueIDProcessed)
	}
	if erOpts.S3BucketID != nil {
		cln.S3BucketID = utils.StringPointer(*erOpts.S3BucketID)
	}
	if erOpts.S3FolderPathProcessed != nil {
		cln.S3FolderPathProcessed = utils.StringPointer(*erOpts.S3FolderPathProcessed)
	}
	if erOpts.S3BucketIDProcessed != nil {
		cln.S3BucketIDProcessed = utils.StringPointer(*erOpts.S3BucketIDProcessed)
	}
	if erOpts.NATSJetStream != nil {
		cln.NATSJetStream = utils.BoolPointer(*erOpts.NATSJetStream)
	}
	if erOpts.NATSConsumerName != nil {
		cln.NATSConsumerName = utils.StringPointer(*erOpts.NATSConsumerName)
	}
	if erOpts.NATSSubject != nil {
		cln.NATSSubject = utils.StringPointer(*erOpts.NATSSubject)
	}
	if erOpts.NATSQueueID != nil {
		cln.NATSQueueID = utils.StringPointer(*erOpts.NATSQueueID)
	}
	if erOpts.NATSJWTFile != nil {
		cln.NATSJWTFile = utils.StringPointer(*erOpts.NATSJWTFile)
	}
	if erOpts.NATSSeedFile != nil {
		cln.NATSSeedFile = utils.StringPointer(*erOpts.NATSSeedFile)
	}
	if erOpts.NATSCertificateAuthority != nil {
		cln.NATSCertificateAuthority = utils.StringPointer(*erOpts.NATSCertificateAuthority)
	}
	if erOpts.NATSClientCertificate != nil {
		cln.NATSClientCertificate = utils.StringPointer(*erOpts.NATSClientCertificate)
	}
	if erOpts.NATSClientKey != nil {
		cln.NATSClientKey = utils.StringPointer(*erOpts.NATSClientKey)
	}
	if erOpts.NATSJetStreamMaxWait != nil {
		cln.NATSJetStreamMaxWait = utils.DurationPointer(*erOpts.NATSJetStreamMaxWait)
	}
	if erOpts.NATSJetStreamProcessed != nil {
		cln.NATSJetStreamProcessed = utils.BoolPointer(*erOpts.NATSJetStreamProcessed)
	}
	if erOpts.NATSSubjectProcessed != nil {
		cln.NATSSubjectProcessed = utils.StringPointer(*erOpts.NATSSubjectProcessed)
	}
	if erOpts.NATSJWTFileProcessed != nil {
		cln.NATSJWTFileProcessed = utils.StringPointer(*erOpts.NATSJWTFileProcessed)
	}
	if erOpts.NATSSeedFileProcessed != nil {
		cln.NATSSeedFileProcessed = utils.StringPointer(*erOpts.NATSSeedFileProcessed)
	}
	if erOpts.NATSCertificateAuthorityProcessed != nil {
		cln.NATSCertificateAuthorityProcessed = utils.StringPointer(*erOpts.NATSCertificateAuthorityProcessed)
	}
	if erOpts.NATSClientCertificateProcessed != nil {
		cln.NATSClientCertificateProcessed = utils.StringPointer(*erOpts.NATSClientCertificateProcessed)
	}
	if erOpts.NATSClientKeyProcessed != nil {
		cln.NATSClientKeyProcessed = utils.StringPointer(*erOpts.NATSClientKeyProcessed)
	}
	if erOpts.NATSJetStreamMaxWaitProcessed != nil {
		cln.NATSJetStreamMaxWaitProcessed = utils.DurationPointer(*erOpts.NATSJetStreamMaxWaitProcessed)
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
	if er.Opts.PgSSLMode != nil {
		opts[utils.PgSSLModeCfg] = *er.Opts.PgSSLMode
	}
	if er.Opts.SQLDBNameProcessed != nil {
		opts[utils.SQLDBNameProcessedCfg] = *er.Opts.SQLDBNameProcessed
	}
	if er.Opts.SQLTableNameProcessed != nil {
		opts[utils.SQLTableNameProcessedCfg] = *er.Opts.SQLTableNameProcessed
	}
	if er.Opts.PgSSLModeProcessed != nil {
		opts[utils.PgSSLModeProcessedCfg] = *er.Opts.PgSSLModeProcessed
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

	initialMP[utils.OptsCfg] = opts

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
