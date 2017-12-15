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

package sessionmanager

import (
	"encoding/json"
	"fmt"
	"math/rand"
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

func (ev SMGenericEvent) HasField(fieldName string) (hasField bool) {
	_, hasField = ev[fieldName]
	return
}

func (self SMGenericEvent) GetName() string {
	result, _ := utils.CastFieldIfToString(self[utils.EVENT_NAME])
	return result
}

func (self SMGenericEvent) GetTOR(fieldName string) string {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.TOR
	}
	result, _ := utils.CastFieldIfToString(self[fieldName])
	return result
}

func (self SMGenericEvent) GetCGRID(oIDFieldName string) string {
	return utils.Sha1(self.GetOriginID(oIDFieldName), self.GetOriginatorIP(utils.META_DEFAULT))
}

// GetOriginID returns the OriginID from event
// fieldName offers the possibility to extract info from other fields, eg: InitialOriginID
func (self SMGenericEvent) GetOriginID(fieldName string) string {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.ACCID
	}
	result, _ := utils.CastFieldIfToString(self[fieldName])
	return result
}

func (self SMGenericEvent) GetSessionIds() []string {
	return []string{self.GetOriginID(utils.META_DEFAULT)}
}

func (self SMGenericEvent) GetDirection(fieldName string) string {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.DIRECTION
	}
	result, _ := utils.CastFieldIfToString(self[fieldName])
	return result
}

func (self SMGenericEvent) GetAccount(fieldName string) string {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.Account
	}
	result, _ := utils.CastFieldIfToString(self[fieldName])
	return result
}

func (self SMGenericEvent) GetSubject(fieldName string) string {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.SUBJECT
	}
	result, _ := utils.CastFieldIfToString(self[fieldName])
	return result
}

func (self SMGenericEvent) GetDestination(fieldName string) string {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.Destination
	}
	result, _ := utils.CastFieldIfToString(self[fieldName])
	return result
}

func (self SMGenericEvent) GetCallDestNr(fieldName string) string {
	return self.GetDestination(fieldName)
}

func (self SMGenericEvent) GetCategory(fieldName string) string {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.Category
	}
	result, _ := utils.CastFieldIfToString(self[fieldName])
	return result
}

func (self SMGenericEvent) GetTenant(fieldName string) string {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.Tenant
	}
	result, _ := utils.CastFieldIfToString(self[fieldName])
	return result
}

func (self SMGenericEvent) GetReqType(fieldName string) string {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.RequestType
	}
	result, _ := utils.CastFieldIfToString(self[fieldName])
	return result
}

func (self SMGenericEvent) GetSetupTime(fieldName, timezone string) (time.Time, error) {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.SetupTime
	}
	result, _ := utils.CastFieldIfToString(self[fieldName])
	return utils.ParseTimeDetectLayout(result, timezone)
}

func (self SMGenericEvent) GetAnswerTime(fieldName, timezone string) (time.Time, error) {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.AnswerTime
	}
	result, _ := utils.CastFieldIfToString(self[fieldName])
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
		fieldName = utils.Usage
	}
	valIf, hasVal := self[fieldName]
	if !hasVal {
		return nilDuration, utils.ErrNotFound
	}
	result, _ := utils.CastFieldIfToString(valIf)
	return utils.ParseDurationWithNanosecs(result)
}

func (self SMGenericEvent) GetLastUsed(fieldName string) (time.Duration, error) {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.LastUsed
	}
	valStr, hasVal := self[fieldName]
	if !hasVal {
		return nilDuration, utils.ErrNotFound
	}
	result, _ := utils.CastFieldIfToString(valStr)
	return utils.ParseDurationWithNanosecs(result)
}

