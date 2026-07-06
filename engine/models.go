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

// Structs here are one to one mapping of the tables and fields
// to be used by gorm orm

type CDRsql struct {
	ID          int64
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

	out["run_id"] = t.RunID
	out["originHost"] = t.OriginHost
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
	out["extraFields"] = t.ExtraFields
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

type AccountJSONMdl struct {
	PK      uint        `gorm:"primary_key"`
	Tenant  string      `index:"0" re:".*"`
	ID      string      `index:"1" re:".*"`
	Account utils.JSONB `gorm:"type:jsonb" index:"2" re:".*"`
}

func (AccountJSONMdl) TableName() string {
	return utils.TBLAccounts
}

type IPProfileMdl struct {
	PK        uint        `gorm:"primary_key"`
	Tenant    string      `index:"0" re:".*"`
	ID        string      `index:"1" re:".*"`
	IPProfile utils.JSONB `gorm:"type:jsonb" index:"2" re:".*"`
}

func (IPProfileMdl) TableName() string {
	return utils.TBLIPProfiles
}

type IPAllocationMdl struct {
	PK           uint        `gorm:"primary_key"`
	Tenant       string      `index:"0" re:".*"`
	ID           string      `index:"1" re:".*"`
	IPAllocation utils.JSONB `gorm:"type:jsonb" index:"2" re:".*"`
}

func (IPAllocationMdl) TableName() string {
	return utils.TBLIPAllocations
}

type ActionProfileJSONMdl struct {
	PK            uint        `gorm:"primary_key"`
	Tenant        string      `index:"0" re:".*"`
	ID            string      `index:"1" re:".*"`
	ActionProfile utils.JSONB `gorm:"type:jsonb" index:"2" re:".*"`
}

func (ActionProfileJSONMdl) TableName() string {
	return utils.TBLActionProfiles
}

type ChargerProfileMdl struct {
	PK             uint        `gorm:"primary_key"`
	Tenant         string      `index:"0" re:".*"`
	ID             string      `index:"1" re:".*"`
	ChargerProfile utils.JSONB `gorm:"type:jsonb" index:"2" re:".*"`
}

func (ChargerProfileMdl) TableName() string {
	return utils.TBLChargerProfiles
}

type AttributeProfileMdl struct {
	PK               uint        `gorm:"primary_key"`
	Tenant           string      `index:"0" re:".*"`
	ID               string      `index:"1" re:".*"`
	AttributeProfile utils.JSONB `gorm:"type:jsonb" index:"2" re:".*"`
}

func (AttributeProfileMdl) TableName() string {
	return utils.TBLAttributeProfiles
}

type ResourceProfileMdl struct {
	PK              uint        `gorm:"primary_key"`
	Tenant          string      `index:"0" re:".*"`
	ID              string      `index:"1" re:".*"`
	ResourceProfile utils.JSONB `gorm:"type:jsonb" index:"2" re:".*"`
}

func (ResourceProfileMdl) TableName() string {
	return utils.TBLResourceProfiles
}

type ResourceJSONMdl struct {
	PK       uint        `gorm:"primary_key"`
	Tenant   string      `index:"0" re:".*"`
	ID       string      `index:"1" re:".*"`
	Resource utils.JSONB `gorm:"type:jsonb" index:"2" re:".*"`
}

func (ResourceJSONMdl) TableName() string {
	return utils.TBLResources
}

type StatQueueProfileMdl struct {
	PK               uint        `gorm:"primary_key"`
	Tenant           string      `index:"0" re:".*"`
	ID               string      `index:"1" re:".*"`
	StatQueueProfile utils.JSONB `gorm:"type:jsonb" index:"2" re:".*"`
}

func (StatQueueProfileMdl) TableName() string {
	return utils.TBLStatQueueProfiles
}

type StatQueueMdl struct {
	PK        uint        `gorm:"primary_key"`
	Tenant    string      `index:"0" re:".*"`
	ID        string      `index:"1" re:".*"`
	StatQueue utils.JSONB `gorm:"type:jsonb" index:"2" re:".*"`
}

func (StatQueueMdl) TableName() string {
	return utils.TBLStatQueues
}

type ThresholdProfileMdl struct {
	PK               uint        `gorm:"primary_key"`
	Tenant           string      `index:"0" re:".*"`
	ID               string      `index:"1" re:".*"`
	ThresholdProfile utils.JSONB `gorm:"type:jsonb" index:"2" re:".*"`
}

func (ThresholdProfileMdl) TableName() string {
	return utils.TBLThresholdProfiles
}

type ThresholdJSONMdl struct {
	PK        uint        `gorm:"primary_key"`
	Tenant    string      `index:"0" re:".*"`
	ID        string      `index:"1" re:".*"`
	Threshold utils.JSONB `gorm:"type:jsonb" index:"2" re:".*"`
}

func (ThresholdJSONMdl) TableName() string {
	return utils.TBLThresholds
}

type FilterJSONMdl struct {
	PK     uint        `gorm:"primary_key"`
	Tenant string      `index:"0" re:".*"`
	ID     string      `index:"1" re:".*"`
	Filter utils.JSONB `gorm:"type:jsonb" index:"2" re:".*"`
}

func (FilterJSONMdl) TableName() string {
	return utils.TBLFilters
}

type RouteProfileMdl struct {
	PK           uint        `gorm:"primary_key"`
	Tenant       string      `index:"0" re:".*"`
	ID           string      `index:"1" re:".*"`
	RouteProfile utils.JSONB `gorm:"type:jsonb" index:"2" re:".*"`
}

func (RouteProfileMdl) TableName() string {
	return utils.TBLRouteProfiles
}

// Doesnt include Rates in RateProfile json, Rates taken from RateMdl using foreign keys
type RateProfileJSONMdl struct {
	PK          uint        `gorm:"primary_key"`
	Tenant      string      `index:"0" re:".*"`
	ID          string      `index:"1" re:".*"`
	RateProfile utils.JSONB `gorm:"type:jsonb" index:"2" re:".*"`
}

func (RateProfileJSONMdl) TableName() string {
	return utils.TBLRateProfiles
}

type RateMdl struct {
	PK            uint        `gorm:"primary_key"`
	Tenant        string      `index:"0" re:".*"`
	ID            string      `index:"1" re:".*"`
	RateProfileID string      `gorm:"foreign_key" index:"3" re:".*"`
	Rate          utils.JSONB `gorm:"type:jsonb" index:"2" re:".*"`
}

func (RateMdl) TableName() string {
	return utils.TBLRates
}

type RankingProfileMdl struct {
	PK             uint        `gorm:"primary_key"`
	Tenant         string      `index:"0" re:".*"`
	ID             string      `index:"1" re:".*"`
	RankingProfile utils.JSONB `gorm:"type:jsonb" index:"2" re:".*"`
}

func (RankingProfileMdl) TableName() string {
	return utils.TBLRankingProfiles
}

type RankingJSONMdl struct {
	PK      uint        `gorm:"primary_key"`
	Tenant  string      `index:"0" re:".*"`
	ID      string      `index:"1" re:".*"`
	Ranking utils.JSONB `gorm:"type:jsonb" index:"2" re:".*"`
}

func (RankingJSONMdl) TableName() string {
	return utils.TBLRankings
}

type TrendProfileMdl struct {
	PK           uint        `gorm:"primary_key"`
	Tenant       string      `index:"0" re:".*"`
	ID           string      `index:"1" re:".*"`
	TrendProfile utils.JSONB `gorm:"type:jsonb" index:"2" re:".*"`
}

func (TrendProfileMdl) TableName() string {
	return utils.TBLTrendProfiles
}

type TrendJSONMdl struct {
	PK     uint        `gorm:"primary_key"`
	Tenant string      `index:"0" re:".*"`
	ID     string      `index:"1" re:".*"`
	Trend  utils.JSONB `gorm:"type:jsonb" index:"2" re:".*"`
}

func (TrendJSONMdl) TableName() string {
	return utils.TBLTrends
}

type LoadInstanceMdl struct {
	Key          string      `gorm:"primary_key"`
	LoadInstance utils.JSONB `gorm:"type:jsonb" index:"0" re:".*"`
}

func (LoadInstanceMdl) TableName() string {
	return utils.LoadInstKey
}

type LoadIDMdl struct {
	PK      uint        `gorm:"primary_key"`
	LoadIDs utils.JSONB `gorm:"type:jsonb" index:"0" re:".*"`
}

func (LoadIDMdl) TableName() string {
	return utils.TBLLoadIDs
}

type IndexMdl struct {
	PK     uint        `gorm:"primary_key"`
	Tenant string      `index:"0" re:".*"`
	Type   string      `index:"1" re:".*"`
	Key    string      `index:"2" re:".*"`
	Value  utils.JSONB `gorm:"type:jsonb" index:"3" re:".*"`
}

func (IndexMdl) TableName() string {
	return utils.TBLIndexes
}
