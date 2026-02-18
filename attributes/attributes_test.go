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

package attributes

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	expTimeAttributes = time.Now().Add(20 * time.Minute)
	attrS             *AttributeS
	dmAtr             *engine.DataManager
	attrEvs           = []*utils.CGREvent{
		{ //matching AttributeProfile1
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"Attribute":      "AttributeProfile1",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				utils.Weight:     "20.0",
			},
			APIOpts: map[string]any{
				utils.OptsContext: utils.MetaSessionS,
			},
		},
		{ //matching AttributeProfile2
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"Attribute": "AttributeProfile2",
			},
			APIOpts: map[string]any{
				utils.OptsContext: utils.MetaSessionS,
			},
		},
		{ //matching AttributeProfilePrefix
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"Attribute": "AttributeProfilePrefix",
			},
			APIOpts: map[string]any{
				utils.OptsContext: utils.MetaSessionS,
			},
		},
		{ //matching AttributeProfilePrefix
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"DistinctMatch": 20,
			},
			APIOpts: map[string]any{
				utils.OptsContext:               utils.MetaSessionS,
				utils.OptsAttributesProcessRuns: 0,
			},
		},
	}
	atrPs = []*utils.AttributeProfile{
		{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "AttributeProfile1",
			FilterIDs: []string{"FLTR_ATTR_1", "*string:~*opts.*context:*sessions"},
			Attributes: []*utils.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + utils.AccountField,
					Value: utils.NewRSRParsersMustCompile("1010", utils.InfieldSep),
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
		{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "AttributeProfile2",
			FilterIDs: []string{"FLTR_ATTR_2", "*string:~*opts.*context:*sessions"},
			Attributes: []*utils.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + utils.AccountField,
					Value: utils.NewRSRParsersMustCompile("1010", utils.InfieldSep),
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
		{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "AttributeProfilePrefix",
			FilterIDs: []string{"FLTR_ATTR_3", "*string:~*opts.*context:*sessions"},
			Attributes: []*utils.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + utils.AccountField,
					Value: utils.NewRSRParsersMustCompile("1010", utils.InfieldSep),
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
		{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "AttributeIDMatch",
			FilterIDs: []string{"*gte:~*req.DistinctMatch:20"},
			Attributes: []*utils.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + utils.AccountField,
					Value: utils.NewRSRParsersMustCompile("1010", utils.InfieldSep),
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
	}
)

func TestParseAtributeUsageDiffVal1(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"usage1": "20",
			"usage2": "35",
		},
	}
	_, err := ParseAttribute(dp, utils.MetaUsageDifference, utils.EmptyString, utils.NewRSRParsersMustCompile("~*req.wrong;~*req.usage2", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %v\n but received %v", utils.ErrNotFound, err)
	}

}

func TestParseAtributeUsageDiffVal2(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"usage1": "20",
			"usage2": "35",
		},
	}
	_, err := ParseAttribute(dp, utils.MetaUsageDifference, utils.EmptyString, utils.NewRSRParsersMustCompile("~*req.usage1;~*req.wrong", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %v\n but received %v", utils.ErrNotFound, err)
	}
}

func TestParseAtributeUsageDiffTimeVal1(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"usage1": "20",
			"usage2": "35s",
		},
	}
	_, err := ParseAttribute(dp, utils.MetaUsageDifference, utils.EmptyString, utils.NewRSRParsersMustCompile("~*req.usage1;~*req.usage2", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString)
	errExp := "Unsupported time format"
	if err == nil || err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}
}

func TestParseAtributeUsageDiffTimeVal2(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"usage1": "20s",
			"usage2": "35",
		},
	}
	_, err := ParseAttribute(dp, utils.MetaUsageDifference, utils.EmptyString, utils.NewRSRParsersMustCompile("~*req.usage1;~*req.usage2", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString)
	errExp := "Unsupported time format"
	if err == nil || err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}
}

func TestParseAtributeSum(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"not_valid": "field",
		},
	}
	_, err := ParseAttribute(dp, utils.MetaSum, utils.EmptyString, utils.NewRSRParsersMustCompile("~*req.valid", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %v\n but received %v", utils.ErrNotFound, err)
	}
}

func TestParseAtributeDifference(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"not_valid": "field",
		},
	}
	_, err := ParseAttribute(dp, utils.MetaDifference, utils.EmptyString, utils.NewRSRParsersMustCompile("~*req.valid", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %v\n but received %v", utils.ErrNotFound, err)
	}
}

func TestParseAtributeMultiply(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"not_valid": "field",
		},
	}
	_, err := ParseAttribute(dp, utils.MetaMultiply, utils.EmptyString, utils.NewRSRParsersMustCompile("~*req.valid", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %v\n but received %v", utils.ErrNotFound, err)
	}
}

func TestParseAtributeDivide(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"not_valid": "field",
		},
	}
	_, err := ParseAttribute(dp, utils.MetaDivide, utils.EmptyString, utils.NewRSRParsersMustCompile("~*req.valid", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %v\n but received %v", utils.ErrNotFound, err)
	}
}

func TestParseAtributeExponentVal1(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"number":   "20",
			"exponent": "2",
		},
	}
	_, err := ParseAttribute(dp, utils.MetaValueExponent, utils.EmptyString, utils.NewRSRParsersMustCompile("~*req.wrong;~*req.exponent", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %v\n but received %v", utils.ErrNotFound, err)
	}
}

func TestParseAtributeExponentVal2(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"number":   "20",
			"exponent": "2",
		},
	}
	_, err := ParseAttribute(dp, utils.MetaValueExponent, utils.EmptyString, utils.NewRSRParsersMustCompile("~*req.number;~*req.wrong", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %v\n but received %v", utils.ErrNotFound, err)
	}
}

func TestParseAtributeExponentWrongNumber(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"number":   "not_a_number",
			"exponent": "2",
		},
	}
	_, err := ParseAttribute(dp, utils.MetaValueExponent, utils.EmptyString, utils.NewRSRParsersMustCompile("~*req.number;~*req.exponent", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString)
	errExp := "invalid value <not_a_number> to *valueExponent"
	if err == nil || err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}
}

func TestParseAtributeExponentWrongExponent(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"number":   "4",
			"exponent": "NaN",
		},
	}
	_, err := ParseAttribute(dp, utils.MetaValueExponent, utils.EmptyString, utils.NewRSRParsersMustCompile("~*req.number;~*req.exponent", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString)
	errExp := `strconv.Atoi: parsing "NaN": invalid syntax`
	if err == nil || err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}
}

func TestParseAtributeUnixTimestampWrongField(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"unix_timestamp": "not_a_unix_timestamp",
		},
	}
	_, err := ParseAttribute(dp, utils.MetaUnixTimestamp, utils.EmptyString, utils.NewRSRParsersMustCompile("~*req.wrong", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %v\n but received %v", utils.ErrNotFound, err)
	}
}

func TestParseAtributeUnixTimestampWrongVal(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"unix_timestamp": "not_a_unix_timestamp",
		},
	}
	_, err := ParseAttribute(dp, utils.MetaUnixTimestamp, utils.EmptyString, utils.NewRSRParsersMustCompile("~*req.unix_timestamp", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString)
	errExp := "Unsupported time format"
	if err == nil || err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}
}

func TestParseAtributePrefixPath(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"prefix": "prfx",
		},
	}
	_, err := ParseAttribute(dp, utils.MetaPrefix, "```", utils.NewRSRParsersMustCompile("~*req.prefix", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString)
	errExp := "Closed unspilit syntax"
	if err == nil || err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}
}

func TestParseAtributePrefixField(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"prefix": "prfx",
		},
	}
	_, err := ParseAttribute(dp, utils.MetaPrefix, utils.EmptyString, utils.NewRSRParsersMustCompile("~*req.wrong", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %v\n but received %v", utils.ErrNotFound, err)
	}
}

func TestParseAtributeSuffixPath(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"suffix": "sfx",
		},
	}
	_, err := ParseAttribute(dp, utils.MetaSuffix, "```", utils.NewRSRParsersMustCompile("~*req.suffix", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString)
	errExp := "Closed unspilit syntax"
	if err == nil || err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}
}

func TestParseAtributeSuffixField(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"suffix": "sfx",
		},
	}
	_, err := ParseAttribute(dp, utils.MetaSuffix, utils.EmptyString, utils.NewRSRParsersMustCompile("~*req.wrong", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %v\n but received %v", utils.ErrNotFound, err)
	}
}

func TestParseAtributeCCUsageLessThanTwo(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"cc1": "sfx",
			"cc2": "sfx",
			// "cc3": "sfx",
		},
	}
	_, err := ParseAttribute(dp, utils.MetaCCUsage, utils.EmptyString, utils.NewRSRParsersMustCompile("~*req.cc1;~*req.cc2", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString)
	errExp := "invalid arguments <[{\"Rules\":\"~*req.cc1\",\"Path\":\"~*req.cc1\"},{\"Rules\":\"~*req.cc2\",\"Path\":\"~*req.cc2\"}]> to *ccUsage"
	if err == nil || err.Error() != errExp {
		t.Errorf("expected %q, received %q", errExp, err)
	}
}

func TestParseAtributeCCUsageField1(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"cc1": "20",
			"cc2": "sfx",
			"cc3": "sfx",
		},
	}
	_, err := ParseAttribute(dp, utils.MetaCCUsage, utils.EmptyString, utils.NewRSRParsersMustCompile("~*req.wrong;~*req.cc2;~*req.cc3", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %v\n but received %v", utils.ErrNotFound, err)
	}
}

func TestParseAtributeCCUsageVal1(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"cc1": "not_valid",
			"cc2": "sfx",
			"cc3": "sfx",
		},
	}
	_, err := ParseAttribute(dp, utils.MetaCCUsage, utils.EmptyString, utils.NewRSRParsersMustCompile("~*req.cc1;~*req.cc2;~*req.cc3", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString)
	errExp := `invalid requestNumber <not_valid> to *ccUsage`
	if err == nil || err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}
}

func TestParseAtributeCCUsageField2(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"cc1": "20",
			"cc2": "sfx",
			"cc3": "sfx",
		},
	}
	_, err := ParseAttribute(dp, utils.MetaCCUsage, utils.EmptyString, utils.NewRSRParsersMustCompile("~*req.cc1;~*req.wrong;~*req.cc3", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %v\n but received %v", utils.ErrNotFound, err)
	}
}

func TestParseAtributeCCUsageVal2(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"cc1": "20",
			"cc2": "sfx",
			"cc3": "sfx",
		},
	}
	_, err := ParseAttribute(dp, utils.MetaCCUsage, utils.EmptyString, utils.NewRSRParsersMustCompile("~*req.cc1;~*req.cc2;~*req.cc3", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString)
	errExp := `invalid usedCCTime <sfx> to *ccUsage`
	if err == nil || err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}
}

func TestParseAtributeCCUsageField3(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"cc1": "20",
			"cc2": "20",
			"cc3": "sfx",
		},
	}
	_, err := ParseAttribute(dp, utils.MetaCCUsage, utils.EmptyString, utils.NewRSRParsersMustCompile("~*req.cc1;~*req.cc2;~*req.wrong", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %v\n but received %v", utils.ErrNotFound, err)
	}
}

func TestParseAtributeCCUsageVal3(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"cc1": "20",
			"cc2": "20",
			"cc3": "sfx",
		},
	}
	_, err := ParseAttribute(dp, utils.MetaCCUsage, utils.EmptyString, utils.NewRSRParsersMustCompile("~*req.cc1;~*req.cc2;~*req.cc3", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString)
	errExp := `invalid debitInterval <sfx> to *ccUsage`
	if err == nil || err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}
}

func TestParseAtributeCCUsageNoErr(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"cc1": "20",
			"cc2": "20",
			"cc3": "20",
		},
	}
	out, err := ParseAttribute(dp, utils.MetaCCUsage, utils.EmptyString, utils.NewRSRParsersMustCompile("~*req.cc1;~*req.cc2;~*req.cc3", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString)
	if err != nil {
		t.Error(err)
	}
	exp := 400 * time.Nanosecond
	if out.(time.Duration) != exp {
		t.Errorf("Expected %v\n but received %v", exp, out)
	}
}

func TestUniqueAlteredFields(t *testing.T) {
	flds := &AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{
			{Fields: []string{"field1"}},
		},
	}
	exp := make(utils.StringSet)
	for _, altered := range flds.AlteredFields {
		exp.AddSlice(altered.Fields)
	}
	if rcv := flds.UniqueAlteredFields(); !reflect.DeepEqual(exp, rcv) {
		t.Errorf("Expected <%+v>, Received <%+v>", exp, rcv)
	}
}

