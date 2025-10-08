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
	"github.com/cgrates/cgrates/analyzers"
)

func NewAnalyzerSv1(aS *analyzers.AnalyzerS) *AnalyzerSv1 {
	return &AnalyzerSv1{aS: aS}
}

type AnalyzerSv1 struct {
	aS *analyzers.AnalyzerS
	ping
}

func (aSv1 *AnalyzerSv1) StringQuery(ctx *context.Context, search *analyzers.QueryArgs, reply *[]map[string]any) error {
	return aSv1.aS.V1StringQuery(ctx, search, reply)
}
