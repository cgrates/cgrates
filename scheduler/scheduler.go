/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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
	"sync"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type Scheduler struct {
	queue       engine.ActionTimingPriorityList
	timer       *time.Timer
	restartLoop chan bool
	sync.Mutex
	storage          engine.RatingStorage
	schedulerStarted bool
}

func NewScheduler(storage engine.RatingStorage) *Scheduler {
	return &Scheduler{
		restartLoop: make(chan bool),
		storage:     storage,
	}
}

func (s *Scheduler) Loop() {
	s.schedulerStarted = true
	for {
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
			go a0.Execute()
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
				utils.Logger.Info(fmt.Sprintf("<Scheduler> Time for action on %v", a0.ActionsID))
			case <-s.restartLoop:
				// nothing to do, just continue the loop
			}
		}
	}
}

func (s *Scheduler) Reload(protect bool) {
	s.loadActionPlans()
	s.restart()
}

func (s *Scheduler) loadActionPlans() {
	s.Lock()
	defer s.Unlock()
	// limit the number of concurrent tasks
	var limit = make(chan bool, 10)
	// execute existing tasks
	for {
		task, err := s.storage.PopTask()
		if err != nil || task == nil {
			break
		}
		limit <- true
		go func() {
			task.Execute()
			<-limit
		}()
	}

	actionPlans, err := s.storage.GetAllActionPlans()
	if err != nil && err != utils.ErrNotFound {
		utils.Logger.Warning(fmt.Sprintf("<Scheduler> Cannot get action plans: %v", err))
	}
	utils.Logger.Info(fmt.Sprintf("<Scheduler> processing %d action plans", len(actionPlans)))
	// recreate the queue
	s.queue = engine.ActionTimingPriorityList{}
	for _, actionPlan := range actionPlans {
		for _, at := range actionPlan.ActionTimings {
			if at.Timing == nil {
				utils.Logger.Warning(fmt.Sprintf("<Scheduler> Nil timing on action plan: %+v, discarding!", at))
				continue
			}
			if at.IsASAP() {
				continue
			}

			now := time.Now()
			if at.GetNextStartTime(now).Before(now) {
				// the task is obsolete, do not add it to the queue
				continue
			}
			at.SetAccountIDs(actionPlan.AccountIDs) // copy the accounts
			s.queue = append(s.queue, at)

		}
	}
	sort.Sort(s.queue)
	utils.Logger.Info(fmt.Sprintf("<Scheduler> queued %d action plans", len(s.queue)))
}

func (s *Scheduler) restart() {
	if s.schedulerStarted {
		s.restartLoop <- true
	}
	if s.timer != nil {
		s.timer.Stop()
	}
}

func (s *Scheduler) GetQueue() engine.ActionTimingPriorityList {
	return s.queue
}
