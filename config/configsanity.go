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
	"errors"
	"fmt"
	"os"

	"github.com/cgrates/cgrates/utils"
)

func (cfg *CGRConfig) checkConfigSanity() error {
	// Rater checks
	if cfg.ralsCfg.Enabled && !cfg.dispatcherSCfg.Enabled {
		if !cfg.statsCfg.Enabled {
			for _, connCfg := range cfg.ralsCfg.StatSConns {
				if connCfg.Address == utils.MetaInternal {
					return fmt.Errorf("<%s> not enabled but requested by <%s> component.",
						utils.StatS, utils.RALService)
				}
			}
		}
		if !cfg.thresholdSCfg.Enabled {
			for _, connCfg := range cfg.ralsCfg.ThresholdSConns {
				if connCfg.Address == utils.MetaInternal {
					return fmt.Errorf("<%s> not enabled but requested by <%s> component.",
						utils.ThresholdS, utils.RALService)
				}
			}
		}
	}
	// CDRServer checks
	if cfg.cdrsCfg.Enabled && !cfg.dispatcherSCfg.Enabled {
		if !cfg.chargerSCfg.Enabled {
			for _, conn := range cfg.cdrsCfg.ChargerSConns {
				if conn.Address == utils.MetaInternal {
					return fmt.Errorf("<%s> not enabled but requested by <%s> component.", utils.Chargers, utils.CDRs)
				}
			}
		}
		if !cfg.ralsCfg.Enabled {
			for _, cdrsRaterConn := range cfg.cdrsCfg.RaterConns {
				if cdrsRaterConn.Address == utils.MetaInternal {
					return fmt.Errorf("<%s> not enabled but requested by <%s> component.", utils.RALService, utils.CDRs)
				}
			}
		}
		if !cfg.attributeSCfg.Enabled {
			for _, connCfg := range cfg.cdrsCfg.AttributeSConns {
				if connCfg.Address == utils.MetaInternal {
					return fmt.Errorf("<%s> not enabled but requested by <%s> component.", utils.AttributeS, utils.CDRs)
				}
			}
		}
		if !cfg.statsCfg.Enabled {
			for _, connCfg := range cfg.cdrsCfg.StatSConns {
				if connCfg.Address == utils.MetaInternal {
					return fmt.Errorf("<%s> not enabled but requested by <%s> component.", utils.StatService, utils.CDRs)
				}
			}
		}
		for _, cdrePrfl := range cfg.cdrsCfg.OnlineCDRExports {
			if _, hasIt := cfg.CdreProfiles[cdrePrfl]; !hasIt {
				return fmt.Errorf("<%s> Cannot find CDR export template with ID: <%s>", utils.CDRs, cdrePrfl)
			}
		}
		if !cfg.thresholdSCfg.Enabled {
			for _, connCfg := range cfg.cdrsCfg.ThresholdSConns {
				if connCfg.Address == utils.MetaInternal {
					return fmt.Errorf("%s not enabled but requested by %s component.", utils.ThresholdS, utils.CDRs)
				}
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
			if !cfg.cdrsCfg.Enabled && !cfg.dispatcherSCfg.Enabled {
				for _, conn := range cdrcInst.CdrsConns {
					if conn.Address == utils.MetaInternal {
						return fmt.Errorf("<%s> not enabled but referenced from <%s>", utils.CDRs, utils.CDRC)
					}
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
	if cfg.sessionSCfg.Enabled && !cfg.dispatcherSCfg.Enabled {
		if cfg.sessionSCfg.TerminateAttempts < 1 {
			return fmt.Errorf("<%s> 'terminate_attempts' should be at least 1", utils.SessionS)
		}
		if !cfg.chargerSCfg.Enabled {
			for _, conn := range cfg.sessionSCfg.ChargerSConns {
				if conn.Address == utils.MetaInternal {
					return fmt.Errorf("<%s> %s not enabled", utils.SessionS, utils.ChargerS)
				}
			}
		}
		if !cfg.ralsCfg.Enabled {
			for _, smgRALsConn := range cfg.sessionSCfg.RALsConns {
				if smgRALsConn.Address == utils.MetaInternal {
					return fmt.Errorf("<%s> %s not enabled but requested by SMGeneric component.", utils.SessionS, utils.RALService)
				}
			}
		}
		if !cfg.resourceSCfg.Enabled {
			for _, conn := range cfg.sessionSCfg.ResSConns {
				if conn.Address == utils.MetaInternal {
					return fmt.Errorf("<%s> %s not enabled but requested by SMGeneric component.", utils.SessionS, utils.ResourceS)
				}
			}
		}
		if !cfg.thresholdSCfg.Enabled {
			for _, conn := range cfg.sessionSCfg.ThreshSConns {
				if conn.Address == utils.MetaInternal {
					return fmt.Errorf("<%s> %s not enabled but requested by SMGeneric component.", utils.SessionS, utils.ThresholdS)
				}
			}
		}
		if !cfg.statsCfg.Enabled {
			for _, conn := range cfg.sessionSCfg.StatSConns {
				if conn.Address == utils.MetaInternal {
					return fmt.Errorf("<%s> %s not enabled but requested by SMGeneric component.", utils.SessionS, utils.StatS)
				}
			}
		}
		if !cfg.supplierSCfg.Enabled {
			for _, conn := range cfg.sessionSCfg.SupplSConns {
				if conn.Address == utils.MetaInternal {
					return fmt.Errorf("<%s> %s not enabled but requested by SMGeneric component.", utils.SessionS, utils.SupplierS)
				}
			}
		}
		if !cfg.attributeSCfg.Enabled {
			for _, conn := range cfg.sessionSCfg.AttrSConns {
				if conn.Address == utils.MetaInternal {
					return fmt.Errorf("<%s> %s not enabled but requested by SMGeneric component.", utils.SessionS, utils.AttributeS)
				}
			}
		}
		if !cfg.cdrsCfg.Enabled {
			for _, smgCDRSConn := range cfg.sessionSCfg.CDRsConns {
				if smgCDRSConn.Address == utils.MetaInternal {
					return fmt.Errorf("<%s> CDRS not enabled but referenced by SMGeneric component", utils.SessionS)
				}
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
		if !cfg.dispatcherSCfg.Enabled && // if dispatcher is enabled all internal connections are managed by it
			!cfg.sessionSCfg.Enabled {
			for _, connCfg := range cfg.fsAgentCfg.SessionSConns {
				if connCfg.Address == utils.MetaInternal {
					return fmt.Errorf("%s not enabled but referenced by %s",
						utils.SessionS, utils.FreeSWITCHAgent)
				}
			}
		}
	}
	// KamailioAgent checks
	if cfg.kamAgentCfg.Enabled {
		if len(cfg.kamAgentCfg.SessionSConns) == 0 {
			return fmt.Errorf("<%s> no %s connections defined",
				utils.KamailioAgent, utils.SessionS)
		}
		if !cfg.dispatcherSCfg.Enabled && // if dispatcher is enabled all internal connections are managed by it
			!cfg.sessionSCfg.Enabled {
			for _, connCfg := range cfg.kamAgentCfg.SessionSConns {
				if connCfg.Address == utils.MetaInternal {
					return fmt.Errorf("%s not enabled but referenced by %s",
						utils.SessionS, utils.KamailioAgent)
				}
			}
		}
	}
	// AsteriskAgent checks
	if cfg.asteriskAgentCfg.Enabled {
		if len(cfg.asteriskAgentCfg.SessionSConns) == 0 {
			return fmt.Errorf("<%s> no %s connections defined",
				utils.AsteriskAgent, utils.SessionS)
		}
		if !cfg.dispatcherSCfg.Enabled && // if dispatcher is enabled all internal connections are managed by it
			!cfg.sessionSCfg.Enabled {
			for _, smAstSMGConn := range cfg.asteriskAgentCfg.SessionSConns {
				if smAstSMGConn.Address == utils.MetaInternal {
					return fmt.Errorf("%s not enabled but referenced by %s",
						utils.SessionS, utils.AsteriskAgent)
				}
			}
		}
	}
	// DAgent checks
	if cfg.diameterAgentCfg.Enabled {
		if len(cfg.diameterAgentCfg.SessionSConns) == 0 {
			return fmt.Errorf("<%s> no %s connections defined",
				utils.DiameterAgent, utils.SessionS)
		}
		if !cfg.dispatcherSCfg.Enabled && // if dispatcher is enabled all internal connections are managed by it
			!cfg.sessionSCfg.Enabled {
			for _, daSMGConn := range cfg.diameterAgentCfg.SessionSConns {
				if daSMGConn.Address == utils.MetaInternal {
					return fmt.Errorf("%s not enabled but referenced by %s",
						utils.SessionS, utils.DiameterAgent)
				}
			}
		}
	}
	if cfg.radiusAgentCfg.Enabled {
		if len(cfg.radiusAgentCfg.SessionSConns) == 0 {
			return fmt.Errorf("<%s> no %s connections defined",
				utils.RadiusAgent, utils.SessionS)
		}
		if !cfg.dispatcherSCfg.Enabled && // if dispatcher is enabled all internal connections are managed by it
			!cfg.sessionSCfg.Enabled {
			for _, raSMGConn := range cfg.radiusAgentCfg.SessionSConns {
				if raSMGConn.Address == utils.MetaInternal {
					return fmt.Errorf("%s not enabled but referenced by %s",
						utils.SessionS, utils.RadiusAgent)
				}
			}
		}
	}
	if cfg.dnsAgentCfg.Enabled {
		if len(cfg.dnsAgentCfg.SessionSConns) == 0 {
			return fmt.Errorf("<%s> no %s connections defined",
				utils.DNSAgent, utils.SessionS)
		}
		if !cfg.dispatcherSCfg.Enabled && // if dispatcher is enabled all internal connections are managed by it
			!cfg.sessionSCfg.Enabled {
			for _, sSConn := range cfg.dnsAgentCfg.SessionSConns {
				if sSConn.Address == utils.MetaInternal {
					return fmt.Errorf("%s not enabled but referenced by %s", utils.SessionS, utils.DNSAgent)
				}
			}
		}
	}
	// HTTPAgent checks
	for _, httpAgentCfg := range cfg.httpAgentCfg {
		// httpAgent checks
		if !cfg.dispatcherSCfg.Enabled && // if dispatcher is enabled all internal connections are managed by it
			cfg.sessionSCfg.Enabled {
			for _, sSConn := range httpAgentCfg.SessionSConns {
				if sSConn.Address == utils.MetaInternal {
					return errors.New("SessionS not enabled but referenced by HttpAgent component")
				}
			}
		}
		if !utils.SliceHasMember([]string{utils.MetaUrl, utils.MetaXml}, httpAgentCfg.RequestPayload) {
			return fmt.Errorf("<%s> unsupported request payload %s",
				utils.HTTPAgent, httpAgentCfg.RequestPayload)
		}
		if !utils.SliceHasMember([]string{utils.MetaTextPlain, utils.MetaXml}, httpAgentCfg.ReplyPayload) {
			return fmt.Errorf("<%s> unsupported reply payload %s",
				utils.HTTPAgent, httpAgentCfg.ReplyPayload)
		}
	}
	if cfg.attributeSCfg.Enabled {
		if cfg.attributeSCfg.ProcessRuns < 1 {
			return fmt.Errorf("<%s> process_runs needs to be bigger than 0", utils.AttributeS)
		}
	}
	if cfg.chargerSCfg.Enabled && !cfg.dispatcherSCfg.Enabled &&
		(cfg.attributeSCfg == nil || !cfg.attributeSCfg.Enabled) {
		for _, connCfg := range cfg.chargerSCfg.AttributeSConns {
			if connCfg.Address == utils.MetaInternal {
				return errors.New("AttributeS not enabled but requested by ChargerS component.")
			}
		}
	}
	// ResourceLimiter checks
	if cfg.resourceSCfg.Enabled && !cfg.thresholdSCfg.Enabled && !cfg.dispatcherSCfg.Enabled {
		for _, connCfg := range cfg.resourceSCfg.ThresholdSConns {
			if connCfg.Address == utils.MetaInternal {
				return errors.New("ThresholdS not enabled but requested by ResourceS component.")
			}
		}
	}
	// StatS checks
	if cfg.statsCfg.Enabled && !cfg.thresholdSCfg.Enabled && !cfg.dispatcherSCfg.Enabled {
		for _, connCfg := range cfg.statsCfg.ThresholdSConns {
			if connCfg.Address == utils.MetaInternal {
				return errors.New("ThresholdS not enabled but requested by StatS component.")
			}
		}
	}
	// SupplierS checks
	if cfg.supplierSCfg.Enabled && !cfg.dispatcherSCfg.Enabled {
		if !cfg.resourceSCfg.Enabled {
			for _, connCfg := range cfg.supplierSCfg.ResourceSConns {
				if connCfg.Address == utils.MetaInternal {
					return errors.New("ResourceS not enabled but requested by SupplierS component.")
				}
			}
		}
		if !cfg.statsCfg.Enabled {
			for _, connCfg := range cfg.supplierSCfg.StatSConns {
				if connCfg.Address == utils.MetaInternal {
					return errors.New("StatS not enabled but requested by SupplierS component.")
				}
			}
		}
		if !cfg.attributeSCfg.Enabled {
			for _, connCfg := range cfg.supplierSCfg.AttributeSConns {
				if connCfg.Address == utils.MetaInternal {
					return errors.New("AttributeS not enabled but requested by SupplierS component.")
				}
			}
		}
	}
	// Scheduler check connection with CDR Server
	if !cfg.cdrsCfg.Enabled && !cfg.dispatcherSCfg.Enabled {
		for _, connCfg := range cfg.schedulerCfg.CDRsConns {
			if connCfg.Address == utils.MetaInternal {
				return errors.New("CDR Server not enabled but requested by Scheduler")
			}
		}
	}
	// EventReader sanity checks
	if cfg.ersCfg.Enabled {
		if !cfg.sessionSCfg.Enabled {
			for _, connCfg := range cfg.ersCfg.SessionSConns {
				if connCfg.Address == utils.MetaInternal {
					return errors.New("SessionS not enabled but requested by EventReader component.")
				}
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
			if key == utils.CacheDiameterMessages || key == utils.CacheClosedSessions {
				if config.Limit == 0 {
					return fmt.Errorf("<%s> %s needs to be != 0 when DataBD is *internal, found 0.", utils.CacheS, key)
				}
				continue
			}
			if config.Limit != 0 {
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
	return nil
}
