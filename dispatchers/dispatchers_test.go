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

package dispatchers

import (
	"reflect"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func TestDispatcherServiceDispatcherProfileForEventGetDispatcherProfileNF(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.DispatcherSCfg().IndexedSelects = false
	connMng := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(&engine.DataDBMock{
		GetKeysForPrefixF: func(*context.Context, string) ([]string, error) {
			return []string{"dpp_cgrates.org:123"}, nil
		},
	}, nil, connMng)
	dsp := &engine.DispatcherProfile{
		Tenant:         "cgrates.org",
		ID:             "321",
		FilterIDs:      []string{"filter", "*string:~*vars.*subsys:" + utils.MetaAccounts},
		Strategy:       "",
		StrategyParams: nil,
		Weight:         0,
		Hosts:          nil,
	}
	err := dm.SetDispatcherProfile(context.TODO(), dsp, false)
	if err == nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrNotImplemented, err)
	}
	fltr := &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "filter",
		Rules:  nil,
	}
	err = dm.SetFilter(context.Background(), fltr, false)
	if err == nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrNotImplemented, err)
	}
	fltrs := engine.NewFilterS(cfg, connMng, dm)
	dss := NewDispatcherService(dm, cfg, fltrs, connMng)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "321",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
			"Password":         "CGRateS.org",
			"RunID":            utils.MetaDefault,
		},
		APIOpts: map[string]interface{}{
			utils.Subsys: utils.MetaAccounts,
		},
	}
	tnt := ev.Tenant
	subsys := utils.IfaceAsString(ev.APIOpts[utils.Subsys])
	_, err = dss.dispatcherProfilesForEvent(context.Background(), tnt, ev, utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.Subsys: subsys,
		},
	})
	expected := utils.ErrNotImplemented
	if err == nil || err != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestDispatcherServiceDispatcherProfileForEventMIIDENotFound(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.DispatcherSCfg().IndexedSelects = false
	connMng := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMng)
	dss := NewDispatcherService(dm, cfg, engine.NewFilterS(cfg, connMng, dm), connMng)
	ev := &utils.CGREvent{}
	tnt := ""
	subsys := utils.IfaceAsString(ev.APIOpts[utils.Subsys])
	_, err := dss.dispatcherProfilesForEvent(context.Background(), tnt, ev, utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.Subsys: subsys,
		},
	})
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrNotFound, err)
	}
}

func (dS *DispatcherService) DispatcherServicePing(ev *utils.CGREvent, reply *string) error {
	*reply = utils.Pong
	return nil
}

