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

package agents

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/radigo"
	"github.com/fiorix/go-diameter/v4/diam"
	"github.com/fiorix/go-diameter/v4/diam/avp"
	"github.com/fiorix/go-diameter/v4/diam/datatype"
)

func TestAgReqSetFields(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	agReq := NewAgentRequest(nil, nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)
	// populate request, emulating the way will be done in HTTPAgent
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.CGRID, PathSlice: []string{utils.CGRID}},
		utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()))
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.ToR, PathSlice: []string{utils.ToR}}, utils.MetaVoice)
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.AccountField, PathSlice: []string{utils.AccountField}}, "1001")
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.Destination, PathSlice: []string{utils.Destination}}, "1002")
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.AnswerTime, PathSlice: []string{utils.AnswerTime}},
		time.Date(2013, 12, 30, 15, 0, 1, 0, time.UTC))
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.RequestType, PathSlice: []string{utils.RequestType}}, utils.MetaPrepaid)
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.Usage, PathSlice: []string{utils.Usage}}, 3*time.Minute)

	cgrRply := &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{
		utils.CapAttributes: {Type: utils.NMMapType, Map: map[string]*utils.DataNode{
			"PaypalAccount": utils.NewLeafNode("cgrates@paypal.com"),
		}},
		utils.CapMaxUsage: utils.NewLeafNode(120 * time.Second),
		utils.Error:       utils.NewLeafNode(""),
	},
	}
	agReq.CGRReply = cgrRply

	tplFlds := []*config.FCTemplate{
		{Tag: "Tenant",
			Path: utils.MetaRep + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep)},
		{Tag: "Account",
			Path: utils.MetaRep + utils.NestingSep + utils.AccountField, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*cgreq.Account", utils.InfieldSep)},
		{Tag: "Destination",
			Path: utils.MetaRep + utils.NestingSep + utils.Destination, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*cgreq.Destination", utils.InfieldSep)},

		{Tag: "RequestedUsageVoice",
			Path: utils.MetaRep + utils.NestingSep + "RequestedUsage", Type: utils.MetaVariable,
			Filters: []string{"*string:~*cgreq.ToR:*voice"},
			Value: config.NewRSRParsersMustCompile(
				"~*cgreq.Usage{*duration_seconds}", utils.InfieldSep)},
		{Tag: "RequestedUsageData",
			Path: utils.MetaRep + utils.NestingSep + "RequestedUsage", Type: utils.MetaVariable,
			Filters: []string{"*string:~*cgreq.ToR:*data"},
			Value: config.NewRSRParsersMustCompile(
				"~*cgreq.Usage{*duration_nanoseconds}", utils.InfieldSep)},
		{Tag: "RequestedUsageSMS",
			Path: utils.MetaRep + utils.NestingSep + "RequestedUsage", Type: utils.MetaVariable,
			Filters: []string{"*string:~*cgreq.ToR:*sms"},
			Value: config.NewRSRParsersMustCompile(
				"~*cgreq.Usage{*duration_nanoseconds}", utils.InfieldSep)},

		{Tag: "AttrPaypalAccount",
			Path: utils.MetaRep + utils.NestingSep + "PaypalAccount", Type: utils.MetaVariable,
			Filters: []string{"*empty:~*cgrep.Error:"},
			Value: config.NewRSRParsersMustCompile(
				"~*cgrep.Attributes.PaypalAccount", utils.InfieldSep)},
		{Tag: "MaxUsage",
			Path: utils.MetaRep + utils.NestingSep + "MaxUsage", Type: utils.MetaVariable,
			Filters: []string{"*empty:~*cgrep.Error:"},
			Value: config.NewRSRParsersMustCompile(
				"~*cgrep.MaxUsage{*duration_seconds}", utils.InfieldSep)},
		{Tag: "Error",
			Path: utils.MetaRep + utils.NestingSep + "Error", Type: utils.MetaVariable,
			Filters: []string{"*rsr:~*cgrep.Error:!^$"},
			Value: config.NewRSRParsersMustCompile(
				"~*cgrep.Error", utils.InfieldSep)},
	}
	for _, v := range tplFlds {
		v.ComputePath()
	}

	eMp := utils.NewOrderedNavigableMap()
	eMp.SetAsSlice(&utils.FullPath{Path: utils.Tenant, PathSlice: []string{utils.Tenant}}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "cgrates.org"}}})
	eMp.SetAsSlice(&utils.FullPath{Path: utils.AccountField, PathSlice: []string{utils.AccountField}}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "1001"}}})
	eMp.SetAsSlice(&utils.FullPath{Path: utils.Destination, PathSlice: []string{utils.Destination}}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "1002"}}})
	eMp.SetAsSlice(&utils.FullPath{Path: "RequestedUsage", PathSlice: []string{"RequestedUsage"}}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "180"}}})
	eMp.SetAsSlice(&utils.FullPath{Path: "PaypalAccount", PathSlice: []string{"PaypalAccount"}}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "cgrates@paypal.com"}}})
	eMp.SetAsSlice(&utils.FullPath{Path: "MaxUsage", PathSlice: []string{"MaxUsage"}}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "120"}}})

	if err := agReq.SetFields(tplFlds); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(agReq.Reply, eMp) {
		t.Log(utils.ToJSON(eMp.GetOrder()))
		t.Log(utils.ToJSON(agReq.Reply.GetOrder()))
		t.Errorf("expecting: %+v,\n received: %+v", eMp, agReq.Reply)
	}
}

