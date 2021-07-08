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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestNewNatsER(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfgIdx := 0
	expected := &NatsER{
		cgrCfg:        cfg,
		cfgIdx:        cfgIdx,
		fltrS:         nil,
		rdrEvents:     nil,
		partialEvents: nil,
		rdrExit:       nil,
		rdrErr:        nil,
	}
	expected.Config().ConcurrentReqs = -1
	if err := expected.createPoster(); err != nil {
		return
	}
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
		Event: map[string]interface{}{
			utils.ToR: "*voice",
		},
		APIOpts: map[string]interface{}{},
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
	data := engine.NewInternalDB(nil, nil, true)
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
