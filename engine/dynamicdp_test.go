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
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/ericlagergren/decimal"
	"github.com/nyaruka/phonenumbers"
	"golang.org/x/exp/slices"
)

func TestDynamicDPnewDynamicDP(t *testing.T) {

	expDDP := &DynamicDP{
		cfg:    config.CgrConfig(),
		tenant: "cgrates.org",
		initialDP: utils.StringSet{
			"test": struct{}{},
		},
		cache: utils.MapStorage{},
		ctx:   context.Background(),
	}

	if rcv := NewDynamicDP(context.Background(), config.CgrConfig(), "cgrates.org",
		utils.StringSet{"test": struct{}{}}, nil); !reflect.DeepEqual(rcv, expDDP) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expDDP), utils.ToJSON(rcv))
	}
}

func TestDynamicDPString(t *testing.T) {

	rcv := &DynamicDP{
		tenant: "cgrates.org",
		initialDP: utils.StringSet{
			"test": struct{}{},
		},
		cache: utils.MapStorage{},
		ctx:   context.Background(),
	}
	exp2 := "[\"test\"]"
	rcv2 := rcv.String()
	if !reflect.DeepEqual(rcv2, exp2) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(exp2), utils.ToJSON(rcv2))
	}
}

func TestDynamicDPFieldAsInterfaceErrFilename(t *testing.T) {

	rcv := &DynamicDP{
		tenant: "cgrates.org",
		initialDP: utils.StringSet{
			"test": struct{}{},
		},
		cache: utils.MapStorage{},
		ctx:   context.Background(),
	}
	_, err := rcv.FieldAsInterface([]string{""})
	if err == nil || err.Error() != "invalid fieldname <[]>" {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			"invalid fieldname <[]>", err)
	}
}

func TestDynamicDPFieldAsInterfaceErrLenFldPath(t *testing.T) {

	rcv := &DynamicDP{
		tenant: "cgrates.org",
		initialDP: utils.StringSet{
			"test": struct{}{},
		},
		cache: utils.MapStorage{},
		ctx:   context.Background(),
	}
	_, err := rcv.FieldAsInterface([]string{})
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ErrNotFound, err)
	}
}

func TestDynamicDPFieldAsInterface(t *testing.T) {

	DDP := &DynamicDP{
		tenant: "cgrates.org",
		initialDP: utils.StringSet{
			"test": struct{}{},
		},
		cache: utils.MapStorage{
			"testField": "testValue",
		},
		ctx: context.Background(),
	}
	result, err := DDP.FieldAsInterface([]string{"testField"})
	if err != nil {
		t.Error(err)
	}
	exp := "testValue"
	if !reflect.DeepEqual(result, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(exp), utils.ToJSON(result))
	}
}

func TestLibphonenumberDPString(t *testing.T) {
	var pInt int32 = 2
	LDP := &libphonenumberDP{
		pNumber: &phonenumbers.PhoneNumber{
			CountryCode: &pInt,
		},
	}
	exp2 := "country_code:2"
	rcv2 := LDP.String()
	if !reflect.DeepEqual(rcv2, exp2) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(exp2), utils.ToJSON(rcv2))
	}
}

func TestLibphonenumberDPFieldAsString(t *testing.T) {
	var pInt int32 = 2
	LDP := &libphonenumberDP{
		pNumber: &phonenumbers.PhoneNumber{
			CountryCode: &pInt,
		},
		cache: utils.MapStorage{
			"testField": "testValue",
		},
	}
	exp2 := "testValue"
	rcv2, err := LDP.FieldAsString([]string{"testField"})
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(rcv2, exp2) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(exp2), utils.ToJSON(rcv2))
	}
}

func TestLibphonenumberDPFieldAsStringError(t *testing.T) {
	var pInt int32 = 2
	LDP := &libphonenumberDP{
		pNumber: &phonenumbers.PhoneNumber{
			CountryCode: &pInt,
		},
		cache: utils.MapStorage{
			"testField": "testValue",
		},
	}
	_, err := LDP.FieldAsString([]string{"testField", "testField2"})
	if err == nil || err.Error() != "WRONG_PATH" {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			"WRONG_PATH", err)
	}
}

