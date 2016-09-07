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
package sessionmanager

import (
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/rpcclient"
)

func NewSMAsterisk(cgrCfg *config.CGRConfig, smg rpcclient.RpcClientConnection) (*SMAsterisk, error) {
	return &SMAsterisk{cgrCfg: cgrCfg, smg: smg}, nil
}

type SMAsterisk struct {
	cgrCfg *config.CGRConfig // Separate from smCfg since there can be multiple
	smg    rpcclient.RpcClientConnection
}

// Called to start the service
func (sma *SMAsterisk) ListenAndServe() error {
	return nil
}

// Called to shutdown the service
func (rls *SMAsterisk) ServiceShutdown() error {
	return nil
}
