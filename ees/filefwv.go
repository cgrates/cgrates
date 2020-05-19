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
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"sync"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewFileFWVee(cgrCfg *config.CGRConfig, cfgIdx int, filterS *engine.FilterS) (fFwv *FileFWVee, err error) {
	fFwv = &FileFWVee{id: cgrCfg.EEsCfg().Exporters[cfgIdx].ID,
		cgrCfg: cgrCfg, cfgIdx: cfgIdx, filterS: filterS}
	err = fFwv.init()
	return
}

// FileFWVee implements EventExporter interface for .fwv files
type FileFWVee struct {
	id      string
	cgrCfg  *config.CGRConfig
	cfgIdx  int // index of config instance within ERsCfg.Readers
	filterS *engine.FilterS
	file    *os.File
	sync.RWMutex

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
func (fFwv *FileFWVee) init() (err error) {
	// create the file
	if fFwv.file, err = os.Create(path.Join(fFwv.cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].ExportPath,
		fFwv.id+utils.Underline+utils.UUIDSha1Prefix()+utils.FWVSuffix)); err != nil {
		return
	}
	fFwv.positiveExports = utils.StringSet{}
	fFwv.negativeExports = utils.StringSet{}
	return fFwv.composeHeader()
}

// ID returns the identificator of this exporter
func (fFwv *FileFWVee) ID() string {
	return fFwv.id
}

// OnEvicted implements EventExporter, doing the cleanup before exit
func (fFwv *FileFWVee) OnEvicted(_ string, _ interface{}) {
	// verify if we need to add the trailer
	if err := fFwv.composeTrailer(); err != nil {
		utils.Logger.Warning(fmt.Sprintf("<%s> Exporter with id: <%s> received error: <%s> when composed trailer",
			utils.EventExporterS, fFwv.id, err.Error()))
	}
	if err := fFwv.file.Close(); err != nil {
		utils.Logger.Warning(fmt.Sprintf("<%s> Exporter with id: <%s> received error: <%s> when closing the file",
			utils.EventExporterS, fFwv.id, err.Error()))
	}
	return
}

// ExportEvent implements EventExporter
func (fFwv *FileFWVee) ExportEvent(cgrEv *utils.CGREvent) (err error) {
	fFwv.Lock()
	defer fFwv.Unlock()
	fFwv.numberOfEvents++
	var records []string
	navMp := utils.MapStorage{utils.MetaReq: cgrEv.Event}
	for _, cfgFld := range fFwv.cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].ContentFields() {
		if pass, err := fFwv.filterS.Pass(cgrEv.Tenant, cfgFld.Filters,
			navMp); err != nil || !pass {
			continue
		}
		val, err := cfgFld.Value.ParseDataProvider(navMp, utils.NestingSep)
		if err != nil {
			if err == utils.ErrNotFound {
				err = utils.ErrPrefix(err, cfgFld.Value.GetRule())
			}
			fFwv.negativeExports.Add(cgrEv.ID)
			return err
		}
		records = append(records, val)
	}
	if aTime, err := cgrEv.FieldAsTime(utils.AnswerTime, fFwv.cgrCfg.GeneralCfg().DefaultTimezone); err == nil {
		if fFwv.firstEventATime.IsZero() || fFwv.firstEventATime.Before(aTime) {
			fFwv.firstEventATime = aTime
		}
		if aTime.After(fFwv.lastEventATime) {
			fFwv.lastEventATime = aTime
		}
	}
	if oID, err := cgrEv.FieldAsInt64(utils.OrderID); err == nil {
		if fFwv.firstExpOrderID > oID || fFwv.firstExpOrderID == 0 {
			fFwv.firstExpOrderID = oID
		}
		if fFwv.lastExpOrderID < oID {
			fFwv.lastExpOrderID = oID
		}
	}
	if cost, err := cgrEv.FieldAsFloat64(utils.Cost); err == nil {
		fFwv.totalCost += cost
	}
	if tor, err := cgrEv.FieldAsString(utils.ToR); err == nil {
		if usage, err := cgrEv.FieldAsDuration(utils.Usage); err == nil {
			switch tor {
			case utils.VOICE:
				fFwv.totalDuration += usage
			case utils.SMS:
				fFwv.totalSmsUsage += usage
			case utils.MMS:
				fFwv.totalMmsUsage += usage
			case utils.GENERIC:
				fFwv.totalGenericUsage += usage
			case utils.DATA:
				fFwv.totalDataUsage += usage
			}
		}
	}
	fFwv.positiveExports.Add(cgrEv.ID)
	for _, record := range append(records, "\n") {
		if _, err = io.WriteString(fFwv.file, record); err != nil {
			return
		}
	}
	return
}

