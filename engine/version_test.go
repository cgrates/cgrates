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
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestVersionCompare(t *testing.T) {
	x := Versions{utils.Accounts: 2, utils.Actions: 2,
		utils.ActionTriggers: 2, utils.ActionPlans: 2,
		utils.SharedGroups: 2, utils.CostDetails: 2}
	y := Versions{utils.Accounts: 1, utils.Actions: 2,
		utils.ActionTriggers: 2, utils.ActionPlans: 2,
		utils.SharedGroups: 2, utils.CostDetails: 2}
	z := Versions{utils.Accounts: 2, utils.Actions: 2,
		utils.ActionTriggers: 2, utils.ActionPlans: 1,
		utils.SharedGroups: 2, utils.CostDetails: 2}
	q := Versions{utils.Accounts: 2, utils.Actions: 2,
		utils.ActionTriggers: 2, utils.ActionPlans: 2,
		utils.SharedGroups: 1, utils.CostDetails: 2}
	c := Versions{utils.CostDetails: 1}
	a := Versions{utils.Accounts: 2, utils.Actions: 2,
		utils.ActionTriggers: 2, utils.ActionPlans: 2,
		utils.SharedGroups: 2, utils.CostDetails: 2,
		utils.SessionSCosts: 1}
	b := Versions{utils.Accounts: 2, utils.Actions: 2,
		utils.ActionTriggers: 2, utils.ActionPlans: 2,
		utils.SharedGroups: 2, utils.CostDetails: 2,
		utils.SessionSCosts: 2}

	message1 := y.Compare(x, utils.Mongo, true)
	if message1 != "cgr-migrator -exec=*accounts" {
		t.Errorf("Error failed to compare to curent version expected: %s received: %s", "cgr-migrator -exec=*accounts", message1)
	}
	message2 := z.Compare(x, utils.Mongo, true)
	if message2 != "cgr-migrator -exec=*action_plans" {
		t.Errorf("Error failed to compare to curent version expected: %s received: %s", "cgr-migrator -exec=*action_plans", message2)
	}
	message3 := q.Compare(x, utils.Mongo, true)
	if message3 != "cgr-migrator -exec=*shared_groups" {
		t.Errorf("Error failed to compare to curent version expected: %s received: %s", "cgr-migrator -exec=*shared_groups", message3)
	}
	message4 := c.Compare(x, utils.Mongo, false)
	if message4 != "cgr-migrator -exec=*cost_details" {
		t.Errorf("Error failed to compare to curent version expected: %s received: %s", "cgr-migrator -exec=*cost_details", message4)
	}
	message5 := a.Compare(b, utils.MySQL, false)
	if message5 != "cgr-migrator -exec=*sessions_costs" {
		t.Errorf("Error failed to compare to curent version expected: %s received: %s", "cgr-migrator -exec=*sessions_costs", message5)
	}
	message6 := a.Compare(b, utils.Postgres, false)
	if message6 != "cgr-migrator -exec=*sessions_costs" {
		t.Errorf("Error failed to compare to curent version expected: %s received: %s", "cgr-migrator -exec=*sessions_costs", message6)
	}
	message7 := y.Compare(x, utils.Redis, true)
	if message7 != "cgr-migrator -exec=*accounts" {
		t.Errorf("Error failed to compare to curent version expected: %s received: %s", "cgr-migrator -exec=*accounts", message7)
	}

	y[utils.Accounts] = 2
	message8 := y.Compare(x, utils.Redis, true)
	if message8 != utils.EmptyString {
		t.Errorf("Expected %+v, received %+v", utils.EmptyString, message8)
	}
}

func TestCurrentDBVersions(t *testing.T) {
	expVersDataDB := Versions{
		utils.StatS: 4, utils.Accounts: 3, utils.Actions: 2,
		utils.ActionTriggers: 2, utils.ActionPlans: 3, utils.SharedGroups: 2,
		utils.Thresholds: 4, utils.Routes: 2, utils.Attributes: 6,
		utils.Timing: 1, utils.RQF: 5, utils.Resource: 1,
		utils.Subscribers: 1, utils.Destinations: 1, utils.ReverseDestinations: 1,
		utils.RatingPlan: 1, utils.RatingProfile: 1, utils.Chargers: 2,
		utils.Dispatchers: 2, utils.LoadIDsVrs: 1,
	}
	expVersStorDB := Versions{
		utils.CostDetails: 2, utils.SessionSCosts: 3, utils.CDRs: 2,
		utils.TpRatingPlans: 1, utils.TpFilters: 1, utils.TpDestinationRates: 1,
		utils.TpActionTriggers: 1, utils.TpAccountActionsV: 1, utils.TpActionPlans: 1,
		utils.TpActions: 1, utils.TpThresholds: 1, utils.TpRoutes: 1,
		utils.TpStats: 1, utils.TpSharedGroups: 1, utils.TpRatingProfiles: 1,
		utils.TpResources: 1, utils.TpRates: 1, utils.TpTiming: 1,
		utils.TpResource: 1, utils.TpDestinations: 1, utils.TpRatingPlan: 1,
		utils.TpRatingProfile: 1, utils.TpChargers: 1, utils.TpDispatchers: 1,
	}
	if vrs := CurrentDBVersions(utils.Mongo, true); !reflect.DeepEqual(expVersDataDB, vrs) {
		t.Errorf("Expectred %+v, received %+v", expVersDataDB, vrs)
	}

	if vrs := CurrentDBVersions(utils.Mongo, false); !reflect.DeepEqual(expVersStorDB, vrs) {
		t.Errorf("Expectred %+v, received %+v", expVersStorDB, vrs)
	}

	if vrs := CurrentDBVersions(utils.Postgres, false); !reflect.DeepEqual(expVersStorDB, vrs) {
		t.Errorf("Expectred %+v, received %+v", expVersStorDB, vrs)
	}

	if vrs := CurrentDBVersions(utils.Redis, true); !reflect.DeepEqual(expVersDataDB, vrs) {
		t.Errorf("Expectred %+v, received %+v", expVersDataDB, vrs)
	}

	if vrs := CurrentDBVersions("NOT_A_DB", true); vrs != nil {
		t.Error(vrs)
	}

	//Compare AllVersions
	expStr := "cgr-migrator"
	if rcv := expVersDataDB.Compare(expVersStorDB, utils.INTERNAL, true); !strings.Contains(rcv, expStr) {
		t.Errorf("Expected %+v, received %+v", expStr, rcv)
	}
}

