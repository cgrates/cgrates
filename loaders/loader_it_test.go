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
package loaders

import (
	"encoding/csv"
	"io/ioutil"
	"net/rpc"
	"os"
	"path"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	loaderCfgPath    string
	loaderCfgDIR     string //run tests for specific configuration
	loaderCfg        *config.CGRConfig
	loaderRPC        *rpc.Client
	customAttributes = "12012000001\t12018209998\n12012000002\t15512580598\n12012000007\t19085199998\n12012000008\t18622784999\n12012000010\t17329440866\n12012000011\t18623689800\n12012000012\t19082050951\n12012000014\t17329440866\n12012000015\t12018209999\n12012000031\t12018209999\n12012000032\t19082050951\n12012000033\t12018209998\n12012000034\t12018209998\n"

	sTestsLoader = []func(t *testing.T){
		testLoaderMakeFolders,
		testLoaderInitCfg,
		testLoaderResetDataDB,
		testLoaderStartEngine,
		testLoaderRPCConn,
		testLoaderPopulateData,
		testLoadFromFilesCsvActionProfile,
		testLoadFromFilesCsvActionProfileOpenError,
		testLoadFromFilesCsvActionProfileLockFolderError,
		testLoaderLoadAttributes,
		testLoaderVerifyOutDir,
		testLoaderCheckAttributes,
		testLoaderResetDataDB,
		testLoaderPopulateDataWithoutMoving,
		testLoaderLoadAttributesWithoutMoving,
		testLoaderVerifyOutDirWithoutMoving,
		testLoaderCheckAttributes,
		testLoaderResetDataDB,
		testLoaderPopulateDataWithSubpath,
		testLoaderLoadAttributesWithSubpath,
		testLoaderVerifyOutDirWithSubpath,
		testLoaderCheckAttributes,
		testLoaderResetDataDB,
		testLoaderPopulateDataWithSubpathWithMove,
		testLoaderLoadAttributesWithoutSubpathWithMove,
		testLoaderVerifyOutDirWithSubpathWithMove,
		testLoaderCheckAttributes,
		testLoaderResetDataDB,
		testLoaderPopulateDataForTemplateLoader,
		testLoaderLoadAttributesForTemplateLoader,
		testLoaderVerifyOutDirForTemplateLoader,
		testLoaderCheckAttributes,
		testLoaderResetDataDB,
		testLoaderPopulateDataForCustomSep,
		testLoaderCheckForCustomSep,
		testLoaderVerifyOutDirForCustomSep,
		testLoaderKillEngine,
	}
)

//Test start here
func TestLoaderIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		loaderCfgDIR = "tutinternal"
	case utils.MetaMySQL:
		loaderCfgDIR = "tutmysql"
	case utils.MetaMongo:
		loaderCfgDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestsLoader {
		t.Run(loaderCfgDIR, stest)
	}
}

func testLoaderInitCfg(t *testing.T) {
	var err error
	loaderCfgPath = path.Join(*dataDir, "conf", "samples", "loaders", loaderCfgDIR)
	loaderCfg, err = config.NewCGRConfigFromPath(loaderCfgPath)
	if err != nil {
		t.Fatal(err)
	}
}

func testLoaderMakeFolders(t *testing.T) {
	// active the loaders here
	for _, dir := range loaderPaths {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal("Error removing folder: ", dir, err)
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal("Error creating folder: ", dir, err)
		}
	}
}

// Wipe out the cdr database
func testLoaderResetDataDB(t *testing.T) {
	if err := engine.InitDataDb(loaderCfg); err != nil {
		t.Fatal(err)
	}
	engine.Cache.Clear(nil)
}

