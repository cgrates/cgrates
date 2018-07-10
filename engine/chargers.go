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

// ChargerProfile is the config for one Charger
type ChargerProfile struct {
	Tenant             string
	ID                 string
	FilterIDs          []string
	ActivationInterval *utils.ActivationInterval // Activation interval
	RunID              string
	AttributeIDs       []string // perform data aliasing based on these Attributes
	Weight             float64
}

func NewChargerService(dm *DataManager, filterS *FilterS,
	attrS rpcclient.RpcClientConnection,
	strgIdxFlds, prfxIdxFlds *[]string) (*ChargerService, error) {
	return &ChargerService{dm: dm, filterS: filterS,
		attrS:       attrS,
		strgIdxFlds: strgIdxFlds,
		prfxIdxFlds: prfxIdxFlds}, nil
}

// ChargerService is performing charging
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

func (cS *ChargerService) processEvent(cgrEv *utils.CGREvent) (cgrEvs []*utils.CGREvent, err error) {
	return
}

// V1ProcessEvent will process the event received via API and return list of events forked
func (cS *ChargerService) V1ProcessEvent(args *utils.CGREvent,
	reply *[]*utils.CGREvent) (err error) {
	if args.Event == nil {
		return utils.NewErrMandatoryIeMissing("Event")
	}
	rply, err := cS.processEvent(args)
	if err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = rply
	return

}
