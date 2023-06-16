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

// getProcessOptions assigns all non-nil fields ending in "Processed" from EventReaderOpts to their counterparts in EventExporterOpts
func getProcessOptions(erOpts *config.EventReaderOpts) (eeOpts *config.EventExporterOpts) {
	eeOpts = new(config.EventExporterOpts)
	if erOpts.AMQPExchangeProcessed != nil {
		if eeOpts.AMQP == nil {
			eeOpts.AMQP = new(config.AMQPOpts)
		}
		eeOpts.AMQP.Exchange = erOpts.AMQPExchangeProcessed
	}
	if erOpts.AMQPExchangeTypeProcessed != nil {
		if eeOpts.AMQP == nil {
			eeOpts.AMQP = new(config.AMQPOpts)
		}
		eeOpts.AMQP.ExchangeType = erOpts.AMQPExchangeTypeProcessed
	}
	if erOpts.AMQPQueueIDProcessed != nil {
		if eeOpts.AMQP == nil {
			eeOpts.AMQP = new(config.AMQPOpts)
		}
		eeOpts.AMQP.QueueID = erOpts.AMQPQueueIDProcessed
	}
	if erOpts.AMQPRoutingKeyProcessed != nil {
		if eeOpts.AMQP == nil {
			eeOpts.AMQP = new(config.AMQPOpts)
		}
		eeOpts.AMQP.RoutingKey = erOpts.AMQPRoutingKeyProcessed
	}
	if erOpts.AMQPUsernameProcessed != nil {
		if eeOpts.AMQP == nil {
			eeOpts.AMQP = new(config.AMQPOpts)
		}
		eeOpts.AMQP.Username = erOpts.AMQPUsernameProcessed
	}
	if erOpts.AMQPPasswordProcessed != nil {
		if eeOpts.AMQP == nil {
			eeOpts.AMQP = new(config.AMQPOpts)
		}
		eeOpts.AMQP.Password = erOpts.AMQPPasswordProcessed
	}
	if erOpts.AWSKeyProcessed != nil {
		if eeOpts.AWS == nil {
			eeOpts.AWS = new(config.AWSOpts)
		}
		eeOpts.AWS.Key = erOpts.AWSKeyProcessed
	}
	if erOpts.AWSRegionProcessed != nil {
		if eeOpts.AWS == nil {
			eeOpts.AWS = new(config.AWSOpts)
		}
		eeOpts.AWS.Region = erOpts.AWSRegionProcessed
	}
	if erOpts.AWSSecretProcessed != nil {
		if eeOpts.AWS == nil {
			eeOpts.AWS = new(config.AWSOpts)
		}
		eeOpts.AWS.Secret = erOpts.AWSSecretProcessed
	}
	if erOpts.AWSTokenProcessed != nil {
		if eeOpts.AWS == nil {
			eeOpts.AWS = new(config.AWSOpts)
		}
		eeOpts.AWS.Token = erOpts.AWSTokenProcessed
	}
	if erOpts.KafkaTopicProcessed != nil {
		if eeOpts.Kafka == nil {
			eeOpts.Kafka = new(config.KafkaOpts)
		}
		eeOpts.Kafka.KafkaTopic = erOpts.KafkaTopicProcessed
	}
	if erOpts.NATSCertificateAuthorityProcessed != nil {
		if eeOpts.NATS == nil {
			eeOpts.NATS = new(config.NATSOpts)
		}
		eeOpts.NATS.CertificateAuthority = erOpts.NATSCertificateAuthorityProcessed
	}
	if erOpts.NATSClientCertificateProcessed != nil {
		if eeOpts.NATS == nil {
			eeOpts.NATS = new(config.NATSOpts)
		}
		eeOpts.NATS.ClientCertificate = erOpts.NATSClientCertificateProcessed
	}
	if erOpts.NATSClientKeyProcessed != nil {
		if eeOpts.NATS == nil {
			eeOpts.NATS = new(config.NATSOpts)
		}
		eeOpts.NATS.ClientKey = erOpts.NATSClientKeyProcessed
	}
	if erOpts.NATSJWTFileProcessed != nil {
		if eeOpts.NATS == nil {
			eeOpts.NATS = new(config.NATSOpts)
		}
		eeOpts.NATS.JWTFile = erOpts.NATSJWTFileProcessed
	}
	if erOpts.NATSJetStreamMaxWaitProcessed != nil {
		if eeOpts.NATS == nil {
			eeOpts.NATS = new(config.NATSOpts)
		}
		eeOpts.NATS.JetStreamMaxWait = erOpts.NATSJetStreamMaxWaitProcessed
	}
	if erOpts.NATSJetStreamProcessed != nil {
		if eeOpts.NATS == nil {
			eeOpts.NATS = new(config.NATSOpts)
		}
		eeOpts.NATS.JetStream = erOpts.NATSJetStreamProcessed
	}
	if erOpts.NATSSeedFileProcessed != nil {
		if eeOpts.NATS == nil {
			eeOpts.NATS = new(config.NATSOpts)
		}
		eeOpts.NATS.SeedFile = erOpts.NATSSeedFileProcessed
	}
	if erOpts.NATSSubjectProcessed != nil {
		if eeOpts.NATS == nil {
			eeOpts.NATS = new(config.NATSOpts)
		}
		eeOpts.NATS.Subject = erOpts.NATSSubjectProcessed
	}
	if erOpts.S3BucketIDProcessed != nil {
		if eeOpts.AWS == nil {
			eeOpts.AWS = new(config.AWSOpts)
		}
		eeOpts.AWS.S3BucketID = erOpts.S3BucketIDProcessed
	}
	if erOpts.S3FolderPathProcessed != nil {
		if eeOpts.AWS == nil {
			eeOpts.AWS = new(config.AWSOpts)
		}
		eeOpts.AWS.S3FolderPath = erOpts.S3FolderPathProcessed
	}
	if erOpts.SQLDBNameProcessed != nil {
		if eeOpts.SQL == nil {
			eeOpts.SQL = new(config.SQLOpts)
		}
		eeOpts.SQL.DBName = erOpts.SQLDBNameProcessed
	}
	if erOpts.SQLTableNameProcessed != nil {
		if eeOpts.SQL == nil {
			eeOpts.SQL = new(config.SQLOpts)
		}
		eeOpts.SQL.TableName = erOpts.SQLTableNameProcessed
	}
	if erOpts.SQSQueueIDProcessed != nil {
		if eeOpts.AWS == nil {
			eeOpts.AWS = new(config.AWSOpts)
		}
		eeOpts.AWS.SQSQueueID = erOpts.SQSQueueIDProcessed
	}
	if erOpts.PgSSLModeProcessed != nil {
		if eeOpts.SQL == nil {
			eeOpts.SQL = new(config.SQLOpts)
		}
		eeOpts.SQL.PgSSLMode = erOpts.PgSSLModeProcessed
	}
	fmt.Println(utils.ToJSON(eeOpts))
	return
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
