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
	"encoding/json"
	"flag"
	"net/http"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	amqpv1 "github.com/Azure/go-amqp"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	/*
		README
		run test for poster with following commands:
		- sqs
			go test -tags=integration -run=TestSQSPoster -sqs
		- s3
			go test -tags=integration -run=TestS3Poster -s3
		- amqpv1
			go test -tags=integration -run=TestAMQPv1Poster -amqpv1
		also configure the credentials from test function
	*/
	itTestSQS    = flag.Bool("sqs", false, "Run the test for SQSPoster")
	itTestS3     = flag.Bool("s3", false, "Run the test for S3Poster")
	itTestAMQPv1 = flag.Bool("amqpv1", false, "Run the test for AMQPv1Poster")
)

type TestContent struct {
	Var1 string
	Var2 string
}

func TestHttpJsonPoster(t *testing.T) {
	SetFailedPostCacheTTL(time.Millisecond)
	content := &TestContent{Var1: "Val1", Var2: "Val2"}
	jsn, _ := json.Marshal(content)
	pstr, err := NewHTTPjsonMapEE(&config.EventExporterCfg{
		ExportPath:     "http://localhost:8080/invalid",
		Attempts:       3,
		FailedPostsDir: "/tmp",
	}, config.CgrConfig(), nil, nil)
	if err != nil {
		t.Error(err)
	}
	if err = ExportWithAttempts(context.Background(), pstr, &HTTPPosterRequest{Body: jsn, Header: make(http.Header)}, ""); err == nil {
		t.Error("Expected error")
	}
	AddFailedPost("/tmp", "http://localhost:8080/invalid", utils.MetaHTTPjsonMap, "test1", jsn, make(map[string]interface{}))
	time.Sleep(5 * time.Millisecond)
	fs, err := filepath.Glob("/tmp/test1*")
	if err != nil {
		t.Fatal(err)
	} else if len(fs) == 0 {
		t.Fatal("Expected at least one file")
	}

	ev, err := NewExportEventsFromFile(fs[0])
	if err != nil {
		t.Fatal(err)
	} else if len(ev.Events) == 0 {
		t.Fatal("Expected at least one event")
	}
	if !reflect.DeepEqual(jsn, ev.Events[0]) {
		t.Errorf("Expecting: %q, received: %q", string(jsn), ev.Events[0])
	}
}

func TestHttpBytesPoster(t *testing.T) {
	SetFailedPostCacheTTL(time.Millisecond)
	content := []byte(`Test
		Test2
		`)
	pstr, err := NewHTTPjsonMapEE(&config.EventExporterCfg{
		ExportPath:     "http://localhost:8080/invalid",
		Attempts:       3,
		FailedPostsDir: "/tmp",
	}, config.CgrConfig(), nil, nil)
	if err != nil {
		t.Error(err)
	}
	if err = ExportWithAttempts(context.Background(), pstr, &HTTPPosterRequest{Body: content, Header: make(http.Header)}, ""); err == nil {
		t.Error("Expected error")
	}
	AddFailedPost("/tmp", "http://localhost:8080/invalid", utils.ContentJSON, "test2", content, make(map[string]interface{}))
	time.Sleep(5 * time.Millisecond)
	fs, err := filepath.Glob("/tmp/test2*")
	if err != nil {
		t.Fatal(err)
	} else if len(fs) == 0 {
		t.Fatal("Expected at least one file")
	}
	ev, err := NewExportEventsFromFile(fs[0])
	if err != nil {
		t.Fatal(err)
	} else if len(ev.Events) == 0 {
		t.Fatal("Expected at least one event")
	}
	if !reflect.DeepEqual(content, ev.Events[0]) {
		t.Errorf("Expecting: %q, received: %q", string(content), ev.Events[0])
	}
}

