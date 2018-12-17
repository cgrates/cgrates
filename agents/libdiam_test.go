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
	"reflect"
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
			diam.NewAVP(avp.ValueDigits, avp.Mbit, 0, datatype.Integer64(10000)),
		}})
	m.NewAVP("Subscription-Id", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(450, avp.Mbit, 0, datatype.Enumerated(1)),              // Subscription-Id-Type
			diam.NewAVP(444, avp.Mbit, 0, datatype.UTF8String("208708000003")), // Subscription-Id-Data
			diam.NewAVP(avp.ValueDigits, avp.Mbit, 0, datatype.Integer64(20000)),
		}})
	m.NewAVP("Subscription-Id", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(450, avp.Mbit, 0, datatype.Enumerated(2)),              // Subscription-Id-Type
			diam.NewAVP(444, avp.Mbit, 0, datatype.UTF8String("208708000004")), // Subscription-Id-Data
			diam.NewAVP(avp.ValueDigits, avp.Mbit, 0, datatype.Integer64(20000)),
		}})

	dP := newDADataProvider(nil, m)
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
	} else if out != eOut { // can be any result since both entries are matching single filter
		t.Errorf("expecting: %v, received: %v", eOut, out)
	}
	eOut = interface{}("208708000004")
	if out, err := dP.FieldAsInterface([]string{"Subscription-Id",
		"Subscription-Id-Data[~Subscription-Id-Type(2)|~Value-Digits(20000)]"}); err != nil { // on multiple filter
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

func TestMessageSetAVPsWithPath(t *testing.T) {
	eMessage := diam.NewRequest(diam.CreditControl, 4, nil)
	eMessage.NewAVP("Session-Id", avp.Mbit, 0,
		datatype.UTF8String("simuhuawei;1449573472;00001"))
	m := diam.NewMessage(diam.CreditControl, diam.RequestFlag, 4,
		eMessage.Header.HopByHopID, eMessage.Header.EndToEndID, nil)
	if err := messageSetAVPsWithPath(m,
		[]string{"Session-Id"}, "simuhuawei;1449573472;00001",
		false, "UTC"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eMessage, m) {
		t.Errorf("Expecting: %+v, received: %+v", eMessage, m)
	}
	// create same attribute twice
	eMessage.NewAVP("Session-Id", avp.Mbit, 0,
		datatype.UTF8String("simuhuawei;1449573472;00002"))
	if err := messageSetAVPsWithPath(m,
		[]string{"Session-Id"}, "simuhuawei;1449573472;00002",
		true, "UTC"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eMessage.AVP, m.AVP) {
		t.Errorf("Expecting: %+v, received: %+v", eMessage, m)
	}
	// overwrite of previous attribute
	eMessage = diam.NewRequest(diam.CreditControl, 4, nil)
	eMessage.NewAVP("Session-Id", avp.Mbit, 0,
		datatype.UTF8String("simuhuawei;1449573472;00001"))
	eMessage.NewAVP("Session-Id", avp.Mbit, 0,
		datatype.UTF8String("simuhuawei;1449573472;00003"))
	if err := messageSetAVPsWithPath(m,
		[]string{"Session-Id"}, "simuhuawei;1449573472;00003",
		false, "UTC"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eMessage.AVP, m.AVP) {
		t.Errorf("Expecting: %+v, received: %+v", eMessage, m)
	}
	// adding a groupped AVP
	eMessage.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(0)),
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("1001")),
		}})
	if err := messageSetAVPsWithPath(m,
		[]string{"Subscription-Id", "Subscription-Id-Type"}, "0",
		false, "UTC"); err != nil {
		t.Error(err)
	}
	if err := messageSetAVPsWithPath(m,
		[]string{"Subscription-Id", "Subscription-Id-Data"}, "1001",
		false, "UTC"); err != nil {
		t.Error(err)
	} else if len(eMessage.AVP) != len(m.AVP) {
		t.Errorf("Expecting: %+v, received: %+v", eMessage, m)
	}
	eMessage.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(1)),
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("1002")),
		}})
	if err := messageSetAVPsWithPath(m,
		[]string{"Subscription-Id", "Subscription-Id-Type"}, "1",
		true, "UTC"); err != nil {
		t.Error(err)
	}
	if err := messageSetAVPsWithPath(m,
		[]string{"Subscription-Id", "Subscription-Id-Data"}, "1002",
		false, "UTC"); err != nil {
		t.Error(err)
	} else if len(eMessage.AVP) != len(m.AVP) {
		t.Errorf("Expecting: %+v, received: %+v", eMessage, m)
	}
}

