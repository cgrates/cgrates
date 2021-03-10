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
