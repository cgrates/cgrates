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
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

var testMap = UserMap{
	table: map[string]map[string]string{
		"test:user":   map[string]string{"t": "v"},
		":user":       map[string]string{"t": "v"},
		"test:":       map[string]string{"t": "v"},
		"test1:user1": map[string]string{"t": "v", "x": "y"},
		"test:masked": map[string]string{"t": "v"},
	},
	index: make(map[string]map[string]bool),
	properties: map[string]*prop{
		"test:masked": &prop{masked: true},
	},
}

var testMap2 = UserMap{
	table: map[string]map[string]string{
		"an:u1": map[string]string{"a": "b", "c": "d"},
		"an:u2": map[string]string{"a": "b"},
	},
	index: make(map[string]map[string]bool),
	properties: map[string]*prop{
		"an:u2": &prop{weight: 10},
	},
}

func TestUsersAdd(t *testing.T) {
	tm := newUserMap(dm, nil)
	var r string
	up := &UserProfile{
		Tenant:   "test",
		UserName: "user",
		Profile: map[string]string{
			"t": "v",
		},
	}
	tm.SetUser(up, &r)
	p, found := tm.table[up.GetId()]
	if r != utils.OK ||
		!found ||
		p["t"] != "v" ||
		len(tm.table) != 1 ||
		len(p) != 1 {
		t.Error("Error setting user: ", tm, len(tm.table))
	}
}

func TestUsersUpdate(t *testing.T) {
	tm := newUserMap(dm, nil)
	var r string
	up := &UserProfile{
		Tenant:   "test",
		UserName: "user",
		Profile: map[string]string{
			"t": "v",
		},
	}
	tm.SetUser(up, &r)
	p, found := tm.table[up.GetId()]
	if r != utils.OK ||
		!found ||
		p["t"] != "v" ||
		len(tm.table) != 1 ||
		len(p) != 1 {
		t.Error("Error setting user: ", tm)
	}
	up.Profile["x"] = "y"
	tm.UpdateUser(up, &r)
	p, found = tm.table[up.GetId()]
	if r != utils.OK ||
		!found ||
		p["x"] != "y" ||
		len(tm.table) != 1 ||
		len(p) != 2 {
		t.Error("Error updating user: ", tm)
	}
}

func TestUsersUpdateNotFound(t *testing.T) {
	tm := newUserMap(dm, nil)
	var r string
	up := &UserProfile{
		Tenant:   "test",
		UserName: "user",
		Profile: map[string]string{
			"t": "v",
		},
	}
	tm.SetUser(up, &r)
	up.UserName = "test1"
	err = tm.UpdateUser(up, &r)
	if err != utils.ErrNotFound {
		t.Error("Error detecting user not found on update: ", err)
	}
}

func TestUsersUpdateInit(t *testing.T) {
	tm := newUserMap(dm, nil)
	var r string
	up := &UserProfile{
		Tenant:   "test",
		UserName: "user",
	}
	tm.SetUser(up, &r)
	up = &UserProfile{
		Tenant:   "test",
		UserName: "user",
		Profile: map[string]string{
			"t": "v",
		},
	}
	tm.UpdateUser(up, &r)
	p, found := tm.table[up.GetId()]
	if r != utils.OK ||
		!found ||
		p["t"] != "v" ||
		len(tm.table) != 1 ||
		len(p) != 1 {
		t.Error("Error updating user: ", tm)
	}
}

func TestUsersRemove(t *testing.T) {
	tm := newUserMap(dm, nil)
	var r string
	up := &UserProfile{
		Tenant:   "test",
		UserName: "user",
		Profile: map[string]string{
			"t": "v",
		},
	}
	tm.SetUser(up, &r)
	p, found := tm.table[up.GetId()]
	if r != utils.OK ||
		!found ||
		p["t"] != "v" ||
		len(tm.table) != 1 ||
		len(p) != 1 {
		t.Error("Error setting user: ", tm)
	}
	tm.RemoveUser(up, &r)
	p, found = tm.table[up.GetId()]
	if r != utils.OK ||
		found ||
		len(tm.table) != 0 {
		t.Error("Error removing user: ", tm)
	}
}

