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

	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/config"
)

func TestNewFlatstoreER(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	expected := &FlatstoreER{
		cgrCfg: cfg,
	}
	cfg.ERsCfg().Readers[0].SourcePath = "/"
	result, err := NewFlatstoreER(cfg, 0, nil, nil, nil, nil)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	result.(*FlatstoreER).cache = nil
	result.(*FlatstoreER).conReqs = nil
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestFlatstoreConfig(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers = []*config.EventReaderCfg{
		{
			ID:             "file_reader1",
			Type:           utils.MetaFileCSV,
			RunDelay:       -1,
			ConcurrentReqs: 1024,
			SourcePath:     "/tmp/ers/in",
			ProcessedPath:  "/tmp/ers/out",
			Tenant:         nil,
			Timezone:       utils.EmptyString,
			Filters:        []string{},
			Flags:          utils.FlagsWithParams{},
			Opts:           map[string]interface{}{utils.FlatstorePrfx + utils.RowLengthOpt: 5},
		},
		{
			ID:             "file_reader2",
			Type:           utils.MetaFileCSV,
			RunDelay:       -1,
			ConcurrentReqs: 1024,
			SourcePath:     "/tmp/ers/in",
			ProcessedPath:  "/tmp/ers/out",
			Tenant:         nil,
			Timezone:       utils.EmptyString,
			Filters:        []string{},
			Flags:          utils.FlagsWithParams{},
			Opts:           map[string]interface{}{utils.FlatstorePrfx + utils.RowLengthOpt: 5},
		},
	}
	expected := cfg.ERsCfg().Readers[0]
	rdr, err := NewFlatstoreER(cfg, 0, nil, nil, nil, nil)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	result := rdr.Config()
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestFlatstoreServeNil(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	result, err := NewFlatstoreER(cfg, 0, nil, nil, nil, nil)
	if err != nil {
		t.Errorf("\nExpected: <%+v>, \nreceived: <%+v>", nil, err)
	}
	expected := &FlatstoreER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     nil,
		cache:     result.(*FlatstoreER).cache,
		rdrDir:    "/var/spool/cgrates/ers/in",
		rdrEvents: nil,
		rdrError:  nil,
		rdrExit:   nil,
		conReqs:   result.(*FlatstoreER).conReqs,
	}
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("\nExpected: <%+v>, \nreceived: <%+v>", expected, result)
	}
	result.Config().RunDelay = time.Duration(0)
	err = result.Serve()
	if err != nil {
		t.Errorf("\nExpected: <%+v>, \nreceived: <%+v>", nil, err)
	}
}
