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
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/go-diameter/diam"
	"github.com/cgrates/go-diameter/diam/avp"
	"github.com/cgrates/go-diameter/diam/datatype"
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
			diam.NewAVP(452, avp.Mbit, 0, datatype.Enumerated(1)),
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
			diam.NewAVP(439, avp.Mbit, 0, datatype.Unsigned32(1)),
		},
	},
	)
	m := diam.NewMessage(diam.CreditControl, diam.RequestFlag, 4,
		eMessage.Header.HopByHopID, eMessage.Header.EndToEndID, nil)

	if err := messageSetAVPsWithPath(m,
		[]string{"Multiple-Services-Credit-Control", "Tariff-Change-Usage"}, "1",
		false, "UTC"); err != nil {
		t.Error(err)
	}

	if err := messageSetAVPsWithPath(m,
		[]string{"Multiple-Services-Credit-Control", "Final-Unit-Indication", "Redirect-Server", "Redirect-Address-Type"}, "2",
		false, "UTC"); err != nil {
		t.Error(err)
	}

	if err := messageSetAVPsWithPath(m,
		[]string{"Multiple-Services-Credit-Control", "Service-Identifier"}, "1",
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
		t.Errorf("Expected %s, recived %s", utils.ToJSON(eMessage), utils.ToJSON(m))
		// t.Errorf("Expecting: %+v \n, received: %+v \n", eMessage, m)
	}
}

// In case we send -1 as CC-Time we get error from go-diameter :
// " strconv.ParseUint: parsing "-1": invalid syntax "
// and crashed
/*
func TestMessageSetAVPsWithPath6(t *testing.T) {
	m := diam.NewMessage(diam.CreditControl, diam.RequestFlag, 4,
		eMessage.Header.HopByHopID, eMessage.Header.EndToEndID, nil)

	if err := messageSetAVPsWithPath(m,
		[]string{"Granted-Service-Unit", "CC-Time"}, "-1",
		false, "UTC"); err != nil {
		t.Error(err)
	}
}
*/

func TestDisectDiamListen(t *testing.T) {
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
	if len(rvc) != 0 {
		t.Errorf("Expecting: %+v \n, received: %+q \n ", expIPs, rvc)
	}

}

func TestUpdateDiamMsgFromNavMap1(t *testing.T) {
	eMessage := diam.NewRequest(diam.CreditControl, 4, nil)
	eMessage.NewAVP("Multiple-Services-Credit-Control", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(430, avp.Mbit, 0, &diam.GroupedAVP{ // 430 code for Final-Unit-Indication
				AVP: []*diam.AVP{
					diam.NewAVP(434, avp.Mbit, 0, &diam.GroupedAVP{ // 434 code for Redirect-Server
						AVP: []*diam.AVP{
							diam.NewAVP(433, avp.Mbit, 0, datatype.Enumerated(2)),                      // 433 code for Redirect-Address-Type
							diam.NewAVP(435, avp.Mbit, 0, datatype.UTF8String("http://172.10.88.88/")), // 435 code for Redirect-Server-Address
						},
					},
					),
				},
			},
			),
			diam.NewAVP(452, avp.Mbit, 0, datatype.Enumerated(2)), // 452 code for Redirect-Address-Type
		},
	},
	)

	m2 := diam.NewMessage(diam.CreditControl, diam.RequestFlag, 4,
		eMessage.Header.HopByHopID, eMessage.Header.EndToEndID, nil)
	m2.NewAVP("Multiple-Services-Credit-Control", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(430, avp.Mbit, 0, &diam.GroupedAVP{ // 430 code for Final-Unit-Indication
			},
			),
		},
	},
	)

	nM := config.NewNavigableMap(nil)
	itm := &config.NMItem{
		Path: []string{"Multiple-Services-Credit-Control", "Final-Unit-Indication", "Redirect-Server", "Redirect-Address-Type"},
		Data: datatype.Enumerated(2),
	}
	nM.Set(itm.Path, []*config.NMItem{itm}, true, true)

	itm = &config.NMItem{
		Path: []string{"Multiple-Services-Credit-Control", "Tariff-Change-Usage"},
		Data: datatype.Enumerated(2),
	}
	nM.Set(itm.Path, []*config.NMItem{itm}, true, true)

	itm = &config.NMItem{
		Path: []string{"Multiple-Services-Credit-Control", "Final-Unit-Indication", "Redirect-Server", "Redirect-Server-Address"},
		Data: "http://172.10.88.88/",
	}
	nM.Set(itm.Path, []*config.NMItem{itm}, true, true)

	if err := updateDiamMsgFromNavMap(m2, nM, ""); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(eMessage.String(), m2.String()) {
		t.Errorf("Expected %s, recived %s", utils.ToJSON(eMessage), utils.ToJSON(m2))
	}
}

