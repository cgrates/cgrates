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

	if erOpts.AMQPOpts != nil {
		initAMQPExporterOpts := func() {
			if eeOpts == nil {
				eeOpts = new(config.EventExporterOpts)
			}
			if eeOpts.AMQP == nil {
				eeOpts.AMQP = new(config.AMQPOpts)
			}
		}

		if erOpts.AMQPOpts.AMQPExchangeProcessed != nil {
			initAMQPExporterOpts()
			eeOpts.AMQP.Exchange = erOpts.AMQPOpts.AMQPExchangeProcessed
		}
		if erOpts.AMQPOpts.AMQPExchangeTypeProcessed != nil {
			initAMQPExporterOpts()
			eeOpts.AMQP.ExchangeType = erOpts.AMQPOpts.AMQPExchangeTypeProcessed
		}
		if erOpts.AMQPOpts.AMQPQueueIDProcessed != nil {
			initAMQPExporterOpts()
			eeOpts.AMQP.QueueID = erOpts.AMQPOpts.AMQPQueueIDProcessed
		}
		if erOpts.AMQPOpts.AMQPRoutingKeyProcessed != nil {
			initAMQPExporterOpts()
			eeOpts.AMQP.RoutingKey = erOpts.AMQPOpts.AMQPRoutingKeyProcessed
		}
		if erOpts.AMQPOpts.AMQPUsernameProcessed != nil {
			initAMQPExporterOpts()
			eeOpts.AMQP.Username = erOpts.AMQPOpts.AMQPUsernameProcessed
		}
		if erOpts.AMQPOpts.AMQPPasswordProcessed != nil {
			initAMQPExporterOpts()
			eeOpts.AMQP.Password = erOpts.AMQPOpts.AMQPPasswordProcessed
		}
	}

	if erOpts.AWSOpts != nil {
		initAWSExporterOpts := func() {
			if eeOpts == nil {
				eeOpts = new(config.EventExporterOpts)
			}
			if eeOpts.AWS == nil {
				eeOpts.AWS = new(config.AWSOpts)
			}
		}

		if erOpts.AWSOpts.AWSKeyProcessed != nil {
			initAWSExporterOpts()
			eeOpts.AWS.Key = erOpts.AWSOpts.AWSKeyProcessed
		}
		if erOpts.AWSOpts.AWSRegionProcessed != nil {
			initAWSExporterOpts()
			eeOpts.AWS.Region = erOpts.AWSOpts.AWSRegionProcessed
		}
		if erOpts.AWSOpts.AWSSecretProcessed != nil {
			initAWSExporterOpts()
			eeOpts.AWS.Secret = erOpts.AWSOpts.AWSSecretProcessed
		}
		if erOpts.AWSOpts.AWSTokenProcessed != nil {
			initAWSExporterOpts()
			eeOpts.AWS.Token = erOpts.AWSOpts.AWSTokenProcessed
		}
		if erOpts.AWSOpts.S3BucketIDProcessed != nil {
			initAWSExporterOpts()
			eeOpts.AWS.S3BucketID = erOpts.AWSOpts.S3BucketIDProcessed
		}
		if erOpts.AWSOpts.S3FolderPathProcessed != nil {
			initAWSExporterOpts()
			eeOpts.AWS.S3FolderPath = erOpts.AWSOpts.S3FolderPathProcessed
		}
		if erOpts.AWSOpts.SQSQueueIDProcessed != nil {
			initAWSExporterOpts()
			eeOpts.AWS.SQSQueueID = erOpts.AWSOpts.SQSQueueIDProcessed
		}
	}

	if erOpts.KafkaOpts != nil {
		if erOpts.KafkaOpts.KafkaTopicProcessed != nil {
			if eeOpts == nil {
				eeOpts = new(config.EventExporterOpts)
			}
			if eeOpts.Kafka == nil {
				eeOpts.Kafka = new(config.KafkaOpts)
			}
			eeOpts.Kafka.KafkaTopic = erOpts.KafkaOpts.KafkaTopicProcessed
		}
	}

	if erOpts.NATSOpts != nil {
		initNATSExporterOpts := func() {
			if eeOpts == nil {
				eeOpts = new(config.EventExporterOpts)
			}
			if eeOpts.NATS == nil {
				eeOpts.NATS = new(config.NATSOpts)
			}
		}

		if erOpts.NATSOpts.CertificateAuthorityProcessed != nil {
			initNATSExporterOpts()
			eeOpts.NATS.CertificateAuthority = erOpts.NATSOpts.CertificateAuthorityProcessed
		}
		if erOpts.NATSOpts.ClientCertificateProcessed != nil {
			initNATSExporterOpts()
			eeOpts.NATS.ClientCertificate = erOpts.NATSOpts.ClientCertificateProcessed
		}
		if erOpts.NATSOpts.ClientKeyProcessed != nil {
			initNATSExporterOpts()
			eeOpts.NATS.ClientKey = erOpts.NATSOpts.ClientKeyProcessed
		}
		if erOpts.NATSOpts.JWTFileProcessed != nil {
			initNATSExporterOpts()
			eeOpts.NATS.JWTFile = erOpts.NATSOpts.JWTFileProcessed
		}
		if erOpts.NATSOpts.JetStreamMaxWaitProcessed != nil {
			initNATSExporterOpts()
			eeOpts.NATS.JetStreamMaxWait = erOpts.NATSOpts.JetStreamMaxWaitProcessed
		}
		if erOpts.NATSOpts.JetStreamProcessed != nil {
			initNATSExporterOpts()
			eeOpts.NATS.JetStream = erOpts.NATSOpts.JetStreamProcessed
		}
		if erOpts.NATSOpts.SeedFileProcessed != nil {
			initNATSExporterOpts()
			eeOpts.NATS.SeedFile = erOpts.NATSOpts.SeedFileProcessed
		}
		if erOpts.NATSOpts.SubjectProcessed != nil {
			initNATSExporterOpts()
			eeOpts.NATS.Subject = erOpts.NATSOpts.SubjectProcessed
		}
	}

	if erOpts.SQLOpts != nil {
		initSQLExporterOpts := func() {
			if eeOpts == nil {
				eeOpts = new(config.EventExporterOpts)
			}
			if eeOpts.SQL == nil {
				eeOpts.SQL = new(config.SQLOpts)
			}
		}

		if erOpts.SQLOpts.SQLDBNameProcessed != nil {
			initSQLExporterOpts()
			eeOpts.SQL.DBName = erOpts.SQLOpts.SQLDBNameProcessed
		}
		if erOpts.SQLOpts.SQLTableNameProcessed != nil {
			initSQLExporterOpts()
			eeOpts.SQL.TableName = erOpts.SQLOpts.SQLTableNameProcessed
		}
		if erOpts.SQLOpts.PgSSLModeProcessed != nil {
			initSQLExporterOpts()
			eeOpts.SQL.PgSSLMode = erOpts.SQLOpts.PgSSLModeProcessed
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
