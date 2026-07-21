// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package actions

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type actResetStat struct {
	tnt     string
	config  *config.CGRConfig
	connMgr *engine.ConnManager
	fltrS   *engine.FilterS
	aCfg    *utils.APAction
}

func (aL *actResetStat) id() string {
	return aL.aCfg.ID
}

func (aL *actResetStat) cfg() *utils.APAction {
	return aL.aCfg
}

// execute implements actioner interface
func (aL *actResetStat) execute(ctx *context.Context, data utils.MapStorage, trgID string) (err error) {
	args := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID(trgID),
		APIOpts:  data[utils.MetaOpts].(map[string]any),
	}
	if args.Tenant == utils.EmptyString { // in case that user pass only ID we populate the tenant from the event
		args.Tenant = aL.tnt
	}
	var statsConns []string
	if statsConns, err = engine.GetConnIDs(ctx, aL.config.ActionSCfg().Conns, utils.MetaStats, aL.tnt, data, nil, aL.fltrS); err != nil {
		return
	}
	var rply string
	return aL.connMgr.Call(ctx, statsConns,
		utils.StatSv1ResetStatQueue, args, &rply)
}

type actResetThreshold struct {
	tnt     string
	config  *config.CGRConfig
	fltrS   *engine.FilterS
	connMgr *engine.ConnManager
	aCfg    *utils.APAction
}

func (aL *actResetThreshold) id() string {
	return aL.aCfg.ID
}

func (aL *actResetThreshold) cfg() *utils.APAction {
	return aL.aCfg
}

// execute implements actioner interface
func (aL *actResetThreshold) execute(ctx *context.Context, data utils.MapStorage, trgID string) (err error) {
	args := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID(trgID),
		APIOpts:  data[utils.MetaOpts].(map[string]any),
	}
	if args.Tenant == utils.EmptyString { // in case that user pass only ID we populate the tenant from the event
		args.Tenant = aL.tnt
	}
	var threshConns []string
	if threshConns, err = engine.GetConnIDs(ctx, aL.config.ActionSCfg().Conns, utils.MetaThresholds, aL.tnt, data, nil, aL.fltrS); err != nil {
		return
	}
	var rply string
	return aL.connMgr.Call(ctx, threshConns,
		utils.ThresholdSv1ResetThreshold, args, &rply)
}
