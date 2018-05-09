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
package migrator

/*
import (
	"encoding/json"
	"testing"
)

var v1TmspsStr1 = `[{"TimeStart":"2016-07-28T02:18:49+02:00","TimeEnd":"2016-07-28T02:19:28+02:00","Cost":0.0117,"RateInterval":{"Timing":{"Years":[],"Months":[],"MonthDays":[],"WeekDays":[],"StartTime":"00:00:00","EndTime":""},"Rating":{"ConnectFee":0.0564,"RoundingMethod":"*middle","RoundingDecimals":4,"MaxCost":0,"MaxCostStrategy":"","Rates":[{"GroupIntervalStart":0,"Value":0.0198,"RateIncrement":1000000000,"RateUnit":60000000000}]},"Weight":10},"DurationIndex":39000000000,"Increments":[{"Duration":1000000000,"Cost":0.0003,"BalanceInfo":{"UnitBalanceUuid":"","MoneyBalanceUuid":"c50c201c405defc3807347444efc62da","AccountId":"cgrates.org:dan"},"BalanceRateInterval":null,"UnitInfo":null,"CompressFactor":39}],"MatchedSubject":"*out:cgrates.org:call:dan","MatchedPrefix":"+311","MatchedDestId":"CST_491_DE001","RatingPlanId":"V_RET_1490_01_V"}]`
var v1TmspsStr2 = `[{"TimeStart":"2016-07-28T01:12:19+02:00","TimeEnd":"2016-07-28T01:12:27+02:00","Cost":0.00046875,"RateInterval":{"Timing":{"Years":[],"Months":[],"MonthDays":[],"WeekDays":[],"StartTime":"00:00:00","EndTime":""},"Rating":{"ConnectFee":0,"RoundingMethod":"*middle","RoundingDecimals":4,"MaxCost":0,"MaxCostStrategy":"","Rates":[{"GroupIntervalStart":0,"Value":0.06,"RateIncrement":1000000000,"RateUnit":1024000000000}]},"Weight":10},"DurationIndex":8000000000,"Increments":null,"MatchedSubject":"*out:cgrates.org:data:danb","MatchedPrefix":"+4900","MatchedDestId":"CST_data_DAT01","RatingPlanId":"M_RET_1409_01_D"}]`

func TestV1CostDetailsAsCostDetails1(t *testing.T) {
	var v1tmsps v1TimeSpans
	if err := json.Unmarshal([]byte(v1TmspsStr1), &v1tmsps); err != nil {
		t.Error(err)
	}
	v1CC := &v1CallCost{Timespans: v1tmsps}
	_ = v1CC.AsCallCost()
	// ToDo: Test here the content

}

func TestV1CostDetailsAsCostDetails2(t *testing.T) {
	var v1tmsps v1TimeSpans
	if err := json.Unmarshal([]byte(v1TmspsStr2), &v1tmsps); err != nil {
		t.Error(err)
	}
	v1CC := &v1CallCost{Timespans: v1tmsps}
	_ = v1CC.AsCallCost()

}
*/
