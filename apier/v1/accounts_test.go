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

package v1

import (
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	apierAcnts            *APIerSv1
	apierAcntsAcntStorage *engine.InternalDB
)

func init() {
	cfg := config.NewDefaultCGRConfig()
	apierAcntsAcntStorage = engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	apierAcnts = &APIerSv1{
		DataManager: engine.NewDataManager(apierAcntsAcntStorage, config.CgrConfig().CacheCfg(), nil),
		Config:      cfg,
	}
}

func TestSetAccounts(t *testing.T) {
	cgrTenant := "cgrates.org"
	iscTenant := "itsyscom.com"
	b10 := &engine.Balance{Value: 10, Weight: 10}
	cgrAcnt1 := &engine.Account{ID: utils.ConcatenatedKey(cgrTenant, "account1"),
		BalanceMap: map[string]engine.Balances{utils.MetaMonetary + utils.MetaOut: {b10}}}
	cgrAcnt2 := &engine.Account{ID: utils.ConcatenatedKey(cgrTenant, "account2"),
		BalanceMap: map[string]engine.Balances{utils.MetaMonetary + utils.MetaOut: {b10}}}
	cgrAcnt3 := &engine.Account{ID: utils.ConcatenatedKey(cgrTenant, "account3"),
		BalanceMap: map[string]engine.Balances{utils.MetaMonetary + utils.MetaOut: {b10}}}
	iscAcnt1 := &engine.Account{ID: utils.ConcatenatedKey(iscTenant, "account1"),
		BalanceMap: map[string]engine.Balances{utils.MetaMonetary + utils.MetaOut: {b10}}}
	iscAcnt2 := &engine.Account{ID: utils.ConcatenatedKey(iscTenant, "account2"),
		BalanceMap: map[string]engine.Balances{utils.MetaMonetary + utils.MetaOut: {b10}}}
	for _, account := range []*engine.Account{cgrAcnt1, cgrAcnt2, cgrAcnt3, iscAcnt1, iscAcnt2} {
		if err := apierAcntsAcntStorage.SetAccountDrv(account); err != nil {
			t.Error(err)
		}
	}
	//apierAcntsAcntStorage.CacheRatingPrefixes(utils.ActionPrefix)
}

/* This was a comment
func TestGetAccountIds(t *testing.T) {
	var accountIds []string
	var attrs AttrGetAccountIds
	if err := apierAcnts.GetAccountIds(attrs, &accountIds); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if len(accountIds) != 5 {
		t.Errorf("Accounts returned: %+v", accountIds)
	}
}
*/

func TestGetAccounts(t *testing.T) {
	var accounts []any
	var attrs utils.AttrGetAccounts
	if err := apierAcnts.GetAccounts(context.Background(), &utils.AttrGetAccounts{Tenant: "cgrates.org"}, &accounts); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if len(accounts) != 3 {
		t.Errorf("Accounts returned: %+v", accounts)
	}
	attrs = utils.AttrGetAccounts{Tenant: "itsyscom.com"}
	if err := apierAcnts.GetAccounts(context.Background(), &attrs, &accounts); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if len(accounts) != 2 {
		t.Errorf("Accounts returned: %+v", accounts)
	}
	attrs = utils.AttrGetAccounts{Tenant: "cgrates.org", AccountIDs: []string{"account1"}}
	if err := apierAcnts.GetAccounts(context.Background(), &attrs, &accounts); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if len(accounts) != 1 {
		t.Errorf("Accounts returned: %+v", accounts)
	}
	attrs = utils.AttrGetAccounts{Tenant: "itsyscom.com", AccountIDs: []string{"INVALID"}}
	if err := apierAcnts.GetAccounts(context.Background(), &attrs, &accounts); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if len(accounts) != 0 {
		t.Errorf("Accounts returned: %+v", accounts)
	}
	attrs = utils.AttrGetAccounts{Tenant: "INVALID"}
	if err := apierAcnts.GetAccounts(context.Background(), &attrs, &accounts); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if len(accounts) != 0 {
		t.Errorf("Accounts returned: %+v", accounts)
	}
}
