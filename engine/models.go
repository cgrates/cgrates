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
	Tag       string `index:"0" re:""`
	Years     string `index:"1" re:""`
	Months    string `index:"2" re:""`
	MonthDays string `index:"3" re:""`
	WeekDays  string `index:"4" re:""`
	Time      string `index:"5" re:""`
	CreatedAt time.Time
}

type TpDestination struct {
	Id        int64
	Tpid      string
	Tag       string `index:"0" re:"(?:\w+\s*,\s*){1}"`
	Prefix    string `index:"1" re:"(?:\+?\d+.?\d*){1}"`
	CreatedAt time.Time
}

type TpRate struct {
	Id                 int64
	Tpid               string
	Tag                string  `index:"0" re:""`
	ConnectFee         float64 `index:"1" re:""`
	Rate               float64 `index:"2" re:""`
	RateUnit           string  `index:"3" re:""`
	RateIncrement      string  `index:"4" re:""`
	GroupIntervalStart string  `index:"5" re:""`
	CreatedAt          time.Time
}

type TpDestinationRate struct {
	Id               int64
	Tpid             string
	Tag              string  `index:"0" re:""`
	DestinationsTag  string  `index:"1" re:""`
	RatesTag         string  `index:"2" re:""`
	RoundingMethod   string  `index:"3" re:""`
	RoundingDecimals int     `index:"4" re:""`
	MaxCost          float64 `index:"5" re:""`
	MaxCostStrategy  string  `index:"6" re:""`
	CreatedAt        time.Time
}

type TpRatingPlan struct {
	Id           int64
	Tpid         string
	Tag          string  `index:"0" re:""`
	DestratesTag string  `index:"1" re:""`
	TimingTag    string  `index:"2" re:""`
	Weight       float64 `index:"3" re:""`
	CreatedAt    time.Time
}

