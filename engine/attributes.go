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
	"github.com/cgrates/cgrates/utils"
)

func NewAttributeService(dm *DataManager, filterS *FilterS,
	stringIndexedFields, prefixIndexedFields *[]string,
	processRuns int) (*AttributeService, error) {
	return &AttributeService{dm: dm, filterS: filterS,
		stringIndexedFields: stringIndexedFields,
		prefixIndexedFields: prefixIndexedFields,
		processRuns:         processRuns}, nil
}

type AttributeService struct {
	dm                  *DataManager
	filterS             *FilterS
	stringIndexedFields *[]string
	prefixIndexedFields *[]string
	processRuns         int
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
func (alS *AttributeService) matchingAttributeProfilesForEvent(args *AttrArgsProcessEvent) (aPrfls AttributeProfiles, err error) {
	var attrIdxKey string
	var attrIDs []string
	contextVal := utils.META_DEFAULT
	if args.Context != nil && *args.Context != "" {
		contextVal = *args.Context
	}
	attrIdxKey = utils.ConcatenatedKey(args.Tenant, contextVal)
	matchingAPs := make(map[string]*AttributeProfile)
	if len(args.AttributeIDs) != 0 {
		attrIDs = args.AttributeIDs
	} else {
		aPrflIDs, err := matchingItemIDsForEvent(args.Event, alS.stringIndexedFields, alS.prefixIndexedFields,
			alS.dm, utils.CacheAttributeFilterIndexes, attrIdxKey, alS.filterS.cfg.FilterSCfg().IndexedSelects)
		if err != nil {
			if err != utils.ErrNotFound {
				return nil, err
			}
			if aPrflIDs, err = matchingItemIDsForEvent(args.Event, alS.stringIndexedFields, alS.prefixIndexedFields,
				alS.dm, utils.CacheAttributeFilterIndexes, utils.ConcatenatedKey(args.Tenant, utils.META_ANY),
				alS.filterS.cfg.FilterSCfg().IndexedSelects); err != nil {
				return nil, err
			}
		}
		attrIDs = aPrflIDs.Slice()
	}
	for _, apID := range attrIDs {
		aPrfl, err := alS.dm.GetAttributeProfile(args.Tenant, apID, true, true, utils.NonTransactional)
		if err != nil {
			if err == utils.ErrNotFound {
				continue
			}
			return nil, err
		}
		if aPrfl.ActivationInterval != nil && args.Time != nil &&
			!aPrfl.ActivationInterval.IsActiveAtTime(*args.Time) { // not active
			continue
		}
		if pass, err := alS.filterS.Pass(args.Tenant, aPrfl.FilterIDs,
			config.NewNavigableMap(args.Event)); err != nil {
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

func (alS *AttributeService) attributeProfileForEvent(args *AttrArgsProcessEvent) (attrPrfl *AttributeProfile, err error) {
	var attrPrfls AttributeProfiles
	if attrPrfls, err = alS.matchingAttributeProfilesForEvent(args); err != nil {
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
	MatchedProfiles []string
	AlteredFields   []string
	CGREvent        *utils.CGREvent
	blocker         bool // internally used to stop further processRuns
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

type AttrArgsProcessEvent struct {
	AttributeIDs []string
	ProcessRuns  *int // number of loops for ProcessEvent
	utils.CGREvent
}

// processEvent will match event with attribute profile and do the necessary replacements
func (alS *AttributeService) processEvent(args *AttrArgsProcessEvent) (
	rply *AttrSProcessEventReply, err error) {
	attrPrf, err := alS.attributeProfileForEvent(args)
	if err != nil {
		if err == utils.ErrNotFound {
			// change the error in case that at least one field need to be processed by attributes
			for _, valIface := range args.CGREvent.Event {
				if valIface == interface{}(utils.MetaAttributes) {
					err = utils.ErrMandatoryIeMissing
					break
				}
			}
		}
		return nil, err
	}
	rply = &AttrSProcessEventReply{
		MatchedProfiles: []string{attrPrf.ID},
		CGREvent:        args.Clone(),
		blocker:         attrPrf.Blocker}
	for fldName, initialMp := range attrPrf.attributesIdx {
		initEvValIf, has := args.Event[fldName]
		if !has {
			anyInitial, hasAny := initialMp[utils.ANY]
			if hasAny && anyInitial.Append { // add field name
				substitute, err := anyInitial.Substitute.ParseEvent(args.Event)
				if err != nil {
					return nil, err
				}
				rply.CGREvent.Event[fldName] = substitute
				rply.AlteredFields = append(rply.AlteredFields, fldName)
			}
			continue
		}
		attrVal, has := initialMp[initEvValIf]
		if !has {
			attrVal, has = initialMp[utils.ANY]
		}
		if has {
			substitute, err := attrVal.Substitute.ParseEvent(args.Event)
			if err != nil {
				return nil, err
			}
			if substitute == utils.META_NONE {
				delete(rply.CGREvent.Event, fldName)
			} else {
				rply.CGREvent.Event[fldName] = substitute
			}
			rply.AlteredFields = append(rply.AlteredFields, fldName)
		}
	}
	for _, valIface := range rply.CGREvent.Event {
		if valIface == interface{}(utils.MetaAttributes) {
			// mandatory IE missing
			return nil, utils.NewCGRError(
				utils.AttributeSv1ProcessEvent,
				utils.ErrMandatoryIeMissing.Error(),
				utils.ErrMandatoryIeMissing.Error(),
				utils.ErrMandatoryIeMissing.Error())
		}
	}
	return
}

func (alS *AttributeService) V1GetAttributeForEvent(args *AttrArgsProcessEvent,
	attrPrfl *AttributeProfile) (err error) {
	attrPrf, err := alS.attributeProfileForEvent(args)
	if err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*attrPrfl = *attrPrf
	return
}

func (alS *AttributeService) V1ProcessEvent(args *AttrArgsProcessEvent,
	reply *AttrSProcessEventReply) (err error) {
	if args.Event == nil {
		return utils.NewErrMandatoryIeMissing("Event")
	}
	if args.ProcessRuns == nil || *args.ProcessRuns == 0 {
		args.ProcessRuns = utils.IntPointer(alS.processRuns)
	}
	var apiRply *AttrSProcessEventReply // aggregate response here
	for i := 0; i < *args.ProcessRuns; i++ {
		evRply, err := alS.processEvent(args)
		if err != nil {
			if err != utils.ErrNotFound && err != utils.ErrMandatoryIeMissing {
				err = utils.NewErrServerError(err)
			} else if i != 0 { // ignore "not found" in a loop different than 0
				err = nil
				break
			}
			return err
		}
		if len(evRply.AlteredFields) != 0 {
			args.CGREvent = *evRply.CGREvent // for next loop
		}
		if apiRply == nil { // first reply
			apiRply = evRply
			if apiRply.blocker {
				break
			}
			continue
		}
		if utils.IsSliceMember(apiRply.MatchedProfiles,
			evRply.MatchedProfiles[0]) { // don't process the same AttributeProfile twice
			break
		}
		apiRply.MatchedProfiles = append(apiRply.MatchedProfiles, evRply.MatchedProfiles[0])
		apiRply.CGREvent = evRply.CGREvent
		for _, fldName := range evRply.AlteredFields {
			if utils.IsSliceMember(apiRply.AlteredFields, fldName) {
				continue // only add processed fieldName once
			}
			apiRply.AlteredFields = append(apiRply.AlteredFields, fldName)
		}
		if evRply.blocker {
			break
		}
	}
	*reply = *apiRply
	return
}
