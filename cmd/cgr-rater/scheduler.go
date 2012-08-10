/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012  Radu Ioan Fericean

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

package main

import (
	"fmt"
	"github.com/cgrates/cgrates/timespans"
	"sort"
	"time"
)

var (
	sched       = new(scheduler)
	timer       *time.Timer
	restartLoop = make(chan byte)
)

type scheduler struct {
	queue timespans.ActionTimingPriotityList
}

func (s scheduler) loop() {
	for {
		if len(s.queue) == 0 {
			<-restartLoop
		}
		a0 := s.queue[0]
		now := time.Now()
		if a0.GetNextStartTime().Equal(now) || a0.GetNextStartTime().Before(now) {
			timespans.Logger.Debug(fmt.Sprintf("%v - %v", a0.Tag, a0.Timing))
			go a0.Execute()
			sched.queue = append(s.queue, a0)
			sched.queue = s.queue[1:]
			sort.Sort(sched.queue)
		} else {
			d := a0.GetNextStartTime().Sub(now)
			timespans.Logger.Info(fmt.Sprintf("Timer set to wait for %v", d))
			timer = time.NewTimer(d)
			select {
			case <-timer.C:
				// timer has expired
				timespans.Logger.Info(fmt.Sprintf("Time for action on %v", s.queue[0]))
			case <-restartLoop:
				// nothing to do, just continue the loop
			}

		}
	}
}

func loadActionTimings(storage timespans.StorageGetter) {
	actionTimings, err := storage.GetAllActionTimings()
	if err != nil {
		timespans.Logger.Warning(fmt.Sprintf("Cannot get action timings: %v", err))
	}
	// recreate the queue
	sched.queue = timespans.ActionTimingPriotityList{}
	for key, ats := range actionTimings {
		toBeSaved := false
		for i, at := range ats {
			toBeSaved = toBeSaved || at.CheckForASAP()
			if at.IsOneTimeRun() {
				go at.Execute()
				// remove it from list
				ats = append(ats[:i], ats[i+1:]...)
			} else {
				sched.queue = append(sched.queue, at)
			}
		}
		if toBeSaved {
			storage.SetActionTimings(key, ats)
		}
	}
	sort.Sort(sched.queue)
}
