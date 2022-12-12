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
