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
	"github.com/cgrates/cgrates/timespans"
	"log"
	"fmt"
	"os"
	"time"
)

var (
	destinations   []*timespans.Destination
	rates          = make(map[string][]*Rate)
	timings        = make(map[string][]*Timing)
	ratesTimings   = make(map[string][]*RateTiming)
	ratingProfiles = make(map[string]CallDescriptors)
)

func loadDestinations() {
	fp, err := os.Open(*destinationsFn)
	if err != nil {
		log.Printf("Could not open destinations file: %v", err)
		return
	}
	defer fp.Close()
	csvReader := csv.NewReader(fp)
	csvReader.Comma = sep
	csvReader.TrailingComma = true
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		tag := record[0]
		if tag == "Tag" {
			// skip header line
			continue
		}
		var dest *timespans.Destination
		for _, d := range destinations {
			if d.Id == tag {
				dest = d
				break
			}
		}
		if dest == nil {
			dest = &timespans.Destination{Id: tag}
			destinations = append(destinations, dest)
		}
		dest.Prefixes = append(dest.Prefixes, record[1:]...)
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
	csvReader.Comma = sep
	csvReader.TrailingComma = true
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
	}
}

func loadTimings() {
	fp, err := os.Open(*timingsFn)
	if err != nil {
		log.Printf("Could not open timings file: %v", err)
		return
	}
	defer fp.Close()
	csvReader := csv.NewReader(fp)
	csvReader.Comma = sep
	csvReader.TrailingComma = true
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		tag := record[0]
		if tag == "Tag" {
			// skip header line
			continue
		}

		t := NewTiming(record[1:]...)
		timings[tag] = append(timings[tag], t)
	}
}

func loadRatesTimings() {
	fp, err := os.Open(*ratestimingsFn)
	if err != nil {
		log.Printf("Could not open rates timings file: %v", err)
		return
	}
	defer fp.Close()
	csvReader := csv.NewReader(fp)
	csvReader.Comma = sep
	csvReader.TrailingComma = true
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		tag := record[0]
		if tag == "Tag" {
			// skip header line
			continue
		}

		ts, exists := timings[record[2]]
		if !exists {
			log.Printf("Could not get timing for tag %v", record[2])
			continue
		}

		for _, t := range ts {
			rt := NewRateTiming(record[1], t)
			ratesTimings[tag] = append(ratesTimings[tag], rt)
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
	csvReader.Comma = sep
	csvReader.TrailingComma = true
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		tag := record[0]
		if tag == "Tenant" {
			// skip header line
			continue
		}
		if len(record) != 7 {
			log.Printf("Malformed rating profile: %v", record)
			continue
		}
		tenant, tor, direction, subject, fallbacksubject := record[0], record[1], record[2], record[3], record[4]
		at, err := time.Parse(time.RFC3339, record[6])
		if err != nil {
			log.Printf("Cannot parse activation time from %v", record[5])
			continue
		}
		rts, exists := ratesTimings[record[5]]
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
				for _, d := range destinations {
					if d.Id == r.DestinationsTag {
						ap.AddInterval(rt.GetInterval(r))
						for _, p := range d.Prefixes { //destinations
							// Search for a CallDescriptor with the same key
							var cd *timespans.CallDescriptor
							for _, c := range ratingProfiles[p] {
								if c.GetKey() == r.DestinationsTag {
									cd = c
								}
							}
							if cd == nil {
								cd = &timespans.CallDescriptor{
									Direction:   direction,
									Tenant:      tenant,
									TOR:         tor,
									Subject:     subject,
									Destination: p,
								}
								ratingProfiles[p] = append(ratingProfiles[p], cd)
							}
							cd.ActivationPeriods = append(cd.ActivationPeriods, ap)
							if fallbacksubject != "" &&
								ratingProfiles[p].getKey(fmt.Sprintf("%s:%s:%s:%s:%s", direction, tenant, tor, subject, timespans.FallbackDestination)) == nil {
								cd = &timespans.CallDescriptor{
									Direction:   direction,
									Tenant:      tenant,
									TOR:         tor,
									Subject:     subject,
									Destination: timespans.FallbackDestination,
									FallbackKey: fmt.Sprintf("%s:%s:%s:%s", direction, tenant, tor, fallbacksubject),
								}
								ratingProfiles[p] = append(ratingProfiles[p], cd)
							}
						}
					}
				}
			}
		}
	}
	log.Print("Call descriptors:")
	for dest, cds := range ratingProfiles {
		log.Print(dest)
		for _, cd := range cds {
			log.Print(cd)
		}
	}
}
