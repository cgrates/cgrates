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

func TestConfigSanityCDRServer(t *testing.T) {
	cfg := NewDefaultCGRConfig()

	cfg.cdrsCfg = &CdrsCfg{
		Enabled: true,
	}

	cfg.cdrsCfg.EEsConns = []string{"test"}
	expected := "<CDRs> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expected %+q, received %+q", expected, err)
	}
	cfg.cdrsCfg.EEsConns = []string{utils.MetaInternal}
	expected = "<EEs> not enabled but requested by <CDRs> component"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expected %+q, received %+q", expected, err)
	}

	cfg.cdrsCfg.ChargerSConns = []string{utils.MetaInternal}
	expected = "<ChargerS> not enabled but requested by <CDRs> component"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.cdrsCfg.ChargerSConns = []string{"test"}
	expected = "<CDRs> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.cdrsCfg.ChargerSConns = []string{}

	cfg.cdrsCfg.AttributeSConns = []string{utils.MetaInternal}
	expected = "<AttributeS> not enabled but requested by <CDRs> component"
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
	expected = "<StatS> not enabled but requested by <CDRs> component"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.cdrsCfg.StatSConns = []string{"test"}
	expected = "<CDRs> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.cdrsCfg.StatSConns = []string{}

	cfg.cdrsCfg.ThresholdSConns = []string{utils.MetaInternal}
	expected = "<ThresholdS> not enabled but requested by <CDRs> component"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.cdrsCfg.ThresholdSConns = []string{"test"}
	expected = "<CDRs> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.cdrsCfg.ThresholdSConns = []string{}

	cfg.cdrsCfg.OnlineCDRExports = []string{"*default", "stringy"}
	expected = "<CDRs> cannot find exporter with ID: <stringy>"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
}

func TestConfigSanityLoaders(t *testing.T) {
	cfg = NewDefaultCGRConfig()
	cfg.loaderCfg = LoaderSCfgs{
		&LoaderSCfg{
			Enabled: true,
			TpInDir: "/not/exist",
			Data: []*LoaderDataType{{
				Type: "strsdfing",
			}},
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
			TpOutDir: "/not/exist",
			Data: []*LoaderDataType{{
				Type: "strsdfing",
			}},
		},
	}
	expected = "<LoaderS> nonexistent folder: /not/exist"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q received: %+q", expected, err)
	}

	cfg.loaderCfg = LoaderSCfgs{
		&LoaderSCfg{
			Enabled:  true,
			TpInDir:  "/",
			TpOutDir: "/",
			Data: []*LoaderDataType{{
				Type: "wrongtype",
			}},
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
			Data: []*LoaderDataType{{
				Type: utils.MetaStats,
				Fields: []*FCTemplate{{
					Type: utils.MetaStats,
					Tag:  "test1",
				}},
			}},
		},
	}
	expected = "<LoaderS> invalid field type *stats for *stats at test1"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.loaderCfg = LoaderSCfgs{
		&LoaderSCfg{
			Enabled:  true,
			TpInDir:  "/",
			TpOutDir: "/",
			Data: []*LoaderDataType{{
				Type: utils.MetaStats,
				Fields: []*FCTemplate{{
					Type: utils.MetaComposed,
					Tag:  "test1",
					Path: utils.EmptyString,
				}},
			}},
		},
	}
	expected = "<LoaderS> MANDATORY_IE_MISSING: [Path] for *stats at test1"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.loaderCfg[0].Data[0].Fields[0].Path = "~req."
	expected = "<LoaderS> Empty field path  for ~req. at Path"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.loaderCfg[0].Data[0].Fields[0].Path = "~*req.Destination"

	cfg.loaderCfg[0].Data[0].Fields[0].Value = RSRParsers{{}}
	cfg.loaderCfg[0].Data[0].Fields[0].Value[0].path = "~req."
	expected = "<LoaderS> Empty field path  for ~req. at Values"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.loaderCfg[0].Data[0].Fields[0].Value[0].path = "*req.Destination"

	cfg.loaderCfg[0].Data[0].Fields[0].Filters = make([]string, 1)
	cfg.loaderCfg[0].Data[0].Fields[0].Filters = []string{"*string:~*req..Field"}
	expected = "<LoaderS> inline parse error for string: <*string:~*req..Field> for [*string:~*req..Field] at Filters"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.loaderCfg[0].Data[0].Fields[0].Filters = []string{"~req.Valid.Field"}
}

