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
	varPrefix = "variable_"
	// Freswitch event proprities names
	SUBJECT                  = varPrefix + utils.CGRSubject
	ACCOUNT                  = varPrefix + utils.CGRAccount
	DESTINATION              = varPrefix + utils.CGRDestination
	REQTYPE                  = varPrefix + utils.CGRReqType //prepaid or postpaid
	CATEGORY                 = varPrefix + utils.CGRCategory
	VAR_CGR_ROUTE            = varPrefix + utils.CGRRoute
	UUID                     = "Unique-ID" // -Unique ID for this call leg
	CSTMID                   = varPrefix + utils.CGRTenant
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
	VarCGRFlags              = varPrefix + utils.CGRFlags
	VarCGROpts               = varPrefix + utils.CGROpts
	CGRResourceAllocation    = "cgr_resource_allocation"
	VAR_CGR_DISCONNECT_CAUSE = varPrefix + utils.CGRDisconnectCause
	VAR_CGR_CMPUTELCR        = varPrefix + utils.CGRComputeLCR
	FsConnID                 = "FsConnID" // used to share connID info in event for remote disconnects
	VarAnswerEpoch           = "variable_answer_epoch"
	VarCGRACD                = varPrefix + utils.CgrAcd
	VarCGROriginHost         = varPrefix + utils.CGROriginHost
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

// Account calling
func (fsev FSEvent) GetAccount(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.StaticValuePrefix) { // Static value
		return fieldName[len(utils.StaticValuePrefix):]
	}
	return utils.FirstNonEmpty(fsev[fieldName], fsev[ACCOUNT], fsev[USERNAME])
}

// Rating subject being charged
func (fsev FSEvent) GetSubject(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.StaticValuePrefix) { // Static value
		return fieldName[len(utils.StaticValuePrefix):]
	}
	return utils.FirstNonEmpty(fsev[fieldName], fsev[SUBJECT], fsev.GetAccount(fieldName))
}

// Charging destination number
func (fsev FSEvent) GetDestination(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.StaticValuePrefix) { // Static value
		return fieldName[len(utils.StaticValuePrefix):]
	}
	return utils.FirstNonEmpty(fsev[fieldName], fsev[DESTINATION],
		fsev[CALL_DEST_NR], fsev[SIP_REQ_USER])
}

// Original dialed destination number, useful in case of unpark
func (fsev FSEvent) GetCallDestNr(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.StaticValuePrefix) { // Static value
		return fieldName[len(utils.StaticValuePrefix):]
	}
	return utils.FirstNonEmpty(fsev[fieldName], fsev[CALL_DEST_NR], fsev[SIP_REQ_USER])
}
func (fsev FSEvent) GetCategory(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.StaticValuePrefix) { // Static value
		return fieldName[len(utils.StaticValuePrefix):]
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
	if strings.HasPrefix(fieldName, utils.StaticValuePrefix) { // Static value
		return fieldName[len(utils.StaticValuePrefix):]
	}
	return utils.FirstNonEmpty(fsev[fieldName], fsev[CSTMID],
		config.CgrConfig().GeneralCfg().DefaultTenant)
}

func (fsev FSEvent) GetReqType(fieldName string) string {
	var reqTypeDetected = ""                     // Used to automatically disable processing of the request
	if fsev["variable_process_cdr"] == "false" { // FS will not generated CDR here
		reqTypeDetected = utils.MetaNone
	} else if fsev["Caller-Dialplan"] == "inline" { // Used for internally generated dialplan, eg refer coming from another box, not in our control
		reqTypeDetected = utils.MetaNone
	}
	if strings.HasPrefix(fieldName, utils.StaticValuePrefix) { // Static value
		return fieldName[len(utils.StaticValuePrefix):]
	}
	return utils.FirstNonEmpty(fsev[fieldName], fsev[REQTYPE],
		reqTypeDetected, config.CgrConfig().GeneralCfg().DefaultReqType)
}

