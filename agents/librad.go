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

package agents

import (
	"net"

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/radigo"
)

// radReplyAppendAttributes appends attributes to a RADIUS reply based on predefined template
func radReplyAppendAttributes(reply *radigo.Packet, rplNM *utils.OrderedNavigableMap) (err error) {
	for el := rplNM.GetFirstElement(); el != nil; el = el.Next() {
		val := el.Value
		var cfgItm *utils.DataLeaf
		if cfgItm, err = rplNM.Field(val); err != nil {
			return
		}

		if cfgItm.Path[0] == MetaRadReplyCode { // Special case used to control the reply code of RADIUS reply
			if err = reply.SetCodeWithName(utils.IfaceAsString(cfgItm.Data)); err != nil {
				return err
			}
			continue
		}
		var attrName, vendorName string
		if len(cfgItm.Path) > 2 {
			vendorName, attrName = cfgItm.Path[0], cfgItm.Path[1]
		} else {
			attrName = cfgItm.Path[0]
		}

		if err = reply.AddAVPWithName(attrName, utils.IfaceAsString(cfgItm.Data), vendorName); err != nil {
			return err
		}
	}
	return
}

// newRADataProvider constructs a DataProvider
func newRADataProvider(req *radigo.Packet) (dP utils.DataProvider) {
	dP = &radiusDP{req: req, cache: utils.MapStorage{}}
	return
}

// radiusDP implements utils.DataProvider, serving as radigo.Packet data decoder
// decoded data is only searched once and cached
type radiusDP struct {
	req   *radigo.Packet
	cache utils.MapStorage
}

// String is part of utils.DataProvider interface
// when called, it will display the already parsed values out of cache
func (pk *radiusDP) String() string {
	return utils.ToIJSON(pk.req) // return ToJSON because Packet don't have a string method
}

// FieldAsInterface is part of utils.DataProvider interface
func (pk *radiusDP) FieldAsInterface(fldPath []string) (data interface{}, err error) {
	if len(fldPath) != 1 {
		return nil, utils.ErrNotFound
	}
	if data, err = pk.cache.FieldAsInterface(fldPath); err != nil {
		if err != utils.ErrNotFound { // item found in cache
			return
		}
		err = nil // cancel previous err
	} else {
		return // data found in cache
	}
	if len(pk.req.AttributesWithName(fldPath[0], "")) != 0 {
		data = pk.req.AttributesWithName(fldPath[0], "")[0].GetStringValue()
	}
	pk.cache.Set(fldPath, data)
	return
}

// FieldAsString is part of utils.DataProvider interface
func (pk *radiusDP) FieldAsString(fldPath []string) (data string, err error) {
	var valIface interface{}
	valIface, err = pk.FieldAsInterface(fldPath)
	if err != nil {
		return
	}
	return utils.IfaceAsString(valIface), nil
}

// RemoteHost is part of utils.DataProvider interface
func (pk *radiusDP) RemoteHost() net.Addr {
	return utils.NewNetAddr(pk.req.RemoteAddr().Network(), pk.req.RemoteAddr().String())
}

//radauthReq is used to authorize a request based on flags
func radauthReq(flags utils.FlagsWithParams, req *radigo.Packet, aReq *AgentRequest, rpl *radigo.Packet) (bool, error) {
	// try to get UserPassword from Vars as slice of NMItems
	nmItems, has := aReq.Vars.Map[utils.UserPassword]
	if !has {
		return false, utils.ErrNotFound
	}
	pass := nmItems.Slice[0].Value.String()
	switch {
	case flags.Has(utils.MetaPAP):
		userPassAvps := req.AttributesWithName(UserPasswordAVP, utils.EmptyString)
		if len(userPassAvps) == 0 {
			return false, utils.NewErrMandatoryIeMissing(UserPasswordAVP)
		}
		if userPassAvps[0].StringValue != pass {
			return false, nil
		}
	case flags.Has(utils.MetaCHAP):
		chapAVPs := req.AttributesWithName(CHAPPasswordAVP, utils.EmptyString)
		if len(chapAVPs) == 0 {
			return false, utils.NewErrMandatoryIeMissing(CHAPPasswordAVP)
		}
		return radigo.AuthenticateCHAP([]byte(pass),
			req.Authenticator[:], chapAVPs[0].RawValue), nil
	case flags.Has(utils.MetaMSCHAPV2):
		msChallenge := req.AttributesWithName(MSCHAPChallengeAVP, MicrosoftVendor)
		if len(msChallenge) == 0 {
			return false, utils.NewErrMandatoryIeMissing(MSCHAPChallengeAVP)
		}
		msResponse := req.AttributesWithName(MSCHAPResponseAVP, MicrosoftVendor)
		if len(msResponse) == 0 {
			return false, utils.NewErrMandatoryIeMissing(MSCHAPResponseAVP)
		}
		vsaMSResponde := msResponse[0].Value.(*radigo.VSA)
		vsaMSChallange := msChallenge[0].Value.(*radigo.VSA)

		userName := req.AttributesWithName("User-Name", utils.EmptyString)[0].StringValue
		passwordFromAttributes := pass

		if len(vsaMSChallange.RawValue) != 16 || len(vsaMSResponde.RawValue) != 50 {
			return false, nil
		}
		ident := vsaMSResponde.RawValue[0]
		peerChallenge := vsaMSResponde.RawValue[2:18]
		peerResponse := vsaMSResponde.RawValue[26:50]
		ntResponse, err := radigo.GenerateNTResponse(vsaMSChallange.RawValue,
			peerChallenge, userName, passwordFromAttributes)
		if err != nil {
			return false, err
		}
		if len(ntResponse) != len(peerResponse) {
			return false, nil
		}
		for i := range ntResponse {
			if ntResponse[i] != peerResponse[i] {
				return false, nil
			}
		}

		authenticatorResponse, err := radigo.GenerateAuthenticatorResponse(vsaMSChallange.RawValue, peerChallenge,
			ntResponse, userName, passwordFromAttributes)
		if err != nil {
			return false, err
		}
		success := make([]byte, 43)
		success[0] = ident
		copy(success[1:], authenticatorResponse)
		// this AVP need to be added to be verified on the client side
		rpl.AddAVPWithName(MSCHAP2SuccessAVP, string(success), MicrosoftVendor)
	default:
		return false, utils.NewErrMandatoryIeMissing(utils.Flags)
	}

	return true, nil
}
