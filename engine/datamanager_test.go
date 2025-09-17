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
	"errors"
	"fmt"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
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
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetDestination: func(ctx *context.Context, args, reply any) error {
				rpl := &Destination{
					Id: "nat", Prefixes: []string{"0257", "0256", "0723"},
				}
				*reply.(**Destination) = rpl
				return nil
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
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
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetAccount: func(ctx *context.Context, args, reply any) error {
				rpl := &Account{
					ID:         "cgrates.org:exp",
					UpdateTime: time.Now(),
				}
				*reply.(**Account) = rpl
				return nil
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
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
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetFilter: func(ctx *context.Context, args, reply any) error {
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
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})

	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
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
	cfg.CacheCfg().Partitions[utils.CacheThresholds].Replicate = true
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheThresholds: {
			Limit:     3,
			Remote:    true,
			APIKey:    "key",
			RouteID:   "route",
			Replicate: true,
		},
	}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetThreshold: func(ctx *context.Context, args, reply any) error {
				rpl := &Threshold{
					Tenant: "cgrates.org",
					ID:     "THD_ACNT_1001",
					Hits:   0,
				}
				*reply.(**Threshold) = rpl
				return nil
			},
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error {

				return errors.New("Can't Replicate")
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches):    clientConn,
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
	Cache.Set(utils.CacheThresholds, utils.ConcatenatedKey("cgrates.org", "THD_ACNT_1001"), nil, []string{}, false, utils.NonTransactional)
	if _, err := dm.GetThreshold("cgrates.org", "THD_ACNT_1001", true, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	SetConnManager(connMgr)
	Cache = NewCacheS(cfg, dm, nil)
	if _, err := dm.GetThreshold("cgrates", "id2", false, true, utils.NonTransactional); err == nil || err.Error() != "Can't Replicate" {
		t.Error(err)
	}
}

