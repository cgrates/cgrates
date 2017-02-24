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
package migrator

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var v1ActionTriggers1 = `{"BalanceType": "*monetary","BalanceDirection": "*out","ThresholdType":"*max_balance", "ThresholdValue" :2, "ActionsId": "TEST_ACTIONS", "Executed": true}`

func TestV1ActionTriggersAsActionTriggers(t *testing.T) {
	atrs := &engine.ActionTrigger{
		Balance: &engine.BalanceFilter{
			Type:       utils.StringPointer(utils.MONETARY),
			Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT)),
		},
		ThresholdType:  utils.TRIGGER_MAX_BALANCE,
		ThresholdValue: 2,
		ActionsID:      "TEST_ACTIONS",
		Executed:       true,
	}
	var v1actstrgrs v1ActionTrigger
	if err := json.Unmarshal([]byte(v1ActionTriggers1), &v1actstrgrs); err != nil {
		t.Error(err)
	}
	if newatrs, err := v1actstrgrs.AsActionTrigger(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(*atrs, newatrs) {
		t.Errorf("Expecting: %+v, received: %+v", *atrs, newatrs)
	}
}
