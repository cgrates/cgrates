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

package agents

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/fsock"
)

// ToDo: Introduce support for RSRFields

const (
	// Freswitch event proprities names
	DIRECTION                = "Call-Direction"
	SUBJECT                  = "variable_" + utils.CGR_SUBJECT
	ACCOUNT                  = "variable_" + utils.CGR_ACCOUNT
	DESTINATION              = "variable_" + utils.CGR_DESTINATION
	REQTYPE                  = "variable_" + utils.CGR_REQTYPE //prepaid or postpaid
	CATEGORY                 = "variable_" + utils.CGR_CATEGORY
	VAR_CGR_SUPPLIER         = "variable_" + utils.CGR_SUPPLIER
	UUID                     = "Unique-ID" // -Unique ID for this call leg
	CSTMID                   = "variable_" + utils.CGR_TENANT
	CALL_DEST_NR             = "Caller-Destination-Number"
	SIP_REQ_USER             = "variable_sip_req_user"
	PARK_TIME                = "Caller-Profile-Created-Time"
	SETUP_TIME               = "Caller-Channel-Created-Time"
	ANSWER_TIME              = "Caller-Channel-Answered-Time"
	END_TIME                 = "Caller-Channel-Hangup-Time"
	DURATION                 = "variable_billsec"
	NAME                     = "Event-Name"
	HEARTBEAT                = "HEARTBEAT"
	ANSWER                   = "CHANNEL_ANSWER"
	HANGUP                   = "CHANNEL_HANGUP_COMPLETE"
	PARK                     = "CHANNEL_PARK"
	AUTH_OK                  = "AUTH_OK"
	DISCONNECT               = "SWITCH DISCONNECT"
	MANAGER_REQUEST          = "MANAGER_REQUEST"
	USERNAME                 = "Caller-Username"
	FS_IPv4                  = "FreeSWITCH-IPv4"
	HANGUP_CAUSE             = "Hangup-Cause"
	PDD_MEDIA_MS             = "variable_progress_mediamsec"
	PDD_NOMEDIA_MS           = "variable_progressmsec"
	IGNOREPARK               = "variable_cgr_ignorepark"
	FS_VARPREFIX             = "variable_"
	VarCGRSubsystems         = "variable_" + utils.CGRSubsystems
	CGRResourceAllocation    = "cgr_resource_allocation"
	VAR_CGR_DISCONNECT_CAUSE = "variable_" + utils.CGR_DISCONNECT_CAUSE
	VAR_CGR_CMPUTELCR        = "variable_" + utils.CGR_COMPUTELCR
	FsConnID                 = "FsConnID" // used to share connID info in event for remote disconnects
	VarAnswerEpoch           = "variable_answer_epoch"
	VarCGRACD                = "variable_" + utils.CGR_ACD
	VarCGROriginHost         = "variable_" + utils.CGROriginHost
)

func NewFSEvent(strEv string) (fsev FSEvent) {
	return fsock.FSEventStrToMap(strEv, nil)
}

// Event type holding a mapping of all event's proprieties
type FSEvent map[string]string

// Nice printing for the event object.
func (fsev FSEvent) String() (result string) {
	for k, v := range fsev {
		result += fmt.Sprintf("%s = %s\n", k, v)
	}
	result += "=============================================================="
	return
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

// Account calling
func (fsev FSEvent) GetAccount(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.FirstNonEmpty(fsev[fieldName], fsev[ACCOUNT], fsev[USERNAME])
}

// Rating subject being charged
func (fsev FSEvent) GetSubject(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.FirstNonEmpty(fsev[fieldName], fsev[SUBJECT], fsev.GetAccount(fieldName))
}

// Charging destination number
func (fsev FSEvent) GetDestination(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.FirstNonEmpty(fsev[fieldName], fsev[DESTINATION],
		fsev[CALL_DEST_NR], fsev[SIP_REQ_USER])
}

// Original dialed destination number, useful in case of unpark
func (fsev FSEvent) GetCallDestNr(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.FirstNonEmpty(fsev[fieldName], fsev[CALL_DEST_NR], fsev[SIP_REQ_USER])
}
func (fsev FSEvent) GetCategory(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.FirstNonEmpty(fsev[fieldName], fsev[CATEGORY],
		config.CgrConfig().GeneralCfg().DefaultCategory)
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
	}
	return utils.FirstNonEmpty(fsev[fieldName], fsev[CSTMID],
		config.CgrConfig().GeneralCfg().DefaultTenant)
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
	}
	return utils.FirstNonEmpty(fsev[fieldName], fsev[REQTYPE],
		reqTypeDetected, config.CgrConfig().GeneralCfg().DefaultReqType)
}

