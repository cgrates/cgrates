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

package ers

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestS3ERServe(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	rdr, err := NewS3ER(cfg, 0, nil, nil, nil, nil, nil)
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
		bucket:    "cgrates_cdrs",
		session:   nil,
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
		bucket:    "cgrates_cdrs",
		session:   nil,
	}
	expEvent := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.Destination: "testdest",
		},
		APIOpts: map[string]any{},
	}
	body := []byte(`{"Destination":"testdest"}`)
	rdr.Config().Fields = []*config.FCTemplate{
		{
			Tag:   "Destination",
			Type:  utils.MetaConstant,
			Value: utils.NewRSRParsersMustCompile("testdest", utils.InfieldSep),
			Path:  "*cgreq.Destination",
		},
	}
	rdr.Config().Fields[0].ComputePath()
	if err := rdr.processMessage(body); err != nil {
		t.Error(err)
	}
	select {
	case data := <-rdr.rdrEvents:
		expEvent.ID = data.cgrEvent.ID
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
		bucket:    "cgrates_cdrs",
		session:   nil,
	}
	rdr.Config().Fields = []*config.FCTemplate{
		{},
	}
	body := []byte(`{"*originID":"testoriginID"}`)
	errExpect := "unsupported type: <>"
	if err := rdr.processMessage(body); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
}

func TestS3ERProcessMessageError2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
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
		bucket:    "cgrates_cdrs",
		session:   nil,
	}
	body := []byte(`{"*originID":"testoriginID"}`)
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
		bucket:    "cgrates_cdrs",
		session:   nil,
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
		bucket:    "cgrates_cdrs",
		session:   nil,
	}

	opts := &config.EventReaderOpts{
		S3BucketID: utils.StringPointer("QueueID"),
		AWSRegion:  utils.StringPointer("AWSRegion"),
		AWSKey:     utils.StringPointer("AWSKey"),
		AWSSecret:  utils.StringPointer("AWSSecret"),
		AWSToken:   utils.StringPointer("AWSToken"),
	}
	rdr.parseOpts(opts)
	if rdr.bucket != *opts.S3BucketID ||
		rdr.awsRegion != *opts.AWSRegion ||
		rdr.awsID != *opts.AWSKey ||
		rdr.awsKey != *opts.AWSSecret ||
		rdr.awsToken != *opts.AWSToken {
		t.Error("Fields do not corespond")
	}
	rdr.Config().Opts = &config.EventReaderOpts{}
	rdr.Config().ProcessedPath = utils.EmptyString
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
		bucket:    "cgrates_cdrs",
		session:   nil,
	}
	if rcv := rdr.isClosed(); rcv != false {
		t.Errorf("Expected %v but received %v", false, true)
	}
	rdr.rdrExit <- struct{}{}
	if rcv := rdr.isClosed(); rcv != true {
		t.Errorf("Expected %v but received %v", true, false)
	}
}
