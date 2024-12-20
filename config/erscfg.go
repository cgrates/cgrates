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

	"github.com/cgrates/cgrates/utils"
)

// ERsCfg the config for ERs
type ERsCfg struct {
	Enabled          bool
	SessionSConns    []string
	EEsConns         []string
	ConcurrentEvents int
	Readers          []*EventReaderCfg
	PartialCacheTTL  time.Duration
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

func (erS *ERsCfg) loadFromJSONCfg(jsnCfg *ERsJsonCfg, msgTemplates map[string][]*FCTemplate, sep string, dfltRdrCfg *EventReaderCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		erS.Enabled = *jsnCfg.Enabled
	}

	if jsnCfg.Sessions_conns != nil {
		erS.SessionSConns = make([]string, 0, len(*jsnCfg.Sessions_conns))
		for _, fID := range *jsnCfg.Sessions_conns {

			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if fID != utils.MetaInternal {
				erS.SessionSConns = append(erS.SessionSConns, fID)
			} else {
				erS.SessionSConns = append(erS.SessionSConns, utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS))
			}
		}
	}
	if jsnCfg.Ees_conns != nil {
		erS.EEsConns = make([]string, 0, len(*jsnCfg.Ees_conns))
		for _, fID := range *jsnCfg.Ees_conns {

			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if fID != utils.MetaInternal {
				erS.EEsConns = append(erS.EEsConns, fID)
			} else {
				erS.EEsConns = append(erS.EEsConns, utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs))
			}
		}
	}

	if jsnCfg.Concurrent_events != nil {
		erS.ConcurrentEvents = *jsnCfg.Concurrent_events
		if erS.ConcurrentEvents < 1 {
			erS.ConcurrentEvents = 1
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
		Enabled:          erS.Enabled,
		SessionSConns:    slices.Clone(erS.SessionSConns),
		EEsConns:         slices.Clone(erS.EEsConns),
		Readers:          make([]*EventReaderCfg, len(erS.Readers)),
		ConcurrentEvents: erS.ConcurrentEvents,
		PartialCacheTTL:  erS.PartialCacheTTL,
	}
	for idx, rdr := range erS.Readers {
		cln.Readers[idx] = rdr.Clone()
	}
	return
}

// AsMapInterface returns the config as a map[string]any
func (erS *ERsCfg) AsMapInterface(separator string) (initialMP map[string]any) {
	initialMP = map[string]any{
		utils.EnabledCfg:          erS.Enabled,
		utils.ConcurrentEventsCfg: erS.ConcurrentEvents,
		utils.PartialCacheTTLCfg:  "0",
	}
	if erS.PartialCacheTTL != 0 {
		initialMP[utils.PartialCacheTTLCfg] = erS.PartialCacheTTL.String()
	}
	if erS.SessionSConns != nil {
		sessionSConns := make([]string, 0, len(erS.SessionSConns))
		for _, item := range erS.SessionSConns {
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS) {
				sessionSConns = append(sessionSConns, utils.MetaInternal)
			} else {
				sessionSConns = append(sessionSConns, item)
			}
		}
		initialMP[utils.SessionSConnsCfg] = sessionSConns
	}
	if erS.EEsConns != nil {
		eesConns := make([]string, 0, len(erS.EEsConns))
		for _, item := range erS.EEsConns {
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs) {
				eesConns = append(eesConns, utils.MetaInternal)
			} else {
				eesConns = append(eesConns, item)
			}
		}
		initialMP[utils.EEsConnsCfg] = eesConns
	}
	if erS.Readers != nil {
		readers := make([]map[string]any, 0, len(erS.Readers))
		for _, item := range erS.Readers {
			readers = append(readers, item.AsMapInterface(separator))
		}
		initialMP[utils.ReadersCfg] = readers
	}
	return
}

type AMQPROpts struct {
	QueueID      *string
	Username     *string
	Password     *string
	ConsumerTag  *string
	Exchange     *string
	ExchangeType *string
	RoutingKey   *string
}

