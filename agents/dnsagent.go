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
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/miekg/dns"
)

// NewDNSAgent is the constructor for DNSAgent
func NewDNSAgent(cgrCfg *config.CGRConfig, fltrS *engine.FilterS,
	connMgr *engine.ConnManager) (da *DNSAgent, err error) {
	da = &DNSAgent{cgrCfg: cgrCfg, fltrS: fltrS, connMgr: connMgr}
	err = da.initDNSServer()
	return
}

// DNSAgent translates DNS requests towards CGRateS infrastructure
type DNSAgent struct {
	cgrCfg  *config.CGRConfig // loaded CGRateS configuration
	fltrS   *engine.FilterS   // connection towards FilterS
	server  *dns.Server
	connMgr *engine.ConnManager
}

// initDNSServer instantiates the DNS server
func (da *DNSAgent) initDNSServer() (err error) {
	handler := dns.HandlerFunc(func(w dns.ResponseWriter, m *dns.Msg) {
		go da.handleMessage(w, m)
	})

	if strings.HasSuffix(da.cgrCfg.DNSAgentCfg().ListenNet, utils.TLSNoCaps) {
		cert, err := tls.LoadX509KeyPair(da.cgrCfg.TLSCfg().ServerCerificate, da.cgrCfg.TLSCfg().ServerKey)
		if err != nil {
			return err
		}

		config := tls.Config{
			Certificates: []tls.Certificate{cert},
		}

		da.server = &dns.Server{
			Addr:      da.cgrCfg.DNSAgentCfg().Listen,
			Net:       "tcp-tls",
			TLSConfig: &config,
			Handler:   handler,
		}
	} else {
		da.server = &dns.Server{Addr: da.cgrCfg.DNSAgentCfg().Listen, Net: da.cgrCfg.DNSAgentCfg().ListenNet, Handler: handler}
	}
	return
}

// ListenAndServe will run the DNS handler doing also the connection to listen address
func (da *DNSAgent) ListenAndServe() (err error) {
	utils.Logger.Info(fmt.Sprintf("<%s> start listening on <%s:%s>",
		utils.DNSAgent, da.cgrCfg.DNSAgentCfg().ListenNet, da.cgrCfg.DNSAgentCfg().Listen))
	return da.server.ListenAndServe()
}

// Reload will reinitialize the server
// this is in order to monitor if we receive error on ListenAndServe
func (da *DNSAgent) Reload() (err error) {
	return da.initDNSServer()
}