func (fsev FSEvent) MissingParameter(timezone string) string {
	if strings.TrimSpace(fsev.GetAccount(utils.MetaDefault)) == "" {
		return utils.AccountField
	}
	if strings.TrimSpace(fsev.GetSubject(utils.MetaDefault)) == "" {
		return utils.Subject
	}
	if strings.TrimSpace(fsev.GetDestination(utils.MetaDefault)) == "" {
		return utils.Destination
	}
	if strings.TrimSpace(fsev.GetCategory(utils.MetaDefault)) == "" {
		return utils.Category
	}
	if strings.TrimSpace(fsev.GetUUID()) == "" {
		return utils.OriginID
	}
	if strings.TrimSpace(fsev.GetTenant(utils.MetaDefault)) == "" {
		return utils.Tenant
	}
	if strings.TrimSpace(fsev.GetCallDestNr(utils.MetaDefault)) == "" {
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
	if strings.HasPrefix(fieldName, utils.StaticValuePrefix) { // Static value
		sTimeStr = fieldName[len(utils.StaticValuePrefix):]
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
	if strings.HasPrefix(fieldName, utils.StaticValuePrefix) { // Static value
		aTimeStr = fieldName[len(utils.StaticValuePrefix):]
	}
	return utils.ParseTimeDetectLayout(aTimeStr, timezone)
}

func (fsev FSEvent) GetEndTime(fieldName, timezone string) (t time.Time, err error) {
	return utils.ParseTimeDetectLayout(fsev[END_TIME], timezone)
}

func (fsev FSEvent) GetDuration(fieldName string) (time.Duration, error) {
	durStr := utils.FirstNonEmpty(fsev[fieldName], fsev[DURATION])
	if strings.HasPrefix(fieldName, utils.StaticValuePrefix) { // Static value
		durStr = fieldName[len(utils.StaticValuePrefix):]
	}
	return utils.ParseDurationWithSecs(durStr)
}

func (fsev FSEvent) GetPdd(fieldName string) (time.Duration, error) {
	var PDDStr string
	if utils.SliceHasMember([]string{utils.MetaDefault, utils.PDD}, fieldName) {
		PDDStr = utils.FirstNonEmpty(fsev[PDD_MEDIA_MS], fsev[PDD_NOMEDIA_MS])
		if len(PDDStr) != 0 {
			PDDStr = PDDStr + "ms" // PDD is in milliseconds and CGR expects it in seconds
		}
	} else if strings.HasPrefix(fieldName, utils.StaticValuePrefix) { // Static value
		PDDStr = fieldName[len(utils.StaticValuePrefix):]
	} else {
		PDDStr = fsev[fieldName]
	}
	return utils.ParseDurationWithSecs(PDDStr)
}

func (fsev FSEvent) GetADC(fieldName string) (time.Duration, error) {
	var ACDStr string
	if utils.SliceHasMember([]string{utils.MetaDefault, utils.ACD}, fieldName) {
		ACDStr = utils.FirstNonEmpty(fsev[VarCGRACD])
		if len(ACDStr) != 0 {
			ACDStr = ACDStr + "s" //  ACD is in seconds and CGR expects it in seconds
		}
	} else if strings.HasPrefix(fieldName, utils.StaticValuePrefix) { // Static value
		ACDStr = fieldName[len(utils.StaticValuePrefix):]
	} else {
		ACDStr = fsev[fieldName]
	}
	return utils.ParseDurationWithSecs(ACDStr)
}

func (fsev FSEvent) GetRoute(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.StaticValuePrefix) { // Static value
		return fieldName[len(utils.StaticValuePrefix):]
	}
	return utils.FirstNonEmpty(fsev[fieldName], fsev[VAR_CGR_ROUTE])
}

func (fsev FSEvent) GetDisconnectCause(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.StaticValuePrefix) { // Static value
		return fieldName[len(utils.StaticValuePrefix):]
	}
	return utils.FirstNonEmpty(fsev[fieldName], fsev[VAR_CGR_DISCONNECT_CAUSE], fsev[HANGUP_CAUSE])
}

func (fsev FSEvent) GetOriginatorIP(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.StaticValuePrefix) { // Static value
		return fieldName[len(utils.StaticValuePrefix):]
	}
	return utils.FirstNonEmpty(fsev[fieldName], fsev[VarCGROriginHost], fsev[FS_IPv4])
}

