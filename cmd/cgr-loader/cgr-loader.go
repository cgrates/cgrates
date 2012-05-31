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
	"github.com/rif/cgrates/timespans"
	"log"
	"os"
	"encoding/csv"
	"time"
)

var (
	separator        = flag.String("separator", ",", "Default field separator")
	redisserver      = flag.String("redisserver", "tcp:127.0.0.1:6379", "redis server address (tcp:127.0.0.1:6379)")
	redisdb          = flag.Int("rdb", 10, "redis database number (10)")
	redispass        = flag.String("pass", "", "redis database password")
	monthsFn         = flag.String("month", "Months.csv", "Months file")
	monthdaysFn      = flag.String("monthdays", "MonthDays.csv", "Month days file")
	weekdaysFn       = flag.String("weekdays", "WeekDays.csv", "Week days file")
	destinationsFn   = flag.String("destinations", "Destinations.csv", "Destinations file")
	ratesFn          = flag.String("rates", "Rates.csv", "Rates file")
	ratestimingFn    = flag.String("ratestiming", "RatesTiming.csv", "Rates timing file")
	ratingprofilesFn = flag.String("ratingprofiles", "RatingProfiles.csv", "Rating profiles file")
)

var (
	months         = make(map[string][]time.Month)
	monthdays      = make(map[string][]int)
	weekdays       = make(map[string][]time.Weekday)
	destinations   = make(map[string][]string)
	rates          = make(map[string][]*Rate)
	ratesTiming    = make(map[string][]*RateTiming)
	ratingProfiles = make(map[string][]*timespans.CallDescriptor)
)

func loadDataSeries() {
	sep := []rune(*separator)[0]
	// MONTHS
	fp, err := os.Open(*monthsFn)
	if err != nil {
		log.Printf("Could not open months file: %v", err)
	} else {
		csvReader := csv.NewReader(fp)
		csvReader.Comma = sep
		for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
			tag := record[0]
			if tag == "Tag" {
				// skip header line
				continue
			}
			for i, m := range record[1:] {
				if m == "1" {
					months[tag] = append(months[tag], time.Month(i+1))
				}
			}
			log.Print(tag, months[tag])
		}
		fp.Close()
	}
	// MONTH DAYS
	fp, err = os.Open(*monthdaysFn)
	if err != nil {
		log.Printf("Could not open month days file: %v", err)
	} else {
		csvReader := csv.NewReader(fp)
		csvReader.Comma = sep
		for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
			tag := record[0]
			if tag == "Tag" {
				// skip header line
				continue
			}
			for i, m := range record[1:] {
				if m == "1" {
					monthdays[tag] = append(monthdays[tag], i+1)
				}
			}
			log.Print(tag, monthdays[tag])
		}
		fp.Close()
	}
	// WEEK DAYS
	fp, err = os.Open(*weekdaysFn)
	if err != nil {
		log.Printf("Could not open week days file: %v", err)
	} else {
		csvReader := csv.NewReader(fp)
		csvReader.Comma = sep
		for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
			tag := record[0]
			if tag == "Tag" {
				// skip header line
				continue
			}
			for i, m := range record[1:] {
				if m == "1" {
					weekdays[tag] = append(weekdays[tag], time.Weekday(((i + 1) % 7)))
				}
			}
			log.Print(tag, weekdays[tag])
		}
		fp.Close()
	}
}

func loadDestinations() {
	fp, err := os.Open(*destinationsFn)
	if err != nil {
		log.Printf("Could not open destinations file: %v", err)
		return
	}
	defer fp.Close()
	csvReader := csv.NewReader(fp)
	sep := []rune(*separator)[0]
	csvReader.Comma = sep
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		tag := record[0]
		if tag == "Tag" {
			// skip header line
			continue
		}
		destinations[tag] = record[1:]
		log.Print(tag, destinations[tag])
	}
}

func loadRates() {
	fp, err := os.Open(*ratesFn)
	if err != nil {
		log.Printf("Could not open rates timing file: %v", err)
		return
	}
	defer fp.Close()
	csvReader := csv.NewReader(fp)
	sep := []rune(*separator)[0]
	csvReader.Comma = sep
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		tag := record[0]
		if tag == "Tag" {
			// skip header line
			continue
		}
		r, err := NewRate(record[1], record[2], record[3], record[4], record[5])
		if err != nil {
			continue
		}
		rates[tag] = append(rates[tag], r)
		log.Print(tag, rates[tag])
	}
}

func loadRatesTiming() {
	fp, err := os.Open(*ratestimingFn)
	if err != nil {
		log.Printf("Could not open rates file: %v", err)
		return
	}
	defer fp.Close()
	csvReader := csv.NewReader(fp)
	sep := []rune(*separator)[0]
	csvReader.Comma = sep
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		tag := record[0]
		if tag == "Tag" {
			// skip header line
			continue
		}

		rt := NewRateTiming(record[1], record[2], record[3], record[4], record[5])
		ratesTiming[tag] = append(ratesTiming[tag], rt)

		log.Print(tag)
		for _, i := range ratesTiming[tag] {
			log.Print(i)
		}
	}
}

func loadRatingProfiles() {
	fp, err := os.Open(*ratingprofilesFn)
	if err != nil {
		log.Printf("Could not open destinations rates file: %v", err)
		return
	}
	defer fp.Close()
	csvReader := csv.NewReader(fp)
	sep := []rune(*separator)[0]
	csvReader.Comma = sep
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		tag := record[0]
		if tag == "Tenant" {
			// skip header line
			continue
		}
		tenant, tor, subject := record[0], record[1], record[2]
		at, err := time.Parse(time.RFC3339, record[5])
		if err != nil {
			log.Printf("Cannot parse activation time from %v", record[5])
			continue
		}
		rts, exists := ratesTiming[record[4]]
		if !exists {
			log.Printf("Could not get rate timing for tag %v", record[4])
			continue
		}
		for _, rt := range rts { // rates timing
			rs, exists := rates[rt.RatesTag]
			if !exists {
				log.Printf("Could not get rates for tag %v", rt.RatesTag)
				continue
			}
			ap := &timespans.ActivationPeriod{
				ActivationTime: at,
			}
			for _, r := range rs { //rates
				ds, exists := destinations[r.DestinationsTag]
				if !exists {
					log.Printf("Could not get destinations for tag %v", r.DestinationsTag)
					continue
				}
				ap.AddInterval(rt.GetInterval(r))
				for _, d := range ds { //destinations
					cd := &timespans.CallDescriptor{
						Tenant:            tenant,
						TOR:               tor,
						Subject:           subject,
						DestinationPrefix: d,
					}
					cd.ActivationPeriods = append(cd.ActivationPeriods, ap)
					ratingProfiles[d] = append(ratingProfiles[d], cd)
				}
			}
		}

		for key, cds := range ratingProfiles {
			log.Print(key)
			for _, cd := range cds {
				log.Print(cd, cd.ActivationPeriods[0])
			}
		}
	}
}

func main() {
	flag.Parse()
	loadDataSeries()
	loadDestinations()
	loadRates()
	loadRatesTiming()
	loadRatingProfiles()
}
