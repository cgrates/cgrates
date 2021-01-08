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

package ees

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestUpdateEEMetrics(t *testing.T) {
	dc, _ := newEEMetrics(utils.EmptyString)
	tnow := time.Now()
	ev := engine.MapEvent{
		utils.AnswerTime: tnow,
		utils.OrderID:    1,
		utils.Cost:       5.5,
		utils.ToR:        utils.MetaVoice,
		utils.Usage:      time.Second,
	}
	exp, _ := newEEMetrics(utils.EmptyString)
	exp[utils.FirstEventATime] = tnow
	exp[utils.LastEventATime] = tnow
	exp[utils.FirstExpOrderID] = int64(1)
	exp[utils.LastExpOrderID] = int64(1)
	exp[utils.TotalCost] = float64(5.5)
	exp[utils.TotalDuration] = time.Second
	exp[utils.TimeNow] = dc[utils.TimeNow]
	if updateEEMetrics(dc, ev, utils.EmptyString); !reflect.DeepEqual(dc, exp) {
		t.Errorf("Expected: %s,received: %s", utils.ToJSON(exp), utils.ToJSON(dc))
	}

	tnow = tnow.Add(24 * time.Hour)
	ev = engine.MapEvent{
		utils.AnswerTime: tnow,
		utils.OrderID:    2,
		utils.Cost:       5.5,
		utils.ToR:        utils.MetaSMS,
		utils.Usage:      time.Second,
	}
	exp[utils.LastEventATime] = tnow
	exp[utils.LastExpOrderID] = int64(2)
	exp[utils.TotalCost] = float64(11)
	exp[utils.TotalSMSUsage] = time.Second
	if updateEEMetrics(dc, ev, utils.EmptyString); !reflect.DeepEqual(dc, exp) {
		t.Errorf("Expected: %s,received: %s", utils.ToJSON(exp), utils.ToJSON(dc))
	}

	tnow = tnow.Add(24 * time.Hour)
	ev = engine.MapEvent{
		utils.AnswerTime: tnow,
		utils.OrderID:    3,
		utils.Cost:       5.5,
		utils.ToR:        utils.MetaMMS,
		utils.Usage:      time.Second,
	}
	exp[utils.LastEventATime] = tnow
	exp[utils.LastExpOrderID] = int64(3)
	exp[utils.TotalCost] = float64(16.5)
	exp[utils.TotalMMSUsage] = time.Second
	if updateEEMetrics(dc, ev, utils.EmptyString); !reflect.DeepEqual(dc, exp) {
		t.Errorf("Expected: %s,received: %s", utils.ToJSON(exp), utils.ToJSON(dc))
	}

	tnow = tnow.Add(24 * time.Hour)
	ev = engine.MapEvent{
		utils.AnswerTime: tnow,
		utils.OrderID:    4,
		utils.Cost:       5.5,
		utils.ToR:        utils.MetaGeneric,
		utils.Usage:      time.Second,
	}
	exp[utils.LastEventATime] = tnow
	exp[utils.LastExpOrderID] = int64(4)
	exp[utils.TotalCost] = float64(22)
	exp[utils.TotalGenericUsage] = time.Second
	if updateEEMetrics(dc, ev, utils.EmptyString); !reflect.DeepEqual(dc, exp) {
		t.Errorf("Expected: %s,received: %s", utils.ToJSON(exp), utils.ToJSON(dc))
	}

	tnow = tnow.Add(24 * time.Hour)
	ev = engine.MapEvent{
		utils.AnswerTime: tnow,
		utils.OrderID:    5,
		utils.Cost:       5.5,
		utils.ToR:        utils.MetaData,
		utils.Usage:      time.Second,
	}
	exp[utils.LastEventATime] = tnow
	exp[utils.LastExpOrderID] = int64(5)
	exp[utils.TotalCost] = float64(27.5)
	exp[utils.TotalDataUsage] = time.Second
	if updateEEMetrics(dc, ev, utils.EmptyString); !reflect.DeepEqual(dc, exp) {
		t.Errorf("Expected: %s,received: %s", utils.ToJSON(exp), utils.ToJSON(dc))
	}
}
