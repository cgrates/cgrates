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

func (erS *ERsCfg) loadFromJsonCfg(jsnCfg *ERsJsonCfg, sep string) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		erS.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Sessions_conns != nil {
		erS.SessionSConns = make([]*RemoteHost, len(*jsnCfg.Sessions_conns))
		for idx, jsnHaCfg := range *jsnCfg.Sessions_conns {
			erS.SessionSConns[idx] = NewDfltRemoteHost()
			erS.SessionSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Readers != nil {
		erS.Readers = make([]*EventReaderCfg, len(*jsnCfg.Readers))
		for idx, rdrs := range *jsnCfg.Readers {
			erS.Readers[idx] = NewDfltEventReaderCfg()
			if err = erS.Readers[idx].loadFromJsonCfg(rdrs, sep); err != nil {
				return err
			}
		}
	}
	return
}

// Clone itself into a new ERsCfg
func (erS *ERsCfg) Clone() (cln *ERsCfg) {
	cln = new(ERsCfg)
	cln.Enabled = erS.Enabled
	cln.SessionSConns = make([]*RemoteHost, len(erS.SessionSConns))
	for idx, sConn := range erS.SessionSConns {
		clonedVal := *sConn
		cln.SessionSConns[idx] = &clonedVal
	}
	cln.Readers = make([]*EventReaderCfg, len(erS.Readers))
	for idx, rdr := range erS.Readers {
		cln.Readers[idx] = rdr.Clone()
	}
	return
}

func NewDfltEventReaderCfg() *EventReaderCfg {
	return new(EventReaderCfg)
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
	SourceID       string
	Tenant         RSRParsers
	Timezone       string
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
	if jsnCfg.Id != nil {
		er.ID = *jsnCfg.Id
	}
	if jsnCfg.Type != nil {
		er.Type = *jsnCfg.Type
	}
	if jsnCfg.Field_separator != nil {
		er.FieldSep = *jsnCfg.Field_separator
	}
	if jsnCfg.Run_delay != nil {
		er.RunDelay = time.Duration(*jsnCfg.Run_delay)
	}
	if jsnCfg.Concurrent_requests != nil {
		er.ConcurrentReqs = *jsnCfg.Concurrent_requests
	}
	if jsnCfg.Source_path != nil {
		er.SourcePath = *jsnCfg.Source_path
	}
	if jsnCfg.Processed_path != nil {
		er.ProcessedPath = *jsnCfg.Processed_path
	}
	if jsnCfg.Xml_root_path != nil {
		er.XmlRootPath = *jsnCfg.Xml_root_path
	}
	if jsnCfg.Source_id != nil {
		er.SourceID = *jsnCfg.Source_id
	}
	if jsnCfg.Tenant != nil {
		if er.Tenant, err = NewRSRParsers(*jsnCfg.Tenant, true, sep); err != nil {
			return err
		}
	}
	if jsnCfg.Timezone != nil {
		er.Timezone = *jsnCfg.Timezone
	}
	if jsnCfg.Filters != nil {
		er.Filters = make([]string, len(*jsnCfg.Filters))
		for i, fltr := range *jsnCfg.Filters {
			er.Filters[i] = fltr
		}
	}
	if jsnCfg.Flags != nil {
		if er.Flags, err = utils.FlagsWithParamsFromSlice(*jsnCfg.Flags); err != nil {
			return
		}
	}
	if jsnCfg.Header_fields != nil {
		if er.Header_fields, err = FCTemplatesFromFCTemplatesJsonCfg(*jsnCfg.Header_fields, sep); err != nil {
			return err
		}
	}
	if jsnCfg.Content_fields != nil {
		if er.Content_fields, err = FCTemplatesFromFCTemplatesJsonCfg(*jsnCfg.Content_fields, sep); err != nil {
			return err
		}
	}
	if jsnCfg.Trailer_fields != nil {
		if er.Trailer_fields, err = FCTemplatesFromFCTemplatesJsonCfg(*jsnCfg.Trailer_fields, sep); err != nil {
			return err
		}
	}
	if jsnCfg.Continue != nil {
		er.Continue = *jsnCfg.Continue
	}
	return
}

//Clone itself into a new EventReaderCfg
func (er *EventReaderCfg) Clone() (cln *EventReaderCfg) {
	cln = new(EventReaderCfg)
	cln.ID = er.ID
	cln.Type = er.Type
	cln.FieldSep = er.FieldSep
	cln.RunDelay = er.RunDelay
	cln.ConcurrentReqs = er.ConcurrentReqs
	cln.SourcePath = er.SourcePath
	cln.ProcessedPath = er.ProcessedPath
	cln.XmlRootPath = er.XmlRootPath
	cln.SourceID = er.SourceID
	cln.Tenant = make(RSRParsers, len(er.Tenant))
	for idx, val := range er.Tenant {
		clnVal := *val
		cln.Tenant[idx] = &clnVal
	}
	cln.Timezone = er.Timezone
	if len(er.Filters) != 0 {
		cln.Filters = make([]string, len(er.Filters))
		for idx, val := range er.Filters {
			cln.Filters[idx] = val
		}
	}
	cln.Flags = er.Flags
	cln.Header_fields = make([]*FCTemplate, len(er.Header_fields))
	for idx, fld := range er.Header_fields {
		cln.Header_fields[idx] = fld.Clone()
	}
	cln.Content_fields = make([]*FCTemplate, len(er.Content_fields))
	for idx, fld := range er.Content_fields {
		cln.Content_fields[idx] = fld.Clone()
	}
	cln.Trailer_fields = make([]*FCTemplate, len(er.Trailer_fields))
	for idx, fld := range er.Trailer_fields {
		cln.Trailer_fields[idx] = fld.Clone()
	}
	cln.Continue = er.Continue
	return
}