func TestAttributeProfileForEventWeightFromDynamicsErr(t *testing.T) {

	defer func() {
		engine.Cache = engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	engine.Cache = engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	cfg := config.NewDefaultCGRConfig()
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	attrS := NewAttributeService(dm, filterS, nil, cfg)

	attrIDs, err := utils.OptAsStringSlice(attrEvs[0].APIOpts, utils.OptsAttributesProfileIDs)
	if err != nil {
		t.Fatal(err)
	}

	value := utils.NewRSRParsersMustCompile("abcd123", utils.RSRSep)

	attrPrf := &utils.AttributeProfile{
		Tenant: "cgrates.org",
		ID:     "ATTR_TEST",
		Attributes: []*utils.Attribute{
			{
				Path:  "*req.Password",
				Type:  utils.MetaPassword,
				Value: value,
			},
		},
		Weights: utils.DynamicWeights{
			{
				FilterIDs: []string{"*stirng:~*req.Account:1001"},
				Weight:    10,
			},
		},
	}

	if err := dm.SetAttributeProfile(context.Background(), attrPrf, true); err != nil {
		t.Fatal(err)
	}

	attrEvs := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		Event: map[string]any{
			"Attribute":      "AttributeProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			utils.Weight:     "20.0",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}

	expErr := "NOT_IMPLEMENTED:*stirng"
	_, err = attrS.attributeProfileForEvent(context.TODO(), attrEvs.Tenant,
		attrIDs, utils.MapStorage{
			utils.MetaReq:  attrEvs.Event,
			utils.MetaOpts: attrEvs.APIOpts,
			utils.MetaVars: utils.MapStorage{
				utils.OptsAttributesProcessRuns: 0,
			},
		}, utils.EmptyString, make(map[string]int), 0, false)
	if err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%+v>, received error <%+v>", expErr, err)
	}

}

func TestAttributeProcessEventBlockerFromDynamicsErr(t *testing.T) {
	defer func() {
		engine.Cache = engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()
	cfg := config.NewDefaultCGRConfig()
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	attrS := NewAttributeService(dm, filterS, nil, cfg)

	value := utils.NewRSRParsersMustCompile("abcd123", utils.RSRSep)

	attrPrf := &utils.AttributeProfile{
		Tenant: "cgrates.org",
		ID:     "ATTR_TEST",
		Attributes: []*utils.Attribute{
			{
				Path:  "*req.Password",
				Type:  utils.MetaPassword,
				Value: value,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				FilterIDs: []string{"*stirng:~*req.Account:1001"},
				Blocker:   false,
			},
		},
	}

	if err := dm.SetAttributeProfile(context.Background(), attrPrf, true); err != nil {
		t.Fatal(err)
	}

	attrEvs := &utils.CGREvent{

		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"Attribute":      "AttributeProfile1",
			"Account":        "1010",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			utils.Weight:     "20.0",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}

	eNM := utils.MapStorage{
		utils.MetaReq:  attrEvs.Event,
		utils.MetaOpts: attrEvs.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}

	expErr := "NOT_IMPLEMENTED:*stirng"
	_, err = attrS.processEvent(context.TODO(), attrEvs.Tenant, attrEvs, eNM, engine.NewDynamicDP(context.TODO(), nil, nil, nil, nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
	if err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%+v>, received error <%+v>", expErr, err)
	}

}

func TestAttributeSProcessEventPassErr(t *testing.T) {

	defer func() {
		engine.Cache = engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	attrS := NewAttributeService(dm, filterS, nil, cfg)

	attrPrf := &utils.AttributeProfile{
		Tenant: "cgrates.org",
		ID:     "ATTR_TEST",
		Attributes: []*utils.Attribute{
			{
				FilterIDs: []string{"*apiban:~*req.<~*req.IP>{*}:*all"},
				Path:      "*req.Password",
				Type:      utils.MetaPassword,
				Value:     utils.NewRSRParsersMustCompile("abcd123", utils.RSRSep),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}

	if err := dm.SetAttributeProfile(context.TODO(), attrPrf, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"PassField": "Test",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}

	eNM := utils.MapStorage{
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
		utils.MetaReq: utils.MapStorage{
			"bannedIP":  "1.2.3.251",
			"bannedIP2": "1.2.3.252",
			"IP":        "1.2.3.253",
			"IP2":       "1.2.3.254",
		},
	}

	expErr := `invalid converter value in string: <*>, err: unsupported converter definition: <*>`
	_, err = attrS.processEvent(context.TODO(), attrPrf.Tenant, ev, eNM, engine.NewDynamicDP(context.TODO(), nil, nil, nil, nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
	if err == nil || err.Error() != expErr {
		t.Errorf("Expected error %s received: %v", expErr, err)
	}

}

func TestAttributeSProcessAttrBlockerFromDynamicsErr(t *testing.T) {
	defer func() {
		engine.Cache = engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	attrS := NewAttributeService(dm, filterS, nil, cfg)

	attrPrf := &utils.AttributeProfile{
		Tenant: "cgrates.org",
		ID:     "ATTR_TEST",
		Attributes: []*utils.Attribute{
			{
				Blockers: utils.DynamicBlockers{{
					FilterIDs: []string{"*stirng:~*req.Account:1001"},
					Blocker:   false,
				}},
				Path:  "*req.Password",
				Type:  utils.MetaPassword,
				Value: utils.NewRSRParsersMustCompile("abcd123", utils.RSRSep),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}

	if err := dm.SetAttributeProfile(context.TODO(), attrPrf, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"PassField": "Test",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}

	eNM := utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}

	expErr := "NOT_IMPLEMENTED:*stirng"
	_, err = attrS.processEvent(context.TODO(), attrPrf.Tenant, ev, eNM, engine.NewDynamicDP(context.TODO(), nil, nil, nil, nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
	if err == nil || err.Error() != expErr {
		t.Errorf("Expected error %s received: %v", expErr, err)
	}

}

func TestAttributeSProcessSubstituteRmvBlockerTrue(t *testing.T) {
	defer func() {
		engine.Cache = engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	attrS := NewAttributeService(dm, filterS, nil, cfg)

	attrPrf := &utils.AttributeProfile{
		Tenant: "cgrates.org",
		ID:     "ATTR_TEST",
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaRemove,
				Type:  utils.MetaVariable,
				Value: utils.NewRSRParsersMustCompile(utils.MetaRemove, utils.RSRSep),
			},
			{
				Blockers: utils.DynamicBlockers{{
					Blocker: true,
				}},
				Path:  "*req.Password",
				Type:  utils.MetaPassword,
				Value: utils.NewRSRParsersMustCompile("abcd123", utils.RSRSep),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}

	if err := dm.SetAttributeProfile(context.TODO(), attrPrf, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"PassField": "Test",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}

	eNM := utils.MapStorage{
		utils.MetaReq:      ev.Event,
		utils.MetaOpts:     ev.APIOpts,
		utils.MetaVariable: utils.MetaRemove,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}

	exp := &AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{{
			MatchedProfileID: "cgrates.org:ATTR_TEST",
			Fields:           []string{utils.MetaRemove, "*req.Password"},
		}},
		CGREvent: ev,
	}

	rcv, err := attrS.processEvent(context.TODO(), attrPrf.Tenant, ev, eNM, engine.NewDynamicDP(context.TODO(), nil, nil, nil, nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(utils.ToJSON(exp), utils.ToJSON(rcv)) {
		t.Errorf("Expected \n<%+v>,\n Received \n<%+v>", utils.ToJSON(exp), utils.ToJSON(rcv))
	}

}

func TestV1GetAttributeForEventAttrProfEventErr(t *testing.T) {

	defer func() {
		engine.Cache = engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	conMng := engine.NewConnManager(cfg)
	db, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: db}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, conMng)
	filterS := engine.NewFilterS(cfg, conMng, dm)
	attr := &utils.AttributeProfile{
		Tenant: "cgrates.org",
		ID:     "AttributeProfile1",
		Attributes: []*utils.Attribute{
			{
				Path:  "*tenant",
				Type:  "*variable",
				Value: utils.NewRSRParsersMustCompile("~*req.Account:s/(.*)@(.*)/${1}.${2}/", utils.InfieldSep),
			},
		},

		Weights: utils.DynamicWeights{
			{
				FilterIDs: []string{"*stirng:~*req.Account:1001"},
				Weight:    20,
			},
		},
	}
	err = dm.SetAttributeProfile(context.Background(), attr, true)
	if err != nil {
		t.Error(err)
	}

	alS := NewAttributeService(dm, filterS, nil, cfg)
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"Attribute": "AttributeProfile1",
		},
		APIOpts: map[string]any{},
	}
	var rply utils.APIAttributeProfile
	expErr := "SERVER_ERROR: NOT_IMPLEMENTED:*stirng"
	err = alS.V1GetAttributeForEvent(context.Background(), ev, &rply)
	if err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%+v>, received error <%+v>", expErr, err)
	}

}

func TestAttributesV1ProcessEventFieldMissingErr(t *testing.T) {

	defer func() {
		engine.Cache = engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.FilterSCfg().ResourceSConns = []string{}
	conMng := engine.NewConnManager(cfg)
	db, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: db}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, conMng)
	filterS := engine.NewFilterS(cfg, conMng, dm)
	attr := &utils.AttributeProfile{
		Tenant: "cgrates.org",
		ID:     "ATTR_CHANGE_TENANT_FROM_USER",
		Attributes: []*utils.Attribute{
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*variable",
				Value:     utils.NewRSRParsersMustCompile("~*req.Account:s/(.*)@(.*)/${1}.${2}/", utils.InfieldSep),
			},
		},
		Blockers: utils.DynamicBlockers{{Blocker: false}},
		Weights:  utils.DynamicWeights{{Weight: 20}},
	}
	err = dm.SetAttributeProfile(context.Background(), attr, true)
	if err != nil {
		t.Error(err)
	}

	alS := NewAttributeService(dm, filterS, nil, cfg)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "123",
		Event: map[string]any{
			"testfield": utils.MetaAttributes,
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProcessRuns: 2,
		},
	}
	var rply AttrSProcessEventReply
	expErr := "MANDATORY_IE_MISSING: [testfield]"
	err = alS.V1ProcessEvent(context.Background(), ev, &rply)
	if err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%+v>, received error <%+v>", expErr, err)
	}

}

func TestParseAtributeUsageDiffDetectLayoutErr2(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"UnixTimeStamp": "1554364297",
			"usage2":        "35",
		},
	}
	_, err := ParseAttribute(dp, utils.MetaUsageDifference, utils.EmptyString, utils.NewRSRParsersMustCompile("~*req.UnixTimeStamp;~*req.usage2", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString)
	errExp := "Unsupported time format"
	if err == nil || err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}
}

func TestParseAtributeMetaPrefixParseDPErr(t *testing.T) {
	dp := utils.MapStorage{}

	_, err := ParseAttribute(dp, utils.MetaPrefix, "constant;`>;q=0.7;expires=3600`;~*req.Account", utils.NewRSRParsersMustCompile("~*req.UnixTimeStamp;~*req.usage2", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString)
	if err != utils.ErrNotFound {
		t.Errorf("Expected %v\n but received %v", utils.ErrNotFound, err)
	}
}

func TestParseAtributeMetaSuffixParseDPErr(t *testing.T) {
	dp := utils.MapStorage{}

	_, err := ParseAttribute(dp, utils.MetaSuffix, "constant;`>;q=0.7;expires=3600`;~*req.Account", utils.NewRSRParsersMustCompile("~*req.UnixTimeStamp;~*req.usage2", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString)
	if err != utils.ErrNotFound {
		t.Errorf("Expected %v\n but received %v", utils.ErrNotFound, err)
	}
}

func TestParseAtributeCCUsageNegativeReqNr(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"cc1": "-1",
			"cc2": "-20",
			"cc3": "-20",
		},
	}
	out, err := ParseAttribute(dp, utils.MetaCCUsage, utils.EmptyString, utils.NewRSRParsersMustCompile("~*req.cc1;~*req.cc2;~*req.cc3", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString)
	if err != nil {
		t.Error(err)
	}
	exp := -20 * time.Nanosecond
	if out.(time.Duration) != exp {
		t.Errorf("Expected %v\n but received %v", exp, out)
	}
}
func TestAttributeFromHTTP(t *testing.T) {
	exp := "Account"
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, exp)

	}))

	defer testServer.Close()
	attrType := utils.MetaHTTP + utils.HashtagSep + utils.IdxStart + testServer.URL + utils.IdxEnd

	attrID := attrType + ":*req.Category:*attributes"
	expAttrPrf1 := &utils.AttributeProfile{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     attrType + ":*req.Category:*attributes",
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Category",
				Type:  attrType,
				Value: utils.NewRSRParsersMustCompile("*attributes", utils.InfieldSep),
			},
		},
	}
	attrPrf, err := utils.NewAttributeFromInline(config.CgrConfig().GeneralCfg().DefaultTenant, attrID)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expAttrPrf1, attrPrf) {
		t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(expAttrPrf1), utils.ToJSON(attrPrf))
	}
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{},
	}

	attr := attrPrf.Attributes[0]
	if out, err := ParseAttribute(dp, attr.Type, attr.Path, attr.Value,
		0, utils.EmptyString, utils.EmptyString); err != nil {
		t.Fatal(err)
	} else if exp != out {
		t.Errorf("Expected %q, Received %q", exp, out)
	}
}

func TestAttributesV1GetAttributeForEventProfileIgnoreOpts(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.FilterSCfg().ResourceSConns = []string{}
	conMng := engine.NewConnManager(cfg)
	db, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: db}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, conMng)
	filterS := engine.NewFilterS(cfg, conMng, dm)
	aA := NewAttributeService(dm, filterS, nil, cfg)
	cfg.AttributeSCfg().Opts.ProfileIgnoreFilters = []*config.DynamicBoolOpt{
		config.NewDynamicBoolOpt(nil, "", true, nil),
	}
	acPrf := &utils.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "AC1",
		FilterIDs: []string{"*string:~*req.Attribute:testAttrValue"},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	if err := dm.SetAttributeProfile(context.Background(), acPrf, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "AcProcessEvent1",
		Event: map[string]any{
			"Attribute": "testAttrValue",
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProcessRuns: 2,
			utils.OptsAttributesProfileIDs:  []string{"AC1"},
			utils.MetaProfileIgnoreFilters:  false,
		},
	}
	rply := &utils.APIAttributeProfile{}
	expected := &utils.APIAttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "AC1",
		FilterIDs: []string{"*string:~*req.Attribute:testAttrValue"},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		Attributes: []*utils.ExternalAttribute{},
	}

	err = aA.V1GetAttributeForEvent(context.Background(), ev, rply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expected), utils.ToJSON(rply))
	}
	// correct filter but ignore filters opt on false
	ev2 := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "AcProcessEvent2",
		Event: map[string]any{
			"Attribute": "testAttrValue2",
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProcessRuns: 2,
			utils.OptsAttributesProfileIDs:  []string{"AC1"},
			utils.MetaProfileIgnoreFilters:  true,
		},
	}
	rply2 := &utils.APIAttributeProfile{}
	expected2 := &utils.APIAttributeProfile{
		Tenant:     "cgrates.org",
		ID:         "AC1",
		FilterIDs:  []string{"*string:~*req.Attribute:testAttrValue"},
		Attributes: []*utils.ExternalAttribute{},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	// with ignore filters on true and with bad filter
	err = aA.V1GetAttributeForEvent(context.Background(), ev2, rply2)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(expected2, rply2) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expected2), utils.ToJSON(rply2))
	}
}

func TestAttributesV1GetAttributeForEventErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.FilterSCfg().ResourceSConns = []string{}
	conMng := engine.NewConnManager(cfg)
	db, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: db}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, conMng)
	filterS := engine.NewFilterS(cfg, conMng, dm)
	attr := &utils.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_CHANGE_TENANT_FROM_USER",
		FilterIDs: []string{"*string:~*req.Account:dan@itsyscom.com|adrian@itsyscom.com"},
		Attributes: []*utils.Attribute{
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*variable",
				Value:     utils.NewRSRParsersMustCompile("~*req.Account:s/(.*)@(.*)/${1}.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*req.Account",
				Type:      "*variable",
				Value:     utils.NewRSRParsersMustCompile("~*req.Account:s/(dan)@(.*)/${1}.${2}/:s/(adrian)@(.*)/andrei.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*composed",
				Value:     utils.NewRSRParsersMustCompile(".co.uk", utils.InfieldSep),
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: false,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	err = dm.SetAttributeProfile(context.Background(), attr, true)
	if err != nil {
		t.Error(err)
	}

	attr2 := &utils.AttributeProfile{
		Tenant: "adrian.itsyscom.com.co.uk",
		ID:     "ATTR_MATCH_TENANT",
		Attributes: []*utils.Attribute{
			{
				FilterIDs: nil,
				Path:      "*req.Password",
				Type:      utils.MetaConstant,
				Value:     utils.NewRSRParsersMustCompile("CGRATES.ORG", utils.InfieldSep),
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: false,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}

	err = dm.SetAttributeProfile(context.Background(), attr2, true)
	if err != nil {
		t.Error(err)
	}

	alS := NewAttributeService(dm, filterS, nil, cfg)
	var ev utils.CGREvent
	rply := &utils.APIAttributeProfile{}
	err = alS.V1GetAttributeForEvent(context.Background(), &ev, rply)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrNotFound, err)
	}

}
func TestAttributePopulateAttrService(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().StringIndexedFields = nil
	cfg.AttributeSCfg().PrefixIndexedFields = nil
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dmAtr = engine.NewDataManager(dbCM, cfg, nil)
	fltrs := engine.NewFilterS(cfg, nil, dmAtr)
	attrS = NewAttributeService(dmAtr, fltrs, nil, cfg)
}

func TestAttributeAddFilters(t *testing.T) {
	fltrAttr1 := &engine.Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_ATTR_1",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Attribute",
				Values:  []string{"AttributeProfile1"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: "~*req.UsageInterval",
				Values:  []string{time.Second.String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: "~*req." + utils.Weight,
				Values:  []string{"9.0"},
			},
		},
	}
	dmAtr.SetFilter(context.Background(), fltrAttr1, true)
	fltrAttr2 := &engine.Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_ATTR_2",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Attribute",
				Values:  []string{"AttributeProfile2"},
			},
		},
	}
	dmAtr.SetFilter(context.Background(), fltrAttr2, true)
	fltrAttrPrefix := &engine.Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_ATTR_3",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Attribute",
				Values:  []string{"AttributeProfilePrefix"},
			},
		},
	}
	dmAtr.SetFilter(context.Background(), fltrAttrPrefix, true)
	fltrAttr4 := &engine.Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_ATTR_4",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: "~*req." + utils.Weight,
				Values:  []string{"200.00"},
			},
		},
	}
	dmAtr.SetFilter(context.Background(), fltrAttr4, true)
}

func TestAttributeCache(t *testing.T) {
	for _, atr := range atrPs {
		if err := dmAtr.SetAttributeProfile(context.TODO(), atr, true); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//verify each attribute from cache
	for _, atr := range atrPs {
		if tempAttr, err := dmAtr.GetAttributeProfile(context.TODO(), atr.Tenant, atr.ID,
			true, false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(atr, tempAttr) {
			t.Errorf("Expecting: %+v, received: %+v", atr, tempAttr)
		}
	}
}

func TestAttributeProfileForEvent(t *testing.T) {
	attrIDs, err := utils.OptAsStringSlice(attrEvs[0].APIOpts, utils.OptsAttributesProfileIDs)
	if err != nil {
		t.Fatal(err)
	}
	atrp, err := attrS.attributeProfileForEvent(context.TODO(), attrEvs[0].Tenant,
		attrIDs, utils.MapStorage{
			utils.MetaReq:  attrEvs[0].Event,
			utils.MetaOpts: attrEvs[0].APIOpts,
			utils.MetaVars: utils.MapStorage{
				utils.OptsAttributesProcessRuns: 0,
			},
		}, utils.EmptyString, make(map[string]int), 0, false)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(atrPs[0], atrp) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(atrPs[0]), utils.ToJSON(atrp))
	}

	attrIDs, err = utils.OptAsStringSlice(attrEvs[1].APIOpts, utils.OptsAttributesProfileIDs)
	if err != nil {
		t.Fatal(err)
	}
	atrp, err = attrS.attributeProfileForEvent(context.TODO(), attrEvs[1].Tenant,
		attrIDs, utils.MapStorage{
			utils.MetaReq:  attrEvs[1].Event,
			utils.MetaOpts: attrEvs[1].APIOpts,
			utils.MetaVars: utils.MapStorage{
				utils.OptsAttributesProcessRuns: 0,
			},
		}, utils.EmptyString, make(map[string]int), 0, false)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(atrPs[1], atrp) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(atrPs[1]), utils.ToJSON(atrp))
	}

	attrIDs, err = utils.OptAsStringSlice(attrEvs[2].APIOpts, utils.OptsAttributesProfileIDs)
	if err != nil {
		t.Fatal(err)
	}
	atrp, err = attrS.attributeProfileForEvent(context.TODO(), attrEvs[2].Tenant,
		attrIDs, utils.MapStorage{
			utils.MetaReq:  attrEvs[2].Event,
			utils.MetaOpts: attrEvs[2].APIOpts,
			utils.MetaVars: utils.MapStorage{
				utils.OptsAttributesProcessRuns: 0,
			},
		}, utils.EmptyString, make(map[string]int), 0, false)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(atrPs[2], atrp) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(atrPs[2]), utils.ToJSON(atrp))
	}
}

func TestAttributeProcessEvent(t *testing.T) {
	attrEvs[0].Event["Account"] = "1010" //Field added in event after process
	eRply := &AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{{
			MatchedProfileID: "cgrates.org:AttributeProfile1",
			Fields:           []string{utils.MetaReq + utils.NestingSep + "Account"},
		}},
		CGREvent: attrEvs[0],
	}
	eNM := utils.MapStorage{
		utils.MetaReq:  attrEvs[0].Event,
		utils.MetaOpts: attrEvs[0].APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}
	atrp, err := attrS.processEvent(context.TODO(), attrEvs[0].Tenant, attrEvs[0], eNM, engine.NewDynamicDP(context.TODO(), nil, nil, nil, nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eRply, atrp) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eRply), utils.ToJSON(atrp))
	}
}

