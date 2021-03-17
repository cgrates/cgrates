/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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

package actions

import (
	"context"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type actResetStat struct {
	tnt     string
	config  *config.CGRConfig
	connMgr *engine.ConnManager
	aCfg    *engine.APAction
}

func (aL *actResetStat) id() string {
	return aL.aCfg.ID
}

func (aL *actResetStat) cfg() *engine.APAction {
	return aL.aCfg
}

// execute implements actioner interface
func (aL *actResetStat) execute(_ context.Context, data utils.MapStorage, trgID string) (err error) {
	args := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID(trgID),
		APIOpts:  data[utils.MetaOpts].(map[string]interface{}),
	}
	if args.Tenant == utils.EmptyString { // in case that user pass only ID we populate the tenant from the event
		args.Tenant = aL.tnt
	}
	var rply string
	return aL.connMgr.Call(aL.config.ActionSCfg().StatSConns, nil,
		utils.StatSv1ResetStatQueue, args, &rply)
}

type actResetThreshold struct {
	tnt     string
	config  *config.CGRConfig
	connMgr *engine.ConnManager
	aCfg    *engine.APAction
}

func (aL *actResetThreshold) id() string {
	return aL.aCfg.ID
}

func (aL *actResetThreshold) cfg() *engine.APAction {
	return aL.aCfg
}

// execute implements actioner interface
func (aL *actResetThreshold) execute(_ context.Context, data utils.MapStorage, trgID string) (err error) {
	args := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID(trgID),
		APIOpts:  data[utils.MetaOpts].(map[string]interface{}),
	}
	if args.Tenant == utils.EmptyString { // in case that user pass only ID we populate the tenant from the event
		args.Tenant = aL.tnt
	}
	var rply string
	return aL.connMgr.Call(aL.config.ActionSCfg().ThresholdSConns, nil,
		utils.ThresholdSv1ResetThreshold, args, &rply)
}