func TestDispatcherauthorizeEvent(t *testing.T) {
	dm := &engine.DataManager{}
	cfg := config.NewDefaultCGRConfig()
	fltr := &engine.FilterS{}
	connMgr := &engine.ConnManager{}
	dsp := NewDispatcherService(dm, cfg, fltr, connMgr)
	ev := &utils.CGREvent{}
	reply := &engine.AttrSProcessEventReply{}
	err := dsp.authorizeEvent(ev, reply)
	expected := "MANDATORY_IE_MISSING: [connIDs]"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestDispatcherAuthorizeEventErr(t *testing.T) {
	dm := &engine.DataManager{}
	cfg := config.NewDefaultCGRConfig()
	fltr := &engine.FilterS{}
	connMgr := &engine.ConnManager{}
	dsp := NewDispatcherService(dm, cfg, fltr, connMgr)
	ev := &utils.CGREvent{}
	reply := &engine.AttrSProcessEventReply{}
	err := dsp.authorizeEvent(ev, reply)
	expected := "MANDATORY_IE_MISSING: [connIDs]"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestDispatcherV1GetProfileForEventErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultTenant = utils.EmptyString
	dsp := NewDispatcherService(nil, cfg, nil, nil)
	ev := &utils.CGREvent{}
	dPfl := &engine.DispatcherProfiles{}
	err := dsp.V1GetProfilesForEvent(context.Background(), ev, dPfl)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestDispatcherV1GetProfileForEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultTenant = utils.EmptyString
	dsp := NewDispatcherService(nil, cfg, nil, nil)
	ev := &utils.CGREvent{}
	dPfl := &engine.DispatcherProfiles{}
	err := dsp.V1GetProfilesForEvent(context.Background(), ev, dPfl)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestDispatcherDispatch(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultTenant = utils.EmptyString
	dsp := NewDispatcherService(nil, cfg, nil, nil)
	ev := &utils.CGREvent{}
	err := dsp.Dispatch(context.TODO(), ev, "", "", "", "")
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestDispatcherAuthorizeError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.DispatcherSCfg().AttributeSConns = []string{"connID"}
	cfg.RPCConns()["connID"] = &config.RPCConn{
		Strategy: rpcclient.PoolFirst,
		PoolSize: 0,
		Conns: []*config.RemoteHost{
			{
				ID:        "",
				Address:   "error",
				Transport: "",
				TLS:       false,
			},
		},
	}
	connMng := engine.NewConnManager(cfg)
	dsp := NewDispatcherService(nil, cfg, nil, connMng)
	err := dsp.authorize("", "cgrates.org", utils.APIMethods)
	expected := "dial tcp: address error: missing port in address"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestDispatcherAuthorizeError2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.DispatcherSCfg().AttributeSConns = []string{utils.APIMethods}
	cfg.RPCConns()[utils.APIMethods] = &config.RPCConn{
		Strategy: rpcclient.PoolFirst,
		PoolSize: 0,
		Conns: []*config.RemoteHost{
			{
				ID:        "",
				Address:   "error",
				Transport: "",
				TLS:       false,
			},
		},
	}
	connMng := engine.NewConnManager(cfg)
	dsp := NewDispatcherService(nil, cfg, nil, connMng)
	err := dsp.authorize("", "cgrates.org", utils.APIMethods)
	expected := "dial tcp: address error: missing port in address"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestDispatcherServiceAuthorizeEvenError1(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(nil, nil, nil)
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	fltr := &engine.FilterS{}
	connMgr := &engine.ConnManager{}
	dsp := NewDispatcherService(dm, cfg, fltr, connMgr)
	cfg.DispatcherSCfg().AttributeSConns = []string{"connID"}
	ev := &utils.CGREvent{}
	reply := &engine.AttrSProcessEventReply{}
	engine.Cache.SetWithoutReplicate(utils.CacheRPCConnections, "connID",
		nil, nil, true, utils.NonTransactional)
	err := dsp.authorizeEvent(ev, reply)
	expected := "UNKNOWN_API_KEY"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
	engine.Cache = cacheInit
}

func TestDispatcherServiceAuthorizeEventError2(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(nil, nil, nil)
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	fltr := &engine.FilterS{}
	connMgr := &engine.ConnManager{}
	dsp := NewDispatcherService(dm, cfg, fltr, connMgr)
	cfg.DispatcherSCfg().AttributeSConns = []string{"connID"}
	ev := &utils.CGREvent{}
	reply := &engine.AttrSProcessEventReply{}
	dh := &engine.DispatcherHost{
		Tenant: "testTenant",
		RemoteHost: &config.RemoteHost{
			ID:        "testID",
			Address:   "",
			Transport: "",
			TLS:       false,
		},
	}
	value := &lazzyDH{dh: dh, cfg: cfg, iPRCCh: nil}
	engine.Cache.SetWithoutReplicate(utils.CacheRPCConnections, "connID",
		value, nil, true, utils.NonTransactional)

	expected := "dial tcp: missing address"
	if err := dsp.authorizeEvent(ev, reply); err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
	engine.Cache = cacheInit
}

type mockTypeCon2 struct{}

func (*mockTypeCon2) Call(ctx *context.Context, serviceMethod string, args, reply interface{}) error {
	return nil
}

func TestDispatcherServiceAuthorizeEventError3(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.DispatcherSCfg().AttributeSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}
	dm := engine.NewDataManager(nil, nil, nil)
	chanRPC := make(chan birpc.ClientConnector, 1)
	chanRPC <- new(mockTypeCon2)
	connMgr := engine.NewConnManager(cfg)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), utils.AttributeSv1, chanRPC)

	dsp := NewDispatcherService(dm, cfg, nil, connMgr)
	ev := &utils.CGREvent{
		Tenant:  "testTenant",
		ID:      "testID",
		Event:   map[string]interface{}{},
		APIOpts: nil,
	}
	dh := &engine.DispatcherHost{
		Tenant: "testTenant",
		RemoteHost: &config.RemoteHost{
			ID:        "testID",
			Address:   rpcclient.InternalRPC,
			Transport: utils.MetaInternal,
			TLS:       false,
		},
	}
	value := &lazzyDH{dh: dh, cfg: cfg, iPRCCh: chanRPC}

	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	engine.Cache.SetWithoutReplicate(utils.CacheRPCConnections, "testID",
		value, nil, true, utils.NonTransactional)
	rply := &engine.AttrSProcessEventReply{}
	if err := dsp.authorizeEvent(ev, rply); err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	engine.Cache = cacheInit
}

type mockTypeCon3 struct{}

func (*mockTypeCon3) Call(ctx *context.Context, serviceMethod string, args, reply interface{}) error {
	eVreply := &engine.AttrSProcessEventReply{
		CGREvent: &utils.CGREvent{
			Tenant: "testTenant",
			ID:     "testID",
			Event: map[string]interface{}{
				utils.APIMethods: "yes",
			},
			APIOpts: nil,
		},
	}
	*reply.(*engine.AttrSProcessEventReply) = *eVreply
	return nil
}

