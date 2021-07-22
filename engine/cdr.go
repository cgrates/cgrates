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
	"encoding/json"
	"math"
	"strconv"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func NewCDRFromExternalCDR(extCdr *ExternalCDR, timezone string) (*CDR, error) {
	var err error
	cdr := &CDR{CGRID: extCdr.CGRID, RunID: extCdr.RunID, OrderID: extCdr.OrderID, ToR: extCdr.ToR,
		OriginID: extCdr.OriginID, OriginHost: extCdr.OriginHost,
		Source: extCdr.Source, RequestType: extCdr.RequestType, Tenant: extCdr.Tenant, Category: extCdr.Category,
		Account: extCdr.Account, Subject: extCdr.Subject, Destination: extCdr.Destination,
		CostSource: extCdr.CostSource, Cost: extCdr.Cost, PreRated: extCdr.PreRated}
	if extCdr.SetupTime != "" {
		if cdr.SetupTime, err = utils.ParseTimeDetectLayout(extCdr.SetupTime, timezone); err != nil {
			return nil, err
		}
	}
	if len(cdr.CGRID) == 0 { // Populate CGRID if not present
		cdr.ComputeCGRID()
	}
	if extCdr.AnswerTime != "" {
		if cdr.AnswerTime, err = utils.ParseTimeDetectLayout(extCdr.AnswerTime, timezone); err != nil {
			return nil, err
		}
	}
	if extCdr.Usage != "" {
		if cdr.Usage, err = utils.ParseDurationWithNanosecs(extCdr.Usage); err != nil {
			return nil, err
		}
	}
	if extCdr.ExtraFields != nil {
		cdr.ExtraFields = make(map[string]string)
	}
	for k, v := range extCdr.ExtraFields {
		cdr.ExtraFields[k] = v
	}
	return cdr, nil
}

type CDR struct {
	CGRID       string
	RunID       string
	OrderID     int64             // Stor order id used as export order id
	OriginHost  string            // represents the IP address of the host generating the CDR (automatically populated by the server)
	Source      string            // formally identifies the source of the CDR (free form field)
	OriginID    string            // represents the unique accounting id given by the telecom switch generating the CDR
	ToR         string            // type of record, meta-field, should map to one of the TORs hardcoded inside the server <*voice|*data|*sms|*generic>
	RequestType string            // matching the supported request types by the **CGRateS**, accepted values are hardcoded in the server <prepaid|postpaid|pseudoprepaid|rated>.
	Tenant      string            // tenant whom this record belongs
	Category    string            // free-form filter for this record, matching the category defined in rating profiles.
	Account     string            // account id (accounting subsystem) the record should be attached to
	Subject     string            // rating subject (rating subsystem) this record should be attached to
	Destination string            // destination to be charged
	SetupTime   time.Time         // set-up time of the event. Supported formats: datetime RFC3339 compatible, SQL datetime (eg: MySQL), unix timestamp.
	AnswerTime  time.Time         // answer time of the event. Supported formats: datetime RFC3339 compatible, SQL datetime (eg: MySQL), unix timestamp.
	Usage       time.Duration     // event usage information (eg: in case of tor=*voice this will represent the total duration of a call)
	ExtraFields map[string]string // Extra fields to be stored in CDR
	ExtraInfo   string            // Container for extra information related to this CDR, eg: populated with error reason in case of error on calculation
	Partial     bool              // Used for partial record processing by ERs
	PreRated    bool              // Mark the CDR as rated so we do not process it during rating
	CostSource  string            // The source of this cost
	Cost        float64           //
}

// AddDefaults will add missing information based on other fields
func (cdr *CDR) AddDefaults(cfg *config.CGRConfig) {
	if cdr.CGRID == utils.EmptyString {
		cdr.ComputeCGRID()
	}
	if cdr.RunID == utils.EmptyString {
		cdr.RunID = utils.MetaDefault
	}
	if cdr.ToR == utils.EmptyString {
		cdr.ToR = utils.MetaVoice
	}
	if cdr.RequestType == utils.EmptyString {
		cdr.RequestType = cfg.GeneralCfg().DefaultReqType
	}
	if cdr.Tenant == utils.EmptyString {
		cdr.Tenant = cfg.GeneralCfg().DefaultTenant
	}
	if cdr.Category == utils.EmptyString {
		cdr.Category = cfg.GeneralCfg().DefaultCategory
	}
	if cdr.Subject == utils.EmptyString {
		cdr.Subject = cdr.Account
	}
}

func (cdr *CDR) ComputeCGRID() {
	cdr.CGRID = utils.Sha1(cdr.OriginID, cdr.OriginHost)
}

// FormatCost formats the cost as string on export
func (cdr *CDR) FormatCost(shiftDecimals, roundDecimals int) string {
	cost := cdr.Cost
	if shiftDecimals != 0 {
		cost = cost * math.Pow10(shiftDecimals)
	}
	return strconv.FormatFloat(cost, 'f', roundDecimals, 64)
}

