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

package engine

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/cgrates/cgrates/utils"
)

func NewSQSPoster(dialURL string, attempts int, fallbackFileDir string) (Poster, error) {
	pstr := &SQSPoster{
		attempts:        attempts,
		fallbackFileDir: fallbackFileDir,
	}
	pstr.parseURL(dialURL)
	return pstr, nil
}

type SQSPoster struct {
	sync.Mutex
	dialURL         string
	awsRegion       string
	awsID           string
	awsKey          string
	awsToken        string
	attempts        int
	fallbackFileDir string
	queueURL        *string
	queueID         string
	// getQueueOnce    sync.Once
	session *session.Session
}

func (pstr *SQSPoster) Close() {}

func (pstr *SQSPoster) parseURL(dialURL string) {
	qry := utils.GetUrlRawArguments(dialURL)

	pstr.dialURL = strings.Split(dialURL, "?")[0]
	pstr.dialURL = strings.TrimSuffix(pstr.dialURL, "/") // used to remove / to point to correct endpoint
	pstr.queueID = defaultQueueID
	if val, has := qry[queueID]; has {
		pstr.queueID = val
	}
	if val, has := qry[utils.AWSRegion]; has {
		pstr.awsRegion = val
	}
	if val, has := qry[utils.AWSKey]; has {
		pstr.awsID = val
	}
	if val, has := qry[utils.AWSSecret]; has {
		pstr.awsKey = val
	}
	if val, has := qry[awsToken]; has {
		pstr.awsToken = val
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

func (pstr *SQSPoster) Post(message []byte, fallbackFileName, _ string) (err error) {
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
		if fallbackFileName != utils.META_NONE {
			utils.Logger.Warning(fmt.Sprintf("<SQSPoster> creating new session, err: %s", err.Error()))
			err = writeToFile(pstr.fallbackFileDir, fallbackFileName, message)
		}
		return err
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
	if err != nil && fallbackFileName != utils.META_NONE {
		utils.Logger.Warning(fmt.Sprintf("<SQSPoster> posting new message, err: %s", err.Error()))
		err = writeToFile(pstr.fallbackFileDir, fallbackFileName, message)
	}
	return err
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
