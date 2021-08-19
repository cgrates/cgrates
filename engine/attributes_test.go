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
	"bytes"
	"fmt"
	"log"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestAttributesShutdown(t *testing.T) {
	alS := &AttributeService{}

	utils.Logger.SetLogLevel(6)
	utils.Logger.SetSyslog(nil)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	exp := []string{
		"CGRateS <> [INFO] <AttributeS> shutdown initialized",
		"CGRateS <> [INFO] <AttributeS> shutdown complete",
	}
	alS.Shutdown()
	rcv := strings.Split(buf.String(), "\n")

	for i := 0; i < 2; i++ {
		rcv[i] = rcv[i][20:]
		if rcv[i] != exp[i] {
			t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp[i], rcv[i])
		}
	}

	utils.Logger.SetLogLevel(0)
}

func TestAttributesSetCloneable(t *testing.T) {
	attr := &AttrArgsProcessEvent{
		AttributeIDs: []string{"ATTR_ID"},
		Context:      utils.StringPointer(utils.MetaAny),
		ProcessRuns:  utils.IntPointer(1),
		clnb:         true,
	}

	exp := &AttrArgsProcessEvent{
		AttributeIDs: []string{"ATTR_ID"},
		Context:      utils.StringPointer(utils.MetaAny),
		ProcessRuns:  utils.IntPointer(1),
		clnb:         false,
	}
	attr.SetCloneable(false)

	if !reflect.DeepEqual(attr, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, attr)
	}
}

func TestAttributesRPCClone(t *testing.T) {
	attr := &AttrArgsProcessEvent{
		AttributeIDs: []string{"ATTR_ID"},
		Context:      utils.StringPointer(utils.MetaAny),
		ProcessRuns:  utils.IntPointer(1),
		CGREvent: &utils.CGREvent{
			Event:   make(map[string]interface{}),
			APIOpts: make(map[string]interface{}),
		},
		clnb: true,
	}

	rcv, err := attr.RPCClone()

	exp := &AttrArgsProcessEvent{
		AttributeIDs: []string{"ATTR_ID"},
		Context:      utils.StringPointer(utils.MetaAny),
		ProcessRuns:  utils.IntPointer(1),
		CGREvent: &utils.CGREvent{
			Event:   make(map[string]interface{}),
			APIOpts: make(map[string]interface{}),
		},
		clnb: false,
	}

	if err != nil {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}

}

func TestAttributesV1GetAttributeForEventNilCGREvent(t *testing.T) {
	alS := &AttributeService{}
	args := &AttrArgsProcessEvent{}
	reply := &AttributeProfile{}

	experr := fmt.Sprintf("MANDATORY_IE_MISSING: [%s]", "CGREvent")
	err := alS.V1GetAttributeForEvent(args, reply)

	if err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestAttributesV1GetAttributeForEventProfileNotFound(t *testing.T) {
	defaultCfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, nil, nil)
	alS := &AttributeService{
		dm:      dm,
		filterS: &FilterS{},
		cgrcfg:  defaultCfg,
	}
	args := &AttrArgsProcessEvent{
		CGREvent: &utils.CGREvent{},
	}
	reply := &AttributeProfile{}

	experr := utils.ErrNotFound
	err := alS.V1GetAttributeForEvent(args, reply)

	if err == nil || err != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestAttributesV1GetAttributeForEvent2(t *testing.T) {
	defaultCfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, nil, nil)
	alS := &AttributeService{
		dm:      dm,
		filterS: &FilterS{},
		cgrcfg:  defaultCfg,
	}
	args := &AttrArgsProcessEvent{
		CGREvent: &utils.CGREvent{},
	}
	reply := &AttributeProfile{}

	experr := utils.ErrNotFound
	err := alS.V1GetAttributeForEvent(args, reply)

	if err == nil || err != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestAttributesV1ProcessEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.FilterSCfg().ResourceSConns = []string{}
	conMng := &ConnManager{}
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, nil, conMng)
	filterS := NewFilterS(cfg, conMng, dm)
	attr := &AttributeProfile{
		Tenant:             "cgrates.org",
		ID:                 "ATTR_CHANGE_TENANT_FROM_USER",
		Contexts:           []string{utils.MetaAny},
		FilterIDs:          []string{"*string:~*req.Account:dan@itsyscom.com|adrian@itsyscom.com"},
		ActivationInterval: nil,
		Attributes: []*Attribute{
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*variable",
				Value:     config.NewRSRParsersMustCompile("~*req.Account:s/(.*)@(.*)/${1}.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*req.Account",
				Type:      "*variable",
				Value:     config.NewRSRParsersMustCompile("~*req.Account:s/(dan)@(.*)/${1}.${2}/:s/(adrian)@(.*)/andrei.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*composed",
				Value:     config.NewRSRParsersMustCompile(".co.uk", utils.InfieldSep),
			},
		},
		Blocker: false,
		Weight:  20,
	}
	err := dm.SetAttributeProfile(attr, true)
	if err != nil {
		t.Error(err)
	}

	attr2 := &AttributeProfile{
		Tenant:             "adrian.itsyscom.com.co.uk",
		ID:                 "ATTR_MATCH_TENANT",
		Contexts:           []string{utils.MetaAny},
		ActivationInterval: nil,
		Attributes: []*Attribute{
			{
				FilterIDs: nil,
				Path:      "*req.Password",
				Type:      utils.MetaConstant,
				Value:     config.NewRSRParsersMustCompile("CGRATES.ORG", utils.InfieldSep),
			},
		},
		Blocker: false,
		Weight:  20,
	}

	err = dm.SetAttributeProfile(attr2, true)
	if err != nil {
		t.Error(err)
	}

	alS := NewAttributeService(dm, filterS, cfg)
	args := &AttrArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "123",
			Event: map[string]interface{}{
				utils.AccountField: "adrian@itsyscom.com",
			},
		},
		ProcessRuns: utils.IntPointer(2),
	}
	rply := &AttrSProcessEventReply{}
	expected := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_CHANGE_TENANT_FROM_USER", "adrian.itsyscom.com.co.uk:ATTR_MATCH_TENANT"},
		AlteredFields:   []string{"*req.Account", "*req.Password", "*tenant"},
		CGREvent: &utils.CGREvent{
			Tenant: "adrian.itsyscom.com.co.uk",
			ID:     "123",
			Time:   nil,
			Event: map[string]interface{}{
				utils.AccountField: "andrei.itsyscom.com",
				"Password":         "CGRATES.ORG",
			},
			APIOpts: map[string]interface{}{},
		},
		blocker: false,
	}
	err = alS.V1ProcessEvent(args, rply)
	sort.Strings(rply.AlteredFields)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expected), utils.ToJSON(rply))
	}
}

