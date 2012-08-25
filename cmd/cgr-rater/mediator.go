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
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/cgrates/cgrates/sessionmanager"
	"github.com/cgrates/cgrates/timespans"
	"os"
	"time"
)

type Mediator struct {
	Connector sessionmanager.Connector
	loggerDb  timespans.DataStorage
	SkipDb    bool
}

/*func readDbRecord(db *sql.DB, searchedUUID string) (cc *timespans.CallCost, timespansText string, err error) {

}*/

func (m *Mediator) parseCSV() {
	flag.Parse()
	file, err := os.Open(mediator_cdr_file)
	defer file.Close()
	if err != nil {
		timespans.Logger.Crit(err.Error())
		os.Exit(1)
	}
	csvReader := csv.NewReader(bufio.NewReader(file))

	for record, ok := csvReader.Read(); ok == nil; record, ok = csvReader.Read() {
		//t, _ := time.Parse("2012-05-21 17:48:20", record[5])		
		var cc *timespans.CallCost
		if !m.SkipDb {
			cc, err = m.GetCostsFromDB(record)
		} else {
			cc, err = m.GetCostsFromRater(record)

		}
		if err != nil {
			timespans.Logger.Err(fmt.Sprintf("Could not get the cost for mediator record (%v): %v", record, err))
		}
		fmt.Println(cc)
	}
}

func (m *Mediator) GetCostsFromDB(record []string) (cc *timespans.CallCost, err error) {
	searchedUUID := record[10]
	cc, err = m.loggerDb.GetCallCostLog(searchedUUID)
	if err != nil {
		cc, err = m.GetCostsFromRater(record)
	}
	return
}

func (m *Mediator) GetCostsFromRater(record []string) (cc *timespans.CallCost, err error) {
	tenant := record[0]
	subject := record[1]
	dest := record[2]
	t1, _ := time.Parse("2012-05-21 17:48:20", record[5])
	t2, _ := time.Parse("2012-05-21 17:48:20", record[6])
	cd := timespans.CallDescriptor{
		Direction:   "OUT",
		Tenant:      tenant,
		TOR:         "0",
		Subject:     subject,
		Destination: dest,
		TimeStart:   t1,
		TimeEnd:     t2}
	err = m.Connector.GetCost(cd, cc)
	return
}
