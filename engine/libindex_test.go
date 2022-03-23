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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestSplitFilterIndexes(t *testing.T) {
	tntGrpIdxKey := "tntCtx:*prefix:~*accounts:1001"
	tntGrp, idxKey, err := splitFilterIndex(tntGrpIdxKey)
	if err != nil {
		t.Error(err)
	}
	expTntGrp := "tntCtx"
	expIdxKey := "*prefix:~*accounts:1001"
	if expTntGrp != tntGrp && expIdxKey != idxKey {
		t.Errorf("Expected %v and %v\n but received %v and %v", expTntGrp, expIdxKey, tntGrp, idxKey)
	}
}

func TestSplitFilterIndexesWrongFormat(t *testing.T) {
	tntGrpIdxKey := "tntCtx:*prefix:~*accounts"
	_, _, err := splitFilterIndex(tntGrpIdxKey)
	errExp := "WRONG_IDX_KEY_FORMAT<tntCtx:*prefix:~*accounts>"
	if errExp != err.Error() {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}
}

func TestComputeIndexes(t *testing.T) {
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
	transactionID := utils.GenUUID()
	indexes, err := ComputeIndexes(context.Background(), dm, "cgrates.org", utils.EmptyString, utils.CacheThresholdFilterIndexes,
		nil, transactionID, func(tnt, id, grp string) (*[]string, error) {
			th, e := dm.GetThresholdProfile(context.Background(), tnt, id, true, false, utils.NonTransactional)
			if e != nil {
				return nil, e
			}
			return utils.SliceStringPointer(utils.CloneStringSlice(th.FilterIDs)), nil
		}, nil)
	if err != nil {
		t.Error(err)
	}
	exp := make(utils.StringSet)
	exp.Add(utils.ConcatenatedKey("cgrates.org", "*string:*req.Account:1001"))
	if !reflect.DeepEqual(exp, indexes) {
		t.Errorf("Expected %v\n but received %v", exp, indexes)
	}
}

func TestComputeIndexesIDsNotNil(t *testing.T) {
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
	transactionID := utils.GenUUID()
	_, err = ComputeIndexes(context.Background(), dm, "cgrates.org", utils.EmptyString, utils.CacheThresholdFilterIndexes,
		&[]string{utils.CacheThresholdFilterIndexes, utils.CacheAccountsFilterIndexes}, transactionID, func(tnt, id, grp string) (*[]string, error) {
			th, e := dm.GetThresholdProfile(context.Background(), tnt, id, true, false, utils.NonTransactional)
			if e != nil {
				return nil, e
			}
			return utils.SliceStringPointer(utils.CloneStringSlice(th.FilterIDs)), nil
		}, nil)
	if err != utils.ErrNotFound {
		t.Errorf("Expected %v\n but received %v", utils.ErrNotFound, err)
	}
}

func TestRemoveIndexFiltersItem(t *testing.T) {
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
	if err := removeIndexFiltersItem(context.Background(), dm, utils.CacheThresholdFilterIndexes, "cgrates.org", "", []string{"account"}); err != nil {
		t.Error(err)
	}
}

func TestRemoveFilterIndexesForFilter(t *testing.T) {
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
	exp := make(utils.StringSet)
	exp.Add(utils.ConcatenatedKey("cgrates.org", "*string:*req.Account:1001"))
	dm := NewDataManager(dataDB, config.CgrConfig().CacheCfg(), connMng)
	if err := removeFilterIndexesForFilter(context.Background(), dm, utils.CacheThresholdFilterIndexes, "cgrates.org", []string{""}, exp); err != nil {
		t.Error(err)
	}
}
