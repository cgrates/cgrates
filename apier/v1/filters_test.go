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
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

type ccMock struct {
	calls map[string]func(args interface{}, reply interface{}) error
}

func (ccM *ccMock) Call(serviceMethod string, args interface{}, reply interface{}) (err error) {
	if call, has := ccM.calls[serviceMethod]; !has {
		return rpcclient.ErrUnsupporteServiceMethod
	} else {
		return call(args, reply)
	}
}

func TestFiltersSetFilterReloadCache(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	cfg.ApierCfg().CachesConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	expArgs := &utils.AttrReloadCacheWithAPIOpts{
		APIOpts: map[string]interface{}{
			utils.CacheOpt: utils.MetaReload,
		},
		Tenant:    "cgrates.org",
		FilterIDs: []string{"cgrates.org:FLTR_ID"},
	}
	ccM := &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.CacheSv1ReloadCache: func(args, reply interface{}) error {
				if !reflect.DeepEqual(args, expArgs) {
					return fmt.Errorf("expected: <%+v>,\nreceived: <%+v>", utils.ToJSON(expArgs), utils.ToJSON(args))
				}
				return nil
			},
		},
	}
	rpcInternal := make(chan rpcclient.ClientConnector, 1)
	rpcInternal <- ccM
	cM := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): rpcInternal,
	})
	apierSv1 := &APIerSv1{
		Config:      cfg,
		DataManager: dm,
		ConnMgr:     cM,
	}
	arg := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			ID: "FLTR_ID",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Account",
					Values:  []string{"1001"},
				},
			},
		},
		APIOpts: map[string]interface{}{
			utils.CacheOpt: utils.MetaReload,
		},
	}
	var reply string

	if err := apierSv1.SetFilter(arg, &reply); err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}

	attrPrf := &engine.AttributeProfileWithAPIOpts{
		AttributeProfile: &engine.AttributeProfile{
			FilterIDs: []string{"FLTR_ID"},
			ID:        "ATTR_ID",
			Contexts:  []string{utils.MetaAny},
			Weight:    10,
			Attributes: []*engine.Attribute{
				{
					Path:  "*req.Account",
					Value: config.NewRSRParsersMustCompile("1003", ";"),
					Type:  utils.MetaConstant,
				},
			},
		},
		APIOpts: map[string]interface{}{
			utils.CacheOpt: utils.MetaNone,
		},
	}

	if err := apierSv1.SetAttributeProfile(attrPrf, &reply); err != nil {
		t.Error(err)
	}

	thPrf := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			ID:        "THD_ID",
			FilterIDs: []string{"FLTR_ID"},
			MaxHits:   10,
			Weight:    10,
		},
		APIOpts: map[string]interface{}{
			utils.CacheOpt: utils.MetaNone,
		},
	}

	if err := apierSv1.SetThresholdProfile(thPrf, &reply); err != nil {
		t.Error(err)
	}

	rsPrf := &engine.ResourceProfileWithAPIOpts{
		ResourceProfile: &engine.ResourceProfile{
			ID:        "RES_ID",
			FilterIDs: []string{"FLTR_ID"},
			Weight:    10,
		},
		APIOpts: map[string]interface{}{
			utils.CacheOpt: utils.MetaNone,
		},
	}

	if err := apierSv1.SetResourceProfile(rsPrf, &reply); err != nil {
		t.Error(err)
	}

	sqPrf := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			ID:        "SQ_ID",
			FilterIDs: []string{"FLTR_ID"},
			Weight:    10,
		},
		APIOpts: map[string]interface{}{
			utils.CacheOpt: utils.MetaNone,
		},
	}

	if err := apierSv1.SetStatQueueProfile(sqPrf, &reply); err != nil {
		t.Error(err)
	}

	dpPrf := &DispatcherWithAPIOpts{
		DispatcherProfile: &engine.DispatcherProfile{
			ID:         "DP_ID",
			FilterIDs:  []string{"FLTR_ID"},
			Subsystems: []string{utils.MetaAny},
		},
		APIOpts: map[string]interface{}{
			utils.CacheOpt: utils.MetaNone,
		},
	}

	if err := apierSv1.SetDispatcherProfile(dpPrf, &reply); err != nil {
		t.Error(err)
	}

	chgPrf := &ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			ID:        "CHG_ID",
			FilterIDs: []string{"FLTR_ID"},
			RunID:     "runID",
		},
		APIOpts: map[string]interface{}{
			utils.CacheOpt: utils.MetaNone,
		},
	}

	if err := apierSv1.SetChargerProfile(chgPrf, &reply); err != nil {
		t.Error(err)
	}

	arg = &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			ID: "FLTR_ID",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Account",
					Values:  []string{"1002"},
				},
			},
		},
		APIOpts: map[string]interface{}{
			utils.CacheOpt: utils.MetaReload,
		},
	}
	expArgs = &utils.AttrReloadCacheWithAPIOpts{
		APIOpts: map[string]interface{}{
			utils.CacheOpt: utils.MetaReload,
		},
		Tenant:                   "cgrates.org",
		FilterIDs:                []string{"cgrates.org:FLTR_ID"},
		AttributeFilterIndexIDs:  []string{"cgrates.org:*any:*string:*req.Account:1001", "cgrates.org:*any:*string:*req.Account:1002"},
		ChargerFilterIndexIDs:    []string{"cgrates.org:*string:*req.Account:1001", "cgrates.org:*string:*req.Account:1002"},
		DispatcherFilterIndexIDs: []string{"cgrates.org:*any:*string:*req.Account:1001", "cgrates.org:*any:*string:*req.Account:1002"},
		ResourceFilterIndexIDs:   []string{"cgrates.org:*string:*req.Account:1001", "cgrates.org:*string:*req.Account:1002"},
		StatFilterIndexIDs:       []string{"cgrates.org:*string:*req.Account:1001", "cgrates.org:*string:*req.Account:1002"},
		ThresholdFilterIndexIDs:  []string{"cgrates.org:*string:*req.Account:1001", "cgrates.org:*string:*req.Account:1002"},
	}

	if err := apierSv1.SetFilter(arg, &reply); err != nil {
		t.Error(err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestFiltersSetFilterClearCache(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	cfg.ApierCfg().CachesConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	expArgs := &utils.AttrCacheIDsWithAPIOpts{
		APIOpts: map[string]interface{}{
			utils.CacheOpt: utils.MetaClear,
		},
		Tenant:   "cgrates.org",
		CacheIDs: []string{utils.CacheFilters},
	}
	ccM := &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.CacheSv1Clear: func(args, reply interface{}) error {
				sort.Strings(args.(*utils.AttrCacheIDsWithAPIOpts).CacheIDs)
				if !reflect.DeepEqual(args, expArgs) {
					return fmt.Errorf("expected: <%+v>,\nreceived: <%+v>", utils.ToJSON(expArgs), utils.ToJSON(args))
				}
				return nil
			},
		},
	}
	rpcInternal := make(chan rpcclient.ClientConnector, 1)
	rpcInternal <- ccM
	cM := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): rpcInternal,
	})
	apierSv1 := &APIerSv1{
		Config:      cfg,
		DataManager: dm,
		ConnMgr:     cM,
	}
	arg := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			ID: "FLTR_ID",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Account",
					Values:  []string{"1001"},
				},
			},
		},
		APIOpts: map[string]interface{}{
			utils.CacheOpt: utils.MetaClear,
		},
	}
	var reply string

	if err := apierSv1.SetFilter(arg, &reply); err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}

	attrPrf := &engine.AttributeProfileWithAPIOpts{
		AttributeProfile: &engine.AttributeProfile{
			FilterIDs: []string{"FLTR_ID"},
			ID:        "ATTR_ID",
			Contexts:  []string{utils.MetaAny},
			Weight:    10,
			Attributes: []*engine.Attribute{
				{
					Path:  "*req.Account",
					Value: config.NewRSRParsersMustCompile("1003", ";"),
					Type:  utils.MetaConstant,
				},
			},
		},
		APIOpts: map[string]interface{}{
			utils.CacheOpt: utils.MetaNone,
		},
	}

	if err := apierSv1.SetAttributeProfile(attrPrf, &reply); err != nil {
		t.Error(err)
	}

	thPrf := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			ID:        "THD_ID",
			FilterIDs: []string{"FLTR_ID"},
			MaxHits:   10,
			Weight:    10,
		},
		APIOpts: map[string]interface{}{
			utils.CacheOpt: utils.MetaNone,
		},
	}

	if err := apierSv1.SetThresholdProfile(thPrf, &reply); err != nil {
		t.Error(err)
	}

	rsPrf := &engine.ResourceProfileWithAPIOpts{
		ResourceProfile: &engine.ResourceProfile{
			ID:        "RES_ID",
			FilterIDs: []string{"FLTR_ID"},
			Weight:    10,
		},
		APIOpts: map[string]interface{}{
			utils.CacheOpt: utils.MetaNone,
		},
	}

	if err := apierSv1.SetResourceProfile(rsPrf, &reply); err != nil {
		t.Error(err)
	}

	sqPrf := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			ID:        "SQ_ID",
			FilterIDs: []string{"FLTR_ID"},
			Weight:    10,
		},
		APIOpts: map[string]interface{}{
			utils.CacheOpt: utils.MetaNone,
		},
	}

	if err := apierSv1.SetStatQueueProfile(sqPrf, &reply); err != nil {
		t.Error(err)
	}

	dpPrf := &DispatcherWithAPIOpts{
		DispatcherProfile: &engine.DispatcherProfile{
			ID:         "DP_ID",
			FilterIDs:  []string{"FLTR_ID"},
			Subsystems: []string{utils.MetaAny},
		},
		APIOpts: map[string]interface{}{
			utils.CacheOpt: utils.MetaNone,
		},
	}

	if err := apierSv1.SetDispatcherProfile(dpPrf, &reply); err != nil {
		t.Error(err)
	}

	chgPrf := &ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			ID:        "CHG_ID",
			FilterIDs: []string{"FLTR_ID"},
			RunID:     "runID",
		},
		APIOpts: map[string]interface{}{
			utils.CacheOpt: utils.MetaNone,
		},
	}

	if err := apierSv1.SetChargerProfile(chgPrf, &reply); err != nil {
		t.Error(err)
	}

	arg = &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			ID: "FLTR_ID",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Account",
					Values:  []string{"1002"},
				},
			},
		},
		APIOpts: map[string]interface{}{
			utils.CacheOpt: utils.MetaClear,
		},
	}
	expArgs = &utils.AttrCacheIDsWithAPIOpts{
		APIOpts: map[string]interface{}{
			utils.CacheOpt: utils.MetaClear,
		},
		Tenant: "cgrates.org",
		CacheIDs: []string{utils.CacheAttributeFilterIndexes, utils.CacheThresholdFilterIndexes,
			utils.CacheResourceFilterIndexes, utils.CacheStatFilterIndexes,
			utils.CacheChargerFilterIndexes, utils.CacheFilters, utils.CacheDispatcherFilterIndexes},
	}
	sort.Strings(expArgs.CacheIDs)

	if err := apierSv1.SetFilter(arg, &reply); err != nil {
		t.Error(err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}