// FieldAsString is used to retrieve fields as string, primary fields are const labeled
func (cdr *CDR) FieldAsString(rsrPrs *config.RSRParser) (parsed string, err error) {
	parsed, err = rsrPrs.ParseDataProviderWithInterfaces(
		cdr.AsMapStorage())
	if err != nil {
		return
	}
	return
}

// FieldsAsString concatenates values of multiple fields defined in template, used eg in CDR templates
func (cdr *CDR) FieldsAsString(rsrFlds config.RSRParsers) string {
	outVal, err := rsrFlds.ParseDataProvider(
		cdr.AsMapStorage())
	if err != nil {
		return ""
	}
	return outVal
}

func (cdr *CDR) Clone() *CDR {
	if cdr == nil {
		return nil
	}
	cln := &CDR{
		CGRID:       cdr.CGRID,
		RunID:       cdr.RunID,
		OrderID:     cdr.OrderID,
		OriginHost:  cdr.OriginHost,
		Source:      cdr.Source,
		OriginID:    cdr.OriginID,
		ToR:         cdr.ToR,
		RequestType: cdr.RequestType,
		Tenant:      cdr.Tenant,
		Category:    cdr.Category,
		Account:     cdr.Account,
		Subject:     cdr.Subject,
		Destination: cdr.Destination,
		SetupTime:   cdr.SetupTime,
		AnswerTime:  cdr.AnswerTime,
		Usage:       cdr.Usage,
		ExtraFields: cdr.ExtraFields,
		ExtraInfo:   cdr.ExtraInfo,
		Partial:     cdr.Partial,
		PreRated:    cdr.PreRated,
		CostSource:  cdr.CostSource,
		Cost:        cdr.Cost,
	}
	if cdr.ExtraFields != nil {
		cln.ExtraFields = make(map[string]string, len(cdr.ExtraFields))
		for key, val := range cdr.ExtraFields {
			cln.ExtraFields[key] = val
		}
	}

	return cln
}

func (cdr *CDR) AsMapStorage() (mp utils.MapStorage) {
	mp = utils.MapStorage{
		utils.MetaReq: cdr.AsMapStringIface(),
	}
	return
}

func (cdr *CDR) AsMapStringIface() (mp map[string]interface{}) {
	mp = make(map[string]interface{})
	for fld, val := range cdr.ExtraFields {
		mp[fld] = val
	}
	mp[utils.CGRID] = cdr.CGRID
	mp[utils.RunID] = cdr.RunID
	mp[utils.OrderID] = cdr.OrderID
	mp[utils.OriginHost] = cdr.OriginHost
	mp[utils.Source] = cdr.Source
	mp[utils.OriginID] = cdr.OriginID
	mp[utils.ToR] = cdr.ToR
	mp[utils.RequestType] = cdr.RequestType
	mp[utils.Tenant] = cdr.Tenant
	mp[utils.Category] = cdr.Category
	mp[utils.AccountField] = cdr.Account
	mp[utils.Subject] = cdr.Subject
	mp[utils.Destination] = cdr.Destination
	mp[utils.SetupTime] = cdr.SetupTime
	mp[utils.AnswerTime] = cdr.AnswerTime
	mp[utils.Usage] = cdr.Usage
	mp[utils.ExtraInfo] = cdr.ExtraInfo
	mp[utils.Partial] = cdr.Partial
	mp[utils.PreRated] = cdr.PreRated
	mp[utils.CostSource] = cdr.CostSource
	mp[utils.Cost] = cdr.Cost
	return
}

func (cdr *CDR) AsExternalCDR() *ExternalCDR {
	var usageStr string
	switch cdr.ToR {
	case utils.MetaVoice: // usage as time
		usageStr = cdr.Usage.String()
	default: // usage as units
		usageStr = strconv.FormatInt(cdr.Usage.Nanoseconds(), 10)
	}
	return &ExternalCDR{CGRID: cdr.CGRID,
		RunID:       cdr.RunID,
		OrderID:     cdr.OrderID,
		OriginHost:  cdr.OriginHost,
		Source:      cdr.Source,
		OriginID:    cdr.OriginID,
		ToR:         cdr.ToR,
		RequestType: cdr.RequestType,
		Tenant:      cdr.Tenant,
		Category:    cdr.Category,
		Account:     cdr.Account,
		Subject:     cdr.Subject,
		Destination: cdr.Destination,
		SetupTime:   cdr.SetupTime.Format(time.RFC3339),
		AnswerTime:  cdr.AnswerTime.Format(time.RFC3339),
		Usage:       usageStr,
		ExtraFields: cdr.ExtraFields,
		CostSource:  cdr.CostSource,
		Cost:        cdr.Cost,
		ExtraInfo:   cdr.ExtraInfo,
		PreRated:    cdr.PreRated,
	}
}

