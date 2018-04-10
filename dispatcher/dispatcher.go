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
	"github.com/cgrates/cgrates/utils"
)

// NewDispatcherService initializes a DispatcherService
func NewDispatcherService() (dspS *DispatcherService, err error) {
	dspS = &DispatcherService{}
	return
}

// DispatcherService  is the service handling dispatcher
type DispatcherService struct {
}

// ListenAndServe will initialize the service
func (spS *DispatcherService) ListenAndServe(exitChan chan bool) error {
	utils.Logger.Info("Starting Dispatcher Service")
	e := <-exitChan
	exitChan <- e // put back for the others listening for shutdown request
	return nil
}

// Shutdown is called to shutdown the service
func (spS *DispatcherService) Shutdown() error {
	utils.Logger.Info(fmt.Sprintf("<%s> service shutdown initialized", utils.DispatcherS))
	utils.Logger.Info(fmt.Sprintf("<%s> service shutdown complete", utils.DispatcherS))
	return nil
}
