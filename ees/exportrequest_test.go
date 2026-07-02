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

package ees

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestExportRequestParseFieldDateTimeDaily(t *testing.T) {
	cacheS := engine.NewCacheS(config.CgrConfig(), nil, nil, nil)
	EventReq := NewExportRequest(map[string]utils.DataStorage{}, "", cacheS, nil, nil, config.CgrConfig().GeneralCfg().RoundingDecimals, config.CgrConfig().GeneralCfg().DefaultTimezone)
	fctTemp := &config.FCTemplate{
		Type:     utils.MetaDateTime,
		Value:    utils.NewRSRParsersMustCompile("*daily", utils.InfieldSep),
		Layout:   "“Mon Jan _2 15:04:05 2006”",
		Timezone: "",
	}

	result, err := EventReq.ParseField(fctTemp)
	if err != nil {
		t.Errorf("Expected %v but received %v", nil, err)
	}

	expected, err := utils.ParseTimeDetectLayout("*daily", utils.FirstNonEmpty(fctTemp.Timezone, config.CgrConfig().GeneralCfg().DefaultTimezone))
	if err != nil {
		t.Errorf("Expected %v but received %v", nil, err)
	}
	strRes := fmt.Sprintf("%v", result)
	finRes, err := time.Parse("“Mon Jan _2 15:04:05 2006”", strRes)
	if err != nil {
		t.Errorf("Expected %v but received %v", nil, err)
	}
	if !reflect.DeepEqual(finRes.Day(), expected.Day()) {
		t.Errorf("Expected %v but received %v", expected, result)
	}
}

func TestExportReqParseFieldDateTimeTimeZone(t *testing.T) {
	cacheS := engine.NewCacheS(config.CgrConfig(), nil, nil, nil)
	EventReq := NewExportRequest(map[string]utils.DataStorage{}, "", cacheS, nil, nil, config.CgrConfig().GeneralCfg().RoundingDecimals, config.CgrConfig().GeneralCfg().DefaultTimezone)
	fctTemp := &config.FCTemplate{
		Type:     utils.MetaDateTime,
		Value:    utils.NewRSRParsersMustCompile("*daily", utils.InfieldSep),
		Layout:   "“Mon Jan _2 15:04:05 2006”",
		Timezone: "Local",
	}

	result, err := EventReq.ParseField(fctTemp)
	if err != nil {
		t.Errorf("Expected %v but received %v", nil, err)
	}

	expected, err := utils.ParseTimeDetectLayout("*daily", utils.FirstNonEmpty(fctTemp.Timezone, config.CgrConfig().GeneralCfg().DefaultTimezone))
	if err != nil {
		t.Errorf("Expected %v but received %v", nil, err)
	}
	strRes := fmt.Sprintf("%v", result)
	finRes, err := time.Parse("“Mon Jan _2 15:04:05 2006”", strRes)
	if err != nil {
		t.Errorf("Expected %v but received %v", nil, err)
	}
	if !reflect.DeepEqual(finRes.Day(), expected.Day()) {
		t.Errorf("Expected %v but received %v", finRes.Day(), expected.Day())
	}
}

func TestExportReqParseFieldDateTimeMonthly(t *testing.T) {
	cacheS := engine.NewCacheS(config.CgrConfig(), nil, nil, nil)
	EventReq := NewExportRequest(map[string]utils.DataStorage{}, "", cacheS, nil, nil, config.CgrConfig().GeneralCfg().RoundingDecimals, config.CgrConfig().GeneralCfg().DefaultTimezone)
	fctTemp := &config.FCTemplate{
		Type:     utils.MetaDateTime,
		Value:    utils.NewRSRParsersMustCompile("*monthly", utils.InfieldSep),
		Layout:   "“Mon Jan _2 15:04:05 2006”",
		Timezone: "Local",
	}
	result, err := EventReq.ParseField(fctTemp)
	if err != nil {
		t.Errorf("Expected %v but received %v", nil, err)
	}

	expected, err := utils.ParseTimeDetectLayout("*monthly", utils.FirstNonEmpty(fctTemp.Timezone, config.CgrConfig().GeneralCfg().DefaultTimezone))
	if err != nil {
		t.Errorf("Expected %v but received %v", nil, err)
	}
	strRes := fmt.Sprintf("%v", result)
	finRes, err := time.Parse("“Mon Jan _2 15:04:05 2006”", strRes)
	if err != nil {
		t.Errorf("Expected %v but received %v", nil, err)
	}
	if !reflect.DeepEqual(finRes.Month(), expected.Month()) {
		t.Errorf("Expected %v but received %v", finRes.Month(), expected.Month())
	}
}

func TestExportReqParseFieldDateTimeMonthlyEstimated(t *testing.T) {
	cacheS := engine.NewCacheS(config.CgrConfig(), nil, nil, nil)
	EventReq := NewExportRequest(map[string]utils.DataStorage{}, "", cacheS, nil, nil, config.CgrConfig().GeneralCfg().RoundingDecimals, config.CgrConfig().GeneralCfg().DefaultTimezone)
	fctTemp := &config.FCTemplate{
		Type:     utils.MetaDateTime,
		Value:    utils.NewRSRParsersMustCompile("*monthlyEstimated", utils.InfieldSep),
		Layout:   "“Mon Jan _2 15:04:05 2006”",
		Timezone: "Local",
	}
	result, err := EventReq.ParseField(fctTemp)
	if err != nil {
		t.Errorf("Expected %v but received %v", nil, err)
	}

	expected, err := utils.ParseTimeDetectLayout("*monthlyEstimated", utils.FirstNonEmpty(fctTemp.Timezone, config.CgrConfig().GeneralCfg().DefaultTimezone))
	if err != nil {
		t.Errorf("Expected %v but received %v", nil, err)
	}
	strRes := fmt.Sprintf("%v", result)
	finRes, err := time.Parse("“Mon Jan _2 15:04:05 2006”", strRes)
	if err != nil {
		t.Errorf("Expected %v but received %v", nil, err)
	}
	if !reflect.DeepEqual(finRes.Month(), expected.Month()) {
		t.Errorf("Expected %v but received %v", finRes.Month(), expected.Month())
	}
}

