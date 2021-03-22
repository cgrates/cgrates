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
	"io"
	"os"
	"path"
	"sync"

	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func NewFileCSVee(cgrCfg *config.CGRConfig, cfgIdx int, filterS *engine.FilterS,
	dc utils.MapStorage) (fCsv *FileCSVee, err error) {
	fCsv = &FileCSVee{id: cgrCfg.EEsCfg().Exporters[cfgIdx].ID,
		cgrCfg: cgrCfg, cfgIdx: cfgIdx, filterS: filterS, dc: dc}
	err = fCsv.init()
	return
}

// FileCSVee implements EventExporter interface for .csv files
type FileCSVee struct {
	id        string
	cgrCfg    *config.CGRConfig
	cfgIdx    int // index of config instance within ERsCfg.Readers
	filterS   *engine.FilterS
	file      io.WriteCloser
	csvWriter *csv.Writer
	sync.RWMutex
	dc utils.MapStorage
}

// init will create all the necessary dependencies, including opening the file
func (fCsv *FileCSVee) init() (err error) {
	// create the file
	filePath := path.Join(fCsv.cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].ExportPath,
		fCsv.id+utils.Underline+utils.UUIDSha1Prefix()+utils.CSVSuffix)
	fCsv.Lock()
	fCsv.dc[utils.ExportPath] = filePath
	fCsv.Unlock()
	if fCsv.file, err = os.Create(filePath); err != nil {
		return
	}
	fCsv.csvWriter = csv.NewWriter(fCsv.file)
	fCsv.csvWriter.Comma = utils.CSVSep
	if len(fCsv.cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].FieldSep) > 0 {
		fCsv.csvWriter.Comma = rune(fCsv.cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].FieldSep[0])
	}
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
	fCsv.Lock()
	defer func() {
		if err != nil {
			fCsv.dc[utils.NegativeExports].(utils.StringSet).Add(cgrEv.ID)
		} else {
			fCsv.dc[utils.PositiveExports].(utils.StringSet).Add(cgrEv.ID)
		}
		fCsv.Unlock()
	}()
	fCsv.dc[utils.NumberOfEvents] = fCsv.dc[utils.NumberOfEvents].(int64) + 1

	var csvRecord []string
	if len(fCsv.cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].ContentFields()) == 0 {
		csvRecord = make([]string, 0, len(cgrEv.Event))
		for _, val := range cgrEv.Event {
			csvRecord = append(csvRecord, utils.IfaceAsString(val))
		}
	} else {
		oNm := map[string]*utils.OrderedNavigableMap{
			utils.MetaExp: utils.NewOrderedNavigableMap(),
		}
		req := utils.MapStorage(cgrEv.Event)
		eeReq := engine.NewEventRequest(req, fCsv.dc, cgrEv.APIOpts,
			fCsv.cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].Tenant,
			fCsv.cgrCfg.GeneralCfg().DefaultTenant,
			utils.FirstNonEmpty(fCsv.cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].Timezone,
				fCsv.cgrCfg.GeneralCfg().DefaultTimezone),
			fCsv.filterS, oNm)

		if err = eeReq.SetFields(fCsv.cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].ContentFields()); err != nil {
			return
		}
		csvRecord = eeReq.OrdNavMP[utils.MetaExp].OrderedFieldsAsStrings()
	}

	updateEEMetrics(fCsv.dc, cgrEv.Event, utils.FirstNonEmpty(fCsv.cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].Timezone,
		fCsv.cgrCfg.GeneralCfg().DefaultTimezone))
	return fCsv.csvWriter.Write(csvRecord)
}

// Compose and cache the header
func (fCsv *FileCSVee) composeHeader() (err error) {
	if len(fCsv.cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].HeaderFields()) == 0 {
		return
	}
	oNm := map[string]*utils.OrderedNavigableMap{
		utils.MetaHdr: utils.NewOrderedNavigableMap(),
	}
	eeReq := engine.NewEventRequest(nil, fCsv.dc, nil,
		fCsv.cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].Tenant,
		fCsv.cgrCfg.GeneralCfg().DefaultTenant,
		utils.FirstNonEmpty(fCsv.cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].Timezone,
			fCsv.cgrCfg.GeneralCfg().DefaultTimezone),
		fCsv.filterS, oNm)
	if err = eeReq.SetFields(fCsv.cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].HeaderFields()); err != nil {
		return
	}
	return fCsv.csvWriter.Write(eeReq.OrdNavMP[utils.MetaHdr].OrderedFieldsAsStrings())
}

// Compose and cache the trailer
func (fCsv *FileCSVee) composeTrailer() (err error) {
	if len(fCsv.cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].TrailerFields()) == 0 {
		return
	}
	oNm := map[string]*utils.OrderedNavigableMap{
		utils.MetaTrl: utils.NewOrderedNavigableMap(),
	}
	eeReq := engine.NewEventRequest(nil, fCsv.dc, nil,
		fCsv.cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].Tenant,
		fCsv.cgrCfg.GeneralCfg().DefaultTenant,
		utils.FirstNonEmpty(fCsv.cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].Timezone,
			fCsv.cgrCfg.GeneralCfg().DefaultTimezone),
		fCsv.filterS, oNm)
	if err = eeReq.SetFields(fCsv.cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].TrailerFields()); err != nil {
		return
	}

	return fCsv.csvWriter.Write(eeReq.OrdNavMP[utils.MetaTrl].OrderedFieldsAsStrings())
}

func (fCsv *FileCSVee) GetMetrics() utils.MapStorage {
	return fCsv.dc.Clone()
}
