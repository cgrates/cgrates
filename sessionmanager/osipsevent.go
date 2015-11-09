/*
Real-time Charging System for Telecom & ISP environments
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
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/osipsdagram"
)

/*
/*&{Name:E_ACC_CDR AttrValues:map[to_tag:5ec6e925 cgr_account:dan setuptime:1 created:1406312794 method:INVITE callid:Y2I5ZDYzMDkzM2YzYjhlZjA2Y2ZhZTJmZTc4MGU4NDI.
	// sip_reason:OK time:1406312795 cgr_reqtype:prepaid cgr_destination:dan cgr_subject:dan sip_code:200 duration:7 from_tag:a5716471] Values:[]}*/

const (
	FROM_TAG                 = "from_tag"
	TO_TAG                   = "to_tag"
	CALLID                   = "callid"
	CGR_CATEGORY             = "cgr_category"
	CGR_REQTYPE              = "cgr_reqtype"
	CGR_TENANT               = "cgr_tenant"
	CGR_SUBJECT              = "cgr_subject"
	CGR_ACCOUNT              = "cgr_account"
	CGR_DESTINATION          = "cgr_destination"
	TIME                     = "time"
	SETUP_DURATION           = "setuptime"
	OSIPS_SETUP_TIME         = "created"
	OSIPS_EVENT_TIME         = "time"
	OSIPS_DURATION           = "duration"
	OSIPS_AUTH_OK            = "AUTH_OK"
	OSIPS_INSUFFICIENT_FUNDS = "INSUFFICIENT_FUNDS"
	OSIPS_DIALOG_ID          = "dialog_id"
	OSIPS_SIPCODE            = "sip_code"
)

func NewOsipsEvent(osipsDagramEvent *osipsdagram.OsipsEvent) (*OsipsEvent, error) {
	return &OsipsEvent{osipsEvent: osipsDagramEvent}, nil
}

type OsipsEvent struct {
	osipsEvent *osipsdagram.OsipsEvent
}

func (osipsev *OsipsEvent) AsEvent(evStr string) engine.Event {
	return osipsev
}

func (osipsev *OsipsEvent) String() string {
	mrsh, _ := json.Marshal(osipsev)
	return string(mrsh)
}

func (osipsev *OsipsEvent) GetName() string {
	return osipsev.osipsEvent.Name
}

func (osipsev *OsipsEvent) GetCgrId(timezone string) string {
	setupTime, _ := osipsev.GetSetupTime(utils.META_DEFAULT, timezone)
	return utils.Sha1(osipsev.GetUUID(), setupTime.UTC().String())
}

func (osipsev *OsipsEvent) GetUUID() string {
	return osipsev.osipsEvent.AttrValues[CALLID]
}

// Returns the dialog identifier which opensips needs to disconnect a dialog
func (osipsev *OsipsEvent) GetSessionIds() []string {
	return strings.Split(osipsev.osipsEvent.AttrValues[OSIPS_DIALOG_ID], ":")
}

func (osipsev *OsipsEvent) GetDirection(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.OUT
}

// Account being charged
func (osipsev *OsipsEvent) GetAccount(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.FirstNonEmpty(osipsev.osipsEvent.AttrValues[fieldName], osipsev.osipsEvent.AttrValues[CGR_ACCOUNT])
}

// Rating subject being charged, falls back on account if missing
func (osipsev *OsipsEvent) GetSubject(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.FirstNonEmpty(osipsev.osipsEvent.AttrValues[fieldName], osipsev.osipsEvent.AttrValues[CGR_SUBJECT], osipsev.GetAccount(fieldName))
}

func (osipsev *OsipsEvent) GetDestination(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.FirstNonEmpty(osipsev.osipsEvent.AttrValues[fieldName], osipsev.osipsEvent.AttrValues[CGR_DESTINATION])
}

func (osipsev *OsipsEvent) GetCallDestNr(fieldName string) string {
	return osipsev.GetDestination(fieldName)
}

func (osipsev *OsipsEvent) GetCategory(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.FirstNonEmpty(osipsev.osipsEvent.AttrValues[fieldName], osipsev.osipsEvent.AttrValues[CGR_CATEGORY], config.CgrConfig().DefaultCategory)
}