func TestAgentRequestSetFields(t *testing.T) {
	req := map[string]interface{}{
		utils.AccountField: 1009,
		utils.Tenant:       "cgrates.org",
	}
	cfg := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, true),
		config.CgrConfig().CacheCfg(), nil)
	ar := NewAgentRequest(utils.MapStorage(req), nil,
		nil, nil, nil, config.NewRSRParsersMustCompile("", utils.NestingSep),
		"cgrates.org", "", engine.NewFilterS(cfg, nil, dm),
		utils.MapStorage(req), utils.MapStorage(req))
	input := []*config.FCTemplate{}
	if err := ar.SetFields(input); err != nil {
		t.Error(err)
	}
	// tplFld.Type == utils.MetaNone
	input = []*config.FCTemplate{{Type: utils.MetaNone}}
	if err := ar.SetFields(input); err != nil {
		t.Error(err)
	}
	// unsupported type: <>
	input = []*config.FCTemplate{{Blocker: true}}
	if err := ar.SetFields(input); err == nil || err.Error() != "unsupported type: <>" {
		t.Error(err)
	}
	// case utils.MetaVars
	input = []*config.FCTemplate{
		{
			Path:  fmt.Sprintf("%s.Account", utils.MetaVars),
			Tag:   fmt.Sprintf("%s.Account", utils.MetaVars),
			Type:  utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~"+utils.MetaReq+".Account", utils.InfieldSep),
		},
	}
	input[0].ComputePath()

	if err := ar.SetFields(input); err != nil {
		t.Error(err)
	} else if val, err := ar.Vars.FieldAsInterface([]string{"Account"}); err != nil {
		t.Error(err)
	} else if nm, ok := val.([]*utils.DataNode); !ok {
		t.Errorf("Expecting NM items<%T>", val)
	} else if len(nm) != 1 {
		t.Error("Expecting one item")
	} else if nm[0].Value.Data != "1009" {
		t.Error("Expecting 1009, received: ", nm[0].Value.Data)
	}

	// case utils.MetaCgreq
	input = []*config.FCTemplate{
		{
			Path:  fmt.Sprintf("%s.Account", utils.MetaCgreq),
			Tag:   fmt.Sprintf("%s.Account", utils.MetaCgreq),
			Type:  utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~"+utils.MetaReq+".Account", utils.InfieldSep),
		},
	}
	input[0].ComputePath()
	if err := ar.SetFields(input); err != nil {
		t.Error(err)
	} else if val, err := ar.CGRRequest.FieldAsInterface([]string{"Account"}); err != nil {
		t.Error(err)
	} else if nm, ok := val.([]*utils.DataNode); !ok {
		t.Error("Expecting NM items")
	} else if len(nm) != 1 {
		t.Error("Expecting one item")
	} else if nm[0].Value.Data != "1009" {
		t.Error("Expecting 1009, received: ", nm[0].Value.Data)
	}

	// case utils.MetaCgrep
	input = []*config.FCTemplate{
		{
			Path:  fmt.Sprintf("%s.Account", utils.MetaCgrep),
			Tag:   fmt.Sprintf("%s.Account", utils.MetaCgrep),
			Type:  utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~"+utils.MetaReq+".Account", utils.InfieldSep),
		},
	}
	input[0].ComputePath()
	if err := ar.SetFields(input); err != nil {
		t.Error(err)
	} else if val, err := ar.CGRReply.FieldAsInterface([]string{"Account"}); err != nil {
		t.Error(err)
	} else if nm, ok := val.([]*utils.DataNode); !ok {
		t.Error("Expecting NM items")
	} else if len(nm) != 1 {
		t.Error("Expecting one item")
	} else if nm[0].Value.Data != "1009" {
		t.Error("Expecting 1009, received: ", nm[0].Value.Data)
	}

	// case utils.MetaRep
	input = []*config.FCTemplate{
		{
			Path:  fmt.Sprintf("%s.Account", utils.MetaRep),
			Tag:   fmt.Sprintf("%s.Account", utils.MetaRep),
			Type:  utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~"+utils.MetaReq+".Account", utils.InfieldSep),
		},
	}
	input[0].ComputePath()
	if err := ar.SetFields(input); err != nil {
		t.Error(err)
	} else if val, err := ar.Reply.FieldAsInterface([]string{"Account"}); err != nil {
		t.Error(err)
	} else if nm, ok := val.([]*utils.DataNode); !ok {
		t.Error("Expecting NM items")
	} else if len(nm) != 1 {
		t.Error("Expecting one item")
	} else if nm[0].Value.Data != "1009" {
		t.Error("Expecting 1009, received: ", nm[0].Value.Data)
	}

	// case utils.MetaDiamreq
	input = []*config.FCTemplate{
		{
			Path:  fmt.Sprintf("%s.Account", utils.MetaDiamreq),
			Tag:   fmt.Sprintf("%s.Account", utils.MetaDiamreq),
			Type:  utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~"+utils.MetaReq+".Account", utils.InfieldSep),
		},
	}
	input[0].ComputePath()
	if err := ar.SetFields(input); err != nil {
		t.Error(err)
	} else if val, err := ar.diamreq.FieldAsInterface([]string{"Account"}); err != nil {
		t.Error(err)
	} else if nm, ok := val.([]*utils.DataNode); !ok {
		t.Error("Expecting NM items")
	} else if len(nm) != 1 {
		t.Error("Expecting one item")
	} else if nm[0].Value.Data != "1009" {
		t.Error("Expecting 1009, received: ", nm[0].Value.Data)
	}

	//MetaComposed
	input = []*config.FCTemplate{
		{
			Path:  fmt.Sprintf("%s.AccountID", utils.MetaVars),
			Tag:   fmt.Sprintf("%s.AccountID", utils.MetaVars),
			Type:  utils.MetaComposed,
			Value: config.NewRSRParsersMustCompile("~"+utils.MetaReq+".Tenant", utils.InfieldSep),
		},
		{
			Path:  fmt.Sprintf("%s.AccountID", utils.MetaVars),
			Tag:   fmt.Sprintf("%s.AccountID", utils.MetaVars),
			Type:  utils.MetaComposed,
			Value: config.NewRSRParsersMustCompile(":", utils.InfieldSep),
		},
		{
			Path:  fmt.Sprintf("%s.AccountID", utils.MetaVars),
			Tag:   fmt.Sprintf("%s.AccountID", utils.MetaVars),
			Type:  utils.MetaComposed,
			Value: config.NewRSRParsersMustCompile("~"+utils.MetaReq+".Account", utils.InfieldSep),
		},
	}
	for _, v := range input {
		v.ComputePath()
	}

	if err := ar.SetFields(input); err != nil {
		t.Error(err)
	} else if val, err := ar.Vars.FieldAsInterface([]string{"AccountID"}); err != nil {
		t.Error(err)
	} else if nm, ok := val.([]*utils.DataNode); !ok {
		t.Error("Expecting NM items")
	} else if len(nm) != 1 {
		t.Error("Expecting one item")
	} else if nm[0].Value.Data != "cgrates.org:1009" {
		t.Error("Expecting 'cgrates.org:1009', received: ", nm[0].Value.Data)
	}

	// MetaConstant
	input = []*config.FCTemplate{
		{
			Path:  fmt.Sprintf("%s.Account", utils.MetaVars),
			Tag:   fmt.Sprintf("%s.Account", utils.MetaVars),
			Type:  utils.MetaConstant,
			Value: config.NewRSRParsersMustCompile("2020", utils.InfieldSep),
		},
	}
	input[0].ComputePath()
	if err := ar.SetFields(input); err != nil {
		t.Error(err)
	} else if val, err := ar.Vars.FieldAsInterface([]string{"Account"}); err != nil {
		t.Error(err)
	} else if nm, ok := val.([]*utils.DataNode); !ok {
		t.Error("Expecting NM items")
	} else if len(nm) != 1 {
		t.Error("Expecting one item")
	} else if nm[0].Value.Data != "2020" {
		t.Error("Expecting 1009, received: ", nm[0].Value.Data)
	}

	// Filters
	input = []*config.FCTemplate{
		{
			Path:    fmt.Sprintf("%s.AccountID", utils.MetaVars),
			Tag:     fmt.Sprintf("%s.AccountID", utils.MetaVars),
			Filters: []string{utils.MetaString + ":~" + utils.MetaVars + ".Account:1003"},
			Type:    utils.MetaConstant,
			Value:   config.NewRSRParsersMustCompile("2021", utils.InfieldSep),
		},
	}
	input[0].ComputePath()
	if err := ar.SetFields(input); err != nil {
		t.Error(err)
	} else if val, err := ar.Vars.FieldAsInterface([]string{"AccountID"}); err != nil {
		t.Error(err)
	} else if nm, ok := val.([]*utils.DataNode); !ok {
		t.Error("Expecting NM items")
	} else if len(nm) != 1 {
		t.Error("Expecting one item ", utils.ToJSON(nm))
	} else if nm[0].Value.Data != "cgrates.org:1009" {
		t.Error("Expecting 'cgrates.org:1009', received: ", nm[0].Value.Data)
	}

	input = []*config.FCTemplate{
		{
			Path:    fmt.Sprintf("%s.Account", utils.MetaVars),
			Tag:     fmt.Sprintf("%s.Account", utils.MetaVars),
			Filters: []string{"Not really a filter"},
			Type:    utils.MetaConstant,
			Value:   config.NewRSRParsersMustCompile("2021", utils.InfieldSep),
		},
	}
	input[0].ComputePath()
	if err := ar.SetFields(input); err == nil || err.Error() != "NOT_FOUND:Not really a filter" {
		t.Errorf("Expecting: 'NOT_FOUND:Not really a filter', received: %+v", err)
	}

	// Blocker: true
	input = []*config.FCTemplate{
		{
			Path:    fmt.Sprintf("%s.Name", utils.MetaVars),
			Tag:     fmt.Sprintf("%s.Name", utils.MetaVars),
			Type:    utils.MetaVariable,
			Value:   config.NewRSRParsersMustCompile("~"+utils.MetaReq+".Account", utils.InfieldSep),
			Blocker: true,
		},
		{
			Path:  fmt.Sprintf("%s.Name", utils.MetaVars),
			Tag:   fmt.Sprintf("%s.Name", utils.MetaVars),
			Type:  utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("1005", utils.InfieldSep),
		},
	}
	for _, v := range input {
		v.ComputePath()
	}
	if err := ar.SetFields(input); err != nil {
		t.Error(err)
	} else if val, err := ar.Vars.FieldAsInterface([]string{"Name"}); err != nil {
		t.Error(err)
	} else if nm, ok := val.([]*utils.DataNode); !ok {
		t.Error("Expecting NM items")
	} else if len(nm) != 1 {
		t.Error("Expecting one item")
	} else if nm[0].Value.Data != "1009" {
		t.Error("Expecting 1009, received: ", nm[0].Value.Data)
	}

	// ErrNotFound
	input = []*config.FCTemplate{
		{
			Path:  fmt.Sprintf("%s.Test", utils.MetaVars),
			Tag:   fmt.Sprintf("%s.Test", utils.MetaVars),
			Type:  utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~"+utils.MetaReq+".Test", utils.InfieldSep),
		},
	}
	input[0].ComputePath()
	if err := ar.SetFields(input); err != nil {
		t.Error(err)
	} else if _, err := ar.Vars.FieldAsInterface([]string{"Test"}); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expecting: %+v, received: %+v", utils.ErrNotFound, err)
	}
	input = []*config.FCTemplate{
		{
			Path:      fmt.Sprintf("%s.Test", utils.MetaVars),
			Tag:       fmt.Sprintf("%s.Test", utils.MetaVars),
			Type:      utils.MetaVariable,
			Value:     config.NewRSRParsersMustCompile("~"+utils.MetaReq+".Test", utils.InfieldSep),
			Mandatory: true,
		},
	}
	input[0].ComputePath()
	if err := ar.SetFields(input); err == nil || err.Error() != "NOT_FOUND:"+utils.MetaVars+".Test" {
		t.Errorf("Expecting: %+v, received: %+v", "NOT_FOUND:"+utils.MetaVars+".Test", err)
	}

	//Not found
	input = []*config.FCTemplate{
		{
			Path:      "wrong",
			Tag:       "wrong",
			Type:      utils.MetaVariable,
			Value:     config.NewRSRParsersMustCompile("~*req.Account", utils.InfieldSep),
			Mandatory: true,
		},
	}
	input[0].ComputePath()
	if err := ar.SetFields(input); err == nil || err.Error() != "unsupported field prefix: <wrong> when set field" {
		t.Errorf("Expecting: %+v, received: %+v", "unsupported field prefix: <wrong> when set field", err)
	}

	// MetaHdr/MetaTrl
	input = []*config.FCTemplate{
		{
			Path:  fmt.Sprintf("%s.Account4", utils.MetaVars),
			Tag:   fmt.Sprintf("%s.Account4", utils.MetaVars),
			Type:  utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~"+utils.MetaHdr+".Account", utils.InfieldSep),
		},
	}
	input[0].ComputePath()
	if err := ar.SetFields(input); err != nil {
		t.Error(err)
	} else if val, err := ar.Vars.FieldAsInterface([]string{"Account4"}); err != nil {
		t.Error(err)
	} else if nm, ok := val.([]*utils.DataNode); !ok {
		t.Error("Expecting NM items")
	} else if len(nm) != 1 {
		t.Error("Expecting one item")
	} else if nm[0].Value.Data != "1009" {
		t.Error("Expecting 1009, received: ", nm[0].Value.Data)
	}

	input = []*config.FCTemplate{
		{
			Path:  fmt.Sprintf("%s.Account5", utils.MetaVars),
			Tag:   fmt.Sprintf("%s.Account5", utils.MetaVars),
			Type:  utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~"+utils.MetaTrl+".Account", utils.InfieldSep),
		},
	}
	input[0].ComputePath()
	if err := ar.SetFields(input); err != nil {
		t.Error(err)
	} else if val, err := ar.Vars.FieldAsInterface([]string{"Account5"}); err != nil {
		t.Error(err)
	} else if nm, ok := val.([]*utils.DataNode); !ok {
		t.Error("Expecting NM items")
	} else if len(nm) != 1 {
		t.Error("Expecting one item")
	} else if nm[0].Value.Data != "1009" {
		t.Error("Expecting 1009, received: ", nm[0].Value.Data)
	}
}

func TestAgReqMaxCost(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	agReq := NewAgentRequest(nil, nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)
	// populate request, emulating the way will be done in HTTPAgent
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.CapMaxUsage, PathSlice: []string{utils.CapMaxUsage}}, utils.NewLeafNode("120s"))

	agReq.CGRReply = &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{
		utils.CapMaxUsage: utils.NewLeafNode(120 * time.Second),
	}}

	tplFlds := []*config.FCTemplate{
		{Tag: "MaxUsage",
			Path: utils.MetaRep + utils.NestingSep + "MaxUsage", Type: utils.MetaVariable,
			Filters: []string{"*rsr:~*cgrep.MaxUsage:>0s"},
			Value: config.NewRSRParsersMustCompile(
				"~*cgrep.MaxUsage{*duration_seconds}", utils.InfieldSep)},
	}
	tplFlds[0].ComputePath()
	eMp := utils.NewOrderedNavigableMap()

	eMp.SetAsSlice(&utils.FullPath{Path: "MaxUsage", PathSlice: []string{"MaxUsage"}}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "120"}}})

	if err := agReq.SetFields(tplFlds); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(agReq.Reply, eMp) {
		t.Log(utils.ToJSON(eMp.GetOrder()))
		t.Log(utils.ToJSON(agReq.Reply.GetOrder()))
		t.Errorf("expecting: %+v,\n received: %+v", eMp, agReq.Reply)
	}
}

func TestAgReqParseFieldDiameter(t *testing.T) {
	//creater diameter message
	m := diam.NewRequest(diam.CreditControl, 4, nil)
	m.NewAVP("Session-Id", avp.Mbit, 0, datatype.UTF8String("simuhuawei;1449573472;00002"))
	m.NewAVP("Subscription-Id", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(450, avp.Mbit, 0, datatype.Enumerated(2)),              // Subscription-Id-Type
			diam.NewAVP(444, avp.Mbit, 0, datatype.UTF8String("208708000004")), // Subscription-Id-Data
			diam.NewAVP(avp.ValueDigits, avp.Mbit, 0, datatype.Integer64(20000)),
		}})
	//create diameterDataProvider
	dP := newDADataProvider(nil, m)
	cfg := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, true),
		config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	//pass the data provider to agent request
	agReq := NewAgentRequest(dP, nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)

	tplFlds := []*config.FCTemplate{
		{Tag: "MandatoryFalse",
			Path: "MandatoryFalse", Type: utils.MetaComposed,
			Value:     config.NewRSRParsersMustCompile("~*req.MandatoryFalse", utils.InfieldSep),
			Mandatory: false},
		{Tag: "MandatoryTrue",
			Path: "MandatoryTrue", Type: utils.MetaComposed,
			Value:     config.NewRSRParsersMustCompile("~*req.MandatoryTrue", utils.InfieldSep),
			Mandatory: true},
		{Tag: "Session-Id", Filters: []string{},
			Path: "Session-Id", Type: utils.MetaComposed,
			Value:     config.NewRSRParsersMustCompile("~*req.Session-Id", utils.InfieldSep),
			Mandatory: true},
	}
	for _, v := range tplFlds {
		v.ComputePath()
	}
	expected := ""
	if out, err := agReq.ParseField(tplFlds[0]); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(out, expected) {
		t.Errorf("expecting: <%+v>, received: <%+v>", expected, out)
	}
	if _, err := agReq.ParseField(tplFlds[1]); err == nil ||
		err.Error() != "Empty source value for fieldID: <MandatoryTrue>" {
		t.Error(err)
	}
	expected = "simuhuawei;1449573472;00002"
	if out, err := agReq.ParseField(tplFlds[2]); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(out, expected) {
		t.Errorf("expecting: <%+v>, received: <%+v>", expected, out)
	}
}

