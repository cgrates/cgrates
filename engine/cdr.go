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
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"io"
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
	if len(extCdr.CostDetails) != 0 {
		cdr.CostDetails = &EventCost{}
		if err = json.Unmarshal([]byte(extCdr.CostDetails), cdr.CostDetails); err != nil {
			return nil, err
		}
		cdr.CostDetails.initCache()
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
	CostDetails *EventCost        // Attach the cost details to CDR when possible
}

type CompressedCostCDR struct {
	CGRID                 string
	RunID                 string
	OriginID              string
	OrderID               int64
	OriginHost            string
	Source                string
	ToR                   string
	RequestType           string
	Tenant                string
	Category              string
	Account               string
	Subject               string
	Destination           string
	SetupTime             time.Time
	AnswerTime            time.Time
	Usage                 time.Duration
	ExtraFields           map[string]string
	ExtraInfo             string
	Partial               bool
	PreRated              bool
	CostSource            string
	Cost                  float64
	CompressedCostDetails []byte
}

func (cdr *CDR) AsCompressedCostCDR(ms Marshaler) (*CompressedCostCDR, error) {
	cdrCompressed := &CompressedCostCDR{
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
	result, err := ms.Marshal(cdr.CostDetails)
	if err != nil {
		return nil, err
	}
	var compressedBuf bytes.Buffer
	gzWriter := gzip.NewWriter(&compressedBuf)

	if _, err := gzWriter.Write(result); err != nil {
		return nil, err
	}
	if err := gzWriter.Close(); err != nil {
		return nil, err
	}
	cdrCompressed.CompressedCostDetails = compressedBuf.Bytes()
	return cdrCompressed, nil
}

func NewCDRFromCompressedCDR(cdrCompressed *CompressedCostCDR, ms Marshaler) (*CDR, error) {
	cdr := &CDR{
		CGRID:       cdrCompressed.CGRID,
		RunID:       cdrCompressed.RunID,
		OrderID:     cdrCompressed.OrderID,
		OriginHost:  cdrCompressed.OriginHost,
		Source:      cdrCompressed.Source,
		OriginID:    cdrCompressed.OriginID,
		ToR:         cdrCompressed.ToR,
		RequestType: cdrCompressed.RequestType,
		Tenant:      cdrCompressed.Tenant,
		Category:    cdrCompressed.Category,
		Account:     cdrCompressed.Account,
		Subject:     cdrCompressed.Subject,
		Destination: cdrCompressed.Destination,
		SetupTime:   cdrCompressed.SetupTime,
		AnswerTime:  cdrCompressed.AnswerTime,
		Usage:       cdrCompressed.Usage,
		ExtraFields: cdrCompressed.ExtraFields,
		ExtraInfo:   cdrCompressed.ExtraInfo,
		Partial:     cdrCompressed.Partial,
		PreRated:    cdrCompressed.PreRated,
		CostSource:  cdrCompressed.CostSource,
		Cost:        cdrCompressed.Cost,
	}
	if cdrCompressed.CompressedCostDetails == nil {
		return cdr, nil
	}
	compressedReader := bytes.NewReader(cdrCompressed.CompressedCostDetails)
	gzReader, err := gzip.NewReader(compressedReader)
	if err != nil {
		return nil, err
	}
	defer gzReader.Close()
	value, err := io.ReadAll(gzReader)
	if err != nil {
		return nil, err
	}
	if err = ms.Unmarshal(value, &cdr.CostDetails); err != nil {
		return nil, err
	}
	return cdr, nil
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

func (cdr *CDR) CostDetailsJson() string {
	mrshled, _ := json.Marshal(cdr.CostDetails)
	return string(mrshled)
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
		CostDetails: cdr.CostDetails.Clone(),
	}
	if cdr.ExtraFields != nil {
		cln.ExtraFields = make(map[string]string, len(cdr.ExtraFields))
		for key, val := range cdr.ExtraFields {
			cln.ExtraFields[key] = val
		}
	}

	return cln
}

// CacheClone returns a clone of CDR used by ltcache CacheCloner
func (cdr *CDR) CacheClone() any {
	return cdr.Clone()
}

func (cdr *CDR) AsMapStorage() (mp utils.MapStorage) {
	mp = utils.MapStorage{
		utils.MetaReq: cdr.AsMapStringIface(),
	}
	if cdr.CostDetails != nil {
		mp[utils.MetaEC] = cdr.CostDetails
	}
	return
}

func (cdr *CDR) AsMapStringIface() (mp map[string]any) {
	mp = make(map[string]any)
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
	if cdr.CostDetails != nil {
		mp[utils.CostDetails] = cdr.CostDetails
	}
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
		CostDetails: cdr.CostDetailsJson(),
		ExtraInfo:   cdr.ExtraInfo,
		PreRated:    cdr.PreRated,
	}
}

func (cdr *CDR) String() string {
	mrsh, _ := json.Marshal(cdr)
	return string(mrsh)
}

