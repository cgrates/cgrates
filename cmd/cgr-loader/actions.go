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
	"os"
	"fmt"
	"strconv"
)

var (
	actions         = make(map[string][]*timespans.Action)
	actionsTimings  = make(map[string][]*timespans.ActionTiming)
	actionsTriggers = make(map[string][]*timespans.ActionTrigger)
	accountActions  = make(map[string][]*timespans.UserBalance)
)

func loadActions() {
	fp, err := os.Open(*actionsFn)
	if err != nil {
		log.Printf("Could not open actions file: %v", err)
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
				Units:      units,
			}
		} else {
			price, percent := 0.0, 0.0
			value, err := strconv.ParseFloat(record[6], 64)
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
			weight, err := strconv.ParseFloat(record[7], 64)
			if err != nil {
				log.Printf("Could not parse action units: %v", err)
				continue
			}
			a = &timespans.Action{
				ActionType: record[1],
				BalanceId:  record[2],
				MinuteBucket: &timespans.MinuteBucket{
					Seconds:       units,
					Weight:        weight,
					Price:         price,
					Percent:       percent,
					DestinationId: record[4],
				},
			}
		}
		actions[tag] = append(actions[tag], a)
	}
	log.Print("Actions:")
	log.Print(actions)
}

func loadActionTimings() {
	fp, err := os.Open(*actiontimingsFn)
	if err != nil {
		log.Printf("Could not open actions timings file: %v", err)
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
			log.Printf("Could not load the timing for tag: %v", record[2])
			continue
		}
		for _, t := range ts {
			at := &timespans.ActionTiming{
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

func loadActionTriggers() {
	fp, err := os.Open(*actiontriggersFn)
	if err != nil {
		log.Printf("Could not open destination balance actions file: %v", err)
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
		value, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			log.Printf("Could not parse action trigger value: %v", err)
			continue
		}
		at := &timespans.ActionTrigger{
			BalanceId:      record[1],
			ThresholdValue: value,
			DestinationId:  record[3],
			ActionsId:      record[4],
		}
		actionsTriggers[tag] = append(actionsTriggers[tag], at)
	}
	log.Print("Actions triggers:")
	log.Print(actionsTriggers)
}

func loadAccountActions() {
	fp, err := os.Open(*accountactionsFn)
	if err != nil {
		log.Printf("Could not open account actions file: %v", err)
		return
	}
	defer fp.Close()
	csvReader := csv.NewReader(fp)
	csvReader.Comma = sep
	csvReader.TrailingComma = true
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		if record[0] == "Tenant" {
			continue
		}
		tag := fmt.Sprintf("%s:%s:%s", record[0], record[1], record[2])
		aTriggers, exists := actionsTriggers[record[4]]
		if !exists {
			log.Printf("Could not get action triggers for tag %v", record[4])
			continue
		}
		aTimingsTag := record[3]
		ub := &timespans.UserBalance{
			Id:             tag,
			ActionTriggers: aTriggers,
		}
		accountActions[tag] = append(accountActions[tag], ub)

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
