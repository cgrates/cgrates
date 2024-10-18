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
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/fiorix/go-diameter/v4/diam"
	"github.com/fiorix/go-diameter/v4/diam/avp"
	"github.com/fiorix/go-diameter/v4/diam/datatype"
)

func TestLibDiamDPFieldAsInterface(t *testing.T) {
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
	eOut := any("simuhuawei;1449573472;00002")
	if out, err := dP.FieldAsInterface([]string{"Session-Id"}); err != nil {
		t.Error(err)
	} else if eOut != out {
		t.Errorf("Expecting: %v, received: %v", eOut, out)
	}
	eOut = any(int64(10000))
	if out, err := dP.FieldAsInterface([]string{"Requested-Service-Unit", "CC-Money", "Unit-Value", "Value-Digits"}); err != nil {
		t.Error(err)
	} else if eOut != out {
		t.Errorf("Expecting: %v, received: %v", eOut, out)
	}
	eOut = any("208708000003") // with filter on second group item
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
	eOut = any("208708000004")
	if out, err := dP.FieldAsInterface([]string{"Subscription-Id",
		"Subscription-Id-Data[~Subscription-Id-Type(2)|~Value-Digits(20000)]"}); err != nil { // on multiple filter
		t.Error(err)
	} else if eOut != out {
		t.Errorf("Expecting: %v, received: %v", eOut, out)
	}
	eOut = any("33708000003")
	if out, err := dP.FieldAsInterface([]string{"Subscription-Id", "Subscription-Id-Data"}); err != nil {
		t.Error(err)
	} else if eOut != out {
		t.Errorf("Expecting: %v, received: %v", eOut, out)
	}
}

func TestLibDiamMessageSetAVPsWithPath(t *testing.T) {
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

func TestLibDiamMessageSetAVPsWithPath2(t *testing.T) {
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

func TestLibDiamMessageSetAVPsWithPath3(t *testing.T) {
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

func TestLibDiamMessageSetAVPsWithPath4(t *testing.T) {
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

func TestLibDiamMessageSetAVPsWithPath5(t *testing.T) {
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
		t.Errorf("Expected %s, received %s", utils.ToJSON(eMessage), utils.ToJSON(m))
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

func TestLibDiamDisectDiamListen(t *testing.T) {
	expIPs := []net.IP{net.ParseIP("192.168.56.203"), net.ParseIP("192.168.57.203")}
	rvc := disectDiamListen("192.168.56.203/192.168.57.203:3869")
	if !reflect.DeepEqual(expIPs, rvc) {
		t.Errorf("Expecting: %+v \n, received: %+v \n ", expIPs, rvc)
	}
	expIPs = []net.IP{net.ParseIP("192.168.56.203")}
	rvc = disectDiamListen("192.168.56.203:3869")
	if !reflect.DeepEqual(expIPs, rvc) {
		t.Errorf("Expecting: %+v \n, received: %+v \n ", expIPs, rvc)
	}
	expIPs = []net.IP{}
	rvc = disectDiamListen(":3869")
	if len(rvc) != 0 {
		t.Errorf("Expecting: %+v \n, received: %+q \n ", expIPs, rvc)
	}

}

func TestLibDiamUpdateDiamMsgFromNavMap1(t *testing.T) {
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

	nM := utils.NewOrderedNavigableMap()
	path := []string{"Multiple-Services-Credit-Control", "Final-Unit-Indication", "Redirect-Server", "Redirect-Address-Type"}
	itm := &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: datatype.Enumerated(2),
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})

	path = []string{"Multiple-Services-Credit-Control", "Tariff-Change-Usage"}
	itm = &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: datatype.Enumerated(2),
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})

	path = []string{"Multiple-Services-Credit-Control", "Final-Unit-Indication", "Redirect-Server", "Redirect-Server-Address"}
	itm = &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: "http://172.10.88.88/",
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})

	if err := updateDiamMsgFromNavMap(m2, nM, ""); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(eMessage.String(), m2.String()) {
		t.Errorf("Expected %s, received %s", utils.ToJSON(eMessage), utils.ToJSON(m2))
	}
}

