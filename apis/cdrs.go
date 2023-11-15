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

package apis

import (
	"fmt"

	"github.com/cgrates/birpc/context"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// GetCDRs retrieves a list of CDRs matching the specified filters.
func (admS AdminSv1) GetCDRs(ctx *context.Context, args *engine.CDRFilters, reply *[]*engine.CDR) error {
	if args.Tenant == utils.EmptyString {
		args.Tenant = admS.cfg.GeneralCfg().DefaultTenant
	}
	fltrs, err := engine.PrepareFilters(ctx, args.FilterIDs, args.Tenant, admS.dm)
	if err != nil {
		return fmt.Errorf("preparing filters failed: %w", err)
	}
	cdrs, err := admS.storDB.GetCDRs(ctx, fltrs, args.APIOpts)
	if err != nil {
		return fmt.Errorf("retrieving CDRs failed: %w", err)
	}
	*reply = cdrs
	return nil
}

// RemoveCDRs removes CDRs matching the specified filters.
func (admS AdminSv1) RemoveCDRs(ctx *context.Context, args *engine.CDRFilters, reply *string) (err error) {
	if args.Tenant == utils.EmptyString {
		args.Tenant = admS.cfg.GeneralCfg().DefaultTenant
	}
	fltrs, err := engine.PrepareFilters(ctx, args.FilterIDs, args.Tenant, admS.dm)
	if err != nil {
		return fmt.Errorf("preparing filters failed: %w", err)
	}
	if err := admS.storDB.RemoveCDRs(ctx, fltrs); err != nil {
		return fmt.Errorf("removing CDRs failed: %w", err)
	}
	*reply = utils.OK
	return
}

// NewCDRsV1 constructs the RPC Object for CDRsV1
func NewCDRsV1(cdrS *engine.CDRServer) *CDRsV1 {
	return &CDRsV1{cdrS: cdrS}
}

// CDRsV1 Exports RPC from CDRs
type CDRsV1 struct {
	ping
	cdrS *engine.CDRServer
}

// ProcessEvent will process the CGREvent
func (cdrSv1 *CDRsV1) ProcessEvent(ctx *context.Context, args *utils.CGREvent,
	reply *string) error {
	return cdrSv1.cdrS.V1ProcessEvent(ctx, args, reply)
}

// ProcessEventWithGet has the same logic with V1ProcessEvent except it adds the proccessed events to the reply
func (cdrSv1 *CDRsV1) ProcessEventWithGet(ctx *context.Context, args *utils.CGREvent,
	reply *[]*utils.EventsWithOpts) error {
	return cdrSv1.cdrS.V1ProcessEventWithGet(ctx, args, reply)
}

// ProcessStoredEvents processes stored events based on provided filters.
func (cdrSv1 *CDRsV1) ProcessStoredEvents(ctx *context.Context, args *engine.CDRFilters,
	reply *string) error {
	return cdrSv1.cdrS.V1ProcessStoredEvents(ctx, args, reply)
}
