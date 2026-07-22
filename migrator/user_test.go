// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrator

import (
	"reflect"
	"sort"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestUserProfile2attributeProfile(t *testing.T) {
	usrCfgIn := config.CgrConfig()
	usrCfgIn.MigratorCgrCfg().UsersFilters = []string{"Account"}
	config.SetCgrConfig(usrCfgIn)
	usrTenant := "cgrates.com"
	users := map[int]*v1UserProfile{
		0: {
			Tenant:   defaultTenant,
			UserName: "1001",
			Masked:   true,
			Profile:  map[string]string{},
			Weight:   10,
		},
		1: {
			Tenant:   usrTenant,
			UserName: "1001",
			Masked:   true,
			Profile: map[string]string{
				"Account": "1002",
				"Subject": "call_1001",
			},
			Weight: 10,
		},
		2: {
			Tenant:   defaultTenant,
			UserName: "1001",
			Masked:   false,
			Profile: map[string]string{
				"Account": "1002",
				"ReqType": "*prepaid",
				"msisdn":  "123423534646752",
			},
			Weight: 10,
		},
		3: {
			Tenant:   usrTenant,
			UserName: "1001",
			Masked:   false,
			Profile: map[string]string{
				"Account": "1002",
				"ReqType": "*prepaid",
			},
			Weight: 10,
		},
		4: {
			Tenant:   usrTenant,
			UserName: "acstmusername",
			Profile: map[string]string{
				"Account": "acnt63",
				"Subject": "acnt63",
				"ReqType": "*prepaid",
				"msisdn":  "12345",
				"imsi":    "12345",
			},
			Weight: 10,
		},
	}
	expected := map[int]*engine.AttributeProfile{
		0: {
			Tenant:             defaultTenant,
			ID:                 "1001",
			Contexts:           []string{utils.MetaAny},
			FilterIDs:          make([]string, 0),
			ActivationInterval: nil,
			Attributes:         []*engine.Attribute{},
			Blocker:            false,
			Weight:             10,
		},
		1: {
			Tenant:             defaultTenant,
			ID:                 "1001",
			Contexts:           []string{utils.MetaAny},
			FilterIDs:          []string{"*string:~*req.Account:1002"},
			ActivationInterval: nil,
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Subject",
					Type:  utils.MetaVariable,
					Value: config.NewRSRParsersMustCompile("call_1001", utils.InfieldSep),
				},
				{
					Path:  utils.MetaTenant,
					Type:  utils.MetaConstant,
					Value: config.NewRSRParsersMustCompile(usrTenant, utils.InfieldSep),
				},
			},
			Blocker: false,
			Weight:  10,
		},
		2: {
			Tenant:   defaultTenant,
			ID:       "1001",
			Contexts: []string{utils.MetaAny},
			FilterIDs: []string{
				"*string:~*req.Account:1002",
			},
			ActivationInterval: nil,
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + utils.RequestType,
					Type:  utils.MetaVariable,
					Value: config.NewRSRParsersMustCompile("*prepaid", utils.InfieldSep),
				},
				{
					Path:  utils.MetaReq + utils.NestingSep + "msisdn",
					Type:  utils.MetaVariable,
					Value: config.NewRSRParsersMustCompile("123423534646752", utils.InfieldSep),
				},
			},
			Blocker: false,
			Weight:  10,
		},
		3: {
			Tenant:             defaultTenant,
			ID:                 "1001",
			Contexts:           []string{utils.MetaAny},
			FilterIDs:          []string{"*string:~*req.Account:1002"},
			ActivationInterval: nil,
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + utils.RequestType,
					Type:  utils.MetaVariable,
					Value: config.NewRSRParsersMustCompile("*prepaid", utils.InfieldSep),
				},
				{
					Path:  utils.MetaTenant,
					Type:  utils.MetaConstant,
					Value: config.NewRSRParsersMustCompile(usrTenant, utils.InfieldSep),
				},
			},
			Blocker: false,
			Weight:  10,
		},
		4: {
			Tenant:   defaultTenant,
			ID:       "acstmusername",
			Contexts: []string{utils.MetaAny},
			FilterIDs: []string{
				"*string:~*req.Account:acnt63",
			},
			ActivationInterval: nil,
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + utils.RequestType,
					Type:  utils.MetaVariable,
					Value: config.NewRSRParsersMustCompile("*prepaid", utils.InfieldSep),
				},
				{
					Path:  utils.MetaReq + utils.NestingSep + utils.Subject,
					Type:  utils.MetaVariable,
					Value: config.NewRSRParsersMustCompile("acnt63", utils.InfieldSep),
				},
				{
					Path:  utils.MetaReq + utils.NestingSep + "imsi",
					Type:  utils.MetaVariable,
					Value: config.NewRSRParsersMustCompile("12345", utils.InfieldSep),
				},
				{
					Path:  utils.MetaReq + utils.NestingSep + "msisdn",
					Type:  utils.MetaVariable,
					Value: config.NewRSRParsersMustCompile("12345", utils.InfieldSep),
				},
				{
					Path:  utils.MetaTenant,
					Type:  utils.MetaConstant,
					Value: config.NewRSRParsersMustCompile(usrTenant, utils.InfieldSep),
				},
			},
			Blocker: false,
			Weight:  10,
		},
	}
	for i := range expected {
		rply := userProfile2attributeProfile(users[i])
		sort.Slice(rply.Attributes, func(i, j int) bool {
			if rply.Attributes[i].Path == rply.Attributes[j].Path {
				return rply.Attributes[i].FilterIDs[0] < rply.Attributes[j].FilterIDs[0]
			}
			return rply.Attributes[i].Path < rply.Attributes[j].Path
		}) // only for test; map returns random keys
		if !reflect.DeepEqual(expected[i], rply) {
			t.Errorf("For %v expected: %s ,\nreceived: %s ", i, utils.ToJSON(expected[i]), utils.ToJSON(rply))
		}
	}
}

