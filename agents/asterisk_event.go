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
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/sessions"
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
	cachedKey := utils.SetupTime
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

func (smaEv *SMAsteriskEvent) Subsystems() string {
	return smaEv.cachedFields[utils.CGRSubsystems]
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
	primaryFields := []string{eventType, channelID, timestamp, utils.SetupTime, utils.CGR_ACCOUNT, utils.CGR_DESTINATION, utils.CGR_REQTYPE,
		utils.CGR_TENANT, utils.CGR_CATEGORY, utils.CGR_SUBJECT, utils.CGR_PDD, utils.CGR_SUPPLIER, utils.CGR_DISCONNECT_CAUSE}
	for cachedKey, cachedVal := range smaEv.cachedFields {
		if !utils.IsSliceMember(primaryFields, cachedKey) {
			extraParams[cachedKey] = cachedVal
		}
	}
	return
}

func (smaEv *SMAsteriskEvent) UpdateCGREvent(cgrEv *utils.CGREvent) error {
	resCGREv := *cgrEv
	switch smaEv.EventType() {
	case ARIChannelStateChange:
		resCGREv.Event[utils.EVENT_NAME] = SMASessionStart
		resCGREv.Event[utils.AnswerTime] = smaEv.Timestamp()
	case ARIChannelDestroyed:
		resCGREv.Event[utils.EVENT_NAME] = SMASessionTerminate
		resCGREv.Event[utils.DISCONNECT_CAUSE] = smaEv.DisconnectCause()
		if _, hasIt := resCGREv.Event[utils.AnswerTime]; !hasIt {
			resCGREv.Event[utils.Usage] = "0s"
		} else {
			if aTime, err := utils.IfaceAsTime(resCGREv.Event[utils.AnswerTime],
				config.CgrConfig().GeneralCfg().DefaultTimezone); err != nil {
				return err
			} else if aTime.IsZero() {
				resCGREv.Event[utils.Usage] = "0s"
			} else {
				actualTime, err := utils.ParseTimeDetectLayout(smaEv.Timestamp(), "")
				if err != nil {
					return err
				}
				resCGREv.Event[utils.Usage] = actualTime.Sub(aTime).String()
			}
		}
	}
	*cgrEv = resCGREv
	return nil
}

func (smaEv *SMAsteriskEvent) AsMapStringInterface() (mp map[string]interface{}) {
	mp = make(map[string]interface{})
	var evName string
	switch smaEv.EventType() {
	case ARIStasisStart:
		evName = SMAAuthorization
	case ARIChannelStateChange:
		evName = SMASessionStart
	case ARIChannelDestroyed:
		evName = SMASessionTerminate
	}
	mp[utils.EVENT_NAME] = evName
	mp[utils.OriginID] = smaEv.ChannelID()
	if smaEv.RequestType() != "" {
		mp[utils.RequestType] = smaEv.RequestType()
	}
	if smaEv.Tenant() != "" {
		mp[utils.Tenant] = smaEv.Tenant()
	}
	if smaEv.Category() != "" {
		mp[utils.Category] = smaEv.Category()
	}
	if smaEv.Subject() != "" {
		mp[utils.Subject] = smaEv.Subject()
	}
	mp[utils.OriginHost] = smaEv.OriginatorIP()
	mp[utils.Account] = smaEv.Account()
	mp[utils.Destination] = smaEv.Destination()
	mp[utils.SetupTime] = smaEv.SetupTime()
	if smaEv.Supplier() != "" {
		mp[utils.SUPPLIER] = smaEv.Supplier()
	}
	for extraKey, extraVal := range smaEv.ExtraParameters() { // Append extraParameters
		mp[extraKey] = extraVal
	}
	mp[utils.Source] = utils.AsteriskAgent
	return
}

