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

import (
	"fmt"
	"strings"

	"github.com/cgrates/cgrates/utils"
)

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

type TpDerivedCharger struct {
	Tbid             int64 `gorm:"primary_key:yes"`
	Tpid             string
	Loadid           string
	Direction        string
	Tenant           string
	Category         string
	Account          string
	Subject          string
	RunId            string
	RunFilter        string
	ReqtypeField     string
	DirectionField   string
	TenantField      string
	CategoryField    string
	AccountField     string
	SubjectField     string
	DestinationField string
	SetupTimeField   string
	AnswerTimeField  string
	DurationField    string
}

func (tpdc *TpDerivedCharger) SetDerivedChargersId(id string) error {
	ids := strings.Split(id, utils.TP_ID_SEP)
	if len(ids) != 6 {
		return fmt.Errorf("Wrong TP Derived Charge Id!")
	}
	tpdc.Direction = ids[0]
	tpdc.Tenant = ids[1]
	tpdc.Category = ids[2]
	tpdc.Account = ids[3]
	tpdc.Subject = ids[4]
	tpdc.Loadid = ids[5]
	return nil
}

type TpCdrStat struct {
	Tbid              int64 `gorm:"primary_key:yes"`
	Tpid              string
	Id                string
	QueueLength       int
	TimeWindow        int64
	Metrics           string
	SetupInterval     string
	Tor               string
	CdrHost           string
	CdrSource         string
	ReqType           string
	Direction         string
	Tenant            string
	Category          string
	Account           string
	Subject           string
	DestinationPrefix string
	UsageInterval     string
	MediationRunIds   string
	RatedAccount      string
	RatedSubject      string
	CostInterval      string
	ActionTriggers    string
}