func TestExportReqParseFieldDateTimeYearly(t *testing.T) {
	cacheS := engine.NewCacheS(config.CgrConfig(), nil, nil, nil)
	EventReq := NewExportRequest(map[string]utils.DataStorage{}, "", cacheS, nil, nil, config.CgrConfig().GeneralCfg().RoundingDecimals, config.CgrConfig().GeneralCfg().DefaultTimezone)
	fctTemp := &config.FCTemplate{
		Type:     utils.MetaDateTime,
		Value:    utils.NewRSRParsersMustCompile("*yearly", utils.InfieldSep),
		Layout:   "“Mon Jan _2 15:04:05 2006”",
		Timezone: "Local",
	}
	result, err := EventReq.ParseField(fctTemp)
	if err != nil {
		t.Errorf("Expected %v but received %v", nil, err)
	}

	expected, err := utils.ParseTimeDetectLayout("*yearly", utils.FirstNonEmpty(fctTemp.Timezone, config.CgrConfig().GeneralCfg().DefaultTimezone))
	if err != nil {
		t.Errorf("Expected %v but received %v", nil, err)
	}
	strRes := fmt.Sprintf("%v", result)
	finRes, err := time.Parse("“Mon Jan _2 15:04:05 2006”", strRes)
	if err != nil {
		t.Errorf("Expected %v but received %v", nil, err)
	}
	if !reflect.DeepEqual(finRes.Year(), expected.Year()) {
		t.Errorf("Expected %v but received %v", finRes.Year(), expected.Year())
	}
}

func TestExportReqParseFieldDateTimeMetaUnlimited(t *testing.T) {
	cacheS := engine.NewCacheS(config.CgrConfig(), nil, nil, nil)
	EventReq := NewExportRequest(map[string]utils.DataStorage{}, "", cacheS, nil, nil, config.CgrConfig().GeneralCfg().RoundingDecimals, config.CgrConfig().GeneralCfg().DefaultTimezone)
	fctTemp := &config.FCTemplate{
		Type:     utils.MetaDateTime,
		Value:    utils.NewRSRParsersMustCompile(utils.MetaUnlimited, utils.InfieldSep),
		Layout:   "“Mon Jan _2 15:04:05 2006”",
		Timezone: "Local",
	}
	result, err := EventReq.ParseField(fctTemp)
	if err != nil {
		t.Errorf("Expected %v but received %v", nil, err)
	}

	expected, err := utils.ParseTimeDetectLayout(utils.MetaUnlimited, utils.FirstNonEmpty(fctTemp.Timezone, config.CgrConfig().GeneralCfg().DefaultTimezone))
	if err != nil {
		t.Errorf("Expected %v but received %v", nil, err)
	}
	strRes := fmt.Sprintf("%v", result)
	finRes, err := time.Parse("“Mon Jan _2 15:04:05 2006”", strRes)
	if err != nil {
		t.Errorf("Expected %v but received %v", nil, err)
	}
	if !reflect.DeepEqual(finRes.Day(), expected.Day()) {
		t.Errorf("Expected %v but received %v", finRes.Day(), expected.Day())
	}
}

func TestExportReqParseFieldDateTimeEmpty(t *testing.T) {
	cacheS := engine.NewCacheS(config.CgrConfig(), nil, nil, nil)
	EventReq := NewExportRequest(map[string]utils.DataStorage{}, "", cacheS, nil, nil, config.CgrConfig().GeneralCfg().RoundingDecimals, config.CgrConfig().GeneralCfg().DefaultTimezone)
	fctTemp := &config.FCTemplate{
		Type:     utils.MetaDateTime,
		Value:    utils.NewRSRParsersMustCompile("", utils.InfieldSep),
		Layout:   "“Mon Jan _2 15:04:05 2006”",
		Timezone: "Local",
	}
	result, err := EventReq.ParseField(fctTemp)
	if err != nil {
		t.Errorf("Expected %v but received %v", nil, err)
	}

	expected, err := utils.ParseTimeDetectLayout("", utils.FirstNonEmpty(fctTemp.Timezone, config.CgrConfig().GeneralCfg().DefaultTimezone))
	if err != nil {
		t.Errorf("Expected %v but received %v", nil, err)
	}
	strRes := fmt.Sprintf("%v", result)
	finRes, err := time.Parse("“Mon Jan _2 15:04:05 2006”", strRes)
	if err != nil {
		t.Errorf("Expected %v but received %v", nil, err)
	}
	if !reflect.DeepEqual(finRes.Day(), expected.Day()) {
		t.Errorf("Expected %v but received %v", finRes.Day(), expected.Day())
	}
}

