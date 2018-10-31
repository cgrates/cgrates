/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNEtS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package v1

import (
	"github.com/cgrates/cgrates/analyzers"
	"github.com/cgrates/cgrates/utils"
)

// NewAnalyzerSv1 initializes AnalyzerSv1
func NewAnalyzerSv1(aS *analyzers.AnalyzerService) *AnalyzerSv1 {
	return &AnalyzerSv1{aS: aS}
}

// Exports RPC from RLs
type AnalyzerSv1 struct {
	aS *analyzers.AnalyzerService
}

// Call implements rpcclient.RpcClientConnection interface for internal RPC
func (aSv1 *AnalyzerSv1) Call(serviceMethod string,
	args interface{}, reply interface{}) error {
	return utils.APIerRPCCall(aSv1, serviceMethod, args, reply)
}

// Ping return pong if the service is active
func (alSv1 *AnalyzerSv1) Ping(ign string, reply *string) error {
	*reply = utils.Pong
	return nil
}
