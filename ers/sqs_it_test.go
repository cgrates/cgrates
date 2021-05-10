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
	"flag"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	itTestSQS = flag.Bool("sqs", false, "Run the test for SQSReader")
)

func TestSQSER(t *testing.T) {
	if !*itTestSQS {
		t.SkipNow()
	}
	cfg, err := config.NewCGRConfigFromJSONStringWithDefaults(`{
"ers": {									// EventReaderService
	"enabled": true,						// starts the EventReader service: <true|false>
	"readers": [
		{
			"id": "sqs",										// identifier of the EventReader profile
			"type": "*sqs_json_map",							// reader type <*file_csv>
			"run_delay":  "-1",									// sleep interval in seconds between consecutive runs, -1 to use automation via inotify or 0 to disable running all together
			"concurrent_requests": 1024,						// maximum simultaneous requests/files to process, 0 for unlimited
			"source_path": "sqs.us-east-2.amazonaws.com",		// read data from this path
			// "processed_path": "/var/spool/cgrates/ers/out",	// move processed data here
			"tenant": "cgrates.org",							// tenant used by import
			"filters": [],										// limit parsing based on the filters
			"flags": [],										// flags to influence the event processing
			"opts": {
				"sqsQueueID": "cgrates-cdrs",
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

	if rdr, err = NewSQSER(cfg, 1, rdrEvents,
		rdrErr, new(engine.FilterS), rdrExit); err != nil {
		t.Fatal(err)
	}
	sqsRdr := rdr.(*SQSER)
	var sess *session.Session
	awsCfg := aws.Config{Endpoint: aws.String(rdr.Config().SourcePath)}
	awsCfg.Region = aws.String(sqsRdr.awsRegion)
	awsCfg.Credentials = credentials.NewStaticCredentials(sqsRdr.awsID, sqsRdr.awsKey, sqsRdr.awsToken)

	if sess, err = session.NewSessionWithOptions(session.Options{Config: awsCfg}); err != nil {
		return
	}
	scv := sqs.New(sess)

	randomCGRID := utils.UUIDSha1Prefix()
	scv.SendMessage(&sqs.SendMessageInput{
		MessageBody: aws.String(fmt.Sprintf(`{"CGRID": "%s"}`, randomCGRID)),
		QueueUrl:    sqsRdr.queueURL,
	})

	if err = rdr.Serve(); err != nil {
		t.Fatal(err)
	}

	select {
	case err = <-rdrErr:
		t.Error(err)
	case ev := <-rdrEvents:
		if ev.rdrCfg.ID != "sqs" {
			t.Errorf("Expected 'sqs' received `%s`", ev.rdrCfg.ID)
		}
		expected := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     ev.cgrEvent.ID,
			Time:   ev.cgrEvent.Time,
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
