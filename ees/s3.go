/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package ees

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// NewS3EE creates a s3 poster
func NewS3EE(cfg *config.EventExporterCfg, em *utils.ExporterMetrics) *S3EE {
	pstr := &S3EE{
		cfg:  cfg,
		em:   em,
		reqs: newConcReq(cfg.ConcurrentRequests),
	}
	pstr.parseOpts(cfg.Opts)
	return pstr
}

// S3EE is a s3 poster
type S3EE struct {
	awsRegion  string
	awsID      string
	awsKey     string
	awsToken   string
	bucket     string
	folderPath string
	session    *session.Session
	up         *s3manager.Uploader

	cfg          *config.EventExporterCfg
	em           *utils.ExporterMetrics
	reqs         *concReq
	sync.RWMutex // protect connection
	bytePreparing
}

func (pstr *S3EE) parseOpts(opts *config.EventExporterOpts) {
	pstr.bucket = utils.DefaultQueueID
	if opts.S3BucketID != nil {
		pstr.bucket = *opts.S3BucketID
	}
	if opts.S3FolderPath != nil {
		pstr.folderPath = *opts.S3FolderPath
	}
	if opts.AWSRegion != nil {
		pstr.awsRegion = *opts.AWSRegion
	}
	if opts.AWSKey != nil {
		pstr.awsID = *opts.AWSKey
	}
	if opts.AWSSecret != nil {
		pstr.awsKey = *opts.AWSSecret
	}
	if opts.AWSToken != nil {
		pstr.awsToken = *opts.AWSToken
	}
}

func (pstr *S3EE) Cfg() *config.EventExporterCfg { return pstr.cfg }

func (pstr *S3EE) Connect() (err error) {
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
	if pstr.up == nil {
		pstr.up, err = s3manager.NewUploader(pstr.session), nil
	}
	return
}

func (pstr *S3EE) ExportEvent(ctx *context.Context, message, extraData any) (err error) {
	pstr.reqs.get()
	pstr.RLock()
	sKey := extraData.(string)
	_, err = pstr.up.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: aws.String(pstr.bucket),

		// Can also use the `filepath` standard library package to modify the
		// filename as need for an S3 object key. Such as turning absolute path
		// to a relative path.
		Key: aws.String(fmt.Sprintf("%s/%s.json", pstr.folderPath, sKey)),

		// The file to be uploaded. io.ReadSeeker is preferred as the Uploader
		// will be able to optimize memory when uploading large content. io.Reader
		// is supported, but will require buffering of the reader's bytes for
		// each part.
		Body: bytes.NewReader(message.([]byte)),
	})
	pstr.RUnlock()
	pstr.reqs.done()
	return
}

func (pstr *S3EE) Close() (_ error) { return }

func (pstr *S3EE) GetMetrics() *utils.ExporterMetrics { return pstr.em }

func (pstr *S3EE) ExtraData(ev *utils.CGREvent) any {
	return utils.ConcatenatedKey(
		utils.FirstNonEmpty(engine.MapEvent(ev.APIOpts).GetStringIgnoreErrors(utils.MetaOriginID), utils.GenUUID()),
		utils.FirstNonEmpty(engine.MapEvent(ev.APIOpts).GetStringIgnoreErrors(utils.MetaRunID), utils.MetaDefault),
	)
}
