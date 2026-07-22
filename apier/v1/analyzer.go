// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/analyzers"
)

// NewAnalyzerSv1 initializes AnalyzerSv1
func NewAnalyzerSv1(aS *analyzers.AnalyzerService) *AnalyzerSv1 {
	return &AnalyzerSv1{aS: aS}
}

// AnalyzerSv1 exports RPC from RLs
type AnalyzerSv1 struct {
	aS *analyzers.AnalyzerService
}

// StringQuery returns a list of API that match the query
func (aSv1 *AnalyzerSv1) StringQuery(ctx *context.Context, search *analyzers.QueryArgs, reply *[]map[string]any) error {
	return aSv1.aS.V1StringQuery(ctx, search, reply)
}
