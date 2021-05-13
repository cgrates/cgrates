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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestNewInvalidReader(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	reader := cfg.ERsCfg().Readers[0]
	reader.Type = "Invalid"
	reader.ID = "InvaidReader"
	cfg.ERsCfg().Readers = append(cfg.ERsCfg().Readers, reader)
	if len(cfg.ERsCfg().Readers) != 2 {
		t.Errorf("Expecting: <2>, received: <%+v>", len(cfg.ERsCfg().Readers))
	}
	if _, err := NewEventReader(cfg, 1, nil, nil, nil, &engine.FilterS{}, nil); err == nil || err.Error() != "unsupported reader type: <Invalid>" {
		t.Errorf("Expecting: <unsupported reader type: <Invalid>>, received: <%+v>", err)
	}
}

func TestNewCsvReader(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
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
	if rcv, err := NewEventReader(cfg, 1, nil, nil, nil, fltr, nil); err != nil {
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
	cfg := config.NewDefaultCGRConfig()
	fltr := &engine.FilterS{}
	reader := cfg.ERsCfg().Readers[0]
	reader.Type = utils.MetaKafkajsonMap
	reader.ID = "file_reader"
	reader.ConcurrentReqs = -1
	cfg.ERsCfg().Readers = append(cfg.ERsCfg().Readers, reader)
	if len(cfg.ERsCfg().Readers) != 2 {
		t.Errorf("Expecting: <2>, received: <%+v>", len(cfg.ERsCfg().Readers))
	}
	expected, err := NewKafkaER(cfg, 1, nil, nil, nil, fltr, nil)
	if err != nil {
		t.Errorf("Expecting: <nil>, received: <%+v>", err)
	}
	if rcv, err := NewEventReader(cfg, 1, nil, nil, nil, fltr, nil); err != nil {
		t.Errorf("Expecting: <nil>, received: <%+v>", err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: <%+v>, received: <%+v>", expected, rcv)
	}
}

func TestNewSQLReader(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fltr := &engine.FilterS{}
	reader := cfg.ERsCfg().Readers[0].Clone()
	reader.Type = utils.MetaSQL
	reader.ID = "file_reader"
	reader.ConcurrentReqs = -1
	reader.Opts = map[string]interface{}{"db_name": "cgrates2"}
	reader.SourcePath = "*mysql://cgrates:CGRateS.org@127.0.0.1:3306"
	reader.ProcessedPath = ""
	cfg.ERsCfg().Readers = append(cfg.ERsCfg().Readers, reader)
	if len(cfg.ERsCfg().Readers) != 2 {
		t.Errorf("Expecting: <2>, received: <%+v>", len(cfg.ERsCfg().Readers))
	}
	expected, err := NewSQLEventReader(cfg, 1, nil, nil, nil, fltr, nil)
	if err != nil {
		t.Errorf("Expecting: <nil>, received: <%+v>", err)
	}
	if rcv, err := NewEventReader(cfg, 1, nil, nil, nil, fltr, nil); err != nil {
		t.Errorf("Expecting: <nil>, received: <%+v>", err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: <%+v>, received: <%+v>", expected, rcv)
	}
}

func TestNewSQLReaderError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fltr := &engine.FilterS{}
	reader := cfg.ERsCfg().Readers[0].Clone()
	reader.Type = utils.MetaSQL
	reader.ID = "file_reader"
	reader.ConcurrentReqs = -1
	reader.Opts = map[string]interface{}{"db_name": "cgrates2"}
	reader.SourcePath = "#"
	reader.ProcessedPath = ""
	expected := "unknown db_type "
	_, err := NewSQLEventReader(cfg, 0, nil, nil, nil, fltr, nil)
	if err == nil || err.Error() != expected {
		t.Errorf("Expecting: <%+v>, received: <%+v>", expected, err)
	}
}

func TestNewFileXMLReader(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fltr := &engine.FilterS{}
	cfg.ERsCfg().Readers[0].Type = utils.MetaFileXML
	expected, err := NewXMLFileER(cfg, 0, nil, nil, nil, fltr, nil)
	if err != nil {
		t.Error(err)
	}
	rcv, err := NewEventReader(cfg, 0, nil, nil, nil, fltr, nil)
	if err != nil {
		t.Error(err)
	} else {
		rcv.(*XMLFileER).conReqs = nil
		expected.(*XMLFileER).conReqs = nil
		if !reflect.DeepEqual(expected, rcv) {
			t.Errorf("Expecting %v but received %v", expected, rcv)
		}
	}
}

func TestNewFileFWVReader(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fltr := &engine.FilterS{}
	cfg.ERsCfg().Readers[0].Type = utils.MetaFileFWV
	expected, err := NewFWVFileER(cfg, 0, nil, nil, nil, fltr, nil)
	if err != nil {
		t.Error(err)
	}
	rcv, err := NewEventReader(cfg, 0, nil, nil, nil, fltr, nil)
	if err != nil {
		t.Error(nil)
	} else {
		rcv.(*FWVFileER).conReqs = nil
		expected.(*FWVFileER).conReqs = nil
		if !reflect.DeepEqual(expected, rcv) {
			t.Errorf("Expecting %v but received %v", expected, rcv)
		}
	}
}

func TestNewJSONReader(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fltr := &engine.FilterS{}
	cfg.ERsCfg().Readers[0].Type = utils.MetaFileJSON
	expected, err := NewJSONFileER(cfg, 0, nil, nil, nil, fltr, nil)
	if err != nil {
		t.Error(err)
	}
	rcv, err := NewEventReader(cfg, 0, nil, nil, nil, fltr, nil)
	if err != nil {
		t.Error(err)
	} else {
		rcv.(*JSONFileER).conReqs = nil
		expected.(*JSONFileER).conReqs = nil
		if !reflect.DeepEqual(expected, rcv) {
			t.Errorf("Expecting %v but received %v", expected, rcv)
		}
	}
}

func TestNewAMQPReader(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fltr := &engine.FilterS{}
	cfg.ERsCfg().Readers[0].Type = utils.MetaAMQPjsonMap
	cfg.ERsCfg().Readers[0].ConcurrentReqs = -1
	exp := &AMQPER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltr,
		rdrEvents: nil,
		rdrExit:   nil,
		rdrErr:    nil,
	}
	exp.dialURL = exp.Config().SourcePath
	exp.Config().ProcessedPath = ""
	exp.setOpts(map[string]interface{}{})
	exp.createPoster()
	var expected EventReader = exp
	rcv, err := NewEventReader(cfg, 0, nil, nil, nil, fltr, nil)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected %v but received %v", expected, rcv)
	}
}

