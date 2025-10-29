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
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestExportReqParseFieldDateTimeDaily(t *testing.T) {
	EventReq := NewExportRequest(map[string]utils.DataStorage{}, "", nil, nil)
	fctTemp := &config.FCTemplate{
		Type:     utils.MetaDateTime,
		Value:    config.NewRSRParsersMustCompile("*daily", utils.InfieldSep),
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
	EventReq := NewExportRequest(map[string]utils.DataStorage{}, "", nil, nil)
	fctTemp := &config.FCTemplate{
		Type:     utils.MetaDateTime,
		Value:    config.NewRSRParsersMustCompile("*daily", utils.InfieldSep),
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
	EventReq := NewExportRequest(map[string]utils.DataStorage{}, "", nil, nil)
	fctTemp := &config.FCTemplate{
		Type:     utils.MetaDateTime,
		Value:    config.NewRSRParsersMustCompile("*monthly", utils.InfieldSep),
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
	EventReq := NewExportRequest(map[string]utils.DataStorage{}, "", nil, nil)
	fctTemp := &config.FCTemplate{
		Type:     utils.MetaDateTime,
		Value:    config.NewRSRParsersMustCompile("*monthly_estimated", utils.InfieldSep),
		Layout:   "“Mon Jan _2 15:04:05 2006”",
		Timezone: "Local",
	}
	result, err := EventReq.ParseField(fctTemp)
	if err != nil {
		t.Errorf("Expected %v but received %v", nil, err)
	}

	expected, err := utils.ParseTimeDetectLayout("*monthly_estimated", utils.FirstNonEmpty(fctTemp.Timezone, config.CgrConfig().GeneralCfg().DefaultTimezone))
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
	EventReq := NewExportRequest(map[string]utils.DataStorage{}, "", nil, nil)
	fctTemp := &config.FCTemplate{
		Type:     utils.MetaDateTime,
		Value:    config.NewRSRParsersMustCompile("*yearly", utils.InfieldSep),
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
	EventReq := NewExportRequest(map[string]utils.DataStorage{}, "", nil, nil)
	fctTemp := &config.FCTemplate{
		Type:     utils.MetaDateTime,
		Value:    config.NewRSRParsersMustCompile(utils.MetaUnlimited, utils.InfieldSep),
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
	EventReq := NewExportRequest(map[string]utils.DataStorage{}, "", nil, nil)
	fctTemp := &config.FCTemplate{
		Type:     utils.MetaDateTime,
		Value:    config.NewRSRParsersMustCompile("", utils.InfieldSep),
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
	EventReq := NewExportRequest(map[string]utils.DataStorage{}, "", nil, nil)
	fctTemp := &config.FCTemplate{
		Type:     utils.MetaDateTime,
		Value:    config.NewRSRParsersMustCompile("*month_endTest", utils.InfieldSep),
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
	EventReq := NewExportRequest(map[string]utils.DataStorage{}, "", nil, nil)
	fctTemp := &config.FCTemplate{
		Type:     utils.MetaDateTime,
		Value:    config.NewRSRParsersMustCompile("*month_endTest", utils.InfieldSep),
		Layout:   "“Mon Jan _2 15:04:05 2006”",
		Timezone: "/",
	}
	_, err := EventReq.ParseField(fctTemp)
	expected := "time: invalid location name"
	if err == nil || err.Error() != expected {
		t.Errorf("Expected <%+v> but received <%+v>", expected, err)
	}
}

func TestExportReqFieldAsINterfaceOnePath(t *testing.T) {
	mS := map[string]utils.DataStorage{
		utils.MetaReq: utils.MapStorage{
			utils.AccountField: "1004",
			utils.Usage:        "20m",
			utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
		},
		utils.MetaOpts: utils.MapStorage{
			utils.APIKey: "attr12345",
		},
		utils.MetaVars: utils.MapStorage{
			utils.RequestType: utils.MetaRated,
			utils.Subsystems:  utils.MetaChargers,
		},
	}
	eventReq := NewExportRequest(mS, "", nil, nil)
	fldPath := []string{utils.MetaReq}
	if val, err := eventReq.FieldAsInterface(fldPath); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(val, mS[utils.MetaReq]) {
		t.Errorf("Expected %+v \n, received %+v", val, mS[utils.MetaReq])
	}
	fldPath = []string{"default"}
	if _, err = eventReq.FieldAsInterface(fldPath); err == nil {
		t.Error("expected error")
	}

	fldPath = []string{utils.MetaOpts}
	if val, err := eventReq.FieldAsInterface(fldPath); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(val, mS[utils.MetaOpts]) {
		t.Errorf("Expected %+v \n, received %+v", val, mS[utils.MetaOpts])
	}

	fldPath = []string{utils.MetaVars}
	if val, err := eventReq.FieldAsInterface(fldPath); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(val, mS[utils.MetaVars]) {
		t.Errorf("Expected %+v \n, received %+v", val, mS[utils.MetaVars])
	}
	fldPath = []string{utils.MetaUCH}
	if _, err = eventReq.FieldAsInterface(fldPath); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
}
func TestEventReqFieldAsInterface(t *testing.T) {
	inData := map[string]utils.DataStorage{
		utils.MetaReq: utils.MapStorage{
			"Account": "1001",
			"Usage":   "10m",
		},
	}
	eventReq := NewExportRequest(inData, "cgrates.org", nil, nil)
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

func TestEventReqNewEventExporter(t *testing.T) {
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
	expected := &ExportRequest{
		inData:  inData,
		filterS: nil,
		tnt:     "cgrates.org",
		ExpData: expData,
	}
	eventReq := NewExportRequest(inData, "cgrates.org", nil, expData)
	if !reflect.DeepEqual(expected, eventReq) {
		t.Errorf("Expected %v \n but received \n %v", expected, eventReq)
	}
}

func TestExportRequestSetAsSlice(t *testing.T) {
	onm := utils.NewOrderedNavigableMap()
	fullpath := &utils.FullPath{
		PathSlice: []string{utils.MetaReq, utils.MetaTenant},
		Path:      utils.MetaTenant,
	}
	value := &utils.DataLeaf{
		Data: "value1",
	}
	onm.Append(fullpath, value)
	expData := map[string]*utils.OrderedNavigableMap{
		"default": onm,
	}

	eeR := &ExportRequest{
		inData: map[string]utils.DataStorage{
			utils.MetaReq: utils.MapStorage{
				"Account": "1001",
				"Usage":   "10m",
			},
			utils.MetaOpts: utils.MapStorage{},
		},
		tnt:     "cgrates.org",
		ExpData: expData,
	}

	fullPath := &utils.FullPath{
		PathSlice: []string{utils.MetaUCH, utils.MetaReq, utils.MetaTenant},
		Path:      utils.MetaTenant,
	}
	val := &utils.DataLeaf{
		Data: "value1",
	}

	if err := eeR.SetAsSlice(fullPath, val); err != nil {
		t.Error(err)
	}
	fullPath.PathSlice[0] = utils.MetaOpts
	if err = eeR.SetAsSlice(fullPath, val); err != nil {
		t.Error(err)
	}
	fullPath.PathSlice[0] = "default"
	if err = eeR.SetAsSlice(fullPath, val); err != nil {
		t.Error(err)
	} else if err = eeR.SetAsSlice(&utils.FullPath{PathSlice: []string{"Val"}}, val); err == nil {
		t.Error(err)
	}
}

func TestExportRequestParseField(t *testing.T) {
	Cache.Clear(nil)
	fctTemp := &config.FCTemplate{
		Type:       utils.MetaMaskedDestination,
		Value:      config.NewRSRParsersMustCompile("*month_endTest", utils.InfieldSep),
		Layout:     "“Mon Jan _2 15:04:05 2006”",
		Timezone:   "Local",
		MaskLen:    3,
		MaskDestID: "dest",
	}
	mS := map[string]utils.DataStorage{
		utils.MetaReq: utils.MapStorage{
			utils.AccountField: "1004",
			utils.Usage:        "20m",
			utils.Destination:  "dest",
		},
		utils.MetaOpts: utils.MapStorage{
			utils.APIKey: "attr12345",
		},
		utils.MetaVars: utils.MapStorage{
			utils.RequestType: utils.MetaRated,
			utils.Subsystems:  utils.MetaChargers,
		},
	}
	eventReq := NewExportRequest(mS, "", nil, nil)

	if _, err := eventReq.ParseField(fctTemp); err != nil {
		t.Error(err)
	}
	fctTemp.Type = utils.MetaFiller
	if _, err = eventReq.ParseField(fctTemp); err != nil {
		t.Error(err)
	}
	fctTemp.Type = utils.MetaGroup
	if _, err = eventReq.ParseField(fctTemp); err != nil {
		t.Error(err)
	}
}

func TestExportRequestAppend(t *testing.T) {
	onm := utils.NewOrderedNavigableMap()
	fullpath := &utils.FullPath{
		PathSlice: []string{utils.MetaReq, utils.MetaTenant},
		Path:      utils.MetaTenant,
	}
	value := &utils.DataLeaf{
		Data: "value1",
	}
	onm.Append(fullpath, value)
	expData := map[string]*utils.OrderedNavigableMap{
		"default": onm,
	}

	eeR := &ExportRequest{
		inData: map[string]utils.DataStorage{
			utils.MetaReq: utils.MapStorage{
				"Account": "1001",
				"Usage":   "10m",
			},
			utils.MetaOpts: utils.MapStorage{},
		},
		tnt:     "cgrates.org",
		ExpData: expData,
	}

	fullPath := &utils.FullPath{
		PathSlice: []string{utils.MetaUCH, utils.MetaReq, utils.MetaTenant},
		Path:      utils.MetaTenant,
	}
	val := &utils.DataLeaf{
		Data: "value1",
	}

	if err := eeR.Append(fullPath, val); err != nil {
		t.Error(err)
	}
	fullPath.PathSlice[0] = utils.MetaOpts
	if err = eeR.Append(fullPath, val); err != nil {
		t.Error(err)
	}
	fullPath.PathSlice[0] = "default"
	if err = eeR.Append(fullPath, val); err != nil {
		t.Error(err)
	} else if err = eeR.Append(&utils.FullPath{PathSlice: []string{"Val"}}, val); err == nil {
		t.Error(err)
	}

}

func TestExportRequestCompose(t *testing.T) {
	onm := utils.NewOrderedNavigableMap()
	fullPath := &utils.FullPath{
		PathSlice: []string{utils.MetaReq, utils.MetaTenant},
		Path:      utils.MetaTenant,
	}
	val := &utils.DataLeaf{
		Data: "value1",
	}
	onm.Append(fullPath, val)

	eeR := &ExportRequest{
		inData: map[string]utils.DataStorage{
			utils.MetaReq: utils.MapStorage{
				"Account": "1001",
				"Usage":   "10m",
			},
			utils.MetaOpts: utils.MapStorage{
				"*opts": "val",
			},
		},
		filterS: nil,
		tnt:     "cgrates.org",
		ExpData: map[string]*utils.OrderedNavigableMap{
			utils.MetaReq: onm,
		},
	}
	if err := eeR.Compose(&utils.FullPath{
		PathSlice: []string{utils.MetaReq},
		Path:      "path"}, &utils.DataLeaf{
		Data: "Value"}); err == nil {
		t.Error(err)
	} else if err = eeR.Compose(&utils.FullPath{
		PathSlice: []string{"default"},
		Path:      "path"}, &utils.DataLeaf{
		Data: "Value"}); err == nil {
		t.Error(err)
	} else if err = eeR.Compose(&utils.FullPath{
		PathSlice: []string{utils.MetaUCH},
		Path:      "pathvalue"}, &utils.DataLeaf{
		Data: "Value"}); err != nil {
		t.Error(err)
	} else if err = eeR.Compose(&utils.FullPath{
		PathSlice: []string{utils.MetaOpts, "*opts"},
		Path:      "pathvalue"}, &utils.DataLeaf{
		Data: "Value"}); err != nil {
		t.Error(err)
	}

}

func TestExportRequestSetFields(t *testing.T) {
	Cache.Clear(nil)
	onm := utils.NewOrderedNavigableMap()
	fullPath := &utils.FullPath{
		PathSlice: []string{utils.MetaReq, utils.MetaAccountID},
		Path:      utils.MetaReq + utils.MetaAccountID,
	}
	val := &utils.DataLeaf{
		Data: "value1",
	}
	onm.Append(fullPath, val)
	cfg := config.NewDefaultCGRConfig()
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(db, config.CgrConfig().CacheCfg(), nil)
	eeR := &ExportRequest{
		inData: map[string]utils.DataStorage{
			utils.MetaReq: utils.MapStorage{
				"Account": "1001",
				"Usage":   "10m",
			},
			utils.MetaOpts: utils.MapStorage{
				"*opts": "val",
			},
		},
		filterS: NewFilterS(cfg, nil, dm),
		tnt:     "cgrates.org",
		ExpData: map[string]*utils.OrderedNavigableMap{
			utils.MetaReq: onm,
		},
	}
	fctTemp := []*config.FCTemplate{
		{
			Type:     utils.MetaComposed,
			Value:    config.NewRSRParsersMustCompile("1003", utils.InfieldSep),
			Timezone: "Local",
			Path:     "<*uch;*opts>",
		},
	}
	if err = eeR.SetFields(fctTemp); err == nil || err.Error() != "unsupported field prefix: <*uch*opts> when set field" {
		t.Error(err)
	}
}

func TestExportRequestFieldAsString(t *testing.T) {
	inData := map[string]utils.DataStorage{
		utils.MetaReq: utils.MapStorage{
			"Account": "1001",
			"Usage":   "10m",
		},
	}
	eventReq := NewExportRequest(inData, "cgrates.org", nil, nil)
	fldPath := []string{utils.MetaReq, "Usage"}
	expVal := "10m"
	if rcv, err := eventReq.FieldAsString(fldPath); err != nil {
		t.Error(err)
	} else if rcv != expVal {
		t.Errorf("expected %v,received %v", expVal, rcv)
	}
	fldPath[0] = utils.MetaUCH
	if _, err = eventReq.FieldAsString(fldPath); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
}
func TestExportRequestParseFieldErr(t *testing.T) {
	mp := utils.NewSecureMapStorage()
	inData := map[string]utils.DataStorage{
		utils.MetaReq: mp,
	}
	EventReq := NewExportRequest(inData, "", nil, nil)
	fctTemp := &config.FCTemplate{
		Type:     utils.MetaMaskedDestination,
		Value:    config.NewRSRParsersMustCompile("*daily", utils.InfieldSep),
		Layout:   "“Mon Jan _2 15:04:05 2006”",
		Timezone: "Local",
	}
	if _, err := EventReq.ParseField(fctTemp); err == nil {
		t.Error(err)
	}

}

func TestExportRequestSetFields2(t *testing.T) {

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
	eventReq := NewExportRequest(inData, "cgrates.org", nil, expData)
	tpFields := []*config.FCTemplate{
		{
			Tag:   "Tenant",
			Path:  utils.MetaReq + utils.NestingSep + utils.AccountField,
			Type:  utils.MetaGroup,
			Value: config.NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep),
		},
	}
	tpFields[0].ComputePath()
	if err := eventReq.SetFields(tpFields); err != nil {
		t.Error(err)
	}
}

func TestExportRequestString(t *testing.T) {
	eeR := &ExportRequest{}
	jsonStr := eeR.String()
	if jsonStr == "" {
		t.Errorf("Expected non-empty JSON string, but got empty")
	}
}
