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
	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// NewSQSER return a new sqs event reader
func NewSQSER(cfg *config.CGRConfig, cfgIdx int,
	rdrEvents chan *erEvent, rdrErr chan error,
	fltrS *engine.FilterS, rdrExit chan struct{}) (er EventReader, err error) {

	rdr := &SQSER{
		cgrCfg:    cfg,
		cfgIdx:    cfgIdx,
		fltrS:     fltrS,
		rdrEvents: rdrEvents,
		rdrExit:   rdrExit,
		rdrErr:    rdrErr,
	}
	if concReq := rdr.Config().ConcurrentReqs; concReq != -1 {
		rdr.cap = make(chan struct{}, concReq)
		for i := 0; i < concReq; i++ {
			rdr.cap <- struct{}{}
		}
	}
	rdr.parseOpts(rdr.Config().Opts)
	rdr.createPoster()
	return rdr, nil
}

// SQSER implements EventReader interface for sqs message
type SQSER struct {
	// sync.RWMutex
	cgrCfg *config.CGRConfig
	cfgIdx int // index of config instance within ERsCfg.Readers
	fltrS  *engine.FilterS

	rdrEvents chan *erEvent // channel to dispatch the events created to
	rdrExit   chan struct{}
	rdrErr    chan error
	cap       chan struct{}

	queueURL  *string
	awsRegion string
	awsID     string
	awsKey    string
	awsToken  string
	queueID   string
	session   *session.Session

	poster engine.Poster
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
	var decodedMessage map[string]interface{}
	if err = json.Unmarshal(body, &decodedMessage); err != nil {
		return
	}

	agReq := agents.NewAgentRequest(
		utils.MapStorage(decodedMessage), nil,
		nil, nil, nil, rdr.Config().Tenant,
		rdr.cgrCfg.GeneralCfg().DefaultTenant,
		utils.FirstNonEmpty(rdr.Config().Timezone,
			rdr.cgrCfg.GeneralCfg().DefaultTimezone),
		rdr.fltrS, nil, nil) // create an AgentRequest
	var pass bool
	if pass, err = rdr.fltrS.Pass(agReq.Tenant, rdr.Config().Filters,
		agReq); err != nil || !pass {
		return
	}
	if err = agReq.SetFields(rdr.Config().Fields); err != nil {
		return
	}
	cgrEv := utils.NMAsCGREvent(agReq.CGRRequest, agReq.Tenant, utils.NestingSep, agReq.Opts)
	rdr.rdrEvents <- &erEvent{
		cgrEvent: cgrEv,
		rdrCfg:   rdr.Config(),
	}
	return
}

func (rdr *SQSER) parseOpts(opts map[string]interface{}) {
	rdr.queueID = utils.DefaultQueueID
	if val, has := opts[utils.SQSQueueID]; has {
		rdr.queueID = utils.IfaceAsString(val)
	}
	if val, has := opts[utils.AWSRegion]; has {
		rdr.awsRegion = utils.IfaceAsString(val)
	}
	if val, has := opts[utils.AWSKey]; has {
		rdr.awsID = utils.IfaceAsString(val)
	}
	if val, has := opts[utils.AWSSecret]; has {
		rdr.awsKey = utils.IfaceAsString(val)
	}
	if val, has := opts[utils.AWSToken]; has {
		rdr.awsToken = utils.IfaceAsString(val)
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
		*(rdr.queueURL) = *result.QueueUrl
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
	utils.Logger.Warning(fmt.Sprintf("<SQSPoster> can not get url for queue with ID=%s because err: %v", rdr.queueID, err))
	return
}

func (rdr *SQSER) readLoop(scv sqsClient) (err error) {
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

func (rdr *SQSER) createPoster() {
	processedOpt := getProcessOptions(rdr.Config().Opts)
	if len(processedOpt) == 0 &&
		len(rdr.Config().ProcessedPath) == 0 {
		return
	}
	rdr.poster = engine.NewSQSPoster(utils.FirstNonEmpty(rdr.Config().ProcessedPath, rdr.Config().SourcePath),
		rdr.cgrCfg.GeneralCfg().PosterAttempts, processedOpt)
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

	if rdr.poster != nil { // post it
		if err = rdr.poster.Post(body, key); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> writing message %s error: %s",
					utils.ERs, key, err.Error()))
			return
		}
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
