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
	"encoding/json"
	"io"
	"os"

	"github.com/DisposaBoy/JsonConfigReader"
)

const (
	GENERAL_JSN     = "general"
	LISTEN_JSN      = "listen"
	TPDB_JSN        = "tariffplan_db"
	DATADB_JSN      = "data_db"
	STORDB_JSN      = "stor_db"
	BALANCER_JSN    = "balancer"
	RALS_JSN        = "rals"
	SCHEDULER_JSN   = "scheduler"
	CDRS_JSN        = "cdrs"
	MEDIATOR_JSN    = "mediator"
	CDRSTATS_JSN    = "cdrstats"
	CDRE_JSN        = "cdre"
	CDRC_JSN        = "cdrc"
	SMGENERIC_JSON  = "sm_generic"
	SMFS_JSN        = "sm_freeswitch"
	SMKAM_JSN       = "sm_kamailio"
	SMOSIPS_JSN     = "sm_opensips"
	SM_JSN          = "session_manager"
	FS_JSN          = "freeswitch"
	KAMAILIO_JSN    = "kamailio"
	OSIPS_JSN       = "opensips"
	DA_JSN          = "diameter_agent"
	HISTSERV_JSN    = "historys"
	PUBSUBSERV_JSN  = "pubsubs"
	ALIASESSERV_JSN = "aliases"
	USERSERV_JSN    = "users"
	MAILER_JSN      = "mailer"
	SURETAX_JSON    = "suretax"
)

// Loads the json config out of io.Reader, eg other sources than file, maybe over http
func NewCgrJsonCfgFromReader(r io.Reader) (*CgrJsonCfg, error) {
	var cgrJsonCfg CgrJsonCfg
	jr := JsonConfigReader.New(r)
	if err := json.NewDecoder(jr).Decode(&cgrJsonCfg); err != nil {
		return nil, err
	}
	return &cgrJsonCfg, nil
}

// Loads the config out of file
func NewCgrJsonCfgFromFile(fpath string) (*CgrJsonCfg, error) {
	cfgFile, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}
	defer cfgFile.Close()
	return NewCgrJsonCfgFromReader(cfgFile)
}

// Main object holding the loaded config as section raw messages
type CgrJsonCfg map[string]*json.RawMessage

func (self CgrJsonCfg) GeneralJsonCfg() (*GeneralJsonCfg, error) {
	rawCfg, hasKey := self[GENERAL_JSN]
	if !hasKey {
		return nil, nil
	}
	cfg := new(GeneralJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) ListenJsonCfg() (*ListenJsonCfg, error) {
	rawCfg, hasKey := self[LISTEN_JSN]
	if !hasKey {
		return nil, nil
	}
	cfg := new(ListenJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) DbJsonCfg(section string) (*DbJsonCfg, error) {
	rawCfg, hasKey := self[section]
	if !hasKey {
		return nil, nil
	}
	cfg := new(DbJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) BalancerJsonCfg() (*BalancerJsonCfg, error) {
	rawCfg, hasKey := self[BALANCER_JSN]
	if !hasKey {
		return nil, nil
	}
	cfg := new(BalancerJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) RalsJsonCfg() (*RalsJsonCfg, error) {
	rawCfg, hasKey := self[RALS_JSN]
	if !hasKey {
		return nil, nil
	}
	cfg := new(RalsJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) SchedulerJsonCfg() (*SchedulerJsonCfg, error) {
	rawCfg, hasKey := self[SCHEDULER_JSN]
	if !hasKey {
		return nil, nil
	}
	cfg := new(SchedulerJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) CdrsJsonCfg() (*CdrsJsonCfg, error) {
	rawCfg, hasKey := self[CDRS_JSN]
	if !hasKey {
		return nil, nil
	}
	cfg := new(CdrsJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) CdrStatsJsonCfg() (*CdrStatsJsonCfg, error) {
	rawCfg, hasKey := self[CDRSTATS_JSN]
	if !hasKey {
		return nil, nil
	}
	cfg := new(CdrStatsJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) CdreJsonCfgs() (map[string]*CdreJsonCfg, error) {
	rawCfg, hasKey := self[CDRE_JSN]
	if !hasKey {
		return nil, nil
	}
	cfg := make(map[string]*CdreJsonCfg)
	if err := json.Unmarshal(*rawCfg, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) CdrcJsonCfg() (map[string]*CdrcJsonCfg, error) {
	rawCfg, hasKey := self[CDRC_JSN]
	if !hasKey {
		return nil, nil
	}
	cfg := make(map[string]*CdrcJsonCfg)
	if err := json.Unmarshal(*rawCfg, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) SmGenericJsonCfg() (*SmGenericJsonCfg, error) {
	rawCfg, hasKey := self[SMGENERIC_JSON]
	if !hasKey {
		return nil, nil
	}
	cfg := new(SmGenericJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) SmFsJsonCfg() (*SmFsJsonCfg, error) {
	rawCfg, hasKey := self[SMFS_JSN]
	if !hasKey {
		return nil, nil
	}
	cfg := new(SmFsJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) SmKamJsonCfg() (*SmKamJsonCfg, error) {
	rawCfg, hasKey := self[SMKAM_JSN]
	if !hasKey {
		return nil, nil
	}
	cfg := new(SmKamJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) SmOsipsJsonCfg() (*SmOsipsJsonCfg, error) {
	rawCfg, hasKey := self[SMOSIPS_JSN]
	if !hasKey {
		return nil, nil
	}
	cfg := new(SmOsipsJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) DiameterAgentJsonCfg() (*DiameterAgentJsonCfg, error) {
	rawCfg, hasKey := self[DA_JSN]
	if !hasKey {
		return nil, nil
	}
	cfg := new(DiameterAgentJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) HistServJsonCfg() (*HistServJsonCfg, error) {
	rawCfg, hasKey := self[HISTSERV_JSN]
	if !hasKey {
		return nil, nil
	}
	cfg := new(HistServJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) PubSubServJsonCfg() (*PubSubServJsonCfg, error) {
	rawCfg, hasKey := self[PUBSUBSERV_JSN]
	if !hasKey {
		return nil, nil
	}
	cfg := new(PubSubServJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) AliasesServJsonCfg() (*AliasesServJsonCfg, error) {
	rawCfg, hasKey := self[ALIASESSERV_JSN]
	if !hasKey {
		return nil, nil
	}
	cfg := new(AliasesServJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) UserServJsonCfg() (*UserServJsonCfg, error) {
	rawCfg, hasKey := self[USERSERV_JSN]
	if !hasKey {
		return nil, nil
	}
	cfg := new(UserServJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) MailerJsonCfg() (*MailerJsonCfg, error) {
	rawCfg, hasKey := self[MAILER_JSN]
	if !hasKey {
		return nil, nil
	}
	cfg := new(MailerJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) SureTaxJsonCfg() (*SureTaxJsonCfg, error) {
	rawCfg, hasKey := self[SURETAX_JSON]
	if !hasKey {
		return nil, nil
	}
	cfg := new(SureTaxJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
