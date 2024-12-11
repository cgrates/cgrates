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
package cdrs

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func TestCDRsNewCDRServer(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()

	dm := &engine.DataManager{}
	fltrs := &engine.FilterS{}
	connMng := &engine.ConnManager{}
	storDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	newCDRSrv := NewCDRServer(cfg, dm, fltrs, connMng, storDB)
	expected := &CDRServer{
		cfg:     cfg,
		dm:      dm,
		guard:   guardian.Guardian,
		fltrS:   fltrs,
		connMgr: connMng,
		db:      storDB,
	}
	if !reflect.DeepEqual(newCDRSrv, expected) {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", expected, newCDRSrv)
	}
}

func TestCDRsChrgrSProcessEventErrMsnConnIDs(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()

	dm := &engine.DataManager{}
	fltrs := &engine.FilterS{}
	connMng := &engine.ConnManager{}
	storDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	newCDRSrv := NewCDRServer(cfg, dm, fltrs, connMng, storDB)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
		APIOpts: map[string]any{
			utils.MetaSubsys:   utils.MetaChargers,
			utils.MetaOriginID: "originID",
		},
	}
	_, err := newCDRSrv.chrgrSProcessEvent(context.Background(), cgrEv)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [connIDs]" {
		t.Errorf("\nExpected <MANDATORY_IE_MISSING: [connIDs]> \n, received <%+v>", err)
	}
}

func TestCDRsAttrSProcessEventNoOpts(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dm := &engine.DataManager{}
	fltrs := &engine.FilterS{}
	connMng := &engine.ConnManager{}
	storDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	newCDRSrv := NewCDRServer(cfg, dm, fltrs, connMng, storDB)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
	}
	err := newCDRSrv.attrSProcessEvent(context.Background(), cgrEv)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [connIDs]" {
		t.Errorf("\nExpected <MANDATORY_IE_MISSING: [connIDs]> \n, received <%+v>", err)
	}
}

func TestCDRsAttrSProcessEvent(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()

	dm := &engine.DataManager{}
	fltrs := &engine.FilterS{}
	connMng := &engine.ConnManager{}
	storDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	newCDRSrv := NewCDRServer(cfg, dm, fltrs, connMng, storDB)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
		APIOpts: map[string]any{
			utils.MetaSubsys: utils.MetaChargers,
		},
	}
	err := newCDRSrv.attrSProcessEvent(context.Background(), cgrEv)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [connIDs]" {
		t.Errorf("\nExpected <MANDATORY_IE_MISSING: [connIDs]> \n, received <%+v>", err)
	}
}

func TestCDRsRateSCostForEventErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()

	dm := &engine.DataManager{}
	fltrs := &engine.FilterS{}
	connMng := &engine.ConnManager{}
	storDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	newCDRSrv := NewCDRServer(cfg, dm, fltrs, connMng, storDB)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
		APIOpts: map[string]any{
			utils.MetaSubsys: utils.MetaChargers,
		},
	}
	err := newCDRSrv.rateSCostForEvent(context.Background(), cgrEv)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [connIDs]" {
		t.Errorf("\nExpected <MANDATORY_IE_MISSING: [connIDs]> \n, received <%+v>", err)
	}
}

func TestCDRsAccountSDebitEventErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()

	dm := &engine.DataManager{}
	fltrs := &engine.FilterS{}
	connMng := &engine.ConnManager{}
	storDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	newCDRSrv := NewCDRServer(cfg, dm, fltrs, connMng, storDB)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
		APIOpts: map[string]any{
			utils.MetaSubsys: utils.MetaChargers,
		},
	}
	err := newCDRSrv.accountSDebitEvent(context.Background(), cgrEv)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [connIDs]" {
		t.Errorf("\nExpected <MANDATORY_IE_MISSING: [connIDs]> \n, received <%+v>", err)
	}
}

func TestCDRsThdSProcessEventErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()

	dm := &engine.DataManager{}
	fltrs := &engine.FilterS{}
	connMng := &engine.ConnManager{}
	storDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	newCDRSrv := NewCDRServer(cfg, dm, fltrs, connMng, storDB)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
	}
	err := newCDRSrv.thdSProcessEvent(context.Background(), cgrEv)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [connIDs]" {
		t.Errorf("\nExpected <MANDATORY_IE_MISSING: [connIDs]> \n, received <%+v>", err)
	}

}

func TestCDRsStatSProcessEventErrMsnConnIDs(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()

	dm := &engine.DataManager{}
	fltrs := &engine.FilterS{}
	connMng := &engine.ConnManager{}
	storDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	newCDRSrv := NewCDRServer(cfg, dm, fltrs, connMng, storDB)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
		APIOpts: map[string]any{
			utils.MetaSubsys: utils.MetaChargers,
		},
	}
	err := newCDRSrv.statSProcessEvent(context.Background(), cgrEv)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [connIDs]" {
		t.Errorf("\nExpected <MANDATORY_IE_MISSING: [connIDs]> \n, received <%+v>", err)
	}

}

func TestCDRsEESProcessEventErrMsnConnIDs(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()

	dm := &engine.DataManager{}
	fltrs := &engine.FilterS{}
	connMng := &engine.ConnManager{}
	storDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	newCDRSrv := NewCDRServer(cfg, dm, fltrs, connMng, storDB)

	cgrEv := &utils.CGREventWithEeIDs{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testID",
			Event: map[string]any{
				"Resources":      "ResourceProfile1",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				"PddInterval":    "1s",
				utils.Weight:     "20.0",
				utils.Usage:      135 * time.Second,
				utils.Cost:       123.0,
			},
			APIOpts: map[string]any{
				utils.MetaSubsys: utils.MetaChargers,
			},
		},
	}
	err := newCDRSrv.eeSProcessEvent(context.Background(), cgrEv)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [connIDs]" {
		t.Errorf("\nExpected <MANDATORY_IE_MISSING: [connIDs]> \n, received <%+v>", err)
	}

}

