//go:build flaky

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
	"reflect"
	"strings"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	storageDb         Storage
	dm3               *DataManager
	versionsConfigDIR string
	versionCfg        *config.CGRConfig

	sTestsITVersions = []func(t *testing.T){
		testInitConfig,
		testInitDataDB,
		testVersionsFlush,
		testVersion,
		testVersionsFlush,
		testVersionsWithEngine,
		testUpdateVersionsAccounts,
		testUpdateVersionsActionPlans,
		testUpdateVersionsActionTriggers,
		testUpdateVersionsActions,
		testUpdateVersionsAttributes,
		testUpdateVersionsChargers,
		testUpdateVersionsDestinations,
		testUpdateVersionsDispatchers,
		testUpdateVersionsLoadIDs,
		testUpdateVersionsRQF,
		testUpdateVersionsRatingPlan,
		testUpdateVersionsRatingProfile,
		testUpdateVersionsResource,
		testUpdateVersionsReverseDestinations,
		testUpdateVersionsRoutes,
		testUpdateVersionsSharedGroups,
		testUpdateVersionsStats,
		testUpdateVersionsSubscribers,
		testUpdateVersionsThresholds,
		testUpdateVersionsTiming,
		testUpdateVersionsCostDetails,
		testUpdateVersionsSessionSCosts,
		testUpdateVersionsCDRs,
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
	if versionCfg, err = config.NewCGRConfigFromPath(path.Join(*utils.DataDir, "conf", "samples", versionsConfigDIR)); err != nil {
		t.Fatal(err)
	}
}

func testInitDataDB(t *testing.T) {
	dbConn, err := NewDataDBConn(versionCfg.DataDbCfg().Type,
		versionCfg.DataDbCfg().Host, versionCfg.DataDbCfg().Port,
		versionCfg.DataDbCfg().Name, versionCfg.DataDbCfg().User,
		versionCfg.DataDbCfg().Password, versionCfg.GeneralCfg().DBDataEncoding,
		versionCfg.DataDbCfg().Opts, versionCfg.DataDbCfg().Items)
	if err != nil {
		log.Fatal(err)
	}
	dm3 = NewDataManager(dbConn, versionCfg.CacheCfg(), nil)

	storageDb, err = NewStorDBConn(versionCfg.StorDbCfg().Type,
		versionCfg.StorDbCfg().Host, versionCfg.StorDbCfg().Port,
		versionCfg.StorDbCfg().Name, versionCfg.StorDbCfg().User,
		versionCfg.StorDbCfg().Password, versionCfg.GeneralCfg().DBDataEncoding,
		versionCfg.StorDbCfg().StringIndexedFields, versionCfg.StorDbCfg().PrefixIndexedFields,
		versionCfg.StorDbCfg().Opts, versionCfg.StorDbCfg().Items)
	if err != nil {
		log.Fatal(err)
	}
}

func testVersionsFlush(t *testing.T) {
	err := dm3.DataDB().Flush("")
	if err != nil {
		t.Error("Error when flushing Mongo ", err.Error())
	}
	if err := storageDb.Flush(path.Join(versionCfg.DataFolderPath, "storage", strings.Trim(versionCfg.StorDbCfg().Type, "*"))); err != nil {
		t.Error(err)
	}
	SetDBVersions(storageDb)
}