func TestLibphonenumberDPFieldAsInterfaceLen0(t *testing.T) {
	var pInt int32 = 2
	LDP := &libphonenumberDP{
		pNumber: &phonenumbers.PhoneNumber{
			CountryCode: &pInt,
		},
		cache: utils.MapStorage{
			"testField": "testValue",
		},
	}
	exp2 := &libphonenumberDP{
		pNumber: &phonenumbers.PhoneNumber{
			CountryCode: &pInt,
		},
		cache: utils.MapStorage{
			"testField": "testValue",
		}}
	exp2.setDefaultFields()

	rcv2, err := LDP.FieldAsInterface([]string{})
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(rcv2, exp2.cache) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			exp2.cache, rcv2)
	}
}

func TestNewLibPhoneNumberDPErr(t *testing.T) {

	number := "badNum"

	if _, err := newLibPhoneNumberDP(number); err != phonenumbers.ErrNotANumber {

		t.Error(err)
	}
}

func TestNewLibPhoneNumberDP(t *testing.T) {

	number := "+355123456789"

	num, err := phonenumbers.ParseAndKeepRawInput(number, utils.EmptyString)
	if err != nil {
		t.Error(err)
	}
	exp := utils.DataProvider(&libphonenumberDP{pNumber: num, cache: make(utils.MapStorage)})

	if err != nil {
		t.Error(err)
	}
	if rcv, err := newLibPhoneNumberDP(number); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected \n<%+v>, Received \n<%+v>", exp, rcv)
	}

}

func TestFieldAsInterfacelibphonenumberDP(t *testing.T) {

	dDP := &libphonenumberDP{
		pNumber: &phonenumbers.PhoneNumber{},
		cache: utils.MapStorage{
			"testField": "testValue",
		},
	}

	expErr := `invalid field path <[fld1 fld2 fld3]> for libphonenumberDP`
	if _, err := dDP.FieldAsInterface([]string{"fld1", "fld2", "fld3"}); err.Error() != expErr {
		t.Error(err)
	}
}

func TestLibphonenumberDPfieldAsInterfaceGeoLocationErr(t *testing.T) {
	tmp := utils.Logger
	defer func() {
		utils.Logger = tmp
	}()

	buf := new(bytes.Buffer)
	utils.Logger = utils.NewStdLoggerWithWriter(buf, "", 7)

	dDP := &libphonenumberDP{
		pNumber: &phonenumbers.PhoneNumber{},
		cache: utils.MapStorage{
			"testField": "testValue",
		},
	}

	if rcv, err := dDP.fieldAsInterface([]string{"GeoLocation"}); err != nil {
		t.Error(rcv, err)
	}
	expErr := "Received error: <language: tag is not well-formed> when getting GeoLocation for number"

	if rcvTxt := buf.String(); !strings.Contains(rcvTxt, expErr) {
		t.Errorf("Expected <%v>, Received <%v>", expErr, rcvTxt)
	}

	buf.Reset()
}

