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
	cgrcfg *config.CGRConfig) *AttributeS {
	return &AttributeS{
		dm:    dm,
		fltrS: filterS,
		cfg:   cgrcfg,
	}
}

// AttributeS the service for the API
type AttributeS struct {
	dm    *DataManager
	fltrS *FilterS
	cfg   *config.CGRConfig
}

// attributeProfileForEvent returns the matching attribute
func (alS *AttributeS) attributeProfileForEvent(ctx *context.Context, tnt string, attrIDs []string,
	evNm utils.MapStorage, lastID string, processedPrfNo map[string]int, profileRuns int, ignoreFilters bool) (matchAttrPrfl *AttributeProfile, err error) {
	if len(attrIDs) == 0 {
		ignoreFilters = false
		aPrflIDs, err := MatchingItemIDsForEvent(ctx, evNm,
			alS.cfg.AttributeSCfg().StringIndexedFields,
			alS.cfg.AttributeSCfg().PrefixIndexedFields,
			alS.cfg.AttributeSCfg().SuffixIndexedFields,
			alS.cfg.AttributeSCfg().ExistsIndexedFields,
			alS.cfg.AttributeSCfg().NotExistsIndexedFields,
			alS.dm, utils.CacheAttributeFilterIndexes, tnt,
			alS.cfg.AttributeSCfg().IndexedSelects,
			alS.cfg.AttributeSCfg().NestedFields,
		)
		if err != nil {
			return nil, err
		}
		attrIDs = aPrflIDs.AsSlice()
	}
	var apWw *apWithWeight
	for _, apID := range attrIDs {
		var aPrfl *AttributeProfile
		aPrfl, err = alS.dm.GetAttributeProfile(ctx, tnt, apID, true, true, utils.NonTransactional)
		if err != nil {
			if err == utils.ErrNotFound {
				continue
			}
			return nil, err
		}
		tntID := aPrfl.TenantIDInline()
		evNm[utils.MetaVars].(utils.MapStorage)[utils.MetaAttrPrfTenantID] = tntID
		if !ignoreFilters {
			if pass, err := alS.fltrS.Pass(ctx, tnt, aPrfl.FilterIDs,
				evNm); err != nil {
				return nil, err
			} else if !pass {
				continue
			}
		}

		var apfWeight float64
		if apfWeight, err = WeightFromDynamics(ctx, aPrfl.Weights,
			alS.fltrS, tnt, evNm); err != nil {
			return
		}
		if (apWw == nil || apWw.weight < apfWeight) &&
			tntID != lastID &&
			(profileRuns <= 0 || processedPrfNo[tntID] < profileRuns) {
			apWw = &apWithWeight{aPrfl, apfWeight}
		}
	}
	// All good, convert from Map to Slice so we can sort
	if apWw == nil {
		return nil, utils.ErrNotFound
	}
	evNm[utils.MetaVars].(utils.MapStorage)[utils.MetaAttrPrfTenantID] = apWw.AttributeProfile.TenantIDInline()
	return apWw.AttributeProfile, nil
}

type FieldsAltered struct {
	MatchedProfileID string
	Fields           []string
}

// UniqueAlteredFields will return all altered fields without duplicates
func (flds *AttrSProcessEventReply) UniqueAlteredFields() (unFlds utils.StringSet) {
	unFlds = make(utils.StringSet)
	for _, altered := range flds.AlteredFields {
		unFlds.AddSlice(altered.Fields)
	}
	return
}

// AttrSProcessEventReply reply used for proccess event
type AttrSProcessEventReply struct {
	AlteredFields []*FieldsAltered
	*utils.CGREvent
	blocker bool // internally used to stop further processRuns
}