func TestAttributeProcessEventWithNotFound(t *testing.T) {
	attrEvs[3].Event["Account"] = "1010" //Field added in event after process
	eNM := utils.MapStorage{
		utils.MetaReq:  attrEvs[3].Event,
		utils.MetaOpts: attrEvs[3].APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}
	if _, err := attrS.processEvent(context.TODO(), attrEvs[0].Tenant, attrEvs[3], eNM,
		engine.NewDynamicDP(context.TODO(), nil, nil, nil, nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0); err == nil || err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestAttributeProcessEventWithIDs(t *testing.T) {
	attrEvs[3].Event["Account"] = "1010" //Field added in event after process
	attrEvs[3].APIOpts[utils.OptsAttributesProfileIDs] = []string{"AttributeIDMatch"}
	eRply := &AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{{
			MatchedProfileID: "cgrates.org:AttributeIDMatch",
			Fields:           []string{utils.MetaReq + utils.NestingSep + "Account"},
		}},
		CGREvent: attrEvs[3],
	}
	eNM := utils.MapStorage{
		utils.MetaReq:  attrEvs[3].Event,
		utils.MetaOpts: attrEvs[3].APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}
	if atrp, err := attrS.processEvent(context.TODO(), attrEvs[0].Tenant, attrEvs[3], eNM, engine.NewDynamicDP(context.TODO(), nil, nil, nil, nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0); err != nil {
	} else if !reflect.DeepEqual(eRply, atrp) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eRply), utils.ToJSON(atrp))
	}
}

func TestAttributeEventReplyDigest(t *testing.T) {
	eRpl := &AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{{
			MatchedProfileID: "cgrates.org:ATTR_1",
			Fields:           []string{utils.AccountField, utils.Subject},
		}},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSProcessEvent",
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.Subject:      "1001",
			},
		},
	}
	expRpl := "Account:1001,Subject:1001"
	val := eRpl.Digest()
	if !reflect.DeepEqual(val, expRpl) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expRpl), utils.ToJSON(val))
	}
}

func TestAttributeEventReplyDigest2(t *testing.T) {
	eRpl := &AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{{
			MatchedProfileID: "cgrates.org:ATTR_1",
			Fields:           []string{},
		}},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSProcessEvent",
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.Subject:      "1001",
			},
		},
	}
	expRpl := ""
	val := eRpl.Digest()
	if !reflect.DeepEqual(val, expRpl) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expRpl), utils.ToJSON(val))
	}
}

func TestAttributeEventReplyDigest3(t *testing.T) {
	eRpl := &AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{{
			MatchedProfileID: "cgrates.org:ATTR_1",
			Fields:           []string{"*req.Subject"},
		}},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSProcessEvent",
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.Subject:      "1001",
			},
		},
	}
	expRpl := "Subject:1001"
	val := eRpl.Digest()
	if !reflect.DeepEqual(val, expRpl) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expRpl), utils.ToJSON(val))
	}
}

func TestAttributeEventReplyDigest4(t *testing.T) {
	eRpl := &AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{{
			MatchedProfileID: "cgrates.org:ATTR_1",
			Fields:           []string{"*req.Subject"},
		}},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSProcessEvent",
			Event: map[string]any{
				utils.AccountField: "1001",
			},
		},
	}
	expRpl := ""
	val := eRpl.Digest()
	if !reflect.DeepEqual(val, expRpl) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expRpl), utils.ToJSON(val))
	}
}

func TestAttributeIndexer(t *testing.T) {
	//refresh the DM
	if err := dmAtr.DataDB()[utils.MetaDefault].Flush(""); err != nil {
		t.Error(err)
	}
	engine.Cache.Clear(nil)
	if test, err := dmAtr.DataDB()[utils.MetaDefault].IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("Expecting: true got :%+v", test)
	}
	expTimeStr := expTimeAttributes.Format("2006-01-02T15-04-05Z")
	attrPrf := &utils.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "AttrPrf",
		FilterIDs: []string{"*string:~*req.Account:1007", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z|" + expTimeStr},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + utils.AccountField,
				Value: utils.NewRSRParsersMustCompile("1010", utils.InfieldSep),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf, true); err != nil {
		t.Error(err)
	}
	eIdxes := map[string]utils.StringSet{
		"*string:*req.Account:1007": {
			"AttrPrf": struct{}{},
		},
	}
	if rcvIdx, err := dmAtr.GetIndexes(context.TODO(), utils.CacheAttributeFilterIndexes,
		attrPrf.Tenant, "", utils.NonTransactional, false, false); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(eIdxes, rcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
		}
	}
	//Set AttributeProfile with new context (*sessions)
	cpAttrPrf := new(utils.AttributeProfile)
	*cpAttrPrf = *attrPrf
	cpAttrPrf.FilterIDs = append(attrPrf.FilterIDs, "*string:~*opts.*context:*sessions")
	eIdxes["*string:*opts.*context:*sessions"] = utils.StringSet{
		"AttrPrf": {},
	}
	if err := dmAtr.SetAttributeProfile(context.TODO(), cpAttrPrf, true); err != nil {
		t.Error(err)
	}
	if rcvIdx, err := dmAtr.GetIndexes(context.TODO(), utils.CacheAttributeFilterIndexes,
		attrPrf.Tenant, "", utils.NonTransactional, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIdxes, rcvIdx) {
		t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(eIdxes), utils.ToJSON(rcvIdx))
	}

	expected := map[string]utils.StringSet{
		"*string:*opts.*context:*sessions": {
			"AttrPrf": {},
		},
		"*string:*req.Account:1007": {
			"AttrPrf": {},
		},
	}
	//verify if old index was deleted ( context *any)
	if rcv, err := dmAtr.GetIndexes(context.TODO(), utils.CacheAttributeFilterIndexes,
		attrPrf.Tenant, "", utils.NonTransactional, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", expected, rcv)
	}
}

func TestAttributeProcessWithMultipleRuns1(t *testing.T) {
	//refresh the DM
	if err := dmAtr.DataDB()[utils.MetaDefault].Flush(""); err != nil {
		t.Error(err)
	}
	if test, err := dmAtr.DataDB()[utils.MetaDefault].IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("Expecting: true got :%+v", test)
	}
	attrPrf1 := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		FilterIDs: []string{"*string:~*req.InitialField:InitialValue", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field1",
				Value: utils.NewRSRParsersMustCompile("Value1", utils.InfieldSep),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	attrPrf2 := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_2",
		FilterIDs: []string{"*string:~*req.Field1:Value1", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Value: utils.NewRSRParsersMustCompile("Value2", utils.InfieldSep),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	attrPrf3 := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_3",
		FilterIDs: []string{"*string:~*req.Field2:Value2", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field3",
				Value: utils.NewRSRParsersMustCompile("Value3", utils.InfieldSep),
			},
			{
				Path:  utils.MetaReq + utils.NestingSep + "Destination",
				Value: utils.NewRSRParsersMustCompile("2044", utils.InfieldSep),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf1, true); err != nil {
		t.Error(err)
	}
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf2, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf3, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"InitialField": "InitialValue",
		},
		APIOpts: map[string]any{
			utils.OptsContext:               utils.MetaSessionS,
			utils.OptsAttributesProcessRuns: 3,
		},
	}

	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
		t.Errorf("Error: %+v", err)
	}
	eRply := AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_1",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field1"},
			},
			{
				MatchedProfileID: "cgrates.org:ATTR_2",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field2"},
			},
			{
				MatchedProfileID: "cgrates.org:ATTR_3",
				Fields: []string{
					utils.MetaReq + utils.NestingSep + "Field3",
					utils.MetaReq + utils.NestingSep + "Destination"},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     reply.CGREvent.ID,
			Event: map[string]any{
				utils.Destination: "2044",
				"InitialField":    "InitialValue",
				"Field1":          "Value1",
				"Field2":          "Value2",
				"Field3":          "Value3",
			},
			APIOpts: map[string]any{
				utils.OptsContext:               utils.MetaSessionS,
				utils.OptsAttributesProcessRuns: 3,
			},
		},
		blocker: false,
	}
	if !reflect.DeepEqual(eRply, reply) {
		t.Fatalf("Expecting %+v, received: %+v", utils.ToJSON(eRply), utils.ToJSON(reply))
	}
}

func TestAttributeProcessWithMultipleRuns2(t *testing.T) {
	//refresh the DM
	if err := dmAtr.DataDB()[utils.MetaDefault].Flush(""); err != nil {
		t.Error(err)
	}
	engine.Cache.Clear(nil)
	if test, err := dmAtr.DataDB()[utils.MetaDefault].IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("Expecting: true got :%+v", test)
	}
	attrPrf1 := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		FilterIDs: []string{"*string:~*req.InitialField:InitialValue", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field1",
				Value: utils.NewRSRParsersMustCompile("Value1", utils.InfieldSep),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	attrPrf2 := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_2",
		FilterIDs: []string{"*string:~*req.Field1:Value1", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Value: utils.NewRSRParsersMustCompile("Value2", utils.InfieldSep),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	attrPrf3 := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_3",
		FilterIDs: []string{"*string:~*req.NotFound:NotFound", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field3",
				Value: utils.NewRSRParsersMustCompile("Value3", utils.InfieldSep),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf1, true); err != nil {
		t.Error(err)
	}
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf2, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf3, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"InitialField": "InitialValue",
		},
		APIOpts: map[string]any{
			utils.OptsContext:               utils.MetaSessionS,
			utils.OptsAttributesProcessRuns: 4,
		},
	}
	eRply := &AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_1",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field1"},
			},
			{
				MatchedProfileID: "cgrates.org:ATTR_2",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field2"},
			},
			{
				MatchedProfileID: "cgrates.org:ATTR_1",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field1"},
			},
			{
				MatchedProfileID: "cgrates.org:ATTR_2",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field2"},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"InitialField": "InitialValue",
				"Field1":       "Value1",
				"Field2":       "Value2",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eRply.AlteredFields, reply.AlteredFields) {
		t.Errorf("Expecting %+v, received: %+v", eRply.AlteredFields, reply.AlteredFields)
	}
	if !reflect.DeepEqual(eRply.CGREvent.Event, reply.CGREvent.Event) {
		t.Errorf("Expecting %+v, received: %+v", eRply.CGREvent.Event, reply.CGREvent.Event)
	}
}

func TestAttributeProcessWithMultipleRuns3(t *testing.T) {
	//refresh the DM
	if err := dmAtr.DataDB()[utils.MetaDefault].Flush(""); err != nil {
		t.Error(err)
	}
	engine.Cache.Clear(nil)
	if test, err := dmAtr.DataDB()[utils.MetaDefault].IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("Expecting: true got :%+v", test)
	}
	attrPrf1 := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		FilterIDs: []string{"*string:~*req.InitialField:InitialValue", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field1",
				Value: utils.NewRSRParsersMustCompile("Value1", utils.InfieldSep),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	attrPrf2 := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_2",
		FilterIDs: []string{"*string:~*req.Field1:Value1", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Value: utils.NewRSRParsersMustCompile("Value2", utils.InfieldSep),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	attrPrf3 := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_3",
		FilterIDs: []string{"*string:~*req.Field2:Value2", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field3",
				Value: utils.NewRSRParsersMustCompile("Value3", utils.InfieldSep),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf1, true); err != nil {
		t.Error(err)
	}
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf2, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf3, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"InitialField": "InitialValue",
		},
		APIOpts: map[string]any{
			utils.OptsContext:               utils.MetaSessionS,
			utils.OptsAttributesProcessRuns: 2,
		},
	}
	eRply := &AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_1",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field1"},
			},
			{
				MatchedProfileID: "cgrates.org:ATTR_2",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field2"},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"InitialField": "InitialValue",
				"Field1":       "Value1",
				"Field2":       "Value2",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
		t.Errorf("Error: %+v", err)
	}
	for idx, altered := range eRply.AlteredFields {
		if altered.MatchedProfileID != reply.AlteredFields[idx].MatchedProfileID {
			t.Errorf("Expecting %+v, received: %+v", altered.MatchedProfileID, reply.AlteredFields[idx].MatchedProfileID)
		}
	}
	if !reflect.DeepEqual(eRply.CGREvent.Event, reply.CGREvent.Event) {
		t.Errorf("Expecting %+v, received: %+v", eRply.CGREvent.Event, reply.CGREvent.Event)
	}
}

func TestAttributeProcessWithMultipleRuns4(t *testing.T) {
	//refresh the DM
	if err := dmAtr.DataDB()[utils.MetaDefault].Flush(""); err != nil {
		t.Error(err)
	}
	engine.Cache.Clear(nil)
	if test, err := dmAtr.DataDB()[utils.MetaDefault].IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("Expecting: true got :%+v", test)
	}
	attrPrf1 := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		FilterIDs: []string{"*string:~*req.InitialField:InitialValue", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field1",
				Value: utils.NewRSRParsersMustCompile("Value1", utils.InfieldSep),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	attrPrf2 := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_2",
		FilterIDs: []string{"*string:~*req.Field1:Value1", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Value: utils.NewRSRParsersMustCompile("Value2", utils.InfieldSep),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf1, true); err != nil {
		t.Error(err)
	}
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf2, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"InitialField": "InitialValue",
		},
		APIOpts: map[string]any{
			utils.OptsContext:               utils.MetaSessionS,
			utils.OptsAttributesProcessRuns: 4,
		},
	}
	eRply := &AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_1",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field1"},
			},
			{
				MatchedProfileID: "cgrates.org:ATTR_2",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field2"},
			},
			{
				MatchedProfileID: "cgrates.org:ATTR_1",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field1"},
			},
			{
				MatchedProfileID: "cgrates.org:ATTR_2",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field2"},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"InitialField": "InitialValue",
				"Field1":       "Value1",
				"Field2":       "Value2",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eRply.AlteredFields, reply.AlteredFields) {
		t.Errorf("Expecting %+v, received: %+v", eRply.AlteredFields, reply.AlteredFields)
	}
	if !reflect.DeepEqual(eRply.CGREvent.Event, reply.CGREvent.Event) {
		t.Errorf("Expecting %+v, received: %+v", eRply.CGREvent.Event, reply.CGREvent.Event)
	}
}

func TestAttributeMultipleProcessWithBlocker(t *testing.T) {
	//refresh the DM
	if err := dmAtr.DataDB()[utils.MetaDefault].Flush(""); err != nil {
		t.Error(err)
	}
	engine.Cache.Clear(nil)
	if test, err := dmAtr.DataDB()[utils.MetaDefault].IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("Expecting: true got :%+v", test)
		return
	}
	attrPrf1 := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		FilterIDs: []string{"*string:~*req.InitialField:InitialValue", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field1",
				Value: utils.NewRSRParsersMustCompile("Value1", utils.InfieldSep),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	attrPrf2 := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_2",
		FilterIDs: []string{"*string:~*req.Field1:Value1", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Value: utils.NewRSRParsersMustCompile("Value2", utils.InfieldSep),
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: true,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	attrPrf3 := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_3",
		FilterIDs: []string{"*string:~*req.Field2:Value2", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field3",
				Value: utils.NewRSRParsersMustCompile("Value3", utils.InfieldSep),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf1, true); err != nil {
		t.Error(err)
	}
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf2, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf3, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"InitialField": "InitialValue",
		},
		APIOpts: map[string]any{
			utils.OptsContext:               utils.MetaSessionS,
			utils.OptsAttributesProcessRuns: 4,
		},
	}
	eRply := &AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_1",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field1"},
			},
			{
				MatchedProfileID: "cgrates.org:ATTR_2",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field2"},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"InitialField": "InitialValue",
				"Field1":       "Value1",
				"Field2":       "Value2",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eRply.AlteredFields, reply.AlteredFields) {
		t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(eRply.AlteredFields), utils.ToJSON(reply.AlteredFields))
	}
	if !reflect.DeepEqual(eRply.CGREvent.Event, reply.CGREvent.Event) {
		t.Errorf("Expecting %+v, received: %+v", eRply.CGREvent.Event, reply.CGREvent.Event)
	}
}

func TestAttributeMultipleProcessWithBlocker2(t *testing.T) {
	//refresh the DM
	if err := dmAtr.DataDB()[utils.MetaDefault].Flush(""); err != nil {
		t.Error(err)
	}
	engine.Cache.Clear(nil)
	if test, err := dmAtr.DataDB()[utils.MetaDefault].IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("Expecting: true got :%+v", test)
	}
	attrPrf1 := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		FilterIDs: []string{"*string:~*req.InitialField:InitialValue", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field1",
				Value: utils.NewRSRParsersMustCompile("Value1", utils.InfieldSep),
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: true,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	attrPrf2 := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_2",
		FilterIDs: []string{"*string:~*req.Field1:Value1", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Value: utils.NewRSRParsersMustCompile("Value2", utils.InfieldSep),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	attrPrf3 := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_3",
		FilterIDs: []string{"*string:~*req.Field2:Value2", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field3",
				Value: utils.NewRSRParsersMustCompile("Value3", utils.InfieldSep),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf1, true); err != nil {
		t.Error(err)
	}
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf2, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf3, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"InitialField": "InitialValue",
		},
		APIOpts: map[string]any{
			utils.OptsContext:               utils.MetaSessionS,
			utils.OptsAttributesProcessRuns: 4,
		},
	}
	eRply := &AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_1",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field1"},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"InitialField": "InitialValue",
				"Field1":       "Value1",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eRply.AlteredFields, reply.AlteredFields) {
		t.Errorf("Expecting %+v, received: %+v", eRply.AlteredFields, reply.AlteredFields)
	}
	if !reflect.DeepEqual(eRply.CGREvent.Event, reply.CGREvent.Event) {
		t.Errorf("Expecting %+v, received: %+v", eRply.CGREvent.Event, reply.CGREvent.Event)
	}
}

