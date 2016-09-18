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

func NewSMAsteriskEvent(ariEv map[string]interface{}, asteriskIP string) *SMAsteriskEvent {
	smsmaEv := &SMAsteriskEvent{ariEv: ariEv, asteriskIP: asteriskIP, cachedFields: make(map[string]string)}
	smsmaEv.parseStasisArgs() // Populate appArgs
	return smsmaEv
}

type SMAsteriskEvent struct { // Standalone struct so we can cache the fields while we parse them
	ariEv        map[string]interface{} // stasis event
	asteriskIP   string
	cachedFields map[string]string // Cache replies here
}

// parseStasisArgs will convert the args passed to Stasis into CGRateS attribute/value pairs understood by CGRateS and store them in cachedFields
// args need to be in the form of []string{"key=value", "key2=value2"}
func (smaEv *SMAsteriskEvent) parseStasisArgs() {
	args, _ := smaEv.ariEv["args"].([]interface{})
	for _, arg := range args {
		if splt := strings.Split(arg.(string), "="); len(splt) > 1 {
			smaEv.cachedFields[splt[0]] = splt[1]
		}
	}
}

func (smaEv *SMAsteriskEvent) OriginatorIP() string {
	return smaEv.asteriskIP
}

func (smaEv *SMAsteriskEvent) EventType() string {
	cachedKey := eventType
	cachedVal, hasIt := smaEv.cachedFields[cachedKey]
	if !hasIt {
		cachedVal, _ = smaEv.ariEv["type"].(string)
		smaEv.cachedFields[cachedKey] = cachedVal
	}
	return cachedVal
}

func (smaEv *SMAsteriskEvent) ChannelID() string {
	cachedKey := channelID
	cachedVal, hasIt := smaEv.cachedFields[cachedKey]
	if !hasIt {
		channelData, _ := smaEv.ariEv["channel"].(map[string]interface{})
		cachedVal, _ = channelData["id"].(string)
		smaEv.cachedFields[cachedKey] = cachedVal
	}
	return cachedVal
}

func (smaEv *SMAsteriskEvent) Timestamp() string {
	cachedKey := timestamp
	cachedVal, hasIt := smaEv.cachedFields[cachedKey]
	if !hasIt {
		cachedVal, _ = smaEv.ariEv["timestamp"].(string)
		smaEv.cachedFields[cachedKey] = cachedVal
	}
	return cachedVal
}

func (smaEv *SMAsteriskEvent) ChannelState() string {
	cachedKey := channelState
	cachedVal, hasIt := smaEv.cachedFields[cachedKey]
	if !hasIt {
		channelData, _ := smaEv.ariEv["channel"].(map[string]interface{})
		cachedVal, _ = channelData["state"].(string)
	}
	return cachedVal
}

func (smaEv *SMAsteriskEvent) SetupTime() string {
	cachedKey := utils.SETUP_TIME
	cachedVal, hasIt := smaEv.cachedFields[cachedKey]
	if !hasIt {
		channelData, _ := smaEv.ariEv["channel"].(map[string]interface{})
		cachedVal, _ = channelData["creationtime"].(string)
		smaEv.cachedFields[cachedKey] = cachedVal
	}
	return cachedVal
}

func (smaEv *SMAsteriskEvent) Account() string {
	cachedKey := utils.CGR_ACCOUNT
	cachedVal, hasIt := smaEv.cachedFields[cachedKey]
	if !hasIt {
		channelData, _ := smaEv.ariEv["channel"].(map[string]interface{})
		callerData, _ := channelData["caller"].(map[string]interface{})
		cachedVal, _ = callerData["number"].(string)
		smaEv.cachedFields[cachedKey] = cachedVal
	}
	return cachedVal
}

func (smaEv *SMAsteriskEvent) Destination() string {
	cachedKey := utils.CGR_DESTINATION
	cachedVal, hasIt := smaEv.cachedFields[cachedKey]
	if !hasIt {
		channelData, _ := smaEv.ariEv["channel"].(map[string]interface{})
		dialplanData, _ := channelData["dialplan"].(map[string]interface{})
		cachedVal, _ = dialplanData["exten"].(string)
		smaEv.cachedFields[cachedKey] = cachedVal
	}
	return cachedVal
}

func (smaEv *SMAsteriskEvent) RequestType() string {
	return smaEv.cachedFields[utils.CGR_REQTYPE]
}

func (smaEv *SMAsteriskEvent) Tenant() string {
	return smaEv.cachedFields[utils.CGR_TENANT]
}

func (smaEv *SMAsteriskEvent) Category() string {
	return smaEv.cachedFields[utils.CGR_CATEGORY]
}

