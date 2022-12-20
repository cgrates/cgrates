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
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func TestDmGetDestinationRemote(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetDataStorage(tmpDm)
	}()
	Cache.Clear(nil)
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().RmtConnID = "rmt"
	cfg.GeneralCfg().NodeID = "node"
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheDestinations: {
			Limit:   3,
			Remote:  true,
			APIKey:  "key",
			RouteID: "route",
		},
	}
	clientConn := make(chan rpcclient.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ReplicatorSv1GetDestination: func(args, reply interface{}) error {
				rpl := &Destination{
					Id: "nat", Prefixes: []string{"0257", "0256", "0723"},
				}
				*reply.(**Destination) = rpl
				return nil
			},
		},
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	exp := &Destination{
		Id: "nat", Prefixes: []string{"0257", "0256", "0723"},
	}
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	config.SetCgrConfig(cfg)
	SetDataStorage(dm)
	if val, err := dm.GetDestination("key", false, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(val, exp) {
		t.Errorf("expected %+v,received %+v", utils.ToJSON(exp), utils.ToJSON(val))
	}
}

func TestDmGetAccountRemote(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetDataStorage(tmpDm)
	}()
	Cache.Clear(nil)
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().RmtConnID = "rmt"
	cfg.GeneralCfg().NodeID = "node"

	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheAccounts: {
			Limit:   3,
			Remote:  true,
			APIKey:  "key",
			RouteID: "route",
		},
	}
	clientConn := make(chan rpcclient.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ReplicatorSv1GetAccount: func(args, reply interface{}) error {
				rpl := &Account{
					ID:         "cgrates.org:exp",
					UpdateTime: time.Now(),
				}
				*reply.(**Account) = rpl
				return nil
			},
		},
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	exp := &Account{
		ID:         "cgrates.org:exp",
		UpdateTime: time.Now(),
	}
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	config.SetCgrConfig(cfg)
	SetDataStorage(dm)
	if val, err := dm.GetAccount("id"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(val.ID, exp.ID) {
		t.Errorf("expected %+v,received %+v", utils.ToJSON(exp), utils.ToJSON(val))
	}
}

func TestDmGetFilterRemote(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetDataStorage(tmpDm)
	}()
	Cache.Clear(nil)
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().RmtConnID = "rmt"
	cfg.GeneralCfg().NodeID = "node"

	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheFilters: {
			Limit:   3,
			Remote:  true,
			APIKey:  "key",
			RouteID: "route",
		},
	}
	clientConn := make(chan rpcclient.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ReplicatorSv1GetFilter: func(args, reply interface{}) error {
				rpl := &Filter{
					Tenant: "cgrates.org",
					ID:     "Filter1",
					Rules: []*FilterRule{
						{
							Element: "~*req.Account",
							Type:    utils.MetaString,
							Values:  []string{"1001", "1002"},
						},
					},
					ActivationInterval: &utils.ActivationInterval{
						ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
						ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
					},
				}
				*reply.(**Filter) = rpl
				return nil
			},
		},
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})

	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	config.SetCgrConfig(cfg)
	SetDataStorage(dm)
	exp := &Filter{
		Tenant: "cgrates.org",
		ID:     "Filter1",
		Rules: []*FilterRule{
			{
				Element: "~*req.Account",
				Type:    utils.MetaString,
				Values:  []string{"1001", "1002"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	if val, err := dm.GetFilter("cgrates", "id2", false, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp.ID, val.ID) {
		t.Errorf("expected %+v,received %+v", utils.ToJSON(exp), utils.ToJSON(val))
	}
}

func TestDMGetThresholdRemote(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetDataStorage(tmpDm)
	}()
	Cache.Clear(nil)
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().RmtConnID = "rmt"
	cfg.GeneralCfg().NodeID = "node"
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheThresholds: {
			Limit:   3,
			Remote:  true,
			APIKey:  "key",
			RouteID: "route",
		},
	}
	clientConn := make(chan rpcclient.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ReplicatorSv1GetThreshold: func(args, reply interface{}) error {
				rpl := &Threshold{
					Tenant: "cgrates.org",
					ID:     "THD_ACNT_1001",
					Hits:   0,
				}
				*reply.(**Threshold) = rpl
				return nil
			},
		},
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	config.SetCgrConfig(cfg)
	SetDataStorage(dm)
	exp := &Threshold{
		Tenant: "cgrates.org",
		ID:     "THD_ACNT_1001",
		Hits:   0,
	}
	if val, err := dm.GetThreshold("cgrates", "id2", false, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, val) {
		t.Errorf("expected %+v,received %+v", utils.ToJSON(exp), utils.ToJSON(val))
	}
}
func TestDMGetThresholdProfileRemote(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetDataStorage(tmpDm)
	}()
	Cache.Clear(nil)
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().RmtConnID = "rmt"
	cfg.GeneralCfg().NodeID = "node"

	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheThresholdProfiles: {
			Limit:   3,
			Remote:  true,
			APIKey:  "key",
			RouteID: "route",
		},
	}
	clientConn := make(chan rpcclient.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ReplicatorSv1GetThresholdProfile: func(args, reply interface{}) error {
				rpl := &ThresholdProfile{
					Tenant: "cgrates.org",
					ID:     "ID",
				}
				*reply.(**ThresholdProfile) = rpl
				return nil
			},
		},
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	config.SetCgrConfig(cfg)
	SetDataStorage(dm)
	exp := &ThresholdProfile{
		Tenant: "cgrates.org",
		ID:     "ID",
	}
	if val, err := dm.GetThresholdProfile("cgrates", "id2", false, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, val) {
		t.Errorf("expected %+v,received %+v", utils.ToJSON(exp), utils.ToJSON(val))
	}
}