func TestUsersGetFull(t *testing.T) {
	up := &UserProfile{
		Tenant:   "test",
		UserName: "user",
		Profile: map[string]string{
			"t": "v",
		},
	}
	results := UserProfiles{}
	testMap.GetUsers(up, &results)
	if len(results) != 3 {
		t.Error("error getting users: ", results)
	}
}

func TestUsersGetFullMasked(t *testing.T) {
	up := &UserProfile{
		Tenant: "test",
	}
	results := UserProfiles{}
	testMap.GetUsers(up, &results)
	if len(results) != 3 {
		t.Error("error getting users: ", results)
	}
}

func TestUsersGetFullUnMasked(t *testing.T) {
	up := &UserProfile{
		Tenant: "test",
		Masked: true,
	}
	results := UserProfiles{}
	testMap.GetUsers(up, &results)
	if len(results) != 4 {
		for _, r := range results {
			t.Logf("U: %+v", r)
		}
		t.Error("error getting users: ", results)
	}
}

func TestUsersGetTenant(t *testing.T) {
	up := &UserProfile{
		Tenant:   "testX",
		UserName: "user",
		Profile: map[string]string{
			"t": "v",
		},
	}
	results := UserProfiles{}
	testMap.GetUsers(up, &results)
	if len(results) != 1 {
		t.Error("error getting users: ", results)
	}
}

func TestUsersGetUserName(t *testing.T) {
	up := &UserProfile{
		Tenant:   "test",
		UserName: "userX",
		Profile: map[string]string{
			"t": "v",
		},
	}
	results := UserProfiles{}
	testMap.GetUsers(up, &results)
	if len(results) != 1 {
		t.Error("error getting users: ", results)
	}
}

func TestUsersGetNotFoundProfile(t *testing.T) {
	up := &UserProfile{
		Tenant:   "test",
		UserName: "user",
		Profile: map[string]string{
			"o": "p",
		},
	}
	results := UserProfiles{}
	testMap.GetUsers(up, &results)
	if len(results) != 3 {
		t.Error("error getting users: ", results)
	}
}

func TestUsersGetMissingTenant(t *testing.T) {
	up := &UserProfile{
		UserName: "user",
		Profile: map[string]string{
			"t": "v",
		},
	}
	results := UserProfiles{}
	testMap.GetUsers(up, &results)
	if len(results) != 3 {
		t.Error("error getting users: ", results)
	}
}

func TestUsersGetMissingUserName(t *testing.T) {
	up := &UserProfile{
		Tenant: "test",
		Profile: map[string]string{
			"t": "v",
		},
	}
	results := UserProfiles{}
	testMap.GetUsers(up, &results)
	if len(results) != 3 {
		t.Error("error getting users: ", results)
	}
}

func TestUsersGetMissingId(t *testing.T) {
	up := &UserProfile{
		Profile: map[string]string{
			"t": "v",
		},
	}
	results := UserProfiles{}
	testMap.GetUsers(up, &results)
	if len(results) != 4 {
		t.Error("error getting users: ", results)
	}
}

func TestUsersGetMissingIdTwo(t *testing.T) {
	up := &UserProfile{
		Profile: map[string]string{
			"t": "v",
			"x": "y",
		},
	}
	results := UserProfiles{}
	testMap.GetUsers(up, &results)
	if len(results) != 4 {
		t.Error("error getting users: ", results)
	}
}

func TestUsersGetMissingIdTwoSort(t *testing.T) {
	up := &UserProfile{
		Profile: map[string]string{
			"t": "v",
			"x": "y",
		},
	}
	results := UserProfiles{}
	testMap.GetUsers(up, &results)
	if len(results) != 4 {
		t.Error("error getting users: ", results)
	}
	if results[0].GetId() != "test1:user1" {
		t.Errorf("Error sorting profiles: %+v", results[0])
	}
}

func TestUsersGetMissingIdTwoSortWeight(t *testing.T) {
	up := &UserProfile{
		Profile: map[string]string{
			"a": "b",
			"c": "d",
		},
	}
	results := UserProfiles{}
	testMap2.GetUsers(up, &results)
	if len(results) != 2 {
		t.Error("error getting users: ", results)
	}
	if results[0].GetId() != "an:u2" {
		t.Errorf("Error sorting profiles: %+v", results[0])
	}
}