func TestAgReqParseFieldRadius(t *testing.T) {
	//creater radius message
	pkt := radigo.NewPacket(radigo.AccountingRequest, 1, dictRad, coder, "CGRateS.org")
	if err := pkt.AddAVPWithName("User-Name", "flopsy", ""); err != nil {
		t.Error(err)
	}
	if err := pkt.AddAVPWithName("Cisco-NAS-Port", "CGR1", "Cisco"); err != nil {
		t.Error(err)
	}
	//create radiusDataProvider
	dP := newRADataProvider(pkt)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	//pass the data provider to agent request
	agReq := NewAgentRequest(dP, nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)
	tplFlds := []*config.FCTemplate{
		{Tag: "MandatoryFalse",
			Path: "MandatoryFalse", Type: utils.MetaComposed,
			Value:     config.NewRSRParsersMustCompile("~*req.MandatoryFalse", utils.InfieldSep),
			Mandatory: false},
		{Tag: "MandatoryTrue",
			Path: "MandatoryTrue", Type: utils.MetaComposed,
			Value:     config.NewRSRParsersMustCompile("~*req.MandatoryTrue", utils.InfieldSep),
			Mandatory: true},
	}
	for _, v := range tplFlds {
		v.ComputePath()
	}
	expected := ""
	if out, err := agReq.ParseField(tplFlds[0]); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(out, expected) {
		t.Errorf("expecting: <%+v>, received: <%+v>", expected, out)
	}
	if _, err := agReq.ParseField(tplFlds[1]); err == nil ||
		err.Error() != "Empty source value for fieldID: <MandatoryTrue>" {
		t.Error(err)
	}
}

func TestAgReqParseFieldHttpUrl(t *testing.T) {
	//creater radius message
	br := bufio.NewReader(strings.NewReader(`GET /cdr?request_type=MOSMS_CDR&timestamp=2008-08-15%2017:49:21&message_date=2008-08-15%2017:49:21&transactionid=100744&CDR_ID=123456&carrierid=1&mcc=222&mnc=10&imsi=235180000000000&msisdn=%2B4977000000000&destination=%2B497700000001&message_status=0&IOT=0&service_id=1 HTTP/1.1
Host: api.cgrates.org

`))
	req, err := http.ReadRequest(br)
	if err != nil {
		t.Error(err)
	}
	//create radiusDataProvider
	dP, _ := newHTTPUrlDP(req)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	//pass the data provider to agent request
	agReq := NewAgentRequest(dP, nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)
	tplFlds := []*config.FCTemplate{
		{Tag: "MandatoryFalse",
			Path: "MandatoryFalse", Type: utils.MetaComposed,
			Value:     config.NewRSRParsersMustCompile("~*req.MandatoryFalse", utils.InfieldSep),
			Mandatory: false},
		{Tag: "MandatoryTrue",
			Path: "MandatoryTrue", Type: utils.MetaComposed,
			Value:     config.NewRSRParsersMustCompile("~*req.MandatoryTrue", utils.InfieldSep),
			Mandatory: true},
	}
	expected := ""
	if out, err := agReq.ParseField(tplFlds[0]); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(out, expected) {
		t.Errorf("expecting: <%+v>, received: <%+v>", expected, out)
	}

	if _, err := agReq.ParseField(tplFlds[1]); err == nil ||
		err.Error() != "Empty source value for fieldID: <MandatoryTrue>" {
		t.Error(err)
	}
}

func TestAgReqParseFieldHttpXml(t *testing.T) {
	//creater radius message
	body := `<complete-success-notification callid="109870">
	<createtime>2005-08-26T14:16:42</createtime>
	<connecttime>2005-08-26T14:16:56</connecttime>
	<endtime>2005-08-26T14:17:34</endtime>
	<reference>My Call Reference</reference>
	<userid>386</userid>
	<username>sampleusername</username>
	<customerid>1</customerid>
	<companyname>Conecto LLC</companyname>
	<totalcost amount="0.21" currency="USD">US$0.21</totalcost>
	<hasrecording>yes</hasrecording>
	<hasvoicemail>no</hasvoicemail>
	<agenttotalcost amount="0.13" currency="USD">US$0.13</agenttotalcost>
	<agentid>44</agentid>
	<callleg calllegid="222146">
		<number>+441624828505</number>
		<description>Isle of Man</description>
		<seconds>38</seconds>
		<perminuterate amount="0.0200" currency="USD">US$0.0200</perminuterate>
		<cost amount="0.0140" currency="USD">US$0.0140</cost>
		<agentperminuterate amount="0.0130" currency="USD">US$0.0130</agentperminuterate>
		<agentcost amount="0.0082" currency="USD">US$0.0082</agentcost>
	</callleg>
	<callleg calllegid="222147">
		<number>+44 7624 494075</number>
		<description>Isle of Man</description>
		<seconds>37</seconds>
		<perminuterate amount="0.2700" currency="USD">US$0.2700</perminuterate>
		<cost amount="0.1890" currency="USD">US$0.1890</cost>
		<agentperminuterate amount="0.1880" currency="USD">US$0.1880</agentperminuterate>
		<agentcost amount="0.1159" currency="USD">US$0.1159</agentcost>
	</callleg>
</complete-success-notification>
`
	req, err := http.NewRequest("POST", "http://localhost:8080/", bytes.NewBuffer([]byte(body)))
	if err != nil {
		t.Error(err)
	}
	//create radiusDataProvider
	dP, _ := newHTTPXmlDP(req)
	cfg := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, true),
		config.CgrConfig().CacheCfg(), nil)

	filterS := engine.NewFilterS(cfg, nil, dm)
	//pass the data provider to agent request
	agReq := NewAgentRequest(dP, nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)
	tplFlds := []*config.FCTemplate{
		{Tag: "MandatoryFalse",
			Path: "MandatoryFalse", Type: utils.MetaComposed,
			Value:     config.NewRSRParsersMustCompile("~*req.MandatoryFalse", utils.InfieldSep),
			Mandatory: false},
		{Tag: "MandatoryTrue",
			Path: "MandatoryTrue", Type: utils.MetaComposed,
			Value:     config.NewRSRParsersMustCompile("~*req.MandatoryTrue", utils.InfieldSep),
			Mandatory: true},
	}
	expected := ""
	if out, err := agReq.ParseField(tplFlds[0]); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(out, expected) {
		t.Errorf("expecting: <%+v>, received: <%+v>", expected, out)
	}
	if _, err := agReq.ParseField(tplFlds[1]); err == nil ||
		err.Error() != "Empty source value for fieldID: <MandatoryTrue>" {
		t.Error(err)
	}
}

func TestAgReqEmptyFilter(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	agReq := NewAgentRequest(nil, nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)
	// populate request, emulating the way will be done in HTTPAgent
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.CGRID, PathSlice: []string{utils.CGRID}},
		utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()))
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.AccountField, PathSlice: []string{utils.AccountField}}, utils.NewLeafNode("1001"))
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.Destination, PathSlice: []string{utils.Destination}}, utils.NewLeafNode("1002"))

	tplFlds := []*config.FCTemplate{
		{Tag: "Tenant", Filters: []string{},
			Path: utils.MetaCgrep + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep)},

		{Tag: "Account", Filters: []string{},
			Path: utils.MetaCgrep + utils.NestingSep + utils.AccountField, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*cgreq.Account", utils.InfieldSep)},
		{Tag: "Destination", Filters: []string{},
			Path: utils.MetaCgrep + utils.NestingSep + utils.Destination, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*cgreq.Destination", utils.InfieldSep)},
	}
	for _, v := range tplFlds {
		v.ComputePath()
	}
	eMp := &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{}}
	eMp.Set([]string{utils.Tenant}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "cgrates.org"}}})
	eMp.Set([]string{utils.AccountField}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "1001"}}})
	eMp.Set([]string{utils.Destination}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "1002"}}})

	if err := agReq.SetFields(tplFlds); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(agReq.CGRReply, eMp) {
		t.Errorf("expecting: %+v,\n received: %+v", eMp, agReq.CGRReply)
	}
}

func TestAgReqMetaExponent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, true),
		config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	agReq := NewAgentRequest(nil, nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)
	agReq.CGRRequest.Set(&utils.FullPath{Path: "Value", PathSlice: []string{"Value"}}, utils.NewLeafNode("2"))
	agReq.CGRRequest.Set(&utils.FullPath{Path: "Exponent", PathSlice: []string{"Exponent"}}, utils.NewLeafNode("2"))

	tplFlds := []*config.FCTemplate{
		{Tag: "TestExpo", Filters: []string{},
			Path: utils.MetaCgrep + utils.NestingSep + "TestExpo", Type: utils.MetaValueExponent,
			Value: config.NewRSRParsersMustCompile("~*cgreq.Value;~*cgreq.Exponent", utils.InfieldSep)},
	}
	tplFlds[0].ComputePath()
	eMp := &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{}}
	eMp.Set([]string{"TestExpo"}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "200"}}})

	if err := agReq.SetFields(tplFlds); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(agReq.CGRReply, eMp) {
		t.Errorf("expecting: %+v,\n received: %+v", eMp, agReq.CGRReply)
	}
}

func TestAgReqFieldAsNone(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	agReq := NewAgentRequest(nil, nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)
	// populate request, emulating the way will be done in HTTPAgent
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.ToR, PathSlice: []string{utils.ToR}}, utils.NewLeafNode(utils.MetaVoice))
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.AccountField, PathSlice: []string{utils.AccountField}}, utils.NewLeafNode("1001"))
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.Destination, PathSlice: []string{utils.Destination}}, utils.NewLeafNode("1002"))

	tplFlds := []*config.FCTemplate{
		{Tag: "Tenant",
			Path: utils.MetaCgrep + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep)},
		{Tag: "Account",
			Path: utils.MetaCgrep + utils.NestingSep + utils.AccountField, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*cgreq.Account", utils.InfieldSep)},
		{Type: utils.MetaNone, Blocker: true},
		{Tag: "Destination",
			Path: utils.MetaCgrep + utils.NestingSep + utils.Destination, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*cgreq.Destination", utils.InfieldSep)},
	}
	for _, v := range tplFlds {
		v.ComputePath()
	}
	eMp := &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{}}
	eMp.Set([]string{utils.Tenant}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "cgrates.org"}}})
	eMp.Set([]string{utils.AccountField}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "1001"}}})
	if err := agReq.SetFields(tplFlds); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(agReq.CGRReply, eMp) {
		t.Errorf("expecting: %+v,\n received: %+v", eMp, agReq.CGRReply)
	}
}

func TestAgReqFieldAsNone2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, true),
		config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	agReq := NewAgentRequest(nil, nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)
	// populate request, emulating the way will be done in HTTPAgent
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.ToR, PathSlice: []string{utils.ToR}}, utils.NewLeafNode(utils.MetaVoice))
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.AccountField, PathSlice: []string{utils.AccountField}}, utils.NewLeafNode("1001"))
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.Destination, PathSlice: []string{utils.Destination}}, utils.NewLeafNode("1002"))

	tplFlds := []*config.FCTemplate{
		{Tag: "Tenant",
			Path: utils.MetaCgrep + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep)},
		{Tag: "Account",
			Path: utils.MetaCgrep + utils.NestingSep + utils.AccountField, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*cgreq.Account", utils.InfieldSep)},
		{Type: utils.MetaNone},
		{Tag: "Destination",
			Path: utils.MetaCgrep + utils.NestingSep + utils.Destination, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*cgreq.Destination", utils.InfieldSep)},
	}
	for _, v := range tplFlds {
		v.ComputePath()
	}
	eMp := &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{}}
	eMp.Set([]string{utils.Tenant}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "cgrates.org"}}})
	eMp.Set([]string{utils.AccountField}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "1001"}}})
	eMp.Set([]string{utils.Destination}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "1002"}}})
	if err := agReq.SetFields(tplFlds); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(agReq.CGRReply, eMp) {
		t.Errorf("expecting: %+v,\n received: %+v", eMp, agReq.CGRReply)
	}
}