func TestConfigSanitySessionS(t *testing.T) {
	cfg = NewDefaultCGRConfig()
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
	expected = "<ChargerS> not enabled but requested by <SessionS> component"
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

	cfg.sessionSCfg.ResSConns = []string{utils.MetaInternal}
	expected = "<ResourceS> not enabled but requested by <SessionS> component"
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
	expected = "<ThresholdS> not enabled but requested by <SessionS> component"
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
	expected = "<StatS> not enabled but requested by <SessionS> component"
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
	expected = "<RouteS> not enabled but requested by <SessionS> component"
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
	expected = "<AttributeS> not enabled but requested by <SessionS> component"
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
	expected = "<CDRs> not enabled but requested by <SessionS> component"
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
	cfg = NewDefaultCGRConfig()
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
	expected = "<SessionS> not enabled but requested by <FreeSWITCHAgent> component"
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
	cfg = NewDefaultCGRConfig()
	cfg.kamAgentCfg = &KamAgentCfg{
		Enabled: true,
	}
	expected := "<KamailioAgent> no SessionS connections defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.kamAgentCfg.SessionSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)}
	expected = "<SessionS> not enabled but requested by <KamailioAgent> component"
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
	cfg = NewDefaultCGRConfig()
	cfg.asteriskAgentCfg = &AsteriskAgentCfg{
		Enabled: true,
	}
	expected := "<AsteriskAgent> no SessionS connections defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.asteriskAgentCfg.SessionSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)}
	expected = "<SessionS> not enabled but requested by <AsteriskAgent> component"
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
	cfg = NewDefaultCGRConfig()

	cfg.templates = FCTemplates{
		utils.MetaEEs: {
			{Tag: "SessionId", Path: utils.EmptyString, Type: "*variable",
				Value: NewRSRParsersMustCompile("~*req.Session-Id", utils.InfieldSep), Mandatory: true},
		},
	}
	cfg.diameterAgentCfg = &DiameterAgentCfg{
		Enabled: true,
		RequestProcessors: []*RequestProcessor{
			{
				ID:       "cgrates",
				Timezone: "Local",
				RequestFields: []*FCTemplate{
					{Tag: "SessionId", Path: utils.EmptyString, Type: "*variable",
						Value: NewRSRParsersMustCompile("~*req.Session-Id", utils.InfieldSep), Mandatory: true},
				},
				ReplyFields: []*FCTemplate{
					{Tag: "SessionId", Path: utils.EmptyString, Type: "*variable",
						Value: NewRSRParsersMustCompile("~*req.Session-Id", utils.InfieldSep), Mandatory: true},
				},
			},
		},
	}

	expected := "<DiameterAgent> no SessionS connections defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.diameterAgentCfg.SessionSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)}
	expected = "<SessionS> not enabled but requested by <DiameterAgent> component"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.diameterAgentCfg.SessionSConns = []string{"test"}
	expected = "<DiameterAgent> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.rpcConns["test"] = nil
	expected = "<DiameterAgent> MANDATORY_IE_MISSING: [Path] for template *ees at SessionId"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.templates[utils.MetaEEs][0].Path = "~Req."
	expected = "<DiameterAgent> Empty field path  for ~Req. at Path"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.templates[utils.MetaEEs][0].Path = "*req.OriginID"

	cfg.templates[utils.MetaEEs][0].Value[0].path = "~Req."
	expected = "<DiameterAgent> Empty field path  for ~Req. at Values"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.templates[utils.MetaEEs][0].Value[0].path = "*req.OriginID"

	cfg.templates[utils.MetaEEs][0].Filters = []string{"*string:~*req..Field"}
	expected = "<DiameterAgent> inline parse error for string: <*string:~*req..Field> for [*string:~*req..Field] at templates"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.templates[utils.MetaEEs][0].Filters = []string{"*string:~*req.VAlid.Field"}
	cfg.templates = nil

	expected = "<DiameterAgent> MANDATORY_IE_MISSING: [Path] for cgrates at SessionId"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.diameterAgentCfg.RequestProcessors[0].RequestFields[0].Type = utils.MetaNone

	cfg.diameterAgentCfg.RequestProcessors[0].RequestFields[0].Filters = []string{"*string:~*req..Field"}
	expected = "<DiameterAgent> inline parse error for string: <*string:~*req..Field> for [*string:~*req..Field] at request_fields"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.diameterAgentCfg.RequestProcessors[0].RequestFields[0].Filters = []string{"*string:~*req.Valid.Field"}

	cfg.diameterAgentCfg.RequestProcessors[0].RequestFields[0].Path = "~REQ."
	expected = "<DiameterAgent> Empty field path  for ~REQ. at Path"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.diameterAgentCfg.RequestProcessors[0].RequestFields[0].Path = "*req.Valid.Field"

	cfg.diameterAgentCfg.RequestProcessors[0].RequestFields[0].Value[0].path = "~REQ."
	expected = "<DiameterAgent> Empty field path  for ~REQ. at Values of request_fields"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.diameterAgentCfg.RequestProcessors[0].RequestFields[0].Value[0].path = "*req.Valid.Field"
	cfg.diameterAgentCfg.RequestProcessors[0].RequestFields = []*FCTemplate{}

	expected = "<DiameterAgent> MANDATORY_IE_MISSING: [Path] for cgrates at SessionId"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.diameterAgentCfg.RequestProcessors[0].ReplyFields[0].Type = utils.MetaNone

	cfg.diameterAgentCfg.RequestProcessors[0].ReplyFields[0].Filters = []string{"*string:~*req..Field"}
	expected = "<DiameterAgent> inline parse error for string: <*string:~*req..Field> for [*string:~*req..Field] at reply_fields"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.diameterAgentCfg.RequestProcessors[0].ReplyFields[0].Filters = []string{"*string:~*req.Valid.Field"}

	cfg.diameterAgentCfg.RequestProcessors[0].ReplyFields[0].Path = "~REQ."
	expected = "<DiameterAgent> Empty field path  for ~REQ. at Path"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.diameterAgentCfg.RequestProcessors[0].ReplyFields[0].Path = "*req.Valid.Field"

	cfg.diameterAgentCfg.RequestProcessors[0].ReplyFields[0].Value[0].path = "~REQ."
	expected = "<DiameterAgent> Empty field path  for ~REQ. at Values of reply_fields"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.diameterAgentCfg.RequestProcessors[0].ReplyFields[0].Value[0].path = "*req.Valid.Field"
	cfg.diameterAgentCfg.RequestProcessors[0].ReplyFields = []*FCTemplate{}

	cfg.diameterAgentCfg.RequestProcessors[0].Filters = []string{"*string:~*req..Field"}
	expected = "<DiameterAgent> inline parse error for string: <*string:~*req..Field> for [*string:~*req..Field] at request_processors"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.diameterAgentCfg.RequestProcessors[0].Filters = []string{"*string:~*req.Valid.Field"}
}

func TestConfigSanityRadiusAgent(t *testing.T) {
	cfg = NewDefaultCGRConfig()
	cfg.radiusAgentCfg = &RadiusAgentCfg{
		Enabled: true,
		RequestProcessors: []*RequestProcessor{
			{
				ID:       "cgrates",
				Timezone: "Local",
				RequestFields: []*FCTemplate{
					{Tag: "SessionId", Path: utils.EmptyString, Type: "*variable",
						Value: NewRSRParsersMustCompile("~*req.Session-Id", utils.InfieldSep), Mandatory: true},
				},
				ReplyFields: []*FCTemplate{
					{Tag: "SessionId", Path: utils.EmptyString, Type: "*variable",
						Value: NewRSRParsersMustCompile("~*req.Session-Id", utils.InfieldSep), Mandatory: true},
				},
			},
		},
	}
	expected := "<RadiusAgent> no SessionS connections defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.radiusAgentCfg.SessionSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)}
	expected = "<SessionS> not enabled but requested by <RadiusAgent> component"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.radiusAgentCfg.SessionSConns = []string{"test"}
	expected = "<RadiusAgent> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.rpcConns["test"] = nil
	expected = "<RadiusAgent> MANDATORY_IE_MISSING: [Path] for cgrates at SessionId"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.radiusAgentCfg.RequestProcessors[0].RequestFields[0].Type = utils.MetaNone

	cfg.radiusAgentCfg.RequestProcessors[0].RequestFields[0].Path = "~req."
	expected = "<RadiusAgent> Empty field path  for ~req. at Path"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.radiusAgentCfg.RequestProcessors[0].RequestFields[0].Path = "*req.OriginID"

	cfg.radiusAgentCfg.RequestProcessors[0].RequestFields[0].Value[0].path = "~req."
	expected = "<RadiusAgent> Empty field path  for ~req. at Values of request_fields"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.radiusAgentCfg.RequestProcessors[0].RequestFields[0].Value[0].path = "*req.OriginID"

	cfg.radiusAgentCfg.RequestProcessors[0].RequestFields[0].Filters = []string{"*empty:*ec"}
	expected = "<RadiusAgent> inline parse error for string: <*empty:*ec> for [*empty:*ec] at request_fields"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.radiusAgentCfg.RequestProcessors[0].RequestFields[0].Filters = []string{"*string:~*req.Account:1001"}
	cfg.radiusAgentCfg.RequestProcessors[0].RequestFields = []*FCTemplate{}

	cfg.radiusAgentCfg.RequestProcessors[0].ReplyFields[0].Path = "~req."
	expected = "<RadiusAgent> Empty field path  for ~req. at Path"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.radiusAgentCfg.RequestProcessors[0].ReplyFields[0].Path = "*req.OriginID"

	cfg.radiusAgentCfg.RequestProcessors[0].ReplyFields[0].Value[0].path = "~req."
	expected = "<RadiusAgent> Empty field path  for ~req. at Values of reply_fields"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.radiusAgentCfg.RequestProcessors[0].ReplyFields[0].Value[0].path = "*req.OriginID"

	cfg.radiusAgentCfg.RequestProcessors[0].ReplyFields[0].Filters = []string{"*empty:*ec"}
	expected = "<RadiusAgent> inline parse error for string: <*empty:*ec> for [*empty:*ec] at reply_fields"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.radiusAgentCfg.RequestProcessors[0].ReplyFields[0].Filters = []string{"*string:~*req.Account:1001"}

	cfg.radiusAgentCfg.RequestProcessors[0].ReplyFields[0].Type = "changed"
	cfg.radiusAgentCfg.RequestProcessors[0].ReplyFields[0].Path = utils.EmptyString
	expected = "<RadiusAgent> MANDATORY_IE_MISSING: [Path] for cgrates at SessionId"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.radiusAgentCfg.RequestProcessors[0].ReplyFields = []*FCTemplate{}

	cfg.radiusAgentCfg.RequestProcessors[0].Filters = []string{"*empty:*ec"}
	expected = "<RadiusAgent> inline parse error for string: <*empty:*ec> for [*empty:*ec] at request_processors"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.radiusAgentCfg.RequestProcessors[0].Filters = []string{"*string:~*req.Account:1001"}
}