// handleMessage is the entry point of all DNS requests
// requests are reaching here asynchronously
func (da *DNSAgent) handleMessage(w dns.ResponseWriter, req *dns.Msg) {
	dnsDP := newDNSDataProvider(req, w)
	reqVars := make(utils.NavigableMap2)
	reqVars[QueryType] = utils.NewNMData(dns.TypeToString[req.Question[0].Qtype])
	rply := new(dns.Msg)
	rply.SetReply(req)
	// message preprocesing
	switch req.Question[0].Qtype {
	case dns.TypeNAPTR:
		reqVars[QueryName] = utils.NewNMData(req.Question[0].Name)
		e164, err := e164FromNAPTR(req.Question[0].Name)
		if err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> decoding NAPTR query: <%s>, err: %s",
					utils.DNSAgent, req.Question[0].Name, err.Error()))
			rply.Rcode = dns.RcodeServerFailure
			dnsWriteMsg(w, rply)
			return
		}
		reqVars[E164Address] = utils.NewNMData(e164)
		reqVars[DomainName] = utils.NewNMData(domainNameFromNAPTR(req.Question[0].Name))
	}
	reqVars[utils.RemoteHost] = utils.NewNMData(w.RemoteAddr().String())
	cgrRplyNM := utils.NavigableMap2{}
	rplyNM := utils.NewOrderedNavigableMap() // share it among different processors
	opts := utils.NewOrderedNavigableMap()
	var processed bool
	var err error
	for _, reqProcessor := range da.cgrCfg.DNSAgentCfg().RequestProcessors {
		var lclProcessed bool
		lclProcessed, err = da.processRequest(
			reqProcessor,
			NewAgentRequest(
				dnsDP, reqVars, &cgrRplyNM, rplyNM,
				opts, reqProcessor.Tenant,
				da.cgrCfg.GeneralCfg().DefaultTenant,
				utils.FirstNonEmpty(da.cgrCfg.DNSAgentCfg().Timezone,
					da.cgrCfg.GeneralCfg().DefaultTimezone),
				da.fltrS, nil, nil))
		if lclProcessed {
			processed = lclProcessed
		}
		if err != nil ||
			(lclProcessed && !reqProcessor.Flags.GetBool(utils.MetaContinue)) {
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
		return
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
	if pass, err := da.fltrS.Pass(agReq.Tenant,
		reqProcessor.Filters, agReq); err != nil || !pass {
		return pass, err
	}
	if err = agReq.SetFields(reqProcessor.RequestFields); err != nil {
		return
	}
	cgrEv := config.NMAsCGREvent(agReq.CGRRequest, agReq.Tenant, utils.NestingSep)
	opts := config.NMAsMapInterface(agReq.Opts, utils.NestingSep)
	var reqType string
	for _, typ := range []string{
		utils.MetaDryRun, utils.MetaAuthorize,
		utils.MetaInitiate, utils.MetaUpdate,
		utils.MetaTerminate, utils.MetaMessage,
		utils.MetaCDRs, utils.MetaEvent, utils.MetaNone} {
		if reqProcessor.Flags.Has(typ) { // request type is identified through flags
			reqType = typ
			break
		}
	}
	var cgrArgs utils.Paginator
	if reqType == utils.MetaAuthorize ||
		reqType == utils.MetaMessage ||
		reqType == utils.MetaEvent {
		if cgrArgs, err = utils.GetRoutePaginatorFromOpts(opts); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<%s> args extraction failed because <%s>",
				utils.DNSAgent, err.Error()))
			err = nil // reset the error and continue the processing
		}
	}
	if reqProcessor.Flags.Has(utils.MetaLog) {
		utils.Logger.Info(
			fmt.Sprintf("<%s> LOG, processorID: <%s>, message: %s",
				utils.DNSAgent, reqProcessor.ID, agReq.Request.String()))
	}
	switch reqType {
	default:
		return false, fmt.Errorf("unknown request type: <%s>", reqType)
	case utils.MetaNone: // do nothing on CGRateS side
	case utils.MetaDryRun:
		utils.Logger.Info(
			fmt.Sprintf("<%s> DRY_RUN, processorID: %s, CGREvent: %s",
				utils.DNSAgent, reqProcessor.ID, utils.ToJSON(cgrEv)))
	case utils.MetaAuthorize:
		authArgs := sessions.NewV1AuthorizeArgs(
			reqProcessor.Flags.GetBool(utils.MetaAttributes),
			reqProcessor.Flags.ParamsSlice(utils.MetaAttributes, utils.MetaIDs),
			reqProcessor.Flags.GetBool(utils.MetaThresholds),
			reqProcessor.Flags.ParamsSlice(utils.MetaThresholds, utils.MetaIDs),
			reqProcessor.Flags.GetBool(utils.MetaStats),
			reqProcessor.Flags.ParamsSlice(utils.MetaStats, utils.MetaIDs),
			reqProcessor.Flags.GetBool(utils.MetaResources),
			reqProcessor.Flags.Has(utils.MetaAccounts),
			reqProcessor.Flags.GetBool(utils.MetaRoutes),
			reqProcessor.Flags.Has(utils.MetaRoutesIgnoreErrors),
			reqProcessor.Flags.Has(utils.MetaRoutesEventCost),
			cgrEv, cgrArgs, reqProcessor.Flags.Has(utils.MetaFD),
			opts,
		)
		rply := new(sessions.V1AuthorizeReply)
		err = da.connMgr.Call(da.cgrCfg.DNSAgentCfg().SessionSConns, nil,
			utils.SessionSv1AuthorizeEvent,
			authArgs, rply)
		if err = agReq.setCGRReply(rply, err); err != nil {
			return
		}
	case utils.MetaInitiate:
		initArgs := sessions.NewV1InitSessionArgs(
			reqProcessor.Flags.GetBool(utils.MetaAttributes),
			reqProcessor.Flags.ParamsSlice(utils.MetaAttributes, utils.MetaIDs),
			reqProcessor.Flags.GetBool(utils.MetaThresholds),
			reqProcessor.Flags.ParamsSlice(utils.MetaThresholds, utils.MetaIDs),
			reqProcessor.Flags.GetBool(utils.MetaStats),
			reqProcessor.Flags.ParamsSlice(utils.MetaStats, utils.MetaIDs),
			reqProcessor.Flags.GetBool(utils.MetaResources),
			reqProcessor.Flags.Has(utils.MetaAccounts),
			cgrEv, reqProcessor.Flags.Has(utils.MetaFD),
			opts)
		rply := new(sessions.V1InitSessionReply)
		err = da.connMgr.Call(da.cgrCfg.DNSAgentCfg().SessionSConns, nil,
			utils.SessionSv1InitiateSession,
			initArgs, rply)
		if err = agReq.setCGRReply(rply, err); err != nil {
			return
		}
	case utils.MetaUpdate:
		updateArgs := sessions.NewV1UpdateSessionArgs(
			reqProcessor.Flags.GetBool(utils.MetaAttributes),
			reqProcessor.Flags.ParamsSlice(utils.MetaAttributes, utils.MetaIDs),
			reqProcessor.Flags.Has(utils.MetaAccounts),
			cgrEv, reqProcessor.Flags.Has(utils.MetaFD),
			opts)
		rply := new(sessions.V1UpdateSessionReply)
		err = da.connMgr.Call(da.cgrCfg.DNSAgentCfg().SessionSConns, nil,
			utils.SessionSv1UpdateSession,
			updateArgs, rply)
		if err = agReq.setCGRReply(rply, err); err != nil {
			return
		}
	case utils.MetaTerminate:
		terminateArgs := sessions.NewV1TerminateSessionArgs(
			reqProcessor.Flags.Has(utils.MetaAccounts),
			reqProcessor.Flags.GetBool(utils.MetaResources),
			reqProcessor.Flags.GetBool(utils.MetaThresholds),
			reqProcessor.Flags.ParamsSlice(utils.MetaThresholds, utils.MetaIDs),
			reqProcessor.Flags.GetBool(utils.MetaStats),
			reqProcessor.Flags.ParamsSlice(utils.MetaStats, utils.MetaIDs),
			cgrEv, reqProcessor.Flags.Has(utils.MetaFD),
			opts)
		rply := utils.StringPointer("")
		err = da.connMgr.Call(da.cgrCfg.DNSAgentCfg().SessionSConns, nil,
			utils.SessionSv1TerminateSession,
			terminateArgs, rply)
		if err = agReq.setCGRReply(nil, err); err != nil {
			return
		}
	case utils.MetaMessage:
		evArgs := sessions.NewV1ProcessMessageArgs(
			reqProcessor.Flags.GetBool(utils.MetaAttributes),
			reqProcessor.Flags.ParamsSlice(utils.MetaAttributes, utils.MetaIDs),
			reqProcessor.Flags.GetBool(utils.MetaThresholds),
			reqProcessor.Flags.ParamsSlice(utils.MetaThresholds, utils.MetaIDs),
			reqProcessor.Flags.GetBool(utils.MetaStats),
			reqProcessor.Flags.ParamsSlice(utils.MetaStats, utils.MetaIDs),
			reqProcessor.Flags.GetBool(utils.MetaResources),
			reqProcessor.Flags.Has(utils.MetaAccounts),
			reqProcessor.Flags.GetBool(utils.MetaRoutes),
			reqProcessor.Flags.Has(utils.MetaRoutesIgnoreErrors),
			reqProcessor.Flags.Has(utils.MetaRoutesEventCost),
			cgrEv, cgrArgs, reqProcessor.Flags.Has(utils.MetaFD),
			opts)
		rply := new(sessions.V1ProcessMessageReply) // need it so rpcclient can clone
		err = da.connMgr.Call(da.cgrCfg.DNSAgentCfg().SessionSConns, nil,
			utils.SessionSv1ProcessMessage,
			evArgs, rply)
		if utils.ErrHasPrefix(err, utils.RalsErrorPrfx) {
			cgrEv.Event[utils.Usage] = 0 // avoid further debits
		} else if evArgs.Debit {
			cgrEv.Event[utils.Usage] = rply.MaxUsage // make sure the CDR reflects the debit
		}
		if err = agReq.setCGRReply(rply, err); err != nil {
			return
		}
	case utils.MetaEvent:
		evArgs := &sessions.V1ProcessEventArgs{
			Flags: reqProcessor.Flags.SliceFlags(),
			CGREventWithOpts: &utils.CGREventWithOpts{
				CGREvent: cgrEv,
				Opts:     opts,
			},
			Paginator: cgrArgs,
		}
		rply := new(sessions.V1ProcessEventReply)
		err = da.connMgr.Call(da.cgrCfg.DNSAgentCfg().SessionSConns, nil,
			utils.SessionSv1ProcessEvent,
			evArgs, rply)
		if utils.ErrHasPrefix(err, utils.RalsErrorPrfx) {
			cgrEv.Event[utils.Usage] = 0 // avoid further debits
		} else if needsMaxUsage(reqProcessor.Flags[utils.MetaRALs]) {
			cgrEv.Event[utils.Usage] = rply.MaxUsage // make sure the CDR reflects the debit
		}
		if err = agReq.setCGRReply(rply, err); err != nil {
			return
		}
	case utils.MetaCDRs: // allow CDR processing
	}
	// separate request so we can capture the Terminate/Event also here
	if reqProcessor.Flags.GetBool(utils.MetaCDRs) &&
		!reqProcessor.Flags.Has(utils.MetaDryRun) {
		rplyCDRs := utils.StringPointer("")
		if err = da.connMgr.Call(da.cgrCfg.DNSAgentCfg().SessionSConns, nil,
			utils.SessionSv1ProcessCDR,
			&utils.CGREventWithOpts{
				CGREvent: cgrEv,
				Opts:     opts,
			}, &rplyCDRs); err != nil {
			agReq.CGRReply.Set(utils.PathItems{{Field: utils.Error}}, utils.NewNMData(err.Error()))
		}
	}
	if err := agReq.SetFields(reqProcessor.ReplyFields); err != nil {
		return false, err
	}
	if reqProcessor.Flags.Has(utils.MetaLog) {
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

// Shutdown stops the DNS server
func (da *DNSAgent) Shutdown() error {
	return da.server.Shutdown()
}
