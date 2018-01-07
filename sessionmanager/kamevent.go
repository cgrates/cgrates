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
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

const (
	EVENT                  = "event"
	CGR_AUTH_REQUEST       = "CGR_AUTH_REQUEST"
	CGR_LCR_REQUEST        = "CGR_LCR_REQUEST"
	CGR_AUTH_REPLY         = "CGR_AUTH_REPLY"
	CGR_LCR_REPLY          = "CGR_LCR_REPLY"
	CGR_SESSION_DISCONNECT = "CGR_SESSION_DISCONNECT"
	CGR_CALL_START         = "CGR_CALL_START"
	CGR_CALL_END           = "CGR_CALL_END"
	CGR_RL_REQUEST         = "CGR_RL_REQUEST"
	CGR_RL_REPLY           = "CGR_RL_REPLY"
	CGR_SETUPTIME          = "cgr_setuptime"
	CGR_ANSWERTIME         = "cgr_answertime"
	CGR_STOPTIME           = "cgr_stoptime"
	CGR_DURATION           = "cgr_duration"
	CGR_PDD                = "cgr_pdd"

	KAM_TR_INDEX = "tr_index"
	KAM_TR_LABEL = "tr_label"
	HASH_ENTRY   = "h_entry"
	HASH_ID      = "h_id"
)

var primaryFields = []string{EVENT, CALLID, FROM_TAG, HASH_ENTRY, HASH_ID, CGR_ACCOUNT, CGR_SUBJECT, CGR_DESTINATION,
	CGR_CATEGORY, CGR_TENANT, CGR_REQTYPE, CGR_ANSWERTIME, CGR_SETUPTIME, CGR_STOPTIME, CGR_DURATION, CGR_PDD, utils.CGR_SUPPLIER, utils.CGR_DISCONNECT_CAUSE}

type KamAuthReply struct {
	Event             string // Kamailio will use this to differentiate between requests and replies
	TransactionIndex  int    // Original transaction index
	TransactionLabel  int    // Original transaction label
	MaxSessionTime    int    // Maximum session time in case of success, -1 for unlimited
	Suppliers         string // List of suppliers, comma separated
	ResourceAllocated bool
	AllocationMessage string
	Error             string // Reply in case of error
}

func (self *KamAuthReply) String() string {
	mrsh, _ := json.Marshal(self)
	return string(mrsh)
}

type KamLcrReply struct {
	Event     string
	Suppliers string
	Error     error
}

func (self *KamLcrReply) String() string {
	self.Event = CGR_LCR_REPLY
	mrsh, _ := json.Marshal(self)
	return string(mrsh)
}

type KamSessionDisconnect struct {
	Event     string
	HashEntry string
	HashId    string
	Reason    string
}

func (self *KamSessionDisconnect) String() string {
	mrsh, _ := json.Marshal(self)
	return string(mrsh)
}

func NewKamEvent(kamEvData []byte) (KamEvent, error) {
	kev := make(map[string]string)
	if err := json.Unmarshal(kamEvData, &kev); err != nil {
		return nil, err
	}
	return kev, nil
}

// Hold events received from Kamailio
type KamEvent map[string]string

// Backwards compatibility, should be AsEvent
func (kev KamEvent) AsEvent(ignored string) engine.Event {
	return engine.Event(kev)
}

func (kev KamEvent) GetName() string {
	return kev[EVENT]
}
func (kev KamEvent) GetCgrId(timezone string) string {
	setupTime, _ := kev.GetSetupTime(utils.META_DEFAULT, timezone)
	return utils.Sha1(kev.GetUUID(), setupTime.UTC().String())
}
func (kev KamEvent) GetUUID() string {
	return kev[CALLID] + ";" + kev[FROM_TAG] // ToTag not available in callStart event
}
func (kev KamEvent) GetSessionIds() []string {
	return []string{kev[HASH_ENTRY], kev[HASH_ID]}
}
func (kev KamEvent) GetDirection(fieldName string) string {
	return utils.OUT
}
func (kev KamEvent) GetAccount(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.FirstNonEmpty(kev[fieldName], kev[CGR_ACCOUNT])
}