// AsCDRsql converts the CDR into the format used for SQL storage
func (cdr *CDR) AsCDRsql(ms Marshaler) (cdrSQL *CDRsql, err error) {
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
	if !cdr.AnswerTime.IsZero() {
		cdrSQL.AnswerTime = utils.TimePointer(cdr.AnswerTime)
	}
	cdrSQL.Usage = cdr.Usage.Nanoseconds()
	cdrSQL.ExtraFields = utils.ToJSON(cdr.ExtraFields)
	cdrSQL.CostSource = cdr.CostSource
	cdrSQL.Cost = cdr.Cost
	if config.CgrConfig().CdrsCfg().CompressStoredCost && cdr.CostDetails != nil {
		result, err := ms.Marshal(cdr.CostDetails)
		if err != nil {
			return nil, err
		}
		var compressBuffer bytes.Buffer
		gzWriter := gzip.NewWriter(&compressBuffer)
		if _, err := gzWriter.Write(result); err != nil {
			return nil, err
		}
		if err := gzWriter.Close(); err != nil {
			return nil, err
		}
		cdrSQL.CostDetails = base64.StdEncoding.EncodeToString(compressBuffer.Bytes())
	} else {
		cdrSQL.CostDetails = utils.ToJSON(cdr.CostDetails)
	}
	cdrSQL.ExtraInfo = cdr.ExtraInfo
	cdrSQL.CreatedAt = time.Now()
	return
}

func (cdr *CDR) AsCGREvent() *utils.CGREvent {
	return &utils.CGREvent{
		Tenant:  cdr.Tenant,
		ID:      utils.UUIDSha1Prefix(),
		Event:   cdr.AsMapStringIface(),
		APIOpts: map[string]any{},
	}
}

// NewCDRFromSQL converts the CDRsql into CDR
func NewCDRFromSQL(cdrSQL *CDRsql, ms Marshaler) (cdr *CDR, err error) {
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
	if cdrSQL.AnswerTime != nil {
		cdr.AnswerTime = *cdrSQL.AnswerTime
	}
	cdr.Usage = time.Duration(cdrSQL.Usage)
	cdr.CostSource = cdrSQL.CostSource
	cdr.Cost = cdrSQL.Cost
	cdr.ExtraInfo = cdrSQL.ExtraInfo
	if cdrSQL.ExtraFields != "" {
		if err = json.Unmarshal([]byte(cdrSQL.ExtraFields), &cdr.ExtraFields); err != nil {
			return nil, err
		}
	}
	if cdrSQL.CostDetails != utils.EmptyString {
		if config.CgrConfig().CdrsCfg().CompressStoredCost {
			decoded, err := base64.StdEncoding.DecodeString(cdrSQL.CostDetails)
			if err != nil {
				return nil, err
			}
			gzReader, err := gzip.NewReader(bytes.NewReader(decoded))
			if err != nil {
				return nil, err
			}
			unCompressCost, err := io.ReadAll(gzReader)
			if err := gzReader.Close(); err != nil {
				return nil, err
			}
			if err != nil {
				return nil, err
			}
			if err = ms.Unmarshal(unCompressCost, &cdr.CostDetails); err != nil {
				return nil, err
			}
		} else {
			if err = json.Unmarshal([]byte(cdrSQL.CostDetails), &cdr.CostDetails); err != nil {
				return nil, err
			}
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

func (uR *UsageRecord) AsCallDescriptor(timezone string, denyNegative bool) (*CallDescriptor, error) {
	var err error
	cd := &CallDescriptor{
		CgrID:               uR.GetId(),
		ToR:                 uR.ToR,
		Tenant:              uR.Tenant,
		Category:            uR.Category,
		Subject:             uR.Subject,
		Account:             uR.Account,
		Destination:         uR.Destination,
		DenyNegativeAccount: denyNegative,
	}
	timeStr := uR.AnswerTime
	if len(timeStr) == 0 { // In case of auth, answer time will not be defined, so take it out of setup one
		timeStr = uR.SetupTime
	}
	if cd.TimeStart, err = utils.ParseTimeDetectLayout(timeStr, timezone); err != nil {
		return nil, err
	}
	if usage, err := utils.ParseDurationWithNanosecs(uR.Usage); err != nil {
		return nil, err
	} else {
		cd.TimeEnd = cd.TimeStart.Add(usage)
	}
	if uR.ExtraFields != nil {
		cd.ExtraFields = make(map[string]string)
	}
	for k, v := range uR.ExtraFields {
		cd.ExtraFields[k] = v
	}
	return cd, nil
}

func (uR *UsageRecord) GetId() string {
	return utils.Sha1(uR.ToR, uR.RequestType, uR.Tenant, uR.Category, uR.Account, uR.Subject, uR.Destination, uR.SetupTime, uR.AnswerTime, uR.Usage)
}

type ExternalCDRWithAPIOpts struct {
	*ExternalCDR
	APIOpts map[string]any
}

type UsageRecordWithAPIOpts struct {
	*UsageRecord
	APIOpts map[string]any
}

type CDRWithAPIOpts struct {
	*CDR
	APIOpts map[string]any
}