type storageMock struct {
	testcase string
}

func (sM *storageMock) Close() {}

func (sM *storageMock) Flush(string) error {
	return nil
}

func (sM *storageMock) GetKeysForPrefix(string) ([]string, error) {
	return nil, nil
}

func (sM *storageMock) RemoveKeysForPrefix(string) error {
	return nil
}

func (sM *storageMock) GetVersions(itm string) (vrs Versions, err error) {
	switch sM.testcase {
	case "GetVersions non-nil error":
		err = fmt.Errorf("GetVersions mock error")
		return nil, err
	case "GetVersions ErrNotFound":
		return nil, utils.ErrNotFound
	case "ErrNotFound empty db":
		return nil, utils.ErrNotFound
	case "ErrNotFound IsDBEmpty error":
		return nil, utils.ErrNotFound
	}
	return nil, nil
}

func (sM *storageMock) SetVersions(vrs Versions, overwrite bool) (err error) {
	switch sM.testcase {
	case "setDBVersions error":
		err = fmt.Errorf("SetVersions error")
		return
	case "ErrNotFound empty db":
		err = fmt.Errorf("OverwriteDBVersions mock error")
		return
	}
	return
}

func (sM *storageMock) RemoveVersions(vrs Versions) (err error) {
	return nil
}

func (sM *storageMock) SelectDatabase(dbName string) (err error) {
	return nil
}

func (sM *storageMock) GetStorageType() string {
	switch sM.testcase {
	case "Compare returns non-nil message":
		return utils.Postgres
	}
	return ""
}

func (sM *storageMock) IsDBEmpty() (resp bool, err error) {
	switch sM.testcase {
	case "ErrNotFound empty db":
		return true, nil
	case "ErrNotFound IsDBEmpty error":
		err = fmt.Errorf("IsDBEmpty mock error")
		return false, err
	}
	return false, nil
}

func TestVersionSetDBVersions(t *testing.T) {
	storage := &storageMock{}

	err := SetDBVersions(storage)

	if err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}
}

func TestVersionisDataDB(t *testing.T) {
	var storage Storage = &MongoStorage{}
	rcv := isDataDB(storage)

	if rcv != false {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", false, rcv)
	}
}

func TestVersionsetDBVersions(t *testing.T) {
	storage := &storageMock{
		testcase: "setDBVersions error",
	}

	experr := "SetVersions error"
	err := setDBVersions(storage, false)

	if err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestVersionCheckVersionsNilError(t *testing.T) {
	storage := &storageMock{}
	err := CheckVersions(storage)

	if err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}
}

func TestVersionCheckVersionsGetVersionsNonNilErr(t *testing.T) {
	storage := &storageMock{
		testcase: "GetVersions non-nil error",
	}

	experr := "GetVersions mock error"
	err := CheckVersions(storage)

	if err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestVersionCheckVersionsGetVersionsErrNotFoundEmptyDB(t *testing.T) {
	storage := &storageMock{
		testcase: "ErrNotFound empty db",
	}

	experr := "OverwriteDBVersions mock error"
	err := CheckVersions(storage)

	if err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestVersionCheckVersionsGetVersionsErrNotFoundNonEmptyDB(t *testing.T) {
	storage := &storageMock{
		testcase: "GetVersions ErrNotFound",
	}

	experr := "No versions defined: please backup cgrates data and run : <cgr-migrator -exec=*set_versions>"
	err := CheckVersions(storage)

	if err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestVersionCheckVersionsGetVersionsErrNotFoundDataDBErr(t *testing.T) {
	storage := &storageMock{
		testcase: "ErrNotFound IsDBEmpty error",
	}

	experr := "IsDBEmpty mock error"
	err := CheckVersions(storage)

	if err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestVersionCheckVersionsCompareNonNilMsg(t *testing.T) {
	storage := &storageMock{
		testcase: "Compare returns non-nil message",
	}

	experr := "Migration needed: please backup cgr data and run :"
	err := CheckVersions(storage)

	if err == nil || !strings.Contains(err.Error(), experr) {
		t.Errorf("expected error to be of type: <Migration needed: please backup cgr data and run : ..., \n received: <%+v>", err)
	}
}
