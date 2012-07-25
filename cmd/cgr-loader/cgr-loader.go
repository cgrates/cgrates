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
	"encoding/csv"
	"flag"
	"github.com/cgrates/cgrates/timespans"
	"log"
	"os"
	"strings"
)

var (
	separator        = flag.String("separator", ",", "Default field separator")
	redissrv         = flag.String("redissrv", "127.0.0.1:6379", "redis server address (tcp:127.0.0.1:6379)")
	redisdb          = flag.Int("redisdb", 10, "redis database number (10)")
	redispass        = flag.String("pass", "", "redis database password")
	flush            = flag.Bool("flush", false, "Flush the database before importing")
	monthsFn         = flag.String("month", "", "Months file")
	monthdaysFn      = flag.String("monthdays", "", "Month days file")
	weekdaysFn       = flag.String("weekdays", "", "Week days file")
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

func writeToDatabase() {
	storage, err := timespans.NewRedisStorage(*redissrv, *redisdb)
	if err != nil {
		log.Fatalf("Could not open database connection: %v", err)
	}
	if *flush {
		storage.Flush()
	}
	log.Print("Destinations")
	for _, d := range destinations {
		storage.SetDestination(d)
		log.Print(d.Id, " : ", d.Prefixes)
	}
	log.Print("Rating profiles")
	for _, cds := range ratingProfiles {
		for _, cd := range cds {
			err = storage.SetActivationPeriodsOrFallback(cd.GetKey(), cd.ActivationPeriods, cd.FallbackKey)
			log.Print(cd.GetKey())
		}
	}
	log.Print("Action timings")
	for k, ats := range actionsTimings {
		storage.SetActionTimings(timespans.ACTION_TIMING_PREFIX+":"+k, ats)
		log.Println(k)
	}
	log.Print("Actions")
	for k, as := range actions {
		storage.SetActions(k, as)
		log.Println(k)
	}
	log.Print("Account actions")
	for _, ub := range accountActions {
		storage.SetUserBalance(ub)
		log.Println(ub.Id)
	}
}

func openFileCSVReader(fn string) (csvReader *csv.Reader, fp *os.File, err error) {
	fp, err = os.Open(fn)
	if err != nil {
		return
	}
	csvReader = csv.NewReader(fp)
	csvReader.Comma = sep
	csvReader.TrailingComma = true
	return
}

func openStringCSVReader(data string) (csvReader *csv.Reader, fp *os.File, err error) {
	csvReader = csv.NewReader(strings.NewReader(data))
	csvReader.Comma = ','
	csvReader.TrailingComma = true
	return
}

type CSVReader struct {
	readerFunc func(string) (*csv.Reader, *os.File, error)
}

func main() {
	flag.Parse()
	sep = []rune(*separator)[0]
	csvr := &CSVReader{openFileCSVReader}
	csvr.loadDestinations(*destinationsFn)
	csvr.loadRates(*ratesFn)
	csvr.loadTimings(*timingsFn)
	csvr.loadRateTimings(*ratetimingsFn)
	csvr.loadRatingProfiles(*ratingprofilesFn)
	csvr.loadActions(*actionsFn)
	csvr.loadActionTimings(*actiontimingsFn)
	csvr.loadActionTriggers(*actiontriggersFn)
	csvr.loadAccountActions(*accountactionsFn)
	writeToDatabase()
}
