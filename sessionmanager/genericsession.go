/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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
	"time"

	"github.com/cgrates/cgrates/engine"
	//"github.com/cgrates/cgrates/utils"
)

// One session handled by SM
type GenericSession struct {
	eventStart SMGenericEvent // Event which started
	stopDebit  chan struct{}  // Channel to communicate with debit loops when closing the session
	connId     string         // Reference towards connection id on the session manager side.
	runId      string         // Keep a reference for the derived run
	cd         *engine.CallDescriptor
	cc         []*engine.CallCost
}

func (self *GenericSession) debitLoop(debitInterval time.Duration) {
	loopIndex := 0
	for {
		select {
		case <-self.stopDebit:
			return
		default:
		}
		time.Sleep(debitInterval)
		loopIndex++
	}
}