func TestDMGetStatQueue(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetDataStorage(tmpDm)
	}()
	Cache.Clear(nil)
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().RmtConnID = "rmt"
	cfg.GeneralCfg().NodeID = "node"

	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheStatQueues: {
			Limit:   3,
			Remote:  true,
			APIKey:  "key",
			RouteID: "route",
		},
	}
	clientConn := make(chan rpcclient.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ReplicatorSv1GetStatQueue: func(args, reply interface{}) error {
				rpl := &StatQueue{
					Tenant: "cgrates.org",
					ID:     "StatsID",
					SQItems: []SQItem{{
						EventID: "ev1",
					}},
				}
				*reply.(**StatQueue) = rpl
				return nil
			},
		},
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	dm.ms = &JSONMarshaler{}
	config.SetCgrConfig(cfg)
	SetDataStorage(dm)
	exp := &StatQueue{
		Tenant: "cgrates.org",
		ID:     "StatsID",
		SQItems: []SQItem{{
			EventID: "ev1",
		}},
	}
	if val, err := dm.GetStatQueue("cgrates", "id2", false, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, val) {
		t.Errorf("expected %+v,received %+v", utils.ToJSON(exp), utils.ToJSON(val))
	}
}

func TestRebuildReverseForPrefix(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetDataStorage(tmpDm)
	}()
	Cache.Clear(nil)
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheReverseDestinations: {
			Limit:  3,
			Remote: true,
		},
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	dm.dataDB = &DataDBMock{}
	db.db.Set(utils.CacheReverseDestinations, utils.ConcatenatedKey(utils.ReverseDestinationPrefix, "item1"), &Destination{}, []string{}, true, utils.NonTransactional)
	if err := dm.RebuildReverseForPrefix(utils.ReverseDestinationPrefix); err == nil || err != utils.ErrNotImplemented {
		t.Error(err)
	}
	dm.dataDB = db
	if err := dm.RebuildReverseForPrefix(utils.ReverseDestinationPrefix); err != nil {
		t.Error(err)
	}

}

func TestDMSetAccount(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetDataStorage(tmpDm)
	}()
	Cache.Clear(nil)
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().RplFiltered = true
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheAccounts: {
			Limit:     3,
			Replicate: true,
			APIKey:    "key",
			RouteID:   "route",
		},
	}
	acc := &Account{
		ID: "vdf:broker",
		BalanceMap: map[string]Balances{
			utils.MetaVoice: {
				&Balance{Value: 20 * float64(time.Second),
					DestinationIDs: utils.NewStringMap("NAT"),
					Weight:         10, RatingSubject: "rif"},
				&Balance{Value: 100 * float64(time.Second),
					DestinationIDs: utils.NewStringMap("RET"), Weight: 20},
			}},
	}
	clientConn := make(chan rpcclient.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ReplicatorSv1SetAccount: func(args, reply interface{}) error {
				accApiOpts, cancast := args.(AccountWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.dataDB.SetAccountDrv(accApiOpts.Account)

				return nil
			},
		},
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	dm.ms = &JSONMarshaler{}
	config.SetCgrConfig(cfg)
	SetDataStorage(dm)
	if err := dm.SetAccount(acc); err != nil {
		t.Error(err)
	}
	var dmnil *DataManager
	if err = dmnil.SetAccount(acc); err == nil || err != utils.ErrNoDatabaseConn {
		t.Error(err)
	}
	dm.dataDB = &DataDBMock{}
	if err = dm.SetAccount(acc); err == nil || err != utils.ErrNotImplemented {
		t.Error(err)
	}
}

func TestDMRemoveAccount(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetDataStorage(tmpDm)
	}()
	Cache.Clear(nil)
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().RplFiltered = true
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheAccounts: {
			Limit:     3,
			Replicate: true,
			APIKey:    "key",
			RouteID:   "route",
		},
	}
	acc := &Account{
		ID: "vdf:broker",
		BalanceMap: map[string]Balances{
			utils.MetaVoice: {
				&Balance{Value: 20 * float64(time.Second),
					DestinationIDs: utils.NewStringMap("NAT"),
					Weight:         10, RatingSubject: "rif"},
				&Balance{Value: 100 * float64(time.Second),
					DestinationIDs: utils.NewStringMap("RET"), Weight: 20},
			}},
	}
	if err = dm.dataDB.SetAccountDrv(acc); err != nil {
		t.Error(err)
	}

	clientConn := make(chan rpcclient.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ReplicatorSv1RemoveAccount: func(args, reply interface{}) error {
				strApiOpts, cancast := args.(utils.StringWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.dataDB.RemoveAccountDrv(strApiOpts.Arg)
				return nil
			},
		},
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	config.SetCgrConfig(cfg)
	SetDataStorage(dm)
	if err = dm.RemoveAccount(acc.ID); err != nil {
		t.Error(err)
	}
	var dmnil *DataManager
	if err = dmnil.RemoveAccount(acc.ID); err == nil || err != utils.ErrNoDatabaseConn {
		t.Error(err)
	}
	dm.dataDB = &DataDBMock{}
	if err = dm.RemoveAccount(acc.ID); err == nil || err != utils.ErrNotImplemented {
		t.Error(err)
	}
}

