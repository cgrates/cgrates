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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestDatamanagerCacheDataFromDBNoPrfxErr(t *testing.T) {
	dm := NewDataManager(nil, nil, nil)
	err := dm.CacheDataFromDB(context.Background(), "", []string{}, false)
	if err == nil || err.Error() != "unsupported cache prefix" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "unsupported cache prefix", err)
	}
}

func TestDatamanagerCacheDataFromDBNoDMErr(t *testing.T) {
	var dm *DataManager
	err := dm.CacheDataFromDB(context.Background(), "", []string{}, false)
	if err == nil || err != utils.ErrNoDatabaseConn {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrNoDatabaseConn, err)
	}
}

func TestDatamanagerCacheDataFromDBNoLimitZeroErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dm := NewDataManager(nil, cfg.CacheCfg(), nil)
	dm.cacheCfg.Partitions = map[string]*config.CacheParamCfg{
		utils.CachePrefixToInstance[utils.AttributeProfilePrefix]: {
			Limit: 0,
		},
	}
	err := dm.CacheDataFromDB(context.Background(), utils.AttributeProfilePrefix, []string{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestDatamanagerCacheDataFromDBMetaAPIBanErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dm := NewDataManager(nil, cfg.CacheCfg(), nil)
	dm.cacheCfg.Partitions = map[string]*config.CacheParamCfg{
		utils.CachePrefixToInstance[utils.MetaAPIBan]: {
			Limit: 1,
		},
	}
	err := dm.CacheDataFromDB(context.Background(), utils.MetaAPIBan, []string{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestDatamanagerCacheDataFromDBMustBeCached(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dm := NewDataManager(nil, cfg.CacheCfg(), nil)
	dm.cacheCfg.Partitions = map[string]*config.CacheParamCfg{
		utils.CachePrefixToInstance[utils.AttributeProfilePrefix]: {
			Limit: 1,
		},
	}
	err := dm.CacheDataFromDB(context.Background(), utils.AttributeProfilePrefix, []string{utils.MetaAny}, true)
	if err != nil {
		t.Error(err)
	}
}

func TestDatamanagerCacheDataFromDBNotCachedResourceProfiles(t *testing.T) {
	// cfg := config.NewDefaultCGRConfig()
	// dm := NewDataManager(nil, cfg.CacheCfg(), nil)
	cfg := config.NewDefaultCGRConfig()
	connMng := NewConnManager(cfg)
	dataDB, err := NewDataDBConn(cfg.DataDbCfg().Type,
		cfg.DataDbCfg().Host, cfg.DataDbCfg().Port,
		cfg.DataDbCfg().Name, cfg.DataDbCfg().User,
		cfg.DataDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
		cfg.DataDbCfg().Opts, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	defer dataDB.Close()
	dm := NewDataManager(dataDB, config.CgrConfig().CacheCfg(), connMng)
	dm.cacheCfg.Partitions = map[string]*config.CacheParamCfg{
		utils.CachePrefixToInstance[utils.ResourceProfilesPrefix]: {
			Limit: 1,
		},
		utils.CacheRPCResponses: {
			Limit: 1,
		},
	}
	err = dm.CacheDataFromDB(context.Background(), utils.ResourceProfilesPrefix, []string{utils.MetaAny}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestDatamanagerCacheDataFromDBNotCachedResources(t *testing.T) {
	// cfg := config.NewDefaultCGRConfig()
	// dm := NewDataManager(nil, cfg.CacheCfg(), nil)
	cfg := config.NewDefaultCGRConfig()
	connMng := NewConnManager(cfg)
	dataDB, err := NewDataDBConn(cfg.DataDbCfg().Type,
		cfg.DataDbCfg().Host, cfg.DataDbCfg().Port,
		cfg.DataDbCfg().Name, cfg.DataDbCfg().User,
		cfg.DataDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
		cfg.DataDbCfg().Opts, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	defer dataDB.Close()
	dm := NewDataManager(dataDB, config.CgrConfig().CacheCfg(), connMng)
	dm.cacheCfg.Partitions = map[string]*config.CacheParamCfg{
		utils.CachePrefixToInstance[utils.ResourcesPrefix]: {
			Limit: 1,
		},
		utils.CacheRPCResponses: {
			Limit: 1,
		},
	}
	err = dm.CacheDataFromDB(context.Background(), utils.ResourcesPrefix, []string{utils.MetaAny}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestDatamanagerCacheDataFromDBNotCachedStatQueueProfiles(t *testing.T) {
	// cfg := config.NewDefaultCGRConfig()
	// dm := NewDataManager(nil, cfg.CacheCfg(), nil)
	cfg := config.NewDefaultCGRConfig()
	connMng := NewConnManager(cfg)
	dataDB, err := NewDataDBConn(cfg.DataDbCfg().Type,
		cfg.DataDbCfg().Host, cfg.DataDbCfg().Port,
		cfg.DataDbCfg().Name, cfg.DataDbCfg().User,
		cfg.DataDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
		cfg.DataDbCfg().Opts, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	defer dataDB.Close()
	dm := NewDataManager(dataDB, config.CgrConfig().CacheCfg(), connMng)
	dm.cacheCfg.Partitions = map[string]*config.CacheParamCfg{
		utils.CachePrefixToInstance[utils.StatQueueProfilePrefix]: {
			Limit: 1,
		},
		utils.CacheRPCResponses: {
			Limit: 1,
		},
	}
	err = dm.CacheDataFromDB(context.Background(), utils.StatQueueProfilePrefix, []string{utils.MetaAny}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestDatamanagerCacheDataFromDBNotCachedStatQueuePrefix(t *testing.T) {
	// cfg := config.NewDefaultCGRConfig()
	// dm := NewDataManager(nil, cfg.CacheCfg(), nil)
	cfg := config.NewDefaultCGRConfig()
	connMng := NewConnManager(cfg)
	dataDB, err := NewDataDBConn(cfg.DataDbCfg().Type,
		cfg.DataDbCfg().Host, cfg.DataDbCfg().Port,
		cfg.DataDbCfg().Name, cfg.DataDbCfg().User,
		cfg.DataDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
		cfg.DataDbCfg().Opts, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	defer dataDB.Close()
	dm := NewDataManager(dataDB, config.CgrConfig().CacheCfg(), connMng)
	dm.cacheCfg.Partitions = map[string]*config.CacheParamCfg{
		utils.CachePrefixToInstance[utils.StatQueuePrefix]: {
			Limit: 1,
		},
		utils.CacheRPCResponses: {
			Limit: 1,
		},
	}
	err = dm.CacheDataFromDB(context.Background(), utils.StatQueuePrefix, []string{utils.MetaAny}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestDatamanagerCacheDataFromDBNotCachedThresholdProfilePrefix(t *testing.T) {
	// cfg := config.NewDefaultCGRConfig()
	// dm := NewDataManager(nil, cfg.CacheCfg(), nil)
	cfg := config.NewDefaultCGRConfig()
	connMng := NewConnManager(cfg)
	dataDB, err := NewDataDBConn(cfg.DataDbCfg().Type,
		cfg.DataDbCfg().Host, cfg.DataDbCfg().Port,
		cfg.DataDbCfg().Name, cfg.DataDbCfg().User,
		cfg.DataDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
		cfg.DataDbCfg().Opts, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	defer dataDB.Close()
	dm := NewDataManager(dataDB, config.CgrConfig().CacheCfg(), connMng)
	dm.cacheCfg.Partitions = map[string]*config.CacheParamCfg{
		utils.CachePrefixToInstance[utils.ThresholdProfilePrefix]: {
			Limit: 1,
		},
		utils.CacheRPCResponses: {
			Limit: 1,
		},
	}
	err = dm.CacheDataFromDB(context.Background(), utils.ThresholdProfilePrefix, []string{utils.MetaAny}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestDatamanagerCacheDataFromDBNotCachedThresholdPrefix(t *testing.T) {
	// cfg := config.NewDefaultCGRConfig()
	// dm := NewDataManager(nil, cfg.CacheCfg(), nil)
	cfg := config.NewDefaultCGRConfig()
	connMng := NewConnManager(cfg)
	dataDB, err := NewDataDBConn(cfg.DataDbCfg().Type,
		cfg.DataDbCfg().Host, cfg.DataDbCfg().Port,
		cfg.DataDbCfg().Name, cfg.DataDbCfg().User,
		cfg.DataDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
		cfg.DataDbCfg().Opts, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	defer dataDB.Close()
	dm := NewDataManager(dataDB, config.CgrConfig().CacheCfg(), connMng)
	dm.cacheCfg.Partitions = map[string]*config.CacheParamCfg{
		utils.CachePrefixToInstance[utils.ThresholdPrefix]: {
			Limit: 1,
		},
		utils.CacheRPCResponses: {
			Limit: 1,
		},
	}
	err = dm.CacheDataFromDB(context.Background(), utils.ThresholdPrefix, []string{utils.MetaAny}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestDatamanagerCacheDataFromDBNotCachedFilterPrefix(t *testing.T) {
	// cfg := config.NewDefaultCGRConfig()
	// dm := NewDataManager(nil, cfg.CacheCfg(), nil)
	cfg := config.NewDefaultCGRConfig()
	connMng := NewConnManager(cfg)
	dataDB, err := NewDataDBConn(cfg.DataDbCfg().Type,
		cfg.DataDbCfg().Host, cfg.DataDbCfg().Port,
		cfg.DataDbCfg().Name, cfg.DataDbCfg().User,
		cfg.DataDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
		cfg.DataDbCfg().Opts, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	defer dataDB.Close()
	dm := NewDataManager(dataDB, config.CgrConfig().CacheCfg(), connMng)
	dm.cacheCfg.Partitions = map[string]*config.CacheParamCfg{
		utils.CachePrefixToInstance[utils.FilterPrefix]: {
			Limit: 1,
		},
		utils.CacheRPCResponses: {
			Limit: 1,
		},
	}
	err = dm.CacheDataFromDB(context.Background(), utils.FilterPrefix, []string{utils.MetaAny}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestDatamanagerCacheDataFromDBNotCachedRouteProfilePrefix(t *testing.T) {
	// cfg := config.NewDefaultCGRConfig()
	// dm := NewDataManager(nil, cfg.CacheCfg(), nil)
	cfg := config.NewDefaultCGRConfig()
	connMng := NewConnManager(cfg)
	dataDB, err := NewDataDBConn(cfg.DataDbCfg().Type,
		cfg.DataDbCfg().Host, cfg.DataDbCfg().Port,
		cfg.DataDbCfg().Name, cfg.DataDbCfg().User,
		cfg.DataDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
		cfg.DataDbCfg().Opts, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	defer dataDB.Close()
	dm := NewDataManager(dataDB, config.CgrConfig().CacheCfg(), connMng)
	dm.cacheCfg.Partitions = map[string]*config.CacheParamCfg{
		utils.CachePrefixToInstance[utils.RouteProfilePrefix]: {
			Limit: 1,
		},
		utils.CacheRPCResponses: {
			Limit: 1,
		},
	}
	err = dm.CacheDataFromDB(context.Background(), utils.RouteProfilePrefix, []string{utils.MetaAny}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestDatamanagerCacheDataFromDBNotCachedAttributeProfilePrefix(t *testing.T) {
	// cfg := config.NewDefaultCGRConfig()
	// dm := NewDataManager(nil, cfg.CacheCfg(), nil)
	cfg := config.NewDefaultCGRConfig()
	connMng := NewConnManager(cfg)
	dataDB, err := NewDataDBConn(cfg.DataDbCfg().Type,
		cfg.DataDbCfg().Host, cfg.DataDbCfg().Port,
		cfg.DataDbCfg().Name, cfg.DataDbCfg().User,
		cfg.DataDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
		cfg.DataDbCfg().Opts, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	defer dataDB.Close()
	dm := NewDataManager(dataDB, config.CgrConfig().CacheCfg(), connMng)
	dm.cacheCfg.Partitions = map[string]*config.CacheParamCfg{
		utils.CachePrefixToInstance[utils.AttributeProfilePrefix]: {
			Limit: 1,
		},
		utils.CacheRPCResponses: {
			Limit: 1,
		},
	}
	err = dm.CacheDataFromDB(context.Background(), utils.AttributeProfilePrefix, []string{utils.MetaAny}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestDatamanagerCacheDataFromDBNotCachedChargerProfilePrefix(t *testing.T) {
	// cfg := config.NewDefaultCGRConfig()
	// dm := NewDataManager(nil, cfg.CacheCfg(), nil)
	cfg := config.NewDefaultCGRConfig()
	connMng := NewConnManager(cfg)
	dataDB, err := NewDataDBConn(cfg.DataDbCfg().Type,
		cfg.DataDbCfg().Host, cfg.DataDbCfg().Port,
		cfg.DataDbCfg().Name, cfg.DataDbCfg().User,
		cfg.DataDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
		cfg.DataDbCfg().Opts, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	defer dataDB.Close()
	dm := NewDataManager(dataDB, config.CgrConfig().CacheCfg(), connMng)
	dm.cacheCfg.Partitions = map[string]*config.CacheParamCfg{
		utils.CachePrefixToInstance[utils.ChargerProfilePrefix]: {
			Limit: 1,
		},
		utils.CacheRPCResponses: {
			Limit: 1,
		},
	}
	err = dm.CacheDataFromDB(context.Background(), utils.ChargerProfilePrefix, []string{utils.MetaAny}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestDatamanagerCacheDataFromDBNotCachedDispatcherProfilePrefix(t *testing.T) {
	// cfg := config.NewDefaultCGRConfig()
	// dm := NewDataManager(nil, cfg.CacheCfg(), nil)
	cfg := config.NewDefaultCGRConfig()
	connMng := NewConnManager(cfg)
	dataDB, err := NewDataDBConn(cfg.DataDbCfg().Type,
		cfg.DataDbCfg().Host, cfg.DataDbCfg().Port,
		cfg.DataDbCfg().Name, cfg.DataDbCfg().User,
		cfg.DataDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
		cfg.DataDbCfg().Opts, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	defer dataDB.Close()
	dm := NewDataManager(dataDB, config.CgrConfig().CacheCfg(), connMng)
	dm.cacheCfg.Partitions = map[string]*config.CacheParamCfg{
		utils.CachePrefixToInstance[utils.DispatcherProfilePrefix]: {
			Limit: 1,
		},
		utils.CacheRPCResponses: {
			Limit: 1,
		},
	}
	err = dm.CacheDataFromDB(context.Background(), utils.DispatcherProfilePrefix, []string{utils.MetaAny}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestDatamanagerCacheDataFromDBNotCachedDispatcherHostPrefix(t *testing.T) {
	// cfg := config.NewDefaultCGRConfig()
	// dm := NewDataManager(nil, cfg.CacheCfg(), nil)
	cfg := config.NewDefaultCGRConfig()
	connMng := NewConnManager(cfg)
	dataDB, err := NewDataDBConn(cfg.DataDbCfg().Type,
		cfg.DataDbCfg().Host, cfg.DataDbCfg().Port,
		cfg.DataDbCfg().Name, cfg.DataDbCfg().User,
		cfg.DataDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
		cfg.DataDbCfg().Opts, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	defer dataDB.Close()
	dm := NewDataManager(dataDB, config.CgrConfig().CacheCfg(), connMng)
	dm.cacheCfg.Partitions = map[string]*config.CacheParamCfg{
		utils.CachePrefixToInstance[utils.DispatcherHostPrefix]: {
			Limit: 1,
		},
		utils.CacheRPCResponses: {
			Limit: 1,
		},
	}
	err = dm.CacheDataFromDB(context.Background(), utils.DispatcherHostPrefix, []string{utils.MetaAny}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestDatamanagerCacheDataFromDBNotCachedAccountFilterIndexPrfx(t *testing.T) {
	// cfg := config.NewDefaultCGRConfig()
	// dm := NewDataManager(nil, cfg.CacheCfg(), nil)
	cfg := config.NewDefaultCGRConfig()
	connMng := NewConnManager(cfg)
	dataDB, err := NewDataDBConn(cfg.DataDbCfg().Type,
		cfg.DataDbCfg().Host, cfg.DataDbCfg().Port,
		cfg.DataDbCfg().Name, cfg.DataDbCfg().User,
		cfg.DataDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
		cfg.DataDbCfg().Opts, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	defer dataDB.Close()
	dm := NewDataManager(dataDB, config.CgrConfig().CacheCfg(), connMng)
	dm.cacheCfg.Partitions = map[string]*config.CacheParamCfg{
		utils.CachePrefixToInstance[utils.AccountFilterIndexPrfx]: {
			Limit: 1,
		},
		utils.CacheRPCResponses: {
			Limit: 1,
		},
	}
	err = dm.CacheDataFromDB(context.Background(), utils.AccountFilterIndexPrfx, []string{utils.MetaAny}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestDatamanagerCacheDataFromDBNotCachedAccountPrefix(t *testing.T) {
	// cfg := config.NewDefaultCGRConfig()
	// dm := NewDataManager(nil, cfg.CacheCfg(), nil)
	cfg := config.NewDefaultCGRConfig()
	connMng := NewConnManager(cfg)
	dataDB, err := NewDataDBConn(cfg.DataDbCfg().Type,
		cfg.DataDbCfg().Host, cfg.DataDbCfg().Port,
		cfg.DataDbCfg().Name, cfg.DataDbCfg().User,
		cfg.DataDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
		cfg.DataDbCfg().Opts, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	defer dataDB.Close()
	dm := NewDataManager(dataDB, config.CgrConfig().CacheCfg(), connMng)
	dm.cacheCfg.Partitions = map[string]*config.CacheParamCfg{
		utils.CachePrefixToInstance[utils.AccountPrefix]: {
			Limit: 1,
		},
		utils.CacheRPCResponses: {
			Limit: 1,
		},
	}
	err = dm.CacheDataFromDB(context.Background(), utils.AccountPrefix, []string{utils.MetaAny}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestDatamanagerCacheDataFromDBNotCachedRateProfilePrefix(t *testing.T) {
	// cfg := config.NewDefaultCGRConfig()
	// dm := NewDataManager(nil, cfg.CacheCfg(), nil)
	cfg := config.NewDefaultCGRConfig()
	connMng := NewConnManager(cfg)
	dataDB, err := NewDataDBConn(cfg.DataDbCfg().Type,
		cfg.DataDbCfg().Host, cfg.DataDbCfg().Port,
		cfg.DataDbCfg().Name, cfg.DataDbCfg().User,
		cfg.DataDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
		cfg.DataDbCfg().Opts, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	defer dataDB.Close()
	dm := NewDataManager(dataDB, config.CgrConfig().CacheCfg(), connMng)
	dm.cacheCfg.Partitions = map[string]*config.CacheParamCfg{
		utils.CachePrefixToInstance[utils.RateProfilePrefix]: {
			Limit: 1,
		},
		utils.CacheRPCResponses: {
			Limit: 1,
		},
	}
	err = dm.CacheDataFromDB(context.Background(), utils.RateProfilePrefix, []string{utils.MetaAny}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestDatamanagerCacheDataFromDBNotCachedActionProfilePrefix(t *testing.T) {
	// cfg := config.NewDefaultCGRConfig()
	// dm := NewDataManager(nil, cfg.CacheCfg(), nil)
	cfg := config.NewDefaultCGRConfig()
	connMng := NewConnManager(cfg)
	dataDB, err := NewDataDBConn(cfg.DataDbCfg().Type,
		cfg.DataDbCfg().Host, cfg.DataDbCfg().Port,
		cfg.DataDbCfg().Name, cfg.DataDbCfg().User,
		cfg.DataDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
		cfg.DataDbCfg().Opts, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	defer dataDB.Close()
	dm := NewDataManager(dataDB, config.CgrConfig().CacheCfg(), connMng)
	dm.cacheCfg.Partitions = map[string]*config.CacheParamCfg{
		utils.CachePrefixToInstance[utils.ActionProfilePrefix]: {
			Limit: 1,
		},
		utils.CacheRPCResponses: {
			Limit: 1,
		},
	}
	err = dm.CacheDataFromDB(context.Background(), utils.ActionProfilePrefix, []string{utils.MetaAny}, false)
	if err != nil {
		t.Error(err)
	}
}