func TestLibphonenumberDPfieldAsInterface(t *testing.T) {

	var pInt int32 = 49
	var nNum uint64 = 17222020
	var leadingZeros int32 = 1

	lDP, err := newLibPhoneNumberDP("+49 17222020")
	if err != nil {
		t.Error(err)
	}
	libphone, canCast := lDP.(*libphonenumberDP)
	if !canCast {
		t.Errorf("cant cast <%v> to a libphonenumberDP", lDP)
	}
	dDP := &libphonenumberDP{
		pNumber: libphone.pNumber,
		cache: utils.MapStorage{
			"testField": "testValue",
		},
	}
	dDP.pNumber.Extension = utils.StringPointer("+")
	dDP.pNumber.PreferredDomesticCarrierCode = utils.StringPointer("49 172")

	exp := any(pInt)
	if rcv, err := dDP.fieldAsInterface([]string{"CountryCode"}); err != nil {
		t.Error(err)
	} else if rcv != exp {
		t.Errorf("Expected <%+v>, Received <%v>", exp, rcv)
	}

	exp = any(nNum)
	if rcv, err := dDP.fieldAsInterface([]string{"NationalNumber"}); err != nil {
		t.Error(err)
	} else if rcv != exp {
		t.Errorf("Expected <%+v>, Received <%v>", exp, rcv)
	}

	exp = any("DE")
	if rcv, err := dDP.fieldAsInterface([]string{"Region"}); err != nil {
		t.Error(err)
	} else if rcv != exp {
		t.Errorf("Expected <%+v>, Received <%v>", exp, rcv)
	}
	exp = any(phonenumbers.PhoneNumberType(11))
	if rcv, err := dDP.fieldAsInterface([]string{"NumberType"}); err != nil {
		t.Error(err)
	} else if rcv != exp {
		t.Errorf("Expected <%+v>, Received <%v><%t>", exp, rcv, rcv)
	}

	exp = any("Deutschland")
	if rcv, err := dDP.fieldAsInterface([]string{"GeoLocation"}); err != nil {
		t.Error(err)
	} else if rcv != exp {
		t.Errorf("Expected <%+v>, Received <%v>", exp, rcv)
	}
	exp = any("Vodafone")
	if rcv, err := dDP.fieldAsInterface([]string{"Carrier"}); err != nil {
		t.Error(err)
	} else if rcv != exp {
		t.Errorf("Expected <%+v>, Received <%v><%T>", exp, rcv, rcv)
	}
	exp = any(0)
	if rcv, err := dDP.fieldAsInterface([]string{"LengthOfNationalDestinationCode"}); err != nil {
		t.Error(err)
	} else if rcv != exp {
		t.Errorf("Expected <%+v>, Received <%v>", exp, rcv)
	}
	exp = any("+49 17222020")
	if rcv, err := dDP.fieldAsInterface([]string{"RawInput"}); err != nil {
		t.Error(err)
	} else if rcv != exp {
		t.Errorf("Expected <%+v>, Received <%v>", exp, rcv)
	}
	exp = any("+")
	if rcv, err := dDP.fieldAsInterface([]string{"Extension"}); err != nil {
		t.Error(err)
	} else if rcv != exp {
		t.Errorf("Expected <%+v>, Received <%v>", exp, rcv)
	}
	exp = any(leadingZeros)
	if rcv, err := dDP.fieldAsInterface([]string{"NumberOfLeadingZeros"}); err != nil {
		t.Error(err)
	} else if rcv != exp {
		t.Errorf("Expected <%+v>, Received <%v>", exp, rcv)
	}
	exp = any(false)
	if rcv, err := dDP.fieldAsInterface([]string{"ItalianLeadingZero"}); err != nil {
		t.Error(err)
	} else if rcv != exp {
		t.Errorf("Expected <%+v>, Received <%v>", exp, rcv)
	}
	exp = any("49 172")
	if rcv, err := dDP.fieldAsInterface([]string{"PreferredDomesticCarrierCode"}); err != nil {
		t.Error(err)
	} else if rcv != exp {
		t.Errorf("Expected <%+v>, Received <%v>", exp, rcv)
	}
	exp = any(phonenumbers.PhoneNumber_FROM_NUMBER_WITH_PLUS_SIGN)
	if rcv, err := dDP.fieldAsInterface([]string{"CountryCodeSource"}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected <%+v>, Received <%v>", exp, rcv)
	}

}