func TestDMGetThresholdRemoteErr(t *testing.T) {
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
	cfg.CacheCfg().Partitions[utils.CacheThresholds].Replicate = true
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheThresholds: {
			Limit:     3,
			Remote:    true,
			APIKey:    "key",
			RouteID:   "route",
			Replicate: true,
		},
	}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetThreshold: func(ctx *context.Context, args, reply any) error {
				return utils.ErrNotFound
			},
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error {
				return errors.New("Can't Replicate")
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches):    clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	config.SetCgrConfig(cfg)
	SetDataStorage(dm)
	SetConnManager(connMgr)
	Cache = NewCacheS(cfg, dm, nil)
	if _, err := dm.GetThreshold("cgrates", "id2", false, true, utils.NonTransactional); err == nil || err.Error() != "Can't Replicate" {
		t.Error(err)
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
	cfg.CacheCfg().Partitions[utils.CacheThresholdProfiles].Replicate = true
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheThresholdProfiles: {
			Limit:     3,
			Remote:    true,
			APIKey:    "key",
			RouteID:   "route",
			Replicate: true,
		},
	}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetThresholdProfile: func(ctx *context.Context, args, reply any) error {
				rpl := &ThresholdProfile{
					Tenant: "cgrates.org",
					ID:     "ID",
				}
				*reply.(**ThresholdProfile) = rpl
				return nil
			},
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error {

				return errors.New("Can't Replicate")
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches):    clientConn,
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
	Cache = NewCacheS(cfg, dm, nil)
	SetConnManager(connMgr)
	if _, err := dm.GetThresholdProfile("cgrates", "id2", false, true, utils.NonTransactional); err == nil || err.Error() != "Can't Replicate" {
		t.Error(err)
	}
}
func TestDMGetThresholdProfileRemoteErr(t *testing.T) {
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
	cfg.CacheCfg().Partitions[utils.CacheThresholdProfiles].Replicate = true
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheThresholdProfiles: {
			Limit:     3,
			Remote:    true,
			APIKey:    "key",
			RouteID:   "route",
			Replicate: true,
		},
	}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetThresholdProfile: func(ctx *context.Context, args, reply any) error {
				return utils.ErrNotFound
			},
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error {
				return errors.New("Can't Replicate")
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches):    clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	config.SetCgrConfig(cfg)
	SetDataStorage(dm)
	SetConnManager(connMgr)
	Cache = NewCacheS(cfg, dm, nil)
	if _, err := dm.GetThresholdProfile("cgrates", "id2", false, true, utils.NonTransactional); err == nil || err.Error() != "Can't Replicate" {
		t.Error(err)
	}
	var dm2 *DataManager
	if _, err := dm2.GetThresholdProfile("cgrates", "id2", false, true, utils.NonTransactional); err == nil || err != utils.ErrNoDatabaseConn {
		t.Error(err)
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
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetStatQueue: func(ctx *context.Context, args, reply any) error {
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
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
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
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
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
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1SetAccount: func(ctx *context.Context, args, reply any) error {
				accApiOpts, cancast := args.(AccountWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.dataDB.SetAccountDrv(accApiOpts.Account)

				return nil
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	dm.ms = &JSONMarshaler{}
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

	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1RemoveAccount: func(ctx *context.Context, args, reply any) error {
				strApiOpts, cancast := args.(utils.StringWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.dataDB.RemoveAccountDrv(strApiOpts.Arg)
				return nil
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
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
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1SetFilter: func(ctx *context.Context, args, reply any) error {
				fltr, cancast := args.(FilterWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.dataDB.SetFilterDrv(fltr.Filter)
				return nil
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
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
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1SetThreshold: func(ctx *context.Context, args, reply any) error {
				thS, cancast := args.(ThresholdWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.dataDB.SetThresholdDrv(thS.Threshold)
				return nil
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
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
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1RemoveThreshold: func(ctx *context.Context, args, reply any) error {
				tntApiOpts, cancast := args.(utils.TenantIDWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.dataDB.RemoveThresholdDrv(tntApiOpts.TenantID.Tenant, tntApiOpts.TenantID.ID)
				return nil
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
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
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1SetReverseDestination: func(ctx *context.Context, args, reply any) error {
				dest, cancast := args.(Destination)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.dataDB.SetReverseDestinationDrv(dest.Id, dest.Prefixes, utils.NonTransactional)
				return nil
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
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
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1SetStatQueue: func(ctx *context.Context, args, reply any) error {
				sqApiOpts, cancast := args.(StatQueueWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.dataDB.SetStatQueueDrv(nil, sqApiOpts.StatQueue)
				return nil
			},
			utils.ReplicatorSv1RemoveStatQueue: func(ctx *context.Context, args, reply any) error {
				tntIDApiOpts, cancast := args.(utils.TenantIDWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.dataDB.RemStatQueueDrv(tntIDApiOpts.Tenant, tntIDApiOpts.ID)
				return nil
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
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
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1SetTiming: func(ctx *context.Context, args, reply any) error {
				tpTimingApiOpts, cancast := args.(utils.TPTimingWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().SetTimingDrv(tpTimingApiOpts.TPTiming)
				return nil
			},
			utils.ReplicatorSv1RemoveTiming: func(ctx *context.Context, args, reply any) error {
				id, cancast := args.(string)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().RemoveTimingDrv(id)
				return nil
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
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
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1SetActionTriggers: func(ctx *context.Context, args, reply any) error {
				setActTrgAOpts, cancast := args.(SetActionTriggersArgWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().SetActionTriggersDrv(setActTrgAOpts.Key, setActTrgAOpts.Attrs)
				return nil
			},
			utils.ReplicatorSv1RemoveActionTriggers: func(ctx *context.Context, args, reply any) error {
				strApiOpts, cancast := args.(utils.StringWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().RemoveActionTriggersDrv(strApiOpts.Arg)
				return nil
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
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
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1SetResourceProfile: func(ctx *context.Context, args, reply any) error {
				rscPrflApiOpts, cancast := args.(ResourceProfileWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().SetResourceProfileDrv(rscPrflApiOpts.ResourceProfile)
				return nil
			},
			utils.ReplicatorSv1SetResource: func(ctx *context.Context, args, reply any) error {
				rscApiOpts, cancast := args.(ResourceWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().SetResourceDrv(rscApiOpts.Resource)
				return nil
			},
			utils.ReplicatorSv1RemoveResourceProfile: func(ctx *context.Context, args, reply any) error {
				tntApiOpts, cancast := args.(utils.TenantIDWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().RemoveResourceProfileDrv(tntApiOpts.Tenant, tntApiOpts.ID)
				return nil
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
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
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1SetSharedGroup: func(ctx *context.Context, args, reply any) error {
				shGrpApiOpts, cancast := args.(SharedGroupWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.dataDB.SetSharedGroupDrv(shGrpApiOpts.SharedGroup)
				return nil
			},
			utils.ReplicatorSv1RemoveSharedGroup: func(ctx *context.Context, args, reply any) error {
				strApiOpts, cancast := args.(utils.StringWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().RemoveSharedGroupDrv(strApiOpts.Arg)
				return nil
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
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
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1SetThresholdProfile: func(ctx *context.Context, args, reply any) error {
				thPApiOpts, cancast := args.(ThresholdProfileWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().SetThresholdProfileDrv(thPApiOpts.ThresholdProfile)
				return nil
			},
			utils.ReplicatorSv1SetThreshold: func(ctx *context.Context, args, reply any) error {
				thApiOpts, cancast := args.(ThresholdWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().SetThresholdDrv(thApiOpts.Threshold)
				return nil
			},
			utils.ReplicatorSv1RemoveThresholdProfile: func(ctx *context.Context, args, reply any) error {
				tntApiOpts, cancast := args.(utils.TenantIDWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().RemThresholdProfileDrv(tntApiOpts.Tenant, tntApiOpts.ID)
				return nil
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
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

func TestDMRemoveThresholdProfileErr(t *testing.T) {
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
		utils.MetaThresholdProfiles: {
			Remote: true,
		},
	}
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetThresholdProfile: func(ctx *context.Context, args, reply any) error {
				return fmt.Errorf("Can't Replicate")
			},
		},
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator): clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	Cache.Set(utils.MetaThresholdProfiles, "cgrates.org:TEST_PROFILE1", nil, []string{}, true, utils.NonTransactional)
	if err := dm.RemoveThresholdProfile("cgrates.org", "TEST_PROFILE1", true); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	Cache.Remove(utils.MetaThresholdProfiles, "cgrates.org:TEST_PROFILE1", true, utils.NonTransactional)
	var dm2 *DataManager
	if err = dm2.RemoveThresholdProfile("cgrates.org", "TEST_PROFILE1", true); err == nil || err != utils.ErrNoDatabaseConn {
		t.Error(err)
	}
	dm2 = NewDataManager(db, cfg.CacheCfg(), nil)
	dm2.dataDB = &DataDBMock{
		GetThresholdProfileDrvF: func(tenant, id string) (tp *ThresholdProfile, err error) {
			return
		},
		RemThresholdProfileDrvF: func(tenant, id string) (err error) {
			return utils.ErrNotImplemented
		},
	}
	if err = dm2.RemoveThresholdProfile("cgrates.org", "TEST_PROFILE1", true); err == nil || err != utils.ErrNotImplemented {
		t.Error(err)
	}
	config.SetCgrConfig(cfg)
	if err = dm.RemoveThresholdProfile("cgrates.org", "TEST_PROFILE1", true); err == nil || err.Error() != "Can't Replicate" {
		t.Error(err)
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
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1SetDispatcherHost: func(ctx *context.Context, args, reply any) error {
				dspApiOpts, cancast := args.(DispatcherHostWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().SetDispatcherHostDrv(dspApiOpts.DispatcherHost)
				return nil
			},
			utils.ReplicatorSv1RemoveDispatcherHost: func(ctx *context.Context, args, reply any) error {
				tntApiOpts, cancast := args.(utils.TenantIDWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().RemoveDispatcherHostDrv(tntApiOpts.Tenant, tntApiOpts.ID)
				return nil
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
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

func TestGetDispatcherHostErr(t *testing.T) {
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
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	cfg.CacheCfg().Partitions[utils.CacheDispatcherHosts].Replicate = true
	cfg.DataDbCfg().RplFiltered = true
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheDispatcherHosts: {
			Limit:     3,
			Replicate: true,
		},
	}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetDispatcherHost: func(ctx *context.Context, args, reply any) error {
				return utils.ErrDSPHostNotFound
			},
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error {
				return errors.New("Can't Replicate")
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches):    clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
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

	if _, err := dm.GetDispatcherHost(dH.Tenant, dH.ID, true, true, utils.NonTransactional); err == nil || err != utils.ErrDSPHostNotFound {
		t.Error(err)
	}
	SetConnManager(connMgr)
	Cache = NewCacheS(cfg, dm, nil)

	if _, err := dm.GetDispatcherHost(dH.Tenant, dH.ID, true, true, utils.NonTransactional); err == nil {
		t.Error(err)
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
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1SetDispatcherHost: func(ctx *context.Context, args, reply any) error {
				chrgPrflApiOpts, cancast := args.(ChargerProfileWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().SetChargerProfileDrv(chrgPrflApiOpts.ChargerProfile)
				return nil
			},
			utils.ReplicatorSv1RemoveChargerProfile: func(ctx *context.Context, args, reply any) error {
				tntApiOpts, cancast := args.(utils.TenantIDWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().RemoveChargerProfileDrv(tntApiOpts.Tenant, tntApiOpts.ID)
				return nil
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
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
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1SetDispatcherProfile: func(ctx *context.Context, args, reply any) error {
				dspApiOpts, cancast := args.(DispatcherProfileWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().SetDispatcherProfileDrv(dspApiOpts.DispatcherProfile)
				return nil
			},
			utils.ReplicatorSv1RemoveDispatcherProfile: func(ctx *context.Context, args, reply any) error {
				tntApiOpts, cancast := args.(utils.TenantIDWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().RemoveDispatcherProfileDrv(tntApiOpts.Tenant, tntApiOpts.ID)
				return nil
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
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
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1SetRouteProfile: func(ctx *context.Context, args, reply any) error {
				routeApiOpts, cancast := args.(RouteProfileWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().SetRouteProfileDrv(routeApiOpts.RouteProfile)
				return nil
			},
			utils.ReplicatorSv1RemoveRouteProfile: func(ctx *context.Context, args, reply any) error {
				tntApiOpts, cancast := args.(utils.TenantIDWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().RemoveRouteProfileDrv(tntApiOpts.Tenant, tntApiOpts.ID)
				return nil
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
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
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1SetRatingPlan: func(ctx *context.Context, args, reply any) error {
				rPnApiOpts, cancast := args.(RatingPlanWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().SetRatingPlanDrv(rPnApiOpts.RatingPlan)
				return nil
			},
			utils.ReplicatorSv1RemoveRatingPlan: func(ctx *context.Context, args, reply any) error {
				strApiOpts, cancast := args.(utils.StringWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().RemoveRatingPlanDrv(strApiOpts.Arg)
				return nil
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
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
	tmpConn := connMgr
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetDataStorage(tmpDm)
		SetConnManager(tmpConn)
	}()
	Cache.Clear(nil)
	cfg.CacheCfg().Partitions[utils.CacheResources].Replicate = true
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheResources: {
			Limit:     3,
			Replicate: true,

			Remote: true,
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
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetResource: func(ctx *context.Context, args, reply any) error {
				tntApiOpts, cancast := args.(*utils.TenantIDWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().GetResourceDrv(tntApiOpts.Tenant, tntApiOpts.ID)
				*reply.(**Resource) = rS
				return nil
			},
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error {
				return errors.New("Can't Replicate")
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator): clientConn,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches):     clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	SetDataStorage(dm)

	if val, err := dm.GetResource(rS.Tenant, rS.ID, false, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rS, val) {
		t.Errorf("expected %v,received %v", utils.ToJSON(rS), utils.ToJSON(val))
	}
	Cache = NewCacheS(cfg, dm, nil)
	SetConnManager(connMgr)
	if _, err := dm.GetResource(rS.Tenant, rS.ID, false, true, utils.NonTransactional); err == nil || err.Error() != "Can't Replicate" {
		t.Error(err)
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
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetResourceProfile: func(ctx *context.Context, args, reply any) error {
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
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator): clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	SetDataStorage(dm)
	if val, err := dm.GetResourceProfile(rsP.Tenant, rsP.ID, false, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(val, rsP) {
		t.Errorf("expected %v,received %v", utils.ToJSON(rsP), utils.ToJSON(val))
	}
}

func TestGetActionTriggers(t *testing.T) {
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
	cfg.DataDbCfg().RplFiltered = true
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	cfg.CacheCfg().Partitions[utils.CacheActionTriggers].Replicate = true
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheActionTriggers: {
			Limit:     3,
			Replicate: true,
			APIKey:    "key",
			RouteID:   "route",
			Remote:    true,
		},
	}
	aT := ActionTriggers{
		&ActionTrigger{
			ID: "Test",
		},
	}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetActionTriggers: func(ctx *context.Context, args, reply any) error {
				strApiOpts, cancast := args.(*utils.StringWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().GetActionTriggersDrv(strApiOpts.Arg)
				*reply.(*ActionTriggers) = aT
				return nil
			},
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error {

				return errors.New("Can't Replicate")
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches):    clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	SetDataStorage(dm)
	Cache.Set(utils.CacheActionTriggers, "Test", ActionTriggers{}, []string{}, false, utils.NonTransactional)
	if val, err := dm.GetActionTriggers(aT[0].ID, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(val, aT) {
		t.Errorf("Expected %v,Received %v", utils.ToJSON(aT), utils.ToJSON(val))
	}
	SetConnManager(connMgr)
	Cache = NewCacheS(cfg, dm, nil)
	if _, err := dm.GetActionTriggers(aT[0].ID, false, utils.NonTransactional); err == nil {
		t.Error(err)
	}

}

func TestGetActionTriggersErr(t *testing.T) {
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
	cfg.DataDbCfg().RplFiltered = true
	cfg.CacheCfg().Partitions[utils.CacheActionTriggers].Replicate = true
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheActionTriggers: {
			Limit:     3,
			Replicate: true,
			APIKey:    "key",
			RouteID:   "route",
			Remote:    true,
		},
	}
	aT := ActionTriggers{
		&ActionTrigger{
			ID: "Test",
		},
	}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetActionTriggers: func(ctx *context.Context, args, reply any) error {
				return utils.ErrNotFound
			},
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error {
				return errors.New("Can't Replicate")
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr1 := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches):    clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr1)
	SetDataStorage(dm)
	if _, err := dm.GetActionTriggers(aT[0].ID, true, utils.NonTransactional); err == nil {
		t.Error(err)
	}
	Cache.Set(utils.CacheActionTriggers, "Test", nil, []string{}, false, utils.NonTransactional)
	if _, err := dm.GetActionTriggers(aT[0].ID, false, utils.NonTransactional); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	SetConnManager(connMgr1)
	Cache = NewCacheS(cfg, dm, nil)
	if _, err := dm.GetActionTriggers(aT[0].ID, true, utils.NonTransactional); err == nil || err.Error() != "Can't Replicate" {
		t.Error(err)
	}
	var dm2 *DataManager
	if _, err = dm2.GetActionTriggers("test", false, utils.NonTransactional); err == nil || err != utils.ErrNoDatabaseConn {
		t.Error(err)
	}
}
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
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetSharedGroup: func(ctx *context.Context, args, reply any) error {
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
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator): clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
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
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetStatQueueProfile: func(ctx *context.Context, args, reply any) error {
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
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator): clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
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
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1SetStatQueueProfile: func(ctx *context.Context, args, reply any) error {
				sqPApiOpts, cancast := args.(StatQueueProfileWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().SetStatQueueProfileDrv(sqPApiOpts.StatQueueProfile)
				return nil
			},
			utils.ReplicatorSv1SetStatQueue: func(ctx *context.Context, args, reply any) error {
				sqApiOpts, cancast := args.(StatQueueWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().SetStatQueueDrv(nil, sqApiOpts.StatQueue)
				return nil
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator): clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	SetDataStorage(dm)
	Cache.Set(utils.CacheStatQueueProfiles, utils.ConcatenatedKey(sqP.Tenant, sqP.ID), &StatQueueProfile{
		QueueLength: 2,
	}, []string{}, true, utils.NonTransactional)
	if err := dm.SetStatQueueProfile(sqP, false); err != nil {
		t.Error(err)
	}
}

func TestDMActionsRemote(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetDataStorage(tmpDm)
	}()
	Cache.Clear(nil)
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.DataDbCfg().RplFiltered = true
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheActions: {
			Limit:     3,
			Replicate: true,
			APIKey:    "key",
			RouteID:   "route",
			Remote:    true,
		},
	}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1SetActions: func(ctx *context.Context, args, reply any) error {
				sArgApiOpts, cancast := args.(SetActionsArgsWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().SetActionsDrv(sArgApiOpts.Key, sArgApiOpts.Acs)
				return nil
			},
			utils.ReplicatorSv1GetActions: func(ctx *context.Context, args, reply any) error {
				strApiOpts, cancast := args.(utils.StringWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().GetActionsDrv(strApiOpts.Arg)
				return nil
			},
			utils.ReplicatorSv1RemoveActions: func(ctx *context.Context, args, reply any) error {
				strApiOpts, cancast := args.(utils.StringWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().RemoveActionsDrv(strApiOpts.Arg)
				return nil
			},
		}}
	acs := Actions{{
		Id:               "SHARED",
		ActionType:       utils.MetaTopUp,
		ExpirationString: utils.MetaUnlimited}}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator): clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	SetDataStorage(dm)
	if err := dm.SetActions("KeyActions", acs); err != nil {
		t.Error(err)
	}
	Cache.Set(utils.CacheActions, "KeyActions", acs, []string{}, true, utils.NonTransactional)
	if val, err := dm.GetActions("KeyActions", false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(acs, val) {
		t.Errorf("expected  %+v,received %+v", utils.ToJSON(acs), utils.ToJSON(val))
	}
	if err := dm.RemoveActions("KeyActions"); err != nil {
		t.Error(err)
	}
	if _, has := db.db.Get(utils.CacheActions, "KeyActions"); has {
		t.Error("shouln't be in db cache")
	}
}

func TestGetDispatcherHost(t *testing.T) {
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
		utils.CacheDispatcherHosts: {
			Limit:   3,
			Remote:  true,
			APIKey:  "key",
			RouteID: "route",
		},
	}
	dh := &DispatcherHost{
		Tenant: "cgrates.org:HostID",
		RemoteHost: &config.RemoteHost{
			ID:        "Host1",
			Address:   "127.0.0.1:2012",
			TLS:       true,
			Transport: utils.MetaJSON,
		},
	}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetDispatcherHost: func(ctx *context.Context, args, reply any) error {
				tntApiOpts, cancast := args.(*utils.TenantIDWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().GetDispatcherHostDrv(tntApiOpts.Tenant, tntApiOpts.ID)
				*reply.(**DispatcherHost) = dh
				return nil
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	SetDataStorage(dm)

	if val, err := dm.GetDispatcherHost("cgrates.org", "HostID", false, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dh, val) {
		t.Errorf("expected %v,received %v", utils.ToJSON(val), utils.ToJSON(dh))
	}
}

func TestGetReverseDestinationRemote(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	tmpConn := connMgr
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		connMgr = tmpConn
		SetDataStorage(tmpDm)
	}()
	Cache.Clear(nil)
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.CacheCfg().Partitions[utils.CacheReverseDestinations].Replicate = true
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	cfg.DataDbCfg().RmtConnID = "rmt"
	cfg.GeneralCfg().NodeID = "node"
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheReverseDestinations: {
			Limit:   3,
			Remote:  true,
			APIKey:  "key",
			RouteID: "route",
		},
	}
	ids := []string{"dest1", "dest2"}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetReverseDestination: func(ctx *context.Context, args, reply any) error {
				strApiOpts, cancast := args.(*utils.StringWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().GetReverseDestinationDrv(strApiOpts.Arg, utils.NonTransactional)
				*reply.(*[]string) = ids
				return nil
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	SetDataStorage(dm)
	if val, err := dm.GetReverseDestination("CRUDReverseDestination", false, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(val, ids) {
		t.Errorf("expected %v,received %v", utils.ToJSON(ids), utils.ToJSON(val))
	}
	Cache = NewCacheS(cfg, dm, nil)
	clientConn2 := make(chan birpc.ClientConnector, 1)
	clientConn2 <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error {
				return errors.New("Can't replicate")
			},
		},
	}
	connMgr2 := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): clientConn2,
	})
	SetConnManager(connMgr2)
	if _, err := dm.GetReverseDestination("CRUDReverseDestination", false, true, utils.NonTransactional); err == nil || err.Error() != "Can't replicate" {
		t.Error(err)
	}
	var dm2 *DataManager
	if _, err := dm2.GetReverseDestination("CRUDReverseDestination", false, true, utils.NonTransactional); err == nil || err != utils.ErrNoDatabaseConn {
		t.Error(err)
	}
}

func TestDMRemoveDestination(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	tmpConn := connMgr
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetDataStorage(tmpDm)
		connMgr = tmpConn
	}()
	Cache.Clear(nil)
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().RplFiltered = true
	cfg.CacheCfg().Partitions[utils.CacheDestinations].Replicate = true
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheDestinations: {
			Limit:     3,
			Replicate: true,
			APIKey:    "key",
			RouteID:   "route",
		},
		utils.CacheReverseDestinations: {
			Remote: true,
		},
	}
	dest := &Destination{
		Id: "nat", Prefixes: []string{"0257", "0256", "0723"},
	}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1RemoveDestination: func(ctx *context.Context, args, reply any) error {
				strApiOpts, cancast := args.(utils.StringWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().RemoveDestinationDrv(strApiOpts.Arg, utils.NonTransactional)
				return nil
			},
			utils.ReplicatorSv1GetReverseDestination: func(ctx *context.Context, args, reply any) error {
				strApiOpts, cancast := args.(utils.StringWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().GetReverseDestinationDrv(strApiOpts.Arg, utils.NonTransactional)
				return nil
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	SetDataStorage(dm)
	dm.DataDB().SetDestinationDrv(dest, utils.NonTransactional)
	if err := dm.RemoveDestination(dest.Id, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if err := dm.RemoveDestination(dest.Id, utils.NonTransactional); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	Cache = NewCacheS(cfg, dm, nil)
	clientConn2 := make(chan birpc.ClientConnector, 1)
	clientConn2 <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1ReplicateRemove: func(ctx *context.Context, args, reply any) error {
				return errors.New("Can't replicate")
			},
		},
	}
	connMgr2 := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): clientConn2,
	})
	SetConnManager(connMgr2)
	if err := dm.RemoveDestination(dest.Id, utils.NonTransactional); err == nil {
		t.Error(err)
	}
	var dm2 *DataManager
	if err := dm2.RemoveDestination(dest.Id, utils.NonTransactional); err == nil {
		t.Error(err)
	}
}

func TestDMRemoveFilter(t *testing.T) {
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
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheFilters: {
			Limit:     3,
			Replicate: true,
			APIKey:    "key",
			RouteID:   "route",
			Remote:    false,
		},
		utils.CacheReverseFilterIndexes: {
			Remote: true,
		},
	}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1RemoveFilter: func(ctx *context.Context, args, reply any) error {
				tntApiOpts, cancast := args.(utils.TenantIDWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.dataDB.RemoveFilterDrv(tntApiOpts.TenantID.Tenant, tntApiOpts.TenantID.ID)
				return nil
			},
			utils.ReplicatorSv1GetIndexes: func(ctx *context.Context, args, reply any) error {
				return nil
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	SetDataStorage(dm)
	fltr := &Filter{
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
	dm.DataDB().SetFilterDrv(fltr)
	if err := dm.RemoveFilter(fltr.Tenant, fltr.ID, true); err == nil {
		t.Error(err)
	}
	if err := dm.RemoveFilter(fltr.Tenant, fltr.ID, false); err != nil {
		t.Error(err)
	}
	if err := dm.RemoveFilter(fltr.Tenant, fltr.ID, false); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	var dm2 *DataManager
	if err := dm2.RemoveFilter(fltr.Tenant, fltr.ID, false); err == nil || err != utils.ErrNoDatabaseConn {
		t.Error(err)
	}
}

func TestRemoveStatQueueProfile(t *testing.T) {
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
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().RplFiltered = true
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheStatQueueProfiles: {
			Limit:     3,
			Replicate: true,
			APIKey:    "key",
			RouteID:   "route",
		},
		utils.CacheStatQueues: {
			Limit:     3,
			Replicate: true,
			APIKey:    "key",
			RouteID:   "route",
		},
		utils.CacheReverseFilterIndexes: {
			Remote: true,
		},
	}
	sQ := &StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "STATS_RES_TEST12",
		FilterIDs: []string{"FLTR_ST_Resource1", "*string:~*req.Account:1001"},
		Weight:    50,
	}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1RemoveStatQueueProfile: func(ctx *context.Context, args, reply any) error {
				tntApiOpts, cancast := args.(utils.TenantIDWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().RemStatQueueProfileDrv(tntApiOpts.Tenant, tntApiOpts.ID)
				return nil
			},
			utils.ReplicatorSv1GetIndexes: func(ctx *context.Context, args, reply any) error {

				return errors.New("Can't replicate")
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	SetDataStorage(dm)
	dm.DataDB().SetStatQueueProfileDrv(sQ)
	if err = dm.RemoveStatQueueProfile(sQ.Tenant, sQ.ID, true); err == nil {
		t.Error(err)
	}
	dm.DataDB().SetStatQueueProfileDrv(sQ)
	if err = dm.RemoveStatQueueProfile(sQ.Tenant, sQ.ID, false); err != nil {
		t.Error(err)
	}
	if err = dm.RemoveStatQueueProfile(sQ.Tenant, sQ.ID, true); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	var dm2 *DataManager
	if err = dm2.RemoveStatQueueProfile(sQ.Tenant, sQ.ID, true); err == nil || err != utils.ErrNoDatabaseConn {
		t.Error(err)
	}
}

func TestDMGetTimingRemote(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	tmpConn := connMgr
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetConnManager(tmpConn)
		SetDataStorage(tmpDm)
	}()
	Cache.Clear(nil)
	cfg.CacheCfg().Partitions[utils.CacheTimings].Replicate = true
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	cfg.DataDbCfg().RmtConnID = "rmt"
	cfg.GeneralCfg().NodeID = "node"
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheTimings: {
			Limit:   3,
			Remote:  true,
			APIKey:  "key",
			RouteID: "route",
		},
	}
	tp := &utils.TPTiming{
		ID:        "MIDNIGHT",
		Years:     utils.Years{2020, 2019},
		Months:    utils.Months{1, 2, 3, 4},
		MonthDays: utils.MonthDays{5, 6, 7, 8},
		WeekDays:  utils.WeekDays{0, 1, 2, 3},
		StartTime: "00:00:00",
		EndTime:   "00:00:01",
	}

	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetTiming: func(ctx *context.Context, args, reply any) error {
				strApiOpts, cancast := args.(*utils.StringWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().GetTimingDrv(strApiOpts.Arg)
				*reply.(**utils.TPTiming) = tp
				return nil
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	SetDataStorage(dm)
	if _, err := dm.GetTiming(tp.ID, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	Cache = NewCacheS(cfg, dm, nil)
	clientConn2 := make(chan birpc.ClientConnector, 1)
	clientConn2 <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error {
				return errors.New("Can't replicate")
			},
		},
	}
	connMgr2 := NewConnManager(cfg, map[string]chan birpc.ClientConnector{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): clientConn2})
	SetConnManager(connMgr2)
	if _, err := dm.GetTiming(tp.ID, true, utils.NonTransactional); err == nil || err.Error() != "Can't replicate" {
		t.Error(err)
	}
	var dm2 *DataManager
	if _, err := dm2.GetTiming(tp.ID, true, utils.NonTransactional); err == nil || err != utils.ErrNoDatabaseConn {
		t.Error(err)
	}
}

func TestDmGetActions(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	tmpConn := connMgr
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetConnManager(tmpConn)
		SetDataStorage(tmpDm)
	}()
	Cache.Clear(nil)
	cfg.CacheCfg().Partitions[utils.CacheTimings].Replicate = true
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().RmtConnID = "rmt"
	cfg.GeneralCfg().NodeID = "node"
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheActions: {
			Limit:   3,
			Remote:  true,
			APIKey:  "key",
			RouteID: "route",
		},
	}
	acs := &Actions{
		{Id: "MINI",
			ActionType:       utils.MetaTopUpReset,
			ExpirationString: utils.MetaUnlimited},
	}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetActions: func(ctx *context.Context, args, reply any) error {
				strApiOpts, cancast := args.(*utils.StringWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().GetActionsDrv(strApiOpts.Arg)
				*reply.(*Actions) = *acs
				return nil
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	SetDataStorage(dm)
	if _, err := dm.GetActions("MINI", true, utils.NonTransactional); err != nil {
		t.Error(err)
	}
}

func TestDMSetLoadIDs(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	tmpConn := connMgr
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetConnManager(tmpConn)
		SetDataStorage(tmpDm)
	}()
	Cache.Clear(nil)
	cfg.CacheCfg().Partitions[utils.CacheTimings].Replicate = true
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheLoadIDs: {
			Limit:     3,
			Replicate: true,
			APIKey:    "key",
			RouteID:   "route",
		},
	}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1SetLoadIDs: func(ctx *context.Context, args, reply any) error {
				ldApiOpts, cancast := args.(*utils.LoadIDsWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().SetLoadIDsDrv(ldApiOpts.LoadIDs)
				return nil
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	SetDataStorage(dm)
	ld := map[string]int64{
		"load1": 23,
		"load2": 22,
	}
	if err := dm.SetLoadIDs(ld); err != nil {
		t.Error(err)
	}
	dm.dataDB = &DataDBMock{}
	if err = dm.SetLoadIDs(ld); err == nil || err != utils.ErrNotImplemented {
		t.Error(err)
	}
	var dm2 *DataManager
	if err = dm2.SetLoadIDs(ld); err == nil || err != utils.ErrNoDatabaseConn {
		t.Error(err)
	}
}

func TestGetItemLoadIDsRemote(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	tmpConn := connMgr
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetConnManager(tmpConn)
		SetDataStorage(tmpDm)
	}()
	Cache.Clear(nil)
	cfg.CacheCfg().Partitions[utils.CacheLoadIDs].Replicate = true
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	cfg.CacheCfg().Partitions[utils.CacheLoadIDs].Remote = true
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheLoadIDs: {
			Limit:   3,
			Remote:  true,
			APIKey:  "key",
			RouteID: "route",
		},
	}
	ld := map[string]int64{
		"load1": 23,
		"load2": 22,
	}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetItemLoadIDs: func(ctx *context.Context, args, reply any) error {
				strApiOpts, cancast := args.(*utils.StringWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				*reply.(*map[string]int64) = ld
				dm.DataDB().GetItemLoadIDsDrv(strApiOpts.Arg)
				return nil
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	SetDataStorage(dm)
	if val, err := dm.GetItemLoadIDs("load1", true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(val, ld) {
		t.Error(err)
	}
	for key := range ld {
		if _, has := Cache.Get(utils.CacheLoadIDs, key); !has {
			t.Error("Item isn't stored on the Cache")
		}
	}

	Cache = NewCacheS(cfg, dm, nil)
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error { return errors.New("Can't replicate") },
		},
	}
	connMgr2 := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): clientconn,
	})
	SetConnManager(connMgr2)
	if _, err := dm.GetItemLoadIDs("load1", true); err == nil {
		t.Error(err)
	}
}
func TestDMItemLoadIDsRemoteErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	tmpConn := connMgr
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetConnManager(tmpConn)
		SetDataStorage(tmpDm)
	}()
	Cache.Clear(nil)
	cfg.CacheCfg().Partitions[utils.CacheLoadIDs].Replicate = true
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	cfg.CacheCfg().Partitions[utils.CacheLoadIDs].Remote = true
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheLoadIDs: {
			Limit:   3,
			Remote:  true,
			APIKey:  "key",
			RouteID: "route",
		},
	}
	ld := map[string]int64{
		"load1": 23,
		"load2": 22,
	}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetItemLoadIDs: func(ctx *context.Context, args, reply any) error {
				*reply.(*map[string]int64) = ld
				return utils.ErrNotFound
			},
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error { return errors.New("Can't replicate") },
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches):    clientConn,
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})

	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	SetConnManager(connMgr)

	Cache = NewCacheS(cfg, dm, nil)
	if _, err := dm.GetItemLoadIDs("load1", true); err == nil || err.Error() != "Can't replicate" {
		t.Error(err)
	}
}

func TestActionPlanRemote(t *testing.T) {
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
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().RplFiltered = true
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheActionPlans: {
			Limit:     3,
			Replicate: true,
			APIKey:    "key",
			RouteID:   "route",
			Remote:    true,
		},
	}
	actPln := &ActionPlan{
		Id:         "TestActionPlansRemoveMember1",
		AccountIDs: utils.StringMap{"one": true},
		ActionTimings: []*ActionTiming{
			{
				Uuid: "uuid1",
				Timing: &RateInterval{
					Timing: &RITiming{
						Years:     utils.Years{2012},
						Months:    utils.Months{},
						MonthDays: utils.MonthDays{},
						WeekDays:  utils.WeekDays{},
						StartTime: utils.MetaASAP,
					},
				},
				Weight:    10,
				ActionsID: "MINI",
			},
		},
	}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1SetActionPlan: func(ctx *context.Context, args, reply any) error {
				setActPlnOpts, cancast := args.(*SetActionPlanArgWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.dataDB.SetActionPlanDrv(setActPlnOpts.Key, setActPlnOpts.Ats)
				return nil
			},
			utils.ReplicatorSv1RemoveActionPlan: func(ctx *context.Context, args, reply any) error {
				strApiOpts, cancast := args.(*utils.StringWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().RemoveActionPlanDrv(strApiOpts.Arg)
				return nil
			},
			utils.ReplicatorSv1GetAllActionPlans: func(ctx *context.Context, args, reply any) error {

				*reply.(*map[string]*ActionPlan) = map[string]*ActionPlan{
					"act_key": actPln,
				}
				return nil
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	SetDataStorage(dm)

	if err := dm.SetActionPlan("act_key", actPln, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if err := dm.RemoveActionPlan("act_key", utils.NonTransactional); err != nil {
		t.Error(err)
	}
	exp := map[string]*ActionPlan{
		"act_key": actPln,
	}
	if val, err := dm.GetAllActionPlans(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(val, exp) {
		t.Errorf("expected %v,received %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
}

func TestAccountActionPlansRemote(t *testing.T) {
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
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().RplFiltered = true
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheAccountActionPlans: {
			Limit:     3,
			Replicate: true,
			APIKey:    "key",
			RouteID:   "route",
			Remote:    true,
		},
	}

	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1SetAccountActionPlans: func(ctx *context.Context, args, reply any) error {
				setActPlnOpts, cancast := args.(*SetActionPlanArgWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.dataDB.SetActionPlanDrv(setActPlnOpts.Key, setActPlnOpts.Ats)
				return nil
			},
			utils.ReplicatorSv1RemAccountActionPlans: func(ctx *context.Context, args, reply any) error {

				return nil
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	SetDataStorage(dm)

	if err := dm.SetAccountActionPlans("acc_ID", []string{"act_pln", "act_pln"}, true); err != nil {
		t.Error(err)
	}
	if err = dm.RemAccountActionPlans("acc_ID", []string{}); err != nil {
		t.Error(err)
	}
}

func TestComputeIndexes(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	thd := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "THD_2",
		FilterIDs: []string{"*string:~*req.Account:1001"},

		MinHits: 0,

		Async: true,
	}
	dm.SetThresholdProfile(thd, false)
	transactionID := utils.GenUUID()
	if _, err := ComputeIndexes(dm, "cgrates.org", utils.EmptyString, utils.CacheThresholdFilterIndexes,
		nil, transactionID, func(tnt, id, grp string) (*[]string, error) {
			th, e := dm.GetThresholdProfile(tnt, id, true, false, utils.NonTransactional)
			if e != nil {
				return nil, e
			}
			return utils.SliceStringPointer(slices.Clone(th.FilterIDs)), nil
		}, nil); err != nil {
		t.Error(err)
	}
}
func TestUpdateFilterIndexRouteIndex(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	dataDB, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)

	oldFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Cost",
				Values:  []string{"unRegVal2"},
			},
		},
	}

	if err := oldFlt.Compile(); err != nil {
		t.Error(err)
	}

	if err := dm.SetFilter(oldFlt, true); err != nil {
		t.Error(err)
	}
	rp := &RouteProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr_test"},

		Sorting:           utils.MetaQOS,
		SortingParameters: []string{"param"},
		Routes: []*Route{{
			ID:         "RT1",
			FilterIDs:  []string{"fltr1"},
			AccountIDs: []string{"acc1"},

			ResourceIDs: []string{"res1"},
			StatIDs:     []string{"stat1"},

			RouteParameters: "params",
		}},
	}

	if err := dm.SetRouteProfile(rp, true); err != nil {
		t.Error(err)
	}

	expindx := map[string]utils.StringSet{
		"*string:*req.Cost:unRegVal2": {"ID": {}},
	}

	getindx, err := dm.GetIndexes(utils.CacheRouteFilterIndexes, cfg.GeneralCfg().DefaultTenant, utils.EmptyString, true, true)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expindx, getindx) {
		t.Errorf("Expected %v, Received %v", utils.ToJSON(expindx), utils.ToJSON(getindx))
	}

	newFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Usage",
				Values:  []string{"10s"},
			},
		},
	}
	if err := newFlt.Compile(); err != nil {
		t.Error(err)
	}
	if err := dm.SetFilter(newFlt, false); err != nil {
		t.Error(err)
	}

	if err := UpdateFilterIndex(dm, oldFlt, newFlt); err != nil {
		t.Error(err)
	}

	expindxNew := map[string]utils.StringSet{
		"*prefix:*req.Usage:10s": {"ID": {}},
	}
	getindxNew, err := dm.GetIndexes(utils.CacheRouteFilterIndexes, cfg.GeneralCfg().DefaultTenant, utils.EmptyString, true, true)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expindxNew, getindxNew) {
		t.Errorf("Expected %v, Received %v", utils.ToJSON(expindxNew), utils.ToJSON(getindxNew))
	}
}

func TestUpdateFilterIndexStatIndex(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	dataDB, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)

	oldFlt := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_STATS_1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Stats",
				Values:  []string{"StatQueueProfile1"},
			},
		},
	}
	if err := oldFlt.Compile(); err != nil {
		t.Error(err)
	}
	if err := dm.SetFilter(oldFlt, true); err != nil {
		t.Error(err)
	}
	sQ := &StatQueueProfile{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "TEST_PROFILE1",
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2020, 4, 18, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2020, 4, 18, 14, 25, 0, 0, time.UTC),
		},
		QueueLength: 10,
		TTL:         10 * time.Second,
		Metrics: []*MetricWithFilters{
			{
				MetricID: "*sum",
			},
			{
				MetricID: "*acd",
			},
		},
		ThresholdIDs: []string{"Val1", "Val2"},
		Blocker:      true,
		Stored:       true,
		Weight:       20,
		MinItems:     1,
		FilterIDs:    []string{"FLTR_STATS_1"},
	}

	if err := dm.SetStatQueueProfile(sQ, true); err != nil {
		t.Error(err)
	}

	expindx := map[string]utils.StringSet{
		"*string:*req.Stats:StatQueueProfile1": {"TEST_PROFILE1": {}},
	}

	if getindx, err := dm.GetIndexes(utils.CacheStatFilterIndexes, cfg.GeneralCfg().DefaultTenant, utils.EmptyString, true, true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expindx, getindx) {
		t.Errorf("Expected %v, Received %v", utils.ToJSON(expindx), utils.ToJSON(getindx))
	}
	newFlt := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_STATS_1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Usage",
				Values:  []string{"10s"},
			},
		},
	}
	if err := newFlt.Compile(); err != nil {
		t.Error(err)
	}
	if err := dm.SetFilter(newFlt, false); err != nil {
		t.Error(err)
	}
	if err := UpdateFilterIndex(dm, oldFlt, newFlt); err != nil {
		t.Error(err)
	}
	expindxNew := map[string]utils.StringSet{
		"*prefix:*req.Usage:10s": {"TEST_PROFILE1": {}},
	}
	if getindx, err := dm.GetIndexes(utils.CacheStatFilterIndexes, cfg.GeneralCfg().DefaultTenant, utils.EmptyString, true, true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expindxNew, getindx) {
		t.Errorf("Expected %v, Received %v", utils.ToJSON(expindx), utils.ToJSON(getindx))
	}

}

func TestDMRatingProfile(t *testing.T) {
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
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().RplFiltered = true
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheRatingProfiles: {
			Limit:     3,
			Replicate: true,
			APIKey:    "key",
			RouteID:   "route",
			Remote:    true,
		},
	}
	rpf := &RatingProfile{
		Id: "*out:TCDDBSWF:call:*any",
		RatingPlanActivations: RatingPlanActivations{&RatingPlanActivation{
			ActivationTime: time.Date(2015, 01, 01, 8, 0, 0, 0, time.UTC),
			RatingPlanId:   "RP_ANY2CNT",
		}},
	}

	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1SetRatingProfile: func(ctx *context.Context, args, reply any) error {
				rtPrfApiOpts, cancast := args.(*RatingProfileWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.dataDB.SetRatingProfileDrv(rtPrfApiOpts.RatingProfile)
				return nil
			},
			utils.ReplicatorSv1RemoveRatingProfile: func(ctx *context.Context, args, reply any) error {
				strApiOpts, cancast := args.(*utils.StringWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().RemoveRatingProfileDrv(strApiOpts.Arg)
				return nil
			},
			utils.ReplicatorSv1GetRatingProfile: func(ctx *context.Context, args, reply any) error {
				strApiOpts, cancast := args.(*utils.StringWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().GetRatingProfileDrv(strApiOpts.Arg)
				*reply.(**RatingProfile) = rpf
				return nil
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	SetDataStorage(dm)
	if err := dm.SetRatingProfile(rpf); err != nil {
		t.Error(err)
	}
	if err := dm.RemoveRatingProfile(rpf.Id); err != nil {
		t.Error(err)
	}
	if _, err := dm.GetRatingProfile(rpf.Id, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}
}

func TestUpdateFilterDispatcherIndex(t *testing.T) {
	tmp := Cache
	tmpDm := dm
	defer func() {
		Cache = tmp
		dm = tmpDm
		config.SetCgrConfig(config.NewDefaultCGRConfig())
	}()
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	dataDB, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	oldFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "DISPATCHER_FLTR1",
		Rules:  []*FilterRule{{Type: utils.MetaString, Element: "~*req.Destination", Values: []string{"ACC1", "ACC2", "~*req.Account"}}},
	}
	if err := oldFlt.Compile(); err != nil {
		t.Error(err)
	}
	if err := dm.SetFilter(oldFlt, true); err != nil {
		t.Error(err)
	}
	disp := &DispatcherProfile{
		Tenant:     "cgrates.org",
		ID:         "Dsp",
		Subsystems: []string{"*any"},
		FilterIDs:  []string{"DISPATCHER_FLTR1"},
		Strategy:   utils.MetaFirst,
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		StrategyParams: map[string]any{},
		Weight:         20,
	}
	if err := dm.SetDispatcherProfile(disp, true); err != nil {
		t.Error(err)
	}

	exp := map[string]utils.StringSet{
		"*string:*req.Destination:ACC1": {"Dsp": {}},
		"*string:*req.Destination:ACC2": {"Dsp": {}},
	}
	if indx, err := dm.GetIndexes(utils.CacheDispatcherFilterIndexes, "cgrates.org:*any", utils.EmptyString, true, true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, indx) {
		t.Errorf("Expected %v, Received %v", utils.ToJSON(exp), utils.ToJSON(indx))
	}
	newFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "DISPATCHER_FLTR1",
		Rules: []*FilterRule{{
			Type:    utils.MetaString,
			Element: "~*req.Charger",
			Values:  []string{"ChargerProfile2"}}},
	}
	if err := newFlt.Compile(); err != nil {
		t.Error(err)
	}
	if err := dm.SetFilter(newFlt, true); err != nil {
		t.Error(err)
	}
	exp = map[string]utils.StringSet{
		"*string:*req.Charger:ChargerProfile2": {"Dsp": {}},
	}
	if indx, err := dm.GetIndexes(utils.CacheDispatcherFilterIndexes, "cgrates.org:*any", utils.EmptyString, true, true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, indx) {
		t.Errorf("Expected %v, Received %v", utils.ToJSON(exp), utils.ToJSON(indx))
	}
}

func TestDMGetRatingPlan(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	tmpConn := connMgr
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetConnManager(tmpConn)
		SetDataStorage(tmpDm)
	}()
	Cache.Clear(nil)
	cfg.CacheCfg().Partitions[utils.CacheLoadIDs].Replicate = true
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	cfg.CacheCfg().Partitions[utils.CacheRatingPlans].Remote = true
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheRatingPlans: {
			Limit:     3,
			Remote:    true,
			APIKey:    "key",
			RouteID:   "route",
			Replicate: true,
		},
	}
	rpL := &RatingPlan{
		Id: "id",
		DestinationRates: map[string]RPRateList{
			"DestinationRates": {&RPRate{Rating: "Rating"}}},
		Ratings: map[string]*RIRate{"Ratings": {ConnectFee: 0.7}},
		Timings: map[string]*RITiming{"Timings": {Months: utils.Months{4}}},
	}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetRatingPlan: func(ctx *context.Context, args, reply any) error {
				*reply.(**RatingPlan) = rpL
				return nil
			},
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error {

				return errors.New("Can't replicate ")
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches):    clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	SetDataStorage(dm)
	if _, err := dm.GetRatingPlan("id", true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

}

func TestDMChargerProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	tmpConn := connMgr
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetConnManager(tmpConn)
		SetDataStorage(tmpDm)
	}()
	Cache.Clear(nil)
	cfg.CacheCfg().Partitions[utils.CacheLoadIDs].Replicate = true
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	cfg.CacheCfg().Partitions[utils.CacheRatingPlans].Remote = true
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheChargerProfiles: {
			Limit:     3,
			Remote:    true,
			APIKey:    "key",
			RouteID:   "route",
			Replicate: true,
		},
	}
	chP := &ChargerProfile{
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
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetChargerProfile: func(ctx *context.Context, args, reply any) error {
				*reply.(**ChargerProfile) = chP
				return nil
			},
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error {

				return errors.New("Can't replicate ")
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches):    clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	SetDataStorage(dm)
	if _, err := dm.GetChargerProfile(chP.Tenant, chP.ID, false, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}
}

func TestDMDispatcherProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	tmpConn := connMgr
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetConnManager(tmpConn)
		SetDataStorage(tmpDm)
	}()
	Cache.Clear(nil)
	cfg.CacheCfg().Partitions[utils.CacheLoadIDs].Replicate = true
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	cfg.CacheCfg().Partitions[utils.CacheRatingPlans].Remote = true
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheDispatcherProfiles: {
			Limit:     3,
			Remote:    true,
			APIKey:    "key",
			RouteID:   "route",
			Replicate: true,
		},
	}
	dPP := &DispatcherProfile{
		Tenant:     cfg.GeneralCfg().DefaultTenant,
		ID:         "Dsp",
		Subsystems: []string{"*any"},
		FilterIDs:  []string{"DISPATCHER_FLTR1"},
		Strategy:   utils.MetaFirst,
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		StrategyParams: map[string]any{},
		Weight:         20,
		Hosts: DispatcherHostProfiles{
			&DispatcherHostProfile{
				ID:        "C1",
				FilterIDs: []string{},
				Weight:    10,
				Params:    map[string]any{"0": "192.168.54.203", utils.MetaRatio: "2"},
				Blocker:   false,
			},
		},
	}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetDispatcherProfile: func(ctx *context.Context, args, reply any) error {
				*reply.(**DispatcherProfile) = dPP
				return nil
			},
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error {

				return errors.New("Can't replicate ")
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches):    clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	SetDataStorage(dm)
	if _, err := dm.GetDispatcherProfile(dPP.Tenant, dPP.ID, false, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

}

func TestCacheDataFromDB(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetDataStorage(tmpDm)
	}()
	Cache.Clear(nil)
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	if err := dm.CacheDataFromDB("INVALID", nil, false); err == nil || err.Error() != utils.UnsupportedCachePrefix {
		t.Error(err)
	}
	rp := &RatingPlan{
		Id: "id",
		DestinationRates: map[string]RPRateList{
			"DestinationRates": {&RPRate{Rating: "Rating"}}},
		Ratings: map[string]*RIRate{"Ratings": {ConnectFee: 0.7}},
		Timings: map[string]*RITiming{"Timings": {Months: utils.Months{4}}},
	}
	if err := dm.SetRatingPlan(rp); err != nil {
		t.Error(err)
	}
	if _, hasIt := Cache.Get(utils.CacheRatingPlans, rp.Id); hasIt {
		t.Error("Already in cache")
	}
	if err := dm.CacheDataFromDB(utils.RatingPlanPrefix, []string{rp.Id}, false); err != nil {
		t.Error(err)
	}
	rP := &RatingProfile{
		Id: "*out:TCDDBSWF:call:*any",
		RatingPlanActivations: RatingPlanActivations{&RatingPlanActivation{
			ActivationTime: time.Date(2015, 01, 01, 8, 0, 0, 0, time.UTC),
			RatingPlanId:   rp.Id,
		}},
	}
	if err := dm.SetRatingProfile(rP); err != nil {
		t.Error(err)
	}
	if _, hasIt := Cache.Get(utils.CacheRatingProfiles, rP.Id); hasIt {
		t.Error("Already in cache")
	}
	if err := dm.CacheDataFromDB(utils.RatingProfilePrefix, []string{rP.Id}, false); err != nil {
		t.Error(err)
	}
	as := Actions{
		&Action{
			ActionType: utils.MetaSetBalance,
			Filters:    []string{"*string:~*req.BalanceMap.*monetary[0].ID:*default", "*lt:~*req.BalanceMap.*monetary[0].Value:0"},
			Balance: &BalanceFilter{
				Type:     utils.StringPointer("*sms"),
				ID:       utils.StringPointer("for_v3hsillmilld500m_sms_ill"),
				Disabled: utils.BoolPointer(true),
			},
			Weight: 9,
		},
	}
	if err := dm.SetActions("test", as); err != nil {
		t.Error(err)
	}
	if _, hasIt := Cache.Get(utils.CacheActions, "test"); hasIt {
		t.Error("Already in cache")
	}
	if err := dm.CacheDataFromDB(utils.ActionPrefix, []string{"test"}, false); err != nil {
		t.Error(err)
	}
	rsPrf := &ResourceProfile{
		Tenant: "tenant.custom",
		ID:     "RES_GRP1",
		FilterIDs: []string{
			"*string:~*req.RequestType:*rated",
		},
		UsageTTL:          10 * time.Microsecond,
		Limit:             10,
		AllocationMessage: "MessageAllocation",
		Blocker:           true,
		Stored:            true,
		Weight:            20,
	}
	if err := dm.SetResourceProfile(rsPrf, false); err != nil {
		t.Error(err)
	}
	if _, hasIt := Cache.Get(utils.CacheResourceProfiles, "test"); hasIt {
		t.Error("Already in cache")
	}
	if err := dm.CacheDataFromDB(utils.ResourceProfilesPrefix, []string{utils.ConcatenatedKey(rsPrf.Tenant, rsPrf.ID)}, false); err != nil {
		t.Error(err)
	}
	ru1 := &ResourceUsage{
		Tenant:     "cgrates.org",
		ID:         "RU1",
		ExpiryTime: time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
		Units:      1,
	}
	rs := &Resource{
		Tenant: "cgrates.org",
		ID:     "RES_GRP1",
		rPrf:   rsPrf,
		TTLIdx: []string{ru1.ID},
		tUsage: utils.Float64Pointer(2),
	}
	if err = dm.SetResource(rs); err != nil {
		t.Error(err)
	}
	if _, hasIt := Cache.Get(utils.CacheResources, utils.ConcatenatedKey(rs.Tenant, rs.ID)); hasIt {
		t.Error("Already in cache")
	}
	if err := dm.CacheDataFromDB(utils.ResourcesPrefix, []string{utils.ConcatenatedKey(rs.Tenant, rs.ID)}, false); err != nil {
		t.Error(err)
	}
	if err := dm.CacheDataFromDB(utils.ResourceFilterIndexes, []string{"*resources:*string:~*req.RequestType:*rated"}, false); err != nil {
		t.Error(err)
	}
	sqPrf := &StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "SQ1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weight:    50,
	}
	if err = dm.SetStatQueueProfile(sqPrf, false); err != nil {
		t.Error(err)
	}
	if _, hasIt := Cache.Get(utils.CacheStatQueueProfiles, utils.ConcatenatedKey(sqPrf.Tenant, sqPrf.ID)); hasIt {
		t.Error("Already in cache")
	}
	if err := dm.CacheDataFromDB(utils.StatQueueProfilePrefix, []string{utils.ConcatenatedKey(sqPrf.Tenant, sqPrf.ID)}, false); err != nil {
		t.Error(err)
	}
	if err := dm.CacheDataFromDB(utils.StatFilterIndexes, []string{"*statqueue_profiles:*string:~*req.Account:1001"}, false); err != nil {
		t.Error(err)
	}
	sq := &StatQueue{
		sqPrfl: sqPrf,
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: map[string]StatMetric{
			utils.MetaTCD: &StatTCD{},
		},
	}
	if err = dm.SetStatQueue(sq); err != nil {
		t.Error(err)
	}
	if _, hasIt := Cache.Get(utils.CacheStatQueues, utils.ConcatenatedKey(sq.Tenant, sq.ID)); hasIt {
		t.Error("Already in cache")
	}
	if err := dm.CacheDataFromDB(utils.StatQueuePrefix, []string{utils.ConcatenatedKey(sq.Tenant, sq.ID)}, false); err != nil {
		t.Error(err)
	}
	thdPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001", "*notstring:~*req.Destination:+49123"},
	}
	if err = dm.SetThresholdProfile(thdPrf, false); err != nil {
		t.Error(err)
	}
	if _, hasIt := Cache.Get(utils.CacheThresholdProfiles, utils.ConcatenatedKey(thdPrf.Tenant, thdPrf.ID)); hasIt {
		t.Error("Already in cache")
	}
	if err := dm.CacheDataFromDB(utils.ThresholdProfilePrefix, []string{utils.ConcatenatedKey(thdPrf.Tenant, thdPrf.ID)}, false); err != nil {
		t.Error(err)
	}
	if err := dm.CacheDataFromDB(utils.ThresholdFilterIndexes, []string{"*threshold_profiles:*string:~*req.Account:1001"}, false); err != nil {
		t.Error(err)
	}
	thd := &Threshold{
		Tenant: "cgrates.org",
		ID:     "TH1",
		tPrfl:  thdPrf,
	}
	if err = dm.SetThreshold(thd); err != nil {
		t.Error(err)
	}
	if _, hasIt := Cache.Get(utils.CacheThresholds, utils.ConcatenatedKey(thd.Tenant, thd.ID)); hasIt {
		t.Error("Already in cache")
	}
	if err := dm.CacheDataFromDB(utils.ThresholdPrefix, []string{utils.ConcatenatedKey(thd.Tenant, thd.ID)}, false); err != nil {
		t.Error(err)
	}
	rPrf := &RouteProfile{
		Tenant:            "cgrates.org",
		ID:                "SUP1",
		FilterIDs:         []string{"*string:~*opts.Account:1001"},
		Weight:            10,
		Sorting:           utils.MetaQOS,
		SortingParameters: []string{},
		Routes: []*Route{{
			ID:            "Sup",
			FilterIDs:     []string{},
			AccountIDs:    []string{"1001"},
			RatingPlanIDs: []string{"RT_PLAN1"},
			ResourceIDs:   []string{"RES1"},
			Weight:        10,
		}},
	}
	if err = dm.SetRouteProfile(rPrf, false); err != nil {
		t.Error(err)
	}
	if _, hasIt := Cache.Get(utils.CacheRouteProfiles, rPrf.TenantID()); hasIt {
		t.Error("Already in cache")
	}
	if err := dm.CacheDataFromDB(utils.RouteProfilePrefix, []string{rPrf.TenantID()}, false); err != nil {
		t.Error(err)
	}
	if err := dm.CacheDataFromDB(utils.RouteFilterIndexes, []string{"*route_profiles:*string:~*opts.Account:1001"}, false); err != nil {
		t.Error(err)
	}
	attrPrf := &AttributeProfile{
		Tenant:             "cgrates.org",
		ID:                 "1001",
		Contexts:           []string{utils.MetaAny},
		FilterIDs:          []string{"*string:~*req.Account:1002"},
		ActivationInterval: nil,
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Subject",
				Type:  utils.MetaVariable,
				Value: config.NewRSRParsersMustCompile("call_1001", utils.InfieldSep),
			},
		},
		Blocker: false,
		Weight:  10,
	}
	if err = dm.SetAttributeProfile(attrPrf, false); err != nil {
		t.Error(err)
	}
	if _, hasIt := Cache.Get(utils.CacheAttributeProfiles, attrPrf.TenantID()); hasIt {
		t.Error("Already in cache")
	}
	if err := dm.CacheDataFromDB(utils.AttributeProfilePrefix, []string{attrPrf.TenantID()}, false); err != nil {
		t.Error(err)
	}
	if err := dm.CacheDataFromDB(utils.AttributeFilterIndexes, []string{"*attribute_profiles:*string:~*req.Account:1002"}, false); err != nil {
		t.Error(err)
	}
	chPrf := &ChargerProfile{
		Tenant:    "cgrates.com",
		ID:        "CHRG_1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		AttributeIDs: []string{"ATTR_1"},
		Weight:       20,
	}
	if err = dm.SetChargerProfile(chPrf, false); err != nil {
		t.Error(err)
	}
	if _, hasIt := Cache.Get(utils.CacheChargerProfiles, chPrf.TenantID()); hasIt {
		t.Error("Already in cache")
	}
	if err := dm.CacheDataFromDB(utils.ChargerProfilePrefix, []string{chPrf.TenantID()}, false); err != nil {
		t.Error(err)
	}
	if err := dm.CacheDataFromDB(utils.ChargerFilterIndexes, []string{"*charger_profiles:*string:~*req.Account:1001"}, false); err != nil {
		t.Error(err)
	}
	dsPrf := &DispatcherProfile{
		Tenant:    "cgrates.org",
		ID:        "Dsp1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Strategy: utils.MetaRandom,
		Weight:   20,
	}
	if err = dm.SetDispatcherProfile(dsPrf, false); err != nil {
		t.Error(err)
	}
	if _, hasIt := Cache.Get(utils.CacheDispatcherProfiles, dsPrf.TenantID()); hasIt {
		t.Error("Already in cache")
	}
	if err := dm.CacheDataFromDB(utils.DispatcherProfilePrefix, []string{dsPrf.TenantID()}, false); err != nil {
		t.Error(err)
	}
	if err := dm.CacheDataFromDB(utils.DispatcherFilterIndexes, []string{"*dispatcher_profiles:*string:~*req.Account:1001"}, false); err != nil {
		t.Error(err)
	}
	dH := &DispatcherHost{
		Tenant: "cgrates.org",
		RemoteHost: &config.RemoteHost{
			ID:        "Host1",
			Address:   "127.0.0.1:2012",
			Transport: utils.MetaJSON,
		},
	}
	if err = dm.SetDispatcherHost(dH); err != nil {
		t.Error(err)
	}
	if _, hasIt := Cache.Get(utils.CacheDispatcherHosts, dH.TenantID()); hasIt {
		t.Error("Already in cache")
	}
	if err := dm.CacheDataFromDB(utils.DispatcherHostPrefix, []string{dH.TenantID()}, false); err != nil {
		t.Error(err)
	}
	if err := dm.CacheDataFromDB(utils.FilterIndexPrfx, []string{"*string:~*req.Account:1001:Dsp1"}, false); err != nil {
		t.Error(err)
	}
}
func TestCacheDataFromDBErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	tmpConn := connMgr
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetDataStorage(tmpDm)
		SetConnManager(tmpConn)
	}()
	Cache.Clear(nil)
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.MetaThresholdProfiles: {
			Remote: true,
		},
	}
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.CacheCfg().Partitions[utils.CacheThresholdProfiles].Replicate = true
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetThresholdProfile: func(ctx *context.Context, args, reply any) error {
				return errors.New("Another Error")
			},
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error {
				return fmt.Errorf("New Error")
			},
		},
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator): clientConn,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches):     clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	Cache = NewCacheS(cfg, dm, nil)
	SetConnManager(connMgr)
	thdPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001", "*notstring:~*req.Destination:+49123"},
	}

	if err := dm.CacheDataFromDB(utils.ThresholdProfilePrefix, []string{utils.ConcatenatedKey(thdPrf.Tenant, thdPrf.ID)}, false); err == nil {
		t.Error(err)
	}
}

func TestDMGetRouteProfile(t *testing.T) {
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
	cfg.CacheCfg().Partitions[utils.CacheRouteProfiles].Replicate = true
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheRouteProfiles: {
			Limit:     3,
			Remote:    true,
			APIKey:    "key",
			RouteID:   "route",
			Replicate: true,
		},
	}
	rpL := &RouteProfile{Tenant: "cgrates.org",
		ID:        "ROUTE_ACNT_1002",
		FilterIDs: []string{"FLTR_ACNT_1002"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2017, 11, 27, 00, 00, 00, 00, time.UTC),
		},
		Sorting: utils.MetaLC,
		Routes: []*Route{
			{
				ID:            "route1",
				RatingPlanIDs: []string{"RP_1002_LOW"},
				Weight:        10,
				Blocker:       false,
			},
		},
		Weight: 10,
	}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetRouteProfile: func(ctx *context.Context, args, reply any) error {
				*reply.(**RouteProfile) = rpL
				return nil
			},
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error {
				return errors.New("Can't Replicate")
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches):    clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	SetDataStorage(dm)
	if val, err := dm.GetRouteProfile(rpL.Tenant, rpL.ID, false, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(val, rpL) {
		t.Errorf("expected %v,received %v", utils.ToJSON(rpL), utils.ToJSON(val))
	}
	Cache = NewCacheS(cfg, dm, nil)
	SetConnManager(connMgr)
	if _, err := dm.GetRouteProfile(rpL.Tenant, rpL.ID, false, true, utils.NonTransactional); err == nil {
		t.Error(err)
	}
}
func TestDMGetRouteProfileErr(t *testing.T) {
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
	cfg.CacheCfg().Partitions[utils.CacheRouteProfiles].Replicate = true
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheRouteProfiles: {
			Limit:     3,
			Remote:    true,
			APIKey:    "key",
			RouteID:   "route",
			Replicate: true,
		},
	}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetRouteProfile: func(ctx *context.Context, args, reply any) error {
				return utils.ErrNotFound
			},
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error {
				return errors.New("Can't Replicate")
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches):    clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	SetDataStorage(dm)
	Cache = NewCacheS(cfg, dm, nil)
	SetConnManager(connMgr)
	if _, err := dm.GetRouteProfile("cgrates.org", "id", false, true, utils.NonTransactional); err == nil || err.Error() != "Can't Replicate" {
		t.Error(err)
	}
	Cache.Set(utils.CacheRouteProfiles, "cgrates.org:id", nil, []string{}, false, utils.NonTransactional)
	if _, err := dm.GetRouteProfile("cgrates.org", "id", true, true, utils.NonTransactional); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	var dm2 *DataManager
	if _, err := dm2.GetRouteProfile("cgrates.org", "id", false, true, utils.NonTransactional); err == nil || err != utils.ErrNoDatabaseConn {
		t.Error(err)
	}

}
func TestUpdateFilterIndexStatErr1(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	dataDB, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	dm.dataDB = &DataDBMock{
		GetIndexesDrvF: func(idxItmType, tntCtx string, idxKey ...string) (indexes map[string]utils.StringSet, err error) {
			return map[string]utils.StringSet{
				utils.CacheStatFilterIndexes: {
					"ATTR_TEST": {},
				},
			}, nil
		},
	}

	oldFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Cost",
				Values:  []string{"unRegVal2"},
			},
		},
	}
	newFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Usage",
				Values:  []string{"10s"},
			},
		},
	}
	if err := UpdateFilterIndex(dm, oldFlt, newFlt); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotImplemented, err)
	}
}

