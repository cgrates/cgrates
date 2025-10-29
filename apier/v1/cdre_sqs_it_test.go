//go:build aws
// +build aws

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

package v1

import (
	"fmt"
	"net/rpc"
	"path"
	"reflect"
	"sort"
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
	sqsCfgPath   string
	sqsCfg       *config.CGRConfig
	sqsRPC       *rpc.Client
	sqsConfigDIR string

	sTestsCDReSQS = []func(t *testing.T){
		testSQSInitCfg,
		testSQSInitDataDb,
		testSQSResetStorDb,
		testSQSStartEngine,
		testSQSRPCConn,
		testSQSAddCDRs,
		testSQSExportCDRs,
		testSQSVerifyExport,
		testSQSKillEngine,
	}
)

func TestSQSExport(t *testing.T) {
	sqsConfigDIR = "cdre"
	for _, stest := range sTestsCDReSQS {
		t.Run(sqsConfigDIR, stest)
	}
}

func testSQSInitCfg(t *testing.T) {
	var err error
	sqsCfgPath = path.Join("/usr/share/cgrates", "conf", "samples", sqsConfigDIR)
	sqsCfg, err = config.NewCGRConfigFromPath(sqsCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	sqsCfg.DataFolderPath = "/usr/share/cgrates" // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(sqsCfg)
}

func testSQSInitDataDb(t *testing.T) {
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
	if _, err := engine.StopStartEngine(sqsCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testSQSRPCConn(t *testing.T) {
	var err error
	sqsRPC, err = newRPCClient(sqsCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testSQSAddCDRs(t *testing.T) {
	storedCdrs := []*engine.CDR{
		{
			CGRID:       "Cdr1",
			OrderID:     101,
			ToR:         utils.VOICE,
			OriginID:    "OriginCDR1",
			OriginHost:  "192.168.1.1",
			Source:      "test",
			RequestType: utils.META_RATED,
			Tenant:      "cgrates.org",
			Category:    "call",
			Account:     "1001",
			Subject:     "1001",
			Destination: "+4986517174963",
			SetupTime:   time.Now(),
			AnswerTime:  time.Now(),
			RunID:       utils.MetaDefault,
			Usage:       time.Duration(10) * time.Second,
			ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
			Cost:        1.01,
		},
		{
			CGRID:       "Cdr2",
			OrderID:     102,
			ToR:         utils.VOICE,
			OriginID:    "OriginCDR2",
			OriginHost:  "192.168.1.1",
			Source:      "test2",
			RequestType: utils.META_RATED,
			Tenant:      "cgrates.org",
			Category:    "call",
			Account:     "1001",
			Subject:     "1001",
			Destination: "+4986517174963",
			SetupTime:   time.Now(),
			AnswerTime:  time.Now(),
			RunID:       utils.MetaDefault,
			Usage:       time.Duration(5) * time.Second,
			ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
			Cost:        1.01,
		},
		{
			CGRID:       "Cdr3",
			OrderID:     103,
			ToR:         utils.VOICE,
			OriginID:    "OriginCDR3",
			OriginHost:  "192.168.1.1",
			Source:      "test2",
			RequestType: utils.META_RATED,
			Tenant:      "cgrates.org",
			Category:    "call",
			Account:     "1001",
			Subject:     "1001",
			Destination: "+4986517174963",
			SetupTime:   time.Now(),
			AnswerTime:  time.Now(),
			RunID:       utils.MetaDefault,
			Usage:       time.Duration(30) * time.Second,
			ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
			Cost:        1.01,
		},
		{
			CGRID:       "Cdr4",
			OrderID:     104,
			ToR:         utils.VOICE,
			OriginID:    "OriginCDR4",
			OriginHost:  "192.168.1.1",
			Source:      "test3",
			RequestType: utils.META_RATED,
			Tenant:      "cgrates.org",
			Category:    "call",
			Account:     "1001",
			Subject:     "1001",
			Destination: "+4986517174963",
			SetupTime:   time.Now(),
			AnswerTime:  time.Time{},
			RunID:       utils.MetaDefault,
			Usage:       time.Duration(0) * time.Second,
			ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
		},
	}
	for _, cdr := range storedCdrs {
		var reply string
		if err := sqsRPC.Call(utils.CDRsV1ProcessCDR, &engine.CDRWithArgDispatcher{CDR: cdr}, &reply); err != nil {
			t.Error("Unexpected error: ", err.Error())
		} else if reply != utils.OK {
			t.Error("Unexpected reply received: ", reply)
		}
	}
	time.Sleep(100 * time.Millisecond)
}

func testSQSExportCDRs(t *testing.T) {
	attr := ArgExportCDRs{
		ExportArgs: map[string]any{
			utils.ExportTemplate: "sqs_exporter",
		},
		Verbose: true,
	}
	var rply RplExportedCDRs
	if err := sqsRPC.Call(utils.APIerSv1ExportCDRs, attr, &rply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(rply.ExportedCGRIDs) != 2 {
		t.Errorf("Unexpected number of CDR exported: %s ", utils.ToJSON(rply))
	}
}

func testSQSVerifyExport(t *testing.T) {
	endpoint := "sqs.eu-central-1.amazonaws.com"
	region := "eu-central-1"
	qname := "cgrates-cdrs"

	var sess *session.Session
	cfg := aws.Config{Endpoint: aws.String(endpoint)}
	cfg.Region = aws.String(region)
	var err error
	cfg.Credentials = credentials.NewStaticCredentials("testKey", "testSecret", "")
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
		MaxNumberOfMessages: aws.Int64(2),
		VisibilityTimeout:   aws.Int64(30),
		WaitTimeSeconds:     aws.Int64(0),
	})
	time.Sleep(time.Second)
	if err != nil {
		t.Error(err)
		return
	}

	expCDRs := []string{
		`{"Account":"1001","CGRID":"Cdr2","Category":"call","Cost":"-1.0000","Destination":"+4986517174963","OriginID":"OriginCDR2","RunID":"*default","Source":"test2","Tenant":"cgrates.org","Usage":"5s"}`,
		`{"Account":"1001","CGRID":"Cdr3","Category":"call","Cost":"-1.0000","Destination":"+4986517174963","OriginID":"OriginCDR3","RunID":"*default","Source":"test2","Tenant":"cgrates.org","Usage":"30s"}`,
	}
	rplyCDRs := make([]string, 0)
	for _, msg := range result.Messages {
		fmt.Println(111)
		rplyCDRs = append(rplyCDRs, *msg.Body)
	}

	sort.Strings(rplyCDRs)
	if len(rplyCDRs) != 2 {
		t.Fatalf("Expected 2 message received: %d", len(rplyCDRs))
	}
	if !reflect.DeepEqual(rplyCDRs, expCDRs) {
		t.Errorf("expected: %s,\nreceived: %s", expCDRs, rplyCDRs)
	}
}

func testSQSKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
