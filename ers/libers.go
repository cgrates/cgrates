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

func getProcessOptions(erOpts *config.EventReaderOpts) (eeOpts *config.EventExporterOpts) {
	if erOpts.AMQPExchangeProcessed != nil {
		if eeOpts == nil {
			eeOpts = new(config.EventExporterOpts)
		}
		eeOpts.AMQPExchange = erOpts.AMQPExchangeProcessed
	}
	if erOpts.AMQPExchangeTypeProcessed != nil {
		if eeOpts == nil {
			eeOpts = new(config.EventExporterOpts)
		}
		eeOpts.AMQPExchangeType = erOpts.AMQPExchangeTypeProcessed
	}
	if erOpts.AMQPQueueIDProcessed != nil {
		if eeOpts == nil {
			eeOpts = new(config.EventExporterOpts)
		}
		eeOpts.AMQPQueueID = erOpts.AMQPQueueIDProcessed
	}
	if erOpts.AMQPRoutingKeyProcessed != nil {
		if eeOpts == nil {
			eeOpts = new(config.EventExporterOpts)
		}
		eeOpts.AMQPRoutingKey = erOpts.AMQPRoutingKeyProcessed
	}
	if erOpts.AWSKeyProcessed != nil {
		if eeOpts == nil {
			eeOpts = new(config.EventExporterOpts)
		}
		eeOpts.AWSKey = erOpts.AWSKeyProcessed
	}
	if erOpts.AWSRegionProcessed != nil {
		if eeOpts == nil {
			eeOpts = new(config.EventExporterOpts)
		}
		eeOpts.AWSRegion = erOpts.AWSRegionProcessed
	}
	if erOpts.AWSSecretProcessed != nil {
		if eeOpts == nil {
			eeOpts = new(config.EventExporterOpts)
		}
		eeOpts.AWSSecret = erOpts.AWSSecretProcessed
	}
	if erOpts.AWSTokenProcessed != nil {
		if eeOpts == nil {
			eeOpts = new(config.EventExporterOpts)
		}
		eeOpts.AWSToken = erOpts.AWSTokenProcessed
	}
	if erOpts.KafkaTopicProcessed != nil {
		if eeOpts == nil {
			eeOpts = new(config.EventExporterOpts)
		}
		eeOpts.KafkaTopic = erOpts.KafkaTopicProcessed
	}
	if erOpts.NATSCertificateAuthorityProcessed != nil {
		if eeOpts == nil {
			eeOpts = new(config.EventExporterOpts)
		}
		eeOpts.NATSCertificateAuthority = erOpts.NATSCertificateAuthorityProcessed
	}
	if erOpts.NATSClientCertificateProcessed != nil {
		if eeOpts == nil {
			eeOpts = new(config.EventExporterOpts)
		}
		eeOpts.NATSClientCertificate = erOpts.NATSClientCertificateProcessed
	}
	if erOpts.NATSClientKeyProcessed != nil {
		if eeOpts == nil {
			eeOpts = new(config.EventExporterOpts)
		}
		eeOpts.NATSClientKey = erOpts.NATSClientKeyProcessed
	}
	if erOpts.NATSJWTFileProcessed != nil {
		if eeOpts == nil {
			eeOpts = new(config.EventExporterOpts)
		}
		eeOpts.NATSJWTFile = erOpts.NATSJWTFileProcessed
	}
	if erOpts.NATSJetStreamMaxWaitProcessed != nil {
		if eeOpts == nil {
			eeOpts = new(config.EventExporterOpts)
		}
		eeOpts.NATSJetStreamMaxWait = erOpts.NATSJetStreamMaxWaitProcessed
	}
	if erOpts.NATSJetStreamProcessed != nil {
		if eeOpts == nil {
			eeOpts = new(config.EventExporterOpts)
		}
		eeOpts.NATSJetStream = erOpts.NATSJetStreamProcessed
	}
	if erOpts.NATSSeedFileProcessed != nil {
		if eeOpts == nil {
			eeOpts = new(config.EventExporterOpts)
		}
		eeOpts.NATSSeedFile = erOpts.NATSSeedFileProcessed
	}
	if erOpts.NATSSubjectProcessed != nil {
		if eeOpts == nil {
			eeOpts = new(config.EventExporterOpts)
		}
		eeOpts.NATSSubject = erOpts.NATSSubjectProcessed
	}
	if erOpts.S3BucketIDProcessed != nil {
		if eeOpts == nil {
			eeOpts = new(config.EventExporterOpts)
		}
		eeOpts.S3BucketID = erOpts.S3BucketIDProcessed
	}
	if erOpts.S3FolderPathProcessed != nil {
		if eeOpts == nil {
			eeOpts = new(config.EventExporterOpts)
		}
		eeOpts.S3FolderPath = erOpts.S3FolderPathProcessed
	}
	if erOpts.SQLDBNameProcessed != nil {
		if eeOpts == nil {
			eeOpts = new(config.EventExporterOpts)
		}
		eeOpts.SQLDBName = erOpts.SQLDBNameProcessed
	}
	if erOpts.SQLTableNameProcessed != nil {
		if eeOpts == nil {
			eeOpts = new(config.EventExporterOpts)
		}
		eeOpts.SQLTableName = erOpts.SQLTableNameProcessed
	}
	if erOpts.SQSQueueIDProcessed != nil {
		if eeOpts == nil {
			eeOpts = new(config.EventExporterOpts)
		}
		eeOpts.SQSQueueID = erOpts.SQSQueueIDProcessed
	}
	if erOpts.PgSSLModeProcessed != nil {
		if eeOpts == nil {
			eeOpts = new(config.EventExporterOpts)
		}
		eeOpts.PgSSLMode = erOpts.PgSSLModeProcessed
	}
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
		fields := make([]interface{}, len(cgrEvs))
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
			Event:   make(map[string]interface{}),
			APIOpts: make(map[string]interface{}),
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
