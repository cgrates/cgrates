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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	nilTime     time.Time
	nilDuration time.Duration
)

type SMGenericEvent map[string]interface{}

func (self SMGenericEvent) GetName() string {
	result, _ := utils.ConvertIfaceToString(self[utils.EVENT_NAME])
	return result
}

func (self SMGenericEvent) GetTOR(fieldName string) string {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.TOR
	}
	result, _ := utils.ConvertIfaceToString(self[fieldName])
	return result
}

func (self SMGenericEvent) GetCgrId(timezone string) string {
	setupTime, _ := self.GetSetupTime(utils.META_DEFAULT, timezone)
	return utils.Sha1(self.GetUUID(), setupTime.UTC().String())
}

func (self SMGenericEvent) GetUUID() string {
	result, _ := utils.ConvertIfaceToString(self[utils.ACCID])
	return result
}

func (self SMGenericEvent) GetSessionIds() []string {
	return []string{self.GetUUID()}
}

func (self SMGenericEvent) GetDirection(fieldName string) string {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.DIRECTION
	}
	result, _ := utils.ConvertIfaceToString(self[fieldName])
	return result
}

func (self SMGenericEvent) GetAccount(fieldName string) string {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.ACCOUNT
	}
	result, _ := utils.ConvertIfaceToString(self[fieldName])
	return result
}

func (self SMGenericEvent) GetSubject(fieldName string) string {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.SUBJECT
	}
	result, _ := utils.ConvertIfaceToString(self[fieldName])
	return result
}

func (self SMGenericEvent) GetDestination(fieldName string) string {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.DESTINATION
	}
	result, _ := utils.ConvertIfaceToString(self[fieldName])
	return result
}

func (self SMGenericEvent) GetCallDestNr(fieldName string) string {
	return self.GetDestination(fieldName)
}

func (self SMGenericEvent) GetCategory(fieldName string) string {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.CATEGORY
	}
	result, _ := utils.ConvertIfaceToString(self[fieldName])
	return result
}

func (self SMGenericEvent) GetTenant(fieldName string) string {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.TENANT
	}
	result, _ := utils.ConvertIfaceToString(self[fieldName])
	return result
}

func (self SMGenericEvent) GetReqType(fieldName string) string {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.REQTYPE
	}
	result, _ := utils.ConvertIfaceToString(self[fieldName])
	return result
}

func (self SMGenericEvent) GetSetupTime(fieldName, timezone string) (time.Time, error) {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.SETUP_TIME
	}
	result, _ := utils.ConvertIfaceToString(self[fieldName])
	return utils.ParseTimeDetectLayout(result, timezone)
}

func (self SMGenericEvent) GetAnswerTime(fieldName, timezone string) (time.Time, error) {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.ANSWER_TIME
	}
	result, _ := utils.ConvertIfaceToString(self[fieldName])
	return utils.ParseTimeDetectLayout(result, timezone)
}

func (self SMGenericEvent) GetEndTime(fieldName, timezone string) (time.Time, error) {
	var nilTime time.Time
	aTime, err := self.GetAnswerTime(utils.META_DEFAULT, timezone)
	if err != nil {
		return nilTime, err
	}
	dur, err := self.GetUsage(utils.META_DEFAULT)
	if err != nil {
		return nilTime, err
	}
	return aTime.Add(dur), nil
}

func (self SMGenericEvent) GetUsage(fieldName string) (time.Duration, error) {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.USAGE
	}
	result, _ := utils.ConvertIfaceToString(self[fieldName])
	return utils.ParseDurationWithSecs(result)
}

func (self SMGenericEvent) GetMaxUsage(fieldName string, cfgMaxUsage time.Duration) (time.Duration, error) {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.USAGE
	}
	maxUsageStr, hasIt := self[fieldName]
	if !hasIt {
		return cfgMaxUsage, nil
	}
	result, _ := utils.ConvertIfaceToString(maxUsageStr)
	return utils.ParseDurationWithSecs(result)
}

func (self SMGenericEvent) GetPdd(fieldName string) (time.Duration, error) {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.PDD
	}
	result, _ := utils.ConvertIfaceToString(self[fieldName])
	return utils.ParseDurationWithSecs(result)
}

func (self SMGenericEvent) GetSupplier(fieldName string) string {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.SUPPLIER
	}
	result, _ := utils.ConvertIfaceToString(self[fieldName])
	return result
}

func (self SMGenericEvent) GetDisconnectCause(fieldName string) string {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.DISCONNECT_CAUSE
	}
	result, _ := utils.ConvertIfaceToString(self[fieldName])
	return result
}

func (self SMGenericEvent) GetOriginatorIP(fieldName string) string {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.CDRHOST
	}
	result, _ := utils.ConvertIfaceToString(self[fieldName])
	return result
}