func TestDispatcherServiceAuthorizeError(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.DispatcherSCfg().AttributeSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}
	dm := engine.NewDataManager(nil, nil, nil)
	chanRPC := make(chan birpc.ClientConnector, 1)
	chanRPC <- new(mockTypeCon3)
	connMgr := engine.NewConnManager(cfg)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), utils.AttributeSv1, chanRPC)

	dsp := NewDispatcherService(dm, cfg, nil, connMgr)
	dh := &engine.DispatcherHost{
		Tenant: "testTenant",
		RemoteHost: &config.RemoteHost{
			ID:        "testID",
			Address:   rpcclient.InternalRPC,
			Transport: utils.MetaInternal,
			TLS:       false,
		},
	}
	value := &lazzyDH{dh: dh, cfg: cfg, iPRCCh: chanRPC}
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	engine.Cache.SetWithoutReplicate(utils.CacheRPCConnections, "testID",
		value, nil, true, utils.NonTransactional)
	expected := "UNAUTHORIZED_API"
	if err := dsp.authorize(utils.APIMethods, "testTenant", "apikey"); err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
	engine.Cache = cacheInit
}

type mockTypeCon4 struct{}

func (*mockTypeCon4) Call(ctx *context.Context, serviceMethod string, args, reply interface{}) error {
	eVreply := &engine.AttrSProcessEventReply{
		CGREvent: &utils.CGREvent{
			Tenant:  "testTenant",
			ID:      "testID",
			Event:   map[string]interface{}{},
			APIOpts: nil,
		},
	}
	*reply.(*engine.AttrSProcessEventReply) = *eVreply
	return nil
}

func TestDispatcherServiceAuthorizeError2(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.DispatcherSCfg().AttributeSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}
	dm := engine.NewDataManager(nil, nil, nil)
	chanRPC := make(chan birpc.ClientConnector, 1)
	chanRPC <- new(mockTypeCon4)
	connMgr := engine.NewConnManager(cfg)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), utils.AttributeSv1, chanRPC)

	dsp := NewDispatcherService(dm, cfg, nil, connMgr)
	dh := &engine.DispatcherHost{
		Tenant: "testTenant",
		RemoteHost: &config.RemoteHost{
			ID:        "testID",
			Address:   rpcclient.InternalRPC,
			Transport: utils.MetaInternal,
			TLS:       false,
		},
	}
	value := &lazzyDH{dh: dh, cfg: cfg, iPRCCh: chanRPC}

	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	engine.Cache.SetWithoutReplicate(utils.CacheRPCConnections, "testID",
		value, nil, true, utils.NonTransactional)
	expected := "NOT_FOUND"
	if err := dsp.authorize(utils.APIMethods, "testTenant", "apikey"); err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
	engine.Cache = cacheInit
}

type mockTypeCon5 struct{}

func (*mockTypeCon5) Call(ctx *context.Context, serviceMethod string, args, reply interface{}) error {
	eVreply := &engine.AttrSProcessEventReply{
		CGREvent: &utils.CGREvent{
			Tenant: "testTenant",
			ID:     "testID",
			Event: map[string]interface{}{
				utils.APIMethods: "testMethod",
			},
			APIOpts: nil,
		},
	}
	*reply.(*engine.AttrSProcessEventReply) = *eVreply
	return nil
}

func TestDispatcherServiceAuthorizeError3(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.DispatcherSCfg().AttributeSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}
	dm := engine.NewDataManager(nil, nil, nil)
	chanRPC := make(chan birpc.ClientConnector, 1)
	chanRPC <- new(mockTypeCon5)
	connMgr := engine.NewConnManager(cfg)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), utils.AttributeSv1, chanRPC)

	dsp := NewDispatcherService(dm, cfg, nil, connMgr)
	dh := &engine.DispatcherHost{
		Tenant: "testTenant",
		RemoteHost: &config.RemoteHost{
			ID:        "testID",
			Address:   rpcclient.InternalRPC,
			Transport: utils.MetaInternal,
			TLS:       false,
		},
	}
	value := &lazzyDH{dh: dh, cfg: cfg, iPRCCh: chanRPC}

	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	engine.Cache.SetWithoutReplicate(utils.CacheRPCConnections, "testID",
		value, nil, true, utils.NonTransactional)
	if err := dsp.authorize("testMethod", "testTenant", "apikey"); err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	engine.Cache = cacheInit
}