func (cdr *CDR) String() string {
	mrsh, _ := json.Marshal(cdr)
	return string(mrsh)
}

// AsCDRsql converts the CDR into the format used for SQL storage
func (cdr *CDR) AsCDRsql() (cdrSQL *CDRsql) {
	cdrSQL = new(CDRsql)
	cdrSQL.Cgrid = cdr.CGRID
	cdrSQL.RunID = cdr.RunID
	cdrSQL.OriginHost = cdr.OriginHost
	cdrSQL.Source = cdr.Source
	cdrSQL.OriginID = cdr.OriginID
	cdrSQL.TOR = cdr.ToR
	cdrSQL.RequestType = cdr.RequestType
	cdrSQL.Tenant = cdr.Tenant
	cdrSQL.Category = cdr.Category
	cdrSQL.Account = cdr.Account
	cdrSQL.Subject = cdr.Subject
	cdrSQL.Destination = cdr.Destination
	cdrSQL.SetupTime = cdr.SetupTime
	cdrSQL.AnswerTime = cdr.AnswerTime
	cdrSQL.Usage = cdr.Usage.Nanoseconds()
	cdrSQL.ExtraFields = utils.ToJSON(cdr.ExtraFields)
	cdrSQL.CostSource = cdr.CostSource
	cdrSQL.Cost = cdr.Cost
	cdrSQL.ExtraInfo = cdr.ExtraInfo
	cdrSQL.CreatedAt = time.Now()
	return
}

func (cdr *CDR) AsCGREvent() *utils.CGREvent {
	return &utils.CGREvent{
		Tenant:  cdr.Tenant,
		ID:      utils.UUIDSha1Prefix(),
		Event:   cdr.AsMapStringIface(),
		APIOpts: map[string]interface{}{},
	}
}

// NewCDRFromSQL converts the CDRsql into CDR
func NewCDRFromSQL(cdrSQL *CDRsql) (cdr *CDR, err error) {
	cdr = new(CDR)
	cdr.CGRID = cdrSQL.Cgrid
	cdr.RunID = cdrSQL.RunID
	cdr.OriginHost = cdrSQL.OriginHost
	cdr.Source = cdrSQL.Source
	cdr.OriginID = cdrSQL.OriginID
	cdr.OrderID = cdrSQL.ID
	cdr.ToR = cdrSQL.TOR
	cdr.RequestType = cdrSQL.RequestType
	cdr.Tenant = cdrSQL.Tenant
	cdr.Category = cdrSQL.Category
	cdr.Account = cdrSQL.Account
	cdr.Subject = cdrSQL.Subject
	cdr.Destination = cdrSQL.Destination
	cdr.SetupTime = cdrSQL.SetupTime
	cdr.AnswerTime = cdrSQL.AnswerTime
	cdr.Usage = time.Duration(cdrSQL.Usage)
	cdr.CostSource = cdrSQL.CostSource
	cdr.Cost = cdrSQL.Cost
	cdr.ExtraInfo = cdrSQL.ExtraInfo
	if cdrSQL.ExtraFields != "" {
		if err = json.Unmarshal([]byte(cdrSQL.ExtraFields), &cdr.ExtraFields); err != nil {
			return nil, err
		}
	}
	return
}

type ExternalCDR struct {
	CGRID       string
	RunID       string
	OrderID     int64
	OriginHost  string
	Source      string
	OriginID    string
	ToR         string
	RequestType string
	Tenant      string
	Category    string
	Account     string
	Subject     string
	Destination string
	SetupTime   string
	AnswerTime  string
	Usage       string
	ExtraFields map[string]string
	CostSource  string
	Cost        float64
	CostDetails string
	ExtraInfo   string
	PreRated    bool // Mark the CDR as rated so we do not process it during mediation
}

// UsageRecord is used when authorizing requests from outside, eg APIerSv1.GetMaxUsage
type UsageRecord struct {
	ToR         string
	RequestType string
	Tenant      string
	Category    string
	Account     string
	Subject     string
	Destination string
	SetupTime   string
	AnswerTime  string
	Usage       string
	ExtraFields map[string]string
}

func (uR *UsageRecord) GetID() string {
	return utils.Sha1(uR.ToR, uR.RequestType, uR.Tenant, uR.Category, uR.Account, uR.Subject, uR.Destination, uR.SetupTime, uR.AnswerTime, uR.Usage)
}

type ExternalCDRWithAPIOpts struct {
	*ExternalCDR
	APIOpts map[string]interface{}
}

type UsageRecordWithAPIOpts struct {
	*UsageRecord
	APIOpts map[string]interface{}
}

type CDRWithAPIOpts struct {
	*CDR
	APIOpts map[string]interface{}
}
