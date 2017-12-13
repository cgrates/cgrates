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
	"fmt"
	"math"
	"strconv"
	"strings"
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
		CostSource: extCdr.CostSource, Cost: extCdr.Cost, Rated: extCdr.Rated}
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
		if err = json.Unmarshal([]byte(extCdr.CostDetails), cdr.CostDetails); err != nil {
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

func NewCDRWithDefaults(cfg *config.CGRConfig) *CDR {
	return &CDR{ToR: utils.VOICE, RequestType: cfg.DefaultReqType,
		Tenant: cfg.DefaultTenant, Category: cfg.DefaultCategory,
		ExtraFields: make(map[string]string), Cost: -1}
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
	Partial     bool              // Used for partial record processing by CDRC
	Rated       bool              // Mark the CDR as rated so we do not process it during rating
	CostSource  string            // The source of this cost
	Cost        float64
	CostDetails *CallCost // Attach the cost details to CDR when possible
}

func (cdr *CDR) CostDetailsJson() string {
	mrshled, _ := json.Marshal(cdr.CostDetails)
	return string(mrshled)
}

func (cdr *CDR) ComputeCGRID() {
	cdr.CGRID = utils.Sha1(cdr.OriginID, cdr.SetupTime.UTC().String())
}

// Used to multiply usage on export
func (cdr *CDR) UsageMultiply(multiplyFactor float64, roundDecimals int) {
	cdr.Usage = time.Duration(int(utils.Round(float64(cdr.Usage.Nanoseconds())*multiplyFactor, roundDecimals, utils.ROUNDING_MIDDLE))) // Rounding down could introduce a slight loss here but only at nanoseconds level
}

// Used to multiply cost on export
func (cdr *CDR) CostMultiply(multiplyFactor float64, roundDecimals int) {
	cdr.Cost = utils.Round(cdr.Cost*multiplyFactor, roundDecimals, utils.ROUNDING_MIDDLE)
}

// Format cost as string on export
func (cdr *CDR) FormatCost(shiftDecimals, roundDecimals int) string {
	cost := cdr.Cost
	if shiftDecimals != 0 {
		cost = cost * math.Pow10(shiftDecimals)
	}
	return strconv.FormatFloat(cost, 'f', roundDecimals, 64)
}

// Formats usage on export
func (cdr *CDR) FormatUsage(layout string) string {
	if utils.IsSliceMember([]string{utils.DATA, utils.SMS, utils.MMS, utils.GENERIC}, cdr.ToR) {
		return strconv.FormatFloat(utils.Round(cdr.Usage.Seconds(), 0, utils.ROUNDING_MIDDLE), 'f', -1, 64)
	}
	switch layout {
	default:
		return strconv.FormatFloat(float64(cdr.Usage.Nanoseconds())/1000000000, 'f', -1, 64)
	}
}

// Used to retrieve fields as string, primary fields are const labeled
func (cdr *CDR) FieldAsString(rsrFld *utils.RSRField) string {
	if rsrFld.IsStatic() { // Static values do not care about headers
		return rsrFld.ParseValue("")
	}
	switch rsrFld.Id {
	case utils.CGRID:
		return rsrFld.ParseValue(cdr.CGRID)
	case utils.ORDERID:
		return rsrFld.ParseValue(strconv.FormatInt(cdr.OrderID, 10))
	case utils.TOR:
		return rsrFld.ParseValue(cdr.ToR)
	case utils.ACCID:
		return rsrFld.ParseValue(cdr.OriginID)
	case utils.CDRHOST:
		return rsrFld.ParseValue(cdr.OriginHost)
	case utils.CDRSOURCE:
		return rsrFld.ParseValue(cdr.Source)
	case utils.REQTYPE:
		return rsrFld.ParseValue(cdr.RequestType)
	case utils.TENANT:
		return rsrFld.ParseValue(cdr.Tenant)
	case utils.CATEGORY:
		return rsrFld.ParseValue(cdr.Category)
	case utils.Account:
		return rsrFld.ParseValue(cdr.Account)
	case utils.SUBJECT:
		return rsrFld.ParseValue(cdr.Subject)
	case utils.Destination:
		return rsrFld.ParseValue(cdr.Destination)
	case utils.SetupTime:
		return rsrFld.ParseValue(cdr.SetupTime.Format(time.RFC3339))
	case utils.AnswerTime:
		return rsrFld.ParseValue(cdr.AnswerTime.Format(time.RFC3339))
	case utils.Usage:
		return cdr.Usage.String()
	case utils.MEDI_RUNID:
		return rsrFld.ParseValue(cdr.RunID)
	case utils.RATED_FLD:
		return rsrFld.ParseValue(strconv.FormatBool(cdr.Rated))
	case utils.COST:
		return rsrFld.ParseValue(strconv.FormatFloat(cdr.Cost, 'f', -1, 64)) // Recommended to use FormatCost
	case utils.COST_DETAILS:
		return rsrFld.ParseValue(cdr.CostDetailsJson())
	case utils.PartialField:
		return rsrFld.ParseValue(strconv.FormatBool(cdr.Partial))
	default:
		return rsrFld.ParseValue(cdr.ExtraFields[rsrFld.Id])
	}
}

// Populates the field with id from value; strings are appended to original one
func (cdr *CDR) ParseFieldValue(fieldId, fieldVal, timezone string) error {
	var err error
	switch fieldId {
	case utils.ORDERID:
		if cdr.OrderID, err = strconv.ParseInt(fieldVal, 10, 64); err != nil {
			return err
		}
	case utils.TOR:
		cdr.ToR += fieldVal
	case utils.MEDI_RUNID:
		cdr.RunID += fieldVal
	case utils.ACCID:
		cdr.OriginID += fieldVal
	case utils.REQTYPE:
		cdr.RequestType += fieldVal
	case utils.TENANT:
		cdr.Tenant += fieldVal
	case utils.CATEGORY:
		cdr.Category += fieldVal
	case utils.Account:
		cdr.Account += fieldVal
	case utils.SUBJECT:
		cdr.Subject += fieldVal
	case utils.Destination:
		cdr.Destination += fieldVal
	case utils.RATED_FLD:
		cdr.Rated, _ = strconv.ParseBool(fieldVal)
	case utils.SetupTime:
		if cdr.SetupTime, err = utils.ParseTimeDetectLayout(fieldVal, timezone); err != nil {
			return fmt.Errorf("Cannot parse answer time field with value: %s, err: %s", fieldVal, err.Error())
		}
	case utils.AnswerTime:
		if cdr.AnswerTime, err = utils.ParseTimeDetectLayout(fieldVal, timezone); err != nil {
			return fmt.Errorf("Cannot parse answer time field with value: %s, err: %s", fieldVal, err.Error())
		}
	case utils.Usage:
		if cdr.Usage, err = utils.ParseDurationWithNanosecs(fieldVal); err != nil {
			return fmt.Errorf("Cannot parse duration field with value: %s, err: %s", fieldVal, err.Error())
		}
	case utils.COST:
		if cdr.Cost, err = strconv.ParseFloat(fieldVal, 64); err != nil {
			return fmt.Errorf("Cannot parse cost field with value: %s, err: %s", fieldVal, err.Error())
		}
	case utils.PartialField:
		cdr.Partial, _ = strconv.ParseBool(fieldVal)
	default: // Extra fields will not match predefined so they all show up here
		cdr.ExtraFields[fieldId] += fieldVal
	}
	return nil
}

// concatenates values of multiple fields defined in template, used eg in CDR templates
func (cdr *CDR) FieldsAsString(rsrFlds utils.RSRFields) string {
	var fldVal string
	for _, rsrFld := range rsrFlds {
		fldVal += cdr.FieldAsString(rsrFld)
	}
	return fldVal
}

func (cdr *CDR) AsCDR(timezone string) *CDR {
	return cdr
}

func (cdr *CDR) Clone() *CDR {
	var clnedCDR CDR
	utils.Clone(cdr, &clnedCDR)
	return &clnedCDR
}

// Used in mediation, primaryMandatory marks whether missing field out of request represents error or can be ignored
func (cdr *CDR) ForkCdr(runId string, RequestTypeFld, tenantFld, categFld, accountFld, subjectFld, destFld, setupTimeFld,
	answerTimeFld, durationFld, ratedFld, costFld *utils.RSRField,
	extraFlds []*utils.RSRField, primaryMandatory bool, timezone string) (*CDR, error) {
	if RequestTypeFld == nil {
		RequestTypeFld, _ = utils.NewRSRField(utils.META_DEFAULT)
	}
	if RequestTypeFld.Id == utils.META_DEFAULT {
		RequestTypeFld.Id = utils.REQTYPE
	}
	if tenantFld == nil {
		tenantFld, _ = utils.NewRSRField(utils.META_DEFAULT)
	}
	if tenantFld.Id == utils.META_DEFAULT {
		tenantFld.Id = utils.TENANT
	}
	if categFld == nil {
		categFld, _ = utils.NewRSRField(utils.META_DEFAULT)
	}
	if categFld.Id == utils.META_DEFAULT {
		categFld.Id = utils.CATEGORY
	}
	if accountFld == nil {
		accountFld, _ = utils.NewRSRField(utils.META_DEFAULT)
	}
	if accountFld.Id == utils.META_DEFAULT {
		accountFld.Id = utils.Account
	}
	if subjectFld == nil {
		subjectFld, _ = utils.NewRSRField(utils.META_DEFAULT)
	}
	if subjectFld.Id == utils.META_DEFAULT {
		subjectFld.Id = utils.SUBJECT
	}
	if destFld == nil {
		destFld, _ = utils.NewRSRField(utils.META_DEFAULT)
	}
	if destFld.Id == utils.META_DEFAULT {
		destFld.Id = utils.Destination
	}
	if setupTimeFld == nil {
		setupTimeFld, _ = utils.NewRSRField(utils.META_DEFAULT)
	}
	if setupTimeFld.Id == utils.META_DEFAULT {
		setupTimeFld.Id = utils.SetupTime
	}
	if answerTimeFld == nil {
		answerTimeFld, _ = utils.NewRSRField(utils.META_DEFAULT)
	}
	if answerTimeFld.Id == utils.META_DEFAULT {
		answerTimeFld.Id = utils.AnswerTime
	}
	if durationFld == nil {
		durationFld, _ = utils.NewRSRField(utils.META_DEFAULT)
	}
	if durationFld.Id == utils.META_DEFAULT {
		durationFld.Id = utils.Usage
	}
	if ratedFld == nil {
		ratedFld, _ = utils.NewRSRField(utils.META_DEFAULT)
	}
	if ratedFld.Id == utils.META_DEFAULT {
		ratedFld.Id = utils.RATED_FLD
	}
	if costFld == nil {
		costFld, _ = utils.NewRSRField(utils.META_DEFAULT)
	}
	if costFld.Id == utils.META_DEFAULT {
		costFld.Id = utils.COST
	}
	var err error
	frkStorCdr := new(CDR)
	frkStorCdr.CGRID = cdr.CGRID
	frkStorCdr.ToR = cdr.ToR
	frkStorCdr.RunID = runId
	frkStorCdr.Cost = -1.0 // Default for non-rated CDR
	frkStorCdr.OriginID = cdr.OriginID
	frkStorCdr.OriginHost = cdr.OriginHost
	frkStorCdr.Source = cdr.Source
	frkStorCdr.RequestType = cdr.FieldAsString(RequestTypeFld)
	if primaryMandatory && len(frkStorCdr.RequestType) == 0 {
		return nil, utils.NewErrMandatoryIeMissing(utils.REQTYPE, RequestTypeFld.Id)
	}
	frkStorCdr.Tenant = cdr.FieldAsString(tenantFld)
	if primaryMandatory && len(frkStorCdr.Tenant) == 0 {
		return nil, utils.NewErrMandatoryIeMissing(utils.TENANT, tenantFld.Id)
	}
	frkStorCdr.Category = cdr.FieldAsString(categFld)
	if primaryMandatory && len(frkStorCdr.Category) == 0 {
		return nil, utils.NewErrMandatoryIeMissing(utils.CATEGORY, categFld.Id)
	}
	frkStorCdr.Account = cdr.FieldAsString(accountFld)
	if primaryMandatory && len(frkStorCdr.Account) == 0 {
		return nil, utils.NewErrMandatoryIeMissing(utils.Account, accountFld.Id)
	}
	frkStorCdr.Subject = cdr.FieldAsString(subjectFld)
	if primaryMandatory && len(frkStorCdr.Subject) == 0 {
		return nil, utils.NewErrMandatoryIeMissing(utils.SUBJECT, subjectFld.Id)
	}
	frkStorCdr.Destination = cdr.FieldAsString(destFld)
	if primaryMandatory && len(frkStorCdr.Destination) == 0 && frkStorCdr.ToR == utils.VOICE {
		return nil, utils.NewErrMandatoryIeMissing(utils.Destination, destFld.Id)
	}
	sTimeStr := cdr.FieldAsString(setupTimeFld)
	if primaryMandatory && len(sTimeStr) == 0 {
		return nil, utils.NewErrMandatoryIeMissing(utils.SetupTime, setupTimeFld.Id)
	} else if frkStorCdr.SetupTime, err = utils.ParseTimeDetectLayout(sTimeStr, timezone); err != nil {
		return nil, err
	}
	aTimeStr := cdr.FieldAsString(answerTimeFld)
	if primaryMandatory && len(aTimeStr) == 0 {
		return nil, utils.NewErrMandatoryIeMissing(utils.AnswerTime, answerTimeFld.Id)
	} else if frkStorCdr.AnswerTime, err = utils.ParseTimeDetectLayout(aTimeStr, timezone); err != nil {
		return nil, err
	}
	durStr := cdr.FieldAsString(durationFld)
	if primaryMandatory && len(durStr) == 0 {
		return nil, utils.NewErrMandatoryIeMissing(utils.Usage, durationFld.Id)
	} else if frkStorCdr.Usage, err = utils.ParseDurationWithNanosecs(durStr); err != nil {
		return nil, err
	}
	ratedStr := cdr.FieldAsString(ratedFld)
	if primaryMandatory && len(ratedStr) == 0 {
		return nil, utils.NewErrMandatoryIeMissing(utils.RATED_FLD, ratedFld.Id)
	} else if frkStorCdr.Rated, err = strconv.ParseBool(ratedStr); err != nil {
		return nil, err
	}
	costStr := cdr.FieldAsString(costFld)
	if primaryMandatory && len(costStr) == 0 {
		return nil, utils.NewErrMandatoryIeMissing(utils.COST, costFld.Id)
	} else if frkStorCdr.Cost, err = strconv.ParseFloat(costStr, 64); err != nil {
		return nil, err
	}
	frkStorCdr.ExtraFields = make(map[string]string, len(extraFlds))
	for _, fld := range extraFlds {
		frkStorCdr.ExtraFields[fld.Id] = cdr.FieldAsString(fld)
	}
	return frkStorCdr, nil
}

func (cdr *CDR) AsExternalCDR() *ExternalCDR {
	var usageStr string
	switch cdr.ToR {
	case utils.VOICE: // usage as time
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
		Rated:       cdr.Rated,
	}
}

// Implementation of Event interface, used in tests
func (cdr *CDR) AsEvent(ignored string) Event {
	return Event(cdr)
}
func (cdr *CDR) ComputeLcr() bool {
	return false
}
func (cdr *CDR) GetName() string {
	return cdr.Source
}
func (cdr *CDR) GetCgrId(timezone string) string {
	return cdr.CGRID
}
func (cdr *CDR) GetUUID() string {
	return cdr.OriginID
}
func (cdr *CDR) GetSessionIds() []string {
	return []string{cdr.GetUUID()}
}
func (cdr *CDR) GetSubject(fieldName string) string {
	if utils.IsSliceMember([]string{utils.SUBJECT, utils.META_DEFAULT, ""}, fieldName) {
		return cdr.Subject
	}
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return cdr.FieldAsString(&utils.RSRField{Id: fieldName})
}
func (cdr *CDR) GetAccount(fieldName string) string {
	if utils.IsSliceMember([]string{utils.Account, utils.META_DEFAULT, ""}, fieldName) {
		return cdr.Account
	}
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return cdr.FieldAsString(&utils.RSRField{Id: fieldName})
}
func (cdr *CDR) GetDestination(fieldName string) string {
	if utils.IsSliceMember([]string{utils.Destination, utils.META_DEFAULT, ""}, fieldName) {
		return cdr.Destination
	}
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return cdr.FieldAsString(&utils.RSRField{Id: fieldName})
}
func (cdr *CDR) GetCallDestNr(fieldName string) string {
	if utils.IsSliceMember([]string{utils.Destination, utils.META_DEFAULT, ""}, fieldName) {
		return cdr.Destination
	}
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return cdr.FieldAsString(&utils.RSRField{Id: fieldName})
}
func (cdr *CDR) GetCategory(fieldName string) string {
	if utils.IsSliceMember([]string{utils.CATEGORY, utils.META_DEFAULT, ""}, fieldName) {
		return cdr.Category
	}
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return cdr.FieldAsString(&utils.RSRField{Id: fieldName})
}
func (cdr *CDR) GetTenant(fieldName string) string {
	if utils.IsSliceMember([]string{utils.TENANT, utils.META_DEFAULT, ""}, fieldName) {
		return cdr.Tenant
	}
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return cdr.FieldAsString(&utils.RSRField{Id: fieldName})
}
func (cdr *CDR) GetReqType(fieldName string) string {
	if utils.IsSliceMember([]string{utils.REQTYPE, utils.META_DEFAULT, ""}, fieldName) {
		return cdr.RequestType
	}
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return cdr.FieldAsString(&utils.RSRField{Id: fieldName})
}
func (cdr *CDR) GetSetupTime(fieldName, timezone string) (time.Time, error) {
	if utils.IsSliceMember([]string{utils.SetupTime, utils.META_DEFAULT, ""}, fieldName) {
		return cdr.SetupTime, nil
	}
	var sTimeVal string
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		sTimeVal = fieldName[len(utils.STATIC_VALUE_PREFIX):]
	} else {
		sTimeVal = cdr.FieldAsString(&utils.RSRField{Id: fieldName})
	}
	return utils.ParseTimeDetectLayout(sTimeVal, timezone)
}
func (cdr *CDR) GetAnswerTime(fieldName, timezone string) (time.Time, error) {
	if utils.IsSliceMember([]string{utils.AnswerTime, utils.META_DEFAULT, ""}, fieldName) {
		return cdr.AnswerTime, nil
	}
	var aTimeVal string
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		aTimeVal = fieldName[len(utils.STATIC_VALUE_PREFIX):]
	} else {
		aTimeVal = cdr.FieldAsString(&utils.RSRField{Id: fieldName})
	}
	return utils.ParseTimeDetectLayout(aTimeVal, timezone)
}
func (cdr *CDR) GetEndTime(fieldName, timezone string) (time.Time, error) {
	return cdr.AnswerTime.Add(cdr.Usage), nil
}
func (cdr *CDR) GetDuration(fieldName string) (time.Duration, error) {
	if utils.IsSliceMember([]string{utils.Usage, utils.META_DEFAULT, ""}, fieldName) {
		return cdr.Usage, nil
	}
	var durVal string
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		durVal = fieldName[len(utils.STATIC_VALUE_PREFIX):]
	} else {
		durVal = cdr.FieldAsString(&utils.RSRField{Id: fieldName})
	}
	return utils.ParseDurationWithNanosecs(durVal)
}
func (cdr *CDR) GetOriginatorIP(fieldName string) string {
	if utils.IsSliceMember([]string{utils.CDRHOST, utils.META_DEFAULT, ""}, fieldName) {
		return cdr.OriginHost
	}
	return cdr.FieldAsString(&utils.RSRField{Id: fieldName})
}
func (cdr *CDR) GetExtraFields() map[string]string {
	return cdr.ExtraFields
}
func (cdr *CDR) MissingParameter(timezone string) bool {
	return len(cdr.OriginID) == 0 ||
		len(cdr.Category) == 0 ||
		len(cdr.Tenant) == 0 ||
		len(cdr.Account) == 0 ||
		len(cdr.Destination) == 0
}
func (cdr *CDR) ParseEventValue(rsrFld *utils.RSRField, timezone string) string {
	return cdr.FieldAsString(rsrFld)
}
func (cdr *CDR) String() string {
	mrsh, _ := json.Marshal(cdr)
	return string(mrsh)
}

