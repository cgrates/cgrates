/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
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
	toExportMap := make(map[string][]interface{})

	storDataTimings, err := self.storDb.GetTPTimings(self.tpID, "")
	if err != nil && err.Error() != utils.ErrNotFound.Error() {
		return err
	}
	storDataModelTimings := APItoModelTimings(storDataTimings)
	toExportMap[utils.TIMINGS_CSV] = make([]interface{}, len(storDataTimings))
	for i, sd := range storDataModelTimings {
		toExportMap[utils.TIMINGS_CSV][i] = sd
	}

	storDataDestinations, err := self.storDb.GetTPDestinations(self.tpID, "")
	if err != nil && err.Error() != utils.ErrNotFound.Error() {
		return err
	}
	toExportMap[utils.DESTINATIONS_CSV] = make([]interface{}, len(storDataDestinations))
	for i, sd := range storDataDestinations {
		sdModel := APItoModelDestination(sd)
		toExportMap[utils.DESTINATIONS_CSV][i] = sdModel[0]
	}

	storDataRates, err := self.storDb.GetTPRates(self.tpID, "")
	if err != nil && err.Error() != utils.ErrNotFound.Error() {
		return err
	}
	toExportMap[utils.RATES_CSV] = make([]interface{}, len(storDataRates))
	for i, sd := range storDataRates {
		sdModel := APItoModelRate(sd)
		toExportMap[utils.RATES_CSV][i] = sdModel[0]
	}

	storDataDestinationRates, err := self.storDb.GetTPDestinationRates(self.tpID, "", nil)
	if err != nil && err.Error() != utils.ErrNotFound.Error() {
		return err
	}
	toExportMap[utils.DESTINATION_RATES_CSV] = make([]interface{}, len(storDataDestinationRates))
	for i, sd := range storDataDestinationRates {
		sdModel := APItoModelDestinationRate(sd)
		toExportMap[utils.DESTINATION_RATES_CSV][i] = sdModel[0]
	}

	storDataRatingPlans, err := self.storDb.GetTPRatingPlans(self.tpID, "", nil)
	if err != nil && err.Error() != utils.ErrNotFound.Error() {
		return err
	}
	toExportMap[utils.RATING_PLANS_CSV] = make([]interface{}, len(storDataRatingPlans))
	for i, sd := range storDataRatingPlans {
		sdModel := APItoModelRatingPlan(sd)
		toExportMap[utils.RATING_PLANS_CSV][i] = sdModel[0]
	}

	storDataRatingProfiles, err := self.storDb.GetTPRatingProfiles(&utils.TPRatingProfile{TPid: self.tpID})
	if err != nil && err.Error() != utils.ErrNotFound.Error() {
		return err
	}
	toExportMap[utils.RATING_PROFILES_CSV] = make([]interface{}, len(storDataRatingProfiles))
	for i, sd := range storDataRatingProfiles {
		sdModel := APItoModelRatingProfile(sd)
		toExportMap[utils.RATING_PROFILES_CSV][i] = sdModel[0]
	}

	storDataSharedGroups, err := self.storDb.GetTPSharedGroups(self.tpID, "")
	if err != nil && err.Error() != utils.ErrNotFound.Error() {
		return err
	}

	toExportMap[utils.SHARED_GROUPS_CSV] = make([]interface{}, len(storDataSharedGroups))
	for i, sd := range storDataSharedGroups {
		sdModel := APItoModelSharedGroup(sd)
		toExportMap[utils.SHARED_GROUPS_CSV][i] = sdModel[0]
	}

	storDataActions, err := self.storDb.GetTPActions(self.tpID, "")
	if err != nil && err.Error() != utils.ErrNotFound.Error() {
		return err
	}
	toExportMap[utils.ACTIONS_CSV] = make([]interface{}, len(storDataActions))
	for i, sd := range storDataActions {
		sdModel := APItoModelAction(sd)
		toExportMap[utils.ACTIONS_CSV][i] = sdModel[0]
	}

	storDataActionPlans, err := self.storDb.GetTPActionPlans(self.tpID, "")
	if err != nil && err.Error() != utils.ErrNotFound.Error() {
		return err
	}
	toExportMap[utils.ACTION_PLANS_CSV] = make([]interface{}, len(storDataActionPlans))
	for i, sd := range storDataActionPlans {
		sdModel := APItoModelActionPlan(sd)
		toExportMap[utils.ACTION_PLANS_CSV][i] = sdModel[0]
	}

	storDataActionTriggers, err := self.storDb.GetTPActionTriggers(self.tpID, "")
	if err != nil && err.Error() != utils.ErrNotFound.Error() {
		return err
	}
	toExportMap[utils.ACTION_TRIGGERS_CSV] = make([]interface{}, len(storDataActionTriggers))
	for i, sd := range storDataActionTriggers {
		sdModel := APItoModelActionTrigger(sd)
		toExportMap[utils.ACTION_TRIGGERS_CSV][i] = sdModel[0]
	}

	storDataAccountActions, err := self.storDb.GetTPAccountActions(&utils.TPAccountActions{TPid: self.tpID})
	if err != nil && err.Error() != utils.ErrNotFound.Error() {
		return err
	}
	toExportMap[utils.ACCOUNT_ACTIONS_CSV] = make([]interface{}, len(storDataAccountActions))
	for i, sd := range storDataAccountActions {
		sdModel := APItoModelAccountAction(sd)
		toExportMap[utils.ACCOUNT_ACTIONS_CSV][i] = sdModel
	}

	storDataDerivedCharges, err := self.storDb.GetTPDerivedChargers(&utils.TPDerivedChargers{TPid: self.tpID})
	if err != nil && err.Error() != utils.ErrNotFound.Error() {
		return err
	}
	toExportMap[utils.DERIVED_CHARGERS_CSV] = make([]interface{}, len(storDataDerivedCharges))
	for i, sd := range storDataDerivedCharges {
		sdModel := APItoModelDerivedCharger(sd)
		toExportMap[utils.DERIVED_CHARGERS_CSV][i] = sdModel[0]
	}

	storDataCdrStats, err := self.storDb.GetTPCdrStats(self.tpID, "")
	if err != nil && err.Error() != utils.ErrNotFound.Error() {
		return err
	}
	toExportMap[utils.CDR_STATS_CSV] = make([]interface{}, len(storDataCdrStats))
	for i, sd := range storDataCdrStats {
		sdModel := APItoModelCdrStat(sd)
		toExportMap[utils.CDR_STATS_CSV][i] = sdModel[0]
	}

	storDataResources, err := self.storDb.GetTPResources(self.tpID, "")
	if err != nil && err.Error() != utils.ErrNotFound.Error() {
		return err
	}
	toExportMap[utils.ResourcesCsv] = make([]interface{}, len(storDataResources))
	for i, sd := range storDataResources {
		sdModel := APItoModelResource(sd)
		toExportMap[utils.ResourcesCsv][i] = sdModel[0]
	}

	storDataStats, err := self.storDb.GetTPStats(self.tpID, "")
	if err != nil && err.Error() != utils.ErrNotFound.Error() {
		return err
	}

	toExportMap[utils.StatsCsv] = make([]interface{}, len(storDataStats))
	for i, sd := range storDataStats {
		sdModel := APItoModelStats(sd)
		toExportMap[utils.StatsCsv][i] = sdModel[0]
	}

	storDataThresholds, err := self.storDb.GetTPThreshold(self.tpID, "")
	if err != nil && err.Error() != utils.ErrNotFound.Error() {
		return err
	}
	toExportMap[utils.ThresholdsCsv] = make([]interface{}, len(storDataThresholds))
	for i, sd := range storDataThresholds {
		sdModel := APItoModelTPThreshold(sd)
		toExportMap[utils.ThresholdsCsv][i] = sdModel[0]
	}

	storDataFilters, err := self.storDb.GetTPFilter(self.tpID, "")
	if err != nil && err.Error() != utils.ErrNotFound.Error() {
		return err
	}
	toExportMap[utils.FiltersCsv] = make([]interface{}, len(storDataFilters))
	for i, sd := range storDataFilters {
		sdModel := APItoModelTPFilter(sd)
		toExportMap[utils.FiltersCsv][i] = sdModel[0]
	}

	storDataUsers, err := self.storDb.GetTPUsers(&utils.TPUsers{TPid: self.tpID})
	if err != nil && err.Error() != utils.ErrNotFound.Error() {
		return err
	}
	toExportMap[utils.USERS_CSV] = make([]interface{}, len(storDataUsers))
	for i, sd := range storDataUsers {
		sdModel := APItoModelUsers(sd)
		toExportMap[utils.USERS_CSV][i] = sdModel[0]
	}

	for fileName, storData := range toExportMap {
		if err := self.writeOut(fileName, storData); err != nil {
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
	for _, fileName := range self.exportedFiles {
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
