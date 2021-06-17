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

var (
	cgrConfigDIR string

	cgrTests = []func(t *testing.T){
		testNewCgrJsonCfgFromHttp,
		testNewCGRConfigFromPath,
		testCGRConfigReloadAttributeS,
		testCGRConfigReloadChargerS,
		testCGRConfigReloadThresholdS,
		testCGRConfigReloadStatS,
		testCGRConfigReloadResourceS,
		testCGRConfigReloadSupplierS,
		testCGRConfigReloadSchedulerS,
		testCGRConfigReloadCDRs,
		testCGRConfigReloadRALs,
		testCGRConfigReloadSessionS,
		testCGRConfigReloadERs,
		testCGRConfigReloadDNSAgent,
		testCGRConfigReloadFreeswitchAgent,
		testCgrCfgV1ReloadConfigSection,
		testCGRConfigReloadConfigFromJSONSessionS,
		testCGRConfigReloadAll,
	}
)

func TestCGRConfig(t *testing.T) {
	for _, test := range cgrTests {
		t.Run("CGRConfig", test)
	}
}

func testNewCgrJsonCfgFromHttp(t *testing.T) {
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

	if _, err = net.DialTimeout(utils.TCP, addr, time.Second); err != nil { // check if site is up
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

func testNewCGRConfigFromPath(t *testing.T) {
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

	if _, err = net.DialTimeout(utils.TCP, addr, time.Second); err != nil { // check if site is up
		return
	}

	if rply, err := NewCGRConfigFromPath(addr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expVal, rply) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(expVal), utils.ToJSON(rply))
	}

}
func testCGRConfigReloadAttributeS(t *testing.T) {
	cfg, err := NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	var reply string
	if err = cfg.V1ReloadConfigFromPath(&ConfigReloadWithArgDispatcher{
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

func testCGRConfigReloadChargerS(t *testing.T) {
	cfg, err := NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	var reply string
	if err = cfg.V1ReloadConfigFromPath(&ConfigReloadWithArgDispatcher{
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
		AttributeSConns:     []string{"*localhost"},
	}
	if !reflect.DeepEqual(expAttr, cfg.ChargerSCfg()) {
		t.Errorf("Expected %s , received: %s ", utils.ToJSON(expAttr), utils.ToJSON(cfg.ChargerSCfg()))
	}
}

func testCGRConfigReloadThresholdS(t *testing.T) {
	cfg, err := NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	var reply string
	if err = cfg.V1ReloadConfigFromPath(&ConfigReloadWithArgDispatcher{
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

func testCGRConfigReloadStatS(t *testing.T) {
	cfg, err := NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	var reply string
	if err = cfg.V1ReloadConfigFromPath(&ConfigReloadWithArgDispatcher{
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
		ThresholdSConns:     []string{utils.MetaLocalHost},
	}
	if !reflect.DeepEqual(expAttr, cfg.StatSCfg()) {
		t.Errorf("Expected %s , received: %s ", utils.ToJSON(expAttr), utils.ToJSON(cfg.StatSCfg()))
	}
}

func testCGRConfigReloadResourceS(t *testing.T) {
	cfg, err := NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	var reply string
	if err = cfg.V1ReloadConfigFromPath(&ConfigReloadWithArgDispatcher{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo2"),
		Section: RESOURCES_JSON,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	expAttr := &ResourceSConfig{
		Enabled:             true,
		StringIndexedFields: &[]string{utils.Account},
		PrefixIndexedFields: &[]string{},
		IndexedSelects:      true,
		ThresholdSConns:     []string{utils.MetaLocalHost},
	}
	if !reflect.DeepEqual(expAttr, cfg.ResourceSCfg()) {
		t.Errorf("Expected %s , received: %s ", utils.ToJSON(expAttr), utils.ToJSON(cfg.ResourceSCfg()))
	}
}

func testCGRConfigReloadSupplierS(t *testing.T) {
	cfg, err := NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	var reply string
	if err = cfg.V1ReloadConfigFromPath(&ConfigReloadWithArgDispatcher{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo2"),
		Section: SupplierSJson,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	expAttr := &SupplierSCfg{
		Enabled:             true,
		StringIndexedFields: &[]string{"LCRProfile"},
		PrefixIndexedFields: &[]string{utils.Destination},
		ResourceSConns:      []string{},
		StatSConns:          []string{},
		AttributeSConns:     []string{},
		RALsConns:           []string{},
		IndexedSelects:      true,
		DefaultRatio:        1,
	}
	if !reflect.DeepEqual(expAttr, cfg.SupplierSCfg()) {
		t.Errorf("Expected %s , received: %s ", utils.ToJSON(expAttr), utils.ToJSON(cfg.SupplierSCfg()))
	}
}

func testCGRConfigReloadSchedulerS(t *testing.T) {
	cfg, err := NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	var reply string
	if err = cfg.V1ReloadConfigFromPath(&ConfigReloadWithArgDispatcher{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo2"),
		Section: SCHEDULER_JSN,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	expAttr := &SchedulerCfg{
		Enabled:   true,
		CDRsConns: []string{utils.MetaLocalHost},
		Filters:   []string{},
	}
	if !reflect.DeepEqual(expAttr, cfg.SchedulerCfg()) {
		t.Errorf("Expected %s , received: %s ", utils.ToJSON(expAttr), utils.ToJSON(cfg.SchedulerCfg()))
	}
}

func testCGRConfigReloadCDRs(t *testing.T) {
	cfg, err := NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	cfg.RalsCfg().Enabled = true
	var reply string
	if err = cfg.V1ReloadConfigFromPath(&ConfigReloadWithArgDispatcher{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo2"),
		Section: CDRS_JSN,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	expAttr := &CdrsCfg{
		Enabled: true,
		ExtraFields: utils.RSRFields{
			utils.NewRSRFieldMustCompile("PayPalAccount"),
			utils.NewRSRFieldMustCompile("LCRProfile"),
			utils.NewRSRFieldMustCompile("ResourceID"),
		},
		ChargerSConns:   []string{utils.MetaLocalHost},
		RaterConns:      []string{},
		AttributeSConns: []string{},
		ThresholdSConns: []string{},
		StatSConns:      []string{},
		SMCostRetries:   5,
		StoreCdrs:       true,
	}
	if !reflect.DeepEqual(expAttr, cfg.CdrsCfg()) {
		t.Errorf("Expected %s , received: %s ", utils.ToJSON(expAttr), utils.ToJSON(cfg.CdrsCfg()))
	}
}

func testCGRConfigReloadRALs(t *testing.T) {
	cfg, err := NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	blMap := cfg.RalsCfg().BalanceRatingSubject
	maxComp := cfg.RalsCfg().MaxComputedUsage
	var reply string
	if err = cfg.V1ReloadConfigFromPath(&ConfigReloadWithArgDispatcher{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo2"),
		Section: RALS_JSN,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	expAttr := &RalsCfg{
		Enabled:                 true,
		RpSubjectPrefixMatching: false,
		RemoveExpired:           true,
		MaxComputedUsage:        maxComp,
		BalanceRatingSubject:    blMap,
		ThresholdSConns:         []string{utils.MetaLocalHost},
		StatSConns:              []string{utils.MetaLocalHost},
		MaxIncrements:           1000000,
	}
	if !reflect.DeepEqual(expAttr, cfg.RalsCfg()) {
		t.Errorf("Expected %s , received: %s ", utils.ToJSON(expAttr), utils.ToJSON(cfg.RalsCfg()))
	}
}

func testCGRConfigReloadSessionS(t *testing.T) {
	cfg, err := NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	cfg.RalsCfg().Enabled = true
	cfg.ChargerSCfg().Enabled = true
	cfg.CdrsCfg().Enabled = true
	var reply string
	if err = cfg.V1ReloadConfigFromPath(&ConfigReloadWithArgDispatcher{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo2"),
		Section: SessionSJson,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	expAttr := &SessionSCfg{
		Enabled:       true,
		ListenBijson:  "127.0.0.1:2014",
		ChargerSConns: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)},
		RALsConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResponder)},
		ResSConns:     []string{utils.MetaLocalHost},
		ThreshSConns:  []string{},
		StatSConns:    []string{},
		SupplSConns:   []string{utils.MetaLocalHost},
		AttrSConns:    []string{utils.MetaLocalHost},
		CDRsConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs)},

		ReplicationConns:  []string{},
		SessionIndexes:    utils.NewStringMap(),
		ClientProtocol:    1,
		TerminateAttempts: 5,
		AlterableFields:   utils.NewStringSet([]string{}),
		DefaultUsage: map[string]time.Duration{
			utils.META_ANY: 3 * time.Hour,
			utils.VOICE:    3 * time.Hour,
			utils.DATA:     1048576,
			utils.SMS:      1,
		},
	}
	if !reflect.DeepEqual(expAttr, cfg.SessionSCfg()) {
		t.Errorf("Expected %s , received: %s ", utils.ToJSON(expAttr), utils.ToJSON(cfg.SessionSCfg()))
	}
}

func testCGRConfigReloadERs(t *testing.T) {
	for _, dir := range []string{"/tmp/ers/in", "/tmp/ers/out"} {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal("Error removing folder: ", dir, err)
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal("Error creating folder: ", dir, err)
		}
	}

	cfg, err := NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	cfg.SessionSCfg().Enabled = true
	var reply string
	if err = cfg.V1ReloadConfigFromPath(&ConfigReloadWithArgDispatcher{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "ers_example"),
		Section: ERsJson,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	flags, _ := utils.FlagsWithParamsFromSlice([]string{"*dryrun"})
	flagsDefault, _ := utils.FlagsWithParamsFromSlice([]string{})
	content := []*FCTemplate{
		{Tag: utils.ToR, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR, Type: utils.MetaVariable, Value: NewRSRParsersMustCompile("~*req.2", true, utils.INFIELD_SEP), Mandatory: true},
		{Tag: utils.OriginID, Path: utils.MetaCgreq + utils.NestingSep + utils.OriginID, Type: utils.MetaVariable, Value: NewRSRParsersMustCompile("~*req.3", true, utils.INFIELD_SEP), Mandatory: true},
		{Tag: utils.RequestType, Path: utils.MetaCgreq + utils.NestingSep + utils.RequestType, Type: utils.MetaVariable, Value: NewRSRParsersMustCompile("~*req.4", true, utils.INFIELD_SEP), Mandatory: true},
		{Tag: utils.Tenant, Path: utils.MetaCgreq + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable, Value: NewRSRParsersMustCompile("~*req.6", true, utils.INFIELD_SEP), Mandatory: true},
		{Tag: utils.Category, Path: utils.MetaCgreq + utils.NestingSep + utils.Category, Type: utils.MetaVariable, Value: NewRSRParsersMustCompile("~*req.7", true, utils.INFIELD_SEP), Mandatory: true},
		{Tag: utils.Account, Path: utils.MetaCgreq + utils.NestingSep + utils.Account, Type: utils.MetaVariable, Value: NewRSRParsersMustCompile("~*req.8", true, utils.INFIELD_SEP), Mandatory: true},
		{Tag: utils.Subject, Path: utils.MetaCgreq + utils.NestingSep + utils.Subject, Type: utils.MetaVariable, Value: NewRSRParsersMustCompile("~*req.9", true, utils.INFIELD_SEP), Mandatory: true},
		{Tag: utils.Destination, Path: utils.MetaCgreq + utils.NestingSep + utils.Destination, Type: utils.MetaVariable, Value: NewRSRParsersMustCompile("~*req.10", true, utils.INFIELD_SEP), Mandatory: true},
		{Tag: utils.SetupTime, Path: utils.MetaCgreq + utils.NestingSep + utils.SetupTime, Type: utils.MetaVariable, Value: NewRSRParsersMustCompile("~*req.11", true, utils.INFIELD_SEP), Mandatory: true},
		{Tag: utils.AnswerTime, Path: utils.MetaCgreq + utils.NestingSep + utils.AnswerTime, Type: utils.MetaVariable, Value: NewRSRParsersMustCompile("~*req.12", true, utils.INFIELD_SEP), Mandatory: true},
		{Tag: utils.Usage, Path: utils.MetaCgreq + utils.NestingSep + utils.Usage, Type: utils.MetaVariable, Value: NewRSRParsersMustCompile("~*req.13", true, utils.INFIELD_SEP), Mandatory: true},
	}
	for _, v := range content {
		v.ComputePath()
	}
	expAttr := &ERsCfg{
		Enabled:       true,
		SessionSConns: []string{utils.MetaLocalHost},
		Readers: []*EventReaderCfg{
			{
				ID:              utils.MetaDefault,
				Type:            utils.META_NONE,
				FieldSep:        ",",
				RunDelay:        0,
				ConcurrentReqs:  1024,
				SourcePath:      "/var/spool/cgrates/ers/in",
				ProcessedPath:   "/var/spool/cgrates/ers/out",
				Filters:         []string{},
				Flags:           flagsDefault,
				Fields:          content,
				CacheDumpFields: []*FCTemplate{},
				XmlRootPath:     utils.HierarchyPath{utils.EmptyString},
			},
			{
				ID:              "file_reader1",
				Type:            utils.MetaFileCSV,
				FieldSep:        ",",
				RunDelay:        -1,
				ConcurrentReqs:  1024,
				SourcePath:      "/tmp/ers/in",
				ProcessedPath:   "/tmp/ers/out",
				Flags:           flags,
				Fields:          content,
				CacheDumpFields: []*FCTemplate{},
				XmlRootPath:     utils.HierarchyPath{utils.EmptyString},
			},
		},
	}
	if !reflect.DeepEqual(expAttr, cfg.ERsCfg()) {
		t.Errorf("Expected %s , received: %s ", utils.ToJSON(expAttr), utils.ToJSON(cfg.ERsCfg()))
	}
}

func testCGRConfigReloadDNSAgent(t *testing.T) {
	cfg, err := NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	cfg.SessionSCfg().Enabled = true
	var reply string
	if err = cfg.V1ReloadConfigFromPath(&ConfigReloadWithArgDispatcher{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "dnsagent_reload"),
		Section: DNSAgentJson,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	expAttr := &DNSAgentCfg{
		Enabled:       true,
		Listen:        ":2053",
		ListenNet:     "udp",
		SessionSConns: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
		// Timezone          string
		// RequestProcessors []*RequestProcessor
	}
	if !reflect.DeepEqual(expAttr, cfg.DNSAgentCfg()) {
		t.Errorf("Expected %s , received: %s ", utils.ToJSON(expAttr), utils.ToJSON(cfg.DNSAgentCfg()))
	}
}

func testCGRConfigReloadFreeswitchAgent(t *testing.T) {
	cfg, err := NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	cfg.SessionSCfg().Enabled = true
	var reply string
	if err = cfg.V1ReloadConfigFromPath(&ConfigReloadWithArgDispatcher{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "freeswitch_reload"),
		Section: FreeSWITCHAgentJSN,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	expAttr := &FsAgentCfg{
		Enabled:           true,
		SessionSConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
		SubscribePark:     true,
		ExtraFields:       RSRParsers{},
		MaxWaitConnection: 2 * time.Second,
		EventSocketConns: []*FsConnCfg{
			&FsConnCfg{
				Address:    "1.2.3.4:8021",
				Password:   "ClueCon",
				Reconnects: 5,
				Alias:      "1.2.3.4:8021",
			},
		},
	}
	if !reflect.DeepEqual(expAttr, cfg.FsAgentCfg()) {
		t.Errorf("Expected %s , received: %s ", utils.ToJSON(expAttr), utils.ToJSON(cfg.FsAgentCfg()))
	}
}

func testCgrCfgV1ReloadConfigSection(t *testing.T) {
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
			"Path":             "*cgreq.ToR",
			"Filters":          nil,
			"Layout":           "",
			"Mandatory":        true,
			"MaskDestID":       "",
			"MaskLen":          0,
			"NewBranch":        false,
			"Padding":          "",
			"RoundingDecimals": 0,
			"Strip":            "",
			"Tag":              "ToR",
			"Timezone":         "",
			"Type":             "*variable",
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
			"Path":             "*cgreq.OriginID",
			"Filters":          nil,
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
			"Type":             "*variable",
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
			"Path":             "*cgreq.RequestType",
			"Filters":          nil,
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
			"Type":             "*variable",
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
			"Path":             "*cgreq.Tenant",
			"Filters":          nil,
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
			"Type":             "*variable",
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
			"Path":             "*cgreq.Category",
			"Filters":          nil,
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
			"Type":             "*variable",
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
			"Path":             "*cgreq.Account",
			"Filters":          nil,
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
			"Type":             "*variable",
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
			"Path":             "*cgreq.Subject",
			"Filters":          nil,
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
			"Type":             "*variable",
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
			"Path":             "*cgreq.Destination",
			"Filters":          nil,
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
			"Type":             "*variable",
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
			"Path":             "*cgreq.SetupTime",
			"Filters":          nil,
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
			"Type":             "*variable",
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
			"Path":             "*cgreq.AnswerTime",
			"Filters":          nil,
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
			"Type":             "*variable",
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
			"Path":             "*cgreq.Usage",
			"Filters":          nil,
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
			"Type":             "*variable",
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
				"PartialCacheExpiryAction": "",
				"PartialRecordCache":       0,
				"CacheDumpFields":          []interface{}{},
				"ConcurrentReqs":           1024,
				"Fields":                   content,
				"FieldSep":                 ",",
				"Filters":                  []interface{}{},
				"Flags":                    map[string]interface{}{},
				"FailedCallsPrefix":        "",
				"ID":                       "*default",
				"ProcessedPath":            "/var/spool/cgrates/ers/out",
				"RowLength":                0,
				"RunDelay":                 0,
				"SourcePath":               "/var/spool/cgrates/ers/in",
				"Tenant":                   nil,
				"Timezone":                 "",
				"Type":                     "*none",
				"XmlRootPath":              []interface{}{utils.EmptyString},
			},
			map[string]interface{}{
				"CacheDumpFields": []interface{}{},
				"ConcurrentReqs":  1024,
				"FieldSep":        ",",
				"Filters":         nil,
				"Flags": map[string]interface{}{
					"*dryrun": []interface{}{},
				},
				"FailedCallsPrefix":        "",
				"PartialCacheExpiryAction": "",
				"PartialRecordCache":       0,
				"ID":                       "file_reader1",
				"ProcessedPath":            "/tmp/ers/out",
				"RowLength":                0,
				"RunDelay":                 -1.,
				"SourcePath":               "/tmp/ers/in",
				"Tenant":                   nil,
				"Timezone":                 "",
				"Type":                     "*file_csv",
				"XmlRootPath":              []interface{}{utils.EmptyString},
				"Fields":                   content,
			},
		},
		"SessionSConns": []string{
			utils.MetaLocalHost,
		},
	}

	cfg, _ := NewDefaultCGRConfig()
	var reply string
	var rcv map[string]interface{}

	if err := cfg.V1ReloadConfigFromPath(&ConfigReloadWithArgDispatcher{
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

func testCGRConfigReloadConfigFromJSONSessionS(t *testing.T) {
	cfg, err := NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	cfg.RalsCfg().Enabled = true
	cfg.ChargerSCfg().Enabled = true
	cfg.CdrsCfg().Enabled = true
	var reply string
	if err = cfg.V1ReloadConfigFromJSON(&JSONReloadWithArgDispatcher{
		JSON: map[string]interface{}{
			"sessions": map[string]interface{}{
				"enabled":          true,
				"resources_conns":  []string{"*localhost"},
				"suppliers_conns":  []string{"*localhost"},
				"attributes_conns": []string{"*localhost"},
				"rals_conns":       []string{"*internal"},
				"cdrs_conns":       []string{"*internal"},
				"chargers_conns":   []string{"*internal"},
			},
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	expAttr := &SessionSCfg{
		Enabled:       true,
		ListenBijson:  "127.0.0.1:2014",
		ChargerSConns: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)},
		RALsConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResponder)},
		ResSConns:     []string{utils.MetaLocalHost},
		ThreshSConns:  []string{},
		StatSConns:    []string{},
		SupplSConns:   []string{utils.MetaLocalHost},
		AttrSConns:    []string{utils.MetaLocalHost},
		CDRsConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs)},

		ReplicationConns:  []string{},
		SessionIndexes:    utils.NewStringMap(),
		ClientProtocol:    1,
		TerminateAttempts: 5,
		AlterableFields:   utils.NewStringSet([]string{}),
		DefaultUsage: map[string]time.Duration{
			utils.META_ANY: 3 * time.Hour,
			utils.VOICE:    3 * time.Hour,
			utils.DATA:     1048576,
			utils.SMS:      1,
		},
	}
	if !reflect.DeepEqual(expAttr, cfg.SessionSCfg()) {
		t.Errorf("Expected %s , received: %s ", utils.ToJSON(expAttr), utils.ToJSON(cfg.SessionSCfg()))
	}
}

func testCGRConfigReloadAll(t *testing.T) {
	cfg, err := NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	cfg.RalsCfg().Enabled = true
	cfg.ChargerSCfg().Enabled = true
	cfg.CdrsCfg().Enabled = true
	var reply string
	if err = cfg.V1ReloadConfigFromPath(&ConfigReloadWithArgDispatcher{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo2"),
		Section: utils.MetaAll,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	expAttr := &SessionSCfg{
		Enabled:       true,
		ListenBijson:  "127.0.0.1:2014",
		ChargerSConns: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)},
		RALsConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResponder)},
		ResSConns:     []string{utils.MetaLocalHost},
		ThreshSConns:  []string{},
		StatSConns:    []string{},
		SupplSConns:   []string{utils.MetaLocalHost},
		AttrSConns:    []string{utils.MetaLocalHost},
		CDRsConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs)},

		ReplicationConns:  []string{},
		SessionIndexes:    utils.NewStringMap(),
		ClientProtocol:    1,
		TerminateAttempts: 5,
		AlterableFields:   utils.NewStringSet([]string{}), DefaultUsage: map[string]time.Duration{
			utils.META_ANY: 3 * time.Hour,
			utils.VOICE:    3 * time.Hour,
			utils.DATA:     1048576,
			utils.SMS:      1,
		},
	}
	if !reflect.DeepEqual(expAttr, cfg.SessionSCfg()) {
		t.Errorf("Expected %s , received: %s ", utils.ToJSON(expAttr), utils.ToJSON(cfg.SessionSCfg()))
	}
}