func TestAttributeProcessValue(t *testing.T) {
	//refresh the DM
	if err := dmAtr.DataDB()[utils.MetaDefault].Flush(""); err != nil {
		t.Error(err)
	}
	engine.Cache.Clear(nil)
	if test, err := dmAtr.DataDB()[utils.MetaDefault].IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("Expecting: true got :%+v", test)
	}
	attrPrf1 := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		FilterIDs: []string{"*string:~*req.Field1:Value1", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Value: utils.NewRSRParsersMustCompile("~*req.Field1", utils.InfieldSep),
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: true,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf1, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"Field1": "Value1",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eRply := &AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_1",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field2"},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"Field1": "Value1",
				"Field2": "Value1",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eRply.AlteredFields, reply.AlteredFields) {
		t.Errorf("Expecting %+v, received: %+v", eRply.AlteredFields, reply.AlteredFields)
	}
	if !reflect.DeepEqual(eRply.CGREvent.Event, reply.CGREvent.Event) {
		t.Errorf("Expecting %+v, received: %+v", eRply.CGREvent.Event, reply.CGREvent.Event)
	}
}

func TestAttributeAttributeFilterIDs(t *testing.T) {
	//refresh the DM
	if err := dmAtr.DataDB()[utils.MetaDefault].Flush(""); err != nil {
		t.Error(err)
	}
	engine.Cache.Clear(nil)
	if test, err := dmAtr.DataDB()[utils.MetaDefault].IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("Expecting: true got :%+v", test)
	}
	attrPrf1 := &utils.AttributeProfile{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "ATTR_1",
		Attributes: []*utils.Attribute{
			{
				FilterIDs: []string{"*string:~*req.PassField:Test"},
				Path:      utils.MetaReq + utils.NestingSep + "PassField",
				Value:     utils.NewRSRParsersMustCompile("Pass", utils.InfieldSep),
			},
			{
				FilterIDs: []string{"*string:~*req.PassField:RandomValue"},
				Path:      utils.MetaReq + utils.NestingSep + "NotPassField",
				Value:     utils.NewRSRParsersMustCompile("NotPass", utils.InfieldSep),
			},
			{
				FilterIDs: []string{"*notexists:~*req.RandomField:"},
				Path:      utils.MetaReq + utils.NestingSep + "RandomField",
				Value:     utils.NewRSRParsersMustCompile("RandomValue", utils.InfieldSep),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf1, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"PassField": "Test",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eRply := &AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_1",
				Fields: []string{utils.MetaReq + utils.NestingSep + "PassField",
					utils.MetaReq + utils.NestingSep + "RandomField"},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"PassField":   "Pass",
				"RandomField": "RandomValue",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eRply.AlteredFields, reply.AlteredFields) {
		t.Errorf("Expecting %+v, received: %+v", eRply.AlteredFields, reply.AlteredFields)
	}
	if !reflect.DeepEqual(eRply.CGREvent.Event, reply.CGREvent.Event) {
		t.Errorf("Expecting %+v, received: %+v", eRply.CGREvent.Event, reply.CGREvent.Event)
	}
}

func TestAttributeProcessEventConstant(t *testing.T) {
	//refresh the DM
	if err := dmAtr.DataDB()[utils.MetaDefault].Flush(""); err != nil {
		t.Error(err)
	}
	engine.Cache.Clear(nil)
	if test, err := dmAtr.DataDB()[utils.MetaDefault].IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("Expecting: true got :%+v", test)
	}
	attrPrf1 := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		FilterIDs: []string{"*string:~*req.Field1:Value1", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaConstant,
				Value: utils.NewRSRParsersMustCompile("ConstVal", utils.InfieldSep),
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: true,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf1, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"Field1": "Value1",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eRply := &AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_1",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field2"},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"Field1": "Value1",
				"Field2": "ConstVal",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eRply.AlteredFields, reply.AlteredFields) {
		t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(eRply.AlteredFields), utils.ToJSON(reply.AlteredFields))
	}
	if !reflect.DeepEqual(eRply.CGREvent.Event, reply.CGREvent.Event) {
		t.Errorf("Expecting %+v, received: %+v", eRply.CGREvent.Event, reply.CGREvent.Event)
	}
}

func TestAttributeProcessEventVariable(t *testing.T) {
	//refresh the DM
	if err := dmAtr.DataDB()[utils.MetaDefault].Flush(""); err != nil {
		t.Error(err)
	}
	engine.Cache.Clear(nil)
	if test, err := dmAtr.DataDB()[utils.MetaDefault].IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("Expecting: true got :%+v", test)
	}
	attrPrf1 := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		FilterIDs: []string{"*string:~*req.Field1:Value1", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaVariable,
				Value: utils.NewRSRParsersMustCompile("~*req.Field1", utils.InfieldSep),
			},
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaVariable,
				Value: utils.NewRSRParsersMustCompile("~*req.TheField", utils.InfieldSep),
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: true,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf1, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"Field1":   "Value1",
			"TheField": "TheVal",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eRply := &AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_1",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field2"},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"Field1":   "Value1",
				"Field2":   "TheVal",
				"TheField": "TheVal",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eRply.AlteredFields, reply.AlteredFields) {
		t.Errorf("Expecting %+v, received: %+v", eRply.AlteredFields, reply.AlteredFields)
	}
	if !reflect.DeepEqual(eRply.CGREvent.Event, reply.CGREvent.Event) {
		t.Errorf("Expecting %+v, received: %+v", eRply.CGREvent.Event, reply.CGREvent.Event)
	}
}

func TestAttributeProcessEventComposed(t *testing.T) {
	//refresh the DM
	if err := dmAtr.DataDB()[utils.MetaDefault].Flush(""); err != nil {
		t.Error(err)
	}
	engine.Cache.Clear(nil)
	if test, err := dmAtr.DataDB()[utils.MetaDefault].IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("Expecting: true got :%+v", test)
	}
	attrPrf1 := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		FilterIDs: []string{"*string:~*req.Field1:Value1", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaComposed,
				Value: utils.NewRSRParsersMustCompile("~*req.Field1", utils.InfieldSep),
			},
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaComposed,
				Value: utils.NewRSRParsersMustCompile("_", utils.InfieldSep),
			},
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaComposed,
				Value: utils.NewRSRParsersMustCompile("~*req.TheField", utils.InfieldSep),
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: true,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf1, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"Field1":   "Value1",
			"TheField": "TheVal",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eRply := &AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_1",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field2"},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"Field1":   "Value1",
				"Field2":   "Value1_TheVal",
				"TheField": "TheVal",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
		t.Fatalf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eRply.AlteredFields, reply.AlteredFields) {
		t.Errorf("Expecting %+v, received: %+v", eRply.AlteredFields, reply.AlteredFields)
	}
	if !reflect.DeepEqual(eRply.CGREvent.Event, reply.CGREvent.Event) {
		t.Errorf("Expecting %+v, received: %+v", eRply.CGREvent.Event, reply.CGREvent.Event)
	}
}

func TestAttributeProcessEventSum(t *testing.T) {
	//refresh the DM
	if err := dmAtr.DataDB()[utils.MetaDefault].Flush(""); err != nil {
		t.Error(err)
	}
	engine.Cache.Clear(nil)
	if test, err := dmAtr.DataDB()[utils.MetaDefault].IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("Expecting: true got :%+v", test)
	}
	attrPrf1 := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		FilterIDs: []string{"*string:~*req.Field1:Value1", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaSum,
				Value: utils.NewRSRParsersMustCompile("10;~*req.NumField;20", utils.InfieldSep),
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: true,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf1, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"Field1":   "Value1",
			"TheField": "TheVal",
			"NumField": "20",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eRply := &AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_1",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field2"},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"Field1":   "Value1",
				"TheField": "TheVal",
				"NumField": "20",
				"Field2":   "50",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eRply.AlteredFields, reply.AlteredFields) {
		t.Errorf("Expecting %+v, received: %+v", eRply.AlteredFields, reply.AlteredFields)
	}
	if !reflect.DeepEqual(eRply.CGREvent.Event, reply.CGREvent.Event) {
		t.Errorf("Expecting %+v, received: %+v", eRply.CGREvent.Event, reply.CGREvent.Event)
	}
}

func TestAttributeProcessEventUsageDifference(t *testing.T) {
	//refresh the DM
	if err := dmAtr.DataDB()[utils.MetaDefault].Flush(""); err != nil {
		t.Error(err)
	}
	engine.Cache.Clear(nil)
	if test, err := dmAtr.DataDB()[utils.MetaDefault].IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("Expecting: true got :%+v", test)
	}
	attrPrf1 := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		FilterIDs: []string{"*string:~*req.Field1:Value1", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaUsageDifference,
				Value: utils.NewRSRParsersMustCompile("~*req.UnixTimeStamp;~*req.UnixTimeStamp2", utils.InfieldSep),
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: true,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf1, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"Field1":         "Value1",
			"TheField":       "TheVal",
			"UnixTimeStamp":  "1554364297",
			"UnixTimeStamp2": "1554364287",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eRply := &AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_1",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field2"},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"Field1":         "Value1",
				"TheField":       "TheVal",
				"UnixTimeStamp":  "1554364297",
				"UnixTimeStamp2": "1554364287",
				"Field2":         "10s",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eRply.AlteredFields, reply.AlteredFields) {
		t.Errorf("Expecting %+v, received: %+v", eRply.AlteredFields, reply.AlteredFields)
	}
	if !reflect.DeepEqual(eRply.CGREvent.Event, reply.CGREvent.Event) {
		t.Errorf("Expecting %+v, received: %+v", eRply.CGREvent.Event, reply.CGREvent.Event)
	}
}

func TestAttributeProcessEventValueExponent(t *testing.T) {
	//refresh the DM
	if err := dmAtr.DataDB()[utils.MetaDefault].Flush(""); err != nil {
		t.Error(err)
	}
	engine.Cache.Clear(nil)
	if test, err := dmAtr.DataDB()[utils.MetaDefault].IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("Expecting: true got :%+v", test)
	}
	attrPrf1 := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		FilterIDs: []string{"*string:~*req.Field1:Value1", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaValueExponent,
				Value: utils.NewRSRParsersMustCompile("~*req.Multiplier;~*req.Pow", utils.InfieldSep),
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: true,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf1, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"Field1":     "Value1",
			"TheField":   "TheVal",
			"Multiplier": "2",
			"Pow":        "3",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eRply := &AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_1",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field2"},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"Field1":     "Value1",
				"TheField":   "TheVal",
				"Multiplier": "2",
				"Pow":        "3",
				"Field2":     "2000",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eRply.AlteredFields, reply.AlteredFields) {
		t.Errorf("Expecting %+v, received: %+v", eRply.AlteredFields, reply.AlteredFields)
	}
	if !reflect.DeepEqual(eRply.CGREvent.Event, reply.CGREvent.Event) {
		t.Errorf("Expecting %+v, received: %+v", eRply.CGREvent.Event, reply.CGREvent.Event)
	}
}

func BenchmarkAttributeProcessEventConstant(b *testing.B) {
	cfg := config.NewDefaultCGRConfig()
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		b.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dmAtr = engine.NewDataManager(dbCM, cfg, nil)
	fltrs := engine.NewFilterS(cfg, nil, dmAtr)
	attrS = NewAttributeService(dmAtr, fltrs, nil, cfg)
	//refresh the DM
	if err := dmAtr.DataDB()[utils.MetaDefault].Flush(""); err != nil {
		b.Error(err)
	}
	engine.Cache.Clear(nil)
	if test, err := dmAtr.DataDB()[utils.MetaDefault].IsDBEmpty(); err != nil {
		b.Error(err)
	} else if test != true {
		b.Errorf("Expecting: true got :%+v", test)
	}
	attrPrf1 := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		FilterIDs: []string{"*string:~*req.Field1:Value1", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaConstant,
				Value: utils.NewRSRParsersMustCompile("ConstVal", utils.InfieldSep),
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: true,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf1, true); err != nil {
		b.Error(err)
	}
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"Field1": "Value1",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	var reply AttrSProcessEventReply
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
			b.Errorf("Error: %+v", err)
		}
	}
}

func BenchmarkAttributeProcessEventVariable(b *testing.B) {
	cfg := config.NewDefaultCGRConfig()
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		b.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dmAtr = engine.NewDataManager(dbCM, cfg, nil)
	fltrs := engine.NewFilterS(cfg, nil, dmAtr)
	attrS = NewAttributeService(dmAtr, fltrs, nil, cfg)

	//refresh the DM
	if err := dmAtr.DataDB()[utils.MetaDefault].Flush(""); err != nil {
		b.Error(err)
	}
	engine.Cache.Clear(nil)
	if test, err := dmAtr.DataDB()[utils.MetaDefault].IsDBEmpty(); err != nil {
		b.Error(err)
	} else if test != true {
		b.Errorf("Expecting: true got :%+v", test)
	}
	attrPrf1 := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		FilterIDs: []string{"*string:~*req.Field1:Value1", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaVariable,
				Value: utils.NewRSRParsersMustCompile("~*req.Field1", utils.InfieldSep),
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: true,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf1, true); err != nil {
		b.Error(err)
	}
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"Field1": "Value1",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	var reply AttrSProcessEventReply
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
			b.Errorf("Error: %+v", err)
		}
	}
}

func TestGetAttributeProfileFromInline(t *testing.T) {
	//refresh the DM
	if err := dmAtr.DataDB()[utils.MetaDefault].Flush(""); err != nil {
		t.Error(err)
	}
	engine.Cache.Clear(nil)
	if test, err := dmAtr.DataDB()[utils.MetaDefault].IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("Expecting: true got :%+v", test)
	}
	attrID := "*sum:*req.Field2:10&~*req.NumField&20"
	expAttrPrf1 := &utils.AttributeProfile{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     attrID,
		Attributes: []*utils.Attribute{{
			Path:  utils.MetaReq + utils.NestingSep + "Field2",
			Type:  utils.MetaSum,
			Value: utils.NewRSRParsersMustCompile("10;~*req.NumField;20", utils.InfieldSep),
		}},
	}
	cfg := config.NewDefaultCGRConfig()
	attr, err := engine.NewDataManager(&engine.DBConnManager{}, cfg, nil).GetAttributeProfile(context.TODO(), config.CgrConfig().GeneralCfg().DefaultTenant, attrID, false, false, "")
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expAttrPrf1, attr) {
		t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(expAttrPrf1), utils.ToJSON(attr))
	}
}

func TestProcessAttributeConstant(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dmAtr = engine.NewDataManager(dbCM, cfg, nil)
	fltrs := engine.NewFilterS(cfg, nil, dmAtr)
	attrS = NewAttributeService(dmAtr, fltrs, nil, cfg)
	attrPrf := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_CONSTANT",
		FilterIDs: []string{"*string:~*req.Field1:Val1", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaConstant,
				Value: utils.NewRSRParsersMustCompile("Val2", utils.InfieldSep),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{ //matching ATTR_CONSTANT
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "TestProcessAttributeConstant",
		Event: map[string]any{
			"Field1":     "Val1",
			utils.Weight: "20.0",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eNM := utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}
	rcv, err := attrS.processEvent(context.TODO(), ev.Tenant, ev, eNM, engine.NewDynamicDP(context.TODO(), nil, nil, nil, nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	ev.Event["Field2"] = "Val2"
	eRply := &AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_CONSTANT",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field2"},
			},
		},
		CGREvent: ev,
	}
	if !reflect.DeepEqual(eRply, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eRply), utils.ToJSON(rcv))
	}
}

func TestProcessAttributeVariable(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dmAtr = engine.NewDataManager(dbCM, cfg, nil)
	engine.Cache.Clear(nil)
	fltrs := engine.NewFilterS(cfg, nil, dmAtr)
	attrS = NewAttributeService(dmAtr, fltrs, nil, cfg)
	attrPrf := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_VARIABLE",
		FilterIDs: []string{"*string:~*req.Field1:Val1", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaVariable,
				Value: utils.NewRSRParsersMustCompile("~*req.RandomField", utils.InfieldSep),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{ //matching ATTR_VARIABLE
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "TestProcessAttributeVariable",
		Event: map[string]any{
			"Field1":      "Val1",
			"RandomField": "Val2",
			utils.Weight:  "20.0",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eNM := utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}
	rcv, err := attrS.processEvent(context.TODO(), ev.Tenant, ev, eNM, engine.NewDynamicDP(context.TODO(), nil, nil, nil, nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	clnEv := ev.Clone()
	clnEv.Event["Field2"] = "Val2"
	eRply := &AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_VARIABLE",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field2"},
			},
		},
		CGREvent: clnEv,
	}
	if !reflect.DeepEqual(eRply, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eRply), utils.ToJSON(rcv))
	}
}

func TestProcessAttributeComposed(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dmAtr = engine.NewDataManager(dbCM, cfg, nil)
	engine.Cache.Clear(nil)
	fltrs := engine.NewFilterS(cfg, nil, dmAtr)
	attrS = NewAttributeService(dmAtr, fltrs, nil, cfg)
	attrPrf := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_COMPOSED",
		FilterIDs: []string{"*string:~*req.Field1:Val1", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaComposed,
				Value: utils.NewRSRParsersMustCompile("~*req.RandomField", utils.InfieldSep),
			},
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaComposed,
				Value: utils.NewRSRParsersMustCompile("~*req.RandomField2", utils.InfieldSep),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{ //matching ATTR_COMPOSED
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "TestProcessAttributeComposed",
		Event: map[string]any{
			"Field1":       "Val1",
			"RandomField":  "Val2",
			"RandomField2": "Concatenated",
			utils.Weight:   "20.0",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eNM := utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}
	rcv, err := attrS.processEvent(context.TODO(), ev.Tenant, ev, eNM, engine.NewDynamicDP(context.TODO(), nil, nil, nil, nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	clnEv := ev.Clone()
	clnEv.Event["Field2"] = "Val2Concatenated"
	eRply := &AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_COMPOSED",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field2"},
			},
		},
		CGREvent: clnEv,
	}
	if !reflect.DeepEqual(eRply, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eRply), utils.ToJSON(rcv))
	}
}

