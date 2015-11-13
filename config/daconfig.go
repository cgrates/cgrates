/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2015 ITsysCOM GmbH

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
	//"time"

	"github.com/cgrates/cgrates/utils"
)

type DiameterAgentCfg struct {
	Enabled           bool   // enables the diameter agent: <true|false>
	Listen            string // address where to listen for diameter requests <x.y.z.y:1234>
	Timezone          string // timezone for timestamps where not specified <""|UTC|Local|$IANA_TZ_DB>
	RequestProcessors []*DARequestProcessor
}

func (self *DiameterAgentCfg) loadFromJsonCfg(jsnCfg *DiameterAgentJsonCfg) error {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		self.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Listen != nil {
		self.Listen = *jsnCfg.Listen
	}
	if jsnCfg.Timezone != nil {
		self.Timezone = *jsnCfg.Timezone
	}
	if jsnCfg.Request_processors != nil {
		for _, reqProcJsn := range *jsnCfg.Request_processors {
			rp := new(DARequestProcessor)
			for _, rpSet := range self.RequestProcessors {
				if reqProcJsn.Id != nil && rpSet.Id == *reqProcJsn.Id {
					rp = rpSet // Will load data into the one set
					break
				}
			}
			if err := rp.loadFromJsonCfg(reqProcJsn); err != nil {
				return nil
			}
		}
	}
	return nil
}

// One Diameter request processor configuration
type DARequestProcessor struct {
	Id                string
	DryRun            bool
	RequestFilter     utils.RSRFields
	ContinueOnSuccess bool
	ContentFields     []*CfgCdrField
}

func (self *DARequestProcessor) loadFromJsonCfg(jsnCfg *DARequestProcessorJsnCfg) error {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Id != nil {
		self.Id = *jsnCfg.Id
	}
	if jsnCfg.Dry_run != nil {
		self.DryRun = *jsnCfg.Dry_run
	}
	var err error
	if jsnCfg.Request_filter != nil {
		if self.RequestFilter, err = utils.ParseRSRFields(*jsnCfg.Request_filter, utils.INFIELD_SEP); err != nil {
			return err
		}
	}
	if jsnCfg.Content_fields != nil {
		if self.ContentFields, err = CfgCdrFieldsFromCdrFieldsJsonCfg(*jsnCfg.Content_fields); err != nil {
			return err
		}
	}
	return nil
}
