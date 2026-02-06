/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package config

import (
	"fmt"
	"math"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/cron"
	"github.com/cgrates/rpcclient"
)

// CheckConfigSanity is used in cgr-engine
func (cfg *CGRConfig) CheckConfigSanity() error {
	return cfg.checkConfigSanity()
}

func (cfg *CGRConfig) checkConfigSanity() error {

	// CDRServer checks
	if cfg.cdrsCfg.Enabled {
		cdrConnEnabledMap := map[string]struct {
			name    string
			enabled bool
		}{
			utils.MetaChargers:   {utils.ChargerS, cfg.chargerSCfg.Enabled},
			utils.MetaAttributes: {utils.AttributeS, cfg.attributeSCfg.Enabled},
			utils.MetaStats:      {utils.StatService, cfg.statsCfg.Enabled},
			utils.MetaThresholds: {utils.ThresholdS, cfg.thresholdSCfg.Enabled},
			utils.MetaEEs:        {utils.EEs, cfg.eesCfg.Enabled},
		}
		for connType, opts := range cfg.cdrsCfg.Conns {
			for _, opt := range opts {
				for _, connID := range opt.Values {
					if info, has := cdrConnEnabledMap[connType]; has {
						if strings.HasPrefix(connID, utils.MetaInternal) && !info.enabled {
							return fmt.Errorf("<%s> not enabled but requested by <%s> component", info.name, utils.CDRs)
						}
					}
					if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
						return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.CDRs, connID)
					}
				}
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
			if !possibleLoaderTypes.Has(data.Type) {
				return fmt.Errorf("<%s> unsupported data type %s", utils.LoaderS, data.Type)
			}

			for _, field := range data.Fields {
				if field.Path == utils.EmptyString {
					return fmt.Errorf("<%s> %s for %s at %s", utils.LoaderS, utils.NewErrMandatoryIeMissing(utils.Path), data.Type, field.Tag)
				}
				if err := utils.IsPathValidForExporters(field.Path); err != nil {
					return fmt.Errorf("<%s> %s for %s at %s", utils.LoaderS, err, field.Path, utils.Path)
				}
				for _, val := range field.Value {
					if err := utils.IsPathValidForExporters(val.Path); err != nil {
						return fmt.Errorf("<%s> %s for %s at %s", utils.LoaderS, err, val.Path, utils.Values)
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
		connEnabledMap := map[string]struct {
			name    string
			enabled bool
		}{
			utils.MetaChargers:   {utils.ChargerS, cfg.chargerSCfg.Enabled},
			utils.MetaResources:  {utils.ResourceS, cfg.resourceSCfg.Enabled},
			utils.MetaIPs:        {utils.IPs, cfg.ipsCfg.Enabled},
			utils.MetaThresholds: {utils.ThresholdS, cfg.thresholdSCfg.Enabled},
			utils.MetaStats:      {utils.StatService, cfg.statsCfg.Enabled},
			utils.MetaRoutes:     {utils.RouteS, cfg.routeSCfg.Enabled},
			utils.MetaAttributes: {utils.AttributeS, cfg.attributeSCfg.Enabled},
			utils.MetaCDRs:       {utils.CDRs, cfg.cdrsCfg.Enabled},
			utils.MetaActions:    {utils.ActionS, cfg.actionSCfg.Enabled},
			utils.MetaRates:      {utils.RateS, cfg.rateSCfg.Enabled},
			utils.MetaAccounts:   {utils.AccountS, cfg.accountSCfg.Enabled},
		}
		for connType, opts := range cfg.sessionSCfg.Conns {
			for _, opt := range opts {
				for _, connID := range opt.Values {
					if info, has := connEnabledMap[connType]; has {
						if strings.HasPrefix(connID, utils.MetaInternal) && !info.enabled {
							return fmt.Errorf("<%s> not enabled but requested by <%s> component", info.name, utils.SessionS)
						}
					}
					if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
						return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.SessionS, connID)
					}
				}
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
		if len(cfg.fsAgentCfg.Conns[utils.MetaSessionS]) == 0 {
			return fmt.Errorf("<%s> no %s connections defined",
				utils.FreeSWITCHAgent, utils.SessionS)
		}
		fsConnEnabledMap := map[string]struct {
			name    string
			enabled bool
		}{
			utils.MetaSessionS: {utils.SessionS, cfg.sessionSCfg.Enabled},
		}
		for connType, opts := range cfg.fsAgentCfg.Conns {
			for _, opt := range opts {
				for _, connID := range opt.Values {
					isInternal := strings.HasPrefix(connID, utils.MetaInternal) || strings.HasPrefix(connID, rpcclient.BiRPCInternal)
					if info, has := fsConnEnabledMap[connType]; has && isInternal && !info.enabled {
						return fmt.Errorf("<%s> not enabled but requested by <%s> component", info.name, utils.FreeSWITCHAgent)
					}
					if _, has := cfg.rpcConns[connID]; !has && !isInternal {
						return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.FreeSWITCHAgent, connID)
					}
				}
			}
		}
	}
	// KamailioAgent checks
	if cfg.kamAgentCfg.Enabled {
		if len(cfg.kamAgentCfg.Conns[utils.MetaSessionS]) == 0 {
			return fmt.Errorf("<%s> no %s connections defined",
				utils.KamailioAgent, utils.SessionS)
		}
		kamConnEnabledMap := map[string]struct {
			name    string
			enabled bool
		}{
			utils.MetaSessionS: {utils.SessionS, cfg.sessionSCfg.Enabled},
		}
		for connType, opts := range cfg.kamAgentCfg.Conns {
			for _, opt := range opts {
				for _, connID := range opt.Values {
					isInternal := strings.HasPrefix(connID, utils.MetaInternal) || strings.HasPrefix(connID, rpcclient.BiRPCInternal)
					if info, has := kamConnEnabledMap[connType]; has && isInternal && !info.enabled {
						return fmt.Errorf("<%s> not enabled but requested by <%s> component", info.name, utils.KamailioAgent)
					}
					if _, has := cfg.rpcConns[connID]; !has && !isInternal {
						return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.KamailioAgent, connID)
					}
				}
			}
		}
	}
	// AsteriskAgent checks
	if cfg.asteriskAgentCfg.Enabled {
		if len(cfg.asteriskAgentCfg.Conns[utils.MetaSessionS]) == 0 {
			return fmt.Errorf("<%s> no %s connections defined",
				utils.AsteriskAgent, utils.SessionS)
		}
		astConnEnabledMap := map[string]struct {
			name    string
			enabled bool
		}{
			utils.MetaSessionS: {utils.SessionS, cfg.sessionSCfg.Enabled},
		}
		for connType, opts := range cfg.asteriskAgentCfg.Conns {
			for _, opt := range opts {
				for _, connID := range opt.Values {
					isInternal := strings.HasPrefix(connID, utils.MetaInternal) || strings.HasPrefix(connID, rpcclient.BiRPCInternal)
					if info, has := astConnEnabledMap[connType]; has && isInternal && !info.enabled {
						return fmt.Errorf("<%s> not enabled but requested by <%s> component", info.name, utils.AsteriskAgent)
					}
					if _, has := cfg.rpcConns[connID]; !has && !isInternal {
						return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.AsteriskAgent, connID)
					}
				}
			}
		}
	}
	// DAgent checks
	if cfg.diameterAgentCfg.Enabled {
		if len(cfg.diameterAgentCfg.Conns[utils.MetaSessionS]) == 0 {
			return fmt.Errorf("<%s> no %s connections defined",
				utils.DiameterAgent, utils.SessionS)
		}
		if len(cfg.diameterAgentCfg.ConnStatusStatQueueIDs) != 0 && len(cfg.diameterAgentCfg.Conns[utils.MetaStats]) == 0 {
			return fmt.Errorf("<%s> stat_queue_ids defined but no %s connections configured",
				utils.DiameterAgent, utils.StatS)
		}
		if len(cfg.diameterAgentCfg.ConnStatusThresholdIDs) != 0 && len(cfg.diameterAgentCfg.Conns[utils.MetaThresholds]) == 0 {
			return fmt.Errorf("<%s> threshold_ids defined but no %s connections configured",
				utils.DiameterAgent, utils.ThresholdS)
		}
		daConnEnabledMap := map[string]struct {
			name    string
			enabled bool
		}{
			utils.MetaSessionS:   {utils.SessionS, cfg.sessionSCfg.Enabled},
			utils.MetaStats:      {utils.StatS, cfg.statsCfg.Enabled},
			utils.MetaThresholds: {utils.ThresholdS, cfg.thresholdSCfg.Enabled},
		}
		for connType, opts := range cfg.diameterAgentCfg.Conns {
			for _, opt := range opts {
				for _, connID := range opt.Values {
					isInternal := strings.HasPrefix(connID, utils.MetaInternal) || strings.HasPrefix(connID, rpcclient.BiRPCInternal)
					if info, has := daConnEnabledMap[connType]; has && isInternal && !info.enabled {
						return fmt.Errorf("<%s> not enabled but requested by <%s> component", info.name, utils.DiameterAgent)
					}
					if _, has := cfg.rpcConns[connID]; !has && !isInternal {
						return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.DiameterAgent, connID)
					}
				}
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
					if err := utils.IsPathValidForExporters(val.Path); err != nil {
						return fmt.Errorf("<%s> %s for %s at %s", utils.DiameterAgent, err, val.Path, utils.Values)
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
					if err := utils.IsPathValidForExporters(val.Path); err != nil {
						return fmt.Errorf("<%s> %s for %s at %s of %s", utils.DiameterAgent, err, val.Path, utils.Values, utils.RequestFieldsCfg)
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
					if err := utils.IsPathValidForExporters(val.Path); err != nil {
						return fmt.Errorf("<%s> %s for %s at %s of %s", utils.DiameterAgent, err, val.Path, utils.Values, utils.ReplyFieldsCfg)
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
		if cfg.radiusAgentCfg.CoATemplate != "" {
			if _, found := cfg.templates[cfg.radiusAgentCfg.CoATemplate]; !found {
				return fmt.Errorf("<%s> CoA Template %s not defined", utils.RadiusAgent, cfg.radiusAgentCfg.CoATemplate)
			}
		}
		if cfg.radiusAgentCfg.DMRTemplate != "" {
			if _, found := cfg.templates[cfg.radiusAgentCfg.DMRTemplate]; !found {
				return fmt.Errorf("<%s> DMR Template %s not defined", utils.RadiusAgent, cfg.radiusAgentCfg.DMRTemplate)
			}
		}
		if len(cfg.radiusAgentCfg.Conns[utils.MetaSessionS]) == 0 {
			return fmt.Errorf("<%s> no %s connections defined",
				utils.RadiusAgent, utils.SessionS)
		}
		raConnEnabledMap := map[string]struct {
			name    string
			enabled bool
		}{
			utils.MetaSessionS:   {utils.SessionS, cfg.sessionSCfg.Enabled},
			utils.MetaStats:      {utils.StatS, cfg.statsCfg.Enabled},
			utils.MetaThresholds: {utils.ThresholdS, cfg.thresholdSCfg.Enabled},
		}
		for connType, opts := range cfg.radiusAgentCfg.Conns {
			for _, opt := range opts {
				for _, connID := range opt.Values {
					isInternal := strings.HasPrefix(connID, utils.MetaInternal) || strings.HasPrefix(connID, rpcclient.BiRPCInternal)
					if info, has := raConnEnabledMap[connType]; has && isInternal && !info.enabled {
						return fmt.Errorf("<%s> not enabled but requested by <%s> component", info.name, utils.RadiusAgent)
					}
					if _, has := cfg.rpcConns[connID]; !has && !isInternal {
						return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.RadiusAgent, connID)
					}
				}
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
					if err := utils.IsPathValidForExporters(val.Path); err != nil {
						return fmt.Errorf("<%s> %s for %s at %s of %s", utils.RadiusAgent, err, val.Path, utils.Values, utils.RequestFieldsCfg)
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
					if err := utils.IsPathValidForExporters(val.Path); err != nil {
						return fmt.Errorf("<%s> %s for %s at %s of %s", utils.RadiusAgent, err, val.Path, utils.Values, utils.ReplyFieldsCfg)
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
		if len(cfg.dnsAgentCfg.Conns[utils.MetaSessionS]) == 0 {
			return fmt.Errorf("<%s> no %s connections defined",
				utils.DNSAgent, utils.SessionS)
		}
		dnsConnEnabledMap := map[string]struct {
			name    string
			enabled bool
		}{
			utils.MetaSessionS:   {utils.SessionS, cfg.sessionSCfg.Enabled},
			utils.MetaStats:      {utils.StatS, cfg.statsCfg.Enabled},
			utils.MetaThresholds: {utils.ThresholdS, cfg.thresholdSCfg.Enabled},
		}
		for connType, opts := range cfg.dnsAgentCfg.Conns {
			for _, opt := range opts {
				for _, connID := range opt.Values {
					isInternal := strings.HasPrefix(connID, utils.MetaInternal) || strings.HasPrefix(connID, rpcclient.BiRPCInternal)
					if info, has := dnsConnEnabledMap[connType]; has && isInternal && !info.enabled {
						return fmt.Errorf("<%s> not enabled but requested by <%s> component", info.name, utils.DNSAgent)
					}
					if _, has := cfg.rpcConns[connID]; !has && !isInternal {
						return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.DNSAgent, connID)
					}
				}
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
					if err := utils.IsPathValidForExporters(val.Path); err != nil {
						return fmt.Errorf("<%s> %s for %s at %s of %s", utils.DNSAgent, err, val.Path, utils.Values, utils.RequestFieldsCfg)
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
					if err := utils.IsPathValidForExporters(val.Path); err != nil {
						return fmt.Errorf("<%s> %s for %s at %s of %s", utils.DNSAgent, err, val.Path, utils.Values, utils.ReplyFieldsCfg)
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
		for _, opt := range httpAgentCfg.Conns[utils.MetaSessionS] {
			for _, connID := range opt.Values {
				isInternal := strings.HasPrefix(connID, utils.MetaInternal) || strings.HasPrefix(connID, rpcclient.BiRPCInternal)
				if isInternal && !cfg.sessionSCfg.Enabled {
					return fmt.Errorf("<%s> not enabled but requested by <%s> HTTPAgent Template", utils.SessionS, httpAgentCfg.ID)
				}
				if _, has := cfg.rpcConns[connID]; !has && !isInternal {
					return fmt.Errorf("<%s> template with ID <%s> has connection with id: <%s> not defined", utils.HTTPAgent, httpAgentCfg.ID, connID)
				}
			}
		}
		for _, opt := range httpAgentCfg.Conns[utils.MetaStats] {
			for _, connID := range opt.Values {
				isInternal := strings.HasPrefix(connID, utils.MetaInternal)
				if isInternal && !cfg.statsCfg.Enabled {
					return fmt.Errorf("<%s> not enabled but requested by <%s> HTTPAgent Template", utils.StatS, httpAgentCfg.ID)
				}
				if _, has := cfg.rpcConns[connID]; !has && !isInternal {
					return fmt.Errorf("<%s> template with ID <%s> has connection with id: <%s> not defined", utils.HTTPAgent, httpAgentCfg.ID, connID)
				}
			}
		}
		for _, opt := range httpAgentCfg.Conns[utils.MetaThresholds] {
			for _, connID := range opt.Values {
				isInternal := strings.HasPrefix(connID, utils.MetaInternal)
				if isInternal && !cfg.thresholdSCfg.Enabled {
					return fmt.Errorf("<%s> not enabled but requested by <%s> HTTPAgent Template", utils.ThresholdS, httpAgentCfg.ID)
				}
				if _, has := cfg.rpcConns[connID]; !has && !isInternal {
					return fmt.Errorf("<%s> template with ID <%s> has connection with id: <%s> not defined", utils.HTTPAgent, httpAgentCfg.ID, connID)
				}
			}
		}
		if !slices.Contains([]string{utils.MetaUrl, utils.MetaXml}, httpAgentCfg.RequestPayload) {
			return fmt.Errorf("<%s> unsupported request payload %s", utils.HTTPAgent, httpAgentCfg.RequestPayload)
		}
		if !slices.Contains([]string{utils.MetaTextPlain, utils.MetaXml}, httpAgentCfg.ReplyPayload) {
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
					if err := utils.IsPathValidForExporters(val.Path); err != nil {
						return fmt.Errorf("<%s> %s for %s at %s of %s", utils.HTTPAgent, err, val.Path, utils.Values, utils.RequestFieldsCfg)
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
					if err := utils.IsPathValidForExporters(val.Path); err != nil {
						return fmt.Errorf("<%s> %s for %s at %s of %s", utils.HTTPAgent, err, val.Path, utils.Values, utils.ReplyFieldsCfg)
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
		if len(cfg.sipAgentCfg.Conns[utils.MetaSessionS]) == 0 {
			return fmt.Errorf("<%s> no %s connections defined",
				utils.SIPAgent, utils.SessionS)
		}
		sipConnEnabledMap := map[string]struct {
			name    string
			enabled bool
		}{
			utils.MetaSessionS:   {utils.SessionS, cfg.sessionSCfg.Enabled},
			utils.MetaStats:      {utils.StatS, cfg.statsCfg.Enabled},
			utils.MetaThresholds: {utils.ThresholdS, cfg.thresholdSCfg.Enabled},
		}
		for connType, opts := range cfg.sipAgentCfg.Conns {
			for _, opt := range opts {
				for _, connID := range opt.Values {
					isInternal := strings.HasPrefix(connID, utils.MetaInternal) || strings.HasPrefix(connID, rpcclient.BiRPCInternal)
					if info, has := sipConnEnabledMap[connType]; has && isInternal && !info.enabled {
						return fmt.Errorf("<%s> not enabled but requested by <%s> component", info.name, utils.SIPAgent)
					}
					if _, has := cfg.rpcConns[connID]; !has && !isInternal {
						return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.SIPAgent, connID)
					}
				}
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
					if err := utils.IsPathValidForExporters(val.Path); err != nil {
						return fmt.Errorf("<%s> %s for %s at %s of %s", utils.SIPAgent, err, val.Path, utils.Values, utils.RequestFieldsCfg)
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
					if err := utils.IsPathValidForExporters(val.Path); err != nil {
						return fmt.Errorf("<%s> %s for %s at %s of %s", utils.SIPAgent, err, val.Path, utils.Values, utils.ReplyFieldsCfg)
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
		for _, opt := range cfg.attributeSCfg.Opts.ProcessRuns {
			if opt.value < 1 {
				return fmt.Errorf("<%s> processRuns needs to be bigger than 0", utils.AttributeS)
			}
		}
	}
	if cfg.chargerSCfg.Enabled {
		for connType, opts := range cfg.chargerSCfg.Conns {
			for _, opt := range opts {
				for _, connID := range opt.Values {
					if connType == utils.MetaAttributes && strings.HasPrefix(connID, utils.MetaInternal) && !cfg.attributeSCfg.Enabled {
						return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.AttributeS, utils.ChargerS)
					}
					if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
						return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.ChargerS, connID)
					}
				}
			}
		}
	}

	// ResourceLimiter checks
	if cfg.resourceSCfg.Enabled {
		for connType, opts := range cfg.resourceSCfg.Conns {
			for _, opt := range opts {
				for _, connID := range opt.Values {
					if connType == utils.MetaThresholds && strings.HasPrefix(connID, utils.MetaInternal) && !cfg.thresholdSCfg.Enabled {
						return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.ThresholdS, utils.ResourceS)
					}
					if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
						return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.ResourceS, connID)
					}
				}
			}
		}
	}
	// StatS checks
	if cfg.statsCfg.Enabled {
		for connType, opts := range cfg.statsCfg.Conns {
			for _, opt := range opts {
				for _, connID := range opt.Values {
					if connType == utils.MetaThresholds && strings.HasPrefix(connID, utils.MetaInternal) && !cfg.thresholdSCfg.Enabled {
						return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.ThresholdS, utils.StatS)
					}
					if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
						return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.StatS, connID)
					}
				}
			}
		}
	}
	// RouteS checks
	if cfg.routeSCfg.Enabled {
		rtConnEnabledMap := map[string]struct {
			name    string
			enabled bool
		}{
			utils.MetaAttributes: {utils.AttributeS, cfg.attributeSCfg.Enabled},
			utils.MetaStats:      {utils.StatService, cfg.statsCfg.Enabled},
			utils.MetaResources:  {utils.ResourceS, cfg.resourceSCfg.Enabled},
		}
		for connType, opts := range cfg.routeSCfg.Conns {
			for _, opt := range opts {
				for _, connID := range opt.Values {
					if info, has := rtConnEnabledMap[connType]; has {
						if strings.HasPrefix(connID, utils.MetaInternal) && !info.enabled {
							return fmt.Errorf("<%s> not enabled but requested by <%s> component", info.name, utils.RouteS)
						}
					}
					if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
						return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.RouteS, connID)
					}
				}
			}
		}
	}
	// EventReader sanity checks
	if cfg.ersCfg.Enabled {
		ersConnEnabledMap := map[string]struct {
			name    string
			enabled bool
		}{
			utils.MetaSessionS:   {utils.SessionS, cfg.sessionSCfg.Enabled},
			utils.MetaEEs:        {utils.EEs, cfg.eesCfg.Enabled},
			utils.MetaStats:      {utils.StatService, cfg.statsCfg.Enabled},
			utils.MetaThresholds: {utils.ThresholdS, cfg.thresholdSCfg.Enabled},
		}
		for connType, opts := range cfg.ersCfg.Conns {
			for _, opt := range opts {
				for _, connID := range opt.Values {
					if info, has := ersConnEnabledMap[connType]; has {
						if strings.HasPrefix(connID, utils.MetaInternal) && !info.enabled {
							return fmt.Errorf("<%s> not enabled but requested by <%s> component", info.name, utils.ERs)
						}
					}
					if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
						return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.ERs, connID)
					}
				}
			}
		}
		for _, rdr := range cfg.ersCfg.Readers {
			if len(rdr.EEsSuccessIDs) != 0 || len(rdr.EEsFailedIDs) != 0 || len(rdr.EEsIDs) != 0 {
				if len(cfg.ersCfg.Conns[utils.MetaEEs]) == 0 || !cfg.eesCfg.Enabled {
					return fmt.Errorf("<%s> connection to <%s> required due to exporter ID references", utils.ERs, utils.EEs)
				}
			}
			exporterIDs := cfg.eesCfg.exporterIDs()
			if hasInternalConnOpt(cfg.ersCfg.Conns[utils.MetaEEs]) {
				for _, eesID := range rdr.EEsIDs {
					if !slices.Contains(exporterIDs, eesID) {
						return fmt.Errorf("<%s> exporter with id %s not defined", utils.ERs, eesID)
					}
				}
				for _, eesID := range rdr.EEsSuccessIDs {
					if !slices.Contains(exporterIDs, eesID) {
						return fmt.Errorf("<%s> exporter with id %s not defined", utils.ERs, eesID)
					}
				}
				for _, eesID := range rdr.EEsFailedIDs {
					if !slices.Contains(exporterIDs, eesID) {
						return fmt.Errorf("<%s> exporter with id %s not defined", utils.ERs, eesID)
					}
				}
			}
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
				if rdr.Opts.PartialOrderField != nil && *rdr.Opts.PartialOrderField == utils.EmptyString {
					return fmt.Errorf("<%s> empty %s for reader with ID: %s", utils.ERs, utils.PartialOrderFieldOpt, rdr.ID)
				}
			}
			if pAct == utils.MetaDumpToFile ||
				pAct == utils.MetaDumpToJSON { // only if the action is *dump_to_file
				path := rdr.ProcessedPath
				if rdr.Opts.PartialPath != nil {
					path = *rdr.Opts.PartialPath
				}
				if _, err := os.Stat(utils.IfaceAsString(path)); err != nil && os.IsNotExist(err) {
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
						if err := utils.IsPathValidForExporters(val.Path); err != nil {
							return fmt.Errorf("<%s> %s for %s at %s of %s", utils.ERs, err, val.Path, utils.Values, utils.CacheDumpFieldsCfg)
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
						if err := utils.IsPathValidForExporters(val.Path); err != nil {
							return fmt.Errorf("<%s> %s for %s at %s of %s", utils.ERs, err, val.Path, utils.Values, utils.FieldsCfg)
						}
					}
				}
				if err := utils.CheckInLineFilter(field.Filters); err != nil {
					return fmt.Errorf("<%s> %s for %s at %s", utils.ERs, err, field.Filters, utils.FieldsCfg)
				}
				// The following sanity check prevents a "slice bounds out of range" panic.
				if rdr.Type == utils.MetaFileXML && len(field.Value) != 0 &&
					!slices.Contains([]string{utils.MetaNone, utils.MetaConstant}, field.Type) {

					// Find the minimum rule length for dynamic utils.RSRParser within the field value.
					minRuleLength := math.MaxInt
					for _, parser := range field.Value {
						if !strings.HasPrefix(parser.Rules, utils.MetaDynReq) {
							continue
						}
						ruleLen := len(strings.Split(parser.Rules, utils.NestingSep))
						minRuleLength = min(minRuleLength, ruleLen)
					}

					// If a dynamic utils.RSRParser is found, verify xmlRootPath length against minRuleLength.
					if minRuleLength != math.MaxInt {
						var rootHP utils.HierarchyPath
						if rdr.Opts.XMLRootPath != nil {
							rootHP = utils.ParseHierarchyPath(*rdr.Opts.XMLRootPath, utils.EmptyString)
						}
						if len(rootHP) >= minRuleLength {
							return fmt.Errorf("<%s> %s for reader %s at %s",
								utils.ERs,
								"xmlRootPath length exceeds value rule elements",
								rdr.ID, field.Tag)
						}
					}
				}
			}
			if err := utils.CheckInLineFilter(rdr.Filters); err != nil {
				return fmt.Errorf("<%s> %s for %s at %s", utils.ERs, err, rdr.Filters, utils.ReadersCfg)
			}
		}
	}
	// EventExporter sanity checks
	if cfg.eesCfg.Enabled {
		for connType, opts := range cfg.eesCfg.Conns {
			for _, opt := range opts {
				for _, connID := range opt.Values {
					if connType == utils.MetaAttributes && strings.HasPrefix(connID, utils.MetaInternal) && !cfg.attributeSCfg.Enabled {
						return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.AttributeS, utils.EEs)
					}
					if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
						return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.EEs, connID)
					}
				}
			}
		}

		// Check cache TTL for file exporters which require positive TTL as
		// content is flushed only upon cache expiration.
		for eeType, cacheCfg := range cfg.eesCfg.Cache {
			if slices.Contains([]string{utils.MetaFileCSV, utils.MetaFileFWV}, eeType) {
				if cacheCfg.TTL <= 0 {
					return fmt.Errorf("<%s> exporter type %q requires positive cache TTL, got %v",
						utils.EEs, eeType, cacheCfg.TTL)
				}
			}
		}

		for _, exp := range cfg.eesCfg.Exporters {
			if !possibleExporterTypes.Has(exp.Type) {
				return fmt.Errorf("<%s> unsupported data type: %s for exporter with ID: %s", utils.EEs, exp.Type, exp.ID)
			}

			if exp.MetricsResetSchedule != "" {
				parser := cron.NewParser(
					cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
				)
				if _, err := parser.Parse(exp.MetricsResetSchedule); err != nil {
					return fmt.Errorf("<%s> invalid cron expression %q in metrics_reset_schedule for exporter %q: %v",
						utils.EEs, exp.MetricsResetSchedule, exp.ID, err)
				}
			}

			switch exp.Type {
			case utils.MetaFileCSV:
				for _, dir := range []string{exp.ExportPath} {
					if dir == utils.MetaBuffer {
						break
					}
					if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
						return fmt.Errorf("<%s> nonexistent folder: %s for exporter with ID: %s", utils.EEs, dir, exp.ID)
					}
				}
				if exp.Opts.CSVFieldSeparator != nil && *exp.Opts.CSVFieldSeparator == utils.EmptyString {
					return fmt.Errorf("<%s> empty %s for exporter with ID: %s", utils.EEs, utils.CSVFieldSepOpt, exp.ID)
				}
			case utils.MetaFileFWV:
				for _, dir := range []string{exp.ExportPath} {
					if dir == utils.MetaBuffer {
						break
					}
					if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
						return fmt.Errorf("<%s> nonexistent folder: %s for exporter with ID: %s", utils.EEs, dir, exp.ID)
					}
				}
			case utils.MetaElastic:
				opts := exp.Opts
				if opts.ElsLogger != nil {
					if !slices.Contains([]string{utils.ElsJson, utils.ElsText, utils.ElsColor}, *opts.ElsLogger) {
						return fmt.Errorf("<%s> invalid elsLogger value for exporter with ID: %s", utils.EEs, exp.ID)
					}
				}
				if opts.ElsRefresh != nil {
					if !slices.Contains([]string{"true", "false", "wait_for"}, *opts.ElsRefresh) {
						return fmt.Errorf("<%s> invalid elsRefresh value for exporter with ID: %s", utils.EEs, exp.ID)
					}
				}
				if opts.ElsOpType != nil {
					if !slices.Contains([]string{"index", "create"}, *opts.ElsOpType) {
						return fmt.Errorf("<%s> invalid elsOpType value for exporter with ID: %s", utils.EEs, exp.ID)
					}
				}
				if opts.ElsCAPath != nil {
					if _, err := os.Stat(*opts.ElsCAPath); os.IsNotExist(err) {
						return fmt.Errorf("<%s> CA certificate file not found at path: %s for exporter with ID: %s", utils.EEs, *opts.ElsCAPath, exp.ID)
					}
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
						if err := utils.IsPathValidForExporters(val.Path); err != nil {
							return fmt.Errorf("<%s> %s for %s at %s of %s", utils.EEs, err, val.Path, utils.Values, utils.FieldsCfg)
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

	// DataDB sanity checks
	hasOneInternalDB := false // used to reutrn error in case more then 1 internaldb is found
	for _, dbcfg := range cfg.dbCfg.DBConns {
		if dbcfg.Type == utils.MetaInternal {
			if hasOneInternalDB {
				return fmt.Errorf("<%s> There can only be 1 internal DB", utils.DB)
			}
			if (dbcfg.Opts.InternalDBDumpInterval != 0 ||
				dbcfg.Opts.InternalDBRewriteInterval != 0) &&
				dbcfg.Opts.InternalDBFileSizeLimit <= 0 {
				return fmt.Errorf("<%s> InternalDBFileSizeLimit field cannot be equal or smaller than 0: <%v>", utils.DB,
					dbcfg.Opts.InternalDBFileSizeLimit)
			}
			hasOneInternalDB = true
		}
		if dbcfg.Type == utils.MetaPostgres {
			if !slices.Contains([]string{utils.PgSSLModeDisable, utils.PgSSLModeAllow,
				utils.PgSSLModePrefer, utils.PgSSLModeRequire, utils.PgSSLModeVerifyCA,
				utils.PgSSLModeVerifyFull}, dbcfg.Opts.PgSSLMode) {
				return fmt.Errorf("<%s> unsupported pgSSLMode (sslmode) in DB configuration", utils.DB)
			}
			if !slices.Contains([]string{utils.PgSSLModeDisable, utils.PgSSLModeAllow, utils.PgSSLModeRequire,
				utils.EmptyString}, dbcfg.Opts.PgSSLCertMode) {
				return fmt.Errorf("<%s> unsupported pgSSLCertMode (sslcertmode) in DB configuration", utils.DB)
			}
		}
	}
	for item, val := range cfg.dbCfg.Items {
		if _, has := cfg.dbCfg.DBConns[val.DBConn]; !has {
			return fmt.Errorf("item's <%s> dbConn <%v>, does not match any db_conns ID", item, val.DBConn)
		}
		found1RmtConns := false
		found1RplConns := false
		for _, dbcfg := range cfg.dbCfg.DBConns {
			for _, connID := range dbcfg.RplConns {
				conn, has := cfg.rpcConns[connID]
				if !has {
					return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.DB, connID)
				}
				for _, rpc := range conn.Conns {
					if rpc.Transport != utils.MetaGOB {
						return fmt.Errorf("<%s> unsupported transport <%s> for connection with ID: <%s>", utils.DB, rpc.Transport, connID)
					}
				}
				found1RplConns = true
			}
			for _, connID := range dbcfg.RmtConns {
				conn, has := cfg.rpcConns[connID]
				if !has {
					return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.DB, connID)
				}
				for _, rpc := range conn.Conns {
					if rpc.Transport != utils.MetaGOB {
						return fmt.Errorf("<%s> unsupported transport <%s> for connection with ID: <%s>", utils.DB, rpc.Transport, connID)
					}
				}
				found1RmtConns = true
			}
		}
		if val.Remote && !found1RmtConns {
			return fmt.Errorf("remote connections required by: <%s>", item)
		}
		if val.Replicate && !found1RplConns {
			return fmt.Errorf("replicate connections required by: <%s>", item)
		}
	}
	// APIer sanity checks
	admConnEnabledMap := map[string]struct {
		name    string
		enabled bool
	}{
		utils.MetaAttributes: {utils.AttributeS, cfg.attributeSCfg.Enabled},
		utils.MetaActions:    {utils.SchedulerS, cfg.actionSCfg.Enabled},
	}
	for connType, opts := range cfg.admS.Conns {
		for _, opt := range opts {
			for _, connID := range opt.Values {
				if info, has := admConnEnabledMap[connType]; has {
					if strings.HasPrefix(connID, utils.MetaInternal) && !info.enabled {
						return fmt.Errorf("<%s> not enabled but requested by <%s> component", info.name, utils.AdminS)
					}
				}
				if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
					return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.AdminS, connID)
				}
			}
		}
	}
	// Cache check
	for _, connID := range cfg.cacheCfg.RemoteConns {
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
	fltrConnEnabledMap := map[string]struct {
		name    string
		enabled bool
	}{
		utils.MetaStats:     {utils.StatS, cfg.statsCfg.Enabled},
		utils.MetaResources: {utils.ResourceS, cfg.resourceSCfg.Enabled},
		utils.MetaAccounts:  {utils.AccountS, cfg.accountSCfg.Enabled},
	}
	for connType, opts := range cfg.filterSCfg.Conns {
		for _, opt := range opts {
			for _, connID := range opt.Values {
				if info, has := fltrConnEnabledMap[connType]; has {
					if strings.HasPrefix(connID, utils.MetaInternal) && !info.enabled {
						return fmt.Errorf("<%s> not enabled but requested by <%s> component", info.name, utils.FilterS)
					}
				}
				if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
					return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.FilterS, connID)
				}
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
				if !slices.Contains([]string{utils.MetaGOB, rpcclient.HTTPjson, utils.MetaJSON, rpcclient.BiRPCJSON, rpcclient.BiRPCGOB}, host.Transport) {
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
		// TTL and CleanupInterval should allways be biger than zero in order to not keep unecesary logs in index
		if cfg.analyzerSCfg.TTL <= 0 {
			return fmt.Errorf("<%s> the TTL needs to be bigger than 0", utils.AnalyzerS)
		}
		for _, connID := range cfg.analyzerSCfg.EEsConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.eesCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.EEs, utils.AnalyzerS)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.AnalyzerS, connID)
			}
		}
		if cfg.analyzerSCfg.CleanupInterval <= 0 {
			return fmt.Errorf("<%s> the CleanupInterval needs to be bigger than 0", utils.AnalyzerS)
		}
	}
	if cfg.prometheusAgentCfg.Enabled {
		if len(cfg.prometheusAgentCfg.Conns[utils.MetaStats]) > 0 &&
			len(cfg.prometheusAgentCfg.StatQueueIDs) == 0 &&
			len(cfg.prometheusAgentCfg.Conns[utils.MetaStats]) != len(cfg.prometheusAgentCfg.Conns[utils.MetaAdminS]) {
			return fmt.Errorf(
				"<%s> when StatQueueIDs is empty, admins_conns must match stats_conns length to fetch StatQueue IDs",
				utils.PrometheusAgent)
		}
		promConnEnabledMap := map[string]struct {
			name    string
			enabled bool
		}{
			utils.MetaAdminS: {utils.AdminS, cfg.admS.Enabled},
			utils.MetaStats:  {utils.StatService, cfg.statsCfg.Enabled},
		}
		for connType, opts := range cfg.prometheusAgentCfg.Conns {
			for _, opt := range opts {
				for _, connID := range opt.Values {
					if info, has := promConnEnabledMap[connType]; has && strings.HasPrefix(connID, utils.MetaInternal) && !info.enabled {
						return fmt.Errorf("<%s> not enabled but requested by <%s> component", info.name, utils.PrometheusAgent)
					}
					if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
						return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.PrometheusAgent, connID)
					}
				}
			}
		}
		if len(cfg.prometheusAgentCfg.Conns[utils.MetaCores]) > 0 {
			if cfg.prometheusAgentCfg.CollectGoMetrics || cfg.prometheusAgentCfg.CollectProcessMetrics {
				return fmt.Errorf("<%s> collect_go_metrics and collect_process_metrics cannot be enabled when using CoreSConns",
					utils.PrometheusAgent)
			}
		}
	}

	return nil
}

// hasInternalConnOpt checks if any DynamicStringSliceOpt contains a *internal connection ID
func hasInternalConnOpt(opts []*DynamicStringSliceOpt) bool {
	for _, opt := range opts {
		for _, connID := range opt.Values {
			if strings.HasPrefix(connID, utils.MetaInternal) {
				return true
			}
		}
	}
	return false
}
