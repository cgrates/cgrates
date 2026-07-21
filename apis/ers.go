// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package apis

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/ers"
)

// NewErSv1 initializes the ErSv1 object.
func NewErSv1(erS *ers.ERService) *ErSv1 {
	return &ErSv1{erS: erS}
}

// ErSv1 represents the RPC object to register for event reader v1 APIs.
type ErSv1 struct {
	erS *ers.ERService
}

// RunReader processes files for the configured reader.
func (s *ErSv1) RunReader(ctx *context.Context, args ers.V1RunReaderParams, reply *string) error {
	return s.erS.V1RunReader(ctx, args, reply)
}
