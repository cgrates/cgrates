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
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestNewSQSER(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	expected := &SQSER{
		cgrCfg:  cfg,
		cfgIdx:  0,
		cap:     nil,
		queueID: "cgrates_cdrs",
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
	rdr, err := NewSQSER(cfg, 0, nil,
		nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	expected.cap = rdr.(*SQSER).cap
	expected.session = rdr.(*SQSER).session

	rdr.(*SQSER).poster = nil
	if !reflect.DeepEqual(rdr, expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, rdr)
	}
}

func TestSQSERServeRunDelay0(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
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
	rdr, err := NewSQSER(cfg, 0, nil,
		nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	rdr.Config().RunDelay = time.Duration(0)
	result := rdr.Serve()
	if result != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, result)
	}
}

func TestSQSERServe(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
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
	rdr, err := NewSQSER(cfg, 0, nil,
		nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	rdr.Config().RunDelay = time.Duration(1)
	result := rdr.Serve()
	if result != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, result)
	}
}

func TestSQSERProcessMessage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	rdr := &SQSER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     new(engine.FilterS),
		rdrEvents: make(chan *erEvent, 1),
		rdrExit:   make(chan struct{}),
		rdrErr:    make(chan error, 1),
		cap:       nil,
		awsRegion: "us-east-2",
		awsID:     "AWSId",
		awsKey:    "AWSAccessKeyId",
		awsToken:  "",
		queueID:   "cgrates_cdrs",
		session:   nil,
		poster:    nil,
	}
	expEvent := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			utils.CGRID: "testCgrId",
		},
		APIOpts: map[string]interface{}{},
	}
	body := []byte(`{"CGRID":"testCgrId"}`)
	rdr.Config().Fields = []*config.FCTemplate{
		{
			Tag:   "CGRID",
			Type:  utils.MetaConstant,
			Value: config.NewRSRParsersMustCompile("testCgrId", utils.InfieldSep),
			Path:  "*cgreq.CGRID",
		},
	}
	rdr.Config().Fields[0].ComputePath()
	if err := rdr.processMessage(body); err != nil {
		t.Error(err)
	}
	select {
	case data := <-rdr.rdrEvents:
		expEvent.ID = data.cgrEvent.ID
		expEvent.Time = data.cgrEvent.Time
		if !reflect.DeepEqual(data.cgrEvent, expEvent) {
			t.Errorf("Expected %v but received %v", utils.ToJSON(expEvent), utils.ToJSON(data.cgrEvent))
		}
	case <-time.After(50 * time.Millisecond):
		t.Error("Time limit exceeded")
	}
}

func TestSQSERProcessMessageError1(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	rdr := &SQSER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     new(engine.FilterS),
		rdrEvents: make(chan *erEvent, 1),
		rdrExit:   make(chan struct{}),
		rdrErr:    make(chan error, 1),
		cap:       nil,
		awsRegion: "us-east-2",
		awsID:     "AWSId",
		awsKey:    "AWSAccessKeyId",
		awsToken:  "",
		queueID:   "cgrates_cdrs",
		session:   nil,
		poster:    nil,
	}
	rdr.Config().Fields = []*config.FCTemplate{
		{},
	}
	body := []byte(`{"CGRID":"testCgrId"}`)
	errExpect := "unsupported type: <>"
	if err := rdr.processMessage(body); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
}

func TestSQSERProcessMessageError2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	cfg.ERsCfg().Readers[0].ProcessedPath = ""
	fltrs := engine.NewFilterS(cfg, nil, dm)
	rdr := &SQSER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		rdrEvents: make(chan *erEvent, 1),
		rdrExit:   make(chan struct{}),
		rdrErr:    make(chan error, 1),
		cap:       nil,
		awsRegion: "us-east-2",
		awsID:     "AWSId",
		awsKey:    "AWSAccessKeyId",
		awsToken:  "",
		queueID:   "cgrates_cdrs",
		session:   nil,
		poster:    nil,
	}
	body := []byte(`{"CGRID":"testCgrId"}`)
	rdr.Config().Filters = []string{"Filter1"}
	errExpect := "NOT_FOUND:Filter1"
	if err := rdr.processMessage(body); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}

	//
	rdr.Config().Filters = []string{"*exists:~*req..Account:"}
	if err := rdr.processMessage(body); err != nil {
		t.Error(err)
	}
}