func TestMessageSetAVPsWithPath2(t *testing.T) {
	eMessage := diam.NewRequest(diam.CreditControl, 4, nil)
	eMessage.NewAVP("Multiple-Services-Credit-Control", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(432, avp.Mbit, 0, datatype.Unsigned32(65000)),
		}})
	eMessage.NewAVP("Multiple-Services-Credit-Control", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(432, avp.Mbit, 0, datatype.Unsigned32(100)),
		}})
	m := diam.NewMessage(diam.CreditControl, diam.RequestFlag, 4,
		eMessage.Header.HopByHopID, eMessage.Header.EndToEndID, nil)
	if err := messageSetAVPsWithPath(m,
		[]string{"Multiple-Services-Credit-Control", "Rating-Group"}, "65000",
		true, "UTC"); err != nil {
		t.Error(err)
	}
	if err := messageSetAVPsWithPath(m,
		[]string{"Multiple-Services-Credit-Control", "Rating-Group"}, "100",
		true, "UTC"); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(eMessage, m) {
		t.Errorf("Expecting: %+v, received: %+v", eMessage, m)
	}
}

func TestMessageSetAVPsWithPath3(t *testing.T) {
	eMessage := diam.NewRequest(diam.CreditControl, 4, nil)
	eMessage.NewAVP("Multiple-Services-Credit-Control", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(430, avp.Mbit, 0, &diam.GroupedAVP{ // 430 code for Final-Unit-Indication
				AVP: []*diam.AVP{
					diam.NewAVP(449, avp.Mbit, 0, datatype.Enumerated(1)), // 449 code for Final-Unit-Action
					diam.NewAVP(434, avp.Mbit, 0, &diam.GroupedAVP{ // 434 code for Redirect-Server
						AVP: []*diam.AVP{
							diam.NewAVP(433, avp.Mbit, 0, datatype.Enumerated(2)),                      // 433 code for Redirect-Address-Type
							diam.NewAVP(435, avp.Mbit, 0, datatype.UTF8String("http://172.10.88.88/")), //435 code for Redirect-Server-Address
						},
					},
					),
				},
			},
			),
		},
	},
	)
	m := diam.NewMessage(diam.CreditControl, diam.RequestFlag, 4,
		eMessage.Header.HopByHopID, eMessage.Header.EndToEndID, nil)

	if err := messageSetAVPsWithPath(m,
		[]string{"Multiple-Services-Credit-Control", "Final-Unit-Indication", "Final-Unit-Action"}, "1",
		false, "UTC"); err != nil {
		t.Error(err)
	}

	if err := messageSetAVPsWithPath(m,
		[]string{"Multiple-Services-Credit-Control", "Final-Unit-Indication", "Redirect-Server", "Redirect-Address-Type"}, "2",
		false, "UTC"); err != nil {
		t.Error(err)
	}

	if err := messageSetAVPsWithPath(m,
		[]string{"Multiple-Services-Credit-Control", "Final-Unit-Indication", "Redirect-Server", "Redirect-Server-Address"}, "http://172.10.88.88/",
		false, "UTC"); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(eMessage.String(), m.String()) {
		t.Errorf("Expecting: %+v \n, received: %+v \n", eMessage, m)
	}
}