// GetSessionTTL retrieves SessionTTL setting out of SMGenericEvent
func (self SMGenericEvent) GetSessionTTL(sesTTL time.Duration,
	cfgSessionTTLMaxDelay *time.Duration) time.Duration {
	valIf, hasVal := self[utils.SessionTTL]
	if hasVal {
		ttlStr, converted := utils.CastFieldIfToString(valIf)
		if !converted {
			utils.Logger.Warning(
				fmt.Sprintf("SMGenericEvent, cannot convert SessionTTL, disabling functionality for event: <%s>",
					self.GetCGRID(utils.META_DEFAULT)))
			return time.Duration(0)
		}
		var err error
		if sesTTL, err = utils.ParseDurationWithNanosecs(ttlStr); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("SMGenericEvent, cannot parse SessionTTL, disabling functionality for event: <%s>",
					self.GetCGRID(utils.META_DEFAULT)))
			return time.Duration(0)
		}
	}
	// Variable sessionTTL
	var sessionTTLMaxDelay int64
	if cfgSessionTTLMaxDelay != nil {
		sessionTTLMaxDelay = cfgSessionTTLMaxDelay.Nanoseconds() / 1000000 // Milliseconds precision
	}
	if sesTTLMaxDelayIf, hasVal := self[utils.SessionTTLMaxDelay]; hasVal {
		maxTTLDelaxStr, converted := utils.CastFieldIfToString(sesTTLMaxDelayIf)
		if !converted {
			utils.Logger.Warning(fmt.Sprintf("SMGenericEvent, cannot convert SessionTTLMaxDelay, disabling functionality for event: <%s>",
				self.GetCGRID(utils.META_DEFAULT)))
			return time.Duration(0)
		}
		if maxTTLDelay, err := utils.ParseDurationWithNanosecs(maxTTLDelaxStr); err != nil {
			utils.Logger.Warning(fmt.Sprintf("SMGenericEvent, cannot parse SessionTTLMaxDelay, disabling functionality for event: <%s>",
				self.GetCGRID(utils.META_DEFAULT)))
			return time.Duration(0)
		} else {
			sessionTTLMaxDelay = maxTTLDelay.Nanoseconds() / 1000000
		}
	}
	if sessionTTLMaxDelay != 0 {
		rand.Seed(time.Now().Unix())
		sesTTL += time.Duration(rand.Int63n(sessionTTLMaxDelay) * 1000000)
	}
	return sesTTL
}

// GetSessionTTLLastUsed retrieves SessionTTLLastUsed setting out of SMGenericEvent
func (self SMGenericEvent) GetSessionTTLLastUsed() *time.Duration {
	valIf, hasVal := self[utils.SessionTTLLastUsed]
	if !hasVal {
		return nil
	}
	ttlStr, converted := utils.CastFieldIfToString(valIf)
	if !converted {
		return nil
	}
	if ttl, err := utils.ParseDurationWithNanosecs(ttlStr); err != nil {
		return nil
	} else {
		return &ttl
	}
}

// GetSessionTTLUsage retrieves SessionTTLUsage setting out of SMGenericEvent
func (self SMGenericEvent) GetSessionTTLUsage() *time.Duration {
	valIf, hasVal := self[utils.SessionTTLUsage]
	if !hasVal {
		return nil
	}
	ttlStr, converted := utils.CastFieldIfToString(valIf)
	if !converted {
		return nil
	}
	if ttl, err := utils.ParseDurationWithNanosecs(ttlStr); err != nil {
		return nil
	} else {
		return &ttl
	}
}

func (self SMGenericEvent) GetMaxUsage(fieldName string, cfgMaxUsage time.Duration) (time.Duration, error) {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.Usage
	}
	maxUsageStr, hasIt := self[fieldName]
	if !hasIt {
		return cfgMaxUsage, nil
	}
	result, _ := utils.CastFieldIfToString(maxUsageStr)
	return utils.ParseDurationWithNanosecs(result)
}

func (self SMGenericEvent) GetPdd(fieldName string) (time.Duration, error) {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.PDD
	}
	result, _ := utils.CastFieldIfToString(self[fieldName])
	return utils.ParseDurationWithNanosecs(result)
}

func (self SMGenericEvent) GetSupplier(fieldName string) string {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.SUPPLIER
	}
	result, _ := utils.CastFieldIfToString(self[fieldName])
	return result
}

func (self SMGenericEvent) GetDisconnectCause(fieldName string) string {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.DISCONNECT_CAUSE
	}
	result, _ := utils.CastFieldIfToString(self[fieldName])
	return result
}

func (self SMGenericEvent) GetOriginatorIP(fieldName string) string {
	if fieldName == utils.META_DEFAULT {
		fieldName = utils.OriginHost
	}
	result, _ := utils.CastFieldIfToString(self[fieldName])
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
		result, _ := utils.CastFieldIfToString(val)
		extraFields[key] = result
	}
	return extraFields
}

