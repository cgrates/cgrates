/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or56
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
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestUpdateReplicationFilters(t *testing.T) {
	tmp := Cache
	cfgTmp := config.CgrConfig()

	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheReplicationHosts] = &config.CacheParamCfg{
		Limit: 1,
	}
	config.SetCgrConfig(cfg)
	Cache = NewCacheS(cfg, nil, nil, nil)

	args := &utils.ArgsGetGroupWithAPIOpts{
		Tenant: "cgrates.org",
		ArgsGetGroup: utils.ArgsGetGroup{
			CacheID: utils.CacheReplicationHosts,
			GroupID: utils.AccountPrefix + "cgrates.org:acc1",
		},
	}
	var reply []string

	UpdateReplicationFilters(utils.AccountPrefix, "cgrates.org:acc1", utils.EmptyString)
	if err := Cache.V1GetGroupItemIDs(context.Background(), args,
		&reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %v, received %v", utils.ErrNotFound, err)
	}

	UpdateReplicationFilters(utils.AccountPrefix, "cgrates.org:acc1", utils.MetaLocalHost)
	expected := []string{utils.AccountPrefix + "cgrates.org:acc1:" + utils.MetaLocalHost}
	if err := Cache.V1GetGroupItemIDs(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %v, received %v", expected, reply)
	}

	Cache = tmp
	config.SetCgrConfig(cfgTmp)
}

func TestReplicateNnReplicatorSv1(t *testing.T) {
	tmp := Cache
	cfgTmp := config.CgrConfig()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	connMgr := NewConnManager(cfg)
	cfg.CacheCfg().Partitions[utils.CacheReplicationHosts] = &config.CacheParamCfg{
		Limit: 1,
	}
	config.SetCgrConfig(cfg)
	Cache = NewCacheS(cfg, nil, nil, nil)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10))
	connId := []string{}
	objIds := "cgrates.org:acc1"
	objType := utils.AccountPrefix
	expErr := "MANDATORY_IE_MISSING: [connIDs]"

	Cache.Set(ctx, utils.CacheReplicationHosts, objType+"cgrates.org:acc1"+utils.ConcatenatedKeySep+utils.MetaLocalHost, utils.MetaLocalHost, []string{objType + "cgrates.org:acc1"}, true, utils.NonTransactional)
	if err := replicate(ctx, connMgr, connId, true, utils.AccountPrefix, objIds, "GET", "args"); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}
	cancel()

	Cache = tmp
	config.SetCgrConfig(cfgTmp)
}

func TestReplicateMultipleIDs(t *testing.T) {
	tmp := Cache
	cfgTmp := config.CgrConfig()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	connMgr := NewConnManager(cfg)
	cfg.CacheCfg().Partitions[utils.CacheReplicationHosts] = &config.CacheParamCfg{
		Limit: 1,
	}
	config.SetCgrConfig(cfg)
	Cache = NewCacheS(cfg, nil, nil, nil)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10))
	connId := []string{}
	objIds := []string{"cgrates.org:acc1"}
	objType := utils.AccountPrefix
	expErr := "MANDATORY_IE_MISSING: [connIDs]"
	if err := replicateMultipleIDs(ctx, connMgr, connId, false, utils.AccountPrefix, objIds, "GET", "args"); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}
	cancel()

	Cache.Set(ctx, utils.CacheReplicationHosts, objType+"cgrates.org:acc1"+utils.ConcatenatedKeySep+utils.MetaLocalHost, utils.MetaLocalHost, []string{objType + "cgrates.org:acc1"}, true, utils.NonTransactional)
	if err := replicateMultipleIDs(ctx, connMgr, connId, true, utils.AccountPrefix, objIds, "GET", "args"); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}
	cancel()

	Cache = tmp
	config.SetCgrConfig(cfgTmp)
}