func TestUpdateFilterIndexRemoveThresholdErr1(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	dataDB, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	dm.dataDB = &DataDBMock{
		GetIndexesDrvF: func(idxItmType, tntCtx string, idxKey ...string) (indexes map[string]utils.StringSet, err error) {
			return map[string]utils.StringSet{
				utils.CacheThresholdFilterIndexes: {
					"ATTR_TEST": {},
				},
			}, nil
		},
	}

	oldFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Cost",
				Values:  []string{"unRegVal2"},
			},
		},
	}
	newFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr_test",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Usage",
				Values:  []string{"10s"},
			},
		},
	}
	if err := UpdateFilterIndex(dm, oldFlt, newFlt); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotImplemented, err)
	}
}

func TestDMAttributeProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmp := Cache
	tmpConn := connMgr
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
		SetConnManager(tmpConn)
		SetDataStorage(tmpDm)
	}()
	Cache.Clear(nil)
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheAttributeProfiles: {
			Limit:     3,
			Remote:    true,
			APIKey:    "key",
			RouteID:   "route",
			Replicate: true,
		},
	}
	attrPrf := &AttributeProfile{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,

		ID:        "ATTR_1001_SIMPLEAUTH",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Contexts:  []string{"simpleauth"},
		Attributes: []*Attribute{
			{
				FilterIDs: []string{},
				Path:      utils.MetaReq + utils.NestingSep + "Password",
				Type:      utils.MetaConstant,
				Value:     config.NewRSRParsersMustCompile("CGRateS.org", utils.InfieldSep),
			},
		},
		Weight: 20.0,
	}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetAttributeProfile: func(ctx *context.Context, args, reply any) error {
				*reply.(**AttributeProfile) = attrPrf
				return nil
			},
			utils.ReplicatorSv1SetAttributeProfile: func(ctx *context.Context, args, reply any) error {
				attrPrfApiOpts, cancast := args.(*AttributeProfileWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().SetAttributeProfileDrv(attrPrfApiOpts.AttributeProfile)
				return nil
			},
			utils.ReplicatorSv1RemoveAttributeProfile: func(ctx *context.Context, args, reply any) error {
				tntApiOpts, cancast := args.(*utils.TenantIDWithAPIOpts)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().RemoveAttributeProfileDrv(tntApiOpts.Tenant, tntApiOpts.ID)
				return nil
			},
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	})
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	SetDataStorage(dm)
	if err := dm.SetAttributeProfile(attrPrf, false); err != nil {
		t.Error(err)
	}
	if err := dm.RemoveAttributeProfile(attrPrf.Tenant, attrPrf.ID, false); err != nil {
		t.Error(err)
	}
	if _, has := db.db.Get(utils.CacheAttributeProfiles, utils.ConcatenatedKey(attrPrf.Tenant, attrPrf.ID)); has {
		t.Error("Should been removed from db")
	}
}