func TestProcessAttributeUsageDifference(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dmAtr = engine.NewDataManager(dbCM, cfg, nil)
	engine.Cache.Clear(nil)
	fltrs := engine.NewFilterS(cfg, nil, dmAtr)
	attrS = NewAttributeService(dmAtr, fltrs, nil, cfg)
	attrPrf := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_USAGE_DIFF",
		FilterIDs: []string{"*string:~*req.Field1:Val1", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaUsageDifference,
				Value: utils.NewRSRParsersMustCompile("~*req.RandomField;~*req.RandomField2", utils.InfieldSep),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{ //matching ATTR_USAGE_DIFF
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "TestProcessAttributeUsageDifference",
		Event: map[string]any{
			"Field1":       "Val1",
			"RandomField":  "1514808000",
			"RandomField2": "1514804400",
			utils.Weight:   "20.0",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eNM := utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}
	rcv, err := attrS.processEvent(context.TODO(), ev.Tenant, ev, eNM, engine.NewDynamicDP(context.TODO(), nil, nil, nil, nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	clnEv := ev.Clone()
	clnEv.Event["Field2"] = "1h0m0s"
	eRply := &AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_USAGE_DIFF",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field2"},
			},
		},
		CGREvent: clnEv,
	}
	if !reflect.DeepEqual(eRply, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eRply), utils.ToJSON(rcv))
	}
}

func TestProcessAttributeSum(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dmAtr = engine.NewDataManager(dbCM, cfg, nil)
	engine.Cache.Clear(nil)
	fltrs := engine.NewFilterS(cfg, nil, dmAtr)
	attrS = NewAttributeService(dmAtr, fltrs, nil, cfg)
	attrPrf := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_SUM",
		FilterIDs: []string{"*string:~*req.Field1:Val1", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaSum,
				Value: utils.NewRSRParsersMustCompile("~*req.RandomField;~*req.RandomField2;10", utils.InfieldSep),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{ //matching ATTR_SUM
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "TestProcessAttributeSum",
		Event: map[string]any{
			"Field1":       "Val1",
			"RandomField":  "1",
			"RandomField2": "5",
			utils.Weight:   "20.0",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eNM := utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}
	rcv, err := attrS.processEvent(context.TODO(), ev.Tenant, ev, eNM, engine.NewDynamicDP(context.TODO(), nil, nil, nil, nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	clnEv := ev.Clone()
	clnEv.Event["Field2"] = "16"
	eRply := &AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_SUM",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field2"},
			},
		},
		CGREvent: clnEv,
	}
	if !reflect.DeepEqual(eRply, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eRply), utils.ToJSON(rcv))
	}
}

func TestProcessAttributeDiff(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dmAtr = engine.NewDataManager(dbCM, cfg, nil)
	engine.Cache.Clear(nil)
	fltrs := engine.NewFilterS(cfg, nil, dmAtr)
	attrS = NewAttributeService(dmAtr, fltrs, nil, cfg)
	attrPrf := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_DIFF",
		FilterIDs: []string{"*string:~*req.Field1:Val1", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaDifference,
				Value: utils.NewRSRParsersMustCompile("55;~*req.RandomField;~*req.RandomField2;10", utils.InfieldSep),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{ //matching ATTR_DIFF
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "TestProcessAttributeDiff",
		Event: map[string]any{
			"Field1":       "Val1",
			"RandomField":  "1",
			"RandomField2": "5",
			utils.Weight:   "20.0",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eNM := utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}
	rcv, err := attrS.processEvent(context.TODO(), ev.Tenant, ev, eNM, engine.NewDynamicDP(context.TODO(), nil, nil, nil, nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	clnEv := ev.Clone()
	clnEv.Event["Field2"] = "39"
	eRply := &AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_DIFF",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field2"},
			},
		},
		CGREvent: clnEv,
	}
	if !reflect.DeepEqual(eRply, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eRply), utils.ToJSON(rcv))
	}
}

func TestProcessAttributeMultiply(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dmAtr = engine.NewDataManager(dbCM, cfg, nil)
	engine.Cache.Clear(nil)
	fltrs := engine.NewFilterS(cfg, nil, dmAtr)
	attrS = NewAttributeService(dmAtr, fltrs, nil, cfg)
	attrPrf := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_MULTIPLY",
		FilterIDs: []string{"*string:~*req.Field1:Val1", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaMultiply,
				Value: utils.NewRSRParsersMustCompile("55;~*req.RandomField;~*req.RandomField2;10", utils.InfieldSep),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{ //matching ATTR_MULTIPLY
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "TestProcessAttributeMultiply",
		Event: map[string]any{
			"Field1":       "Val1",
			"RandomField":  "1",
			"RandomField2": "5",
			utils.Weight:   "20.0",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eNM := utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}
	rcv, err := attrS.processEvent(context.TODO(), ev.Tenant, ev, eNM, engine.NewDynamicDP(context.TODO(), nil, nil, nil, nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	clnEv := ev.Clone()
	clnEv.Event["Field2"] = "2750"
	eRply := &AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_MULTIPLY",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field2"},
			},
		},
		CGREvent: clnEv,
	}
	if !reflect.DeepEqual(eRply, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eRply), utils.ToJSON(rcv))
	}
}

func TestProcessAttributeDivide(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dmAtr = engine.NewDataManager(dbCM, cfg, nil)
	engine.Cache.Clear(nil)
	fltrs := engine.NewFilterS(cfg, nil, dmAtr)
	attrS = NewAttributeService(dmAtr, fltrs, nil, cfg)
	attrPrf := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_DIVIDE",
		FilterIDs: []string{"*string:~*req.Field1:Val1", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaDivide,
				Value: utils.NewRSRParsersMustCompile("55.0;~*req.RandomField;~*req.RandomField2;4", utils.InfieldSep),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{ //matching ATTR_DIVIDE
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "TestProcessAttributeDivide",
		Event: map[string]any{
			"Field1":       "Val1",
			"RandomField":  "1",
			"RandomField2": "5",
			utils.Weight:   "20.0",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eNM := utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}
	rcv, err := attrS.processEvent(context.TODO(), ev.Tenant, ev, eNM, engine.NewDynamicDP(context.TODO(), nil, nil, nil, nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	clnEv := ev.Clone()
	clnEv.Event["Field2"] = "2.75"
	eRply := &AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_DIVIDE",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field2"},
			},
		},
		CGREvent: clnEv,
	}
	if !reflect.DeepEqual(eRply, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eRply), utils.ToJSON(rcv))
	}
}

func TestProcessAttributeValueExponent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dmAtr = engine.NewDataManager(dbCM, cfg, nil)
	engine.Cache.Clear(nil)
	fltrs := engine.NewFilterS(cfg, nil, dmAtr)
	attrS = NewAttributeService(dmAtr, fltrs, nil, cfg)
	attrPrf := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_VAL_EXP",
		FilterIDs: []string{"*string:~*req.Field1:Val1", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaValueExponent,
				Value: utils.NewRSRParsersMustCompile("~*req.RandomField2;4", utils.InfieldSep),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{ //matching ATTR_VAL_EXP
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "TestProcessAttributeValueExponent",
		Event: map[string]any{
			"Field1":       "Val1",
			"RandomField":  "1",
			"RandomField2": "5",
			utils.Weight:   "20.0",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eNM := utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}
	rcv, err := attrS.processEvent(context.TODO(), ev.Tenant, ev, eNM, engine.NewDynamicDP(context.TODO(), nil, nil, nil, nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	clnEv := ev.Clone()
	clnEv.Event["Field2"] = "50000"
	eRply := &AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_VAL_EXP",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field2"},
			},
		},
		CGREvent: clnEv,
	}
	if !reflect.DeepEqual(eRply, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eRply), utils.ToJSON(rcv))
	}
}

func TestProcessAttributeUnixTimeStamp(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dmAtr = engine.NewDataManager(dbCM, cfg, nil)
	engine.Cache.Clear(nil)
	fltrs := engine.NewFilterS(cfg, nil, dmAtr)
	attrS = NewAttributeService(dmAtr, fltrs, nil, cfg)
	attrPrf := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_UNIX_TIMESTAMP",
		FilterIDs: []string{"*string:~*req.Field1:Val1", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaUnixTimestamp,
				Value: utils.NewRSRParsersMustCompile("~*req.RandomField2", utils.InfieldSep),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{ //matching ATTR_UNIX_TIMESTAMP
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "TestProcessAttributeUnixTimeStamp",
		Event: map[string]any{
			"Field1":       "Val1",
			"RandomField":  "1",
			"RandomField2": "2013-12-30T15:00:01Z",
			utils.Weight:   "20.0",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eNM := utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}
	rcv, err := attrS.processEvent(context.TODO(), ev.Tenant, ev, eNM, engine.NewDynamicDP(context.TODO(), nil, nil, nil, nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	clnEv := ev.Clone()
	clnEv.Event["Field2"] = "1388415601"
	eRply := &AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_UNIX_TIMESTAMP",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field2"},
			},
		},
		CGREvent: clnEv,
	}
	if !reflect.DeepEqual(eRply, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eRply), utils.ToJSON(rcv))
	}
}

func TestProcessAttributePrefix(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dmAtr = engine.NewDataManager(dbCM, cfg, nil)
	engine.Cache.Clear(nil)
	fltrs := engine.NewFilterS(cfg, nil, dmAtr)
	attrS = NewAttributeService(dmAtr, fltrs, nil, cfg)
	attrPrf := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_PREFIX",
		FilterIDs: []string{"*string:~*req.ATTR:ATTR_PREFIX", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaPrefix,
				Value: utils.NewRSRParsersMustCompile("abc_", utils.InfieldSep),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{ //matching ATTR_VAL_EXP
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "TestProcessAttributeValueExponent",
		Event: map[string]any{
			"ATTR":       "ATTR_PREFIX",
			"Field2":     "Val2",
			utils.Weight: "20.0",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eNM := utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}
	rcv, err := attrS.processEvent(context.TODO(), ev.Tenant, ev, eNM, engine.NewDynamicDP(context.TODO(), nil, nil, nil, nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	clnEv := ev.Clone()
	clnEv.Event["Field2"] = "abc_Val2"
	eRply := &AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_PREFIX",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field2"},
			},
		},
		CGREvent: clnEv,
	}
	if !reflect.DeepEqual(eRply, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eRply), utils.ToJSON(rcv))
	}
}

func TestProcessAttributeSuffix(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dmAtr = engine.NewDataManager(dbCM, cfg, nil)
	engine.Cache.Clear(nil)
	fltrs := engine.NewFilterS(cfg, nil, dmAtr)
	attrS = NewAttributeService(dmAtr, fltrs, nil, cfg)
	attrPrf := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_SUFFIX",
		FilterIDs: []string{"*string:~*req.ATTR:ATTR_SUFFIX", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaSuffix,
				Value: utils.NewRSRParsersMustCompile("_abc", utils.InfieldSep),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{ //matching ATTR_VAL_EXP
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "TestProcessAttributeValueExponent",
		Event: map[string]any{
			"ATTR":       "ATTR_SUFFIX",
			"Field2":     "Val2",
			utils.Weight: "20.0",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eNM := utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}
	rcv, err := attrS.processEvent(context.TODO(), ev.Tenant, ev, eNM, engine.NewDynamicDP(context.TODO(), nil, nil, nil, nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	clnEv := ev.Clone()
	clnEv.Event["Field2"] = "Val2_abc"
	eRply := &AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_SUFFIX",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field2"},
			},
		},
		CGREvent: clnEv,
	}
	if !reflect.DeepEqual(eRply, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eRply), utils.ToJSON(rcv))
	}
}

func TestAttributeIndexSelectsFalse(t *testing.T) {
	// change the IndexedSelects to false
	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().StringIndexedFields = nil
	cfg.AttributeSCfg().PrefixIndexedFields = nil
	cfg.AttributeSCfg().IndexedSelects = false
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dmAtr = engine.NewDataManager(dbCM, cfg, nil)
	fltrs := engine.NewFilterS(cfg, nil, dmAtr)
	attrS = NewAttributeService(dmAtr, fltrs, nil, cfg)

	//refresh the DM
	if err := dmAtr.DataDB()[utils.MetaDefault].Flush(""); err != nil {
		t.Error(err)
	}
	if test, err := dmAtr.DataDB()[utils.MetaDefault].IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("Expecting: true got :%+v", test)
	}
	expTimeStr := expTimeAttributes.Format("2006-01-02T15:04:05Z")
	attrPrf := &utils.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "AttrPrf",
		FilterIDs: []string{"*string:~*req.Account:1007", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z|" + expTimeStr, "*string:~*opts.*context:*cdrs"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + utils.AccountField,
				Value: utils.NewRSRParsersMustCompile("1010", utils.InfieldSep),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf, true); err != nil {
		t.Error(err)
	}

	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"Account": "1007",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}

	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected not found, reveiced: %+v", err)
	}

}

func TestProcessAttributeWithSameWeight(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dmAtr = engine.NewDataManager(dbCM, cfg, nil)
	engine.Cache.Clear(nil)
	fltrs := engine.NewFilterS(cfg, nil, dmAtr)
	attrS = NewAttributeService(dmAtr, fltrs, nil, cfg)
	attrPrf := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		FilterIDs: []string{"*string:~*req.Field1:Val1", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaVariable,
				Value: utils.NewRSRParsersMustCompile("~*req.RandomField", utils.InfieldSep),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	attrPrf2 := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_2",
		FilterIDs: []string{"*string:~*req.Field1:Val1", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field3",
				Type:  utils.MetaVariable,
				Value: utils.NewRSRParsersMustCompile("~*req.RandomField", utils.InfieldSep),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf, true); err != nil {
		t.Error(err)
	}
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf2, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{ //matching ATTR_UNIX_TIMESTAMP
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "TestProcessAttributeUnixTimeStamp",
		Event: map[string]any{
			"Field1":      "Val1",
			"RandomField": "1",
			utils.Weight:  "20.0",
		},
		APIOpts: map[string]any{
			utils.OptsContext:               utils.MetaSessionS,
			utils.OptsAttributesProcessRuns: 2,
		},
	}
	var rcv AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.TODO(), ev, &rcv); err != nil {
		t.Errorf("Error: %+v", err)
	}
	clnEv := ev.Clone()
	clnEv.Event["Field2"] = "1"
	clnEv.Event["Field3"] = "1"
	eRply := AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_1",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field2"},
			},
			{
				MatchedProfileID: "cgrates.org:ATTR_2",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field3"},
			},
		},
		CGREvent: clnEv,
	}
	sort.Slice(rcv.AlteredFields, func(i, j int) bool {
		return rcv.AlteredFields[i].MatchedProfileID < rcv.AlteredFields[j].MatchedProfileID
	})
	if !reflect.DeepEqual(eRply, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eRply), utils.ToJSON(rcv))
	}
}

func TestAttributeMultipleProcessWithFiltersExists(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().IndexedSelects = false
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dmAtr = engine.NewDataManager(dbCM, cfg, nil)
	engine.Cache.Clear(nil)
	fltrs := engine.NewFilterS(cfg, nil, dmAtr)
	attrS = NewAttributeService(dmAtr, fltrs, nil, cfg)
	attrPrf1Exists := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1_EXISTS",
		FilterIDs: []string{"*exists:~*req.InitialField:", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field1",
				Value: utils.NewRSRParsersMustCompile("Value1", utils.InfieldSep),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	attrPrf2Exists := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_2_EXISTS",
		FilterIDs: []string{"*exists:~*req.Field1:", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Value: utils.NewRSRParsersMustCompile("Value2", utils.InfieldSep),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf1Exists, true); err != nil {
		t.Error(err)
	}
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf2Exists, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	// Add attribute in DM
	if _, err := dmAtr.GetAttributeProfile(context.TODO(), attrPrf1Exists.Tenant, attrPrf1Exists.ID, true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if _, err := dmAtr.GetAttributeProfile(context.TODO(), attrPrf2Exists.Tenant, attrPrf2Exists.ID, true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"InitialField": "InitialValue",
		},
		APIOpts: map[string]any{
			utils.OptsContext:               utils.MetaSessionS,
			utils.OptsAttributesProcessRuns: 4,
		},
	}
	eRply := &AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_1_EXISTS",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field1"},
			},
			{
				MatchedProfileID: "cgrates.org:ATTR_2_EXISTS",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field2"},
			},
			{
				MatchedProfileID: "cgrates.org:ATTR_1_EXISTS",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field1"},
			},
			{
				MatchedProfileID: "cgrates.org:ATTR_2_EXISTS",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field2"},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"InitialField": "InitialValue",
				"Field1":       "Value1",
				"Field2":       "Value2",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eRply.AlteredFields, reply.AlteredFields) {
		t.Errorf("Expecting %+v, received: %+v", eRply.AlteredFields, reply.AlteredFields)
	}
	if !reflect.DeepEqual(eRply.CGREvent.Event, reply.CGREvent.Event) {
		t.Errorf("Expecting %+v, received: %+v", eRply.CGREvent.Event, reply.CGREvent.Event)
	}
}

