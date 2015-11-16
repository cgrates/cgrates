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
	queue       engine.ActionPlanPriotityList
	timer       *time.Timer
	restartLoop chan bool
	sync.Mutex
}

func NewScheduler() *Scheduler {
	return &Scheduler{restartLoop: make(chan bool)}
}

func (s *Scheduler) Loop() {
	for {
		for len(s.queue) == 0 { //hang here if empty
			<-s.restartLoop
		}
		s.Lock()
		a0 := s.queue[0]
		//utils.Logger.Info(fmt.Sprintf("Scheduler qeue length: %v", len(s.qeue)))
		now := time.Now()
		start := a0.GetNextStartTime(now)
		if start.Equal(now) || start.Before(now) {
			go a0.Execute()
			// if after execute the next start time is in the past then
			// do not add it to the queue
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
			//utils.Logger.Info(fmt.Sprintf("Timer set to wait for %v", d))
			s.timer = time.NewTimer(d)
			select {
			case <-s.timer.C:
				// timer has expired
				utils.Logger.Info(fmt.Sprintf("Time for action on %v", a0))
			case <-s.restartLoop:
				// nothing to do, just continue the loop
			}
		}
	}
}

func (s *Scheduler) LoadActionPlans(storage engine.RatingStorage) {
	actionPlans, err := storage.GetAllActionPlans()
	if err != nil && err != utils.ErrNotFound {
		utils.Logger.Warning(fmt.Sprintf("Cannot get action plans: %v", err))
	}
	// recreate the queue
	s.Lock()
	s.queue = engine.ActionPlanPriotityList{}
	for key, aps := range actionPlans {
		toBeSaved := false
		isAsap := false
		newApls := make([]*engine.ActionPlan, 0) // will remove the one time runs from the database
		for _, ap := range aps {
			if ap.Timing == nil {
				utils.Logger.Warning(fmt.Sprintf("<Scheduler> Nil timing on action plan: %+v, discarding!", ap))
				continue
			}
			isAsap = ap.IsASAP()
			toBeSaved = toBeSaved || isAsap
			if isAsap {
				if len(ap.AccountIds) > 0 {
					utils.Logger.Info(fmt.Sprintf("Time for one time action on %v", key))
				}
				ap.Execute()
				ap.AccountIds = make([]string, 0)
			} else {
				now := time.Now()
				if ap.GetNextStartTime(now).Before(now) {
					// the task is obsolete, do not add it to the queue
					continue
				}
				s.queue = append(s.queue, ap)
			}
			// save even asap action plans with empty account id list
			newApls = append(newApls, ap)
		}
		if toBeSaved {
			engine.Guardian.Guard(func() (interface{}, error) {
				storage.SetActionPlans(key, newApls)
				storage.CacheRatingPrefixValues(map[string][]string{utils.ACTION_PLAN_PREFIX: []string{utils.ACTION_PLAN_PREFIX + key}})
				return 0, nil
			}, 0, utils.ACTION_PLAN_PREFIX)
		}
	}
	sort.Sort(s.queue)
	s.Unlock()
}

func (s *Scheduler) Restart() {
	s.restartLoop <- true
	if s.timer != nil {
		s.timer.Stop()
	}
}

func (s *Scheduler) GetQueue() engine.ActionPlanPriotityList {
	return s.queue
}
