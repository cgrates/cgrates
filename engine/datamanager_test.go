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
package engine

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestDmSetSupplierProfileRpl(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	defer func() {
		cfg2, _ := config.NewDefaultCGRConfig()
		config.SetCgrConfig(cfg2)
	}()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	cfg.DataDbCfg().Items[utils.MetaSupplierProfiles].Replicate = true
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1SetSupplierProfile: func(ctx *context.Context, args, reply any) error {
				*reply.(*string) = utils.OK
				return nil
			},
		},
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	fltrSupp1 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_SUPP_1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Supplier",
				Values:  []string{"SupplierProfile2"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: "~*req.PddInterval",
				Values:  []string{(1 * time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"15.0"},
			},
		},
	}
	if err := dm.SetFilter(fltrSupp1); err != nil {
		t.Error(err)
	}
	fltrSupp2 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_SUPP_2",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Supplier",
				Values:  []string{"SupplierProfilePrefix"},
			},
		},
	}
	if err := dm.SetFilter(fltrSupp2); err != nil {
		t.Error(err)
	}
	supp := &SupplierProfile{
		Tenant:            "cgrates.org",
		ID:                "SUP1",
		FilterIDs:         []string{"FLTR_SUPP_1"},
		Weight:            10,
		Sorting:           utils.MetaQOS,
		SortingParameters: []string{},
		Suppliers: []*Supplier{
			{
				ID:            "Sup",
				FilterIDs:     []string{},
				AccountIDs:    []string{"1001"},
				RatingPlanIDs: []string{"RT_PLAN1"},
				ResourceIDs:   []string{"RES1"},
				Weight:        10,
			},
		},
	}
	config.SetCgrConfig(cfg)
	if err := dm.SetSupplierProfile(supp, true); err != nil {
		t.Error(err)
	}
	supp1 := &SupplierProfile{
		Tenant:            "cgrates.org",
		ID:                "SUP1",
		FilterIDs:         []string{"FLTR_SUPP_2"},
		Weight:            10,
		Sorting:           utils.MetaQOS,
		SortingParameters: []string{},
		Suppliers: []*Supplier{
			{
				ID:            "Sup",
				FilterIDs:     []string{},
				AccountIDs:    []string{"1001"},
				RatingPlanIDs: []string{"RT_PLAN1"},
				ResourceIDs:   []string{"RES1"},
				Weight:        10,
			},
		},
	}
	if err := dm.SetSupplierProfile(supp1, true); err != nil {
		t.Error(err)
	}
	if err := dm.RemoveSupplierProfile("cgrates.org", supp1.ID, utils.NonTransactional, true); err != nil {
		t.Error(err)
	}
}

func TestDmMatchFilterIndexFromKey(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	defer func() {
		cfg2, _ := config.NewDefaultCGRConfig()
		config.SetCgrConfig(cfg2)
	}()
	Cache.Clear(nil)
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	cfg.DataDbCfg().Items[utils.MetaFilterIndexes].Remote = true
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1MatchFilterIndex: func(ctx *context.Context, args, reply any) error {
				return nil
			},
		},
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	fltr := &Filter{
		Tenant: "cgrates.org",
		ID:     "RES_FLT_1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Account,
				Values:  []string{"1002"},
			},
		},
	}
	if err := dm.SetFilter(fltr); err != nil {
		t.Error(err)
	}
	rp := &ResourceProfile{
		Tenant:       "cgrates.org",
		ID:           "RES1",
		FilterIDs:    []string{"RES_FLT_1"},
		UsageTTL:     time.Second,
		Limit:        1,
		Weight:       10,
		ThresholdIDs: []string{"TH1"},
	}
	if err := dm.SetResourceProfile(rp, true); err != nil {
		t.Error(err)
	}
	config.SetCgrConfig(cfg)
	if err := dm.MatchFilterIndexFromKey(utils.CacheResourceFilterIndexes, "cgrates.org:*string:Account:1002"); err != nil {
		t.Error(err)
	}
}

func TestCacheDataFromDB(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	chgS := ChargerProfiles{
		&ChargerProfile{
			Tenant:    "cgrates.org",
			ID:        "Charger1",
			FilterIDs: []string{"*string:~*req.Account:1015", "*gt:~*req.Usage:10"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
			},
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{"*none"},
			Weight:       20,
		},
		&ChargerProfile{
			Tenant:    "cgrates.com",
			ID:        "CHRG_1",
			FilterIDs: []string{"*string:Account:1001"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			AttributeIDs: []string{"ATTR_1"},
			Weight:       20,
		},
	}
	dest := &Destination{
		Id: "DEST", Prefixes: []string{"1004", "1002", "1002"},
	}
	dm.SetDestination(dest, "")
	dm.SetReverseDestination(dest, "")

	for _, chg := range chgS {
		if err := dm.SetChargerProfile(chg, true); err != nil {
			t.Error(err)
		}
	}
	if err := dm.CacheDataFromDB(utils.ChargerProfilePrefix, nil, false); err != nil {
		t.Error(err)
	}

	if err := dm.CacheDataFromDB(utils.DESTINATION_PREFIX, nil, false); err != nil {
		t.Error(err)
	}

	if err := dm.CacheDataFromDB(utils.REVERSE_DESTINATION_PREFIX, nil, false); err != nil {
		t.Error(err)
	}
}

func TestCacheDataFromDBFilterIndexes(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	fltr := &Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_ATTR_1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Attribute",
				Values:  []string{"AttributeProfile1"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: "~*req.UsageInterval",
				Values:  []string{(1 * time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: "~*req." + utils.Weight,
				Values:  []string{"9.0"},
			},
		},
	}
	dm.SetFilter(fltr)
	attr := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_1001_SIMPLEAUTH",
		FilterIDs: []string{"FLTR_ATTR_1"},
		Contexts:  []string{"simpleauth"},
		Attributes: []*Attribute{
			{
				FilterIDs: []string{},
				Path:      utils.MetaReq + utils.NestingSep + "Password",
				Type:      utils.META_CONSTANT,
				Value:     config.NewRSRParsersMustCompile("CGRateS.org", true, utils.INFIELD_SEP),
			},
		},
		Weight: 20.0,
	}
	dm.SetAttributeProfile(attr, true)
	if err := dm.CacheDataFromDB(utils.AttributeFilterIndexes, nil, false); err != nil {
		t.Error(err)
	}
	fltr2 := &Filter{
		Tenant: "cgrates.org",
		ID:     "RS_FLT",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Destination",
				Values:  []string{"1002", "1003"},
			},
		},
	}
	dm.SetFilter(fltr2)
	rsc := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"RS_FLT"},
		ThresholdIDs:      []string{utils.META_NONE},
		AllocationMessage: "Approved",
		Weight:            10,
		Limit:             10,
		UsageTTL:          time.Minute,
		Stored:            true,
	}
	dm.SetResourceProfile(rsc, true)
	if err := dm.CacheDataFromDB(utils.ResourceFilterIndexes, nil, false); err != nil {
		t.Error(err)
	}
	statFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_STATS_1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Account,
				Values:  []string{"1001"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: "~*req.UsageInterval",
				Values:  []string{(1 * time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage,
				Values:  []string{(1 * time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"9.0"},
			},
		},
	}
	dm.SetFilter(statFlt)
	sqP := &StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "DistinctMetricProfile",
		QueueLength: 10,
		FilterIDs:   []string{"FLTR_STATS_1"},
		TTL:         time.Duration(10) * time.Second,
		Metrics: []*MetricWithFilters{
			{
				MetricID: utils.MetaDDC,
			},
		},
		ThresholdIDs: []string{utils.META_NONE},
		Stored:       true,
		Weight:       20,
	}
	dm.SetStatQueueProfile(sqP, true)

	if err := dm.CacheDataFromDB(utils.StatFilterIndexes, nil, false); err != nil {
		t.Error(err)
	}
	thFltr := &Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_TH_2",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Threshold",
				Values:  []string{"TH_2"},
			},
			{
				Type:    utils.MetaPrefix,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Destination,
				Values:  []string{"100"},
			},
		},
	}
	dm.SetFilter(thFltr)
	thP := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "THD_AccDisableAndLog",
		FilterIDs: []string{"FLTR_TH_2"},
		MaxHits:   -1,
		MinSleep:  time.Duration(1 * time.Second),
		Weight:    30.0,
		ActionIDs: []string{"DISABLE_LOG"},
	}
	dm.SetThresholdProfile(thP, true)
	if err := dm.CacheDataFromDB(utils.ThresholdFilterIndexes, nil, false); err != nil {
		t.Error(err)
	}
	suppFltr := &Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_SUPP_1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Supplier",
				Values:  []string{"SupplierProfile2"},
			}},
	}
	dm.SetFilter(suppFltr)
	supp := &SupplierProfile{
		Tenant:    "cgrates.org",
		ID:        "SPP_1",
		FilterIDs: []string{"FLTR_SUPP_1"},

		Sorting:           utils.MetaLC,
		SortingParameters: []string{},
		Suppliers: []*Supplier{
			{
				ID:                 "supplier1",
				RatingPlanIDs:      []string{"RPL_2"},
				ResourceIDs:        []string{"ResGroup2", "ResGroup4"},
				StatIDs:            []string{"Stat3"},
				Weight:             10,
				Blocker:            false,
				SupplierParameters: utils.EmptyString,
			},
		},
		Weight: 20,
	}
	dm.SetSupplierProfile(supp, true)
	if err := dm.CacheDataFromDB(utils.SupplierFilterIndexes, nil, false); err != nil {
		t.Error(err)
	}
	ddpFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "DSP_FLT",
		Rules: []*FilterRule{
			{
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Account,
				Type:    utils.MetaString,
				Values:  []string{"2009"},
			},
		},
	}
	dm.SetFilter(ddpFlt)
	dpp := &DispatcherProfile{
		Tenant:     "cgrates.org",
		ID:         "DSP_Test1",
		FilterIDs:  []string{"DSP_FLT"},
		Strategy:   utils.MetaFirst,
		Subsystems: []string{utils.MetaAttributes, utils.MetaSessionS},
		Weight:     20,
	}
	if err := dm.SetDispatcherProfile(dpp, true); err != nil {
		t.Error(err)
	}
	if err := dm.CacheDataFromDB(utils.DispatcherFilterIndexes, nil, false); err != nil {
		t.Error(err)
	}
	chgFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "FLT_CPP",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + "Charger",
				Values:  []string{"Charger1"},
			},
		},
	}
	dm.SetFilter(chgFlt)
	cpp := &ChargerProfile{
		Tenant: "cgrates.org",
		ID:     "Default",
		FilterIDs: []string{"*string:~*req.Destination:+1442",
			"*prefix:~*opts.Accounts:1002;1004"},
		RunID:        utils.MetaDefault,
		AttributeIDs: []string{"*none"},
		Weight:       20,
	}
	if err := dm.SetChargerProfile(cpp, true); err != nil {
		t.Error(err)
	}
	if err := dm.CacheDataFromDB(utils.ChargerFilterIndexes, nil, false); err != nil {
		t.Error(err)
	}
}

