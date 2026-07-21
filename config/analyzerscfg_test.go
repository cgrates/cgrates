// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestAnalyzerSCfgloadFromJsonCfg(t *testing.T) {
	jsonCfg := &AnalyzerSJsonCfg{
		Enabled: utils.BoolPointer(false),
		Opts: &AnalyzerSOptsJson{
			ExporterIDs: []*DynamicStringSliceOpt{
				{
					FilterIDs: []string{"filtr1"},
				},
			},
		},
	}
	expected := &AnalyzerSCfg{
		Enabled:         false,
		CleanupInterval: time.Hour,
		DBPath:          "/var/spool/cgrates/analyzers",
		IndexType:       utils.MetaScorch,
		Conns:           map[string][]*DynamicConns{},
		TTL:             24 * time.Hour,
		Opts: &AnalyzerSOpts{
			ExporterIDs: []*DynamicStringSliceOpt{
				{
					FilterIDs: []string{"filtr1"},
				},
			},
		},
	}
	jsnCfg := NewDefaultCGRConfig()
	if err := jsnCfg.analyzerSCfg.loadFromJSONCfg(jsonCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(jsnCfg.analyzerSCfg, expected) {
		t.Errorf("Expected %+v \n, received %+v", expected, jsnCfg.analyzerSCfg)
	}
	jsonCfg = nil
	if err := jsnCfg.analyzerSCfg.loadFromJSONCfg(jsonCfg); err != nil {
		t.Error(err)
	}
}

func TestAnalyzerSCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
		"analyzers":{},
    }
}`
	eMap := map[string]any{
		utils.EnabledCfg:         false,
		utils.CleanupIntervalCfg: "1h0m0s",
		utils.DBPathCfg:          "/var/spool/cgrates/analyzers",
		utils.IndexTypeCfg:       utils.MetaScorch,
		utils.ConnsCfg:           map[string][]*DynamicConns{},
		utils.TTLCfg:             "24h0m0s",
		utils.OptsCfg: map[string]any{
			utils.MetaExporterIDs: []*DynamicStringSliceOpt{},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.analyzerSCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected: %+v , received: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestAnalyzerSCfgAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
		"analyzers":{
            "enabled": true,
			"conns": {"*ees": [{"connIDs": ["*localhost"]}]},
        },
    }
}`
	eMap := map[string]any{
		utils.EnabledCfg:         true,
		utils.CleanupIntervalCfg: "1h0m0s",
		utils.DBPathCfg:          "/var/spool/cgrates/analyzers",
		utils.IndexTypeCfg:       utils.MetaScorch,
		utils.ConnsCfg:           map[string][]*DynamicConns{utils.MetaEEs: {{ConnIDs: []string{"*localhost"}}}},
		utils.TTLCfg:             "24h0m0s",
		utils.OptsCfg: map[string]any{
			utils.MetaExporterIDs: []*DynamicStringSliceOpt{},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.analyzerSCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected: %+v , received: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestAnalyzerSCfgloadFromJsonCfgErr(t *testing.T) {
	jsonCfg := &AnalyzerSJsonCfg{
		Cleanup_interval: utils.StringPointer("24ha"),
	}
	jsnCfg := NewDefaultCGRConfig()
	if err := jsnCfg.analyzerSCfg.loadFromJSONCfg(jsonCfg); err == nil {
		t.Errorf("Expected error received nil")
	}
	jsonCfg = &AnalyzerSJsonCfg{
		Ttl: utils.StringPointer("24ha"),
	}
	jsnCfg = NewDefaultCGRConfig()
	if err := jsnCfg.analyzerSCfg.loadFromJSONCfg(jsonCfg); err == nil {
		t.Errorf("Expected error received nil")
	}
}

func TestAnalyzerSCfgClone(t *testing.T) {
	cS := &AnalyzerSCfg{
		Enabled:         false,
		CleanupInterval: time.Hour,
		DBPath:          "/var/spool/cgrates/analyzers",
		IndexType:       utils.MetaScorch,
		Conns:           map[string][]*DynamicConns{utils.MetaEEs: {{ConnIDs: []string{"*internal"}}}},
		TTL:             24 * time.Hour,
	}
	rcv := cS.Clone()
	if !reflect.DeepEqual(cS, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cS), utils.ToJSON(rcv))
	}
	if rcv.DBPath = ""; cS.DBPath != "/var/spool/cgrates/analyzers" {
		t.Errorf("Expected clone to not modify the cloned")
	}
}