func TestCDRsNewMapEventFromReqForm(t *testing.T) {
	httpReq := &http.Request{
		Form: url.Values{
			"value1": {"value2"},
		},
	}
	result, err := newMapEventFromReqForm(httpReq)
	if err != nil {
		t.Errorf("\nExpected <nil> \n, received <%+v>", err)
	}
	expected := engine.MapEvent{
		"value1": "value2",
		"Source": "",
	}
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", expected, result)
	}
}

type ccMock struct {
	calls map[string]func(ctx *context.Context, args any, reply any) error
}

func (ccM *ccMock) Call(ctx *context.Context, serviceMethod string, args any, reply any) (err error) {
	if call, has := ccM.calls[serviceMethod]; !has {
		return rpcclient.ErrUnsupporteServiceMethod
	} else {
		return call(ctx, args, reply)
	}
}

func TestCDRsAttrSProcessEventMock(t *testing.T) {
	testCache := engine.Cache
	tmpC := config.CgrConfig()
	defer func() {
		engine.Cache = testCache
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().AttributeSConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaAttributes)}

	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMng := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	storDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	newCDRSrv := NewCDRServer(cfg, dm, fltrs, connMng, storDB)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.AttributeSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				*reply.(*engine.AttrSProcessEventReply) = engine.AttrSProcessEventReply{
					AlteredFields: []*engine.FieldsAltered{},
				}
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaAttributes), utils.AttributeSv1, rpcInternal)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
		APIOpts: map[string]any{
			utils.MetaSubsys: utils.MetaAttributes,
		},
	}
	err := newCDRSrv.attrSProcessEvent(context.Background(), cgrEv)
	if err != nil {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", nil, err)
	}
	expected := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaCDRs,
			utils.MetaSubsys:  utils.MetaCDRs,
		},
	}
	if !reflect.DeepEqual(expected, cgrEv) {
		t.Errorf("\nExpected <%+v> \n,received <%+v>", expected, cgrEv)
	}
}

func TestCDRsAttrSProcessEventMockNotFoundErr(t *testing.T) {
	testCache := engine.Cache
	tmpC := config.CgrConfig()
	defer func() {
		engine.Cache = testCache
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().AttributeSConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaAttributes)}

	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMng := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	storDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	newCDRSrv := NewCDRServer(cfg, dm, fltrs, connMng, storDB)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.AttributeSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				*reply.(*engine.AttrSProcessEventReply) = engine.AttrSProcessEventReply{
					AlteredFields: []*engine.FieldsAltered{{
						Fields: []string{},
					}},
				}
				return utils.ErrNotFound
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaAttributes), utils.AttributeSv1, rpcInternal)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
		APIOpts: map[string]any{
			utils.MetaSubsys: utils.MetaAttributes,
		},
	}
	err := newCDRSrv.attrSProcessEvent(context.Background(), cgrEv)
	if err != nil {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", nil, err)
	}
	expected := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaCDRs,
			utils.MetaSubsys:  utils.MetaCDRs,
		},
	}
	if !reflect.DeepEqual(expected, cgrEv) {
		t.Errorf("\nExpected <%+v> \n,received <%+v>", expected, cgrEv)
	}
}

func TestCDRsAttrSProcessEventMockNotEmptyAF(t *testing.T) {
	testCache := engine.Cache
	tmpC := config.CgrConfig()
	defer func() {
		engine.Cache = testCache
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().AttributeSConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaAttributes)}

	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMng := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	storDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	newCDRSrv := NewCDRServer(cfg, dm, fltrs, connMng, storDB)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.AttributeSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				*reply.(*engine.AttrSProcessEventReply) = engine.AttrSProcessEventReply{
					AlteredFields: []*engine.FieldsAltered{{
						Fields: []string{utils.AccountField},
					}},
					CGREvent: &utils.CGREvent{
						Tenant: "cgrates.org",
						ID:     "testID",
						Event: map[string]any{
							utils.AccountField: "1001",
							"Resources":        "ResourceProfile1",
							utils.AnswerTime:   time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
							"UsageInterval":    "1s",
							"PddInterval":      "1s",
							utils.Weight:       "20.0",
							utils.Usage:        135 * time.Second,
							utils.Cost:         123.0,
						},
						APIOpts: map[string]any{
							utils.MetaSubsys:   utils.MetaAttributes,
							utils.AccountField: "1001",
						},
					},
				}
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaAttributes), utils.AttributeSv1, rpcInternal)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			utils.AccountField: "1001",
			"Resources":        "ResourceProfile1",
			utils.AnswerTime:   time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":    "1s",
			"PddInterval":      "1s",
			utils.Weight:       "20.0",
			utils.Usage:        135 * time.Second,
			utils.Cost:         123.0,
		},
		APIOpts: map[string]any{
			utils.MetaSubsys: utils.MetaAttributes,
		},
	}
	err := newCDRSrv.attrSProcessEvent(context.Background(), cgrEv)
	if err != nil {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", nil, err)
	}
	expected := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			utils.AccountField: "1001",
			"Resources":        "ResourceProfile1",
			utils.AnswerTime:   time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":    "1s",
			"PddInterval":      "1s",
			utils.Weight:       "20.0",
			utils.Usage:        135 * time.Second,
			utils.Cost:         123.0,
		},
		APIOpts: map[string]any{
			utils.AccountField: "1001",
			utils.MetaSubsys:   utils.MetaAttributes,
		},
	}
	if !reflect.DeepEqual(expected, cgrEv) {
		t.Errorf("\nExpected <%+v> \n,received <%+v>", expected, cgrEv)
	}
}

