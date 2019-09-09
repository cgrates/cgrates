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

func (erS *ERsCfg) loadFromJsonCfg(jsnCfg *ERsJsonCfg, sep string, dfltRdrCfg *EventReaderCfg) (err error) {
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
	return erS.appendERsReaders(jsnCfg.Readers, sep, dfltRdrCfg)
}

func (ers *ERsCfg) appendERsReaders(jsnReaders *[]*EventReaderJsonCfg, sep string,
	dfltRdrCfg *EventReaderCfg) (err error) {
	if jsnReaders == nil {
		return
	}
	for _, jsnReader := range *jsnReaders {
		rdr := new(EventReaderCfg)
		if dfltRdrCfg != nil {
			rdr = dfltRdrCfg.Clone()
		}
		var haveID bool
		if jsnReader.Id != nil {
			for _, reader := range ers.Readers {
				if reader.ID == *jsnReader.Id {
					rdr = reader
					haveID = true
					break
				}
			}
		}

		if err := rdr.loadFromJsonCfg(jsnReader, sep); err != nil {
			return err
		}
		if !haveID {
			ers.Readers = append(ers.Readers, rdr)
		}

	}
	return nil
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
	Filters        []string
	Flags          utils.FlagsWithParams
	HeaderFields   []*FCTemplate
	ContentFields  []*FCTemplate
	TrailerFields  []*FCTemplate
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
		if er.HeaderFields, err = FCTemplatesFromFCTemplatesJsonCfg(*jsnCfg.Header_fields, sep); err != nil {
			return err
		}
	}
	if jsnCfg.Content_fields != nil {
		if er.ContentFields, err = FCTemplatesFromFCTemplatesJsonCfg(*jsnCfg.Content_fields, sep); err != nil {
			return err
		}
	}
	if jsnCfg.Trailer_fields != nil {
		if er.TrailerFields, err = FCTemplatesFromFCTemplatesJsonCfg(*jsnCfg.Trailer_fields, sep); err != nil {
			return err
		}
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
	if len(er.Tenant) != 0 {
		cln.Tenant = make(RSRParsers, len(er.Tenant))
		for idx, val := range er.Tenant {
			clnVal := *val
			cln.Tenant[idx] = &clnVal
		}
	}
	cln.Timezone = er.Timezone
	if len(er.Filters) != 0 {
		cln.Filters = make([]string, len(er.Filters))
		for idx, val := range er.Filters {
			cln.Filters[idx] = val
		}
	}
	cln.Flags = er.Flags
	cln.HeaderFields = make([]*FCTemplate, len(er.HeaderFields))
	for idx, fld := range er.HeaderFields {
		cln.HeaderFields[idx] = fld.Clone()
	}
	cln.ContentFields = make([]*FCTemplate, len(er.ContentFields))
	for idx, fld := range er.ContentFields {
		cln.ContentFields[idx] = fld.Clone()
	}
	cln.TrailerFields = make([]*FCTemplate, len(er.TrailerFields))
	for idx, fld := range er.TrailerFields {
		cln.TrailerFields[idx] = fld.Clone()
	}
	return
}