func (self SMGenericEvent) GetCdrSource() string {
	return utils.SMG + "_" + self.GetName()
}

func (self SMGenericEvent) GetExtraFields() map[string]string {
	extraFields := make(map[string]string)
	for key, val := range self {
		primaryFields := append(utils.PrimaryCdrFields, utils.EVENT_NAME)
		if utils.IsSliceMember(primaryFields, key) {
			continue
		}
		result, _ := utils.ConvertIfaceToString(val)
		extraFields[key] = result
	}
	return extraFields
}

func (self SMGenericEvent) MissingParameter(timezone string) bool {
	switch self.GetName() {
	case utils.CGR_AUTHORIZATION:
		if setupTime, err := self.GetSetupTime(utils.META_DEFAULT, timezone); err != nil || setupTime == nilTime {
			return true
		}
		return len(self.GetAccount(utils.META_DEFAULT)) == 0 ||
			len(self.GetDestination(utils.META_DEFAULT)) == 0

	case utils.CGR_SESSION_START:
		return false
	case utils.CGR_SESSION_UPDATE:
		return false
	case utils.CGR_SESSION_END:
		return false
	case utils.CGR_LCR_REQUEST:
		return false
	}
	return true // Unhandled event
}

func (self SMGenericEvent) ParseEventValue(rsrFld *utils.RSRField, timezone string) string {
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
		dur, _ := self.GetUsage(utils.META_DEFAULT)
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

func (self SMGenericEvent) PassesFieldFilter(*utils.RSRField) (bool, string) {
	return true, ""
}

func (self SMGenericEvent) AsStoredCdr(cfg *config.CGRConfig, timezone string) *engine.CDR {
	storCdr := engine.NewCDRWithDefaults(cfg)
	storCdr.CGRID = self.GetCgrId(timezone)
	storCdr.ToR = utils.FirstNonEmpty(self.GetTOR(utils.META_DEFAULT), storCdr.ToR) // Keep default if none in the event
	storCdr.OriginID = self.GetUUID()
	storCdr.OriginHost = self.GetOriginatorIP(utils.META_DEFAULT)
	storCdr.Source = self.GetCdrSource()
	storCdr.RequestType = utils.FirstNonEmpty(self.GetReqType(utils.META_DEFAULT), storCdr.RequestType)
	storCdr.Direction = utils.FirstNonEmpty(self.GetDirection(utils.META_DEFAULT), storCdr.Direction)
	storCdr.Tenant = utils.FirstNonEmpty(self.GetTenant(utils.META_DEFAULT), storCdr.Tenant)
	storCdr.Category = utils.FirstNonEmpty(self.GetCategory(utils.META_DEFAULT), storCdr.Category)
	storCdr.Account = self.GetAccount(utils.META_DEFAULT)
	storCdr.Subject = self.GetSubject(utils.META_DEFAULT)
	storCdr.Destination = self.GetDestination(utils.META_DEFAULT)
	storCdr.SetupTime, _ = self.GetSetupTime(utils.META_DEFAULT, timezone)
	storCdr.AnswerTime, _ = self.GetAnswerTime(utils.META_DEFAULT, timezone)
	storCdr.Usage, _ = self.GetUsage(utils.META_DEFAULT)
	storCdr.PDD, _ = self.GetPdd(utils.META_DEFAULT)
	storCdr.Supplier = self.GetSupplier(utils.META_DEFAULT)
	storCdr.DisconnectCause = self.GetDisconnectCause(utils.META_DEFAULT)
	storCdr.ExtraFields = self.GetExtraFields()
	storCdr.Cost = -1
	return storCdr
}

func (self SMGenericEvent) String() string {
	jsn, _ := json.Marshal(self)
	return string(jsn)
}

func (self SMGenericEvent) ComputeLcr() bool {
	computeLcr, _ := self[utils.COMPUTE_LCR].(bool)
	return computeLcr
}

func (self SMGenericEvent) AsLcrRequest() *engine.LcrRequest {
	setupTimeStr, _ := utils.ConvertIfaceToString(self[utils.SETUP_TIME])
	usageStr, _ := utils.ConvertIfaceToString(self[utils.USAGE])
	return &engine.LcrRequest{
		Direction:   self.GetDirection(utils.META_DEFAULT),
		Tenant:      self.GetTenant(utils.META_DEFAULT),
		Category:    self.GetCategory(utils.META_DEFAULT),
		Account:     self.GetAccount(utils.META_DEFAULT),
		Subject:     self.GetSubject(utils.META_DEFAULT),
		Destination: self.GetDestination(utils.META_DEFAULT),
		SetupTime:   utils.FirstNonEmpty(setupTimeStr),
		Duration:    usageStr,
	}
}
