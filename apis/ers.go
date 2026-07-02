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
