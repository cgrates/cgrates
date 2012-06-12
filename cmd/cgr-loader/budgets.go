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
	primaryBalanceActions    []*timespans.Action
	destinatioBalanceActions []*timespans.Action
)

func loadPrimaryBalanceActions() {
	fp, err := os.Open(*primaryBalanceActionsFn)
	if err != nil {
		log.Printf("Could not open primary balance actions file: %v", err)
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
		log.Print(tag, primaryBalanceActions)
	}
}

func loadDestinationBalanceActions() {
	fp, err := os.Open(*destinationBalanceActionsFn)
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
		log.Print(tag, destinatioBalanceActions)
	}
}
