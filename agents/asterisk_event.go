// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package agents

import (
	"errors"
	"strings"

	"github.com/cgrates/cgrates/utils"
)

func NewSMAsteriskEvent(ariEv map[string]any) *SMAsteriskEvent {
	smaEv := &SMAsteriskEvent{
		ariEv: ariEv,
	}
	return smaEv
}

type SMAsteriskEvent struct {
	ariEv map[string]any // raw stasis/ARI event
}

// EventType returns the ARI event type (StasisStart, ChannelStateChange, ...).
func (smaEv *SMAsteriskEvent) EventType() string {
	evType, _ := smaEv.ariEv["type"].(string)
	return evType
}

// ChannelID returns the id of the channel the event refers to.
func (smaEv *SMAsteriskEvent) ChannelID() string {
	channelData, _ := smaEv.ariEv[ARIChannel].(map[string]any)
	chID, _ := channelData["id"].(string)
	return chID
}

func (smaEv *SMAsteriskEvent) ChannelState() string {
	channelData, _ := smaEv.ariEv[ARIChannel].(map[string]any)
	state, _ := channelData["state"].(string)
	return state
}

func (smaEv *SMAsteriskEvent) String() string {
	return utils.ToJSON(smaEv.ariEv)
}

func (smaEv *SMAsteriskEvent) FieldAsInterface(fldPath []string) (any, error) {
	if len(fldPath) == 0 {
		return nil, errors.New("empty field path")
	}
	return utils.MapStorage(smaEv.ariEv).FieldAsInterface(fldPath)
}

func (smaEv *SMAsteriskEvent) setCGRRequest(nm *utils.OrderedNavigableMap, opts utils.MapStorage) (err error) {
	chnl, _ := smaEv.ariEv[ARIChannel].(map[string]any)
	chVars, _ := chnl[ARIChannelVars].(map[string]any)
	for k, v := range chVars {
		if strings.HasPrefix(k, utils.Meta) {
			opts[k] = v
		}
	}
	for k, v := range smaEv.ariEv {
		if err = nm.Set(&utils.FullPath{Path: k, PathSlice: []string{k}}, v); err != nil {
			return
		}
	}
	return
}

func (smaEv *SMAsteriskEvent) FieldAsString(fldPath []string) (string, error) {
	val, err := smaEv.FieldAsInterface(fldPath)
	if err != nil {
		return "", err
	}
	return utils.IfaceAsString(val), nil
}