func TestLibDiamUpdateDiamMsgFromNavMap2(t *testing.T) {
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

	nM := utils.NewOrderedNavigableMap()
	path := []string{"Multiple-Services-Credit-Control", "Final-Unit-Indication", "Redirect-Server", "Redirect-Address-Type"}
	itm := &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: datatype.Enumerated(2),
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})

	path = []string{"Multiple-Services-Credit-Control", "Tariff-Change-Usage"}
	itm = &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: datatype.Enumerated(2),
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})

	path = []string{"Multiple-Services-Credit-Control", "Final-Unit-Indication", "Redirect-Server", "Redirect-Server-Address"}
	itm = &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data:      "http://172.10.88.88/",
		NewBranch: true,
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})

	path = []string{"Multiple-Services-Credit-Control", "Final-Unit-Indication", "Redirect-Server", "Redirect-Address-Type"}
	itm = &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: datatype.Enumerated(2),
	}}
	nM.Append(
		&utils.FullPath{
			Path:      strings.Join(path, utils.NestingSep),
			PathSlice: []string{"Multiple-Services-Credit-Control", "Final-Unit-Indication", "Redirect-Server", "Redirect-Address-Type"},
		}, itm.Value)
	if err := updateDiamMsgFromNavMap(m2, nM, ""); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(eMessage.String(), m2.String()) {
		t.Errorf("Expected %s, received %s", utils.ToJSON(eMessage), utils.ToJSON(m2))
	}
}

func TestLibDiamUpdateDiamMsgFromNavMap3(t *testing.T) {
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

	nM := utils.NewOrderedNavigableMap()

	path := []string{"Multiple-Services-Credit-Control", "Final-Unit-Indication", "Final-Unit-Action"}
	itm := &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: datatype.Enumerated(1),
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})

	path = []string{"Multiple-Services-Credit-Control", "Final-Unit-Indication", "Redirect-Server", "Redirect-Address-Type"}
	itm = &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: datatype.Enumerated(2),
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})

	path = []string{"Multiple-Services-Credit-Control", "Tariff-Change-Usage"}
	itm = &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: datatype.Enumerated(2),
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})

	path = []string{"Multiple-Services-Credit-Control", "Final-Unit-Indication", "Redirect-Server", "Redirect-Server-Address"}
	itm = &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: "http://172.10.88.88/",
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})

	if err := updateDiamMsgFromNavMap(m2, nM, ""); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(eMessage.String(), m2.String()) {
		t.Errorf("Expected %s, received %s", utils.ToJSON(eMessage), utils.ToJSON(m2))
	}
}

func TestLibDiamUpdateDiamMsgFromNavMap4(t *testing.T) {
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

	nM := utils.NewOrderedNavigableMap()

	path := []string{"Multiple-Services-Credit-Control", "Final-Unit-Indication", "Final-Unit-Action"}
	itm := &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: datatype.Enumerated(1),
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})

	path = []string{"Multiple-Services-Credit-Control", "Final-Unit-Indication", "Redirect-Server", "Redirect-Address-Type"}
	itm = &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: datatype.Enumerated(2),
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})

	path = []string{"Multiple-Services-Credit-Control", "Tariff-Change-Usage"}
	itm = &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: datatype.Enumerated(2),
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})

	path = []string{"Multiple-Services-Credit-Control", "Final-Unit-Indication", "Redirect-Server", "Redirect-Server-Address"}
	itm = &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: "http://172.10.88.88/",
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})

	path = []string{"Multiple-Services-Credit-Control", "Final-Unit-Indication", "Redirect-Server", "Redirect-Server-Address"}
	itm2 := &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data:      "http://172.10.88.88/",
		NewBranch: true,
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm, itm2})

	path = []string{"Granted-Service-Unit", "CC-Time"}
	itm = &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: datatype.Unsigned32(10),
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})

	if err := updateDiamMsgFromNavMap(m2, nM, ""); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(eMessage.String(), m2.String()) {
		t.Errorf("Expected %s, received %s", utils.ToJSON(eMessage), utils.ToJSON(m2))
	}
}

func TestLibDiamAVPAsIface(t *testing.T) {
	args := diam.NewAVP(435, avp.Mbit, 0, datatype.Address("127.0.0.1"))
	var exp any = net.IP([]byte("127.0.0.1"))
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

	args = diam.NewAVP(257, avp.Mbit, 0, datatype.Address("10.170.248.140"))
	exp = net.IP([]byte("10.170.248.140"))
	if rply, err := diamAVPAsIface(args); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected<%T>: %v ,received<%T>: %v ", exp, exp, rply, rply)
	}
}