func TestDispatcherServiceDispatcherProfileForEventErrNil(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.DispatcherSCfg().IndexedSelects = false
	connMng := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMng)
	dsp := &engine.DispatcherProfile{
		Tenant:         "cgrates.org",
		ID:             "123",
		FilterIDs:      []string{"*string:~*vars.*subsys:" + utils.MetaAccounts},
		Strategy:       "",
		StrategyParams: nil,
		Weight:         0,
		Hosts:          nil,
	}
	err := dm.SetDispatcherProfile(context.TODO(), dsp, false)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	dss := NewDispatcherService(dm, cfg, engine.NewFilterS(cfg, connMng, dm), connMng)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "123",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
			"Password":         "CGRateS.org",
			"RunID":            utils.MetaDefault,
		},
		APIOpts: map[string]interface{}{
			utils.Subsys: utils.MetaAccounts,
		},
	}
	tnt := ev.Tenant
	subsys := utils.IfaceAsString(ev.APIOpts[utils.Subsys])
	_, err = dss.dispatcherProfilesForEvent(context.Background(), tnt, ev, utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.Subsys: subsys,
		},
	})
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
}

func TestDispatcherV1GetProfileForEventReturn(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.DispatcherSCfg().IndexedSelects = false
	connMng := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMng)
	dsp := &engine.DispatcherProfile{
		Tenant:         "cgrates.org",
		ID:             "123",
		FilterIDs:      []string{"*string:~*vars.*subsys:" + utils.MetaAccounts},
		Strategy:       "",
		StrategyParams: nil,
		Weight:         0,
		Hosts:          nil,
	}
	err := dm.SetDispatcherProfile(context.TODO(), dsp, false)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	dss := NewDispatcherService(dm, cfg, engine.NewFilterS(cfg, connMng, dm), connMng)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "123",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
			"Password":         "CGRateS.org",
			"RunID":            utils.MetaDefault,
		},
		APIOpts: map[string]interface{}{
			utils.Subsys: utils.MetaAccounts,
		},
	}
	tnt := ev.Tenant
	subsys := utils.IfaceAsString(ev.APIOpts[utils.Subsys])
	_, err = dss.dispatcherProfilesForEvent(context.Background(), tnt, ev, utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.Subsys: subsys,
		},
	})
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	dPfl := &engine.DispatcherProfiles{}
	err = dss.V1GetProfilesForEvent(context.Background(), ev, dPfl)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestDispatcherServiceDispatcherProfileForEventErrNotFound(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.DispatcherSCfg().IndexedSelects = false
	connMng := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMng)
	dsp := &engine.DispatcherProfile{
		Tenant:         "cgrates.org",
		ID:             "123",
		FilterIDs:      []string{"*string:~*vars.*subsys:" + utils.MetaAccounts},
		Strategy:       "",
		StrategyParams: nil,
		Weight:         0,
		Hosts:          nil,
	}
	err := dm.SetDispatcherProfile(context.TODO(), dsp, false)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	dss := NewDispatcherService(dm, cfg, engine.NewFilterS(cfg, connMng, dm), connMng)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "123",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
			"Password":         "CGRateS.org",
			"RunID":            utils.MetaDefault,
		},
		APIOpts: map[string]interface{}{
			utils.Subsys: utils.MetaAny,
		},
	}
	tnt := ev.Tenant
	subsys := utils.IfaceAsString(ev.APIOpts[utils.Subsys])
	_, err = dss.dispatcherProfilesForEvent(context.Background(), tnt, ev, utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.Subsys: subsys,
		},
	})
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrNotFound, err)
	}
}

func TestDispatcherServiceDispatcherProfileForEventErrNotFound2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.DispatcherSCfg().IndexedSelects = false
	connMng := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMng)
	dsp := &engine.DispatcherProfile{
		Tenant:         "cgrates.org",
		ID:             "123",
		FilterIDs:      []string{"*string:~*vars.*subsys:" + utils.MetaAccounts},
		Strategy:       "",
		StrategyParams: nil,
		Weight:         0,
		Hosts:          nil,
	}
	err := dm.SetDispatcherProfile(context.TODO(), dsp, false)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	dss := NewDispatcherService(dm, cfg, engine.NewFilterS(cfg, connMng, dm), connMng)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "123",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
			"Password":         "CGRateS.org",
			"RunID":            utils.MetaDefault,
		},
		APIOpts: map[string]interface{}{
			utils.Subsys: utils.MetaAccounts,
		},
	}
	tnt := ""
	subsys := utils.IfaceAsString(ev.APIOpts[utils.Subsys])
	_, err = dss.dispatcherProfilesForEvent(context.Background(), tnt, ev, utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.Subsys: subsys,
		},
	})
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrNotFound, err)
	}
}