func TestCDRsChrgrSProcessEvent(t *testing.T) {
	testCache := engine.Cache
	tmpC := config.CgrConfig()
	defer func() {
		engine.Cache = testCache
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaChargers)}

	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMng := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	storDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	newCDRSrv := NewCDRServer(cfg, dm, fltrs, connMng, storDB)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ChargerSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				*reply.(*[]*engine.ChrgSProcessEventReply) = []*engine.ChrgSProcessEventReply{
					{
						ChargerSProfile: "string",
					},
				}
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaChargers), utils.ChargerSv1, rpcInternal)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
		APIOpts: map[string]any{
			utils.MetaSubsys: utils.MetaChargers,
		},
	}
	result, err := newCDRSrv.chrgrSProcessEvent(context.Background(), cgrEv)
	if err != nil {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", nil, err)
	}
	var expecte *utils.CGREvent
	expected := []*utils.CGREvent{
		expecte,
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", expected, result)
	}

}

func TestCDRsRateProcessEventMock(t *testing.T) {
	testCache := engine.Cache
	tmpC := config.CgrConfig()
	defer func() {
		engine.Cache = testCache
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().RateSConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaRates)}

	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMng := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	storDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	newCDRSrv := NewCDRServer(cfg, dm, fltrs, connMng, storDB)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.RateSv1CostForEvent: func(ctx *context.Context, args, reply any) error {
				*reply.(*utils.RateProfileCost) = utils.RateProfileCost{}
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaRates), utils.RateSv1, rpcInternal)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
		APIOpts: map[string]any{
			utils.MetaSubsys: utils.MetaRates,
		},
	}
	err := newCDRSrv.rateSCostForEvent(context.Background(), cgrEv)
	if err != nil {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", nil, err)
	}
	expected := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{

			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
		APIOpts: map[string]any{
			utils.MetaRateSCost: utils.RateProfileCost{},
			utils.MetaSubsys:    utils.MetaRates,
		},
	}
	if !reflect.DeepEqual(expected, cgrEv) {
		t.Errorf("\nExpected <%+v> \n,received <%+v>", expected, cgrEv)
	}
}

func TestCDRsAccountProcessEventMock(t *testing.T) {
	testCache := engine.Cache
	tmpC := config.CgrConfig()
	defer func() {
		engine.Cache = testCache
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().AccountSConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaAccounts)}

	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMng := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	storDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	newCDRSrv := NewCDRServer(cfg, dm, fltrs, connMng, storDB)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.AccountSv1DebitAbstracts: func(ctx *context.Context, args, reply any) error {
				*reply.(*utils.EventCharges) = utils.EventCharges{}
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaAccounts), utils.AccountSv1, rpcInternal)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{

			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
		APIOpts: map[string]any{
			utils.MetaAccountSCost: &utils.EventCharges{},
			utils.MetaSubsys:       utils.MetaAccounts,
		},
	}
	err := newCDRSrv.accountSDebitEvent(context.Background(), cgrEv)
	if err != nil {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", nil, err)
	}
	expected := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{

			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
		APIOpts: map[string]any{
			utils.MetaAccountSCost: cgrEv.APIOpts[utils.MetaAccountSCost],
			utils.MetaSubsys:       utils.MetaAccounts,
		},
	}
	if !reflect.DeepEqual(expected, cgrEv) {
		t.Errorf("\nExpected <%+v> \n,received <%+v>", expected, cgrEv)
	}
}

func TestCDRsThdSProcessEventMock(t *testing.T) {
	testCache := engine.Cache
	tmpC := config.CgrConfig()
	defer func() {
		engine.Cache = testCache
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().ThresholdSConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaThresholds)}

	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMng := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	storDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	newCDRSrv := NewCDRServer(cfg, dm, fltrs, connMng, storDB)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ThresholdSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				*reply.(*[]string) = []string{"testID"}
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaThresholds), utils.ThresholdSv1, rpcInternal)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
		APIOpts: nil,
	}
	err := newCDRSrv.thdSProcessEvent(context.Background(), cgrEv)
	if err != nil {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", nil, err)
	}
	expected := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
		APIOpts: nil,
	}
	if !reflect.DeepEqual(expected, cgrEv) {
		t.Errorf("\nExpected <%+v> \n,received <%+v>", expected, cgrEv)
	}
}

func TestCDRsThdSProcessEventMockNotfound(t *testing.T) {
	testCache := engine.Cache
	tmpC := config.CgrConfig()
	defer func() {
		engine.Cache = testCache
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().ThresholdSConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaThresholds)}

	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMng := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	storDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	newCDRSrv := NewCDRServer(cfg, dm, fltrs, connMng, storDB)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ThresholdSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				return utils.ErrNotFound
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaThresholds), utils.ThresholdSv1, rpcInternal)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
		APIOpts: nil,
	}
	err := newCDRSrv.thdSProcessEvent(context.Background(), cgrEv)
	if err != nil {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", nil, err)
	}
	expected := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
		APIOpts: nil,
	}
	if !reflect.DeepEqual(expected, cgrEv) {
		t.Errorf("\nExpected <%+v> \n,received <%+v>", expected, cgrEv)
	}
}

