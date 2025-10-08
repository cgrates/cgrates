/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

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
	ping
}

// ExportTariffPlan is the API executed to export tariff plan items
func (tpE *TPeSv1) ExportTariffPlan(ctx *context.Context, args *tpes.ArgsExportTP, reply *[]byte) error {
	return tpE.tpes.V1ExportTariffPlan(ctx, args, reply)
}