func TestDispatcherServiceDispatcherProfileForEventErrNotFoundFilter(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.DispatcherSCfg().IndexedSelects = false
	connMng := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMng)
	dsp := &engine.DispatcherProfile{
		Tenant:         "cgrates.org",
		ID:             "123",
		FilterIDs:      []string{"filter", "*string:~*vars.*subsys:" + utils.MetaAccounts},
		Strategy:       "",
		StrategyParams: nil,
		Weight:         0,
		Hosts:          nil,
	}
	err := dm.SetDispatcherProfile(context.TODO(), dsp, false)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	fltrs := engine.NewFilterS(cfg, connMng, dm)
	dss := NewDispatcherService(dm, cfg, fltrs, connMng)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "123",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
			"Password":         "CGRateS.org",
			"RunID":            utils.MetaDefault,
		},
		APIOpts: map[string]interface{}{
			utils.Subsys: utils.MetaAccounts,
		},
	}
	tnt := ev.Tenant
	subsys := utils.IfaceAsString(ev.APIOpts[utils.Subsys])
	_, err = dss.dispatcherProfilesForEvent(context.Background(), tnt, ev, utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.Subsys: subsys,
		},
	})
	if err == nil || err.Error() != "NOT_FOUND:filter" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "NOT_FOUND:filter", err)
	}
}

func TestDispatcherServiceDispatchDspErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.DispatcherSCfg().IndexedSelects = false
	connMng := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMng)
	dsp := &engine.DispatcherProfile{
		Tenant:         "cgrates.org",
		ID:             "123",
		FilterIDs:      []string{"*string:~*vars.*subsys:" + utils.MetaAccounts},
		Strategy:       "",
		StrategyParams: nil,
		Weight:         0,
		Hosts:          nil,
	}
	err := dm.SetDispatcherProfile(context.TODO(), dsp, false)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	dss := NewDispatcherService(dm, cfg, engine.NewFilterS(cfg, connMng, dm), connMng)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "123",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
			"Password":         "CGRateS.org",
			"RunID":            utils.MetaDefault,
		},
		APIOpts: map[string]interface{}{
			utils.Subsys: utils.MetaAccounts,
		},
	}
	subsys := utils.IfaceAsString(ev.APIOpts[utils.Subsys])
	err = dss.Dispatch(context.TODO(), ev, subsys, "", "", "")
	expected := "DISPATCHER_ERROR:unsupported dispatch strategy: <>"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
	dss.Shutdown()
}

func TestDispatcherServiceDispatchDspErrHostNotFound(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.DispatcherSCfg().IndexedSelects = false
	connMng := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMng)
	dsp := &engine.DispatcherProfile{
		Tenant:         "cgrates.org",
		ID:             "123",
		FilterIDs:      []string{"*string:~*vars.*subsys:" + utils.MetaAccounts},
		StrategyParams: make(map[string]interface{}),
		Strategy:       utils.MetaWeight,
		Weight:         0,
		Hosts:          nil,
	}
	newCache := engine.NewCacheS(cfg, dm, nil)
	value, errDsp := newDispatcher(dsp)
	if errDsp != nil {
		t.Fatal(errDsp)
	}
	ctx := &context.Context{}
	engine.Cache = newCache
	engine.Cache.Set(ctx, utils.CacheDispatchers, dsp.TenantID(), value, nil, true, utils.EmptyString)

	err := dm.SetDispatcherProfile(context.TODO(), dsp, false)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	dss := NewDispatcherService(dm, cfg, engine.NewFilterS(cfg, connMng, dm), connMng)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "123",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
			"Password":         "CGRateS.org",
			"RunID":            utils.MetaDefault,
		},
		APIOpts: map[string]interface{}{
			utils.Subsys: utils.MetaAccounts,
		},
	}
	subsys := utils.IfaceAsString(ev.APIOpts[utils.Subsys])
	err = dss.Dispatch(context.TODO(), ev, subsys, "", "", "")
	expected := "HOST_NOT_FOUND"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
	engine.Cache = cacheInit
}

func TestDispatcherServiceDispatcherProfileForEventFoundFilter(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.DispatcherSCfg().IndexedSelects = false
	connMng := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMng)
	dsp := &engine.DispatcherProfile{
		Tenant:         "cgrates.org",
		ID:             "123",
		FilterIDs:      []string{"*string:~*req.RunID:1", "*string:~*vars.*subsys:" + utils.MetaAccounts},
		Strategy:       "",
		StrategyParams: nil,
		Weight:         0,
		Hosts:          nil,
	}
	err := dm.SetDispatcherProfile(context.TODO(), dsp, false)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	fltrs := engine.NewFilterS(cfg, connMng, dm)
	dss := NewDispatcherService(dm, cfg, fltrs, connMng)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "123",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
			"Password":         "CGRateS.org",
			"RunID":            utils.MetaDefault,
		},
		APIOpts: map[string]interface{}{
			utils.Subsys: utils.MetaAccounts,
		},
	}
	tnt := ev.Tenant
	subsys := utils.IfaceAsString(ev.APIOpts[utils.Subsys])
	_, err = dss.dispatcherProfilesForEvent(context.Background(), tnt, ev, utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.Subsys: subsys,
		},
	})
	if err == nil || err.Error() != "NOT_FOUND" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "NOT_FOUND:filter", err)
	}
}

