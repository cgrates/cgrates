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
	"bytes"
	"errors"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

func TestCacheSSetWithReplicateTrue(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)
	args := &utils.ArgCacheReplicateSet{
		CacheID: utils.CacheAccounts,
		ItemID:  "itemID",
		Value: &utils.CachedRPCResponse{
			Result: "reply",
			Error:  nil},
	}
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicationConnsCfg)}
	cfg.CacheCfg().Partitions = map[string]*config.CacheParamCfg{
		args.CacheID: {
			Replicate: true,
		},
	}

	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- &ccMock{
		calls: map[string]func(_ *context.Context, args any, reply any) error{
			utils.CacheSv1ReplicateSet: func(_ *context.Context, args, reply any) error {
				argCache, canCast := args.(*utils.ArgCacheReplicateSet)
				if !canCast {
					return errors.New("cannot cast")
				}
				Cache.Set(nil, argCache.CacheID, argCache.ItemID, argCache.Value, nil, true, utils.EmptyString)
				*reply.(*string) = utils.OK
				return nil
			},
		},
	}
	connMgr := NewConnManager(cfg)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicationConnsCfg), utils.CacheSv1, clientconn)

	stopchan := make(chan struct{}, 1)
	close(stopchan)

	cacheS := NewCacheS(cfg, dm, connMgr, nil)

	if err := cacheS.SetWithReplicate(context.Background(), args); err != nil {
		t.Error(err)
	}

	expectedVal := &utils.CachedRPCResponse{
		Result: "reply",
		Error:  nil,
	}
	if val, ok := Cache.Get(utils.CacheAccounts, "itemID"); !ok {
		t.Errorf("Expected value")
	} else {
		valConverted, canCast := val.(*utils.CachedRPCResponse)
		if !canCast {
			t.Error("Should cast")
		}
		if valConverted.Error != nil {
			t.Errorf("Expected error <%v>, Received error <%v>", expectedVal.Error, valConverted.Error)
		}
		if !reflect.DeepEqual(expectedVal.Result, valConverted.Result) {
			t.Errorf("Expected %v, received %v", utils.ToJSON(expectedVal), utils.ToJSON(valConverted))
		}
	}
}

func TestCacheSSetWithReplicateFalse(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)
	args := &utils.ArgCacheReplicateSet{
		CacheID:  utils.CacheAccounts,
		ItemID:   "itemID",
		Value:    &utils.CachedRPCResponse{Result: "reply", Error: nil},
		GroupIDs: []string{"groupId", "groupId"},
	}
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicationConnsCfg)}
	cfg.CacheCfg().Partitions = map[string]*config.CacheParamCfg{
		args.CacheID: {
			Replicate: false,
		},
	}

	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	connMgr := NewConnManager(cfg)

	stopchan := make(chan struct{}, 1)
	close(stopchan)

	cacheS := NewCacheS(cfg, dm, connMgr, nil)

	if err := cacheS.SetWithReplicate(context.Background(), args); err != nil {
		t.Error(err)
	}
}

func TestCacheSGetWithRemote(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)
	args := &utils.ArgsGetCacheItemWithAPIOpts{
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheAccounts,
			ItemID:  "itemId",
		},
	}
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cfg.CacheCfg().RemoteConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg)}
	cfg.CacheCfg().Partitions = map[string]*config.CacheParamCfg{
		args.CacheID: {
			Remote: true,
		},
	}
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- &ccMock{
		calls: map[string]func(_ *context.Context, args any, reply any) error{
			utils.CacheSv1GetItem: func(_ *context.Context, args, reply any) error {
				var valBack string = "test_value_was_set"
				*reply.(*any) = valBack
				return nil
			},
		},
	}
	connMgr := NewConnManager(cfg)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg), utils.CacheSv1GetItem, clientconn)

	stopchan := make(chan struct{}, 1)
	close(stopchan)

	cacheS := NewCacheS(cfg, dm, connMgr, nil)

	// first we have to set the value in order to get it from our mock
	cacheS.Set(context.Background(), utils.CacheAccounts, "itemId", "test_value_was_set", []string{}, true, utils.NonTransactional)
	var reply any
	expected := "test_value_was_set"
	if err := cacheS.V1GetItemWithRemote(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else {
		strVal, canCast := reply.(string)
		if !canCast {
			t.Error("must be a string")
		}
		if strVal != expected {
			t.Errorf("Expected %v, received %v", expected, strVal)
		}
	}
}

