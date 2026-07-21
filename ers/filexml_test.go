// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ers

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/config"
)

func TestERSNewXMLFileER(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	expected := &XMLFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		cache:     cacheS,
		fltrS:     nil,
		sourceDir: "/var/spool/cgrates/ers/in",
		rdrEvents: nil,
		rdrError:  nil,
		rdrExit:   nil,
		conReqs:   nil,
	}
	result, err := NewXMLFileER(cfg, 0, nil, nil, nil, cacheS, nil, nil)
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
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	cfg.ERsCfg().Readers[0] = &config.EventReaderCfg{
		ID:             utils.MetaDefault,
		Type:           utils.MetaNone,
		RunDelay:       0,
		ConcurrentReqs: 0,
		SourcePath:     "/var/spool/cgrates/ers/in",
		ProcessedPath:  "/var/spool/cgrates/ers/out",
		Filters:        []string{},
		Opts:           &config.EventReaderOpts{},
	}
	result1, err := NewXMLFileER(cfg, 0, nil, nil, nil, cacheS, nil, nil)
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
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	cfg.ERsCfg().Readers[0] = &config.EventReaderCfg{
		ID:             utils.MetaDefault,
		Type:           utils.MetaNone,
		RunDelay:       0,
		ConcurrentReqs: 0,
		SourcePath:     "/var/spool/cgrates/ers/in",
		ProcessedPath:  "/var/spool/cgrates/ers/out",
		Filters:        []string{},
		Opts:           &config.EventReaderOpts{},
	}
	result1, err := NewXMLFileER(cfg, 0, nil, nil, nil, cacheS, nil, nil)
	if err != nil {
		t.Errorf("\nExpected: <%+v>, \nreceived: <%+v>", nil, err)
	}
	err = result1.Serve()
	if err != nil {
		t.Errorf("\nExpected: <%+v>, \nreceived: <%+v>", nil, err)
	}
}
