// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/dispatchers"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
)

func NewGuardianSv1() *GuardianSv1 {
	return &GuardianSv1{}
}

type GuardianSv1 struct{}

// RemoteLock will lock a key from remote
func (self *GuardianSv1) RemoteLock(attr dispatchers.AttrRemoteLockWithApiKey, reply *string) (err error) {
	*reply = guardian.Guardian.GuardIDs(attr.ReferenceID, attr.Timeout, attr.LockIDs...)
	return
}

// RemoteUnlock will unlock a key from remote based on reference ID
func (self *GuardianSv1) RemoteUnlock(refID dispatchers.AttrRemoteUnlockWithApiKey, reply *[]string) (err error) {
	*reply = guardian.Guardian.UnguardIDs(refID.RefID)
	return
}

// Ping return pong if the service is active
func (self *GuardianSv1) Ping(ign *utils.CGREventWithArgDispatcher, reply *string) error {
	*reply = utils.Pong
	return nil
}

// Call implements birpc.ClientConnector interface for internal RPC
func (self *GuardianSv1) Call(ctx *context.Context, serviceMethod string,
	args any, reply any) error {
	return utils.APIerRPCCall(self, serviceMethod, args, reply)
}