func TestCacheSGetWithRemoteFalse(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)
	args := &utils.ArgsGetCacheItemWithAPIOpts{

		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheAccounts,
			ItemID:  "itemId",
		},
	}
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cfg.CacheCfg().RemoteConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg)}
	cfg.CacheCfg().Partitions = map[string]*config.CacheParamCfg{
		args.CacheID: {
			Remote: false,
		},
	}

	connMgr := NewConnManager(cfg)

	stopchan := make(chan struct{}, 1)
	close(stopchan)

	cacheS := NewCacheS(cfg, dm, connMgr, nil)

	var reply any = utils.OK
	if err := cacheS.V1GetItemWithRemote(context.Background(), args, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotFound, err)
	}
}
func TestRemoveWithoutReplicate(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	connMgr := NewConnManager(cfg)
	chS := NewCacheS(cfg, dm, connMgr, nil)

	chS.tCache.Set(utils.CacheAccounts, "itemId", "value", nil, true, utils.NonTransactional)

	chS.RemoveWithoutReplicate(utils.CacheAccounts, "itemId", true, utils.NonTransactional)
	if _, has := chS.tCache.Get(utils.CacheAccounts, "itemId"); has {
		t.Error("This itemId shouldn't exist")
	}

}

func TestV1GetItemExpiryTimeFromCacheErr(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)
	args := &utils.ArgsGetCacheItemWithAPIOpts{
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheAccounts,
			ItemID:  "itemId",
		},
	}
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cfg.CacheCfg().RemoteConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg)}
	cfg.CacheCfg().Partitions = map[string]*config.CacheParamCfg{}

	cacheS := NewCacheS(cfg, dm, connMgr, nil)

	var reply time.Time
	if err := cacheS.V1GetItemExpiryTime(context.Background(), args, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotFound, err)
	}

}

func TestV1GetItemErr(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)
	args := &utils.ArgsGetCacheItemWithAPIOpts{
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheAccounts,
			ItemID:  "itemId",
		},
	}
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cfg.CacheCfg().RemoteConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg)}
	cfg.CacheCfg().Partitions = map[string]*config.CacheParamCfg{}

	cacheS := NewCacheS(cfg, dm, connMgr, nil)

	var reply any
	if err := cacheS.V1GetItem(context.Background(), args, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotFound, err)
	}

}
func TestV1GetItemIDsErr(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)
	args := &utils.ArgsGetCacheItemIDsWithAPIOpts{
		ArgsGetCacheItemIDs: utils.ArgsGetCacheItemIDs{
			CacheID:      utils.CacheAccounts,
			ItemIDPrefix: "itemId",
		},
	}
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cfg.CacheCfg().RemoteConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg)}
	cfg.CacheCfg().Partitions = map[string]*config.CacheParamCfg{}

	cacheS := NewCacheS(cfg, dm, connMgr, nil)

	var reply []string
	if err := cacheS.V1GetItemIDs(context.Background(), args, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotFound, err)
	}

}

func TestCacheSGetWithRemoteQueryErr(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)
	args := &utils.ArgsGetCacheItemWithAPIOpts{
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheAccounts,
			ItemID:  "itemId",
		},
	}
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cfg.CacheCfg().RemoteConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg)}
	cfg.CacheCfg().Partitions = map[string]*config.CacheParamCfg{
		args.CacheID: {

			Remote: true,
		},
	}
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- &ccMock{
		calls: map[string]func(_ *context.Context, args any, reply any) error{
			utils.CacheSv1GetItem: func(_ *context.Context, args, reply any) error {
				return utils.ErrNotFound
			},
		},
	}
	connMgr := NewConnManager(cfg)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg), utils.CacheSv1GetItem, clientconn)

	cacheS := NewCacheS(cfg, dm, connMgr, nil)

	expErr := utils.ErrNotFound
	if _, err := cacheS.GetWithRemote(context.Background(), args); err == nil || err != expErr {
		t.Error(err)
	}
}

func TestCacheSGetWithRemoteTCacheGet(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)
	args := &utils.ArgsGetCacheItemWithAPIOpts{
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.Accounts,
			ItemID:  "itemId",
		},
	}
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cfg.CacheCfg().RemoteConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg)}

	var customRply any = utils.ArgsGetCacheItem{
		CacheID: utils.Accounts,
		ItemID:  "itemId",
	}
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- &ccMock{
		calls: map[string]func(_ *context.Context, args any, reply any) error{
			utils.CacheSv1GetItem: func(_ *context.Context, args, reply any) error {
				*reply.(*any) = customRply
				return nil
			},
		},
	}
	connMgr := NewConnManager(cfg)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg), utils.CacheSv1GetItem, clientconn)

	cacheS := NewCacheS(cfg, dm, connMgr, nil)

	exp := "expected value"
	cacheS.tCache.Set(utils.Accounts, "itemId", exp, []string{}, true, utils.EmptyString)

	if rcv, err := cacheS.GetWithRemote(context.Background(), args); err != nil {
		t.Error(err)
	} else if rcv != exp {
		t.Errorf("Expected <%v>, received <%v>", exp, rcv)
	}
}

