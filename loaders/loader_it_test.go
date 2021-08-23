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
	"io/ioutil"
	"net/rpc"
	"os"
	"path"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	loaderCfgPath               string
	loaderCfgDIR                string //run tests for specific configuration
	loaderCfg                   *config.CGRConfig
	loaderRPC                   *rpc.Client
	loaderPathIn, loaderPathOut string

	sTestsLoader = []func(t *testing.T){
		testLoaderMakeFolders,
		testLoaderInitCfg,
		testLoaderResetDataDB,
		testLoaderStartEngine,
		testLoaderRPCConn,
		testLoaderPopulateData,
		testLoaderLoadAttributes,
		testLoaderVerifyOutDir,
		testLoaderCheckAttributes,
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

	loaderCfgDIR = "tutmysql"
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
	loaderCfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(loaderCfg)
}

func testLoaderMakeFolders(t *testing.T) {
	// active the loaders here
	for _, dir := range []string{"/tmp/In", "/tmp/Out"} {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal("Error removing folder: ", dir, err)
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal("Error creating folder: ", dir, err)
		}
	}
	loaderPathIn = "/tmp/In"
	loaderPathOut = "/tmp/Out"
}

// Wipe out the cdr database
func testLoaderResetDataDB(t *testing.T) {
	if err := engine.InitDataDb(loaderCfg); err != nil {
		t.Fatal(err)
	}
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
	if err := os.Rename(tmpFilePath, path.Join(loaderPathIn, fileName)); err != nil {
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
	if outContent1, err := ioutil.ReadFile(path.Join(loaderPathOut, utils.AttributesCsv)); err != nil {
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
			&engine.Attribute{
				FilterIDs: []string{"*string:~*req.Field1:Initial"},
				Path:      utils.MetaReq + utils.NestingSep + "Field1",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("Sub1", true, utils.INFIELD_SEP),
			},
			&engine.Attribute{
				FilterIDs: []string{},
				Path:      utils.MetaReq + utils.NestingSep + "Field2",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("Sub2", true, utils.INFIELD_SEP),
			}},
		Blocker: true,
		Weight:  20,
	}
	if *encoding == utils.MetaGOB { // gob threats empty slices as nil values
		eAttrPrf.Attributes[1].FilterIDs = nil
	}
	var reply *engine.AttributeProfile
	if err := loaderRPC.Call(utils.APIerSv1GetAttributeProfile,
		&utils.TenantIDWithArgDispatcher{
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

func testLoaderKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
