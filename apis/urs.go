// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package apis

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/ers"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

// URFiltersFromReader used on GetURs
type URFiltersFromReader struct {
	Tenant    string
	FilterIDs []string       // Filter for URs
	ReaderID  string         // ReaderID where the URs will be read from (ERs ReaderID)
	APIOpts   map[string]any // options for pagination
}

// // URFiltersFromExporter used on URs Set or Remove
// type URFiltersFromExporter struct {
// 	URFilters   *utils.URFilters // URs which will overwrite an existing UR or set a new one using EEs
// 	ExporterIDs []string         // ExporterIDs where the URs will be exported to (EEs ExporterIDs)
// }

// GetURs retrieves a list of URs matching the specified filters.
func (admS *AdminSv1) GetURs(ctx *context.Context, args *URFiltersFromReader, reply *[]map[string]*utils.CGREvent) error {
	if args.Tenant == utils.EmptyString {
		args.Tenant = admS.cfg.GeneralCfg().DefaultTenant
	}
	ersConns, err := engine.GetConnIDs(context.TODO(),
		admS.cfg.AdminSCfg().Conns, utils.MetaERs, args.Tenant, nil, nil,
		admS.fltrS)
	if err != nil {
		return err
	}

	ersReplies := new([]*sessions.V1ProcessEventReply)
	err = admS.connMgr.Call(context.TODO(), ersConns,
		utils.ErSv1ProcessReaderEvents, &ers.V1RunReaderParams{
			Tenant:   args.Tenant,
			ReaderID: args.ReaderID,
			Filters:  args.FilterIDs,
			APIOpts:  args.APIOpts,
		}, ersReplies)
	if err != nil {
		return err
	}

	for _, ersReply := range *ersReplies {
		if ersReply.UsageRecords != nil {
			*reply = append(*reply, ersReply.UsageRecords)
		}
	}
	return nil
}

// // SetURs sets or overwrites URs matching the specified filters.
// func (admS *AdminSv1) SetURs(ctx *context.Context, args *URFiltersFromExporter, reply *string) (err error) {
// 	if args.URFilters.Tenant == utils.EmptyString {
// 		args.URFilters.Tenant = admS.cfg.GeneralCfg().DefaultTenant
// 	}
// 	fltrs, err := engine.GetFilters(ctx, args.URFilters.FilterIDs, args.URFilters.Tenant, admS.dm)
// 	if err != nil {
// 		return fmt.Errorf("preparing filters failed: %w", err)
// 	}
// 	if err := admS.dm.SetURs(ctx, fltrs); err != nil {
// 		return fmt.Errorf("removing URs failed: %w", err)
// 	}
// 	*reply = utils.OK
// 	return
// }

// // RemoveURs removes URs matching the specified filters.
// func (admS *AdminSv1) RemoveURs(ctx *context.Context, args *URFiltersFromExporter, reply *string) (err error) {
// 	if args.URFilters.Tenant == utils.EmptyString {
// 		args.URFilters.Tenant = admS.cfg.GeneralCfg().DefaultTenant
// 	}
// 	fltrs, err := engine.GetFilters(ctx, args.URFilters.FilterIDs, args.URFilters.Tenant, admS.dm)
// 	if err != nil {
// 		return fmt.Errorf("preparing filters failed: %w", err)
// 	}
// 	if err := admS.dm.RemoveURs(ctx, fltrs); err != nil {
// 		return fmt.Errorf("removing URs failed: %w", err)
// 	}
// 	*reply = utils.OK
// 	return
// }
