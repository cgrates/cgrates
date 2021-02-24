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
	"fmt"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

// actionTarget returns the target attached to an action
func actionTarget(act string) string {
	switch act {
	case utils.MetaResetStatQueue:
		return utils.MetaStats
	case utils.MetaResetThreshold:
		return utils.MetaThresholds
	case utils.MetaAddBalance, utils.MetaSetBalance, utils.MetaRemBalance:
		return utils.MetaAccounts
	default:
		return utils.MetaNone
	}
}

func newScheduledActs(ctx context.Context, tenant, apID, trgTyp, trgID, schedule string,
	data utils.MapStorage, acts []actioner) (sActs *scheduledActs) {
	return &scheduledActs{
		tenant:   tenant,
		apID:     apID,
		trgTyp:   trgTyp,
		trgID:    trgID,
		schedule: schedule,
		ctx:      ctx,
		data:     data,
		acts:     acts,
		cch:      ltcache.NewTransCache(map[string]*ltcache.CacheConfig{}),
	}
}

// scheduled is a set of actions which will be executed directly or by the cron.schedule
type scheduledActs struct {
	tenant   string
	apID     string
	trgTyp   string
	trgID    string
	schedule string
	ctx      context.Context
	data     utils.MapStorage
	acts     []actioner

	cch *ltcache.TransCache // cache data between actions here
}

// Execute is called when we want the ActionProfile to be executed
func (s *scheduledActs) ScheduledExecute() {
	s.Execute()
}

// Execute notifies possible errors on execution
func (s *scheduledActs) Execute() (err error) {
	var partExec bool
	for _, act := range s.acts {
		//ctx, cancel := context.WithTimeout(s.ctx, act.cfg().TTL)
		if err := act.execute(s.ctx, s.data, s.trgID); err != nil {
			utils.Logger.Warning(fmt.Sprintf("executing action: <%s>, error: <%s>", act.id(), err))
			partExec = true
		}
	}
	// postexec here
	if partExec {
		err = utils.ErrPartiallyExecuted
	}
	return
}

// postExec will save data which was modified in actions and unlock guardian
func (s *scheduledActs) postExec() (err error) {
	return
}

// newActionersFromActions constructs multiple actioners out of APAction configurations
func newActionersFromActions(cfg *config.CGRConfig, fltrS *engine.FilterS, dm *engine.DataManager,
	connMgr *engine.ConnManager, aCfgs []*engine.APAction, tnt string) (acts []actioner, err error) {
	acts = make([]actioner, len(aCfgs))
	for i, aCfg := range aCfgs {
		if acts[i], err = newActioner(cfg, fltrS, dm, connMgr, aCfg, tnt); err != nil {
			return nil, err
		}
	}
	return
}

// newAction is the constructor to create actioner
func newActioner(cfg *config.CGRConfig, fltrS *engine.FilterS, dm *engine.DataManager,
	connMgr *engine.ConnManager, aCfg *engine.APAction, tnt string) (act actioner, err error) {
	switch aCfg.Type {
	case utils.MetaLog:
		return &actLog{aCfg}, nil
	case utils.CDRLog:
		return &actCDRLog{cfg, fltrS, connMgr, aCfg}, nil
	case utils.MetaHTTPPost:
		return &actHTTPPost{cfg, aCfg}, nil
	case utils.MetaExport:
		return &actExport{tnt, cfg, connMgr, aCfg}, nil
	case utils.MetaResetStatQueue:
		return &actResetStat{tnt, cfg, connMgr, aCfg}, nil
	case utils.MetaResetThreshold:
		return &actResetThreshold{tnt, cfg, connMgr, aCfg}, nil
	case utils.MetaAddBalance:
		return &actSetBalance{cfg, connMgr, aCfg, tnt, false}, nil
	case utils.MetaSetBalance:
		return &actSetBalance{cfg, connMgr, aCfg, tnt, true}, nil
	case utils.MetaRemBalance:
		return &actRemBalance{cfg, connMgr, aCfg, tnt}, nil
	default:
		return nil, fmt.Errorf("unsupported action type: <%s>", aCfg.Type)

	}
}

// actioner is implemented by each action type
type actioner interface {
	id() string
	cfg() *engine.APAction
	execute(ctx context.Context, data utils.MapStorage, trgID string) (err error)
}
