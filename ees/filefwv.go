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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewFileFWVee(cfg *config.EventExporterCfg, cgrCfg *config.CGRConfig, filterS *engine.FilterS, dc *utils.SafeMapStorage) (fFwv *FileFWVee, err error) {
	fFwv = &FileFWVee{
		cfg: cfg,
		dc:  dc,

		cgrCfg:  cgrCfg,
		filterS: filterS,
	}
	err = fFwv.init()
	return
}

// FileFWVee implements EventExporter interface for .fwv files
type FileFWVee struct {
	cfg  *config.EventExporterCfg
	dc   *utils.SafeMapStorage
	file io.WriteCloser
	sync.Mutex
	slicePreparing

	// for header and trailer composing
	cgrCfg  *config.CGRConfig
	filterS *engine.FilterS
}

// init will create all the necessary dependencies, including opening the file
func (fFwv *FileFWVee) init() (err error) {
	filePath := path.Join(fFwv.Cfg().ExportPath,
		fFwv.Cfg().ID+utils.Underline+utils.UUIDSha1Prefix()+utils.FWVSuffix)
	fFwv.dc.Lock()
	fFwv.dc.MapStorage[utils.ExportPath] = filePath
	fFwv.dc.Unlock()
	// create the file
	if fFwv.file, err = os.Create(filePath); err != nil {
		return
	}
	return fFwv.composeHeader()
}

// Compose and cache the header
func (fFwv *FileFWVee) composeHeader() (err error) {
	if len(fFwv.Cfg().HeaderFields()) == 0 {
		return
	}
	var exp *utils.OrderedNavigableMap
	if exp, err = composeHeaderTrailer(context.Background(), utils.MetaHdr, fFwv.Cfg().HeaderFields(), fFwv.dc, fFwv.cgrCfg, fFwv.filterS); err != nil {
		return
	}
	for _, record := range exp.OrderedFieldsAsStrings() {
		if _, err = io.WriteString(fFwv.file, record); err != nil {
			return
		}
	}
	_, err = io.WriteString(fFwv.file, "\n")
	return
}

// Compose and cache the trailer
func (fFwv *FileFWVee) composeTrailer() (err error) {
	if len(fFwv.Cfg().TrailerFields()) == 0 {
		return
	}
	var exp *utils.OrderedNavigableMap
	if exp, err = composeHeaderTrailer(context.Background(), utils.MetaTrl, fFwv.Cfg().TrailerFields(), fFwv.dc, fFwv.cgrCfg, fFwv.filterS); err != nil {
		return
	}
	for _, record := range exp.OrderedFieldsAsStrings() {
		if _, err = io.WriteString(fFwv.file, record); err != nil {
			return
		}
	}
	_, err = io.WriteString(fFwv.file, "\n")
	return
}

func (fFwv *FileFWVee) Cfg() *config.EventExporterCfg { return fFwv.cfg }

func (fFwv *FileFWVee) Connect() (_ error) { return }

func (fFwv *FileFWVee) ExportEvent(_ *context.Context, records interface{}, _ string) (err error) {
	fFwv.Lock() // make sure that only one event is writen in file at once
	defer fFwv.Unlock()
	for _, record := range records.([]string) {
		if _, err = io.WriteString(fFwv.file, record); err != nil {
			return
		}
	}
	_, err = io.WriteString(fFwv.file, "\n")
	return
}

func (fFwv *FileFWVee) Close() (err error) {
	fFwv.Lock()
	defer fFwv.Unlock()
	// verify if we need to add the trailer
	if err = fFwv.composeTrailer(); err != nil {
		utils.Logger.Warning(fmt.Sprintf("<%s> Exporter with id: <%s> received error: <%s> when composed trailer",
			utils.EEs, fFwv.Cfg().ID, err.Error()))
	}
	if err = fFwv.file.Close(); err != nil {
		utils.Logger.Warning(fmt.Sprintf("<%s> Exporter with id: <%s> received error: <%s> when closing the file",
			utils.EEs, fFwv.Cfg().ID, err.Error()))
	}
	return
}

func (fFwv *FileFWVee) GetMetrics() *utils.SafeMapStorage { return fFwv.dc }
