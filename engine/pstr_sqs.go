/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT MetaAny WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package engine

import (
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/cgrates/cgrates/utils"
)

// NewSQSPoster creates a poster for sqs
func NewSQSPoster(dialURL string, attempts int, opts map[string]interface{}) Poster {
	pstr := &SQSPoster{
		attempts: attempts,
	}
	pstr.parseOpts(opts)
	return pstr
}

// SQSPoster is a poster for sqs
type SQSPoster struct {
	sync.Mutex
	dialURL   string
	awsRegion string
	awsID     string
	awsKey    string
	awsToken  string
	attempts  int
	queueURL  *string
	queueID   string
	// getQueueOnce    sync.Once
	session *session.Session
}

// Close for Poster interface
func (pstr *SQSPoster) Close() {}

func (pstr *SQSPoster) parseOpts(opts map[string]interface{}) {
	pstr.queueID = utils.DefaultQueueID
	if val, has := opts[utils.QueueID]; has {
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
	pstr.getQueueURL()
}

func (pstr *SQSPoster) getQueueURL() (err error) {
	if pstr.queueURL != nil {
		return nil
	}
	// pstr.getQueueOnce.Do(func() {
	var svc *sqs.SQS
	if svc, err = pstr.newPosterSession(); err != nil {
		return
	}
	var result *sqs.GetQueueUrlOutput
	if result, err = svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(pstr.queueID),
	}); err == nil {
		pstr.queueURL = new(string)
		*(pstr.queueURL) = *result.QueueUrl
		return
	}
	if aerr, ok := err.(awserr.Error); ok && aerr.Code() == sqs.ErrCodeQueueDoesNotExist {
		// For CreateQueue
		var createResult *sqs.CreateQueueOutput
		if createResult, err = svc.CreateQueue(&sqs.CreateQueueInput{
			QueueName: aws.String(pstr.queueID),
		}); err == nil {
			pstr.queueURL = new(string)
			*(pstr.queueURL) = *createResult.QueueUrl
			return
		}
	}
	utils.Logger.Warning(fmt.Sprintf("<SQSPoster> can not get url for queue with ID=%s because err: %v", pstr.queueID, err))
	// })
	return err
}

// Post is the method being called when we need to post anything in the queue
func (pstr *SQSPoster) Post(message []byte, _ string) (err error) {
	var svc *sqs.SQS
	fib := utils.Fib()

	for i := 0; i < pstr.attempts; i++ {
		if svc, err = pstr.newPosterSession(); err == nil {
			break
		}
		if i+1 < pstr.attempts {
			time.Sleep(time.Duration(fib()) * time.Second)
		}
	}
	if err != nil {
		utils.Logger.Warning(fmt.Sprintf("<SQSPoster> creating new session, err: %s", err.Error()))
		return
	}

	for i := 0; i < pstr.attempts; i++ {
		if _, err = svc.SendMessage(
			&sqs.SendMessageInput{
				MessageBody: aws.String(string(message)),
				QueueUrl:    pstr.queueURL,
			},
		); err == nil {
			break
		}
		if i+1 < pstr.attempts {
			time.Sleep(time.Duration(fib()) * time.Second)
		}
	}
	if err != nil {
		utils.Logger.Warning(fmt.Sprintf("<SQSPoster> posting new message, err: %s", err.Error()))
	}
	return
}

func (pstr *SQSPoster) newPosterSession() (s *sqs.SQS, err error) {
	pstr.Lock()
	defer pstr.Unlock()
	if pstr.session == nil {
		var ses *session.Session
		cfg := aws.Config{Endpoint: aws.String(pstr.dialURL)}
		if len(pstr.awsRegion) != 0 {
			cfg.Region = aws.String(pstr.awsRegion)
		}
		if len(pstr.awsID) != 0 &&
			len(pstr.awsKey) != 0 {
			cfg.Credentials = credentials.NewStaticCredentials(pstr.awsID, pstr.awsKey, pstr.awsToken)
		}
		ses, err = session.NewSessionWithOptions(
			session.Options{
				Config: cfg,
			},
		)
		if err != nil {
			return nil, err
		}
		pstr.session = ses
	}
	return sqs.New(pstr.session), nil
}