func TestDynamicDPfieldAsInterfaceMetaLibPhoneNumber(t *testing.T) {

	dDP := &DynamicDP{
		tenant: "cgrates.org",

		cache: utils.MapStorage{
			"testField": "testValue",
		},
		ctx: context.Background(),
	}

	dp, err := newLibPhoneNumberDP("+4917642092123")
	if err != nil {
		t.Error(err)
	}
	expAsField, err := dp.FieldAsInterface([]string{})
	if err != nil {
		t.Error(err)
	}
	exp := any(expAsField)
	if rcv, err := dDP.fieldAsInterface([]string{utils.MetaLibPhoneNumber, "+4917642092123"}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected \n<%v>, Received \n<%v>", exp, rcv)
	}
}

func TestDynamicDPfieldAsInterfaceErrMetaLibPhoneNumber(t *testing.T) {

	dDP := &DynamicDP{
		tenant: "cgrates.org",

		cache: utils.MapStorage{
			"testField": "testValue",
		},
		ctx: context.Background(),
	}

	if _, err := dDP.fieldAsInterface([]string{utils.MetaLibPhoneNumber, "inexistentNum"}); err == nil || err != phonenumbers.ErrNotANumber {
		t.Error(err)
	}
}

func TestDynamicDPfieldAsInterfaceNotFound(t *testing.T) {
	Cache.Clear(nil)

	ms := utils.MapStorage{}
	dDp := NewDynamicDP(context.Background(), config.CgrConfig(), "cgrates.org", ms, nil)

	if _, err := dDp.fieldAsInterface([]string{"inexistentfld1", "inexistentfld2"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
}

func TestDynamicDPfieldAsInterfaceErrMetaStats(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.FilterSCfg().Conns[utils.MetaResources] = []*config.DynamicStringSliceOpt{
		{Values: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)}},
	}
	dDP := &DynamicDP{
		cfg:    cfg,
		tenant: "cgrates.org",
		cache: utils.MapStorage{
			"testField": "testValue",
		},
		ctx: context.Background(),
	}

	cc := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ResourceSv1GetResourceWithConfig: func(ctx *context.Context, args, reply any) error {
				return utils.ErrNotImplemented
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- cc
	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources), utils.ResourceSv1, rpcInternal)
	dDP.fs = NewFilterS(cfg, cM, nil)

	if _, err := dDP.fieldAsInterface([]string{utils.MetaResources, "fld1"}); err == nil || err != utils.ErrNotImplemented {
		t.Error(err)
	}
}

