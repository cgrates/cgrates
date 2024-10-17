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

package engine

import (
	"fmt"
	"strings"
	"time"

	"github.com/cgrates/cgrates/utils"
)

// Structs here are one to one mapping of the tables and fields
// to be used by gorm orm

type TimingMdl struct {
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

func (TimingMdl) TableName() string {
	return utils.TBLTPTimings
}

type DestinationMdl struct {
	Id        int64
	Tpid      string
	Tag       string `index:"0" re:".*"`
	Prefix    string `index:"1" re:".*"`
	CreatedAt time.Time
}

func (DestinationMdl) TableName() string {
	return utils.TBLTPDestinations
}

type RateMdl struct {
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

func (RateMdl) TableName() string {
	return utils.TBLTPRates
}

type DestinationRateMdl struct {
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

func (DestinationRateMdl) TableName() string {
	return utils.TBLTPDestinationRates
}

type RatingPlanMdl struct {
	Id           int64
	Tpid         string
	Tag          string  `index:"0" re:".*"`
	DestratesTag string  `index:"1" re:".*"`
	TimingTag    string  `index:"2" re:".*"`
	Weight       float64 `index:"3" re:".*"`
	CreatedAt    time.Time
}

func (RatingPlanMdl) TableName() string {
	return utils.TBLTPRatingPlans
}

type RatingProfileMdl struct {
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

func (RatingProfileMdl) TableName() string {
	return utils.TBLTPRatingProfiles
}

type ActionMdl struct {
	Id              int64
	Tpid            string
	Tag             string  `index:"0" re:".*"`
	Action          string  `index:"1" re:".*"`
	ExtraParameters string  `index:"2" re:".*"`
	Filters         string  `index:"3" re:".*"`
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

func (ActionMdl) TableName() string {
	return utils.TBLTPActions
}

type ActionPlanMdl struct {
	Id         int64
	Tpid       string
	Tag        string  `index:"0" re:".*"`
	ActionsTag string  `index:"1" re:".*"`
	TimingTag  string  `index:"2" re:".*"`
	Weight     float64 `index:"3" re:".*"`
	CreatedAt  time.Time
}

func (ActionPlanMdl) TableName() string {
	return utils.TBLTPActionPlans
}

type ActionTriggerMdl struct {
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

func (ActionTriggerMdl) TableName() string {
	return utils.TBLTPActionTriggers
}

type AccountActionMdl struct {
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

func (AccountActionMdl) TableName() string {
	return utils.TBLTPAccountActions
}

func (aa *AccountActionMdl) SetAccountActionId(id string) error {
	ids := strings.Split(id, utils.ConcatenatedKeySep)
	if len(ids) != 3 {
		return fmt.Errorf("Wrong TP Account Action Id: %s", id)
	}
	aa.Loadid = ids[0]
	aa.Tenant = ids[1]
	aa.Account = ids[2]
	return nil
}

func (aa *AccountActionMdl) GetAccountActionId() string {
	return utils.ConcatenatedKey(aa.Tenant, aa.Account)
}

type SharedGroupMdl struct {
	Id            int64
	Tpid          string
	Tag           string `index:"0" re:".*"`
	Account       string `index:"1" re:".*"`
	Strategy      string `index:"2" re:".*"`
	RatingSubject string `index:"3" re:".*"`
	CreatedAt     time.Time
}

func (SharedGroupMdl) TableName() string {
	return utils.TBLTPSharedGroups
}

type ResourceMdl struct {
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

func (ResourceMdl) TableName() string {
	return utils.TBLTPResources
}

type StatMdl struct {
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

func (StatMdl) TableName() string {
	return utils.TBLTPStats
}

type RankingsMdl struct {
	PK                uint `gorm:"primary_key"`
	Tpid              string
	Tenant            string `index:"0" re:".*"`
	ID                string `index:"1" re:".*"`
	Schedule          string `index:"2" re:".*"`
	StatIDs           string `index:"3" re:".*"`
	MetricIDs         string `index:"4" re:".*"`
	Sorting           string `index:"5" re:".*"`
	SortingParameters string `index:"6" re:".*"`
	Stored            bool   `index:"7" re:".*"`
	ThresholdIDs      string `index:"8" re:".*"`
	CreatedAt         time.Time
}

func (RankingsMdl) TableName() string {
	return utils.TBLTPRankings
}

type TrendsMdl struct {
	PK              uint `gorm:"primary_key"`
	Tpid            string
	Tenant          string  `index:"0" re:".*"`
	ID              string  `index:"1" re:".*"`
	Schedule        string  `index:"2" re:".*"`
	StatID          string  `index:"3" re:".*"`
	Metrics         string  `index:"4" re:".*"`
	TTL             string  `index:"5" re:".*"`
	QueueLength     int     `index:"6" re:".*"`
	MinItems        int     `index:"7" re:".*"`
	CorrelationType string  `index:"8" re:".*"`
	Tolerance       float64 `index:"9"  re:".*"`
	Stored          bool    `index:"10" re:".*"`
	ThresholdIDs    string  `index:"11" re:".*"`
	CreatedAt       time.Time
}

func (TrendsMdl) TableName() string {
	return utils.TBLTPTrends
}

type ThresholdMdl struct {
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

func (ThresholdMdl) TableName() string {
	return utils.TBLTPThresholds
}

type FilterMdl struct {
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

func (FilterMdl) TableName() string {
	return utils.TBLTPFilters
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
	AnswerTime  *time.Time
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

type RouteMdl struct {
	PK                 uint `gorm:"primary_key"`
	Tpid               string
	Tenant             string  `index:"0" re:".*"`
	ID                 string  `index:"1" re:".*"`
	FilterIDs          string  `index:"2" re:".*"`
	ActivationInterval string  `index:"3" re:".*"`
	Sorting            string  `index:"4" re:".*"`
	SortingParameters  string  `index:"5" re:".*"`
	RouteID            string  `index:"6" re:".*"`
	RouteFilterIDs     string  `index:"7" re:".*"`
	RouteAccountIDs    string  `index:"8" re:".*"`
	RouteRatingplanIDs string  `index:"9" re:".*"`
	RouteResourceIDs   string  `index:"10" re:".*"`
	RouteStatIDs       string  `index:"11" re:".*"`
	RouteWeight        float64 `index:"12" re:".*"`
	RouteBlocker       bool    `index:"13" re:".*"`
	RouteParameters    string  `index:"14" re:".*"`
	Weight             float64 `index:"15" re:".*"`
	CreatedAt          time.Time
}

func (RouteMdl) TableName() string {
	return utils.TBLTPRoutes
}

type AttributeMdl struct {
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

func (AttributeMdl) TableName() string {
	return utils.TBLTPAttributes
}

type ChargerMdl struct {
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

func (ChargerMdl) TableName() string {
	return utils.TBLTPChargers
}

type DispatcherProfileMdl struct {
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

func (DispatcherProfileMdl) TableName() string {
	return utils.TBLTPDispatchers
}

type DispatcherHostMdl struct {
	PK                   uint   `gorm:"primary_key"`
	Tpid                 string //
	Tenant               string `index:"0" re:".*"`
	ID                   string `index:"1" re:".*"`
	Address              string `index:"2" re:".*"`
	Transport            string `index:"3" re:".*"`
	ConnectAttempts      int    `index:"4" re:".*"`
	Reconnects           int    `index:"5" re:".*"`
	MaxReconnectInterval string `index:"6" re:".*"`
	ConnectTimeout       string `index:"7" re:".*"`
	ReplyTimeout         string `index:"8" re:".*"`
	TLS                  bool   `index:"9" re:".*"`
	ClientKey            string `index:"10" re:".*"`
	ClientCertificate    string `index:"11" re:".*"`
	CaCertificate        string `index:"12" re:".*"`
	CreatedAt            time.Time
}

func (DispatcherHostMdl) TableName() string {
	return utils.TBLTPDispatcherHosts
}
