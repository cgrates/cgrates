// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ees

import (
	"github.com/cgrates/birpc/context"
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

func (vEe *VirtualEE) Cfg() *config.EventExporterCfg                { return vEe.cfg }
func (vEe *VirtualEE) Connect() error                               { return nil }
func (vEe *VirtualEE) ExportEvent(*context.Context, any, any) error { return nil }
func (vEe *VirtualEE) Close() error                                 { return nil }
func (vEe *VirtualEE) GetMetrics() *utils.ExporterMetrics           { return vEe.em }
func (vEe *VirtualEE) ExtraData(*utils.CGREvent) any                { return nil }
func (vEe *VirtualEE) PrepareMap(mp *utils.CGREvent) (any, error)   { return nil, nil }
func (vEe *VirtualEE) PrepareOrderMap(*utils.OrderedNavigableMap) (any, error) {
	return nil, nil
}