func TestDmSetFilter(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetDataStorage(tmpDm)
	}()
	Cache.Clear(nil)
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().RplFiltered = true
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheFilters: {
			Limit:     3,
			Replicate: true,
			APIKey:    "key",
			RouteID:   "route",
		},
	}
	filter := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_CP_1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Charger",
				Values:  []string{"ChargerProfile1"},
			},
		},
	}
	clientConn := make(chan rpcclient.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ReplicatorSv1SetFilter: func(args, reply interface{}) error {
				fltr, cancast := args.(FilterWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.dataDB.SetFilterDrv(fltr.Filter)
				return nil
			},
		},
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	config.SetCgrConfig(cfg)
	SetDataStorage(dm)
	if err := dm.SetFilter(filter, false); err != nil {
		t.Error(err)
	}
	var dmnil *DataManager
	if err = dmnil.SetFilter(filter, false); err == nil || err != utils.ErrNoDatabaseConn {
		t.Error(err)
	}
}

func TestDMSetThreshold(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetDataStorage(tmpDm)
	}()
	Cache.Clear(nil)
	Cache.Clear(nil)
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().RplFiltered = true
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheThresholds: {
			Limit:     3,
			Replicate: true,
			APIKey:    "key",
			RouteID:   "route",
		},
	}
	thS := &Threshold{
		Tenant: "cgrates.org",
		ID:     "THD_ACNT_1001",
		Hits:   0,
	}
	clientConn := make(chan rpcclient.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ReplicatorSv1SetThreshold: func(args, reply interface{}) error {
				thS, cancast := args.(ThresholdWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.dataDB.SetThresholdDrv(thS.Threshold)
				return nil
			},
		},
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	config.SetCgrConfig(cfg)
	SetDataStorage(dm)

	if err = dm.SetThreshold(thS); err != nil {
		t.Error(err)
	}
	dm.dataDB = &DataDBMock{}
	if err = dm.SetThreshold(thS); err == nil || err != utils.ErrNotImplemented {
		t.Error(err)
	}
}

func TestDmRemoveThreshold(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetDataStorage(tmpDm)
	}()

	Cache.Clear(nil)
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().RplFiltered = true
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheThresholds: {
			Limit:     3,
			Replicate: true,
			APIKey:    "key",
			RouteID:   "route",
		},
	}
	thS := &Threshold{
		Tenant: "cgrates.org",
		ID:     "THD_ACNT_1001",
		Hits:   0,
	}
	clientConn := make(chan rpcclient.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ReplicatorSv1RemoveThreshold: func(args, reply interface{}) error {
				tntApiOpts, cancast := args.(utils.TenantIDWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.dataDB.RemoveThresholdDrv(tntApiOpts.TenantID.Tenant, tntApiOpts.TenantID.ID)
				return nil
			},
		},
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	config.SetCgrConfig(cfg)
	SetDataStorage(dm)
	if err := dm.RemoveThreshold(thS.Tenant, thS.ID); err != nil {
		t.Error(err)
	}
	dm.dataDB = &DataDBMock{}
	if err = dm.RemoveThreshold(thS.Tenant, thS.ID); err == nil || err != utils.ErrNotImplemented {
		t.Error(err)
	}
}

func TestDMReverseDestinationRemote(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetDataStorage(tmpDm)
	}()
	Cache.Clear(nil)
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().RplFiltered = true
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheReverseDestinations: {
			Limit:     3,
			Replicate: true,
			APIKey:    "key",
			RouteID:   "route",
		},
	}
	clientConn := make(chan rpcclient.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ReplicatorSv1SetReverseDestination: func(args, reply interface{}) error {
				dest, cancast := args.(Destination)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.dataDB.SetReverseDestinationDrv(dest.Id, dest.Prefixes, utils.NonTransactional)
				return nil
			},
		},
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	config.SetCgrConfig(cfg)
	SetDataStorage(dm)
	dest := &Destination{
		Id: "nat", Prefixes: []string{"0257", "0256", "0723"},
	}
	if err := dm.SetReverseDestination(dest.Id, dest.Prefixes, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	exp := []string{"nat"}
	for _, prf := range dest.Prefixes {
		if val, err := dm.dataDB.GetReverseDestinationDrv(prf, utils.NonTransactional); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(val, exp) {
			t.Errorf("expected %v,received %v", exp, val)
		}
	}
}

