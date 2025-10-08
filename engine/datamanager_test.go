/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
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
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/baningo"
	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestDatamanagerCacheDataFromDBNoPrfxErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dm := NewDataManager(nil, cfg, nil)
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
	dm := NewDataManager(nil, cfg, nil)
	dm.cfg.CacheCfg().Partitions = map[string]*config.CacheParamCfg{
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
	dm := NewDataManager(nil, cfg, nil)
	dm.cfg.CacheCfg().Partitions = map[string]*config.CacheParamCfg{
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
	dm := NewDataManager(nil, cfg, nil)
	dm.cfg.CacheCfg().Partitions = map[string]*config.CacheParamCfg{
		utils.CachePrefixToInstance[utils.AttributeProfilePrefix]: {
			Limit: 1,
		},
	}
	err := dm.CacheDataFromDB(context.Background(), utils.AttributeProfilePrefix, []string{utils.MetaAny}, true)
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

	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)

	dm := NewDataManager(data, cfg, nil)
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

	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)

	dm := NewDataManager(data, cfg, nil)
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

	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)
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
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)
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

	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)
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

	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)
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

	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)
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

	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)
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

	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

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
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)
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

	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)
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

	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)
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

	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

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

	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)
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

	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)
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
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)
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

	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)
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
	if err := dm.SetAccount(context.Background(), ap, false); err != utils.ErrNoDatabaseConn {
		t.Errorf("Expected error <%v>, Received <%v>", expErr, err)
	}
}

func TestDMSetAccountcheckFiltersErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

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
		Opts:         make(map[string]any),
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
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)
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
		Opts:         make(map[string]any),
		ThresholdIDs: []string{utils.MetaNone},
	}

	expErr := utils.ErrNotImplemented
	if err := dm.SetAccount(context.Background(), ap, true); err != expErr {
		t.Errorf("Expected error <%v>, Received <%v>", expErr, err)
	}
}

func TestDMSetAccountSetAccountDrvErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)
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
		Opts:         make(map[string]any),
		ThresholdIDs: []string{utils.MetaNone},
	}

	expErr := utils.ErrNotImplemented
	if err := dm.SetAccount(context.Background(), ap, true); err != expErr {
		t.Errorf("Expected error <%v>, Received <%v>", expErr, err)
	}
}

func TestDMSetAccountupdatedIndexesErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)
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
		Opts:         make(map[string]any),
		ThresholdIDs: []string{utils.MetaNone},
	}

	expErr := utils.ErrNotImplemented
	if err := dm.SetAccount(context.Background(), ap, true); err != expErr {
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

	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)
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
		Opts:         make(map[string]any),
		ThresholdIDs: []string{utils.MetaNone},
	}
	// tests replicete
	dm.SetAccount(context.Background(), ap, false)
}

func TestDMRemoveThresholdProfileNilDM(t *testing.T) {

	var dm *DataManager

	expErr := utils.ErrNoDatabaseConn
	if err := dm.RemoveThresholdProfile(context.Background(), utils.CGRateSorg, "ThrPrf1", false); err != utils.ErrNoDatabaseConn {
		t.Errorf("Expected error <%v>, Received <%v>", expErr, err)
	}
}

func TestDMRemoveThresholdProfileGetErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)
	dm.dataDB = &DataDBMock{
		GetThresholdProfileDrvF: func(ctx *context.Context, tenant, id string) (tp *ThresholdProfile, err error) {
			return nil, utils.ErrNotImplemented
		},
	}

	expErr := utils.ErrNotImplemented
	if err := dm.RemoveThresholdProfile(context.Background(), utils.CGRateSorg, "ThrPrf1", false); err != expErr {
		t.Errorf("Expected error <%v>, Received <%v>", expErr, err)
	}
}

func TestDMRemoveThresholdProfileRmvErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)
	dm.dataDB = &DataDBMock{
		GetThresholdProfileDrvF: func(ctx *context.Context, tenant, id string) (tp *ThresholdProfile, err error) {
			return nil, nil
		},
		RemThresholdProfileDrvF: func(ctx *context.Context, tenant, id string) (err error) { return utils.ErrNotImplemented },
	}

	expErr := utils.ErrNotImplemented
	if err := dm.RemoveThresholdProfile(context.Background(), utils.CGRateSorg, "ThrPrf1", false); err != expErr {
		t.Errorf("Expected error <%v>, Received <%v>", expErr, err)
	}
}

func TestDMRemoveThresholdProfileOldThrNil(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)
	dm.dataDB = &DataDBMock{
		GetThresholdProfileDrvF: func(ctx *context.Context, tenant, id string) (tp *ThresholdProfile, err error) {
			return nil, nil
		},
		RemThresholdProfileDrvF: func(ctx *context.Context, tenant, id string) (err error) { return nil },
	}

	expErr := utils.ErrNotFound
	if err := dm.RemoveThresholdProfile(context.Background(), utils.CGRateSorg, "ThrPrf1", false); err != expErr {
		t.Errorf("Expected error <%v>, Received <%v>", expErr, err)
	}
}

func TestDMRemoveThresholdProfileIndxTrueErr1(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)
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
	if err := dm.RemoveThresholdProfile(context.Background(), utils.CGRateSorg, "THD_2", true); err != expErr {
		t.Errorf("Expected error <%v>, Received <%v>", expErr, err)
	}
}

func TestDMRemoveThresholdProfileIndxTrueErr2(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)
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
	if err := dm.RemoveThresholdProfile(context.Background(), utils.CGRateSorg, "THD_2", true); err != expErr {
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

	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)
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

func TestDMSetThresholdErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)
	dm.dataDB = &DataDBMock{
		SetThresholdDrvF: func(ctx *context.Context, t *Threshold) error { return utils.ErrNotImplemented },
	}

	th := &Threshold{
		Tenant: "cgrates.org",
		ID:     "TH_1",
		Hits:   0,
	}
	expErr := utils.ErrNotImplemented
	if err := dm.SetThreshold(context.Background(), th); err != expErr {
		t.Errorf("Expected error <%v>, Received <%v>", expErr, err)
	}
}

func TestDMSetThresholdReplicateTrue(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaThresholds].Replicate = true
	config.SetCgrConfig(cfg)
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	th := &Threshold{
		Tenant: "cgrates.org",
		ID:     "TH_1",
		Hits:   0,
	}

	// tests replicate

	dm.SetThreshold(context.Background(), th)
}

func TestDMRemoveThresholdReplicateTrue(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaThresholds].Replicate = true
	config.SetCgrConfig(cfg)
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	th := &Threshold{
		Tenant: "cgrates.org",
		ID:     "TH_1",
		Hits:   0,
	}

	if err := dm.DataDB().SetThresholdDrv(context.Background(), th); err != nil {
		t.Error(err)
	}

	rcv, err := dm.GetThreshold(context.Background(), utils.CGRateSorg, "TH_1", true, false, utils.NonTransactional)
	if err != nil {
		t.Error(err)
	} else if *th != *rcv {
		t.Errorf("Expected <%+v> , Received <%+v>", th, rcv)
	}

	// tests replicate

	if err := dm.RemoveThreshold(context.Background(), "cgrates.org", "TH_1"); err != nil {
		t.Error(err)
	}
	if getRcv, err := dm.GetThreshold(context.Background(), utils.CGRateSorg, "TH_1", true, false, utils.NonTransactional); err != utils.ErrNotFound {
		t.Error(err)
	} else if getRcv != nil {
		t.Errorf("Expected <%+v>, \nReceived <%+v>\n", nil, getRcv)
	}
}

func TestDMGetThresholdCacheGetErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	if err := Cache.Set(context.Background(), utils.CacheThresholds, utils.ConcatenatedKey(utils.CGRateSorg, "TH_1"), nil, []string{}, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	_, err := dm.GetThreshold(context.Background(), utils.CGRateSorg, "TH_1", true, false, utils.NonTransactional)
	if err != utils.ErrNotFound {
		t.Error(err)
	}
}

// unfinished, getting **Threshold GetThresholdDrv outputing *, we need plain
// func TestDMGetThresholdSetThErr(t *testing.T) {
// 	tmp := Cache
// 	cfgtmp := config.CgrConfig()
// 	tmpCM := connMgr
// 	defer func() {
// 		Cache = tmp
// 		config.SetCgrConfig(cfgtmp)
// 		connMgr = tmpCM
// 	}()
// 	Cache.Clear(nil)

// 	cfg := config.NewDefaultCGRConfig()
// 	cfg.DataDbCfg().Items[utils.MetaThresholds].Remote = true
// 	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
// 		utils.RemoteConnsCfg)}
// 	config.SetCgrConfig(cfg)
// 	data , _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)

// 	th := &Threshold{
// 		Tenant: "cgrates.org",
// 		ID:     "TH_1",
// 		Hits:   0,
// 	}

// 	cc := &ccMock{
// 		calls: map[string]func(ctx *context.Context, args any, reply any) error{
// 			utils.ReplicatorSv1GetThreshold: func(ctx *context.Context, args, reply any) error {
// 				rplCast, canCast := reply.(*Threshold)
// 				if !canCast {
// 					t.Errorf("Wrong argument type : %T", reply)
// 					return nil
// 				}
// 				*rplCast = *th
// 				return nil
// 			},
// 		},
// 	}

// 	rpcInternal := make(chan birpc.ClientConnector, 1)
// 	rpcInternal <- cc
// 	cM := NewConnManager(cfg)
// 	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg), utils.ReplicatorSv1, rpcInternal)
// 	dm := NewDataManager(data, cfg.CacheCfg(), cM)

// 	_, err := dm.GetThreshold(context.Background(), utils.CGRateSorg, "TH_1", false, false, utils.NonTransactional)
// 	if  err != utils.ErrNotFound {
// 		t.Error(err)
// 	}
// }

func TestDMGetThresholdSetThCacheSetErr(t *testing.T) {
	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.CacheCfg().Partitions[utils.MetaThresholds].Replicate = true
	config.SetCgrConfig(cfg)
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error {

				return utils.ErrNotImplemented
			},
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.CacheSv1, cc)
	dm := NewDataManager(data, cfg, cM)
	Cache = NewCacheS(cfg, dm, cM, nil)

	expErr := utils.ErrNotImplemented
	_, err := dm.GetThreshold(context.Background(), utils.CGRateSorg, "TH_1", false, true, utils.NonTransactional)
	if err != expErr {
		t.Errorf("Expected <%v> , Received <%v>", expErr, err)
	}
}

func TestDMSetStatQueueNewErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)
	experr := "marshal mock error"
	dm.ms = mockMarshal(experr)

	sq := &StatQueue{
		SQMetrics: map[string]StatMetric{
			"key": statMetricMock(""),
		},
	}

	if err := dm.SetStatQueue(context.Background(), sq); err == nil || err.Error() != experr {
		t.Errorf("Expected error <%v>, Received <%v>", experr, err)
	}
}

func TestDMSetStatQueueSetDrvErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)
	dm.dataDB = &DataDBMock{
		SetStatQueueDrvF: func(ctx *context.Context, ssq *StoredStatQueue, sq *StatQueue) error { return utils.ErrNotImplemented },
	}

	sq := &StatQueue{

		SQMetrics: map[string]StatMetric{
			utils.MetaASR: &StatASR{
				Metric: &Metric{
					Value: utils.NewDecimal(1, 0),
					Count: 1,
					Events: map[string]*DecimalWithCompress{
						"cgrates.org:TestStatRemExpired_1": {Stat: utils.NewDecimalFromFloat64(1), CompressFactor: 1},
					},
				},
			},
		},
		sqPrfl: &StatQueueProfile{
			Metrics: []*MetricWithFilters{
				{
					MetricID: utils.MetaASR,
				},
			},
		},
	}

	expErr := utils.ErrNotImplemented
	if err := dm.SetStatQueue(context.Background(), sq); err != expErr {
		t.Errorf("Expected error <%v>, Received <%v>", expErr, err)
	}
}

func TestDMSetStatQueueReplicateTrue(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaStatQueues].Replicate = true
	config.SetCgrConfig(cfg)
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	sq := &StatQueue{

		SQMetrics: map[string]StatMetric{
			utils.MetaASR: &StatASR{
				Metric: &Metric{
					Value: utils.NewDecimal(1, 0),
					Count: 1,
					Events: map[string]*DecimalWithCompress{
						"cgrates.org:TestStatRemExpired_1": {Stat: utils.NewDecimalFromFloat64(1), CompressFactor: 1},
					},
				},
			},
		},
		sqPrfl: &StatQueueProfile{
			Metrics: []*MetricWithFilters{
				{
					MetricID: utils.MetaASR,
				},
			},
		},
	}

	// tests replicate
	dm.SetStatQueue(context.Background(), sq)
}

func TestDMRemoveStatQueueNildb(t *testing.T) {
	var dm *DataManager

	if err := dm.RemoveStatQueue(context.Background(), utils.CGRateSorg, "SQ99"); err != utils.ErrNoDatabaseConn {
		t.Error(err)
	}

}
func TestDMRemoveStatQueueErrDrv(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	dm.dataDB = &DataDBMock{
		RemStatQueueDrvF: func(ctx *context.Context, tenant, id string) (err error) { return utils.ErrNotImplemented },
	}

	if err := dm.RemoveStatQueue(context.Background(), utils.CGRateSorg, "SQ99"); err != utils.ErrNotImplemented {
		t.Error(err)
	}

}

func TestDMRemoveStatQueueReplicate(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaStatQueues].Replicate = true
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	config.SetCgrConfig(cfg)

	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1SetStatQueue: func(ctx *context.Context, args, reply any) error {

				return nil
			},
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.ReplicatorSv1, cc)
	dm := NewDataManager(data, cfg, cM)

	sq := &StatQueue{
		Tenant: utils.CGRateSorg,
		ID:     "sqid99",
		SQMetrics: map[string]StatMetric{
			utils.MetaASR: &StatASR{
				Metric: &Metric{
					Value: utils.NewDecimal(1, 0),
					Count: 1,
					Events: map[string]*DecimalWithCompress{
						"cgrates.org:TestStatRemExpired_1": {Stat: utils.NewDecimalFromFloat64(1), CompressFactor: 1},
					},
				},
			},
		},
		sqPrfl: &StatQueueProfile{
			Metrics: []*MetricWithFilters{
				{
					MetricID: utils.MetaASR,
				},
			},
		},
	}

	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}

	if rcv, err := dm.GetStatQueue(context.Background(), utils.CGRateSorg, "sqid99", true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, sq) {
		t.Errorf("\nexpected \n<%v> \nreceived \n<%v>\n", sq, rcv)
	}

	//tests replicate
	if err := dm.RemoveStatQueue(context.Background(), utils.CGRateSorg, "sqid99"); err != nil {
		t.Error(err)
	}

	if _, err := dm.GetStatQueue(context.Background(), utils.CGRateSorg, "sqid99", true, false, utils.NonTransactional); err != utils.ErrNotFound {
		t.Error(err)
	}

}

func TestDMGetStatQueueProfileErrNildm(t *testing.T) {
	var dm *DataManager
	if _, err := dm.GetStatQueueProfile(context.Background(), utils.CGRateSorg, "sqp99", false, false, utils.NonTransactional); err != utils.ErrNoDatabaseConn {
		t.Error(err)
	}
}
func TestDMGetStatQueueProfileErrNilCacheRead(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	tntID := utils.ConcatenatedKey(utils.CGRateSorg, "sqp99")

	var setVal any
	if err := Cache.Set(context.Background(), utils.CacheStatQueueProfiles, tntID, setVal, []string{}, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	if _, err := dm.GetStatQueueProfile(context.Background(), utils.CGRateSorg, "sqp99", true, false, utils.NonTransactional); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func TestDMGetStatQueueProfileErrRemote(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaStatQueueProfiles].Remote = true
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg)}

	config.SetCgrConfig(cfg)

	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetStatQueueProfile: func(ctx *context.Context, args, reply any) error {

				return nil
			},
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg), utils.ReplicatorSv1, cc)

	dm := NewDataManager(data, cfg, cM)
	sqp := &StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "sqp99",
		FilterIDs:   []string{"fltr_test"},
		QueueLength: 10,
		TTL:         10 * time.Second,
		Metrics: []*MetricWithFilters{
			{
				MetricID: "*sum#~*req.Usage",
			},
		},
		ThresholdIDs: []string{},
		Stored:       true,
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		MinItems: 1,
	}

	dm.dataDB = &DataDBMock{
		GetStatQueueProfileDrvF: func(ctx *context.Context, tenant, id string) (sq *StatQueueProfile, err error) {
			return sqp, utils.ErrNotFound
		},
		SetStatQueueProfileDrvF: func(ctx *context.Context, sq *StatQueueProfile) (err error) { return utils.ErrNotImplemented },
	}

	if _, err := dm.GetStatQueueProfile(context.Background(), utils.CGRateSorg, "sqp99", false, false, utils.NonTransactional); err != utils.ErrNotImplemented {
		t.Error(err)
	}
}

func TestDMGetStatQueueProfileErrCacheWrite(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.CacheCfg().Partitions[utils.CacheStatQueueProfiles].Replicate = true
	config.SetCgrConfig(cfg)

	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error { return utils.ErrNotImplemented },
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.CacheSv1, cc)

	dm := NewDataManager(data, cfg, cM)
	sqp := &StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "sqp99",
		FilterIDs:   []string{"fltr_test"},
		QueueLength: 10,
		TTL:         10 * time.Second,
		Metrics: []*MetricWithFilters{
			{
				MetricID: "*sum#~*req.Usage",
			},
		},
		ThresholdIDs: []string{},
		Stored:       true,
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		MinItems: 1,
	}

	dm.dataDB = &DataDBMock{
		GetStatQueueProfileDrvF: func(ctx *context.Context, tenant, id string) (sq *StatQueueProfile, err error) {
			return sqp, utils.ErrNotFound
		},
	}

	Cache = NewCacheS(cfg, dm, cM, nil)

	if _, err := dm.GetStatQueueProfile(context.Background(), utils.CGRateSorg, "sqp99", false, true, utils.NonTransactional); err != utils.ErrNotImplemented {
		t.Error(err)
	}

}

func TestDMGetStatQueueProfileErr2CacheWrite(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.CacheCfg().Partitions[utils.CacheStatQueueProfiles].Replicate = true
	config.SetCgrConfig(cfg)

	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error { return utils.ErrNotImplemented },
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.CacheSv1, cc)

	dm := NewDataManager(data, cfg, cM)
	sqp := &StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "sqp99",
		FilterIDs:   []string{"fltr_test"},
		QueueLength: 10,
		TTL:         10 * time.Second,
		Metrics: []*MetricWithFilters{
			{
				MetricID: "*sum#~*req.Usage",
			},
		},
		ThresholdIDs: []string{},
		Stored:       true,
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		MinItems: 1,
	}

	if err := dm.DataDB().SetStatQueueProfileDrv(context.Background(), sqp); err != nil {
		t.Error(err)
	}

	Cache = NewCacheS(cfg, dm, cM, nil)

	if _, err := dm.GetStatQueueProfile(context.Background(), utils.CGRateSorg, "sqp99", false, true, utils.NonTransactional); err != utils.ErrNotImplemented {
		t.Error(err)
	}

}

func TestDMGetThresholdProfileSetThErr2(t *testing.T) {
	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.CacheCfg().Partitions[utils.MetaThresholds].Replicate = true

	config.SetCgrConfig(cfg)
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error {

				return utils.ErrNotImplemented
			},
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.CacheSv1, cc)
	dm := NewDataManager(data, cfg, cM)
	dm.dataDB = &DataDBMock{
		GetThresholdDrvF: func(ctx *context.Context, tenant, id string) (*Threshold, error) {
			return &Threshold{}, nil
		},
	}

	Cache = NewCacheS(cfg, dm, cM, nil)

	expErr := utils.ErrNotImplemented
	_, err := dm.GetThreshold(context.Background(), utils.CGRateSorg, "TH_1", false, true, utils.NonTransactional)
	if err != expErr {
		t.Errorf("Expected <%v> , Received <%v>", expErr, err)
	}
}

func TestDMGetThresholdGetThProflErr(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	if err := Cache.Set(context.Background(), utils.CacheThresholdProfiles, utils.ConcatenatedKey(utils.CGRateSorg, "TH_1"), nil, []string{}, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	expErr := utils.ErrNotFound
	_, err := dm.GetThresholdProfile(context.Background(), utils.CGRateSorg, "TH_1", true, true, utils.NonTransactional)
	if err != expErr {
		t.Errorf("Expected <%v> , Received <%v>", expErr, err)
	}
}

func TestDMGetThresholdProfileDMErr(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	var dm *DataManager

	expErr := utils.ErrNoDatabaseConn
	_, err := dm.GetThresholdProfile(context.Background(), utils.CGRateSorg, "TH_1", true, true, utils.NonTransactional)
	if err != expErr {
		t.Errorf("Expected <%v> , Received <%v>", expErr, err)
	}
}

// unfinished
// func TestGetThresholdProfileSetThErr(t *testing.T) {
// 	tmp := Cache
// 	cfgtmp := config.CgrConfig()
// 	tmpCM := connMgr
// 	defer func() {
// 		Cache = tmp
// 		config.SetCgrConfig(cfgtmp)
// 		connMgr = tmpCM
// 	}()
// 	Cache.Clear(nil)

// 	cfg := config.NewDefaultCGRConfig()
// 	cfg.DataDbCfg().Items[utils.MetaThresholdProfiles].Remote = true
// 	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
// 		utils.RemoteConnsCfg)}
// 	config.SetCgrConfig(cfg)
// 	data , _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)

// 	th := &Threshold{
// 		Tenant: "cgrates.org",
// 		ID:     "TH_1",
// 		Hits:   0,
// 	}

// 	cc := &ccMock{
// 		calls: map[string]func(ctx *context.Context, args any, reply any) error{
// 			utils.ReplicatorSv1GetThresholdProfile: func(ctx *context.Context, args, reply any) error {
// 				rplCast, canCast := reply.(*Threshold)
// 				if !canCast {
// 					t.Errorf("Wrong argument type : %T", reply)
// 					return nil
// 				}
// 				*rplCast = *th
// 				return nil
// 			},
// 		},
// 	}

// 	rpcInternal := make(chan birpc.ClientConnector, 1)
// 	rpcInternal <- cc
// 	cM := NewConnManager(cfg)
// 	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg), utils.ReplicatorSv1, rpcInternal)
// 	dm := NewDataManager(data, cfg.CacheCfg(), cM)

// 	_, err := dm.GetThresholdProfile(context.Background(), utils.CGRateSorg, "TH_1", false, false, utils.NonTransactional)
// 	if  err != utils.ErrNotFound {
// 		t.Error(err)
// 	}
// }

func TestDMGetThresholdProfileSetThPrfErr(t *testing.T) {
	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.CacheCfg().Partitions[utils.MetaThresholdProfiles].Replicate = true
	config.SetCgrConfig(cfg)
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error {

				return utils.ErrNotImplemented
			},
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.CacheSv1, cc)
	dm := NewDataManager(data, cfg, cM)
	Cache = NewCacheS(cfg, dm, cM, nil)

	expErr := utils.ErrNotImplemented
	_, err := dm.GetThresholdProfile(context.Background(), utils.CGRateSorg, "TH_1", false, true, utils.NonTransactional)
	if err != expErr {
		t.Errorf("Expected <%v> , Received <%v>", expErr, err)
	}
}

func TestDMGetThresholdProfileSetThPrfErr2(t *testing.T) {
	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.CacheCfg().Partitions[utils.MetaThresholdProfiles].Replicate = true
	config.SetCgrConfig(cfg)
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error {

				return utils.ErrNotImplemented
			},
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.CacheSv1, cc)
	dm := NewDataManager(data, cfg, cM)
	dm.dataDB = &DataDBMock{
		GetThresholdProfileDrvF: func(ctx *context.Context, tenant, id string) (tp *ThresholdProfile, err error) {
			return &ThresholdProfile{}, nil
		},
	}
	Cache = NewCacheS(cfg, dm, cM, nil)

	expErr := utils.ErrNotImplemented
	_, err := dm.GetThresholdProfile(context.Background(), utils.CGRateSorg, "TH_1", false, true, utils.NonTransactional)
	if err != expErr {
		t.Errorf("Expected <%v> , Received <%v>", expErr, err)
	}
}

