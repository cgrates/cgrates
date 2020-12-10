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
)

func newScheduledActs(tenant, apID, acntID string, ctx context.Context,
	data utils.RWDataProvider, acts []actioner) (sActs *scheduledActs) {
	return &scheduledActs{tenant, apID, acntID, ctx, data, acts}
}

// scheduled is a set of actions which will be executed directly or by the cron.schedule
type scheduledActs struct {
	tenant, apID, acntID string
	ctx                  context.Context
	data                 utils.RWDataProvider
	acts                 []actioner
}

// Execute is called when we want the ActionProfile to be executed
func (s *scheduledActs) ScheduledExecute() {
	s.Execute()
}

// Execute notifies possible errors on execution
func (s *scheduledActs) Execute() (err error) {
	var partExec bool
	for _, act := range s.acts {
		if err := act.execute(nil, s.data); err != nil {
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
func newActionersFromActions(cfg *config.CGRConfig, fltrS *engine.FilterS, dm *engine.DataManager,
	aCfgs []*engine.APAction) (acts []actioner, err error) {
	acts = make([]actioner, len(aCfgs))
	for i, aCfg := range aCfgs {
		if acts[i], err = newActioner(cfg, fltrS, dm, aCfg); err != nil {
			return nil, err
		}
	}
	return
}

// newAction is the constructor to create actioner
func newActioner(cfg *config.CGRConfig, fltrS *engine.FilterS, dm *engine.DataManager,
	aCfg *engine.APAction) (act actioner, err error) {
	switch aCfg.Type {
	case utils.LOG:
		return &actLog{aCfg}, nil
	default:
		return nil, fmt.Errorf("unsupported action type: <%s>", aCfg.Type)

	}
	return
}

// actioner is implemented by each action type
type actioner interface {
	id() string
	execute(ctx context.Context, data utils.RWDataProvider) (err error)
}

// actLogger will log data to CGRateS logger
type actLog struct {
	aCfg *engine.APAction
}

func (aL *actLog) id() string {
	return aL.aCfg.ID
}

// execute implements actioner interface
func (aL *actLog) execute(ctx context.Context, data utils.RWDataProvider) (err error) {
	return
}
