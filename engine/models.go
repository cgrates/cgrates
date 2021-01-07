/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT MetaAny WARRANTY; without even the implied warranty of
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
	Tag       string `index:"0" re:"\w+\s*,\s*"`
	Years     string `index:"1" re:"\*any\s*,\s*|(?:\d{1,4};?)+\s*,\s*|\s*,\s*"`
	Months    string `index:"2" re:"\*any\s*,\s*|(?:\d{1,4};?)+\s*,\s*|\s*,\s*"`
	MonthDays string `index:"3" re:"\*any\s*,\s*|(?:\d{1,4};?)+\s*,\s*|\s*,\s*"`
	WeekDays  string `index:"4" re:"\*any\s*,\s*|(?:\d{1,4};?)+\s*,\s*|\s*,\s*"`
	Time      string `index:"5" re:"\d{2}:\d{2}:\d{2}|\*asap"`
	CreatedAt time.Time
}

func (TimingMdl) TableName() string {
	return utils.TBLTPTimings
}

type DestinationMdl struct {
	Id        int64
	Tpid      string
	Tag       string `index:"0" re:"\w+\s*,\s*"`
	Prefix    string `index:"1" re:"\+?\d+.?\d*"`
	CreatedAt time.Time
}

func (DestinationMdl) TableName() string {
	return utils.TBLTPDestinations
}

type RateMdl struct {
	Id                 int64
	Tpid               string
	Tag                string  `index:"0" re:"\w+\s*"`
	ConnectFee         float64 `index:"1" re:"\d+\.*\d*s*"`
	Rate               float64 `index:"2" re:"\d+\.*\d*s*"`
	RateUnit           string  `index:"3" re:"\d+\.*\d*(ns|us|µs|ms|s|m|h)*\s*"`
	RateIncrement      string  `index:"4" re:"\d+\.*\d*(ns|us|µs|ms|s|m|h)*\s*"`
	GroupIntervalStart string  `index:"5" re:"\d+\.*\d*(ns|us|µs|ms|s|m|h)*\s*"`
	CreatedAt          time.Time
}

func (RateMdl) TableName() string {
	return utils.TBLTPRates
}

type DestinationRateMdl struct {
	Id               int64
	Tpid             string
	Tag              string  `index:"0" re:"\w+\s*"`
	DestinationsTag  string  `index:"1" re:"\w+\s*|\*any"`
	RatesTag         string  `index:"2" re:"\w+\s*"`
	RoundingMethod   string  `index:"3" re:"\*up|\*down|\*middle"`
	RoundingDecimals int     `index:"4" re:"\d+"`
	MaxCost          float64 `index:"5" re:"\d+\.*\d*s*"`
	MaxCostStrategy  string  `index:"6" re:"\*free|\*disconnect"`
	CreatedAt        time.Time
}

func (DestinationRateMdl) TableName() string {
	return utils.TBLTPDestinationRates
}

type RatingPlanMdl struct {
	Id           int64
	Tpid         string
	Tag          string  `index:"0" re:"\w+\s*,\s*"`
	DestratesTag string  `index:"1" re:"\w+\s*,\s*|\*any"`
	TimingTag    string  `index:"2" re:"\w+\s*,\s*|\*any"`
	Weight       float64 `index:"3" re:"\d+.?\d*"`
	CreatedAt    time.Time
}

func (RatingPlanMdl) TableName() string {
	return utils.TBLTPRatingPlans
}

type RatingProfileMdl struct {
	Id               int64
	Tpid             string
	Loadid           string
	Tenant           string `index:"0" re:"[0-9A-Za-z_\.]+\s*"`
	Category         string `index:"1" re:"\w+\s*"`
	Subject          string `index:"2" re:"\*any\s*|(\w+;?)+\s*"`
	ActivationTime   string `index:"3" re:"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z"`
	RatingPlanTag    string `index:"4" re:"\w+\s*"`
	FallbackSubjects string `index:"5" re:"\w+\s*"`
	CreatedAt        time.Time
}