func TestSQSERProcessMessageError3(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	rdr := &SQSER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     new(engine.FilterS),
		rdrEvents: make(chan *erEvent, 1),
		rdrExit:   make(chan struct{}),
		rdrErr:    make(chan error, 1),
		cap:       nil,
		awsRegion: "us-east-2",
		awsID:     "AWSId",
		awsKey:    "AWSAccessKeyId",
		awsToken:  "",
		queueID:   "cgrates_cdrs",
		session:   nil,
		poster:    nil,
	}
	body := []byte("invalid_format")
	errExpect := "invalid character 'i' looking for beginning of value"
	if err := rdr.processMessage(body); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
}

func TestSQSERParseOpts(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	rdr := &SQSER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     new(engine.FilterS),
		rdrEvents: make(chan *erEvent, 1),
		rdrExit:   make(chan struct{}),
		rdrErr:    make(chan error, 1),
		cap:       nil,
		awsRegion: "us-east-2",
		awsID:     "AWSId",
		awsKey:    "AWSAccessKeyId",
		awsToken:  "",
		queueID:   "cgrates_cdrs",
		session:   nil,
		poster:    nil,
	}

	opts := map[string]interface{}{
		utils.SQSQueueID: "QueueID",
		utils.AWSRegion:  "AWSRegion",
		utils.AWSKey:     "AWSKey",
		utils.AWSSecret:  "AWSSecret",
		utils.AWSToken:   "AWSToken",
	}
	rdr.parseOpts(opts)
	if rdr.queueID != opts[utils.SQSQueueID] || rdr.awsRegion != opts[utils.AWSRegion] || rdr.awsID != opts[utils.AWSKey] || rdr.awsKey != opts[utils.AWSSecret] || rdr.awsToken != opts[utils.AWSToken] {
		t.Error("Fields do not corespond")
	}
	rdr.Config().Opts = map[string]interface{}{}
	rdr.Config().ProcessedPath = utils.EmptyString
	rdr.createPoster()
}

func TestSQSERIsClosed(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	rdr := &SQSER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     new(engine.FilterS),
		rdrEvents: make(chan *erEvent, 1),
		rdrExit:   make(chan struct{}, 1),
		rdrErr:    make(chan error, 1),
		cap:       nil,
		awsRegion: "us-east-2",
		awsID:     "AWSId",
		awsKey:    "AWSAccessKeyId",
		awsToken:  "",
		queueID:   "cgrates_cdrs",
		session:   nil,
		poster:    nil,
	}
	if rcv := rdr.isClosed(); rcv != false {
		t.Errorf("Expected %v but received %v", false, true)
	}
	rdr.rdrExit <- struct{}{}
	if rcv := rdr.isClosed(); rcv != true {
		t.Errorf("Expected %v but received %v", true, false)
	}
}

// Mock the SCV
type sqsClientMock struct {
	ReceiveMessageF func(input *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error)
	DeleteMessageF  func(input *sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error)
	GetQueueUrlF    func(input *sqs.GetQueueUrlInput) (*sqs.GetQueueUrlOutput, error)
	CreateQueueF    func(input *sqs.CreateQueueInput) (*sqs.CreateQueueOutput, error)
}

func (s *sqsClientMock) ReceiveMessage(input *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error) {
	if s.ReceiveMessageF != nil {
		return s.ReceiveMessageF(input)
	}
	return nil, utils.ErrNotFound
}

func (s *sqsClientMock) DeleteMessage(input *sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error) {
	if s.DeleteMessageF != nil {
		return s.DeleteMessageF(input)
	}
	return nil, utils.ErrNotImplemented
}

func (s *sqsClientMock) GetQueueUrl(input *sqs.GetQueueUrlInput) (*sqs.GetQueueUrlOutput, error) {
	if s.GetQueueUrlF != nil {
		return s.GetQueueUrlF(input)
	}
	return nil, nil
}

func (s *sqsClientMock) CreateQueue(input *sqs.CreateQueueInput) (*sqs.CreateQueueOutput, error) {
	if s.CreateQueueF != nil {
		return s.CreateQueueF(input)
	}
	return nil, utils.ErrInvalidPath
}