func TestAttributesV1ProcessEventErrorMetaSum(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.FilterSCfg().ResourceSConns = []string{}
	conMng := &ConnManager{}
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, nil, conMng)
	filterS := NewFilterS(cfg, conMng, dm)
	attr := &AttributeProfile{
		Tenant:             "cgrates.org",
		ID:                 "ATTR_CHANGE_TENANT_FROM_USER",
		Contexts:           []string{utils.MetaAny},
		FilterIDs:          []string{"*string:~*req.Account:dan@itsyscom.com|adrian@itsyscom.com"},
		ActivationInterval: nil,
		Attributes: []*Attribute{
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*variable",
				Value:     config.NewRSRParsersMustCompile("~*req.Account:s/(.*)@(.*)/${1}.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*req.Account",
				Type:      "*variable",
				Value:     config.NewRSRParsersMustCompile("~*req.Account:s/(dan)@(.*)/${1}.${2}/:s/(adrian)@(.*)/andrei.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*composed",
				Value:     config.NewRSRParsersMustCompile(".co.uk", utils.InfieldSep),
			},
		},
		Blocker: false,
		Weight:  20,
	}
	err := dm.SetAttributeProfile(attr, true)
	if err != nil {
		t.Error(err)
	}

	attr2 := &AttributeProfile{
		Tenant:             "adrian.itsyscom.com.co.uk",
		ID:                 "ATTR_MATCH_TENANT",
		Contexts:           []string{utils.MetaAny},
		ActivationInterval: nil,
		Attributes: []*Attribute{
			{
				FilterIDs: nil,
				Path:      "*req.Password",
				Type:      utils.MetaSum,
				Value:     config.NewRSRParsersMustCompile("CGRATES.ORG", utils.InfieldSep),
			},
		},
		Blocker: false,
		Weight:  20,
	}

	err = dm.SetAttributeProfile(attr2, true)
	if err != nil {
		t.Error(err)
	}

	alS := NewAttributeService(dm, filterS, cfg)
	args := &AttrArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "123",
			Event: map[string]interface{}{
				utils.AccountField: "adrian@itsyscom.com",
			},
		},
		ProcessRuns: utils.IntPointer(2),
	}
	rply := &AttrSProcessEventReply{}
	err = alS.V1ProcessEvent(args, rply)
	sort.Strings(rply.AlteredFields)
	expErr := "SERVER_ERROR: NotEnoughParameters"
	if err == nil || err.Error() != expErr {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}

}