func TestDynamicDPfieldAsInterfaceErrMetaAccounts(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.FilterSCfg().Conns[utils.MetaAccounts] = []*config.DynamicStringSliceOpt{
		{Values: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts)}},
	}
	dDP := &DynamicDP{
		cfg:    cfg,
		tenant: "cgrates.org",
		cache: utils.MapStorage{
			"testField": "testValue",
		},
		ctx: context.Background(),
	}

	customRply := &utils.Account{
		Tenant:    "cgrates.org",
		ID:        "1004",
		FilterIDs: []string{"*string:~*req.Account:1004"},
		Balances: map[string]*utils.Balance{
			"ConcreteBalance1": {
				ID: "ConcreteBalance1",
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
				Type:  utils.MetaConcrete,
				Units: &utils.Decimal{Big: decimal.New(1000, 0)},
				CostIncrements: []*utils.CostIncrement{
					{
						FilterIDs:    []string{"*string:~*req.ToR:*data"},
						Increment:    &utils.Decimal{Big: decimal.New(1, 0)},
						FixedFee:     &utils.Decimal{Big: decimal.New(0, 0)},
						RecurrentFee: &utils.Decimal{Big: decimal.New(1, 0)},
					},
				},
			},
		},
	}

	cc := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.AccountSv1GetAccount: func(ctx *context.Context, args, reply any) error {
				rplCast, canCast := reply.(*utils.Account)
				if !canCast {
					t.Errorf("Wrong argument type : %T", reply)
					return nil
				}
				*rplCast = *customRply
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- cc
	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts), utils.AccountSv1, rpcInternal)
	dDP.fs = NewFilterS(cfg, cM, nil)

	exp := &utils.Decimal{Big: decimal.New(1000, 0)}

	if rcv, err := dDP.fieldAsInterface([]string{utils.MetaAccounts, "1001", "Balances", "ConcreteBalance1", "Units"}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected: <%+v>,\nreceived: \n<%+v>", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestDynamicDPfieldAsInterfaceMetaResources(t *testing.T) {
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.FilterSCfg().Conns[utils.MetaResources] = []*config.DynamicStringSliceOpt{
		{Values: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)}},
	}
	dDP := &DynamicDP{
		cfg:    cfg,
		tenant: "cgrates.org",
		cache: utils.MapStorage{
			"testField": "testValue",
		},
		ctx: context.Background(),
	}

	customRply := &utils.ResourceWithConfig{
		Resource: &utils.Resource{
			Tenant: "cgrates.org",
			ID:     "ResGroup2",
			Usages: map[string]*utils.ResourceUsage{
				"RU1": {
					Tenant: "cgrates.org",
					ID:     "RU1",
					Units:  9,
				},
			},
		},
		Config: &utils.ResourceProfile{
			Tenant:            "cgrates.org",
			ID:                "ResGroup2",
			FilterIDs:         []string{"*string:~*req.Account:1001"},
			Limit:             10,
			AllocationMessage: "Approved",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				}},
			ThresholdIDs: []string{utils.MetaNone},
		},
	}

	cc := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ResourceSv1GetResourceWithConfig: func(ctx *context.Context, args, reply any) error {
				rplCast, canCast := reply.(*utils.ResourceWithConfig)
				if !canCast {
					t.Errorf("Wrong argument type : %T", reply)
					return nil
				}
				*rplCast = *customRply
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- cc
	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources), utils.ResourceSv1, rpcInternal)
	dDP.fs = NewFilterS(cfg, cM, nil)

	var exp float64 = 9

	if rcv, err := dDP.fieldAsInterface([]string{utils.MetaResources, "ResGroup2", "TotalUsage"}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected: <%+v>,\nreceived: \n<%+v>", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestDynamicDPfieldAsInterfaceMetaStats(t *testing.T) {
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.FilterSCfg().Conns[utils.MetaStats] = []*config.DynamicStringSliceOpt{
		{Values: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)}},
	}
	dDP := &DynamicDP{
		cfg:    cfg,
		tenant: "cgrates.org",
		cache: utils.MapStorage{
			"testField": "testValue",
		},
		ctx: context.Background(),
	}

	customRply := &map[string]*utils.Decimal{
		utils.MetaACD: utils.NewDecimal(1, 0),
		utils.MetaTCC: utils.NewDecimal(2, 0),
	}

	cc := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.StatSv1GetQueueDecimalMetrics: func(ctx *context.Context, args, reply any) error {
				rplCast, canCast := reply.(*map[string]*utils.Decimal)
				if !canCast {
					t.Errorf("Wrong argument type : %T", reply)
					return nil
				}
				*rplCast = *customRply
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- cc
	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats), utils.StatSv1, rpcInternal)
	dDP.fs = NewFilterS(cfg, cM, nil)

	exp := utils.NewDecimal(1, 0)

	if rcv, err := dDP.fieldAsInterface([]string{utils.MetaStats, "Stats2", utils.MetaACD}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(utils.ToJSON(exp), utils.ToJSON(rcv)) {
		t.Errorf("expected: <%+v><%T>,\nreceived: \n<%+v><%T>", utils.ToJSON(exp), exp, utils.ToJSON(rcv), rcv)
	}
}