func (RatingProfileMdl) TableName() string {
	return utils.TBLTPRatingProfiles
}

type ActionMdl struct {
	Id              int64
	Tpid            string
	Tag             string  `index:"0" re:"\w+\s*"`
	Action          string  `index:"1" re:"\*\w+\s*"`
	ExtraParameters string  `index:"2" re:"\S+\s*"`
	Filter          string  `index:"3" re:"\S+\s*"`
	BalanceTag      string  `index:"4" re:"\w+\s*"`
	BalanceType     string  `index:"5" re:"\*\w+\s*"`
	Categories      string  `index:"6" re:""`
	DestinationTags string  `index:"7" re:"\*any|\w+\s*"`
	RatingSubject   string  `index:"8" re:"\w+\s*"`
	SharedGroups    string  `index:"9" re:"[0-9A-Za-z_;]*"`
	ExpiryTime      string  `index:"10" re:"\*\w+\s*|\+\d+[smh]\s*|\d+\s*"`
	TimingTags      string  `index:"11" re:"[0-9A-Za-z_;]*|\*any"`
	Units           string  `index:"12" re:"\d+\s*"`
	BalanceWeight   string  `index:"13" re:"\d+\.?\d*\s*"`
	BalanceBlocker  string  `index:"14" re:""`
	BalanceDisabled string  `index:"15" re:""`
	Weight          float64 `index:"16" re:"\d+\.?\d*\s*"`
	CreatedAt       time.Time
}

func (ActionMdl) TableName() string {
	return utils.TBLTPActions
}

type ActionPlanMdl struct {
	Id         int64
	Tpid       string
	Tag        string  `index:"0" re:"\w+\s*,\s*"`
	ActionsTag string  `index:"1" re:"\w+\s*,\s*"`
	TimingTag  string  `index:"2" re:"\w+\s*,\s*"|\*any`
	Weight     float64 `index:"3" re:"\d+\.?\d*"`
	CreatedAt  time.Time
}

func (ActionPlanMdl) TableName() string {
	return utils.TBLTPActionPlans
}

type ActionTriggerMdl struct {
	Id                     int64
	Tpid                   string
	Tag                    string  `index:"0" re:"\w+"`
	UniqueId               string  `index:"1" re:"\w+"`
	ThresholdType          string  `index:"2" re:"\*\w+"`
	ThresholdValue         float64 `index:"3" re:"\d+\.?\d*"`
	Recurrent              bool    `index:"4" re:"true|false|"`
	MinSleep               string  `index:"5" re:"\d+[smh]?"`
	ExpiryTime             string  `index:"6" re:""`
	ActivationTime         string  `index:"7" re:""`
	BalanceTag             string  `index:"8" re:"\w+\s*"`
	BalanceType            string  `index:"9" re:"\*\w+"`
	BalanceCategories      string  `index:"10" re:""`
	BalanceDestinationTags string  `index:"11" re:"\w+|\*any"`
	BalanceRatingSubject   string  `index:"12" re:"\w+|\*any"`
	BalanceSharedGroups    string  `index:"13" re:"\w+|\*any"`
	BalanceExpiryTime      string  `index:"14" re:"\*\w+\s*|\+\d+[smh]\s*|\d+\s*"`
	BalanceTimingTags      string  `index:"15" re:"[0-9A-Za-z_;]*|\*any"`
	BalanceWeight          string  `index:"16" re:"\d+\.?\d*"`
	BalanceBlocker         string  `index:"17" re:""`
	BalanceDisabled        string  `index:"18" re:""`
	ActionsTag             string  `index:"19" re:"\w+"`
	Weight                 float64 `index:"20" re:"\d+\.?\d*"`
	CreatedAt              time.Time
}

func (ActionTriggerMdl) TableName() string {
	return utils.TBLTPActionTriggers
}