func TestSQSPoster(t *testing.T) {
	if !*itTestSQS {
		return
	}
	cfg1 := config.NewDefaultCGRConfig()

	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg1.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)

	//#####################################
	// update this variables
	endpoint := "https://sqs.us-east-2.amazonaws.com"
	region := "us-east-2"
	awsKey := "replace-this-with-your-secret-key"
	awsSecret := "replace-this-with-your-secret"
	qname := "cgrates-cdrs"

	opts := map[string]interface{}{
		utils.AWSRegion:  region,
		utils.AWSKey:     awsKey,
		utils.AWSSecret:  awsSecret,
		utils.SQSQueueID: qname,
	}
	//#####################################

	body := "testString"

	pstr := NewSQSee(&config.EventExporterCfg{
		ExportPath: endpoint,
		Attempts:   5,
		Opts:       opts,
	}, nil)
	if err := ExportWithAttempts(context.Background(), pstr, []byte(body), ""); err != nil {
		t.Fatal(err)
	}

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

	// Create a SQS service client.
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
		VisibilityTimeout:   aws.Int64(30), // 20 seconds
		WaitTimeSeconds:     aws.Int64(0),
	})

	if err != nil {
		t.Error(err)
		return
	}

	if len(result.Messages) != 1 {
		t.Fatalf("Expected 1 message received: %d", len(result.Messages))
	}
	if result.Messages[0].Body == nil {
		t.Fatal("No Msg Body")
	}
	if *result.Messages[0].Body != body {
		t.Errorf("Expected: %q, received: %q", body, *result.Messages[0].Body)
	}
}

func TestS3Poster(t *testing.T) {
	if !*itTestS3 {
		return
	}
	cfg1 := config.NewDefaultCGRConfig()

	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg1.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)

	//#####################################
	// update this variables
	endpoint := "http://s3.us-east-2.amazonaws.com"
	region := "us-east-2"
	awsKey := "replace-this-with-your-secret-key"
	awsSecret := "replace-this-with-your-secret"
	qname := "cgrates-cdrs"

	opts := map[string]interface{}{
		utils.AWSRegion: region,
		utils.AWSKey:    awsKey,
		utils.AWSSecret: awsSecret,
		utils.S3Bucket:  qname,
	}
	//#####################################

	body := "testString"
	key := "key1234"
	pstr := NewS3EE(&config.EventExporterCfg{
		ExportPath: endpoint,
		Attempts:   5,
		Opts:       opts,
	}, nil)
	if err := ExportWithAttempts(context.Background(), pstr, []byte(body), key); err != nil {
		t.Fatal(err)
	}
	key += ".json"
	var sess *session.Session
	cfg := aws.Config{Endpoint: aws.String(endpoint)}
	cfg.Region = aws.String(region)

	cfg.Credentials = credentials.NewStaticCredentials(awsKey, awsSecret, "")
	var err error
	sess, err = session.NewSessionWithOptions(
		session.Options{
			Config: cfg,
		},
	)
	s31 := s3.New(sess)
	s31.DeleteObject(&s3.DeleteObjectInput{})
	file := aws.NewWriteAtBuffer([]byte{})
	// Create a SQS service client.
	svc := s3manager.NewDownloader(sess)

	if _, err = svc.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(qname),
			Key:    aws.String(key),
		}); err != nil {
		t.Fatalf("Unable to download item %v", err)
	}

	if rply := string(file.Bytes()); rply != body {
		t.Errorf("Expected: %q, received: %q", body, rply)
	}
}

func TestAMQPv1Poster(t *testing.T) {
	if !*itTestAMQPv1 {
		return
	}
	cfg1 := config.NewDefaultCGRConfig()

	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg1.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)

	//#####################################
	// update this variables
	endpoint := "amqps://RootManageSharedAccessKey:UlfIJ%2But11L0ZzA%2Fgpje8biFJeQihpWibJsUhaOi1DU%3D@cdrscgrates.servicebus.windows.net"
	qname := "cgrates-cdrs"
	opts := map[string]interface{}{
		utils.AMQPQueueID: qname,
	}
	//#####################################

	body := "testString"

	pstr := NewAMQPv1EE(&config.EventExporterCfg{
		ExportPath: endpoint,
		Attempts:   5,
		Opts:       opts,
	}, nil)
	if err := ExportWithAttempts(context.Background(), pstr, []byte(body), ""); err != nil {
		t.Fatal(err)
	}
	// Create client
	client, err := amqpv1.Dial(endpoint)
	if err != nil {
		t.Fatal("Dialing AMQP server:", err)
	}
	defer client.Close()

	// Open a session
	session, err := client.NewSession()
	if err != nil {
		t.Fatal("Creating AMQP session:", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	// Create a receiver
	receiver, err := session.NewReceiver(
		amqpv1.LinkSourceAddress("/" + qname),
	)
	if err != nil {
		t.Fatal("Creating receiver link:", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		receiver.Close(ctx)
		cancel()
	}()

	// Receive next message
	msg, err := receiver.Receive(ctx)
	cancel()
	if err != nil {
		t.Fatal("Reading message from AMQP:", err)
	}

	// Accept message
	msg.Accept(ctx)
	if rply := string(msg.GetData()); rply != body {
		t.Errorf("Expected: %q, received: %q", body, rply)
	}
}