func TestDPFilterSConns(t *testing.T) {
	acc1 := &utils.Account{
		Tenant: "cgrates.org",
		ID:     "1001",
		Balances: map[string]*utils.Balance{
			"Bal1": {
				ID:    "Bal1",
				Type:  utils.MetaConcrete,
				Units: &utils.Decimal{Big: decimal.New(1000, 0)},
			},
		},
	}

	registerAccountsMock := func(cM *ConnManager, connName string) {
		cc := &ccMock{
			calls: map[string]func(ctx *context.Context, args any, reply any) error{
				utils.AccountSv1GetAccount: func(ctx *context.Context, args, reply any) error {
					rpl, ok := reply.(*utils.Account)
					if !ok {
						return fmt.Errorf("wrong reply type %T", reply)
					}
					*rpl = *acc1
					return nil
				},
			},
		}
		ch := make(chan birpc.ClientConnector, 1)
		ch <- cc
		cM.AddInternalConn(connName, utils.AccountSv1, ch)
	}

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1001",
			utils.Destination:  "2003",
		},
	}

	t.Run("CommonConnWithReqFilter", func(t *testing.T) {
		cfg, err := config.NewCGRConfigFromJSONStringWithDefaults(`{
			"sessions": {
				"conns": {
					"*attributes": [
						{"FilterIDs": ["*string:~*req.Account:1001"], "Values": ["conn"]},
					]
				}
			}
		}`)
		if err != nil {
			t.Fatal(err)
		}
		dataDB, _ := NewInternalDB(nil, nil, nil, nil)
		dbCM := NewDBConnManager(map[string]DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
		dm := NewDataManager(dbCM, cfg, nil)
		fS := NewFilterS(cfg, nil, dm)
		NewConnManager(cfg)

		got, err := GetConnIDs(context.TODO(), cfg.SessionSCfg().Conns[utils.MetaAttributes], "cgrates.org", ev.AsDataProvider(), fS)
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(got, []string{"conn"}) {
			t.Errorf("expected [conn], got %v", got)
		}
	})

	t.Run("CommonConnOptWithFilterConn", func(t *testing.T) {
		cfg, err := config.NewCGRConfigFromJSONStringWithDefaults(`{
			"filters": {
				"conns": {
					"*accounts": [{"Values": ["*internal"]}]
				}
			},
			"chargers": {
				"conns": {
					"*attributes": [{"Values": ["conn2"],"FilterIDs":["*gte:~*accounts.1001.Balances[Bal1].Units:20"]}]
				}
			}
		}`)
		if err != nil {
			t.Fatal(err)
		}

		dataDB, _ := NewInternalDB(nil, nil, nil, nil)
		dbCM := NewDBConnManager(map[string]DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
		dm := NewDataManager(dbCM, cfg, nil)
		cM := NewConnManager(cfg)
		registerAccountsMock(cM, utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts))
		fS := NewFilterS(cfg, cM, dm)

		got, err := GetConnIDs(context.TODO(), cfg.ChargerSCfg().Conns[utils.MetaAttributes], "cgrates.org", ev.AsDataProvider(), fS)
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(got, []string{"conn2"}) {
			t.Errorf("expected [conn2], got %v", got)
		}
	})

	t.Run("ConnOptWithFilter", func(t *testing.T) {
		cfg, err := config.NewCGRConfigFromJSONStringWithDefaults(`{
			"filters": {
				"conns": {
					"*accounts": [
						{"FilterIDs": ["*string:~*req.Account:1001"], "Values": ["*internal"]},
					]
				}
			},
		}`)
		if err != nil {
			t.Fatal(err)
		}

		dataDB, _ := NewInternalDB(nil, nil, nil, nil)
		dbCM := NewDBConnManager(map[string]DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
		dm := NewDataManager(dbCM, cfg, nil)
		fS := NewFilterS(cfg, nil, dm)
		cM := NewConnManager(cfg)
		registerAccountsMock(cM, utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts))

		got, err := GetConnIDs(context.TODO(), cfg.FilterSCfg().Conns[utils.MetaAccounts], "cgrates.org", ev.AsDataProvider(), fS)
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(got, []string{"*internal:*accounts"}) {
			t.Errorf("expected [*internal], got %v", got)
		}
	})
}
