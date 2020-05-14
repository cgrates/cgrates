/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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

package ees

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func NewFileCSVee(cgrCfg *config.CGRConfig, cfgIdx int, filterS *engine.FilterS) (fCsv *FileCSVee, err error) {
	fCsv = &FileCSVee{cgrCfg: cgrCfg, cfgIdx: cfgIdx, filterS: filterS}
	err = fCsv.init()
	return
}

// FileCSVee implements EventExporter interface for .csv files
type FileCSVee struct {
	id        string
	cgrCfg    *config.CGRConfig
	cfgIdx    int // index of config instance within ERsCfg.Readers
	filterS   *engine.FilterS
	file      *os.File
	csvWriter *csv.Writer

	firstEventATime, lastEventATime time.Time
	numberOfEvents                  int
	totalDuration, totalDataUsage, totalSmsUsage,
	totalMmsUsage, totalGenericUsage time.Duration
	totalCost                       float64
	firstExpOrderID, lastExpOrderID int64
	positiveExports                 utils.StringSet
	negativeExports                 utils.StringSet
}

// init will create all the necessary dependencies, including opening the file
func (fCsv *FileCSVee) init() (err error) {
	if fCsv.file, err = os.Create(fCsv.cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].ExportPath); err != nil {
		return
	}
	fCsv.csvWriter = csv.NewWriter(fCsv.file)
	fCsv.csvWriter.Comma = utils.CSV_SEP
	if len(fCsv.cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].FieldSep) > 0 {
		fCsv.csvWriter.Comma = rune(fCsv.cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].FieldSep[0])
	}
	fCsv.positiveExports = utils.StringSet{}
	fCsv.negativeExports = utils.StringSet{}
	return fCsv.composeHeader()
}

// ID returns the identificator of this exporter
func (fCsv *FileCSVee) ID() string {
	return fCsv.id
}

// OnEvicted implements EventExporter, doing the cleanup before exit
func (fCsv *FileCSVee) OnEvicted(_ string, _ interface{}) {
	// verify if we need to add the trailer
	if err := fCsv.composeTrailer(); err != nil {
		utils.Logger.Warning(fmt.Sprintf("<%s> Exporter with id: <%s> received error: <%s> when composed trailer",
			utils.EventExporterS, fCsv.id, err.Error()))
	}
	fCsv.csvWriter.Flush()
	if err := fCsv.file.Close(); err != nil {
		utils.Logger.Warning(fmt.Sprintf("<%s> Exporter with id: <%s> received error: <%s> when closing the file",
			utils.EventExporterS, fCsv.id, err.Error()))
	}
	return
}

// ExportEvent implements EventExporter
func (fCsv *FileCSVee) ExportEvent(cgrEv *utils.CGREvent) (err error) {
	// convert cgrEvent in export record
	fCsv.numberOfEvents++
	var csvRecord []string
	navMp := config.NewNavigableMap(map[string]interface{}{
		utils.MetaReq: cgrEv.Event,
	})
	for _, cfgFld := range fCsv.cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].ContentFields {
		if pass, err := fCsv.filterS.Pass(cgrEv.Tenant, cfgFld.Filters,
			navMp); err != nil || !pass {
			continue
		}
		val, err := cfgFld.Value.ParseDataProvider(navMp, fCsv.cgrCfg.GeneralCfg().RSRSep)
		if err != nil {
			fCsv.negativeExports.Add(cgrEv.ID)
			return err
		}
		csvRecord = append(csvRecord, val)
	}
	if aTime, err := cgrEv.FieldAsTime(utils.AnswerTime, fCsv.cgrCfg.GeneralCfg().DefaultTimezone); err == nil {
		if fCsv.firstEventATime.IsZero() || fCsv.firstEventATime.Before(aTime) {
			fCsv.firstEventATime = aTime
		}
		if aTime.After(fCsv.lastEventATime) {
			fCsv.lastEventATime = aTime
		}
	}
	if oID, err := cgrEv.FieldAsInt64(utils.OrderID); err == nil {
		if fCsv.firstExpOrderID > oID || fCsv.firstExpOrderID == 0 {
			fCsv.firstExpOrderID = oID
		}
		if fCsv.lastExpOrderID < oID {
			fCsv.lastExpOrderID = oID
		}
	}
	if cost, err := cgrEv.FieldAsFloat64(utils.Cost); err == nil {
		fCsv.totalCost += cost
	}
	if tor, err := cgrEv.FieldAsString(utils.ToR); err == nil {
		if usage, err := cgrEv.FieldAsDuration(utils.Usage); err == nil {
			switch tor {
			case utils.VOICE:
				fCsv.totalDuration += usage
			case utils.SMS:
				fCsv.totalSmsUsage += usage
			case utils.MMS:
				fCsv.totalMmsUsage += usage
			case utils.GENERIC:
				fCsv.totalGenericUsage += usage
			case utils.DATA:
				fCsv.totalDataUsage += usage
			}
		}
	}
	fCsv.positiveExports.Add(cgrEv.ID)
	fCsv.csvWriter.Write(csvRecord)
	return
}

