// +build integration

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
package config

import (
	"net"
	"os"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestNewCgrJsonCfgFromHttp(t *testing.T) {
	addr := "https://raw.githubusercontent.com/cgrates/cgrates/master/data/conf/samples/tutmongo/cgrates.json"
	expVal, err := NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	err = expVal.loadConfigFromPath(path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo"),
		[]func(*CgrJsonCfg) error{expVal.loadFromJsonCfg})
	if err != nil {
		t.Fatal(err)
	}

	if _, err = net.DialTimeout("tcp", addr, time.Second); err != nil { // check if site is up
		return
	}

	rply, err := NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	if err = rply.loadConfigFromPath(addr, []func(*CgrJsonCfg) error{rply.loadFromJsonCfg}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expVal, rply) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(expVal), utils.ToJSON(rply))
	}

}

func TestNewCGRConfigFromPath(t *testing.T) {
	for key, val := range map[string]string{"LOGGER": "*syslog", "LOG_LEVEL": "6", "TLS_VERIFY": "false", "ROUND_DEC": "5",
		"DB_ENCODING": "*msgpack", "TP_EXPORT_DIR": "/var/spool/cgrates/tpe", "FAILED_POSTS_DIR": "/var/spool/cgrates/failed_posts",
		"DF_TENANT": "cgrates.org", "TIMEZONE": "Local"} {
		os.Setenv(key, val)
	}
	addr := "https://raw.githubusercontent.com/cgrates/cgrates/master/data/conf/samples/multifiles/a.json;https://raw.githubusercontent.com/cgrates/cgrates/master/data/conf/samples/multifiles/b/b.json;https://raw.githubusercontent.com/cgrates/cgrates/master/data/conf/samples/multifiles/c.json;https://raw.githubusercontent.com/cgrates/cgrates/master/data/conf/samples/multifiles/d.json"
	expVal, err := NewCGRConfigFromPath(path.Join("/usr", "share", "cgrates", "conf", "samples", "multifiles"))
	if err != nil {
		t.Fatal(err)
	}

	if _, err = net.DialTimeout("tcp", addr, time.Second); err != nil { // check if site is up
		return
	}

	if rply, err := NewCGRConfigFromPath(addr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expVal, rply) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(expVal), utils.ToJSON(rply))
	}

}
func TestCGRConfigReloadAttributeS(t *testing.T) {
	cfg, err := NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	var reply string
	if err = cfg.V1ReloadConfig(&ConfigReloadWithArgDispatcher{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo2"),
		Section: ATTRIBUTE_JSN,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	expAttr := &AttributeSCfg{
		Enabled:             true,
		StringIndexedFields: &[]string{utils.Account},
		PrefixIndexedFields: &[]string{},
		IndexedSelects:      true,
		ProcessRuns:         1,
	}
	if !reflect.DeepEqual(expAttr, cfg.AttributeSCfg()) {
		t.Errorf("Expected %s , received: %s ", utils.ToJSON(expAttr), utils.ToJSON(cfg.AttributeSCfg()))
	}
}

func TestCGRConfigReloadChargerS(t *testing.T) {
	cfg, err := NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	var reply string
	if err = cfg.V1ReloadConfig(&ConfigReloadWithArgDispatcher{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo2"),
		Section: ChargerSCfgJson,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	expAttr := &ChargerSCfg{
		Enabled:             true,
		StringIndexedFields: &[]string{utils.Account},
		PrefixIndexedFields: &[]string{},
		IndexedSelects:      true,
		AttributeSConns: []*RemoteHost{
			&RemoteHost{
				Address:   "127.0.0.1:2012",
				Transport: utils.MetaJSONrpc,
			},
		},
	}
	if !reflect.DeepEqual(expAttr, cfg.ChargerSCfg()) {
		t.Errorf("Expected %s , received: %s ", utils.ToJSON(expAttr), utils.ToJSON(cfg.ChargerSCfg()))
	}
}

func TestCGRConfigReloadThresholdS(t *testing.T) {
	cfg, err := NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	var reply string
	if err = cfg.V1ReloadConfig(&ConfigReloadWithArgDispatcher{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo2"),
		Section: THRESHOLDS_JSON,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	expAttr := &ThresholdSCfg{
		Enabled:             true,
		StringIndexedFields: &[]string{utils.Account},
		PrefixIndexedFields: &[]string{},
		IndexedSelects:      true,
	}
	if !reflect.DeepEqual(expAttr, cfg.ThresholdSCfg()) {
		t.Errorf("Expected %s , received: %s ", utils.ToJSON(expAttr), utils.ToJSON(cfg.ThresholdSCfg()))
	}
}

func TestCGRConfigReloadStatS(t *testing.T) {
	cfg, err := NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	var reply string
	if err = cfg.V1ReloadConfig(&ConfigReloadWithArgDispatcher{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo2"),
		Section: STATS_JSON,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	expAttr := &StatSCfg{
		Enabled:             true,
		StringIndexedFields: &[]string{utils.Account},
		PrefixIndexedFields: &[]string{},
		IndexedSelects:      true,
		ThresholdSConns: []*RemoteHost{
			&RemoteHost{Address: "127.0.0.1:2012", Transport: utils.MetaJSONrpc},
		},
	}
	if !reflect.DeepEqual(expAttr, cfg.StatSCfg()) {
		t.Errorf("Expected %s , received: %s ", utils.ToJSON(expAttr), utils.ToJSON(cfg.StatSCfg()))
	}
}

func TestCgrCfgV1ReloadConfigSection(t *testing.T) {
	for _, dir := range []string{"/tmp/ers/in", "/tmp/ers/out"} {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal("Error removing folder: ", dir, err)
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal("Error creating folder: ", dir, err)
		}
	}
	content := []interface{}{
		map[string]interface{}{
			"AttributeID":      "",
			"Blocker":          false,
			"BreakOnSuccess":   false,
			"CostShiftDigits":  0,
			"FieldId":          "ToR",
			"Filters":          nil,
			"HandlerId":        "",
			"Layout":           "",
			"Mandatory":        true,
			"MaskDestID":       "",
			"MaskLen":          0,
			"NewBranch":        false,
			"Padding":          "",
			"RoundingDecimals": 0,
			"Strip":            "",
			"Tag":              "TOR",
			"Timezone":         "",
			"Type":             "*composed",
			"Value": []interface{}{
				map[string]interface{}{
					"AllFiltersMatch": true,
					"Rules":           "~*req.2",
				}},
			"Width": 0,
		},
		map[string]interface{}{
			"AttributeID":      "",
			"Blocker":          false,
			"BreakOnSuccess":   false,
			"CostShiftDigits":  0,
			"FieldId":          "OriginID",
			"Filters":          nil,
			"HandlerId":        "",
			"Layout":           "",
			"Mandatory":        true,
			"MaskDestID":       "",
			"MaskLen":          0,
			"NewBranch":        false,
			"Padding":          "",
			"RoundingDecimals": 0,
			"Strip":            "",
			"Tag":              "OriginID",
			"Timezone":         "",
			"Type":             "*composed",
			"Value": []interface{}{
				map[string]interface{}{
					"AllFiltersMatch": true,
					"Rules":           "~*req.3",
				}},
			"Width": 0,
		},
		map[string]interface{}{
			"AttributeID":      "",
			"Blocker":          false,
			"BreakOnSuccess":   false,
			"CostShiftDigits":  0,
			"FieldId":          "RequestType",
			"Filters":          nil,
			"HandlerId":        "",
			"Layout":           "",
			"Mandatory":        true,
			"MaskDestID":       "",
			"MaskLen":          0,
			"NewBranch":        false,
			"Padding":          "",
			"RoundingDecimals": 0,
			"Strip":            "",
			"Tag":              "RequestType",
			"Timezone":         "",
			"Type":             "*composed",
			"Value": []interface{}{
				map[string]interface{}{
					"AllFiltersMatch": true,
					"Rules":           "~*req.4",
				}},
			"Width": 0,
		},
		map[string]interface{}{
			"AttributeID":      "",
			"Blocker":          false,
			"BreakOnSuccess":   false,
			"CostShiftDigits":  0,
			"FieldId":          "Tenant",
			"Filters":          nil,
			"HandlerId":        "",
			"Layout":           "",
			"Mandatory":        true,
			"MaskDestID":       "",
			"MaskLen":          0,
			"NewBranch":        false,
			"Padding":          "",
			"RoundingDecimals": 0,
			"Strip":            "",
			"Tag":              "Tenant",
			"Timezone":         "",
			"Type":             "*composed",
			"Value": []interface{}{
				map[string]interface{}{
					"AllFiltersMatch": true,
					"Rules":           "~*req.6",
				}},
			"Width": 0,
		},
		map[string]interface{}{
			"AttributeID":      "",
			"Blocker":          false,
			"BreakOnSuccess":   false,
			"CostShiftDigits":  0,
			"FieldId":          "Category",
			"Filters":          nil,
			"HandlerId":        "",
			"Layout":           "",
			"Mandatory":        true,
			"MaskDestID":       "",
			"MaskLen":          0,
			"NewBranch":        false,
			"Padding":          "",
			"RoundingDecimals": 0,
			"Strip":            "",
			"Tag":              "Category",
			"Timezone":         "",
			"Type":             "*composed",
			"Value": []interface{}{
				map[string]interface{}{
					"AllFiltersMatch": true,
					"Rules":           "~*req.7",
				}},
			"Width": 0,
		},
		map[string]interface{}{
			"AttributeID":      "",
			"Blocker":          false,
			"BreakOnSuccess":   false,
			"CostShiftDigits":  0,
			"FieldId":          "Account",
			"Filters":          nil,
			"HandlerId":        "",
			"Layout":           "",
			"Mandatory":        true,
			"MaskDestID":       "",
			"MaskLen":          0,
			"NewBranch":        false,
			"Padding":          "",
			"RoundingDecimals": 0,
			"Strip":            "",
			"Tag":              "Account",
			"Timezone":         "",
			"Type":             "*composed",
			"Value": []interface{}{
				map[string]interface{}{
					"AllFiltersMatch": true,
					"Rules":           "~*req.8",
				}},
			"Width": 0,
		},
		map[string]interface{}{
			"AttributeID":      "",
			"Blocker":          false,
			"BreakOnSuccess":   false,
			"CostShiftDigits":  0,
			"FieldId":          "Subject",
			"Filters":          nil,
			"HandlerId":        "",
			"Layout":           "",
			"Mandatory":        true,
			"MaskDestID":       "",
			"MaskLen":          0,
			"NewBranch":        false,
			"Padding":          "",
			"RoundingDecimals": 0,
			"Strip":            "",
			"Tag":              "Subject",
			"Timezone":         "",
			"Type":             "*composed",
			"Value": []interface{}{
				map[string]interface{}{
					"AllFiltersMatch": true,
					"Rules":           "~*req.9",
				}},
			"Width": 0,
		},
		map[string]interface{}{
			"AttributeID":      "",
			"Blocker":          false,
			"BreakOnSuccess":   false,
			"CostShiftDigits":  0,
			"FieldId":          "Destination",
			"Filters":          nil,
			"HandlerId":        "",
			"Layout":           "",
			"Mandatory":        true,
			"MaskDestID":       "",
			"MaskLen":          0,
			"NewBranch":        false,
			"Padding":          "",
			"RoundingDecimals": 0,
			"Strip":            "",
			"Tag":              "Destination",
			"Timezone":         "",
			"Type":             "*composed",
			"Value": []interface{}{
				map[string]interface{}{
					"AllFiltersMatch": true,
					"Rules":           "~*req.10",
				}},
			"Width": 0,
		},
		map[string]interface{}{
			"AttributeID":      "",
			"Blocker":          false,
			"BreakOnSuccess":   false,
			"CostShiftDigits":  0,
			"FieldId":          "SetupTime",
			"Filters":          nil,
			"HandlerId":        "",
			"Layout":           "",
			"Mandatory":        true,
			"MaskDestID":       "",
			"MaskLen":          0,
			"NewBranch":        false,
			"Padding":          "",
			"RoundingDecimals": 0,
			"Strip":            "",
			"Tag":              "SetupTime",
			"Timezone":         "",
			"Type":             "*composed",
			"Value": []interface{}{
				map[string]interface{}{
					"AllFiltersMatch": true,
					"Rules":           "~*req.11",
				}},
			"Width": 0,
		},
		map[string]interface{}{
			"AttributeID":      "",
			"Blocker":          false,
			"BreakOnSuccess":   false,
			"CostShiftDigits":  0,
			"FieldId":          "AnswerTime",
			"Filters":          nil,
			"HandlerId":        "",
			"Layout":           "",
			"Mandatory":        true,
			"MaskDestID":       "",
			"MaskLen":          0,
			"NewBranch":        false,
			"Padding":          "",
			"RoundingDecimals": 0,
			"Strip":            "",
			"Tag":              "AnswerTime",
			"Timezone":         "",
			"Type":             "*composed",
			"Value": []interface{}{
				map[string]interface{}{
					"AllFiltersMatch": true,
					"Rules":           "~*req.12",
				}},
			"Width": 0,
		},
		map[string]interface{}{
			"AttributeID":      "",
			"Blocker":          false,
			"BreakOnSuccess":   false,
			"CostShiftDigits":  0,
			"FieldId":          "Usage",
			"Filters":          nil,
			"HandlerId":        "",
			"Layout":           "",
			"Mandatory":        true,
			"MaskDestID":       "",
			"MaskLen":          0,
			"NewBranch":        false,
			"Padding":          "",
			"RoundingDecimals": 0,
			"Strip":            "",
			"Tag":              "Usage",
			"Timezone":         "",
			"Type":             "*composed",
			"Value": []interface{}{
				map[string]interface{}{
					"AllFiltersMatch": true,
					"Rules":           "~*req.13",
				}},
			"Width": 0,
		},
	}
	expected := map[string]interface{}{
		"Enabled": true,
		"Readers": []interface{}{
			map[string]interface{}{
				"ConcurrentReqs": 1024,
				"ContentFields":  content,
				"FieldSep":       ",",
				"Filters":        []interface{}{},
				"Flags":          map[string]interface{}{},
				"HeaderFields":   []interface{}{},
				"ID":             "*default",
				"ProcessedPath":  "/var/spool/cgrates/cdrc/out",
				"RunDelay":       0,
				"SourcePath":     "/var/spool/cgrates/cdrc/in",
				"Tenant":         nil,
				"Timezone":       "",
				"TrailerFields":  []interface{}{},
				"Type":           "*file_csv",
				"XmlRootPath":    "",
			},
			map[string]interface{}{
				"ConcurrentReqs": 1024,
				"FieldSep":       ",",
				"Filters":        nil,
				"Flags": map[string]interface{}{
					"*dryrun": []interface{}{},
				},
				"HeaderFields":  []interface{}{},
				"ID":            "file_reader1",
				"ProcessedPath": "/tmp/ers/out",
				"RunDelay":      -1.,
				"SourcePath":    "/tmp/ers/in",
				"Tenant":        nil,
				"Timezone":      "",
				"TrailerFields": []interface{}{},
				"Type":          "*file_csv",
				"XmlRootPath":   "",
				"ContentFields": content,
			},
		},
		"SessionSConns": []interface{}{
			map[string]interface{}{
				"Address":     "127.0.0.1:2012",
				"Synchronous": false,
				"TLS":         false,
				"Transport":   "*json",
			},
		},
	}

	cfg, _ := NewDefaultCGRConfig()
	var reply string
	var rcv map[string]interface{}

	if err := cfg.V1ReloadConfig(&ConfigReloadWithArgDispatcher{
		Path:    "/usr/share/cgrates/conf/samples/ers_example",
		Section: ERsJson,
	}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("Expected: %s \n,received: %s", utils.OK, reply)
	}

	if err := cfg.V1GetConfigSection(&StringWithArgDispatcher{Section: ERsJson}, &rcv); err != nil {
		t.Error(err)
	} else if utils.ToJSON(expected) != utils.ToJSON(rcv) {
		t.Errorf("Expected: %+v, \n received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	for _, dir := range []string{"/tmp/ers/in", "/tmp/ers/out"} {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal("Error removing folder: ", dir, err)
		}
	}
}
