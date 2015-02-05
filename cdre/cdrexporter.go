/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2014 ITsysCOM GmbH

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

package cdre

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"io"
	"os"
	"strconv"
	"time"
)

const (
	COST_DETAILS          = "cost_details"
	CONCATENATED_CDRFIELD = "concatenated_cdrfield"
	DATETIME              = "datetime"
	META_EXPORTID         = "export_id"
	META_TIMENOW          = "time_now"
	META_FIRSTCDRATIME    = "first_cdr_atime"
	META_LASTCDRATIME     = "last_cdr_atime"
	META_NRCDRS           = "cdrs_number"
	META_DURCDRS          = "cdrs_duration"
	META_SMSUSAGE         = "sms_usage"
	META_DATAUSAGE        = "data_usage"
	META_COSTCDRS         = "cdrs_cost"
	META_MASKDESTINATION  = "mask_destination"
	META_FORMATCOST       = "format_cost"
)

var err error

func NewCdrExporter(cdrs []*utils.StoredCdr, logDb engine.LogStorage, exportTpl *config.CdreConfig, cdrFormat string, fieldSeparator rune, exportId string,
	dataUsageMultiplyFactor, smsUsageMultiplyFactor, costMultiplyFactor float64, costShiftDigits, roundDecimals, cgrPrecision int, maskDestId string, maskLen int, httpSkipTlsCheck bool) (*CdrExporter, error) {
	if len(cdrs) == 0 { // Nothing to export
		return nil, nil
	}
	cdre := &CdrExporter{
		cdrs:                    cdrs,
		logDb:                   logDb,
		exportTemplate:          exportTpl,
		cdrFormat:               cdrFormat,
		fieldSeparator:          fieldSeparator,
		exportId:                exportId,
		dataUsageMultiplyFactor: dataUsageMultiplyFactor,
		smsUsageMultiplyFactor:  smsUsageMultiplyFactor,
		costMultiplyFactor:      costMultiplyFactor,
		costShiftDigits:         costShiftDigits,
		roundDecimals:           roundDecimals,
		cgrPrecision:            cgrPrecision,
		maskDestId:              maskDestId,
		httpSkipTlsCheck:        httpSkipTlsCheck,
		maskLen:                 maskLen,
		negativeExports:         make(map[string]string),
	}
	if err := cdre.processCdrs(); err != nil {
		return nil, err
	}
	return cdre, nil
}

type CdrExporter struct {
	cdrs           []*utils.StoredCdr
	logDb          engine.LogStorage // Used to extract cost_details if these are requested
	exportTemplate *config.CdreConfig
	cdrFormat      string // csv, fwv
	fieldSeparator rune
	exportId       string // Unique identifier or this export
	dataUsageMultiplyFactor,
	smsUsageMultiplyFactor, // Multiply the SMS usage (eg: some billing systems billing them as minutes)
	costMultiplyFactor float64
	costShiftDigits, roundDecimals, cgrPrecision int
	maskDestId                                   string
	maskLen                                      int
	httpSkipTlsCheck                             bool
	header, trailer                              []string   // Header and Trailer fields
	content                                      [][]string // Rows of cdr fields
	firstCdrATime, lastCdrATime                  time.Time
	numberOfRecords                              int
	totalDuration, totalDataUsage, totalSmsUsage time.Duration

	totalCost                       float64
	firstExpOrderId, lastExpOrderId int64
	positiveExports                 []string          // CGRIds of successfully exported CDRs
	negativeExports                 map[string]string // CgrIds of failed exports
}

// Return Json marshaled callCost attached to
// Keep it separately so we test only this part in local tests
func (cdre *CdrExporter) getCdrCostDetails(cgrId, runId string) (string, error) {
	cc, err := cdre.logDb.GetCallCostLog(cgrId, "", runId)
	if err != nil {
		return "", err
	} else if cc == nil {
		return "", nil
	}
	ccJson, _ := json.Marshal(cc)
	return string(ccJson), nil
}

