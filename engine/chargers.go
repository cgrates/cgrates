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

package engine

import (
	"fmt"

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func NewChargerService(dm *DataManager, filterS *FilterS,
	attrS rpcclient.RpcClientConnection,
	strgIdxFlds, prfxIdxFlds *[]string) (*ChargerService, error) {
	return &ChargerService{dm: dm, filterS: filterS,
		attrS:       attrS,
		strgIdxFlds: strgIdxFlds,
		prfxIdxFlds: prfxIdxFlds}, nil
}

type ChargerService struct {
	dm          *DataManager
	filterS     *FilterS
	attrS       rpcclient.RpcClientConnection
	strgIdxFlds *[]string
	prfxIdxFlds *[]string
}

// ListenAndServe will initialize the service
func (cS *ChargerService) ListenAndServe(exitChan chan bool) (err error) {
	utils.Logger.Info(fmt.Sprintf("Starting %s", utils.ChargerS))
	e := <-exitChan
	exitChan <- e
	return
}

// Shutdown is called to shutdown the service
func (cS *ChargerService) Shutdown() (err error) {
	utils.Logger.Info(fmt.Sprintf("<%s> shutdown initialized", utils.ChargerS))
	utils.Logger.Info(fmt.Sprintf("<%s> shutdown complete", utils.ChargerS))
	return
}

type ChargerProfile struct {
	Tenant             string
	ID                 string
	FilterIDs          []string
	ActivationInterval *utils.ActivationInterval // Activation interval
	RunID              string
	AttributeIDs       []string
	Weight             float64
}