func TestAttributesV1ProcessEventErrorMetaDifference(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.FilterSCfg().ResourceSConns = []string{}
	conMng := &ConnManager{}
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, nil, conMng)
	filterS := NewFilterS(cfg, conMng, dm)
	attr := &AttributeProfile{
		Tenant:             "cgrates.org",
		ID:                 "ATTR_CHANGE_TENANT_FROM_USER",
		Contexts:           []string{utils.MetaAny},
		FilterIDs:          []string{"*string:~*req.Account:dan@itsyscom.com|adrian@itsyscom.com"},
		ActivationInterval: nil,
		Attributes: []*Attribute{
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*variable",
				Value:     config.NewRSRParsersMustCompile("~*req.Account:s/(.*)@(.*)/${1}.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*req.Account",
				Type:      "*variable",
				Value:     config.NewRSRParsersMustCompile("~*req.Account:s/(dan)@(.*)/${1}.${2}/:s/(adrian)@(.*)/andrei.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*composed",
				Value:     config.NewRSRParsersMustCompile(".co.uk", utils.InfieldSep),
			},
		},
		Blocker: false,
		Weight:  20,
	}
	err := dm.SetAttributeProfile(attr, true)
	if err != nil {
		t.Error(err)
	}

	attr2 := &AttributeProfile{
		Tenant:             "adrian.itsyscom.com.co.uk",
		ID:                 "ATTR_MATCH_TENANT",
		Contexts:           []string{utils.MetaAny},
		ActivationInterval: nil,
		Attributes: []*Attribute{
			{
				FilterIDs: nil,
				Path:      "*req.Password",
				Type:      utils.MetaDifference,
				Value:     config.NewRSRParsersMustCompile("CGRATES.ORG", utils.InfieldSep),
			},
		},
		Blocker: false,
		Weight:  20,
	}

	err = dm.SetAttributeProfile(attr2, true)
	if err != nil {
		t.Error(err)
	}

	alS := NewAttributeService(dm, filterS, cfg)
	args := &AttrArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "123",
			Event: map[string]interface{}{
				utils.AccountField: "adrian@itsyscom.com",
			},
		},
		ProcessRuns: utils.IntPointer(2),
	}
	rply := &AttrSProcessEventReply{}
	err = alS.V1ProcessEvent(args, rply)
	sort.Strings(rply.AlteredFields)
	expErr := "SERVER_ERROR: NotEnoughParameters"
	if err == nil || err.Error() != expErr {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}

}

func TestAttributesV1ProcessEventErrorMetaValueExponent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.FilterSCfg().ResourceSConns = []string{}
	conMng := &ConnManager{}
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, nil, conMng)
	filterS := NewFilterS(cfg, conMng, dm)
	attr := &AttributeProfile{
		Tenant:             "cgrates.org",
		ID:                 "ATTR_CHANGE_TENANT_FROM_USER",
		Contexts:           []string{utils.MetaAny},
		FilterIDs:          []string{"*string:~*req.Account:dan@itsyscom.com|adrian@itsyscom.com"},
		ActivationInterval: nil,
		Attributes: []*Attribute{
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*variable",
				Value:     config.NewRSRParsersMustCompile("~*req.Account:s/(.*)@(.*)/${1}.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*req.Account",
				Type:      "*variable",
				Value:     config.NewRSRParsersMustCompile("~*req.Account:s/(dan)@(.*)/${1}.${2}/:s/(adrian)@(.*)/andrei.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*composed",
				Value:     config.NewRSRParsersMustCompile(".co.uk", utils.InfieldSep),
			},
		},
		Blocker: false,
		Weight:  20,
	}
	err := dm.SetAttributeProfile(attr, true)
	if err != nil {
		t.Error(err)
	}

	attr2 := &AttributeProfile{
		Tenant:             "adrian.itsyscom.com.co.uk",
		ID:                 "ATTR_MATCH_TENANT",
		Contexts:           []string{utils.MetaAny},
		ActivationInterval: nil,
		Attributes: []*Attribute{
			{
				FilterIDs: nil,
				Path:      "*req.Password",
				Type:      utils.MetaValueExponent,
				Value:     config.NewRSRParsersMustCompile("CGRATES.ORG", utils.InfieldSep),
			},
		},
		Blocker: false,
		Weight:  20,
	}

	err = dm.SetAttributeProfile(attr2, true)
	if err != nil {
		t.Error(err)
	}

	alS := NewAttributeService(dm, filterS, cfg)
	args := &AttrArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "123",
			Event: map[string]interface{}{
				utils.AccountField: "adrian@itsyscom.com",
			},
		},
		ProcessRuns: utils.IntPointer(2),
	}
	rply := &AttrSProcessEventReply{}
	err = alS.V1ProcessEvent(args, rply)
	sort.Strings(rply.AlteredFields)
	expErr := "SERVER_ERROR: invalid arguments <[{\"Rules\":\"CGRATES.ORG\"}]> to *value_exponent"
	if err == nil || err.Error() != expErr {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}

}

