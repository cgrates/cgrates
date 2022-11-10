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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

const (
	EVENT                  = "event"
	CGR_AUTH_REQUEST       = "CGR_AUTH_REQUEST"
	CGR_AUTH_REPLY         = "CGR_AUTH_REPLY"
	CGR_SESSION_DISCONNECT = "CGR_SESSION_DISCONNECT"
	CGR_CALL_START         = "CGR_CALL_START"
	CGR_CALL_END           = "CGR_CALL_END"
	CGR_PROCESS_MESSAGE    = "CGR_PROCESS_MESSAGE"
	CGR_PROCESS_CDR        = "CGR_PROCESS_CDR"
	KamTRIndex             = "tr_index"
	KamTRLabel             = "tr_label"
	KamHashEntry           = "h_entry"
	KamHashID              = "h_id"
	KamReplyRoute          = "reply_route"
	EvapiConnID            = "EvapiConnID" // used to share connID info in event for remote disconnects
	CGR_DLG_LIST           = "CGR_DLG_LIST"
)

var (
	kamReservedEventFields = utils.NewStringSet([]string{EVENT, KamTRIndex, KamTRLabel, utils.CGRFlags, KamReplyRoute})
	// kamReservedCDRFields   = append(kamReservedEventFields, KamHashEntry, KamHashID) // HashEntry and id are needed in events for disconnects
)

func NewKamSessionDisconnect(hEntry, hID, reason string) *KamSessionDisconnect {
	return &KamSessionDisconnect{
		Event:     CGR_SESSION_DISCONNECT,
		HashEntry: hEntry,
		HashId:    hID,
		Reason:    reason}
}

type KamSessionDisconnect struct {
	Event     string
	HashEntry string
	HashId    string
	Reason    string
}

func (ksd *KamSessionDisconnect) String() string {
	return utils.ToJSON(ksd)
}

// NewKamEvent parses bytes received over the wire from Kamailio into KamEvent
func NewKamEvent(kamEvData []byte, alias, adress string) (KamEvent, error) {
	kev := make(map[string]string)
	if err := json.Unmarshal(kamEvData, &kev); err != nil {
		return nil, err
	}
	kev[utils.OriginHost] = utils.FirstNonEmpty(kev[utils.OriginHost], alias, adress)
	return kev, nil
}

// KamEvent represents one event received from Kamailio
type KamEvent map[string]string

func (kev KamEvent) MissingParameter() bool {
	switch kev[EVENT] {
	case CGR_AUTH_REQUEST:
		return utils.IsSliceMember([]string{
			kev[KamTRIndex],
			kev[KamTRLabel],
		}, "")
	case CGR_CALL_START:
		return utils.IsSliceMember([]string{
			kev[KamHashEntry],
			kev[KamHashID],
			kev[utils.OriginID],
			kev[utils.AnswerTime],
			kev[utils.AccountField],
			kev[utils.Destination],
		}, "")
	case CGR_CALL_END:
		return utils.IsSliceMember([]string{
			kev[utils.OriginID],
			kev[utils.AnswerTime],
			kev[utils.AccountField],
			kev[utils.Destination],
		}, "")
	case CGR_PROCESS_MESSAGE:
		// TRIndex and TRLabel must exist in order to know where to send back the response
		mndPrm := []string{kev[KamTRIndex], kev[KamTRLabel]}
		_, has := kev[utils.CGRFlags]
		// in case that the user populate cgr_flags we treat it like a ProcessEvent
		// and expect to have the required fields
		if has {
			mndPrm = append(mndPrm, kev[utils.OriginID],
				kev[utils.AnswerTime],
				kev[utils.AccountField],
				kev[utils.Destination])
		}
		return utils.IsSliceMember(mndPrm, "")
	case CGR_PROCESS_CDR:
		// TRIndex and TRLabel must exist in order to know where to send back the response
		return utils.IsSliceMember([]string{
			kev[KamTRIndex],
			kev[KamTRLabel],
			kev[utils.OriginID],
		}, "")
	default: // no/unsupported event
		return true
	}

}

