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
	"github.com/cgrates/cgrates/utils"
)

const (
	ARIStasisStart = "StasisStart"
)

func NewARIEvent(ev map[string]interface{}) *ARIEvent {
	return &ARIEvent{ev: ev}
}

type ARIEvent struct { // Standalone struct so we can cache the fields while we parse them
	ev        map[string]interface{}
	evType    *string
	channelID *string
}

func (aev *ARIEvent) Type() string {
	if aev.evType == nil {
		typ, _ := aev.ev["type"].(string)
		aev.evType = utils.StringPointer(typ)
	}
	return *aev.evType
}

func (aev ARIEvent) ChannelID() string {
	if aev.channelID == nil {
		channelData, _ := aev.ev["channel"].(map[string]interface{})
		channelID, _ := channelData["id"].(string)
		aev.channelID = utils.StringPointer(channelID)
	}
	return *aev.channelID
}

func (aev ARIEvent) AsSMGenericSessionStart() (smgEv SMGenericEvent, err error) {
	smgEv = SMGenericEvent{utils.EVENT_NAME: utils.CGR_SESSION_START}
	return smgEv, nil
}