func (amqpr *AMQPROpts) loadFromJSONCfg(jsnCfg *EventReaderOptsJson) (err error) {
	if jsnCfg.AMQPQueueID != nil {
		amqpr.QueueID = jsnCfg.AMQPQueueID
	}
	if jsnCfg.AMQPUsername != nil {
		amqpr.Username = jsnCfg.AMQPUsername
	}
	if jsnCfg.AMQPPassword != nil {
		amqpr.Password = jsnCfg.AMQPPassword
	}
	if jsnCfg.AMQPConsumerTag != nil {
		amqpr.ConsumerTag = jsnCfg.AMQPConsumerTag
	}
	if jsnCfg.AMQPExchange != nil {
		amqpr.Exchange = jsnCfg.AMQPExchange
	}
	if jsnCfg.AMQPExchangeType != nil {
		amqpr.ExchangeType = jsnCfg.AMQPExchangeType
	}
	if jsnCfg.AMQPRoutingKey != nil {
		amqpr.RoutingKey = jsnCfg.AMQPRoutingKey
	}
	return
}

type KafkaROpts struct {
	Topic         *string
	GroupID       *string
	MaxWait       *time.Duration
	TLS           *bool
	CAPath        *string
	SkipTLSVerify *bool
}

func (kafkaROpts *KafkaROpts) loadFromJSONCfg(jsnCfg *EventReaderOptsJson) (err error) {
	if jsnCfg.KafkaTopic != nil {
		kafkaROpts.Topic = jsnCfg.KafkaTopic
	}
	if jsnCfg.KafkaGroupID != nil {
		kafkaROpts.GroupID = jsnCfg.KafkaGroupID
	}
	if jsnCfg.KafkaMaxWait != nil {
		var kafkaMaxWait time.Duration
		if kafkaMaxWait, err = utils.ParseDurationWithNanosecs(*jsnCfg.KafkaMaxWait); err != nil {
			return
		}
		kafkaROpts.MaxWait = utils.DurationPointer(kafkaMaxWait)
	}
	if jsnCfg.KafkaTLS != nil {
		kafkaROpts.TLS = jsnCfg.KafkaTLS
	}
	if jsnCfg.KafkaCAPath != nil {
		kafkaROpts.CAPath = jsnCfg.KafkaCAPath
	}
	if jsnCfg.KafkaSkipTLSVerify != nil {
		kafkaROpts.SkipTLSVerify = jsnCfg.KafkaSkipTLSVerify
	}
	return
}

type SQLROpts struct {
	DBName              *string
	TableName           *string
	DeleteIndexedFields *[]string
	PgSSLMode           *string
}

func (sqlOpts *SQLROpts) loadFromJSONCfg(jsnCfg *EventReaderOptsJson) (err error) {
	if jsnCfg.SQLDBName != nil {
		sqlOpts.DBName = jsnCfg.SQLDBName
	}
	if jsnCfg.SQLTableName != nil {
		sqlOpts.TableName = jsnCfg.SQLTableName
	}
	if jsnCfg.SQLDeleteIndexedFields != nil {
		dif := make([]string, len(*jsnCfg.SQLDeleteIndexedFields))
		copy(dif, *jsnCfg.SQLDeleteIndexedFields)
		sqlOpts.DeleteIndexedFields = &dif
	}
	if jsnCfg.PgSSLMode != nil {
		sqlOpts.PgSSLMode = jsnCfg.PgSSLMode
	}
	return
}

type AWSROpts struct {
	Region     *string
	Key        *string
	Secret     *string
	Token      *string
	SQSQueueID *string
	S3BucketID *string
}

func (awsROpts *AWSROpts) loadFromJSONCfg(jsnCfg *EventReaderOptsJson) (err error) {

	if jsnCfg.AWSRegion != nil {
		awsROpts.Region = jsnCfg.AWSRegion
	}
	if jsnCfg.AWSKey != nil {
		awsROpts.Key = jsnCfg.AWSKey
	}
	if jsnCfg.AWSSecret != nil {
		awsROpts.Secret = jsnCfg.AWSSecret
	}
	if jsnCfg.AWSToken != nil {
		awsROpts.Token = jsnCfg.AWSToken
	}
	if jsnCfg.SQSQueueID != nil {
		awsROpts.SQSQueueID = jsnCfg.SQSQueueID
	}
	if jsnCfg.S3BucketID != nil {
		awsROpts.S3BucketID = jsnCfg.S3BucketID
	}
	return
}

