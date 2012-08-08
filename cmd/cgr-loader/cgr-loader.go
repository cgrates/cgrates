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
	"path"
)

var (
	separator        = flag.String("separator", ",", "Default field separator")
	redissrv         = flag.String("redissrv", "127.0.0.1:6379", "redis server address (tcp:127.0.0.1:6379)")
	redisdb          = flag.Int("redisdb", 10, "redis database number (10)")
	redispass        = flag.String("pass", "", "redis database password")
	flush            = flag.Bool("flush", false, "Flush the database before importing")
	dataPath         = flag.String("path", ".", "The path containing the data files")
	destinationsFn   = "Destinations.csv"
	ratesFn          = "Rates.csv"
	timingsFn        = "Timings.csv"
	ratetimingsFn    = "RateTimings.csv"
	ratingprofilesFn = "RatingProfiles.csv"
	actionsFn        = "Actions.csv"
	actiontimingsFn  = "ActionTimings.csv"
	actiontriggersFn = "ActionTriggers.csv"
	accountactionsFn = "AccountActions.csv"
	sep              rune
)

func main() {
	flag.Parse()
	sep = []rune(*separator)[0]
	csvr := timespans.NewFileCSVReader()
	err := csvr.LoadDestinations(path.Join(*dataPath, destinationsFn), sep)
	err = csvr.LoadRates(path.Join(*dataPath, ratesFn), sep)
	if err != nil {
		log.Fatal(err)
	}
	err = csvr.LoadTimings(path.Join(*dataPath, timingsFn), sep)
	if err != nil {
		log.Fatal(err)
	}
	err = csvr.LoadRateTimings(path.Join(*dataPath, ratetimingsFn), sep)
	if err != nil {
		log.Fatal(err)
	}
	err = csvr.LoadRatingProfiles(path.Join(*dataPath, ratingprofilesFn), sep)
	if err != nil {
		log.Fatal(err)
	}
	err = csvr.LoadActions(path.Join(*dataPath, actionsFn), sep)
	if err != nil {
		log.Fatal(err)
	}
	err = csvr.LoadActionTimings(path.Join(*dataPath, actiontimingsFn), sep)
	if err != nil {
		log.Fatal(err)
	}
	err = csvr.LoadActionTriggers(path.Join(*dataPath, actiontriggersFn), sep)
	if err != nil {
		log.Fatal(err)
	}
	err = csvr.LoadAccountActions(path.Join(*dataPath, accountactionsFn), sep)
	if err != nil {
		log.Fatal(err)
	}
	storage, err := timespans.NewRedisStorage(*redissrv, *redisdb, *redispass)
	if err != nil {
		log.Fatal("Could not open database connection: %v", err)
	}
	csvr.WriteToDatabase(storage, *flush, true)
}
