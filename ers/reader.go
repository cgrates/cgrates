// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ers

import (
	"fmt"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// EventReader .
type EventReader interface {
	Config() *config.EventReaderCfg // return it's configuration
	Serve() error                   // subscribe the reader on the path
}

// NewEventReader instantiates the event reader based on configuration at index
func NewEventReader(cfg *config.CGRConfig, cfgIdx int, rdrEvents, partialEvents chan *erEvent, rdrErr chan error,
	cache *engine.CacheS, fltrS *engine.FilterS, rdrExit chan struct{}, dm *engine.DataManager) (EventReader, error) {
	switch cfg.ERsCfg().Readers[cfgIdx].Type {
	case utils.MetaFileCSV:
		return NewCSVFileER(cfg, cfgIdx, rdrEvents, partialEvents, rdrErr, cache, fltrS, rdrExit)
	case utils.MetaFileXML:
		return NewXMLFileER(cfg, cfgIdx, rdrEvents, partialEvents, rdrErr, cache, fltrS, rdrExit)
	case utils.MetaFileFWV:
		return NewFWVFileER(cfg, cfgIdx, rdrEvents, partialEvents, rdrErr, cache, fltrS, rdrExit)
	case utils.MetaKafkajsonMap:
		return NewKafkaER(cfg, cfgIdx, rdrEvents, partialEvents, rdrErr, cache, fltrS, rdrExit)
	case utils.MetaSQL:
		return NewSQLEventReader(cfg, cfgIdx, rdrEvents, partialEvents, rdrErr, cache, fltrS, rdrExit, dm)
	case utils.MetaFileJSON:
		return NewJSONFileER(cfg, cfgIdx, rdrEvents, partialEvents, rdrErr, cache, fltrS, rdrExit)
	case utils.MetaAMQPjsonMap:
		return NewAMQPER(cfg, cfgIdx, rdrEvents, partialEvents, rdrErr, cache, fltrS, rdrExit)
	case utils.MetaS3jsonMap:
		return NewS3ER(cfg, cfgIdx, rdrEvents, partialEvents, rdrErr, cache, fltrS, rdrExit)
	case utils.MetaSQSjsonMap:
		return NewSQSER(cfg, cfgIdx, rdrEvents, partialEvents, rdrErr, cache, fltrS, rdrExit)
	case utils.MetaAMQPV1jsonMap:
		return NewAMQPv1ER(cfg, cfgIdx, rdrEvents, partialEvents, rdrErr, cache, fltrS, rdrExit)
	case utils.MetaNATSJSONMap:
		return NewNatsER(cfg, cfgIdx, rdrEvents, partialEvents, rdrErr, cache, fltrS, rdrExit)
	case utils.MetaCgrcdr:
		return NewCgrCdr(cfg, cfgIdx, rdrEvents, partialEvents, rdrErr, fltrS, rdrExit, dm)
	}
	return nil, fmt.Errorf("unsupported reader type: <%s>", cfg.ERsCfg().Readers[cfgIdx].Type)
}
