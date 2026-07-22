// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestUpdateReplicationFilters(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := NewLocker(cfg)
	cfg.CacheCfg().Partitions[utils.CacheReplicationHosts] = &config.CacheParamCfg{
		Limit: 1,
	}
	cacheS := NewCacheS(cfg, nil, nil, nil, locker)
	data, _ := NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := NewDBConnManager(map[string]DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(cacheS)

	args := &utils.ArgsGetGroupWithAPIOpts{
		Tenant: "cgrates.org",
		ArgsGetGroup: utils.ArgsGetGroup{
			CacheID: utils.CacheReplicationHosts,
			GroupID: utils.AccountPrefix + "cgrates.org:acc1",
		},
	}
	var reply []string

	dm.UpdateReplicationFilters(utils.AccountPrefix, "cgrates.org:acc1", "")
	if err := cacheS.V1GetGroupItemIDs(context.Background(), args,
		&reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %v, received %v", utils.ErrNotFound, err)
	}

	dm.UpdateReplicationFilters(utils.AccountPrefix, "cgrates.org:acc1", utils.MetaLocalHost)
	expected := []string{utils.AccountPrefix + "cgrates.org:acc1:" + utils.MetaLocalHost}
	if err := cacheS.V1GetGroupItemIDs(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %v, received %v", expected, reply)
	}
}

func TestReplicateNnReplicatorSv1(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := NewLocker(cfg)
	cfg.CacheCfg().Partitions[utils.CacheReplicationHosts] = &config.CacheParamCfg{
		Limit: 1,
	}
	cacheS := NewCacheS(cfg, nil, nil, nil, locker)
	connMgr := NewConnManager(cfg)
	connMgr.SetCache(cacheS)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10))
	connId := []string{}
	objIds := "cgrates.org:acc1"
	objType := utils.AccountPrefix
	expErr := "MANDATORY_IE_MISSING: [connIDs]"

	cacheS.Set(ctx, utils.CacheReplicationHosts, objType+"cgrates.org:acc1"+utils.ConcatenatedKeySep+utils.MetaLocalHost, utils.MetaLocalHost, []string{objType + "cgrates.org:acc1"}, true, utils.NonTransactional)
	if err := replicate(ctx, connMgr, connId, true, utils.AccountPrefix, objIds, "GET", "args"); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}
	cancel()
}

func TestReplicateMultipleIDs(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := NewLocker(cfg)
	cfg.CacheCfg().Partitions[utils.CacheReplicationHosts] = &config.CacheParamCfg{
		Limit: 1,
	}
	cacheS := NewCacheS(cfg, nil, nil, nil, locker)
	connMgr := NewConnManager(cfg)
	connMgr.SetCache(cacheS)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10))
	connId := []string{}
	objIds := []string{"cgrates.org:acc1"}
	objType := utils.AccountPrefix
	expErr := "MANDATORY_IE_MISSING: [connIDs]"
	if err := replicateMultipleIDs(ctx, connMgr, connId, false, utils.AccountPrefix, objIds, "GET", "args"); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}
	cancel()

	cacheS.Set(ctx, utils.CacheReplicationHosts, objType+"cgrates.org:acc1"+utils.ConcatenatedKeySep+utils.MetaLocalHost, utils.MetaLocalHost, []string{objType + "cgrates.org:acc1"}, true, utils.NonTransactional)
	if err := replicateMultipleIDs(ctx, connMgr, connId, true, utils.AccountPrefix, objIds, "GET", "args"); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}
	cancel()
}