func TestAgReqSetField2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	agReq := NewAgentRequest(nil, nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)
	// populate request, emulating the way will be done in HTTPAgent
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.ToR, PathSlice: []string{utils.ToR}}, utils.NewLeafNode(utils.MetaVoice))
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.AccountField, PathSlice: []string{utils.AccountField}}, utils.NewLeafNode("1001"))
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.Destination, PathSlice: []string{utils.Destination}}, utils.NewLeafNode("1002"))
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.AnswerTime, PathSlice: []string{utils.AnswerTime}}, utils.NewLeafNode(
		time.Date(2013, 12, 30, 15, 0, 1, 0, time.UTC)))
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.RequestType, PathSlice: []string{utils.RequestType}}, utils.NewLeafNode(utils.MetaPrepaid))

	agReq.CGRReply = &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{}}

	tplFlds := []*config.FCTemplate{
		{Tag: "Tenant",
			Path: utils.MetaCgrep + utils.NestingSep + utils.Tenant, Type: utils.MetaComposed,
			Value: config.NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep)},
		{Tag: "Account",
			Path: utils.MetaCgrep + utils.NestingSep + utils.AccountField, Type: utils.MetaComposed,
			Value: config.NewRSRParsersMustCompile("~*cgreq.Account", utils.InfieldSep)},
		{Tag: "Destination",
			Path: utils.MetaCgrep + utils.NestingSep + utils.Destination, Type: utils.MetaComposed,
			Value: config.NewRSRParsersMustCompile("~*cgreq.Destination", utils.InfieldSep)},
		{Tag: "Usage",
			Path: utils.MetaCgrep + utils.NestingSep + utils.Usage, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("30s", utils.InfieldSep)},
		{Tag: "CalculatedUsage",
			Path: utils.MetaCgrep + utils.NestingSep + "CalculatedUsage",
			Type: "*difference", Value: config.NewRSRParsersMustCompile("~*cgreq.AnswerTime;~*cgrep.Usage", utils.InfieldSep),
		},
	}
	for _, v := range tplFlds {
		v.ComputePath()
	}
	eMp := &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{}}
	eMp.Set([]string{utils.Tenant}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "cgrates.org"}}})
	eMp.Set([]string{utils.AccountField}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "1001"}}})
	eMp.Set([]string{utils.Destination}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "1002"}}})
	eMp.Set([]string{"Usage"}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "30s"}}})
	eMp.Set([]string{"CalculatedUsage"}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: time.Date(2013, 12, 30, 14, 59, 31, 0, time.UTC)}}})

	if err := agReq.SetFields(tplFlds); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(agReq.CGRReply, eMp) {
		t.Errorf("expecting: %+v,\n received: %+v", eMp, agReq.CGRReply)
	}
}

func TestAgReqFieldAsInterface(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, true),
		config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	agReq := NewAgentRequest(nil, nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)
	// populate request, emulating the way will be done in HTTPAgent
	agReq.CGRRequest = utils.NewOrderedNavigableMap()
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.Usage, PathSlice: []string{utils.Usage}}, &utils.DataNode{Type: utils.NMSliceType, Slice: []*utils.DataNode{{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: 3 * time.Minute}}}})
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.ToR, PathSlice: []string{utils.ToR}}, &utils.DataNode{Type: utils.NMSliceType, Slice: []*utils.DataNode{{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: utils.MetaVoice}}}})
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.AccountField, PathSlice: []string{utils.AccountField}}, utils.NewLeafNode("1001"))
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.Destination, PathSlice: []string{utils.Destination}}, utils.NewLeafNode("1002"))

	path := []string{utils.MetaCgreq, utils.Usage}
	var expVal interface{}
	expVal = 3 * time.Minute
	if rply, err := agReq.FieldAsInterface(path); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rply, expVal) {
		t.Errorf("Expected %v , received: %v", utils.ToJSON(expVal), utils.ToJSON(rply))
	}

	path = []string{utils.MetaCgreq, utils.ToR}
	expVal = utils.MetaVoice
	if rply, err := agReq.FieldAsInterface(path); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rply, expVal) {
		t.Errorf("Expected %v , received: %v", utils.ToJSON(expVal), utils.ToJSON(rply))
	}

	path = []string{utils.MetaCgreq, utils.AccountField}
	expVal = "1001"
	if rply, err := agReq.FieldAsInterface(path); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rply, expVal) {
		t.Errorf("Expected %v , received: %v", utils.ToJSON(expVal), utils.ToJSON(rply))
	}

	path = []string{utils.MetaCgreq, utils.Destination}
	expVal = "1002"
	if rply, err := agReq.FieldAsInterface(path); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rply, expVal) {
		t.Errorf("Expected %v , received: %v", utils.ToJSON(expVal), utils.ToJSON(rply))
	}
}

func TestAgReqNewARWithCGRRplyAndRply(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)

	rply := utils.NewOrderedNavigableMap()
	rply.Set(&utils.FullPath{
		Path:      "FirstLevel.SecondLevel.Fld1",
		PathSlice: []string{"FirstLevel", "SecondLevel", "Fld1"}}, utils.NewLeafNode("Val1"))
	cgrRply := &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{
		utils.CapAttributes: {Type: utils.NMMapType, Map: map[string]*utils.DataNode{
			"PaypalAccount": utils.NewLeafNode("cgrates@paypal.com"),
		}},
		utils.CapMaxUsage: utils.NewLeafNode(120 * time.Second),
		utils.Error:       utils.NewLeafNode(""),
	}}

	agReq := NewAgentRequest(nil, nil, cgrRply, rply, nil, nil, "cgrates.org", "", filterS, nil, nil)

	tplFlds := []*config.FCTemplate{
		{Tag: "Fld1",
			Path: utils.MetaCgreq + utils.NestingSep + "Fld1", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*rep.FirstLevel.SecondLevel.Fld1", utils.InfieldSep)},
		{Tag: "Fld2",
			Path: utils.MetaCgreq + utils.NestingSep + "Fld2", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*cgrep.Attributes.PaypalAccount", utils.InfieldSep)},
	}
	for _, v := range tplFlds {
		v.ComputePath()
	}

	eMp := utils.NewOrderedNavigableMap()
	eMp.SetAsSlice(&utils.FullPath{Path: "Fld1", PathSlice: []string{"Fld1"}}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "Val1"}}})
	eMp.SetAsSlice(&utils.FullPath{Path: "Fld2", PathSlice: []string{"Fld2"}}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "cgrates@paypal.com"}}})

	if err := agReq.SetFields(tplFlds); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(agReq.CGRRequest, eMp) {
		t.Errorf("expecting: %+v,\n received: %+v", eMp, agReq.CGRRequest)
	}
}

func TestAgReqSetCGRReplyWithError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, true),
		config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)

	rply := utils.NewOrderedNavigableMap()
	rply.Set(&utils.FullPath{
		Path:      "FirstLevel.SecondLevel.Fld1",
		PathSlice: []string{"FirstLevel", "SecondLevel", "Fld1"}}, utils.NewLeafNode("Val1"))
	agReq := NewAgentRequest(nil, nil, nil, rply, nil, nil, "cgrates.org", "", filterS, nil, nil)

	agReq.setCGRReply(nil, utils.ErrNotFound)

	tplFlds := []*config.FCTemplate{
		{Tag: "Fld1",
			Path: utils.MetaCgreq + utils.NestingSep + "Fld1", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*rep.FirstLevel.SecondLevel.Fld1", utils.InfieldSep)},
		{Tag: "Fld2",
			Path: utils.MetaCgreq + utils.NestingSep + "Fld2", Type: utils.MetaVariable,
			Value:     config.NewRSRParsersMustCompile("~*cgrep.Attributes.PaypalAccount", utils.InfieldSep),
			Mandatory: true},
	}
	for _, v := range tplFlds {
		v.ComputePath()
	}
	if err := agReq.SetFields(tplFlds); err == nil ||
		err.Error() != "NOT_FOUND:Fld2" {
		t.Error(err)
	}
}

type myEv map[string]*utils.DataNode

func (ev myEv) AsNavigableMap() map[string]*utils.DataNode {
	return ev
}

func TestAgReqSetCGRReplyWithoutError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)

	rply := utils.NewOrderedNavigableMap()
	rply.Set(&utils.FullPath{
		Path:      "FirstLevel.SecondLevel.Fld1",
		PathSlice: []string{"FirstLevel", "SecondLevel", "Fld1"}}, utils.NewLeafNode("Val1"))

	myEv := myEv{
		utils.CapAttributes: &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{
			"PaypalAccount": utils.NewLeafNode("cgrates@paypal.com"),
		}},
		utils.CapMaxUsage: utils.NewLeafNode(120 * time.Second),
		utils.Error:       utils.NewLeafNode(""),
	}

	agReq := NewAgentRequest(nil, nil, nil, rply, nil,
		nil, "cgrates.org", "", filterS, nil, nil)

	agReq.setCGRReply(myEv, nil)

	tplFlds := []*config.FCTemplate{
		{Tag: "Fld1",
			Path: utils.MetaCgreq + utils.NestingSep + "Fld1", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*rep.FirstLevel.SecondLevel.Fld1", utils.InfieldSep)},
		{Tag: "Fld2",
			Path: utils.MetaCgreq + utils.NestingSep + "Fld2", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*cgrep.Attributes.PaypalAccount", utils.InfieldSep)},
	}
	for _, v := range tplFlds {
		v.ComputePath()
	}
	eMp := utils.NewOrderedNavigableMap()
	eMp.SetAsSlice(&utils.FullPath{Path: "Fld1", PathSlice: []string{"Fld1"}}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "Val1"}}})
	eMp.SetAsSlice(&utils.FullPath{Path: "Fld2", PathSlice: []string{"Fld2"}}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "cgrates@paypal.com"}}})

	if err := agReq.SetFields(tplFlds); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(agReq.CGRRequest, eMp) {
		t.Log(utils.ToJSON(eMp.GetOrder()))
		t.Log(utils.ToJSON(agReq.CGRRequest.GetOrder()))
		t.Errorf("expecting: %+v,\n received: %+v", eMp, agReq.CGRRequest)
	}
}

func TestAgReqParseFieldMetaCCUsage(t *testing.T) {
	//creater diameter message
	m := diam.NewRequest(diam.CreditControl, 4, nil)
	m.NewAVP("Session-Id", avp.Mbit, 0, datatype.UTF8String("simuhuawei;1449573472;00002"))
	m.NewAVP("Subscription-Id", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(450, avp.Mbit, 0, datatype.Enumerated(2)),              // Subscription-Id-Type
			diam.NewAVP(444, avp.Mbit, 0, datatype.UTF8String("208708000004")), // Subscription-Id-Data
			diam.NewAVP(avp.ValueDigits, avp.Mbit, 0, datatype.Integer64(20000)),
		}})
	//create diameterDataProvider
	dP := newDADataProvider(nil, m)
	cfg := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, true),
		config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	//pass the data provider to agent request
	agReq := NewAgentRequest(dP, nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)

	tplFlds := []*config.FCTemplate{
		{Tag: "CCUsage", Filters: []string{},
			Path: "CCUsage", Type: utils.MetaCCUsage,
			Value:     config.NewRSRParsersMustCompile("~*req.Session-Id", utils.InfieldSep),
			Mandatory: true},
	}
	tplFlds[0].ComputePath()

	if _, err := agReq.ParseField(tplFlds[0]); err == nil ||
		err.Error() != `invalid arguments <[{"Rules":"~*req.Session-Id"}]> to *cc_usage` {
		t.Error(err)
	}

	tplFlds = []*config.FCTemplate{
		{Tag: "CCUsage", Filters: []string{},
			Path: "CCUsage", Type: utils.MetaCCUsage,
			Value:     config.NewRSRParsersMustCompile("~*req.Session-Id;12s;12s", utils.InfieldSep),
			Mandatory: true},
	}
	tplFlds[0].ComputePath()
	if _, err := agReq.ParseField(tplFlds[0]); err == nil ||
		err.Error() != `invalid requestNumber <simuhuawei;1449573472;00002> to *cc_usage` {
		t.Error(err)
	}

	tplFlds = []*config.FCTemplate{
		{Tag: "CCUsage", Filters: []string{},
			Path: "CCUsage", Type: utils.MetaCCUsage,
			Value:     config.NewRSRParsersMustCompile("10;~*req.Session-Id;12s", utils.InfieldSep),
			Mandatory: true},
	}
	tplFlds[0].ComputePath()
	if _, err := agReq.ParseField(tplFlds[0]); err == nil ||
		err.Error() != `invalid usedCCTime <simuhuawei;1449573472;00002> to *cc_usage` {
		t.Error(err)
	}

	tplFlds = []*config.FCTemplate{
		{Tag: "CCUsage", Filters: []string{},
			Path: "CCUsage", Type: utils.MetaCCUsage,
			Value:     config.NewRSRParsersMustCompile("10;12s;~*req.Session-Id", utils.InfieldSep),
			Mandatory: true},
	}
	tplFlds[0].ComputePath()
	if _, err := agReq.ParseField(tplFlds[0]); err == nil ||
		err.Error() != `invalid debitInterval <simuhuawei;1449573472;00002> to *cc_usage` {
		t.Error(err)
	}

	tplFlds = []*config.FCTemplate{
		{Tag: "CCUsage", Filters: []string{},
			Path: "CCUsage", Type: utils.MetaCCUsage,
			Value:     config.NewRSRParsersMustCompile("3;10s;5s", utils.InfieldSep),
			Mandatory: true},
	}
	tplFlds[0].ComputePath()
	//5s*2 + 10s
	expected := 20 * time.Second
	if out, err := agReq.ParseField(tplFlds[0]); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(out, expected) {
		t.Errorf("expecting: <%+v>, received: <%+v>", expected, out)
	}
}