func (fsev FSEvent) MissingParameter(timezone string) string {
	if strings.TrimSpace(fsev.GetDirection(utils.META_DEFAULT)) == "" {
		return utils.Direction
	}
	if strings.TrimSpace(fsev.GetAccount(utils.META_DEFAULT)) == "" {
		return utils.Account
	}
	if strings.TrimSpace(fsev.GetSubject(utils.META_DEFAULT)) == "" {
		return utils.Subject
	}
	if strings.TrimSpace(fsev.GetDestination(utils.META_DEFAULT)) == "" {
		return utils.Destination
	}
	if strings.TrimSpace(fsev.GetCategory(utils.META_DEFAULT)) == "" {
		return utils.Category
	}
	if strings.TrimSpace(fsev.GetUUID()) == "" {
		return utils.OriginID
	}
	if strings.TrimSpace(fsev.GetTenant(utils.META_DEFAULT)) == "" {
		return utils.Tenant
	}
	if strings.TrimSpace(fsev.GetCallDestNr(utils.META_DEFAULT)) == "" {
		return CALL_DEST_NR
	}
	return ""
}

func (fsev FSEvent) GetSetupTime(fieldName, timezone string) (t time.Time, err error) {
	fsSTimeStr, hasKey := fsev[SETUP_TIME]
	if hasKey && fsSTimeStr != "0" {
		// Discard the nanoseconds information since MySQL cannot store them in early versions and csv uses default seconds so CGRID will not corelate
		fsSTimeStr = fsSTimeStr[:len(fsSTimeStr)-6]
	}
	sTimeStr := utils.FirstNonEmpty(fsev[fieldName], fsSTimeStr)
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		sTimeStr = fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.ParseTimeDetectLayout(sTimeStr, timezone)
}

func (fsev FSEvent) GetAnswerTime(fieldName, timezone string) (t time.Time, err error) {
	fsATimeStr, hasKey := fsev[ANSWER_TIME]
	if hasKey && fsATimeStr != "0" {
		// Discard the nanoseconds information since MySQL cannot store them in early versions and csv uses default seconds so CGRID will not corelate
		fsATimeStr = fsATimeStr[:len(fsATimeStr)-6]
	}
	aTimeStr := utils.FirstNonEmpty(fsev[fieldName], fsATimeStr)
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		aTimeStr = fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.ParseTimeDetectLayout(aTimeStr, timezone)
}

func (fsev FSEvent) GetEndTime(fieldName, timezone string) (t time.Time, err error) {
	return utils.ParseTimeDetectLayout(fsev[END_TIME], timezone)
}

func (fsev FSEvent) GetDuration(fieldName string) (time.Duration, error) {
	durStr := utils.FirstNonEmpty(fsev[fieldName], fsev[DURATION])
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		durStr = fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.ParseDurationWithSecs(durStr)
}

func (fsev FSEvent) GetPdd(fieldName string) (time.Duration, error) {
	var PDDStr string
	if utils.IsSliceMember([]string{utils.PDD, utils.META_DEFAULT}, fieldName) {
		PDDStr = utils.FirstNonEmpty(fsev[PDD_MEDIA_MS], fsev[PDD_NOMEDIA_MS])
		if len(PDDStr) != 0 {
			PDDStr = PDDStr + "ms" // PDD is in milliseconds and CGR expects it in seconds
		}
	} else if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		PDDStr = fieldName[len(utils.STATIC_VALUE_PREFIX):]
	} else {
		PDDStr = fsev[fieldName]
	}
	return utils.ParseDurationWithSecs(PDDStr)
}

func (fsev FSEvent) GetADC(fieldName string) (time.Duration, error) {
	var ACDStr string
	if utils.IsSliceMember([]string{utils.ACD, utils.META_DEFAULT}, fieldName) {
		ACDStr = utils.FirstNonEmpty(fsev[VarCGRACD])
		if len(ACDStr) != 0 {
			ACDStr = ACDStr + "s" //  ACD is in seconds and CGR expects it in seconds
		}
	} else if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		ACDStr = fieldName[len(utils.STATIC_VALUE_PREFIX):]
	} else {
		ACDStr = fsev[fieldName]
	}
	return utils.ParseDurationWithSecs(ACDStr)
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
	}
	return utils.FirstNonEmpty(fsev[fieldName], fsev[VarCGROriginHost], fsev[FS_IPv4])
}

