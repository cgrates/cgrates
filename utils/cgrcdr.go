/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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

package utils

import (
	"net/http"
	"time"
	"errors"
	"fmt"
)

func NewCgrCdrFromHttpReq(req *http.Request) (CgrCdr, error) {
	if req.Form == nil {
		if err := req.ParseForm(); err != nil {
			return nil, err
		}
	}
	cgrCdr := make(CgrCdr)
	cgrCdr[CDRHOST] = req.RemoteAddr
	for k, vals := range req.Form {
		cgrCdr[k] = vals[0] // We only support the first value for now, if more are provided it is considered remote's fault
	}
	return cgrCdr, nil
}

type CgrCdr map[string]string

func (cgrCdr CgrCdr) GetCgrId() string {
	return FSCgrId(cgrCdr[ACCID])
}

func (cgrCdr CgrCdr) GetAccId() string {
	return cgrCdr[ACCID]
}

func (cgrCdr CgrCdr) GetCdrHost() string {
	return cgrCdr[CDRHOST]
}

func (cgrCdr CgrCdr) GetCdrSource() string {
	return cgrCdr[CDRSOURCE]
}
func (cgrCdr CgrCdr) GetDirection() string {
	//TODO: implement direction
	return "*out"
}
func (cgrCdr CgrCdr) GetOrigId() string {
	return cgrCdr[CDRHOST]
}
func (cgrCdr CgrCdr) GetSubject() string {
	return cgrCdr[SUBJECT]
}
func (cgrCdr CgrCdr) GetAccount() string {
	return cgrCdr[ACCOUNT]
}

// Charging destination number
func (cgrCdr CgrCdr) GetDestination() string {
	return cgrCdr[DESTINATION]
}

func (cgrCdr CgrCdr) GetTOR() string {
	return cgrCdr[TOR]
}

func (cgrCdr CgrCdr) GetTenant() string {
	return cgrCdr[TENANT]
}
func (cgrCdr CgrCdr) GetReqType() string {
	return cgrCdr[REQTYPE]
}
func (cgrCdr CgrCdr) GetExtraFields() map[string]string {
	extraFields := make(map[string]string)
	for k, v := range cgrCdr {
		if !IsSliceMember(PrimaryCdrFields, k) {
			extraFields[k] = v
		}
	}
	return extraFields
}
func (cgrCdr CgrCdr) GetAnswerTime() (t time.Time, err error) {
	return ParseTimeDetectLayout(cgrCdr[ANSWER_TIME])
}

// Extracts duration as considered by the telecom switch
func (cgrCdr CgrCdr) GetDuration() time.Duration {
	dur, _ := ParseDurationWithSecs(cgrCdr[DURATION])
	return dur
}