func TestAgReqParseFieldMetaUsageDifference(t *testing.T) {
	//creater diameter message
	m := diam.NewRequest(diam.CreditControl, 4, nil)
	m.NewAVP("Session-Id", avp.Mbit, 0, datatype.UTF8String("simuhuawei;1449573472;00002"))
	m.NewAVP("Subscription-Id", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(450, avp.Mbit, 0, datatype.Enumerated(2)),              // Subscription-Id-Type
			diam.NewAVP(444, avp.Mbit, 0, datatype.UTF8String("208708000004")), // Subscription-Id-Data
			diam.NewAVP(avp.ValueDigits, avp.Mbit, 0, datatype.Integer64(20000)),
		}})
	//create diameterDataProvider
	dP := newDADataProvider(nil, m)
	cfg := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, true),
		config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	//pass the data provider to agent request
	agReq := NewAgentRequest(dP, nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)

	tplFlds := []*config.FCTemplate{
		{Tag: "Usage", Filters: []string{},
			Path: "Usage", Type: utils.MetaUsageDifference,
			Value:     config.NewRSRParsersMustCompile("~*req.Session-Id", utils.InfieldSep),
			Mandatory: true},
	}
	if _, err := agReq.ParseField(tplFlds[0]); err == nil ||
		err.Error() != `invalid arguments <[{"Rules":"~*req.Session-Id"}]> to *usage_difference` {
		t.Error(err)
	}

	tplFlds = []*config.FCTemplate{
		{Tag: "Usage", Filters: []string{},
			Path: "Usage", Type: utils.MetaUsageDifference,
			Value:     config.NewRSRParsersMustCompile("1560325161;~*req.Session-Id", utils.InfieldSep),
			Mandatory: true},
	}
	if _, err := agReq.ParseField(tplFlds[0]); err == nil ||
		err.Error() != `Unsupported time format` {
		t.Error(err)
	}

	tplFlds = []*config.FCTemplate{
		{Tag: "Usage", Filters: []string{},
			Path: "Usage", Type: utils.MetaUsageDifference,
			Value:     config.NewRSRParsersMustCompile("~*req.Session-Id;1560325161", utils.InfieldSep),
			Mandatory: true},
	}
	if _, err := agReq.ParseField(tplFlds[0]); err == nil ||
		err.Error() != `Unsupported time format` {
		t.Error(err)
	}

	tplFlds = []*config.FCTemplate{
		{Tag: "Usage", Filters: []string{},
			Path: "Usage", Type: utils.MetaUsageDifference,
			Value:     config.NewRSRParsersMustCompile("1560325161;1560325151", utils.InfieldSep),
			Mandatory: true},
	}
	expected := "10s"
	if out, err := agReq.ParseField(tplFlds[0]); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(out, expected) {
		t.Errorf("expecting: <%+v>, received: <%+v>", expected, out)
	}
}

func TestAgReqParseFieldMetaSum(t *testing.T) {
	//creater diameter message
	m := diam.NewRequest(diam.CreditControl, 4, nil)
	m.NewAVP("Session-Id", avp.Mbit, 0, datatype.UTF8String("simuhuawei;1449573472;00002"))
	m.NewAVP("Subscription-Id", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(450, avp.Mbit, 0, datatype.Enumerated(2)),              // Subscription-Id-Type
			diam.NewAVP(444, avp.Mbit, 0, datatype.UTF8String("208708000004")), // Subscription-Id-Data
			diam.NewAVP(avp.ValueDigits, avp.Mbit, 0, datatype.Integer64(20000)),
		}})
	//create diameterDataProvider
	dP := newDADataProvider(nil, m)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	//pass the data provider to agent request
	agReq := NewAgentRequest(dP, nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)

	tplFlds := []*config.FCTemplate{
		{Tag: "Sum", Filters: []string{},
			Path: "Sum", Type: utils.MetaSum,
			Value:     config.NewRSRParsersMustCompile("15;~*req.Session-Id", utils.InfieldSep),
			Mandatory: true},
	}
	if _, err := agReq.ParseField(tplFlds[0]); err == nil ||
		err.Error() != `strconv.ParseInt: parsing "simuhuawei;1449573472;00002": invalid syntax` {
		t.Error(err)
	}

	tplFlds = []*config.FCTemplate{
		{Tag: "Sum", Filters: []string{},
			Path: "Sum", Type: utils.MetaSum,
			Value:     config.NewRSRParsersMustCompile("15;15", utils.InfieldSep),
			Mandatory: true},
	}
	expected := int64(30)
	if out, err := agReq.ParseField(tplFlds[0]); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(out, expected) {
		t.Errorf("expecting: <%+v>, %T received: <%+v> %T", expected, expected, out, out)
	}
}

func TestAgReqParseFieldMetaDifference(t *testing.T) {
	//creater diameter message
	m := diam.NewRequest(diam.CreditControl, 4, nil)
	m.NewAVP("Session-Id", avp.Mbit, 0, datatype.UTF8String("simuhuawei;1449573472;00002"))
	m.NewAVP("Subscription-Id", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(450, avp.Mbit, 0, datatype.Enumerated(2)),              // Subscription-Id-Type
			diam.NewAVP(444, avp.Mbit, 0, datatype.UTF8String("208708000004")), // Subscription-Id-Data
			diam.NewAVP(avp.ValueDigits, avp.Mbit, 0, datatype.Integer64(20000)),
		}})
	//create diameterDataProvider
	dP := newDADataProvider(nil, m)
	cfg := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, true),
		config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	//pass the data provider to agent request
	agReq := NewAgentRequest(dP, nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)

	tplFlds := []*config.FCTemplate{
		{Tag: "Diff", Filters: []string{},
			Path: "Diff", Type: utils.MetaDifference,
			Value:     config.NewRSRParsersMustCompile("15;~*req.Session-Id", utils.InfieldSep),
			Mandatory: true},
	}
	if _, err := agReq.ParseField(tplFlds[0]); err == nil ||
		err.Error() != `strconv.ParseInt: parsing "simuhuawei;1449573472;00002": invalid syntax` {
		t.Error(err)
	}

	tplFlds = []*config.FCTemplate{
		{Tag: "Diff", Filters: []string{},
			Path: "Diff", Type: utils.MetaDifference,
			Value:     config.NewRSRParsersMustCompile("15;12;2", utils.InfieldSep),
			Mandatory: true},
	}
	expected := int64(1)
	if out, err := agReq.ParseField(tplFlds[0]); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(out, expected) {
		t.Errorf("expecting: <%+v>, %T received: <%+v> %T", expected, expected, out, out)
	}
}

func TestAgReqParseFieldMetaMultiply(t *testing.T) {
	//creater diameter message
	m := diam.NewRequest(diam.CreditControl, 4, nil)
	m.NewAVP("Session-Id", avp.Mbit, 0, datatype.UTF8String("simuhuawei;1449573472;00002"))
	m.NewAVP("Subscription-Id", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(450, avp.Mbit, 0, datatype.Enumerated(2)),              // Subscription-Id-Type
			diam.NewAVP(444, avp.Mbit, 0, datatype.UTF8String("208708000004")), // Subscription-Id-Data
			diam.NewAVP(avp.ValueDigits, avp.Mbit, 0, datatype.Integer64(20000)),
		}})
	//create diameterDataProvider
	dP := newDADataProvider(nil, m)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	//pass the data provider to agent request
	agReq := NewAgentRequest(dP, nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)

	tplFlds := []*config.FCTemplate{
		{Tag: "Multiply", Filters: []string{},
			Path: "Multiply", Type: utils.MetaMultiply,
			Value:     config.NewRSRParsersMustCompile("15;~*req.Session-Id", utils.InfieldSep),
			Mandatory: true},
	}
	if _, err := agReq.ParseField(tplFlds[0]); err == nil ||
		err.Error() != `strconv.ParseInt: parsing "simuhuawei;1449573472;00002": invalid syntax` {
		t.Error(err)
	}

	tplFlds = []*config.FCTemplate{
		{Tag: "Multiply", Filters: []string{},
			Path: "Multiply", Type: utils.MetaMultiply,
			Value:     config.NewRSRParsersMustCompile("15;15", utils.InfieldSep),
			Mandatory: true},
	}
	expected := int64(225)
	if out, err := agReq.ParseField(tplFlds[0]); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(out, expected) {
		t.Errorf("expecting: <%+v>, %T received: <%+v> %T", expected, expected, out, out)
	}
}

func TestAgReqParseFieldMetaDivide(t *testing.T) {
	//creater diameter message
	m := diam.NewRequest(diam.CreditControl, 4, nil)
	m.NewAVP("Session-Id", avp.Mbit, 0, datatype.UTF8String("simuhuawei;1449573472;00002"))
	m.NewAVP("Subscription-Id", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(450, avp.Mbit, 0, datatype.Enumerated(2)),              // Subscription-Id-Type
			diam.NewAVP(444, avp.Mbit, 0, datatype.UTF8String("208708000004")), // Subscription-Id-Data
			diam.NewAVP(avp.ValueDigits, avp.Mbit, 0, datatype.Integer64(20000)),
		}})
	//create diameterDataProvider
	dP := newDADataProvider(nil, m)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	//pass the data provider to agent request
	agReq := NewAgentRequest(dP, nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)

	tplFlds := []*config.FCTemplate{
		{Tag: "Divide", Filters: []string{},
			Path: "Divide", Type: utils.MetaDivide,
			Value:     config.NewRSRParsersMustCompile("15;~*req.Session-Id", utils.InfieldSep),
			Mandatory: true},
	}
	if _, err := agReq.ParseField(tplFlds[0]); err == nil ||
		err.Error() != `strconv.ParseInt: parsing "simuhuawei;1449573472;00002": invalid syntax` {
		t.Error(err)
	}

	tplFlds = []*config.FCTemplate{
		{Tag: "Divide", Filters: []string{},
			Path: "Divide", Type: utils.MetaDivide,
			Value:     config.NewRSRParsersMustCompile("15;3", utils.InfieldSep),
			Mandatory: true},
	}
	expected := int64(5)
	if out, err := agReq.ParseField(tplFlds[0]); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(out, expected) {
		t.Errorf("expecting: <%+v>, %T received: <%+v> %T", expected, expected, out, out)
	}
}

