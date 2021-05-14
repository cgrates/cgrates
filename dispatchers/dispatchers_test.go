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
	"testing"
	"time"

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
	rpcCl := map[string]chan birpc.ClientConnector{}
	connMng := engine.NewConnManager(cfg, rpcCl)
	dm := engine.NewDataManager(&engine.DataDBMock{
		GetKeysForPrefixF: func(*context.Context, string) ([]string, error) {
			return []string{"dpp_cgrates.org:123"}, nil
		},
	}, nil, connMng)
	dsp := &engine.DispatcherProfile{
		Tenant:         "cgrates.org",
		ID:             "321",
		Subsystems:     []string{utils.MetaAccounts},
		FilterIDs:      []string{"filter"},
		Strategy:       "",
		StrategyParams: nil,
		Weight:         0,
		Hosts:          nil,
	}
	err := dm.SetDispatcherProfile(dsp, false)
	if err == nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrNotImplemented, err)
	}
	fltr := &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "filter",
		Rules:  nil,
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(1999, 2, 3, 4, 5, 6, 700000000, time.UTC),
			ExpiryTime:     time.Date(2000, 2, 3, 4, 5, 6, 700000000, time.UTC),
		},
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
	_, err = dss.dispatcherProfileForEvent(tnt, ev, subsys)
	expected := utils.ErrNotImplemented
	if err == nil || err != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestDispatcherServiceDispatcherProfileForEventMIIDENotFound(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.DispatcherSCfg().IndexedSelects = false
	rpcCl := map[string]chan birpc.ClientConnector{}
	connMng := engine.NewConnManager(cfg, rpcCl)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMng)
	dss := NewDispatcherService(dm, cfg, nil, connMng)
	ev := &utils.CGREvent{}
	tnt := ""
	subsys := utils.IfaceAsString(ev.APIOpts[utils.Subsys])
	_, err := dss.dispatcherProfileForEvent(tnt, ev, subsys)
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
	dPfl := &engine.DispatcherProfile{}
	err := dsp.V1GetProfileForEvent(ev, dPfl)
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
	dPfl := &engine.DispatcherProfile{}
	err := dsp.V1GetProfileForEvent(ev, dPfl)
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
	err := dsp.Dispatch(ev, "", "", "", "")
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
				ID:          "",
				Address:     "error",
				Transport:   "",
				Synchronous: false,
				TLS:         false,
			},
		},
	}
	connMng := engine.NewConnManager(cfg, nil)
	dsp := NewDispatcherService(nil, cfg, nil, connMng)
	err := dsp.authorize("", "cgrates.org", utils.APIMethods, nil)
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
				ID:          "",
				Address:     "error",
				Transport:   "",
				Synchronous: false,
				TLS:         false,
			},
		},
	}
	connMng := engine.NewConnManager(cfg, nil)
	dsp := NewDispatcherService(nil, cfg, nil, connMng)
	err := dsp.authorize("", "cgrates.org", utils.APIMethods, nil)
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
	value := &engine.DispatcherHost{
		Tenant: "testTenant",
		RemoteHost: &config.RemoteHost{
			ID:          "testID",
			Address:     "",
			Transport:   "",
			Synchronous: false,
			TLS:         false,
		},
	}
	engine.Cache.SetWithoutReplicate(utils.CacheRPCConnections, "connID",
		value, nil, true, utils.NonTransactional)
	err := dsp.authorizeEvent(ev, reply)
	expected := "dial tcp: missing address"
	if err == nil || err.Error() != expected {
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
	rpcInt := map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes): chanRPC,
	}
	connMgr := engine.NewConnManager(cfg, rpcInt)

	dsp := NewDispatcherService(dm, cfg, nil, connMgr)
	ev := &utils.CGREvent{
		Tenant:  "testTenant",
		ID:      "testID",
		Time:    nil,
		Event:   map[string]interface{}{},
		APIOpts: nil,
	}
	value := &engine.DispatcherHost{
		Tenant: "testTenant",
		RemoteHost: &config.RemoteHost{
			ID:          "testID",
			Address:     rpcclient.InternalRPC,
			Transport:   utils.MetaInternal,
			Synchronous: false,
			TLS:         false,
		},
	}
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	engine.Cache.SetWithoutReplicate(utils.CacheRPCConnections, "testID",
		value, nil, true, utils.NonTransactional)
	rply := &engine.AttrSProcessEventReply{}
	err := dsp.authorizeEvent(ev, rply)
	if err != nil {
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
			Time:   nil,
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
	rpcInt := map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes): chanRPC,
	}
	connMgr := engine.NewConnManager(cfg, rpcInt)

	dsp := NewDispatcherService(dm, cfg, nil, connMgr)
	value := &engine.DispatcherHost{
		Tenant: "testTenant",
		RemoteHost: &config.RemoteHost{
			ID:          "testID",
			Address:     rpcclient.InternalRPC,
			Transport:   utils.MetaInternal,
			Synchronous: false,
			TLS:         false,
		},
	}
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	engine.Cache.SetWithoutReplicate(utils.CacheRPCConnections, "testID",
		value, nil, true, utils.NonTransactional)
	err := dsp.authorize(utils.APIMethods, "testTenant", "apikey", &time.Time{})
	expected := "UNAUTHORIZED_API"
	if err == nil || err.Error() != expected {
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
			Time:    nil,
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
	rpcInt := map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes): chanRPC,
	}
	connMgr := engine.NewConnManager(cfg, rpcInt)

	dsp := NewDispatcherService(dm, cfg, nil, connMgr)
	value := &engine.DispatcherHost{
		Tenant: "testTenant",
		RemoteHost: &config.RemoteHost{
			ID:          "testID",
			Address:     rpcclient.InternalRPC,
			Transport:   utils.MetaInternal,
			Synchronous: false,
			TLS:         false,
		},
	}
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	engine.Cache.SetWithoutReplicate(utils.CacheRPCConnections, "testID",
		value, nil, true, utils.NonTransactional)
	err := dsp.authorize(utils.APIMethods, "testTenant", "apikey", &time.Time{})
	expected := "NOT_FOUND"
	if err == nil || err.Error() != expected {
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
			Time:   nil,
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
	rpcInt := map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes): chanRPC,
	}
	connMgr := engine.NewConnManager(cfg, rpcInt)

	dsp := NewDispatcherService(dm, cfg, nil, connMgr)
	value := &engine.DispatcherHost{
		Tenant: "testTenant",
		RemoteHost: &config.RemoteHost{
			ID:          "testID",
			Address:     rpcclient.InternalRPC,
			Transport:   utils.MetaInternal,
			Synchronous: false,
			TLS:         false,
		},
	}
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	engine.Cache.SetWithoutReplicate(utils.CacheRPCConnections, "testID",
		value, nil, true, utils.NonTransactional)
	err := dsp.authorize("testMethod", "testTenant", "apikey", &time.Time{})
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	engine.Cache = cacheInit
}