func TestAttributeMultipleProcessWithFiltersNotEmpty(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().IndexedSelects = false
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dmAtr = engine.NewDataManager(dbCM, cfg, nil)
	engine.Cache.Clear(nil)
	fltrs := engine.NewFilterS(cfg, nil, dmAtr)
	attrS = NewAttributeService(dmAtr, fltrs, nil, cfg)
	attrPrf1NotEmpty := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1_NOTEMPTY",
		FilterIDs: []string{"*notempty:~*req.InitialField:", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field1",
				Value: utils.NewRSRParsersMustCompile("Value1", utils.InfieldSep),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	attrPrf2NotEmpty := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_2_NOTEMPTY",
		FilterIDs: []string{"*notempty:~*req.Field1:", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Value: utils.NewRSRParsersMustCompile("Value2", utils.InfieldSep),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf1NotEmpty, true); err != nil {
		t.Error(err)
	}
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf2NotEmpty, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	// Add attribute in DM
	if _, err := dmAtr.GetAttributeProfile(context.TODO(), attrPrf1NotEmpty.Tenant, attrPrf1NotEmpty.ID, true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if _, err := dmAtr.GetAttributeProfile(context.TODO(), attrPrf2NotEmpty.Tenant, attrPrf2NotEmpty.ID, true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"InitialField": "InitialValue",
		},
		APIOpts: map[string]any{
			utils.OptsContext:               utils.MetaSessionS,
			utils.OptsAttributesProcessRuns: 4,
		},
	}
	eRply := &AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_1_NOTEMPTY",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field1"},
			},
			{
				MatchedProfileID: "cgrates.org:ATTR_2_NOTEMPTY",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field2"},
			},
			{
				MatchedProfileID: "cgrates.org:ATTR_1_NOTEMPTY",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field1"},
			},
			{
				MatchedProfileID: "cgrates.org:ATTR_2_NOTEMPTY",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field2"},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"InitialField": "InitialValue",
				"Field1":       "Value1",
				"Field2":       "Value2",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eRply.AlteredFields, reply.AlteredFields) {
		t.Errorf("Expecting %+v, received: %+v", eRply.AlteredFields, reply.AlteredFields)
	}
	if !reflect.DeepEqual(eRply.CGREvent.Event, reply.CGREvent.Event) {
		t.Errorf("Expecting %+v, received: %+v", eRply.CGREvent.Event, reply.CGREvent.Event)
	}
}

func TestAttributeMetaTenant(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().IndexedSelects = false
	idb, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: idb}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	engine.Cache.Clear(nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	attrS = NewAttributeService(dm, fltrs, nil, cfg)
	attr1 := &utils.AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_TNT",
		FilterIDs: []string{"*string:~*opts.*context:*sessions"},
		Attributes: []*utils.Attribute{{
			Type:  utils.MetaPrefix,
			Path:  utils.MetaTenant,
			Value: utils.NewRSRParsersMustCompile("prfx_", utils.InfieldSep),
		}},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}

	// Add attribute in DM
	if err := dm.SetAttributeProfile(context.TODO(), attr1, true); err != nil {
		t.Error(err)
	}
	if _, err := dm.GetAttributeProfile(context.TODO(), attr1.Tenant, attr1.ID, true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		Event:  map[string]any{},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eRply := AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_TNT",
				Fields:           []string{utils.MetaTenant},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: "prfx_" + config.CgrConfig().GeneralCfg().DefaultTenant,
			Event:  map[string]any{},
			APIOpts: map[string]any{
				utils.OptsContext: utils.MetaSessionS,
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(eRply, reply) {
		t.Errorf("Expecting %s, received: %s", utils.ToJSON(eRply), utils.ToJSON(reply))
	}
}

func TestAttributesPorcessEventMatchingProcessRuns(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().Enabled = true
	cfg.AttributeSCfg().IndexedSelects = false
	db, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: db}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	fltrS := engine.NewFilterS(cfg, nil, dm)
	fltr := &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "Process_Runs_Fltr",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaGreaterThan,
				Element: "~*vars.*processRuns",
				Values:  []string{"1"},
			},
		},
	}
	if err := dm.SetFilter(context.Background(), fltr, true); err != nil {
		t.Error(err)
	}

	attrPfr := &utils.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_ProcessRuns",
		FilterIDs: []string{"Process_Runs_Fltr"},
		Attributes: []*utils.Attribute{
			{
				Path:  "*req.CompanyName",
				Type:  utils.MetaVariable,
				Value: utils.NewRSRParsersMustCompile("ITSYS COMMUNICATIONS SRL", utils.InfieldSep),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	// this I'll match first, no fltr and processRuns will be 1
	attrPfr2 := &utils.AttributeProfile{
		Tenant: "cgrates.org",
		ID:     "ATTR_MatchSecond",
		Attributes: []*utils.Attribute{
			{
				Path:  "*req.Password",
				Type:  utils.MetaVariable,
				Value: utils.NewRSRParsersMustCompile("CGRateS.org", utils.InfieldSep),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}

	attrPfr.Compile()
	fltr.Compile()
	attrPfr2.Compile()
	if err := dm.SetAttributeProfile(context.Background(), attrPfr, true); err != nil {
		t.Error(err)
	}
	if err := dm.SetAttributeProfile(context.Background(), attrPfr2, true); err != nil {
		t.Error(err)
	}

	attr := NewAttributeService(dm, fltrS, nil, cfg)

	ev := &utils.CGREvent{
		Event: map[string]any{
			"Account":     "pc_test",
			"CompanyName": "MY_company_will_be_changed",
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProcessRuns: 2,
		},
	}
	reply := &AttrSProcessEventReply{}
	expReply := &AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_MatchSecond",
				Fields:           []string{"*req.Password"},
			},
			{
				MatchedProfileID: "cgrates.org:ATTR_ProcessRuns",
				Fields:           []string{"*req.CompanyName"},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]any{
				"Account":     "pc_test",
				"CompanyName": "ITSYS COMMUNICATIONS SRL",
				"Password":    "CGRateS.org",
			},
			APIOpts: map[string]any{
				utils.OptsAttributesProcessRuns: 2,
			},
		},
	}
	if err := attr.V1ProcessEvent(context.Background(), ev, reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expReply, reply) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expReply), utils.ToJSON(reply))
	}
}

func TestAttributeMultipleProfileRunns(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().IndexedSelects = false
	idb, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: idb}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	engine.Cache.Clear(nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	attrS = NewAttributeService(dm, fltrs, nil, cfg)
	attrPrf1Exists := &utils.AttributeProfile{
		Tenant:    cfg.GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		FilterIDs: []string{},
		Attributes: []*utils.Attribute{{
			Path:  utils.MetaReq + utils.NestingSep + "Field1",
			Value: utils.NewRSRParsersMustCompile("Value1", utils.InfieldSep),
		}},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	attrPrf2Exists := &utils.AttributeProfile{
		Tenant:    cfg.GeneralCfg().DefaultTenant,
		ID:        "ATTR_2",
		FilterIDs: []string{},
		Attributes: []*utils.Attribute{{
			Path:  utils.MetaReq + utils.NestingSep + "Field2",
			Value: utils.NewRSRParsersMustCompile("Value2", utils.InfieldSep),
		}},
		Weights: utils.DynamicWeights{
			{
				Weight: 5,
			},
		},
	}
	// Add attribute in DM
	if err := dm.SetAttributeProfile(context.TODO(), attrPrf1Exists, true); err != nil {
		t.Error(err)
	}
	if err := dm.SetAttributeProfile(context.TODO(), attrPrf2Exists, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	// Add attribute in DM
	if _, err := dm.GetAttributeProfile(context.TODO(), attrPrf1Exists.Tenant, attrPrf1Exists.ID, true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if _, err := dm.GetAttributeProfile(context.TODO(), attrPrf2Exists.Tenant, attrPrf2Exists.ID, true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"InitialField": "InitialValue",
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProfileRuns: 2,
			utils.OptsAttributesProcessRuns: 40,
		},
	}
	eRply := AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_1",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field1"},
			},
			{
				MatchedProfileID: "cgrates.org:ATTR_2",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field2"},
			},
			{
				MatchedProfileID: "cgrates.org:ATTR_1",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field1"},
			},
			{
				MatchedProfileID: "cgrates.org:ATTR_2",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field2"},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: cfg.GeneralCfg().DefaultTenant,
			ID:     ev.ID,
			Event: map[string]any{
				"InitialField": "InitialValue",
				"Field1":       "Value1",
				"Field2":       "Value2",
			},
			APIOpts: map[string]any{
				utils.OptsAttributesProfileRuns: 2,
				utils.OptsAttributesProcessRuns: 40,
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eRply, reply) {
		t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(eRply), utils.ToJSON(reply))
	}

	ev = &utils.CGREvent{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"InitialField": "InitialValue",
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProfileRuns: 1,
			utils.OptsAttributesProcessRuns: 40,
		},
	}
	eRply = AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_1",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field1"},
			},
			{
				MatchedProfileID: "cgrates.org:ATTR_2",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Field2"},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: cfg.GeneralCfg().DefaultTenant,
			ID:     ev.ID,
			Event: map[string]any{
				"InitialField": "InitialValue",
				"Field1":       "Value1",
				"Field2":       "Value2",
			},
			APIOpts: map[string]any{
				utils.OptsAttributesProfileRuns: 1,
				utils.OptsAttributesProcessRuns: 40,
			},
		},
	}
	reply = AttrSProcessEventReply{}
	if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eRply, reply) {
		t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(eRply), utils.ToJSON(reply))
	}
}

func TestAttributesV1ProcessEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.FilterSCfg().ResourceSConns = []string{}
	conMng := engine.NewConnManager(cfg)
	db, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: db}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, conMng)
	filterS := engine.NewFilterS(cfg, conMng, dm)
	attr := &utils.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_CHANGE_TENANT_FROM_USER",
		FilterIDs: []string{"*string:~*req.Account:dan@itsyscom.com|adrian@itsyscom.com"},
		Attributes: []*utils.Attribute{
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*variable",
				Value:     utils.NewRSRParsersMustCompile("~*req.Account:s/(.*)@(.*)/${1}.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*req.Account",
				Type:      "*variable",
				Value:     utils.NewRSRParsersMustCompile("~*req.Account:s/(dan)@(.*)/${1}.${2}/:s/(adrian)@(.*)/andrei.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*composed",
				Value:     utils.NewRSRParsersMustCompile(".co.uk", utils.InfieldSep),
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: false,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	err = dm.SetAttributeProfile(context.Background(), attr, true)
	if err != nil {
		t.Error(err)
	}

	attr2 := &utils.AttributeProfile{
		Tenant: "adrian.itsyscom.com.co.uk",
		ID:     "ATTR_MATCH_TENANT",
		Attributes: []*utils.Attribute{
			{
				FilterIDs: nil,
				Path:      "*req.Password",
				Type:      utils.MetaConstant,
				Value:     utils.NewRSRParsersMustCompile("CGRATES.ORG", utils.InfieldSep),
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: false,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}

	err = dm.SetAttributeProfile(context.Background(), attr2, true)
	if err != nil {
		t.Error(err)
	}

	alS := NewAttributeService(dm, filterS, nil, cfg)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "123",
		Event: map[string]any{
			utils.AccountField: "adrian@itsyscom.com",
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProcessRuns: 2,
		},
	}
	rply := &AttrSProcessEventReply{}
	expected := &AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_CHANGE_TENANT_FROM_USER",
				Fields: []string{utils.MetaReq + utils.NestingSep + "Account",
					"*tenant"},
			},
			{
				MatchedProfileID: "adrian.itsyscom.com.co.uk:ATTR_MATCH_TENANT",
				Fields:           []string{"*req.Password"},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: "adrian.itsyscom.com.co.uk",
			ID:     "123",
			Event: map[string]any{
				utils.AccountField: "andrei.itsyscom.com",
				"Password":         "CGRATES.ORG",
			},
			APIOpts: map[string]any{
				utils.OptsAttributesProcessRuns: 2,
			},
		},
		blocker: false,
	}
	err = alS.V1ProcessEvent(context.Background(), ev, rply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if sort.Slice(rply.AlteredFields[0].Fields, func(i, j int) bool {
		return rply.AlteredFields[0].Fields[i] < rply.AlteredFields[0].Fields[j]
	}); !reflect.DeepEqual(expected, rply) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expected), utils.ToJSON(rply))
	}
}

func TestAttributesV1ProcessEventErrorMetaSum(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.FilterSCfg().ResourceSConns = []string{}
	conMng := engine.NewConnManager(cfg)
	engine.Cache.Clear(nil)
	db, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: db}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, conMng)
	filterS := engine.NewFilterS(cfg, conMng, dm)
	attr := &utils.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_CHANGE_TENANT_FROM_USER",
		FilterIDs: []string{"*string:~*req.Account:dan@itsyscom.com|adrian@itsyscom.com"},
		Attributes: []*utils.Attribute{
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*variable",
				Value:     utils.NewRSRParsersMustCompile("~*req.Account:s/(.*)@(.*)/${1}.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*req.Account",
				Type:      "*variable",
				Value:     utils.NewRSRParsersMustCompile("~*req.Account:s/(dan)@(.*)/${1}.${2}/:s/(adrian)@(.*)/andrei.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*composed",
				Value:     utils.NewRSRParsersMustCompile(".co.uk", utils.InfieldSep),
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: false,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	err = dm.SetAttributeProfile(context.Background(), attr, true)
	if err != nil {
		t.Error(err)
	}

	attr2 := &utils.AttributeProfile{
		Tenant: "adrian.itsyscom.com.co.uk",
		ID:     "ATTR_MATCH_TENANT",
		Attributes: []*utils.Attribute{
			{
				FilterIDs: nil,
				Path:      "*req.Password",
				Type:      utils.MetaSum,
				Value:     utils.NewRSRParsersMustCompile("CGRATES.ORG", utils.InfieldSep),
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: false,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}

	err = dm.SetAttributeProfile(context.Background(), attr2, true)
	if err != nil {
		t.Error(err)
	}

	alS := NewAttributeService(dm, filterS, nil, cfg)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "123",
		Event: map[string]any{
			utils.AccountField: "adrian@itsyscom.com",
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProcessRuns: 2,
		},
	}
	rply := &AttrSProcessEventReply{}
	err = alS.V1ProcessEvent(context.Background(), ev, rply)
	expErr := "SERVER_ERROR: NotEnoughParameters"
	if err == nil || err.Error() != expErr {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expErr, err)
	}

}

func TestAttributesV1ProcessEventErrorMetaDifference(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.FilterSCfg().ResourceSConns = []string{}
	conMng := engine.NewConnManager(cfg)
	db, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: db}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, conMng)
	filterS := engine.NewFilterS(cfg, conMng, dm)
	attr := &utils.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_CHANGE_TENANT_FROM_USER",
		FilterIDs: []string{"*string:~*req.Account:dan@itsyscom.com|adrian@itsyscom.com"},
		Attributes: []*utils.Attribute{
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*variable",
				Value:     utils.NewRSRParsersMustCompile("~*req.Account:s/(.*)@(.*)/${1}.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*req.Account",
				Type:      "*variable",
				Value:     utils.NewRSRParsersMustCompile("~*req.Account:s/(dan)@(.*)/${1}.${2}/:s/(adrian)@(.*)/andrei.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*composed",
				Value:     utils.NewRSRParsersMustCompile(".co.uk", utils.InfieldSep),
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: false,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	err = dm.SetAttributeProfile(context.Background(), attr, true)
	if err != nil {
		t.Error(err)
	}

	attr2 := &utils.AttributeProfile{
		Tenant: "adrian.itsyscom.com.co.uk",
		ID:     "ATTR_MATCH_TENANT",
		Attributes: []*utils.Attribute{
			{
				FilterIDs: nil,
				Path:      "*req.Password",
				Type:      utils.MetaDifference,
				Value:     utils.NewRSRParsersMustCompile("CGRATES.ORG", utils.InfieldSep),
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: false,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}

	err = dm.SetAttributeProfile(context.Background(), attr2, true)
	if err != nil {
		t.Error(err)
	}

	alS := NewAttributeService(dm, filterS, nil, cfg)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "123",
		Event: map[string]any{
			utils.AccountField: "adrian@itsyscom.com",
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProcessRuns: 2,
		},
	}
	rply := &AttrSProcessEventReply{}
	err = alS.V1ProcessEvent(context.Background(), ev, rply)
	expErr := "SERVER_ERROR: NotEnoughParameters"
	if err == nil || err.Error() != expErr {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expErr, err)
	}

}

func TestAttributesV1ProcessEventErrorMetaValueExponent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.FilterSCfg().ResourceSConns = []string{}
	conMng := engine.NewConnManager(cfg)
	engine.Cache.Clear(nil)
	db, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: db}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, conMng)
	filterS := engine.NewFilterS(cfg, conMng, dm)
	attr := &utils.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_CHANGE_TENANT_FROM_USER",
		FilterIDs: []string{"*string:~*req.Account:dan@itsyscom.com|adrian@itsyscom.com"},
		Attributes: []*utils.Attribute{
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*variable",
				Value:     utils.NewRSRParsersMustCompile("~*req.Account:s/(.*)@(.*)/${1}.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*req.Account",
				Type:      "*variable",
				Value:     utils.NewRSRParsersMustCompile("~*req.Account:s/(dan)@(.*)/${1}.${2}/:s/(adrian)@(.*)/andrei.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*composed",
				Value:     utils.NewRSRParsersMustCompile(".co.uk", utils.InfieldSep),
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: false,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	err = dm.SetAttributeProfile(context.Background(), attr, true)
	if err != nil {
		t.Error(err)
	}

	attr2 := &utils.AttributeProfile{
		Tenant: "adrian.itsyscom.com.co.uk",
		ID:     "ATTR_MATCH_TENANT",
		Attributes: []*utils.Attribute{
			{
				FilterIDs: nil,
				Path:      "*req.Password",
				Type:      utils.MetaValueExponent,
				Value:     utils.NewRSRParsersMustCompile("CGRATES.ORG", utils.InfieldSep),
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: false,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}

	err = dm.SetAttributeProfile(context.Background(), attr2, true)
	if err != nil {
		t.Error(err)
	}

	alS := NewAttributeService(dm, filterS, nil, cfg)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "123",
		Event: map[string]any{
			utils.AccountField: "adrian@itsyscom.com",
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProcessRuns: 2,
		},
	}
	rply := &AttrSProcessEventReply{}
	err = alS.V1ProcessEvent(context.Background(), ev, rply)
	expErr := "SERVER_ERROR: invalid arguments <[{\"Rules\":\"CGRATES.ORG\",\"Path\":\"CGRATES.ORG\"}]> to *valueExponent"
	if err == nil || err.Error() != expErr {
		t.Errorf("expected %q, received %q", expErr, err)
	}

}

func TestAttributesattributeProfileForEventNoDBConn(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	dataDB, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	alS := &AttributeS{
		cfg:   cfg,
		dm:    dm,
		fltrS: engine.NewFilterS(cfg, nil, dm),
	}

	postpaid, err := utils.NewRSRParsers(utils.MetaPostpaid, utils.InfieldSep)
	if err != nil {
		t.Error(err)
	}
	ap1 := &utils.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Attributes: []*utils.Attribute{
			{
				Path:  "*req.RequestType",
				Type:  utils.MetaConstant,
				Value: postpaid,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	err = alS.dm.SetAttributeProfile(context.Background(), ap1, true)
	if err != nil {
		t.Error(err)
	}

	ap2 := &utils.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_2",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Attributes: []*utils.Attribute{
			{
				Path:  "*req.RequestType",
				Type:  utils.MetaConstant,
				Value: postpaid,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	err = alS.dm.SetAttributeProfile(context.Background(), ap2, true)
	if err != nil {
		t.Error(err)
	}

	tnt := "cgrates.org"
	evNm := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			utils.AccountField: "1001",
		},
		utils.MetaVars: utils.MapStorage{},
	}
	lastID := ""
	alS.dm = nil

	if rcv, err := alS.attributeProfileForEvent(context.Background(), tnt, []string{"ATTR_3"}, evNm, lastID, make(map[string]int), 0, false); err == nil || err != utils.ErrNoDatabaseConn {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNoDatabaseConn, err)
	} else if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}
}

func TestAttributesattributeProfileForEventErrNotFound(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	dataDB, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	alS := &AttributeS{
		cfg:   cfg,
		dm:    dm,
		fltrS: engine.NewFilterS(cfg, nil, dm),
	}

	apNil := &utils.AttributeProfile{}
	err = alS.dm.SetAttributeProfile(context.Background(), apNil, true)
	if err != nil {
		t.Error(err)
	}

	tnt := ""
	evNm := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			utils.AccountField: "1001",
		},
		utils.MetaVars: utils.MapStorage{},
	}
	lastID := ""

	if rcv, err := alS.attributeProfileForEvent(context.Background(), tnt, []string{"ATTR_3"}, evNm, lastID, make(map[string]int), 0, false); err == nil || err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	} else if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}
}

