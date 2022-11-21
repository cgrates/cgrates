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
	"github.com/cgrates/rpcclient"
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

func TestAttributesV1GetAttributeForEventNilCGREvent(t *testing.T) {
	alS := &AttributeService{}
	reply := &AttributeProfile{}

	experr := fmt.Sprintf("MANDATORY_IE_MISSING: [%s]", "CGREvent")
	err := alS.V1GetAttributeForEvent(nil, reply)

	if err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestAttributesV1GetAttributeForEventProfileNotFound(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, nil, nil)
	alS := &AttributeService{
		dm:      dm,
		filterS: &FilterS{},
		cgrcfg:  cfg,
	}
	args := &utils.CGREvent{}
	reply := &AttributeProfile{}

	experr := utils.ErrNotFound
	err := alS.V1GetAttributeForEvent(args, reply)

	if err == nil || err != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestAttributesV1GetAttributeForEvent2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, nil, nil)
	alS := &AttributeService{
		dm:      dm,
		filterS: &FilterS{},
		cgrcfg:  cfg,
	}
	args := &utils.CGREvent{}
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
	conMng := NewConnManager(cfg, make(map[string]chan rpcclient.ClientConnector))
	dm := NewDataManager(NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items), nil, conMng)
	filterS := NewFilterS(cfg, conMng, dm)
	if err := dm.SetAttributeProfile(&AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_CHANGE_TENANT_FROM_USER",
		Contexts:  []string{utils.MetaAny},
		FilterIDs: []string{"*string:~*req.Account:dan@itsyscom.com|adrian@itsyscom.com"},
		Attributes: []*Attribute{
			{
				Path:  "*tenant",
				Type:  "*variable",
				Value: config.NewRSRParsersMustCompile("~*req.Account:s/(.*)@(.*)/${1}.${2}/", utils.InfieldSep),
			},
			{
				Path:  "*req.Account",
				Type:  "*variable",
				Value: config.NewRSRParsersMustCompile("~*req.Account:s/(dan)@(.*)/${1}.${2}/:s/(adrian)@(.*)/andrei.${2}/", utils.InfieldSep),
			},
			{
				Path:  "*tenant",
				Type:  "*composed",
				Value: config.NewRSRParsersMustCompile(".co.uk", utils.InfieldSep),
			},
		},
		Blocker: false,
		Weight:  20,
	}, true); err != nil {
		t.Error(err)
	}

	if err := dm.SetAttributeProfile(&AttributeProfile{
		Tenant:   "adrian.itsyscom.com.co.uk",
		ID:       "ATTR_MATCH_TENANT",
		Contexts: []string{utils.MetaAny},
		Attributes: []*Attribute{
			{
				Path:  "*req.Password",
				Type:  utils.MetaConstant,
				Value: config.NewRSRParsersMustCompile("CGRATES.ORG", utils.InfieldSep),
			},
		},
		Blocker: false,
		Weight:  20,
	}, true); err != nil {
		t.Error(err)
	}

	alS := NewAttributeService(dm, filterS, cfg)
	var rply AttrSProcessEventReply
	expected := AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_CHANGE_TENANT_FROM_USER", "adrian.itsyscom.com.co.uk:ATTR_MATCH_TENANT"},
		AlteredFields:   []string{"*req.Account", "*req.Password", "*tenant"},
		CGREvent: &utils.CGREvent{
			Tenant: "adrian.itsyscom.com.co.uk",
			Time:   nil,
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
	if err = alS.V1ProcessEvent(&utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			utils.AccountField: "adrian@itsyscom.com",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAttributesProcessRuns: 2,
		},
	}, &rply); err != nil {
		t.Errorf("Expected <%+v>, received <%+v>", nil, err)
	} else if sort.Strings(rply.AlteredFields); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expected <%+v>, received <%+v>", utils.ToJSON(expected), utils.ToJSON(rply))
	}
}

