// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ers

import (
	"fmt"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type EventReader interface {
	Config() *config.EventReaderCfg // return it's configuration
	Serve() error                   // subscribe the reader on the path
}

// NewEventReader instantiates the event reader based on configuration at index
func NewEventReader(cfg *config.CGRConfig, cfgIdx int,
	rdrEvents chan *erEvent, rdrErr chan error,
	fltrS *engine.FilterS, rdrExit chan struct{}) (er EventReader, err error) {
	switch cfg.ERsCfg().Readers[cfgIdx].Type {
	default:
		err = fmt.Errorf("unsupported reader type: <%s>", cfg.ERsCfg().Readers[cfgIdx].Type)
	case utils.MetaFileCSV:
		return NewCSVFileER(cfg, cfgIdx, rdrEvents, rdrErr, fltrS, rdrExit)
	case utils.MetaPartialCSV:
		return NewPartialCSVFileER(cfg, cfgIdx, rdrEvents, rdrErr, fltrS, rdrExit)
	case utils.MetaFileXML:
		return NewXMLFileER(cfg, cfgIdx, rdrEvents, rdrErr, fltrS, rdrExit)
	case utils.MetaFileFWV:
		return NewFWVFileERER(cfg, cfgIdx, rdrEvents, rdrErr, fltrS, rdrExit)
	case utils.MetaKafkajsonMap:
		return NewKafkaER(cfg, cfgIdx, rdrEvents, rdrErr, fltrS, rdrExit)
	case utils.MetaSQL:
		return NewSQLEventReader(cfg, cfgIdx, rdrEvents, rdrErr, fltrS, rdrExit)
	case utils.MetaFlatstore:
		return NewFlatstoreER(cfg, cfgIdx, rdrEvents, rdrErr, fltrS, rdrExit)
	}
	return
}
