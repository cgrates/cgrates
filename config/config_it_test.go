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

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

var (
	cgrTests = []func(t *testing.T){
		testNewCgrJsonCfgFromHttp,
		testNewCGRConfigFromPath,
		testCGRConfigReloadAttributeS,
		testCGRConfigReloadChargerSDryRun,
		testCGRConfigReloadChargerS,
		testCGRConfigReloadThresholdS,
		testCGRConfigReloadStatS,
		testCGRConfigReloadResourceS,
		testCGRConfigReloadSupplierS,
		testCGRConfigReloadCDRs,
		testCGRConfigReloadRALs,
		testCGRConfigReloadSessionS,
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

	err = expVal.loadConfigFromPath(path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo"),
		[]func(*CgrJsonCfg) error{expVal.loadFromJSONCfg}, false)
	if err != nil {
		t.Fatal(err)
	}

	if _, err = net.DialTimeout(utils.TCP, addr, time.Second); err != nil { // check if site is up
		return
	}

	rply := NewDefaultCGRConfig()
	if err = rply.loadConfigFromPath(addr, []func(*CgrJsonCfg) error{rply.loadFromJSONCfg}, false); err != nil {
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
	cfg := NewDefaultCGRConfig()
	for _, section := range sortedCfgSections {
		cfg.rldChans[section] = make(chan struct{}, 1)
	}
	var reply string
	if err = cfg.V1ReloadConfig(&ReloadArgs{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo2"),
		Section: ATTRIBUTE_JSN,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	expAttr := &AttributeSCfg{
		Enabled:             true,
		ApierSConns:         []string{},
		ResourceSConns:      []string{},
		StatSConns:          []string{},
		StringIndexedFields: &[]string{utils.MetaReq + utils.NestingSep + utils.AccountField},
		PrefixIndexedFields: &[]string{},
		SuffixIndexedFields: &[]string{},
		IndexedSelects:      true,
		ProcessRuns:         1,
	}
	if !reflect.DeepEqual(expAttr, cfg.AttributeSCfg()) {
		t.Errorf("Expected %s , received: %s ", utils.ToJSON(expAttr), utils.ToJSON(cfg.AttributeSCfg()))
	}
}

func testCGRConfigReloadChargerSDryRun(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	for _, section := range sortedCfgSections {
		cfg.rldChans[section] = make(chan struct{}, 1)
	}
	var reply string
	if err = cfg.V1ReloadConfig(&ReloadArgs{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo2"),
		Section: ChargerSCfgJson,
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
	for _, section := range sortedCfgSections {
		cfg.rldChans[section] = make(chan struct{}, 1)
	}
	var reply string
	if err = cfg.V1ReloadConfig(&ReloadArgs{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo2"),
		Section: ChargerSCfgJson,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	expAttr := &ChargerSCfg{
		Enabled:             true,
		StringIndexedFields: &[]string{utils.MetaReq + utils.NestingSep + utils.AccountField},
		PrefixIndexedFields: &[]string{},
		SuffixIndexedFields: &[]string{},
		IndexedSelects:      true,
		AttributeSConns:     []string{"*localhost"},
	}
	if !reflect.DeepEqual(expAttr, cfg.ChargerSCfg()) {
		t.Errorf("Expected %s , received: %s ", utils.ToJSON(expAttr), utils.ToJSON(cfg.ChargerSCfg()))
	}
}

func testCGRConfigReloadThresholdS(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	for _, section := range sortedCfgSections {
		cfg.rldChans[section] = make(chan struct{}, 1)
	}
	var reply string
	if err = cfg.V1ReloadConfig(&ReloadArgs{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo2"),
		Section: THRESHOLDS_JSON,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	expAttr := &ThresholdSCfg{
		Enabled:             true,
		StringIndexedFields: &[]string{utils.MetaReq + utils.NestingSep + utils.AccountField},
		PrefixIndexedFields: &[]string{},
		SuffixIndexedFields: &[]string{},
		IndexedSelects:      true,
	}
	if !reflect.DeepEqual(expAttr, cfg.ThresholdSCfg()) {
		t.Errorf("Expected %s , received: %s ", utils.ToJSON(expAttr), utils.ToJSON(cfg.ThresholdSCfg()))
	}
}

func testCGRConfigReloadStatS(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	for _, section := range sortedCfgSections {
		cfg.rldChans[section] = make(chan struct{}, 1)
	}
	var reply string
	if err = cfg.V1ReloadConfig(&ReloadArgs{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo2"),
		Section: STATS_JSON,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	expAttr := &StatSCfg{
		Enabled:             true,
		StringIndexedFields: &[]string{utils.MetaReq + utils.NestingSep + utils.AccountField},
		PrefixIndexedFields: &[]string{},
		SuffixIndexedFields: &[]string{},
		IndexedSelects:      true,
		ThresholdSConns:     []string{utils.MetaLocalHost},
	}
	if !reflect.DeepEqual(expAttr, cfg.StatSCfg()) {
		t.Errorf("Expected %s , received: %s ", utils.ToJSON(expAttr), utils.ToJSON(cfg.StatSCfg()))
	}
}

func testCGRConfigReloadResourceS(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	for _, section := range sortedCfgSections {
		cfg.rldChans[section] = make(chan struct{}, 1)
	}
	var reply string
	if err = cfg.V1ReloadConfig(&ReloadArgs{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo2"),
		Section: RESOURCES_JSON,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	expAttr := &ResourceSConfig{
		Enabled:             true,
		StringIndexedFields: &[]string{utils.MetaReq + utils.NestingSep + utils.AccountField},
		PrefixIndexedFields: &[]string{},
		SuffixIndexedFields: &[]string{},
		IndexedSelects:      true,
		ThresholdSConns:     []string{utils.MetaLocalHost},
	}
	if !reflect.DeepEqual(expAttr, cfg.ResourceSCfg()) {
		t.Errorf("Expected %s , received: %s ", utils.ToJSON(expAttr), utils.ToJSON(cfg.ResourceSCfg()))
	}
}

func testCGRConfigReloadSupplierS(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	for _, section := range sortedCfgSections {
		cfg.rldChans[section] = make(chan struct{}, 1)
	}
	var reply string
	if err = cfg.V1ReloadConfig(&ReloadArgs{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo2"),
		Section: RouteSJson,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	expAttr := &RouteSCfg{
		Enabled:             true,
		StringIndexedFields: &[]string{"*req.LCRProfile"},
		PrefixIndexedFields: &[]string{utils.MetaReq + utils.NestingSep + utils.Destination},
		SuffixIndexedFields: &[]string{},
		ResourceSConns:      []string{},
		StatSConns:          []string{},
		AttributeSConns:     []string{},
		RALsConns:           []string{},
		IndexedSelects:      true,
		DefaultRatio:        1,
	}
	if !reflect.DeepEqual(expAttr, cfg.RouteSCfg()) {
		t.Errorf("Expected %s , received: %s ", utils.ToJSON(expAttr), utils.ToJSON(cfg.RouteSCfg()))
	}
}

func testCGRConfigV1ReloadConfigFromPathInvalidSection(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	for _, section := range sortedCfgSections {
		cfg.rldChans[section] = make(chan struct{}, 1)
	}
	expectedErr := "Invalid section: <InvalidSection>"
	var reply string
	if err := cfg.V1ReloadConfig(&ReloadArgs{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo2"),
		Section: "InvalidSection",
	}, &reply); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v. received %+v", expectedErr, err)
	}

	expectedErr = utils.NewErrMandatoryIeMissing("Path").Error()
	if err := cfg.V1ReloadConfig(&ReloadArgs{
		Section: "InvalidSection",
	}, &reply); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v. received %+v", expectedErr, err)
	}
}

func testV1ReloadConfigFromPathConfigSanity(t *testing.T) {
	expectedErr := "<AttributeS> not enabled but requested by <ChargerS> component"
	var reply string
	cfg := NewDefaultCGRConfig()
	for _, section := range sortedCfgSections {
		cfg.rldChans[section] = make(chan struct{}, 1)
	}
	if err := cfg.V1ReloadConfig(&ReloadArgs{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutinternal"),
		Section: ChargerSCfgJson}, &reply); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func testLoadConfigFromHTTPValidURL(t *testing.T) {
	cfg := NewDefaultCGRConfig()

	url := "https://raw.githubusercontent.com/cgrates/cgrates/master/data/conf/samples/multifiles/a.json"
	if err := cfg.loadConfigFromHTTP(url, nil); err != nil {
		t.Error(err)
	}
}

func testCGRConfigReloadCDRs(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	for _, section := range sortedCfgSections {
		cfg.rldChans[section] = make(chan struct{}, 1)
	}
	cfg.RalsCfg().Enabled = true
	var reply string
	if err = cfg.V1ReloadConfig(&ReloadArgs{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo2"),
		Section: CDRS_JSN,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	rsr, err := NewRSRParsersFromSlice([]string{"~*req.PayPalAccount", "~*req.LCRProfile", "~*req.ResourceID"})
	if err != nil {
		t.Fatal(err)
	}
	expAttr := &CdrsCfg{
		Enabled:         true,
		ExtraFields:     rsr,
		ChargerSConns:   []string{utils.MetaLocalHost},
		RaterConns:      []string{},
		AttributeSConns: []string{},
		ThresholdSConns: []string{},
		StatSConns:      []string{},
		SMCostRetries:   5,
		StoreCdrs:       true,
		ActionSConns:    []string{},
		EEsConns:        []string{utils.MetaLocalHost},
	}
	if !reflect.DeepEqual(expAttr, cfg.CdrsCfg()) {
		t.Errorf("Expected %s , received: %s ", utils.ToJSON(expAttr), utils.ToJSON(cfg.CdrsCfg()))
	}
}

func testCGRConfigReloadRALs(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	for _, section := range sortedCfgSections {
		cfg.rldChans[section] = make(chan struct{}, 1)
	}
	blMap := cfg.RalsCfg().BalanceRatingSubject
	maxComp := cfg.RalsCfg().MaxComputedUsage
	var reply string
	if err = cfg.V1ReloadConfig(&ReloadArgs{
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
		CacheSConns:             []string{utils.MetaInternal + utils.InInFieldSep + utils.MetaCaches},
		ThresholdSConns:         []string{utils.MetaLocalHost},
		StatSConns:              []string{utils.MetaLocalHost},
		MaxIncrements:           1000000,
		DynaprepaidActionPlans:  []string{},
	}
	if !reflect.DeepEqual(expAttr, cfg.RalsCfg()) {
		t.Errorf("Expected %s , received: %s ", utils.ToJSON(expAttr), utils.ToJSON(cfg.RalsCfg()))
	}
}

func testCGRConfigReloadSessionS(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	for _, section := range sortedCfgSections {
		cfg.rldChans[section] = make(chan struct{}, 1)
	}
	cfg.RalsCfg().Enabled = true
	cfg.ChargerSCfg().Enabled = true
	cfg.CdrsCfg().Enabled = true
	var reply string
	if err = cfg.V1ReloadConfig(&ReloadArgs{
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
		RouteSConns:   []string{utils.MetaLocalHost},
		AttrSConns:    []string{utils.MetaLocalHost},
		CDRsConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs)},

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

	cfg := NewDefaultCGRConfig()
	for _, section := range sortedCfgSections {
		cfg.rldChans[section] = make(chan struct{}, 1)
	}
	cfg.SessionSCfg().Enabled = true
	var reply string
	if err = cfg.V1ReloadConfig(&ReloadArgs{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "ers_example"),
		Section: ERsJson,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	flags := utils.FlagsWithParamsFromSlice([]string{"*dryrun"})
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
				ID:               utils.MetaDefault,
				Type:             utils.MetaNone,
				RowLength:        0,
				FieldSep:         ",",
				HeaderDefineChar: ":",
				RunDelay:         0,
				ConcurrentReqs:   1024,
				SourcePath:       "/var/spool/cgrates/ers/in",
				ProcessedPath:    "/var/spool/cgrates/ers/out",
				Filters:          []string{},
				Flags:            flagsDefault,
				Fields:           content,
				CacheDumpFields:  []*FCTemplate{},
				XMLRootPath:      utils.HierarchyPath{utils.EmptyString},
				Opts:             make(map[string]interface{}),
			},
			{
				ID:               "file_reader1",
				Type:             utils.MetaFileCSV,
				FieldSep:         ",",
				HeaderDefineChar: ":",
				RunDelay:         -1,
				ConcurrentReqs:   1024,
				SourcePath:       "/tmp/ers/in",
				ProcessedPath:    "/tmp/ers/out",
				Filters:          []string{},
				Flags:            flags,
				Fields:           content,
				CacheDumpFields:  []*FCTemplate{},
				XMLRootPath:      utils.HierarchyPath{utils.EmptyString},
				Opts:             make(map[string]interface{}),
			},
		},
	}
	if !reflect.DeepEqual(expAttr, cfg.ERsCfg()) {
		t.Errorf("Expected %s,\n received: %s ", utils.ToJSON(expAttr), utils.ToJSON(cfg.ERsCfg()))
	}
}

func testCGRConfigReloadDNSAgent(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	for _, section := range sortedCfgSections {
		cfg.rldChans[section] = make(chan struct{}, 1)
	}
	cfg.SessionSCfg().Enabled = true
	var reply string
	if err = cfg.V1ReloadConfig(&ReloadArgs{
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
	cfg := NewDefaultCGRConfig()
	for _, section := range sortedCfgSections {
		cfg.rldChans[section] = make(chan struct{}, 1)
	}
	cfg.SessionSCfg().Enabled = true
	var reply string
	if err = cfg.V1ReloadConfig(&ReloadArgs{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "freeswitch_reload"),
		Section: FreeSWITCHAgentJSN,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	expAttr := &FsAgentCfg{
		Enabled:           true,
		SessionSConns:     []string{utils.ConcatenatedKey(rpcclient.BiRPCInternal, utils.MetaSessionS)},
		SubscribePark:     true,
		ExtraFields:       RSRParsers{},
		MaxWaitConnection: 2 * time.Second,
		EventSocketConns: []*FsConnCfg{
			{
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
			"path":      "*cgreq.ToR",
			"mandatory": true,
			"tag":       "ToR",
			"type":      "*variable",
			"value":     "~*req.2",
		},
		map[string]interface{}{
			"path":      "*cgreq.OriginID",
			"mandatory": true,
			"tag":       "OriginID",
			"type":      "*variable",
			"value":     "~*req.3",
		},
		map[string]interface{}{
			"path":      "*cgreq.RequestType",
			"mandatory": true,
			"tag":       "RequestType",
			"type":      "*variable",
			"value":     "~*req.4",
		},
		map[string]interface{}{
			"path":      "*cgreq.Tenant",
			"mandatory": true,
			"tag":       "Tenant",
			"type":      "*variable",
			"value":     "~*req.6",
		},
		map[string]interface{}{
			"path":      "*cgreq.Category",
			"mandatory": true,
			"tag":       "Category",
			"type":      "*variable",
			"value":     "~*req.7",
		},
		map[string]interface{}{
			"path":      "*cgreq.Account",
			"mandatory": true,
			"tag":       "Account",
			"type":      "*variable",
			"value":     "~*req.8",
		},
		map[string]interface{}{
			"path":      "*cgreq.Subject",
			"mandatory": true,
			"tag":       "Subject",
			"type":      "*variable",
			"value":     "~*req.9",
		},
		map[string]interface{}{
			"path":      "*cgreq.Destination",
			"mandatory": true,
			"tag":       "Destination",
			"type":      "*variable",
			"value":     "~*req.10",
		},
		map[string]interface{}{
			"path":      "*cgreq.SetupTime",
			"mandatory": true,
			"tag":       "SetupTime",
			"type":      "*variable",
			"value":     "~*req.11",
		},
		map[string]interface{}{
			"path":      "*cgreq.AnswerTime",
			"mandatory": true,
			"tag":       "AnswerTime",
			"type":      "*variable",
			"value":     "~*req.12",
		},
		map[string]interface{}{
			"path":      "*cgreq.Usage",
			"mandatory": true,
			"tag":       "Usage",
			"type":      "*variable",
			"value":     "~*req.13",
		},
	}
	expected := map[string]interface{}{
		"enabled": true,
		"readers": []interface{}{
			map[string]interface{}{
				"partial_cache_expiry_action": "",
				"partial_record_cache":        "0",
				"cache_dump_fields":           []interface{}{},
				"concurrent_requests":         1024,
				"fields":                      content,
				"field_separator":             ",",
				"header_define_character":     ":",
				"filters":                     []string{},
				"flags":                       []string{},
				"failed_calls_prefix":         "",
				"id":                          "*default",
				"processed_path":              "/var/spool/cgrates/ers/out",
				"row_length":                  0,
				"run_delay":                   "0",
				"source_path":                 "/var/spool/cgrates/ers/in",
				"tenant":                      "",
				"timezone":                    "",
				"type":                        utils.MetaNone,
				"xml_root_path":               []interface{}{utils.EmptyString},
				"opts":                        make(map[string]interface{}),
			},
			map[string]interface{}{
				"cache_dump_fields":           []interface{}{},
				"concurrent_requests":         1024,
				"field_separator":             ",",
				"header_define_character":     ":",
				"filters":                     []string{},
				"flags":                       []string{"*dryrun"},
				"failed_calls_prefix":         "",
				"partial_cache_expiry_action": "",
				"partial_record_cache":        "0",
				"id":                          "file_reader1",
				"processed_path":              "/tmp/ers/out",
				"row_length":                  0,
				"run_delay":                   "-1",
				"source_path":                 "/tmp/ers/in",
				"tenant":                      "",
				"timezone":                    "",
				"type":                        "*file_csv",
				"xml_root_path":               []interface{}{utils.EmptyString},
				"fields":                      content,
				"opts":                        make(map[string]interface{}),
			},
		},
		"sessions_conns": []string{
			utils.MetaLocalHost,
		},
	}

	cfg := NewDefaultCGRConfig()
	for _, section := range sortedCfgSections {
		cfg.rldChans[section] = make(chan struct{}, 1)
	}
	var reply string
	var rcv map[string]interface{}

	if err := cfg.V1ReloadConfig(&ReloadArgs{
		Path:    "/usr/share/cgrates/conf/samples/ers_example",
		Section: ERsJson,
	}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("Expected: %s \n,received: %s", utils.OK, reply)
	}

	expected = map[string]interface{}{
		ERsJson: expected,
	}
	if err := cfg.V1GetConfig(&SectionWithAPIOpts{Section: ERsJson}, &rcv); err != nil {
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
	for _, section := range sortedCfgSections {
		cfg.rldChans[section] = make(chan struct{}, 1)
	}
	cfg.RalsCfg().Enabled = true
	cfg.ChargerSCfg().Enabled = true
	cfg.CdrsCfg().Enabled = true
	var reply string
	if err = cfg.V1SetConfig(&SetConfigArgs{
		Config: map[string]interface{}{
			"sessions": map[string]interface{}{
				"enabled":          true,
				"resources_conns":  []string{"*localhost"},
				"routes_conns":     []string{"*localhost"},
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
		RouteSConns:   []string{utils.MetaLocalHost},
		AttrSConns:    []string{utils.MetaLocalHost},
		CDRsConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs)},

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
	}
	if !reflect.DeepEqual(expAttr, cfg.SessionSCfg()) {
		t.Errorf("Expected %s , received: %s ", utils.ToJSON(expAttr), utils.ToJSON(cfg.SessionSCfg()))
	}
}

func testCGRConfigReloadConfigFromStringSessionS(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	for _, section := range sortedCfgSections {
		cfg.rldChans[section] = make(chan struct{}, 1)
	}
	cfg.RalsCfg().Enabled = true
	cfg.ChargerSCfg().Enabled = true
	cfg.CdrsCfg().Enabled = true
	var reply string
	if err = cfg.V1SetConfigFromJSON(&SetConfigFromJSONArgs{
		Config: `{"sessions":{
				"enabled":          true,
				"resources_conns":  ["*localhost"],
				"routes_conns":     ["*localhost"],
				"attributes_conns": ["*localhost"],
				"rals_conns":       ["*internal"],
				"cdrs_conns":       ["*internal"],
				"chargers_conns":   ["*localhost"]
				}}`}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	expAttr := &SessionSCfg{
		Enabled:       true,
		ListenBijson:  "127.0.0.1:2014",
		ChargerSConns: []string{utils.MetaLocalHost},
		RALsConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResponder)},
		ResSConns:     []string{utils.MetaLocalHost},
		ThreshSConns:  []string{},
		StatSConns:    []string{},
		RouteSConns:   []string{utils.MetaLocalHost},
		AttrSConns:    []string{utils.MetaLocalHost},
		CDRsConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs)},

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
	}
	if !reflect.DeepEqual(expAttr, cfg.SessionSCfg()) {
		t.Errorf("Expected %s , received: %s ", utils.ToJSON(expAttr), utils.ToJSON(cfg.SessionSCfg()))
	}

	var rcv string
	expected := `{"sessions":{"alterable_fields":[],"attributes_conns":["*localhost"],"cdrs_conns":["*internal"],"channel_sync_interval":"0","chargers_conns":["*localhost"],"client_protocol":1,"debit_interval":"0","default_usage":{"*any":"3h0m0s","*data":"1048576","*sms":"1","*voice":"3h0m0s"},"enabled":true,"listen_bigob":"","listen_bijson":"127.0.0.1:2014","min_dur_low_balance":"0","rals_conns":["*internal"],"replication_conns":[],"resources_conns":["*localhost"],"routes_conns":["*localhost"],"scheduler_conns":[],"session_indexes":[],"session_ttl":"0","stats_conns":[],"stir":{"allowed_attest":["*any"],"default_attest":"A","payload_maxduration":"-1","privatekey_path":"","publickey_path":""},"store_session_costs":false,"terminate_attempts":5,"thresholds_conns":[]}}`
	if err := cfg.V1GetConfigAsJSON(&SectionWithAPIOpts{Section: SessionSJson}, &rcv); err != nil {
		t.Error(err)
	} else if expected != rcv {
		t.Errorf("Expected: %+q, \n received: %s", expected, rcv)
	}
}

func testCGRConfigReloadAll(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	for _, section := range sortedCfgSections {
		cfg.rldChans[section] = make(chan struct{}, 1)
	}
	cfg.RalsCfg().Enabled = true
	cfg.ChargerSCfg().Enabled = true
	cfg.CdrsCfg().Enabled = true
	var reply string
	if err = cfg.V1ReloadConfig(&ReloadArgs{
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
		RouteSConns:   []string{utils.MetaLocalHost},
		AttrSConns:    []string{utils.MetaLocalHost},
		CDRsConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs)},

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
	}
	if !reflect.DeepEqual(expAttr, cfg.SessionSCfg()) {
		t.Errorf("Expected %s , received: %s ", utils.ToJSON(expAttr), utils.ToJSON(cfg.SessionSCfg()))
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
	handleConfigSFolder(flPath, w)

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
	cfg, err := NewCGRConfigFromPath("/usr/share/cgrates/conf/samples/diamagent_internal/")
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
	if err = cfg.loadConfigFromFolder("/usr/share/cgrates/conf/samples/",
		[]func(jsonCfg *CgrJsonCfg) error{cfg.loadFromJSONCfg},
		false); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func testLoadConfigFromFolderOpenError(t *testing.T) {
	newDir := "/tmp/testLoadConfigFromFolderOpenError"
	if err = os.MkdirAll(newDir, 0755); err != nil {
		t.Fatal(err)
	}
	cfg := NewDefaultCGRConfig()

	expected := "open /tmp/testLoadConfigFromFolderOpenError/notes.json: no such file or directory"
	if err = cfg.loadConfigFromFile(path.Join(newDir, "notes.json"),
		[]func(jsonCfg *CgrJsonCfg) error{cfg.loadFromJSONCfg},
		false); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	if err = os.RemoveAll(newDir); err != nil {
		t.Fatal(err)
	}
}

func testLoadConfigFromFolderNoConfigFound(t *testing.T) {
	newDir := "/tmp/[]"
	if err = os.MkdirAll(newDir, 0755); err != nil {
		t.Fatal(err)
	}
	cfg := NewDefaultCGRConfig()

	if err = cfg.loadConfigFromFolder(newDir,
		[]func(jsonCfg *CgrJsonCfg) error{cfg.loadFromJSONCfg},
		false); err == nil || err != filepath.ErrBadPattern {
		t.Errorf("Expected %+v, received %+v", filepath.ErrBadPattern, err)
	}
	if err = os.RemoveAll(newDir); err != nil {
		t.Fatal(err)
	}
}

func testLoadConfigFromPathInvalidArgument(t *testing.T) {
	cfg := NewDefaultCGRConfig()

	expected := "stat /\x00: invalid argument"
	if err = cfg.loadConfigFromPath("/\x00",
		[]func(jsonCfg *CgrJsonCfg) error{cfg.loadFromJSONCfg},
		false); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func testLoadConfigFromPathValidPath(t *testing.T) {
	newDir := "/tmp/testLoadConfigFromPathValidPath"
	if err = os.MkdirAll(newDir, 0755); err != nil {
		t.Fatal(err)
	}
	cfg := NewDefaultCGRConfig()

	expected := "No config file found on path /tmp/testLoadConfigFromPathValidPath"
	if err = cfg.loadConfigFromPath(newDir,
		[]func(jsonCfg *CgrJsonCfg) error{cfg.loadFromJSONCfg},
		false); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	if err = os.RemoveAll(newDir); err != nil {
		t.Fatal(err)
	}
}

func testLoadConfigFromPathFile(t *testing.T) {
	newDir := "/tmp/testLoadConfigFromPathFile"
	if err = os.MkdirAll(newDir, 0755); err != nil {
		t.Fatal(err)
	}
	cfg := NewDefaultCGRConfig()

	expected := "No config file found on path /tmp/testLoadConfigFromPathFile"
	if err = cfg.loadConfigFromPath(newDir,
		[]func(jsonCfg *CgrJsonCfg) error{cfg.loadFromJSONCfg},
		false); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	if err = os.RemoveAll(newDir); err != nil {
		t.Fatal(err)
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

	if err = os.Remove(path.Join(flPath, "random.json")); err != nil {
		t.Fatal(err)
	}

	if err = os.RemoveAll(flPath); err != nil {
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
	handleConfigSFolder(flPath, w)

	if err = os.Remove(path.Join(flPath, "random.json")); err != nil {
		t.Fatal(err)
	}

	if err = os.RemoveAll(flPath); err != nil {
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