type NATSROpts struct {
	JetStream            *bool
	ConsumerName         *string
	StreamName           *string
	Subject              *string
	QueueID              *string
	JWTFile              *string
	SeedFile             *string
	CertificateAuthority *string
	ClientCertificate    *string
	ClientKey            *string
	JetStreamMaxWait     *time.Duration
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
	return
}

type CSVROpts struct {
	PartialCSVFieldSeparator *string
	RowLength                *int
	FieldSeparator           *string
	HeaderDefineChar         *string
	LazyQuotes               *bool
}

func (csvROpts *CSVROpts) loadFromJSONCfg(jsnCfg *EventReaderOptsJson) (err error) {
	if jsnCfg.PartialCSVFieldSeparator != nil {
		csvROpts.PartialCSVFieldSeparator = jsnCfg.PartialCSVFieldSeparator
	}
	if jsnCfg.CSVRowLength != nil {
		csvROpts.RowLength = jsnCfg.CSVRowLength
	}
	if jsnCfg.CSVFieldSeparator != nil {
		csvROpts.FieldSeparator = jsnCfg.CSVFieldSeparator
	}
	if jsnCfg.CSVHeaderDefineChar != nil {
		csvROpts.HeaderDefineChar = jsnCfg.CSVHeaderDefineChar
	}
	if jsnCfg.CSVLazyQuotes != nil {
		csvROpts.LazyQuotes = jsnCfg.CSVLazyQuotes
	}
	return
}

type EventReaderOpts struct {
	PartialPath        *string
	PartialCacheAction *string
	PartialOrderField  *string
	XMLRootPath        *string
	CSV                *CSVROpts
	AMQP               *AMQPROpts
	AWS                *AWSROpts
	NATS               *NATSROpts
	Kafka              *KafkaROpts
	SQL                *SQLROpts
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
	Filters              []string
	Flags                utils.FlagsWithParams
	Reconnects           int
	MaxReconnectInterval time.Duration
	EEsIDs               []string
	EEsSuccessIDs        []string
	EEsFailedIDs         []string
	Opts                 *EventReaderOpts
	Fields               []*FCTemplate
	PartialCommitFields  []*FCTemplate
	CacheDumpFields      []*FCTemplate
}