func TestLibDiamNewDiamDataType(t *testing.T) {
	argType := datatype.AddressType
	argVal := "127.0.0.1"
	var exp datatype.Type = datatype.Address(net.ParseIP("127.0.0.1"))
	if rply, err := newDiamDataType(argType, argVal, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected<%T>: %v ,received<%T>: %v ", exp, exp, rply, rply)
	}

	argType = datatype.DiameterIdentityType
	argVal = "diam1"
	exp = datatype.DiameterIdentity("diam1")
	if rply, err := newDiamDataType(argType, argVal, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected<%T>: %v ,received<%T>: %v ", exp, exp, rply, rply)
	}

	argType = datatype.DiameterURIType
	argVal = "http://172.10.88.88/"
	exp = datatype.DiameterURI("http://172.10.88.88/")
	if rply, err := newDiamDataType(argType, argVal, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected<%T>: %v ,received<%T>: %v ", exp, exp, rply, rply)
	}

	argType = datatype.EnumeratedType
	argVal = "00"
	exp = datatype.Enumerated(int32(0))
	if rply, err := newDiamDataType(argType, argVal, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected<%T>: %v ,received<%T>: %v ", exp, exp, rply, rply)
	}

	argVal = "0.0"
	if rply, err := newDiamDataType(argType, argVal, ""); err == nil {
		t.Errorf("Expected err received: err: %v, rply %v", err, rply)
	}

	argType = datatype.Float32Type
	argVal = "00"
	exp = datatype.Float32(float32(0))
	if rply, err := newDiamDataType(argType, argVal, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected<%T>: %v ,received<%T>: %v ", exp, exp, rply, rply)
	}

	argVal = "0.0A"
	if rply, err := newDiamDataType(argType, argVal, ""); err == nil {
		t.Errorf("Expected err received: err: %v, rply %v", err, rply)
	}

	argType = datatype.Float64Type
	argVal = "00"
	exp = datatype.Float64(float64(0))
	if rply, err := newDiamDataType(argType, argVal, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected<%T>: %v ,received<%T>: %v ", exp, exp, rply, rply)
	}

	argVal = "0.0A"
	if rply, err := newDiamDataType(argType, argVal, ""); err == nil {
		t.Errorf("Expected err received: err: %v, rply %v", err, rply)
	}

	argType = datatype.IPFilterRuleType
	argVal = "filter1"
	exp = datatype.IPFilterRule("filter1")
	if rply, err := newDiamDataType(argType, argVal, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected<%T>: %v ,received<%T>: %v ", exp, exp, rply, rply)
	}

	argType = datatype.IPv4Type
	argVal = "127.0.0.1"
	exp = datatype.IPv4(net.ParseIP("127.0.0.1"))
	if rply, err := newDiamDataType(argType, argVal, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected<%T>: %v ,received<%T>: %v ", exp, exp, rply, rply)
	}

	argType = datatype.Integer32Type
	argVal = "00"
	exp = datatype.Integer32(int32(0))
	if rply, err := newDiamDataType(argType, argVal, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected<%T>: %v ,received<%T>: %v ", exp, exp, rply, rply)
	}

	argVal = "0.0"
	if rply, err := newDiamDataType(argType, argVal, ""); err == nil {
		t.Errorf("Expected err received: err: %v, rply %v", err, rply)
	}

	argType = datatype.Integer64Type
	argVal = "00"
	exp = datatype.Integer64(int64(0))
	if rply, err := newDiamDataType(argType, argVal, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected<%T>: %v ,received<%T>: %v ", exp, exp, rply, rply)
	}

	argVal = "0.0"
	if rply, err := newDiamDataType(argType, argVal, ""); err == nil {
		t.Errorf("Expected err received: err: %v, rply %v", err, rply)
	}

	argType = datatype.OctetStringType
	argVal = "diam1"
	exp = datatype.OctetString("diam1")
	if rply, err := newDiamDataType(argType, argVal, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected<%T>: %v ,received<%T>: %v ", exp, exp, rply, rply)
	}

	argType = datatype.QoSFilterRuleType
	argVal = "diam1"
	exp = datatype.QoSFilterRule("diam1")
	if rply, err := newDiamDataType(argType, argVal, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected<%T>: %v ,received<%T>: %v ", exp, exp, rply, rply)
	}

	argType = datatype.UTF8StringType
	argVal = "diam1"
	exp = datatype.UTF8String("diam1")
	if rply, err := newDiamDataType(argType, argVal, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected<%T>: %v ,received<%T>: %v ", exp, exp, rply, rply)
	}

	argType = datatype.Unsigned32Type
	argVal = "00"
	exp = datatype.Unsigned32(uint32(0))
	if rply, err := newDiamDataType(argType, argVal, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected<%T>: %v ,received<%T>: %v ", exp, exp, rply, rply)
	}

	argVal = "0.0"
	if rply, err := newDiamDataType(argType, argVal, ""); err == nil {
		t.Errorf("Expected err received: err: %v, rply %v", err, rply)
	}

	argType = datatype.Unsigned64Type
	argVal = "00"
	exp = datatype.Unsigned64(uint64(0))
	if rply, err := newDiamDataType(argType, argVal, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected<%T>: %v ,received<%T>: %v ", exp, exp, rply, rply)
	}

	argVal = "0.0"
	if rply, err := newDiamDataType(argType, argVal, ""); err == nil {
		t.Errorf("Expected err received: err: %v, rply %v", err, rply)
	}

	now := time.Now().UTC()

	argType = datatype.TimeType
	argVal = now.String()
	exp = datatype.Time(now)
	if rply, err := newDiamDataType(argType, argVal, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected<%T>: %v ,received<%T>: %v ", exp, exp, rply, rply)
	}

	argVal = "0.0"
	if rply, err := newDiamDataType(argType, argVal, ""); err == nil {
		t.Errorf("Expected err received: err: %v, rply %v", err, rply)
	}

	argType = datatype.GroupedType
	argVal = "{}"
	if rply, err := newDiamDataType(argType, argVal, ""); err == nil {
		t.Errorf("Expected err received: err: %v, rply %v", err, rply)
	}

	argType = datatype.AddressType
	argVal = "10.170.248.140"
	exp = datatype.Address(net.ParseIP("10.170.248.140"))
	if rply, err := newDiamDataType(argType, argVal, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected<%T>: %v ,received<%T>: %v ", exp, exp, rply, rply)
	}
}

