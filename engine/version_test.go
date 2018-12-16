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
	message1 := y.Compare(x, utils.MONGO, true)
	if message1 != "cgr-migrator -migrate=*accounts" {
		t.Errorf("Error failed to compare to curent version expected: %s received: %s", "cgr-migrator -migrate=*accounts", message1)
	}
	message2 := z.Compare(x, utils.MONGO, true)
	if message2 != "cgr-migrator -migrate=*action_plans" {
		t.Errorf("Error failed to compare to curent version expected: %s received: %s", "cgr-migrator -migrate=*action_plans", message2)
	}
	message3 := q.Compare(x, utils.MONGO, true)
	if message3 != "cgr-migrator -migrate=*shared_groups" {
		t.Errorf("Error failed to compare to curent version expected: %s received: %s", "cgr-migrator -migrate=*shared_groups", message3)
	}
	message4 := c.Compare(x, utils.MONGO, false)
	if message4 != "cgr-migrator -migrate=*cost_details" {
		t.Errorf("Error failed to compare to curent version expected: %s received: %s", "cgr-migrator -migrate=*cost_details", message4)
	}
	message5 := a.Compare(b, utils.MYSQL, false)
	if message5 != "cgr-migrator -migrate=*sessions_costs" {
		t.Errorf("Error failed to compare to curent version expected: %s received: %s", "cgr-migrator -migrate=*sessions_costs", message5)
	}
	message6 := a.Compare(b, utils.POSTGRES, false)
	if message6 != "cgr-migrator -migrate=*sessions_costs" {
		t.Errorf("Error failed to compare to curent version expected: %s received: %s", "cgr-migrator -migrate=*sessions_costs", message6)
	}

}