func (osipsev *OsipsEvent) GetTenant(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.FirstNonEmpty(osipsev.osipsEvent.AttrValues[fieldName], osipsev.osipsEvent.AttrValues[CGR_TENANT], config.CgrConfig().DefaultTenant)
}
func (osipsev *OsipsEvent) GetReqType(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.FirstNonEmpty(osipsev.osipsEvent.AttrValues[fieldName], osipsev.osipsEvent.AttrValues[CGR_REQTYPE], config.CgrConfig().DefaultReqType)
}
func (osipsev *OsipsEvent) GetSetupTime(fieldName, timezone string) (time.Time, error) {
	sTimeStr := utils.FirstNonEmpty(osipsev.osipsEvent.AttrValues[fieldName], osipsev.osipsEvent.AttrValues[OSIPS_SETUP_TIME], osipsev.osipsEvent.AttrValues[OSIPS_EVENT_TIME])
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		sTimeStr = fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.ParseTimeDetectLayout(sTimeStr, timezone)
}
func (osipsev *OsipsEvent) GetAnswerTime(fieldName, timezone string) (time.Time, error) {
	aTimeStr := utils.FirstNonEmpty(osipsev.osipsEvent.AttrValues[fieldName], osipsev.osipsEvent.AttrValues[TIME])
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		aTimeStr = fieldName[len(utils.STATIC_VALUE_PREFIX):]
	} else if fieldName == utils.META_DEFAULT {
		aTimeStr = osipsev.osipsEvent.AttrValues[TIME]
	}
	return utils.ParseTimeDetectLayout(aTimeStr, timezone)
}
func (osipsev *OsipsEvent) GetEndTime(fieldName, timezone string) (time.Time, error) {
	var nilTime time.Time
	aTime, err := osipsev.GetAnswerTime(utils.META_DEFAULT, timezone)
	if err != nil {
		return nilTime, err
	}
	dur, err := osipsev.GetDuration(utils.META_DEFAULT)
	if err != nil {
		return nilTime, err
	}
	return aTime.Add(dur), nil
}
func (osipsev *OsipsEvent) GetDuration(fieldName string) (time.Duration, error) {
	durStr := utils.FirstNonEmpty(osipsev.osipsEvent.AttrValues[fieldName], osipsev.osipsEvent.AttrValues[OSIPS_DURATION])
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		durStr = fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.ParseDurationWithSecs(durStr)
}
func (osipsev *OsipsEvent) GetPdd(fieldName string) (time.Duration, error) {
	var pddStr string
	if utils.IsSliceMember([]string{utils.PDD, utils.META_DEFAULT}, fieldName) {
		pddStr = osipsev.osipsEvent.AttrValues[CGR_PDD]
	} else if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		pddStr = fieldName[len(utils.STATIC_VALUE_PREFIX):]
	} else {
		pddStr = osipsev.osipsEvent.AttrValues[fieldName]
	}
	return utils.ParseDurationWithSecs(pddStr)
}
func (osipsev *OsipsEvent) GetSupplier(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.FirstNonEmpty(osipsev.osipsEvent.AttrValues[fieldName], osipsev.osipsEvent.AttrValues[utils.CGR_SUPPLIER])
}
func (osipsev *OsipsEvent) GetDisconnectCause(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.FirstNonEmpty(osipsev.osipsEvent.AttrValues[fieldName], osipsev.osipsEvent.AttrValues[OSIPS_SIPCODE], osipsev.osipsEvent.AttrValues[utils.DISCONNECT_CAUSE])
}
func (osipsEv *OsipsEvent) GetOriginatorIP(fieldName string) string {
	if osipsEv.osipsEvent == nil || osipsEv.osipsEvent.OriginatorAddress == nil {
		return ""
	}
	return osipsEv.osipsEvent.OriginatorAddress.IP.String()
}
func (osipsev *OsipsEvent) MissingParameter() bool {
	var nilTime time.Time
	if osipsev.GetName() == "E_ACC_EVENT" && osipsev.osipsEvent.AttrValues["method"] == "INVITE" {
		return len(osipsev.GetUUID()) == 0 ||
			len(osipsev.GetAccount(utils.META_DEFAULT)) == 0 ||
			len(osipsev.GetDestination(utils.META_DEFAULT)) == 0 ||
			len(osipsev.osipsEvent.AttrValues[OSIPS_DIALOG_ID]) == 0
	} else if osipsev.GetName() == "E_ACC_EVENT" && osipsev.osipsEvent.AttrValues["method"] == "BYE" {
		return len(osipsev.osipsEvent.AttrValues[OSIPS_DIALOG_ID]) == 0 ||
			len(osipsev.osipsEvent.AttrValues[TIME]) == 0
	} else if osipsev.GetName() == "E_ACC_EVENT" && osipsev.osipsEvent.AttrValues["method"] == "UPDATE" { // Updated event out of start/stop
		// Data needed when stopping a prepaid loop or building a CDR with start/stop event
		setupTime, err := osipsev.GetSetupTime(TIME, config.CgrConfig().DefaultTimezone)
		if err != nil || setupTime.Equal(nilTime) {
			return true
		}
		aTime, err := osipsev.GetAnswerTime(utils.META_DEFAULT, config.CgrConfig().DefaultTimezone)
		if err != nil || aTime.Equal(nilTime) {
			return true
		}
		endTime, err := osipsev.GetEndTime(utils.META_DEFAULT, config.CgrConfig().DefaultTimezone)
		if err != nil || endTime.Equal(nilTime) {
			return true
		}
		_, err = osipsev.GetDuration(utils.META_DEFAULT)
		if err != nil {
			return true
		}
		if osipsev.osipsEvent.AttrValues[OSIPS_DIALOG_ID] == "" {
			return true
		}
		return false
	}
	return true
}
func (osipsev *OsipsEvent) ParseEventValue(fld *utils.RSRField, timezone string) string {
	return ""
}
func (osipsev *OsipsEvent) PassesFieldFilter(*utils.RSRField) (bool, string) {
	return false, ""
}
func (osipsev *OsipsEvent) GetExtraFields() map[string]string {
	primaryFields := []string{TO_TAG, SETUP_DURATION, OSIPS_SETUP_TIME, "method", "callid", "sip_reason", OSIPS_EVENT_TIME, "sip_code", "duration", "from_tag", "dialog_id",
		CGR_TENANT, CGR_CATEGORY, CGR_REQTYPE, CGR_ACCOUNT, CGR_SUBJECT, CGR_DESTINATION, utils.CGR_SUPPLIER, CGR_PDD}
	extraFields := make(map[string]string)
	for field, val := range osipsev.osipsEvent.AttrValues {
		if !utils.IsSliceMember(primaryFields, field) {
			extraFields[field] = val
		}
	}
	return extraFields
}

