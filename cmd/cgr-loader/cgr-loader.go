/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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
	"github.com/cgrates/cgrates/rater"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/cgrates/config"
	"log"
	"path"
	"regexp"
)

var (
	//separator = flag.String("separator", ",", "Default field separator")
	cgrConfig,_ = config.NewDefaultCGRConfig()
	data_db_type = flag.String("datadb_type", cgrConfig.DataDBType, "The type of the dataDb database (redis|mongo|postgres|mysql)")
	data_db_host = flag.String("datadb_host", cgrConfig.DataDBHost, "The dataDb host to connect to.")
	data_db_port = flag.String("datadb_port", cgrConfig.DataDBPort, "The dataDb port to bind to.")
	data_db_name = flag.String("datadb_name", cgrConfig.DataDBName, "The name/number of the dataDb to connect to.")
	data_db_user = flag.String("datadb_user", cgrConfig.DataDBUser, "The dataDb user to sign in as.")
	data_db_pass = flag.String("datadb_passwd", cgrConfig.DataDBPass,  "The dataDb user's password.")

	stor_db_type = flag.String("stordb_type", cgrConfig.StorDBType, "The type of the storDb database (redis|mongo|postgres|mysql)")
	stor_db_host = flag.String("stordb_host", cgrConfig.StorDBHost, "The storDb host to connect to.")
	stor_db_port = flag.String("stordb_port", cgrConfig.StorDBPort, "The storDb port to bind to.")
	stor_db_name = flag.String("stordb_name", cgrConfig.StorDBName, "The name/number of the storDb to connect to.")
	stor_db_user = flag.String("stordb_user", cgrConfig.StorDBUser, "The storDb user to sign in as.")
	stor_db_pass = flag.String("stordb_passwd", cgrConfig.StorDBPass, "The storDb user's password.")

	flush    = flag.Bool("flush", false, "Flush the database before importing")
	tpid = flag.String("tpid", "", "The tariff plan id from the database")
	dataPath = flag.String("path", ".", "The path containing the data files")
	version  = flag.Bool("version", false, "Prints the application version.")
	importer = flag.Bool("import", false, "Import to storDb instead of directly loading to dataDb")

	sep                      rune
)

type validator struct {
	fn      string
	re      *regexp.Regexp
	message string
}