func TestConfigSanityDNSAgent(t *testing.T) {
	cfg = NewDefaultCGRConfig()
	cfg.dnsAgentCfg = &DNSAgentCfg{
		Enabled: true,
		RequestProcessors: []*RequestProcessor{
			{
				ID:       "cgrates",
				Timezone: "Local",
				RequestFields: []*FCTemplate{
					{Tag: "SessionId", Path: utils.EmptyString, Type: "*variable",
						Value: NewRSRParsersMustCompile("~*req.Session-Id", utils.InfieldSep), Mandatory: true},
				},
				ReplyFields: []*FCTemplate{
					{Tag: "SessionId", Path: utils.EmptyString, Type: "*variable",
						Value: NewRSRParsersMustCompile("~*req.Session-Id", utils.InfieldSep), Mandatory: true},
				},
			},
		},
	}
	expected := "<DNSAgent> no SessionS connections defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.dnsAgentCfg.SessionSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)}
	expected = "<SessionS> not enabled but requested by <DNSAgent> component"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.dnsAgentCfg.SessionSConns = []string{"test"}
	expected = "<DNSAgent> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.rpcConns["test"] = nil
	expected = "<DNSAgent> MANDATORY_IE_MISSING: [Path] for cgrates at SessionId"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.dnsAgentCfg.RequestProcessors[0].RequestFields[0].Type = utils.MetaNone

	cfg.dnsAgentCfg.RequestProcessors[0].RequestFields[0].Value[0].path = "~req."
	expected = "<DNSAgent> Empty field path  for ~req. at Values of request_fields"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.dnsAgentCfg.RequestProcessors[0].RequestFields[0].Value[0].path = "*req.Value"

	cfg.dnsAgentCfg.RequestProcessors[0].RequestFields[0].Filters = []string{"*empty:~*req"}
	expected = "<DNSAgent> inline parse error for string: <*empty:~*req> for [*empty:~*req] at request_fields"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.dnsAgentCfg.RequestProcessors[0].RequestFields[0].Filters = []string{"*string:~*req.Account:1001"}

	cfg.dnsAgentCfg.RequestProcessors[0].RequestFields[0].Path = "~req."
	expected = "<DNSAgent> Empty field path  for ~req. at Path"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.dnsAgentCfg.RequestProcessors[0].RequestFields[0].Path = "*req.Value"
	cfg.dnsAgentCfg.RequestProcessors[0].RequestFields = []*FCTemplate{}

	cfg.dnsAgentCfg.RequestProcessors[0].ReplyFields[0].Path = "~req."
	cfg.dnsAgentCfg.RequestProcessors[0].ReplyFields[0].Type = "CHANGED"
	expected = "<DNSAgent> Empty field path  for ~req. at Path"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.dnsAgentCfg.RequestProcessors[0].ReplyFields[0].Path = utils.EmptyString

	expected = "<DNSAgent> MANDATORY_IE_MISSING: [Path] for cgrates at SessionId"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.dnsAgentCfg.RequestProcessors[0].ReplyFields[0].Path = "*empty:*ec"

	cfg.dnsAgentCfg.RequestProcessors[0].ReplyFields[0].Value[0].path = "~req."
	expected = "<DNSAgent> Empty field path  for ~req. at Values of reply_fields"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.dnsAgentCfg.RequestProcessors[0].ReplyFields[0].Value[0].path = "*empty:*ec"

	cfg.dnsAgentCfg.RequestProcessors[0].ReplyFields[0].Filters = []string{"*empty:*ec"}
	expected = "<DNSAgent> inline parse error for string: <*empty:*ec> for [*empty:*ec] at reply_fields"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.dnsAgentCfg.RequestProcessors[0].ReplyFields = []*FCTemplate{}

	cfg.dnsAgentCfg.RequestProcessors[0].Filters = []string{"*empty:*ec"}
	expected = "<DNSAgent> inline parse error for string: <*empty:*ec> for [*empty:*ec] at request_processors"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.dnsAgentCfg.RequestProcessors[0].Filters = []string{"*string:~*req.Account:1001"}

}

