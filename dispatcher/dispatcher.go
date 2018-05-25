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

package dispatcher

import (
	"fmt"
	"reflect"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewDispatcherService initializes a DispatcherService
func NewDispatcherService(dm *engine.DataManager, rals, resS, thdS,
	statS, splS, attrS, sessionS rpcclient.RpcClientConnection) (*DispatcherService, error) {
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
	return &DispatcherService{dm: dm,
		rals:     rals,
		resS:     resS,
		thdS:     thdS,
		statS:    statS,
		splS:     splS,
		attrS:    attrS,
		sessionS: sessionS}, nil
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
}

// ListenAndServe will initialize the service
func (dS *DispatcherService) ListenAndServe(exitChan chan bool) error {
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

func (dS *DispatcherService) ThresholdSv1Ping(ign string, reply *string) error {
	if dS.thdS != nil {
		if err := dS.thdS.Call(utils.ThresholdSv1Ping, ign, reply); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<DispatcherS> error: %s ThresholdS.", err.Error()))
		}
	}
	return nil
}

func (dS *DispatcherService) ThresholdSv1GetThresholdIDs(tenant string, tIDs *[]string) error {
	if dS.thdS != nil {
		if err := dS.thdS.Call(utils.ThresholdSv1GetThresholdIDs, tenant, tIDs); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<DispatcherS> error: %s ThresholdS.", err.Error()))
		}
	}
	return nil
}

func (dS *DispatcherService) ThresholdSv1GetThreshold(tntID *utils.TenantID, t *engine.Threshold) error {
	if dS.thdS != nil {
		if err := dS.thdS.Call(utils.ThresholdSv1GetThreshold, tntID, t); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<DispatcherS> error: %s ThresholdS.", err.Error()))
		}
	}
	return nil
}

func (dS *DispatcherService) ThresholdSv1ProcessEvent(args *engine.ArgsProcessEvent, tIDs *[]string) error {
	if dS.thdS != nil {
		if err := dS.thdS.Call(utils.ThresholdSv1ProcessEvent, args, tIDs); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<DispatcherS> error: %s ThresholdS.", err.Error()))
		}
	}
	return nil
}

func (dS *DispatcherService) StatSv1Ping(ign string, reply *string) error {
	if dS.statS != nil {
		if err := dS.statS.Call(utils.StatSv1Ping, ign, reply); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<DispatcherS> error: %s StatS.", err.Error()))
		}
	}
	return nil
}