func TestUpdateFilterResourceIndexErr(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	dataDB, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	dm.dataDB = &DataDBMock{
		GetIndexesDrvF: func(idxItmType, tntCtx string, idxKey ...string) (indexes map[string]utils.StringSet, err error) {
			return map[string]utils.StringSet{
				utils.CacheResourceFilterIndexes: {
					"ATTR_TEST": {},
				},
			}, nil
		},
	}
	oldFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "RSC_FLTR1",
		Rules:  []*FilterRule{{Type: utils.MetaString, Element: "~*req.Destination", Values: []string{"ACC1", "ACC2", "~*req.Account"}}},
	}
	newFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "RSC_FLTR1",
		Rules: []*FilterRule{
			{
				Element: "~*req.Account",
				Type:    utils.MetaString,
				Values:  []string{"1001", "1002"},
			},
		}}
	if err := UpdateFilterIndex(dm, oldFlt, newFlt); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotImplemented, err)
	}
}
func TestUpdateFilterRouteIndexErr(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	dataDB, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	dm.dataDB = &DataDBMock{
		GetIndexesDrvF: func(idxItmType, tntCtx string, idxKey ...string) (indexes map[string]utils.StringSet, err error) {
			return map[string]utils.StringSet{
				utils.CacheRouteFilterIndexes: {
					"ATTR_TEST": {},
				},
			}, nil
		},
	}
	oldFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_SUPP_2",
		Rules: []*FilterRule{{
			Type:    utils.MetaString,
			Element: "~*req.Route",
			Values:  []string{"RouteProfile2"},
		}},
	}
	newFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_SUPP_2",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: "~*req.PddInterval",
				Values:  []string{(time.Second).String()},
			},
		}}
	if err := UpdateFilterIndex(dm, oldFlt, newFlt); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotImplemented, err)
	}
}

