/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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

type TpTiming struct {
	Tbid      int64 `gorm:"primary_key:yes"`
	Tpid      string
	Id        string
	Years     string
	Months    string
	MonthDays string
	WeekDays  string
	Time      string
}

type TpDestination struct {
	Tbid   int64 `gorm:"primary_key:yes"`
	Tpid   string
	Id     string
	Prefix string
}

type TpRate struct {
	Tbid               int64 `gorm:"primary_key:yes"`
	Tpid               string
	Id                 string
	ConnectFee         float64
	Rate               float64
	RateUnit           string
	RateIncrement      string
	GroupIntervalStart string
}

type TpDestinationRate struct {
	Tbid             int64 `gorm:"primary_key:yes"`
	Tpid             string
	Id               string
	DestinationsId   string
	RatesId          string
	RoundingMethod   string
	RoundingDecimals int
}

type TpRatingPlan struct {
	Tbid        int64 `gorm:"primary_key:yes"`
	Tpid        string
	Id          string
	DestratesId string
	TimingId    string
	Weight      float64
}
