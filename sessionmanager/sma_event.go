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
	"strings"

	"github.com/cgrates/cgrates/utils"
)

const (
	ARIStasisStart = "StasisStart"
)

func NewSMAsteriskEvent(ariEv map[string]interface{}) *SMAsteriskEvent {
	smsmaEv := &SMAsteriskEvent{ariEv: ariEv}
	smsmaEv.parseStasisArgs() // Populate appArgs
	return smsmaEv
}

type SMAsteriskEvent struct { // Standalone struct so we can cache the fields while we parse them
	ariEv   map[string]interface{} // stasis event
	appArgs map[string]string      // parsed stasis args
	// cached values start here
	evType    *string
	channelID *string
}

// parseStasisArgs will convert the args passed to Stasis into CGRateS attribute/value pairs understood by CGRateS
// args need to be in the form of []string{"key=value", "key2=value2"}
func (smaEv *SMAsteriskEvent) parseStasisArgs() {
	smaEv.appArgs = make(map[string]string)
	args := smaEv.ariEv["args"].([]interface{})
	for _, arg := range args {
		if splt := strings.Split(arg.(string), "="); len(splt) > 1 {
			smaEv.appArgs[splt[0]] = splt[1]
		}
	}
}

func (smaEv *SMAsteriskEvent) Type() string {
	if smaEv.evType == nil {
		typ, _ := smaEv.ariEv["type"].(string)
		smaEv.evType = utils.StringPointer(typ)
	}
	return *smaEv.evType
}

func (smaEv *SMAsteriskEvent) ChannelID() string {
	if smaEv.channelID == nil {
		channelData, _ := smaEv.ariEv["channel"].(map[string]interface{})
		channelID, _ := channelData["id"].(string)
		smaEv.channelID = utils.StringPointer(channelID)
	}
	return *smaEv.channelID
}

func (smaEv *SMAsteriskEvent) AsSMGenericSessionStart() (smgEv SMGenericEvent, err error) {
	smgEv = SMGenericEvent{utils.EVENT_NAME: utils.CGR_SESSION_START}
	return smgEv, nil
}