func TestUpdateFilterChargersIndexErr(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	dataDB, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	dm.dataDB = &DataDBMock{
		GetIndexesDrvF: func(idxItmType, tntCtx string, idxKey ...string) (indexes map[string]utils.StringSet, err error) {
			return map[string]utils.StringSet{
				utils.CacheChargerFilterIndexes: {
					"ATTR_TEST": {},
				},
			}, nil
		},
	}
	oldFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_CP_2",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Charger",
				Values:  []string{"ChargerProfile2"},
			},
		}}

	newFlt := &Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_CP_2",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"200.00"},
			},
		}}
	if err := UpdateFilterIndex(dm, oldFlt, newFlt); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotImplemented, err)
	}
}

func TestDmIndexes(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
	}()
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheResourceFilterIndexes: {
			Replicate: true,
		},
	}
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1SetIndexes: func(ctx *context.Context, args, reply any) error {
				setcastIndxArg, cancast := args.(*utils.SetIndexesArg)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().SetIndexesDrv(setcastIndxArg.IdxItmType, setcastIndxArg.TntCtx, setcastIndxArg.Indexes, true, utils.NonTransactional)
				return nil
			},
			utils.ReplicatorSv1RemoveIndexes: func(ctx *context.Context, args, reply any) error {
				gIdxArg, cancast := args.(*utils.GetIndexesArg)
				if !cancast {
					return utils.ErrNotConvertible
				}
				dm.DataDB().RemoveIndexesDrv(gIdxArg.IdxItmType, gIdxArg.Tenant, utils.EmptyString)
				return nil
			},
		},
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator): clientConn,
	})
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	idxes := map[string]utils.StringSet{
		"*string:Account:1001": {
			"RL1": struct{}{},
		},
		"*string:Account:1002": {
			"RL1": struct{}{},
			"RL2": struct{}{},
		},
		utils.ConcatenatedKey(utils.MetaNone, utils.MetaAny, utils.MetaAny): {
			"RL4": struct{}{},
			"RL5": struct{}{},
		},
	}
	if err := dm.SetIndexes(utils.CacheResourceFilterIndexes,
		"cgrates.org", idxes, false, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if err := dm.RemoveIndexes(utils.CacheResourceFilterIndexes, "cgrates.org"); err != nil {
		t.Error(err)
	}
}

