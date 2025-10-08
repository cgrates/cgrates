/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package agents

import (
	"crypto/tls"
	"fmt"
	"strings"
	"sync"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
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
	sync.RWMutex
	cgrCfg  *config.CGRConfig // loaded CGRateS configuration
	fltrS   *engine.FilterS   // connection towards FilterS
	servers []*dns.Server
	connMgr *engine.ConnManager
}

// initDNSServer instantiates the DNS server
func (da *DNSAgent) initDNSServer() (_ error) {
	da.servers = make([]*dns.Server, 0, len(da.cgrCfg.DNSAgentCfg().Listeners))
	for i := range da.cgrCfg.DNSAgentCfg().Listeners {
		da.servers = append(da.servers, &dns.Server{
			Addr: da.cgrCfg.DNSAgentCfg().Listeners[i].Address,
			Net:  da.cgrCfg.DNSAgentCfg().Listeners[i].Network,
			Handler: dns.HandlerFunc(func(w dns.ResponseWriter, m *dns.Msg) {
				go da.handleMessage(w, m)
			}),
		})
		if strings.HasSuffix(da.cgrCfg.DNSAgentCfg().Listeners[i].Network, utils.TLSNoCaps) {
			cert, err := tls.LoadX509KeyPair(da.cgrCfg.TLSCfg().ServerCerificate, da.cgrCfg.TLSCfg().ServerKey)
			if err != nil {
				return fmt.Errorf("load certificate error <%v>", err)
			}
			da.servers[i].Net = "tcp-tls"
			da.servers[i].TLSConfig = &tls.Config{
				Certificates: []tls.Certificate{cert},
			}
		}
	}
	return
}

// ListenAndServe will run the DNS handler doing also the connection to listen address
func (da *DNSAgent) ListenAndServe(stopChan chan struct{}) error {
	errChan := make(chan error)
	for _, server := range da.servers {
		utils.Logger.Info(fmt.Sprintf("<%s> start listening on <%s:%s>",
			utils.DNSAgent, server.Net, server.Addr))
		go func(srv *dns.Server) {
			err := srv.ListenAndServe()
			if err != nil {
				utils.Logger.Warning(fmt.Sprintf("<%s> error <%v>, on ListenAndServe <%s:%s>",
					utils.DNSAgent, err, srv.Net, srv.Addr))
				if strings.Contains(err.Error(), "address already in use") {
					return
				}
				errChan <- err
			}
		}(server)
	}

	select {
	case <-stopChan:
		return da.Shutdown()
	case err := <-errChan:
		if shtdErr := da.Shutdown(); shtdErr != nil {
			return shtdErr
		}
		return err
	}

}

// Reload will reinitialize the server
// this is in order to monitor if we receive error on ListenAndServe
func (da *DNSAgent) Reload() (err error) {
	return da.initDNSServer()
}

// handleMessage is the entry point of all DNS requests
// requests are reaching here asynchronously
func (da *DNSAgent) handleMessage(w dns.ResponseWriter, req *dns.Msg) {
	dnsDP := newDnsDP(req)

	rply := newDnsReply(req)
	rmtAddr := w.RemoteAddr().String()
	for _, q := range req.Question {
		if processed, err := da.handleQuestion(dnsDP, rply, &q, rmtAddr); err != nil ||
			!processed {
			rply := newDnsReply(req)
			rply.Rcode = dns.RcodeServerFailure
			dnsWriteMsg(w, rply)
			return
		}
	}

	if err := dnsWriteMsg(w, rply); err != nil { // failed sending, most probably content issue
		rply := newDnsReply(req)
		rply.Rcode = dns.RcodeServerFailure
		dnsWriteMsg(w, rply)
	}
}

// Shutdown stops the DNS server
func (da *DNSAgent) Shutdown() error {
	var err error
	for _, server := range da.servers {
		shtdErr := server.Shutdown()
		if shtdErr == nil {
			continue
		}
		utils.Logger.Warning(fmt.Sprintf("<%s> error <%v>, on Shutdown <%s:%s>",
			utils.DNSAgent, shtdErr, server.Net, server.Addr))
		if shtdErr.Error() != "dns: server not started" {
			err = shtdErr
		}
	}
	return err
}

// handleMessage is the entry point of all DNS requests
// requests are reaching here asynchronously
func (da *DNSAgent) handleQuestion(dnsDP utils.DataProvider, rply *dns.Msg, q *dns.Question, rmtAddr string) (processed bool, err error) {
	reqVars := &utils.DataNode{
		Type: utils.NMMapType,
		Map: map[string]*utils.DataNode{
			utils.DNSQueryType: utils.NewLeafNode(dns.TypeToString[q.Qtype]),
			utils.DNSQueryName: utils.NewLeafNode(q.Name),
			utils.RemoteHost:   utils.NewLeafNode(rmtAddr),
		},
	}
	// message preprocesing
	cgrRplyNM := &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}
	rplyNM := utils.NewOrderedNavigableMap() // share it among different processors
	opts := utils.MapStorage{}
	for _, reqProcessor := range da.cgrCfg.DNSAgentCfg().RequestProcessors {
		var lclProcessed bool
		if lclProcessed, err = processRequest(
			context.TODO(),
			reqProcessor,
			NewAgentRequest(
				dnsDP, reqVars, cgrRplyNM, rplyNM,
				opts, reqProcessor.Tenant,
				da.cgrCfg.GeneralCfg().DefaultTenant,
				utils.FirstNonEmpty(da.cgrCfg.DNSAgentCfg().Timezone,
					da.cgrCfg.GeneralCfg().DefaultTimezone),
				da.fltrS, nil),
			utils.DNSAgent, da.connMgr,
			da.cgrCfg.DNSAgentCfg().SessionSConns,
			da.cgrCfg.DNSAgentCfg().StatSConns,
			da.cgrCfg.DNSAgentCfg().ThresholdSConns,
			da.fltrS); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing message: %s from %s",
					utils.DNSAgent, err.Error(), dnsDP, rmtAddr))
			return
		}
		processed = processed || lclProcessed
		if lclProcessed && !reqProcessor.Flags.GetBool(utils.MetaContinue) {
			break
		}
	}
	if !processed {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> no request processor enabled, ignoring message %s from %s",
				utils.DNSAgent, dnsDP, rmtAddr))
		return
	}
	if err = updateDNSMsgFromNM(rply, rplyNM, q.Qtype, q.Name); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: %s updating answer: %s from NM %s",
				utils.DNSAgent, err.Error(), utils.ToJSON(rply), utils.ToJSON(rplyNM)))
	}
	return
}
