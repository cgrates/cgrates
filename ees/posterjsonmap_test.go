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

package ees

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestPosterJsonMapID(t *testing.T) {
	pstrEE := &PosterJSONMapEE{
		id: "3",
	}
	if rcv := pstrEE.ID(); !reflect.DeepEqual(rcv, "3") {
		t.Errorf("Expected %+v but got %+v", "3", rcv)
	}
}

func TestPosterJsonMapGetMetrics(t *testing.T) {
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	if err != nil {
		t.Error(err)
	}
	pstrEE := &PosterJSONMapEE{
		dc: dc,
	}

	if rcv := pstrEE.GetMetrics(); !reflect.DeepEqual(rcv, pstrEE.dc) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(rcv), utils.ToJSON(pstrEE.dc))
	}
}

func TestPosterJsonMapNewPosterJSONMapEECase2(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaAMQPV1jsonMap
	cgrCfg.EEsCfg().Exporters[0].ExportPath = utils.EmptyString
	filterS := engine.NewFilterS(cgrCfg, nil, nil)
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	if err != nil {
		t.Error(err)
	}
	pstrJSON, err := NewPosterJSONMapEE(cgrCfg, 0, filterS, dc)
	if err != nil {
		t.Error(err)
	}
	pstrJSONExpect := engine.NewAMQPv1Poster(cgrCfg.EEsCfg().Exporters[0].ExportPath,
		cgrCfg.EEsCfg().Exporters[0].Attempts, cgrCfg.EEsCfg().Exporters[0].Opts)
	if !reflect.DeepEqual(pstrJSON.poster, pstrJSONExpect) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(pstrJSONExpect), utils.ToJSON(pstrJSON.poster))
	}
}

func TestPosterJsonMapNewPosterJSONMapEECase3(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaSQSjsonMap
	cgrCfg.EEsCfg().Exporters[0].ExportPath = utils.EmptyString
	filterS := engine.NewFilterS(cgrCfg, nil, nil)
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	if err != nil {
		t.Error(err)
	}
	pstrJSON, err := NewPosterJSONMapEE(cgrCfg, 0, filterS, dc)
	if err != nil {
		t.Error(err)
	}

	if _, canCast := pstrJSON.poster.(*engine.SQSPoster); !canCast {
		t.Error("Can't cast")
	}
}

func TestPosterJsonMapNewPosterJSONMapEECase4(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaKafkajsonMap
	filterS := engine.NewFilterS(cgrCfg, nil, nil)
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	if err != nil {
		t.Error(err)
	}
	pstrJSON, err := NewPosterJSONMapEE(cgrCfg, 0, filterS, dc)
	if err != nil {
		t.Error(err)
	}
	pstrJSONExpect := engine.NewKafkaPoster(cgrCfg.EEsCfg().Exporters[0].ExportPath,
		cgrCfg.EEsCfg().Exporters[0].Attempts, cgrCfg.EEsCfg().Exporters[0].Opts)
	if !reflect.DeepEqual(pstrJSON.poster, pstrJSONExpect) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(pstrJSONExpect), utils.ToJSON(pstrJSON.poster))
	}
}

func TestPosterJsonMapNewPosterJSONMapEECase5(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaS3jsonMap
	filterS := engine.NewFilterS(cgrCfg, nil, nil)
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	if err != nil {
		t.Error(err)
	}
	pstrJSON, err := NewPosterJSONMapEE(cgrCfg, 0, filterS, dc)
	if err != nil {
		t.Error(err)
	}
	pstrJSONExpect := engine.NewS3Poster(cgrCfg.EEsCfg().Exporters[0].ExportPath,
		cgrCfg.EEsCfg().Exporters[0].Attempts, cgrCfg.EEsCfg().Exporters[0].Opts)
	if !reflect.DeepEqual(pstrJSON.poster, pstrJSONExpect) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(pstrJSONExpect), utils.ToJSON(pstrJSON.poster))
	}
}

func TestPosterJsonMapExportEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaSQSjsonMap
	cgrEv := new(utils.CGREvent)
	newIDb := engine.NewInternalDB(nil, nil, true)
	newDM := engine.NewDataManager(newIDb, cgrCfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cgrCfg, nil, newDM)
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	if err != nil {
		t.Error(err)
	}

	pstrEE, err := NewPosterJSONMapEE(cgrCfg, 0, filterS, dc)
	if err != nil {
		t.Error(err)
	}
	cgrEv.Event = map[string]interface{}{
		"test": "string",
	}
	cgrCfg.EEsCfg().Exporters[pstrEE.cfgIdx].Fields = []*config.FCTemplate{
		{
			Path: "*exp.1", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*req.field1", utils.InfieldSep),
		},
		{
			Path: "*exp.2", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("*req.field2", utils.InfieldSep),
		},
	}
	for _, field := range cgrCfg.EEsCfg().Exporters[pstrEE.cfgIdx].Fields {
		field.ComputePath()
	}
	errExpect := "MissingRegion: could not find region configuration"
	if err := pstrEE.ExportEvent(cgrEv); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %q but received %q", errExpect, err)
	}
	cgrCfg.EEsCfg().Exporters[pstrEE.cfgIdx].ComputeFields()
	if err := pstrEE.ExportEvent(cgrEv); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %q but received %q", errExpect, err)
	}
	cgrCfg.EEsCfg().Exporters[0].Fields = []*config.FCTemplate{
		{
			Path: "*exp.1", Type: utils.MetaVariable,
			Value:   config.NewRSRParsersMustCompile("~*req.field1", utils.InfieldSep),
			Filters: []string{"*wrong-type"},
		},
		{
			Path: "*exp.1", Type: utils.MetaVariable,
			Value:   config.NewRSRParsersMustCompile("~*req.field1", utils.InfieldSep),
			Filters: []string{"*wrong-type"},
		},
	}
	for _, field := range cgrCfg.EEsCfg().Exporters[0].Fields {
		field.ComputePath()
	}
	cgrCfg.EEsCfg().Exporters[0].ComputeFields()
	errExpect = "inline parse error for string: <*wrong-type>"
	if err := pstrEE.ExportEvent(cgrEv); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %q but received %q", errExpect, err)
	}
	//test marshal invalid input
	cgrEv.Event = map[string]interface{}{
		"test": make(chan int),
	}
	cgrCfg.EEsCfg().Exporters[0].Fields = []*config.FCTemplate{{}}
	for _, field := range cgrCfg.EEsCfg().Exporters[0].Fields {
		field.ComputePath()
	}
	cgrCfg.EEsCfg().Exporters[0].ComputeFields()
	errExpect = "json: unsupported type: chan int"
	if err := pstrEE.ExportEvent(cgrEv); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %q but received %q", errExpect, err)
	}
	pstrEE.OnEvicted("test", "test")
}

// func TestPosterJsonMapExportEvent2(t *testing.T) {
// 	cgrCfg := config.NewDefaultCGRConfig()
// 	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaSQSjsonMap
// 	cgrEv := new(utils.CGREvent)
// 	newIDb := engine.NewInternalDB(nil, nil, true)
// 	newDM := engine.NewDataManager(newIDb, cgrCfg.CacheCfg(), nil)
// 	filterS := engine.NewFilterS(cgrCfg, nil, newDM)
// 	dc, err := newEEMetrics(utils.FirstNonEmpty(
// 		"Local",
// 		utils.EmptyString,
// 	))
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
// 		defer r.Body.Close()
// 		// body, err := io.ReadAll(r.Body)
// 		// if err != nil {
// 		// 	t.Error(err)
// 		// }
// 		// if strBody := string(body); strBody != bodyExpect {
// 		// 	t.Errorf("Expected %q but received %q", bodyExpect, strBody)
// 		// }
// 		rw.WriteHeader(http.StatusOK)
// 	}))
// 	defer srv.Close()
// 	cgrCfg.EEsCfg().Exporters[0].ExportPath = srv.URL + "/"
// 	pstrEE, err := NewPosterJSONMapEE(cgrCfg, 0, filterS, dc)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	cgrEv.Event = map[string]interface{}{
// 		"test": "string",
// 	}
// 	cgrCfg.EEsCfg().Exporters[pstrEE.cfgIdx].Fields = []*config.FCTemplate{
// 		{
// 			Path: "*exp.1", Type: utils.MetaVariable,
// 			Value: config.NewRSRParsersMustCompile("~*req.field1", utils.InfieldSep),
// 		},
// 		{
// 			Path: "*exp.2", Type: utils.MetaVariable,
// 			Value: config.NewRSRParsersMustCompile("*req.field2", utils.InfieldSep),
// 		},
// 	}
// 	for _, field := range cgrCfg.EEsCfg().Exporters[pstrEE.cfgIdx].Fields {
// 		field.ComputePath()
// 	}
// 	cgrCfg.EEsCfg().Exporters[pstrEE.cfgIdx].ComputeFields()
// 	if err := pstrEE.ExportEvent(cgrEv); err != nil {
// 		t.Error(err)
// 	}
// }
