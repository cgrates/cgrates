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
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/fsock"
)

// ToDo: Introduce support for RSRFields

// Event type holding a mapping of all event's proprieties
type FSEvent map[string]string

const (
	// Freswitch event proprities names
	DIRECTION          = "Call-Direction"
	SUBJECT            = "variable_" + utils.CGR_SUBJECT
	ACCOUNT            = "variable_" + utils.CGR_ACCOUNT
	DESTINATION        = "variable_" + utils.CGR_DESTINATION
	REQTYPE            = "variable_" + utils.CGR_REQTYPE //prepaid or postpaid
	Category           = "variable_" + utils.CGR_CATEGORY
	VAR_CGR_SUPPLIER   = "variable_" + utils.CGR_SUPPLIER
	UUID               = "Unique-ID" // -Unique ID for this call leg
	CSTMID             = "variable_" + utils.CGR_TENANT
	CALL_DEST_NR       = "Caller-Destination-Number"
	SIP_REQ_USER       = "variable_sip_req_user"
	PARK_TIME          = "Caller-Profile-Created-Time"
	SETUP_TIME         = "Caller-Channel-Created-Time"
	ANSWER_TIME        = "Caller-Channel-Answered-Time"
	END_TIME           = "Caller-Channel-Hangup-Time"
	DURATION           = "variable_billsec"
	NAME               = "Event-Name"
	HEARTBEAT          = "HEARTBEAT"
	ANSWER             = "CHANNEL_ANSWER"
	HANGUP             = "CHANNEL_HANGUP_COMPLETE"
	PARK               = "CHANNEL_PARK"
	AUTH_OK            = "+AUTH_OK"
	DISCONNECT         = "+SWITCH DISCONNECT"
	INSUFFICIENT_FUNDS = "-INSUFFICIENT_FUNDS"
	MISSING_PARAMETER  = "-MISSING_PARAMETER"
	SYSTEM_ERROR       = "-SYSTEM_ERROR"
	MANAGER_REQUEST    = "+MANAGER_REQUEST"
	USERNAME           = "Caller-Username"
	FS_IPv4            = "FreeSWITCH-IPv4"
	HANGUP_CAUSE       = "Hangup-Cause"
	PDD_MEDIA_MS       = "variable_progress_mediamsec"
	PDD_NOMEDIA_MS     = "variable_progressmsec"
	IGNOREPARK         = "variable_cgr_ignorepark"

	VAR_CGR_DISCONNECT_CAUSE = "variable_" + utils.CGR_DISCONNECT_CAUSE
	VAR_CGR_CMPUTELCR        = "variable_" + utils.CGR_COMPUTELCR
)

// Nice printing for the event object.
func (fsev FSEvent) String() (result string) {
	for k, v := range fsev {
		result += fmt.Sprintf("%s = %s\n", k, v)
	}
	result += "=============================================================="
	return
}

// Loads the new event data from a body of text containing the key value proprieties.
// It stores the parsed proprieties in the internal map.
func (fsev FSEvent) AsEvent(body string) engine.Event {
	fsev = fsock.FSEventStrToMap(body, nil)
	return fsev
}

func (fsev FSEvent) GetName() string {
	return fsev[NAME]
}
func (fsev FSEvent) GetDirection(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	//TODO: implement direction
	return utils.OUT
}
func (fsev FSEvent) GetSubject(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	} else if fieldName == utils.META_DEFAULT {
		return utils.FirstNonEmpty(fsev[SUBJECT], fsev[USERNAME])
	}
	return utils.FirstNonEmpty(fsev[fieldName], fsev[SUBJECT], fsev[USERNAME])
}

func (fsev FSEvent) GetAccount(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	} else if fieldName == utils.META_DEFAULT {
		return utils.FirstNonEmpty(fsev[ACCOUNT], fsev[USERNAME])
	}
	return utils.FirstNonEmpty(fsev[fieldName], fsev[ACCOUNT], fsev[USERNAME])
}

// Charging destination number
func (fsev FSEvent) GetDestination(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	} else if fieldName == utils.META_DEFAULT {
		return utils.FirstNonEmpty(fsev[DESTINATION], fsev[CALL_DEST_NR], fsev[SIP_REQ_USER])
	}
	return utils.FirstNonEmpty(fsev[fieldName], fsev[DESTINATION], fsev[CALL_DEST_NR], fsev[SIP_REQ_USER])
}