func TestAgReqParseFieldMetaValueExponent(t *testing.T) {
	//creater diameter message
	m := diam.NewRequest(diam.CreditControl, 4, nil)
	m.NewAVP("Session-Id", avp.Mbit, 0, datatype.UTF8String("simuhuawei;1449573472;00002"))
	m.NewAVP("Subscription-Id", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(450, avp.Mbit, 0, datatype.Enumerated(2)),              // Subscription-Id-Type
			diam.NewAVP(444, avp.Mbit, 0, datatype.UTF8String("208708000004")), // Subscription-Id-Data
			diam.NewAVP(avp.ValueDigits, avp.Mbit, 0, datatype.Integer64(20000)),
		}})
	//create diameterDataProvider
	dP := newDADataProvider(nil, m)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	//pass the data provider to agent request
	agReq := NewAgentRequest(dP, nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)

	tplFlds := []*config.FCTemplate{
		{Tag: "ValExp", Filters: []string{},
			Path: "ValExp", Type: utils.MetaValueExponent,
			Value:     config.NewRSRParsersMustCompile("~*req.Session-Id", utils.InfieldSep),
			Mandatory: true},
	}
	if _, err := agReq.ParseField(tplFlds[0]); err == nil ||
		err.Error() != `invalid arguments <[{"Rules":"~*req.Session-Id"}]> to *value_exponent` {
		t.Error(err)
	}

	tplFlds = []*config.FCTemplate{
		{Tag: "ValExp", Filters: []string{},
			Path: "ValExp", Type: utils.MetaValueExponent,
			Value:     config.NewRSRParsersMustCompile("15;~*req.Session-Id", utils.InfieldSep),
			Mandatory: true},
	}
	if _, err := agReq.ParseField(tplFlds[0]); err == nil ||
		err.Error() != `strconv.Atoi: parsing "simuhuawei;1449573472;00002": invalid syntax` {
		t.Error(err)
	}

	tplFlds = []*config.FCTemplate{
		{Tag: "ValExp", Filters: []string{},
			Path: "ValExp", Type: utils.MetaValueExponent,
			Value:     config.NewRSRParsersMustCompile("~*req.Session-Id;15", utils.InfieldSep),
			Mandatory: true},
	}
	if _, err := agReq.ParseField(tplFlds[0]); err == nil ||
		err.Error() != `invalid value <simuhuawei;1449573472;00002> to *value_exponent` {
		t.Error(err)
	}
	tplFlds = []*config.FCTemplate{
		{Tag: "ValExp", Filters: []string{},
			Path: "ValExp", Type: utils.MetaValueExponent,
			Value:     config.NewRSRParsersMustCompile("2;3", utils.InfieldSep),
			Mandatory: true},
	}
	expected := "2000"
	if out, err := agReq.ParseField(tplFlds[0]); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(out, expected) {
		t.Errorf("expecting: <%+v>, %T received: <%+v> %T", expected, expected, out, out)
	}
}

func TestAgReqOverwrite(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	agReq := NewAgentRequest(nil, nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)
	// populate request, emulating the way will be done in HTTPAgent
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.ToR, PathSlice: []string{utils.ToR}}, utils.NewLeafNode(utils.MetaVoice))
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.AccountField, PathSlice: []string{utils.AccountField}}, utils.NewLeafNode("1001"))
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.Destination, PathSlice: []string{utils.Destination}}, utils.NewLeafNode("1002"))
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.AnswerTime, PathSlice: []string{utils.AnswerTime}}, utils.NewLeafNode(
		time.Date(2013, 12, 30, 15, 0, 1, 0, time.UTC)))
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.RequestType, PathSlice: []string{utils.RequestType}}, utils.NewLeafNode(utils.MetaPrepaid))

	agReq.CGRReply = &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{}}

	tplFlds := []*config.FCTemplate{
		{Tag: "Account",
			Path: utils.MetaCgrep + utils.NestingSep + utils.AccountField, Type: utils.MetaComposed,
			Value: config.NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep)},
		{Tag: "Account",
			Path: utils.MetaCgrep + utils.NestingSep + utils.AccountField, Type: utils.MetaComposed,
			Value: config.NewRSRParsersMustCompile(":", utils.InfieldSep)},
		{Tag: "Account",
			Path: utils.MetaCgrep + utils.NestingSep + utils.AccountField, Type: utils.MetaComposed,
			Value: config.NewRSRParsersMustCompile("~*cgreq.Account", utils.InfieldSep)},
		{Tag: "Account",
			Path: utils.MetaCgrep + utils.NestingSep + utils.AccountField, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("OverwrittenAccount", utils.InfieldSep)},
		{Tag: "Account",
			Path: utils.MetaCgrep + utils.NestingSep + utils.AccountField, Type: utils.MetaComposed,
			Value: config.NewRSRParsersMustCompile("WithComposed", utils.InfieldSep)},
	}
	for _, v := range tplFlds {
		v.ComputePath()
	}
	if err := agReq.SetFields(tplFlds); err != nil {
		t.Error(err)
	}

	if rcv, err := agReq.CGRReply.FieldAsInterface([]string{utils.AccountField}); err != nil {
		t.Error(err)
	} else if sls, canCast := rcv.([]*utils.DataNode); !canCast {
		t.Errorf("Cannot cast to &utils.NMSlice %+v", rcv)
	} else if len(sls) != 1 {
		t.Errorf("expecting: %+v, \n received: %+v ", 1, len(sls))
	} else if (sls)[0].Value.Data != "OverwrittenAccountWithComposed" {
		t.Errorf("expecting: %+v, \n received: %+v ",
			"OverwrittenAccountWithComposed", (sls)[0].Value.Data)
	}
}

func TestAgReqGroupType(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	agReq := NewAgentRequest(nil, nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)
	// populate request, emulating the way will be done in HTTPAgent
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.ToR, PathSlice: []string{utils.ToR}}, utils.NewLeafNode(utils.MetaVoice))
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.AccountField, PathSlice: []string{utils.AccountField}}, utils.NewLeafNode("1001"))
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.Destination, PathSlice: []string{utils.Destination}}, utils.NewLeafNode("1002"))
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.AnswerTime, PathSlice: []string{utils.AnswerTime}}, utils.NewLeafNode(
		time.Date(2013, 12, 30, 15, 0, 1, 0, time.UTC)))
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.RequestType, PathSlice: []string{utils.RequestType}}, utils.NewLeafNode(utils.MetaPrepaid))

	agReq.CGRReply = &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{}}

	tplFlds := []*config.FCTemplate{
		{Tag: "Account",
			Path: utils.MetaCgrep + utils.NestingSep + utils.AccountField, Type: utils.MetaGroup,
			Value: config.NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep)},
		{Tag: "Account",
			Path: utils.MetaCgrep + utils.NestingSep + utils.AccountField, Type: utils.MetaGroup,
			Value: config.NewRSRParsersMustCompile("test", utils.InfieldSep)},
	}
	for _, v := range tplFlds {
		v.ComputePath()
	}
	if err := agReq.SetFields(tplFlds); err != nil {
		t.Error(err)
	}

	if rcv, err := agReq.CGRReply.FieldAsInterface([]string{utils.AccountField}); err != nil {
		t.Error(err)
	} else if sls, canCast := rcv.([]*utils.DataNode); !canCast {
		t.Errorf("Cannot cast to &utils.NMSlice %+v", rcv)
	} else if len(sls) != 2 {
		t.Errorf("expecting: %+v, \n received: %+v ", 1, len(sls))
	} else if (sls)[0].Value.Data != "cgrates.org" {
		t.Errorf("expecting: %+v, \n received: %+v ", "cgrates.org", (sls)[0].Value.Data)
	} else if (sls)[1].Value.Data != "test" {
		t.Errorf("expecting: %+v, \n received: %+v ", "test", (sls)[1].Value.Data)
	}
}

func TestAgReqSetFieldsInTmp(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	agReq := NewAgentRequest(nil, nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.AccountField, PathSlice: []string{utils.AccountField}}, utils.NewLeafNode("1001"))

	tplFlds := []*config.FCTemplate{
		{Tag: "Tenant",
			Path: utils.MetaTmp + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep)},
		{Tag: "Account",
			Path: utils.MetaTmp + utils.NestingSep + utils.AccountField, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*cgreq.Account", utils.InfieldSep)},
	}
	for _, v := range tplFlds {
		v.ComputePath()
	}
	eMp := &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{}}
	eMp.Set([]string{utils.Tenant}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "cgrates.org"}}})
	eMp.Set([]string{utils.AccountField}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "1001"}}})

	if err := agReq.SetFields(tplFlds); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(agReq.tmp, eMp) {
		t.Errorf("expecting: %+v,\n received: %+v", eMp, agReq.tmp)
	}
}

func TestAgReqSetFieldsIp2Hex(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	agReq := NewAgentRequest(nil, nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)
	agReq.CGRRequest.Set(&utils.FullPath{Path: "IP", PathSlice: []string{"IP"}}, utils.NewLeafNode("62.87.114.244"))

	tplFlds := []*config.FCTemplate{
		{Tag: "IP",
			Path: utils.MetaTmp + utils.NestingSep + "IP", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*cgreq.IP{*ip2hex}", utils.InfieldSep)},
	}
	for _, v := range tplFlds {
		v.ComputePath()
	}
	eMp := &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{}}
	eMp.Set([]string{"IP"}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "0x3e5772f4"}}})

	if err := agReq.SetFields(tplFlds); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(agReq.tmp, eMp) {
		t.Errorf("expecting: %+v,\n received: %+v", eMp, agReq.tmp)
	}
}

func TestAgReqSetFieldsString2Hex(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	agReq := NewAgentRequest(nil, nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)
	agReq.CGRRequest.Set(&utils.FullPath{Path: "CustomField", PathSlice: []string{"CustomField"}}, utils.NewLeafNode(string([]byte{0x94, 0x71, 0x02, 0x31, 0x01, 0x59})))

	tplFlds := []*config.FCTemplate{
		{Tag: "CustomField",
			Path: utils.MetaTmp + utils.NestingSep + "CustomField", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*cgreq.CustomField{*string2hex}", utils.InfieldSep)},
	}
	for _, v := range tplFlds {
		v.ComputePath()
	}
	eMp := &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{}}
	eMp.Set([]string{"CustomField"}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "0x947102310159"}}})

	if err := agReq.SetFields(tplFlds); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(agReq.tmp, eMp) {
		t.Errorf("expecting: %+v,\n received: %+v", eMp, agReq.tmp)
	}
}