func TestFilterIndexesRmtRpl(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	Cache.Clear(nil)
	cfg.DataDbCfg().Items[utils.MetaFilterIndexes].Remote = true
	cfg.DataDbCfg().Items[utils.MetaFilterIndexes].Replicate = true
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	defer func() {
		cfg2, _ := config.NewDefaultCGRConfig()
		config.SetCgrConfig(cfg2)
	}()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	expIndx := map[string]utils.StringMap{
		"*string:Account:1001": {
			"RL1": true,
		},
		"*string:Account:1002": {
			"RL1": true,
			"RL2": true,
		},
	}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- clMock(func(ctx *context.Context, m string, a, r any) error {
		if m == utils.ReplicatorSv1SetFilterIndexes {
			setFltrIndxArg, concat := a.(*utils.SetFilterIndexesArg)
			if !concat {
				return errors.New("Can't convert interfacea")
			}
			if err := dm.DataDB().SetFilterIndexesDrv(setFltrIndxArg.CacheID, setFltrIndxArg.ItemIDPrefix, setFltrIndxArg.Indexes, false, utils.EmptyString); err == nil {
				*r.(*string) = utils.OK
			}
			return nil
		} else if m == utils.ReplicatorSv1GetFilterIndexes {

			rpl := expIndx
			*r.(*map[string]utils.StringMap) = rpl
			return nil
		}
		return utils.ErrNotImplemented
	})
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	idx := map[string]utils.StringMap{
		"*string:Account:1001": {
			"DSP1": true,
			"DSP2": true,
		},
		"*suffix:*opts.Destination:+100": {
			"Dsp1": true,
			"Dsp2": true,
		},
	}
	config.SetCgrConfig(cfg)
	if err := dm.SetFilterIndexes(utils.CacheDispatcherProfiles, "cgrates.org", idx, false, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if err := dm.RemoveFilterIndexes(utils.CacheDispatcherProfiles, "cgrates.org"); err != nil {
		t.Error(err)
	}
	if rcvIdx, err := dm.GetFilterIndexes(utils.CacheResourceProfiles, "cgrates.org", utils.EmptyString, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expIndx, rcvIdx) {
		t.Errorf("Expected %+v,Received %+v", utils.ToJSON(expIndx), utils.ToJSON(idx))
	}
}

func TestStatQueueProfileIndx(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	fltrs := []*Filter{
		{
			Tenant: "cgrates.org",
			ID:     "SQ_FLT_1",
			Rules: []*FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Destination",
					Values:  []string{"1002", "1003", "1004"},
				},
			},
		},
		{
			Tenant: "cgrates.org",
			ID:     "SQ_FLT_2",
			Rules: []*FilterRule{

				{
					Type:    utils.MetaGreaterOrEqual,
					Element: "~*req.UsageInterval",
					Values:  []string{(1 * time.Second).String()},
				},
				{
					Type:    utils.MetaGreaterOrEqual,
					Element: "~*req." + utils.Weight,
					Values:  []string{"9.0"},
				},
			},
		}}
	for _, flt := range fltrs {
		if err := dm.SetFilter(flt); err != nil {
			t.Error(err)
		}
	}
	sqP := &StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "SQ_1",
		FilterIDs:   []string{"SQ_FLT_1"},
		QueueLength: 10,
		TTL:         time.Duration(0) * time.Second,
		Metrics: []*MetricWithFilters{
			{
				MetricID: "*asr",
			},
			{
				MetricID: utils.MetaACD,
			},
			{
				MetricID: "*acc",
			},
		},
		ThresholdIDs: []string{"Test"},
		Blocker:      false,
		Stored:       true,
		Weight:       float64(0),
		MinItems:     0,
	}
	if err := dm.SetStatQueueProfile(sqP, true); err != nil {
		t.Error(err)
	}
	sqP = &StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "SQ_1",
		FilterIDs:   []string{"SQ_FLT_2"},
		QueueLength: 10,
		TTL:         time.Duration(0) * time.Second,
		Metrics: []*MetricWithFilters{
			{
				MetricID: "*asr",
			},
			{
				MetricID: utils.MetaACD,
			},
			{
				MetricID: "*acc",
			},
		},
		ThresholdIDs: []string{"Test"},
		Blocker:      false,
		Stored:       true,
		Weight:       float64(0),
		MinItems:     0,
	}
	if err := dm.SetStatQueueProfile(sqP, true); err != nil {
		t.Error(err)
	}
	if err := dm.RemoveStatQueueProfile("cgrates.org", "SQ_1", utils.NonTransactional, true); err != nil {
		t.Error(err)
	}
}