func TestUpdateDiamMsgFromNavMap2(t *testing.T) {
	eMessage := diam.NewRequest(diam.CreditControl, 4, nil)
	eMessage.NewAVP("Multiple-Services-Credit-Control", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(430, avp.Mbit, 0, &diam.GroupedAVP{ // 430 code for Final-Unit-Indication
				AVP: []*diam.AVP{
					diam.NewAVP(434, avp.Mbit, 0, &diam.GroupedAVP{ // 434 code for Redirect-Server
						AVP: []*diam.AVP{
							diam.NewAVP(433, avp.Mbit, 0, datatype.Enumerated(2)), // 433 code for Redirect-Address-Type
						},
					},
					),
				},
			},
			),
			diam.NewAVP(452, avp.Mbit, 0, datatype.Enumerated(2)), // 452 code for Redirect-Address-Type
		},
	},
	)
	eMessage.NewAVP("Multiple-Services-Credit-Control", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(430, avp.Mbit, 0, &diam.GroupedAVP{ // 430 code for Final-Unit-Indication
				AVP: []*diam.AVP{
					diam.NewAVP(434, avp.Mbit, 0, &diam.GroupedAVP{ // 434 code for Redirect-Server
						AVP: []*diam.AVP{
							diam.NewAVP(435, avp.Mbit, 0, datatype.UTF8String("http://172.10.88.88/")), // 435 code for Redirect-Server-Address
							diam.NewAVP(433, avp.Mbit, 0, datatype.Enumerated(2)),                      // 433 code for Redirect-Address-Type
						},
					},
					),
				},
			},
			),
		},
	},
	)

	m2 := diam.NewMessage(diam.CreditControl, diam.RequestFlag, 4,
		eMessage.Header.HopByHopID, eMessage.Header.EndToEndID, nil)
	m2.NewAVP("Multiple-Services-Credit-Control", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(430, avp.Mbit, 0, &diam.GroupedAVP{ // 430 code for Final-Unit-Indication
			},
			),
		},
	},
	)

	nM := config.NewNavigableMap(nil)
	itm := &config.NMItem{
		Path: []string{"Multiple-Services-Credit-Control", "Final-Unit-Indication", "Redirect-Server", "Redirect-Address-Type"},
		Data: datatype.Enumerated(2),
	}
	nM.Set(itm.Path, []*config.NMItem{itm}, true, true)

	itm = &config.NMItem{
		Path: []string{"Multiple-Services-Credit-Control", "Tariff-Change-Usage"},
		Data: datatype.Enumerated(2),
	}
	nM.Set(itm.Path, []*config.NMItem{itm}, true, true)

	itm = &config.NMItem{
		Path:   []string{"Multiple-Services-Credit-Control", "Final-Unit-Indication", "Redirect-Server", "Redirect-Server-Address"},
		Data:   "http://172.10.88.88/",
		Config: &config.FCTemplate{NewBranch: true},
	}
	nM.Set(itm.Path, []*config.NMItem{itm}, true, true)

	itm = &config.NMItem{
		Path: []string{"Multiple-Services-Credit-Control", "Final-Unit-Indication", "Redirect-Server", "Redirect-Address-Type"},
		Data: datatype.Enumerated(2),
	}
	nM.Set(itm.Path, []*config.NMItem{itm}, true, true)

	if err := updateDiamMsgFromNavMap(m2, nM, ""); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(eMessage.String(), m2.String()) {
		t.Errorf("Expected %s, recived %s", utils.ToJSON(eMessage), utils.ToJSON(m2))
	}
}

