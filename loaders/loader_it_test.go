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
package loaders

import (
	"encoding/csv"
	"errors"
	"io"
	"os"
	"path"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/birpc/jsonrpc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	loaderCfgPath    string
	loaderCfgDIR     string //run tests for specific configuration
	loaderCfg        *config.CGRConfig
	loaderRPC        *birpc.Client
	customAttributes = "12012000001\t12018209998\n12012000002\t15512580598\n12012000007\t19085199998\n12012000008\t18622784999\n12012000010\t17329440866\n12012000011\t18623689800\n12012000012\t19082050951\n12012000014\t17329440866\n12012000015\t12018209999\n12012000031\t12018209999\n12012000032\t19082050951\n12012000033\t12018209998\n12012000034\t12018209998\n"

	sTestsLoader = []func(t *testing.T){
		testLoaderMakeFolders,
		testLoaderInitCfg,
		testLoaderResetDataDB,
		testLoaderStartEngine,
		testLoaderRPCConn,
		testLoaderPopulateData,
		testProcessFile,
		testProcessFileLockFolder,
		testProcessFileUnableToOpen,
		testProcessFileAllFilesPresent,
		testProcessFileRenameError,
		testAllFilesPresentEmptyCSV,
		testIsFolderLocked,
		testNewLockFolder,
		testNewLockFolderNotFound,
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

// Test start here
func TestLoaderIT(t *testing.T) {
	switch *utils.DBType {
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
	loaderCfgPath = path.Join(*utils.DataDir, "conf", "samples", "loaders", loaderCfgDIR)
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
	switch *utils.Encoding {
	case utils.MetaJSON:
		loaderRPC, err = jsonrpc.Dial(utils.TCP, loaderCfg.ListenCfg().RPCJSONListen)
	case utils.MetaGOB:
		loaderRPC, err = birpc.Dial(utils.TCP, loaderCfg.ListenCfg().RPCGOBListen)
	default:
		loaderRPC, err = nil, errors.New("UNSUPPORTED_RPC")
	}
	if err != nil {
		t.Fatal(err)
	}
}

func testLoaderPopulateData(t *testing.T) {
	fileName := utils.AttributesCsv
	tmpFilePath := path.Join("/tmp", fileName)
	if err := os.WriteFile(tmpFilePath, []byte(AttributesCSVContent), os.ModePerm); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.Rename(tmpFilePath, path.Join("/tmp/In", fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

func testLoaderLoadAttributes(t *testing.T) {
	var reply string
	if err := loaderRPC.Call(context.Background(), utils.LoaderSv1Load,
		&ArgsProcessFolder{LoaderID: "CustomLoader"}, &reply); err != nil {
		t.Error(err)
	}
}

func testLoaderVerifyOutDir(t *testing.T) {
	time.Sleep(100 * time.Millisecond)
	if outContent1, err := os.ReadFile(path.Join("/tmp/Out", utils.AttributesCsv)); err != nil {
		t.Error(err)
	} else if AttributesCSVContent != string(outContent1) {
		t.Errorf("Expecting: %q, received: %q", AttributesCSVContent, string(outContent1))
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
				Value:     config.NewRSRParsersMustCompile("Sub1", utils.InfieldSep),
			},
			{
				FilterIDs: []string{},
				Path:      utils.MetaReq + utils.NestingSep + "Field2",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("Sub2", utils.InfieldSep),
			}},
		Blocker: true,
		Weight:  20,
	}
	if *utils.Encoding == utils.MetaGOB { // gob threats empty slices as nil values
		eAttrPrf.Attributes[1].FilterIDs = nil
	}
	var reply *engine.AttributeProfile
	if err := loaderRPC.Call(context.Background(), utils.APIerSv1GetAttributeProfile,
		&utils.TenantIDWithAPIOpts{
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
	if err := os.WriteFile(tmpFilePath, []byte(AttributesCSVContent), os.ModePerm); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.Rename(tmpFilePath, path.Join("/tmp/LoaderIn", fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

func testLoaderLoadAttributesWithoutMoving(t *testing.T) {
	var reply string
	if err := loaderRPC.Call(context.Background(), utils.LoaderSv1Load,
		&ArgsProcessFolder{LoaderID: "WithoutMoveToOut"}, &reply); err != nil {
		t.Error(err)
	}
}

func testLoaderVerifyOutDirWithoutMoving(t *testing.T) {
	time.Sleep(100 * time.Millisecond)
	// we expect that after the LoaderS process the file leave in in the input folder
	if outContent1, err := os.ReadFile(path.Join("/tmp/LoaderIn", utils.AttributesCsv)); err != nil {
		t.Error(err)
	} else if AttributesCSVContent != string(outContent1) {
		t.Errorf("Expecting: %q, received: %q", AttributesCSVContent, string(outContent1))
	}
}

func testLoaderPopulateDataWithSubpath(t *testing.T) {
	fileName := utils.AttributesCsv
	tmpFilePath := path.Join("/tmp/", fileName)
	if err := os.WriteFile(tmpFilePath, []byte(AttributesCSVContent), os.ModePerm); err != nil {
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
	if err := loaderRPC.Call(context.Background(), utils.LoaderSv1Load,
		&ArgsProcessFolder{LoaderID: "SubpathLoaderWithoutMove"}, &reply); err != nil {
		t.Error(err)
	}
}

func testLoaderVerifyOutDirWithSubpath(t *testing.T) {
	time.Sleep(100 * time.Millisecond)
	// we expect that after the LoaderS process the file leave in in the input folder
	if outContent1, err := os.ReadFile(path.Join("/tmp/SubpathWithoutMove/folder1", utils.AttributesCsv)); err != nil {
		t.Error(err)
	} else if AttributesCSVContent != string(outContent1) {
		t.Errorf("Expecting: %q, received: %q", AttributesCSVContent, string(outContent1))
	}
}

func testLoaderPopulateDataWithSubpathWithMove(t *testing.T) {
	fileName := utils.AttributesCsv
	tmpFilePath := path.Join("/tmp/", fileName)
	if err := os.WriteFile(tmpFilePath, []byte(AttributesCSVContent), os.ModePerm); err != nil {
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
	if err := loaderRPC.Call(context.Background(), utils.LoaderSv1Load,
		&ArgsProcessFolder{LoaderID: "SubpathLoaderWithMove"}, &reply); err != nil {
		t.Error(err)
	}
}

func testLoaderVerifyOutDirWithSubpathWithMove(t *testing.T) {
	time.Sleep(100 * time.Millisecond)
	if outContent1, err := os.ReadFile(path.Join("/tmp/SubpathOut/folder1", utils.AttributesCsv)); err != nil {
		t.Error(err)
	} else if AttributesCSVContent != string(outContent1) {
		t.Errorf("Expecting: %q, received: %q", AttributesCSVContent, string(outContent1))
	}
}

func testLoaderPopulateDataForTemplateLoader(t *testing.T) {
	fileName := utils.AttributesCsv
	tmpFilePath := path.Join("/tmp/", fileName)
	if err := os.WriteFile(tmpFilePath, []byte(AttributesCSVContent), os.ModePerm); err != nil {
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
	if err := loaderRPC.Call(context.Background(), utils.LoaderSv1Load,
		&ArgsProcessFolder{LoaderID: "LoaderWithTemplate"}, &reply); err != nil {
		t.Error(err)
	}
}

func testLoaderVerifyOutDirForTemplateLoader(t *testing.T) {
	time.Sleep(100 * time.Millisecond)
	if outContent1, err := os.ReadFile(path.Join("/tmp/templateLoaderOut", utils.AttributesCsv)); err != nil {
		t.Error(err)
	} else if AttributesCSVContent != string(outContent1) {
		t.Errorf("Expecting: %q, received: %q", AttributesCSVContent, string(outContent1))
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
	if err := os.WriteFile(tmpFilePath, []byte(customAttributes), os.ModePerm); err != nil {
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
				Type:      utils.MetaConstant,
				Value:     config.NewRSRParsersMustCompile("12018209998", utils.InfieldSep),
			},
		},
	}
	if *utils.Encoding == utils.MetaGOB { // gob threats empty slices as nil values
		eAttrPrf.Attributes[0].FilterIDs = nil
	}
	var reply *engine.AttributeProfile
	if err := loaderRPC.Call(context.Background(), utils.APIerSv1GetAttributeProfile,
		&utils.TenantIDWithAPIOpts{
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
	if outContent1, err := os.ReadFile(path.Join("/tmp/customSepLoaderOut", utils.Attributes)); err != nil {
		t.Error(err)
	} else if customAttributes != string(outContent1) {
		t.Errorf("Expecting: %q, received: %q", customAttributes, string(outContent1))
	}
}

func testProcessFile(t *testing.T) {
	flPath := "/tmp/testProcessFile"
	if err := os.MkdirAll(flPath, os.ModePerm); err != nil {
		t.Error(err)
	}
	file, err := os.Create(path.Join(flPath, utils.ResourcesCsv))
	if err != nil {
		t.Error(err)
	}
	file.Write([]byte(`
#Tenant[0],ID[1]
cgrates.org,NewRes1
`))
	file.Close()

	data := engine.NewInternalDB(nil, nil, true, loaderCfg.DataDbCfg().Items)
	ldr := &Loader{
		ldrID:         "testProcessFile",
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		fieldSep:      utils.FieldsSep,
		tpInDir:       flPath,
		tpOutDir:      "/tmp",
		lockFilepath:  "/tmp/testProcessFile/.lck",
		bufLoaderData: make(map[string][]LoaderData),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaResources: {
			{Tag: "Tenant",
				Path:      "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Tag: "ID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
		},
	}

	//loader file is empty (loaderType will be empty)
	if err := ldr.processFile(utils.ResourcesCsv); err != nil {
		t.Error(err)
	}

	// 	resCsv := `
	// #Tenant[0],ID[1]
	// cgrates.org,NewRes1
	// `
	// 	rdr := io.NopCloser(strings.NewReader(resCsv))

	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaResources: {
			utils.ResourcesCsv: &openedCSVFile{
				rdr:    io.NopCloser(nil),
				csvRdr: csv.NewReader(nil),
			},
		},
	}

	expRes := &engine.ResourceProfile{
		Tenant:       "cgrates.org",
		ID:           "NewRes1",
		FilterIDs:    []string{},
		ThresholdIDs: []string{},
	}

	//successfully processed the file
	if err := ldr.processFile(utils.ResourcesCsv); err != nil {
		t.Error(err)
	}

	//get ResourceProfile and compare
	if rcv, err := ldr.dm.GetResourceProfile(expRes.Tenant, expRes.ID, true, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, expRes) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expRes), utils.ToJSON(rcv))
	}

	if err := ldr.dm.RemoveResourceProfile(expRes.Tenant, expRes.ID, true); err != nil {
		t.Error(err)
	}

	file, err = os.Create(path.Join(flPath, utils.ResourcesCsv))
	if err != nil {
		t.Error(err)
	}
	file.Write([]byte(`
#Tenant[0],ID[1]
cgrates.org,NewRes1
`))
	file.Close()

	//cannot move file when tpOutDir is empty
	ldr.tpOutDir = utils.EmptyString
	if err := ldr.processFile(utils.ResourcesCsv); err != nil {
		t.Error(err)
	}

	if err := os.Remove(path.Join("/tmp", utils.ResourcesCsv)); err != nil {
		t.Error(err)
	} else if err := os.RemoveAll(flPath); err != nil {
		t.Error(err)
	}
}

func testProcessFileAllFilesPresent(t *testing.T) {
	flPath := "/tmp/testProcessFile"
	if err := os.MkdirAll(flPath, os.ModePerm); err != nil {
		t.Error(err)
	}
	file, err := os.Create(path.Join(flPath, "inexistent.csv"))
	if err != nil {
		t.Error(err)
	}

	file.Write([]byte(`
#Tenant[0],ID[1]
cgrates.org,NewRes1
`))
	file.Close()

	data := engine.NewInternalDB(nil, nil, true, loaderCfg.DataDbCfg().Items)
	ldr := &Loader{
		ldrID:         "testProcessFile",
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		fieldSep:      utils.FieldsSep,
		tpInDir:       flPath,
		tpOutDir:      "/tmp",
		lockFilepath:  utils.ResourcesCsv,
		bufLoaderData: make(map[string][]LoaderData),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaResources: {
			{Tag: "Tenant",
				Path:      "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.FieldsSep),
				Mandatory: true},
			{Tag: "ID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.FieldsSep),
				Mandatory: true},
		},
	}

	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaResources: {
			"inexistent.csv":    nil,
			utils.AttributesCsv: nil,
		},
	}

	if err := ldr.processFile("inexistent.csv"); err != nil {
		t.Error(err)
	}

	if err := os.Remove(path.Join(flPath, "inexistent.csv")); err != nil {
		t.Error(err)
	} else if err := os.Remove(flPath); err != nil {
		t.Error(err)
	}
}

func testProcessFileLockFolder(t *testing.T) {
	flPath := "/tmp/testProcessFileLockFolder"
	if err := os.MkdirAll(flPath, os.ModePerm); err != nil {
		t.Error(err)
	}
	_, err := os.Create(path.Join(flPath, utils.ResourcesCsv))
	if err != nil {
		t.Error(err)
	}

	ldr := &Loader{
		ldrID:        "testProcessFileLockFolder",
		tpInDir:      flPath,
		tpOutDir:     "/tmp",
		lockFilepath: "/tmp/test/.cgr.lck",
		fieldSep:     utils.InfieldSep,
	}

	resCsv := `
#Tenant[0],ID[1]
cgrates.org,NewRes1
`
	rdr := io.NopCloser(strings.NewReader(resCsv))

	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaResources: {
			utils.ResourcesCsv: &openedCSVFile{
				fileName: utils.ResourcesCsv,
				rdr:      rdr,
			},
		},
	}

	//unable to lock the folder, because lockFileName is missing
	expected := "open /tmp/test/.cgr.lck: no such file or directory"
	if err := ldr.processFile(utils.ResourcesCsv); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	if err := os.Remove(path.Join(flPath, utils.ResourcesCsv)); err != nil {
		t.Error(err)
	} else if err := os.Remove(flPath); err != nil {
		t.Error(err)
	}
}

func testProcessFileUnableToOpen(t *testing.T) {
	flPath := "/tmp/testProcessFileUnableToOpen"
	if err := os.MkdirAll(flPath, os.ModePerm); err != nil {
		t.Error(err)
	}

	ldr := &Loader{
		ldrID:        "testProcessFile",
		tpInDir:      flPath,
		fieldSep:     ",",
		lockFilepath: utils.MetaResources,
	}
	resCsv := `
#Tenant[0],ID[1]
cgrates.org,NewRes1
`
	rdr := io.NopCloser(strings.NewReader(resCsv))

	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaResources: {
			`resources`: &openedCSVFile{
				fileName: utils.ResourcesCsv,
				rdr:      rdr,
			},
		},
	}

	//unable to lock the folder, because lockFileName is missing
	expected := "open /tmp/testProcessFileUnableToOpen/resources: no such file or directory"
	if err := ldr.processFile(`resources`); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	if err := os.Remove(flPath); err != nil {
		t.Error(err)
	}
}

func testProcessFileRenameError(t *testing.T) {
	flPath1 := "/tmp/testProcessFileLockFolder"
	if err := os.MkdirAll(flPath1, os.ModePerm); err != nil {
		t.Error(err)
	}
	data := engine.NewInternalDB(nil, nil, true, loaderCfg.DataDbCfg().Items)
	ldr := &Loader{
		ldrID:         "testProcessFileRenameError",
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		fieldSep:      utils.FieldsSep,
		tpInDir:       flPath1,
		tpOutDir:      "INEXISTING_FILE",
		lockFilepath:  utils.ResourcesCsv,
		bufLoaderData: make(map[string][]LoaderData),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaResources: {
			{Tag: "Tenant",
				Path:      "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Tag: "ID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
		},
	}

	// 	resCsv := `
	// #Tenant[0],ID[1]
	// cgrates.org,NewRes1
	// `
	// 	rdr := io.NopCloser(strings.NewReader(resCsv))

	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaResources: {
			utils.ResourcesCsv: &openedCSVFile{
				rdr:    io.NopCloser(nil),
				csvRdr: csv.NewReader(nil),
			},
		},
	}

	file, err := os.Create(path.Join(flPath1, utils.ResourcesCsv))
	if err != nil {
		t.Error(err)
	}
	file.Write([]byte(`
#Tenant[0],ID[1]
cgrates.org,NewRes1
`))
	file.Close()

	expected := "rename /tmp/testProcessFileLockFolder/Resources.csv INEXISTING_FILE/Resources.csv: no such file or directory"
	if err := ldr.processFile(utils.ResourcesCsv); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	if err := os.RemoveAll(flPath1); err != nil {
		t.Error(err)
	}
}

func testAllFilesPresentEmptyCSV(t *testing.T) {
	ldr := &Loader{
		ldrID:         "testProcessFileRenameError",
		lockFilepath:  utils.ResourcesCsv,
		bufLoaderData: make(map[string][]LoaderData),
		timezone:      "UTC",
	}
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaResources: {
			utils.ResourcesCsv: nil,
		},
	}
	if rcv := ldr.allFilesPresent(utils.MetaResources); rcv {
		t.Errorf("Expecting false")
	}
}

func testIsFolderLocked(t *testing.T) {
	flPath := "/tmp/testIsFolderLocked"
	ldr := &Loader{
		ldrID:         "TestLoadAndRemoveResources",
		tpInDir:       flPath,
		lockFilepath:  utils.EmptyString,
		bufLoaderData: make(map[string][]LoaderData),
		timezone:      "UTC",
	}
	expected := "stat /\x00: invalid argument"
	if _, err := ldr.isFolderLocked(); err != nil {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func testNewLockFolder(t *testing.T) {
	pathL := "/tmp/testNewLockFolder/"
	if err := os.MkdirAll(pathL, os.ModePerm); err != nil {
		t.Error(err)
	}

	_, err := os.Create(path.Join(pathL, utils.ResourcesCsv))
	if err != nil {
		t.Error(err)
	}

	ldr := &Loader{
		ldrID:         "testNewLockFolder",
		tpInDir:       "",
		lockFilepath:  pathL + utils.ResourcesCsv,
		bufLoaderData: make(map[string][]LoaderData),
		timezone:      "UTC",
	}

	if err := ldr.lockFolder(); err != nil {
		t.Error(err)
	}
	if err := os.RemoveAll(pathL); err != nil {
		t.Error(err)
	}
}

func testNewLockFolderNotFound(t *testing.T) {
	pathL := "/tmp/testNewLockFolder/"
	ldr := &Loader{
		ldrID:         "testNewLockFolder",
		tpInDir:       "",
		lockFilepath:  pathL + utils.ResourcesCsv,
		bufLoaderData: make(map[string][]LoaderData),
		timezone:      "UTC",
	}

	errExpect := "open /tmp/testNewLockFolder/Resources.csv: no such file or directory"
	if err := ldr.lockFolder(); err == nil || err.Error() != errExpect {
		t.Error(err)
	}
}

func testNewIsFolderLock(t *testing.T) {
	pathL := "/tmp/testNewLockFolder/"
	if err := os.MkdirAll(pathL, os.ModePerm); err != nil {
		t.Error(err)
	}

	_, err := os.Create(path.Join(pathL, utils.ResourcesCsv))
	if err != nil {
		t.Error(err)
	}

	ldr := &Loader{
		ldrID:         "testNewLockFolder",
		tpInDir:       "",
		lockFilepath:  pathL + utils.ResourcesCsv,
		bufLoaderData: make(map[string][]LoaderData),
		timezone:      "UTC",
	}

	if err := ldr.lockFolder(); err != nil {
		t.Error(err)
	}

	isLocked, err := ldr.isFolderLocked()
	if !isLocked {
		t.Error("Expected the file to be locked")
	}

	if err := os.RemoveAll(pathL); err != nil {
		t.Error(err)
	}
}