func (cdre *CdrExporter) getCombimedCdrFieldVal(processedCdr *utils.StoredCdr, cfgCdrFld *config.CfgCdrField) (string, error) {
	var combinedVal string // Will result as combination of the field values, filters must match
	for _, filterRule := range cfgCdrFld.Filter {
		fltrPass, ftrPassValue := processedCdr.PassesFieldFilter(filterRule)
		if !fltrPass {
			return "", nil
		}
		for _, cdr := range cdre.cdrs {
			if cdr.CgrId != processedCdr.CgrId {
				continue // We only care about cdrs with same primary cdr behind
			}
			if cdr.FieldAsString(&utils.RSRField{Id: filterRule.Id}) == ftrPassValue { // First CDR with filte
				for _, rsrRule := range cfgCdrFld.Value {
					combinedVal += cdr.FieldAsString(rsrRule)
				}
			}
		}
	}
	return combinedVal, nil
}

// Check if the destination should be masked in output
func (cdre *CdrExporter) maskedDestination(destination string) bool {
	if len(cdre.maskDestId) != 0 && engine.CachedDestHasPrefix(cdre.maskDestId, destination) {
		return true
	}
	return false
}

func (cdre *CdrExporter) getDateTimeFieldVal(cdr *utils.StoredCdr, cfgCdrFld *config.CfgCdrField) (string, error) {
	if len(cfgCdrFld.Value) == 0 {
		return "", nil
	}
	for _, fltrRl := range cfgCdrFld.Filter {
		if fltrPass, _ := cdr.PassesFieldFilter(fltrRl); !fltrPass {
			return "", fmt.Errorf("Field: %s not matching filter rule %v", fltrRl.Id, fltrRl)
		}
	}
	layout := cfgCdrFld.Layout
	if len(layout) == 0 {
		layout = time.RFC3339
	}
	if dtFld, err := utils.ParseTimeDetectLayout(cdr.FieldAsString(cfgCdrFld.Value[0])); err != nil { // Only one rule makes sense here
		return "", err
	} else {
		return dtFld.Format(layout), nil
	}
}

// Extracts the value specified by cfgHdr out of cdr
func (cdre *CdrExporter) cdrFieldValue(cdr *utils.StoredCdr, cfgCdrFld *config.CfgCdrField) (string, error) {
	for _, fltrRl := range cfgCdrFld.Filter {
		if fltrPass, _ := cdr.PassesFieldFilter(fltrRl); !fltrPass {
			return "", fmt.Errorf("Field: %s not matching filter rule %v", fltrRl.Id, fltrRl)
		}
	}
	layout := cfgCdrFld.Layout
	if len(layout) == 0 {
		layout = time.RFC3339
	}
	var retVal string // Concatenate the resulting values
	for _, rsrFld := range cfgCdrFld.Value {
		var cdrVal string
		switch rsrFld.Id {
		case COST_DETAILS: // Special case when we need to further extract cost_details out of logDb
			if cdrVal, err = cdre.getCdrCostDetails(cdr.CgrId, cdr.MediationRunId); err != nil {
				return "", err
			}
		case utils.COST:
			cdrVal = cdr.FormatCost(cdre.costShiftDigits, cdre.roundDecimals)
		case utils.USAGE:
			cdrVal = cdr.FormatUsage(layout)
		case utils.SETUP_TIME:
			cdrVal = cdr.SetupTime.Format(layout)
		case utils.ANSWER_TIME: // Format time based on layout
			cdrVal = cdr.AnswerTime.Format(layout)
		case utils.DESTINATION:
			cdrVal = cdr.FieldAsString(rsrFld)
			if cdre.maskLen != -1 && cdre.maskedDestination(cdrVal) {
				cdrVal = MaskDestination(cdrVal, cdre.maskLen)
			}
		default:
			cdrVal = cdr.FieldAsString(rsrFld)
		}
		retVal += cdrVal
	}
	return retVal, nil
}

