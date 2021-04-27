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