func TestAttributesattributeProfileForEventAnyCtxFalseNotFound(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, true)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	alS := &AttributeService{
		cgrcfg:  cfg,
		dm:      dm,
		filterS: NewFilterS(cfg, nil, dm),
	}
	alS.cgrcfg.AttributeSCfg().AnyContext = false

	postpaid, err := config.NewRSRParsers(utils.MetaPostpaid, utils.InfieldSep)
	if err != nil {
		t.Error(err)
	}
	ap1 := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_1",
		FilterIDs: []string{"*string:~*req.Account:1002"},
		Contexts:  []string{utils.MetaSessionS},
		Attributes: []*Attribute{
			{
				Path:  "*req.RequestType",
				Type:  utils.MetaConstant,
				Value: postpaid,
			},
		},
		Weight: 20,
	}
	err = alS.dm.SetAttributeProfile(ap1, true)
	if err != nil {
		t.Error(err)
	}

	ap2 := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_2",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Contexts:  []string{utils.MetaAny},
		Attributes: []*Attribute{
			{
				Path:  "*req.RequestType",
				Type:  utils.MetaConstant,
				Value: postpaid,
			},
		},
		Weight: 10,
	}
	err = alS.dm.SetAttributeProfile(ap2, true)
	if err != nil {
		t.Error(err)
	}

	tnt := "cgrates.org"
	ctx := utils.StringPointer(utils.MetaSessionS)
	evNm := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			utils.AccountField: "1001",
		},
		utils.MetaVars: utils.MapStorage{},
	}
	lastID := ""

	if rcv, err := alS.attributeProfileForEvent(tnt, ctx, nil, nil, evNm,
		lastID); err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	} else if !reflect.DeepEqual(rcv, ap2) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", ap2, rcv)
	}

	lastID = "cgrates.org:ATTR_2"

	if rcv, err := alS.attributeProfileForEvent(tnt, ctx, nil, nil, evNm,
		lastID); err == nil || err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	} else if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}
}

func TestAttributesattributeProfileForEventAnyCtxFalseFound(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, true)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	alS := &AttributeService{
		cgrcfg:  cfg,
		dm:      dm,
		filterS: NewFilterS(cfg, nil, dm),
	}
	alS.cgrcfg.AttributeSCfg().AnyContext = false

	postpaid, err := config.NewRSRParsers(utils.MetaPostpaid, utils.InfieldSep)
	if err != nil {
		t.Error(err)
	}
	ap1 := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Contexts:  []string{utils.MetaSessionS},
		Attributes: []*Attribute{
			{
				Path:  "*req.RequestType",
				Type:  utils.MetaConstant,
				Value: postpaid,
			},
		},
		Weight: 20,
	}
	err = alS.dm.SetAttributeProfile(ap1, true)
	if err != nil {
		t.Error(err)
	}

	ap2 := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_2",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Contexts:  []string{utils.MetaAny},
		Attributes: []*Attribute{
			{
				Path:  "*req.RequestType",
				Type:  utils.MetaConstant,
				Value: postpaid,
			},
		},
		Weight: 10,
	}
	err = alS.dm.SetAttributeProfile(ap2, true)
	if err != nil {
		t.Error(err)
	}

	tnt := "cgrates.org"
	ctx := utils.StringPointer(utils.MetaSessionS)
	evNm := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			utils.AccountField: "1001",
		},
		utils.MetaVars: utils.MapStorage{},
	}
	lastID := ""

	if rcv, err := alS.attributeProfileForEvent(tnt, ctx, nil, nil, evNm,
		lastID); err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	} else if !reflect.DeepEqual(rcv, ap1) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", ap1, rcv)
	}
}

func TestAttributesattributeProfileForEventAnyCtxTrueBothFound(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, true)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	alS := &AttributeService{
		cgrcfg:  cfg,
		dm:      dm,
		filterS: NewFilterS(cfg, nil, dm),
	}

	postpaid, err := config.NewRSRParsers(utils.MetaPostpaid, utils.InfieldSep)
	if err != nil {
		t.Error(err)
	}
	ap1 := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Contexts:  []string{utils.MetaSessionS},
		Attributes: []*Attribute{
			{
				Path:  "*req.RequestType",
				Type:  utils.MetaConstant,
				Value: postpaid,
			},
		},
		Weight: 20,
	}
	err = alS.dm.SetAttributeProfile(ap1, true)
	if err != nil {
		t.Error(err)
	}

	ap2 := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_2",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Contexts:  []string{utils.MetaAny},
		Attributes: []*Attribute{
			{
				Path:  "*req.RequestType",
				Type:  utils.MetaConstant,
				Value: postpaid,
			},
		},
		Weight: 10,
	}
	err = alS.dm.SetAttributeProfile(ap2, true)
	if err != nil {
		t.Error(err)
	}

	tnt := "cgrates.org"
	ctx := utils.StringPointer(utils.MetaSessionS)
	evNm := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			utils.AccountField: "1001",
		},
		utils.MetaVars: utils.MapStorage{},
	}
	lastID := ""

	if rcv, err := alS.attributeProfileForEvent(tnt, ctx, nil, nil, evNm,
		lastID); err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	} else if !reflect.DeepEqual(rcv, ap1) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", ap1, rcv)
	}

	ap2.Weight = 30
	err = alS.dm.SetAttributeProfile(ap2, true)
	if err != nil {
		t.Error(err)
	}

	if rcv, err := alS.attributeProfileForEvent(tnt, ctx, nil, nil, evNm,
		lastID); err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	} else if !reflect.DeepEqual(rcv, ap2) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", ap2, rcv)
	}
}