func TestDMCacheDataFromDBResourceProfilesPrefix(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	rp := &utils.ResourceProfile{
		Tenant:    "cgrates.org",
		ID:        "RL2",
		FilterIDs: []string{"fltr_test"},
		Weights: utils.DynamicWeights{
			{
				Weight: 100,
			}},
		Limit:             2,
		ThresholdIDs:      []string{"TEST_ACTIONS"},
		Blocker:           false,
		UsageTTL:          time.Millisecond,
		AllocationMessage: "ALLOC",
	}

	if err := dm.SetResourceProfile(context.Background(), rp, false); err != nil {
		t.Error(err)
	}

	if _, ok := Cache.Get(utils.CacheResourceProfiles, "cgrates.org:RL2"); ok {
		t.Error("expected ok to be false")
	}

	if err := dm.CacheDataFromDB(context.Background(), utils.ResourceProfilesPrefix, []string{utils.MetaAny}, false); err != nil {
		t.Error(err)
	}

	if rcv, ok := Cache.Get(utils.CacheResourceProfiles, "cgrates.org:RL2"); !ok {
		t.Error("expected ok to be true")
	} else if !reflect.DeepEqual(rcv, rp) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", rp, rcv)
	}

}

func TestDMCacheDataFromDBResourcesPrefix(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	rs := &utils.Resource{
		Tenant: "cgrates.org",
		ID:     "ResGroup2",
		Usages: map[string]*utils.ResourceUsage{
			"RU1": {
				Tenant: "cgrates.org",
				ID:     "RU1",
				Units:  9,
			},
		},
	}

	if err := dm.SetResource(context.Background(), rs); err != nil {
		t.Error(err)
	}

	if _, ok := Cache.Get(utils.CacheResources, "cgrates.org:ResGroup2"); ok {
		t.Error("expected ok to be false")
	}

	if err := dm.CacheDataFromDB(context.Background(), utils.ResourcesPrefix, []string{utils.MetaAny}, false); err != nil {
		t.Error(err)
	}

	if rcv, ok := Cache.Get(utils.CacheResources, "cgrates.org:ResGroup2"); !ok {
		t.Error("expected ok to be true")
	} else if !reflect.DeepEqual(rcv, rs) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", rs, rcv)
	}

}

func TestDMCacheDataFromDBStatQueueProfilePrefix(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	sQP := &StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "StatQueueProfile3",
		FilterIDs:   []string{"fltr_test"},
		QueueLength: 10,
		TTL:         10 * time.Second,
		Metrics: []*MetricWithFilters{
			{
				MetricID: "*sum#~*req.Usage",
			},
		},
		ThresholdIDs: []string{},
		Stored:       true,
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		MinItems: 1,
	}

	if err := dm.SetStatQueueProfile(context.Background(), sQP, false); err != nil {
		t.Error(err)
	}

	if _, ok := Cache.Get(utils.CacheStatQueueProfiles, "cgrates.org:StatQueueProfile3"); ok {
		t.Error("expected ok to be false")
	}

	if err := dm.CacheDataFromDB(context.Background(), utils.StatQueueProfilePrefix, []string{utils.MetaAny}, false); err != nil {
		t.Error(err)
	}

	if rcv, ok := Cache.Get(utils.CacheStatQueueProfiles, "cgrates.org:StatQueueProfile3"); !ok {
		t.Error("expected ok to be true")
	} else if !reflect.DeepEqual(rcv, sQP) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", sQP, rcv)
	}

}

func TestDMCacheDataFromDBStatQueuePrefix(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	sq := &StatQueue{
		Tenant: "cgrates.org",
		ID:     "SQ1",
		SQItems: []SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: make(map[string]StatMetric),
	}

	if err := dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}

	if _, ok := Cache.Get(utils.CacheStatQueues, "cgrates.org:SQ1"); ok {
		t.Error("expected ok to be false")
	}

	if err := dm.CacheDataFromDB(context.Background(), utils.StatQueuePrefix, []string{utils.MetaAny}, false); err != nil {
		t.Error(err)
	}

	if rcv, ok := Cache.Get(utils.CacheStatQueues, "cgrates.org:SQ1"); !ok {
		t.Error("expected ok to be true")
	} else if !reflect.DeepEqual(rcv, sq) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", sq, rcv)
	}

}

func TestDMCacheDataFromDBThresholdProfilePrefix(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	thP := &ThresholdProfile{
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
	}

	if err := dm.SetThresholdProfile(context.Background(), thP, false); err != nil {
		t.Error(err)
	}

	if _, ok := Cache.Get(utils.CacheThresholdProfiles, "cgrates.org:THD_2"); ok {
		t.Error("expected ok to be false")
	}

	if err := dm.CacheDataFromDB(context.Background(), utils.ThresholdProfilePrefix, []string{utils.MetaAny}, false); err != nil {
		t.Error(err)
	}

	if rcv, ok := Cache.Get(utils.CacheThresholdProfiles, "cgrates.org:THD_2"); !ok {
		t.Error("expected ok to be true")
	} else if !reflect.DeepEqual(rcv, thP) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", thP, rcv)
	}

}

func TestDMCacheDataFromDBThresholdPrefix(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	th := &Threshold{
		Tenant: "cgrates.org",
		ID:     "TH_3",
		Hits:   0,
	}

	if err := dm.SetThreshold(context.Background(), th); err != nil {
		t.Error(err)
	}

	if _, ok := Cache.Get(utils.CacheThresholds, "cgrates.org:TH_3"); ok {
		t.Error("expected ok to be false")
	}

	if err := dm.CacheDataFromDB(context.Background(), utils.ThresholdPrefix, []string{utils.MetaAny}, false); err != nil {
		t.Error(err)
	}

	if rcv, ok := Cache.Get(utils.CacheThresholds, "cgrates.org:TH_3"); !ok {
		t.Error("expected ok to be true")
	} else if !reflect.DeepEqual(rcv, th) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", th, rcv)
	}

}

func TestDMCacheDataFromDBFilterPrefix(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	fltr := &Filter{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "FLTR_CP_2",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Charger",
				Values:  []string{"ChargerProfile2"},
			},
		},
	}

	if err := dm.SetFilter(context.Background(), fltr, false); err != nil {
		t.Error(err)
	}

	if _, ok := Cache.Get(utils.CacheFilters, "cgrates.org:FLTR_CP_2"); ok {
		t.Error("expected ok to be false")
	}

	if err := dm.CacheDataFromDB(context.Background(), utils.FilterPrefix, []string{utils.MetaAny}, false); err != nil {
		t.Error(err)
	}

	if rcv, ok := Cache.Get(utils.CacheFilters, "cgrates.org:FLTR_CP_2"); !ok {
		t.Error("expected ok to be true")
	} else if !reflect.DeepEqual(rcv, fltr) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", fltr, rcv)
	}

}

func TestDMCacheDataFromDBRouteProfilePrefix(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	routeProf := &utils.RouteProfile{

		Tenant:            "cgrates.org",
		ID:                "RP_1",
		FilterIDs:         []string{"fltr_test"},
		Weights:           utils.DynamicWeights{{}},
		Sorting:           utils.MetaQOS,
		SortingParameters: []string{"param"},
		Routes: []*utils.Route{{
			ID:             "RT1",
			FilterIDs:      []string{"fltr1"},
			AccountIDs:     []string{"acc1"},
			RateProfileIDs: []string{"rp1"},
			ResourceIDs:    []string{"res1"},
			StatIDs:        []string{"stat1"},
			Weights:        utils.DynamicWeights{{}},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: true,
				},
			},
			RouteParameters: "params",
		}},
	}

	if err := dm.SetRouteProfile(context.Background(), routeProf, false); err != nil {
		t.Error(err)
	}

	if _, ok := Cache.Get(utils.CacheRouteProfiles, "cgrates.org:RP_1"); ok {
		t.Error("expected ok to be false")
	}

	if err := dm.CacheDataFromDB(context.Background(), utils.RouteProfilePrefix, []string{utils.MetaAny}, false); err != nil {
		t.Error(err)
	}

	if rcv, ok := Cache.Get(utils.CacheRouteProfiles, "cgrates.org:RP_1"); !ok {
		t.Error("expected ok to be true")
	} else if !reflect.DeepEqual(rcv, routeProf) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", routeProf, rcv)
	}

}

func TestDMCacheDataFromDBChargerProfilePrefix(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	cpp := &utils.ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "CPP_1",
		FilterIDs:    []string{"FLTR_CP_1"},
		RunID:        "TestRunID",
		AttributeIDs: []string{"*none"},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}

	if err := dm.SetChargerProfile(context.Background(), cpp, false); err != nil {
		t.Error(err)
	}

	if _, ok := Cache.Get(utils.CacheChargerProfiles, "cgrates.org:CPP_1"); ok {
		t.Error("expected ok to be false")
	}

	if err := dm.CacheDataFromDB(context.Background(), utils.ChargerProfilePrefix, []string{utils.MetaAny}, false); err != nil {
		t.Error(err)
	}

	if rcv, ok := Cache.Get(utils.CacheChargerProfiles, "cgrates.org:CPP_1"); !ok {
		t.Error("expected ok to be true")
	} else if !reflect.DeepEqual(rcv, cpp) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", cpp, rcv)
	}

}

func TestDMCacheDataFromDBRateProfilePrefix(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	rpp := &utils.RateProfile{
		ID:        "RP1",
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.Destination:1234"},
		Rates: map[string]*utils.Rate{
			"RT1": {
				ID: "RT1",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(1, 2),
						Unit:          utils.NewDecimal(int64(time.Second), 0),
						Increment:     utils.NewDecimal(int64(time.Second), 0),
					},
				},
			},
		},
	}

	if err := dm.SetRateProfile(context.Background(), rpp, false, false); err != nil {
		t.Error(err)
	}

	if _, ok := Cache.Get(utils.CacheRateProfiles, "cgrates.org:RP1"); ok {
		t.Error("expected ok to be false")
	}

	if err := dm.CacheDataFromDB(context.Background(), utils.RateProfilePrefix, []string{utils.MetaAny}, false); err != nil {
		t.Error(err)
	}

	if rcv, ok := Cache.Get(utils.CacheRateProfiles, "cgrates.org:RP1"); !ok {
		t.Error("expected ok to be true")
	} else if !reflect.DeepEqual(rcv, rpp) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", rpp, rcv)
	}

}

func TestDMCacheDataFromDBActionProfilePrefix(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	ap := &utils.ActionProfile{

		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1"},
		Weights: utils.DynamicWeights{
			{
				Weight: 65,
			},
		},
		Schedule: "* * * * *",
		Targets:  map[string]utils.StringSet{utils.MetaAccounts: {"1001": {}}},
		Actions:  []*utils.APAction{{}},
	}

	if err := dm.SetActionProfile(context.Background(), ap, false); err != nil {
		t.Error(err)
	}

	if _, ok := Cache.Get(utils.CacheActionProfiles, "cgrates.org:ID"); ok {
		t.Error("expected ok to be false")
	}

	if err := dm.CacheDataFromDB(context.Background(), utils.ActionProfilePrefix, []string{utils.MetaAny}, false); err != nil {
		t.Error(err)
	}

	if rcv, ok := Cache.Get(utils.CacheActionProfiles, "cgrates.org:ID"); !ok {
		t.Error("expected ok to be true")
	} else if !reflect.DeepEqual(rcv, ap) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", ap, rcv)
	}

}

func TestDMCacheDataFromDBAttributeFilterIndexes(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	indexes := map[string]utils.StringSet{"*string:*req.Account:1002": {"ATTR1": {}, "ATTR2": {}}}

	if err := dm.SetIndexes(context.Background(), utils.CacheAttributeFilterIndexes, "cgrates.org", indexes, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	if _, ok := Cache.Get(utils.CacheAttributeFilterIndexes, utils.ConcatenatedKey("cgrates.org", "*string:*req.Account:1002")); ok {
		t.Error("expected ok to be false")
	}

	if err := dm.CacheDataFromDB(context.Background(), utils.AttributeFilterIndexes, []string{utils.MetaAny}, false); err != nil {
		t.Error(err)
	}

	exp := utils.StringSet{"ATTR1": {}, "ATTR2": {}}

	if rcv, ok := Cache.Get(utils.CacheAttributeFilterIndexes, utils.ConcatenatedKey("cgrates.org", "*string:*req.Account:1002")); !ok {
		t.Error("expected ok to be true")
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", exp, rcv)
	}

}

func TestDMCacheDataFromDBResourceFilterIndexes(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	indexes := map[string]utils.StringSet{"*string:*req.Account:1002": {"ATTR1": {}, "ATTR2": {}}}

	if err := dm.SetIndexes(context.Background(), utils.CacheResourceFilterIndexes, "cgrates.org", indexes, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	if _, ok := Cache.Get(utils.CacheResourceFilterIndexes, utils.ConcatenatedKey("cgrates.org", "*string:*req.Account:1002")); ok {
		t.Error("expected ok to be false")
	}

	if err := dm.CacheDataFromDB(context.Background(), utils.ResourceFilterIndexes, []string{utils.MetaAny}, false); err != nil {
		t.Error(err)
	}

	exp := utils.StringSet{"ATTR1": {}, "ATTR2": {}}

	if rcv, ok := Cache.Get(utils.CacheResourceFilterIndexes, utils.ConcatenatedKey("cgrates.org", "*string:*req.Account:1002")); !ok {
		t.Error("expected ok to be true")
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", exp, rcv)
	}

}

func TestDMCacheDataFromDBStatFilterIndexes(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	indexes := map[string]utils.StringSet{"*string:*req.Account:1002": {"ATTR1": {}, "ATTR2": {}}}

	if err := dm.SetIndexes(context.Background(), utils.CacheStatFilterIndexes, "cgrates.org", indexes, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	if _, ok := Cache.Get(utils.CacheStatFilterIndexes, utils.ConcatenatedKey("cgrates.org", "*string:*req.Account:1002")); ok {
		t.Error("expected ok to be false")
	}

	if err := dm.CacheDataFromDB(context.Background(), utils.StatFilterIndexes, []string{utils.MetaAny}, false); err != nil {
		t.Error(err)
	}

	exp := utils.StringSet{"ATTR1": {}, "ATTR2": {}}

	if rcv, ok := Cache.Get(utils.CacheStatFilterIndexes, utils.ConcatenatedKey("cgrates.org", "*string:*req.Account:1002")); !ok {
		t.Error("expected ok to be true")
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", exp, rcv)
	}

}

func TestDMCacheDataFromDBThresholdFilterIndexes(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	indexes := map[string]utils.StringSet{"*string:*req.Account:1002": {"ATTR1": {}, "ATTR2": {}}}

	if err := dm.SetIndexes(context.Background(), utils.CacheThresholdFilterIndexes, "cgrates.org", indexes, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	if _, ok := Cache.Get(utils.CacheThresholdFilterIndexes, utils.ConcatenatedKey("cgrates.org", "*string:*req.Account:1002")); ok {
		t.Error("expected ok to be false")
	}

	if err := dm.CacheDataFromDB(context.Background(), utils.ThresholdFilterIndexes, []string{utils.MetaAny}, false); err != nil {
		t.Error(err)
	}

	exp := utils.StringSet{"ATTR1": {}, "ATTR2": {}}

	if rcv, ok := Cache.Get(utils.CacheThresholdFilterIndexes, utils.ConcatenatedKey("cgrates.org", "*string:*req.Account:1002")); !ok {
		t.Error("expected ok to be true")
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", exp, rcv)
	}

}

func TestDMCacheDataFromDBRouteFilterIndexes(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	indexes := map[string]utils.StringSet{"*string:*req.Account:1002": {"ATTR1": {}, "ATTR2": {}}}

	if err := dm.SetIndexes(context.Background(), utils.CacheRouteFilterIndexes, "cgrates.org", indexes, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	if _, ok := Cache.Get(utils.CacheRouteFilterIndexes, utils.ConcatenatedKey("cgrates.org", "*string:*req.Account:1002")); ok {
		t.Error("expected ok to be false")
	}

	if err := dm.CacheDataFromDB(context.Background(), utils.RouteFilterIndexes, []string{utils.MetaAny}, false); err != nil {
		t.Error(err)
	}

	exp := utils.StringSet{"ATTR1": {}, "ATTR2": {}}

	if rcv, ok := Cache.Get(utils.CacheRouteFilterIndexes, utils.ConcatenatedKey("cgrates.org", "*string:*req.Account:1002")); !ok {
		t.Error("expected ok to be true")
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", exp, rcv)
	}

}

func TestDMCacheDataFromDBChargerFilterIndexes(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	indexes := map[string]utils.StringSet{"*string:*req.Account:1002": {"ATTR1": {}, "ATTR2": {}}}

	if err := dm.SetIndexes(context.Background(), utils.CacheChargerFilterIndexes, "cgrates.org", indexes, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	if _, ok := Cache.Get(utils.CacheChargerFilterIndexes, utils.ConcatenatedKey("cgrates.org", "*string:*req.Account:1002")); ok {
		t.Error("expected ok to be false")
	}

	if err := dm.CacheDataFromDB(context.Background(), utils.ChargerFilterIndexes, []string{utils.MetaAny}, false); err != nil {
		t.Error(err)
	}

	exp := utils.StringSet{"ATTR1": {}, "ATTR2": {}}

	if rcv, ok := Cache.Get(utils.CacheChargerFilterIndexes, utils.ConcatenatedKey("cgrates.org", "*string:*req.Account:1002")); !ok {
		t.Error("expected ok to be true")
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", exp, rcv)
	}

}

func TestDMCacheDataFromDBRateProfilesFilterIndexPrfx(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	indexes := map[string]utils.StringSet{"*string:*req.Account:1002": {"ATTR1": {}, "ATTR2": {}}}

	if err := dm.SetIndexes(context.Background(), utils.CacheRateProfilesFilterIndexes, "cgrates.org", indexes, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	if _, ok := Cache.Get(utils.CacheRateProfilesFilterIndexes, utils.ConcatenatedKey("cgrates.org", "*string:*req.Account:1002")); ok {
		t.Error("expected ok to be false")
	}

	if err := dm.CacheDataFromDB(context.Background(), utils.RateProfilesFilterIndexPrfx, []string{utils.MetaAny}, false); err != nil {
		t.Error(err)
	}

	exp := utils.StringSet{"ATTR1": {}, "ATTR2": {}}

	if rcv, ok := Cache.Get(utils.CacheRateProfilesFilterIndexes, utils.ConcatenatedKey("cgrates.org", "*string:*req.Account:1002")); !ok {
		t.Error("expected ok to be true")
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", exp, rcv)
	}

}

func TestDMCacheDataFromDBRateFilterIndexPrfx(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	indexes := map[string]utils.StringSet{"*string:*req.Account:1002": {"ATTR1": {}, "ATTR2": {}}}

	if err := dm.SetIndexes(context.Background(), utils.CacheRateFilterIndexes, "cgrates.org", indexes, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	if _, ok := Cache.Get(utils.CacheRateFilterIndexes, utils.ConcatenatedKey("cgrates.org", "*string:*req.Account:1002")); ok {
		t.Error("expected ok to be false")
	}

	if err := dm.CacheDataFromDB(context.Background(), utils.RateFilterIndexPrfx, []string{utils.MetaAny}, false); err != nil {
		t.Error(err)
	}

	exp := utils.StringSet{"ATTR1": {}, "ATTR2": {}}

	if rcv, ok := Cache.Get(utils.CacheRateFilterIndexes, utils.ConcatenatedKey("cgrates.org", "*string:*req.Account:1002")); !ok {
		t.Error("expected ok to be true")
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", exp, rcv)
	}

}

func TestDMCacheDataFromDBActionProfilesFilterIndexPrfx(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	indexes := map[string]utils.StringSet{"*string:*req.Account:1002": {"ATTR1": {}, "ATTR2": {}}}

	if err := dm.SetIndexes(context.Background(), utils.CacheActionProfilesFilterIndexes, "cgrates.org", indexes, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	if _, ok := Cache.Get(utils.CacheActionProfilesFilterIndexes, utils.ConcatenatedKey("cgrates.org", "*string:*req.Account:1002")); ok {
		t.Error("expected ok to be false")
	}

	if err := dm.CacheDataFromDB(context.Background(), utils.ActionProfilesFilterIndexPrfx, []string{utils.MetaAny}, false); err != nil {
		t.Error(err)
	}

	exp := utils.StringSet{"ATTR1": {}, "ATTR2": {}}

	if rcv, ok := Cache.Get(utils.CacheActionProfilesFilterIndexes, utils.ConcatenatedKey("cgrates.org", "*string:*req.Account:1002")); !ok {
		t.Error("expected ok to be true")
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", exp, rcv)
	}

}

func TestDMCacheDataFromDBFilterIndexPrfx(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	indexes := map[string]utils.StringSet{"*string:*req.Account:1002": {"ATTR1": {}, "ATTR2": {}}}

	if err := dm.SetIndexes(context.Background(), utils.CacheReverseFilterIndexes, "cgrates.org", indexes, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	if _, ok := Cache.Get(utils.CacheReverseFilterIndexes, utils.ConcatenatedKey("cgrates.org", "*string:*req.Account:1002")); ok {
		t.Error("expected ok to be false")
	}

	if err := dm.CacheDataFromDB(context.Background(), utils.FilterIndexPrfx, []string{utils.MetaAny}, false); err != nil {
		t.Error(err)
	}

	exp := utils.StringSet{"ATTR1": {}, "ATTR2": {}}

	if rcv, ok := Cache.Get(utils.CacheReverseFilterIndexes, utils.ConcatenatedKey("cgrates.org", "*string:*req.Account:1002")); !ok {
		t.Error("expected ok to be true")
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", exp, rcv)
	}

}

func TestDMCacheDataFromDBAttributeFilterIndexErr(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	errExp := "WRONG_IDX_KEY_FORMAT<tntCtx:*prefix:~*accounts>"
	if err := dm.CacheDataFromDB(context.Background(), utils.AttributeFilterIndexes, []string{"tntCtx:*prefix:~*accounts"}, false); errExp != err.Error() {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}
}

func TestDMCacheDataFromDBResourceFilterIndexErr(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	errExp := "WRONG_IDX_KEY_FORMAT<tntCtx:*prefix:~*accounts>"
	if err := dm.CacheDataFromDB(context.Background(), utils.ResourceFilterIndexes, []string{"tntCtx:*prefix:~*accounts"}, false); errExp != err.Error() {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}
}

func TestDMCacheDataFromDBStatFilterIndexErr(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	errExp := "WRONG_IDX_KEY_FORMAT<tntCtx:*prefix:~*accounts>"
	if err := dm.CacheDataFromDB(context.Background(), utils.StatFilterIndexes, []string{"tntCtx:*prefix:~*accounts"}, false); errExp != err.Error() {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}
}

func TestDMCacheDataFromDBThresholdFilterIndexesErr(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	errExp := "WRONG_IDX_KEY_FORMAT<tntCtx:*prefix:~*accounts>"
	if err := dm.CacheDataFromDB(context.Background(), utils.ThresholdFilterIndexes, []string{"tntCtx:*prefix:~*accounts"}, false); errExp != err.Error() {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}
}

func TestDMCacheDataFromDBRouteFilterIndexesErr(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	errExp := "WRONG_IDX_KEY_FORMAT<tntCtx:*prefix:~*accounts>"
	if err := dm.CacheDataFromDB(context.Background(), utils.RouteFilterIndexes, []string{"tntCtx:*prefix:~*accounts"}, false); errExp != err.Error() {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}
}

func TestDMCacheDataFromDBChargerFilterIndexesErr(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	errExp := "WRONG_IDX_KEY_FORMAT<tntCtx:*prefix:~*accounts>"
	if err := dm.CacheDataFromDB(context.Background(), utils.ChargerFilterIndexes, []string{"tntCtx:*prefix:~*accounts"}, false); errExp != err.Error() {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}
}

func TestDMCacheDataFromDBRateProfilesFilterIndexPrfxErr(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	errExp := "WRONG_IDX_KEY_FORMAT<tntCtx:*prefix:~*accounts>"
	if err := dm.CacheDataFromDB(context.Background(), utils.RateProfilesFilterIndexPrfx, []string{"tntCtx:*prefix:~*accounts"}, false); errExp != err.Error() {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}
}

func TestDMCacheDataFromDBRateFilterIndexPrfxErr(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	errExp := "WRONG_IDX_KEY_FORMAT<tntCtx:*prefix:~*accounts>"
	if err := dm.CacheDataFromDB(context.Background(), utils.RateFilterIndexPrfx, []string{"tntCtx:*prefix:~*accounts"}, false); errExp != err.Error() {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}
}

func TestDMCacheDataFromDBActionProfilesFilterIndexPrfxErr(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	errExp := "WRONG_IDX_KEY_FORMAT<tntCtx:*prefix:~*accounts>"
	if err := dm.CacheDataFromDB(context.Background(), utils.ActionProfilesFilterIndexPrfx, []string{"tntCtx:*prefix:~*accounts"}, false); errExp != err.Error() {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}
}

func TestDMGetAccountNil(t *testing.T) {

	var dm *DataManager

	expErr := utils.ErrNoDatabaseConn
	if _, err := dm.GetAccount(context.Background(), utils.CGRateSorg, "1002"); err != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}
}

func TestDMGetAccountReplicate(t *testing.T) {
	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaAccounts].Remote = true
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg)}
	config.SetCgrConfig(cfg)
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)

	cc := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetAccount: func(ctx *context.Context, args, reply any) error {
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- cc

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg), utils.ReplicatorSv1, rpcInternal)
	dm := NewDataManager(data, cfg, cM)

	ap := &utils.Account{
		Tenant: "cgrates.org",
		ID:     "1002",
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
		Opts:         make(map[string]any),
		ThresholdIDs: []string{utils.MetaNone},
	}

	dm.dataDB = &DataDBMock{
		GetAccountDrvF: func(ctx *context.Context, str1, str2 string) (*utils.Account, error) {
			return ap, utils.ErrNotFound
		},
		SetAccountDrvF: func(ctx *context.Context, profile *utils.Account) error {
			return nil
		},
	}

	// tests replicate
	if rcv, err := dm.GetAccount(context.Background(), utils.CGRateSorg, "1002"); err != nil {
		t.Error(err, rcv)
	} else if rcv != ap {
		t.Errorf("Expected <%v>, received <%v>", ap, rcv)
	}
}

func TestDMGetRateProfileRatesNil(t *testing.T) {

	var dm *DataManager

	expErr := utils.ErrNoDatabaseConn
	if _, _, err := dm.GetRateProfileRates(context.Background(), &utils.ArgsSubItemIDs{}, false); err != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}
}

func TestDMGetRateProfileRatesOK(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	rps := &utils.RateProfile{
		ID:        "test_ID1",
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.Destination:1234"},
		Rates: map[string]*utils.Rate{
			"RT1": {
				ID: "RT1",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(1, 2),
						Unit:          utils.NewDecimal(int64(time.Second), 0),
						Increment:     utils.NewDecimal(int64(time.Second), 0),
					},
				},
			},
		},
	}

	dm.DataDB().SetRateProfileDrv(context.Background(), rps, true)

	args := &utils.ArgsSubItemIDs{
		Tenant:      "cgrates.org",
		ProfileID:   "test_ID1",
		ItemsPrefix: "RT1",
	}

	exp := []*utils.Rate{
		{
			ID: "RT1",
			IntervalRates: []*utils.IntervalRate{
				{
					IntervalStart: utils.NewDecimal(0, 0),
					RecurrentFee:  utils.NewDecimal(1, 2),
					Unit:          utils.NewDecimal(int64(time.Second), 0),
					Increment:     utils.NewDecimal(int64(time.Second), 0),
				},
			},
		},
	}

	if _, rcv, err := dm.GetRateProfileRates(context.Background(), args, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(utils.ToJSON(exp), utils.ToJSON(rcv)) {
		t.Errorf("Expected \n<%+v>,\nreceived \n<%+v>", exp, rcv)
	}
}

func TestDMSetLoadIDsNil(t *testing.T) {

	var dm *DataManager

	expErr := utils.ErrNoDatabaseConn
	if err := dm.SetLoadIDs(context.Background(), map[string]int64{}); err != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestDMSetLoadIDsDrvErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	dm.dataDB = &DataDBMock{
		SetLoadIDsDrvF: func(ctx *context.Context, loadIDs map[string]int64) error { return utils.ErrNotImplemented },
	}

	itmLIDs := map[string]int64{
		"ID_1": 21,
	}

	expErr := utils.ErrNotImplemented
	if err := dm.SetLoadIDs(context.Background(), itmLIDs); err != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestDMSetLoadIDsReplicate(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaLoadIDs].Replicate = true
	config.SetCgrConfig(cfg)
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	itmLIDs := map[string]int64{
		"ID_1": 21,
	}

	// tests Replicate
	dm.SetLoadIDs(context.Background(), itmLIDs)

}

func TestDMCheckFiltersErrBadReference(t *testing.T) {

	var dm *DataManager

	expErr := "broken reference to filter: <*string:~*req.Account>"
	if err := dm.checkFilters(context.Background(), utils.CGRateSorg, []string{"*string:~*req.Account"}); expErr != err.Error() {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestDMCheckFiltersErrBadPath(t *testing.T) {

	var dm *DataManager

	expErr := `Path is missing  for filter <{"Tenant":"cgrates.org","ID":"*string:~missing path:chp1","Rules":[{"Type":"*string","Element":"~missing path","Values":["chp1"]}]}>`
	if err := dm.checkFilters(context.Background(), utils.CGRateSorg, []string{"*string:~missing path:chp1"}); expErr != err.Error() {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestDMCheckFiltersErrBrokenReferenceCache(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	var val any
	if err := Cache.Set(context.Background(), utils.CacheFilters, utils.ConcatenatedKey(utils.CGRateSorg, "fltr1"), val, []string{}, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	expErr := `broken reference to filter: <fltr1>`
	if err := dm.checkFilters(context.Background(), utils.CGRateSorg, []string{"fltr1"}); expErr != err.Error() {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestDMCheckFiltersErrCall(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaFilters].Remote = true
	config.SetCgrConfig(cfg)
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	dm.dataDB = &DataDBMock{
		HasDataDrvF: func(ctx *context.Context, category, subject, tenant string) (bool, error) {
			return false, utils.ErrNotFound
		},
	}

	expErr := `broken reference to filter: <fltr1>`
	if err := dm.checkFilters(context.Background(), utils.CGRateSorg, []string{"fltr1"}); expErr != err.Error() {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestGetAPIBanErrSingleCacheWrite(t *testing.T) {
	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.APIBanCfg().Keys = []string{"testKey"}
	cfg.CacheCfg().Partitions[utils.MetaAPIBan].Replicate = true
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	config.SetCgrConfig(cfg)

	var counter int
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		responses := map[string]struct {
			code int
			body []byte
		}{
			"/testKey/check/1.2.3.251": {code: http.StatusOK, body: []byte(`{"ipaddress":["1.2.3.251"], "ID":"987654321"}`)},
		}
		if val, has := responses[r.URL.EscapedPath()]; has {
			w.WriteHeader(val.code)
			if val.body != nil {
				w.Write(val.body)
			}
			return
		}
		counter++
		w.WriteHeader(http.StatusOK)
		if counter < 2 {
			_, _ = w.Write([]byte(`{"ipaddress": ["1.2.3.251", "ID": "100"}`))
		} else {
			_, _ = w.Write([]byte(`{"ID": "none"}`))
			counter = 0
		}
	}))
	defer testServer.Close()
	baningo.RootURL = testServer.URL + "/"

	config.SetCgrConfig(cfg)
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error {

				return utils.ErrNotImplemented
			},
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.CacheSv1, cc)
	dm := NewDataManager(data, cfg, cM)
	Cache = NewCacheS(cfg, dm, cM, nil)

	if _, err := GetAPIBan(context.Background(), "1.2.3.251", []string{"testKey"}, true, false, true); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotImplemented, err)
	}
}

func TestGetAPIBanErrMultipleCacheWrite(t *testing.T) {
	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.APIBanCfg().Keys = []string{"testKey"}
	cfg.CacheCfg().Partitions[utils.MetaAPIBan].Replicate = true
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	config.SetCgrConfig(cfg)

	var counter int
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		responses := map[string]struct {
			code int
			body []byte
		}{
			"/testKey/check/1.2.3.251": {code: http.StatusOK, body: []byte(`{"ipaddress":["1.2.3.251"], "ID":"987654321"}`)},
		}
		if val, has := responses[r.URL.EscapedPath()]; has {
			w.WriteHeader(val.code)
			if val.body != nil {
				w.Write(val.body)
			}
			return
		}
		counter++
		w.WriteHeader(http.StatusOK)
		if counter < 2 {
			_, _ = w.Write([]byte(`{"ipaddress": ["1.2.3.251", "1.2.3.252"], "ID": "100"}`))
		} else {
			_, _ = w.Write([]byte(`{"ID": "none"}`))
			counter = 0
		}
	}))
	defer testServer.Close()
	baningo.RootURL = testServer.URL + "/"

	config.SetCgrConfig(cfg)
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error {

				return utils.ErrNotImplemented
			},
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.CacheSv1, cc)
	dm := NewDataManager(data, cfg, cM)
	Cache = NewCacheS(cfg, dm, cM, nil)

	if _, err := GetAPIBan(context.Background(), "1.2.3.251", []string{"testKey"}, false, false, true); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotImplemented, err)
	}
}