func TestExportReqParseFieldDateTimeMonthEnd(t *testing.T) {
	cacheS := engine.NewCacheS(config.CgrConfig(), nil, nil, nil)
	EventReq := NewExportRequest(map[string]utils.DataStorage{}, "", cacheS, nil, nil, config.CgrConfig().GeneralCfg().RoundingDecimals, config.CgrConfig().GeneralCfg().DefaultTimezone)
	fctTemp := &config.FCTemplate{
		Type:     utils.MetaDateTime,
		Value:    utils.NewRSRParsersMustCompile("*month_endTest", utils.InfieldSep),
		Layout:   "“Mon Jan _2 15:04:05 2006”",
		Timezone: "Local",
	}
	result, err := EventReq.ParseField(fctTemp)
	if err != nil {
		t.Errorf("Expected %v but received %v", nil, err)
	}

	expected, err := utils.ParseTimeDetectLayout("*month_endTest", utils.FirstNonEmpty(fctTemp.Timezone, config.CgrConfig().GeneralCfg().DefaultTimezone))
	if err != nil {
		t.Errorf("Expected %v but received %v", nil, err)
	}
	strRes := fmt.Sprintf("%v", result)
	finRes, err := time.Parse("“Mon Jan _2 15:04:05 2006”", strRes)
	if err != nil {
		t.Errorf("Expected %v but received %v", nil, err)
	}
	if !reflect.DeepEqual(finRes.Day(), expected.Day()) {
		t.Errorf("Expected %v but received %v", finRes.Day(), expected.Day())
	}
}

func TestExportReqParseFieldDateTimeError(t *testing.T) {
	cacheS := engine.NewCacheS(config.CgrConfig(), nil, nil, nil)
	EventReq := NewExportRequest(map[string]utils.DataStorage{}, "", cacheS, nil, nil, config.CgrConfig().GeneralCfg().RoundingDecimals, config.CgrConfig().GeneralCfg().DefaultTimezone)
	fctTemp := &config.FCTemplate{
		Type:     utils.MetaDateTime,
		Value:    utils.NewRSRParsersMustCompile("*month_endTest", utils.InfieldSep),
		Layout:   "“Mon Jan _2 15:04:05 2006”",
		Timezone: "/",
	}
	_, err := EventReq.ParseField(fctTemp)
	expected := "time: invalid location name"
	if err == nil || err.Error() != expected {
		t.Errorf("Expected <%+v> but received <%+v>", expected, err)
	}
}

func TestExportReqParseFieldDateTimeError2(t *testing.T) {
	prsr, err := utils.NewRSRParsersFromSlice([]string{"2.", "~*opts.*originID<~*opts.Converter>"})
	if err != nil {
		t.Fatal(err)
	}
	mS := map[string]utils.DataStorage{
		utils.MetaOpts: utils.MapStorage{
			utils.AccountField: "1002",
			utils.Usage:        "20m",
		},
	}
	cacheS := engine.NewCacheS(config.CgrConfig(), nil, nil, nil)
	EventReq := NewExportRequest(mS, "", cacheS, nil, nil, config.CgrConfig().GeneralCfg().RoundingDecimals, config.CgrConfig().GeneralCfg().DefaultTimezone)
	fctTemp := &config.FCTemplate{
		Type:     utils.MetaDateTime,
		Value:    prsr,
		Layout:   "“Mon Jan _2 15:04:05 2006”",
		Timezone: "/",
	}
	expected := utils.ErrNotFound
	if _, err = EventReq.ParseField(fctTemp); err == nil || err != expected {
		t.Errorf("Expected <%+v> but received <%+v>", expected, err)
	}
}

func TestExportReqFieldAsInterface(t *testing.T) {
	inData := map[string]utils.DataStorage{
		utils.MetaReq: utils.MapStorage{
			"Account": "1001",
			"Usage":   "10m",
		},
	}
	cacheS := engine.NewCacheS(config.CgrConfig(), nil, nil, nil)
	eventReq := NewExportRequest(inData, "cgrates.org", cacheS, nil, nil, config.CgrConfig().GeneralCfg().RoundingDecimals, config.CgrConfig().GeneralCfg().DefaultTimezone)
	fldPath := []string{utils.MetaReq, "Usage"}
	expVal := "10m"
	if rcv, err := eventReq.FieldAsInterface(fldPath); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, expVal) {
		t.Errorf("Expected %+v \n but received \n %+v", expVal, rcv)
	}

	expVal = "cgrates.org"
	fldPath = []string{utils.MetaTenant}
	if rcv, err := eventReq.FieldAsInterface(fldPath); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, expVal) {
		t.Errorf("Expected %+v \n but received \n %+v", expVal, rcv)
	}
}

func TestExportReqNewEventExporter(t *testing.T) {
	inData := map[string]utils.DataStorage{
		utils.MetaReq: utils.MapStorage{
			"Account": "1001",
			"Usage":   "10m",
		},
	}
	onm := utils.NewOrderedNavigableMap()
	fullPath := &utils.FullPath{
		PathSlice: []string{utils.MetaReq, utils.MetaTenant},
		Path:      utils.MetaTenant,
	}
	val := &utils.DataLeaf{
		Data: "value1",
	}
	onm.Append(fullPath, val)
	expData := map[string]*utils.OrderedNavigableMap{
		utils.MetaReq: onm,
	}
	cacheS := engine.NewCacheS(config.CgrConfig(), nil, nil, nil)
	expected := &ExportRequest{
		inData:           inData,
		filterS:          nil,
		tnt:              "cgrates.org",
		cache:            cacheS,
		ExpData:          expData,
		roundingDecimals: config.CgrConfig().GeneralCfg().RoundingDecimals,
		timezone:         config.CgrConfig().GeneralCfg().DefaultTimezone,
	}
	eventReq := NewExportRequest(inData, "cgrates.org", cacheS, nil, expData, config.CgrConfig().GeneralCfg().RoundingDecimals, config.CgrConfig().GeneralCfg().DefaultTimezone)
	if !reflect.DeepEqual(expected, eventReq) {
		t.Errorf("Expected %v \n but received \n %v", expected, eventReq)
	}
}

