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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
	"github.com/fiorix/go-diameter/diam"
	"github.com/fiorix/go-diameter/diam/datatype"
	"github.com/fiorix/go-diameter/diam/sm"
)

func NewDiameterAgent(cgrCfg *config.CGRConfig, smg *rpcclient.RpcClient) (*DiameterAgent, error) {
	da := &DiameterAgent{cgrCfg: cgrCfg, smg: smg}
	dictsDir := cgrCfg.DiameterAgentCfg().DictionariesDir
	if len(dictsDir) != 0 {
		if err := loadDictionaries(dictsDir, "DiameterAgent"); err != nil {
			return nil, err
		}
	}
	return da, nil
}

type DiameterAgent struct {
	cgrCfg *config.CGRConfig
	smg    *rpcclient.RpcClient // Connection towards CGR-SMG component
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

func (self *DiameterAgent) handleCCR(c diam.Conn, m *diam.Message) {
	//utils.Logger.Warning(fmt.Sprintf("<DiameterAgent> Received CCR message from %s:\n%s", c.RemoteAddr(), m))
	var ccr CCR
	if err := m.Unmarshal(&ccr); err != nil {
		utils.Logger.Err(fmt.Sprintf("<DiameterAgent> Unmarshaling message: %s, error: %s", m, err))
		return
	}
	ccr.diamMessage = m // Save it for later searches inside AVPs
	//utils.Logger.Debug(fmt.Sprintf("CCR unrmashaled: %+v", ccr))
	cca := NewCCAFromCCR(&ccr)
	cca.OriginHost = self.cgrCfg.DiameterAgentCfg().OriginHost
	cca.OriginRealm = self.cgrCfg.DiameterAgentCfg().OriginRealm
	cca.GrantedServiceUnit.CCTime = 300
	cca.ResultCode = diam.Success
	if dmtA, err := cca.AsDiameterMessage(); err != nil {
		utils.Logger.Err(fmt.Sprintf("<DiameterAgent> Failed to convert cca as diameter message, error: %s", err.Error()))
		return
	} else if _, err := dmtA.WriteTo(c); err != nil {
		utils.Logger.Err(fmt.Sprintf("<DiameterAgent> Failed to write message to %s: %s\n%s\n", c.RemoteAddr(), err, dmtA))
		return
	}

}

func (self *DiameterAgent) handleALL(c diam.Conn, m *diam.Message) {
	utils.Logger.Warning(fmt.Sprintf("<DiameterAgent> Received unexpected message from %s:\n%s", c.RemoteAddr(), m))
}

func (self *DiameterAgent) ListenAndServe() error {
	return diam.ListenAndServe(self.cgrCfg.DiameterAgentCfg().Listen, self.handlers(), nil)
}