func (cdr *CDR) combimedCdrFieldVal(cfgCdrFld *config.CfgCdrField, groupCDRs []*CDR) (string, error) {
	var combinedVal string // Will result as combination of the field values, filters must match
	for _, filterRule := range cfgCdrFld.FieldFilter {
		if !filterRule.FilterPasses(cdr.FieldAsString(&utils.RSRField{Id: filterRule.Id})) { // Filter will activate the rule to extract the content
			continue
		}
		pairingVal := cdr.FieldAsString(filterRule)
		for _, grpCDR := range groupCDRs {
			if cdr.CGRID != grpCDR.CGRID {
				continue // We only care about cdrs with same primary cdr behind
			}
			if grpCDR.FieldAsString(&utils.RSRField{Id: filterRule.Id}) != pairingVal { // First CDR with field equal with ours
				continue
			}
			for _, rsrRule := range cfgCdrFld.Value {
				combinedVal += grpCDR.FieldAsString(rsrRule)
			}
		}
	}
	return combinedVal, nil
}

// Extracts the value specified by cfgHdr out of cdr, used for export values
func (cdr *CDR) exportFieldValue(cfgCdrFld *config.CfgCdrField) (string, error) {
	passesFilters := true
	for _, cdfFltr := range cfgCdrFld.FieldFilter {
		if !cdfFltr.FilterPasses(cdr.FieldAsString(cdfFltr)) {
			passesFilters = false
			break
		}
	}
	if !passesFilters { // Not passes filters, ignore this replication
		return "", fmt.Errorf("filters not passing")
	}
	var retVal string // Concatenate the resulting values
	for _, rsrFld := range cfgCdrFld.Value {
		var cdrVal string
		switch rsrFld.Id {
		case utils.COST:
			cdrVal = cdr.FormatCost(cfgCdrFld.CostShiftDigits, cfgCdrFld.RoundingDecimals)
		case utils.Usage:
			cdrVal = cdr.FormatUsage(cfgCdrFld.Layout)
		case utils.SetupTime:
			cdrVal = cdr.SetupTime.Format(cfgCdrFld.Layout)
		case utils.AnswerTime: // Format time based on layout
			cdrVal = cdr.AnswerTime.Format(cfgCdrFld.Layout)
		case utils.Destination:
			cdrVal = cdr.FieldAsString(rsrFld)
			if cfgCdrFld.MaskLen != -1 && len(cfgCdrFld.MaskDestID) != 0 && CachedDestHasPrefix(cfgCdrFld.MaskDestID, cdrVal) {
				cdrVal = utils.MaskSuffix(cdrVal, cfgCdrFld.MaskLen)
			}
		default:
			cdrVal = cdr.FieldAsString(rsrFld)
		}
		retVal += cdrVal
	}
	return retVal, nil
}

