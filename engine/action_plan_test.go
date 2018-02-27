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

func TestActionPlanClone(t *testing.T) {
	at1 := &ActionPlan{
		Id:         "test",
		AccountIDs: utils.StringMap{"one": true, "two": true, "three": true},
		//ActionTimings: []*ActionTiming{},
	}
	clned, err := at1.Clone()
	if err != nil {
		t.Error(err)
	}
	at1Cloned := clned.(*ActionPlan)
	if !reflect.DeepEqual(at1, at1Cloned) {
		t.Errorf("Expecting: %+v, received: %+v", at1, at1Cloned)
	}
}

func TestCacheGetCloned(t *testing.T) {
	at1 := &ActionPlan{
		Id:         "test",
		AccountIDs: utils.StringMap{"one": true, "two": true, "three": true},
	}
	Cache.Set(utils.CacheActionPlans, "MYTESTAPL", at1, nil, true, "")
	clned, err := Cache.GetCloned(utils.CacheActionPlans, "MYTESTAPL")
	if err != nil {
		t.Error(err)
	}
	at1Cloned := clned.(*ActionPlan)
	if !reflect.DeepEqual(at1, at1Cloned) {
		t.Errorf("Expecting: %+v, received: %+v", at1, at1Cloned)
	}
}
