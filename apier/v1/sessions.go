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
	"github.com/cgrates/cgrates/sessionmanager"
	"github.com/cgrates/cgrates/utils"
)

func NewSessionSv1(sm *sessionmanager.SMGeneric) *SessionSv1 {
	return &SessionSv1{SMG: sm}
}

// Exports RPC from SessionSv1
type SessionSv1 struct {
	SMG *sessionmanager.SMGeneric
}

// Publishes BiJSONRPC methods exported by SessionSv1
func (ssv1 *SessionSv1) Handlers() map[string]interface{} {
	return map[string]interface{}{
		"SessionSv1.InitiateSession":  ssv1.SMG.BiRPCv1InitiateSession,
		"SessionSv1.UpdateSession":    ssv1.SMG.BiRPCv1UpdateSession,
		"SessionSv1.TerminateSession": ssv1.SMG.BiRPCv1TerminateSession,
		"SessionSv1.ProcessCDR":       ssv1.SMG.BiRPCv1ProcessCDR,
	}
}

// Called on session start, returns the maximum number of seconds the session can last
func (ssv1 *SessionSv1) InitiateSession(args *sessionmanager.V1InitSessionArgs,
	rply *sessionmanager.V1InitSessionReply) error {
	return ssv1.SMG.BiRPCv1InitiateSession(nil, args, rply)
}

// Interim updates, returns remaining duration from the rater
func (ssv1 *SessionSv1) UpdateSession(args *sessionmanager.V1UpdateSessionArgs,
	rply *sessionmanager.V1UpdateSessionReply) error {
	return ssv1.SMG.BiRPCv1UpdateSession(nil, args, rply)
}

// Called on session end, should stop debit loop
func (ssv1 *SessionSv1) TerminateSession(args *sessionmanager.V1TerminateSessionArgs,
	rply *string) error {
	return ssv1.SMG.BiRPCv1TerminateSession(nil, args, rply)
}

// Called on session end, should stop debit loop
func (ssv1 *SessionSv1) ProcessCDR(cgrEv utils.CGREvent, rply *string) error {
	return ssv1.SMG.BiRPCv1ProcessCDR(nil, cgrEv, rply)
}
