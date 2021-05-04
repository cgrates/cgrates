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
	"strings"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type AttrGetFilterIndexes struct {
	Tenant      string
	Context     string
	ItemType    string
	FilterType  string
	FilterField string
	FilterValue string
	utils.Paginator
}

type AttrRemFilterIndexes struct {
	Tenant   string
	Context  string
	ItemType string
}

func (apierSv1 *APIerSv1) RemoveFilterIndexes(arg *AttrRemFilterIndexes, reply *string) (err error) {
	if missing := utils.MissingStructFields(arg, []string{"ItemType"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	tntCtx := arg.Tenant
	switch arg.ItemType {
	case utils.MetaThresholds:
		arg.ItemType = utils.CacheThresholdFilterIndexes
	case utils.MetaRoutes:
		arg.ItemType = utils.CacheRouteFilterIndexes
	case utils.MetaStats:
		arg.ItemType = utils.CacheStatFilterIndexes
	case utils.MetaResources:
		arg.ItemType = utils.CacheResourceFilterIndexes
	case utils.MetaChargers:
		arg.ItemType = utils.CacheChargerFilterIndexes
	case utils.MetaDispatchers:
		if missing := utils.MissingStructFields(arg, []string{"Context"}); len(missing) != 0 { //Params missing
			return utils.NewErrMandatoryIeMissing(missing...)
		}
		arg.ItemType = utils.CacheDispatcherFilterIndexes
		tntCtx = utils.ConcatenatedKey(tnt, arg.Context)
	case utils.MetaAttributes:
		if missing := utils.MissingStructFields(arg, []string{"Context"}); len(missing) != 0 { //Params missing
			return utils.NewErrMandatoryIeMissing(missing...)
		}
		arg.ItemType = utils.CacheAttributeFilterIndexes
		tntCtx = utils.ConcatenatedKey(tnt, arg.Context)
	}
	if err = apierSv1.DataManager.RemoveIndexes(arg.ItemType, tntCtx, utils.EmptyString); err != nil {
		return
	}
	*reply = utils.OK
	return
}

func (apierSv1 *APIerSv1) GetFilterIndexes(arg *AttrGetFilterIndexes, reply *[]string) (err error) {
	var indexes map[string]utils.StringSet
	var indexedSlice []string
	indexesFilter := make(map[string]utils.StringSet)
	if missing := utils.MissingStructFields(arg, []string{"ItemType"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	tntCtx := arg.Tenant
	switch arg.ItemType {
	case utils.MetaThresholds:
		arg.ItemType = utils.CacheThresholdFilterIndexes
	case utils.MetaRoutes:
		arg.ItemType = utils.CacheRouteFilterIndexes
	case utils.MetaStats:
		arg.ItemType = utils.CacheStatFilterIndexes
	case utils.MetaResources:
		arg.ItemType = utils.CacheResourceFilterIndexes
	case utils.MetaChargers:
		arg.ItemType = utils.CacheChargerFilterIndexes
	case utils.MetaDispatchers:
		if missing := utils.MissingStructFields(arg, []string{"Context"}); len(missing) != 0 { //Params missing
			return utils.NewErrMandatoryIeMissing(missing...)
		}
		arg.ItemType = utils.CacheDispatcherFilterIndexes
		tntCtx = utils.ConcatenatedKey(tnt, arg.Context)
	case utils.MetaAttributes:
		if missing := utils.MissingStructFields(arg, []string{"Context"}); len(missing) != 0 { //Params missing
			return utils.NewErrMandatoryIeMissing(missing...)
		}
		arg.ItemType = utils.CacheAttributeFilterIndexes
		tntCtx = utils.ConcatenatedKey(tnt, arg.Context)
	}
	if indexes, err = apierSv1.DataManager.GetIndexes(
		arg.ItemType, tntCtx, utils.EmptyString, true, true); err != nil {
		return
	}
	if arg.FilterType != utils.EmptyString {
		for val, strmap := range indexes {
			if strings.HasPrefix(val, arg.FilterType) {
				indexesFilter[val] = strmap
				for _, value := range strmap.AsSlice() {
					indexedSlice = append(indexedSlice, utils.ConcatenatedKey(val, value))
				}
			}
		}
		if len(indexedSlice) == 0 {
			return utils.ErrNotFound
		}
	}
	if arg.FilterField != utils.EmptyString {
		if len(indexedSlice) == 0 {
			indexesFilter = make(map[string]utils.StringSet)
			for val, strmap := range indexes {
				if strings.Index(val, arg.FilterField) != -1 {
					indexesFilter[val] = strmap
					for _, value := range strmap.AsSlice() {
						indexedSlice = append(indexedSlice, utils.ConcatenatedKey(val, value))
					}
				}
			}
			if len(indexedSlice) == 0 {
				return utils.ErrNotFound
			}
		} else {
			var cloneIndexSlice []string
			for val, strmap := range indexesFilter {
				if strings.Index(val, arg.FilterField) != -1 {
					for _, value := range strmap.AsSlice() {
						cloneIndexSlice = append(cloneIndexSlice, utils.ConcatenatedKey(val, value))
					}
				}
			}
			if len(cloneIndexSlice) == 0 {
				return utils.ErrNotFound
			}
			indexedSlice = cloneIndexSlice
		}
	}
	if arg.FilterValue != utils.EmptyString {
		if len(indexedSlice) == 0 {
			for val, strmap := range indexes {
				if strings.Index(val, arg.FilterValue) != -1 {
					for _, value := range strmap.AsSlice() {
						indexedSlice = append(indexedSlice, utils.ConcatenatedKey(val, value))
					}
				}
			}
			if len(indexedSlice) == 0 {
				return utils.ErrNotFound
			}
		} else {
			var cloneIndexSlice []string
			for val, strmap := range indexesFilter {
				if strings.Index(val, arg.FilterValue) != -1 {
					for _, value := range strmap.AsSlice() {
						cloneIndexSlice = append(cloneIndexSlice, utils.ConcatenatedKey(val, value))
					}
				}
			}
			if len(cloneIndexSlice) == 0 {
				return utils.ErrNotFound
			}
			indexedSlice = cloneIndexSlice
		}
	}
	if len(indexedSlice) == 0 {
		for val, strmap := range indexes {
			for _, value := range strmap.AsSlice() {
				indexedSlice = append(indexedSlice, utils.ConcatenatedKey(val, value))
			}
		}
	}
	if arg.Paginator.Limit != nil || arg.Paginator.Offset != nil {
		*reply = arg.Paginator.PaginateStringSlice(indexedSlice)
	} else {
		*reply = indexedSlice
	}
	return nil
}

// ComputeFilterIndexes selects which index filters to recompute
func (apierSv1 *APIerSv1) ComputeFilterIndexes(args *utils.ArgsComputeFilterIndexes, reply *string) (err error) {
	transactionID := utils.GenUUID()
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}

	//ThresholdProfile Indexes
	if args.ThresholdS {
		if args.ThresholdS, err = engine.ComputeIndexes(apierSv1.DataManager, tnt, args.Context, utils.CacheThresholdFilterIndexes,
			nil, transactionID, func(tnt, id, ctx string) (*[]string, error) {
				th, e := apierSv1.DataManager.GetThresholdProfile(tnt, id, true, false, utils.NonTransactional)
				if e != nil {
					return nil, e
				}
				fltrIDs := make([]string, len(th.FilterIDs))
				for i, fltrID := range th.FilterIDs {
					fltrIDs[i] = fltrID
				}
				return &fltrIDs, nil
			}, nil); err != nil && err != utils.ErrNotFound {
			return utils.APIErrorHandler(err)
		}
	}
	//StatQueueProfile Indexes
	if args.StatS {
		if args.StatS, err = engine.ComputeIndexes(apierSv1.DataManager, tnt, args.Context, utils.CacheStatFilterIndexes,
			nil, transactionID, func(tnt, id, ctx string) (*[]string, error) {
				sq, e := apierSv1.DataManager.GetStatQueueProfile(tnt, id, true, false, utils.NonTransactional)
				if e != nil {
					return nil, e
				}
				fltrIDs := make([]string, len(sq.FilterIDs))
				for i, fltrID := range sq.FilterIDs {
					fltrIDs[i] = fltrID
				}
				return &fltrIDs, nil
			}, nil); err != nil && err != utils.ErrNotFound {
			return utils.APIErrorHandler(err)
		}
	}
	//ResourceProfile Indexes
	if args.ResourceS {
		if args.ResourceS, err = engine.ComputeIndexes(apierSv1.DataManager, tnt, args.Context, utils.CacheResourceFilterIndexes,
			nil, transactionID, func(tnt, id, ctx string) (*[]string, error) {
				rp, e := apierSv1.DataManager.GetResourceProfile(tnt, id, true, false, utils.NonTransactional)
				if e != nil {
					return nil, e
				}
				fltrIDs := make([]string, len(rp.FilterIDs))
				for i, fltrID := range rp.FilterIDs {
					fltrIDs[i] = fltrID
				}
				return &fltrIDs, nil
			}, nil); err != nil && err != utils.ErrNotFound {
			return utils.APIErrorHandler(err)
		}
	}
	//SupplierProfile Indexes
	if args.RouteS {
		if args.RouteS, err = engine.ComputeIndexes(apierSv1.DataManager, tnt, args.Context, utils.CacheRouteFilterIndexes,
			nil, transactionID, func(tnt, id, ctx string) (*[]string, error) {
				rp, e := apierSv1.DataManager.GetRouteProfile(tnt, id, true, false, utils.NonTransactional)
				if e != nil {
					return nil, e
				}
				fltrIDs := make([]string, len(rp.FilterIDs))
				for i, fltrID := range rp.FilterIDs {
					fltrIDs[i] = fltrID
				}
				return &fltrIDs, nil
			}, nil); err != nil && err != utils.ErrNotFound {
			return utils.APIErrorHandler(err)
		}
	}
	//AttributeProfile Indexes
	if args.AttributeS {
		if args.AttributeS, err = engine.ComputeIndexes(apierSv1.DataManager, tnt, args.Context, utils.CacheAttributeFilterIndexes,
			nil, transactionID, func(tnt, id, ctx string) (*[]string, error) {
				ap, e := apierSv1.DataManager.GetAttributeProfile(tnt, id, true, false, utils.NonTransactional)
				if e != nil {
					return nil, e
				}
				if !utils.IsSliceMember(ap.Contexts, ctx) {
					return nil, nil
				}
				fltrIDs := make([]string, len(ap.FilterIDs))
				for i, fltrID := range ap.FilterIDs {
					fltrIDs[i] = fltrID
				}

				return &fltrIDs, nil
			}, nil); err != nil && err != utils.ErrNotFound {
			return utils.APIErrorHandler(err)
		}
	}
	//ChargerProfile  Indexes
	if args.ChargerS {
		if args.ChargerS, err = engine.ComputeIndexes(apierSv1.DataManager, tnt, args.Context, utils.CacheChargerFilterIndexes,
			nil, transactionID, func(tnt, id, ctx string) (*[]string, error) {
				ap, e := apierSv1.DataManager.GetChargerProfile(tnt, id, true, false, utils.NonTransactional)
				if e != nil {
					return nil, e
				}
				fltrIDs := make([]string, len(ap.FilterIDs))
				for i, fltrID := range ap.FilterIDs {
					fltrIDs[i] = fltrID
				}
				return &fltrIDs, nil
			}, nil); err != nil && err != utils.ErrNotFound {
			return utils.APIErrorHandler(err)
		}
	}
	//DispatcherProfile Indexes
	if args.DispatcherS {
		if args.DispatcherS, err = engine.ComputeIndexes(apierSv1.DataManager, tnt, args.Context, utils.CacheDispatcherFilterIndexes,
			nil, transactionID, func(tnt, id, ctx string) (*[]string, error) {
				dsp, e := apierSv1.DataManager.GetDispatcherProfile(tnt, id, true, false, utils.NonTransactional)
				if e != nil {
					return nil, e
				}
				if !utils.IsSliceMember(dsp.Subsystems, ctx) {
					return nil, nil
				}
				fltrIDs := make([]string, len(dsp.FilterIDs))
				for i, fltrID := range dsp.FilterIDs {
					fltrIDs[i] = fltrID
				}
				return &fltrIDs, nil
			}, nil); err != nil && err != utils.ErrNotFound {
			return utils.APIErrorHandler(err)
		}
	}

	tntCtx := args.Tenant
	if args.Context != utils.EmptyString {
		tntCtx = utils.ConcatenatedKey(tnt, args.Context)
	}
	//Now we move from tmpKey to the right key for each type
	//ThresholdProfile Indexes
	if args.ThresholdS {
		if err = apierSv1.DataManager.SetIndexes(utils.CacheThresholdFilterIndexes, tnt, nil, true, transactionID); err != nil {
			return
		}
	}
	//StatQueueProfile Indexes
	if args.StatS {
		if err = apierSv1.DataManager.SetIndexes(utils.CacheStatFilterIndexes, tnt, nil, true, transactionID); err != nil {
			return
		}
	}
	//ResourceProfile Indexes
	if args.ResourceS {
		if err = apierSv1.DataManager.SetIndexes(utils.CacheResourceFilterIndexes, tnt, nil, true, transactionID); err != nil {
			return
		}
	}
	//RouteProfile Indexes
	if args.RouteS {
		if err = apierSv1.DataManager.SetIndexes(utils.CacheRouteFilterIndexes, tnt, nil, true, transactionID); err != nil {
			return
		}
	}
	//AttributeProfile Indexes
	if args.AttributeS {
		if err = apierSv1.DataManager.SetIndexes(utils.CacheAttributeFilterIndexes, tntCtx, nil, true, transactionID); err != nil {
			return
		}
	}
	//ChargerProfile Indexes
	if args.ChargerS {
		if err = apierSv1.DataManager.SetIndexes(utils.CacheChargerFilterIndexes, tnt, nil, true, transactionID); err != nil {
			return
		}
	}
	//DispatcherProfile Indexes
	if args.DispatcherS {
		if err = apierSv1.DataManager.SetIndexes(utils.CacheDispatcherFilterIndexes, tntCtx, nil, true, transactionID); err != nil {
			return
		}
	}
	*reply = utils.OK
	return nil
}

// ComputeFilterIndexIDs computes specific filter indexes
func (apierSv1 *APIerSv1) ComputeFilterIndexIDs(args *utils.ArgsComputeFilterIndexIDs, reply *string) (err error) {
	transactionID := utils.NonTransactional
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	//ThresholdProfile Indexes
	if _, err = engine.ComputeIndexes(apierSv1.DataManager, tnt, args.Context, utils.CacheThresholdFilterIndexes,
		&args.ThresholdIDs, transactionID, func(tnt, id, ctx string) (*[]string, error) {
			th, e := apierSv1.DataManager.GetThresholdProfile(tnt, id, true, false, utils.NonTransactional)
			if e != nil {
				return nil, e
			}
			fltrIDs := make([]string, len(th.FilterIDs))
			for i, fltrID := range th.FilterIDs {
				fltrIDs[i] = fltrID
			}
			return &fltrIDs, nil
		}, nil); err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	//StatQueueProfile Indexes
	if _, err = engine.ComputeIndexes(apierSv1.DataManager, tnt, args.Context, utils.CacheStatFilterIndexes,
		&args.StatIDs, transactionID, func(tnt, id, ctx string) (*[]string, error) {
			sq, e := apierSv1.DataManager.GetStatQueueProfile(tnt, id, true, false, utils.NonTransactional)
			if e != nil {
				return nil, e
			}
			fltrIDs := make([]string, len(sq.FilterIDs))
			for i, fltrID := range sq.FilterIDs {
				fltrIDs[i] = fltrID
			}
			return &fltrIDs, nil
		}, nil); err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	//ResourceProfile Indexes
	if _, err = engine.ComputeIndexes(apierSv1.DataManager, tnt, args.Context, utils.CacheResourceFilterIndexes,
		&args.ResourceIDs, transactionID, func(tnt, id, ctx string) (*[]string, error) {
			rp, e := apierSv1.DataManager.GetResourceProfile(tnt, id, true, false, utils.NonTransactional)
			if e != nil {
				return nil, e
			}
			fltrIDs := make([]string, len(rp.FilterIDs))
			for i, fltrID := range rp.FilterIDs {
				fltrIDs[i] = fltrID
			}
			return &fltrIDs, nil
		}, nil); err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	//RouteProfile Indexes
	if _, err = engine.ComputeIndexes(apierSv1.DataManager, tnt, args.Context, utils.CacheRouteFilterIndexes,
		&args.RouteIDs, transactionID, func(tnt, id, ctx string) (*[]string, error) {
			rp, e := apierSv1.DataManager.GetRouteProfile(tnt, id, true, false, utils.NonTransactional)
			if e != nil {
				return nil, e
			}
			fltrIDs := make([]string, len(rp.FilterIDs))
			for i, fltrID := range rp.FilterIDs {
				fltrIDs[i] = fltrID
			}
			return &fltrIDs, nil
		}, nil); err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	//AttributeProfile Indexes
	if _, err = engine.ComputeIndexes(apierSv1.DataManager, tnt, args.Context, utils.CacheAttributeFilterIndexes,
		&args.AttributeIDs, transactionID, func(tnt, id, ctx string) (*[]string, error) {
			ap, e := apierSv1.DataManager.GetAttributeProfile(tnt, id, true, false, utils.NonTransactional)
			if e != nil {
				return nil, e
			}
			if !utils.IsSliceMember(ap.Contexts, ctx) {
				return nil, nil
			}
			fltrIDs := make([]string, len(ap.FilterIDs))
			for i, fltrID := range ap.FilterIDs {
				fltrIDs[i] = fltrID
			}
			return &fltrIDs, nil
		}, nil); err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	//ChargerProfile  Indexes
	if _, err = engine.ComputeIndexes(apierSv1.DataManager, tnt, args.Context, utils.CacheChargerFilterIndexes,
		&args.ChargerIDs, transactionID, func(tnt, id, ctx string) (*[]string, error) {
			ap, e := apierSv1.DataManager.GetChargerProfile(tnt, id, true, false, utils.NonTransactional)
			if e != nil {
				return nil, e
			}
			fltrIDs := make([]string, len(ap.FilterIDs))
			for i, fltrID := range ap.FilterIDs {
				fltrIDs[i] = fltrID
			}
			return &fltrIDs, nil
		}, nil); err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	//DispatcherProfile Indexes
	if _, err = engine.ComputeIndexes(apierSv1.DataManager, tnt, args.Context, utils.CacheDispatcherFilterIndexes,
		&args.DispatcherIDs, transactionID, func(tnt, id, ctx string) (*[]string, error) {
			dsp, e := apierSv1.DataManager.GetDispatcherProfile(tnt, id, true, false, utils.NonTransactional)
			if e != nil {
				return nil, e
			}
			if !utils.IsSliceMember(dsp.Subsystems, ctx) {
				return nil, nil
			}
			fltrIDs := make([]string, len(dsp.FilterIDs))
			for i, fltrID := range dsp.FilterIDs {
				fltrIDs[i] = fltrID
			}
			return &fltrIDs, nil
		}, nil); err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}
