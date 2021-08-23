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

//because we import dispatchers in APIerSv1 we will send information as map[string]interface{}
func testDspApierSetAttributes(t *testing.T) {
	ev := &map[string]interface{}{
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
