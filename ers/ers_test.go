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

func TestERsNewERService(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	fltrS := &engine.FilterS{}
	expected := &ERService{cfg: cfg,
		filterS:   fltrS,
		rdrs:      make(map[string]EventReader),
		rdrPaths:  make(map[string]string),
		stopLsn:   make(map[string]chan struct{}),
		rdrEvents: make(chan *erEvent),
		rdrErr:    make(chan error),
		stopChan:  nil}
	rcv := NewERService(cfg, fltrS, nil, nil)

	if !reflect.DeepEqual(expected.cfg, rcv.cfg) {
		t.Errorf("Expecting: <%+v>, received: <%+v>", expected.cfg, rcv.cfg)
	} else if !reflect.DeepEqual(expected.filterS, rcv.filterS) {
		t.Errorf("Expecting: <%+v>, received: <%+v>", expected.filterS, rcv.filterS)
	}
}

func TestERsAddReader(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	fltrS := &engine.FilterS{}
	erS := NewERService(cfg, fltrS, nil, nil)
	reader := cfg.ERsCfg().Readers[0]
	reader.Type = utils.MetaFileCSV
	reader.ID = "file_reader"
	reader.RunDelay = time.Duration(0)
	cfg.ERsCfg().Readers = append(cfg.ERsCfg().Readers, reader)
	if len(cfg.ERsCfg().Readers) != 2 {
		t.Errorf("Expecting: <2>, received: <%+v>", len(cfg.ERsCfg().Readers))
	}
	if err := erS.addReader("file_reader", 1); err != nil {
		t.Errorf("Expecting: <nil>, received: <%+v>", len(cfg.ERsCfg().Readers))
	} else if len(erS.rdrs) != 1 {
		t.Errorf("Expecting: <2>, received: <%+v>", len(erS.rdrs))
	} else if !reflect.DeepEqual(erS.rdrs["file_reader"].Config(), reader) {
		t.Errorf("Expecting: <%+v>, received: <%+v>", reader, erS.rdrs["file_reader"].Config())
	}
}
