//go:build integration
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

package ees

import (
	"flag"
	"fmt"
	"net/rpc"
	"path"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	awsKey     = flag.String("awsKey", utils.EmptyString, "Access key ID for IAM user")
	awsSecret  = flag.String("awsSecret", utils.EmptyString, "Secret access key")
	sqsConfDir string
	sqsCfgPath string
	sqsCfg     *config.CGRConfig
	sqsRPC     *rpc.Client

	sTestsSQS = []func(t *testing.T){
		testSQSLoadConfig,
		testSQSResetDataDB,
		testSQSResetStorDb,
		testSQSStartEngine,
		testSQSRPCConn,
		testSQSExportEvent,
		// testSQSVerifyExport,
		testStopCgrEngine,
	}
)

func TestSQSExport(t *testing.T) {
	if awsKey == nil || *awsKey == utils.EmptyString ||
		awsSecret == nil || *awsSecret == utils.EmptyString {
		t.SkipNow()
	}
	sqsConfDir = "ees_s3&sqs"
	for _, stest := range sTestsSQS {
		t.Run(sqsConfDir, stest)
	}
}

func testSQSLoadConfig(t *testing.T) {
	var err error
	sqsCfgPath = path.Join(*dataDir, "conf", "samples", sqsConfDir)
	if sqsCfg, err = config.NewCGRConfigFromPath(sqsCfgPath); err != nil {
		t.Error(err)
	}
	for _, value := range sqsCfg.EEsCfg().Exporters {
		if value.ID == "sqs_test_file" {
			value.ExportPath = fmt.Sprintf("https://sqs.eu-central-1.amazonaws.com/?awsRegion=eu-central-1&awsKey=%s&awsSecret=%s", *awsKey, *awsSecret)
			value.Opts.AWSKey = awsKey
			value.Opts.AWSSecret = awsSecret
		}
	}
}

func testSQSResetDataDB(t *testing.T) {
	if err := engine.InitDataDb(sqsCfg); err != nil {
		t.Fatal(err)
	}
}

func testSQSResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(sqsCfg); err != nil {
		t.Fatal(err)
	}
}

func testSQSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(sqsCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testSQSRPCConn(t *testing.T) {
	var err error
	sqsRPC, err = newRPCClient(sqsCfg.ListenCfg())
	if err != nil {
		t.Fatal(err)
	}
}

func testSQSExportEvent(t *testing.T) {
	ev := &engine.CGREventWithEeIDs{
		EeIDs: []string{"sqs_test_file"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "voiceEvent",
			Time:   utils.TimePointer(time.Now()),
			Event: map[string]interface{}{
				utils.CGRID:        utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()),
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "dsafdsaf",
				utils.OriginHost:   "192.168.1.1",
				utils.RequestType:  utils.MetaRated,
				utils.Tenant:       "cgrates.org",
				utils.Category:     "call",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Unix(1383813745, 0).UTC(),
				utils.AnswerTime:   time.Unix(1383813746, 0).UTC(),
				utils.Usage:        10 * time.Second,
				utils.RunID:        utils.MetaDefault,
				utils.Cost:         1.01,
			},
		},
	}

	var reply map[string]utils.MapStorage
	if err := sqsRPC.Call(utils.EeSv1ProcessEvent, ev, &reply); err != nil {
		t.Error(err)
	}

	time.Sleep(time.Second)
}

func testSQSVerifyExport(t *testing.T) {
	endpoint := fmt.Sprintf("https://sqs.eu-central-1.amazonaws.com/?awsRegion=eu-central-1&awsKey=%s&awsSecret=%s", *awsKey, *awsSecret)
	region := "eu-central-1"
	qname := "testQueue"

	var sess *session.Session
	cfg := aws.Config{Endpoint: aws.String(endpoint)}
	cfg.Region = aws.String(region)
	var err error
	cfg.Credentials = credentials.NewStaticCredentials(*awsKey, *awsSecret, "")
	sess, err = session.NewSessionWithOptions(
		session.Options{
			Config: cfg,
		},
	)
	if err != nil {
		t.Error(err)
	}

	svc := sqs.New(sess)

	resultURL, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(qname),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == sqs.ErrCodeQueueDoesNotExist {
			t.Fatalf("Unable to find queue %q.", qname)
		}
		t.Fatalf("Unable to queue %q, %v.", qname, err)
	}

	result, err := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl:            resultURL.QueueUrl,
		MaxNumberOfMessages: aws.Int64(1),
		VisibilityTimeout:   aws.Int64(30),
		WaitTimeSeconds:     aws.Int64(0),
	})

	if err != nil {
		t.Error(err)
		return
	}

	expected := `{"Account":"1001","CGRID":"dbafe9c8614c785a65aabd116dd3959c3c56f7f6","Category":"call","Destination":"1002","OriginID":"dsafdsaf","RequestType":"*rated","RunID":"*default","Subject":"1001","Tenant":"cgrates.org","ToR":"*voice"}`
	if len(result.Messages) != 1 {
		t.Fatalf("Expected 1 message received: %d", len(result.Messages))
	}
	if result.Messages[0].Body == nil {
		t.Fatal("No Msg Body")
	}
	if *result.Messages[0].Body != expected {
		t.Errorf("Expected: %q, received: %q", expected, *result.Messages[0].Body)
	}
}
