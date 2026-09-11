// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package agents

import (
	"encoding/json"

	"github.com/cgrates/cgrates/utils"
)

const (
	EVENT                  = "event"
	CGR_SESSION_DISCONNECT = "CGR_SESSION_DISCONNECT"
	KamTRIndex             = "trIndex"
	KamTRLabel             = "trLabel"
	KamHashEntry           = "hEntry"
	KamHashID              = "hID"
	KamReplyRoute          = "replyRoute"
	KamReplyEvent          = "Event"
	KamReplyTRIndex        = "TransactionIndex"
	KamReplyTRLabel        = "TransactionLabel"
	EvapiConnID            = "evapiConnID"
	CGR_DLG_LIST           = "CGR_DLG_LIST"
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

// String is used for pretty printing event in logs
func (kev KamEvent) String() string {
	return utils.ToJSON(kev)
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
	CallId    string `json:"call-id"`
	Caller    *kamCallerDlg
	Variables []struct {
		CgrOriginID   string `json:"cgrOriginID,omitempty"`
		CgrOriginHost string `json:"cgrOriginhost,omitempty"`
	} `json:"variables"`
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
func (kev KamEvent) GetOptions() (mp map[string]any) {
	mp = make(map[string]any)
	for k := range utils.CGROptionsSet {
		if val, has := kev[k]; has {
			mp[k] = val
		}
	}
	return
}

func (kev KamEvent) setCGRRequest(nm *utils.OrderedNavigableMap, connIdx int) (err error) {
	for k, v := range kev {
		if utils.CGROptionsSet.Has(k) {
			continue
		}
		if err = nm.Set(&utils.FullPath{Path: k, PathSlice: []string{k}}, v); err != nil {
			return
		}
	}
	return nm.Set(&utils.FullPath{Path: EvapiConnID, PathSlice: []string{EvapiConnID}}, connIdx)
}