func TestUsersAddIndex(t *testing.T) {
	var r string
	testMap.AddIndex([]string{"t"}, &r)
	if r != utils.OK ||
		len(testMap.index) != 1 ||
		len(testMap.index[utils.ConcatenatedKey("t", "v")]) != 5 {
		t.Error("error adding index: ", testMap.index)
	}
}

func TestUsersAddIndexFull(t *testing.T) {
	var r string
	testMap.index = make(map[string]map[string]bool) // reset index
	testMap.AddIndex([]string{"t", "x", "UserName", "Tenant"}, &r)
	if r != utils.OK ||
		len(testMap.index) != 7 ||
		len(testMap.index[utils.ConcatenatedKey("t", "v")]) != 5 {
		t.Error("error adding index: ", testMap.index)
	}
}

func TestUsersAddIndexNone(t *testing.T) {
	var r string
	testMap.index = make(map[string]map[string]bool) // reset index
	testMap.AddIndex([]string{"test"}, &r)
	if r != utils.OK ||
		len(testMap.index) != 0 {
		t.Error("error adding index: ", testMap.index)
	}
}

func TestUsersGetFullindex(t *testing.T) {
	var r string
	testMap.index = make(map[string]map[string]bool) // reset index
	testMap.AddIndex([]string{"t", "x", "UserName", "Tenant"}, &r)
	up := &UserProfile{
		Tenant:   "test",
		UserName: "user",
		Profile: map[string]string{
			"t": "v",
		},
	}
	results := UserProfiles{}
	testMap.GetUsers(up, &results)
	if len(results) != 3 {
		t.Error("error getting users: ", results)
	}
}

func TestUsersGetTenantindex(t *testing.T) {
	var r string
	testMap.index = make(map[string]map[string]bool) // reset index
	testMap.AddIndex([]string{"t", "x", "UserName", "Tenant"}, &r)
	up := &UserProfile{
		Tenant:   "testX",
		UserName: "user",
		Profile: map[string]string{
			"t": "v",
		},
	}
	results := UserProfiles{}
	testMap.GetUsers(up, &results)
	if len(results) != 1 {
		t.Error("error getting users: ", results)
	}
}

func TestUsersGetUserNameindex(t *testing.T) {
	var r string
	testMap.index = make(map[string]map[string]bool) // reset index
	testMap.AddIndex([]string{"t", "x", "UserName", "Tenant"}, &r)
	up := &UserProfile{
		Tenant:   "test",
		UserName: "userX",
		Profile: map[string]string{
			"t": "v",
		},
	}
	results := UserProfiles{}
	testMap.GetUsers(up, &results)
	if len(results) != 1 {
		t.Error("error getting users: ", results)
	}
}

func TestUsersGetNotFoundProfileindex(t *testing.T) {
	var r string
	testMap.index = make(map[string]map[string]bool) // reset index
	testMap.AddIndex([]string{"t", "x", "UserName", "Tenant"}, &r)
	up := &UserProfile{
		Tenant:   "test",
		UserName: "user",
		Profile: map[string]string{
			"o": "p",
		},
	}
	results := UserProfiles{}
	testMap.GetUsers(up, &results)
	if len(results) != 3 {
		t.Error("error getting users: ", results)
	}
}

func TestUsersGetMissingTenantindex(t *testing.T) {
	var r string
	testMap.index = make(map[string]map[string]bool) // reset index
	testMap.AddIndex([]string{"t", "x", "UserName", "Tenant"}, &r)
	up := &UserProfile{
		UserName: "user",
		Profile: map[string]string{
			"t": "v",
		},
	}
	results := UserProfiles{}
	testMap.GetUsers(up, &results)
	if len(results) != 3 {
		t.Error("error getting users: ", results)
	}
}