func TestConfigSanityHTTPAgent1(t *testing.T) {
	cfg := NewDefaultCGRConfig()

	cfg.httpAgentCfg = HTTPAgentCfgs{
		&HTTPAgentCfg{
			SessionSConns: []string{utils.MetaInternal},
			RequestProcessors: []*RequestProcessor{
				{
					ID:       "cgrates",
					Timezone: "Local",
					RequestFields: []*FCTemplate{
						{Tag: "SessionId", Path: utils.EmptyString, Type: "*variable",
							Value: NewRSRParsersMustCompile("~*req.Session-Id", utils.InfieldSep), Mandatory: true},
					},
					ReplyFields: []*FCTemplate{
						{Tag: "SessionId", Path: utils.EmptyString, Type: "*variable",
							Value: NewRSRParsersMustCompile("~*req.Session-Id", utils.InfieldSep), Mandatory: true},
					},
				},
			},
		},
	}
	expected := "<SessionS> not enabled but requested by <> HTTPAgent Template"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.httpAgentCfg[0].SessionSConns = []string{"test"}
	expected = "<HTTPAgent> template with ID <> has connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.httpAgentCfg[0].SessionSConns = []string{}

	cfg.httpAgentCfg[0].RequestPayload = "test"
	expected = "<HTTPAgent> unsupported request payload test"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.httpAgentCfg[0].RequestPayload = utils.MetaXml

	cfg.httpAgentCfg[0].ReplyPayload = "test"
	expected = "<HTTPAgent> unsupported reply payload test"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.httpAgentCfg[0].ReplyPayload = utils.MetaXml

	expected = "<HTTPAgent> MANDATORY_IE_MISSING: [Path] for cgrates at SessionId"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.httpAgentCfg[0].RequestProcessors[0].RequestFields[0].Type = utils.MetaNone
	expected = "<HTTPAgent> MANDATORY_IE_MISSING: [Path] for cgrates at SessionId"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.httpAgentCfg[0].RequestProcessors[0].RequestFields[0].Path = "~Field1..Field2[0]"
	expected = "<HTTPAgent> Empty field path  for ~Field1..Field2[0] at Path"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.httpAgentCfg[0].RequestProcessors[0].RequestFields[0].Value = RSRParsers{
		&RSRParser{
			Rules: "~*req.Session-Id2",
			path:  "~Field1..Field2[0]",
		},
	}
	cfg.httpAgentCfg[0].RequestProcessors[0].RequestFields[0].Path = utils.EmptyString
	expected = "<HTTPAgent> Empty field path  for ~Field1..Field2[0] at Values of request_fields"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.httpAgentCfg[0].RequestProcessors[0].RequestFields[0].Value = RSRParsers{
		&RSRParser{
			Rules: "~*req.Session-Id2",
			path:  "~Field1.Field2[0]",
		},
	}
	cfg.httpAgentCfg[0].RequestProcessors[0].RequestFields[0].Filters = []string{"*empty:~Field1..Field2[0]:*Test3:*Test4"}
	expected = "<HTTPAgent> Empty field path  for <*empty:~Field1..Field2[0]:*Test3:*Test4> for [*empty:~Field1..Field2[0]:*Test3:*Test4] at request_fields"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.httpAgentCfg[0].RequestProcessors[0].RequestFields[0].Filters = nil

	cfg.httpAgentCfg[0].RequestProcessors[0].ReplyFields[0].Path = "~Field1..Field2[0]"
	expected = "<HTTPAgent> Empty field path  for ~Field1..Field2[0] at Path"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.httpAgentCfg[0].RequestProcessors[0].ReplyFields[0].Value = RSRParsers{
		&RSRParser{
			Rules: "~*req.Session-Id2",
			path:  "~Field1..Field2[0]",
		},
	}
	cfg.httpAgentCfg[0].RequestProcessors[0].ReplyFields[0].Path = utils.EmptyString
	cfg.httpAgentCfg[0].RequestProcessors[0].ReplyFields[0].Type = utils.MetaNone
	expected = "<HTTPAgent> Empty field path  for ~Field1..Field2[0] at Values of reply_fields"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.httpAgentCfg[0].RequestProcessors[0].ReplyFields[0].Value = RSRParsers{
		&RSRParser{
			Rules: "~*req.Session-Id2",
			path:  "~Field1.Field2[0]",
		},
	}
	cfg.httpAgentCfg[0].RequestProcessors[0].ReplyFields[0].Filters = []string{"*empty:~Field1..Field2[0]:*Test3:*Test4"}
	expected = "<HTTPAgent> Empty field path  for <*empty:~Field1..Field2[0]:*Test3:*Test4> for [*empty:~Field1..Field2[0]:*Test3:*Test4] at reply_fields"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.httpAgentCfg[0].RequestProcessors[0].ReplyFields[0].Filters = nil

	cfg.httpAgentCfg[0].RequestProcessors[0].Filters = []string{"*empty:~Field1..Field2[0]:*Test3:*Test4"}
	expected = "<HTTPAgent> Empty field path  for <*empty:~Field1..Field2[0]:*Test3:*Test4> for [*empty:~Field1..Field2[0]:*Test3:*Test4] at request_processors"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
}

func TestConfigSanitySipAgent(t *testing.T) {
	cfg := NewDefaultCGRConfig()

	cfg.sipAgentCfg = &SIPAgentCfg{
		Enabled: true,
		RequestProcessors: []*RequestProcessor{
			{
				ID:       "cgrates",
				Timezone: "Local",
				RequestFields: []*FCTemplate{
					{Tag: "SessionId", Path: utils.EmptyString, Type: "*variable",
						Value: NewRSRParsersMustCompile("~*req.Session-Id", utils.InfieldSep), Mandatory: true},
				},
				ReplyFields: []*FCTemplate{
					{Tag: "SessionId", Path: utils.EmptyString, Type: "*variable",
						Value: NewRSRParsersMustCompile("~*req.Session-Id", utils.InfieldSep), Mandatory: true},
				},
			},
		},
	}

	expected := "<SIPAgent> no SessionS connections defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	expected = "<SessionS> not enabled but requested by <SIPAgent> component"
	cfg.sipAgentCfg.SessionSConns = []string{utils.MetaInternal}
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	expected = "<SIPAgent> connection with id: <test> not defined"
	cfg.sipAgentCfg.SessionSConns = []string{"test"}
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	//Request fields
	cfg.rpcConns["test"] = nil
	expected = "<SIPAgent> MANDATORY_IE_MISSING: [Path] for cgrates at SessionId"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.sipAgentCfg.RequestProcessors[0].RequestFields[0].Type = utils.MetaNone
	expected = "<SIPAgent> MANDATORY_IE_MISSING: [Path] for cgrates at SessionId"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.sipAgentCfg.RequestProcessors[0].RequestFields[0].Path = "~Field1..Field2[0]"
	expected = "<SIPAgent> Empty field path  for ~Field1..Field2[0] at Path"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.sipAgentCfg.RequestProcessors[0].RequestFields[0].Value = RSRParsers{
		&RSRParser{
			Rules: "~*req.Session-Id2",
			path:  "~Field1..Field2[0]",
		},
	}
	cfg.sipAgentCfg.RequestProcessors[0].RequestFields[0].Path = utils.EmptyString
	expected = "<SIPAgent> Empty field path  for ~Field1..Field2[0] at Values of request_fields"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.sipAgentCfg.RequestProcessors[0].RequestFields[0].Value = RSRParsers{
		&RSRParser{
			Rules: "~*req.Session-Id2",
			path:  "~Field1.Field2[0]",
		},
	}
	cfg.sipAgentCfg.RequestProcessors[0].RequestFields[0].Filters = []string{"*empty:~Field1..Field2[0]:*Test3:*Test4"}
	expected = "<SIPAgent> Empty field path  for <*empty:~Field1..Field2[0]:*Test3:*Test4> for [*empty:~Field1..Field2[0]:*Test3:*Test4] at request_fields"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.sipAgentCfg.RequestProcessors[0].RequestFields[0].Filters = nil

	//Reply fields
	cfg.sipAgentCfg.RequestProcessors[0].ReplyFields[0].Path = "~Field1..Field2[0]"
	expected = "<SIPAgent> Empty field path  for ~Field1..Field2[0] at Path"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.sipAgentCfg.RequestProcessors[0].ReplyFields[0].Value = RSRParsers{
		&RSRParser{
			Rules: "~*req.Session-Id2",
			path:  "~Field1..Field2[0]",
		},
	}
	cfg.sipAgentCfg.RequestProcessors[0].ReplyFields[0].Path = utils.EmptyString
	cfg.sipAgentCfg.RequestProcessors[0].ReplyFields[0].Type = utils.MetaNone
	expected = "<SIPAgent> Empty field path  for ~Field1..Field2[0] at Values of reply_fields"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.sipAgentCfg.RequestProcessors[0].ReplyFields[0].Value = RSRParsers{
		&RSRParser{
			Rules: "~*req.Session-Id2",
			path:  "~Field1.Field2[0]",
		},
	}
	cfg.sipAgentCfg.RequestProcessors[0].ReplyFields[0].Filters = []string{"*empty:~Field1..Field2[0]:*Test3:*Test4"}
	expected = "<SIPAgent> Empty field path  for <*empty:~Field1..Field2[0]:*Test3:*Test4> for [*empty:~Field1..Field2[0]:*Test3:*Test4] at reply_fields"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.sipAgentCfg.RequestProcessors[0].ReplyFields[0].Filters = nil

	cfg.sipAgentCfg.RequestProcessors[0].Filters = []string{"*empty:~Field1..Field2[0]:*Test3:*Test4"}
	expected = "<SIPAgent> Empty field path  for <*empty:~Field1..Field2[0]:*Test3:*Test4> for [*empty:~Field1..Field2[0]:*Test3:*Test4] at request_processors"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
}