func TestMigratorGetId(t *testing.T) {
	testTenant := "cgrates.org"
	testUserName := "testUser"
	expectedID := utils.ConcatenatedKey(testTenant, testUserName)
	userProfile := &v1UserProfile{
		Tenant:   testTenant,
		UserName: testUserName,
		Masked:   true,
		Profile:  map[string]string{"key1": "value1"},
		Weight:   3.14,
	}
	actualID := userProfile.GetId()
	if actualID != expectedID {
		t.Errorf("Expected GetId() to return %s, got %s", expectedID, actualID)
	}
	expectedProfile := map[string]string{"key1": "value1"}
	if !reflect.DeepEqual(userProfile.Profile, expectedProfile) {
		t.Errorf("Expected Profile to remain unchanged, got %#v", userProfile.Profile)
	}
	if userProfile.Masked != true {
		t.Errorf("Expected Masked flag to remain true")
	}
	if userProfile.Weight != 3.14 {
		t.Errorf("Expected Weight to remain unchanged")
	}
}

func TestMigratorSetId(t *testing.T) {
	testTenant := "cgrates"
	sampleUserName := "usertest123"
	validID := utils.ConcatenatedKey(testTenant, sampleUserName)
	invalidID := testTenant + sampleUserName
	var testCases = []struct {
		name    string
		id      string
		want    string
		wantErr error
	}{
		{
			name:    "Valid ID format",
			id:      validID,
			want:    testTenant,
			wantErr: nil,
		},
		{
			name:    "Invalid ID format",
			id:      invalidID,
			want:    "",
			wantErr: utils.ErrInvalidKey,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			userProfile := &v1UserProfile{}
			err := userProfile.SetId(tc.id)
			if err != tc.wantErr {
				t.Errorf("Expected error: %v, got: %v", tc.wantErr, err)
			}
			if userProfile.Tenant != tc.want {
				t.Errorf("Expected Tenant to be %s, got %s", tc.want, userProfile.Tenant)
			}
		})
	}
}
