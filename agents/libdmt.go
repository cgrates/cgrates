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
	"github.com/cgrates/cgrates/sessionmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/fiorix/go-diameter/diam"
	"github.com/fiorix/go-diameter/diam/avp"
	"github.com/fiorix/go-diameter/diam/datatype"
	"github.com/fiorix/go-diameter/diam/dict"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const (
	META_CCR_USAGE          = "*ccr_usage"
	META_CCR_SMG_EVENT_NAME = "*ccr_smg_event_name"
	DIAMETER_CCR            = "DIAMETER_CCR"
)

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
func disectUsageForCCR(usage time.Duration, debitInterval time.Duration, callEnded bool) (reqType, reqNr, reqCCTime, usedCCTime int) {
	usageSecs := usage.Seconds()
	debitIntervalSecs := debitInterval.Seconds()
	reqType = 1
	if usage > 0 {
		reqType = 2
	}
	if callEnded {
		reqType = 3
	}
	reqNr = int(usageSecs / debitIntervalSecs)
	if callEnded {
		reqNr += 1
	}
	ccTimeFloat := debitInterval.Seconds()
	if callEnded {
		ccTimeFloat = math.Mod(usageSecs, debitIntervalSecs)
	}
	if reqType == 1 { // Initial does not have usedCCTime
		reqCCTime = int(ccTimeFloat)
	} else if reqType == 2 {
		reqCCTime = int(ccTimeFloat)
		usedCCTime = int(math.Mod(usageSecs, debitIntervalSecs))
	} else if reqType == 3 {
		usedCCTime = int(ccTimeFloat) // Termination does not have requestCCTime
	}
	return
}

func usageFromCCR(reqType, reqNr, reqCCTime, usedCCTime int, debitIterval time.Duration) time.Duration {
	dISecs := debitIterval.Seconds()
	var ccTime int
	if reqType == 3 {
		reqNr -= 1 // decrease request number to reach the real number
		ccTime = usedCCTime + (int(dISecs) * reqNr)
	} else {
		ccTime = int(dISecs)
	}
	return time.Duration(ccTime) * time.Second
}

// Utility function to convert from StoredCdr to CCR struct
func storedCdrToCCR(cdr *engine.CDR, originHost, originRealm string, vendorId int, productName string,
	firmwareRev int, debitInterval time.Duration, callEnded bool) *CCR {
	//sid := "session;" + strconv.Itoa(int(rand.Uint32()))
	reqType, reqNr, reqCCTime, usedCCTime := disectUsageForCCR(cdr.Usage, debitInterval, callEnded)
	ccr := &CCR{SessionId: cdr.CGRID, OriginHost: originHost, OriginRealm: originRealm, DestinationHost: originHost, DestinationRealm: originRealm,
		AuthApplicationId: 4, ServiceContextId: cdr.ExtraFields["Service-Context-Id"], CCRequestType: reqType, CCRequestNumber: reqNr, EventTimestamp: cdr.AnswerTime,
		ServiceIdentifier: 0}
	ccr.SubscriptionId = make([]struct {
		SubscriptionIdType int    `avp:"Subscription-Id-Type"`
		SubscriptionIdData string `avp:"Subscription-Id-Data"`
	}, 1)
	ccr.SubscriptionId[0].SubscriptionIdType = 0
	ccr.SubscriptionId[0].SubscriptionIdData = cdr.Account
	ccr.RequestedServiceUnit.CCTime = reqCCTime
	ccr.UsedServiceUnit.CCTime = usedCCTime
	ccr.ServiceInformation.INInformation.CallingPartyAddress = cdr.Account
	ccr.ServiceInformation.INInformation.CalledPartyAddress = cdr.Destination
	ccr.ServiceInformation.INInformation.RealCalledNumber = cdr.Destination
	ccr.ServiceInformation.INInformation.ChargeFlowType = 0
	ccr.ServiceInformation.INInformation.CallingVlrNumber = cdr.ExtraFields["Calling-Vlr-Number"]
	ccr.ServiceInformation.INInformation.CallingCellIDOrSAI = cdr.ExtraFields["Calling-CellID-Or-SAI"]
	ccr.ServiceInformation.INInformation.BearerCapability = cdr.ExtraFields["Bearer-Capability"]
	ccr.ServiceInformation.INInformation.CallReferenceNumber = cdr.CGRID
	ccr.ServiceInformation.INInformation.TimeZone = 0
	ccr.ServiceInformation.INInformation.SSPTime = cdr.ExtraFields["SSP-Time"]
	return ccr
}

