/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

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
