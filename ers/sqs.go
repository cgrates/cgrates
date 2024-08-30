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
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// NewSQSER return a new sqs event reader
func NewSQSER(cfg *config.CGRConfig, cfgIdx int, rdrEvents, partialEvents chan *erEvent, rdrErr chan error,
	fltrS *engine.FilterS, rdrExit chan struct{}) (EventReader, error) {
	rdr := &SQSER{
		cgrCfg:        cfg,
		cfgIdx:        cfgIdx,
		fltrS:         fltrS,
		rdrEvents:     rdrEvents,
		partialEvents: partialEvents,
		rdrExit:       rdrExit,
		rdrErr:        rdrErr,
	}
	if concReq := rdr.Config().ConcurrentReqs; concReq != -1 {
		rdr.cap = make(chan struct{}, concReq)
		for i := 0; i < concReq; i++ {
			rdr.cap <- struct{}{}
		}
	}
	rdr.parseOpts(rdr.Config().Opts)
	return rdr, nil
}

// SQSER implements EventReader interface for sqs message
type SQSER struct {
	// sync.RWMutex
	cgrCfg *config.CGRConfig
	cfgIdx int // index of config instance within ERsCfg.Readers
	fltrS  *engine.FilterS

	rdrEvents     chan *erEvent // channel to dispatch the events created to
	partialEvents chan *erEvent // channel to dispatch the partial events created to
	rdrExit       chan struct{}
	rdrErr        chan error
	cap           chan struct{}

	queueURL  *string
	awsRegion string
	awsID     string
	awsKey    string
	awsToken  string
	queueID   string
	session   *session.Session
}

type sqsClient interface {
	ReceiveMessage(input *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error)
	DeleteMessage(input *sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error)
	GetQueueUrl(input *sqs.GetQueueUrlInput) (*sqs.GetQueueUrlOutput, error)
	CreateQueue(input *sqs.CreateQueueInput) (*sqs.CreateQueueOutput, error)
}

// Config returns the curent configuration
func (rdr *SQSER) Config() *config.EventReaderCfg {
	return rdr.cgrCfg.ERsCfg().Readers[rdr.cfgIdx]
}

// Serve will start the gorutines needed to watch the sqs topic
func (rdr *SQSER) Serve() (err error) {
	if rdr.Config().RunDelay == time.Duration(0) { // 0 disables the automatic read, maybe done per API
		return
	}
	go rdr.readLoop(sqs.New(rdr.session)) // read until the connection is closed
	return
}

func (rdr *SQSER) processMessage(body []byte) (err error) {
	var decodedMessage map[string]any
	if err = json.Unmarshal(body, &decodedMessage); err != nil {
		return
	}
	reqVars := &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{utils.MetaReaderID: utils.NewLeafNode(rdr.cgrCfg.ERsCfg().Readers[rdr.cfgIdx].ID)}}
	agReq := agents.NewAgentRequest(
		utils.MapStorage(decodedMessage), reqVars,
		nil, nil, nil, rdr.Config().Tenant,
		rdr.cgrCfg.GeneralCfg().DefaultTenant,
		utils.FirstNonEmpty(rdr.Config().Timezone,
			rdr.cgrCfg.GeneralCfg().DefaultTimezone),
		rdr.fltrS, nil) // create an AgentRequest
	var pass bool
	if pass, err = rdr.fltrS.Pass(context.TODO(), agReq.Tenant, rdr.Config().Filters,
		agReq); err != nil || !pass {
		return
	}
	if err = agReq.SetFields(rdr.Config().Fields); err != nil {
		return
	}
	cgrEv := utils.NMAsCGREvent(agReq.CGRRequest, agReq.Tenant, utils.NestingSep, agReq.Opts)
	rdrEv := rdr.rdrEvents
	cgrEv.APIOpts = make(map[string]any)
	if _, isPartial := cgrEv.APIOpts[utils.PartialOpt]; isPartial {
		rdrEv = rdr.partialEvents
	}
	rdrEv <- &erEvent{
		cgrEvent: cgrEv,
		rdrCfg:   rdr.Config(),
	}
	return
}