func TestCDRsStatSProcessEventMock(t *testing.T) {
	testCache := engine.Cache
	tmpC := config.CgrConfig()
	defer func() {
		engine.Cache = testCache
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().StatSConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaStats)}

	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMng := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	storDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	newCDRSrv := NewCDRServer(cfg, dm, fltrs, connMng, storDB)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.StatSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				*reply.(*[]string) = []string{"testID"}
				return utils.ErrNotFound
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaStats), utils.ThresholdSv1, rpcInternal)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
		APIOpts: nil,
	}
	err := newCDRSrv.statSProcessEvent(context.Background(), cgrEv)
	if err != nil {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", nil, err)
	}
	expected := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
		APIOpts: nil,
	}
	if !reflect.DeepEqual(expected, cgrEv) {
		t.Errorf("\nExpected <%+v> \n,received <%+v>", expected, cgrEv)
	}
}

func TestCDRsEESProcessEventMock(t *testing.T) {
	testCache := engine.Cache
	tmpC := config.CgrConfig()
	defer func() {
		engine.Cache = testCache
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs)}

	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMng := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	storDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	newCDRSrv := NewCDRServer(cfg, dm, fltrs, connMng, storDB)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				*reply.(*map[string]map[string]any) = map[string]map[string]any{}
				return utils.ErrNotFound
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs), utils.EeSv1, rpcInternal)

	cgrEv := &utils.CGREventWithEeIDs{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testID",
			Event: map[string]any{
				"Resources":      "ResourceProfile1",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				"PddInterval":    "1s",
				utils.Weight:     "20.0",
				utils.Usage:      135 * time.Second,
				utils.Cost:       123.0,
			},
			APIOpts: nil,
		},
	}
	err := newCDRSrv.eeSProcessEvent(context.Background(), cgrEv)
	if err != nil {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", nil, err)
	}
	expected := &utils.CGREventWithEeIDs{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testID",
			Event: map[string]any{
				"Resources":      "ResourceProfile1",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				"PddInterval":    "1s",
				utils.Weight:     "20.0",
				utils.Usage:      135 * time.Second,
				utils.Cost:       123.0,
			},
			APIOpts: nil,
		},
	}
	if !reflect.DeepEqual(expected, cgrEv) {
		t.Errorf("\nExpected <%+v> \n,received <%+v>", expected, cgrEv)
	}
}

func TestCDRsProcessEventMock(t *testing.T) {
	testCache := engine.Cache
	tmpC := config.CgrConfig()
	defer func() {
		engine.Cache = testCache
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs)}

	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMng := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	storDB := engine.NewInternalDB(nil, nil, nil)
	newCDRSrv := NewCDRServer(cfg, dm, fltrs, connMng, storDB)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				*reply.(*map[string]map[string]any) = map[string]map[string]any{}
				return utils.ErrNotFound
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs), utils.ThresholdSv1, rpcInternal)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
		APIOpts: nil,
	}
	_, err := newCDRSrv.processEvents(context.Background(), []*utils.CGREvent{cgrEv})
	if err != nil {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", nil, err)
	}
	expected := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
		APIOpts: map[string]any{},
	}
	delete(cgrEv.APIOpts, utils.MetaCDRID) // ignore autogenerated *cdr field when comparing
	if !reflect.DeepEqual(expected, cgrEv) {
		t.Errorf("\nExpected <%+v> \n,received <%+v>", expected, cgrEv)
	}
}

func TestCDRsProcessEventMockSkipOpts(t *testing.T) {
	testCache := engine.Cache
	tmpC := config.CgrConfig()
	defer func() {
		engine.Cache = testCache
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs)}

	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMng := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	storDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	newCDRSrv := NewCDRServer(cfg, dm, fltrs, connMng, storDB)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				*reply.(*map[string]map[string]any) = map[string]map[string]any{}
				return utils.ErrNotFound
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs), utils.ThresholdSv1, rpcInternal)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
		APIOpts: nil,
	}
	_, err := newCDRSrv.processEvents(context.Background(), []*utils.CGREvent{cgrEv})
	if err != nil {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", nil, err)
	}
	expected := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
		APIOpts: map[string]any{},
	}
	delete(cgrEv.APIOpts, utils.MetaCDRID) // ignore autogenerated *cdr field when comparing
	if !reflect.DeepEqual(expected, cgrEv) {
		t.Errorf("\nExpected <%+v> \n,received <%+v>", expected, cgrEv)
	}
}

func TestCDRsNewMapEventFromReqFormErr(t *testing.T) {
	httpReq := &http.Request{
		URL: &url.URL{
			RawQuery: "%0x",
		},
	}
	_, err := newMapEventFromReqForm(httpReq)
	errExpect := `invalid URL escape "%0x"`
	if err == nil || err.Error() != errExpect {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", errExpect, err)
	}

}

func TestCDRsProcessEventMockAttrsErr(t *testing.T) {
	testCache := engine.Cache
	tmpC := config.CgrConfig()
	defer func() {
		engine.Cache = testCache
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs)}
	cfg.CdrsCfg().Opts.Attributes = []*utils.DynamicBoolOpt{
		{
			Value: false,
		},
	}

	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMng := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	storDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	newCDRSrv := NewCDRServer(cfg, dm, fltrs, connMng, storDB)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				*reply.(*map[string]map[string]any) = map[string]map[string]any{}
				return utils.ErrNotFound
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs), utils.ThresholdSv1, rpcInternal)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
		APIOpts: map[string]any{
			utils.MetaAttributes: true,
			"*context":           utils.MetaCDRs,
		},
	}
	_, err := newCDRSrv.processEvents(context.Background(), []*utils.CGREvent{cgrEv})
	if err == nil || err.Error() != "ATTRIBUTES_ERROR:MANDATORY_IE_MISSING: [connIDs]" {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", "ATTRIBUTES_ERROR:MANDATORY_IE_MISSING: [connIDs]", err)
	}
}

