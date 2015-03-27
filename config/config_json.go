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
	"github.com/DisposaBoy/JsonConfigReader"
	"io"
	"os"
)

const (
	GENERAL_JSN      = "general"
	LISTEN_JSN       = "listen"
	RATINGDB_JSN     = "rating_db"
	ACCOUNTINGDB_JSN = "accounting_db"
	STORDB_JSN       = "stor_db"
	BALANCER_JSN     = "balancer"
	RATER_JSN        = "rater"
	SCHEDULER_JSN    = "scheduler"
	CDRS_JSN         = "cdrs"
	MEDIATOR_JSN     = "mediator"
	CDRSTATS_JSN     = "cdr_stats"
	CDRE_JSN         = "cdre"
	CDRC_JSN         = "cdrc"
	SMFS_JSN         = "sm_freeswitch"
	SMKAM_JSN        = "sm_kamailio"
	SMOSIPS_JSN      = "sm_opensips"
	SM_JSN           = "session_manager"
	FS_JSN           = "freeswitch"
	KAMAILIO_JSN     = "kamailio"
	OSIPS_JSN        = "opensips"
	HISTSERV_JSN     = "history_server"
	HISTAGENT_JSN    = "history_agent"
	MAILER_JSN       = "mailer"
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

func (self CgrJsonCfg) RaterJsonCfg() (*RaterJsonCfg, error) {
	rawCfg, hasKey := self[RATER_JSN]
	if !hasKey {
		return nil, nil
	}
	cfg := new(RaterJsonCfg)
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

func (self CgrJsonCfg) HistAgentJsonCfg() (*HistAgentJsonCfg, error) {
	rawCfg, hasKey := self[HISTAGENT_JSN]
	if !hasKey {
		return nil, nil
	}
	cfg := new(HistAgentJsonCfg)
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
