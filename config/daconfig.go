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

type DiameterAgentCfg struct {
	Enabled            bool   // enables the diameter agent: <true|false>
	Listen             string // address where to listen for diameter requests <x.y.z.y:1234>
	DictionariesDir    string
	SMGenericConns     []*HaPoolConfig // connections towards SMG component
	PubSubConns        []*HaPoolConfig // connection towards pubsubs
	CreateCDR          bool
	CDRRequiresSession bool
	DebitInterval      time.Duration
	Timezone           string // timezone for timestamps where not specified <""|UTC|Local|$IANA_TZ_DB>
	OriginHost         string
	OriginRealm        string
	VendorId           int
	ProductName        string
	RequestProcessors  []*DARequestProcessor
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
	if jsnCfg.Dictionaries_dir != nil {
		self.DictionariesDir = *jsnCfg.Dictionaries_dir
	}
	if jsnCfg.Sm_generic_conns != nil {
		self.SMGenericConns = make([]*HaPoolConfig, len(*jsnCfg.Sm_generic_conns))
		for idx, jsnHaCfg := range *jsnCfg.Sm_generic_conns {
			self.SMGenericConns[idx] = NewDfltHaPoolConfig()
			self.SMGenericConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Pubsubs_conns != nil {
		self.PubSubConns = make([]*HaPoolConfig, len(*jsnCfg.Pubsubs_conns))
		for idx, jsnHaCfg := range *jsnCfg.Pubsubs_conns {
			self.PubSubConns[idx] = NewDfltHaPoolConfig()
			self.PubSubConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Create_cdr != nil {
		self.CreateCDR = *jsnCfg.Create_cdr
	}
	if jsnCfg.Cdr_requires_session != nil {
		self.CDRRequiresSession = *jsnCfg.Cdr_requires_session
	}
	if jsnCfg.Debit_interval != nil {
		var err error
		if self.DebitInterval, err = utils.ParseDurationWithSecs(*jsnCfg.Debit_interval); err != nil {
			return err
		}
	}
	if jsnCfg.Timezone != nil {
		self.Timezone = *jsnCfg.Timezone
	}
	if jsnCfg.Origin_host != nil {
		self.OriginHost = *jsnCfg.Origin_host
	}
	if jsnCfg.Origin_realm != nil {
		self.OriginRealm = *jsnCfg.Origin_realm
	}
	if jsnCfg.Vendor_id != nil {
		self.VendorId = *jsnCfg.Vendor_id
	}
	if jsnCfg.Product_name != nil {
		self.ProductName = *jsnCfg.Product_name
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
			self.RequestProcessors = append(self.RequestProcessors, rp)
		}
	}
	return nil
}

// One Diameter request processor configuration
type DARequestProcessor struct {
	Id                string
	DryRun            bool
	PublishEvent      bool
	RequestFilter     utils.RSRFields
	Flags             utils.StringMap // Various flags to influence behavior
	ContinueOnSuccess bool
	AppendCCA         bool
	CCRFields         []*CfgCdrField
	CCAFields         []*CfgCdrField
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
	if jsnCfg.Publish_event != nil {
		self.PublishEvent = *jsnCfg.Publish_event
	}
	var err error
	if jsnCfg.Request_filter != nil {
		if self.RequestFilter, err = utils.ParseRSRFields(*jsnCfg.Request_filter, utils.INFIELD_SEP); err != nil {
			return err
		}
	}
	if jsnCfg.Flags != nil {
		self.Flags = utils.StringMapFromSlice(*jsnCfg.Flags)
	}
	if jsnCfg.Continue_on_success != nil {
		self.ContinueOnSuccess = *jsnCfg.Continue_on_success
	}
	if jsnCfg.Append_cca != nil {
		self.AppendCCA = *jsnCfg.Append_cca
	}
	if jsnCfg.CCR_fields != nil {
		if self.CCRFields, err = CfgCdrFieldsFromCdrFieldsJsonCfg(*jsnCfg.CCR_fields); err != nil {
			return err
		}
	}
	if jsnCfg.CCA_fields != nil {
		if self.CCAFields, err = CfgCdrFieldsFromCdrFieldsJsonCfg(*jsnCfg.CCA_fields); err != nil {
			return err
		}
	}
	return nil
}
