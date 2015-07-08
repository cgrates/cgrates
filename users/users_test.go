package users

import (
	"testing"

	"github.com/cgrates/cgrates/utils"
)

var testMap = UserMap{
	"test:user":   map[string]string{"t": "v"},
	":user":       map[string]string{"t": "v"},
	"test:":       map[string]string{"t": "v"},
	"test1:user1": map[string]string{"t": "v", "x": "y"},
}

func TestUsersAdd(t *testing.T) {
	tm := UserMap{}
	var r string
	up := UserProfile{
		Tenant:   "test",
		UserName: "user",
		Profile: map[string]string{
			"t": "v",
		},
	}
	tm.SetUser(up, &r)
	p, found := tm[up.GetId()]
	if r != utils.OK ||
		!found ||
		p["t"] != "v" ||
		len(tm) != 1 ||
		len(p) != 1 {
		t.Error("Error setting user: ", tm)
	}
}

func TestUsersUpdate(t *testing.T) {
	tm := UserMap{}
	var r string
	up := UserProfile{
		Tenant:   "test",
		UserName: "user",
		Profile: map[string]string{
			"t": "v",
		},
	}
	tm.SetUser(up, &r)
	p, found := tm[up.GetId()]
	if r != utils.OK ||
		!found ||
		p["t"] != "v" ||
		len(tm) != 1 ||
		len(p) != 1 {
		t.Error("Error setting user: ", tm)
	}
	up.Profile["x"] = "y"
	tm.UpdateUser(up, &r)
	p, found = tm[up.GetId()]
	if r != utils.OK ||
		!found ||
		p["x"] != "y" ||
		len(tm) != 1 ||
		len(p) != 2 {
		t.Error("Error updating user: ", tm)
	}
}

func TestUsersUpdateNotFound(t *testing.T) {
	tm := UserMap{}
	var r string
	up := UserProfile{
		Tenant:   "test",
		UserName: "user",
		Profile: map[string]string{
			"t": "v",
		},
	}
	tm.SetUser(up, &r)
	up.UserName = "test1"
	err := tm.UpdateUser(up, &r)
	if err != utils.ErrNotFound {
		t.Error("Error detecting user not found on update: ", err)
	}
}

func TestUsersUpdateInit(t *testing.T) {
	tm := UserMap{}
	var r string
	up := UserProfile{
		Tenant:   "test",
		UserName: "user",
	}
	tm.SetUser(up, &r)
	up = UserProfile{
		Tenant:   "test",
		UserName: "user",
		Profile: map[string]string{
			"t": "v",
		},
	}
	tm.UpdateUser(up, &r)
	p, found := tm[up.GetId()]
	if r != utils.OK ||
		!found ||
		p["t"] != "v" ||
		len(tm) != 1 ||
		len(p) != 1 {
		t.Error("Error updating user: ", tm)
	}
}

func TestUsersRemove(t *testing.T) {
	tm := UserMap{}
	var r string
	up := UserProfile{
		Tenant:   "test",
		UserName: "user",
		Profile: map[string]string{
			"t": "v",
		},
	}
	tm.SetUser(up, &r)
	p, found := tm[up.GetId()]
	if r != utils.OK ||
		!found ||
		p["t"] != "v" ||
		len(tm) != 1 ||
		len(p) != 1 {
		t.Error("Error setting user: ", tm)
	}
	tm.RemoveUser(up, &r)
	p, found = tm[up.GetId()]
	if r != utils.OK ||
		found ||
		len(tm) != 0 {
		t.Error("Error removing user: ", tm)
	}
}

func TestUsersGetFull(t *testing.T) {
	up := UserProfile{
		Tenant:   "test",
		UserName: "user",
		Profile: map[string]string{
			"t": "v",
		},
	}
	results := make([]*UserProfile, 0)
	testMap.GetUsers(up, &results)
	if len(results) != 1 {
		t.Error("error getting users: ", results)
	}
}

func TestUsersGetTenant(t *testing.T) {
	up := UserProfile{
		Tenant:   "testX",
		UserName: "user",
		Profile: map[string]string{
			"t": "v",
		},
	}
	results := make([]*UserProfile, 0)
	testMap.GetUsers(up, &results)
	if len(results) != 0 {
		t.Error("error getting users: ", results)
	}
}

func TestUsersGetUserName(t *testing.T) {
	up := UserProfile{
		Tenant:   "test",
		UserName: "userX",
		Profile: map[string]string{
			"t": "v",
		},
	}
	results := make([]*UserProfile, 0)
	testMap.GetUsers(up, &results)
	if len(results) != 0 {
		t.Error("error getting users: ", results)
	}
}

func TestUsersGetNotFoundProfile(t *testing.T) {
	up := UserProfile{
		Tenant:   "test",
		UserName: "user",
		Profile: map[string]string{
			"o": "p",
		},
	}
	results := make([]*UserProfile, 0)
	testMap.GetUsers(up, &results)
	if len(results) != 0 {
		t.Error("error getting users: ", results)
	}
}

func TestUsersGetMissingTenant(t *testing.T) {
	up := UserProfile{
		UserName: "user",
		Profile: map[string]string{
			"t": "v",
		},
	}
	results := make([]*UserProfile, 0)
	testMap.GetUsers(up, &results)
	if len(results) != 2 {
		t.Error("error getting users: ", results)
	}
}

func TestUsersGetMissingUserName(t *testing.T) {
	up := UserProfile{
		Tenant: "test",
		Profile: map[string]string{
			"t": "v",
		},
	}
	results := make([]*UserProfile, 0)
	testMap.GetUsers(up, &results)
	if len(results) != 2 {
		t.Error("error getting users: ", results)
	}
}

func TestUsersGetMissingId(t *testing.T) {
	up := UserProfile{
		Profile: map[string]string{
			"t": "v",
		},
	}
	results := make([]*UserProfile, 0)
	testMap.GetUsers(up, &results)
	if len(results) != 4 {
		t.Error("error getting users: ", results)
	}
}

func TestUsersGetMissingIdTwo(t *testing.T) {
	up := UserProfile{
		Profile: map[string]string{
			"t": "v",
			"x": "y",
		},
	}
	results := make([]*UserProfile, 0)
	testMap.GetUsers(up, &results)
	if len(results) != 1 {
		t.Error("error getting users: ", results)
	}
}
