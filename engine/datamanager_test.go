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
