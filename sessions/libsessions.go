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

package sessions

import (
	"math/rand"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// SessionSClient is the interface implemented by Agents which are able to
// communicate bidirectionally with SessionS and remote Communication Switch
type SessionSClient interface {
	Call(serviceMethod string, args interface{}, reply interface{}) error
	V1DisconnectSession(args utils.AttrDisconnectSession, reply *string) (err error)
	V1GetActiveSessionIDs(ignParam string, sessionIDs *[]*SessionID) (err error)
}

// getSessionTTL retrieves SessionTTL setting out of ev
// if SessionTTLMaxDelay is present in ev, the return is randomized
func getSessionTTL(ev *engine.SafEvent, cfgSessionTTL time.Duration,
	cfgSessionTTLMaxDelay *time.Duration) (ttl time.Duration, err error) {
	if ttl, err = ev.GetDuration(utils.SessionTTL); err != nil {
		if err != utils.ErrNotFound {
			return
		}
		err = nil
		ttl = cfgSessionTTL
	}
	if ttl == 0 {
		return
	}
	// random delay computation
	var sessionTTLMaxDelay int64
	maxDelay, err := ev.GetDuration(utils.SessionTTLMaxDelay)
	if err != nil {
		if err != utils.ErrNotFound {
			return
		}
		err = nil // clear the error for return
		if cfgSessionTTLMaxDelay != nil {
			maxDelay = *cfgSessionTTLMaxDelay
		}
	}
	sessionTTLMaxDelay = maxDelay.Nanoseconds() / 1000000 // Milliseconds precision for randomness
	if sessionTTLMaxDelay != 0 {
		rand.Seed(time.Now().Unix())
		ttl += time.Duration(rand.Int63n(sessionTTLMaxDelay) * 1000000)
	}
	return
}

func GetSetCGRID(ev *engine.SafEvent) (cgrID string) {
	cgrID = ev.GetStringIgnoreErrors(utils.CGRID)
	if cgrID == "" {
		cgrID = utils.Sha1(ev.GetStringIgnoreErrors(utils.OriginID),
			ev.GetStringIgnoreErrors(utils.OriginHost))
		ev.Set(utils.CGRID, cgrID)
	}
	return
}
