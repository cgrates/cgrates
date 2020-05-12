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
	"os"
	"strings"
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
	id         string
	cgrCfg     *config.CGRConfig
	cfgIdx     int // index of config instance within ERsCfg.Readers
	filterS    *engine.FilterS
	file       *os.File
	csvWriter  *csv.Writer
	csvRecords [][]string
	//hdrFields
	//trlFields
	//cntFields
	firstEventATime, lastEventATime time.Time
	numberOfRecords                 int
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
	return
}

// ID returns the identificator of this exporter
func (fCsv *FileCSVee) ID() string {
	return fCsv.id
}

// OnEvicted implements EventExporter, doing the cleanup before exit
func (fCsv *FileCSVee) OnEvicted(_ string, _ interface{}) {
	// verify if we need to add the trailer

	fCsv.csvWriter.Flush()
	fCsv.file.Close()
	return
}

// ExportEvent implements EventExporter
func (fCsv *FileCSVee) ExportEvent(cgrEv *utils.CGREvent) (err error) {
	// convert cgrEvent in export record
	fCsv.numberOfRecords++
	var csvRecord []string
	navMp := config.NewNavigableMap(map[string]interface{}{
		utils.MetaReq: cgrEv.Event,
	})
	for _, cfgFld := range fCsv.cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].Fields {
		if !strings.HasPrefix(cfgFld.Path, utils.MetaReq) {
			continue
		}
		if pass, err := fCsv.filterS.Pass(cgrEv.Tenant, cfgFld.Filters,
			navMp); err != nil || !pass {
			continue
		}
		val, err := cfgFld.Value.ParseDataProvider(navMp, fCsv.cgrCfg.GeneralCfg().RSRSep)
		if err != nil {
			fCsv.negativeExports.Add(cgrEv.ID)
			return
		}
		csvRecord = append(csvRecord, val)
	}
	if cost, err := cgrEv.FieldAsFloat64(utils.Cost); err == nil {
		fCsv.totalCost += cost
	}
	fCsv.positiveExports.Add(cgrEv.ID)
	fCsv.csvWriter.Write(csvRecord)
	return
}

//// Handle various meta functions used in header/trailer
//func (fCsv *FileCSVee) metaHandler(tag, arg string) (string, error) {
//	switch tag {
//	case metaExportID:
//		return cdre.exportID, nil
//	case metaTimeNow:
//		return time.Now().Format(arg), nil
//	case metaFirstCDRAtime:
//		return cdre.firstCdrATime.Format(arg), nil
//	case metaLastCDRAtime:
//		return cdre.lastCdrATime.Format(arg), nil
//	case metaNrCDRs:
//		return strconv.Itoa(cdre.numberOfRecords), nil
//	case metaDurCDRs:
//		cdr := &CDR{ToR: utils.VOICE, Usage: cdre.totalDuration}
//		return cdr.FieldAsString(&config.RSRParser{Rules: "~" + utils.Usage, AllFiltersMatch: true})
//	case metaSMSUsage:
//		cdr := &CDR{ToR: utils.SMS, Usage: cdre.totalDuration}
//		return cdr.FieldAsString(&config.RSRParser{Rules: "~" + utils.Usage, AllFiltersMatch: true})
//	case metaMMSUsage:
//		cdr := &CDR{ToR: utils.MMS, Usage: cdre.totalDuration}
//		return cdr.FieldAsString(&config.RSRParser{Rules: "~" + utils.Usage, AllFiltersMatch: true})
//	case metaGenericUsage:
//		cdr := &CDR{ToR: utils.GENERIC, Usage: cdre.totalDuration}
//		return cdr.FieldAsString(&config.RSRParser{Rules: "~" + utils.Usage, AllFiltersMatch: true})
//	case metaDataUsage:
//		cdr := &CDR{ToR: utils.DATA, Usage: cdre.totalDuration}
//		return cdr.FieldAsString(&config.RSRParser{Rules: "~" + utils.Usage, AllFiltersMatch: true})
//	case metaCostCDRs:
//		return strconv.FormatFloat(utils.Round(cdre.totalCost,
//			globalRoundingDecimals, utils.ROUNDING_MIDDLE), 'f', -1, 64), nil
//	default:
//		return "", fmt.Errorf("Unsupported METATAG: %s", tag)
//	}
//}

// Compose and cache the header
func (fCsv *FileCSVee) composeHeader() (err error) {
	var csvRecord []string
	for _, cfgFld := range fCsv.cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].Fields {
		if !strings.HasPrefix(cfgFld.Path, utils.MetaHdr) {
			continue
		}
		val, err := cfgFld.Value.ParseValue(utils.EmptyString)
		if err != nil {
			return
		}
		csvRecord = append(csvRecord, val)
	}
	fCsv.csvWriter.Write(csvRecord)
	return nil
}

// Compose and cache the trailer
func (fCsv *FileCSVee) composeTrailer() (err error) {
	for _, cfgFld := range fCsv.cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].Fields {
		if !strings.HasPrefix(cfgFld.Path, utils.MetaTrl) {
			continue
		}
		//var outVal string
		//switch cfgFld.Type {
		//case utils.META_FILLER:
		//	out, err := cfgFld.Value.ParseValue(utils.EmptyString)
		//	if err != nil {
		//		return err
		//	}
		//	outVal = out
		//	cfgFld.Padding = utils.MetaRight
		//case utils.META_CONSTANT:
		//	out, err := cfgFld.Value.ParseValue(utils.EmptyString)
		//	if err != nil {
		//		return err
		//	}
		//	outVal = out
		//case utils.META_HANDLER:
		//	out, err := cfgFld.Value.ParseValue(utils.EmptyString)
		//	if err != nil {
		//		return err
		//	}
		//	outVal, err = cdre.metaHandler(out, cfgFld.Layout)
		//default:
		//	return fmt.Errorf("Unsupported field type: %s", cfgFld.Type)
		//}
		//if err != nil {
		//	utils.Logger.Err(fmt.Sprintf("<CdreFw> Cannot export CDR trailer, field: %s, error: %s", cfgFld.Tag, err.Error()))
		//	return err
		//}
		//fmtOut := outVal
		//if fmtOut, err = utils.FmtFieldWidth(cfgFld.Tag, outVal, cfgFld.Width, cfgFld.Strip, cfgFld.Padding, cfgFld.Mandatory); err != nil {
		//	utils.Logger.Err(fmt.Sprintf("<CdreFw> Cannot export CDR trailer, field: %s, error: %s", cfgFld.Tag, err.Error()))
		//	return err
		//}
		//cdre.Lock()
		//cdre.trailer = append(cdre.trailer, fmtOut)
		//cdre.Unlock()
	}
	return nil
}
