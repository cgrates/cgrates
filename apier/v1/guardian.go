// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/dispatchers"
	"github.com/cgrates/cgrates/guardian"
)

func NewGuardianSv1() *GuardianSv1 {
	return &GuardianSv1{}
}

type GuardianSv1 struct{}

// RemoteLock will lock a key from remote
func (self *GuardianSv1) RemoteLock(ctx *context.Context, attr *dispatchers.AttrRemoteLockWithAPIOpts, reply *string) (err error) {
	*reply = guardian.Guardian.GuardIDs(attr.ReferenceID, attr.Timeout, attr.LockIDs...)
	return
}

// RemoteUnlock will unlock a key from remote based on reference ID
func (self *GuardianSv1) RemoteUnlock(ctx *context.Context, refID *dispatchers.AttrRemoteUnlockWithAPIOpts, reply *[]string) (err error) {
	*reply = guardian.Guardian.UnguardIDs(refID.RefID)
	return
}
