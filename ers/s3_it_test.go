// +build integration

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
	"bytes"
	"flag"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	itTestS3 = flag.Bool("s3", false, "Run the test for S3Reader")
)

func TestS3ER(t *testing.T) {
	if !*itTestS3 {
		t.SkipNow()
	}
	cfg, err := config.NewCGRConfigFromJSONStringWithDefaults(`{
"ers": {									// EventReaderService
	"enabled": true,						// starts the EventReader service: <true|false>
	"readers": [
		{
			"id": "s3",										// identifier of the EventReader profile
			"type": "*s3_json_map",							// reader type <*file_csv>
			"run_delay":  "-1",									// sleep interval in seconds between consecutive runs, -1 to use automation via inotify or 0 to disable running all together
			"concurrent_requests": 1024,						// maximum simultaneous requests/files to process, 0 for unlimited
			"source_path": "s3.us-east-2.amazonaws.com",		// read data from this path
			// "processed_path": "/var/spool/cgrates/ers/out",	// move processed data here
			"tenant": "cgrates.org",							// tenant used by import
			"filters": [],										// limit parsing based on the filters
			"flags": [],										// flags to influence the event processing
			"opts": {
				"s3BucketID": "cgrates-cdrs",
				"awsRegion": "us-east-2",
				"awsKey": "AWSAccessKeyId",
				"awsSecret": "AWSSecretKey",
				// "awsToken": "".
			},
			"fields":[									// import fields template, tag will match internally CDR field, in case of .csv value will be represented by index of the field value
				{"tag": "CGRID", "type": "*composed", "value": "~*req.CGRID", "path": "*cgreq.CGRID"},
			],
		},
	],
},
}`)
	if err != nil {
		t.Fatal(err)
	}

	rdrEvents = make(chan *erEvent, 1)
	rdrErr = make(chan error, 1)
	rdrExit = make(chan struct{}, 1)

	if rdr, err = NewS3ER(cfg, 1, rdrEvents, make(chan *erEvent, 1),
		rdrErr, new(engine.FilterS), rdrExit); err != nil {
		t.Fatal(err)
	}
	s3Rdr := rdr.(*S3ER)
	var sess *session.Session
	awsCfg := aws.Config{Endpoint: aws.String(rdr.Config().SourcePath)}
	awsCfg.Region = aws.String(s3Rdr.awsRegion)
	awsCfg.Credentials = credentials.NewStaticCredentials(s3Rdr.awsID, s3Rdr.awsKey, s3Rdr.awsToken)

	if sess, err = session.NewSessionWithOptions(session.Options{Config: awsCfg}); err != nil {
		return
	}
	scv := s3manager.NewUploader(sess)

	randomCGRID := utils.UUIDSha1Prefix()
	scv.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s3Rdr.bucket),
		Key:    aws.String("home/test.json"),
		Body:   bytes.NewReader([]byte(fmt.Sprintf(`{"CGRID": "%s"}`, randomCGRID))),
	})

	if err = rdr.Serve(); err != nil {
		t.Fatal(err)
	}

	select {
	case err = <-rdrErr:
		t.Error(err)
	case ev := <-rdrEvents:
		if ev.rdrCfg.ID != "s3" {
			t.Errorf("Expected 's3' received `%s`", ev.rdrCfg.ID)
		}
		expected := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     ev.cgrEvent.ID,
			Event: map[string]interface{}{
				"CGRID": randomCGRID,
			},
		}
		if !reflect.DeepEqual(ev.cgrEvent, expected) {
			t.Errorf("Expected %s ,received %s", utils.ToJSON(expected), utils.ToJSON(ev.cgrEvent))
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Timeout")
	}
	close(rdrExit)
}

func TestNewS3ER(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	expected := &S3ER{
		cgrCfg:    cfg,
		cfgIdx:    1,
		fltrS:     nil,
		rdrEvents: nil,
		rdrExit:   nil,
		rdrErr:    nil,
		cap:       nil,
		awsRegion: "",
		awsID:     "",
		awsKey:    "",
		awsToken:  "",
		bucket:    "cgrates_cdrs",
		session:   nil,
		poster:    nil,
	}
	cfg.ERsCfg().Readers = []*config.EventReaderCfg{
		{
			ID:             utils.MetaDefault,
			Type:           utils.MetaNone,
			RunDelay:       0,
			ConcurrentReqs: -1,
			SourcePath:     "/var/spool/cgrates/ers/in",
			ProcessedPath:  "/var/spool/cgrates/ers/out",
			Filters:        []string{},
			Opts:           make(map[string]interface{}),
		},
		{
			ID:             utils.MetaDefault,
			Type:           utils.MetaNone,
			RunDelay:       0,
			ConcurrentReqs: -1,
			SourcePath:     "/var/spool/cgrates/ers/in",
			ProcessedPath:  "/var/spool/cgrates/ers/out",
			Filters:        []string{},
			Opts:           make(map[string]interface{}),
		},
	}

	rdr, err := NewS3ER(cfg, 1, nil, nil,
		nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rdr, expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, rdr)
	}
}

func TestNewS3ERCase2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	expected := &S3ER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     nil,
		rdrEvents: nil,
		rdrExit:   nil,
		rdrErr:    nil,
		cap:       nil,
		awsRegion: "",
		awsID:     "",
		awsKey:    "",
		awsToken:  "",
		bucket:    "cgrates_cdrs",
		session:   nil,
		poster:    nil,
	}
	cfg.ERsCfg().Readers = []*config.EventReaderCfg{
		{
			ID:             utils.MetaDefault,
			Type:           utils.MetaNone,
			RunDelay:       0,
			ConcurrentReqs: 1,
			SourcePath:     "/var/spool/cgrates/ers/in",
			ProcessedPath:  "/var/spool/cgrates/ers/out",
			Filters:        []string{},
			Opts:           make(map[string]interface{}),
		},
	}

	rdr, err := NewS3ER(cfg, 0, nil, nil,
		nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	expected.cap = rdr.(*S3ER).cap
	if !reflect.DeepEqual(rdr, expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, rdr)
	}
}
