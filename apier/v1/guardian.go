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

package v1

import (
	"time"

	"github.com/cgrates/cgrates/guardian"
)

type GuardianSv1 struct{}

type AttrRemoteLock struct {
	ReferenceID string        // reference ID for this lock if available
	LockIDs     []string      // List of IDs to obtain lock for
	Timeout     time.Duration // Automatically unlock on timeout
}

// RemoteLock will lock a key from remote
func (self *GuardianSv1) RemoteLock(attr AttrRemoteLock, reply *string) (err error) {
	*reply = guardian.Guardian.GuardIDs(attr.ReferenceID, attr.Timeout, attr.LockIDs...)
	return
}

// RemoteUnlock will unlock a key from remote based on reference ID
func (self *GuardianSv1) RemoteUnlock(refID string, reply *[]string) (err error) {
	*reply = guardian.Guardian.UnguardIDs(refID)
	return
}
