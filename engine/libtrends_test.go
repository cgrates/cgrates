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

func TestTrendGetTrendLabel(t *testing.T) {
	now := time.Now()
	t1 := now.Add(-2 * time.Second)
	t2 := now.Add(-time.Second)
	trnd1 := &Trend{
		RWMutex:  &sync.RWMutex{},
		Tenant:   "cgrates.org",
		ID:       "TestTrendGetTrendLabel",
		RunTimes: []time.Time{t1, t2, now},
		Metrics: map[time.Time]map[string]*MetricWithTrend{
			t1: {utils.MetaTCD: {utils.MetaTCD, float64(9 * time.Second), utils.NotAvailable}, utils.MetaTCC: {utils.MetaTCC, 9.0, utils.NotAvailable}},
			t2: {utils.MetaTCD: {utils.MetaTCD, float64(10 * time.Second), utils.MetaPositive}, utils.MetaTCC: {utils.MetaTCC, 10.0, utils.MetaPositive}}},
	}
	trnd1.computeIndexes()
	if lbl := trnd1.getTrendLabel(utils.MetaTCD, float64(11*time.Second), 0.0); lbl != utils.MetaPositive {
		t.Errorf("Expecting: <%q> got <%q>", utils.MetaPositive, lbl)
	}
	if lbl := trnd1.getTrendLabel(utils.MetaTCD, float64(11*time.Second), 9.0); lbl != utils.MetaPositive {
		t.Errorf("Expecting: <%q> got <%q>", utils.MetaPositive, lbl)
	}
	if lbl := trnd1.getTrendLabel(utils.MetaTCD, float64(11*time.Second), 10.0); lbl != utils.MetaConstant {
		t.Errorf("Expecting: <%q> got <%q>", utils.MetaConstant, lbl)
	}
	if lbl := trnd1.getTrendLabel(utils.MetaTCD, float64(9*time.Second), 9.0); lbl != utils.MetaNegative {
		t.Errorf("Expecting: <%q> got <%q>", utils.MetaNegative, lbl)
	}
	if lbl := trnd1.getTrendLabel(utils.MetaTCD, float64(9*time.Second), 10.0); lbl != utils.MetaConstant {
		t.Errorf("Expecting: <%q> got <%q>", utils.MetaConstant, lbl)
	}
	if lbl := trnd1.getTrendLabel(utils.MetaACD, float64(9*time.Second), 0.0); lbl != utils.NotAvailable {
		t.Errorf("Expecting: <%q> got <%q>", utils.NotAvailable, lbl)
	}
}
