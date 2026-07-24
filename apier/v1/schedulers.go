// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// NewSchedulerSv1 returns the API for SchedulerS
func NewSchedulerSv1(cgrcfg *config.CGRConfig) *SchedulerSv1 {
	return &SchedulerSv1{cgrcfg: cgrcfg}
}

// SchedulerSv1 is the RPC object implementing scheduler APIs
type SchedulerSv1 struct {
	cgrcfg *config.CGRConfig
}

// Reload reloads scheduler instructions
func (schdSv1 *SchedulerSv1) Reload(arg *utils.CGREventWithArgDispatcher, reply *string) error {
	schdSv1.cgrcfg.GetReloadChan(config.SCHEDULER_JSN) <- struct{}{}
	*reply = utils.OK
	return nil
}

// Ping returns Pong
func (schdSv1 *SchedulerSv1) Ping(ign *utils.CGREventWithArgDispatcher, reply *string) error {
	*reply = utils.Pong
	return nil
}

// Call implements birpc.ClientConnector interface for internal RPC
func (schdSv1 *SchedulerSv1) Call(ctx *context.Context, serviceMethod string,
	args any, reply any) error {
	return utils.APIerRPCCall(schdSv1, serviceMethod, args, reply)
}
