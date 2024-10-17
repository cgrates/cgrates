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
	"slices"

	"github.com/nyaruka/phonenumbers"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func newDynamicDP(resConns, stsConns, apiConns, trdConns []string,
	tenant string, initialDP utils.DataProvider) *dynamicDP {
	return &dynamicDP{
		resConns:  resConns,
		stsConns:  stsConns,
		apiConns:  apiConns,
		trdConns:  trdConns,
		tenant:    tenant,
		initialDP: initialDP,
		cache:     utils.MapStorage{},
	}
}

type dynamicDP struct {
	resConns  []string
	stsConns  []string
	apiConns  []string
	trdConns  []string
	tenant    string
	initialDP utils.DataProvider

	cache utils.MapStorage
}

func (dDP *dynamicDP) String() string { return dDP.initialDP.String() }

func (dDP *dynamicDP) FieldAsString(fldPath []string) (string, error) {
	val, err := dDP.FieldAsInterface(fldPath)
	if err != nil {
		return "", err
	}
	return utils.IfaceAsString(val), nil
}

var initialDPPrefixes = utils.NewStringSet([]string{
	utils.MetaReq, utils.MetaVars, utils.MetaCgreq,
	utils.MetaCgrep, utils.MetaRep, utils.MetaAct,
	utils.MetaEC, utils.MetaUCH, utils.MetaOpts,
	utils.MetaHdr, utils.MetaTrl, utils.MetaCfg,
	utils.MetaTenant, utils.MetaEventTimestamp})

func (dDP *dynamicDP) FieldAsInterface(fldPath []string) (val any, err error) {
	if len(fldPath) == 0 {
		return nil, utils.ErrNotFound
	}

	// Ensure type for supported path elements to allow calling their specific
	// FieldAsInterface method.
	if len(fldPath) > 3 && fldPath[0] == utils.MetaReq &&
		slices.Contains([]string{utils.CostDetails, utils.AccountSummary}, fldPath[1]) {
		if mp, canCast := dDP.initialDP.(utils.MapStorage); canCast {
			if event, canCast := mp[utils.MetaReq].(map[string]any); canCast {
				if field, has := event[fldPath[1]]; has {
					switch fldPath[1] {
					case utils.CostDetails:
						return processEventCostField(fldPath[2:], field, event)
					case utils.AccountSummary:
						return processAccountSummaryField(fldPath[2:], field, event)
					}
				}
			}
		}
	}

	if initialDPPrefixes.Has(fldPath[0]) {
		return dDP.initialDP.FieldAsInterface(fldPath)
	}
	val, err = dDP.cache.FieldAsInterface(fldPath)
	if err == utils.ErrNotFound { // in case not found in cache try to populate it
		return dDP.fieldAsInterface(fldPath)
	}
	return
}