func TestAttributesattributeProfileForEventAnyCtxTrueErrMatching(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, true)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	alS := &AttributeService{
		cgrcfg:  cfg,
		dm:      dm,
		filterS: NewFilterS(cfg, nil, dm),
	}

	postpaid, err := config.NewRSRParsers(utils.MetaPostpaid, utils.InfieldSep)
	if err != nil {
		t.Error(err)
	}
	ap1 := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Contexts:  []string{utils.MetaSessionS},
		Attributes: []*Attribute{
			{
				Path:  "*req.RequestType",
				Type:  utils.MetaConstant,
				Value: postpaid,
			},
		},
		Weight: 20,
	}
	err = alS.dm.SetAttributeProfile(ap1, true)
	if err != nil {
		t.Error(err)
	}

	ap2 := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_2",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Contexts:  []string{utils.MetaAny},
		Attributes: []*Attribute{
			{
				Path:  "*req.RequestType",
				Type:  utils.MetaConstant,
				Value: postpaid,
			},
		},
		Weight: 10,
	}
	err = alS.dm.SetAttributeProfile(ap2, true)
	if err != nil {
		t.Error(err)
	}

	tnt := "cgrates.org"
	ctx := utils.StringPointer(utils.MetaSessionS)
	evNm := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			utils.AccountField: "1001",
		},
		utils.MetaVars: utils.MapStorage{},
	}
	lastID := ""

	dbm := &DataDBMock{
		GetKeysForPrefixF: func(s string) ([]string, error) {
			return nil, utils.ErrExists
		},
	}
	alS.cgrcfg.AttributeSCfg().IndexedSelects = false
	alS.dm = NewDataManager(dbm, cfg.CacheCfg(), nil)

	if rcv, err := alS.attributeProfileForEvent(tnt, ctx, nil, nil, evNm,
		lastID); err == nil || err != utils.ErrExists {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrExists, err)
	} else if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}
}

func TestAttributesattributeProfileForEventAnyCtxTrueNotFound(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, true)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	alS := &AttributeService{
		cgrcfg:  cfg,
		dm:      dm,
		filterS: NewFilterS(cfg, nil, dm),
	}

	postpaid, err := config.NewRSRParsers(utils.MetaPostpaid, utils.InfieldSep)
	if err != nil {
		t.Error(err)
	}
	ap1 := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_1",
		FilterIDs: []string{"*string:~*req.Account:1002"},
		Contexts:  []string{utils.MetaSessionS},
		Attributes: []*Attribute{
			{
				Path:  "*req.RequestType",
				Type:  utils.MetaConstant,
				Value: postpaid,
			},
		},
		Weight: 20,
	}
	err = alS.dm.SetAttributeProfile(ap1, true)
	if err != nil {
		t.Error(err)
	}

	ap2 := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_2",
		FilterIDs: []string{"*string:~*req.Account:1002"},
		Contexts:  []string{utils.MetaAny},
		Attributes: []*Attribute{
			{
				Path:  "*req.RequestType",
				Type:  utils.MetaConstant,
				Value: postpaid,
			},
		},
		Weight: 10,
	}
	err = alS.dm.SetAttributeProfile(ap2, true)
	if err != nil {
		t.Error(err)
	}

	tnt := "cgrates.org"
	ctx := utils.StringPointer(utils.MetaSessionS)
	evNm := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			utils.AccountField: "1001",
		},
		utils.MetaVars: utils.MapStorage{},
	}
	lastID := ""

	if rcv, err := alS.attributeProfileForEvent(tnt, ctx, nil, nil, evNm,
		lastID); err == nil || err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	} else if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}
}

func TestAttributesattributeProfileForEventNoDBConn(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, true)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	alS := &AttributeService{
		cgrcfg:  cfg,
		dm:      dm,
		filterS: NewFilterS(cfg, nil, dm),
	}

	postpaid, err := config.NewRSRParsers(utils.MetaPostpaid, utils.InfieldSep)
	if err != nil {
		t.Error(err)
	}
	ap1 := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Contexts:  []string{utils.MetaSessionS},
		Attributes: []*Attribute{
			{
				Path:  "*req.RequestType",
				Type:  utils.MetaConstant,
				Value: postpaid,
			},
		},
		Weight: 20,
	}
	err = alS.dm.SetAttributeProfile(ap1, true)
	if err != nil {
		t.Error(err)
	}

	ap2 := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_2",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Contexts:  []string{utils.MetaAny},
		Attributes: []*Attribute{
			{
				Path:  "*req.RequestType",
				Type:  utils.MetaConstant,
				Value: postpaid,
			},
		},
		Weight: 10,
	}
	err = alS.dm.SetAttributeProfile(ap2, true)
	if err != nil {
		t.Error(err)
	}

	tnt := "cgrates.org"
	ctx := utils.StringPointer(utils.MetaSessionS)
	evNm := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			utils.AccountField: "1001",
		},
		utils.MetaVars: utils.MapStorage{},
	}
	lastID := ""
	alS.dm = nil

	if rcv, err := alS.attributeProfileForEvent(tnt, ctx, []string{"ATTR_3"}, nil, evNm,
		lastID); err == nil || err != utils.ErrNoDatabaseConn {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNoDatabaseConn, err)
	} else if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}
}