func TestDmDispatcherHost(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	dppH := &DispatcherHost{
		Tenant: "cgrates.org",
		ID:     "ALL1",
		Conns: []*config.RemoteHost{
			{
				Address:   "127.0.0.1:2012",
				Transport: utils.MetaJSON,
				TLS:       true,
			},
			{
				Address:   "127.0.0.1:3012",
				Transport: utils.MetaJSON,
			},
		},
	}
	if err := dm.SetDispatcherHost(dppH); err != nil {
		t.Error(err)
	}
	if _, err := dm.GetDispatcherHost("cgrates.org", "ALL1", false, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if err := dm.RemoveDispatcherHost("cgrates.org", "ALL1", utils.NonTransactional); err != nil {
		t.Error(err)
	}
}

func TestDmRemoveThresholdProfile(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	thP := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "THD_ACNT_1001",
		FilterIDs: []string{"*prefix:Destination:46"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
		},
		MaxHits:   -1,
		MinSleep:  time.Duration(0),
		Blocker:   false,
		Weight:    10.0,
		ActionIDs: []string{"TOPUP_MONETARY_10"},
		Async:     false,
	}
	dm.SetThresholdProfile(thP, true)
	if err := dm.RemoveThresholdProfile("cgrates.org", "THD_ACNT_1001", utils.NonTransactional, true); err != nil {
		t.Error(err)
	}
}

func TestDMReplicateMultipleIds(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	tmpConn := connMgr
	defer func() {
		SetConnManager(tmpConn)
	}()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- clMock(func(ctx *context.Context, serviceMethod string, args, _ any) error {
		if serviceMethod == utils.ReplicatorSv1RemoveAccount {
			keyList, err := dm.DataDB().GetKeysForPrefix(args.(string))
			if err != nil {
				return utils.ErrNotFound
			}
			for _, key := range keyList {
				dm.RemoveAccount(key[len(utils.ACCOUNT_PREFIX):])
			}
			return nil
		}
		return utils.ErrNotImplemented
	})
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	accs := []*Account{
		{
			ID: "cgrates.org:1001",
			BalanceMap: map[string]Balances{
				utils.MONETARY: {
					&Balance{Value: 10},
				},
				utils.VOICE: {
					&Balance{Value: 10, Weight: 20, DestinationIDs: utils.NewStringMap("GER")},
					&Balance{Weight: 10, DestinationIDs: utils.NewStringMap("ENG")},
				},
			},
		},
		{
			ID: "cgrates.org:1002",
			BalanceMap: map[string]Balances{
				utils.MONETARY: {
					&Balance{
						Weight:         30,
						Value:          12,
						DestinationIDs: utils.NewStringMap("DEST"),
					},
				},
			},
		},
	}
	connIds := make([]string, len(accs))
	for i, acc := range accs {
		dm.SetAccount(acc)
		connIds[i] = acc.ID
		//UpdateReplicationFilters(utils.ACCOUNT_PREFIX, connIds[i], utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1))
	}

	if err := replicateMultipleIDs(connMgr, []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}, true, utils.ACCOUNT_PREFIX, connIds, utils.ReplicatorSv1RemoveAccount, "cgrates.org"); err != nil {
		t.Error(err)
	} else if ids, err := dm.DataDB().GetKeysForPrefix("cgrates"); len(ids) > 0 || err != nil {
		t.Error(err)
	}

}

func TestDmUpdateReverseDestination(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	dst := &Destination{Id: "DEST1", Prefixes: []string{"+456", "+457", "+458"}}
	dst2 := &Destination{Id: "DEST2", Prefixes: []string{"+466", "467", "468"}}
	if err := dm.SetReverseDestination(dst, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if err := dm.SetDestination(dst2, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	for i := range dst.Prefixes {
		if rcv, err := dm.GetReverseDestination(dst.Prefixes[i], true, utils.NonTransactional); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(rcv, []string{dst.Id}) {
			t.Errorf("Expected  %v,Received %v", utils.ToJSON(rcv), utils.ToJSON([]string{dst.Id}))
		}
	}
	if err := dm.UpdateReverseDestination(dst, dst2, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	for i := range dst2.Prefixes {
		if rcv, err := dm.GetReverseDestination(dst2.Prefixes[i], true, utils.NonTransactional); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(rcv, []string{dst2.Id}) {
			t.Errorf("Expected  %v,Received %v", utils.ToJSON(rcv), utils.ToJSON([]string{dst2.Id}))
		}
	}

}

func TestActionTriggerRplRmt(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	attrs := ActionTriggers{
		{
			Balance: &BalanceFilter{
				Type: utils.StringPointer(utils.MONETARY),
			},
			ThresholdType:  utils.TRIGGER_MAX_BALANCE,
			ThresholdValue: 2,
		},
		{
			UniqueID:      "TestTR1",
			ThresholdType: utils.TRIGGER_MAX_EVENT_COUNTER,
			Balance: &BalanceFilter{
				Type:   utils.StringPointer(utils.MONETARY),
				Weight: utils.Float64Pointer(10),
			},
		},
	}
	if err := dm.SetActionTriggers("TEST_ACTIONS", attrs, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if vals, err := dm.GetActionTriggers("TEST_ACTIONS", true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(attrs, vals) {
		t.Errorf("Expected %v,Receive %v", attrs, vals)
	}
}

func TestDMRemoveAttributeProfile(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	attrPrf := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_1",
		FilterIDs: []string{"*string:~*req.Account:1007"},
		Contexts:  []string{utils.MetaSessionS, utils.MetaCDRs},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 1, 14, 0, 0, 0, 0, time.UTC)},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + utils.Account,
				Value: config.NewRSRParsersMustCompile("1001", true, utils.INFIELD_SEP),
			},
			{
				Path:  utils.MetaReq + utils.NestingSep + utils.Subject,
				Value: config.NewRSRParsersMustCompile("1001", true, utils.INFIELD_SEP),
			},
		},
		Weight: 10.0,
	}
	dm.SetAttributeProfile(attrPrf, true)
	if err := dm.RemoveAttributeProfile("cgrates.org", "ATTR_1", utils.NonTransactional, true); err != nil {
		t.Error(err)
	}
}

func TestThresholdProfileSetWithIndex(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	fltr1 := &Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_TH_2",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Threshold",
				Values:  []string{"TH_2"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"15.0"},
			},
		},
	}
	dm.SetFilter(fltr1)
	thp := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "THD_AccDisableAndLog",
		FilterIDs: []string{"FLTR_TH_2"},
		MaxHits:   -1,
		MinSleep:  time.Duration(1 * time.Second),
		Weight:    20.0,
		Async:     true,
		ActionIDs: []string{"DISABLE_LOG"},
	}
	dm.SetThresholdProfile(thp, true)
	fltr2 := &Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_TH_3",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Threshold",
				Values:  []string{"THD"},
			},
		},
	}
	dm.SetFilter(fltr2)
	thp = &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "THD_AccDisableAndLog",
		FilterIDs: []string{"FLTR_TH_3"},
		MaxHits:   -1,
		MinSleep:  time.Duration(1 * time.Second),
		Weight:    20.0,
		Async:     true,
		ActionIDs: []string{"DISABLE_LOG"},
	}
	if err := dm.SetThresholdProfile(thp, true); err != nil {
		t.Error(err)
	}
}