func TestUsersGetMissingUserNameindex(t *testing.T) {
	var r string
	testMap.index = make(map[string]map[string]bool) // reset index
	testMap.AddIndex([]string{"t", "x", "UserName", "Tenant"}, &r)
	up := &UserProfile{
		Tenant: "test",
		Profile: map[string]string{
			"t": "v",
		},
	}
	results := UserProfiles{}
	testMap.GetUsers(up, &results)
	if len(results) != 3 {
		t.Error("error getting users: ", results)
	}
}

func TestUsersGetMissingIdindex(t *testing.T) {
	var r string
	testMap.index = make(map[string]map[string]bool) // reset index
	testMap.AddIndex([]string{"t", "x", "UserName", "Tenant"}, &r)
	up := &UserProfile{
		Profile: map[string]string{
			"t": "v",
		},
	}
	results := UserProfiles{}
	testMap.GetUsers(up, &results)
	if len(results) != 4 {
		t.Error("error getting users: ", results)
	}
}

func TestUsersGetMissingIdTwoINdex(t *testing.T) {
	var r string
	testMap.index = make(map[string]map[string]bool) // reset index
	testMap.AddIndex([]string{"t", "x", "UserName", "Tenant"}, &r)
	up := &UserProfile{
		Profile: map[string]string{
			"t": "v",
			"x": "y",
		},
	}
	results := UserProfiles{}
	testMap.GetUsers(up, &results)
	if len(results) != 4 {
		t.Error("error getting users: ", results)
	}
}

func TestUsersAddUpdateRemoveIndexes(t *testing.T) {
	tm := newUserMap(dm, nil)
	var r string
	tm.AddIndex([]string{"t"}, &r)
	if len(tm.index) != 0 {
		t.Error("error adding indexes: ", tm.index)
	}
	tm.SetUser(&UserProfile{
		Tenant:   "test",
		UserName: "user",
		Profile: map[string]string{
			"t": "v",
		},
	}, &r)
	if len(tm.index) != 1 || !tm.index["t:v"]["test:user"] {
		t.Error("error adding indexes: ", tm.index)
	}
	tm.SetUser(&UserProfile{
		Tenant:   "test",
		UserName: "best",
		Profile: map[string]string{
			"t": "v",
		},
	}, &r)
	if len(tm.index) != 1 ||
		!tm.index["t:v"]["test:user"] ||
		!tm.index["t:v"]["test:best"] {
		t.Error("error adding indexes: ", tm.index)
	}
	tm.UpdateUser(&UserProfile{
		Tenant:   "test",
		UserName: "best",
		Profile: map[string]string{
			"t": "v1",
		},
	}, &r)
	if len(tm.index) != 2 ||
		!tm.index["t:v"]["test:user"] ||
		!tm.index["t:v1"]["test:best"] {
		t.Error("error adding indexes: ", tm.index)
	}
	tm.UpdateUser(&UserProfile{
		Tenant:   "test",
		UserName: "best",
		Profile: map[string]string{
			"t": "v",
		},
	}, &r)
	if len(tm.index) != 1 ||
		!tm.index["t:v"]["test:user"] ||
		!tm.index["t:v"]["test:best"] {
		t.Error("error adding indexes: ", tm.index)
	}
	tm.RemoveUser(&UserProfile{
		Tenant:   "test",
		UserName: "best",
		Profile: map[string]string{
			"t": "v",
		},
	}, &r)
	if len(tm.index) != 1 ||
		!tm.index["t:v"]["test:user"] ||
		tm.index["t:v"]["test:best"] {
		t.Error("error adding indexes: ", tm.index)
	}
	tm.RemoveUser(&UserProfile{
		Tenant:   "test",
		UserName: "user",
		Profile: map[string]string{
			"t": "v",
		},
	}, &r)
	if len(tm.index) != 0 {
		t.Error("error adding indexes: ", tm.index)
	}
}

