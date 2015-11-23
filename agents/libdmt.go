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
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/fiorix/go-diameter/diam"
	"github.com/fiorix/go-diameter/diam/avp"
	"github.com/fiorix/go-diameter/diam/datatype"
	"github.com/fiorix/go-diameter/diam/dict"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func loadDictionaries(dictsDir, componentId string) error {
	fi, err := os.Stat(dictsDir)
	if err != nil {
		if strings.HasSuffix(err.Error(), "no such file or directory") {
			return fmt.Errorf("<DiameterAgent> Invalid dictionaries folder: <%s>", dictsDir)
		}
		return err
	} else if !fi.IsDir() { // If config dir defined, needs to exist
		return fmt.Errorf("<DiameterAgent> Path: <%s> is not a directory", dictsDir)
	}
	return filepath.Walk(dictsDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}
		cfgFiles, err := filepath.Glob(filepath.Join(path, "*.xml")) // Only consider .xml files
		if err != nil {
			return err
		}
		if cfgFiles == nil { // No need of processing further since there are no dictionary files in the folder
			return nil
		}
		for _, filePath := range cfgFiles {
			utils.Logger.Info(fmt.Sprintf("<%s> Loading dictionary out of file %s", componentId, filePath))
			if err := dict.Default.LoadFile(filePath); err != nil {
				return err
			}
		}
		return nil
	})

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

func usageFromCCR(reqType, reqNr, ccTime int, debitIterval time.Duration) time.Duration {
	dISecs := debitIterval.Seconds()
	if reqType == 3 {
		reqNr -= 1 // decrease request number to reach the real number
	}
	ccTime += int(dISecs) * reqNr
	return time.Duration(ccTime) * time.Second
}

// Utility function to convert from StoredCdr to CCR struct
func storedCdrToCCR(cdr *engine.StoredCdr, originHost, originRealm string, vendorId int, productName string,
	firmwareRev int, debitInterval time.Duration, callEnded bool) *CCR {
	sid := "session;" + strconv.Itoa(int(rand.Uint32()))
	reqType, reqNr, ccTime := disectUsageForCCR(cdr.Usage, debitInterval, callEnded)
	ccr := &CCR{SessionId: sid, OriginHost: originHost, OriginRealm: originRealm, DestinationHost: originHost, DestinationRealm: originRealm,
		AuthApplicationId: 4, ServiceContextId: cdr.ExtraFields["Service-Context-Id"], CCRequestType: reqType, CCRequestNumber: reqNr, EventTimestamp: cdr.AnswerTime,
		ServiceIdentifier: 0}
	ccr.SubscriptionId = make([]struct {
		SubscriptionIdType int    `avp:"Subscription-Id-Type"`
		SubscriptionIdData string `avp:"Subscription-Id-Data"`
	}, 1)
	ccr.SubscriptionId[0].SubscriptionIdType = 0
	ccr.SubscriptionId[0].SubscriptionIdData = cdr.Account
	ccr.RequestedServiceUnit.CCTime = ccTime
	ccr.ServiceInformation.INInformation.CallingPartyAddress = cdr.Account
	ccr.ServiceInformation.INInformation.CalledPartyAddress = cdr.Destination
	ccr.ServiceInformation.INInformation.RealCalledNumber = cdr.Destination
	ccr.ServiceInformation.INInformation.ChargeFlowType = 0
	ccr.ServiceInformation.INInformation.CallingVlrNumber = cdr.ExtraFields["Calling-Vlr-Number"]
	ccr.ServiceInformation.INInformation.CallingCellIDOrSAI = cdr.ExtraFields["Calling-CellID-Or-SAI"]
	ccr.ServiceInformation.INInformation.BearerCapability = cdr.ExtraFields["Bearer-Capability"]
	ccr.ServiceInformation.INInformation.CallReferenceNumber = cdr.CgrId
	ccr.ServiceInformation.INInformation.TimeZone = 0
	ccr.ServiceInformation.INInformation.SSPTime = cdr.ExtraFields["SSP-Time"]
	return ccr
}

