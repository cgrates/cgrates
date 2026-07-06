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
	"time"

	"github.com/cgrates/cgrates/utils"
)

type ResourceMdl struct {
	PK                uint `gorm:"primary_key"`
	Tpid              string
	Tenant            string `index:"0" re:".*"`
	ID                string `index:"1" re:".*"`
	FilterIDs         string `index:"2" re:".*"`
	Weights           string `index:"3" re:".*"`
	UsageTTL          string `index:"4" re:".*"`
	Limit             string `index:"5" re:".*"`
	AllocationMessage string `index:"6" re:".*"`
	Blocker           bool   `index:"7" re:".*"`
	Stored            bool   `index:"8" re:".*"`
	ThresholdIDs      string `index:"9" re:".*"`
	CreatedAt         time.Time
}

func (ResourceMdl) TableName() string {
	return utils.TBLTPResources
}

type IPMdl struct {
	PK            uint `gorm:"primary_key"`
	Tpid          string
	Tenant        string `index:"0" re:".*"`
	ID            string `index:"1" re:".*"`
	FilterIDs     string `index:"2" re:".*"`
	Weights       string `index:"3" re:".*"`
	TTL           string `index:"4" re:".*"`
	Stored        bool   `index:"5" re:".*"`
	PoolID        string `index:"6" re:".*"`
	PoolFilterIDs string `index:"7" re:".*"`
	PoolType      string `index:"8" re:".*"`
	PoolRange     string `index:"9" re:".*"`
	PoolStrategy  string `index:"10" re:".*"`
	PoolMessage   string `index:"11" re:".*"`
	PoolWeights   string `index:"12" re:".*"`
	PoolBlockers  string `index:"13" re:".*"`
	CreatedAt     time.Time
}

func (IPMdl) TableName() string {
	return utils.TBLTPIPs
}

type StatMdl struct {
	PK              uint `gorm:"primary_key"`
	Tpid            string
	Tenant          string `index:"0" re:".*"`
	ID              string `index:"1" re:".*"`
	FilterIDs       string `index:"2" re:".*"`
	Weights         string `index:"3" re:".*"`
	Blockers        string `index:"4" re:".*"`
	QueueLength     int    `index:"5" re:".*"`
	TTL             string `index:"6" re:".*"`
	MinItems        int    `index:"7" re:".*"`
	Stored          bool   `index:"8" re:".*"`
	ThresholdIDs    string `index:"9" re:".*"`
	MetricIDs       string `index:"10" re:".*"`
	MetricFilterIDs string `index:"11" re:".*"`
	MetricBlockers  string `index:"12" re:".*"`
	CreatedAt       time.Time
}

func (StatMdl) TableName() string {
	return utils.TBLTPStats
}

