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

func TestDataManagerDataDB(t *testing.T) {
	var dm *DataManager
	rcv := dm.DataDB()
	if rcv != nil {
		t.Errorf("Expected DataDB to be nil, Received <%+v>", rcv)
	}
}

func TestDataManagerSetFilterDMNil(t *testing.T) {
	expErr := utils.ErrNoDatabaseConn
	var dm *DataManager
	err := dm.SetFilter(context.Background(), nil, true)
	if err != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}
}

func TestDataManagerSetFilterErrConnID(t *testing.T) {
	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaFilters].Remote = true
	config.SetCgrConfig(cfg)

	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)

	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	fltr := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr1",
		Rules: []*FilterRule{{
			Type:    utils.MetaString,
			Element: "~*req.Account",
			Values:  []string{"1001", "1002"},
		}},
	}

	expErr := "MANDATORY_IE_MISSING: [connIDs]"
	err := dm.SetFilter(context.Background(), fltr, true)
	if err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}
}

func TestDataManagerSetFilterErrSetFilterDrv(t *testing.T) {
	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()

	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)

	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	dm.dataDB = &DataDBMock{
		GetFilterDrvF: func(ctx *context.Context, str1 string, str2 string) (*Filter, error) {
			return nil, utils.ErrNotFound
		},
	}
	fltr := &Filter{
		Tenant: "cgrates.org",
		ID:     "fltr1",
		Rules: []*FilterRule{{
			Type:    utils.MetaString,
			Element: "~*req.Account",
			Values:  []string{"1001", "1002"},
		}},
	}

	err := dm.SetFilter(context.Background(), fltr, true)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotImplemented, err)
	}
}

func TestDataManagerSetFilterErrUpdateFilterIndex(t *testing.T) {
	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()

	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), cM)
	dm.dataDB = &DataDBMock{
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return nil, utils.ErrNotImplemented
		},
		SetFilterDrvF: func(ctx *context.Context, fltr *Filter) error { return nil },
	}
	fltr := &Filter{
		Tenant: "cgrates.org",
		ID:     "*stirng:~*req.Account:1001",
		Rules: []*FilterRule{{
			Type:    utils.MetaString,
			Element: "~*req.Account",
			Values:  []string{"1001", "1002"},
		}},
	}

	err := dm.SetFilter(context.Background(), fltr, true)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotImplemented, err)
	}
}

func TestDataManagerSetFilterErrItemReplicate(t *testing.T) {
	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaFilters].Replicate = true
	config.SetCgrConfig(cfg)
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), cM)
	dm.dataDB = &DataDBMock{
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return nil, nil
		},
		SetFilterDrvF: func(ctx *context.Context, fltr *Filter) error { return nil },
	}
	fltr := &Filter{
		Tenant: "cgrates.org",
		ID:     "*stirng:~*req.Account:1001",
		Rules: []*FilterRule{{
			Type:    utils.MetaString,
			Element: "~*req.Account",
			Values:  []string{"1001", "1002"},
		}},
	}

	expErr := "MANDATORY_IE_MISSING: [connIDs]"
	err := dm.SetFilter(context.Background(), fltr, true)
	if err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}
}

