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
	eOut = interface{}(uint32(33))
	if out, err := dP.FieldAsInterface([]string{"Requested-Service-Unit", "CC-Money", "Currency-Code"}); err != nil {
		t.Error(err)
	} else if eOut != out {
		t.Errorf("Expecting: %v, received: %v", eOut, out)
	}
}
