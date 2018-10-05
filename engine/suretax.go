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

package engine

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var sureTaxClient *http.Client // Cache the client here if in use

// Init a new request to be sent out to SureTax
func NewSureTaxRequest(cdr *CDR, stCfg *config.SureTaxCfg) (*SureTaxRequest, error) {
	if stCfg == nil {
		return nil, errors.New("Invalid SureTax config.")
	}
	aTimeLoc := cdr.AnswerTime.In(stCfg.Timezone)
	revenue := utils.Round(cdr.Cost, 4, utils.ROUNDING_MIDDLE)
	unts, err := strconv.ParseInt(cdr.FieldsAsStringWithRSRFields(stCfg.Units), 10, 64)
	if err != nil {
		return nil, err
	}
	taxExempt := []string{}
	definedTaxExtempt := cdr.FieldsAsStringWithRSRFields(stCfg.TaxExemptionCodeList)
	if len(definedTaxExtempt) != 0 {
		taxExempt = strings.Split(cdr.FieldsAsStringWithRSRFields(stCfg.TaxExemptionCodeList), ",")
	}
	stReq := new(STRequest)
	stReq.ClientNumber = stCfg.ClientNumber
	stReq.BusinessUnit = stCfg.BusinessUnit
	stReq.ValidationKey = stCfg.ValidationKey
	stReq.DataYear = strconv.Itoa(aTimeLoc.Year())
	stReq.DataMonth = strconv.Itoa(int(aTimeLoc.Month()))
	stReq.TotalRevenue = revenue
	stReq.ReturnFileCode = stCfg.ReturnFileCode
	stReq.ClientTracking = cdr.FieldsAsStringWithRSRFields(stCfg.ClientTracking)
	stReq.ResponseGroup = stCfg.ResponseGroup
	stReq.ResponseType = stCfg.ResponseType
	stReq.ItemList = []*STRequestItem{
		{
			CustomerNumber:       cdr.FieldsAsStringWithRSRFields(stCfg.CustomerNumber),
			OrigNumber:           cdr.FieldsAsStringWithRSRFields(stCfg.OrigNumber),
			TermNumber:           cdr.FieldsAsStringWithRSRFields(stCfg.TermNumber),
			BillToNumber:         cdr.FieldsAsStringWithRSRFields(stCfg.BillToNumber),
			Zipcode:              cdr.FieldsAsStringWithRSRFields(stCfg.Zipcode),
			Plus4:                cdr.FieldsAsStringWithRSRFields(stCfg.Plus4),
			P2PZipcode:           cdr.FieldsAsStringWithRSRFields(stCfg.P2PZipcode),
			P2PPlus4:             cdr.FieldsAsStringWithRSRFields(stCfg.P2PPlus4),
			TransDate:            aTimeLoc.Format("2006-01-02T15:04:05"),
			Revenue:              revenue,
			Units:                unts,
			UnitType:             cdr.FieldsAsStringWithRSRFields(stCfg.UnitType),
			Seconds:              int64(cdr.Usage.Seconds()),
			TaxIncludedCode:      cdr.FieldsAsStringWithRSRFields(stCfg.TaxIncluded),
			TaxSitusRule:         cdr.FieldsAsStringWithRSRFields(stCfg.TaxSitusRule),
			TransTypeCode:        cdr.FieldsAsStringWithRSRFields(stCfg.TransTypeCode),
			SalesTypeCode:        cdr.FieldsAsStringWithRSRFields(stCfg.SalesTypeCode),
			RegulatoryCode:       stCfg.RegulatoryCode,
			TaxExemptionCodeList: taxExempt,
		},
	}
	jsnContent, err := json.Marshal(stReq)
	if err != nil {
		return nil, err
	}
	return &SureTaxRequest{Request: string(jsnContent)}, nil
}

// SureTax JSON Request
type SureTaxRequest struct {
	Request string `json:"request"` // SureTax Requires us to encapsulate the content into a request element
}

// SureTax JSON Response
type SureTaxResponse struct {
	D string // SureTax requires encapsulating reply into a D object
}

// SureTax Request type
type STRequest struct {
	ClientNumber      string           // Client ID Number – provided by SureTax. Required. Max Len: 10
	BusinessUnit      string           // Client’s Business Unit. Value for this field is not required. Max Len: 20
	ValidationKey     string           // Validation Key provided by SureTax. Required for client access to API function. Max Len: 36
	DataYear          string           // Required. YYYY – Year to use for tax calculation purposes
	DataMonth         string           // Required. MM – Month to use for tax calculation purposes. Leading zero is preferred.
	TotalRevenue      float64          // Required. Format: $$$$$$$$$.CCCC. For Negative charges, the first position should have a minus ‘-‘ indicator.
	ReturnFileCode    string           // Required. 0 – Default.Q – Quote purposes – taxes are computed and returned in the response message for generating quotes.
	ClientTracking    string           // Field for client transaction tracking. This value will be provided in the response data. Value for this field is not required, but preferred. Max Len: 100
	IndustryExemption string           // Reserved for future use.
	ResponseGroup     string           // Required. Determines how taxes are grouped for the response.
	ResponseType      string           // Required. Determines the granularity of taxes and (optionally) the decimal precision for the tax calculations and amounts in the response.
	ItemList          []*STRequestItem // List of Item records
}

