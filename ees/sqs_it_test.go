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
	"path"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	runSQSTest = flag.Bool("sqs_ees", false, "Run the integration test for the SQS exporter")
	awsKey     string
	awsSecret  string
	sqsConfDir string
	sqsCfgPath string
	sqsCfg     *config.CGRConfig
	sqsRPC     *birpc.Client

	sTestsSQS = []func(t *testing.T){
		testSQSLoadConfig,
		testSQSResetDBs,
		testSQSStartEngine,
		testSQSRPCConn,
		testSQSExportEvent,
		testSQSVerifyExport,
		testStopCgrEngine,
	}
)

func TestSQSExport(t *testing.T) {
	if !*runSQSTest {
		t.SkipNow()
	}
	sqsConfDir = "ees_cloud"
	for _, stest := range sTestsSQS {
		t.Run(sqsConfDir, stest)
	}
}

func testSQSLoadConfig(t *testing.T) {
	var err error
	sqsCfgPath = path.Join(*dataDir, "conf", "samples", sqsConfDir)
	if sqsCfg, err = config.NewCGRConfigFromPath(context.Background(), sqsCfgPath); err != nil {
		t.Error(err)
	}
	for _, value := range sqsCfg.EEsCfg().Exporters {
		if value.ID == "sqs_test_file" {
			awsKey = *value.Opts.AWSKey
			awsSecret = *value.Opts.AWSSecret
		}
	}
}

func testSQSResetDBs(t *testing.T) {
	if err := engine.InitDataDB(sqsCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDB(sqsCfg); err != nil {
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
	sqsRPC, err = engine.NewRPCClient(sqsCfg.ListenCfg(), *encoding)
	if err != nil {
		t.Fatal(err)
	}
}

func testSQSExportEvent(t *testing.T) {
	ev := &utils.CGREventWithEeIDs{
		EeIDs: []string{"sqs_test_file"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "voiceEvent",
			Event: map[string]any{
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
			APIOpts: map[string]any{
				utils.MetaOriginID: utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()),
			},
		},
	}

	var reply map[string]utils.MapStorage
	if err := sqsRPC.Call(context.Background(), utils.EeSv1ProcessEvent, ev, &reply); err != nil {
		t.Error(err)
	}

	time.Sleep(2 * time.Second)
}

func testSQSVerifyExport(t *testing.T) {
	endpoint := "sqs.eu-central-1.amazonaws.com"
	region := "eu-central-1"
	qname := "testQueue"

	var sess *session.Session
	cfg := aws.Config{Endpoint: aws.String(endpoint)}
	cfg.Region = aws.String(region)
	var err error
	cfg.Credentials = credentials.NewStaticCredentials(awsKey, awsSecret, "")
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

	expected := `{"Account":"1001","Category":"call","Destination":"1002","OriginID":"dsafdsaf","RequestType":"*rated","RunID":"*default","Subject":"1001","Tenant":"cgrates.org","ToR":"*voice"}`
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