func TestExportRequestString(t *testing.T) {
	inData := map[string]utils.DataStorage{
		utils.MetaReq: utils.MapStorage{
			"Account": "1001",
			"Usage":   "10m",
		},
	}
	onm := utils.NewOrderedNavigableMap()

	expData := map[string]*utils.OrderedNavigableMap{
		utils.MetaReq: onm,
	}

	cacheS := engine.NewCacheS(config.CgrConfig(), nil, nil, nil)
	eeR := NewExportRequest(inData, "cgrates.org", cacheS, nil, expData, config.CgrConfig().GeneralCfg().RoundingDecimals, config.CgrConfig().GeneralCfg().DefaultTimezone)

	exp := utils.ToIJSON(eeR)

	if rcv := eeR.String(); rcv != exp {
		t.Error(rcv)
	}

}

func TestExportReqFieldAsInterfaceBadPrefix(t *testing.T) {
	inData := map[string]utils.DataStorage{
		utils.MetaReq: utils.MapStorage{
			"Account": "1001",
			"Usage":   "10m",
		},
	}
	cacheS := engine.NewCacheS(config.CgrConfig(), nil, nil, nil)
	eventReq := NewExportRequest(inData, "cgrates.org", cacheS, nil, nil, config.CgrConfig().GeneralCfg().RoundingDecimals, config.CgrConfig().GeneralCfg().DefaultTimezone)

	fldPath := []string{"inexistant"}
	expErr := "unsupported field prefix: <inexistant>"
	if _, err := eventReq.FieldAsInterface(fldPath); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%+v>, received error <%+v>", expErr, err)
	}
}

func TestExportReqFieldAsInterfaceMetaUCHErr(t *testing.T) {

	inData := map[string]utils.DataStorage{
		utils.MetaReq: utils.MapStorage{
			"Account": "1001",
			"Usage":   "10m",
		},
	}
	cacheS := engine.NewCacheS(config.CgrConfig(), nil, nil, nil)
	eventReq := NewExportRequest(inData, "cgrates.org", cacheS, nil, nil, config.CgrConfig().GeneralCfg().RoundingDecimals, config.CgrConfig().GeneralCfg().DefaultTimezone)

	fldPath := []string{utils.MetaUCH}
	if _, err := eventReq.FieldAsInterface(fldPath); err != utils.ErrNotFound {
		t.Errorf("Expected error <%+v>, received error <%+v>", utils.ErrNotFound, err)
	}
}

func TestExportReqFieldAsInterfaceNMSliceType(t *testing.T) {

	inData := map[string]utils.DataStorage{
		utils.MetaReq: utils.MapStorage{
			"Slice": &utils.DataNode{
				Type: utils.NMSliceType,
				Slice: []*utils.DataNode{
					{
						Type: utils.NMDataType,
						Value: &utils.DataLeaf{
							Data: "cgrates.org",
						},
					},
				},
			},
		},
	}

	cacheS := engine.NewCacheS(config.CgrConfig(), nil, nil, nil)
	eventReq := NewExportRequest(inData, "cgrates.org", cacheS, nil, nil, config.CgrConfig().GeneralCfg().RoundingDecimals, config.CgrConfig().GeneralCfg().DefaultTimezone)
	fldPath := []string{utils.MetaReq, "Slice"}
	expVal := "cgrates.org"
	if rcv, err := eventReq.FieldAsInterface(fldPath); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, expVal) {
		t.Errorf("Expected %+v \n but received \n %+v", expVal, rcv)
	}

}

func TestExportReqFieldAsStringOK(t *testing.T) {
	inData := map[string]utils.DataStorage{
		utils.MetaReq: utils.MapStorage{
			"Account": "1001",
			"Usage":   "10m",
		},
	}
	cacheS := engine.NewCacheS(config.CgrConfig(), nil, nil, nil)
	eventReq := NewExportRequest(inData, "cgrates.org", cacheS, nil, nil, config.CgrConfig().GeneralCfg().RoundingDecimals, config.CgrConfig().GeneralCfg().DefaultTimezone)
	fldPath := []string{utils.MetaReq, "Usage"}
	expVal := "10m"
	if rcv, err := eventReq.FieldAsString(fldPath); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, expVal) {
		t.Errorf("Expected %+v \n but received \n %+v", expVal, rcv)
	}
}

func TestExportRequestParseFieldMetaFiller(t *testing.T) {
	cacheS := engine.NewCacheS(config.CgrConfig(), nil, nil, nil)
	EventReq := NewExportRequest(map[string]utils.DataStorage{}, "", cacheS, nil, nil, config.CgrConfig().GeneralCfg().RoundingDecimals, config.CgrConfig().GeneralCfg().DefaultTimezone)
	fctTemp := &config.FCTemplate{
		Type:     utils.MetaFiller,
		Value:    utils.NewRSRParsersMustCompile("*daily", utils.InfieldSep),
		Layout:   "“Mon Jan _2 15:04:05 2006”",
		Timezone: "",
	}

	rcv, err := EventReq.ParseField(fctTemp)
	if err != nil {
		t.Error(err)
	} else if rcv != utils.MetaDaily {
		t.Errorf("Expected %v but received %v", utils.MetaDaily, rcv)
	}

}

