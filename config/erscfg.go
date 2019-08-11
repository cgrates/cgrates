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

package config

import (
	"time"

	"github.com/cgrates/cgrates/utils"
)

type ERsCfg struct {
	Enabled       bool
	SessionSConns []*RemoteHost
	Readers       []*EventReaderCfg
}

func (erS *ERsJsonCfg) loadFromJsonCfg(jsnCfg *ERsJsonCfg, sep string) (err error) {
	if jsnCfg == nil {
		return
	}
	return
}

// Clone itself into a new ERsCfg
func (erS *ERsCfg) Clone() (cln *ERsCfg) {
	return
}

type EventReaderCfg struct {
	ID             string
	Type           string
	FieldSep       string
	RunDelay       time.Duration
	ConcurrentReqs int
	SourcePath     string
	ProcessedPath  string
	XmlRootPath    string
	Tenant         RSRParsers
	Timezone       string
	SourceID       string
	Filters        []string
	Flags          utils.FlagsWithParams
	Header_fields  []*FCTemplate
	Content_fields []*FCTemplate
	Trailer_fields []*FCTemplate
	Continue       bool
}

func (er *EventReaderCfg) loadFromJsonCfg(jsnCfg *EventReaderJsonCfg, sep string) (err error) {
	if jsnCfg == nil {
		return
	}
	return
}
