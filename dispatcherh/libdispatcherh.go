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
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// Registar handdle for httpServer to register the dispatcher hosts
func Registar(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	result, errMessage := utils.OK, utils.EmptyString
	var err error
	var id *json.RawMessage
	if id, err = register(r); err != nil {
		result, errMessage = utils.EmptyString, err.Error()
	}
	if err := utils.WriteServerResponse(w, id, result, errMessage); err != nil {
		utils.Logger.Warning(fmt.Sprintf("<%s> Failed to write resonse because: %s",
			utils.DispatcherH, err))
	}
}

func register(req *http.Request) (*json.RawMessage, error) {
	sReq, err := utils.DecodeServerRequest(req.Body)
	if err != nil {
		utils.Logger.Warning(fmt.Sprintf("<%s> Failed to decode request because: %s",
			utils.DispatcherH, err))
		return nil, err
	}
	if sReq.Method != utils.DispatcherHv1RegisterHosts {
		err = errors.New("rpc: can't find service " + sReq.Method)
		utils.Logger.Warning(fmt.Sprintf("<%s> Failed to register hosts because: %s",
			utils.DispatcherH, err))
		return sReq.Id, err
	}
	var dHs []*engine.DispatcherHost
	params := []interface{}{dHs}
	if err = json.Unmarshal(*sReq.Params, &params); err != nil {
		utils.Logger.Warning(fmt.Sprintf("<%s> Failed to decode params because: %s",
			utils.DispatcherH, err))
		return sReq.Id, err
	}
	var addr string
	if addr, err = getIP(req); err != nil {
		utils.Logger.Warning(fmt.Sprintf("<%s> Failed to obtain the remote IP because: %s",
			utils.DispatcherH, err))
		return sReq.Id, err
	}

	for _, dH := range dHs {
		if len(dH.Conns) != 1 { // ignore the hosts with no connections or more
			continue
		}
		dH.Conns[0].Address = addr + dH.Conns[0].Address // the address contains the port
		if err = engine.Cache.Set(utils.CacheDispatcherHosts, dH.Tenant, dH, nil,
			false, utils.NonTransactional); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<%s> Failed to set DispatcherHost <%s> in cache because: %s",
				utils.DispatcherH, dH.TenantID(), err))
			continue
		}
	}
	return sReq.Id, nil
}

func getIP(r *http.Request) (ip string, err error) {
	ip = r.Header.Get("X-REAL-IP")
	if net.ParseIP(ip) != nil {
		return
	}
	for _, ip = range strings.Split(r.Header.Get("X-FORWARDED-FOR"), utils.FIELDS_SEP) {
		if net.ParseIP(ip) != nil {
			return
		}
	}
	if ip, _, err = net.SplitHostPort(r.RemoteAddr); err != nil {
		return
	}
	if net.ParseIP(ip) != nil {
		return
	}
	ip = utils.EmptyString
	err = fmt.Errorf("no valid ip found")
	return
}

func getConnCfg(cfg *config.CGRConfig, transport string, tmpl *config.RemoteHost) (conn *config.RemoteHost, err error) {
	var address string
	var extraPath string
	switch transport {
	case utils.MetaJSON:
		if tmpl.TLS {
			address = cfg.ListenCfg().RPCJSONTLSListen
		} else {
			address = cfg.ListenCfg().RPCJSONListen
		}
	case utils.MetaGOB:
		if tmpl.TLS {
			address = cfg.ListenCfg().RPCGOBTLSListen
		} else {
			address = cfg.ListenCfg().RPCGOBListen
		}
	case rpcclient.HTTPjson:
		if tmpl.TLS {
			address = cfg.ListenCfg().HTTPTLSListen
		} else {
			address = cfg.ListenCfg().HTTPListen
		}
		extraPath = cfg.HTTPCfg().HTTPJsonRPCURL
	}
	var port string
	if _, port, err = net.SplitHostPort(address); err != nil {
		return
	}
	conn = &config.RemoteHost{
		Address:     ":" + port + extraPath,
		Synchronous: tmpl.Synchronous,
		TLS:         tmpl.TLS,
		Transport:   transport,
	}
	return
}