func TestDMStatQueueRemote(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetDataStorage(tmpDm)
	}()
	Cache.Clear(nil)
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().RplFiltered = true
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheStatQueues: {
			Limit:     3,
			Replicate: true,
			APIKey:    "key",
			RouteID:   "route",
		},
	}
	clientConn := make(chan rpcclient.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ReplicatorSv1SetStatQueue: func(args, reply interface{}) error {
				sqApiOpts, cancast := args.(StatQueueWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.dataDB.SetStatQueueDrv(nil, sqApiOpts.StatQueue)
				return nil
			},
			utils.ReplicatorSv1RemoveStatQueue: func(args, reply interface{}) error {
				tntIDApiOpts, cancast := args.(utils.TenantIDWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.dataDB.RemStatQueueDrv(tntIDApiOpts.Tenant, tntIDApiOpts.ID)
				return nil
			},
		},
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	config.SetCgrConfig(cfg)
	SetDataStorage(dm)
	sq := &StatQueue{
		Tenant:  "cgrates.org",
		ID:      "SQ1",
		SQItems: []SQItem{},
		SQMetrics: map[string]StatMetric{
			utils.MetaTCD: &StatTCD{
				Events: make(map[string]*DurationWithCompress),
			},
		},
	}
	if err := dm.SetStatQueue(sq); err != nil {
		t.Error(err)
	}
	if val, err := dm.GetStatQueue(sq.Tenant, sq.ID, true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(val, sq) {
		t.Errorf("expected %v,received %v", utils.ToJSON(sq), utils.ToJSON(val))
	}
	if err = dm.RemoveStatQueue(sq.Tenant, sq.ID); err != nil {
		t.Error(err)
	}
	if _, has := db.db.Get(utils.CacheStatQueues, utils.ConcatenatedKey(sq.Tenant, sq.ID)); has {
		t.Error("should been removed")
	}
}

func TestDmTimingR(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetDataStorage(tmpDm)
	}()
	Cache.Clear(nil)
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().RplFiltered = true
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheTimings: {
			Limit:     3,
			Replicate: true,
			APIKey:    "key",
			RouteID:   "route",
		},
	}
	clientConn := make(chan rpcclient.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ReplicatorSv1SetTiming: func(args, reply interface{}) error {
				tpTimingApiOpts, cancast := args.(utils.TPTimingWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().SetTimingDrv(tpTimingApiOpts.TPTiming)
				return nil
			},
			utils.ReplicatorSv1RemoveTiming: func(args, reply interface{}) error {
				id, cancast := args.(string)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().RemoveTimingDrv(id)
				return nil
			},
		},
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	config.SetCgrConfig(cfg)
	SetDataStorage(dm)
	tp := &utils.TPTiming{
		ID:        "MIDNIGHT",
		Years:     utils.Years{2020, 2019},
		Months:    utils.Months{1, 2, 3, 4},
		MonthDays: utils.MonthDays{5, 6, 7, 8},
		WeekDays:  utils.WeekDays{0, 1, 2, 3},
		StartTime: "00:00:00",
		EndTime:   "00:00:01",
	}
	if err := dm.SetTiming(tp); err != nil {
		t.Error(err)
	}
	if val, err := dm.GetTiming(tp.ID, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tp, val) {
		t.Errorf("expected %v,received %v", utils.ToJSON(tp), utils.ToJSON(val))
	}
	if err = dm.RemoveTiming(tp.ID, utils.NonTransactional); err != nil {
		t.Error(err)
	}
}

func TestDMSetActionTriggers(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetDataStorage(tmpDm)
	}()
	Cache.Clear(nil)
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().RplFiltered = true
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheActionTriggers: {
			Limit:     3,
			Replicate: true,
			APIKey:    "key",
			RouteID:   "route",
		},
	}
	clientConn := make(chan rpcclient.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ReplicatorSv1SetActionTriggers: func(args, reply interface{}) error {
				setActTrgAOpts, cancast := args.(SetActionTriggersArgWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().SetActionTriggersDrv(setActTrgAOpts.Key, setActTrgAOpts.Attrs)
				return nil
			},
			utils.ReplicatorSv1RemoveActionTriggers: func(args, reply interface{}) error {
				strApiOpts, cancast := args.(utils.StringWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().RemoveActionTriggersDrv(strApiOpts.Arg)
				return nil
			},
		},
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	config.SetCgrConfig(cfg)
	SetDataStorage(dm)
	attrs := ActionTriggers{
		&ActionTrigger{
			Balance: &BalanceFilter{
				Type: utils.StringPointer(utils.MetaMonetary)},
			ThresholdValue: 2, ThresholdType: utils.TriggerMaxEventCounter,
			ActionsID: "TEST_ACTIONS"}}
	if err := dm.SetActionTriggers("act_ID", attrs); err != nil {
		t.Error(err)
	}
	if val, err := dm.GetActionTriggers("act_ID", false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(attrs, val) {
		t.Errorf("expected %v,received %v", utils.ToJSON(attrs), utils.ToJSON(val))
	}

	if err = dm.RemoveActionTriggers("act_ID", utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if _, has := db.db.Get(utils.CacheActionTriggers, "act_ID"); has {
		t.Error("should been removed")
	}
}

func TestDMResourceProfileRemote(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetDataStorage(tmpDm)
	}()
	Cache.Clear(nil)
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().RplFiltered = true
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheResourceProfiles: {
			Limit:     3,
			Replicate: true,
			APIKey:    "key",
			RouteID:   "route",
		},
		utils.CacheResources: {
			Limit:     3,
			Replicate: true,
			APIKey:    "key",
			RouteID:   "route",
		},
	}
	clientConn := make(chan rpcclient.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ReplicatorSv1SetResourceProfile: func(args, reply interface{}) error {
				rscPrflApiOpts, cancast := args.(ResourceProfileWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().SetResourceProfileDrv(rscPrflApiOpts.ResourceProfile)
				return nil
			},
			utils.ReplicatorSv1SetResource: func(args, reply interface{}) error {
				rscApiOpts, cancast := args.(ResourceWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().SetResourceDrv(rscApiOpts.Resource)
				return nil
			},
			utils.ReplicatorSv1RemoveResourceProfile: func(args, reply interface{}) error {
				tntApiOpts, cancast := args.(utils.TenantIDWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().RemoveResourceProfileDrv(tntApiOpts.Tenant, tntApiOpts.ID)
				return nil
			},
		},
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	config.SetCgrConfig(cfg)
	SetDataStorage(dm)
	rp := &ResourceProfile{
		Tenant:    "cgrates.org",
		ID:        "RES_ULTIMITED",
		FilterIDs: []string{"*string:~*req.CustomField:UnlimitedEvent"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	if err := dm.SetResourceProfile(rp, false); err != nil {
		t.Error(err)
	}
	if val, err := dm.GetResourceProfile(rp.Tenant, rp.ID, true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rp, val) {
		t.Errorf("expected %v,received %v", utils.ToJSON(rp), utils.ToJSON(val))
	}
	expRes := &Resource{
		Tenant: rp.Tenant,
		ID:     rp.ID,
		Usages: map[string]*ResourceUsage{},
	}
	if val, err := dm.GetResource(rp.Tenant, rp.ID, true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expRes, val) {
		t.Errorf("expected %v,received %v", utils.ToJSON(rp), utils.ToJSON(val))
	}
	if err := dm.RemoveResourceProfile(rp.Tenant, rp.ID, false); err != nil {
		t.Error(err)
	}
	if _, has := db.db.Get(utils.CacheResourceProfiles, utils.ConcatenatedKey(rp.Tenant, rp.ID)); has {
		t.Error("should been removed")
	}
}

func TestDmSharedGroup(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetDataStorage(tmpDm)
	}()
	Cache.Clear(nil)
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().RplFiltered = true
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheSharedGroups: {
			Limit:     3,
			Replicate: true,
			APIKey:    "key",
			RouteID:   "route",
		},
	}
	clientConn := make(chan rpcclient.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ReplicatorSv1SetSharedGroup: func(args, reply interface{}) error {
				shGrpApiOpts, cancast := args.(SharedGroupWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.dataDB.SetSharedGroupDrv(shGrpApiOpts.SharedGroup)
				return nil
			},
			utils.ReplicatorSv1RemoveSharedGroup: func(args, reply interface{}) error {
				strApiOpts, cancast := args.(utils.StringWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().RemoveSharedGroupDrv(strApiOpts.Arg)
				return nil
			},
		},
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	config.SetCgrConfig(cfg)
	SetDataStorage(dm)
	sg := &SharedGroup{
		Id: "SG2",
		AccountParameters: map[string]*SharingParameters{
			"*any": {
				Strategy:      "*lowest",
				RatingSubject: "one",
			},
		},
	}
	if err := dm.SetSharedGroup(sg); err != nil {
		t.Error(err)
	}
	if val, err := dm.GetSharedGroup(sg.Id, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(val, sg) {
		t.Errorf("expected %+v,received %+v", utils.ToJSON(sg), utils.ToJSON(val))
	}
	if err := dm.RemoveSharedGroup(sg.Id, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if _, has := db.db.Get(utils.CacheSharedGroups, sg.Id); has {
		t.Error("should been removed")
	}
}

func TestDMThresholdProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetDataStorage(tmpDm)
	}()
	Cache.Clear(nil)
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().RplFiltered = true
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheThresholdProfiles: {
			Limit:     3,
			Replicate: true,
			APIKey:    "key",
			RouteID:   "route",
		},
		utils.CacheThresholds: {
			Limit:     3,
			Replicate: true,
			APIKey:    "key",
			RouteID:   "route",
		},
	}
	clientConn := make(chan rpcclient.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ReplicatorSv1SetThresholdProfile: func(args, reply interface{}) error {
				thPApiOpts, cancast := args.(ThresholdProfileWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().SetThresholdProfileDrv(thPApiOpts.ThresholdProfile)
				return nil
			},
			utils.ReplicatorSv1SetThreshold: func(args, reply interface{}) error {
				thApiOpts, cancast := args.(ThresholdWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().SetThresholdDrv(thApiOpts.Threshold)
				return nil
			},
			utils.ReplicatorSv1RemoveThresholdProfile: func(args, reply interface{}) error {
				tntApiOpts, cancast := args.(utils.TenantIDWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().RemThresholdProfileDrv(tntApiOpts.Tenant, tntApiOpts.ID)
				return nil
			},
		},
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	config.SetCgrConfig(cfg)
	SetDataStorage(dm)
	th := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "THD_Test",
		FilterIDs: []string{"*string:Account:test"},
		MaxHits:   -1,
		MinSleep:  time.Second,
		Blocker:   false,
		Weight:    20.0,
		ActionIDs: []string{"ACT_LOG"},
		Async:     false,
	}
	if err := dm.SetThresholdProfile(th, false); err != nil {
		t.Error(err)
	}
	if val, err := dm.GetThresholdProfile(th.Tenant, th.ID, true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(th, val) {
		t.Errorf("expected %v,received %v", utils.ToJSON(th), utils.ToJSON(val))
	}
	if err := dm.RemoveThresholdProfile(th.Tenant, th.ID, false); err != nil {
		t.Error(err)
	}
	if _, has := db.db.Get(utils.CacheThresholdProfiles, utils.ConcatenatedKey(th.Tenant, th.ID)); has {
		t.Error("should receive error")
	}
}

func TestDmDispatcherHost(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetDataStorage(tmpDm)
	}()
	Cache.Clear(nil)
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().RplFiltered = true
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheDispatcherHosts: {
			Limit:     3,
			Replicate: true,
			APIKey:    "key",
			RouteID:   "route",
		},
	}
	clientConn := make(chan rpcclient.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ReplicatorSv1SetDispatcherHost: func(args, reply interface{}) error {
				dspApiOpts, cancast := args.(DispatcherHostWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().SetDispatcherHostDrv(dspApiOpts.DispatcherHost)
				return nil
			},
			utils.ReplicatorSv1RemoveDispatcherHost: func(args, reply interface{}) error {
				tntApiOpts, cancast := args.(utils.TenantIDWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().RemoveDispatcherHostDrv(tntApiOpts.Tenant, tntApiOpts.ID)
				return nil
			},
		},
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	config.SetCgrConfig(cfg)
	SetDataStorage(dm)
	dH := &DispatcherHost{
		Tenant: "testTenant",
		RemoteHost: &config.RemoteHost{
			ID:        "testID",
			Address:   rpcclient.InternalRPC,
			Transport: utils.MetaInternal,
			TLS:       false,
		},
	}
	if err := dm.SetDispatcherHost(dH); err != nil {
		t.Error(err)
	}
	if val, err := dm.GetDispatcherHost(dH.Tenant, dH.ID, true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dH, val) {
		t.Errorf("expected %v,received %v", utils.ToJSON(dH), utils.ToJSON(val))
	}
	if err = dm.RemoveDispatcherHost(dH.Tenant, dH.ID); err != nil {
		t.Error(err)
	}
	if _, has := db.db.Get(utils.CacheDispatcherHosts, utils.ConcatenatedKey(dH.Tenant, dH.ID)); has {
		t.Error("has not been removed from the cache")
	}
}

func TestChargerProfileRemote(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetDataStorage(tmpDm)
	}()
	Cache.Clear(nil)
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().RplFiltered = true
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheChargerProfiles: {
			Limit:     3,
			Replicate: true,
			APIKey:    "key",
			RouteID:   "route",
		},
	}
	clientConn := make(chan rpcclient.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ReplicatorSv1SetDispatcherHost: func(args, reply interface{}) error {
				chrgPrflApiOpts, cancast := args.(ChargerProfileWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().SetChargerProfileDrv(chrgPrflApiOpts.ChargerProfile)
				return nil
			},
			utils.ReplicatorSv1RemoveChargerProfile: func(args, reply interface{}) error {
				tntApiOpts, cancast := args.(utils.TenantIDWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().RemoveChargerProfileDrv(tntApiOpts.Tenant, tntApiOpts.ID)
				return nil
			},
		},
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	config.SetCgrConfig(cfg)
	SetDataStorage(dm)
	chrPrf := &ChargerProfile{
		Tenant:    "cgrates.org",
		ID:        "CPP_3",
		FilterIDs: []string{"FLTR_CP_3"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		RunID:        "*rated",
		AttributeIDs: []string{"ATTR_1"},
		Weight:       20,
	}
	if err := dm.SetChargerProfile(chrPrf, false); err != nil {
		t.Error(err)
	}
	if val, err := dm.GetChargerProfile(chrPrf.Tenant, chrPrf.ID, true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(chrPrf, val) {
		t.Errorf("expected %v,received %v", utils.ToJSON(val), utils.ToJSON(chrPrf))
	}
	if err = dm.RemoveChargerProfile(chrPrf.Tenant, chrPrf.ID, false); err != nil {
		t.Error(err)
	}
	if _, has := db.db.Get(utils.CacheChargerProfiles, chrPrf.TenantID()); has {
		t.Error("should been removed from the cache")
	}
}

func TestDispatcherProfileRemote(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetDataStorage(tmpDm)
	}()
	Cache.Clear(nil)
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().RplFiltered = true
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheDispatcherProfiles: {
			Limit:     3,
			Replicate: true,
			APIKey:    "key",
			RouteID:   "route",
		},
	}
	clientConn := make(chan rpcclient.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ReplicatorSv1SetDispatcherProfile: func(args, reply interface{}) error {
				dspApiOpts, cancast := args.(DispatcherProfileWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().SetDispatcherProfileDrv(dspApiOpts.DispatcherProfile)
				return nil
			},
			utils.ReplicatorSv1RemoveDispatcherProfile: func(args, reply interface{}) error {
				tntApiOpts, cancast := args.(utils.TenantIDWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().RemoveDispatcherProfileDrv(tntApiOpts.Tenant, tntApiOpts.ID)
				return nil
			},
		},
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	config.SetCgrConfig(cfg)
	SetDataStorage(dm)
	dsp := &DispatcherProfile{
		Tenant:     "cgrates.org",
		ID:         "Dsp1",
		FilterIDs:  []string{"*string:~*req.Account:1001"},
		Subsystems: []string{utils.MetaAny},
		Strategy:   utils.MetaFirst,
		Weight:     20,
	}
	if err := dm.SetDispatcherProfile(dsp, false); err != nil {
		t.Error(err)
	}
	if val, err := dm.GetDispatcherProfile(dsp.Tenant, dsp.ID, true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(val, dsp) {
		t.Errorf("expected %v,received %v", utils.ToJSON(dsp), utils.ToJSON(val))
	}
	if err := dm.RemoveDispatcherProfile(dsp.Tenant, dsp.ID, false); err != nil {
		t.Error(err)
	}
	if _, has := db.db.Get(utils.CacheDispatcherProfiles, dsp.TenantID()); has {
		t.Error("should been removed from the cache")
	}
}

func TestRouteProfileRemote(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetDataStorage(tmpDm)
	}()
	Cache.Clear(nil)
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().RplFiltered = true
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheRouteProfiles: {
			Limit:     3,
			Replicate: true,
			APIKey:    "key",
			RouteID:   "route",
		},
	}
	clientConn := make(chan rpcclient.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ReplicatorSv1SetRouteProfile: func(args, reply interface{}) error {
				routeApiOpts, cancast := args.(RouteProfileWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().SetRouteProfileDrv(routeApiOpts.RouteProfile)
				return nil
			},
			utils.ReplicatorSv1RemoveRouteProfile: func(args, reply interface{}) error {
				tntApiOpts, cancast := args.(utils.TenantIDWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().RemoveRouteProfileDrv(tntApiOpts.Tenant, tntApiOpts.ID)
				return nil
			},
		},
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	config.SetCgrConfig(cfg)
	SetDataStorage(dm)
	rpp := &RouteProfile{
		Tenant:    "cgrates.org",
		ID:        "ROUTE_ACNT_1002",
		FilterIDs: []string{"FLTR_ACNT_1002"},
	}
	if err := dm.SetRouteProfile(rpp, false); err != nil {
		t.Error(err)
	}
	if val, err := dm.GetRouteProfile(rpp.Tenant, rpp.ID, true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(val, rpp) {
		t.Errorf("expected %v,received %v", utils.ToJSON(rpp), utils.ToJSON(val))
	}
	if err := dm.RemoveRouteProfile(rpp.Tenant, rpp.ID, false); err != nil {
		t.Error(err)
	}
	if _, has := db.db.Get(utils.CacheRouteProfiles, rpp.TenantID()); has {
		t.Error("should been removed from the cache")
	}
}

func TestRatingPlanRemote(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetDataStorage(tmpDm)
	}()
	Cache.Clear(nil)
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().RplFiltered = true
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheRatingPlans: {
			Limit:     3,
			Replicate: true,
			APIKey:    "key",
			RouteID:   "route",
		},
	}
	clientConn := make(chan rpcclient.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ReplicatorSv1SetRatingPlan: func(args, reply interface{}) error {
				rPnApiOpts, cancast := args.(RatingPlanWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().SetRatingPlanDrv(rPnApiOpts.RatingPlan)
				return nil
			},
			utils.ReplicatorSv1RemoveRatingPlan: func(args, reply interface{}) error {
				strApiOpts, cancast := args.(utils.StringWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().RemoveRatingPlanDrv(strApiOpts.Arg)
				return nil
			},
		},
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	config.SetCgrConfig(cfg)
	SetDataStorage(dm)
	rP := &RatingPlan{
		Id: "RP1",
		Timings: map[string]*RITiming{
			"30eab300": {
				Years:     utils.Years{},
				Months:    utils.Months{},
				MonthDays: utils.MonthDays{},
				WeekDays:  utils.WeekDays{},
				StartTime: "00:00:00",
			},
		},
	}
	if err := dm.SetRatingPlan(rP); err != nil {
		t.Error(err)
	}
	if val, err := dm.GetRatingPlan(rP.Id, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(val, rP) {
		t.Errorf("expected %v,received %v", utils.ToJSON(rP), utils.ToJSON(val))
	}
	if err = dm.RemoveRatingPlan(rP.Id, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if _, has := db.db.Get(utils.CacheRatingPlans, rP.Id); has {
		t.Error("should been removed from the caches")
	}
}

func TestGetResourceRemote(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetDataStorage(tmpDm)
	}()
	Cache.Clear(nil)
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.DataDbCfg().RplFiltered = true
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheResources: {
			Limit:     3,
			Replicate: true,
			APIKey:    "key",
			RouteID:   "route",
			Remote:    true,
		},
	}

	rS := &Resource{
		Tenant: "cgrates.org",
		ID:     "ResGroup3",
		Usages: map[string]*ResourceUsage{
			"651a8db2-4f67-4cf8-b622-169e8a482e21": {
				Tenant: "cgrates.org",
				ID:     "651a8db2-4f67-4cf8-b622-169e8a482e21",
				Units:  2,
			},
		},
	}
	clientConn := make(chan rpcclient.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ReplicatorSv1GetResource: func(args, reply interface{}) error {
				tntApiOpts, cancast := args.(*utils.TenantIDWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().GetResourceDrv(tntApiOpts.Tenant, tntApiOpts.ID)
				*reply.(**Resource) = rS
				return nil
			},
		},
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator): clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	config.SetCgrConfig(cfg)
	SetDataStorage(dm)

	if val, err := dm.GetResource(rS.Tenant, rS.ID, false, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rS, val) {
		t.Errorf("expected %v,received %v", utils.ToJSON(rS), utils.ToJSON(val))
	}
}

func TestGetResourceProfileRemote(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetDataStorage(tmpDm)
	}()
	Cache.Clear(nil)
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.DataDbCfg().RplFiltered = true
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheResourceProfiles: {
			Limit:     3,
			Replicate: true,
			APIKey:    "key",
			RouteID:   "route",
			Remote:    true,
		},
	}
	rsP := &ResourceProfile{
		ID:                "RES_ULTIMITED2",
		UsageTTL:          time.Nanosecond,
		Limit:             10,
		AllocationMessage: "MessageAllocation",
		Blocker:           true,
		Stored:            true,
		Weight:            20,
		ThresholdIDs:      []string{"Val1"},
	}
	clientConn := make(chan rpcclient.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ReplicatorSv1GetResourceProfile: func(args, reply interface{}) error {
				tntApiOpts, cancast := args.(*utils.TenantIDWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().GetResourceProfileDrv(tntApiOpts.Tenant, tntApiOpts.ID)
				*reply.(**ResourceProfile) = rsP
				return nil
			},
		},
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator): clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	config.SetCgrConfig(cfg)
	SetDataStorage(dm)
	if val, err := dm.GetResourceProfile(rsP.Tenant, rsP.ID, false, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(val, rsP) {
		t.Errorf("expected %v,received %v", utils.ToJSON(rsP), utils.ToJSON(val))
	}
}

// func TestGetActionTriggers(t *testing.T) {
// 	cfg := config.NewDefaultCGRConfig()
// 	tmpDm := dm
// 	tmp := Cache
// 	defer func() {
// 		config.SetCgrConfig(config.NewDefaultCGRConfig())
// 		Cache = tmp
// 		SetDataStorage(tmpDm)
// 	}()
// 	Cache.Clear(nil)
// 	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
// 	cfg.DataDbCfg().RplFiltered = true
// 	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
// 		utils.CacheActionTriggers: {
// 			Limit:     3,
// 			Replicate: true,
// 			APIKey:    "key",
// 			RouteID:   "route",
// 			Remote:    true,
// 		},
// 	}
// 	aT := ActionTriggers{
// 		&ActionTrigger{
// 			ID: "Test",
// 		},
// 	}
// 	clientConn := make(chan rpcclient.ClientConnector, 1)
// 	clientConn <- &ccMock{
// 		calls: map[string]func(args interface{}, reply interface{}) error{
// 			utils.ReplicatorSv1GetActionTriggers: func(args, reply interface{}) error {
// 				strApiOpts, cancast := args.(*utils.StringWithAPIOpts)
// 				if !cancast {
// 					return utils.ErrNotConvertible
// 				}
// 				dm.DataDB().GetActionTriggersDrv(strApiOpts.Arg)
// 				*reply.(*ActionTriggers) = aT
// 				return nil
// 			},
// 		},
// 	}
// 	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
// 	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
// 		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
// 	})
// 	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
// 	config.SetCgrConfig(cfg)
// 	SetDataStorage(dm)
// 	Cache.Set(utils.CacheActionTriggers, "Test", ActionTriggers{}, []string{}, false, utils.NonTransactional)
// 	if _, err := dm.GetActionTriggers(aT[0].ID, false, utils.NonTransactional); err != nil {
// 		t.Error(err)
// 	}
// }

func TestGetSharedGroupRemote(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetDataStorage(tmpDm)
	}()
	Cache.Clear(nil)
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.DataDbCfg().RplFiltered = true
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheSharedGroups: {
			Limit:     3,
			Replicate: true,
			APIKey:    "key",
			RouteID:   "route",
			Remote:    true,
		},
	}
	shG := &SharedGroup{
		Id: "testID",
		AccountParameters: map[string]*SharingParameters{
			"string1": {
				Strategy:      "strategyTEST1",
				RatingSubject: "RatingSubjectTEST1",
			},
			"string2": {
				Strategy:      "strategyTEST2",
				RatingSubject: "RatingSubjectTEST2",
			},
		},
		MemberIds: utils.StringMap{
			"string1": true,
			"string2": false,
		},
	}
	clientConn := make(chan rpcclient.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ReplicatorSv1GetSharedGroup: func(args, reply interface{}) error {
				strApiOpts, cancast := args.(*utils.StringWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().GetSharedGroupDrv(strApiOpts.Arg)
				*reply.(**SharedGroup) = shG
				return nil
			},
		},
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator): clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	config.SetCgrConfig(cfg)
	SetDataStorage(dm)
	if val, err := dm.GetSharedGroup(shG.Id, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(val, shG) {
		t.Errorf("expected %v,received %v", utils.ToJSON(shG), utils.ToJSON(val))
	}
}

func TestGetStatQueueProfileRemote(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetDataStorage(tmpDm)
	}()
	Cache.Clear(nil)
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.DataDbCfg().RplFiltered = true
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheStatQueueProfiles: {
			Limit:     3,
			Replicate: true,
			APIKey:    "key",
			RouteID:   "route",
			Remote:    true,
		},
	}
	sqP := &StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "SQ_ID",
		FilterIDs: []string{"FLTR_ID"},
		Weight:    10,
	}
	clientConn := make(chan rpcclient.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ReplicatorSv1GetStatQueueProfile: func(args, reply interface{}) error {
				tntApiOpts, cancast := args.(*utils.TenantIDWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().GetStatQueueProfileDrv(tntApiOpts.Tenant, tntApiOpts.ID)
				*reply.(**StatQueueProfile) = sqP
				return nil
			},
		},
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator): clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	config.SetCgrConfig(cfg)
	SetDataStorage(dm)
	if val, err := dm.GetStatQueueProfile(sqP.Tenant, sqP.ID, true, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(val, sqP) {
		t.Errorf("expected %v,received %v", utils.ToJSON(sqP), utils.ToJSON(val))
	}
}

func TestStatQueueProfileRemote(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetDataStorage(tmpDm)
	}()
	Cache.Clear(nil)
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.DataDbCfg().RplFiltered = true
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheStatQueueProfiles: {
			Limit:     3,
			Replicate: true,
			APIKey:    "key",
			RouteID:   "route",
			Remote:    true,
		},
		utils.CacheStatQueues: {
			Limit:     3,
			Replicate: true,
			APIKey:    "key",
			RouteID:   "route",
			Remote:    true,
		},
	}
	sqP := &StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "SQ_ID",
		FilterIDs: []string{"FLTR_ID"},
		Weight:    10,
	}
	clientConn := make(chan rpcclient.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ReplicatorSv1SetStatQueueProfile: func(args, reply interface{}) error {
				sqPApiOpts, cancast := args.(StatQueueProfileWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().SetStatQueueProfileDrv(sqPApiOpts.StatQueueProfile)
				return nil
			},
			utils.ReplicatorSv1SetStatQueue: func(args, reply interface{}) error {
				sqApiOpts, cancast := args.(StatQueueWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().SetStatQueueDrv(nil, sqApiOpts.StatQueue)
				return nil
			},
		},
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator): clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	config.SetCgrConfig(cfg)
	SetDataStorage(dm)
	Cache.Set(utils.CacheStatQueueProfiles, utils.ConcatenatedKey(sqP.Tenant, sqP.ID), &StatQueueProfile{
		QueueLength: 2,
	}, []string{}, true, utils.NonTransactional)
	if err := dm.SetStatQueueProfile(sqP, false); err != nil {
		t.Error(err)
	}
}