func TestDiffAnalyzerSJsonCfg(t *testing.T) {
	var d *AnalyzerSJsonCfg

	v1 := &AnalyzerSCfg{
		Enabled:         false,
		DBPath:          "",
		IndexType:       utils.MetaPrefix,
		TTL:             2 * time.Minute,
		CleanupInterval: time.Hour,
		Opts:            &AnalyzerSOpts{},
	}

	v2 := &AnalyzerSCfg{
		Enabled:         true,
		DBPath:          "/var/spool/cgrates/analyzers",
		IndexType:       utils.MetaString,
		TTL:             3 * time.Minute,
		Conns:           map[string][]*DynamicConns{utils.MetaEEs: {{ConnIDs: []string{"*internal"}}}},
		CleanupInterval: 30 * time.Minute,
		Opts:            &AnalyzerSOpts{},
	}

	expected := &AnalyzerSJsonCfg{
		Enabled:          utils.BoolPointer(true),
		Db_path:          utils.StringPointer("/var/spool/cgrates/analyzers"),
		Index_type:       utils.StringPointer(utils.MetaString),
		Ttl:              utils.StringPointer("3m0s"),
		Conns:            map[string][]*DynamicConns{utils.MetaEEs: {{ConnIDs: []string{"*internal"}}}},
		Cleanup_interval: utils.StringPointer("30m0s"),
		Opts:             &AnalyzerSOptsJson{},
	}

	rcv := diffAnalyzerSJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v2 = v1
	expected2 := &AnalyzerSJsonCfg{
		Opts: &AnalyzerSOptsJson{},
	}
	rcv = diffAnalyzerSJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected2) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected2), utils.ToJSON(rcv))
	}
}

func TestAnalyzerSCloneSection(t *testing.T) {
	anlCfg := &AnalyzerSCfg{
		Enabled:         true,
		DBPath:          "/var/spool/cgrates/analyzers",
		IndexType:       utils.MetaString,
		TTL:             3 * time.Minute,
		CleanupInterval: 30 * time.Minute,
	}

	exp := &AnalyzerSCfg{
		Enabled:         true,
		DBPath:          "/var/spool/cgrates/analyzers",
		IndexType:       utils.MetaString,
		TTL:             3 * time.Minute,
		CleanupInterval: 30 * time.Minute,
	}
	rcv := anlCfg.CloneSection()
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestToLoadFromJSONCfg(t *testing.T) {
	anzOpts := &AnalyzerSOpts{
		ExporterIDs: []*DynamicStringSliceOpt{
			{
				FilterIDs: []string{"filtr1"},
			},
		},
	}
	var jsonAnzOpts *AnalyzerSOptsJson

	temp := *anzOpts

	anzOpts.loadFromJSONCfg(jsonAnzOpts)
	if !reflect.DeepEqual(temp, *anzOpts) {
		t.Errorf("Expected anzOpts to not change, was<%v>, now is <%v>", temp, *anzOpts)
	}
}

func TestDiffAnalyzerSOptsJsonCfg(t *testing.T) {
	d := &AnalyzerSOptsJson{}
	v1 := &AnalyzerSOpts{
		ExporterIDs: []*DynamicStringSliceOpt{},
	}
	v2 := &AnalyzerSOpts{
		ExporterIDs: []*DynamicStringSliceOpt{{
			FilterIDs: []string{"test"},
		},
		},
	}
	exp := &AnalyzerSOptsJson{

		ExporterIDs: []*DynamicStringSliceOpt{{
			FilterIDs: []string{"test"},
		},
		},
	}
	if rcv := diffAnalyzerSOptsJsonCfg(d, v1, v2); utils.ToJSON(rcv) != utils.ToJSON(exp) {
		t.Errorf("Expected <%v> \n Received \n <%v>", exp, rcv)
	}
}
