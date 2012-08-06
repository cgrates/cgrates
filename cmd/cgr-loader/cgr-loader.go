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
	"fmt"
	"github.com/cgrates/cgrates/timespans"
	"os"
)

var (
	separator        = flag.String("separator", ",", "Default field separator")
	redissrv         = flag.String("redissrv", "127.0.0.1:6379", "redis server address (tcp:127.0.0.1:6379)")
	redisdb          = flag.Int("redisdb", 10, "redis database number (10)")
	redispass        = flag.String("pass", "", "redis database password")
	flush            = flag.Bool("flush", false, "Flush the database before importing")
	destinationsFn   = flag.String("destinations", "", "Destinations file")
	ratesFn          = flag.String("rates", "", "Rates file")
	timingsFn        = flag.String("timings", "", "Timings file")
	ratetimingsFn    = flag.String("ratetimings", "", "Rates timings file")
	ratingprofilesFn = flag.String("ratingprofiles", "", "Rating profiles file")
	actionsFn        = flag.String("actions", "", "Actions file")
	actiontimingsFn  = flag.String("actiontimings", "", "Actions timings file")
	actiontriggersFn = flag.String("actiontriggers", "", "Actions triggers file")
	accountactionsFn = flag.String("accountactions", "", "Account actions file")
	sep              rune
)

func main() {
	flag.Parse()
	sep = []rune(*separator)[0]
	csvr := timespans.NewFileCSVReader()
	csvr.LoadDestinations(*destinationsFn, sep)
	csvr.LoadRates(*ratesFn, sep)
	csvr.LoadTimings(*timingsFn, sep)
	csvr.LoadRateTimings(*ratetimingsFn, sep)
	csvr.LoadRatingProfiles(*ratingprofilesFn, sep)
	csvr.LoadActions(*actionsFn, sep)
	csvr.LoadActionTimings(*actiontimingsFn, sep)
	csvr.LoadActionTriggers(*actiontriggersFn, sep)
	csvr.LoadAccountActions(*accountactionsFn, sep)
	storage, err := timespans.NewRedisStorage(*redissrv, *redisdb, *redispass)
	if err != nil {
		timespans.Logger.Crit(fmt.Sprintf("Could not open database connection: %v", err))
		os.Exit(1)
	}
	csvr.WriteToDatabase(storage, *flush, true)
}