func (cdr *CDR) formatField(cfgFld *config.CfgCdrField, httpSkipTlsCheck bool, groupedCDRs []*CDR) (fmtOut string, err error) {
	layout := cfgFld.Layout
	if layout == "" {
		layout = time.RFC3339
	}
	var outVal string
	switch cfgFld.Type {
	case utils.META_FILLER:
		outVal = cfgFld.Value.Id()
		cfgFld.Padding = "right"
	case utils.META_CONSTANT:
		outVal = cfgFld.Value.Id()
	case utils.MetaDateTime: // Convert the requested field value into datetime with layout
		rawVal, err := cdr.exportFieldValue(cfgFld)
		if err != nil {
			return "", err
		}
		if dtFld, err := utils.ParseTimeDetectLayout(rawVal, cfgFld.Timezone); err != nil { // Only one rule makes sense here
			return "", err
		} else {
			outVal = dtFld.Format(layout)
		}
	case utils.META_HTTP_POST:
		var outValByte []byte
		httpAddr := cfgFld.Value.Id()
		jsn, err := json.Marshal(cdr)
		if err != nil {
			return "", err
		}
		if len(httpAddr) == 0 {
			err = fmt.Errorf("Empty http address for field %s type %s", cfgFld.Tag, cfgFld.Type)
		} else if outValByte, err = utils.HttpJsonPost(httpAddr, httpSkipTlsCheck, jsn); err == nil {
			outVal = string(outValByte)
			if len(outVal) == 0 && cfgFld.Mandatory {
				err = fmt.Errorf("Empty result for http_post field: %s", cfgFld.Tag)
			}
		}
	case utils.META_COMBIMED:
		outVal, err = cdr.combimedCdrFieldVal(cfgFld, groupedCDRs)
	case utils.META_COMPOSED:
		outVal, err = cdr.exportFieldValue(cfgFld)
	case utils.MetaMaskedDestination:
		if len(cfgFld.MaskDestID) != 0 && CachedDestHasPrefix(cfgFld.MaskDestID, cdr.Destination) {
			outVal = "1"
		} else {
			outVal = "0"
		}
	}
	if err != nil {
		return "", err
	}
	return utils.FmtFieldWidth(cfgFld.Tag, outVal, cfgFld.Width, cfgFld.Strip, cfgFld.Padding, cfgFld.Mandatory)

}

