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
	"strconv"
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
	ratingprofilesFn = flag.String("ratingprofiles", "RatingProfiles.csv", "Rating profiles file")
)

var (
	months         = make(map[string][]time.Month)
	monthdays      = make(map[string][]int)
	weekdays       = make(map[string][]time.Weekday)
	destinations   = make(map[string][]string)
	rates          = make(map[string][]*timespans.Interval)
	ratingProfiles = make(map[string][]*timespans.ActivationPeriod)
)

func loadDataSeries() {
	// MONTHS
	fp, err := os.Open(*monthsFn)
	if err != nil {
		log.Printf("Could not open months file: %v", err)
	}
	csvReader := csv.NewReader(fp)
	sep := []rune(*separator)[0]
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

	// MONTH DAYS
	fp, err = os.Open(*monthdaysFn)
	if err != nil {
		log.Printf("Could not open month days file: %v", err)
	}
	csvReader = csv.NewReader(fp)
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

	// WEEK DAYS
	fp, err = os.Open(*weekdaysFn)
	if err != nil {
		log.Printf("Could not open week days file: %v", err)
	}
	csvReader = csv.NewReader(fp)
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

func loadDestinations() {
	fp, err := os.Open(*destinationsFn)
	defer fp.Close()
	if err != nil {
		log.Printf("Could not open destinations file: %v", err)
		return
	}
	csvReader := csv.NewReader(fp)
	sep := []rune(*separator)[0]
	csvReader.Comma = sep
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		tag := record[0]
		if tag == "Tag" {
			// skip header line
			continue
		}
		for _, m := range record[1:] {
			destinations[tag] = append(destinations[tag], m)
		}
		log.Print(tag, destinations[tag])
	}
}

func loadRates() {
	fp, err := os.Open(*ratesFn)
	defer fp.Close()
	if err != nil {
		log.Printf("Could not open rates file: %v", err)
		return
	}
	csvReader := csv.NewReader(fp)
	sep := []rune(*separator)[0]
	csvReader.Comma = sep
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		tag := record[0]
		if tag == "Tag" {
			// skip header line
			continue
		}
		if len(record) < 9 {
			log.Printf("Malformed rates record: %v", record)
			continue
		}
		cf, err := strconv.ParseFloat(record[5], 64)
		if err != nil {
			log.Printf("Error parsing connect fee from: %v", record)
			continue
		}
		p, err := strconv.ParseFloat(record[6], 64)
		if err != nil {
			log.Printf("Error parsing price from: %v", record)
			continue
		}
		bu, err := strconv.ParseFloat(record[7], 64)
		if err != nil {
			log.Printf("Error parsing billing unit from: %v", record)
			continue
		}
		w, err := strconv.ParseFloat(record[8], 64)
		if err != nil {
			log.Printf("Error parsing weight from: %v", record)
			continue
		}
		i := &timespans.Interval{
			Months:      timespans.Months(months[record[1]]),
			MonthDays:   timespans.MonthDays(monthdays[record[2]]),
			WeekDays:    timespans.WeekDays(weekdays[record[3]]),
			StartTime:   record[4],
			ConnectFee:  cf,
			Price:       p,
			BillingUnit: bu,
			Weight:      w,
		}
		rates[tag] = append(rates[tag], i)

		log.Print(tag)
		for _, i := range rates[tag] {
			log.Print(i)
		}
	}
}

func loadRatingProfiles() {
	fp, err := os.Open(*destinationsFn)
	defer fp.Close()
	if err != nil {
		log.Printf("Could not open destinations file: %v", err)
		return
	}
	csvReader := csv.NewReader(fp)
	sep := []rune(*separator)[0]
	csvReader.Comma = sep
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		tag := record[0]
		if tag == "Tag" {
			// skip header line
			continue
		}
		for _, m := range record[1:] {
			destinations[tag] = append(destinations[tag], m)
		}
		log.Print(tag, destinations[tag])
	}
}

func main() {
	flag.Parse()
	loadDataSeries()
	loadDestinations()
	loadRates()
}
