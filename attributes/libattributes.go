// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	connsCfg map[string][]*config.DynamicConns, connMgr *engine.ConnManager, subsys string,
	cgrEv *utils.CGREvent) (reply *ProcessEventReply, err error) {
	var conns []string
	if conns, err = engine.GetConnIDs(ctx, connsCfg, utils.MetaAttributes,
		cgrEv.Tenant, cgrEv.AsDataProvider(), nil, fltrS); err != nil {
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
	reply = &ProcessEventReply{}
	err = connMgr.Call(ctx, conns, utils.AttributeSv1ProcessEvent,
		cgrEv, reply)
	return
}
