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
	"sync"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestTrendGetTrendGrowth(t *testing.T) {
	now := time.Now()
	t1 := now.Add(-time.Second)
	t2 := now.Add(-2 * time.Second)
	t3 := now.Add(-3 * time.Second)
	trnd1 := &Trend{
		RWMutex:  &sync.RWMutex{},
		Tenant:   "cgrates.org",
		ID:       "TestTrendGetTrendLabel",
		RunTimes: []time.Time{t3, t2, t1},
		Metrics: map[time.Time]map[string]*MetricWithTrend{
			t3: {utils.MetaTCD: {utils.MetaTCD, float64(41 * time.Second), -1.0, utils.NotAvailable}, utils.MetaTCC: {utils.MetaTCC, 41.0, -1.0, utils.NotAvailable}},
			t2: {utils.MetaTCD: {utils.MetaTCD, float64(9 * time.Second), -78.048, utils.MetaNegative}, utils.MetaTCC: {utils.MetaTCC, 9.0, -78.048, utils.MetaNegative}},
			t1: {utils.MetaTCD: {utils.MetaTCD, float64(10 * time.Second), 11.11111, utils.MetaPositive}, utils.MetaTCC: {utils.MetaTCC, 10.0, 11.11111, utils.MetaPositive}}},
	}
	trnd1.computeIndexes()
	if _, err := trnd1.getTrendGrowth(utils.MetaTCD, float64(11*time.Second), utils.NotAvailable, 5); err != utils.ErrCorrelationUndefined {
		t.Error(err)
	}
	if growth, err := trnd1.getTrendGrowth(utils.MetaTCD, float64(11*time.Second), utils.MetaLast, 5); err != nil || growth != 10.0 {
		t.Errorf("Expecting: <%f> got <%f>, err: %v", 10.0, growth, err)
	}
	if growth, err := trnd1.getTrendGrowth(utils.MetaTCD, float64(11*time.Second), utils.MetaAverage, 5); err != nil || growth != -45.0 {
		t.Errorf("Expecting: <%f> got <%f>, err: %v", -45.0, growth, err)
	}
}

func TestTrendGetTrendLabel(t *testing.T) {
	now := time.Now()
	t1 := now.Add(-time.Second)
	t2 := now.Add(-2 * time.Second)
	t3 := now.Add(-3 * time.Second)
	trnd1 := &Trend{
		RWMutex:  sync.RWMutex{},
		Tenant:   "cgrates.org",
		ID:       "TestTrendGetTrendLabel",
		RunTimes: []time.Time{t3, t2, t1},
		Metrics: map[time.Time]map[string]*MetricWithTrend{
			t3: {utils.MetaTCD: {utils.MetaTCD, float64(41 * time.Second), -1.0, utils.NotAvailable}, utils.MetaTCC: {utils.MetaTCC, 41.0, -1.0, utils.NotAvailable}},
			t2: {utils.MetaTCD: {utils.MetaTCD, float64(9 * time.Second), -78.048, utils.MetaNegative}, utils.MetaTCC: {utils.MetaTCC, 9.0, -78.048, utils.MetaNegative}},
			t1: {utils.MetaTCD: {utils.MetaTCD, float64(10 * time.Second), 11.11111, utils.MetaPositive}, utils.MetaTCC: {utils.MetaTCC, 10.0, 11.11111, utils.MetaPositive}}},
	}
	trnd1.computeIndexes()
	expct := utils.MetaPositive
	if lbl := trnd1.getTrendLabel(11.0, 0.0); lbl != expct {
		t.Errorf("Expecting: <%q> got <%q>", expct, lbl)
	}
	if lbl := trnd1.getTrendLabel(11.0, 10.0); lbl != expct {
		t.Errorf("Expecting: <%q> got <%q>", expct, lbl)
	}
	expct = utils.MetaConstant
	if lbl := trnd1.getTrendLabel(11.0, 11.0); lbl != expct {
		t.Errorf("Expecting: <%q> got <%q>", expct, lbl)
	}
	expct = utils.MetaNegative
	if lbl := trnd1.getTrendLabel(-9.0, 8.0); lbl != expct {
		t.Errorf("Expecting: <%q> got <%q>", expct, lbl)
	}
	expct = utils.MetaConstant
	if lbl := trnd1.getTrendLabel(-9.0, 10.0); lbl != expct {
		t.Errorf("Expecting: <%q> got <%q>", expct, lbl)
	}
}