func TestSQSERReadMsg(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	rdr := &SQSER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     new(engine.FilterS),
		rdrEvents: make(chan *erEvent, 1),
		rdrExit:   make(chan struct{}, 1),
		rdrErr:    make(chan error, 1),
		cap:       nil,
		awsRegion: "us-east-2",
		awsID:     "AWSId",
		awsKey:    "AWSAccessKeyId",
		awsToken:  "",
		queueID:   "cgrates_cdrs",
		// queueURL:  utils.StringPointer("url"),
		session: nil,
		poster:  nil,
	}
	awsCfg := aws.Config{Endpoint: aws.String(rdr.Config().SourcePath)}
	rdr.session, _ = session.NewSessionWithOptions(
		session.Options{
			Config: awsCfg,
		},
	)

	rdr.Config().ConcurrentReqs = -1
	rdr.Config().Fields = []*config.FCTemplate{
		{
			Tag:   "Tor",
			Type:  utils.MetaConstant,
			Value: config.NewRSRParsersMustCompile("*voice", utils.InfieldSep),
			Path:  "*cgreq.ToR",
		},
	}
	rdr.Config().Fields[0].ComputePath()
	receiveMessage := func(input *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error) {
		return nil, nil
	}
	deleteMessage := func(input *sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error) {
		return nil, nil
	}
	scv := &sqsClientMock{
		ReceiveMessageF: receiveMessage,
		DeleteMessageF:  deleteMessage,
	}
	msg := &sqs.Message{
		Body:          utils.StringPointer(`{"msgBody":"BODY"}`),
		MessageId:     utils.StringPointer(`{"msgId":"MESSAGE"}`),
		ReceiptHandle: utils.StringPointer(`{"msgReceiptHandle":"RECEIPT_HANDLE"}`),
	}
	if err := rdr.readMsg(scv, msg); err != nil {
		t.Error(err)
	}
}

func TestSQSERReadMsgError1(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	rdr := &SQSER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     new(engine.FilterS),
		rdrEvents: make(chan *erEvent, 1),
		rdrExit:   make(chan struct{}, 1),
		rdrErr:    make(chan error, 1),
		cap:       nil,
		awsRegion: "us-east-2",
		awsID:     "AWSId",
		awsKey:    "AWSAccessKeyId",
		awsToken:  "",
		queueID:   "cgrates_cdrs",
		// queueURL:  utils.StringPointer("url"),
		session: nil,
		poster:  nil,
	}
	awsCfg := aws.Config{Endpoint: aws.String(rdr.Config().SourcePath)}
	rdr.session, _ = session.NewSessionWithOptions(
		session.Options{
			Config: awsCfg,
		},
	)
	rdr.Config().ConcurrentReqs = -1
	rdr.Config().Fields = []*config.FCTemplate{
		{
			Tag:   "Tor",
			Type:  utils.MetaConstant,
			Value: config.NewRSRParsersMustCompile("*voice", utils.InfieldSep),
			Path:  "*cgreq.ToR",
		},
	}
	rdr.Config().Fields[0].ComputePath()
	scv := &sqs.SQS{}
	msg := &sqs.Message{
		Body:          utils.StringPointer(`{"msgBody":"BODY"`),
		MessageId:     utils.StringPointer(`{"msgId":"MESSAGE"}`),
		ReceiptHandle: utils.StringPointer(`{"msgReceiptHandle":"RECEIPT_HANDLE"}`),
	}
	errExp := "unexpected end of JSON input"
	if err := rdr.readMsg(scv, msg); err == nil || err.Error() != errExp {
		t.Errorf("Expected %v but received %v", errExp, err)
	}
}

func TestSQSERReadMsgError2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	rdr := &SQSER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     new(engine.FilterS),
		rdrEvents: make(chan *erEvent, 1),
		rdrExit:   make(chan struct{}, 1),
		rdrErr:    make(chan error, 1),
		cap:       nil,
		awsRegion: "us-east-2",
		awsID:     "AWSId",
		awsKey:    "AWSAccessKeyId",
		awsToken:  "",
		queueID:   "cgrates_cdrs",
		session:   nil,
		poster:    nil,
	}
	awsCfg := aws.Config{Endpoint: aws.String(rdr.Config().SourcePath)}
	rdr.session, _ = session.NewSessionWithOptions(
		session.Options{
			Config: awsCfg,
		},
	)
	rdr.Config().ConcurrentReqs = -1
	rdr.Config().Fields = []*config.FCTemplate{
		{
			Tag:   "Tor",
			Type:  utils.MetaConstant,
			Value: config.NewRSRParsersMustCompile("*voice", utils.InfieldSep),
			Path:  "*cgreq.ToR",
		},
	}
	rdr.Config().Fields[0].ComputePath()
	receiveMessage := func(input *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error) {
		return nil, nil
	}
	scv := &sqsClientMock{
		ReceiveMessageF: receiveMessage,
	}
	msg := &sqs.Message{
		Body:          utils.StringPointer(`{"msgBody":"BODY"}`),
		MessageId:     utils.StringPointer(`{"msgId":"MESSAGE"}`),
		ReceiptHandle: utils.StringPointer(`{"msgReceiptHandle":"RECEIPT_HANDLE"}`),
	}
	errExp := "NOT_IMPLEMENTED"
	if err := rdr.readMsg(scv, msg); err == nil || err.Error() != errExp {
		t.Errorf("Expected %v but received %v", errExp, err)
	}
}

