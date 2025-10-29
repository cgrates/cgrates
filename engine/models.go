/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
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
	Tag       string `index:"0" re:".*"`
	Years     string `index:"1" re:".*"`
	Months    string `index:"2" re:".*"`
	MonthDays string `index:"3" re:".*"`
	WeekDays  string `index:"4" re:".*"`
	Time      string `index:"5" re:".*"`
	CreatedAt time.Time
}

type TpDestination struct {
	Id        int64
	Tpid      string
	Tag       string `index:"0" re:".*"`
	Prefix    string `index:"1" re:".*"`
	CreatedAt time.Time
}

type TpRate struct {
	Id                 int64
	Tpid               string
	Tag                string  `index:"0" re:".*"`
	ConnectFee         float64 `index:"1" re:".*"`
	Rate               float64 `index:"2" re:".*"`
	RateUnit           string  `index:"3" re:".*"`
	RateIncrement      string  `index:"4" re:".*"`
	GroupIntervalStart string  `index:"5" re:".*"`
	CreatedAt          time.Time
}

type TpDestinationRate struct {
	Id               int64
	Tpid             string
	Tag              string  `index:"0" re:".*"`
	DestinationsTag  string  `index:"1" re:".*"`
	RatesTag         string  `index:"2" re:".*"`
	RoundingMethod   string  `index:"3" re:".*"`
	RoundingDecimals int     `index:"4" re:".*"`
	MaxCost          float64 `index:"5" re:".*"`
	MaxCostStrategy  string  `index:"6" re:".*"`
	CreatedAt        time.Time
}

type TpRatingPlan struct {
	Id           int64
	Tpid         string
	Tag          string  `index:"0" re:".*"`
	DestratesTag string  `index:"1" re:".*"`
	TimingTag    string  `index:"2" re:".*"`
	Weight       float64 `index:"3" re:".*"`
	CreatedAt    time.Time
}

type TpRatingProfile struct {
	Id               int64
	Tpid             string
	Loadid           string
	Tenant           string `index:"0" re:".*"`
	Category         string `index:"1" re:".*"`
	Subject          string `index:"2" re:".*"`
	ActivationTime   string `index:"3" re:".*"`
	RatingPlanTag    string `index:"4" re:".*"`
	FallbackSubjects string `index:"5" re:".*"`
	CreatedAt        time.Time
}

type TpAction struct {
	Id              int64
	Tpid            string
	Tag             string  `index:"0" re:".*"`
	Action          string  `index:"1" re:".*"`
	ExtraParameters string  `index:"2" re:".*"`
	Filter          string  `index:"3" re:".*"`
	BalanceTag      string  `index:"4" re:".*"`
	BalanceType     string  `index:"5" re:".*"`
	Categories      string  `index:"6" re:".*"`
	DestinationTags string  `index:"7" re:".*"`
	RatingSubject   string  `index:"8" re:".*"`
	SharedGroups    string  `index:"9" re:".*"`
	ExpiryTime      string  `index:"10" re:".*"`
	TimingTags      string  `index:"11" re:".*"`
	Units           string  `index:"12" re:".*"`
	BalanceWeight   string  `index:"13" re:".*"`
	BalanceBlocker  string  `index:"14" re:".*"`
	BalanceDisabled string  `index:"15" re:".*"`
	Weight          float64 `index:"16" re:".*"`
	CreatedAt       time.Time
}

type TpActionPlan struct {
	Id         int64
	Tpid       string
	Tag        string  `index:"0" re:".*"`
	ActionsTag string  `index:"1" re:".*"`
	TimingTag  string  `index:"2" re:".*"`
	Weight     float64 `index:"3" re:".*"`
	CreatedAt  time.Time
}

type TpActionTrigger struct {
	Id                     int64
	Tpid                   string
	Tag                    string  `index:"0" re:".*"`
	UniqueId               string  `index:"1" re:".*"`
	ThresholdType          string  `index:"2" re:".*"`
	ThresholdValue         float64 `index:"3" re:".*"`
	Recurrent              bool    `index:"4" re:".*"`
	MinSleep               string  `index:"5" re:".*"`
	ExpiryTime             string  `index:"6" re:".*"`
	ActivationTime         string  `index:"7" re:".*"`
	BalanceTag             string  `index:"8" re:".*"`
	BalanceType            string  `index:"9" re:".*"`
	BalanceCategories      string  `index:"10" re:".*"`
	BalanceDestinationTags string  `index:"11" re:".*"`
	BalanceRatingSubject   string  `index:"12" re:".*"`
	BalanceSharedGroups    string  `index:"13" re:".*"`
	BalanceExpiryTime      string  `index:"14" re:".*"`
	BalanceTimingTags      string  `index:"15" re:".*"`
	BalanceWeight          string  `index:"16" re:".*"`
	BalanceBlocker         string  `index:"17" re:".*"`
	BalanceDisabled        string  `index:"18" re:".*"`
	ActionsTag             string  `index:"19" re:".*"`
	Weight                 float64 `index:"20" re:".*"`
	CreatedAt              time.Time
}

