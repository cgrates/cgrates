/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2015 ITsysCOM GmbH

This program is free software: you can Storagetribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, ornt
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITH*out ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Kinda standard of internal CDR, complies to CDR interface also
type StoredCdr struct {
	CgrId          string
	OrderId        int64             // Stor order id used as export order id
	TOR            string            // type of record, meta-field, should map to one of the TORs hardcoded inside the server <*voice|*data|*sms>
	AccId          string            // represents the unique accounting id given by the telecom switch generating the CDR
	CdrHost        string            // represents the IP address of the host generating the CDR (automatically populated by the server)
	CdrSource      string            // formally identifies the source of the CDR (free form field)
	ReqType        string            // matching the supported request types by the **CGRateS**, accepted values are hardcoded in the server <prepaid|postpaid|pseudoprepaid|rated>.
	Direction      string            // matching the supported direction identifiers of the CGRateS <*out>
	Tenant         string            // tenant whom this record belongs
	Category       string            // free-form filter for this record, matching the category defined in rating profiles.
	Account        string            // account id (accounting subsystem) the record should be attached to
	Subject        string            // rating subject (rating subsystem) this record should be attached to
	Destination    string            // destination to be charged
	SetupTime      time.Time         // set-up time of the event. Supported formats: datetime RFC3339 compatible, SQL datetime (eg: MySQL), unix timestamp.
	AnswerTime     time.Time         // answer time of the event. Supported formats: datetime RFC3339 compatible, SQL datetime (eg: MySQL), unix timestamp.
	Usage          time.Duration     // event usage information (eg: in case of tor=*voice this will represent the total duration of a call)
	ExtraFields    map[string]string // Extra fields to be stored in CDR
	MediationRunId string
	RatedAccount   string // Populated out of rating data
	RatedSubject   string
	Cost           float64
	Rated          bool // Mark the CDR as rated so we do not process it during mediation
}

// Used to multiply usage on export
func (storedCdr *StoredCdr) UsageMultiply(multiplyFactor float64, roundDecimals int) {
	storedCdr.Usage = time.Duration(int(Round(float64(storedCdr.Usage.Nanoseconds())*multiplyFactor, roundDecimals, ROUNDING_MIDDLE))) // Rounding down could introduce a slight loss here but only at nanoseconds level
}

// Used to multiply cost on export
func (storedCdr *StoredCdr) CostMultiply(multiplyFactor float64, roundDecimals int) {
	storedCdr.Cost = Round(storedCdr.Cost*multiplyFactor, roundDecimals, ROUNDING_MIDDLE)
}

// Format cost as string on export
func (storedCdr *StoredCdr) FormatCost(shiftDecimals, roundDecimals int) string {
	cost := storedCdr.Cost
	if shiftDecimals != 0 {
		cost = cost * math.Pow10(shiftDecimals)
	}
	return strconv.FormatFloat(cost, 'f', roundDecimals, 64)
}

// Formats usage on export
func (storedCdr *StoredCdr) FormatUsage(layout string) string {
	if IsSliceMember([]string{DATA, SMS}, storedCdr.TOR) {
		return strconv.FormatFloat(Round(storedCdr.Usage.Seconds(), 0, ROUNDING_MIDDLE), 'f', -1, 64)
	}
	switch layout {
	default:
		return strconv.FormatFloat(float64(storedCdr.Usage.Nanoseconds())/1000000000, 'f', -1, 64)
	}
}