func TestUsersUsageRecordGetLoadUserProfile(t *testing.T) {
	userService = &UserMap{
		table: map[string]map[string]string{
			"test:user":   map[string]string{utils.ToR: "01", "RequestType": "1", "Direction": "*out", "Category": "c1", "Account": "dan", "Subject": "0723", "Destination": "+401", "SetupTime": "s1", "AnswerTime": "t1", "Usage": "10"},
			":user":       map[string]string{utils.ToR: "02", "RequestType": "2", "Direction": "*out", "Category": "c2", "Account": "ivo", "Subject": "0724", "Destination": "+402", "SetupTime": "s2", "AnswerTime": "t2", "Usage": "11"},
			"test:":       map[string]string{utils.ToR: "03", "RequestType": "3", "Direction": "*out", "Category": "c3", "Account": "elloy", "Subject": "0725", "Destination": "+403", "SetupTime": "s3", "AnswerTime": "t3", "Usage": "12"},
			"test1:user1": map[string]string{utils.ToR: "04", "RequestType": "4", "Direction": "*out", "Category": "call", "Account": "rif", "Subject": "0726", "Destination": "+404", "SetupTime": "s4", "AnswerTime": "t4", "Usage": "13"},
		},
		index: make(map[string]map[string]bool),
	}

	ur := &UsageRecord{
		ToR:         utils.USERS,
		RequestType: utils.USERS,
		Tenant:      "",
		Category:    "call",
		Account:     utils.USERS,
		Subject:     utils.USERS,
		Destination: utils.USERS,
		SetupTime:   utils.USERS,
		AnswerTime:  utils.USERS,
		Usage:       "13",
	}

	err := LoadUserProfile(ur, "")
	if err != nil {
		t.Error("Error loading user profile: ", err)
	}
	expected := &UsageRecord{
		ToR:         "04",
		RequestType: "4",
		Tenant:      "",
		Category:    "call",
		Account:     "rif",
		Subject:     "0726",
		Destination: "+404",
		SetupTime:   "s4",
		AnswerTime:  "t4",
		Usage:       "13",
	}
	if !reflect.DeepEqual(ur, expected) {
		t.Errorf("Expected: %+v got: %+v", expected, ur)
	}
}

func TestUsersExternalCDRGetLoadUserProfileExtraFields(t *testing.T) {
	userService = &UserMap{
		table: map[string]map[string]string{
			"test:user":   map[string]string{utils.ToR: "01", "RequestType": "1", "Direction": "*out", "Category": "c1", "Account": "dan", "Subject": "0723", "Destination": "+401", "SetupTime": "s1", "AnswerTime": "t1", "Usage": "10"},
			":user":       map[string]string{utils.ToR: "02", "RequestType": "2", "Direction": "*out", "Category": "c2", "Account": "ivo", "Subject": "0724", "Destination": "+402", "SetupTime": "s2", "AnswerTime": "t2", "Usage": "11"},
			"test:":       map[string]string{utils.ToR: "03", "RequestType": "3", "Direction": "*out", "Category": "c3", "Account": "elloy", "Subject": "0725", "Destination": "+403", "SetupTime": "s3", "AnswerTime": "t3", "Usage": "12"},
			"test1:user1": map[string]string{utils.ToR: "04", "RequestType": "4", "Direction": "*out", "Category": "call", "Account": "rif", "Subject": "0726", "Destination": "+404", "SetupTime": "s4", "AnswerTime": "t4", "Usage": "13", "Test": "1"},
		},
		index: make(map[string]map[string]bool),
	}

	ur := &ExternalCDR{
		ToR:         utils.USERS,
		RequestType: utils.USERS,
		Tenant:      "",
		Category:    "call",
		Account:     utils.USERS,
		Subject:     utils.USERS,
		Destination: utils.USERS,
		SetupTime:   utils.USERS,
		AnswerTime:  utils.USERS,
		Usage:       "13",
		ExtraFields: map[string]string{
			"Test": "1",
		},
	}

	err := LoadUserProfile(ur, "ExtraFields")
	if err != nil {
		t.Error("Error loading user profile: ", err)
	}
	expected := &ExternalCDR{
		ToR:         "04",
		RequestType: "4",
		Tenant:      "",
		Category:    "call",
		Account:     "rif",
		Subject:     "0726",
		Destination: "+404",
		SetupTime:   "s4",
		AnswerTime:  "t4",
		Usage:       "13",
		ExtraFields: map[string]string{
			"Test": "1",
		},
	}
	if !reflect.DeepEqual(ur, expected) {
		t.Errorf("Expected: %+v got: %+v", expected, ur)
	}
}

