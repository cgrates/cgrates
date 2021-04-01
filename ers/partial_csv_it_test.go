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

package ers

import (
	"net/rpc"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"
	"time"

	v1 "github.com/cgrates/cgrates/apier/v1"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	partCfgPath string
	partCfgDIR  string
	partCfg     *config.CGRConfig
	partRPC     *rpc.Client

	partTests = []func(t *testing.T){
		testCreateDirs,
		testPartITInitConfig,
		testPartITInitCdrDb,
		testPartITResetDataDb,
		testPartITStartEngine,
		testPartITRpcConn,
		testPartITLoadTPFromFolder,
		testPartITHandleCdr1File,
		testPartITHandleCdr2File,
		testPartITHandleCdr3File,
		testPartITVerifyFiles,
		testPartITAnalyseCDRs,
		testCleanupFiles,
		testPartITKillEngine,
	}

	partCsvFileContent1 = `4986517174963,004986517174964,DE-National,04.07.2016 18:58:55,04.07.2016 18:58:55,1,65,Peak,0.014560,498651,partial
4986517174964,004986517174963,DE-National,04.07.2016 20:58:55,04.07.2016 20:58:55,0,74,Offpeak,0.003360,498651,complete
`

	partCsvFileContent2 = `4986517174963,004986517174964,DE-National,04.07.2016 19:00:00,04.07.2016 18:58:55,0,15,Offpeak,0.003360,498651,partial`
	partCsvFileContent3 = `4986517174964,004986517174960,DE-National,04.07.2016 19:05:55,04.07.2016 19:05:55,0,23,Offpeak,0.003360,498651,partial`

	eCacheDumpFile1 = `4986517174963_004986517174964_04.07.2016 18:58:55,1467658800,*rated,086517174963,+4986517174964,2016-07-04T18:58:55Z,2016-07-04T18:58:55Z,15s,-1.00000
4986517174963_004986517174964_04.07.2016 18:58:55,1467658735,*rated,086517174963,+4986517174964,2016-07-04T18:58:55Z,2016-07-04T18:58:55Z,1m5s,-1.00000
`
)

func TestPartReadFile(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		partCfgDIR = "ers_internal"
	case utils.MetaMySQL:
		partCfgDIR = "ers_mysql"
	case utils.MetaMongo:
		partCfgDIR = "ers_mongo"
	case utils.MetaPostgres:
		partCfgDIR = "ers_postgres"
	default:
		t.Fatal("Unknown Database type")
	}
	for _, test := range partTests {
		t.Run(partCfgDIR, test)
	}
}

