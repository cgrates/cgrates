// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ers

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestKafkasetOpts(t *testing.T) {
	k := new(KafkaER)
	k.dialURL = "localhost:2013"
	expKafka := &KafkaER{
		dialURL: "localhost:2013",
		topic:   "cdrs",
		groupID: "new",
		maxWait: time.Second,
	}
	if err := k.setOpts(&config.EventReaderOpts{
		KafkaTopic:   utils.StringPointer("cdrs"),
		KafkaGroupID: utils.StringPointer("new"),
		KafkaMaxWait: utils.DurationPointer(time.Second),
	}); err != nil {
		t.Fatal(err)
	} else if expKafka.dialURL != k.dialURL {
		t.Errorf("Expected: %s ,received: %s", expKafka.dialURL, k.dialURL)
	} else if expKafka.topic != k.topic {
		t.Errorf("Expected: %s ,received: %s", expKafka.topic, k.topic)
	} else if expKafka.groupID != k.groupID {
		t.Errorf("Expected: %s ,received: %s", expKafka.groupID, k.groupID)
	} else if expKafka.maxWait != k.maxWait {
		t.Errorf("Expected: %s ,received: %s", expKafka.maxWait, k.maxWait)
	}
	k = new(KafkaER)
	k.dialURL = "localhost:2013"
	expKafka = &KafkaER{
		dialURL: "localhost:2013",
		topic:   "cgrates",
		groupID: "cgrates",
		maxWait: time.Millisecond,
	}
	if err := k.setOpts(&config.EventReaderOpts{}); err != nil {
		t.Fatal(err)
	} else if expKafka.dialURL != k.dialURL {
		t.Errorf("Expected: %s ,received: %s", expKafka.dialURL, k.dialURL)
	} else if expKafka.topic != k.topic {
		t.Errorf("Expected: %s ,received: %s", expKafka.topic, k.topic)
	} else if expKafka.groupID != k.groupID {
		t.Errorf("Expected: %s ,received: %s", expKafka.groupID, k.groupID)
	} else if expKafka.maxWait != k.maxWait {
		t.Errorf("Expected: %s ,received: %s", expKafka.maxWait, k.maxWait)
	}

	k = new(KafkaER)
	k.dialURL = "127.0.0.1:2013"
	expKafka = &KafkaER{
		dialURL: "127.0.0.1:2013",
		topic:   "cdrs",
		groupID: "new",
		maxWait: time.Second,
	}
	if err := k.setOpts(&config.EventReaderOpts{
		KafkaTopic:   utils.StringPointer("cdrs"),
		KafkaGroupID: utils.StringPointer("new"),
		KafkaMaxWait: utils.DurationPointer(time.Second),
	}); err != nil {
		t.Fatal(err)
	} else if expKafka.dialURL != k.dialURL {
		t.Errorf("Expected: %s ,received: %s", expKafka.dialURL, k.dialURL)
	} else if expKafka.topic != k.topic {
		t.Errorf("Expected: %s ,received: %s", expKafka.topic, k.topic)
	} else if expKafka.groupID != k.groupID {
		t.Errorf("Expected: %s ,received: %s", expKafka.groupID, k.groupID)
	} else if expKafka.maxWait != k.maxWait {
		t.Errorf("Expected: %s ,received: %s", expKafka.maxWait, k.maxWait)
	}
}

func TestKafkaERServe(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	fltrS := new(engine.FilterS)
	rdrEvents := make(chan *erEvent, 1)
	rdrExit := make(chan struct{}, 1)
	rdrErr := make(chan error, 1)
	rdr, err := NewKafkaER(cfg, 0, rdrEvents, make(chan *erEvent, 1), rdrErr, cacheS, fltrS, rdrExit)
	if err != nil {
		t.Error(err)
	}
	if err := rdr.Serve(); err != nil {
		t.Error(err)
	}
	rdr.Config().RunDelay = 1 * time.Millisecond
	if err := rdr.Serve(); err != nil {
		t.Error(err)
	}
	rdr.Config().Opts = &config.EventReaderOpts{}
	rdr.Config().ProcessedPath = ""
	close(rdrExit)
}

func TestKafkaERServe2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	rdr := &KafkaER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     new(engine.FilterS),
		rdrEvents: make(chan *erEvent, 1),
		rdrExit:   make(chan struct{}, 1),
		rdrErr:    make(chan error, 1),
		dialURL:   "testURL",
		groupID:   "testGroupID",
		topic:     "testTopic",
		maxWait:   time.Duration(1),
		cap:       make(chan struct{}, 1),
	}
	rdr.rdrExit <- struct{}{}
	rdr.Config().RunDelay = 1 * time.Millisecond
	if err := rdr.Serve(); err != nil {
		t.Error(err)
	}
}

func TestKafkaERProcessMessage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	rdr := &KafkaER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     new(engine.FilterS),
		rdrEvents: make(chan *erEvent, 1),
		rdrExit:   make(chan struct{}, 1),
		rdrErr:    make(chan error, 1),
		dialURL:   "testURL",
		groupID:   "testGroupID",
		topic:     "testTopic",
		maxWait:   time.Duration(1),
		cap:       make(chan struct{}, 1),
	}
	expEvent := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.ToR: "*voice",
		},
		APIOpts: map[string]any{},
	}
	rdr.Config().Fields = []*config.FCTemplate{
		{
			Tag:   "Tor",
			Type:  utils.MetaConstant,
			Value: utils.NewRSRParsersMustCompile("*voice", utils.InfieldSep),
			Path:  "*cgreq.ToR",
		},
	}
	rdr.Config().Fields[0].ComputePath()

	msg := []byte(`{"test":"input"}`)
	if err := rdr.processMessage(msg); err != nil {
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

func TestKafkaERProcessMessageError1(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	rdr := &KafkaER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     new(engine.FilterS),
		rdrEvents: make(chan *erEvent, 1),
		rdrExit:   make(chan struct{}, 1),
		rdrErr:    make(chan error, 1),
		dialURL:   "testURL",
		groupID:   "testGroupID",
		topic:     "testTopic",
		maxWait:   time.Duration(1),
		cap:       make(chan struct{}, 1),
	}
	rdr.Config().Fields = []*config.FCTemplate{
		{},
	}
	msg := []byte(`{"test":"input"}`)
	errExpect := "unsupported type: <>"
	if err := rdr.processMessage(msg); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
}

func TestKafkaERProcessMessageError2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dm.SetCache(cacheS)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	rdr := &KafkaER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		rdrEvents: make(chan *erEvent, 1),
		rdrExit:   make(chan struct{}, 1),
		rdrErr:    make(chan error, 1),
		dialURL:   "testURL",
		groupID:   "testGroupID",
		topic:     "testTopic",
		maxWait:   time.Duration(1),
		cap:       make(chan struct{}, 1),
	}
	rdr.Config().Filters = []string{"Filter1"}
	msg := []byte(`{"test":"input"}`)
	errExpect := "NOT_FOUND:Filter1"
	if err := rdr.processMessage(msg); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
}

func TestKafkaERProcessMessageError3(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	rdr := &KafkaER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     new(engine.FilterS),
		rdrEvents: make(chan *erEvent, 1),
		rdrExit:   make(chan struct{}, 1),
		rdrErr:    make(chan error, 1),
		dialURL:   "testURL",
		groupID:   "testGroupID",
		topic:     "testTopic",
		maxWait:   time.Duration(1),
		cap:       make(chan struct{}, 1),
	}
	msg := []byte(`{"invalid":"input"`)
	errExpect := "unexpected end of JSON input"
	if err := rdr.processMessage(msg); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
}