func TestUsersExternalCDRGetLoadUserProfileExtraFieldsNotFound(t *testing.T) {
	userService = &UserMap{
		table: map[string]map[string]string{
			"test:user":   map[string]string{utils.ToR: "01", "RequestType": "1", "Direction": "*out", "Category": "c1", "Account": "dan", "Subject": "0723", "Destination": "+401", "SetupTime": "s1", "AnswerTime": "t1", "Usage": "10"},
			":user":       map[string]string{utils.ToR: "02", "RequestType": "2", "Direction": "*out", "Category": "c2", "Account": "ivo", "Subject": "0724", "Destination": "+402", "SetupTime": "s2", "AnswerTime": "t2", "Usage": "11"},
			"test:":       map[string]string{utils.ToR: "03", "RequestType": "3", "Direction": "*out", "Category": "c3", "Account": "elloy", "Subject": "0725", "Destination": "+403", "SetupTime": "s3", "AnswerTime": "t3", "Usage": "12"},
			"test1:user1": map[string]string{utils.ToR: "04", "RequestType": "4", "Direction": "*out", "Category": "call", "Account": "rif", "Subject": "0726", "Destination": "+404", "SetupTime": "s4", "AnswerTime": "t4", "Usage": "13", "Test": "2"},
		},
		index: make(map[string]map[string]bool),
	}

	ur := &ExternalCDR{
		ToR:         utils.USERS,
		RequestType: utils.USERS,
		Tenant:      "",
		Category:    "call",
		Account:     utils.USERS,
		Subject:     utils.USERS,
		Destination: utils.USERS,
		SetupTime:   utils.USERS,
		AnswerTime:  utils.USERS,
		Usage:       "13",
		ExtraFields: map[string]string{
			"Test": "1",
		},
	}

	err := LoadUserProfile(ur, "ExtraFields")
	if err != utils.ErrUserNotFound {
		t.Error("Error detecting err in loading user profile: ", err)
	}
}

func TestUsersExternalCDRGetLoadUserProfileExtraFieldsSet(t *testing.T) {
	userService = &UserMap{
		table: map[string]map[string]string{
			"test:user":   map[string]string{utils.ToR: "01", "RequestType": "1", "Direction": "*out", "Category": "c1", "Account": "dan", "Subject": "0723", "Destination": "+401", "SetupTime": "s1", "AnswerTime": "t1", "Usage": "10"},
			":user":       map[string]string{utils.ToR: "02", "RequestType": "2", "Direction": "*out", "Category": "c2", "Account": "ivo", "Subject": "0724", "Destination": "+402", "SetupTime": "s2", "AnswerTime": "t2", "Usage": "11"},
			"test:":       map[string]string{utils.ToR: "03", "RequestType": "3", "Direction": "*out", "Category": "c3", "Account": "elloy", "Subject": "0725", "Destination": "+403", "SetupTime": "s3", "AnswerTime": "t3", "Usage": "12"},
			"test1:user1": map[string]string{utils.ToR: "04", "RequestType": "4", "Direction": "*out", "Category": "call", "Account": "rif", "Subject": "0726", "Destination": "+404", "SetupTime": "s4", "AnswerTime": "t4", "Usage": "13", "Test": "1", "Best": "BestValue"},
		},
		index: make(map[string]map[string]bool),
	}

	ur := &ExternalCDR{
		ToR:         utils.USERS,
		RequestType: utils.USERS,
		Tenant:      "",
		Category:    "call",
		Account:     utils.USERS,
		Subject:     utils.USERS,
		Destination: utils.USERS,
		SetupTime:   utils.USERS,
		AnswerTime:  utils.USERS,
		Usage:       "13",
		ExtraFields: map[string]string{
			"Test": "1",
			"Best": utils.USERS,
		},
	}

	err := LoadUserProfile(ur, "ExtraFields")
	if err != nil {
		t.Error("Error loading user profile: ", err)
	}
	expected := &ExternalCDR{
		ToR:         "04",
		RequestType: "4",
		Tenant:      "",
		Category:    "call",
		Account:     "rif",
		Subject:     "0726",
		Destination: "+404",
		SetupTime:   "s4",
		AnswerTime:  "t4",
		Usage:       "13",
		ExtraFields: map[string]string{
			"Test": "1",
			"Best": "BestValue",
		},
	}
	if !reflect.DeepEqual(ur, expected) {
		t.Errorf("Expected: %+v got: %+v", expected, ur)
	}
}

