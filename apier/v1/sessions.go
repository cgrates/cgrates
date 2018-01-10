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
	"github.com/cenk/rpc2"
	"github.com/cgrates/cgrates/sessionmanager"
	"github.com/cgrates/cgrates/utils"
)

func NewSessionSv1(sm *sessionmanager.SMGeneric) *SessionSv1 {
	return &SessionSv1{SMG: sm}
}

// SessionSv1 exports RPC from SessionSv1
type SessionSv1 struct {
	SMG *sessionmanager.SMGeneric
}

// Publishes BiJSONRPC methods exported by SessionSv1
func (ssv1 *SessionSv1) Handlers() map[string]interface{} {
	return map[string]interface{}{
		utils.SessionSv1AuthorizeEvent:   ssv1.BiRpcAuthorizeEvent,
		utils.SessionSv1InitiateSession:  ssv1.BiRpcInitiateSession,
		utils.SessionSv1UpdateSession:    ssv1.BiRpcUpdateSession,
		utils.SessionSv1TerminateSession: ssv1.BiRpcTerminateSession,
		utils.SessionSv1ProcessCDR:       ssv1.BiRpcProcessCDR,
		utils.SessionSv1ProcessEvent:     ssv1.BiRpcProcessEvent,
	}
}

func (ssv1 *SessionSv1) AuthorizeEvent(args *sessionmanager.V1AuthorizeArgs,
	rply *sessionmanager.V1AuthorizeReply) error {
	return ssv1.SMG.BiRPCv1AuthorizeEvent(nil, args, rply)
}

func (ssv1 *SessionSv1) InitiateSession(args *sessionmanager.V1InitSessionArgs,
	rply *sessionmanager.V1InitSessionReply) error {
	return ssv1.SMG.BiRPCv1InitiateSession(nil, args, rply)
}

func (ssv1 *SessionSv1) UpdateSession(args *sessionmanager.V1UpdateSessionArgs,
	rply *sessionmanager.V1UpdateSessionReply) error {
	return ssv1.SMG.BiRPCv1UpdateSession(nil, args, rply)
}

func (ssv1 *SessionSv1) TerminateSession(args *sessionmanager.V1TerminateSessionArgs,
	rply *string) error {
	return ssv1.SMG.BiRPCv1TerminateSession(nil, args, rply)
}

func (ssv1 *SessionSv1) ProcessCDR(cgrEv utils.CGREvent, rply *string) error {
	return ssv1.SMG.BiRPCv1ProcessCDR(nil, cgrEv, rply)
}

func (ssv1 *SessionSv1) ProcessEvent(args *sessionmanager.V1ProcessEventArgs,
	rply *sessionmanager.V1ProcessEventReply) error {
	return ssv1.SMG.BiRPCv1ProcessEvent(nil, args, rply)
}

func (ssv1 *SessionSv1) BiRpcAuthorizeEvent(clnt *rpc2.Client, args *sessionmanager.V1AuthorizeArgs,
	rply *sessionmanager.V1AuthorizeReply) error {
	return ssv1.SMG.BiRPCv1AuthorizeEvent(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRpcInitiateSession(clnt *rpc2.Client, args *sessionmanager.V1InitSessionArgs,
	rply *sessionmanager.V1InitSessionReply) error {
	return ssv1.SMG.BiRPCv1InitiateSession(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRpcUpdateSession(clnt *rpc2.Client, args *sessionmanager.V1UpdateSessionArgs,
	rply *sessionmanager.V1UpdateSessionReply) error {
	return ssv1.SMG.BiRPCv1UpdateSession(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRpcTerminateSession(clnt *rpc2.Client, args *sessionmanager.V1TerminateSessionArgs,
	rply *string) error {
	return ssv1.SMG.BiRPCv1TerminateSession(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRpcProcessCDR(clnt *rpc2.Client, cgrEv utils.CGREvent, rply *string) error {
	return ssv1.SMG.BiRPCv1ProcessCDR(clnt, cgrEv, rply)
}

func (ssv1 *SessionSv1) BiRpcProcessEvent(clnt *rpc2.Client, args *sessionmanager.V1ProcessEventArgs,
	rply *sessionmanager.V1ProcessEventReply) error {
	return ssv1.SMG.BiRPCv1ProcessEvent(clnt, args, rply)
}
