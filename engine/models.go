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
	"time"

	"github.com/cgrates/cgrates/utils"
)

// Structs here are one to one mapping of the tables and fields
// to be used by gorm orm

type ResourceMdl struct {
	PK                uint `gorm:"primary_key"`
	Tpid              string
	Tenant            string  `index:"0" re:""`
	ID                string  `index:"1" re:""`
	FilterIDs         string  `index:"2" re:""`
	Weight            float64 `index:"3" re:"\d+\.?\d*"`
	UsageTTL          string  `index:"4" re:""`
	Limit             string  `index:"5" re:""`
	AllocationMessage string  `index:"6" re:""`
	Blocker           bool    `index:"7" re:""`
	Stored            bool    `index:"8" re:""`
	ThresholdIDs      string  `index:"9" re:""`
	CreatedAt         time.Time
}

func (ResourceMdl) TableName() string {
	return utils.TBLTPResources
}

type StatMdl struct {
	PK              uint `gorm:"primary_key"`
	Tpid            string
	Tenant          string  `index:"0" re:""`
	ID              string  `index:"1" re:""`
	FilterIDs       string  `index:"2" re:""`
	Weight          float64 `index:"3" re:"\d+\.?\d*"`
	QueueLength     int     `index:"4" re:""`
	TTL             string  `index:"5" re:""`
	MinItems        int     `index:"6" re:""`
	MetricIDs       string  `index:"7" re:""`
	MetricFilterIDs string  `index:"8" re:""`
	Stored          bool    `index:"9" re:""`
	Blocker         bool    `index:"10" re:""`
	ThresholdIDs    string  `index:"11" re:""`
	CreatedAt       time.Time
}

func (StatMdl) TableName() string {
	return utils.TBLTPStats
}

type ThresholdMdl struct {
	PK               uint `gorm:"primary_key"`
	Tpid             string
	Tenant           string  `index:"0" re:""`
	ID               string  `index:"1" re:""`
	FilterIDs        string  `index:"2" re:""`
	Weight           float64 `index:"3" re:"\d+\.?\d*"`
	MaxHits          int     `index:"4" re:""`
	MinHits          int     `index:"5" re:""`
	MinSleep         string  `index:"6" re:""`
	Blocker          bool    `index:"7" re:""`
	ActionProfileIDs string  `index:"8" re:""`
	Async            bool    `index:"9" re:""`
	CreatedAt        time.Time
}

func (ThresholdMdl) TableName() string {
	return utils.TBLTPThresholds
}

type FilterMdl struct {
	PK        uint `gorm:"primary_key"`
	Tpid      string
	Tenant    string `index:"0" re:""`
	ID        string `index:"1" re:""`
	Type      string `index:"2" re:"^\*[A-Za-z].*"`
	Element   string `index:"3" re:""`
	Values    string `index:"4" re:""`
	CreatedAt time.Time
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
	Tenant              string `index:"0" re:""`
	ID                  string `index:"1" re:""`
	FilterIDs           string `index:"2" re:""`
	Weights             string `index:"3" re:""`
	Sorting             string `index:"4" re:""`
	SortingParameters   string `index:"5" re:""`
	RouteID             string `index:"6" re:""`
	RouteFilterIDs      string `index:"7" re:""`
	RouteAccountIDs     string `index:"8" re:""`
	RouteRateProfileIDs string `index:"9" re:""`
	RouteResourceIDs    string `index:"10" re:""`
	RouteStatIDs        string `index:"11" re:""`
	RouteWeights        string `index:"12" re:""`
	RouteBlocker        bool   `index:"13" re:""`
	RouteParameters     string `index:"14" re:""`
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
	FilterIDs          string  `index:"2" re:""`
	Weight             float64 `index:"3" re:"\d+\.?\d*"`
	AttributeFilterIDs string  `index:"4" re:""`
	Path               string  `index:"5" re:""`
	Type               string  `index:"6" re:""`
	Value              string  `index:"7" re:""`
	Blocker            bool    `index:"8" re:""`
	CreatedAt          time.Time
}

func (AttributeMdl) TableName() string {
	return utils.TBLTPAttributes
}

type ChargerMdl struct {
	PK           uint `gorm:"primary_key"`
	Tpid         string
	Tenant       string  `index:"0" re:""`
	ID           string  `index:"1" re:""`
	FilterIDs    string  `index:"2" re:""`
	Weight       float64 `index:"3" re:"\d+\.?\d*"`
	RunID        string  `index:"4" re:""`
	AttributeIDs string  `index:"5" re:""`
	CreatedAt    time.Time
}

func (ChargerMdl) TableName() string {
	return utils.TBLTPChargers
}

