/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or56
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT MetaAny WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package engine

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestModelsTimingMdlTableName(t *testing.T) {
	testStruct := TimingMdl{}
	exp := utils.TBLTPTimings
	result := testStruct.TableName()
	if !reflect.DeepEqual(exp, result) {
		t.Errorf("\nExpected <%+v>,\nreceived <%+v>", exp, result)
	}
}

func TestModelsDestinationMdlTableName(t *testing.T) {
	testStruct := DestinationMdl{}
	exp := utils.TBLTPDestinations
	result := testStruct.TableName()
	if !reflect.DeepEqual(exp, result) {
		t.Errorf("\nExpected <%+v>,\nreceived <%+v>", exp, result)
	}
}

func TestModelsRateMdlTableName(t *testing.T) {
	testStruct := RateMdl{}
	exp := utils.TBLTPRates
	result := testStruct.TableName()
	if !reflect.DeepEqual(exp, result) {
		t.Errorf("\nExpected <%+v>,\nreceived <%+v>", exp, result)
	}
}

func TestModelsDestinationRateMdlTableName(t *testing.T) {
	testStruct := DestinationRateMdl{}
	exp := utils.TBLTPDestinationRates
	result := testStruct.TableName()
	if !reflect.DeepEqual(exp, result) {
		t.Errorf("\nExpected <%+v>,\nreceived <%+v>", exp, result)
	}
}

func TestModelsRatingPlanMdlTableName(t *testing.T) {
	testStruct := RatingPlanMdl{}
	exp := utils.TBLTPRatingPlans
	result := testStruct.TableName()
	if !reflect.DeepEqual(exp, result) {
		t.Errorf("\nExpected <%+v>,\nreceived <%+v>", exp, result)
	}
}

func TestModelsRatingProfileMdlTableName(t *testing.T) {
	testStruct := RatingProfileMdl{}
	exp := utils.TBLTPRatingProfiles
	result := testStruct.TableName()
	if !reflect.DeepEqual(exp, result) {
		t.Errorf("\nExpected <%+v>,\nreceived <%+v>", exp, result)
	}
}

func TestModelsActionMdlTableName(t *testing.T) {
	testStruct := ActionMdl{}
	exp := utils.TBLTPActions
	result := testStruct.TableName()
	if !reflect.DeepEqual(exp, result) {
		t.Errorf("\nExpected <%+v>,\nreceived <%+v>", exp, result)
	}
}

func TestModelsActionPlanMdlTableName(t *testing.T) {
	testStruct := ActionPlanMdl{}
	exp := utils.TBLTPActionPlans
	result := testStruct.TableName()
	if !reflect.DeepEqual(exp, result) {
		t.Errorf("\nExpected <%+v>,\nreceived <%+v>", exp, result)
	}
}

func TestModelsActionTriggerMdlTableName(t *testing.T) {
	testStruct := ActionTriggerMdl{}
	exp := utils.TBLTPActionTriggers
	result := testStruct.TableName()
	if !reflect.DeepEqual(exp, result) {
		t.Errorf("\nExpected <%+v>,\nreceived <%+v>", exp, result)
	}
}

func TestModelsAccountActionMdlTableName(t *testing.T) {
	testStruct := AccountActionMdl{}
	exp := utils.TBLTPAccountActions
	result := testStruct.TableName()
	if !reflect.DeepEqual(exp, result) {
		t.Errorf("\nExpected <%+v>,\nreceived <%+v>", exp, result)
	}
}

func TestModelsSetAccountActionId(t *testing.T) {
	testStruct := &AccountActionMdl{
		Id:      0,
		Loadid:  "",
		Tenant:  "",
		Account: "",
	}

	err := testStruct.SetAccountActionId("id1:id2:id3")
	if err != nil {
		t.Errorf("\nExpected <%+v>,\nreceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual("id1", testStruct.Loadid) {
		t.Errorf("\nExpected <%+v>,\nreceived <%+v>", "id1", testStruct.Loadid)
	}
	if !reflect.DeepEqual("id2", testStruct.Tenant) {
		t.Errorf("\nExpected <%+v>,\nreceived <%+v>", "id2", testStruct.Tenant)
	}
	if !reflect.DeepEqual("id3", testStruct.Account) {
		t.Errorf("\nExpected <%+v>,\nreceived <%+v>", "id3", testStruct.Account)
	}

}

func TestModelsSetAccountActionIdError(t *testing.T) {
	testStruct := &AccountActionMdl{
		Id:      0,
		Loadid:  "",
		Tenant:  "",
		Account: "",
	}

	err := testStruct.SetAccountActionId("id1;id2;id3")
	if err == nil || err.Error() != "Wrong TP Account Action Id: id1;id2;id3" {
		t.Errorf("\nExpected <Wrong TP Account Action Id: id1;id2;id3>,\nreceived <%+v>", err)
	}

}

func TestModelGetAccountActionId(t *testing.T) {
	testStruct := &AccountActionMdl{
		Id:      0,
		Tpid:    "",
		Loadid:  "",
		Tenant:  "id1",
		Account: "id2",
	}
	result := testStruct.GetAccountActionId()
	exp := utils.ConcatenatedKey(testStruct.Tenant, testStruct.Account)
	if !reflect.DeepEqual(exp, result) {
		t.Errorf("\nExpected <%+v> ,\nreceived <%+v>", exp, result)
	}
}
