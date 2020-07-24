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
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// NewAttributeService returns a new AttributeService
func NewAttributeService(dm *DataManager, filterS *FilterS,
	cgrcfg *config.CGRConfig) (*AttributeService, error) {
	return &AttributeService{
		dm:      dm,
		filterS: filterS,
		cgrcfg:  cgrcfg,
	}, nil
}

// AttributeService the service for the API
type AttributeService struct {
	dm      *DataManager
	filterS *FilterS
	cgrcfg  *config.CGRConfig
}

// ListenAndServe will initialize the service
func (alS *AttributeService) ListenAndServe(exitChan chan bool) (err error) {
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.AttributeS))
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
func (alS *AttributeService) attributeProfileForEvent(args *AttrArgsProcessEvent, evNm utils.MapStorage, lastID string) (matchAttrPrfl *AttributeProfile, err error) {
	var attrIDs []string
	contextVal := utils.MetaDefault
	if args.Context != nil && *args.Context != "" {
		contextVal = *args.Context
	}
	attrIdxKey := utils.ConcatenatedKey(args.Tenant, contextVal)
	if len(args.AttributeIDs) != 0 {
		attrIDs = args.AttributeIDs
	} else {
		aPrflIDs, err := MatchingItemIDsForEvent(args.Event,
			alS.cgrcfg.AttributeSCfg().StringIndexedFields,
			alS.cgrcfg.AttributeSCfg().PrefixIndexedFields,
			alS.dm, utils.CacheAttributeFilterIndexes, attrIdxKey,
			alS.cgrcfg.AttributeSCfg().IndexedSelects,
			alS.cgrcfg.AttributeSCfg().NestedFields,
		)
		if err != nil {
			if err != utils.ErrNotFound {
				return nil, err
			}
			if aPrflIDs, err = MatchingItemIDsForEvent(args.Event,
				alS.cgrcfg.AttributeSCfg().StringIndexedFields,
				alS.cgrcfg.AttributeSCfg().PrefixIndexedFields,
				alS.dm, utils.CacheAttributeFilterIndexes,
				utils.ConcatenatedKey(args.Tenant, utils.META_ANY),
				alS.cgrcfg.AttributeSCfg().IndexedSelects,
				alS.cgrcfg.AttributeSCfg().NestedFields); err != nil {
				return nil, err
			}
		}
		attrIDs = aPrflIDs.AsSlice()
	}
	for _, apID := range attrIDs {
		aPrfl, err := alS.dm.GetAttributeProfile(args.Tenant, apID, true, true, utils.NonTransactional)
		if err != nil {
			if err == utils.ErrNotFound {
				continue
			}
			return nil, err
		}
		if !(len(aPrfl.Contexts) == 1 && aPrfl.Contexts[0] == utils.META_ANY) &&
			!utils.IsSliceMember(aPrfl.Contexts, contextVal) {
			continue
		}
		if aPrfl.ActivationInterval != nil && args.Time != nil &&
			!aPrfl.ActivationInterval.IsActiveAtTime(*args.Time) { // not active
			continue
		}
		if pass, err := alS.filterS.Pass(args.Tenant, aPrfl.FilterIDs,
			evNm); err != nil {
			return nil, err
		} else if !pass {
			continue
		}
		if (matchAttrPrfl == nil || matchAttrPrfl.Weight < aPrfl.Weight) &&
			apID != lastID {
			matchAttrPrfl = aPrfl
		}
	}
	// All good, convert from Map to Slice so we can sort
	if matchAttrPrfl == nil {
		return nil, utils.ErrNotFound
	}
	return
}

// AttrSProcessEventReply reply used for proccess event
type AttrSProcessEventReply struct {
	MatchedProfiles []string
	AlteredFields   []string
	CGREvent        *utils.CGREvent
	Opts            map[string]interface{}
	blocker         bool // internally used to stop further processRuns
}

