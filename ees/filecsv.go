// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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

func NewFileCSVee(cfg *config.EventExporterCfg,
	cgrCfg *config.CGRConfig, filterS *engine.FilterS,
	em *utils.ExporterMetrics) (fCsv *FileCSVee, err error) {
	fCsv = &FileCSVee{
		cfg: cfg,
		em:  em,

		cgrCfg:  cgrCfg,
		filterS: filterS,
	}
	err = fCsv.init()
	return
}

// FileCSVee implements EventExporter interface for .csv files
type FileCSVee struct {
	cfg       *config.EventExporterCfg
	em        *utils.ExporterMetrics
	file      io.WriteCloser
	csvWriter *csv.Writer
	sync.Mutex
	slicePreparing
	// for header and trailer composing
	cgrCfg  *config.CGRConfig
	filterS *engine.FilterS
}

func (fCsv *FileCSVee) init() (err error) {
	fCsv.Lock()
	defer fCsv.Unlock()
	// create the file
	filePath := path.Join(fCsv.Cfg().ExportPath,
		fCsv.Cfg().ID+utils.Underline+utils.UUIDSha1Prefix()+utils.CSVSuffix)
	fCsv.em.Set([]string{utils.ExportPath}, filePath)
	if fCsv.file, err = os.Create(filePath); err != nil {
		return
	}
	fCsv.csvWriter = csv.NewWriter(fCsv.file)
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
		if exp, err = composeHeaderTrailer(utils.MetaHdr, fCsv.Cfg().HeaderFields(), fCsv.em, fCsv.cgrCfg, fCsv.filterS); err != nil {
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
		if exp, err = composeHeaderTrailer(utils.MetaTrl, fCsv.Cfg().TrailerFields(), fCsv.em, fCsv.cgrCfg, fCsv.filterS); err != nil {
			return
		}
		return fCsv.csvWriter.Write(exp.OrderedFieldsAsStrings())
	}
	return
}

func (fCsv *FileCSVee) Cfg() *config.EventExporterCfg { return fCsv.cfg }

func (fCsv *FileCSVee) Connect() (_ error) { return }

func (fCsv *FileCSVee) ExportEvent(ev any, _ string) error {
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
	if err = fCsv.file.Close(); err != nil {
		utils.Logger.Warning(fmt.Sprintf("<%s> Exporter with id: <%s> received error: <%s> when closing the file",
			utils.EEs, fCsv.Cfg().ID, err.Error()))
	}
	return
}

func (fCsv *FileCSVee) GetMetrics() *utils.ExporterMetrics { return fCsv.em }