func TestAttributesattributeProfileForEventErrNotFound(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, true)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	alS := &AttributeService{
		cgrcfg:  cfg,
		dm:      dm,
		filterS: NewFilterS(cfg, nil, dm),
	}

	apNil := &AttributeProfile{}
	err = alS.dm.SetAttributeProfile(apNil, true)
	if err != nil {
		t.Error(err)
	}

	tnt := ""
	ctx := utils.StringPointer(utils.MetaSessionS)
	evNm := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			utils.AccountField: "1001",
		},
		utils.MetaVars: utils.MapStorage{},
	}
	lastID := ""

	if rcv, err := alS.attributeProfileForEvent(tnt, ctx, []string{"ATTR_3"}, nil, evNm,
		lastID); err == nil || err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	} else if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}
}

func TestAttributesattributeProfileForEventNotActive(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, true)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	alS := &AttributeService{
		cgrcfg:  cfg,
		dm:      dm,
		filterS: NewFilterS(cfg, nil, dm),
	}

	postpaid, err := config.NewRSRParsers(utils.MetaPostpaid, utils.InfieldSep)
	if err != nil {
		t.Error(err)
	}
	ap := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Contexts:  []string{utils.MetaSessionS},
		Attributes: []*Attribute{
			{
				Path:  "*req.RequestType",
				Type:  utils.MetaConstant,
				Value: postpaid,
			},
		},
		Weight: 20,
	}

	ctx := utils.StringPointer(utils.MetaSessionS)
	evNm := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			utils.AccountField: "1001",
		},
	}
	lastID := ""

	ap.ActivationInterval = &utils.ActivationInterval{
		ExpiryTime: time.Date(2021, 5, 14, 15, 0, 0, 0, time.UTC),
	}
	err = alS.dm.SetAttributeProfile(ap, true)
	if err != nil {
		t.Error(err)
	}
	actTime := utils.TimePointer(time.Date(2021, 5, 14, 16, 0, 0, 0, time.UTC))
	tnt := "cgrates.org"

	if rcv, err := alS.attributeProfileForEvent(tnt, ctx, nil, actTime, evNm,
		lastID); err == nil || err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	} else if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}
}

