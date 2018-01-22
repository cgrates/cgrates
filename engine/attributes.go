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
	"fmt"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
)

func NewAttributeService(dm *DataManager, filterS *FilterS, indexedFields []string) (*AttributeService, error) {
	return &AttributeService{dm: dm, filterS: filterS, indexedFields: indexedFields}, nil
}

type AttributeService struct {
	dm            *DataManager
	filterS       *FilterS
	indexedFields []string
}

// ListenAndServe will initialize the service
func (alS *AttributeService) ListenAndServe(exitChan chan bool) (err error) {
	utils.Logger.Info("Starting Attribute service")
	e := <-exitChan
	exitChan <- e // put back for the others listening for shutdown request
	return
}

// Shutdown is called to shutdown the service
func (alS *AttributeService) Shutdown() (err error) {
	utils.Logger.Info(fmt.Sprintf("<%s> shutdown initialized", utils.AttributeS))
	utils.Logger.Info(fmt.Sprintf("<%s> shutdown complete", utils.AttributeS))
	return
}

// matchingAttributeProfilesForEvent returns ordered list of matching resources which are active by the time of the call
func (alS *AttributeService) matchingAttributeProfilesForEvent(ev *utils.CGREvent) (aPrfls AttributeProfiles, err error) {
	var attrIdxKey string
	contextVal := utils.META_DEFAULT
	if ev.Context != nil && *ev.Context != "" {
		contextVal = *ev.Context
	}
	attrIdxKey = utils.ConcatenatedKey(ev.Tenant, contextVal)
	matchingAPs := make(map[string]*AttributeProfile)
	aPrflIDs, err := matchingItemIDsForEvent(ev.Event, alS.indexedFields,
		alS.dm, utils.AttributeFilterIndexes+attrIdxKey, MetaString)
	if err != nil {
		return nil, err
	}
	lockIDs := utils.PrefixSliceItems(aPrflIDs.Slice(), utils.AttributeFilterIndexes)
	guardian.Guardian.GuardIDs(config.CgrConfig().LockingTimeout, lockIDs...)
	defer guardian.Guardian.UnguardIDs(lockIDs...)
	for apID := range aPrflIDs {
		aPrfl, err := alS.dm.GetAttributeProfile(ev.Tenant, apID, false, utils.NonTransactional)
		if err != nil {
			if err == utils.ErrNotFound {
				continue
			}
			return nil, err
		}
		evTime := time.Now()
		if ev.Time != nil {
			evTime = *ev.Time
		}
		if aPrfl.ActivationInterval != nil &&
			!aPrfl.ActivationInterval.IsActiveAtTime(evTime) { // not active
			continue
		}
		if pass, err := alS.filterS.PassFiltersForEvent(ev.Tenant,
			ev.Event, aPrfl.FilterIDs); err != nil {
			return nil, err
		} else if !pass {
			continue
		}
		matchingAPs[apID] = aPrfl
	}
	// All good, convert from Map to Slice so we can sort
	aPrfls = make(AttributeProfiles, len(matchingAPs))
	i := 0
	for _, aPrfl := range matchingAPs {
		aPrfls[i] = aPrfl
		i++
	}
	aPrfls.Sort()
	return
}

func (alS *AttributeService) attributeProfileForEvent(ev *utils.CGREvent) (attrPrfl *AttributeProfile, err error) {
	var attrPrfls AttributeProfiles
	if attrPrfls, err = alS.matchingAttributeProfilesForEvent(ev); err != nil {
		return
	} else if len(attrPrfls) == 0 {
		return nil, utils.ErrNotFound
	}
	return attrPrfls[0], nil
}

// AttrSFldNameValue is a helper struct for AttrSDigest deserialization
type AttrSFieldNameValue struct {
	FieldName  string
	FieldValue string
}

type AttrSProcessEventReply struct {
	MatchedProfile string
	AlteredFields  []string
	CGREvent       *utils.CGREvent
}

// Digest returns serialized version of alteredFields in AttrSProcessEventReply
// format fldName1:fldVal1,fldName2:fldVal2
func (attrReply *AttrSProcessEventReply) Digest() (rplyDigest string) {
	for i, fld := range attrReply.AlteredFields {
		if i != 0 {
			rplyDigest += utils.FIELDS_SEP
		}
		fldStrVal, _ := attrReply.CGREvent.FieldAsString(fld)
		rplyDigest += fld + utils.InInFieldSep + fldStrVal
	}
	return
}

// processEvent will match event with attribute profile and do the necessary replacements
func (alS *AttributeService) processEvent(ev *utils.CGREvent) (rply *AttrSProcessEventReply, err error) {
	attrPrf, err := alS.attributeProfileForEvent(ev)
	if err != nil {
		return nil, err
	}
	rply = &AttrSProcessEventReply{MatchedProfile: attrPrf.ID, CGREvent: ev.Clone()}
	for fldName, intialMp := range attrPrf.Attributes {
		initEvValIf, has := ev.Event[fldName]
		if !has { // we don't have initial in event, try append
			if anyInitial, has := intialMp[utils.ANY]; has && anyInitial.Append {
				rply.CGREvent.Event[fldName] = anyInitial.Substitute
				rply.AlteredFields = append(rply.AlteredFields, fldName)
			}
			continue
		}
		initEvVal, cast := utils.CastFieldIfToString(initEvValIf)
		if !cast {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> ev: %s, cannot cast field: %+v to string",
					utils.AttributeS, ev, fldName))
			continue
		}
		attrVal, has := intialMp[initEvVal]
		if !has {
			attrVal, has = intialMp[utils.ANY]
		}
		if has {
			rply.CGREvent.Event[fldName] = attrVal.Substitute
			rply.AlteredFields = append(rply.AlteredFields, fldName)
		}
		for _, valIface := range rply.CGREvent.Event {
			if valIface == interface{}(utils.MetaAttributes) {
				return nil, utils.NewCGRError(utils.AttributeSv1ProcessEvent,
					utils.AttributesNotFoundCaps,
					utils.AttributesNotFound,
					utils.AttributesNotFound)
			}
		}
	}
	return
}

func (alS *AttributeService) V1GetAttributeForEvent(ev *utils.CGREvent,
	extattrPrf *ExternalAttributeProfile) (err error) {
	attrPrf, err := alS.attributeProfileForEvent(ev)
	if err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	eattrPrfl := NewExternalAttributeProfileFromAttributeProfile(attrPrf)
	*extattrPrf = *eattrPrfl
	return
}

func (alS *AttributeService) V1ProcessEvent(ev *utils.CGREvent,
	reply *AttrSProcessEventReply) (err error) {
	evReply, err := alS.processEvent(ev)
	if err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = *evReply
	return
}
