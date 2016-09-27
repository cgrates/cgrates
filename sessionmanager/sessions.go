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
	"sync"
	"time"

	"github.com/cgrates/cgrates/engine"
)

func NewSessions() *Sessions {
	return &Sessions{
		sessionsMux: new(sync.Mutex),
		guard:       engine.Guardian,
	}
}

type Sessions struct {
	sessions    []*Session
	sessionsMux *sync.Mutex          // Lock the list operations
	guard       *engine.GuardianLock // Used to lock on uuid
}

func (self *Sessions) indexSession(s *Session) {
	self.sessionsMux.Lock()
	self.sessions = append(self.sessions, s)
	self.sessionsMux.Unlock()
}

func (self *Sessions) getSessions() []*Session {
	self.sessionsMux.Lock()
	defer self.sessionsMux.Unlock()
	return self.sessions
}

// Searches and return the session with the specifed uuid
func (self *Sessions) getSession(uuid string) *Session {
	self.sessionsMux.Lock()
	defer self.sessionsMux.Unlock()
	for _, s := range self.sessions {
		if s.eventStart.GetUUID() == uuid {
			return s
		}
	}
	return nil
}

// Remove session from session list, removes all related in case of multiple runs, true if item was found
func (self *Sessions) unindexSession(uuid string) bool {
	self.sessionsMux.Lock()
	defer self.sessionsMux.Unlock()
	for i, ss := range self.sessions {
		if ss.eventStart.GetUUID() == uuid {
			self.sessions = append(self.sessions[:i], self.sessions[i+1:]...)
			return true
		}
	}
	return false
}

func (self *Sessions) removeSession(s *Session, evStop engine.Event) error {
	_, err := self.guard.Guard(func() (interface{}, error) { // Lock it on UUID level
		if !self.unindexSession(s.eventStart.GetUUID()) { // Unreference it early so we avoid concurrency
			return nil, nil // Did not find the session so no need to close it anymore
		}
		if err := s.Close(evStop); err != nil { // Stop loop, refund advanced charges and save the costs deducted so far to database
			return nil, err
		}
		return nil, nil
	}, time.Duration(2)*time.Second, s.eventStart.GetUUID())
	return err
}