// Compose and cache the header
func (fFwv *FileFWVee) composeHeader() (err error) {
	if len(fFwv.cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].HeaderFields()) == 0 {
		return
	}
	var records []string
	for _, cfgFld := range fFwv.cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].HeaderFields() {
		var outVal string
		switch cfgFld.Type {
		case utils.META_CONSTANT:
			outVal, err = cfgFld.Value.ParseValue(utils.EmptyString)
			if err != nil {
				if err == utils.ErrNotFound {
					err = utils.ErrPrefix(err, cfgFld.Value.GetRule())
				}
				return err
			}
		case utils.MetaExportID:
			outVal = fFwv.id
		case utils.MetaTimeNow:
			outVal = time.Now().String()
		default:
			return fmt.Errorf("unsupported type in header for field: <%+v>", utils.ToJSON(cfgFld))
		}
		fmtOut := outVal
		if fmtOut, err = utils.FmtFieldWidth(cfgFld.Tag, outVal, cfgFld.Width,
			cfgFld.Strip, cfgFld.Padding, cfgFld.Mandatory); err != nil {
			return err
		}
		records = append(records, fmtOut)
	}
	for _, record := range append(records, "\n") {
		if _, err = io.WriteString(fFwv.file, record); err != nil {
			return
		}
	}
	return
}

// Compose and cache the trailer
func (fFwv *FileFWVee) composeTrailer() (err error) {
	if len(fFwv.cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].TrailerFields()) == 0 {
		return
	}
	var records []string
	for _, cfgFld := range fFwv.cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].TrailerFields() {
		var val string
		switch cfgFld.Type {
		case utils.META_CONSTANT:
			val, err = cfgFld.Value.ParseValue(utils.EmptyString)
			if err != nil {
				if err == utils.ErrNotFound {
					err = utils.ErrPrefix(err, cfgFld.Value.GetRule())
				}
				return err
			}
		case utils.MetaExportID:
			val = fFwv.id
		case utils.MetaTimeNow:
			val = time.Now().String()
		case utils.MetaFirstEventATime:
			val = fFwv.firstEventATime.Format(cfgFld.Layout)
		case utils.MetaLastEventATime:
			val = fFwv.lastEventATime.Format(cfgFld.Layout)
		case utils.MetaEventNumber:
			val = strconv.Itoa(fFwv.numberOfEvents)
		case utils.MetaEventCost:
			rounding := fFwv.cgrCfg.GeneralCfg().RoundingDecimals
			if cfgFld.RoundingDecimals != nil {
				rounding = *cfgFld.RoundingDecimals
			}
			val = strconv.FormatFloat(utils.Round(fFwv.totalCost,
				rounding, utils.ROUNDING_MIDDLE), 'f', -1, 64)
		case utils.MetaVoiceUsage:
			val = fFwv.totalDuration.String()
		case utils.MetaDataUsage:
			val = strconv.Itoa(int(fFwv.totalDataUsage.Nanoseconds()))
		case utils.MetaSMSUsage:
			val = strconv.Itoa(int(fFwv.totalSmsUsage.Nanoseconds()))
		case utils.MetaMMSUsage:
			val = strconv.Itoa(int(fFwv.totalMmsUsage.Nanoseconds()))
		case utils.MetaGenericUsage:
			val = strconv.Itoa(int(fFwv.totalGenericUsage.Nanoseconds()))
		case utils.MetaNegativeExports:
			val = strconv.Itoa(len(fFwv.negativeExports.AsSlice()))
		case utils.MetaPositiveExports:
			val = strconv.Itoa(len(fFwv.positiveExports.AsSlice()))
		default:
			return fmt.Errorf("unsupported type in trailer for field: <%+v>", utils.ToJSON(cfgFld))
		}
		fmtOut := val
		if fmtOut, err = utils.FmtFieldWidth(cfgFld.Tag, val, cfgFld.Width,
			cfgFld.Strip, cfgFld.Padding, cfgFld.Mandatory); err != nil {
			return err
		}
		records = append(records, fmtOut)
	}
	for _, record := range append(records, "\n") {
		if _, err = io.WriteString(fFwv.file, record); err != nil {
			return
		}
	}
	return
}
