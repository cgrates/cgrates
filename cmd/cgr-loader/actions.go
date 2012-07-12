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
	"github.com/cgrates/cgrates/timespans"
	"log"
	"fmt"
	"strconv"
)

var (
	actions         = make(map[string][]*timespans.Action)
	actionsTimings  = make(map[string][]*timespans.ActionTiming)
	actionsTriggers = make(map[string][]*timespans.ActionTrigger)
	accountActions  []*timespans.UserBalance
)

func (csvr *CSVReader) loadActions(fn string) {
	csvReader, fp, err := csvr.readerFunc(fn)
	if err != nil {
		return
	}
	if fp != nil {
		defer fp.Close()
	}
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		tag := record[0]
		if tag == "Tag" {
			// skip header line
			continue
		}
		units, err := strconv.ParseFloat(record[3], 64)
		if err != nil {
			log.Printf("Could not parse action units: %v", err)
			continue
		}
		var a *timespans.Action
		if record[2] != timespans.MINUTES {
			a = &timespans.Action{
				ActionType: record[1],
				BalanceId:  record[2],
				Direction:  record[3],
				Units:      units,
			}
		} else {
			price, percent := 0.0, 0.0
			value, err := strconv.ParseFloat(record[7], 64)
			if err != nil {
				log.Printf("Could not parse action price: %v", err)
				continue
			}
			if record[5] == timespans.PERCENT {
				percent = value
			}
			if record[5] == timespans.ABSOLUTE {
				price = value
			}
			minutesWeight, err := strconv.ParseFloat(record[8], 64)
			if err != nil {
				log.Printf("Could not parse action minutes weight: %v", err)
				continue
			}
			weight, err := strconv.ParseFloat(record[9], 64)
			if err != nil {
				log.Printf("Could not parse action weight: %v", err)
				continue
			}
			a = &timespans.Action{
				ActionType: record[1],
				BalanceId:  record[2],
				Direction:  record[3],
				Weight:     weight,
				MinuteBucket: &timespans.MinuteBucket{
					Seconds:       units,
					Weight:        minutesWeight,
					Price:         price,
					Percent:       percent,
					DestinationId: record[5],
				},
			}
		}
		actions[tag] = append(actions[tag], a)
	}
	log.Print("Actions:")
	log.Print(actions)
}

func (csvr *CSVReader) loadActionTimings(fn string) {
	csvReader, fp, err := csvr.readerFunc(fn)
	if err != nil {
		return
	}
	if fp != nil {
		defer fp.Close()
	}
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		tag := record[0]
		if tag == "Tag" {
			// skip header line
			continue
		}

		ts, exists := timings[record[2]]
		if !exists {
			log.Printf("Could not load the timing for tag: %v", record[2])
			continue
		}
		weight, err := strconv.ParseFloat(record[3], 64)
		if err != nil {
			log.Printf("Could not parse action timing weight: %v", err)
			continue
		}
		for _, t := range ts {
			at := &timespans.ActionTiming{
				Tag:    record[2],
				Weight: weight,
				Timing: &timespans.Interval{
					Months:    t.Months,
					MonthDays: t.MonthDays,
					WeekDays:  t.WeekDays,
					StartTime: t.StartTime,
				},
				ActionsId: record[1],
			}
			actionsTimings[tag] = append(actionsTimings[tag], at)
		}

	}
	log.Print("Actions timings:")
	log.Print(actionsTimings)
}

func (csvr *CSVReader) loadActionTriggers(fn string) {
	csvReader, fp, err := csvr.readerFunc(fn)
	if err != nil {
		return
	}
	if fp != nil {
		defer fp.Close()
	}
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		tag := record[0]
		if tag == "Tag" {
			// skip header line
			continue
		}
		value, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			log.Printf("Could not parse action trigger value: %v", err)
			continue
		}
		weight, err := strconv.ParseFloat(record[5], 64)
		if err != nil {
			log.Printf("Could not parse action trigger weight: %v", err)
			continue
		}
		at := &timespans.ActionTrigger{
			BalanceId:      record[1],
			ThresholdValue: value,
			DestinationId:  record[3],
			ActionsId:      record[4],
			Weight:         weight,
		}
		actionsTriggers[tag] = append(actionsTriggers[tag], at)
	}
	log.Print("Actions triggers:")
	log.Print(actionsTriggers)
}

func (csvr *CSVReader) loadAccountActions(fn string) {
	csvReader, fp, err := csvr.readerFunc(fn)
	if err != nil {
		return
	}
	if fp != nil {
		defer fp.Close()
	}
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		if record[0] == "Tenant" {
			continue
		}
		tag := fmt.Sprintf("%s:%s:%s", record[2], record[0], record[1])
		aTriggers, exists := actionsTriggers[record[4]]
		if !exists {
			log.Printf("Could not get action triggers for tag %v", record[4])
			continue
		}
		aTimingsTag := record[3]
		ub := &timespans.UserBalance{
			Type:           timespans.UB_TYPE_PREPAID,
			Id:             tag,
			ActionTriggers: aTriggers,
		}
		accountActions = append(accountActions, ub)

		aTimings, exists := actionsTimings[aTimingsTag]
		if !exists {
			log.Printf("Could not get action triggers for tag %v", aTimingsTag)
			// must not continue here
		}
		for _, at := range aTimings {
			at.UserBalanceIds = append(at.UserBalanceIds, tag)
		}
	}
	log.Print("Account actions:")
	log.Print(accountActions)
}
