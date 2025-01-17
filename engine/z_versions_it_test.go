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
package engine

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"path"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	dm3               *DataManager
	versionsConfigDIR string
	vrsCfg            *config.CGRConfig
	sTestsITVersions  = []func(t *testing.T){
		testInitConfig,
		testInitDataDB,
		testVersionsFlush,
		testVersion,
		testVersionsFlush,
	}
)

func TestVersionsIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		t.SkipNow()
	case utils.MetaMySQL:
		versionsConfigDIR = "tutmysql"
	case utils.MetaMongo:
		versionsConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		versionsConfigDIR = "tutpostgres"
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestsITVersions {
		t.Run(versionsConfigDIR, stest)
	}
}

func testInitConfig(t *testing.T) {
	var err error
	if vrsCfg, err = config.NewCGRConfigFromPath(context.Background(), path.Join(*utils.DataDir, "conf", "samples", versionsConfigDIR)); err != nil {
		t.Fatal(err)
	}
}

func testInitDataDB(t *testing.T) {
	dbConn, err := NewDataDBConn(vrsCfg.DataDbCfg().Type,
		vrsCfg.DataDbCfg().Host, vrsCfg.DataDbCfg().Port,
		vrsCfg.DataDbCfg().Name, vrsCfg.DataDbCfg().User,
		vrsCfg.DataDbCfg().Password, vrsCfg.GeneralCfg().DBDataEncoding,
		vrsCfg.DataDbCfg().Opts, vrsCfg.DataDbCfg().Items)
	if err != nil {
		log.Fatal(err)
	}
	dm3 = NewDataManager(dbConn, vrsCfg.CacheCfg(), nil)

	if err != nil {
		log.Fatal(err)
	}
}

func testVersionsFlush(t *testing.T) {
	err := dm3.DataDB().Flush("")
	if err != nil {
		t.Error("Error when flushing Mongo ", err.Error())
	}

}

func testVersion(t *testing.T) {
	var test string
	var currentVersion Versions
	var testVersion Versions
	dataDbVersions := CurrentDataDBVersions()

	allVersions := make(Versions)
	for k, v := range dataDbVersions {
		allVersions[k] = v
	}

	storType := dm3.DataDB().GetStorageType()
	switch storType {
	case utils.MetaInternal:
		currentVersion = allVersions
		testVersion = allVersions
		testVersion[utils.Accounts] = 1
		test = "Migration needed: please backup cgr data and run: <cgr-migrator -exec=*accounts>"
	case utils.MetaMongo, utils.MetaRedis:
		currentVersion = dataDbVersions
		testVersion = dataDbVersions
		testVersion[utils.Accounts] = 1

		test = "Migration needed: please backup cgr data and run: <cgr-migrator -exec=*accounts>"
	}

	//dataDB
	if _, rcvErr := dm3.DataDB().GetVersions(""); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := CheckVersions(dm3.DataDB()); err != nil {
		t.Error(err)
	}
	if rcv, err := dm3.DataDB().GetVersions(""); err != nil {
		t.Error(err)
	} else if len(currentVersion) != len(rcv) {
		t.Errorf("Expecting: %v, received: %v", currentVersion, rcv)
	}
	if err := dm3.DataDB().RemoveVersions(currentVersion); err != nil {
		t.Error(err)
	}
	if _, rcvErr := dm3.DataDB().GetVersions(""); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := dm3.DataDB().SetVersions(testVersion, false); err != nil {
		t.Error(err)
	}
	if err := CheckVersions(dm3.DataDB()); err.Error() != test {
		t.Error(err)
	}
	if err := dm3.DataDB().RemoveVersions(testVersion); err != nil {
		t.Error(err)
	}
	switch storType {
	case utils.MetaInternal:
		currentVersion = allVersions
		testVersion = allVersions
		testVersion[utils.Accounts] = 1
		test = "Migration needed: please backup cgr data and run: <cgr-migrator -exec=*accounts>"
	case utils.MetaMongo, utils.MetaPostgres, utils.MetaMySQL:
		testVersion = allVersions
		testVersion[utils.CostDetails] = 1
		test = "Migration needed: please backup cgr data and run: <cgr-migrator -exec=*cost_details>"
	}

}

