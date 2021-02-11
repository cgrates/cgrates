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

	"github.com/cgrates/cgrates/config"

	"github.com/cgrates/rpcclient"

	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/utils"
)

func TestNewAccountBalanceOperators(t *testing.T) {
	acntPrf := &utils.AccountProfile{
		ID:     "TEST_ID",
		Tenant: "cgrates.org",
		Balances: map[string]*utils.Balance{
			"BL0": {
				ID:   "BALANCE1",
				Type: utils.MetaAbstract,
			},
			"BL1": {
				ID:   "BALANCE1",
				Type: utils.MetaConcrete,
				CostIncrements: []*utils.CostIncrement{
					{
						Increment: utils.NewDecimal(int64(time.Second), 0),
					},
				},
			},
		},
	}
	config := config.NewDefaultCGRConfig()
	filters := engine.NewFilterS(config, nil, nil)

	concrete, err := newBalanceOperator(acntPrf.Balances["BL1"], nil, filters, nil, nil, nil)
	if err != nil {
		t.Error(err)
	}
	var cncrtBlncs []*concreteBalance
	cncrtBlncs = append(cncrtBlncs, concrete.(*concreteBalance))

	expected := &abstractBalance{
		blnCfg:     acntPrf.Balances["BL0"],
		fltrS:      filters,
		cncrtBlncs: cncrtBlncs,
	}

	if blcOp, err := newAccountBalanceOperators(acntPrf, filters, nil,
		nil, nil); err != nil {
		t.Error(err)
	} else {
		rcv := blcOp[0].(*abstractBalance)
		if !reflect.DeepEqual(expected, rcv) {
			t.Errorf("Expected %+v, received %+v", expected, rcv)
		}
	}

	acntPrf.Balances["BL1"].Type = "INVALID_TYPE"
	expectedErr := "unsupported balance type: <INVALID_TYPE>"
	if _, err := newAccountBalanceOperators(acntPrf, filters, nil,
		nil, nil); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

type testMockCall struct {
	calls map[string]func(args interface{}, reply interface{}) error
}

func (tS *testMockCall) Call(method string, args interface{}, rply interface{}) error {
	if call, has := tS.calls[method]; !has {
		return rpcclient.ErrUnsupporteServiceMethod
	} else {
		return call(args, rply)
	}
}

func TestProcessAttributeS(t *testing.T) {
	engine.Cache.Clear(nil)

	config := config.NewDefaultCGRConfig()
	sTestMock := &testMockCall{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.AttributeSv1ProcessEvent: func(args interface{}, reply interface{}) error {
				return utils.ErrNotImplemented
			},
		},
	}
	chanInternal := make(chan rpcclient.ClientConnector, 1)
	chanInternal <- sTestMock
	connMgr := engine.NewConnManager(config, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes): chanInternal,
	})
	cgrEvent := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TEST_ID1",
		Opts: map[string]interface{}{
			utils.OptsAttributesProcessRuns: "20",
		},
	}

	attrsConns := []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}

	if _, err := processAttributeS(connMgr, cgrEvent, attrsConns, nil); err == nil || err != utils.ErrNotImplemented {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotImplemented, err)
	}
}