// CallControl Request
type CCR struct {
	SessionId         string    `avp:"Session-Id"`
	OriginHost        string    `avp:"Origin-Host"`
	OriginRealm       string    `avp:"Origin-Realm"`
	DestinationHost   string    `avp:"Destination-Host"`
	DestinationRealm  string    `avp:"Destination-Realm"`
	AuthApplicationId int       `avp:"Auth-Application-Id"`
	ServiceContextId  string    `avp:"Service-Context-Id"`
	CCRequestType     int       `avp:"CC-Request-Type"`
	CCRequestNumber   int       `avp:"CC-Request-Number"`
	EventTimestamp    time.Time `avp:"Event-Timestamp"`
	ServiceIdentifier int       `avp:"Service-Identifier"`
	SubscriptionId    []struct {
		SubscriptionIdType int    `avp:"Subscription-Id-Type"`
		SubscriptionIdData string `avp:"Subscription-Id-Data"`
	} `avp:"Subscription-Id"`
	RequestedServiceUnit struct {
		CCTime int `avp:"CC-Time"`
	} `avp:"Requested-Service-Unit"`
	ServiceInformation struct {
		INInformation struct {
			CallingPartyAddress string `avp:"Calling-Party-Address"`
			CalledPartyAddress  string `avp:"Called-Party-Address"`
			RealCalledNumber    string `avp:"Real-Called-Number"`
			ChargeFlowType      int    `avp:"Charge-Flow-Type"`
			CallingVlrNumber    string `avp:"Calling-Vlr-Number"`
			CallingCellIDOrSAI  string `avp:"Calling-CellID-Or-SAI"`
			BearerCapability    string `avp:"Bearer-Capability"`
			CallReferenceNumber string `avp:"Call-Reference-Number"`
			MSCAddress          string `avp:"MSC-Address"`
			TimeZone            int    `avp:"Time-Zone"`
			CalledPartyNP       string `avp:"Called-Party-NP"`
			SSPTime             string `avp:"SSP-Time"`
		} `avp:"IN-Information"`
	} `avp:"Service-Information"`
	diamMessage *diam.Message // Used to parse fields with CGR templates
}

