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
	if rcv, err := mergePartialEvents(cgrEvs, confg.ERsCfg().Readers[0], fltrS, confg.GeneralCfg().DefaultTenant,
		confg.GeneralCfg().DefaultTimezone, confg.GeneralCfg().RSRSep); err != nil {
		t.Error(err)
	} else {
		rcv.ID = utils.EmptyString
		if !reflect.DeepEqual(rcv, exp) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(exp), utils.ToJSON(rcv))
		}
	}
}