func TestUpdateDiamMsgFromNavMap3(t *testing.T) {
	eMessage := diam.NewRequest(diam.CreditControl, 4, nil)
	eMessage.NewAVP("Multiple-Services-Credit-Control", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(430, avp.Mbit, 0, &diam.GroupedAVP{ // 430 code for Final-Unit-Indication
				AVP: []*diam.AVP{
					diam.NewAVP(449, avp.Mbit, 0, datatype.Enumerated(1)), // 449 code for Final-Unit-Action
					diam.NewAVP(434, avp.Mbit, 0, &diam.GroupedAVP{ // 434 code for Redirect-Server
						AVP: []*diam.AVP{
							diam.NewAVP(433, avp.Mbit, 0, datatype.Enumerated(2)),                      // 433 code for Redirect-Address-Type
							diam.NewAVP(435, avp.Mbit, 0, datatype.UTF8String("http://172.10.88.88/")), // 435 code for Redirect-Server-Address
						},
					},
					),
				},
			},
			),
			diam.NewAVP(452, avp.Mbit, 0, datatype.Enumerated(2)), // 452 code for Redirect-Address-Type
		},
	},
	)

	m2 := diam.NewMessage(diam.CreditControl, diam.RequestFlag, 4,
		eMessage.Header.HopByHopID, eMessage.Header.EndToEndID, nil)

	nM := config.NewNavigableMap(nil)

	itm := &config.NMItem{
		Path: []string{"Multiple-Services-Credit-Control", "Final-Unit-Indication", "Final-Unit-Action"},
		Data: datatype.Enumerated(1),
	}
	nM.Set(itm.Path, []*config.NMItem{itm}, true, true)

	itm = &config.NMItem{
		Path: []string{"Multiple-Services-Credit-Control", "Final-Unit-Indication", "Redirect-Server", "Redirect-Address-Type"},
		Data: datatype.Enumerated(2),
	}
	nM.Set(itm.Path, []*config.NMItem{itm}, true, true)

	itm = &config.NMItem{
		Path: []string{"Multiple-Services-Credit-Control", "Tariff-Change-Usage"},
		Data: datatype.Enumerated(2),
	}
	nM.Set(itm.Path, []*config.NMItem{itm}, true, true)

	itm = &config.NMItem{
		Path: []string{"Multiple-Services-Credit-Control", "Final-Unit-Indication", "Redirect-Server", "Redirect-Server-Address"},
		Data: "http://172.10.88.88/",
	}
	nM.Set(itm.Path, []*config.NMItem{itm}, true, true)

	if err := updateDiamMsgFromNavMap(m2, nM, ""); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(eMessage.String(), m2.String()) {
		t.Errorf("Expected %s, recived %s", utils.ToJSON(eMessage), utils.ToJSON(m2))
	}
}