func TestConfigSanityAttributesCfg(t *testing.T) {
	cfg := NewDefaultCGRConfig()

	cfg.attributeSCfg = &AttributeSCfg{
		Enabled: true,
		Opts: &AttributesOpts{
			ProcessRuns: 0,
		},
	}
	expected := "<AttributeS> processRuns needs to be bigger than 0"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
}

func TestConfigSanityChargerS(t *testing.T) {
	cfg := NewDefaultCGRConfig()

	cfg.chargerSCfg = &ChargerSCfg{
		Enabled:         true,
		AttributeSConns: []string{utils.MetaInternal},
	}
	expected := "<AttributeS> not enabled but requested by <ChargerS> component"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.chargerSCfg.AttributeSConns = []string{"test"}
	expected = "<ChargerS> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
}

func TestConfigSanityResourceLimiter(t *testing.T) {
	cfg = NewDefaultCGRConfig()
	cfg.resourceSCfg = &ResourceSConfig{
		Enabled:         true,
		ThresholdSConns: []string{utils.MetaInternal},
	}
	expected := "<ThresholdS> not enabled but requested by <ResourceS> component"
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
	cfg = NewDefaultCGRConfig()
	cfg.statsCfg = &StatSCfg{
		Enabled:         true,
		ThresholdSConns: []string{utils.MetaInternal},
	}
	expected := "<ThresholdS> not enabled but requested by <Stats> component"
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
	cfg = NewDefaultCGRConfig()
	cfg.routeSCfg.Enabled = true

	cfg.routeSCfg.ResourceSConns = []string{utils.MetaInternal}
	expected := "<ResourceS> not enabled but requested by <RouteS> component"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.routeSCfg.ResourceSConns = []string{"test"}
	expected = "<RouteS> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.routeSCfg.ResourceSConns = []string{}

	cfg.routeSCfg.StatSConns = []string{utils.MetaInternal}
	expected = "<StatS> not enabled but requested by <RouteS> component"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.routeSCfg.StatSConns = []string{"test"}
	expected = "<RouteS> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.routeSCfg.StatSConns = []string{}

	cfg.routeSCfg.AttributeSConns = []string{utils.MetaInternal}
	expected = "<AttributeS> not enabled but requested by <RouteS> component"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.routeSCfg.AttributeSConns = []string{"test"}
	expected = "<RouteS> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.routeSCfg.AttributeSConns = []string{}
}

func TestConfigSanityEventReader(t *testing.T) {
	cfg = NewDefaultCGRConfig()
	cfg.ersCfg = &ERsCfg{
		Enabled:       true,
		SessionSConns: []string{"unexistedConn"},
	}
	expected := "<ERs> connection with id: <unexistedConn> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.ersCfg.SessionSConns = []string{utils.MetaInternal}
	expected = "<SessionS> not enabled but requested by <ERs> component"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.sessionSCfg.Enabled = true
	cfg.ersCfg.Readers = []*EventReaderCfg{{
		ID:   "test",
		Type: "wrongtype",
	}}
	expected = "<ERs> unsupported data type: wrongtype for reader with ID: test"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.ersCfg.Readers = []*EventReaderCfg{{
		ID:            "test2",
		Type:          utils.MetaFileCSV,
		ProcessedPath: "not/a/path",
		Opts: map[string]interface{}{
			utils.PartialCacheActionOpt: utils.MetaNone,
		},
	}}
	expected = "<ERs> nonexistent folder: not/a/path for reader with ID: test2"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.ersCfg.Readers = []*EventReaderCfg{{
		ID:            "test3",
		Type:          utils.MetaFileCSV,
		ProcessedPath: "/",
		SourcePath:    "/",
		Opts: map[string]interface{}{
			"csvFieldSeparator":         "",
			utils.PartialCacheActionOpt: utils.MetaNone,
		},
	}}
	expected = "<ERs> empty csvFieldSeparator for reader with ID: test3"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.ersCfg.Readers[0] = &EventReaderCfg{
		ID:       "test4",
		Type:     utils.MetaKafkajsonMap,
		RunDelay: 1,
		Opts: map[string]interface{}{
			"csvFieldSeparator":         ",",
			utils.PartialCacheActionOpt: utils.MetaNone,
		},
	}
	expected = "<ERs> the RunDelay field can not be bigger than zero for reader with ID: test4"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.ersCfg.Readers[0] = &EventReaderCfg{
		ID:            "test5",
		Type:          utils.MetaFileXML,
		RunDelay:      0,
		ProcessedPath: "not/a/path",
		SourcePath:    "not/a/path",
		Opts: map[string]interface{}{
			utils.PartialCacheActionOpt: utils.MetaNone,
		},
	}
	expected = "<ERs> nonexistent folder: not/a/path for reader with ID: test5"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.ersCfg.Readers[0] = &EventReaderCfg{
		ID:            "test5",
		Type:          utils.MetaFileFWV,
		RunDelay:      0,
		ProcessedPath: "not/a/path",
		SourcePath:    "not/a/path",
		Opts: map[string]interface{}{
			utils.PartialCacheActionOpt: utils.MetaNone,
		},
	}
	expected = "<ERs> nonexistent folder: not/a/path for reader with ID: test5"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.ersCfg = &ERsCfg{
		Enabled: true,
		Readers: []*EventReaderCfg{
			{
				Type: utils.MetaKafkajsonMap,
				CacheDumpFields: []*FCTemplate{
					{Tag: "SessionId", Path: utils.EmptyString, Type: "*variable",
						Value: NewRSRParsersMustCompile("~*req.Session-Id", utils.InfieldSep), Mandatory: true},
				},
				Fields: []*FCTemplate{
					{Tag: "SessionId", Path: utils.EmptyString, Type: "*variable",
						Value: NewRSRParsersMustCompile("~*req.Session-Id", utils.InfieldSep), Mandatory: true},
				},
				Opts: map[string]interface{}{
					utils.PartialCacheActionOpt: utils.MetaNone,
				},
			},
		},
	}

	//CacheDumpFields
	expected = "<ERs> MANDATORY_IE_MISSING: [Path] for  at SessionId"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	expected = "<ERs> MANDATORY_IE_MISSING: [Path] for  at SessionId"
	cfg.ersCfg.Readers[0].CacheDumpFields[0].Type = utils.MetaNone
	if err := cfg.CheckConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.ersCfg.Readers[0].CacheDumpFields[0].Path = "~Field1..Field2[0]"
	expected = "<ERs> Empty field path  for ~Field1..Field2[0] at Path"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.ersCfg.Readers[0].CacheDumpFields[0].Value = RSRParsers{
		&RSRParser{
			Rules: "~*req.Session-Id2",
			path:  "~Field1..Field2[0]",
		},
	}
	cfg.ersCfg.Readers[0].CacheDumpFields[0].Path = "~Field1.Field3[0]"
	cfg.ersCfg.Readers[0].CacheDumpFields[0].Type = utils.MetaUnixTimestamp
	expected = "<ERs> Empty field path  for ~Field1..Field2[0] at Values of cache_dump_fields"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.ersCfg.Readers[0].CacheDumpFields[0].Type = utils.MetaNone
	cfg.ersCfg.Readers[0].CacheDumpFields[0].Value = RSRParsers{
		&RSRParser{
			Rules: "~*req.Session-Id2",
			path:  "~Field1.Field2[0]",
		},
	}
	cfg.ersCfg.Readers[0].CacheDumpFields[0].Filters = []string{"*empty:~Field1..Field2[0]:*Test3:*Test4"}
	expected = "<ERs> Empty field path  for <*empty:~Field1..Field2[0]:*Test3:*Test4> for [*empty:~Field1..Field2[0]:*Test3:*Test4] at cache_dump_fields"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.ersCfg.Readers[0].CacheDumpFields[0].Filters = nil

	//Fields
	cfg.ersCfg.Readers[0].Fields[0].Path = "~Field1..Field2[0]"
	expected = "<ERs> Empty field path  for ~Field1..Field2[0] at Path"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.ersCfg.Readers[0].Fields[0].Value = RSRParsers{
		&RSRParser{
			Rules: "~*req.Session-Id2",
			path:  "~Field1..Field2[0]",
		},
	}
	cfg.ersCfg.Readers[0].Fields[0].Path = "~Field1.Field3[0]"
	cfg.ersCfg.Readers[0].Fields[0].Type = utils.MetaUnixTimestamp
	expected = "<ERs> Empty field path  for ~Field1..Field2[0] at Values of fields"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.ersCfg.Readers[0].Fields[0].Type = utils.MetaNone
	cfg.ersCfg.Readers[0].Fields[0].Value = RSRParsers{
		&RSRParser{
			Rules: "~*req.Session-Id2",
			path:  "~Field1.Field2[0]",
		},
	}
	cfg.ersCfg.Readers[0].Fields[0].Filters = []string{"*empty:~Field1..Field2[0]:*Test3:*Test4"}
	expected = "<ERs> Empty field path  for <*empty:~Field1..Field2[0]:*Test3:*Test4> for [*empty:~Field1..Field2[0]:*Test3:*Test4] at fields"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.ersCfg.Readers[0].Fields[0].Filters = nil

	cfg.ersCfg.Readers[0].Filters = []string{"*empty:~Field1..Field2[0]:*Test3:*Test4"}
	expected = "<ERs> Empty field path  for <*empty:~Field1..Field2[0]:*Test3:*Test4> for [*empty:~Field1..Field2[0]:*Test3:*Test4] at readers"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
}