func TestAttributesattributeProfileForEventErrPass(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	dataDB, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	alS := &AttributeS{
		cfg:   cfg,
		dm:    dm,
		fltrS: engine.NewFilterS(cfg, nil, dm),
	}

	postpaid, err := utils.NewRSRParsers(utils.MetaPostpaid, utils.InfieldSep)
	if err != nil {
		t.Error(err)
	}
	ap := &utils.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Attributes: []*utils.Attribute{
			{
				Path:  "*req.RequestType",
				Type:  utils.MetaConstant,
				Value: postpaid,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	err = alS.dm.SetAttributeProfile(context.Background(), ap, true)
	if err != nil {
		t.Error(err)
	}

	tnt := "cgrates.org"
	evNm := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			utils.AccountField: "1001",
		},
		utils.MetaVars: utils.MapStorage{},
	}
	lastID := ""

	evNm = utils.MapStorage{
		utils.MetaReq:  1,
		utils.MetaVars: utils.MapStorage{},
	}

	if rcv, err := alS.attributeProfileForEvent(context.Background(), tnt, []string{"ATTR_1"}, evNm, lastID, make(map[string]int), 0, false); err == nil || err != utils.ErrWrongPath {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrWrongPath, err)
	} else if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}
}

func TestAttributesParseAttributeSIPCID(t *testing.T) {
	exp := "12345;1001;1002"
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"cid":  "12345",
			"to":   "1001",
			"from": "1002",
		},
	}
	if out, err := ParseAttribute(dp, utils.MetaSIPCID, utils.EmptyString, utils.NewRSRParsersMustCompile("~*req.cid;~*req.to;~*req.from", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString); err != nil {
		t.Fatal(err)
	} else if exp != out {
		t.Errorf("Expected %q, Received %q", exp, out)
	}

	dp = utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"cid":  "12345",
			"to":   "1002",
			"from": "1001",
		},
	}
	if out, err := ParseAttribute(dp, utils.MetaSIPCID, utils.EmptyString, utils.NewRSRParsersMustCompile("~*req.cid;~*req.to;~*req.from", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString); err != nil {
		t.Fatal(err)
	} else if exp != out {
		t.Errorf("Expected %q, Received %q", exp, out)
	}

	exp = "12345;1001;1002;1003"
	dp = utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"cid":   "12345",
			"to":    "1001",
			"from":  "1002",
			"extra": "1003",
		},
	}
	if out, err := ParseAttribute(dp, utils.MetaSIPCID, utils.EmptyString, utils.NewRSRParsersMustCompile("~*req.cid;~*req.to;~*req.extra;~*req.from",
		utils.InfieldSep), 0, utils.EmptyString, utils.EmptyString); err != nil {
		t.Fatal(err)
	} else if exp != out {
		t.Errorf("Expected %q, Received %q", exp, out)
	}

	dp = utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"cid":   "12345",
			"to":    "1002",
			"from":  "1001",
			"extra": "1003",
		},
	}
	if out, err := ParseAttribute(dp, utils.MetaSIPCID, utils.EmptyString, utils.NewRSRParsersMustCompile("~*req.cid;~*req.extra;~*req.to;~*req.from",
		utils.InfieldSep), 0, utils.EmptyString, utils.EmptyString); err != nil {
		t.Fatal(err)
	} else if exp != out {
		t.Errorf("Expected %q, Received %q", exp, out)
	}

	dp = utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"cid": "12345",
		},
	}
	if _, err := ParseAttribute(dp, utils.MetaSIPCID, utils.EmptyString, utils.NewRSRParsersMustCompile("~*req.cid;~*req.extra;~*req.to;~*req.from", utils.
		InfieldSep), 0, utils.EmptyString, utils.EmptyString); err != utils.ErrNotFound {
		t.Errorf("Expected <%+v>, received <%+v>", utils.ErrNotFound, err)
	}
}

func TestAttributesParseAttributeSIPCIDWrongPathErr(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"cid":  "12345",
			"to":   "1001",
			"from": "1002",
		},
		utils.MetaOpts: 13,
	}
	value := utils.NewRSRParsersMustCompile("~*req.cid;~*req.to;~*req.from;~*opts.WrongPath", utils.InfieldSep)
	if _, err := ParseAttribute(dp, utils.MetaSIPCID, utils.EmptyString, value,
		0, time.UTC.String(), utils.EmptyString); err == nil ||
		err.Error() != utils.ErrWrongPath.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrWrongPath, err)
	}
}

func TestAttributesParseAttributeSIPCIDNotFoundErr(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"to":   "1001",
			"from": "1002",
		},
	}
	value := utils.NewRSRParsersMustCompile("~*req.cid;~*req.to;~*req.from", utils.InfieldSep)
	if _, err := ParseAttribute(dp, utils.MetaSIPCID, utils.EmptyString, value,
		0, time.UTC.String(), utils.EmptyString); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestAttributesParseAttributeSIPCIDInvalidArguments(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"to":   "1001",
			"from": "1002",
		},
	}
	value := utils.RSRParsers{}
	experr := `invalid number of arguments <[]> to *sipcid`
	if _, err := ParseAttribute(dp, utils.MetaSIPCID, utils.EmptyString, value,
		0, time.UTC.String(), utils.EmptyString); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestAttributesV1ProcessEventMultipleRuns1(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().IndexedSelects = false
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	alS := NewAttributeService(dm, filterS, nil, cfg)

	postpaid := utils.NewRSRParsersMustCompile(utils.MetaPostpaid, utils.InfieldSep)
	pw := utils.NewRSRParsersMustCompile("CGRateS.org", utils.InfieldSep)

	ap1 := &utils.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR1",
		FilterIDs: []string{"*notexists:~*vars.*processedProfileIDs[<~*vars.*apTenantID>]:"},
		Attributes: []*utils.Attribute{
			{
				Path:  "*req.Password",
				Type:  utils.MetaConstant,
				Value: pw,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	err = alS.dm.SetAttributeProfile(context.Background(), ap1, true)
	if err != nil {
		t.Error(err)
	}

	ap2 := &utils.AttributeProfile{
		Tenant: "cgrates.org",
		ID:     "ATTR2",
		Attributes: []*utils.Attribute{
			{
				Path:  "*req.RequestType",
				Type:  utils.MetaConstant,
				Value: postpaid,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	err = alS.dm.SetAttributeProfile(context.Background(), ap2, true)
	if err != nil {
		t.Error(err)
	}

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "AttrProcessEventMultipleRuns",
		Event: map[string]any{
			"Password": "passwd",
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProcessRuns: 4,
			utils.OptsAttributesProfileIDs:  []string{"ATTR1", "ATTR2"},
		},
	}
	reply := &AttrSProcessEventReply{}
	exp := &AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR2",
				Fields:           []string{"*req.RequestType"},
			},
			{
				MatchedProfileID: "cgrates.org:ATTR1",
				Fields:           []string{"*req.Password"},
			},
			{
				MatchedProfileID: "cgrates.org:ATTR2",
				Fields:           []string{"*req.RequestType"},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "AttrProcessEventMultipleRuns",
			Event: map[string]any{
				"Password":        "CGRateS.org",
				utils.RequestType: utils.MetaPostpaid,
			},
			APIOpts: map[string]any{
				utils.OptsAttributesProfileIDs:  []string{"ATTR1", "ATTR2"},
				utils.OptsAttributesProcessRuns: 4,
			},
		},
	}

	if err := alS.V1ProcessEvent(context.Background(), ev, reply); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(reply, exp) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(exp), utils.ToJSON(reply))
		}
	}
}

func TestAttributesV1ProcessEventMultipleRuns2(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().IndexedSelects = false
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	alS := NewAttributeService(dm, filterS, nil, cfg)

	postpaid := utils.NewRSRParsersMustCompile(utils.MetaPostpaid, utils.InfieldSep)
	pw := utils.NewRSRParsersMustCompile("CGRateS.org", utils.InfieldSep)
	paypal := utils.NewRSRParsersMustCompile("cgrates@paypal.com", utils.InfieldSep)

	ap1 := &utils.AttributeProfile{
		Tenant: "cgrates.org",
		ID:     "ATTR1",
		Attributes: []*utils.Attribute{
			{
				Path:  "*req.Password",
				Type:  utils.MetaConstant,
				Value: pw,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	err = alS.dm.SetAttributeProfile(context.Background(), ap1, true)
	if err != nil {
		t.Error(err)
	}

	ap2 := &utils.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR2",
		FilterIDs: []string{"*exists:~*vars.*processedProfileIDs[cgrates.org:ATTR1]:"},
		Attributes: []*utils.Attribute{
			{
				Path:  "*req.RequestType",
				Type:  utils.MetaConstant,
				Value: postpaid,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	err = alS.dm.SetAttributeProfile(context.Background(), ap2, true)
	if err != nil {
		t.Error(err)
	}

	ap3 := &utils.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR3",
		FilterIDs: []string{"*exists:~*vars.*processedProfileIDs[cgrates.org:ATTR2]:"},
		Attributes: []*utils.Attribute{
			{
				Path:  "*req.PaypalAccount",
				Type:  utils.MetaConstant,
				Value: paypal,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	err = alS.dm.SetAttributeProfile(context.Background(), ap3, true)
	if err != nil {
		t.Error(err)
	}

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "AttrProcessEventMultipleRuns",
		Event:  map[string]any{},
		APIOpts: map[string]any{
			utils.OptsAttributesProcessRuns: 3,
		},
	}

	reply := &AttrSProcessEventReply{}
	exp := &AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR1",
				Fields:           []string{"*req.Password"},
			},
			{
				MatchedProfileID: "cgrates.org:ATTR2",
				Fields:           []string{"*req.RequestType"},
			},
			{
				MatchedProfileID: "cgrates.org:ATTR3",
				Fields:           []string{"*req.PaypalAccount"},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "AttrProcessEventMultipleRuns",
			Event: map[string]any{
				"Password":        "CGRateS.org",
				"PaypalAccount":   "cgrates@paypal.com",
				utils.RequestType: utils.MetaPostpaid,
			},
			APIOpts: map[string]any{
				utils.OptsAttributesProcessRuns: 3,
			},
		},
	}
	if err := alS.V1ProcessEvent(context.Background(), ev, reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(exp), utils.ToJSON(reply))
	}
}

func TestAttributesV1GetAttributeForEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.FilterSCfg().ResourceSConns = []string{}
	conMng := engine.NewConnManager(cfg)
	db, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: db}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, conMng)
	filterS := engine.NewFilterS(cfg, conMng, dm)
	attr := &utils.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_CHANGE_TENANT_FROM_USER",
		FilterIDs: []string{"*string:~*req.Account:dan@itsyscom.com|adrian@itsyscom.com"},
		Attributes: []*utils.Attribute{
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*variable",
				Value:     utils.NewRSRParsersMustCompile("~*req.Account:s/(.*)@(.*)/${1}.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*req.Account",
				Type:      "*variable",
				Value:     utils.NewRSRParsersMustCompile("~*req.Account:s/(dan)@(.*)/${1}.${2}/:s/(adrian)@(.*)/andrei.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*composed",
				Value:     utils.NewRSRParsersMustCompile(".co.uk", utils.InfieldSep),
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: false,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	err = dm.SetAttributeProfile(context.Background(), attr, true)
	if err != nil {
		t.Error(err)
	}

	attr2 := &utils.AttributeProfile{
		Tenant: "adrian.itsyscom.com.co.uk",
		ID:     "ATTR_MATCH_TENANT",
		Attributes: []*utils.Attribute{
			{
				FilterIDs: nil,
				Path:      "*req.Password",
				Type:      utils.MetaConstant,
				Value:     utils.NewRSRParsersMustCompile("CGRATES.ORG", utils.InfieldSep),
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: false,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}

	err = dm.SetAttributeProfile(context.Background(), attr2, true)
	if err != nil {
		t.Error(err)
	}

	alS := NewAttributeService(dm, filterS, nil, cfg)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "123",
		Event: map[string]any{
			utils.AccountField: "adrian@itsyscom.com",
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProcessRuns: 2,
		},
	}
	rply := &utils.APIAttributeProfile{}
	expected := &utils.APIAttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_CHANGE_TENANT_FROM_USER",
		FilterIDs: []string{"*string:~*req.Account:dan@itsyscom.com|adrian@itsyscom.com"},
		Attributes: []*utils.ExternalAttribute{
			{
				Path:  "*tenant",
				Type:  "*variable",
				Value: "~*req.Account:s/(.*)@(.*)/${1}.${2}/",
			},
			{
				Path:  "*req.Account",
				Type:  "*variable",
				Value: "~*req.Account:s/(dan)@(.*)/${1}.${2}/:s/(adrian)@(.*)/andrei.${2}/",
			},
			{
				Path:  "*tenant",
				Type:  "*composed",
				Value: ".co.uk",
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: false,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}

	err = alS.V1GetAttributeForEvent(context.Background(), ev, rply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expected), utils.ToJSON(rply))
	}
}

func TestAttributesV1GetAttributeForEventErrorBoolOpts(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.FilterSCfg().ResourceSConns = []string{}
	conMng := engine.NewConnManager(cfg)
	db, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: db}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, conMng)
	filterS := engine.NewFilterS(cfg, conMng, dm)
	attr := &utils.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_CHANGE_TENANT_FROM_USER",
		FilterIDs: []string{"*string:~*req.Account:dan@itsyscom.com|adrian@itsyscom.com"},
		Attributes: []*utils.Attribute{
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*variable",
				Value:     utils.NewRSRParsersMustCompile("~*req.Account:s/(.*)@(.*)/${1}.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*req.Account",
				Type:      "*variable",
				Value:     utils.NewRSRParsersMustCompile("~*req.Account:s/(dan)@(.*)/${1}.${2}/:s/(adrian)@(.*)/andrei.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*composed",
				Value:     utils.NewRSRParsersMustCompile(".co.uk", utils.InfieldSep),
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: false,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	err = dm.SetAttributeProfile(context.Background(), attr, true)
	if err != nil {
		t.Error(err)
	}

	attr2 := &utils.AttributeProfile{
		Tenant: "adrian.itsyscom.com.co.uk",
		ID:     "ATTR_MATCH_TENANT",
		Attributes: []*utils.Attribute{
			{
				FilterIDs: nil,
				Path:      "*req.Password",
				Type:      utils.MetaConstant,
				Value:     utils.NewRSRParsersMustCompile("CGRATES.ORG", utils.InfieldSep),
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: false,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}

	err = dm.SetAttributeProfile(context.Background(), attr2, true)
	if err != nil {
		t.Error(err)
	}

	alS := NewAttributeService(dm, filterS, nil, cfg)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "123",
		Event: map[string]any{
			utils.AccountField: "adrian@itsyscom.com",
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProcessRuns: 2,
			utils.MetaProfileIgnoreFilters:  time.Second,
		},
	}
	rply := &utils.APIAttributeProfile{}

	err = alS.V1GetAttributeForEvent(context.Background(), ev, rply)
	if err == nil || err.Error() != "cannot convert field: 1s to bool" {
		t.Errorf("\nExpected <cannot convert field: 1s to bool>, \nReceived <%+v>", err)
	}

}

func TestAttributesV1GetAttributeForEventErrorNil(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.FilterSCfg().ResourceSConns = []string{}
	conMng := engine.NewConnManager(cfg)
	db, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: db}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, conMng)
	filterS := engine.NewFilterS(cfg, conMng, dm)
	attr := &utils.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_CHANGE_TENANT_FROM_USER",
		FilterIDs: []string{"*string:~*req.Account:dan@itsyscom.com|adrian@itsyscom.com"},
		Attributes: []*utils.Attribute{
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*variable",
				Value:     utils.NewRSRParsersMustCompile("~*req.Account:s/(.*)@(.*)/${1}.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*req.Account",
				Type:      "*variable",
				Value:     utils.NewRSRParsersMustCompile("~*req.Account:s/(dan)@(.*)/${1}.${2}/:s/(adrian)@(.*)/andrei.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*composed",
				Value:     utils.NewRSRParsersMustCompile(".co.uk", utils.InfieldSep),
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: false,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	err = dm.SetAttributeProfile(context.Background(), attr, true)
	if err != nil {
		t.Error(err)
	}

	attr2 := &utils.AttributeProfile{
		Tenant: "adrian.itsyscom.com.co.uk",
		ID:     "ATTR_MATCH_TENANT",
		Attributes: []*utils.Attribute{
			{
				FilterIDs: nil,
				Path:      "*req.Password",
				Type:      utils.MetaConstant,
				Value:     utils.NewRSRParsersMustCompile("CGRATES.ORG", utils.InfieldSep),
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: false,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}

	err = dm.SetAttributeProfile(context.Background(), attr2, true)
	if err != nil {
		t.Error(err)
	}

	alS := NewAttributeService(dm, filterS, nil, cfg)
	rply := &utils.APIAttributeProfile{}

	err = alS.V1GetAttributeForEvent(context.Background(), nil, rply)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [CGREvent]" {
		t.Errorf("\nExpected <MANDATORY_IE_MISSING: [CGREvent]>, \nReceived <%+v>", err)
	}

}

func TestAttributesV1GetAttributeForEventErrOptsI(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.FilterSCfg().ResourceSConns = []string{}
	conMng := engine.NewConnManager(cfg)
	db, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: db}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, conMng)
	filterS := engine.NewFilterS(cfg, conMng, dm)
	attr := &utils.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_CHANGE_TENANT_FROM_USER",
		FilterIDs: []string{"*string:~*req.Account:dan@itsyscom.com|adrian@itsyscom.com"},
		Attributes: []*utils.Attribute{
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*variable",
				Value:     utils.NewRSRParsersMustCompile("~*req.Account:s/(.*)@(.*)/${1}.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*req.Account",
				Type:      "*variable",
				Value:     utils.NewRSRParsersMustCompile("~*req.Account:s/(dan)@(.*)/${1}.${2}/:s/(adrian)@(.*)/andrei.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*composed",
				Value:     utils.NewRSRParsersMustCompile(".co.uk", utils.InfieldSep),
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: false,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	err = dm.SetAttributeProfile(context.Background(), attr, true)
	if err != nil {
		t.Error(err)
	}

	attr2 := &utils.AttributeProfile{
		Tenant: "adrian.itsyscom.com.co.uk",
		ID:     "ATTR_MATCH_TENANT",
		Attributes: []*utils.Attribute{
			{
				FilterIDs: nil,
				Path:      "*req.Password",
				Type:      utils.MetaConstant,
				Value:     utils.NewRSRParsersMustCompile("CGRATES.ORG", utils.InfieldSep),
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: false,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}

	err = dm.SetAttributeProfile(context.Background(), attr2, true)
	if err != nil {
		t.Error(err)
	}

	alS := NewAttributeService(dm, filterS, nil, cfg)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "123",
		Event: map[string]any{
			utils.AccountField: "adrian@itsyscom.com",
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProcessRuns: 2,
			utils.OptsAttributesProfileIDs:  time.Second,
		},
	}
	rply := &utils.APIAttributeProfile{}

	err = alS.V1GetAttributeForEvent(context.Background(), ev, rply)
	if err == nil || err.Error() != "cannot convert field: 1s to []string" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "cannot convert field: 1s to []string", err)
	}

}
func TestAttributesProcessEventProfileIgnoreFilters(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	aA := NewAttributeService(dm, filterS, nil, cfg)
	cfg.AttributeSCfg().Opts.ProfileIgnoreFilters = []*config.DynamicBoolOpt{
		config.NewDynamicBoolOpt(nil, "", true, nil),
	}
	acPrf := &utils.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "AC1",
		FilterIDs: []string{"*string:~*req.Attribute:testAttrValue"},
	}
	if err := dm.SetAttributeProfile(context.Background(), acPrf, true); err != nil {
		t.Error(err)
	}
	//should match the attr profile for event because the option is false but the filter matches
	args2 := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "AcProcessEvent",
		Event: map[string]any{
			"Attribute": "testAttrValue",
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProfileIDs: []string{"AC1"},
			utils.MetaProfileIgnoreFilters: false,
		},
	}
	eNM := utils.MapStorage{
		utils.MetaReq:  args2.Event,
		utils.MetaOpts: args2.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}
	exp2 := &AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:AC1",
				Fields:           []string{},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "AcProcessEvent",
			Event: map[string]any{
				"Attribute": "testAttrValue",
			},
			APIOpts: map[string]any{
				utils.OptsAttributesProfileIDs: []string{"AC1"},
				utils.MetaProfileIgnoreFilters: false,
			},
		},
	}
	if rcv2, err := aA.processEvent(context.Background(), args2.Tenant, args2, eNM, engine.NewDynamicDP(context.TODO(), nil, nil, nil, nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv2, exp2) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(exp2), utils.ToJSON(rcv2))
	}
	//should match the attr profile for event because the option is true even if the filter doesn't match
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "AcProcessEvent2",
		Event: map[string]any{
			"Attribute": "testAttrValue2",
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProfileIDs: []string{"AC1"},
			utils.MetaProfileIgnoreFilters: true,
		},
	}
	eNM2 := utils.MapStorage{
		utils.MetaReq:  args.Event,
		utils.MetaOpts: args.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}
	exp := &AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:AC1",
				Fields:           []string{},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "AcProcessEvent2",
			Event: map[string]any{
				"Attribute": "testAttrValue2",
			},
			APIOpts: map[string]any{
				utils.OptsAttributesProfileIDs: []string{"AC1"},
				utils.MetaProfileIgnoreFilters: true,
			},
		},
	}
	if rcv, err := aA.processEvent(context.Background(), args.Tenant, args, eNM2, engine.NewDynamicDP(context.TODO(), nil, nil, nil, nil, nil, nil, "cgrates.org", eNM2), utils.EmptyString, make(map[string]int), 0); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestAttributeServicesProcessEventGetStringSliceOptsError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	aA := NewAttributeService(dm, filterS, nil, cfg)
	args2 := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "AcProcessEvent",
		APIOpts: map[string]any{
			utils.OptsAttributesProfileIDs: time.Second,
		},
	}
	eNM := utils.MapStorage{
		utils.MetaReq:  args2.Event,
		utils.MetaOpts: args2.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}
	_, err = aA.processEvent(context.Background(), args2.Tenant, args2, eNM, engine.NewDynamicDP(context.TODO(), nil, nil, nil, nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
	if err == nil || err.Error() != "cannot convert field: 1s to []string" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "cannot convert field: 1s to []string", err)
	}
}

