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
	"sync"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/miekg/dns"
)

// NewDNSAgent is the constructor for DNSAgent
func NewDNSAgent(cgrCfg *config.CGRConfig, fltrS *engine.FilterS,
	connMgr *engine.ConnManager, caps *engine.Caps) (da *DNSAgent) {
	da = &DNSAgent{
		cgrCfg:  cgrCfg,
		fltrS:   fltrS,
		connMgr: connMgr,
		caps:    caps,
	}
	return
}

// DNSAgent translates DNS requests towards CGRateS infrastructure
type DNSAgent struct {
	sync.RWMutex
	cgrCfg  *config.CGRConfig // loaded CGRateS configuration
	connMgr *engine.ConnManager
	caps    *engine.Caps
	fltrS   *engine.FilterS // connection towards FilterS
	servers []*dns.Server
}

type DNSListener struct {
	server  *dns.Server // server holds the DNS server configuration.
	errChan chan error  // used to communicate errors encountered during DNS listenAndServe
}

// listenAndServe will open a DNS Listener using the configured DNSListener server
func (dal *DNSListener) listenAndServe() {
	utils.Logger.Info(fmt.Sprintf("<%s> start listening on <%s:%s>",
		utils.DNSAgent, dal.server.Net, dal.server.Addr))
	if err := dal.server.ListenAndServe(); err != nil {
		dal.errChan <- err
		return
	}
}

// ShutdownListeners shuts down all listeners in servers slice
func (da *DNSAgent) ShutdownListeners() (err error) {
	da.RLock()
	defer da.RUnlock()
	for _, server := range da.servers {
		utils.Logger.Info(fmt.Sprintf("<%s> Shutting down <%s:%s>",
			utils.DNSAgent, server.Net, server.Addr))
		ctx, cancel := context.WithTimeout(context.Background(), da.cgrCfg.CoreSCfg().ShutdownTimeout) // pass timeout context from configs to server.ShutdownContext
		defer cancel()
		if shErr := server.ShutdownContext(ctx); shErr != nil {
			if shErr.Error() != "dns: server not started" { // ignore error if server is not started
				err = shErr
			}
		}
	}
	return
}

// configureTLS provides TLS certificates to the DNS server
func configureTLS(server *dns.Server, certificateFile, keyFile string) error {
	cert, err := tls.LoadX509KeyPair(certificateFile, keyFile)
	if err != nil {
		return fmt.Errorf("load certificate error <%v>", err)
	}
	server.Net = "tcp-tls"
	server.TLSConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	return nil
}

// ListenAndServe will create dns server configs, run the DNS handler, and start listeners from the servers created
func (da *DNSAgent) ListenAndServe(stopChan chan struct{}) (err error) {
	errChan := make(chan error, 1) // errors coming dnsAL.listenAndServe()
	da.servers = make([]*dns.Server, 0, len(da.cgrCfg.DNSAgentCfg().Listeners))
	// populate servers
	for i := range da.cgrCfg.DNSAgentCfg().Listeners {
		da.servers = append(da.servers, &dns.Server{
			Addr: da.cgrCfg.DNSAgentCfg().Listeners[i].Address,
			Net:  da.cgrCfg.DNSAgentCfg().Listeners[i].Network,
			Handler: dns.HandlerFunc(func(w dns.ResponseWriter, m *dns.Msg) {
				go handleMessage(w, m, da.cgrCfg, da.fltrS, da.connMgr, da.caps)
			}),
		})
		if strings.HasSuffix(da.cgrCfg.DNSAgentCfg().Listeners[i].Network, utils.TLSNoCaps) {
			if err := configureTLS(da.servers[i], da.cgrCfg.TLSCfg().ServerCerificate,
				da.cgrCfg.TLSCfg().ServerKey); err != nil {
				return err
			}
		}
	}
	for _, server := range da.servers { // populate DNSListeners and start listenAndServe
		dnsAL := DNSListener{
			server:  server,
			errChan: errChan,
		}
		go dnsAL.listenAndServe()
	}
	select {
	case err := <-errChan: // if error, shutdown engine
		return err
	case <-stopChan: // wait signal from service shutdown to close this function
		return nil
	}
}

// handleMessage is the entry point of all DNS requests
// requests are reaching here asynchronously
func handleMessage(w dns.ResponseWriter, req *dns.Msg, cfg *config.CGRConfig,
	fltrS *engine.FilterS, connMgr *engine.ConnManager, caps *engine.Caps) {
	if caps.IsLimited() {
		if err := caps.Allocate(); err != nil {
			rply := newDnsReply(req)
			rply.Rcode = dns.RcodeRefused
			dnsWriteMsg(w, rply)
			return
		}
		defer caps.Deallocate()
	}
	dnsDP := newDnsDP(req)

	rply := newDnsReply(req)
	rmtAddr := w.RemoteAddr().String()
	for _, q := range req.Question {
		if processed, err := handleQuestion(dnsDP, rply, &q, rmtAddr, cfg, fltrS, connMgr); err != nil ||
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

// handleMessage is the entry point of all DNS requests
// requests are reaching here asynchronously
func handleQuestion(dnsDP utils.DataProvider, rply *dns.Msg, q *dns.Question, rmtAddr string,
	cfg *config.CGRConfig, fltrS *engine.FilterS, connMgr *engine.ConnManager) (processed bool, err error) {
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
	for _, reqProcessor := range cfg.DNSAgentCfg().RequestProcessors {
		var lclProcessed bool
		if lclProcessed, err = processRequest(
			context.TODO(),
			reqProcessor,
			NewAgentRequest(
				dnsDP, reqVars, cgrRplyNM, rplyNM,
				opts, reqProcessor.Tenant,
				cfg.GeneralCfg().DefaultTenant,
				utils.FirstNonEmpty(
					cfg.DNSAgentCfg().Timezone,
					cfg.GeneralCfg().DefaultTimezone,
				),
				fltrS, nil),
			utils.DNSAgent, connMgr,
			cfg.DNSAgentCfg().SessionSConns,
			fltrS); err != nil {
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
