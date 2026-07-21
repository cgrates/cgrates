// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
func (cdrS *CdrSv1) ProcessEvent(ctx *context.Context, args *utils.CGREvent, reply *string) (err error) {
	return cdrS.cdrs.V1ProcessEvent(ctx, args, reply)
}

// V1ProcessEventWithGet has the same logic with V1ProcessEvent except it adds the proccessed events to the reply
func (cdrS *CdrSv1) ProcessEventWithGet(ctx *context.Context, args *utils.CGREvent, evs *[]*utils.EventsWithOpts) (err error) {
	return cdrS.cdrs.V1ProcessEventWithGet(ctx, args, evs)
}

// V1ProcessStoredEvents processes stored events based on provided filters.
func (cdrS *CdrSv1) ProcessStoredEvents(ctx *context.Context, args *utils.CDRFilters, reply *string) (err error) {
	return cdrS.cdrs.V1ProcessStoredEvents(ctx, args, reply)
}
