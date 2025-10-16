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
	"fmt"

	"github.com/cgrates/birpc/context"

	"github.com/cgrates/cgrates/cdrs"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// GetCDRs retrieves a list of CDRs matching the specified filters.
func (admS AdminSv1) GetCDRs(ctx *context.Context, args *utils.CDRFilters, reply *[]*utils.CDR) error {
	if args.Tenant == utils.EmptyString {
		args.Tenant = admS.cfg.GeneralCfg().DefaultTenant
	}
	fltrs, err := engine.GetFilters(ctx, args.FilterIDs, args.Tenant, admS.dm)
	if err != nil {
		return fmt.Errorf("preparing filters failed: %w", err)
	}
	cdrs, err := admS.dm.GetCDRs(ctx, fltrs, args.APIOpts)
	if err != nil {
		return fmt.Errorf("retrieving CDRs failed: %w", err)
	}
	*reply = cdrs
	return nil
}

// RemoveCDRs removes CDRs matching the specified filters.
func (admS AdminSv1) RemoveCDRs(ctx *context.Context, args *utils.CDRFilters, reply *string) (err error) {
	if args.Tenant == utils.EmptyString {
		args.Tenant = admS.cfg.GeneralCfg().DefaultTenant
	}
	fltrs, err := engine.GetFilters(ctx, args.FilterIDs, args.Tenant, admS.dm)
	if err != nil {
		return fmt.Errorf("preparing filters failed: %w", err)
	}
	if err := admS.dm.RemoveCDRs(ctx, fltrs); err != nil {
		return fmt.Errorf("removing CDRs failed: %w", err)
	}
	*reply = utils.OK
	return
}

// NewCdrSv1 initializes the CdrSv1 object.
func NewCdrSv1(cdrs *cdrs.CDRServer) *CdrSv1 {
	return &CdrSv1{cdrs: cdrs}
}

// CdrSv1 represents the RPC object to register for cdrs v1 APIs.
type CdrSv1 struct {
	cdrs *cdrs.CDRServer
}

// V1ProcessEvent will process the CGREvent
func (cdrS *CdrSv1) V1ProcessEvent(ctx *context.Context, args *utils.CGREvent, reply *string) (err error) {
	return cdrS.cdrs.V1ProcessEvent(ctx, args, reply)
}

// V1ProcessEventWithGet has the same logic with V1ProcessEvent except it adds the proccessed events to the reply
func (cdrS *CdrSv1) V1ProcessEventWithGet(ctx *context.Context, args *utils.CGREvent, evs *[]*utils.EventsWithOpts) (err error) {
	return cdrS.cdrs.V1ProcessEventWithGet(ctx, args, evs)
}

// V1ProcessStoredEvents processes stored events based on provided filters.
func (cdrS *CdrSv1) V1ProcessStoredEvents(ctx *context.Context, args *utils.CDRFilters, reply *string) (err error) {
	return cdrS.cdrs.V1ProcessStoredEvents(ctx, args, reply)
}
