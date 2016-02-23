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
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net"
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
	META_CCR_USAGE       = "*ccr_usage"
	META_CCA_USAGE       = "*cca_usage"
	META_VALUE_EXPONENT  = "*value_exponent"
	DIAMETER_CCR         = "DIAMETER_CCR"
	DiameterRatingFailed = 5031
	CGRError             = "CGRError"
	CGRMaxUsage          = "CGRMaxUsage"
	CGRResultCode        = "CGRResultCode"
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
func metaHandler(m *diam.Message, tag, arg string, dur time.Duration) (string, error) {
	switch tag {
	case META_CCR_USAGE:
		var ok bool
		var reqType datatype.Enumerated
		var reqNr, reqUnit, usedUnit datatype.Unsigned32
		if ccReqTypeAvp, err := m.FindAVP("CC-Request-Type", 0); err != nil {
			return "", err
		} else if ccReqTypeAvp == nil {
			return "", errors.New("CC-Request-Type not found")
		} else if reqType, ok = ccReqTypeAvp.Data.(datatype.Enumerated); !ok {
			return "", fmt.Errorf("CC-Request-Type must be Enumerated and not %v", ccReqTypeAvp.Data.Type())
		}
		if ccReqNrAvp, err := m.FindAVP("CC-Request-Number", 0); err != nil {
			return "", err
		} else if ccReqNrAvp == nil {
			return "", errors.New("CC-Request-Number not found")
		} else if reqNr, ok = ccReqNrAvp.Data.(datatype.Unsigned32); !ok {
			return "", fmt.Errorf("CC-Request-Number must be Unsigned32 and not %v", ccReqNrAvp.Data.Type())
		}
		switch reqType {
		case datatype.Enumerated(1), datatype.Enumerated(2):
			if reqUnitAVPs, err := m.FindAVPsWithPath([]interface{}{"Requested-Service-Unit", "CC-Time"}, dict.UndefinedVendorID); err != nil {
				return "", err
			} else if len(reqUnitAVPs) == 0 {
				return "", errors.New("Requested-Service-Unit>CC-Time not found")
			} else if reqUnit, ok = reqUnitAVPs[0].Data.(datatype.Unsigned32); !ok {
				return "", fmt.Errorf("Requested-Service-Unit>CC-Time must be Unsigned32 and not %v", reqUnitAVPs[0].Data.Type())
			}
		case datatype.Enumerated(3), datatype.Enumerated(4):
			if usedUnitAVPs, err := m.FindAVPsWithPath([]interface{}{"Used-Service-Unit", "CC-Time"}, dict.UndefinedVendorID); err != nil {
				return "", err
			} else if len(usedUnitAVPs) != 0 {
				if usedUnit, ok = usedUnitAVPs[0].Data.(datatype.Unsigned32); !ok {
					return "", fmt.Errorf("Used-Service-Unit>CC-Time must be Unsigned32 and not %v", usedUnitAVPs[0].Data.Type())
				}
			}
		}
		usage := usageFromCCR(int(reqType), int(reqNr), int(reqUnit), int(usedUnit), dur)
		return strconv.FormatFloat(usage.Seconds(), 'f', -1, 64), nil
	}
	return "", nil
}

// metaValueExponent will multiply the float value with the exponent provided.
// Expects 2 arguments in template separated by |
func metaValueExponent(m *diam.Message, argsTpl utils.RSRFields, roundingDecimals int) (string, error) {
	valStr := composedFieldvalue(m, argsTpl, 0)
	handlerArgs := strings.Split(valStr, utils.HandlerArgSep)
	if len(handlerArgs) != 2 {
		return "", errors.New("Unexpected number of arguments")
	}
	val, err := strconv.ParseFloat(handlerArgs[0], 64)
	if err != nil {
		return "", err
	}
	exp, err := strconv.Atoi(handlerArgs[1])
	if err != nil {
		return "", err
	}
	res := val * math.Pow10(exp)
	return strconv.FormatFloat(utils.Round(res, roundingDecimals, utils.ROUNDING_MIDDLE), 'f', -1, 64), nil
}

// splitIntoInterface is used to split a string into []interface{} instead of []string
func splitIntoInterface(content, sep string) []interface{} {
	spltStr := strings.Split(content, sep)
	spltIf := make([]interface{}, len(spltStr))
	for i, val := range spltStr {
		spltIf[i] = val
	}
	return spltIf
}

