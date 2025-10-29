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
	"net/rpc"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	s3CfgPath   string
	s3Cfg       *config.CGRConfig
	s3RPC       *rpc.Client
	s3ConfigDIR string

	sTestsCDReS3 = []func(t *testing.T){
		testS3InitCfg,
		testS3InitDataDb,
		testS3ResetStorDb,
		testS3StartEngine,
		testS3RPCConn,
		testS3AddCDRs,
		testS3ExportCDRs,
		testS3VerifyExport,
		testS3KillEngine,
	}
)

func TestS3Export(t *testing.T) {
	s3ConfigDIR = "cdre"
	for _, stest := range sTestsCDReS3 {
		t.Run(s3ConfigDIR, stest)
	}
}

func testS3InitCfg(t *testing.T) {
	var err error
	s3CfgPath = path.Join("/usr/share/cgrates", "conf", "samples", s3ConfigDIR)
	s3Cfg, err = config.NewCGRConfigFromPath(s3CfgPath)
	if err != nil {
		t.Fatal(err)
	}
	s3Cfg.DataFolderPath = "/usr/share/cgrates" // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(s3Cfg)
}

func testS3InitDataDb(t *testing.T) {
	if err := engine.InitDataDb(s3Cfg); err != nil {
		t.Fatal(err)
	}
}

func testS3ResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(s3Cfg); err != nil {
		t.Fatal(err)
	}
}

func testS3StartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(s3CfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testS3RPCConn(t *testing.T) {
	var err error
	s3RPC, err = newRPCClient(s3Cfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testS3AddCDRs(t *testing.T) {
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
		if err := s3RPC.Call(utils.CDRsV1ProcessCDR, &engine.CDRWithArgDispatcher{CDR: cdr}, &reply); err != nil {
			t.Error("Unexpected error: ", err.Error())
		} else if reply != utils.OK {
			t.Error("Unexpected reply received: ", reply)
		}
	}
	time.Sleep(100 * time.Millisecond)
}

func testS3ExportCDRs(t *testing.T) {
	attr := ArgExportCDRs{
		ExportArgs: map[string]any{
			utils.ExportTemplate: "s3_exporter",
		},
		Verbose: true,
	}
	var rply RplExportedCDRs
	if err := s3RPC.Call(utils.APIerSv1ExportCDRs, attr, &rply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(rply.ExportedCGRIDs) != 2 {
		t.Errorf("Unexpected number of CDR exported: %s ", utils.ToJSON(rply))
	}
}

func testS3VerifyExport(t *testing.T) {
	endpoint := "s3.eu-central-1.amazonaws.com"
	region := "eu-central-1"
	qname := "cgrates-cdrs"

	keys := []string{"Cdr2:*default.json", "Cdr3:*default.json"}

	var sess *session.Session
	cfg := aws.Config{Endpoint: aws.String(endpoint)}
	cfg.Region = aws.String(region)

	cfg.Credentials = credentials.NewStaticCredentials("testKey", "testSecret", "")
	var err error
	sess, err = session.NewSessionWithOptions(
		session.Options{
			Config: cfg,
		},
	)
	if err != nil {
		t.Error(err)
	}
	s3Clnt := s3.New(sess)
	s3Clnt.DeleteObject(&s3.DeleteObjectInput{})
	file := aws.NewWriteAtBuffer([]byte{})
	svc := s3manager.NewDownloader(sess)

	expCDRs := []string{
		`{"Account":"1001","CGRID":"Cdr2","Category":"call","Cost":"-1.0000","Destination":"+4986517174963","OriginID":"OriginCDR2","RunID":"*default","Source":"test2","Tenant":"cgrates.org","Usage":"5s"}`,
		`{"Account":"1001","CGRID":"Cdr3","Category":"call","Cost":"-1.0000","Destination":"+4986517174963","OriginID":"OriginCDR3","RunID":"*default","Source":"test2","Tenant":"cgrates.org","Usage":"30s"}`,
	}
	rplyCDRs := make([]string, 0)
	for _, key := range keys {
		if _, err = svc.Download(file,
			&s3.GetObjectInput{
				Bucket: aws.String(qname),
				Key:    aws.String(key),
			}); err != nil {
			t.Fatalf("Unable to download item %v", err)
		}
		rplyCDRs = append(rplyCDRs, string(file.Bytes()))
	}

	if !reflect.DeepEqual(rplyCDRs, expCDRs) {
		t.Errorf("expected: %s,\nreceived: %s", expCDRs, rplyCDRs)
	}
}

func testS3KillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