func TestAttributesattributeProfileForEventErrPass(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, true)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	alS := &AttributeService{
		cgrcfg:  cfg,
		dm:      dm,
		filterS: NewFilterS(cfg, nil, dm),
	}

	postpaid, err := config.NewRSRParsers(utils.MetaPostpaid, utils.InfieldSep)
	if err != nil {
		t.Error(err)
	}
	ap := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Contexts:  []string{utils.MetaSessionS},
		Attributes: []*Attribute{
			{
				Path:  "*req.RequestType",
				Type:  utils.MetaConstant,
				Value: postpaid,
			},
		},
		Weight: 20,
	}
	err = alS.dm.SetAttributeProfile(ap, true)
	if err != nil {
		t.Error(err)
	}

	tnt := "cgrates.org"
	ctx := utils.StringPointer(utils.MetaSessionS)
	evNm := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			utils.AccountField: "1001",
		},
	}
	lastID := ""

	evNm = utils.MapStorage{
		utils.MetaReq:  1,
		utils.MetaVars: utils.MapStorage{},
	}

	if rcv, err := alS.attributeProfileForEvent(tnt, ctx, []string{"ATTR_1"}, nil, evNm,
		lastID); err == nil || err != utils.ErrWrongPath {
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
	if out, err := ParseAttribute(dp, utils.MetaSIPCID, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.cid;~*req.to;~*req.from", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString, utils.InfieldSep); err != nil {
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
	if out, err := ParseAttribute(dp, utils.MetaSIPCID, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.cid;~*req.to;~*req.from", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString, utils.InfieldSep); err != nil {
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
	if out, err := ParseAttribute(dp, utils.MetaSIPCID, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.cid;~*req.to;~*req.extra;~*req.from",
		utils.InfieldSep), 0, utils.EmptyString, utils.EmptyString, utils.InfieldSep); err != nil {
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
	if out, err := ParseAttribute(dp, utils.MetaSIPCID, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.cid;~*req.extra;~*req.to;~*req.from",
		utils.InfieldSep), 0, utils.EmptyString, utils.EmptyString, utils.InfieldSep); err != nil {
		t.Fatal(err)
	} else if exp != out {
		t.Errorf("Expected %q, Received %q", exp, out)
	}

	dp = utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"cid": "12345",
		},
	}
	if _, err := ParseAttribute(dp, utils.MetaSIPCID, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.cid;~*req.extra;~*req.to;~*req.from", utils.
		InfieldSep), 0, utils.EmptyString, utils.EmptyString, utils.InfieldSep); err != utils.ErrNotFound {
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
	value := config.NewRSRParsersMustCompile("~*req.cid;~*req.to;~*req.from;~*opts.WrongPath", utils.InfieldSep)
	if _, err := ParseAttribute(dp, utils.MetaSIPCID, utils.EmptyString, value,
		0, time.UTC.String(), utils.EmptyString, utils.InfieldSep); err == nil ||
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
	value := config.NewRSRParsersMustCompile("~*req.cid;~*req.to;~*req.from", utils.InfieldSep)
	if _, err := ParseAttribute(dp, utils.MetaSIPCID, utils.EmptyString, value,
		0, time.UTC.String(), utils.EmptyString, utils.InfieldSep); err == nil ||
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
	value := config.RSRParsers{}
	experr := `invalid number of arguments <[]> to *sipcid`
	if _, err := ParseAttribute(dp, utils.MetaSIPCID, utils.EmptyString, value,
		0, time.UTC.String(), utils.EmptyString, utils.InfieldSep); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestAttributesV1ProcessEventMultipleRuns1(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().IndexedSelects = false
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	Cache = NewCacheS(cfg, dm, nil)
	alS := NewAttributeService(dm, filterS, cfg)

	postpaid := config.NewRSRParsersMustCompile(utils.MetaPostpaid, utils.InfieldSep)
	pw := config.NewRSRParsersMustCompile("CGRateS.org", utils.InfieldSep)

	ap1 := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR1",
		FilterIDs: []string{"*notexists:~*vars.*processedProfileIDs[<~*vars.*apTenantID>]:"},
		Contexts:  []string{utils.MetaAny},
		Attributes: []*Attribute{
			{
				Path:  "*req.Password",
				Type:  utils.MetaConstant,
				Value: pw,
			},
		},
		Weight: 10,
	}
	err = alS.dm.SetAttributeProfile(ap1, true)
	if err != nil {
		t.Error(err)
	}

	ap2 := &AttributeProfile{
		Tenant:   "cgrates.org",
		ID:       "ATTR2",
		Contexts: []string{utils.MetaAny},
		Attributes: []*Attribute{
			{
				Path:  "*req.RequestType",
				Type:  utils.MetaConstant,
				Value: postpaid,
			},
		},
		Weight: 20,
	}
	err = alS.dm.SetAttributeProfile(ap2, true)
	if err != nil {
		t.Error(err)
	}

	args := &AttrArgsProcessEvent{
		AttributeIDs: []string{"ATTR1", "ATTR2"},
		Context:      utils.StringPointer(utils.MetaAny),
		ProcessRuns:  utils.IntPointer(3),
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "AttrProcessEventMultipleRuns",
			Event: map[string]interface{}{
				"Password": "passwd",
			},
		},
	}
	reply := &AttrSProcessEventReply{}
	exp := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR2", "cgrates.org:ATTR1", "cgrates.org:ATTR2"},
		AlteredFields:   []string{"*req.Password", "*req.RequestType"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "AttrProcessEventMultipleRuns",
			Event: map[string]interface{}{
				"Password":        "CGRateS.org",
				utils.RequestType: utils.MetaPostpaid,
			},
			APIOpts: make(map[string]interface{}),
		},
	}

	if err := alS.V1ProcessEvent(args, reply); err != nil {
		t.Error(err)
	} else {
		sort.Strings(reply.AlteredFields)
		if !reflect.DeepEqual(reply, exp) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(exp), utils.ToJSON(reply))
		}
	}
}