func TestAttributesV1ProcessEventErrorMetaSum(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.FilterSCfg().ResourceSConns = []string{}
	conMng := NewConnManager(cfg, make(map[string]chan rpcclient.ClientConnector))
	dm := NewDataManager(NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items), nil, conMng)
	filterS := NewFilterS(cfg, conMng, dm)

	if err := dm.SetAttributeProfile(&AttributeProfile{
		Tenant:   "cgrates.org",
		ID:       "ATTR_MATCH_TENANT",
		Contexts: []string{utils.MetaAny},
		Attributes: []*Attribute{
			{
				Path:  "*req.Password",
				Type:  utils.MetaSum,
				Value: config.NewRSRParsersMustCompile("CGRATES.ORG", utils.InfieldSep),
			},
		},
		Blocker: false,
		Weight:  20,
	}, true); err != nil {
		t.Error(err)
	}

	alS := NewAttributeService(dm, filterS, cfg)
	var rply AttrSProcessEventReply
	expErr := "SERVER_ERROR: NotEnoughParameters"
	if err = alS.V1ProcessEvent(&utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			utils.AccountField: "adrian@itsyscom.com",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAttributesProcessRuns: 2,
		},
	}, &rply); err == nil || err.Error() != expErr {
		t.Errorf("Expected <%+v>, received <%+v>", expErr, err)
	}

}

func TestAttributesV1ProcessEventErrorMetaDifference(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.FilterSCfg().ResourceSConns = []string{}
	conMng := NewConnManager(cfg, make(map[string]chan rpcclient.ClientConnector))
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, nil, conMng)
	filterS := NewFilterS(cfg, conMng, dm)

	if err := dm.SetAttributeProfile(&AttributeProfile{
		Tenant:   "cgrates.org",
		ID:       "ATTR_MATCH_TENANT",
		Contexts: []string{utils.MetaAny},
		Attributes: []*Attribute{
			{
				Path:  "*req.Password",
				Type:  utils.MetaDifference,
				Value: config.NewRSRParsersMustCompile("CGRATES.ORG", utils.InfieldSep),
			},
		},
		Blocker: false,
		Weight:  20,
	}, true); err != nil {
		t.Error(err)
	}

	alS := NewAttributeService(dm, filterS, cfg)
	var rply AttrSProcessEventReply
	expErr := "SERVER_ERROR: NotEnoughParameters"
	if err := alS.V1ProcessEvent(&utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			utils.AccountField: "adrian@itsyscom.com",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAttributesProcessRuns: 2,
		},
	}, &rply); err == nil || err.Error() != expErr {
		t.Errorf("Expected <%+v>, received <%+v>", expErr, err)
	}

}

func TestAttributesV1ProcessEventErrorMetaValueExponent(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.FilterSCfg().ResourceSConns = []string{}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, nil, nil)
	filterS := NewFilterS(cfg, nil, dm)

	if err := dm.SetAttributeProfile(&AttributeProfile{
		Tenant:   "cgrates.org",
		ID:       "ATTR_MATCH_TENANT",
		Contexts: []string{utils.MetaAny},
		Attributes: []*Attribute{
			{
				Path:  "*req.Password",
				Type:  utils.MetaValueExponent,
				Value: config.NewRSRParsersMustCompile("CGRATES.ORG", utils.InfieldSep),
			},
		},
		Blocker: false,
		Weight:  20,
	}, true); err != nil {
		t.Error(err)
	}

	alS := NewAttributeService(dm, filterS, cfg)
	var rply AttrSProcessEventReply
	expErr := "SERVER_ERROR: invalid arguments <[{\"Rules\":\"CGRATES.ORG\"}]> to *value_exponent"
	if err := alS.V1ProcessEvent(&utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			utils.AccountField: "adrian@itsyscom.com",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAttributesProcessRuns: 2,
		},
	}, &rply); err == nil || err.Error() != expErr {
		t.Errorf("Expected <%+v>, received <%+v>", expErr, err)
	}

}