func TestUsersCallDescLoadUserProfile(t *testing.T) {
	userService = &UserMap{
		table: map[string]map[string]string{
			"cgrates.org:dan":      map[string]string{"RequestType": "*prepaid", "Category": "call1", "Account": "dan", "Subject": "dan", "Cli": "+4986517174963"},
			"cgrates.org:danvoice": map[string]string{utils.ToR: "*voice", "RequestType": "*prepaid", "Category": "call1", "Account": "dan", "Subject": "0723"},
			"cgrates:rif":          map[string]string{"RequestType": "*postpaid", "Direction": "*out", "Category": "call", "Account": "rif", "Subject": "0726"},
		},
		index: make(map[string]map[string]bool),
	}
	startTime := time.Now()
	cd := &CallDescriptor{
		TOR:         "*sms",
		Tenant:      utils.USERS,
		Category:    utils.USERS,
		Subject:     utils.USERS,
		Account:     utils.USERS,
		Destination: "+4986517174963",
		TimeStart:   startTime,
		TimeEnd:     startTime.Add(time.Duration(1) * time.Minute),
		ExtraFields: map[string]string{"Cli": "+4986517174963"},
	}
	expected := &CallDescriptor{
		TOR:         "*sms",
		Tenant:      "cgrates.org",
		Category:    "call1",
		Account:     "dan",
		Subject:     "dan",
		Destination: "+4986517174963",
		TimeStart:   startTime,
		TimeEnd:     startTime.Add(time.Duration(1) * time.Minute),
		ExtraFields: map[string]string{"Cli": "+4986517174963"},
	}
	err := LoadUserProfile(cd, "ExtraFields")
	if err != nil {
		t.Error("Error loading user profile: ", err)
	}
	if !reflect.DeepEqual(expected, cd) {
		t.Errorf("Expected: %+v got: %+v", expected, cd)
	}
}

func TestUsersCDRLoadUserProfile(t *testing.T) {
	userService = &UserMap{
		table: map[string]map[string]string{
			"cgrates.org:dan":      map[string]string{"RequestType": "*prepaid", "Category": "call1", "Account": "dan", "Subject": "dan", "Cli": "+4986517174963"},
			"cgrates.org:danvoice": map[string]string{utils.ToR: "*voice", "RequestType": "*prepaid", "Category": "call1", "Account": "dan", "Subject": "0723"},
			"cgrates:rif":          map[string]string{"RequestType": "*postpaid", "Direction": "*out", "Category": "call", "Account": "rif", "Subject": "0726"},
		},
		index: make(map[string]map[string]bool),
	}
	startTime := time.Now()
	cdr := &CDR{
		ToR:         "*sms",
		RequestType: utils.USERS,
		Tenant:      utils.USERS,
		Category:    utils.USERS,
		Account:     utils.USERS,
		Subject:     utils.USERS,
		Destination: "+4986517174963",
		SetupTime:   startTime,
		AnswerTime:  startTime,
		Usage:       time.Duration(1) * time.Minute,
		ExtraFields: map[string]string{"Cli": "+4986517174963"},
	}
	expected := &CDR{
		ToR:         "*sms",
		RequestType: "*prepaid",
		Tenant:      "cgrates.org",
		Category:    "call1",
		Account:     "dan",
		Subject:     "dan",
		Destination: "+4986517174963",
		SetupTime:   startTime,
		AnswerTime:  startTime,
		Usage:       time.Duration(1) * time.Minute,
		ExtraFields: map[string]string{"Cli": "+4986517174963"},
	}
	err := LoadUserProfile(cdr, "ExtraFields")
	if err != nil {
		t.Error("Error loading user profile: ", err)
	}
	if !reflect.DeepEqual(expected, cdr) {
		t.Errorf("Expected: %+v got: %+v", expected, cdr)
	}
}
