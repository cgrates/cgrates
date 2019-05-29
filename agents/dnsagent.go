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
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
	"github.com/miekg/dns"
)

// NewDNSAgent is the constructor for DNSAgent
func NewDNSAgent(cgrCfg *config.CGRConfig, fltrS *engine.FilterS,
	sS rpcclient.RpcClientConnection) (da *DNSAgent, err error) {
	da = &DNSAgent{cgrCfg: cgrCfg, fltrS: fltrS, sS: sS}
	return
}

// DNSAgent translates DNS requests towards CGRateS infrastructure
type DNSAgent struct {
	cgrCfg *config.CGRConfig             // loaded CGRateS configuration
	fltrS  *engine.FilterS               // connection towards FilterS
	sS     rpcclient.RpcClientConnection // connection towards CGR-SessionS component
}

// ListenAndServe will run the DNS handler doing also the connection to listen address
func (da *DNSAgent) ListenAndServe() error {
	utils.Logger.Info(fmt.Sprintf("<%s> start listening on <%s:%s>",
		utils.DNSAgent, da.cgrCfg.DNSAgentCfg().ListenNet, da.cgrCfg.DNSAgentCfg().Listen))
	if strings.HasSuffix(da.cgrCfg.DNSAgentCfg().ListenNet, utils.TLSNoCaps) {
		return dns.ListenAndServeTLS(
			da.cgrCfg.DNSAgentCfg().Listen,
			da.cgrCfg.TlsCfg().ServerCerificate,
			da.cgrCfg.TlsCfg().ServerKey,
			dns.HandlerFunc(
				func(w dns.ResponseWriter, m *dns.Msg) {
					go da.handleMessage(w, m)
				}),
		)
	}
	return dns.ListenAndServe(
		da.cgrCfg.DNSAgentCfg().Listen,
		da.cgrCfg.DNSAgentCfg().ListenNet,
		dns.HandlerFunc(
			func(w dns.ResponseWriter, m *dns.Msg) {
				go da.handleMessage(w, m)
			}),
	)
}

// handleMessage is the entry point of all DNS requests
// requests are reaching here asynchronously
func (da *DNSAgent) handleMessage(w dns.ResponseWriter, req *dns.Msg) {
	dnsDP := newDNSDataProvider(req, w)
	reqVars := make(map[string]interface{})
	reqVars[QueryType] = dns.TypeToString[req.Question[0].Qtype]
	rply := new(dns.Msg)
	rply.SetReply(req)
	// message preprocesing
	switch req.Question[0].Qtype {
	case dns.TypeNAPTR:
		e164, err := e164FromNAPTR(req.Question[0].Name)
		if err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> decoding NAPTR query: <%s>, err: %s",
					utils.DNSAgent, req.Question[0].Name, err.Error()))
			rply.Rcode = dns.RcodeServerFailure
			dnsWriteMsg(w, rply)
			return
		}
		reqVars[E164Address] = e164
	}
	rplyNM := config.NewNavigableMap(nil) // share it among different processors
	var processed bool
	var err error
	for _, reqProcessor := range da.cgrCfg.DNSAgentCfg().RequestProcessors {
		var lclProcessed bool
		lclProcessed, err = da.processRequest(
			reqProcessor,
			newAgentRequest(
				dnsDP, reqVars, rplyNM,
				reqProcessor.Tenant,
				da.cgrCfg.GeneralCfg().DefaultTenant,
				utils.FirstNonEmpty(da.cgrCfg.DNSAgentCfg().Timezone,
					da.cgrCfg.GeneralCfg().DefaultTimezone),
				da.fltrS))
		if lclProcessed {
			processed = lclProcessed
		}
		if err != nil ||
			(lclProcessed && !reqProcessor.ContinueOnSuccess) {
			break
		}
	}
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: %s processing message: %s from %s",
				utils.DNSAgent, err.Error(), req, w.RemoteAddr()))
		rply.Rcode = dns.RcodeServerFailure
		dnsWriteMsg(w, rply)
		return
	} else if !processed {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> no request processor enabled, ignoring message %s from %s",
				utils.DNSAgent, req, w.RemoteAddr()))
		rply.Rcode = dns.RcodeServerFailure
		dnsWriteMsg(w, rply)
		return
	}
	if err = updateDNSMsgFromNM(rply, rplyNM); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: %s updating answer: %s from NM %s",
				utils.DNSAgent, err.Error(), utils.ToJSON(rply), utils.ToJSON(rplyNM)))
		rply.Rcode = dns.RcodeServerFailure
		dnsWriteMsg(w, rply)
	}
	if err = dnsWriteMsg(w, rply); err != nil { // failed sending, most probably content issue
		rply = new(dns.Msg)
		rply.SetReply(req)
		rply.Rcode = dns.RcodeServerFailure
		dnsWriteMsg(w, rply)
	}
	return
}