func TestLibDiamAvpGroupIface(t *testing.T) {
	avps := diam.NewRequest(diam.CreditControl, 4, nil)
	avps.NewAVP("Multiple-Services-Credit-Control", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(432, avp.Mbit, 0, datatype.Unsigned32(1)),
		}})
	avps.NewAVP("Multiple-Services-Credit-Control", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(432, avp.Mbit, 0, datatype.Unsigned32(99)),
		}})
	dP := newDADataProvider(nil, avps)
	eOut := any(uint32(1))
	if out, err := dP.FieldAsInterface([]string{"Multiple-Services-Credit-Control", "Rating-Group"}); err != nil {
		t.Error(err)
	} else if eOut != out {
		t.Errorf("Expecting: %v, received: %v", eOut, out)
	}
	dP.(*diameterDP).cache = utils.MapStorage{}
	if out, err := dP.FieldAsInterface([]string{"Multiple-Services-Credit-Control", "Rating-Group[~Rating-Group(1)]"}); err != nil {
		t.Error(err)
	} else if eOut != out {
		t.Errorf("Expecting: %v, received: %v", eOut, out)
	}
	dP.(*diameterDP).cache = utils.MapStorage{}
	eOut = any(uint32(99))
	if out, err := dP.FieldAsInterface([]string{"Multiple-Services-Credit-Control", "Rating-Group[1]"}); err != nil {
		t.Error(err)
	} else if eOut != out {
		t.Errorf("Expecting: %v, received: %v", eOut, out)
	}
	dP.(*diameterDP).cache = utils.MapStorage{}
	if out, err := dP.FieldAsInterface([]string{"Multiple-Services-Credit-Control", "Rating-Group[~Rating-Group(99)]"}); err != nil {
		t.Error(err)
	} else if eOut != out {
		t.Errorf("Expecting: %v, received: %v", eOut, out)
	}
	dP.(*diameterDP).cache = utils.MapStorage{}
	if _, err := dP.FieldAsInterface([]string{"Multiple-Services-Credit-Control", "Rating-Group[~Rating-Group(10)]"}); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func TestLibDiamFilterWithDiameterDP(t *testing.T) {
	avps := diam.NewRequest(diam.CreditControl, 4, nil)
	avps.NewAVP("Multiple-Services-Credit-Control", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(432, avp.Mbit, 0, datatype.Unsigned32(1)),
		}})
	avps.NewAVP("Multiple-Services-Credit-Control", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(432, avp.Mbit, 0, datatype.Unsigned32(99)),
		}})
	dP := newDADataProvider(nil, avps)
	cfg := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items),
		config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	agReq := NewAgentRequest(dP, nil, nil, nil, nil, nil, "cgrates.org", "", filterS, nil)

	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*exists:~*req.Multiple-Services-Credit-Control.Rating-Group[~Rating-Group(99)]:"}, agReq); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Exptected true, received: %+v", pass)
	}

	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*exists:~*req.Multiple-Services-Credit-Control.Rating-Group[~Rating-Group(10)]:"}, agReq); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Exptected false, received: %+v", pass)
	}

	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*string:~*req.Multiple-Services-Credit-Control.Rating-Group[~Rating-Group(10)]:12"}, agReq); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Exptected false, received: %+v", pass)
	}

	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*string:~*req.Multiple-Services-Credit-Control.Rating-Group[~Rating-Group(1)]:1"}, agReq); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Exptected true, received: %+v", pass)
	}
}

