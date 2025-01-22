//go:build integration
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
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

var (
	cgrTests = []func(t *testing.T){
		testNewCgrJsonCfgFromHttp,
		testNewCGRConfigFromPath,
		testCGRConfigReloadAttributeS,
		testCGRConfigReloadAttributeSWithDB,
		testCGRConfigReloadChargerSDryRun,
		testCGRConfigReloadChargerS,
		testCGRConfigReloadThresholdS,
		testCGRConfigReloadStatS,
		testCGRConfigReloadResourceS,
		testCGRConfigReloadSupplierS,
		testCGRConfigReloadERs,
		testCGRConfigReloadDNSAgent,
		testCGRConfigReloadFreeswitchAgent,
		testCgrCfgV1ReloadConfigSection,
		testCGRConfigV1ReloadConfigFromPathInvalidSection,
		testV1ReloadConfigFromPathConfigSanity,
		testLoadConfigFromHTTPValidURL,
		testCGRConfigReloadConfigFromJSONSessionS,
		testCGRConfigReloadConfigFromStringSessionS,
		testCGRConfigReloadAll,
		testHttpHandlerConfigSForNotExistFile,
		testHttpHandlerConfigSForFile,
		testHttpHandlerConfigSForNotExistFolder,
		testHttpHandlerConfigSForFolder,
		testHttpHandlerConfigSInvalidPath,
		testLoadConfigFromPathInvalidArgument,
		testLoadConfigFromPathValidPath,
		testLoadConfigFromPathFile,
		testLoadConfigFromFolderFileNotFound,
		testLoadConfigFromFolderNoConfigFound,
		testLoadConfigFromFolderOpenError,
		testApisLoadFromPath,
		testHandleConfigSFolderError,
		testHandleConfigSFilesError,
		testHandleConfigSFileErrorWrite,
		testHandleConfigSFolderErrorWrite,
	}
)

func TestCGRConfig(t *testing.T) {
	for _, test := range cgrTests {
		t.Run("CGRConfig", test)
	}
}

func testNewCgrJsonCfgFromHttp(t *testing.T) {
	addr := "https://raw.githubusercontent.com/cgrates/cgrates/master/data/conf/samples/tutmongo/cgrates.json"
	expVal := NewDefaultCGRConfig()

	err := loadConfigFromPath(context.Background(), path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo"),
		expVal.sections, false, expVal)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := net.DialTimeout(utils.TCP, addr, time.Second); err != nil { // check if site is up
		return
	}

	rply := NewDefaultCGRConfig()
	if err := loadConfigFromPath(context.Background(), addr, rply.sections, false, rply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expVal, rply) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(expVal), utils.ToJSON(rply))
	}

}