// Part of event interface
func (cdr *CDR) AsMapStringIface() (mp map[string]interface{}, err error) {
	mp = make(map[string]interface{})
	for k, v := range cdr.ExtraFields {
		mp[k] = v
	}
	mp[utils.CGRID] = cdr.CGRID
	mp[utils.MEDI_RUNID] = cdr.RunID
	mp[utils.ORDERID] = cdr.OrderID
	mp[utils.CDRHOST] = cdr.OriginHost
	mp[utils.CDRSOURCE] = cdr.Source
	mp[utils.ACCID] = cdr.OriginID
	mp[utils.TOR] = cdr.ToR
	mp[utils.REQTYPE] = cdr.RequestType
	mp[utils.TENANT] = cdr.Tenant
	mp[utils.CATEGORY] = cdr.Category
	mp[utils.Account] = cdr.Account
	mp[utils.SUBJECT] = cdr.Subject
	mp[utils.Destination] = cdr.Destination
	mp[utils.SetupTime] = cdr.SetupTime
	mp[utils.AnswerTime] = cdr.AnswerTime
	mp[utils.Usage] = cdr.Usage
	mp[utils.CostSource] = cdr.CostSource
	mp[utils.COST] = cdr.Cost
	mp[utils.COST_DETAILS] = cdr.CostDetails
	mp[utils.ExtraInfo] = cdr.ExtraInfo
	mp[utils.RATED] = cdr.Rated
	mp[utils.PartialField] = cdr.Partial
	return
}

