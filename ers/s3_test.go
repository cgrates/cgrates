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
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestS3ERServe(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	rdr, err := NewS3ER(cfg, 0, nil,
		nil, nil, nil)
	if err != nil {
		t.Error(err)
	}
	rdr.Config().RunDelay = 1 * time.Millisecond
	if err := rdr.Serve(); err != nil {
		t.Error(err)
	}
}

func TestS3ERServe2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	rdr := &S3ER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     nil,
		rdrEvents: nil,
		rdrExit:   nil,
		rdrErr:    nil,
		cap:       nil,
		awsRegion: "us-east-2",
		awsID:     "AWSId",
		awsKey:    "AWSAccessKeyId",
		awsToken:  "",
		queueID:   "cgrates_cdrs",
		session:   nil,
		poster:    nil,
	}
	if err := rdr.Serve(); err != nil {
		t.Error(err)
	}
}

func TestS3ERProcessMessage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	rdr := &S3ER{
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

func TestS3ERProcessMessageError1(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	rdr := &S3ER{
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

func TestS3ERProcessMessageError2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	cfg.ERsCfg().Readers[0].ProcessedPath = ""
	fltrs := engine.NewFilterS(cfg, nil, dm)
	rdr := &S3ER{
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

func TestS3ERProcessMessageError3(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	rdr := &S3ER{
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

func TestS3ERParseOpts(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	rdr := &S3ER{
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
		utils.QueueID:   "QueueID",
		utils.AWSRegion: "AWSRegion",
		utils.AWSKey:    "AWSKey",
		utils.AWSSecret: "AWSSecret",
		utils.AWSToken:  "AWSToken",
	}
	rdr.parseOpts(opts)
	if rdr.queueID != opts[utils.QueueID] || rdr.awsRegion != opts[utils.AWSRegion] || rdr.awsID != opts[utils.AWSKey] || rdr.awsKey != opts[utils.AWSSecret] || rdr.awsToken != opts[utils.AWSToken] {
		t.Error("Fields do not corespond")
	}
	rdr.Config().Opts = map[string]interface{}{}
	rdr.Config().ProcessedPath = utils.EmptyString
	rdr.createPoster()
}

func TestS3ERIsClosed(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	rdr := &S3ER{
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

type s3ClientMock struct {
	ListObjectsV2PagesF func(input *s3.ListObjectsV2Input, fn func(*s3.ListObjectsV2Output, bool) bool) error
	GetObjectF          func(input *s3.GetObjectInput) (*s3.GetObjectOutput, error)
	DeleteObjectF       func(input *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error)
}

func (s *s3ClientMock) ListObjectsV2Pages(input *s3.ListObjectsV2Input, fn func(*s3.ListObjectsV2Output, bool) bool) error {
	if s.ListObjectsV2PagesF != nil {
		return s.ListObjectsV2PagesF(input, fn)
	}
	return utils.ErrNotFound
}

func (s *s3ClientMock) GetObject(input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	if s.GetObjectF != nil {
		return s.GetObjectF(input)
	}
	return nil, utils.ErrNotImplemented
}

func (s *s3ClientMock) DeleteObject(input *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error) {
	// return nil, nil
	if s.DeleteObjectF != nil {
		return s.DeleteObjectF(input)
	}
	return nil, utils.ErrInvalidPath
}

func TestS3ERReadLoop(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	rdr := &S3ER{
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
	listObjects := func(input *s3.ListObjectsV2Input, fn func(*s3.ListObjectsV2Output, bool) bool) error {
		return nil
	}
	scv := &s3ClientMock{
		ListObjectsV2PagesF: listObjects,
	}
	if err := rdr.readLoop(scv); err != nil {
		t.Error(err)
	}
}

func TestS3ERReadLoopIsClosed(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	rdr := &S3ER{
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
	listObjects := func(input *s3.ListObjectsV2Input, fn func(*s3.ListObjectsV2Output, bool) bool) error {
		return nil
	}
	scv := &s3ClientMock{
		ListObjectsV2PagesF: listObjects,
	}
	rdr.rdrExit <- struct{}{}
	if err := rdr.readLoop(scv); err != nil {
		t.Error(err)
	}
}

func TestS3ERReadMsg(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	rdr := &S3ER{
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
	// rdr.poster = engine.NewS3Poster(rdr.Config().SourcePath, 1, make(map[string]interface{}))
	rdr.Config().SourcePath = rdr.awsRegion
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

	listObjects := func(input *s3.ListObjectsV2Input, fn func(*s3.ListObjectsV2Output, bool) bool) error {
		return nil
	}
	getObject := func(input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
		return &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewBuffer([]byte(`{"key":"value"}`)))}, nil
	}
	deleteObject := func(input *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error) {
		return nil, nil
	}
	scv := &s3ClientMock{
		ListObjectsV2PagesF: listObjects,
		GetObjectF:          getObject,
		DeleteObjectF:       deleteObject,
	}
	if err := rdr.readMsg(scv, "AWSKey"); err != nil {
		t.Error(err)
	}
}

func TestS3ERReadMsgError1(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	rdr := &S3ER{
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
		session:   nil,
		poster:    nil,
	}
	rdr.Config().ConcurrentReqs = 1
	listObjects := func(input *s3.ListObjectsV2Input, fn func(*s3.ListObjectsV2Output, bool) bool) error {
		return nil
	}
	getObject := func(input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
		return &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewBuffer([]byte(`{"key":"value"}`)))}, nil
	}
	deleteObject := func(input *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error) {
		return nil, nil
	}
	scv := &s3ClientMock{
		ListObjectsV2PagesF: listObjects,
		GetObjectF:          getObject,
		DeleteObjectF:       deleteObject,
	}
	rdr.cap <- struct{}{}
	errExp := "NOT_FOUND:ToR"
	if err := rdr.readMsg(scv, "AWSKey"); err == nil || err.Error() != errExp {
		t.Errorf("Expected %v but received %v", errExp, err)
	}
}

func TestS3ERReadMsgError2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	rdr := &S3ER{
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
		session:   nil,
		poster:    nil,
	}
	rdr.Config().ConcurrentReqs = 1
	scv := &s3ClientMock{}
	rdr.cap <- struct{}{}
	rdr.rdrExit <- struct{}{}
	if err := rdr.readMsg(scv, "AWSKey"); err != nil {
		t.Error(err)
	}
}

func TestS3ERReadMsgError3(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	rdr := &S3ER{
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
		session:   nil,
		poster:    nil,
	}
	rdr.Config().ConcurrentReqs = -1
	scv := &s3ClientMock{}
	errExp := "NOT_IMPLEMENTED"
	if err := rdr.readMsg(scv, "AWSKey"); err == nil || err.Error() != errExp {
		t.Errorf("Expected %v but received %v", errExp, err)
	}
}

func TestS3ERReadMsgError4(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	rdr := &S3ER{
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
		session:   nil,
		poster:    nil,
	}
	rdr.Config().SourcePath = rdr.awsRegion
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
	listObjects := func(input *s3.ListObjectsV2Input, fn func(*s3.ListObjectsV2Output, bool) bool) error {
		return nil
	}
	getObject := func(input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
		return &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewBuffer([]byte(`{"key":"value"}`)))}, nil
	}
	deleteObject := func(input *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error) {
		return nil, utils.ErrInvalidPath
	}
	scv := &s3ClientMock{
		ListObjectsV2PagesF: listObjects,
		GetObjectF:          getObject,
		DeleteObjectF:       deleteObject,
	}
	errExp := "INVALID_PATH"
	if err := rdr.readMsg(scv, "AWSKey"); err == nil || err.Error() != errExp {
		t.Errorf("Expected %v but received %v", errExp, err)
	}
}
