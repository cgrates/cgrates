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
	KamTRIndex             = "tr_index"
	KamTRLabel             = "tr_label"
	KamHashEntry           = "h_entry"
	KamHashID              = "h_id"
	KamReplyRoute          = "reply_route"
	KamCGRContext          = "cgr_context"
	EvapiConnID            = "EvapiConnID" // used to share connID info in event for remote disconnects
	CGR_DLG_LIST           = "CGR_DLG_LIST"
)

var (
	kamReservedEventFields = []string{EVENT, KamTRIndex, KamTRLabel, utils.CGRSubsystems, KamCGRContext, KamReplyRoute}
	kamReservedCDRFields   = append(kamReservedEventFields, KamHashEntry, KamHashID) // HashEntry and id are needed in events for disconnects
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

func (self *KamSessionDisconnect) String() string {
	mrsh, _ := json.Marshal(self)
	return string(mrsh)
}

// NewKamEvent parses bytes received over the wire from Kamailio into KamEvent
func NewKamEvent(kamEvData []byte) (KamEvent, error) {
	kev := make(map[string]string)
	if err := json.Unmarshal(kamEvData, &kev); err != nil {
		return nil, err
	}
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
			kev[utils.Account],
			kev[utils.Destination],
		}, "")
	case CGR_CALL_END:
		return utils.IsSliceMember([]string{
			kev[utils.OriginID],
			kev[utils.AnswerTime],
			kev[utils.Account],
			kev[utils.Destination],
		}, "")
	default: // no/unsupported event
		return true
	}

}

// AsMapStringIface converts KamEvent into event used by other subsystems
func (kev KamEvent) AsMapStringInterface() (mp map[string]interface{}) {
	mp = make(map[string]interface{})
	for k, v := range kev {
		if k == utils.Usage {
			v += "s" // mark the Usage as seconds
		}
		if !utils.IsSliceMember(kamReservedEventFields, k) { // reserved attributes not getting into event
			mp[k] = v
		}
	}
	if _, has := mp[utils.Source]; !has {
		mp[utils.Source] = utils.KamailioAgent
	}
	return
}

// AsCDR converts KamEvent into CGREvent
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
	default: // no/unsupported event
		return
	}
	cgrEv = &utils.CGREvent{
		Tenant: utils.FirstNonEmpty(kev[utils.Tenant],
			config.CgrConfig().GeneralCfg().DefaultTenant),
		ID:    utils.UUIDSha1Prefix(),
		Time:  &sTime,
		Event: kev.AsMapStringInterface(),
	}
	if ctx, has := kev[KamCGRContext]; has {
		cgrEv.Context = utils.StringPointer(ctx)
	}
	return cgrEv, nil
}

// String is used for pretty printing event in logs
func (kev KamEvent) String() string {
	mrsh, _ := json.Marshal(kev)
	return string(mrsh)
}

func (kev KamEvent) V1AuthorizeArgs() (args *sessions.V1AuthorizeArgs) {
	cgrEv, err := kev.AsCGREvent(config.CgrConfig().GeneralCfg().DefaultTimezone)
	if err != nil {
		return
	}
	args = &sessions.V1AuthorizeArgs{
		GetMaxUsage: true,
		CGREvent:    *cgrEv,
	}
	subsystems, has := kev[utils.CGRSubsystems]
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

// AsKamAuthReply builds up a Kamailio AuthReply based on arguments and reply from SessionS
func (kev KamEvent) AsKamAuthReply(authArgs *sessions.V1AuthorizeArgs,
	authReply *sessions.V1AuthorizeReply, rplyErr error) (kar *KamAuthReply, err error) {
	evName := CGR_AUTH_REPLY
	if kamRouReply, has := kev[KamReplyRoute]; has {
		evName = kamRouReply
	}
	kar = &KamAuthReply{Event: evName,
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
	if authArgs.AuthorizeResources {
		kar.ResourceAllocation = *authReply.ResourceAllocation
	}
	if authArgs.GetMaxUsage {
		if *authReply.MaxUsage == -1 { // For calls different than unlimited, set limits
			kar.MaxUsage = -1
		} else {
			kar.MaxUsage = int(utils.Round(authReply.MaxUsage.Seconds(), 0, utils.ROUNDING_MIDDLE))
		}
	}
	if authArgs.GetSuppliers && authReply.Suppliers != nil {
		kar.Suppliers = authReply.Suppliers.Digest()
	}

	if authArgs.ProcessThresholds {
		kar.Thresholds = strings.Join(*authReply.ThresholdIDs, utils.FIELDS_SEP)
	}
	if authArgs.ProcessStats {
		kar.StatQueues = strings.Join(*authReply.StatQueueIDs, utils.FIELDS_SEP)
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
		InitSession: true,
		CGREvent:    *cgrEv,
	}
	subsystems, has := kev[utils.CGRSubsystems]
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
func (kev KamEvent) V1TerminateSessionArgs() (args *sessions.V1TerminateSessionArgs) {
	cgrEv, err := kev.AsCGREvent(config.CgrConfig().GeneralCfg().DefaultTimezone)
	if err != nil {
		return
	}
	args = &sessions.V1TerminateSessionArgs{ // defaults
		TerminateSession: true,
		CGREvent:         *cgrEv,
	}
	subsystems, has := kev[utils.CGRSubsystems]
	if !has {
		return
	}
	args.TerminateSession = strings.Index(subsystems, utils.MetaAccounts) != -1
	args.ReleaseResources = strings.Index(subsystems, utils.MetaResources) != -1
	args.ProcessThresholds = strings.Index(subsystems, utils.MetaThresholds) != -1
	args.ProcessStats = strings.Index(subsystems, utils.MetaStats) != -1
	return
}

type KamAuthReply struct {
	Event              string // Kamailio will use this to differentiate between requests and replies
	TransactionIndex   string // Original transaction index
	TransactionLabel   string // Original transaction label
	Attributes         string
	ResourceAllocation string
	MaxUsage           int    // Maximum session time in case of success, -1 for unlimited
	Suppliers          string // List of suppliers, comma separated
	Thresholds         string
	StatQueues         string
	Error              string // Reply in case of error
}

func (self *KamAuthReply) String() string {
	mrsh, _ := json.Marshal(self)
	return string(mrsh)
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

func (self *KamDlgReply) String() string {
	mrsh, _ := json.Marshal(self)
	return string(mrsh)
}
