// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ees

import (
	"fmt"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func NewLogEE(cfg *config.EventExporterCfg, em *utils.ExporterMetrics) *LogEE {
	return &LogEE{
		cfg: cfg,
		em:  em,
	}
}

// LogEE implements EventExporter interface for .csv files
type LogEE struct {
	cfg *config.EventExporterCfg
	em  *utils.ExporterMetrics
}

func (vEe *LogEE) Cfg() *config.EventExporterCfg { return vEe.cfg }
func (vEe *LogEE) Connect() error                { return nil }
func (vEe *LogEE) ExportEvent(mp any, _ string) error {
	utils.Logger.Info(
		fmt.Sprintf("<%s> <%s> exported: <%s>",
			utils.EEs, vEe.Cfg().ID, utils.ToJSON(mp)))
	return nil
}
func (vEe *LogEE) Close() error                       { return nil }
func (vEe *LogEE) GetMetrics() *utils.ExporterMetrics { return vEe.em }
func (vEe *LogEE) PrepareMap(mp *utils.CGREvent) (any, error) {
	return mp.Event, nil
}
func (vEe *LogEE) PrepareOrderMap(mp *utils.OrderedNavigableMap) (any, error) {
	return mp.AsMap(), nil
}