func (dDP *dynamicDP) fieldAsInterface(fldPath []string) (val any, err error) {
	if len(fldPath) < 2 {
		return nil, fmt.Errorf("invalid fieldname <%s>", fldPath)
	}
	switch fldPath[0] {
	case utils.MetaAccounts:
		// sample of fieldName : ~*accounts.1001.BalanceMap.*monetary[0].Value
		// split the field name in 3 parts
		// fieldNameType (~*accounts), accountID(1001) and queried part (BalanceMap.*monetary[0].Value)

		var account Account
		if err = connMgr.Call(context.TODO(), dDP.apiConns, utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: dDP.tenant, Account: fldPath[1]}, &account); err != nil {
			return
		}
		//construct dataProvider from account and set it further
		dDP.cache.Set(fldPath[:2], account)
		return account.FieldAsInterface(fldPath[2:])
	case utils.MetaResources:
		// sample of fieldName : ~*resources.ResourceID.Field
		var reply ResourceWithConfig
		if err := connMgr.Call(context.TODO(), dDP.resConns, utils.ResourceSv1GetResourceWithConfig,
			&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: dDP.tenant, ID: fldPath[1]}}, &reply); err != nil {
			return nil, err
		}
		dp := config.NewObjectDP(&reply)
		dDP.cache.Set(fldPath[:2], dp)
		return dp.FieldAsInterface(fldPath[2:])
	case utils.MetaStats:
		// sample of fieldName : ~*stats.StatID.*acd
		var statValues map[string]float64

		if err := connMgr.Call(context.TODO(), dDP.stsConns, utils.StatSv1GetQueueFloatMetrics,
			&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: dDP.tenant, ID: fldPath[1]}},
			&statValues); err != nil {
			return nil, err
		}
		for k, v := range statValues {
			dDP.cache.Set([]string{utils.MetaStats, fldPath[1], k}, v)
		}
		return dDP.cache.FieldAsInterface(fldPath)
	case utils.MetaTrends:
		//sample of fieldName : ~*trends.TrendID.*acd.Value
		var trendSum TrendSummary
		if err := connMgr.Call(context.TODO(), dDP.trdConns, utils.TrendSv1GetTrendSummary, &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: dDP.tenant, ID: fldPath[1]}}, &trendSum); err != nil {
			return nil, err
		}
		dp := config.NewObjectDP(trendSum)
		dDP.cache.Set(fldPath[:2], dp)
		return dp.FieldAsInterface(fldPath[2:])
	case utils.MetaLibPhoneNumber:
		// sample of fieldName ~*libphonenumber.<~*req.Destination>
		// or ~*libphonenumber.<~*req.Destination>.Carrier
		dp, err := newLibPhoneNumberDP(fldPath[1])
		if err != nil {
			return nil, err
		}
		dDP.cache.Set(fldPath[:2], dp)
		return dp.FieldAsInterface(fldPath[2:])
	case utils.MetaAsm:
		// sample of fieldName ~*asm.BalanceSummaries.HolidayBalance.Value
		stringReq, err := dDP.initialDP.FieldAsString([]string{utils.MetaReq})
		if err != nil {
			return nil, err
		}
		acntSummary, err := NewAccountSummaryFromJSON(stringReq)
		if err != nil {
			return nil, err
		}
		dDP.cache.Set(fldPath[:1], acntSummary)
		return acntSummary.FieldAsInterface(fldPath[1:])
	default: // in case of constant we give an empty DataProvider ( empty navigable map )
	}
	return nil, utils.ErrNotFound
}

func newLibPhoneNumberDP(number string) (dp utils.DataProvider, err error) {
	num, err := phonenumbers.ParseAndKeepRawInput(number, utils.EmptyString)
	if err != nil {
		return nil, err
	}
	return &libphonenumberDP{pNumber: num, cache: make(utils.MapStorage)}, nil

}

type libphonenumberDP struct {
	pNumber *phonenumbers.PhoneNumber
	cache   utils.MapStorage
}

func (dDP *libphonenumberDP) String() string { return dDP.pNumber.String() }

func (dDP *libphonenumberDP) FieldAsString(fldPath []string) (string, error) {
	val, err := dDP.FieldAsInterface(fldPath)
	if err != nil {
		return "", err
	}
	return utils.IfaceAsString(val), nil
}

func (dDP *libphonenumberDP) FieldAsInterface(fldPath []string) (val any, err error) {
	if len(fldPath) == 0 {
		dDP.setDefaultFields()
		val = dDP.cache
		return
	}
	if val, err = dDP.cache.FieldAsInterface(fldPath); err == utils.ErrNotFound { // in case not found in cache try to populate it
		return dDP.fieldAsInterface(fldPath)
	}
	return
}

