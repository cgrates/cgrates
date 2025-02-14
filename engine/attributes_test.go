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
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestParseAtributeUsageDiffVal1(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"usage1": "20",
			"usage2": "35",
		},
	}
	_, err := ParseAttribute(dp, utils.MetaUsageDifference, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.wrong;~*req.usage2", utils.InfieldSep),
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
	_, err := ParseAttribute(dp, utils.MetaUsageDifference, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.usage1;~*req.wrong", utils.InfieldSep),
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
	_, err := ParseAttribute(dp, utils.MetaUsageDifference, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.usage1;~*req.usage2", utils.InfieldSep),
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
	_, err := ParseAttribute(dp, utils.MetaUsageDifference, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.usage1;~*req.usage2", utils.InfieldSep),
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
	_, err := ParseAttribute(dp, utils.MetaSum, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.valid", utils.InfieldSep),
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
	_, err := ParseAttribute(dp, utils.MetaDifference, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.valid", utils.InfieldSep),
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
	_, err := ParseAttribute(dp, utils.MetaMultiply, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.valid", utils.InfieldSep),
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
	_, err := ParseAttribute(dp, utils.MetaDivide, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.valid", utils.InfieldSep),
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
	_, err := ParseAttribute(dp, utils.MetaValueExponent, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.wrong;~*req.exponent", utils.InfieldSep),
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
	_, err := ParseAttribute(dp, utils.MetaValueExponent, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.number;~*req.wrong", utils.InfieldSep),
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
	_, err := ParseAttribute(dp, utils.MetaValueExponent, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.number;~*req.exponent", utils.InfieldSep),
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
	_, err := ParseAttribute(dp, utils.MetaValueExponent, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.number;~*req.exponent", utils.InfieldSep),
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
	_, err := ParseAttribute(dp, utils.MetaUnixTimestamp, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.wrong", utils.InfieldSep),
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
	_, err := ParseAttribute(dp, utils.MetaUnixTimestamp, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.unix_timestamp", utils.InfieldSep),
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
	_, err := ParseAttribute(dp, utils.MetaPrefix, "```", config.NewRSRParsersMustCompile("~*req.prefix", utils.InfieldSep),
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
	_, err := ParseAttribute(dp, utils.MetaPrefix, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.wrong", utils.InfieldSep),
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
	_, err := ParseAttribute(dp, utils.MetaSuffix, "```", config.NewRSRParsersMustCompile("~*req.suffix", utils.InfieldSep),
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
	_, err := ParseAttribute(dp, utils.MetaSuffix, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.wrong", utils.InfieldSep),
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
	_, err := ParseAttribute(dp, utils.MetaCCUsage, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.cc1;~*req.cc2", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString)
	errExp := `invalid arguments <[{"Rules":"~*req.cc1"},{"Rules":"~*req.cc2"}]> to *ccUsage`
	if err == nil || err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
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
	_, err := ParseAttribute(dp, utils.MetaCCUsage, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.wrong;~*req.cc2;~*req.cc3", utils.InfieldSep),
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
	_, err := ParseAttribute(dp, utils.MetaCCUsage, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.cc1;~*req.cc2;~*req.cc3", utils.InfieldSep),
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
	_, err := ParseAttribute(dp, utils.MetaCCUsage, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.cc1;~*req.wrong;~*req.cc3", utils.InfieldSep),
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
	_, err := ParseAttribute(dp, utils.MetaCCUsage, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.cc1;~*req.cc2;~*req.cc3", utils.InfieldSep),
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
	_, err := ParseAttribute(dp, utils.MetaCCUsage, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.cc1;~*req.cc2;~*req.wrong", utils.InfieldSep),
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
	_, err := ParseAttribute(dp, utils.MetaCCUsage, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.cc1;~*req.cc2;~*req.cc3", utils.InfieldSep),
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
	out, err := ParseAttribute(dp, utils.MetaCCUsage, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.cc1;~*req.cc2;~*req.cc3", utils.InfieldSep),
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
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	attrS := NewAttributeService(dm, filterS, cfg)

	attrIDs, err := utils.OptAsStringSlice(attrEvs[0].APIOpts, utils.OptsAttributesProfileIDs)
	if err != nil {
		t.Fatal(err)
	}

	value := config.NewRSRParsersMustCompile("abcd123", utils.RSRSep)

	attrPrf := &AttributeProfile{
		Tenant: "cgrates.org",
		ID:     "ATTR_TEST",
		Attributes: []*Attribute{
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
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	attrS := NewAttributeService(dm, filterS, cfg)

	value := config.NewRSRParsersMustCompile("abcd123", utils.RSRSep)

	attrPrf := &AttributeProfile{
		Tenant: "cgrates.org",
		ID:     "ATTR_TEST",
		Attributes: []*Attribute{
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
	_, err := attrS.processEvent(context.TODO(), attrEvs.Tenant, attrEvs, eNM, newDynamicDP(context.TODO(), nil, nil, nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
	if err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%+v>, received error <%+v>", expErr, err)
	}

}

func TestAttributeSProcessEventPassErr(t *testing.T) {

	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	attrS := NewAttributeService(dm, filterS, cfg)

	attrPrf := &AttributeProfile{
		Tenant: "cgrates.org",
		ID:     "ATTR_TEST",
		Attributes: []*Attribute{
			{
				FilterIDs: []string{"*apiban:~*req.<~*req.IP>{*}:*all"},
				Path:      "*req.Password",
				Type:      utils.MetaPassword,
				Value:     config.NewRSRParsersMustCompile("abcd123", utils.RSRSep),
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
	_, err := attrS.processEvent(context.TODO(), attrPrf.Tenant, ev, eNM, newDynamicDP(context.TODO(), nil, nil, nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
	if err == nil || err.Error() != expErr {
		t.Errorf("Expected error %s received: %v", expErr, err)
	}

}

func TestAttributeSProcessAttrBlockerFromDynamicsErr(t *testing.T) {
	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	attrS := NewAttributeService(dm, filterS, cfg)

	attrPrf := &AttributeProfile{
		Tenant: "cgrates.org",
		ID:     "ATTR_TEST",
		Attributes: []*Attribute{
			{
				Blockers: utils.DynamicBlockers{{
					FilterIDs: []string{"*stirng:~*req.Account:1001"},
					Blocker:   false,
				}},
				Path:  "*req.Password",
				Type:  utils.MetaPassword,
				Value: config.NewRSRParsersMustCompile("abcd123", utils.RSRSep),
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
	_, err := attrS.processEvent(context.TODO(), attrPrf.Tenant, ev, eNM, newDynamicDP(context.TODO(), nil, nil, nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
	if err == nil || err.Error() != expErr {
		t.Errorf("Expected error %s received: %v", expErr, err)
	}

}

func TestAttributeSProcessSubstituteRmvBlockerTrue(t *testing.T) {
	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	attrS := NewAttributeService(dm, filterS, cfg)

	attrPrf := &AttributeProfile{
		Tenant: "cgrates.org",
		ID:     "ATTR_TEST",
		Attributes: []*Attribute{
			{
				Path:  utils.MetaRemove,
				Type:  utils.MetaVariable,
				Value: config.NewRSRParsersMustCompile(utils.MetaRemove, utils.RSRSep),
			},
			{
				Blockers: utils.DynamicBlockers{{
					Blocker: true,
				}},
				Path:  "*req.Password",
				Type:  utils.MetaPassword,
				Value: config.NewRSRParsersMustCompile("abcd123", utils.RSRSep),
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

	rcv, err := attrS.processEvent(context.TODO(), attrPrf.Tenant, ev, eNM, newDynamicDP(context.TODO(), nil, nil, nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(utils.ToJSON(exp), utils.ToJSON(rcv)) {
		t.Errorf("Expected \n<%+v>,\n Received \n<%+v>", utils.ToJSON(exp), utils.ToJSON(rcv))
	}

}

func TestV1GetAttributeForEventAttrProfEventErr(t *testing.T) {

	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	conMng := NewConnManager(cfg)
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), conMng)
	filterS := NewFilterS(cfg, conMng, dm)
	attr := &AttributeProfile{
		Tenant: "cgrates.org",
		ID:     "AttributeProfile1",
		Attributes: []*Attribute{
			{
				Path:  "*tenant",
				Type:  "*variable",
				Value: config.NewRSRParsersMustCompile("~*req.Account:s/(.*)@(.*)/${1}.${2}/", utils.InfieldSep),
			},
		},

		Weights: utils.DynamicWeights{
			{
				FilterIDs: []string{"*stirng:~*req.Account:1001"},
				Weight:    20,
			},
		},
	}
	err := dm.SetAttributeProfile(context.Background(), attr, true)
	if err != nil {
		t.Error(err)
	}

	alS := NewAttributeService(dm, filterS, cfg)
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"Attribute": "AttributeProfile1",
		},
		APIOpts: map[string]any{},
	}
	var rply APIAttributeProfile
	expErr := "SERVER_ERROR: NOT_IMPLEMENTED:*stirng"
	err = alS.V1GetAttributeForEvent(context.Background(), ev, &rply)
	if err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%+v>, received error <%+v>", expErr, err)
	}

}

func TestAttributesV1ProcessEventFieldMissingErr(t *testing.T) {

	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.FilterSCfg().ResourceSConns = []string{}
	conMng := NewConnManager(cfg)
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, nil, conMng)
	filterS := NewFilterS(cfg, conMng, dm)
	attr := &AttributeProfile{
		Tenant: "cgrates.org",
		ID:     "ATTR_CHANGE_TENANT_FROM_USER",
		Attributes: []*Attribute{
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*variable",
				Value:     config.NewRSRParsersMustCompile("~*req.Account:s/(.*)@(.*)/${1}.${2}/", utils.InfieldSep),
			},
		},
		Blockers: utils.DynamicBlockers{{Blocker: false}},
		Weights:  utils.DynamicWeights{{Weight: 20}},
	}
	err := dm.SetAttributeProfile(context.Background(), attr, true)
	if err != nil {
		t.Error(err)
	}

	alS := NewAttributeService(dm, filterS, cfg)
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
	_, err := ParseAttribute(dp, utils.MetaUsageDifference, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.UnixTimeStamp;~*req.usage2", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString)
	errExp := "Unsupported time format"
	if err == nil || err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}
}

func TestParseAtributeMetaPrefixParseDPErr(t *testing.T) {
	dp := utils.MapStorage{}

	_, err := ParseAttribute(dp, utils.MetaPrefix, "constant;`>;q=0.7;expires=3600`;~*req.Account", config.NewRSRParsersMustCompile("~*req.UnixTimeStamp;~*req.usage2", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString)
	if err != utils.ErrNotFound {
		t.Errorf("Expected %v\n but received %v", utils.ErrNotFound, err)
	}
}

func TestParseAtributeMetaSuffixParseDPErr(t *testing.T) {
	dp := utils.MapStorage{}

	_, err := ParseAttribute(dp, utils.MetaSuffix, "constant;`>;q=0.7;expires=3600`;~*req.Account", config.NewRSRParsersMustCompile("~*req.UnixTimeStamp;~*req.usage2", utils.InfieldSep),
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
	out, err := ParseAttribute(dp, utils.MetaCCUsage, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.cc1;~*req.cc2;~*req.cc3", utils.InfieldSep),
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
	expAttrPrf1 := &AttributeProfile{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     attrType + ":*req.Category:*attributes",
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Category",
				Type:  attrType,
				Value: config.NewRSRParsersMustCompile("*attributes", utils.InfieldSep),
			},
		},
	}
	attrPrf, err := NewAttributeFromInline(config.CgrConfig().GeneralCfg().DefaultTenant, attrID)
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