// Used in mediation, fieldsMandatory marks whether missing field out of request represents error or can be ignored
func(cgrCdr CgrCdr) AsRatedCdr(runId, reqTypeFld, directionFld, tenantFld, torFld, accountFld, subjectFld, destFld, answerTimeFld, durationFld string, extraFlds []string, fieldsMandatory bool) (*RatedCDR, error) {
	if IsSliceMember([]string{runId, reqTypeFld, directionFld, tenantFld, torFld, accountFld, subjectFld, destFld, answerTimeFld, durationFld}, "") {
		return nil, errors.New(fmt.Sprintf("%s:FieldName", ERR_MANDATORY_IE_MISSING)) // All input field names are mandatory
		}
	var err error
	var hasKey bool
	var aTimeStr, durStr string
	rtCdr := new(RatedCDR)
	rtCdr.MediationRunId = runId
	rtCdr.Cost = -1.0 // Default for non-rated CDR
	if rtCdr.AccId, hasKey = cgrCdr[ACCID]; !hasKey {
		if fieldsMandatory {
			return nil, errors.New(fmt.Sprintf("%s:%s", ERR_MANDATORY_IE_MISSING, ACCID))
		} else { // Not mandatory, cgrid needs however to be unique
			rtCdr.CgrId = GenUUID()
		}
	} else { // hasKey, use it to generate cgrid
		rtCdr.CgrId = FSCgrId(rtCdr.AccId)
	}
	if rtCdr.CdrHost, hasKey = cgrCdr[CDRHOST]; !hasKey && fieldsMandatory {
		return nil, errors.New(fmt.Sprintf("%s:%s", ERR_MANDATORY_IE_MISSING, CDRHOST))
	}
	if rtCdr.CdrSource, hasKey = cgrCdr[CDRSOURCE]; !hasKey && fieldsMandatory {
		return nil, errors.New(fmt.Sprintf("%s:%s", ERR_MANDATORY_IE_MISSING, CDRSOURCE))
	}
	if rtCdr.ReqType, hasKey = cgrCdr[reqTypeFld]; !hasKey && fieldsMandatory {
		return nil, errors.New(fmt.Sprintf("%s:%s", ERR_MANDATORY_IE_MISSING, reqTypeFld))
	}
	if rtCdr.Direction, hasKey = cgrCdr[directionFld]; !hasKey && fieldsMandatory {
		return nil, errors.New(fmt.Sprintf("%s:%s", ERR_MANDATORY_IE_MISSING, directionFld))
	}
	if rtCdr.Tenant, hasKey = cgrCdr[tenantFld]; !hasKey && fieldsMandatory {
		return nil, errors.New(fmt.Sprintf("%s:%s", ERR_MANDATORY_IE_MISSING, tenantFld))
	}
	if rtCdr.TOR, hasKey = cgrCdr[torFld]; !hasKey && fieldsMandatory {
		return nil, errors.New(fmt.Sprintf("%s:%s", ERR_MANDATORY_IE_MISSING, torFld))
	}
	if rtCdr.Account, hasKey = cgrCdr[accountFld]; !hasKey && fieldsMandatory {
		return nil, errors.New(fmt.Sprintf("%s:%s", ERR_MANDATORY_IE_MISSING, accountFld))
	}
	if rtCdr.Subject, hasKey = cgrCdr[subjectFld]; !hasKey && fieldsMandatory {
		return nil, errors.New(fmt.Sprintf("%s:%s", ERR_MANDATORY_IE_MISSING, subjectFld))
	}
	if rtCdr.Destination, hasKey = cgrCdr[destFld]; !hasKey && fieldsMandatory {
		return nil, errors.New(fmt.Sprintf("%s:%s", ERR_MANDATORY_IE_MISSING, destFld))
	}
	if aTimeStr, hasKey = cgrCdr[answerTimeFld]; !hasKey && fieldsMandatory {
		return nil, errors.New(fmt.Sprintf("%s:%s", ERR_MANDATORY_IE_MISSING, answerTimeFld))
	} else {
		if rtCdr.AnswerTime, err = ParseTimeDetectLayout(aTimeStr); err != nil && fieldsMandatory {
			return nil, err
		}
	}
	if durStr, hasKey = cgrCdr[durationFld]; !hasKey && fieldsMandatory { 
		return nil, errors.New(fmt.Sprintf("%s:%s", ERR_MANDATORY_IE_MISSING, durationFld))
	} else {
		if rtCdr.Duration, err = ParseDurationWithSecs(durStr); err != nil && fieldsMandatory {
			return nil, err
		}
	}
	rtCdr.ExtraFields = make(map[string]string, len(extraFlds))
	for _, fldName := range extraFlds {
		if fldVal, hasKey := cgrCdr[fldName]; !hasKey && fieldsMandatory {
			return nil, errors.New(fmt.Sprintf("%s:%s", ERR_MANDATORY_IE_MISSING, fldName))
		} else {
			rtCdr.ExtraFields[fldName] = fldVal
		}
	}
	return rtCdr, nil
}
