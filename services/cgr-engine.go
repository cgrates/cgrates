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

package services

import (
	"fmt"
	"sync"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

type CGREngine struct {
	cfg *config.CGRConfig

	srvManager *servmanager.ServiceManager
	srvDep     map[string]*sync.WaitGroup
	cmConns    map[string]chan birpc.ClientConnector
}

func (cgr *CGREngine) AddService(service servmanager.Service, connName, apiPrefix string,
	iConnCh chan birpc.ClientConnector) {
	cgr.srvManager.AddServices(service)
	cgr.cmConns[utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)] = iConnCh
	cgr.srvDep[service.ServiceName()] = new(sync.WaitGroup)
	engine.IntRPC.AddInternalRPCClient(apiPrefix, iConnCh)
}

func (cgr *CGREngine) InitConfigFromPath(path string, nodeID string) (err error) {
	// Init config
	if cgr.cfg, err = config.NewCGRConfigFromPath(path); err != nil {
		err = fmt.Errorf("could not parse config: <%s>", err)
		return
	}
	if cgr.cfg.ConfigDBCfg().Type != utils.MetaInternal {
		var d config.ConfigDB
		if d, err = engine.NewDataDBConn(cgr.cfg.ConfigDBCfg().Type,
			cgr.cfg.ConfigDBCfg().Host, cgr.cfg.ConfigDBCfg().Port,
			cgr.cfg.ConfigDBCfg().Name, cgr.cfg.ConfigDBCfg().User,
			cgr.cfg.ConfigDBCfg().Password, cgr.cfg.GeneralCfg().DBDataEncoding,
			cgr.cfg.ConfigDBCfg().Opts); err != nil { // Cannot configure getter database, show stopper
			err = fmt.Errorf("could not configure configDB: <%s>", err)
			return
		}
		if err = cgr.cfg.LoadFromDB(d); err != nil {
			err = fmt.Errorf("could not parse config from DB: <%s>", err)
			return
		}
	}
	if nodeID != utils.EmptyString {
		cgr.cfg.GeneralCfg().NodeID = nodeID
	}
	config.SetCgrConfig(cgr.cfg) // Share the config object
	return
}