// debitInterval is the configured debitInterval, in sync with the diameter client one
func NewCCRFromDiameterMessage(m *diam.Message, debitInterval time.Duration) (*CCR, error) {
	var ccr CCR
	if err := m.Unmarshal(&ccr); err != nil {
		return nil, err
	}
	ccr.diamMessage = m
	ccr.debitInterval = debitInterval
	return &ccr, nil
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
	SubscriptionId    []struct {
		SubscriptionIdType int    `avp:"Subscription-Id-Type"`
		SubscriptionIdData string `avp:"Subscription-Id-Data"`
	} `avp:"Subscription-Id"`
	ServiceIdentifier    int `avp:"Service-Identifier"`
	RequestedServiceUnit struct {
		CCTime int `avp:"CC-Time"`
	} `avp:"Requested-Service-Unit"`
	UsedServiceUnit struct {
		CCTime int `avp:"CC-Time"`
	} `avp:"Used-Service-Unit"`
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
	diamMessage   *diam.Message // Used to parse fields with CGR templates
	debitInterval time.Duration // Configured debit interval
}

// Used when sending from client to agent
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
	for _, subscriptionId := range self.SubscriptionId {
		if _, err := m.NewAVP("Subscription-Id", avp.Mbit, 0, &diam.GroupedAVP{
			AVP: []*diam.AVP{
				diam.NewAVP(450, avp.Mbit, 0, datatype.Enumerated(subscriptionId.SubscriptionIdType)), // Subscription-Id-Type
				diam.NewAVP(444, avp.Mbit, 0, datatype.UTF8String(subscriptionId.SubscriptionIdData)), // Subscription-Id-Data
			}}); err != nil {
			return nil, err
		}
	}
	if _, err := m.NewAVP("Service-Identifier", avp.Mbit, 0, datatype.Unsigned32(self.ServiceIdentifier)); err != nil {
		return nil, err
	}
	if _, err := m.NewAVP("Requested-Service-Unit", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(420, avp.Mbit, 0, datatype.Unsigned32(self.RequestedServiceUnit.CCTime))}}); err != nil { // CC-Time
		return nil, err
	}
	if _, err := m.NewAVP("Used-Service-Unit", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(420, avp.Mbit, 0, datatype.Unsigned32(self.UsedServiceUnit.CCTime))}}); err != nil { // CC-Time
		return nil, err
	}
	if _, err := m.NewAVP(873, avp.Mbit, 10415, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(20300, avp.Mbit, 2011, &diam.GroupedAVP{ // IN-Information
				AVP: []*diam.AVP{
					diam.NewAVP(831, avp.Mbit, 10415, datatype.UTF8String(self.ServiceInformation.INInformation.CallingPartyAddress)),  // Calling-Party-Address
					diam.NewAVP(832, avp.Mbit, 10415, datatype.UTF8String(self.ServiceInformation.INInformation.CalledPartyAddress)),   // Called-Party-Address
					diam.NewAVP(20327, avp.Mbit, 2011, datatype.UTF8String(self.ServiceInformation.INInformation.RealCalledNumber)),    // Real-Called-Number
					diam.NewAVP(20339, avp.Mbit, 2011, datatype.Unsigned32(self.ServiceInformation.INInformation.ChargeFlowType)),      // Charge-Flow-Type
					diam.NewAVP(20302, avp.Mbit, 2011, datatype.UTF8String(self.ServiceInformation.INInformation.CallingVlrNumber)),    // Calling-Vlr-Number
					diam.NewAVP(20303, avp.Mbit, 2011, datatype.UTF8String(self.ServiceInformation.INInformation.CallingCellIDOrSAI)),  // Calling-CellID-Or-SAI
					diam.NewAVP(20313, avp.Mbit, 2011, datatype.UTF8String(self.ServiceInformation.INInformation.BearerCapability)),    // Bearer-Capability
					diam.NewAVP(20321, avp.Mbit, 2011, datatype.UTF8String(self.ServiceInformation.INInformation.CallReferenceNumber)), // Call-Reference-Number
					diam.NewAVP(20322, avp.Mbit, 2011, datatype.UTF8String(self.ServiceInformation.INInformation.MSCAddress)),          // MSC-Address
					diam.NewAVP(20324, avp.Mbit, 2011, datatype.Unsigned32(self.ServiceInformation.INInformation.TimeZone)),            // Time-Zone
					diam.NewAVP(20385, avp.Mbit, 2011, datatype.UTF8String(self.ServiceInformation.INInformation.CalledPartyNP)),       // Called-Party-NP
					diam.NewAVP(20386, avp.Mbit, 2011, datatype.UTF8String(self.ServiceInformation.INInformation.SSPTime)),             // SSP-Time
				},
			}),
		}}); err != nil {
		return nil, err
	}
	return m, nil
}

// Not the cleanest but most efficient way to retrieve a string from AVP since there are string methods on all datatypes
// and the output is always in teh form "DataType{real_string}Padding:x"
func avpValAsString(a *diam.AVP) string {
	dataVal := a.Data.String()
	startIdx := strings.Index(dataVal, "{")
	endIdx := strings.Index(dataVal, "}")
	if startIdx == 0 || endIdx == 0 {
		return ""
	}
	return dataVal[startIdx+1 : endIdx]
}

// Handler for meta functions
func (self *CCR) metaHandler(tag, arg string) (string, error) {
	switch tag {
	case META_CCR_USAGE:
		usage := usageFromCCR(self.CCRequestType, self.CCRequestNumber, self.RequestedServiceUnit.CCTime, self.UsedServiceUnit.CCTime, self.debitInterval)
		return strconv.FormatFloat(usage.Seconds(), 'f', -1, 64), nil
	}
	return "", nil
}