// Original dialed destination number, useful in case of unpark
func (fsev FSEvent) GetCallDestNr(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	} else if fieldName == utils.META_DEFAULT {
		return utils.FirstNonEmpty(fsev[CALL_DEST_NR], fsev[SIP_REQ_USER])
	}
	return utils.FirstNonEmpty(fsev[fieldName], fsev[CALL_DEST_NR], fsev[SIP_REQ_USER])
}
func (fsev FSEvent) GetCategory(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	} else if fieldName == utils.META_DEFAULT {
		return utils.FirstNonEmpty(fsev[Category], config.CgrConfig().DefaultCategory)
	}
	return utils.FirstNonEmpty(fsev[fieldName], fsev[Category], config.CgrConfig().DefaultCategory)
}
func (fsev FSEvent) GetCgrId() string {
	setupTime, _ := fsev.GetSetupTime(utils.META_DEFAULT)
	return utils.Sha1(fsev[UUID], setupTime.UTC().String())
}
func (fsev FSEvent) GetUUID() string {
	return fsev[UUID]
}
func (fsev FSEvent) GetSessionIds() []string {
	return []string{fsev.GetUUID()}
}
func (fsev FSEvent) GetTenant(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	} else if fieldName == utils.META_DEFAULT {
		return utils.FirstNonEmpty(fsev[CSTMID], config.CgrConfig().DefaultTenant)
	}
	return utils.FirstNonEmpty(fsev[fieldName], fsev[CSTMID], config.CgrConfig().DefaultTenant)
}
func (fsev FSEvent) GetReqType(fieldName string) string {
	var reqTypeDetected = ""                     // Used to automatically disable processing of the request
	if fsev["variable_process_cdr"] == "false" { // FS will not generated CDR here
		reqTypeDetected = utils.META_NONE
	} else if fsev["Caller-Dialplan"] == "inline" { // Used for internally generated dialplan, eg refer coming from another box, not in our control
		reqTypeDetected = utils.META_NONE
	}
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	} else if fieldName == utils.META_DEFAULT {
		return utils.FirstNonEmpty(fsev[REQTYPE], reqTypeDetected, config.CgrConfig().DefaultReqType)
	}
	return utils.FirstNonEmpty(fsev[fieldName], fsev[REQTYPE], reqTypeDetected, config.CgrConfig().DefaultReqType)
}
func (fsev FSEvent) MissingParameter() bool {
	return strings.TrimSpace(fsev.GetDirection(utils.META_DEFAULT)) == "" ||
		strings.TrimSpace(fsev.GetSubject(utils.META_DEFAULT)) == "" ||
		strings.TrimSpace(fsev.GetAccount(utils.META_DEFAULT)) == "" ||
		strings.TrimSpace(fsev.GetDestination(utils.META_DEFAULT)) == "" ||
		strings.TrimSpace(fsev.GetCategory(utils.META_DEFAULT)) == "" ||
		strings.TrimSpace(fsev.GetUUID()) == "" ||
		strings.TrimSpace(fsev.GetTenant(utils.META_DEFAULT)) == "" ||
		strings.TrimSpace(fsev.GetCallDestNr(utils.META_DEFAULT)) == ""
}
func (fsev FSEvent) GetSetupTime(fieldName string) (t time.Time, err error) {
	fsSTimeStr, hasKey := fsev[SETUP_TIME]
	if hasKey && fsSTimeStr != "0" {
		// Discard the nanoseconds information since MySQL cannot store them in early versions and csv uses default seconds so cgrid will not corelate
		fsSTimeStr = fsSTimeStr[:len(fsSTimeStr)-6]
	}
	sTimeStr := utils.FirstNonEmpty(fsev[fieldName], fsSTimeStr)
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		sTimeStr = fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.ParseTimeDetectLayout(sTimeStr)
}
func (fsev FSEvent) GetAnswerTime(fieldName string) (t time.Time, err error) {
	fsATimeStr, hasKey := fsev[ANSWER_TIME]
	if hasKey && fsATimeStr != "0" {
		// Discard the nanoseconds information since MySQL cannot store them in early versions and csv uses default seconds so cgrid will not corelate
		fsATimeStr = fsATimeStr[:len(fsATimeStr)-6]
	}
	aTimeStr := utils.FirstNonEmpty(fsev[fieldName], fsATimeStr)
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		aTimeStr = fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.ParseTimeDetectLayout(aTimeStr)
}

func (fsev FSEvent) GetEndTime() (t time.Time, err error) {
	return utils.ParseTimeDetectLayout(fsev[END_TIME])
}

func (fsev FSEvent) GetDuration(fieldName string) (time.Duration, error) {
	durStr := utils.FirstNonEmpty(fsev[fieldName], fsev[DURATION])
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		durStr = fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.ParseDurationWithSecs(durStr)
}