func TestAttributesattributeProfileForEventAnyCtxFalseNotFound(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	Cache.Clear(nil)
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
		lastID, make(map[string]int), 0, false); err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	} else if !reflect.DeepEqual(rcv, ap2) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", ap2, rcv)
	}

	lastID = "cgrates.org:ATTR_2"

	if rcv, err := alS.attributeProfileForEvent(tnt, ctx, nil, nil, evNm,
		lastID, make(map[string]int), 0, false); err == nil || err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	} else if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}
}

func TestAttributesattributeProfileForEventAnyCtxFalseFound(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	Cache.Clear(nil)
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
		lastID, make(map[string]int), 0, false); err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	} else if !reflect.DeepEqual(rcv, ap1) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", ap1, rcv)
	}
}

func TestAttributesattributeProfileForEventAnyCtxTrueBothFound(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	Cache.Clear(nil)
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
		lastID, make(map[string]int), 0, false); err != nil {
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
		lastID, make(map[string]int), 0, false); err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	} else if !reflect.DeepEqual(rcv, ap2) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", ap2, rcv)
	}
}

func TestAttributesattributeProfileForEventAnyCtxTrueErrMatching(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	Cache.Clear(nil)
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
		lastID, make(map[string]int), 0, false); err == nil || err != utils.ErrExists {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrExists, err)
	} else if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}
}

func TestAttributesattributeProfileForEventAnyCtxTrueNotFound(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	Cache.Clear(nil)
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
		lastID, make(map[string]int), 0, false); err == nil || err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	} else if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}
}

func TestAttributesattributeProfileForEventNoDBConn(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	Cache.Clear(nil)
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
		lastID, make(map[string]int), 0, false); err == nil || err != utils.ErrNoDatabaseConn {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNoDatabaseConn, err)
	} else if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}
}

func TestAttributesattributeProfileForEventErrNotFound(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	Cache.Clear(nil)
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
		lastID, make(map[string]int), 0, false); err == nil || err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	} else if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}
}

func TestAttributesattributeProfileForEventNotActive(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	Cache.Clear(nil)
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
		lastID, make(map[string]int), 0, false); err == nil || err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	} else if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}
}