func TestAttributesV1ProcessEventMultipleRuns2(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().IndexedSelects = false
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	Cache = NewCacheS(cfg, dm, nil)
	alS := NewAttributeService(dm, filterS, cfg)

	postpaid := config.NewRSRParsersMustCompile(utils.MetaPostpaid, utils.InfieldSep)
	pw := config.NewRSRParsersMustCompile("CGRateS.org", utils.InfieldSep)
	paypal := config.NewRSRParsersMustCompile("cgrates@paypal.com", utils.InfieldSep)

	ap1 := &AttributeProfile{
		Tenant:   "cgrates.org",
		ID:       "ATTR1",
		Contexts: []string{utils.MetaAny},
		Attributes: []*Attribute{
			{
				Path:  "*req.Password",
				Type:  utils.MetaConstant,
				Value: pw,
			},
		},
		Weight: 10,
	}
	err = alS.dm.SetAttributeProfile(ap1, true)
	if err != nil {
		t.Error(err)
	}

	ap2 := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR2",
		FilterIDs: []string{"*exists:~*vars.*processedProfileIDs[cgrates.org:ATTR1]:"},
		Contexts:  []string{utils.MetaAny},
		Attributes: []*Attribute{
			{
				Path:  "*req.RequestType",
				Type:  utils.MetaConstant,
				Value: postpaid,
			},
		},
		Weight: 20,
	}
	err = alS.dm.SetAttributeProfile(ap2, true)
	if err != nil {
		t.Error(err)
	}

	ap3 := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR3",
		FilterIDs: []string{"*exists:~*vars.*processedProfileIDs[cgrates.org:ATTR2]:"},
		Contexts:  []string{utils.MetaAny},
		Attributes: []*Attribute{
			{
				Path:  "*req.PaypalAccount",
				Type:  utils.MetaConstant,
				Value: paypal,
			},
		},
		Weight: 30,
	}
	err = alS.dm.SetAttributeProfile(ap3, true)
	if err != nil {
		t.Error(err)
	}

	args := &AttrArgsProcessEvent{
		Context:     utils.StringPointer(utils.MetaAny),
		ProcessRuns: utils.IntPointer(3),
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "AttrProcessEventMultipleRuns",
			Event:  map[string]interface{}{},
		},
	}

	reply := &AttrSProcessEventReply{}
	exp := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR1", "cgrates.org:ATTR2", "cgrates.org:ATTR3"},
		AlteredFields:   []string{"*req.Password", "*req.PaypalAccount", "*req.RequestType"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "AttrProcessEventMultipleRuns",
			Event: map[string]interface{}{
				"Password":        "CGRateS.org",
				"PaypalAccount":   "cgrates@paypal.com",
				utils.RequestType: utils.MetaPostpaid,
			},
			APIOpts: make(map[string]interface{}),
		},
	}
	if err := alS.V1ProcessEvent(args, reply); err != nil {
		t.Error(err)
	} else {
		sort.Strings(reply.AlteredFields)
		if !reflect.DeepEqual(reply, exp) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(exp), utils.ToJSON(reply))
		}
	}
}

func TestAttributesPorcessEventMatchingProcessRuns(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().Enabled = true
	cfg.AttributeSCfg().IndexedSelects = false
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	fltrS := NewFilterS(cfg, nil, dm)
	fltr := &Filter{
		Tenant: "cgrates.org",
		ID:     "Process_Runs_Fltr",
		Rules: []*FilterRule{
			/*
				{
					Type: utils.MetaString,
					Element: "~*req.Account",
					Values: []string{"pc_test"},
				},

			*/
			{
				Type:    utils.MetaGreaterThan,
				Element: "~*vars.*processRuns",
				Values:  []string{"1"},
			},
		},
	}
	if err := dm.SetFilter(fltr, true); err != nil {
		t.Error(err)
	}

	attrPfr := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_ProcessRuns",
		Contexts:  []string{"*any"},
		FilterIDs: []string{"Process_Runs_Fltr"},
		Attributes: []*Attribute{
			{
				Path:  "*req.CompanyName",
				Type:  utils.MetaVariable,
				Value: config.NewRSRParsersMustCompile("ITSYS COMMUNICATIONS SRL", utils.InfieldSep),
			},
		},
		Weight: 20,
	}
	// this I'll match first, no fltr and processRuns will be 1
	attrPfr2 := &AttributeProfile{
		Tenant:   "cgrates.org",
		ID:       "ATTR_MatchSecond",
		Contexts: []string{"*any"},
		Attributes: []*Attribute{
			{
				Path:  "*req.Password",
				Type:  utils.MetaVariable,
				Value: config.NewRSRParsersMustCompile("CGRateS.org", utils.InfieldSep),
			},
		},
		Weight: 10,
	}

	attrPfr.Compile()
	fltr.Compile()
	attrPfr2.Compile()
	if err := dm.SetAttributeProfile(attrPfr, true); err != nil {
		t.Error(err)
	}
	if err := dm.SetAttributeProfile(attrPfr2, true); err != nil {
		t.Error(err)
	}

	attr := NewAttributeService(dm, fltrS, cfg)

	args := &AttrArgsProcessEvent{
		ProcessRuns: utils.IntPointer(2),
		Context:     utils.StringPointer(utils.MetaAny),
		CGREvent: &utils.CGREvent{
			Event: map[string]interface{}{
				"Account":     "pc_test",
				"CompanyName": "MY_company_will_be_changed",
			},
		},
	}
	reply := &AttrSProcessEventReply{}
	expReply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_MatchSecond", "cgrates.org:ATTR_ProcessRuns"},
		AlteredFields:   []string{"*req.Password", "*req.CompanyName"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				"Account":     "pc_test",
				"CompanyName": "ITSYS COMMUNICATIONS SRL",
				"Password":    "CGRateS.org",
			},
			APIOpts: map[string]interface{}{},
		},
	}
	if err := attr.V1ProcessEvent(args, reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expReply, reply) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expReply), utils.ToJSON(reply))
	}
}