// Used to retrieve fields as string, primary fields are const labeled
func (storedCdr *StoredCdr) FieldAsString(rsrFld *RSRField) string {
	switch rsrFld.Id {
	case CGRID:
		return rsrFld.ParseValue(storedCdr.CgrId)
	case ORDERID:
		return rsrFld.ParseValue(strconv.FormatInt(storedCdr.OrderId, 10))
	case TOR:
		return rsrFld.ParseValue(storedCdr.TOR)
	case ACCID:
		return rsrFld.ParseValue(storedCdr.AccId)
	case CDRHOST:
		return rsrFld.ParseValue(storedCdr.CdrHost)
	case CDRSOURCE:
		return rsrFld.ParseValue(storedCdr.CdrSource)
	case REQTYPE:
		return rsrFld.ParseValue(storedCdr.ReqType)
	case DIRECTION:
		return rsrFld.ParseValue(storedCdr.Direction)
	case TENANT:
		return rsrFld.ParseValue(storedCdr.Tenant)
	case CATEGORY:
		return rsrFld.ParseValue(storedCdr.Category)
	case ACCOUNT:
		return rsrFld.ParseValue(storedCdr.Account)
	case SUBJECT:
		return rsrFld.ParseValue(storedCdr.Subject)
	case DESTINATION:
		return rsrFld.ParseValue(storedCdr.Destination)
	case SETUP_TIME:
		return rsrFld.ParseValue(storedCdr.SetupTime.String())
	case ANSWER_TIME:
		return rsrFld.ParseValue(storedCdr.AnswerTime.String())
	case USAGE:
		return strconv.FormatFloat(Round(storedCdr.Usage.Seconds(), 0, ROUNDING_MIDDLE), 'f', -1, 64)
	case MEDI_RUNID:
		return rsrFld.ParseValue(storedCdr.MediationRunId)
	case RATED_ACCOUNT:
		return rsrFld.ParseValue(storedCdr.RatedAccount)
	case RATED_SUBJECT:
		return rsrFld.ParseValue(storedCdr.RatedSubject)
	case COST:
		return rsrFld.ParseValue(strconv.FormatFloat(storedCdr.Cost, 'f', -1, 64)) // Recommended to use FormatCost
	default:
		return rsrFld.ParseValue(storedCdr.ExtraFields[rsrFld.Id])
	}
}

func (storedCdr *StoredCdr) PassesFieldFilter(fieldFilter *RSRField) (bool, string) {
	if fieldFilter == nil {
		return true, ""
	}
	if fieldFilter.IsStatic() && storedCdr.FieldAsString(&RSRField{Id: fieldFilter.Id}) == storedCdr.FieldAsString(fieldFilter) {
		return true, storedCdr.FieldAsString(&RSRField{Id: fieldFilter.Id})
	}
	preparedFilter := &RSRField{Id: fieldFilter.Id, RSRules: make([]*ReSearchReplace, len(fieldFilter.RSRules))} // Reset rules so they do not point towards same structures as original fieldFilter
	for idx := range fieldFilter.RSRules {
		// Hardcode the template with maximum of 5 groups ordered
		preparedFilter.RSRules[idx] = &ReSearchReplace{SearchRegexp: fieldFilter.RSRules[idx].SearchRegexp, ReplaceTemplate: FILTER_REGEXP_TPL}
	}
	preparedVal := storedCdr.FieldAsString(preparedFilter)
	filteredValue := storedCdr.FieldAsString(fieldFilter)
	if preparedFilter.RegexpMatched() && (len(preparedVal) == 0 || preparedVal == filteredValue) {
		return true, filteredValue
	}
	return false, ""
}

func (storedCdr *StoredCdr) AsStoredCdr() *StoredCdr {
	return storedCdr
}

// Ability to send the CgrCdr remotely to another CDR server, we do not include rating variables for now
func (storedCdr *StoredCdr) AsHttpForm() url.Values {
	v := url.Values{}
	for fld, val := range storedCdr.ExtraFields {
		v.Set(fld, val)
	}
	v.Set(TOR, storedCdr.TOR)
	v.Set(ACCID, storedCdr.AccId)
	v.Set(CDRHOST, storedCdr.CdrHost)
	v.Set(CDRSOURCE, storedCdr.CdrSource)
	v.Set(REQTYPE, storedCdr.ReqType)
	v.Set(DIRECTION, storedCdr.Direction)
	v.Set(TENANT, storedCdr.Tenant)
	v.Set(CATEGORY, storedCdr.Category)
	v.Set(ACCOUNT, storedCdr.Account)
	v.Set(SUBJECT, storedCdr.Subject)
	v.Set(DESTINATION, storedCdr.Destination)
	v.Set(SETUP_TIME, storedCdr.SetupTime.String())
	v.Set(ANSWER_TIME, storedCdr.AnswerTime.String())
	v.Set(USAGE, storedCdr.FormatUsage(SECONDS))
	return v
}

