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
	"bytes"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/cgrates/cgrates/utils"
)

func NewS3Poster(dialURL string, attempts int, fallbackFileDir string) (Poster, error) {
	pstr := &S3Poster{
		attempts:        attempts,
		fallbackFileDir: fallbackFileDir,
	}
	err := pstr.parseURL(dialURL)
	if err != nil {
		return nil, err
	}
	return pstr, nil
}

type S3Poster struct {
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

func (pstr *S3Poster) Close() {}

func (pstr *S3Poster) parseURL(dialURL string) (err error) {
	u, err := url.Parse(dialURL)
	if err != nil {
		return err
	}
	qry := u.Query()

	pstr.dialURL = strings.Split(dialURL, "?")[0]
	pstr.dialURL = strings.TrimSuffix(pstr.dialURL, "/") // used to remove / to point to correct endpoint
	pstr.queueID = defaultQueueID
	if vals, has := qry[queueID]; has && len(vals) != 0 {
		pstr.queueID = vals[0]
	}
	if vals, has := qry[awsRegion]; has && len(vals) != 0 {
		pstr.awsRegion = vals[0]
	} else {
		utils.Logger.Warning("<S3Poster> No region present for AWS.")
	}
	if vals, has := qry[awsID]; has && len(vals) != 0 {
		pstr.awsID = vals[0]
	} else {
		utils.Logger.Warning("<S3Poster> No access key ID present for AWS.")
	}
	if vals, has := qry[awsSecret]; has && len(vals) != 0 {
		pstr.awsKey = vals[0]
	} else {
		utils.Logger.Warning("<S3Poster> No secret access key present for AWS.")
	}
	if vals, has := qry[awsToken]; has && len(vals) != 0 {
		pstr.awsToken = vals[0]
	} else {
		utils.Logger.Warning("<S3Poster> No session token present for AWS.")
	}
	return nil
}

func (pstr *S3Poster) Post(message []byte, fallbackFileName string) (err error) {
	var svc *s3manager.Uploader
	fib := utils.Fib()

	for i := 0; i < pstr.attempts; i++ {
		if svc, err = pstr.newPosterSession(); err == nil {
			break
		}
		time.Sleep(time.Duration(fib()) * time.Second)
	}
	if err != nil {
		if fallbackFileName != utils.META_NONE {
			utils.Logger.Warning(fmt.Sprintf("<S3Poster> creating new session, err: %s", err.Error()))
			err = writeToFile(pstr.fallbackFileDir, fallbackFileName, message)
		}
		return err
	}

	for i := 0; i < pstr.attempts; i++ {
		if _, err = svc.Upload(&s3manager.UploadInput{
			Bucket: aws.String(pstr.queueID),

			// Can also use the `filepath` standard library package to modify the
			// filename as need for an S3 object key. Such as turning absolute path
			// to a relative path.
			Key: aws.String(fallbackFileName),

			// The file to be uploaded. io.ReadSeeker is preferred as the Uploader
			// will be able to optimize memory when uploading large content. io.Reader
			// is supported, but will require buffering of the reader's bytes for
			// each part.
			Body: bytes.NewReader(message),
		}); err == nil {
			break
		}
	}
	if err != nil && fallbackFileName != utils.META_NONE {
		utils.Logger.Warning(fmt.Sprintf("<S3Poster> posting new message, err: %s", err.Error()))
		err = writeToFile(pstr.fallbackFileDir, fallbackFileName, message)
	}
	return err
}

func (pstr *S3Poster) newPosterSession() (s *s3manager.Uploader, err error) {
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
	return s3manager.NewUploader(pstr.session), nil
}