func TestCDRsProcessEventMockAttrsErrBoolOpts(t *testing.T) {
	testCache := engine.Cache
	tmpC := config.CgrConfig()
	defer func() {
		engine.Cache = testCache
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs)}
	cfg.CdrsCfg().Opts.Attributes = []*utils.DynamicBoolOpt{
		{
			Value: false,
		},
	}

	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMng := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	storDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	newCDRSrv := NewCDRServer(cfg, dm, fltrs, connMng, storDB)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				*reply.(*map[string]map[string]any) = map[string]map[string]any{}
				return utils.ErrNotFound
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs), utils.ThresholdSv1, rpcInternal)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
		APIOpts: map[string]any{
			utils.MetaAttributes: time.Second,
			"*context":           utils.MetaCDRs,
		},
	}
	expectedErr := `retrieving *attributes option failed: cannot convert field: 1s to bool`
	_, err := newCDRSrv.processEvents(context.Background(), []*utils.CGREvent{cgrEv})
	if err == nil || err.Error() != expectedErr {
		t.Errorf("expected <%v>, received <%v>", expectedErr, err)
	}
}

func TestCDRsProcessEventMockChrgsErr(t *testing.T) {
	testCache := engine.Cache
	tmpC := config.CgrConfig()
	defer func() {
		engine.Cache = testCache
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs)}
	cfg.CdrsCfg().Opts.Attributes = []*utils.DynamicBoolOpt{
		{
			Value: false,
		},
	}

	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMng := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	storDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	newCDRSrv := NewCDRServer(cfg, dm, fltrs, connMng, storDB)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				*reply.(*map[string]map[string]any) = map[string]map[string]any{}
				return utils.ErrNotFound
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs), utils.ThresholdSv1, rpcInternal)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
		APIOpts: map[string]any{
			utils.MetaChargers: true,
			"*context":         utils.MetaCDRs,
		},
	}
	_, err := newCDRSrv.processEvents(context.Background(), []*utils.CGREvent{cgrEv})
	if err == nil || err.Error() != "CHARGERS_ERROR:MANDATORY_IE_MISSING: [connIDs]" {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", "CHARGERS_ERROR:MANDATORY_IE_MISSING: [connIDs]", err)
	}

}

func TestCDRsProcessEventMockChrgsErrBoolOpts(t *testing.T) {
	testCache := engine.Cache
	tmpC := config.CgrConfig()
	defer func() {
		engine.Cache = testCache
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs)}
	cfg.CdrsCfg().Opts.Attributes = []*utils.DynamicBoolOpt{
		{
			Value: false,
		},
	}

	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMng := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	storDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	newCDRSrv := NewCDRServer(cfg, dm, fltrs, connMng, storDB)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				*reply.(*map[string]map[string]any) = map[string]map[string]any{}
				return utils.ErrNotFound
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs), utils.ThresholdSv1, rpcInternal)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
		APIOpts: map[string]any{
			utils.MetaChargers: time.Second,
			"*context":         utils.MetaCDRs,
		},
	}
	expectedErr := `retrieving *chargers option failed: cannot convert field: 1s to bool`
	_, err := newCDRSrv.processEvents(context.Background(), []*utils.CGREvent{cgrEv})
	if err == nil || err.Error() != expectedErr {
		t.Errorf("expected <%v>, received <%v>", expectedErr, err)
	}

}

func TestCDRsProcessEventMockRateSErr(t *testing.T) {
	testCache := engine.Cache
	tmpC := config.CgrConfig()
	defer func() {
		engine.Cache = testCache
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs)}
	cfg.CdrsCfg().Opts.Attributes = []*utils.DynamicBoolOpt{
		{
			Value: false,
		},
	}

	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMng := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	storDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	newCDRSrv := NewCDRServer(cfg, dm, fltrs, connMng, storDB)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				*reply.(*map[string]map[string]any) = map[string]map[string]any{}
				return utils.ErrNotFound
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs), utils.ThresholdSv1, rpcInternal)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
		APIOpts: map[string]any{
			utils.MetaRates: true,
			"*context":      utils.MetaCDRs,
		},
	}
	_, err := newCDRSrv.processEvents(context.Background(), []*utils.CGREvent{cgrEv})
	if err == nil || err.Error() != "PARTIALLY_EXECUTED" {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", "PARTIALLY_EXECUTED", err)
	}

}

func TestCDRsProcessEventMockRateSErrBoolOpts(t *testing.T) {
	testCache := engine.Cache
	tmpC := config.CgrConfig()
	defer func() {
		engine.Cache = testCache
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs)}
	cfg.CdrsCfg().Opts.Attributes = []*utils.DynamicBoolOpt{
		{
			Value: false,
		},
	}

	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMng := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	storDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	newCDRSrv := NewCDRServer(cfg, dm, fltrs, connMng, storDB)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				*reply.(*map[string]map[string]any) = map[string]map[string]any{}
				return utils.ErrNotFound
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs), utils.ThresholdSv1, rpcInternal)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
		APIOpts: map[string]any{
			utils.MetaRates: time.Second,
			"*context":      utils.MetaCDRs,
		},
	}
	expectedErr := `retrieving *rates option failed: cannot convert field: 1s to bool`
	_, err := newCDRSrv.processEvents(context.Background(), []*utils.CGREvent{cgrEv})
	if err == nil || err.Error() != expectedErr {
		t.Errorf("expected <%v>, received <%v>", expectedErr, err)
	}

}