func TestDmCheckFilters(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
	}()
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.MetaFilters: {
			Remote: true,
		},
	}
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetFilter: func(ctx *context.Context, args, reply any) error {
				fltr := &Filter{
					ID:     "FLTR_1",
					Tenant: "cgrates.org",
					Rules: []*FilterRule{
						{
							Type:    utils.MetaString,
							Element: "~*req.Account",
							Values:  []string{"1001", "1002"},
						},
					},
				}
				*reply.(*Filter) = *fltr
				return nil
			},
		},
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator): clientConn,
	})
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	if err := dm.checkFilters("cgrates.org", []string{"FLTR_1"}); err == nil || err.Error() != "broken reference to filter: <FLTR_1>" {
		t.Error(err)
	}
}

func TestRemoveFilterIndexes(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
	}()
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheThresholdFilterIndexes: {
			Remote: true,
		},
	}
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetIndexes: func(ctx *context.Context, args, reply any) error {
				return utils.ErrNotImplemented
			},
		},
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator): clientConn,
	})
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	fp3 := &Filter{
		Tenant: "cgrates.org",
		ID:     "Filter3",
		Rules: []*FilterRule{
			{
				Element: "~*req.Destination",
				Type:    utils.MetaString,
				Values:  []string{"30", "50"},
			},
		}}
	if err := dm.SetFilter(fp3, true); err != nil {
		t.Error(err)
	}

	if err := removeFilterIndexesForFilter(dm, utils.CacheThresholdFilterIndexes, "cgrates.org", []string{"Filter3"}, utils.StringSet{
		"Filter3:THD1": {},
	}); err != nil {
		t.Error(err)
	}
	config.SetCgrConfig(cfg)
	if err := removeFilterIndexesForFilter(dm, utils.CacheThresholdFilterIndexes, "cgrates.org", []string{"Filter3"}, utils.StringSet{
		"Filter3:THD1": {},
	}); err == nil || err != utils.ErrUnsupporteServiceMethod {
		t.Error(err)
	}
}

func TestGetDispatcherProfileErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpCache := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmpCache
	}()
	Cache.Clear(nil)
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.MetaDispatcherProfiles: {
			Remote: true,
		},
	}
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetDispatcherProfile: func(ctx *context.Context, args, reply any) error {
				return utils.ErrDSPProfileNotFound
			},
		},
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator): clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	// dp := &DispatcherProfile{
	// 	Tenant:     "cgrates.org",
	// 	ID:         "DSP_1",
	// 	Subsystems: []string{"*any"},
	// 	FilterIDs:  []string{"*string:~*req.Account:1001"},
	// 	ActivationInterval: &utils.ActivationInterval{
	// 		ActivationTime: time.Date(2014, 7, 29, 15, 00, 0, 0, time.UTC),
	// 	},
	// }
	Cache.Set(utils.MetaDispatcherProfiles, "cgrates:Dsp1", nil, []string{}, true, utils.NonTransactional)
	if _, err := dm.GetDispatcherProfile("cgrates.org", "Dsp1", true, false, utils.NonTransactional); err == nil || err != utils.ErrDSPProfileNotFound {
		t.Error(err)
	}
	var dm2 *DataManager
	if _, err := dm2.GetDispatcherProfile("cgrates.org", "Dsp1", false, true, utils.NonTransactional); err == nil || err != utils.ErrNoDatabaseConn {
		t.Error(err)
	}
	config.SetCgrConfig(cfg)
	if _, err := dm.GetDispatcherProfile("cgrates.org", "Dsp1", false, true, utils.NonTransactional); err == nil || err != utils.ErrDSPProfileNotFound {
		t.Error(err)
	}
}