func main() {
	flag.Parse()
	if *version {
		fmt.Println("CGRateS " + utils.VERSION)
		return
	}
	var err error
	var db rater.DataStorage
	if *importer { // Loader has importer function, we need connection to storDb
		db, err = rater.ConfigureDatabase(*stor_db_type, *stor_db_host, *stor_db_port, *stor_db_name, *stor_db_user, *stor_db_pass)
	} else { // Loader function, need connection directly to dataDb
		db, err = rater.ConfigureDatabase(*data_db_type, *data_db_host, *data_db_port, *data_db_name, *data_db_user, *data_db_pass)
	}
	defer db.Close()
	if err != nil {
		log.Fatalf("Could not open database connection: %v", err)
	}

	if *tpid != "" && *dataPath != "" {
		log.Fatal("You can read either from db or from files, not both.")
	}
	var loader rater.TPLoader
	if *dataPath != "" {
		dataFilesValidators := []*validator{
			&validator{utils.DESTINATIONS_CSV,
				regexp.MustCompile(`(?:\w+\s*,\s*){1}(?:\d+.?\d*){1}$`),
				"Tag[0-9A-Za-z_],Prefix[0-9]"},
			&validator{utils.TIMINGS_CSV,
				regexp.MustCompile(`(?:\w+\s*,\s*){1}(?:\*all\s*,\s*|(?:\d{1,4};?)+\s*,\s*|\s*,\s*){4}(?:\d{2}:\d{2}:\d{2}|\*asap){1}$`),
				"Tag[0-9A-Za-z_],Years[0-9;]|*all|<empty>,Months[0-9;]|*all|<empty>,MonthDays[0-9;]|*all|<empty>,WeekDays[0-9;]|*all|<empty>,Time[0-9:]|*asap(00:00:00)"},
			&validator{utils.RATES_CSV,
				regexp.MustCompile(`(?:\w+\s*,\s*){2}(?:\d+.?\d*,?){4}$`),
				"Tag[0-9A-Za-z_],ConnectFee[0-9.],Price[0-9.],PricedUnits[0-9.],RateIncrement[0-9.]"},
			&validator{utils.DESTINATION_RATES_CSV,
				regexp.MustCompile(`(?:\w+\s*,\s*){2}(?:\d+.?\d*,?){4}$`),
				"Tag[0-9A-Za-z_],DestinationsTag[0-9A-Za-z_],RateTag[0-9A-Za-z_]"},
			&validator{utils.DESTRATE_TIMINGS_CSV,
				regexp.MustCompile(`(?:\w+\s*,\s*){3}(?:\d+.?\d*){1}$`),
				"Tag[0-9A-Za-z_],DestinationRatesTag[0-9A-Za-z_],TimingProfile[0-9A-Za-z_],Weight[0-9.]"},
			&validator{utils.RATE_PROFILES_CSV,
				regexp.MustCompile(`(?:\w+\s*,\s*){1}(?:\d+\s*,\s*){1}(?:OUT\s*,\s*|IN\s*,\s*){1}(?:\*all\s*,\s*|[\w:\.]+\s*,\s*){1}(?:\w*\s*,\s*){1}(?:\w+\s*,\s*){1}(?:\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z){1}$`),
				"Tenant[0-9A-Za-z_],TOR[0-9],Direction OUT|IN,Subject[0-9A-Za-z_:.]|*all,RatesFallbackSubject[0-9A-Za-z_]|<empty>,RatesTimingTag[0-9A-Za-z_],ActivationTime[[0-9T:X]] (2012-01-01T00:00:00Z)"},
			&validator{utils.ACTIONS_CSV,
				regexp.MustCompile(`(?:\w+\s*,\s*){3}(?:OUT\s*,\s*|IN\s*,\s*){1}(?:\d+\s*,\s*){1}(?:\w+\s*,\s*|\*all\s*,\s*){1}(?:ABSOLUTE\s*,\s*|PERCENT\s*,\s*|\s*,\s*){1}(?:\d*\.?\d*\s*,?\s*){3}$`),
				"Tag[0-9A-Za-z_],Action[0-9A-Za-z_],BalanceTag[0-9A-Za-z_],Direction OUT|IN,Units[0-9],DestinationTag[0-9A-Za-z_]|*all,PriceType ABSOLUT|PERCENT,PriceValue[0-9.],MinutesWeight[0-9.],Weight[0-9.]"},
			&validator{utils.ACTION_TIMINGS_CSV,
				regexp.MustCompile(`(?:\w+\s*,\s*){3}(?:\d+\.?\d*){1}`),
				"Tag[0-9A-Za-z_],ActionsTag[0-9A-Za-z_],TimingTag[0-9A-Za-z_],Weight[0-9.]"},
			&validator{utils.ACTION_TRIGGERS_CSV,
				regexp.MustCompile(`(?:\w+\s*,\s*){1}(?:MONETARY\s*,\s*|SMS\s*,\s*|MINUTES\s*,\s*|INTERNET\s*,\s*|INTERNET_TIME\s*,\s*){1}(?:OUT\s*,\s*|IN\s*,\s*){1}(?:\d+\.?\d*\s*,\s*){1}(?:\w+\s*,\s*|\*all\s*,\s*){1}(?:\w+\s*,\s*){1}(?:\d+\.?\d*){1}$`),
				"Tag[0-9A-Za-z_],BalanceTag MONETARY|SMS|MINUTES|INTERNET|INTERNET_TIME,Direction OUT|IN,ThresholdValue[0-9.],DestinationTag[0-9A-Za-z_]|*all,ActionsTag[0-9A-Za-z_],Weight[0-9.]"},
			&validator{utils.ACCOUNT_ACTIONS_CSV,
				regexp.MustCompile(`(?:\w+\s*,\s*){1}(?:[\w:.]+\s*,\s*){1}(?:OUT\s*,\s*|IN\s*,\s*){1}(?:\w+\s*,?\s*){2}$`),
				"Tenant[0-9A-Za-z_],Account[0-9A-Za-z_:.],Direction OUT|IN,ActionTimingsTag[0-9A-Za-z_],ActionTriggersTag[0-9A-Za-z_]"},
		}
		for _, v := range dataFilesValidators {
			err := rater.ValidateCSVData(path.Join(*dataPath, v.fn), v.re)
			if err != nil {
				log.Fatal(err, "\n\t", v.message)
			}
		}
		//sep = []rune(*separator)[0]
		loader = rater.NewFileCSVReader(db, ',', utils.DESTINATIONS_CSV, utils.TIMINGS_CSV, utils.RATES_CSV, utils.DESTINATION_RATES_CSV, utils.DESTRATE_TIMINGS_CSV, utils.RATE_PROFILES_CSV, utils.ACTIONS_CSV, utils.ACTION_TIMINGS_CSV, utils.ACTION_TRIGGERS_CSV, utils.ACCOUNT_ACTIONS_CSV)
	}

	if *tpid != "" {
		loader = rater.NewDbReader(db, db, *tpid)
	}

	err = loader.LoadDestinations()
	if err != nil {
		log.Fatal(err)
	}
	err = loader.LoadTimings()
	if err != nil {
		log.Fatal(err)
	}
	err = loader.LoadRates()
	if err != nil {
		log.Fatal(err)
	}
	err = loader.LoadDestinationRates()
	if err != nil {
		log.Fatal(err)
	}
	err = loader.LoadDestinationRateTimings()
	if err != nil {
		log.Fatal(err)
	}
	err = loader.LoadRatingProfiles()
	if err != nil {
		log.Fatal(err)
	}
	err = loader.LoadActions()
	if err != nil {
		log.Fatal(err)
	}
	err = loader.LoadActionTimings()
	if err != nil {
		log.Fatal(err)
	}
	err = loader.LoadActionTriggers()
	if err != nil {
		log.Fatal(err)
	}
	err = loader.LoadAccountActions()
	if err != nil {
		log.Fatal(err)
	}

	// write maps to database
	if err := loader.WriteToDatabase(*flush, true); err != nil {
		log.Fatal("Could not write to database: ", err)
	}
}
