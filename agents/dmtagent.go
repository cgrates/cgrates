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
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
	"github.com/fiorix/go-diameter/diam"
	"github.com/fiorix/go-diameter/diam/datatype"
	"github.com/fiorix/go-diameter/diam/sm"
)

func NewDiameterAgent(cgrCfg *config.CGRConfig, sessionS rpcclient.RpcClientConnection,
	pubsubs rpcclient.RpcClientConnection) (*DiameterAgent, error) {
	da := &DiameterAgent{cgrCfg: cgrCfg, sessionS: sessionS,
		pubsubs: pubsubs, connMux: new(sync.Mutex)}
	if reflect.ValueOf(da.pubsubs).IsNil() {
		da.pubsubs = nil // Empty it so we can check it later
	}
	dictsDir := cgrCfg.DiameterAgentCfg().DictionariesDir
	if len(dictsDir) != 0 {
		if err := loadDictionaries(dictsDir, "DiameterAgent"); err != nil {
			return nil, err
		}
	}
	return da, nil
}

type DiameterAgent struct {
	cgrCfg   *config.CGRConfig
	sessionS rpcclient.RpcClientConnection // Connection towards CGR-SMG component
	pubsubs  rpcclient.RpcClientConnection // Connection towards CGR-PubSub component
	connMux  *sync.Mutex                   // Protect connection for read/write
}

// Creates the message handlers
func (self *DiameterAgent) handlers() diam.Handler {
	settings := &sm.Settings{
		OriginHost:       datatype.DiameterIdentity(self.cgrCfg.DiameterAgentCfg().OriginHost),
		OriginRealm:      datatype.DiameterIdentity(self.cgrCfg.DiameterAgentCfg().OriginRealm),
		VendorID:         datatype.Unsigned32(self.cgrCfg.DiameterAgentCfg().VendorId),
		ProductName:      datatype.UTF8String(self.cgrCfg.DiameterAgentCfg().ProductName),
		FirmwareRevision: datatype.Unsigned32(utils.DIAMETER_FIRMWARE_REVISION),
	}
	dSM := sm.New(settings)
	dSM.HandleFunc("CCR", self.handleCCR)
	dSM.HandleFunc("ALL", self.handleALL)
	go func() {
		for err := range dSM.ErrorReports() {
			utils.Logger.Err(fmt.Sprintf("<DiameterAgent> StateMachine error: %+v", err))
		}
	}()
	return dSM
}

