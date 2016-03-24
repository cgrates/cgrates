/*
Real-time Charging System for Telecom & ISP environments
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
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/sessionmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/fiorix/go-diameter/diam"
	"github.com/fiorix/go-diameter/diam/avp"
	"github.com/fiorix/go-diameter/diam/datatype"
	"github.com/fiorix/go-diameter/diam/dict"
)

func TestDisectUsageForCCR(t *testing.T) {
	if reqType, reqNr, reqCCTime, usedCCTime := disectUsageForCCR(time.Duration(0)*time.Second, time.Duration(300)*time.Second, false); reqType != 1 || reqNr != 0 || reqCCTime != 300 || usedCCTime != 0 {
		t.Error(reqType, reqNr, reqCCTime, usedCCTime)
	}
	if reqType, reqNr, reqCCTime, usedCCTime := disectUsageForCCR(time.Duration(35)*time.Second, time.Duration(300)*time.Second, false); reqType != 2 || reqNr != 0 || reqCCTime != 300 || usedCCTime != 35 {
		t.Error(reqType, reqNr, reqCCTime, usedCCTime)
	}
	if reqType, reqNr, reqCCTime, usedCCTime := disectUsageForCCR(time.Duration(935)*time.Second, time.Duration(300)*time.Second, false); reqType != 2 || reqNr != 3 || reqCCTime != 300 || usedCCTime != 35 {
		t.Error(reqType, reqNr, reqCCTime, usedCCTime)
	}
	if reqType, reqNr, reqCCTime, usedCCTime := disectUsageForCCR(time.Duration(35)*time.Second, time.Duration(300)*time.Second, true); reqType != 3 || reqNr != 1 || reqCCTime != 0 || usedCCTime != 35 {
		t.Error(reqType, reqNr, reqCCTime, usedCCTime)
	}
	if reqType, reqNr, reqCCTime, usedCCTime := disectUsageForCCR(time.Duration(610)*time.Second, time.Duration(300)*time.Second, true); reqType != 3 || reqNr != 3 || reqCCTime != 0 || usedCCTime != 10 {
		t.Error(reqType, reqNr, reqCCTime, usedCCTime)
	}
	if reqType, reqNr, reqCCTime, usedCCTime := disectUsageForCCR(time.Duration(935)*time.Second, time.Duration(300)*time.Second, true); reqType != 3 || reqNr != 4 || reqCCTime != 0 || usedCCTime != 35 {
		t.Error(reqType, reqNr, reqCCTime, usedCCTime)
	}
}

func TestUsageFromCCR(t *testing.T) {
	if usage := usageFromCCR(1, 0, 300, 0, time.Duration(300)*time.Second); usage != time.Duration(300)*time.Second {
		t.Error(usage)
	}
	if usage := usageFromCCR(2, 0, 300, 300, time.Duration(300)*time.Second); usage != time.Duration(300)*time.Second {
		t.Error(usage)
	}
	if usage := usageFromCCR(2, 3, 300, 300, time.Duration(300)*time.Second); usage != time.Duration(300)*time.Second {
		t.Error(usage.Seconds())
	}
	if usage := usageFromCCR(3, 3, 0, 10, time.Duration(300)*time.Second); usage != time.Duration(610)*time.Second {
		t.Error(usage)
	}
	if usage := usageFromCCR(3, 4, 0, 35, time.Duration(300)*time.Second); usage != time.Duration(935)*time.Second {
		t.Error(usage)
	}
	if usage := usageFromCCR(3, 1, 0, 35, time.Duration(300)*time.Second); usage != time.Duration(35)*time.Second {
		t.Error(usage)
	}
	if usage := usageFromCCR(1, 0, 360, 0, time.Duration(360)*time.Second); usage != time.Duration(360)*time.Second {
		t.Error(usage)
	}
}

func TestAvpValAsString(t *testing.T) {
	originHostStr := "unit_test"
	a := diam.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity(originHostStr))
	if avpValStr := avpValAsString(a); avpValStr != originHostStr {
		t.Errorf("Expected: %s, received: %s", originHostStr, avpValStr)
	}
}

func TestMetaValueExponent(t *testing.T) {
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
	if val, err := metaValueExponent(m, utils.ParseRSRFieldsMustCompile("Requested-Service-Unit>CC-Money>Unit-Value>Value-Digits;^|;Requested-Service-Unit>CC-Money>Unit-Value>Exponent", utils.INFIELD_SEP), 10); err != nil {
		t.Error(err)
	} else if val != "0.1" {
		t.Error("Received: ", val)
	}
	if _, err = metaValueExponent(m, utils.ParseRSRFieldsMustCompile("Requested-Service-Unit>CC-Money>Unit-Value>Value-Digits;Requested-Service-Unit>CC-Money>Unit-Value>Exponent", utils.INFIELD_SEP), 10); err == nil {
		t.Error("Should have received error") // Insufficient number arguments
	}
}

func TestMetaSum(t *testing.T) {
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
	if val, err := metaSum(m, utils.ParseRSRFieldsMustCompile("Requested-Service-Unit>CC-Money>Unit-Value>Value-Digits;^|;Requested-Service-Unit>CC-Money>Unit-Value>Exponent", utils.INFIELD_SEP), 0, 10); err != nil {
		t.Error(err)
	} else if val != "9995" {
		t.Error("Received: ", val)
	}
	if _, err = metaSum(m, utils.ParseRSRFieldsMustCompile("Requested-Service-Unit>CC-Money>Unit-Value>Value-Digits;Requested-Service-Unit>CC-Money>Unit-Value>Exponent", utils.INFIELD_SEP), 0, 10); err == nil {
		t.Error("Should have received error") // Insufficient number arguments
	}
}

func TestFieldOutVal(t *testing.T) {
	m := diam.NewRequest(diam.CreditControl, 4, nil)
	m.NewAVP("Session-Id", avp.Mbit, 0, datatype.UTF8String("simuhuawei;1449573472;00002"))
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
	m.NewAVP("Service-Identifier", avp.Mbit, 0, datatype.Unsigned32(0))
	m.NewAVP("Requested-Service-Unit", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(420, avp.Mbit, 0, datatype.Unsigned32(360))}}) // CC-Time
	cfgFld := &config.CfgCdrField{Tag: "StaticTest", Type: utils.META_COMPOSED, FieldId: utils.TOR,
		Value: utils.ParseRSRFieldsMustCompile("^*voice", utils.INFIELD_SEP), Mandatory: true}
	eOut := "*voice"
	if fldOut, err := fieldOutVal(m, cfgFld, time.Duration(0), nil); err != nil {
		t.Error(err)
	} else if fldOut != eOut {
		t.Errorf("Expecting: %s, received: %s", eOut, fldOut)
	}
	cfgFld = &config.CfgCdrField{Tag: "ComposedTest", Type: utils.META_COMPOSED, FieldId: utils.DESTINATION,
		Value: utils.ParseRSRFieldsMustCompile("Requested-Service-Unit>CC-Time", utils.INFIELD_SEP), Mandatory: true}
	eOut = "360"
	if fldOut, err := fieldOutVal(m, cfgFld, time.Duration(0), nil); err != nil {
		t.Error(err)
	} else if fldOut != eOut {
		t.Errorf("Expecting: %s, received: %s", eOut, fldOut)
	}
	// With filter on ProcessorVars
	cfgFld = &config.CfgCdrField{Tag: "ComposedTestWithProcessorVarsFilter", Type: utils.META_COMPOSED, FieldId: utils.DESTINATION,
		FieldFilter: utils.ParseRSRFieldsMustCompile("CGRError(INSUFFICIENT_CREDIT)", utils.INFIELD_SEP),
		Value:       utils.ParseRSRFieldsMustCompile("Requested-Service-Unit>CC-Time", utils.INFIELD_SEP), Mandatory: true}
	if _, err := fieldOutVal(m, cfgFld, time.Duration(0), nil); err == nil {
		t.Error("Should have error")
	}
	eOut = "360"
	if fldOut, err := fieldOutVal(m, cfgFld, time.Duration(0), map[string]string{"CGRError": "INSUFFICIENT_CREDIT"}); err != nil {
		t.Error(err)
	} else if fldOut != eOut {
		t.Errorf("Expecting: %s, received: %s", eOut, fldOut)
	}
	// Without filter, we shoud get always the first subscriptionId
	cfgFld = &config.CfgCdrField{Tag: "Grouped1", Type: utils.MetaGrouped, FieldId: "Account",
		Value: utils.ParseRSRFieldsMustCompile("Subscription-Id>Subscription-Id-Data", utils.INFIELD_SEP), Mandatory: true}
	eOut = "33708000003"
	if fldOut, err := fieldOutVal(m, cfgFld, time.Duration(0), nil); err != nil {
		t.Error(err)
	} else if fldOut != eOut {
		t.Errorf("Expecting: %s, received: %s", eOut, fldOut)
	}
	// Without groupedAVP, we shoud get the first subscriptionId
	cfgFld = &config.CfgCdrField{Tag: "Grouped2", Type: utils.MetaGrouped, FieldId: "Account",
		FieldFilter: utils.ParseRSRFieldsMustCompile("Subscription-Id>Subscription-Id-Type(1)", utils.INFIELD_SEP),
		Value:       utils.ParseRSRFieldsMustCompile("Subscription-Id>Subscription-Id-Data", utils.INFIELD_SEP), Mandatory: true}
	eOut = "208708000003"
	if fldOut, err := fieldOutVal(m, cfgFld, time.Duration(0), nil); err != nil {
		t.Error(err)
	} else if fldOut != eOut {
		t.Errorf("Expecting: %s, received: %s", eOut, fldOut)
	}
}

func TestSerializeAVPValueFromString(t *testing.T) {
	dictAVP, _ := dict.Default.FindAVP(4, "Session-Id")
	eValByte := []byte("simuhuawei;1449573472;00002")
	if valByte, err := serializeAVPValueFromString(dictAVP, "simuhuawei;1449573472;00002", "UTC"); err != nil {
		t.Error(err)
	} else if !bytes.Equal(eValByte, valByte) {
		t.Errorf("Expecting: %+v, received: %+v", eValByte, valByte)
	}
	dictAVP, _ = dict.Default.FindAVP(4, "Result-Code")
	eValByte = make([]byte, 4)
	binary.BigEndian.PutUint32(eValByte, uint32(5031))
	if valByte, err := serializeAVPValueFromString(dictAVP, "5031", "UTC"); err != nil {
		t.Error(err)
	} else if !bytes.Equal(eValByte, valByte) {
		t.Errorf("Expecting: %+v, received: %+v", eValByte, valByte)
	}
}

func TestMessageSetAVPsWithPath(t *testing.T) {
	eMessage := diam.NewRequest(diam.CreditControl, 4, nil)
	eMessage.NewAVP("Session-Id", avp.Mbit, 0, datatype.UTF8String("simuhuawei;1449573472;00002"))
	m := diam.NewMessage(diam.CreditControl, diam.RequestFlag, 4, eMessage.Header.HopByHopID, eMessage.Header.EndToEndID, nil)
	if err := messageSetAVPsWithPath(m, []interface{}{"Session-Id", "Unknown"}, "simuhuawei;1449573472;00002", false, "UTC"); err == nil || err.Error() != "Could not find AVP Unknown" {
		t.Error(err)
	}
	if err := messageSetAVPsWithPath(m, []interface{}{"Session-Id"}, "simuhuawei;1449573472;00002", false, "UTC"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eMessage, m) {
		t.Errorf("Expecting: %+v, received: %+v", eMessage, m)
	}
	// test append
	eMessage.NewAVP("Session-Id", avp.Mbit, 0, datatype.UTF8String("simuhuawei;1449573472;00003"))
	if err := messageSetAVPsWithPath(m, []interface{}{"Session-Id"}, "simuhuawei;1449573472;00003", true, "UTC"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eMessage, m) {
		t.Errorf("Expecting: %+v, received: %+v", eMessage, m)
	}
	// test overwrite
	eMessage = diam.NewRequest(diam.CreditControl, 4, nil)
	eMessage.NewAVP("Session-Id", avp.Mbit, 0, datatype.UTF8String("simuhuawei;1449573472;00002"))
	m = diam.NewMessage(diam.CreditControl, diam.RequestFlag, 4, eMessage.Header.HopByHopID, eMessage.Header.EndToEndID, nil)
	if err := messageSetAVPsWithPath(m, []interface{}{"Session-Id"}, "simuhuawei;1449573472;00001", false, "UTC"); err != nil {
		t.Error(err)
	}
	if err := messageSetAVPsWithPath(m, []interface{}{"Session-Id"}, "simuhuawei;1449573472;00002", false, "UTC"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eMessage, m) {
		t.Errorf("Expecting: %+v, received: %+v", eMessage, m)
	}
	eMessage = diam.NewRequest(diam.CreditControl, 4, nil)
	eMessage.NewAVP("Subscription-Id", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(444, avp.Mbit, 0, datatype.UTF8String("33708000003")), // Subscription-Id-Data
		}})
	m = diam.NewMessage(diam.CreditControl, diam.RequestFlag, 4, eMessage.Header.HopByHopID, eMessage.Header.EndToEndID, nil)
	if err := messageSetAVPsWithPath(m, []interface{}{"Subscription-Id", "Subscription-Id-Data"}, "33708000003", false, "UTC"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eMessage, m) {
		t.Errorf("Expecting: %+v, received: %+v", eMessage, m)
	}
	// test append
	eMessage.NewAVP("Subscription-Id", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(450, avp.Mbit, 0, datatype.Enumerated(0)), // Subscription-Id-Data
		}})
	if err := messageSetAVPsWithPath(m, []interface{}{"Subscription-Id", "Subscription-Id-Type"}, "0", true, "UTC"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eMessage, m) {
		t.Errorf("Expecting: %+v, received: %+v", eMessage, m)
	}
	// test group append
	eMessage = diam.NewRequest(diam.CreditControl, 4, nil)
	eMessage.NewAVP("Subscription-Id", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(450, avp.Mbit, 0, datatype.Enumerated(0)),             // Subscription-Id-Data
			diam.NewAVP(444, avp.Mbit, 0, datatype.UTF8String("33708000003")), // Subscription-Id-Data
		}})
	eMsgSrl, _ := eMessage.Serialize()
	m = diam.NewMessage(diam.CreditControl, diam.RequestFlag, 4, eMessage.Header.HopByHopID, eMessage.Header.EndToEndID, nil)
	if err := messageSetAVPsWithPath(m, []interface{}{"Subscription-Id", "Subscription-Id-Type"}, "0", false, "UTC"); err != nil {
		t.Error(err)
	}
	if err := messageSetAVPsWithPath(m, []interface{}{"Subscription-Id", "Subscription-Id-Data"}, "33708000003", false, "UTC"); err != nil {
		t.Error(err)
	} else {
		mSrl, _ := m.Serialize()
		if !bytes.Equal(eMsgSrl, mSrl) {
			t.Errorf("Expecting: %+v, received: %+v", eMessage, m)
		}
	}
	eMessage = diam.NewRequest(diam.CreditControl, 4, nil)
	eMessage.NewAVP("Granted-Service-Unit", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(420, avp.Mbit, 0, datatype.Unsigned32(300)), // Subscription-Id-Data
		}})
	m = diam.NewMessage(diam.CreditControl, diam.RequestFlag, 4, eMessage.Header.HopByHopID, eMessage.Header.EndToEndID, nil)
	if err := messageSetAVPsWithPath(m, []interface{}{"Granted-Service-Unit", "CC-Time"}, "300", false, "UTC"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eMessage, m) {
		t.Errorf("Expecting: %+v, received: %+v", eMessage, m)
	}
	// Multiple append
	eMessage = diam.NewRequest(diam.CreditControl, 4, nil)
	eMessage.NewAVP("Multiple-Services-Credit-Control", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(431, avp.Mbit, 0, &diam.GroupedAVP{ // Granted-Service-Unit
				AVP: []*diam.AVP{
					diam.NewAVP(420, avp.Mbit, 0, datatype.Unsigned32(3600)),
					diam.NewAVP(421, avp.Mbit, 0, datatype.Unsigned64(153600)), // "CC-Total-Octets"
				},
			}),
			diam.NewAVP(432, avp.Mbit, 0, datatype.Unsigned32(10)),
		},
	})
	eMessage.NewAVP("Multiple-Services-Credit-Control", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(431, avp.Mbit, 0, &diam.GroupedAVP{ // Granted-Service-Unit
				AVP: []*diam.AVP{
					diam.NewAVP(420, avp.Mbit, 0, datatype.Unsigned32(2600)),
					diam.NewAVP(421, avp.Mbit, 0, datatype.Unsigned64(143600)), // "CC-Total-Octets"
				},
			}),
			diam.NewAVP(432, avp.Mbit, 0, datatype.Unsigned32(11)), // Rating-Group
		},
	})
	m = diam.NewMessage(diam.CreditControl, diam.RequestFlag, 4, eMessage.Header.HopByHopID, eMessage.Header.EndToEndID, nil)
	if err := messageSetAVPsWithPath(m, []interface{}{"Multiple-Services-Credit-Control", "Granted-Service-Unit", "CC-Time"}, "3600", false, "UTC"); err != nil {
		t.Error(err)
	}
	if err := messageSetAVPsWithPath(m, []interface{}{"Multiple-Services-Credit-Control", "Granted-Service-Unit", "CC-Total-Octets"}, "153600", false, "UTC"); err != nil {
		t.Error(err)
	}
	if err := messageSetAVPsWithPath(m, []interface{}{"Multiple-Services-Credit-Control", "Rating-Group"}, "10", false, "UTC"); err != nil {
		t.Error(err)
	}
	if err := messageSetAVPsWithPath(m, []interface{}{"Multiple-Services-Credit-Control", "Granted-Service-Unit", "CC-Time"}, "2600", true, "UTC"); err != nil {
		t.Error(err)
	}
	if err := messageSetAVPsWithPath(m, []interface{}{"Multiple-Services-Credit-Control", "Granted-Service-Unit", "CC-Total-Octets"}, "143600", false, "UTC"); err != nil {
		t.Error(err)
	}
	if err := messageSetAVPsWithPath(m, []interface{}{"Multiple-Services-Credit-Control", "Rating-Group"}, "11", false, "UTC"); err != nil {
		t.Error(err)
	}
	if fmt.Sprintf("%q", eMessage) != fmt.Sprintf("%q", m) { // test with fmt since reflect.DeepEqual does not perform properly here
		t.Errorf("Expecting: %+v, received: %+v", eMessage, m)
	}
}

func TestCCASetProcessorAVPs(t *testing.T) {
	ccr := &CCR{ // Bare information, just the one needed for answer
		SessionId:         "routinga;1442095190;1476802709",
		AuthApplicationId: 4,
		CCRequestType:     1,
		CCRequestNumber:   0,
	}
	ccr.diamMessage = ccr.AsBareDiameterMessage()
	ccr.diamMessage.NewAVP("Subscription-Id", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(450, avp.Mbit, 0, datatype.Enumerated(0)),             // Subscription-Id-Type
			diam.NewAVP(444, avp.Mbit, 0, datatype.UTF8String("33708000003")), // Subscription-Id-Data
		}})
	ccr.debitInterval = time.Duration(300) * time.Second
	cca := NewBareCCAFromCCR(ccr, "CGR-DA", "cgrates.org")
	reqProcessor := &config.DARequestProcessor{Id: "UNIT_TEST", // Set template for tests
		CCAFields: []*config.CfgCdrField{
			&config.CfgCdrField{Tag: "Subscription-Id/Subscription-Id-Type", Type: utils.META_COMPOSED,
				FieldId: "Subscription-Id>Subscription-Id-Type",
				Value:   utils.ParseRSRFieldsMustCompile("Subscription-Id>Subscription-Id-Type", utils.INFIELD_SEP), Mandatory: true},
			&config.CfgCdrField{Tag: "Subscription-Id/Subscription-Id-Data", Type: utils.META_COMPOSED,
				FieldId: "Subscription-Id>Subscription-Id-Data",
				Value:   utils.ParseRSRFieldsMustCompile("Subscription-Id>Subscription-Id-Data", utils.INFIELD_SEP), Mandatory: true},
		},
	}
	eMessage := cca.AsDiameterMessage()
	eMessage.NewAVP("Subscription-Id", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(450, avp.Mbit, 0, datatype.Enumerated(0)),             // Subscription-Id-Type
			diam.NewAVP(444, avp.Mbit, 0, datatype.UTF8String("33708000003")), // Subscription-Id-Data
		}})
	if err := cca.SetProcessorAVPs(reqProcessor, map[string]string{}); err != nil {
		t.Error(err)
	} else if ccaMsg := cca.AsDiameterMessage(); !reflect.DeepEqual(eMessage, ccaMsg) {
		t.Errorf("Expecting: %+v, received: %+v", eMessage, ccaMsg)
	}
}

func TestCCRAsSMGenericEvent(t *testing.T) {
	ccr := &CCR{ // Bare information, just the one needed for answer
		SessionId:         "ccrasgen1",
		AuthApplicationId: 4,
		CCRequestType:     3,
	}
	ccr.diamMessage = ccr.AsBareDiameterMessage()
	ccr.diamMessage.NewAVP("Multiple-Services-Credit-Control", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(446, avp.Mbit, 0, &diam.GroupedAVP{ // Used-Service-Unit
				AVP: []*diam.AVP{
					diam.NewAVP(420, avp.Mbit, 0, datatype.Unsigned32(17)),   // CC-Time
					diam.NewAVP(412, avp.Mbit, 0, datatype.Unsigned64(1341)), // CC-Input-Octets
					diam.NewAVP(414, avp.Mbit, 0, datatype.Unsigned64(3079)), // CC-Output-Octets
				},
			}),
			diam.NewAVP(432, avp.Mbit, 0, datatype.Unsigned32(99)),
		},
	})
	ccr.diamMessage.NewAVP("Multiple-Services-Credit-Control", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(446, avp.Mbit, 0, &diam.GroupedAVP{ // Used-Service-Unit
				AVP: []*diam.AVP{
					diam.NewAVP(452, avp.Mbit, 0, datatype.Enumerated(0)),     // Tariff-Change-Usage
					diam.NewAVP(420, avp.Mbit, 0, datatype.Unsigned32(20)),    // CC-Time
					diam.NewAVP(412, avp.Mbit, 0, datatype.Unsigned64(8046)),  // CC-Input-Octets
					diam.NewAVP(414, avp.Mbit, 0, datatype.Unsigned64(46193)), // CC-Output-Octets
				},
			}),
			diam.NewAVP(432, avp.Mbit, 0, datatype.Unsigned32(1)),
		},
	})
	ccr.diamMessage.NewAVP("FramedIPAddress", avp.Mbit, 0, datatype.OctetString("0AE40041"))
	cfgFlds := make([]*config.CfgCdrField, 0)
	eSMGEv := sessionmanager.SMGenericEvent{"EventName": "DIAMETER_CCR"}
	if rSMGEv, err := ccr.AsSMGenericEvent(cfgFlds); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSMGEv, rSMGEv) {
		t.Errorf("Expecting: %+v, received: %+v", eSMGEv, rSMGEv)
	}
	cfgFlds = []*config.CfgCdrField{
		&config.CfgCdrField{
			Tag:         "LastUsed",
			FieldFilter: utils.ParseRSRFieldsMustCompile("~Multiple-Services-Credit-Control>Used-Service-Unit>CC-Input-Octets:s/^(.*)$/test/(test);Multiple-Services-Credit-Control>Rating-Group(1)", utils.INFIELD_SEP),
			FieldId:     "LastUsed",
			Type:        "*handler",
			HandlerId:   "*sum",
			Value:       utils.ParseRSRFieldsMustCompile("Multiple-Services-Credit-Control>Used-Service-Unit>CC-Input-Octets;^|;Multiple-Services-Credit-Control>Used-Service-Unit>CC-Output-Octets", utils.INFIELD_SEP),
			Mandatory:   true,
		},
	}
	eSMGEv = sessionmanager.SMGenericEvent{"EventName": "DIAMETER_CCR", "LastUsed": "54239"}
	if rSMGEv, err := ccr.AsSMGenericEvent(cfgFlds); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSMGEv, rSMGEv) {
		t.Errorf("Expecting: %+v, received: %+v", eSMGEv, rSMGEv)
	}
	cfgFlds = []*config.CfgCdrField{
		&config.CfgCdrField{
			Tag:         "LastUsed",
			FieldFilter: utils.ParseRSRFieldsMustCompile("~Multiple-Services-Credit-Control>Used-Service-Unit>CC-Input-Octets:s/^(.*)$/test/(test);Multiple-Services-Credit-Control>Rating-Group(99)", utils.INFIELD_SEP),
			FieldId:     "LastUsed",
			Type:        "*handler",
			HandlerId:   "*sum",
			Value:       utils.ParseRSRFieldsMustCompile("Multiple-Services-Credit-Control>Used-Service-Unit>CC-Input-Octets;^|;Multiple-Services-Credit-Control>Used-Service-Unit>CC-Output-Octets", utils.INFIELD_SEP),
			Mandatory:   true,
		},
	}
	eSMGEv = sessionmanager.SMGenericEvent{"EventName": "DIAMETER_CCR", "LastUsed": "4420"}
	if rSMGEv, err := ccr.AsSMGenericEvent(cfgFlds); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSMGEv, rSMGEv) {
		t.Errorf("Expecting: %+v, received: %+v", eSMGEv, rSMGEv)
	}
}