// Handle various meta functions used in header/trailer
func (cdre *CdrExporter) metaHandler(tag, arg string) (string, error) {
	switch tag {
	case META_EXPORTID:
		return cdre.exportId, nil
	case META_TIMENOW:
		return time.Now().Format(arg), nil
	case META_FIRSTCDRATIME:
		return cdre.firstCdrATime.Format(arg), nil
	case META_LASTCDRATIME:
		return cdre.lastCdrATime.Format(arg), nil
	case META_NRCDRS:
		return strconv.Itoa(cdre.numberOfRecords), nil
	case META_DURCDRS:
		emulatedCdr := &utils.StoredCdr{TOR: utils.VOICE, Usage: cdre.totalDuration}
		return emulatedCdr.FormatUsage(arg), nil
	case META_SMSUSAGE:
		emulatedCdr := &utils.StoredCdr{TOR: utils.SMS, Usage: cdre.totalSmsUsage}
		return emulatedCdr.FormatUsage(arg), nil
	case META_DATAUSAGE:
		emulatedCdr := &utils.StoredCdr{TOR: utils.DATA, Usage: cdre.totalDataUsage}
		return emulatedCdr.FormatUsage(arg), nil
	case META_COSTCDRS:
		return strconv.FormatFloat(utils.Round(cdre.totalCost, cdre.roundDecimals, utils.ROUNDING_MIDDLE), 'f', -1, 64), nil
	case META_MASKDESTINATION:
		if cdre.maskedDestination(arg) {
			return "1", nil
		}
		return "0", nil
	default:
		return "", fmt.Errorf("Unsupported METATAG: %s", tag)
	}
	return "", nil
}

// Compose and cache the header
func (cdre *CdrExporter) composeHeader() error {
	for _, cfgFld := range cdre.exportTemplate.HeaderFields {
		var outVal string
		switch cfgFld.Type {
		case utils.FILLER:
			for _, rsrFld := range cfgFld.Value {
				outVal += rsrFld.ParseValue("")
			}
			cfgFld.Padding = "right"
		case utils.CONSTANT:
			for _, rsrFld := range cfgFld.Value {
				outVal += rsrFld.ParseValue("")
			}
		case utils.METATAG:
			outVal, err = cdre.metaHandler(cfgFld.CdrFieldId, cfgFld.Layout)
		default:
			return fmt.Errorf("Unsupported field type: %s", cfgFld.Type)
		}
		if err != nil {
			engine.Logger.Err(fmt.Sprintf("<CdreFw> Cannot export CDR header, field %s, error: %s", cfgFld.Tag, err.Error()))
			return err
		}
		fmtOut := outVal
		if fmtOut, err = FmtFieldWidth(outVal, cfgFld.Width, cfgFld.Strip, cfgFld.Padding, cfgFld.Mandatory); err != nil {
			engine.Logger.Err(fmt.Sprintf("<CdreFw> Cannot export CDR header, field %s, error: %s", cfgFld.Tag, err.Error()))
			return err
		}
		cdre.header = append(cdre.header, fmtOut)
	}
	return nil
}

// Compose and cache the trailer
func (cdre *CdrExporter) composeTrailer() error {
	for _, cfgFld := range cdre.exportTemplate.TrailerFields {
		var outVal string
		switch cfgFld.Type {
		case utils.FILLER:
			for _, rsrFld := range cfgFld.Value {
				outVal += rsrFld.ParseValue("")
			}
			cfgFld.Padding = "right"
		case utils.CONSTANT:
			for _, rsrFld := range cfgFld.Value {
				outVal += rsrFld.ParseValue("")
			}
		case utils.METATAG:
			outVal, err = cdre.metaHandler(cfgFld.CdrFieldId, cfgFld.Layout)
		default:
			return fmt.Errorf("Unsupported field type: %s", cfgFld.Type)
		}
		if err != nil {
			engine.Logger.Err(fmt.Sprintf("<CdreFw> Cannot export CDR trailer, field: %s, error: %s", cfgFld.Tag, err.Error()))
			return err
		}
		fmtOut := outVal
		if fmtOut, err = FmtFieldWidth(outVal, cfgFld.Width, cfgFld.Strip, cfgFld.Padding, cfgFld.Mandatory); err != nil {
			engine.Logger.Err(fmt.Sprintf("<CdreFw> Cannot export CDR trailer, field: %s, error: %s", cfgFld.Tag, err.Error()))
			return err
		}
		cdre.trailer = append(cdre.trailer, fmtOut)
	}
	return nil
}

