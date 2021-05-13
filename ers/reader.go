/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
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
func NewEventReader(cfg *config.CGRConfig, cfgIdx int,
	rdrEvents, partialEvents chan *erEvent, rdrErr chan error,
	fltrS *engine.FilterS, rdrExit chan struct{}) (er EventReader, err error) {
	switch cfg.ERsCfg().Readers[cfgIdx].Type {
	default:
		err = fmt.Errorf("unsupported reader type: <%s>", cfg.ERsCfg().Readers[cfgIdx].Type)
	case utils.MetaFileCSV:
		return NewCSVFileER(cfg, cfgIdx, rdrEvents, partialEvents, rdrErr, fltrS, rdrExit)
	case utils.MetaFileXML:
		return NewXMLFileER(cfg, cfgIdx, rdrEvents, partialEvents, rdrErr, fltrS, rdrExit)
	case utils.MetaFileFWV:
		return NewFWVFileER(cfg, cfgIdx, rdrEvents, partialEvents, rdrErr, fltrS, rdrExit)
	case utils.MetaKafkajsonMap:
		return NewKafkaER(cfg, cfgIdx, rdrEvents, partialEvents, rdrErr, fltrS, rdrExit)
	case utils.MetaSQL:
		return NewSQLEventReader(cfg, cfgIdx, rdrEvents, partialEvents, rdrErr, fltrS, rdrExit)
	case utils.MetaFileJSON:
		return NewJSONFileER(cfg, cfgIdx, rdrEvents, partialEvents, rdrErr, fltrS, rdrExit)
	case utils.MetaAMQPjsonMap:
		return NewAMQPER(cfg, cfgIdx, rdrEvents, partialEvents, rdrErr, fltrS, rdrExit)
	case utils.MetaS3jsonMap:
		return NewS3ER(cfg, cfgIdx, rdrEvents, partialEvents, rdrErr, fltrS, rdrExit)
	case utils.MetaSQSjsonMap:
		return NewSQSER(cfg, cfgIdx, rdrEvents, partialEvents, rdrErr, fltrS, rdrExit)
	case utils.MetaAMQPV1jsonMap:
		return NewAMQPv1ER(cfg, cfgIdx, rdrEvents, partialEvents, rdrErr, fltrS, rdrExit)
	}
	return
}