func testVersion(t *testing.T) {
	var test string
	var currentVersion Versions
	var testVersion Versions
	dataDbVersions := CurrentDataDBVersions()
	storDbVersions := CurrentStorDBVersions()

	allVersions := make(Versions)
	for k, v := range dataDbVersions {
		allVersions[k] = v
	}
	for k, v := range storDbVersions {
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
	if err = dm3.DataDB().RemoveVersions(currentVersion); err != nil {
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
	if err = dm3.DataDB().RemoveVersions(testVersion); err != nil {
		t.Error(err)
	}
	storType = storageDb.GetStorageType()
	switch storType {
	case utils.MetaInternal:
		currentVersion = allVersions
		testVersion = allVersions
		testVersion[utils.Accounts] = 1
		test = "Migration needed: please backup cgr data and run: <cgr-migrator -exec=*accounts>"
	case utils.MetaMongo, utils.MetaPostgres, utils.MetaMySQL:
		currentVersion = storDbVersions
		testVersion = allVersions
		testVersion[utils.CostDetails] = 1
		test = "Migration needed: please backup cgr data and run: <cgr-migrator -exec=*cost_details>"
	}
	//storageDb

	if err := CheckVersions(storageDb); err != nil {
		t.Error(err)
	}
	if rcv, err := storageDb.GetVersions(""); err != nil {
		t.Error(err)
	} else if len(currentVersion) != len(rcv) {
		t.Errorf("Expecting: %v, received: %v", currentVersion, rcv)
	}
	if err = storageDb.RemoveVersions(currentVersion); err != nil {
		t.Error(err)
	}
	if _, rcvErr := storageDb.GetVersions(""); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := storageDb.SetVersions(testVersion, false); err != nil {
		t.Error(err)
	}
	if err := CheckVersions(storageDb); err != nil && err.Error() != test {
		t.Error(err)
	}
	if err = storageDb.RemoveVersions(testVersion); err != nil {
		t.Error(err)
	}

}

func testVersionsWithEngine(t *testing.T) {
	cmd := exec.Command("cgr-engine", fmt.Sprintf(`-config_path=/usr/share/cgrates/conf/samples/%s`, versionsConfigDIR), `-scheduled_shutdown=4ms`)
	output := bytes.NewBuffer(nil)
	cmd.Stderr = output
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	if output.String() != utils.EmptyString {
		t.Fatalf("Expected empty but received: %q", output.String())
	}
	dataDbQueryVersions, err := dm3.DataDB().GetVersions("")
	if err != nil {
		t.Error(err)
	}
	storDbQueryVersions, err := storageDb.GetVersions("")
	if err != nil {
		t.Error(err)
	}
	expectDataDb := CurrentDataDBVersions()
	expectStorDb := CurrentStorDBVersions()
	if !reflect.DeepEqual(dataDbQueryVersions, expectDataDb) {
		t.Fatalf("Expected %v \n  but received \n %v", utils.ToJSON(expectDataDb), utils.ToJSON(dataDbQueryVersions))
	} else if !reflect.DeepEqual(storDbQueryVersions, expectStorDb) {
		t.Fatalf("Expected %v \n  but received \n %v", utils.ToJSON(expectStorDb), utils.ToJSON(storDbQueryVersions))
	}
}

// Tests for DataDB
// We do a test for each version field in order to test them as a unit and not at as a whole.
func testUpdateVersionsAccounts(t *testing.T) {
	newVersions := CurrentDataDBVersions()
	newVersions[utils.Accounts] = 2
	if err := dm3.DataDB().SetVersions(newVersions, true); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("cgr-engine", fmt.Sprintf(`-config_path=/usr/share/cgrates/conf/samples/%s`, versionsConfigDIR), `-scheduled_shutdown=4ms`)
	output := bytes.NewBuffer(nil)
	cmd.Stderr = output
	if err := cmd.Run(); err == nil ||
		err.Error() != "exit status 1" {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatalf("expected: %s, \nreceived: %s", "exit status 1", err)
	}
	errExpect := "Migration needed: please backup cgr data and run: <cgr-migrator -exec=*accounts>"
	if !strings.Contains(output.String(), errExpect) {
		t.Errorf("expected %s \nto contain: %s", output.String(), errExpect)
	}
}

func testUpdateVersionsActionPlans(t *testing.T) {
	newVersions := CurrentDataDBVersions()
	newVersions[utils.ActionPlans] = 2
	if err := dm3.DataDB().SetVersions(newVersions, true); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("cgr-engine", fmt.Sprintf(`-config_path=/usr/share/cgrates/conf/samples/%s`, versionsConfigDIR), `-scheduled_shutdown=4ms`)
	output := bytes.NewBuffer(nil)
	cmd.Stderr = output
	if err := cmd.Run(); err == nil ||
		err.Error() != "exit status 1" {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatalf("expected: %s, \nreceived: %s", "exit status 1", err)
	}
	errExpect := "Migration needed: please backup cgr data and run: <cgr-migrator -exec=*action_plans>"
	if !strings.Contains(output.String(), errExpect) {
		t.Errorf("expected %s \nto contain: %s", output.String(), errExpect)
	}
}

func testUpdateVersionsActionTriggers(t *testing.T) {
	newVersions := CurrentDataDBVersions()
	newVersions[utils.ActionTriggers] = 1
	if err := dm3.DataDB().SetVersions(newVersions, true); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("cgr-engine", fmt.Sprintf(`-config_path=/usr/share/cgrates/conf/samples/%s`, versionsConfigDIR), `-scheduled_shutdown=4ms`)
	output := bytes.NewBuffer(nil)
	cmd.Stderr = output
	if err := cmd.Run(); err == nil ||
		err.Error() != "exit status 1" {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatalf("expected: %s, \nreceived: %s", "exit status 1", err)
	}
	errExpect := "Migration needed: please backup cgr data and run: <cgr-migrator -exec=*action_triggers>"
	if !strings.Contains(output.String(), errExpect) {
		t.Errorf("expected %s \nto contain: %s", output.String(), errExpect)
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
	cmd.Stderr = output
	if err := cmd.Run(); err == nil ||
		err.Error() != "exit status 1" {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatalf("expected: %s, \nreceived: %s", "exit status 1", err)
	}
	errExpect := "Migration needed: please backup cgr data and run: <cgr-migrator -exec=*actions>"
	if !strings.Contains(output.String(), errExpect) {
		t.Errorf("expected %s \nto contain: %s", output.String(), errExpect)
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
	cmd.Stderr = output
	if err := cmd.Run(); err == nil ||
		err.Error() != "exit status 1" {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatalf("expected: %s, \nreceived: %s", "exit status 1", err)
	}
	errExpect := "Migration needed: please backup cgr data and run: <cgr-migrator -exec=*chargers>"
	if !strings.Contains(output.String(), errExpect) {
		t.Errorf("expected %s \nto contain: %s", output.String(), errExpect)
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
	cmd.Stderr = output
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
	cmd.Stderr = output
	if err := cmd.Run(); err == nil ||
		err.Error() != "exit status 1" {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatalf("expected: %s, \nreceived: %s", "exit status 1", err)
	}
	errExpect := "Migration needed: please backup cgr data and run: <cgr-migrator -exec=*attributes>"
	if !strings.Contains(output.String(), errExpect) {
		t.Errorf("expected %s \nto contain: %s", output.String(), errExpect)
	}
}

func testUpdateVersionsDispatchers(t *testing.T) {
	newVersions := CurrentDataDBVersions()
	newVersions[utils.Dispatchers] = 1
	if err := dm3.DataDB().SetVersions(newVersions, true); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("cgr-engine", fmt.Sprintf(`-config_path=/usr/share/cgrates/conf/samples/%s`, versionsConfigDIR), `-scheduled_shutdown=4ms`)
	output := bytes.NewBuffer(nil)
	cmd.Stderr = output
	if err := cmd.Run(); err == nil ||
		err.Error() != "exit status 1" {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatalf("expected: %s, \nreceived: %s", "exit status 1", err)
	}
	errExpect := "Migration needed: please backup cgr data and run: <cgr-migrator -exec=*dispatchers>"
	if !strings.Contains(output.String(), errExpect) {
		t.Errorf("expected %s \nto contain: %s", output.String(), errExpect)
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
	cmd.Stderr = output
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
	cmd.Stderr = output
	if err := cmd.Run(); err == nil ||
		err.Error() != "exit status 1" {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatalf("expected: %s, \nreceived: %s", "exit status 1", err)
	}
	errExpect := "Migration needed: please backup cgr data and run: <cgr-migrator -exec=*filters>"
	if !strings.Contains(output.String(), errExpect) {
		t.Errorf("expected %s \nto contain: %s", output.String(), errExpect)
	}
}

func testUpdateVersionsRatingPlan(t *testing.T) {
	newVersions := CurrentDataDBVersions()
	newVersions[utils.RatingPlan] = 0
	if err := dm3.DataDB().SetVersions(newVersions, true); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("cgr-engine", fmt.Sprintf(`-config_path=/usr/share/cgrates/conf/samples/%s`, versionsConfigDIR), `-scheduled_shutdown=4ms`)
	output := bytes.NewBuffer(nil)
	cmd.Stderr = output
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

func testUpdateVersionsRatingProfile(t *testing.T) {
	newVersions := CurrentDataDBVersions()
	newVersions[utils.RatingProfile] = 0
	if err := dm3.DataDB().SetVersions(newVersions, true); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("cgr-engine", fmt.Sprintf(`-config_path=/usr/share/cgrates/conf/samples/%s`, versionsConfigDIR), `-scheduled_shutdown=4ms`)
	output := bytes.NewBuffer(nil)
	cmd.Stderr = output
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

func testUpdateVersionsResource(t *testing.T) {
	newVersions := CurrentDataDBVersions()
	newVersions[utils.Resource] = 0
	if err := dm3.DataDB().SetVersions(newVersions, true); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("cgr-engine", fmt.Sprintf(`-config_path=/usr/share/cgrates/conf/samples/%s`, versionsConfigDIR), `-scheduled_shutdown=4ms`)
	output := bytes.NewBuffer(nil)
	cmd.Stderr = output
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

func testUpdateVersionsReverseDestinations(t *testing.T) {
	newVersions := CurrentDataDBVersions()
	newVersions[utils.ReverseDestinations] = 0
	if err := dm3.DataDB().SetVersions(newVersions, true); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("cgr-engine", fmt.Sprintf(`-config_path=/usr/share/cgrates/conf/samples/%s`, versionsConfigDIR), `-scheduled_shutdown=4ms`)
	output := bytes.NewBuffer(nil)
	cmd.Stderr = output
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
	cmd.Stderr = output
	if err := cmd.Run(); err == nil ||
		err.Error() != "exit status 1" {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatalf("expected: %s, \nreceived: %s", "exit status 1", err)
	}
	errExpect := "Migration needed: please backup cgr data and run: <cgr-migrator -exec=*routes>"
	if !strings.Contains(output.String(), errExpect) {
		t.Errorf("expected %s \nto contain: %s", output.String(), errExpect)
	}
}

func testUpdateVersionsSharedGroups(t *testing.T) {
	newVersions := CurrentDataDBVersions()
	newVersions[utils.SharedGroups] = 1
	if err := dm3.DataDB().SetVersions(newVersions, true); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("cgr-engine", fmt.Sprintf(`-config_path=/usr/share/cgrates/conf/samples/%s`, versionsConfigDIR), `-scheduled_shutdown=4ms`)
	output := bytes.NewBuffer(nil)
	cmd.Stderr = output
	if err := cmd.Run(); err == nil ||
		err.Error() != "exit status 1" {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatalf("expected: %s, \nreceived: %s", "exit status 1", err)
	}
	errExpect := "Migration needed: please backup cgr data and run: <cgr-migrator -exec=*shared_groups>"
	if !strings.Contains(output.String(), errExpect) {
		t.Errorf("expected %s \nto contain: %s", output.String(), errExpect)
	}
}

func testUpdateVersionsStats(t *testing.T) {
	newVersions := CurrentDataDBVersions()
	newVersions[utils.StatS] = 3
	if err := dm3.DataDB().SetVersions(newVersions, true); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("cgr-engine", fmt.Sprintf(`-config_path=/usr/share/cgrates/conf/samples/%s`, versionsConfigDIR), `-scheduled_shutdown=4ms`)
	output := bytes.NewBuffer(nil)
	cmd.Stderr = output
	if err := cmd.Run(); err == nil ||
		err.Error() != "exit status 1" {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatalf("expected: %s, \nreceived: %s", "exit status 1", err)
	}
	errExpect := "Migration needed: please backup cgr data and run: <cgr-migrator -exec=*stats>"
	if !strings.Contains(output.String(), errExpect) {
		t.Errorf("expected %s \nto contain: %s", output.String(), errExpect)
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
	cmd.Stderr = output
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
	cmd.Stderr = output
	if err := cmd.Run(); err == nil ||
		err.Error() != "exit status 1" {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatalf("expected: %s, \nreceived: %s", "exit status 1", err)
	}
	errExpect := "Migration needed: please backup cgr data and run: <cgr-migrator -exec=*thresholds>"
	if !strings.Contains(output.String(), errExpect) {
		t.Errorf("expected %s \nto contain: %s", output.String(), errExpect)
	}
}

func testUpdateVersionsTiming(t *testing.T) {
	newVersions := CurrentDataDBVersions()
	newVersions[utils.Timing] = 0
	if err := dm3.DataDB().SetVersions(newVersions, true); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("cgr-engine", fmt.Sprintf(`-config_path=/usr/share/cgrates/conf/samples/%s`, versionsConfigDIR), `-scheduled_shutdown=4ms`)
	output := bytes.NewBuffer(nil)
	cmd.Stderr = output
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

// Tests for StorDB
func testUpdateVersionsCostDetails(t *testing.T) {
	newVersions := CurrentStorDBVersions()
	newVersions[utils.CostDetails] = 1
	if err := storageDb.SetVersions(newVersions, true); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("cgr-engine", fmt.Sprintf(`-config_path=/usr/share/cgrates/conf/samples/%s`, versionsConfigDIR), `-scheduled_shutdown=4ms`)
	output := bytes.NewBuffer(nil)
	cmd.Stderr = output
	if err := cmd.Run(); err == nil ||
		err.Error() != "exit status 1" {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatalf("expected: %s, \nreceived: %s", "exit status 1", err)
	}
	errExpect := "Migration needed: please backup cgr data and run: <cgr-migrator -exec=*cost_details>"
	if !strings.Contains(output.String(), errExpect) {
		t.Errorf("expected %s \nto contain: %s", output.String(), errExpect)
	}
}

func testUpdateVersionsSessionSCosts(t *testing.T) {
	newVersions := CurrentStorDBVersions()
	newVersions[utils.SessionSCosts] = 2
	if err := storageDb.SetVersions(newVersions, true); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("cgr-engine", fmt.Sprintf(`-config_path=/usr/share/cgrates/conf/samples/%s`, versionsConfigDIR), `-scheduled_shutdown=4ms`)
	output := bytes.NewBuffer(nil)
	cmd.Stderr = output
	if err := cmd.Run(); err == nil ||
		err.Error() != "exit status 1" {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatalf("expected: %s, \nreceived: %s", "exit status 1", err)
	}
	errExpect := "Migration needed: please backup cgr data and run: <cgr-migrator -exec=*sessions_costs>"
	if !strings.Contains(output.String(), errExpect) {
		t.Errorf("expected %s \nto contain: %s", output.String(), errExpect)
	}
}

func testUpdateVersionsCDRs(t *testing.T) {
	newVersions := CurrentStorDBVersions()
	newVersions[utils.CDRs] = 1
	if err := storageDb.SetVersions(newVersions, true); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("cgr-engine", fmt.Sprintf(`-config_path=/usr/share/cgrates/conf/samples/%s`, versionsConfigDIR), `-scheduled_shutdown=4ms`)
	output := bytes.NewBuffer(nil)
	cmd.Stderr = output
	if err := cmd.Run(); err == nil ||
		err.Error() != "exit status 1" {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatalf("expected: %s, \nreceived: %s", "exit status 1", err)
	}
	errExpect := "Migration needed: please backup cgr data and run: <cgr-migrator -exec=*cdrs>"
	if !strings.Contains(output.String(), errExpect) {
		t.Errorf("expected %s \nto contain: %s", output.String(), errExpect)
	}
}
