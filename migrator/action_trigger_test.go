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
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestV1ActionTriggersAsActionTriggers(t *testing.T) {
	tim := time.Date(0001, time.January, 1, 2, 0, 0, 0, time.UTC)
	v1atrs := &v1ActionTrigger{
		Id:                    "Test",
		BalanceType:           "*monetary",
		BalanceDirection:      "*out",
		ThresholdType:         "*max_balance",
		ThresholdValue:        2,
		ActionsId:             "TEST_ACTIONS",
		Executed:              true,
		BalanceExpirationDate: tim,
	}
	atrs := &engine.ActionTrigger{
		ID: "Test",
		Balance: &engine.BalanceFilter{
			ExpirationDate: utils.TimePointer(tim),
			Type:           utils.StringPointer(utils.MetaMonetary),
		},
		ExpirationDate:    tim,
		LastExecutionTime: tim,
		ActivationDate:    tim,
		ThresholdType:     utils.TriggerMaxBalance,
		ThresholdValue:    2,
		ActionsID:         "TEST_ACTIONS",
		Executed:          true,
	}

	newatrs := v1atrs.AsActionTrigger()
	if !reflect.DeepEqual(atrs, newatrs) {
		t.Errorf("Expecting: %+v, received: %+v", atrs, newatrs)
	}
}
