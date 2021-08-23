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
	"strings"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

type AttrGetFilterIndexes struct {
	Tenant      string
	Context     string
	ItemType    string
	FilterType  string
	FilterField string
	FilterValue string
	APIOpts     map[string]interface{}
	utils.Paginator
}

type AttrRemFilterIndexes struct {
	Tenant   string
	Context  string
	ItemType string
	APIOpts  map[string]interface{}
}

func (adms *AdminSv1) RemoveFilterIndexes(ctx *context.Context, arg *AttrRemFilterIndexes, reply *string) (err error) {
	if missing := utils.MissingStructFields(arg, []string{"ItemType"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	tntCtx := tnt
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
	case utils.MetaAccounts:
		arg.ItemType = utils.CacheAccountsFilterIndexes
	case utils.MetaActions:
		arg.ItemType = utils.CacheActionProfilesFilterIndexes
	case utils.MetaRateProfiles:
		arg.ItemType = utils.CacheRateProfilesFilterIndexes
	case utils.MetaRateProfileRates:
		if missing := utils.MissingStructFields(arg, []string{"Context"}); len(missing) != 0 {
			return utils.NewErrMandatoryIeMissing(missing...)
		}
		arg.ItemType = utils.CacheRateFilterIndexes
		tntCtx = utils.ConcatenatedKey(tnt, arg.Context)
	case utils.MetaDispatchers:
		arg.ItemType = utils.CacheDispatcherFilterIndexes
	case utils.MetaAttributes:
		arg.ItemType = utils.CacheAttributeFilterIndexes
	}
	if err = adms.dm.RemoveIndexes(ctx, arg.ItemType, tntCtx, utils.EmptyString); err != nil {
		return
	}
	//generate a loadID for CacheFilterIndexes and store it in database
	if err := adms.dm.SetLoadIDs(ctx,
		map[string]int64{arg.ItemType: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	if err := adms.callCacheForRemoveIndexes(ctx, utils.IfaceAsString(arg.APIOpts[utils.CacheOpt]), arg.Tenant,
		arg.ItemType, []string{utils.MetaAny}, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return
}

func (adms *AdminSv1) GetFilterIndexes(ctx *context.Context, arg *AttrGetFilterIndexes, reply *[]string) (err error) {
	var indexes map[string]utils.StringSet
	var indexedSlice []string
	indexesFilter := make(map[string]utils.StringSet)
	if missing := utils.MissingStructFields(arg, []string{"ItemType"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	tntCtx := tnt
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
	case utils.MetaAccounts:
		arg.ItemType = utils.CacheAccountsFilterIndexes
	case utils.MetaActions:
		arg.ItemType = utils.CacheActionProfilesFilterIndexes
	case utils.MetaRateProfiles:
		arg.ItemType = utils.CacheRateProfilesFilterIndexes
	case utils.MetaRateProfileRates:
		if missing := utils.MissingStructFields(arg, []string{"Context"}); len(missing) != 0 {
			return utils.NewErrMandatoryIeMissing(missing...)
		}
		arg.ItemType = utils.CacheRateFilterIndexes
		tntCtx = utils.ConcatenatedKey(tnt, arg.Context)
	case utils.MetaDispatchers:
		arg.ItemType = utils.CacheDispatcherFilterIndexes
	case utils.MetaAttributes:
		arg.ItemType = utils.CacheAttributeFilterIndexes
	}
	if indexes, err = adms.dm.GetIndexes(ctx,
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
func (adms *AdminSv1) ComputeFilterIndexes(ctx *context.Context, args *utils.ArgsComputeFilterIndexes, reply *string) (err error) {
	transactionID := utils.GenUUID()
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	cacheIDs := make(map[string][]string)

	var indexes utils.StringSet
	//ThresholdProfile Indexes
	if args.ThresholdS {
		cacheIDs[utils.CacheThresholdFilterIndexes] = []string{utils.MetaAny}
		if indexes, err = engine.ComputeIndexes(ctx, adms.dm, tnt, utils.EmptyString, utils.CacheThresholdFilterIndexes,
			nil, transactionID, func(tnt, id, grp string) (*[]string, error) {
				th, e := adms.dm.GetThresholdProfile(ctx, tnt, id, true, false, utils.NonTransactional)
				if e != nil {
					return nil, e
				}
				return utils.SliceStringPointer(utils.CloneStringSlice(th.FilterIDs)), nil
			}, nil); err != nil && err != utils.ErrNotFound {
			return utils.APIErrorHandler(err)
		}
		args.ThresholdS = indexes.Size() != 0
	}
	//StatQueueProfile Indexes
	if args.StatS {
		cacheIDs[utils.CacheStatFilterIndexes] = []string{utils.MetaAny}
		if indexes, err = engine.ComputeIndexes(ctx, adms.dm, tnt, utils.EmptyString, utils.CacheStatFilterIndexes,
			nil, transactionID, func(tnt, id, grp string) (*[]string, error) {
				sq, e := adms.dm.GetStatQueueProfile(ctx, tnt, id, true, false, utils.NonTransactional)
				if e != nil {
					return nil, e
				}
				return utils.SliceStringPointer(utils.CloneStringSlice(sq.FilterIDs)), nil
			}, nil); err != nil && err != utils.ErrNotFound {
			return utils.APIErrorHandler(err)
		}
		args.StatS = indexes.Size() != 0
	}
	//ResourceProfile Indexes
	if args.ResourceS {
		cacheIDs[utils.CacheResourceFilterIndexes] = []string{utils.MetaAny}
		if indexes, err = engine.ComputeIndexes(ctx, adms.dm, tnt, utils.EmptyString, utils.CacheResourceFilterIndexes,
			nil, transactionID, func(tnt, id, grp string) (*[]string, error) {
				rp, e := adms.dm.GetResourceProfile(ctx, tnt, id, true, false, utils.NonTransactional)
				if e != nil {
					return nil, e
				}
				return utils.SliceStringPointer(utils.CloneStringSlice(rp.FilterIDs)), nil
			}, nil); err != nil && err != utils.ErrNotFound {
			return utils.APIErrorHandler(err)
		}
		args.ResourceS = indexes.Size() != 0
	}
	//RouteSProfile Indexes
	if args.RouteS {
		cacheIDs[utils.CacheRouteFilterIndexes] = []string{utils.MetaAny}
		if indexes, err = engine.ComputeIndexes(ctx, adms.dm, tnt, utils.EmptyString, utils.CacheRouteFilterIndexes,
			nil, transactionID, func(tnt, id, grp string) (*[]string, error) {
				rp, e := adms.dm.GetRouteProfile(ctx, tnt, id, true, false, utils.NonTransactional)
				if e != nil {
					return nil, e
				}
				return utils.SliceStringPointer(utils.CloneStringSlice(rp.FilterIDs)), nil
			}, nil); err != nil && err != utils.ErrNotFound {
			return utils.APIErrorHandler(err)
		}
		args.RouteS = indexes.Size() != 0
	}
	//AttributeProfile Indexes
	if args.AttributeS {
		cacheIDs[utils.CacheAttributeFilterIndexes] = []string{utils.MetaAny}
		if indexes, err = engine.ComputeIndexes(ctx, adms.dm, tnt, utils.EmptyString, utils.CacheAttributeFilterIndexes,
			nil, transactionID, func(tnt, id, grp string) (*[]string, error) {
				attr, e := adms.dm.GetAttributeProfile(ctx, tnt, id, true, false, utils.NonTransactional)
				if e != nil {
					return nil, e
				}
				return utils.SliceStringPointer(utils.CloneStringSlice(attr.FilterIDs)), nil
			}, nil); err != nil && err != utils.ErrNotFound {
			return utils.APIErrorHandler(err)
		}
		args.AttributeS = indexes.Size() != 0
	}
	//ChargerProfile  Indexes
	if args.ChargerS {
		cacheIDs[utils.CacheChargerFilterIndexes] = []string{utils.MetaAny}
		if indexes, err = engine.ComputeIndexes(ctx, adms.dm, tnt, utils.EmptyString, utils.CacheChargerFilterIndexes,
			nil, transactionID, func(tnt, id, grp string) (*[]string, error) {
				ch, e := adms.dm.GetChargerProfile(ctx, tnt, id, true, false, utils.NonTransactional)
				if e != nil {
					return nil, e
				}
				return utils.SliceStringPointer(utils.CloneStringSlice(ch.FilterIDs)), nil
			}, nil); err != nil && err != utils.ErrNotFound {
			return utils.APIErrorHandler(err)
		}
		args.ChargerS = indexes.Size() != 0
	}
	//AccountFilter Indexes
	if args.AccountS {
		cacheIDs[utils.CacheAccountsFilterIndexes] = []string{utils.MetaAny}
		if indexes, err = engine.ComputeIndexes(ctx, adms.dm, tnt, utils.EmptyString, utils.CacheAccountsFilterIndexes,
			nil, transactionID, func(tnt, id, grp string) (*[]string, error) {
				acnts, e := adms.dm.GetAccount(ctx, tnt, id)
				if e != nil {
					return nil, e
				}
				return utils.SliceStringPointer(utils.CloneStringSlice(acnts.FilterIDs)), nil
			}, nil); err != nil && err != utils.ErrNotFound {
			return utils.APIErrorHandler(err)
		}
		args.AccountS = indexes.Size() != 0
	}
	//ActionFilter Indexes
	if args.ActionS {
		cacheIDs[utils.CacheActionProfilesFilterIndexes] = []string{utils.MetaAny}
		if indexes, err = engine.ComputeIndexes(ctx, adms.dm, tnt, utils.EmptyString, utils.CacheActionProfilesFilterIndexes,
			nil, transactionID, func(tnt, id, grp string) (*[]string, error) {
				act, e := adms.dm.GetActionProfile(ctx, tnt, id, true, false, utils.NonTransactional)
				if e != nil {
					return nil, e
				}
				return utils.SliceStringPointer(utils.CloneStringSlice(act.FilterIDs)), nil
			}, nil); err != nil && err != utils.ErrNotFound {
			return utils.APIErrorHandler(err)
		}
		args.ActionS = indexes.Size() != 0
	}
	var ratePrf []string
	if args.RateS {
		cacheIDs[utils.CacheRateProfilesFilterIndexes] = []string{utils.MetaAny}
		if indexes, err = engine.ComputeIndexes(ctx, adms.dm, tnt, utils.EmptyString, utils.CacheRateProfilesFilterIndexes,
			nil, transactionID, func(tnt, id, grp string) (*[]string, error) {
				rtPrf, e := adms.dm.GetRateProfile(ctx, tnt, id, true, false, utils.NonTransactional)
				if e != nil {
					return nil, e
				}
				ratePrf = append(ratePrf, utils.ConcatenatedKey(tnt, id))
				rtIds := make([]string, 0, len(rtPrf.Rates))
				for key := range rtPrf.Rates {
					rtIds = append(rtIds, key)
				}
				cacheIDs[utils.CacheRateFilterIndexes] = rtIds
				_, e = engine.ComputeIndexes(ctx, adms.dm, tnt, id, utils.CacheRateFilterIndexes,
					&rtIds, transactionID, func(_, id, _ string) (*[]string, error) {
						return utils.SliceStringPointer(utils.CloneStringSlice(rtPrf.Rates[id].FilterIDs)), nil
					}, nil)
				if e != nil {
					return nil, e
				}
				return utils.SliceStringPointer(utils.CloneStringSlice(rtPrf.FilterIDs)), nil
			}, nil); err != nil {
			return utils.APIErrorHandler(err)
		}
		args.RateS = indexes.Size() != 0
	}
	//DispatcherProfile Indexes
	if args.DispatcherS {
		cacheIDs[utils.CacheDispatcherFilterIndexes] = []string{utils.MetaAny}
		if indexes, err = engine.ComputeIndexes(ctx, adms.dm, tnt, utils.EmptyString, utils.CacheDispatcherFilterIndexes,
			nil, transactionID, func(tnt, id, grp string) (*[]string, error) {
				dsp, e := adms.dm.GetDispatcherProfile(ctx, tnt, id, true, false, utils.NonTransactional)
				if e != nil {
					return nil, e
				}
				return utils.SliceStringPointer(utils.CloneStringSlice(dsp.FilterIDs)), nil
			}, nil); err != nil && err != utils.ErrNotFound {
			return utils.APIErrorHandler(err)
		}
		args.DispatcherS = indexes.Size() != 0
	}

	//Now we move from tmpKey to the right key for each type
	//ThresholdProfile Indexes
	if args.ThresholdS {
		if err = adms.dm.SetIndexes(ctx, utils.CacheThresholdFilterIndexes, tnt, nil, true, transactionID); err != nil {
			return
		}
	}
	//StatQueueProfile Indexes
	if args.StatS {
		if err = adms.dm.SetIndexes(ctx, utils.CacheStatFilterIndexes, tnt, nil, true, transactionID); err != nil {
			return
		}
	}
	//ResourceProfile Indexes
	if args.ResourceS {
		if err = adms.dm.SetIndexes(ctx, utils.CacheResourceFilterIndexes, tnt, nil, true, transactionID); err != nil {
			return
		}
	}
	//RouteProfile Indexes
	if args.RouteS {
		if err = adms.dm.SetIndexes(ctx, utils.CacheRouteFilterIndexes, tnt, nil, true, transactionID); err != nil {
			return
		}
	}
	//AttributeProfile Indexes
	if args.AttributeS {
		if err = adms.dm.SetIndexes(ctx, utils.CacheAttributeFilterIndexes, tnt, nil, true, transactionID); err != nil {
			return
		}
	}
	//ChargerProfile Indexes
	if args.ChargerS {
		if err = adms.dm.SetIndexes(ctx, utils.CacheChargerFilterIndexes, tnt, nil, true, transactionID); err != nil {
			return
		}
	}
	//AccountProfile Indexes
	if args.AccountS {
		if err = adms.dm.SetIndexes(ctx, utils.CacheAccountsFilterIndexes, tnt, nil, true, transactionID); err != nil {
			return err
		}
	}
	//ActionProfile Indexes
	if args.ActionS {
		if err = adms.dm.SetIndexes(ctx, utils.CacheActionProfilesFilterIndexes, tnt, nil, true, transactionID); err != nil {
			return err
		}
	}
	//RateProfile Indexes
	if args.RateS {
		if err = adms.dm.SetIndexes(ctx, utils.CacheRateProfilesFilterIndexes, tnt, nil, true, transactionID); err != nil {
			return err
		}
		for _, tntId := range ratePrf {
			if err = adms.dm.SetIndexes(ctx, utils.CacheRateFilterIndexes, tntId, nil, true, transactionID); err != nil {
				return err
			}

		}
	}
	//DispatcherProfile Indexes
	if args.DispatcherS {
		if err = adms.dm.SetIndexes(ctx, utils.CacheDispatcherFilterIndexes, tnt, nil, true, transactionID); err != nil {
			return
		}
	}
	//generate a load
	//ID for CacheFilterIndexes and store it in database
	loadIDs := make(map[string]int64)
	timeNow := time.Now().UnixNano()
	for idx := range cacheIDs {
		loadIDs[idx] = timeNow
	}
	if err := adms.dm.SetLoadIDs(ctx, loadIDs); err != nil {
		return utils.APIErrorHandler(err)
	}
	if err := adms.callCacheForComputeIndexes(ctx, utils.IfaceAsString(args.APIOpts[utils.CacheOpt]),
		args.Tenant, cacheIDs, args.APIOpts); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// ComputeFilterIndexIDs computes specific filter indexes
func (adms *AdminSv1) ComputeFilterIndexIDs(ctx *context.Context, args *utils.ArgsComputeFilterIndexIDs, reply *string) (err error) {
	transactionID := utils.NonTransactional
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	var indexes utils.StringSet
	cacheIDs := make(map[string][]string)
	//ThresholdProfile Indexes
	if indexes, err = engine.ComputeIndexes(ctx, adms.dm, tnt, utils.EmptyString, utils.CacheThresholdFilterIndexes,
		&args.ThresholdIDs, transactionID, func(tnt, id, grp string) (*[]string, error) {
			th, e := adms.dm.GetThresholdProfile(ctx, tnt, id, true, false, utils.NonTransactional)
			if e != nil {
				return nil, e
			}
			return utils.SliceStringPointer(utils.CloneStringSlice(th.FilterIDs)), nil
		}, nil); err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	if indexes.Size() != 0 {
		cacheIDs[utils.CacheThresholdFilterIndexes] = indexes.AsSlice()
	}
	//StatQueueProfile Indexes
	if indexes, err = engine.ComputeIndexes(ctx, adms.dm, tnt, utils.EmptyString, utils.CacheStatFilterIndexes,
		&args.StatIDs, transactionID, func(tnt, id, grp string) (*[]string, error) {
			sq, e := adms.dm.GetStatQueueProfile(ctx, tnt, id, true, false, utils.NonTransactional)
			if e != nil {
				return nil, e
			}
			cacheIDs[utils.CacheStatFilterIndexes] = []string{sq.ID}
			return utils.SliceStringPointer(utils.CloneStringSlice(sq.FilterIDs)), nil
		}, nil); err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	if indexes.Size() != 0 {
		cacheIDs[utils.CacheStatFilterIndexes] = indexes.AsSlice()
	}
	//ResourceProfile Indexes
	if indexes, err = engine.ComputeIndexes(ctx, adms.dm, tnt, utils.EmptyString, utils.CacheResourceFilterIndexes,
		&args.ResourceIDs, transactionID, func(tnt, id, grp string) (*[]string, error) {
			rp, e := adms.dm.GetResourceProfile(ctx, tnt, id, true, false, utils.NonTransactional)
			if e != nil {
				return nil, e
			}
			cacheIDs[utils.CacheResourceFilterIndexes] = []string{rp.ID}
			return utils.SliceStringPointer(utils.CloneStringSlice(rp.FilterIDs)), nil
		}, nil); err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	if indexes.Size() != 0 {
		cacheIDs[utils.CacheResourceFilterIndexes] = indexes.AsSlice()
	}
	//RouteProfile Indexes
	if indexes, err = engine.ComputeIndexes(ctx, adms.dm, tnt, utils.EmptyString, utils.CacheRouteFilterIndexes,
		&args.RouteIDs, transactionID, func(tnt, id, grp string) (*[]string, error) {
			rp, e := adms.dm.GetRouteProfile(ctx, tnt, id, true, false, utils.NonTransactional)
			if e != nil {
				return nil, e
			}
			cacheIDs[utils.CacheRouteFilterIndexes] = []string{rp.ID}
			return utils.SliceStringPointer(utils.CloneStringSlice(rp.FilterIDs)), nil
		}, nil); err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	if indexes.Size() != 0 {
		cacheIDs[utils.CacheRouteFilterIndexes] = indexes.AsSlice()
	}
	//AttributeProfile Indexes
	if indexes, err = engine.ComputeIndexes(ctx, adms.dm, tnt, utils.EmptyString, utils.CacheAttributeFilterIndexes,
		&args.AttributeIDs, transactionID, func(tnt, id, grp string) (*[]string, error) {
			attr, e := adms.dm.GetAttributeProfile(ctx, tnt, id, true, false, utils.NonTransactional)
			if e != nil {
				return nil, e
			}
			return utils.SliceStringPointer(utils.CloneStringSlice(attr.FilterIDs)), nil
		}, nil); err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	if indexes.Size() != 0 {
		cacheIDs[utils.CacheAttributeFilterIndexes] = indexes.AsSlice()
	}
	//ChargerProfile  Indexes
	if indexes, err = engine.ComputeIndexes(ctx, adms.dm, tnt, utils.EmptyString, utils.CacheChargerFilterIndexes,
		&args.ChargerIDs, transactionID, func(tnt, id, grp string) (*[]string, error) {
			ch, e := adms.dm.GetChargerProfile(ctx, tnt, id, true, false, utils.NonTransactional)
			if e != nil {
				return nil, e
			}
			return utils.SliceStringPointer(utils.CloneStringSlice(ch.FilterIDs)), nil
		}, nil); err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	if indexes.Size() != 0 {
		cacheIDs[utils.CacheChargerFilterIndexes] = indexes.AsSlice()
	}
	//AccountIndexes
	if indexes, err = engine.ComputeIndexes(ctx, adms.dm, tnt, utils.EmptyString, utils.CacheAccountsFilterIndexes,
		&args.AccountIDs, transactionID, func(tnt, id, grp string) (*[]string, error) {
			acc, e := adms.dm.GetAccount(ctx, tnt, id)
			if e != nil {
				return nil, e
			}
			return utils.SliceStringPointer(utils.CloneStringSlice(acc.FilterIDs)), nil
		}, nil); err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	if indexes.Size() != 0 {
		cacheIDs[utils.CacheAccountsFilterIndexes] = indexes.AsSlice()
	}
	//ActionProfile Indexes
	if indexes, err = engine.ComputeIndexes(ctx, adms.dm, tnt, utils.EmptyString, utils.CacheActionProfilesFilterIndexes,
		&args.ActionProfileIDs, transactionID, func(tnt, id, grp string) (*[]string, error) {
			act, e := adms.dm.GetActionProfile(ctx, tnt, id, true, false, utils.NonTransactional)
			if e != nil {
				return nil, e
			}
			return utils.SliceStringPointer(utils.CloneStringSlice(act.FilterIDs)), nil
		}, nil); err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	if indexes.Size() != 0 {
		cacheIDs[utils.CacheActionProfilesFilterIndexes] = indexes.AsSlice()
	}
	//RateProfile Indexes
	if _, err = engine.ComputeIndexes(ctx, adms.dm, tnt, utils.EmptyString, utils.CacheRateProfilesFilterIndexes,
		&args.RateProfileIDs, transactionID, func(tnt, id, grp string) (*[]string, error) {
			rpr, e := adms.dm.GetRateProfile(ctx, tnt, id, true, false, utils.NonTransactional)
			if e != nil {
				return nil, e
			}
			rtIds := make([]string, 0, len(rpr.Rates))
			for key := range rpr.Rates {
				rtIds = append(rtIds, key)
			}
			indexesRate, e := engine.ComputeIndexes(ctx, adms.dm, tnt, id, utils.CacheRateFilterIndexes,
				&rtIds, transactionID, func(_, id, _ string) (*[]string, error) {
					return utils.SliceStringPointer(utils.CloneStringSlice(rpr.Rates[id].FilterIDs)), nil
				}, nil)
			if e != nil {
				return nil, e
			}
			if indexesRate.Size() != 0 {
				cacheIDs[utils.CacheRateFilterIndexes] = indexesRate.AsSlice()
			}
			return utils.SliceStringPointer(utils.CloneStringSlice(rpr.FilterIDs)), nil
		}, nil); err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	if indexes.Size() != 0 {
		cacheIDs[utils.CacheRateProfilesFilterIndexes] = indexes.AsSlice()
	}
	//DispatcherProfile Indexes
	if indexes, err = engine.ComputeIndexes(ctx, adms.dm, tnt, utils.EmptyString, utils.CacheDispatcherFilterIndexes,
		&args.DispatcherIDs, transactionID, func(tnt, id, grp string) (*[]string, error) {
			dsp, e := adms.dm.GetDispatcherProfile(ctx, tnt, id, true, false, utils.NonTransactional)
			if e != nil {
				return nil, e
			}
			return utils.SliceStringPointer(utils.CloneStringSlice(dsp.FilterIDs)), nil
		}, nil); err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	if indexes.Size() != 0 {
		cacheIDs[utils.CacheDispatcherFilterIndexes] = indexes.AsSlice()
	}

	loadIDs := make(map[string]int64)
	timeNow := time.Now().UnixNano()
	for idx := range cacheIDs {
		loadIDs[idx] = timeNow
	}
	if err := adms.dm.SetLoadIDs(ctx, loadIDs); err != nil {
		return utils.APIErrorHandler(err)
	}
	if err := adms.callCacheForComputeIndexes(ctx, utils.IfaceAsString(args.APIOpts[utils.CacheOpt]),
		args.Tenant, cacheIDs, args.APIOpts); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (adms *AdminSv1) GetReverseFilterHealth(ctx *context.Context, args *engine.IndexHealthArgs, reply *map[string]*engine.ReverseFilterIHReply) (err error) {
	objCaches := map[string]*ltcache.Cache{utils.CacheRateFilterIndexes: ltcache.NewCache(-1, 0, false, nil)}
	for indxType := range utils.CacheIndexesToPrefix {
		objCaches[indxType] = ltcache.NewCache(-1, 0, false, nil)
	}

	*reply, err = engine.GetRevFltrIdxHealth(ctx, adms.dm,
		ltcache.NewCache(args.FilterCacheLimit, args.FilterCacheTTL, args.FilterCacheStaticTTL, nil),
		ltcache.NewCache(args.IndexCacheLimit, args.IndexCacheTTL, args.IndexCacheStaticTTL, nil),
		objCaches,
	)
	return
}

func (adms *AdminSv1) GetThresholdsIndexesHealth(ctx *context.Context, args *engine.IndexHealthArgs, reply *engine.FilterIHReply) error {
	rp, err := engine.GetFltrIdxHealth(ctx, adms.dm,
		ltcache.NewCache(args.FilterCacheLimit, args.FilterCacheTTL, args.FilterCacheStaticTTL, nil),
		ltcache.NewCache(args.IndexCacheLimit, args.IndexCacheTTL, args.IndexCacheStaticTTL, nil),
		ltcache.NewCache(args.ObjectCacheLimit, args.ObjectCacheTTL, args.ObjectCacheStaticTTL, nil),
		utils.CacheThresholdFilterIndexes,
	)
	if err != nil {
		return err
	}
	*reply = *rp
	return nil
}

func (adms *AdminSv1) GetResourcesIndexesHealth(ctx *context.Context, args *engine.IndexHealthArgs, reply *engine.FilterIHReply) error {
	rp, err := engine.GetFltrIdxHealth(ctx, adms.dm,
		ltcache.NewCache(args.FilterCacheLimit, args.FilterCacheTTL, args.FilterCacheStaticTTL, nil),
		ltcache.NewCache(args.IndexCacheLimit, args.IndexCacheTTL, args.IndexCacheStaticTTL, nil),
		ltcache.NewCache(args.ObjectCacheLimit, args.ObjectCacheTTL, args.ObjectCacheStaticTTL, nil),
		utils.CacheResourceFilterIndexes,
	)
	if err != nil {
		return err
	}
	*reply = *rp
	return nil
}

func (adms *AdminSv1) GetStatsIndexesHealth(ctx *context.Context, args *engine.IndexHealthArgs, reply *engine.FilterIHReply) error {
	rp, err := engine.GetFltrIdxHealth(ctx, adms.dm,
		ltcache.NewCache(args.FilterCacheLimit, args.FilterCacheTTL, args.FilterCacheStaticTTL, nil),
		ltcache.NewCache(args.IndexCacheLimit, args.IndexCacheTTL, args.IndexCacheStaticTTL, nil),
		ltcache.NewCache(args.ObjectCacheLimit, args.ObjectCacheTTL, args.ObjectCacheStaticTTL, nil),
		utils.CacheStatFilterIndexes,
	)
	if err != nil {
		return err
	}
	*reply = *rp
	return nil
}

func (adms *AdminSv1) GetRoutesIndexesHealth(ctx *context.Context, args *engine.IndexHealthArgs, reply *engine.FilterIHReply) error {
	rp, err := engine.GetFltrIdxHealth(ctx, adms.dm,
		ltcache.NewCache(args.FilterCacheLimit, args.FilterCacheTTL, args.FilterCacheStaticTTL, nil),
		ltcache.NewCache(args.IndexCacheLimit, args.IndexCacheTTL, args.IndexCacheStaticTTL, nil),
		ltcache.NewCache(args.ObjectCacheLimit, args.ObjectCacheTTL, args.ObjectCacheStaticTTL, nil),
		utils.CacheRouteFilterIndexes,
	)
	if err != nil {
		return err
	}
	*reply = *rp
	return nil
}

func (adms *AdminSv1) GetAttributesIndexesHealth(ctx *context.Context, args *engine.IndexHealthArgs, reply *engine.FilterIHReply) error {
	rp, err := engine.GetFltrIdxHealth(ctx, adms.dm,
		ltcache.NewCache(args.FilterCacheLimit, args.FilterCacheTTL, args.FilterCacheStaticTTL, nil),
		ltcache.NewCache(args.IndexCacheLimit, args.IndexCacheTTL, args.IndexCacheStaticTTL, nil),
		ltcache.NewCache(args.ObjectCacheLimit, args.ObjectCacheTTL, args.ObjectCacheStaticTTL, nil),
		utils.CacheAttributeFilterIndexes,
	)
	if err != nil {
		return err
	}
	*reply = *rp
	return nil
}

func (adms *AdminSv1) GetChargersIndexesHealth(ctx *context.Context, args *engine.IndexHealthArgs, reply *engine.FilterIHReply) error {
	rp, err := engine.GetFltrIdxHealth(ctx, adms.dm,
		ltcache.NewCache(args.FilterCacheLimit, args.FilterCacheTTL, args.FilterCacheStaticTTL, nil),
		ltcache.NewCache(args.IndexCacheLimit, args.IndexCacheTTL, args.IndexCacheStaticTTL, nil),
		ltcache.NewCache(args.ObjectCacheLimit, args.ObjectCacheTTL, args.ObjectCacheStaticTTL, nil),
		utils.CacheChargerFilterIndexes,
	)
	if err != nil {
		return err
	}
	*reply = *rp
	return nil
}

func (adms *AdminSv1) GetDispatchersIndexesHealth(ctx *context.Context, args *engine.IndexHealthArgs, reply *engine.FilterIHReply) error {
	rp, err := engine.GetFltrIdxHealth(ctx, adms.dm,
		ltcache.NewCache(args.FilterCacheLimit, args.FilterCacheTTL, args.FilterCacheStaticTTL, nil),
		ltcache.NewCache(args.IndexCacheLimit, args.IndexCacheTTL, args.IndexCacheStaticTTL, nil),
		ltcache.NewCache(args.ObjectCacheLimit, args.ObjectCacheTTL, args.ObjectCacheStaticTTL, nil),
		utils.CacheDispatcherFilterIndexes,
	)
	if err != nil {
		return err
	}
	*reply = *rp
	return nil
}

func (adms *AdminSv1) GetRateProfilesIndexesHealth(ctx *context.Context, args *engine.IndexHealthArgs, reply *engine.FilterIHReply) error {
	rp, err := engine.GetFltrIdxHealth(ctx, adms.dm,
		ltcache.NewCache(args.FilterCacheLimit, args.FilterCacheTTL, args.FilterCacheStaticTTL, nil),
		ltcache.NewCache(args.IndexCacheLimit, args.IndexCacheTTL, args.IndexCacheStaticTTL, nil),
		ltcache.NewCache(args.ObjectCacheLimit, args.ObjectCacheTTL, args.ObjectCacheStaticTTL, nil),
		utils.CacheRateProfilesFilterIndexes,
	)
	if err != nil {
		return err
	}
	*reply = *rp
	return nil
}

func (adms *AdminSv1) GetActionsIndexesHealth(ctx *context.Context, args *engine.IndexHealthArgs, reply *engine.FilterIHReply) error {
	rp, err := engine.GetFltrIdxHealth(ctx, adms.dm,
		ltcache.NewCache(args.FilterCacheLimit, args.FilterCacheTTL, args.FilterCacheStaticTTL, nil),
		ltcache.NewCache(args.IndexCacheLimit, args.IndexCacheTTL, args.IndexCacheStaticTTL, nil),
		ltcache.NewCache(args.ObjectCacheLimit, args.ObjectCacheTTL, args.ObjectCacheStaticTTL, nil),
		utils.CacheActionProfilesFilterIndexes,
	)
	if err != nil {
		return err
	}
	*reply = *rp
	return nil
}

func (adms *AdminSv1) GetAccountsIndexesHealth(ctx *context.Context, args *engine.IndexHealthArgs, reply *engine.FilterIHReply) error {
	rp, err := engine.GetFltrIdxHealth(ctx, adms.dm,
		ltcache.NewCache(args.FilterCacheLimit, args.FilterCacheTTL, args.FilterCacheStaticTTL, nil),
		ltcache.NewCache(args.IndexCacheLimit, args.IndexCacheTTL, args.IndexCacheStaticTTL, nil),
		ltcache.NewCache(args.ObjectCacheLimit, args.ObjectCacheTTL, args.ObjectCacheStaticTTL, nil),
		utils.CacheAccountsFilterIndexes,
	)
	if err != nil {
		return err
	}
	*reply = *rp
	return nil
}

func (adms *AdminSv1) GetRateRatesIndexesHealth(ctx *context.Context, args *engine.IndexHealthArgs, reply *engine.FilterIHReply) error {
	rp, err := engine.GetFltrIdxHealthForRateRates(ctx, adms.dm,
		ltcache.NewCache(args.FilterCacheLimit, args.FilterCacheTTL, args.FilterCacheStaticTTL, nil),
		ltcache.NewCache(args.IndexCacheLimit, args.IndexCacheTTL, args.IndexCacheStaticTTL, nil),
		ltcache.NewCache(args.ObjectCacheLimit, args.ObjectCacheTTL, args.ObjectCacheStaticTTL, nil),
	)
	if err != nil {
		return err
	}
	*reply = *rp
	return nil
}