func TestGetAPIBanErrNoBanCacheSet(t *testing.T) {
	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.APIBanCfg().Keys = []string{"testKey"}
	cfg.CacheCfg().Partitions[utils.MetaAPIBan].Replicate = true
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	config.SetCgrConfig(cfg)

	var counter int
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		responses := map[string]struct {
			code int
			body []byte
		}{
			"/testKey/check/1.2.3.251": {code: http.StatusOK, body: []byte(`{"ipaddress":["1.2.3.251"], "ID":"987654321"}`)},
			"/testKey/check/1.2.3.254": {code: http.StatusBadRequest, body: []byte(`{"ipaddress":["not blocked"], "ID":"none"}`)},
		}
		if val, has := responses[r.URL.EscapedPath()]; has {
			w.WriteHeader(val.code)
			if val.body != nil {
				w.Write(val.body)
			}
			return
		}
		counter++
		w.WriteHeader(http.StatusOK)
		if counter < 2 {
			_, _ = w.Write([]byte(`{"ipaddress": ["1.2.3.251", "1.2.3.252"], "ID": "100"}`))
		} else {
			_, _ = w.Write([]byte(`{"ID": "none"}`))
			counter = 0
		}
	}))
	defer testServer.Close()
	baningo.RootURL = testServer.URL + "/"

	config.SetCgrConfig(cfg)
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error {

				return utils.ErrNotImplemented
			},
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.CacheSv1, cc)
	dm := NewDataManager(data, cfg, cM)
	Cache = NewCacheS(cfg, dm, cM, nil)

	if _, err := GetAPIBan(context.Background(), "1.2.3.254", []string{}, false, false, true); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotImplemented, err)
	}
}

func TestDMRemoveIndexesErrNilDm(t *testing.T) {

	var dm *DataManager

	if err := dm.RemoveIndexes(context.Background(), "indxItmtype", "cgrates.org", "indxkey"); err != utils.ErrNoDatabaseConn {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNoDatabaseConn, err)
	}

}

func TestDMRemoveIndexesErrDrv(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	dm.dataDB = &DataDBMock{
		RemoveIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey string) error {
			return utils.ErrNotImplemented
		},
	}

	if err := dm.RemoveIndexes(context.Background(), "indxItmtype", "cgrates.org", "indxkey"); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestDMRemoveIndexesReplicate(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.CacheAttributeFilterIndexes].Replicate = true
	config.SetCgrConfig(cfg)

	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	indexes := map[string]utils.StringSet{"*string:*req.Account:1002": {"ATTR1": {}, "ATTR2": {}}}

	if err := dm.dataDB.SetIndexesDrv(context.Background(), utils.CacheAttributeFilterIndexes, "cgrates.org", indexes, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	_, err := dm.dataDB.GetIndexesDrv(context.Background(), utils.CacheAttributeFilterIndexes, "cgrates.org", utils.EmptyString, utils.NonTransactional)
	if err != nil {
		t.Error(err)
	}

	if err := dm.RemoveIndexes(context.Background(), utils.CacheAttributeFilterIndexes, "cgrates.org", utils.EmptyString); err != nil {
		t.Errorf("Expected error <%v>, received error <%v>", nil, err)
	}

	_, err = dm.dataDB.GetIndexesDrv(context.Background(), utils.CacheAttributeFilterIndexes, "cgrates.org", utils.EmptyString, utils.NonTransactional)
	if err != utils.ErrNotFound {
		t.Error(err)
	}

}

func TestDMSetIndexesErrNilDm(t *testing.T) {

	var dm *DataManager

	indexes := map[string]utils.StringSet{"*string:*req.Account:1002": {"ATTR1": {}, "ATTR2": {}}}

	if err := dm.SetIndexes(context.Background(), utils.CacheAttributeFilterIndexes, utils.CGRateSorg, indexes, true, utils.NonTransactional); err != utils.ErrNoDatabaseConn {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNoDatabaseConn, err)
	}

}

func TestDMSetIndexesReplicate(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.CacheAttributeFilterIndexes].Replicate = true
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	config.SetCgrConfig(cfg)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1SetIndexes: func(ctx *context.Context, args, reply any) error {
				return nil

			},
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.ReplicatorSv1, cc)

	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg, cM)

	indexes := map[string]utils.StringSet{"*string:*req.Account:1002": {"ATTR1": {}, "ATTR2": {}}}

	if err := dm.SetIndexes(context.Background(), utils.CacheAttributeFilterIndexes, utils.CGRateSorg, indexes, true, utils.NonTransactional); err != nil {
		t.Errorf("Expected error <%v>, received error <%v>", nil, err)
	}
	if _, err := dm.GetIndexes(context.Background(), utils.CacheAttributeFilterIndexes, utils.CGRateSorg, utils.EmptyString, utils.NonTransactional, true, false); err != nil {
		t.Error(err)
	}

}

func TestDMGetIndexesErrSetIdxDrv(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.CacheAttributeFilterIndexes].Remote = true
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg)}
	config.SetCgrConfig(cfg)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetIndexes: func(ctx *context.Context, args, reply any) error {
				return nil

			},
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg), utils.ReplicatorSv1, cc)

	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg, cM)

	indexes2 := map[string]utils.StringSet{"*string:*req.Account:1002": {"ATTR1": {}, "ATTR2": {}}}

	dm.dataDB = &DataDBMock{
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return indexes2, utils.ErrNotFound
		},

		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return utils.ErrNotFound
		},
	}

	if _, err := dm.GetIndexes(context.Background(), utils.CacheAttributeFilterIndexes, utils.CGRateSorg, "idxKey", utils.NonTransactional, false, true); err != utils.ErrNotFound {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotFound, err)
	}

}

func TestDMGetIndexesErrCacheSet(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheAttributeFilterIndexes].Replicate = true
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	config.SetCgrConfig(cfg)
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error {

				return utils.ErrNotImplemented
			},
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.CacheSv1, cc)
	dm := NewDataManager(data, cfg, cM)
	Cache = NewCacheS(cfg, dm, cM, nil)

	if _, err := dm.GetIndexes(context.Background(), utils.CacheAttributeFilterIndexes, utils.CGRateSorg, "idxKey", utils.NonTransactional, false, true); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestDMGetIndexesErrCacheWriteSet(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheAttributeFilterIndexes].Replicate = true
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	config.SetCgrConfig(cfg)
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error {

				return utils.ErrNotImplemented
			},
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.CacheSv1, cc)
	dm := NewDataManager(data, cfg, cM)
	Cache = NewCacheS(cfg, dm, cM, nil)

	indexes := map[string]utils.StringSet{"*string:*req.Account:1002": {"ATTR1": {}, "ATTR2": {}}}

	if err := dm.SetIndexes(context.Background(), utils.CacheAttributeFilterIndexes, utils.CGRateSorg, indexes, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	if _, err := dm.GetIndexes(context.Background(), utils.CacheAttributeFilterIndexes, utils.CGRateSorg, utils.EmptyString, utils.NonTransactional, false, true); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestDMRemoveActionProfileErrNilDM(t *testing.T) {

	var dm *DataManager

	if err := dm.RemoveActionProfile(context.Background(), "cgrates.org", "AP1", false); err != utils.ErrNoDatabaseConn {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNoDatabaseConn, err)
	}

}

func TestDMRemoveActionProfileErrGetActionProf(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	dm.dataDB = &DataDBMock{
		GetActionProfileDrvF: func(ctx *context.Context, tenant, ID string) (*utils.ActionProfile, error) {
			return nil, utils.ErrNotImplemented
		},
	}

	if err := dm.RemoveActionProfile(context.Background(), "cgrates.org", "AP1", false); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestDMRemoveActionProfileErrRemvProfDrv(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	ap := &utils.ActionProfile{

		Tenant:    "cgrates.org",
		ID:        "AP1",
		FilterIDs: []string{"fltr1"},
		Weights: utils.DynamicWeights{
			{
				Weight: 65,
			},
		},
		Schedule: "* * * * *",
		Targets:  map[string]utils.StringSet{utils.MetaAccounts: {"1001": {}}},
		Actions:  []*utils.APAction{{}},
	}

	dm.dataDB = &DataDBMock{
		GetActionProfileDrvF: func(ctx *context.Context, tenant, ID string) (*utils.ActionProfile, error) {
			return ap, nil
		},
		RemoveActionProfileDrvF: func(ctx *context.Context, tenant, ID string) error {
			return utils.ErrNotImplemented
		},
	}

	if err := dm.RemoveActionProfile(context.Background(), "cgrates.org", "AP1", false); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestDMCacheDataFromDBPrefixKeysErr(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	dm.dataDB = &DataDBMock{
		GetKeysForPrefixF: func(ctx *context.Context, s string) ([]string, error) { return []string{}, utils.ErrNotImplemented },
	}

	if err := dm.CacheDataFromDB(context.Background(), utils.FilterIndexPrfx, []string{utils.MetaAny}, false); err.Error() != utils.ErrNotImplemented.Error() {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestDMGetFilterCacheReadGetErr(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	if err := Cache.Set(context.Background(), utils.CacheFilters, utils.ConcatenatedKey(utils.CGRateSorg, "fltr_for_prf"), nil, []string{}, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	if _, err := dm.GetFilter(context.Background(), "cgrates.org", "fltr_for_prf", true, false, utils.GenUUID()); err != utils.ErrNotFound {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotFound, err)
	}

}

func TestDMGetFilterNilDMErr(t *testing.T) {

	var dm *DataManager

	if _, err := dm.GetFilter(context.Background(), "cgrates.org", "fltr_for_prf", false, false, utils.GenUUID()); err != utils.ErrNoDatabaseConn {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNoDatabaseConn, err)
	}

}

func TestDMGetThresholdSetThPrflDrvErr(t *testing.T) {
	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaThresholdProfiles].Remote = true
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg)}
	config.SetCgrConfig(cfg)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetThresholdProfile: func(ctx *context.Context, args, reply any) error { return nil },
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg), utils.ReplicatorSv1, cc)

	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg, cM)

	th := &ThresholdProfile{
		Tenant:           "cgrates.org",
		ID:               "THD_100",
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
	}

	dm.dataDB = &DataDBMock{
		GetThresholdProfileDrvF: func(ctx *context.Context, tenant, id string) (tp *ThresholdProfile, err error) {
			return th, utils.ErrNotFound
		},
	}

	if _, err := dm.GetThresholdProfile(context.Background(), "cgrates.org", "THD_100", false, false, utils.NonTransactional); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestDMSetThresholdProfileNilDM(t *testing.T) {

	var dm *DataManager
	th := &ThresholdProfile{
		Tenant:           "cgrates.org",
		ID:               "THD_100",
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
	}
	if err := dm.SetThresholdProfile(context.Background(), th, false); err != utils.ErrNoDatabaseConn {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNoDatabaseConn, err)
	}

}

func TestDMSetThresholdProfileWithIndexErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	th := &ThresholdProfile{
		Tenant:           "cgrates.org",
		ID:               "THD_100",
		FilterIDs:        []string{"*string*req.Account1001"},
		ActionProfileIDs: []string{"actPrfID"},
		MaxHits:          7,
		MinHits:          0,
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		Async: true,
	}

	expErr := "broken reference to filter: <*string*req.Account1001> for item with ID: cgrates.org:THD_100"
	if err := dm.SetThresholdProfile(context.Background(), th, true); err.Error() != expErr {
		t.Errorf("Expected error <%v>, received error <%v>", expErr, err)
	}

}

func TestDMSetThresholdProfileGetThPrfErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	dm.dataDB = &DataDBMock{
		GetThresholdProfileDrvF: func(ctx *context.Context, tenant, id string) (tp *ThresholdProfile, err error) {
			return nil, utils.ErrNotImplemented
		},
	}

	th := &ThresholdProfile{
		Tenant:           "cgrates.org",
		ID:               "THD_100",
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
	}

	if err := dm.SetThresholdProfile(context.Background(), th, false); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestDMSetThresholdProfileSetThPrflDrvErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	th := &ThresholdProfile{
		Tenant:           "cgrates.org",
		ID:               "THD_100",
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
	}

	dm.dataDB = &DataDBMock{
		GetThresholdProfileDrvF: func(ctx *context.Context, tenant, id string) (tp *ThresholdProfile, err error) {
			return th, nil
		},
		SetThresholdProfileDrvF: func(ctx *context.Context, tp *ThresholdProfile) (err error) { return utils.ErrNotFound },
	}

	if err := dm.SetThresholdProfile(context.Background(), th, false); err != utils.ErrNotFound {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotFound, err)
	}

}

func TestDMSetThresholdProfileUpdatedIndexesErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	dm.dataDB = &DataDBMock{
		GetThresholdProfileDrvF: func(ctx *context.Context, tenant, id string) (tp *ThresholdProfile, err error) {
			return &ThresholdProfile{
				Tenant: utils.CGRateSorg,
			}, nil
		},
		SetThresholdProfileDrvF: func(ctx *context.Context, tp *ThresholdProfile) (err error) { return nil },
	}

	th := &ThresholdProfile{
		Tenant:           "cgrates.org",
		ID:               "THD_100",
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
	}

	if err := dm.SetThresholdProfile(context.Background(), th, true); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestDMSetThresholdProfileReplicateErr(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaThresholdProfiles].Replicate = true
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	config.SetCgrConfig(cfg)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1SetThresholdProfile: func(ctx *context.Context, args, reply any) error { return utils.ErrNotImplemented },
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.ReplicatorSv1, cc)

	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg, cM)

	th := &ThresholdProfile{
		Tenant:           "cgrates.org",
		ID:               "THD_100",
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
	}

	if err := dm.SetThresholdProfile(context.Background(), th, true); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestDMGetStatQueueCacheGetErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	if err := Cache.Set(context.Background(), utils.CacheStatQueues, utils.ConcatenatedKey(utils.CGRateSorg, "SQ1"), nil, []string{}, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	_, err := dm.GetStatQueue(context.Background(), utils.CGRateSorg, "SQ1", true, false, utils.NonTransactional)
	if err != utils.ErrNotFound {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotFound, err)
	}
}

func TestDMGetStatQueueNewStoredStatQueueErr(t *testing.T) {

	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
		config.SetCgrConfig(cfgtmp)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaStatQueues].Remote = true
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg)}
	config.SetCgrConfig(cfg)

	stq := &StatQueue{
		Tenant: "cgrates.org",
		ID:     "sq01",
		SQMetrics: map[string]StatMetric{
			"key": statMetricMock(""),
		},
		SQItems: []SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
	}

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetStatQueue: func(ctx *context.Context, args, reply any) error {
				return nil
			},
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg), utils.ReplicatorSv1, cc)

	data := &DataDBMock{
		GetStatQueueDrvF: func(ctx *context.Context, tenant, id string) (sq *StatQueue, err error) {
			return stq, utils.ErrNotFound
		},
	}
	dm := NewDataManager(data, cfg, cM)

	experr := "marshal mock error"
	dm.ms = mockMarshal(experr)

	if _, err := dm.GetStatQueue(context.Background(), utils.CGRateSorg, "sq01", false, false, utils.NonTransactional); err == nil || err.Error() != experr {
		t.Errorf("Expected error <%v>, received error <%v>", experr, err)
	}

}

