// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ees

import (
	"fmt"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func NewVirtualEE(cfg *config.EventExporterCfg, em *utils.ExporterMetrics) *VirtualEE {
	return &VirtualEE{
		cfg: cfg,
		em:  em,
	}
}

// VirtualEE implements EventExporter interface for .csv files
type VirtualEE struct {
	cfg *config.EventExporterCfg
	em  *utils.ExporterMetrics
}

func (vEe *VirtualEE) Cfg() *config.EventExporterCfg { return vEe.cfg }
func (vEe *VirtualEE) Connect() error                { return nil }

func (vEe *VirtualEE) ExportEvent(payload any, _ string) error {
	utils.Logger.Info(
		fmt.Sprintf("<%s> <%s> exported: <%s>",
			utils.EEs, vEe.Cfg().ID, utils.ToJSON(payload)))
	return nil
}

func (vEe *VirtualEE) Close() error                       { return nil }
func (vEe *VirtualEE) GetMetrics() *utils.ExporterMetrics { return vEe.em }

func (vEe *VirtualEE) PrepareMap(cgrEv *utils.CGREvent) (any, error) {
	return cgrEv.Event, nil
}

func (vEe *VirtualEE) PrepareOrderMap(onm *utils.OrderedNavigableMap) (any, error) {
	return onm.AsMap(), nil
}