func (kev KamEvent) GetSubject(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.FirstNonEmpty(kev[fieldName], kev[CGR_SUBJECT], kev.GetAccount(fieldName))
}
func (kev KamEvent) GetDestination(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.FirstNonEmpty(kev[fieldName], kev[CGR_DESTINATION])
}
func (kev KamEvent) GetCallDestNr(fieldName string) string {
	return kev.GetDestination(utils.META_DEFAULT)
}
func (kev KamEvent) GetCategory(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.FirstNonEmpty(kev[fieldName], kev[CGR_CATEGORY], config.CgrConfig().DefaultCategory)
}
func (kev KamEvent) GetTenant(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.FirstNonEmpty(kev[fieldName], kev[CGR_TENANT], config.CgrConfig().DefaultTenant)
}
func (kev KamEvent) GetReqType(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.FirstNonEmpty(kev[fieldName], kev[CGR_REQTYPE], config.CgrConfig().DefaultReqType)
}
func (kev KamEvent) GetAnswerTime(fieldName, timezone string) (time.Time, error) {
	aTimeStr := utils.FirstNonEmpty(kev[fieldName], kev[CGR_ANSWERTIME])
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		aTimeStr = fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.ParseTimeDetectLayout(aTimeStr, timezone)
}
func (kev KamEvent) GetSetupTime(fieldName, timezone string) (time.Time, error) {
	sTimeStr := utils.FirstNonEmpty(kev[fieldName], kev[CGR_SETUPTIME], kev[CGR_ANSWERTIME])
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		sTimeStr = fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.ParseTimeDetectLayout(sTimeStr, timezone)
}
func (kev KamEvent) GetEndTime(fieldName, timezone string) (time.Time, error) {
	return utils.ParseTimeDetectLayout(kev[CGR_STOPTIME], timezone)
}
func (kev KamEvent) GetDuration(fieldName string) (time.Duration, error) {
	durStr := utils.FirstNonEmpty(kev[fieldName], kev[CGR_DURATION])
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		durStr = fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.ParseDurationWithSecs(durStr)
}
func (kev KamEvent) GetPdd(fieldName string) (time.Duration, error) {
	var pddStr string
	if utils.IsSliceMember([]string{utils.PDD, utils.META_DEFAULT}, fieldName) {
		pddStr = kev[CGR_PDD]
	} else if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		pddStr = fieldName[len(utils.STATIC_VALUE_PREFIX):]
	} else {
		pddStr = kev[fieldName]
	}
	return utils.ParseDurationWithSecs(pddStr)
}
func (kev KamEvent) GetSupplier(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.FirstNonEmpty(kev[fieldName], kev[utils.CGR_SUPPLIER])
}

func (kev KamEvent) GetDisconnectCause(fieldName string) string {
	if strings.HasPrefix(fieldName, utils.STATIC_VALUE_PREFIX) { // Static value
		return fieldName[len(utils.STATIC_VALUE_PREFIX):]
	}
	return utils.FirstNonEmpty(kev[fieldName], kev[utils.CGR_DISCONNECT_CAUSE])
}

//ToDo: extract the IP of the kamailio server generating the event
func (kev KamEvent) GetOriginatorIP(string) string {
	return "127.0.0.1"
}
func (kev KamEvent) GetExtraFields() map[string]string {
	extraFields := make(map[string]string)
	for field, val := range kev {
		if !utils.IsSliceMember(primaryFields, field) {
			extraFields[field] = val
		}
	}
	return extraFields
}
func (kev KamEvent) GetCdrSource() string {
	return "KAMAILIO_" + kev.GetName()
}