// Write individual cdr into content buffer, build stats
func (cdre *CdrExporter) processCdr(cdr *utils.StoredCdr) error {
	if cdr == nil || len(cdr.CgrId) == 0 { // We do not export empty CDRs
		return nil
	}
	if cdre.dataUsageMultiplyFactor != 0.0 && cdr.TOR == utils.DATA {
		cdr.UsageMultiply(cdre.dataUsageMultiplyFactor, cdre.cgrPrecision)
	}
	if cdre.smsUsageMultiplyFactor != 0 && cdr.TOR == utils.SMS {
		cdr.UsageMultiply(cdre.smsUsageMultiplyFactor, cdre.cgrPrecision)
	}
	if cdre.costMultiplyFactor != 0.0 {
		cdr.CostMultiply(cdre.smsUsageMultiplyFactor, cdre.cgrPrecision)
	}
	var err error
	cdrRow := make([]string, len(cdre.exportTemplate.ContentFields))
	for idx, cfgFld := range cdre.exportTemplate.ContentFields {
		var outVal string
		switch cfgFld.Type {
		case utils.FILLER:
			for _, rsrFld := range cfgFld.Value {
				outVal += rsrFld.ParseValue("")
			}
			cfgFld.Padding = "right"
		case utils.CONSTANT:
			for _, rsrFld := range cfgFld.Value {
				outVal += rsrFld.ParseValue("")
			}
		case utils.CDRFIELD:
			outVal, err = cdre.cdrFieldValue(cdr, cfgFld)
		case DATETIME:
			outVal, err = cdre.getDateTimeFieldVal(cdr, cfgFld)
		case utils.HTTP_POST:
			var outValByte []byte
			var httpAddr string
			for _, rsrFld := range cfgFld.Value {
				httpAddr += rsrFld.ParseValue("")
			}
			if outValByte, err = utils.HttpJsonPost(httpAddr, cdre.httpSkipTlsCheck, cdr); err == nil {
				outVal = string(outValByte)
				if len(outVal) == 0 && cfgFld.Mandatory {
					err = fmt.Errorf("Empty result for http_post field: %s", cfgFld.Tag)
				}
			}
		case utils.COMBIMED:
			outVal, err = cdre.getCombimedCdrFieldVal(cdr, cfgFld)
		case utils.METATAG:
			if cfgFld.CdrFieldId == META_MASKDESTINATION {
				outVal, err = cdre.metaHandler(cfgFld.CdrFieldId, cdr.FieldAsString(&utils.RSRField{Id: utils.DESTINATION}))
			} else {
				outVal, err = cdre.metaHandler(cfgFld.CdrFieldId, cfgFld.Layout)
			}
		}
		if err != nil {
			engine.Logger.Err(fmt.Sprintf("<CdreFw> Cannot export CDR with cgrid: %s and runid: %s, error: %s", cdr.CgrId, cdr.MediationRunId, err.Error()))
			return err
		}
		fmtOut := outVal
		if fmtOut, err = FmtFieldWidth(outVal, cfgFld.Width, cfgFld.Strip, cfgFld.Padding, cfgFld.Mandatory); err != nil {
			engine.Logger.Err(fmt.Sprintf("<CdreFw> Cannot export CDR with cgrid: %s, runid: %s, fieldName: %s, fieldValue: %s, error: %s", cdr.CgrId, cdr.MediationRunId, cfgFld.Tag, outVal, err.Error()))
			return err
		}
		cdrRow[idx] += fmtOut
	}
	if len(cdrRow) == 0 { // No CDR data, most likely no configuration fields defined
		return nil
	} else {
		cdre.content = append(cdre.content, cdrRow)
	}
	// Done with writing content, compute stats here
	if cdre.firstCdrATime.IsZero() || cdr.AnswerTime.Before(cdre.firstCdrATime) {
		cdre.firstCdrATime = cdr.AnswerTime
	}
	if cdr.AnswerTime.After(cdre.lastCdrATime) {
		cdre.lastCdrATime = cdr.AnswerTime
	}
	cdre.numberOfRecords += 1
	if cdr.TOR == utils.VOICE { // Only count duration for non data cdrs
		cdre.totalDuration += cdr.Usage
	}
	if cdr.TOR == utils.SMS { // Count usage for SMS
		cdre.totalSmsUsage += cdr.Usage
	}
	if cdr.TOR == utils.DATA { // Count usage for SMS
		cdre.totalDataUsage += cdr.Usage
	}
	if cdr.Cost != -1 {
		cdre.totalCost += cdr.Cost
		cdre.totalCost = utils.Round(cdre.totalCost, cdre.roundDecimals, utils.ROUNDING_MIDDLE)
	}
	if cdre.firstExpOrderId > cdr.OrderId || cdre.firstExpOrderId == 0 {
		cdre.firstExpOrderId = cdr.OrderId
	}
	if cdre.lastExpOrderId < cdr.OrderId {
		cdre.lastExpOrderId = cdr.OrderId
	}
	return nil
}

