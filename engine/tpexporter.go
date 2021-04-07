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

func NewTPExporter(storDB LoadStorage, tpID, expPath, fileFormat, sep string, compress bool) (*TPExporter, error) {
	if len(tpID) == 0 {
		return nil, errors.New("Missing TPid")
	}
	if utils.CSV != fileFormat {
		return nil, errors.New("Unsupported file format")
	}
	tpExp := &TPExporter{
		storDB:     storDB,
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
	storDB        LoadStorage   // StorDb connection handle
	tpID          string        // Load data on this tpid
	exportPath    string        // Directory path to export to
	fileFormat    string        // The file format <csv>
	sep           rune          // Separator in the csv file
	compress      bool          // Use ZIP to compress the folder
	cacheBuff     *bytes.Buffer // Will be written in case of no output folder is specified
	zipWritter    *zip.Writer   // Populated in case of needing to write zipped content
	exportedFiles []string
}

func (tpExp *TPExporter) Run() error {
	tpExp.removeFiles() // Make sure we clean the folder before starting with new one
	var withError bool
	toExportMap := make(map[string][]interface{})

	storDataTimings, err := tpExp.storDB.GetTPTimings(tpExp.tpID, "")
	if err != nil && err.Error() != utils.ErrNotFound.Error() {
		utils.Logger.Warning(fmt.Sprintf("<%s> error: %s, when getting %s from stordb for export", utils.ApierS, err, utils.TpTiming))
		withError = true

	}
	storDataModelTimings := APItoModelTimings(storDataTimings)
	for i, sd := range storDataModelTimings {
		toExportMap[utils.TimingsCsv][i] = sd
	}

	storDataDestinations, err := tpExp.storDB.GetTPDestinations(tpExp.tpID, "")
	if err != nil && err.Error() != utils.ErrNotFound.Error() {
		utils.Logger.Warning(fmt.Sprintf("<%s> error: %s, when getting %s from stordb for export", utils.ApierS, err, utils.TpDestinations))
		withError = true
	}
	for _, sd := range storDataDestinations {
		sdModels := APItoModelDestination(sd)
		for _, sdModel := range sdModels {
			toExportMap[utils.DestinationsCsv] = append(toExportMap[utils.DestinationsCsv], sdModel)
		}
	}

	storDataResources, err := tpExp.storDB.GetTPResources(tpExp.tpID, "", "")
	if err != nil && err.Error() != utils.ErrNotFound.Error() {
		utils.Logger.Warning(fmt.Sprintf("<%s> error: %s, when getting %s from stordb for export", utils.ApierS, err, utils.TpResources))
		withError = true
	}
	for _, sd := range storDataResources {
		sdModels := APItoModelResource(sd)
		for _, sdModel := range sdModels {
			toExportMap[utils.ResourcesCsv] = append(toExportMap[utils.ResourcesCsv], sdModel)
		}
	}

	storDataStats, err := tpExp.storDB.GetTPStats(tpExp.tpID, "", "")
	if err != nil && err.Error() != utils.ErrNotFound.Error() {
		utils.Logger.Warning(fmt.Sprintf("<%s> error: %s, when getting %s from stordb for export", utils.ApierS, err, utils.TpStats))
		withError = true
	}
	for _, sd := range storDataStats {
		sdModels := APItoModelStats(sd)
		for _, sdModel := range sdModels {
			toExportMap[utils.StatsCsv] = append(toExportMap[utils.StatsCsv], sdModel)
		}
	}

	storDataThresholds, err := tpExp.storDB.GetTPThresholds(tpExp.tpID, "", "")
	if err != nil && err.Error() != utils.ErrNotFound.Error() {
		utils.Logger.Warning(fmt.Sprintf("<%s> error: %s, when getting %s from stordb for export", utils.ApierS, err, utils.TpThresholds))
		withError = true
	}
	for _, sd := range storDataThresholds {
		sdModels := APItoModelTPThreshold(sd)
		for _, sdModel := range sdModels {
			toExportMap[utils.ThresholdsCsv] = append(toExportMap[utils.ThresholdsCsv], sdModel)
		}
	}

	storDataFilters, err := tpExp.storDB.GetTPFilters(tpExp.tpID, "", "")
	if err != nil && err.Error() != utils.ErrNotFound.Error() {
		utils.Logger.Warning(fmt.Sprintf("<%s> error: %s, when getting %s from stordb for export", utils.ApierS, err, utils.TpFilters))
		withError = true
	}
	for _, sd := range storDataFilters {
		sdModels := APItoModelTPFilter(sd)
		for _, sdModel := range sdModels {
			toExportMap[utils.FiltersCsv] = append(toExportMap[utils.FiltersCsv], sdModel)
		}
	}

	storDataRoutes, err := tpExp.storDB.GetTPRoutes(tpExp.tpID, "", "")
	if err != nil && err.Error() != utils.ErrNotFound.Error() {
		utils.Logger.Warning(fmt.Sprintf("<%s> error: %s, when getting %s from stordb for export", utils.ApierS, err, utils.TpRoutes))
		withError = true
	}
	for _, sd := range storDataRoutes {
		sdModels := APItoModelTPRoutes(sd)
		for _, sdModel := range sdModels {
			toExportMap[utils.RoutesCsv] = append(toExportMap[utils.RoutesCsv], sdModel)
		}
	}

	storeDataAttributes, err := tpExp.storDB.GetTPAttributes(tpExp.tpID, "", "")
	if err != nil && err.Error() != utils.ErrNotFound.Error() {
		utils.Logger.Warning(fmt.Sprintf("<%s> error: %s, when getting %s from stordb for export", utils.ApierS, err, utils.TpAttributes))
		withError = true
	}
	for _, sd := range storeDataAttributes {
		sdModels := APItoModelTPAttribute(sd)
		for _, sdModel := range sdModels {
			toExportMap[utils.AttributesCsv] = append(toExportMap[utils.AttributesCsv], sdModel)
		}
	}

	storDataChargers, err := tpExp.storDB.GetTPChargers(tpExp.tpID, "", "")
	if err != nil && err.Error() != utils.ErrNotFound.Error() {
		utils.Logger.Warning(fmt.Sprintf("<%s> error: %s, when getting %s from stordb for export", utils.ApierS, err, utils.TpChargers))
		withError = true
	}
	for _, sd := range storDataChargers {
		sdModels := APItoModelTPCharger(sd)
		for _, sdModel := range sdModels {
			toExportMap[utils.ChargersCsv] = append(toExportMap[utils.ChargersCsv], sdModel)
		}
	}

	storDataDispatcherProfiles, err := tpExp.storDB.GetTPDispatcherProfiles(tpExp.tpID, "", "")
	if err != nil && err.Error() != utils.ErrNotFound.Error() {
		utils.Logger.Warning(fmt.Sprintf("<%s> error: %s, when getting %s from stordb for export", utils.ApierS, err, utils.TpDispatcherProfiles))
		withError = true
	}
	for _, sd := range storDataDispatcherProfiles {
		sdModels := APItoModelTPDispatcherProfile(sd)
		for _, sdModel := range sdModels {
			toExportMap[utils.DispatcherProfilesCsv] = append(toExportMap[utils.DispatcherProfilesCsv], sdModel)
		}
	}

	storDataDispatcherHosts, err := tpExp.storDB.GetTPDispatcherHosts(tpExp.tpID, "", "")
	if err != nil && err.Error() != utils.ErrNotFound.Error() {
		utils.Logger.Warning(fmt.Sprintf("<%s> error: %s, when getting %s from stordb for export", utils.ApierS, err, utils.TpDispatcherHosts))
		withError = true
	}
	for _, sd := range storDataDispatcherHosts {
		toExportMap[utils.DispatcherHostsCsv] = append(toExportMap[utils.DispatcherHostsCsv], APItoModelTPDispatcherHost(sd))
	}

	storDataRateProfiles, err := tpExp.storDB.GetTPRateProfiles(tpExp.tpID, "", "")
	if err != nil && err.Error() != utils.ErrNotFound.Error() {
		utils.Logger.Warning(fmt.Sprintf("<%s> error: %s, when getting %s from stordb for export", utils.ApierS, err, utils.TpRateProfiles))
		withError = true
	}
	for _, sd := range storDataRateProfiles {
		sdModels := APItoModelTPRateProfile(sd)
		for _, sdModel := range sdModels {
			toExportMap[utils.RateProfilesCsv] = append(toExportMap[utils.RateProfilesCsv], sdModel)
		}
	}

	storDataActionProfiles, err := tpExp.storDB.GetTPActionProfiles(tpExp.tpID, "", "")
	if err != nil && err.Error() != utils.ErrNotFound.Error() {
		utils.Logger.Warning(fmt.Sprintf("<%s> error: %s, when getting %s from stordb for export", utils.ApierS, err, utils.TpActionProfiles))
		withError = true
	}
	for _, sd := range storDataActionProfiles {
		sdModels := APItoModelTPActionProfile(sd)
		for _, sdModel := range sdModels {
			toExportMap[utils.ActionProfilesCsv] = append(toExportMap[utils.ActionProfilesCsv], sdModel)
		}
	}

	storDataAccountProfiles, err := tpExp.storDB.GetTPAccounts(tpExp.tpID, "", "")
	if err != nil && err.Error() != utils.ErrNotFound.Error() {
		utils.Logger.Warning(fmt.Sprintf("<%s> error: %s, when getting %s from stordb for export", utils.ApierS, err, utils.TpAccountProfiles))
		withError = true
	}
	for _, sd := range storDataAccountProfiles {
		sdModels := APItoModelTPAccount(sd)
		for _, sdModel := range sdModels {
			toExportMap[utils.AccountProfilesCsv] = append(toExportMap[utils.AccountProfilesCsv], sdModel)
		}
	}

	if len(toExportMap) == 0 { // if we don't have anything to export we return not found error
		return utils.ErrNotFound
	}

	for fileName, storData := range toExportMap {
		if err := tpExp.writeOut(fileName, storData); err != nil {
			tpExp.removeFiles()
			return err
		}
		tpExp.exportedFiles = append(tpExp.exportedFiles, fileName)
	}

	if tpExp.compress {
		if err := tpExp.zipWritter.Close(); err != nil {
			return err
		}
	}
	if withError { // if we export something but have error we return partially executed
		return utils.ErrPartiallyExecuted
	}
	return nil
}

// Some export did not end up well, remove the files here
func (tpExp *TPExporter) removeFiles() error {
	if len(tpExp.exportPath) == 0 {
		return nil
	}
	for _, fileName := range tpExp.exportedFiles {
		os.Remove(path.Join(tpExp.exportPath, fileName))
	}
	return nil
}

// General method to write the content out to a file on path or zip archive
func (tpExp *TPExporter) writeOut(fileName string, tpData []interface{}) error {
	if len(tpData) == 0 {
		return nil
	}
	var fWriter io.Writer
	var writerOut utils.NopFlushWriter
	var err error

	if tpExp.compress {
		if fWriter, err = tpExp.zipWritter.Create(fileName); err != nil {
			return err
		}
	} else if len(tpExp.exportPath) != 0 {
		if f, err := os.Create(path.Join(tpExp.exportPath, fileName)); err != nil {
			return err
		} else {
			fWriter = f
			defer f.Close()
		}

	} else {
		fWriter = new(bytes.Buffer)
	}

	switch tpExp.fileFormat {
	case utils.CSV:
		csvWriter := csv.NewWriter(fWriter)
		csvWriter.Comma = tpExp.sep
		writerOut = csvWriter
	default:
		writerOut = utils.NewNopFlushWriter(fWriter)
	}
	for _, tpItem := range tpData {
		record, err := CsvDump(tpItem)
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

func (tpExp *TPExporter) ExportStats() *utils.ExportedTPStats {
	return &utils.ExportedTPStats{ExportPath: tpExp.exportPath, ExportedFiles: tpExp.exportedFiles, Compressed: tpExp.compress}
}

func (tpExp *TPExporter) GetCacheBuffer() *bytes.Buffer {
	return tpExp.cacheBuff
}