func (fsev FSEvent) GetPdd(fieldName string) (time.Duration, error) {
	var pddStr string
	if utils.IsSliceMember([]string{utils.PDD, utils.META_DEFAULT}, fieldName) {
		pddStr = utils.FirstNonEmpty(fsev[PDD_MEDIA_MS], fsev[PDD_NOMEDIA_MS])
		if len(pddStr) != 0 {
			pddStr = pddStr + "ms" // PDD is in milliseconds and CGR expects it in seconds
		}
	} else if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		pddStr = fieldName[len(utils.STATIC_VALUE_PREFIX):]
	} else {
		pddStr = fsev[fieldName]
	}
	return utils.ParseDurationWithSecs(pddStr)
}

func (fsev FSEvent) GetSupplier(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.FirstNonEmpty(fsev[fieldName], fsev[VAR_CGR_SUPPLIER])
}

func (fsev FSEvent) GetDisconnectCause(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.FirstNonEmpty(fsev[fieldName], fsev[VAR_CGR_DISCONNECT_CAUSE], fsev[HANGUP_CAUSE])
}

func (fsev FSEvent) GetOriginatorIP(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	} else if fieldName == utils.META_DEFAULT {
		return fsev[FS_IPv4]
	}
	return utils.FirstNonEmpty(fsev[fieldName], fsev[FS_IPv4])
}

func (fsev FSEvent) GetExtraFields() map[string]string {
	extraFields := make(map[string]string)
	for _, fldRule := range config.CgrConfig().SmFsConfig.CdrExtraFields {
		extraFields[fldRule.Id] = fsev.ParseEventValue(fldRule)
	}
	return extraFields
}

// Used in derived charging and sittuations when we need to run regexp on fields
func (fsev FSEvent) ParseEventValue(rsrFld *utils.RSRField) string {
	switch rsrFld.Id {
	case utils.CGRID:
		return rsrFld.ParseValue(fsev.GetCgrId())
	case utils.TOR:
		return rsrFld.ParseValue(utils.VOICE)
	case utils.ACCID:
		return rsrFld.ParseValue(fsev.GetUUID())
	case utils.CDRHOST:
		return rsrFld.ParseValue(fsev["FreeSWITCH-IPv4"])
	case utils.CDRSOURCE:
		return rsrFld.ParseValue("FS_EVENT")
	case utils.REQTYPE:
		return rsrFld.ParseValue(fsev.GetReqType(""))
	case utils.DIRECTION:
		return rsrFld.ParseValue(fsev.GetDirection(""))
	case utils.TENANT:
		return rsrFld.ParseValue(fsev.GetTenant(""))
	case utils.CATEGORY:
		return rsrFld.ParseValue(fsev.GetCategory(""))
	case utils.ACCOUNT:
		return rsrFld.ParseValue(fsev.GetAccount(""))
	case utils.SUBJECT:
		return rsrFld.ParseValue(fsev.GetSubject(""))
	case utils.DESTINATION:
		return rsrFld.ParseValue(fsev.GetDestination(""))
	case utils.SETUP_TIME:
		st, _ := fsev.GetSetupTime("")
		return rsrFld.ParseValue(st.String())
	case utils.ANSWER_TIME:
		at, _ := fsev.GetAnswerTime("")
		return rsrFld.ParseValue(at.String())
	case utils.USAGE:
		dur, _ := fsev.GetDuration("")
		return rsrFld.ParseValue(strconv.FormatInt(dur.Nanoseconds(), 10))
	case utils.PDD:
		pdd, _ := fsev.GetPdd(utils.META_DEFAULT)
		return rsrFld.ParseValue(strconv.FormatFloat(pdd.Seconds(), 'f', -1, 64))
	case utils.SUPPLIER:
		return rsrFld.ParseValue(fsev.GetSupplier(""))
	case utils.DISCONNECT_CAUSE:
		return rsrFld.ParseValue(fsev.GetDisconnectCause(""))
	case utils.MEDI_RUNID:
		return rsrFld.ParseValue(utils.DEFAULT_RUNID)
	case utils.COST:
		return rsrFld.ParseValue(strconv.FormatFloat(-1, 'f', -1, 64)) // Recommended to use FormatCost
	default:
		return rsrFld.ParseValue(fsev[rsrFld.Id])
	}
}