func TestCacheSV1ReplicateRemove(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)
	args := &utils.ArgCacheReplicateRemove{

		CacheID: utils.CacheAccounts,
		ItemID:  "itemId",
		Tenant:  utils.CGRateSorg,
	}
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicationConnsCfg)}
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- &ccMock{
		calls: map[string]func(_ *context.Context, args any, reply any) error{
			utils.CacheSv1ReplicateRemove: func(_ *context.Context, args, reply any) error {
				var valBack string = utils.OK
				*reply.(*any) = valBack
				return nil
			},
		},
	}
	connMgr := NewConnManager(cfg)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicationConnsCfg), utils.CacheSv1, clientconn)

	cacheS := NewCacheS(cfg, dm, connMgr, nil)

	var reply string
	if err := cacheS.V1ReplicateRemove(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected reply <%v>, Received <%v>", utils.OK, reply)
	}
}

func TestCacheSReplicateRemove(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicationConnsCfg)}
	cfg.CacheCfg().Partitions = map[string]*config.CacheParamCfg{
		utils.CacheAccounts: {
			Replicate: true,
		},
	}
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- &ccMock{
		calls: map[string]func(_ *context.Context, args any, reply any) error{
			utils.CacheSv1ReplicateRemove: func(_ *context.Context, args, reply any) error {
				if err := Cache.Remove(context.Background(), utils.CacheAccounts, "itemId", true, utils.NonTransactional); err != nil {
					t.Error(err)
				}
				return nil
			},
		},
	}
	connMgr := NewConnManager(cfg)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicationConnsCfg), utils.CacheSv1, clientconn)

	cacheS := NewCacheS(cfg, dm, connMgr, nil)

	if err := Cache.Set(context.Background(), utils.CacheAccounts, "itemId", "val", nil, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	if err := cacheS.ReplicateRemove(context.Background(), utils.CacheAccounts, "itemId"); err != nil {
		t.Error(err)
	}

	if rcv, ok := Cache.Get(utils.CacheAccounts, "itemId"); ok != false || rcv != nil {
		t.Errorf("Expected rcv <nil>, Received <%v>, OK <%v>", rcv, ok)
	}
}

func TestCacheSV1ReplicateSet(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cacheS := NewCacheS(cfg, dm, connMgr, nil)

	args := &utils.ArgCacheReplicateSet{
		Tenant: utils.CGRateSorg,

		CacheID:  utils.CacheAccounts,
		ItemID:   "itemId",
		Value:    "valinterface",
		GroupIDs: []string{},
	}

	if err := Cache.Set(context.Background(), utils.CacheAccounts, "itemId", "valinterface", []string{}, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	exp := utils.OK
	var reply string
	if err := cacheS.V1ReplicateSet(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if exp != reply {
		t.Errorf("Expected rcv <%v>, Received <%v>", exp, reply)
	}

	getExp := "valinterface"
	if rcv, ok := Cache.Get(utils.CacheAccounts, "itemId"); !ok {
		t.Errorf("Cache.Get did not receive ok, received <%v>", rcv)
	} else if rcv != getExp {
		t.Errorf("Expected rcv <%v>, Received <%v>", getExp, rcv)
	}

}

func TestCacheSV1ReplicateSetErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cacheS := NewCacheS(cfg, dm, connMgr, nil)
	fltr := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaRegex,
				Element: "~*req.Account",
				Values:  []string{"^(?!On.*On\\s.+?wrote:)(On\\s(.+?)wrote:)$"},
			},
		},
	}
	args := &utils.ArgCacheReplicateSet{
		Tenant: utils.CGRateSorg,

		CacheID:  utils.CacheAccounts,
		ItemID:   "itemId",
		Value:    fltr,
		GroupIDs: []string{},
	}

	expErr := "error parsing regexp: invalid or unsupported Perl syntax: `(?!`"
	var reply string
	if err := cacheS.V1ReplicateSet(context.Background(), args, &reply); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, received error <%v>", expErr, err)
	}

}

