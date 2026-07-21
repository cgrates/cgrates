// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
}

func (aSv1 *AnalyzerSv1) StringQuery(ctx *context.Context, search *analyzers.QueryArgs, reply *[]map[string]any) error {
	return aSv1.aS.V1StringQuery(ctx, search, reply)
}
