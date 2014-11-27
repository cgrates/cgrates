/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2014 ITsysCOM

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

func TestGetRatingProfileForPrefix(t *testing.T) {
	cd := &CallDescriptor{
		TimeStart:   time.Date(2013, 11, 18, 13, 45, 1, 0, time.UTC),
		TimeEnd:     time.Date(2013, 11, 18, 13, 47, 30, 0, time.UTC),
		Tenant:      "vdf",
		Category:    "0",
		Direction:   OUTBOUND,
		Subject:     "fallback1",
		Destination: "0256098",
	}
	cd.LoadRatingPlans()
	if len(cd.RatingInfos) != 3 || !cd.continousRatingInfos() {
		t.Logf("0: %+v", cd.RatingInfos[0])
		t.Logf("1: %+v", cd.RatingInfos[1])
		t.Logf("2: %+v", cd.RatingInfos[2])
		t.Errorf("Error loading rating information: %+v %+v", cd.RatingInfos, cd.continousRatingInfos())
	}
}

func TestGetRatingProfileForPrefixFirstEmpty(t *testing.T) {
	cd := &CallDescriptor{
		TimeStart:   time.Date(2013, 11, 18, 13, 44, 1, 0, time.UTC),
		TimeEnd:     time.Date(2013, 11, 18, 13, 47, 30, 0, time.UTC),
		Tenant:      "vdf",
		Category:    "0",
		Direction:   OUTBOUND,
		Subject:     "fallback1",
		Destination: "0256098",
	}
	cd.LoadRatingPlans()
	if len(cd.RatingInfos) != 4 || !cd.continousRatingInfos() {
		t.Logf("0: %+v", cd.RatingInfos[0])
		t.Logf("1: %+v", cd.RatingInfos[1])
		t.Logf("2: %+v", cd.RatingInfos[2])
		t.Logf("3: %+v", cd.RatingInfos[3])
		t.Errorf("Error loading rating information: %+v %+v", cd.RatingInfos, cd.continousRatingInfos())
	}
}