func TestCacheSCacheDataFromDB(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cacheS := NewCacheS(cfg, dm, connMgr, nil)

	attrs := &utils.AttrReloadCacheWithAPIOpts{
		Tenant:              utils.CGRateSorg,
		AttributeProfileIDs: []string{"cgrates.org:TEST_ATTRIBUTES_TEST"},
	}
	atrPrfl := &AttributeProfile{
		Tenant: utils.CGRateSorg,
		ID:     "TEST_ATTRIBUTES_TEST",
		Attributes: []*Attribute{
			{
				Path:  "*opts.RateSProfile",
				Type:  utils.MetaConstant,
				Value: config.NewRSRParsersMustCompile("RP_2", utils.InfieldSep),
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: false,
			},
		},
	}
	if err := dm.SetAttributeProfile(context.Background(), atrPrfl, true); err != nil {
		t.Error(err)
	}
	if _, err := dm.GetAttributeProfile(context.Background(), utils.CGRateSorg, "TEST_ATTRIBUTES_TEST", true, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	exp := utils.OK
	var reply string
	if err := cacheS.cacheDataFromDB(context.Background(), attrs, &reply, false); err != nil {
		t.Error(err)
	} else if exp != reply {
		t.Errorf("Expected rcv <%v>, Received <%v>", exp, reply)
	}

	if rcv, ok := Cache.Get(utils.CacheAttributeProfiles, "cgrates.org:TEST_ATTRIBUTES_TEST"); !ok {
		t.Errorf("Cache.Get did not receive ok, received <%v>", rcv)
	} else if !reflect.DeepEqual(rcv, atrPrfl) {
		t.Errorf("Expected rcv <%v>, Received <%v>", atrPrfl, rcv)
	}
}

func TestCacheScacheDataFromDBErrCacheDataFromDB(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cacheS := NewCacheS(cfg, nil, connMgr, nil)

	attrs := &utils.AttrReloadCacheWithAPIOpts{
		Tenant:              utils.CGRateSorg,
		AttributeProfileIDs: []string{"cgrates.org:TEST_ATTRIBUTES_TEST"},
	}

	expErr := utils.ErrNoDatabaseConn
	var reply string

	if err := cacheS.cacheDataFromDB(context.Background(), attrs, &reply, true); err == nil || err != expErr {
		t.Errorf("Expected error <%v>, received error <%v>", expErr, err)
	}

}

func TestCacheScacheDataFromDBErrGetItemLoadIDs(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cacheS := NewCacheS(cfg, dm, connMgr, nil)

	attrs := &utils.AttrReloadCacheWithAPIOpts{
		Tenant:              utils.CGRateSorg,
		AttributeProfileIDs: []string{"cgrates.org:TEST_ATTRIBUTES_TEST"},
	}

	dm.dataDB = &DataDBMock{
		GetItemLoadIDsDrvF: func(ctx *context.Context, itemIDPrefix string) (loadIDs map[string]int64, err error) {
			return nil, utils.ErrNotImplemented
		},
	}

	expErr := utils.ErrNotImplemented
	var reply string

	if err := cacheS.cacheDataFromDB(context.Background(), attrs, &reply, true); err == nil || err != expErr {
		t.Errorf("Expected error <%v>, received error <%v>", expErr, err)
	}

}

func TestCacheSV1LoadCache(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cacheS := NewCacheS(cfg, dm, connMgr, nil)

	attrs := &utils.AttrReloadCacheWithAPIOpts{
		Tenant:              utils.CGRateSorg,
		AttributeProfileIDs: []string{"cgrates.org:TEST_ATTRIBUTES_TEST"},
	}
	atrPrfl := &AttributeProfile{
		Tenant: utils.CGRateSorg,
		ID:     "TEST_ATTRIBUTES_TEST",
		Attributes: []*Attribute{
			{
				Path:  "*opts.RateSProfile",
				Type:  utils.MetaConstant,
				Value: config.NewRSRParsersMustCompile("RP_2", utils.InfieldSep),
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: false,
			},
		},
	}
	if err := dm.SetAttributeProfile(context.Background(), atrPrfl, true); err != nil {
		t.Error(err)
	}

	exp := utils.OK
	var reply string
	if err := cacheS.V1LoadCache(context.Background(), attrs, &reply); err != nil {
		t.Error(err)
	} else if exp != reply {
		t.Errorf("Expected rcv <%v>, Received <%v>", exp, reply)
	}

	if rcv, ok := Cache.Get(utils.CacheAttributeProfiles, "cgrates.org:TEST_ATTRIBUTES_TEST"); !ok {
		t.Errorf("Cache.Get did not receive ok, received <%v>", rcv)
	} else if !reflect.DeepEqual(rcv, atrPrfl) {
		t.Errorf("Expected rcv <%v>, Received <%v>", atrPrfl, rcv)
	}
}

func TestCacheSV1ReloadCache(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cacheS := NewCacheS(cfg, dm, connMgr, nil)

	attrs := &utils.AttrReloadCacheWithAPIOpts{
		Tenant:              utils.CGRateSorg,
		AttributeProfileIDs: []string{"cgrates.org:TEST_ATTRIBUTES_TEST"},
	}

	atrPrfl := &AttributeProfile{
		Tenant: utils.CGRateSorg,
		ID:     "TEST_ATTRIBUTES_TEST",
		Attributes: []*Attribute{
			{
				Path:  "*opts.RateSProfile",
				Type:  utils.MetaConstant,
				Value: config.NewRSRParsersMustCompile("RP_2", utils.InfieldSep),
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: false,
			},
		},
	}
	if err := dm.SetAttributeProfile(context.Background(), atrPrfl, true); err != nil {
		t.Error(err)
	}
	if _, err := dm.GetAttributeProfile(context.Background(), utils.CGRateSorg, "TEST_ATTRIBUTES_TEST", true, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	exp := utils.OK
	var reply string
	if err := cacheS.V1ReloadCache(context.Background(), attrs, &reply); err != nil {
		t.Error(err)
	} else if exp != reply {
		t.Errorf("Expected rcv <%v>, Received <%v>", exp, reply)
	}

	if rcv, ok := Cache.Get(utils.CacheAttributeProfiles, "cgrates.org:TEST_ATTRIBUTES_TEST"); !ok {
		t.Errorf("Cache.Get did not receive ok, received <%v>", rcv)
	} else if !reflect.DeepEqual(rcv, atrPrfl) {
		t.Errorf("Expected rcv <%v>, Received <%v>", atrPrfl, rcv)
	}
}

func TestCacheSV1RemoveGroup(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cacheS := NewCacheS(cfg, dm, connMgr, nil)

	args := &utils.ArgsGetGroupWithAPIOpts{
		Tenant:  "cgrates.org",
		APIOpts: map[string]any{},
		ArgsGetGroup: utils.ArgsGetGroup{
			CacheID: utils.CacheAccounts,
			GroupID: "Group",
		},
	}

	if err := Cache.Set(context.Background(), utils.CacheAccounts, "itemId", "valinterface", []string{"Group", "group2"}, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	exp := utils.OK
	var reply string
	if err := cacheS.V1RemoveGroup(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if exp != reply {
		t.Errorf("Expected rcv <%v>, Received <%v>", exp, reply)
	}

	var hasRply bool
	if err := cacheS.V1HasGroup(context.Background(), args, &hasRply); err != nil {
		t.Error(err)
	} else if hasRply {
		t.Error("There are groups in cacheS")
	}

}

func TestV1GetCacheStats(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cacheS := NewCacheS(cfg, dm, connMgr, nil)

	args := &utils.AttrCacheIDsWithAPIOpts{
		APIOpts: map[string]any{
			utils.MetaSubsys: utils.MetaChargers,
		},
		CacheIDs: []string{"cacheId1"},
		Tenant:   "cgrates.org",
	}

	if err := Cache.Set(context.Background(), "cacheId1", "itemId", "valinterface", []string{"GroupId"}, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	exp := cacheS.tCache.GetCacheStats(args.CacheIDs)
	var reply map[string]*ltcache.CacheStats
	if err := cacheS.V1GetCacheStats(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, exp) {
		t.Errorf("Expected rcv <%v>, Received <%v>", exp, reply)
	}

}

func TestCacheSV1Clear(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cacheS := NewCacheS(cfg, dm, connMgr, nil)

	args := &utils.AttrCacheIDsWithAPIOpts{
		APIOpts: map[string]any{
			utils.MetaSubsys: utils.MetaChargers,
		},
		CacheIDs: []string{"cacheId1", "cacheId2", "cacheId3", "cacheId4"},
		Tenant:   "cgrates.org",
	}

	if err := Cache.Set(context.Background(), "cacheId1", "itemId", "valinterface", []string{"GroupId"}, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	var reply string
	if err := cacheS.V1Clear(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected <%v>, Received <%v>", utils.OK, reply)
	}

	if rcv, ok := Cache.Get(utils.CacheAccounts, "cacheId1"); ok {
		t.Errorf("Cache.Get ok shouldnt be true, received <%v>", rcv)
	} else if rcv != nil {
		t.Errorf("Expected <%v>, Received <%v>", nil, rcv)
	}

}

func TestCacheSV1RemoveItems(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cacheS := NewCacheS(cfg, dm, connMgr, nil)

	args := &utils.AttrReloadCacheWithAPIOpts{
		Tenant:              utils.CGRateSorg,
		AttributeProfileIDs: []string{"cgrates.org:TEST_ATTRIBUTES_TEST"},
	}

	if err := cacheS.Set(context.Background(), utils.CacheAttributeProfiles, "cgrates.org:TEST_ATTRIBUTES_TEST", "valinterface", []string{}, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	expNotRemovedGet := "valinterface"
	if rcv, ok := cacheS.Get(utils.CacheAttributeProfiles, "cgrates.org:TEST_ATTRIBUTES_TEST"); !ok {
		t.Errorf("Cache.Get did not receive ok, received <%v>", rcv)
	} else if rcv != expNotRemovedGet {
		t.Errorf("Expected <%v>, Received <%v>", expNotRemovedGet, rcv)
	}
	var reply string
	if err := cacheS.V1RemoveItems(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected <%v>, Received <%v>", utils.OK, reply)
	}

	if rcv, ok := cacheS.Get(utils.CacheAttributeProfiles, "cgrates.org:TEST_ATTRIBUTES_TEST"); ok {
		t.Errorf("Cache.Get shouldnt receive ok, received <%v>", rcv)
	} else if rcv != nil {
		t.Errorf("Expected <%v>, Received <%v>", nil, rcv)
	}

}

func TestCacheSV1RemoveSingular(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cacheS := NewCacheS(cfg, dm, connMgr, nil)
	args := &utils.ArgsGetCacheItemWithAPIOpts{
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheAccounts,
			ItemID:  "itemId",
		},
	}

	if err := cacheS.Set(context.Background(), utils.CacheAccounts, "itemId", "valinterface", []string{}, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	expNotRemovedGet := "valinterface"
	if rcv, ok := cacheS.Get(utils.CacheAccounts, "itemId"); !ok {
		t.Errorf("Cache.Get did not receive ok, received <%v>", rcv)
	} else if rcv != expNotRemovedGet {
		t.Errorf("Expected <%v>, Received <%v>", expNotRemovedGet, rcv)
	}
	var reply string
	if err := cacheS.V1RemoveItem(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected <%v>, Received <%v>", utils.OK, reply)
	}

	if rcv, ok := cacheS.Get(utils.CacheAccounts, "itemId"); ok {
		t.Errorf("Cache.Get shouldnt receive ok, received <%v>", rcv)
	} else if rcv != nil {
		t.Errorf("Expected <%v>, Received <%v>", nil, rcv)
	}

}

func TestCacheSV1GetItemExpiryTime(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	args := &utils.ArgsGetCacheItemWithAPIOpts{
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheAccounts,
			ItemID:  "itemId",
		},
	}
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions = map[string]*config.CacheParamCfg{
		args.CacheID: {
			Limit: 1,
			TTL:   5 * time.Second,
		},
	}
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	cacheS := NewCacheS(cfg, dm, connMgr, nil)

	if err := cacheS.Set(context.Background(), utils.CacheAccounts, "itemId", "valinterface", []string{}, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	exp := time.Now().Year()
	var reply time.Time
	if err := cacheS.V1GetItemExpiryTime(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if reply.Year() != exp {
		t.Errorf("Expected <%v>, Received <%v>", exp, reply)
	}

}

func TestV1GetItemSingular(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	args := &utils.ArgsGetCacheItemWithAPIOpts{
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheAccounts,
			ItemID:  "itemId",
		},
	}
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	cacheS := NewCacheS(cfg, dm, connMgr, nil)

	if err := cacheS.Set(context.Background(), utils.CacheAccounts, "itemId", "valinterface", []string{}, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	exp := any("valinterface")
	var reply any
	if err := cacheS.V1GetItem(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if reply != exp {
		t.Errorf("Expected <%v>, Received <%v>", exp, reply)
	}

}

func TestCacheSV1HasItem(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	args := &utils.ArgsGetCacheItemWithAPIOpts{
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheAccounts,
			ItemID:  "itemId",
		},
	}
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	cacheS := NewCacheS(cfg, dm, connMgr, nil)

	if err := cacheS.Set(context.Background(), utils.CacheAccounts, "itemId", "valinterface", []string{}, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	exp := true
	var reply bool
	if err := cacheS.V1HasItem(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if reply != exp {
		t.Errorf("Expected <%v>, Received <%v>", exp, reply)
	}

}

func TestCacheSV1GetItemIDs(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	args := &utils.ArgsGetCacheItemIDsWithAPIOpts{
		ArgsGetCacheItemIDs: utils.ArgsGetCacheItemIDs{
			CacheID:      utils.CacheAccounts,
			ItemIDPrefix: "itemId",
		},
	}
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	cacheS := NewCacheS(cfg, dm, connMgr, nil)

	if err := cacheS.Set(context.Background(), utils.CacheAccounts, "itemId", "valinterface", []string{}, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	exp := []string{"itemId"}
	var reply []string
	if err := cacheS.V1GetItemIDs(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, reply) {
		t.Errorf("Expected <%v>, Received <%v>", exp, reply)
	}

}

type ccCloner struct {
	mckField string
}

func (cc *ccCloner) Clone() (any, error) {
	cc.mckField = "value"
	return cc, nil
}

func TestCacheSGetCloned(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	cacheS := NewCacheS(cfg, dm, connMgr, nil)

	intfCloneVal := new(ccCloner)

	if err := cacheS.Set(context.Background(), utils.MetaAccounts, "itemId", intfCloneVal, nil, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	expVal := intfCloneVal
	expGetVal, ok := cacheS.Get(utils.MetaAccounts, "itemId")
	if !ok {
		t.Errorf("Cache.Get should receive ok, received <%v>", expGetVal)
	} else if expGetVal != expVal {
		t.Errorf("Expected <%v>, Received <%v>", expVal, expGetVal)
	}

	if rcv, err := cacheS.GetCloned(utils.MetaAccounts, "itemId"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expGetVal, rcv) {
		t.Errorf("Expected <%v>, Received <%v>", expGetVal, rcv)
	}

}

func TestCacheSGetPrecacheChannel(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	cacheS := NewCacheS(cfg, dm, connMgr, nil)

	if err := cacheS.Set(context.Background(), utils.MetaAccounts, "itemId", "valinterface", nil, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	exp := cacheS.pcItems[utils.MetaAccounts]

	if rcv := cacheS.GetPrecacheChannel(utils.MetaAccounts); exp != rcv {
		t.Errorf("Expected <%v>, Received <%v>", exp, rcv)
	}

}

func TestCacheSV1PrecacheStatusDefault(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	args := &utils.AttrCacheIDsWithAPIOpts{
		APIOpts: map[string]any{
			utils.MetaSubsys: utils.MetaChargers,
		},
		CacheIDs: []string{},
		Tenant:   "cgrates.org",
	}

	cacheS := NewCacheS(cfg, dm, connMgr, nil)

	expArgs := &utils.AttrCacheIDsWithAPIOpts{
		APIOpts: map[string]any{
			utils.MetaSubsys: utils.MetaChargers,
		},
		CacheIDs: []string{},
		Tenant:   "cgrates.org",
	}
	expArgs.CacheIDs = utils.CachePartitions.AsSlice()
	pCacheStatus := make(map[string]string)
	for _, cacheID := range expArgs.CacheIDs {
		pCacheStatus[cacheID] = utils.MetaPrecaching
	}

	exp := pCacheStatus

	var reply map[string]string
	if err := cacheS.V1PrecacheStatus(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, exp) {
		t.Errorf("Expected \n<%v>,\nReceived \n<%v>", exp, reply)
	}

}

func TestCacheSV1PrecacheStatusErrUnknownCacheID(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	args := &utils.AttrCacheIDsWithAPIOpts{
		APIOpts: map[string]any{
			utils.MetaSubsys: utils.MetaChargers,
		},
		CacheIDs: []string{"Inproper ID"},
		Tenant:   "cgrates.org",
	}

	cacheS := NewCacheS(cfg, dm, connMgr, nil)

	expErr := "unknown cacheID: Inproper ID"
	var reply map[string]string
	if err := cacheS.V1PrecacheStatus(context.Background(), args, &reply); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, error received <%v>", expErr, err)
	}

}

func TestCacheSV1PrecacheStatusMetaReady(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	args := &utils.AttrCacheIDsWithAPIOpts{
		APIOpts: map[string]any{
			utils.MetaSubsys: utils.MetaChargers,
		},
		CacheIDs: []string{utils.MetaAccounts},
		Tenant:   "cgrates.org",
	}

	cacheS := NewCacheS(cfg, dm, connMgr, nil)

	if err := Cache.Set(context.Background(), utils.MetaAccounts, "itemId", "valinterface", nil, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	go func() {
		cacheS.GetPrecacheChannel(utils.MetaAccounts) <- struct{}{}
	}()
	time.Sleep(10 * time.Millisecond)

	pCacheStatus := make(map[string]string)
	pCacheStatus[utils.MetaAccounts] = utils.MetaReady

	exp := pCacheStatus
	var reply map[string]string
	if err := cacheS.V1PrecacheStatus(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, exp) {
		t.Errorf("Expected \n<%v>,\nReceived \n<%v>", exp, reply)
	}

}

func TestCacheSPrecachePartitions(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions = map[string]*config.CacheParamCfg{
		utils.CacheAccounts: {
			Precache: false,
		},
		utils.MetaAttributeProfiles: {
			Precache: true,
		},
	}
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	cacheS := NewCacheS(cfg, dm, connMgr, nil)

	atrPrfl := &AttributeProfile{
		Tenant: utils.CGRateSorg,
		ID:     "TEST_ATTRIBUTES_TEST",
		Attributes: []*Attribute{
			{
				Path:  "*opts.RateSProfile",
				Type:  utils.MetaConstant,
				Value: config.NewRSRParsersMustCompile("RP_2", utils.InfieldSep),
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: false,
			},
		},
	}
	if err := dm.SetAttributeProfile(context.Background(), atrPrfl, true); err != nil {
		t.Error(err)
	}
	if _, err := dm.GetAttributeProfile(context.Background(), utils.CGRateSorg, "TEST_ATTRIBUTES_TEST", true, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	cacheS.Precache(utils.NewSyncedChan())
	time.Sleep(10 * time.Millisecond)

	if rcv, ok := Cache.Get(utils.CacheAttributeProfiles, "cgrates.org:TEST_ATTRIBUTES_TEST"); !ok {
		t.Errorf("Cache.Get did not receive ok, received <%v>", rcv)
	} else if !reflect.DeepEqual(rcv, atrPrfl) {
		t.Errorf("Expected rcv <%v>, Received <%v>", atrPrfl, rcv)
	}

}

func TestCacheSPrecacheErr(t *testing.T) {

	tmp := Cache
	tmpLog := utils.Logger
	defer func() {
		Cache = tmp
		utils.Logger = tmpLog
	}()

	Cache.Clear(nil)

	buf := new(bytes.Buffer)
	utils.Logger = utils.NewStdLoggerWithWriter(buf, "", 7)

	args := &utils.ArgCacheReplicateSet{
		CacheID: utils.CacheAccounts,
	}
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions = map[string]*config.CacheParamCfg{
		args.CacheID: {
			Precache: true,
		},
	}

	cacheS := NewCacheS(cfg, nil, connMgr, nil)

	cacheS.Precache(utils.NewSyncedChan())
	time.Sleep(10 * time.Millisecond)
	expErr := "<CacheS> precaching cacheID <*accounts>, got error: NO_DATABASE_CONNECTION"

	if rcvTxt := buf.String(); !strings.Contains(rcvTxt, expErr) {
		t.Errorf("Expected <%v>, Received <%v>", expErr, rcvTxt)
	}

	buf.Reset()

}

func TestCacheSBeginTransaction(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	cacheS := NewCacheS(cfg, dm, connMgr, nil)

	expFormat := `........-....-....-....-............`
	rcv := cacheS.BeginTransaction()
	if matched, err := regexp.Match(expFormat, []byte(rcv)); err != nil {
		t.Error(err)
	} else if !matched {
		t.Errorf("Unexpected transaction format, Received <%v>", rcv)
	}

}

func TestCacheSRollbackTransaction(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	cacheS := NewCacheS(cfg, dm, connMgr, nil)

	expFormat := `........-....-....-....-............`
	tranId := cacheS.BeginTransaction()
	if matched, err := regexp.Match(expFormat, []byte(tranId)); err != nil {
		t.Error(err)
	} else if !matched {
		t.Errorf("Unexpected transaction format, Received <%v>", tranId)
	}

	if err := cacheS.Set(context.Background(), utils.CacheAccounts, "itemId", "valinterface", []string{}, true, tranId); err != nil {
		t.Error(err)
	}

	if rcv, ok := cacheS.Get(utils.CacheAccounts, "itemId"); !ok {
		t.Errorf("Cache.Get should receive ok, received <%v>", rcv)
	} else if rcv != "valinterface" {
		t.Errorf("Expected <%v>, Received <%v>", "valinterface", rcv)
	}

	// destroys a transaction from transactions buffer
	cacheS.RollbackTransaction(tranId)

}

func TestCacheSCommitTransaction(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	cacheS := NewCacheS(cfg, dm, connMgr, nil)

	expFormat := `........-....-....-....-............`
	tranId := cacheS.BeginTransaction()
	if matched, err := regexp.Match(expFormat, []byte(tranId)); err != nil {
		t.Error(err)
	} else if !matched {
		t.Errorf("Unexpected transaction format, Received <%v>", tranId)
	}

	if err := cacheS.Set(context.Background(), utils.CacheAccounts, "itemId", "valinterface", []string{}, true, tranId); err != nil {
		t.Error(err)
	}

	if rcv, ok := cacheS.Get(utils.CacheAccounts, "itemId"); !ok {
		t.Errorf("Cache.Get should receive ok, received <%v>", rcv)
	} else if rcv != "valinterface" {
		t.Errorf("Expected <%v>, Received <%v>", "valinterface", rcv)
	}

	// executes the actions in a transaction buffer
	cacheS.CommitTransaction(tranId)

}
