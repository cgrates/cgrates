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

package engine

import (
	"errors"
	"fmt"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func NewChargerService(dm *DataManager, filterS *FilterS,
	attrS rpcclient.RpcClientConnection,
	cfg *config.CGRConfig) (*ChargerService, error) {
	return &ChargerService{dm: dm, filterS: filterS,
		attrS: attrS, cfg: cfg}, nil
}

// ChargerService is performing charging
type ChargerService struct {
	dm      *DataManager
	filterS *FilterS
	attrS   rpcclient.RpcClientConnection
	cfg     *config.CGRConfig
}

// ListenAndServe will initialize the service
func (cS *ChargerService) ListenAndServe(exitChan chan bool) (err error) {
	utils.Logger.Info(fmt.Sprintf("Starting %s", utils.ChargerS))
	e := <-exitChan
	exitChan <- e
	return
}

// Shutdown is called to shutdown the service
func (cS *ChargerService) Shutdown() (err error) {
	utils.Logger.Info(fmt.Sprintf("<%s> shutdown initialized", utils.ChargerS))
	utils.Logger.Info(fmt.Sprintf("<%s> shutdown complete", utils.ChargerS))
	return
}

// matchingChargingProfilesForEvent returns ordered list of matching chargers which are active by the time of the function call
func (cS *ChargerService) matchingChargerProfilesForEvent(cgrEv *utils.CGREvent) (cPs ChargerProfiles, err error) {
	cpIDs, err := matchingItemIDsForEvent(cgrEv.Event,
		cS.cfg.ChargerSCfg().StringIndexedFields, cS.cfg.ChargerSCfg().PrefixIndexedFields,
		cS.dm, utils.CacheChargerFilterIndexes, cgrEv.Tenant, cS.cfg.FilterSCfg().IndexedSelects)
	if err != nil {
		return nil, err
	}
	matchingCPs := make(map[string]*ChargerProfile)
	for cpID := range cpIDs {
		cP, err := cS.dm.GetChargerProfile(cgrEv.Tenant, cpID, false, utils.NonTransactional)
		if err != nil {
			if err == utils.ErrNotFound {
				continue
			}
			return nil, err
		}
		if cP.ActivationInterval != nil && cgrEv.Time != nil &&
			!cP.ActivationInterval.IsActiveAtTime(*cgrEv.Time) { // not active
			continue
		}
		if pass, err := cS.filterS.Pass(cgrEv.Tenant, cP.FilterIDs,
			NewNavigableMap(cgrEv.Event)); err != nil {
			return nil, err
		} else if !pass {
			continue
		}
		matchingCPs[cpID] = cP
	}
	cPs = make(ChargerProfiles, len(matchingCPs))
	i := 0
	for _, cP := range matchingCPs {
		cPs[i] = cP
		i++
	}
	cPs.Sort()
	return
}

func (cS *ChargerService) processEvent(cgrEv *utils.CGREvent) (cgrEvs []*utils.CGREvent, err error) {
	var cPs ChargerProfiles
	if cPs, err = cS.matchingChargerProfilesForEvent(cgrEv); err != nil {
		return nil, err
	}
	cgrEvs = make([]*utils.CGREvent, len(cPs))
	for i, cP := range cPs {
		cgrEvs[i] = cgrEv.Clone()
		cgrEvs[i].Event[utils.RunID] = cP.RunID
		if len(cP.AttributeIDs) != 0 { // Attributes should process the event
			if cS.attrS == nil {
				return nil, errors.New("no connection to AttributeS")
			}
			if cgrEvs[i].Context == nil {
				cgrEvs[i].Context = utils.StringPointer(utils.MetaChargers)
			}
			var rply AttrSProcessEventReply
			if err = cS.attrS.Call(utils.AttributeSv1ProcessEvent,
				&AttrArgsProcessEvent{cP.AttributeIDs, *cgrEvs[i]},
				&rply); err != nil {
				return nil, err
			}
			if len(rply.AlteredFields) != 0 {
				cgrEvs[i] = rply.CGREvent // modified event by attributeS
			}
		}
	}
	return
}

// V1ProcessEvent will process the event received via API and return list of events forked
func (cS *ChargerService) V1ProcessEvent(args *utils.CGREvent,
	reply *[]*utils.CGREvent) (err error) {
	if args.Event == nil {
		return utils.NewErrMandatoryIeMissing("Event")
	}
	rply, err := cS.processEvent(args)
	if err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = rply
	return
}

// V1GetChargersForEvent exposes the list of ordered matching ChargingProfiles for an event
func (cS *ChargerService) V1GetChargersForEvent(args *utils.CGREvent,
	rply *ChargerProfiles) (err error) {
	cPs, err := cS.matchingChargerProfilesForEvent(args)
	if err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*rply = cPs
	return
}
