// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ers

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestLibErsMergePartialEvents(t *testing.T) {
	confg := config.NewDefaultCGRConfig()
	locker := engine.NewGuardianLocker(confg)
	cacheS := engine.NewCacheS(confg, nil, nil, nil, locker)
	fltrS := engine.NewFilterS(confg, nil, nil)
	cgrEvs := []*utils.CGREvent{
		{
			Tenant: "cgrates.org",
			ID:     "ev1",
			Event: map[string]any{
				"EvField1":       "Value1",
				"EvField2":       "Value4",
				utils.AnswerTime: 6.,
			},
			APIOpts: map[string]any{
				"Field1": "Value1",
				"Field2": "Value2",
			},
		},
		{
			Tenant: "cgrates.org",
			ID:     "ev2",
			Event: map[string]any{
				"EvField3":       "Value1",
				"EvField2":       "Value2",
				utils.AnswerTime: 4.,
			},
			APIOpts: map[string]any{
				"Field4": "Value2",
				"Field2": "Value3",
			},
		},
		{
			Tenant: "cgrates.org",
			ID:     "ev3",
			Event: map[string]any{
				"EvField2":       "Value2",
				"EvField4":       "Value4",
				"EvField3":       "Value3",
				utils.AnswerTime: 8.,
			},
			APIOpts: map[string]any{
				"Field3": "Value3",
				"Field4": "Value4",
			},
		},
	}
	exp := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AnswerTime: 8.,
			"EvField1":       "Value1",
			"EvField2":       "Value2",
			"EvField3":       "Value3",
			"EvField4":       "Value4",
		},
		APIOpts: map[string]any{
			"Field1": "Value1",
			"Field2": "Value2",
			"Field3": "Value3",
			"Field4": "Value4",
		},
	}
	if rcv, err := mergePartialEvents(cgrEvs, confg.ERsCfg().Readers[0], confg, cacheS, fltrS, confg.GeneralCfg().DefaultTenant,
		confg.GeneralCfg().DefaultTimezone); err != nil {
		t.Error(err)
	} else {
		rcv.ID = utils.EmptyString
		if !reflect.DeepEqual(rcv, exp) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(exp), utils.ToJSON(rcv))
		}
	}
}