type TpRatingProfile struct {
	Id               int64
	Tpid             string
	Loadid           string
	Direction        string `index:"0" re:""`
	Tenant           string `index:"1" re:""`
	Category         string `index:"2" re:""`
	Subject          string `index:"3" re:""`
	ActivationTime   string `index:"4" re:""`
	RatingPlanTag    string `index:"5" re:""`
	FallbackSubjects string `index:"6" re:""`
	CdrStatQueueIds  string `index:"7" re:""`
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

func (rpf *TpRatingProfile) GetRatingProfileId() string {
	return utils.ConcatenatedKey(rpf.Loadid, rpf.Direction, rpf.Tenant, rpf.Category, rpf.Subject)
}

type TpLcrRule struct {
	Id             int64
	Tpid           string
	Direction      string `index:"0" re:""`
	Tenant         string `index:"1" re:""`
	Category       string `index:"2" re:""`
	Account        string `index:"3" re:""`
	Subject        string `index:"4" re:""`
	DestinationTag string `index:"5" re:""`
	RpCategory     string `index:"6" re:""`
	Strategy       string `index:"7" re:""`
	StrategyParams string `index:"8" re:""`
	ActivationTime string `index:"9" re:""`
	Weight         float64
	CreatedAt      time.Time
}

func (lcr *TpLcrRule) SetLcrRuleId(id string) error {
	ids := strings.Split(id, utils.CONCATENATED_KEY_SEP)
	if len(ids) != 5 {
		return fmt.Errorf("wrong LcrRule Id: %s", id)
	}
	lcr.Direction = ids[0]
	lcr.Tenant = ids[2]
	lcr.Category = ids[3]
	lcr.Account = ids[3]
	lcr.Subject = ids[5]
	return nil
}

func (lcr *TpLcrRule) GetLcrRuleId() string {
	return utils.LCRKey(lcr.Direction, lcr.Tenant, lcr.Category, lcr.Account, lcr.Subject)
}

type TpAction struct {
	Id              int64
	Tpid            string
	Tag             string  `index:"0" re:""`
	Action          string  `index:"1" re:""`
	ExtraParameters string  `index:"2" re:""`
	BalanceTag      string  `index:"3" re:""`
	BalanceType     string  `index:"4" re:""`
	Direction       string  `index:"5" re:""`
	Category        string  `index:"6" re:""`
	DestinationTags string  `index:"7" re:""`
	RatingSubject   string  `index:"8" re:""`
	SharedGroup     string  `index:"9" re:""`
	ExpiryTime      string  `index:"10" re:""`
	TimingTags      string  `index:"11" re:""`
	Units           float64 `index:"12" re:""`
	BalanceWeight   float64 `index:"13" re:""`
	Weight          float64 `index:"14" re:""`
	CreatedAt       time.Time
}

type TpActionPlan struct {
	Id         int64
	Tpid       string
	Tag        string  `index:"0" re:""`
	ActionsTag string  `index:"1" re:""`
	TimingTag  string  `index:"2" re:""`
	Weight     float64 `index:"3" re:""`
	CreatedAt  time.Time
}

type TpActionTrigger struct {
	Id                     int64
	Tpid                   string
	Tag                    string  `index:"0" re:""`
	UniqueId               string  `index:"1" re:""`
	ThresholdType          string  `index:"2" re:""`
	ThresholdValue         float64 `index:"3" re:""`
	Recurrent              bool    `index:"4" re:""`
	MinSleep               string  `index:"5" re:""`
	BalanceTag             string  `index:"6" re:""`
	BalanceType            string  `index:"7" re:""`
	BalanceDirection       string  `index:"8" re:""`
	BalanceDestinationTags string  `index:"9" re:""`
	BalanceWeight          float64 `index:"10" re:""`
	BalanceExpiryTime      string  `index:"11" re:""`
	BalanceTimingTags      string  `index:"12" re:""`
	BalanceRatingSubject   string  `index:"13" re:""`
	BalanceCategory        string  `index:"14" re:""`
	BalanceSharedGroup     string  `index:"15" re:""`
	MinQueuedItems         int     `index:"16" re:""`
	ActionsTag             string  `index:"17" re:""`
	Weight                 float64 `index:"18" re:""`
	CreatedAt              time.Time
}

type TpAccountAction struct {
	Id                int64
	Tpid              string
	Loadid            string
	Direction         string `index:"0" re:""`
	Tenant            string `index:"1" re:""`
	Account           string `index:"2" re:""`
	ActionPlanTag     string `index:"3" re:""`
	ActionTriggersTag string `index:"4" re:""`
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

func (aa *TpAccountAction) GetAccountActionId() string {
	return utils.AccountKey(aa.Tenant, aa.Account, aa.Direction)
}

type TpSharedGroup struct {
	Id            int64
	Tpid          string
	Tag           string `index:"0" re:""`
	Account       string `index:"0" re:""`
	Strategy      string `index:"0" re:""`
	RatingSubject string `index:"0" re:""`
	CreatedAt     time.Time
}

type TpDerivedCharger struct {
	Id                   int64
	Tpid                 string
	Loadid               string
	Direction            string `index:"0" re:""`
	Tenant               string `index:"1" re:""`
	Category             string `index:"2" re:""`
	Account              string `index:"3" re:""`
	Subject              string `index:"4" re:""`
	Runid                string `index:"5" re:""`
	RunFilters           string `index:"6" re:""`
	ReqTypeField         string `index:"7" re:""`
	DirectionField       string `index:"8" re:""`
	TenantField          string `index:"9" re:""`
	CategoryField        string `index:"10" re:""`
	AccountField         string `index:"11" re:""`
	SubjectField         string `index:"12" re:""`
	DestinationField     string `index:"13" re:""`
	SetupTimeField       string `index:"14" re:""`
	AnswerTimeField      string `index:"15" re:""`
	UsageField           string `index:"16" re:""`
	SupplierField        string `index:"17" re:""`
	DisconnectCauseField string `index:"18" re:""`
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

func (tpdc *TpDerivedCharger) GetDerivedChargersId() string {
	return utils.ConcatenatedKey(tpdc.Loadid, tpdc.Direction, tpdc.Tenant, tpdc.Category, tpdc.Account, tpdc.Subject)
}

type TpCdrStat struct {
	Id                  int64
	Tpid                string
	Tag                 string `index:"0" re:""`
	QueueLength         int    `index:"1" re:""`
	TimeWindow          string `index:"2" re:""`
	Metrics             string `index:"3" re:""`
	SetupInterval       string `index:"4" re:""`
	Tors                string `index:"5" re:""`
	CdrHosts            string `index:"6" re:""`
	CdrSources          string `index:"7" re:""`
	ReqTypes            string `index:"8" re:""`
	Directions          string `index:"9" re:""`
	Tenants             string `index:"10" re:""`
	Categories          string `index:"11" re:""`
	Accounts            string `index:"12" re:""`
	Subjects            string `index:"13" re:""`
	DestinationPrefixes string `index:"14" re:""`
	UsageInterval       string `index:"15" re:""`
	Suppliers           string `index:"16" re:""`
	DisconnectCauses    string `index:"17" re:""`
	MediationRunids     string `index:"18" re:""`
	RatedAccounts       string `index:"19" re:""`
	RatedSubjects       string `index:"20" re:""`
	CostInterval        string `index:"21" re:""`
	ActionTriggers      string `index:"22" re:""`
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
