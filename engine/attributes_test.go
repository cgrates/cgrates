/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOev.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package engine

import (
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestAttributesV1ProcessEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.FilterSCfg().ResourceSConns = []string{}
	conMng := &ConnManager{}
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, nil, conMng)
	filterS := NewFilterS(cfg, conMng, dm)
	attr := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_CHANGE_TENANT_FROM_USER",
		FilterIDs: []string{"*string:~*req.Account:dan@itsyscom.com|adrian@itsyscom.com"},
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
	err := dm.SetAttributeProfile(context.Background(), attr, true)
	if err != nil {
		t.Error(err)
	}

	attr2 := &AttributeProfile{
		Tenant: "adrian.itsyscom.com.co.uk",
		ID:     "ATTR_MATCH_TENANT",
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

	err = dm.SetAttributeProfile(context.Background(), attr2, true)
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
			APIOpts: map[string]interface{}{
				utils.OptsAttributesProcessRuns: 2,
			},
		},
	}
	rply := &AttrSProcessEventReply{}
	expected := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_CHANGE_TENANT_FROM_USER", "adrian.itsyscom.com.co.uk:ATTR_MATCH_TENANT"},
		AlteredFields:   []string{"*req.Account", "*req.Password", "*tenant"},
		CGREvent: &utils.CGREvent{
			Tenant: "adrian.itsyscom.com.co.uk",
			ID:     "123",
			Event: map[string]interface{}{
				utils.AccountField: "andrei.itsyscom.com",
				"Password":         "CGRATES.ORG",
			},
			APIOpts: map[string]interface{}{
				utils.OptsAttributesProcessRuns: 2,
			},
		},
		blocker: false,
	}
	err = alS.V1ProcessEvent(context.Background(), args, rply)
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
		Tenant:    "cgrates.org",
		ID:        "ATTR_CHANGE_TENANT_FROM_USER",
		FilterIDs: []string{"*string:~*req.Account:dan@itsyscom.com|adrian@itsyscom.com"},
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
	err := dm.SetAttributeProfile(context.Background(), attr, true)
	if err != nil {
		t.Error(err)
	}

	attr2 := &AttributeProfile{
		Tenant: "adrian.itsyscom.com.co.uk",
		ID:     "ATTR_MATCH_TENANT",
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

	err = dm.SetAttributeProfile(context.Background(), attr2, true)
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
			APIOpts: map[string]interface{}{
				utils.OptsAttributesProcessRuns: 2,
			},
		},
	}
	rply := &AttrSProcessEventReply{}
	err = alS.V1ProcessEvent(context.Background(), args, rply)
	sort.Strings(rply.AlteredFields)
	expErr := "SERVER_ERROR: NotEnoughParameters"
	if err == nil || err.Error() != expErr {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expErr, err)
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
		Tenant:    "cgrates.org",
		ID:        "ATTR_CHANGE_TENANT_FROM_USER",
		FilterIDs: []string{"*string:~*req.Account:dan@itsyscom.com|adrian@itsyscom.com"},
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
	err := dm.SetAttributeProfile(context.Background(), attr, true)
	if err != nil {
		t.Error(err)
	}

	attr2 := &AttributeProfile{
		Tenant: "adrian.itsyscom.com.co.uk",
		ID:     "ATTR_MATCH_TENANT",
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

	err = dm.SetAttributeProfile(context.Background(), attr2, true)
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
			APIOpts: map[string]interface{}{
				utils.OptsAttributesProcessRuns: 2,
			},
		},
	}
	rply := &AttrSProcessEventReply{}
	err = alS.V1ProcessEvent(context.Background(), args, rply)
	sort.Strings(rply.AlteredFields)
	expErr := "SERVER_ERROR: NotEnoughParameters"
	if err == nil || err.Error() != expErr {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expErr, err)
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
		Tenant:    "cgrates.org",
		ID:        "ATTR_CHANGE_TENANT_FROM_USER",
		FilterIDs: []string{"*string:~*req.Account:dan@itsyscom.com|adrian@itsyscom.com"},
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
	err := dm.SetAttributeProfile(context.Background(), attr, true)
	if err != nil {
		t.Error(err)
	}

	attr2 := &AttributeProfile{
		Tenant: "adrian.itsyscom.com.co.uk",
		ID:     "ATTR_MATCH_TENANT",
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

	err = dm.SetAttributeProfile(context.Background(), attr2, true)
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
			APIOpts: map[string]interface{}{
				utils.OptsAttributesProcessRuns: 2,
			},
		},
	}
	rply := &AttrSProcessEventReply{}
	err = alS.V1ProcessEvent(context.Background(), args, rply)
	sort.Strings(rply.AlteredFields)
	expErr := "SERVER_ERROR: invalid arguments <[{\"Rules\":\"CGRATES.ORG\"}]> to *valueExponent"
	if err == nil || err.Error() != expErr {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expErr, err)
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
		Attributes: []*Attribute{
			{
				Path:  "*req.RequestType",
				Type:  utils.MetaConstant,
				Value: postpaid,
			},
		},
		Weight: 20,
	}
	err = alS.dm.SetAttributeProfile(context.Background(), ap1, true)
	if err != nil {
		t.Error(err)
	}

	ap2 := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_2",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Attributes: []*Attribute{
			{
				Path:  "*req.RequestType",
				Type:  utils.MetaConstant,
				Value: postpaid,
			},
		},
		Weight: 10,
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

	if rcv, err := alS.attributeProfileForEvent(context.Background(), tnt, []string{"ATTR_3"}, evNm, lastID, make(map[string]int), 0); err == nil || err != utils.ErrNoDatabaseConn {
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

	if rcv, err := alS.attributeProfileForEvent(context.Background(), tnt, []string{"ATTR_3"}, evNm, lastID, make(map[string]int), 0); err == nil || err != utils.ErrNotFound {
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
		Attributes: []*Attribute{
			{
				Path:  "*req.RequestType",
				Type:  utils.MetaConstant,
				Value: postpaid,
			},
		},
		Weight: 20,
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

	if rcv, err := alS.attributeProfileForEvent(context.Background(), tnt, []string{"ATTR_1"}, evNm, lastID, make(map[string]int), 0); err == nil || err != utils.ErrWrongPath {
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
		Attributes: []*Attribute{
			{
				Path:  "*req.Password",
				Type:  utils.MetaConstant,
				Value: pw,
			},
		},
		Weight: 10,
	}
	err = alS.dm.SetAttributeProfile(context.Background(), ap1, true)
	if err != nil {
		t.Error(err)
	}

	ap2 := &AttributeProfile{
		Tenant: "cgrates.org",
		ID:     "ATTR2",
		Attributes: []*Attribute{
			{
				Path:  "*req.RequestType",
				Type:  utils.MetaConstant,
				Value: postpaid,
			},
		},
		Weight: 20,
	}
	err = alS.dm.SetAttributeProfile(context.Background(), ap2, true)
	if err != nil {
		t.Error(err)
	}

	args := &AttrArgsProcessEvent{
		AttributeIDs: []string{"ATTR1", "ATTR2"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "AttrProcessEventMultipleRuns",
			Event: map[string]interface{}{
				"Password": "passwd",
			},
			APIOpts: map[string]interface{}{
				utils.OptsAttributesProcessRuns: 4,
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
			APIOpts: map[string]interface{}{
				utils.OptsAttributesProcessRuns: 4,
			},
		},
	}

	if err := alS.V1ProcessEvent(context.Background(), args, reply); err != nil {
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
		Tenant: "cgrates.org",
		ID:     "ATTR1",
		Attributes: []*Attribute{
			{
				Path:  "*req.Password",
				Type:  utils.MetaConstant,
				Value: pw,
			},
		},
		Weight: 10,
	}
	err = alS.dm.SetAttributeProfile(context.Background(), ap1, true)
	if err != nil {
		t.Error(err)
	}

	ap2 := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR2",
		FilterIDs: []string{"*exists:~*vars.*processedProfileIDs[cgrates.org:ATTR1]:"},
		Attributes: []*Attribute{
			{
				Path:  "*req.RequestType",
				Type:  utils.MetaConstant,
				Value: postpaid,
			},
		},
		Weight: 20,
	}
	err = alS.dm.SetAttributeProfile(context.Background(), ap2, true)
	if err != nil {
		t.Error(err)
	}

	ap3 := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR3",
		FilterIDs: []string{"*exists:~*vars.*processedProfileIDs[cgrates.org:ATTR2]:"},
		Attributes: []*Attribute{
			{
				Path:  "*req.PaypalAccount",
				Type:  utils.MetaConstant,
				Value: paypal,
			},
		},
		Weight: 30,
	}
	err = alS.dm.SetAttributeProfile(context.Background(), ap3, true)
	if err != nil {
		t.Error(err)
	}

	args := &AttrArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "AttrProcessEventMultipleRuns",
			Event:  map[string]interface{}{},
			APIOpts: map[string]interface{}{
				utils.OptsAttributesProcessRuns: 3,
			},
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
			APIOpts: map[string]interface{}{
				utils.OptsAttributesProcessRuns: 3,
			},
		},
	}
	if err := alS.V1ProcessEvent(context.Background(), args, reply); err != nil {
		t.Error(err)
	} else {
		sort.Strings(reply.AlteredFields)
		if !reflect.DeepEqual(reply, exp) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(exp), utils.ToJSON(reply))
		}
	}
}
