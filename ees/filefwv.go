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
	"sync"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewFileFWVee(cgrCfg *config.CGRConfig, cfgIdx int, filterS *engine.FilterS, dc utils.MapStorage) (fFwv *FileFWVee, err error) {
	fFwv = &FileFWVee{id: cgrCfg.EEsCfg().Exporters[cfgIdx].ID,
		cgrCfg: cgrCfg, cfgIdx: cfgIdx, filterS: filterS, dc: dc}
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
	dc      utils.MapStorage
	sync.RWMutex
}

// init will create all the necessary dependencies, including opening the file
func (fFwv *FileFWVee) init() (err error) {
	filePath := path.Join(fFwv.cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].ExportPath,
		fFwv.id+utils.Underline+utils.UUIDSha1Prefix()+utils.FWVSuffix)
	fFwv.Lock()
	fFwv.dc[utils.ExportPath] = filePath
	fFwv.Unlock()
	// create the file
	if fFwv.file, err = os.Create(filePath); err != nil {
		return
	}
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
	defer func() {
		if err != nil {
			fFwv.dc[utils.NegativeExports].(utils.StringSet).Add(cgrEv.ID)
		} else {
			fFwv.dc[utils.PositiveExports].(utils.StringSet).Add(cgrEv.ID)
		}
		fFwv.Unlock()
	}()
	fFwv.dc[utils.NumberOfEvents] = fFwv.dc[utils.NumberOfEvents].(int64) + 1
	var records []string
	if len(fFwv.cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].ContentFields()) == 0 {
		for _, val := range cgrEv.Event {
			records = append(records, utils.IfaceAsString(val))
		}
	} else {
		req := utils.MapStorage(cgrEv.Event)
		eeReq := NewEventExporterRequest(req, fFwv.dc, cgrEv.Opts,
			fFwv.cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].Tenant,
			fFwv.cgrCfg.GeneralCfg().DefaultTenant,
			utils.FirstNonEmpty(fFwv.cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].Timezone,
				fFwv.cgrCfg.GeneralCfg().DefaultTimezone),
			fFwv.filterS)

		if err = eeReq.SetFields(fFwv.cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].ContentFields()); err != nil {
			return
		}
		for el := eeReq.cnt.GetFirstElement(); el != nil; el = el.Next() {
			var strVal string
			if strVal, err = eeReq.cnt.FieldAsString(el.Value.Slice()); err != nil {
				return
			}
			records = append(records, strVal)
		}
	}

	updateEEMetrics(fFwv.dc, cgrEv.Event, utils.FirstNonEmpty(fFwv.cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].Timezone,
		fFwv.cgrCfg.GeneralCfg().DefaultTimezone))
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
	eeReq := NewEventExporterRequest(nil, fFwv.dc, nil,
		fFwv.cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].Tenant,
		fFwv.cgrCfg.GeneralCfg().DefaultTenant,
		utils.FirstNonEmpty(fFwv.cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].Timezone,
			fFwv.cgrCfg.GeneralCfg().DefaultTimezone),
		fFwv.filterS)
	if err = eeReq.SetFields(fFwv.cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].HeaderFields()); err != nil {
		return
	}
	for el := eeReq.hdr.GetFirstElement(); el != nil; el = el.Next() {
		var strVal string
		if strVal, err = eeReq.hdr.FieldAsString(el.Value.Slice()); err != nil {
			return
		}
		records = append(records, strVal)
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
	eeReq := NewEventExporterRequest(nil, fFwv.dc, nil,
		fFwv.cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].Tenant,
		fFwv.cgrCfg.GeneralCfg().DefaultTenant,
		utils.FirstNonEmpty(fFwv.cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].Timezone,
			fFwv.cgrCfg.GeneralCfg().DefaultTimezone),
		fFwv.filterS)
	if err = eeReq.SetFields(fFwv.cgrCfg.EEsCfg().Exporters[fFwv.cfgIdx].TrailerFields()); err != nil {
		return
	}
	for el := eeReq.trl.GetFirstElement(); el != nil; el = el.Next() {
		var strVal string
		if strVal, err = eeReq.trl.FieldAsString(el.Value.Slice()); err != nil {
			return
		}
		records = append(records, strVal)
	}
	for _, record := range append(records, "\n") {
		if _, err = io.WriteString(fFwv.file, record); err != nil {
			return
		}
	}
	return
}

func (fFwv *FileFWVee) GetMetrics() utils.MapStorage {
	return fFwv.dc.Clone()
}