func (dDP *libphonenumberDP) fieldAsInterface(fldPath []string) (val any, err error) {
	if len(fldPath) != 1 {
		return nil, fmt.Errorf("invalid field path <%+v> for libphonenumberDP", fldPath)
	}
	switch fldPath[0] {
	case "CountryCode":
		val = dDP.pNumber.GetCountryCode()
	case "NationalNumber":
		val = dDP.pNumber.GetNationalNumber()
	case "Region":
		val = phonenumbers.GetRegionCodeForNumber(dDP.pNumber)
	case "NumberType":
		val = phonenumbers.GetNumberType(dDP.pNumber)
	case "GeoLocation":
		regCode := phonenumbers.GetRegionCodeForNumber(dDP.pNumber)
		geoLocation, err := phonenumbers.GetGeocodingForNumber(dDP.pNumber, regCode)
		if err != nil {
			utils.Logger.Warning(fmt.Sprintf("Received error: <%+v> when getting GeoLocation for number %+v", err, dDP.pNumber))
		}
		val = geoLocation
	case "Carrier":
		carrier, err := phonenumbers.GetCarrierForNumber(dDP.pNumber, phonenumbers.GetRegionCodeForNumber(dDP.pNumber))
		if err != nil {
			utils.Logger.Warning(fmt.Sprintf("Received error: <%+v> when getting Carrier for number %+v", err, dDP.pNumber))
		}
		val = carrier
	case "LengthOfNationalDestinationCode":
		val = phonenumbers.GetLengthOfNationalDestinationCode(dDP.pNumber)
	case "RawInput":
		val = dDP.pNumber.GetRawInput()
	case "Extension":
		val = dDP.pNumber.GetExtension()
	case "NumberOfLeadingZeros":
		val = dDP.pNumber.GetNumberOfLeadingZeros()
	case "ItalianLeadingZero":
		val = dDP.pNumber.GetItalianLeadingZero()
	case "PreferredDomesticCarrierCode":
		val = dDP.pNumber.GetPreferredDomesticCarrierCode()
	case "CountryCodeSource":
		val = dDP.pNumber.GetCountryCodeSource()

	}
	dDP.cache[fldPath[0]] = val
	return
}

func (dDP *libphonenumberDP) setDefaultFields() {
	if _, has := dDP.cache["CountryCode"]; !has {
		dDP.cache["CountryCode"] = dDP.pNumber.GetCountryCode()
	}
	if _, has := dDP.cache["NationalNumber"]; !has {
		dDP.cache["NationalNumber"] = dDP.pNumber.GetNationalNumber()
	}
	if _, has := dDP.cache["Region"]; !has {
		dDP.cache["Region"] = phonenumbers.GetRegionCodeForNumber(dDP.pNumber)
	}
	if _, has := dDP.cache["NumberType"]; !has {
		dDP.cache["NumberType"] = phonenumbers.GetNumberType(dDP.pNumber)
	}
	if _, has := dDP.cache["GeoLocation"]; !has {
		geoLocation, err := phonenumbers.GetGeocodingForNumber(dDP.pNumber, phonenumbers.GetRegionCodeForNumber(dDP.pNumber))
		if err != nil {
			utils.Logger.Warning(fmt.Sprintf("Received error: <%+v> when getting GeoLocation for number %+v", err, dDP.pNumber))
		}
		dDP.cache["GeoLocation"] = geoLocation
	}
	if _, has := dDP.cache["Carrier"]; !has {
		carrier, err := phonenumbers.GetCarrierForNumber(dDP.pNumber, phonenumbers.GetRegionCodeForNumber(dDP.pNumber))
		if err != nil {
			utils.Logger.Warning(fmt.Sprintf("Received error: <%+v> when getting Carrier for number %+v", err, dDP.pNumber))
		}
		dDP.cache["Carrier"] = carrier
	}
	if _, has := dDP.cache["LengthOfNationalDestinationCode"]; !has {
		dDP.cache["LengthOfNationalDestinationCode"] = phonenumbers.GetLengthOfNationalDestinationCode(dDP.pNumber)
	}
	if _, has := dDP.cache["RawInput"]; !has {
		dDP.cache["RawInput"] = dDP.pNumber.GetRawInput()
	}
	if _, has := dDP.cache["Extension"]; !has {
		dDP.cache["Extension"] = dDP.pNumber.GetExtension()
	}
	if _, has := dDP.cache["NumberOfLeadingZeros"]; !has {
		dDP.cache["NumberOfLeadingZeros"] = dDP.pNumber.GetNumberOfLeadingZeros()
	}
	if _, has := dDP.cache["ItalianLeadingZero"]; !has {
		dDP.cache["ItalianLeadingZero"] = dDP.pNumber.GetItalianLeadingZero()
	}
	if _, has := dDP.cache["PreferredDomesticCarrierCode"]; !has {
		dDP.cache["PreferredDomesticCarrierCode"] = dDP.pNumber.GetPreferredDomesticCarrierCode()
	}
	if _, has := dDP.cache["CountryCodeSource"]; !has {
		dDP.cache["CountryCodeSource"] = dDP.pNumber.GetCountryCodeSource()
	}
}