type AccountActionMdl struct {
	Id                int64
	Tpid              string
	Loadid            string
	Tenant            string `index:"0" re:"\w+\s*"`
	Account           string `index:"1" re:"(\w+;?)+\s*"`
	ActionPlanTag     string `index:"2" re:"\w+\s*"`
	ActionTriggersTag string `index:"3" re:"\w+\s*"`
	AllowNegative     bool   `index:"4" re:""`
	Disabled          bool   `index:"5" re:""`
	CreatedAt         time.Time
}

func (AccountActionMdl) TableName() string {
	return utils.TBLTPAccountActions
}

func (aa *AccountActionMdl) SetAccountActionId(id string) error {
	ids := strings.Split(id, utils.CONCATENATED_KEY_SEP)
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
	Tag           string `index:"0" re:"\w+\s*"`
	Account       string `index:"1" re:"\*?\w+\s*"`
	Strategy      string `index:"2" re:"\*\w+\s*"`
	RatingSubject string `index:"3" re:"\*?\w]+\s*"`
	CreatedAt     time.Time
}

func (SharedGroupMdl) TableName() string {
	return utils.TBLTPSharedGroups
}

type ResourceMdl struct {
	PK                 uint `gorm:"primary_key"`
	Tpid               string
	Tenant             string  `index:"0" re:""`
	ID                 string  `index:"1" re:""`
	FilterIDs          string  `index:"2" re:""`
	ActivationInterval string  `index:"3" re:""`
	UsageTTL           string  `index:"4" re:""`
	Limit              string  `index:"5" re:""`
	AllocationMessage  string  `index:"6" re:""`
	Blocker            bool    `index:"7" re:""`
	Stored             bool    `index:"8" re:""`
	Weight             float64 `index:"9" re:"\d+\.?\d*"`
	ThresholdIDs       string  `index:"10" re:""`
	CreatedAt          time.Time
}

func (ResourceMdl) TableName() string {
	return utils.TBLTPResources
}

type StatMdl struct {
	PK                 uint `gorm:"primary_key"`
	Tpid               string
	Tenant             string  `index:"0" re:""`
	ID                 string  `index:"1" re:""`
	FilterIDs          string  `index:"2" re:""`
	ActivationInterval string  `index:"3" re:""`
	QueueLength        int     `index:"4" re:""`
	TTL                string  `index:"5" re:""`
	MinItems           int     `index:"6" re:""`
	MetricIDs          string  `index:"7" re:""`
	MetricFilterIDs    string  `index:"8" re:""`
	Stored             bool    `index:"9" re:""`
	Blocker            bool    `index:"10" re:""`
	Weight             float64 `index:"11" re:"\d+\.?\d*"`
	ThresholdIDs       string  `index:"12" re:""`
	CreatedAt          time.Time
}

func (StatMdl) TableName() string {
	return utils.TBLTPStats
}

type ThresholdMdl struct {
	PK                 uint `gorm:"primary_key"`
	Tpid               string
	Tenant             string  `index:"0" re:""`
	ID                 string  `index:"1" re:""`
	FilterIDs          string  `index:"2" re:""`
	ActivationInterval string  `index:"3" re:""`
	MaxHits            int     `index:"4" re:""`
	MinHits            int     `index:"5" re:""`
	MinSleep           string  `index:"6" re:""`
	Blocker            bool    `index:"7" re:""`
	Weight             float64 `index:"8" re:"\d+\.?\d*"`
	ActionIDs          string  `index:"9" re:""`
	Async              bool    `index:"10" re:""`
	CreatedAt          time.Time
}

func (ThresholdMdl) TableName() string {
	return utils.TBLTPThresholds
}

