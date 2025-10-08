//go:build integration
// +build integration

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

// import (
// 	"flag"
// 	"fmt"
// 	"net/rpc"
// 	"path"
// 	"testing"
// 	"time"

// 	"github.com/aws/aws-sdk-go/aws"
// 	"github.com/aws/aws-sdk-go/aws/credentials"
// 	"github.com/aws/aws-sdk-go/aws/session"
// 	"github.com/aws/aws-sdk-go/service/s3"
// 	"github.com/aws/aws-sdk-go/service/s3/s3manager"
// 	"github.com/cgrates/birpc/context"
// 	"github.com/cgrates/cgrates/config"
// 	"github.com/cgrates/cgrates/engine"
// 	"github.com/cgrates/cgrates/utils"
// )

// var (
// 	runS3Test = flag.Bool("s3_ees", false, "Run the integration test for the S3 exporter")
// 	originID  string // needed to compute the key when verifying export
// 	s3ConfDir string
// 	s3CfgPath string
// 	s3Cfg     *config.CGRConfig
// 	s3RPC     *rpc.Client

// 	sTestsS3 = []func(t *testing.T){
// 		testS3LoadConfig,
// 		testS3ResetDataDB,
// 		testS3StartEngine,
// 		testS3RPCConn,
// 		testS3ExportEvent,
// 		testS3VerifyExport,
// 		testStopCgrEngine,
// 	}
// )

// func TestS3Export(t *testing.T) {
// 	if !*runS3Test {
// 		t.SkipNow()
// 	}
// 	s3ConfDir = "ees_cloud"
// 	for _, stest := range sTestsS3 {
// 		t.Run(s3ConfDir, stest)
// 	}
// }

// func testS3LoadConfig(t *testing.T) {
// 	var err error
// 	s3CfgPath = path.Join(*utils.DataDir, "conf", "samples", s3ConfDir)
// 	if s3Cfg, err = config.NewCGRConfigFromPath(context.Background(), s3CfgPath); err != nil {
// 		t.Error(err)
// 	}
// 	for _, value := range s3Cfg.EEsCfg().Exporters {
// 		if value.ID == "sqs_test_file" {
// 			awsKey = *value.Opts.AWSKey
// 			awsSecret = *value.Opts.AWSSecret
// 		}
// 	}
// }

// func testS3ResetDataDB(t *testing.T) {
// 	if err := engine.InitDataDB(s3Cfg); err != nil {
// 		t.Fatal(err)
// 	}
// }

// func testS3StartEngine(t *testing.T) {
// 	if _, err := engine.StopStartEngine(s3CfgPath, *utils.WaitRater); err != nil {
// 		t.Fatal(err)
// 	}
// }

// func testS3RPCConn(t *testing.T) {
// 	var err error
// 	s3RPC, err = engine.NewRPCClient(s3Cfg.ListenCfg(), *utils.Encoding)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// }

// func testS3ExportEvent(t *testing.T) {
// 	originID = utils.Sha1("abcdef", time.Unix(1383813745, 0).UTC().String())
// 	ev := &utils.CGREventWithEeIDs{
// 		EeIDs: []string{"s3_test_file"},
// 		CGREvent: &utils.CGREvent{
// 			Tenant: "cgrates.org",
// 			ID:     "dataEvent",
// 			Event: map[string]any{
// 				utils.ToR:          utils.MetaData,
// 				utils.OriginID:     "abcdef",
// 				utils.OriginHost:   "192.168.1.1",
// 				utils.RequestType:  utils.MetaRated,
// 				utils.Tenant:       "AnotherTenant",
// 				utils.Category:     "call", //for data CDR use different Tenant
// 				utils.AccountField: "1001",
// 				utils.Subject:      "1001",
// 				utils.Destination:  "1002",
// 				utils.SetupTime:    time.Unix(1383813745, 0).UTC(),
// 				utils.AnswerTime:   time.Unix(1383813746, 0).UTC(),
// 				utils.Usage:        10 * time.Nanosecond,
// 				utils.RunID:        utils.MetaDefault,
// 				utils.Cost:         0.012,
// 			},
// 			APIOpts: map[string]any{
// 				utils.MetaOriginID: originID,
// 			},
// 		},
// 	}

// 	var reply map[string]utils.MapStorage
// 	if err := s3RPC.Call(utils.EeSv1ProcessEvent, ev, &reply); err != nil {
// 		t.Error(err)
// 	}
// 	time.Sleep(2 * time.Second)
// }

// func testS3VerifyExport(t *testing.T) {
// 	endpoint := "s3.eu-central-1.amazonaws.com"
// 	region := "eu-central-1"
// 	qname := "cgrates-cdrs"

// 	key := fmt.Sprintf("%s/%s:%s.json", "", originID, utils.MetaDefault)

// 	var sess *session.Session
// 	cfg := aws.Config{Endpoint: aws.String(endpoint)}
// 	cfg.Region = aws.String(region)

// 	cfg.Credentials = credentials.NewStaticCredentials(awsKey, awsSecret, "")
// 	var err error
// 	sess, err = session.NewSessionWithOptions(
// 		session.Options{
// 			Config: cfg,
// 		},
// 	)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	s3Clnt := s3.New(sess)
// 	s3Clnt.DeleteObject(&s3.DeleteObjectInput{})
// 	file := aws.NewWriteAtBuffer([]byte{})
// 	svc := s3manager.NewDownloader(sess)

// 	if _, err = svc.Download(file,
// 		&s3.GetObjectInput{
// 			Bucket: aws.String(qname),
// 			Key:    aws.String(key),
// 		}); err != nil {
// 		t.Fatalf("Unable to download item %v", err)
// 	}

// 	expected := `{"Account":"1001","Category":"call","Destination":"1002","OriginID":"abcdef","RequestType":"*rated","RunID":"*default","Subject":"1001","Tenant":"AnotherTenant","ToR":"*data"}`
// 	if rply := string(file.Bytes()); rply != expected {
// 		t.Errorf("Expected: %q, received: %q", expected, rply)
// 	}
// }