func TestDataManagerRemoveFilterNildm(t *testing.T) {

	var dm *DataManager

	expErr := utils.ErrNoDatabaseConn
	err := dm.RemoveFilter(context.Background(), utils.CGRateSorg, "fltr1", true)
	if err != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestDataManagerRemoveFilterErrGetFilter(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()

	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), cM)
	dm.dataDB = &DataDBMock{
		GetFilterDrvF: func(ctx *context.Context, str1, str2 string) (*Filter, error) {
			return nil, utils.ErrNotImplemented
		},
	}

	expErr := utils.ErrNotImplemented
	err := dm.RemoveFilter(context.Background(), utils.CGRateSorg, "fltr1", true)
	if err != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestDataManagerRemoveFilterErrGetIndexes(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()

	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), cM)
	dm.dataDB = &DataDBMock{
		GetFilterDrvF: func(ctx *context.Context, str1, str2 string) (*Filter, error) {
			return nil, utils.ErrNotFound
		},
	}

	expErr := utils.ErrNotImplemented
	err := dm.RemoveFilter(context.Background(), utils.CGRateSorg, "fltr1", true)
	if err != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestDataManagerRemoveFilterErrGetIndexesBrokenReference(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()

	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), cM)
	dm.dataDB = &DataDBMock{
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return nil, nil
		},
	}

	fltrId := "*stirng:~*req.Account:1001:4fields"
	expErr := "cannot remove filter <cgrates.org:*stirng:~*req.Account:1001:4fields> because will broken the reference to following items: null"
	err := dm.RemoveFilter(context.Background(), utils.CGRateSorg, fltrId, true)
	if err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestDataManagerRemoveFilterErrRemoveFilterDrv(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()

	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), cM)
	dm.dataDB = &DataDBMock{
		GetFilterDrvF: func(ctx *context.Context, str1 string, str2 string) (*Filter, error) {
			return &Filter{}, nil
		},
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return map[string]utils.StringSet{}, utils.ErrNotFound
		},
		RemoveFilterDrvF: func(ctx *context.Context, str1 string, str2 string) error {
			return utils.ErrNotImplemented
		},
	}

	fltrId := "fltr1"
	expErr := utils.ErrNotImplemented
	err := dm.RemoveFilter(context.Background(), utils.CGRateSorg, fltrId, true)
	if err != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestDataManagerRemoveFilterErrNilOldFltr(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()

	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), cM)

	var fltrId string
	var tnt string
	expErr := utils.ErrNotFound
	err := dm.RemoveFilter(context.Background(), tnt, fltrId, false)
	if err != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestDataManagerRemoveFilterReplicateTrue(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaFilters].Replicate = true
	cfg.DataDbCfg().RplConns = []string{}
	config.SetCgrConfig(cfg)
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), cM)
	dm.dataDB = &DataDBMock{
		GetFilterDrvF: func(ctx *context.Context, str1 string, str2 string) (*Filter, error) {
			return &Filter{}, nil
		},

		RemoveFilterDrvF: func(ctx *context.Context, str1 string, str2 string) error {
			return nil
		},
	}

	tnt := utils.CGRateSorg
	fltrId := "*stirng:~*req.Account:1001"

	// tested replicate
	dm.RemoveFilter(context.Background(), tnt, fltrId, false)

}