func TestExportRequestParseFieldMetaGroup(t *testing.T) {
	cacheS := engine.NewCacheS(config.CgrConfig(), nil, nil, nil)
	EventReq := NewExportRequest(map[string]utils.DataStorage{}, "", cacheS, nil, nil, config.CgrConfig().GeneralCfg().RoundingDecimals, config.CgrConfig().GeneralCfg().DefaultTimezone)
	fctTemp := &config.FCTemplate{
		Type:     utils.MetaGroup,
		Value:    utils.NewRSRParsersMustCompile("*daily", utils.InfieldSep),
		Layout:   "“Mon Jan _2 15:04:05 2006”",
		Timezone: "",
	}

	rcv, err := EventReq.ParseField(fctTemp)
	if err != nil {
		t.Error(err)
	} else if rcv != utils.MetaDaily {
		t.Errorf("Expected %v but received %v", utils.MetaDaily, rcv)
	}

}

func TestExportRequestSetAsSliceMetaUCH(t *testing.T) {
	cacheS := engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)

	inData := map[string]utils.DataStorage{
		utils.MetaReq: utils.MapStorage{
			"Account": "1001",
			"Usage":   "10m",
		},
	}
	onm := utils.NewOrderedNavigableMap()

	expData := map[string]*utils.OrderedNavigableMap{
		utils.MetaReq: onm,
	}

	eeR := NewExportRequest(inData, "cgrates.org", cacheS, nil, expData, config.CgrConfig().GeneralCfg().RoundingDecimals, config.CgrConfig().GeneralCfg().DefaultTimezone)

	fullPath := &utils.FullPath{
		PathSlice: []string{utils.MetaUCH},
		Path:      "*uch;Tenant",
	}
	val := &utils.DataLeaf{
		Data: "cgrates.org",
	}
	if err := eeR.SetAsSlice(fullPath, val); err != nil {
		t.Errorf("Expected error <%v> but received <%v>", nil, err)
	}

	if rcv, ok := cacheS.Get(utils.CacheUCH, "Tenant"); !ok {
		t.Error("Couldnt receive from cache")
	} else if rcv != "cgrates.org" {
		t.Errorf("Expected \n<%v>,\n but received \n<%v>", "cgrates.org", rcv)
	}

}

func TestExportRequestSetAsSliceMetaOpts(t *testing.T) {

	inData := map[string]utils.DataStorage{
		utils.MetaOpts: utils.MapStorage{},
	}
	onm := utils.NewOrderedNavigableMap()

	expData := map[string]*utils.OrderedNavigableMap{
		utils.MetaReq: onm,
	}

	cacheS := engine.NewCacheS(config.CgrConfig(), nil, nil, nil)
	eeR := NewExportRequest(inData, "cgrates.org", cacheS, nil, expData, config.CgrConfig().GeneralCfg().RoundingDecimals, config.CgrConfig().GeneralCfg().DefaultTimezone)

	fullPath := &utils.FullPath{
		PathSlice: []string{utils.MetaOpts, "Tenant"},
	}
	val := &utils.DataLeaf{
		Data: "cgrates.org",
	}
	if err := eeR.SetAsSlice(fullPath, val); err != nil {
		t.Errorf("Expected error <%v> but received <%v>", nil, err)
	}

	exp := map[string]utils.DataStorage{
		utils.MetaOpts: utils.MapStorage{
			"Tenant": "cgrates.org",
		},
	}

	if !reflect.DeepEqual(eeR.inData[utils.MetaOpts], exp[utils.MetaOpts]) {
		t.Errorf("Expected \n<%v>,\n but received \n<%v>", exp, eeR.inData[utils.MetaOpts])
	}

}

func TestExportRequestSetAsSliceExpDataErr(t *testing.T) {

	inData := map[string]utils.DataStorage{
		utils.MetaOpts: utils.MapStorage{},
	}
	onm := utils.NewOrderedNavigableMap()

	expData := map[string]*utils.OrderedNavigableMap{
		utils.MetaReq: onm,
	}

	cacheS := engine.NewCacheS(config.CgrConfig(), nil, nil, nil)
	eeR := NewExportRequest(inData, "cgrates.org", cacheS, nil, expData, config.CgrConfig().GeneralCfg().RoundingDecimals, config.CgrConfig().GeneralCfg().DefaultTimezone)

	fullPath := &utils.FullPath{
		PathSlice: []string{"Inexistant field"},
	}
	val := &utils.DataLeaf{
		Data: "cgrates.org",
	}
	expErr := "unsupported field prefix: <Inexistant field> when set field"
	if err := eeR.SetAsSlice(fullPath, val); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v> but received <%v>", expErr, err)
	}

}

