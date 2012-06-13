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
)

var (
	actions         = make(map[string][]*timespans.Action)
	actionsTimings  []*timespans.Action
	actionsTriggers []*timespans.Action
	accountActions  []*timespans.Action
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
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		tag := record[0]
		if tag == "Tag" {
			// skip header line
			continue
		}
		//primaryBalanceActions = append(primaryBalanceActions, record[1:]...)
		log.Print(tag, actions)
	}
}

func loadActionsTimings() {
	fp, err := os.Open(*actionstimingsFn)
	if err != nil {
		log.Printf("Could not open actions timings file: %v", err)
		return
	}
	defer fp.Close()
	csvReader := csv.NewReader(fp)
	csvReader.Comma = sep
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		tag := record[0]
		if tag == "Tag" {
			// skip header line
			continue
		}
		//destinatioBalanceActions = append(destinatioBalanceActions, record[1:]...)
		log.Print(tag, actionsTimings)
	}
}

func loadActionsTriggers() {
	fp, err := os.Open(*actionstriggersFn)
	if err != nil {
		log.Printf("Could not open destination balance actions file: %v", err)
		return
	}
	defer fp.Close()
	csvReader := csv.NewReader(fp)
	csvReader.Comma = sep
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		tag := record[0]
		if tag == "Tag" {
			// skip header line
			continue
		}
		//destinatioBalanceActions = append(destinatioBalanceActions, record[1:]...)
		log.Print(tag, actionsTriggers)
	}
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
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		tag := record[0]
		if tag == "Tag" {
			// skip header line
			continue
		}
		//destinatioBalanceActions = append(destinatioBalanceActions, record[1:]...)
		log.Print(tag, accountActions)
	}
}
