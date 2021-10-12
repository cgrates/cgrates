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

package config

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
	"github.com/ericlagergren/decimal"
)

func TestAccountSCfgLoadFromJSONCfg(t *testing.T) {
	jsonCfg := &AccountSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Attributes_conns:      &[]string{utils.MetaInternal},
		Rates_conns:           &[]string{utils.MetaInternal},
		Thresholds_conns:      &[]string{utils.MetaInternal},
		Indexed_selects:       utils.BoolPointer(false),
		String_indexed_fields: &[]string{"*req.index1"},
		Prefix_indexed_fields: &[]string{"*req.index1"},
		Suffix_indexed_fields: &[]string{"*req.index1"},
		Nested_fields:         utils.BoolPointer(true),
		Max_iterations:        utils.IntPointer(1000),
		Max_usage:             utils.StringPointer("200h"),
	}
	usage, err := utils.NewDecimalFromUsage("200h")
	if err != nil {
		t.Error(err)
	}
	expected := &AccountSCfg{
		Enabled:             true,
		AttributeSConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)},
		RateSConns:          []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRateS)},
		ThresholdSConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)},
		IndexedSelects:      false,
		StringIndexedFields: &[]string{"*req.index1"},
		PrefixIndexedFields: &[]string{"*req.index1"},
		SuffixIndexedFields: &[]string{"*req.index1"},
		NestedFields:        true,
		MaxIterations:       1000,
		MaxUsage:            usage,
		Opts: &AccountsOpts{
			AccountIDs: []*utils.DynamicStringSliceOpt{
				{
					FilterIDs: []string{utils.MetaDefault},
					Value:     []string{},
				},
			},
			Usage: []*utils.DynamicDecimalBigOpt{
				{
					FilterIDs: []string{utils.MetaDefault},
					Value:     decimal.New(int64(72*time.Hour), 0),
				},
			},
			ProfileIgnoreFilters: []*utils.DynamicBoolOpt{
				{
					FilterIDs: []string{utils.MetaDefault},
					Value:     false,
				},
			},
		},
	}
	jsnCfg := NewDefaultCGRConfig()
	if err = jsnCfg.accountSCfg.loadFromJSONCfg(jsonCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsnCfg.accountSCfg) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expected), utils.ToJSON(jsnCfg.accountSCfg))
	}

	//test for empty config
	jsonCfg = nil
	if err = jsnCfg.accountSCfg.loadFromJSONCfg(jsonCfg); err != nil {
		t.Error(err)
	}
}

func TestAccountSCfgLoadFromJSONCfgOptsErr(t *testing.T) {
	accOpts := &AccountsOpts{
		AccountIDs: []*utils.DynamicStringSliceOpt{
			{
				Value: []string{"1001", "1002"},
			},
		},
		Usage: []*utils.DynamicDecimalBigOpt{
			{
				Value: new(decimal.Big).SetUint64(2),
			},
		},
		ProfileIgnoreFilters: []*utils.DynamicBoolOpt{
			{
				Value: false,
			},
		},
	}

	jsnCfg := &AccountsOptsJson{
		AccountIDs: map[string][]string{},
		Usage: map[string]string{
			"Value": "error",
		},
		ProfileIgnoreFilters: map[string]bool{},
	}
	errExp := "can't convert <error> to decimal"
	if err := accOpts.loadFromJSONCfg(jsnCfg); err == nil || err.Error() != errExp {
		t.Errorf("Expected %v \n but received \n %v", errExp, err)
	}

	//also test for empty json config
	jsnCfg = nil
	if err := accOpts.loadFromJSONCfg((jsnCfg)); err != nil {
		t.Error(err)
	}
}