func (kev KamEvent) MissingParameter(timezone string) bool {
	var nullTime time.Time
	switch kev.GetName() {
	case CGR_AUTH_REQUEST:
		if setupTime, err := kev.GetSetupTime(utils.META_DEFAULT, timezone); err != nil || setupTime == nullTime {
			return true
		}
		return len(kev.GetAccount(utils.META_DEFAULT)) == 0 ||
			len(kev.GetDestination(utils.META_DEFAULT)) == 0 ||
			len(kev[KAM_TR_INDEX]) == 0 || len(kev[KAM_TR_LABEL]) == 0
	case CGR_LCR_REQUEST:
		return len(kev.GetAccount(utils.META_DEFAULT)) == 0 ||
			len(kev.GetDestination(utils.META_DEFAULT)) == 0 ||
			len(kev[KAM_TR_INDEX]) == 0 || len(kev[KAM_TR_LABEL]) == 0
	case CGR_CALL_START:
		if aTime, err := kev.GetAnswerTime(utils.META_DEFAULT, timezone); err != nil || aTime == nullTime {
			return true
		}
		return len(kev.GetUUID()) == 0 ||
			len(kev.GetAccount(utils.META_DEFAULT)) == 0 ||
			len(kev.GetDestination(utils.META_DEFAULT)) == 0 ||
			len(kev[HASH_ENTRY]) == 0 || len(kev[HASH_ID]) == 0
	case CGR_CALL_END:
		return len(kev.GetUUID()) == 0 ||
			len(kev.GetAccount(utils.META_DEFAULT)) == 0 ||
			len(kev.GetDestination(utils.META_DEFAULT)) == 0 ||
			len(kev[CGR_DURATION]) == 0
	default:
		return true
	}

}

// Useful for CDR generation
func (kev KamEvent) ParseEventValue(rsrFld *utils.RSRField, timezone string) string {
	sTime, _ := kev.GetSetupTime(utils.META_DEFAULT, config.CgrConfig().DefaultTimezone)
	aTime, _ := kev.GetAnswerTime(utils.META_DEFAULT, config.CgrConfig().DefaultTimezone)
	duration, _ := kev.GetDuration(utils.META_DEFAULT)
	switch rsrFld.Id {
	case utils.CGRID:
		return rsrFld.ParseValue(kev.GetCgrId(timezone))
	case utils.TOR:
		return rsrFld.ParseValue(utils.VOICE)
	case utils.OriginID:
		return rsrFld.ParseValue(kev.GetUUID())
	case utils.OriginHost:
		return rsrFld.ParseValue(kev.GetOriginatorIP(utils.META_DEFAULT))
	case utils.Source:
		return rsrFld.ParseValue(kev.GetCdrSource())
	case utils.RequestType:
		return rsrFld.ParseValue(kev.GetReqType(utils.META_DEFAULT))
	case utils.Direction:
		return rsrFld.ParseValue(kev.GetDirection(utils.META_DEFAULT))
	case utils.Tenant:
		return rsrFld.ParseValue(kev.GetTenant(utils.META_DEFAULT))
	case utils.Category:
		return rsrFld.ParseValue(kev.GetCategory(utils.META_DEFAULT))
	case utils.Account:
		return rsrFld.ParseValue(kev.GetAccount(utils.META_DEFAULT))
	case utils.Subject:
		return rsrFld.ParseValue(kev.GetSubject(utils.META_DEFAULT))
	case utils.Destination:
		return rsrFld.ParseValue(kev.GetDestination(utils.META_DEFAULT))
	case utils.SetupTime:
		return rsrFld.ParseValue(sTime.String())
	case utils.AnswerTime:
		return rsrFld.ParseValue(aTime.String())
	case utils.Usage:
		return rsrFld.ParseValue(strconv.FormatFloat(utils.Round(duration.Seconds(), 0, utils.ROUNDING_MIDDLE), 'f', -1, 64))
	case utils.PDD:
		return rsrFld.ParseValue(strconv.FormatFloat(utils.Round(duration.Seconds(), 0, utils.ROUNDING_MIDDLE), 'f', -1, 64))
	case utils.SUPPLIER:
		return rsrFld.ParseValue(kev.GetSupplier(utils.META_DEFAULT))
	case utils.DISCONNECT_CAUSE:
		return rsrFld.ParseValue(kev.GetDisconnectCause(utils.META_DEFAULT))
	case utils.MEDI_RUNID:
		return rsrFld.ParseValue(utils.META_DEFAULT)
	case utils.COST:
		return rsrFld.ParseValue("-1.0")
	default:
		return rsrFld.ParseValue(kev.GetExtraFields()[rsrFld.Id])
	}
}
func (kev KamEvent) PassesFieldFilter(*utils.RSRField) (bool, string) {
	return false, ""
}

