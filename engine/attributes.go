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
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// NewAttributeService returns a new AttributeService
func NewAttributeService(dm *DataManager, filterS *FilterS,
	cgrcfg *config.CGRConfig) *AttributeService {
	return &AttributeService{
		dm:      dm,
		filterS: filterS,
		cgrcfg:  cgrcfg,
	}
}

// AttributeService the service for the API
type AttributeService struct {
	dm      *DataManager
	filterS *FilterS
	cgrcfg  *config.CGRConfig
}

// Shutdown is called to shutdown the service
func (alS *AttributeService) Shutdown() {
	utils.Logger.Info(fmt.Sprintf("<%s> shutdown initialized", utils.AttributeS))
	utils.Logger.Info(fmt.Sprintf("<%s> shutdown complete", utils.AttributeS))
}

// attributeProfileForEvent returns the matching attribute
func (alS *AttributeService) attributeProfileForEvent(tnt string, ctx *string, attrsIDs []string, actTime *time.Time, evNm utils.MapStorage,
	lastID string, processedPrfNo map[string]int, profileRuns int, ignoreFilters bool) (matchAttrPrfl *AttributeProfile, err error) {
	var attrIDs []string
	contextVal := utils.MetaDefault
	if ctx != nil && *ctx != "" {
		contextVal = *ctx
	}
	attrIdxKey := utils.ConcatenatedKey(tnt, contextVal)
	if len(attrsIDs) != 0 {
		attrIDs = attrsIDs
	} else {
		ignoreFilters = false
		aPrflIDs, err := MatchingItemIDsForEvent(evNm,
			alS.cgrcfg.AttributeSCfg().StringIndexedFields,
			alS.cgrcfg.AttributeSCfg().PrefixIndexedFields,
			alS.cgrcfg.AttributeSCfg().SuffixIndexedFields,
			alS.dm, utils.CacheAttributeFilterIndexes, attrIdxKey,
			alS.cgrcfg.AttributeSCfg().IndexedSelects,
			alS.cgrcfg.AttributeSCfg().NestedFields,
		)
		if err != nil &&
			err != utils.ErrNotFound {
			return nil, err
		}
		if err == utils.ErrNotFound ||
			alS.cgrcfg.AttributeSCfg().AnyContext {
			aPrflAnyIDs, err := MatchingItemIDsForEvent(evNm,
				alS.cgrcfg.AttributeSCfg().StringIndexedFields,
				alS.cgrcfg.AttributeSCfg().PrefixIndexedFields,
				alS.cgrcfg.AttributeSCfg().SuffixIndexedFields,
				alS.dm, utils.CacheAttributeFilterIndexes,
				utils.ConcatenatedKey(tnt, utils.MetaAny),
				alS.cgrcfg.AttributeSCfg().IndexedSelects,
				alS.cgrcfg.AttributeSCfg().NestedFields)
			if aPrflIDs.Size() == 0 {
				if err != nil { // return the error if no attribute matched the needed context
					return nil, err
				}
				aPrflIDs = aPrflAnyIDs
			} else if err == nil && aPrflAnyIDs.Size() != 0 {
				aPrflIDs = utils.JoinStringSet(aPrflIDs, aPrflAnyIDs)
			}
		}
		attrIDs = aPrflIDs.AsSlice()
	}
	for _, apID := range attrIDs {
		aPrfl, err := alS.dm.GetAttributeProfile(tnt, apID, true, true, utils.NonTransactional)
		if err != nil {
			if err == utils.ErrNotFound {
				continue
			}
			return nil, err
		}
		if !(len(aPrfl.Contexts) == 1 && aPrfl.Contexts[0] == utils.MetaAny) &&
			!slices.Contains(aPrfl.Contexts, contextVal) {
			continue
		}
		if aPrfl.ActivationInterval != nil && actTime != nil &&
			!aPrfl.ActivationInterval.IsActiveAtTime(*actTime) { // not active
			continue
		}
		tntID := aPrfl.TenantIDInline()
		(evNm[utils.MetaVars].(utils.MapStorage))[utils.MetaAttrPrfTenantID] = tntID
		if !ignoreFilters {
			if pass, err := alS.filterS.Pass(tnt, aPrfl.FilterIDs,
				evNm); err != nil {
				return nil, err
			} else if !pass {
				continue
			}
		}
		if (matchAttrPrfl == nil || matchAttrPrfl.Weight < aPrfl.Weight) &&
			tntID != lastID &&
			(profileRuns <= 0 || processedPrfNo[tntID] < profileRuns) {
			matchAttrPrfl = aPrfl
		}
	}
	// All good, convert from Map to Slice so we can sort
	if matchAttrPrfl == nil {
		return nil, utils.ErrNotFound
	}
	(evNm[utils.MetaVars].(utils.MapStorage))[utils.MetaAttrPrfTenantID] = matchAttrPrfl.TenantIDInline()
	return
}

