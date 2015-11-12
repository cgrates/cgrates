/*
Real-time Charging System for Telecom & ISP environments
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
	"errors"
	"fmt"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// One session handled by SM
type SMGSession struct {
	eventStart SMGenericEvent // Event which started
	stopDebit  chan struct{}  // Channel to communicate with debit loops when closing the session
	connId     string         // Reference towards connection id on the session manager side.
	runId      string         // Keep a reference for the derived run
	timezone   string
	rater      engine.Connector // Connector to Rater service
	cdrsrv     engine.Connector // Connector to CDRS service
	extconns   *SMGExternalConnections
	cd         *engine.CallDescriptor
	cc         []*engine.CallCost
}

// Called in case of automatic debits
func (self *SMGSession) debitLoop(debitInterval time.Duration) {
	loopIndex := 0
	for {
		select {
		case <-self.stopDebit:
			return
		default:
		}
		if maxDebit, err := self.debit(debitInterval); err != nil {
			utils.Logger.Err(fmt.Sprintf("<SMGeneric> Could not complete debit opperation on session: %s, error: %s", self.eventStart.GetUUID(), err.Error()))
			disconnectReason := SYSTEM_ERROR
			if err.Error() == utils.ErrUnauthorizedDestination.Error() {
				disconnectReason = UNAUTHORIZED_DESTINATION
			}
			if err := self.disconnectSession(disconnectReason); err != nil {
				utils.Logger.Err(fmt.Sprintf("<SMGeneric> Could not disconnect session: %s, error: %s", self.eventStart.GetUUID(), err.Error()))
			}
			return
		} else if maxDebit < debitInterval {
			time.Sleep(maxDebit)
			if err := self.disconnectSession(INSUFFICIENT_FUNDS); err != nil {
				utils.Logger.Err(fmt.Sprintf("<SMGeneric> Could not disconnect session: %s, error: %s", self.eventStart.GetUUID(), err.Error()))
			}
			return
		}
		time.Sleep(debitInterval)
		loopIndex++
	}
}

// Attempts to debit a duration, returns maximum duration which can be debitted or error
func (self *SMGSession) debit(dur time.Duration) (time.Duration, error) {
	return nilDuration, nil
}

// Attempts to refund a duration, error on failure
func (self *SMGSession) refund(dur time.Duration) error {
	return nil
}

// Session has ended, check debits and refund the extra charged duration
func (self *SMGSession) close(endTime time.Time) error {
	return nil
}

// Send disconnect order to remote connection
func (self *SMGSession) disconnectSession(reason string) error {
	type AttrDisconnectSession struct {
		EventStart map[string]interface{}
		Reason     string
	}
	conn := self.extconns.GetConnection(self.connId)
	if conn == nil {
		return ErrConnectionNotFound
	}
	var reply string
	if err := conn.Call("SMGClientV1.DisconnectSession", AttrDisconnectSession{EventStart: self.eventStart, Reason: reason}, &reply); err != nil {
		return err
	} else if reply != utils.OK {
		return errors.New(fmt.Sprintf("Unexpected disconnect reply: %s", reply))
	}
	return nil
}

// Merge the sum of costs and sends it to CDRS for storage
func (self *SMGSession) saveOperations(reason string) error {
	if len(self.cc) == 0 {
		return nil // There are no costs to save, ignore the operation
	}
	firstCC := self.cc[0]
	for _, cc := range self.cc[1:] {
		firstCC.Merge(cc)
	}
	var reply string
	err := self.cdrsrv.LogCallCost(&engine.CallCostLog{
		CgrId:          self.eventStart.GetCgrId(self.timezone),
		Source:         utils.SESSION_MANAGER_SOURCE,
		RunId:          self.runId,
		CallCost:       firstCC,
		CheckDuplicate: true,
	}, &reply)
	// this is a protection against the case when the close event is missed for some reason
	// when the cdr arrives to cdrserver because our callcost is not there it will be rated
	// as postpaid. When the close event finally arives we have to refund everything
	if err != nil {
		if err == utils.ErrExists {
			self.refund(self.cd.TimeEnd.Sub(self.cd.TimeStart)) // Refund entire duration
		} else {
			return err
		}
	}
	return nil
}
