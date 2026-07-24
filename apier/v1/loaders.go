// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/loaders"
	"github.com/cgrates/cgrates/utils"
)

func NewLoaderSv1(ldrS *loaders.LoaderService) *LoaderSv1 {
	return &LoaderSv1{ldrS: ldrS}
}

// Exports RPC from LoaderService
type LoaderSv1 struct {
	ldrS *loaders.LoaderService
}

// Call implements birpc.ClientConnector interface for internal RPC
func (ldrSv1 *LoaderSv1) Call(ctx *context.Context, serviceMethod string,
	args any, reply any) error {
	return utils.APIerRPCCall(ldrSv1, serviceMethod, args, reply)
}

func (ldrSv1 *LoaderSv1) Load(args *loaders.ArgsProcessFolder,
	rply *string) error {
	return ldrSv1.ldrS.V1Load(args, rply)
}

func (ldrSv1 *LoaderSv1) Remove(args *loaders.ArgsProcessFolder,
	rply *string) error {
	return ldrSv1.ldrS.V1Remove(args, rply)
}

func (rsv1 *LoaderSv1) Ping(ign *utils.CGREventWithArgDispatcher, reply *string) error {
	*reply = utils.Pong
	return nil
}
