/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
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

package scheduler

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type Scheduler struct {
	sync.RWMutex
	queue                           engine.ActionTimingPriorityList
	timer                           *time.Timer
	restartLoop                     chan struct{}
	dm                              *engine.DataManager
	cfg                             *config.CGRConfig
	fltrS                           *engine.FilterS
	schedulerStarted                bool
	actStatsInterval                time.Duration                 // How long time to keep the stats in memory
	actSucessChan, actFailedChan    chan *engine.Action           // ActionPlan will pass actions via these channels
	aSMux, aFMux                    sync.RWMutex                  // protect schedStats
	actSuccessStats, actFailedStats map[string]map[time.Time]bool // keep here stats regarding executed actions, map[actionType]map[execTime]bool
}

func NewScheduler(dm *engine.DataManager, cfg *config.CGRConfig,
	fltrS *engine.FilterS) (s *Scheduler) {
	s = &Scheduler{
		restartLoop: make(chan struct{}),
		dm:          dm,
		cfg:         cfg,
		fltrS:       fltrS,
	}
	s.Reload()
	return
}

func (s *Scheduler) updateActStats(act *engine.Action, isFailed bool) {
	mux := &s.aSMux
	statsMp := s.actSuccessStats
	if isFailed {
		mux = &s.aFMux
		statsMp = s.actFailedStats
	}
	now := time.Now()
	mux.Lock()
	for aType := range statsMp {
		for t := range statsMp[aType] {
			if now.Sub(t) > s.actStatsInterval {
				delete(statsMp[aType], t)
				if len(statsMp[aType]) == 0 {
					delete(statsMp, aType)
				}
			}
		}
	}
	if act == nil {
		return
	}
	if _, hasIt := statsMp[act.ActionType]; !hasIt {
		statsMp[act.ActionType] = make(map[time.Time]bool)
	}
	statsMp[act.ActionType][now] = true
	mux.Unlock()
}

func (s *Scheduler) Loop() {
	s.schedulerStarted = true
	for {
		if !s.schedulerStarted { // shutdown requested
			break
		}
		for len(s.queue) == 0 { //hang here if empty
			<-s.restartLoop
		}
		utils.Logger.Info(fmt.Sprintf("<Scheduler> Scheduler queue length: %v", len(s.queue)))
		s.Lock()
		a0 := s.queue[0]
		utils.Logger.Info(fmt.Sprintf("<Scheduler> Action: %s", a0.ActionsID))
		now := time.Now()
		start := a0.GetNextStartTime(now)
		if start.Equal(now) || start.Before(now) {
			go a0.Execute(s.actSucessChan, s.actFailedChan)
			// if after execute the next start time is in the past then
			// do not add it to the queue
			a0.ResetStartTimeCache()
			now = time.Now().Add(time.Second)
			start = a0.GetNextStartTime(now)
			if start.Before(now) {
				s.queue = s.queue[1:]
			} else {
				s.queue = append(s.queue, a0)
				s.queue = s.queue[1:]
				sort.Sort(s.queue)
			}
			s.Unlock()
		} else {
			s.Unlock()
			d := a0.GetNextStartTime(now).Sub(now)
			utils.Logger.Info(fmt.Sprintf("<Scheduler> Time to next action (%s): %v", a0.ActionsID, d))
			s.timer = time.NewTimer(d)
			select {
			case <-s.timer.C:
				// timer has expired
				utils.Logger.Info(fmt.Sprintf("<Scheduler> Time for action on %s", a0.ActionsID))
			case <-s.restartLoop:
				// nothing to do, just continue the loop
			}
		}
	}
}

func (s *Scheduler) Reload() {
	s.loadActionPlans()
	s.restart()
}

// loadTasks loads the tasks
// this will push the tasks that did not match
// the filters before exiting the function
func (s *Scheduler) loadTasks() {
	// limit the number of concurrent tasks
	limit := make(chan struct{}, 10)
	// if some task don't mach the filter save them in this slice
	// in oreder to push them back when finish executing them
	var unexecutedTasks []*engine.Task
	// execute existing tasks
	for {
		task, err := s.dm.DataDB().PopTask()
		if err != nil || task == nil {
			break
		}
		if pass, err := s.fltrS.Pass(s.cfg.GeneralCfg().DefaultTenant,
			s.cfg.SchedulerCfg().Filters, task); err != nil || !pass {
			if err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: <%s> querying filters for path: <%+v>, not executing task <%s> on account <%s>",
						utils.SchedulerS, err.Error(), s.cfg.SchedulerCfg().Filters, task.ActionsID, task.AccountID))
			}
			// we do not push the task back as this may cause an infinite loop
			// push it when the function is done and we stopped the for
			// do not use defer here as the functions are exeucted
			// from the last one to the first
			unexecutedTasks = append(unexecutedTasks, task)
			continue
		}
		limit <- struct{}{}
		go func() {
			utils.Logger.Info(fmt.Sprintf("<%s> executing task %s on account %s",
				utils.SchedulerS, task.ActionsID, task.AccountID))
			task.Execute()
			<-limit
		}()
	}
	for _, t := range unexecutedTasks {
		if err := s.dm.DataDB().PushTask(t); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> failed pushing task <%s> back to DataDB, err <%s>",
					utils.SchedulerS, t.ActionsID, err.Error()))
		}
	}
}

