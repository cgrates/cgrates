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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/radigo"
	"github.com/cgrates/rpcclient"
)

const (
	MetaRadReqType       = "*radReqType"
	MetaRadAuth          = "*radAuth"
	MetaRadAcctStart     = "*radAcctStart"
	MetaRadAcctUpdate    = "*radAcctUpdate"
	MetaRadAcctStop      = "*radAcctStop"
	MetaRadReplyCode     = "*radReplyCode"
	MetaUsageDifference  = "*usage_difference"
	RadAcctStart         = "Start"
	RadAcctInterimUpdate = "Interim-Update"
	RadAcctStop          = "Stop"
)

func NewRadiusAgent(cgrCfg *config.CGRConfig, sessionS rpcclient.RpcClientConnection) (ra *RadiusAgent, err error) {
	dts := make(map[string]*radigo.Dictionary, len(cgrCfg.RadiusAgentCfg().ClientDictionaries))
	for clntID, dictPath := range cgrCfg.RadiusAgentCfg().ClientDictionaries {
		utils.Logger.Info(
			fmt.Sprintf("<%s> loading dictionary for clientID: <%s> out of path <%s>",
				utils.RadiusAgent, clntID, dictPath))
		if dts[clntID], err = radigo.NewDictionaryFromFolderWithRFC2865(dictPath); err != nil {
			return
		}
	}
	dicts := radigo.NewDictionaries(dts)
	ra = &RadiusAgent{cgrCfg: cgrCfg, sessionS: sessionS}
	secrets := radigo.NewSecrets(cgrCfg.RadiusAgentCfg().ClientSecrets)
	ra.rsAuth = radigo.NewServer(cgrCfg.RadiusAgentCfg().ListenNet,
		cgrCfg.RadiusAgentCfg().ListenAuth, secrets, dicts,
		map[radigo.PacketCode]func(*radigo.Packet) (*radigo.Packet, error){
			radigo.AccessRequest: ra.handleAuth}, nil)
	ra.rsAcct = radigo.NewServer(cgrCfg.RadiusAgentCfg().ListenNet,
		cgrCfg.RadiusAgentCfg().ListenAcct, secrets, dicts,
		map[radigo.PacketCode]func(*radigo.Packet) (*radigo.Packet, error){
			radigo.AccountingRequest: ra.handleAcct}, nil)
	return
}

type RadiusAgent struct {
	cgrCfg   *config.CGRConfig             // reference for future config reloads
	sessionS rpcclient.RpcClientConnection // Connection towards CGR-SessionS component
	rsAuth   *radigo.Server
	rsAcct   *radigo.Server
}

// handleAuth handles RADIUS Authorization request
func (ra *RadiusAgent) handleAuth(req *radigo.Packet) (rpl *radigo.Packet, err error) {
	req.SetAVPValues() // populate string values in AVPs
	procVars := processorVars{
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
		utils.Logger.Err(fmt.Sprintf("<%s> error: <%s> ignoring request: %s, process vars: %+v",
			utils.RadiusAgent, err.Error(), utils.ToJSON(req), procVars))
		return nil, nil
	} else if !processed {
		utils.Logger.Err(fmt.Sprintf("<%s> no request processor enabled, ignoring request %s, process vars: %+v",
			utils.RadiusAgent, utils.ToJSON(req), procVars))
		return nil, nil
	}
	return
}

