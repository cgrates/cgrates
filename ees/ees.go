/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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

package ees

import (
	"fmt"
	"sync"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// NewERService instantiates the EEService
func NewEEService(cfg *config.CGRConfig, filterS *engine.FilterS,
	connMgr *engine.ConnManager) *EEService {
	return &EEService{
		cfg:     cfg,
		filterS: filterS,
		connMgr: connMgr,
		ees:     make(map[string]EventExporter),
	}
}

// EEService is managing the EventExporters
type EEService struct {
	cfg     *config.CGRConfig
	filterS *engine.FilterS
	connMgr *engine.ConnManager

	ees    map[string]EventExporter // map[eeType]EventExporterID
	eesMux sync.RWMutex             // protects the ees
}

// ListenAndServe keeps the service alive
func (eeS *EEService) ListenAndServe(exitChan chan bool, cfgRld chan struct{}) (err error) {
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s>",
		utils.CoreS, utils.EventExporterService))
	for {
		select {
		case e := <-exitChan: // global exit
			eeS.Shutdown()
			exitChan <- e // put back for the others listening for shutdown request
			break
		case rld := <-cfgRld: // configuration was reloaded, destroy the cache
			cfgRld <- rld
			utils.Logger.Info(fmt.Sprintf("<%s> reloading configuration internals.",
				utils.EventExporterService))
			eeS.eesMux.Lock()
			eeS.ees = make(map[string]EventExporter)
			eeS.eesMux.Unlock()

		}
	}
	return
}

// Shutdown is called to shutdown the service
func (eeS *EEService) Shutdown() (err error) {
	utils.Logger.Info(fmt.Sprintf("<%s> shutdown <%s>", utils.CoreS, utils.EventExporterService))
	return
}

// ProcessEvent will be called each time a new event is received from readers
func (eeS *EEService) V1ProcessEvent(cgrEv *utils.CGREventWithOpts) (err error) {
	eeS.cfg.RLocks(config.EEsJson)
	defer eeS.cfg.RUnlocks(config.EEsJson)

	for _, eeCfg := range eeS.cfg.EEsCfg().Exporters {

		if eeCfg.Flags.GetBool(utils.MetaAttributes) {
			var rplyEv engine.AttrSProcessEventReply
			attrArgs := &engine.AttrArgsProcessEvent{
				AttributeIDs: eeCfg.AttributeSIDs,
				Context: utils.StringPointer(
					utils.FirstNonEmpty(
						eeCfg.AttributeSCtx,
						utils.IfaceAsString(cgrEv.Opts[utils.Context]),
						utils.MetaEEs)),
				CGREvent:      cgrEv.CGREvent,
				ArgDispatcher: cgrEv.ArgDispatcher,
			}
			if err = eeS.connMgr.Call(
				eeS.cfg.EEsCfg().AttributeSConns, nil,
				utils.AttributeSv1ProcessEvent,
				attrArgs, &rplyEv); err == nil && len(rplyEv.AlteredFields) != 0 {
				cgrEv.CGREvent = rplyEv.CGREvent
				cgrEv.Opts = rplyEv.Opts
			} else if err != nil {
				if err.Error() != utils.ErrNotFound.Error() {
					return
				}
				err = nil // cancel ErrNotFound
			}
		}

		eeS.eesMux.RLock()
		ee, has := eeS.ees[eeCfg.ID]
		eeS.eesMux.RUnlock()
		if !has {
			if ee, err = NewEventExporter(eeCfg); err != nil {
				return
			}
			eeS.ees[eeCfg.ID] = ee
		}
		if err = ee.ExportEvent(cgrEv.CGREvent); err != nil {
			return
		}
	}
	return
}