// Used in place where we need to export the CDR based on an export template
// ExportRecord is a []string to keep it compatible with encoding/csv Writer
func (cdr *CDR) AsExportRecord(exportFields []*config.CfgCdrField, httpSkipTlsCheck bool, groupedCDRs []*CDR, roundingDecs int) (expRecord []string, err error) {
	expRecord = make([]string, len(exportFields))
	for idx, cfgFld := range exportFields {
		if roundingDecs != 0 {
			clnFld := new(config.CfgCdrField) // Clone so we can modify the rounding decimals without affecting the template
			*clnFld = *cfgFld
			clnFld.RoundingDecimals = roundingDecs
			cfgFld = clnFld
		}
		if fmtOut, err := cdr.formatField(cfgFld, httpSkipTlsCheck, groupedCDRs); err != nil {
			return nil, err
		} else {
			expRecord[idx] += fmtOut
		}
	}
	return expRecord, nil
}

// AsExportMap converts the CDR into a map[string]string based on export template
// Used in real-time replication as well as remote exports
func (cdr *CDR) AsExportMap(exportFields []*config.CfgCdrField, httpSkipTlsCheck bool, groupedCDRs []*CDR, roundingDecs int) (expMap map[string]string, err error) {
	expMap = make(map[string]string)
	for _, cfgFld := range exportFields {
		if roundingDecs != 0 {
			clnFld := new(config.CfgCdrField) // Clone so we can modify the rounding decimals without affecting the template
			*clnFld = *cfgFld
			clnFld.RoundingDecimals = roundingDecs
			cfgFld = clnFld
		}
		if fmtOut, err := cdr.formatField(cfgFld, httpSkipTlsCheck, groupedCDRs); err != nil {
			return nil, err
		} else {
			expMap[cfgFld.FieldId] += fmtOut
		}
	}
	return
}