// ToDo: do it with reflect in the future
func (self *CCR) AsDiameterMessage() (*diam.Message, error) {
	m := diam.NewRequest(diam.CreditControl, 4, nil)
	if _, err := m.NewAVP("Session-Id", avp.Mbit, 0, datatype.UTF8String(self.SessionId)); err != nil {
		return nil, err
	}
	if _, err := m.NewAVP("Origin-Host", avp.Mbit, 0, datatype.DiameterIdentity(self.OriginHost)); err != nil {
		return nil, err
	}
	if _, err := m.NewAVP("Origin-Realm", avp.Mbit, 0, datatype.DiameterIdentity(self.OriginRealm)); err != nil {
		return nil, err
	}
	if _, err := m.NewAVP("Destination-Host", avp.Mbit, 0, datatype.DiameterIdentity(self.DestinationHost)); err != nil {
		return nil, err
	}
	if _, err := m.NewAVP("Destination-Realm", avp.Mbit, 0, datatype.DiameterIdentity(self.DestinationRealm)); err != nil {
		return nil, err
	}
	if _, err := m.NewAVP("Auth-Application-Id", avp.Mbit, 0, datatype.Unsigned32(self.AuthApplicationId)); err != nil {
		return nil, err
	}
	if _, err := m.NewAVP("Service-Context-Id", avp.Mbit, 0, datatype.UTF8String(self.ServiceContextId)); err != nil {
		return nil, err
	}
	if _, err := m.NewAVP("CC-Request-Type", avp.Mbit, 0, datatype.Enumerated(self.CCRequestType)); err != nil {
		return nil, err
	}
	if _, err := m.NewAVP("CC-Request-Number", avp.Mbit, 0, datatype.Enumerated(self.CCRequestNumber)); err != nil {
		return nil, err
	}
	if _, err := m.NewAVP("Event-Timestamp", avp.Mbit, 0, datatype.Time(self.EventTimestamp)); err != nil {
		return nil, err
	}

	subscriptionIdType, err := m.Dictionary().FindAVP(m.Header.ApplicationID, "Subscription-Id-Type")
	if err != nil {
		return nil, err
	}
	subscriptionIdData, err := m.Dictionary().FindAVP(m.Header.ApplicationID, "Subscription-Id-Data")
	if err != nil {
		return nil, err
	}
	for _, subscriptionId := range self.SubscriptionId {
		if _, err := m.NewAVP("Subscription-Id", avp.Mbit, 0, &diam.GroupedAVP{
			AVP: []*diam.AVP{
				diam.NewAVP(subscriptionIdType.Code, avp.Mbit, 0, datatype.Enumerated(subscriptionId.SubscriptionIdType)),
				diam.NewAVP(subscriptionIdData.Code, avp.Mbit, 0, datatype.UTF8String(subscriptionId.SubscriptionIdData)),
			}}); err != nil {
			return nil, err
		}
	}
	if _, err := m.NewAVP("Service-Identifier", avp.Mbit, 0, datatype.Unsigned32(self.ServiceIdentifier)); err != nil {
		return nil, err
	}
	ccTimeAvp, err := m.Dictionary().FindAVP(m.Header.ApplicationID, "CC-Time")
	if err != nil {
		return nil, err
	}
	if _, err := m.NewAVP("Requested-Service-Unit", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(ccTimeAvp.Code, avp.Mbit, 0, datatype.Unsigned32(self.RequestedServiceUnit.CCTime))}}); err != nil {
		return nil, err
	}
	inInformation, err := m.Dictionary().FindAVP(m.Header.ApplicationID, "IN-Information")
	if err != nil {
		return nil, err
	}
	callingPartyAddress, err := m.Dictionary().FindAVP(m.Header.ApplicationID, "Calling-Party-Address")
	if err != nil {
		return nil, err
	}
	calledPartyAddress, err := m.Dictionary().FindAVP(m.Header.ApplicationID, "Called-Party-Address")
	if err != nil {
		return nil, err
	}
	realCalledNumber, err := m.Dictionary().FindAVP(m.Header.ApplicationID, "Real-Called-Number")
	if err != nil {
		return nil, err
	}
	chargeFlowType, err := m.Dictionary().FindAVP(m.Header.ApplicationID, "Charge-Flow-Type")
	if err != nil {
		return nil, err
	}
	callingVlrNumber, err := m.Dictionary().FindAVP(m.Header.ApplicationID, "Calling-Vlr-Number")
	if err != nil {
		return nil, err
	}
	callingCellIdOrSai, err := m.Dictionary().FindAVP(m.Header.ApplicationID, "Calling-CellID-Or-SAI")
	if err != nil {
		return nil, err
	}
	bearerCapability, err := m.Dictionary().FindAVP(m.Header.ApplicationID, "Bearer-Capability")
	if err != nil {
		return nil, err
	}
	callReferenceNumber, err := m.Dictionary().FindAVP(m.Header.ApplicationID, "Call-Reference-Number")
	if err != nil {
		return nil, err
	}
	mscAddress, err := m.Dictionary().FindAVP(m.Header.ApplicationID, "MSC-Address")
	if err != nil {
		return nil, err
	}
	timeZone, err := m.Dictionary().FindAVP(m.Header.ApplicationID, "Time-Zone")
	if err != nil {
		return nil, err
	}
	calledPartyNP, err := m.Dictionary().FindAVP(m.Header.ApplicationID, "Called-Party-NP")
	if err != nil {
		return nil, err
	}
	sspTime, err := m.Dictionary().FindAVP(m.Header.ApplicationID, "SSP-Time")
	if err != nil {
		return nil, err
	}
	if _, err := m.NewAVP("Service-Information", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(inInformation.Code, avp.Mbit, 0, &diam.GroupedAVP{
				AVP: []*diam.AVP{
					diam.NewAVP(callingPartyAddress.Code, avp.Mbit, 0, datatype.UTF8String(self.ServiceInformation.INInformation.CallingPartyAddress)),
					diam.NewAVP(calledPartyAddress.Code, avp.Mbit, 0, datatype.UTF8String(self.ServiceInformation.INInformation.CalledPartyAddress)),
					diam.NewAVP(realCalledNumber.Code, avp.Mbit, 0, datatype.UTF8String(self.ServiceInformation.INInformation.RealCalledNumber)),
					diam.NewAVP(chargeFlowType.Code, avp.Mbit, 0, datatype.Unsigned32(self.ServiceInformation.INInformation.ChargeFlowType)),
					diam.NewAVP(callingVlrNumber.Code, avp.Mbit, 0, datatype.UTF8String(self.ServiceInformation.INInformation.CallingVlrNumber)),
					diam.NewAVP(callingCellIdOrSai.Code, avp.Mbit, 0, datatype.UTF8String(self.ServiceInformation.INInformation.CallingCellIDOrSAI)),
					diam.NewAVP(bearerCapability.Code, avp.Mbit, 0, datatype.UTF8String(self.ServiceInformation.INInformation.BearerCapability)),
					diam.NewAVP(callReferenceNumber.Code, avp.Mbit, 0, datatype.UTF8String(self.ServiceInformation.INInformation.CallReferenceNumber)),
					diam.NewAVP(mscAddress.Code, avp.Mbit, 0, datatype.UTF8String(self.ServiceInformation.INInformation.MSCAddress)),
					diam.NewAVP(timeZone.Code, avp.Mbit, 0, datatype.UTF8String(self.ServiceInformation.INInformation.TimeZone)),
					diam.NewAVP(calledPartyNP.Code, avp.Mbit, 0, datatype.UTF8String(self.ServiceInformation.INInformation.CalledPartyNP)),
					diam.NewAVP(sspTime.Code, avp.Mbit, 0, datatype.UTF8String(self.ServiceInformation.INInformation.SSPTime)),
				},
			}),
		}}); err != nil {
		return nil, err
	}
	return m, nil
}