func TestUpdateDiamMsgFromNavMap4(t *testing.T) {
	eMessage := diam.NewRequest(diam.CreditControl, 4, nil)
	eMessage.NewAVP("Multiple-Services-Credit-Control", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(430, avp.Mbit, 0, &diam.GroupedAVP{ // 430 code for Final-Unit-Indication
				AVP: []*diam.AVP{
					diam.NewAVP(449, avp.Mbit, 0, datatype.Enumerated(1)), // 449 code for Final-Unit-Action
					diam.NewAVP(434, avp.Mbit, 0, &diam.GroupedAVP{ // 434 code for Redirect-Server
						AVP: []*diam.AVP{
							diam.NewAVP(433, avp.Mbit, 0, datatype.Enumerated(2)),                      // 433 code for Redirect-Address-Type
							diam.NewAVP(435, avp.Mbit, 0, datatype.UTF8String("http://172.10.88.88/")), // 435 code for Redirect-Server-Address
						},
					},
					),
				},
			},
			),
			diam.NewAVP(452, avp.Mbit, 0, datatype.Enumerated(2)), // 452 code for Redirect-Address-Type
		},
	},
	)
	eMessage.NewAVP("Multiple-Services-Credit-Control", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(430, avp.Mbit, 0, &diam.GroupedAVP{ // 430 code for Final-Unit-Indication
				AVP: []*diam.AVP{
					diam.NewAVP(434, avp.Mbit, 0, &diam.GroupedAVP{ // 434 code for Redirect-Server
						AVP: []*diam.AVP{
							diam.NewAVP(435, avp.Mbit, 0, datatype.UTF8String("http://172.10.88.88/")), // 435 code for Redirect-Server-Address
						},
					},
					),
				},
			},
			),
		},
	},
	)
	eMessage.NewAVP("Granted-Service-Unit", avp.Mbit, 0, &diam.GroupedAVP{ // 431 code for Granted-Service-Unit
		AVP: []*diam.AVP{
			diam.NewAVP(420, avp.Mbit, 0, datatype.Unsigned32(10)), // 420 code for CC-Time
		},
	},
	)

	m2 := diam.NewMessage(diam.CreditControl, diam.RequestFlag, 4,
		eMessage.Header.HopByHopID, eMessage.Header.EndToEndID, nil)

	nM := config.NewNavigableMap(nil)

	itm := &config.NMItem{
		Path: []string{"Multiple-Services-Credit-Control", "Final-Unit-Indication", "Final-Unit-Action"},
		Data: datatype.Enumerated(1),
	}
	nM.Set(itm.Path, []*config.NMItem{itm}, true, true)

	itm = &config.NMItem{
		Path: []string{"Multiple-Services-Credit-Control", "Final-Unit-Indication", "Redirect-Server", "Redirect-Address-Type"},
		Data: datatype.Enumerated(2),
	}
	nM.Set(itm.Path, []*config.NMItem{itm}, true, true)

	itm = &config.NMItem{
		Path: []string{"Multiple-Services-Credit-Control", "Tariff-Change-Usage"},
		Data: datatype.Enumerated(2),
	}
	nM.Set(itm.Path, []*config.NMItem{itm}, true, true)

	itm = &config.NMItem{
		Path: []string{"Multiple-Services-Credit-Control", "Final-Unit-Indication", "Redirect-Server", "Redirect-Server-Address"},
		Data: "http://172.10.88.88/",
	}
	nM.Set(itm.Path, []*config.NMItem{itm}, true, true)

	itm = &config.NMItem{
		Path:   []string{"Multiple-Services-Credit-Control", "Final-Unit-Indication", "Redirect-Server", "Redirect-Server-Address"},
		Data:   "http://172.10.88.88/",
		Config: &config.FCTemplate{NewBranch: true},
	}
	nM.Set(itm.Path, []*config.NMItem{itm}, true, true)

	itm = &config.NMItem{
		Path: []string{"Granted-Service-Unit", "CC-Time"},
		Data: datatype.Unsigned32(10),
	}
	nM.Set(itm.Path, []*config.NMItem{itm}, true, true)

	if err := updateDiamMsgFromNavMap(m2, nM, ""); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(eMessage.String(), m2.String()) {
		t.Errorf("Expected %s, recived %s", utils.ToJSON(eMessage), utils.ToJSON(m2))
	}
}