func (da DiameterAgent) processCCR(ccr *CCR, reqProcessor *config.DARequestProcessor,
	procVars processorVars, cca *CCA) (processed bool, err error) {
	passesAllFilters := true
	for _, fldFilter := range reqProcessor.RequestFilter {
		if passes, _ := passesFieldFilter(ccr.diamMessage, fldFilter, nil); !passes {
			passesAllFilters = false
		}
	}
	if !passesAllFilters { // Not going with this processor further
		return false, nil
	}
	if reqProcessor.DryRun { // DryRun should log the matching processor as well as the received CCR
		utils.Logger.Info(fmt.Sprintf("<DiameterAgent> RequestProcessor: %s", reqProcessor.Id))
		utils.Logger.Info(fmt.Sprintf("<DiameterAgent> CCR message: %s", ccr.diamMessage))
	}
	if !reqProcessor.AppendCCA {
		*cca = *NewBareCCAFromCCR(ccr, da.cgrCfg.DiameterAgentCfg().OriginHost, da.cgrCfg.DiameterAgentCfg().OriginRealm)
		procVars = make(processorVars)
	}
	smgEv, err := ccr.AsSMGenericEvent(reqProcessor.CCRFields)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<DiameterAgent> Processing message: %+v AsSMGenericEvent, error: %s", ccr.diamMessage, err))
		*cca = *NewBareCCAFromCCR(ccr, da.cgrCfg.DiameterAgentCfg().OriginHost, da.cgrCfg.DiameterAgentCfg().OriginRealm)
		if err := messageSetAVPsWithPath(cca.diamMessage, []interface{}{"Result-Code"}, strconv.Itoa(DiameterRatingFailed),
			false, da.cgrCfg.DiameterAgentCfg().Timezone); err != nil {
			utils.Logger.Err(fmt.Sprintf("<DiameterAgent> Processing message: %+v messageSetAVPsWithPath, error: %s", cca.diamMessage, err.Error()))
			return false, err
		}
		return false, ErrDiameterRatingFailed
	}
	if len(reqProcessor.Flags) != 0 {
		smgEv[utils.CGRFlags] = reqProcessor.Flags.String() // Populate CGRFlags automatically
		for flag, val := range reqProcessor.Flags {
			procVars[flag] = val
		}
	}
	if reqProcessor.PublishEvent && da.pubsubs != nil {
		evt, err := smgEv.AsMapStringString()
		if err != nil {
			*cca = *NewBareCCAFromCCR(ccr, da.cgrCfg.DiameterAgentCfg().OriginHost, da.cgrCfg.DiameterAgentCfg().OriginRealm)
			if err := messageSetAVPsWithPath(cca.diamMessage, []interface{}{"Result-Code"}, strconv.Itoa(DiameterRatingFailed),
				false, da.cgrCfg.DiameterAgentCfg().Timezone); err != nil {
				return false, err
			}
			utils.Logger.Err(fmt.Sprintf("<DiameterAgent> Processing message: %+v failed converting SMGEvent to pubsub one, error: %s", ccr.diamMessage, err))
			return false, ErrDiameterRatingFailed
		}
		var reply string
		if err := da.pubsubs.Call("PubSubV1.Publish", engine.CgrEvent(evt), &reply); err != nil {
			*cca = *NewBareCCAFromCCR(ccr, da.cgrCfg.DiameterAgentCfg().OriginHost, da.cgrCfg.DiameterAgentCfg().OriginRealm)
			if err := messageSetAVPsWithPath(cca.diamMessage, []interface{}{"Result-Code"}, strconv.Itoa(DiameterRatingFailed),
				false, da.cgrCfg.DiameterAgentCfg().Timezone); err != nil {
				return false, err
			}
			utils.Logger.Err(fmt.Sprintf("<DiameterAgent> Processing message: %+v failed publishing event, error: %s", ccr.diamMessage, err))
			return false, ErrDiameterRatingFailed
		}
	}
	if reqProcessor.DryRun { // DryRun does not send over network
		utils.Logger.Info(fmt.Sprintf("<DiameterAgent> SMGenericEvent: %+v", smgEv))
		procVars[CGRResultCode] = strconv.Itoa(diam.LimitedSuccess)
	} else { // Query SessionS over APIs
		var tnt string
		if tntIf, has := smgEv[utils.Tenant]; has {
			if tntStr, canCast := utils.CastFieldIfToString(tntIf); canCast {
				tnt = tntStr
			}
		}
		cgrEv := &utils.CGREvent{
			Tenant: utils.FirstNonEmpty(tnt,
				config.CgrConfig().DefaultTenant),
			ID:    "dmt:" + utils.UUIDSha1Prefix(),
			Time:  utils.TimePointer(time.Now()),
			Event: smgEv,
		}
		switch ccr.CCRequestType {
		case 1:
			var initReply sessions.V1InitSessionReply
			err = da.sessionS.Call(utils.SessionSv1InitiateSession,
				procVars.asV1InitSessionArgs(cgrEv), &initReply)
			if procVars[utils.MetaCGRReply], err = utils.NewCGRReply(&initReply, err); err != nil {
				return
			}
		case 2:
			var updateReply sessions.V1UpdateSessionReply
			err = da.sessionS.Call(utils.SessionSv1UpdateSession,
				procVars.asV1UpdateSessionArgs(cgrEv), &updateReply)
			if procVars[utils.MetaCGRReply], err = utils.NewCGRReply(&updateReply, err); err != nil {
				return
			}
		case 3, 4: // Handle them together since we generate CDR for them
			var rpl string
			if ccr.CCRequestType == 3 {
				if err = da.sessionS.Call(utils.SessionSv1TerminateSession,
					procVars.asV1TerminateSessionArgs(cgrEv), &rpl); err != nil {
					procVars[utils.MetaCGRReply] = utils.CGRReply{utils.Error: err.Error()}
				}
			} else if ccr.CCRequestType == 4 {
				var evntRply sessions.V1ProcessEventReply
				err = da.sessionS.Call(utils.SessionSv1ProcessEvent,
					procVars.asV1ProcessEventArgs(cgrEv), &evntRply)
				if procVars[utils.MetaCGRReply], err = utils.NewCGRReply(&evntRply, err); err != nil {
					return
				}
			}
			if da.cgrCfg.DiameterAgentCfg().CreateCDR &&
				(!da.cgrCfg.DiameterAgentCfg().CDRRequiresSession || err == nil ||
					!strings.HasSuffix(err.Error(), utils.ErrNoActiveSession.Error())) { // Check if CDR requires session
				if errCdr := da.sessionS.Call(utils.SessionSv1ProcessCDR, *cgrEv, &rpl); errCdr != nil {
					err = errCdr
					procVars[utils.MetaCGRReply] = utils.CGRReply{utils.Error: err.Error()}
				}
			}
		}
	}
	diamCode := strconv.Itoa(diam.Success)
	if procVars.hasVar(CGRResultCode) {
		diamCode, _ = procVars.valAsString(CGRResultCode)
	}
	if err := messageSetAVPsWithPath(cca.diamMessage, []interface{}{"Result-Code"}, diamCode,
		false, da.cgrCfg.DiameterAgentCfg().Timezone); err != nil {
		return false, err
	}
	if err := cca.SetProcessorAVPs(reqProcessor, procVars); err != nil {
		if err := messageSetAVPsWithPath(cca.diamMessage, []interface{}{"Result-Code"}, strconv.Itoa(DiameterRatingFailed),
			false, da.cgrCfg.DiameterAgentCfg().Timezone); err != nil {
			return false, err
		}
		utils.Logger.Err(fmt.Sprintf("<DiameterAgent> CCA SetProcessorAVPs for message: %+v, error: %s", ccr.diamMessage, err))
		return false, ErrDiameterRatingFailed
	}
	if reqProcessor.DryRun {
		utils.Logger.Info(fmt.Sprintf("<DiameterAgent> CCA message: %s", cca.diamMessage))
	}
	return true, nil
}