// AsCDRsTBL converts the CDR into the format used for SQL storage
func (cdr *CDR) AsCDRsql() (cdrSql *CDRsql) {
	cdrSql = new(CDRsql)
	cdrSql.Cgrid = cdr.CGRID
	cdrSql.RunID = cdr.RunID
	cdrSql.OriginHost = cdr.OriginHost
	cdrSql.Source = cdr.Source
	cdrSql.OriginID = cdr.OriginID
	cdrSql.TOR = cdr.ToR
	cdrSql.RequestType = cdr.RequestType
	cdrSql.Tenant = cdr.Tenant
	cdrSql.Category = cdr.Category
	cdrSql.Account = cdr.Account
	cdrSql.Subject = cdr.Subject
	cdrSql.Destination = cdr.Destination
	cdrSql.SetupTime = cdr.SetupTime
	cdrSql.AnswerTime = cdr.AnswerTime
	cdrSql.Usage = cdr.Usage.Nanoseconds()
	cdrSql.ExtraFields = utils.ToJSON(cdr.ExtraFields)
	cdrSql.CostSource = cdr.CostSource
	cdrSql.Cost = cdr.Cost
	cdrSql.CostDetails = utils.ToJSON(cdr.CostDetails)
	cdrSql.ExtraInfo = cdr.ExtraInfo
	cdrSql.CreatedAt = time.Now()
	return
}

