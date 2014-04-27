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

package cdrs

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

const (
	// Freswitch event property names
	FS_CDR_MAP      = "variables"
	FS_DIRECTION    = "direction"
	FS_SUBJECT      = "cgr_subject"
	FS_ACCOUNT      = "cgr_account"
	FS_DESTINATION  = "cgr_destination"
	FS_REQTYPE      = "cgr_reqtype" //prepaid or postpaid
	FS_TOR          = "cgr_tor"
	FS_UUID         = "uuid" // -Unique ID for this call leg
	FS_CSTMID       = "cgr_tenant"
	FS_CALL_DEST_NR = "dialed_extension"
	FS_PARK_TIME    = "start_epoch"
	FS_SETUP_TIME   = "start_epoch"
	FS_ANSWER_TIME  = "answer_epoch"
	FS_HANGUP_TIME  = "end_epoch"
	FS_DURATION     = "billsec"
	FS_USERNAME     = "user_name"
	FS_IP           = "sip_local_network_addr"
	FS_CDR_SOURCE   = "freeswitch_json"
	FS_SIP_REQUSER  = "sip_req_user" // Apps like FusionPBX do not set dialed_extension, alternative being destination_number but that comes in customer profile, not in vars
)

type FSCdr struct {
	vars map[string]string
	body map[string]interface{} // keeps the loaded body for extra field search
}

func (fsCdr FSCdr) New(body []byte) (utils.RawCDR, error) {
	fsCdr.vars = make(map[string]string)
	var err error
	if err = json.Unmarshal(body, &fsCdr.body); err == nil {
		if variables, ok := fsCdr.body[FS_CDR_MAP]; ok {
			if variables, ok := variables.(map[string]interface{}); ok {
				for k, v := range variables {
					fsCdr.vars[k] = v.(string)
				}
			}
			return fsCdr, nil
		}
	}
	return nil, err
}

func (fsCdr FSCdr) GetCgrId() string {
	setupTime, _ := fsCdr.GetSetupTime()
	return utils.Sha1(fsCdr.vars[FS_UUID], setupTime.String())
}
func (fsCdr FSCdr) GetAccId() string {
	return fsCdr.vars[FS_UUID]
}
func (fsCdr FSCdr) GetCdrHost() string {
	return fsCdr.vars[FS_IP]
}
func (fsCdr FSCdr) GetCdrSource() string {
	return FS_CDR_SOURCE
}
func (fsCdr FSCdr) GetDirection() string {
	//TODO: implement direction, not related to FS_DIRECTION but traffic towards or from subject/account
	return "*out"
}
func (fsCdr FSCdr) GetSubject() string {
	return utils.FirstNonEmpty(fsCdr.vars[FS_SUBJECT], fsCdr.vars[FS_USERNAME])
}
func (fsCdr FSCdr) GetAccount() string {
	return utils.FirstNonEmpty(fsCdr.vars[FS_ACCOUNT], fsCdr.vars[FS_USERNAME])
}

// Charging destination number
func (fsCdr FSCdr) GetDestination() string {
	return utils.FirstNonEmpty(fsCdr.vars[FS_DESTINATION], fsCdr.vars[FS_CALL_DEST_NR], fsCdr.vars[FS_SIP_REQUSER])
}

func (fsCdr FSCdr) GetTOR() string {
	return utils.FirstNonEmpty(fsCdr.vars[FS_TOR], cfg.DefaultTOR)
}

func (fsCdr FSCdr) GetTenant() string {
	return utils.FirstNonEmpty(fsCdr.vars[FS_CSTMID], cfg.DefaultTenant)
}
func (fsCdr FSCdr) GetReqType() string {
	return utils.FirstNonEmpty(fsCdr.vars[FS_REQTYPE], cfg.DefaultReqType)
}
func (fsCdr FSCdr) GetExtraFields() map[string]string {
	extraFields := make(map[string]string, len(cfg.CDRSExtraFields))
	for _, field := range cfg.CDRSExtraFields {
		origFieldVal, foundInVars := fsCdr.vars[field.Id]
		if !foundInVars {
			origFieldVal = fsCdr.searchExtraField(field.Id, fsCdr.body)
		}
		extraFields[field.Id] = field.ParseValue(origFieldVal)
	}
	return extraFields
}

func (fsCdr FSCdr) searchExtraField(field string, body map[string]interface{}) (result string) {
	for key, value := range body {
		switch v := value.(type) {
		case string:
			if key == field {
				return v
			}
		case map[string]interface{}:
			if result = fsCdr.searchExtraField(field, v); result != "" {
				return
			}
		case []interface{}:
			for _, item := range v {
				if otherMap, ok := item.(map[string]interface{}); ok {
					if result = fsCdr.searchExtraField(field, otherMap); result != "" {
						return
					}
				} else {
					engine.Logger.Warning(fmt.Sprintf("Slice with no maps: %v", reflect.TypeOf(item)))
				}
			}
		default:
			engine.Logger.Warning(fmt.Sprintf("Unexpected type: %v", reflect.TypeOf(v)))
		}
	}
	return
}