func TestDispatcherServiceDispatcherProfileForEventErrNil(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.DispatcherSCfg().IndexedSelects = false
	rpcCl := map[string]chan birpc.ClientConnector{}
	connMng := engine.NewConnManager(cfg, rpcCl)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMng)
	dsp := &engine.DispatcherProfile{
		Tenant:         "cgrates.org",
		ID:             "123",
		Subsystems:     []string{utils.MetaAccounts},
		FilterIDs:      nil,
		Strategy:       "",
		StrategyParams: nil,
		Weight:         0,
		Hosts:          nil,
	}
	err := dm.SetDispatcherProfile(dsp, false)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	dss := NewDispatcherService(dm, cfg, nil, connMng)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "123",
		Time:   nil,
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
	_, err = dss.dispatcherProfileForEvent(tnt, ev, subsys)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
}

func TestDispatcherV1GetProfileForEventReturn(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.DispatcherSCfg().IndexedSelects = false
	rpcCl := map[string]chan birpc.ClientConnector{}
	connMng := engine.NewConnManager(cfg, rpcCl)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMng)
	dsp := &engine.DispatcherProfile{
		Tenant:         "cgrates.org",
		ID:             "123",
		Subsystems:     []string{utils.MetaAccounts},
		FilterIDs:      nil,
		Strategy:       "",
		StrategyParams: nil,
		Weight:         0,
		Hosts:          nil,
	}
	err := dm.SetDispatcherProfile(dsp, false)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	dss := NewDispatcherService(dm, cfg, nil, connMng)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "123",
		Time:   nil,
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
	_, err = dss.dispatcherProfileForEvent(tnt, ev, subsys)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	dPfl := &engine.DispatcherProfile{}
	err = dss.V1GetProfileForEvent(ev, dPfl)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestDispatcherServiceDispatcherProfileForEventErrNotFound(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.DispatcherSCfg().IndexedSelects = false
	rpcCl := map[string]chan birpc.ClientConnector{}
	connMng := engine.NewConnManager(cfg, rpcCl)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMng)
	dsp := &engine.DispatcherProfile{
		Tenant:         "cgrates.org",
		ID:             "123",
		Subsystems:     []string{utils.MetaAccounts},
		FilterIDs:      nil,
		Strategy:       "",
		StrategyParams: nil,
		Weight:         0,
		Hosts:          nil,
	}
	err := dm.SetDispatcherProfile(dsp, false)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	dss := NewDispatcherService(dm, cfg, nil, connMng)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "123",
		Time:   nil,
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
	_, err = dss.dispatcherProfileForEvent(tnt, ev, subsys)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrNotFound, err)
	}
}