// avpsWithPath is used to find AVPs by specifying RSRField as filter
func avpsWithPath(m *diam.Message, rsrFld *utils.RSRField) ([]*diam.AVP, error) {
	return m.FindAVPsWithPath(splitIntoInterface(rsrFld.Id, utils.HIERARCHY_SEP), dict.UndefinedVendorID)
}

// Follows the implementation in the StorCdr
func passesFieldFilter(m *diam.Message, fieldFilter *utils.RSRField) (bool, int) {
	if fieldFilter == nil {
		return true, 0
	}
	avps, err := avpsWithPath(m, fieldFilter)
	if err != nil {
		return false, 0
	} else if len(avps) == 0 {
		return false, 0 // No AVPs with field filter ID
	}
	for avpIdx, avpVal := range avps { // First match wins due to index
		if fieldFilter.FilterPasses(avpValAsString(avpVal)) {
			return true, avpIdx
		}
	}
	return false, 0
}

func composedFieldvalue(m *diam.Message, outTpl utils.RSRFields, avpIdx int) string {
	var outVal string
	for _, rsrTpl := range outTpl {
		if rsrTpl.IsStatic() {
			outVal += rsrTpl.ParseValue("")
		} else {
			matchingAvps, err := avpsWithPath(m, rsrTpl)
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
			outVal += rsrTpl.ParseValue(avpValAsString(matchingAvps[avpIdx]))
		}
	}
	return outVal
}

// Used to return the encoded value based on what AVP understands for it's type
func serializeAVPValueFromString(dictAVP *dict.AVP, valStr, timezone string) ([]byte, error) {
	switch dictAVP.Data.Type {
	case datatype.OctetStringType, datatype.DiameterIdentityType, datatype.DiameterURIType, datatype.IPFilterRuleType, datatype.QoSFilterRuleType, datatype.UTF8StringType:
		return []byte(valStr), nil
	case datatype.AddressType:
		return []byte(net.ParseIP(valStr)), nil
	case datatype.EnumeratedType, datatype.Integer32Type, datatype.Integer64Type, datatype.Unsigned32Type, datatype.Unsigned64Type:
		i, err := strconv.Atoi(valStr)
		if err != nil {
			return nil, err
		}
		return datatype.Enumerated(i).Serialize(), nil
	case datatype.Float32Type:
		f, err := strconv.ParseFloat(valStr, 32)
		if err != nil {
			return nil, err
		}
		return datatype.Float32(f).Serialize(), nil
	case datatype.Float64Type:
		f, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			return nil, err
		}
		return datatype.Float64(f).Serialize(), nil
	case datatype.GroupedType:
		return nil, errors.New("GroupedType not supported for serialization")
	case datatype.IPv4Type:
		return datatype.IPv4(net.ParseIP(valStr)).Serialize(), nil
	case datatype.TimeType:
		t, err := utils.ParseTimeDetectLayout(valStr, timezone)
		if err != nil {
			return nil, err
		}
		return datatype.Time(t).Serialize(), nil
	default:
		return nil, fmt.Errorf("Unsupported type for serialization: %v", dictAVP.Data.Type)
	}
}

var ErrFilterNotPassing = errors.New("Filter not passing")

func fieldOutVal(m *diam.Message, cfgFld *config.CfgCdrField, extraParam interface{}) (fmtValOut string, err error) {
	var outVal string
	passAtIndex := -1
	passedAllFilters := true
	for _, fldFilter := range cfgFld.FieldFilter {
		var pass bool
		if pass, passAtIndex = passesFieldFilter(m, fldFilter); !pass {
			passedAllFilters = false
			break
		}
	}
	if !passedAllFilters {
		return "", ErrFilterNotPassing // Not matching field filters, will have it empty
	}
	if passAtIndex == -1 {
		passAtIndex = 0 // No filter
	}
	switch cfgFld.Type {
	case utils.META_FILLER:
		outVal = cfgFld.Value.Id()
		cfgFld.Padding = "right"
	case utils.META_CONSTANT:
		outVal = cfgFld.Value.Id()
	case utils.META_HANDLER:
		if cfgFld.HandlerId == META_CCA_USAGE { // Exception, usage is passed in the dur variable by CCA
			outVal = strconv.FormatFloat(extraParam.(float64), 'f', -1, 64)
		} else if cfgFld.HandlerId == META_VALUE_EXPONENT {
			outVal, err = metaValueExponent(m, cfgFld.Value, 10) // FixMe: add here configured number of decimals
		} else {
			outVal, err = metaHandler(m, cfgFld.HandlerId, cfgFld.Layout, extraParam.(time.Duration))
			if err != nil {
				utils.Logger.Warning(fmt.Sprintf("<Diameter> Ignoring processing of metafunction: %s, error: %s", cfgFld.HandlerId, err.Error()))
			}
		}
	case utils.META_COMPOSED:
		outVal = composedFieldvalue(m, cfgFld.Value, 0)
	case utils.MetaGrouped: // GroupedAVP
		outVal = composedFieldvalue(m, cfgFld.Value, passAtIndex)
	}
	if fmtValOut, err = utils.FmtFieldWidth(outVal, cfgFld.Width, cfgFld.Strip, cfgFld.Padding, cfgFld.Mandatory); err != nil {
		utils.Logger.Warning(fmt.Sprintf("<Diameter> Error when processing field template with tag: %s, error: %s", cfgFld.Tag, err.Error()))
		return "", err
	}
	return fmtValOut, nil
}

