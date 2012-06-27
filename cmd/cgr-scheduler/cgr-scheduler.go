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
	timer       *time.Timer
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
			a0.Execute()
			s.queue = append(s.queue, a0)
			s.queue = s.queue[1:]
			sort.Sort(s.queue)
		} else {
			d := a0.GetNextStartTime().Sub(now)
			timer = time.NewTimer(d)
			<-timer.C
		}
	}
}

// Listens for the HUP system signal and gracefuly reloads the timers from database.
func stopSingnalHandler() {
	log.Print("Handling HUP signal...")
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP)
	sig := <-c

	log.Printf("Caught signal %v, reloading action timings.\n", sig)
	loadActionTimings()
}

func loadActionTimings() (actionTimings []*timespans.ActionTiming, err error) {
	storage, err := timespans.NewRedisStorage(*redisserver, *redisdb)
	defer storage.Close()
	if err != nil {
		log.Fatalf("Could not open database connection: %v", err)
	}
	actionTimings, err = storage.GetAllActionTimings()
	return
}

func main() {
	flag.Parse()
	actionTimings, err := loadActionTimings()
	if err != nil {
		log.Fatalf("Cannot get action timings:", err)
	}
	s := scheduler{}
	s.queue = append(s.queue, actionTimings...)
	go stopSingnalHandler()
	s.loop()
}