// Used in mediation, primaryMandatory marks whether missing field out of request represents error or can be ignored
func (storedCdr *StoredCdr) ForkCdr(runId string, reqTypeFld, directionFld, tenantFld, categFld, accountFld, subjectFld, destFld, setupTimeFld, answerTimeFld, durationFld *RSRField,
	extraFlds []*RSRField, primaryMandatory bool) (*StoredCdr, error) {
	if reqTypeFld == nil {
		reqTypeFld, _ = NewRSRField(META_DEFAULT)
	}
	if reqTypeFld.Id == META_DEFAULT {
		reqTypeFld.Id = REQTYPE
	}
	if directionFld == nil {
		directionFld, _ = NewRSRField(META_DEFAULT)
	}
	if directionFld.Id == META_DEFAULT {
		directionFld.Id = DIRECTION
	}
	if tenantFld == nil {
		tenantFld, _ = NewRSRField(META_DEFAULT)
	}
	if tenantFld.Id == META_DEFAULT {
		tenantFld.Id = TENANT
	}
	if categFld == nil {
		categFld, _ = NewRSRField(META_DEFAULT)
	}
	if categFld.Id == META_DEFAULT {
		categFld.Id = CATEGORY
	}
	if accountFld == nil {
		accountFld, _ = NewRSRField(META_DEFAULT)
	}
	if accountFld.Id == META_DEFAULT {
		accountFld.Id = ACCOUNT
	}
	if subjectFld == nil {
		subjectFld, _ = NewRSRField(META_DEFAULT)
	}
	if subjectFld.Id == META_DEFAULT {
		subjectFld.Id = SUBJECT
	}
	if destFld == nil {
		destFld, _ = NewRSRField(META_DEFAULT)
	}
	if destFld.Id == META_DEFAULT {
		destFld.Id = DESTINATION
	}
	if setupTimeFld == nil {
		setupTimeFld, _ = NewRSRField(META_DEFAULT)
	}
	if setupTimeFld.Id == META_DEFAULT {
		setupTimeFld.Id = SETUP_TIME
	}
	if answerTimeFld == nil {
		answerTimeFld, _ = NewRSRField(META_DEFAULT)
	}
	if answerTimeFld.Id == META_DEFAULT {
		answerTimeFld.Id = ANSWER_TIME
	}
	if durationFld == nil {
		durationFld, _ = NewRSRField(META_DEFAULT)
	}
	if durationFld.Id == META_DEFAULT {
		durationFld.Id = USAGE
	}
	var err error
	frkStorCdr := new(StoredCdr)
	frkStorCdr.CgrId = storedCdr.CgrId
	frkStorCdr.TOR = storedCdr.TOR
	frkStorCdr.MediationRunId = runId
	frkStorCdr.Cost = -1.0 // Default for non-rated CDR
	frkStorCdr.AccId = storedCdr.AccId
	frkStorCdr.CdrHost = storedCdr.CdrHost
	frkStorCdr.CdrSource = storedCdr.CdrSource
	frkStorCdr.ReqType = storedCdr.FieldAsString(reqTypeFld)
	if primaryMandatory && len(frkStorCdr.ReqType) == 0 {
		return nil, errors.New(fmt.Sprintf("%s:%s:%s", ERR_MANDATORY_IE_MISSING, REQTYPE, reqTypeFld.Id))
	}
	frkStorCdr.Direction = storedCdr.FieldAsString(directionFld)
	if primaryMandatory && len(frkStorCdr.Direction) == 0 {
		return nil, errors.New(fmt.Sprintf("%s:%s:%s", ERR_MANDATORY_IE_MISSING, DIRECTION, directionFld.Id))
	}
	frkStorCdr.Tenant = storedCdr.FieldAsString(tenantFld)
	if primaryMandatory && len(frkStorCdr.Tenant) == 0 {
		return nil, errors.New(fmt.Sprintf("%s:%s:%s", ERR_MANDATORY_IE_MISSING, TENANT, tenantFld.Id))
	}
	frkStorCdr.Category = storedCdr.FieldAsString(categFld)
	if primaryMandatory && len(frkStorCdr.Category) == 0 {
		return nil, errors.New(fmt.Sprintf("%s:%s:%s", ERR_MANDATORY_IE_MISSING, CATEGORY, categFld.Id))
	}
	frkStorCdr.Account = storedCdr.FieldAsString(accountFld)
	if primaryMandatory && len(frkStorCdr.Account) == 0 {
		return nil, errors.New(fmt.Sprintf("%s:%s:%s", ERR_MANDATORY_IE_MISSING, ACCOUNT, accountFld.Id))
	}
	frkStorCdr.Subject = storedCdr.FieldAsString(subjectFld)
	if primaryMandatory && len(frkStorCdr.Subject) == 0 {
		return nil, errors.New(fmt.Sprintf("%s:%s:%s", ERR_MANDATORY_IE_MISSING, SUBJECT, subjectFld.Id))
	}
	frkStorCdr.Destination = storedCdr.FieldAsString(destFld)
	if primaryMandatory && len(frkStorCdr.Destination) == 0 && frkStorCdr.TOR == VOICE {
		return nil, errors.New(fmt.Sprintf("%s:%s:%s", ERR_MANDATORY_IE_MISSING, DESTINATION, destFld.Id))
	}
	sTimeStr := storedCdr.FieldAsString(setupTimeFld)
	if primaryMandatory && len(sTimeStr) == 0 {
		return nil, errors.New(fmt.Sprintf("%s:%s:%s", ERR_MANDATORY_IE_MISSING, SETUP_TIME, setupTimeFld.Id))
	} else if frkStorCdr.SetupTime, err = ParseTimeDetectLayout(sTimeStr); err != nil {
		return nil, err
	}
	aTimeStr := storedCdr.FieldAsString(answerTimeFld)
	if primaryMandatory && len(aTimeStr) == 0 {
		return nil, errors.New(fmt.Sprintf("%s:%s:%s", ERR_MANDATORY_IE_MISSING, ANSWER_TIME, answerTimeFld.Id))
	} else if frkStorCdr.AnswerTime, err = ParseTimeDetectLayout(aTimeStr); err != nil {
		return nil, err
	}
	durStr := storedCdr.FieldAsString(durationFld)
	if primaryMandatory && len(durStr) == 0 {
		return nil, errors.New(fmt.Sprintf("%s:%s:%s", ERR_MANDATORY_IE_MISSING, USAGE, durationFld.Id))
	} else if frkStorCdr.Usage, err = ParseDurationWithSecs(durStr); err != nil {
		return nil, err
	}
	frkStorCdr.ExtraFields = make(map[string]string, len(extraFlds))
	for _, fld := range extraFlds {
		frkStorCdr.ExtraFields[fld.Id] = storedCdr.FieldAsString(fld)
	}
	return frkStorCdr, nil
}

