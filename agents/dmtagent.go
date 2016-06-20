/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2015 ITsysCOM GmbH

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
	"math"
	"strconv"
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
	"github.com/fiorix/go-diameter/diam"
	"github.com/fiorix/go-diameter/diam/datatype"
	"github.com/fiorix/go-diameter/diam/sm"
)

const CCRQueueLen = 1000 // Maximum number of messages waiting to be processed

func NewDiameterAgent(cgrCfg *config.CGRConfig, smg rpcclient.RpcClientConnection, pubsubs rpcclient.RpcClientConnection) (*DiameterAgent, error) {
	da := &DiameterAgent{cgrCfg: cgrCfg, smg: smg, pubsubs: pubsubs, ccrQueue: make(chan *DiameterMessage, CCRQueueLen)}
	dictsDir := cgrCfg.DiameterAgentCfg().DictionariesDir
	if len(dictsDir) != 0 {
		if err := loadDictionaries(dictsDir, "DiameterAgent"); err != nil {
			return nil, err
		}
	}
	go da.ccrServer()
	return da, nil
}

type DiameterMessage struct {
	Message    *diam.Message
	Connection diam.Conn
}

type DiameterAgent struct {
	cgrCfg   *config.CGRConfig
	smg      rpcclient.RpcClientConnection // Connection towards CGR-SMG component
	pubsubs  rpcclient.RpcClientConnection // Connection towards CGR-PubSub component
	ccrQueue chan *DiameterMessage
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

func (self DiameterAgent) processCCR(ccr *CCR, reqProcessor *config.DARequestProcessor, processorVars map[string]string, cca *CCA) (bool, error) {
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
		*cca = *NewBareCCAFromCCR(ccr, self.cgrCfg.DiameterAgentCfg().OriginHost, self.cgrCfg.DiameterAgentCfg().OriginRealm)
	}
	smgEv, err := ccr.AsSMGenericEvent(reqProcessor.CCRFields)
	if err != nil {
		*cca = *NewBareCCAFromCCR(ccr, self.cgrCfg.DiameterAgentCfg().OriginHost, self.cgrCfg.DiameterAgentCfg().OriginRealm)
		if err := messageSetAVPsWithPath(cca.diamMessage, []interface{}{"Result-Code"}, strconv.Itoa(DiameterRatingFailed),
			false, self.cgrCfg.DiameterAgentCfg().Timezone); err != nil {
			return false, err
		}
		utils.Logger.Err(fmt.Sprintf("<DiameterAgent> Processing message: %+v AsSMGenericEvent, error: %s", ccr.diamMessage, err))
		return false, ErrDiameterRatingFailed
	}
	if len(reqProcessor.Flags) != 0 {
		smgEv[utils.CGRFlags] = reqProcessor.Flags.String() // Populate CGRFlags automatically
	}
	if reqProcessor.PublishEvent && self.pubsubs != nil {
		evt, err := smgEv.AsMapStringString()
		if err != nil {
			*cca = *NewBareCCAFromCCR(ccr, self.cgrCfg.DiameterAgentCfg().OriginHost, self.cgrCfg.DiameterAgentCfg().OriginRealm)
			if err := messageSetAVPsWithPath(cca.diamMessage, []interface{}{"Result-Code"}, strconv.Itoa(DiameterRatingFailed),
				false, self.cgrCfg.DiameterAgentCfg().Timezone); err != nil {
				return false, err
			}
			utils.Logger.Err(fmt.Sprintf("<DiameterAgent> Processing message: %+v failed converting SMGEvent to pubsub one, error: %s", ccr.diamMessage, err))
			return false, ErrDiameterRatingFailed
		}
		var reply string
		if err := self.pubsubs.Call("PubSubV1.Publish", engine.CgrEvent(evt), &reply); err != nil {
			*cca = *NewBareCCAFromCCR(ccr, self.cgrCfg.DiameterAgentCfg().OriginHost, self.cgrCfg.DiameterAgentCfg().OriginRealm)
			if err := messageSetAVPsWithPath(cca.diamMessage, []interface{}{"Result-Code"}, strconv.Itoa(DiameterRatingFailed),
				false, self.cgrCfg.DiameterAgentCfg().Timezone); err != nil {
				return false, err
			}
			utils.Logger.Err(fmt.Sprintf("<DiameterAgent> Processing message: %+v failed publishing event, error: %s", ccr.diamMessage, err))
			return false, ErrDiameterRatingFailed
		}
	}
	var maxUsage float64
	processorVars[CGRResultCode] = strconv.Itoa(diam.Success)
	processorVars[CGRError] = ""
	if reqProcessor.DryRun { // DryRun does not send over network
		utils.Logger.Info(fmt.Sprintf("<DiameterAgent> SMGenericEvent: %+v", smgEv))
		processorVars[CGRResultCode] = strconv.Itoa(diam.LimitedSuccess)
	} else { // Find out maxUsage over APIs
		switch ccr.CCRequestType {
		case 1:
			err = self.smg.Call("SMGenericV1.InitiateSession", smgEv, &maxUsage)
		case 2:
			err = self.smg.Call("SMGenericV1.UpdateSession", smgEv, &maxUsage)
		case 3, 4: // Handle them together since we generate CDR for them
			var rpl string
			if ccr.CCRequestType == 3 {
				err = self.smg.Call("SMGenericV1.TerminateSession", smgEv, &rpl)
			} else if ccr.CCRequestType == 4 {
				err = self.smg.Call("SMGenericV1.ChargeEvent", smgEv, &maxUsage)
				if maxUsage == 0 {
					smgEv[utils.USAGE] = 0 // For CDR not to debit
				}
			}
			if self.cgrCfg.DiameterAgentCfg().CreateCDR {
				if errCdr := self.smg.Call("SMGenericV1.ProcessCDR", smgEv, &rpl); errCdr != nil {
					err = errCdr
				}
			}
		}
		if err != nil {
			utils.Logger.Err(fmt.Sprintf("<DiameterAgent> Processing message: %+v, API error: %s", ccr.diamMessage, err))
			switch { // Prettify some errors
			case strings.HasSuffix(err.Error(), utils.ErrAccountNotFound.Error()):
				processorVars[CGRError] = utils.ErrAccountNotFound.Error()
			case strings.HasSuffix(err.Error(), utils.ErrUserNotFound.Error()):
				processorVars[CGRError] = utils.ErrUserNotFound.Error()
			case strings.HasSuffix(err.Error(), utils.ErrInsufficientCredit.Error()):
				processorVars[CGRError] = utils.ErrInsufficientCredit.Error()
			case strings.HasSuffix(err.Error(), utils.ErrAccountDisabled.Error()):
				processorVars[CGRError] = utils.ErrAccountDisabled.Error()
			case strings.HasSuffix(err.Error(), utils.ErrRatingPlanNotFound.Error()):
				processorVars[CGRError] = utils.ErrRatingPlanNotFound.Error()
			case strings.HasSuffix(err.Error(), utils.ErrUnauthorizedDestination.Error()):
				processorVars[CGRError] = utils.ErrUnauthorizedDestination.Error()
			default: // Unknown error
				processorVars[CGRError] = err.Error()
				processorVars[CGRResultCode] = strconv.Itoa(DiameterRatingFailed)
			}
		}
		if maxUsage < 0 {
			maxUsage = 0
		}
		if prevMaxUsageStr, hasKey := processorVars[CGRMaxUsage]; hasKey {
			prevMaxUsage, _ := strconv.ParseFloat(prevMaxUsageStr, 64)
			if prevMaxUsage < maxUsage {
				maxUsage = prevMaxUsage
			}
		}
		processorVars[CGRMaxUsage] = strconv.FormatFloat(maxUsage, 'f', -1, 64)
	}
	if err := messageSetAVPsWithPath(cca.diamMessage, []interface{}{"Result-Code"}, processorVars[CGRResultCode],
		false, self.cgrCfg.DiameterAgentCfg().Timezone); err != nil {
		return false, err
	}
	if err := cca.SetProcessorAVPs(reqProcessor, processorVars); err != nil {
		if err := messageSetAVPsWithPath(cca.diamMessage, []interface{}{"Result-Code"}, strconv.Itoa(DiameterRatingFailed),
			false, self.cgrCfg.DiameterAgentCfg().Timezone); err != nil {
			return false, err
		}
		utils.Logger.Err(fmt.Sprintf("<DiameterAgent> CCA SetProcessorAVPs for message: %+v, error: %s", ccr.diamMessage, err))
		return false, ErrDiameterRatingFailed
	}
	return true, nil
}