func TestDiamAVPAsIface(t *testing.T) {
	args := diam.NewAVP(435, avp.Mbit, 0, datatype.Address("127.0.0.1"))
	var exp interface{} = net.IP([]byte("127.0.0.1"))
	if rply, err := diamAVPAsIface(args); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected<%T>: %v ,received<%T>: %v ", exp, exp, rply, rply)
	}

	args = diam.NewAVP(435, avp.Mbit, 0, datatype.DiameterIdentity("diam1"))
	exp = "diam1"
	if rply, err := diamAVPAsIface(args); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected<%T>: %v ,received<%T>: %v ", exp, exp, rply, rply)
	}

	args = diam.NewAVP(435, avp.Mbit, 0, datatype.DiameterURI("http://172.10.88.88/"))
	exp = "http://172.10.88.88/"
	if rply, err := diamAVPAsIface(args); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected<%T>: %v ,received<%T>: %v ", exp, exp, rply, rply)
	}

	args = diam.NewAVP(450, avp.Mbit, 0, datatype.Enumerated(10))
	exp = int32(10)
	if rply, err := diamAVPAsIface(args); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected<%T>: %v ,received<%T>: %v ", exp, exp, rply, rply)
	}

	args = diam.NewAVP(450, avp.Mbit, 0, datatype.Float32(10.25))
	exp = float32(10.25)
	if rply, err := diamAVPAsIface(args); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected<%T>: %v ,received<%T>: %v ", exp, exp, rply, rply)
	}

	args = diam.NewAVP(450, avp.Mbit, 0, datatype.Float64(10.25))
	exp = float64(10.25)
	if rply, err := diamAVPAsIface(args); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected<%T>: %v ,received<%T>: %v ", exp, exp, rply, rply)
	}

	args = diam.NewAVP(450, avp.Mbit, 0, datatype.IPFilterRule("fltr1"))
	exp = "fltr1"
	if rply, err := diamAVPAsIface(args); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected<%T>: %v ,received<%T>: %v ", exp, exp, rply, rply)
	}

	args = diam.NewAVP(435, avp.Mbit, 0, datatype.IPv4("127.0.0.1"))
	exp = net.IP([]byte("127.0.0.1"))
	if rply, err := diamAVPAsIface(args); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected<%T>: %v ,received<%T>: %v ", exp, exp, rply, rply)
	}

	args = diam.NewAVP(450, avp.Mbit, 0, datatype.Integer32(10))
	exp = int32(10)
	if rply, err := diamAVPAsIface(args); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected<%T>: %v ,received<%T>: %v ", exp, exp, rply, rply)
	}

	args = diam.NewAVP(450, avp.Mbit, 0, datatype.Integer64(10))
	exp = int64(10)
	if rply, err := diamAVPAsIface(args); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected<%T>: %v ,received<%T>: %v ", exp, exp, rply, rply)
	}

	args = diam.NewAVP(450, avp.Mbit, 0, datatype.OctetString("diam1"))
	exp = "diam1"
	if rply, err := diamAVPAsIface(args); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected<%T>: %v ,received<%T>: %v ", exp, exp, rply, rply)
	}

	args = diam.NewAVP(450, avp.Mbit, 0, datatype.QoSFilterRule("diam1"))
	exp = "diam1"
	if rply, err := diamAVPAsIface(args); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected<%T>: %v ,received<%T>: %v ", exp, exp, rply, rply)
	}

	args = diam.NewAVP(450, avp.Mbit, 0, datatype.UTF8String("diam1"))
	exp = "diam1"
	if rply, err := diamAVPAsIface(args); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected<%T>: %v ,received<%T>: %v ", exp, exp, rply, rply)
	}

	args = diam.NewAVP(450, avp.Mbit, 0, datatype.Unsigned32(10))
	exp = uint32(10)
	if rply, err := diamAVPAsIface(args); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected<%T>: %v ,received<%T>: %v ", exp, exp, rply, rply)
	}

	args = diam.NewAVP(450, avp.Mbit, 0, datatype.Unsigned64(10))
	exp = uint64(10)
	if rply, err := diamAVPAsIface(args); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected<%T>: %v ,received<%T>: %v ", exp, exp, rply, rply)
	}

	now := time.Now()
	args = diam.NewAVP(450, avp.Mbit, 0, datatype.Time(now))
	exp = now
	if rply, err := diamAVPAsIface(args); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected<%T>: %v ,received<%T>: %v ", exp, exp, rply, rply)
	}

	args = diam.NewAVP(434, avp.Mbit, 0, &diam.GroupedAVP{ // 434 code for Redirect-Server
		AVP: []*diam.AVP{
			diam.NewAVP(435, avp.Mbit, 0, datatype.UTF8String("http://172.10.88.88/")), // 435 code for Redirect-Server-Address
		},
	})
	if rply, err := diamAVPAsIface(args); err == nil {
		t.Errorf("Expected err received: err: %v, rply %v", err, rply)
	}
}
