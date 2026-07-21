// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package admins

import (
	"fmt"

	"github.com/cgrates/birpc/context"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// GetCDRs retrieves a list of CDRs matching the specified filters.
func (admS AdminS) V1GetCDRs(ctx *context.Context, args *utils.CDRFilters, reply *[]*utils.CDR) error {
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
func (admS AdminS) V1RemoveCDRs(ctx *context.Context, args *utils.CDRFilters, reply *string) (err error) {
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
