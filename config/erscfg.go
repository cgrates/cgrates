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
	"slices"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// ERsCfg the config for ERs
type ERsCfg struct {
	Enabled         bool
	SessionSConns   []string
	EEsConns        []string
	Readers         []*EventReaderCfg
	PartialCacheTTL time.Duration
}

// ReaderCfg iterates over the Readers slice and returns the reader
// configuration associated with the specified "id". If none were found, the
// method will return nil.
func (erS *ERsCfg) ReaderCfg(id string) *EventReaderCfg {
	for _, rdr := range erS.Readers {
		if rdr.ID == id {
			return rdr
		}
	}
	return nil
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
	if jsnCfg.SessionSConns != nil {
		erS.SessionSConns = updateInternalConns(*jsnCfg.SessionSConns, utils.MetaSessionS)
	}
	if jsnCfg.EEsConns != nil {
		erS.EEsConns = updateInternalConns(*jsnCfg.EEsConns, utils.MetaEEs)
	}
	if jsnCfg.PartialCacheTTL != nil {
		if erS.PartialCacheTTL, err = utils.ParseDurationWithNanosecs(*jsnCfg.PartialCacheTTL); err != nil {
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
		if jsnReader.ID != nil {
			for _, reader := range erS.Readers {
				if reader.ID == *jsnReader.ID {
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
		SessionSConns:   slices.Clone(erS.SessionSConns),
		EEsConns:        slices.Clone(erS.EEsConns),
		Readers:         make([]*EventReaderCfg, len(erS.Readers)),
		PartialCacheTTL: erS.PartialCacheTTL,
	}
	for idx, rdr := range erS.Readers {
		cln.Readers[idx] = rdr.Clone()
	}
	return
}

// AsMapInterface returns the config as a map[string]any
func (erS ERsCfg) AsMapInterface(separator string) any {
	mp := map[string]any{
		utils.EnabledCfg:         erS.Enabled,
		utils.PartialCacheTTLCfg: "0",
	}
	if erS.PartialCacheTTL != 0 {
		mp[utils.PartialCacheTTLCfg] = erS.PartialCacheTTL.String()
	}
	if erS.SessionSConns != nil {
		mp[utils.SessionSConnsCfg] = getInternalJSONConns(erS.SessionSConns)
	}
	if erS.EEsConns != nil {
		mp[utils.EEsConnsCfg] = getInternalJSONConns(erS.EEsConns)
	}
	if erS.Readers != nil {
		readers := make([]map[string]any, len(erS.Readers))
		for i, item := range erS.Readers {
			readers[i] = item.AsMapInterface(separator)
		}
		mp[utils.ReadersCfg] = readers
	}
	return mp
}

type EventReaderOpts struct {
	PartialPath              *string
	PartialCacheAction       *string
	PartialOrderField        *string
	PartialCSVFieldSeparator *string
	CSVRowLength             *int
	CSVFieldSeparator        *string
	CSVHeaderDefineChar      *string
	CSVLazyQuotes            *bool
	XMLRootPath              *string
	AMQPQueueID              *string
	AMQPUsername             *string
	AMQPPassword             *string
	AMQPConsumerTag          *string
	AMQPExchange             *string
	AMQPExchangeType         *string
	AMQPRoutingKey           *string
	KafkaTopic               *string
	KafkaGroupID             *string
	KafkaMaxWait             *time.Duration
	KafkaTLS                 *bool
	KafkaCAPath              *string
	KafkaSkipTLSVerify       *bool
	SQLDBName                *string
	SQLTableName             *string
	PgSSLMode                *string
	AWSRegion                *string
	AWSKey                   *string
	AWSSecret                *string
	AWSToken                 *string
	SQSQueueID               *string
	S3BucketID               *string
	NATSJetStream            *bool
	NATSConsumerName         *string
	NATSStreamName           *string
	NATSSubject              *string
	NATSQueueID              *string
	NATSJWTFile              *string
	NATSSeedFile             *string
	NATSCertificateAuthority *string
	NATSClientCertificate    *string
	NATSClientKey            *string
	NATSJetStreamMaxWait     *time.Duration
}

// EventReaderCfg the event for the Event Reader
type EventReaderCfg struct {
	ID   string
	Type string

	// RunDelay determines how the Serve method initiates the reading process.
	// 	- A value of 0 disables automatic reading, allowing manual control, possibly through an API.
	// 	- A value of -1 enables watching directory changes indefinitely, applicable for file-based readers.
	// 	- Any positive duration sets a fixed time interval for automatic reading cycles.
	RunDelay time.Duration

	// StartDelay adds a delay before starting reading loop
	StartDelay           time.Duration
	ConcurrentReqs       int
	SourcePath           string
	ProcessedPath        string
	Tenant               RSRParsers
	Timezone             string
	EEsSuccessIDs        []string
	EEsFailedIDs         []string
	Filters              []string
	Flags                utils.FlagsWithParams
	Reconnects           int
	MaxReconnectInterval time.Duration
	Opts                 *EventReaderOpts
	Fields               []*FCTemplate
	PartialCommitFields  []*FCTemplate
	CacheDumpFields      []*FCTemplate
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
	if jsnCfg.AMQPUsername != nil {
		erOpts.AMQPUsername = jsnCfg.AMQPUsername
	}
	if jsnCfg.AMQPPassword != nil {
		erOpts.AMQPPassword = jsnCfg.AMQPPassword
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
	if jsnCfg.KafkaTLS != nil {
		erOpts.KafkaTLS = jsnCfg.KafkaTLS
	}
	if jsnCfg.KafkaCAPath != nil {
		erOpts.KafkaCAPath = jsnCfg.KafkaCAPath
	}
	if jsnCfg.KafkaSkipTLSVerify != nil {
		erOpts.KafkaSkipTLSVerify = jsnCfg.KafkaSkipTLSVerify
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
	if jsnCfg.SQSQueueID != nil {
		erOpts.SQSQueueID = jsnCfg.SQSQueueID
	}
	if jsnCfg.S3BucketID != nil {
		erOpts.S3BucketID = jsnCfg.S3BucketID
	}
	if jsnCfg.NATSJetStream != nil {
		erOpts.NATSJetStream = jsnCfg.NATSJetStream
	}
	if jsnCfg.NATSConsumerName != nil {
		erOpts.NATSConsumerName = jsnCfg.NATSConsumerName
	}
	if jsnCfg.NATSStreamName != nil {
		erOpts.NATSStreamName = jsnCfg.NATSStreamName
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
	return
}

func (er *EventReaderCfg) loadFromJSONCfg(jsnCfg *EventReaderJsonCfg, msgTemplates map[string][]*FCTemplate, sep string) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.ID != nil {
		er.ID = *jsnCfg.ID
	}
	if jsnCfg.Type != nil {
		er.Type = *jsnCfg.Type
	}
	if jsnCfg.RunDelay != nil {
		if er.RunDelay, err = utils.ParseDurationWithNanosecs(*jsnCfg.RunDelay); err != nil {
			return
		}
	}
	if jsnCfg.StartDelay != nil {
		if er.StartDelay, err = utils.ParseDurationWithNanosecs(*jsnCfg.StartDelay); err != nil {
			return
		}
	}
	if jsnCfg.ConcurrentRequests != nil {
		er.ConcurrentReqs = *jsnCfg.ConcurrentRequests
	}
	if jsnCfg.SourcePath != nil {
		er.SourcePath = *jsnCfg.SourcePath
	}
	if jsnCfg.ProcessedPath != nil {
		er.ProcessedPath = *jsnCfg.ProcessedPath
	}
	if jsnCfg.Tenant != nil {
		if er.Tenant, err = NewRSRParsers(*jsnCfg.Tenant, sep); err != nil {
			return err
		}
	}
	if jsnCfg.Timezone != nil {
		er.Timezone = *jsnCfg.Timezone
	}
	if jsnCfg.EEsSuccessIDs != nil {
		er.EEsSuccessIDs = make([]string, len(*jsnCfg.EEsSuccessIDs))
		copy(er.EEsSuccessIDs, *jsnCfg.EEsSuccessIDs)
	}
	if jsnCfg.EEsFailedIDs != nil {
		er.EEsFailedIDs = make([]string, len(*jsnCfg.EEsFailedIDs))
		copy(er.EEsFailedIDs, *jsnCfg.EEsFailedIDs)
	}
	if jsnCfg.Filters != nil {
		er.Filters = slices.Clone(*jsnCfg.Filters)
	}
	if jsnCfg.Flags != nil {
		er.Flags = utils.FlagsWithParamsFromSlice(*jsnCfg.Flags)
	}
	if jsnCfg.Reconnects != nil {
		er.Reconnects = *jsnCfg.Reconnects
	}
	if jsnCfg.MaxReconnectInterval != nil {
		if er.MaxReconnectInterval, err = utils.ParseDurationWithNanosecs(*jsnCfg.MaxReconnectInterval); err != nil {
			return err
		}
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
	if jsnCfg.CacheDumpFields != nil {
		if er.CacheDumpFields, err = FCTemplatesFromFCTemplatesJSONCfg(*jsnCfg.CacheDumpFields, sep); err != nil {
			return err
		}
		if tpls, err := InflateTemplates(er.CacheDumpFields, msgTemplates); err != nil {
			return err
		} else if tpls != nil {
			er.CacheDumpFields = tpls
		}
	}
	if jsnCfg.PartialCommitFields != nil {
		if er.PartialCommitFields, err = FCTemplatesFromFCTemplatesJSONCfg(*jsnCfg.PartialCommitFields, sep); err != nil {
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
	if erOpts.PartialCSVFieldSeparator != nil {
		cln.PartialCSVFieldSeparator = new(string)
		*cln.PartialCSVFieldSeparator = *erOpts.PartialCSVFieldSeparator
	}
	if erOpts.CSVRowLength != nil {
		cln.CSVRowLength = new(int)
		*cln.CSVRowLength = *erOpts.CSVRowLength
	}
	if erOpts.CSVFieldSeparator != nil {
		cln.CSVFieldSeparator = new(string)
		*cln.CSVFieldSeparator = *erOpts.CSVFieldSeparator
	}
	if erOpts.CSVHeaderDefineChar != nil {
		cln.CSVHeaderDefineChar = new(string)
		*cln.CSVHeaderDefineChar = *erOpts.CSVHeaderDefineChar
	}
	if erOpts.CSVLazyQuotes != nil {
		cln.CSVLazyQuotes = new(bool)
		*cln.CSVLazyQuotes = *erOpts.CSVLazyQuotes
	}
	if erOpts.XMLRootPath != nil {
		cln.XMLRootPath = new(string)
		*cln.XMLRootPath = *erOpts.XMLRootPath
	}
	if erOpts.AMQPQueueID != nil {
		cln.AMQPQueueID = new(string)
		*cln.AMQPQueueID = *erOpts.AMQPQueueID
	}
	if erOpts.AMQPUsername != nil {
		cln.AMQPUsername = new(string)
		*cln.AMQPUsername = *erOpts.AMQPUsername
	}
	if erOpts.AMQPPassword != nil {
		cln.AMQPPassword = new(string)
		*cln.AMQPPassword = *erOpts.AMQPPassword
	}
	if erOpts.AMQPConsumerTag != nil {
		cln.AMQPConsumerTag = new(string)
		*cln.AMQPConsumerTag = *erOpts.AMQPConsumerTag
	}
	if erOpts.AMQPExchange != nil {
		cln.AMQPExchange = new(string)
		*cln.AMQPExchange = *erOpts.AMQPExchange
	}
	if erOpts.AMQPExchangeType != nil {
		cln.AMQPExchangeType = new(string)
		*cln.AMQPExchangeType = *erOpts.AMQPExchangeType
	}
	if erOpts.AMQPRoutingKey != nil {
		cln.AMQPRoutingKey = new(string)
		*cln.AMQPRoutingKey = *erOpts.AMQPRoutingKey
	}
	if erOpts.KafkaTopic != nil {
		cln.KafkaTopic = new(string)
		*cln.KafkaTopic = *erOpts.KafkaTopic
	}
	if erOpts.KafkaGroupID != nil {
		cln.KafkaGroupID = new(string)
		*cln.KafkaGroupID = *erOpts.KafkaGroupID
	}
	if erOpts.KafkaMaxWait != nil {
		cln.KafkaMaxWait = new(time.Duration)
		*cln.KafkaMaxWait = *erOpts.KafkaMaxWait
	}
	if erOpts.KafkaTLS != nil {
		cln.KafkaTLS = new(bool)
		*cln.KafkaTLS = *erOpts.KafkaTLS
	}
	if erOpts.KafkaCAPath != nil {
		cln.KafkaCAPath = new(string)
		*cln.KafkaCAPath = *erOpts.KafkaCAPath
	}
	if erOpts.KafkaSkipTLSVerify != nil {
		cln.KafkaSkipTLSVerify = new(bool)
		*cln.KafkaSkipTLSVerify = *erOpts.KafkaSkipTLSVerify
	}
	if erOpts.SQLDBName != nil {
		cln.SQLDBName = new(string)
		*cln.SQLDBName = *erOpts.SQLDBName
	}
	if erOpts.SQLTableName != nil {
		cln.SQLTableName = new(string)
		*cln.SQLTableName = *erOpts.SQLTableName
	}
	if erOpts.PgSSLMode != nil {
		cln.PgSSLMode = new(string)
		*cln.PgSSLMode = *erOpts.PgSSLMode
	}
	if erOpts.AWSRegion != nil {
		cln.AWSRegion = new(string)
		*cln.AWSRegion = *erOpts.AWSRegion
	}
	if erOpts.AWSKey != nil {
		cln.AWSKey = new(string)
		*cln.AWSKey = *erOpts.AWSKey
	}
	if erOpts.AWSSecret != nil {
		cln.AWSSecret = new(string)
		*cln.AWSSecret = *erOpts.AWSSecret
	}
	if erOpts.AWSToken != nil {
		cln.AWSToken = new(string)
		*cln.AWSToken = *erOpts.AWSToken
	}
	if erOpts.SQSQueueID != nil {
		cln.SQSQueueID = new(string)
		*cln.SQSQueueID = *erOpts.SQSQueueID
	}
	if erOpts.S3BucketID != nil {
		cln.S3BucketID = new(string)
		*cln.S3BucketID = *erOpts.S3BucketID
	}
	if erOpts.NATSJetStream != nil {
		cln.NATSJetStream = new(bool)
		*cln.NATSJetStream = *erOpts.NATSJetStream
	}
	if erOpts.NATSConsumerName != nil {
		cln.NATSConsumerName = new(string)
		*cln.NATSConsumerName = *erOpts.NATSConsumerName
	}
	if erOpts.NATSStreamName != nil {
		cln.NATSStreamName = new(string)
		*cln.NATSStreamName = *erOpts.NATSStreamName
	}
	if erOpts.NATSSubject != nil {
		cln.NATSSubject = new(string)
		*cln.NATSSubject = *erOpts.NATSSubject
	}
	if erOpts.NATSQueueID != nil {
		cln.NATSQueueID = new(string)
		*cln.NATSQueueID = *erOpts.NATSQueueID
	}
	if erOpts.NATSJWTFile != nil {
		cln.NATSJWTFile = new(string)
		*cln.NATSJWTFile = *erOpts.NATSJWTFile
	}
	if erOpts.NATSSeedFile != nil {
		cln.NATSSeedFile = new(string)
		*cln.NATSSeedFile = *erOpts.NATSSeedFile
	}
	if erOpts.NATSCertificateAuthority != nil {
		cln.NATSCertificateAuthority = new(string)
		*cln.NATSCertificateAuthority = *erOpts.NATSCertificateAuthority
	}
	if erOpts.NATSClientCertificate != nil {
		cln.NATSClientCertificate = new(string)
		*cln.NATSClientCertificate = *erOpts.NATSClientCertificate
	}
	if erOpts.NATSClientKey != nil {
		cln.NATSClientKey = new(string)
		*cln.NATSClientKey = *erOpts.NATSClientKey
	}
	if erOpts.NATSJetStreamMaxWait != nil {
		cln.NATSJetStreamMaxWait = new(time.Duration)
		*cln.NATSJetStreamMaxWait = *erOpts.NATSJetStreamMaxWait
	}
	return cln
}

// Clone returns a deep copy of EventReaderCfg
func (er EventReaderCfg) Clone() (cln *EventReaderCfg) {
	cln = &EventReaderCfg{
		ID:                   er.ID,
		Type:                 er.Type,
		StartDelay:           er.StartDelay,
		RunDelay:             er.RunDelay,
		ConcurrentReqs:       er.ConcurrentReqs,
		SourcePath:           er.SourcePath,
		ProcessedPath:        er.ProcessedPath,
		Tenant:               er.Tenant.Clone(),
		Timezone:             er.Timezone,
		EEsSuccessIDs:        slices.Clone(er.EEsSuccessIDs),
		EEsFailedIDs:         slices.Clone(er.EEsFailedIDs),
		Filters:              slices.Clone(er.Filters),
		Flags:                er.Flags.Clone(),
		Reconnects:           er.Reconnects,
		MaxReconnectInterval: er.MaxReconnectInterval,
		Opts:                 er.Opts.Clone(),
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
	if er.Opts.AMQPUsername != nil {
		opts[utils.AMQPUsername] = *er.Opts.AMQPUsername
	}
	if er.Opts.AMQPPassword != nil {
		opts[utils.AMQPPassword] = *er.Opts.AMQPPassword
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
	if er.Opts.KafkaTopic != nil {
		opts[utils.KafkaTopic] = *er.Opts.KafkaTopic
	}
	if er.Opts.KafkaGroupID != nil {
		opts[utils.KafkaGroupID] = *er.Opts.KafkaGroupID
	}
	if er.Opts.KafkaMaxWait != nil {
		opts[utils.KafkaMaxWait] = er.Opts.KafkaMaxWait.String()
	}
	if er.Opts.KafkaTLS != nil {
		opts[utils.KafkaTLS] = *er.Opts.KafkaTLS
	}
	if er.Opts.KafkaCAPath != nil {
		opts[utils.KafkaCAPath] = *er.Opts.KafkaCAPath
	}
	if er.Opts.KafkaSkipTLSVerify != nil {
		opts[utils.KafkaSkipTLSVerify] = *er.Opts.KafkaSkipTLSVerify
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
	if er.Opts.SQSQueueID != nil {
		opts[utils.SQSQueueID] = *er.Opts.SQSQueueID
	}
	if er.Opts.S3BucketID != nil {
		opts[utils.S3Bucket] = *er.Opts.S3BucketID
	}
	if er.Opts.NATSJetStream != nil {
		opts[utils.NatsJetStream] = *er.Opts.NATSJetStream
	}
	if er.Opts.NATSConsumerName != nil {
		opts[utils.NatsConsumerName] = *er.Opts.NATSConsumerName
	}
	if er.Opts.NATSStreamName != nil {
		opts[utils.NatsStreamName] = *er.Opts.NATSStreamName
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

	initialMP = map[string]any{
		utils.IDCfg:                   er.ID,
		utils.TypeCfg:                 er.Type,
		utils.ConcurrentRequestsCfg:   er.ConcurrentReqs,
		utils.SourcePathCfg:           er.SourcePath,
		utils.ProcessedPathCfg:        er.ProcessedPath,
		utils.TenantCfg:               er.Tenant.GetRule(separator),
		utils.TimezoneCfg:             er.Timezone,
		utils.FiltersCfg:              er.Filters,
		utils.FlagsCfg:                []string{},
		utils.RunDelayCfg:             "0",
		utils.StartDelayCfg:           "0",
		utils.ReconnectsCfg:           er.Reconnects,
		utils.MaxReconnectIntervalCfg: "0",
		utils.OptsCfg:                 opts,
	}
	if er.EEsSuccessIDs != nil {
		initialMP[utils.EEsSuccessIDsCfg] = er.EEsSuccessIDs
	}
	if er.EEsFailedIDs != nil {
		initialMP[utils.EEsFailedIDsCfg] = er.EEsFailedIDs
	}
	if er.MaxReconnectInterval != 0 {
		initialMP[utils.MaxReconnectIntervalCfg] = er.MaxReconnectInterval.String()
	}
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
	if er.StartDelay > 0 {
		initialMP[utils.StartDelayCfg] = er.StartDelay.String()
	}
	return
}

type EventReaderOptsJson struct {
	PartialPath              *string `json:"partialPath"`
	PartialCacheAction       *string `json:"partialCacheAction"`
	PartialOrderField        *string `json:"partialOrderField"`
	PartialCSVFieldSeparator *string `json:"partialcsvFieldSeparator"`
	CSVRowLength             *int    `json:"csvRowLength"`
	CSVFieldSeparator        *string `json:"csvFieldSeparator"`
	CSVHeaderDefineChar      *string `json:"csvHeaderDefineChar"`
	CSVLazyQuotes            *bool   `json:"csvLazyQuotes"`
	XMLRootPath              *string `json:"xmlRootPath"`
	AMQPQueueID              *string `json:"amqpQueueID"`
	AMQPUsername             *string `json:"amqpUsername"`
	AMQPPassword             *string `json:"amqpPassword"`
	AMQPConsumerTag          *string `json:"amqpConsumerTag"`
	AMQPExchange             *string `json:"amqpExchange"`
	AMQPExchangeType         *string `json:"amqpExchangeType"`
	AMQPRoutingKey           *string `json:"amqpRoutingKey"`
	KafkaTopic               *string `json:"kafkaTopic"`
	KafkaGroupID             *string `json:"kafkaGroupID"`
	KafkaMaxWait             *string `json:"kafkaMaxWait"`
	KafkaTLS                 *bool   `json:"kafkaTLS"`
	KafkaCAPath              *string `json:"kafkaCAPath"`
	KafkaSkipTLSVerify       *bool   `json:"kafkaSkipTLSVerify"`
	SQLDBName                *string `json:"sqlDBName"`
	SQLTableName             *string `json:"sqlTableName"`
	PgSSLMode                *string `json:"pgSSLMode"`
	AWSRegion                *string `json:"awsRegion"`
	AWSKey                   *string `json:"awsKey"`
	AWSSecret                *string `json:"awsSecret"`
	AWSToken                 *string `json:"awsToken"`
	SQSQueueID               *string `json:"sqsQueueID"`
	S3BucketID               *string `json:"s3BucketID"`
	NATSJetStream            *bool   `json:"natsJetStream"`
	NATSConsumerName         *string `json:"natsConsumerName"`
	NATSStreamName           *string `json:"natsStreamName"`
	NATSSubject              *string `json:"natsSubject"`
	NATSQueueID              *string `json:"natsQueueID"`
	NATSJWTFile              *string `json:"natsJWTFile"`
	NATSSeedFile             *string `json:"natsSeedFile"`
	NATSCertificateAuthority *string `json:"natsCertificateAuthority"`
	NATSClientCertificate    *string `json:"natsClientCertificate"`
	NATSClientKey            *string `json:"natsClientKey"`
	NATSJetStreamMaxWait     *string `json:"natsJetStreamMaxWait"`
}

// EventReaderSJsonCfg is the configuration of a single EventReader
type EventReaderJsonCfg struct {
	ID                   *string               `json:"id"`
	Type                 *string               `json:"type"`
	RunDelay             *string               `json:"run_delay"`
	StartDelay           *string               `json:"start_delay"`
	ConcurrentRequests   *int                  `json:"concurrent_requests"`
	SourcePath           *string               `json:"source_path"`
	ProcessedPath        *string               `json:"processed_path"`
	Tenant               *string               `json:"tenant"`
	Timezone             *string               `json:"timezone"`
	EEsSuccessIDs        *[]string             `json:"ees_success_ids"`
	EEsFailedIDs         *[]string             `json:"ees_failed_ids"`
	Filters              *[]string             `json:"filters"`
	Flags                *[]string             `json:"flags"`
	Reconnects           *int                  `json:"reconnects"`
	MaxReconnectInterval *string               `json:"max_reconnect_interval"`
	Opts                 *EventReaderOptsJson  `json:"opts"`
	Fields               *[]*FcTemplateJsonCfg `json:"fields"`
	PartialCommitFields  *[]*FcTemplateJsonCfg `json:"partial_commit_fields"`
	CacheDumpFields      *[]*FcTemplateJsonCfg `json:"cache_dump_fields"`
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
	if v2.AMQPUsername != nil {
		if v1.AMQPUsername == nil ||
			*v1.AMQPUsername != *v2.AMQPUsername {
			d.AMQPUsername = v2.AMQPUsername
		}
	} else {
		d.AMQPUsername = nil
	}
	if v2.AMQPPassword != nil {
		if v1.AMQPPassword == nil ||
			*v1.AMQPPassword != *v2.AMQPPassword {
			d.AMQPPassword = v2.AMQPPassword
		}
	} else {
		d.AMQPPassword = nil
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
	if v2.KafkaTLS != nil {
		if v1.KafkaTLS == nil ||
			*v1.KafkaTLS != *v2.KafkaTLS {
			d.KafkaTLS = v2.KafkaTLS
		}
	} else {
		d.KafkaTLS = nil
	}
	if v2.KafkaCAPath != nil {
		if v1.KafkaCAPath == nil ||
			*v1.KafkaCAPath != *v2.KafkaCAPath {
			d.KafkaCAPath = v2.KafkaCAPath
		}
	} else {
		d.KafkaCAPath = nil
	}
	if v2.KafkaSkipTLSVerify != nil {
		if v1.KafkaSkipTLSVerify == nil ||
			*v1.KafkaSkipTLSVerify != *v2.KafkaSkipTLSVerify {
			d.KafkaSkipTLSVerify = v2.KafkaSkipTLSVerify
		}
	} else {
		d.KafkaSkipTLSVerify = nil
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
	if v2.PgSSLMode != nil {
		if v1.PgSSLMode == nil ||
			*v1.PgSSLMode != *v2.PgSSLMode {
			d.PgSSLMode = v2.PgSSLMode
		}
	} else {
		d.PgSSLMode = nil
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
	if v2.SQSQueueID != nil {
		if v1.SQSQueueID == nil ||
			*v1.SQSQueueID != *v2.SQSQueueID {
			d.SQSQueueID = v2.SQSQueueID
		}
	} else {
		d.SQSQueueID = nil
	}
	if v2.S3BucketID != nil {
		if v1.S3BucketID == nil ||
			*v1.S3BucketID != *v2.S3BucketID {
			d.S3BucketID = v2.S3BucketID
		}
	} else {
		d.S3BucketID = nil
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
	if v2.NATSStreamName != nil {
		if v1.NATSStreamName == nil ||
			*v1.NATSStreamName != *v2.NATSStreamName {
			d.NATSStreamName = v2.NATSStreamName
		}
	} else {
		d.NATSStreamName = nil
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
	return d
}

func diffEventReaderJsonCfg(d *EventReaderJsonCfg, v1, v2 *EventReaderCfg, separator string) *EventReaderJsonCfg {
	if d == nil {
		d = new(EventReaderJsonCfg)
	}
	if v1.ID != v2.ID {
		d.ID = utils.StringPointer(v2.ID)
	}
	if v1.Type != v2.Type {
		d.Type = utils.StringPointer(v2.Type)
	}
	if v1.RunDelay != v2.RunDelay {
		d.RunDelay = utils.StringPointer(v2.RunDelay.String())
	}
	if v1.StartDelay != v2.StartDelay {
		d.StartDelay = utils.StringPointer(v2.StartDelay.String())
	}
	if v1.ConcurrentReqs != v2.ConcurrentReqs {
		d.ConcurrentRequests = utils.IntPointer(v2.ConcurrentReqs)
	}
	if v1.SourcePath != v2.SourcePath {
		d.SourcePath = utils.StringPointer(v2.SourcePath)
	}
	if v1.ProcessedPath != v2.ProcessedPath {
		d.ProcessedPath = utils.StringPointer(v2.ProcessedPath)
	}
	if v1.Reconnects != v2.Reconnects {
		d.Reconnects = utils.IntPointer(v2.ConcurrentReqs)
	}
	if v1.MaxReconnectInterval != v2.MaxReconnectInterval {
		d.MaxReconnectInterval = utils.StringPointer(v2.MaxReconnectInterval.String())
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
	if !slices.Equal(v1.EEsSuccessIDs, v2.EEsSuccessIDs) {
		d.EEsSuccessIDs = &v2.EEsSuccessIDs
	}
	if !slices.Equal(v1.EEsFailedIDs, v2.EEsFailedIDs) {
		d.EEsFailedIDs = &v2.EEsFailedIDs
	}
	if !slices.Equal(v1.Filters, v2.Filters) {
		d.Filters = &v2.Filters
	}
	flgs1 := v1.Flags.SliceFlags()
	flgs2 := v2.Flags.SliceFlags()
	if !slices.Equal(flgs1, flgs2) {
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
	if d.PartialCommitFields != nil {
		pcf = *d.PartialCommitFields
	}
	pcf = diffFcTemplateJsonCfg(pcf, v1.PartialCommitFields, v2.PartialCommitFields, separator)
	if pcf != nil {
		d.PartialCommitFields = &pcf
	}

	var cdf []*FcTemplateJsonCfg
	if d.CacheDumpFields != nil {
		cdf = *d.CacheDumpFields
	}
	cdf = diffFcTemplateJsonCfg(cdf, v1.CacheDumpFields, v2.CacheDumpFields, separator)
	if cdf != nil {
		d.CacheDumpFields = &cdf
	}

	return d
}

func getEventReaderJsonCfg(d []*EventReaderJsonCfg, id string) (*EventReaderJsonCfg, int) {
	for i, v := range d {
		if v.ID != nil && *v.ID == id {
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
	Enabled         *bool                  `json:"enabled"`
	SessionSConns   *[]string              `json:"sessions_conns"`
	EEsConns        *[]string              `json:"ees_conns"`
	Readers         *[]*EventReaderJsonCfg `json:"readers"`
	PartialCacheTTL *string                `json:"partial_cache_ttl"`
}

func diffERsJsonCfg(d *ERsJsonCfg, v1, v2 *ERsCfg, separator string) *ERsJsonCfg {
	if d == nil {
		d = new(ERsJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if !slices.Equal(v1.SessionSConns, v2.SessionSConns) {
		d.SessionSConns = utils.SliceStringPointer(getInternalJSONConns(v2.SessionSConns))
	}
	if !slices.Equal(v1.EEsConns, v2.EEsConns) {
		d.EEsConns = utils.SliceStringPointer(getInternalJSONConns(v2.EEsConns))
	}
	if v1.PartialCacheTTL != v2.PartialCacheTTL {
		d.PartialCacheTTL = utils.StringPointer(v2.PartialCacheTTL.String())
	}
	d.Readers = diffEventReadersJsonCfg(d.Readers, v1.Readers, v2.Readers, separator)
	return d
}
