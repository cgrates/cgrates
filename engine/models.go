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

type TpTiming struct {
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

type TpDestination struct {
	Id        int64
	Tpid      string
	Tag       string `index:"0" re:"\w+\s*,\s*"`
	Prefix    string `index:"1" re:"\+?\d+.?\d*"`
	CreatedAt time.Time
}

type TpRate struct {
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

type TpDestinationRate struct {
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

type TpRatingPlan struct {
	Id           int64
	Tpid         string
	Tag          string  `index:"0" re:"\w+\s*,\s*"`
	DestratesTag string  `index:"1" re:"\w+\s*,\s*|\*any"`
	TimingTag    string  `index:"2" re:"\w+\s*,\s*|\*any"`
	Weight       float64 `index:"3" re:"\d+.?\d*"`
	CreatedAt    time.Time
}

type TpRatingProfile struct {
	Id               int64
	Tpid             string
	Loadid           string
	Direction        string `index:"0" re:"\*out\s*"`
	Tenant           string `index:"1" re:"[0-9A-Za-z_\.]+\s*"`
	Category         string `index:"2" re:"\w+\s*"`
	Subject          string `index:"3" re:"\*any\s*|(\w+;?)+\s*"`
	ActivationTime   string `index:"4" re:"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z"`
	RatingPlanTag    string `index:"5" re:"\w+\s*"`
	FallbackSubjects string `index:"6" re:"\w+\s*"`
	CdrStatQueueIds  string `index:"7" re:"\w+\s*"`
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

type TpAction struct {
	Id              int64
	Tpid            string
	Tag             string  `index:"0" re:"\w+\s*"`
	Action          string  `index:"1" re:"\*\w+\s*"`
	ExtraParameters string  `index:"2" re:"\S+\s*"`
	Filter          string  `index:"3" re:"\S+\s*"`
	BalanceTag      string  `index:"4" re:"\w+\s*"`
	BalanceType     string  `index:"5" re:"\*\w+\s*"`
	Directions      string  `index:"6" re:""`
	Categories      string  `index:"7" re:""`
	DestinationTags string  `index:"8" re:"\*any|\w+\s*"`
	RatingSubject   string  `index:"9" re:"\w+\s*"`
	SharedGroups    string  `index:"10" re:"[0-9A-Za-z_;]*"`
	ExpiryTime      string  `index:"11" re:"\*\w+\s*|\+\d+[smh]\s*|\d+\s*"`
	TimingTags      string  `index:"12" re:"[0-9A-Za-z_;]*|\*any"`
	Units           string  `index:"13" re:"\d+\s*"`
	BalanceWeight   string  `index:"14" re:"\d+\.?\d*\s*"`
	BalanceBlocker  string  `index:"15" re:""`
	BalanceDisabled string  `index:"16" re:""`
	Weight          float64 `index:"17" re:"\d+\.?\d*\s*"`
	CreatedAt       time.Time
}

type TpActionPlan struct {
	Id         int64
	Tpid       string
	Tag        string  `index:"0" re:"\w+\s*,\s*"`
	ActionsTag string  `index:"1" re:"\w+\s*,\s*"`
	TimingTag  string  `index:"2" re:"\w+\s*,\s*"|\*any`
	Weight     float64 `index:"3" re:"\d+\.?\d*"`
	CreatedAt  time.Time
}

type TpActionTrigger struct {
	Id                     int64
	Tpid                   string
	Tag                    string  `index:"0" re:"\w+"`
	UniqueId               string  `index:"1" re:"\w+"`
	ThresholdType          string  `index:"2" re:"\*\w+"`
	ThresholdValue         float64 `index:"3" re:"\d+\.?\d*"`
	Recurrent              bool    `index:"4" re:"true|false"`
	MinSleep               string  `index:"5" re:"\d+[smh]?"`
	ExpiryTime             string  `index:"6" re:""`
	ActivationTime         string  `index:"7" re:""`
	BalanceTag             string  `index:"8" re:"\w+\s*"`
	BalanceType            string  `index:"9" re:"\*\w+"`
	BalanceDirections      string  `index:"10" re:"\*out"`
	BalanceCategories      string  `index:"11" re:""`
	BalanceDestinationTags string  `index:"12" re:"\w+|\*any"`
	BalanceRatingSubject   string  `index:"13" re:"\w+|\*any"`
	BalanceSharedGroups    string  `index:"14" re:"\w+|\*any"`
	BalanceExpiryTime      string  `index:"15" re:"\*\w+\s*|\+\d+[smh]\s*|\d+\s*"`
	BalanceTimingTags      string  `index:"16" re:"[0-9A-Za-z_;]*|\*any"`
	BalanceWeight          string  `index:"17" re:"\d+\.?\d*"`
	BalanceBlocker         string  `index:"18" re:""`
	BalanceDisabled        string  `index:"19" re:""`
	MinQueuedItems         int     `index:"20" re:"\d+"`
	ActionsTag             string  `index:"21" re:"\w+"`
	Weight                 float64 `index:"22" re:"\d+\.?\d*"`
	CreatedAt              time.Time
}

type TpAccountAction struct {
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
	return utils.AccountKey(aa.Tenant, aa.Account)
}

type TpSharedGroup struct {
	Id            int64
	Tpid          string
	Tag           string `index:"0" re:"\w+\s*"`
	Account       string `index:"1" re:"\*?\w+\s*"`
	Strategy      string `index:"2" re:"\*\w+\s*"`
	RatingSubject string `index:"3" re:"\*?\w]+\s*"`
	CreatedAt     time.Time
}

type TpDerivedCharger struct {
	Id                   int64
	Tpid                 string
	Loadid               string
	Direction            string `index:"0" re:"\*out"`
	Tenant               string `index:"1" re:"[0-9A-Za-z_\.]+\s*"`
	Category             string `index:"2" re:"\w+\s*"`
	Account              string `index:"3" re:"\w+\s*"`
	Subject              string `index:"4" re:"\*any\s*|\w+\s*"`
	DestinationIds       string `index:"5" re:""`
	Runid                string `index:"6" re:"\w+\s*"`
	RunFilters           string `index:"7" re:"[~^]*[0-9A-Za-z_/:().+]+\s*"`
	ReqTypeField         string `index:"8" re:"\*default\s*|[~^*]*[0-9A-Za-z_/:().+]+\s*"`
	DirectionField       string `index:"9" re:"\*default\s*|[~^]*[0-9A-Za-z_/:().+]+\s*"`
	TenantField          string `index:"10" re:"\*default\s*|[~^]*[0-9A-Za-z_/:().+]+\s*"`
	CategoryField        string `index:"11" re:"\*default\s*|[~^]*[0-9A-Za-z_/:().+]+\s*"`
	AccountField         string `index:"12" re:"\*default\s*|[~^]*[0-9A-Za-z_/:().+]+\s*"`
	SubjectField         string `index:"13" re:"\*default\s*|[~^]*[0-9A-Za-z_/:().+]+\s*"`
	DestinationField     string `index:"14" re:"\*default\s*|[~^]*[0-9A-Za-z_/:().+]+\s*"`
	SetupTimeField       string `index:"15" re:"\*default\s*|[~^]*[0-9A-Za-z_/:().+]+\s*"`
	PddField             string `index:"16" re:"\*default\s*|[~^]*[0-9A-Za-z_/:().+]+\s*"`
	AnswerTimeField      string `index:"17" re:"\*default\s*|[~^]*[0-9A-Za-z_/:().+]+\s*"`
	UsageField           string `index:"18" re:"\*default\s*|[~^]*[0-9A-Za-z_/:().+]+\s*"`
	SupplierField        string `index:"19" re:"\*default\s*|[~^]*[0-9A-Za-z_/:().+]+\s*"`
	DisconnectCauseField string `index:"20" re:"\*default\s*|[~^]*[0-9A-Za-z_/:().+]+\s*"`
	RatedField           string `index:"21" re:"\*default\s*|[~^]*[0-9A-Za-z_/:().+]+\s*"`
	CostField            string `index:"22" re:"\*default\s*|[~^]*[0-9A-Za-z_/:().+]+\s*"`
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

type TpUser struct {
	Id             int64
	Tpid           string
	Tenant         string  `index:"0" re:""`
	UserName       string  `index:"1" re:""`
	Masked         bool    `index:"2" re:""`
	AttributeName  string  `index:"3" re:""`
	AttributeValue string  `index:"4" re:""`
	Weight         float64 `index:"5" re:""`
	CreatedAt      time.Time
}

func (tu *TpUser) GetId() string {
	return utils.ConcatenatedKey(tu.Tenant, tu.UserName)
}

func (tu *TpUser) SetId(id string) error {
	vals := strings.Split(id, utils.CONCATENATED_KEY_SEP)
	if len(vals) != 2 {
		return utils.ErrInvalidKey
	}
	tu.Tenant = vals[0]
	tu.UserName = vals[1]
	return nil
}

type TpAlias struct {
	Id            int64
	Tpid          string
	Direction     string  `index:"0" re:""`
	Tenant        string  `index:"1" re:""`
	Category      string  `index:"2" re:""`
	Account       string  `index:"3" re:""`
	Subject       string  `index:"4" re:""`
	DestinationId string  `index:"5" re:""`
	Context       string  `index:"6" re:""`
	Target        string  `index:"7" re:""`
	Original      string  `index:"8" re:""`
	Alias         string  `index:"9" re:""`
	Weight        float64 `index:"10" re:""`
}

func (ta *TpAlias) TableName() string {
	return utils.TBLTPAliases
}

func (ta *TpAlias) SetId(id string) error {
	vals := strings.Split(id, utils.CONCATENATED_KEY_SEP)
	if len(vals) != 6 {
		return utils.ErrInvalidKey
	}
	ta.Direction = vals[0]
	ta.Tenant = vals[1]
	ta.Category = vals[2]
	ta.Account = vals[3]
	ta.Subject = vals[4]
	ta.Context = vals[5]
	return nil
}

func (ta *TpAlias) GetId() string {
	return utils.ConcatenatedKey(ta.Direction, ta.Tenant, ta.Category, ta.Account, ta.Subject, ta.Context)
}

type TpResource struct {
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

type TpStats struct {
	PK                 uint `gorm:"primary_key"`
	Tpid               string
	Tenant             string  `index:"0" re:""`
	ID                 string  `index:"1" re:""`
	FilterIDs          string  `index:"2" re:""`
	ActivationInterval string  `index:"3" re:""`
	QueueLength        int     `index:"4" re:""`
	TTL                string  `index:"5" re:""`
	Metrics            string  `index:"6" re:""`
	Parameters         string  `index:"7" re:""`
	Blocker            bool    `index:"8" re:""`
	Stored             bool    `index:"9" re:""`
	Weight             float64 `index:"10" re:"\d+\.?\d*"`
	MinItems           int     `index:"11" re:""`
	ThresholdIDs       string  `index:"12" re:""`
	CreatedAt          time.Time
}

type TpThreshold struct {
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

type TpFilter struct {
	PK                 uint `gorm:"primary_key"`
	Tpid               string
	Tenant             string `index:"0" re:""`
	ID                 string `index:"1" re:""`
	FilterType         string `index:"2" re:"^\*[A-Za-z].*"`
	FilterFieldName    string `index:"3" re:""`
	FilterFieldValues  string `index:"4" re:""`
	ActivationInterval string `index:"5" re:""`
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

type SessionsCostsSQL struct {
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

func (t SessionsCostsSQL) TableName() string {
	return utils.SessionsCostsTBL
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
	Tenant                string  `index:"0" re:""`
	ID                    string  `index:"1" re:""`
	FilterIDs             string  `index:"2" re:""`
	ActivationInterval    string  `index:"3" re:""`
	Sorting               string  `index:"4" re:""`
	SortingParameters     string  `index:"5" re:""`
	SupplierID            string  `index:"6" re:""`
	SupplierFilterIDs     string  `index:"7" re:""`
	SupplierAccountIDs    string  `index:"8" re:""`
	SupplierRatingplanIDs string  `index:"9" re:""`
	SupplierResourceIDs   string  `index:"10" re:""`
	SupplierStatIDs       string  `index:"11" re:""`
	SupplierWeight        float64 `index:"12" re:"\d+\.?\d*"`
	SupplierBlocker       bool    `index:"13" re:""`
	SupplierParameters    string  `index:"14" re:""`
	Weight                float64 `index:"15" re:"\d+\.?\d*"`
	CreatedAt             time.Time
}

type TPAttribute struct {
	PK                 uint `gorm:"primary_key"`
	Tpid               string
	Tenant             string  `index:"0" re:""`
	ID                 string  `index:"1" re:""`
	Contexts           string  `index:"2" re:""`
	FilterIDs          string  `index:"3" re:""`
	ActivationInterval string  `index:"4" re:""`
	FieldName          string  `index:"5" re:""`
	Initial            string  `index:"6" re:""`
	Substitute         string  `index:"7" re:""`
	Append             bool    `index:"8" re:""`
	Blocker            bool    `index:"9" re:""`
	Weight             float64 `index:"10" re:"\d+\.?\d*"`
	CreatedAt          time.Time
}

type TPCharger struct {
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