func TestAttributeServicesProcessEventGetBoolOptsError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	aA := NewAttributeService(dm, filterS, nil, cfg)
	args2 := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "AcProcessEvent",
		APIOpts: map[string]any{
			utils.MetaProfileIgnoreFilters: time.Second,
		},
	}
	eNM := utils.MapStorage{
		utils.MetaReq:  args2.Event,
		utils.MetaOpts: args2.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}
	_, err = aA.processEvent(context.Background(), args2.Tenant, args2, eNM, engine.NewDynamicDP(context.TODO(), nil, nil, nil, nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
	if err == nil || err.Error() != "cannot convert field: 1s to bool" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "cannot convert field: 1s to bool", err)
	}
}

func TestAttributesParseAttributeMetaNone(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"cid":  "12345",
			"to":   "1001",
			"from": "1002",
		},
	}
	if out, err := ParseAttribute(dp, utils.MetaNone, utils.EmptyString, utils.NewRSRParsersMustCompile("~*req.cid;~*req.to;~*req.from", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString); err != nil {
		t.Fatal(err)
	} else if out != nil {
		t.Errorf("Expected %+v, Received %+v", nil, out)
	}
}

func TestAttributesParseAttributeMetaUsageDifferenceBadValError(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"cid":  "12345",
			"to":   "1001",
			"from": "1002",
		},
	}
	_, err := ParseAttribute(dp, utils.MetaUsageDifference, utils.EmptyString, utils.NewRSRParsersMustCompile("", utils.InfieldSep), 0, utils.EmptyString, utils.EmptyString)

	if err == nil || err.Error() != "invalid arguments <null> to *usageDifference" {
		t.Fatal(err)
	}
}

func TestAttributesParseAttributeMetaCCUsageError(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"cid":  "12345",
			"to":   "1001",
			"from": "1002",
		},
	}
	_, err := ParseAttribute(dp, utils.MetaCCUsage, utils.EmptyString, utils.NewRSRParsersMustCompile("::;~*req.to;~*req.from", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString)
	if err == nil || err.Error() != "invalid requestNumber <::> to *ccUsage" {
		t.Fatal(err)
	}
}

func TestAttributesProcessEventSetError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	aA := NewAttributeService(dm, filterS, nil, cfg)
	acPrf := &utils.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "AC1",
		FilterIDs: []string{"*string:~*req.Attribute:testAttrValue"},
		Attributes: []*utils.Attribute{
			{
				Path: "",
			},
		},
	}
	if err := dm.SetAttributeProfile(context.Background(), acPrf, true); err != nil {
		t.Error(err)
	}

	args2 := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "AcProcessEvent",
		Event: map[string]any{
			"Attribute": "testAttrValue",
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProfileIDs: []string{"AC1"},
		},
	}
	eNM := utils.MapStorage{
		utils.MetaReq:  args2.Event,
		utils.MetaOpts: args2.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}

	if _, err := aA.processEvent(context.Background(), args2.Tenant, args2, eNM, engine.NewDynamicDP(context.TODO(), nil, nil, nil, nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0); err != nil {
		t.Error(err)
	}
}
func TestAttributesAttributeServiceV1PrcssEvPrcssRunsGetIntOptsErr(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().IndexedSelects = false
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	alS := NewAttributeService(dm, filterS, nil, cfg)
	pw := utils.NewRSRParsersMustCompile("CGRateS.org", utils.InfieldSep)

	ap1 := &utils.AttributeProfile{
		Tenant: "cgrates.org",
		ID:     "ATTR1",
		Attributes: []*utils.Attribute{
			{
				Path:  "*req.Password",
				Type:  utils.MetaConstant,
				Value: pw,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	err = alS.dm.SetAttributeProfile(context.Background(), ap1, true)
	if err != nil {
		t.Error(err)
	}

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "AttrProcessEventMultipleRuns",
		Event:  map[string]any{},
		APIOpts: map[string]any{
			utils.OptsAttributesProcessRuns: "errVal",
		},
	}

	reply := &AttrSProcessEventReply{}
	exrErr := `strconv.Atoi: parsing "errVal": invalid syntax`
	if err := alS.V1ProcessEvent(context.Background(), ev, reply); err == nil || err.Error() != exrErr {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", exrErr, err)
	}
}

func TestAttributesAttributeServiceV1PrcssEvProfRunsGetIntOptsErr(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().IndexedSelects = false
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	alS := NewAttributeService(dm, filterS, nil, cfg)
	pw := utils.NewRSRParsersMustCompile("CGRateS.org", utils.InfieldSep)

	ap1 := &utils.AttributeProfile{
		Tenant: "cgrates.org",
		ID:     "ATTR1",
		Attributes: []*utils.Attribute{
			{
				Path:  "*req.Password",
				Type:  utils.MetaConstant,
				Value: pw,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	err = alS.dm.SetAttributeProfile(context.Background(), ap1, true)
	if err != nil {
		t.Error(err)
	}

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "AttrProcessEventMultipleRuns",
		Event:  map[string]any{},
		APIOpts: map[string]any{
			utils.OptsAttributesProfileRuns: "errVal",
		},
	}

	reply := &AttrSProcessEventReply{}
	exrErr := `strconv.Atoi: parsing "errVal": invalid syntax`
	if err := alS.V1ProcessEvent(context.Background(), ev, reply); err == nil || err.Error() != exrErr {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", exrErr, err)
	}
}

func TestAttributesParseAttributeMetaGeneric(t *testing.T) {
	exp := "1234510011002"
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"cid":  "12345",
			"to":   "1001",
			"from": "1002",
		},
	}
	if out, err := ParseAttribute(dp, utils.MetaGeneric, utils.EmptyString, utils.NewRSRParsersMustCompile("~*req.cid;~*req.to;~*req.from", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString); err != nil {
		t.Fatal(err)
	} else if exp != out {
		t.Errorf("Expected %q, Received %q", exp, out)
	}
}

func TestAttributesParseAttributeError(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"cid":  "12345",
			"to":   "1001",
			"from": "1002",
		},
	}
	if _, err := ParseAttribute(dp, "badType", utils.EmptyString, utils.NewRSRParsersMustCompile("~*req.cid;~*req.to;~*req.from", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString); err == nil || err.Error() != "unsupported type: <badType>" {
		t.Errorf("Expected %q, Received %q", "unsupported type: <badType>", err)
	}
}

func TestAttributesProcessEventPasswordAttribute(t *testing.T) {
	tmp := engine.Cache
	tmpC := config.CgrConfig()
	defer func() {
		engine.Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	attrS := NewAttributeService(dm, filterS, nil, cfg)

	value := utils.NewRSRParsersMustCompile("abcd123", utils.RSRSep)

	attrPrf := &utils.AttributeProfile{
		Tenant: "cgrates.org",
		ID:     "ATTR_TEST",
		Attributes: []*utils.Attribute{
			{
				Path:  "*req.Password",
				Type:  utils.MetaPassword,
				Value: value,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}

	if err := dm.SetAttributeProfile(context.Background(), attrPrf, true); err != nil {
		t.Fatal(err)
	}

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EventHashPw",
		Event: map[string]any{
			"Password": "321dcba",
		},
	}

	exp := AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_TEST",
				Fields:           []string{"*req.Password"},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "EventHashPw",
			Event: map[string]any{
				"Password": "abcd123",
			},
			APIOpts: map[string]any{},
		},
	}
	var hashedPw string
	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.Background(), cgrEv, &reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(reply.AlteredFields, exp.AlteredFields) {
		t.Fatalf("expected: <%+v>,\nreceived: <%+v>",
			utils.ToJSON(exp.AlteredFields), utils.ToJSON(reply.AlteredFields))
	} else {
		hashedPw = utils.IfaceAsString(reply.CGREvent.Event["Password"])
		if !utils.VerifyHash(hashedPw, "abcd123") {
			t.Fatalf("expected: <%+v>, \nreceived: <%+v>", "abcd123", hashedPw)
		}
		exp.CGREvent.Event["Password"] = hashedPw
		if !reflect.DeepEqual(reply.CGREvent, exp.CGREvent) {
			t.Fatalf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(exp.CGREvent), utils.ToJSON(reply.CGREvent))
		}
	}

	value = utils.NewRSRParsersMustCompile(hashedPw, utils.RSRSep)
	expAttrPrf := &utils.AttributeProfile{
		Tenant: "cgrates.org",
		ID:     "ATTR_TEST",
		Attributes: []*utils.Attribute{
			{
				Path:  "*req.Password",
				Type:  utils.MetaConstant,
				Value: value,
			},
		},
		Weights: make(utils.DynamicWeights, 1),
	}
	expAttrPrf.Weights[0] = &utils.DynamicWeight{

		Weight: 10,
	}
	if rcvAttrPrf, err := dm.GetAttributeProfile(context.Background(), attrPrf.Tenant, attrPrf.ID, true, true,
		utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(rcvAttrPrf, expAttrPrf) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expAttrPrf), utils.ToJSON(rcvAttrPrf))
	}
}

func TestAttributesSetAttributeProfilePasswordAttr(t *testing.T) {
	tmp := engine.Cache
	tmpC := config.CgrConfig()
	defer func() {
		engine.Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)

	value := utils.NewRSRParsersMustCompile("abcd123", utils.RSRSep)
	attrPrf := &utils.AttributeProfile{
		Tenant: "cgrates.org",
		ID:     "ATTR_TEST",
		Attributes: []*utils.Attribute{
			{
				Path:  "*req.Password",
				Type:  utils.MetaPassword,
				Value: value,
			},
		},
		Weights: make(utils.DynamicWeights, 1),
	}
	attrPrf.Weights[0] = &utils.DynamicWeight{

		Weight: 20,
	}
	if err := dm.SetAttributeProfile(context.Background(), attrPrf, true); err != nil {
		t.Fatal(err)
	}

	exp := &utils.AttributeProfile{
		Tenant: "cgrates.org",
		ID:     "ATTR_TEST",
		Attributes: []*utils.Attribute{
			{
				Path: "*req.Password",
				Type: utils.MetaConstant,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}

	if rcv, err := dm.GetAttributeProfile(context.Background(), attrPrf.Tenant, attrPrf.ID, true, true,
		utils.NonTransactional); err != nil {
		t.Error(err)
	} else if hashedPw := rcv.Attributes[0].Value.GetRule(); !utils.VerifyHash(hashedPw, "abcd123") {
		t.Errorf("Received an incorrect password")
	} else {
		rcv.Attributes[0].Value = nil
		if !reflect.DeepEqual(rcv, exp) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(exp), utils.ToJSON(rcv))
		}
	}
}
