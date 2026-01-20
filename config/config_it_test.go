//go:build integration
// +build integration

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/
package config

import (
	"fmt"
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
		testCGRConfigReloadChargerSDryRun,
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

	err := expVal.loadConfigFromPath(path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo"),
		[]func(*CgrJsonCfg) error{expVal.loadFromJSONCfg}, false)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := net.DialTimeout(utils.TCP, addr, time.Second); err != nil { // check if site is up
		return
	}

	rply := NewDefaultCGRConfig()
	if err := rply.loadConfigFromPath(addr, []func(*CgrJsonCfg) error{rply.loadFromJSONCfg}, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expVal, rply) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(expVal), utils.ToJSON(rply))
	}
}

func testNewCGRConfigFromPath(t *testing.T) {
	for key, val := range map[string]string{"LOGGER": "*syslog", "LOG_LEVEL": "6", "ROUND_DEC": "5",
		"DB_ENCODING": "*msgpack", "TP_EXPORT_DIR": "/var/spool/cgrates/tpe", "DF_TENANT": "cgrates.org",
		"TIMEZONE": "Local"} {
		os.Setenv(key, val)
	}
	addr := "https://raw.githubusercontent.com/cgrates/cgrates/master/data/conf/samples/multifiles/a.json;https://raw.githubusercontent.com/cgrates/cgrates/master/data/conf/samples/multifiles/b/b.json;https://raw.githubusercontent.com/cgrates/cgrates/master/data/conf/samples/multifiles/c.json;https://raw.githubusercontent.com/cgrates/cgrates/master/data/conf/samples/multifiles/d.json"
	expVal, err := NewCGRConfigFromPath(path.Join("/usr", "share", "cgrates", "conf", "samples", "multifiles"))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := net.DialTimeout(utils.TCP, addr, time.Second); err != nil { // check if site is up
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
	if err := cfg.V1ReloadConfig(context.Background(),
		&ReloadArgs{
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
		ExistsIndexedFields: &[]string{},
		IndexedSelects:      true,
		AnyContext:          true,
		Opts: &AttributesOpts{
			ProcessRuns: 1,
			ProfileIDs:  []string{},
		},
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
	if err := cfg.V1ReloadConfig(context.Background(),
		&ReloadArgs{
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
	if err := cfg.V1ReloadConfig(context.Background(),
		&ReloadArgs{
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
		ExistsIndexedFields: &[]string{},
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
	if err := cfg.V1ReloadConfig(context.Background(),
		&ReloadArgs{
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
		ExistsIndexedFields: &[]string{},
		IndexedSelects:      true,
		Opts: &ThresholdsOpts{
			ProfileIDs: []string{},
		},
		EEsConns: []string{},
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
	if err := cfg.V1ReloadConfig(context.Background(),
		&ReloadArgs{
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
		ExistsIndexedFields: &[]string{},
		IndexedSelects:      true,
		ThresholdSConns:     []string{utils.MetaLocalHost},
		Opts: &StatsOpts{
			ProfileIDs: []string{},
		},
		EEsConns: []string{},
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
	if err := cfg.V1ReloadConfig(context.Background(),
		&ReloadArgs{
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
		ExistsIndexedFields: &[]string{},
		IndexedSelects:      true,
		ThresholdSConns:     []string{utils.MetaLocalHost},
		Opts: &ResourcesOpts{
			UsageID: utils.EmptyString,
			Units:   1,
		},
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
	if err := cfg.V1ReloadConfig(context.Background(),
		&ReloadArgs{
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
		ExistsIndexedFields: &[]string{},
		ResourceSConns:      []string{},
		StatSConns:          []string{},
		AttributeSConns:     []string{},
		RALsConns:           []string{},
		IndexedSelects:      true,
		DefaultRatio:        1,
		Opts: &RoutesOpts{
			Context:      utils.MetaRoutes,
			IgnoreErrors: false,
			MaxCost:      utils.EmptyString,
		},
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
	if err := cfg.V1ReloadConfig(context.Background(),
		&ReloadArgs{
			Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo2"),
			Section: "InvalidSection",
		}, &reply); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v. received %+v", expectedErr, err)
	}

	expectedErr = utils.NewErrMandatoryIeMissing("Path").Error()
	if err := cfg.V1ReloadConfig(context.Background(),
		&ReloadArgs{
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
	if err := cfg.V1ReloadConfig(context.Background(),
		&ReloadArgs{
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

func testCGRConfigReloadSchedulerS(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	for _, section := range sortedCfgSections {
		cfg.rldChans[section] = make(chan struct{}, 1)
	}
	var reply string
	if err := cfg.V1ReloadConfig(context.Background(),
		&ReloadArgs{
			Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo2"),
			Section: SCHEDULER_JSN,
		}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	expAttr := &SchedulerCfg{
		Enabled:                true,
		CDRsConns:              []string{utils.MetaLocalHost},
		ThreshSConns:           []string{},
		StatSConns:             []string{},
		Filters:                []string{},
		DynaprepaidActionPlans: []string{},
	}
	if !reflect.DeepEqual(expAttr, cfg.SchedulerCfg()) {
		t.Errorf("Expected %s , received: %s ", utils.ToJSON(expAttr), utils.ToJSON(cfg.SchedulerCfg()))
	}
}

func testCGRConfigReloadCDRs(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	for _, section := range sortedCfgSections {
		cfg.rldChans[section] = make(chan struct{}, 1)
	}
	cfg.RalsCfg().Enabled = true
	var reply string
	if err := cfg.V1ReloadConfig(context.Background(),
		&ReloadArgs{
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
		SchedulerConns:  []string{},
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
	if err := cfg.V1ReloadConfig(context.Background(),
		&ReloadArgs{
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
		SessionSConns:           []string{},
		MaxIncrements:           1000000,
		FallbackDepth:           3,
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
	if err := cfg.V1ReloadConfig(context.Background(),
		&ReloadArgs{
			Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo2"),
			Section: SessionSJson,
		}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	expAttr := &SessionSCfg{
		Enabled:         true,
		ChargerSConns:   []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)},
		RALsConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResponder)},
		IPsConns:        []string{},
		ResourceSConns:  []string{utils.MetaLocalHost},
		ThresholdSConns: []string{},
		StatSConns:      []string{},
		RouteSConns:     []string{utils.MetaLocalHost},
		AttributeSConns: []string{utils.MetaLocalHost},
		CDRsConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs)},

		ReplicationConns:  []string{},
		SessionIndexes:    utils.StringSet{},
		ClientProtocol:    2,
		TerminateAttempts: 5,
		AlterableFields:   utils.NewStringSet([]string{}),
		STIRCfg: &STIRcfg{
			AllowedAttest:      utils.NewStringSet([]string{utils.MetaAny}),
			PayloadMaxduration: -1,
			DefaultAttest:      "A",
		},
		SchedulerConns: []string{},
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
	if err := cfg.V1ReloadConfig(context.Background(),
		&ReloadArgs{
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
		Enabled:          true,
		SessionSConns:    []string{utils.MetaLocalHost},
		EEsConns:         []string{},
		StatSConns:       []string{},
		ThresholdSConns:  []string{},
		ConcurrentEvents: 1,
		Readers: []*EventReaderCfg{
			{
				ID:                   utils.MetaDefault,
				Type:                 utils.MetaNone,
				RunDelay:             0,
				ConcurrentReqs:       1024,
				SourcePath:           "/var/spool/cgrates/ers/in",
				ProcessedPath:        "/var/spool/cgrates/ers/out",
				Filters:              []string{},
				Flags:                flagsDefault,
				Fields:               content,
				CacheDumpFields:      []*FCTemplate{},
				PartialCommitFields:  []*FCTemplate{},
				Reconnects:           -1,
				MaxReconnectInterval: 5 * time.Minute,
				EEsIDs:               []string{},
				EEsSuccessIDs:        []string{},
				EEsFailedIDs:         []string{},
				Opts: &EventReaderOpts{
					CSV: &CSVROpts{
						FieldSeparator:   utils.StringPointer(utils.FieldsSep),
						HeaderDefineChar: utils.StringPointer(utils.InInFieldSep),
						RowLength:        utils.IntPointer(0),
					},
					AMQP:               &AMQPROpts{},
					AWS:                &AWSROpts{},
					SQL:                &SQLROpts{},
					Kafka:              &KafkaROpts{},
					PartialOrderField:  utils.StringPointer("~*req.AnswerTime"),
					PartialCacheAction: utils.StringPointer(utils.MetaNone),
					NATS: &NATSROpts{
						Subject: utils.StringPointer("cgrates_cdrs"),
					},
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
				EEsIDs:               []string{},
				EEsSuccessIDs:        []string{},
				EEsFailedIDs:         []string{},
				Opts: &EventReaderOpts{
					CSV: &CSVROpts{
						FieldSeparator:   utils.StringPointer(utils.FieldsSep),
						HeaderDefineChar: utils.StringPointer(utils.InInFieldSep),
						RowLength:        utils.IntPointer(0),
					},
					AMQP:               &AMQPROpts{},
					AWS:                &AWSROpts{},
					Kafka:              &KafkaROpts{},
					SQL:                &SQLROpts{},
					PartialOrderField:  utils.StringPointer("~*req.AnswerTime"),
					PartialCacheAction: utils.StringPointer(utils.MetaNone),
					NATS: &NATSROpts{
						Subject: utils.StringPointer("cgrates_cdrs"),
					},
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
	for _, section := range sortedCfgSections {
		cfg.rldChans[section] = make(chan struct{}, 1)
	}
	cfg.SessionSCfg().Enabled = true
	var reply string
	if err := cfg.V1ReloadConfig(context.Background(),
		&ReloadArgs{
			Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "dnsagent_reload"),
			Section: DNSAgentJson,
		}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	expAttr := &DNSAgentCfg{
		Enabled: true,
		Listeners: []DnsListener{
			{
				Address: ":2053",
				Network: "udp",
			},
			{
				Address: ":2054",
				Network: "tcp",
			},
		},
		SessionSConns:   []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
		StatSConns:      []string{},
		ThresholdSConns: []string{},
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
	if err := cfg.V1ReloadConfig(context.Background(),
		&ReloadArgs{
			Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "freeswitch_reload"),
			Section: FreeSWITCHAgentJSN,
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
		SchedTransferExtension: "CGRateS",
		EventSocketConns: []*FsConnCfg{
			{
				Address:      "1.2.3.4:8021",
				Password:     "ClueCon",
				Reconnects:   5,
				ReplyTimeout: time.Minute,
				Alias:        "1.2.3.4:8021",
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
		"concurrent_events": 1,
		"readers": []any{
			map[string]any{
				"id":                          utils.MetaDefault,
				"cache_dump_fields":           []any{},
				"concurrent_requests":         1024,
				"fields":                      content,
				"filters":                     []string{},
				"flags":                       []string{},
				"run_delay":                   "0",
				"start_delay":                 "0",
				"source_path":                 "/var/spool/cgrates/ers/in",
				"processed_path":              "/var/spool/cgrates/ers/out",
				"tenant":                      "",
				"timezone":                    "",
				"type":                        utils.MetaNone,
				utils.ReconnectsCfg:           -1,
				utils.MaxReconnectIntervalCfg: "5m0s",
				"opts": map[string]any{
					"csvFieldSeparator":   ",",
					"csvHeaderDefineChar": ":",
					"csvRowLength":        0.,
					"partialOrderField":   "~*req.AnswerTime",
					"partialCacheAction":  utils.MetaNone,
					"natsSubject":         "cgrates_cdrs",
				},
				"partial_commit_fields": []any{},
			},
			map[string]any{
				"cache_dump_fields":           []any{},
				"concurrent_requests":         1024,
				"filters":                     []string{},
				"flags":                       []string{"*dryrun"},
				"id":                          "file_reader1",
				"processed_path":              "/tmp/ers/out",
				"run_delay":                   "-1",
				"start_delay":                 "0",
				"source_path":                 "/tmp/ers/in",
				"tenant":                      "",
				"timezone":                    "",
				"type":                        "*file_csv",
				"fields":                      content,
				utils.ReconnectsCfg:           -1,
				utils.MaxReconnectIntervalCfg: "5m0s",
				"opts": map[string]any{
					"csvFieldSeparator":   ",",
					"csvHeaderDefineChar": ":",
					"csvRowLength":        0.,
					"partialOrderField":   "~*req.AnswerTime",
					"partialCacheAction":  utils.MetaNone,
					"natsSubject":         "cgrates_cdrs",
				},
				"partial_commit_fields": []any{},
			},
		},
		"sessions_conns": []string{
			utils.MetaLocalHost,
		},
		utils.EEsConnsCfg:        []string{},
		utils.StatSConnsCfg:      []string{},
		utils.ThresholdSConnsCfg: []string{},
	}

	cfg := NewDefaultCGRConfig()
	for _, section := range sortedCfgSections {
		cfg.rldChans[section] = make(chan struct{}, 1)
	}
	var reply string
	var rcv map[string]any

	if err := cfg.V1ReloadConfig(context.Background(),
		&ReloadArgs{
			Path:    "/usr/share/cgrates/conf/samples/ers_example",
			Section: ERsJson,
		}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("Expected: %s \n,received: %s", utils.OK, reply)
	}

	expected = map[string]any{
		ERsJson: expected,
	}
	if err := cfg.V1GetConfig(context.Background(), &SectionWithAPIOpts{Section: ERsJson}, &rcv); err != nil {
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
	if err := cfg.V1SetConfig(context.Background(),
		&SetConfigArgs{
			Config: map[string]any{
				"sessions": map[string]any{
					"enabled":          true,
					"ips_conns":        []string{"*localhost"},
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
		Enabled:         true,
		ChargerSConns:   []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)},
		RALsConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResponder)},
		IPsConns:        []string{utils.MetaLocalHost},
		ResourceSConns:  []string{utils.MetaLocalHost},
		ThresholdSConns: []string{},
		StatSConns:      []string{},
		RouteSConns:     []string{utils.MetaLocalHost},
		AttributeSConns: []string{utils.MetaLocalHost},
		CDRsConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs)},

		ReplicationConns:  []string{},
		SessionIndexes:    utils.StringSet{},
		ClientProtocol:    2,
		TerminateAttempts: 5,
		AlterableFields:   utils.NewStringSet([]string{}),
		STIRCfg: &STIRcfg{
			AllowedAttest:      utils.NewStringSet([]string{utils.MetaAny}),
			PayloadMaxduration: -1,
			DefaultAttest:      "A",
		},
		SchedulerConns: []string{},
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
	if err := cfg.V1SetConfigFromJSON(context.Background(),
		&SetConfigFromJSONArgs{
			Config: `{
	"sessions":{
		"enabled":          true,
		"ips_conns":  		["*localhost"],
		"resources_conns":  ["*localhost"],
		"routes_conns":     ["*localhost"],
		"attributes_conns": ["*localhost"],
		"rals_conns":       ["*internal"],
		"cdrs_conns":       ["*internal"],
		"chargers_conns":   ["*localhost"]
	}
}`}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	expAttr := &SessionSCfg{
		Enabled:         true,
		ChargerSConns:   []string{utils.MetaLocalHost},
		RALsConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResponder)},
		IPsConns:        []string{utils.MetaLocalHost},
		ResourceSConns:  []string{utils.MetaLocalHost},
		ThresholdSConns: []string{},
		StatSConns:      []string{},
		RouteSConns:     []string{utils.MetaLocalHost},
		AttributeSConns: []string{utils.MetaLocalHost},
		CDRsConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs)},

		ReplicationConns:  []string{},
		SessionIndexes:    utils.StringSet{},
		ClientProtocol:    2,
		TerminateAttempts: 5,
		AlterableFields:   utils.NewStringSet([]string{}),
		STIRCfg: &STIRcfg{
			AllowedAttest:      utils.NewStringSet([]string{utils.MetaAny}),
			PayloadMaxduration: -1,
			DefaultAttest:      "A",
		},
		SchedulerConns: []string{},
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
	expected := `{"sessions":{"alterable_fields":[],"attributes_conns":["*localhost"],"backup_interval":"0","cdrs_conns":["*internal"],"channel_sync_interval":"0","chargers_conns":["*localhost"],"client_protocol":2,"debit_interval":"0","default_usage":{"*any":"3h0m0s","*data":"1048576","*sms":"1","*voice":"3h0m0s"},"enabled":true,"ips_conns":["*localhost"],"min_dur_low_balance":"0","rals_conns":["*internal"],"replication_conns":[],"resources_conns":["*localhost"],"routes_conns":["*localhost"],"scheduler_conns":[],"session_indexes":[],"session_ttl":"0","stale_chan_max_extra_usage":"0","stats_conns":[],"stir":{"allowed_attest":["*any"],"default_attest":"A","payload_maxduration":"-1","privatekey_path":"","publickey_path":""},"store_session_costs":false,"terminate_attempts":5,"thresholds_conns":[]}}`

	if err := cfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Section: SessionSJson}, &rcv); err != nil {
		t.Error(err)
	} else if expected != rcv {
		t.Errorf("Expected: %s, \n received: %s", expected, rcv)
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
	if err := cfg.V1ReloadConfig(context.Background(),
		&ReloadArgs{
			Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo2"),
			Section: utils.MetaAll,
		}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	expAttr := &SessionSCfg{
		Enabled:         true,
		ChargerSConns:   []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)},
		RALsConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResponder)},
		IPsConns:        []string{},
		ResourceSConns:  []string{utils.MetaLocalHost},
		ThresholdSConns: []string{},
		StatSConns:      []string{},
		RouteSConns:     []string{utils.MetaLocalHost},
		AttributeSConns: []string{utils.MetaLocalHost},
		CDRsConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs)},

		ReplicationConns:  []string{},
		SessionIndexes:    utils.StringSet{},
		ClientProtocol:    2,
		TerminateAttempts: 5,
		AlterableFields:   utils.NewStringSet([]string{}),
		STIRCfg: &STIRcfg{
			AllowedAttest:      utils.NewStringSet([]string{utils.MetaAny}),
			PayloadMaxduration: -1,
			DefaultAttest:      "A",
		},
		SchedulerConns: []string{},
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
	if err := cfg.loadConfigFromFolder("/usr/share/cgrates/conf/samples/",
		[]func(jsonCfg *CgrJsonCfg) error{cfg.loadFromJSONCfg},
		false); err == nil || err.Error() != expected {
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
	if err := cfg.loadConfigFromFile(path.Join(newDir, "notes.json"),
		[]func(jsonCfg *CgrJsonCfg) error{cfg.loadFromJSONCfg},
		false); err == nil || err.Error() != expected {
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

	if err := cfg.loadConfigFromFolder(newDir,
		[]func(jsonCfg *CgrJsonCfg) error{cfg.loadFromJSONCfg},
		false); err == nil || err != filepath.ErrBadPattern {
		t.Errorf("Expected %+v, received %+v", filepath.ErrBadPattern, err)
	}
	if err := os.RemoveAll(newDir); err != nil {
		t.Fatal(err)
	}
}

func testLoadConfigFromPathInvalidArgument(t *testing.T) {
	cfg := NewDefaultCGRConfig()

	expected := "stat /\x00: invalid argument"
	if err := cfg.loadConfigFromPath("/\x00",
		[]func(jsonCfg *CgrJsonCfg) error{cfg.loadFromJSONCfg},
		false); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func testLoadConfigFromPathValidPath(t *testing.T) {
	newDir := "/tmp/testLoadConfigFromPathValidPath"
	if err := os.MkdirAll(newDir, 0755); err != nil {
		t.Fatal(err)
	}
	cfg := NewDefaultCGRConfig()

	expected := "No config file found on path /tmp/testLoadConfigFromPathValidPath"
	if err := cfg.loadConfigFromPath(newDir,
		[]func(jsonCfg *CgrJsonCfg) error{cfg.loadFromJSONCfg},
		false); err == nil || err.Error() != expected {
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

	expected := "No config file found on path /tmp/testLoadConfigFromPathFile"
	if err := cfg.loadConfigFromPath(newDir,
		[]func(jsonCfg *CgrJsonCfg) error{cfg.loadFromJSONCfg},
		false); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	if err := os.RemoveAll(newDir); err != nil {
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
	handleConfigSFolder(flPath, w)

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

func (responseTest) Write(p []byte) (int, error) {
	return 0, fmt.Errorf("Invalid section")
}

func (responseTest) WriteHeader(statusCode int) {}