func (self *DiameterAgent) handlerCCR(c diam.Conn, m *diam.Message) {
	ccr, err := NewCCRFromDiameterMessage(m, self.cgrCfg.DiameterAgentCfg().DebitInterval)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<DiameterAgent> Unmarshaling message: %s, error: %s", m, err))
		return
	}
	cca := NewBareCCAFromCCR(ccr, self.cgrCfg.DiameterAgentCfg().OriginHost, self.cgrCfg.DiameterAgentCfg().OriginRealm)
	var processed, lclProcessed bool
	procVars := make(processorVars) // Shared between processors
	for _, reqProcessor := range self.cgrCfg.DiameterAgentCfg().RequestProcessors {
		lclProcessed, err = self.processCCR(ccr, reqProcessor, procVars, cca)
		if lclProcessed { // Process local so we don't overwrite globally
			processed = lclProcessed
		}
		if err != nil || (lclProcessed && !reqProcessor.ContinueOnSuccess) {
			break
		}
	}
	if err != nil && err != ErrDiameterRatingFailed {
		utils.Logger.Err(fmt.Sprintf("<DiameterAgent> CCA SetProcessorAVPs for message: %+v, error: %s", ccr.diamMessage, err))
		return
	} else if !processed {
		utils.Logger.Err(fmt.Sprintf("<DiameterAgent> No request processor enabled for CCR: %s, ignoring request", ccr.diamMessage))
		return
	}
	self.connMux.Lock()
	defer self.connMux.Unlock()
	if _, err := cca.AsDiameterMessage().WriteTo(c); err != nil {
		utils.Logger.Err(fmt.Sprintf("<DiameterAgent> Failed to write message to %s: %s\n%s\n", c.RemoteAddr(), err, cca.AsDiameterMessage()))
		return
	}
}

// Simply dispatch the handling in goroutines
// Could be futher improved with rate control
func (self *DiameterAgent) handleCCR(c diam.Conn, m *diam.Message) {
	go self.handlerCCR(c, m)
}

func (self *DiameterAgent) handleALL(c diam.Conn, m *diam.Message) {
	utils.Logger.Warning(fmt.Sprintf("<DiameterAgent> Received unexpected message from %s:\n%s", c.RemoteAddr(), m))
}

func (self *DiameterAgent) ListenAndServe() error {
	return diam.ListenAndServe(self.cgrCfg.DiameterAgentCfg().Listen, self.handlers(), nil)
}
