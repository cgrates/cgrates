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
	"strings"

	"github.com/cgrates/cgrates/utils"
)

func (cfg *CGRConfig) checkConfigSanity() error {
	// Rater checks
	if cfg.ralsCfg.Enabled {
		for _, connID := range cfg.ralsCfg.StatSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.statsCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component.", utils.StatService, utils.RALService)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> Connection with id: <%s> not defined", utils.RALService, connID)
			}
		}
		for _, connID := range cfg.ralsCfg.ThresholdSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.thresholdSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component.", utils.ThresholdS, utils.RALService)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> Connection with id: <%s> not defined", utils.RALService, connID)
			}
		}
	}
	// CDRServer checks
	if cfg.cdrsCfg.Enabled {
		for _, connID := range cfg.cdrsCfg.ChargerSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.chargerSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component.", utils.ChargerS, utils.CDRs)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> Connection with id: <%s> not defined", utils.CDRs, connID)
			}
		}
		for _, connID := range cfg.cdrsCfg.RaterConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.ralsCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component.", utils.RALService, utils.CDRs)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> Connection with id: <%s> not defined", utils.CDRs, connID)
			}
		}
		for _, connID := range cfg.cdrsCfg.AttributeSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.attributeSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component.", utils.AttributeS, utils.CDRs)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> Connection with id: <%s> not defined", utils.CDRs, connID)
			}
		}
		for _, connID := range cfg.cdrsCfg.StatSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.statsCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component.", utils.StatService, utils.CDRs)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> Connection with id: <%s> not defined", utils.CDRs, connID)
			}
		}
		for _, connID := range cfg.cdrsCfg.ThresholdSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.thresholdSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component.", utils.ThresholdS, utils.CDRs)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> Connection with id: <%s> not defined", utils.CDRs, connID)
			}
		}
		for _, cdrePrfl := range cfg.cdrsCfg.OnlineCDRExports {
			if _, hasIt := cfg.CdreProfiles[cdrePrfl]; !hasIt {
				return fmt.Errorf("<%s> Cannot find CDR export template with ID: <%s>", utils.CDRs, cdrePrfl)
			}
		}
	}
	// CDRC sanity checks
	for _, cdrcCfgs := range cfg.CdrcProfiles {
		for _, cdrcInst := range cdrcCfgs {
			if !cdrcInst.Enabled {
				continue
			}
			if len(cdrcInst.CdrsConns) == 0 {
				return fmt.Errorf("<%s> Instance: %s, %s enabled but no %s defined!", utils.CDRC, cdrcInst.ID, utils.CDRC, utils.CDRs)
			}
			for _, connID := range cdrcInst.CdrsConns {
				if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.cdrsCfg.Enabled {
					return fmt.Errorf("<%s> not enabled but requested by <%s> cdrcProfile", utils.CDRs, cdrcInst.ID)
				}
				if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
					return fmt.Errorf("<%s> Connection with id: <%s> not defined", utils.CDRs, connID)
				}
			}
			if len(cdrcInst.ContentFields) == 0 {
				return fmt.Errorf("<%s> enabled but no fields to be processed defined!", utils.CDRC)
			}
		}
	}
	// Loaders sanity checks
	for _, ldrSCfg := range cfg.loaderCfg {
		if !ldrSCfg.Enabled {
			continue
		}
		for _, dir := range []string{ldrSCfg.TpInDir, ldrSCfg.TpOutDir} {
			if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
				return fmt.Errorf("<%s> Nonexistent folder: %s", utils.LoaderS, dir)
			}
		}
		for _, data := range ldrSCfg.Data {
			if !posibleLoaderTypes.Has(data.Type) {
				return fmt.Errorf("<%s> unsupported data type %s", utils.LoaderS, data.Type)
			}

			for _, field := range data.Fields {
				if field.Type != utils.META_COMPOSED && field.Type != utils.MetaString && field.Type != utils.MetaVariable {
					return fmt.Errorf("<%s> invalid field type %s for %s at %s", utils.LoaderS, field.Type, data.Type, field.Tag)
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
				return fmt.Errorf("<%s> not enabled but requested by <%s> component.", utils.ChargerS, utils.SessionS)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> Connection with id: <%s> not defined", utils.SessionS, connID)
			}
		}
		for _, connID := range cfg.sessionSCfg.RALsConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.ralsCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component.", utils.RALService, utils.SessionS)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> Connection with id: <%s> not defined", utils.SessionS, connID)
			}
		}
		for _, connID := range cfg.sessionSCfg.ResSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.resourceSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component.", utils.ResourceS, utils.SessionS)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> Connection with id: <%s> not defined", utils.SessionS, connID)
			}
		}
		for _, connID := range cfg.sessionSCfg.ThreshSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.thresholdSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component.", utils.ThresholdS, utils.SessionS)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> Connection with id: <%s> not defined", utils.SessionS, connID)
			}
		}
		for _, connID := range cfg.sessionSCfg.StatSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.statsCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component.", utils.StatService, utils.SessionS)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> Connection with id: <%s> not defined", utils.SessionS, connID)
			}
		}
		for _, connID := range cfg.sessionSCfg.SupplSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.supplierSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component.", utils.SupplierS, utils.SessionS)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> Connection with id: <%s> not defined", utils.SessionS, connID)
			}
		}
		for _, connID := range cfg.sessionSCfg.AttrSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.attributeSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component.", utils.AttributeS, utils.SessionS)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> Connection with id: <%s> not defined", utils.SessionS, connID)
			}
		}
		for _, connID := range cfg.sessionSCfg.CDRsConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.cdrsCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component.", utils.CDRs, utils.SessionS)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> Connection with id: <%s> not defined", utils.SessionS, connID)
			}
		}
		for _, connID := range cfg.sessionSCfg.ReplicationConns {
			if _, has := cfg.rpcConns[connID]; !has {
				return fmt.Errorf("<%s> Connection with id: <%s> not defined", utils.SessionS, connID)
			}
		}
		if cfg.cacheCfg[utils.CacheClosedSessions].Limit == 0 {
			return fmt.Errorf("<%s> %s needs to be != 0, received: %d", utils.CacheS, utils.CacheClosedSessions, cfg.cacheCfg[utils.CacheClosedSessions].Limit)
		}
	}

	// FreeSWITCHAgent checks
	if cfg.fsAgentCfg.Enabled {
		if len(cfg.fsAgentCfg.SessionSConns) == 0 {
			return fmt.Errorf("<%s> no %s connections defined",
				utils.FreeSWITCHAgent, utils.SessionS)
		}
		for _, connID := range cfg.fsAgentCfg.SessionSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.sessionSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component.", utils.SessionS, utils.FreeSWITCHAgent)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> Connection with id: <%s> not defined", utils.FreeSWITCHAgent, connID)
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
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.sessionSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component.", utils.SessionS, utils.KamailioAgent)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> Connection with id: <%s> not defined", utils.KamailioAgent, connID)
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
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.sessionSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component.", utils.SessionS, utils.AsteriskAgent)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> Connection with id: <%s> not defined", utils.AsteriskAgent, connID)
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
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.sessionSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component.", utils.SessionS, utils.DiameterAgent)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> Connection with id: <%s> not defined", utils.DiameterAgent, connID)
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
				return fmt.Errorf("<%s> not enabled but requested by <%s> component.", utils.SessionS, utils.RadiusAgent)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> Connection with id: <%s> not defined", utils.RadiusAgent, connID)
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
				return fmt.Errorf("<%s> not enabled but requested by <%s> component.", utils.SessionS, utils.DNSAgent)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> Connection with id: <%s> not defined", utils.DNSAgent, connID)
			}
		}
	}
	// HTTPAgent checks
	for _, httpAgentCfg := range cfg.httpAgentCfg {
		// httpAgent checks
		for _, connID := range httpAgentCfg.SessionSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.sessionSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> HTTPAgent Template.", utils.SessionS, httpAgentCfg.ID)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("HTTPAgent Templae with ID <%s> has connection with id: <%s> not defined", httpAgentCfg.ID, connID)
			}
		}
		if !utils.SliceHasMember([]string{utils.MetaUrl, utils.MetaXml}, httpAgentCfg.RequestPayload) {
			return fmt.Errorf("<%s> unsupported request payload %s", utils.HTTPAgent, httpAgentCfg.RequestPayload)
		}
		if !utils.SliceHasMember([]string{utils.MetaTextPlain, utils.MetaXml}, httpAgentCfg.ReplyPayload) {
			return fmt.Errorf("<%s> unsupported reply payload %s", utils.HTTPAgent, httpAgentCfg.ReplyPayload)
		}
	}
	if cfg.attributeSCfg.Enabled {
		if cfg.attributeSCfg.ProcessRuns < 1 {
			return fmt.Errorf("<%s> process_runs needs to be bigger than 0", utils.AttributeS)
		}
	}
	if cfg.chargerSCfg.Enabled {
		for _, connID := range cfg.chargerSCfg.AttributeSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.attributeSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component.", utils.AttributeS, utils.ChargerS)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> Connection with id: <%s> not defined", utils.ChargerS, connID)
			}
		}
	}
	// ResourceLimiter checks
	if cfg.resourceSCfg.Enabled {
		for _, connID := range cfg.resourceSCfg.ThresholdSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.thresholdSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component.", utils.ThresholdS, utils.ResourceS)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> Connection with id: <%s> not defined", utils.ResourceS, connID)
			}
		}
	}
	// StatS checks
	if cfg.statsCfg.Enabled {
		for _, connID := range cfg.statsCfg.ThresholdSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.thresholdSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component.", utils.ThresholdS, utils.StatS)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> Connection with id: <%s> not defined", utils.StatS, connID)
			}
		}
	}
	// SupplierS checks
	if cfg.supplierSCfg.Enabled {
		for _, connID := range cfg.supplierSCfg.AttributeSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.attributeSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component.", utils.AttributeS, utils.SupplierS)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> Connection with id: <%s> not defined", utils.SupplierS, connID)
			}
		}
		for _, connID := range cfg.supplierSCfg.StatSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.statsCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component.", utils.StatService, utils.SupplierS)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> Connection with id: <%s> not defined", utils.SupplierS, connID)
			}
		}
		for _, connID := range cfg.supplierSCfg.ResourceSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.resourceSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component.", utils.ResourceS, utils.SupplierS)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> Connection with id: <%s> not defined", utils.SupplierS, connID)
			}
		}
	}
	// Scheduler check connection with CDR Server
	if cfg.schedulerCfg.Enabled {
		for _, connID := range cfg.schedulerCfg.CDRsConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.cdrsCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component.", utils.CDRs, utils.SchedulerS)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> Connection with id: <%s> not defined", utils.SchedulerS, connID)
			}
		}
	}
	// EventReader sanity checks
	if cfg.ersCfg.Enabled {
		for _, connID := range cfg.ersCfg.SessionSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.sessionSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component.", utils.SessionS, utils.ERs)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> Connection with id: <%s> not defined", utils.ERs, connID)
			}
		}
		for _, rdr := range cfg.ersCfg.Readers {
			if !possibleReaderTypes.Has(rdr.Type) {
				return fmt.Errorf("<%s> unsupported data type: %s for reader with ID: %s", utils.ERs, rdr.Type, rdr.ID)
			}

			if rdr.Type == utils.MetaFileCSV {
				for _, dir := range []string{rdr.ProcessedPath, rdr.SourcePath} {
					if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
						return fmt.Errorf("<%s> Nonexistent folder: %s for reader with ID: %s", utils.ERs, dir, rdr.ID)
					}
				}
				if rdr.FieldSep == utils.EmptyString {
					return fmt.Errorf("<%s> empty FieldSep for reader with ID: %s", utils.ERs, rdr.ID)
				}
			}
			if rdr.Type == utils.MetaKafkajsonMap && rdr.RunDelay > 0 {
				return fmt.Errorf("<%s> RunDelay field can not be bigger than zero for reader with ID: %s", utils.ERs, rdr.ID)
			}
		}
	}
	// StorDB sanity checks
	if cfg.storDbCfg.Type == utils.POSTGRES {
		if !utils.IsSliceMember([]string{utils.PostgressSSLModeDisable, utils.PostgressSSLModeAllow,
			utils.PostgressSSLModePrefer, utils.PostgressSSLModeRequire, utils.PostgressSSLModeVerifyCa,
			utils.PostgressSSLModeVerifyFull}, cfg.storDbCfg.SSLMode) {
			return fmt.Errorf("<%s> Unsuported sslmode for storDB", utils.StorDB)
		}
	}
	// DataDB sanity checks
	if cfg.dataDbCfg.DataDbType == utils.INTERNAL {
		for key, config := range cfg.cacheCfg {
			if utils.CacheDataDBPartitions.Has(key) && config.Limit != 0 {
				return fmt.Errorf("<%s> %s needs to be 0 when DataBD is *internal, received : %d", utils.CacheS, key, config.Limit)
			}
		}
		if cfg.resourceSCfg.Enabled == true && cfg.resourceSCfg.StoreInterval != -1 {
			return fmt.Errorf("<%s> StoreInterval needs to be -1 when DataBD is *internal, received : %d", utils.ResourceS, cfg.resourceSCfg.StoreInterval)
		}
		if cfg.statsCfg.Enabled == true && cfg.statsCfg.StoreInterval != -1 {
			return fmt.Errorf("<%s> StoreInterval needs to be -1 when DataBD is *internal, received : %d", utils.StatS, cfg.statsCfg.StoreInterval)
		}
		if cfg.thresholdSCfg.Enabled == true && cfg.thresholdSCfg.StoreInterval != -1 {
			return fmt.Errorf("<%s> StoreInterval needs to be -1 when DataBD is *internal, received : %d", utils.ThresholdS, cfg.thresholdSCfg.StoreInterval)
		}
	}
	for item, val := range cfg.dataDbCfg.Items {
		if val.Remote == true && len(cfg.dataDbCfg.RmtConns) == 0 {
			return fmt.Errorf("Remote connections required by: <%s>", item)
		}
		if val.Replicate == true && len(cfg.dataDbCfg.RplConns) == 0 {
			return fmt.Errorf("Replicate connections required by: <%s>", item)
		}
	}
	// APIer sanity checks
	for _, connID := range cfg.apier.AttributeSConns {
		if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.attributeSCfg.Enabled {
			return fmt.Errorf("<%s> not enabled but requested by <%s> component.", utils.AttributeS, utils.ApierV1)
		}
		if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
			return fmt.Errorf("<%s> Connection with id: <%s> not defined", utils.ApierV1, connID)
		}
	}
	for _, connID := range cfg.apier.SchedulerConns {
		if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.schedulerCfg.Enabled {
			return fmt.Errorf("<%s> not enabled but requested by <%s> component.", utils.SchedulerS, utils.ApierV1)
		}
		if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
			return fmt.Errorf("<%s> Connection with id: <%s> not defined", utils.ApierV1, connID)
		}
	}
	// Dispatcher sanity check
	if cfg.dispatcherSCfg.Enabled {
		for _, connID := range cfg.dispatcherSCfg.AttributeSConns {
			if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.attributeSCfg.Enabled {
				return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.AttributeS, utils.DispatcherS)
			}
			if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
				return fmt.Errorf("<%s> Connection with id: <%s> not defined", utils.DispatcherS, connID)
			}
		}
	}
	// Cache partitions check
	for cacheID := range cfg.cacheCfg {
		if !utils.CachePartitions.Has(cacheID) {
			return fmt.Errorf("<%s> partition <%s> not defined", utils.CacheS, cacheID)
		}
	}
	// FilterS sanity check
	for _, connID := range cfg.filterSCfg.StatSConns {
		if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.statsCfg.Enabled {
			return fmt.Errorf("<%s> not enabled but requested by <%s> component.", utils.StatS, utils.FilterS)
		}
		if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
			return fmt.Errorf("<%s> Connection with id: <%s> not defined", utils.FilterS, connID)
		}
	}
	for _, connID := range cfg.filterSCfg.ResourceSConns {
		if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.resourceSCfg.Enabled {
			return fmt.Errorf("<%s> not enabled but requested by <%s> component.", utils.ResourceS, utils.FilterS)
		}
		if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
			return fmt.Errorf("<%s> Connection with id: <%s> not defined", utils.FilterS, connID)
		}
	}
	for _, connID := range cfg.filterSCfg.RALsConns {
		if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.ralsCfg.Enabled {
			return fmt.Errorf("<%s> not enabled but requested by <%s> component.", utils.RALService, utils.FilterS)
		}
		if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
			return fmt.Errorf("<%s> Connection with id: <%s> not defined", utils.FilterS, connID)
		}
	}

	return nil
}
