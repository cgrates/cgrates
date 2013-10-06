/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

This program is free software: you can Storagetribute it and/or modify
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
	"testing"
	"time"
)

func TestRpAddAPIfNotPresent(t *testing.T) {
	ap1 := &RatingPlan{ActivationTime: time.Date(2012, time.July, 2, 14, 24, 30, 0, time.UTC)}
	ap2 := &RatingPlan{ActivationTime: time.Date(2012, time.July, 2, 14, 24, 30, 0, time.UTC)}
	ap3 := &RatingPlan{ActivationTime: time.Date(2012, time.July, 2, 14, 24, 30, 1, time.UTC)}
	rp := &RatingProfile{}
	rp.AddRatingPlanIfNotPresent("test", ap1)
	rp.AddRatingPlanIfNotPresent("test", ap2)
	if len(rp.DestinationMap["test"]) != 1 {
		t.Error("Wronfully appended activation period ;)", len(rp.DestinationMap["test"]))
	}
	rp.AddRatingPlanIfNotPresent("test", ap3)
	if len(rp.DestinationMap["test"]) != 2 {
		t.Error("Wronfully not appended activation period ;)", len(rp.DestinationMap["test"]))
	}
}