func (cdr *CDR) AsCGREvent() *utils.CGREvent {
	cdrIf, _ := cdr.AsMapStringIface()
	return &utils.CGREvent{
		Tenant: cdr.Tenant,
		ID:     utils.UUIDSha1Prefix(),
		Event:  cdrIf,
	}
}

// UpdateFromCGREvent will update CDR with event fields from CGREvent
func (cdr *CDR) UpdateFromCGREvent(cgrEv *utils.CGREvent, fields []string) (err error) {
	for _, fldName := range fields {
		switch fldName {
		case utils.CDRHOST:
			if cdr.OriginHost, err = cgrEv.FieldAsString(fldName); err != nil {
				return
			}
		case utils.CDRSOURCE:
			if cdr.Source, err = cgrEv.FieldAsString(fldName); err != nil {
				return
			}
		case utils.TOR:
			if cdr.ToR, err = cgrEv.FieldAsString(fldName); err != nil {
				return
			}
		case utils.REQTYPE:
			if cdr.RequestType, err = cgrEv.FieldAsString(fldName); err != nil {
				return
			}
		case utils.TENANT:
			if cdr.Tenant, err = cgrEv.FieldAsString(fldName); err != nil {
				return
			}
		case utils.CATEGORY:
			if cdr.Category, err = cgrEv.FieldAsString(fldName); err != nil {
				return
			}
		case utils.Account:
			if cdr.Account, err = cgrEv.FieldAsString(fldName); err != nil {
				return
			}
		case utils.SUBJECT:
			if cdr.Subject, err = cgrEv.FieldAsString(fldName); err != nil {
				return
			}
		case utils.Destination:
			if cdr.Destination, err = cgrEv.FieldAsString(fldName); err != nil {
				return
			}
		case utils.SetupTime:
			if cdr.SetupTime, err = cgrEv.FieldAsTime(fldName, config.CgrConfig().DefaultTimezone); err != nil {
				return
			}
		case utils.AnswerTime:
			if cdr.AnswerTime, err = cgrEv.FieldAsTime(fldName, config.CgrConfig().DefaultTimezone); err != nil {
				return
			}
		case utils.Usage:
			if cdr.Usage, err = cgrEv.FieldAsDuration(fldName); err != nil {
				return
			}
		default:
			fldVal, err := cgrEv.FieldAsString(fldName)
			if err != nil {
				return err
			}
			cdr.ExtraFields[fldName] = fldVal
		}
	}
	return
}