func (s *Scheduler) loadActionPlans() {
	s.Lock()
	defer s.Unlock()
	s.loadTasks()

	actionPlans, err := s.dm.GetAllActionPlans()
	if err != nil && err != utils.ErrNotFound {
		utils.Logger.Warning(fmt.Sprintf("<Scheduler> Cannot get action plans: %v", err))
	}
	utils.Logger.Info(fmt.Sprintf("<Scheduler> processing %d action plans", len(actionPlans)))
	// recreate the queue
	s.queue = engine.ActionTimingPriorityList{}
	for _, actionPlan := range actionPlans {
		if actionPlan == nil {
			continue
		}
		for _, at := range actionPlan.ActionTimings {
			if at.Timing == nil {
				utils.Logger.Warning(fmt.Sprintf("<Scheduler> Nil timing on action plan: %+v, discarding!", at))
				continue
			}
			if at.IsASAP() {
				continue // should be already executed as task
			}
			now := time.Now()
			if at.GetNextStartTime(now).Before(now) {
				// the task is obsolete, do not add it to the queue
				continue
			}
			at.SetAccountIDs(actionPlan.AccountIDs) // copy the accounts
			at.SetActionPlanID(actionPlan.Id)
			for _, task := range at.Tasks() {
				if pass, err := s.fltrS.Pass(s.cfg.GeneralCfg().DefaultTenant,
					s.cfg.SchedulerCfg().Filters, task); err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> error: <%s> querying filters for path: <%+v>, not executing action <%s> on account <%s>",
							utils.SchedulerS, err.Error(), s.cfg.SchedulerCfg().Filters, task.ActionsID, task.AccountID))
					at.RemoveAccountID(task.AccountID)
				} else if !pass {
					at.RemoveAccountID(task.AccountID)
				}
			}
			s.queue = append(s.queue, at)
		}
	}
	sort.Sort(s.queue)
	utils.Logger.Info(fmt.Sprintf("<Scheduler> queued %d action plans", len(s.queue)))
}

func (s *Scheduler) restart() {
	if s.schedulerStarted {
		s.restartLoop <- struct{}{}
	}
	if s.timer != nil {
		s.timer.Stop()
	}
}

type ArgsGetScheduledActions struct {
	Tenant, Account    *string
	TimeStart, TimeEnd *time.Time // Filter based on next runTime
	utils.Paginator
}

type ScheduledAction struct {
	NextRunTime                               time.Time
	Accounts                                  int // Number of acccounts this action will run on
	ActionPlanID, ActionTimingUUID, ActionsID string
}

func (s *Scheduler) GetScheduledActions(fltr ArgsGetScheduledActions) (schedActions []*ScheduledAction) {
	s.RLock()
	for _, at := range s.queue {
		sas := &ScheduledAction{NextRunTime: at.GetNextStartTime(time.Now()), Accounts: len(at.GetAccountIDs()),
			ActionPlanID: at.GetActionPlanID(), ActionTimingUUID: at.Uuid, ActionsID: at.ActionsID}
		if fltr.TimeStart != nil && !fltr.TimeStart.IsZero() && sas.NextRunTime.Before(*fltr.TimeStart) {
			continue // need to match the filter interval
		}
		if fltr.TimeEnd != nil && !fltr.TimeEnd.IsZero() && (sas.NextRunTime.After(*fltr.TimeEnd) || sas.NextRunTime.Equal(*fltr.TimeEnd)) {
			continue
		}
		// filter on account
		if fltr.Tenant != nil || fltr.Account != nil {
			found := false
			for accID := range at.GetAccountIDs() {
				split := strings.Split(accID, utils.ConcatenatedKeySep)
				if len(split) != 2 {
					continue // malformed account id
				}
				if fltr.Tenant != nil && *fltr.Tenant != split[0] {
					continue
				}
				if fltr.Account != nil && *fltr.Account != split[1] {
					continue
				}
				found = true
				break
			}
			if !found {
				continue
			}
		}
		schedActions = append(schedActions, sas)
	}
	if fltr.Paginator.Offset != nil {
		if *fltr.Paginator.Offset <= len(schedActions) {
			schedActions = schedActions[*fltr.Paginator.Offset:]
		}
	}
	if fltr.Paginator.Limit != nil {
		if *fltr.Paginator.Limit <= len(schedActions) {
			schedActions = schedActions[:*fltr.Paginator.Limit]
		}
	}
	s.RUnlock()
	return
}

func (s *Scheduler) Shutdown() {
	s.schedulerStarted = false  // disable loop on next run
	s.restartLoop <- struct{}{} // cancel waiting tasks
	if s.timer != nil {
		s.timer.Stop()
	}
}
