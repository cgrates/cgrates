/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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

package accounts

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/ericlagergren/decimal"
)

func TestActSetAccountBalance(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)

	acntID := "TestActSetAccount"
	diktats := []*utils.BalDiktat{
		{
			Path:  "*balance.Concrete1",
			Value: ";10",
		},
	}

	expected := "NO_DATA_BASE_CONNECTION"
	if err := actSetAccount(nil, "cgrates.org", acntID, diktats, false); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)

	expected = "WRONG_PATH"
	if err := actSetAccount(dm, "cgrates.org", acntID, diktats, false); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	diktats[0].Path = "*balance.Concrete1.NOT_A_FIELD"

	if err := actSetAccount(dm, "cgrates.org", acntID, diktats, false); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	diktats[0].Path = "*balance.Concrete1.Weights"

	expectedAcc := &utils.AccountProfile{
		Tenant: "cgrates.org",
		ID:     acntID,
		Balances: map[string]*utils.Balance{
			"Concrete1": {
				ID:    "Concrete1",
				Type:  utils.MetaConcrete,
				Units: &utils.Decimal{decimal.New(0, 0)},
				Weights: []*utils.DynamicWeight{
					{
						Weight: 10,
					},
				},
				CostIncrements: []*utils.CostIncrement{
					{
						FilterIDs:    []string{"*string:~*req.ToR:*voice"},
						Increment:    utils.NewDecimal(int64(time.Second), 0),
						RecurrentFee: utils.NewDecimal(0, 0),
					},
					{
						FilterIDs:    []string{"*string:~*req.ToR:*data"},
						Increment:    utils.NewDecimal(1024*1024, 0),
						RecurrentFee: utils.NewDecimal(0, 0),
					},
					{
						FilterIDs:    []string{"*string:~*req.ToR:*sms"},
						Increment:    utils.NewDecimal(1, 0),
						RecurrentFee: utils.NewDecimal(0, 0),
					},
				},
			},
		},
	}
	if err := actSetAccount(dm, "cgrates.org", acntID, diktats, false); err != nil {
		t.Error(err)
	} else if rcv, err := dm.GetAccountProfile("cgrates.org", acntID); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, expectedAcc) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expectedAcc), utils.ToJSON(rcv))
	}
}

func TestActSetAccount(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)

	acntID := "TestActSetAccount"
	diktats := []*utils.BalDiktat{
		{
			Path:  "*accountFilterIDs",
			Value: "10",
		},
	}

	expected := "WRONG_PATH"
	if err := actSetAccount(dm, "cgrates.org", acntID, diktats, false); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	diktats[0].Path = "*account"

	if err := actSetAccount(dm, "cgrates.org", acntID, diktats, false); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	diktats[0].Path = "*account.Weights"

	expected = "invalid DynamicWeight format for string <10>"
	if err := actSetAccount(dm, "cgrates.org", acntID, diktats, false); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	diktats[0].Value = ";10"

	expectedAcc := &utils.AccountProfile{
		Tenant: "cgrates.org",
		ID:     acntID,
		Weights: []*utils.DynamicWeight{
			{
				Weight: 10,
			},
		},
	}
	if err := actSetAccount(dm, "cgrates.org", acntID, diktats, false); err != nil {
		t.Error(err)
	} else if rcv, err := dm.GetAccountProfile("cgrates.org", acntID); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, expectedAcc) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expectedAcc), utils.ToJSON(rcv))
	}
}