func TestDispatcherServiceDispatcherProfileForEventErrNotFound2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.DispatcherSCfg().IndexedSelects = false
	rpcCl := map[string]chan birpc.ClientConnector{}
	connMng := engine.NewConnManager(cfg, rpcCl)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMng)
	dsp := &engine.DispatcherProfile{
		Tenant:         "cgrates.org",
		ID:             "123",
		Subsystems:     []string{utils.MetaAccounts},
		FilterIDs:      nil,
		Strategy:       "",
		StrategyParams: nil,
		Weight:         0,
		Hosts:          nil,
	}
	err := dm.SetDispatcherProfile(dsp, false)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	dss := NewDispatcherService(dm, cfg, nil, connMng)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "123",
		Time:   nil,
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
	_, err = dss.dispatcherProfileForEvent(tnt, ev, subsys)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrNotFound, err)
	}
}

func TestDispatcherServiceDispatcherProfileForEventErrNotFoundFilter(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.DispatcherSCfg().IndexedSelects = false
	rpcCl := map[string]chan birpc.ClientConnector{}
	connMng := engine.NewConnManager(cfg, rpcCl)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMng)
	dsp := &engine.DispatcherProfile{
		Tenant:         "cgrates.org",
		ID:             "123",
		Subsystems:     []string{utils.MetaAccounts},
		FilterIDs:      []string{"filter"},
		Strategy:       "",
		StrategyParams: nil,
		Weight:         0,
		Hosts:          nil,
	}
	err := dm.SetDispatcherProfile(dsp, false)
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
	_, err = dss.dispatcherProfileForEvent(tnt, ev, subsys)
	if err == nil || err.Error() != "NOT_FOUND:filter" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "NOT_FOUND:filter", err)
	}
}

func TestDispatcherServiceDispatchDspErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.DispatcherSCfg().IndexedSelects = false
	rpcCl := map[string]chan birpc.ClientConnector{}
	connMng := engine.NewConnManager(cfg, rpcCl)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMng)
	dsp := &engine.DispatcherProfile{
		Tenant:         "cgrates.org",
		ID:             "123",
		Subsystems:     []string{utils.MetaAccounts},
		FilterIDs:      nil,
		Strategy:       "",
		StrategyParams: nil,
		Weight:         0,
		Hosts:          nil,
	}
	err := dm.SetDispatcherProfile(dsp, false)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	dss := NewDispatcherService(dm, cfg, nil, connMng)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "123",
		Time:   nil,
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
	err = dss.Dispatch(ev, subsys, "", "", "")
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
	rpcCl := map[string]chan birpc.ClientConnector{}
	connMng := engine.NewConnManager(cfg, rpcCl)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMng)
	dsp := &engine.DispatcherProfile{
		Tenant:         "cgrates.org",
		ID:             "123",
		Subsystems:     []string{utils.MetaAccounts},
		FilterIDs:      nil,
		StrategyParams: make(map[string]interface{}),
		Strategy:       utils.MetaWeight,
		Weight:         0,
		Hosts:          nil,
	}
	newCache := engine.NewCacheS(cfg, dm, nil)
	value, errDsp := newDispatcher(dm, dsp)
	if errDsp != nil {
		t.Fatal(errDsp)
	}
	ctx := &context.Context{}
	engine.Cache = newCache
	engine.Cache.Set(ctx, utils.CacheDispatchers, dsp.TenantID(), value, nil, true, utils.EmptyString)

	err := dm.SetDispatcherProfile(dsp, false)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	dss := NewDispatcherService(dm, cfg, nil, connMng)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "123",
		Time:   nil,
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
	err = dss.Dispatch(ev, subsys, "", "", "")
	expected := "HOST_NOT_FOUND"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
	engine.Cache = cacheInit
}

func TestDispatcherServiceDispatcherProfileForEventFoundFilter(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.DispatcherSCfg().IndexedSelects = false
	rpcCl := map[string]chan birpc.ClientConnector{}
	connMng := engine.NewConnManager(cfg, rpcCl)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMng)
	dsp := &engine.DispatcherProfile{
		Tenant:         "cgrates.org",
		ID:             "123",
		Subsystems:     []string{utils.MetaAccounts},
		FilterIDs:      []string{"filter"},
		Strategy:       "",
		StrategyParams: nil,
		Weight:         0,
		Hosts:          nil,
	}
	err := dm.SetDispatcherProfile(dsp, false)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	fltr := &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "filter",
		Rules:  nil,
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(1999, 2, 3, 4, 5, 6, 700000000, time.UTC),
			ExpiryTime:     time.Date(2000, 2, 3, 4, 5, 6, 700000000, time.UTC),
		},
	}
	err = dm.SetFilter(context.Background(), fltr, false)
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
	_, err = dss.dispatcherProfileForEvent(tnt, ev, subsys)
	if err == nil || err.Error() != "NOT_FOUND" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "NOT_FOUND:filter", err)
	}
}