// handleAcct handles RADIUS Accounting request
// supports: Acct-Status-Type = Start, Interim-Update, Stop
func (ra *RadiusAgent) handleAcct(req *radigo.Packet) (rpl *radigo.Packet, err error) {
	req.SetAVPValues() // populate string values in AVPs
	procVars := make(processorVars)
	if avps := req.AttributesWithName("Acct-Status-Type", ""); len(avps) != 0 { // populate accounting type
		switch avps[0].GetStringValue() { // first AVP found will give out the type of accounting
		case RadAcctStart:
			procVars[MetaRadReqType] = MetaRadAcctStart
		case RadAcctInterimUpdate:
			procVars[MetaRadReqType] = MetaRadAcctUpdate
		case RadAcctStop:
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
		utils.Logger.Err(fmt.Sprintf("<%s> error: <%s> ignoring request: %s, process vars: %+v",
			utils.RadiusAgent, err.Error(), utils.ToJSON(req), procVars))
		return nil, nil
	} else if !processed {
		utils.Logger.Err(fmt.Sprintf("<%s> no request processor enabled, ignoring request %s, process vars: %+v",
			utils.RadiusAgent, utils.ToJSON(req), procVars))
		return nil, nil
	}
	return
}

// processRequest represents one processor processing the request
func (ra *RadiusAgent) processRequest(reqProcessor *config.RARequestProcessor,
	req *radigo.Packet, procVars processorVars, reply *radigo.Packet) (processed bool, err error) {
	passesAllFilters := true
	for _, fldFilter := range reqProcessor.RequestFilter {
		if !radPassesFieldFilter(req, procVars, fldFilter) {
			passesAllFilters = false
			break
		}
	}
	if !passesAllFilters { // Not going with this processor further
		return false, nil
	}
	for k, v := range reqProcessor.Flags { // update procVars with flags from processor
		procVars[k] = strconv.FormatBool(v)
	}
	if reqProcessor.DryRun {
		utils.Logger.Info(fmt.Sprintf("<%s> DRY_RUN, RADIUS request: %s", utils.RadiusAgent, utils.ToJSON(req)))
		utils.Logger.Info(fmt.Sprintf("<%s> DRY_RUN, process variabiles: %+v", utils.RadiusAgent, procVars))
	}
	cgrEv, err := radReqAsCGREvent(req, procVars, reqProcessor.Flags, reqProcessor.RequestFields)
	if err != nil {
		return false, err
	}
	if reqProcessor.DryRun {
		utils.Logger.Info(fmt.Sprintf("<%s> DRY_RUN, CGREvent: %s",utils.RadiusAgent, utils.ToJSON(cgrEv)))
	} else { // process with RPC
		switch procVars[MetaRadReqType] {
		case MetaRadAuth:
			var authReply sessions.V1AuthorizeReply
			err = ra.sessionS.Call(utils.SessionSv1AuthorizeEvent,
				radV1AuthorizeArgs(cgrEv, procVars), &authReply)
			if procVars[utils.MetaCGRReply], err = utils.NewCGRReply(&authReply, err); err != nil {
				return
			}
		case MetaRadAcctStart:
			var initReply sessions.V1InitSessionReply
			err = ra.sessionS.Call(utils.SessionSv1InitiateSession,
				radV1InitSessionArgs(cgrEv, procVars), &initReply)
			if procVars[utils.MetaCGRReply], err = utils.NewCGRReply(&initReply, err); err != nil {
				return
			}
		case MetaRadAcctUpdate:
			var updateReply sessions.V1UpdateSessionReply
			err = ra.sessionS.Call(utils.SessionSv1UpdateSession,
				radV1UpdateSessionArgs(cgrEv, procVars), &updateReply)
			if procVars[utils.MetaCGRReply], err = utils.NewCGRReply(&updateReply, err); err != nil {
				return
			}
		case MetaRadAcctStop:
			var rpl string
			if err = ra.sessionS.Call(utils.SessionSv1TerminateSession,
				radV1TerminateSessionArgs(cgrEv, procVars), &rpl); err != nil {
				procVars[utils.MetaCGRReply] = &utils.CGRReply{utils.Error: err.Error()}
			}
			if ra.cgrCfg.RadiusAgentCfg().CreateCDR {
				if errCdr := ra.sessionS.Call(utils.SessionSv1ProcessCDR, *cgrEv, &rpl); errCdr != nil {
					procVars[utils.MetaCGRReply] = &utils.CGRReply{utils.Error: err.Error()}
					err = errCdr
				}

			}
			if err != nil {
				return
			}
		default:
			err = fmt.Errorf("unsupported radius request type: <%s>", procVars[MetaRadReqType])
		}
	}
	if err := radReplyAppendAttributes(reply, procVars, reqProcessor.ReplyFields); err != nil {
		return false, err
	}
	if reqProcessor.DryRun {
		utils.Logger.Info(fmt.Sprintf("<RadiusAgent> DRY_RUN, radius reply: %+v", reply))
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