// AttrSProcessEventReply reply used for proccess event
type AttrSProcessEventReply struct {
	MatchedProfiles []string
	AlteredFields   []string
	*utils.CGREvent
	blocker bool // internally used to stop further processRuns
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
			rplyDigest += utils.FieldsSep
		}
		fldStrVal, _ := attrReply.CGREvent.FieldAsString(fld)
		rplyDigest += fld + utils.InInFieldSep + fldStrVal
	}
	return
}

// processEvent will match event with attribute profile and do the necessary replacements
func (alS *AttributeService) processEvent(tnt string, args *utils.CGREvent, evNm utils.MapStorage, dynDP utils.DataProvider,
	lastID string, processedPrfNo map[string]int, profileRuns int) (
	rply *AttrSProcessEventReply, err error) {
	context := alS.cgrcfg.AttributeSCfg().Opts.Context
	if opt, has := args.APIOpts[utils.OptsContext]; has {
		context = utils.StringPointer(utils.IfaceAsString(opt))
	}
	var attrIDs []string
	if attrIDs, err = utils.GetStringSliceOpts(args, alS.cgrcfg.AttributeSCfg().Opts.ProfileIDs, utils.OptsAttributesProfileIDs); err != nil {
		return
	}
	var ignFilters bool
	if ignFilters, err = utils.GetBoolOpts(args, alS.cgrcfg.AttributeSCfg().Opts.ProfileIgnoreFilters,
		utils.OptsAttributesProfileIgnoreFilters); err != nil {
		return
	}
	var attrPrf *AttributeProfile
	if attrPrf, err = alS.attributeProfileForEvent(tnt, context, attrIDs, args.Time, evNm, lastID, processedPrfNo, profileRuns, ignFilters); err != nil {
		return
	}
	rply = &AttrSProcessEventReply{
		MatchedProfiles: []string{attrPrf.TenantIDInline()},
		CGREvent:        args,
		blocker:         attrPrf.Blocker,
	}
	rply.Tenant = tnt
	for _, attribute := range attrPrf.Attributes {
		//in case that we have filter for attribute send them to FilterS to be processed
		if len(attribute.FilterIDs) != 0 {
			var pass bool
			if pass, err = alS.filterS.Pass(tnt, attribute.FilterIDs,
				evNm); err != nil {
				return
			} else if !pass {
				continue
			}
		}
		var out any
		if out, err = ParseAttribute(dynDP, utils.FirstNonEmpty(attribute.Type, utils.MetaVariable), utils.DynamicDataPrefix+attribute.Path, attribute.Value, alS.cgrcfg.GeneralCfg().RoundingDecimals, alS.cgrcfg.GeneralCfg().DefaultTimezone, time.RFC3339, alS.cgrcfg.GeneralCfg().RSRSep); err != nil {
			rply = nil
			return
		}
		substitute := utils.IfaceAsString(out)
		//add only once the Path in AlteredFields
		if !slices.Contains(rply.AlteredFields, attribute.Path) {
			rply.AlteredFields = append(rply.AlteredFields, attribute.Path)
		}
		if attribute.Path == utils.MetaTenant {
			if attribute.Type == utils.MetaComposed {
				rply.CGREvent.Tenant += substitute
			} else {
				rply.CGREvent.Tenant = substitute
			}
			evNm[utils.MetaTenant] = substitute
			continue
		}
		if substitute == utils.MetaRemove {
			evNm.Remove(utils.SplitPath(attribute.Path, utils.NestingSep[0], -1))
			continue
		}
		if attribute.Type == utils.MetaComposed {
			var val string
			if val, err = evNm.FieldAsString(utils.SplitPath(attribute.Path, utils.NestingSep[0], -1)); err != nil && err != utils.ErrNotFound {
				rply = nil
				return
			}
			substitute = val + substitute
		}
		if err = evNm.Set(utils.SplitPath(attribute.Path, utils.NestingSep[0], -1), substitute); err != nil {
			rply = nil
			return
		}
	}
	return
}

