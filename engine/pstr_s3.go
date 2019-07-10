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
	pstr.parseURL(dialURL)
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
	queueID         string
	folderPath      string
	session         *session.Session
}

func (pstr *S3Poster) Close() {}

func (pstr *S3Poster) parseURL(dialURL string) {
	qry := utils.GetUrlRawArguments(dialURL)

	pstr.dialURL = strings.Split(dialURL, "?")[0]
	pstr.dialURL = strings.TrimSuffix(pstr.dialURL, "/") // used to remove / to point to correct endpoint
	pstr.queueID = defaultQueueID
	if val, has := qry[queueID]; has {
		pstr.queueID = val
	}
	if val, has := qry[folderPath]; has {
		pstr.folderPath = val
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
}

func (pstr *S3Poster) Post(message []byte, fallbackFileName, key string) (err error) {
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
			Key: aws.String(fmt.Sprintf("%s/%s.json", pstr.folderPath, key)),

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
