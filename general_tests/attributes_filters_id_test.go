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

package general_tests

import (
	"net/rpc"
	"path"
	"testing"

	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	attrFltrCfgPath     string
	attrFltrCfg         *config.CGRConfig
	attrFltrRPC         *rpc.Client
	alsPrfFltrConfigDIR string
	sTestsAlsFltrPrf    = []func(t *testing.T){
		testAttributeFltrSInitCfg,
		testAttributeFltrSInitDataDb,
		testAttributeFltrSResetStorDb,
		testAttributeFltrSStartEngine,
		testAttributeFltrSRPCConn,
		testAttributeFltrSetAttrProfileAndFltr,
		testAttributeFltrSStopEngine,
	}
)

func TestAttributeFilterSIT(t *testing.T) {
	switch *dbType {
	case utils.MetaMySQL:
		alsPrfFltrConfigDIR = "mysql_attributes"
		/*
			case utils.MetaMongo:
				alsPrfConfigDIR = "tutmongo"

		*/
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsAlsFltrPrf {
		t.Run(alsPrfFltrConfigDIR, stest)
	}
}

func testAttributeFltrSInitCfg(t *testing.T) {
	var err error
	attrFltrCfgPath = path.Join(*dataDir, "conf", "samples", alsPrfFltrConfigDIR)
	attrFltrCfg, err = config.NewCGRConfigFromPath(attrFltrCfgPath)
	if err != nil {
		t.Error(err)
	}
	attrFltrCfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
}

func testAttributeFltrSInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(attrFltrCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testAttributeFltrSResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(attrFltrCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testAttributeFltrSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(attrFltrCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testAttributeFltrSRPCConn(t *testing.T) {
	var err error
	attrFltrRPC, err = newRPCClient(attrFltrCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testAttributeFltrSetAttrProfileAndFltr(t *testing.T) {
	filter := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_1",
			Rules: []*engine.FilterRule{{
				Element: "~*req.Subject",
				Type:    "*prefix",
				Values:  []string{"48"},
			}},
		},
	}
	var result string
	if err := attrFltrRPC.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	alsPrf := &engine.AttributeProfileWithAPIOpts{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ApierTest",
			Contexts:  []string{utils.MetaSessionS},
			FilterIDs: []string{"FLTR_1"},
			Attributes: []*engine.Attribute{
				{
					FilterIDs: []string{"*string:~*req.FL1:In1"},
					Path:      "FL1",
					Value:     config.NewRSRParsersMustCompile("Al1", utils.InfieldSep),
				},
			},
			Weight: 20,
		},
	}
	if err := attrFltrRPC.Call(utils.APIerSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	ev := &engine.AttrArgsProcessEvent{
		Context: utils.StringPointer(utils.MetaSessionS),
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				"Subject": "45",
			},
			APIOpts: map[string]interface{}{},
		},
	}
	var rplyEv engine.AttrSProcessEventReply
	if err := attrFltrRPC.Call(utils.AttributeSv1ProcessEvent,
		ev, &rplyEv); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}

	filter = &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_1",
			Rules: []*engine.FilterRule{{
				Element: "~*req.Subject",
				Type:    "*prefix",
				Values:  []string{"44"},
			}},
		},
	}
	if err := attrFltrRPC.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	//same event for process
	ev = &engine.AttrArgsProcessEvent{
		Context: utils.StringPointer(utils.MetaSessionS),
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				"Subject": "4444",
			},
			APIOpts: map[string]interface{}{},
		},
	}
	if err := attrFltrRPC.Call(utils.AttributeSv1ProcessEvent,
		ev, &rplyEv); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}
}

func testAttributeFltrSStopEngine(t *testing.T) {
	if err := engine.KillEngine(accDelay); err != nil {
		t.Error(err)
	}
}