func TestNewAMQPv1Reader(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fltr := &engine.FilterS{}
	cfg.ERsCfg().Readers[0].Type = utils.MetaAMQPV1jsonMap
	cfg.ERsCfg().Readers[0].ConcurrentReqs = -1
	exp := &AMQPv1ER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltr,
		rdrEvents: nil,
		rdrExit:   nil,
		rdrErr:    nil,
	}
	exp.Config().ProcessedPath = ""
	exp.Config().Opts = map[string]interface{}{}
	exp.createPoster()
	var expected EventReader = exp
	rcv, err := NewEventReader(cfg, 0, nil, nil, nil, fltr, nil)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected \n%v but received \n%v", expected, rcv)
	}
}

func TestNewS3Reader(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fltr := &engine.FilterS{}
	cfg.ERsCfg().Readers[0].Type = utils.MetaS3jsonMap
	cfg.ERsCfg().Readers[0].ConcurrentReqs = -1
	exp := &S3ER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltr,
		rdrEvents: nil,
		rdrExit:   nil,
		rdrErr:    nil,
		bucket:    "cgrates_cdrs",
	}
	exp.Config().ProcessedPath = ""
	exp.Config().Opts = map[string]interface{}{}
	exp.createPoster()
	var expected EventReader = exp
	rcv, err := NewEventReader(cfg, 0, nil, nil, nil, fltr, nil)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected \n%v but received \n%v", expected, rcv)
	}
}

func TestNewSQSReader(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fltr := &engine.FilterS{}
	cfg.ERsCfg().Readers[0].Type = utils.MetaSQSjsonMap
	cfg.ERsCfg().Readers[0].ConcurrentReqs = -1
	exp := &SQSER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltr,
		rdrEvents: nil,
		rdrExit:   nil,
		rdrErr:    nil,
		queueID:   "cgrates_cdrs",
	}
	exp.Config().SourcePath = "string"
	// var err error
	// awsCfg := aws.Config{Endpoint: aws.String(exp.Config().SourcePath)}
	// exp.session, err = session.NewSessionWithOptions(
	// 	session.Options{
	// 		Config: awsCfg,
	// 	},
	// )
	// if err != nil {
	// 	t.Error(err)
	// }
	exp.Config().ProcessedPath = ""
	exp.Config().Opts = map[string]interface{}{}
	exp.createPoster()
	var expected EventReader = exp
	rcv, err := NewEventReader(cfg, 0, nil, nil, nil, fltr, nil)
	exp.session = rcv.(*SQSER).session
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected \n%v but received \n%v", expected, rcv)
	}
}