func (smaEv *SMAsteriskEvent) Subject() string {
	return smaEv.cachedFields[utils.CGR_SUBJECT]
}

func (smaEv *SMAsteriskEvent) PDD() string {
	return smaEv.cachedFields[utils.CGR_PDD]
}

func (smaEv *SMAsteriskEvent) Supplier() string {
	return smaEv.cachedFields[utils.CGR_SUPPLIER]
}

func (smaEv *SMAsteriskEvent) DisconnectCause() string {
	cachedKey := utils.CGR_DISCONNECT_CAUSE
	cachedVal, hasIt := smaEv.cachedFields[cachedKey]
	if !hasIt {
		cachedVal, _ = smaEv.ariEv["cause_txt"].(string)
		smaEv.cachedFields[cachedKey] = cachedVal
	}
	return cachedVal
}

func (smaEv *SMAsteriskEvent) ExtraParameters() (extraParams map[string]string) {
	extraParams = make(map[string]string)
	primaryFields := []string{eventType, channelID, timestamp, utils.SETUP_TIME, utils.CGR_ACCOUNT, utils.CGR_DESTINATION, utils.CGR_REQTYPE,
		utils.CGR_TENANT, utils.CGR_CATEGORY, utils.CGR_SUBJECT, utils.CGR_PDD, utils.CGR_SUPPLIER, utils.CGR_DISCONNECT_CAUSE}
	for cachedKey, cachedVal := range smaEv.cachedFields {
		if !utils.IsSliceMember(primaryFields, cachedKey) {
			extraParams[cachedKey] = cachedVal
		}
	}
	return
}

func (smaEv *SMAsteriskEvent) AsSMGenericEvent() *SMGenericEvent {
	var evName string
	switch smaEv.EventType() {
	case ARIStasisStart:
		evName = SMAAuthorization
	case ARIChannelStateChange:
		evName = SMASessionStart
	case ARIChannelDestroyed:
		evName = SMASessionTerminate
	}
	smgEv := SMGenericEvent{utils.EVENT_NAME: evName}
	smgEv[utils.ACCID] = smaEv.ChannelID()
	if smaEv.RequestType() != "" {
		smgEv[utils.REQTYPE] = smaEv.RequestType()
	}
	if smaEv.Tenant() != "" {
		smgEv[utils.TENANT] = smaEv.Tenant()
	}
	if smaEv.Category() != "" {
		smgEv[utils.CATEGORY] = smaEv.Category()
	}
	if smaEv.Subject() != "" {
		smgEv[utils.SUBJECT] = smaEv.Subject()
	}
	smgEv[utils.CDRHOST] = smaEv.OriginatorIP()
	smgEv[utils.ACCOUNT] = smaEv.Account()
	smgEv[utils.DESTINATION] = smaEv.Destination()
	smgEv[utils.SETUP_TIME] = smaEv.Timestamp()
	if smaEv.Supplier() != "" {
		smgEv[utils.SUPPLIER] = smaEv.Supplier()
	}
	for extraKey, extraVal := range smaEv.ExtraParameters() { // Append extraParameters
		smgEv[extraKey] = extraVal
	}
	return &smgEv
}

// Updates fields in smgEv based on own fields
// Using pointer so we update it directly in cache
func (smaEv *SMAsteriskEvent) UpdateSMGEvent(smgEv *SMGenericEvent) error {
	resSMGEv := *smgEv
	switch smaEv.EventType() {
	case ARIChannelStateChange:
		if smaEv.ChannelState() == channelUp {
			resSMGEv[utils.EVENT_NAME] = SMASessionStart
			resSMGEv[utils.ANSWER_TIME] = smaEv.Timestamp()
		}
	case ARIChannelDestroyed:
		resSMGEv[utils.EVENT_NAME] = SMASessionTerminate
		resSMGEv[utils.DISCONNECT_CAUSE] = smaEv.DisconnectCause()
		if _, hasIt := resSMGEv[utils.ANSWER_TIME]; !hasIt {
			resSMGEv[utils.USAGE] = "0s"
		} else {
			if aTime, err := smgEv.GetAnswerTime(utils.META_DEFAULT, ""); err != nil {
				return err
			} else if aTime.IsZero() {
				resSMGEv[utils.USAGE] = "0s"
			} else {
				actualTime, err := utils.ParseTimeDetectLayout(smaEv.Timestamp(), "")
				if err != nil {
					return err
				}
				resSMGEv[utils.USAGE] = actualTime.Sub(aTime).String()
			}
		}
	}
	*smgEv = resSMGEv
	return nil
}
