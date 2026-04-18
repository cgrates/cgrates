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

package attributes

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// cfg.SessionSCfg().Conns[utils.MetaAttributes]
// AttributeCProcessEvent is a wrapper to unify processing from the client side from multiple subsystems
func AttributeScProcessEvent(ctx *context.Context, fltrS *engine.FilterS,
	connsCfg []*config.DynamicConns, connMgr *engine.ConnManager, subsys string,
	cgrEv *utils.CGREvent) (reply *AttrSProcessEventReply, err error) {
	var conns []string
	if conns, err = engine.GetConnIDs(ctx, connsCfg,
		cgrEv.Tenant, cgrEv.AsDataProvider(), fltrS); err != nil {
		return
	} else if len(conns) == 0 {
		return nil, utils.NewErrNotConnected(utils.AttributeS)
	}
	if cgrEv.APIOpts == nil {
		cgrEv.APIOpts = make(engine.MapEvent)
	}
	cgrEv.APIOpts[utils.MetaSubsys] = subsys
	cgrEv.APIOpts[utils.OptsContext] = utils.FirstNonEmpty(
		utils.IfaceAsString(cgrEv.APIOpts[utils.OptsContext]),
		subsys)
	err = connMgr.Call(ctx, conns, utils.AttributeSv1ProcessEvent,
		cgrEv, reply)
	return
}