// Start CGR Engine
func testLoaderStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(loaderCfgPath, 100); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testLoaderRPCConn(t *testing.T) {
	var err error
	loaderRPC, err = newRPCClient(loaderCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testLoaderPopulateData(t *testing.T) {
	fileName := utils.AttributesCsv
	tmpFilePath := path.Join("/tmp", fileName)
	if err := ioutil.WriteFile(tmpFilePath, []byte(engine.AttributesCSVContent), 0777); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.Rename(tmpFilePath, path.Join("/tmp/In", fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

func testLoaderLoadAttributes(t *testing.T) {
	var reply string
	if err := loaderRPC.Call(utils.LoaderSv1Load,
		&ArgsProcessFolder{LoaderID: "CustomLoader"}, &reply); err != nil {
		t.Error(err)
	}
}

func testLoaderVerifyOutDir(t *testing.T) {
	time.Sleep(100 * time.Millisecond)
	if outContent1, err := ioutil.ReadFile(path.Join("/tmp/Out", utils.AttributesCsv)); err != nil {
		t.Error(err)
	} else if engine.AttributesCSVContent != string(outContent1) {
		t.Errorf("Expecting: %q, received: %q", engine.AttributesCSVContent, string(outContent1))
	}
}

func testLoaderCheckAttributes(t *testing.T) {
	eAttrPrf := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		Contexts:  []string{"con1", "con2", "con3"},
		FilterIDs: []string{"*string:~*req.Account:1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC)},
		Attributes: []*engine.Attribute{
			{
				FilterIDs: []string{"*string:~*req.Field1:Initial"},
				Path:      utils.MetaReq + utils.NestingSep + "Field1",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("Sub1", utils.INFIELD_SEP),
			},
			{
				FilterIDs: []string{},
				Path:      utils.MetaReq + utils.NestingSep + "Field2",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("Sub2", utils.INFIELD_SEP),
			}},
		Blocker: true,
		Weight:  20,
	}
	if *encoding == utils.MetaGOB { // gob threats empty slices as nil values
		eAttrPrf.Attributes[1].FilterIDs = nil
	}
	var reply *engine.AttributeProfile
	if err := loaderRPC.Call(utils.APIerSv1GetAttributeProfile,
		&utils.TenantIDWithOpts{
			TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ALS1"},
		}, &reply); err != nil {
		t.Fatal(err)
	}
	eAttrPrf.Compile()
	reply.Compile()
	sort.Strings(eAttrPrf.Contexts)
	sort.Strings(reply.Contexts)
	if !reflect.DeepEqual(eAttrPrf, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(eAttrPrf), utils.ToJSON(reply))
	}
}

func testLoaderPopulateDataWithoutMoving(t *testing.T) {
	fileName := utils.AttributesCsv
	tmpFilePath := path.Join("/tmp/", fileName)
	if err := ioutil.WriteFile(tmpFilePath, []byte(engine.AttributesCSVContent), 0777); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.Rename(tmpFilePath, path.Join("/tmp/LoaderIn", fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

func testLoaderLoadAttributesWithoutMoving(t *testing.T) {
	var reply string
	if err := loaderRPC.Call(utils.LoaderSv1Load,
		&ArgsProcessFolder{LoaderID: "WithoutMoveToOut"}, &reply); err != nil {
		t.Error(err)
	}
}

func testLoaderVerifyOutDirWithoutMoving(t *testing.T) {
	time.Sleep(100 * time.Millisecond)
	// we expect that after the LoaderS process the file leave in in the input folder
	if outContent1, err := ioutil.ReadFile(path.Join("/tmp/LoaderIn", utils.AttributesCsv)); err != nil {
		t.Error(err)
	} else if engine.AttributesCSVContent != string(outContent1) {
		t.Errorf("Expecting: %q, received: %q", engine.AttributesCSVContent, string(outContent1))
	}
}

func testLoaderPopulateDataWithSubpath(t *testing.T) {
	fileName := utils.AttributesCsv
	tmpFilePath := path.Join("/tmp/", fileName)
	if err := ioutil.WriteFile(tmpFilePath, []byte(engine.AttributesCSVContent), 0777); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.MkdirAll("/tmp/SubpathWithoutMove/folder1", 0755); err != nil {
		t.Fatal("Error creating folder: /tmp/SubpathWithoutMove/folder1", err)
	}
	if err := os.Rename(tmpFilePath, path.Join("/tmp/SubpathWithoutMove/folder1", fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

func testLoaderLoadAttributesWithSubpath(t *testing.T) {
	var reply string
	if err := loaderRPC.Call(utils.LoaderSv1Load,
		&ArgsProcessFolder{LoaderID: "SubpathLoaderWithoutMove"}, &reply); err != nil {
		t.Error(err)
	}
}

func testLoaderVerifyOutDirWithSubpath(t *testing.T) {
	time.Sleep(100 * time.Millisecond)
	// we expect that after the LoaderS process the file leave in in the input folder
	if outContent1, err := ioutil.ReadFile(path.Join("/tmp/SubpathWithoutMove/folder1", utils.AttributesCsv)); err != nil {
		t.Error(err)
	} else if engine.AttributesCSVContent != string(outContent1) {
		t.Errorf("Expecting: %q, received: %q", engine.AttributesCSVContent, string(outContent1))
	}
}

func testLoaderPopulateDataWithSubpathWithMove(t *testing.T) {
	fileName := utils.AttributesCsv
	tmpFilePath := path.Join("/tmp/", fileName)
	if err := ioutil.WriteFile(tmpFilePath, []byte(engine.AttributesCSVContent), 0777); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.MkdirAll("/tmp/SubpathLoaderWithMove/folder1", 0755); err != nil {
		t.Fatal("Error creating folder: /tmp/SubpathLoaderWithMove/folder1", err)
	}
	if err := os.Rename(tmpFilePath, path.Join("/tmp/SubpathLoaderWithMove/folder1", fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

func testLoaderLoadAttributesWithoutSubpathWithMove(t *testing.T) {
	var reply string
	if err := loaderRPC.Call(utils.LoaderSv1Load,
		&ArgsProcessFolder{LoaderID: "SubpathLoaderWithMove"}, &reply); err != nil {
		t.Error(err)
	}
}

func testLoaderVerifyOutDirWithSubpathWithMove(t *testing.T) {
	time.Sleep(100 * time.Millisecond)
	if outContent1, err := ioutil.ReadFile(path.Join("/tmp/SubpathOut/folder1", utils.AttributesCsv)); err != nil {
		t.Error(err)
	} else if engine.AttributesCSVContent != string(outContent1) {
		t.Errorf("Expecting: %q, received: %q", engine.AttributesCSVContent, string(outContent1))
	}
}

func testLoaderPopulateDataForTemplateLoader(t *testing.T) {
	fileName := utils.AttributesCsv
	tmpFilePath := path.Join("/tmp/", fileName)
	if err := ioutil.WriteFile(tmpFilePath, []byte(engine.AttributesCSVContent), 0777); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.MkdirAll("/tmp/templateLoaderIn", 0755); err != nil {
		t.Fatal("Error creating folder: /tmp/templateLoaderIn", err)
	}
	if err := os.Rename(tmpFilePath, path.Join("/tmp/templateLoaderIn", fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

func testLoaderLoadAttributesForTemplateLoader(t *testing.T) {
	var reply string
	if err := loaderRPC.Call(utils.LoaderSv1Load,
		&ArgsProcessFolder{LoaderID: "LoaderWithTemplate"}, &reply); err != nil {
		t.Error(err)
	}
}

func testLoaderVerifyOutDirForTemplateLoader(t *testing.T) {
	time.Sleep(100 * time.Millisecond)
	if outContent1, err := ioutil.ReadFile(path.Join("/tmp/templateLoaderOut", utils.AttributesCsv)); err != nil {
		t.Error(err)
	} else if engine.AttributesCSVContent != string(outContent1) {
		t.Errorf("Expecting: %q, received: %q", engine.AttributesCSVContent, string(outContent1))
	}
}

func testLoaderKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}

func testLoaderPopulateDataForCustomSep(t *testing.T) {
	fileName := utils.Attributes
	tmpFilePath := path.Join("/tmp/", fileName)
	if err := ioutil.WriteFile(tmpFilePath, []byte(customAttributes), 0777); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.MkdirAll("/tmp/customSepLoaderIn", 0755); err != nil {
		t.Fatal("Error creating folder: /tmp/customSepLoaderIn", err)
	}
	if err := os.Rename(tmpFilePath, path.Join("/tmp/customSepLoaderIn", fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testLoaderCheckForCustomSep(t *testing.T) {
	eAttrPrf := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_12012000001",
		Contexts:  []string{"*any"},
		FilterIDs: []string{"*string:~*req.Destination:12012000001"},
		Attributes: []*engine.Attribute{
			{
				FilterIDs: []string{},
				Path:      "*req.Destination",
				Type:      utils.META_CONSTANT,
				Value:     config.NewRSRParsersMustCompile("12018209998", utils.INFIELD_SEP),
			},
		},
	}
	if *encoding == utils.MetaGOB { // gob threats empty slices as nil values
		eAttrPrf.Attributes[0].FilterIDs = nil
	}
	var reply *engine.AttributeProfile
	if err := loaderRPC.Call(utils.APIerSv1GetAttributeProfile,
		&utils.TenantIDWithOpts{
			TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_12012000001"},
		}, &reply); err != nil {
		t.Fatal(err)
	}
	eAttrPrf.Compile()
	reply.Compile()
	if !reflect.DeepEqual(eAttrPrf, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(eAttrPrf), utils.ToJSON(reply))
	}
}

func testLoaderVerifyOutDirForCustomSep(t *testing.T) {
	time.Sleep(100 * time.Millisecond)
	if outContent1, err := ioutil.ReadFile(path.Join("/tmp/customSepLoaderOut", utils.Attributes)); err != nil {
		t.Error(err)
	} else if customAttributes != string(outContent1) {
		t.Errorf("Expecting: %q, received: %q", customAttributes, string(outContent1))
	}
}

func testLoadFromFilesCsvActionProfile(t *testing.T) {
	flPath := "/tmp/TestLoadFromFilesCsvActionProfile"
	if err := os.MkdirAll(flPath, 0777); err != nil {
		t.Error(err)
	}
	newFile, err := os.Create(path.Join(flPath, "ActionProfiles.csv"))
	if err != nil {
		t.Error(err)
	}
	newFile.Write([]byte(`
#Tenant[0],ID[1]
cgrates.org,SET_ACTPROFILE_3
`))
	content, err := ioutil.ReadFile(path.Join(flPath, "ActionProfiles.csv"))
	if err != nil {
		t.Error(err)
	}
	newFile.Close()

	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestRemoveActionProfileContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		tpInDir:       flPath,
		tpOutDir:      utils.EmptyString,
		lockFilename:  "ActionProfiles.csv",
		fieldSep:      ",",
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaActionProfiles: {
			{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.INFIELD_SEP),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.INFIELD_SEP),
				Mandatory: true},
		},
	}

	rdr := ioutil.NopCloser(strings.NewReader(string(content)))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaActionProfiles: {
			utils.ActionProfilesCsv: &openedCSVFile{
				fileName: utils.ActionProfilesCsv,
				rdr:      rdr,
				csvRdr:   csvRdr,
			},
		},
	}
	if err := ldr.ProcessFolder(utils.EmptyString, utils.MetaStore, true); err != nil {
		t.Error(err)
	}
	expACtPrf := &engine.ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "SET_ACTPROFILE_3",
		FilterIDs: []string{},
		Targets:   map[string]utils.StringSet{},
		Actions:   []*engine.APAction{},
	}
	if rcv, err := ldr.dm.GetActionProfile(expACtPrf.Tenant, expACtPrf.ID,
		true, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expACtPrf, rcv) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expACtPrf), utils.ToJSON(rcv))
	}

	//checking the error by adding a caching method
	ldr.connMgr = engine.NewConnManager(config.NewDefaultCGRConfig(), nil)
	ldr.cacheConns = []string{utils.MetaInternal}
	rdr = ioutil.NopCloser(strings.NewReader(string(content)))
	csvRdr = csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaActionProfiles: {
			utils.ActionProfilesCsv: &openedCSVFile{
				fileName: utils.ActionProfilesCsv,
				rdr:      rdr,
				csvRdr:   csvRdr,
			},
		},
	}
	expected := "UNSUPPORTED_SERVICE_METHOD"
	if err := ldr.ProcessFolder(utils.MetaReload, utils.MetaStore, true); err == nil || err.Error() != expected {
		t.Error(err)
	}

	if err = os.RemoveAll(flPath); err != nil {
		t.Fatal(err)
	}
}

func testLoadFromFilesCsvActionProfileOpenError(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestRemoveActionProfileContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		tpInDir:       "/tmp/testLoadFromFilesCsvActionProfileOpenError",
		timezone:      "UTC",
	}
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaActionProfiles: {
			utils.ActionProfilesCsv: &openedCSVFile{
				fileName: utils.ActionProfilesCsv,
			},
		},
	}
	expectedErr := "open /tmp/testLoadFromFilesCsvActionProfileOpenError/ActionProfiles.csv: not a directory"
	if err := ldr.ProcessFolder(utils.EmptyString, utils.MetaStore, true); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}

	//if stopOnError is on true, the error is avoided,but instead will get a logger.warning message
	if err := ldr.ProcessFolder(utils.EmptyString, utils.MetaStore, false); err != nil {
		t.Error(err)
	}
}

func testLoadFromFilesCsvActionProfileLockFolderError(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestRemoveActionProfileContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaActionProfiles: {
			utils.ActionProfilesCsv: &openedCSVFile{
				fileName: utils.ActionProfilesCsv,
			},
		},
	}
	expectedErr := "open : no such file or directory"
	if err := ldr.ProcessFolder(utils.EmptyString, utils.MetaStore, true); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}