func TestSQSERReadMsgError3(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	rdr := &SQSER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     new(engine.FilterS),
		rdrEvents: make(chan *erEvent, 1),
		rdrExit:   make(chan struct{}, 1),
		rdrErr:    make(chan error, 1),
		cap:       nil,
		awsRegion: "us-east-2",
		awsID:     "AWSId",
		awsKey:    "AWSAccessKeyId",
		awsToken:  "",
		queueID:   "cgrates_cdrs",
		session:   nil,
		poster:    engine.NewSQSPoster("url", 1, make(map[string]interface{})),
	}
	awsCfg := aws.Config{Endpoint: aws.String(rdr.Config().SourcePath)}
	rdr.session, _ = session.NewSessionWithOptions(
		session.Options{
			Config: awsCfg,
		},
	)
	rdr.Config().ConcurrentReqs = -1
	rdr.Config().Fields = []*config.FCTemplate{
		{
			Tag:   "Tor",
			Type:  utils.MetaConstant,
			Value: config.NewRSRParsersMustCompile("*voice", utils.InfieldSep),
			Path:  "*cgreq.ToR",
		},
	}
	rdr.Config().Fields[0].ComputePath()
	receiveMessage := func(input *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error) {
		return nil, nil
	}
	deleteMessage := func(input *sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error) {
		return nil, nil
	}
	scv := &sqsClientMock{
		ReceiveMessageF: receiveMessage,
		DeleteMessageF:  deleteMessage,
	}
	msg := &sqs.Message{
		Body:          utils.StringPointer(`{"msgBody":"BODY"}`),
		MessageId:     utils.StringPointer(`{"msgId":"MESSAGE"}`),
		ReceiptHandle: utils.StringPointer(`{"msgReceiptHandle":"RECEIPT_HANDLE"}`),
	}
	errExp := "MissingRegion: could not find region configuration"
	if err := rdr.readMsg(scv, msg); err == nil || err.Error() != errExp {
		t.Errorf("Expected %v but received %v", errExp, err)
	}
}

func TestSQSERReadLoop(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	rdr := &SQSER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     new(engine.FilterS),
		rdrEvents: make(chan *erEvent, 1),
		rdrExit:   make(chan struct{}, 1),
		rdrErr:    make(chan error, 1),
		cap:       make(chan struct{}, 1),
		awsRegion: "us-east-2",
		awsID:     "AWSId",
		awsKey:    "AWSAccessKeyId",
		awsToken:  "",
		queueID:   "cgrates_cdrs",
		queueURL:  utils.StringPointer("testQueueURL"),
		session:   nil,
		poster:    nil,
	}
	rdr.cap <- struct{}{}
	rdr.Config().ConcurrentReqs = 1
	counter := 0
	receiveMessage := func(input *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error) {
		msg := &sqs.ReceiveMessageOutput{
			Messages: []*sqs.Message{
				{
					Body:      utils.StringPointer(`{"msgBody":"BODY"`),
					MessageId: utils.StringPointer(`{"msgId":"MESSAGE"}`),
				},
			},
		}
		if counter == 0 {
			counter++
			return msg, nil
		}
		return nil, utils.ErrNotImplemented
	}
	scv := &sqsClientMock{
		ReceiveMessageF: receiveMessage,
	}
	errExpect := "NOT_IMPLEMENTED"
	if err := rdr.readLoop(scv); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
}

