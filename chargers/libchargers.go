// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	connsCfg map[string][]*config.DynamicConns, connMgr *engine.ConnManager, cache *engine.CacheS, subsys string,
	cgrEv *utils.CGREvent) ([]*ChrgSProcessEventReply, error) {
	conns, err := engine.GetConnIDs(ctx, connsCfg, utils.MetaChargers,
		cgrEv.Tenant, cgrEv.AsDataProvider(), nil, fltrS)
	if err != nil {
		return nil, err

	}
	if len(conns) == 0 {
		return nil, utils.NewErrNotConnected(utils.ChargerS)
	}
	if x, ok := cache.Get(utils.CacheEventCharges, cgrEv.ID); ok && x != nil {
		return x.([]*ChrgSProcessEventReply), nil
	}
	var chrgrs []*ChrgSProcessEventReply
	if err = connMgr.Call(ctx, conns,
		utils.ChargerSv1ProcessEvent, cgrEv, &chrgrs); err != nil {
		err = utils.NewErrChargerS(err)
	}

	if errCh := cache.Set(ctx, utils.CacheEventCharges, cgrEv.ID, chrgrs, nil,
		true, utils.NonTransactional); errCh != nil {
		return nil, errCh
	}
	return chrgrs, err
}