// messageAddAVPsWithPath will dynamically add AVPs into the message
// 	append:	append to the message, on false overwrite if AVP is single or add to group if AVP is Grouped
func messageSetAVPsWithPath(m *diam.Message, path []interface{}, avpValStr string, appnd bool, timezone string) error {
	if len(path) == 0 {
		return errors.New("Empty path as AVP filter")
	}
	dictAVPs := make([]*dict.AVP, len(path)) // for each subpath, one dictionary AVP
	for i, subpath := range path {
		if dictAVP, err := m.Dictionary().FindAVP(m.Header.ApplicationID, subpath); err != nil {
			return err
		} else if dictAVP == nil {
			return fmt.Errorf("Cannot find AVP with id: %s", path[len(path)-1])
		} else {
			dictAVPs[i] = dictAVP
		}
	}
	if dictAVPs[len(path)-1].Data.Type == diam.GroupedAVPType {
		return errors.New("Last AVP in path needs not to be GroupedAVP")
	}
	var msgAVP *diam.AVP // Keep a reference here towards last AVP
	lastAVPIdx := len(path) - 1
	for i := lastAVPIdx; i >= 0; i-- {
		var typeVal datatype.Type
		if i == lastAVPIdx {
			avpValByte, err := serializeAVPValueFromString(dictAVPs[i], avpValStr, timezone)
			if err != nil {
				return err
			}
			typeVal, err = datatype.Decode(dictAVPs[i].Data.Type, avpValByte)
			if err != nil {
				return err
			}
		} else {
			typeVal = &diam.GroupedAVP{
				AVP: []*diam.AVP{msgAVP}}
		}
		newMsgAVP := diam.NewAVP(dictAVPs[i].Code, avp.Mbit, dictAVPs[i].VendorID, typeVal) // FixMe: maybe Mbit with dictionary one
		if i == lastAVPIdx-1 && !appnd {                                                    // last AVP needs to be appended in group
			avps, _ := m.FindAVPsWithPath(path[:lastAVPIdx], dict.UndefinedVendorID)
			if len(avps) != 0 { // Group AVP already in the message
				prevGrpData := avps[0].Data.(*diam.GroupedAVP)
				prevGrpData.AVP = append(prevGrpData.AVP, msgAVP)
				m.Header.MessageLength += uint32(msgAVP.Len())
				return nil
			}
		}
		msgAVP = newMsgAVP
	}
	if !appnd { // Not group AVP, replace the previous set one with this one
		avps, _ := m.FindAVPsWithPath(path, dict.UndefinedVendorID)
		if len(avps) != 0 { // Group AVP already in the message
			m.Header.MessageLength -= uint32(avps[0].Len()) // decrease message length since we overwrite
			*avps[0] = *msgAVP
			m.Header.MessageLength += uint32(msgAVP.Len())
			return nil
		}
	}
	m.AVP = append(m.AVP, msgAVP)
	m.Header.MessageLength += uint32(msgAVP.Len())
	return nil
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
// FixMe: strip it down to mandatory bare structure format by RFC 4006
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

// AsBareDiameterMessage converts CCR into a bare DiameterMessage
// Compatible with the required fields of CCA
func (self *CCR) AsBareDiameterMessage() *diam.Message {
	m := diam.NewRequest(diam.CreditControl, 4, nil)
	m.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(self.SessionId))
	m.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity(self.OriginHost))
	m.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity(self.OriginRealm))
	m.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(self.AuthApplicationId))
	m.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(self.CCRequestType))
	m.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(self.CCRequestNumber))
	return m
}

