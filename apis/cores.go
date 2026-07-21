// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package apis

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/utils"
)

func NewCoreSv1(cS *cores.CoreS) *CoreSv1 {
	return &CoreSv1{cS: cS}
}

// CoreSv1 exports RPC from RLs
type CoreSv1 struct {
	cS *cores.CoreS
}

func (cS *CoreSv1) Status(ctx *context.Context, params *cores.V1StatusParams, reply *map[string]any) error {
	return cS.cS.V1Status(ctx, params, reply)
}

// Sleep is used to test the concurrent requests mechanism
func (cS *CoreSv1) Sleep(ctx *context.Context, arg *utils.DurationArgs, reply *string) error {
	return cS.cS.V1Sleep(ctx, arg, reply)
}

func (cS *CoreSv1) Shutdown(ctx *context.Context, args *utils.CGREvent, reply *string) error {
	return cS.cS.V1Shutdown(ctx, args, reply)
}

// StartCPUProfiling starts CPU profiling and saves the profile to the specified path.
func (cS *CoreSv1) StartCPUProfiling(ctx *context.Context, args *utils.DirectoryArgs, reply *string) error {
	return cS.cS.V1StartCPUProfiling(ctx, args, reply)
}

// StopCPUProfiling stops CPU Profiling.
func (cS *CoreSv1) StopCPUProfiling(ctx *context.Context, args *utils.TenantWithAPIOpts, reply *string) error {
	return cS.cS.V1StopCPUProfiling(ctx, args, reply)
}

// StartMemoryProfiling starts memory profiling in the specified directory.
func (cS *CoreSv1) StartMemoryProfiling(ctx *context.Context, params cores.MemoryProfilingParams, reply *string) error {
	return cS.cS.V1StartMemoryProfiling(ctx, params, reply)
}

// StopMemoryProfiling stops memory profiling.
func (cS *CoreSv1) StopMemoryProfiling(ctx *context.Context, params utils.TenantWithAPIOpts, reply *string) error {
	return cS.cS.V1StopMemoryProfiling(ctx, params, reply)
}

// DescribeMethods returns descriptors for registered RPC methods.
func (cS *CoreSv1) DescribeMethods(ctx *context.Context, args *utils.DescribeMethodsArgs, reply *[]utils.MethodDescriptor) error {
	return cS.cS.V1DescribeMethods(ctx, args, reply)
}

// Panic is used print the Message sent as a panic.
func (cS *CoreSv1) Panic(ctx *context.Context, args *utils.PanicMessageArgs, reply *string) error {
	return cS.cS.V1Panic(ctx, args, reply)
}
