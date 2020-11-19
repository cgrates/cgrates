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

package dispatcherh

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewRegisterArgs creates the arguments for register hosts API
func NewRegisterArgs(cfg *config.CGRConfig, tnt string, hostCfgs []*config.DispatcherHRegistarCfg) (rargs *RegisterArgs, err error) {
	rargs = &RegisterArgs{
		Tenant: tnt,
		Opts:   make(map[string]interface{}),
		Hosts:  make([]*RegisterHostCfg, len(hostCfgs)),
	}
	for i, hostCfg := range hostCfgs {
		var port string
		if port, err = getConnPort(cfg,
			hostCfg.RegisterTransport,
			hostCfg.RegisterTLS); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<%s> Unable to get the port because : %s",
				utils.DispatcherH, err))
			return
		}
		rargs.Hosts[i] = &RegisterHostCfg{
			ID:        hostCfg.ID,
			Port:      port,
			Transport: hostCfg.RegisterTransport,
			TLS:       hostCfg.RegisterTLS,
		}
	}
	return
}

// RegisterArgs the arguments to register the dispacher host
type RegisterArgs struct {
	Tenant string
	Opts   map[string]interface{}
	Hosts  []*RegisterHostCfg
}

// RegisterHostCfg the host config used to register
type RegisterHostCfg struct {
	ID        string
	Port      string
	Transport string
	TLS       bool
}

// AsDispatcherHosts converts the arguments to DispatcherHosts
func (rargs *RegisterArgs) AsDispatcherHosts(ip string) (dHs []*engine.DispatcherHost) {
	dHs = make([]*engine.DispatcherHost, len(rargs.Hosts))
	for i, hCfg := range rargs.Hosts {
		dHs[i] = hCfg.AsDispatcherHost(rargs.Tenant, ip)
	}
	return
}

// AsDispatcherHost converts the arguments to DispatcherHosts
func (rhc *RegisterHostCfg) AsDispatcherHost(tnt, ip string) *engine.DispatcherHost {
	return &engine.DispatcherHost{
		Tenant: tnt,
		ID:     rhc.ID,
		Conn: &config.RemoteHost{
			Address:   ip + ":" + rhc.Port,
			Transport: rhc.Transport,
			TLS:       rhc.TLS,
		},
	}
}

// NewUnregisterArgs creates the arguments for unregister hosts API
func NewUnregisterArgs(tnt string, hostCfgs []*config.DispatcherHRegistarCfg) (uargs *UnregisterArgs) {
	uargs = &UnregisterArgs{
		Tenant: tnt,
		Opts:   make(map[string]interface{}),
		IDs:    make([]string, len(hostCfgs)),
	}
	for i, hostCfg := range hostCfgs {
		uargs.IDs[i] = hostCfg.ID
	}
	return
}

// UnregisterArgs the arguments to unregister the dispacher host
type UnregisterArgs struct {
	Tenant string
	Opts   map[string]interface{}
	IDs    []string
}

// Registar handdle for httpServer to register the dispatcher hosts
func Registar(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	var result interface{} = utils.OK
	var errMessage interface{}
	var err error
	var id *json.RawMessage
	if id, err = register(r); err != nil {
		result, errMessage = nil, err.Error()
	}
	if err := utils.WriteServerResponse(w, id, result, errMessage); err != nil {
		utils.Logger.Warning(fmt.Sprintf("<%s> Failed to write resonse because: %s",
			utils.DispatcherH, err))
	}
}

func register(req *http.Request) (*json.RawMessage, error) {
	id := json.RawMessage("0")
	sReq, err := utils.DecodeServerRequest(req.Body)
	if err != nil {
		utils.Logger.Warning(fmt.Sprintf("<%s> Failed to decode request because: %s",
			utils.DispatcherH, err))
		return &id, err
	}
	var hasErrors bool
	switch sReq.Method {
	default:
		err = errors.New("rpc: can't find service " + sReq.Method)
		utils.Logger.Warning(fmt.Sprintf("<%s> Failed to register hosts because: %s",
			utils.DispatcherH, err))
		return sReq.Id, err
	case utils.DispatcherHv1UnregisterHosts:
		var args UnregisterArgs
		params := []interface{}{&args}
		if err = json.Unmarshal(*sReq.Params, &params); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<%s> Failed to decode params because: %s",
				utils.DispatcherH, err))
			return sReq.Id, err
		}
		for _, id := range args.IDs {
			if err = engine.Cache.Remove(utils.CacheDispatcherHosts, utils.ConcatenatedKey(args.Tenant, id), true, utils.NonTransactional); err != nil {
				utils.Logger.Warning(fmt.Sprintf("<%s> Failed to remove DispatcherHost <%s> from cache because: %s",
					utils.DispatcherH, id, err))
				hasErrors = true
				continue
			}
		}

	case utils.DispatcherHv1RegisterHosts:
		var dHs RegisterArgs
		params := []interface{}{&dHs}
		if err = json.Unmarshal(*sReq.Params, &params); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<%s> Failed to decode params because: %s",
				utils.DispatcherH, err))
			return sReq.Id, err
		}
		var addr string
		if addr, err = utils.GetRemoteIP(req); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<%s> Failed to obtain the remote IP because: %s",
				utils.DispatcherH, err))
			return sReq.Id, err
		}

		for _, dH := range dHs.AsDispatcherHosts(addr) {
			if err = engine.Cache.Set(utils.CacheDispatcherHosts, dH.TenantID(), dH, nil,
				true, utils.NonTransactional); err != nil {
				utils.Logger.Warning(fmt.Sprintf("<%s> Failed to set DispatcherHost <%s> in cache because: %s",
					utils.DispatcherH, dH.TenantID(), err))
				hasErrors = true
				continue
			}
		}
	}
	if hasErrors {
		return sReq.Id, utils.ErrPartiallyExecuted
	}
	return sReq.Id, nil
}

func getConnPort(cfg *config.CGRConfig, transport string, tls bool) (port string, err error) {
	var address string
	var extraPath string
	switch transport {
	case utils.MetaJSON:
		if tls {
			address = cfg.ListenCfg().RPCJSONTLSListen
		} else {
			address = cfg.ListenCfg().RPCJSONListen
		}
	case utils.MetaGOB:
		if tls {
			address = cfg.ListenCfg().RPCGOBTLSListen
		} else {
			address = cfg.ListenCfg().RPCGOBListen
		}
	case rpcclient.HTTPjson:
		if tls {
			address = cfg.ListenCfg().HTTPTLSListen
		} else {
			address = cfg.ListenCfg().HTTPListen
		}
		extraPath = cfg.HTTPCfg().HTTPJsonRPCURL
	}
	if _, port, err = net.SplitHostPort(address); err != nil {
		return
	}
	port += extraPath
	return
}
