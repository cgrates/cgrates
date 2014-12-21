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
	"fmt"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/osipsdagram"
	"strings"
	"time"
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
	OSIPS_DURATION           = "duration"
	OSIPS_AUTH_OK            = "AUTH_OK"
	OSIPS_INSUFFICIENT_FUNDS = "INSUFFICIENT_FUNDS"
)

func NewOsipsEvent(osipsDagramEvent *osipsdagram.OsipsEvent) (*OsipsEvent, error) {
	return &OsipsEvent{osipsEvent: osipsDagramEvent}, nil
}

type OsipsEvent struct {
	osipsEvent *osipsdagram.OsipsEvent
}

func (osipsev *OsipsEvent) New(evStr string) Event {
	return osipsev
}

func (osipsev *OsipsEvent) String() string {
	return fmt.Sprintf("%+v", osipsev)
}

func (osipsev *OsipsEvent) GetName() string {
	return osipsev.osipsEvent.Name
}

func (osipsev *OsipsEvent) GetCgrId() string {
	setupTime, _ := osipsev.GetSetupTime(utils.META_DEFAULT)
	return utils.Sha1(osipsev.GetUUID(), setupTime.UTC().String())
}

func (osipsev *OsipsEvent) GetUUID() string {
	return osipsev.osipsEvent.AttrValues[CALLID] + ";" + osipsev.osipsEvent.AttrValues[FROM_TAG] + ";" + osipsev.osipsEvent.AttrValues[TO_TAG]
}

func (osipsev *OsipsEvent) GetDirection(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.OUT
}

func (osipsev *OsipsEvent) GetSubject(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.FirstNonEmpty(osipsev.osipsEvent.AttrValues[fieldName], osipsev.osipsEvent.AttrValues[CGR_SUBJECT])
}

func (osipsev *OsipsEvent) GetAccount(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.FirstNonEmpty(osipsev.osipsEvent.AttrValues[fieldName], osipsev.osipsEvent.AttrValues[CGR_ACCOUNT])
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
func (osipsev *OsipsEvent) GetSetupTime(fieldName string) (time.Time, error) {
	sTimeStr := utils.FirstNonEmpty(osipsev.osipsEvent.AttrValues[fieldName], osipsev.osipsEvent.AttrValues[OSIPS_SETUP_TIME])
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		sTimeStr = fieldName[len(utils.STATIC_VALUE_PREFIX):]
	} else if fieldName == utils.META_DEFAULT {
		sTimeStr = osipsev.osipsEvent.AttrValues[OSIPS_SETUP_TIME]
	}
	return utils.ParseTimeDetectLayout(sTimeStr)
}
func (osipsev *OsipsEvent) GetAnswerTime(fieldName string) (time.Time, error) {
	aTimeStr := utils.FirstNonEmpty(osipsev.osipsEvent.AttrValues[fieldName], osipsev.osipsEvent.AttrValues[TIME])
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		aTimeStr = fieldName[len(utils.STATIC_VALUE_PREFIX):]
	} else if fieldName == utils.META_DEFAULT {
		aTimeStr = osipsev.osipsEvent.AttrValues[TIME]
	}
	return utils.ParseTimeDetectLayout(aTimeStr)
}
func (osipsev *OsipsEvent) GetEndTime() (time.Time, error) {
	var nilTime time.Time
	aTime, err := osipsev.GetAnswerTime(utils.META_DEFAULT)
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
func (osipsEv *OsipsEvent) GetOriginatorIP(fieldName string) string {
	if osipsEv.osipsEvent == nil || osipsEv.osipsEvent.OriginatorAddress == nil {
		return ""
	}
	return osipsEv.osipsEvent.OriginatorAddress.IP.String()
}
func (osipsev *OsipsEvent) MissingParameter(eventName string) bool {
	return len(osipsev.GetUUID()) == 0 ||
		len(osipsev.GetAccount(utils.META_DEFAULT)) == 0 ||
		len(osipsev.GetSubject(utils.META_DEFAULT)) == 0 ||
		len(osipsev.GetDestination(utils.META_DEFAULT)) == 0
}
func (osipsev *OsipsEvent) ParseEventValue(*utils.RSRField) string {
	return ""
}
func (osipsev *OsipsEvent) PassesFieldFilter(*utils.RSRField) (bool, string) {
	return false, ""
}
func (osipsev *OsipsEvent) GetExtraFields() map[string]string {
	primaryFields := []string{"to_tag", "setuptime", "created", "method", "callid", "sip_reason", "time", "sip_code", "duration", "from_tag",
		"cgr_tenant", "cgr_category", "cgr_reqtype", "cgr_account", "cgr_subject", "cgr_destination"}
	extraFields := make(map[string]string)
	for field, val := range osipsev.osipsEvent.AttrValues {
		if !utils.IsSliceMember(primaryFields, field) {
			extraFields[field] = val
		}
	}
	return extraFields
}

func (osipsEv *OsipsEvent) AsStoredCdr() *utils.StoredCdr {
	storCdr := new(utils.StoredCdr)
	storCdr.CgrId = osipsEv.GetCgrId()
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
	storCdr.SetupTime, _ = osipsEv.GetSetupTime(utils.META_DEFAULT)
	storCdr.AnswerTime, _ = osipsEv.GetAnswerTime(utils.META_DEFAULT)
	storCdr.Usage, _ = osipsEv.GetDuration(utils.META_DEFAULT)
	storCdr.ExtraFields = osipsEv.GetExtraFields()
	storCdr.Cost = -1
	return storCdr
}
