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
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// CheckConfigSanity is used in cgr-engine
func (cfg *CGRConfig) CheckConfigSanity() error {
	return cfg.checkConfigSanity()
}

func (cfg *CGRConfig) checkConfigSanity() error {
	// Rater checks
	if cfg.ralsCfg.Enabled {
		for _, connID := range cfg.ralsCfg.StatSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.statsCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.StatService, utils.RALService)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.RALService, connID)
			}
		}
		for _, connID := range cfg.ralsCfg.ThresholdSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.thresholdSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.ThresholdS, utils.RALService)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.RALService, connID)
			}
		}
	}
	// CDRServer checks
	if cfg.cdrsCfg.Enabled {
		for _, connID := range cfg.cdrsCfg.ChargerSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.chargerSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.ChargerS, utils.CDRs)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.CDRs, connID)
			}
		}
		for _, connID := range cfg.cdrsCfg.RaterConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.ralsCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.RALService, utils.CDRs)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.CDRs, connID)
			}
		}
		for _, connID := range cfg.cdrsCfg.AttributeSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.attributeSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.AttributeS, utils.CDRs)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.CDRs, connID)
			}
		}
		for _, connID := range cfg.cdrsCfg.StatSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.statsCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.StatService, utils.CDRs)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.CDRs, connID)
			}
		}
		for _, connID := range cfg.cdrsCfg.ThresholdSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.thresholdSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.ThresholdS, utils.CDRs)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.CDRs, connID)
			}
		}
		for _, expID := range cfg.cdrsCfg.OnlineCDRExports {
			has := false
			for _, ee := range cfg.eesCfg.Exporters {
				if ee.ID == expID {
					has = true
					break
				}
			}
			if !has {
				return fmt.Errorf("<%s> cannot find exporter with ID: <%s>", utils.CDRs, expID)
			}
		}
		for _, connID := range cfg.cdrsCfg.EEsConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.eesCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.EEs, utils.CDRs)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.CDRs, connID)
			}
		}
	}
	// Loaders sanity checks
	for _, ldrSCfg := range cfg.loaderCfg {
		if !ldrSCfg.Enabled {
			continue
		}
		if _, err := os.Stat(ldrSCfg.TpInDir); err != nil && os.IsNotExist(err) { // if loader is enabled tpInDir must exist
			return fmt.Errorf("<%s> nonexistent folder: %s", utils.LoaderS, ldrSCfg.TpInDir)
		}
		if ldrSCfg.TpOutDir != utils.EmptyString { // tpOutDir support empty string for no moving files after process
			if _, err := os.Stat(ldrSCfg.TpOutDir); err != nil && os.IsNotExist(err) {
				return fmt.Errorf("<%s> nonexistent folder: %s", utils.LoaderS, ldrSCfg.TpOutDir)
			}
		}
		if ldrSCfg.LockFilePath != utils.EmptyString { // tpOutDir support empty string for no moving files after process
			pathL := ldrSCfg.GetLockFilePath()
			if _, err := os.Stat(path.Dir(pathL)); err != nil && os.IsNotExist(err) {
				return fmt.Errorf("<%s> nonexistent folder: %s", utils.LoaderS, pathL)
			}
		}
		for _, data := range ldrSCfg.Data {
			if !posibleLoaderTypes.Has(data.Type) {
				return fmt.Errorf("<%s> unsupported data type %s", utils.LoaderS, data.Type)
			}

			for _, field := range data.Fields {
				if field.Type != utils.MetaComposed && field.Type != utils.MetaString && field.Type != utils.MetaVariable {
					return fmt.Errorf("<%s> invalid field type %s for %s at %s", utils.LoaderS, field.Type, data.Type, field.Tag)
				}
				if field.Path == utils.EmptyString {
					return fmt.Errorf("<%s> %s for %s at %s", utils.LoaderS, utils.NewErrMandatoryIeMissing(utils.Path), data.Type, field.Tag)
				}
				if err := utils.IsPathValidForExporters(field.Path); err != nil {
					return fmt.Errorf("<%s> %s for %s at %s", utils.LoaderS, err, field.Path, utils.Path)
				}
				for _, val := range field.Value {
					if err := utils.IsPathValidForExporters(val.path); err != nil {
						return fmt.Errorf("<%s> %s for %s at %s", utils.LoaderS, err, val.path, utils.Values)
					}
				}
				if err := utils.CheckInLineFilter(field.Filters); err != nil {
					return fmt.Errorf("<%s> %s for %s at %s", utils.LoaderS, err, field.Filters, utils.Filters)
				}
			}
		}
	}
	// SessionS checks
	if cfg.sessionSCfg.Enabled {
		if cfg.sessionSCfg.TerminateAttempts < 1 {
			return fmt.Errorf("<%s> 'terminate_attempts' should be at least 1", utils.SessionS)
		}
		for _, connID := range cfg.sessionSCfg.ChargerSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.chargerSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.ChargerS, utils.SessionS)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.SessionS, connID)
			}
		}
		for _, connID := range cfg.sessionSCfg.RALsConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.ralsCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.RALService, utils.SessionS)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.SessionS, connID)
			}
		}
		for _, connID := range cfg.sessionSCfg.ResSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.resourceSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.ResourceS, utils.SessionS)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.SessionS, connID)
			}
		}
		for _, connID := range cfg.sessionSCfg.ThreshSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.thresholdSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.ThresholdS, utils.SessionS)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.SessionS, connID)
			}
		}
		for _, connID := range cfg.sessionSCfg.StatSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.statsCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.StatService, utils.SessionS)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.SessionS, connID)
			}
		}
		for _, connID := range cfg.sessionSCfg.RouteSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.routeSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.RouteS, utils.SessionS)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.SessionS, connID)
			}
		}
		for _, connID := range cfg.sessionSCfg.AttrSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.attributeSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.AttributeS, utils.SessionS)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.SessionS, connID)
			}
		}
		for _, connID := range cfg.sessionSCfg.CDRsConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.cdrsCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.CDRs, utils.SessionS)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.SessionS, connID)
			}
		}
		for _, connID := range cfg.sessionSCfg.ReplicationConns {
			if _, has := cfg.rpcConns[connID]; !has {
				return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.SessionS, connID)
			}
		}
		if cfg.cacheCfg.Partitions[utils.CacheClosedSessions].Limit == 0 {
			return fmt.Errorf("<%s> %s needs to be != 0, received: %d", utils.CacheS, utils.CacheClosedSessions,
				cfg.cacheCfg.Partitions[utils.CacheClosedSessions].Limit)
		}
		for alfld := range cfg.sessionSCfg.AlterableFields {
			if utils.ProtectedSFlds.Has(alfld) {
				return fmt.Errorf("<%s> the following protected field can't be altered by session: <%s>", utils.SessionS, alfld)
			}
		}
	}

	// FreeSWITCHAgent checks
	if cfg.fsAgentCfg.Enabled {
		if len(cfg.fsAgentCfg.SessionSConns) == 0 {
			return fmt.Errorf("<%s> no %s connections defined",
				utils.FreeSWITCHAgent, utils.SessionS)
		}
		for _, connID := range cfg.fsAgentCfg.SessionSConns {
			isInternal := strings.HasPrefix(connID, utils.MetaInternal) || strings.HasPrefix(connID, rpcclient.BiRPCInternal)
			if isInternal && !cfg.sessionSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.SessionS, utils.FreeSWITCHAgent)
			}
			if _, has := cfg.rpcConns[connID]; !has && !isInternal {
				return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.FreeSWITCHAgent, connID)
			}
		}
	}
	// KamailioAgent checks
	if cfg.kamAgentCfg.Enabled {
		if len(cfg.kamAgentCfg.SessionSConns) == 0 {
			return fmt.Errorf("<%s> no %s connections defined",
				utils.KamailioAgent, utils.SessionS)
		}
		for _, connID := range cfg.kamAgentCfg.SessionSConns {
			isInternal := strings.HasPrefix(connID, utils.MetaInternal) || strings.HasPrefix(connID, rpcclient.BiRPCInternal)
			if isInternal && !cfg.sessionSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.SessionS, utils.KamailioAgent)
			}
			if _, has := cfg.rpcConns[connID]; !has && !isInternal {
				return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.KamailioAgent, connID)
			}
		}
	}
	// AsteriskAgent checks
	if cfg.asteriskAgentCfg.Enabled {
		if len(cfg.asteriskAgentCfg.SessionSConns) == 0 {
			return fmt.Errorf("<%s> no %s connections defined",
				utils.AsteriskAgent, utils.SessionS)
		}
		for _, connID := range cfg.asteriskAgentCfg.SessionSConns {
			isInternal := strings.HasPrefix(connID, utils.MetaInternal) || strings.HasPrefix(connID, rpcclient.BiRPCInternal)
			if isInternal && !cfg.sessionSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.SessionS, utils.AsteriskAgent)
			}
			if _, has := cfg.rpcConns[connID]; !has && !isInternal {
				return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.AsteriskAgent, connID)
			}
		}
	}
	// DAgent checks
	if cfg.diameterAgentCfg.Enabled {
		if len(cfg.diameterAgentCfg.SessionSConns) == 0 {
			return fmt.Errorf("<%s> no %s connections defined",
				utils.DiameterAgent, utils.SessionS)
		}
		for _, connID := range cfg.diameterAgentCfg.SessionSConns {
			isInternal := strings.HasPrefix(connID, utils.MetaInternal) || strings.HasPrefix(connID, rpcclient.BiRPCInternal)
			if isInternal && !cfg.sessionSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.SessionS, utils.DiameterAgent)
			}
			if _, has := cfg.rpcConns[connID]; !has && !isInternal {
				return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.DiameterAgent, connID)
			}
		}
		for prf, tmp := range cfg.templates {
			for _, field := range tmp {
				if field.Type != utils.MetaNone && field.Path == utils.EmptyString {
					return fmt.Errorf("<%s> %s for template %s at %s", utils.DiameterAgent, utils.NewErrMandatoryIeMissing(utils.Path), prf, field.Tag)
				}
				if err := utils.IsPathValidForExporters(field.Path); err != nil {
					return fmt.Errorf("<%s> %s for %s at %s", utils.DiameterAgent, err, field.Path, utils.Path)
				}
				for _, val := range field.Value {
					if err := utils.IsPathValidForExporters(val.path); err != nil {
						return fmt.Errorf("<%s> %s for %s at %s", utils.DiameterAgent, err, val.path, utils.Values)
					}
				}
				if err := utils.CheckInLineFilter(field.Filters); err != nil {
					return fmt.Errorf("<%s> %s for %s at %s", utils.DiameterAgent, err, field.Filters, utils.TemplatesCfg)
				}
			}
		}
		for _, req := range cfg.diameterAgentCfg.RequestProcessors {
			for _, field := range req.RequestFields {
				if field.Type != utils.MetaNone &&
					field.Type != utils.MetaTemplate &&
					field.Path == utils.EmptyString {
					return fmt.Errorf("<%s> %s for %s at %s", utils.DiameterAgent, utils.NewErrMandatoryIeMissing(utils.Path), req.ID, field.Tag)
				}
				if err := utils.IsPathValidForExporters(field.Path); err != nil {
					return fmt.Errorf("<%s> %s for %s at %s", utils.DiameterAgent, err, field.Path, utils.Path)
				}
				for _, val := range field.Value {
					if err := utils.IsPathValidForExporters(val.path); err != nil {
						return fmt.Errorf("<%s> %s for %s at %s of %s", utils.DiameterAgent, err, val.path, utils.Values, utils.RequestFieldsCfg)
					}
				}
				if err := utils.CheckInLineFilter(field.Filters); err != nil {
					return fmt.Errorf("<%s> %s for %s at %s", utils.DiameterAgent, err, field.Filters, utils.RequestFieldsCfg)
				}
			}
			for _, field := range req.ReplyFields {
				if field.Type != utils.MetaNone &&
					field.Type != utils.MetaTemplate &&
					field.Path == utils.EmptyString {
					return fmt.Errorf("<%s> %s for %s at %s", utils.DiameterAgent, utils.NewErrMandatoryIeMissing(utils.Path), req.ID, field.Tag)
				}
				if err := utils.IsPathValidForExporters(field.Path); err != nil {
					return fmt.Errorf("<%s> %s for %s at %s", utils.DiameterAgent, err, field.Path, utils.Path)
				}
				for _, val := range field.Value {
					if err := utils.IsPathValidForExporters(val.path); err != nil {
						return fmt.Errorf("<%s> %s for %s at %s of %s", utils.DiameterAgent, err, val.path, utils.Values, utils.ReplyFieldsCfg)
					}
				}
				if err := utils.CheckInLineFilter(field.Filters); err != nil {
					return fmt.Errorf("<%s> %s for %s at %s", utils.DiameterAgent, err, field.Filters, utils.ReplyFieldsCfg)
				}
			}
			if err := utils.CheckInLineFilter(req.Filters); err != nil {
				return fmt.Errorf("<%s> %s for %s at %s", utils.DiameterAgent, err, req.Filters, utils.RequestProcessorsCfg)
			}
		}
	}
	//Radius Agent
	if cfg.radiusAgentCfg.Enabled {
		if len(cfg.radiusAgentCfg.SessionSConns) == 0 {
			return fmt.Errorf("<%s> no %s connections defined",
				utils.RadiusAgent, utils.SessionS)
		}
		for _, connID := range cfg.radiusAgentCfg.SessionSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.sessionSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.SessionS, utils.RadiusAgent)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.RadiusAgent, connID)
			}
		}
		for _, req := range cfg.radiusAgentCfg.RequestProcessors {
			for _, field := range req.RequestFields {
				if field.Type != utils.MetaNone && field.Path == utils.EmptyString {
					return fmt.Errorf("<%s> %s for %s at %s", utils.RadiusAgent, utils.NewErrMandatoryIeMissing(utils.Path), req.ID, field.Tag)
				}
				if err := utils.IsPathValidForExporters(field.Path); err != nil {
					return fmt.Errorf("<%s> %s for %s at %s", utils.RadiusAgent, err, field.Path, utils.Path)
				}
				for _, val := range field.Value {
					if err := utils.IsPathValidForExporters(val.path); err != nil {
						return fmt.Errorf("<%s> %s for %s at %s of %s", utils.RadiusAgent, err, val.path, utils.Values, utils.RequestFieldsCfg)
					}
				}
				if err := utils.CheckInLineFilter(field.Filters); err != nil {
					return fmt.Errorf("<%s> %s for %s at %s", utils.RadiusAgent, err, field.Filters, utils.RequestFieldsCfg)
				}
			}
			for _, field := range req.ReplyFields {
				if field.Type != utils.MetaNone && field.Path == utils.EmptyString {
					return fmt.Errorf("<%s> %s for %s at %s", utils.RadiusAgent, utils.NewErrMandatoryIeMissing(utils.Path), req.ID, field.Tag)
				}
				if err := utils.IsPathValidForExporters(field.Path); err != nil {
					return fmt.Errorf("<%s> %s for %s at %s", utils.RadiusAgent, err, field.Path, utils.Path)
				}
				for _, val := range field.Value {
					if err := utils.IsPathValidForExporters(val.path); err != nil {
						return fmt.Errorf("<%s> %s for %s at %s of %s", utils.RadiusAgent, err, val.path, utils.Values, utils.ReplyFieldsCfg)
					}
				}
				if err := utils.CheckInLineFilter(field.Filters); err != nil {
					return fmt.Errorf("<%s> %s for %s at %s", utils.RadiusAgent, err, field.Filters, utils.ReplyFieldsCfg)
				}
			}
			if err := utils.CheckInLineFilter(req.Filters); err != nil {
				return fmt.Errorf("<%s> %s for %s at %s", utils.RadiusAgent, err, req.Filters, utils.RequestProcessorsCfg)
			}
		}
	}
	//DNS Agent
	if cfg.dnsAgentCfg.Enabled {
		if len(cfg.dnsAgentCfg.SessionSConns) == 0 {
			return fmt.Errorf("<%s> no %s connections defined",
				utils.DNSAgent, utils.SessionS)
		}
		for _, connID := range cfg.dnsAgentCfg.SessionSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.sessionSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.SessionS, utils.DNSAgent)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.DNSAgent, connID)
			}
		}
		for _, req := range cfg.dnsAgentCfg.RequestProcessors {
			for _, field := range req.RequestFields {
				if field.Type != utils.MetaNone && field.Path == utils.EmptyString {
					return fmt.Errorf("<%s> %s for %s at %s", utils.DNSAgent, utils.NewErrMandatoryIeMissing(utils.Path), req.ID, field.Tag)
				}
				if err := utils.IsPathValidForExporters(field.Path); err != nil {
					return fmt.Errorf("<%s> %s for %s at %s", utils.DNSAgent, err, field.Path, utils.Path)
				}
				for _, val := range field.Value {
					if err := utils.IsPathValidForExporters(val.path); err != nil {
						return fmt.Errorf("<%s> %s for %s at %s of %s", utils.DNSAgent, err, val.path, utils.Values, utils.RequestFieldsCfg)
					}
				}
				if err := utils.CheckInLineFilter(field.Filters); err != nil {
					return fmt.Errorf("<%s> %s for %s at %s", utils.DNSAgent, err, field.Filters, utils.RequestFieldsCfg)
				}
			}
			for _, field := range req.ReplyFields {
				if field.Type != utils.MetaNone && field.Path == utils.EmptyString {
					return fmt.Errorf("<%s> %s for %s at %s", utils.DNSAgent, utils.NewErrMandatoryIeMissing(utils.Path), req.ID, field.Tag)
				}
				if err := utils.IsPathValidForExporters(field.Path); err != nil {
					return fmt.Errorf("<%s> %s for %s at %s", utils.DNSAgent, err, field.Path, utils.Path)
				}
				for _, val := range field.Value {
					if err := utils.IsPathValidForExporters(val.path); err != nil {
						return fmt.Errorf("<%s> %s for %s at %s of %s", utils.DNSAgent, err, val.path, utils.Values, utils.ReplyFieldsCfg)
					}
				}
				if err := utils.CheckInLineFilter(field.Filters); err != nil {
					return fmt.Errorf("<%s> %s for %s at %s", utils.DNSAgent, err, field.Filters, utils.ReplyFieldsCfg)
				}
			}
			if err := utils.CheckInLineFilter(req.Filters); err != nil {
				return fmt.Errorf("<%s> %s for %s at %s", utils.DNSAgent, err, req.Filters, utils.RequestProcessorsCfg)
			}
		}
	}
	// HTTPAgent checks
	for _, httpAgentCfg := range cfg.httpAgentCfg {
		// httpAgent checks
		for _, connID := range httpAgentCfg.SessionSConns {
			isInternal := strings.HasPrefix(connID, utils.MetaInternal) || strings.HasPrefix(connID, rpcclient.BiRPCInternal)
			if isInternal && !cfg.sessionSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> HTTPAgent Template", utils.SessionS, httpAgentCfg.ID)
			}
			if _, has := cfg.rpcConns[connID]; !has && !isInternal {
				return fmt.Errorf("<%s> template with ID <%s> has connection with id: <%s> not defined", utils.HTTPAgent, httpAgentCfg.ID, connID)
			}
		}
		if !utils.SliceHasMember([]string{utils.MetaUrl, utils.MetaXml}, httpAgentCfg.RequestPayload) {
			return fmt.Errorf("<%s> unsupported request payload %s", utils.HTTPAgent, httpAgentCfg.RequestPayload)
		}
		if !utils.SliceHasMember([]string{utils.MetaTextPlain, utils.MetaXml}, httpAgentCfg.ReplyPayload) {
			return fmt.Errorf("<%s> unsupported reply payload %s", utils.HTTPAgent, httpAgentCfg.ReplyPayload)
		}
		for _, req := range httpAgentCfg.RequestProcessors {
			for _, field := range req.RequestFields {
				if field.Type != utils.MetaNone && field.Path == utils.EmptyString {
					return fmt.Errorf("<%s> %s for %s at %s", utils.HTTPAgent, utils.NewErrMandatoryIeMissing(utils.Path), req.ID, field.Tag)
				}
				if err := utils.IsPathValidForExporters(field.Path); err != nil {
					return fmt.Errorf("<%s> %s for %s at %s", utils.HTTPAgent, err, field.Path, utils.Path)
				}
				for _, val := range field.Value {
					if err := utils.IsPathValidForExporters(val.path); err != nil {
						return fmt.Errorf("<%s> %s for %s at %s of %s", utils.HTTPAgent, err, val.path, utils.Values, utils.RequestFieldsCfg)
					}
				}
				if err := utils.CheckInLineFilter(field.Filters); err != nil {
					return fmt.Errorf("<%s> %s for %s at %s", utils.HTTPAgent, err, field.Filters, utils.RequestFieldsCfg)
				}
			}
			for _, field := range req.ReplyFields {
				if field.Type != utils.MetaNone && field.Path == utils.EmptyString {
					return fmt.Errorf("<%s> %s for %s at %s", utils.HTTPAgent, utils.NewErrMandatoryIeMissing(utils.Path), req.ID, field.Tag)
				}
				if err := utils.IsPathValidForExporters(field.Path); err != nil {
					return fmt.Errorf("<%s> %s for %s at %s", utils.HTTPAgent, err, field.Path, utils.Path)
				}
				for _, val := range field.Value {
					if err := utils.IsPathValidForExporters(val.path); err != nil {
						return fmt.Errorf("<%s> %s for %s at %s of %s", utils.HTTPAgent, err, val.path, utils.Values, utils.ReplyFieldsCfg)
					}
				}
				if err := utils.CheckInLineFilter(field.Filters); err != nil {
					return fmt.Errorf("<%s> %s for %s at %s", utils.HTTPAgent, err, field.Filters, utils.ReplyFieldsCfg)
				}
			}
			if err := utils.CheckInLineFilter(req.Filters); err != nil {
				return fmt.Errorf("<%s> %s for %s at %s", utils.HTTPAgent, err, req.Filters, utils.RequestProcessorsCfg)
			}
		}
	}

	//SIP Agent
	if cfg.sipAgentCfg.Enabled {
		if len(cfg.sipAgentCfg.SessionSConns) == 0 {
			return fmt.Errorf("<%s> no %s connections defined",
				utils.SIPAgent, utils.SessionS)
		}
		for _, connID := range cfg.sipAgentCfg.SessionSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.sessionSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.SessionS, utils.SIPAgent)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.SIPAgent, connID)
			}
		}
		for _, req := range cfg.sipAgentCfg.RequestProcessors {
			for _, field := range req.RequestFields {
				if field.Type != utils.MetaNone && field.Path == utils.EmptyString {
					return fmt.Errorf("<%s> %s for %s at %s", utils.SIPAgent, utils.NewErrMandatoryIeMissing(utils.Path), req.ID, field.Tag)
				}
				if err := utils.IsPathValidForExporters(field.Path); err != nil {
					return fmt.Errorf("<%s> %s for %s at %s", utils.SIPAgent, err, field.Path, utils.Path)
				}
				for _, val := range field.Value {
					if err := utils.IsPathValidForExporters(val.path); err != nil {
						return fmt.Errorf("<%s> %s for %s at %s of %s", utils.SIPAgent, err, val.path, utils.Values, utils.RequestFieldsCfg)
					}
				}
				if err := utils.CheckInLineFilter(field.Filters); err != nil {
					return fmt.Errorf("<%s> %s for %s at %s", utils.SIPAgent, err, field.Filters, utils.RequestFieldsCfg)
				}
			}
			for _, field := range req.ReplyFields {
				if field.Type != utils.MetaNone && field.Path == utils.EmptyString {
					return fmt.Errorf("<%s> %s for %s at %s", utils.SIPAgent, utils.NewErrMandatoryIeMissing(utils.Path), req.ID, field.Tag)
				}
				if err := utils.IsPathValidForExporters(field.Path); err != nil {
					return fmt.Errorf("<%s> %s for %s at %s", utils.SIPAgent, err, field.Path, utils.Path)
				}
				for _, val := range field.Value {
					if err := utils.IsPathValidForExporters(val.path); err != nil {
						return fmt.Errorf("<%s> %s for %s at %s of %s", utils.SIPAgent, err, val.path, utils.Values, utils.ReplyFieldsCfg)
					}
				}
				if err := utils.CheckInLineFilter(field.Filters); err != nil {
					return fmt.Errorf("<%s> %s for %s at %s", utils.SIPAgent, err, field.Filters, utils.ReplyFieldsCfg)
				}
			}
			if err := utils.CheckInLineFilter(req.Filters); err != nil {
				return fmt.Errorf("<%s> %s for %s at %s", utils.SIPAgent, err, req.Filters, utils.RequestProcessorsCfg)
			}
		}
	}

	if cfg.attributeSCfg.Enabled {
		if cfg.attributeSCfg.Opts.ProcessRuns < 1 {
			return fmt.Errorf("<%s> process_runs needs to be bigger than 0", utils.AttributeS)
		}
	}
	if cfg.chargerSCfg.Enabled {
		for _, connID := range cfg.chargerSCfg.AttributeSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.attributeSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.AttributeS, utils.ChargerS)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.ChargerS, connID)
			}
		}
	}
	// ResourceLimiter checks
	if cfg.resourceSCfg.Enabled {
		for _, connID := range cfg.resourceSCfg.ThresholdSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.thresholdSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.ThresholdS, utils.ResourceS)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.ResourceS, connID)
			}
		}
	}
	// StatS checks
	if cfg.statsCfg.Enabled {
		for _, connID := range cfg.statsCfg.ThresholdSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.thresholdSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.ThresholdS, utils.StatS)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.StatS, connID)
			}
		}
	}
	// RouteS checks
	if cfg.routeSCfg.Enabled {
		for _, connID := range cfg.routeSCfg.AttributeSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.attributeSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.AttributeS, utils.RouteS)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.RouteS, connID)
			}
		}
		for _, connID := range cfg.routeSCfg.StatSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.statsCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.StatService, utils.RouteS)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.RouteS, connID)
			}
		}
		for _, connID := range cfg.routeSCfg.ResourceSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.resourceSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.ResourceS, utils.RouteS)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.RouteS, connID)
			}
		}
		for _, connID := range cfg.routeSCfg.RALsConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.ralsCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.RALService, utils.RouteS)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.RouteS, connID)
			}
		}
	}
	// Scheduler check connection with CDR Server
	if cfg.schedulerCfg.Enabled {
		for _, connID := range cfg.schedulerCfg.CDRsConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.cdrsCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.CDRs, utils.SchedulerS)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.SchedulerS, connID)
			}
		}
		for _, connID := range cfg.schedulerCfg.ThreshSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.thresholdSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.ThresholdS, utils.SchedulerS)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.SchedulerS, connID)
			}
		}
		for _, connID := range cfg.schedulerCfg.StatSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.statsCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.StatS, utils.SchedulerS)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.SchedulerS, connID)
			}
		}
		if err := utils.CheckInLineFilter(cfg.schedulerCfg.Filters); err != nil {
			return fmt.Errorf("<%s> got %s in %s", utils.SchedulerS, err, utils.Filters)
		}

	}
	// EventReader sanity checks
	if cfg.ersCfg.Enabled {
		for _, connID := range cfg.ersCfg.SessionSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.sessionSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.SessionS, utils.ERs)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.ERs, connID)
			}
		}
		for _, rdr := range cfg.ersCfg.Readers {
			if !possibleReaderTypes.Has(rdr.Type) {
				return fmt.Errorf("<%s> unsupported data type: %s for reader with ID: %s", utils.ERs, rdr.Type, rdr.ID)
			}
			var pAct string
			if rdr.Opts.PartialCacheAction != nil {
				pAct = *rdr.Opts.PartialCacheAction
			}
			if pAct != utils.MetaDumpToFile &&
				pAct != utils.MetaNone &&
				pAct != utils.MetaPostCDR &&
				pAct != utils.MetaDumpToJSON {
				return fmt.Errorf("<%s> wrong partial expiry action for reader with ID: %s", utils.ERs, rdr.ID)
			}
			if pAct != utils.MetaNone { // if is *none we do not process the evicted events
				if rdr.Opts.PartialOrderField != nil && *rdr.Opts.PartialOrderField == utils.EmptyString { // the field we order after must not be empty
					return fmt.Errorf("<%s> empty %s for reader with ID: %s", utils.ERs, utils.PartialOrderFieldOpt, rdr.ID)
				}
			}
			if pAct == utils.MetaDumpToFile ||
				pAct == utils.MetaDumpToJSON { // only if the action is *dump_to_file
				path := rdr.ProcessedPath
				if rdr.Opts.PartialPath != nil {
					path = *rdr.Opts.PartialPath
				}
				if _, err := os.Stat(path); err != nil && os.IsNotExist(err) {
					return fmt.Errorf("<%s> nonexistent partial folder: %s for reader with ID: %s", utils.ERs, path, rdr.ID)
				}
				if pAct == utils.MetaDumpToFile {
					if rdr.Opts.PartialCSVFieldSeparator != nil && // the separtor must not be empty
						*rdr.Opts.PartialCSVFieldSeparator == utils.EmptyString {
						return fmt.Errorf("<%s> empty %s for reader with ID: %s", utils.ERs, utils.PartialCSVFieldSepartorOpt, rdr.ID)
					}
				}
			}
			switch rdr.Type {
			case utils.MetaFileCSV:
				paths := []string{rdr.ProcessedPath, rdr.SourcePath}
				if rdr.ProcessedPath == utils.EmptyString {
					paths = []string{rdr.SourcePath}
				}
				for _, dir := range paths {
					if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
						return fmt.Errorf("<%s> nonexistent folder: %s for reader with ID: %s", utils.ERs, dir, rdr.ID)
					}
				}
				if rdr.Opts.CSVFieldSeparator != nil &&
					*rdr.Opts.CSVFieldSeparator == utils.EmptyString {
					return fmt.Errorf("<%s> empty %s for reader with ID: %s", utils.ERs, utils.CSVFieldSepOpt, rdr.ID)
				}
			case utils.MetaKafkajsonMap:
				if rdr.RunDelay > 0 {
					return fmt.Errorf("<%s> the RunDelay field can not be bigger than zero for reader with ID: %s", utils.ERs, rdr.ID)
				}
			case utils.MetaFileXML, utils.MetaFileFWV, utils.MetaFileJSON:
				for _, dir := range []string{rdr.ProcessedPath, rdr.SourcePath} {
					if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
						return fmt.Errorf("<%s> nonexistent folder: %s for reader with ID: %s", utils.ERs, dir, rdr.ID)
					}
				}
			}
			for _, field := range rdr.CacheDumpFields {
				if field.Type != utils.MetaNone && field.Path == utils.EmptyString {
					return fmt.Errorf("<%s> %s for %s at %s", utils.ERs, utils.NewErrMandatoryIeMissing(utils.Path), rdr.ID, field.Tag)
				}
				if err := utils.IsPathValidForExporters(field.Path); err != nil {
					return fmt.Errorf("<%s> %s for %s at %s", utils.ERs, err, field.Path, utils.Path)
				}
				if field.Type == utils.MetaVariable ||
					field.Type == utils.MetaComposed ||
					field.Type == utils.MetaGroup ||
					field.Type == utils.MetaUsageDifference ||
					field.Type == utils.MetaCCUsage ||
					field.Type == utils.MetaSum ||
					field.Type == utils.MetaDifference ||
					field.Type == utils.MetaMultiply ||
					field.Type == utils.MetaDivide ||
					field.Type == utils.MetaValueExponent ||
					field.Type == utils.MetaUnixTimestamp ||
					field.Type == utils.MetaSIPCID {
					for _, val := range field.Value {
						if err := utils.IsPathValidForExporters(val.path); err != nil {
							return fmt.Errorf("<%s> %s for %s at %s of %s", utils.ERs, err, val.path, utils.Values, utils.CacheDumpFieldsCfg)
						}
					}
				}
				if err := utils.CheckInLineFilter(field.Filters); err != nil {
					return fmt.Errorf("<%s> %s for %s at %s", utils.ERs, err, field.Filters, utils.CacheDumpFieldsCfg)
				}
			}
			for _, field := range rdr.Fields {
				if field.Type != utils.MetaNone && field.Path == utils.EmptyString {
					return fmt.Errorf("<%s> %s for %s at %s", utils.ERs, utils.NewErrMandatoryIeMissing(utils.Path), rdr.ID, field.Tag)
				}
				if err := utils.IsPathValidForExporters(field.Path); err != nil {
					return fmt.Errorf("<%s> %s for %s at %s", utils.ERs, err, field.Path, utils.Path)
				}
				if field.Type == utils.MetaVariable ||
					field.Type == utils.MetaComposed ||
					field.Type == utils.MetaGroup ||
					field.Type == utils.MetaUsageDifference ||
					field.Type == utils.MetaCCUsage ||
					field.Type == utils.MetaSum ||
					field.Type == utils.MetaDifference ||
					field.Type == utils.MetaMultiply ||
					field.Type == utils.MetaDivide ||
					field.Type == utils.MetaValueExponent ||
					field.Type == utils.MetaUnixTimestamp ||
					field.Type == utils.MetaSIPCID {
					for _, val := range field.Value {
						if err := utils.IsPathValidForExporters(val.path); err != nil {
							return fmt.Errorf("<%s> %s for %s at %s of %s", utils.ERs, err, val.path, utils.Values, utils.FieldsCfg)
						}
					}
				}
				if err := utils.CheckInLineFilter(field.Filters); err != nil {
					return fmt.Errorf("<%s> %s for %s at %s", utils.ERs, err, field.Filters, utils.FieldsCfg)
				}
			}
			if err := utils.CheckInLineFilter(rdr.Filters); err != nil {
				return fmt.Errorf("<%s> %s for %s at %s", utils.ERs, err, rdr.Filters, utils.ReadersCfg)
			}
		}
	}
	// EventExporter sanity checks
	if cfg.eesCfg.Enabled {
		for _, connID := range cfg.eesCfg.AttributeSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.attributeSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.AttributeS, utils.EEs)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.EEs, connID)
			}
		}
		for _, exp := range cfg.eesCfg.Exporters {
			if !possibleExporterTypes.Has(exp.Type) {
				return fmt.Errorf("<%s> unsupported data type: %s for exporter with ID: %s", utils.EEs, exp.Type, exp.ID)
			}

			switch exp.Type {
			case utils.MetaFileCSV:
				for _, dir := range []string{exp.ExportPath} {
					if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
						return fmt.Errorf("<%s> nonexistent folder: %s for exporter with ID: %s", utils.EEs, dir, exp.ID)
					}
				}
				if exp.Opts.CSVFieldSeparator != nil && *exp.Opts.CSVFieldSeparator == utils.EmptyString {
					return fmt.Errorf("<%s> empty %s for exporter with ID: %s", utils.EEs, utils.CSVFieldSepOpt, exp.ID)
				}
			case utils.MetaFileFWV:
				for _, dir := range []string{exp.ExportPath} {
					if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
						return fmt.Errorf("<%s> nonexistent folder: %s for exporter with ID: %s", utils.EEs, dir, exp.ID)
					}
				}
			case utils.MetaSQL:
				if len(exp.ContentFields()) == 0 {
					return fmt.Errorf("<%s> empty content fields for exporter with ID: %s", utils.EEs, exp.ID)
				}
			}
			for _, field := range exp.Fields {
				if field.Type != utils.MetaNone && field.Path == utils.EmptyString {
					return fmt.Errorf("<%s> %s for %s at %s", utils.EEs, utils.NewErrMandatoryIeMissing(utils.Path), exp.ID, field.Tag)
				}
				if err := utils.IsPathValidForExporters(field.Path); err != nil {
					return fmt.Errorf("<%s> %s for %s at %s", utils.EEs, err, field.Path, utils.Path)
				}
				if field.Type == utils.MetaVariable ||
					field.Type == utils.MetaComposed ||
					field.Type == utils.MetaGroup ||
					field.Type == utils.MetaUsageDifference ||
					field.Type == utils.MetaCCUsage ||
					field.Type == utils.MetaSum ||
					field.Type == utils.MetaDifference ||
					field.Type == utils.MetaMultiply ||
					field.Type == utils.MetaDivide ||
					field.Type == utils.MetaValueExponent ||
					field.Type == utils.MetaUnixTimestamp ||
					field.Type == utils.MetaSIPCID {
					for _, val := range field.Value {
						if err := utils.IsPathValidForExporters(val.path); err != nil {
							return fmt.Errorf("<%s> %s for %s at %s of %s", utils.EEs, err, val.path, utils.Values, utils.FieldsCfg)
						}
					}
				}
				if err := utils.CheckInLineFilter(field.Filters); err != nil {
					return fmt.Errorf("<%s> %s for %s at %s", utils.EEs, err, field.Filters, utils.FieldsCfg)
				}
			}
			if err := utils.CheckInLineFilter(exp.Filters); err != nil {
				return fmt.Errorf("<%s> %s for %s at %s", utils.EEs, err, exp.Filters, utils.ExportersCfg)
			}
		}
	}
	// StorDB sanity checks
	if cfg.storDbCfg.Type == utils.Postgres {
		if !utils.IsSliceMember([]string{utils.PostgresSSLModeDisable, utils.PostgressSSLModeAllow,
			utils.PostgresSSLModePrefer, utils.PostgressSSLModeRequire, utils.PostgresSSLModeVerifyCa,
			utils.PostgresSSLModeVerifyFull}, cfg.storDbCfg.Opts.PgSSLMode) {
			return fmt.Errorf("<%s> unsupported sslmode for storDB", utils.StorDB)
		}
	}
	// DataDB sanity checks
	if cfg.dataDbCfg.Type == utils.Internal {
		for key, config := range cfg.cacheCfg.Partitions {
			if utils.DataDBPartitions.Has(key) && config.Limit != 0 {
				return fmt.Errorf("<%s> %s needs to be 0 when DataBD is *internal, received : %d", utils.CacheS, key, config.Limit)
			}
		}
		if cfg.resourceSCfg.Enabled && cfg.resourceSCfg.StoreInterval != -1 {
			return fmt.Errorf("<%s> the StoreInterval field needs to be -1 when DataBD is *internal, received : %d", utils.ResourceS, cfg.resourceSCfg.StoreInterval)
		}
		if cfg.statsCfg.Enabled && cfg.statsCfg.StoreInterval != -1 {
			return fmt.Errorf("<%s> the StoreInterval field needs to be -1 when DataBD is *internal, received : %d", utils.StatS, cfg.statsCfg.StoreInterval)
		}
		if cfg.thresholdSCfg.Enabled && cfg.thresholdSCfg.StoreInterval != -1 {
			return fmt.Errorf("<%s> the StoreInterval field needs to be -1 when DataBD is *internal, received : %d", utils.ThresholdS, cfg.thresholdSCfg.StoreInterval)
		}
	}
	for item, val := range cfg.dataDbCfg.Items {
		if val.Remote && len(cfg.dataDbCfg.RmtConns) == 0 {
			return fmt.Errorf("remote connections required by: <%s>", item)
		}
		if val.Replicate && len(cfg.dataDbCfg.RplConns) == 0 {
			return fmt.Errorf("replicate connections required by: <%s>", item)
		}
	}
	for _, connID := range cfg.dataDbCfg.RplConns {
		conn, has := cfg.rpcConns[connID]
		if !has {
			return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.DataDB, connID)
		}
		for _, rpc := range conn.Conns {
			if rpc.Transport != utils.MetaGOB {
				return fmt.Errorf("<%s> unsupported transport <%s> for connection with ID: <%s>", utils.DataDB, rpc.Transport, connID)
			}
		}
	}
	for _, connID := range cfg.dataDbCfg.RmtConns {
		conn, has := cfg.rpcConns[connID]
		if !has {
			return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.DataDB, connID)
		}
		for _, rpc := range conn.Conns {
			if rpc.Transport != utils.MetaGOB {
				return fmt.Errorf("<%s> unsupported transport <%s> for connection with ID: <%s>", utils.DataDB, rpc.Transport, connID)
			}
		}
	}
	// APIer sanity checks
	for _, connID := range cfg.apier.AttributeSConns {
		if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.attributeSCfg.Enabled {
			return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.AttributeS, utils.APIerSv1)
		}
		if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
			return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.APIerSv1, connID)
		}
	}
	for _, connID := range cfg.apier.SchedulerConns {
		if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.schedulerCfg.Enabled {
			return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.SchedulerS, utils.APIerSv1)
		}
		if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
			return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.APIerSv1, connID)
		}
	}
	// Dispatcher sanity check
	if cfg.dispatcherSCfg.Enabled {
		for _, connID := range cfg.dispatcherSCfg.AttributeSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.attributeSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.AttributeS, utils.DispatcherS)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.DispatcherS, connID)
			}
		}
	}
	// Cache check
	for _, connID := range cfg.cacheCfg.ReplicationConns {
		conn, has := cfg.rpcConns[connID]
		if !has {
			return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.CacheS, connID)
		}
		for _, rpc := range conn.Conns {
			if rpc.Transport != utils.MetaGOB {
				return fmt.Errorf("<%s> unsupported transport <%s> for connection with ID: <%s>", utils.CacheS, rpc.Transport, connID)
			}
		}
	}
	for cacheID, itm := range cfg.cacheCfg.Partitions {
		if !utils.CachePartitions.Has(cacheID) {
			return fmt.Errorf("<%s> partition <%s> not defined", utils.CacheS, cacheID)
		}
		if cacheID == utils.CacheRPCConnections &&
			itm.Replicate {
			return fmt.Errorf("<%s> partition <%s> does not support replication", utils.CacheS, cacheID) // deadlock prevention
		}
	}
	// FilterS sanity check
	for _, connID := range cfg.filterSCfg.StatSConns {
		if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.statsCfg.Enabled {
			return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.StatS, utils.FilterS)
		}
		if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
			return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.FilterS, connID)
		}
	}
	for _, connID := range cfg.filterSCfg.ResourceSConns {
		if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.resourceSCfg.Enabled {
			return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.ResourceS, utils.FilterS)
		}
		if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
			return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.FilterS, connID)
		}
	}
	for _, connID := range cfg.filterSCfg.ApierSConns {
		if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.apier.Enabled {
			return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.ApierS, utils.FilterS)
		}
		if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
			return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.FilterS, connID)
		}
	}

	if len(cfg.registrarCCfg.Dispatchers.RegistrarSConns) != 0 {
		if len(cfg.registrarCCfg.Dispatchers.Hosts) == 0 {
			return fmt.Errorf("<%s> missing dispatcher host IDs", utils.RegistrarC)
		}
		if cfg.registrarCCfg.Dispatchers.RefreshInterval <= 0 {
			return fmt.Errorf("<%s> the register imterval needs to be bigger than 0", utils.RegistrarC)
		}
		for tnt, hosts := range cfg.registrarCCfg.Dispatchers.Hosts {
			for _, host := range hosts {
				if !utils.SliceHasMember([]string{utils.MetaGOB, rpcclient.HTTPjson, utils.MetaJSON, rpcclient.BiRPCJSON, rpcclient.BiRPCGOB}, host.Transport) {
					return fmt.Errorf("<%s> unsupported transport <%s> for host <%s>", utils.RegistrarC, host.Transport, utils.ConcatenatedKey(tnt, host.ID))
				}
			}
		}
		for _, connID := range cfg.registrarCCfg.Dispatchers.RegistrarSConns {
			if connID == utils.MetaInternal {
				return fmt.Errorf("<%s> internal connection IDs are not supported", utils.RegistrarC)
			}
			connCfg, has := cfg.rpcConns[connID]
			if !has {
				return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.RegistrarC, connID)
			}
			if len(connCfg.Conns) != 1 {
				return fmt.Errorf("<%s> connection with id: <%s> needs to have only one host", utils.RegistrarC, connID)
			}
			if connCfg.Conns[0].Transport != rpcclient.HTTPjson {
				return fmt.Errorf("<%s> connection with id: <%s> unsupported transport <%s>", utils.RegistrarC, connID, connCfg.Conns[0].Transport)
			}
		}
	}

	if len(cfg.registrarCCfg.RPC.RegistrarSConns) != 0 {
		if len(cfg.registrarCCfg.RPC.Hosts) == 0 {
			return fmt.Errorf("<%s> missing RPC host IDs", utils.RegistrarC)
		}
		if cfg.registrarCCfg.RPC.RefreshInterval <= 0 {
			return fmt.Errorf("<%s> the register imterval needs to be bigger than 0", utils.RegistrarC)
		}
		for tnt, hosts := range cfg.registrarCCfg.RPC.Hosts {
			for _, host := range hosts {
				if !utils.SliceHasMember([]string{utils.MetaGOB, rpcclient.HTTPjson, utils.MetaJSON, rpcclient.BiRPCJSON, rpcclient.BiRPCGOB}, host.Transport) {
					return fmt.Errorf("<%s> unsupported transport <%s> for host <%s>", utils.RegistrarC, host.Transport, utils.ConcatenatedKey(tnt, host.ID))
				}
			}
		}
		for _, connID := range cfg.registrarCCfg.RPC.RegistrarSConns {
			if connID == utils.MetaInternal {
				return fmt.Errorf("<%s> internal connection IDs are not supported", utils.RegistrarC)
			}
			connCfg, has := cfg.rpcConns[connID]
			if !has {
				return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.RegistrarC, connID)
			}
			if len(connCfg.Conns) != 1 {
				return fmt.Errorf("<%s> connection with id: <%s> needs to have only one host", utils.RegistrarC, connID)
			}
			if connCfg.Conns[0].Transport != rpcclient.HTTPjson {
				return fmt.Errorf("<%s> connection with id: <%s> unsupported transport <%s>", utils.RegistrarC, connID, connCfg.Conns[0].Transport)
			}
		}
	}

	if cfg.analyzerSCfg.Enabled {
		if !utils.AnzIndexType.Has(cfg.analyzerSCfg.IndexType) {
			return fmt.Errorf("<%s> unsupported index type: %q", utils.AnalyzerS, cfg.analyzerSCfg.IndexType)
		}
		if cfg.analyzerSCfg.IndexType != utils.MetaInternal {
			if _, err := os.Stat(cfg.analyzerSCfg.DBPath); err != nil && os.IsNotExist(err) {
				return fmt.Errorf("<%s> nonexistent DB folder: %q", utils.AnalyzerS, cfg.analyzerSCfg.DBPath)
			}
		}
		// TTL and CleanupInterval should always be biger than zero in order to not keep unecesary logs in index
		if cfg.analyzerSCfg.TTL <= 0 {
			return fmt.Errorf("<%s> the TTL needs to be bigger than 0", utils.AnalyzerS)
		}
		if cfg.analyzerSCfg.CleanupInterval <= 0 {
			return fmt.Errorf("<%s> the CleanupInterval needs to be bigger than 0", utils.AnalyzerS)
		}
	}

	return nil
}
