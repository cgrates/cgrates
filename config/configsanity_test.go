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
	cfg.statsCfg.Enabled = true

	cfg.ralsCfg.ThresholdSConns = []string{utils.MetaInternal}
	expected = "<ThresholdS> not enabled but requested by <RALs> component."
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
	cfg.chargerSCfg.Enabled = true

	cfg.cdrsCfg.RaterConns = []string{utils.MetaInternal}

	expected = "<RALs> not enabled but requested by <CDRs> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.ralsCfg.Enabled = true

	cfg.cdrsCfg.AttributeSConns = []string{utils.MetaInternal}
	expected = "<AttributeS> not enabled but requested by <CDRs> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.attributeSCfg.Enabled = true

	cfg.cdrsCfg.StatSConns = []string{utils.MetaInternal}
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

	cfg.cdrsCfg.ThresholdSConns = []string{utils.MetaInternal}
	expected = "<ThresholdS> not enabled but requested by <CDRs> component."
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
				Enabled:   true,
				ID:        "test",
				CdrsConns: []string{utils.MetaInternal},
			},
		},
	}
	expected = "<CDRs> not enabled but requested by <test> cdrcProfile"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.CdrcProfiles = map[string][]*CdrcCfg{
		"test": []*CdrcCfg{
			&CdrcCfg{
				Enabled:       true,
				ID:            "test",
				CdrsConns:     []string{utils.MetaInternal},
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
	expected := "<LoaderS> Nonexistent folder: /not/exist"
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
	cfg.chargerSCfg.Enabled = true

	cfg.sessionSCfg.RALsConns = []string{utils.MetaInternal}
	expected = "<RALs> not enabled but requested by <SessionS> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.ralsCfg.Enabled = true

	cfg.sessionSCfg.ResSConns = []string{utils.MetaInternal}
	expected = "<ResourceS> not enabled but requested by <SessionS> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.resourceSCfg.Enabled = true

	cfg.sessionSCfg.ThreshSConns = []string{utils.MetaInternal}
	expected = "<ThresholdS> not enabled but requested by <SessionS> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.thresholdSCfg.Enabled = true

	cfg.sessionSCfg.StatSConns = []string{utils.MetaInternal}
	expected = "<StatS> not enabled but requested by <SessionS> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.statsCfg.Enabled = true

	cfg.sessionSCfg.SupplSConns = []string{utils.MetaInternal}
	expected = "<SupplierS> not enabled but requested by <SessionS> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.supplierSCfg.Enabled = true

	cfg.sessionSCfg.AttrSConns = []string{utils.MetaInternal}
	expected = "<AttributeS> not enabled but requested by <SessionS> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.attributeSCfg.Enabled = true

	cfg.sessionSCfg.CDRsConns = []string{utils.MetaInternal}
	expected = "<CDRs> not enabled but requested by <SessionS> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.cdrsCfg.Enabled = true

	cfg.cacheCfg[utils.CacheClosedSessions].Limit = 0
	expected = "<CacheS> *closed_sessions needs to be != 0, received: 0"
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
		Enabled: true,
		SessionSConns: []*RemoteHost{
			&RemoteHost{
				Address: utils.MetaInternal,
			},
		},
	}
	expected = "<SessionS> not enabled but referenced by <FreeSWITCHAgent>"
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

	cfg.kamAgentCfg.SessionSConns = []*RemoteHost{
		&RemoteHost{
			Address: utils.MetaInternal,
		},
	}
	expected = "<SessionS> not enabled but referenced by <KamailioAgent>"
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

	cfg.asteriskAgentCfg.SessionSConns = []*RemoteHost{
		&RemoteHost{
			Address: utils.MetaInternal,
		},
	}
	expected = "<SessionS> not enabled but referenced by <AsteriskAgent>"
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

	cfg.diameterAgentCfg.SessionSConns = []*RemoteHost{
		&RemoteHost{
			Address: utils.MetaInternal,
		},
	}
	expected = "<SessionS> not enabled but referenced by <DiameterAgent>"
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

	cfg.radiusAgentCfg.SessionSConns = []*RemoteHost{
		&RemoteHost{
			Address: utils.MetaInternal,
		},
	}
	expected = "<SessionS> not enabled but referenced by <RadiusAgent>"
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

	cfg.dnsAgentCfg.SessionSConns = []*RemoteHost{
		&RemoteHost{
			Address: utils.MetaInternal,
		},
	}
	expected = "<SessionS> not enabled but referenced by <DNSAgent>"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
}

func TestConfigSanityHTTPAgent(t *testing.T) {
	cfg, _ = NewDefaultCGRConfig()
	cfg.sessionSCfg.Enabled = true
	cfg.httpAgentCfg = HttpAgentCfgs{
		&HttpAgentCfg{
			SessionSConns: []*RemoteHost{
				&RemoteHost{
					Address: utils.MetaInternal,
				},
			},
		},
	}
	expected := "<SessionS> not enabled but referenced by <HTTPAgent> component"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.sessionSCfg.Enabled = false

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
	cfg.chargerSCfg.AttributeSConns = []string{"Invalid"}

	expected = "<ChargerS> Connection with id: <Invalid> not defined"
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
}

func TestConfigSanitySupplierS(t *testing.T) {
	cfg, _ = NewDefaultCGRConfig()
	cfg.supplierSCfg.Enabled = true

	cfg.supplierSCfg.ResourceSConns = []string{utils.MetaInternal}

	expected := "<ResourceS> not enabled but requested by <SupplierS> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.resourceSCfg.Enabled = true

	cfg.supplierSCfg.StatSConns = []string{utils.MetaInternal}
	expected = "<StatS> not enabled but requested by <SupplierS> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.statsCfg.Enabled = true

	cfg.supplierSCfg.AttributeSConns = []string{utils.MetaInternal}
	expected = "<AttributeS> not enabled but requested by <SupplierS> component."
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

}

func TestConfigSanityEventReader(t *testing.T) {
	cfg, _ = NewDefaultCGRConfig()
	cfg.ersCfg = &ERsCfg{
		Enabled:       true,
		SessionSConns: []string{"unexistedConn"},
	}
	expected := "<ERs> Connection with id: <unexistedConn> not defined"
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
	expected = "<ERs> Nonexistent folder: not/a/path for reader with ID: test2"
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
	cfg.ersCfg.Readers[0].FieldSep = utils.InInFieldSep

	cfg.ersCfg.Readers[0].ID = "test4"
	cfg.ersCfg.Readers[0].Type = utils.MetaKafkajsonMap
	cfg.ersCfg.Readers[0].RunDelay = 1
	expected = "<ERs> RunDelay field can not be bigger than zero for reader with ID: test4"
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
	expected := "<stor_db> Unsuported sslmode for storDB"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
}

func TestConfigSanityDataDB(t *testing.T) {
	cfg, _ = NewDefaultCGRConfig()
	cfg.dataDbCfg.DataDbType = utils.INTERNAL
	cfg.cacheCfg = CacheCfg{
		utils.CacheDiameterMessages: &CacheParamCfg{
			Limit: 0,
		},
	}
	expected := "<CacheS> *diameter_messages needs to be != 0 when DataBD is *internal, found 0."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.cacheCfg = CacheCfg{
		utils.CacheDiameterMessages: &CacheParamCfg{
			Limit: 1,
		},
	}
	if err := cfg.checkConfigSanity(); err != nil {
		t.Errorf("Expecting: nil  received: %+q", err)
	}

	cfg.cacheCfg = CacheCfg{
		"test": &CacheParamCfg{
			Limit: 1,
		},
	}
	expected = "<CacheS> test needs to be 0 when DataBD is *internal, received : 1"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.cacheCfg["test"].Limit = 0

	cfg.resourceSCfg.Enabled = true
	expected = "<ResourceS> StoreInterval needs to be -1 when DataBD is *internal, received : 0"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.resourceSCfg.Enabled = false

	cfg.statsCfg.Enabled = true
	expected = "<Stats> StoreInterval needs to be -1 when DataBD is *internal, received : 0"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.statsCfg.Enabled = false

	cfg.thresholdSCfg.Enabled = true
	expected = "<ThresholdS> StoreInterval needs to be -1 when DataBD is *internal, received : 0"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.thresholdSCfg.Enabled = false

	cfg.dataDbCfg.Items = map[string]*ItemRmtRplOpt{
		"test1": &ItemRmtRplOpt{
			Remote: true,
		},
	}
	expected = "Remote connections required by: <test1>"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.dataDbCfg.Items = map[string]*ItemRmtRplOpt{
		"test2": &ItemRmtRplOpt{
			Remote:    false,
			Replicate: true,
		},
	}
	expected = "Replicate connections required by: <test2>"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

}