func TestDispatcherServiceDispatcherProfileForEventNotNotFound(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.DispatcherSCfg().IndexedSelects = true
	connMng := engine.NewConnManager(cfg)
	var cnt int

	dm := engine.NewDataManager(&engine.DataDBMock{
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			if cnt == 0 {
				cnt++
				return map[string]utils.StringSet{
					idxKey: {"cgrates.org:dsp1": {}},
				}, nil
			}
			return nil, utils.ErrNotImplemented
		},
	}, nil, connMng)
	fltrs := engine.NewFilterS(cfg, connMng, dm)
	dss := NewDispatcherService(dm, cfg, fltrs, connMng)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "123",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
			"Password":         "CGRateS.org",
			"RunID":            utils.MetaDefault,
		},
		APIOpts: map[string]interface{}{
			utils.Subsys: utils.MetaAccounts,
		},
	}
	tnt := ev.Tenant
	subsys := utils.IfaceAsString(ev.APIOpts[utils.Subsys])
	_, err := dss.dispatcherProfilesForEvent(context.Background(), tnt, ev, utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.Subsys: subsys,
		},
	})
	expected := utils.ErrNotImplemented
	if err == nil || err != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestDispatcherServiceDispatcherProfileForEventGetDispatcherError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.DispatcherSCfg().IndexedSelects = false
	connMng := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMng)
	dsp := &engine.DispatcherProfile{
		Tenant:         "cgrates.org",
		ID:             "123",
		FilterIDs:      []string{"*string:~*req.RunID:1", "*string:~*vars.*subsys:" + utils.MetaAccounts},
		Strategy:       "",
		StrategyParams: nil,
		Weight:         0,
		Hosts:          nil,
	}
	err := dm.SetDispatcherProfile(context.TODO(), dsp, false)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	fltrs := engine.NewFilterS(cfg, connMng, dm)
	dss := NewDispatcherService(dm, cfg, fltrs, connMng)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "123",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
			"Password":         "CGRateS.org",
			"RunID":            utils.MetaDefault,
		},
		APIOpts: map[string]interface{}{
			utils.Subsys: utils.MetaAccounts,
		},
	}
	tnt := ev.Tenant
	subsys := utils.IfaceAsString(ev.APIOpts[utils.Subsys])
	_, err = dss.dispatcherProfilesForEvent(context.Background(), tnt, ev, utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.Subsys: subsys,
		},
	})
	if err == nil || err.Error() != "NOT_FOUND" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "NOT_FOUND:filter", err)
	}
}

func TestDispatcherServiceDispatchDspErrHostNotFound2(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.DispatcherSCfg().IndexedSelects = false
	connMng := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMng)
	dsp := &engine.DispatcherProfile{
		Tenant:         "cgrates.org",
		ID:             "123",
		FilterIDs:      []string{"*string:~*vars.*subsys:" + utils.MetaAccounts},
		StrategyParams: make(map[string]interface{}),
		Strategy:       utils.MetaWeight,
		Weight:         0,
		Hosts:          nil,
	}
	newCache := engine.NewCacheS(cfg, dm, nil)
	value, errDsp := newDispatcher(dsp)
	if errDsp != nil {
		t.Fatal(errDsp)
	}
	ctx := &context.Context{}
	engine.Cache = newCache
	engine.Cache.Set(ctx, utils.CacheDispatchers, dsp.TenantID(), value, nil, true, utils.EmptyString)

	err := dm.SetDispatcherProfile(context.TODO(), dsp, false)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	dss := NewDispatcherService(dm, cfg, engine.NewFilterS(cfg, connMng, dm), connMng)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "123",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
			"Password":         "CGRateS.org",
			"RunID":            utils.MetaDefault,
		},
		APIOpts: map[string]interface{}{
			utils.Subsys: utils.MetaAccounts,
		},
	}
	subsys := utils.IfaceAsString(ev.APIOpts[utils.Subsys])
	err = dss.Dispatch(context.TODO(), ev, subsys, "", "", "")
	expected := "HOST_NOT_FOUND"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
	engine.Cache = cacheInit
}

type mockTypeConSetCache struct{}

func (*mockTypeConSetCache) Call(ctx *context.Context, serviceMethod string, args, reply interface{}) error {
	return utils.ErrNotImplemented
}

