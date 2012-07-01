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
	"flag"
	"github.com/cgrates/cgrates/timespans"
	"log"
	"os"
	"sort"
	"time"
	"os/signal"
	"syscall"
)

var (
	separator   = flag.String("separator", ",", "Default field separator")
	redisserver = flag.String("redisserver", "tcp:127.0.0.1:6379", "redis server address (tcp:127.0.0.1:6379)")
	redisdb     = flag.Int("rdb", 10, "redis database number (10)")
	redispass   = flag.String("pass", "", "redis database password")
	storage     timespans.StorageGetter
	timer       *time.Timer
	restartLoop = make(chan byte)
	s           = scheduler{}
)

/*
Structure to store action timings according to next activation time.
*/
type actiontimingqueue []*timespans.ActionTiming

func (atq actiontimingqueue) Len() int {
	return len(atq)
}

func (atq actiontimingqueue) Swap(i, j int) {
	atq[i], atq[j] = atq[j], atq[i]
}

func (atq actiontimingqueue) Less(j, i int) bool {
	return atq[j].GetNextStartTime().Before(atq[i].GetNextStartTime())
}

type scheduler struct {
	queue actiontimingqueue
}

func (s scheduler) loop() {
	for {
		a0 := s.queue[0]
		now := time.Now()
		if a0.GetNextStartTime().Equal(now) || a0.GetNextStartTime().Before(now) {
			log.Printf("%v - %v", a0.Tag, a0.Timing)
			log.Print(a0.GetNextStartTime(), now)
			go a0.Execute()
			s.queue = append(s.queue, a0)
			s.queue = s.queue[1:]
			sort.Sort(s.queue)
		} else {
			d := a0.GetNextStartTime().Sub(now)
			log.Printf("Timer set to wait for %v", d)
			timer = time.NewTimer(d)
			select {
			case <-timer.C:
				// timer has expired
				log.Printf("Time for action on %v", s.queue[0])
			case <-restartLoop:
				// nothing to do, just continue the loop
			}

		}
	}
}

// Listens for the HUP system signal and gracefuly reloads the timers from database.
func stopSingnalHandler() {
	log.Print("Handling HUP signal...")
	for {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGHUP)
		sig := <-c

		log.Printf("Caught signal %v, reloading action timings.\n", sig)
		loadActionTimings()
		// check the tip of the queue for new actions
		restartLoop <- 1
		timer.Stop()
	}
}

func loadActionTimings() {
	actionTimings, err := storage.GetAllActionTimings()
	if err != nil {
		log.Fatalf("Cannot get action timings:", err)
	}
	// recreate the queue
	s.queue = actiontimingqueue{}
	for _, at := range actionTimings {
		if at.IsOneTimeRun() {
			log.Print("Executing: ", at)
			go at.Execute()
			continue
		}
		s.queue = append(s.queue, at)
	}
	sort.Sort(s.queue)
}

func main() {
	flag.Parse()
	var err error
	storage, err = timespans.NewRedisStorage(*redisserver, *redisdb)
	if err != nil {
		log.Fatalf("Could not open database connection: %v", err)
	}
	defer storage.Close()
	timespans.SetStorageGetter(storage)
	loadActionTimings()
	go stopSingnalHandler()
	s.loop()
}
