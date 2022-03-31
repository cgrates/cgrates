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
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestRemoveFilterIndexes(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	connMgr := engine.NewConnManager(cfg)
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	adms := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: connMgr,
		ping:    struct{}{},
	}
	var reply string

	// Thresholds
	args := &AttrRemFilterIndexes{
		Tenant:   "cgrates.org",
		Context:  "cgrates",
		ItemType: utils.MetaThresholds,
		APIOpts: map[string]interface{}{
			utils.MetaCache: utils.MetaNone,
		},
	}
	if err := adms.RemoveFilterIndexes(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
	if reply != utils.OK {
		t.Error("Expected OK")
	}
	if args.ItemType != utils.CacheThresholdFilterIndexes {
		t.Errorf("Expected %v\n but received %v", utils.CacheThresholdFilterIndexes, args.ItemType)
	}

	//Routes
	args = &AttrRemFilterIndexes{
		Tenant:   "cgrates.org",
		Context:  "cgrates",
		ItemType: utils.MetaRoutes,
		APIOpts: map[string]interface{}{
			utils.MetaCache: utils.MetaNone,
		},
	}
	if err := adms.RemoveFilterIndexes(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
	if reply != utils.OK {
		t.Error("Expected OK")
	}
	if args.ItemType != utils.CacheRouteFilterIndexes {
		t.Errorf("Expected %v\n but received %v", utils.CacheRouteFilterIndexes, args.ItemType)
	}

	//Stats
	args = &AttrRemFilterIndexes{
		Tenant:   "cgrates.org",
		Context:  "cgrates",
		ItemType: utils.MetaStats,
		APIOpts: map[string]interface{}{
			utils.MetaCache: utils.MetaNone,
		},
	}
	if err := adms.RemoveFilterIndexes(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
	if reply != utils.OK {
		t.Error("Expected OK")
	}
	if args.ItemType != utils.CacheStatFilterIndexes {
		t.Errorf("Expected %v\n but received %v", utils.CacheStatFilterIndexes, args.ItemType)
	}

	//Resources
	args = &AttrRemFilterIndexes{
		Tenant:   "cgrates.org",
		Context:  "cgrates",
		ItemType: utils.MetaResources,
		APIOpts: map[string]interface{}{
			utils.MetaCache: utils.MetaNone,
		},
	}
	if err := adms.RemoveFilterIndexes(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
	if reply != utils.OK {
		t.Error("Expected OK")
	}
	if args.ItemType != utils.CacheResourceFilterIndexes {
		t.Errorf("Expected %v\n but received %v", utils.CacheResourceFilterIndexes, args.ItemType)
	}

	//Chargers
	args = &AttrRemFilterIndexes{
		Tenant:   "cgrates.org",
		Context:  "cgrates",
		ItemType: utils.MetaChargers,
		APIOpts: map[string]interface{}{
			utils.MetaCache: utils.MetaNone,
		},
	}
	if err := adms.RemoveFilterIndexes(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
	if reply != utils.OK {
		t.Error("Expected OK")
	}
	if args.ItemType != utils.CacheChargerFilterIndexes {
		t.Errorf("Expected %v\n but received %v", utils.CacheChargerFilterIndexes, args.ItemType)
	}

	//Accounts
	args = &AttrRemFilterIndexes{
		Tenant:   "cgrates.org",
		Context:  "cgrates",
		ItemType: utils.MetaAccounts,
		APIOpts: map[string]interface{}{
			utils.MetaCache: utils.MetaNone,
		},
	}
	if err := adms.RemoveFilterIndexes(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
	if reply != utils.OK {
		t.Error("Expected OK")
	}
	if args.ItemType != utils.CacheAccountsFilterIndexes {
		t.Errorf("Expected %v\n but received %v", utils.CacheAccountsFilterIndexes, args.ItemType)
	}

	//Actions
	args = &AttrRemFilterIndexes{
		Tenant:   "cgrates.org",
		Context:  "cgrates",
		ItemType: utils.MetaActions,
		APIOpts: map[string]interface{}{
			utils.MetaCache: utils.MetaNone,
		},
	}
	if err := adms.RemoveFilterIndexes(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
	if reply != utils.OK {
		t.Error("Expected OK")
	}
	if args.ItemType != utils.CacheActionProfilesFilterIndexes {
		t.Errorf("Expected %v\n but received %v", utils.CacheActionProfilesFilterIndexes, args.ItemType)
	}

	//RateProfile
	args = &AttrRemFilterIndexes{
		Tenant:   "cgrates.org",
		Context:  "cgrates",
		ItemType: utils.MetaRateProfiles,
		APIOpts: map[string]interface{}{
			utils.MetaCache: utils.MetaNone,
		},
	}
	if err := adms.RemoveFilterIndexes(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
	if reply != utils.OK {
		t.Error("Expected OK")
	}
	if args.ItemType != utils.CacheRateProfilesFilterIndexes {
		t.Errorf("Expected %v\n but received %v", utils.CacheRateProfilesFilterIndexes, args.ItemType)
	}

	//RateProfileRates
	args = &AttrRemFilterIndexes{
		Tenant:   "cgrates.org",
		Context:  "cgrates",
		ItemType: utils.MetaRateProfileRates,
		APIOpts: map[string]interface{}{
			utils.MetaCache: utils.MetaNone,
		},
	}
	if err := adms.RemoveFilterIndexes(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
	if reply != utils.OK {
		t.Error("Expected OK")
	}
	if args.ItemType != utils.CacheRateFilterIndexes {
		t.Errorf("Expected %v\n but received %v", utils.CacheRateFilterIndexes, args.ItemType)
	}

	//Dispatchers
	args = &AttrRemFilterIndexes{
		Tenant:   "cgrates.org",
		Context:  "cgrates",
		ItemType: utils.MetaDispatchers,
		APIOpts: map[string]interface{}{
			utils.MetaCache: utils.MetaNone,
		},
	}
	if err := adms.RemoveFilterIndexes(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
	if reply != utils.OK {
		t.Error("Expected OK")
	}
	if args.ItemType != utils.CacheDispatcherFilterIndexes {
		t.Errorf("Expected %v\n but received %v", utils.DispatcherFilterIndexes, args.ItemType)
	}

	//Attributes
	args = &AttrRemFilterIndexes{
		Tenant:   "cgrates.org",
		Context:  "cgrates",
		ItemType: utils.MetaAttributes,
		APIOpts: map[string]interface{}{
			utils.MetaCache: utils.MetaNone,
		},
	}
	if err := adms.RemoveFilterIndexes(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
	if reply != utils.OK {
		t.Error("Expected OK")
	}
	if args.ItemType != utils.CacheAttributeFilterIndexes {
		t.Errorf("Expected %v\n but received %v", utils.AttributeFilterIndexes, args.ItemType)
	}
}

// func TestGetFilterIndexes(t *testing.T) {
// cfg := config.NewDefaultCGRConfig()
// data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
// dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
// connMgr := engine.NewConnManager(cfg)
// cfg.AdminSCfg().CachesConns = []string{"*internal"}
// adms := &AdminSv1{
// 	cfg:     cfg,
// 	dm:      dm,
// 	connMgr: connMgr,
// 	ping:    struct{}{},
// }

// 	adms.dm.SetIndexes(context.Background(), utils.CacheThresholdFilterIndexes, "cgrates", nil, true, utils.GenUUID())

// 	var reply []string
// 	args := &AttrGetFilterIndexes{
// 		Tenant:   "cgrates.org",
// 		Context:  "cgrates",
// 		ItemType: utils.MetaThresholds,
// 		APIOpts: map[string]interface{}{
// 			utils.MetaCache: utils.MetaNone,
// 		},
// 	}

// 	if err := adms.GetFilterIndexes(context.Background(), args, &reply); err != nil {
// 		t.Error(err)
// 	}
// }

func TestComputeFilterIndexes(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	connMgr := engine.NewConnManager(cfg)
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	adms := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: connMgr,
		ping:    struct{}{},
	}

	var reply string

	//Thresholds
	args := &utils.ArgsComputeFilterIndexes{
		Tenant:     "cgrates.org",
		ThresholdS: true,
		APIOpts: map[string]interface{}{
			utils.MetaCache: utils.MetaNone,
		},
	}

	if err := adms.ComputeFilterIndexes(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}

	//StatQueueProfile
	args = &utils.ArgsComputeFilterIndexes{
		Tenant: "cgrates.org",
		StatS:  true,
		APIOpts: map[string]interface{}{
			utils.MetaCache: utils.MetaNone,
		},
	}

	if err := adms.ComputeFilterIndexes(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}

	//ResourceProfile
	args = &utils.ArgsComputeFilterIndexes{
		Tenant:    "cgrates.org",
		ResourceS: true,
		APIOpts: map[string]interface{}{
			utils.MetaCache: utils.MetaNone,
		},
	}

	if err := adms.ComputeFilterIndexes(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}

	//Routes
	args = &utils.ArgsComputeFilterIndexes{
		Tenant: "cgrates.org",
		RouteS: true,
		APIOpts: map[string]interface{}{
			utils.MetaCache: utils.MetaNone,
		},
	}

	if err := adms.ComputeFilterIndexes(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}

	//AttributeProfile
	args = &utils.ArgsComputeFilterIndexes{
		Tenant:     "cgrates.org",
		AttributeS: true,
		APIOpts: map[string]interface{}{
			utils.MetaCache: utils.MetaNone,
		},
	}

	if err := adms.ComputeFilterIndexes(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}

	//ChargerProfile
	args = &utils.ArgsComputeFilterIndexes{
		Tenant:   "cgrates.org",
		ChargerS: true,
		APIOpts: map[string]interface{}{
			utils.MetaCache: utils.MetaNone,
		},
	}

	if err := adms.ComputeFilterIndexes(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}

	//Account
	args = &utils.ArgsComputeFilterIndexes{
		Tenant:   "cgrates.org",
		AccountS: true,
		APIOpts: map[string]interface{}{
			utils.MetaCache: utils.MetaNone,
		},
	}

	if err := adms.ComputeFilterIndexes(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}

	///Actions
	args = &utils.ArgsComputeFilterIndexes{
		Tenant:  "cgrates.org",
		ActionS: true,
		APIOpts: map[string]interface{}{
			utils.MetaCache: utils.MetaNone,
		},
	}

	if err := adms.ComputeFilterIndexes(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}

	//Rates
	args = &utils.ArgsComputeFilterIndexes{
		Tenant: "cgrates.org",
		RateS:  true,
		APIOpts: map[string]interface{}{
			utils.MetaCache: utils.MetaNone,
		},
	}

	if err := adms.ComputeFilterIndexes(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}

	//DispatcherProfile
	args = &utils.ArgsComputeFilterIndexes{
		Tenant:      "cgrates.org",
		DispatcherS: true,
		APIOpts: map[string]interface{}{
			utils.MetaCache: utils.MetaNone,
		},
	}

	if err := adms.ComputeFilterIndexes(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
}
