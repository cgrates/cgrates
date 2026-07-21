// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package apis

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/loaders"
)

func NewLoaderSv1(ldrS *loaders.LoaderS) *LoaderSv1 {
	return &LoaderSv1{ldrS: ldrS}
}

// LoaderSv1 Exports RPC from LoaderService
type LoaderSv1 struct {
	ldrS *loaders.LoaderS
}

func (ldrSv1 *LoaderSv1) Run(ctx *context.Context, args *loaders.ArgsProcessFolder,
	rply *string) error {
	return ldrSv1.ldrS.V1Run(ctx, args, rply)
}

func (ldrSv1 *LoaderSv1) ImportZip(ctx *context.Context, args *loaders.ArgsProcessZip,
	rply *string) error {
	return ldrSv1.ldrS.V1ImportZip(ctx, args, rply)
}
