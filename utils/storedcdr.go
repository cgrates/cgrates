/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2014 ITsysCOM GmbH

This program is free software: you can Storagetribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
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
	"errors"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"time"
)

// Kinda standard of internal CDR, complies to CDR interface also
type StoredCdr struct {
	CgrId          string
	OrderId        int64 // Stor order id used as export order id
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
	Usage          time.Duration
	ExtraFields    map[string]string
	MediationRunId string
	Cost           float64
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
	case HOURS:
		return strconv.FormatFloat(Round(storedCdr.Usage.Seconds(), 0, ROUNDING_MIDDLE), 'f', -1, 64)
	case MINUTES:
		return strconv.FormatFloat(Round(storedCdr.Usage.Seconds(), 0, ROUNDING_MIDDLE), 'f', -1, 64)
	case SECONDS:
		return strconv.FormatFloat(Round(storedCdr.Usage.Seconds(), 0, ROUNDING_MIDDLE), 'f', -1, 64)
	default:
		return strconv.FormatInt(storedCdr.Usage.Nanoseconds(), 10)
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
		//if IsSliceMember([]string{DATA, SMS}, storedCdr.TOR) {
		//	return strconv.FormatFloat(Round(storedCdr.Usage.Seconds(), 0, ROUNDING_MIDDLE), 'f', -1, 64)
		//}
		return rsrFld.ParseValue(strconv.FormatInt(storedCdr.Usage.Nanoseconds(), 10))
	case MEDI_RUNID:
		return rsrFld.ParseValue(storedCdr.MediationRunId)
	case COST:
		return rsrFld.ParseValue(strconv.FormatFloat(storedCdr.Cost, 'f', -1, 64)) // Recommended to use FormatCost
	default:
		return rsrFld.ParseValue(storedCdr.ExtraFields[rsrFld.Id])
	}
}

func (storedCdr *StoredCdr) AsStoredCdr() *StoredCdr {
	return storedCdr
}

// Ability to send the CgrCdr remotely to another CDR server
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
	v.Set(USAGE, strconv.FormatInt(storedCdr.Usage.Nanoseconds(), 10))
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
	} else if frkStorCdr.Usage, err = ParseDurationWithNanosecs(durStr); err != nil {
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
		Cost:           storedCdr.Cost,
	}
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
	Cost           float64
}