func TestCDRsProcessEventMockAcntsErr(t *testing.T) {
	testCache := engine.Cache
	tmpC := config.CgrConfig()
	defer func() {
		engine.Cache = testCache
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs)}
	cfg.CdrsCfg().Opts.Attributes = []*utils.DynamicBoolOpt{
		{
			Value: false,
		},
	}

	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMng := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	storDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	newCDRSrv := NewCDRServer(cfg, dm, fltrs, connMng, storDB)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				*reply.(*map[string]map[string]any) = map[string]map[string]any{}
				return utils.ErrNotFound
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs), utils.ThresholdSv1, rpcInternal)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
		APIOpts: map[string]any{
			utils.MetaAccounts: true,
			"*context":         utils.MetaCDRs,
		},
	}
	_, err := newCDRSrv.processEvents(context.Background(), []*utils.CGREvent{cgrEv})
	if err == nil || err.Error() != "PARTIALLY_EXECUTED" {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", "PARTIALLY_EXECUTED", err)
	}

}

func TestCDRsProcessEventMockAcntsErrBoolOpts(t *testing.T) {
	testCache := engine.Cache
	tmpC := config.CgrConfig()
	defer func() {
		engine.Cache = testCache
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs)}
	cfg.CdrsCfg().Opts.Attributes = []*utils.DynamicBoolOpt{
		{
			Value: false,
		},
	}

	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMng := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	storDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	newCDRSrv := NewCDRServer(cfg, dm, fltrs, connMng, storDB)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				*reply.(*map[string]map[string]any) = map[string]map[string]any{}
				return utils.ErrNotFound
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs), utils.ThresholdSv1, rpcInternal)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
		APIOpts: map[string]any{
			utils.MetaAccounts: time.Second,
			"*context":         utils.MetaCDRs,
		},
	}
	expectedErr := `retrieving *accounts option failed: cannot convert field: 1s to bool`
	_, err := newCDRSrv.processEvents(context.Background(), []*utils.CGREvent{cgrEv})
	if err == nil || err.Error() != expectedErr {
		t.Errorf("expected <%v>, received <%v>", expectedErr, err)
	}

}

func TestCDRsProcessEventMockExportErr(t *testing.T) {
	testCache := engine.Cache
	tmpC := config.CgrConfig()
	defer func() {
		engine.Cache = testCache
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs)}
	cfg.CdrsCfg().Opts.Attributes = []*utils.DynamicBoolOpt{
		{

			Value: false,
		},
	}

	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMng := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	storDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	newCDRSrv := NewCDRServer(cfg, dm, fltrs, connMng, storDB)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				*reply.(*map[string]map[string]any) = map[string]map[string]any{}
				return utils.ErrExists
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs), utils.ThresholdSv1, rpcInternal)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
		APIOpts: map[string]any{
			utils.OptsCDRsExport: true,
			"*context":           utils.MetaCDRs,
		},
	}
	_, err := newCDRSrv.processEvents(context.Background(), []*utils.CGREvent{cgrEv})
	if err == nil || err.Error() != "PARTIALLY_EXECUTED" {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", "PARTIALLY_EXECUTED", err)
	}

}

func TestCDRsProcessEventMockExportErrBoolOpts(t *testing.T) {
	testCache := engine.Cache
	tmpC := config.CgrConfig()
	defer func() {
		engine.Cache = testCache
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs)}
	cfg.CdrsCfg().Opts.Attributes = []*utils.DynamicBoolOpt{
		{
			Value: false,
		},
	}

	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMng := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	storDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	newCDRSrv := NewCDRServer(cfg, dm, fltrs, connMng, storDB)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				*reply.(*map[string]map[string]any) = map[string]map[string]any{}
				return utils.ErrExists
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs), utils.ThresholdSv1, rpcInternal)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
		APIOpts: map[string]any{
			utils.OptsCDRsExport: time.Second,
			"*context":           utils.MetaCDRs,
		},
	}
	expectedErr := `retrieving *cdrsExport option failed: cannot convert field: 1s to bool`
	_, err := newCDRSrv.processEvents(context.Background(), []*utils.CGREvent{cgrEv})
	if err == nil || err.Error() != expectedErr {
		t.Errorf("expected <%v>, received <%v>", expectedErr, err)
	}

}

func TestCDRsProcessEventMockThdsErr(t *testing.T) {
	testCache := engine.Cache
	tmpC := config.CgrConfig()
	defer func() {
		engine.Cache = testCache
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs)}
	cfg.CdrsCfg().Opts.Attributes = []*utils.DynamicBoolOpt{
		{
			Value: false,
		},
	}

	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMng := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	storDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	newCDRSrv := NewCDRServer(cfg, dm, fltrs, connMng, storDB)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				*reply.(*map[string]map[string]any) = map[string]map[string]any{}
				return utils.ErrNotFound
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs), utils.ThresholdSv1, rpcInternal)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
		APIOpts: map[string]any{
			utils.MetaThresholds: true,
			"*context":           utils.MetaCDRs,
		},
	}
	_, err := newCDRSrv.processEvents(context.Background(), []*utils.CGREvent{cgrEv})
	if err == nil || err.Error() != "PARTIALLY_EXECUTED" {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", "PARTIALLY_EXECUTED", err)
	}
}

