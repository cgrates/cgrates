/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

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

package engine

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"path"
	"unicode/utf8"

	"github.com/cgrates/cgrates/utils"
)

var TPExportFormats = []string{utils.DRYRUN, utils.CSV}

func NewTPExporter(storDb LoadStorage, tpID, dir, fileFormat, sep string, dryRunBuff *bytes.Buffer) (*TPExporter, error) {
	if !utils.IsSliceMember(TPExportFormats, fileFormat) {
		return nil, errors.New("Unsupported file format")
	}
	tpDirExp := &TPExporter{
		storDb:       storDb,
		tpID:         tpID,
		dirPath:      dir,
		fileFormat:   fileFormat,
		dryRunBuffer: dryRunBuff,
	}
	runeSep, _ := utf8.DecodeRuneInString(sep)
	if runeSep == utf8.RuneError {
		return nil, fmt.Errorf("Invalid field separator: %s", sep)
	} else {
		tpDirExp.sep = runeSep
	}
	return tpDirExp, nil
}

// Export TariffPlan to a folder
type TPExporter struct {
	storDb       LoadStorage   // StorDb connection handle
	tpID         string        // Load data on this tpid
	dirPath      string        // Directory path to export to
	fileFormat   string        // The file format <csv>
	sep          rune          // Separator in the csv file
	dryRunBuffer *bytes.Buffer // Will be written in case of dryRun so we can read it from tests
}

func (self *TPExporter) Run() error {
	for _, fHandler := range []func() error{
		self.exportTimings,
		self.exportDestinations,
		self.exportRates,
		self.exportDestinationRates,
		self.exportRatingPlans,
		self.exportRatingProfiles,
		self.exportSharedGroups,
		self.exportActions,
		self.exportActionPlans,
		self.exportActionTriggers,
		self.exportAccountActions,
		self.exportDerivedChargers,
		self.exportCdrStats,
	} {
		if err := fHandler(); err != nil {
			self.removeFiles()
			return err
		}
	}
	return nil
}

// Some export did not end up well, remove the files here
func (self *TPExporter) removeFiles() error {
	return nil
}

// General method to write the content out to a file
func (self *TPExporter) writeOut(fileName string, tpData []utils.ExportedData) error {
	var writerOut utils.CgrRecordWriter
	if self.fileFormat == utils.DRYRUN {
		writerOut = utils.NewCgrIORecordWriter(self.dryRunBuffer)
	} else { // For now the only supported here is csv
		fileOut, err := os.Create(path.Join(self.dirPath, fileName))
		if err != nil {
			return err
		}
		defer fileOut.Close()
		csvWriter := csv.NewWriter(fileOut)
		csvWriter.Comma = self.sep
		writerOut = csvWriter
	}
	for _, tpItem := range tpData {
		for _, record := range tpItem.AsExportSlice() {
			if err := writerOut.Write(record); err != nil {
				return err
			}
		}
	}
	return nil
}

func (self *TPExporter) exportTimings() error {
	return nil
}

func (self *TPExporter) exportDestinations() error {
	return nil
}

func (self *TPExporter) exportRates() error {
	fileName := utils.RATES_CSV
	storData, err := self.storDb.GetTpRates(self.tpID, "")
	if err != nil {
		return nil
	}
	exportedData := make([]utils.ExportedData, len(storData))
	idx := 0
	for _, tpItem := range storData {
		exportedData[idx] = tpItem
		idx += 1
	}
	if err := self.writeOut(fileName, exportedData); err != nil {
		return err
	}
	return nil
}

func (self *TPExporter) exportDestinationRates() error {
	fileName := utils.DESTINATION_RATES_CSV
	storData, err := self.storDb.GetTpDestinationRates(self.tpID, "", nil)
	if err != nil {
		return nil
	}
	exportedData := make([]utils.ExportedData, len(storData))
	idx := 0
	for _, tpItem := range storData {
		exportedData[idx] = tpItem
		idx += 1
	}
	if err := self.writeOut(fileName, exportedData); err != nil {
		return err
	}
	return nil
}

func (self *TPExporter) exportRatingPlans() error {
	return nil
}

func (self *TPExporter) exportRatingProfiles() error {
	fileName := utils.RATING_PROFILES_CSV
	storData, err := self.storDb.GetTpRatingProfiles(&utils.TPRatingProfile{TPid: self.tpID})
	if err != nil {
		return nil
	}
	exportedData := make([]utils.ExportedData, len(storData))
	idx := 0
	for _, tpItem := range storData {
		exportedData[idx] = tpItem
		idx += 1
	}
	if err := self.writeOut(fileName, exportedData); err != nil {
		return err
	}
	return nil
}

func (self *TPExporter) exportSharedGroups() error {
	return nil
}

func (self *TPExporter) exportActions() error {
	return nil
}

func (self *TPExporter) exportActionPlans() error {
	return nil
}

func (self *TPExporter) exportActionTriggers() error {
	return nil
}

func (self *TPExporter) exportAccountActions() error {
	fileName := utils.ACCOUNT_ACTIONS_CSV
	storData, err := self.storDb.GetTpAccountActions(&utils.TPAccountActions{TPid: self.tpID})
	if err != nil {
		return nil
	}
	exportedData := make([]utils.ExportedData, len(storData))
	idx := 0
	for _, tpItem := range storData {
		exportedData[idx] = tpItem
		idx += 1
	}
	if err := self.writeOut(fileName, exportedData); err != nil {
		return err
	}
	return nil
}

func (self *TPExporter) exportDerivedChargers() error {
	fileName := utils.DERIVED_CHARGERS_CSV
	storData, err := self.storDb.GetTpDerivedChargers(&utils.TPDerivedChargers{TPid: self.tpID})
	if err != nil {
		return nil
	}
	exportedData := make([]utils.ExportedData, len(storData))
	idx := 0
	for _, tpItem := range storData {
		exportedData[idx] = tpItem
		idx += 1
	}
	if err := self.writeOut(fileName, exportedData); err != nil {
		return err
	}
	return nil
}

func (self *TPExporter) exportCdrStats() error {
	return nil
}