// Dispatches the ccr to the queue, useful so we offload the receiving buffer
func (self *DiameterAgent) handleCCR(c diam.Conn, m *diam.Message) {
	self.ccrQueue <- &DiameterMessage{Message: m, Connection: c}
}

// Asynchronously processes CCR messages
func (self *DiameterAgent) ccrServer() {
	utils.Logger.Info("<DiameterAgent> Starting the CCR Server")
	for {
		if lenCCRQueue := len(self.ccrQueue); lenCCRQueue != 0 && math.Mod(float64(lenCCRQueue), 100) == 0 {
			utils.Logger.Warning(fmt.Sprintf("<DiameterAgent> WARNING: CCR message queue loaded with %d items", lenCCRQueue))
		}
		dmtMsg := <-self.ccrQueue
		ccr, err := NewCCRFromDiameterMessage(dmtMsg.Message, self.cgrCfg.DiameterAgentCfg().DebitInterval)
		if err != nil {
			utils.Logger.Err(fmt.Sprintf("<DiameterAgent> Unmarshaling message: %s, error: %s", dmtMsg.Message, err))
			continue
		}
		cca := NewBareCCAFromCCR(ccr, self.cgrCfg.DiameterAgentCfg().OriginHost, self.cgrCfg.DiameterAgentCfg().OriginRealm)
		var processed, lclProcessed bool
		processorVars := make(map[string]string) // Shared between processors
		for _, reqProcessor := range self.cgrCfg.DiameterAgentCfg().RequestProcessors {
			lclProcessed, err = self.processCCR(ccr, reqProcessor, processorVars, cca)
			if lclProcessed { // Process local so we don't overwrite globally
				processed = lclProcessed
			}
			if err != nil || (lclProcessed && !reqProcessor.ContinueOnSuccess) {
				break
			}
		}
		if err != nil && err != ErrDiameterRatingFailed {
			utils.Logger.Err(fmt.Sprintf("<DiameterAgent> CCA SetProcessorAVPs for message: %+v, error: %s", ccr.diamMessage, err))
			continue
		} else if !processed {
			utils.Logger.Err(fmt.Sprintf("<DiameterAgent> No request processor enabled for CCR: %s, ignoring request", ccr.diamMessage))
			continue
		}
		if _, err := cca.AsDiameterMessage().WriteTo(dmtMsg.Connection); err != nil {
			utils.Logger.Err(fmt.Sprintf("<DiameterAgent> Failed to write message to %s: %s\n%s\n", dmtMsg.Connection.RemoteAddr(), err, cca.AsDiameterMessage()))
			continue
		}
	}
}

func (self *DiameterAgent) handleALL(c diam.Conn, m *diam.Message) {
	utils.Logger.Warning(fmt.Sprintf("<DiameterAgent> Received unexpected message from %s:\n%s", c.RemoteAddr(), m))
}

func (self *DiameterAgent) ListenAndServe() error {
	return diam.ListenAndServe(self.cgrCfg.DiameterAgentCfg().Listen, self.handlers(), nil)
}
