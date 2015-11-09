/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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

package sessionmanager

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	nilTime     time.Time
	nilDuration time.Duration
)

type GenericEvent map[string]interface{}

func (self GenericEvent) GetName() string {
	result, _ := utils.ConvertIfaceToString(self[utils.EVENT_NAME])
	return result
}

func (self GenericEvent) GetCgrId(timezone string) string {
	setupTime, _ := self.GetSetupTime(utils.META_DEFAULT, timezone)
	return utils.Sha1(self.GetUUID(), setupTime.UTC().String())
}

func (self GenericEvent) GetUUID() string {
	result, _ := utils.ConvertIfaceToString(self[utils.ACCID])
	return result
}

func (self GenericEvent) GetSessionIds() []string {
	return []string{self.GetUUID()}
}

func (self GenericEvent) GetDirection(fieldName string) string {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.DIRECTION
	}
	result, _ := utils.ConvertIfaceToString(self[fieldName])
	return result
}

func (self GenericEvent) GetAccount(fieldName string) string {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.ACCOUNT
	}
	result, _ := utils.ConvertIfaceToString(self[fieldName])
	return result
}

func (self GenericEvent) GetSubject(fieldName string) string {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.SUBJECT
	}
	result, _ := utils.ConvertIfaceToString(self[fieldName])
	return result
}

func (self GenericEvent) GetDestination(fieldName string) string {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.DESTINATION
	}
	result, _ := utils.ConvertIfaceToString(self[fieldName])
	return result
}

func (self GenericEvent) GetCallDestNr(fieldName string) string {
	return self.GetDestination(fieldName)
}

func (self GenericEvent) GetCategory(fieldName string) string {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.CATEGORY
	}
	result, _ := utils.ConvertIfaceToString(self[fieldName])
	return result
}

func (self GenericEvent) GetTenant(fieldName string) string {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.TENANT
	}
	result, _ := utils.ConvertIfaceToString(self[fieldName])
	return result
}

func (self GenericEvent) GetReqType(fieldName string) string {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.REQTYPE
	}
	result, _ := utils.ConvertIfaceToString(self[fieldName])
	return result
}

func (self GenericEvent) GetSetupTime(fieldName, timezone string) (time.Time, error) {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.SETUP_TIME
	}
	result, _ := utils.ConvertIfaceToString(self[fieldName])
	return utils.ParseTimeDetectLayout(result, timezone)
}

func (self GenericEvent) GetAnswerTime(fieldName, timezone string) (time.Time, error) {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.ANSWER_TIME
	}
	result, _ := utils.ConvertIfaceToString(self[fieldName])
	return utils.ParseTimeDetectLayout(result, timezone)
}

func (self GenericEvent) GetEndTime(fieldName, timezone string) (time.Time, error) {
	var nilTime time.Time
	aTime, err := self.GetAnswerTime(utils.META_DEFAULT, timezone)
	if err != nil {
		return nilTime, err
	}
	dur, err := self.GetDuration(utils.META_DEFAULT)
	if err != nil {
		return nilTime, err
	}
	return aTime.Add(dur), nil
}

func (self GenericEvent) GetDuration(fieldName string) (time.Duration, error) {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.USAGE
	}
	result, _ := utils.ConvertIfaceToString(self[fieldName])
	return utils.ParseDurationWithSecs(result)
}

func (self GenericEvent) GetPdd(fieldName string) (time.Duration, error) {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.PDD
	}
	result, _ := utils.ConvertIfaceToString(self[fieldName])
	return utils.ParseDurationWithSecs(result)
}

func (self GenericEvent) GetSupplier(fieldName string) string {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.SUPPLIER
	}
	result, _ := utils.ConvertIfaceToString(self[fieldName])
	return result
}

func (self GenericEvent) GetDisconnectCause(fieldName string) string {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.DISCONNECT_CAUSE
	}
	result, _ := utils.ConvertIfaceToString(self[fieldName])
	return result
}

func (self GenericEvent) GetOriginatorIP(fieldName string) string {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.CDRSOURCE
	}
	result, _ := utils.ConvertIfaceToString(self[fieldName])
	return result
}