func TestAccountsCfLoadConfigError(t *testing.T) {
	accountsJson := &AccountSJsonCfg{
		Max_usage: utils.StringPointer("invalid_Decimal"),
	}
	actsCfg := new(AccountSCfg)
	expected := "strconv.ParseInt: parsing \"invalid_Decimal\": invalid syntax"
	if err := actsCfg.loadFromJSONCfg(accountsJson); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestAccountSCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
"accounts": {								
	"enabled": true,						
	"indexed_selects": false,			
	"attributes_conns": ["*internal:*attributes"],
	"rates_conns": ["*internal:*rates"],
	"thresholds_conns": ["*internal:*thresholds"],					
	"string_indexed_fields": ["*req.index1"],			
	"prefix_indexed_fields": ["*req.index1"],			
	"suffix_indexed_fields": ["*req.index1"],			
	"nested_fields": true,			
    "max_iterations": 100,
    "max_usage": "72h",
},	
}`

	eMap := map[string]interface{}{
		utils.EnabledCfg:             true,
		utils.IndexedSelectsCfg:      false,
		utils.AttributeSConnsCfg:     []string{utils.MetaInternal},
		utils.RateSConnsCfg:          []string{utils.MetaInternal},
		utils.ThresholdSConnsCfg:     []string{utils.MetaInternal},
		utils.StringIndexedFieldsCfg: []string{"*req.index1"},
		utils.PrefixIndexedFieldsCfg: []string{"*req.index1"},
		utils.SuffixIndexedFieldsCfg: []string{"*req.index1"},
		utils.NestedFieldsCfg:        true,
		utils.MaxIterations:          100,
		utils.MaxUsage:               "259200000000000", // 72h in ns
		utils.OptsCfg: map[string]interface{}{
			utils.MetaAccountIDsCfg: map[string][]string{
				utils.MetaDefault: {},
			},
			utils.MetaUsage: map[string]string{
				utils.MetaDefault: decimal.New(int64(72*time.Hour), 0).String(),
			},
			utils.MetaProfileIgnoreFilters: map[string]bool{
				utils.MetaDefault: false,
			},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.accountSCfg.AsMapInterface(""); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected: %+v\n Received: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestAccountSCfgClone(t *testing.T) {
	usage, err := utils.NewDecimalFromUsage("24h")
	if err != nil {
		t.Error(err)
	}
	ban := &AccountSCfg{
		Enabled:             true,
		IndexedSelects:      false,
		AttributeSConns:     []string{"*req.index1"},
		RateSConns:          []string{"*req.index1"},
		ThresholdSConns:     []string{"*req.index1"},
		StringIndexedFields: &[]string{"*req.index1"},
		PrefixIndexedFields: &[]string{"*req.index1", "*req.index2"},
		SuffixIndexedFields: &[]string{"*req.index1"},
		NestedFields:        true,
		MaxIterations:       1000,
		MaxUsage:            usage,
		Opts:                &AccountsOpts{},
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("\nExpected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
	if (rcv.AttributeSConns)[0] = utils.EmptyString; (ban.AttributeSConns)[0] != "*req.index1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if (rcv.RateSConns)[0] = utils.EmptyString; (ban.RateSConns)[0] != "*req.index1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if (rcv.ThresholdSConns)[0] = utils.EmptyString; (ban.ThresholdSConns)[0] != "*req.index1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if (*rcv.StringIndexedFields)[0] = utils.EmptyString; (*ban.StringIndexedFields)[0] != "*req.index1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if (*rcv.PrefixIndexedFields)[0] = utils.EmptyString; (*ban.PrefixIndexedFields)[0] != "*req.index1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if (*rcv.SuffixIndexedFields)[0] = utils.EmptyString; (*ban.SuffixIndexedFields)[0] != "*req.index1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
}

func TestDiffAccountSJsonCfg(t *testing.T) {
	var d *AccountSJsonCfg

	v1 := &AccountSCfg{
		Enabled:             true,
		AttributeSConns:     []string{"*localhost"},
		RateSConns:          []string{},
		ThresholdSConns:     []string{},
		IndexedSelects:      true,
		StringIndexedFields: &[]string{"~*req.Index1"},
		PrefixIndexedFields: &[]string{},
		SuffixIndexedFields: &[]string{},
		NestedFields:        true,
		MaxIterations:       1,
		MaxUsage:            nil,
		Opts: &AccountsOpts{
			AccountIDs: []*utils.DynamicStringSliceOpt{
				{
					Value: []string{"ACC1"},
				},
			},
		},
	}

	v2 := &AccountSCfg{
		Enabled:             false,
		AttributeSConns:     []string{"*localhost", "*birpc"},
		RateSConns:          []string{"*localhost"},
		ThresholdSConns:     []string{"*localhost"},
		IndexedSelects:      false,
		StringIndexedFields: &[]string{"~*req.Index1"},
		PrefixIndexedFields: &[]string{},
		SuffixIndexedFields: &[]string{},
		NestedFields:        false,
		MaxIterations:       3,
		MaxUsage:            utils.NewDecimal(60, 0),
		Opts: &AccountsOpts{
			AccountIDs: []*utils.DynamicStringSliceOpt{
				{
					Value: []string{"ACC2"},
				},
			},
		},
	}

	expected1 := &AccountSJsonCfg{
		Enabled:               utils.BoolPointer(false),
		Indexed_selects:       utils.BoolPointer(false),
		Attributes_conns:      &[]string{"*localhost", "*birpc"},
		Rates_conns:           &[]string{"*localhost"},
		Thresholds_conns:      &[]string{"*localhost"},
		String_indexed_fields: nil,
		Prefix_indexed_fields: nil,
		Suffix_indexed_fields: nil,
		Nested_fields:         utils.BoolPointer(false),
		Max_iterations:        utils.IntPointer(3),
		Max_usage:             utils.StringPointer("60"),
		Opts: &AccountsOptsJson{
			AccountIDs: map[string][]string{
				utils.EmptyString: {"ACC2"},
			},
		},
	}

	rcv := diffAccountSJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected1) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected1), utils.ToJSON(rcv))
	}

	//MaxUsage is nil in v2
	v2_2 := &AccountSCfg{
		Enabled:             false,
		AttributeSConns:     []string{"*localhost", "*birpc"},
		RateSConns:          []string{"*localhost"},
		ThresholdSConns:     []string{"*localhost"},
		IndexedSelects:      false,
		StringIndexedFields: &[]string{"~*req.Index1"},
		PrefixIndexedFields: &[]string{},
		SuffixIndexedFields: &[]string{},
		NestedFields:        false,
		MaxIterations:       3,
		MaxUsage:            nil,
		Opts: &AccountsOpts{
			AccountIDs: []*utils.DynamicStringSliceOpt{
				{
					Value: []string{"ACC2"},
				},
			},
		},
	}

	expected2 := &AccountSJsonCfg{
		Enabled:               utils.BoolPointer(false),
		Indexed_selects:       utils.BoolPointer(false),
		Attributes_conns:      &[]string{"*localhost", "*birpc"},
		Rates_conns:           &[]string{"*localhost"},
		Thresholds_conns:      &[]string{"*localhost"},
		String_indexed_fields: nil,
		Prefix_indexed_fields: nil,
		Suffix_indexed_fields: nil,
		Nested_fields:         utils.BoolPointer(false),
		Max_iterations:        utils.IntPointer(3),
		Max_usage:             nil,
		Opts: &AccountsOptsJson{
			AccountIDs: map[string][]string{
				utils.EmptyString: {"ACC2"},
			},
		},
	}

	rcv = diffAccountSJsonCfg(d, v1, v2_2)
	if !reflect.DeepEqual(rcv, expected2) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected2), utils.ToJSON(rcv))
	}

	//Make the two Accounts equal in order to get a nil "d"

	v2_3 := v1
	expected3 := &AccountSJsonCfg{
		Enabled:               nil,
		Indexed_selects:       nil,
		Attributes_conns:      nil,
		Rates_conns:           nil,
		Thresholds_conns:      nil,
		String_indexed_fields: nil,
		Prefix_indexed_fields: nil,
		Suffix_indexed_fields: nil,
		Nested_fields:         nil,
		Max_iterations:        nil,
		Max_usage:             nil,
		Opts:                  &AccountsOptsJson{},
	}

	rcv = diffAccountSJsonCfg(d, v1, v2_3)
	if !reflect.DeepEqual(rcv, expected3) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected3), utils.ToJSON(rcv))
	}
}

func TestAccountSCloneSection(t *testing.T) {
	acS := &AccountSCfg{
		Enabled:             true,
		AttributeSConns:     []string{"*localhost"},
		RateSConns:          []string{},
		ThresholdSConns:     []string{},
		IndexedSelects:      true,
		StringIndexedFields: &[]string{"~*req.Index1"},
		PrefixIndexedFields: &[]string{},
		SuffixIndexedFields: &[]string{},
		NestedFields:        true,
		MaxIterations:       1,
		MaxUsage:            nil,
		Opts: &AccountsOpts{
			AccountIDs: []*utils.DynamicStringSliceOpt{
				{
					Value: []string{"ACC1"},
				},
			},
		},
	}

	rcv := acS.CloneSection()
	if !reflect.DeepEqual(acS, rcv.(*AccountSCfg)) {
		t.Errorf("Expected %v \n but received \n %v", acS, rcv.(*AccountSCfg))
	}
}
