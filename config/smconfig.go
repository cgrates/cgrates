/*
Real-time Charging System for Telecom & ISP environments
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
	"github.com/cgrates/cgrates/utils"
	"time"
)

// Returns the first cached default value for a SM-FreeSWITCH connection
func NewDfltFsConnConfig() *FsConnConfig {
	if dfltFsConnConfig == nil {
		return new(FsConnConfig) // No defaults, most probably we are building the defaults now
	}
	return dfltFsConnConfig
}

// One connection to FreeSWITCH server
type FsConnConfig struct {
	Server     string
	Password   string
	Reconnects int
}

func (self *FsConnConfig) loadFromJsonCfg(jsnCfg *FsConnJsonCfg) error {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Server != nil {
		self.Server = *jsnCfg.Server
	}
	if jsnCfg.Password != nil {
		self.Password = *jsnCfg.Password
	}
	if jsnCfg.Reconnects != nil {
		self.Reconnects = *jsnCfg.Reconnects
	}
	return nil
}

type SmFsConfig struct {
	Enabled             bool
	Rater               string
	Cdrs                string
	CdrExtraFields      []string
	DebitInterval       time.Duration
	MinCallDuration     time.Duration
	MaxCallDuration     time.Duration
	MinDurLowBalance    time.Duration
	LowBalanceAnnFile   string
	EmptyBalanceContext string
	EmptyBalanceAnnFile string
	Connections         []*FsConnConfig
}

func (self *SmFsConfig) loadFromJsonCfg(jsnCfg *SmFsJsonCfg) error {
	if jsnCfg == nil {
		return nil
	}
	var err error
	if jsnCfg.Enabled != nil {
		self.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Rater != nil {
		self.Rater = *jsnCfg.Rater
	}
	if jsnCfg.Cdrs != nil {
		self.Cdrs = *jsnCfg.Cdrs
	}
	if jsnCfg.Cdr_extra_fields != nil {
		self.CdrExtraFields = *jsnCfg.Cdr_extra_fields
	}
	if jsnCfg.Debit_interval != nil {
		if self.DebitInterval, err = utils.ParseDurationWithSecs(*jsnCfg.Debit_interval); err != nil {
			return err
		}
	}
	if jsnCfg.Min_call_duration != nil {
		if self.MinCallDuration, err = utils.ParseDurationWithSecs(*jsnCfg.Min_call_duration); err != nil {
			return err
		}
	}
	if jsnCfg.Max_call_duration != nil {
		if self.MaxCallDuration, err = utils.ParseDurationWithSecs(*jsnCfg.Max_call_duration); err != nil {
			return err
		}
	}
	if jsnCfg.Min_dur_low_balance != nil {
		if self.MinDurLowBalance, err = utils.ParseDurationWithSecs(*jsnCfg.Min_dur_low_balance); err != nil {
			return err
		}
	}
	if jsnCfg.Low_balance_ann_file != nil {
		self.LowBalanceAnnFile = *jsnCfg.Low_balance_ann_file
	}
	if jsnCfg.Empty_balance_context != nil {
		self.EmptyBalanceContext = *jsnCfg.Empty_balance_context
	}
	if jsnCfg.Empty_balance_ann_file != nil {
		self.EmptyBalanceAnnFile = *jsnCfg.Empty_balance_ann_file
	}
	if jsnCfg.Connections != nil {
		self.Connections = make([]*FsConnConfig, len(*jsnCfg.Connections))
		for idx, jsnConnCfg := range *jsnCfg.Connections {
			self.Connections[idx] = NewDfltFsConnConfig()
			self.Connections[idx].loadFromJsonCfg(jsnConnCfg)
		}
	}
	return nil
}

// Returns the first cached default value for a SM-FreeSWITCH connection
func NewDfltKamConnConfig() *KamConnConfig {
	if dfltKamConnConfig == nil {
		return new(KamConnConfig) // No defaults, most probably we are building the defaults now
	}
	return dfltKamConnConfig
}

// Represents one connection instance towards Kamailio
type KamConnConfig struct {
	EvapiAddr  string
	Reconnects int
}

func (self *KamConnConfig) loadFromJsonCfg(jsnCfg *KamConnJsonCfg) error {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Evapi_addr != nil {
		self.EvapiAddr = *jsnCfg.Evapi_addr
	}
	if jsnCfg.Reconnects != nil {
		self.Reconnects = *jsnCfg.Reconnects
	}
	return nil
}

// SM-Kamailio config section
type SmKamConfig struct {
	Enabled         bool
	Rater           string
	Cdrs            string
	DebitInterval   time.Duration
	MinCallDuration time.Duration
	MaxCallDuration time.Duration
	Connections     []*KamConnConfig
}

func (self *SmKamConfig) loadFromJsonCfg(jsnCfg *SmKamJsonCfg) error {
	if jsnCfg == nil {
		return nil
	}
	var err error
	if jsnCfg.Enabled != nil {
		self.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Rater != nil {
		self.Rater = *jsnCfg.Rater
	}
	if jsnCfg.Cdrs != nil {
		self.Cdrs = *jsnCfg.Cdrs
	}
	if jsnCfg.Debit_interval != nil {
		if self.DebitInterval, err = utils.ParseDurationWithSecs(*jsnCfg.Debit_interval); err != nil {
			return err
		}
	}
	if jsnCfg.Min_call_duration != nil {
		if self.MinCallDuration, err = utils.ParseDurationWithSecs(*jsnCfg.Min_call_duration); err != nil {
			return err
		}
	}
	if jsnCfg.Max_call_duration != nil {
		if self.MaxCallDuration, err = utils.ParseDurationWithSecs(*jsnCfg.Max_call_duration); err != nil {
			return err
		}
	}
	if jsnCfg.Connections != nil {
		self.Connections = make([]*KamConnConfig, len(*jsnCfg.Connections))
		for idx, jsnConnCfg := range *jsnCfg.Connections {
			self.Connections[idx] = NewDfltKamConnConfig()
			self.Connections[idx].loadFromJsonCfg(jsnConnCfg)
		}
	}
	return nil
}

// Returns the first cached default value for a SM-FreeSWITCH connection
func NewDfltOsipsConnConfig() *OsipsConnConfig {
	if dfltOsipsConnConfig == nil {
		return new(OsipsConnConfig) // No defaults, most probably we are building the defaults now
	}
	return dfltOsipsConnConfig
}

// Represents one connection instance towards OpenSIPS
type OsipsConnConfig struct {
	MiAddr     string
	Reconnects int
}

func (self *OsipsConnConfig) loadFromJsonCfg(jsnCfg *OsipsConnJsonCfg) error {
	if jsnCfg.Mi_addr != nil {
		self.MiAddr = *jsnCfg.Mi_addr
	}
	if jsnCfg.Reconnects != nil {
		self.Reconnects = *jsnCfg.Reconnects
	}
	return nil
}

// SM-OpenSIPS config section
type SmOsipsConfig struct {
	Enabled                 bool
	ListenUdp               string
	Rater                   string
	Cdrs                    string
	DebitInterval           time.Duration
	MinCallDuration         time.Duration
	MaxCallDuration         time.Duration
	EventsSubscribeInterval time.Duration
	Connections             []*OsipsConnConfig
}

func (self *SmOsipsConfig) loadFromJsonCfg(jsnCfg *SmOsipsJsonCfg) error {
	var err error
	if jsnCfg.Enabled != nil {
		self.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Listen_udp != nil {
		self.ListenUdp = *jsnCfg.Listen_udp
	}
	if jsnCfg.Rater != nil {
		self.Rater = *jsnCfg.Rater
	}
	if jsnCfg.Cdrs != nil {
		self.Cdrs = *jsnCfg.Cdrs
	}
	if jsnCfg.Debit_interval != nil {
		if self.DebitInterval, err = utils.ParseDurationWithSecs(*jsnCfg.Debit_interval); err != nil {
			return err
		}
	}
	if jsnCfg.Min_call_duration != nil {
		if self.MinCallDuration, err = utils.ParseDurationWithSecs(*jsnCfg.Min_call_duration); err != nil {
			return err
		}
	}
	if jsnCfg.Max_call_duration != nil {
		if self.MaxCallDuration, err = utils.ParseDurationWithSecs(*jsnCfg.Max_call_duration); err != nil {
			return err
		}
	}
	if jsnCfg.Events_subscribe_interval != nil {
		if self.EventsSubscribeInterval, err = utils.ParseDurationWithSecs(*jsnCfg.Events_subscribe_interval); err != nil {
			return err
		}
	}
	if jsnCfg.Connections != nil {
		self.Connections = make([]*OsipsConnConfig, len(*jsnCfg.Connections))
		for idx, jsnConnCfg := range *jsnCfg.Connections {
			self.Connections[idx] = NewDfltOsipsConnConfig()
			self.Connections[idx].loadFromJsonCfg(jsnConnCfg)
		}
	}
	return nil
}
