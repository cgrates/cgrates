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
	"time"

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
		0, utils.EmptyString, utils.EmptyString, utils.InfieldSep)
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
		0, utils.EmptyString, utils.EmptyString, utils.InfieldSep)
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
		0, utils.EmptyString, utils.EmptyString, utils.InfieldSep)
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
		0, utils.EmptyString, utils.EmptyString, utils.InfieldSep)
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
		0, utils.EmptyString, utils.EmptyString, utils.InfieldSep)
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
		0, utils.EmptyString, utils.EmptyString, utils.InfieldSep)
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
		0, utils.EmptyString, utils.EmptyString, utils.InfieldSep)
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
		0, utils.EmptyString, utils.EmptyString, utils.InfieldSep)
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
		0, utils.EmptyString, utils.EmptyString, utils.InfieldSep)
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
		0, utils.EmptyString, utils.EmptyString, utils.InfieldSep)
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
		0, utils.EmptyString, utils.EmptyString, utils.InfieldSep)
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
		0, utils.EmptyString, utils.EmptyString, utils.InfieldSep)
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
		0, utils.EmptyString, utils.EmptyString, utils.InfieldSep)
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
		0, utils.EmptyString, utils.EmptyString, utils.InfieldSep)
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
		0, utils.EmptyString, utils.EmptyString, utils.InfieldSep)
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
		0, utils.EmptyString, utils.EmptyString, utils.InfieldSep)
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
		0, utils.EmptyString, utils.EmptyString, utils.InfieldSep)
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
		0, utils.EmptyString, utils.EmptyString, utils.InfieldSep)
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
		0, utils.EmptyString, utils.EmptyString, utils.InfieldSep)
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
		0, utils.EmptyString, utils.EmptyString, utils.InfieldSep)
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
		0, utils.EmptyString, utils.EmptyString, utils.InfieldSep)
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
		0, utils.EmptyString, utils.EmptyString, utils.InfieldSep)
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
		0, utils.EmptyString, utils.EmptyString, utils.InfieldSep)
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
		0, utils.EmptyString, utils.EmptyString, utils.InfieldSep)
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
		0, utils.EmptyString, utils.EmptyString, utils.InfieldSep)
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
		0, utils.EmptyString, utils.EmptyString, utils.InfieldSep)
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
