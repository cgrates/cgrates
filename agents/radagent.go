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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/radigo"
	"github.com/cgrates/rpcclient"
)

func NewRadiusAgent(cgrCfg *config.CGRConfig, smg rpcclient.RpcClientConnection) (ra *RadiusAgent, err error) {
	dicts := make(map[string]*radigo.Dictionary, len(cgrCfg.RadiusAgentCfg().ClientDictionaries))
	for clntID, dictPath := range cgrCfg.RadiusAgentCfg().ClientDictionaries {
		if dicts[clntID], err = radigo.NewDictionaryFromFolderWithRFC2865(dictPath); err != nil {
			return
		}
	}
	ra = &RadiusAgent{cgrCfg: cgrCfg, smg: smg}
	ra.rsAuth = radigo.NewServer(cgrCfg.RadiusAgentCfg().ListenNet, cgrCfg.RadiusAgentCfg().ListenAuth, cgrCfg.RadiusAgentCfg().ClientSecrets, dicts,
		map[radigo.PacketCode]func(*radigo.Packet) (*radigo.Packet, error){radigo.AccessRequest: ra.handleAuth}, nil)
	ra.rsAcct = radigo.NewServer(cgrCfg.RadiusAgentCfg().ListenNet, cgrCfg.RadiusAgentCfg().ListenAcct, cgrCfg.RadiusAgentCfg().ClientSecrets, dicts,
		map[radigo.PacketCode]func(*radigo.Packet) (*radigo.Packet, error){radigo.AccountingRequest: ra.handleAcct}, nil)
	return

}

type RadiusAgent struct {
	cgrCfg *config.CGRConfig             // reference for future config reloads
	smg    rpcclient.RpcClientConnection // Connection towards CGR-SMG component
	rsAuth *radigo.Server
	rsAcct *radigo.Server
}

func (ra *RadiusAgent) handleAuth(req *radigo.Packet) (rpl *radigo.Packet, err error) {
	utils.Logger.Debug(fmt.Sprintf("RadiusAgent handleAuth, received request: %+v", req))
	rpl = req.Reply()
	rpl.Code = radigo.AccessAccept
	for _, avp := range req.AVPs {
		rpl.AVPs = append(rpl.AVPs, avp)
	}
	return
}

// RadiusAgent handleAcct, received req: &{RWMutex:{w:{state:0 sema:0} writerSem:0 readerSem:0 readerCount:0 readerWait:0} dict:0xc4202e5840 secret:CGRateS.org Code:AccountingRequest Identifier:143 Authenticator:[67 77 204 122 189 209 219 22 9 176 15 228 24 246 183 7] AVPs:[0xc42023c230 0xc42023c2a0 0xc42023c310 0xc42023c460 0xc42023c4d0 0xc42023c540 0xc42023c850 0xc42023ce00 0xc42023d180 0xc42023d1f0 0xc42023d260]}
// Identifier:144 Authenticator:[192 197 33 53 203 181 16 117 204 143 172 174 231 245 81 116] AVPs:[0xc42023d5e0 0xc42023d650 0xc42023d880 0xc42023d8f0 0xc42023da40 0xc42023db20 0xc42023dc70 0xc42023dd50 0xc42023ddc0 0xc42023de30 0xc42023dea0]}
func (ra *RadiusAgent) handleAcct(req *radigo.Packet) (rpl *radigo.Packet, err error) {
	req.SetAVPValues()
	utils.Logger.Debug(fmt.Sprintf("Received request: %s", utils.ToJSON(req)))
	rpl = req.Reply()
	rpl.Code = radigo.AccountingResponse
	rpl.SetAVPValues()
	return
}

func (ra *RadiusAgent) processRequest(req *radigo.Packet, reqProcessor *config.RARequestProcessor) (processed bool, err error) {
	passesAllFilters := true
	for _, fldFilter := range reqProcessor.RequestFilter {
		if passes, _ := radPassesFieldFilter(req, fldFilter, nil); !passes {
			passesAllFilters = false
		}
	}
	if !passesAllFilters { // Not going with this processor further
		return false, nil
	}
	return
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