func TestDataManagerRemoveAccountNilDM(t *testing.T) {

	var dm *DataManager

	expErr := utils.ErrNoDatabaseConn
	err := dm.RemoveAccount(context.Background(), utils.CGRateSorg, "acc1", false)
	if err != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestDataManagerRemoveAccountErrGetAccount(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()

	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), cM)
	dm.dataDB = &DataDBMock{
		GetAccountDrvF: func(ctx *context.Context, str1, str2 string) (*utils.Account, error) {
			return nil, utils.ErrNotImplemented
		},
	}

	expErr := utils.ErrNotImplemented
	err := dm.RemoveAccount(context.Background(), utils.CGRateSorg, "fltr1", false)
	if err != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestDataManagerRemoveAccountErrRemoveAccountDrv(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()

	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), cM)
	dm.dataDB = &DataDBMock{
		GetAccountDrvF: func(ctx *context.Context, str1, str2 string) (*utils.Account, error) {
			return &utils.Account{}, nil
		},
		RemoveAccountDrvF: func(ctx *context.Context, str1, str2 string) error {
			return utils.ErrNotImplemented
		},
	}

	expErr := utils.ErrNotImplemented
	err := dm.RemoveAccount(context.Background(), utils.CGRateSorg, "fltr1", false)
	if err != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestDataManagerRemoveAccountErrNiloldRpp(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()

	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), cM)

	var fltrId string
	var tnt string
	expErr := utils.ErrNotFound
	err := dm.RemoveAccount(context.Background(), tnt, fltrId, false)
	if err != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestDataManagerRemoveAccountErrRemoveItemFromFilterIndex(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()

	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), cM)
	dm.dataDB = &DataDBMock{
		GetAccountDrvF: func(ctx *context.Context, str1, str2 string) (*utils.Account, error) {
			return &utils.Account{}, nil
		},
		RemoveAccountDrvF: func(ctx *context.Context, str1, str2 string) error {
			return nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return utils.ErrNotImplemented
		},
	}

	tnt := utils.CGRateSorg
	fltrId := "*stirng:~*req.Account:1001"
	expErr := utils.ErrNotImplemented
	err := dm.RemoveAccount(context.Background(), tnt, fltrId, true)
	if err != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestDataManagerRemoveAccountErrRemoveIndexFiltersItem(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()

	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), cM)
	dm.dataDB = &DataDBMock{
		GetAccountDrvF: func(ctx *context.Context, str1, str2 string) (*utils.Account, error) {
			return &utils.Account{
				FilterIDs: []string{"fltr1"},
			}, nil
		},
		RemoveAccountDrvF: func(ctx *context.Context, str1, str2 string) error {
			return nil
		},
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return nil, utils.ErrNotImplemented
		},
	}

	tnt := utils.CGRateSorg
	fltrId := "*stirng:~*req.Account:1001"
	expErr := utils.ErrNotImplemented
	err := dm.RemoveAccount(context.Background(), tnt, fltrId, true)
	if err != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestDataManagerRemoveAccountReplicateTrue(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaFilters].Replicate = true
	cfg.DataDbCfg().RplConns = []string{}
	config.SetCgrConfig(cfg)
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), cM)
	dm.dataDB = &DataDBMock{
		GetAccountDrvF: func(ctx *context.Context, str1, str2 string) (*utils.Account, error) {
			return &utils.Account{}, nil
		},
		RemoveAccountDrvF: func(ctx *context.Context, str1, str2 string) error {
			return nil
		},
	}

	tnt := utils.CGRateSorg
	fltrId := "*stirng:~*req.Account:1001"

	// tested replicate
	dm.RemoveAccount(context.Background(), tnt, fltrId, false)

}

func TestDataManagerRemoveDispatcherHostErrNilDM(t *testing.T) {

	var dm *DataManager
	if err := dm.RemoveDispatcherHost(context.Background(), utils.CGRateSorg, "*stirng:~*req.Account:1001"); err == nil || err != utils.ErrNoDatabaseConn {
		t.Error(err)
	}

}

func TestDataManagerRemoveDispatcherHostErroldDppNil(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), cM)

	if err := dm.RemoveDispatcherHost(context.Background(), utils.CGRateSorg, "*stirng:~*req.Account:1001"); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}

}

func TestDataManagerRemoveDispatcherHostErrGetDisp(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), cM)

	dm.dataDB = &DataDBMock{
		GetDispatcherHostDrvF: func(ctx *context.Context, s1, s2 string) (*DispatcherHost, error) {
			return nil, utils.ErrNotImplemented
		},
	}

	if err := dm.RemoveDispatcherHost(context.Background(), utils.CGRateSorg, "*stirng:~*req.Account:1001"); err == nil || err != utils.ErrNotImplemented {
		t.Error(err)
	}

}

func TestDataManagerRemoveDispatcherHostErrRemoveDisp(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), cM)

	dm.dataDB = &DataDBMock{
		GetDispatcherHostDrvF: func(ctx *context.Context, s1, s2 string) (*DispatcherHost, error) {
			return &DispatcherHost{}, nil
		},
		RemoveDispatcherHostDrvF: func(ctx *context.Context, s1, s2 string) error {
			return utils.ErrNotImplemented
		},
	}

	if err := dm.RemoveDispatcherHost(context.Background(), utils.CGRateSorg, "*stirng:~*req.Account:1001"); err == nil || err != utils.ErrNotImplemented {
		t.Error(err)
	}

}