func (osipsev *OsipsEvent) DialogId() string {
	return osipsev.osipsEvent.AttrValues[OSIPS_DIALOG_ID]
}

func (osipsEv *OsipsEvent) AsStoredCdr(timezone string) *engine.StoredCdr {
	storCdr := new(engine.StoredCdr)
	storCdr.CgrId = osipsEv.GetCgrId(timezone)
	storCdr.TOR = utils.VOICE
	storCdr.AccId = osipsEv.GetUUID()
	storCdr.CdrHost = osipsEv.GetOriginatorIP(utils.META_DEFAULT)
	storCdr.CdrSource = "OSIPS_" + osipsEv.GetName()
	storCdr.ReqType = osipsEv.GetReqType(utils.META_DEFAULT)
	storCdr.Direction = osipsEv.GetDirection(utils.META_DEFAULT)
	storCdr.Tenant = osipsEv.GetTenant(utils.META_DEFAULT)
	storCdr.Category = osipsEv.GetCategory(utils.META_DEFAULT)
	storCdr.Account = osipsEv.GetAccount(utils.META_DEFAULT)
	storCdr.Subject = osipsEv.GetSubject(utils.META_DEFAULT)
	storCdr.Destination = osipsEv.GetDestination(utils.META_DEFAULT)
	storCdr.SetupTime, _ = osipsEv.GetSetupTime(utils.META_DEFAULT, timezone)
	storCdr.AnswerTime, _ = osipsEv.GetAnswerTime(utils.META_DEFAULT, timezone)
	storCdr.Usage, _ = osipsEv.GetDuration(utils.META_DEFAULT)
	storCdr.Pdd, _ = osipsEv.GetPdd(utils.META_DEFAULT)
	storCdr.Supplier = osipsEv.GetSupplier(utils.META_DEFAULT)
	storCdr.DisconnectCause = osipsEv.GetDisconnectCause(utils.META_DEFAULT)
	storCdr.ExtraFields = osipsEv.GetExtraFields()
	storCdr.Cost = -1
	return storCdr
}

// Computes duration out of setup time of the callEnd
func (osipsEv *OsipsEvent) updateDurationFromEvent(updatedOsipsEv *OsipsEvent) error {
	endTime, err := updatedOsipsEv.GetSetupTime(TIME, config.CgrConfig().DefaultTimezone)
	if err != nil {
		return err
	}
	answerTime, err := osipsEv.GetAnswerTime(utils.META_DEFAULT, config.CgrConfig().DefaultTimezone)
	osipsEv.osipsEvent.AttrValues[OSIPS_DURATION] = endTime.Sub(answerTime).String()
	osipsEv.osipsEvent.AttrValues["method"] = "UPDATE" // So we can know it is an end event
	osipsEv.osipsEvent.AttrValues[OSIPS_SIPCODE] = updatedOsipsEv.osipsEvent.AttrValues[OSIPS_SIPCODE]
	return nil
}

func (osipsEv *OsipsEvent) ComputeLcr() bool {
	if computeLcr, err := strconv.ParseBool(osipsEv.osipsEvent.AttrValues[utils.CGR_COMPUTELCR]); err != nil {
		return false
	} else {
		return computeLcr
	}
}
