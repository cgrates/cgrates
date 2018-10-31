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

package dispatchers

import (
	"fmt"
	"reflect"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewDispatcherService initializes a DispatcherService
func NewDispatcherService(dm *engine.DataManager, rals, resS, thdS,
	statS, splS, attrS, sessionS, chargerS rpcclient.RpcClientConnection) (*DispatcherService, error) {
	if rals != nil && reflect.ValueOf(rals).IsNil() {
		rals = nil
	}
	if resS != nil && reflect.ValueOf(resS).IsNil() {
		resS = nil
	}
	if thdS != nil && reflect.ValueOf(thdS).IsNil() {
		thdS = nil
	}
	if statS != nil && reflect.ValueOf(statS).IsNil() {
		statS = nil
	}
	if splS != nil && reflect.ValueOf(splS).IsNil() {
		splS = nil
	}
	if attrS != nil && reflect.ValueOf(attrS).IsNil() {
		attrS = nil
	}
	if sessionS != nil && reflect.ValueOf(sessionS).IsNil() {
		sessionS = nil
	}
	if chargerS != nil && reflect.ValueOf(chargerS).IsNil() {
		chargerS = nil
	}
	return &DispatcherService{dm: dm,
		rals:     rals,
		resS:     resS,
		thdS:     thdS,
		statS:    statS,
		splS:     splS,
		attrS:    attrS,
		sessionS: sessionS,
		chargerS: chargerS}, nil
}

// DispatcherService  is the service handling dispatcher
type DispatcherService struct {
	dm       *engine.DataManager
	rals     rpcclient.RpcClientConnection // RALs connections
	resS     rpcclient.RpcClientConnection // ResourceS connections
	thdS     rpcclient.RpcClientConnection // ThresholdS connections
	statS    rpcclient.RpcClientConnection // StatS connections
	splS     rpcclient.RpcClientConnection // SupplierS connections
	attrS    rpcclient.RpcClientConnection // AttributeS connections
	sessionS rpcclient.RpcClientConnection // SessionS server connections
	chargerS rpcclient.RpcClientConnection // ChargerS server connections
}

// ListenAndServe will initialize the service
func (dS *DispatcherService) ListenAndServe(exitChan chan bool) error {
	utils.Logger.Info("Starting Dispatcher service")
	e := <-exitChan
	exitChan <- e // put back for the others listening for shutdown request
	return nil
}

// Shutdown is called to shutdown the service
func (dS *DispatcherService) Shutdown() error {
	utils.Logger.Info(fmt.Sprintf("<%s> service shutdown initialized", utils.DispatcherS))
	utils.Logger.Info(fmt.Sprintf("<%s> service shutdown complete", utils.DispatcherS))
	return nil
}

func (dS *DispatcherService) authorizeEvent(ev *utils.CGREvent,
	reply *engine.AttrSProcessEventReply) (err error) {
	if dS.attrS == nil {
		return utils.NewErrNotConnected(utils.AttributeS)
	}
	if err = dS.attrS.Call(utils.AttributeSv1ProcessEvent, ev, reply); err != nil {
		if err.Error() == utils.ErrNotFound.Error() {
			err = utils.ErrUnknownApiKey
		}
		return
	}
	return
}

func (dS *DispatcherService) authorize(method, tenant, apiKey string, evTime *time.Time) (err error) {
	if apiKey == "" {
		return utils.NewErrMandatoryIeMissing(utils.APIKey)
	}
	ev := &utils.CGREvent{
		Tenant:  tenant,
		ID:      utils.UUIDSha1Prefix(),
		Context: utils.StringPointer(utils.MetaAuth),
		Time:    evTime,
		Event: map[string]interface{}{
			utils.APIKey: apiKey,
		},
	}
	var rplyEv engine.AttrSProcessEventReply
	if err = dS.authorizeEvent(ev, &rplyEv); err != nil {
		return
	}
	var apiMethods string
	if apiMethods, err = rplyEv.CGREvent.FieldAsString(utils.APIMethods); err != nil {
		return
	}
	if !ParseStringMap(apiMethods).HasKey(method) {
		return utils.ErrUnauthorizedApi
	}
	return
}