// Part of SureTax Request
type STRequestItem struct {
	LineNumber           string   // Used to identify an item within the request. If no value is provided, requests are numbered sequentially. Max Len: 40
	InvoiceNumber        string   // Used for tax aggregation by Invoice. Must be alphanumeric. Max Len: 40
	CustomerNumber       string   // Used for tax aggregation by Customer. Must be alphanumeric. Max Len: 40
	OrigNumber           string   // Required when using Tax Situs Rule 01 or 03. Format: NPANXXNNNN
	TermNumber           string   // Required when using Tax Situs Rule 01. Format: NPANXXNNNN
	BillToNumber         string   // Required when using Tax Situs Rule 01 or 02. Format: NPANXXNNNN
	Zipcode              string   // Required when using Tax Situs Rule 04, 05, or 14.
	Plus4                string   // Zip code extension in format: 9999 (not applicable for Tax Situs Rule 14)
	P2PZipcode           string   // Secondary zip code in format: 99999 (US or US territory) or X9X9X9 (Canadian)
	P2PPlus4             string   // Secondary zip code extension in format: 99999 (US or US territory) or X9X9X9         (Canadian)
	TransDate            string   // Required. Date of transaction. Valid date formats include: MM/DD/YYYY, MM-DD-YYYY, YYYY-MM-DDTHH:MM:SS
	Revenue              float64  // Required. Format: $$$$$$$$$.CCCC. For Negative charges, the first position should have a minus ‘-‘indicator.
	Units                int64    // Required. Units representing number of “lines” or unique charges contained within the revenue. This value is essentially a multiplier on unit-based fees (e.g. E911 fees). Format: 99999. Default should be 1 (one unit).
	UnitType             string   // Required. 00 – Default / Number of unique access lines.
	Seconds              int64    // Required. Duration of call in seconds. Format 99999. Default should be 1.
	TaxIncludedCode      string   // Required. Values: 0 – Default (No Tax Included) 1 – Tax Included in Revenue
	TaxSitusRule         string   // Required.
	TransTypeCode        string   // Required. Transaction Type Indicator.
	SalesTypeCode        string   // Required. Values: R – Residential customer (default) B – Business customer I – Industrial customer L – Lifeline customer
	RegulatoryCode       string   // Required. Provider Type.
	TaxExemptionCodeList []string // Required. Tax Exemption to be applied to this item only.
}

type STResponse struct {
	Successful     string           // Response will be either ‘Y' or ‘N' : Y = Success / Success with Item error N = Failure
	ResponseCode   string           // ResponseCode: 9999 – Request was successful. 1101-1400 – Range of values for a failed request (no processing occurred) 9001 – Request was successful, but items within the request have errors. The specific items with errors are provided in the ItemMessages field.
	HeaderMessage  string           // Response message: For ResponseCode 9999 – “Success”For ResponseCode 9001 – “Success with Item errors”.  For ResponseCode 1100-1400 – Unsuccessful / declined web request.
	ItemMessages   []*STItemMessage // This field contains a list of items that were not able to be processed due to bad or invalid data (see Response Code of “9001”).
	ClientTracking string           // Client transaction tracking provided in web request.
	TotalTax       string           // Total Tax – a total of all taxes included in the TaxList
	TransId        int              // Transaction ID – provided by SureTax
	GroupList      []*STGroup       // contains one-to-many Groups
}

// Part of the SureTax Response
type STItemMessage struct {
	LineNumber   string // value corresponding to the line number in the web request
	ResponseCode string // a value in the range 9100-9400
	Message      string // the error message corresponding to the ResponseCode
}

// Part of the SureTax Response
type STGroup struct {
	StateCode      string       // Tax State
	InvoiceNumber  string       // Invoice Number
	CustomerNumber string       // Customer number
	TaxList        []*STTaxItem // contains one-to-many Tax Items
}

// Part of the SureTax Response
type STTaxItem struct {
	TaxTypeCode string // Tax Type Code
	TaxTypeDesc string // Tax Type Description
	TaxAmount   string // Tax Amount
}

func SureTaxProcessCdr(cdr *CDR) error {
	stCfg := config.CgrConfig().SureTaxCfg()
	if stCfg == nil {
		return errors.New("Invalid SureTax configuration")
	}
	if sureTaxClient == nil { // First time used, init the client here
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: config.CgrConfig().GeneralCfg().HttpSkipTlsVerify,
			},
		}
		sureTaxClient = &http.Client{Transport: tr}
	}
	req, err := NewSureTaxRequest(cdr, stCfg)
	if err != nil {
		return err
	}
	jsnContent, err := json.Marshal(req)
	if err != nil {
		return err
	}
	resp, err := sureTaxClient.Post(stCfg.Url, "application/json", bytes.NewBuffer(jsnContent))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode > 299 {
		return fmt.Errorf("Unexpected status code received: %d", resp.StatusCode)
	}
	var respFull SureTaxResponse
	if err := json.Unmarshal(respBody, &respFull); err != nil {
		return err
	}
	var stResp STResponse
	if err := json.Unmarshal([]byte(respFull.D), &stResp); err != nil {
		return err
	}
	if stResp.ResponseCode != "9999" {
		cdr.ExtraInfo = stResp.HeaderMessage
		return nil // No error because the request was processed by SureTax, error will be in the ExtraInfo
	}
	// Write cost to CDR
	totalTax, err := strconv.ParseFloat(stResp.TotalTax, 64)
	if err != nil {
		cdr.ExtraInfo = err.Error()
	}
	if !stCfg.IncludeLocalCost {
		cdr.Cost = utils.Round(totalTax,
			config.CgrConfig().GeneralCfg().RoundingDecimals,
			utils.ROUNDING_MIDDLE)
	} else {
		cdr.Cost = utils.Round(cdr.Cost+totalTax,
			config.CgrConfig().GeneralCfg().RoundingDecimals,
			utils.ROUNDING_MIDDLE)
	}
	// Add response into extra fields to be available for later review
	cdr.ExtraFields[utils.META_SURETAX] = respFull.D
	return nil
}