func TestDMGetStatQueueSetStatQueueDrvErr(t *testing.T) {

	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
		config.SetCgrConfig(cfgtmp)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaStatQueues].Remote = true
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg)}
	config.SetCgrConfig(cfg)

	stq := &StatQueue{
		Tenant: "cgrates.org",
		ID:     "sq01",
		SQItems: []SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: make(map[string]StatMetric),
	}

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetStatQueue: func(ctx *context.Context, args, reply any) error {
				return nil
			},
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg), utils.ReplicatorSv1, cc)

	data := &DataDBMock{
		GetStatQueueDrvF: func(ctx *context.Context, tenant, id string) (sq *StatQueue, err error) {
			return stq, utils.ErrNotFound
		},
		SetStatQueueDrvF: func(ctx *context.Context, ssq *StoredStatQueue, sq *StatQueue) error { return utils.ErrNotImplemented },
	}
	dm := NewDataManager(data, cfg, cM)

	if _, err := dm.GetStatQueue(context.Background(), utils.CGRateSorg, "sq01", false, false, utils.NonTransactional); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestDMGetStatQueueCacheWriteErr(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.CacheCfg().Partitions[utils.CacheStatQueues].Replicate = true
	config.SetCgrConfig(cfg)

	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error { return utils.ErrNotImplemented },
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.CacheSv1, cc)

	dm := NewDataManager(data, cfg, cM)

	ssq := &StoredStatQueue{
		Tenant: "cgrates.org",
		ID:     "ssq01",
		SQItems: []SQItem{
			{
				EventID: "testEventID",
			},
		},
		SQMetrics: map[string][]byte{
			utils.MetaTCD: []byte(""),
		},
		Compressed: true,
	}

	sq := &StatQueue{
		Tenant: "cgrates.org",
		ID:     "sq01",
		SQItems: []SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: make(map[string]StatMetric),
	}

	if err := dm.DataDB().SetStatQueueDrv(context.Background(), ssq, sq); err != nil {
		t.Error(err)
	}

	Cache = NewCacheS(cfg, dm, cM, nil)

	if _, err := dm.GetStatQueue(context.Background(), utils.CGRateSorg, "sq01", false, true, utils.NonTransactional); err != utils.ErrNotImplemented {
		t.Errorf("Expected %v\n but received %v", utils.ErrNotImplemented, err)
	}

}

func TestDMCacheDataFromDBAccountFilterIndexPrfx(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	indexes := map[string]utils.StringSet{"*string:*req.Account:1002": {"ATTR1": {}, "ATTR2": {}}}

	if err := dm.SetIndexes(context.Background(), utils.CacheAccountsFilterIndexes, "cgrates.org", indexes, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	if _, ok := Cache.Get(utils.CacheAccountsFilterIndexes, utils.ConcatenatedKey("cgrates.org", "*string:*req.Account:1002")); ok {
		t.Error("expected ok to be false")
	}

	if err := dm.CacheDataFromDB(context.Background(), utils.AccountFilterIndexPrfx, []string{utils.MetaAny}, false); err != nil {
		t.Error(err)
	}

	exp := utils.StringSet{"ATTR1": {}, "ATTR2": {}}

	if rcv, ok := Cache.Get(utils.CacheAccountsFilterIndexes, utils.ConcatenatedKey("cgrates.org", "*string:*req.Account:1002")); !ok {
		t.Error("expected ok to be true")
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", exp, rcv)
	}

}

func TestDMCacheDataFromDBAccountFilterIndexPrfxErr(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	errExp := "WRONG_IDX_KEY_FORMAT<tntCtx:*prefix:~*accounts>"
	if err := dm.CacheDataFromDB(context.Background(), utils.AccountFilterIndexPrfx, []string{"tntCtx:*prefix:~*accounts"}, false); errExp != err.Error() {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}
}

func TestDMSetStatQueueProfileNilDm(t *testing.T) {

	var dm *DataManager

	sqp := &StatQueueProfile{}

	if err := dm.SetStatQueueProfile(context.Background(), sqp, false); err != utils.ErrNoDatabaseConn {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNoDatabaseConn, err)
	}

}

func TestDMSetStatQueueProfileCheckFiltrsErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	sqp := &StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "sqp001",
		FilterIDs:   []string{":::"},
		QueueLength: 10,
		TTL:         10 * time.Second,
		Metrics: []*MetricWithFilters{
			{
				MetricID: "*sum#~*req.Usage",
			},
		},
		ThresholdIDs: []string{},
		Stored:       true,
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		MinItems: 1,
	}

	expErr := "broken reference to filter: <:::> for item with ID: cgrates.org:sqp001"
	if err := dm.SetStatQueueProfile(context.Background(), sqp, true); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, received error <%v>", expErr, err)
	}

}

func TestDMSetStatQueueProfileGetStatQProflErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	sqp := &StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "sqp002",
		FilterIDs:   []string{"*string:~*req.Account:1001"},
		QueueLength: 10,
		TTL:         10 * time.Second,
		Metrics: []*MetricWithFilters{
			{
				MetricID: "*sum#~*req.Usage",
			},
		},
		ThresholdIDs: []string{},
		Stored:       true,
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		MinItems: 1,
	}

	dm.dataDB = &DataDBMock{
		GetStatQueueProfileDrvF: func(ctx *context.Context, tenant, id string) (sq *StatQueueProfile, err error) {
			return sqp, utils.ErrNotImplemented
		},
	}

	expErr := utils.ErrNotImplemented
	if err := dm.SetStatQueueProfile(context.Background(), sqp, false); err != expErr {
		t.Errorf("Expected error <%v>, received error <%v>", expErr, err)
	}

}

func TestDMSetStatQueueProfileSetStatQPrflDrvErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	sqp := &StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "sqp003",
		FilterIDs:   []string{"*string:~*req.Account:1001"},
		QueueLength: 10,
		TTL:         10 * time.Second,
		Metrics: []*MetricWithFilters{
			{
				MetricID: "*sum#~*req.Usage",
			},
		},
		ThresholdIDs: []string{},
		Stored:       true,
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		MinItems: 1,
	}

	dm.dataDB = &DataDBMock{
		GetStatQueueProfileDrvF: func(ctx *context.Context, tenant, id string) (sq *StatQueueProfile, err error) { return sqp, nil },
		SetStatQueueProfileDrvF: func(ctx *context.Context, sq *StatQueueProfile) (err error) { return utils.ErrNotImplemented },
	}

	expErr := utils.ErrNotImplemented
	if err := dm.SetStatQueueProfile(context.Background(), sqp, false); err != expErr {
		t.Errorf("Expected error <%v>, received error <%v>", expErr, err)
	}

}

func TestDMGetResourceCacheGetErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	if err := Cache.Set(context.Background(), utils.CacheResources, utils.ConcatenatedKey(utils.CGRateSorg, "rsrc1"), nil, []string{}, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	_, err := dm.GetResource(context.Background(), utils.CGRateSorg, "rsrc1", true, false, utils.NonTransactional)
	if err != utils.ErrNotFound {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotFound, err)
	}
}

func TestDMGetResourceNilDmErr(t *testing.T) {

	var dm *DataManager

	_, err := dm.GetResource(context.Background(), utils.CGRateSorg, "rsrc1", false, false, utils.NonTransactional)
	if err != utils.ErrNoDatabaseConn {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNoDatabaseConn, err)
	}
}

func TestDMGetResourceSetResourceDrvErr(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaResources].Remote = true
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg)}
	config.SetCgrConfig(cfg)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetResource: func(ctx *context.Context, args, reply any) error {
				return nil
			},
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg), utils.ReplicatorSv1, cc)

	data := &DataDBMock{
		GetResourceDrvF: func(ctx *context.Context, tenant, id string) (*utils.Resource, error) {
			return &utils.Resource{}, utils.ErrNotFound
		},
		SetResourceDrvF: func(ctx *context.Context, r *utils.Resource) error { return utils.ErrNotImplemented },
	}
	dm := NewDataManager(data, cfg, cM)

	if _, err := dm.GetResource(context.Background(), utils.CGRateSorg, "ResGroup2", false, false, utils.NonTransactional); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestDMGetResourceCacheWriteErr1(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.CacheCfg().Partitions[utils.CacheResources].Replicate = true
	config.SetCgrConfig(cfg)

	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error { return utils.ErrNotImplemented },
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.CacheSv1, cc)
	dm := NewDataManager(data, cfg, cM)

	Cache = NewCacheS(cfg, dm, cM, nil)

	if _, err := dm.GetResource(context.Background(), utils.CGRateSorg, "ResGroup2", false, true, utils.NonTransactional); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestDMGetResourceCacheWriteErr2(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.CacheCfg().Partitions[utils.CacheResources].Replicate = true
	config.SetCgrConfig(cfg)

	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error { return utils.ErrNotImplemented },
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.CacheSv1, cc)
	dm := NewDataManager(data, cfg, cM)

	rs := &utils.Resource{
		Tenant: "cgrates.org",
		ID:     "ResGroup2",
		Usages: map[string]*utils.ResourceUsage{
			"RU1": {
				Tenant: "cgrates.org",
				ID:     "RU1",
				Units:  9,
			},
		},
	}

	if err := dm.dataDB.SetResourceDrv(context.Background(), rs); err != nil {
		t.Error(err)
	}

	Cache = NewCacheS(cfg, dm, cM, nil)

	if _, err := dm.GetResource(context.Background(), utils.CGRateSorg, "ResGroup2", false, true, utils.NonTransactional); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestDMSetResourceSetResourceDrvErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		SetResourceDrvF: func(ctx *context.Context, r *utils.Resource) error { return utils.ErrNotImplemented },
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	rs := &utils.Resource{
		Tenant: "cgrates.org",
		ID:     "ResGroup2",
		Usages: map[string]*utils.ResourceUsage{
			"RU1": {
				Tenant: "cgrates.org",
				ID:     "RU1",
				Units:  9,
			},
		},
	}

	if err := dm.SetResource(context.Background(), rs); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestDMSetResourceNilDmErr(t *testing.T) {

	var dm *DataManager

	if err := dm.RemoveResource(context.Background(), "cgrates.org", "ResGroup2"); err != utils.ErrNoDatabaseConn {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNoDatabaseConn, err)
	}

}

func TestDMRemoveResourceRemoveResourceDrvErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		RemoveResourceDrvF: func(ctx *context.Context, tnt, id string) error { return utils.ErrNotImplemented },
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	if err := dm.RemoveResource(context.Background(), "cgrates.org", "ResGroup2"); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestDMRemoveResourceReplicateErr(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaResources].Replicate = true
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	config.SetCgrConfig(cfg)

	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1RemoveResource: func(ctx *context.Context, args, reply any) error {
				return nil
			},
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.ReplicatorSv1, cc)
	dm := NewDataManager(data, cfg, cM)

	// test replicate
	if err := dm.RemoveResource(context.Background(), "cgrates.org", "ResGroup2"); err != nil {
		t.Errorf("Expected error <nil>, received error <%v>", err)
	}

}

func TestDMGetResourceProfileCacheGetErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	if err := Cache.Set(context.Background(), utils.CacheResourceProfiles, utils.ConcatenatedKey(utils.CGRateSorg, "rsrc1"), nil, []string{}, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	_, err := dm.GetResourceProfile(context.Background(), utils.CGRateSorg, "rsrc1", true, false, utils.NonTransactional)
	if err != utils.ErrNotFound {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotFound, err)
	}
}

func TestDMGetResourceProfileNilDmErr(t *testing.T) {

	var dm *DataManager

	_, err := dm.GetResourceProfile(context.Background(), utils.CGRateSorg, "rsrc1", false, false, utils.NonTransactional)
	if err != utils.ErrNoDatabaseConn {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNoDatabaseConn, err)
	}
}

func TestDMGetResourceProfileSetResourceProfileDrvErr(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaResourceProfiles].Remote = true
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg)}
	config.SetCgrConfig(cfg)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetResourceProfile: func(ctx *context.Context, args, reply any) error {
				return nil
			},
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg), utils.ReplicatorSv1, cc)

	data := &DataDBMock{
		GetResourceProfileDrvF: func(ctx *context.Context, tnt, id string) (*utils.ResourceProfile, error) {
			return &utils.ResourceProfile{}, utils.ErrNotFound
		},
		SetResourceProfileDrvF: func(ctx *context.Context, rp *utils.ResourceProfile) error { return utils.ErrNotImplemented },
	}
	dm := NewDataManager(data, cfg, cM)

	if _, err := dm.GetResourceProfile(context.Background(), utils.CGRateSorg, "rsrc1", false, false, utils.NonTransactional); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestDMGetResourceProfileCacheWriteErr1(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.CacheCfg().Partitions[utils.CacheResourceProfiles].Replicate = true
	config.SetCgrConfig(cfg)

	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error { return utils.ErrNotImplemented },
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.CacheSv1, cc)
	dm := NewDataManager(data, cfg, cM)

	Cache = NewCacheS(cfg, dm, cM, nil)

	if _, err := dm.GetResourceProfile(context.Background(), utils.CGRateSorg, "ResGroup2", false, true, utils.NonTransactional); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestDMGetResourceProfileCacheWriteErr2(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.CacheCfg().Partitions[utils.CacheResourceProfiles].Replicate = true
	config.SetCgrConfig(cfg)

	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error { return utils.ErrNotImplemented },
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.CacheSv1, cc)
	dm := NewDataManager(data, cfg, cM)

	rp := &utils.ResourceProfile{
		Tenant:    "cgrates.org",
		ID:        "RL2",
		FilterIDs: []string{"fltr_test"},
		Weights: utils.DynamicWeights{
			{
				Weight: 100,
			}},
		Limit:             2,
		ThresholdIDs:      []string{"TEST_ACTIONS"},
		Blocker:           false,
		UsageTTL:          time.Millisecond,
		AllocationMessage: "ALLOC",
	}
	if err := dm.dataDB.SetResourceProfileDrv(context.Background(), rp); err != nil {
		t.Error(err)
	}

	Cache = NewCacheS(cfg, dm, cM, nil)

	if _, err := dm.GetResourceProfile(context.Background(), utils.CGRateSorg, "RL2", false, true, utils.NonTransactional); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestDMGetFilterSetFilterDrvErr(t *testing.T) {
	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaFilters].Remote = true
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg)}
	config.SetCgrConfig(cfg)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetFilter: func(ctx *context.Context, args, reply any) error { return nil },
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg), utils.ReplicatorSv1, cc)

	data := &DataDBMock{
		GetFilterDrvF: func(ctx *context.Context, str1, str2 string) (*Filter, error) { return &Filter{}, utils.ErrNotFound },
		SetFilterDrvF: func(ctx *context.Context, fltr *Filter) error { return utils.ErrNotImplemented },
	}

	dm := NewDataManager(data, cfg, cM)

	if _, err := dm.GetFilter(context.Background(), "cgrates.org", "fltr2", false, true, utils.NonTransactional); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestDMGetFilterCacheWriteErr1(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.CacheCfg().Partitions[utils.CacheFilters].Replicate = true
	config.SetCgrConfig(cfg)

	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error { return utils.ErrNotImplemented },
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.CacheSv1, cc)
	dm := NewDataManager(data, cfg, cM)

	Cache = NewCacheS(cfg, dm, cM, nil)

	if _, err := dm.GetFilter(context.Background(), utils.CGRateSorg, "fltr2", false, true, utils.NonTransactional); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestDMGetFilterCacheWriteErr2(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.CacheCfg().Partitions[utils.CacheFilters].Replicate = true
	config.SetCgrConfig(cfg)

	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error { return utils.ErrNotImplemented },
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.CacheSv1, cc)
	dm := NewDataManager(data, cfg, cM)

	f := &Filter{
		Tenant: utils.CGRateSorg,
		ID:     "fltr2",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Charger",
				Values:  []string{"ChargerProfile2"},
			},
		},
	}

	if err := dm.dataDB.SetFilterDrv(context.Background(), f); err != nil {
		t.Error(err)
	}

	Cache = NewCacheS(cfg, dm, cM, nil)

	if _, err := dm.GetFilter(context.Background(), utils.CGRateSorg, "fltr2", false, true, utils.NonTransactional); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestDMGetThresholdSetThresholdDrvErr(t *testing.T) {
	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaThresholds].Remote = true
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg)}
	config.SetCgrConfig(cfg)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetThreshold: func(ctx *context.Context, args, reply any) error { return nil },
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg), utils.ReplicatorSv1, cc)

	data := &DataDBMock{
		GetThresholdDrvF: func(ctx *context.Context, tenant, id string) (*Threshold, error) {
			return &Threshold{}, utils.ErrNotFound
		},
		SetThresholdDrvF: func(ctx *context.Context, t *Threshold) error { return utils.ErrNotImplemented },
	}

	dm := NewDataManager(data, cfg, cM)

	if _, err := dm.GetThreshold(context.Background(), "cgrates.org", "TH1", false, false, utils.NonTransactional); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestDMSetResourceProfileNilDm(t *testing.T) {

	var dm *DataManager

	if err := dm.SetResourceProfile(context.Background(), &utils.ResourceProfile{}, false); err != utils.ErrNoDatabaseConn {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNoDatabaseConn, err)
	}

}

func TestDMSetResourceProfileGetResourceProfileErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetResourceProfileDrvF: func(ctx *context.Context, tnt, id string) (*utils.ResourceProfile, error) {
			return &utils.ResourceProfile{}, utils.ErrNotImplemented
		},
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	if err := dm.SetResourceProfile(context.Background(), &utils.ResourceProfile{}, false); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}
func TestDMSetResourceProfileSetResourceProfileDrvErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetResourceProfileDrvF: func(ctx *context.Context, tnt, id string) (*utils.ResourceProfile, error) {
			return &utils.ResourceProfile{}, nil
		},
		SetResourceProfileDrvF: func(ctx *context.Context, rp *utils.ResourceProfile) error { return utils.ErrNotImplemented },
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	if err := dm.SetResourceProfile(context.Background(), &utils.ResourceProfile{}, false); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestDMSetResourceProfileUpdatedIndexesErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetResourceProfileDrvF: func(ctx *context.Context, tnt, id string) (*utils.ResourceProfile, error) {
			return &utils.ResourceProfile{}, nil
		},
		SetResourceProfileDrvF: func(ctx *context.Context, rp *utils.ResourceProfile) error { return nil },
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	rp := &utils.ResourceProfile{
		Tenant:    "cgrates.org",
		ID:        "RL1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 100,
			}},
		Limit:        2,
		ThresholdIDs: []string{"TEST_ACTIONS"},

		UsageTTL:          time.Millisecond,
		AllocationMessage: "ALLOC",
	}

	if err := dm.SetResourceProfile(context.Background(), rp, true); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestDMSetResourceProfileErr(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaResourceProfiles].Replicate = true
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	config.SetCgrConfig(cfg)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1SetResourceProfile: func(ctx *context.Context, args, reply any) error { return utils.ErrNotImplemented },
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.ReplicatorSv1, cc)

	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg, cM)

	if err := dm.SetResourceProfile(context.Background(), &utils.ResourceProfile{}, false); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestDMRemoveResourceProfileNilDm(t *testing.T) {

	var dm *DataManager

	if err := dm.RemoveResourceProfile(context.Background(), utils.CGRateSorg, "RSP1", false); err != utils.ErrNoDatabaseConn {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNoDatabaseConn, err)
	}

}

func TestDMRemoveResourceProfileGetResourceProfileErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetResourceProfileDrvF: func(ctx *context.Context, tnt, id string) (*utils.ResourceProfile, error) {
			return &utils.ResourceProfile{}, utils.ErrNotImplemented
		},
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	if err := dm.RemoveResourceProfile(context.Background(), utils.CGRateSorg, "RSP1", false); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestDMRemoveResourceProfileRemoveResourceProfileDrvErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetResourceProfileDrvF: func(ctx *context.Context, tnt, id string) (*utils.ResourceProfile, error) {
			return &utils.ResourceProfile{}, nil
		},
		RemoveResourceProfileDrvF: func(ctx *context.Context, tnt, id string) error { return utils.ErrNotImplemented },
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	if err := dm.RemoveResourceProfile(context.Background(), utils.CGRateSorg, "RSP1", false); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestDMRemoveResourceProfileOldResErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	if err := dm.RemoveResourceProfile(context.Background(), utils.CGRateSorg, "RSP1", false); err != utils.ErrNotFound {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotFound, err)
	}

}

func TestDMRemoveResourceProfileRemoveItemFromFilterIndexErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	rp := &utils.ResourceProfile{
		Tenant:    "cgrates.org",
		ID:        "RSP1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 100,
			}},
		Limit:        2,
		ThresholdIDs: []string{"TEST_ACTIONS"},

		UsageTTL:          time.Millisecond,
		AllocationMessage: "ALLOC",
	}

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetResourceProfileDrvF: func(ctx *context.Context, tnt, id string) (*utils.ResourceProfile, error) { return rp, nil },
		RemoveResourceProfileDrvF: func(ctx *context.Context, tnt, id string) error {
			return nil
		},
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	if err := dm.RemoveResourceProfile(context.Background(), utils.CGRateSorg, rp.ID, true); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestDMRemoveResourceProfileRemoveIndexFiltersItemErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)
	rp := &utils.ResourceProfile{
		Tenant:    "cgrates.org",
		ID:        "RSP1",
		FilterIDs: []string{"fltrID"},
		Weights: utils.DynamicWeights{
			{
				Weight: 100,
			}},
		Limit:        2,
		ThresholdIDs: []string{"TEST_ACTIONS"},

		UsageTTL:          time.Millisecond,
		AllocationMessage: "ALLOC",
	}

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetResourceProfileDrvF: func(ctx *context.Context, tnt, id string) (*utils.ResourceProfile, error) { return rp, nil },
		RemoveResourceProfileDrvF: func(ctx *context.Context, tnt, id string) error {
			return nil
		},
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	if err := dm.RemoveResourceProfile(context.Background(), utils.CGRateSorg, rp.ID, true); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}
func TestDMRemoveResourceProfileReplicate(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	rp := &utils.ResourceProfile{
		Tenant:    "cgrates.org",
		ID:        "RSP1",
		FilterIDs: []string{"fltrID"},
		Weights: utils.DynamicWeights{
			{
				Weight: 100,
			}},
		Limit:        2,
		ThresholdIDs: []string{"TEST_ACTIONS"},

		UsageTTL:          time.Millisecond,
		AllocationMessage: "ALLOC",
	}

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaResourceProfiles].Replicate = true
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	config.SetCgrConfig(cfg)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1SetResourceProfile: func(ctx *context.Context, args, reply any) error { return utils.ErrNotImplemented },
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.ReplicatorSv1, cc)
	data := &DataDBMock{
		GetResourceProfileDrvF: func(ctx *context.Context, tnt, id string) (*utils.ResourceProfile, error) { return rp, nil },
		RemoveResourceProfileDrvF: func(ctx *context.Context, tnt, id string) error {
			return nil
		},
	}
	dm := NewDataManager(data, cfg, cM)

	// tests replicate
	if err := dm.RemoveResourceProfile(context.Background(), utils.CGRateSorg, rp.ID, false); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestDMHasDataNilDmErr(t *testing.T) {

	var dm *DataManager

	if _, err := dm.HasData("Category", "subj", "cgrates.org"); err != utils.ErrNoDatabaseConn {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNoDatabaseConn, err)
	}

}

func TestDMHasDataOK(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	fltrTh1 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_TH_1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Threshold",
				Values:  []string{"TH_1"},
			},
		},
	}
	if err := dm.SetFilter(context.Background(), fltrTh1, true); err != nil {
		t.Error(err)
	}

	if has, err := dm.HasData(utils.FilterPrefix, fltrTh1.ID, fltrTh1.Tenant); err != nil {
		t.Error(err)
	} else if !has {
		t.Errorf("Expected to have data")
	}

}

func TestDMGetRouteProfileCacheGetErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	if err := Cache.Set(context.Background(), utils.CacheRouteProfiles, utils.ConcatenatedKey(utils.CGRateSorg, "rsrc1"), nil, []string{}, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	_, err := dm.GetRouteProfile(context.Background(), utils.CGRateSorg, "rsrc1", true, false, utils.NonTransactional)
	if err != utils.ErrNotFound {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotFound, err)
	}
}

func TestDMGetRouteProfileNilDmErr(t *testing.T) {

	var dm *DataManager

	_, err := dm.GetRouteProfile(context.Background(), utils.CGRateSorg, "rsrc1", false, false, utils.NonTransactional)
	if err != utils.ErrNoDatabaseConn {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNoDatabaseConn, err)
	}
}

func TestDMGetRouteProfileSetRouteProfileDrvErr(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	rp := &utils.RouteProfile{

		Tenant:            "cgrates.org",
		ID:                "RP_1",
		FilterIDs:         []string{"fltr_test"},
		Weights:           utils.DynamicWeights{{}},
		Sorting:           utils.MetaQOS,
		SortingParameters: []string{"param"},
		Routes: []*utils.Route{{
			ID:             "RT1",
			FilterIDs:      []string{"fltr1"},
			AccountIDs:     []string{"acc1"},
			RateProfileIDs: []string{"rp1"},
			ResourceIDs:    []string{"res1"},
			StatIDs:        []string{"stat1"},
			Weights:        utils.DynamicWeights{{}},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: true,
				},
			},
			RouteParameters: "params",
		}},
	}

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaRouteProfiles].Remote = true
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg)}
	config.SetCgrConfig(cfg)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetRouteProfile: func(ctx *context.Context, args, reply any) error { return nil },
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg), utils.ReplicatorSv1, cc)
	data := &DataDBMock{
		GetRouteProfileDrvF: func(ctx *context.Context, tnt, id string) (*utils.RouteProfile, error) { return rp, utils.ErrNotFound },
		SetRouteProfileDrvF: func(ctx *context.Context, rtPrf *utils.RouteProfile) error { return utils.ErrNotImplemented },
	}
	dm := NewDataManager(data, cfg, cM)

	_, err := dm.GetRouteProfile(context.Background(), utils.CGRateSorg, rp.ID, false, false, utils.NonTransactional)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}
}

func TestDMGetRouteProfileCacheWriteErr1(t *testing.T) {

	rp := &utils.RouteProfile{

		Tenant:            "cgrates.org",
		ID:                "RP_1",
		FilterIDs:         []string{"fltr_test"},
		Weights:           utils.DynamicWeights{{}},
		Sorting:           utils.MetaQOS,
		SortingParameters: []string{"param"},
		Routes: []*utils.Route{{
			ID:             "RT1",
			FilterIDs:      []string{"fltr1"},
			AccountIDs:     []string{"acc1"},
			RateProfileIDs: []string{"rp1"},
			ResourceIDs:    []string{"res1"},
			StatIDs:        []string{"stat1"},
			Weights:        utils.DynamicWeights{{}},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: true,
				},
			},
			RouteParameters: "params",
		}},
	}

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.CacheCfg().Partitions[utils.CacheRouteProfiles].Replicate = true
	config.SetCgrConfig(cfg)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error {

				return utils.ErrNotImplemented
			},
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.CacheSv1, cc)

	data := &DataDBMock{
		GetRouteProfileDrvF: func(ctx *context.Context, tnt, id string) (*utils.RouteProfile, error) { return rp, utils.ErrNotFound },
	}
	dm := NewDataManager(data, cfg, cM)

	Cache = NewCacheS(cfg, dm, cM, nil)

	_, err := dm.GetRouteProfile(context.Background(), utils.CGRateSorg, rp.ID, false, true, utils.NonTransactional)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}
}

func TestDMGetRouteProfileCacheWriteErr2(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.CacheCfg().Partitions[utils.CacheRouteProfiles].Replicate = true
	config.SetCgrConfig(cfg)

	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error { return utils.ErrNotImplemented },
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.CacheSv1, cc)
	dm := NewDataManager(data, cfg, cM)

	rp := &utils.RouteProfile{

		Tenant:            "cgrates.org",
		ID:                "RP_1",
		FilterIDs:         []string{"fltr_test"},
		Weights:           utils.DynamicWeights{{}},
		Sorting:           utils.MetaQOS,
		SortingParameters: []string{"param"},
		Routes: []*utils.Route{{
			ID:             "RT1",
			FilterIDs:      []string{"fltr1"},
			AccountIDs:     []string{"acc1"},
			RateProfileIDs: []string{"rp1"},
			ResourceIDs:    []string{"res1"},
			StatIDs:        []string{"stat1"},
			Weights:        utils.DynamicWeights{{}},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: true,
				},
			},
			RouteParameters: "params",
		}},
	}

	if err := dm.dataDB.SetRouteProfileDrv(context.Background(), rp); err != nil {
		t.Error(err)
	}

	Cache = NewCacheS(cfg, dm, cM, nil)

	if _, err := dm.GetRouteProfile(context.Background(), utils.CGRateSorg, rp.ID, false, true, utils.NonTransactional); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestDMSetRouteProfileNoDMErr(t *testing.T) {
	var dm *DataManager
	err := dm.SetRouteProfile(context.Background(), &utils.RouteProfile{}, false)
	if err != utils.ErrNoDatabaseConn {
		t.Errorf("\nExpected error <%+v>, \nReceived error <%+v>", utils.ErrNoDatabaseConn, err)
	}
}

func TestDMSetRouteProfileCheckFiltersErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	rpp := &utils.RouteProfile{
		Tenant:            "cgrates.org",
		ID:                "RP_1",
		FilterIDs:         []string{":::"},
		Weights:           utils.DynamicWeights{{}},
		Sorting:           utils.MetaQOS,
		SortingParameters: []string{"param"},
		Routes: []*utils.Route{{
			ID:             "RT1",
			FilterIDs:      []string{"fltr1"},
			AccountIDs:     []string{"acc1"},
			RateProfileIDs: []string{"rp1"},
			ResourceIDs:    []string{"res1"},
			StatIDs:        []string{"stat1"},
			Weights:        utils.DynamicWeights{{}},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: true,
				},
			},
			RouteParameters: "params",
		}},
	}

	expErr := "broken reference to filter: <:::> for item with ID: cgrates.org:RP_1"
	if err := dm.SetRouteProfile(context.Background(), rpp, true); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, received error <%v>", expErr, err)
	}
}

func TestDMSetRouteProfileGetRouteProfileErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetRouteProfileDrvF: func(ctx *context.Context, tnt, id string) (*utils.RouteProfile, error) {
			return &utils.RouteProfile{}, utils.ErrNotImplemented
		},
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	rpp := &utils.RouteProfile{
		Tenant:            "cgrates.org",
		ID:                "RP_1",
		FilterIDs:         []string{"FilterID1"},
		Weights:           utils.DynamicWeights{{}},
		Sorting:           utils.MetaQOS,
		SortingParameters: []string{"param"},
		Routes: []*utils.Route{{
			ID:             "RT1",
			FilterIDs:      []string{"fltr1"},
			AccountIDs:     []string{"acc1"},
			RateProfileIDs: []string{"rp1"},
			ResourceIDs:    []string{"res1"},
			StatIDs:        []string{"stat1"},
			Weights:        utils.DynamicWeights{{}},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: true,
				},
			},
			RouteParameters: "params",
		}},
	}

	if err := dm.SetRouteProfile(context.Background(), rpp, false); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}
}

func TestDMSetRouteProfileSetRouteProfileDrvErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetRouteProfileDrvF: func(ctx *context.Context, tnt, id string) (*utils.RouteProfile, error) {
			return &utils.RouteProfile{}, nil
		},
		SetRouteProfileDrvF: func(ctx *context.Context, rtPrf *utils.RouteProfile) error { return utils.ErrNotImplemented },
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	rpp := &utils.RouteProfile{
		Tenant:            "cgrates.org",
		ID:                "RP_1",
		FilterIDs:         []string{"FilterID1"},
		Weights:           utils.DynamicWeights{{}},
		Sorting:           utils.MetaQOS,
		SortingParameters: []string{"param"},
		Routes: []*utils.Route{{
			ID:             "RT1",
			FilterIDs:      []string{"fltr1"},
			AccountIDs:     []string{"acc1"},
			RateProfileIDs: []string{"rp1"},
			ResourceIDs:    []string{"res1"},
			StatIDs:        []string{"stat1"},
			Weights:        utils.DynamicWeights{{}},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: true,
				},
			},
			RouteParameters: "params",
		}},
	}

	if err := dm.SetRouteProfile(context.Background(), rpp, false); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}
}

func TestDMSetRouteProfileUpdatedIndexesErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetRouteProfileDrvF: func(ctx *context.Context, tnt, id string) (*utils.RouteProfile, error) {
			return &utils.RouteProfile{}, nil
		},
		SetRouteProfileDrvF: func(ctx *context.Context, rtPrf *utils.RouteProfile) error { return nil },
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	rpp := &utils.RouteProfile{
		Tenant:            "cgrates.org",
		ID:                "RP_1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		Weights:           utils.DynamicWeights{{}},
		Sorting:           utils.MetaQOS,
		SortingParameters: []string{"param"},
		Routes: []*utils.Route{{
			ID:             "RT1",
			FilterIDs:      []string{"fltr1"},
			AccountIDs:     []string{"acc1"},
			RateProfileIDs: []string{"rp1"},
			ResourceIDs:    []string{"res1"},
			StatIDs:        []string{"stat1"},
			Weights:        utils.DynamicWeights{{}},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: true,
				},
			},
			RouteParameters: "params",
		}},
	}

	if err := dm.SetRouteProfile(context.Background(), rpp, true); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}
}

func TestDMSetRouteProfileReplicate(t *testing.T) {

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	rpp := &utils.RouteProfile{
		Tenant:            "cgrates.org",
		ID:                "RP_1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		Weights:           utils.DynamicWeights{{}},
		Sorting:           utils.MetaQOS,
		SortingParameters: []string{"param"},
		Routes: []*utils.Route{{
			ID:             "RT1",
			FilterIDs:      []string{"fltr1"},
			AccountIDs:     []string{"acc1"},
			RateProfileIDs: []string{"rp1"},
			ResourceIDs:    []string{"res1"},
			StatIDs:        []string{"stat1"},
			Weights:        utils.DynamicWeights{{}},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: true,
				},
			},
			RouteParameters: "params",
		}},
	}

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaRouteProfiles].Replicate = true
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	config.SetCgrConfig(cfg)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1SetRouteProfile: func(ctx *context.Context, args, reply any) error { return utils.ErrNotImplemented },
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.ReplicatorSv1, cc)
	data := &DataDBMock{
		GetRouteProfileDrvF: func(ctx *context.Context, tnt, id string) (*utils.RouteProfile, error) { return rpp, nil },
		SetRouteProfileDrvF: func(ctx *context.Context, rtPrf *utils.RouteProfile) error { return nil },
	}
	dm := NewDataManager(data, cfg, cM)

	// tests replicate
	if err := dm.SetRouteProfile(context.Background(), rpp, false); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestDMRemoveRouteProfileNoDMErr(t *testing.T) {
	var dm *DataManager
	err := dm.RemoveRouteProfile(context.Background(), "cgrates.org", "RP_1", false)
	if err != utils.ErrNoDatabaseConn {
		t.Errorf("\nExpected error <%+v>, \nReceived error <%+v>", utils.ErrNoDatabaseConn, err)
	}
}

func TestDMRemoveRouteProfileGetRouteProfileErr(t *testing.T) {

	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetRouteProfileDrvF: func(ctx *context.Context, tnt, id string) (*utils.RouteProfile, error) {
			return &utils.RouteProfile{}, utils.ErrNotImplemented
		},
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	rpp := &utils.RouteProfile{
		Tenant:            "cgrates.org",
		ID:                "RP_1",
		FilterIDs:         []string{"FilterID1"},
		Weights:           utils.DynamicWeights{{}},
		Sorting:           utils.MetaQOS,
		SortingParameters: []string{"param"},
		Routes: []*utils.Route{{
			ID:             "RT1",
			FilterIDs:      []string{"fltr1"},
			AccountIDs:     []string{"acc1"},
			RateProfileIDs: []string{"rp1"},
			ResourceIDs:    []string{"res1"},
			StatIDs:        []string{"stat1"},
			Weights:        utils.DynamicWeights{{}},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: true,
				},
			},
			RouteParameters: "params",
		}},
	}

	err := dm.RemoveRouteProfile(context.Background(), rpp.Tenant, rpp.ID, false)
	if err != utils.ErrNotImplemented {
		t.Errorf("\nExpected error <%+v>, \nReceived error <%+v>", utils.ErrNotImplemented, err)
	}

}

func TestDMRemoveRouteProfileRemoveRouteProfileDrvErr(t *testing.T) {

	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetRouteProfileDrvF: func(ctx *context.Context, tnt, id string) (*utils.RouteProfile, error) {
			return &utils.RouteProfile{}, nil
		},
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	rpp := &utils.RouteProfile{
		Tenant:            "cgrates.org",
		ID:                "RP_1",
		FilterIDs:         []string{"FilterID1"},
		Weights:           utils.DynamicWeights{{}},
		Sorting:           utils.MetaQOS,
		SortingParameters: []string{"param"},
		Routes: []*utils.Route{{
			ID:             "RT1",
			FilterIDs:      []string{"fltr1"},
			AccountIDs:     []string{"acc1"},
			RateProfileIDs: []string{"rp1"},
			ResourceIDs:    []string{"res1"},
			StatIDs:        []string{"stat1"},
			Weights:        utils.DynamicWeights{{}},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: true,
				},
			},
			RouteParameters: "params",
		}},
	}

	err := dm.RemoveRouteProfile(context.Background(), rpp.Tenant, rpp.ID, false)
	if err != utils.ErrNotImplemented {
		t.Errorf("\nExpected error <%+v>, \nReceived error <%+v>", utils.ErrNotImplemented, err)
	}

}

func TestDMRemoveRouteProfileNilOldRppErr(t *testing.T) {

	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	var Id string
	var tnt string
	err := dm.RemoveRouteProfile(context.Background(), tnt, Id, false)
	if err != utils.ErrNotFound {
		t.Errorf("\nExpected error <%+v>, \nReceived error <%+v>", utils.ErrNotFound, err)
	}

}

func TestDMRemoveRouteProfileRmvItemFromFiltrIndexErr(t *testing.T) {

	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetRouteProfileDrvF: func(ctx *context.Context, tnt, id string) (*utils.RouteProfile, error) {
			return &utils.RouteProfile{}, nil
		},
		RemoveRouteProfileDrvF: func(ctx *context.Context, tnt, id string) error { return nil },
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	rpp := &utils.RouteProfile{
		Tenant:            "cgrates.org",
		ID:                "RP_1",
		FilterIDs:         []string{"FilterID1"},
		Weights:           utils.DynamicWeights{{}},
		Sorting:           utils.MetaQOS,
		SortingParameters: []string{"param"},
		Routes: []*utils.Route{{
			ID:             "RT1",
			FilterIDs:      []string{"fltr1"},
			AccountIDs:     []string{"acc1"},
			RateProfileIDs: []string{"rp1"},
			ResourceIDs:    []string{"res1"},
			StatIDs:        []string{"stat1"},
			Weights:        utils.DynamicWeights{{}},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: true,
				},
			},
			RouteParameters: "params",
		}},
	}

	err := dm.RemoveRouteProfile(context.Background(), rpp.Tenant, rpp.ID, true)
	if err != utils.ErrNotImplemented {
		t.Errorf("\nExpected error <%+v>, \nReceived error <%+v>", utils.ErrNotImplemented, err)
	}

}

func TestDMRemoveRouteProfileRmvIndexFiltersItemErr(t *testing.T) {

	Cache.Clear(nil)

	rpp := &utils.RouteProfile{
		Tenant:            "cgrates.org",
		ID:                "RP_1",
		FilterIDs:         []string{"fltrID"},
		Weights:           utils.DynamicWeights{{}},
		Sorting:           utils.MetaQOS,
		SortingParameters: []string{"param"},
		Routes: []*utils.Route{{
			ID:             "RT1",
			FilterIDs:      []string{"fltr1"},
			AccountIDs:     []string{"acc1"},
			RateProfileIDs: []string{"rp1"},
			ResourceIDs:    []string{"res1"},
			StatIDs:        []string{"stat1"},
			Weights:        utils.DynamicWeights{{}},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: true,
				},
			},
			RouteParameters: "params",
		}},
	}

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetRouteProfileDrvF:    func(ctx *context.Context, tnt, id string) (*utils.RouteProfile, error) { return rpp, nil },
		RemoveRouteProfileDrvF: func(ctx *context.Context, tnt, id string) error { return nil },
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	err := dm.RemoveRouteProfile(context.Background(), rpp.Tenant, rpp.ID, true)
	if err != utils.ErrNotImplemented {
		t.Errorf("\nExpected error <%+v>, \nReceived error <%+v>", utils.ErrNotImplemented, err)
	}

}

func TestDMRemoveRouteProfileReplicate(t *testing.T) {

	cfgtmp := config.CgrConfig()
	defer func() {
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	rpp := &utils.RouteProfile{
		Tenant:            "cgrates.org",
		ID:                "RP_1",
		FilterIDs:         []string{"fltrID"},
		Weights:           utils.DynamicWeights{{}},
		Sorting:           utils.MetaQOS,
		SortingParameters: []string{"param"},
		Routes: []*utils.Route{{
			ID:             "RT1",
			FilterIDs:      []string{"fltr1"},
			AccountIDs:     []string{"acc1"},
			RateProfileIDs: []string{"rp1"},
			ResourceIDs:    []string{"res1"},
			StatIDs:        []string{"stat1"},
			Weights:        utils.DynamicWeights{{}},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: true,
				},
			},
			RouteParameters: "params",
		}},
	}

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaRouteProfiles].Replicate = true
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	config.SetCgrConfig(cfg)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1RemoveRouteProfile: func(ctx *context.Context, args, reply any) error { return utils.ErrNotImplemented },
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.ReplicatorSv1, cc)
	data := &DataDBMock{
		GetRouteProfileDrvF:    func(ctx *context.Context, tnt, id string) (*utils.RouteProfile, error) { return rpp, nil },
		RemoveRouteProfileDrvF: func(ctx *context.Context, tnt, id string) error { return nil },
	}
	dm := NewDataManager(data, cfg, cM)

	// tests replicate
	dm.RemoveRouteProfile(context.Background(), rpp.Tenant, rpp.ID, false)

}

func TestDMRemoveAttributeProfileNoDMErr(t *testing.T) {
	var dm *DataManager
	err := dm.RemoveAttributeProfile(context.Background(), "cgrates.org", "ap_1", false)
	if err != utils.ErrNoDatabaseConn {
		t.Errorf("\nExpected error <%+v>, \nReceived error <%+v>", utils.ErrNoDatabaseConn, err)
	}
}

func TestDMRemoveAttributeProfileGetAttributeProfileErr(t *testing.T) {

	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetAttributeProfileDrvF: func(ctx *context.Context, str1, str2 string) (*utils.AttributeProfile, error) {
			return &utils.AttributeProfile{}, utils.ErrNotImplemented
		},
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	attrPrfl := &utils.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_ID",
		FilterIDs: []string{"FLTR_ACNT_1"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Account",
				Value: utils.NewRSRParsersMustCompile("1001", utils.InfieldSep),
			},
		},
		Weights: make(utils.DynamicWeights, 1),
	}

	err := dm.RemoveAttributeProfile(context.Background(), attrPrfl.Tenant, attrPrfl.ID, false)
	if err != utils.ErrNotImplemented {
		t.Errorf("\nExpected error <%+v>, \nReceived error <%+v>", utils.ErrNotImplemented, err)
	}

}

func TestDMRemoveAttributeProfileRemoveAttributeProfileDrvErr(t *testing.T) {

	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetAttributeProfileDrvF: func(ctx *context.Context, str1, str2 string) (*utils.AttributeProfile, error) {
			return &utils.AttributeProfile{}, nil
		},
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	attrPrfl := &utils.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_ID",
		FilterIDs: []string{"FLTR_ACNT_1"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Account",
				Value: utils.NewRSRParsersMustCompile("1001", utils.InfieldSep),
			},
		},
		Weights: make(utils.DynamicWeights, 1),
	}

	err := dm.RemoveAttributeProfile(context.Background(), attrPrfl.Tenant, attrPrfl.ID, false)
	if err != utils.ErrNotImplemented {
		t.Errorf("\nExpected error <%+v>, \nReceived error <%+v>", utils.ErrNotImplemented, err)
	}

}

func TestDMRemoveAttributeProfileNilOldAttrErr(t *testing.T) {

	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	var Id string
	var tnt string
	err := dm.RemoveAttributeProfile(context.Background(), tnt, Id, false)
	if err != utils.ErrNotFound {
		t.Errorf("\nExpected error <%+v>, \nReceived error <%+v>", utils.ErrNotFound, err)
	}

}

func TestDMRemoveAttributeProfileRmvItemFromFiltrIndexErr(t *testing.T) {

	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetAttributeProfileDrvF: func(ctx *context.Context, str1, str2 string) (*utils.AttributeProfile, error) {
			return &utils.AttributeProfile{}, nil
		},
		RemoveAttributeProfileDrvF: func(ctx *context.Context, str1, str2 string) error { return nil },
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	attrPrfl := &utils.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_ID",
		FilterIDs: []string{"FLTR_ACNT_1"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Account",
				Value: utils.NewRSRParsersMustCompile("1001", utils.InfieldSep),
			},
		},
		Weights: make(utils.DynamicWeights, 1),
	}

	err := dm.RemoveAttributeProfile(context.Background(), attrPrfl.Tenant, attrPrfl.ID, true)
	if err != utils.ErrNotImplemented {
		t.Errorf("\nExpected error <%+v>, \nReceived error <%+v>", utils.ErrNotImplemented, err)
	}

}

func TestDMRemoveAttributeProfileRmvIndexFiltersItemErr(t *testing.T) {

	Cache.Clear(nil)

	attrPrfl := &utils.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_ID",
		FilterIDs: []string{"FLTR_ACNT_1"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Account",
				Value: utils.NewRSRParsersMustCompile("1001", utils.InfieldSep),
			},
		},
		Weights: make(utils.DynamicWeights, 1),
	}

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetAttributeProfileDrvF: func(ctx *context.Context, str1, str2 string) (*utils.AttributeProfile, error) {
			return attrPrfl, nil
		},
		RemoveAttributeProfileDrvF: func(ctx *context.Context, str1, str2 string) error { return nil },
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	err := dm.RemoveAttributeProfile(context.Background(), attrPrfl.Tenant, attrPrfl.ID, true)
	if err != utils.ErrNotImplemented {
		t.Errorf("\nExpected error <%+v>, \nReceived error <%+v>", utils.ErrNotImplemented, err)
	}

}

func TestDMRemoveAttributeProfileReplicate(t *testing.T) {

	cfgtmp := config.CgrConfig()
	defer func() {
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	attrPrfl := &utils.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_ID",
		FilterIDs: []string{"FLTR_ACNT_1"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Account",
				Value: utils.NewRSRParsersMustCompile("1001", utils.InfieldSep),
			},
		},
		Weights: make(utils.DynamicWeights, 1),
	}

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaAttributeProfiles].Replicate = true
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	config.SetCgrConfig(cfg)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1RemoveRouteProfile: func(ctx *context.Context, args, reply any) error { return utils.ErrNotImplemented },
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.ReplicatorSv1, cc)
	data := &DataDBMock{
		GetAttributeProfileDrvF: func(ctx *context.Context, str1, str2 string) (*utils.AttributeProfile, error) {
			return attrPrfl, nil
		},
		RemoveAttributeProfileDrvF: func(ctx *context.Context, str1, str2 string) error { return nil },
	}
	dm := NewDataManager(data, cfg, cM)

	// tests replicate
	dm.RemoveAttributeProfile(context.Background(), attrPrfl.Tenant, attrPrfl.ID, false)

}