type FilterMdl struct {
	PK                 uint `gorm:"primary_key"`
	Tpid               string
	Tenant             string `index:"0" re:""`
	ID                 string `index:"1" re:""`
	Type               string `index:"2" re:"^\*[A-Za-z].*"`
	Element            string `index:"3" re:""`
	Values             string `index:"4" re:""`
	ActivationInterval string `index:"5" re:""`
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

func (t CDRsql) AsMapStringInterface() (out map[string]interface{}) {
	out = make(map[string]interface{})
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
	PK                  uint `gorm:"primary_key"`
	Tpid                string
	Tenant              string  `index:"0" re:""`
	ID                  string  `index:"1" re:""`
	FilterIDs           string  `index:"2" re:""`
	ActivationInterval  string  `index:"3" re:""`
	Sorting             string  `index:"4" re:""`
	SortingParameters   string  `index:"5" re:""`
	RouteID             string  `index:"6" re:""`
	RouteFilterIDs      string  `index:"7" re:""`
	RouteAccountIDs     string  `index:"8" re:""`
	RouteRatingplanIDs  string  `index:"9" re:""`
	RouteRateProfileIDs string  `index:"10" re:""`
	RouteResourceIDs    string  `index:"11" re:""`
	RouteStatIDs        string  `index:"12" re:""`
	RouteWeight         float64 `index:"13" re:"\d+\.?\d*"`
	RouteBlocker        bool    `index:"14" re:""`
	RouteParameters     string  `index:"15" re:""`
	Weight              float64 `index:"16" re:"\d+\.?\d*"`
	CreatedAt           time.Time
}

func (RouteMdl) TableName() string {
	return utils.TBLTPRoutes
}

type AttributeMdl struct {
	PK                 uint `gorm:"primary_key"`
	Tpid               string
	Tenant             string  `index:"0" re:""`
	ID                 string  `index:"1" re:""`
	Contexts           string  `index:"2" re:""`
	FilterIDs          string  `index:"3" re:""`
	ActivationInterval string  `index:"4" re:""`
	AttributeFilterIDs string  `index:"5" re:""`
	Path               string  `index:"6" re:""`
	Type               string  `index:"7" re:""`
	Value              string  `index:"8" re:""`
	Blocker            bool    `index:"9" re:""`
	Weight             float64 `index:"10" re:"\d+\.?\d*"`
	CreatedAt          time.Time
}

func (AttributeMdl) TableName() string {
	return utils.TBLTPAttributes
}

type ChargerMdl struct {
	PK                 uint `gorm:"primary_key"`
	Tpid               string
	Tenant             string  `index:"0" re:""`
	ID                 string  `index:"1" re:""`
	FilterIDs          string  `index:"2" re:""`
	ActivationInterval string  `index:"3" re:""`
	RunID              string  `index:"4" re:""`
	AttributeIDs       string  `index:"5" re:""`
	Weight             float64 `index:"6" re:"\d+\.?\d*"`
	CreatedAt          time.Time
}

func (ChargerMdl) TableName() string {
	return utils.TBLTPChargers
}

type DispatcherProfileMdl struct {
	PK                 uint    `gorm:"primary_key"`
	Tpid               string  //
	Tenant             string  `index:"0" re:""`
	ID                 string  `index:"1" re:""`
	Subsystems         string  `index:"2" re:""`
	FilterIDs          string  `index:"3" re:""`
	ActivationInterval string  `index:"4" re:""`
	Strategy           string  `index:"5" re:""`
	StrategyParameters string  `index:"6" re:""`
	ConnID             string  `index:"7" re:""`
	ConnFilterIDs      string  `index:"8" re:""`
	ConnWeight         float64 `index:"9" re:"\d+\.?\d*"`
	ConnBlocker        bool    `index:"10" re:""`
	ConnParameters     string  `index:"11" re:""`
	Weight             float64 `index:"12" re:"\d+\.?\d*"`
	CreatedAt          time.Time
}

func (DispatcherProfileMdl) TableName() string {
	return utils.TBLTPDispatchers
}

type DispatcherHostMdl struct {
	PK        uint   `gorm:"primary_key"`
	Tpid      string //
	Tenant    string `index:"0" re:""`
	ID        string `index:"1" re:""`
	Address   string `index:"2" re:""`
	Transport string `index:"3" re:""`
	TLS       bool   `index:"4" re:""`
	CreatedAt time.Time
}

func (DispatcherHostMdl) TableName() string {
	return utils.TBLTPDispatcherHosts
}

type RateProfileMdl struct {
	PK                  uint `gorm:"primary_key"`
	Tpid                string
	Tenant              string  `index:"0" re:""`
	ID                  string  `index:"1" re:""`
	FilterIDs           string  `index:"2" re:""`
	ActivationInterval  string  `index:"3" re:""`
	Weight              float64 `index:"4" re:"\d+\.?\d*"`
	RoundingMethod      string  `index:"5" re:""`
	RoundingDecimals    int     `index:"6" re:""`
	MinCost             float64 `index:"7"  re:"\d+\.?\d*""`
	MaxCost             float64 `index:"8"  re:"\d+\.?\d*"`
	MaxCostStrategy     string  `index:"9" re:""`
	RateID              string  `index:"10" re:""`
	RateFilterIDs       string  `index:"11" re:""`
	RateActivationTimes string  `index:"12" re:""`
	RateWeight          float64 `index:"13" re:"\d+\.?\d*"`
	RateBlocker         bool    `index:"14" re:""`
	RateIntervalStart   string  `index:"15" re:""`
	RateFixedFee        float64 `index:"16" re:"\d+\.?\d*"`
	RateRecurrentFee    float64 `index:"17" re:"\d+\.?\d*"`
	RateUnit            string  `index:"18" re:""`
	RateIncrement       string  `index:"19" re:""`

	CreatedAt time.Time
}

func (RateProfileMdl) TableName() string {
	return utils.TBLTPRateProfiles
}

type ActionProfileMdl struct {
	PK                 uint `gorm:"primary_key"`
	Tpid               string
	Tenant             string  `index:"0" re:""`
	ID                 string  `index:"1" re:""`
	FilterIDs          string  `index:"2" re:""`
	ActivationInterval string  `index:"3" re:""`
	Weight             float64 `index:"4" re:"\d+\.?\d*"`
	Schedule           string  `index:"5" re:""`
	TargetType         string  `index:"6" re:""`
	TargetIDs          string  `index:"7" re:""`
	ActionID           string  `index:"8" re:""`
	ActionFilterIDs    string  `index:"9" re:""`
	ActionBlocker      bool    `index:"10" re:""`
	ActionTTL          string  `index:"11" re:""`
	ActionType         string  `index:"12" re:""`
	ActionOpts         string  `index:"13" re:""`
	ActionPath         string  `index:"14" re:""`
	ActionValue        string  `index:"15" re:""`

	CreatedAt time.Time
}

func (ActionProfileMdl) TableName() string {
	return utils.TBLTPActionProfiles
}

type AccountProfileMdl struct {
	PK                    uint `gorm:"primary_key"`
	Tpid                  string
	Tenant                string  `index:"0" re:""`
	ID                    string  `index:"1" re:""`
	FilterIDs             string  `index:"2" re:""`
	ActivationInterval    string  `index:"3" re:""`
	Weight                float64 `index:"4" re:"\d+\.?\d*"`
	BalanceID             string  `index:"5" re:""`
	BalanceFilterIDs      string  `index:"6" re:""`
	BalanceWeight         float64 `index:"7" re:"\d+\.?\d*"`
	BalanceBlocker        bool    `index:"8" re:""`
	BalanceType           string  `index:"9" re:""`
	BalanceOpts           string  `index:"10" re:""`
	BalanceCostIncrements string  `index:"11" re:""`
	BalanceCostAttributes string  `index:"12" re:""`
	BalanceUnitFactors    string  `index:"13" re:""`
	BalanceUnits          float64 `index:"14" re:"\d+\.?\d*"`
	ThresholdIDs          string  `index:"15" re:""`
	CreatedAt             time.Time
}

func (AccountProfileMdl) TableName() string {
	return utils.TBLTPAccountProfiles
}