func TestConfigSanityEventExporter(t *testing.T) {
	cfg := NewDefaultCGRConfig()

	cfg.eesCfg = &EEsCfg{
		Enabled:         true,
		AttributeSConns: []string{utils.MetaInternal},
		Exporters: []*EventExporterCfg{
			{
				Fields: []*FCTemplate{
					{Tag: "SessionId", Path: utils.EmptyString, Type: "*variable",
						Value: NewRSRParsersMustCompile("~*req.Session-Id", utils.InfieldSep), Mandatory: true},
				},
			},
		},
	}

	expected := "<AttributeS> not enabled but requested by <EEs> component"
	if err := cfg.CheckConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.eesCfg.AttributeSConns = []string{"test"}
	expected = "<EEs> connection with id: <test> not defined"
	if err := cfg.CheckConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.eesCfg.AttributeSConns = []string{}

	expected = "<EEs> unsupported data type:  for exporter with ID: "
	if err := cfg.CheckConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.eesCfg.Exporters[0].Type = utils.MetaHTTPPost
	expected = "<EEs> MANDATORY_IE_MISSING: [Path] for  at SessionId"
	if err := cfg.CheckConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.eesCfg.Exporters[0].Type = utils.MetaFileCSV
	cfg.eesCfg.Exporters[0].ExportPath = "/"
	cfg.eesCfg.Exporters[0].Opts = map[string]interface{}{utils.CSVFieldSepOpt: ""}
	expected = "<EEs> empty csvFieldSeparator for exporter with ID: "
	if err := cfg.CheckConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.eesCfg.Exporters[0].ExportPath = "randomPath"
	expected = "<EEs> nonexistent folder: randomPath for exporter with ID: "
	if err := cfg.CheckConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.eesCfg.Exporters[0].Type = utils.MetaFileFWV
	expected = "<EEs> nonexistent folder: randomPath for exporter with ID: "
	if err := cfg.CheckConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.eesCfg.Exporters[0].Type = utils.MetaSQL
	expected = "<EEs> empty content fields for exporter with ID: "
	if err := cfg.CheckConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.eesCfg.Exporters[0].Type = utils.MetaHTTPPost
	cfg.eesCfg.Exporters[0].Fields[0].Path = "~Field1..Field2[0]"
	expected = "<EEs> Empty field path  for ~Field1..Field2[0] at Path"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.eesCfg.Exporters[0].Fields[0].Value = RSRParsers{
		&RSRParser{
			Rules: "~*req.Session-Id2",
			path:  "~Field1..Field2[0]",
		},
	}
	cfg.eesCfg.Exporters[0].Fields[0].Path = "~Field1.Field3[0]"
	cfg.eesCfg.Exporters[0].Fields[0].Type = utils.MetaUnixTimestamp
	expected = "<EEs> Empty field path  for ~Field1..Field2[0] at Values of fields"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.eesCfg.Exporters[0].Fields[0].Type = utils.MetaNone
	cfg.eesCfg.Exporters[0].Fields[0].Value = RSRParsers{
		&RSRParser{
			Rules: "~*req.Session-Id2",
			path:  "~Field1.Field2[0]",
		},
	}
	cfg.eesCfg.Exporters[0].Fields[0].Filters = []string{"*empty:~Field1..Field2[0]:*Test3:*Test4"}
	expected = "<EEs> Empty field path  for <*empty:~Field1..Field2[0]:*Test3:*Test4> for [*empty:~Field1..Field2[0]:*Test3:*Test4] at fields"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.eesCfg.Exporters[0].Fields[0].Filters = nil

	cfg.eesCfg.Exporters[0].Filters = []string{"*empty:~Field1..Field2[0]:*Test3:*Test4"}
	expected = "<EEs> Empty field path  for <*empty:~Field1..Field2[0]:*Test3:*Test4> for [*empty:~Field1..Field2[0]:*Test3:*Test4] at exporters"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
}