func TestDMRemoveChargerProfileNoDMErr(t *testing.T) {
	var dm *DataManager
	err := dm.RemoveChargerProfile(context.Background(), "cgrates.org", "cp_1", false)
	if err != utils.ErrNoDatabaseConn {
		t.Errorf("\nExpected error <%+v>, \nReceived error <%+v>", utils.ErrNoDatabaseConn, err)
	}
}

func TestDMRemoveChargerProfileGetChargerProfileErr(t *testing.T) {

	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetChargerProfileDrvF: func(ctx *context.Context, tnt, id string) (*utils.ChargerProfile, error) {
			return &utils.ChargerProfile{}, utils.ErrNotImplemented
		},
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	cpp := &utils.ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "CPP_1",
		FilterIDs:    []string{"FLTR_CP_1"},
		RunID:        "TestRunID",
		AttributeIDs: []string{"*none"},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}

	err := dm.RemoveChargerProfile(context.Background(), cpp.Tenant, cpp.ID, false)
	if err != utils.ErrNotImplemented {
		t.Errorf("\nExpected error <%+v>, \nReceived error <%+v>", utils.ErrNotImplemented, err)
	}

}

func TestDMRemoveChargerProfileRemoveChargerProfileDrvErr(t *testing.T) {

	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetChargerProfileDrvF: func(ctx *context.Context, tnt, id string) (*utils.ChargerProfile, error) {
			return &utils.ChargerProfile{}, nil
		},
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	cpp := &utils.ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "CPP_1",
		FilterIDs:    []string{"FLTR_CP_1"},
		RunID:        "TestRunID",
		AttributeIDs: []string{"*none"},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}

	err := dm.RemoveChargerProfile(context.Background(), cpp.Tenant, cpp.ID, false)
	if err != utils.ErrNotImplemented {
		t.Errorf("\nExpected error <%+v>, \nReceived error <%+v>", utils.ErrNotImplemented, err)
	}

}

func TestDMRemoveChargerProfileNilOldCppErr(t *testing.T) {

	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	var Id string
	var tnt string
	err := dm.RemoveChargerProfile(context.Background(), tnt, Id, false)
	if err != utils.ErrNotFound {
		t.Errorf("\nExpected error <%+v>, \nReceived error <%+v>", utils.ErrNotFound, err)
	}

}

func TestDMRemoveChargerProfileRmvItemFromFiltrIndexErr(t *testing.T) {

	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetChargerProfileDrvF: func(ctx *context.Context, tnt, id string) (*utils.ChargerProfile, error) {
			return &utils.ChargerProfile{}, nil
		},
		RemoveChargerProfileDrvF: func(ctx *context.Context, chr, rpl string) error { return nil },
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	cpp := &utils.ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "CPP_1",
		FilterIDs:    []string{"FLTR_CP_1"},
		RunID:        "TestRunID",
		AttributeIDs: []string{"*none"},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}

	err := dm.RemoveChargerProfile(context.Background(), cpp.Tenant, cpp.ID, true)
	if err != utils.ErrNotImplemented {
		t.Errorf("\nExpected error <%+v>, \nReceived error <%+v>", utils.ErrNotImplemented, err)
	}

}

func TestDMRemoveChargerProfileRmvIndexFiltersItemErr(t *testing.T) {

	Cache.Clear(nil)

	cpp := &utils.ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "CPP_1",
		FilterIDs:    []string{"FLTR_CP_1"},
		RunID:        "TestRunID",
		AttributeIDs: []string{"*none"},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetChargerProfileDrvF: func(ctx *context.Context, tnt, id string) (*utils.ChargerProfile, error) {
			return cpp, nil
		},
		RemoveChargerProfileDrvF: func(ctx *context.Context, chr, rpl string) error { return nil },
	}

	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	err := dm.RemoveChargerProfile(context.Background(), cpp.Tenant, cpp.ID, true)
	if err != utils.ErrNotImplemented {
		t.Errorf("\nExpected error <%+v>, \nReceived error <%+v>", utils.ErrNotImplemented, err)
	}

}

func TestDMRemoveChargerProfileReplicate(t *testing.T) {

	cfgtmp := config.CgrConfig()
	defer func() {
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cpp := &utils.ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "CPP_1",
		FilterIDs:    []string{"FLTR_CP_1"},
		RunID:        "TestRunID",
		AttributeIDs: []string{"*none"},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaChargerProfiles].Replicate = true
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	config.SetCgrConfig(cfg)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1RemoveChargerProfile: func(ctx *context.Context, args, reply any) error { return utils.ErrNotImplemented },
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.ReplicatorSv1, cc)
	data := &DataDBMock{
		GetChargerProfileDrvF: func(ctx *context.Context, tnt, id string) (*utils.ChargerProfile, error) {
			return cpp, nil
		},
		RemoveChargerProfileDrvF: func(ctx *context.Context, chr, rpl string) error { return nil },
	}
	dm := NewDataManager(data, cfg, cM)

	// tests replicate
	dm.RemoveChargerProfile(context.Background(), cpp.Tenant, cpp.ID, false)

}

func TestDMRemoveRateProfileNoDMErr(t *testing.T) {
	var dm *DataManager
	err := dm.RemoveRateProfile(context.Background(), "cgrates.org", "Rp_1", false)
	if err != utils.ErrNoDatabaseConn {
		t.Errorf("\nExpected error <%+v>, \nReceived error <%+v>", utils.ErrNoDatabaseConn, err)
	}
}

func TestDMRemoveRateProfileGetRateProfileErr(t *testing.T) {

	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetRateProfileDrvF: func(ctx *context.Context, s1, s2 string) (*utils.RateProfile, error) {
			return &utils.RateProfile{}, utils.ErrNotImplemented
		},
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	rpp := &utils.RateProfile{
		ID:        "test_ID1",
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.Destination:1234"},
		Rates: map[string]*utils.Rate{
			"RT1": {
				ID: "RT1",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(1, 2),
						Unit:          utils.NewDecimal(int64(time.Second), 0),
						Increment:     utils.NewDecimal(int64(time.Second), 0),
					},
				},
			},
		},
	}

	err := dm.RemoveRateProfile(context.Background(), rpp.Tenant, rpp.ID, false)
	if err != utils.ErrNotImplemented {
		t.Errorf("\nExpected error <%+v>, \nReceived error <%+v>", utils.ErrNotImplemented, err)
	}

}

func TestDMRemoveRateProfileRemoveRateProfileDrvErr(t *testing.T) {

	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetRateProfileDrvF: func(ctx *context.Context, s1, s2 string) (*utils.RateProfile, error) {
			return &utils.RateProfile{}, nil
		},
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	rpp := &utils.RateProfile{
		ID:        "test_ID1",
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.Destination:1234"},
		Rates: map[string]*utils.Rate{
			"RT1": {
				ID: "RT1",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(1, 2),
						Unit:          utils.NewDecimal(int64(time.Second), 0),
						Increment:     utils.NewDecimal(int64(time.Second), 0),
					},
				},
			},
		},
	}

	err := dm.RemoveRateProfile(context.Background(), rpp.Tenant, rpp.ID, false)
	if err != utils.ErrNotImplemented {
		t.Errorf("\nExpected error <%+v>, \nReceived error <%+v>", utils.ErrNotImplemented, err)
	}

}

func TestDMRemoveRateProfileNilOldRppErr(t *testing.T) {

	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	var Id string
	var tnt string
	err := dm.RemoveRateProfile(context.Background(), tnt, Id, false)
	if err != utils.ErrNotFound {
		t.Errorf("\nExpected error <%+v>, \nReceived error <%+v>", utils.ErrNotFound, err)
	}

}

func TestDMRemoveRateProfileRemoveIndexFiltersItemErr1(t *testing.T) {

	Cache.Clear(nil)

	rpp := &utils.RateProfile{
		ID:        "test_ID1",
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.Destination:1234"},
		Rates: map[string]*utils.Rate{
			"RT1": {
				FilterIDs: []string{"FltrId1"},
				ID:        "RT1",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(1, 2),
						Unit:          utils.NewDecimal(int64(time.Second), 0),
						Increment:     utils.NewDecimal(int64(time.Second), 0),
					},
				},
			},
		},
	}

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetRateProfileDrvF: func(ctx *context.Context, s1, s2 string) (*utils.RateProfile, error) {
			return rpp, nil
		},
		RemoveRateProfileDrvF: func(ctx *context.Context, str1, str2 string, rateIDs *[]string) error { return nil },
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	err := dm.RemoveRateProfile(context.Background(), rpp.Tenant, rpp.ID, true)
	if err != utils.ErrNotImplemented {
		t.Errorf("\nExpected error <%+v>, \nReceived error <%+v>", utils.ErrNotImplemented, err)
	}

}
func TestDMRemoveRateProfileRemoveIndexFiltersItemErr2(t *testing.T) {

	Cache.Clear(nil)

	rpp := &utils.RateProfile{
		ID:        "test_ID1",
		Tenant:    "cgrates.org",
		FilterIDs: []string{"fltrId1"},
	}

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetRateProfileDrvF: func(ctx *context.Context, s1, s2 string) (*utils.RateProfile, error) {
			return rpp, nil
		},
		RemoveRateProfileDrvF: func(ctx *context.Context, str1, str2 string, rateIDs *[]string) error { return nil },
	}

	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	err := dm.RemoveRateProfile(context.Background(), rpp.Tenant, rpp.ID, true)
	if err != utils.ErrNotImplemented {
		t.Errorf("\nExpected error <%+v>, \nReceived error <%+v>", utils.ErrNotImplemented, err)
	}

}

func TestDMRemoveRateProfileRmvItemFromFiltrIndexErr(t *testing.T) {

	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetRateProfileDrvF: func(ctx *context.Context, s1, s2 string) (*utils.RateProfile, error) {
			return &utils.RateProfile{}, nil
		},
		RemoveRateProfileDrvF: func(ctx *context.Context, str1, str2 string, rateIDs *[]string) error { return nil },
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	rpp := &utils.RateProfile{
		ID:        "test_ID1",
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.Destination:1234"},
		Rates: map[string]*utils.Rate{
			"RT1": {
				ID: "RT1",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(1, 2),
						Unit:          utils.NewDecimal(int64(time.Second), 0),
						Increment:     utils.NewDecimal(int64(time.Second), 0),
					},
				},
			},
		},
	}

	err := dm.RemoveRateProfile(context.Background(), rpp.Tenant, rpp.ID, true)
	if err != utils.ErrNotImplemented {
		t.Errorf("\nExpected error <%+v>, \nReceived error <%+v>", utils.ErrNotImplemented, err)
	}

}

func TestDMRemoveRateProfileRmvIndexFiltersItemErr(t *testing.T) {

	Cache.Clear(nil)

	rpp := &utils.RateProfile{
		ID:        "test_ID1",
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.Destination:1234"},
		Rates: map[string]*utils.Rate{
			"RT1": {
				ID: "RT1",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(1, 2),
						Unit:          utils.NewDecimal(int64(time.Second), 0),
						Increment:     utils.NewDecimal(int64(time.Second), 0),
					},
				},
			},
		},
	}

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetRateProfileDrvF: func(ctx *context.Context, s1, s2 string) (*utils.RateProfile, error) {
			return rpp, nil
		},
		RemoveRateProfileDrvF: func(ctx *context.Context, str1, str2 string, rateIDs *[]string) error { return nil },
	}

	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	err := dm.RemoveRateProfile(context.Background(), rpp.Tenant, rpp.ID, true)
	if err != utils.ErrNotImplemented {
		t.Errorf("\nExpected error <%+v>, \nReceived error <%+v>", utils.ErrNotImplemented, err)
	}

}

func TestDMRemoveRateProfileReplicate(t *testing.T) {

	cfgtmp := config.CgrConfig()
	defer func() {
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	rpp := &utils.RateProfile{
		ID:        "test_ID1",
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.Destination:1234"},
		Rates: map[string]*utils.Rate{
			"RT1": {
				ID: "RT1",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(1, 2),
						Unit:          utils.NewDecimal(int64(time.Second), 0),
						Increment:     utils.NewDecimal(int64(time.Second), 0),
					},
				},
			},
		},
	}

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaRateProfiles].Replicate = true
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	config.SetCgrConfig(cfg)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1RemoveRateProfile: func(ctx *context.Context, args, reply any) error { return utils.ErrNotImplemented },
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.ReplicatorSv1, cc)
	data := &DataDBMock{
		GetRateProfileDrvF: func(ctx *context.Context, s1, s2 string) (*utils.RateProfile, error) {
			return rpp, nil
		},
		RemoveRateProfileDrvF: func(ctx *context.Context, str1, str2 string, rateIDs *[]string) error { return nil },
	}
	dm := NewDataManager(data, cfg, cM)

	// tests replicate
	dm.RemoveRateProfile(context.Background(), rpp.Tenant, rpp.ID, false)

}

func TestDMRemoveActionProfileNilOldActErr(t *testing.T) {

	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	var Id string
	var tnt string
	err := dm.RemoveActionProfile(context.Background(), tnt, Id, false)
	if err != utils.ErrNotFound {
		t.Errorf("\nExpected error <%+v>, \nReceived error <%+v>", utils.ErrNotFound, err)
	}

}

func TestDMRemoveActionProfileRmvItemFromFiltrIndexErr(t *testing.T) {

	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetActionProfileDrvF: func(ctx *context.Context, tenant, ID string) (*utils.ActionProfile, error) {
			return &utils.ActionProfile{}, nil
		},
		RemoveActionProfileDrvF: func(ctx *context.Context, tenant, ID string) error { return nil },
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	ap := &utils.ActionProfile{

		Tenant:    "cgrates.org",
		ID:        "AP1",
		FilterIDs: []string{"fltr1"},
		Weights: utils.DynamicWeights{
			{
				Weight: 65,
			},
		},
		Schedule: "* * * * *",
		Targets:  map[string]utils.StringSet{utils.MetaAccounts: {"1001": {}}},
		Actions:  []*utils.APAction{{}},
	}

	err := dm.RemoveActionProfile(context.Background(), ap.Tenant, ap.ID, true)
	if err != utils.ErrNotImplemented {
		t.Errorf("\nExpected error <%+v>, \nReceived error <%+v>", utils.ErrNotImplemented, err)
	}

}

func TestDMRemoveActionProfileRmvIndexFiltersItemErr(t *testing.T) {

	Cache.Clear(nil)

	ap := &utils.ActionProfile{

		Tenant:    "cgrates.org",
		ID:        "AP1",
		FilterIDs: []string{"fltr1"},
		Weights: utils.DynamicWeights{
			{
				Weight: 65,
			},
		},
		Schedule: "* * * * *",
		Targets:  map[string]utils.StringSet{utils.MetaAccounts: {"1001": {}}},
		Actions:  []*utils.APAction{{}},
	}

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetActionProfileDrvF:    func(ctx *context.Context, tenant, ID string) (*utils.ActionProfile, error) { return ap, nil },
		RemoveActionProfileDrvF: func(ctx *context.Context, tenant, ID string) error { return nil },
	}

	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	err := dm.RemoveActionProfile(context.Background(), ap.Tenant, ap.ID, true)
	if err != utils.ErrNotImplemented {
		t.Errorf("\nExpected error <%+v>, \nReceived error <%+v>", utils.ErrNotImplemented, err)
	}

}

func TestDMRemoveActionProfileReplicate(t *testing.T) {

	cfgtmp := config.CgrConfig()
	defer func() {
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	ap := &utils.ActionProfile{

		Tenant:    "cgrates.org",
		ID:        "AP1",
		FilterIDs: []string{"fltr1"},
		Weights: utils.DynamicWeights{
			{
				Weight: 65,
			},
		},
		Schedule: "* * * * *",
		Targets:  map[string]utils.StringSet{utils.MetaAccounts: {"1001": {}}},
		Actions:  []*utils.APAction{{}},
	}

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaActionProfiles].Replicate = true
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	config.SetCgrConfig(cfg)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1RemoveActionProfile: func(ctx *context.Context, args, reply any) error { return utils.ErrNotImplemented },
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.ReplicatorSv1, cc)
	data := &DataDBMock{
		GetActionProfileDrvF:    func(ctx *context.Context, tenant, ID string) (*utils.ActionProfile, error) { return ap, nil },
		RemoveActionProfileDrvF: func(ctx *context.Context, tenant, ID string) error { return nil },
	}
	dm := NewDataManager(data, cfg, cM)

	// tests replicate
	dm.RemoveActionProfile(context.Background(), ap.Tenant, ap.ID, false)

}

func TestDMSetAttributeProfileNoDMErr(t *testing.T) {
	var dm *DataManager
	err := dm.SetAttributeProfile(context.Background(), &utils.AttributeProfile{}, false)
	if err != utils.ErrNoDatabaseConn {
		t.Errorf("\nExpected error <%+v>, \nReceived error <%+v>", utils.ErrNoDatabaseConn, err)
	}
}

func TestDMSetAttributeProfileCheckFiltersErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	attrPrfl := &utils.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_ID",
		FilterIDs: []string{":::"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Account",
				Value: utils.NewRSRParsersMustCompile("1001", utils.InfieldSep),
			},
		},
		Weights: make(utils.DynamicWeights, 1),
	}

	expErr := "broken reference to filter: <:::> for item with ID: cgrates.org:ATTR_ID"
	if err := dm.SetAttributeProfile(context.Background(), attrPrfl, true); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, received error <%v>", expErr, err)
	}
}

func TestDMSetAttributeProfileGetAttributeProfileErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetAttributeProfileDrvF: func(ctx *context.Context, str1, str2 string) (*utils.AttributeProfile, error) {
			return &utils.AttributeProfile{}, utils.ErrNotImplemented
		},
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	attrPrfl := &utils.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_ID",
		FilterIDs: []string{"filtrId1"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Account",
				Value: utils.NewRSRParsersMustCompile("1001", utils.InfieldSep),
			},
		},
		Weights: make(utils.DynamicWeights, 1),
	}

	if err := dm.SetAttributeProfile(context.Background(), attrPrfl, false); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}
}

func TestDMSetAttributeProfileSetAttributeProfileDrvErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetAttributeProfileDrvF: func(ctx *context.Context, str1, str2 string) (*utils.AttributeProfile, error) {
			return &utils.AttributeProfile{}, nil
		},
		SetAttributeProfileDrvF: func(ctx *context.Context, attr *utils.AttributeProfile) error { return utils.ErrNotImplemented },
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	attrPrfl := &utils.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_ID",
		FilterIDs: []string{"filtrId1"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Account",
				Value: utils.NewRSRParsersMustCompile("1001", utils.InfieldSep),
			},
		},
		Weights: make(utils.DynamicWeights, 1),
	}
	if err := dm.SetAttributeProfile(context.Background(), attrPrfl, false); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}
}

func TestDMSetAttributeProfileUpdatedIndexesErr(t *testing.T) {

	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetAttributeProfileDrvF: func(ctx *context.Context, str1, str2 string) (*utils.AttributeProfile, error) {
			return &utils.AttributeProfile{}, nil
		},
		SetAttributeProfileDrvF: func(ctx *context.Context, attr *utils.AttributeProfile) error { return nil },
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	attrPrfl := &utils.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_ID",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Account",
				Value: utils.NewRSRParsersMustCompile("1001", utils.InfieldSep),
			},
		},
		Weights: make(utils.DynamicWeights, 1),
	}

	if err := dm.SetAttributeProfile(context.Background(), attrPrfl, true); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}
}

func TestDMSetAttributeProfileReplicate(t *testing.T) {

	cfgtmp := config.CgrConfig()
	defer func() {
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	attrPrfl := &utils.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_ID",
		FilterIDs: []string{"filtrId1"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Account",
				Value: utils.NewRSRParsersMustCompile("1001", utils.InfieldSep),
			},
		},
		Weights: make(utils.DynamicWeights, 1),
	}

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaAttributeProfiles].Replicate = true
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	config.SetCgrConfig(cfg)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1SetAttributeProfile: func(ctx *context.Context, args, reply any) error { return utils.ErrNotImplemented },
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.ReplicatorSv1, cc)
	data := &DataDBMock{
		GetAttributeProfileDrvF: func(ctx *context.Context, str1, str2 string) (*utils.AttributeProfile, error) {
			return attrPrfl, nil
		},
		SetAttributeProfileDrvF: func(ctx *context.Context, attr *utils.AttributeProfile) error { return nil },
	}
	dm := NewDataManager(data, cfg, cM)

	// tests replicate
	if err := dm.SetAttributeProfile(context.Background(), attrPrfl, false); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestDMGetAttributeProfileNilDmErr(t *testing.T) {

	var dm *DataManager

	_, err := dm.GetAttributeProfile(context.Background(), utils.CGRateSorg, "ap_1", false, false, utils.NonTransactional)
	if err != utils.ErrNoDatabaseConn {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNoDatabaseConn, err)
	}
}

func TestDMGetAttributeProfileSetAttributeProfileDrvErr(t *testing.T) {

	cfgtmp := config.CgrConfig()
	defer func() {
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	attrPrfl := &utils.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_ID",
		FilterIDs: []string{"filtrId1"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Account",
				Value: utils.NewRSRParsersMustCompile("1001", utils.InfieldSep),
			},
		},
		Weights: make(utils.DynamicWeights, 1),
	}

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaAttributeProfiles].Remote = true
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg)}
	config.SetCgrConfig(cfg)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetAttributeProfile: func(ctx *context.Context, args, reply any) error { return nil },
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg), utils.ReplicatorSv1, cc)
	data := &DataDBMock{
		GetAttributeProfileDrvF: func(ctx *context.Context, str1, str2 string) (*utils.AttributeProfile, error) {
			return attrPrfl, utils.ErrNotFound
		},
		SetAttributeProfileDrvF: func(ctx *context.Context, attr *utils.AttributeProfile) error { return utils.ErrNotImplemented },
	}
	dm := NewDataManager(data, cfg, cM)

	_, err := dm.GetAttributeProfile(context.Background(), utils.CGRateSorg, attrPrfl.ID, false, false, utils.NonTransactional)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}
}

func TestDMSetAttributeProfileComputeHashErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetAttributeProfileDrvF: func(ctx *context.Context, str1, str2 string) (*utils.AttributeProfile, error) {
			return &utils.AttributeProfile{}, nil
		},
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)
	value := utils.NewRSRParsersMustCompile("31 0a 0a 32 0a 0a 33 0a 0a 34 0a 0a 35 0a 0a 36 0a 0a 37 0a 0a 38 0a 0a 39 0a 0a 31 30 0a 0a 31", utils.RSRSep)
	attrPrfl := &utils.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_ID",
		FilterIDs: []string{"filtrId1"},
		Attributes: []*utils.Attribute{
			{
				Type:  utils.MetaPassword,
				Value: value,
			},
		},
		Weights: make(utils.DynamicWeights, 1),
	}

	expErr := "bcrypt: password length exceeds 72 bytes"
	if err := dm.SetAttributeProfile(context.Background(), attrPrfl, false); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, received error <%v>", expErr, err)
	}
}

func TestDMSetChargerProfileNoDMErr(t *testing.T) {
	var dm *DataManager
	err := dm.SetChargerProfile(context.Background(), &utils.ChargerProfile{}, false)
	if err != utils.ErrNoDatabaseConn {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrNoDatabaseConn, err)
	}
}

func TestDMSetChargerProfileCheckFiltersErr(t *testing.T) {

	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	cpp := &utils.ChargerProfile{
		Tenant:    "cgrates.org",
		ID:        "CHP_1",
		FilterIDs: []string{"*string*req.Account1001"},

		RunID:        "*rated",
		AttributeIDs: []string{"ATTR_1"},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}

	expErr := "broken reference to filter: <*string*req.Account1001> for item with ID: cgrates.org:CHP_1"
	if err := dm.SetChargerProfile(context.Background(), cpp, true); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, received error <%v>", expErr, err)
	}

}

func TestDMSetChargerProfileGetChargerProfileErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetChargerProfileDrvF: func(ctx *context.Context, tnt, id string) (*utils.ChargerProfile, error) {
			return &utils.ChargerProfile{}, utils.ErrNotImplemented
		},
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	cpp := &utils.ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "CHP_1",
		FilterIDs:    []string{"FltrId1"},
		RunID:        "*rated",
		AttributeIDs: []string{"ATTR_1"},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}

	if err := dm.SetChargerProfile(context.Background(), cpp, false); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}
}

func TestDMSetChargerProfileSetChargerProfileDrvErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetChargerProfileDrvF: func(ctx *context.Context, tnt, id string) (*utils.ChargerProfile, error) {
			return &utils.ChargerProfile{}, nil
		},
		SetChargerProfileDrvF: func(ctx *context.Context, chr *utils.ChargerProfile) (err error) { return utils.ErrNotImplemented },
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	cpp := &utils.ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "CHP_1",
		FilterIDs:    []string{"FltrId1"},
		RunID:        "*rated",
		AttributeIDs: []string{"ATTR_1"},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	if err := dm.SetChargerProfile(context.Background(), cpp, false); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}
}

func TestDMSetChargerProfileUpdatedIndexesErr(t *testing.T) {

	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetChargerProfileDrvF: func(ctx *context.Context, tnt, id string) (*utils.ChargerProfile, error) {
			return &utils.ChargerProfile{}, nil
		},
		SetChargerProfileDrvF: func(ctx *context.Context, chr *utils.ChargerProfile) (err error) { return nil },
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	cpp := &utils.ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "CHP_1",
		FilterIDs:    []string{"*string:~*req.Account:1001"},
		RunID:        "*rated",
		AttributeIDs: []string{"ATTR_1"},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}

	if err := dm.SetChargerProfile(context.Background(), cpp, true); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}
}

func TestDMSetChargerProfileReplicate(t *testing.T) {

	cfgtmp := config.CgrConfig()
	defer func() {
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cpp := &utils.ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "CHP_1",
		FilterIDs:    []string{"*string:~*req.Account:1001"},
		RunID:        "*rated",
		AttributeIDs: []string{"ATTR_1"},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaChargerProfiles].Replicate = true
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	config.SetCgrConfig(cfg)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1SetChargerProfile: func(ctx *context.Context, args, reply any) error { return utils.ErrNotImplemented },
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.ReplicatorSv1, cc)
	data := &DataDBMock{
		GetChargerProfileDrvF: func(ctx *context.Context, tnt, id string) (*utils.ChargerProfile, error) {
			return cpp, nil
		},
		SetChargerProfileDrvF: func(ctx *context.Context, chr *utils.ChargerProfile) (err error) { return nil },
	}
	dm := NewDataManager(data, cfg, cM)

	// tests replicate
	if err := dm.SetChargerProfile(context.Background(), cpp, false); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestDMSetActionProfileNoDMErr(t *testing.T) {
	var dm *DataManager
	err := dm.SetActionProfile(context.Background(), &utils.ActionProfile{}, false)
	if err != utils.ErrNoDatabaseConn {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrNoDatabaseConn, err)
	}
}

func TestDMSetActionProfileCheckFiltersErr(t *testing.T) {

	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	ap := &utils.ActionProfile{

		Tenant:    "cgrates.org",
		ID:        "AP1",
		FilterIDs: []string{"*string*req.Account1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 65,
			},
		},
		Schedule: "* * * * *",
		Targets:  map[string]utils.StringSet{utils.MetaAccounts: {"1001": {}}},
		Actions:  []*utils.APAction{{}},
	}

	expErr := "broken reference to filter: <*string*req.Account1001> for item with ID: cgrates.org:AP1"
	if err := dm.SetActionProfile(context.Background(), ap, true); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, received error <%v>", expErr, err)
	}

}

func TestDMSetActionProfileGetActionProfileErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetActionProfileDrvF: func(ctx *context.Context, tenant, ID string) (*utils.ActionProfile, error) {
			return &utils.ActionProfile{}, utils.ErrNotImplemented
		},
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	ap := &utils.ActionProfile{

		Tenant:    "cgrates.org",
		ID:        "AP1",
		FilterIDs: []string{"*string*req.Account1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 65,
			},
		},
		Schedule: "* * * * *",
		Targets:  map[string]utils.StringSet{utils.MetaAccounts: {"1001": {}}},
		Actions:  []*utils.APAction{{}},
	}

	if err := dm.SetActionProfile(context.Background(), ap, false); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}
}

func TestDMSetActionProfileSetActionProfileDrvErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetActionProfileDrvF: func(ctx *context.Context, tenant, ID string) (*utils.ActionProfile, error) {
			return &utils.ActionProfile{}, nil
		},
		SetActionProfileDrvF: func(ctx *context.Context, ap *utils.ActionProfile) error { return utils.ErrNotImplemented },
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	ap := &utils.ActionProfile{

		Tenant:    "cgrates.org",
		ID:        "AP1",
		FilterIDs: []string{"*string*req.Account1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 65,
			},
		},
		Schedule: "* * * * *",
		Targets:  map[string]utils.StringSet{utils.MetaAccounts: {"1001": {}}},
		Actions:  []*utils.APAction{{}},
	}
	if err := dm.SetActionProfile(context.Background(), ap, false); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}
}

func TestDMSetActionProfileUpdatedIndexesErr(t *testing.T) {

	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetActionProfileDrvF: func(ctx *context.Context, tenant, ID string) (*utils.ActionProfile, error) {
			return &utils.ActionProfile{}, nil
		},
		SetActionProfileDrvF: func(ctx *context.Context, ap *utils.ActionProfile) error { return nil },
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	ap := &utils.ActionProfile{

		Tenant:    "cgrates.org",
		ID:        "AP1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 65,
			},
		},
		Schedule: "* * * * *",
		Targets:  map[string]utils.StringSet{utils.MetaAccounts: {"1001": {}}},
		Actions:  []*utils.APAction{{}},
	}

	if err := dm.SetActionProfile(context.Background(), ap, true); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}
}

func TestDMSetActionProfileReplicate(t *testing.T) {

	cfgtmp := config.CgrConfig()
	defer func() {
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	ap := &utils.ActionProfile{

		Tenant:    "cgrates.org",
		ID:        "AP1",
		FilterIDs: []string{"*string*req.Account1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 65,
			},
		},
		Schedule: "* * * * *",
		Targets:  map[string]utils.StringSet{utils.MetaAccounts: {"1001": {}}},
		Actions:  []*utils.APAction{{}},
	}

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaActionProfiles].Replicate = true
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	config.SetCgrConfig(cfg)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1SetActionProfile: func(ctx *context.Context, args, reply any) error { return utils.ErrNotImplemented },
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.ReplicatorSv1, cc)
	data := &DataDBMock{
		GetActionProfileDrvF: func(ctx *context.Context, tenant, ID string) (*utils.ActionProfile, error) {
			return ap, nil
		},
		SetActionProfileDrvF: func(ctx *context.Context, ap *utils.ActionProfile) error { return nil },
	}
	dm := NewDataManager(data, cfg, cM)

	// tests replicate
	if err := dm.SetActionProfile(context.Background(), ap, false); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestDMSetRateProfileNoDMErr(t *testing.T) {
	var dm *DataManager
	err := dm.SetRateProfile(context.Background(), &utils.RateProfile{}, false, false)
	if err != utils.ErrNoDatabaseConn {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrNoDatabaseConn, err)
	}
}

func TestDMSetRateProfileRatesProfileCheckFiltersErr(t *testing.T) {

	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	rpp := &utils.RateProfile{
		ID:        "test_ID1",
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string*req.Account1001"},
		Rates: map[string]*utils.Rate{
			"RT1": {
				ID:        "RT1",
				FilterIDs: []string{"*string*req.Account1001"},
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(1, 2),
						Unit:          utils.NewDecimal(int64(time.Second), 0),
						Increment:     utils.NewDecimal(int64(time.Second), 0),
					},
				},
			},
		},
	}

	expErr := "broken reference to filter: <*string*req.Account1001> for item with ID: cgrates.org:test_ID1"
	if err := dm.SetRateProfile(context.Background(), rpp, false, false); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, received error <%v>", expErr, err)
	}

}

func TestDMSetRateProfileRatesCheckFiltersErr(t *testing.T) {

	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	rpp := &utils.RateProfile{
		ID:        "test_ID1",
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Rates: map[string]*utils.Rate{
			"RT1": {
				ID:        "RT1",
				FilterIDs: []string{"*string*req.Account1001"},
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(1, 2),
						Unit:          utils.NewDecimal(int64(time.Second), 0),
						Increment:     utils.NewDecimal(int64(time.Second), 0),
					},
				},
			},
		},
	}

	expErr := "broken reference to filter: <*string*req.Account1001> for item with ID: RT1"
	if err := dm.SetRateProfile(context.Background(), rpp, false, false); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, received error <%v>", expErr, err)
	}

}

func TestDMSetRateProfileGetRateProfileErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetRateProfileDrvF: func(ctx *context.Context, s1, s2 string) (*utils.RateProfile, error) {
			return &utils.RateProfile{}, utils.ErrNotImplemented
		},
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	rpp := &utils.RateProfile{
		ID:        "test_ID1",
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:*req.Account:1001"},
		Rates: map[string]*utils.Rate{
			"RT1": {
				ID: "RT1",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(1, 2),
						Unit:          utils.NewDecimal(int64(time.Second), 0),
						Increment:     utils.NewDecimal(int64(time.Second), 0),
					},
				},
			},
		},
	}

	if err := dm.SetRateProfile(context.Background(), rpp, false, false); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}
}

func TestDMSetRateProfileUpdatedIndexesErr(t *testing.T) {

	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetRateProfileDrvF: func(ctx *context.Context, s1, s2 string) (*utils.RateProfile, error) {
			return &utils.RateProfile{}, nil
		},
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	rpp := &utils.RateProfile{
		ID:        "test_ID1",
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:*req.Account:1001"},
		Rates: map[string]*utils.Rate{
			"RT1": {
				ID: "RT1",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(1, 2),
						Unit:          utils.NewDecimal(int64(time.Second), 0),
						Increment:     utils.NewDecimal(int64(time.Second), 0),
					},
				},
			},
		},
	}

	if err := dm.SetRateProfile(context.Background(), rpp, true, true); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}
}

func TestDMSetRateProfileRatesSetRateProfileDrvErr(t *testing.T) {
	Cache.Clear(nil)

	rpp := &utils.RateProfile{
		ID:        "test_ID1",
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Rates:     map[string]*utils.Rate{},
	}

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetRateProfileDrvF: func(ctx *context.Context, s1, s2 string) (*utils.RateProfile, error) {
			return rpp, nil
		},
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	if err := dm.SetRateProfile(context.Background(), rpp, false, false); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}
}

func TestDMSetRateProfileReplicate(t *testing.T) {

	cfgtmp := config.CgrConfig()
	defer func() {
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	rpp := &utils.RateProfile{
		ID:        "test_ID1",
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Rates:     map[string]*utils.Rate{},
	}

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaRateProfiles].Replicate = true
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	config.SetCgrConfig(cfg)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1SetRateProfile: func(ctx *context.Context, args, reply any) error { return utils.ErrNotImplemented },
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.ReplicatorSv1, cc)
	data := &DataDBMock{
		GetRateProfileDrvF: func(ctx *context.Context, s1, s2 string) (*utils.RateProfile, error) {
			return rpp, nil
		},
		SetRateProfileDrvF: func(ctx *context.Context, rp *utils.RateProfile, b bool) error { return nil },
	}
	dm := NewDataManager(data, cfg, cM)

	// tests replicate
	if err := dm.SetRateProfile(context.Background(), rpp, false, false); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestDMGetActionProfileCacheGetErr(t *testing.T) {

	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	if err := Cache.Set(context.Background(), utils.CacheActionProfiles, utils.ConcatenatedKey(utils.CGRateSorg, "ap1"), nil, []string{}, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	_, err := dm.GetActionProfile(context.Background(), utils.CGRateSorg, "ap1", true, false, utils.NonTransactional)
	if err != utils.ErrNotFound {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotFound, err)
	}
}

func TestDMGetActionProfileCacheGet(t *testing.T) {

	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	val := &utils.ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "AP1",
		FilterIDs: []string{"*string*req.Account1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 65,
			},
		},
		Schedule: "* * * * *",
		Targets:  map[string]utils.StringSet{utils.MetaAccounts: {"1001": {}}},
		Actions:  []*utils.APAction{{}}}

	if err := Cache.Set(context.Background(), utils.CacheActionProfiles, utils.ConcatenatedKey(utils.CGRateSorg, "ap1"), val, []string{}, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	if rcv, err := dm.GetActionProfile(context.Background(), utils.CGRateSorg, "ap1", true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, val) {
		t.Errorf("Expected <%v>, received <%v>", val, rcv)
	}
}

func TestDMGetActionProfileNilDmErr(t *testing.T) {

	var dm *DataManager

	_, err := dm.GetActionProfile(context.Background(), utils.CGRateSorg, "ap1", false, false, utils.NonTransactional)
	if err != utils.ErrNoDatabaseConn {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNoDatabaseConn, err)
	}
}

func TestDMGetActionProfileSetActionProfileDrvErr(t *testing.T) {

	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
		config.SetCgrConfig(cfgtmp)
	}()

	ap := &utils.ActionProfile{

		Tenant:    "cgrates.org",
		ID:        "AP1",
		FilterIDs: []string{"*string*req.Account1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 65,
			},
		},
		Schedule: "* * * * *",
		Targets:  map[string]utils.StringSet{utils.MetaAccounts: {"1001": {}}},
		Actions:  []*utils.APAction{{}},
	}

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaActionProfiles].Remote = true
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg)}
	config.SetCgrConfig(cfg)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetActionProfile: func(ctx *context.Context, args, reply any) error { return nil },
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg), utils.ReplicatorSv1, cc)
	data := &DataDBMock{
		GetActionProfileDrvF: func(ctx *context.Context, tenant, ID string) (*utils.ActionProfile, error) {
			return ap, utils.ErrNotFound
		},
		SetActionProfileDrvF: func(ctx *context.Context, ap *utils.ActionProfile) error { return utils.ErrNotImplemented },
	}
	dm := NewDataManager(data, cfg, cM)

	_, err := dm.GetActionProfile(context.Background(), utils.CGRateSorg, ap.ID, false, false, utils.NonTransactional)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}
}

func TestDMGetActionProfileCacheWriteErr1(t *testing.T) {

	ap := &utils.ActionProfile{

		Tenant:    "cgrates.org",
		ID:        "AP1",
		FilterIDs: []string{"*string*req.Account1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 65,
			},
		},
		Schedule: "* * * * *",
		Targets:  map[string]utils.StringSet{utils.MetaAccounts: {"1001": {}}},
		Actions:  []*utils.APAction{{}},
	}

	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
		config.SetCgrConfig(cfgtmp)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.CacheCfg().Partitions[utils.CacheActionProfiles].Replicate = true
	config.SetCgrConfig(cfg)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error {

				return utils.ErrNotImplemented
			},
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.CacheSv1, cc)

	data := &DataDBMock{
		GetActionProfileDrvF: func(ctx *context.Context, tenant, ID string) (*utils.ActionProfile, error) {
			return ap, utils.ErrNotFound
		},
	}
	dm := NewDataManager(data, cfg, cM)

	Cache = NewCacheS(cfg, dm, cM, nil)

	_, err := dm.GetActionProfile(context.Background(), utils.CGRateSorg, ap.ID, false, true, utils.NonTransactional)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}
}

func TestDMGetActionProfileCacheWriteErr2(t *testing.T) {

	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
		config.SetCgrConfig(cfgtmp)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.CacheCfg().Partitions[utils.CacheActionProfiles].Replicate = true
	config.SetCgrConfig(cfg)

	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error { return utils.ErrNotImplemented },
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.CacheSv1, cc)
	dm := NewDataManager(data, cfg, cM)

	ap := &utils.ActionProfile{

		Tenant:    "cgrates.org",
		ID:        "AP1",
		FilterIDs: []string{"*string*req.Account1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 65,
			},
		},
		Schedule: "* * * * *",
		Targets:  map[string]utils.StringSet{utils.MetaAccounts: {"1001": {}}},
		Actions:  []*utils.APAction{{}},
	}

	if err := dm.dataDB.SetActionProfileDrv(context.Background(), ap); err != nil {
		t.Error(err)
	}

	Cache = NewCacheS(cfg, dm, cM, nil)

	if _, err := dm.GetActionProfile(context.Background(), utils.CGRateSorg, ap.ID, false, true, utils.NonTransactional); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestDMGetAttributeProfileCacheGetErr(t *testing.T) {

	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	if err := Cache.Set(context.Background(), utils.CacheAttributeProfiles, utils.ConcatenatedKey(utils.CGRateSorg, "ap1"), nil, []string{}, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	_, err := dm.GetAttributeProfile(context.Background(), utils.CGRateSorg, "ap1", true, false, utils.NonTransactional)
	if err != utils.ErrNotFound {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotFound, err)
	}
}

func TestDMGetAttributeProfileCacheWriteErr1(t *testing.T) {

	attrPrfl := &utils.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_ID",
		FilterIDs: []string{"filtrId1"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Account",
				Value: utils.NewRSRParsersMustCompile("1001", utils.InfieldSep),
			},
		},
		Weights: make(utils.DynamicWeights, 1),
	}

	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
		config.SetCgrConfig(cfgtmp)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.CacheCfg().Partitions[utils.CacheAttributeProfiles].Replicate = true
	config.SetCgrConfig(cfg)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error {

				return utils.ErrNotImplemented
			},
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.CacheSv1, cc)

	data := &DataDBMock{
		GetAttributeProfileDrvF: func(ctx *context.Context, str1, str2 string) (*utils.AttributeProfile, error) {
			return attrPrfl, utils.ErrNotFound
		},
	}
	dm := NewDataManager(data, cfg, cM)

	Cache = NewCacheS(cfg, dm, cM, nil)

	_, err := dm.GetAttributeProfile(context.Background(), utils.CGRateSorg, attrPrfl.ID, false, true, utils.NonTransactional)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}
}

func TestDMGetAttributeProfileCacheWriteErr2(t *testing.T) {

	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
		config.SetCgrConfig(cfgtmp)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.CacheCfg().Partitions[utils.CacheAttributeProfiles].Replicate = true
	config.SetCgrConfig(cfg)

	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error { return utils.ErrNotImplemented },
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.CacheSv1, cc)
	dm := NewDataManager(data, cfg, cM)

	attrPrfl := &utils.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_ID",
		FilterIDs: []string{"filtrId1"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Account",
				Value: utils.NewRSRParsersMustCompile("1001", utils.InfieldSep),
			},
		},
		Weights: make(utils.DynamicWeights, 1),
	}

	if err := dm.dataDB.SetAttributeProfileDrv(context.Background(), attrPrfl); err != nil {
		t.Error(err)
	}

	Cache = NewCacheS(cfg, dm, cM, nil)

	if _, err := dm.GetAttributeProfile(context.Background(), utils.CGRateSorg, attrPrfl.ID, false, true, utils.NonTransactional); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestDMGGetChargerProfileCacheGetErr(t *testing.T) {

	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	if err := Cache.Set(context.Background(), utils.CacheChargerProfiles, utils.ConcatenatedKey(utils.CGRateSorg, "chp1"), nil, []string{}, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	_, err := dm.GetChargerProfile(context.Background(), utils.CGRateSorg, "chp1", true, false, utils.NonTransactional)
	if err != utils.ErrNotFound {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotFound, err)
	}
}

func TestDMGetChargerProfileNilDmErr(t *testing.T) {

	var dm *DataManager

	_, err := dm.GetChargerProfile(context.Background(), utils.CGRateSorg, "chp1", false, false, utils.NonTransactional)
	if err != utils.ErrNoDatabaseConn {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNoDatabaseConn, err)
	}
}

func TestDMSetStatQueueProfileUpdatedIndexesErr(t *testing.T) {

	defer func() { Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil) }()

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetStatQueueProfileDrvF: func(ctx *context.Context, tenant, id string) (sq *StatQueueProfile, err error) {
			return &StatQueueProfile{}, nil
		},
		SetStatQueueProfileDrvF: func(ctx *context.Context, sq *StatQueueProfile) (err error) { return nil },
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	sqp := &StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "sqp003",
		FilterIDs:   []string{"*string:~*req.Account:1001"},
		QueueLength: 10,
		TTL:         10 * time.Second,
		Metrics: []*MetricWithFilters{
			{
				MetricID: "*sum#~*req.Usage",
			},
		},
		ThresholdIDs: []string{},
		Stored:       true,
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		MinItems: 1,
	}

	if err := dm.SetStatQueueProfile(context.Background(), sqp, true); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestDMSetStatQueueProfileReplicate(t *testing.T) {

	cfgtmp := config.CgrConfig()
	defer func() {
		config.SetCgrConfig(cfgtmp)
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	sqp := &StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "sqp003",
		FilterIDs:   []string{"*string:~*req.Account:1001"},
		QueueLength: 10,
		TTL:         10 * time.Second,
		Metrics: []*MetricWithFilters{
			{
				MetricID: "*sum#~*req.Usage",
			},
		},
		ThresholdIDs: []string{},
		Stored:       true,
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		MinItems: 1,
	}

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaStatQueueProfiles].Replicate = true
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	config.SetCgrConfig(cfg)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1SetStatQueueProfile: func(ctx *context.Context, args, reply any) error { return utils.ErrNotImplemented },
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.ReplicatorSv1, cc)
	data := &DataDBMock{

		GetStatQueueProfileDrvF: func(ctx *context.Context, tenant, id string) (sq *StatQueueProfile, err error) {
			return sqp, nil
		},
		SetStatQueueProfileDrvF: func(ctx *context.Context, sq *StatQueueProfile) (err error) { return nil },
	}
	dm := NewDataManager(data, cfg, cM)

	// tests replicate
	if err := dm.SetStatQueueProfile(context.Background(), sqp, false); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestDMSetStatQueueProfileNewStatQueueNilOldStsErr(t *testing.T) {

	defer func() {

		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	var minItems int
	sqp := &StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "sqp003",
		FilterIDs:   []string{"*string:~*req.Account:1001"},
		QueueLength: 10,
		TTL:         10 * time.Second,
		Metrics: []*MetricWithFilters{
			{
				MetricID: "invalid",
			},
		},
		ThresholdIDs: []string{},
		Stored:       true,
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		MinItems: minItems,
	}

	cfg := config.NewDefaultCGRConfig()
	cM := NewConnManager(cfg)
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg, cM)

	experr := fmt.Sprintf("unsupported metric type <%s>", sqp.Metrics[0].MetricID)

	if err := dm.SetStatQueueProfile(context.Background(), sqp, false); err == nil || err.Error() != experr {
		t.Errorf("\nexpected error: %v, \nreceived error: %v", experr, err)
	}

}

func TestDMSetStatQueueProfileNewStatQueueErr(t *testing.T) {

	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	var minItems int
	sqp := &StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "sqp003",
		FilterIDs:   []string{"*string:~*req.Account:1001"},
		QueueLength: 10,
		TTL:         10 * time.Second,
		Metrics: []*MetricWithFilters{
			{
				MetricID: "invalid",
			},
		},
		ThresholdIDs: []string{},
		Stored:       true,
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		MinItems: minItems,
	}

	cfg := config.NewDefaultCGRConfig()
	cM := NewConnManager(cfg)
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg, cM)

	if err := dm.dataDB.SetStatQueueProfileDrv(context.Background(), sqp); err != nil {
		t.Error(err)
	}

	experr := fmt.Sprintf("unsupported metric type <%s>", sqp.Metrics[0].MetricID)
	if err := dm.SetStatQueueProfile(context.Background(), sqp, false); err == nil || err.Error() != experr {
		t.Errorf("\nexpected error: %v, \nreceived error: %v", experr, err)
	}

}

func TestDMRemoveStatQueueProfileNilDMErr(t *testing.T) {

	var dm *DataManager

	err := dm.RemoveStatQueueProfile(context.Background(), "tnt", "Id", false)
	if err != utils.ErrNoDatabaseConn {
		t.Errorf("\nExpected error <%+v>, \nReceived error <%+v>", utils.ErrNoDatabaseConn, err)
	}

}