func TestMessageSetAVPsWithPath4(t *testing.T) {
	eMessage := diam.NewRequest(diam.CreditControl, 4, nil)
	eMessage.NewAVP("Multiple-Services-Credit-Control", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(430, avp.Mbit, 0, &diam.GroupedAVP{ // 430 code for Final-Unit-Indication
				AVP: []*diam.AVP{
					diam.NewAVP(434, avp.Mbit, 0, &diam.GroupedAVP{ // 434 code for Redirect-Server
						AVP: []*diam.AVP{
							diam.NewAVP(433, avp.Mbit, 0, datatype.Enumerated(2)),                      // 433 code for Redirect-Address-Type
							diam.NewAVP(435, avp.Mbit, 0, datatype.UTF8String("http://172.10.88.88/")), //435 code for Redirect-Server-Address
						},
					},
					),
					diam.NewAVP(449, avp.Mbit, 0, datatype.Enumerated(1)), // 449 code for Final-Unit-Action
				},
			},
			),
		},
	},
	)
	m := diam.NewMessage(diam.CreditControl, diam.RequestFlag, 4,
		eMessage.Header.HopByHopID, eMessage.Header.EndToEndID, nil)

	if err := messageSetAVPsWithPath(m,
		[]string{"Multiple-Services-Credit-Control", "Final-Unit-Indication", "Redirect-Server", "Redirect-Address-Type"}, "2",
		false, "UTC"); err != nil {
		t.Error(err)
	}
	if err := messageSetAVPsWithPath(m,
		[]string{"Multiple-Services-Credit-Control", "Final-Unit-Indication", "Final-Unit-Action"}, "1",
		false, "UTC"); err != nil {
		t.Error(err)
	}

	if err := messageSetAVPsWithPath(m,
		[]string{"Multiple-Services-Credit-Control", "Final-Unit-Indication", "Redirect-Server", "Redirect-Server-Address"}, "http://172.10.88.88/",
		false, "UTC"); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(eMessage.String(), m.String()) {
		t.Errorf("Expecting: %+v \n, received: %+v \n", eMessage, m)
	}
}

func TestMessageSetAVPsWithPath5(t *testing.T) {
	eMessage := diam.NewRequest(diam.CreditControl, 4, nil)
	eMessage.NewAVP("Multiple-Services-Credit-Control", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(430, avp.Mbit, 0, &diam.GroupedAVP{ // 430 code for Final-Unit-Indication
				AVP: []*diam.AVP{
					diam.NewAVP(434, avp.Mbit, 0, &diam.GroupedAVP{ // 434 code for Redirect-Server
						AVP: []*diam.AVP{
							diam.NewAVP(433, avp.Mbit, 0, datatype.Enumerated(2)),                      // 433 code for Redirect-Address-Type
							diam.NewAVP(435, avp.Mbit, 0, datatype.UTF8String("http://172.10.88.88/")), //435 code for Redirect-Server-Address
						},
					},
					),
					diam.NewAVP(449, avp.Mbit, 0, datatype.Enumerated(1)), // 449 code for Final-Unit-Action
				},
			},
			),
		},
	},
	)
	m := diam.NewMessage(diam.CreditControl, diam.RequestFlag, 4,
		eMessage.Header.HopByHopID, eMessage.Header.EndToEndID, nil)

	if err := messageSetAVPsWithPath(m,
		[]string{"Multiple-Services-Credit-Control", "Final-Unit-Indication", "Redirect-Server", "Redirect-Address-Type"}, "2",
		false, "UTC"); err != nil {
		t.Error(err)
	}

	if err := messageSetAVPsWithPath(m,
		[]string{"Multiple-Services-Credit-Control", "Final-Unit-Indication", "Redirect-Server", "Redirect-Server-Address"}, "http://172.10.88.88/",
		false, "UTC"); err != nil {
		t.Error(err)
	}
	if err := messageSetAVPsWithPath(m,
		[]string{"Multiple-Services-Credit-Control", "Final-Unit-Indication", "Final-Unit-Action"}, "1",
		false, "UTC"); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(eMessage.String(), m.String()) {
		t.Errorf("Expecting: %+v \n, received: %+v \n", eMessage, m)
	}
}

func TestdisectDiamListen(t *testing.T) {
	expIPs := []string{"192.168.56.203", "192.168.57.203"}
	rvc := disectDiamListen("192.168.56.203/192.168.57.203:3869")
	if !reflect.DeepEqual(expIPs, rvc) {
		t.Errorf("Expecting: %+v \n, received: %+v \n ", expIPs, rvc)
	}
	expIPs = []string{"192.168.56.203"}
	rvc = disectDiamListen("192.168.56.203:3869")
	if !reflect.DeepEqual(expIPs, rvc) {
		t.Errorf("Expecting: %+v \n, received: %+v \n ", expIPs, rvc)
	}
	expIPs = []string{}
	rvc = disectDiamListen(":3869")
	if !reflect.DeepEqual(expIPs, rvc) {
		t.Errorf("Expecting: %+v \n, received: %+v \n ", expIPs, rvc)
	}

}
