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
		Enabled:    true,
		StatSConns: []string{utils.MetaInternal},
	}
	expected := "<StatS> not enabled but requested by <RALs> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.ralsCfg.StatSConns = []string{"test"}
	expected = "<RALs> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.ralsCfg.StatSConns = []string{}
	cfg.ralsCfg.ThresholdSConns = []string{utils.MetaInternal}
	expected = "<ThresholdS> not enabled but requested by <RALs> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.ralsCfg.ThresholdSConns = []string{"test"}
	expected = "<RALs> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
}

func TestConfigSanityCDRServer(t *testing.T) {
	cfg, _ := NewDefaultCGRConfig()

	cfg.cdrsCfg = &CdrsCfg{
		Enabled:       true,
		ChargerSConns: []string{utils.MetaInternal},
	}
	expected := "<ChargerS> not enabled but requested by <CDRs> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.cdrsCfg.ChargerSConns = []string{"test"}
	expected = "<CDRs> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.cdrsCfg.ChargerSConns = []string{}

	cfg.cdrsCfg.RaterConns = []string{utils.MetaInternal}

	expected = "<RALs> not enabled but requested by <CDRs> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.cdrsCfg.RaterConns = []string{"test"}
	expected = "<CDRs> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.cdrsCfg.RaterConns = []string{}

	cfg.cdrsCfg.AttributeSConns = []string{utils.MetaInternal}
	expected = "<AttributeS> not enabled but requested by <CDRs> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.cdrsCfg.AttributeSConns = []string{"test"}
	expected = "<CDRs> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.cdrsCfg.AttributeSConns = []string{}

	cfg.cdrsCfg.StatSConns = []string{utils.MetaInternal}
	expected = "<StatS> not enabled but requested by <CDRs> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.cdrsCfg.StatSConns = []string{"test"}
	expected = "<CDRs> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.cdrsCfg.StatSConns = []string{}

	cfg.cdrsCfg.OnlineCDRExports = []string{"stringy"}
	cfg.CdreProfiles = map[string]*CdreCfg{"stringx": &CdreCfg{}}
	expected = "<CDRs> cannot find CDR export template with ID: <stringy>"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.cdrsCfg.ThresholdSConns = []string{"test"}
	expected = "<CDRs> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.cdrsCfg.ThresholdSConns = []string{}

	cfg.cdrsCfg.ThresholdSConns = []string{utils.MetaInternal}
	expected = "<ThresholdS> not enabled but requested by <CDRs> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
}

func TestConfigSanityLoaders(t *testing.T) {
	cfg, _ = NewDefaultCGRConfig()
	cfg.loaderCfg = LoaderSCfgs{
		&LoaderSCfg{
			Enabled: true,
			TpInDir: "/not/exist",
			Data: []*LoaderDataType{
				&LoaderDataType{
					Type: "strsdfing",
				},
			},
		},
	}
	expected := "<LoaderS> nonexistent folder: /not/exist"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.loaderCfg = LoaderSCfgs{
		&LoaderSCfg{
			Enabled:  true,
			TpInDir:  "/",
			TpOutDir: "/",
			Data: []*LoaderDataType{
				&LoaderDataType{
					Type: "wrongtype",
				},
			},
		},
	}
	expected = "<LoaderS> unsupported data type wrongtype"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.loaderCfg = LoaderSCfgs{
		&LoaderSCfg{
			Enabled:  true,
			TpInDir:  "/",
			TpOutDir: "/",
			Data: []*LoaderDataType{
				&LoaderDataType{
					Type: utils.MetaStats,
					Fields: []*FCTemplate{
						&FCTemplate{
							Type: utils.MetaStats,
							Tag:  "test1",
						},
					},
				},
			},
		},
	}
	expected = "<LoaderS> invalid field type *stats for *stats at test1"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
}