func TestDmAllActionPlans(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	apS := []*ActionPlan{
		{
			Id:         "AP1",
			AccountIDs: utils.StringMap{"cgrates.org:1001": true},
			ActionTimings: []*ActionTiming{
				{
					Uuid: utils.GenUUID(),
					Timing: &RateInterval{
						Timing: &RITiming{
							Years:     utils.Years{2022},
							Months:    utils.Months{},
							MonthDays: utils.MonthDays{},
							WeekDays:  utils.WeekDays{},
							StartTime: utils.ASAP,
						},
					},
					Weight:    10,
					ActionsID: "ACT_1",
				},
			},
		},
		{
			Id:         "AP2",
			AccountIDs: utils.StringMap{"cgrates.org:1001": true},
			ActionTimings: []*ActionTiming{{
				Uuid: utils.GenUUID(),
				Timing: &RateInterval{
					Timing: &RITiming{
						Years:     utils.Years{2022},
						Months:    utils.Months{},
						MonthDays: utils.MonthDays{},
						WeekDays:  utils.WeekDays{},
						StartTime: utils.ASAP,
					},
				},
				Weight:    10,
				ActionsID: "ACT_2",
			},
			},
		}}
	expMap := make(map[string]*ActionPlan)
	for _, ap := range apS {
		dm.SetActionPlan(ap.Id, ap, true, utils.NonTransactional)
		expMap[ap.Id] = ap
	}
	if rpl, err := dm.GetAllActionPlans(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expMap, rpl) {
		t.Errorf("Expected %+v,Received %+v", utils.ToJSON(expMap), utils.ToJSON(rpl))
	}
}

func TestDMRemoveCHP(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	chp := &ChargerProfile{
		Tenant:    "cgrates.org",
		ID:        "Ch1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 00, 0, 0, time.UTC),
		},
		RunID:        "*rated",
		AttributeIDs: []string{"ATTR_1001_SIMPLEAUTH"},
		Weight:       20,
	}
	dm.SetChargerProfile(chp, true)
	if err := dm.RemoveChargerProfile("cgrates.org", "Ch1", utils.NonTransactional, true); err != nil {
		t.Error(err)
	}
}

func TestDMGetTiming(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	timing := &utils.TPTiming{
		ID:        "WEEKENDS",
		Years:     utils.Years{},
		Months:    utils.Months{},
		MonthDays: utils.MonthDays{},
		WeekDays:  utils.WeekDays{time.Saturday, time.Sunday},
		StartTime: "00:00:00",
	}
	dm.SetTiming(timing)
	if _, err := dm.GetTiming("WEEKENDS", true, utils.NonTransactional); err != nil {
		t.Error(err)
	}
}

func TestDmDispatcherProfile(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	fltr := &Filter{
		Tenant: "cgrates.org",
		ID:     "DSP_FLT",
		Rules: []*FilterRule{
			{
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Account,
				Type:    utils.MetaString,
				Values:  []string{"2009"},
			},
		},
	}
	fltr2 := &Filter{
		Tenant: "cgrates.org",
		ID:     "DSP_FLT2",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"15.0"},
			},
		},
	}
	dm.SetFilter(fltr)
	dm.SetFilter(fltr2)
	dpp := &DispatcherProfile{
		Tenant:     "cgrates.org",
		ID:         "DSP1",
		FilterIDs:  []string{"DSP_FLT"},
		Strategy:   utils.MetaFirst,
		Subsystems: []string{utils.MetaAttributes, utils.MetaSessionS},
		Weight:     20,
		Hosts: DispatcherHostProfiles{
			&DispatcherHostProfile{
				ID:        "ALL2",
				FilterIDs: []string{},
				Weight:    20,
				Params:    make(map[string]any),
			},
			&DispatcherHostProfile{
				ID:        "ALL",
				FilterIDs: []string{},
				Weight:    10,
				Params:    make(map[string]any),
			},
		},
	}
	dm.SetDispatcherProfile(dpp, true)
	dpp.FilterIDs = []string{"DSP_FLT2"}
	if err := dm.SetDispatcherProfile(dpp, true); err != nil {
		t.Error(err)
	}

}

func TestDmGetSQRemote(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaStatQueues].Remote = true
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	Cache.Clear(nil)
	defer func() {
		cfg2, _ := config.NewDefaultCGRConfig()
		config.SetCgrConfig(cfg2)
	}()
	clientConnn := make(chan birpc.ClientConnector, 1)
	clientConnn <- clMock(func(ctx *context.Context, serviceMethod string, _, reply any) error {
		if serviceMethod == utils.ReplicatorSv1GetStatQueue {

			*reply.(*StoredStatQueue) = StoredStatQueue{
				Tenant:    "cgrates.org",
				ID:        "SQ1",
				SQItems:   []SQItem{},
				SQMetrics: map[string][]byte{},
			}
			return nil
		}

		return utils.ErrNotFound
	})
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConnn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	config.SetCgrConfig(cfg)
	if _, err := dm.GetStatQueue("cgrates.org", "SQ1", false, true, utils.NonTransactional); err == nil {
		t.Error(err)
	}
	// unfinished
}

func TestRemoveThresholdRpl(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaThresholds].Replicate = true
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	Cache.Clear(nil)
	defer func() {
		cfg2, _ := config.NewDefaultCGRConfig()
		config.SetCgrConfig(cfg2)
	}()
	clientConnn := make(chan birpc.ClientConnector, 1)
	clientConnn <- clMock(func(ctx *context.Context, serviceMethod string, _, reply any) error {
		if serviceMethod == utils.ReplicatorSv1RemoveThreshold {

			return nil
		}

		return utils.ErrNotFound
	})
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConnn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	config.SetCgrConfig(cfg)
	if err := dm.RemoveThreshold("cgrates.org", "TH1", utils.NonTransactional); err != nil {
		t.Error(err)
	}

}

func TestRemoveDispatcherPrfRpl(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaDispatcherProfiles].Replicate = true
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	Cache.Clear(nil)
	defer func() {
		cfg2, _ := config.NewDefaultCGRConfig()
		config.SetCgrConfig(cfg2)
	}()
	clientConnn := make(chan birpc.ClientConnector, 1)
	clientConnn <- clMock(func(ctx *context.Context, serviceMethod string, _, reply any) error {
		if serviceMethod == utils.ReplicatorSv1RemoveDispatcherProfile {

			return nil
		}

		return utils.ErrNotFound
	})
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConnn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	dpp := &DispatcherProfile{
		Tenant:     "cgrates.org",
		ID:         "DSP_Test2",
		Subsystems: []string{utils.MetaAttributes},
		Weight:     20,
	}
	dm.SetDispatcherProfile(dpp, true)
	config.SetCgrConfig(cfg)
	if err := dm.RemoveDispatcherProfile("cgrates.org", "DSP_Test2", utils.NonTransactional, true); err != nil {
		t.Error(err)
	}
}

func TestDMReconnect(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	newCfg := *cfg.DataDbCfg()
	newCfg.DataDbType = "*internal"
	if err := dm.Reconnect(cfg.GeneralCfg().DBDataEncoding, &newCfg); err != nil {
		t.Error(err)
	}
}