func testNewCGRConfigFromPath(t *testing.T) {
	for key, val := range map[string]string{"LOGGER": "*syslog", "LOG_LEVEL": "6", "ROUND_DEC": "5",
		"DB_ENCODING": "*msgpack", "TP_EXPORT_DIR": "/var/spool/cgrates/tpe", "FAILED_POSTS_DIR": "/var/spool/cgrates/failed_posts",
		"DF_TENANT": "cgrates.org", "TIMEZONE": "Local"} {
		if err := os.Setenv(key, val); err != nil {
			t.Error(err)
		}
	}
	addr := "https://raw.githubusercontent.com/cgrates/cgrates/master/data/conf/samples/multifiles/a.json;https://raw.githubusercontent.com/cgrates/cgrates/master/data/conf/samples/multifiles/b/b.json;https://raw.githubusercontent.com/cgrates/cgrates/master/data/conf/samples/multifiles/c.json;https://raw.githubusercontent.com/cgrates/cgrates/master/data/conf/samples/multifiles/d.json"
	expVal, err := NewCGRConfigFromPath(context.Background(), path.Join("/usr", "share", "cgrates", "conf", "samples", "multifiles"))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := net.DialTimeout(utils.TCP, addr, time.Second); err != nil { // check if site is up
		return
	}

	if rply, err := NewCGRConfigFromPath(context.Background(), addr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expVal, rply) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(expVal), utils.ToJSON(rply))
	}

}
func testCGRConfigReloadAttributeS(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	cfg.rldCh = make(chan string, 100)
	var reply string
	cfg.ConfigPath = path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo2")
	if err := cfg.V1ReloadConfig(context.Background(), &ReloadArgs{
		Section: AttributeSJSON,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	expAttr := &AttributeSCfg{
		Enabled:                true,
		AccountSConns:          []string{},
		ResourceSConns:         []string{},
		StatSConns:             []string{},
		StringIndexedFields:    &[]string{utils.MetaReq + utils.NestingSep + utils.AccountField},
		PrefixIndexedFields:    &[]string{},
		SuffixIndexedFields:    &[]string{},
		ExistsIndexedFields:    &[]string{},
		NotExistsIndexedFields: &[]string{},
		IndexedSelects:         true,
		Opts: &AttributesOpts{
			ProfileIDs:           []*DynamicStringSliceOpt{},
			ProcessRuns:          []*DynamicIntOpt{{value: AttributesProcessRunsDftOpt}},
			ProfileRuns:          []*DynamicIntOpt{{value: AttributesProfileRunsDftOpt}},
			ProfileIgnoreFilters: []*DynamicBoolOpt{{value: AttributesProfileIgnoreFiltersDftOpt}},
		}}
	if !reflect.DeepEqual(expAttr, cfg.AttributeSCfg()) {
		t.Errorf("Expected %s , received: %s ", utils.ToJSON(expAttr), utils.ToJSON(cfg.AttributeSCfg()))
	}
}

func testCGRConfigReloadAttributeSWithDB(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	cfg.rldCh = make(chan string, 100)
	cfg.db = make(CgrJsonCfg)
	if err := cfg.db.SetSection(context.Background(), AttributeSJSON, &AttributeSJsonCfg{
		Stats_conns: &[]string{utils.MetaLocalHost},
	}); err != nil {
		t.Fatal(err)
	}
	var reply string
	cfg.ConfigPath = path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo2")
	if err := cfg.V1ReloadConfig(context.Background(), &ReloadArgs{
		Section: AttributeSJSON,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	expAttr := &AttributeSCfg{
		Enabled:                true,
		ResourceSConns:         []string{},
		AccountSConns:          []string{},
		StatSConns:             []string{utils.MetaLocalHost},
		StringIndexedFields:    &[]string{utils.MetaReq + utils.NestingSep + utils.AccountField},
		PrefixIndexedFields:    &[]string{},
		SuffixIndexedFields:    &[]string{},
		ExistsIndexedFields:    &[]string{},
		NotExistsIndexedFields: &[]string{},
		IndexedSelects:         true,
		Opts: &AttributesOpts{
			ProfileIDs:           []*DynamicStringSliceOpt{},
			ProcessRuns:          []*DynamicIntOpt{{value: AttributesProcessRunsDftOpt}},
			ProfileRuns:          []*DynamicIntOpt{{value: AttributesProfileRunsDftOpt}},
			ProfileIgnoreFilters: []*DynamicBoolOpt{{value: AccountsProfileIgnoreFiltersDftOpt}},
		},
	}
	if !reflect.DeepEqual(expAttr, cfg.AttributeSCfg()) {
		t.Errorf("Expected %s , received: %s ", utils.ToJSON(expAttr), utils.ToJSON(cfg.AttributeSCfg()))
	}
}

func testCGRConfigReloadChargerSDryRun(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	cfg.rldCh = make(chan string, 100)
	var reply string
	cfg.ConfigPath = path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo2")
	if err := cfg.V1ReloadConfig(context.Background(), &ReloadArgs{
		Section: ChargerSJSON,
		DryRun:  true,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	ecfg := NewDefaultCGRConfig()

	if !reflect.DeepEqual(ecfg.ChargerSCfg(), cfg.ChargerSCfg()) {
		t.Errorf("Expected %s , received: %s ", utils.ToJSON(ecfg.ChargerSCfg()), utils.ToJSON(cfg.ChargerSCfg()))
	}
}

func testCGRConfigReloadChargerS(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	cfg.rldCh = make(chan string, 100)
	var reply string
	cfg.ConfigPath = path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo2")
	if err := cfg.V1ReloadConfig(context.Background(), &ReloadArgs{
		Section: ChargerSJSON,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	expAttr := &ChargerSCfg{
		Enabled:                true,
		StringIndexedFields:    &[]string{utils.MetaReq + utils.NestingSep + utils.AccountField},
		PrefixIndexedFields:    &[]string{},
		SuffixIndexedFields:    &[]string{},
		ExistsIndexedFields:    &[]string{},
		NotExistsIndexedFields: &[]string{},
		IndexedSelects:         true,
		AttributeSConns:        []string{"*localhost"},
	}
	if !reflect.DeepEqual(expAttr, cfg.ChargerSCfg()) {
		t.Errorf("Expected %s , received: %s ", utils.ToJSON(expAttr), utils.ToJSON(cfg.ChargerSCfg()))
	}
}

func testCGRConfigReloadThresholdS(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	cfg.rldCh = make(chan string, 100)
	var reply string
	cfg.ConfigPath = path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo2")
	if err := cfg.V1ReloadConfig(context.Background(), &ReloadArgs{Section: ThresholdSJSON}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	expAttr := &ThresholdSCfg{
		Enabled:                true,
		StringIndexedFields:    &[]string{utils.MetaReq + utils.NestingSep + utils.AccountField},
		PrefixIndexedFields:    &[]string{},
		SuffixIndexedFields:    &[]string{},
		ExistsIndexedFields:    &[]string{},
		NotExistsIndexedFields: &[]string{},
		ActionSConns:           []string{},
		IndexedSelects:         true,
		Opts: &ThresholdsOpts{
			ProfileIDs:           []*DynamicStringSliceOpt{},
			ProfileIgnoreFilters: []*DynamicBoolOpt{{value: ThresholdsProfileIgnoreFiltersDftOpt}},
		}}
	if !reflect.DeepEqual(expAttr, cfg.ThresholdSCfg()) {
		t.Errorf("Expected %s , received: %s ", utils.ToJSON(expAttr), utils.ToJSON(cfg.ThresholdSCfg()))
	}
}

func testCGRConfigReloadStatS(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	cfg.rldCh = make(chan string, 100)
	var reply string
	cfg.ConfigPath = path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo2")
	if err := cfg.V1ReloadConfig(context.Background(), &ReloadArgs{
		Section: StatSJSON,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	expAttr := &StatSCfg{
		Enabled:                true,
		StringIndexedFields:    &[]string{utils.MetaReq + utils.NestingSep + utils.AccountField},
		PrefixIndexedFields:    &[]string{},
		SuffixIndexedFields:    &[]string{},
		ExistsIndexedFields:    &[]string{},
		NotExistsIndexedFields: &[]string{},
		IndexedSelects:         true,
		ThresholdSConns:        []string{utils.MetaLocalHost},
		Opts: &StatsOpts{
			ProfileIDs:           []*DynamicStringSliceOpt{},
			ProfileIgnoreFilters: []*DynamicBoolOpt{{value: StatsProfileIgnoreFilters}},
			RoundingDecimals:     []*DynamicIntOpt{},
			PrometheusStatIDs:    []*DynamicStringSliceOpt{},
		},
		EEsConns: []string{},
	}
	if !reflect.DeepEqual(expAttr, cfg.StatSCfg()) {
		t.Errorf("Expected %s , received: %s ", utils.ToJSON(expAttr), utils.ToJSON(cfg.StatSCfg()))
	}
}

func testCGRConfigReloadResourceS(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	cfg.rldCh = make(chan string, 100)
	var reply string
	cfg.ConfigPath = path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo2")
	if err := cfg.V1ReloadConfig(context.Background(), &ReloadArgs{
		Section: ResourceSJSON,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	expAttr := &ResourceSConfig{
		Enabled:                true,
		StringIndexedFields:    &[]string{utils.MetaReq + utils.NestingSep + utils.AccountField},
		PrefixIndexedFields:    &[]string{},
		SuffixIndexedFields:    &[]string{},
		ExistsIndexedFields:    &[]string{},
		NotExistsIndexedFields: &[]string{},
		IndexedSelects:         true,
		ThresholdSConns:        []string{utils.MetaLocalHost},
		Opts: &ResourcesOpts{
			UsageID:  []*DynamicStringOpt{{value: ResourcesUsageIDDftOpt}},
			UsageTTL: []*DynamicDurationOpt{{value: ResourcesUsageTTLDftOpt}},
			Units:    []*DynamicFloat64Opt{{value: ResourcesUnitsDftOpt}},
		},
	}
	if !reflect.DeepEqual(expAttr, cfg.ResourceSCfg()) {
		t.Errorf("Expected %s , received: %s ", utils.ToJSON(expAttr), utils.ToJSON(cfg.ResourceSCfg()))
	}
}

func testCGRConfigReloadSupplierS(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	cfg.rldCh = make(chan string, 100)
	var reply string
	cfg.ConfigPath = path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo2")
	if err := cfg.V1ReloadConfig(context.Background(), &ReloadArgs{
		Section: RouteSJSON,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	expAttr := &RouteSCfg{
		Enabled:                true,
		StringIndexedFields:    &[]string{"*req.LCRProfile"},
		PrefixIndexedFields:    &[]string{utils.MetaReq + utils.NestingSep + utils.Destination},
		SuffixIndexedFields:    &[]string{},
		ExistsIndexedFields:    &[]string{},
		NotExistsIndexedFields: &[]string{},
		ResourceSConns:         []string{},
		StatSConns:             []string{},
		AttributeSConns:        []string{},
		RateSConns:             []string{},
		AccountSConns:          []string{},
		IndexedSelects:         true,
		DefaultRatio:           1,
		Opts: &RoutesOpts{
			Context:      []*DynamicStringOpt{{value: RoutesContextDftOpt}},
			ProfileCount: []*DynamicIntPointerOpt{{value: RoutesProfileCountDftOpt}},
			IgnoreErrors: []*DynamicBoolOpt{{value: RoutesIgnoreErrorsDftOpt}},
			MaxCost:      []*DynamicInterfaceOpt{{Value: RoutesMaxCostDftOpt}},
			Limit:        []*DynamicIntPointerOpt{},
			Offset:       []*DynamicIntPointerOpt{},
			Usage:        []*DynamicDecimalOpt{{value: RatesUsageDftOpt}},
			MaxItems:     []*DynamicIntPointerOpt{},
		},
	}
	if !reflect.DeepEqual(expAttr, cfg.RouteSCfg()) {
		t.Errorf("Expected %s , received: %s ", utils.ToJSON(expAttr), utils.ToJSON(cfg.RouteSCfg()))
	}
}

func testCGRConfigV1ReloadConfigFromPathInvalidSection(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	cfg.rldCh = make(chan string, 100)
	expectedErr := "Invalid section: <InvalidSection> "
	var reply string
	cfg.ConfigPath = path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo2")
	if err := cfg.V1ReloadConfig(context.Background(), &ReloadArgs{
		Section: "InvalidSection",
	}, &reply); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %q. received %q", expectedErr, err.Error())
	}
}

func testV1ReloadConfigFromPathConfigSanity(t *testing.T) {
	expectedErr := "<AttributeS> not enabled but requested by <ChargerS> component"
	var reply string
	cfg := NewDefaultCGRConfig()
	cfg.rldCh = make(chan string, 100)
	cfg.ConfigPath = path.Join("/usr", "share", "cgrates", "conf", "samples", "tutinternal")
	if err := cfg.V1ReloadConfig(context.Background(), &ReloadArgs{
		Section: ChargerSJSON}, &reply); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func testLoadConfigFromHTTPValidURL(t *testing.T) {
	cfg := NewDefaultCGRConfig()

	url := "https://raw.githubusercontent.com/cgrates/cgrates/master/data/conf/samples/multifiles/a.json"
	if err := loadConfigFromHTTP(context.Background(), url, cfg.sections, cfg); err != nil {
		t.Error(err)
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

	cfg := NewDefaultCGRConfig()
	cfg.rldCh = make(chan string, 100)
	cfg.SessionSCfg().Enabled = true
	var reply string
	cfg.ConfigPath = path.Join("/usr", "share", "cgrates", "conf", "samples", "ers_example")
	if err := cfg.V1ReloadConfig(context.Background(), &ReloadArgs{
		Section: ERsJSON,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	flags := utils.FlagsWithParamsFromSlice([]string{"*dryRun"})
	flagsDefault := utils.FlagsWithParamsFromSlice([]string{})
	content := []*FCTemplate{
		{Tag: utils.ToR, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR, Type: utils.MetaVariable, Value: NewRSRParsersMustCompile("~*req.2", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
		{Tag: utils.OriginID, Path: utils.MetaCgreq + utils.NestingSep + utils.OriginID, Type: utils.MetaVariable, Value: NewRSRParsersMustCompile("~*req.3", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
		{Tag: utils.RequestType, Path: utils.MetaCgreq + utils.NestingSep + utils.RequestType, Type: utils.MetaVariable, Value: NewRSRParsersMustCompile("~*req.4", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
		{Tag: utils.Tenant, Path: utils.MetaCgreq + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable, Value: NewRSRParsersMustCompile("~*req.6", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
		{Tag: utils.Category, Path: utils.MetaCgreq + utils.NestingSep + utils.Category, Type: utils.MetaVariable, Value: NewRSRParsersMustCompile("~*req.7", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
		{Tag: utils.AccountField, Path: utils.MetaCgreq + utils.NestingSep + utils.AccountField, Type: utils.MetaVariable, Value: NewRSRParsersMustCompile("~*req.8", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
		{Tag: utils.Subject, Path: utils.MetaCgreq + utils.NestingSep + utils.Subject, Type: utils.MetaVariable, Value: NewRSRParsersMustCompile("~*req.9", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
		{Tag: utils.Destination, Path: utils.MetaCgreq + utils.NestingSep + utils.Destination, Type: utils.MetaVariable, Value: NewRSRParsersMustCompile("~*req.10", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
		{Tag: utils.SetupTime, Path: utils.MetaCgreq + utils.NestingSep + utils.SetupTime, Type: utils.MetaVariable, Value: NewRSRParsersMustCompile("~*req.11", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
		{Tag: utils.AnswerTime, Path: utils.MetaCgreq + utils.NestingSep + utils.AnswerTime, Type: utils.MetaVariable, Value: NewRSRParsersMustCompile("~*req.12", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
		{Tag: utils.Usage, Path: utils.MetaCgreq + utils.NestingSep + utils.Usage, Type: utils.MetaVariable, Value: NewRSRParsersMustCompile("~*req.13", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
	}
	for _, v := range content {
		v.ComputePath()
	}
	expAttr := &ERsCfg{
		Enabled:       true,
		SessionSConns: []string{utils.MetaLocalHost},
		Readers: []*EventReaderCfg{
			{
				ID:                   utils.MetaDefault,
				Type:                 utils.MetaNone,
				RunDelay:             0,
				ConcurrentReqs:       1024,
				SourcePath:           "/var/spool/cgrates/ers/in",
				ProcessedPath:        "/var/spool/cgrates/ers/out",
				Filters:              []string{},
				Reconnects:           -1,
				MaxReconnectInterval: 5 * time.Minute,
				Flags:                flagsDefault,
				Fields:               content,
				CacheDumpFields:      []*FCTemplate{},
				PartialCommitFields:  []*FCTemplate{},
				Opts: &EventReaderOpts{
					CSVFieldSeparator:   utils.StringPointer(","),
					PartialCacheAction:  utils.StringPointer(utils.MetaNone),
					CSVHeaderDefineChar: utils.StringPointer(":"),
					CSVRowLength:        utils.IntPointer(0),
					PartialOrderField:   utils.StringPointer("~*req.AnswerTime"),
					NATSSubject:         utils.StringPointer("cgrates_cdrs"),
				},
			},
			{
				ID:                   "file_reader1",
				Type:                 utils.MetaFileCSV,
				RunDelay:             -1,
				ConcurrentReqs:       1024,
				SourcePath:           "/tmp/ers/in",
				ProcessedPath:        "/tmp/ers/out",
				Filters:              []string{},
				Flags:                flags,
				Fields:               content,
				CacheDumpFields:      []*FCTemplate{},
				PartialCommitFields:  []*FCTemplate{},
				Reconnects:           -1,
				MaxReconnectInterval: 5 * time.Minute,
				Opts: &EventReaderOpts{
					CSVFieldSeparator:   utils.StringPointer(","),
					CSVHeaderDefineChar: utils.StringPointer(":"),
					CSVRowLength:        utils.IntPointer(0),
					PartialOrderField:   utils.StringPointer("~*req.AnswerTime"),
					PartialCacheAction:  utils.StringPointer(utils.MetaNone),

					NATSSubject: utils.StringPointer("cgrates_cdrs"),
				},
			},
		},
		PartialCacheTTL: time.Second,
	}
	if !reflect.DeepEqual(expAttr, cfg.ERsCfg()) {
		t.Errorf("Expected %s,\n received: %s ", utils.ToJSON(expAttr), utils.ToJSON(cfg.ERsCfg()))
	}
}

func testCGRConfigReloadDNSAgent(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	cfg.rldCh = make(chan string, 100)
	cfg.SessionSCfg().Enabled = true
	var reply string
	cfg.ConfigPath = path.Join("/usr", "share", "cgrates", "conf", "samples", "dnsagent_reload")
	if err := cfg.V1ReloadConfig(context.Background(), &ReloadArgs{
		Section: DNSAgentJSON,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	expAttr := &DNSAgentCfg{
		Enabled: true,
		Listeners: []Listener{
			{
				Address: ":2053",
				Network: "udp",
			},
			{
				Address: ":2054",
				Network: "tcp",
			},
		},
		SessionSConns: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
		// Timezone          string
		// RequestProcessors []*RequestProcessor
	}
	if !reflect.DeepEqual(expAttr, cfg.DNSAgentCfg()) {
		t.Errorf("Expected %s , received: %s ", utils.ToJSON(expAttr), utils.ToJSON(cfg.DNSAgentCfg()))
	}
}

func testCGRConfigReloadFreeswitchAgent(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	cfg.rldCh = make(chan string, 100)
	cfg.SessionSCfg().Enabled = true
	var reply string
	cfg.ConfigPath = path.Join("/usr", "share", "cgrates", "conf", "samples", "freeswitch_reload")
	if err := cfg.V1ReloadConfig(context.Background(), &ReloadArgs{
		Section: FreeSWITCHAgentJSON,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	expAttr := &FsAgentCfg{
		Enabled:                true,
		SessionSConns:          []string{utils.ConcatenatedKey(rpcclient.BiRPCInternal, utils.MetaSessionS)},
		SubscribePark:          true,
		ExtraFields:            RSRParsers{},
		MaxWaitConnection:      2 * time.Second,
		ActiveSessionDelimiter: ",",
		EventSocketConns: []*FsConnCfg{
			{
				Address:      "1.2.3.4:8021",
				Password:     "ClueCon",
				Reconnects:   5,
				Alias:        "1.2.3.4:8021",
				ReplyTimeout: time.Minute,
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
	content := []any{
		map[string]any{
			"path":      "*cgreq.ToR",
			"mandatory": true,
			"tag":       "ToR",
			"type":      "*variable",
			"value":     "~*req.2",
		},
		map[string]any{
			"path":      "*cgreq.OriginID",
			"mandatory": true,
			"tag":       "OriginID",
			"type":      "*variable",
			"value":     "~*req.3",
		},
		map[string]any{
			"path":      "*cgreq.RequestType",
			"mandatory": true,
			"tag":       "RequestType",
			"type":      "*variable",
			"value":     "~*req.4",
		},
		map[string]any{
			"path":      "*cgreq.Tenant",
			"mandatory": true,
			"tag":       "Tenant",
			"type":      "*variable",
			"value":     "~*req.6",
		},
		map[string]any{
			"path":      "*cgreq.Category",
			"mandatory": true,
			"tag":       "Category",
			"type":      "*variable",
			"value":     "~*req.7",
		},
		map[string]any{
			"path":      "*cgreq.Account",
			"mandatory": true,
			"tag":       "Account",
			"type":      "*variable",
			"value":     "~*req.8",
		},
		map[string]any{
			"path":      "*cgreq.Subject",
			"mandatory": true,
			"tag":       "Subject",
			"type":      "*variable",
			"value":     "~*req.9",
		},
		map[string]any{
			"path":      "*cgreq.Destination",
			"mandatory": true,
			"tag":       "Destination",
			"type":      "*variable",
			"value":     "~*req.10",
		},
		map[string]any{
			"path":      "*cgreq.SetupTime",
			"mandatory": true,
			"tag":       "SetupTime",
			"type":      "*variable",
			"value":     "~*req.11",
		},
		map[string]any{
			"path":      "*cgreq.AnswerTime",
			"mandatory": true,
			"tag":       "AnswerTime",
			"type":      "*variable",
			"value":     "~*req.12",
		},
		map[string]any{
			"path":      "*cgreq.Usage",
			"mandatory": true,
			"tag":       "Usage",
			"type":      "*variable",
			"value":     "~*req.13",
		},
	}
	expected := map[string]any{
		"enabled":           true,
		"partial_cache_ttl": "1s",
		"readers": []any{
			map[string]any{
				"id":                     utils.MetaDefault,
				"cache_dump_fields":      []any{},
				"concurrent_requests":    1024,
				"fields":                 content,
				"filters":                []string{},
				"flags":                  []string{},
				"run_delay":              "0",
				"source_path":            "/var/spool/cgrates/ers/in",
				"processed_path":         "/var/spool/cgrates/ers/out",
				"tenant":                 "",
				"timezone":               "",
				"type":                   utils.MetaNone,
				"reconnects":             -1,
				"max_reconnect_interval": "5m0s",
				"opts": map[string]any{
					"csvFieldSeparator":         ",",
					"csvHeaderDefineChar":       ":",
					"csvRowLength":              0.,
					"partialOrderField":         "~*req.AnswerTime",
					utils.PartialCacheActionOpt: utils.MetaNone,
					"natsSubject":               "cgrates_cdrs",
				},
				"partial_commit_fields": []any{},
			},
			map[string]any{
				"cache_dump_fields":      []any{},
				"concurrent_requests":    1024,
				"filters":                []string{},
				"flags":                  []string{"*dryRun"},
				"id":                     "file_reader1",
				"processed_path":         "/tmp/ers/out",
				"run_delay":              "-1",
				"source_path":            "/tmp/ers/in",
				"tenant":                 "",
				"timezone":               "",
				"type":                   "*fileCSV",
				"fields":                 content,
				"reconnects":             -1,
				"max_reconnect_interval": "5m0s",
				"opts": map[string]any{
					"csvFieldSeparator":         ",",
					"csvHeaderDefineChar":       ":",
					"csvRowLength":              0.,
					"partialOrderField":         "~*req.AnswerTime",
					utils.PartialCacheActionOpt: utils.MetaNone,
					"natsSubject":               "cgrates_cdrs",
				},
				"partial_commit_fields": []any{},
			},
		},
		"sessions_conns": []string{
			utils.MetaLocalHost,
		},
	}

	cfg := NewDefaultCGRConfig()
	cfg.rldCh = make(chan string, 100)
	var reply string
	var rcv map[string]any

	cfg.ConfigPath = "/usr/share/cgrates/conf/samples/ers_example"
	if err := cfg.V1ReloadConfig(context.Background(), &ReloadArgs{
		Section: ERsJSON,
	}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("Expected: %s \n,received: %s", utils.OK, reply)
	}

	expected = map[string]any{
		ERsJSON: expected,
	}
	if err := cfg.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{ERsJSON}}, &rcv); err != nil {
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
	cfg := NewDefaultCGRConfig()
	cfg.rldCh = make(chan string, 100)
	cfg.ChargerSCfg().Enabled = true
	cfg.CdrsCfg().Enabled = true
	var reply string
	if err := cfg.V1SetConfig(context.Background(), &SetConfigArgs{
		Config: map[string]any{
			"sessions": map[string]any{
				"enabled":          true,
				"resources_conns":  []string{"*localhost"},
				"routes_conns":     []string{"*localhost"},
				"attributes_conns": []string{"*localhost"},
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
		Enabled:         true,
		ListenBijson:    "127.0.0.1:2014",
		ChargerSConns:   []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)},
		ResourceSConns:  []string{utils.MetaLocalHost},
		ThresholdSConns: []string{},
		StatSConns:      []string{},
		AccountSConns:   []string{},
		RateSConns:      []string{},
		RouteSConns:     []string{utils.MetaLocalHost},
		AttributeSConns: []string{utils.MetaLocalHost},
		CDRsConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs)},

		ReplicationConns:  []string{},
		SessionIndexes:    utils.StringSet{},
		ClientProtocol:    1,
		TerminateAttempts: 5,
		AlterableFields:   utils.NewStringSet([]string{}),
		STIRCfg: &STIRcfg{
			AllowedAttest:      utils.NewStringSet([]string{utils.MetaAny}),
			PayloadMaxduration: -1,
			DefaultAttest:      "A",
		},
		ActionSConns: []string{},
		DefaultUsage: map[string]time.Duration{
			utils.MetaAny:   3 * time.Hour,
			utils.MetaVoice: 3 * time.Hour,
			utils.MetaData:  1048576,
			utils.MetaSMS:   1,
		},
		Opts: &SessionsOpts{
			Accounts:               []*DynamicBoolOpt{{}},
			Attributes:             []*DynamicBoolOpt{{}},
			CDRs:                   []*DynamicBoolOpt{{}},
			Chargers:               []*DynamicBoolOpt{{}},
			Resources:              []*DynamicBoolOpt{{}},
			Routes:                 []*DynamicBoolOpt{{}},
			Stats:                  []*DynamicBoolOpt{{}},
			Thresholds:             []*DynamicBoolOpt{{}},
			Initiate:               []*DynamicBoolOpt{{}},
			Update:                 []*DynamicBoolOpt{{}},
			Terminate:              []*DynamicBoolOpt{{}},
			Message:                []*DynamicBoolOpt{{}},
			AttributesDerivedReply: []*DynamicBoolOpt{{}},
			BlockerError:           []*DynamicBoolOpt{{}},
			CDRsDerivedReply:       []*DynamicBoolOpt{{}},
			ResourcesAuthorize:     []*DynamicBoolOpt{{}},
			ResourcesAllocate:      []*DynamicBoolOpt{{}},
			ResourcesRelease:       []*DynamicBoolOpt{{}},
			ResourcesDerivedReply:  []*DynamicBoolOpt{{}},
			RoutesDerivedReply:     []*DynamicBoolOpt{{}},
			StatsDerivedReply:      []*DynamicBoolOpt{{}},
			ThresholdsDerivedReply: []*DynamicBoolOpt{{}},
			MaxUsage:               []*DynamicBoolOpt{{}},
			ForceUsage:             []*DynamicBoolOpt{},
			TTL:                    []*DynamicDurationOpt{{value: SessionsTTLDftOpt}},
			Chargeable:             []*DynamicBoolOpt{{value: SessionsChargeableDftOpt}},
			TTLLastUsage:           []*DynamicDurationPointerOpt{},
			TTLLastUsed:            []*DynamicDurationPointerOpt{},
			DebitInterval:          []*DynamicDurationOpt{{value: SessionsDebitIntervalDftOpt}},
			TTLMaxDelay:            []*DynamicDurationOpt{{value: SessionsTTLMaxDelayDftOpt}},
			TTLUsage:               []*DynamicDurationPointerOpt{},
		},
	}
	if !reflect.DeepEqual(expAttr, cfg.SessionSCfg()) {
		t.Errorf("Expected %s ,\n received: %s ", utils.ToJSON(expAttr), utils.ToJSON(cfg.SessionSCfg()))
	}
}

func testCGRConfigReloadConfigFromStringSessionS(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	cfg.rldCh = make(chan string, 100)
	cfg.ChargerSCfg().Enabled = true
	cfg.CdrsCfg().Enabled = true
	var reply string
	if err := cfg.V1SetConfigFromJSON(context.Background(), &SetConfigFromJSONArgs{
		Config: `{"sessions":{
				"enabled":          true,
				"resources_conns":  ["*localhost"],
				"routes_conns":     ["*localhost"],
				"attributes_conns": ["*localhost"],
				"cdrs_conns":       ["*internal"],
				"chargers_conns":   ["*localhost"]
				}}`}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	expAttr := &SessionSCfg{
		Enabled:         true,
		ListenBijson:    "127.0.0.1:2014",
		ChargerSConns:   []string{utils.MetaLocalHost},
		ResourceSConns:  []string{utils.MetaLocalHost},
		ThresholdSConns: []string{},
		StatSConns:      []string{},
		AccountSConns:   []string{},
		RateSConns:      []string{},
		RouteSConns:     []string{utils.MetaLocalHost},
		AttributeSConns: []string{utils.MetaLocalHost},
		CDRsConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs)},

		ReplicationConns:  []string{},
		SessionIndexes:    utils.StringSet{},
		ClientProtocol:    1,
		TerminateAttempts: 5,
		AlterableFields:   utils.NewStringSet([]string{}),
		STIRCfg: &STIRcfg{
			AllowedAttest:      utils.NewStringSet([]string{utils.MetaAny}),
			PayloadMaxduration: -1,
			DefaultAttest:      "A",
		},
		ActionSConns: []string{},
		DefaultUsage: map[string]time.Duration{
			utils.MetaAny:   3 * time.Hour,
			utils.MetaVoice: 3 * time.Hour,
			utils.MetaData:  1048576,
			utils.MetaSMS:   1,
		},
		Opts: &SessionsOpts{
			Accounts:               []*DynamicBoolOpt{{}},
			Attributes:             []*DynamicBoolOpt{{}},
			CDRs:                   []*DynamicBoolOpt{{}},
			Chargers:               []*DynamicBoolOpt{{}},
			Resources:              []*DynamicBoolOpt{{}},
			Routes:                 []*DynamicBoolOpt{{}},
			Stats:                  []*DynamicBoolOpt{{}},
			Thresholds:             []*DynamicBoolOpt{{}},
			Initiate:               []*DynamicBoolOpt{{}},
			Update:                 []*DynamicBoolOpt{{}},
			Terminate:              []*DynamicBoolOpt{{}},
			Message:                []*DynamicBoolOpt{{}},
			AttributesDerivedReply: []*DynamicBoolOpt{{}},
			BlockerError:           []*DynamicBoolOpt{{}},
			CDRsDerivedReply:       []*DynamicBoolOpt{{}},
			ResourcesAuthorize:     []*DynamicBoolOpt{{}},
			ResourcesAllocate:      []*DynamicBoolOpt{{}},
			ResourcesRelease:       []*DynamicBoolOpt{{}},
			ResourcesDerivedReply:  []*DynamicBoolOpt{{}},
			RoutesDerivedReply:     []*DynamicBoolOpt{{}},
			StatsDerivedReply:      []*DynamicBoolOpt{{}},
			ThresholdsDerivedReply: []*DynamicBoolOpt{{}},
			MaxUsage:               []*DynamicBoolOpt{{}},
			ForceUsage:             []*DynamicBoolOpt{},
			TTL:                    []*DynamicDurationOpt{{value: SessionsTTLDftOpt}},
			Chargeable:             []*DynamicBoolOpt{{value: SessionsChargeableDftOpt}},
			TTLLastUsage:           []*DynamicDurationPointerOpt{},
			TTLLastUsed:            []*DynamicDurationPointerOpt{},
			DebitInterval:          []*DynamicDurationOpt{{value: SessionsDebitIntervalDftOpt}},
			TTLMaxDelay:            []*DynamicDurationOpt{{value: SessionsTTLMaxDelayDftOpt}},
			TTLUsage:               []*DynamicDurationPointerOpt{},
		},
	}
	if !reflect.DeepEqual(expAttr, cfg.SessionSCfg()) {
		t.Errorf("Expected %s ,\n received: %s ", utils.ToJSON(expAttr), utils.ToJSON(cfg.SessionSCfg()))
	}

	var rcv string
	expected := `{"sessions":{"accounts_conns":[],"actions_conns":[],"alterable_fields":[],"attributes_conns":["*localhost"],"cdrs_conns":["*internal"],"channel_sync_interval":"0","chargers_conns":["*localhost"],"client_protocol":1,"default_usage":{"*any":"3h0m0s","*data":"1048576","*sms":"1","*voice":"3h0m0s"},"enabled":true,"listen_bigob":"","listen_bijson":"127.0.0.1:2014","min_dur_low_balance":"0","opts":{"*accounts":[{"FilterIDs":null,"Tenant":""}],"*attributes":[{"FilterIDs":null,"Tenant":""}],"*attributesDerivedReply":[{"FilterIDs":null,"Tenant":""}],"*blockerError":[{"FilterIDs":null,"Tenant":""}],"*cdrs":[{"FilterIDs":null,"Tenant":""}],"*cdrsDerivedReply":[{"FilterIDs":null,"Tenant":""}],"*chargeable":[{"FilterIDs":null,"Tenant":""}],"*chargers":[{"FilterIDs":null,"Tenant":""}],"*debitInterval":[{"FilterIDs":null,"Tenant":""}],"*forceUsage":[],"*initiate":[{"FilterIDs":null,"Tenant":""}],"*maxUsage":[{"FilterIDs":null,"Tenant":""}],"*message":[{"FilterIDs":null,"Tenant":""}],"*resources":[{"FilterIDs":null,"Tenant":""}],"*resourcesAllocate":[{"FilterIDs":null,"Tenant":""}],"*resourcesAuthorize":[{"FilterIDs":null,"Tenant":""}],"*resourcesDerivedReply":[{"FilterIDs":null,"Tenant":""}],"*resourcesRelease":[{"FilterIDs":null,"Tenant":""}],"*routes":[{"FilterIDs":null,"Tenant":""}],"*routesDerivedReply":[{"FilterIDs":null,"Tenant":""}],"*stats":[{"FilterIDs":null,"Tenant":""}],"*statsDerivedReply":[{"FilterIDs":null,"Tenant":""}],"*terminate":[{"FilterIDs":null,"Tenant":""}],"*thresholds":[{"FilterIDs":null,"Tenant":""}],"*thresholdsDerivedReply":[{"FilterIDs":null,"Tenant":""}],"*ttl":[{"FilterIDs":null,"Tenant":""}],"*ttlLastUsage":[],"*ttlLastUsed":[],"*ttlMaxDelay":[{"FilterIDs":null,"Tenant":""}],"*ttlUsage":[],"*update":[{"FilterIDs":null,"Tenant":""}]},"rates_conns":[],"replication_conns":[],"resources_conns":["*localhost"],"routes_conns":["*localhost"],"session_indexes":[],"stats_conns":[],"stir":{"allowed_attest":["*any"],"default_attest":"A","payload_maxduration":"-1","privatekey_path":"","publickey_path":""},"store_session_costs":false,"terminate_attempts":5,"thresholds_conns":[]}}`
	if err := cfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{SessionSJSON}}, &rcv); err != nil {
		t.Error(err)
	} else if expected != rcv {
		t.Errorf("Expected: %+s, \n received: %s", expected, rcv)
	}
}

func testCGRConfigReloadAll(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	cfg.rldCh = make(chan string, 100)
	cfg.ChargerSCfg().Enabled = true
	cfg.CdrsCfg().Enabled = true
	var reply string
	cfg.ConfigPath = path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo2")
	if err := cfg.V1ReloadConfig(context.Background(), &ReloadArgs{
		Section: utils.MetaAll,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	expAttr := &SessionSCfg{
		Enabled:         true,
		ListenBijson:    "127.0.0.1:2014",
		ChargerSConns:   []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)},
		ResourceSConns:  []string{utils.MetaLocalHost},
		ThresholdSConns: []string{},
		StatSConns:      []string{},
		AccountSConns:   []string{},
		RateSConns:      []string{},
		RouteSConns:     []string{utils.MetaLocalHost},
		AttributeSConns: []string{utils.MetaLocalHost},
		CDRsConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs)},

		ReplicationConns:  []string{},
		SessionIndexes:    utils.StringSet{},
		ClientProtocol:    1,
		TerminateAttempts: 5,
		AlterableFields:   utils.NewStringSet([]string{}),
		STIRCfg: &STIRcfg{
			AllowedAttest:      utils.NewStringSet([]string{utils.MetaAny}),
			PayloadMaxduration: -1,
			DefaultAttest:      "A",
		},
		ActionSConns: make([]string, 0),
		DefaultUsage: map[string]time.Duration{
			utils.MetaAny:   3 * time.Hour,
			utils.MetaVoice: 3 * time.Hour,
			utils.MetaData:  1048576,
			utils.MetaSMS:   1,
		},
		Opts: &SessionsOpts{
			Accounts:               []*DynamicBoolOpt{{}},
			Attributes:             []*DynamicBoolOpt{{}},
			CDRs:                   []*DynamicBoolOpt{{}},
			Chargers:               []*DynamicBoolOpt{{}},
			Resources:              []*DynamicBoolOpt{{}},
			Routes:                 []*DynamicBoolOpt{{}},
			Stats:                  []*DynamicBoolOpt{{}},
			Thresholds:             []*DynamicBoolOpt{{}},
			Initiate:               []*DynamicBoolOpt{{}},
			Update:                 []*DynamicBoolOpt{{}},
			Terminate:              []*DynamicBoolOpt{{}},
			Message:                []*DynamicBoolOpt{{}},
			AttributesDerivedReply: []*DynamicBoolOpt{{}},
			BlockerError:           []*DynamicBoolOpt{{}},
			CDRsDerivedReply:       []*DynamicBoolOpt{{}},
			ResourcesAuthorize:     []*DynamicBoolOpt{{}},
			ResourcesAllocate:      []*DynamicBoolOpt{{}},
			ResourcesRelease:       []*DynamicBoolOpt{{}},
			ResourcesDerivedReply:  []*DynamicBoolOpt{{}},
			RoutesDerivedReply:     []*DynamicBoolOpt{{}},
			StatsDerivedReply:      []*DynamicBoolOpt{{}},
			ThresholdsDerivedReply: []*DynamicBoolOpt{{}},
			MaxUsage:               []*DynamicBoolOpt{{}},
			ForceUsage:             []*DynamicBoolOpt{},
			TTL:                    []*DynamicDurationOpt{{value: SessionsTTLDftOpt}},
			Chargeable:             []*DynamicBoolOpt{{value: SessionsChargeableDftOpt}},
			TTLLastUsage:           []*DynamicDurationPointerOpt{},
			TTLLastUsed:            []*DynamicDurationPointerOpt{},
			DebitInterval:          []*DynamicDurationOpt{{value: SessionsDebitIntervalDftOpt}},
			TTLMaxDelay:            []*DynamicDurationOpt{{value: SessionsTTLMaxDelayDftOpt}},
			TTLUsage:               []*DynamicDurationPointerOpt{},
		},
	}
	if !reflect.DeepEqual(expAttr, cfg.SessionSCfg()) {
		t.Errorf("Expected %s ,\n received: %s ", utils.ToJSON(expAttr), utils.ToJSON(cfg.SessionSCfg()))
	}
}

func testHttpHandlerConfigSForNotExistFile(t *testing.T) {
	cgrCfg.configSCfg.RootDir = "/usr/share/cgrates/"
	req := httptest.NewRequest("GET", "http://127.0.0.1/conf/samples/NotExists/cgrates.json", nil)
	w := httptest.NewRecorder()
	HandlerConfigS(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.Status != "404 Not Found" {
		t.Errorf("Expected %+v , received: %+v ", "200 OK", resp.Status)
	}
	httpBodyMsgError := "stat /usr/share/cgrates/conf/samples/NotExists/cgrates.json: no such file or directory"
	if httpBodyMsgError != string(body) {
		t.Errorf("Expected %s , received: %s ", httpBodyMsgError, string(body))
	}
}

func testHandleConfigSFolderError(t *testing.T) {
	flPath := "/usr/share/cgrates/conf/samples/NotExists/cgrates.json"
	w := httptest.NewRecorder()
	handleConfigSFolder(context.Background(), flPath, w)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	expected := "path:\"/usr/share/cgrates/conf/samples/NotExists/cgrates.json\" is not reachable"
	if expected != string(body) {
		t.Errorf("Expected %s , received: %s ", expected, string(body))
	}
}

func testHttpHandlerConfigSInvalidPath(t *testing.T) {
	cgrCfg.configSCfg.RootDir = "/usr/share/\x00/"
	req := httptest.NewRequest("GET", "http://127.0.0.1/conf/samples/tutmysql/cgrates.json", nil)
	w := httptest.NewRecorder()
	HandlerConfigS(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.Status != "500 Internal Server Error" {
		t.Errorf("Expected:%+v, received:%+v", "200 OK", resp.Status)
	}
	httpBodyMsgError := "stat /usr/share/\x00/conf/samples/tutmysql/cgrates.json: invalid argument"
	if httpBodyMsgError != string(body) {
		t.Errorf("Received:%q, expected:%q", string(body), httpBodyMsgError)
	}
}

func testHttpHandlerConfigSForFile(t *testing.T) {
	cgrCfg.configSCfg.RootDir = "/usr/share/cgrates/"
	req := httptest.NewRequest("GET", "http://127.0.0.1/conf/samples/tutmysql/cgrates.json", nil)
	w := httptest.NewRecorder()
	HandlerConfigS(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.Status != "200 OK" {
		t.Errorf("Expected %+v , received: %+v ", "200 OK", resp.Status)
	}
	if dat, err := os.ReadFile("/usr/share/cgrates/conf/samples/tutmysql/cgrates.json"); err != nil {
		t.Error(err)
	} else if string(dat) != string(body) {
		t.Errorf("Expected %s , received: %s ", string(dat), string(body))
	}
}

func testHttpHandlerConfigSForNotExistFolder(t *testing.T) {
	cgrCfg.configSCfg.RootDir = "/usr/share/cgrates/"
	req := httptest.NewRequest("GET", "http://127.0.0.1/conf/samples/NotExists/", nil)
	w := httptest.NewRecorder()
	HandlerConfigS(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.Status != "404 Not Found" {
		t.Errorf("Expected %+v , received: %+v ", "200 OK", resp.Status)
	}
	httpBodyMsgError := "stat /usr/share/cgrates/conf/samples/NotExists: no such file or directory"
	if httpBodyMsgError != string(body) {
		t.Errorf("Expected %s , received: %s ", httpBodyMsgError, string(body))
	}
}

func testHttpHandlerConfigSForFolder(t *testing.T) {
	cgrCfg.configSCfg.RootDir = "/usr/share/cgrates/"
	req := httptest.NewRequest("GET", "http://127.0.0.1/conf/samples/diamagent_internal/", nil)
	w := httptest.NewRecorder()
	HandlerConfigS(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.Status != "200 OK" {
		t.Errorf("Expected %+v , received: %+v ", "200 OK", resp.Status)
	}
	cfg, err := NewCGRConfigFromPath(context.Background(), "/usr/share/cgrates/conf/samples/diamagent_internal/")
	if err != nil {
		t.Error(err)
	}
	mp := cfg.AsMapInterface(cfg.generalCfg.RSRSep)

	str := utils.ToJSON(mp)
	// we compare the length of the string because flags is a map and we receive it in different order
	if len(str) != len(string(body)) {
		t.Errorf("Expected %s ,\n\n received: %s ", str, string(body))
	}
}

func testLoadConfigFromFolderFileNotFound(t *testing.T) {
	cfg := NewDefaultCGRConfig()

	expected := "file </usr/share/cgrates/conf/samples/docker/cgrates.json>:NOT_FOUND:ENV_VAR:DOCKER_IP"
	if err := loadConfigFromFolder(context.Background(), "/usr/share/cgrates/conf/samples/docker",
		cfg.sections, false, cfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func testLoadConfigFromFolderOpenError(t *testing.T) {
	newDir := "/tmp/testLoadConfigFromFolderOpenError"
	if err := os.MkdirAll(newDir, 0755); err != nil {
		t.Fatal(err)
	}
	cfg := NewDefaultCGRConfig()

	expected := "open /tmp/testLoadConfigFromFolderOpenError/notes.json: no such file or directory"
	if err := loadConfigFromFile(context.Background(), path.Join(newDir, "notes.json"),
		cfg.sections, false, cfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	if err := os.RemoveAll(newDir); err != nil {
		t.Fatal(err)
	}
}

func testLoadConfigFromFolderNoConfigFound(t *testing.T) {
	newDir := "/tmp/[]"
	if err := os.MkdirAll(newDir, 0755); err != nil {
		t.Fatal(err)
	}
	cfg := NewDefaultCGRConfig()

	if err := loadConfigFromFolder(context.Background(), newDir,
		cfg.sections, false, cfg); err == nil || err != filepath.ErrBadPattern {
		t.Errorf("Expected %+v, received %+v", filepath.ErrBadPattern, err)
	}
	if err := os.RemoveAll(newDir); err != nil {
		t.Fatal(err)
	}
}

func testLoadConfigFromPathInvalidArgument(t *testing.T) {
	cfg := NewDefaultCGRConfig()

	expected := "stat /\x00: invalid argument"
	if err := loadConfigFromPath(context.Background(), "/\x00",
		cfg.sections, false, cfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func testLoadConfigFromPathValidPath(t *testing.T) {
	newDir := "/tmp/testLoadConfigFromPathValidPath"
	if err := os.MkdirAll(newDir, 0755); err != nil {
		t.Fatal(err)
	}
	cfg := NewDefaultCGRConfig()

	expected := "No config file found on path /tmp/testLoadConfigFromPathValidPath "
	if err := loadConfigFromPath(context.Background(), newDir,
		cfg.sections, false, cfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	if err := os.RemoveAll(newDir); err != nil {
		t.Fatal(err)
	}
}

func testLoadConfigFromPathFile(t *testing.T) {
	newDir := "/tmp/testLoadConfigFromPathFile"
	if err := os.MkdirAll(newDir, 0755); err != nil {
		t.Fatal(err)
	}
	cfg := NewDefaultCGRConfig()

	expected := "No config file found on path /tmp/testLoadConfigFromPathFile "
	if err := loadConfigFromPath(context.Background(), newDir,
		cfg.sections, false, cfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	if err := os.RemoveAll(newDir); err != nil {
		t.Fatal(err)
	}
}

func testApisLoadFromPath(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	// pathL := "/tmp/test.json"
	if err := os.Mkdir(path.Join("/tmp", "TestApisLoadFromPath"), 0777); err != nil {
		t.Error(err)
	}
	file, err := os.Create(path.Join("/tmp", "TestApisLoadFromPath", "test.json"))
	if err != nil {
		t.Error(err)
	}
	defer file.Close()
	data := `
	{
		// CGRateS Configuration file
		//
		
		
		"general": {
			"reply_timeout": "50s"
		},
		
		"logger": {
			"level": 7,
		},
		
		"listen": {
			"rpc_json": ":2012",
			"rpc_gob": ":2013",
			"http": ":2080"
		},
		
		
		"data_db": {
			"db_type": "*internal"
		},
			
		
		"rals": {
			"enabled": true,
			"thresholds_conns": ["*internal"],
			"max_increments":3000000
		},
		
		
		"schedulers": {
			"enabled": true,
			"cdrs_conns": ["*internal"],
			"stats_conns": ["*localhost"]
		},
		
		
		"cdrs": {
			"enabled": true,
			"chargers_conns":["*internal"]
		},
		
		
		"attributes": {
			"enabled": true,
			"stats_conns": ["*localhost"],
			"resources_conns": ["*localhost"],
			"accounts_conns": ["*localhost"]
		},
		
		
		"chargers": {
			"enabled": true,
			"attributes_conns": ["*internal"]
		},
		
		
		"resources": {
			"enabled": true,
			"store_interval": "-1",
			"thresholds_conns": ["*internal"]
		},
		
		
		"stats": {
			"enabled": true,
			"store_interval": "-1",
			"thresholds_conns": ["*internal"]
		},
		
		"thresholds": {
			"enabled": true,
			"store_interval": "-1"
		},
		
		
		"routes": {
			"enabled": true,
			"prefix_indexed_fields":["*req.Destination"],
			"stats_conns": ["*internal"],
			"resources_conns": ["*internal"],
		},
		
		
		"sessions": {
			"enabled": true,
			"routes_conns": ["*internal"],
			"resources_conns": ["*internal"],
			"attributes_conns": ["*internal"],
			"cdrs_conns": ["*internal"],
			"chargers_conns": ["*internal"]
		},
		
		
		"admins": {
			"enabled": true,
			"scheduler_conns": ["*internal"]
		},
		
		
		"rates": {
			"enabled": true
		},
		
		
		"actions": {
			"enabled": true,
			"accounts_conns": ["*localhost"]
		},
		
		
		"accounts": {
			"enabled": true
		},
		
		
		"filters": {
			"stats_conns": ["*internal"],
			"resources_conns": ["*internal"],
			"accounts_conns": ["*internal"],
		},
		
		}
	`
	_, err = file.Write([]byte(data))
	if err != nil {
		t.Error(err)
	}

	if err := cfg.LoadFromPath(context.Background(), path.Join("/tmp", "TestApisLoadFromPath")); err != nil {
		t.Error(err)
	}

	if err := os.RemoveAll(path.Join("/tmp", "TestApisLoadFromPath")); err != nil {
		t.Error(err)
	}
}

func testHandleConfigSFilesError(t *testing.T) {
	flPath := "/usr/share/cgrates/conf/samples/NotExists/cgrates.json"
	w := httptest.NewRecorder()
	handleConfigSFile(flPath, w)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	expected := "open /usr/share/cgrates/conf/samples/NotExists/cgrates.json: no such file or directory"
	if expected != string(body) {
		t.Errorf("Expected %s , received: %s ", expected, string(body))
	}
}

func testHandleConfigSFileErrorWrite(t *testing.T) {
	flPath := "/tmp/testHandleConfigSFilesErrorWrite"
	if err := os.MkdirAll(flPath, 0777); err != nil {
		t.Fatal(err)
	}
	newFile, err := os.Create(path.Join(flPath, "random.json"))
	if err != nil {
		t.Error(err)
	}
	newFile.Write([]byte(`{}`))
	newFile.Close()

	w := new(responseTest)
	handleConfigSFile(path.Join(flPath, "random.json"), w)

	if err := os.Remove(path.Join(flPath, "random.json")); err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll(flPath); err != nil {
		t.Fatal(err)
	}
}

func testHandleConfigSFolderErrorWrite(t *testing.T) {
	flPath := "/tmp/testHandleConfigSFilesErrorWrite"
	if err := os.MkdirAll(flPath, 0777); err != nil {
		t.Fatal(err)
	}
	newFile, err := os.Create(path.Join(flPath, "random.json"))
	if err != nil {
		t.Error(err)
	}
	newFile.Write([]byte(`{}`))
	newFile.Close()

	w := new(responseTest)
	handleConfigSFolder(context.Background(), flPath, w)

	if err := os.Remove(path.Join(flPath, "random.json")); err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll(flPath); err != nil {
		t.Fatal(err)
	}
}

type responseTest struct{}

func (responseTest) Header() http.Header {
	return nil
}

func (responseTest) Write([]byte) (int, error) {
	return 0, errors.New("Invalid section")
}

func (responseTest) WriteHeader(int) {}

func TestGetLockFilePath(t *testing.T) {
	l := LoaderSCfg{
		ID:           "file",
		LockFilePath: "../tmp/file.txt",
		TpInDir:      "/home",
	}

	exp := "/tmp/file.txt"
	pathL := l.GetLockFilePath()
	if pathL != exp {
		t.Errorf("Expected %s \n but received \n %s", exp, pathL)
	}

	l.LockFilePath = "file.txt"
	pathL = l.GetLockFilePath()
	exp = "/home/file.txt"
	if pathL != exp {
		t.Errorf("Expected %s \n but received \n %s", exp, pathL)
	}

	if err := os.Mkdir(path.Join("/tmp", "TestGetLockFilePath"), 0777); err != nil {
		t.Error(err)
	}
	l.LockFilePath = "TestGetLockFilePath"
	l.TpInDir = "/tmp"
	pathL = l.GetLockFilePath()
	exp = "/tmp/TestGetLockFilePath/file.lck"
	if pathL != exp {
		t.Errorf("Expected %s \n but received \n %s", exp, pathL)
	}
	if err := os.RemoveAll(path.Join("/tmp", "TestGetLockFilePath")); err != nil {
		t.Error(err)
	}
}

func TestReloadCfgInDb(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	db := &CgrJsonCfg{}
	cfg.db = db
	cfg.attributeSCfg = &AttributeSCfg{
		Enabled:                true,
		ResourceSConns:         []string{"*internal"},
		StatSConns:             []string{"*internal"},
		AccountSConns:          []string{"*internal"},
		IndexedSelects:         false,
		StringIndexedFields:    &[]string{"field1"},
		SuffixIndexedFields:    &[]string{"field1"},
		PrefixIndexedFields:    &[]string{"field1"},
		ExistsIndexedFields:    &[]string{"field1"},
		NotExistsIndexedFields: &[]string{"field1"},
		Opts: &AttributesOpts{
			ProcessRuns: []*DynamicIntOpt{
				{
					FilterIDs: []string{},
					value:     2,
				},
			},
		},
		NestedFields: true,
	}
	var reply string
	cfg.sections = newSections(cfg)
	cfg.rldCh = make(chan string, 100)
	cfg.ConfigPath = path.Join("/usr", "share", "cgrates", "conf", "samples", "attributes_internal")
	jsn := &AttributeSJsonCfg{
		Enabled:                  utils.BoolPointer(false),
		Resources_conns:          &[]string{"*localhost"},
		Stats_conns:              &[]string{"*localhost"},
		Accounts_conns:           &[]string{"*localhost"},
		Indexed_selects:          utils.BoolPointer(true),
		String_indexed_fields:    &[]string{"field2"},
		Suffix_indexed_fields:    &[]string{"field2"},
		Prefix_indexed_fields:    &[]string{"field2"},
		Exists_indexed_fields:    &[]string{"field2"},
		Notexists_indexed_fields: &[]string{"field2"},
		Opts: &AttributesOptsJson{
			ProcessRuns: []*DynamicInterfaceOpt{
				{
					Value: "3",
				},
			},
		},
		Nested_fields: utils.BoolPointer(false),
	}
	db.SetSection(context.Background(), AttributeSJSON, jsn)
	expected := &AttributeSCfg{
		Enabled:                false,
		ResourceSConns:         []string{"*localhost"},
		StatSConns:             []string{"*localhost"},
		AccountSConns:          []string{"*localhost"},
		IndexedSelects:         true,
		StringIndexedFields:    &[]string{"field2"},
		SuffixIndexedFields:    &[]string{"field2"},
		PrefixIndexedFields:    &[]string{"field2"},
		ExistsIndexedFields:    &[]string{"field2"},
		NotExistsIndexedFields: &[]string{"field2"},
		Opts: &AttributesOpts{
			ProcessRuns: []*DynamicIntOpt{
				{
					FilterIDs: nil,
					value:     3,
				},
				{
					FilterIDs: []string{},
					value:     2,
				},
			},
		},
		NestedFields: false,
	}
	args2 := &ReloadArgs{
		Section: AttributeSJSON,
	}
	if err := cfg.V1ReloadConfig(context.Background(), args2, &reply); err != nil {
		t.Error(err)
	}

	rcv := cfg.AttributeSCfg()
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}
