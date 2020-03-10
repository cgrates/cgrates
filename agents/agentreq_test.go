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
	"github.com/fiorix/go-diameter/diam"
	"github.com/fiorix/go-diameter/diam/avp"
	"github.com/fiorix/go-diameter/diam/datatype"
)

func TestAgReqSetFields(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	agReq := NewAgentRequest(nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)
	// populate request, emulating the way will be done in HTTPAgent
	agReq.CGRRequest.Set([]*utils.PathItem{{Field: utils.CGRID}},
		utils.NewNMInterface(utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String())))
	agReq.CGRRequest.Set([]*utils.PathItem{{Field: utils.ToR}}, utils.NewNMInterface(utils.VOICE))
	agReq.CGRRequest.Set([]*utils.PathItem{{Field: utils.Account}}, utils.NewNMInterface("1001"))
	agReq.CGRRequest.Set([]*utils.PathItem{{Field: utils.Destination}}, utils.NewNMInterface("1002"))
	agReq.CGRRequest.Set([]*utils.PathItem{{Field: utils.AnswerTime}},
		utils.NewNMInterface(time.Date(2013, 12, 30, 15, 0, 1, 0, time.UTC)))
	agReq.CGRRequest.Set([]*utils.PathItem{{Field: utils.RequestType}}, utils.NewNMInterface(utils.META_PREPAID))
	agReq.CGRRequest.Set([]*utils.PathItem{{Field: utils.Usage}}, utils.NewNMInterface(time.Duration(3*time.Minute)))

	cgrRply := utils.NavigableMap2{
		utils.CapAttributes: utils.NavigableMap2{
			"PaypalAccount": utils.NewNMInterface("cgrates@paypal.com"),
		},
		utils.CapMaxUsage: utils.NewNMInterface(time.Duration(120 * time.Second)),
		utils.Error:       utils.NewNMInterface(""),
	}
	agReq.CGRReply = &cgrRply

	tplFlds := []*config.FCTemplate{
		&config.FCTemplate{Tag: "Tenant",
			Path: utils.PathItems{{Field: utils.MetaRep}, {Field: utils.Tenant}}, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Tag: "Account",
			Path: utils.PathItems{{Field: utils.MetaRep}, {Field: utils.Account}}, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*cgreq.Account", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Tag: "Destination",
			Path: utils.PathItems{{Field: utils.MetaRep}, {Field: utils.Destination}}, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*cgreq.Destination", true, utils.INFIELD_SEP)},

		&config.FCTemplate{Tag: "RequestedUsageVoice",
			Path: utils.PathItems{{Field: utils.MetaRep}, {Field: "RequestedUsage"}}, Type: utils.MetaVariable,
			Filters: []string{"*string:~*cgreq.ToR:*voice"},
			Value: config.NewRSRParsersMustCompile(
				"~*cgreq.Usage{*duration_seconds}", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Tag: "RequestedUsageData",
			Path: utils.PathItems{{Field: utils.MetaRep}, {Field: "RequestedUsage"}}, Type: utils.MetaVariable,
			Filters: []string{"*string:~*cgreq.ToR:*data"},
			Value: config.NewRSRParsersMustCompile(
				"~*cgreq.Usage{*duration_nanoseconds}", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Tag: "RequestedUsageSMS",
			Path: utils.PathItems{{Field: utils.MetaRep}, {Field: "RequestedUsage"}}, Type: utils.MetaVariable,
			Filters: []string{"*string:~*cgreq.ToR:*sms"},
			Value: config.NewRSRParsersMustCompile(
				"~*cgreq.Usage{*duration_nanoseconds}", true, utils.INFIELD_SEP)},

		&config.FCTemplate{Tag: "AttrPaypalAccount",
			Path: utils.PathItems{{Field: utils.MetaRep}, {Field: "PaypalAccount"}}, Type: utils.MetaVariable,
			Filters: []string{"*empty:~*cgrep.Error:"},
			Value: config.NewRSRParsersMustCompile(
				"~*cgrep.Attributes.PaypalAccount", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Tag: "MaxUsage",
			Path: utils.PathItems{{Field: utils.MetaRep}, {Field: "MaxUsage"}}, Type: utils.MetaVariable,
			Filters: []string{"*empty:~*cgrep.Error:"},
			Value: config.NewRSRParsersMustCompile(
				"~*cgrep.MaxUsage{*duration_seconds}", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Tag: "Error",
			Path: utils.PathItems{{Field: utils.MetaRep}, {Field: "Error"}}, Type: utils.MetaVariable,
			Filters: []string{"*notempty:~*cgrep.Error:"},
			Value: config.NewRSRParsersMustCompile(
				"~*cgrep.Error", true, utils.INFIELD_SEP)},
	}
	eMp := utils.NewOrderedNavigableMap()
	eMp.Set([]*utils.PathItem{{Field: utils.Tenant}}, &utils.NMSlice{
		&config.NMItem{Data: "cgrates.org", Path: []string{utils.Tenant},
			Config: tplFlds[0]}})
	eMp.Set([]*utils.PathItem{{Field: utils.Account}}, &utils.NMSlice{
		&config.NMItem{Data: "1001", Path: []string{utils.Account},
			Config: tplFlds[1]}})
	eMp.Set([]*utils.PathItem{{Field: utils.Destination}}, &utils.NMSlice{
		&config.NMItem{Data: "1002", Path: []string{utils.Destination},
			Config: tplFlds[2]}})
	eMp.Set([]*utils.PathItem{{Field: "RequestedUsage"}}, &utils.NMSlice{
		&config.NMItem{Data: "180", Path: []string{"RequestedUsage"},
			Config: tplFlds[3]}})
	eMp.Set([]*utils.PathItem{{Field: "PaypalAccount"}}, &utils.NMSlice{
		&config.NMItem{Data: "cgrates@paypal.com", Path: []string{"PaypalAccount"},
			Config: tplFlds[6]}})
	eMp.Set([]*utils.PathItem{{Field: "MaxUsage"}}, &utils.NMSlice{
		&config.NMItem{Data: "120", Path: []string{"MaxUsage"},
			Config: tplFlds[7]}})

	if err := agReq.SetFields(tplFlds); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(agReq.Reply, eMp) {
		t.Errorf("expecting: %+v,\n received: %+v", eMp, agReq.Reply)
	}
}

func TestAgentRequestSetFields(t *testing.T) {
	req := map[string]interface{}{
		utils.Account: 1009,
		utils.Tenant:  "cgrates.org",
	}
	cfg, _ := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items),
		config.CgrConfig().CacheCfg(), nil)
	ar := NewAgentRequest(utils.NavigableMap(req), nil,
		nil, nil, config.NewRSRParsersMustCompile("", false, utils.NestingSep),
		"cgrates.org", "", engine.NewFilterS(cfg, nil, dm),
		utils.NavigableMap(req), utils.NavigableMap(req))
	input := []*config.FCTemplate{}
	if err := ar.SetFields(input); err != nil {
		t.Error(err)
	}
	// tplFld.Type == utils.META_NONE
	input = []*config.FCTemplate{&config.FCTemplate{Type: utils.META_NONE}}
	if err := ar.SetFields(input); err != nil {
		t.Error(err)
	}
	// unsupported type: <>
	input = []*config.FCTemplate{&config.FCTemplate{Blocker: true}}
	if err := ar.SetFields(input); err == nil || err.Error() != "unsupported type: <>" {
		t.Error(err)
	}
	// case utils.MetaVars
	input = []*config.FCTemplate{
		&config.FCTemplate{
			Path:  utils.PathItems{{Field: utils.MetaVars}, {Field: "Account"}},
			Tag:   fmt.Sprintf("%s.Account", utils.MetaVars),
			Type:  utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~"+utils.MetaReq+".Account", false, ";"),
		},
	}
	if err := ar.SetFields(input); err != nil {
		t.Error(err)
	} else if val, err := ar.Vars.FieldAsInterface([]string{"Account"}); err != nil {
		t.Error(err)
	} else if nm, ok := val.(*utils.NMSlice); !ok {
		t.Error("Expecting NM items")
	} else if len(*nm) != 1 {
		t.Error("Expecting one item")
	} else if (*nm)[0].Interface() != "1009" {
		t.Error("Expecting 1009, received: ", (*nm)[0])
	}

	// case utils.MetaCgreq
	input = []*config.FCTemplate{
		&config.FCTemplate{
			Path:  utils.PathItems{{Field: utils.MetaCgreq}, {Field: "Account"}},
			Tag:   fmt.Sprintf("%s.Account", utils.MetaCgreq),
			Type:  utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~"+utils.MetaReq+".Account", false, ";"),
		},
	}
	if err := ar.SetFields(input); err != nil {
		t.Error(err)
	} else if val, err := ar.CGRRequest.FieldAsInterface([]string{"Account"}); err != nil {
		t.Error(err)
	} else if nm, ok := val.(*utils.NMSlice); !ok {
		t.Error("Expecting NM items")
	} else if len(*nm) != 1 {
		t.Error("Expecting one item")
	} else if (*nm)[0].Interface() != "1009" {
		t.Error("Expecting 1009, received: ", (*nm)[0])
	}

	// case utils.MetaCgrep
	input = []*config.FCTemplate{
		&config.FCTemplate{
			Path:  utils.PathItems{{Field: utils.MetaCgrep}, {Field: "Account"}},
			Tag:   fmt.Sprintf("%s.Account", utils.MetaCgrep),
			Type:  utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~"+utils.MetaReq+".Account", false, ";"),
		},
	}
	if err := ar.SetFields(input); err != nil {
		t.Error(err)
	} else if val, err := ar.CGRReply.FieldAsInterface([]string{"Account"}); err != nil {
		t.Error(err)
	} else if nm, ok := val.(*utils.NMSlice); !ok {
		t.Error("Expecting NM items")
	} else if len(*nm) != 1 {
		t.Error("Expecting one item")
	} else if (*nm)[0].Interface() != "1009" {
		t.Error("Expecting 1009, received: ", (*nm)[0])
	}

	// case utils.MetaRep
	input = []*config.FCTemplate{
		&config.FCTemplate{
			Path:  utils.PathItems{{Field: utils.MetaRep}, {Field: "Account"}},
			Tag:   fmt.Sprintf("%s.Account", utils.MetaRep),
			Type:  utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~"+utils.MetaReq+".Account", false, ";"),
		},
	}
	if err := ar.SetFields(input); err != nil {
		t.Error(err)
	} else if val, err := ar.Reply.FieldAsInterface([]string{"Account"}); err != nil {
		t.Error(err)
	} else if nm, ok := val.(*utils.NMSlice); !ok {
		t.Error("Expecting NM items")
	} else if len(*nm) != 1 {
		t.Error("Expecting one item")
	} else if (*nm)[0].Interface() != "1009" {
		t.Error("Expecting 1009, received: ", (*nm)[0])
	}

	// case utils.MetaDiamreq
	input = []*config.FCTemplate{
		&config.FCTemplate{
			Path:  utils.PathItems{{Field: utils.MetaDiamreq}, {Field: "Account"}},
			Tag:   fmt.Sprintf("%s.Account", utils.MetaDiamreq),
			Type:  utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~"+utils.MetaReq+".Account", false, ";"),
		},
	}
	if err := ar.SetFields(input); err != nil {
		t.Error(err)
	} else if val, err := ar.diamreq.FieldAsInterface([]string{"Account"}); err != nil {
		t.Error(err)
	} else if nm, ok := val.(*utils.NMSlice); !ok {
		t.Error("Expecting NM items")
	} else if len(*nm) != 1 {
		t.Error("Expecting one item")
	} else if (*nm)[0].Interface() != "1009" {
		t.Error("Expecting 1009, received: ", (*nm)[0])
	}

	//META_COMPOSED
	input = []*config.FCTemplate{
		&config.FCTemplate{
			Path:  utils.PathItems{{Field: utils.MetaVars}, {Field: "AccountID"}},
			Tag:   fmt.Sprintf("%s.AccountID", utils.MetaVars),
			Type:  utils.META_COMPOSED,
			Value: config.NewRSRParsersMustCompile("~"+utils.MetaReq+".Tenant", false, ";"),
		},
		&config.FCTemplate{
			Path:  utils.PathItems{{Field: utils.MetaVars}, {Field: "AccountID"}},
			Tag:   fmt.Sprintf("%s.AccountID", utils.MetaVars),
			Type:  utils.META_COMPOSED,
			Value: config.NewRSRParsersMustCompile(":", false, ";"),
		},
		&config.FCTemplate{
			Path:  utils.PathItems{{Field: utils.MetaVars}, {Field: "AccountID"}},
			Tag:   fmt.Sprintf("%s.AccountID", utils.MetaVars),
			Type:  utils.META_COMPOSED,
			Value: config.NewRSRParsersMustCompile("~"+utils.MetaReq+".Account", false, ";"),
		},
	}
	if err := ar.SetFields(input); err != nil {
		t.Error(err)
	} else if val, err := ar.Vars.FieldAsInterface([]string{"AccountID"}); err != nil {
		t.Error(err)
	} else if nm, ok := val.(*utils.NMSlice); !ok {
		t.Error("Expecting NM items")
	} else if len(*nm) != 1 {
		t.Error("Expecting one item")
	} else if (*nm)[0].Interface() != "cgrates.org:1009" {
		t.Error("Expecting cgrates.org:1009, received: ", (*nm)[0])
	}

	// META_CONSTANT
	input = []*config.FCTemplate{
		&config.FCTemplate{
			Path:  utils.PathItems{{Field: utils.MetaVars}, {Field: "Account"}},
			Tag:   fmt.Sprintf("%s.Account", utils.MetaVars),
			Type:  utils.META_CONSTANT,
			Value: config.NewRSRParsersMustCompile("2020", false, ";"),
		},
	}
	if err := ar.SetFields(input); err != nil {
		t.Error(err)
	} else if val, err := ar.Vars.FieldAsInterface([]string{"Account"}); err != nil {
		t.Error(err)
	} else if nm, ok := val.(*utils.NMSlice); !ok {
		t.Error("Expecting NM items")
	} else if len(*nm) != 1 {
		t.Error("Expecting one item")
	} else if nmi, canCast := (*nm)[0].(*config.NMItem); !canCast {
		t.Error("Expecting NM item")
	} else if nmi.Data != "2020" {
		t.Error("Expecting 1009, received: ", nmi.Data)
	}

	// Filters
	input = []*config.FCTemplate{
		&config.FCTemplate{
			Path:    utils.PathItems{{Field: utils.MetaVars}, {Field: "AccountID"}},
			Tag:     fmt.Sprintf("%s.AccountID", utils.MetaVars),
			Filters: []string{utils.MetaString + ":~" + utils.MetaVars + ".Account:1003"},
			Type:    utils.META_CONSTANT,
			Value:   config.NewRSRParsersMustCompile("2021", false, ";"),
		},
	}
	if err := ar.SetFields(input); err != nil {
		t.Error(err)
	} else if val, err := ar.Vars.FieldAsInterface([]string{"AccountID"}); err != nil {
		t.Error(err)
	} else if nm, ok := val.(*utils.NMSlice); !ok {
		t.Error("Expecting NM items")
	} else if len(*nm) != 1 {
		t.Error("Expecting one item")
	} else if (*nm)[0].Interface() != "cgrates.org:1009" {
		t.Error("Expecting cgrates.org:1009, received: ", (*nm)[0])
	}

	input = []*config.FCTemplate{
		&config.FCTemplate{
			Path:    utils.PathItems{{Field: utils.MetaVars}, {Field: "Account"}},
			Tag:     fmt.Sprintf("%s.Account", utils.MetaVars),
			Filters: []string{"Not really a filter"},
			Type:    utils.META_CONSTANT,
			Value:   config.NewRSRParsersMustCompile("2021", false, ";"),
		},
	}
	if err := ar.SetFields(input); err == nil || err.Error() != "NOT_FOUND:Not really a filter" {
		t.Errorf("Expecting: 'NOT_FOUND:Not really a filter', received: %+v", err)
	}

	// Blocker: true
	input = []*config.FCTemplate{
		&config.FCTemplate{
			Path:    utils.PathItems{{Field: utils.MetaVars}, {Field: "Name"}},
			Tag:     fmt.Sprintf("%s.Name", utils.MetaVars),
			Type:    utils.MetaVariable,
			Value:   config.NewRSRParsersMustCompile("~"+utils.MetaReq+".Account", false, ";"),
			Blocker: true,
		},
		&config.FCTemplate{
			Path:  utils.PathItems{{Field: utils.MetaVars}, {Field: "Name"}},
			Tag:   fmt.Sprintf("%s.Name", utils.MetaVars),
			Type:  utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("1005", false, ";"),
		},
	}
	if err := ar.SetFields(input); err != nil {
		t.Error(err)
	} else if val, err := ar.Vars.FieldAsInterface([]string{"Name"}); err != nil {
		t.Error(err)
	} else if nm, ok := val.(*utils.NMSlice); !ok {
		t.Error("Expecting NM items")
	} else if len(*nm) != 1 {
		t.Error("Expecting one item")
	} else if (*nm)[0].Interface() != "1009" {
		t.Error("Expecting 1009, received: ", (*nm)[0])
	}

	// ErrNotFound
	input = []*config.FCTemplate{
		&config.FCTemplate{
			Path:  utils.PathItems{{Field: utils.MetaVars}, {Field: "Test"}},
			Tag:   fmt.Sprintf("%s.Test", utils.MetaVars),
			Type:  utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~"+utils.MetaReq+".Test", false, ";"),
		},
	}
	if err := ar.SetFields(input); err != nil {
		t.Error(err)
	} else if _, err := ar.Vars.FieldAsInterface([]string{"Test"}); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expecting: %+v, received: %+v", utils.ErrNotFound, err)
	}
	input = []*config.FCTemplate{
		&config.FCTemplate{
			Path:      utils.PathItems{{Field: utils.MetaVars}, {Field: "Test"}},
			Tag:       fmt.Sprintf("%s.Test", utils.MetaVars),
			Type:      utils.MetaVariable,
			Value:     config.NewRSRParsersMustCompile("~"+utils.MetaReq+".Test", false, ";"),
			Mandatory: true,
		},
	}
	if err := ar.SetFields(input); err == nil || err.Error() != "NOT_FOUND:"+utils.MetaVars+".Test" {
		t.Errorf("Expecting: %+v, received: %+v", "NOT_FOUND:"+utils.MetaVars+".Test", err)
	}

	//Not found
	input = []*config.FCTemplate{
		&config.FCTemplate{
			Path:      utils.PathItems{{Field: "wrong"}},
			Tag:       "wrong",
			Type:      utils.MetaVariable,
			Value:     config.NewRSRParsersMustCompile("~*req.Account", false, ";"),
			Mandatory: true,
		},
	}
	if err := ar.SetFields(input); err == nil || err.Error() != "unsupported field prefix: <wrong>" {
		t.Errorf("Expecting: %+v, received: %+v", "unsupported field prefix: <wrong>", err)
	}

	// MetaHdr/MetaTrl
	input = []*config.FCTemplate{
		&config.FCTemplate{
			Path:  utils.PathItems{{Field: utils.MetaVars}, {Field: "Account4"}},
			Tag:   fmt.Sprintf("%s.Account4", utils.MetaVars),
			Type:  utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~"+utils.MetaHdr+".Account", false, ";"),
		},
	}
	if err := ar.SetFields(input); err != nil {
		t.Error(err)
	} else if val, err := ar.Vars.FieldAsInterface([]string{"Account4"}); err != nil {
		t.Error(err)
	} else if nm, ok := val.(*utils.NMSlice); !ok {
		t.Error("Expecting NM items")
	} else if len(*nm) != 1 {
		t.Error("Expecting one item")
	} else if (*nm)[0].Interface() != "1009" {
		t.Error("Expecting 1009, received: ", (*nm)[0])
	}

	input = []*config.FCTemplate{
		&config.FCTemplate{
			Path:  utils.PathItems{{Field: utils.MetaVars}, {Field: "Account5"}},
			Tag:   fmt.Sprintf("%s.Account5", utils.MetaVars),
			Type:  utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~"+utils.MetaTrl+".Account", false, ";"),
		},
	}
	if err := ar.SetFields(input); err != nil {
		t.Error(err)
	} else if val, err := ar.Vars.FieldAsInterface([]string{"Account5"}); err != nil {
		t.Error(err)
	} else if nm, ok := val.(*utils.NMSlice); !ok {
		t.Error("Expecting NM items")
	} else if len(*nm) != 1 {
		t.Error("Expecting one item")
	} else if (*nm)[0].Interface() != "1009" {
		t.Error("Expecting 1009, received: ", (*nm)[0])
	}
}

func TestAgReqMaxCost(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	agReq := NewAgentRequest(nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)
	// populate request, emulating the way will be done in HTTPAgent
	agReq.CGRRequest.Set([]*utils.PathItem{{Field: utils.CapMaxUsage}}, utils.NewNMInterface("120s"))

	cgrRply := utils.NavigableMap2{
		utils.CapMaxUsage: utils.NewNMInterface(time.Duration(120 * time.Second)),
	}
	agReq.CGRReply = &cgrRply

	tplFlds := []*config.FCTemplate{
		&config.FCTemplate{Tag: "MaxUsage",
			Path: utils.PathItems{{Field: utils.MetaRep}, {Field: "MaxUsage"}}, Type: utils.MetaVariable,
			Filters: []string{"*rsr::~*cgrep.MaxUsage(>0s)"},
			Value: config.NewRSRParsersMustCompile(
				"~*cgrep.MaxUsage{*duration_seconds}", true, utils.INFIELD_SEP)},
	}
	eMp := utils.NewOrderedNavigableMap()

	eMp.Set([]*utils.PathItem{{Field: "MaxUsage"}}, &utils.NMSlice{
		&config.NMItem{Data: "120", Path: []string{"MaxUsage"},
			Config: tplFlds[0]}})

	if err := agReq.SetFields(tplFlds); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(agReq.Reply, eMp) {
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
	cfg, _ := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items),
		config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	//pass the data provider to agent request
	agReq := NewAgentRequest(dP, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)

	tplFlds := []*config.FCTemplate{
		&config.FCTemplate{Tag: "MandatoryFalse",
			Path: utils.PathItems{{Field: "MandatoryFalse"}}, Type: utils.META_COMPOSED,
			Value:     config.NewRSRParsersMustCompile("~*req.MandatoryFalse", true, utils.INFIELD_SEP),
			Mandatory: false},
		&config.FCTemplate{Tag: "MandatoryTrue",
			Path: utils.PathItems{{Field: "MandatoryTrue"}}, Type: utils.META_COMPOSED,
			Value:     config.NewRSRParsersMustCompile("~*req.MandatoryTrue", true, utils.INFIELD_SEP),
			Mandatory: true},
		&config.FCTemplate{Tag: "Session-Id", Filters: []string{},
			Path: utils.PathItems{{Field: "Session-Id"}}, Type: utils.META_COMPOSED,
			Value:     config.NewRSRParsersMustCompile("~*req.Session-Id", true, utils.INFIELD_SEP),
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
	cfg, _ := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	//pass the data provider to agent request
	agReq := NewAgentRequest(dP, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)
	tplFlds := []*config.FCTemplate{
		&config.FCTemplate{Tag: "MandatoryFalse",
			Path: utils.PathItems{{Field: "MandatoryFalse"}}, Type: utils.META_COMPOSED,
			Value:     config.NewRSRParsersMustCompile("~*req.MandatoryFalse", true, utils.INFIELD_SEP),
			Mandatory: false},
		&config.FCTemplate{Tag: "MandatoryTrue",
			Path: utils.PathItems{{Field: "MandatoryTrue"}}, Type: utils.META_COMPOSED,
			Value:     config.NewRSRParsersMustCompile("~*req.MandatoryTrue", true, utils.INFIELD_SEP),
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
	cfg, _ := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	//pass the data provider to agent request
	agReq := NewAgentRequest(dP, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)
	tplFlds := []*config.FCTemplate{
		&config.FCTemplate{Tag: "MandatoryFalse",
			Path: utils.PathItems{{Field: "MandatoryFalse"}}, Type: utils.META_COMPOSED,
			Value:     config.NewRSRParsersMustCompile("~*req.MandatoryFalse", true, utils.INFIELD_SEP),
			Mandatory: false},
		&config.FCTemplate{Tag: "MandatoryTrue",
			Path: utils.PathItems{{Field: "MandatoryTrue"}}, Type: utils.META_COMPOSED,
			Value:     config.NewRSRParsersMustCompile("~*req.MandatoryTrue", true, utils.INFIELD_SEP),
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
	cfg, _ := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items),
		config.CgrConfig().CacheCfg(), nil)

	filterS := engine.NewFilterS(cfg, nil, dm)
	//pass the data provider to agent request
	agReq := NewAgentRequest(dP, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)
	tplFlds := []*config.FCTemplate{
		&config.FCTemplate{Tag: "MandatoryFalse",
			Path: utils.PathItems{{Field: "MandatoryFalse"}}, Type: utils.META_COMPOSED,
			Value:     config.NewRSRParsersMustCompile("~*req.MandatoryFalse", true, utils.INFIELD_SEP),
			Mandatory: false},
		&config.FCTemplate{Tag: "MandatoryTrue",
			Path: utils.PathItems{{Field: "MandatoryTrue"}}, Type: utils.META_COMPOSED,
			Value:     config.NewRSRParsersMustCompile("~*req.MandatoryTrue", true, utils.INFIELD_SEP),
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
	cfg, _ := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	agReq := NewAgentRequest(nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)
	// populate request, emulating the way will be done in HTTPAgent
	agReq.CGRRequest.Set([]*utils.PathItem{{Field: utils.CGRID}},
		utils.NewNMInterface(utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String())))
	agReq.CGRRequest.Set([]*utils.PathItem{{Field: utils.Account}}, utils.NewNMInterface("1001"))
	agReq.CGRRequest.Set([]*utils.PathItem{{Field: utils.Destination}}, utils.NewNMInterface("1002"))

	tplFlds := []*config.FCTemplate{
		&config.FCTemplate{Tag: "Tenant", Filters: []string{},
			Path: utils.PathItems{{Field: utils.MetaCgrep}, {Field: utils.Tenant}}, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP)},

		&config.FCTemplate{Tag: "Account", Filters: []string{},
			Path: utils.PathItems{{Field: utils.MetaCgrep}, {Field: utils.Account}}, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*cgreq.Account", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Tag: "Destination", Filters: []string{},
			Path: utils.PathItems{{Field: utils.MetaCgrep}, {Field: utils.Destination}}, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*cgreq.Destination", true, utils.INFIELD_SEP)},
	}
	eMp := &utils.NavigableMap2{}
	eMp.Set([]*utils.PathItem{{Field: utils.Tenant}}, &utils.NMSlice{
		&config.NMItem{Data: "cgrates.org", Path: []string{utils.Tenant},
			Config: tplFlds[0]}})
	eMp.Set([]*utils.PathItem{{Field: utils.Account}}, &utils.NMSlice{
		&config.NMItem{Data: "1001", Path: []string{utils.Account},
			Config: tplFlds[1]}})
	eMp.Set([]*utils.PathItem{{Field: utils.Destination}}, &utils.NMSlice{
		&config.NMItem{Data: "1002", Path: []string{utils.Destination},
			Config: tplFlds[2]}})

	if err := agReq.SetFields(tplFlds); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(agReq.CGRReply, eMp) {
		t.Errorf("expecting: %+v,\n received: %+v", eMp, agReq.CGRReply)
	}
}

func TestAgReqMetaExponent(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items),
		config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	agReq := NewAgentRequest(nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)
	agReq.CGRRequest.Set([]*utils.PathItem{{Field: "Value"}}, utils.NewNMInterface("2"))
	agReq.CGRRequest.Set([]*utils.PathItem{{Field: "Exponent"}}, utils.NewNMInterface("2"))

	tplFlds := []*config.FCTemplate{
		&config.FCTemplate{Tag: "TestExpo", Filters: []string{},
			Path: utils.PathItems{{Field: utils.MetaCgrep}, {Field: "TestExpo"}}, Type: utils.MetaValueExponent,
			Value: config.NewRSRParsersMustCompile("~*cgreq.Value;~*cgreq.Exponent", true, utils.INFIELD_SEP)},
	}
	eMp := &utils.NavigableMap2{}
	eMp.Set([]*utils.PathItem{{Field: "TestExpo"}}, &utils.NMSlice{
		&config.NMItem{Data: "200", Path: []string{"TestExpo"},
			Config: tplFlds[0]}})

	if err := agReq.SetFields(tplFlds); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(agReq.CGRReply, eMp) {
		t.Errorf("expecting: %+v,\n received: %+v", eMp, agReq.CGRReply)
	}
}

func TestAgReqFieldAsNone(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	agReq := NewAgentRequest(nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)
	// populate request, emulating the way will be done in HTTPAgent
	agReq.CGRRequest.Set([]*utils.PathItem{{Field: utils.ToR}}, utils.NewNMInterface(utils.VOICE))
	agReq.CGRRequest.Set([]*utils.PathItem{{Field: utils.Account}}, utils.NewNMInterface("1001"))
	agReq.CGRRequest.Set([]*utils.PathItem{{Field: utils.Destination}}, utils.NewNMInterface("1002"))

	tplFlds := []*config.FCTemplate{
		&config.FCTemplate{Tag: "Tenant",
			Path: utils.PathItems{{Field: utils.MetaCgrep}, {Field: utils.Tenant}}, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Tag: "Account",
			Path: utils.PathItems{{Field: utils.MetaCgrep}, {Field: utils.Account}}, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*cgreq.Account", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Type: utils.META_NONE, Blocker: true},
		&config.FCTemplate{Tag: "Destination",
			Path: utils.PathItems{{Field: utils.MetaCgrep}, {Field: utils.Destination}}, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*cgreq.Destination", true, utils.INFIELD_SEP)},
	}
	eMp := &utils.NavigableMap2{}
	eMp.Set([]*utils.PathItem{{Field: utils.Tenant}}, &utils.NMSlice{
		&config.NMItem{Data: "cgrates.org", Path: []string{utils.Tenant},
			Config: tplFlds[0]}})
	eMp.Set([]*utils.PathItem{{Field: utils.Account}}, &utils.NMSlice{
		&config.NMItem{Data: "1001", Path: []string{utils.Account},
			Config: tplFlds[1]}})
	if err := agReq.SetFields(tplFlds); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(agReq.CGRReply, eMp) {
		t.Errorf("expecting: %+v,\n received: %+v", eMp, agReq.CGRReply)
	}
}

func TestAgReqFieldAsNone2(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items),
		config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	agReq := NewAgentRequest(nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)
	// populate request, emulating the way will be done in HTTPAgent
	agReq.CGRRequest.Set([]*utils.PathItem{{Field: utils.ToR}}, utils.NewNMInterface(utils.VOICE))
	agReq.CGRRequest.Set([]*utils.PathItem{{Field: utils.Account}}, utils.NewNMInterface("1001"))
	agReq.CGRRequest.Set([]*utils.PathItem{{Field: utils.Destination}}, utils.NewNMInterface("1002"))

	tplFlds := []*config.FCTemplate{
		&config.FCTemplate{Tag: "Tenant",
			Path: utils.PathItems{{Field: utils.MetaCgrep}, {Field: utils.Tenant}}, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Tag: "Account",
			Path: utils.PathItems{{Field: utils.MetaCgrep}, {Field: utils.Account}}, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*cgreq.Account", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Type: utils.META_NONE},
		&config.FCTemplate{Tag: "Destination",
			Path: utils.PathItems{{Field: utils.MetaCgrep}, {Field: utils.Destination}}, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*cgreq.Destination", true, utils.INFIELD_SEP)},
	}
	eMp := &utils.NavigableMap2{}
	eMp.Set([]*utils.PathItem{{Field: utils.Tenant}}, &utils.NMSlice{
		&config.NMItem{Data: "cgrates.org", Path: []string{utils.Tenant},
			Config: tplFlds[0]}})
	eMp.Set([]*utils.PathItem{{Field: utils.Account}}, &utils.NMSlice{
		&config.NMItem{Data: "1001", Path: []string{utils.Account},
			Config: tplFlds[1]}})
	eMp.Set([]*utils.PathItem{{Field: utils.Destination}}, &utils.NMSlice{
		&config.NMItem{Data: "1002", Path: []string{utils.Destination},
			Config: tplFlds[3]}})
	if err := agReq.SetFields(tplFlds); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(agReq.CGRReply, eMp) {
		t.Errorf("expecting: %+v,\n received: %+v", eMp, agReq.CGRReply)
	}
}

func TestAgReqSetField2(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	agReq := NewAgentRequest(nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)
	// populate request, emulating the way will be done in HTTPAgent
	agReq.CGRRequest.Set([]*utils.PathItem{{Field: utils.ToR}}, utils.NewNMInterface(utils.VOICE))
	agReq.CGRRequest.Set([]*utils.PathItem{{Field: utils.Account}}, utils.NewNMInterface("1001"))
	agReq.CGRRequest.Set([]*utils.PathItem{{Field: utils.Destination}}, utils.NewNMInterface("1002"))
	agReq.CGRRequest.Set([]*utils.PathItem{{Field: utils.AnswerTime}},
		utils.NewNMInterface(time.Date(2013, 12, 30, 15, 0, 1, 0, time.UTC)))
	agReq.CGRRequest.Set([]*utils.PathItem{{Field: utils.RequestType}}, utils.NewNMInterface(utils.META_PREPAID))

	agReq.CGRReply = &utils.NavigableMap2{}

	tplFlds := []*config.FCTemplate{
		&config.FCTemplate{Tag: "Tenant",
			Path: utils.PathItems{{Field: utils.MetaCgrep}, {Field: utils.Tenant}}, Type: utils.META_COMPOSED,
			Value: config.NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Tag: "Account",
			Path: utils.PathItems{{Field: utils.MetaCgrep}, {Field: utils.Account}}, Type: utils.META_COMPOSED,
			Value: config.NewRSRParsersMustCompile("~*cgreq.Account", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Tag: "Destination",
			Path: utils.PathItems{{Field: utils.MetaCgrep}, {Field: utils.Destination}}, Type: utils.META_COMPOSED,
			Value: config.NewRSRParsersMustCompile("~*cgreq.Destination", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Tag: "Usage",
			Path: utils.PathItems{{Field: utils.MetaCgrep}, {Field: utils.Usage}}, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("30s", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Tag: "CalculatedUsage",
			Path: utils.PathItems{{Field: utils.MetaCgrep}, {Field: "CalculatedUsage"}},
			Type: "*difference", Value: config.NewRSRParsersMustCompile("~*cgreq.AnswerTime;~*cgrep.Usage", true, utils.INFIELD_SEP),
		},
	}
	eMp := &utils.NavigableMap2{}
	eMp.Set([]*utils.PathItem{{Field: utils.Tenant}}, &utils.NMSlice{
		&config.NMItem{Data: "cgrates.org", Path: []string{utils.Tenant},
			Config: tplFlds[0]}})
	eMp.Set([]*utils.PathItem{{Field: utils.Account}}, &utils.NMSlice{
		&config.NMItem{Data: "1001", Path: []string{utils.Account},
			Config: tplFlds[1]}})
	eMp.Set([]*utils.PathItem{{Field: utils.Destination}}, &utils.NMSlice{
		&config.NMItem{Data: "1002", Path: []string{utils.Destination},
			Config: tplFlds[2]}})
	eMp.Set([]*utils.PathItem{{Field: "Usage"}}, &utils.NMSlice{
		&config.NMItem{Data: "30s", Path: []string{"Usage"},
			Config: tplFlds[3]}})
	eMp.Set([]*utils.PathItem{{Field: "CalculatedUsage"}}, &utils.NMSlice{
		&config.NMItem{Data: time.Date(2013, 12, 30, 14, 59, 31, 0, time.UTC), Path: []string{"CalculatedUsage"},
			Config: tplFlds[4]}})

	if err := agReq.SetFields(tplFlds); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(agReq.CGRReply, eMp) {
		t.Errorf("expecting: %+v,\n received: %+v", eMp, agReq.CGRReply)
	}
}

func TestAgReqFieldAsInterface(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items),
		config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	agReq := NewAgentRequest(nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)
	// populate request, emulating the way will be done in HTTPAgent
	agReq.CGRRequest = utils.NewOrderedNavigableMap()
	agReq.CGRRequest.Set([]*utils.PathItem{{Field: utils.Usage}}, &utils.NMSlice{&config.NMItem{Data: 3 * time.Minute}})
	agReq.CGRRequest.Set([]*utils.PathItem{{Field: utils.ToR}}, &utils.NMSlice{&config.NMItem{Data: utils.VOICE}})
	agReq.CGRRequest.Set([]*utils.PathItem{{Field: utils.Account}}, utils.NewNMInterface("1001"))
	agReq.CGRRequest.Set([]*utils.PathItem{{Field: utils.Destination}}, utils.NewNMInterface("1002"))

	path := []string{utils.MetaCgreq, utils.Usage}
	var expVal interface{}
	expVal = &utils.NMSlice{&config.NMItem{Data: 3 * time.Minute}}
	if rply, err := agReq.FieldAsInterface(path); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rply, expVal) {
		t.Errorf("Expected %v , received: %v", utils.ToJSON(expVal), utils.ToJSON(rply))
	}

	path = []string{utils.MetaCgreq, utils.ToR}
	expVal = &utils.NMSlice{&config.NMItem{Data: utils.VOICE}}
	if rply, err := agReq.FieldAsInterface(path); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rply, expVal) {
		t.Errorf("Expected %v , received: %v", utils.ToJSON(expVal), utils.ToJSON(rply))
	}

	path = []string{utils.MetaCgreq, utils.Account}
	expVal = "1001"
	if rply, err := agReq.FieldAsInterface(path); err != nil {
		t.Error(err)
	} else if nm, canCast := rply.(utils.NM); !canCast {
		t.Errorf("Cant cast %T to NM interface", rply)
	} else if !reflect.DeepEqual(nm.Interface(), expVal) {
		t.Errorf("Expected %v , received: %v", utils.ToJSON(expVal), utils.ToJSON(nm.Interface()))
	}

	path = []string{utils.MetaCgreq, utils.Destination}
	expVal = "1002"
	if rply, err := agReq.FieldAsInterface(path); err != nil {
		t.Error(err)
	} else if nm, canCast := rply.(utils.NM); !canCast {
		t.Errorf("Cant cast %T to NM interface", rply)
	} else if !reflect.DeepEqual(nm.Interface(), expVal) {
		t.Errorf("Expected %v , received: %v", utils.ToJSON(expVal), utils.ToJSON(nm.Interface()))
	}
}

func TestAgReqNewARWithCGRRplyAndRply(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)

	rply := utils.NewOrderedNavigableMap()
	rply.Set([]*utils.PathItem{{Field: "FirstLevel"}, {Field: "SecondLevel"}, {Field: "Fld1"}}, utils.NewNMInterface("Val1"))

	cgrRply := &utils.NavigableMap2{
		utils.CapAttributes: utils.NavigableMap2{
			"PaypalAccount": utils.NewNMInterface("cgrates@paypal.com"),
		},
		utils.CapMaxUsage: utils.NewNMInterface(time.Duration(120 * time.Second)),
		utils.Error:       utils.NewNMInterface(""),
	}

	agReq := NewAgentRequest(nil, nil, cgrRply, rply, nil, "cgrates.org", "", filterS, nil, nil)

	tplFlds := []*config.FCTemplate{
		&config.FCTemplate{Tag: "Fld1",
			Path: utils.PathItems{{Field: utils.MetaCgreq}, {Field: "Fld1"}}, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*rep.FirstLevel.SecondLevel.Fld1", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Tag: "Fld2",
			Path: utils.PathItems{{Field: utils.MetaCgreq}, {Field: "Fld2"}}, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*cgrep.Attributes.PaypalAccount", true, utils.INFIELD_SEP)},
	}

	eMp := utils.NewOrderedNavigableMap()
	eMp.Set([]*utils.PathItem{{Field: "Fld1"}}, &utils.NMSlice{
		&config.NMItem{Data: "Val1", Path: []string{"Fld1"},
			Config: tplFlds[0]}})
	eMp.Set([]*utils.PathItem{{Field: "Fld2"}}, &utils.NMSlice{
		&config.NMItem{Data: "cgrates@paypal.com", Path: []string{"Fld2"},
			Config: tplFlds[1]}})

	if err := agReq.SetFields(tplFlds); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(agReq.CGRRequest, eMp) {
		t.Errorf("expecting: %+v,\n received: %+v", eMp, agReq.CGRRequest)
	}
}

func TestAgReqSetCGRReplyWithError(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items),
		config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)

	rply := utils.NewOrderedNavigableMap()
	rply.Set([]*utils.PathItem{{Field: "FirstLevel"}, {Field: "SecondLevel"}, {Field: "Fld1"}}, utils.NewNMInterface("Val1"))
	agReq := NewAgentRequest(nil, nil, nil, rply, nil, "cgrates.org", "", filterS, nil, nil)

	agReq.setCGRReply(nil, utils.ErrNotFound)

	tplFlds := []*config.FCTemplate{
		&config.FCTemplate{Tag: "Fld1",
			Path: utils.PathItems{{Field: utils.MetaCgreq}, {Field: "Fld1"}}, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*rep.FirstLevel.SecondLevel.Fld1", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Tag: "Fld2",
			Path: utils.PathItems{{Field: utils.MetaCgreq}, {Field: "Fld2"}}, Type: utils.MetaVariable,
			Value:     config.NewRSRParsersMustCompile("~*cgrep.Attributes.PaypalAccount", true, utils.INFIELD_SEP),
			Mandatory: true},
	}

	if err := agReq.SetFields(tplFlds); err == nil ||
		err.Error() != "NOT_FOUND:Fld2" {
		t.Error(err)
	}
}

type myEv map[string]utils.NM

func (ev myEv) AsNavigableMap() utils.NavigableMap2 {
	return utils.NavigableMap2(ev)
}

func TestAgReqSetCGRReplyWithoutError(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)

	rply := utils.NewOrderedNavigableMap()
	rply.Set([]*utils.PathItem{{Field: "FirstLevel"}, {Field: "SecondLevel"}, {Field: "Fld1"}}, utils.NewNMInterface("Val1"))

	myEv := myEv{
		utils.CapAttributes: utils.NavigableMap2{
			"PaypalAccount": utils.NewNMInterface("cgrates@paypal.com"),
		},
		utils.CapMaxUsage: utils.NewNMInterface(time.Duration(120 * time.Second)),
		utils.Error:       utils.NewNMInterface(""),
	}

	agReq := NewAgentRequest(nil, nil, nil, rply,
		nil, "cgrates.org", "", filterS, nil, nil)

	agReq.setCGRReply(myEv, nil)

	tplFlds := []*config.FCTemplate{
		&config.FCTemplate{Tag: "Fld1",
			Path: utils.PathItems{{Field: utils.MetaCgreq}, {Field: "Fld1"}}, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*rep.FirstLevel.SecondLevel.Fld1", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Tag: "Fld2",
			Path: utils.PathItems{{Field: utils.MetaCgreq}, {Field: "Fld2"}}, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*cgrep.Attributes.PaypalAccount", true, utils.INFIELD_SEP)},
	}

	eMp := utils.NewOrderedNavigableMap()
	eMp.Set([]*utils.PathItem{{Field: "Fld1"}}, &utils.NMSlice{
		&config.NMItem{Data: "Val1", Path: []string{"Fld1"},
			Config: tplFlds[0]}})
	eMp.Set([]*utils.PathItem{{Field: "Fld2"}}, &utils.NMSlice{
		&config.NMItem{Data: "cgrates@paypal.com", Path: []string{"Fld2"},
			Config: tplFlds[1]}})

	if err := agReq.SetFields(tplFlds); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(agReq.CGRRequest, eMp) {
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
	cfg, _ := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items),
		config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	//pass the data provider to agent request
	agReq := NewAgentRequest(dP, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)

	tplFlds := []*config.FCTemplate{
		&config.FCTemplate{Tag: "CCUsage", Filters: []string{},
			Path: utils.PathItems{{Field: "CCUsage"}}, Type: utils.MetaCCUsage,
			Value:     config.NewRSRParsersMustCompile("~*req.Session-Id", true, utils.INFIELD_SEP),
			Mandatory: true},
	}
	if _, err := agReq.ParseField(tplFlds[0]); err == nil ||
		err.Error() != `invalid arguments <[{"Rules":"~*req.Session-Id","AllFiltersMatch":true}]> to *cc_usage` {
		t.Error(err)
	}

	tplFlds = []*config.FCTemplate{
		&config.FCTemplate{Tag: "CCUsage", Filters: []string{},
			Path: utils.PathItems{{Field: "CCUsage"}}, Type: utils.MetaCCUsage,
			Value:     config.NewRSRParsersMustCompile("~*req.Session-Id;12s;12s", true, utils.INFIELD_SEP),
			Mandatory: true},
	}
	if _, err := agReq.ParseField(tplFlds[0]); err == nil ||
		err.Error() != `invalid requestNumber <simuhuawei;1449573472;00002> to *cc_usage` {
		t.Error(err)
	}

	tplFlds = []*config.FCTemplate{
		&config.FCTemplate{Tag: "CCUsage", Filters: []string{},
			Path: utils.PathItems{{Field: "CCUsage"}}, Type: utils.MetaCCUsage,
			Value:     config.NewRSRParsersMustCompile("10;~*req.Session-Id;12s", true, utils.INFIELD_SEP),
			Mandatory: true},
	}
	if _, err := agReq.ParseField(tplFlds[0]); err == nil ||
		err.Error() != `invalid usedCCTime <simuhuawei;1449573472;00002> to *cc_usage` {
		t.Error(err)
	}

	tplFlds = []*config.FCTemplate{
		&config.FCTemplate{Tag: "CCUsage", Filters: []string{},
			Path: utils.PathItems{{Field: "CCUsage"}}, Type: utils.MetaCCUsage,
			Value:     config.NewRSRParsersMustCompile("10;12s;~*req.Session-Id", true, utils.INFIELD_SEP),
			Mandatory: true},
	}
	if _, err := agReq.ParseField(tplFlds[0]); err == nil ||
		err.Error() != `invalid debitInterval <simuhuawei;1449573472;00002> to *cc_usage` {
		t.Error(err)
	}

	tplFlds = []*config.FCTemplate{
		&config.FCTemplate{Tag: "CCUsage", Filters: []string{},
			Path: utils.PathItems{{Field: "CCUsage"}}, Type: utils.MetaCCUsage,
			Value:     config.NewRSRParsersMustCompile("3;10s;5s", true, utils.INFIELD_SEP),
			Mandatory: true},
	}
	//5s*2 + 10s
	expected := time.Duration(20 * time.Second)
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
	cfg, _ := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items),
		config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	//pass the data provider to agent request
	agReq := NewAgentRequest(dP, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)

	tplFlds := []*config.FCTemplate{
		&config.FCTemplate{Tag: "Usage", Filters: []string{},
			Path: utils.PathItems{{Field: "Usage"}}, Type: utils.META_USAGE_DIFFERENCE,
			Value:     config.NewRSRParsersMustCompile("~*req.Session-Id", true, utils.INFIELD_SEP),
			Mandatory: true},
	}
	if _, err := agReq.ParseField(tplFlds[0]); err == nil ||
		err.Error() != `invalid arguments <[{"Rules":"~*req.Session-Id","AllFiltersMatch":true}]> to *usage_difference` {
		t.Error(err)
	}

	tplFlds = []*config.FCTemplate{
		&config.FCTemplate{Tag: "Usage", Filters: []string{},
			Path: utils.PathItems{{Field: "Usage"}}, Type: utils.META_USAGE_DIFFERENCE,
			Value:     config.NewRSRParsersMustCompile("1560325161;~*req.Session-Id", true, utils.INFIELD_SEP),
			Mandatory: true},
	}
	if _, err := agReq.ParseField(tplFlds[0]); err == nil ||
		err.Error() != `Unsupported time format` {
		t.Error(err)
	}

	tplFlds = []*config.FCTemplate{
		&config.FCTemplate{Tag: "Usage", Filters: []string{},
			Path: utils.PathItems{{Field: "Usage"}}, Type: utils.META_USAGE_DIFFERENCE,
			Value:     config.NewRSRParsersMustCompile("~*req.Session-Id;1560325161", true, utils.INFIELD_SEP),
			Mandatory: true},
	}
	if _, err := agReq.ParseField(tplFlds[0]); err == nil ||
		err.Error() != `Unsupported time format` {
		t.Error(err)
	}

	tplFlds = []*config.FCTemplate{
		&config.FCTemplate{Tag: "Usage", Filters: []string{},
			Path: utils.PathItems{{Field: "Usage"}}, Type: utils.META_USAGE_DIFFERENCE,
			Value:     config.NewRSRParsersMustCompile("1560325161;1560325151", true, utils.INFIELD_SEP),
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
	cfg, _ := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	//pass the data provider to agent request
	agReq := NewAgentRequest(dP, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)

	tplFlds := []*config.FCTemplate{
		&config.FCTemplate{Tag: "Sum", Filters: []string{},
			Path: utils.PathItems{{Field: "Sum"}}, Type: utils.MetaSum,
			Value:     config.NewRSRParsersMustCompile("15;~*req.Session-Id", true, utils.INFIELD_SEP),
			Mandatory: true},
	}
	if _, err := agReq.ParseField(tplFlds[0]); err == nil ||
		err.Error() != `strconv.ParseInt: parsing "simuhuawei;1449573472;00002": invalid syntax` {
		t.Error(err)
	}

	tplFlds = []*config.FCTemplate{
		&config.FCTemplate{Tag: "Sum", Filters: []string{},
			Path: utils.PathItems{{Field: "Sum"}}, Type: utils.MetaSum,
			Value:     config.NewRSRParsersMustCompile("15;15", true, utils.INFIELD_SEP),
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
	cfg, _ := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items),
		config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	//pass the data provider to agent request
	agReq := NewAgentRequest(dP, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)

	tplFlds := []*config.FCTemplate{
		&config.FCTemplate{Tag: "Diff", Filters: []string{},
			Path: utils.PathItems{{Field: "Diff"}}, Type: utils.MetaDifference,
			Value:     config.NewRSRParsersMustCompile("15;~*req.Session-Id", true, utils.INFIELD_SEP),
			Mandatory: true},
	}
	if _, err := agReq.ParseField(tplFlds[0]); err == nil ||
		err.Error() != `strconv.ParseInt: parsing "simuhuawei;1449573472;00002": invalid syntax` {
		t.Error(err)
	}

	tplFlds = []*config.FCTemplate{
		&config.FCTemplate{Tag: "Diff", Filters: []string{},
			Path: utils.PathItems{{Field: "Diff"}}, Type: utils.MetaDifference,
			Value:     config.NewRSRParsersMustCompile("15;12;2", true, utils.INFIELD_SEP),
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
	cfg, _ := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	//pass the data provider to agent request
	agReq := NewAgentRequest(dP, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)

	tplFlds := []*config.FCTemplate{
		&config.FCTemplate{Tag: "Multiply", Filters: []string{},
			Path: utils.PathItems{{Field: "Multiply"}}, Type: utils.MetaMultiply,
			Value:     config.NewRSRParsersMustCompile("15;~*req.Session-Id", true, utils.INFIELD_SEP),
			Mandatory: true},
	}
	if _, err := agReq.ParseField(tplFlds[0]); err == nil ||
		err.Error() != `strconv.ParseInt: parsing "simuhuawei;1449573472;00002": invalid syntax` {
		t.Error(err)
	}

	tplFlds = []*config.FCTemplate{
		&config.FCTemplate{Tag: "Multiply", Filters: []string{},
			Path: utils.PathItems{{Field: "Multiply"}}, Type: utils.MetaMultiply,
			Value:     config.NewRSRParsersMustCompile("15;15", true, utils.INFIELD_SEP),
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
	cfg, _ := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	//pass the data provider to agent request
	agReq := NewAgentRequest(dP, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)

	tplFlds := []*config.FCTemplate{
		&config.FCTemplate{Tag: "Divide", Filters: []string{},
			Path: utils.PathItems{{Field: "Divide"}}, Type: utils.MetaDivide,
			Value:     config.NewRSRParsersMustCompile("15;~*req.Session-Id", true, utils.INFIELD_SEP),
			Mandatory: true},
	}
	if _, err := agReq.ParseField(tplFlds[0]); err == nil ||
		err.Error() != `strconv.ParseInt: parsing "simuhuawei;1449573472;00002": invalid syntax` {
		t.Error(err)
	}

	tplFlds = []*config.FCTemplate{
		&config.FCTemplate{Tag: "Divide", Filters: []string{},
			Path: utils.PathItems{{Field: "Divide"}}, Type: utils.MetaDivide,
			Value:     config.NewRSRParsersMustCompile("15;3", true, utils.INFIELD_SEP),
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
	cfg, _ := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	//pass the data provider to agent request
	agReq := NewAgentRequest(dP, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)

	tplFlds := []*config.FCTemplate{
		&config.FCTemplate{Tag: "ValExp", Filters: []string{},
			Path: utils.PathItems{{Field: "ValExp"}}, Type: utils.MetaValueExponent,
			Value:     config.NewRSRParsersMustCompile("~*req.Session-Id", true, utils.INFIELD_SEP),
			Mandatory: true},
	}
	if _, err := agReq.ParseField(tplFlds[0]); err == nil ||
		err.Error() != `invalid arguments <[{"Rules":"~*req.Session-Id","AllFiltersMatch":true}]> to *value_exponent` {
		t.Error(err)
	}

	tplFlds = []*config.FCTemplate{
		&config.FCTemplate{Tag: "ValExp", Filters: []string{},
			Path: utils.PathItems{{Field: "ValExp"}}, Type: utils.MetaValueExponent,
			Value:     config.NewRSRParsersMustCompile("15;~*req.Session-Id", true, utils.INFIELD_SEP),
			Mandatory: true},
	}
	if _, err := agReq.ParseField(tplFlds[0]); err == nil ||
		err.Error() != `strconv.Atoi: parsing "simuhuawei;1449573472;00002": invalid syntax` {
		t.Error(err)
	}

	tplFlds = []*config.FCTemplate{
		&config.FCTemplate{Tag: "ValExp", Filters: []string{},
			Path: utils.PathItems{{Field: "ValExp"}}, Type: utils.MetaValueExponent,
			Value:     config.NewRSRParsersMustCompile("~*req.Session-Id;15", true, utils.INFIELD_SEP),
			Mandatory: true},
	}
	if _, err := agReq.ParseField(tplFlds[0]); err == nil ||
		err.Error() != `invalid value <simuhuawei;1449573472;00002> to *value_exponent` {
		t.Error(err)
	}
	tplFlds = []*config.FCTemplate{
		&config.FCTemplate{Tag: "ValExp", Filters: []string{},
			Path: utils.PathItems{{Field: "ValExp"}}, Type: utils.MetaValueExponent,
			Value:     config.NewRSRParsersMustCompile("2;3", true, utils.INFIELD_SEP),
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
	cfg, _ := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	agReq := NewAgentRequest(nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)
	// populate request, emulating the way will be done in HTTPAgent
	agReq.CGRRequest.Set([]*utils.PathItem{{Field: utils.ToR}}, utils.NewNMInterface(utils.VOICE))
	agReq.CGRRequest.Set([]*utils.PathItem{{Field: utils.Account}}, utils.NewNMInterface("1001"))
	agReq.CGRRequest.Set([]*utils.PathItem{{Field: utils.Destination}}, utils.NewNMInterface("1002"))
	agReq.CGRRequest.Set([]*utils.PathItem{{Field: utils.AnswerTime}},
		utils.NewNMInterface(time.Date(2013, 12, 30, 15, 0, 1, 0, time.UTC)))
	agReq.CGRRequest.Set([]*utils.PathItem{{Field: utils.RequestType}}, utils.NewNMInterface(utils.META_PREPAID))

	agReq.CGRReply = &utils.NavigableMap2{}

	tplFlds := []*config.FCTemplate{
		&config.FCTemplate{Tag: "Account",
			Path: utils.PathItems{{Field: utils.MetaCgrep}, {Field: utils.Account}}, Type: utils.META_COMPOSED,
			Value: config.NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Tag: "Account",
			Path: utils.PathItems{{Field: utils.MetaCgrep}, {Field: utils.Account}}, Type: utils.META_COMPOSED,
			Value: config.NewRSRParsersMustCompile(":", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Tag: "Account",
			Path: utils.PathItems{{Field: utils.MetaCgrep}, {Field: utils.Account}}, Type: utils.META_COMPOSED,
			Value: config.NewRSRParsersMustCompile("~*cgreq.Account", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Tag: "Account",
			Path: utils.PathItems{{Field: utils.MetaCgrep}, {Field: utils.Account}}, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("OverwrittenAccount", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Tag: "Account",
			Path: utils.PathItems{{Field: utils.MetaCgrep}, {Field: utils.Account}}, Type: utils.META_COMPOSED,
			Value: config.NewRSRParsersMustCompile("WithComposed", true, utils.INFIELD_SEP)},
	}

	if err := agReq.SetFields(tplFlds); err != nil {
		t.Error(err)
	}

	if rcv, err := agReq.CGRReply.FieldAsInterface([]string{utils.Account}); err != nil {
		t.Error(err)
	} else if sls, canCast := rcv.(*utils.NMSlice); !canCast {
		t.Errorf("Cannot cast to &utils.NMSlice %+v", rcv)
	} else if len(*sls) != 1 {
		t.Errorf("expecting: %+v, \n received: %+v ", 1, len(*sls))
	} else if nmi, canCast := (*sls)[0].(*config.NMItem); !canCast {
		t.Error("Expecting NM item")
	} else if nmi.Data != "OverwrittenAccountWithComposed" {
		t.Errorf("expecting: %+v, \n received: %+v ",
			"OverwrittenAccountWithComposed", nmi.Data)
	}
}

func TestAgReqGroupType(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	agReq := NewAgentRequest(nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)
	// populate request, emulating the way will be done in HTTPAgent
	agReq.CGRRequest.Set([]*utils.PathItem{{Field: utils.ToR}}, utils.NewNMInterface(utils.VOICE))
	agReq.CGRRequest.Set([]*utils.PathItem{{Field: utils.Account}}, utils.NewNMInterface("1001"))
	agReq.CGRRequest.Set([]*utils.PathItem{{Field: utils.Destination}}, utils.NewNMInterface("1002"))
	agReq.CGRRequest.Set([]*utils.PathItem{{Field: utils.AnswerTime}},
		utils.NewNMInterface(time.Date(2013, 12, 30, 15, 0, 1, 0, time.UTC)))
	agReq.CGRRequest.Set([]*utils.PathItem{{Field: utils.RequestType}}, utils.NewNMInterface(utils.META_PREPAID))

	agReq.CGRReply = &utils.NavigableMap2{}

	tplFlds := []*config.FCTemplate{
		&config.FCTemplate{Tag: "Account",
			Path: utils.PathItems{{Field: utils.MetaCgrep}, {Field: utils.Account}}, Type: utils.MetaGroup,
			Value: config.NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Tag: "Account",
			Path: utils.PathItems{{Field: utils.MetaCgrep}, {Field: utils.Account}}, Type: utils.MetaGroup,
			Value: config.NewRSRParsersMustCompile("test", true, utils.INFIELD_SEP)},
	}

	if err := agReq.SetFields(tplFlds); err != nil {
		t.Error(err)
	}

	if rcv, err := agReq.CGRReply.FieldAsInterface([]string{utils.Account}); err != nil {
		t.Error(err)
	} else if sls, canCast := rcv.(*utils.NMSlice); !canCast {
		t.Errorf("Cannot cast to &utils.NMSlice %+v", rcv)
	} else if len(*sls) != 2 {
		t.Errorf("expecting: %+v, \n received: %+v ", 1, len(*sls))
	} else if nmi, canCast := (*sls)[0].(*config.NMItem); !canCast {
		t.Error("Expecting NM item")
	} else if nmi.Data != "cgrates.org" {
		t.Errorf("expecting: %+v, \n received: %+v ", "cgrates.org", nmi.Data)
	} else if nmi, canCast := (*sls)[1].(*config.NMItem); !canCast {
		t.Error("Expecting NM item")
	} else if nmi.Data != "test" {
		t.Errorf("expecting: %+v, \n received: %+v ", "test", nmi.Data)
	}
}

func TestAgReqSetFieldsInTmp(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	agReq := NewAgentRequest(nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil, nil)
	agReq.CGRRequest.Set([]*utils.PathItem{{Field: utils.Account}}, utils.NewNMInterface("1001"))

	tplFlds := []*config.FCTemplate{
		&config.FCTemplate{Tag: "Tenant",
			Path: utils.PathItems{{Field: utils.MetaTmp}, {Field: utils.Tenant}}, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Tag: "Account",
			Path: utils.PathItems{{Field: utils.MetaTmp}, {Field: utils.Account}}, Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*cgreq.Account", true, utils.INFIELD_SEP)},
	}
	eMp := utils.NavigableMap2{}
	eMp.Set([]*utils.PathItem{{Field: utils.Tenant}}, &utils.NMSlice{
		&config.NMItem{Data: "cgrates.org", Path: []string{utils.Tenant},
			Config: tplFlds[0]}})
	eMp.Set([]*utils.PathItem{{Field: utils.Account}}, &utils.NMSlice{
		&config.NMItem{Data: "1001", Path: []string{utils.Account},
			Config: tplFlds[1]}})

	if err := agReq.SetFields(tplFlds); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(agReq.tmp, eMp) {
		t.Errorf("expecting: %+v,\n received: %+v", eMp, agReq.tmp)
	}
}
