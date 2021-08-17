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

package ees

import (
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// NewSQSee creates a poster for sqs
func NewSQSee(cfg *config.EventExporterCfg, dc *utils.SafeMapStorage) *SQSee {
	pstr := &SQSee{
		cfg:  cfg,
		dc:   dc,
		reqs: newConcReq(cfg.ConcurrentRequests),
	}
	pstr.parseOpts(cfg.Opts)
	return pstr
}

// SQSee is a poster for sqs
type SQSee struct {
	awsRegion string
	awsID     string
	awsKey    string
	awsToken  string
	queueURL  *string
	queueID   string
	session   *session.Session
	svc       *sqs.SQS

	cfg          *config.EventExporterCfg
	dc           *utils.SafeMapStorage
	reqs         *concReq
	sync.RWMutex // protect connection
	bytePreparing
}

func (pstr *SQSee) parseOpts(opts map[string]interface{}) {
	pstr.queueID = utils.DefaultQueueID
	if val, has := opts[utils.SQSQueueID]; has {
		pstr.queueID = utils.IfaceAsString(val)
	}
	if val, has := opts[utils.AWSRegion]; has {
		pstr.awsRegion = utils.IfaceAsString(val)
	}
	if val, has := opts[utils.AWSKey]; has {
		pstr.awsID = utils.IfaceAsString(val)
	}
	if val, has := opts[utils.AWSSecret]; has {
		pstr.awsKey = utils.IfaceAsString(val)
	}
	if val, has := opts[utils.AWSToken]; has {
		pstr.awsToken = utils.IfaceAsString(val)
	}
}

func (pstr *SQSee) Cfg() *config.EventExporterCfg { return pstr.cfg }

func (pstr *SQSee) Connect() (err error) {
	pstr.Lock()
	defer pstr.Unlock()
	if pstr.session == nil {
		cfg := aws.Config{Endpoint: aws.String(pstr.Cfg().ExportPath)}
		if len(pstr.awsRegion) != 0 {
			cfg.Region = aws.String(pstr.awsRegion)
		}
		if len(pstr.awsID) != 0 &&
			len(pstr.awsKey) != 0 {
			cfg.Credentials = credentials.NewStaticCredentials(pstr.awsID, pstr.awsKey, pstr.awsToken)
		}
		pstr.session, err = session.NewSessionWithOptions(
			session.Options{
				Config: cfg,
			},
		)
		if err != nil {
			return
		}
	}
	if pstr.svc == nil {
		pstr.svc = sqs.New(pstr.session)
	}
	if pstr.queueURL != nil {
		return
	}
	var result *sqs.GetQueueUrlOutput
	if result, err = pstr.svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(pstr.queueID),
	}); err == nil {
		pstr.queueURL = new(string)
		*(pstr.queueURL) = *result.QueueUrl
		return
	}
	if aerr, ok := err.(awserr.Error); ok && aerr.Code() == sqs.ErrCodeQueueDoesNotExist {
		// For CreateQueue
		var createResult *sqs.CreateQueueOutput
		if createResult, err = pstr.svc.CreateQueue(&sqs.CreateQueueInput{
			QueueName: aws.String(pstr.queueID),
		}); err == nil {
			pstr.queueURL = new(string)
			*(pstr.queueURL) = *createResult.QueueUrl
			return
		}
	}
	return
}

func (pstr *SQSee) ExportEvent(message interface{}, _ string) (err error) {
	pstr.reqs.get()
	pstr.RLock()
	_, err = pstr.svc.SendMessage(
		&sqs.SendMessageInput{
			MessageBody: aws.String(string(message.([]byte))),
			QueueUrl:    pstr.queueURL,
		},
	)
	pstr.RUnlock()
	pstr.reqs.done()
	return
}

func (pstr *SQSee) Close() (_ error) { return }

func (pstr *SQSee) GetMetrics() *utils.SafeMapStorage { return pstr.dc }