func TestConfigSanityCache(t *testing.T) {
	cfg := NewDefaultCGRConfig()

	cfg.cacheCfg = &CacheCfg{
		ReplicationConns: []string{"test"},
	}
	expected := "<CacheS> connection with id: <test> not defined"
	if err := cfg.CheckConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.cacheCfg.ReplicationConns = []string{utils.MetaLocalHost}
	expected = "<CacheS> unsuported transport <*json> for connection with ID: <*localhost>"
	if err := cfg.CheckConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
}

func TestConfigSanityRegistrarCRPC(t *testing.T) {
	cfg := NewDefaultCGRConfig()

	cfg.registrarCCfg = &RegistrarCCfgs{
		RPC: &RegistrarCCfg{
			RegistrarSConns: []string{utils.MetaLocalHost},
			Hosts: map[string][]*RemoteHost{
				"hosts": {},
			},
		},
		Dispatchers: &RegistrarCCfg{},
	}

	expected := "<RegistrarC> the register imterval needs to be bigger than 0"
	if err := cfg.CheckConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.registrarCCfg.RPC.Hosts = nil
	expected = "<RegistrarC> missing RPC host IDs"
	if err := cfg.CheckConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.registrarCCfg.RPC.RefreshInterval = 2
	cfg.registrarCCfg.RPC.Hosts = map[string][]*RemoteHost{
		"hosts": {
			{
				ID: "randomID",
			},
		},
	}
	expected = "<RegistrarC> unsupported transport <> for host <hosts:randomID>"
	if err := cfg.CheckConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.registrarCCfg.RPC.Hosts["hosts"][0].Transport = utils.MetaJSON

	cfg.registrarCCfg.RPC.RegistrarSConns = []string{utils.MetaInternal}
	expected = "<RegistrarC> internal connection IDs are not supported"
	if err := cfg.CheckConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.registrarCCfg.RPC.RegistrarSConns = []string{utils.MetaLocalHost}
	expected = "<RegistrarC> connection with id: <*localhost> unsupported transport <*json>"
	if err := cfg.CheckConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.registrarCCfg.RPC.RegistrarSConns = []string{"*conn1"}
	expected = "<RegistrarC> connection with id: <*conn1> not defined"
	if err := cfg.CheckConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.rpcConns = RPCConns{
		utils.MetaLocalHost: {},
		"*conn1":            {},
	}
	expected = "<RegistrarC> connection with id: <*conn1> needs to have only one host"
	if err := cfg.CheckConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
}
func TestConfigSanityRegistrarCDispatcher(t *testing.T) {
	cfg := NewDefaultCGRConfig()

	cfg.registrarCCfg = &RegistrarCCfgs{
		Dispatchers: &RegistrarCCfg{
			RegistrarSConns: []string{utils.MetaLocalHost},
			Hosts: map[string][]*RemoteHost{
				"hosts": {},
			},
		},
		RPC: &RegistrarCCfg{},
	}

	expected := "<RegistrarC> the register imterval needs to be bigger than 0"
	if err := cfg.CheckConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.registrarCCfg.Dispatchers.Hosts = nil
	cfg.registrarCCfg.Dispatchers.RefreshInterval = 2
	cfg.registrarCCfg.Dispatchers.Hosts = map[string][]*RemoteHost{
		"hosts": {
			{
				ID: "randomID",
			},
		},
	}
	expected = "<RegistrarC> unsupported transport <> for host <hosts:randomID>"
	if err := cfg.CheckConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.registrarCCfg.Dispatchers.Hosts["hosts"][0].Transport = utils.MetaJSON

	cfg.registrarCCfg.Dispatchers.RegistrarSConns = []string{utils.MetaInternal}
	expected = "<RegistrarC> internal connection IDs are not supported"
	if err := cfg.CheckConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.registrarCCfg.Dispatchers.RegistrarSConns = []string{utils.MetaLocalHost}
	expected = "<RegistrarC> connection with id: <*localhost> unsupported transport <*json>"
	if err := cfg.CheckConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.registrarCCfg.Dispatchers.RegistrarSConns = []string{"*conn1"}
	expected = "<RegistrarC> connection with id: <*conn1> not defined"
	if err := cfg.CheckConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.rpcConns = RPCConns{
		utils.MetaLocalHost: {},
		"*conn1":            {},
	}
	expected = "<RegistrarC> connection with id: <*conn1> needs to have only one host"
	if err := cfg.CheckConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
}

func TestConfigSanityStorDB(t *testing.T) {
	cfg = NewDefaultCGRConfig()
	cfg.storDbCfg = &StorDbCfg{
		Type: utils.Postgres,
		Opts: map[string]interface{}{
			utils.SSLModeCfg: "wrongSSLMode",
		},
	}
	expected := "<stor_db> unsuported sslmode for storDB"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
}

func TestConfigSanityAnalyzer(t *testing.T) {
	cfg := NewDefaultCGRConfig()

	cfg.analyzerSCfg = &AnalyzerSCfg{
		Enabled: true,
		DBPath:  "/",
	}

	expected := "<AnalyzerS> unsuported index type: \"\""
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.analyzerSCfg.DBPath = "/"

	cfg.analyzerSCfg.IndexType = utils.MetaScorch
	expected = "<AnalyzerS> the TTL needs to be bigger than 0"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.analyzerSCfg.TTL = 1
	expected = "<AnalyzerS> the CleanupInterval needs to be bigger than 0"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
}

func TestConfigSanityDataDB(t *testing.T) {
	cfg = NewDefaultCGRConfig()
	cfg.dataDbCfg.Type = utils.Internal

	cfg.cacheCfg = &CacheCfg{
		Partitions: map[string]*CacheParamCfg{},
	}
	if err := cfg.checkConfigSanity(); err != nil {
		t.Error(err)
	}
	cfg.cacheCfg = &CacheCfg{
		Partitions: map[string]*CacheParamCfg{
			utils.CacheAccounts: {
				Limit: 1,
			},
		},
	}
	expected := "<CacheS> *accounts needs to be 0 when DataBD is *internal, received : 1"
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
		"test1": {
			Remote: true,
		},
	}
	expected = "remote connections required by: <test1>"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}

	cfg.dataDbCfg.Items = map[string]*ItemOpt{
		"test2": {
			Remote:    false,
			Replicate: true,
		},
	}
	expected = "replicate connections required by: <test2>"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	//RpcConns
	cfg.dataDbCfg.RplConns = []string{"test1"}
	expected = "<data_db> connection with id: <test1> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.dataDbCfg.RplConns = []string{utils.MetaInternal}
	cfg.rpcConns[utils.MetaInternal].Conns = []*RemoteHost{
		{
			Transport: utils.MetaNone,
		},
	}
	expected = "<data_db> unsuported transport <*none> for connection with ID: <*internal>"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.dataDbCfg.RplConns = []string{}
	cfg.dataDbCfg.Items = map[string]*ItemOpt{}
	//RmtConns
	cfg.dataDbCfg.RmtConns = []string{"test2"}
	expected = "<data_db> connection with id: <test2> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {

		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.dataDbCfg.RmtConns = []string{utils.MetaInternal}
	cfg.rpcConns[utils.MetaInternal].Conns = []*RemoteHost{
		{
			Transport: utils.MetaNone,
		},
	}
	expected = "<data_db> unsuported transport <*none> for connection with ID: <*internal>"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
}

