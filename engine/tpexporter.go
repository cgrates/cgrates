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
	"archive/zip"
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"unicode/utf8"

	"github.com/cgrates/cgrates/utils"
)

var (
	TPExportFormats = []string{utils.CSV}
)

func NewTPExporter(storDb LoadStorage, tpID, expPath, fileFormat, sep string, compress bool) (*TPExporter, error) {
	if len(tpID) == 0 {
		return nil, errors.New("Missing TPid")
	}
	if !utils.IsSliceMember(TPExportFormats, fileFormat) {
		return nil, errors.New("Unsupported file format")
	}
	tpExp := &TPExporter{
		storDb:     storDb,
		tpID:       tpID,
		exportPath: expPath,
		fileFormat: fileFormat,
		compress:   compress,
		cacheBuff:  new(bytes.Buffer),
	}
	runeSep, _ := utf8.DecodeRuneInString(sep)
	if runeSep == utf8.RuneError {
		return nil, fmt.Errorf("Invalid field separator: %s", sep)
	} else {
		tpExp.sep = runeSep
	}
	if compress {
		if len(tpExp.exportPath) == 0 {
			tpExp.zipWritter = zip.NewWriter(tpExp.cacheBuff)
		} else {
			if fileOut, err := os.Create(path.Join(tpExp.exportPath, "tpexport.zip")); err != nil {
				return nil, err
			} else {
				tpExp.zipWritter = zip.NewWriter(fileOut)
			}
		}
	}
	return tpExp, nil
}

// Export TariffPlan to a folder
type TPExporter struct {
	storDb        LoadStorage   // StorDb connection handle
	tpID          string        // Load data on this tpid
	exportPath    string        // Directory path to export to
	fileFormat    string        // The file format <csv>
	sep           rune          // Separator in the csv file
	compress      bool          // Use ZIP to compress the folder
	cacheBuff     *bytes.Buffer // Will be written in case of no output folder is specified
	zipWritter    *zip.Writer   // Populated in case of needing to write zipped content
	exportedFiles []string
}

func (self *TPExporter) Run() error {
	self.removeFiles() // Make sure we clean the folder before starting with new one
	toExportMap := make(map[string]interface{})

	storData, err := self.storDb.GetTpTimings(self.tpID, "")
	if err != nil {
		return err
	}
	toExportMap[utils.TIMINGS_CSV] = storData

	storData, err := self.storDb.GetTpDestinations(self.tpID, "")
	if err != nil {
		return err
	}
	toExportMap[utils.DESTINATIONS_CSV] = storData

	storData, err := self.storDb.GetTpRates(self.tpID, "")
	if err != nil {
		return err
	}
	toExportMap[utils.RATES_CSV] = storData

	storData, err = self.storDb.GetTpRates(self.tpID, "")
	if err != nil {
		return err
	}
	toExportMap[utils.RATES_CSV] = storData

	storData, err = self.storDb.GetTpDestinationRates(self.tpID, "", nil)
	if err != nil {
		return err
	}
	toExportMap[utils.DESTINATION_RATES_CSV] = storData

	storData, err = self.storDb.GetTpRatingPlans(self.tpID, "", nil)
	if err != nil {
		return err
	}
	toExportMap[utils.RATING_PLANS_CSV] = storData

	storData, err = self.storDb.GetTpRatingProfiles(&utils.TPRatingProfile{TPid: self.tpID})
	if err != nil {
		return err
	}
	toExportMap[utils.RATING_PROFILE_CSV] = storData

	storData, err = self.storDb.GetTpSharedGroups(self.tpID, "")
	if err != nil {
		return err
	}
	toExportMap[utils.SHARED_GROUPS_CSV] = storData

	storData, err = self.storDb.GetTpActions(self.tpID, "")
	if err != nil {
		return err
	}
	toExportMap[utils.ACTIONS_CSV] = storData

	storData, err = self.storDb.GetTpActionPlans(self.tpID, "")
	if err != nil {
		return err
	}
	toExportMap[utils.ACTION_PLANS_CSV] = storData

	storData, err = self.storDb.GetTpActionTriggers(self.tpID, "")
	if err != nil {
		return err
	}
	toExportMap[utils.ACTIONS_TRIGGERS_CSV] = storData

	storData, err = self.storDb.GetTpAccountActions(&utils.TPAccountActions{TPid: self.tpID})
	if err != nil {
		return err
	}
	toExportMap[utils.ACCOUNT_ACTIONS_CSV] = storData

	storData, err = self.storDb.GetTpDerivedChargers(&utils.TPDerivedChargers{TPid: self.tpID})
	if err != nil {
		return err
	}
	toExportMap[utils.DERIVED_CHARGERS_CSV] = storData

	storData, err = self.storDb.GetTpCdrStats(self.tpID, "")
	if err != nil {
		return err
	}
	toExportMap[utils.CDR_STATS_CSV] = storData

	for fileName, storData := range toExportMap {
		for _, tpItem := range storData {
			exportedData = append(exportedData, tpItem)
		}
		if err := self.writeOut(fileName, exportedData); err != nil {
			self.removeFiles()
			return err
		}
		self.exportedFiles = append(self.exportedFiles, fileName)
	}

	if self.compress {
		if err := self.zipWritter.Close(); err != nil {
			return err
		}
	}
	return nil
}

// Some export did not end up well, remove the files here
func (self *TPExporter) removeFiles() error {
	if len(self.exportPath) == 0 {
		return nil
	}
	for _, fileName := range exportedFiles {
		os.Remove(path.Join(self.exportPath, fileName))
	}
	return nil
}

// General method to write the content out to a file on path or zip archive
func (self *TPExporter) writeOut(fileName string, tpData []interface{}) error {
	if len(tpData) == 0 {
		return nil
	}
	var fWriter io.Writer
	var writerOut utils.CgrRecordWriter
	var err error

	if self.compress {
		if fWriter, err = self.zipWritter.Create(fileName); err != nil {
			return err
		}
	} else if len(self.exportPath) != 0 {
		if f, err := os.Create(path.Join(self.exportPath, fileName)); err != nil {
			return err
		} else {
			fWriter = f
			defer f.Close()
		}

	} else {
		fWriter = new(bytes.Buffer)
	}

	switch self.fileFormat {
	case utils.CSV:
		csvWriter := csv.NewWriter(fWriter)
		csvWriter.Comma = self.sep
		writerOut = csvWriter
	default:
		writerOut = utils.NewCgrIORecordWriter(fWriter)
	}

	for _, tpItem := range tpData {
		record, err := csvDump(tpItem)
		if err != nil {
			return err
		}
		if err := writerOut.Write(record); err != nil {
			return err
		}
	}
	writerOut.Flush() // In case of .csv will dump data on hdd
	return nil
}

func (self *TPExporter) ExportStats() *utils.ExportedTPStats {
	return &utils.ExportedTPStats{ExportPath: self.exportPath, ExportedFiles: self.exportedFiles, Compressed: self.compress}
}

func (self *TPExporter) GetCacheBuffer() *bytes.Buffer {
	return self.cacheBuff
}