func (fsev FSEvent) PassesFieldFilter(fieldFilter *utils.RSRField) (bool, string) {
	// Keep in sync (or merge) with StoredCdr.PassesFieldFielter()
	if fieldFilter == nil {
		return true, ""
	}
	if fieldFilter.IsStatic() && fsev.ParseEventValue(&utils.RSRField{Id: fieldFilter.Id}) == fsev.ParseEventValue(fieldFilter) {
		return true, fsev.ParseEventValue(&utils.RSRField{Id: fieldFilter.Id})
	}
	preparedFilter := &utils.RSRField{Id: fieldFilter.Id, RSRules: make([]*utils.ReSearchReplace, len(fieldFilter.RSRules))} // Reset rules so they do not point towards same structures as original fieldFilter
	for idx := range fieldFilter.RSRules {
		// Hardcode the template with maximum of 5 groups ordered
		preparedFilter.RSRules[idx] = &utils.ReSearchReplace{SearchRegexp: fieldFilter.RSRules[idx].SearchRegexp, ReplaceTemplate: utils.FILTER_REGEXP_TPL}
	}
	preparedVal := fsev.ParseEventValue(preparedFilter)
	filteredValue := fsev.ParseEventValue(fieldFilter)
	if preparedFilter.RegexpMatched() && (len(preparedVal) == 0 || preparedVal == filteredValue) {
		return true, filteredValue
	}
	return false, ""
}

func (fsev FSEvent) AsStoredCdr() *engine.StoredCdr {
	storCdr := new(engine.StoredCdr)
	storCdr.CgrId = fsev.GetCgrId()
	storCdr.TOR = utils.VOICE
	storCdr.AccId = fsev.GetUUID()
	storCdr.CdrHost = fsev.GetOriginatorIP(utils.META_DEFAULT)
	storCdr.CdrSource = "FS_" + fsev.GetName()
	storCdr.ReqType = fsev.GetReqType(utils.META_DEFAULT)
	storCdr.Direction = fsev.GetDirection(utils.META_DEFAULT)
	storCdr.Tenant = fsev.GetTenant(utils.META_DEFAULT)
	storCdr.Category = fsev.GetCategory(utils.META_DEFAULT)
	storCdr.Account = fsev.GetAccount(utils.META_DEFAULT)
	storCdr.Subject = fsev.GetSubject(utils.META_DEFAULT)
	storCdr.Destination = fsev.GetDestination(utils.META_DEFAULT)
	storCdr.SetupTime, _ = fsev.GetSetupTime(utils.META_DEFAULT)
	storCdr.AnswerTime, _ = fsev.GetAnswerTime(utils.META_DEFAULT)
	storCdr.Usage, _ = fsev.GetDuration(utils.META_DEFAULT)
	storCdr.Pdd, _ = fsev.GetPdd(utils.META_DEFAULT)
	storCdr.ExtraFields = fsev.GetExtraFields()
	storCdr.Cost = -1
	storCdr.Supplier = fsev.GetSupplier(utils.META_DEFAULT)
	storCdr.DisconnectCause = fsev.GetDisconnectCause(utils.META_DEFAULT)
	return storCdr
}

func (fsev FSEvent) ComputeLcr() bool {
	if computeLcr, err := strconv.ParseBool(fsev[VAR_CGR_CMPUTELCR]); err != nil {
		return false
	} else {
		return computeLcr
	}
}

// Converts into CallDescriptor due to responder interface needs
func (fsev FSEvent) AsCallDescriptor() (*engine.CallDescriptor, error) {
	lcrReq := &engine.LcrRequest{
		Direction:   fsev.GetDirection(utils.META_DEFAULT),
		Tenant:      fsev.GetTenant(utils.META_DEFAULT),
		Category:    fsev.GetCategory(utils.META_DEFAULT),
		Account:     fsev.GetAccount(utils.META_DEFAULT),
		Subject:     fsev.GetSubject(utils.META_DEFAULT),
		Destination: fsev.GetDestination(utils.META_DEFAULT),
		StartTime:   utils.FirstNonEmpty(fsev[SETUP_TIME], fsev[ANSWER_TIME]),
		Duration:    fsev[DURATION],
	}
	return lcrReq.AsCallDescriptor()
}

// Converts a slice of strings into a FS array string, contains len(array) at first index since FS does not support len(ARRAY::) for now
func SliceAsFsArray(slc []string) string {
	arry := ""
	if len(slc) == 0 {
		return arry
	}
	for idx, itm := range slc {
		if idx == 0 {
			arry = fmt.Sprintf("ARRAY::%d|:%s", len(slc), itm)
		} else {
			arry += "|:" + itm
		}
	}
	return arry
}