func (self SMGenericEvent) GetFieldAsString(fieldName string) (string, error) {
	valIf, hasVal := self[fieldName]
	if !hasVal {
		return "", utils.ErrNotFound
	}
	result, converted := utils.CastFieldIfToString(valIf)
	if !converted {
		return "", utils.ErrNotConvertible
	}
	return result, nil
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
		return rsrFld.ParseValue(self.GetCGRID(utils.META_DEFAULT))
	case utils.TOR:
		return rsrFld.ParseValue(utils.VOICE)
	case utils.ACCID:
		return rsrFld.ParseValue(self.GetOriginID(utils.META_DEFAULT))
	case utils.OriginHost:
		return rsrFld.ParseValue(self.GetOriginatorIP(utils.META_DEFAULT))
	case utils.Source:
		return rsrFld.ParseValue(self.GetName())
	case utils.RequestType:
		return rsrFld.ParseValue(self.GetReqType(utils.META_DEFAULT))
	case utils.DIRECTION:
		return rsrFld.ParseValue(self.GetDirection(utils.META_DEFAULT))
	case utils.Tenant:
		return rsrFld.ParseValue(self.GetTenant(utils.META_DEFAULT))
	case utils.Category:
		return rsrFld.ParseValue(self.GetCategory(utils.META_DEFAULT))
	case utils.Account:
		return rsrFld.ParseValue(self.GetAccount(utils.META_DEFAULT))
	case utils.SUBJECT:
		return rsrFld.ParseValue(self.GetSubject(utils.META_DEFAULT))
	case utils.Destination:
		return rsrFld.ParseValue(self.GetDestination(utils.META_DEFAULT))
	case utils.SetupTime:
		st, _ := self.GetSetupTime(utils.META_DEFAULT, timezone)
		return rsrFld.ParseValue(st.String())
	case utils.AnswerTime:
		at, _ := self.GetAnswerTime(utils.META_DEFAULT, timezone)
		return rsrFld.ParseValue(at.String())
	case utils.Usage:
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
		strVal, _ := utils.CastFieldIfToString(self[rsrFld.Id])
		val := rsrFld.ParseValue(strVal)
		return val
	}
	return ""
}

func (self SMGenericEvent) PassesFieldFilter(*utils.RSRField) (bool, string) {
	return true, ""
}

func (self SMGenericEvent) AsCDR(cfg *config.CGRConfig, timezone string) *engine.CDR {
	storCdr := engine.NewCDRWithDefaults(cfg)
	storCdr.CGRID = self.GetCGRID(utils.META_DEFAULT)
	storCdr.ToR = utils.FirstNonEmpty(self.GetTOR(utils.META_DEFAULT), storCdr.ToR) // Keep default if none in the event
	storCdr.OriginID = self.GetOriginID(utils.META_DEFAULT)
	storCdr.OriginHost = self.GetOriginatorIP(utils.META_DEFAULT)
	storCdr.Source = self.GetCdrSource()
	storCdr.RequestType = utils.FirstNonEmpty(self.GetReqType(utils.META_DEFAULT), storCdr.RequestType)
	storCdr.Tenant = utils.FirstNonEmpty(self.GetTenant(utils.META_DEFAULT), storCdr.Tenant)
	storCdr.Category = utils.FirstNonEmpty(self.GetCategory(utils.META_DEFAULT), storCdr.Category)
	storCdr.Account = self.GetAccount(utils.META_DEFAULT)
	storCdr.Subject = self.GetSubject(utils.META_DEFAULT)
	storCdr.Destination = self.GetDestination(utils.META_DEFAULT)
	storCdr.SetupTime, _ = self.GetSetupTime(utils.META_DEFAULT, timezone)
	storCdr.AnswerTime, _ = self.GetAnswerTime(utils.META_DEFAULT, timezone)
	storCdr.Usage, _ = self.GetUsage(utils.META_DEFAULT)
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
	setupTimeStr, _ := utils.CastFieldIfToString(self[utils.SetupTime])
	usageStr, _ := utils.CastFieldIfToString(self[utils.Usage])
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

// AsMapStringString Converts into map[string]string, used for example as pubsub event
func (self SMGenericEvent) AsMapStringString() (map[string]string, error) {
	mp := make(map[string]string, len(self))
	for k, v := range self {
		if strV, casts := utils.CastIfToString(v); !casts {
			return nil, fmt.Errorf("Value %+v does not cast to string", v)
		} else {
			mp[k] = strV
		}
	}
	return mp, nil
}

func (self SMGenericEvent) Clone() SMGenericEvent {
	evOut := make(SMGenericEvent, len(self))
	for key, val := range self {
		evOut[key] = val
	}
	return evOut
}
