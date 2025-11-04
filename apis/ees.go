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