func TestExportRequestSetAsSliceDefaultOK(t *testing.T) {

	inData := map[string]utils.DataStorage{
		utils.MetaOpts: utils.MapStorage{},
	}
	onm := utils.NewOrderedNavigableMap()

	expData := map[string]*utils.OrderedNavigableMap{
		utils.MetaExp: onm,
	}

	cacheS := engine.NewCacheS(config.CgrConfig(), nil, nil, nil)
	eeR := NewExportRequest(inData, "cgrates.org", cacheS, nil, expData, config.CgrConfig().GeneralCfg().RoundingDecimals, config.CgrConfig().GeneralCfg().DefaultTimezone)

	fullPath := &utils.FullPath{
		PathSlice: []string{utils.MetaExp, "Tenant"},
		Path:      "*uch.Tenant",
	}

	val := &utils.DataLeaf{
		Data: "cgrates.org",
	}
	exp := `{
 "Map": {
  "Tenant": {
   "Value": {
    "Data": "cgrates.org"
   }
  }
 }
}`

	if err := eeR.SetAsSlice(fullPath, val); err != nil {
		t.Error(err)
	} else if eeR.ExpData[utils.MetaExp].String() != exp {
		t.Errorf("Expected \n<%v>,\n Received <%v>", exp, eeR.ExpData[utils.MetaExp].String())
	}

}

func TestExportRequestAppendMetaUCH(t *testing.T) {
	cacheS := engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)

	inData := map[string]utils.DataStorage{
		utils.MetaReq: utils.MapStorage{
			"Account": "1001",
			"Usage":   "10m",
		},
	}
	onm := utils.NewOrderedNavigableMap()

	expData := map[string]*utils.OrderedNavigableMap{
		utils.MetaReq: onm,
	}

	eeR := NewExportRequest(inData, "cgrates.org", cacheS, nil, expData, config.CgrConfig().GeneralCfg().RoundingDecimals, config.CgrConfig().GeneralCfg().DefaultTimezone)

	fullPath := &utils.FullPath{
		PathSlice: []string{utils.MetaUCH},
		Path:      "*uch;Tenant",
	}
	val := &utils.DataLeaf{
		Data: "cgrates.org",
	}
	if err := eeR.Append(fullPath, val); err != nil {
		t.Errorf("Expected error <%v> but received <%v>", nil, err)
	}

	if rcv, ok := cacheS.Get(utils.CacheUCH, "Tenant"); !ok {
		t.Error("Couldnt receive from cache")
	} else if rcv != "cgrates.org" {
		t.Errorf("Expected \n<%v>,\n but received \n<%v>", "cgrates.org", rcv)
	}

}

func TestExportRequestAppendMetaOpts(t *testing.T) {

	inData := map[string]utils.DataStorage{
		utils.MetaOpts: utils.MapStorage{},
	}
	onm := utils.NewOrderedNavigableMap()

	expData := map[string]*utils.OrderedNavigableMap{
		utils.MetaReq: onm,
	}

	cacheS := engine.NewCacheS(config.CgrConfig(), nil, nil, nil)
	eeR := NewExportRequest(inData, "cgrates.org", cacheS, nil, expData, config.CgrConfig().GeneralCfg().RoundingDecimals, config.CgrConfig().GeneralCfg().DefaultTimezone)

	fullPath := &utils.FullPath{
		PathSlice: []string{utils.MetaOpts, "Tenant"},
	}
	val := &utils.DataLeaf{
		Data: "cgrates.org",
	}
	if err := eeR.Append(fullPath, val); err != nil {
		t.Errorf("Expected error <%v> but received <%v>", nil, err)
	}

	exp := map[string]utils.DataStorage{
		utils.MetaOpts: utils.MapStorage{
			"Tenant": "cgrates.org",
		},
	}

	if !reflect.DeepEqual(eeR.inData[utils.MetaOpts], exp[utils.MetaOpts]) {
		t.Errorf("Expected \n<%v>,\n but received \n<%v>", exp, eeR.inData[utils.MetaOpts])
	}

}

func TestExportRequestAppendExpDataErr(t *testing.T) {

	inData := map[string]utils.DataStorage{
		utils.MetaOpts: utils.MapStorage{},
	}
	onm := utils.NewOrderedNavigableMap()

	expData := map[string]*utils.OrderedNavigableMap{
		utils.MetaReq: onm,
	}

	cacheS := engine.NewCacheS(config.CgrConfig(), nil, nil, nil)
	eeR := NewExportRequest(inData, "cgrates.org", cacheS, nil, expData, config.CgrConfig().GeneralCfg().RoundingDecimals, config.CgrConfig().GeneralCfg().DefaultTimezone)

	fullPath := &utils.FullPath{
		PathSlice: []string{"Inexistant field"},
	}
	val := &utils.DataLeaf{
		Data: "cgrates.org",
	}
	expErr := "unsupported field prefix: <Inexistant field> when set field"
	if err := eeR.Append(fullPath, val); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v> but received <%v>", expErr, err)
	}

}

func TestExportRequestAppendDefaultOK(t *testing.T) {

	inData := map[string]utils.DataStorage{
		utils.MetaOpts: utils.MapStorage{},
	}
	onm := utils.NewOrderedNavigableMap()

	expData := map[string]*utils.OrderedNavigableMap{
		utils.MetaExp: onm,
	}

	cacheS := engine.NewCacheS(config.CgrConfig(), nil, nil, nil)
	eeR := NewExportRequest(inData, "cgrates.org", cacheS, nil, expData, config.CgrConfig().GeneralCfg().RoundingDecimals, config.CgrConfig().GeneralCfg().DefaultTimezone)

	fullPath := &utils.FullPath{
		PathSlice: []string{utils.MetaExp, "Tenant"},
		Path:      "*uch.Tenant",
	}

	val := &utils.DataLeaf{
		Data: "cgrates.org",
	}
	exp := `{
 "Map": {
  "Tenant": {
   "Slice": [
    {
     "Value": {
      "Data": "cgrates.org"
     }
    }
   ]
  }
 }
}`

	if err := eeR.Append(fullPath, val); err != nil {
		t.Error(err)
	} else if eeR.ExpData[utils.MetaExp].String() != exp {
		t.Errorf("Expected \n<%v>,\n Received <%v>", exp, eeR.ExpData[utils.MetaExp].String())
	}

}