func (rdr *SQSER) parseOpts(opts *config.EventReaderOpts) {
	rdr.queueID = utils.DefaultQueueID
	if opts.SQSQueueID != nil {
		rdr.queueID = *opts.SQSQueueID
	}
	if opts.AWSRegion != nil {
		rdr.awsRegion = *opts.AWSRegion
	}
	if opts.AWSKey != nil {
		rdr.awsID = *opts.AWSKey
	}
	if opts.AWSSecret != nil {
		rdr.awsKey = *opts.AWSSecret
	}
	if opts.AWSToken != nil {
		rdr.awsToken = *opts.AWSToken
	}
	rdr.getQueueURL()
}

func (rdr *SQSER) getQueueURL() (err error) {
	if rdr.queueURL != nil {
		return nil
	}
	if err = rdr.newSession(); err != nil {
		return
	}
	return rdr.getQueueURLWithClient(sqs.New(rdr.session))
}

func (rdr *SQSER) getQueueURLWithClient(svc sqsClient) (err error) {
	var result *sqs.GetQueueUrlOutput
	if result, err = svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(rdr.queueID),
	}); err == nil {
		rdr.queueURL = new(string)
		*rdr.queueURL = *result.QueueUrl
		return
	}
	if aerr, ok := err.(awserr.Error); ok && aerr.Code() == sqs.ErrCodeQueueDoesNotExist {
		// For CreateQueue
		var createResult *sqs.CreateQueueOutput
		if createResult, err = svc.CreateQueue(&sqs.CreateQueueInput{
			QueueName: aws.String(rdr.queueID),
		}); err == nil {
			rdr.queueURL = utils.StringPointer(*createResult.QueueUrl)
			return
		}
	}
	return
}

func (rdr *SQSER) readLoop(scv sqsClient) (err error) {
	// scv := sqs.New(rdr.session)
	for !rdr.isClosed() {
		if rdr.Config().ConcurrentReqs != -1 {
			<-rdr.cap // do not try to read if the limit is reached
		}
		var msgs *sqs.ReceiveMessageOutput
		if msgs, err = scv.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl:            rdr.queueURL,
			MaxNumberOfMessages: aws.Int64(1),
			WaitTimeSeconds:     aws.Int64(1),
		}); err != nil {
			return
		}
		if len(msgs.Messages) != 0 {
			go rdr.readMsg(scv, msgs.Messages[0])
		} else if rdr.Config().ConcurrentReqs != -1 {
			rdr.cap <- struct{}{}
		}

	}

	return
}

func (rdr *SQSER) isClosed() bool {
	select {
	case <-rdr.rdrExit:
		return true
	default:
		return false
	}
}

func (rdr *SQSER) readMsg(scv sqsClient, msg *sqs.Message) (err error) {
	if rdr.Config().ConcurrentReqs != -1 {
		defer func() { rdr.cap <- struct{}{} }()
	}
	body := []byte(*msg.Body)
	key := *msg.MessageId
	if err = rdr.processMessage(body); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> processing message %s error: %s",
				utils.ERs, key, err.Error()))
		return
	}
	if _, err = scv.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      rdr.queueURL,
		ReceiptHandle: msg.ReceiptHandle,
	}); err != nil {
		rdr.rdrErr <- err
		return
	}
	return
}

func (rdr *SQSER) newSession() (err error) {
	cfg := aws.Config{Endpoint: aws.String(rdr.Config().SourcePath)}
	if len(rdr.awsRegion) != 0 {
		cfg.Region = aws.String(rdr.awsRegion)
	}
	if len(rdr.awsID) != 0 &&
		len(rdr.awsKey) != 0 {
		cfg.Credentials = credentials.NewStaticCredentials(rdr.awsID, rdr.awsKey, rdr.awsToken)
	}
	rdr.session, err = session.NewSessionWithOptions(
		session.Options{
			Config: cfg,
		},
	)
	return
}
