// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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

func NewFileFWVee(cfg *config.EventExporterCfg, cgrCfg *config.CGRConfig, cache *engine.CacheS, filterS *engine.FilterS, em *utils.ExporterMetrics, writer io.Writer) (fFwv *FileFWVee, err error) {
	fFwv = &FileFWVee{
		cfg: cfg,
		em:  em,

		cgrCfg:  cgrCfg,
		cache:   cache,
		filterS: filterS,
	}
	err = fFwv.init(writer)
	return fFwv, err
}

// FileFWVee implements EventExporter interface for .fwv files
type FileFWVee struct {
	cfg    *config.EventExporterCfg
	em     *utils.ExporterMetrics
	writer io.WriteCloser
	sync.Mutex
	slicePreparing

	// for header and trailer composing
	cgrCfg  *config.CGRConfig
	cache   *engine.CacheS
	filterS *engine.FilterS
}

// init will create all the necessary dependencies, including opening the file
func (fFwv *FileFWVee) init(writer io.Writer) (err error) {
	filePath := path.Join(fFwv.Cfg().ExportPath,
		fFwv.Cfg().ID+utils.Underline+utils.UUIDSha1Prefix()+utils.FWVSuffix)
	fFwv.em.Set([]string{utils.ExportPath}, filePath)
	// create the file
	if fFwv.cfg.ExportPath == utils.MetaBuffer {
		fFwv.writer = &buffer{writer}
	} else if fFwv.writer, err = os.Create(filePath); err != nil {
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
	if exp, err = composeHeaderTrailer(context.Background(), utils.MetaHdr, fFwv.Cfg().HeaderFields(), fFwv.em, fFwv.cgrCfg, fFwv.cache, fFwv.filterS); err != nil {
		return
	}
	for _, record := range exp.OrderedFieldsAsStrings() {
		if _, err = io.WriteString(fFwv.writer, record); err != nil {
			return
		}
	}
	_, err = io.WriteString(fFwv.writer, "\n")
	return
}

// Compose and cache the trailer
func (fFwv *FileFWVee) composeTrailer() (err error) {
	if len(fFwv.Cfg().TrailerFields()) == 0 {
		return
	}
	var exp *utils.OrderedNavigableMap
	if exp, err = composeHeaderTrailer(context.Background(), utils.MetaTrl, fFwv.Cfg().TrailerFields(), fFwv.em, fFwv.cgrCfg, fFwv.cache, fFwv.filterS); err != nil {
		return
	}
	for _, record := range exp.OrderedFieldsAsStrings() {
		if _, err = io.WriteString(fFwv.writer, record); err != nil {
			return
		}
	}
	_, err = io.WriteString(fFwv.writer, "\n")
	return
}

func (fFwv *FileFWVee) Cfg() *config.EventExporterCfg { return fFwv.cfg }

func (fFwv *FileFWVee) Connect() (_ error) { return }

func (fFwv *FileFWVee) ExportEvent(_ *context.Context, records, _ any) (err error) {
	fFwv.Lock() // make sure that only one event is writen in file at once
	defer fFwv.Unlock()
	for _, record := range records.([]string) {
		if _, err = io.WriteString(fFwv.writer, record); err != nil {
			return
		}
	}
	_, err = io.WriteString(fFwv.writer, "\n")
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
	if err = fFwv.writer.Close(); err != nil {
		utils.Logger.Warning(fmt.Sprintf("<%s> Exporter with id: <%s> received error: <%s> when closing the file",
			utils.EEs, fFwv.Cfg().ID, err.Error()))
	}
	return
}

func (fFwv *FileFWVee) GetMetrics() *utils.ExporterMetrics { return fFwv.em }

func (fFwv *FileFWVee) ExtraData(ev *utils.CGREvent) any { return nil }