func testUpdateVersionsAccounts(t *testing.T) {
	newVersions := CurrentDataDBVersions()
	newVersions[utils.Accounts] = 2
	if err := dm3.DataDB().SetVersions(newVersions, true); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("cgr-engine", fmt.Sprintf(`-config_path=/usr/share/cgrates/conf/samples/%s`, versionsConfigDIR), `-scheduled_shutdown=4ms`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	errExpect := "Migration needed: please backup cgr data and run: <cgr-migrator -exec=*accounts>\n"
	if output.String() != errExpect {
		t.Fatalf("Expected %q \n but received: \n %q", errExpect, output.String())
	}
}

func testUpdateVersionsActions(t *testing.T) {
	newVersions := CurrentDataDBVersions()
	newVersions[utils.Actions] = 1
	if err := dm3.DataDB().SetVersions(newVersions, true); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("cgr-engine", fmt.Sprintf(`-config_path=/usr/share/cgrates/conf/samples/%s`, versionsConfigDIR), `-scheduled_shutdown=4ms`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	errExpect := "Migration needed: please backup cgr data and run: <cgr-migrator -exec=*actions>\n"
	if output.String() != errExpect {
		t.Fatalf("Expected %q \n but received: \n %q", errExpect, output.String())
	}
}

func testUpdateVersionsChargers(t *testing.T) {
	newVersions := CurrentDataDBVersions()
	newVersions[utils.Chargers] = 1
	if err := dm3.DataDB().SetVersions(newVersions, true); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("cgr-engine", fmt.Sprintf(`-config_path=/usr/share/cgrates/conf/samples/%s`, versionsConfigDIR), `-scheduled_shutdown=4ms`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	errExpect := "Migration needed: please backup cgr data and run: <cgr-migrator -exec=*chargers>\n"
	if output.String() != errExpect {
		t.Fatalf("Expected %q \n but received: \n %q", errExpect, output.String())
	}
}

func testUpdateVersionsDestinations(t *testing.T) {
	newVersions := CurrentDataDBVersions()
	newVersions[utils.Destination] = 0
	if err := dm3.DataDB().SetVersions(newVersions, true); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("cgr-engine", fmt.Sprintf(`-config_path=/usr/share/cgrates/conf/samples/%s`, versionsConfigDIR), `-scheduled_shutdown=4ms`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	errExpect := utils.EmptyString
	if output.String() != errExpect {
		t.Fatalf("Expected %q \n but received: \n %q", errExpect, output.String())
	}
}
func testUpdateVersionsAttributes(t *testing.T) {
	newVersions := CurrentDataDBVersions()
	newVersions[utils.Attributes] = 3
	if err := dm3.DataDB().SetVersions(newVersions, true); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("cgr-engine", fmt.Sprintf(`-config_path=/usr/share/cgrates/conf/samples/%s`, versionsConfigDIR), `-scheduled_shutdown=4ms`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	errExpect := "Migration needed: please backup cgr data and run: <cgr-migrator -exec=*attributes>\n"
	if output.String() != errExpect {
		t.Fatalf("Expected %q \n but received: \n %q", errExpect, output.String())
	}
}

func testUpdateVersionsLoadIDs(t *testing.T) {
	newVersions := CurrentDataDBVersions()
	delete(newVersions, utils.LoadIDsVrs)
	if err := dm3.DataDB().SetVersions(newVersions, true); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("cgr-engine", fmt.Sprintf(`-config_path=/usr/share/cgrates/conf/samples/%s`, versionsConfigDIR), `-scheduled_shutdown=4ms`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	errExpect := utils.EmptyString
	if output.String() != errExpect {
		t.Fatalf("Expected %q \n but received: \n %q", errExpect, output.String())
	}
}

func testUpdateVersionsRQF(t *testing.T) {
	newVersions := CurrentDataDBVersions()
	newVersions[utils.RQF] = 2
	if err := dm3.DataDB().SetVersions(newVersions, true); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("cgr-engine", fmt.Sprintf(`-config_path=/usr/share/cgrates/conf/samples/%s`, versionsConfigDIR), `-scheduled_shutdown=4ms`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	errExpect := "Migration needed: please backup cgr data and run: <cgr-migrator -exec=*filters>\n"
	if output.String() != errExpect {
		t.Fatalf("Expected %q \n but received: \n %q", errExpect, output.String())
	}
}

func testUpdateVersionsResource(t *testing.T) {
	newVersions := CurrentDataDBVersions()
	newVersions[utils.Resource] = 0
	if err := dm3.DataDB().SetVersions(newVersions, true); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("cgr-engine", fmt.Sprintf(`-config_path=/usr/share/cgrates/conf/samples/%s`, versionsConfigDIR), `-scheduled_shutdown=4ms`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	errExpect := utils.EmptyString
	if output.String() != errExpect {
		t.Fatalf("Expected %q \n but received: \n %q", errExpect, output.String())
	}
}

func testUpdateVersionsRoutes(t *testing.T) {
	newVersions := CurrentDataDBVersions()
	newVersions[utils.Routes] = 1
	if err := dm3.DataDB().SetVersions(newVersions, true); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("cgr-engine", fmt.Sprintf(`-config_path=/usr/share/cgrates/conf/samples/%s`, versionsConfigDIR), `-scheduled_shutdown=4ms`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	errExpect := "Migration needed: please backup cgr data and run: <cgr-migrator -exec=*routes>\n"
	if output.String() != errExpect {
		t.Fatalf("Expected %q \n but received: \n %q", errExpect, output.String())
	}
}

func testUpdateVersionsStats(t *testing.T) {
	newVersions := CurrentDataDBVersions()
	newVersions[utils.Stats] = 3
	if err := dm3.DataDB().SetVersions(newVersions, true); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("cgr-engine", fmt.Sprintf(`-config_path=/usr/share/cgrates/conf/samples/%s`, versionsConfigDIR), `-scheduled_shutdown=4ms`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	errExpect := "Migration needed: please backup cgr data and run: <cgr-migrator -exec=*stats>\n"
	if output.String() != errExpect {
		t.Fatalf("Expected %q \n but received: \n %q", errExpect, output.String())
	}
}

func testUpdateVersionsSubscribers(t *testing.T) {
	newVersions := CurrentDataDBVersions()
	newVersions[utils.Subscribers] = 0
	if err := dm3.DataDB().SetVersions(newVersions, true); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("cgr-engine", fmt.Sprintf(`-config_path=/usr/share/cgrates/conf/samples/%s`, versionsConfigDIR), `-scheduled_shutdown=4ms`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	errExpect := utils.EmptyString
	if output.String() != errExpect {
		t.Fatalf("Expected %q \n but received: \n %q", errExpect, output.String())
	}
}

func testUpdateVersionsThresholds(t *testing.T) {
	newVersions := CurrentDataDBVersions()
	newVersions[utils.Thresholds] = 2
	if err := dm3.DataDB().SetVersions(newVersions, true); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("cgr-engine", fmt.Sprintf(`-config_path=/usr/share/cgrates/conf/samples/%s`, versionsConfigDIR), `-scheduled_shutdown=4ms`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	errExpect := "Migration needed: please backup cgr data and run: <cgr-migrator -exec=*thresholds>\n"
	if output.String() != errExpect {
		t.Fatalf("Expected %q \n but received: \n %q", errExpect, output.String())
	}
}