func TestAgReqSetFieldsWithRemove(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	agReq := NewAgentRequest(nil, nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)
	// populate request, emulating the way will be done in HTTPAgent
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.CGRID, PathSlice: []string{utils.CGRID}},
		utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()))
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.ToR, PathSlice: []string{utils.ToR}}, utils.NewLeafNode(utils.MetaVoice))
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.AccountField, PathSlice: []string{utils.AccountField}}, utils.NewLeafNode("1001"))
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.Destination, PathSlice: []string{utils.Destination}}, utils.NewLeafNode("1002"))
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.AnswerTime, PathSlice: []string{utils.AnswerTime}}, utils.NewLeafNode(
		time.Date(2013, 12, 30, 15, 0, 1, 0, time.UTC)))
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.RequestType, PathSlice: []string{utils.RequestType}}, utils.NewLeafNode(utils.MetaPrepaid))
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.Usage, PathSlice: []string{utils.Usage}}, utils.NewLeafNode(3*time.Minute))

	agReq.CGRReply = &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{
		utils.CapAttributes: {Type: utils.NMMapType, Map: map[string]*utils.DataNode{
			"PaypalAccount": utils.NewLeafNode("cgrates@paypal.com"),
		}},
		utils.CapMaxUsage: utils.NewLeafNode(120 * time.Second),
		utils.Error:       utils.NewLeafNode(""),
	}}

	tplFlds := []*config.FCTemplate{
		{Tag: "Tenant",
			Path: utils.MetaRep + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep)},
		{Tag: "Account",
			Path: utils.MetaRep + utils.NestingSep + utils.AccountField, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*cgreq.Account", utils.InfieldSep)},
		{Tag: "Destination",
			Path: utils.MetaRep + utils.NestingSep + utils.Destination, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*cgreq.Destination", utils.InfieldSep)},

		{Tag: "RequestedUsageVoice",
			Path: utils.MetaRep + utils.NestingSep + "RequestedUsage", Type: utils.MetaVariable,
			Filters: []string{"*string:~*cgreq.ToR:*voice"},
			Value: config.NewRSRParsersMustCompile(
				"~*cgreq.Usage{*duration_seconds}", utils.InfieldSep)},
		{Tag: "RequestedUsageData",
			Path: utils.MetaRep + utils.NestingSep + "RequestedUsage", Type: utils.MetaVariable,
			Filters: []string{"*string:~*cgreq.ToR:*data"},
			Value: config.NewRSRParsersMustCompile(
				"~*cgreq.Usage{*duration_nanoseconds}", utils.InfieldSep)},
		{Tag: "RequestedUsageSMS",
			Path: utils.MetaRep + utils.NestingSep + "RequestedUsage", Type: utils.MetaVariable,
			Filters: []string{"*string:~*cgreq.ToR:*sms"},
			Value: config.NewRSRParsersMustCompile(
				"~*cgreq.Usage{*duration_nanoseconds}", utils.InfieldSep)},

		{Tag: "AttrPaypalAccount",
			Path: utils.MetaRep + utils.NestingSep + "PaypalAccount", Type: utils.MetaVariable,
			Filters: []string{"*empty:~*cgrep.Error:"},
			Value: config.NewRSRParsersMustCompile(
				"~*cgrep.Attributes.PaypalAccount", utils.InfieldSep)},
		{Tag: "MaxUsage",
			Path: utils.MetaRep + utils.NestingSep + "MaxUsage", Type: utils.MetaVariable,
			Filters: []string{"*empty:~*cgrep.Error:"},
			Value: config.NewRSRParsersMustCompile(
				"~*cgrep.MaxUsage{*duration_seconds}", utils.InfieldSep)},
		{Tag: "Error",
			Path: utils.MetaRep + utils.NestingSep + "Error", Type: utils.MetaVariable,
			Filters: []string{"*rsr:~*cgrep.Error:!^$"},
			Value: config.NewRSRParsersMustCompile(
				"~*cgrep.Error", utils.InfieldSep)},
	}
	for _, v := range tplFlds {
		v.ComputePath()
	}
	eMp := utils.NewOrderedNavigableMap()
	eMp.SetAsSlice(&utils.FullPath{Path: utils.Tenant, PathSlice: []string{utils.Tenant}}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "cgrates.org"}}})
	eMp.SetAsSlice(&utils.FullPath{Path: utils.AccountField, PathSlice: []string{utils.AccountField}}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "1001"}}})
	eMp.SetAsSlice(&utils.FullPath{Path: utils.Destination, PathSlice: []string{utils.Destination}}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "1002"}}})
	eMp.SetAsSlice(&utils.FullPath{Path: "RequestedUsage", PathSlice: []string{"RequestedUsage"}}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "180"}}})
	eMp.SetAsSlice(&utils.FullPath{Path: "PaypalAccount", PathSlice: []string{"PaypalAccount"}}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "cgrates@paypal.com"}}})
	eMp.SetAsSlice(&utils.FullPath{Path: "MaxUsage", PathSlice: []string{"MaxUsage"}}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "120"}}})

	if err := agReq.SetFields(tplFlds); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(agReq.Reply, eMp) {
		t.Errorf("expecting: %+v,\n received: %+v", eMp, agReq.Reply)
	}

	tplFldsRemove := []*config.FCTemplate{
		{Tag: "Tenant",
			Path: utils.MetaRep + utils.NestingSep + utils.Tenant, Type: utils.MetaRemove},
		{Tag: "Account",
			Path: utils.MetaRep + utils.NestingSep + utils.AccountField, Type: utils.MetaRemove},
	}
	for _, v := range tplFldsRemove {
		v.ComputePath()
	}
	eMpRemove := utils.NewOrderedNavigableMap()
	eMpRemove.SetAsSlice(&utils.FullPath{Path: utils.Destination, PathSlice: []string{utils.Destination}}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "1002"}}})
	eMpRemove.SetAsSlice(&utils.FullPath{Path: "RequestedUsage", PathSlice: []string{"RequestedUsage"}}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "180"}}})
	eMpRemove.SetAsSlice(&utils.FullPath{Path: "PaypalAccount", PathSlice: []string{"PaypalAccount"}}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "cgrates@paypal.com"}}})
	eMpRemove.SetAsSlice(&utils.FullPath{Path: "MaxUsage", PathSlice: []string{"MaxUsage"}}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "120"}}})

	if err := agReq.SetFields(tplFldsRemove); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(agReq.Reply, eMpRemove) { // when compare ignore orders
		t.Errorf("expecting: %+v,\n received: %+v", eMpRemove, agReq.Reply)
	}

	tplFldsRemove = []*config.FCTemplate{
		{Path: utils.MetaRep, Type: utils.MetaRemoveAll},
	}
	tplFldsRemove[0].ComputePath()
	eMpRemove = utils.NewOrderedNavigableMap()
	if err := agReq.SetFields(tplFldsRemove); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(agReq.Reply.GetOrder(), eMpRemove.GetOrder()) {
		t.Errorf("expecting: %+v,\n received: %+v", eMpRemove, agReq.Reply)
	}
}

func TestAgReqSetFieldsInCache(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	engine.NewCacheS(cfg, dm, nil)
	agReq := NewAgentRequest(nil, nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.AccountField, PathSlice: []string{utils.AccountField}}, utils.NewLeafNode("1001"))

	tplFlds := []*config.FCTemplate{
		{Tag: "Tenant",
			Path: utils.MetaUCH + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep)},
		{Tag: "Account",
			Path: utils.MetaUCH + utils.NestingSep + utils.AccountField, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*cgreq.Account", utils.InfieldSep)},
	}
	for _, v := range tplFlds {
		v.ComputePath()
	}
	if err := agReq.SetFields(tplFlds); err != nil {
		t.Error(err)
	}

	if val, err := agReq.FieldAsInterface([]string{utils.MetaUCH, utils.Tenant}); err != nil {
		t.Error(err)
	} else if val != "cgrates.org" {
		t.Errorf("expecting: %+v, \n received: %+v ", "cgrates.org", utils.ToJSON(val))
	}

	if val, err := agReq.FieldAsInterface([]string{utils.MetaUCH, utils.AccountField}); err != nil {
		t.Error(err)
	} else if val != "1001" {
		t.Errorf("expecting: %+v, \n received: %+v ", "1001", utils.ToJSON(val))
	}

	if _, err := agReq.FieldAsInterface([]string{utils.MetaUCH, "Unexist"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
}

func TestAgReqSetFieldsInCacheWithTimeOut(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)

	cfg.CacheCfg().Partitions[utils.CacheUCH].TTL = time.Second
	engine.Cache = engine.NewCacheS(cfg, dm, nil)
	agReq := NewAgentRequest(nil, nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.AccountField, PathSlice: []string{utils.AccountField}}, utils.NewLeafNode("1001"))

	tplFlds := []*config.FCTemplate{
		{Tag: "Tenant",
			Path: utils.MetaUCH + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep)},
		{Tag: "Account",
			Path: utils.MetaUCH + utils.NestingSep + utils.AccountField, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*cgreq.Account", utils.InfieldSep)},
	}
	for _, v := range tplFlds {
		v.ComputePath()
	}
	if err := agReq.SetFields(tplFlds); err != nil {
		t.Error(err)
	}

	if val, err := agReq.FieldAsInterface([]string{utils.MetaUCH, utils.Tenant}); err != nil {
		t.Error(err)
	} else if val != "cgrates.org" {
		t.Errorf("expecting: %+v, \n received: %+v ", "cgrates.org", utils.ToJSON(val))
	}

	if val, err := agReq.FieldAsInterface([]string{utils.MetaUCH, utils.AccountField}); err != nil {
		t.Error(err)
	} else if val != "1001" {
		t.Errorf("expecting: %+v, \n received: %+v ", "1001", utils.ToJSON(val))
	}

	if _, err := agReq.FieldAsInterface([]string{utils.MetaUCH, "Unexist"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	// give enough time to Cache to remove ttl the *uch
	time.Sleep(2 * time.Second)
	if _, err := agReq.FieldAsInterface([]string{utils.MetaUCH, utils.Tenant}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	if _, err := agReq.FieldAsInterface([]string{utils.MetaUCH, utils.AccountField}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
}

func TestAgReqFiltersInsideField(t *testing.T) {
	//simulate the diameter request
	m := diam.NewRequest(diam.CreditControl, 4, nil)
	m.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String("bb97be2b9f37c2be9614fff71c8b1d08b1acbff8"))
	m.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("192.168.1.1"))
	m.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	m.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	m.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(3))
	m.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(2))
	m.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	m.NewAVP(avp.DestinationRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	m.NewAVP(avp.ServiceContextID, avp.Mbit, 0, datatype.UTF8String("voice@DiamItCCRInit"))
	m.NewAVP(avp.EventTimestamp, avp.Mbit, 0, datatype.Time(time.Date(2018, 10, 4, 15, 12, 20, 0, time.UTC)))
	m.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(450, avp.Mbit, 0, datatype.Enumerated(0)),      // Subscription-Id-Type
			diam.NewAVP(444, avp.Mbit, 0, datatype.UTF8String("1006")), // Subscription-Id-Data
		}})
	m.NewAVP(avp.ServiceIdentifier, avp.Mbit, 0, datatype.Unsigned32(0))
	m.NewAVP(avp.RequestedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(420, avp.Mbit, 0, datatype.Unsigned32(0))}})
	m.NewAVP(avp.UsedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(420, avp.Mbit, 0, datatype.Unsigned32(250))}})
	m.NewAVP(873, avp.Mbit, 10415, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(20300, avp.Mbit, 2011, &diam.GroupedAVP{ // IN-Information
				AVP: []*diam.AVP{
					diam.NewAVP(831, avp.Mbit, 10415, datatype.UTF8String("1006")),                                      // Calling-Party-Address
					diam.NewAVP(832, avp.Mbit, 10415, datatype.UTF8String("1002")),                                      // Called-Party-Address
					diam.NewAVP(20327, avp.Mbit, 2011, datatype.UTF8String("1002")),                                     // Real-Called-Number
					diam.NewAVP(20339, avp.Mbit, 2011, datatype.Unsigned32(0)),                                          // Charge-Flow-Type
					diam.NewAVP(20302, avp.Mbit, 2011, datatype.UTF8String("")),                                         // Calling-Vlr-Number
					diam.NewAVP(20303, avp.Mbit, 2011, datatype.UTF8String("")),                                         // Calling-CellID-Or-SAI
					diam.NewAVP(20313, avp.Mbit, 2011, datatype.OctetString("")),                                        // Bearer-Capability
					diam.NewAVP(20321, avp.Mbit, 2011, datatype.UTF8String("bb97be2b9f37c2be9614fff71c8b1d08b1acbff8")), // Call-Reference-Number
					diam.NewAVP(20322, avp.Mbit, 2011, datatype.UTF8String("")),                                         // MSC-Address
					diam.NewAVP(20324, avp.Mbit, 2011, datatype.Unsigned32(0)),                                          // Time-Zone
					diam.NewAVP(20385, avp.Mbit, 2011, datatype.UTF8String("")),                                         // Called-Party-NP
					diam.NewAVP(20386, avp.Mbit, 2011, datatype.UTF8String("")),                                         // SSP-Time
				},
			}),
		}})
	//create diameterDataProvider
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	//pass the data provider to agent request
	agReq := NewAgentRequest(newDADataProvider(nil, m), nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)

	tplFlds := []*config.FCTemplate{
		{Tag: "Usage",
			Path: utils.MetaCgreq + utils.NestingSep + utils.Usage, Type: utils.MetaCCUsage,
			Value: config.NewRSRParsersMustCompile("~*req.CC-Request-Number;~*req.Used-Service-Unit.CC-Time:s/(.*)/${1}s/;5m",
				utils.InfieldSep)},
		{Tag: "AnswerTime",
			Path: utils.MetaCgreq + utils.NestingSep + utils.AnswerTime, Type: utils.MetaDifference,
			Filters: []string{"*gt:~*cgreq.Usage:0s"}, // populate answer time if usage is greater than zero
			Value:   config.NewRSRParsersMustCompile("~*req.Event-Timestamp;~*cgreq.Usage", utils.InfieldSep)},
	}
	for _, v := range tplFlds {
		v.ComputePath()
	}
	if err := agReq.SetFields(tplFlds); err != nil {
		t.Error(err)
	}
	if val, err := agReq.FieldAsInterface([]string{utils.MetaCgreq, utils.AnswerTime}); err != nil {
		t.Error(err)
	} else if !val.(time.Time).Equal(time.Date(2018, 10, 4, 15, 3, 10, 0, time.UTC)) {
		t.Errorf("expecting: %+v, \n received: %+v ", time.Date(2018, 10, 4, 15, 3, 10, 0, time.UTC), val)
	}
}

