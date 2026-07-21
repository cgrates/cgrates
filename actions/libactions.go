// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package actions

import (
	"fmt"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

// actionTarget returns the target attached to an action
func actionTarget(act string) string {
	switch act {
	case utils.MetaResetStatQueue, utils.MetaDynamicStats:
		return utils.MetaStats
	case utils.MetaResetThreshold, utils.MetaDynamicThreshold:
		return utils.MetaThresholds
	case utils.MetaAddBalance, utils.MetaSetBalance, utils.MetaRemBalance:
		return utils.MetaAccounts
	case utils.MetaDynamicAttribute:
		return utils.MetaAttributes
	case utils.MetaDynamicResource:
		return utils.MetaResources
	case utils.MetaDynamicTrend:
		return utils.MetaTrends
	case utils.MetaDynamicRanking:
		return utils.MetaRankings
	case utils.MetaDynamicFilter:
		return utils.MetaFilters
	case utils.MetaDynamicRoute:
		return utils.MetaRoutes
	case utils.MetaDynamicRate:
		return utils.MetaRates
	case utils.MetaDynamicIP:
		return utils.MetaIPs
	case utils.MetaDynamicAction:
		return utils.MetaActions
	default:
		return utils.MetaNone
	}
}

func newScheduledActs(ctx *context.Context, tenant, apID, trgTyp, trgID, schedule string,
	ignFilters bool, data utils.MapStorage, acts []actioner) (sActs *scheduledActs) {
	return &scheduledActs{
		tenant:     tenant,
		apID:       apID,
		trgTyp:     trgTyp,
		trgID:      trgID,
		schedule:   schedule,
		ignFilters: ignFilters,
		data:       data,
		acts:       acts,
		cch:        ltcache.NewTransCache(map[string]*ltcache.CacheConfig{}, false),
	}
}

// scheduled is a set of actions which will be executed directly or by the cron.schedule
type scheduledActs struct {
	tenant     string
	apID       string
	trgTyp     string
	trgID      string
	schedule   string
	ignFilters bool
	data       utils.MapStorage
	acts       []actioner

	cch *ltcache.TransCache // cache data between actions here
}

// Execute notifies possible errors on execution
func (s *scheduledActs) Execute(ctx *context.Context) (err error) {
	var partExec bool
	for _, act := range s.acts {
		//ctx, cancel := context.WithTimeout(s.ctx, act.cfg().TTL)
		if err := act.execute(ctx, s.data, s.trgID); err != nil {
			utils.Logger.Warning(fmt.Sprintf("executing action: <%s>, error: <%s>", act.id(), err))
			partExec = true
		}
	}
	if partExec {
		err = utils.ErrPartiallyExecuted
	}
	return
}

// newActionersFromActions constructs multiple actioners out of APAction configurations
func newActionersFromActions(ctx *context.Context, cgrEv *utils.CGREvent, cfg *config.CGRConfig, cache *engine.CacheS, fltrS *engine.FilterS, dm *engine.DataManager,
	connMgr *engine.ConnManager, aCfgs []*utils.APAction, tnt string) (acts []actioner, err error) {
	acts = make([]actioner, len(aCfgs))
	for i, aCfg := range aCfgs {
		if acts[i], err = newActioner(ctx, cgrEv, cfg, cache, fltrS, dm, connMgr, aCfg, tnt); err != nil {
			return nil, err
		}
	}
	return acts, nil
}

// newAction is the constructor to create actioner
func newActioner(ctx *context.Context, cgrEv *utils.CGREvent, cfg *config.CGRConfig, cache *engine.CacheS, fltrS *engine.FilterS, dm *engine.DataManager,
	connMgr *engine.ConnManager, aCfg *utils.APAction, tnt string) (act actioner, err error) {
	switch aCfg.Type {
	case utils.MetaLog:
		return &actLog{aCfg}, nil
	case utils.CDRLog:
		return &actCDRLog{cfg, cache, fltrS, connMgr, aCfg}, nil
	case utils.MetaHTTPPost:
		return newActHTTPPost(ctx, tnt, cgrEv, cache, fltrS, cfg, aCfg)
	case utils.MetaExport:
		return &actExport{tnt, cfg, connMgr, fltrS, aCfg}, nil
	case utils.MetaResetStatQueue:
		return &actResetStat{tnt, cfg, connMgr, fltrS, aCfg}, nil
	case utils.MetaResetThreshold:
		return &actResetThreshold{tnt, cfg, fltrS, connMgr, aCfg}, nil
	case utils.MetaAddBalance:
		return &actSetBalance{cfg, connMgr, fltrS, aCfg, tnt, false}, nil
	case utils.MetaSetBalance:
		return &actSetBalance{cfg, connMgr, fltrS, aCfg, tnt, true}, nil
	case utils.MetaRemBalance:
		return &actRemBalance{cfg, connMgr, fltrS, aCfg, tnt}, nil
	case utils.MetaDynamicThreshold:
		return &actDynamicThreshold{cfg, connMgr, fltrS, aCfg, tnt, cgrEv}, nil
	case utils.MetaDynamicStats:
		return &actDynamicStats{cfg, connMgr, fltrS, aCfg, tnt, cgrEv}, nil
	case utils.MetaDynamicAttribute:
		return &actDynamicAttribute{cfg, connMgr, fltrS, aCfg, tnt, cgrEv}, nil
	case utils.MetaDynamicResource:
		return &actDynamicResource{cfg, connMgr, fltrS, aCfg, tnt, cgrEv}, nil
	case utils.MetaDynamicTrend:
		return &actDynamicTrend{cfg, connMgr, fltrS, aCfg, tnt, cgrEv}, nil
	case utils.MetaDynamicRanking:
		return &actDynamicRanking{cfg, connMgr, fltrS, aCfg, tnt, cgrEv}, nil
	case utils.MetaDynamicFilter:
		return &actDynamicFilter{cfg, connMgr, fltrS, aCfg, tnt, cgrEv}, nil
	case utils.MetaDynamicRoute:
		return &actDynamicRoute{cfg, connMgr, dm, fltrS, aCfg, tnt, cgrEv}, nil
	case utils.MetaDynamicRate:
		return &actDynamicRate{cfg, connMgr, fltrS, aCfg, tnt, cgrEv}, nil
	case utils.MetaDynamicIP:
		return &actDynamicIP{cfg, connMgr, fltrS, aCfg, tnt, cgrEv}, nil
	case utils.MetaDynamicAction:
		return &actDynamicAction{cfg, connMgr, fltrS, aCfg, tnt, cgrEv}, nil
	default:
		return nil, fmt.Errorf("unsupported action type: <%s>", aCfg.Type)

	}
}

// actioner is implemented by each action type
type actioner interface {
	id() string
	cfg() *utils.APAction
	execute(ctx *context.Context, data utils.MapStorage, trgID string) (err error)
}
