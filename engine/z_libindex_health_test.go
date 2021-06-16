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

/*
func TestHealthAccountAction(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	if err := dm.SetAccountActionPlans("1001", []string{"AP1", "AP2"}, true); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetActionPlan("AP2", &ActionPlan{
		Id:            "AP2",
		AccountIDs:    utils.NewStringMap("1002"),
		ActionTimings: []*ActionTiming{{}},
	}, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}
	if rply, err := GetAccountActionPlanIndexHealth(dm, -1, -1, -1, -1, false, false); err != nil {
		t.Fatal(err)
	} else {
		t.Error(utils.ToJSON(rply))
	}

}
*/