// V1GetAttributeForEvent returns the AttributeProfile that matches the event
func (alS *AttributeService) V1GetAttributeForEvent(ctx *context.Context, args *utils.CGREvent,
	attrPrfl *AttributeProfile) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = alS.cgrcfg.GeneralCfg().DefaultTenant
	}
	context := alS.cgrcfg.AttributeSCfg().Opts.Context
	if opt, has := args.APIOpts[utils.OptsContext]; has {
		context = utils.StringPointer(utils.IfaceAsString(opt))
	}
	var attrIDs []string
	if attrIDs, err = utils.GetStringSliceOpts(args, alS.cgrcfg.AttributeSCfg().Opts.ProfileIDs, utils.OptsAttributesProfileIDs); err != nil {
		return
	}
	var ignFilters bool
	if ignFilters, err = utils.GetBoolOpts(args, alS.cgrcfg.AttributeSCfg().Opts.ProfileIgnoreFilters,
		utils.OptsAttributesProfileIgnoreFilters); err != nil {
		return
	}
	attrPrf, err := alS.attributeProfileForEvent(tnt, context, attrIDs, args.Time, utils.MapStorage{
		utils.MetaReq:  args.Event,
		utils.MetaOpts: args.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.MetaProcessRuns: 0,
		},
	}, utils.EmptyString, make(map[string]int), 0, ignFilters)
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
func (alS *AttributeService) V1ProcessEvent(ctx *context.Context, args *utils.CGREvent,
	reply *AttrSProcessEventReply) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	if args.Event == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = alS.cgrcfg.GeneralCfg().DefaultTenant
	}
	var processRuns int
	if processRuns, err = utils.GetIntOpts(args, alS.cgrcfg.AttributeSCfg().Opts.ProcessRuns,
		utils.OptsAttributesProcessRuns); err != nil {
		return
	}
	var profileRuns int
	if profileRuns, err = utils.GetIntOpts(args, alS.cgrcfg.AttributeSCfg().Opts.ProfileRuns,
		utils.OptsAttributesProfileRuns); err != nil {
		return
	}
	args = args.Clone()
	processedPrf := make(utils.StringSet)
	processedPrfNo := make(map[string]int)
	eNV := utils.MapStorage{
		utils.MetaReq:  args.Event,
		utils.MetaOpts: args.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.MetaProcessRuns:         0,
			utils.MetaProcessedProfileIDs: processedPrf,
		},
		utils.MetaTenant: tnt,
	}
	var lastID string
	matchedIDs := make([]string, 0, processRuns)
	alteredFields := make(utils.StringSet)
	dynDP := newDynamicDP(alS.cgrcfg.AttributeSCfg().ResourceSConns,
		alS.cgrcfg.AttributeSCfg().StatSConns, alS.cgrcfg.AttributeSCfg().ApierSConns, nil, args.Tenant, eNV)
	for i := 0; i < processRuns; i++ {
		(eNV[utils.MetaVars].(utils.MapStorage))[utils.MetaProcessRuns] = i + 1
		var evRply *AttrSProcessEventReply
		evRply, err = alS.processEvent(tnt, args, eNV, dynDP, lastID, processedPrfNo, profileRuns)
		if err != nil {
			if err != utils.ErrNotFound {
				err = utils.NewErrServerError(err)
			} else if i != 0 { // ignore "not found" in a loop different than 0
				err = nil
			}
			break
		}
		args.Tenant = evRply.CGREvent.Tenant
		tnt = evRply.CGREvent.Tenant
		lastID = evRply.MatchedProfiles[0]
		matchedIDs = append(matchedIDs, lastID)
		processedPrf.Add(lastID)
		processedPrfNo[lastID] = processedPrfNo[lastID] + 1
		for _, fldName := range evRply.AlteredFields {
			alteredFields.Add(fldName)
		}
		if evRply.blocker {
			break
		}
	}
	if err == nil || err == utils.ErrNotFound {
		// Make sure the requested fields were populated
		for field, val := range args.Event {
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
		CGREvent:        args,
	}
	return
}