func (fsev FSEvent) GetExtraFields() map[string]string {
	extraFields := make(map[string]string)
	for _, fldRule := range config.CgrConfig().FsAgentCfg().ExtraFields {
		if parsed, err := fsev.ParseEventValue(fldRule,
			config.CgrConfig().GeneralCfg().DefaultTimezone); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<%s> error: %s parsing event rule: %+v", utils.FreeSWITCHAgent, err.Error(), fldRule))
		} else {
			extraFields[fldRule.AttrName()] = parsed
		}
	}
	return extraFields
}

// Used in derived charging and sittuations when we need to run regexp on fields
func (fsev FSEvent) ParseEventValue(rsrFld *config.RSRParser, timezone string) (parsed string, err error) {
	switch rsrFld.AttrName() {
	case utils.ToR:
		return rsrFld.ParseValue(utils.VOICE)
	case utils.OriginID:
		return rsrFld.ParseValue(fsev.GetUUID())
	case utils.OriginHost:
		return rsrFld.ParseValue(utils.FirstNonEmpty(fsev[VarCGROriginHost], fsev[FS_IPv4]))
	case utils.Source:
		return rsrFld.ParseValue("FS_EVENT")
	case utils.RequestType:
		return rsrFld.ParseValue(fsev.GetReqType(""))
	case utils.Direction:
		return rsrFld.ParseValue(fsev.GetDirection(""))
	case utils.Tenant:
		return rsrFld.ParseValue(fsev.GetTenant(""))
	case utils.Category:
		return rsrFld.ParseValue(fsev.GetCategory(""))
	case utils.Account:
		return rsrFld.ParseValue(fsev.GetAccount(""))
	case utils.Subject:
		return rsrFld.ParseValue(fsev.GetSubject(""))
	case utils.Destination:
		return rsrFld.ParseValue(fsev.GetDestination(""))
	case utils.SetupTime:
		st, _ := fsev.GetSetupTime("", timezone)
		return rsrFld.ParseValue(st.String())
	case utils.AnswerTime:
		at, _ := fsev.GetAnswerTime("", timezone)
		return rsrFld.ParseValue(at.String())
	case utils.Usage:
		dur, _ := fsev.GetDuration("")
		return rsrFld.ParseValue(strconv.FormatInt(dur.Nanoseconds(), 10))
	case utils.PDD:
		PDD, _ := fsev.GetPdd(utils.META_DEFAULT)
		return rsrFld.ParseValue(strconv.FormatFloat(PDD.Seconds(), 'f', -1, 64))
	case utils.SUPPLIER:
		return rsrFld.ParseValue(fsev.GetSupplier(""))
	case utils.DISCONNECT_CAUSE:
		return rsrFld.ParseValue(fsev.GetDisconnectCause(""))
	case utils.RunID:
		return rsrFld.ParseValue(utils.DEFAULT_RUNID)
	case utils.Cost:
		return rsrFld.ParseValue(strconv.FormatFloat(-1, 'f', -1, 64)) // Recommended to use FormatCost
	default:
		if parsed, err = rsrFld.ParseValue(fsev[rsrFld.AttrName()]); err != nil {
			parsed, err = rsrFld.ParseValue(fsev[FS_VARPREFIX+rsrFld.AttrName()])
		}
		return
	}
}

// AsCGREvent converts FSEvent into CGREvent
func (fsev FSEvent) AsCGREvent(timezone string) (cgrEv *utils.CGREvent, err error) {
	sTime, err := fsev.GetSetupTime(utils.META_DEFAULT, timezone)
	if err != nil {
		return nil, err
	}
	cgrEv = &utils.CGREvent{
		Tenant: fsev.GetTenant(utils.META_DEFAULT),
		ID:     utils.UUIDSha1Prefix(),
		Time:   &sTime,
		Event:  fsev.AsMapStringInterface(timezone),
	}
	return cgrEv, nil
}