func TestDispatcherServiceDispatchDspErrHostNotFound3(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.CacheCfg().Partitions[utils.CacheDispatchers] = &config.CacheParamCfg{
		Replicate: true,
	}
	cfg.DispatcherSCfg().IndexedSelects = false
	chanRPC := make(chan birpc.ClientConnector, 1)
	chanRPC <- new(mockTypeConSetCache)
	connMgr := engine.NewConnManager(cfg)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.ReplicatorSv1, chanRPC)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	dsp := &engine.DispatcherProfile{
		Tenant:         "cgrates.org",
		ID:             "123",
		FilterIDs:      []string{"*string:~*vars.*subsys:" + utils.MetaAccounts},
		StrategyParams: make(map[string]interface{}),
		Strategy:       utils.MetaWeight,
		Weight:         0,
		Hosts:          nil,
	}
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache

	err := dm.SetDispatcherProfile(context.TODO(), dsp, false)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	dss := NewDispatcherService(dm, cfg, engine.NewFilterS(cfg, connMgr, dm), connMgr)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "123",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
			"Password":         "CGRateS.org",
			"RunID":            utils.MetaDefault,
		},
		APIOpts: map[string]interface{}{
			utils.Subsys: utils.MetaAccounts,
		},
	}
	subsys := utils.IfaceAsString(ev.APIOpts[utils.Subsys])
	err = dss.Dispatch(context.TODO(), ev, subsys, "", "", "")
	expected := "DISPATCHER_ERROR:NOT_IMPLEMENTED"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
	engine.Cache = cacheInit
}

func TestDispatchersdispatcherProfileForEventAnySSfalseFirstNotFound(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil)

	dS := &DispatcherService{
		cfg:   cfg,
		dm:    dm,
		fltrS: engine.NewFilterS(cfg, nil, dm),
	}

	dsp1 := &engine.DispatcherProfile{
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.Account:1002", "*string:~*vars.*subsys:" + utils.MetaSessionS},
		ID:        "DSP_1",
		Strategy:  "*weight",
		Weight:    10,
	}
	err := dS.dm.SetDispatcherProfile(context.TODO(), dsp1, true)
	if err != nil {
		t.Error(err)
	}

	dsp2 := &engine.DispatcherProfile{
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		ID:        "DSP_2",
		Strategy:  "*weight",
		Weight:    20,
	}
	err = dS.dm.SetDispatcherProfile(context.TODO(), dsp2, true)
	if err != nil {
		t.Error(err)
	}

	tnt := "cgrates.org"
	ev := &utils.CGREvent{
		Tenant: tnt,
		Event: map[string]interface{}{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]interface{}{
			utils.OptsDispatchersProfilesCount: 1,
		},
	}
	subsys := utils.MetaSessionS

	if rcv, err := dS.dispatcherProfilesForEvent(context.Background(), tnt, ev, utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.Subsys: subsys,
		},
	}); err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	} else if len(rcv) != 1 {
		t.Errorf("Unexpected number of profiles:%v", len(rcv))
	} else if !reflect.DeepEqual(rcv[0], dsp2) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", dsp2, rcv)
	}
}

func TestDispatchersdispatcherProfileForEventAnySSfalseFound(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil)

	dS := &DispatcherService{
		cfg:   cfg,
		dm:    dm,
		fltrS: engine.NewFilterS(cfg, nil, dm),
	}

	dsp1 := &engine.DispatcherProfile{
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.Account:1001", "*string:~*vars.*subsys:" + utils.MetaSessionS},
		ID:        "DSP_1",
		Strategy:  "*weight",
		Weight:    20,
	}
	err := dS.dm.SetDispatcherProfile(context.TODO(), dsp1, true)
	if err != nil {
		t.Error(err)
	}

	dsp2 := &engine.DispatcherProfile{
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		ID:        "DSP_2",
		Strategy:  "*weight",
		Weight:    10,
	}
	err = dS.dm.SetDispatcherProfile(context.TODO(), dsp2, true)
	if err != nil {
		t.Error(err)
	}

	tnt := "cgrates.org"
	ev := &utils.CGREvent{
		Tenant: tnt,
		Event: map[string]interface{}{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]interface{}{
			utils.OptsDispatchersProfilesCount: 1,
		},
	}
	subsys := utils.MetaSessionS

	if rcv, err := dS.dispatcherProfilesForEvent(context.Background(), tnt, ev, utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.Subsys: subsys,
		},
	}); err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	} else if len(rcv) != 1 {
		t.Errorf("Unexpected number of profiles:%v", len(rcv))
	} else if !reflect.DeepEqual(rcv[0], dsp1) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", dsp1, rcv)
	}
}

