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
	"log"
	"path"
	"regexp"
	"strconv"
)

const (
	POSTGRES = "postgres"
	MONGO    = "mongo"
	REDIS    = "redis"
)

var (
	//separator = flag.String("separator", ",", "Default field separator")
	db_type = flag.String("dbtype", REDIS, "The type of the database (redis|mongo|postgres)")
	db_host = flag.String("dbhost", "localhost", "The database host to connect to.")
	db_port = flag.String("dbport", "6379", "The database port to bind to.")
	db_name = flag.String("dbname", "10", "he name/number of the database to connect to.")
	db_user = flag.String("dbuser", "", "The database user to sign in as.")
	db_pass = flag.String("dbpass", "", "The database user's password.")

	flush    = flag.Bool("flush", false, "Flush the database before importing")
	dataPath = flag.String("path", ".", "The path containing the data files")
	version  = flag.Bool("version", false, "Prints the application version.")

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

type validator struct {
	fn      string
	re      *regexp.Regexp
	message string
}

func main() {
	flag.Parse()
	if *version {
		fmt.Println("CGRateS " + timespans.VERSION)
		return
	}
	dataFilesValidators := []*validator{
		&validator{destinationsFn,
			regexp.MustCompile(`(?:\w+\s*,\s*){1}(?:\d+.?\d*){1}$`),
			"Tag[0-9A-Za-z_],Prefix[0-9]"},
		&validator{ratesFn,
			regexp.MustCompile(`(?:\w+\s*,\s*){2}(?:\d+.?\d*){4}$`),
			"Tag[0-9A-Za-z_],DestinationsTag[0-9A-Za-z_],ConnectFee[0-9.],Price[0-9.],PricedUnits[0-9.],RateIncrement[0-9.]"},
		&validator{timingsFn,
			regexp.MustCompile(`(?:\w+\s*,\s*){1}(?:\*all\s*,\s*|(?:\d{1,4};?)+\s*,\s*|\s*,\s*){4}(?:\d{2}:\d{2}:\d{2}|\*asap){1}$`),
			"Tag[0-9A-Za-z_],Years[0-9;]|*all|<empty>,Months[0-9;]|*all|<empty>,MonthDays[0-9;]|*all|<empty>,WeekDays[0-9;]|*all|<empty>,Time[0-9:]|*asap(00:00:00)"},
		&validator{ratetimingsFn,
			regexp.MustCompile(`(?:\w+\s*,\s*){3}(?:\d+.?\d*){1}$`),
			"Tag[0-9A-Za-z_],RatesTag[0-9A-Za-z_],TimingProfile[0-9A-Za-z_],Weight[0-9.]"},
		&validator{ratingprofilesFn,
			regexp.MustCompile(`(?:\w+\s*,\s*){1}(?:\d+\s*,\s*){1}(?:OUT\s*,\s*|IN\s*,\s*){1}(?:\*all\s*,\s*|[\w:\.]+\s*,\s*){1}(?:\w*\s*,\s*){1}(?:\w+\s*,\s*){1}(?:\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z){1}$`),
			"Tenant[0-9A-Za-z_],TOR[0-9],Direction OUT|IN,Subject[0-9A-Za-z_:.]|*all,RatesFallbackSubject[0-9A-Za-z_]|<empty>,RatesTimingTag[0-9A-Za-z_],ActivationTime[[0-9T:X]] (2012-01-01T00:00:00Z)"},
		&validator{actionsFn,
			regexp.MustCompile(`(?:\w+\s*,\s*){3}(?:OUT\s*,\s*|IN\s*,\s*){1}(?:\d+\s*,\s*){1}(?:\w+\s*,\s*|\*all\s*,\s*){1}(?:ABSOLUTE\s*,\s*|PERCENT\s*,\s*|\s*,\s*){1}(?:\d*\.?\d*\s*,?\s*){3}$`),
			"Tag[0-9A-Za-z_],Action[0-9A-Za-z_],BalanceTag[0-9A-Za-z_],Direction OUT|IN,Units[0-9],DestinationTag[0-9A-Za-z_]|*all,PriceType ABSOLUT|PERCENT,PriceValue[0-9.],MinutesWeight[0-9.],Weight[0-9.]"},
		&validator{actiontimingsFn,
			regexp.MustCompile(`(?:\w+\s*,\s*){3}(?:\d+\.?\d*){1}`),
			"Tag[0-9A-Za-z_],ActionsTag[0-9A-Za-z_],TimingTag[0-9A-Za-z_],Weight[0-9.]"},
		&validator{actiontriggersFn,
			regexp.MustCompile(`(?:\w+\s*,\s*){1}(?:MONETARY\s*,\s*|SMS\s*,\s*|MINUTES\s*,\s*|INTERNET\s*,\s*|INTERNET_TIME\s*,\s*){1}(?:OUT\s*,\s*|IN\s*,\s*){1}(?:\d+\.?\d*\s*,\s*){1}(?:\w+\s*,\s*|\*all\s*,\s*){1}(?:\w+\s*,\s*){1}(?:\d+\.?\d*){1}$`),
			"Tag[0-9A-Za-z_],BalanceTag MONETARY|SMS|MINUTES|INTERNET|INTERNET_TIME,Direction OUT|IN,ThresholdValue[0-9.],DestinationTag[0-9A-Za-z_]|*all,ActionsTag[0-9A-Za-z_],Weight[0-9.]"},
		&validator{accountactionsFn,
			regexp.MustCompile(`(?:\w+\s*,\s*){1}(?:[\w:.]+\s*,\s*){1}(?:OUT\s*,\s*|IN\s*,\s*){1}(?:\w+\s*,?\s*){2}$`),
			"Tenant[0-9A-Za-z_],Account[0-9A-Za-z_:.],Direction OUT|IN,ActionTimingsTag[0-9A-Za-z_],ActionTriggersTag[0-9A-Za-z_]"},
	}
	for _, v := range dataFilesValidators {
		err := timespans.ValidateCSVData(path.Join(*dataPath, v.fn), v.re)
		if err != nil {
			log.Fatal(err, "\n\t", v.message)
		}
	}

	//sep = []rune(*separator)[0]
	sep = ','
	csvr := timespans.NewFileCSVReader()
	err := csvr.LoadDestinations(path.Join(*dataPath, destinationsFn), sep)
	if err != nil {
		log.Fatal(err)
	}
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
	var getter timespans.DataStorage
	switch *db_type {
	case REDIS:
		db_nb, err := strconv.Atoi(*db_name)
		if err != nil {
			log.Fatal("Redis db name must be an integer!")
		}
		if *db_port != "" {
			*db_host += ":" + *db_port
		}
		getter, err = timespans.NewRedisStorage(*db_host, db_nb, *db_pass)
	case MONGO:
		getter, err = timespans.NewMongoStorage(*db_host, *db_port, *db_name, *db_user, *db_pass)
	case POSTGRES:
		getter, err = timespans.NewPostgresStorage(*db_host, *db_port, *db_name, *db_user, *db_pass)
	default:
		log.Fatal("Unknown data db type, exiting!")
	}

	if err != nil {
		log.Fatalf("Could not open database connection: %v", err)
	}

	// writing to database
	csvr.WriteToDatabase(getter, *flush, true)
}