type TpAccountAction struct {
	Id                int64
	Tpid              string
	Loadid            string
	Tenant            string `index:"0" re:".*"`
	Account           string `index:"1" re:".*"`
	ActionPlanTag     string `index:"2" re:".*"`
	ActionTriggersTag string `index:"3" re:".*"`
	AllowNegative     bool   `index:"4" re:".*"`
	Disabled          bool   `index:"5" re:".*"`
	CreatedAt         time.Time
}

func (aa *TpAccountAction) SetAccountActionId(id string) error {
	ids := strings.Split(id, utils.CONCATENATED_KEY_SEP)
	if len(ids) != 3 {
		return fmt.Errorf("Wrong TP Account Action Id: %s", id)
	}
	aa.Loadid = ids[0]
	aa.Tenant = ids[1]
	aa.Account = ids[2]
	return nil
}

func (aa *TpAccountAction) GetAccountActionId() string {
	return utils.ConcatenatedKey(aa.Tenant, aa.Account)
}

type TpSharedGroup struct {
	Id            int64
	Tpid          string
	Tag           string `index:"0" re:".*"`
	Account       string `index:"1" re:".*"`
	Strategy      string `index:"2" re:".*"`
	RatingSubject string `index:"3" re:".*"`
	CreatedAt     time.Time
}

type TpResource struct {
	PK                 uint `gorm:"primary_key"`
	Tpid               string
	Tenant             string  `index:"0" re:".*"`
	ID                 string  `index:"1" re:".*"`
	FilterIDs          string  `index:"2" re:".*"`
	ActivationInterval string  `index:"3" re:".*"`
	UsageTTL           string  `index:"4" re:".*"`
	Limit              string  `index:"5" re:".*"`
	AllocationMessage  string  `index:"6" re:".*"`
	Blocker            bool    `index:"7" re:".*"`
	Stored             bool    `index:"8" re:".*"`
	Weight             float64 `index:"9" re:".*"`
	ThresholdIDs       string  `index:"10" re:".*"`
	CreatedAt          time.Time
}

type TpStat struct {
	PK                 uint `gorm:"primary_key"`
	Tpid               string
	Tenant             string  `index:"0" re:".*"`
	ID                 string  `index:"1" re:".*"`
	FilterIDs          string  `index:"2" re:".*"`
	ActivationInterval string  `index:"3" re:".*"`
	QueueLength        int     `index:"4" re:".*"`
	TTL                string  `index:"5" re:".*"`
	MinItems           int     `index:"6" re:".*"`
	MetricIDs          string  `index:"7" re:".*"`
	MetricFilterIDs    string  `index:"8" re:".*"`
	Stored             bool    `index:"9" re:".*"`
	Blocker            bool    `index:"10" re:".*"`
	Weight             float64 `index:"11" re:".*"`
	ThresholdIDs       string  `index:"12" re:".*"`
	CreatedAt          time.Time
}

type TpThreshold struct {
	PK                 uint `gorm:"primary_key"`
	Tpid               string
	Tenant             string  `index:"0" re:".*"`
	ID                 string  `index:"1" re:".*"`
	FilterIDs          string  `index:"2" re:".*"`
	ActivationInterval string  `index:"3" re:".*"`
	MaxHits            int     `index:"4" re:".*"`
	MinHits            int     `index:"5" re:".*"`
	MinSleep           string  `index:"6" re:".*"`
	Blocker            bool    `index:"7" re:".*"`
	Weight             float64 `index:"8" re:".*"`
	ActionIDs          string  `index:"9" re:".*"`
	Async              bool    `index:"10" re:".*"`
	CreatedAt          time.Time
}

type TpFilter struct {
	PK                 uint `gorm:"primary_key"`
	Tpid               string
	Tenant             string `index:"0" re:".*"`
	ID                 string `index:"1" re:".*"`
	Type               string `index:"2" re:".*"`
	Element            string `index:"3" re:".*"`
	Values             string `index:"4" re:".*"`
	ActivationInterval string `index:"5" re:".*"`
	CreatedAt          time.Time
}