func ParseAttribute(dp utils.DataProvider, attrType, path string, value config.RSRParsers, roundingDec int, timeZone, layout, rsrSep string) (
	out any, err error) {
	switch attrType {
	case utils.MetaNone:
		return
	case utils.MetaConstant:
		out, err = value.ParseValue(utils.EmptyString)
	case utils.MetaVariable, utils.MetaComposed:
		out, err = value.ParseDataProvider(dp)
	case utils.MetaUsageDifference:
		if len(value) != 2 {
			return "", fmt.Errorf("invalid arguments <%s> to %s",
				utils.ToJSON(value), utils.MetaUsageDifference)
		}
		var strVal1 string
		if strVal1, err = value[0].ParseDataProvider(dp); err != nil {
			return
		}
		var strVal2 string
		if strVal2, err = value[1].ParseDataProvider(dp); err != nil {
			return
		}
		var tEnd time.Time
		if tEnd, err = utils.ParseTimeDetectLayout(strVal1, utils.EmptyString); err != nil {
			return
		}
		var tStart time.Time
		if tStart, err = utils.ParseTimeDetectLayout(strVal2, utils.EmptyString); err != nil {
			return
		}
		out = tEnd.Sub(tStart).String()
	case utils.MetaSum:
		var ifaceVals []any
		if ifaceVals, err = value.GetIfaceFromValues(dp); err != nil {
			return
		}
		out, err = utils.Sum(ifaceVals...)
	case utils.MetaDifference:
		var ifaceVals []any
		if ifaceVals, err = value.GetIfaceFromValues(dp); err != nil {
			return
		}
		out, err = utils.Difference(timeZone, ifaceVals...)
	case utils.MetaMultiply:
		var ifaceVals []any
		if ifaceVals, err = value.GetIfaceFromValues(dp); err != nil {
			return
		}
		out, err = utils.Multiply(ifaceVals...)
	case utils.MetaDivide:
		var ifaceVals []any
		if ifaceVals, err = value.GetIfaceFromValues(dp); err != nil {
			return
		}
		out, err = utils.Divide(ifaceVals...)
	case utils.MetaValueExponent:
		if len(value) != 2 {
			return "", fmt.Errorf("invalid arguments <%s> to %s",
				utils.ToJSON(value), utils.MetaValueExponent)
		}
		var strVal1 string
		if strVal1, err = value[0].ParseDataProvider(dp); err != nil {
			return
		}
		var val float64
		if val, err = strconv.ParseFloat(strVal1, 64); err != nil {
			return "", fmt.Errorf("invalid value <%s> to %s",
				strVal1, utils.MetaValueExponent)
		}
		var strVal2 string
		if strVal2, err = value[1].ParseDataProvider(dp); err != nil {
			return
		}
		var exp int
		if exp, err = strconv.Atoi(strVal2); err != nil {
			return
		}
		out = strconv.FormatFloat(utils.Round(val*math.Pow10(exp),
			roundingDec, utils.MetaRoundingMiddle), 'f', -1, 64)
	case utils.MetaUnixTimestamp:
		var val string
		if val, err = value.ParseDataProvider(dp); err != nil {
			return
		}
		var t time.Time
		if t, err = utils.ParseTimeDetectLayout(val, timeZone); err != nil {
			return
		}
		out = strconv.Itoa(int(t.Unix()))
	case utils.MetaDateTime: // Convert the requested field value into datetime with layout
		var val string
		if val, err = value.ParseDataProvider(dp); err != nil {
			return
		}
		var dtFld time.Time
		dtFld, err = utils.ParseTimeDetectLayout(val, timeZone)
		if err != nil {
			return
		}
		out = dtFld.Format(layout)
	case utils.MetaPrefix:
		var pathRsr config.RSRParsers
		pathRsr, err = config.NewRSRParsers(path, rsrSep)
		if err != nil {
			return
		}
		var pathVal string
		if pathVal, err = pathRsr.ParseDataProvider(dp); err != nil {
			return
		}
		var val string
		if val, err = value.ParseDataProvider(dp); err != nil {
			return
		}
		out = val + pathVal
	case utils.MetaSuffix:
		var pathRsr config.RSRParsers
		pathRsr, err = config.NewRSRParsers(path, rsrSep)
		if err != nil {
			return
		}
		var pathVal string
		if pathVal, err = pathRsr.ParseDataProvider(dp); err != nil {
			return
		}
		var val string
		if val, err = value.ParseDataProvider(dp); err != nil {
			return
		}
		out = pathVal + val
	case utils.MetaCCUsage:
		if len(value) != 3 {
			return nil, fmt.Errorf("invalid arguments <%s> to %s",
				utils.ToJSON(value), utils.MetaCCUsage)
		}
		var strVal1 string
		if strVal1, err = value[0].ParseDataProvider(dp); err != nil {
			return
		}
		var reqNr int64
		if reqNr, err = strconv.ParseInt(strVal1, 10, 64); err != nil {
			err = fmt.Errorf("invalid requestNumber <%s> to %s",
				strVal1, utils.MetaCCUsage)
			return
		}
		var strVal2 string
		if strVal2, err = value[1].ParseDataProvider(dp); err != nil {
			return
		}
		var usedCCTime time.Duration
		if usedCCTime, err = utils.ParseDurationWithNanosecs(strVal2); err != nil {
			err = fmt.Errorf("invalid usedCCTime <%s> to %s",
				strVal2, utils.MetaCCUsage)
			return
		}
		var strVal3 string
		if strVal3, err = value[2].ParseDataProvider(dp); err != nil {
			return
		}
		var debitItvl time.Duration
		if debitItvl, err = utils.ParseDurationWithNanosecs(strVal3); err != nil {
			err = fmt.Errorf("invalid debitInterval <%s> to %s",
				strVal3, utils.MetaCCUsage)
			return
		}
		if reqNr--; reqNr < 0 { // terminate will be ignored (init request should always be 0)
			reqNr = 0
		}
		return usedCCTime + time.Duration(debitItvl.Nanoseconds()*reqNr), nil
	case utils.MetaSIPCID:
		if len(value) < 1 {
			return nil, fmt.Errorf("invalid number of arguments <%s> to %s",
				utils.ToJSON(value), utils.MetaSIPCID)
		}
		values := make([]string, 1, len(value))
		if values[0], err = value[0].ParseDataProvider(dp); err != nil {
			return
		}
		for _, val := range value[1:] {
			var valStr string
			if valStr, err = val.ParseDataProvider(dp); err != nil && err != utils.ErrNotFound {
				return
			}
			if len(valStr) != 0 && err != utils.ErrNotFound {
				values = append(values, valStr)
			}
		}

		sort.Strings(values[1:])
		out = strings.Join(values, utils.InfieldSep)
	default:
		if strings.HasPrefix(attrType, utils.MetaHTTP) {
			out, err = externalAttributeAPI(attrType, dp)
			break
		}
		return utils.EmptyString, fmt.Errorf("unsupported type: <%s>", attrType)
	}
	return
}