type RankingMdl struct {
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

func (RankingMdl) TableName() string {
	return utils.TBLTPRankings
}

type TrendMdl struct {
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

func (TrendMdl) TableName() string {
	return utils.TBLTPTrends
}

type ThresholdMdl struct {
	PK               uint `gorm:"primary_key"`
	Tpid             string
	Tenant           string `index:"0" re:".*"`
	ID               string `index:"1" re:".*"`
	FilterIDs        string `index:"2" re:".*"`
	Weights          string `index:"3" re:".*"`
	MaxHits          int    `index:"4" re:".*"`
	MinHits          int    `index:"5" re:".*"`
	MinSleep         string `index:"6" re:".*"`
	Blocker          bool   `index:"7" re:".*"`
	AttributeIDs     string `index:"8"   re:".*"`
	ActionProfileIDs string `index:"9" re:".*"`
	Async            bool   `index:"10" re:".*"`
	EeIDs            string `index:"11"  re:".*"`

	CreatedAt time.Time
}

func (ThresholdMdl) TableName() string {
	return utils.TBLTPThresholds
}

type FilterMdl struct {
	PK        uint `gorm:"primary_key"`
	Tpid      string
	Tenant    string `index:"0" re:".*"`
	ID        string `index:"1" re:".*"`
	Type      string `index:"2" re:".*"`
	Element   string `index:"3" re:".*"`
	Values    string `index:"4" re:".*"`
	CreatedAt time.Time
}

func (FilterMdl) TableName() string {
	return utils.TBLTPFilters
}

type RouteMdl struct {
	PK                  uint `gorm:"primary_key"`
	Tpid                string
	Tenant              string `index:"0" re:".*"`
	ID                  string `index:"1" re:".*"`
	FilterIDs           string `index:"2" re:".*"`
	Weights             string `index:"3" re:".*"`
	Blockers            string `index:"4" re:".*"`
	Sorting             string `index:"5" re:".*"`
	SortingParameters   string `index:"6" re:".*"`
	RouteID             string `index:"7" re:".*"`
	RouteFilterIDs      string `index:"8" re:".*"`
	RouteAccountIDs     string `index:"9" re:".*"`
	RouteRateProfileIDs string `index:"10" re:".*"`
	RouteResourceIDs    string `index:"11" re:".*"`
	RouteStatIDs        string `index:"12" re:".*"`
	RouteWeights        string `index:"13" re:".*"`
	RouteBlockers       string `index:"14" re:".*"`
	RouteParameters     string `index:"15" re:".*"`
	CreatedAt           time.Time
}

func (RouteMdl) TableName() string {
	return utils.TBLTPRoutes
}

type AttributeMdl struct {
	PK                 uint `gorm:"primary_key"`
	Tpid               string
	Tenant             string `index:"0" re:".*"`
	ID                 string `index:"1" re:".*"`
	FilterIDs          string `index:"2" re:".*"`
	Weights            string `index:"3" re:".*"`
	Blockers           string `index:"4" re:".*"`
	AttributeFilterIDs string `index:"5" re:".*"`
	AttributeBlockers  string `index:"6" re:".*"`
	Path               string `index:"7" re:".*"`
	Type               string `index:"8" re:".*"`
	Value              string `index:"9" re:".*"`
	CreatedAt          time.Time
}

func (AttributeMdl) TableName() string {
	return utils.TBLTPAttributes
}

type ChargerMdl struct {
	PK           uint `gorm:"primary_key"`
	Tpid         string
	Tenant       string `index:"0" re:".*"`
	ID           string `index:"1" re:".*"`
	FilterIDs    string `index:"2" re:".*"`
	Weights      string `index:"3" re:".*"`
	Blockers     string `index:"4" re:".*"`
	RunID        string `index:"5" re:".*"`
	AttributeIDs string `index:"6" re:".*"`
	CreatedAt    time.Time
}

func (ChargerMdl) TableName() string {
	return utils.TBLTPChargers
}

type RateProfileMdl struct {
	PK                  uint `gorm:"primary_key"`
	Tpid                string
	Tenant              string  `index:"0" re:".*"`
	ID                  string  `index:"1" re:".*"`
	FilterIDs           string  `index:"2" re:".*"`
	Weights             string  `index:"3" re:".*"`
	MinCost             float64 `index:"4"  re:".*"`
	MaxCost             float64 `index:"5"  re:".*"`
	MaxCostStrategy     string  `index:"6" re:".*"`
	RateID              string  `index:"7" re:".*"`
	RateFilterIDs       string  `index:"8" re:".*"`
	RateActivationTimes string  `index:"9" re:".*"`
	RateWeights         string  `index:"10" re:".*"`
	RateBlocker         bool    `index:"11" re:".*"`
	RateIntervalStart   string  `index:"12" re:".*"`
	RateFixedFee        float64 `index:"13" re:".*"`
	RateRecurrentFee    float64 `index:"14" re:".*"`
	RateUnit            string  `index:"15" re:".*"`
	RateIncrement       string  `index:"16" re:".*"`

	CreatedAt time.Time
}

func (RateProfileMdl) TableName() string {
	return utils.TBLTPRateProfiles
}

type ActionProfileMdl struct {
	PK                     uint `gorm:"primary_key"`
	Tpid                   string
	Tenant                 string `index:"0" re:".*"`
	ID                     string `index:"1" re:".*"`
	FilterIDs              string `index:"2" re:".*"`
	Weights                string `index:"3" re:".*"`
	Blockers               string `index:"4" re:".*"`
	Schedule               string `index:"5" re:".*"`
	TargetType             string `index:"6" re:".*"`
	TargetIDs              string `index:"7" re:".*"`
	ActionID               string `index:"8" re:".*"`
	ActionFilterIDs        string `index:"9" re:".*"`
	ActionTTL              string `index:"10" re:".*"`
	ActionType             string `index:"11" re:".*"`
	ActionOpts             string `index:"12" re:".*"`
	ActionWeights          string `index:"13" re:".*"`
	ActionBlockers         string `index:"14" re:".*"`
	ActionDiktatsID        string `index:"15" re:".*"`
	ActionDiktatsFilterIDs string `index:"16" re:".*"`
	ActionDiktatsOpts      string `index:"17" re:".*"`
	ActionDiktatsWeights   string `index:"18" re:".*"`
	ActionDiktatsBlockers  string `index:"19" re:".*"`

	CreatedAt time.Time
}

func (ActionProfileMdl) TableName() string {
	return utils.TBLTPActionProfiles
}

type AccountMdl struct {
	PK                    uint `gorm:"primary_key"`
	Tpid                  string
	Tenant                string `index:"0" re:".*"`
	ID                    string `index:"1" re:".*"`
	FilterIDs             string `index:"2" re:".*"`
	Weights               string `index:"3" re:".*"`
	Blockers              string `index:"4" re:".*"`
	Opts                  string `index:"5" re:".*"`
	BalanceID             string `index:"6" re:".*"`
	BalanceFilterIDs      string `index:"7" re:".*"`
	BalanceWeights        string `index:"8" re:".*"`
	BalanceBlockers       string `index:"9" re:".*"`
	BalanceType           string `index:"10" re:".*"`
	BalanceUnits          string `index:"11" re:".*"`
	BalanceUnitFactors    string `index:"12" re:".*"`
	BalanceOpts           string `index:"13" re:".*"`
	BalanceCostIncrements string `index:"14" re:".*"`
	BalanceAttributeIDs   string `index:"15" re:".*"`
	BalanceRateProfileIDs string `index:"16" re:".*"`
	ThresholdIDs          string `index:"17" re:".*"`
	CreatedAt             time.Time
}

func (AccountMdl) TableName() string {
	return utils.TBLTPAccounts
}
