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
	Id        int64
	Tpid      string
	Tag       string
	Years     string
	Months    string
	MonthDays string
	WeekDays  string
	Time      string
}

type TpDestination struct {
	Id     int64
	Tpid   string
	Tag    string
	Prefix string
}

type TpRate struct {
	Id                 int64
	Tpid               string
	Tag                string
	ConnectFee         float64
	Rate               float64
	RateUnit           string
	RateIncrement      string
	GroupIntervalStart string
}

type TpDestinationRate struct {
	Id               int64
	Tpid             string
	Tag              string
	DestinationsTag  string
	RatesTag         string
	RoundingMethod   string
	RoundingDecimals int
}

type TpRatingPlan struct {
	Id           int64
	Tpid         string
	Tag          string
	DestratesTag string
	TimingTag    string
	Weight       float64
}

type TpRatingProfile struct {
	Id               int64
	Tpid             string
	Loadid           string
	Direction        string
	Tenant           string
	Category         string
	Subject          string
	ActivationTime   string
	RatingPlanTag    string
	FallbackSubjects string
}

func (rpf *TpRatingProfile) SetRatingProfileId(id string) error {
	ids := strings.Split(id, utils.CONCATENATED_KEY_SEP)
	if len(ids) != 5 {
		return fmt.Errorf("Wrong TP Rating Profile Id: %s", id)
	}
	rpf.Loadid = ids[0]
	rpf.Direction = ids[1]
	rpf.Tenant = ids[2]
	rpf.Category = ids[3]
	rpf.Subject = ids[4]
	return nil
}

type TpLcrRules struct {
	Id             int64
	Tpid           string
	Direction      string
	Tenant         string
	Customer       string
	DestinationTag string
	Category       string
	Strategy       string
	Suppliers      string
	ActivatinTime  string
	Weight         float64
}

type TpAction struct {
	Id              int64
	Tpid            string
	Tag             string
	Action          string
	BalanceType     string
	Direction       string
	Units           float64
	ExpiryTime      string
	DestinationTag  string
	RatingSubject   string
	Category        string
	SharedGroup     string
	BalanceWeight   float64
	ExtraParameters string
	Weight          float64
}

type TpActionPlan struct {
	Id         int64
	Tpid       string
	Tag        string
	ActionsTag string
	TimingTag  string
	Weight     float64
}

type TpActionTrigger struct {
	Id                   int64
	Tpid                 string
	Tag                  string
	BalanceType          string
	Direction            string
	ThresholdType        string
	ThresholdValue       float64
	Recurrent            int
	MinSleep             int64
	DestinationTag       string
	BalanceWeight        float64
	BalanceExpiryTime    string
	BalanceRatingSubject string
	BalanceCategory      string
	BalanceSharedGroup   string
	MinQueuedItems       int
	ActionsTag           string
	Weight               float64
}

type TpAccountAction struct {
	Id                int64
	Tpid              string
	Loadid            string
	Direction         string
	Tenant            string
	Account           string
	ActionPlanTag     string
	ActionTriggersTag string
}

func (aa *TpAccountAction) SetAccountActionId(id string) error {
	ids := strings.Split(id, utils.CONCATENATED_KEY_SEP)
	if len(ids) != 4 {
		return fmt.Errorf("Wrong TP Account Action Id: %s", id)
	}
	aa.Loadid = ids[0]
	aa.Direction = ids[1]
	aa.Tenant = ids[2]
	aa.Account = ids[3]
	return nil
}

type TpSharedGroup struct {
	Id            int64
	Tpid          string
	Tag           string
	Account       string
	Strategy      string
	RatingSubject string
}

type TpDerivedCharger struct {
	Id               int64
	Tpid             string
	Loadid           string
	Direction        string
	Tenant           string
	Category         string
	Account          string
	Subject          string
	RunId            string
	RunFilters       string
	ReqTypeField     string
	DirectionField   string
	TenantField      string
	CategoryField    string
	AccountField     string
	SubjectField     string
	DestinationField string
	SetupTimeField   string
	AnswerTimeField  string
	UsageField       string
}

func (tpdc *TpDerivedCharger) SetDerivedChargersId(id string) error {
	ids := strings.Split(id, utils.CONCATENATED_KEY_SEP)
	if len(ids) != 6 {
		return fmt.Errorf("Wrong TP Derived Charger Id: %s", id)
	}
	tpdc.Loadid = ids[0]
	tpdc.Direction = ids[1]
	tpdc.Tenant = ids[2]
	tpdc.Category = ids[3]
	tpdc.Account = ids[4]
	tpdc.Subject = ids[5]
	return nil
}

type TpCdrStat struct {
	Id                int64
	Tpid              string
	Tag               string
	QueueLength       int
	TimeWindow        string
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