// Used with RLs
func (fsev FSEvent) AsMapStringInterface(timezone string) map[string]interface{} {
	mp := make(map[string]interface{})
	for fld, val := range fsev.GetExtraFields() {
		mp[fld] = val
	}
	mp[utils.ToR] = utils.VOICE
	mp[utils.OriginID] = fsev.GetUUID()
	mp[utils.OriginHost] = fsev.GetOriginatorIP(utils.META_DEFAULT)
	mp[utils.Source] = "FS_" + fsev.GetName()
	mp[utils.RequestType] = fsev.GetReqType(utils.META_DEFAULT)
	mp[utils.Direction] = fsev.GetDirection(utils.META_DEFAULT)
	mp[utils.Tenant] = fsev.GetTenant(utils.META_DEFAULT)
	mp[utils.Category] = fsev.GetCategory(utils.META_DEFAULT)
	mp[utils.Account] = fsev.GetAccount(utils.META_DEFAULT)
	mp[utils.Subject] = fsev.GetSubject(utils.META_DEFAULT)
	mp[utils.Destination] = fsev.GetDestination(utils.META_DEFAULT)
	mp[utils.SetupTime], _ = fsev.GetSetupTime(utils.META_DEFAULT, timezone)
	mp[utils.AnswerTime], _ = fsev.GetAnswerTime(utils.META_DEFAULT, timezone)
	mp[utils.Usage], _ = fsev.GetDuration(utils.META_DEFAULT)
	mp[utils.PDD], _ = fsev.GetPdd(utils.META_DEFAULT)
	mp[utils.ACD], _ = fsev.GetADC(utils.META_DEFAULT)
	mp[utils.COST] = -1.0
	mp[utils.SUPPLIER] = fsev.GetSupplier(utils.META_DEFAULT)
	mp[utils.DISCONNECT_CAUSE] = fsev.GetDisconnectCause(utils.META_DEFAULT)
	return mp
}

// V1AuthorizeArgs returns the arguments used in SMGv1.Authorize
func (fsev FSEvent) V1AuthorizeArgs() (args *sessions.V1AuthorizeArgs) {
	cgrEv, err := fsev.AsCGREvent(config.CgrConfig().GeneralCfg().DefaultTimezone)
	if err != nil {
		return
	}
	cgrEv.Event[utils.Usage] = config.CgrConfig().SessionSCfg().MaxCallDuration // no billsec available in auth
	args = &sessions.V1AuthorizeArgs{                                           // defaults
		GetMaxUsage: true,
		CGREvent:    *cgrEv,
	}
	subsystems, has := fsev[VarCGRSubsystems]
	if !has {
		return
	}
	args.GetMaxUsage = strings.Index(subsystems, utils.MetaAccounts) != -1
	args.AuthorizeResources = strings.Index(subsystems, utils.MetaResources) != -1
	args.GetSuppliers = strings.Index(subsystems, utils.MetaSuppliers) != -1
	args.SuppliersIgnoreErrors = strings.Index(subsystems, utils.MetaSuppliersIgnoreErrors) != -1
	if strings.Index(subsystems, utils.MetaSuppliersEventCost) != -1 {
		args.SuppliersMaxCost = utils.MetaEventCost
	}
	args.GetAttributes = strings.Index(subsystems, utils.MetaAttributes) != -1
	args.ProcessThresholds = strings.Index(subsystems, utils.MetaThresholds) != -1
	args.ProcessStats = strings.Index(subsystems, utils.MetaStats) != -1
	return
}

// V1InitSessionArgs returns the arguments used in SessionSv1.InitSession
func (fsev FSEvent) V1InitSessionArgs() (args *sessions.V1InitSessionArgs) {
	cgrEv, err := fsev.AsCGREvent(config.CgrConfig().GeneralCfg().DefaultTimezone)
	if err != nil {
		return
	}
	args = &sessions.V1InitSessionArgs{ // defaults
		InitSession: true,
		CGREvent:    *cgrEv,
	}
	subsystems, has := fsev[VarCGRSubsystems]
	if !has {
		return
	}
	args.InitSession = strings.Index(subsystems, utils.MetaAccounts) != -1
	args.AllocateResources = strings.Index(subsystems, utils.MetaResources) != -1
	args.GetAttributes = strings.Index(subsystems, utils.MetaAttributes) != -1
	args.ProcessThresholds = strings.Index(subsystems, utils.MetaThresholds) != -1
	args.ProcessStats = strings.Index(subsystems, utils.MetaStats) != -1
	return
}

// V1TerminateSessionArgs returns the arguments used in SMGv1.TerminateSession
func (fsev FSEvent) V1TerminateSessionArgs() (args *sessions.V1TerminateSessionArgs) {
	cgrEv, err := fsev.AsCGREvent(config.CgrConfig().GeneralCfg().DefaultTimezone)
	if err != nil {
		return
	}
	args = &sessions.V1TerminateSessionArgs{ // defaults
		TerminateSession: true,
		CGREvent:         *cgrEv,
	}
	subsystems, has := fsev[VarCGRSubsystems]
	if !has {
		return
	}
	args.TerminateSession = strings.Index(subsystems, utils.MetaAccounts) != -1
	args.ReleaseResources = strings.Index(subsystems, utils.MetaResources) != -1
	args.ProcessThresholds = strings.Index(subsystems, utils.MetaThresholds) != -1
	args.ProcessStats = strings.Index(subsystems, utils.MetaStats) != -1
	return
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