func TestRemoveIndexFiltersItem(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpCache := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmpCache
	}()
	Cache.Clear(nil)
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	fltr := &Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_ACCOUNT_1001",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Account",
				Values:  []string{"1001"},
			},
		},
	}
	if err := dm.SetFilter(fltr, false); err != nil {
		t.Error(err)
	}
	thd := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "THD_ACNT_1001",
		FilterIDs: []string{"FLTR_ACCOUNT_1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
		},
		MaxHits:   -1,
		MinHits:   0,
		MinSleep:  0,
		Blocker:   false,
		Weight:    10.0,
		ActionIDs: []string{"TOPUP_MONETARY_10"},
		Async:     false,
	}
	if err := dm.SetThresholdProfile(thd, false); err != nil {
		t.Error(err)
	}
	indexes := map[string]utils.StringSet{
		"*string:*req.Account:1001": {
			"THD_ACNT_1001": struct{}{},
		},
	}

	if err := dm.SetIndexes(utils.CacheReverseFilterIndexes, "cgrates.org:FLTR_ACCOUNT_1001:*threshold_filter_indexes",
		indexes, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	if err := removeIndexFiltersItem(dm, utils.CacheThresholdFilterIndexes, "cgrates", "THD_ACNT_1001", []string{"FLTR_ACCOUNT_1001"}); err != nil {
		t.Error(err)
	}

}

func TestDmRemoveRouteProfileErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
	}()
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	oldRp := &RouteProfile{
		Tenant:  "cgrates.org",
		ID:      "RouteProfilePrefix4",
		Sorting: utils.MetaWeight,
		Routes: []*Route{
			{
				ID:              "route1",
				Weight:          10.0,
				RouteParameters: "param1",
			},
		},
		Weight: 0,
	}
	dm.dataDB = &DataDBMock{
		GetRouteProfileDrvF: func(tenant, id string) (rp *RouteProfile, err error) {
			return oldRp, nil
		},
		RemoveRouteProfileDrvF: func(tenant, id string) error {
			return nil
		},
	}
	if err := dm.RemoveRouteProfile("cgrates.org", "RP1", true); err == nil || err != utils.ErrNotImplemented {
		t.Error(err)
	}
	dm.dataDB = &DataDBMock{
		GetRouteProfileDrvF: func(tenant, id string) (rp *RouteProfile, err error) {
			return oldRp, nil
		},
	}
	if err := dm.RemoveRouteProfile("cgrates.org", "RP1", true); err == nil || err != utils.ErrNotImplemented {
		t.Error(err)
	}
	dm.dataDB = &DataDBMock{}
	if err := dm.RemoveRouteProfile("cgrates.org", "RP1", true); err == nil || err != utils.ErrNotImplemented {
		t.Error(err)
	}
}

func TestDmCheckFiltersRmt(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	defer func() {
		cfg2 := config.NewDefaultCGRConfig()
		config.SetCgrConfig(cfg2)
	}()
	Cache.Clear(nil)
	cfg.DataDbCfg().Items[utils.MetaFilters].Remote = true
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- clMock(func(serviceMethod string, _, _ any) error {
		if serviceMethod == utils.ReplicatorSv1GetFilter {

			return nil
		}
		return utils.ErrNotFound
	})
	dm := NewDataManager(db, cfg.CacheCfg(), NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	}))
	dm.SetFilter(&Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Route",
				Values:  []string{"RouteProfile2"},
			},
		},
	}, true)
	config.SetCgrConfig(cfg)
	if err := dm.checkFilters("cgrates.org", []string{"*string:~*req.Destination:1002", "*gte:~*req.Duration:20m", "FLTR1", "FLTR2"}); err == nil {
		t.Error(err)
	}
	//unfinished
}

func TestDmRebuildReverseForPrefix(t *testing.T) {
	testCases := []struct {
		desc      string
		prefix    string
		expectErr bool
	}{
		{
			desc:      "Valid prefix - ReverseDestinationPrefix",
			prefix:    utils.ReverseDestinationPrefix,
			expectErr: false,
		},
		{
			desc:      "Valid prefix - AccountActionPlansPrefix",
			prefix:    utils.AccountActionPlansPrefix,
			expectErr: false,
		},
		{
			desc:      "Invalid prefix",
			prefix:    "invalid_prefix",
			expectErr: true,
		},
	}
	cfg := config.NewDefaultCGRConfig()
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	dm.SetDestination(&Destination{
		Id:       "Dest",
		Prefixes: []string{"1001", "1002"},
	}, "")

	apls := []*ActionPlan{
		{
			Id:         "DisableBal",
			AccountIDs: utils.StringMap{"cgrates:1001": true},
		},
		{
			Id:         "MoreMinutes",
			AccountIDs: utils.StringMap{"cgrates:1002": true},
		},
	}
	dm.SetAccount(&Account{ID: "cgrates:org:1001"})
	dm.SetAccount(&Account{ID: "cgrates:org:1002"})
	for _, apl := range apls {
		dm.SetActionPlan(apl.Id, apl, true, "")

	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			err := dm.RebuildReverseForPrefix(tc.prefix)
			if tc.expectErr {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

func TestDmUpdateReverseDestination(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	dst := &Destination{Id: "OldDest", Prefixes: []string{"+494", "+495", "+496"}}
	dst2 := &Destination{Id: "NewDest", Prefixes: []string{"+497", "+498", "+499"}}
	if _, rcvErr := dm.GetReverseDestination(dst.Id, false, true, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := dm.SetReverseDestination(dst.Id, dst.Prefixes, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	for i := range dst.Prefixes {
		if rcv, err := dm.GetReverseDestination(dst.Prefixes[i], false, true, utils.NonTransactional); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual([]string{dst.Id}, rcv) {
			t.Errorf("Expecting: %v, received: %v", []string{dst.Id}, rcv)
		}
	}
	if err := dm.UpdateReverseDestination(dst, dst2, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	for i := range dst.Prefixes {
		if rcv, err := dm.GetReverseDestination(dst2.Prefixes[i], false, true, utils.NonTransactional); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual([]string{dst2.Id}, rcv) {
			t.Errorf("Expecting: %v, received: %v", []string{dst.Id}, rcv)
		}
	}
}

func TestIndxFilterContains(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	idb, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(idb, cfg.CacheCfg(), nil)

	ft := &Filter{
		Tenant: "cgrates.org",
		ID:     "Filter1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaContains,
				Element: "~*req.EventType",
				Values:  []string{"Ev"},
			},
		},
	}
	if err := dm.SetFilter(ft, true); err != nil {
		t.Error(err)
	}
	fs := &Filter{
		Tenant: "cgrates.org",
		ID:     "Filter2",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.EventType",
				Values:  []string{"Event1"},
			},
		},
	}
	if err := dm.SetFilter(fs, true); err != nil {
		t.Error(err)
	}

	th := &ThresholdProfile{
		Tenant:             "cgrates.org",
		ID:                 "THD_Test",
		ActivationInterval: &utils.ActivationInterval{},
		FilterIDs:          []string{"Filter1", "Filter2"},
		MaxHits:            12,
		MinSleep:           0,
		Blocker:            true,
		Weight:             1.4,
		ActionIDs:          []string{},
	}

	if err := dm.SetThresholdProfile(th, true); err != nil {
		t.Error(err)
	}
	eIdxes := map[string]utils.StringSet{
		"*string:*req.EventType:Event1": {
			"THD_Test": struct{}{},
		},
	}
	if rcvIdx, err := dm.GetIndexes(
		utils.CacheThresholdFilterIndexes, th.Tenant,
		utils.EmptyString, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIdxes, rcvIdx) {
		t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
	}

}
