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
		MatchedProfiles: []string{"ATTR_CHANGE_TENANT_FROM_USER", "ATTR_MATCH_TENANT"},
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

func TestAttributesattributeProfileForEvent(t *testing.T) {
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
	}
	lastID := ""
	alS.cgrcfg.AttributeSCfg().AnyContext = false

	if rcv, err := alS.attributeProfileForEvent(tnt, ctx, nil, nil, evNm,
		lastID); err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	} else if !reflect.DeepEqual(rcv, ap2) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", ap2, rcv)
	}

	lastID = "ATTR_2"

	if rcv, err := alS.attributeProfileForEvent(tnt, ctx, nil, nil, evNm,
		lastID); err == nil || err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	} else if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}

	Cache.Clear(nil)
	ap1.FilterIDs = []string{"*string:~*req.Account:1001"}
	err = alS.dm.SetAttributeProfile(ap1, true)
	if err != nil {
		t.Error(err)
	}
	err = alS.dm.SetAttributeProfile(ap2, true)
	if err != nil {
		t.Error(err)
	}

	if rcv, err := alS.attributeProfileForEvent(tnt, ctx, nil, nil, evNm,
		lastID); err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	} else if !reflect.DeepEqual(rcv, ap1) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", ap1, rcv)
	}

	Cache.Clear(nil)
	alS.cgrcfg.AttributeSCfg().AnyContext = true
	err = alS.dm.SetAttributeProfile(ap1, true)
	if err != nil {
		t.Error(err)
	}
	err = alS.dm.SetAttributeProfile(ap2, true)
	if err != nil {
		t.Error(err)
	}

	if rcv, err := alS.attributeProfileForEvent(tnt, ctx, nil, nil, evNm,
		lastID); err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	} else if !reflect.DeepEqual(rcv, ap1) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", ap1, rcv)
	}

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

	alS.cgrcfg.AttributeSCfg().IndexedSelects = true
	alS.dm = nil

	if rcv, err := alS.attributeProfileForEvent(tnt, ctx, []string{"ATTR_3"}, nil, evNm,
		lastID); err == nil || err != utils.ErrNoDatabaseConn {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNoDatabaseConn, err)
	} else if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}

	apNil := &AttributeProfile{}
	tnt = ""
	alS.dm = dm
	err = alS.dm.SetAttributeProfile(apNil, true)
	if err != nil {
		t.Error(err)
	}
	if rcv, err := alS.attributeProfileForEvent(tnt, ctx, []string{"ATTR_3"}, nil, evNm,
		lastID); err == nil || err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	} else if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}

	Cache.Clear(nil)
	ap1.ActivationInterval = &utils.ActivationInterval{
		ExpiryTime: time.Date(2021, 5, 14, 15, 0, 0, 0, time.UTC),
	}
	err = alS.dm.SetAttributeProfile(ap1, true)
	if err != nil {
		t.Error(err)
	}
	actTime := utils.TimePointer(time.Date(2021, 5, 14, 16, 0, 0, 0, time.UTC))
	tnt = "cgrates.org"
	if rcv, err := alS.attributeProfileForEvent(tnt, ctx, nil, actTime, evNm,
		lastID); err == nil || err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	} else if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}

	Cache.Clear(nil)
	ap1.ActivationInterval = nil
	err = alS.dm.SetAttributeProfile(ap1, true)
	if err != nil {
		t.Error(err)
	}
	evNm = utils.MapStorage{
		utils.MetaReq: 1,
	}

	if rcv, err := alS.attributeProfileForEvent(tnt, ctx, []string{"ATTR_1"}, nil, evNm,
		lastID); err == nil || err != utils.ErrWrongPath {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrWrongPath, err)
	} else if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}
}
