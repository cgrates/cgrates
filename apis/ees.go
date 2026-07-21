// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package apis

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/ees"
	"github.com/cgrates/cgrates/utils"
)

// NewEeSv1 initializes the EeSv1 object.
func NewEeSv1(ees *ees.EeS) *EeSv1 {
	return &EeSv1{ees: ees}
}

// EeSv1 represents the RPC object to register for ips v1 APIs.
type EeSv1 struct {
	ees *ees.EeS
}

func (s *EeSv1) ProcessEvent(ctx *context.Context, cgrEv *utils.CGREventWithEeIDs, rply *map[string]map[string]any) error {
	return s.ees.V1ProcessEvent(ctx, cgrEv, rply)
}

// V1ArchiveEventsInReply should archive the events sent with existing exporters. The zipped content should be returned back as a reply.
func (s *EeSv1) ArchiveEventsInReply(ctx *context.Context, args *ees.ArchiveEventsArgs, reply *[]byte) error {
	return s.ees.V1ArchiveEventsInReply(ctx, args, reply)
}

// V1ResetExporterMetrics resets the metrics for a specific exporter identified by ExporterID.
// Returns utils.ErrNotFound if the exporter is not found in the cache.
func (s *EeSv1) ResetExporterMetrics(ctx *context.Context, params ees.V1ResetExporterMetricsParams, reply *string) error {
	return s.ees.V1ResetExporterMetrics(ctx, params, reply)
}