// AsCDR converts AsteriskEvent into CGREvent
func (smaEv *SMAsteriskEvent) AsCGREvent(timezone string) (cgrEv *utils.CGREvent, err error) {
	setupTime, err := utils.ParseTimeDetectLayout(
		smaEv.Timestamp(), timezone)
	if err != nil {
		return
	}
	cgrEv = &utils.CGREvent{
		Tenant: utils.FirstNonEmpty(smaEv.Tenant(),
			config.CgrConfig().GeneralCfg().DefaultTenant),
		ID:    utils.UUIDSha1Prefix(),
		Time:  &setupTime,
		Event: smaEv.AsMapStringInterface(),
	}
	return cgrEv, nil
}

func (smaEv *SMAsteriskEvent) V1AuthorizeArgs() (args *sessions.V1AuthorizeArgs) {
	cgrEv, err := smaEv.AsCGREvent(config.CgrConfig().GeneralCfg().DefaultTimezone)
	if err != nil {
		return
	}
	args = &sessions.V1AuthorizeArgs{
		GetMaxUsage: true,
		CGREvent:    *cgrEv,
	}
	if smaEv.Subsystems() == utils.EmptyString {
		utils.Logger.Err(fmt.Sprintf("<%s> cgr_subsystems variable is not set",
			utils.AsteriskAgent))
		return
	}
	args.GetMaxUsage = strings.Index(smaEv.Subsystems(), utils.MetaAccounts) != -1
	args.AuthorizeResources = strings.Index(smaEv.Subsystems(), utils.MetaResources) != -1
	args.GetSuppliers = strings.Index(smaEv.Subsystems(), utils.MetaSuppliers) != -1
	args.SuppliersIgnoreErrors = strings.Index(smaEv.Subsystems(), utils.MetaSuppliersIgnoreErrors) != -1
	if strings.Index(smaEv.Subsystems(), utils.MetaSuppliersEventCost) != -1 {
		args.SuppliersMaxCost = utils.MetaEventCost
	}
	args.GetAttributes = strings.Index(smaEv.Subsystems(), utils.MetaAttributes) != -1
	args.ProcessThresholds = strings.Index(smaEv.Subsystems(), utils.MetaThresholds) != -1
	args.ProcessStats = strings.Index(smaEv.Subsystems(), utils.MetaStats) != -1
	return
}

func (smaEv *SMAsteriskEvent) V1InitSessionArgs(cgrEv utils.CGREvent) (args *sessions.V1InitSessionArgs) {
	args = &sessions.V1InitSessionArgs{ // defaults
		InitSession: true,
		CGREvent:    cgrEv,
	}
	subsystems, err := cgrEv.FieldAsString(utils.CGRSubsystems)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> event: %s don't have cgr_subsystems variable",
			utils.AsteriskAgent, utils.ToJSON(cgrEv)))
		return
	}
	args.InitSession = strings.Index(subsystems, utils.MetaAccounts) != -1
	args.AllocateResources = strings.Index(subsystems, utils.MetaResources) != -1
	args.GetAttributes = strings.Index(subsystems, utils.MetaAttributes) != -1
	args.ProcessThresholds = strings.Index(subsystems, utils.MetaThresholds) != -1
	args.ProcessStats = strings.Index(subsystems, utils.MetaStats) != -1
	return
}

func (smaEv *SMAsteriskEvent) V1TerminateSessionArgs(cgrEv utils.CGREvent) (args *sessions.V1TerminateSessionArgs) {
	args = &sessions.V1TerminateSessionArgs{ // defaults
		TerminateSession: true,
		CGREvent:         cgrEv,
	}
	subsystems, err := cgrEv.FieldAsString(utils.CGRSubsystems)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> event: %s don't have cgr_subsystems variable",
			utils.AsteriskAgent, utils.ToJSON(cgrEv)))
		return
	}
	args.TerminateSession = strings.Index(subsystems, utils.MetaAccounts) != -1
	args.ReleaseResources = strings.Index(subsystems, utils.MetaResources) != -1
	args.ProcessThresholds = strings.Index(subsystems, utils.MetaThresholds) != -1
	args.ProcessStats = strings.Index(subsystems, utils.MetaStats) != -1
	return
}
