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
package engine

import (
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
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
	clientConn := make(chan rpcclient.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ReplicatorSv1SetSupplierProfile: func(args, reply interface{}) error {
				*reply.(*string) = utils.OK
				return nil
			},
		},
	}
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
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
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
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
	if err := dm.MatchFilterIndexFromKey(utils.CacheResourceFilterIndexes, "cgrates.org:*string:Account:1002"); err == nil {
		t.Error(err)
	}
	//unifinished
}