func TestConfigSanitySessionS(t *testing.T) {
	cfg, _ = NewDefaultCGRConfig()
	cfg.sessionSCfg = &SessionSCfg{
		Enabled:           true,
		TerminateAttempts: 0,
	}
	expected := "<SessionS> 'terminate_attempts' should be at least 1"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.sessionSCfg.TerminateAttempts = 1

	cfg.sessionSCfg.ChargerSConns = []string{utils.MetaInternal}
	expected = "<ChargerS> not enabled but requested by <SessionS> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.sessionSCfg.ChargerSConns = []string{"test"}
	expected = "<SessionS> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.sessionSCfg.ChargerSConns = []string{}
	cfg.chargerSCfg.Enabled = true

	cfg.sessionSCfg.RALsConns = []string{utils.MetaInternal}
	expected = "<RALs> not enabled but requested by <SessionS> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.sessionSCfg.RALsConns = []string{"test"}
	expected = "<SessionS> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.sessionSCfg.RALsConns = []string{}
	cfg.ralsCfg.Enabled = true

	cfg.sessionSCfg.ResSConns = []string{utils.MetaInternal}
	expected = "<ResourceS> not enabled but requested by <SessionS> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.sessionSCfg.ResSConns = []string{"test"}
	expected = "<SessionS> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.sessionSCfg.ResSConns = []string{}
	cfg.resourceSCfg.Enabled = true

	cfg.sessionSCfg.ThreshSConns = []string{utils.MetaInternal}
	expected = "<ThresholdS> not enabled but requested by <SessionS> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.sessionSCfg.ThreshSConns = []string{"test"}
	expected = "<SessionS> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.sessionSCfg.ThreshSConns = []string{}
	cfg.thresholdSCfg.Enabled = true

	cfg.sessionSCfg.StatSConns = []string{utils.MetaInternal}
	expected = "<StatS> not enabled but requested by <SessionS> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.sessionSCfg.StatSConns = []string{"test"}
	expected = "<SessionS> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.sessionSCfg.StatSConns = []string{}
	cfg.statsCfg.Enabled = true

	cfg.sessionSCfg.RouteSConns = []string{utils.MetaInternal}
	expected = "<RouteS> not enabled but requested by <SessionS> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.sessionSCfg.RouteSConns = []string{"test"}
	expected = "<SessionS> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.sessionSCfg.RouteSConns = []string{}
	cfg.routeSCfg.Enabled = true

	cfg.sessionSCfg.AttrSConns = []string{utils.MetaInternal}
	expected = "<AttributeS> not enabled but requested by <SessionS> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.sessionSCfg.AttrSConns = []string{"test"}
	expected = "<SessionS> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.sessionSCfg.AttrSConns = []string{}
	cfg.attributeSCfg.Enabled = true

	cfg.sessionSCfg.CDRsConns = []string{utils.MetaInternal}
	expected = "<CDRs> not enabled but requested by <SessionS> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.sessionSCfg.CDRsConns = []string{"test"}
	expected = "<SessionS> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.sessionSCfg.CDRsConns = []string{}
	cfg.cdrsCfg.Enabled = true
	cfg.sessionSCfg.ReplicationConns = []string{"test"}
	expected = "<SessionS> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.sessionSCfg.ReplicationConns = []string{}

	cfg.cacheCfg.Partitions[utils.CacheClosedSessions].Limit = 0
	expected = "<CacheS> *closed_sessions needs to be != 0, received: 0"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.cacheCfg.Partitions[utils.CacheClosedSessions].Limit = 1
	expected = "<SessionS> the following protected field can't be altered by session: <CGRID>"
	cfg.sessionSCfg.AlterableFields = utils.NewStringSet([]string{utils.CGRID})
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
}

func TestConfigSanityFreeSWITCHAgent(t *testing.T) {
	cfg, _ = NewDefaultCGRConfig()
	cfg.fsAgentCfg = &FsAgentCfg{
		Enabled: true,
	}

	expected := "<FreeSWITCHAgent> no SessionS connections defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.fsAgentCfg = &FsAgentCfg{
		Enabled:       true,
		SessionSConns: []string{utils.MetaInternal},
	}
	expected = "<SessionS> not enabled but requested by <FreeSWITCHAgent> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.fsAgentCfg.SessionSConns = []string{"test"}
	expected = "<FreeSWITCHAgent> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
}

func TestConfigSanityKamailioAgent(t *testing.T) {
	cfg, _ = NewDefaultCGRConfig()
	cfg.kamAgentCfg = &KamAgentCfg{
		Enabled: true,
	}
	expected := "<KamailioAgent> no SessionS connections defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.kamAgentCfg.SessionSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)}
	expected = "<SessionS> not enabled but requested by <KamailioAgent> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.kamAgentCfg.SessionSConns = []string{"test"}
	expected = "<KamailioAgent> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
}

func TestConfigSanityAsteriskAgent(t *testing.T) {
	cfg, _ = NewDefaultCGRConfig()
	cfg.asteriskAgentCfg = &AsteriskAgentCfg{
		Enabled: true,
	}
	expected := "<AsteriskAgent> no SessionS connections defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.asteriskAgentCfg.SessionSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)}
	expected = "<SessionS> not enabled but requested by <AsteriskAgent> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.asteriskAgentCfg.SessionSConns = []string{"test"}
	expected = "<AsteriskAgent> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
}