func TestLibDiamHeaderLen(t *testing.T) {
	tests := []struct {
		name    string
		avp     *diam.AVP
		wantLen int
	}{
		{
			name:    "len equals",
			avp:     &diam.AVP{Flags: avp.Vbit},
			wantLen: 12,
		},
		{
			name:    "len not set",
			avp:     &diam.AVP{},
			wantLen: 8,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLen := headerLen(tt.avp)
			if gotLen != tt.wantLen {
				t.Errorf("headerLen() returned %d, want %d", gotLen, tt.wantLen)
			}
		})
	}
}

func TestLibDiamAVPAsString(t *testing.T) {
	testCases := []struct {
		name      string
		dAVP      *diam.AVP
		expected  string
		expectErr bool
	}{
		{
			name:     "IPv4 Address AVP",
			dAVP:     &diam.AVP{Data: datatype.Address(net.IPv4(192, 168, 0, 1))},
			expected: "192.168.0.1",
		},
		{
			name:     "Diameter Identity AVP",
			dAVP:     &diam.AVP{Data: datatype.DiameterIdentity("cgrates.com")},
			expected: "cgrates.com",
		},
		{
			name:     "Time AVP",
			dAVP:     &diam.AVP{Data: datatype.Time(time.Now())},
			expected: time.Now().Format(time.RFC3339),
		},
		{
			name:      "Nil AVP",
			dAVP:      nil,
			expectErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := diamAVPAsString(tc.dAVP)
			if tc.expectErr {
				if err == nil {
					t.Error("Expected an error, but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestLibDiamUpdateAVPLength(t *testing.T) {

	testCases := []struct {
		name     string
		avps     []*diam.AVP
		expected int
	}{
		{
			name: "Single AVP without GroupedAVP",
			avps: []*diam.AVP{
				{Length: 10},
				{Length: 20},
				{Length: 15},
			},
			expected: 45,
		},
		{
			name:     "Empty AVP slice",
			avps:     []*diam.AVP{},
			expected: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			result := updateAVPLength(tc.avps)

			if result != tc.expected {
				t.Errorf("Expected length %d, got %d", tc.expected, result)
			}
		})
	}
}

func TestLibDiamUpdateAVPLengthWithGroupedAVP(t *testing.T) {
	groupedAVP := &diam.AVP{Length: 10}
	avps := []*diam.AVP{
		{Length: 10},
		{
			Data: &diam.GroupedAVP{
				AVP: []*diam.AVP{
					groupedAVP,
					{Length: 0},
				},
			},
		},
		{Length: 15},
	}
	expected := 10 + headerLen(avps[1]) + groupedAVP.Length + avps[2].Length
	result := updateAVPLength(avps)
	if result != expected {
		t.Errorf("Expected length %d, got %d", expected, result)
	}
}

func TestLibDiamDiameterDPString(t *testing.T) {
	msg := &diam.Message{
		Header: &diam.Header{
			Version:       1,
			MessageLength: 20,
			CommandFlags:  2,
			CommandCode:   272,
			ApplicationID: 0,
			HopByHopID:    12345,
			EndToEndID:    67890,
		},
	}
	dp := &diameterDP{
		m: msg,
	}
	result := dp.String()
	expected := msg.String()
	if result != expected {
		t.Errorf("Expected %q, but got %q", expected, result)
	}
}

func TestLibDiamLoadDictionaries(t *testing.T) {
	tempDir := t.TempDir()
	dictsDir := filepath.Join(tempDir, "dicts")
	err := os.Mkdir(dictsDir, os.ModePerm)
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	testCases := []struct {
		name             string
		dictsDir         string
		expectedErrorMsg string
	}{
		{
			name:     "Valid Directory",
			dictsDir: dictsDir,
		},
		{
			name:             "Invalid Directory",
			dictsDir:         filepath.Join(tempDir, "nonexistent", "directory"),
			expectedErrorMsg: "Invalid dictionaries folder",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := loadDictionaries(tc.dictsDir, "testComponent")
			if tc.expectedErrorMsg == "" && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tc.expectedErrorMsg != "" && !strings.Contains(err.Error(), tc.expectedErrorMsg) {
				t.Errorf("Expected error containing %q, got: %v", tc.expectedErrorMsg, err)
			}
		})
	}
}
