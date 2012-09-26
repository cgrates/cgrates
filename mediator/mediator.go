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

type mediatorFieldIdxs []int

// Extends goconf to provide us the slice with indexes we need for multiple mediation
func (mfi *mediatorFieldIdxs) Load(idxs string) error {
	cfgStrIdxs := strings.Split(idxs, ",")
	if len(cfgStrIdxs) == 0 {
		return fmt.Errorf("Undefined %s", idxs)
	}
	for _, cfgStrIdx := range cfgStrIdxs {
		if cfgIntIdx, errConv := strconv.Atoi(cfgStrIdx); errConv != nil || cfgStrIdx == "" {
			return fmt.Errorf("All mediator index members (%s) must be ints", idxs)
		} else {
			*mfi = append(*mfi, cfgIntIdx)
		}
	}
	return nil
}

type Mediator struct {
	connector     timespans.Connector
	loggerDb      timespans.DataStorage
	skipDb        bool
	outputDir     string
	pseudoPrepaid bool
	directionIndexs,
	torIndexs,
	tenantIndexs,
	subjectIndexs,
	accountIndexs,
	destinationIndexs,
	timeStartIndexs,
	durationIndexs,
	uuidIndexs mediatorFieldIdxs
}

// Creates a new mediator object parsing the indexses
func NewMediator(connector timespans.Connector,
	loggerDb timespans.DataStorage,
	skipDb bool,
	outputDir string,
	pseudoPrepaid bool,
	directionIndexs, torIndexs, tenantIndexs, subjectIndexs, accountIndexs, destinationIndexs,
	timeStartIndexs, durationIndexs, uuidIndexs string) (m *Mediator, err error) {
	m = &Mediator{
		connector:     connector,
		loggerDb:      loggerDb,
		skipDb:        skipDb,
		outputDir:     outputDir,
		pseudoPrepaid: pseudoPrepaid,
	}
	idxs := []string{directionIndexs, torIndexs, tenantIndexs, subjectIndexs, accountIndexs,
		destinationIndexs, timeStartIndexs, durationIndexs, uuidIndexs}
	objs := []mediatorFieldIdxs{m.directionIndexs, m.torIndexs, m.tenantIndexs, m.subjectIndexs,
		m.accountIndexs, m.destinationIndexs, m.timeStartIndexs, m.durationIndexs, m.uuidIndexs}
	for i, o := range objs {
		err = o.Load(idxs[i])
		if err != nil {
			return
		}
	}
	if !m.validateIndexses() {
		err = fmt.Errorf("All members must have the same length")
	}
	return
}

// Make sure all indexes are having same lenght
func (m *Mediator) validateIndexses() bool {
	refLen := len(m.subjectIndexs)
	for _, fldIdxs := range []mediatorFieldIdxs{m.directionIndexs, m.torIndexs, m.tenantIndexs,
		m.accountIndexs, m.destinationIndexs, m.timeStartIndexs, m.durationIndexs, m.uuidIndexs} {
		if len(fldIdxs) != refLen {
			return false
		}
	}
	return true
}

// Watch the specified folder for file moves and parse the files on events
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

// Parse the files and get cost for every record
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
		for runIdx := range m.subjectIndexs { // Query costs for every run index given by subject
			if runIdx == 0 && !m.skipDb { // The first index is matching the session manager one
				cc, err = m.getCostsFromDB(record, runIdx)
				if err != nil || cc == nil { // Fallback on rater if no db record found
					cc, err = m.getCostsFromRater(record, runIdx)
				}
			} else {
				cc, err = m.getCostsFromRater(record, runIdx)
			}
			cost := "-1"
			if err != nil {
				timespans.Logger.Err(fmt.Sprintf("Could not get the cost for mediator record with uuid:%s and subject:%s - %s", record[m.uuidIndexs[runIdx]], record[m.subjectIndexs[runIdx]], err.Error()))
			} else {
				cost = strconv.FormatFloat(cc.ConnectFee+cc.Cost, 'f', -1, 64)
				timespans.Logger.Debug(fmt.Sprintf("Calculated for uuid:%s, subject:%s cost: %v", record[m.uuidIndexs[runIdx]], record[m.subjectIndexs[runIdx]], cost))
			}
			record = append(record, cost)
		}
		w.WriteString(strings.Join(record, ",") + "\n")
	}
	w.Flush()
	return
}

// Retrive the cost from logging database
func (m *Mediator) getCostsFromDB(record []string, runIdx int) (cc *timespans.CallCost, err error) {
	searchedUUID := record[m.uuidIndexs[runIdx]]
	cc, err = m.loggerDb.GetCallCostLog(searchedUUID, timespans.SESSION_MANAGER_SOURCE)
	return
}

// Retrive the cost from rater
func (m *Mediator) getCostsFromRater(record []string, runIdx int) (cc *timespans.CallCost, err error) {
	d, err := time.ParseDuration(record[m.durationIndexs[runIdx]] + "s")
	if err != nil {
		return
	}

	cc = &timespans.CallCost{}
	if d.Seconds() == 0 { // failed call,  returning empty callcost, no error
		return cc, nil
	}
	t1, err := time.Parse("2006-01-02 15:04:05", record[m.timeStartIndexs[runIdx]])
	if err != nil {
		return
	}
	cd := timespans.CallDescriptor{
		Direction:   "OUT", //record[m.directionIndexs[runIdx]] TODO: fix me
		Tenant:      record[m.tenantIndexs[runIdx]],
		TOR:         record[m.torIndexs[runIdx]],
		Subject:     record[m.subjectIndexs[runIdx]],
		Account:     record[m.accountIndexs[runIdx]],
		Destination: record[m.destinationIndexs[runIdx]],
		TimeStart:   t1,
		TimeEnd:     t1.Add(d)}
	if m.pseudoPrepaid {
		err = m.connector.Debit(cd, cc)
	} else {
		err = m.connector.GetCost(cd, cc)
	}
	if err != nil {
		m.loggerDb.LogError(record[m.uuidIndexs[runIdx]], timespans.MEDIATOR_SOURCE, err.Error())
	} else {
		m.loggerDb.LogCallCost(record[m.uuidIndexs[runIdx]], timespans.MEDIATOR_SOURCE, cc)
	}
	return
}
