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
	"encoding/json"
	"fmt"

	"github.com/cgrates/cgrates/agents"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

// actionTarget returns the target attached to an action
func actionTarget(act string) (trgt string) {
	switch act {
	default:
		trgt = utils.MetaNone
	}
	return
}

func newScheduledActs(tenant, apID, trgTyp, trgID, schedule string,
	ctx context.Context, data utils.MapStorage, acts []actioner) (sActs *scheduledActs) {
	return &scheduledActs{tenant, apID, trgTyp, trgID, schedule, ctx, data, acts,
		ltcache.NewTransCache(map[string]*ltcache.CacheConfig{})}
}

// scheduled is a set of actions which will be executed directly or by the cron.schedule
type scheduledActs struct {
	tenant, apID, trgTyp, trgID string
	schedule                    string
	ctx                         context.Context
	data                        utils.MapStorage
	acts                        []actioner

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
		if err := act.execute(s.ctx, s.data); err != nil {
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
	connMgr *engine.ConnManager, aCfgs []*engine.APAction) (acts []actioner, err error) {
	acts = make([]actioner, len(aCfgs))
	for i, aCfg := range aCfgs {
		if acts[i], err = newActioner(cfg, fltrS, dm, connMgr, aCfg); err != nil {
			return nil, err
		}
	}
	return
}

// newAction is the constructor to create actioner
func newActioner(cfg *config.CGRConfig, fltrS *engine.FilterS, dm *engine.DataManager,
	connMgr *engine.ConnManager, aCfg *engine.APAction) (act actioner, err error) {
	switch aCfg.Type {
	case utils.MetaLog:
		return &actLog{aCfg}, nil
	case utils.CDRLog:
		return &actCDRLog{config: cfg, connMgr: connMgr, aCfg: aCfg, filterS: fltrS}, nil
	default:
		return nil, fmt.Errorf("unsupported action type: <%s>", aCfg.Type)

	}
}

// actioner is implemented by each action type
type actioner interface {
	id() string
	cfg() *engine.APAction
	execute(ctx context.Context, data utils.MapStorage) (err error)
}

// actLogger will log data to CGRateS logger
type actLog struct {
	aCfg *engine.APAction
}

func (aL *actLog) id() string {
	return aL.aCfg.ID
}

func (aL *actLog) cfg() *engine.APAction {
	return aL.aCfg
}

// execute implements actioner interface
func (aL *actLog) execute(ctx context.Context, data utils.MapStorage) (err error) {
	var body []byte
	if body, err = json.Marshal(data); err != nil {
		return
	}
	utils.Logger.Info(fmt.Sprintf("LOG Event: %s", body))
	return
}

// actCDRLog will log data to CGRateS logger
type actCDRLog struct {
	config  *config.CGRConfig
	filterS *engine.FilterS
	connMgr *engine.ConnManager
	aCfg    *engine.APAction
}

func (aL *actCDRLog) id() string {
	return aL.aCfg.ID
}

func (aL *actCDRLog) cfg() *engine.APAction {
	return aL.aCfg
}

// execute implements actioner interface
func (aL *actCDRLog) execute(ctx context.Context, data utils.MapStorage) (err error) {
	if len(aL.config.ActionSCfg().CDRsConns) == 0 {
		return fmt.Errorf("no connection with CDR Server")
	}
	template := aL.config.TemplatesCfg()[utils.MetaCdrLog]
	if id, has := aL.cfg().Opts[utils.MetaTemplateID]; has { // if templateID is not present we use default template
		template = aL.config.TemplatesCfg()[utils.IfaceAsString(id)]
	}
	// split the data into Request and Opts to send as parameters to AgentRequest
	req := data[utils.MetaReq].(map[string]interface{})
	reqNm := utils.MapStorage{}
	for key, val := range req {
		reqNm[key] = val
	}

	opts := data[utils.MetaOpts].(map[string]interface{})
	optsNm := utils.NewOrderedNavigableMap()
	for key, val := range opts {
		optsNm.Set(utils.NewFullPath(key, utils.NestingSep), utils.NewNMData(val))
	}

	// construct an AgentRequest so we can build the reply and send it to CDRServer
	agReq := agents.NewAgentRequest(reqNm, nil, nil, nil, optsNm, nil, aL.config.GeneralCfg().DefaultTenant, aL.config.GeneralCfg().DefaultTimezone, aL.filterS, nil, nil)

	if err = agReq.SetFields(template); err != nil {
		return
	}
	var rply string
	if err := aL.connMgr.Call(config.CgrConfig().SchedulerCfg().CDRsConns, nil,
		utils.CDRsV1ProcessEvent,
		&engine.ArgV1ProcessEvent{
			Flags:    []string{utils.ConcatenatedKey(utils.MetaChargers, "false")}, // do not try to get the chargers for cdrlog
			CGREvent: *config.NMAsCGREvent(agReq.Reply, agReq.Tenant, utils.NestingSep, agReq.Opts),
		}, &rply); err != nil {
		return err
	}

	return
}