// Used when sending from client to agent
func (self *CCR) AsDiameterMessage() (*diam.Message, error) {
	m := self.AsBareDiameterMessage()
	if _, err := m.NewAVP("Destination-Host", avp.Mbit, 0, datatype.DiameterIdentity(self.DestinationHost)); err != nil {
		return nil, err
	}
	if _, err := m.NewAVP("Destination-Realm", avp.Mbit, 0, datatype.DiameterIdentity(self.DestinationRealm)); err != nil {
		return nil, err
	}
	if _, err := m.NewAVP("Service-Context-Id", avp.Mbit, 0, datatype.UTF8String(self.ServiceContextId)); err != nil {
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

// Extracts data out of CCR into a SMGenericEvent based on the configured template
func (self *CCR) AsSMGenericEvent(cfgFlds []*config.CfgCdrField) (sessionmanager.SMGenericEvent, error) {
	outMap := make(map[string]string) // work with it so we can append values to keys
	outMap[utils.EVENT_NAME] = DIAMETER_CCR
	for _, cfgFld := range cfgFlds {
		fmtOut, err := fieldOutVal(self.diamMessage, cfgFld, self.debitInterval)
		if err != nil {
			if err == ErrFilterNotPassing {
				continue // Do nothing in case of Filter not passing
			}
			return nil, err
		}
		if _, hasKey := outMap[cfgFld.FieldId]; hasKey && cfgFld.Append {
			outMap[cfgFld.FieldId] += fmtOut
		} else {
			outMap[cfgFld.FieldId] = fmtOut

		}
	}
	return sessionmanager.SMGenericEvent(utils.ConvertMapValStrIf(outMap)), nil
}

func NewBareCCAFromCCR(ccr *CCR, originHost, originRealm string) *CCA {
	cca := &CCA{SessionId: ccr.SessionId, AuthApplicationId: ccr.AuthApplicationId, CCRequestType: ccr.CCRequestType, CCRequestNumber: ccr.CCRequestNumber,
		OriginHost: originHost, OriginRealm: originRealm,
		diamMessage: diam.NewMessage(ccr.diamMessage.Header.CommandCode, ccr.diamMessage.Header.CommandFlags&^diam.RequestFlag, ccr.diamMessage.Header.ApplicationID,
			ccr.diamMessage.Header.HopByHopID, ccr.diamMessage.Header.EndToEndID, ccr.diamMessage.Dictionary()), ccrMessage: ccr.diamMessage, debitInterval: ccr.debitInterval,
	}
	cca.diamMessage = cca.AsBareDiameterMessage() // Add the required fields to the diameterMessage
	return cca
}

// Call Control Answer, bare structure so we can dynamically manage adding it's fields
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
	ccrMessage    *diam.Message
	diamMessage   *diam.Message
	debitInterval time.Duration
	timezone      string
}

// AsBareDiameterMessage converts CCA into a bare DiameterMessage
func (self *CCA) AsBareDiameterMessage() *diam.Message {
	var m diam.Message
	utils.Clone(self.diamMessage, &m)
	m.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(self.SessionId))
	m.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity(self.OriginHost))
	m.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity(self.OriginRealm))
	m.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(self.AuthApplicationId))
	m.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(self.CCRequestType))
	m.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Enumerated(self.CCRequestNumber))
	m.NewAVP(avp.ResultCode, avp.Mbit, 0, datatype.Unsigned32(self.ResultCode))
	return &m
}

// AsDiameterMessage returns the diameter.Message which can be later written on network
func (self *CCA) AsDiameterMessage() *diam.Message {
	return self.diamMessage
}

// SetProcessorAVPs will add AVPs to self.diameterMessage based on template defined in processor.CCAFields
func (self *CCA) SetProcessorAVPs(reqProcessor *config.DARequestProcessor, processorVars map[string]string) error {
	for _, cfgFld := range reqProcessor.CCAFields {
		fmtOut, err := fieldOutVal(self.ccrMessage, cfgFld, processorVars)
		if err == ErrFilterNotPassing { // Field not in or filter not passing, try match in answer
			fmtOut, err = fieldOutVal(self.diamMessage, cfgFld, processorVars)
		}
		if err != nil {
			if err == ErrFilterNotPassing {
				continue
			}
			return err
		}
		if err := messageSetAVPsWithPath(self.diamMessage, splitIntoInterface(cfgFld.FieldId, utils.HIERARCHY_SEP), fmtOut, cfgFld.Append, self.timezone); err != nil {
			return err
		}
	}
	return nil
}
