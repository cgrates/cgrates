//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package dispatchers

// for the moment we dispable Apier through dispatcher
// until we figured out a better sollution in case of gob server
/*
import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var sTestsDspApier = []func(t *testing.T){
	testDspApierSetAttributes,
	testDspApierGetAttributes,
	testDspApierUnkownAPiKey,
}

//Test start here
func TestDspApierITMySQL(t *testing.T) {
	testDsp(t, sTestsDspApier, "TestDspApier", "all", "all2", "dispatchers_mysql", "tutorial", "oldtutorial", "dispatchers")
}

func TestDspApierITMongo(t *testing.T) {
	testDsp(t, sTestsDspApier, "TestDspApier", "all", "all2", "dispatchers_mongo", "tutorial", "oldtutorial", "dispatchers")
}

//because we import dispatchers in APIerSv1 we will send information as map[string]any
func testDspApierSetAttributes(t *testing.T) {
	ev := &map[string]any{
		utils.Tenant: "cgrates.org",
		"ID":         "ATTR_Dispatcher",
		"Contexts":   []string{utils.MetaSessionS},
		"FilterIDs":  []string{"*string:~Account:1234"},
		"ActivationInterval": &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		"Attributes": []*engine.Attribute{
			{
				Path: utils.MetaReq + utils.NestingSep + utils.Subject,
				Value: config.RSRParsers{
					&config.RSRParser{
						Rules:           "roam",
						AllFiltersMatch: true,
					},
				},
			},
		},
		"Weight":     10,
		utils.APIKey: utils.StringPointer("apier12345"),
	}
	var result string
	if err := dispEngine.RPC.Call(utils.APIerSv1SetAttributeProfile, ev, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

}

func testDspApierGetAttributes(t *testing.T) {
	var reply *engine.AttributeProfile
	alsPrf := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_Dispatcher",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.Account:1234"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		Attributes: []*engine.Attribute{
			{
				Path: utils.MetaReq + utils.NestingSep +  utils.Subject,
				Value: config.RSRParsers{
					&config.RSRParser{
						Rules:           "roam",
						AllFiltersMatch: true,
					},
				},
			},
		},
		Weight: 10,
	}
	alsPrf.Compile()
	if err := dispEngine.RPC.Call(utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithArgDispatcher{
			TenantID:      &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_Dispatcher"},
			ArgDispatcher: &utils.ArgDispatcher{APIKey: utils.StringPointer("apier12345")},
		}, &reply); err != nil {
		t.Fatal(err)
	}
	reply.Compile()
	if !reflect.DeepEqual(alsPrf, reply) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf, reply)
	}

}

func testDspApierUnkownAPiKey(t *testing.T) {
	var reply *engine.AttributeProfile
	if err := dispEngine.RPC.Call(utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithArgDispatcher{
			TenantID:      &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_Dispatcher"},
			ArgDispatcher: &utils.ArgDispatcher{APIKey: utils.StringPointer("RandomApiKey")},
		}, &reply); err == nil || err.Error() != utils.ErrUnknownApiKey.Error() {
		t.Fatal(err)
	}
}
*/
