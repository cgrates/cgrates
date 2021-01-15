/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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

package actions

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestMatchingActionProfilesForEvent(t *testing.T) {
	defaultCfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filters := engine.NewFilterS(defaultCfg, nil, dm)
	acts := NewActionS(defaultCfg, filters, dm)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TEST_ACTIONS1",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
			utils.Destination:  1002,
		},
	}

	actPrf := &engine.ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "test_id1",
		FilterIDs: []string{"*string:~*req.Account:1001;1002;1003", "*prefix:~*req.Destination:10"},
		Actions: []*engine.APAction{
			{
				ID:        "TOPUP",
				FilterIDs: []string{},
				Type:      "*topup",
				Path:      "~*balance.TestBalance.Value",
				Value:     config.NewRSRParsersMustCompile("10", defaultCfg.GeneralCfg().RSRSep),
			},
		},
	}

	if err := acts.dm.SetActionProfile(actPrf, true); err != nil {
		t.Error(err)
	}

	expActionPrf := engine.ActionProfiles{actPrf}

	if rcv, err := acts.matchingActionProfilesForEvent("cgrates.org", cgrEv, []string{}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, expActionPrf) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expActionPrf), utils.ToJSON(rcv))
	}
}