func TestAttributesattributeProfileForEventErrPass(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	Cache.Clear(nil)
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
		lastID, make(map[string]int), 0, false); err == nil || err != utils.ErrWrongPath {
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
	if val, err := ParseAttribute(dp, utils.MetaSIPCID, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.cid;~*req.extra;~*req.to;~*req.from", utils.
		InfieldSep), 0, utils.EmptyString, utils.EmptyString, utils.InfieldSep); err != utils.ErrNotFound {
		t.Errorf("Expected <%+v>, received <%+v>", utils.ErrNotFound, err)
	} else if val != "12345" {
		t.Errorf("received  %+v", val)
	}

	if val, err := ParseAttribute(dp, utils.MetaNone, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.cid;~*req.extra;~*req.to;~*req.from", utils.
		InfieldSep), 0, utils.EmptyString, utils.EmptyString, utils.InfieldSep); err != nil {
		t.Errorf("received <%+v>", err)
	} else if val != nil {
		t.Errorf("received  %+v", val)
	}

	if _, err := ParseAttribute(dp, utils.MetaUsageDifference, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.cid;~*req.extra;~*req.to;~*req.from", utils.
		InfieldSep), 0, utils.EmptyString, utils.EmptyString, utils.InfieldSep); err == nil {
		t.Errorf(",received <%+v>", err)
	}

	if _, err = ParseAttribute(utils.MapStorage{}, utils.MetaSum, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.cid;~*req.extra;~*req.to;~*req.from", utils.
		InfieldSep), 0, utils.EmptyString, utils.EmptyString, utils.InfieldSep); err == nil || err != utils.ErrNotFound {
		t.Errorf("expected <%+v>received <%+v>", utils.ErrNotFound, err)
	}

	if _, err = ParseAttribute(utils.MapStorage{}, utils.MetaDifference, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.cid;~*req.extra;~*req.to;~*req.from", utils.
		InfieldSep), 0, utils.EmptyString, utils.EmptyString, utils.InfieldSep); err == nil || err != utils.ErrNotFound {
		t.Errorf("expected <%+v>received <%+v>", utils.ErrNotFound, err)
	} else if _, err = ParseAttribute(utils.MapStorage{}, utils.MetaMultiply, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.cid;~*req.extra;~*req.to;~*req.from", utils.
		InfieldSep), 0, utils.EmptyString, utils.EmptyString, utils.InfieldSep); err == nil || err != utils.ErrNotFound {
		t.Errorf("expected <%+v>received <%+v>", utils.ErrNotFound, err)
	} else if _, err = ParseAttribute(utils.MapStorage{}, utils.MetaDivide, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.cid;~*req.extra;~*req.to;~*req.from", utils.
		InfieldSep), 0, utils.EmptyString, utils.EmptyString, utils.InfieldSep); err == nil || err != utils.ErrNotFound {
		t.Errorf("expected <%+v>received <%+v>", utils.ErrNotFound, err)
	} else if _, err = ParseAttribute(utils.MapStorage{}, utils.MetaUnixTimestamp, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.cid;~*req.extra;~*req.to;~*req.from", utils.
		InfieldSep), 0, utils.EmptyString, utils.EmptyString, utils.InfieldSep); err == nil || err != utils.ErrNotFound {
		t.Errorf("expected <%+v>received <%+v>", utils.ErrNotFound, err)
	} else if _, err = ParseAttribute(utils.MapStorage{}, utils.MetaDateTime, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.cid;~*req.extra;~*req.to;~*req.from", utils.
		InfieldSep), 0, utils.EmptyString, utils.EmptyString, utils.InfieldSep); err == nil || err != utils.ErrNotFound {
		t.Errorf("expected <%+v>received <%+v>", utils.ErrNotFound, err)
	} else if _, err = ParseAttribute(utils.MapStorage{}, utils.MetaPrefix, "`val", config.NewRSRParsersMustCompile("~*req.cid;~*req.extra;~*req.to;~*req.from", utils.
		InfieldSep), 0, utils.EmptyString, utils.EmptyString, utils.InfieldSep); err == nil || err.Error() != "Unclosed unspilit syntax" {
		t.Errorf("expected <%+v>received <%+v>", "Unclosed unspilit syntax", err)
	} else if _, err = ParseAttribute(utils.MapStorage{}, utils.MetaSuffix, "`val", config.NewRSRParsersMustCompile("~*req.cid;~*req.extra;~*req.to;~*req.from", utils.
		InfieldSep), 0, utils.EmptyString, utils.EmptyString, utils.InfieldSep); err == nil || err.Error() != "Unclosed unspilit syntax" {
		t.Errorf("expected <%+v>received <%+v>", "Unclosed unspilit syntax", err)
	} else if _, err = ParseAttribute(utils.MapStorage{}, utils.MetaCCUsage, "`val", config.NewRSRParsersMustCompile("~*req.cid;~*req.extra;~*req.to;~*req.from", utils.
		InfieldSep), 0, utils.EmptyString, utils.EmptyString, utils.InfieldSep); err == nil {
		t.Errorf("received <%+v>", err)
	} else if _, err = ParseAttribute(utils.MapStorage{}, "default", utils.EmptyString, config.NewRSRParsersMustCompile("~*req.cid;~*req.extra;~*req.to;~*req.from", utils.
		InfieldSep), 0, utils.EmptyString, utils.EmptyString, utils.InfieldSep); err == nil {
		t.Errorf("received <%+v>", err)
	} else if _, err = ParseAttribute(utils.MapStorage{}, "default", utils.EmptyString, config.NewRSRParsersMustCompile("~*req.cid;~*req.extra", utils.
		InfieldSep), 0, utils.EmptyString, utils.EmptyString, utils.InfieldSep); err == nil {
		t.Errorf("received <%+v>", err)
	} else if _, err = ParseAttribute(utils.MapStorage{utils.MetaReq: utils.MapStorage{"cid": "12345"}}, utils.MetaUsageDifference, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.cid;~*req.extra", utils.InfieldSep), 0, utils.EmptyString, utils.EmptyString, utils.InfieldSep); err == nil || err != utils.ErrNotFound {
		t.Errorf("expected <%+v>received <%+v>", utils.ErrNotFound, err)
	} else if _, err = ParseAttribute(utils.MapStorage{utils.MetaReq: utils.MapStorage{"cid": "12345", "extra": "1003"}}, utils.MetaUsageDifference, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.cid;~*req.extra", utils.
		InfieldSep), 0, utils.EmptyString, utils.EmptyString, utils.InfieldSep); err == nil || err.Error() != "Unsupported time format" {
		t.Errorf("expected <%+v>received <%+v>", "Unsupported time format", err)
	} else if _, err = ParseAttribute(utils.MapStorage{utils.MetaReq: utils.MapStorage{"cid": "0", "extra": "1003"}}, utils.MetaUsageDifference, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.cid;~*req.extra", utils.
		InfieldSep), 0, utils.EmptyString, utils.EmptyString, utils.InfieldSep); err == nil || err.Error() != "Unsupported time format" {
		t.Errorf("expected <%+v>received <%+v>", "Unsupported time format", err)
	} else if _, err = ParseAttribute(utils.MapStorage{utils.MetaReq: utils.MapStorage{}}, utils.MetaUsageDifference, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.cid;~*req.extra", utils.
		InfieldSep), 0, utils.EmptyString, utils.EmptyString, utils.InfieldSep); err == nil || err != utils.ErrNotFound {
		t.Errorf("expected <%+v>received <%+v>", utils.ErrNotFound, err)
	} else if _, err = ParseAttribute(utils.MapStorage{utils.MetaReq: utils.MapStorage{"extra": "1003", "to": "1001"}}, utils.MetaCCUsage, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.cid;~*req.extra;~*req.to", utils.
		InfieldSep), 0, utils.EmptyString, utils.EmptyString, utils.InfieldSep); err == nil || err != utils.ErrNotFound {
		t.Errorf("expected <%+v>received <%+v>", utils.ErrNotFound, err)
	} else if _, err = ParseAttribute(utils.MapStorage{utils.MetaReq: utils.MapStorage{"cid": "val", "extra": "1003", "to": "1001"}}, utils.MetaCCUsage, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.cid;~*req.extra;~*req.to", utils.
		InfieldSep), 0, utils.EmptyString, utils.EmptyString, utils.InfieldSep); err == nil {
		t.Errorf("received <%+v>", err)
	} else if _, err = ParseAttribute(utils.MapStorage{utils.MetaReq: utils.MapStorage{"cid": "12233", "to": "1001"}}, utils.MetaCCUsage, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.cid;~*req.extra;~*req.to", utils.
		InfieldSep), 0, utils.EmptyString, utils.EmptyString, utils.InfieldSep); err == nil || err != utils.ErrNotFound {
		t.Errorf("expected <%+v>received <%+v>", utils.ErrNotFound, err)
	} else if _, err = ParseAttribute(utils.MapStorage{utils.MetaReq: utils.MapStorage{"cid": "12232", "extra": "val", "to": "1001"}}, utils.MetaCCUsage, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.cid;~*req.extra;~*req.to", utils.
		InfieldSep), 0, utils.EmptyString, utils.EmptyString, utils.InfieldSep); err == nil {
		t.Errorf("received <%+v>", err)
	} else if _, err = ParseAttribute(utils.MapStorage{utils.MetaReq: utils.MapStorage{"cid": "12233", "extra": "1001"}}, utils.MetaCCUsage, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.cid;~*req.extra;~*req.to", utils.
		InfieldSep), 0, utils.EmptyString, utils.EmptyString, utils.InfieldSep); err == nil || err != utils.ErrNotFound {
		t.Errorf("expected <%+v>received <%+v>", utils.ErrNotFound, err)
	} else if _, err = ParseAttribute(utils.MapStorage{utils.MetaReq: utils.MapStorage{"cid": "12232", "extra": "1001", "to": "val"}}, utils.MetaCCUsage, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.cid;~*req.extra;~*req.to", utils.
		InfieldSep), 0, utils.EmptyString, utils.EmptyString, utils.InfieldSep); err == nil {
		t.Errorf("received <%+v>", err)
	}
	expDur := 100000 * time.Nanosecond
	if val, err := ParseAttribute(utils.MapStorage{utils.MetaReq: utils.MapStorage{"cid": "1000", "extra": "100", "to": "100"}}, utils.MetaCCUsage, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.cid;~*req.extra;~*req.to", utils.
		InfieldSep), 0, utils.EmptyString, utils.EmptyString, utils.InfieldSep); err != nil {
		t.Errorf("received <%+v>", err)
	} else if val != expDur {
		t.Errorf("expected %v,received %v", expDur, val)
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
	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().IndexedSelects = false
	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	Cache.Clear(nil)
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

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "AttrProcessEventMultipleRuns",
		Event: map[string]interface{}{
			"Password": "passwd",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAttributesProcessRuns: 3,
			utils.OptsContext:               utils.MetaAny,
			utils.OptsAttributesProfileIDs:  []string{"ATTR1", "ATTR2"},
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
				utils.OptsAttributesProcessRuns: 3,
				utils.OptsContext:               utils.MetaAny,
				utils.OptsAttributesProfileIDs:  []string{"ATTR1", "ATTR2"},
			},
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

	if err := alS.V1ProcessEvent(nil, reply); err == nil {
		t.Error("expected error")
	}
}

