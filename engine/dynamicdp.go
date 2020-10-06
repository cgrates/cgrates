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
	"net"

	"github.com/nyaruka/phonenumbers"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func newDynamicDP(resConns, stsConns, apiConns []string,
	tenant string, initialDP utils.DataProvider) *dynamicDP {
	return &dynamicDP{
		resConns:  resConns,
		stsConns:  stsConns,
		apiConns:  apiConns,
		tenant:    tenant,
		initialDP: initialDP,
		cache:     utils.MapStorage{},
	}
}

type dynamicDP struct {
	resConns  []string
	stsConns  []string
	apiConns  []string
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

func (dDP *dynamicDP) RemoteHost() net.Addr {
	return utils.LocalAddr()
}

var initialDPPrefixes = utils.NewStringSet([]string{utils.MetaReq, utils.MetaVars,
	utils.MetaCgreq, utils.MetaCgrep, utils.MetaRep, utils.MetaCGRAReq,
	utils.MetaAct, utils.MetaEC, utils.MetaUCH, utils.MetaOpts})

func (dDP *dynamicDP) FieldAsInterface(fldPath []string) (val interface{}, err error) {
	if len(fldPath) == 0 {
		return nil, utils.ErrNotFound
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

func (dDP *dynamicDP) fieldAsInterface(fldPath []string) (val interface{}, err error) {
	if len(fldPath) < 2 {
		return nil, fmt.Errorf("invalid fieldname <%s>", fldPath)
	}
	switch fldPath[0] {
	case utils.MetaAccounts:
		// sample of fieldName : ~*accounts.1001.BalanceMap.*monetary[0].Value
		// split the field name in 3 parts
		// fieldNameType (~*accounts), accountID(1001) and queried part (BalanceMap.*monetary[0].Value)

		var account Account
		if err = connMgr.Call(dDP.apiConns, nil, utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: dDP.tenant, Account: fldPath[1]}, &account); err != nil {
			return
		}
		//construct dataProvider from account and set it further
		dp := config.NewObjectDP(account)
		dDP.cache.Set(fldPath[:2], dp)
		return dp.FieldAsInterface(fldPath[2:])
	case utils.MetaResources:
		// sample of fieldName : ~*resources.ResourceID.Field
		var reply *Resource
		if err := connMgr.Call(dDP.resConns, nil, utils.ResourceSv1GetResource,
			&utils.TenantID{Tenant: dDP.tenant, ID: fldPath[1]}, &reply); err != nil {
			return nil, err
		}
		dp := config.NewObjectDP(reply)
		dDP.cache.Set(fldPath[:2], dp)
		return dp.FieldAsInterface(fldPath[2:])
	case utils.MetaStats:
		// sample of fieldName : ~*stats.StatID.*acd
		var statValues map[string]float64

		if err := connMgr.Call(dDP.stsConns, nil, utils.StatSv1GetQueueFloatMetrics,
			&utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: dDP.tenant, ID: fldPath[1]}},
			&statValues); err != nil {
			return nil, err
		}
		for k, v := range statValues {
			dDP.cache.Set([]string{utils.MetaStats, fldPath[1], k}, v)
		}
		return dDP.cache.FieldAsInterface(fldPath)
	case utils.MetaLibPhoneNumber:
		if len(fldPath) < 3 {
			return nil, fmt.Errorf("invalid fieldname <%s> for libphonenumber", fldPath)
		}
		// sample of fieldName ~*libphonenumber.*req.Destination
		// or ~*libphonenumber.*req.Destination.Carrier
		fieldFromDP, err := dDP.initialDP.FieldAsString(fldPath[1:3])
		if err != nil {
			utils.Logger.Warning(fmt.Sprintf("Received error: <%+v> when getting Destination for libphonenumber", err))
			return nil, err
		}
		num, err := phonenumbers.Parse(fieldFromDP, utils.EmptyString)
		if err != nil {
			return nil, err
		}
		// add the fields from libphonenumber
		dDP.cache.Set([]string{utils.MetaLibPhoneNumber, fieldFromDP, "CountryCode"}, num.CountryCode)
		dDP.cache.Set([]string{utils.MetaLibPhoneNumber, fieldFromDP, "NationalNumber"}, num.GetNationalNumber())
		dDP.cache.Set([]string{utils.MetaLibPhoneNumber, fieldFromDP, "Region"}, phonenumbers.GetRegionCodeForNumber(num))
		dDP.cache.Set([]string{utils.MetaLibPhoneNumber, fieldFromDP, "NumberType"}, phonenumbers.GetNumberType(num))
		geoLocation, err := phonenumbers.GetGeocodingForNumber(num, phonenumbers.GetRegionCodeForNumber(num))
		if err != nil {
			utils.Logger.Warning(fmt.Sprintf("Received error: <%+v> when getting GeoLocation for number %+v", err, num))
		}
		dDP.cache.Set([]string{utils.MetaLibPhoneNumber, fieldFromDP, "GeoLocation"}, geoLocation)
		carrier, err := phonenumbers.GetCarrierForNumber(num, phonenumbers.GetRegionCodeForNumber(num))
		if err != nil {
			utils.Logger.Warning(fmt.Sprintf("Received error: <%+v> when getting Carrier for number %+v", err, num))
		}
		dDP.cache.Set([]string{utils.MetaLibPhoneNumber, fieldFromDP, "Carrier"}, carrier)
		path := []string{utils.MetaLibPhoneNumber, fieldFromDP}
		if len(fldPath) == 4 {
			path = append(path, fldPath[3])
		}
		return dDP.cache.FieldAsInterface(path)
	default: // in case of constant we give an empty DataProvider ( empty navigable map )
	}
	return nil, utils.ErrNotFound
}