func TestDMRemAccountActionPlans(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	acc := &Account{
		ID: "cgrates.org:1001",
		BalanceMap: map[string]Balances{
			utils.MONETARY: {
				&Balance{
					Value: 10,
				},
			}}}
	dm.SetAccount(acc)
	apIDs := []string{"PACKAGE_10_SHARED_A_5", "USE_SHARED_A"}
	if err := dm.SetAccountActionPlans(acc.ID, apIDs, false); err != nil {
		t.Error(err)
	}
	if err := dm.RemAccountActionPlans(acc.ID, apIDs); err != nil {
		t.Error(err)
	}

}

func TestDMGetDispacherHost(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaDispatcherHosts].Remote = true
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	Cache.Clear(nil)
	defer func() {
		cfg2, _ := config.NewDefaultCGRConfig()
		config.SetCgrConfig(cfg2)
	}()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- clMock(func(ctx *context.Context, serviceMethod string, _, reply any) error {
		if serviceMethod == utils.ReplicatorSv1GetDispatcherHost {

			rpl := &DispatcherHost{
				Tenant: "	cgrates.org",
				ID:     "DP_1",
			}
			*reply.(**DispatcherHost) = rpl
			return nil
		}
		return utils.ErrNotImplemented
	})
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	config.SetCgrConfig(cfg)
	if _, err := dm.GetDispatcherHost("cgrates.org", "DP_1", false, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}
}

func TestDmRemoveStatQueue(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().Items[utils.MetaStatQueues].Replicate = true
	defer func() {
		cfg2, _ := config.NewDefaultCGRConfig()
		config.SetCgrConfig(cfg2)
	}()
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- clMock(func(ctx *context.Context, serviceMethod string, _, _ any) error {
		if serviceMethod == utils.ReplicatorSv1RemoveStatQueue {
			return nil
		}
		return utils.ErrNotImplemented
	})
	connMgr := NewConnManager(cfg,
		map[string]chan birpc.ClientConnector{
			utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
		})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	config.SetCgrConfig(cfg)
	if err := dm.RemoveStatQueue("cgrates.org", "SQ1", utils.NonTransactional); err != nil {
		t.Error(err)
	}

}

func TestRemoveResource(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	defer func() {
		cfg2, _ := config.NewDefaultCGRConfig()
		config.SetCgrConfig(cfg2)
	}()
	cfg.DataDbCfg().Items[utils.MetaResources].Replicate = true
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)

	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- clMock(func(ctx *context.Context, serviceMethod string, _, _ any) error {
		if utils.ReplicatorSv1RemoveResource == serviceMethod {
			return nil
		}
		return utils.ErrNotImplemented
	})
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)

	if err := dm.RemoveResource("cgrates.org", "R1", utils.NonTransactional); err != nil {
		t.Error(err)
	}

}

func TestSetReverseDestinastionRpl(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	defer func() {
		cfg2, _ := config.NewDefaultCGRConfig()
		config.SetCgrConfig(cfg2)
	}()
	cfg.DataDbCfg().Items[utils.MetaReverseDestinations].Replicate = true
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)

	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- clMock(func(ctx *context.Context, serviceMethod string, _, _ any) error {
		if utils.ReplicatorSv1SetReverseDestination == serviceMethod {
			return nil
		}
		return utils.ErrNotImplemented
	})
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)

	if err := dm.RemoveResource("cgrates.org", "R1", utils.NonTransactional); err != nil {
		t.Error(err)
	}

}

func TestDMGetThresholdRmt(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	defer func() {
		cfg2, _ := config.NewDefaultCGRConfig()
		config.SetCgrConfig(cfg2)
	}()
	Cache.Clear(nil)
	cfg.DataDbCfg().Items[utils.MetaThresholds].Remote = true
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- clMock(func(ctx *context.Context, serviceMethod string, _, reply any) error {
		if serviceMethod == utils.ReplicatorSv1GetThreshold {
			rpl := &Threshold{
				Tenant: "cgrates.org", ID: "THD_Stat", Hits: 1,
			}
			*reply.(**Threshold) = rpl
			return nil
		}
		return utils.ErrNotImplemented
	})
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	config.SetCgrConfig(cfg)
	if _, err := dm.GetThreshold("cgrates.org", "THD_Stat", false, true, ""); err != nil {
		t.Error(err)
	}
}

func TestDmGetRatingPlanRmt(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaRatingPlans].Remote = true
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	defer func() {
		cfg2, _ := config.NewDefaultCGRConfig()
		config.SetCgrConfig(cfg2)
	}()
	Cache.Clear(nil)
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- clMock(func(ctx *context.Context, serviceMethod string, _, reply any) error {
		if serviceMethod == utils.ReplicatorSv1GetRatingPlan {
			rpl := &RatingPlan{
				Id: "RP1",
				Ratings: map[string]*RIRate{
					"b457f86d": {
						ConnectFee: 0,
						Rates: []*Rate{
							{
								GroupIntervalStart: 0,
								Value:              0.03,
								RateIncrement:      time.Second,
								RateUnit:           time.Second,
							},
						},
						RoundingMethod:   utils.ROUNDING_MIDDLE,
						RoundingDecimals: 4,
					},
				},
			}
			*reply.(**RatingPlan) = rpl
			return nil
		}
		return utils.ErrNotImplemented
	})
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	config.SetCgrConfig(cfg)
	if _, err := dm.GetRatingPlan("RP1", true, ""); err != nil {
		t.Error(err)
	}
}

func TestDMGetTimingRmt(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	cfg.DataDbCfg().Items[utils.MetaTimings].Remote = true
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	Cache.Clear(nil)
	defer func() {
		cfg2, _ := config.NewDefaultCGRConfig()
		config.SetCgrConfig(cfg2)
	}()
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- clMock(func(ctx *context.Context, serviceMethod string, _, reply any) error {
		if serviceMethod == utils.ReplicatorSv1GetTiming {
			rpl := &utils.TPTiming{
				ID:        "WEEKENDS",
				Years:     utils.Years{},
				Months:    utils.Months{},
				MonthDays: utils.MonthDays{},
				WeekDays:  utils.WeekDays{time.Saturday, time.Sunday},
				StartTime: "00:00:00",
			}
			*reply.(**utils.TPTiming) = rpl

			return nil
		}
		return utils.ErrNotImplemented
	})
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	config.SetCgrConfig(cfg)

	if _, err := dm.GetTiming("ALWAYS", true, utils.NonTransactional); err != nil {
		t.Error(err)
	}
}

func TestSetResourceProfileRPl(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	cfg.DataDbCfg().Items[utils.MetaResourceProfile].Remote = true
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	Cache.Clear(nil)
	defer func() {
		cfg2, _ := config.NewDefaultCGRConfig()
		config.SetCgrConfig(cfg2)
	}()
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- clMock(func(ctx *context.Context, serviceMethod string, _, reply any) error {
		if serviceMethod == utils.ReplicatorSv1GetResourceProfile {
			rpl := &ResourceProfile{
				Tenant:    "cgrates.org",
				ID:        "RES_TEST",
				FilterIDs: []string{"FLTR_1"},
				ActivationInterval: &utils.ActivationInterval{
					ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
					ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				},
				UsageTTL:          time.Duration(-1),
				Limit:             2,
				AllocationMessage: "",
				Weight:            20,
				ThresholdIDs:      []string{utils.META_NONE},
			}
			*reply.(**ResourceProfile) = rpl

			return nil
		}
		return utils.ErrNotImplemented
	})
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	config.SetCgrConfig(cfg)

	if _, err := dm.GetResourceProfile("cgrates.org", "RES_TEST", false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	}
}

