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

	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/config"
)

func TestERSNewXMLFileER(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	expected := &XMLFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     nil,
		rdrDir:    "/var/spool/cgrates/ers/in",
		rdrEvents: nil,
		rdrError:  nil,
		rdrExit:   nil,
		conReqs:   nil,
	}
	result, err := NewXMLFileER(cfg, 0, nil, nil, nil, nil, nil)
	if err != nil {
		t.Errorf("\nExpected: <%+v>, \nreceived: <%+v>", nil, err)
	}
	expected.conReqs = result.(*XMLFileER).conReqs
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("\nExpected: <%+v>, \nreceived: <%+v>", expected, result)
	}
}

func TestERSXMLFileERConfig(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers[0] = &config.EventReaderCfg{
		ID:             utils.MetaDefault,
		Type:           utils.MetaNone,
		RunDelay:       0,
		ConcurrentReqs: 0,
		SourcePath:     "/var/spool/cgrates/ers/in",
		ProcessedPath:  "/var/spool/cgrates/ers/out",
		Filters:        []string{},
		Opts:           make(map[string]interface{}),
	}
	result1, err := NewXMLFileER(cfg, 0, nil, nil, nil, nil, nil)
	if err != nil {
		t.Errorf("\nExpected: <%+v>, \nreceived: <%+v>", nil, err)
	}
	result2 := result1.Config()
	if !reflect.DeepEqual(result2, cfg.ERsCfg().Readers[0]) {
		t.Errorf("\nExpected: <%+v>, \nreceived: <%+v>", result2, cfg.ERsCfg().Readers[0])
	}
}

func TestERSXMLFileERServeNil(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers[0] = &config.EventReaderCfg{
		ID:             utils.MetaDefault,
		Type:           utils.MetaNone,
		RunDelay:       0,
		ConcurrentReqs: 0,
		SourcePath:     "/var/spool/cgrates/ers/in",
		ProcessedPath:  "/var/spool/cgrates/ers/out",
		Filters:        []string{},
		Opts:           make(map[string]interface{}),
	}
	result1, err := NewXMLFileER(cfg, 0, nil, nil, nil, nil, nil)
	if err != nil {
		t.Errorf("\nExpected: <%+v>, \nreceived: <%+v>", nil, err)
	}
	err = result1.Serve()
	if err != nil {
		t.Errorf("\nExpected: <%+v>, \nreceived: <%+v>", nil, err)
	}
}
