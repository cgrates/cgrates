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
	"testing"

	"github.com/fiorix/go-diameter/diam"
	"github.com/fiorix/go-diameter/diam/avp"
	"github.com/fiorix/go-diameter/diam/datatype"
)

func TestDPFieldAsInterface(t *testing.T) {
	m := diam.NewRequest(diam.CreditControl, 4, nil)
	m.NewAVP("Session-Id", avp.Mbit, 0, datatype.UTF8String("simuhuawei;1449573472;00002"))
	m.NewAVP(avp.RequestedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.CCMoney, avp.Mbit, 0, &diam.GroupedAVP{
				AVP: []*diam.AVP{
					diam.NewAVP(avp.UnitValue, avp.Mbit, 0, &diam.GroupedAVP{
						AVP: []*diam.AVP{
							diam.NewAVP(avp.ValueDigits, avp.Mbit, 0, datatype.Integer64(10000)),
							diam.NewAVP(avp.Exponent, avp.Mbit, 0, datatype.Integer32(-5)),
						},
					}),
					diam.NewAVP(avp.CurrencyCode, avp.Mbit, 0, datatype.Unsigned32(33)),
				},
			}),
		},
	})
	m.NewAVP("Subscription-Id", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(450, avp.Mbit, 0, datatype.Enumerated(0)),             // Subscription-Id-Type
			diam.NewAVP(444, avp.Mbit, 0, datatype.UTF8String("33708000003")), // Subscription-Id-Data
		}})
	m.NewAVP("Subscription-Id", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(450, avp.Mbit, 0, datatype.Enumerated(1)),              // Subscription-Id-Type
			diam.NewAVP(444, avp.Mbit, 0, datatype.UTF8String("208708000003")), // Subscription-Id-Data
		}})

	dP := newDADataProvider(m)
	eOut := interface{}("simuhuawei;1449573472;00002")
	if out, err := dP.FieldAsInterface([]string{"Session-Id"}); err != nil {
		t.Error(err)
	} else if eOut != out {
		t.Errorf("Expecting: %v, received: %v", eOut, out)
	}
	eOut = interface{}(int64(10000))
	if out, err := dP.FieldAsInterface([]string{"Requested-Service-Unit", "CC-Money", "Unit-Value", "Value-Digits"}); err != nil {
		t.Error(err)
	} else if eOut != out {
		t.Errorf("Expecting: %v, received: %v", eOut, out)
	}
	eOut = interface{}("208708000003") // with filter on second group item
	if out, err := dP.FieldAsInterface([]string{"Subscription-Id",
		"Subscription-Id-Data[1]"}); err != nil { // on index
		t.Error(err)
	} else if eOut != out {
		t.Errorf("Expecting: %v, received: %v", eOut, out)
	}
	if out, err := dP.FieldAsInterface([]string{"Subscription-Id",
		"Subscription-Id-Data[~Subscription-Id-Type(1)]"}); err != nil { // on filter
		t.Error(err)
	} else if eOut != out {
		t.Errorf("Expecting: %v, received: %v", eOut, out)
	}
	eOut = interface{}("33708000003")
	if out, err := dP.FieldAsInterface([]string{"Subscription-Id", "Subscription-Id-Data"}); err != nil {
		t.Error(err)
	} else if eOut != out {
		t.Errorf("Expecting: %v, received: %v", eOut, out)
	}
}
