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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
)

func NewAttributeService(dm *DataManager, filterS *FilterS,
	stringIndexedFields, prefixIndexedFields *[]string) (*AttributeService, error) {
	return &AttributeService{dm: dm, filterS: filterS,
		stringIndexedFields: stringIndexedFields,
		prefixIndexedFields: prefixIndexedFields}, nil
}

type AttributeService struct {
	dm                  *DataManager
	filterS             *FilterS
	stringIndexedFields *[]string
	prefixIndexedFields *[]string
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
	aPrflIDs, err := matchingItemIDsForEvent(ev.Event, alS.stringIndexedFields, alS.prefixIndexedFields,
		alS.dm, utils.AttributeFilterIndexes, attrIdxKey)
	if err != nil {
		if err != utils.ErrNotFound {
			return nil, err
		}
		if aPrflIDs, err = matchingItemIDsForEvent(ev.Event, alS.stringIndexedFields, alS.prefixIndexedFields,
			alS.dm, utils.CacheAttributeFilterIndexes, utils.ConcatenatedKey(ev.Tenant, utils.META_ANY)); err != nil {
			return nil, err
		}
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
		if aPrfl.ActivationInterval != nil && ev.Time != nil &&
			!aPrfl.ActivationInterval.IsActiveAtTime(*ev.Time) { // not active
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
		if _, has := attrReply.CGREvent.Event[fld]; !has {
			continue //maybe removed
		}
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
	for fldName, initialMp := range attrPrf.attributes {
		initEvValIf, has := ev.Event[fldName]
		if !has {
			anyInitial, hasAny := initialMp[utils.ANY]
			if hasAny && anyInitial.Append &&
				initialMp[utils.ANY].Substitute != interface{}(utils.META_NONE) {
				rply.CGREvent.Event[fldName] = anyInitial.Substitute
			}
			rply.AlteredFields = append(rply.AlteredFields, fldName)
			continue
		}
		attrVal, has := initialMp[initEvValIf]
		if !has {
			attrVal, has = initialMp[utils.ANY]
		}
		if has {
			if attrVal.Substitute == interface{}(utils.META_NONE) {
				delete(rply.CGREvent.Event, fldName)
			} else {
				rply.CGREvent.Event[fldName] = attrVal.Substitute
			}
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
	attrPrfl *AttributeProfile) (err error) {
	attrPrf, err := alS.attributeProfileForEvent(ev)
	if err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*attrPrfl = *attrPrf
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
