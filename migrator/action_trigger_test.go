// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
			Type:           utils.StringPointer(utils.MONETARY),
		},
		ExpirationDate:    tim,
		LastExecutionTime: tim,
		ActivationDate:    tim,
		ThresholdType:     utils.TRIGGER_MAX_BALANCE,
		ThresholdValue:    2,
		ActionsID:         "TEST_ACTIONS",
		Executed:          true,
	}

	newatrs := v1atrs.AsActionTrigger()
	if !reflect.DeepEqual(atrs, newatrs) {
		t.Errorf("Expecting: %+v, received: %+v", atrs, newatrs)
	}
}
