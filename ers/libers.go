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

	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func getProcessOptions(opts *config.EventReaderOpts) (proc *config.EventExporterOpts, populated bool) {
	proc = &config.EventExporterOpts{}
	if opts.AMQPExchangeProcessed != nil {
		proc.AMQPExchange = opts.AMQPExchangeProcessed
		populated = true
	}
	if opts.AMQPExchangeTypeProcessed != nil {
		proc.AMQPExchangeType = opts.AMQPExchangeTypeProcessed
		populated = true
	}
	if opts.AMQPQueueIDProcessed != nil {
		proc.AMQPQueueID = opts.AMQPQueueIDProcessed
		populated = true
	}
	if opts.AMQPRoutingKeyProcessed != nil {
		proc.AMQPRoutingKey = opts.AMQPRoutingKeyProcessed
		populated = true
	}
	if opts.AWSKeyProcessed != nil {
		proc.AWSKey = opts.AWSKeyProcessed
		populated = true
	}
	if opts.AWSRegionProcessed != nil {
		proc.AWSRegion = opts.AWSRegionProcessed
		populated = true
	}
	if opts.AWSSecretProcessed != nil {
		proc.AWSSecret = opts.AWSSecretProcessed
		populated = true
	}
	if opts.AWSTokenProcessed != nil {
		proc.AWSToken = opts.AWSTokenProcessed
		populated = true
	}
	if opts.KafkaTopicProcessed != nil {
		proc.KafkaTopic = opts.KafkaTopicProcessed
		populated = true
	}
	if opts.NATSCertificateAuthorityProcessed != nil {
		proc.NATSCertificateAuthority = opts.NATSCertificateAuthorityProcessed
		populated = true
	}
	if opts.NATSClientCertificateProcessed != nil {
		proc.NATSClientCertificate = opts.NATSClientCertificateProcessed
		populated = true
	}
	if opts.NATSClientKeyProcessed != nil {
		proc.NATSClientKey = opts.NATSClientKeyProcessed
		populated = true
	}
	if opts.NATSJWTFileProcessed != nil {
		proc.NATSJWTFile = opts.NATSJWTFileProcessed
		populated = true
	}
	if opts.NATSJetStreamMaxWaitProcessed != nil {
		proc.NATSJetStreamMaxWait = opts.NATSJetStreamMaxWaitProcessed
		populated = true
	}
	if opts.NATSJetStreamProcessed != nil {
		proc.NATSJetStream = opts.NATSJetStreamProcessed
		populated = true
	}
	if opts.NATSSeedFileProcessed != nil {
		proc.NATSSeedFile = opts.NATSSeedFileProcessed
		populated = true
	}
	if opts.NATSSubjectProcessed != nil {
		proc.NATSSubject = opts.NATSSubjectProcessed
		populated = true
	}
	if opts.S3BucketIDProcessed != nil {
		proc.S3BucketID = opts.S3BucketIDProcessed
		populated = true
	}
	if opts.S3FolderPathProcessed != nil {
		proc.S3FolderPath = opts.S3FolderPathProcessed
		populated = true
	}
	if opts.SQLDBNameProcessed != nil {
		proc.SQLDBName = opts.SQLDBNameProcessed
		populated = true
	}
	if opts.SQLTableNameProcessed != nil {
		proc.SQLTableName = opts.SQLTableNameProcessed
		populated = true
	}
	if opts.SQSQueueIDProcessed != nil {
		proc.SQSQueueID = opts.SQSQueueIDProcessed
		populated = true
	}
	if opts.SSLModeProcessed != nil {
		proc.SSLMode = opts.SSLModeProcessed
		populated = true
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