// GetOriginHost returns the first non empty between: fsev[VarCGROriginHost], conns[connId].cfg.Alias and fsev[FS_IPv4]
func (fsev FSEvent) GetOriginHost() string {
	return utils.FirstNonEmpty(fsev[VarCGROriginHost], fsev[FS_IPv4])
}

func (fsev FSEvent) GetExtraFields() map[string]string {
	extraFields := make(map[string]string)
	const dynprefix string = utils.MetaDynReq + utils.NestingSep
	for _, fldRule := range config.CgrConfig().FsAgentCfg().ExtraFields {
		if !strings.HasPrefix(fldRule.Rules, dynprefix) {
			continue
		}
		attrName := fldRule.AttrName()[5:]
		if parsed, err := fsev.ParseEventValue(attrName, fldRule,
			config.CgrConfig().GeneralCfg().DefaultTimezone); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<%s> error: %s parsing event rule: %+v", utils.FreeSWITCHAgent, err.Error(), fldRule))
		} else {
			extraFields[attrName] = parsed
		}
	}
	return extraFields
}

// Used in derived charging and sittuations when we need to run regexp on fields
func (fsev FSEvent) ParseEventValue(attrName string, rsrFld *config.RSRParser, timezone string) (parsed string, err error) {
	switch attrName {
	case utils.ToR:
		return rsrFld.ParseValue(utils.MetaVoice)
	case utils.OriginID:
		return rsrFld.ParseValue(fsev.GetUUID())
	case utils.OriginHost:
		return rsrFld.ParseValue(fsev.GetOriginHost())
	case utils.Source:
		return rsrFld.ParseValue("FS_EVENT")
	case utils.RequestType:
		return rsrFld.ParseValue(fsev.GetReqType(""))
	case utils.Tenant:
		return rsrFld.ParseValue(fsev.GetTenant(""))
	case utils.Category:
		return rsrFld.ParseValue(fsev.GetCategory(""))
	case utils.AccountField:
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
		PDD, _ := fsev.GetPdd(utils.MetaDefault)
		return rsrFld.ParseValue(strconv.FormatFloat(PDD.Seconds(), 'f', -1, 64))
	case utils.Route:
		return rsrFld.ParseValue(fsev.GetRoute(""))
	case utils.DisconnectCause:
		return rsrFld.ParseValue(fsev.GetDisconnectCause(""))
	case utils.RunID:
		return rsrFld.ParseValue(utils.MetaDefault)
	case utils.Cost:
		return rsrFld.ParseValue(strconv.FormatFloat(-1, 'f', -1, 64)) // Recommended to use FormatCost
	default:
		if parsed, err = rsrFld.ParseValue(fsev[attrName]); err != nil {
			parsed, err = rsrFld.ParseValue(fsev[FS_VARPREFIX+attrName])
		}
		return
	}
}

// AsCGREvent converts FSEvent into CGREvent
func (fsev FSEvent) AsCGREvent(timezone string) (cgrEv *utils.CGREvent, err error) {
	sTime, err := fsev.GetSetupTime(utils.MetaDefault, timezone)
	if err != nil {
		return nil, err
	}
	cgrEv = &utils.CGREvent{
		Tenant:  fsev.GetTenant(utils.MetaDefault),
		ID:      utils.UUIDSha1Prefix(),
		Time:    &sTime,
		Event:   fsev.AsMapStringInterface(timezone),
		APIOpts: fsev.GetOptions(),
	}
	return cgrEv, nil
}