func (self *CCR) avpsWithPath(rsrFld *utils.RSRField) ([]*diam.AVP, error) {
	hierarchyPath := strings.Split(rsrFld.Id, utils.HIERARCHY_SEP)
	hpIf := make([]interface{}, len(hierarchyPath))
	for i, val := range hierarchyPath {
		hpIf[i] = val
	}
	return self.diamMessage.FindAVPsWithPath(hpIf, dict.UndefinedVendorID)
}

// Follows the implementation in the StorCdr
func (self *CCR) passesFieldFilter(fieldFilter *utils.RSRField) (bool, int) {
	if fieldFilter == nil {
		return true, 0
	}
	avps, err := self.avpsWithPath(fieldFilter)
	if err != nil {
		return false, 0
	} else if len(avps) == 0 {
		return true, 0
	}
	for avpIdx, avpVal := range avps {
		if fieldFilter.FilterPasses(avpValAsString(avpVal)) {
			return true, avpIdx
		}
	}
	return false, 0
}

func (self *CCR) eventFieldValue(fldTpl utils.RSRFields, avpIdx int) string {
	var outVal string
	for _, rsrTpl := range fldTpl {
		if rsrTpl.IsStatic() {
			outVal += rsrTpl.ParseValue("")
		} else {
			matchingAvps, err := self.avpsWithPath(rsrTpl)
			if err != nil || len(matchingAvps) == 0 {
				utils.Logger.Warning(fmt.Sprintf("<Diameter> Cannot find AVP for field template with id: %s, ignoring.", rsrTpl.Id))
				continue // Filter not matching
			}
			if len(matchingAvps) <= avpIdx {
				utils.Logger.Warning(fmt.Sprintf("<Diameter> Cannot retrieve AVP with index %d for field template with id: %s", avpIdx, rsrTpl.Id))
				continue // Not convertible, ignore
			}
			if matchingAvps[0].Data.Type() == diam.GroupedAVPType {
				utils.Logger.Warning(fmt.Sprintf("<Diameter> Value for field template with id: %s is matching a group AVP, ignoring.", rsrTpl.Id))
				continue // Not convertible, ignore
			}
			outVal += avpValAsString(matchingAvps[avpIdx])
		}
	}
	return outVal
}

func (self *CCR) fieldOutVal(cfgFld *config.CfgCdrField) (fmtValOut string, err error) {
	var outVal string
	switch cfgFld.Type {
	case utils.META_FILLER:
		outVal = cfgFld.Value.Id()
		cfgFld.Padding = "right"
	case utils.META_CONSTANT:
		outVal = cfgFld.Value.Id()
	case utils.META_HANDLER:
		outVal, err = self.metaHandler(cfgFld.HandlerId, cfgFld.Layout)
		if err != nil {
			utils.Logger.Warning(fmt.Sprintf("<Diameter> Ignoring processing of metafunction: %s, error: %s", cfgFld.HandlerId, err.Error()))
		}
	case utils.META_COMPOSED:
		outVal = self.eventFieldValue(cfgFld.Value, 0)
	case utils.MetaGrouped: // GroupedAVP
		passAtIndex := -1
		matchedAllFilters := true
		for _, fldFilter := range cfgFld.FieldFilter {
			var pass bool
			if pass, passAtIndex = self.passesFieldFilter(fldFilter); !pass {
				matchedAllFilters = false
				break
			}
		}
		if !matchedAllFilters {
			return "", nil // Not matching field filters, will have it empty
		}
		if passAtIndex == -1 {
			passAtIndex = 0 // No filter
		}
		outVal = self.eventFieldValue(cfgFld.Value, passAtIndex)
	}
	if fmtValOut, err = utils.FmtFieldWidth(outVal, cfgFld.Width, cfgFld.Strip, cfgFld.Padding, cfgFld.Mandatory); err != nil {
		utils.Logger.Warning(fmt.Sprintf("<Diameter> Error when processing field template with tag: %s, error: %s", cfgFld.Tag, err.Error()))
		return "", err
	}
	return fmtValOut, nil
}

// Extracts data out of CCR into a SMGenericEvent based on the configured template
func (self *CCR) AsSMGenericEvent(cfgFlds []*config.CfgCdrField) (sessionmanager.SMGenericEvent, error) {
	outMap := make(map[string]string) // work with it so we can append values to keys
	outMap[utils.EVENT_NAME] = DIAMETER_CCR
	for _, cfgFld := range cfgFlds {
		fmtOut, err := self.fieldOutVal(cfgFld)
		if err != nil {
			return nil, err
		}
		if _, hasKey := outMap[cfgFld.FieldId]; !hasKey {
			outMap[cfgFld.FieldId] = fmtOut
		} else { // If already there, postpend
			outMap[cfgFld.FieldId] += fmtOut
		}
	}
	return sessionmanager.SMGenericEvent(utils.ConvertMapValStrIf(outMap)), nil
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
