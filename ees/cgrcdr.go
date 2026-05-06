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

package ees

import (
	"fmt"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewCgrcdr(cfg *config.EventExporterCfg, em *utils.ExporterMetrics,
	dm *engine.DataManager) (*Cgrcdr, error) {
	if dm == nil {
		return nil, fmt.Errorf("%s exporter requires DataManager", utils.MetaCgrcdr)
	}
	return &Cgrcdr{cfg: cfg, em: em, dm: dm}, nil
}

type Cgrcdr struct {
	cfg *config.EventExporterCfg
	em  *utils.ExporterMetrics
	dm  *engine.DataManager
}

func (cgr *Cgrcdr) Cfg() *config.EventExporterCfg                           { return cgr.cfg }
func (cgr *Cgrcdr) Connect() error                                          { return nil }
func (cgr *Cgrcdr) Close() error                                            { return nil }
func (cgr *Cgrcdr) GetMetrics() *utils.ExporterMetrics                      { return cgr.em }
func (cgr *Cgrcdr) ExtraData(ev *utils.CGREvent) any                        { return ev }
func (cgr *Cgrcdr) PrepareMap(*utils.CGREvent) (any, error)                 { return nil, nil }
func (cgr *Cgrcdr) PrepareOrderMap(*utils.OrderedNavigableMap) (any, error) { return nil, nil }

func (cgr *Cgrcdr) ExportEvent(ctx *context.Context, _, extraData any) error {
	cgrEv, ok := extraData.(*utils.CGREvent)
	if !ok {
		return fmt.Errorf("unexpected extraData type %T", extraData)
	}
	if cgrEv.APIOpts == nil {
		cgrEv.APIOpts = make(map[string]any)
	}
	if _, has := cgrEv.APIOpts[utils.MetaCDRID]; !has {
		cgrEv.APIOpts[utils.MetaCDRID] = utils.GetUniqueCDRID(cgrEv)
	}
	if err := cgr.dm.SetCDR(ctx, cgrEv, false); err != nil {
		if err != utils.ErrExists {
			return fmt.Errorf("storing CDR %s failed: %w", utils.ToJSON(cgrEv), err)
		}
		if err = cgr.dm.SetCDR(ctx, cgrEv, true); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: <%s> updating CDR %+v",
					utils.CDRs, err.Error(), utils.ToJSON(cgrEv)))
			return utils.ErrPartiallyExecuted
		}
	}
	return nil
}