func (self GenericEvent) GetExtraFields() map[string]string {
	extraFields := make(map[string]string)
	for key, val := range self {
		if utils.IsSliceMember(utils.PrimaryCdrFields, key) {
			continue
		}
		result, _ := utils.ConvertIfaceToString(val)
		extraFields[key] = result
	}
	return extraFields
}

func (self GenericEvent) MissingParameter(timezone string) bool {
	switch self.GetName() {
	case utils.CGR_AUTHORIZATION:
		if setupTime, err := self.GetSetupTime(utils.META_DEFAULT, timezone); err != nil || setupTime == nilTime {
			return true
		}
		return len(self.GetAccount(utils.META_DEFAULT)) == 0 ||
			len(self.GetDestination(utils.META_DEFAULT)) == 0
	}
	return false
}

func (self GenericEvent) ParseEventValue(rsrFld *utils.RSRField, timezone string) string {
	switch rsrFld.Id {
	case utils.CGRID:
		return rsrFld.ParseValue(self.GetCgrId(timezone))
	case utils.TOR:
		return rsrFld.ParseValue(utils.VOICE)
	case utils.ACCID:
		return rsrFld.ParseValue(self.GetUUID())
	case utils.CDRHOST:
		return rsrFld.ParseValue(self.GetOriginatorIP(utils.META_DEFAULT))
	case utils.CDRSOURCE:
		return rsrFld.ParseValue(self.GetName())
	case utils.REQTYPE:
		return rsrFld.ParseValue(self.GetReqType(utils.META_DEFAULT))
	case utils.DIRECTION:
		return rsrFld.ParseValue(self.GetDirection(utils.META_DEFAULT))
	case utils.TENANT:
		return rsrFld.ParseValue(self.GetTenant(utils.META_DEFAULT))
	case utils.CATEGORY:
		return rsrFld.ParseValue(self.GetCategory(utils.META_DEFAULT))
	case utils.ACCOUNT:
		return rsrFld.ParseValue(self.GetAccount(utils.META_DEFAULT))
	case utils.SUBJECT:
		return rsrFld.ParseValue(self.GetSubject(utils.META_DEFAULT))
	case utils.DESTINATION:
		return rsrFld.ParseValue(self.GetDestination(utils.META_DEFAULT))
	case utils.SETUP_TIME:
		st, _ := self.GetSetupTime(utils.META_DEFAULT, timezone)
		return rsrFld.ParseValue(st.String())
	case utils.ANSWER_TIME:
		at, _ := self.GetAnswerTime(utils.META_DEFAULT, timezone)
		return rsrFld.ParseValue(at.String())
	case utils.USAGE:
		dur, _ := self.GetDuration(utils.META_DEFAULT)
		return rsrFld.ParseValue(strconv.FormatInt(dur.Nanoseconds(), 10))
	case utils.PDD:
		pdd, _ := self.GetPdd(utils.META_DEFAULT)
		return rsrFld.ParseValue(strconv.FormatFloat(pdd.Seconds(), 'f', -1, 64))
	case utils.SUPPLIER:
		return rsrFld.ParseValue(self.GetSupplier(utils.META_DEFAULT))
	case utils.DISCONNECT_CAUSE:
		return rsrFld.ParseValue(self.GetDisconnectCause(utils.META_DEFAULT))
	case utils.MEDI_RUNID:
		return rsrFld.ParseValue(utils.META_DEFAULT)
	case utils.COST:
		return rsrFld.ParseValue(strconv.FormatFloat(-1, 'f', -1, 64)) // Recommended to use FormatCost
	default:
		strVal, _ := utils.ConvertIfaceToString(self[rsrFld.Id])
		val := rsrFld.ParseValue(strVal)
		return val
	}
	return ""
}

func (self GenericEvent) PassesFieldFilter(*utils.RSRField) (bool, string) {
	return true, ""
}

func (self GenericEvent) AsStoredCdr(timezone string) *engine.StoredCdr {
	return nil
}

func (self GenericEvent) String() string {
	jsn, _ := json.Marshal(self)
	return string(jsn)
}

func (self GenericEvent) AsEvent(timezone string) engine.Event {
	return self
}

func (self GenericEvent) ComputeLcr() bool {
	computeLcr, _ := self[utils.COMPUTE_LCR].(bool)
	return computeLcr
}
