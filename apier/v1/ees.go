// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/ees"
	"github.com/cgrates/cgrates/engine"
)

func NewEeSv1(eeS *ees.EventExporterS) *EeSv1 {
	return &EeSv1{eeS: eeS}
}

type EeSv1 struct {
	eeS *ees.EventExporterS
}

// ProcessEvent triggers exports on EEs side
func (eeSv1 *EeSv1) ProcessEvent(ctx *context.Context, args *engine.CGREventWithEeIDs,
	reply *map[string]map[string]any) error {
	return eeSv1.eeS.V1ProcessEvent(ctx, args, reply)
}

// V1ResetExporterMetrics resets the metrics for a specific exporter identified by ExporterID.
// Returns utils.ErrNotFound if the exporter is not found in the cache.
func (eeSv1 *EeSv1) ResetExporterMetrics(ctx *context.Context, params ees.V1ResetExporterMetricsParams, reply *string) error {
	return eeSv1.eeS.V1ResetExporterMetrics(ctx, params, reply)
}