func TestAttributesV1ProcessEventMultipleRuns2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().IndexedSelects = false
	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	Cache.Clear(nil)
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

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "AttrProcessEventMultipleRuns",
		Event:  map[string]interface{}{},
		APIOpts: map[string]interface{}{
			utils.OptsAttributesProcessRuns: 3,
			utils.OptsContext:               utils.MetaAny,
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
				utils.OptsContext:               utils.MetaAny,
			},
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

func TestArgeesUnmarshalJSON(t *testing.T) {
	cgr := &CGREventWithEeIDs{
		EeIDs: []string{"eeID1", "eeID2", "eeID3", "eeID$"},
		clnb:  true,
		CGREvent: &utils.CGREvent{
			Event: map[string]interface{}{
				utils.CostDetails: "22",
			},
		},
	}
	if err := cgr.UnmarshalJSON([]byte("val")); err == nil {
		t.Error(err)
	} else if err = cgr.UnmarshalJSON([]byte(`{
		"EeIDs":["eeid1","eeid2"],
		"Event":{
				"CostDetails":{
				 "CGRID":"id1"	}
			 }
	  }`)); err != nil {
		t.Error(err)
	}
}
func TestArgeesRPCClone(t *testing.T) {

	attr := &CGREventWithEeIDs{
		EeIDs: []string{"eeid1", "eeid2"},
		CGREvent: &utils.CGREvent{
			Tenant:  "cgrates.org",
			ID:      "id",
			Time:    &time.Time{},
			Event:   map[string]interface{}{},
			APIOpts: map[string]interface{}{},
		},
		clnb: false,
	}
	if val, err := attr.RPCClone(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(val, attr) {
		t.Errorf("expected %v,received %v", utils.ToJSON(attr), utils.ToJSON(val))
	}
	attr.clnb = true
	exp := &CGREventWithEeIDs{
		EeIDs: []string{"eeid1", "eeid2"},
		CGREvent: &utils.CGREvent{
			Tenant:  "cgrates.org",
			ID:      "id",
			Time:    &time.Time{},
			Event:   map[string]interface{}{},
			APIOpts: map[string]interface{}{},
		},
	}

	if val, err := attr.RPCClone(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(val, exp) {
		t.Errorf("expected %v,received %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
}