type CDRsql struct {
	ID          int64
	Cgrid       string
	RunID       string
	OriginHost  string
	Source      string
	OriginID    string
	TOR         string
	RequestType string
	Tenant      string
	Category    string
	Account     string
	Subject     string
	Destination string
	SetupTime   time.Time
	AnswerTime  time.Time
	Usage       int64
	ExtraFields string
	CostSource  string
	Cost        float64
	CostDetails string
	ExtraInfo   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

func (t CDRsql) TableName() string {
	return utils.CDRsTBL
}

func (t CDRsql) AsMapStringInterface() (out map[string]any) {
	out = make(map[string]any)
	// out["id"] = t.ID // ignore ID
	out["cgrid"] = t.Cgrid
	out["run_id"] = t.RunID
	out["origin_host"] = t.OriginHost
	out["source"] = t.Source
	out["origin_id"] = t.OriginID
	out["tor"] = t.TOR
	out["request_type"] = t.RequestType
	out["tenant"] = t.Tenant
	out["category"] = t.Category
	out["account"] = t.Account
	out["subject"] = t.Subject
	out["destination"] = t.Destination
	out["setup_time"] = t.SetupTime
	out["answer_time"] = t.AnswerTime
	out["usage"] = t.Usage
	out["extra_fields"] = t.ExtraFields
	out["cost_source"] = t.CostSource
	out["cost"] = t.Cost
	out["cost_details"] = t.CostDetails
	out["extra_info"] = t.ExtraInfo
	out["created_at"] = t.CreatedAt
	out["updated_at"] = t.UpdatedAt
	// out["deleted_at"] = t.DeletedAt // ignore DeletedAt
	return

}

type SessionCostsSQL struct {
	ID          int64
	Cgrid       string
	RunID       string
	OriginHost  string
	OriginID    string
	CostSource  string
	Usage       int64
	CostDetails string
	CreatedAt   time.Time
	DeletedAt   *time.Time
}

func (t SessionCostsSQL) TableName() string {
	return utils.SessionCostsTBL
}

type TBLVersion struct {
	ID      uint
	Item    string
	Version int64
}

func (t TBLVersion) TableName() string {
	return utils.TBLVersions
}

type TpSupplier struct {
	PK                    uint `gorm:"primary_key"`
	Tpid                  string
	Tenant                string  `index:"0" re:".*"`
	ID                    string  `index:"1" re:".*"`
	FilterIDs             string  `index:"2" re:".*"`
	ActivationInterval    string  `index:"3" re:".*"`
	Sorting               string  `index:"4" re:".*"`
	SortingParameters     string  `index:"5" re:".*"`
	SupplierID            string  `index:"6" re:".*"`
	SupplierFilterIDs     string  `index:"7" re:".*"`
	SupplierAccountIDs    string  `index:"8" re:".*"`
	SupplierRatingplanIDs string  `index:"9" re:".*"`
	SupplierResourceIDs   string  `index:"10" re:".*"`
	SupplierStatIDs       string  `index:"11" re:".*"`
	SupplierWeight        float64 `index:"12" re:".*"`
	SupplierBlocker       bool    `index:"13" re:".*"`
	SupplierParameters    string  `index:"14" re:".*"`
	Weight                float64 `index:"15" re:".*"`
	CreatedAt             time.Time
}

type TPAttribute struct {
	PK                 uint `gorm:"primary_key"`
	Tpid               string
	Tenant             string  `index:"0" re:".*"`
	ID                 string  `index:"1" re:".*"`
	Contexts           string  `index:"2" re:".*"`
	FilterIDs          string  `index:"3" re:".*"`
	ActivationInterval string  `index:"4" re:".*"`
	AttributeFilterIDs string  `index:"5" re:".*"`
	Path               string  `index:"6" re:".*"`
	Type               string  `index:"7" re:".*"`
	Value              string  `index:"8" re:".*"`
	Blocker            bool    `index:"9" re:".*"`
	Weight             float64 `index:"10" re:".*"`
	CreatedAt          time.Time
}

type TPCharger struct {
	PK                 uint `gorm:"primary_key"`
	Tpid               string
	Tenant             string  `index:"0" re:".*"`
	ID                 string  `index:"1" re:".*"`
	FilterIDs          string  `index:"2" re:".*"`
	ActivationInterval string  `index:"3" re:".*"`
	RunID              string  `index:"4" re:".*"`
	AttributeIDs       string  `index:"5" re:".*"`
	Weight             float64 `index:"6" re:".*"`
	CreatedAt          time.Time
}

type TPDispatcherProfile struct {
	PK                 uint    `gorm:"primary_key"`
	Tpid               string  //
	Tenant             string  `index:"0" re:".*"`
	ID                 string  `index:"1" re:".*"`
	Subsystems         string  `index:"2" re:".*"`
	FilterIDs          string  `index:"3" re:".*"`
	ActivationInterval string  `index:"4" re:".*"`
	Strategy           string  `index:"5" re:".*"`
	StrategyParameters string  `index:"6" re:".*"`
	ConnID             string  `index:"7" re:".*"`
	ConnFilterIDs      string  `index:"8" re:".*"`
	ConnWeight         float64 `index:"9" re:".*"`
	ConnBlocker        bool    `index:"10" re:".*"`
	ConnParameters     string  `index:"11" re:".*"`
	Weight             float64 `index:"12" re:".*"`
	CreatedAt          time.Time
}

type TPDispatcherHost struct {
	PK        uint   `gorm:"primary_key"`
	Tpid      string //
	Tenant    string `index:"0" re:".*"`
	ID        string `index:"1" re:".*"`
	Address   string `index:"2" re:".*"`
	Transport string `index:"3" re:".*"`
	TLS       bool   `index:"4" re:".*"`
	CreatedAt time.Time
}
