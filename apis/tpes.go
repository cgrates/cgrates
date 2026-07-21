// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package apis

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/tpes"
)

func NewTPeSv1(tpes *tpes.TPeS) *TPeSv1 {
	return &TPeSv1{tpes: tpes}
}

type TPeSv1 struct {
	tpes *tpes.TPeS
}

// ExportTariffPlan is the API executed to export tariff plan items
func (tpE *TPeSv1) ExportTariffPlan(ctx *context.Context, args *tpes.ArgsExportTP, reply *[]byte) error {
	return tpE.tpes.V1ExportTariffPlan(ctx, args, reply)
}