func (fsCdr FSCdr) GetSetupTime() (t time.Time, err error) {
	return utils.ParseTimeDetectLayout(fsCdr.vars[FS_SETUP_TIME])
}
func (fsCdr FSCdr) GetAnswerTime() (t time.Time, err error) {
	return utils.ParseTimeDetectLayout(fsCdr.vars[FS_ANSWER_TIME])
}
func (fsCdr FSCdr) GetHangupTime() (t time.Time, err error) {
	return utils.ParseTimeDetectLayout(fsCdr.vars[FS_HANGUP_TIME])
}

// Extracts duration as considered by the telecom switch
func (fsCdr FSCdr) GetDuration() (time.Duration, error) {
	return utils.ParseDurationWithSecs(fsCdr.vars[FS_DURATION])
}

func (fsCdr FSCdr) Store() (result string, err error) {
	result += fsCdr.GetCgrId() + "|"
	result += fsCdr.GetAccId() + "|"
	result += fsCdr.GetCdrHost() + "|"
	result += fsCdr.GetDirection() + "|"
	result += fsCdr.GetSubject() + "|"
	result += fsCdr.GetAccount() + "|"
	result += fsCdr.GetDestination() + "|"
	result += fsCdr.GetTOR() + "|"
	result += fsCdr.GetAccId() + "|"
	result += fsCdr.GetTenant() + "|"
	result += fsCdr.GetReqType() + "|"
	st, err := fsCdr.GetAnswerTime()
	if err != nil {
		return "", err
	}
	result += strconv.FormatInt(st.UnixNano(), 10) + "|"
	et, err := fsCdr.GetHangupTime()
	if err != nil {
		return "", err
	}
	result += strconv.FormatInt(et.UnixNano(), 10) + "|"
	dur, _ := fsCdr.GetDuration()
	result += strconv.FormatInt(int64(dur.Seconds()), 10) + "|"
	return
}

func (fsCdr FSCdr) Restore(input string) error {
	return errors.New("Not implemented")
}