func TestDMGetActionTriggersRmt(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	cfg.DataDbCfg().Items[utils.MetaActionTriggers].Remote = true
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	Cache.Clear(nil)
	defer func() {
		cfg2, _ := config.NewDefaultCGRConfig()
		config.SetCgrConfig(cfg2)
	}()
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- clMock(func(ctx *context.Context, serviceMethod string, _, reply any) error {
		if serviceMethod == utils.ReplicatorSv1GetActionTriggers {
			rpl := ActionTriggers{
				&ActionTrigger{
					Recurrent: true,
					Balance: &BalanceFilter{
						Type:           utils.StringPointer(utils.VOICE),
						ExpirationDate: utils.TimePointer(time.Date(2023, 3, 12, 0, 0, 0, 0, time.UTC)),
						Weight:         utils.Float64Pointer(20),
						DestinationIDs: &utils.StringMap{
							"DEST1": true,
						},
					},
					ThresholdValue: 2,
					ThresholdType:  utils.TRIGGER_BALANCE_EXPIRED,
					ActionsID:      "ACT_1",
					Executed:       false,
				},
			}
			*reply.(*ActionTriggers) = rpl
			return nil
		}
		return utils.ErrNotImplemented
	})
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	config.SetCgrConfig(cfg)

	if _, err := dm.GetActionTriggers("ACT_1", true, ""); err == nil {
		t.Error(err)
	}
	//unfinished
}

func TestDMRemSQPRepl(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	defer func() {
		cfg2, _ := config.NewDefaultCGRConfig()
		config.SetCgrConfig(cfg2)
	}()
	cfg.DataDbCfg().Items[utils.MetaStatQueueProfiles].Replicate = true
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- clMock(func(ctx *context.Context, serviceMethod string, _, _ any) error {
		if serviceMethod == utils.ReplicatorSv1RemoveStatQueueProfile {

			return nil
		}
		return utils.ErrNotImplemented
	})
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	dm.SetStatQueueProfile(&StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "SQ1",
		QueueLength: 10,
		TTL:         time.Duration(0) * time.Second,
		Metrics: []*MetricWithFilters{
			{
				MetricID: utils.MetaACD,
			},
		},
		Blocker:  false,
		Stored:   true,
		Weight:   float64(0),
		MinItems: 0}, true)
	config.SetCgrConfig(cfg)

	if err := dm.RemoveStatQueueProfile("cgrates.org", "SQ1", "", true); err != nil {
		t.Error(err)
	}
}

func TestDmRemoveFilter(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaFilters].Replicate = true
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	defer func() {
		cfg2, _ := config.NewDefaultCGRConfig()
		config.SetCgrConfig(cfg2)
	}()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- clMock(func(ctx *context.Context, serviceMethod string, _, _ any) error {
		if serviceMethod == utils.ReplicatorSv1RemoveFilter {

			return nil
		}
		return utils.ErrNotImplemented
	})
	dm := NewDataManager(db, cfg.CacheCfg(), NewConnManager(cfg, map[string]chan birpc.ClientConnector{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn}))
	dm.SetFilter(&Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_CP",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Charger",
				Values:  []string{"ChargerProfile2"},
			},
		},
	})
	dm.SetChargerProfile(&ChargerProfile{
		Tenant:    "cgrates.org",
		ID:        "Ch1",
		FilterIDs: []string{"FLTR_CP"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 00, 0, 0, time.UTC),
		},
		RunID:        "*rated",
		AttributeIDs: []string{"ATTR_1001_SIMPLEAUTH"},
		Weight:       20,
	}, true)
	config.SetCgrConfig(cfg)
	if err := dm.RemoveFilter("cgrates.org", "FLTR_CP", utils.NonTransactional); err == nil {
		t.Error(err)
	}
	dm.RemoveChargerProfile("cgrates.org", "Ch1", "", true)
	if err := dm.RemoveFilter("cgrates.org", "FLTR_CP", utils.NonTransactional); err != nil {
		t.Error(err)
	}
}

func TestDMGetSupplierProfile(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	Cache.Clear(nil)
	cfg.DataDbCfg().Items[utils.MetaSupplierProfiles].Remote = true
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	defer func() {
		cfg2, _ := config.NewDefaultCGRConfig()
		config.SetCgrConfig(cfg2)
	}()
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- clMock(func(ctx *context.Context, serviceMethod string, _, reply any) error {
		if serviceMethod == utils.ReplicatorSv1GetSupplierProfile {
			rpl := &SupplierProfile{}
			*reply.(**SupplierProfile) = rpl
			return nil
		}
		return utils.ErrNotFound
	})
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	}))
	config.SetCgrConfig(cfg)
	if _, err := dm.GetSupplierProfile("cgrates.org", "SPL1", false, false, ""); err != nil {
		t.Error(err)
	}
}

func TestConnManagerCallWithConnIDs(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	connId := utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)
	cfg.RPCConns()[connId] = &config.RPCConn{
		Conns: []*config.RemoteHost{
			{
				ID:        connId,
				Address:   "127.0.0.1:2012",
				Transport: utils.MetaJSON,
				TLS:       true,
			},
		},
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{})
	if err := connMgr.CallWithConnIDs([]string{connId}, utils.StringSet{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): {}}, utils.ReplicatorSv1GetAccount, nil, nil); err == nil {

		t.Error(err)
	}
}

func TestDMRemResourceProfile(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	defer func() {
		cfg2, _ := config.NewDefaultCGRConfig()
		config.SetCgrConfig(cfg2)
	}()
	cfg.DataDbCfg().Items[utils.MetaResourceProfile].Replicate = true
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- clMock(func(ctx *context.Context, serviceMethod string, _, _ any) error {
		if serviceMethod == utils.ReplicatorSv1RemoveResourceProfile {
			return nil
		}
		return utils.ErrNotImplemented
	})
	dm := NewDataManager(db, cfg.CacheCfg(), NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	}))
	rsP := &ResourceProfile{
		Tenant:    "cgrates.org",
		ID:        "RES_TEST",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		UsageTTL:          time.Duration(-1),
		Limit:             2,
		AllocationMessage: "Account1Channels",
		Weight:            20,
		ThresholdIDs:      []string{utils.META_NONE},
	}
	config.SetCgrConfig(cfg)
	dm.SetResourceProfile(rsP, true)
	if err := dm.RemoveResourceProfile("cgrates.org", "RES_TEST", utils.NonTransactional, true); err != nil {
		t.Error(err)
	}
}

func TestDMSetSQPrf(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	defer func() {
		cfg2, _ := config.NewDefaultCGRConfig()
		config.SetCgrConfig(cfg2)
	}()
	Cache.Clear(nil)
	cfg.DataDbCfg().Items[utils.MetaStatQueueProfiles].Replicate = true
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- clMock(func(ctx *context.Context, serviceMethod string, _, _ any) error {
		if serviceMethod == utils.ReplicatorSv1SetStatQueueProfile {
			return nil
		}
		return utils.ErrNotImplemented
	})
	dm := NewDataManager(db, cfg.CacheCfg(), NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	}))
	sqp := &StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "DistinctMetricProfile",
		QueueLength: 10,
		TTL:         time.Duration(10) * time.Second,
		Metrics: []*MetricWithFilters{
			{
				MetricID: utils.MetaDDC,
			},
		},
		ThresholdIDs: []string{utils.META_NONE},
		Stored:       true,
		Weight:       20,
	}
	config.SetCgrConfig(cfg)
	if err := dm.SetStatQueueProfile(sqp, true); err != nil {
		t.Error(err)
	}

}