// NewCDRFromSQL converts the CDRsql into CDR
func NewCDRFromSQL(cdrSql *CDRsql) (cdr *CDR, err error) {
	cdr = new(CDR)
	cdr.CGRID = cdrSql.Cgrid
	cdr.RunID = cdrSql.RunID
	cdr.OriginHost = cdrSql.OriginHost
	cdr.Source = cdrSql.Source
	cdr.OriginID = cdrSql.OriginID
	cdr.OrderID = cdrSql.ID
	cdr.ToR = cdrSql.TOR
	cdr.RequestType = cdrSql.RequestType
	cdr.Tenant = cdrSql.Tenant
	cdr.Category = cdrSql.Category
	cdr.Account = cdrSql.Account
	cdr.Subject = cdrSql.Subject
	cdr.Destination = cdrSql.Destination
	cdr.SetupTime = cdrSql.SetupTime
	cdr.AnswerTime = cdrSql.AnswerTime
	cdr.Usage = time.Duration(cdrSql.Usage)
	cdr.CostSource = cdrSql.CostSource
	cdr.Cost = cdrSql.Cost
	cdr.ExtraInfo = cdrSql.ExtraInfo
	if cdrSql.ExtraFields != "" {
		if err = json.Unmarshal([]byte(cdrSql.ExtraFields), &cdr.ExtraFields); err != nil {
			return nil, err
		}
	}
	if cdrSql.CostDetails != "" {
		if err = json.Unmarshal([]byte(cdrSql.CostDetails), &cdr.CostDetails); err != nil {
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
	Rated       bool // Mark the CDR as rated so we do not process it during mediation
}

// Used when authorizing requests from outside, eg ApierV1.GetMaxUsage
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

func (self *UsageRecord) AsCDR(timezone string) (*CDR, error) {
	var err error
	cdr := &CDR{CGRID: self.GetId(), ToR: self.ToR, RequestType: self.RequestType, Tenant: self.Tenant,
		Category: self.Category, Account: self.Account, Subject: self.Subject, Destination: self.Destination}
	if cdr.SetupTime, err = utils.ParseTimeDetectLayout(self.SetupTime, timezone); err != nil {
		return nil, err
	}
	if cdr.AnswerTime, err = utils.ParseTimeDetectLayout(self.AnswerTime, timezone); err != nil {
		return nil, err
	}
	if cdr.Usage, err = utils.ParseDurationWithNanosecs(self.Usage); err != nil {
		return nil, err
	}
	if self.ExtraFields != nil {
		cdr.ExtraFields = make(map[string]string)
	}
	for k, v := range self.ExtraFields {
		cdr.ExtraFields[k] = v
	}
	return cdr, nil
}

func (self *UsageRecord) AsCallDescriptor(timezone string, denyNegative bool) (*CallDescriptor, error) {
	var err error
	cd := &CallDescriptor{
		CgrID:               self.GetId(),
		TOR:                 self.ToR,
		Direction:           utils.OUT,
		Tenant:              self.Tenant,
		Category:            self.Category,
		Subject:             self.Subject,
		Account:             self.Account,
		Destination:         self.Destination,
		DenyNegativeAccount: denyNegative,
	}
	timeStr := self.AnswerTime
	if len(timeStr) == 0 { // In case of auth, answer time will not be defined, so take it out of setup one
		timeStr = self.SetupTime
	}
	if cd.TimeStart, err = utils.ParseTimeDetectLayout(timeStr, timezone); err != nil {
		return nil, err
	}
	if usage, err := utils.ParseDurationWithNanosecs(self.Usage); err != nil {
		return nil, err
	} else {
		cd.TimeEnd = cd.TimeStart.Add(usage)
	}
	if self.ExtraFields != nil {
		cd.ExtraFields = make(map[string]string)
	}
	for k, v := range self.ExtraFields {
		cd.ExtraFields[k] = v
	}
	return cd, nil
}

func (self *UsageRecord) GetId() string {
	return utils.Sha1(self.ToR, self.RequestType, self.Tenant, self.Category, self.Account, self.Subject, self.Destination, self.SetupTime, self.AnswerTime, self.Usage)
}
