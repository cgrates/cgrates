// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ers

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestNewInvalidReader(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	reader := cfg.ERsCfg().Readers[0]
	reader.Type = "Invalid"
	reader.ID = "InvaidReader"
	cfg.ERsCfg().Readers = append(cfg.ERsCfg().Readers, reader)
	if len(cfg.ERsCfg().Readers) != 2 {
		t.Errorf("Expecting: <2>, received: <%+v>", len(cfg.ERsCfg().Readers))
	}
	if _, err := NewEventReader(cfg, 1, nil, nil, &engine.FilterS{}, nil); err == nil || err.Error() != "unsupported reader type: <Invalid>" {
		t.Errorf("Expecting: <unsupported reader type: <Invalid>>, received: <%+v>", err)
	}
}

func TestNewCsvReader(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	fltr := &engine.FilterS{}
	reader := cfg.ERsCfg().Readers[0]
	reader.Type = utils.MetaFileCSV
	reader.ID = "file_reader"
	cfg.ERsCfg().Readers = append(cfg.ERsCfg().Readers, reader)
	if len(cfg.ERsCfg().Readers) != 2 {
		t.Errorf("Expecting: <2>, received: <%+v>", len(cfg.ERsCfg().Readers))
	}
	exp := &CSVFileER{
		cgrCfg:    cfg,
		cfgIdx:    1,
		fltrS:     fltr,
		rdrDir:    cfg.ERsCfg().Readers[1].SourcePath,
		rdrEvents: nil,
		rdrError:  nil,
		rdrExit:   nil,
		conReqs:   nil}
	var expected EventReader = exp
	if rcv, err := NewEventReader(cfg, 1, nil, nil, fltr, nil); err != nil {
		t.Errorf("Expecting: <nil>, received: <%+v>", err)
	} else {
		// because we use function make to init the channel when we create the EventReader reflect.DeepEqual
		// says it doesn't match
		rcv.(*CSVFileER).conReqs = nil
		if !reflect.DeepEqual(expected, rcv) {
			t.Errorf("Expecting: <%+v>, received: <%+v>", expected, rcv)
		}
	}
}

func TestNewKafkaReader(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	fltr := &engine.FilterS{}
	reader := cfg.ERsCfg().Readers[0]
	reader.Type = utils.MetaKafkajsonMap
	reader.ID = "file_reader"
	reader.ConcurrentReqs = -1
	cfg.ERsCfg().Readers = append(cfg.ERsCfg().Readers, reader)
	if len(cfg.ERsCfg().Readers) != 2 {
		t.Errorf("Expecting: <2>, received: <%+v>", len(cfg.ERsCfg().Readers))
	}
	expected, err := NewKafkaER(cfg, 1, nil, nil, fltr, nil)
	if err != nil {
		t.Errorf("Expecting: <nil>, received: <%+v>", err)
	}
	if rcv, err := NewEventReader(cfg, 1, nil, nil, fltr, nil); err != nil {
		t.Errorf("Expecting: <nil>, received: <%+v>", err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: <%+v>, received: <%+v>", expected, rcv)
	}
}

func TestNewSQLReader(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	fltr := &engine.FilterS{}
	reader := cfg.ERsCfg().Readers[0]
	reader.Type = utils.MetaSQL
	reader.ID = "file_reader"
	reader.ConcurrentReqs = -1
	reader.SourcePath = "*mysql://cgrates:CGRateS.org@127.0.0.1:3306?db_name=cgrates2"
	cfg.ERsCfg().Readers = append(cfg.ERsCfg().Readers, reader)
	if len(cfg.ERsCfg().Readers) != 2 {
		t.Errorf("Expecting: <2>, received: <%+v>", len(cfg.ERsCfg().Readers))
	}
	expected, err := NewSQLEventReader(cfg, 1, nil, nil, fltr, nil)
	if err != nil {
		t.Errorf("Expecting: <nil>, received: <%+v>", err)
	}
	if rcv, err := NewEventReader(cfg, 1, nil, nil, fltr, nil); err != nil {
		t.Errorf("Expecting: <nil>, received: <%+v>", err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: <%+v>, received: <%+v>", expected, rcv)
	}
}