func TestConfigSanityAPIer(t *testing.T) {
	cfg = NewDefaultCGRConfig()
	cfg.admS.AttributeSConns = []string{utils.MetaInternal}

	if err := cfg.checkConfigSanity(); err == nil || err.Error() != "<AttributeS> not enabled but requested by <AdminS> component" {
		t.Error(err)
	}
	cfg.admS.AttributeSConns = []string{"test"}
	expected := "<AdminS> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.admS.AttributeSConns = []string{utils.MetaInternal}
	cfg.attributeSCfg.Enabled = true
	cfg.admS.ActionSConns = []string{utils.MetaInternal}

	if err := cfg.checkConfigSanity(); err == nil || err.Error() != "<SchedulerS> not enabled but requested by <AdminS> component" {
		t.Error(err)
	}
	cfg.admS.ActionSConns = []string{"test"}
	expected = "<AdminS> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
}

func TestConfigSanityDispatcher(t *testing.T) {
	cfg = NewDefaultCGRConfig()
	cfg.dispatcherSCfg = &DispatcherSCfg{
		Enabled:         true,
		AttributeSConns: []string{utils.MetaInternal},
	}
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != "<AttributeS> not enabled but requested by <DispatcherS> component" {
		t.Error(err)
	}
	cfg.dispatcherSCfg.AttributeSConns = []string{"test"}
	expected := "<DispatcherS> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
}

func TestConfigSanityCacheS(t *testing.T) {
	cfg = NewDefaultCGRConfig()

	cfg.cacheCfg.Partitions = map[string]*CacheParamCfg{"wrong_partition_name": {Limit: 10}}
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != "<CacheS> partition <wrong_partition_name> not defined" {
		t.Error(err)
	}

	cfg.cacheCfg.Partitions = map[string]*CacheParamCfg{utils.CacheLoadIDs: {Limit: 9}}
	if err := cfg.checkConfigSanity(); err != nil {
		t.Error(err)
	}
}

func TestConfigSanityFilterS(t *testing.T) {
	cfg = NewDefaultCGRConfig()
	cfg.filterSCfg.StatSConns = []string{utils.MetaInternal}

	if err := cfg.checkConfigSanity(); err == nil || err.Error() != "<Stats> not enabled but requested by <FilterS> component" {
		t.Error(err)
	}
	cfg.filterSCfg.StatSConns = []string{"test"}
	expected := "<FilterS> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.filterSCfg.StatSConns = []string{}

	cfg.filterSCfg.ResourceSConns = []string{utils.MetaInternal}

	if err := cfg.checkConfigSanity(); err == nil || err.Error() != "<ResourceS> not enabled but requested by <FilterS> component" {
		t.Error(err)
	}
	cfg.filterSCfg.ResourceSConns = []string{"test"}
	expected = "<FilterS> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.filterSCfg.ResourceSConns = []string{}

	cfg.filterSCfg.AccountSConns = []string{utils.MetaInternal}

	if err := cfg.checkConfigSanity(); err == nil || err.Error() != "<AccountS> not enabled but requested by <FilterS> component" {
		t.Error(err)
	}
	cfg.filterSCfg.AccountSConns = []string{"test"}
	expected = "<FilterS> connection with id: <test> not defined"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
}

func TestConfigSanityErs(t *testing.T) {
	cfg = NewDefaultCGRConfig()
	cfg.ersCfg.Enabled = true

	cfg.ersCfg.SessionSConns = []string{}
	cfg.ersCfg.Readers = []*EventReaderCfg{
		{
			Type: utils.MetaNone,
			ID:   "rdrID",
			Opts: map[string]interface{}{
				utils.PartialCacheActionOpt: utils.MetaPostCDR,
			},
		},
	}

	cfg.ersCfg.Readers[0].Opts = map[string]interface{}{
		utils.PartialOrderFieldOpt:  utils.EmptyString,
		utils.PartialCacheActionOpt: "invalid",
	}
	expected := "<ERs> wrong partial expiry action for reader with ID: rdrID"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("expected: <%v>,\n received: <%v>", expected, err)
	}

	cfg.ersCfg.Readers[0].Opts = map[string]interface{}{
		utils.PartialOrderFieldOpt:  utils.EmptyString,
		utils.PartialCacheActionOpt: utils.MetaPostCDR,
	}
	expected = "<ERs> empty partialOrderField for reader with ID: rdrID"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("expected: <%v>,\n received: <%v>", expected, err)
	}

	cfg.ersCfg.Readers[0].ProcessedPath = "/tmp"
	cfg.ersCfg.Readers[0].Opts = map[string]interface{}{
		utils.PartialCacheActionOpt:      utils.MetaDumpToFile,
		utils.PartialCSVFieldSepartorOpt: utils.EmptyString,
	}
	expected = "<ERs> empty partialcsvFieldSeparator for reader with ID: rdrID"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("expected: <%v>,\n received: <%v>", expected, err)
	}

	cfg.ersCfg.Readers[0].ProcessedPath = "/tmp"
	cfg.ersCfg.Readers[0].Opts = map[string]interface{}{
		utils.PartialCacheActionOpt:      utils.MetaDumpToFile,
		utils.PartialCSVFieldSepartorOpt: utils.FieldsSep,
	}
	expected = "<ERs> empty cache_dump_fields for reader with ID: rdrID"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("expected: <%v>,\n received: <%v>", expected, err)
	}

	cfg.ersCfg.Readers[0].Opts = map[string]interface{}{
		utils.PartialOrderFieldOpt:  "non_empty",
		utils.PartialCacheActionOpt: utils.MetaDumpToFile,
		utils.PartialPathOpt:        "path",
	}
	expected = "<ERs> nonexistent partial folder: path for reader with ID: rdrID"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("expected: <%v>,\n received: <%v>", expected, err)
	}

	cfg.ersCfg.Readers[0].Opts = map[string]interface{}{
		utils.PartialCSVFieldSepartorOpt: utils.EmptyString,
	}

	cfg.ersCfg.Readers[0].Type = utils.MetaFileCSV
	cfg.ersCfg.Readers[0].ProcessedPath = utils.EmptyString
	cfg.ersCfg.Readers[0].SourcePath = "/tmp"
	cfg.ersCfg.Readers[0].ID = utils.MetaDefault
	cfg.ersCfg.Readers[0].Opts = map[string]interface{}{
		utils.CSVRowLengthOpt:       "can't convert",
		utils.CSVLazyQuotes:         "can't convert",
		utils.PartialCacheActionOpt: utils.MetaNone,
	}
	expected = `<ERs> error when converting csvRowLength: <strconv.ParseInt: parsing "can't convert": invalid syntax> for reader with ID: *default`
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("expected: <%v>,\n received: <%v>", expected, err)
	}

	cfg.ersCfg.Readers[0].Opts[utils.CSVRowLengthOpt] = "2"
	expected = `<ERs> error when converting csvLazyQuotes: <strconv.ParseBool: parsing "can't convert": invalid syntax> for reader with ID: *default`
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("expected: <%v>,\n received: <%v>", expected, err)
	}
}
