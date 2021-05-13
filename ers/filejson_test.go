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
	"sync"
	"testing"

	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/config"
)

func TestNewJSONFileER(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfgIdx := 0
	expected := &JSONFileER{
		RWMutex:   sync.RWMutex{},
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     nil,
		rdrEvents: nil,
		rdrError:  nil,
		rdrExit:   nil,
		conReqs:   nil,
	}
	cfg.ERsCfg().Readers[0].ConcurrentReqs = 1
	cfg.ERsCfg().Readers[0].SourcePath = "/"
	result, err := NewJSONFileER(cfg, cfgIdx, nil, nil, nil, nil, nil)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	result.(*JSONFileER).conReqs = nil
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestFileJSONConfig(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfgIdx := 0
	cfg.ERsCfg().Readers[cfgIdx] = &config.EventReaderCfg{
		ID:             utils.MetaDefault,
		Type:           utils.MetaNone,
		RunDelay:       0,
		ConcurrentReqs: 1024,
		SourcePath:     "/var/spool/cgrates/ers/in",
		ProcessedPath:  "/var/spool/cgrates/ers/out",
		Tenant:         nil,
		Timezone:       utils.EmptyString,
		Filters:        []string{},
		Flags:          utils.FlagsWithParams{},
		Fields:         []*config.FCTemplate{},
	}
	rdr := &JSONFileER{
		cgrCfg: cfg,
		cfgIdx: cfgIdx,
	}
	expected := &config.EventReaderCfg{
		ID:             utils.MetaDefault,
		Type:           utils.MetaNone,
		RunDelay:       0,
		ConcurrentReqs: 1024,
		SourcePath:     "/var/spool/cgrates/ers/in",
		ProcessedPath:  "/var/spool/cgrates/ers/out",
		Tenant:         nil,
		Timezone:       utils.EmptyString,
		Filters:        []string{},
		Flags:          utils.FlagsWithParams{},
		Fields:         []*config.FCTemplate{},
	}
	result := rdr.Config()
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}
