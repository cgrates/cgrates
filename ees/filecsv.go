/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package ees

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path"
	"sync"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func NewFileCSVee(cfg *config.EventExporterCfg,
	cgrCfg *config.CGRConfig, filterS *engine.FilterS,
	dc *utils.ExporterMetrics, wrtr io.WriteCloser) (fCsv *FileCSVee, err error) {
	fCsv = &FileCSVee{
		cfg:     cfg,
		dc:      dc,
		wrtr:    wrtr,
		cgrCfg:  cgrCfg,
		filterS: filterS,
	}
	err = fCsv.init(wrtr)
	return
}

// FileCSVee implements EventExporter interface for .csv files
type FileCSVee struct {
	cfg       *config.EventExporterCfg
	dc        *utils.ExporterMetrics
	wrtr      io.WriteCloser // writer for the csv
	csvWriter *csv.Writer
	sync.Mutex
	slicePreparing
	// for header and trailer composing
	cgrCfg  *config.CGRConfig
	filterS *engine.FilterS
}

func (fCsv *FileCSVee) init(wrtr io.WriteCloser) (err error) {
	fCsv.Lock()
	defer fCsv.Unlock()
	// create the file
	filePath := path.Join(fCsv.Cfg().ExportPath,
		fCsv.Cfg().ID+utils.Underline+utils.UUIDSha1Prefix()+utils.CSVSuffix)
	fCsv.dc.Set([]string{utils.ExportPath}, filePath)
	if fCsv.cfg.ExportPath == utils.MetaBuffer {
		fCsv.wrtr = wrtr
	} else if fCsv.wrtr, err = os.Create(filePath); err != nil {
		return
	}
	fCsv.csvWriter = csv.NewWriter(fCsv.wrtr)
	fCsv.csvWriter.Comma = utils.CSVSep
	if fCsv.Cfg().Opts.CSVFieldSeparator != nil {
		fCsv.csvWriter.Comma = rune((*fCsv.Cfg().Opts.CSVFieldSeparator)[0])
	}
	return fCsv.composeHeader()
}

// Compose and cache the header
func (fCsv *FileCSVee) composeHeader() (err error) {
	if len(fCsv.Cfg().HeaderFields()) != 0 {
		var exp *utils.OrderedNavigableMap
		if exp, err = composeHeaderTrailer(context.Background(), utils.MetaHdr, fCsv.Cfg().HeaderFields(), fCsv.dc, fCsv.cgrCfg, fCsv.filterS); err != nil {
			return
		}
		return fCsv.csvWriter.Write(exp.OrderedFieldsAsStrings())
	}
	return
}

// Compose and cache the trailer
func (fCsv *FileCSVee) composeTrailer() (err error) {
	if len(fCsv.Cfg().TrailerFields()) != 0 {
		var exp *utils.OrderedNavigableMap
		if exp, err = composeHeaderTrailer(context.Background(), utils.MetaTrl, fCsv.Cfg().TrailerFields(), fCsv.dc, fCsv.cgrCfg, fCsv.filterS); err != nil {
			return
		}
		return fCsv.csvWriter.Write(exp.OrderedFieldsAsStrings())
	}
	return
}

func (fCsv *FileCSVee) Cfg() *config.EventExporterCfg { return fCsv.cfg }

func (fCsv *FileCSVee) Connect() (_ error) { return }

func (fCsv *FileCSVee) ExportEvent(_ *context.Context, ev, _ any) error {
	fCsv.Lock() // make sure that only one event is writen in file at once
	defer fCsv.Unlock()
	return fCsv.csvWriter.Write(ev.([]string))
}

func (fCsv *FileCSVee) Close() (err error) {
	fCsv.Lock()
	defer fCsv.Unlock()
	// verify if we need to add the trailer
	if err = fCsv.composeTrailer(); err != nil {
		utils.Logger.Warning(fmt.Sprintf("<%s> Exporter with id: <%s> received error: <%s> when composed trailer",
			utils.EEs, fCsv.Cfg().ID, err.Error()))
	}
	fCsv.csvWriter.Flush()
	if err = fCsv.wrtr.Close(); err != nil {
		utils.Logger.Warning(fmt.Sprintf("<%s> Exporter with id: <%s> received error: <%s> when closing the file",
			utils.EEs, fCsv.Cfg().ID, err.Error()))
	}
	return
}

func (fCsv *FileCSVee) GetMetrics() *utils.ExporterMetrics { return fCsv.dc }

func (fCsv *FileCSVee) ExtraData(ev *utils.CGREvent) any { return nil }

// Buffers cannot be closed, they just Reset. We implement our struct and used it for writer field in FileCSVee to be available for WriterCloser interface
type buffer struct {
	io.Writer
}

func (buf *buffer) Close() (err error) {
	return
}
