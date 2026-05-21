/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package agents

import (
	"errors"

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
	channelData, _ := smaEv.ariEv["channel"].(map[string]any)
	chID, _ := channelData["id"].(string)
	return chID
}

func (smaEv *SMAsteriskEvent) ChannelState() string {
	channelData, _ := smaEv.ariEv["channel"].(map[string]any)
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

func (smaEv *SMAsteriskEvent) FieldAsString(fldPath []string) (string, error) {
	val, err := smaEv.FieldAsInterface(fldPath)
	if err != nil {
		return "", err
	}
	return utils.IfaceAsString(val), nil
}
