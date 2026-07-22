// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/utils"
)

func NewCoreSv1(cS *cores.CoreService) *CoreSv1 {
	return &CoreSv1{cS: cS}
}

// CoreSv1 exports RPC from RLs
type CoreSv1 struct {
	cS *cores.CoreService
}

func (cS *CoreSv1) Status(ctx *context.Context, params *cores.V1StatusParams, reply *map[string]any) error {
	return cS.cS.V1Status(ctx, params, reply)
}

// Sleep is used to test the concurrent requests mechanism
func (cS *CoreSv1) Sleep(ctx *context.Context, args *utils.DurationArgs, reply *string) error {
	return cS.cS.V1Sleep(ctx, args, reply)
}

// StartCPUProfiling is used to start CPUProfiling in the given path
func (cS *CoreSv1) StartCPUProfiling(ctx *context.Context, args *utils.DirectoryArgs, reply *string) error {
	return cS.cS.V1StartCPUProfiling(ctx, args, reply)
}

// StopCPUProfiling is used to stop CPUProfiling. The file should be written on the path
// where the CPUProfiling already started
func (cS *CoreSv1) StopCPUProfiling(ctx *context.Context, args *utils.TenantWithAPIOpts, reply *string) error {
	return cS.cS.V1StopCPUProfiling(ctx, args, reply)
}

// StartMemoryProfiling starts memory profiling in the specified directory.
func (cS *CoreSv1) StartMemoryProfiling(ctx *context.Context, params cores.MemoryProfilingParams, reply *string) error {
	return cS.cS.V1StartMemoryProfiling(ctx, params, reply)
}

// StopMemoryProfiling stops memory profiling and writes the final profile.
func (cS *CoreSv1) StopMemoryProfiling(ctx *context.Context, params utils.TenantWithAPIOpts, reply *string) error {
	return cS.cS.V1StopMemoryProfiling(ctx, params, reply)

}

func (cS *CoreSv1) Panic(ctx *context.Context, args *utils.PanicMessageArgs, reply *string) error {
	return cS.cS.V1Panic(ctx, args, reply)
}