func TestDispatcherServiceDispatcherProfileForEventNotNotFound(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.DispatcherSCfg().IndexedSelects = true
	rpcCl := map[string]chan birpc.ClientConnector{}
	connMng := engine.NewConnManager(cfg, rpcCl)
	var cnt int

	dm := engine.NewDataManager(&engine.DataDBMock{
		GetIndexesDrvF: func(idxItmType, tntCtx, idxKey string) (indexes map[string]utils.StringSet, err error) {
			if cnt == 0 {
				cnt++
				return map[string]utils.StringSet{
					idxKey: utils.StringSet{"cgrates.org:dsp1": {}},
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
	_, err := dss.dispatcherProfileForEvent(tnt, ev, subsys)
	expected := utils.ErrNotImplemented
	if err == nil || err != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestDispatcherServiceDispatcherProfileForEventGetDispatcherError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.DispatcherSCfg().IndexedSelects = false
	rpcCl := map[string]chan birpc.ClientConnector{}
	connMng := engine.NewConnManager(cfg, rpcCl)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMng)
	dsp := &engine.DispatcherProfile{
		Tenant:         "cgrates.org",
		ID:             "123",
		Subsystems:     []string{utils.MetaAccounts},
		FilterIDs:      []string{"filter"},
		Strategy:       "",
		StrategyParams: nil,
		Weight:         0,
		Hosts:          nil,
	}
	err := dm.SetDispatcherProfile(dsp, false)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	fltr := &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "filter",
		Rules:  nil,
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(1999, 2, 3, 4, 5, 6, 700000000, time.UTC),
			ExpiryTime:     time.Date(2000, 2, 3, 4, 5, 6, 700000000, time.UTC),
		},
	}
	err = dm.SetFilter(context.Background(), fltr, false)
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
	_, err = dss.dispatcherProfileForEvent(tnt, ev, subsys)
	if err == nil || err.Error() != "NOT_FOUND" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "NOT_FOUND:filter", err)
	}
}

func TestDispatcherServiceDispatchDspErrHostNotFound2(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.DispatcherSCfg().IndexedSelects = false
	rpcCl := map[string]chan birpc.ClientConnector{}
	connMng := engine.NewConnManager(cfg, rpcCl)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMng)
	dsp := &engine.DispatcherProfile{
		Tenant:         "cgrates.org",
		ID:             "123",
		Subsystems:     []string{utils.MetaAccounts},
		FilterIDs:      nil,
		StrategyParams: make(map[string]interface{}),
		Strategy:       utils.MetaWeight,
		Weight:         0,
		Hosts:          nil,
	}
	newCache := engine.NewCacheS(cfg, dm, nil)
	value, errDsp := newDispatcher(dm, dsp)
	if errDsp != nil {
		t.Fatal(errDsp)
	}
	ctx := &context.Context{}
	engine.Cache = newCache
	engine.Cache.Set(ctx, utils.CacheDispatchers, dsp.TenantID(), value, nil, true, utils.EmptyString)

	err := dm.SetDispatcherProfile(dsp, false)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	dss := NewDispatcherService(dm, cfg, nil, connMng)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "123",
		Time:   nil,
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
	err = dss.Dispatch(ev, subsys, "", "", "")
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
	rpcInt := map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator): chanRPC,
	}
	connMgr := engine.NewConnManager(cfg, rpcInt)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	dsp := &engine.DispatcherProfile{
		Tenant:         "cgrates.org",
		ID:             "123",
		Subsystems:     []string{utils.MetaAccounts},
		FilterIDs:      nil,
		StrategyParams: make(map[string]interface{}),
		Strategy:       utils.MetaWeight,
		Weight:         0,
		Hosts:          nil,
	}
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache

	err := dm.SetDispatcherProfile(dsp, false)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	dss := NewDispatcherService(dm, cfg, nil, connMgr)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "123",
		Time:   nil,
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
	err = dss.Dispatch(ev, subsys, "", "", "")
	expected := "DISPATCHER_ERROR:NOT_IMPLEMENTED"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
	engine.Cache = cacheInit
}

func (dS *DispatcherService) DispatcherServiceTest(ev *utils.CGREvent, reply *string) (error, interface{}) {
	*reply = utils.Pong
	return nil, nil
}
