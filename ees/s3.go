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
	"bytes"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// NewS3EE creates a s3 poster
func NewS3EE(cfg *config.EventExporterCfg, dc *utils.SafeMapStorage) *S3EE {
	pstr := &S3EE{
		cfg:  cfg,
		dc:   dc,
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
	dc           *utils.SafeMapStorage
	reqs         *concReq
	sync.RWMutex // protect connection
	bytePreparing
}

func (pstr *S3EE) parseOpts(opts map[string]interface{}) {
	pstr.bucket = utils.DefaultQueueID
	if val, has := opts[utils.S3Bucket]; has {
		pstr.bucket = utils.IfaceAsString(val)
	}
	if val, has := opts[utils.S3FolderPath]; has {
		pstr.folderPath = utils.IfaceAsString(val)
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

func (pstr *S3EE) ExportEvent(ctx *context.Context, message interface{}, key string) (err error) {
	pstr.reqs.get()
	pstr.RLock()
	_, err = pstr.up.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: aws.String(pstr.bucket),

		// Can also use the `filepath` standard library package to modify the
		// filename as need for an S3 object key. Such as turning absolute path
		// to a relative path.
		Key: aws.String(fmt.Sprintf("%s/%s.json", pstr.folderPath, key)),

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

func (pstr *S3EE) GetMetrics() *utils.SafeMapStorage { return pstr.dc }
