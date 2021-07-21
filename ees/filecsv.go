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

	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func NewFileCSVee(cgrCfg *config.CGRConfig, cfgIdx int, filterS *engine.FilterS,
	dc *utils.SafeMapStorage) (fCsv *FileCSVee, err error) {
	fCsv = &FileCSVee{
		id:      cgrCfg.EEsCfg().Exporters[cfgIdx].ID,
		cgrCfg:  cgrCfg,
		cfgIdx:  cfgIdx,
		filterS: filterS,
		dc:      dc,
		reqs:    newConcReq(cgrCfg.EEsCfg().Exporters[cfgIdx].ConcurrentRequests),
	}
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
	dc        *utils.SafeMapStorage
	reqs      *concReq
}

// init will create all the necessary dependencies, including opening the file
func (fCsv *FileCSVee) init() (err error) {
	// create the file
	filePath := path.Join(fCsv.cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].ExportPath,
		fCsv.id+utils.Underline+utils.UUIDSha1Prefix()+utils.CSVSuffix)
	fCsv.dc.Lock()
	fCsv.dc.MapStorage[utils.ExportPath] = filePath
	fCsv.dc.Unlock()
	if fCsv.file, err = os.Create(filePath); err != nil {
		return
	}
	fCsv.csvWriter = csv.NewWriter(fCsv.file)
	fCsv.csvWriter.Comma = utils.CSVSep
	if fieldSep, has := fCsv.cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].Opts[utils.CSVFieldSepOpt]; has {
		fCsv.csvWriter.Comma = rune(utils.IfaceAsString(fieldSep)[0])
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
}

// ExportEvent implements EventExporter
func (fCsv *FileCSVee) ExportEvent(cgrEv *utils.CGREvent) (err error) {
	fCsv.reqs.get()
	defer func() {
		updateEEMetrics(fCsv.dc, cgrEv.ID, cgrEv.Event, err != nil, utils.FirstNonEmpty(fCsv.cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].Timezone,
			fCsv.cgrCfg.GeneralCfg().DefaultTimezone))
		fCsv.reqs.done()
	}()
	fCsv.dc.Lock()
	fCsv.dc.MapStorage[utils.NumberOfEvents] = fCsv.dc.MapStorage[utils.NumberOfEvents].(int64) + 1
	fCsv.dc.Unlock()

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
		eeReq := engine.NewExportRequest(map[string]utils.DataStorage{
			utils.MetaReq:  utils.MapStorage(cgrEv.Event),
			utils.MetaDC:   fCsv.dc,
			utils.MetaOpts: utils.MapStorage(cgrEv.APIOpts),
			utils.MetaCfg:  fCsv.cgrCfg.GetDataProvider(),
		}, utils.FirstNonEmpty(cgrEv.Tenant, fCsv.cgrCfg.GeneralCfg().DefaultTenant),
			fCsv.filterS, oNm)

		if err = eeReq.SetFields(fCsv.cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].ContentFields()); err != nil {
			return
		}
		csvRecord = eeReq.ExpData[utils.MetaExp].OrderedFieldsAsStrings()
	}

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
	eeReq := engine.NewExportRequest(map[string]utils.DataStorage{
		utils.MetaDC:  fCsv.dc,
		utils.MetaCfg: fCsv.cgrCfg.GetDataProvider(),
	}, fCsv.cgrCfg.GeneralCfg().DefaultTenant,
		fCsv.filterS, oNm)
	if err = eeReq.SetFields(fCsv.cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].HeaderFields()); err != nil {
		return
	}
	return fCsv.csvWriter.Write(eeReq.ExpData[utils.MetaHdr].OrderedFieldsAsStrings())
}

// Compose and cache the trailer
func (fCsv *FileCSVee) composeTrailer() (err error) {
	if len(fCsv.cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].TrailerFields()) == 0 {
		return
	}
	oNm := map[string]*utils.OrderedNavigableMap{
		utils.MetaTrl: utils.NewOrderedNavigableMap(),
	}
	eeReq := engine.NewExportRequest(map[string]utils.DataStorage{
		utils.MetaDC:  fCsv.dc,
		utils.MetaCfg: fCsv.cgrCfg.GetDataProvider(),
	}, fCsv.cgrCfg.GeneralCfg().DefaultTenant,
		fCsv.filterS, oNm)
	if err = eeReq.SetFields(fCsv.cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].TrailerFields()); err != nil {
		return
	}

	return fCsv.csvWriter.Write(eeReq.ExpData[utils.MetaTrl].OrderedFieldsAsStrings())
}

func (fCsv *FileCSVee) GetMetrics() *utils.SafeMapStorage {
	return fCsv.dc.Clone()
}