// AsMapStringInterface converts KamEvent into event used by other subsystems
func (kev KamEvent) AsMapStringInterface() (mp map[string]interface{}) {
	mp = make(map[string]interface{})
	for k, v := range kev {
		if k == utils.Usage {
			v += "s" // mark the Usage as seconds
		}
		if !kamReservedEventFields.Has(k) && // reserved attributes not getting into event
			!utils.CGROptionsSet.Has(k) { // also omit the options
			mp[k] = v
		}
	}
	if _, has := mp[utils.Source]; !has {
		mp[utils.Source] = utils.KamailioAgent
	}
	if _, has := mp[utils.RequestType]; !has {
		mp[utils.RequestType] = config.CgrConfig().GeneralCfg().DefaultReqType
	}
	return
}

// AsCGREvent converts KamEvent into CGREvent
func (kev KamEvent) AsCGREvent(timezone string) (cgrEv *utils.CGREvent, err error) {
	var sTime time.Time
	switch kev[EVENT] {
	case CGR_AUTH_REQUEST:
		sTimePrv, err := utils.ParseTimeDetectLayout(kev[utils.SetupTime], timezone)
		if err != nil {
			return nil, err
		}
		sTime = sTimePrv
	case CGR_CALL_START:
		sTimePrv, err := utils.ParseTimeDetectLayout(kev[utils.AnswerTime], timezone)
		if err != nil {
			return nil, err
		}
		sTime = sTimePrv
	case CGR_CALL_END:
		sTimePrv, err := utils.ParseTimeDetectLayout(kev[utils.AnswerTime], timezone)
		if err != nil {
			return nil, err
		}
		sTime = sTimePrv
	case CGR_PROCESS_MESSAGE:
		sTimePrv, err := utils.ParseTimeDetectLayout(kev[utils.AnswerTime], timezone)
		if err != nil {
			return nil, err
		}
		sTime = sTimePrv
	case CGR_PROCESS_CDR:
		sTime = time.Now()
	default: // no/unsupported event
		return
	}
	cgrEv = &utils.CGREvent{
		Tenant: utils.FirstNonEmpty(kev[utils.Tenant],
			config.CgrConfig().GeneralCfg().DefaultTenant),
		ID:      utils.UUIDSha1Prefix(),
		Time:    &sTime,
		Event:   kev.AsMapStringInterface(),
		APIOpts: kev.GetOptions(),
	}
	return cgrEv, nil
}

// String is used for pretty printing event in logs
func (kev KamEvent) String() string {
	return utils.ToJSON(kev)
}

// V1AuthorizeArgs returns the arguments used in SessionSv1.AuthorizeEvent
func (kev KamEvent) V1AuthorizeArgs() (args *sessions.V1AuthorizeArgs) {
	cgrEv, err := kev.AsCGREvent(config.CgrConfig().GeneralCfg().DefaultTimezone)
	if err != nil {
		return
	}
	args = &sessions.V1AuthorizeArgs{
		CGREvent: cgrEv,
	}
	subsystems, has := kev[utils.CGRFlags]
	if !has {
		utils.Logger.Warning(fmt.Sprintf("<%s> cgr_flags variable is not set, using defaults",
			utils.KamailioAgent))
		args.GetMaxUsage = true
		return
	}
	args.ParseFlags(subsystems, utils.InfieldSep)
	return
}

// AsKamAuthReply builds up a Kamailio AuthReply based on arguments and reply from SessionS
func (kev KamEvent) AsKamAuthReply(authArgs *sessions.V1AuthorizeArgs,
	authReply *sessions.V1AuthorizeReply, rplyErr error) (kar *KamReply, err error) {
	evName := CGR_AUTH_REPLY
	if kamRouReply, has := kev[KamReplyRoute]; has {
		evName = kamRouReply
	}
	kar = &KamReply{Event: evName,
		TransactionIndex: kev[KamTRIndex],
		TransactionLabel: kev[KamTRLabel],
	}
	if rplyErr != nil {
		kar.Error = rplyErr.Error()
		return
	}
	if authArgs.GetAttributes && authReply.Attributes != nil {
		kar.Attributes = authReply.Attributes.Digest()
	}
	if authArgs.AuthorizeResources && authReply.ResourceAllocation != nil {
		kar.ResourceAllocation = *authReply.ResourceAllocation
	}
	if authArgs.GetMaxUsage {
		if authReply.MaxUsage != nil {
			kar.MaxUsage = int(utils.Round(authReply.MaxUsage.Seconds(), 0, utils.MetaRoundingMiddle))
		} else {
			kar.MaxUsage = 0
		}
	}
	if authArgs.GetRoutes && authReply.RouteProfiles != nil {
		kar.Routes = authReply.RouteProfiles.Digest()
	}

	if authArgs.ProcessThresholds && authReply.ThresholdIDs != nil {
		kar.Thresholds = strings.Join(*authReply.ThresholdIDs, utils.FieldsSep)
	}
	if authArgs.ProcessStats && authReply.StatQueueIDs != nil {
		kar.StatQueues = strings.Join(*authReply.StatQueueIDs, utils.FieldsSep)
	}
	return
}