func TestConfigSanityDAgent(t *testing.T) {
	cfg, _ = NewDefaultCGRConfig()
	cfg.diameterAgentCfg = &DiameterAgentCfg{
		Enabled: true,
	}
	expected := "<DiameterAgent> no SessionS connections defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.diameterAgentCfg.SessionSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)}
	expected = "<SessionS> not enabled but requested by <DiameterAgent> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.diameterAgentCfg.SessionSConns = []string{"test"}
	expected = "<DiameterAgent> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
}

func TestConfigSanityRadiusAgent(t *testing.T) {
	cfg, _ = NewDefaultCGRConfig()
	cfg.radiusAgentCfg = &RadiusAgentCfg{
		Enabled: true,
	}
	expected := "<RadiusAgent> no SessionS connections defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.radiusAgentCfg.SessionSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)}
	expected = "<SessionS> not enabled but requested by <RadiusAgent> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.radiusAgentCfg.SessionSConns = []string{"test"}
	expected = "<RadiusAgent> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
}

func TestConfigSanityDNSAgent(t *testing.T) {
	cfg, _ = NewDefaultCGRConfig()
	cfg.dnsAgentCfg = &DNSAgentCfg{
		Enabled: true,
	}
	expected := "<DNSAgent> no SessionS connections defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.dnsAgentCfg.SessionSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)}
	expected = "<SessionS> not enabled but requested by <DNSAgent> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.dnsAgentCfg.SessionSConns = []string{"test"}
	expected = "<DNSAgent> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
}

func TestConfigSanityHTTPAgent(t *testing.T) {
	cfg, _ = NewDefaultCGRConfig()
	cfg.sessionSCfg.Enabled = false
	cfg.httpAgentCfg = HttpAgentCfgs{
		&HttpAgentCfg{
			ID:            "Test",
			SessionSConns: []string{utils.MetaInternal},
		},
	}
	expected := "<SessionS> not enabled but requested by <Test> HTTPAgent Template."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.httpAgentCfg[0].SessionSConns = []string{"test"}
	expected = "<HTTPAgent> template with ID <Test> has connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.httpAgentCfg[0].SessionSConns = []string{}
	cfg.sessionSCfg.Enabled = true

	cfg.httpAgentCfg = HttpAgentCfgs{
		&HttpAgentCfg{
			RequestPayload: "test",
		},
	}
	expected = "<HTTPAgent> unsupported request payload test"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.httpAgentCfg = HttpAgentCfgs{
		&HttpAgentCfg{
			RequestPayload: utils.MetaUrl,
			ReplyPayload:   "test",
		},
	}
	expected = "<HTTPAgent> unsupported reply payload test"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.httpAgentCfg[0].ReplyPayload = utils.MetaTextPlain

	cfg.attributeSCfg = &AttributeSCfg{
		Enabled:     true,
		ProcessRuns: 0,
	}
	expected = "<AttributeS> process_runs needs to be bigger than 0"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.attributeSCfg = &AttributeSCfg{
		ProcessRuns: 1,
		Enabled:     false,
	}
	cfg.chargerSCfg.Enabled = true
	cfg.attributeSCfg.Enabled = false
	cfg.chargerSCfg.AttributeSConns = []string{utils.MetaInternal}
	expected = "<AttributeS> not enabled but requested by <ChargerS> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.chargerSCfg.AttributeSConns = []string{"Invalid"}

	expected = "<ChargerS> connection with id: <Invalid> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
}

func TestConfigSanityResourceLimiter(t *testing.T) {
	cfg, _ = NewDefaultCGRConfig()
	cfg.resourceSCfg = &ResourceSConfig{
		Enabled:         true,
		ThresholdSConns: []string{utils.MetaInternal},
	}
	expected := "<ThresholdS> not enabled but requested by <ResourceS> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.resourceSCfg.ThresholdSConns = []string{"test"}
	expected = "<ResourceS> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
}

func TestConfigSanityStatS(t *testing.T) {
	cfg, _ = NewDefaultCGRConfig()
	cfg.statsCfg = &StatSCfg{
		Enabled:         true,
		ThresholdSConns: []string{utils.MetaInternal},
	}
	expected := "<ThresholdS> not enabled but requested by <Stats> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.statsCfg.ThresholdSConns = []string{"test"}
	expected = "<Stats> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
}

