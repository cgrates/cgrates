/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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
	"time"

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
	CreatedAt time.Time
}

type TpDestination struct {
	Id        int64
	Tpid      string
	Tag       string
	Prefix    string
	CreatedAt time.Time
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
	CreatedAt          time.Time
}

type TpDestinationRate struct {
	Id               int64
	Tpid             string
	Tag              string
	DestinationsTag  string
	RatesTag         string
	RoundingMethod   string
	RoundingDecimals int
	MaxCost          float64
	MaxCostStrategy  string
	CreatedAt        time.Time
}

type TpRatingPlan struct {
	Id           int64
	Tpid         string
	Tag          string
	DestratesTag string
	TimingTag    string
	Weight       float64
	CreatedAt    time.Time
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
	CdrStatQueueIds  string
	CreatedAt        time.Time
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
	CreatedAt      time.Time
}

type TpAction struct {
	Id              int64
	Tpid            string
	Tag             string
	Action          string
	BalanceTag      string
	BalanceType     string
	Direction       string
	Units           float64
	ExpiryTime      string
	TimingTags      string
	DestinationTag  string
	RatingSubject   string
	Category        string
	SharedGroup     string
	BalanceWeight   float64
	ExtraParameters string
	Weight          float64
	CreatedAt       time.Time
}

type TpActionPlan struct {
	Id         int64
	Tpid       string
	Tag        string
	ActionsTag string
	TimingTag  string
	Weight     float64
	CreatedAt  time.Time
}

type TpActionTrigger struct {
	Id                    int64
	Tpid                  string
	Tag                   string
	UniqueId              string
	ThresholdType         string
	ThresholdValue        float64
	Recurrent             bool
	MinSleep              string
	BalanceTag            string
	BalanceType           string
	BalanceDirection      string
	BalanceDestinationTag string
	BalanceWeight         float64
	BalanceExpiryTime     string
	BalanceTimingTags     string
	BalanceRatingSubject  string
	BalanceCategory       string
	BalanceSharedGroup    string
	MinQueuedItems        int
	ActionsTag            string
	Weight                float64
	CreatedAt             time.Time
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
	CreatedAt         time.Time
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
	CreatedAt     time.Time
}

type TpDerivedCharger struct {
	Id                   int64
	Tpid                 string
	Loadid               string
	Direction            string
	Tenant               string
	Category             string
	Account              string
	Subject              string
	Runid                string
	RunFilters           string
	ReqTypeField         string
	DirectionField       string
	TenantField          string
	CategoryField        string
	AccountField         string
	SubjectField         string
	DestinationField     string
	SetupTimeField       string
	AnswerTimeField      string
	UsageField           string
	SupplierField        string
	DisconnectCauseField string
	CreatedAt            time.Time
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
	Id                  int64
	Tpid                string
	Tag                 string
	QueueLength         int
	TimeWindow          string
	Metrics             string
	SetupInterval       string
	Tors                string
	CdrHosts            string
	CdrSources          string
	ReqTypes            string
	Directions          string
	Tenants             string
	Categories          string
	Accounts            string
	Subjects            string
	DestinationPrefixes string
	UsageInterval       string
	Suppliers           string
	DisconnectCauses    string
	MediationRunids     string
	RatedAccounts       string
	RatedSubjects       string
	CostInterval        string
	ActionTriggers      string
	CreatedAt           time.Time
}

type TblCdrsPrimary struct {
	Id              int64
	Cgrid           string
	Tor             string
	Accid           string
	Cdrhost         string
	Cdrsource       string
	Reqtype         string
	Direction       string
	Tenant          string
	Category        string
	Account         string
	Subject         string
	Destination     string
	SetupTime       time.Time
	AnswerTime      time.Time
	Usage           float64
	Supplier        string
	DisconnectCause string
	CreatedAt       time.Time
	DeletedAt       time.Time
}

func (t TblCdrsPrimary) TableName() string {
	return utils.TBL_CDRS_PRIMARY
}

type TblCdrsExtra struct {
	Id          int64
	Cgrid       string
	ExtraFields string
	CreatedAt   time.Time
	DeletedAt   time.Time
}

func (t TblCdrsExtra) TableName() string {
	return utils.TBL_CDRS_EXTRA
}

type TblCostDetail struct {
	Id          int64
	Cgrid       string
	Runid       string
	Tor         string
	Direction   string
	Tenant      string
	Category    string
	Account     string
	Subject     string
	Destination string
	Cost        float64
	Timespans   string
	CostSource  string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   time.Time
}

func (t TblCostDetail) TableName() string {
	return utils.TBL_COST_DETAILS
}

type TblRatedCdr struct {
	Id              int64
	Cgrid           string
	Runid           string
	Reqtype         string
	Direction       string
	Tenant          string
	Category        string
	Account         string
	Subject         string
	Destination     string
	SetupTime       time.Time
	AnswerTime      time.Time
	Usage           float64
	Supplier        string
	DisconnectCause string
	Cost            float64
	ExtraInfo       string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       time.Time
}

func (t TblRatedCdr) TableName() string {
	return utils.TBL_RATED_CDRS
}
