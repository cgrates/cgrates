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

package utils

import (
	"reflect"
	"testing"
)

func TestTPDestinationAsExportSlice(t *testing.T) {
	tpDst := &TPDestination{
		TPid:          "TEST_TPID",
		DestinationId: "TEST_DEST",
		Prefixes:      []string{"49", "49176", "49151"},
	}
	expectedSlc := [][]string{
		[]string{"TEST_TPID", "TEST_DEST", "49"},
		[]string{"TEST_TPID", "TEST_DEST", "49176"},
		[]string{"TEST_TPID", "TEST_DEST", "49151"},
	}
	if slc := tpDst.AsExportSlice(); !reflect.DeepEqual(expectedSlc, slc) {
		t.Errorf("Expecting: %+v, received: %+v", expectedSlc, slc)
	}
}

func TestTPRateAsExportSlice(t *testing.T) {
	tpRate := &TPRate{
		TPid:   "TEST_TPID",
		RateId: "TEST_RATEID",
		RateSlots: []*RateSlot{
			&RateSlot{
				ConnectFee:         0.100,
				Rate:               0.200,
				RateUnit:           "60",
				RateIncrement:      "60",
				GroupIntervalStart: "0"},
			&RateSlot{
				ConnectFee:         0.0,
				Rate:               0.1,
				RateUnit:           "1",
				RateIncrement:      "60",
				GroupIntervalStart: "60"},
		},
	}
	expectedSlc := [][]string{
		[]string{"TEST_TPID", "TEST_RATEID", "0.1", "0.2", "60", "60", "0"},
		[]string{"TEST_TPID", "TEST_RATEID", "0", "0.1", "1", "60", "60"},
	}
	if slc := tpRate.AsExportSlice(); !reflect.DeepEqual(expectedSlc, slc) {
		t.Errorf("Expecting: %+v, received: %+v", expectedSlc, slc)
	}
}

func TestTPDestinationRateAsExportSlice(t *testing.T) {
	tpDstRate := &TPDestinationRate{
		TPid:              "TEST_TPID",
		DestinationRateId: "TEST_DSTRATE",
		DestinationRates: []*DestinationRate{
			&DestinationRate{
				DestinationId:    "TEST_DEST1",
				RateId:           "TEST_RATE1",
				RoundingMethod:   "*up",
				RoundingDecimals: 4},
			&DestinationRate{
				DestinationId:    "TEST_DEST2",
				RateId:           "TEST_RATE2",
				RoundingMethod:   "*up",
				RoundingDecimals: 4},
		},
	}
	expectedSlc := [][]string{
		[]string{"TEST_TPID", "TEST_DSTRATE", "TEST_DEST1", "TEST_RATE1", "*up", "4"},
		[]string{"TEST_TPID", "TEST_DSTRATE", "TEST_DEST2", "TEST_RATE2", "*up", "4"},
	}
	if slc := tpDstRate.AsExportSlice(); !reflect.DeepEqual(expectedSlc, slc) {
		t.Errorf("Expecting: %+v, received: %+v", expectedSlc, slc)
	}

}
