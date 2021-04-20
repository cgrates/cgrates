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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func (dS *DispatcherService) DispatcherServicePing(ev *utils.CGREvent, reply *string) error {
	*reply = utils.Pong
	return nil
}

func TestDispatcherCall1(t *testing.T) {
	dS := &DispatcherService{}
	var reply string
	if err := dS.Call(utils.DispatcherServicePing, &utils.CGREvent{}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Expected: %s , received: %s", utils.Pong, reply)
	}
}

func TestDispatcherCall2(t *testing.T) {
	dS := &DispatcherService{}
	var reply string
	if err := dS.Call("DispatcherServicePing", &utils.CGREvent{}, &reply); err == nil || err.Error() != rpcclient.ErrUnsupporteServiceMethod.Error() {
		t.Error(err)
	}
	if err := dS.Call("DispatcherService.Pong", &utils.CGREvent{}, &reply); err == nil || err.Error() != rpcclient.ErrUnsupporteServiceMethod.Error() {
		t.Error(err)
	}
	dS.Shutdown()
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
	expected := "DISPATCHER_ERROR:NOT_FOUND"
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
	expected := "DISPATCHER_ERROR:NOT_FOUND"
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
	expected := "DISPATCHER_ERROR:NOT_FOUND"
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

/*
func TestDispatcherServiceAuthorizeEventError3(t *testing.T) {
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
	tmp := engine.IntRPC
	engine.IntRPC = map[string]*rpcclient.RPCClient{}
	chanRPC := make(chan rpcclient.ClientConnector, 1)
	chanRPC <- new(mockTypeCon)
	engine.IntRPC.AddInternalRPCClient(utils.AttributeSv1Ping, chanRPC)
	engine.Cache.SetWithoutReplicate(utils.CacheRPCConnections, "connID",
		value, nil, true, utils.NonTransactional)
	err := dsp.authorizeEvent(ev, reply)
	expected := "dial tcp: missing address"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
	engine.Cache = cacheInit
	engine.IntRPC = tmp
}
*/
