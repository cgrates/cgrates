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
	"errors"
	"flag"
	"os"
	"path"
	"reflect"
	"sort"
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

		testLoaderResetDataDB,
		populateData("/tmp/In"),
		runLoader("CustomLoader"),
		verifyOutput("/tmp/Out"),
		testLoaderCheckAttributes,

		testLoaderResetDataDB,
		populateData("/tmp/LoaderIn"),
		runLoader("WithoutMoveToOut"),
		verifyOutput("/tmp/LoaderIn"),
		testLoaderCheckAttributes,

		testLoaderResetDataDB,
		populateData("/tmp/SubpathWithoutMove/folder1"),
		runLoader("SubpathLoaderWithoutMove"),
		verifyOutput("/tmp/SubpathWithoutMove/folder1"),
		testLoaderCheckAttributes,

		testLoaderResetDataDB,
		populateData("/tmp/SubpathLoaderWithMove/folder1"),
		runLoader("SubpathLoaderWithMove"),
		verifyOutput("/tmp/SubpathOut/folder1"),
		testLoaderCheckAttributes,

		testLoaderResetDataDB,
		populateData("/tmp/templateLoaderIn"),
		runLoader("LoaderWithTemplate"),

		testLoaderResetDataDB,
		populateData("/tmp/templateLoaderOut"),
		testLoaderCheckAttributes,

		testLoaderResetDataDB,
		testLoaderPopulateDataForCustomSep,
		testLoaderCheckForCustomSep,
		testLoaderVerifyOutDirForCustomSep,

		testLoaderKillEngine,
	}
)

var (
	// waitRater = flag.Int("wait_rater", 200, "Number of miliseconds to wait for rater to start and cache")
	dataDir  = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")
	encoding = flag.String("rpc", utils.MetaJSON, "what encoding whould be used for rpc comunication")
	dbType   = flag.String("dbtype", utils.MetaInternal, "The type of DataBase (Internal/Mongo/mySql)")
)

var loaderPaths = []string{"/tmp/In", "/tmp/Out", "/tmp/LoaderIn", "/tmp/SubpathWithoutMove",
	"/tmp/SubpathLoaderWithMove", "/tmp/SubpathOut", "/tmp/templateLoaderIn", "/tmp/templateLoaderOut",
	"/tmp/customSepLoaderIn", "/tmp/customSepLoaderOut"}

func newRPCClient(cfg *config.ListenCfg) (c *birpc.Client, err error) {
	switch *encoding {
	case utils.MetaJSON:
		return jsonrpc.Dial(utils.TCP, cfg.RPCJSONListen)
	case utils.MetaGOB:
		return birpc.Dial(utils.TCP, cfg.RPCGOBListen)
	default:
		return nil, errors.New("UNSUPPORTED_RPC")
	}
}

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
	if loaderCfg, err = config.NewCGRConfigFromPath(context.Background(), loaderCfgPath); err != nil {
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
	if err := engine.InitDataDB(loaderCfg); err != nil {
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
	if loaderRPC, err = newRPCClient(loaderCfg.ListenCfg()); err != nil {
		t.Fatal(err)
	}
}

func populateData(inPath string) func(t *testing.T) {
	return func(t *testing.T) {
		if err := os.MkdirAll(inPath, 0755); err != nil {
			t.Fatal(inPath, err)
		}

		f, err := os.CreateTemp(utils.EmptyString, utils.AttributesCsv)
		if err != nil {
			t.Fatal(inPath, err)
		}
		if _, err := f.WriteString(engine.AttributesCSVContent); err != nil {
			t.Fatal(inPath, err)
		}
		if err = f.Sync(); err != nil {
			t.Fatal(inPath, err)
		}
		if err = f.Close(); err != nil {
			t.Fatal(inPath, err)
		}

		if err := os.Rename(f.Name(), path.Join(inPath, utils.AttributesCsv)); err != nil {
			t.Fatalf("Error moving file to processing directory(%s): %v", inPath, err)
		}
	}
}

func runLoader(loaderID string) func(t *testing.T) {
	return func(t *testing.T) {
		var reply string
		if err := loaderRPC.Call(context.Background(), utils.LoaderSv1Run,
			&ArgsProcessFolder{LoaderID: loaderID}, &reply); err != nil {
			t.Fatal(loaderID, err)
		} else if reply != utils.OK {
			t.Fatalf("<%s> Expected: %q, received: %q", loaderID, utils.OK, reply)
		}
	}
}

func verifyOutput(outPath string) func(t *testing.T) {
	return func(t *testing.T) {
		time.Sleep(100 * time.Millisecond)
		if outContent1, err := os.ReadFile(path.Join(outPath, utils.AttributesCsv)); err != nil {
			t.Fatal(outPath, err)
		} else if engine.AttributesCSVContent != string(outContent1) {
			t.Errorf("<%s>Expecting: %q, received: %q", outPath, engine.AttributesCSVContent, string(outContent1))
		}
	}
}

func testLoaderCheckAttributes(t *testing.T) {
	eAttrPrf := &engine.APIAttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		FilterIDs: []string{"*string:~*req.Account:1001", "*string:~*opts.*context:con1", "*string:~*opts.*context:con2|con3"},
		Attributes: []*engine.ExternalAttribute{{
			FilterIDs: []string{"*string:~*req.Field1:Initial"},
			Path:      utils.MetaReq + utils.NestingSep + "Field1",
			Type:      utils.MetaVariable,
			Value:     "Sub1",
		}, {
			Path:  utils.MetaReq + utils.NestingSep + "Field2",
			Type:  utils.MetaVariable,
			Value: "Sub2",
		}},
		Blocker: true,
		Weight:  20,
	}
	if *encoding == utils.MetaGOB { // gob threats empty slices as nil values
		eAttrPrf.Attributes[1].FilterIDs = nil
	}
	var reply *engine.APIAttributeProfile
	if err := loaderRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ALS1"},
		}, &reply); err != nil {
		t.Fatal(err)
	}
	sort.Strings(reply.FilterIDs)
	sort.Strings(eAttrPrf.FilterIDs)
	if !reflect.DeepEqual(eAttrPrf, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(eAttrPrf), utils.ToJSON(reply))
	}
}

func testLoaderPopulateDataForCustomSep(t *testing.T) {
	tmpFilePath := path.Join("/tmp/", utils.Attributes)
	if err := os.WriteFile(tmpFilePath, []byte(customAttributes), 0777); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.MkdirAll("/tmp/customSepLoaderIn", 0755); err != nil {
		t.Fatal("Error creating folder: /tmp/customSepLoaderIn", err)
	}
	if err := os.Rename(tmpFilePath, path.Join("/tmp/customSepLoaderIn", utils.Attributes)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testLoaderCheckForCustomSep(t *testing.T) {
	eAttrPrf := &engine.APIAttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_12012000001",
		FilterIDs: []string{"*string:~*req.Destination:12012000001"},
		Attributes: []*engine.ExternalAttribute{
			{
				Path:  "*req.Destination",
				Type:  utils.MetaConstant,
				Value: "12018209998",
			},
		},
	}
	if *encoding == utils.MetaGOB { // gob threats empty slices as nil values
		eAttrPrf.Attributes[0].FilterIDs = nil
	}
	var reply *engine.APIAttributeProfile
	if err := loaderRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_12012000001"},
		}, &reply); err != nil {
		t.Fatal(err)
	}
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

func testLoaderKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