func (erOpts *EventReaderOpts) loadFromJSONCfg(jsnCfg *EventReaderOptsJson) (err error) {
	if jsnCfg == nil {
		return
	}
	if err = erOpts.AMQP.loadFromJSONCfg(jsnCfg); err != nil {
		return
	}
	if err = erOpts.AWS.loadFromJSONCfg(jsnCfg); err != nil {
		return
	}
	if err = erOpts.Kafka.loadFromJSONCfg(jsnCfg); err != nil {
		return
	}
	if err = erOpts.NATS.loadFromJSONCfg(jsnCfg); err != nil {
		return
	}
	if err = erOpts.SQL.loadFromJSONCfg(jsnCfg); err != nil {
		return
	}
	if err = erOpts.CSV.loadFromJSONCfg(jsnCfg); err != nil {
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
	if jsnCfg.Start_delay != nil {
		if er.StartDelay, err = utils.ParseDurationWithNanosecs(*jsnCfg.Start_delay); err != nil {
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
		copy(er.Filters, *jsnCfg.Filters)
	}
	if jsnCfg.Flags != nil {
		er.Flags = utils.FlagsWithParamsFromSlice(*jsnCfg.Flags)
	}
	if jsnCfg.Reconnects != nil {
		er.Reconnects = *jsnCfg.Reconnects
	}
	if jsnCfg.Max_reconnect_interval != nil {
		if er.MaxReconnectInterval, err = utils.ParseDurationWithNanosecs(*jsnCfg.Max_reconnect_interval); err != nil {
			return err
		}
	}
	if jsnCfg.Ees_ids != nil {
		er.EEsIDs = make([]string, len(*jsnCfg.Ees_ids))
		copy(er.EEsIDs, *jsnCfg.Ees_ids)
	}
	if jsnCfg.Ees_success_ids != nil {
		er.EEsSuccessIDs = make([]string, len(*jsnCfg.Ees_success_ids))
		copy(er.EEsSuccessIDs, *jsnCfg.Ees_success_ids)
	}
	if jsnCfg.Ees_failed_ids != nil {
		er.EEsFailedIDs = make([]string, len(*jsnCfg.Ees_failed_ids))
		copy(er.EEsFailedIDs, *jsnCfg.Ees_failed_ids)
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
	if amqpOpts.QueueID != nil {
		cln.QueueID = new(string)
		*cln.QueueID = *amqpOpts.QueueID
	}
	if amqpOpts.Username != nil {
		cln.Username = new(string)
		*cln.Username = *amqpOpts.Username
	}
	if amqpOpts.Password != nil {
		cln.Password = new(string)
		*cln.Password = *amqpOpts.Password
	}
	if amqpOpts.ConsumerTag != nil {
		cln.ConsumerTag = new(string)
		*cln.ConsumerTag = *amqpOpts.ConsumerTag
	}
	if amqpOpts.Exchange != nil {
		cln.Exchange = new(string)
		*cln.Exchange = *amqpOpts.Exchange
	}
	if amqpOpts.ExchangeType != nil {
		cln.ExchangeType = new(string)
		*cln.ExchangeType = *amqpOpts.ExchangeType
	}
	if amqpOpts.RoutingKey != nil {
		cln.RoutingKey = new(string)
		*cln.RoutingKey = *amqpOpts.RoutingKey
	}
	return cln
}

func (csvOpts *CSVROpts) Clone() *CSVROpts {
	cln := &CSVROpts{}
	if csvOpts.PartialCSVFieldSeparator != nil {
		cln.PartialCSVFieldSeparator = new(string)
		*cln.PartialCSVFieldSeparator = *csvOpts.PartialCSVFieldSeparator
	}
	if csvOpts.RowLength != nil {
		cln.RowLength = new(int)
		*cln.RowLength = *csvOpts.RowLength
	}
	if csvOpts.FieldSeparator != nil {
		cln.FieldSeparator = new(string)
		*cln.FieldSeparator = *csvOpts.FieldSeparator
	}
	if csvOpts.HeaderDefineChar != nil {
		cln.HeaderDefineChar = new(string)
		*cln.HeaderDefineChar = *csvOpts.HeaderDefineChar
	}
	if csvOpts.LazyQuotes != nil {
		cln.LazyQuotes = new(bool)
		*cln.LazyQuotes = *csvOpts.LazyQuotes
	}
	return cln
}
func (kafkaOpts *KafkaROpts) Clone() *KafkaROpts {
	cln := &KafkaROpts{}
	if kafkaOpts.Topic != nil {
		cln.Topic = new(string)
		*cln.Topic = *kafkaOpts.Topic
	}
	if kafkaOpts.GroupID != nil {
		cln.GroupID = new(string)
		*cln.GroupID = *kafkaOpts.GroupID
	}
	if kafkaOpts.MaxWait != nil {
		cln.MaxWait = new(time.Duration)
		*cln.MaxWait = *kafkaOpts.MaxWait
	}
	if kafkaOpts.TLS != nil {
		cln.TLS = new(bool)
		*cln.TLS = *kafkaOpts.TLS
	}
	if kafkaOpts.CAPath != nil {
		cln.CAPath = new(string)
		*cln.CAPath = *kafkaOpts.CAPath
	}
	if kafkaOpts.SkipTLSVerify != nil {
		cln.SkipTLSVerify = new(bool)
		*cln.SkipTLSVerify = *kafkaOpts.SkipTLSVerify
	}
	return cln
}

func (sqlOpts *SQLROpts) Clone() *SQLROpts {
	cln := &SQLROpts{}
	if sqlOpts.DBName != nil {
		cln.DBName = new(string)
		*cln.DBName = *sqlOpts.DBName
	}
	if sqlOpts.TableName != nil {
		cln.TableName = new(string)
		*cln.TableName = *sqlOpts.TableName
	}
	if sqlOpts.DeleteIndexedFields != nil {
		idx := make([]string, len(*sqlOpts.DeleteIndexedFields))
		copy(idx, *sqlOpts.DeleteIndexedFields)
		cln.DeleteIndexedFields = &idx
	}
	if sqlOpts.PgSSLMode != nil {
		cln.PgSSLMode = new(string)
		*cln.PgSSLMode = *sqlOpts.PgSSLMode
	}
	return cln
}

func (awsOpt *AWSROpts) Clone() *AWSROpts {
	cln := &AWSROpts{}
	if awsOpt.Region != nil {
		cln.Region = new(string)
		*cln.Region = *awsOpt.Region
	}
	if awsOpt.Key != nil {
		cln.Key = new(string)
		*cln.Key = *awsOpt.Key
	}
	if awsOpt.Secret != nil {
		cln.Secret = new(string)
		*cln.Secret = *awsOpt.Secret
	}
	if awsOpt.Token != nil {
		cln.Token = new(string)
		*cln.Token = *awsOpt.Token
	}
	if awsOpt.SQSQueueID != nil {
		cln.SQSQueueID = new(string)
		*cln.SQSQueueID = *awsOpt.SQSQueueID
	}
	if awsOpt.S3BucketID != nil {
		cln.S3BucketID = new(string)
		*cln.S3BucketID = *awsOpt.S3BucketID
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
	if erOpts.CSV != nil {
		cln.CSV = erOpts.CSV.Clone()
	}
	if erOpts.XMLRootPath != nil {
		cln.XMLRootPath = new(string)
		*cln.XMLRootPath = *erOpts.XMLRootPath
	}
	if erOpts.AMQP != nil {
		cln.AMQP = erOpts.AMQP.Clone()
	}
	if erOpts.NATS != nil {
		cln.NATS = erOpts.NATS.Clone()
	}
	if erOpts.Kafka != nil {
		cln.Kafka = erOpts.Kafka.Clone()
	}
	if erOpts.SQL != nil {
		cln.SQL = erOpts.SQL.Clone()
	}
	if erOpts.AWS != nil {
		cln.AWS = erOpts.AWS.Clone()
	}

	return cln
}

// Clone returns a deep copy of EventReaderCfg
func (er EventReaderCfg) Clone() (cln *EventReaderCfg) {
	cln = &EventReaderCfg{
		ID:                   er.ID,
		Type:                 er.Type,
		RunDelay:             er.RunDelay,
		StartDelay:           er.StartDelay,
		ConcurrentReqs:       er.ConcurrentReqs,
		SourcePath:           er.SourcePath,
		ProcessedPath:        er.ProcessedPath,
		Filters:              slices.Clone(er.Filters),
		Tenant:               er.Tenant.Clone(),
		Timezone:             er.Timezone,
		Flags:                er.Flags.Clone(),
		Reconnects:           er.Reconnects,
		MaxReconnectInterval: er.MaxReconnectInterval,
		EEsIDs:               slices.Clone(er.EEsIDs),
		EEsSuccessIDs:        slices.Clone(er.EEsSuccessIDs),
		EEsFailedIDs:         slices.Clone(er.EEsFailedIDs),
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

	if csvOpts := er.Opts.CSV; csvOpts != nil {
		if csvOpts.PartialCSVFieldSeparator != nil {
			opts[utils.PartialCSVFieldSepartorOpt] = *csvOpts.PartialCSVFieldSeparator
		}
		if csvOpts.RowLength != nil {
			opts[utils.CSVRowLengthOpt] = *csvOpts.RowLength
		}
		if csvOpts.FieldSeparator != nil {
			opts[utils.CSVFieldSepOpt] = *csvOpts.FieldSeparator
		}
		if csvOpts.HeaderDefineChar != nil {
			opts[utils.HeaderDefineCharOpt] = *csvOpts.HeaderDefineChar
		}
		if csvOpts.LazyQuotes != nil {
			opts[utils.CSVLazyQuotes] = *csvOpts.LazyQuotes
		}
	}
	if er.Opts.XMLRootPath != nil {
		opts[utils.XMLRootPathOpt] = *er.Opts.XMLRootPath
	}
	if amqpOpts := er.Opts.AMQP; amqpOpts != nil {
		if amqpOpts.QueueID != nil {
			opts[utils.AMQPQueueID] = *amqpOpts.QueueID
		}
		if amqpOpts.Username != nil {
			opts[utils.AMQPUsername] = *amqpOpts.Username
		}
		if amqpOpts.Password != nil {
			opts[utils.AMQPPassword] = *amqpOpts.Password
		}
		if amqpOpts.ConsumerTag != nil {
			opts[utils.AMQPConsumerTag] = *amqpOpts.ConsumerTag
		}
		if amqpOpts.Exchange != nil {
			opts[utils.AMQPExchange] = *amqpOpts.Exchange
		}
		if amqpOpts.ExchangeType != nil {
			opts[utils.AMQPExchangeType] = *amqpOpts.ExchangeType
		}
		if amqpOpts.RoutingKey != nil {
			opts[utils.AMQPRoutingKey] = *amqpOpts.RoutingKey
		}
	}

	if kafkaOpts := er.Opts.Kafka; kafkaOpts != nil {
		if kafkaOpts.Topic != nil {
			opts[utils.KafkaTopic] = *kafkaOpts.Topic
		}
		if kafkaOpts.GroupID != nil {
			opts[utils.KafkaGroupID] = *kafkaOpts.GroupID
		}
		if kafkaOpts.MaxWait != nil {
			opts[utils.KafkaMaxWait] = kafkaOpts.MaxWait.String()
		}
		if kafkaOpts.TLS != nil {
			opts[utils.KafkaTLS] = *kafkaOpts.TLS
		}
		if kafkaOpts.CAPath != nil {
			opts[utils.KafkaCAPath] = *kafkaOpts.CAPath
		}
		if kafkaOpts.SkipTLSVerify != nil {
			opts[utils.KafkaSkipTLSVerify] = *kafkaOpts.SkipTLSVerify
		}
	}

	if sqlOpts := er.Opts.SQL; sqlOpts != nil {
		if sqlOpts.DBName != nil {
			opts[utils.SQLDBNameOpt] = *sqlOpts.DBName
		}
		if sqlOpts.TableName != nil {
			opts[utils.SQLTableNameOpt] = *sqlOpts.TableName
		}
		if sqlOpts.DeleteIndexedFields != nil {
			deleteIndexedFields := make([]string, len(*sqlOpts.DeleteIndexedFields))
			copy(deleteIndexedFields, *sqlOpts.DeleteIndexedFields)
			opts[utils.SQLDeleteIndexedFieldsOpt] = deleteIndexedFields
		}
		if sqlOpts.PgSSLMode != nil {
			opts[utils.PgSSLModeCfg] = *sqlOpts.PgSSLMode
		}
	}

	if awsOpts := er.Opts.AWS; awsOpts != nil {
		if awsOpts.Region != nil {
			opts[utils.AWSRegion] = *awsOpts.Region
		}
		if awsOpts.Key != nil {
			opts[utils.AWSKey] = *awsOpts.Key
		}
		if awsOpts.Secret != nil {
			opts[utils.AWSSecret] = *awsOpts.Secret
		}
		if awsOpts.Token != nil {
			opts[utils.AWSToken] = *awsOpts.Token
		}
		if awsOpts.SQSQueueID != nil {
			opts[utils.SQSQueueID] = *awsOpts.SQSQueueID
		}
		if awsOpts.S3BucketID != nil {
			opts[utils.S3Bucket] = *awsOpts.S3BucketID
		}
	}

	if natsOpts := er.Opts.NATS; natsOpts != nil {
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

	if len(er.EEsIDs) != 0 {
		initialMP[utils.EEsIDsCfg] = er.EEsIDs
	}
	if len(er.EEsSuccessIDs) != 0 {
		initialMP[utils.EEsSuccessIDsCfg] = er.EEsSuccessIDs
	}
	if len(er.EEsFailedIDs) != 0 {
		initialMP[utils.EEsFailedIDsCfg] = er.EEsFailedIDs
	}

	if er.MaxReconnectInterval != 0 {
		initialMP[utils.MaxReconnectIntervalCfg] = er.MaxReconnectInterval.String()
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
	if er.StartDelay > 0 {
		initialMP[utils.StartDelayCfg] = er.StartDelay.String()
	}
	return
}