func TestConfigSanityRouteS(t *testing.T) {
	cfg, _ = NewDefaultCGRConfig()
	cfg.routeSCfg.Enabled = true

	cfg.routeSCfg.ResourceSConns = []string{utils.MetaInternal}

	expected := "<ResourceS> not enabled but requested by <RouteS> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.routeSCfg.ResourceSConns = []string{"test"}
	expected = "<RouteS> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.routeSCfg.ResourceSConns = []string{utils.MetaInternal}
	cfg.resourceSCfg.Enabled = true

	cfg.routeSCfg.StatSConns = []string{utils.MetaInternal}
	expected = "<StatS> not enabled but requested by <RouteS> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.routeSCfg.StatSConns = []string{"test"}
	expected = "<RouteS> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.routeSCfg.StatSConns = []string{utils.MetaInternal}
	cfg.statsCfg.Enabled = true

	cfg.routeSCfg.AttributeSConns = []string{utils.MetaInternal}
	expected = "<AttributeS> not enabled but requested by <RouteS> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.routeSCfg.AttributeSConns = []string{"test"}
	expected = "<RouteS> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
}

func TestConfigSanityScheduler(t *testing.T) {
	cfg, _ = NewDefaultCGRConfig()
	cfg.schedulerCfg = &SchedulerCfg{
		Enabled:   true,
		CDRsConns: []string{utils.MetaInternal},
	}
	expected := "<CDRs> not enabled but requested by <SchedulerS> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.schedulerCfg.CDRsConns = []string{"test"}
	expected = "<SchedulerS> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

}

func TestConfigSanityEventReader(t *testing.T) {
	cfg, _ = NewDefaultCGRConfig()
	cfg.ersCfg = &ERsCfg{
		Enabled:       true,
		SessionSConns: []string{"unexistedConn"},
	}
	expected := "<ERs> connection with id: <unexistedConn> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.ersCfg.SessionSConns = []string{utils.MetaInternal}
	expected = "<SessionS> not enabled but requested by <ERs> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.sessionSCfg.Enabled = true
	cfg.ersCfg.Readers = []*EventReaderCfg{
		&EventReaderCfg{
			ID:   "test",
			Type: "wrongtype",
		},
	}
	expected = "<ERs> unsupported data type: wrongtype for reader with ID: test"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.ersCfg.Readers = []*EventReaderCfg{
		&EventReaderCfg{
			ID:            "test2",
			Type:          utils.MetaFileCSV,
			ProcessedPath: "not/a/path",
		},
	}
	expected = "<ERs> nonexistent folder: not/a/path for reader with ID: test2"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.ersCfg.Readers = []*EventReaderCfg{
		&EventReaderCfg{
			ID:            "test3",
			Type:          utils.MetaFileCSV,
			ProcessedPath: "/",
			SourcePath:    "/",
			FieldSep:      "",
		},
	}
	expected = "<ERs> empty FieldSep for reader with ID: test3"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.ersCfg.Readers[0] = &EventReaderCfg{
		ID:       "test4",
		Type:     utils.MetaKafkajsonMap,
		RunDelay: 1,
		FieldSep: utils.InInFieldSep,
	}
	expected = "<ERs> the RunDelay field can not be bigger than zero for reader with ID: test4"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.ersCfg.Readers[0] = &EventReaderCfg{
		ID:            "test5",
		Type:          utils.MetaFileXML,
		RunDelay:      0,
		FieldSep:      utils.InInFieldSep,
		ProcessedPath: "not/a/path",
		SourcePath:    "not/a/path",
	}
	expected = "<ERs> nonexistent folder: not/a/path for reader with ID: test5"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.ersCfg.Readers[0] = &EventReaderCfg{
		ID:            "test5",
		Type:          utils.MetaFileFWV,
		RunDelay:      0,
		FieldSep:      utils.InInFieldSep,
		ProcessedPath: "not/a/path",
		SourcePath:    "not/a/path",
	}
	expected = "<ERs> nonexistent folder: not/a/path for reader with ID: test5"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
}

func TestConfigSanityStorDB(t *testing.T) {
	cfg, _ = NewDefaultCGRConfig()
	cfg.storDbCfg = &StorDbCfg{
		Type:    utils.POSTGRES,
		SSLMode: "wrongSSLMode",
	}
	expected := "<stor_db> unsuported sslmode for storDB"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
}