// Compose and cache the header
func (fCsv *FileCSVee) composeHeader() (err error) {
	if len(fCsv.cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].HeaderFields) == 0 {
		return
	}
	var csvRecord []string
	for _, cfgFld := range fCsv.cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].HeaderFields {
		val, err := cfgFld.Value.ParseValue(utils.EmptyString)
		if err != nil {
			return err
		}
		csvRecord = append(csvRecord, val)
	}
	return fCsv.csvWriter.Write(csvRecord)
}

// Compose and cache the trailer
func (fCsv *FileCSVee) composeTrailer() (err error) {
	if len(fCsv.cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].TrailerFields) == 0 {
		return
	}
	var csvRecord []string
	for _, cfgFld := range fCsv.cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].TrailerFields {
		switch cfgFld.Type {
		case utils.MetaExportID:
			csvRecord = append(csvRecord, fCsv.id)
		case utils.MetaTimeNow:
			csvRecord = append(csvRecord, time.Now().String())
		case utils.MetaFirstEventATime:
			csvRecord = append(csvRecord, fCsv.firstEventATime.Format(cfgFld.Layout))
		case utils.MetaLastEventATime:
			csvRecord = append(csvRecord, fCsv.lastEventATime.Format(cfgFld.Layout))
		case utils.MetaEventNumber:
			csvRecord = append(csvRecord, strconv.Itoa(fCsv.numberOfEvents))
		case utils.MetaEventCost:
			rounding := fCsv.cgrCfg.GeneralCfg().RoundingDecimals
			if cfgFld.RoundingDecimals != nil {
				rounding = *cfgFld.RoundingDecimals
			}
			csvRecord = append(csvRecord, strconv.FormatFloat(utils.Round(fCsv.totalCost,
				rounding, utils.ROUNDING_MIDDLE), 'f', -1, 64))
		case utils.MetaVoiceUsage:
			csvRecord = append(csvRecord, fCsv.totalDuration.String())
		case utils.MetaDataUsage:
			csvRecord = append(csvRecord, strconv.Itoa(int(fCsv.totalDataUsage.Nanoseconds())))
		case utils.MetaSMSUsage:
			csvRecord = append(csvRecord, strconv.Itoa(int(fCsv.totalSmsUsage.Nanoseconds())))
		case utils.MetaMMSUsage:
			csvRecord = append(csvRecord, strconv.Itoa(int(fCsv.totalMmsUsage.Nanoseconds())))
		case utils.MetaGenericUsage:
			csvRecord = append(csvRecord, strconv.Itoa(int(fCsv.totalGenericUsage.Nanoseconds())))
		case utils.MetaNegativeExports:
			csvRecord = append(csvRecord, strconv.Itoa(len(fCsv.negativeExports.AsSlice())))
		case utils.MetaPositiveExports:
			csvRecord = append(csvRecord, strconv.Itoa(len(fCsv.positiveExports.AsSlice())))
		}
	}
	return fCsv.csvWriter.Write(csvRecord)
}
