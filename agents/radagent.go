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
	"strconv"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/radigo"
	"github.com/cgrates/rpcclient"
)

const (
	MetaRadReplyCode    = "*radReplyCode"
	MetaRadAuth         = "*radAuthReq"
	MetaRadAcctStart    = "*radAcctStart"
	MetaRadAcctUpdate   = "*radAcctUpdate"
	MetaRadAcctStop     = "*radAcctStop"
	MetaRadAcctEvent    = "*radAcctEvent"
	MetaCGRReply        = "*cgrReply"
	MetaCGRMaxUsage     = "*cgrMaxUsage"
	MetaCGRError        = "*cgrError"
	MetaRadReqType      = "*radReqType"
	EvRadiusReq         = "RADIUS_REQUEST"
	MetaUsageDifference = "*usage_difference"
)

func NewRadiusAgent(cgrCfg *config.CGRConfig, smg rpcclient.RpcClientConnection) (ra *RadiusAgent, err error) {
	dicts := make(map[string]*radigo.Dictionary, len(cgrCfg.RadiusAgentCfg().ClientDictionaries))
	for clntID, dictPath := range cgrCfg.RadiusAgentCfg().ClientDictionaries {
		if dicts[clntID], err = radigo.NewDictionaryFromFolderWithRFC2865(dictPath); err != nil {
			return
		}
	}
	ra = &RadiusAgent{cgrCfg: cgrCfg, smg: smg}
	ra.rsAuth = radigo.NewServer(cgrCfg.RadiusAgentCfg().ListenNet,
		cgrCfg.RadiusAgentCfg().ListenAuth, cgrCfg.RadiusAgentCfg().ClientSecrets, dicts,
		map[radigo.PacketCode]func(*radigo.Packet) (*radigo.Packet, error){
			radigo.AccessRequest: ra.handleAuth}, nil)
	ra.rsAcct = radigo.NewServer(cgrCfg.RadiusAgentCfg().ListenNet,
		cgrCfg.RadiusAgentCfg().ListenAcct, cgrCfg.RadiusAgentCfg().ClientSecrets, dicts,
		map[radigo.PacketCode]func(*radigo.Packet) (*radigo.Packet, error){
			radigo.AccountingRequest: ra.handleAcct}, nil)
	return

}

type RadiusAgent struct {
	cgrCfg *config.CGRConfig             // reference for future config reloads
	smg    rpcclient.RpcClientConnection // Connection towards CGR-SMG component
	rsAuth *radigo.Server
	rsAcct *radigo.Server
}

// handleAuth handles RADIUS Authorization request
func (ra *RadiusAgent) handleAuth(req *radigo.Packet) (rpl *radigo.Packet, err error) {
	req.SetAVPValues() // populate string values in AVPs
	procVars := map[string]string{
		MetaRadReqType: MetaRadAuth,
	}
	rpl = req.Reply()
	rpl.Code = radigo.AccessAccept
	var processed bool
	for _, reqProcessor := range ra.cgrCfg.RadiusAgentCfg().RequestProcessors {
		var lclProcessed bool
		if lclProcessed, err = ra.processRequest(reqProcessor, req, procVars, rpl); lclProcessed {
			processed = lclProcessed
		}
		if err != nil || (lclProcessed && !reqProcessor.ContinueOnSuccess) {
			break
		}
	}
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<RadiusAgent> error: <%s> ignoring request: %s, process vars: %+v",
			err.Error(), utils.ToJSON(req), procVars))
		return nil, nil
	} else if !processed {
		utils.Logger.Err(fmt.Sprintf("<RadiusAgent> No request processor enabled, ignoring request %s, process vars: %+v",
			utils.ToJSON(req), procVars))
		return nil, nil
	}
	return
}

// handleAcct handles RADIUS Accounting request
// supports: Acct-Status-Type = Start, Interim-Update, Stop
func (ra *RadiusAgent) handleAcct(req *radigo.Packet) (rpl *radigo.Packet, err error) {
	req.SetAVPValues() // populate string values in AVPs
	procVars := make(map[string]string)
	if avps := req.AttributesWithName("Acct-Status-Type", ""); len(avps) != 0 { // populate accounting type
		switch avps[0].GetStringValue() { // first AVP found will give out the type of accounting
		case "Start":
			procVars[MetaRadReqType] = MetaRadAcctStart
		case "Interim-Update":
			procVars[MetaRadReqType] = MetaRadAcctUpdate
		case "Stop":
			procVars[MetaRadReqType] = MetaRadAcctStop
		}
	}
	rpl = req.Reply()
	rpl.Code = radigo.AccountingResponse
	var processed bool
	for _, reqProcessor := range ra.cgrCfg.RadiusAgentCfg().RequestProcessors {
		var lclProcessed bool
		if lclProcessed, err = ra.processRequest(reqProcessor, req, procVars, rpl); lclProcessed {
			processed = lclProcessed
		}
		if err != nil || (lclProcessed && !reqProcessor.ContinueOnSuccess) {
			break
		}
	}
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<RadiusAgent> error: <%s> ignoring request: %s, process vars: %+v",
			err.Error(), utils.ToJSON(req), procVars))
		return nil, nil
	} else if !processed {
		utils.Logger.Err(fmt.Sprintf("<RadiusAgent> No request processor enabled, ignoring request %s, process vars: %+v",
			utils.ToJSON(req), procVars))
		return nil, nil
	}
	return
}