// Digest returns serialized version of alteredFields in AttrSProcessEventReply
// format fldName1:fldVal1,fldName2:fldVal2
func (attrReply *AttrSProcessEventReply) Digest() (rplyDigest string) {
	for i, fld := range attrReply.AlteredFields {
		fld = strings.TrimPrefix(fld, utils.MetaReq+utils.NestingSep)
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

// AttrArgsProcessEvent arguments used for proccess event
type AttrArgsProcessEvent struct {
	AttributeIDs []string
	Context      *string // attach the event to a context
	ProcessRuns  *int    // number of loops for ProcessEvent
	Opts         map[string]interface{}
	*utils.CGREvent
	*utils.ArgDispatcher
}

// processEvent will match event with attribute profile and do the necessary replacements
func (alS *AttributeService) processEvent(args *AttrArgsProcessEvent, evNm utils.MapStorage, lastID string) (
	rply *AttrSProcessEventReply, err error) {
	var attrPrf *AttributeProfile
	if attrPrf, err = alS.attributeProfileForEvent(args, evNm, lastID); err != nil {
		return
	}
	rply = &AttrSProcessEventReply{
		MatchedProfiles: []string{attrPrf.ID},
		CGREvent:        args.CGREvent,
		Opts:            args.Opts,
		blocker:         attrPrf.Blocker,
	}
	for _, attribute := range attrPrf.Attributes {
		//in case that we have filter for attribute send them to FilterS to be processed
		if len(attribute.FilterIDs) != 0 {
			var pass bool
			if pass, err = alS.filterS.Pass(args.Tenant, attribute.FilterIDs,
				evNm); err != nil {
				return
			} else if !pass {
				continue
			}
		}
		var substitute string
		switch attribute.Type {
		case utils.META_CONSTANT:
			substitute, err = attribute.Value.ParseValue(utils.EmptyString)
		case utils.MetaVariable, utils.META_COMPOSED:
			substitute, err = attribute.Value.ParseDataProvider(evNm)
		case utils.META_USAGE_DIFFERENCE:
			if len(attribute.Value) != 2 {
				return nil, fmt.Errorf("invalid arguments <%s>", utils.ToJSON(attribute.Value))
			}
			var strVal1 string
			if strVal1, err = attribute.Value[0].ParseDataProvider(evNm); err != nil {
				rply = nil
				return
			}
			var strVal2 string
			if strVal2, err = attribute.Value[1].ParseDataProvider(evNm); err != nil {
				rply = nil
				return
			}
			var tEnd time.Time
			if tEnd, err = utils.ParseTimeDetectLayout(strVal1, utils.EmptyString); err != nil {
				rply = nil
				return
			}
			var tStart time.Time
			if tStart, err = utils.ParseTimeDetectLayout(strVal2, utils.EmptyString); err != nil {
				rply = nil
				return
			}
			substitute = tEnd.Sub(tStart).String()
		case utils.MetaSum:
			var ifaceVals []interface{}
			if ifaceVals, err = getIfaceFromValues(attribute.Value, evNm); err != nil {
				rply = nil
				return
			}
			var ifaceSum interface{}
			if ifaceSum, err = utils.Sum(ifaceVals...); err != nil {
				rply = nil
				return
			}
			substitute = utils.IfaceAsString(ifaceSum)
		case utils.MetaDifference:
			var ifaceVals []interface{}
			if ifaceVals, err = getIfaceFromValues(attribute.Value, evNm); err != nil {
				rply = nil
				return
			}
			var ifaceSum interface{}
			if ifaceSum, err = utils.Difference(ifaceVals...); err != nil {
				rply = nil
				return
			}
			substitute = utils.IfaceAsString(ifaceSum)
		case utils.MetaMultiply:
			var ifaceVals []interface{}
			if ifaceVals, err = getIfaceFromValues(attribute.Value, evNm); err != nil {
				rply = nil
				return
			}
			var ifaceSum interface{}
			if ifaceSum, err = utils.Multiply(ifaceVals...); err != nil {
				rply = nil
				return
			}
			substitute = utils.IfaceAsString(ifaceSum)
		case utils.MetaDivide:
			var ifaceVals []interface{}
			if ifaceVals, err = getIfaceFromValues(attribute.Value, evNm); err != nil {
				rply = nil
				return
			}
			var ifaceSum interface{}
			if ifaceSum, err = utils.Divide(ifaceVals...); err != nil {
				rply = nil
				return
			}
			substitute = utils.IfaceAsString(ifaceSum)
		case utils.MetaValueExponent:
			if len(attribute.Value) != 2 {
				return nil, fmt.Errorf("invalid arguments <%s> to %s",
					utils.ToJSON(attribute.Value), utils.MetaValueExponent)
			}
			var strVal1 string
			if strVal1, err = attribute.Value[0].ParseDataProvider(evNm); err != nil {
				rply = nil
				return
			}
			var val float64
			if val, err = strconv.ParseFloat(strVal1, 64); err != nil {
				return nil, fmt.Errorf("invalid value <%s> to %s",
					strVal1, utils.MetaValueExponent)
			}
			var strVal2 string
			if strVal2, err = attribute.Value[1].ParseDataProvider(evNm); err != nil {
				rply = nil
				return
			}
			var exp int
			if exp, err = strconv.Atoi(strVal2); err != nil {
				rply = nil
				return
			}
			substitute = strconv.FormatFloat(utils.Round(val*math.Pow10(exp),
				alS.cgrcfg.GeneralCfg().RoundingDecimals, utils.ROUNDING_MIDDLE), 'f', -1, 64)
		case utils.MetaUnixTimestamp:
			var val string
			if val, err = attribute.Value.ParseDataProvider(evNm); err != nil {
				rply = nil
				return
			}
			var t time.Time
			if t, err = utils.ParseTimeDetectLayout(val, alS.cgrcfg.GeneralCfg().DefaultTimezone); err != nil {
				rply = nil
				return
			}
			substitute = strconv.Itoa(int(t.Unix()))
		default: // backwards compatible in case that Type is empty
			substitute, err = attribute.Value.ParseDataProvider(evNm)
		}

		if err != nil {
			rply = nil
			return
		}
		//add only once the Path in AlteredFields
		if !utils.IsSliceMember(rply.AlteredFields, attribute.Path) {
			rply.AlteredFields = append(rply.AlteredFields, attribute.Path)
		}
		if attribute.Path == utils.MetaTenant {
			if attribute.Type == utils.META_COMPOSED {
				rply.CGREvent.Tenant += substitute
			} else {
				rply.CGREvent.Tenant = substitute
			}
			continue
		}
		if substitute == utils.MetaRemove {
			evNm.Remove(strings.Split(attribute.Path, utils.NestingSep))
			continue
		}
		if attribute.Type == utils.META_COMPOSED {
			var val string
			if val, err = evNm.FieldAsString(strings.Split(attribute.Path, utils.NestingSep)); err != nil && err != utils.ErrNotFound {
				rply = nil
				return
			}
			substitute = val + substitute
		}
		if err = evNm.Set(strings.Split(attribute.Path, utils.NestingSep), substitute); err != nil {
			rply = nil
			return
		}
	}
	return
}

// V1GetAttributeForEvent returns the AttributeProfile that matches the event
func (alS *AttributeService) V1GetAttributeForEvent(args *AttrArgsProcessEvent,
	attrPrfl *AttributeProfile) (err error) {
	if args.CGREvent == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	attrPrf, err := alS.attributeProfileForEvent(args, utils.MapStorage{
		utils.MetaReq:  args.CGREvent.Event,
		utils.MetaOpts: args.Opts,
		utils.MetaVars: utils.MapStorage{
			utils.ProcessRuns: utils.NewNMData(0),
		},
	}, utils.EmptyString)
	if err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*attrPrfl = *attrPrf
	return
}

// V1ProcessEvent proccess the event and returns the result
func (alS *AttributeService) V1ProcessEvent(args *AttrArgsProcessEvent,
	reply *AttrSProcessEventReply) (err error) {
	if args.CGREvent == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	if args.Event == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	processRuns := alS.cgrcfg.AttributeSCfg().ProcessRuns
	if args.ProcessRuns != nil && *args.ProcessRuns != 0 {
		processRuns = *args.ProcessRuns
	}
	args.CGREvent = args.CGREvent.Clone()
	eNV := utils.MapStorage{
		utils.MetaReq:  args.CGREvent.Event,
		utils.MetaOpts: args.Opts,
		utils.MetaVars: utils.MapStorage{
			utils.ProcessRuns: utils.NewNMData(0),
		},
	}
	var lastID string
	matchedIDs := make([]string, 0, processRuns)
	alteredFields := make(utils.StringSet)
	for i := 0; i < processRuns; i++ {
		(eNV[utils.MetaVars].(utils.MapStorage))[utils.ProcessRuns] = utils.NewNMData(i + 1)
		var evRply *AttrSProcessEventReply
		evRply, err = alS.processEvent(args, eNV, lastID)
		if err != nil {
			if err != utils.ErrNotFound {
				err = utils.NewErrServerError(err)
			} else if i != 0 { // ignore "not found" in a loop different than 0
				err = nil
			}
			break
		}
		args.CGREvent.Tenant = evRply.CGREvent.Tenant
		lastID = evRply.MatchedProfiles[0]
		matchedIDs = append(matchedIDs, lastID)
		for _, fldName := range evRply.AlteredFields {
			alteredFields.Add(fldName)
		}
		if evRply.blocker {
			break
		}
	}
	if err == nil || err == utils.ErrNotFound {
		// Make sure the requested fields were populated
		for field, val := range args.CGREvent.Event {
			if val == utils.MetaAttributes {
				// mandatory IE missing
				err = utils.NewErrMandatoryIeMissing(field)
				return
			}
		}
	} else {
		return
	}

	*reply = AttrSProcessEventReply{
		MatchedProfiles: matchedIDs,
		AlteredFields:   alteredFields.AsSlice(),
		CGREvent:        args.CGREvent,
		Opts:            args.Opts,
	}
	return
}