type DispatcherProfileMdl struct {
	PK                 uint    `gorm:"primary_key"`
	Tpid               string  //
	Tenant             string  `index:"0" re:""`
	ID                 string  `index:"1" re:""`
	FilterIDs          string  `index:"2" re:""`
	Weight             float64 `index:"3" re:"\d+\.?\d*"`
	Strategy           string  `index:"4" re:""`
	StrategyParameters string  `index:"5" re:""`
	ConnID             string  `index:"6" re:""`
	ConnFilterIDs      string  `index:"7" re:""`
	ConnWeight         float64 `index:"8" re:"\d+\.?\d*"`
	ConnBlocker        bool    `index:"9" re:""`
	ConnParameters     string  `index:"10" re:""`
	CreatedAt          time.Time
}

func (DispatcherProfileMdl) TableName() string {
	return utils.TBLTPDispatchers
}

type DispatcherHostMdl struct {
	PK                uint   `gorm:"primary_key"`
	Tpid              string //
	Tenant            string `index:"0" re:""`
	ID                string `index:"1" re:""`
	Address           string `index:"2" re:""`
	Transport         string `index:"3" re:""`
	ConnectAttempts   int    `index:"4" re:""`
	Reconnects        int    `index:"5" re:""`
	ConnectTimeout    string `index:"6" re:""`
	ReplyTimeout      string `index:"7" re:""`
	TLS               bool   `index:"8" re:""`
	ClientKey         string `index:"9" re:""`
	ClientCertificate string `index:"10" re:""`
	CaCertificate     string `index:"11" re:""`
	CreatedAt         time.Time
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
	Weights             string  `index:"3" re:""`
	MinCost             float64 `index:"4"  re:"\d+\.?\d*""`
	MaxCost             float64 `index:"5"  re:"\d+\.?\d*"`
	MaxCostStrategy     string  `index:"6" re:""`
	RateID              string  `index:"7" re:""`
	RateFilterIDs       string  `index:"8" re:""`
	RateActivationTimes string  `index:"9" re:""`
	RateWeights         string  `index:"10" re:""`
	RateBlocker         bool    `index:"11" re:""`
	RateIntervalStart   string  `index:"12" re:""`
	RateFixedFee        float64 `index:"13" re:"\d+\.?\d*"`
	RateRecurrentFee    float64 `index:"14" re:"\d+\.?\d*"`
	RateUnit            string  `index:"15" re:""`
	RateIncrement       string  `index:"16" re:""`

	CreatedAt time.Time
}

func (RateProfileMdl) TableName() string {
	return utils.TBLTPRateProfiles
}

type ActionProfileMdl struct {
	PK              uint `gorm:"primary_key"`
	Tpid            string
	Tenant          string  `index:"0" re:""`
	ID              string  `index:"1" re:""`
	FilterIDs       string  `index:"2" re:""`
	Weight          float64 `index:"3" re:"\d+\.?\d*"`
	Schedule        string  `index:"4" re:""`
	TargetType      string  `index:"5" re:""`
	TargetIDs       string  `index:"6" re:""`
	ActionID        string  `index:"7" re:""`
	ActionFilterIDs string  `index:"8" re:""`
	ActionBlocker   bool    `index:"9" re:""`
	ActionTTL       string  `index:"10" re:""`
	ActionType      string  `index:"11" re:""`
	ActionOpts      string  `index:"12" re:""`
	ActionPath      string  `index:"13" re:""`
	ActionValue     string  `index:"14" re:""`

	CreatedAt time.Time
}

func (ActionProfileMdl) TableName() string {
	return utils.TBLTPActionProfiles
}

type AccountMdl struct {
	PK                    uint `gorm:"primary_key"`
	Tpid                  string
	Tenant                string  `index:"0" re:""`
	ID                    string  `index:"1" re:""`
	FilterIDs             string  `index:"2" re:""`
	Weights               string  `index:"3" re:""`
	Opts                  string  `index:"4" re:""`
	BalanceID             string  `index:"5" re:""`
	BalanceFilterIDs      string  `index:"6" re:""`
	BalanceWeights        string  `index:"7" re:""`
	BalanceType           string  `index:"8" re:""`
	BalanceUnits          float64 `index:"9" re:"\d+\.?\d*"`
	BalanceUnitFactors    string  `index:"10" re:""`
	BalanceOpts           string  `index:"11" re:""`
	BalanceCostIncrements string  `index:"12" re:""`
	BalanceAttributeIDs   string  `index:"13" re:""`
	BalanceRateProfileIDs string  `index:"14" re:""`
	ThresholdIDs          string  `index:"15" re:""`
	CreatedAt             time.Time
}

func (AccountMdl) TableName() string {
	return utils.TBLTPAccounts
}