// processRequest represents one processor processing the request
func (ra *RadiusAgent) processRequest(reqProcessor *config.RARequestProcessor,
	req *radigo.Packet, processorVars map[string]string, reply *radigo.Packet) (processed bool, err error) {
	passesAllFilters := true
	for _, fldFilter := range reqProcessor.RequestFilter {
		if !radPassesFieldFilter(req, processorVars, fldFilter) {
			passesAllFilters = false
			break
		}
	}
	if !passesAllFilters { // Not going with this processor further
		return false, nil
	}
	for k, v := range reqProcessor.Flags { // update processorVars with flags from processor
		processorVars[k] = strconv.FormatBool(v)
	}
	if reqProcessor.DryRun {
		utils.Logger.Info(fmt.Sprintf("<RadiusAgent> DRY_RUN, RADIUS request: %s", utils.ToJSON(req)))
		utils.Logger.Info(fmt.Sprintf("<RadiusAgent> DRY_RUN, process variabiles: %+v", processorVars))
	}
	smgEv, err := radReqAsSMGEvent(req, processorVars, reqProcessor.Flags, reqProcessor.RequestFields)
	if err != nil {
		return false, err
	}
	if reqProcessor.DryRun {
		utils.Logger.Info(fmt.Sprintf("<RadiusAgent> DRY_RUN, SMGEvent: %+v", smgEv))
	} else { // process with RPC
		var maxUsage time.Duration
		var cgrReply interface{} // so we can store it in processorsVars
		switch processorVars[MetaRadReqType] {
		case MetaRadAuth: // auth attempt, make sure that MaxUsage is enough
			if err = ra.smg.Call("SMGenericV2.GetMaxUsage", smgEv, &maxUsage); err != nil {
				processorVars[MetaCGRError] = err.Error()
				return
			}
			if reqUsage, has := smgEv[utils.USAGE]; !has { // usage was not requested, decide based on 0
				if maxUsage == 0 {
					reply.Code = radigo.AccessReject
				}
			} else if reqUsage.(time.Duration) < maxUsage {
				reply.Code = radigo.AccessReject
			}
		case MetaRadAcctStart:
			err = ra.smg.Call("SMGenericV2.InitiateSession", smgEv, &maxUsage)
			cgrReply = maxUsage
		case MetaRadAcctUpdate:
			err = ra.smg.Call("SMGenericV2.UpdateSession", smgEv, &maxUsage)
			cgrReply = maxUsage
		case MetaRadAcctStop:
			var rpl string
			err = ra.smg.Call("SMGenericV1.TerminateSession", smgEv, &rpl)
			cgrReply = rpl
			if ra.cgrCfg.RadiusAgentCfg().CreateCDR {
				if errCdr := ra.smg.Call("SMGenericV1.ProcessCDR", smgEv, &rpl); errCdr != nil {
					err = errCdr
				} else {
					cgrReply = rpl
				}
			}
		default:
			err = fmt.Errorf("unsupported radius request type: <%s>", processorVars[MetaRadReqType])
		}
		if err != nil {
			processorVars[MetaCGRError] = err.Error()
			return false, err
		}
		processorVars[MetaCGRReply] = utils.ToJSON(cgrReply)
		processorVars[MetaCGRMaxUsage] = strconv.Itoa(int(maxUsage))
	}
	if err := radReplyAppendAttributes(reply, processorVars, reqProcessor.ReplyFields); err != nil {
		return false, err
	}
	if reqProcessor.DryRun {
		utils.Logger.Info(fmt.Sprintf("<RadiusAgent> DRY_RUN, radius reply: %s", reply))
	}
	return true, nil
}

func (ra *RadiusAgent) ListenAndServe() (err error) {
	var errListen chan error
	go func() {
		utils.Logger.Info(fmt.Sprintf("<RadiusAgent> Start listening for auth requests on <%s>", ra.cgrCfg.RadiusAgentCfg().ListenAuth))
		if err := ra.rsAuth.ListenAndServe(); err != nil {
			errListen <- err
		}
	}()
	go func() {
		utils.Logger.Info(fmt.Sprintf("<RadiusAgent> Start listening for acct req on <%s>", ra.cgrCfg.RadiusAgentCfg().ListenAcct))
		if err := ra.rsAcct.ListenAndServe(); err != nil {
			errListen <- err
		}
	}()
	err = <-errListen
	return
}
