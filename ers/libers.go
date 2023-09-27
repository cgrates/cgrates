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

package ers

import (
	"fmt"
	"sort"
	"time"

	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// getProcessedOptions assigns all non-nil fields ending in "Processed" from EventReaderOpts to their counterparts in EventExporterOpts
func getProcessedOptions(erOpts *config.EventReaderOpts) *config.EventExporterOpts {
	var eeOpts *config.EventExporterOpts

	if erOpts.AMQP != nil {
		initAMQPExporterOpts := func() {
			if eeOpts == nil {
				eeOpts = new(config.EventExporterOpts)
			}
			if eeOpts.AMQP == nil {
				eeOpts.AMQP = new(config.AMQPOpts)
			}
		}

		if erOpts.AMQP.ExchangeProcessed != nil {
			initAMQPExporterOpts()
			eeOpts.AMQP.Exchange = erOpts.AMQP.ExchangeProcessed
		}
		if erOpts.AMQP.ExchangeTypeProcessed != nil {
			initAMQPExporterOpts()
			eeOpts.AMQP.ExchangeType = erOpts.AMQP.ExchangeTypeProcessed
		}
		if erOpts.AMQP.QueueIDProcessed != nil {
			initAMQPExporterOpts()
			eeOpts.AMQP.QueueID = erOpts.AMQP.QueueIDProcessed
		}
		if erOpts.AMQP.RoutingKeyProcessed != nil {
			initAMQPExporterOpts()
			eeOpts.AMQP.RoutingKey = erOpts.AMQP.RoutingKeyProcessed
		}
		if erOpts.AMQP.UsernameProcessed != nil {
			initAMQPExporterOpts()
			eeOpts.AMQP.Username = erOpts.AMQP.UsernameProcessed
		}
		if erOpts.AMQP.PasswordProcessed != nil {
			initAMQPExporterOpts()
			eeOpts.AMQP.Password = erOpts.AMQP.PasswordProcessed
		}
	}

	if erOpts.AWS != nil {
		initAWSExporterOpts := func() {
			if eeOpts == nil {
				eeOpts = new(config.EventExporterOpts)
			}
			if eeOpts.AWS == nil {
				eeOpts.AWS = new(config.AWSOpts)
			}
		}

		if erOpts.AWS.KeyProcessed != nil {
			initAWSExporterOpts()
			eeOpts.AWS.Key = erOpts.AWS.KeyProcessed
		}
		if erOpts.AWS.RegionProcessed != nil {
			initAWSExporterOpts()
			eeOpts.AWS.Region = erOpts.AWS.RegionProcessed
		}
		if erOpts.AWS.SecretProcessed != nil {
			initAWSExporterOpts()
			eeOpts.AWS.Secret = erOpts.AWS.SecretProcessed
		}
		if erOpts.AWS.TokenProcessed != nil {
			initAWSExporterOpts()
			eeOpts.AWS.Token = erOpts.AWS.TokenProcessed
		}
		if erOpts.AWS.S3BucketIDProcessed != nil {
			initAWSExporterOpts()
			eeOpts.AWS.S3BucketID = erOpts.AWS.S3BucketIDProcessed
		}
		if erOpts.AWS.S3FolderPathProcessed != nil {
			initAWSExporterOpts()
			eeOpts.AWS.S3FolderPath = erOpts.AWS.S3FolderPathProcessed
		}
		if erOpts.AWS.SQSQueueIDProcessed != nil {
			initAWSExporterOpts()
			eeOpts.AWS.SQSQueueID = erOpts.AWS.SQSQueueIDProcessed
		}
	}

	if erOpts.Kafka != nil {
		if erOpts.Kafka.TopicProcessed != nil {
			if eeOpts == nil {
				eeOpts = new(config.EventExporterOpts)
			}
			if eeOpts.Kafka == nil {
				eeOpts.Kafka = new(config.KafkaOpts)
			}
			eeOpts.Kafka.KafkaTopic = erOpts.Kafka.TopicProcessed
		}
	}

	if erOpts.NATS != nil {
		initNATSExporterOpts := func() {
			if eeOpts == nil {
				eeOpts = new(config.EventExporterOpts)
			}
			if eeOpts.NATS == nil {
				eeOpts.NATS = new(config.NATSOpts)
			}
		}

		if erOpts.NATS.CertificateAuthorityProcessed != nil {
			initNATSExporterOpts()
			eeOpts.NATS.CertificateAuthority = erOpts.NATS.CertificateAuthorityProcessed
		}
		if erOpts.NATS.ClientCertificateProcessed != nil {
			initNATSExporterOpts()
			eeOpts.NATS.ClientCertificate = erOpts.NATS.ClientCertificateProcessed
		}
		if erOpts.NATS.ClientKeyProcessed != nil {
			initNATSExporterOpts()
			eeOpts.NATS.ClientKey = erOpts.NATS.ClientKeyProcessed
		}
		if erOpts.NATS.JWTFileProcessed != nil {
			initNATSExporterOpts()
			eeOpts.NATS.JWTFile = erOpts.NATS.JWTFileProcessed
		}
		if erOpts.NATS.JetStreamMaxWaitProcessed != nil {
			initNATSExporterOpts()
			eeOpts.NATS.JetStreamMaxWait = erOpts.NATS.JetStreamMaxWaitProcessed
		}
		if erOpts.NATS.JetStreamProcessed != nil {
			initNATSExporterOpts()
			eeOpts.NATS.JetStream = erOpts.NATS.JetStreamProcessed
		}
		if erOpts.NATS.SeedFileProcessed != nil {
			initNATSExporterOpts()
			eeOpts.NATS.SeedFile = erOpts.NATS.SeedFileProcessed
		}
		if erOpts.NATS.SubjectProcessed != nil {
			initNATSExporterOpts()
			eeOpts.NATS.Subject = erOpts.NATS.SubjectProcessed
		}
	}

	if erOpts.SQL != nil {
		initSQLExporterOpts := func() {
			if eeOpts == nil {
				eeOpts = new(config.EventExporterOpts)
			}
			if eeOpts.SQL == nil {
				eeOpts.SQL = new(config.SQLOpts)
			}
		}

		if erOpts.SQL.DBNameProcessed != nil {
			initSQLExporterOpts()
			eeOpts.SQL.DBName = erOpts.SQL.DBNameProcessed
		}
		if erOpts.SQL.TableNameProcessed != nil {
			initSQLExporterOpts()
			eeOpts.SQL.TableName = erOpts.SQL.TableNameProcessed
		}
		if erOpts.SQL.PgSSLModeProcessed != nil {
			initSQLExporterOpts()
			eeOpts.SQL.PgSSLMode = erOpts.SQL.PgSSLModeProcessed
		}
	}

	return eeOpts
}

// mergePartialEvents will unite the events using the reader configuration
func mergePartialEvents(cgrEvs []*utils.CGREvent, cfg *config.EventReaderCfg, fltrS *engine.FilterS, dftTnt, dftTmz, rsrSep string) (cgrEv *utils.CGREvent, err error) {
	cgrEv = cgrEvs[0]     // by default there is at least one event
	if len(cgrEvs) != 1 { // need to merge the incoming events
		// prepare the field after which the events are ordered
		var ordFld string
		if cfg.Opts.PartialOrderField != nil {
			ordFld = *cfg.Opts.PartialOrderField
		}
		var ordPath config.RSRParsers
		if ordPath, err = config.NewRSRParsers(ordFld, rsrSep); err != nil { // convert the option to rsrParsers
			return nil, err
		}

		// get the field as interface in a slice
		fields := make([]any, len(cgrEvs))
		for i, ev := range cgrEvs {
			if fields[i], err = ordPath.ParseDataProviderWithInterfaces(ev.AsDataProvider()); err != nil {
				return
			}
			if fldStr, castStr := fields[i].(string); castStr { // attempt converting string since deserialization fails here (ie: time.Time fields)
				fields[i] = utils.StringToInterface(fldStr)
			}
		}
		//sort CGREvents based on partialOrderFieldOpt
		sort.Slice(cgrEvs, func(i, j int) bool {
			gt, serr := utils.GreaterThan(fields[i], fields[j], true)
			if serr != nil { // save the last non nil error
				err = serr
			}
			return !gt
		})
		if err != nil { // the fields are not comparable
			return
		}

		// compose the CGREvent from slice
		cgrEv = &utils.CGREvent{
			Tenant:  cgrEvs[0].Tenant,
			ID:      utils.UUIDSha1Prefix(),
			Time:    utils.TimePointer(time.Now()),
			Event:   make(map[string]any),
			APIOpts: make(map[string]any),
		}
		for _, ev := range cgrEvs { // merge the maps
			for key, value := range ev.Event {
				cgrEv.Event[key] = value
			}
			for key, val := range ev.APIOpts {
				cgrEv.APIOpts[key] = val
			}
		}
	}
	if len(cfg.PartialCommitFields) != 0 { // apply the partial commit template
		agReq := agents.NewAgentRequest(
			utils.MapStorage(cgrEv.Event), nil,
			nil, nil, cgrEv.APIOpts, cfg.Tenant, dftTnt,
			utils.FirstNonEmpty(cfg.Timezone, dftTmz),
			fltrS, nil) // create an AgentRequest
		if err = agReq.SetFields(cfg.PartialCommitFields); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> processing partial event: <%s>, ignoring due to error: <%s>",
					utils.ERs, utils.ToJSON(cgrEv), err.Error()))
			return
		}
		if ev := utils.NMAsCGREvent(agReq.CGRRequest, agReq.Tenant,
			utils.NestingSep, agReq.Opts); ev != nil { // add the modified fields in the event
			for k, v := range ev.Event {
				cgrEv.Event[k] = v
			}
		}
	}
	return
}