func TestDispatchersdispatcherProfileForEventAnySSfalseNotFound(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil)

	dS := &DispatcherService{
		cfg:   cfg,
		dm:    dm,
		fltrS: engine.NewFilterS(cfg, nil, dm),
	}

	dsp1 := &engine.DispatcherProfile{
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.Account:1002", "*string:~*vars.*subsys:" + utils.MetaSessionS},
		ID:        "DSP_1",
		Strategy:  "*weight",
		Weight:    20,
	}
	err := dS.dm.SetDispatcherProfile(context.TODO(), dsp1, true)
	if err != nil {
		t.Error(err)
	}

	dsp2 := &engine.DispatcherProfile{
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.Account:1002"},
		ID:        "DSP_2",
		Strategy:  "*weight",
		Weight:    10,
	}
	err = dS.dm.SetDispatcherProfile(context.TODO(), dsp2, true)
	if err != nil {
		t.Error(err)
	}

	tnt := "cgrates.org"
	ev := &utils.CGREvent{
		Tenant: tnt,
		Event: map[string]interface{}{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]interface{}{
			utils.OptsDispatchersProfilesCount: 1,
		},
	}
	subsys := utils.MetaSessionS

	if rcv, err := dS.dispatcherProfilesForEvent(context.Background(), tnt, ev, utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.Subsys: subsys,
		},
	}); err == nil || err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	} else if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}
}

func TestDispatchersdispatcherProfileForEventAnySStrueNotFound(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil)

	dS := &DispatcherService{
		cfg:   cfg,
		dm:    dm,
		fltrS: engine.NewFilterS(cfg, nil, dm),
	}

	dsp1 := &engine.DispatcherProfile{
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.Account:1002", "*string:~*vars.*subsys:" + utils.MetaSessionS},
		ID:        "DSP_1",
		Strategy:  "*weight",
		Weight:    20,
	}
	err := dS.dm.SetDispatcherProfile(context.TODO(), dsp1, true)
	if err != nil {
		t.Error(err)
	}

	dsp2 := &engine.DispatcherProfile{
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.Account:1002"},
		ID:        "DSP_2",
		Strategy:  "*weight",
		Weight:    10,
	}
	err = dS.dm.SetDispatcherProfile(context.TODO(), dsp2, true)
	if err != nil {
		t.Error(err)
	}

	tnt := "cgrates.org"
	ev := &utils.CGREvent{
		Tenant: tnt,
		Event: map[string]interface{}{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]interface{}{
			utils.OptsDispatchersProfilesCount: 1,
		},
	}
	subsys := utils.MetaSessionS

	if rcv, err := dS.dispatcherProfilesForEvent(context.Background(), tnt, ev, utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.Subsys: subsys,
		},
	}); err == nil || err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	} else if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}
}

func TestDispatchersdispatcherProfileForEventAnySStrueBothFound(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil)

	dS := &DispatcherService{
		cfg:   cfg,
		dm:    dm,
		fltrS: engine.NewFilterS(cfg, nil, dm),
	}

	dsp1 := &engine.DispatcherProfile{
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.Account:1001", "*string:~*vars.*subsys:" + utils.MetaSessionS},
		ID:        "DSP_1",
		Strategy:  "*weight",
		Weight:    10,
	}
	err := dS.dm.SetDispatcherProfile(context.TODO(), dsp1, true)
	if err != nil {
		t.Error(err)
	}

	dsp2 := &engine.DispatcherProfile{
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		ID:        "DSP_2",
		Strategy:  "*weight",
		Weight:    20,
	}
	err = dS.dm.SetDispatcherProfile(context.TODO(), dsp2, true)
	if err != nil {
		t.Error(err)
	}

	tnt := "cgrates.org"
	ev := &utils.CGREvent{
		Tenant: tnt,
		Event: map[string]interface{}{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]interface{}{
			utils.OptsDispatchersProfilesCount: 1,
		},
	}
	subsys := utils.MetaSessionS

	if rcv, err := dS.dispatcherProfilesForEvent(context.Background(), tnt, ev, utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.Subsys: subsys,
		},
	}); err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	} else if len(rcv) != 1 {
		t.Errorf("Unexpected number of profiles:%v", len(rcv))
	} else if !reflect.DeepEqual(rcv[0], dsp2) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", dsp2, rcv)
	}

	dsp1.Weight = 30
	err = dS.dm.SetDispatcherProfile(context.TODO(), dsp1, true)
	if err != nil {
		t.Error(err)
	}

	if rcv, err := dS.dispatcherProfilesForEvent(context.Background(), tnt, ev, utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.Subsys: subsys,
		},
	}); err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	} else if len(rcv) != 1 {
		t.Errorf("Unexpected number of profiles:%v", len(rcv))
	} else if !reflect.DeepEqual(rcv[0], dsp1) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", dsp1, rcv)
	}
}
