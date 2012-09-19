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

package mediator

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/cgrates/cgrates/inotify"
	"github.com/cgrates/cgrates/timespans"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

type csvindex int

type Mediator struct {
	connector timespans.Connector
	loggerDb  timespans.DataStorage
	skipDb    bool
	outputDir string
	directionIndex,
	torIndex,
	tenantIndex,
	subjectIndex,
	accountIndex,
	destinationIndex,
	timeStartIndex,
	timeEndIndex csvindex
}

func NewMediator(connector timespans.Connector, loggerDb timespans.DataStorage, skipDb bool, outputDir, directionIndex, torIndex, tenantIndex, subjectIndex, accountIndex, destinationIndex, timeStartIndex, timeEndIndex string) (*Mediator, error) {
	m := &Mediator{
		connector: connector,
		loggerDb:  loggerDb,
		skipDb:    skipDb,
		outputDir: outputDir,
	}
	i, err := strconv.Atoi(directionIndex)
	if err != nil {
		return nil, err
	}
	m.directionIndex = csvindex(i)
	i, err = strconv.Atoi(torIndex)
	if err != nil {
		return nil, err
	}
	m.torIndex = csvindex(i)
	i, err = strconv.Atoi(tenantIndex)
	if err != nil {
		return nil, err
	}
	m.tenantIndex = csvindex(i)
	i, err = strconv.Atoi(subjectIndex)
	if err != nil {
		return nil, err
	}
	m.subjectIndex = csvindex(i)
	i, err = strconv.Atoi(accountIndex)
	if err != nil {
		return nil, err
	}
	m.accountIndex = csvindex(i)
	i, err = strconv.Atoi(destinationIndex)
	if err != nil {
		return nil, err
	}
	m.destinationIndex = csvindex(i)
	i, err = strconv.Atoi(timeStartIndex)
	if err != nil {
		return nil, err
	}
	m.timeStartIndex = csvindex(i)
	i, err = strconv.Atoi(timeEndIndex)
	if err != nil {
		return nil, err
	}
	m.timeEndIndex = csvindex(i)
	return m, nil
}

func (m *Mediator) TrackCDRFiles(cdrPath string) (err error) {
	watcher, err := inotify.NewWatcher()
	if err != nil {
		return
	}
	err = watcher.Watch(cdrPath)
	if err != nil {
		return
	}
	timespans.Logger.Info(fmt.Sprintf("Monitoring %v for file moves.", cdrPath))
	for {
		select {
		case ev := <-watcher.Event:
			if ev.Mask&inotify.IN_MOVED_TO != 0 {
				timespans.Logger.Info(fmt.Sprintf("Started to parse %v", ev.Name))
				err = m.parseCSV(ev.Name)
				if err != nil {
					return err
				}
			}
		case err := <-watcher.Error:
			timespans.Logger.Err(fmt.Sprintf("Inotify error: ", err))
		}
	}
	return
}

func (m *Mediator) parseCSV(cdrfn string) (err error) {
	flag.Parse()
	file, err := os.Open(cdrfn)
	defer file.Close()
	if err != nil {
		timespans.Logger.Crit(err.Error())
		os.Exit(1)
	}
	csvReader := csv.NewReader(bufio.NewReader(file))

	_, fn := path.Split(cdrfn)
	fout, err := os.Create(path.Join(m.outputDir, fn))
	if err != nil {
		return err
	}
	defer fout.Close()

	w := bufio.NewWriter(fout)

	for record, ok := csvReader.Read(); ok == nil; record, ok = csvReader.Read() {
		//t, _ := time.Parse("2012-05-21 17:48:20", record[5])		
		var cc *timespans.CallCost
		if !m.skipDb {
			cc, err = m.GetCostsFromDB(record)
		} else {
			cc, err = m.GetCostsFromRater(record)

		}
		if err != nil {
			timespans.Logger.Err(fmt.Sprintf("Could not get the cost for mediator record (%v): %v", record, err))
		} else {
			record = append(record, strconv.FormatFloat(cc.ConnectFee+cc.Cost, 'f', -1, 64))
		}
		w.WriteString(strings.Join(record, ","))
	}
	return
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
	t1, err := time.Parse("2012-05-21 17:48:20", record[m.timeStartIndex])
	if err != nil {
		return
	}
	t2, err := time.Parse("2012-05-21 17:48:20", record[m.timeEndIndex])
	if err != nil {
		return
	}
	cd := timespans.CallDescriptor{
		Direction:   "OUT", //record[m.directionIndex] TODO: fix me
		Tenant:      record[m.tenantIndex],
		TOR:         record[m.torIndex],
		Subject:     record[m.subjectIndex],
		Account:     record[m.accountIndex],
		Destination: record[m.destinationIndex],
		TimeStart:   t1,
		TimeEnd:     t2}
	err = m.connector.GetCost(cd, cc)
	return
}
