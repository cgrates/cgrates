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
	"math"
	"strconv"

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
func (alS *AttributeService) attributeProfileForEvent(args *AttrArgsProcessEvent) (matchAttrPrfl *AttributeProfile, err error) {
	var attrIDs []string
	contextVal := utils.META_DEFAULT
	if args.Context != nil && *args.Context != "" {
		contextVal = *args.Context
	}
	attrIdxKey := utils.ConcatenatedKey(args.Tenant, contextVal)
	if len(args.AttributeIDs) != 0 {
		attrIDs = args.AttributeIDs
	} else {
		aPrflIDs, err := MatchingItemIDsForEvent(args.Event, alS.stringIndexedFields, alS.prefixIndexedFields,
			alS.dm, utils.CacheAttributeFilterIndexes, attrIdxKey, alS.filterS.cfg.AttributeSCfg().IndexedSelects)
		if err != nil {
			if err != utils.ErrNotFound {
				return nil, err
			}
			if aPrflIDs, err = MatchingItemIDsForEvent(args.Event, alS.stringIndexedFields, alS.prefixIndexedFields,
				alS.dm, utils.CacheAttributeFilterIndexes, utils.ConcatenatedKey(args.Tenant, utils.META_ANY),
				alS.filterS.cfg.AttributeSCfg().IndexedSelects); err != nil {
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
		if matchAttrPrfl == nil || matchAttrPrfl.Weight < aPrfl.Weight {
			matchAttrPrfl = aPrfl
		}
	}
	// All good, convert from Map to Slice so we can sort
	if matchAttrPrfl == nil {
		return nil, utils.ErrNotFound
	}
	return
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
	Context      *string // attach the event to a context
	ProcessRuns  *int    // number of loops for ProcessEvent
	*utils.CGREvent
	*utils.ArgDispatcher
}

// processEvent will match event with attribute profile and do the necessary replacements
func (alS *AttributeService) processEvent(args *AttrArgsProcessEvent) (
	rply *AttrSProcessEventReply, err error) {
	attrPrf, err := alS.attributeProfileForEvent(args)
	if err != nil {
		return nil, err
	}
	rply = &AttrSProcessEventReply{
		MatchedProfiles: []string{attrPrf.ID},
		CGREvent:        args.Clone(),
		blocker:         attrPrf.Blocker}

	for _, attribute := range attrPrf.Attributes {
		//in case that we have filter for attribute send them to FilterS to be processed
		if len(attribute.FilterIDs) != 0 {
			if pass, err := alS.filterS.Pass(args.Tenant, attribute.FilterIDs,
				config.NewNavigableMap(args.Event)); err != nil {
				return nil, err
			} else if !pass {
				continue
			}
		}
		var substitute string
		var err error
		switch attribute.Type {
		case utils.META_CONSTANT:
			substitute, err = attribute.Value.ParseValue(utils.EmptyString)
		case utils.MetaVariable, utils.META_COMPOSED:
			substitute, err = attribute.Value.ParseEvent(args.Event)
		case utils.META_USAGE_DIFFERENCE:
			if len(attribute.Value) != 2 {
				return nil, fmt.Errorf("invalid arguments <%s>", utils.ToJSON(attribute.Value))
			}
			strVal1, err := attribute.Value[0].ParseEvent(args.Event)
			if err != nil {
				return nil, err
			}
			strVal2, err := attribute.Value[1].ParseEvent(args.Event)
			if err != nil {
				return nil, err
			}
			tEnd, err := utils.ParseTimeDetectLayout(strVal1, utils.EmptyString)
			if err != nil {
				return nil, err
			}
			tStart, err := utils.ParseTimeDetectLayout(strVal2, utils.EmptyString)
			if err != nil {
				return nil, err
			}
			substitute = tEnd.Sub(tStart).String()
		case utils.MetaSum:
			iFaceVals := make([]interface{}, len(attribute.Value))
			for i, val := range attribute.Value {
				strVal, err := val.ParseEvent(args.Event)
				if err != nil {
					return nil, err
				}
				iFaceVals[i] = utils.StringToInterface(strVal)
			}
			ifaceSum, err := utils.Sum(iFaceVals...)
			if err != nil {
				return nil, err
			}
			substitute, err = utils.IfaceAsString(ifaceSum)
		case utils.MetaValueExponent:
			if len(attribute.Value) != 2 {
				return nil, fmt.Errorf("invalid arguments <%s> to %s",
					utils.ToJSON(attribute.Value), utils.MetaValueExponent)
			}
			strVal1, err := attribute.Value[0].ParseEvent(args.Event) // String Value
			if err != nil {
				return nil, err
			}
			val, err := strconv.ParseFloat(strVal1, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid value <%s> to %s",
					strVal1, utils.MetaValueExponent)
			}
			strVal2, err := attribute.Value[1].ParseEvent(args.Event) // String Exponent
			if err != nil {
				return nil, err
			}
			exp, err := strconv.Atoi(strVal2)
			if err != nil {
				return nil, err
			}
			substitute = strconv.FormatFloat(utils.Round(val*math.Pow10(exp),
				config.CgrConfig().GeneralCfg().RoundingDecimals, utils.ROUNDING_MIDDLE), 'f', -1, 64)
		default: // backwards compatible in case that Type is empty
			substitute, err = attribute.Value.ParseEvent(args.Event)
		}

		if err != nil {
			return nil, err
		}
		//add only once the FieldName in AlteredFields
		if !utils.IsSliceMember(rply.AlteredFields, attribute.FieldName) {
			rply.AlteredFields = append(rply.AlteredFields, attribute.FieldName)
		}
		if substitute == utils.META_NONE {
			delete(rply.CGREvent.Event, attribute.FieldName)
			continue
		}
		if attribute.Type == utils.META_COMPOSED {
			evStrVal, err := utils.IfaceAsString(rply.CGREvent.Event[attribute.FieldName])
			if err != nil {
				return nil, err
			}
			substitute = evStrVal + substitute
		}
		rply.CGREvent.Event[attribute.FieldName] = substitute

	}
	return
}

func (alS *AttributeService) V1GetAttributeForEvent(args *AttrArgsProcessEvent,
	attrPrfl *AttributeProfile) (err error) {
	if args.CGREvent == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
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
	if args.CGREvent == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	if args.Event == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	if args.ProcessRuns == nil || *args.ProcessRuns == 0 {
		args.ProcessRuns = utils.IntPointer(alS.processRuns)
	}
	var apiRply *AttrSProcessEventReply // aggregate response here
	for i := 0; i < *args.ProcessRuns; i++ {
		var evRply *AttrSProcessEventReply
		evRply, err = alS.processEvent(args)
		if err != nil {
			if err != utils.ErrNotFound {
				err = utils.NewErrServerError(err)
			} else if i != 0 { // ignore "not found" in a loop different than 0
				err = nil
			}
			break
		}
		if len(evRply.AlteredFields) != 0 {
			args.CGREvent = evRply.CGREvent // for next loop
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
	// Make sure the requested fields were populated
	if err == utils.ErrNotFound {
		for _, valIface := range args.CGREvent.Event {
			if valIface == interface{}(utils.MetaAttributes) {
				err = utils.ErrMandatoryIeMissing
				break
			}
		}
	} else if err == nil {
		for _, valIface := range apiRply.CGREvent.Event {
			if valIface == interface{}(utils.MetaAttributes) {
				// mandatory IE missing
				err = utils.ErrMandatoryIeMissing
				break
			}
		}
	}
	if err != nil {
		return
	}
	*reply = *apiRply
	return
}