func TestExportRequestComposeMetaUCHNotOK(t *testing.T) {
	cacheS := engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)

	inData := map[string]utils.DataStorage{
		utils.MetaReq: utils.MapStorage{
			"Account": "1001",
			"Usage":   "10m",
		},
	}
	onm := utils.NewOrderedNavigableMap()

	expData := map[string]*utils.OrderedNavigableMap{
		utils.MetaReq: onm,
	}

	eeR := NewExportRequest(inData, "cgrates.org", cacheS, nil, expData, config.CgrConfig().GeneralCfg().RoundingDecimals, config.CgrConfig().GeneralCfg().DefaultTimezone)

	fullPath := &utils.FullPath{
		PathSlice: []string{utils.MetaUCH},
		Path:      "*uch;Tenant",
	}
	val := &utils.DataLeaf{
		Data: "cgrates.org",
	}
	if err := eeR.Compose(fullPath, val); err != nil {
		t.Errorf("Expected error <%v> but received <%v>", nil, err)
	}

	if rcv, ok := cacheS.Get(utils.CacheUCH, "Tenant"); !ok {
		t.Error("Couldnt receive from cache")
	} else if rcv != "cgrates.org" {
		t.Errorf("Expected \n<%v>,\n but received \n<%v>", "cgrates.org", rcv)
	}

}

func TestExportRequestComposeMetaUCHPathSet(t *testing.T) {
	cacheS := engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)

	inData := map[string]utils.DataStorage{
		utils.MetaReq: utils.MapStorage{
			"Account": "1001",
			"Usage":   "10m",
		},
	}
	onm := utils.NewOrderedNavigableMap()

	expData := map[string]*utils.OrderedNavigableMap{
		utils.MetaReq: onm,
	}

	eeR := NewExportRequest(inData, "cgrates.org", cacheS, nil, expData, config.CgrConfig().GeneralCfg().RoundingDecimals, config.CgrConfig().GeneralCfg().DefaultTimezone)

	fullPath := &utils.FullPath{
		PathSlice: []string{utils.MetaUCH},
		Path:      "*uch;Tenant",
	}
	val := &utils.DataLeaf{
		Data: "cgrates.org",
	}

	if err := cacheS.Set(context.Background(), utils.CacheUCH, "Tenant", "Extra", []string{}, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	if err := eeR.Compose(fullPath, val); err != nil {
		t.Errorf("Expected error <%v> but received <%v>", nil, err)
	}

	if rcv, ok := cacheS.Get(utils.CacheUCH, "Tenant"); !ok {
		t.Error("Couldnt receive from cache")
	} else if rcv != "Extracgrates.org" {
		t.Errorf("Expected \n<%v>,\n but received \n<%v>", "cgrates.org", rcv)
	}

}

func TestExportRequestComposeMetaOptsOK(t *testing.T) {

	inData := map[string]utils.DataStorage{
		utils.MetaOpts: utils.MapStorage{},
	}
	onm := utils.NewOrderedNavigableMap()

	expData := map[string]*utils.OrderedNavigableMap{
		utils.MetaReq: onm,
	}

	cacheS := engine.NewCacheS(config.CgrConfig(), nil, nil, nil)
	eeR := NewExportRequest(inData, "cgrates.org", cacheS, nil, expData, config.CgrConfig().GeneralCfg().RoundingDecimals, config.CgrConfig().GeneralCfg().DefaultTimezone)

	fullPath := &utils.FullPath{
		PathSlice: []string{utils.MetaOpts, "Tenant"},
	}
	val := &utils.DataLeaf{
		Data: "cgrates.org",
	}
	if err := eeR.Compose(fullPath, val); err != nil {
		t.Errorf("Expected error <%v> but received <%v>", nil, err)
	}

	exp := map[string]utils.DataStorage{
		utils.MetaOpts: utils.MapStorage{
			"Tenant": "cgrates.org",
		},
	}

	if !reflect.DeepEqual(eeR.inData[utils.MetaOpts], exp[utils.MetaOpts]) {
		t.Errorf("Expected \n<%v>,\n but received \n<%v>", exp, eeR.inData[utils.MetaOpts])
	}

}

func TestExportRequestComposeMetaOptsFoundOK(t *testing.T) {

	inData := map[string]utils.DataStorage{
		utils.MetaOpts: utils.MapStorage{
			"Account": "1001",
			"Usage":   "10m",
			"Tenant":  "Extra",
		},
	}

	cacheS := engine.NewCacheS(config.CgrConfig(), nil, nil, nil)
	eeR := NewExportRequest(inData, "cgrates.org", cacheS, nil, nil, config.CgrConfig().GeneralCfg().RoundingDecimals, config.CgrConfig().GeneralCfg().DefaultTimezone)

	fullPath := &utils.FullPath{
		PathSlice: []string{utils.MetaOpts, "Tenant"},
	}
	val := &utils.DataLeaf{
		Data: "cgrates.org",
	}

	if err := eeR.Compose(fullPath, val); err != nil {
		t.Error(err)
	}
	exp := map[string]utils.DataStorage{
		utils.MetaOpts: utils.MapStorage{
			"Account": "1001",
			"Usage":   "10m",
			"Tenant":  "Extracgrates.org",
		},
	}

	if !reflect.DeepEqual(eeR.inData[utils.MetaOpts], exp[utils.MetaOpts]) {
		t.Errorf("Expected \n<%v>,\n but received \n<%v>", exp, eeR.inData[utils.MetaOpts])
	}

}