func TestRemoveAttributeProfile(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	defer func() {
		cfg2, _ := config.NewDefaultCGRConfig()
		config.SetCgrConfig(cfg2)
	}()
	cfg.DataDbCfg().Items[utils.MetaAttributeProfiles].Replicate = true
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- clMock(func(_ *context.Context, serviceMethod string, _, _ any) error {
		if serviceMethod == utils.ReplicatorSv1RemoveAttributeProfile {
			return nil
		}
		return utils.ErrNotImplemented
	})

	dm := NewDataManager(NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items), cfg.CacheCfg(), NewConnManager(cfg, map[string]chan context.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes): clientConn,
	}))

	tests := []struct {
		name          string
		tenant        string
		id            string
		transactionID string
		withIndex     bool
		wantError     bool
	}{
		{
			name:      "Remove Nonexistent Attribute Profile",
			tenant:    "cgrates.org",
			id:        "non_exst",
			withIndex: true,
			wantError: true,
		},
		{
			name:      "Remove AttributeProfile non-existing filter",
			tenant:    "cgrates.org",
			id:        "ATTR_1002_SIMPLEAUTH",
			withIndex: true,
			wantError: true,
		},
		{
			name:      "Remove an attribute profile without a specific index",
			tenant:    "cgrates.org",
			id:        "ATTR_NO_FLTR",
			withIndex: false,
			wantError: false,
		},
	}
	attributes := []struct {
		index     bool
		attribute *AttributeProfile
	}{
		{
			index: true,
			attribute: &AttributeProfile{
				Tenant:    "cgrates.org",
				ID:        "ATTR_1002_SIMPLEAUTH",
				FilterIDs: []string{"FLT_NIL"},
				Contexts:  []string{"simpleauth"},
				Weight:    20.0,
			},
		}, {
			attribute: &AttributeProfile{
				Tenant:   "cgrates.org",
				ID:       "ATTR_NO_FLTR",
				Contexts: []string{utils.MetaSessionS, utils.META_ANY},
				Attributes: []*Attribute{
					{
						Path:  utils.MetaReq + utils.NestingSep + utils.CGRID,
						Value: config.NewRSRParsersMustCompile("test_generated_id", true, utils.INFIELD_SEP),
					},
				},
				Weight: 20.0,
			},
			index: false,
		},
	}
	for _, attr := range attributes {
		dm.SetAttributeProfile(attr.attribute, attr.index)
	}

	config.SetCgrConfig(cfg)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dm.RemoveAttributeProfile(tt.tenant, tt.id, tt.transactionID, tt.withIndex)
			if (err != nil) != tt.wantError {
				t.Errorf("RemoveAttributeProfile() error = %v, wantError %v", err, tt.wantError)
				return
			}
		})
	}
}

func TestDataManagerRemoveChargerProfile(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	defer func() {
		cfg2, _ := config.NewDefaultCGRConfig()
		config.SetCgrConfig(cfg2)
	}()
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().Items[utils.MetaChargerProfiles].Replicate = true
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- clMock(func(_ *context.Context, serviceMethod string, _, _ any) error {
		if serviceMethod == utils.ReplicatorSv1RemoveChargerProfile {
			return nil
		}
		return utils.ErrNotFound
	})
	dm := NewDataManager(db, cfg.CacheCfg(), NewConnManager(cfg, map[string]chan context.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	}))
	testCases := []struct {
		name           string
		tenant         string
		id             string
		transactionID  string
		withIndex      bool
		expectedErr    bool
		chargerProfile *ChargerProfile
	}{
		{
			name:          "RemoveChargerProfile - Not Found",
			tenant:        "cgrates.org",
			id:            "cpp1",
			transactionID: "",
			withIndex:     false,
			expectedErr:   true,
		},
		{
			name:          "RemoveChargerProfile - Broken Filter",
			tenant:        "cgrates.org",
			id:            "cpp2",
			transactionID: "",
			withIndex:     true,
			chargerProfile: &ChargerProfile{
				Tenant:    "cgrates.org",
				ID:        "cpp2",
				FilterIDs: []string{"FLTR_CP_2"},
				ActivationInterval: &utils.ActivationInterval{
					ActivationTime: time.Date(2023, 7, 14, 14, 25, 0, 0, time.UTC),
				},
				RunID:        "*rated",
				AttributeIDs: []string{"ATTR_1"},
				Weight:       20,
			},
			expectedErr: true,
		},
		{
			name:          "RemoveChargerProfile - Successful Removal",
			tenant:        "cgrates.org",
			id:            "cpp1",
			transactionID: "",
			withIndex:     false,
			expectedErr:   false,
			chargerProfile: &ChargerProfile{
				Tenant:    "cgrates.org",
				ID:        "cpp1",
				FilterIDs: []string{"*string:Account:1001"},
				RunID:     "*default",
				Weight:    20,
			},
		},
	}
	config.SetCgrConfig(cfg)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.chargerProfile != nil {
				dm.SetChargerProfile(tc.chargerProfile, true)
			}
			err := dm.RemoveChargerProfile(tc.tenant, tc.id, tc.transactionID, tc.withIndex)
			if (err != nil) != tc.expectedErr {
				t.Errorf("Expected error: %v, received error: %v", tc.expectedErr, err)
			}
		})
	}
}

func TestDMGetFilter(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	defer func() {
		cfg2, _ := config.NewDefaultCGRConfig()
		config.SetCgrConfig(cfg2)
	}()
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().Items[utils.MetaFilters].Remote = true
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- clMock(func(_ *context.Context, serviceMethod string, _, reply any) error {
		if serviceMethod == utils.ReplicatorSv1GetFilter {
			flt := &Filter{
				Tenant: "cgrates.org",
				ID:     "FLTR_2",
				Rules: []*FilterRule{
					{
						Type:    "*prefix",
						Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Destination,
						Values:  []string{"10", "20"},
					},
				},
			}
			*reply.(**Filter) = flt
			return nil
		}
		return utils.ErrNotFound
	})
	dm := NewDataManager(db, cfg.CacheCfg(), NewConnManager(cfg, map[string]chan context.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	}))
	Cache.Clear(nil)
	testCases := []struct {
		name          string
		dm            *DataManager
		tenant        string
		id            string
		cacheRead     bool
		cachWrite     bool
		transactionID string
		expectedErr   bool
		expected      *Filter
	}{
		{
			name:          "GetFilter - Stored In Cache",
			dm:            dm,
			tenant:        "cgrates.org",
			cacheRead:     true,
			id:            "FLT_1",
			cachWrite:     false,
			transactionID: utils.NonTransactional,
			expectedErr:   false,
			expected: &Filter{
				Tenant: "cgrates.org",
				ID:     "FLT_1",
				Rules: []*FilterRule{
					{
						Type:    "*string",
						Element: "~*req.Destination",
						Values:  []string{"1002"},
					},
				},
			},
		},
		{
			name:          "GetFilter - Nil DataManager",
			dm:            nil,
			cacheRead:     false,
			cachWrite:     false,
			transactionID: utils.NonTransactional,

			tenant:      "cgrates.org",
			expectedErr: true,
		},
		{
			name:        "GetFilter - Inline Filter Error",
			dm:          dm,
			tenant:      "cgrates.org",
			id:          "*stringAccount:Error",
			expectedErr: true,
		},
		{
			name:          "GetFilter - Remote",
			dm:            dm,
			tenant:        "cgrates.org",
			id:            "FLT_2",
			expectedErr:   false,
			cacheRead:     false,
			cachWrite:     true,
			transactionID: utils.NonTransactional,
		},
	}
	Cache.Set(utils.CacheFilters, utils.ConcatenatedKey("cgrates.org", "FLT_1"), &Filter{
		Tenant: "cgrates.org",
		ID:     "FLT_1",
		Rules: []*FilterRule{
			{
				Type:    "*string",
				Element: "~*req.Destination",
				Values:  []string{"1002"},
			},
		},
	}, []string{}, true, "")
	config.SetCgrConfig(cfg)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			val, err := GetFilter(tc.dm, tc.tenant, tc.id, tc.cacheRead, tc.cachWrite, tc.transactionID)
			if (err != nil) != tc.expectedErr {
				t.Errorf("Expected error: %v, received error: %v", tc.expectedErr, err)
			} else if tc.expected != nil {
				if !reflect.DeepEqual(val, tc.expected) {
					t.Errorf("Expected: %v, received: %v", tc.expected, val)
				}
			}
		})
	}
}