// Used with RLs
func (fsev FSEvent) AsMapStringInterface(timezone string) map[string]interface{} {
	mp := make(map[string]interface{})
	for fld, val := range fsev.GetExtraFields() {
		mp[fld] = val
	}
	mp[utils.ToR] = utils.MetaVoice
	mp[utils.OriginID] = fsev.GetUUID()
	mp[utils.OriginHost] = fsev.GetOriginHost()
	mp[utils.Source] = "FS_" + fsev.GetName()
	mp[utils.RequestType] = fsev.GetReqType(utils.MetaDefault)
	mp[utils.Tenant] = fsev.GetTenant(utils.MetaDefault)
	mp[utils.Category] = fsev.GetCategory(utils.MetaDefault)
	mp[utils.AccountField] = fsev.GetAccount(utils.MetaDefault)
	mp[utils.Subject] = fsev.GetSubject(utils.MetaDefault)
	mp[utils.Destination] = fsev.GetDestination(utils.MetaDefault)
	mp[utils.SetupTime], _ = fsev.GetSetupTime(utils.MetaDefault, timezone)
	mp[utils.AnswerTime], _ = fsev.GetAnswerTime(utils.MetaDefault, timezone)
	mp[utils.Usage], _ = fsev.GetDuration(utils.MetaDefault)
	mp[utils.PDD], _ = fsev.GetPdd(utils.MetaDefault)
	mp[utils.ACD], _ = fsev.GetADC(utils.MetaDefault)
	mp[utils.Cost] = -1.0
	mp[utils.Route] = fsev.GetRoute(utils.MetaDefault)
	mp[utils.DisconnectCause] = fsev.GetDisconnectCause(utils.MetaDefault)
	return mp
}

// V1AuthorizeArgs returns the arguments used in SMGv1.Authorize
func (fsev FSEvent) V1AuthorizeArgs() (args *sessions.V1AuthorizeArgs) {
	cgrEv, err := fsev.AsCGREvent(config.CgrConfig().GeneralCfg().DefaultTimezone)
	if err != nil {
		return
	}
	cgrEv.Event[utils.Usage] = config.CgrConfig().SessionSCfg().GetDefaultUsage(utils.IfaceAsString(cgrEv.Event[utils.ToR])) // no billsec available in auth
	args = &sessions.V1AuthorizeArgs{                                                                                        // defaults
		CGREvent: cgrEv,
	}
	subsystems, has := fsev[VarCGRFlags]
	if !has {
		args.GetMaxUsage = true
		return
	}
	args.ParseFlags(subsystems)
	return
}

// V1InitSessionArgs returns the arguments used in SessionSv1.InitSession
func (fsev FSEvent) V1InitSessionArgs() (args *sessions.V1InitSessionArgs) {
	cgrEv, err := fsev.AsCGREvent(config.CgrConfig().GeneralCfg().DefaultTimezone)
	if err != nil {
		return
	}
	args = &sessions.V1InitSessionArgs{ // defaults
		CGREvent: cgrEv,
	}
	subsystems, has := fsev[VarCGRFlags]
	if !has {
		args.InitSession = true
		return
	}
	args.ParseFlags(subsystems)
	return
}

// V1TerminateSessionArgs returns the arguments used in SMGv1.TerminateSession
func (fsev FSEvent) V1TerminateSessionArgs() (args *sessions.V1TerminateSessionArgs) {
	cgrEv, err := fsev.AsCGREvent(config.CgrConfig().GeneralCfg().DefaultTimezone)
	if err != nil {
		return
	}
	args = &sessions.V1TerminateSessionArgs{ // defaults
		CGREvent: cgrEv,
	}
	subsystems, has := fsev[VarCGRFlags]
	if !has {
		args.TerminateSession = true
		return
	}
	args.ParseFlags(subsystems)
	return
}

// SliceAsFsArray Converts a slice of strings into a FS array string, contains len(array) at first index since FS does not support len(ARRAY::) for now
func SliceAsFsArray(slc []string) (arry string) {
	if len(slc) == 0 {
		return
	}
	arry = fmt.Sprintf("ARRAY::%d", len(slc))
	for _, itm := range slc {
		arry += "|:" + itm
	}
	return
}

// GetOptions returns the posible options
func (fsev FSEvent) GetOptions() (mp map[string]interface{}) {
	mp = make(map[string]interface{})
	opts, has := fsev[VarCGROpts]
	if !has {
		return
	}
	for _, opt := range strings.Split(opts, utils.FieldsSep) {
		spltOpt := strings.SplitN(opt, utils.InInFieldSep, 2)
		if len(spltOpt) != 2 {
			continue
		}
		mp[spltOpt[0]] = spltOpt[1]
	}
	return
}