func TestSQSERReadLoop2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	rdr := &SQSER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     new(engine.FilterS),
		rdrEvents: make(chan *erEvent, 1),
		rdrExit:   make(chan struct{}, 1),
		rdrErr:    make(chan error, 1),
		cap:       make(chan struct{}, 1),
		awsRegion: "us-east-2",
		awsID:     "AWSId",
		awsKey:    "AWSAccessKeyId",
		awsToken:  "",
		queueID:   "cgrates_cdrs",
		queueURL:  utils.StringPointer("testQueueURL"),
		session:   nil,
		poster:    nil,
	}
	rdr.cap <- struct{}{}
	rdr.Config().ConcurrentReqs = 1
	counter := 0
	receiveMessage := func(input *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error) {
		msg := &sqs.ReceiveMessageOutput{
			Messages: []*sqs.Message{},
		}
		if counter == 0 {
			counter++
			return msg, nil
		}
		return nil, utils.ErrNotImplemented
	}
	scv := &sqsClientMock{
		ReceiveMessageF: receiveMessage,
	}
	errExpect := "NOT_IMPLEMENTED"
	if err := rdr.readLoop(scv); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
	rdr.rdrExit <- struct{}{}
	if err := rdr.readLoop(scv); err != nil {
		t.Error(err)
	}
}

func TestSQSERGetQueueURL(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	rdr := &SQSER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     new(engine.FilterS),
		rdrEvents: make(chan *erEvent, 1),
		rdrExit:   make(chan struct{}, 1),
		rdrErr:    make(chan error, 1),
		cap:       nil,
		awsRegion: "us-east-2",
		awsID:     "AWSId",
		awsKey:    "AWSAccessKeyId",
		awsToken:  "",
		queueID:   "cgrates_cdrs",
		session:   nil,
		poster:    nil,
	}
	// scv := &sqsClientMock{}
	rdr.queueURL = utils.StringPointer("queueURL")
	if err := rdr.getQueueURL(); err != nil {
		t.Error(err)
	}
}

func TestSQSERGetQueueURLWithClient(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	rdr := &SQSER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     new(engine.FilterS),
		rdrEvents: make(chan *erEvent, 1),
		rdrExit:   make(chan struct{}, 1),
		rdrErr:    make(chan error, 1),
		cap:       nil,
		awsRegion: "us-east-2",
		awsID:     "AWSId",
		awsKey:    "AWSAccessKeyId",
		awsToken:  "",
		queueID:   "cgrates_cdrs",
		session:   nil,
		poster:    nil,
	}
	getQueueUrl := func(input *sqs.GetQueueUrlInput) (*sqs.GetQueueUrlOutput, error) {
		output := &sqs.GetQueueUrlOutput{
			QueueUrl: utils.StringPointer("queueURL"),
		}
		return output, nil
	}
	scv := &sqsClientMock{
		GetQueueUrlF: getQueueUrl,
	}
	// rdr.queueURL = utils.StringPointer("queueURL")
	if err := rdr.getQueueURLWithClient(scv); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(rdr.queueURL, utils.StringPointer("queueURL")) {
		t.Errorf("Expected %v but received %v", "queueURL", rdr.queueURL)
	}
}

type awserrMock struct {
	error
}

func (awserrMock) Code() string {
	return sqs.ErrCodeQueueDoesNotExist
}

func (awserrMock) Message() string {
	return ""
}

func (awserrMock) OrigErr() error {
	return utils.ErrNotImplemented
}

func TestSQSERGetQueueURLWithClient2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	rdr := &SQSER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     new(engine.FilterS),
		rdrEvents: make(chan *erEvent, 1),
		rdrExit:   make(chan struct{}, 1),
		rdrErr:    make(chan error, 1),
		cap:       nil,
		awsRegion: "us-east-2",
		awsID:     "AWSId",
		awsKey:    "AWSAccessKeyId",
		awsToken:  "",
		queueID:   "cgrates_cdrs",
		session:   nil,
		poster:    nil,
	}
	getQueueUrl := func(input *sqs.GetQueueUrlInput) (output *sqs.GetQueueUrlOutput, err error) {
		output = &sqs.GetQueueUrlOutput{
			QueueUrl: utils.StringPointer("queueURL"),
		}
		aerr := &awserrMock{}
		return output, aerr
	}
	createQueue := func(input *sqs.CreateQueueInput) (*sqs.CreateQueueOutput, error) {
		output := &sqs.CreateQueueOutput{
			QueueUrl: utils.StringPointer("queueURL"),
		}
		return output, nil
	}
	scv := &sqsClientMock{
		GetQueueUrlF: getQueueUrl,
		CreateQueueF: createQueue,
	}
	// rdr.queueURL = utils.StringPointer("queueURL")
	if err := rdr.getQueueURLWithClient(scv); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(rdr.queueURL, utils.StringPointer("queueURL")) {
		t.Errorf("Expected %v but received %v", "queueURL", rdr.queueURL)
	}
}