func TestConfigSanityDataDB(t *testing.T) {
	cfg, _ = NewDefaultCGRConfig()
	cfg.dataDbCfg.DataDbType = utils.INTERNAL

	cfg.cacheCfg = &CacheCfg{
		Partitions: map[string]*CacheParamCfg{
			utils.CacheTimings: &CacheParamCfg{
				Limit: 0,
			},
		},
	}
	if err := cfg.checkConfigSanity(); err != nil {
		t.Error(err)
	}

	cfg.cacheCfg = &CacheCfg{
		Partitions: map[string]*CacheParamCfg{
			utils.CacheAccounts: &CacheParamCfg{
				Limit: 1,
			},
		},
	}
	expected := "<CacheS> *accounts needs to be 0 when DataBD is *internal, received : 1"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.cacheCfg.Partitions[utils.CacheAccounts].Limit = 0
	cfg.resourceSCfg.Enabled = true
	expected = "<ResourceS> the StoreInterval field needs to be -1 when DataBD is *internal, received : 0"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.resourceSCfg.Enabled = false

	cfg.statsCfg.Enabled = true
	expected = "<Stats> the StoreInterval field needs to be -1 when DataBD is *internal, received : 0"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.statsCfg.Enabled = false

	cfg.thresholdSCfg.Enabled = true
	expected = "<ThresholdS> the StoreInterval field needs to be -1 when DataBD is *internal, received : 0"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.thresholdSCfg.Enabled = false

	cfg.dataDbCfg.Items = map[string]*ItemOpt{
		"test1": &ItemOpt{
			Remote: true,
		},
	}
	expected = "remote connections required by: <test1>"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.dataDbCfg.Items = map[string]*ItemOpt{
		"test2": &ItemOpt{
			Remote:    false,
			Replicate: true,
		},
	}
	expected = "replicate connections required by: <test2>"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

}

func TestConfigSanityAPIer(t *testing.T) {
	cfg, _ = NewDefaultCGRConfig()
	cfg.apier.AttributeSConns = []string{utils.MetaInternal}

	if err := cfg.checkConfigSanity(); err == nil || err.Error() != "<AttributeS> not enabled but requested by <APIerSv1> component." {
		t.Error(err)
	}
	cfg.apier.AttributeSConns = []string{"test"}
	expected := "<APIerSv1> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.apier.AttributeSConns = []string{utils.MetaInternal}
	cfg.attributeSCfg.Enabled = true
	cfg.apier.SchedulerConns = []string{utils.MetaInternal}

	if err := cfg.checkConfigSanity(); err == nil || err.Error() != "<SchedulerS> not enabled but requested by <APIerSv1> component." {
		t.Error(err)
	}
	cfg.apier.SchedulerConns = []string{"test"}
	expected = "<APIerSv1> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
}

func TestConfigSanityDispatcher(t *testing.T) {
	cfg, _ = NewDefaultCGRConfig()
	cfg.dispatcherSCfg = &DispatcherSCfg{
		Enabled:         true,
		AttributeSConns: []string{utils.MetaInternal},
	}
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != "<AttributeS> not enabled but requested by <DispatcherS> component." {
		t.Error(err)
	}
	cfg.dispatcherSCfg.AttributeSConns = []string{"test"}
	expected := "<DispatcherS> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
}

func TestConfigSanityCacheS(t *testing.T) {
	cfg, _ = NewDefaultCGRConfig()

	cfg.cacheCfg.Partitions = map[string]*CacheParamCfg{"wrong_partition_name": &CacheParamCfg{Limit: 10}}
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != "<CacheS> partition <wrong_partition_name> not defined" {
		t.Error(err)
	}

	cfg.cacheCfg.Partitions = map[string]*CacheParamCfg{utils.CacheLoadIDs: &CacheParamCfg{Limit: 9}}
	if err := cfg.checkConfigSanity(); err != nil {
		t.Error(err)
	}
}

func TestConfigSanityFilterS(t *testing.T) {
	cfg, _ = NewDefaultCGRConfig()
	cfg.filterSCfg.StatSConns = []string{utils.MetaInternal}

	if err := cfg.checkConfigSanity(); err == nil || err.Error() != "<Stats> not enabled but requested by <FilterS> component." {
		t.Error(err)
	}
	cfg.filterSCfg.StatSConns = []string{"test"}
	expected := "<FilterS> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.filterSCfg.StatSConns = []string{}

	cfg.filterSCfg.ResourceSConns = []string{utils.MetaInternal}

	if err := cfg.checkConfigSanity(); err == nil || err.Error() != "<ResourceS> not enabled but requested by <FilterS> component." {
		t.Error(err)
	}
	cfg.filterSCfg.ResourceSConns = []string{"test"}
	expected = "<FilterS> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.filterSCfg.ResourceSConns = []string{}

	cfg.filterSCfg.ApierSConns = []string{utils.MetaInternal}

	if err := cfg.checkConfigSanity(); err == nil || err.Error() != "<ApierS> not enabled but requested by <FilterS> component." {
		t.Error(err)
	}
	cfg.filterSCfg.ApierSConns = []string{"test"}
	expected = "<FilterS> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
}