func NewCCAFromCCR(ccr *CCR) *CCA {
	return &CCA{SessionId: ccr.SessionId, AuthApplicationId: ccr.AuthApplicationId, CCRequestType: ccr.CCRequestType, CCRequestNumber: ccr.CCRequestNumber,
		diamMessage: diam.NewMessage(ccr.diamMessage.Header.CommandCode, ccr.diamMessage.Header.CommandFlags&^diam.RequestFlag, ccr.diamMessage.Header.ApplicationID,
			ccr.diamMessage.Header.HopByHopID, ccr.diamMessage.Header.EndToEndID, ccr.diamMessage.Dictionary()),
	}
}

// Call Control Answer
type CCA struct {
	SessionId          string `avp:"Session-Id"`
	OriginHost         string `avp:"Origin-Host"`
	OriginRealm        string `avp:"Origin-Realm"`
	AuthApplicationId  int    `avp:"Auth-Application-Id"`
	CCRequestType      int    `avp:"CC-Request-Type"`
	CCRequestNumber    int    `avp:"CC-Request-Number"`
	ResultCode         int    `avp:"Result-Code"`
	GrantedServiceUnit struct {
		CCTime int `avp:"CC-Time"`
	} `avp:"Granted-Service-Unit"`
	diamMessage *diam.Message
}

// Converts itself into DiameterMessage
func (self *CCA) AsDiameterMessage() (*diam.Message, error) {
	if _, err := self.diamMessage.NewAVP("Session-Id", avp.Mbit, 0, datatype.UTF8String(self.SessionId)); err != nil {
		return nil, err
	}
	if _, err := self.diamMessage.NewAVP("Origin-Host", avp.Mbit, 0, datatype.DiameterIdentity(self.OriginHost)); err != nil {
		return nil, err
	}
	if _, err := self.diamMessage.NewAVP("Origin-Realm", avp.Mbit, 0, datatype.DiameterIdentity(self.OriginRealm)); err != nil {
		return nil, err
	}
	if _, err := self.diamMessage.NewAVP("Auth-Application-Id", avp.Mbit, 0, datatype.Unsigned32(self.AuthApplicationId)); err != nil {
		return nil, err
	}
	if _, err := self.diamMessage.NewAVP("CC-Request-Type", avp.Mbit, 0, datatype.Enumerated(self.CCRequestType)); err != nil {
		return nil, err
	}
	if _, err := self.diamMessage.NewAVP("CC-Request-Number", avp.Mbit, 0, datatype.Enumerated(self.CCRequestNumber)); err != nil {
		return nil, err
	}
	if _, err := self.diamMessage.NewAVP(avp.ResultCode, avp.Mbit, 0, datatype.Unsigned32(self.ResultCode)); err != nil {
		return nil, err
	}
	ccTimeAvp, err := self.diamMessage.Dictionary().FindAVP(self.diamMessage.Header.ApplicationID, "CC-Time")
	if err != nil {
		return nil, err
	}
	if _, err := self.diamMessage.NewAVP("Granted-Service-Unit", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(ccTimeAvp.Code, avp.Mbit, 0, datatype.Unsigned32(self.GrantedServiceUnit.CCTime))}}); err != nil {
		return nil, err
	}
	return self.diamMessage, nil

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