// Digest returns serialized version of alteredFields in AttrSProcessEventReply
// format fldName1:fldVal1,fldName2:fldVal2
func (attrReply *AttrSProcessEventReply) Digest() (rplyDigest string) {
	for idx, altered := range attrReply.AlteredFields {
		for idxFlds, fldName := range altered.Fields {
			fldName = strings.TrimPrefix(fldName, utils.MetaReq+utils.NestingSep)
			if _, has := attrReply.CGREvent.Event[fldName]; !has {
				continue //maybe removed
			}
			if idx != 0 || idxFlds != 0 {
				rplyDigest += utils.FieldsSep
			}
			fldStrVal, _ := attrReply.CGREvent.FieldAsString(fldName)
			rplyDigest += fldName + utils.InInFieldSep + fldStrVal
		}
	}
	return
}

// processEvent will match event with attribute profile and do the necessary replacements
func (alS *AttributeS) processEvent(ctx *context.Context, tnt string, args *utils.CGREvent, evNm utils.MapStorage, dynDP utils.DataProvider,
	lastID string, processedPrfNo map[string]int, profileRuns int) (rply *AttrSProcessEventReply, err error) {
	var attrIDs []string
	if attrIDs, err = GetStringSliceOpts(ctx, args.Tenant, args, alS.fltrS, alS.cfg.AttributeSCfg().Opts.ProfileIDs,
		config.AttributesProfileIDsDftOpt, utils.OptsAttributesProfileIDs); err != nil {
		return
	}
	var ignFilters bool
	if ignFilters, err = GetBoolOpts(ctx, tnt, evNm, alS.fltrS, alS.cfg.AttributeSCfg().Opts.ProfileIgnoreFilters,
		config.AttributesProfileIgnoreFiltersDftOpt, utils.MetaProfileIgnoreFilters); err != nil {
		return
	}
	var attrPrf *AttributeProfile
	if attrPrf, err = alS.attributeProfileForEvent(ctx, tnt, attrIDs, evNm, lastID, processedPrfNo, profileRuns, ignFilters); err != nil {
		return
	}
	var blocker bool
	if blocker, err = BlockerFromDynamics(ctx, attrPrf.Blockers, alS.fltrS, tnt, evNm); err != nil {
		return
	}
	rply = &AttrSProcessEventReply{
		AlteredFields: []*FieldsAltered{{
			MatchedProfileID: attrPrf.TenantIDInline(),
			Fields:           []string{},
		}},
		CGREvent: args,
		blocker:  blocker,
	}
	rply.Tenant = tnt
	for _, attribute := range attrPrf.Attributes {
		//in case that we have filter for attribute send them to FilterS to be processed
		if len(attribute.FilterIDs) != 0 {
			var pass bool
			if pass, err = alS.fltrS.Pass(ctx, tnt, attribute.FilterIDs,
				evNm); err != nil {
				return
			} else if !pass {
				continue
			}
		}
		var out any
		if out, err = ParseAttribute(dynDP, utils.FirstNonEmpty(attribute.Type, utils.MetaVariable), utils.DynamicDataPrefix+attribute.Path, attribute.Value, alS.cfg.GeneralCfg().RoundingDecimals, alS.cfg.GeneralCfg().DefaultTimezone, time.RFC3339, alS.cfg.GeneralCfg().RSRSep); err != nil {
			rply = nil
			return
		}
		substitute := utils.IfaceAsString(out)
		//add only once the Path in AlteredFields
		if !slices.Contains(rply.AlteredFields[0].Fields, attribute.Path) {
			rply.AlteredFields[0].Fields = append(rply.AlteredFields[0].Fields, attribute.Path)
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
		var blocker bool
		if blocker, err = BlockerFromDynamics(ctx, attribute.Blockers, alS.fltrS, tnt, evNm); err != nil {
			rply = nil
			return
		}
		if blocker {
			break
		}
	}
	return
}

// V1GetAttributeForEvent returns the AttributeProfile that matches the event
func (alS *AttributeS) V1GetAttributeForEvent(ctx *context.Context, args *utils.CGREvent,
	attrPrfl *APIAttributeProfile) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = alS.cfg.GeneralCfg().DefaultTenant
	}
	var attrIDs []string
	if attrIDs, err = GetStringSliceOpts(ctx, args.Tenant, args, alS.fltrS, alS.cfg.AttributeSCfg().Opts.ProfileIDs,
		config.AttributesProfileIDsDftOpt, utils.OptsAttributesProfileIDs); err != nil {
		return
	}
	evNM := utils.MapStorage{
		utils.MetaReq:  args.Event,
		utils.MetaOpts: args.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}
	var ignFilters bool
	if ignFilters, err = GetBoolOpts(ctx, tnt, evNM, alS.fltrS, alS.cfg.AttributeSCfg().Opts.ProfileIgnoreFilters,
		config.AttributesProfileIgnoreFiltersDftOpt, utils.MetaProfileIgnoreFilters); err != nil {
		return
	}
	attrPrf, err := alS.attributeProfileForEvent(ctx, tnt, attrIDs, evNM,
		utils.EmptyString, make(map[string]int), 0, ignFilters)
	if err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*attrPrfl = *NewAPIAttributeProfile(attrPrf)
	return
}