func TestExportRequestComposeDefaultOK(t *testing.T) {

	inData := map[string]utils.DataStorage{
		utils.MetaOpts: utils.MapStorage{},
	}
	onm := utils.NewOrderedNavigableMap()

	expData := map[string]*utils.OrderedNavigableMap{
		utils.MetaExp: onm,
	}

	cacheS := engine.NewCacheS(config.CgrConfig(), nil, nil, nil)
	eeR := NewExportRequest(inData, "cgrates.org", cacheS, nil, expData, config.CgrConfig().GeneralCfg().RoundingDecimals, config.CgrConfig().GeneralCfg().DefaultTimezone)

	fullPath := &utils.FullPath{
		PathSlice: []string{utils.MetaExp, "Tenant"},
		Path:      "*uch.Tenant",
	}

	val := &utils.DataLeaf{
		Data: "cgrates.org",
	}
	exp := `{
 "Map": {
  "Tenant": {
   "Value": {
    "Data": "cgrates.org"
   }
  }
 }
}`

	if err := eeR.Compose(fullPath, val); err != nil {
		t.Error(err)
	} else if eeR.ExpData[utils.MetaExp].String() != exp {
		t.Errorf("Expected \n<%v>,\n Received <%v>", exp, eeR.ExpData[utils.MetaExp].String())
	}

}
func TestExportRequestComposeExpDataErr(t *testing.T) {

	inData := map[string]utils.DataStorage{
		utils.MetaOpts: utils.MapStorage{},
	}
	onm := utils.NewOrderedNavigableMap()

	expData := map[string]*utils.OrderedNavigableMap{
		utils.MetaReq: onm,
	}

	cacheS := engine.NewCacheS(config.CgrConfig(), nil, nil, nil)
	eeR := NewExportRequest(inData, "cgrates.org", cacheS, nil, expData, config.CgrConfig().GeneralCfg().RoundingDecimals, config.CgrConfig().GeneralCfg().DefaultTimezone)

	fullPath := &utils.FullPath{
		PathSlice: []string{"Inexistant field"},
	}
	val := &utils.DataLeaf{
		Data: "cgrates.org",
	}
	expErr := "unsupported field prefix: <Inexistant field> when set field"
	if err := eeR.Compose(fullPath, val); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v> but received <%v>", expErr, err)
	}

}

func TestExportRequestSetFieldsPassErr(t *testing.T) {

	inData := map[string]utils.DataStorage{
		utils.MetaOpts: utils.MapStorage{},
	}
	onm := utils.NewOrderedNavigableMap()

	expData := map[string]*utils.OrderedNavigableMap{
		utils.MetaReq: onm,
	}

	cfg := config.NewDefaultCGRConfig()
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	dm.SetCache(cacheS)
	fltr := engine.NewFilterS(cfg, nil, dm)

	tplFlds := []*config.FCTemplate{
		{
			Tag:     "Tor",
			Type:    utils.MetaConstant,
			Value:   utils.NewRSRParsersMustCompile("*voice", utils.InfieldSep),
			Path:    "*cgreq.ToR",
			Filters: []string{"inexistant"},
		},
	}

	eeR := NewExportRequest(inData, "cgrates.org", cacheS, fltr, expData, config.CgrConfig().GeneralCfg().RoundingDecimals, config.CgrConfig().GeneralCfg().DefaultTimezone)

	expErr := "NOT_FOUND:inexistant"
	if err := eeR.SetFields(context.Background(), tplFlds); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v> but received <%v>", expErr, err)
	}

}

func TestExportRequestSetFieldsPassFalse(t *testing.T) {

	inData := map[string]utils.DataStorage{

		utils.MetaOpts: utils.MapStorage{},
	}
	onm := utils.NewOrderedNavigableMap()

	expData := map[string]*utils.OrderedNavigableMap{
		utils.MetaReq: onm,
	}

	cfg := config.NewDefaultCGRConfig()
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	dm.SetCache(cacheS)
	fltr := engine.NewFilterS(cfg, nil, dm)

	tplFlds := []*config.FCTemplate{
		{
			Tag:     "Tor",
			Type:    utils.MetaConstant,
			Value:   utils.NewRSRParsersMustCompile("*voice", utils.InfieldSep),
			Path:    "*cgreq.ToR",
			Filters: []string{"*gt:~*opts.*rateSCost.Cost:0.5"},
		},
	}

	eeR := NewExportRequest(inData, "cgrates.org", cacheS, fltr, expData, config.CgrConfig().GeneralCfg().RoundingDecimals, config.CgrConfig().GeneralCfg().DefaultTimezone)
	pastEeR := eeR
	if err := eeR.SetFields(context.Background(), tplFlds); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(pastEeR, eeR) {
		t.Errorf("Expected \n<%v>,\n Received <%v>", pastEeR, eeR)
	}

}