func TestDataManagerRemoveDispatcherHostReplicateTrue(t *testing.T) {
	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaDispatcherHosts].Replicate = true
	cfg.DataDbCfg().RplConns = []string{}
	config.SetCgrConfig(cfg)
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), cM)

	dm.dataDB = &DataDBMock{
		GetDispatcherHostDrvF: func(ctx *context.Context, s1, s2 string) (*DispatcherHost, error) {
			return &DispatcherHost{}, nil
		},
		RemoveDispatcherHostDrvF: func(ctx *context.Context, s1, s2 string) error {
			return nil
		},
	}

	// tested replicate
	dm.RemoveDispatcherHost(context.Background(), utils.CGRateSorg, "*stirng:~*req.Account:1001")

}

func TestDataManagerSetDispatcherHostErrNilDM(t *testing.T) {

	var dm *DataManager
	if err := dm.SetDispatcherHost(context.Background(), nil); err == nil || err != utils.ErrNoDatabaseConn {
		t.Error(err)
	}

}

func TestDataManagerSetDispatcherHostErrDataDB(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), cM)
	dm.dataDB = &DataDBMock{
		SetDispatcherHostDrvF: func(ctx *context.Context, dh *DispatcherHost) error {
			return utils.ErrNotImplemented
		},
	}
	defer data.Close()
	if err := dm.SetDispatcherHost(context.Background(), nil); err == nil || err != utils.ErrNotImplemented {
		t.Error(err)
	}

}

func TestDataManagerSetDispatcherHostReplicateTrue(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaDispatcherHosts].Replicate = true

	config.SetCgrConfig(cfg)
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)

	dm := NewDataManager(data, cfg.CacheCfg(), nil)

	dpp := &DispatcherHost{
		Tenant: utils.CGRateSorg,
		RemoteHost: &config.RemoteHost{
			ID:                   "ID",
			Address:              "127.0.0.1",
			Transport:            utils.MetaJSON,
			ConnectAttempts:      1,
			Reconnects:           1,
			MaxReconnectInterval: time.Minute,
			ConnectTimeout:       time.Nanosecond,
			ReplyTimeout:         time.Nanosecond,
			TLS:                  true,
			ClientKey:            "key",
			ClientCertificate:    "ce",
			CaCertificate:        "ca",
		},
	}
	// tested replicate
	dm.SetDispatcherHost(context.Background(), dpp)

}

func TestDMRemoveAccountReplicate(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaAccounts].Replicate = true
	config.SetCgrConfig(cfg)

	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), cM)
	dm.dataDB = &DataDBMock{
		GetAccountDrvF: func(ctx *context.Context, str1, str2 string) (*utils.Account, error) {
			return &utils.Account{}, nil
		},
		RemoveAccountDrvF: func(ctx *context.Context, str1, str2 string) error {
			return nil
		},
	}

	// tested replicate
	if err := dm.RemoveAccount(context.Background(), utils.CGRateSorg, "accId", false); err != nil {
		t.Error(err)
	}
}

func TestDMSetAccountNilDM(t *testing.T) {

	var dm *DataManager
	ap := &utils.Account{}

	expErr := utils.ErrNoDatabaseConn
	if err := dm.SetAccount(context.Background(), ap, false); err == nil || err != utils.ErrNoDatabaseConn {
		t.Errorf("Expected error <%v>, Received <%v>", expErr, err)
	}
}

func TestDMSetAccountcheckFiltersErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), cM)

	ap := &utils.Account{
		Tenant: "cgrates.org",
		ID:     "accId",
		Weights: []*utils.DynamicWeight{
			{
				Weight: 0,
			},
		},
		FilterIDs: []string{":::"},
		Balances: map[string]*utils.Balance{
			"AbstractBalance1": {
				ID: "AbstractBalance1",
				Weights: utils.DynamicWeights{
					{
						Weight: 25,
					},
				},
				Type:  utils.MetaAbstract,
				Units: utils.NewDecimal(int64(40*time.Second), 0),
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    utils.NewDecimal(int64(time.Second), 0),
						FixedFee:     utils.NewDecimal(0, 0),
						RecurrentFee: utils.NewDecimal(0, 0),
					},
				},
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				FilterIDs: []string{"fltrID"},
				Blocker:   true,
			},
		},
		Opts:         make(map[string]interface{}),
		ThresholdIDs: []string{utils.MetaNone},
	}

	expErr := "broken reference to filter: <:::> for item with ID: cgrates.org:accId"
	if err := dm.SetAccount(context.Background(), ap, true); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received <%v>", expErr, err)
	}
}

func TestDMSetAccountGetAccountErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), cM)
	dm.dataDB = &DataDBMock{
		GetAccountDrvF: func(ctx *context.Context, str1, str2 string) (*utils.Account, error) {
			return nil, utils.ErrNotImplemented
		},
	}

	ap := &utils.Account{
		Tenant: "cgrates.org",
		ID:     "accId",
		Weights: []*utils.DynamicWeight{
			{
				Weight: 0,
			},
		},
		FilterIDs: []string{"*stirng:~*req.Account:1001"},
		Balances: map[string]*utils.Balance{
			"AbstractBalance1": {
				ID: "AbstractBalance1",
				Weights: utils.DynamicWeights{
					{
						Weight: 25,
					},
				},
				Type:  utils.MetaAbstract,
				Units: utils.NewDecimal(int64(40*time.Second), 0),
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    utils.NewDecimal(int64(time.Second), 0),
						FixedFee:     utils.NewDecimal(0, 0),
						RecurrentFee: utils.NewDecimal(0, 0),
					},
				},
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				FilterIDs: []string{"fltrID"},
				Blocker:   true,
			},
		},
		Opts:         make(map[string]interface{}),
		ThresholdIDs: []string{utils.MetaNone},
	}

	expErr := utils.ErrNotImplemented
	if err := dm.SetAccount(context.Background(), ap, true); err == nil || err != expErr {
		t.Errorf("Expected error <%v>, Received <%v>", expErr, err)
	}
}

func TestDMSetAccountSetAccountDrvErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), cM)
	dm.dataDB = &DataDBMock{
		GetAccountDrvF: func(ctx *context.Context, str1, str2 string) (*utils.Account, error) {
			return &utils.Account{
				Tenant: "cgrates.org",
			}, nil
		},
		SetAccountDrvF: func(ctx *context.Context, profile *utils.Account) error {
			return utils.ErrNotImplemented
		},
	}

	ap := &utils.Account{
		Tenant: "cgrates.org",
		ID:     "accId",
		Weights: []*utils.DynamicWeight{
			{
				Weight: 0,
			},
		},
		FilterIDs: []string{"*stirng:~*req.Account:1001"},
		Balances: map[string]*utils.Balance{
			"AbstractBalance1": {
				ID: "AbstractBalance1",
				Weights: utils.DynamicWeights{
					{
						Weight: 25,
					},
				},
				Type:  utils.MetaAbstract,
				Units: utils.NewDecimal(int64(40*time.Second), 0),
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    utils.NewDecimal(int64(time.Second), 0),
						FixedFee:     utils.NewDecimal(0, 0),
						RecurrentFee: utils.NewDecimal(0, 0),
					},
				},
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				FilterIDs: []string{"fltrID"},
				Blocker:   true,
			},
		},
		Opts:         make(map[string]interface{}),
		ThresholdIDs: []string{utils.MetaNone},
	}

	expErr := utils.ErrNotImplemented
	if err := dm.SetAccount(context.Background(), ap, true); err == nil || err != expErr {
		t.Errorf("Expected error <%v>, Received <%v>", expErr, err)
	}
}

func TestDMSetAccountupdatedIndexesErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), cM)
	dm.dataDB = &DataDBMock{
		GetAccountDrvF: func(ctx *context.Context, str1, str2 string) (*utils.Account, error) {
			return &utils.Account{
				Tenant: "cgrates.org",
			}, nil
		},
		SetAccountDrvF: func(ctx *context.Context, profile *utils.Account) error {
			return nil
		},
	}

	ap := &utils.Account{
		Tenant: "cgrates.org",
		ID:     "accId",
		Weights: []*utils.DynamicWeight{
			{
				Weight: 0,
			},
		},
		FilterIDs: []string{"*stirng:~*req.Account:1001"},
		Balances: map[string]*utils.Balance{
			"AbstractBalance1": {
				ID: "AbstractBalance1",
				Weights: utils.DynamicWeights{
					{
						Weight: 25,
					},
				},
				Type:  utils.MetaAbstract,
				Units: utils.NewDecimal(int64(40*time.Second), 0),
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    utils.NewDecimal(int64(time.Second), 0),
						FixedFee:     utils.NewDecimal(0, 0),
						RecurrentFee: utils.NewDecimal(0, 0),
					},
				},
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				FilterIDs: []string{"fltrID"},
				Blocker:   true,
			},
		},
		Opts:         make(map[string]interface{}),
		ThresholdIDs: []string{utils.MetaNone},
	}

	expErr := utils.ErrNotImplemented
	if err := dm.SetAccount(context.Background(), ap, true); err == nil || err != expErr {
		t.Errorf("Expected error <%v>, Received <%v>", expErr, err)
	}
}

func TestDMSetAccountReplicateTrue(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaAccounts].Replicate = true
	config.SetCgrConfig(cfg)

	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), cM)
	dm.dataDB = &DataDBMock{
		GetAccountDrvF: func(ctx *context.Context, str1, str2 string) (*utils.Account, error) {
			return &utils.Account{}, nil
		},
		SetAccountDrvF: func(ctx *context.Context, profile *utils.Account) error {
			return nil
		},
	}

	ap := &utils.Account{
		Tenant: "cgrates.org",
		ID:     "accId",
		Weights: []*utils.DynamicWeight{
			{
				Weight: 0,
			},
		},
		FilterIDs: []string{"*stirng:~*req.Account:1001"},
		Balances: map[string]*utils.Balance{
			"AbstractBalance1": {
				ID: "AbstractBalance1",
				Weights: utils.DynamicWeights{
					{
						Weight: 25,
					},
				},
				Type:  utils.MetaAbstract,
				Units: utils.NewDecimal(int64(40*time.Second), 0),
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    utils.NewDecimal(int64(time.Second), 0),
						FixedFee:     utils.NewDecimal(0, 0),
						RecurrentFee: utils.NewDecimal(0, 0),
					},
				},
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				FilterIDs: []string{"fltrID"},
				Blocker:   true,
			},
		},
		Opts:         make(map[string]interface{}),
		ThresholdIDs: []string{utils.MetaNone},
	}
	// tests replicete
	dm.SetAccount(context.Background(), ap, false)
}

func TestDMRemoveThresholdProfileNilDM(t *testing.T) {

	var dm *DataManager

	expErr := utils.ErrNoDatabaseConn
	if err := dm.RemoveThresholdProfile(context.Background(), utils.CGRateSorg, "ThrPrf1", false); err == nil || err != utils.ErrNoDatabaseConn {
		t.Errorf("Expected error <%v>, Received <%v>", expErr, err)
	}
}

func TestDMRemoveThresholdProfileGetErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), cM)
	dm.dataDB = &DataDBMock{
		GetThresholdProfileDrvF: func(ctx *context.Context, tenant, id string) (tp *ThresholdProfile, err error) {
			return nil, utils.ErrNotImplemented
		},
	}

	expErr := utils.ErrNotImplemented
	if err := dm.RemoveThresholdProfile(context.Background(), utils.CGRateSorg, "ThrPrf1", false); err == nil || err != expErr {
		t.Errorf("Expected error <%v>, Received <%v>", expErr, err)
	}
}

func TestDMRemoveThresholdProfileRmvErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), cM)
	dm.dataDB = &DataDBMock{
		GetThresholdProfileDrvF: func(ctx *context.Context, tenant, id string) (tp *ThresholdProfile, err error) {
			return nil, nil
		},
		RemThresholdProfileDrvF: func(ctx *context.Context, tenant, id string) (err error) { return utils.ErrNotImplemented },
	}

	expErr := utils.ErrNotImplemented
	if err := dm.RemoveThresholdProfile(context.Background(), utils.CGRateSorg, "ThrPrf1", false); err == nil || err != expErr {
		t.Errorf("Expected error <%v>, Received <%v>", expErr, err)
	}
}

func TestDMRemoveThresholdProfileOldThrNil(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), cM)
	dm.dataDB = &DataDBMock{
		GetThresholdProfileDrvF: func(ctx *context.Context, tenant, id string) (tp *ThresholdProfile, err error) {
			return nil, nil
		},
		RemThresholdProfileDrvF: func(ctx *context.Context, tenant, id string) (err error) { return nil },
	}

	expErr := utils.ErrNotFound
	if err := dm.RemoveThresholdProfile(context.Background(), utils.CGRateSorg, "ThrPrf1", false); err == nil || err != expErr {
		t.Errorf("Expected error <%v>, Received <%v>", expErr, err)
	}
}

func TestDMRemoveThresholdProfileIndxTrueErr1(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), cM)
	dm.dataDB = &DataDBMock{
		GetThresholdProfileDrvF: func(ctx *context.Context, tenant, id string) (tp *ThresholdProfile, err error) {
			return &ThresholdProfile{
				Tenant:           "cgrates.org",
				ID:               "THD_2",
				FilterIDs:        []string{"*string:~*req.Account:1001"},
				ActionProfileIDs: []string{"actPrfID"},
				MaxHits:          7,
				MinHits:          0,
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
				Async: true,
			}, nil
		},
		RemThresholdProfileDrvF: func(ctx *context.Context, tenant, id string) (err error) { return nil },
	}

	expErr := utils.ErrNotImplemented
	if err := dm.RemoveThresholdProfile(context.Background(), utils.CGRateSorg, "THD_2", true); err == nil || err != expErr {
		t.Errorf("Expected error <%v>, Received <%v>", expErr, err)
	}
}

func TestDMRemoveThresholdProfileIndxTrueErr2(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), cM)
	dm.dataDB = &DataDBMock{
		GetThresholdProfileDrvF: func(ctx *context.Context, tenant, id string) (tp *ThresholdProfile, err error) {
			return &ThresholdProfile{
				Tenant:           "cgrates.org",
				ID:               "THD_2",
				FilterIDs:        []string{"*string:~*req.Account:1001", "noPrefix"},
				ActionProfileIDs: []string{"actPrfID"},
				MaxHits:          7,
				MinHits:          0,
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
				Async: true,
			}, nil
		},
		RemThresholdProfileDrvF: func(ctx *context.Context, tenant, id string) (err error) { return nil },
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return nil, utils.ErrNotImplemented
		},
	}

	expErr := utils.ErrNotImplemented
	if err := dm.RemoveThresholdProfile(context.Background(), utils.CGRateSorg, "THD_2", true); err == nil || err != expErr {
		t.Errorf("Expected error <%v>, Received <%v>", expErr, err)
	}
}

func TestDMRemoveThresholdProfileReplicateTrue(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaThresholdProfiles].Replicate = true
	config.SetCgrConfig(cfg)

	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), cM)
	dm.dataDB = &DataDBMock{
		GetThresholdProfileDrvF: func(ctx *context.Context, tenant, id string) (tp *ThresholdProfile, err error) {
			return &ThresholdProfile{
				Tenant:           "cgrates.org",
				ID:               "THD_2",
				FilterIDs:        []string{"*string:~*req.Account:1001"},
				ActionProfileIDs: []string{"actPrfID"},
				MaxHits:          7,
				MinHits:          0,
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
				Async: true,
			}, nil
		},
		RemThresholdProfileDrvF: func(ctx *context.Context, tenant, id string) (err error) { return nil },
	}

	// tests replicate

	dm.RemoveThresholdProfile(context.Background(), utils.CGRateSorg, "THD_2", false)
}
