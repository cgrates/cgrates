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

// Structs here are one to one mapping of the tables and fields
// to be used by gorm orm

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

type TpLcrRules struct {
	Tbid          int64 `gorm:"primary_key:yes"`
	Tpid          string
	Direction     string
	Tenant        string
	Customer      string
	DestinationId string
	Category      string
	Strategy      string
	Suppliers     string
	ActivatinTime string
	Weight        float64
}

type TpAction struct {
	Tbid            int64 `gorm:"primary_key:yes"`
	Tpid            string
	Id              string
	Action          string
	BalanceType     string
	Direction       string
	Units           float64
	ExpiryTime      string
	DestinationId   string
	RatingSubject   string
	SharedGroup     string
	BalanceWeight   float64
	ExtraParameters string
	Weight          float64
}

type TpActionPlan struct {
	Tbid      int64 `gorm:"primary_key:yes"`
	Tpid      string
	Id        string
	ActionsId string
	TimingId  string
	Weight    float64
}

type TpActionTrigger struct {
	Tbid                 int64 `gorm:"primary_key:yes"`
	Tpid                 string
	Id                   string
	BalanceType          string
	Direction            string
	ThresholdType        string
	ThresholdValue       float64
	Recurrent            int
	MinSleep             int64
	DestinationId        string
	BalanceWeight        float64
	BalanceExpiryTime    string
	BalanceRatingSubject string
	BalanceSharedGroup   string
	MinQueuedItems       int
	ActionsId            string
	Weight               float64
}

type TpSharedGroup struct {
	Tbid          int64 `gorm:"primary_key:yes"`
	Tpid          string
	Id            string
	Account       string
	Strategy      string
	RatingSubject string
}