func TestDMRemoveStatQueueProfileGetStatQueueProfileErr(t *testing.T) {

	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetStatQueueProfileDrvF: func(ctx *context.Context, tenant, id string) (sq *StatQueueProfile, err error) {
			return &StatQueueProfile{}, utils.ErrNotImplemented
		},
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	err := dm.RemoveStatQueueProfile(context.Background(), "tnt", "Id", false)
	if err != utils.ErrNotImplemented {
		t.Errorf("\nExpected error <%+v>, \nReceived error <%+v>", utils.ErrNotImplemented, err)
	}

}

func TestDMRemoveStatQueueProfileRemStatQueueProfileDrvErr(t *testing.T) {

	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetStatQueueProfileDrvF: func(ctx *context.Context, tenant, id string) (sq *StatQueueProfile, err error) {
			return &StatQueueProfile{}, nil
		},
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	sqp := &StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "sqp003",
		FilterIDs:   []string{"*string:~*req.Account:1001"},
		QueueLength: 10,
		TTL:         10 * time.Second,
		Metrics: []*MetricWithFilters{
			{
				MetricID: "*sum#~*req.Usage",
			},
		},
		ThresholdIDs: []string{},
		Stored:       true,
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		MinItems: 1,
	}

	err := dm.RemoveStatQueueProfile(context.Background(), sqp.Tenant, sqp.ID, true)
	if err != utils.ErrNotImplemented {
		t.Errorf("\nExpected error <%+v>, \nReceived error <%+v>", utils.ErrNotImplemented, err)
	}

}

func TestDMRemoveStatQueueProfileNilOldStsErr(t *testing.T) {

	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	var Id string
	var tnt string
	err := dm.RemoveStatQueueProfile(context.Background(), tnt, Id, false)
	if err != utils.ErrNotFound {
		t.Errorf("\nExpected error <%+v>, \nReceived error <%+v>", utils.ErrNotFound, err)
	}

}

func TestDMRemoveStatQueueProfileRmvItemFromFiltrIndexErr(t *testing.T) {

	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetStatQueueProfileDrvF: func(ctx *context.Context, tenant, id string) (sq *StatQueueProfile, err error) {
			return &StatQueueProfile{}, nil
		},
		RemStatQueueProfileDrvF: func(ctx *context.Context, tenant, ID string) error { return nil },
	}
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	sqp := &StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "sqp003",
		FilterIDs:   []string{"*string:~*req.Account:1001"},
		QueueLength: 10,
		TTL:         10 * time.Second,
		Metrics: []*MetricWithFilters{
			{
				MetricID: "*sum#~*req.Usage",
			},
		},
		ThresholdIDs: []string{},
		Stored:       true,
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		MinItems: 1,
	}

	err := dm.RemoveStatQueueProfile(context.Background(), sqp.Tenant, sqp.ID, true)
	if err != utils.ErrNotImplemented {
		t.Errorf("\nExpected error <%+v>, \nReceived error <%+v>", utils.ErrNotImplemented, err)
	}

}

func TestDMRemoveStatQueueProfileRmvIndexFiltersItemErr(t *testing.T) {

	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	sqp := &StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "sqp003",
		FilterIDs:   []string{"fltrID1"},
		QueueLength: 10,
		TTL:         10 * time.Second,
		Metrics: []*MetricWithFilters{
			{
				MetricID: "*sum#~*req.Usage",
			},
		},
		ThresholdIDs: []string{},
		Stored:       true,
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		MinItems: 1,
	}

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{
		GetStatQueueProfileDrvF: func(ctx *context.Context, tenant, id string) (sq *StatQueueProfile, err error) {
			return sqp, nil
		},
		RemStatQueueProfileDrvF: func(ctx *context.Context, tenant, ID string) error { return nil },
	}

	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	err := dm.RemoveStatQueueProfile(context.Background(), sqp.Tenant, sqp.ID, true)
	if err != utils.ErrNotImplemented {
		t.Errorf("\nExpected error <%+v>, \nReceived error <%+v>", utils.ErrNotImplemented, err)
	}

}

func TestDMRemoveStatQueueProfileReplicate(t *testing.T) {

	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
		config.SetCgrConfig(cfgtmp)
	}()

	sqp := &StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "sqp003",
		FilterIDs:   []string{"*string:~*req.Account:1001"},
		QueueLength: 10,
		TTL:         10 * time.Second,
		Metrics: []*MetricWithFilters{
			{
				MetricID: "*sum#~*req.Usage",
			},
		},
		ThresholdIDs: []string{},
		Stored:       true,
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		MinItems: 1,
	}

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaStatQueueProfiles].Replicate = true
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	config.SetCgrConfig(cfg)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1RemoveStatQueueProfile: func(ctx *context.Context, args, reply any) error { return utils.ErrNotImplemented },
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.ReplicatorSv1, cc)
	data := &DataDBMock{
		GetStatQueueProfileDrvF: func(ctx *context.Context, tenant, id string) (sq *StatQueueProfile, err error) {
			return sqp, nil
		},
		RemStatQueueProfileDrvF: func(ctx *context.Context, tenant, ID string) error { return nil },
	}
	dm := NewDataManager(data, cfg, cM)

	// tests replicate
	dm.RemoveStatQueueProfile(context.Background(), sqp.Tenant, sqp.ID, false)

}

func TestDMGetStatQueueCacheWriteErr1(t *testing.T) {

	sq := &StatQueue{
		Tenant: "cgrates.org",
		ID:     "sq01",
		SQItems: []SQItem{
			{
				EventID: "SqProcessEvent",
			},
		},
		SQMetrics: make(map[string]StatMetric),
	}

	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
		config.SetCgrConfig(cfgtmp)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.CacheCfg().Partitions[utils.CacheStatQueues].Replicate = true
	config.SetCgrConfig(cfg)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error {

				return utils.ErrNotImplemented
			},
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.CacheSv1, cc)

	data := &DataDBMock{
		GetStatQueueDrvF: func(ctx *context.Context, tenant, id string) (sq *StatQueue, err error) { return sq, utils.ErrNotFound },
	}
	dm := NewDataManager(data, cfg, cM)

	Cache = NewCacheS(cfg, dm, cM, nil)

	_, err := dm.GetStatQueue(context.Background(), utils.CGRateSorg, sq.ID, false, true, utils.NonTransactional)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}
}

func TestDMGetChargerProfileSetChargerProfileDrvErr(t *testing.T) {

	cfgtmp := config.CgrConfig()
	defer func() {
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cpp := &utils.ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "CHP_1",
		FilterIDs:    []string{"*string:~*req.Account:1001"},
		RunID:        "*rated",
		AttributeIDs: []string{"ATTR_1"},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaChargerProfiles].Remote = true
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg)}
	config.SetCgrConfig(cfg)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetChargerProfile: func(ctx *context.Context, args, reply any) error { return nil },
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg), utils.ReplicatorSv1, cc)
	data := &DataDBMock{
		GetChargerProfileDrvF: func(ctx *context.Context, tnt, id string) (*utils.ChargerProfile, error) {
			return cpp, utils.ErrNotFound
		},
		SetChargerProfileDrvF: func(ctx *context.Context, chr *utils.ChargerProfile) (err error) { return utils.ErrNotImplemented },
	}
	dm := NewDataManager(data, cfg, cM)

	_, err := dm.GetChargerProfile(context.Background(), utils.CGRateSorg, cpp.ID, false, false, utils.NonTransactional)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}
}

func TestDMGetChargerProfileCacheWriteErr1(t *testing.T) {

	cpp := &utils.ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "CHP_1",
		FilterIDs:    []string{"*string:~*req.Account:1001"},
		RunID:        "*rated",
		AttributeIDs: []string{"ATTR_1"},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}

	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
		config.SetCgrConfig(cfgtmp)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.CacheCfg().Partitions[utils.CacheChargerProfiles].Replicate = true
	config.SetCgrConfig(cfg)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error {

				return utils.ErrNotImplemented
			},
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.CacheSv1, cc)

	data := &DataDBMock{
		GetChargerProfileDrvF: func(ctx *context.Context, tnt, id string) (*utils.ChargerProfile, error) {
			return cpp, utils.ErrNotFound
		},
	}
	dm := NewDataManager(data, cfg, cM)

	Cache = NewCacheS(cfg, dm, cM, nil)

	_, err := dm.GetChargerProfile(context.Background(), utils.CGRateSorg, cpp.ID, false, true, utils.NonTransactional)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}
}

func TestDMGetChargerProfileCacheWriteErr2(t *testing.T) {

	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
		config.SetCgrConfig(cfgtmp)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.CacheCfg().Partitions[utils.CacheChargerProfiles].Replicate = true
	config.SetCgrConfig(cfg)

	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error { return utils.ErrNotImplemented },
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.CacheSv1, cc)
	dm := NewDataManager(data, cfg, cM)

	cpp := &utils.ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "CHP_1",
		FilterIDs:    []string{"*string:~*req.Account:1001"},
		RunID:        "*rated",
		AttributeIDs: []string{"ATTR_1"},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}

	if err := dm.dataDB.SetChargerProfileDrv(context.Background(), cpp); err != nil {
		t.Error(err)
	}

	Cache = NewCacheS(cfg, dm, cM, nil)

	if _, err := dm.GetChargerProfile(context.Background(), utils.CGRateSorg, cpp.ID, false, true, utils.NonTransactional); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}
}

func TestDMGetItemLoadIDsNilDM(t *testing.T) {

	var dm *DataManager

	_, err := dm.GetItemLoadIDs(context.Background(), "", false)
	if err != utils.ErrNoDatabaseConn {
		t.Errorf("\nExpected error <%+v>, \nReceived error <%+v>", utils.ErrNoDatabaseConn, err)
	}

}

func TestDMGetItemLoadIDsSetSetLoadIDsDrvErr(t *testing.T) {

	cfgtmp := config.CgrConfig()
	defer func() {
		config.SetCgrConfig(cfgtmp)
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items[utils.MetaLoadIDs].Remote = true
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg)}
	config.SetCgrConfig(cfg)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetItemLoadIDs: func(ctx *context.Context, args, reply any) error { return nil },
		},
	}

	itmLIDs := map[string]int64{
		"ID_1": 21,
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg), utils.ReplicatorSv1, cc)
	data := &DataDBMock{
		GetItemLoadIDsDrvF: func(ctx *context.Context, itemIDPrefix string) (loadIDs map[string]int64, err error) {
			return itmLIDs, utils.ErrNotFound
		},
		SetLoadIDsDrvF: func(ctx *context.Context, loadIDs map[string]int64) error { return utils.ErrNotImplemented },
	}
	dm := NewDataManager(data, cfg, cM)

	_, err := dm.GetItemLoadIDs(context.Background(), cfg.GeneralCfg().DefaultTenant, false)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}
}

func TestDMGetItemLoadIDsCacheWriteErr1(t *testing.T) {

	itmLIDs := map[string]int64{
		"ID_1": 21,
	}

	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
		config.SetCgrConfig(cfgtmp)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.CacheCfg().Partitions[utils.CacheLoadIDs].Replicate = true
	config.SetCgrConfig(cfg)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error {

				return utils.ErrNotImplemented
			},
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.CacheSv1, cc)

	data := &DataDBMock{
		GetItemLoadIDsDrvF: func(ctx *context.Context, itemIDPrefix string) (loadIDs map[string]int64, err error) {
			return itmLIDs, utils.ErrNotFound
		},
	}

	dm := NewDataManager(data, cfg, cM)

	Cache = NewCacheS(cfg, dm, cM, nil)

	_, err := dm.GetItemLoadIDs(context.Background(), cfg.GeneralCfg().DefaultTenant, true)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}
}

func TestDMGetItemLoadIDsCacheWriteErr2(t *testing.T) {

	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
		config.SetCgrConfig(cfgtmp)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.CacheCfg().Partitions[utils.CacheLoadIDs].Replicate = true
	config.SetCgrConfig(cfg)

	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error { return utils.ErrNotImplemented },
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.CacheSv1, cc)
	dm := NewDataManager(data, cfg, cM)

	itmLIDs := map[string]int64{
		"ID_1": 21,
	}

	if err := dm.dataDB.SetLoadIDsDrv(context.Background(), itmLIDs); err != nil {
		t.Error(err)
	}

	Cache = NewCacheS(cfg, dm, cM, nil)

	_, err := dm.GetItemLoadIDs(context.Background(), cfg.GeneralCfg().DefaultTenant, true)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestDMGetRateProfileCacheGetOK(t *testing.T) {

	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	rP := &utils.RateProfile{
		ID:        "test_ID1",
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:*req.Account:1001"},
		Rates: map[string]*utils.Rate{
			"RT1": {
				ID: "RT1",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(1, 2),
						Unit:          utils.NewDecimal(int64(time.Second), 0),
						Increment:     utils.NewDecimal(int64(time.Second), 0),
					},
				},
			},
		},
	}

	if err := Cache.Set(context.Background(), utils.CacheRateProfiles, utils.ConcatenatedKey(utils.CGRateSorg, "rp1"), rP, []string{}, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	if rcv, err := dm.GetRateProfile(context.Background(), utils.CGRateSorg, "rp1", true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, rP) {
		t.Errorf("Expected <%v>, received <%v>", rP, rcv)
	}
}

func TestDMGetRateProfileCacheGetErr(t *testing.T) {

	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, cfg, cM)

	if err := Cache.Set(context.Background(), utils.CacheRateProfiles, utils.ConcatenatedKey(utils.CGRateSorg, "rp1"), nil, []string{}, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	_, err := dm.GetRateProfile(context.Background(), utils.CGRateSorg, "rp1", true, false, utils.NonTransactional)
	if err != utils.ErrNotFound {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNotFound, err)
	}
}

func TestDMGetRateProfileNildm(t *testing.T) {

	var dm *DataManager

	_, err := dm.GetRateProfile(context.Background(), utils.CGRateSorg, "rp1", false, false, utils.NonTransactional)
	if err != utils.ErrNoDatabaseConn {
		t.Errorf("Expected error <%v>, received error <%v>", utils.ErrNoDatabaseConn, err)
	}
}

func TestDMResourcesUpdateResource(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	idb, err := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dm := NewDataManager(idb, cfg, nil)
	Cache.Clear(nil)
	res := &utils.ResourceProfile{
		Tenant:   "cgrates.org",
		ID:       "RES1",
		UsageTTL: 0,
		Limit:    10,
		Stored:   true,
	}
	r := &utils.Resource{
		Tenant: res.Tenant,
		ID:     res.ID,
		Usages: map[string]*utils.ResourceUsage{
			"jkbdfgs": {
				Tenant:     res.Tenant,
				ID:         "jkbdfgs",
				ExpiryTime: time.Now(),
				Units:      5,
			},
		},
		TTLIdx: []string{"jkbdfgs"},
	}
	expR := &utils.Resource{
		Tenant: res.Tenant,
		ID:     res.ID,
		Usages: make(map[string]*utils.ResourceUsage),
	}
	if err := dm.SetResourceProfile(context.Background(), res, true); err != nil {
		t.Fatal(err)
	}

	if r, err := dm.GetResource(context.Background(), res.Tenant, res.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(r, expR) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expR), utils.ToJSON(r))
	}

	if err := dm.RemoveResource(context.Background(), r.Tenant, r.ID); err != nil {
		t.Fatal(err)
	}

	if err := dm.SetResourceProfile(context.Background(), res, true); err != nil {
		t.Fatal(err)
	}

	if r, err := dm.GetResource(context.Background(), res.Tenant, res.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(r, expR) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expR), utils.ToJSON(r))
	}

	if err := dm.SetResource(context.Background(), r); err != nil {
		t.Fatal(err)
	}

	res = &utils.ResourceProfile{
		Tenant:   "cgrates.org",
		ID:       "RES1",
		UsageTTL: 0,
		Limit:    5,
		Stored:   true,
	}
	if err := dm.SetResourceProfile(context.Background(), res, true); err != nil {
		t.Fatal(err)
	}
	if r, err := dm.GetResource(context.Background(), res.Tenant, res.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(r, expR) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expR), utils.ToJSON(r))
	}

	if err := dm.SetResource(context.Background(), r); err != nil {
		t.Fatal(err)
	}

	res = &utils.ResourceProfile{
		Tenant:   "cgrates.org",
		ID:       "RES1",
		UsageTTL: 10,
		Limit:    5,
		Stored:   true,
	}
	if err := dm.SetResourceProfile(context.Background(), res, true); err != nil {
		t.Fatal(err)
	}
	if r, err := dm.GetResource(context.Background(), res.Tenant, res.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(r, expR) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expR), utils.ToJSON(r))
	}

	if err := dm.SetResource(context.Background(), r); err != nil {
		t.Fatal(err)
	}

	res = &utils.ResourceProfile{
		Tenant:   "cgrates.org",
		ID:       "RES1",
		UsageTTL: 10,
		Limit:    5,
		Stored:   false,
	}
	if err := dm.SetResourceProfile(context.Background(), res, true); err != nil {
		t.Fatal(err)
	}
	if r, err := dm.GetResource(context.Background(), res.Tenant, res.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(r, expR) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expR), utils.ToJSON(r))
	}

	if err := dm.RemoveResourceProfile(context.Background(), res.Tenant, res.ID, true); err != nil {
		t.Fatal(err)
	}

	if _, err := dm.GetResource(context.Background(), res.Tenant, res.ID, false, false, utils.NonTransactional); err != utils.ErrNotFound {
		t.Fatal(err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestDMGetTrend(t *testing.T) {
	var dm *DataManager
	if _, err := dm.GetTrend(context.Background(), utils.CGRateSorg, "TrendNil", false, false, utils.NonTransactional); err != utils.ErrNoDatabaseConn {
		t.Errorf("expected ErrNoDatabaseConn, got %v", err)
	}

	tmp := Cache
	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)
	}()
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm = NewDataManager(data, cfg, cM)

	tntID := utils.ConcatenatedKey(utils.CGRateSorg, "TrendCacheNil")
	if err := Cache.Set(context.Background(), utils.CacheTrends, tntID, nil, []string{}, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}
	if _, err := dm.GetTrend(context.Background(), utils.CGRateSorg, "TrendCacheNil", true, false, utils.NonTransactional); err != utils.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}

	Cache.Clear(nil)
	tr := &utils.Trend{
		Tenant: "cgrates.org",
		ID:     "TrendOk",
		RunTimes: []time.Time{
			time.Now(),
		},
		Metrics: map[time.Time]map[string]*utils.MetricWithTrend{
			time.Now(): {
				"metric1": {
					ID:          "metric1",
					Value:       5,
					TrendGrowth: 0.5,
					TrendLabel:  "*positive",
				},
			},
		},
	}
	if err := dm.dataDB.SetTrendDrv(context.Background(), tr); err != nil {
		t.Fatalf("failed SetTrendDrv: %v", err)
	}
	got, err := dm.GetTrend(context.Background(), utils.CGRateSorg, "TrendOk", false, true, utils.NonTransactional)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "TrendOk" {
		t.Fatalf("expected TrendOk, got %s", got.ID)
	}

	Cache.Clear(nil)
	if _, err := dm.GetTrend(context.Background(), utils.CGRateSorg, "TrendNotFound", false, true, utils.NonTransactional); err != nil && err != utils.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}

	Cache.Clear(nil)
	cfg.DataDbCfg().Items[utils.MetaTrends].Remote = true
	cfg.DataDbCfg().RmtConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg)}
	config.SetCgrConfig(cfg)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1GetTrend: func(ctx *context.Context, args, reply any) error {
				return utils.ErrNotFound
			},
		},
	}
	cM = NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.RemoteConnsCfg), utils.ReplicatorSv1, cc)

	dm = NewDataManager(data, cfg, cM)
	if _, err := dm.GetTrend(context.Background(), utils.CGRateSorg, "TrendRemote", false, true, utils.NonTransactional); err != nil && err != utils.ErrNotFound {
		t.Errorf("expected ErrNotFound or nil, got %v", err)
	}
}

func TestDataDBGetIPProfileDrv(t *testing.T) {
	ctx := &context.Context{}
	tenant := "cgrates.org"
	id := "ip1"

	dbNil := &DataDBMock{}
	profile, err := dbNil.GetIPProfileDrv(ctx, tenant, id)
	if profile != nil {
		t.Errorf("expected profile to be nil, got %+v", profile)
	}
	if err != utils.ErrNotImplemented {
		t.Errorf("expected ErrNotImplemented, got %v", err)
	}

	expectedProfile := &utils.IPProfile{
		Tenant:    tenant,
		ID:        id,
		FilterIDs: []string{"f1", "f2"},
		Weights:   nil,
		TTL:       5 * time.Minute,
		Stored:    true,
		Pools:     nil,
	}

	dbMock := &DataDBMock{
		GetIPProfileDrvF: func(ctx *context.Context, tnt, iid string) (*utils.IPProfile, error) {
			if tnt != tenant || iid != id {
				t.Errorf("unexpected arguments: got %s, %s", tnt, iid)
			}
			return expectedProfile, nil
		},
	}

	profile2, err2 := dbMock.GetIPProfileDrv(ctx, tenant, id)
	if err2 != nil {
		t.Errorf("expected no error, got %v", err2)
	}
	if profile2 != expectedProfile {
		t.Errorf("expected profile %+v, got %+v", expectedProfile, profile2)
	}
}

func TestDataDBSetIPProfileDrv(t *testing.T) {
	ctx := &context.Context{}
	ipp := &utils.IPProfile{
		Tenant:    "cgrates.org",
		ID:        "ip1",
		FilterIDs: []string{"f1", "f2"},
		TTL:       5 * time.Minute,
		Stored:    true,
	}

	dbNil := &DataDBMock{}
	err := dbNil.SetIPProfileDrv(ctx, ipp)
	if err != utils.ErrNotImplemented {
		t.Errorf("expected ErrNotImplemented, got %v", err)
	}

	called := false
	dbMock := &DataDBMock{
		SetIPProfileDrvF: func(ctx *context.Context, ip *utils.IPProfile) error {
			called = true
			if ip != ipp {
				t.Errorf("expected IPProfile %+v, got %+v", ipp, ip)
			}
			return nil
		},
	}

	err2 := dbMock.SetIPProfileDrv(ctx, ipp)
	if err2 != nil {
		t.Errorf("expected no error, got %v", err2)
	}
	if !called {
		t.Errorf("expected SetIPProfileDrvF to be called")
	}

	expectedErr := utils.ErrNotFound
	dbMock2 := &DataDBMock{
		SetIPProfileDrvF: func(ctx *context.Context, ip *utils.IPProfile) error {
			return expectedErr
		},
	}

	err3 := dbMock2.SetIPProfileDrv(ctx, ipp)
	if err3 != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err3)
	}
}

func TestDataDBRemoveIPProfileDrv(t *testing.T) {
	ctx := &context.Context{}
	tenant := "cgrates.org"
	id := "ip1"

	dbNil := &DataDBMock{}
	err := dbNil.RemoveIPProfileDrv(ctx, tenant, id)
	if err != utils.ErrNotImplemented {
		t.Errorf("expected ErrNotImplemented, got %v", err)
	}

	called := false
	dbMock := &DataDBMock{
		RemoveIPProfileDrvF: func(ctx *context.Context, tntArg, idArg string) error {
			called = true
			if tntArg != tenant || idArg != id {
				t.Errorf("unexpected arguments: got %s, %s", tntArg, idArg)
			}
			return nil
		},
	}

	err2 := dbMock.RemoveIPProfileDrv(ctx, tenant, id)
	if err2 != nil {
		t.Errorf("expected no error, got %v", err2)
	}
	if !called {
		t.Errorf("expected RemoveIPProfileDrvF to be called")
	}

	expectedErr := utils.ErrNotFound
	dbMock2 := &DataDBMock{
		RemoveIPProfileDrvF: func(ctx *context.Context, tntArg, idArg string) error {
			return expectedErr
		},
	}

	err3 := dbMock2.RemoveIPProfileDrv(ctx, tenant, id)
	if err3 != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err3)
	}
}