func TestCDRsProcessEventMockThdsErrBoolOpts(t *testing.T) {
	testCache := engine.Cache
	tmpC := config.CgrConfig()
	defer func() {
		engine.Cache = testCache
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs)}
	cfg.CdrsCfg().Opts.Attributes = []*utils.DynamicBoolOpt{
		{
			Value: false,
		},
	}

	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMng := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	storDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	newCDRSrv := NewCDRServer(cfg, dm, fltrs, connMng, storDB)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				*reply.(*map[string]map[string]any) = map[string]map[string]any{}
				return utils.ErrNotFound
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs), utils.ThresholdSv1, rpcInternal)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
		APIOpts: map[string]any{
			utils.MetaThresholds: time.Second,
			"*context":           utils.MetaCDRs,
		},
	}
	expectedErr := `retrieving *thresholds option failed: cannot convert field: 1s to bool`
	_, err := newCDRSrv.processEvents(context.Background(), []*utils.CGREvent{cgrEv})
	if err == nil || err.Error() != expectedErr {
		t.Errorf("expected <%v>, received <%v>", expectedErr, err)
	}

}

func TestCDRsProcessEventMockStatsErr(t *testing.T) {
	testCache := engine.Cache
	tmpC := config.CgrConfig()
	defer func() {
		engine.Cache = testCache
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs)}
	cfg.CdrsCfg().Opts.Attributes = []*utils.DynamicBoolOpt{
		{
			Value: false,
		},
	}

	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMng := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	storDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	newCDRSrv := NewCDRServer(cfg, dm, fltrs, connMng, storDB)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				*reply.(*map[string]map[string]any) = map[string]map[string]any{}
				return utils.ErrExists
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs), utils.ThresholdSv1, rpcInternal)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
		APIOpts: map[string]any{
			utils.MetaStats: true,
			"*context":      utils.MetaCDRs,
		},
	}
	_, err := newCDRSrv.processEvents(context.Background(), []*utils.CGREvent{cgrEv})
	if err == nil || err.Error() != "PARTIALLY_EXECUTED" {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", "PARTIALLY_EXECUTED", err)
	}

}

func TestCDRsProcessEventMockStatsErrGetBoolOpts(t *testing.T) {
	testCache := engine.Cache
	tmpC := config.CgrConfig()
	defer func() {
		engine.Cache = testCache
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs)}
	cfg.CdrsCfg().Opts.Attributes = []*utils.DynamicBoolOpt{
		{
			Value: false,
		},
	}

	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMng := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	storDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	newCDRSrv := NewCDRServer(cfg, dm, fltrs, connMng, storDB)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				*reply.(*map[string]map[string]any) = map[string]map[string]any{}
				return utils.ErrExists
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs), utils.ThresholdSv1, rpcInternal)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
		APIOpts: map[string]any{
			utils.MetaStats: time.Second,
			"*context":      utils.MetaCDRs,
		},
	}
	expectedErr := `retrieving *stats option failed: cannot convert field: 1s to bool`
	_, err := newCDRSrv.processEvents(context.Background(), []*utils.CGREvent{cgrEv})
	if err == nil || err.Error() != expectedErr {
		t.Errorf("expected <%v>, received <%v>", expectedErr, err)
	}

}

func TestCDRsChrgrSProcessEventEmptyChrgrs(t *testing.T) {
	testCache := engine.Cache
	tmpC := config.CgrConfig()
	defer func() {
		engine.Cache = testCache
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaChargers)}

	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMng := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	storDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	newCDRSrv := NewCDRServer(cfg, dm, fltrs, connMng, storDB)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ChargerSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaChargers), utils.ChargerSv1, rpcInternal)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
		APIOpts: map[string]any{
			utils.MetaSubsys: utils.MetaChargers,
		},
	}
	_, err := newCDRSrv.chrgrSProcessEvent(context.Background(), cgrEv)
	if err != nil {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", nil, err)
	}

}

func TestCDRServerAccountSRefundCharges(t *testing.T) {
	testCache := engine.Cache
	tmpC := config.CgrConfig()
	defer func() {
		engine.Cache = testCache
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().AccountSConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.AccountSConnsCfg)}

	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMng := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	storDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	newCDRSrv := NewCDRServer(cfg, dm, fltrs, connMng, storDB)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.AccountSv1RefundCharges: func(ctx *context.Context, args, reply any) error {
				*reply.(*string) = utils.OK
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.AccountSConnsCfg), utils.AccountSv1, rpcInternal)

	apiOpts := map[string]any{
		utils.MetaAccountSCost: &utils.EventCharges{},
		utils.MetaSubsys:       utils.AccountSConnsCfg,
	}
	eChrgs := &utils.EventCharges{
		Abstracts: utils.NewDecimal(500, 0),
		Concretes: utils.NewDecimal(400, 0),
		Charges: []*utils.ChargeEntry{
			{
				ChargingID:     "GENUUID", //will be changed
				CompressFactor: 1,
			},
			{
				ChargingID:     "GENUUID2", //will be changed
				CompressFactor: 1,
			},
		},
		Accounting: map[string]*utils.AccountCharge{
			"GENUUID2": {
				BalanceID:    "CB2",
				Units:        utils.NewDecimal(2, 0),
				BalanceLimit: utils.NewDecimal(-1, 0),
			},
			"GENUUID": {
				BalanceID:    "CB1",
				Units:        utils.NewDecimal(7, 0),
				BalanceLimit: utils.NewDecimal(-200, 0),
				UnitFactorID: "GENNUUID_FACTOR",
			},
		},
		UnitFactors: map[string]*utils.UnitFactor{
			"GENNUUID_FACTOR": {
				Factor: utils.NewDecimal(100, 0),
			},
		},
		Rating:   make(map[string]*utils.RateSInterval),
		Rates:    make(map[string]*utils.IntervalRate),
		Accounts: make(map[string]*utils.Account),
	}
	err := newCDRSrv.accountSRefundCharges(context.Background(), "cgrates.org", eChrgs, apiOpts)
	if err != nil {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", nil, err)
	}
}
func TestCDRServerAccountSRefundChargesErr(t *testing.T) {
	testCache := engine.Cache
	tmpC := config.CgrConfig()
	defer func() {
		engine.Cache = testCache
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().AccountSConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.AccountSConnsCfg)}

	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMng := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	storDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	newCDRSrv := NewCDRServer(cfg, dm, fltrs, connMng, storDB)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ChargerSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				*reply.(*string) = utils.OK
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.AccountSConnsCfg), utils.AccountSv1, rpcInternal)

	apiOpts := map[string]any{
		utils.MetaAccountSCost: &utils.EventCharges{},
		utils.MetaSubsys:       utils.AccountSConnsCfg,
	}
	eChrgs := &utils.EventCharges{
		Abstracts: utils.NewDecimal(500, 0),
		Concretes: utils.NewDecimal(400, 0),
	}
	expErr := "UNSUPPORTED_SERVICE_METHOD"
	err := newCDRSrv.accountSRefundCharges(context.Background(), "cgrates.org", eChrgs, apiOpts)
	if err == nil || err.Error() != expErr {
		t.Errorf("\nExpected error <%v> \n, received error <%v>", expErr, err)
	}

}

