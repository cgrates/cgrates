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

var (
	TPExportFormats = []string{utils.DRYRUN, utils.CSV}
	exportedFiles   = []string{utils.TIMINGS_CSV, utils.DESTINATIONS_CSV, utils.RATES_CSV, utils.DESTINATION_RATES_CSV, utils.RATING_PLANS_CSV, utils.RATING_PROFILES_CSV,
		utils.SHARED_GROUPS_CSV, utils.ACTIONS_CSV, utils.ACTION_PLANS_CSV, utils.ACTION_TRIGGERS_CSV, utils.ACCOUNT_ACTIONS_CSV, utils.DERIVED_CHARGERS_CSV, utils.CDR_STATS_CSV}
)

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
	for _, fileName := range exportedFiles {
		os.Remove(path.Join(self.dirPath, fileName))
	}
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
	//fileName := exportedFiles[0] // Define it out of group so we make sure it is cleaned up by removeFiles
	return nil
}

func (self *TPExporter) exportDestinations() error {
	fileName := exportedFiles[1]
	storData, err := self.storDb.GetTpDestinations(self.tpID, "")
	if err != nil {
		return nil
	}
	exportedData := make([]utils.ExportedData, len(storData))
	idx := 0
	for _, dst := range storData {
		exportedData[idx] = &utils.TPDestination{TPid: self.tpID, DestinationId: dst.Id, Prefixes: dst.Prefixes}
		idx += 1
	}
	if err := self.writeOut(fileName, exportedData); err != nil {
		return err
	}
	return nil
}

func (self *TPExporter) exportRates() error {
	fileName := exportedFiles[2]
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
	fileName := exportedFiles[3]
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
	fileName := exportedFiles[4]
	storData, err := self.storDb.GetTpRatingPlans(self.tpID, "", nil)
	if err != nil {
		return nil
	}
	exportedData := make([]utils.ExportedData, len(storData))
	idx := 0
	for rpId, rpBinding := range storData {
		exportedData[idx] = &utils.TPRatingPlan{TPid: self.tpID, RatingPlanId: rpId, RatingPlanBindings: rpBinding}
		idx += 1
	}
	if err := self.writeOut(fileName, exportedData); err != nil {
		return err
	}
	return nil
}

func (self *TPExporter) exportRatingProfiles() error {
	fileName := exportedFiles[5]
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
	fileName := exportedFiles[6]
	storData, err := self.storDb.GetTpSharedGroups(self.tpID, "")
	if err != nil {
		return nil
	}
	exportedData := make([]utils.ExportedData, len(storData))
	idx := 0
	for sgId, sg := range storData {
		exportedData[idx] = &utils.TPSharedGroups{TPid: self.tpID, SharedGroupsId: sgId, SharedGroups: sg}
		idx += 1
	}
	if err := self.writeOut(fileName, exportedData); err != nil {
		return err
	}
	return nil
}

func (self *TPExporter) exportActions() error {
	fileName := exportedFiles[7]
	storData, err := self.storDb.GetTpActions(self.tpID, "")
	if err != nil {
		return nil
	}
	exportedData := make([]utils.ExportedData, len(storData))
	idx := 0
	for actsId, acts := range storData {
		exportedData[idx] = &utils.TPActions{TPid: self.tpID, ActionsId: actsId, Actions: acts}
		idx += 1
	}
	if err := self.writeOut(fileName, exportedData); err != nil {
		return err
	}
	return nil
}

func (self *TPExporter) exportActionPlans() error {
	fileName := exportedFiles[8]
	storData, err := self.storDb.GetTPActionTimings(self.tpID, "")
	if err != nil {
		return nil
	}
	exportedData := make([]utils.ExportedData, len(storData))
	idx := 0
	for apId, ats := range storData {
		exportedData[idx] = &utils.TPActionPlan{TPid: self.tpID, Id: apId, ActionPlan: ats}
		idx += 1
	}
	if err := self.writeOut(fileName, exportedData); err != nil {
		return err
	}
	return nil
}

func (self *TPExporter) exportActionTriggers() error {
	fileName := exportedFiles[9]
	storData, err := self.storDb.GetTpActionTriggers(self.tpID, "")
	if err != nil {
		return nil
	}
	exportedData := make([]utils.ExportedData, len(storData))
	idx := 0
	for atId, ats := range storData {
		exportedData[idx] = &utils.TPActionTriggers{TPid: self.tpID, ActionTriggersId: atId, ActionTriggers: ats}
		idx += 1
	}
	if err := self.writeOut(fileName, exportedData); err != nil {
		return err
	}
	return nil
}

func (self *TPExporter) exportAccountActions() error {
	fileName := exportedFiles[10]
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
	fileName := exportedFiles[11]
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
	fileName := exportedFiles[12]
	storData, err := self.storDb.GetTpCdrStats(self.tpID, "")
	if err != nil {
		return nil
	}
	exportedData := make([]utils.ExportedData, len(storData))
	idx := 0
	for cdrstId, cdrsts := range storData {
		exportedData[idx] = &utils.TPCdrStats{TPid: self.tpID, CdrStatsId: cdrstId, CdrStats: cdrsts}
		idx += 1
	}
	if err := self.writeOut(fileName, exportedData); err != nil {
		return err
	}
	return nil
}
