/*
Real-time Charging System for Telecom & ISP environments
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
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestIndexStringFilters(t *testing.T) {
	rls := []*ResourceLimit{
		&ResourceLimit{
			ID:     "RL1",
			Weight: 20,
			Filters: []*RequestFilter{
				&RequestFilter{Type: MetaString, FieldName: "Account", Values: []string{"1001", "1002"}},
				&RequestFilter{Type: MetaRSRFields, Values: []string{"Subject(~^1.*1$)", "Destination(1002)"},
					rsrFields: utils.ParseRSRFieldsMustCompile("Subject(~^1.*1$);Destination(1002)", utils.INFIELD_SEP),
				}},
			ActivationTime: time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
			Limit:          2,
		},
		&ResourceLimit{
			ID:     "RL2",
			Weight: 10,
			Filters: []*RequestFilter{
				&RequestFilter{Type: MetaString, FieldName: "Account", Values: []string{"dan", "1002"}},
				&RequestFilter{Type: MetaString, FieldName: "Subject", Values: []string{"dan"}},
			},
			ActivationTime: time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
			Limit:          1,
		},
		&ResourceLimit{
			ID:     "RL4",
			Weight: 10,
			Filters: []*RequestFilter{
				&RequestFilter{Type: MetaStringPrefix, FieldName: "Destination", Values: []string{"+49"}},
			},
			ActivationTime: time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
			Limit:          1,
		},
		&ResourceLimit{
			ID:     "RL5",
			Weight: 10,
			Filters: []*RequestFilter{
				&RequestFilter{Type: MetaStringPrefix, FieldName: "Destination", Values: []string{"+40"}},
			},
			ActivationTime: time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
			Limit:          1,
		},
	}
	for _, rl := range rls {
		CacheSet(utils.ResourceLimitsPrefix+rl.ID, rl)
	}
	rLS := new(ResourceLimiterService)
	eIndexes := map[string]map[string]utils.StringMap{
		"Account": map[string]utils.StringMap{
			"1001": utils.StringMap{
				"RL1": true,
			},
			"1002": utils.StringMap{
				"RL1": true,
				"RL2": true,
			},
			"dan": utils.StringMap{
				"RL2": true,
			},
		},
		"Subject": map[string]utils.StringMap{
			"dan": utils.StringMap{
				"RL2": true,
			},
		},
		utils.NOT_AVAILABLE: map[string]utils.StringMap{
			utils.NOT_AVAILABLE: utils.StringMap{
				"RL4": true,
				"RL5": true,
			},
		},
	}
	if err := rLS.indexStringFilters(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIndexes, rLS.stringIndexes) {
		t.Errorf("Expecting: %+v, received: %+v", eIndexes, rLS.stringIndexes)
	}
	rl3 := &ResourceLimit{
		ID:     "RL3",
		Weight: 10,
		Filters: []*RequestFilter{
			&RequestFilter{Type: MetaString, FieldName: "Subject", Values: []string{"dan"}},
			&RequestFilter{Type: MetaString, FieldName: "Subject", Values: []string{"1003"}},
		},
		ActivationTime: time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
		Limit:          1,
	}
	CacheSet(utils.ResourceLimitsPrefix+rl3.ID, rl3)
	eIndexes = map[string]map[string]utils.StringMap{
		"Account": map[string]utils.StringMap{
			"1001": utils.StringMap{
				"RL1": true,
			},
			"1002": utils.StringMap{
				"RL1": true,
				"RL2": true,
			},
			"dan": utils.StringMap{
				"RL2": true,
			},
		},
		"Subject": map[string]utils.StringMap{
			"dan": utils.StringMap{
				"RL2": true,
				"RL3": true,
			},
			"1003": utils.StringMap{
				"RL3": true,
			},
		},
		utils.NOT_AVAILABLE: map[string]utils.StringMap{
			utils.NOT_AVAILABLE: utils.StringMap{
				"RL4": true,
				"RL5": true,
			},
		},
	}
	// Test index update
	if err := rLS.indexStringFilters([]string{rl3.ID}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIndexes, rLS.stringIndexes) {
		t.Errorf("Expecting: %+v, received: %+v", eIndexes, rLS.stringIndexes)
	}
}
