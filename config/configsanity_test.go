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
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestConfigSanityRater(t *testing.T) {
	cfg, _ := NewDefaultCGRConfig()

	cfg.ralsCfg = &RalsCfg{
		Enabled: true,
		StatSConns: []*RemoteHost{
			&RemoteHost{
				Address: utils.MetaInternal,
			},
		},
	}
	expected := "<Stats> not enabled but requested by <RALs> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.statsCfg.Enabled = true

	cfg.ralsCfg.ThresholdSConns = []*RemoteHost{
		&RemoteHost{
			Address: utils.MetaInternal,
		},
	}
	expected = "<ThresholdS> not enabled but requested by <RALs> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
}

func TestConfigSanityCDRServer(t *testing.T) {
	cfg, _ := NewDefaultCGRConfig()

	cfg.cdrsCfg = &CdrsCfg{
		Enabled: true,
		ChargerSConns: []*RemoteHost{
			&RemoteHost{
				Address: utils.MetaInternal,
			},
		},
	}
	expected := "<Chargers> not enabled but requested by <CDRs> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.chargerSCfg.Enabled = true

	cfg.cdrsCfg.RaterConns = []*RemoteHost{
		&RemoteHost{
			Address: utils.MetaInternal,
		},
	}
	expected = "<RALs> not enabled but requested by <CDRs> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.ralsCfg.Enabled = true

	cfg.cdrsCfg.AttributeSConns = []*RemoteHost{
		&RemoteHost{
			Address: utils.MetaInternal,
		},
	}
	expected = "<AttributeS> not enabled but requested by <CDRs> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.attributeSCfg.Enabled = true

	cfg.cdrsCfg.StatSConns = []*RemoteHost{
		&RemoteHost{
			Address: utils.MetaInternal,
		},
	}
	expected = "<StatS> not enabled but requested by <CDRs> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.statsCfg.Enabled = true

	cfg.cdrsCfg.OnlineCDRExports = []string{"stringy"}
	cfg.CdreProfiles = map[string]*CdreCfg{"stringx": &CdreCfg{}}
	expected = "<CDRs> Cannot find CDR export template with ID: <stringy>"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.cdrsCfg.OnlineCDRExports = []string{"stringx"}

	cfg.cdrsCfg.ThresholdSConns = []*RemoteHost{
		&RemoteHost{
			Address: utils.MetaInternal,
		},
	}
	expected = "ThresholdS not enabled but requested by CDRs component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
}

func TestConfigSanityCDRC(t *testing.T) {
	cfg, _ = NewDefaultCGRConfig()
	cfg.CdrcProfiles = map[string][]*CdrcCfg{
		"test": []*CdrcCfg{
			&CdrcCfg{
				Enabled: true,
			},
		},
	}
	expected := "<cdrc> Instance: , cdrc enabled but no CDRs defined!"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.CdrcProfiles = map[string][]*CdrcCfg{
		"test": []*CdrcCfg{
			&CdrcCfg{
				Enabled: true,
				CdrsConns: []*RemoteHost{
					&RemoteHost{Address: utils.MetaInternal},
				},
			},
		},
	}
	expected = "<CDRs> not enabled but referenced from <cdrc>"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.CdrcProfiles = map[string][]*CdrcCfg{
		"test": []*CdrcCfg{
			&CdrcCfg{
				Enabled: true,
				CdrsConns: []*RemoteHost{
					&RemoteHost{Address: utils.MetaInternal},
				},
				ContentFields: []*FCTemplate{},
			},
		},
	}
	cfg.cdrsCfg.Enabled = true
	expected = "<cdrc> enabled but no fields to be processed defined!"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
}
