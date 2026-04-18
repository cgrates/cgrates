/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package chargers

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// cfg.SessionSCfg().Conns[utils.MetaChargers]
// ChargerScProcessEvent is a wrapper to unify processing from the client side from multiple subsystems
func ChargerScProcessEvent(ctx *context.Context, fltrS *engine.FilterS,
	connsCfg []*config.DynamicConns, connMgr *engine.ConnManager, subsys string,
	cgrEv *utils.CGREvent) (chrgrs []*ChrgSProcessEventReply, err error) {
	var conns []string
	if conns, err = engine.GetConnIDs(ctx, connsCfg,
		cgrEv.Tenant, cgrEv.AsDataProvider(), fltrS); err != nil {
		return
	}
	if len(conns) == 0 {
		return nil, utils.NewErrNotConnected(utils.ChargerS)
	}
	if x, ok := engine.Cache.Get(utils.CacheEventCharges, cgrEv.ID); ok && x != nil {
		return x.([]*ChrgSProcessEventReply), nil
	}
	if err = connMgr.Call(ctx, conns,
		utils.ChargerSv1ProcessEvent, cgrEv, &chrgrs); err != nil {
		err = utils.NewErrChargerS(err)
	}

	if errCh := engine.Cache.Set(ctx, utils.CacheEventCharges, cgrEv.ID, chrgrs, nil,
		true, utils.NonTransactional); errCh != nil {
		return nil, errCh
	}
	return
}
