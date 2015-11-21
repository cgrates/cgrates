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

/*
Build various type of packets here
*/

import (
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/fiorix/go-diameter/diam"
	"github.com/fiorix/go-diameter/diam/avp"
	"github.com/fiorix/go-diameter/diam/datatype"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Returns reqType, requestNr and ccTime in seconds
func disectUsageForCCR(usage time.Duration, debitInterval time.Duration, callEnded bool) (int, int, int) {
	usageSecs := usage.Seconds()
	debitIntervalSecs := debitInterval.Seconds()
	reqType := 1
	if usage > 0 {
		reqType = 2
	}
	if callEnded {
		reqType = 3
	}
	reqNr := int(usageSecs / debitIntervalSecs)
	if callEnded {
		reqNr += 1
	}
	ccTime := debitInterval.Seconds()
	if callEnded {
		ccTime = math.Mod(usageSecs, debitIntervalSecs)
	}
	return reqType, reqNr, int(ccTime)
}

func getUsageFromCCR(reqType, reqNr, ccTime int, debitIterval time.Duration) time.Duration {
	dISecs := debitIterval.Seconds()
	if reqType == 3 {
		reqNr -= 1 // decrease request number to reach the real number
	}
	ccTime += int(dISecs) * reqNr
	return time.Duration(ccTime) * time.Second
}

func storedCdrToCCR(cdr *engine.StoredCdr, originHost, originRealm string, vendorId int, productName string, firmwareRev int, debitInterval time.Duration, callEnded bool) *diam.Message {
	sid := "session;" + strconv.Itoa(int(rand.Uint32()))
	reqType, reqNr, ccTime := disectUsageForCCR(cdr.Usage, debitInterval, callEnded)
	m := diam.NewRequest(272, 4, nil)
	m.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(sid))
	m.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity(originHost))
	m.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity(originRealm))
	m.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.DiameterIdentity(originHost))
	m.NewAVP(avp.DestinationRealm, avp.Mbit, 0, datatype.DiameterIdentity(originRealm))
	m.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	m.NewAVP(avp.ServiceContextID, avp.Mbit, 0, datatype.UTF8String("voice@huawei.com"))
	m.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(reqType))
	m.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Enumerated(reqNr))
	m.NewAVP(avp.EventTimestamp, avp.Mbit, 0, datatype.Time(cdr.AnswerTime))
	m.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(0)),
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String(cdr.Account)),
		}})
	m.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(1)),
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("20921006232651")),
		}})
	m.NewAVP(avp.ServiceIdentifier, avp.Mbit, 0, datatype.Unsigned32(0))
	m.NewAVP(avp.RequestedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.CCTime, avp.Mbit, 0, datatype.Unsigned32(ccTime))}})
	/*
		m.NewAVP(avp.ServiceInformation, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(20300, avp.Mbit, 0, &diam.GroupedAVP{ // IN-Information
				AVP: []*diam.AVP{
					diam.NewAVP(avp.CallingPartyAddress, avp.Mbit, 0, datatype.UTF8String(cdr.Account)),
					diam.NewAVP(avp.CalledPartyAddress, avp.Mbit, 0, datatype.UTF8String(cdr.Destination)),
					diam.NewAVP(20327, avp.Mbit, 0, datatype.UTF8String(cdr.Destination)),  // Real-Called-Number
					diam.NewAVP(20339, avp.Mbit, 0, datatype.Unsigned32(0)),                // Charge-Flow-Type
					diam.NewAVP(20302, avp.Mbit, 0, datatype.UTF8String("33657954968")),    // Calling-Vlr-Number
					diam.NewAVP(20303, avp.Mbit, 0, datatype.UTF8String("31901485301525")), // Calling-CellID-Or-SAI
					diam.NewAVP(avp.BearerCapability, avp.Mbit, 0, datatype.UTF8String("31901485301525")),
					diam.NewAVP(20321, avp.Mbit, 0, datatype.UTF8String("31901485301525")), // Call-Reference-Number
					diam.NewAVP(avp.MSCAddress, avp.Mbit, 0, datatype.UTF8String("")),
					diam.NewAVP(20324, avp.Mbit, 0, datatype.UTF8String("0")),              // Time-Zone
					diam.NewAVP(20385, avp.Mbit, 0, datatype.UTF8String("")),               // Called-Party-NP
					diam.NewAVP(20386, avp.Mbit, 0, datatype.UTF8String("20091020120101")), // SSP-Time
				},
			}),
		}})
	*/
	return m
}

// Extracts the value out of a specific field in diameter message, able to go into multiple layers in the form of field1>field2>field3
func dmtMessageFieldValue(dm *diam.Message, fieldId string) string {
	//fieldNameLevels := strings.Split(fieldId, ">")
	return ""

}

// Converts Diameter CCR message into StoredCdr based on field template
func ccrToStoredCdr(ccr *diam.Message, tpl []*config.CfgCdrField) (*engine.StoredCdr, error) {
	return nil, nil
}