// Builds header, content and trailers
func (cdre *CdrExporter) processCdrs() error {
	for _, cdr := range cdre.cdrs {
		if err := cdre.processCdr(cdr); err != nil {
			cdre.negativeExports[cdr.CgrId] = err.Error()
		} else {
			cdre.positiveExports = append(cdre.positiveExports, cdr.CgrId)
		}
	}
	// Process header and trailer after processing cdrs since the metatag functions can access stats out of built cdrs
	if cdre.exportTemplate.HeaderFields != nil {
		if err := cdre.composeHeader(); err != nil {
			return err
		}
	}
	if cdre.exportTemplate.TrailerFields != nil {
		if err := cdre.composeTrailer(); err != nil {
			return err
		}
	}
	return nil
}

// Simple write method
func (cdre *CdrExporter) writeOut(ioWriter io.Writer) error {
	if len(cdre.header) != 0 {
		for _, fld := range append(cdre.header, "\n") {
			if _, err := io.WriteString(ioWriter, fld); err != nil {
				return err
			}
		}
	}
	for _, cdrContent := range cdre.content {
		for _, cdrFld := range append(cdrContent, "\n") {
			if _, err := io.WriteString(ioWriter, cdrFld); err != nil {
				return err
			}
		}
	}
	if len(cdre.trailer) != 0 {
		for _, fld := range append(cdre.trailer, "\n") {
			if _, err := io.WriteString(ioWriter, fld); err != nil {
				return err
			}
		}
	}
	return nil
}

// csvWriter specific method
func (cdre *CdrExporter) writeCsv(csvWriter *csv.Writer) error {
	csvWriter.Comma = cdre.fieldSeparator
	if len(cdre.header) != 0 {
		if err := csvWriter.Write(cdre.header); err != nil {
			return err
		}
	}
	for _, cdrContent := range cdre.content {
		if err := csvWriter.Write(cdrContent); err != nil {
			return err
		}
	}
	if len(cdre.trailer) != 0 {
		if err := csvWriter.Write(cdre.trailer); err != nil {
			return err
		}
	}
	csvWriter.Flush()
	return nil
}

// General method to write the content out to a file
func (cdre *CdrExporter) WriteToFile(filePath string) error {
	fileOut, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer fileOut.Close()
	switch cdre.cdrFormat {
	case utils.DRYRUN:
		return nil
	case utils.CDRE_FIXED_WIDTH:
		if err := cdre.writeOut(fileOut); err != nil {
			return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
		}
	case utils.CSV:
		csvWriter := csv.NewWriter(fileOut)
		if err := cdre.writeCsv(csvWriter); err != nil {
			return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
		}
	}
	return nil
}

// Return the first exported Cdr OrderId
func (cdre *CdrExporter) FirstOrderId() int64 {
	return cdre.firstExpOrderId
}

// Return the last exported Cdr OrderId
func (cdre *CdrExporter) LastOrderId() int64 {
	return cdre.lastExpOrderId
}

// Return total cost in the exported cdrs
func (cdre *CdrExporter) TotalCost() float64 {
	return cdre.totalCost
}

func (cdre *CdrExporter) TotalExportedCdrs() int {
	return cdre.numberOfRecords
}

// Return successfully exported CgrIds
func (cdre *CdrExporter) PositiveExports() []string {
	return cdre.positiveExports
}

// Return failed exported CgrIds together with the reason
func (cdre *CdrExporter) NegativeExports() map[string]string {
	return cdre.negativeExports
}
