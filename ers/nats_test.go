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

func TestNewNatsER(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfgIdx := 0
	expected := &NatsER{
		cgrCfg: cfg,
		cfgIdx: cfgIdx,
	}
	expected.Config().ConcurrentReqs = -1
	rdr, err := NewNatsER(cfg, cfgIdx, nil, nil, nil, nil, nil)
	if err != nil {
		t.Error(err)
	}
	rdr.(*NatsER).opts = nil
	if !reflect.DeepEqual(expected.opts, rdr.(*NatsER).opts) {
		t.Errorf("Expected <%+v> \n but received \n <%+v>", expected.opts, rdr.(*NatsER).opts)
	}
}

func TestNatsERProcessMessage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	rdr := &NatsER{
		cgrCfg:        cfg,
		cfgIdx:        0,
		fltrS:         new(engine.FilterS),
		rdrEvents:     make(chan *erEvent, 1),
		partialEvents: make(chan *erEvent, 1),
		rdrExit:       make(chan struct{}, 1),
		rdrErr:        make(chan error, 1),
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
			Value: config.NewRSRParsersMustCompile("*voice", utils.InfieldSep),
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
		expEvent.Time = data.cgrEvent.Time
		if !reflect.DeepEqual(data.cgrEvent, expEvent) {
			t.Errorf("Expected %v but received %v", utils.ToJSON(expEvent), utils.ToJSON(data.cgrEvent))
		}
	case <-time.After(50 * time.Millisecond):
		t.Error("Time limit exceeded")
	}
}

func TestNatsERProcessMessageError1(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	rdr := &NatsER{
		cgrCfg:        cfg,
		cfgIdx:        0,
		fltrS:         new(engine.FilterS),
		rdrEvents:     make(chan *erEvent, 1),
		partialEvents: make(chan *erEvent, 1),
		rdrExit:       make(chan struct{}, 1),
		rdrErr:        make(chan error, 1),
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

func TestNatsERProcessMessageError2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, err := engine.NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	rdr := &NatsER{
		cgrCfg:        cfg,
		cfgIdx:        0,
		fltrS:         fltrs,
		rdrEvents:     make(chan *erEvent, 1),
		partialEvents: make(chan *erEvent, 1),
		rdrExit:       make(chan struct{}, 1),
		rdrErr:        make(chan error, 1),
	}
	rdr.Config().Filters = []string{"Filter1"}
	msg := []byte(`{"test":"input"}`)
	errExpect := "NOT_FOUND:Filter1"
	if err := rdr.processMessage(msg); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
}

func TestNatsERProcessMessageError3(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	rdr := &NatsER{
		cgrCfg:        cfg,
		cfgIdx:        0,
		fltrS:         new(engine.FilterS),
		rdrEvents:     make(chan *erEvent, 1),
		partialEvents: make(chan *erEvent, 1),
		rdrExit:       make(chan struct{}, 1),
		rdrErr:        make(chan error, 1),
	}
	msg := []byte(`{"invalid":"input"`)
	errExpect := "unexpected end of JSON input"
	if err := rdr.processMessage(msg); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
}
