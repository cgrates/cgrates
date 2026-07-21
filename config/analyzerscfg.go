// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"maps"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// AnalyzerSCfg is the configuration of analyzer service
type AnalyzerSCfg struct {
	Enabled         bool
	DBPath          string
	IndexType       string
	TTL             time.Duration
	Conns           map[string][]*DynamicConns
	CleanupInterval time.Duration
	Opts            *AnalyzerSOpts
}

type AnalyzerSOpts struct {
	ExporterIDs []*DynamicStringSliceOpt
}

// loadAnalyzerCgrCfg loads the Analyzer section of the configuration
func (alS *AnalyzerSCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnAnalyzerCgrCfg := new(AnalyzerSJsonCfg)
	if err = jsnCfg.GetSection(ctx, AnalyzerSJSON, jsnAnalyzerCgrCfg); err != nil {
		return
	}
	return alS.loadFromJSONCfg(jsnAnalyzerCgrCfg)
}

func (anzOpts *AnalyzerSOpts) loadFromJSONCfg(jsonAnzOpts *AnalyzerSOptsJson) {
	if jsonAnzOpts == nil {
		return
	}
	if jsonAnzOpts.ExporterIDs != nil {
		anzOpts.ExporterIDs = append(anzOpts.ExporterIDs, jsonAnzOpts.ExporterIDs...)
	}
}

func (alS *AnalyzerSCfg) loadFromJSONCfg(jsnCfg *AnalyzerSJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		alS.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Db_path != nil {
		alS.DBPath = *jsnCfg.Db_path
	}
	if jsnCfg.Index_type != nil {
		alS.IndexType = *jsnCfg.Index_type
	}
	if jsnCfg.Ttl != nil {
		if alS.TTL, err = time.ParseDuration(*jsnCfg.Ttl); err != nil {
			return
		}
	}
	if jsnCfg.Conns != nil {
		tagged := tagConns(jsnCfg.Conns)
		maps.Copy(alS.Conns, tagged)
	}
	if jsnCfg.Cleanup_interval != nil {
		if alS.CleanupInterval, err = time.ParseDuration(*jsnCfg.Cleanup_interval); err != nil {
			return
		}
	}
	if jsnCfg.Opts != nil {
		alS.Opts.loadFromJSONCfg(jsnCfg.Opts)
	}
	return nil
}

// AsMapInterface returns the config as a map[string]any
func (alS AnalyzerSCfg) AsMapInterface() any {
	opts := map[string]any{
		utils.MetaExporterIDs: alS.Opts.ExporterIDs,
	}
	mp := map[string]any{
		utils.EnabledCfg:         alS.Enabled,
		utils.DBPathCfg:          alS.DBPath,
		utils.IndexTypeCfg:       alS.IndexType,
		utils.TTLCfg:             alS.TTL.String(),
		utils.CleanupIntervalCfg: alS.CleanupInterval.String(),
		utils.OptsCfg:            opts,
	}
	mp[utils.ConnsCfg] = stripConns(alS.Conns)
	return mp
}

func (AnalyzerSCfg) SName() string             { return AnalyzerSJSON }
func (alS AnalyzerSCfg) CloneSection() Section { return alS.Clone() }

// Clone returns a deep copy of AnalyzerSCfg
func (alS AnalyzerSCfg) Clone() (cln *AnalyzerSCfg) {
	cln = &AnalyzerSCfg{
		Enabled:         alS.Enabled,
		DBPath:          alS.DBPath,
		IndexType:       alS.IndexType,
		TTL:             alS.TTL,
		Conns:           CloneConnsMap(alS.Conns),
		CleanupInterval: alS.CleanupInterval,
		Opts:            alS.Opts.Clone(),
	}
	return
}

func (anzOpts *AnalyzerSOpts) Clone() *AnalyzerSOpts {
	if anzOpts == nil {
		return nil
	}
	return &AnalyzerSOpts{
		ExporterIDs: []*DynamicStringSliceOpt(anzOpts.ExporterIDs),
	}
}

type AnalyzerSOptsJson struct {
	ExporterIDs []*DynamicStringSliceOpt `json:"*exporterIDs"`
}

// Analyzer service json config section
type AnalyzerSJsonCfg struct {
	Enabled          *bool
	Db_path          *string `json:"dbPath"`
	Index_type       *string `json:"indexType"`
	Ttl              *string
	Conns            map[string][]*DynamicConns `json:"conns,omitempty"`
	Cleanup_interval *string                    `json:"cleanupInterval"`
	Opts             *AnalyzerSOptsJson
}

func diffAnalyzerSOptsJsonCfg(d *AnalyzerSOptsJson, v1, v2 *AnalyzerSOpts) *AnalyzerSOptsJson {
	if d == nil {
		d = new(AnalyzerSOptsJson)
	}
	if !DynamicStringSliceOptEqual(v1.ExporterIDs, v2.ExporterIDs) {
		d.ExporterIDs = v2.ExporterIDs
	}
	return d
}

func diffAnalyzerSJsonCfg(d *AnalyzerSJsonCfg, v1, v2 *AnalyzerSCfg) *AnalyzerSJsonCfg {
	if d == nil {
		d = new(AnalyzerSJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if v1.DBPath != v2.DBPath {
		d.Db_path = utils.StringPointer(v2.DBPath)
	}
	if v1.IndexType != v2.IndexType {
		d.Index_type = utils.StringPointer(v2.IndexType)
	}
	if v1.TTL != v2.TTL {
		d.Ttl = utils.StringPointer(v2.TTL.String())
	}
	if !ConnsMapEqual(v1.Conns, v2.Conns) {
		d.Conns = stripConns(v2.Conns)
	}
	if v1.CleanupInterval != v2.CleanupInterval {
		d.Cleanup_interval = utils.StringPointer(v2.CleanupInterval.String())
	}
	d.Opts = diffAnalyzerSOptsJsonCfg(d.Opts, v1.Opts, v2.Opts)
	return d
}
