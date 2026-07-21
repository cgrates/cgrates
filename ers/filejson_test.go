// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ers

import (
	"reflect"
	"sync"
	"testing"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/config"
)

func TestNewJSONFileER(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	cfgIdx := 0
	expected := &JSONFileER{
		RWMutex:   sync.RWMutex{},
		cgrCfg:    cfg,
		cfgIdx:    0,
		cache:     cacheS,
		fltrS:     nil,
		rdrEvents: nil,
		rdrError:  nil,
		rdrExit:   nil,
		conReqs:   nil,
	}
	cfg.ERsCfg().Readers[0].ConcurrentReqs = 1
	cfg.ERsCfg().Readers[0].SourcePath = "/"
	result, err := NewJSONFileER(cfg, cfgIdx, nil, nil, nil, cacheS, nil, nil)
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