func (da *DNSAgent) processRequest(reqProcessor *config.RequestProcessor,
	agReq *AgentRequest) (processed bool, err error) {
	if pass, err := da.fltrS.Pass(agReq.tenant,
		reqProcessor.Filters, agReq); err != nil || !pass {
		return pass, err
	}
	if agReq.CGRRequest, err = agReq.AsNavigableMap(reqProcessor.RequestFields); err != nil {
		return
	}
	cgrEv := agReq.CGRRequest.AsCGREvent(agReq.tenant, utils.NestingSep)
	var reqType string
	for _, typ := range []string{
		utils.MetaDryRun, utils.MetaAuth,
		utils.MetaInitiate, utils.MetaUpdate,
		utils.MetaTerminate, utils.MetaEvent,
		utils.MetaCDRs, utils.META_NONE} {
		if reqProcessor.Flags.HasKey(typ) { // request type is identified through flags
			reqType = typ
			break
		}
	}
	cgrArgs := cgrEv.ConsumeArgs(reqProcessor.Flags.HasKey(utils.MetaDispatchers), reqType == utils.MetaAuth || reqType == utils.MetaEvent)
	if reqProcessor.Flags.HasKey(utils.MetaLog) {
		utils.Logger.Info(
			fmt.Sprintf("<%s> LOG, processorID: <%s>, message: %s",
				utils.DNSAgent, reqProcessor.ID, agReq.Request.String()))
	}
	switch reqType {
	default:
		return false, fmt.Errorf("unknown request type: <%s>", reqType)
	case utils.META_NONE: // do nothing on CGRateS side
	case utils.MetaDryRun:
		utils.Logger.Info(
			fmt.Sprintf("<%s> DRY_RUN, processorID: %s, CGREvent: %s",
				utils.DNSAgent, reqProcessor.ID, utils.ToJSON(cgrEv)))
	case utils.MetaAuth:
		authArgs := sessions.NewV1AuthorizeArgs(
			reqProcessor.Flags.HasKey(utils.MetaAttributes),
			reqProcessor.Flags.HasKey(utils.MetaResources),
			reqProcessor.Flags.HasKey(utils.MetaAccounts),
			reqProcessor.Flags.HasKey(utils.MetaThresholds),
			reqProcessor.Flags.HasKey(utils.MetaStats),
			reqProcessor.Flags.HasKey(utils.MetaSuppliers),
			reqProcessor.Flags.HasKey(utils.MetaSuppliersIgnoreErrors),
			reqProcessor.Flags.HasKey(utils.MetaSuppliersEventCost),
			cgrEv, cgrArgs.ArgDispatcher, *cgrArgs.SupplierPaginator)
		var authReply sessions.V1AuthorizeReply
		err = da.sS.Call(utils.SessionSv1AuthorizeEvent,
			authArgs, &authReply)
		if agReq.CGRReply, err = NewCGRReply(&authReply, err); err != nil {
			return
		}
	case utils.MetaInitiate:
		initArgs := sessions.NewV1InitSessionArgs(
			reqProcessor.Flags.HasKey(utils.MetaAttributes),
			reqProcessor.Flags.HasKey(utils.MetaResources),
			reqProcessor.Flags.HasKey(utils.MetaAccounts),
			reqProcessor.Flags.HasKey(utils.MetaThresholds),
			reqProcessor.Flags.HasKey(utils.MetaStats),
			cgrEv, cgrArgs.ArgDispatcher)
		var initReply sessions.V1InitSessionReply
		err = da.sS.Call(utils.SessionSv1InitiateSession,
			initArgs, &initReply)
		if agReq.CGRReply, err = NewCGRReply(&initReply, err); err != nil {
			return
		}
	case utils.MetaUpdate:
		updateArgs := sessions.NewV1UpdateSessionArgs(
			reqProcessor.Flags.HasKey(utils.MetaAttributes),
			reqProcessor.Flags.HasKey(utils.MetaAccounts),
			cgrEv, cgrArgs.ArgDispatcher)
		var updateReply sessions.V1UpdateSessionReply
		err = da.sS.Call(utils.SessionSv1UpdateSession,
			updateArgs, &updateReply)
		if agReq.CGRReply, err = NewCGRReply(&updateReply, err); err != nil {
			return
		}
	case utils.MetaTerminate:
		terminateArgs := sessions.NewV1TerminateSessionArgs(
			reqProcessor.Flags.HasKey(utils.MetaAccounts),
			reqProcessor.Flags.HasKey(utils.MetaResources),
			reqProcessor.Flags.HasKey(utils.MetaThresholds),
			reqProcessor.Flags.HasKey(utils.MetaStats),
			cgrEv, cgrArgs.ArgDispatcher)
		var tRply string
		err = da.sS.Call(utils.SessionSv1TerminateSession,
			terminateArgs, &tRply)
		if agReq.CGRReply, err = NewCGRReply(nil, err); err != nil {
			return
		}
	case utils.MetaEvent:
		evArgs := sessions.NewV1ProcessEventArgs(
			reqProcessor.Flags.HasKey(utils.MetaResources),
			reqProcessor.Flags.HasKey(utils.MetaAccounts),
			reqProcessor.Flags.HasKey(utils.MetaAttributes),
			reqProcessor.Flags.HasKey(utils.MetaThresholds),
			reqProcessor.Flags.HasKey(utils.MetaStats),
			reqProcessor.Flags.HasKey(utils.MetaSuppliers),
			reqProcessor.Flags.HasKey(utils.MetaSuppliersIgnoreErrors),
			reqProcessor.Flags.HasKey(utils.MetaSuppliersEventCost),
			cgrEv, cgrArgs.ArgDispatcher, *cgrArgs.SupplierPaginator)
		var eventRply sessions.V1ProcessEventReply
		err = da.sS.Call(utils.SessionSv1ProcessEvent,
			evArgs, &eventRply)
		if utils.ErrHasPrefix(err, utils.RalsErrorPrfx) {
			cgrEv.Event[utils.Usage] = 0 // avoid further debits
		} else if eventRply.MaxUsage != nil {
			cgrEv.Event[utils.Usage] = *eventRply.MaxUsage // make sure the CDR reflects the debit
		}
		if agReq.CGRReply, err = NewCGRReply(&eventRply, err); err != nil {
			return
		}
	case utils.MetaCDRs: // allow CDR processing
	}
	// separate request so we can capture the Terminate/Event also here
	if reqProcessor.Flags.HasKey(utils.MetaCDRs) &&
		!reqProcessor.Flags.HasKey(utils.MetaDryRun) {
		var rplyCDRs string
		if err = da.sS.Call(utils.SessionSv1ProcessCDR,
			&utils.CGREventWithArgDispatcher{CGREvent: cgrEv,
				ArgDispatcher: cgrArgs.ArgDispatcher}, &rplyCDRs); err != nil {
			agReq.CGRReply.Set([]string{utils.Error}, err.Error(), false, false)
		}
	}
	if nM, err := agReq.AsNavigableMap(reqProcessor.ReplyFields); err != nil {
		return false, err
	} else {
		agReq.Reply.Merge(nM)
	}
	if reqProcessor.Flags.HasKey(utils.MetaLog) {
		utils.Logger.Info(
			fmt.Sprintf("<%s> LOG, reply: %s",
				utils.DNSAgent, agReq.Reply))
	}
	if reqType == utils.MetaDryRun {
		utils.Logger.Info(
			fmt.Sprintf("<%s> DRY_RUN, reply: %s",
				utils.DNSAgent, agReq.Reply))
	}
	return true, nil
}