func (storedCdr *StoredCdr) AsCgrCdrOut() *CgrCdrOut {
	return &CgrCdrOut{CgrId: storedCdr.CgrId,
		OrderId:        storedCdr.OrderId,
		TOR:            storedCdr.TOR,
		AccId:          storedCdr.AccId,
		CdrHost:        storedCdr.CdrHost,
		CdrSource:      storedCdr.CdrSource,
		ReqType:        storedCdr.ReqType,
		Direction:      storedCdr.Direction,
		Tenant:         storedCdr.Tenant,
		Category:       storedCdr.Category,
		Account:        storedCdr.Account,
		Subject:        storedCdr.Subject,
		Destination:    storedCdr.Destination,
		SetupTime:      storedCdr.SetupTime,
		AnswerTime:     storedCdr.AnswerTime,
		Usage:          storedCdr.Usage.Seconds(),
		ExtraFields:    storedCdr.ExtraFields,
		MediationRunId: storedCdr.MediationRunId,
		RatedAccount:   storedCdr.RatedAccount,
		RatedSubject:   storedCdr.RatedSubject,
		Cost:           storedCdr.Cost,
	}
}

// Implementation of Event interface, used in tests
func (storedCdr *StoredCdr) AsEvent(ignored string) Event {
	return Event(storedCdr)
}
func (storedCdr *StoredCdr) GetName() string {
	return storedCdr.CdrSource
}
func (storedCdr *StoredCdr) GetCgrId() string {
	return storedCdr.CgrId
}
func (storedCdr *StoredCdr) GetUUID() string {
	return storedCdr.AccId
}
func (storedCdr *StoredCdr) GetSessionIds() []string {
	return []string{storedCdr.GetUUID()}
}
func (storedCdr *StoredCdr) GetDirection(fieldName string) string {
	if IsSliceMember([]string{DIRECTION, META_DEFAULT}, fieldName) {
		return storedCdr.Direction
	}
	if strings.HasPrefix(fieldName, STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(STATIC_VALUE_PREFIX):]
	}
	return storedCdr.FieldAsString(&RSRField{Id: fieldName})
}
func (storedCdr *StoredCdr) GetSubject(fieldName string) string {
	if IsSliceMember([]string{SUBJECT, META_DEFAULT}, fieldName) {
		return storedCdr.Subject
	}
	if strings.HasPrefix(fieldName, STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(STATIC_VALUE_PREFIX):]
	}
	return storedCdr.FieldAsString(&RSRField{Id: fieldName})
}
func (storedCdr *StoredCdr) GetAccount(fieldName string) string {
	if IsSliceMember([]string{ACCOUNT, META_DEFAULT}, fieldName) {
		return storedCdr.Account
	}
	if strings.HasPrefix(fieldName, STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(STATIC_VALUE_PREFIX):]
	}
	return storedCdr.FieldAsString(&RSRField{Id: fieldName})
}
func (storedCdr *StoredCdr) GetDestination(fieldName string) string {
	if IsSliceMember([]string{DESTINATION, META_DEFAULT}, fieldName) {
		return storedCdr.Destination
	}
	if strings.HasPrefix(fieldName, STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(STATIC_VALUE_PREFIX):]
	}
	return storedCdr.FieldAsString(&RSRField{Id: fieldName})
}
func (storedCdr *StoredCdr) GetCallDestNr(fieldName string) string {
	if IsSliceMember([]string{DESTINATION, META_DEFAULT}, fieldName) {
		return storedCdr.Destination
	}
	if strings.HasPrefix(fieldName, STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(STATIC_VALUE_PREFIX):]
	}
	return storedCdr.FieldAsString(&RSRField{Id: fieldName})
}
func (storedCdr *StoredCdr) GetCategory(fieldName string) string {
	if IsSliceMember([]string{CATEGORY, META_DEFAULT}, fieldName) {
		return storedCdr.Category
	}
	if strings.HasPrefix(fieldName, STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(STATIC_VALUE_PREFIX):]
	}
	return storedCdr.FieldAsString(&RSRField{Id: fieldName})
}
func (storedCdr *StoredCdr) GetTenant(fieldName string) string {
	if IsSliceMember([]string{TENANT, META_DEFAULT}, fieldName) {
		return storedCdr.Tenant
	}
	if strings.HasPrefix(fieldName, STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(STATIC_VALUE_PREFIX):]
	}
	return storedCdr.FieldAsString(&RSRField{Id: fieldName})
}
func (storedCdr *StoredCdr) GetReqType(fieldName string) string {
	if IsSliceMember([]string{REQTYPE, META_DEFAULT}, fieldName) {
		return storedCdr.ReqType
	}
	if strings.HasPrefix(fieldName, STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(STATIC_VALUE_PREFIX):]
	}
	return storedCdr.FieldAsString(&RSRField{Id: fieldName})
}
func (storedCdr *StoredCdr) GetSetupTime(fieldName string) (time.Time, error) {
	if IsSliceMember([]string{SETUP_TIME, META_DEFAULT}, fieldName) {
		return storedCdr.SetupTime, nil
	}
	var sTimeVal string
	if strings.HasPrefix(fieldName, STATIC_VALUE_PREFIX) { // Static value
		sTimeVal = fieldName[len(STATIC_VALUE_PREFIX):]
	} else {
		sTimeVal = storedCdr.FieldAsString(&RSRField{Id: fieldName})
	}
	return ParseTimeDetectLayout(sTimeVal)
}
func (storedCdr *StoredCdr) GetAnswerTime(fieldName string) (time.Time, error) {
	if IsSliceMember([]string{ANSWER_TIME, META_DEFAULT}, fieldName) {
		return storedCdr.AnswerTime, nil
	}
	var aTimeVal string
	if strings.HasPrefix(fieldName, STATIC_VALUE_PREFIX) { // Static value
		aTimeVal = fieldName[len(STATIC_VALUE_PREFIX):]
	} else {
		aTimeVal = storedCdr.FieldAsString(&RSRField{Id: fieldName})
	}
	return ParseTimeDetectLayout(aTimeVal)
}
func (storedCdr *StoredCdr) GetEndTime() (time.Time, error) {
	return storedCdr.AnswerTime.Add(storedCdr.Usage), nil
}
func (storedCdr *StoredCdr) GetDuration(fieldName string) (time.Duration, error) {
	if IsSliceMember([]string{USAGE, META_DEFAULT}, fieldName) {
		return storedCdr.Usage, nil
	}
	var durVal string
	if strings.HasPrefix(fieldName, STATIC_VALUE_PREFIX) { // Static value
		durVal = fieldName[len(STATIC_VALUE_PREFIX):]
	} else {
		durVal = storedCdr.FieldAsString(&RSRField{Id: fieldName})
	}
	return ParseDurationWithSecs(durVal)
}
func (storedCdr *StoredCdr) GetOriginatorIP(fieldName string) string {
	if IsSliceMember([]string{CDRHOST, META_DEFAULT}, fieldName) {
		return storedCdr.CdrHost
	}
	return storedCdr.FieldAsString(&RSRField{Id: fieldName})
}
func (storedCdr *StoredCdr) GetExtraFields() map[string]string {
	return storedCdr.ExtraFields
}
func (storedCdr *StoredCdr) MissingParameter() bool {
	return len(storedCdr.AccId) == 0 ||
		len(storedCdr.Category) == 0 ||
		len(storedCdr.Tenant) == 0 ||
		len(storedCdr.Account) == 0 ||
		len(storedCdr.Destination) == 0
}
func (storedCdr *StoredCdr) ParseEventValue(rsrFld *RSRField) string {
	return storedCdr.FieldAsString(rsrFld)
}
func (storedCdr *StoredCdr) String() string {
	mrsh, _ := json.Marshal(storedCdr)
	return string(mrsh)
}

type CgrCdrOut struct {
	CgrId          string
	OrderId        int64
	TOR            string
	AccId          string
	CdrHost        string
	CdrSource      string
	ReqType        string
	Direction      string
	Tenant         string
	Category       string
	Account        string
	Subject        string
	Destination    string
	SetupTime      time.Time
	AnswerTime     time.Time
	Usage          float64
	ExtraFields    map[string]string
	MediationRunId string
	RatedAccount   string
	RatedSubject   string
	Cost           float64
}
