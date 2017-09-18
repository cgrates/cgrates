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
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

var sq *StatQueue

func TestStatQueuesSort(t *testing.T) {
	sInsts := StatQueues{
		&StatQueue{sqPrfl: &StatQueueProfile{ID: "FIRST", Weight: 30.0}},
		&StatQueue{sqPrfl: &StatQueueProfile{ID: "SECOND", Weight: 40.0}},
		&StatQueue{sqPrfl: &StatQueueProfile{ID: "THIRD", Weight: 30.0}},
		&StatQueue{sqPrfl: &StatQueueProfile{ID: "FOURTH", Weight: 35.0}},
	}
	sInsts.Sort()
	eSInst := StatQueues{
		&StatQueue{sqPrfl: &StatQueueProfile{ID: "SECOND", Weight: 40.0}},
		&StatQueue{sqPrfl: &StatQueueProfile{ID: "FOURTH", Weight: 35.0}},
		&StatQueue{sqPrfl: &StatQueueProfile{ID: "FIRST", Weight: 30.0}},
		&StatQueue{sqPrfl: &StatQueueProfile{ID: "THIRD", Weight: 30.0}},
	}
	if !reflect.DeepEqual(eSInst, sInsts) {
		t.Errorf("expecting: %+v, received: %+v", eSInst, sInsts)
	}
}

func TestRemEventWithID(t *testing.T) {
	sq = &StatQueue{
		SQMetrics: map[string]StatMetric{
			utils.MetaASR: &StatASR{
				Answered: 1,
				Count:    2,
				Events: map[string]bool{
					"cgrates.org:TestRemEventWithID_1": true,
					"cgrates.org:TestRemEventWithID_2": false,
				},
			},
		},
	}
	if asrIf := sq.SQMetrics[utils.MetaASR].GetValue(); asrIf.(float64) != 50 {
		t.Errorf("received ASR: %v", asrIf)
	}
	sq.remEventWithID("cgrates.org:TestRemEventWithID_1")
	if asrIf := sq.SQMetrics[utils.MetaASR].GetValue(); asrIf.(float64) != 0 {
		t.Errorf("received ASR: %v", asrIf)
	}
	sq.remEventWithID("cgrates.org:TestRemEventWithID_2")
	if asrIf := sq.SQMetrics[utils.MetaASR].GetValue(); asrIf.(float64) != -1 {
		t.Errorf("received ASR: %v", asrIf)
	}
}