func (kev KamEvent) AsCDR(timezone string) *engine.CDR {
	storCdr := new(engine.CDR)
	storCdr.CGRID = kev.GetCgrId(timezone)
	storCdr.ToR = utils.VOICE
	storCdr.OriginID = kev.GetUUID()
	storCdr.OriginHost = kev.GetOriginatorIP(utils.META_DEFAULT)
	storCdr.Source = kev.GetCdrSource()
	storCdr.RequestType = kev.GetReqType(utils.META_DEFAULT)
	storCdr.Tenant = kev.GetTenant(utils.META_DEFAULT)
	storCdr.Category = kev.GetCategory(utils.META_DEFAULT)
	storCdr.Account = kev.GetAccount(utils.META_DEFAULT)
	storCdr.Subject = kev.GetSubject(utils.META_DEFAULT)
	storCdr.Destination = kev.GetDestination(utils.META_DEFAULT)
	storCdr.SetupTime, _ = kev.GetSetupTime(utils.META_DEFAULT, timezone)
	storCdr.AnswerTime, _ = kev.GetAnswerTime(utils.META_DEFAULT, timezone)
	storCdr.Usage, _ = kev.GetDuration(utils.META_DEFAULT)
	storCdr.ExtraFields = kev.GetExtraFields()
	storCdr.Cost = -1

	return storCdr
}

func (kev KamEvent) String() string {
	mrsh, _ := json.Marshal(kev)
	return string(mrsh)
}

func (kev KamEvent) AsKamAuthReply(maxSessionTime float64, suppliers string,
	resAllocated bool, allocationMessage string, rplyErr error) (kar *KamAuthReply, err error) {
	kar = &KamAuthReply{Event: CGR_AUTH_REPLY, Suppliers: suppliers,
		ResourceAllocated: resAllocated, AllocationMessage: allocationMessage}
	if rplyErr != nil {
		kar.Error = rplyErr.Error()
	}
	if _, hasIt := kev[KAM_TR_INDEX]; !hasIt {
		return nil, utils.NewErrMandatoryIeMissing(KAM_TR_INDEX, "")
	}
	if kar.TransactionIndex, err = strconv.Atoi(kev[KAM_TR_INDEX]); err != nil {
		return nil, err
	}
	if _, hasIt := kev[KAM_TR_LABEL]; !hasIt {
		return nil, utils.NewErrMandatoryIeMissing(KAM_TR_LABEL, "")
	}
	if kar.TransactionLabel, err = strconv.Atoi(kev[KAM_TR_LABEL]); err != nil {
		return nil, err
	}
	if maxSessionTime != -1 { // Convert maxSessionTime from nanoseconds into seconds
		maxSessionDur := time.Duration(maxSessionTime)
		maxSessionTime = maxSessionDur.Seconds()
	}
	kar.MaxSessionTime = int(utils.Round(maxSessionTime, 0, utils.ROUNDING_MIDDLE))

	return kar, nil
}

// Converts into CallDescriptor due to responder interface needs
func (kev KamEvent) AsCallDescriptor() (*engine.CallDescriptor, error) {
	lcrReq := &engine.LcrRequest{
		Direction:   kev.GetDirection(utils.META_DEFAULT),
		Tenant:      kev.GetTenant(utils.META_DEFAULT),
		Category:    kev.GetCategory(utils.META_DEFAULT),
		Account:     kev.GetAccount(utils.META_DEFAULT),
		Subject:     kev.GetSubject(utils.META_DEFAULT),
		Destination: kev.GetDestination(utils.META_DEFAULT),
		SetupTime:   utils.FirstNonEmpty(kev[CGR_SETUPTIME], kev[CGR_ANSWERTIME]),
		Duration:    kev[CGR_DURATION],
	}
	return lcrReq.AsCallDescriptor(config.CgrConfig().DefaultTimezone)
}

func (kev KamEvent) ComputeLcr() bool {
	if computeLcr, err := strconv.ParseBool(kev[utils.CGR_COMPUTELCR]); err != nil {
		return false
	} else {
		return computeLcr
	}
}

func (kev KamEvent) AsMapStringIface() (mp map[string]interface{}, err error) {
	mp = make(map[string]interface{}, len(kev))
	for k, v := range kev {
		mp[k] = v
	}
	return
}
