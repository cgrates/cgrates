// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/ers"
)

func NewErSv1(erS *ers.ERService) *ErSv1 {
	return &ErSv1{erS: erS}
}

type ErSv1 struct {
	erS *ers.ERService
}

// V1RunReader processes files in the configured directory for the given reader. This function handles files
// based on the reader's type and configuration. Only available for readers that are not processing files
// automatically (RunDelay should equal 0).
//
// Note: This API is not safe to call concurrently for the same reader. Ensure the current files finish being
// processed before calling again.
func (eeSv1 *ErSv1) RunReader(ctx *context.Context, params ers.V1RunReaderParams, reply *string) error {
	return eeSv1.erS.V1RunReader(ctx, params, reply)
}