func TestDmRemoveThresholdProfileErrs(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	testCases := []struct {
		name          string
		tenant        string
		id            string
		transactionID string
		withIndex     bool
	}{
		{
			name:          "RemoveThresholdProfile - Profile not Found",
			tenant:        "cgrates.org",
			id:            "THP_1",
			transactionID: "",
		},
		{
			name:          "RemoveThresholdProfile - Broken Filter",
			tenant:        "cgrates.org",
			id:            "THP_2",
			transactionID: "",
			withIndex:     true,
		},
	}
	dm.SetThresholdProfile(&ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "THP_2",
		FilterIDs: []string{"FLT2"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2024, 7, 14, 14, 35, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2024, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		MaxHits:  -1,
		MinSleep: time.Duration(5 * time.Minute),
		Blocker:  false,
		Weight:   20.0,
	}, true)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := dm.RemoveThresholdProfile(tc.tenant, tc.id, tc.transactionID, tc.withIndex)
			if err == nil {
				t.Error(err)
			}
		})
	}
}

func TestDmGetAccountActionPlans(t *testing.T) {

	cfg, _ := config.NewDefaultCGRConfig()

	defer func() {
		cfg2, _ := config.NewDefaultCGRConfig()
		config.SetCgrConfig(cfg2)
	}()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().Items[utils.MetaAccountActionPlans].Remote = true
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- clMock(func(_ *context.Context, serviceMethod string, _, _ any) error {
		if serviceMethod == utils.ReplicatorSv1GetAccountActionPlans {

			return nil
		}
		return utils.ErrNotImplemented
	})
	dm := NewDataManager(db, cfg.CacheCfg(), NewConnManager(
		cfg, map[string]chan birpc.ClientConnector{
			utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
		},
	))
	Cache.Clear(nil)
	testCases := []struct {
		name      string
		acntId    string
		cacheRead bool
		error     bool
	}{
		{
			name:      "Not Found in Cache",
			acntId:    "Actions1",
			cacheRead: true,
			error:     true,
		},
		{
			name:      "Get AccountActions Remote",
			acntId:    "Actions2",
			cacheRead: false,
			error:     false,
		},
	}
	Cache.Set(utils.CacheAccountActionPlans, "Actions1", nil, []string{}, true, "")
	config.SetCgrConfig(cfg)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := dm.GetAccountActionPlans(tc.acntId, tc.cacheRead, false, utils.NonTransactional); (err != nil) != tc.error {
				t.Error(err)
			}
		})
	}
}

func TestDmSetFilter(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	defer func() {
		cfg2, _ := config.NewDefaultCGRConfig()
		config.SetCgrConfig(cfg2)
	}()
	Cache.Clear(nil)
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	cfg.DataDbCfg().Items[utils.MetaFilters].Replicate = true
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- clMock(func(_ *context.Context, serviceMethod string, _, _ any) error {
		if serviceMethod == utils.ReplicatorSv1SetFilter {
			return nil
		}
		return utils.ErrNotImplemented
	})
	dm := NewDataManager(db, cfg.CacheCfg(), NewConnManager(cfg, map[string]chan context.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	}))
	fltr := &Filter{
		Tenant: "cgrates.org",
		ID:     "FLT_1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Account",
				Values:  []string{"1001"},
			},
		},
	}

	testCases := []struct {
		name   string
		fltr   *Filter
		expErr bool
	}{
		{
			name:   "GetFilter Remote",
			fltr:   fltr,
			expErr: false,
		},
	}
	config.SetCgrConfig(cfg)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if err := dm.SetFilter(tc.fltr); (err != nil) != tc.expErr {
				t.Error(err)
			}
		})
	}

}

func TestDmRemAccountActionPlans(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	cfg.DataDbCfg().Items[utils.MetaAccountActionPlans].Replicate = true
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- clMock(func(_ *context.Context, serviceMethod string, _, _ any) error {
		if serviceMethod == utils.ReplicatorSv1RemAccountActionPlans {
			return nil
		}
		return utils.ErrNotImplemented
	})
	dm := NewDataManager(db, cfg.CacheCfg(), NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	}))
	defer func() {
		cfg2, _ := config.NewDefaultCGRConfig()
		config.SetCgrConfig(cfg2)
	}()
	testCases := []struct {
		name   string
		acnId  string
		apIDs  []string
		expErr bool
	}{
		{
			name:   "DataManager.RemAccountActionPlans NotFound",
			acnId:  "cgrates.org:1001",
			apIDs:  []string{"Actions1"},
			expErr: true,
		},
		{
			name:   "DataManager.RemAccountActionPlans Replicate",
			acnId:  "cgrates.org:1001",
			apIDs:  []string{"ACTIONS1"},
			expErr: false,
		},
	}
	config.SetCgrConfig(cfg)
	dm.SetAccountActionPlans(testCases[1].acnId, testCases[1].apIDs, true)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if err := dm.RemAccountActionPlans(tc.acnId, tc.apIDs); (err != nil) != tc.expErr {
				t.Error("Expected error ,received nil")
			}
		})
	}
}

func TestDMSetActionPlanRpl(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- clMock(func(_ *context.Context, serviceMethod string, _, _ any) error {
		if serviceMethod == utils.ReplicatorSv1SetActionPlan {
			return nil
		}
		return utils.ErrNotImplemented
	})
	dm := NewDataManager(db, cfg.CacheCfg(), NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	}))
	defer func() {
		cfg2, _ := config.NewDefaultCGRConfig()
		config.SetCgrConfig(cfg2)
	}()
	cfg.DataDbCfg().Items[utils.MetaAccountActionPlans].Replicate = true
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	testCases := []struct {
		name      string
		key       string
		ats       *ActionPlan
		overwrite bool
		expErr    bool
	}{
		{
			name: "Remove Action",
			key:  "Actions1",
			ats: &ActionPlan{
				Id:            "Actions1",
				AccountIDs:    utils.StringMap{"cgrates.org: 1001": true},
				ActionTimings: []*ActionTiming{},
			},
			expErr:    false,
			overwrite: true,
		},
		{
			name: "Sett ActionPlans Replication",
			key:  "MORE_MINUTES",
			ats: &ActionPlan{
				Id: "MORE_MINUTES",
				ActionTimings: []*ActionTiming{
					{
						Timing: &RateInterval{
							Timing: &RITiming{
								Years:     utils.Years{2022},
								Months:    utils.Months{},
								MonthDays: utils.MonthDays{},
								WeekDays:  utils.WeekDays{},
								StartTime: utils.ASAP,
							},
						},
						Weight:    10,
						ActionsID: "MINI",
					},
				},
			},
			overwrite: false,
			expErr:    false,
		},
	}
	config.SetCgrConfig(cfg)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if err := dm.SetActionPlan(tc.key, tc.ats, tc.overwrite, utils.NonTransactional); (err != nil) != tc.expErr {
				t.Errorf("Expected error:%v ,received %v", tc.expErr, err)
			}
		})
	}
}

func TestDataManager(t *testing.T) {
	var dm *DataManager

	rcv := dm.DataDB()

	if rcv != nil {
		t.Error(rcv)
	}
}