func testPartITInitConfig(t *testing.T) {
	var err error
	partCfgPath = path.Join(*dataDir, "conf", "samples", partCfgDIR)
	if partCfg, err = config.NewCGRConfigFromPath(partCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

// InitDb so we can rely on count
func testPartITInitCdrDb(t *testing.T) {
	if err := engine.InitStorDB(partCfg); err != nil {
		t.Fatal(err)
	}
}

// Remove data in both rating and accounting db
func testPartITResetDataDb(t *testing.T) {
	if err := engine.InitDataDB(partCfg); err != nil {
		t.Fatal(err)
	}
}

func testPartITStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(partCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testPartITRpcConn(t *testing.T) {
	var err error
	partRPC, err = newRPCClient(partCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testPartITLoadTPFromFolder(t *testing.T) {
	//add a default charger
	chargerProfile := &v1.ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "Default",
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{"*none"},
			Weight:       20,
		},
	}
	var result string
	if err := partRPC.Call(utils.APIerSv1SetChargerProfile, chargerProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

// The default scenario, out of ers defined in .cfg file
func testPartITHandleCdr1File(t *testing.T) {
	fileName := "file1.csv"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := os.WriteFile(tmpFilePath, []byte(partCsvFileContent1), 0644); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.Rename(tmpFilePath, path.Join("/tmp/partErs1/in", fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

// The default scenario, out of ers defined in .cfg file
func testPartITHandleCdr2File(t *testing.T) {
	fileName := "file2.csv"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := os.WriteFile(tmpFilePath, []byte(partCsvFileContent2), 0644); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.Rename(tmpFilePath, path.Join("/tmp/partErs1/in", fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

// The default scenario, out of ers defined in .cfg file
func testPartITHandleCdr3File(t *testing.T) {
	fileName := "file3.csv"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := os.WriteFile(tmpFilePath, []byte(partCsvFileContent3), 0644); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.Rename(tmpFilePath, path.Join("/tmp/partErs2/in", fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
	time.Sleep(time.Second)
}

func testPartITVerifyFiles(t *testing.T) {
	filesInDir, _ := os.ReadDir("/tmp/partErs1/out/")
	if len(filesInDir) == 0 {
		t.Errorf("No files found in folder: <%s>", "/tmp/partErs1/out")
	}
	var fileName string
	for _, file := range filesInDir { // First file in directory is the one we need, harder to find it's name out of config
		if strings.HasPrefix(file.Name(), "4986517174963_004986517174964") {
			fileName = file.Name()
			break
		}
	}
	if contentCacheDump, err := os.ReadFile(path.Join("/tmp/partErs1/out", fileName)); err != nil {
		t.Error(err)
	} else if len(eCacheDumpFile1) != len(string(contentCacheDump)) {
		t.Errorf("Expecting: %q, \n received: %q", eCacheDumpFile1, string(contentCacheDump))
	}
}

func testPartITAnalyseCDRs(t *testing.T) {
	var reply []*engine.ExternalCDR
	if err := partRPC.Call(utils.APIerSv2GetCDRs, &utils.RPCCDRsFilter{}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 2 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
	if err := partRPC.Call(utils.APIerSv2GetCDRs, &utils.RPCCDRsFilter{DestinationPrefixes: []string{"+4986517174963"}}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
	if err := partRPC.Call(utils.APIerSv2GetCDRs, &utils.RPCCDRsFilter{DestinationPrefixes: []string{"+4986517174960"}}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
}

func testPartITKillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}

func TestNewPartialCSVFileER(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fltr := &engine.FilterS{}
	result, err := NewPartialCSVFileER(cfg, 0, nil, nil, fltr, nil)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	expected := &PartialCSVFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltr,
		cache:     result.(*PartialCSVFileER).cache,
		rdrDir:    "/var/spool/cgrates/ers/in",
		rdrEvents: nil,
		rdrError:  nil,
		rdrExit:   nil,
		conReqs:   result.(*PartialCSVFileER).conReqs,
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestNewPartialCSVFileERCase2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers[0].SourcePath = "/"
	fltr := &engine.FilterS{}
	result, err := NewPartialCSVFileER(cfg, 0, nil, nil, fltr, nil)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	expected := &PartialCSVFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltr,
		cache:     result.(*PartialCSVFileER).cache,
		rdrDir:    "",
		rdrEvents: nil,
		rdrError:  nil,
		rdrExit:   nil,
		conReqs:   result.(*PartialCSVFileER).conReqs,
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestNewPartialCSVFileERCase3(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers[0].PartialCacheExpiryAction = utils.MetaDumpToFile
	fltr := &engine.FilterS{}
	result, err := NewPartialCSVFileER(cfg, 0, nil, nil, fltr, nil)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	expected := &PartialCSVFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltr,
		cache:     result.(*PartialCSVFileER).cache,
		rdrDir:    "/var/spool/cgrates/ers/in",
		rdrEvents: nil,
		rdrError:  nil,
		rdrExit:   nil,
		conReqs:   result.(*PartialCSVFileER).conReqs,
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestPartialCSVConfig(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers = []*config.EventReaderCfg{
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
			XMLRootPath:      utils.HierarchyPath{utils.EmptyString},
			Tenant:           nil,
			Timezone:         utils.EmptyString,
			Filters:          []string{},
			Flags:            utils.FlagsWithParams{},
			Fields:           nil,
			CacheDumpFields:  nil,
			Opts:             make(map[string]interface{}),
		},
	}
	fltr := &engine.FilterS{}
	testStruct := &PartialCSVFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltr,
		cache:     nil,
		rdrDir:    "/var/spool/cgrates/ers/in",
		rdrEvents: nil,
		rdrError:  nil,
		rdrExit:   nil,
		conReqs:   nil,
	}
	expected := &config.EventReaderCfg{
		ID:               utils.MetaDefault,
		Type:             utils.MetaNone,
		RowLength:        0,
		FieldSep:         ",",
		HeaderDefineChar: ":",
		RunDelay:         0,
		ConcurrentReqs:   1024,
		SourcePath:       "/var/spool/cgrates/ers/in",
		ProcessedPath:    "/var/spool/cgrates/ers/out",
		XMLRootPath:      utils.HierarchyPath{utils.EmptyString},
		Tenant:           nil,
		Timezone:         utils.EmptyString,
		Filters:          []string{},
		Flags:            utils.FlagsWithParams{},
		Fields:           nil,
		CacheDumpFields:  nil,
		Opts:             make(map[string]interface{}),
	}
	result := testStruct.Config()
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestPartialCSVServe1(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers = []*config.EventReaderCfg{
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
			XMLRootPath:      utils.HierarchyPath{utils.EmptyString},
			Tenant:           nil,
			Timezone:         utils.EmptyString,
			Filters:          []string{},
			Flags:            utils.FlagsWithParams{},
			Fields:           nil,
			CacheDumpFields:  nil,
			Opts:             make(map[string]interface{}),
		},
	}
	fltr := &engine.FilterS{}
	testStruct := &PartialCSVFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltr,
		cache:     nil,
		rdrDir:    "/var/spool/cgrates/ers/in",
		rdrEvents: nil,
		rdrError:  nil,
		rdrExit:   nil,
		conReqs:   nil,
	}
	result := testStruct.Serve()
	if result != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, result)
	}
}

func TestPartialCSVServe3(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers = []*config.EventReaderCfg{
		{
			ID:               utils.MetaDefault,
			Type:             utils.MetaNone,
			RowLength:        0,
			FieldSep:         ",",
			HeaderDefineChar: ":",
			RunDelay:         1,
			ConcurrentReqs:   1024,
			SourcePath:       "/var/spool/cgrates/ers/in",
			ProcessedPath:    "/var/spool/cgrates/ers/out",
			XMLRootPath:      utils.HierarchyPath{utils.EmptyString},
			Tenant:           nil,
			Timezone:         utils.EmptyString,
			Filters:          []string{},
			Flags:            utils.FlagsWithParams{},
			Fields:           nil,
			CacheDumpFields:  nil,
			Opts:             make(map[string]interface{}),
		},
	}
	fltr := &engine.FilterS{}
	testStruct := &PartialCSVFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltr,
		cache:     nil,
		rdrDir:    "/var/spool/cgrates/ers/in",
		rdrEvents: nil,
		rdrError:  nil,
		rdrExit:   nil,
		conReqs:   nil,
	}

	err := testStruct.Serve()
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
}

func TestPartialCSVServe4(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers = []*config.EventReaderCfg{
		{
			ID:               utils.MetaDefault,
			Type:             utils.MetaNone,
			RowLength:        0,
			FieldSep:         ",",
			HeaderDefineChar: ":",
			RunDelay:         1,
			ConcurrentReqs:   1024,
			SourcePath:       "/var/spool/cgrates/ers/in",
			ProcessedPath:    "/var/spool/cgrates/ers/out",
			XMLRootPath:      utils.HierarchyPath{utils.EmptyString},
			Tenant:           nil,
			Timezone:         utils.EmptyString,
			Filters:          []string{},
			Flags:            utils.FlagsWithParams{},
			Fields:           nil,
			CacheDumpFields:  nil,
			Opts:             make(map[string]interface{}),
		},
	}
	fltr := &engine.FilterS{}
	testStruct := &PartialCSVFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltr,
		cache:     nil,
		rdrDir:    "/var/spool/cgrates/ers/in",
		rdrEvents: nil,
		rdrError:  nil,
		rdrExit:   make(chan struct{}, 1),
		conReqs:   nil,
	}
	testStruct.rdrExit <- struct{}{}
	err := testStruct.Serve()
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
}

func TestPartialCSVProcessFile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fltr := &engine.FilterS{}
	testStruct := &PartialCSVFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltr,
		cache:     nil,
		rdrDir:    "/var/spool/cgrates/ers/in",
		rdrEvents: nil,
		rdrError:  nil,
		rdrExit:   make(chan struct{}, 1),
		conReqs:   nil,
	}
	err := testStruct.processFile("", "")
	if err == nil || err.Error() != "open : no such file or directory" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "open : no such file or directory", err)
	}
}

func TestPartialCSVProcessFile2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fltr := &engine.FilterS{}
	testStruct := &PartialCSVFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltr,
		cache:     nil,
		rdrDir:    "/var/spool/cgrates/ers/in",
		rdrEvents: nil,
		rdrError:  nil,
		rdrExit:   make(chan struct{}, 1),
		conReqs:   make(chan struct{}, 1),
	}
	testStruct.conReqs <- struct{}{}
	err := testStruct.processFile("", "")
	if err == nil || err.Error() != "open : no such file or directory" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "open : no such file or directory", err)
	}
}

func TestPartialCSVServe2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers = []*config.EventReaderCfg{
		{
			ID:               utils.MetaDefault,
			Type:             utils.MetaNone,
			RowLength:        0,
			FieldSep:         ",",
			HeaderDefineChar: ":",
			RunDelay:         -1,
			ConcurrentReqs:   1024,
			SourcePath:       "/var/spool/cgrates/ers/in",
			ProcessedPath:    "/var/spool/cgrates/ers/out",
			XMLRootPath:      utils.HierarchyPath{utils.EmptyString},
			Tenant:           nil,
			Timezone:         utils.EmptyString,
			Filters:          []string{},
			Flags:            utils.FlagsWithParams{},
			Fields:           nil,
			CacheDumpFields:  nil,
			Opts:             make(map[string]interface{}),
		},
	}
	fltr := &engine.FilterS{}
	testStruct := &PartialCSVFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltr,
		cache:     nil,
		rdrDir:    "/var/spool/cgrates/ers/in",
		rdrEvents: nil,
		rdrError:  nil,
		rdrExit:   nil,
		conReqs:   nil,
	}

	err := testStruct.Serve()
	if err == nil || err.Error() != "no such file or directory" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "no such file or directory", err)
	}
}