func TestPopulateCost(t *testing.T) {

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.Usage: "10s",
		},
		APIOpts: map[string]any{
			utils.MetaAccountSCost: &utils.EventCharges{
				Concretes: utils.NewDecimal(400, 0),
			},
		},
	}
	exp := utils.NewDecimal(400, 0)
	if rcv := populateCost(ev.APIOpts); !reflect.DeepEqual(exp, rcv) {
		t.Errorf("Expected <%+v>, Received <%+v>", exp, rcv)
	}
	ev = &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.Usage: "10s",
		},
		APIOpts: map[string]any{

			utils.MetaRateSCost: utils.RateProfileCost{

				Cost: utils.NewDecimal(400, 0),
			},
		},
	}
	if rcv := populateCost(ev.APIOpts); !reflect.DeepEqual(exp, rcv) {
		t.Errorf("Expected <%+v>, Received <%+v>", exp, rcv)
	}
	ev = &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.Usage: "10s",
		},
		APIOpts: map[string]any{

			utils.MetaCost: 102.1,
		},
	}
	if rcv := populateCost(ev.APIOpts); rcv != nil {
		t.Errorf("Expected <%+v>, Received <%+v>", nil, rcv)
	}
}
func TestCDRsProcessEventMockThdsEcCostIface(t *testing.T) {
	testCache := engine.Cache
	tmpC := config.CgrConfig()
	defer func() {
		engine.Cache = testCache
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().AccountSConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaAccounts)}
	cfg.CdrsCfg().Opts.Attributes = []*utils.DynamicBoolOpt{
		{
			Value: false,
		},
	}

	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMng := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	storDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	newCDRSrv := NewCDRServer(cfg, dm, fltrs, connMng, storDB)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{

			utils.AccountSv1DebitAbstracts: func(ctx *context.Context, args, reply any) error {
				*reply.(*utils.EventCharges) = utils.EventCharges{
					Concretes: utils.NewDecimal(400, 0),
				}
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaAccounts), utils.AccountSv1, rpcInternal)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
		APIOpts: map[string]any{
			utils.MetaAccounts: true,
			"*context":         utils.MetaCDRs,
			utils.MetaAccountSCost: map[string]any{
				"Concretes": utils.NewDecimal(400, 0),
			},
		},
	}
	_, err := newCDRSrv.processEvents(context.Background(), []*utils.CGREvent{cgrEv})
	if err == nil || err.Error() != "PARTIALLY_EXECUTED" {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", "PARTIALLY_EXECUTED", err)
	}
}

func TestCDRsProcessEventMockThdsEcCostIfaceMarshalErr(t *testing.T) {
	testCache := engine.Cache
	tmpC := config.CgrConfig()
	defer func() {
		engine.Cache = testCache
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()

	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMng := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	storDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	newCDRSrv := NewCDRServer(cfg, dm, fltrs, connMng, storDB)

	rpcInternal := make(chan birpc.ClientConnector, 1)

	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaAccounts), utils.AccountSv1, rpcInternal)

	cgrEv := &utils.CGREvent{
		APIOpts: map[string]any{
			utils.MetaAccounts: true,
			"*context":         utils.MetaCDRs,
			utils.MetaAccountSCost: map[string]any{
				"Concretes": make(chan string),
			},
		},
	}
	_, err := newCDRSrv.processEvents(context.Background(), []*utils.CGREvent{cgrEv})
	if err == nil || err.Error() != "json: unsupported type: chan string" {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", "json: unsupported type: chan string", err)
	}
}

func TestCDRsProcessEventMockThdsEcCostIfaceUnmarshalErr(t *testing.T) {
	testCache := engine.Cache
	tmpC := config.CgrConfig()
	defer func() {
		engine.Cache = testCache
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()

	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMng := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	storDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	newCDRSrv := NewCDRServer(cfg, dm, fltrs, connMng, storDB)

	rpcInternal := make(chan birpc.ClientConnector, 1)

	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaAccounts), utils.AccountSv1, rpcInternal)

	cgrEv := &utils.CGREvent{
		APIOpts: map[string]any{
			utils.MetaAccounts: true,
			"*context":         utils.MetaCDRs,
			utils.MetaAccountSCost: map[string]any{
				"Charges": "not unmarshable",
			},
		},
	}
	expErr := "json: cannot unmarshal string into Go struct field EventCharges.Charges of type []*utils.ChargeEntry"
	_, err := newCDRSrv.processEvents(context.Background(), []*utils.CGREvent{cgrEv})
	if err == nil || err.Error() != expErr {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", expErr, err)
	}
}