// V1ProcessEvent proccess the event and returns the result
func (alS *AttributeS) V1ProcessEvent(ctx *context.Context, args *utils.CGREvent,
	reply *AttrSProcessEventReply) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = alS.cfg.GeneralCfg().DefaultTenant
	}

	var processRuns int
	if processRuns, err = GetIntOpts(ctx, tnt, args, alS.fltrS, alS.cfg.AttributeSCfg().Opts.ProcessRuns,
		config.AttributesProcessRunsDftOpt, utils.OptsAttributesProcessRuns); err != nil {
		return
	}

	var profileRuns int
	if profileRuns, err = GetIntOpts(ctx, tnt, args, alS.fltrS, alS.cfg.AttributeSCfg().Opts.ProfileRuns,
		config.AttributesProfileRunsDftOpt, utils.OptsAttributesProfileRuns); err != nil {
		return
	}
	args = args.Clone()
	processedPrf := make(utils.StringSet)
	processedPrfNo := make(map[string]int)
	eNV := utils.MapStorage{
		utils.MetaVars: utils.MapStorage{
			utils.MetaProcessRunsCfg:      0,
			utils.MetaProcessedProfileIDs: processedPrf,
		},
		utils.MetaTenant: tnt,
	}
	if args.APIOpts != nil {
		eNV[utils.MetaOpts] = args.APIOpts
	}
	if args.Event != nil {
		eNV[utils.MetaReq] = args.Event
	}

	var lastID string
	matchedIDs := []*FieldsAltered{}
	dynDP := newDynamicDP(ctx, alS.cfg.AttributeSCfg().ResourceSConns,
		alS.cfg.AttributeSCfg().StatSConns, alS.cfg.AttributeSCfg().AccountSConns, nil, nil, args.Tenant, eNV)
	for i := 0; i < processRuns; i++ {
		eNV[utils.MetaVars].(utils.MapStorage)[utils.MetaProcessRunsCfg] = i + 1
		var evRply *AttrSProcessEventReply
		evRply, err = alS.processEvent(ctx, tnt, args, eNV, dynDP, lastID, processedPrfNo, profileRuns)
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

		lastID = evRply.AlteredFields[0].MatchedProfileID
		altered := &FieldsAltered{
			MatchedProfileID: lastID,
			Fields:           make([]string, len(evRply.AlteredFields[0].Fields)),
		}
		processedPrf.Add(lastID)
		processedPrfNo[lastID] = processedPrfNo[lastID] + 1
		copy(altered.Fields, evRply.AlteredFields[0].Fields)
		matchedIDs = append(matchedIDs, altered)
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
		AlteredFields: matchedIDs,
		CGREvent:      args,
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
	case utils.MetaGeneric:
		out, err = value.ParseDataProviderWithInterfaces2(dp)
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
