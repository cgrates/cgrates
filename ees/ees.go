/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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

package ees

import (
	"fmt"
	"sync"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// NewERService instantiates the EEService
func NewEEService(cfg *config.CGRConfig, filterS *engine.FilterS,
	stopChan chan struct{}, connMgr *engine.ConnManager) *EEService {
	return &EEService{
		cfg:     cfg,
		filterS: filterS,
		connMgr: connMgr,
		ees:     make(map[string]EventExporter),
	}
}

// EEService is managing the EventExporters
type EEService struct {
	sync.RWMutex
	cfg     *config.CGRConfig
	filterS *engine.FilterS

	connMgr *engine.ConnManager
	ees     map[string]EventExporter // map[eeType]EventExporter
}

// ListenAndServe keeps the service alive
func (eeS *EEService) ListenAndServe(exitChan chan bool) (err error) {
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s>",
		utils.CoreS, utils.EventExporterService))
	e := <-exitChan
	eeS.Shutdown()
	exitChan <- e // put back for the others listening for shutdown request
	return
}

// Shutdown is called to shutdown the service
func (eeS *EEService) Shutdown() (err error) {
	utils.Logger.Info(fmt.Sprintf("<%s> shutdown <%s>", utils.CoreS, utils.EventExporterService))
	return
}

// ProcessEvent will be called each time a new event is received from readers
func (eeS *EEService) V1ProcessEvent(cgrEv *utils.CGREvent) (err error) {
	return
}