// V1InitSessionArgs returns the arguments used in SessionSv1.InitSession
func (kev KamEvent) V1InitSessionArgs() (args *sessions.V1InitSessionArgs) {
	cgrEv, err := kev.AsCGREvent(config.CgrConfig().GeneralCfg().DefaultTimezone)
	if err != nil {
		return
	}
	args = &sessions.V1InitSessionArgs{ // defaults

		CGREvent: cgrEv,
	}
	subsystems, has := kev[utils.CGRFlags]
	if !has {
		utils.Logger.Warning(fmt.Sprintf("<%s> cgr_flags is not set, using defaults",
			utils.FreeSWITCHAgent))
		args.InitSession = true
		return
	}
	args.ParseFlags(subsystems, utils.InfieldSep)
	return
}

// V1ProcessMessageArgs returns the arguments used in SessionSv1.ProcessMessage
func (kev KamEvent) V1ProcessMessageArgs() (args *sessions.V1ProcessMessageArgs) {
	cgrEv, err := kev.AsCGREvent(config.CgrConfig().GeneralCfg().DefaultTimezone)
	if err != nil {
		return
	}
	args = &sessions.V1ProcessMessageArgs{ // defaults

		CGREvent: cgrEv,
	}
	subsystems, has := kev[utils.CGRFlags]
	if !has {
		utils.Logger.Warning(fmt.Sprintf("<%s> cgr_flags is not set, using defaults",
			utils.FreeSWITCHAgent))
		return
	}
	args.ParseFlags(subsystems, utils.InfieldSep)
	return
}

// V1ProcessCDRArgs returns the arguments used in SessionSv1.ProcessCDR
func (kev KamEvent) V1ProcessCDRArgs() (args *utils.CGREvent) {
	var err error
	if args, err = kev.AsCGREvent(config.CgrConfig().GeneralCfg().DefaultTimezone); err != nil {
		return
	}
	return
}

// AsKamProcessMessageReply builds up a Kamailio ProcessEvent based on arguments and reply from SessionS
func (kev KamEvent) AsKamProcessMessageReply(procEvArgs *sessions.V1ProcessMessageArgs,
	procEvReply *sessions.V1ProcessMessageReply, rplyErr error) (kar *KamReply, err error) {
	evName := CGR_PROCESS_MESSAGE
	if kamRouReply, has := kev[KamReplyRoute]; has {
		evName = kamRouReply
	}
	kar = &KamReply{Event: evName,
		TransactionIndex: kev[KamTRIndex],
		TransactionLabel: kev[KamTRLabel],
	}
	if rplyErr != nil {
		kar.Error = rplyErr.Error()
		return
	}
	if procEvArgs.GetAttributes && procEvReply.Attributes != nil {
		kar.Attributes = procEvReply.Attributes.Digest()
	}
	if procEvArgs.AllocateResources {
		kar.ResourceAllocation = *procEvReply.ResourceAllocation
	}
	if procEvArgs.Debit {
		kar.MaxUsage = int(utils.Round(procEvReply.MaxUsage.Seconds(), 0, utils.MetaRoundingMiddle))
	}
	if procEvArgs.GetRoutes && procEvReply.RouteProfiles != nil {
		kar.Routes = procEvReply.RouteProfiles.Digest()
	}

	if procEvArgs.ProcessThresholds {
		kar.Thresholds = strings.Join(*procEvReply.ThresholdIDs, utils.FieldsSep)
	}
	if procEvArgs.ProcessStats {
		kar.StatQueues = strings.Join(*procEvReply.StatQueueIDs, utils.FieldsSep)
	}
	return
}

