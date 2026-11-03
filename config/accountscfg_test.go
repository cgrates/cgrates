// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestAccountSCfgLoadFromJSONCfg(t *testing.T) {
	jsonCfg := &AccountSJsonCfg{
		Enabled: utils.BoolPointer(true),
		Conns: map[string][]*DynamicConns{
			utils.MetaAttributes: {{ConnIDs: []string{utils.MetaInternal}}},
			utils.MetaRates:      {{ConnIDs: []string{utils.MetaInternal}}},
			utils.MetaThresholds: {{ConnIDs: []string{utils.MetaInternal}}},
		},
		Indexed_selects:          utils.BoolPointer(false),
		String_indexed_fields:    &[]string{"*req.index1"},
		Prefix_indexed_fields:    &[]string{"*req.index1"},
		Suffix_indexed_fields:    &[]string{"*req.index1"},
		Exists_indexed_fields:    &[]string{"*req.index1"},
		Notexists_indexed_fields: &[]string{"*req.index1"},
		Nested_fields:            utils.BoolPointer(true),
		Max_iterations:           utils.IntPointer(1000),
		Max_usage:                utils.StringPointer("200h"),
		Opts: &AccountsOptsJson{
			ProfileIDs: []*DynamicStringSliceOpt{
				{
					Tenant: "cgrates.org",
					Values: []string{"ACC1"},
				},
			},
			Usage: []*DynamicInterfaceOpt{
				{
					FilterIDs: []string{"fld1"},
					Tenant:    "cgrates.org",
					Value:     AccountsUsageDftOpt,
				},
			},
			ProfileIgnoreFilters: []*DynamicInterfaceOpt{
				{
					Tenant: "cgrates.org",
					Value:  true,
				},
			},
			Thresholds: []*DynamicInterfaceOpt{
				{
					Tenant: "cgrates.org",
					Value:  true,
				},
			},
			ThresholdIDs: []*DynamicStringSliceOpt{
				{
					Tenant: "cgrates.org",
					Values: []string{"THR1"},
				},
			},
		},
	}
	usage, err := utils.NewDecimalFromUsage("200h")
	if err != nil {
		t.Error(err)
	}
	expected := &AccountSCfg{
		Enabled: true,
		Conns: map[string][]*DynamicConns{
			utils.MetaAttributes: {{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}}},
			utils.MetaRates:      {{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRates)}}},
			utils.MetaThresholds: {{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)}}},
		},
		IndexedSelects:         false,
		StringIndexedFields:    &[]string{"*req.index1"},
		PrefixIndexedFields:    &[]string{"*req.index1"},
		SuffixIndexedFields:    &[]string{"*req.index1"},
		ExistsIndexedFields:    &[]string{"*req.index1"},
		NotExistsIndexedFields: &[]string{"*req.index1"},
		NestedFields:           true,
		MaxIterations:          1000,
		MaxUsage:               usage,
		Opts: &AccountsOpts{
			ProfileIDs: []*DynamicStringSliceOpt{
				{
					Tenant: "cgrates.org",
					Values: []string{"ACC1"},
				},
			},
			Usage: []*DynamicDecimalOpt{
				NewDynamicDecimalOpt([]string{"fld1"}, "cgrates.org", AccountsUsageDftOpt, nil),
				NewDynamicDecimalOpt(nil, "", AccountsUsageDftOpt, nil),
			},
			ProfileIgnoreFilters: []*DynamicBoolOpt{
				{
					FilterIDs: nil,
					Tenant:    "cgrates.org",
					value:     true,
				},
				{},
			},
			Thresholds: []*DynamicBoolOpt{
				{
					Tenant: "cgrates.org",
					value:  true,
				},
			},
			ThresholdIDs: []*DynamicStringSliceOpt{
				{
					Tenant: "cgrates.org",
					Values: []string{"THR1"},
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
		ProfileIDs: []*DynamicStringSliceOpt{
			{
				Values: []string{"1001", "1002"},
			},
		},
		Usage: []*DynamicDecimalOpt{
			NewDynamicDecimalOpt(nil, "", utils.NewDecimal(2, 0), nil),
		},
		ProfileIgnoreFilters: []*DynamicBoolOpt{
			{
				value: false,
			},
		},
	}

	jsnCfg := &AccountsOptsJson{
		ProfileIDs: []*DynamicStringSliceOpt{
			{},
		},
		Usage: []*DynamicInterfaceOpt{
			{
				Value: "error",
			},
		},
		ProfileIgnoreFilters: []*DynamicInterfaceOpt{
			{
				Value: "false",
			},
		},
	}
	errExp := "can't convert <error> to decimal"
	if err := accOpts.loadFromJSONCfg(jsnCfg); err == nil || err.Error() != errExp {
		t.Errorf("Expected %v \n but received \n %v", errExp, err)
	}

	//also test for empty json config
	jsnCfg = nil
	if err := accOpts.loadFromJSONCfg(jsnCfg); err != nil {
		t.Error(err)
	}
}

func TestAccountsCfLoadConfigError(t *testing.T) {
	accountsJson := &AccountSJsonCfg{
		Max_usage: utils.StringPointer("invalid_Decimal"),
	}
	actsCfg := new(AccountSCfg)
	expected := "can't convert <invalid_Decimal> to decimal"
	if err := actsCfg.loadFromJSONCfg(accountsJson); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestAccountSCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
"accounts": {
	"enabled": true,
	"indexedSelects": false,
	"conns": {
		"*attributes": [{"connIDs": ["*internal:*attributes"]}],
		"*rates":      [{"connIDs": ["*internal:*rates"]}],
		"*thresholds": [{"connIDs": ["*internal:*thresholds"]}]
	},
	"stringIndexedFields": ["*req.index1"],
	"prefixIndexedFields": ["*req.index1"],
	"suffixIndexedFields": ["*req.index1"],
	"existsIndexedFields": ["*req.index1"],
	"notExistsIndexedFields": ["*req.index1"],
	"nestedFields": true,
    "maxIterations": 100,
    "maxUsage": "72h",
	"opts": {
		"*profileIDs": [
			{
				"tenant": "cgrates.org", 
				"values": ["ACC1"]
			}
		],
		"*profileIgnoreFilters": [{"value": false}],
		"*thresholdIDs": [
			{
				"tenant": "cgrates.org", 
				"values": ["THR1"]
			}
		],
		"*thresholds": [
			{
				"tenant": "cgrates.org", 
				"value": true
			}
		]
	}
},
}`

	eMap := map[string]any{
		utils.EnabledCfg:        true,
		utils.IndexedSelectsCfg: false,
		utils.ConnsCfg: map[string][]*DynamicConns{
			utils.MetaAttributes: {{ConnIDs: []string{utils.MetaInternal}}},
			utils.MetaRates:      {{ConnIDs: []string{utils.MetaInternal}}},
			utils.MetaThresholds: {{ConnIDs: []string{utils.MetaInternal}}},
		},
		utils.StringIndexedFieldsCfg:    []string{"*req.index1"},
		utils.PrefixIndexedFieldsCfg:    []string{"*req.index1"},
		utils.SuffixIndexedFieldsCfg:    []string{"*req.index1"},
		utils.ExistsIndexedFieldsCfg:    []string{"*req.index1"},
		utils.NotExistsIndexedFieldsCfg: []string{"*req.index1"},
		utils.NestedFieldsCfg:           true,
		utils.MaxIterations:             100,
		utils.MaxUsage:                  "259200000000000", // 72h in ns
		utils.OptsCfg: map[string]any{
			utils.MetaProfileIDs: []*DynamicStringSliceOpt{
				{
					Tenant: "cgrates.org",
					Values: []string{"ACC1"},
				},
			},
			utils.MetaUsage: []*DynamicDecimalOpt{NewDynamicDecimalOpt(nil, "", AccountsUsageDftOpt, nil)},
			utils.MetaProfileIgnoreFilters: []*DynamicBoolOpt{
				{
					value: false,
				},
				{},
			},
			utils.MetaThresholds: []*DynamicBoolOpt{
				{
					Tenant: "cgrates.org",
					value:  true,
				},
			},
			utils.MetaThresholdIDs: []*DynamicStringSliceOpt{
				{
					Tenant: "cgrates.org",
					Values: []string{"THR1"},
				},
			},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.accountSCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected: %+v\n Received: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestAccountSCfgClone(t *testing.T) {
	usage, err := utils.NewDecimalFromUsage("24h")
	if err != nil {
		t.Error(err)
	}
	ban := &AccountSCfg{
		Enabled:        true,
		IndexedSelects: false,
		Conns: map[string][]*DynamicConns{
			utils.MetaAttributes: {{ConnIDs: []string{"*req.index1"}}},
			utils.MetaRates:      {{ConnIDs: []string{"*req.index1"}}},
			utils.MetaThresholds: {{ConnIDs: []string{"*req.index1"}}},
		},
		StringIndexedFields: &[]string{"*req.index1"},
		PrefixIndexedFields: &[]string{"*req.index1", "*req.index2"},
		SuffixIndexedFields: &[]string{"*req.index1"},
		NestedFields:        true,
		MaxIterations:       1000,
		MaxUsage:            usage,
		Opts: &AccountsOpts{
			ProfileIDs: []*DynamicStringSliceOpt{
				{
					Tenant: "cgrates.org",
					Values: []string{"ACC1"},
				},
			},
			Usage: []*DynamicDecimalOpt{
				NewDynamicDecimalOpt(nil, "cgrates.org", utils.NewDecimal(1, 0), nil),
			},
			ProfileIgnoreFilters: []*DynamicBoolOpt{
				{
					Tenant: "cgrates.org",
					value:  true,
				},
				{},
			},
			Thresholds: []*DynamicBoolOpt{
				{
					Tenant: "cgrates.org",
					value:  true,
				},
			},
			ThresholdIDs: []*DynamicStringSliceOpt{
				{
					Tenant: "cgrates.org",
					Values: []string{"THR1"},
				},
			},
		},
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("\nExpected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
	if rcv.Conns[utils.MetaAttributes][0].ConnIDs[0] = utils.EmptyString; ban.Conns[utils.MetaAttributes][0].ConnIDs[0] != "*req.index1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.Conns[utils.MetaRates][0].ConnIDs[0] = utils.EmptyString; ban.Conns[utils.MetaRates][0].ConnIDs[0] != "*req.index1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.Conns[utils.MetaThresholds][0].ConnIDs[0] = utils.EmptyString; ban.Conns[utils.MetaThresholds][0].ConnIDs[0] != "*req.index1" {
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
	if !reflect.DeepEqual(*rcv.Opts.ProfileIDs[0], *ban.Opts.ProfileIDs[0]) {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if !reflect.DeepEqual(*rcv.Opts.Usage[0], *ban.Opts.Usage[0]) {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if !reflect.DeepEqual(*rcv.Opts.ProfileIgnoreFilters[0], *ban.Opts.ProfileIgnoreFilters[0]) {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if !reflect.DeepEqual(*rcv.Opts.Thresholds[0], *ban.Opts.Thresholds[0]) {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if !reflect.DeepEqual(*rcv.Opts.ThresholdIDs[0], *ban.Opts.ThresholdIDs[0]) {
		t.Errorf("Expected clone to not modify the cloned")
	}
}

func TestDiffAccountSJsonCfg(t *testing.T) {
	var d *AccountSJsonCfg

	v1 := &AccountSCfg{
		Enabled: true,
		Conns: map[string][]*DynamicConns{
			utils.MetaAttributes: {{ConnIDs: []string{"*localhost"}}},
			utils.MetaRates:      {{ConnIDs: []string{}}},
			utils.MetaThresholds: {{ConnIDs: []string{}}},
		},
		IndexedSelects:      true,
		StringIndexedFields: &[]string{"~*req.Index1"},
		PrefixIndexedFields: &[]string{},
		SuffixIndexedFields: &[]string{},
		NestedFields:        true,
		MaxIterations:       1,
		MaxUsage:            nil,
		Opts: &AccountsOpts{
			ProfileIDs: []*DynamicStringSliceOpt{
				{
					Tenant: "cgrates.org",
					Values: []string{"ACC1"},
				},
			},
			Usage: []*DynamicDecimalOpt{
				NewDynamicDecimalOpt(nil, "cgrates.org", utils.NewDecimal(1, 0), nil),
			},
			ProfileIgnoreFilters: []*DynamicBoolOpt{
				{
					Tenant: "cgrates.org",
					value:  true,
				},
			},
			Thresholds: []*DynamicBoolOpt{
				{
					Tenant: "cgrates",
					value:  false,
				},
			},
			ThresholdIDs: []*DynamicStringSliceOpt{
				{
					Tenant: "cgrates",
					Values: []string{"TH1"},
				},
			},
		},
	}

	v2 := &AccountSCfg{
		Enabled: false,
		Conns: map[string][]*DynamicConns{
			utils.MetaAttributes: {{ConnIDs: []string{"*localhost", "*birpc"}}},
			utils.MetaRates:      {{ConnIDs: []string{"*localhost"}}},
			utils.MetaThresholds: {{ConnIDs: []string{"*localhost"}}},
		},
		IndexedSelects:      false,
		StringIndexedFields: &[]string{"~*req.Index1"},
		PrefixIndexedFields: &[]string{},
		SuffixIndexedFields: &[]string{},
		NestedFields:        false,
		MaxIterations:       3,
		MaxUsage:            utils.NewDecimal(60, 0),
		Opts: &AccountsOpts{
			ProfileIDs: []*DynamicStringSliceOpt{
				{
					Tenant: "cgrates.net",
					Values: []string{"ACC2"},
				},
			},
			Usage: []*DynamicDecimalOpt{
				NewDynamicDecimalOpt([]string{"fld1", "fld2"}, "cgrates.org", utils.NewDecimal(2, 0), nil),
			},
			ProfileIgnoreFilters: []*DynamicBoolOpt{
				{
					Tenant: "cgrates.net",
					value:  false,
				},
			},
			Thresholds: []*DynamicBoolOpt{
				{
					Tenant: "cgrates.org",
					value:  true,
				},
			},
			ThresholdIDs: []*DynamicStringSliceOpt{
				{
					Tenant: "cgrates.org",
					Values: []string{"THR1"},
				},
			},
		},
	}

	expected1 := &AccountSJsonCfg{
		Enabled:         utils.BoolPointer(false),
		Indexed_selects: utils.BoolPointer(false),
		Conns: map[string][]*DynamicConns{
			utils.MetaAttributes: {{ConnIDs: []string{"*localhost", "*birpc"}}},
			utils.MetaRates:      {{ConnIDs: []string{"*localhost"}}},
			utils.MetaThresholds: {{ConnIDs: []string{"*localhost"}}},
		},
		String_indexed_fields: nil,
		Prefix_indexed_fields: nil,
		Suffix_indexed_fields: nil,
		Nested_fields:         utils.BoolPointer(false),
		Max_iterations:        utils.IntPointer(3),
		Max_usage:             utils.StringPointer("60"),
		Opts: &AccountsOptsJson{
			ProfileIDs: []*DynamicStringSliceOpt{
				{
					Tenant: "cgrates.net",
					Values: []string{"ACC2"},
				},
			},
			Usage: []*DynamicInterfaceOpt{
				{
					FilterIDs: []string{"fld1", "fld2"},
					Tenant:    "cgrates.org",
					Value:     "2",
				},
			},
			ProfileIgnoreFilters: []*DynamicInterfaceOpt{
				{
					Tenant: "cgrates.net",
					Value:  false,
				},
			},
			Thresholds: []*DynamicInterfaceOpt{
				{
					Tenant: "cgrates.org",
					Value:  true,
				},
			},
			ThresholdIDs: []*DynamicStringSliceOpt{
				{
					Tenant: "cgrates.org",
					Values: []string{"THR1"},
				},
			},
		},
	}

	rcv := diffAccountSJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected1) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected1), utils.ToJSON(rcv))
	}

	//Make the two Accounts equal in order to get a nil "d"

	v2_3 := v1
	expected3 := &AccountSJsonCfg{
		Enabled:               nil,
		Indexed_selects:       nil,
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
		Enabled: true,
		Conns: map[string][]*DynamicConns{
			utils.MetaAttributes: {{ConnIDs: []string{"*localhost"}}},
			utils.MetaRates:      {{ConnIDs: []string{}}},
			utils.MetaThresholds: {{ConnIDs: []string{}}},
		},
		IndexedSelects:      true,
		StringIndexedFields: &[]string{"~*req.Index1"},
		PrefixIndexedFields: &[]string{},
		SuffixIndexedFields: &[]string{},
		NestedFields:        true,
		MaxIterations:       1,
		MaxUsage:            nil,
		Opts: &AccountsOpts{
			ProfileIDs: []*DynamicStringSliceOpt{
				{
					Values: []string{"ACC1"},
				},
			},
		},
	}

	rcv := acS.CloneSection()
	if !reflect.DeepEqual(acS, rcv.(*AccountSCfg)) {
		t.Errorf("Expected %v \n but received \n %v", acS, rcv.(*AccountSCfg))
	}
}

func TestAccountSCfgThresholdsOpts(t *testing.T) {
	cfgJSONStr := `{
"accounts": {
	"opts": {
		"*thresholdIDs": [
			{
				"tenant": "cgrates.org",
				"values": ["THR_ACC1","THR_ACC2"]
			}
		],
		"*thresholds": [
         {
             "tenant": "cgrates.org",
             "filterIDs": ["*string:~*req.Account:1001"],
             "value": true
         }
        ]
	}
}
}`
	expThresholds := []*DynamicBoolOpt{
		{
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			value:     true,
		},
	}
	expThresholdIDs := []*DynamicStringSliceOpt{
		{
			Tenant: "cgrates.org",
			Values: []string{"THR_ACC1", "THR_ACC2"},
		},
	}
	cfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expThresholds, cfg.accountSCfg.Opts.Thresholds) {
		t.Errorf("Expecting %v,\n received %v", utils.ToJSON(expThresholds), utils.ToJSON(cfg.accountSCfg.Opts.Thresholds))
	}
	if !reflect.DeepEqual(expThresholdIDs, cfg.accountSCfg.Opts.ThresholdIDs) {
		t.Errorf("Expecting %v,\n received %v", utils.ToJSON(expThresholdIDs), utils.ToJSON(cfg.accountSCfg.Opts.ThresholdIDs))
	}
}

func TestAccountsOptsLoadFromJSONCfgThresholdsErr(t *testing.T) {
	accOpts := &AccountsOpts{}
	jsnCfg := &AccountsOptsJson{
		Thresholds: []*DynamicInterfaceOpt{
			{
				Value: "error",
			},
		},
		ThresholdIDs: []*DynamicStringSliceOpt{
			{
				Values: []string{"THD1"},
			},
		},
	}
	expErr := `strconv.ParseBool: parsing "error": invalid syntax`
	if err := accOpts.loadFromJSONCfg(jsnCfg); err == nil || err.Error() != expErr {
		t.Errorf("Expected %v, recieved %v", expErr, err)
	}
}