// Used in extra mediation
func (fsCdr FSCdr) AsStoredCdr(runId, reqTypeFld, directionFld, tenantFld, torFld, accountFld, subjectFld, destFld, setupTimeFld, answerTimeFld, durationFld string, extraFlds []string, fieldsMandatory bool) (*utils.StoredCdr, error) {
	if utils.IsSliceMember([]string{runId, reqTypeFld, directionFld, tenantFld, torFld, accountFld, subjectFld, destFld, answerTimeFld, durationFld}, "") {
		return nil, errors.New(fmt.Sprintf("%s:FieldName", utils.ERR_MANDATORY_IE_MISSING)) // All input field names are mandatory
	}
	var err error
	var hasKey bool
	var sTimeStr, aTimeStr, durStr string
	rtCdr := new(utils.StoredCdr)
	rtCdr.MediationRunId = runId
	rtCdr.Cost = -1.0 // Default for non-rated CDR
	if rtCdr.AccId = fsCdr.GetAccId(); len(rtCdr.AccId) == 0 {
		if fieldsMandatory {
			return nil, errors.New(fmt.Sprintf("%s:%s", utils.ERR_MANDATORY_IE_MISSING, utils.ACCID))
		} else { // Not mandatory, need to generate here CgrId
			rtCdr.CgrId = utils.GenUUID()
		}
	}
	if rtCdr.CdrHost = fsCdr.GetCdrHost(); len(rtCdr.CdrHost) == 0 && fieldsMandatory {
		return nil, errors.New(fmt.Sprintf("%s:%s", utils.ERR_MANDATORY_IE_MISSING, utils.CDRHOST))
	}
	if rtCdr.CdrSource = fsCdr.GetCdrSource(); len(rtCdr.CdrSource) == 0 && fieldsMandatory {
		return nil, errors.New(fmt.Sprintf("%s:%s", utils.ERR_MANDATORY_IE_MISSING, utils.CDRSOURCE))
	}
	if strings.HasPrefix(reqTypeFld, utils.STATIC_VALUE_PREFIX) { // Values starting with prefix are not dynamically populated
		rtCdr.ReqType = reqTypeFld[1:]
	} else if rtCdr.ReqType, hasKey = fsCdr.vars[reqTypeFld]; !hasKey && fieldsMandatory {
		return nil, errors.New(fmt.Sprintf("%s:%s", utils.ERR_MANDATORY_IE_MISSING, reqTypeFld))
	}
	if strings.HasPrefix(directionFld, utils.STATIC_VALUE_PREFIX) {
		rtCdr.Direction = directionFld[1:]
	} else if rtCdr.Direction, hasKey = fsCdr.vars[directionFld]; !hasKey && fieldsMandatory {
		return nil, errors.New(fmt.Sprintf("%s:%s", utils.ERR_MANDATORY_IE_MISSING, directionFld))
	}
	if strings.HasPrefix(tenantFld, utils.STATIC_VALUE_PREFIX) {
		rtCdr.Tenant = tenantFld[1:]
	} else if rtCdr.Tenant, hasKey = fsCdr.vars[tenantFld]; !hasKey && fieldsMandatory {
		return nil, errors.New(fmt.Sprintf("%s:%s", utils.ERR_MANDATORY_IE_MISSING, tenantFld))
	}
	if strings.HasPrefix(torFld, utils.STATIC_VALUE_PREFIX) {
		rtCdr.TOR = torFld[1:]
	} else if rtCdr.TOR, hasKey = fsCdr.vars[torFld]; !hasKey && fieldsMandatory {
		return nil, errors.New(fmt.Sprintf("%s:%s", utils.ERR_MANDATORY_IE_MISSING, torFld))
	}
	if strings.HasPrefix(accountFld, utils.STATIC_VALUE_PREFIX) {
		rtCdr.Account = accountFld[1:]
	} else if rtCdr.Account, hasKey = fsCdr.vars[accountFld]; !hasKey && fieldsMandatory {
		return nil, errors.New(fmt.Sprintf("%s:%s", utils.ERR_MANDATORY_IE_MISSING, accountFld))
	}
	if strings.HasPrefix(subjectFld, utils.STATIC_VALUE_PREFIX) {
		rtCdr.Subject = subjectFld[1:]
	} else if rtCdr.Subject, hasKey = fsCdr.vars[subjectFld]; !hasKey && fieldsMandatory {
		return nil, errors.New(fmt.Sprintf("%s:%s", utils.ERR_MANDATORY_IE_MISSING, subjectFld))
	}
	if strings.HasPrefix(destFld, utils.STATIC_VALUE_PREFIX) {
		rtCdr.Destination = destFld[1:]
	} else if rtCdr.Destination, hasKey = fsCdr.vars[destFld]; !hasKey && fieldsMandatory {
		return nil, errors.New(fmt.Sprintf("%s:%s", utils.ERR_MANDATORY_IE_MISSING, destFld))
	}
	if sTimeStr, hasKey = fsCdr.vars[setupTimeFld]; !hasKey && fieldsMandatory && !strings.HasPrefix(setupTimeFld, utils.STATIC_VALUE_PREFIX) {
		return nil, errors.New(fmt.Sprintf("%s:%s", utils.ERR_MANDATORY_IE_MISSING, setupTimeFld))
	} else {
		if strings.HasPrefix(setupTimeFld, utils.STATIC_VALUE_PREFIX) {
			sTimeStr = setupTimeFld[1:]
		}
		if rtCdr.SetupTime, err = utils.ParseTimeDetectLayout(sTimeStr); err != nil && fieldsMandatory {
			return nil, err
		}
	}
	if aTimeStr, hasKey = fsCdr.vars[answerTimeFld]; !hasKey && fieldsMandatory && !strings.HasPrefix(answerTimeFld, utils.STATIC_VALUE_PREFIX) {
		return nil, errors.New(fmt.Sprintf("%s:%s", utils.ERR_MANDATORY_IE_MISSING, answerTimeFld))
	} else {
		if strings.HasPrefix(answerTimeFld, utils.STATIC_VALUE_PREFIX) {
			aTimeStr = answerTimeFld[1:]
		}
		if rtCdr.AnswerTime, err = utils.ParseTimeDetectLayout(aTimeStr); err != nil && fieldsMandatory {
			return nil, err
		}
	}
	if durStr, hasKey = fsCdr.vars[durationFld]; !hasKey && fieldsMandatory && !strings.HasPrefix(durationFld, utils.STATIC_VALUE_PREFIX) {
		return nil, errors.New(fmt.Sprintf("%s:%s", utils.ERR_MANDATORY_IE_MISSING, durationFld))
	} else {
		if strings.HasPrefix(durationFld, utils.STATIC_VALUE_PREFIX) {
			durStr = durationFld[1:]
		}
		if rtCdr.Duration, err = utils.ParseDurationWithSecs(durStr); err != nil && fieldsMandatory {
			return nil, err
		}
	}
	rtCdr.CgrId = utils.Sha1(rtCdr.AccId, rtCdr.SetupTime.String())
	rtCdr.ExtraFields = make(map[string]string, len(extraFlds))
	for _, fldName := range extraFlds {
		if fldVal, hasKey := fsCdr.vars[fldName]; !hasKey && fieldsMandatory {
			return nil, errors.New(fmt.Sprintf("%s:%s", utils.ERR_MANDATORY_IE_MISSING, fldName))
		} else {
			rtCdr.ExtraFields[fldName] = fldVal
		}
	}
	return rtCdr, nil
}