// AsKamProcessCDRReply builds up a Kamailio ProcessEvent based on arguments and reply from SessionS
func (kev KamEvent) AsKamProcessCDRReply(cgrEvWithArgDisp *utils.CGREvent,
	rply *string, rplyErr error) (kar *KamReply, err error) {
	evName := CGR_PROCESS_CDR
	if kamRouReply, has := kev[KamReplyRoute]; has {
		evName = kamRouReply
	}
	kar = &KamReply{Event: evName,
		TransactionIndex: kev[KamTRIndex],
		TransactionLabel: kev[KamTRLabel],
	}
	if rplyErr != nil {
		kar.Error = rplyErr.Error()
	}
	return
}

// AsKamProcessMessageEmptyReply builds up a Kamailio ProcessEventEmpty
func (kev KamEvent) AsKamProcessMessageEmptyReply() (kar *KamReply) {
	evName := CGR_PROCESS_MESSAGE
	if kamRouReply, has := kev[KamReplyRoute]; has {
		evName = kamRouReply
	}
	kar = &KamReply{Event: evName,
		TransactionIndex: kev[KamTRIndex],
		TransactionLabel: kev[KamTRLabel],
	}
	return
}

// V1TerminateSessionArgs returns the arguments used in SMGv1.TerminateSession
func (kev KamEvent) V1TerminateSessionArgs() (args *sessions.V1TerminateSessionArgs) {
	cgrEv, err := kev.AsCGREvent(utils.FirstNonEmpty(
		config.CgrConfig().KamAgentCfg().Timezone,
		config.CgrConfig().GeneralCfg().DefaultTimezone))
	if err != nil {
		return
	}
	args = &sessions.V1TerminateSessionArgs{ // defaults
		TerminateSession: true,
		CGREvent:         cgrEv,
	}
	subsystems, has := kev[utils.CGRFlags]
	if !has {
		utils.Logger.Warning(fmt.Sprintf("<%s> cgr_flags is not set, using defaults",
			utils.FreeSWITCHAgent))
		return
	}
	args.ParseFlags(subsystems, utils.InfieldSep)
	return
}

// KamReply will be used to send back to kamailio from
// Authrization,ProcessEvent and ProcessEvent empty (pingPong)
type KamReply struct {
	Event              string // Kamailio will use this to differentiate between requests and replies
	TransactionIndex   string // Original transaction index
	TransactionLabel   string // Original transaction label
	Attributes         string
	ResourceAllocation string
	MaxUsage           int    // Maximum session time in case of success, -1 for unlimited
	Routes             string // List of routes, comma separated
	Thresholds         string
	StatQueues         string
	Error              string // Reply in case of error
}

func (krply *KamReply) String() string {
	return utils.ToJSON(krply)
}

type KamDlgReply struct {
	Event        string
	Jsonrpl_body *kamJsonDlgBody
}

type kamJsonDlgBody struct {
	Id      int
	Jsonrpc string
	Result  []*kamDlgInfo
}

type kamDlgInfo struct {
	CallId string `json:"call-id"`
	Caller *kamCallerDlg
}

type kamCallerDlg struct {
	Tag string
}

// NewKamDlgReply parses bytes received over the wire from Kamailio into KamDlgReply
func NewKamDlgReply(kamEvData []byte) (rpl KamDlgReply, err error) {
	if err = json.Unmarshal(kamEvData, &rpl); err != nil {
		return
	}
	return
}

func (kdr *KamDlgReply) String() string {
	return utils.ToJSON(kdr)
}

// GetOptions returns the posible options
func (kev KamEvent) GetOptions() (mp map[string]interface{}) {
	mp = make(map[string]interface{})
	for k := range utils.CGROptionsSet {
		if val, has := kev[k]; has {
			mp[k] = val
		}
	}
	return
}
