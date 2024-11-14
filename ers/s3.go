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
	"io"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// NewS3ER return a new s3 event reader
func NewS3ER(cfg *config.CGRConfig, cfgIdx int,
	rdrEvents, partialEvents chan *erEvent, rdrErr chan error,
	fltrS *engine.FilterS, rdrExit chan struct{}) (er EventReader, err error) {

	rdr := &S3ER{
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
	}
	rdr.parseOpts(rdr.Config().Opts)
	return rdr, nil
}

// S3ER implements EventReader interface for s3 message
type S3ER struct {
	// sync.RWMutex
	cgrCfg *config.CGRConfig
	cfgIdx int // index of config instance within ERsCfg.Readers
	fltrS  *engine.FilterS

	rdrEvents     chan *erEvent // channel to dispatch the events created to
	partialEvents chan *erEvent // channel to dispatch the partial events created to
	rdrExit       chan struct{}
	rdrErr        chan error
	cap           chan struct{}

	awsRegion string
	awsID     string
	awsKey    string
	awsToken  string
	bucket    string
	session   *session.Session
}

type s3Client interface {
	ListObjectsV2Pages(input *s3.ListObjectsV2Input, fn func(*s3.ListObjectsV2Output, bool) bool) error
	GetObject(input *s3.GetObjectInput) (*s3.GetObjectOutput, error)
	DeleteObject(input *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error)
}

// Config returns the curent configuration
func (rdr *S3ER) Config() *config.EventReaderCfg {
	return rdr.cgrCfg.ERsCfg().Readers[rdr.cfgIdx]
}

// Serve will start the gorutines needed to watch the s3 topic
func (rdr *S3ER) Serve() (err error) {
	var sess *session.Session
	cfg := aws.Config{Endpoint: aws.String(rdr.Config().SourcePath)}
	if len(rdr.awsRegion) != 0 {
		cfg.Region = aws.String(rdr.awsRegion)
	}
	if len(rdr.awsID) != 0 &&
		len(rdr.awsKey) != 0 {
		cfg.Credentials = credentials.NewStaticCredentials(rdr.awsID, rdr.awsKey, rdr.awsToken)
	}
	if sess, err = session.NewSessionWithOptions(session.Options{Config: cfg}); err != nil {
		return
	}
	rdr.session = sess

	if rdr.Config().RunDelay == time.Duration(0) { // 0 disables the automatic read, maybe done per API
		return
	}
	go rdr.readLoop(s3.New(rdr.session)) // read until the connection is closed
	return
}

func (rdr *S3ER) processMessage(body []byte) (err error) {
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
	if pass, err = rdr.fltrS.Pass(agReq.Tenant, rdr.Config().Filters,
		agReq); err != nil || !pass {
		return
	}
	if err = agReq.SetFields(rdr.Config().Fields); err != nil {
		return
	}
	cgrEv := utils.NMAsCGREvent(agReq.CGRRequest, agReq.Tenant, utils.NestingSep, agReq.Opts)
	rdrEv := rdr.rdrEvents
	if _, isPartial := cgrEv.APIOpts[utils.PartialOpt]; isPartial {
		rdrEv = rdr.partialEvents
	}
	rdrEv <- &erEvent{
		cgrEvent: cgrEv,
		rdrCfg:   rdr.Config(),
	}
	return
}

func (rdr *S3ER) parseOpts(opts *config.EventReaderOpts) {
	rdr.bucket = utils.DefaultQueueID
	if awsOpts := opts.AWS; awsOpts != nil {
		if awsOpts.S3BucketID != nil {
			rdr.bucket = *awsOpts.S3BucketID
		}
		if awsOpts.Region != nil {
			rdr.awsRegion = *awsOpts.Region
		}
		if awsOpts.Key != nil {
			rdr.awsID = *awsOpts.Key
		}
		if awsOpts.Secret != nil {
			rdr.awsKey = *awsOpts.Secret
		}
		if awsOpts.Token != nil {
			rdr.awsToken = *awsOpts.Token
		}
	}
}

func (rdr *S3ER) readLoop(scv s3Client) (err error) {
	if rdr.Config().StartDelay > 0 {
		select {
		case <-time.After(rdr.Config().StartDelay):
		case <-rdr.rdrExit:
			return
		}
	}
	var keys []string
	if err = scv.ListObjectsV2Pages(&s3.ListObjectsV2Input{Bucket: aws.String(rdr.bucket)},
		func(lovo *s3.ListObjectsV2Output, _ bool) bool {
			for _, objMeta := range lovo.Contents {
				if objMeta.Key != nil {
					keys = append(keys, *objMeta.Key)
				}
			}
			return !rdr.isClosed()
		}); err != nil {
		rdr.rdrErr <- err
		return
	}
	if rdr.isClosed() {
		return
	}
	for _, key := range keys {
		go rdr.readMsg(scv, key)
	}
	return
}

func (rdr *S3ER) isClosed() bool {
	select {
	case <-rdr.rdrExit:
		return true
	default:
		return false
	}
}

func (rdr *S3ER) readMsg(scv s3Client, key string) (err error) {
	if rdr.Config().ConcurrentReqs != -1 {
		rdr.cap <- struct{}{} // do not try to read if the limit is reached
		defer func() { <-rdr.cap }()
	}
	if rdr.isClosed() {
		return
	}

	obj, err := scv.GetObject(&s3.GetObjectInput{Bucket: &rdr.bucket, Key: &key})
	if err != nil {
		rdr.rdrErr <- err
		return
	}
	var msg []byte
	if msg, err = io.ReadAll(obj.Body); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> decoding message %s error: %s",
				utils.ERs, key, err.Error()))
		return
	}
	obj.Body.Close()
	if err = rdr.processMessage(msg); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> processing message %s error: %s",
				utils.ERs, key, err.Error()))
		return
	}
	if _, err = scv.DeleteObject(&s3.DeleteObjectInput{Bucket: &rdr.bucket, Key: &key}); err != nil {
		rdr.rdrErr <- err
		return
	}

	return
}