func TestAgReqDynamicPath(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	agReq := NewAgentRequest(nil, nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)
	// populate request, emulating the way will be done in HTTPAgent
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.ToR, PathSlice: []string{utils.ToR}}, utils.NewLeafNode(utils.MetaVoice))
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.AccountField, PathSlice: []string{utils.AccountField}}, utils.NewLeafNode("1001"))
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.Destination, PathSlice: []string{utils.Destination}}, utils.NewLeafNode("1002"))
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.AnswerTime, PathSlice: []string{utils.AnswerTime}}, utils.NewLeafNode(
		time.Date(2013, 12, 30, 15, 0, 1, 0, time.UTC)))
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.RequestType, PathSlice: []string{utils.RequestType}}, utils.NewLeafNode(utils.MetaPrepaid))
	agReq.CGRRequest.Set(&utils.FullPath{Path: "Routes.CGR_ROUTE1", PathSlice: []string{"Routes", "CGR_ROUTE1"}}, utils.NewLeafNode(1001))
	agReq.CGRRequest.Set(&utils.FullPath{Path: "Routes.CGR_ROUTE2", PathSlice: []string{"Routes", "CGR_ROUTE2"}}, utils.NewLeafNode(1002))
	agReq.CGRRequest.Set(&utils.FullPath{Path: "BestRoute", PathSlice: []string{"BestRoute"}}, utils.NewLeafNode("ROUTE1"))

	agReq.CGRReply = &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{}}
	val1, err := config.NewRSRParsersFromSlice([]string{"~*cgreq.Routes.<CGR_;~*cgreq.BestRoute>"})
	if err != nil {
		t.Error(err)
	}
	tplFlds := []*config.FCTemplate{
		{Tag: "Tenant",
			Path: utils.MetaCgrep + utils.NestingSep + utils.Tenant, Type: utils.MetaComposed,
			Value: config.NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep)},
		{Tag: "Account",
			Path: utils.MetaCgrep + utils.NestingSep + utils.AccountField, Type: utils.MetaComposed,
			Value: config.NewRSRParsersMustCompile("~*cgreq.Account", utils.InfieldSep)},
		{Tag: "Destination",
			Path: utils.MetaCgrep + utils.NestingSep + utils.Destination, Type: utils.MetaComposed,
			Value: config.NewRSRParsersMustCompile("~*cgreq.Destination", utils.InfieldSep)},
		{Tag: "Usage",
			Path: utils.MetaCgrep + utils.NestingSep + utils.Usage, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("30s", utils.InfieldSep)},
		{Tag: "Route",
			Path: utils.MetaCgrep + utils.NestingSep + "Route",
			Type: utils.MetaVariable, Value: val1,
		},
		{Tag: "Route2",
			Path: utils.MetaCgrep + utils.NestingSep + "Route2.<CGR_;~*cgreq.BestRoute>",
			Type: utils.MetaVariable, Value: config.NewRSRParsersMustCompile("~*cgreq.Routes[CGR_ROUTE2]", utils.InfieldSep),
		},
	}
	for _, v := range tplFlds {
		v.ComputePath()
	}
	eMp := &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{}}
	eMp.Set([]string{utils.Tenant}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "cgrates.org"}}})
	eMp.Set([]string{utils.AccountField}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "1001"}}})
	eMp.Set([]string{utils.Destination}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "1002"}}})
	eMp.Set([]string{"Usage"}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "30s"}}})
	eMp.Set([]string{"Route"}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "1001"}}})
	eMp.Set([]string{"Route2", "CGR_ROUTE1"}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "1002"}}})

	if err := agReq.SetFields(tplFlds); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(agReq.CGRReply, eMp) {
		t.Errorf("expecting: %+v,\n received: %+v", utils.ToJSON(eMp), utils.ToJSON(agReq.CGRReply))
	}
}

func TestAgReqRoundingDecimals(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	agReq := NewAgentRequest(nil, nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)
	// populate request, emulating the way will be done in HTTPAgent
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.ToR, PathSlice: []string{utils.ToR}}, utils.NewLeafNode(utils.MetaVoice))
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.AccountField, PathSlice: []string{utils.AccountField}}, utils.NewLeafNode("1001"))
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.Destination, PathSlice: []string{utils.Destination}}, utils.NewLeafNode("1002"))
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.AnswerTime, PathSlice: []string{utils.AnswerTime}}, utils.NewLeafNode(
		time.Date(2013, 12, 30, 15, 0, 1, 0, time.UTC)))
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.RequestType, PathSlice: []string{utils.RequestType}}, utils.NewLeafNode(utils.MetaPrepaid))
	agReq.CGRRequest.Set(&utils.FullPath{Path: utils.Cost, PathSlice: []string{utils.Cost}}, utils.NewLeafNode(12.12645))

	agReq.CGRReply = &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{}}

	tplFlds := []*config.FCTemplate{
		{Tag: "Cost",
			Path: utils.MetaCgrep + utils.NestingSep + utils.Cost, Type: utils.MetaComposed,
			Value: config.NewRSRParsersMustCompile("~*cgreq.Cost{*round:3}", utils.InfieldSep)},
	}
	for _, v := range tplFlds {
		v.ComputePath()
	}
	if err := agReq.SetFields(tplFlds); err != nil {
		t.Error(err)
	}

	if rcv, err := agReq.CGRReply.FieldAsInterface([]string{utils.Cost}); err != nil {
		t.Error(err)
	} else if sls, canCast := rcv.([]*utils.DataNode); !canCast {
		t.Errorf("Cannot cast to &utils.NMSlice %+v", rcv)
	} else if len(sls) != 1 {
		t.Errorf("expecting: %+v, \n received: %+v ", 1, len(sls))
	} else if (sls)[0].Value.Data != "12.126" {
		t.Errorf("expecting: %+v, \n received: %+v",
			"12.126", (sls)[0].Value.Data)
	}
}

/*
$go test -bench=.  -run=^$ -benchtime=10s -count=3
goos: linux
goarch: amd64
pkg: github.com/cgrates/cgrates/agents
BenchmarkAgReqSetField-16    	  990145	     12012 ns/op
BenchmarkAgReqSetField-16    	 1000000	     12478 ns/op
BenchmarkAgReqSetField-16    	  904732	     13250 ns/op
PASS
ok  	github.com/cgrates/cgrates/agents	36.788s
*/
func BenchmarkAgReqSetField(b *testing.B) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	tplFlds := []*config.FCTemplate{
		{Tag: "Tenant",
			Path: utils.MetaCgrep + utils.NestingSep + utils.Tenant, Type: utils.MetaComposed,
			Value: config.NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep)},
		{Tag: "Account",
			Path: utils.MetaCgrep + utils.NestingSep + utils.AccountField + "[0].ID", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*cgreq.Account", utils.InfieldSep)},
		{Tag: "Account2",
			Path: utils.MetaCgrep + utils.NestingSep + utils.AccountField + "[1].ID", Type: utils.MetaConstant,
			Value: config.NewRSRParsersMustCompile("1003", utils.InfieldSep)},
	}
	for _, v := range tplFlds {
		v.ComputePath()
	}
	eMp := &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{}}
	eMp.Set([]string{utils.Tenant}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "cgrates.org"}}})
	eMp.Set([]string{utils.AccountField, "0", "ID"}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "1001"}}})
	eMp.Set([]string{utils.AccountField, "1", "ID"}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "1003"}}})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		agReq := NewAgentRequest(nil, nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)
		// populate request, emulating the way will be done in HTTPAgent
		agReq.CGRRequest.Set(&utils.FullPath{Path: utils.ToR, PathSlice: []string{utils.ToR}}, utils.NewLeafNode(utils.MetaVoice))
		agReq.CGRRequest.Set(&utils.FullPath{Path: utils.AccountField, PathSlice: []string{utils.AccountField}}, utils.NewLeafNode("1001"))
		agReq.CGRRequest.Set(&utils.FullPath{Path: utils.Destination, PathSlice: []string{utils.Destination}}, utils.NewLeafNode("1002"))
		agReq.CGRRequest.Set(&utils.FullPath{Path: utils.AnswerTime, PathSlice: []string{utils.AnswerTime}}, utils.NewLeafNode(
			time.Date(2013, 12, 30, 15, 0, 1, 0, time.UTC)))
		agReq.CGRRequest.Set(&utils.FullPath{Path: utils.RequestType, PathSlice: []string{utils.RequestType}}, utils.NewLeafNode(utils.MetaPrepaid))
		agReq.CGRReply = &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{}}

		if err := agReq.SetFields(tplFlds); err != nil {
			b.Error(err)
		} else if !reflect.DeepEqual(agReq.CGRReply, eMp) {
			b.Errorf("expecting: %+v,\n received: %+v", eMp, agReq.CGRReply)
		}
	}
}

func TestNeedsMaxUsage(t *testing.T) {
	if needsMaxUsage(nil) {
		t.Error("Expected empty flag to not need maxUsage")
	}

	if needsMaxUsage(utils.FlagParams{}) {
		t.Error("Expected empty flag to not need maxUsage")
	}

	if needsMaxUsage(utils.FlagParams{utils.MetaIDs: {"ID1", "ID2"}}) {
		t.Error("Expected flag to not need maxUsage")
	}
	if !needsMaxUsage(utils.FlagParams{utils.MetaIDs: {"ID1", "ID2"}, utils.MetaInitiate: {}}) {
		t.Error("Expected flag to need maxUsage")
	}
}

func TestAgReqSetFieldsFromCfg(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)

	agReq := NewAgentRequest(nil, nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)

	tplFlds := []*config.FCTemplate{
		{Tag: "CfgField",
			Path: utils.MetaCgreq + utils.NestingSep + "NodeID", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*cfg.general.node_id", utils.InfieldSep)},
	}
	for _, v := range tplFlds {
		v.ComputePath()
	}
	if err := agReq.SetFields(tplFlds); err != nil {
		t.Error(err)
	}

	if val, err := agReq.FieldAsInterface([]string{utils.MetaCgreq, "NodeID"}); err != nil {
		t.Error(err)
	} else if val != config.CgrConfig().GeneralCfg().NodeID {
		t.Errorf("expecting: %+v, \n received: %+v ", config.CgrConfig().GeneralCfg().NodeID, utils.ToJSON(val))
	}

}

func TestFieldAsInterface(t *testing.T) {
	fldPath := []string{utils.MetaOpts, utils.AccountField}
	ar := &AgentRequest{
		Request:    nil,
		Vars:       &utils.DataNode{},
		CGRRequest: &utils.OrderedNavigableMap{},
		diamreq:    nil,
		tmp:        &utils.DataNode{},
		Opts: utils.MapStorage{
			utils.AccountField: "Field1",
		},
	}

	rcv, err := ar.FieldAsInterface(fldPath)
	rcvExpect := ar.Opts[utils.AccountField]
	if err != nil {
		t.Error(err)
	} else if rcv != rcvExpect {
		t.Errorf("Expected %v but received %v", rcvExpect, rcv)
	}

	//default case
	fldPath = []string{utils.MetaNone}
	_, err = ar.FieldAsInterface(fldPath)
	errExpect := "unsupported field prefix: <*none>"
	if err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
}
