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
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/streadway/amqp"
)

const (
	META_EXPORTID      = "*export_id"
	META_TIMENOW       = "*time_now"
	META_FIRSTCDRATIME = "*first_cdr_atime"
	META_LASTCDRATIME  = "*last_cdr_atime"
	META_NRCDRS        = "*cdrs_number"
	META_DURCDRS       = "*cdrs_duration"
	META_SMSUSAGE      = "*sms_usage"
	META_MMSUSAGE      = "*mms_usage"
	META_GENERICUSAGE  = "*generic_usage"
	META_DATAUSAGE     = "*data_usage"
	META_COSTCDRS      = "*cdrs_cost"
	META_FORMATCOST    = "*format_cost"
)

func NewCDRExporter(cdrs []*CDR, exportTemplate *config.CdreCfg, exportFormat, exportPath, fallbackPath, exportID string,
	synchronous bool, attempts int, fieldSeparator rune, usageMultiplyFactor utils.FieldMultiplyFactor,
	costMultiplyFactor float64, roundingDecimals int, httpSkipTlsCheck bool, httpPoster *HTTPPoster, filterS *FilterS) (*CDRExporter, error) {
	if len(cdrs) == 0 { // Nothing to export
		return nil, nil
	}
	cdre := &CDRExporter{
		cdrs:                cdrs,
		exportTemplate:      exportTemplate,
		exportFormat:        exportFormat,
		exportPath:          exportPath,
		fallbackPath:        fallbackPath,
		exportID:            exportID,
		synchronous:         synchronous,
		attempts:            attempts,
		fieldSeparator:      fieldSeparator,
		usageMultiplyFactor: usageMultiplyFactor,
		costMultiplyFactor:  costMultiplyFactor,
		roundingDecimals:    roundingDecimals,
		httpSkipTlsCheck:    httpSkipTlsCheck,
		httpPoster:          httpPoster,
		negativeExports:     make(map[string]string),
		filterS:             filterS,
	}
	return cdre, nil
}

type CDRExporter struct {
	sync.RWMutex
	cdrs                []*CDR
	exportTemplate      *config.CdreCfg
	exportFormat        string
	exportPath          string
	fallbackPath        string // folder where we save failed CDRs
	exportID            string // Unique identifier or this export
	synchronous         bool
	attempts            int
	fieldSeparator      rune
	usageMultiplyFactor utils.FieldMultiplyFactor
	costMultiplyFactor  float64
	roundingDecimals    int
	httpSkipTlsCheck    bool
	httpPoster          *HTTPPoster

	header, trailer []string   // Header and Trailer fields
	content         [][]string // Rows of cdr fields

	firstCdrATime, lastCdrATime time.Time
	numberOfRecords             int
	totalDuration, totalDataUsage, totalSmsUsage,
	totalMmsUsage, totalGenericUsage time.Duration
	totalCost                       float64
	firstExpOrderId, lastExpOrderId int64
	positiveExports                 []string          // CGRIDs of successfully exported CDRs
	negativeExports                 map[string]string // CGRIDs of failed exports

	filterS *FilterS
}

// Handle various meta functions used in header/trailer
func (cdre *CDRExporter) metaHandler(tag, arg string) (string, error) {
	switch tag {
	case META_EXPORTID:
		return cdre.exportID, nil
	case META_TIMENOW:
		return time.Now().Format(arg), nil
	case META_FIRSTCDRATIME:
		return cdre.firstCdrATime.Format(arg), nil
	case META_LASTCDRATIME:
		return cdre.lastCdrATime.Format(arg), nil
	case META_NRCDRS:
		return strconv.Itoa(cdre.numberOfRecords), nil
	case META_DURCDRS:
		cdr := &CDR{ToR: utils.VOICE, Usage: cdre.totalDuration}
		return cdr.FieldAsString(&config.RSRParser{Rules: "~" + utils.Usage, AllFiltersMatch: true})
	case META_SMSUSAGE:
		cdr := &CDR{ToR: utils.SMS, Usage: cdre.totalDuration}
		return cdr.FieldAsString(&config.RSRParser{Rules: "~" + utils.Usage, AllFiltersMatch: true})
	case META_MMSUSAGE:
		cdr := &CDR{ToR: utils.MMS, Usage: cdre.totalDuration}
		return cdr.FieldAsString(&config.RSRParser{Rules: "~" + utils.Usage, AllFiltersMatch: true})
	case META_GENERICUSAGE:
		cdr := &CDR{ToR: utils.GENERIC, Usage: cdre.totalDuration}
		return cdr.FieldAsString(&config.RSRParser{Rules: "~" + utils.Usage, AllFiltersMatch: true})
	case META_DATAUSAGE:
		cdr := &CDR{ToR: utils.DATA, Usage: cdre.totalDuration}
		return cdr.FieldAsString(&config.RSRParser{Rules: "~" + utils.Usage, AllFiltersMatch: true})
	case META_COSTCDRS:
		return strconv.FormatFloat(utils.Round(cdre.totalCost,
			cdre.roundingDecimals, utils.ROUNDING_MIDDLE), 'f', -1, 64), nil
	default:
		return "", fmt.Errorf("Unsupported METATAG: %s", tag)
	}
}

// Compose and cache the header
func (cdre *CDRExporter) composeHeader() (err error) {
	for _, cfgFld := range cdre.exportTemplate.HeaderFields {
		if len(cfgFld.Filters) != 0 {
			//check filter if pass
		}
		var outVal string
		switch cfgFld.Type {
		case utils.META_FILLER:
			out, err := cfgFld.Value.ParseValue(utils.EmptyString)
			if err != nil {
				return err
			}
			outVal = out
			cfgFld.Padding = "right"
		case utils.META_CONSTANT:
			out, err := cfgFld.Value.ParseValue(utils.EmptyString)
			if err != nil {
				return err
			}
			outVal = out
		case utils.META_HANDLER:
			out, err := cfgFld.Value.ParseValue(utils.EmptyString)
			if err != nil {
				return err
			}
			outVal, err = cdre.metaHandler(out, cfgFld.Layout)
		default:
			return fmt.Errorf("Unsupported field type: %s", cfgFld.Type)
		}
		if err != nil {
			utils.Logger.Err(fmt.Sprintf("<CdreFw> Cannot export CDR header, field %s, error: %s", cfgFld.Tag, err.Error()))
			return err
		}
		fmtOut := outVal
		if fmtOut, err = utils.FmtFieldWidth(cfgFld.Tag, outVal, cfgFld.Width, cfgFld.Strip, cfgFld.Padding, cfgFld.Mandatory); err != nil {
			utils.Logger.Err(fmt.Sprintf("<CdreFw> Cannot export CDR header, field %s, error: %s", cfgFld.Tag, err.Error()))
			return err
		}
		cdre.Lock()
		cdre.header = append(cdre.header, fmtOut)
		cdre.Unlock()
	}
	return nil
}

// Compose and cache the trailer
func (cdre *CDRExporter) composeTrailer() (err error) {
	for _, cfgFld := range cdre.exportTemplate.TrailerFields {
		if len(cfgFld.Filters) != 0 {
			//check filter if pass
		}
		var outVal string
		switch cfgFld.Type {
		case utils.META_FILLER:
			out, err := cfgFld.Value.ParseValue(utils.EmptyString)
			if err != nil {
				return err
			}
			outVal = out
			cfgFld.Padding = "right"
		case utils.META_CONSTANT:
			out, err := cfgFld.Value.ParseValue(utils.EmptyString)
			if err != nil {
				return err
			}
			outVal = out
		case utils.META_HANDLER:
			out, err := cfgFld.Value.ParseValue(utils.EmptyString)
			if err != nil {
				return err
			}
			outVal, err = cdre.metaHandler(out, cfgFld.Layout)
		default:
			return fmt.Errorf("Unsupported field type: %s", cfgFld.Type)
		}
		if err != nil {
			utils.Logger.Err(fmt.Sprintf("<CdreFw> Cannot export CDR trailer, field: %s, error: %s", cfgFld.Tag, err.Error()))
			return err
		}
		fmtOut := outVal
		if fmtOut, err = utils.FmtFieldWidth(cfgFld.Tag, outVal, cfgFld.Width, cfgFld.Strip, cfgFld.Padding, cfgFld.Mandatory); err != nil {
			utils.Logger.Err(fmt.Sprintf("<CdreFw> Cannot export CDR trailer, field: %s, error: %s", cfgFld.Tag, err.Error()))
			return err
		}
		cdre.Lock()
		cdre.trailer = append(cdre.trailer, fmtOut)
		cdre.Unlock()
	}
	return nil
}

func (cdre *CDRExporter) postCdr(cdr *CDR) (err error) {
	var body interface{}
	switch cdre.exportFormat {
	case utils.MetaHTTPjsonCDR, utils.MetaAMQPjsonCDR:
		jsn, err := json.Marshal(cdr)
		if err != nil {
			return err
		}
		body = jsn
	case utils.MetaHTTPjsonMap, utils.MetaAMQPjsonMap:
		expMp, err := cdr.AsExportMap(cdre.exportTemplate.ContentFields, cdre.httpSkipTlsCheck, nil, cdre.roundingDecimals, cdre.filterS)
		if err != nil {
			return err
		}
		jsn, err := json.Marshal(expMp)
		if err != nil {
			return err
		}
		body = jsn
	case utils.META_HTTP_POST:
		expMp, err := cdr.AsExportMap(cdre.exportTemplate.ContentFields, cdre.httpSkipTlsCheck, nil, cdre.roundingDecimals, cdre.filterS)
		if err != nil {
			return err
		}
		vals := url.Values{}
		for fld, val := range expMp {
			vals.Set(fld, val)
		}
		body = vals
	default:
		err = fmt.Errorf("unsupported exportFormat: <%s>", cdre.exportFormat)
	}
	if err != nil {
		return
	}
	// compute fallbackPath
	fallbackPath := utils.META_NONE
	ffn := &utils.FallbackFileName{Module: utils.CDRPoster, Transport: cdre.exportFormat, Address: cdre.exportPath, RequestID: utils.GenUUID()}
	fallbackFileName := ffn.AsString()
	if cdre.fallbackPath != utils.META_NONE { // not none, need fallback
		fallbackPath = path.Join(cdre.fallbackPath, fallbackFileName)
	}
	switch cdre.exportFormat {
	case utils.MetaHTTPjsonCDR, utils.MetaHTTPjsonMap, utils.MetaHTTPjson, utils.META_HTTP_POST:
		_, err = cdre.httpPoster.Post(cdre.exportPath, utils.PosterTransportContentTypes[cdre.exportFormat], body, cdre.attempts, fallbackPath)
	case utils.MetaAMQPjsonCDR, utils.MetaAMQPjsonMap:
		var amqpPoster *AMQPPoster
		amqpPoster, err = AMQPPostersCache.GetAMQPPoster(cdre.exportPath, cdre.attempts, cdre.fallbackPath)
		if err == nil { // error will be checked bellow
			var chn *amqp.Channel
			chn, err = amqpPoster.Post(
				nil, utils.PosterTransportContentTypes[cdre.exportFormat], body.([]byte), fallbackFileName)
			if chn != nil {
				chn.Close()
			}
		}
	}
	return
}

// Write individual cdr into content buffer, build stats
func (cdre *CDRExporter) processCDR(cdr *CDR) (err error) {
	if cdr.ExtraFields == nil { // Avoid assignment in nil map if not initialized
		cdr.ExtraFields = make(map[string]string)
	}
	// Usage multiply, find config based on ToR field or *any
	for _, key := range []string{cdr.ToR, utils.ANY} {
		if uM, hasIt := cdre.usageMultiplyFactor[key]; hasIt && uM != 1.0 {
			cdr.UsageMultiply(uM, cdre.roundingDecimals)
			break
		}
	}
	if cdre.costMultiplyFactor != 0.0 {
		cdr.CostMultiply(cdre.costMultiplyFactor, cdre.roundingDecimals)
	}
	switch cdre.exportFormat {
	case utils.MetaFileFWV, utils.MetaFileCSV:
		var cdrRow []string
		cdrRow, err = cdr.AsExportRecord(cdre.exportTemplate.ContentFields, cdre.httpSkipTlsCheck, cdre.cdrs, cdre.roundingDecimals, cdre.filterS)
		if len(cdrRow) == 0 && err == nil { // No CDR data, most likely no configuration fields defined
			return
		} else {
			cdre.Lock()
			cdre.content = append(cdre.content, cdrRow)
			cdre.Unlock()
		}
	default: // attempt posting CDR
		err = cdre.postCdr(cdr)
	}
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<CDRE> Cannot export CDR with CGRID: %s and runid: %s, error: %s", cdr.CGRID, cdr.RunID, err.Error()))
		return
	}
	// Done with writing content, compute stats here
	cdre.Lock()
	defer cdre.Unlock()
	if cdre.firstCdrATime.IsZero() || cdr.AnswerTime.Before(cdre.firstCdrATime) {
		cdre.firstCdrATime = cdr.AnswerTime
	}
	if cdr.AnswerTime.After(cdre.lastCdrATime) {
		cdre.lastCdrATime = cdr.AnswerTime
	}
	cdre.numberOfRecords += 1
	if cdr.ToR == utils.VOICE { // Only count duration for non data cdrs
		cdre.totalDuration += cdr.Usage
	}
	if cdr.ToR == utils.SMS { // Count usage for SMS
		cdre.totalSmsUsage += cdr.Usage
	}
	if cdr.ToR == utils.MMS { // Count usage for MMS
		cdre.totalMmsUsage += cdr.Usage
	}
	if cdr.ToR == utils.GENERIC { // Count usage for GENERIC
		cdre.totalGenericUsage += cdr.Usage
	}
	if cdr.ToR == utils.DATA { // Count usage for DATA
		cdre.totalDataUsage += cdr.Usage
	}
	if cdr.Cost != -1 {
		cdre.totalCost += cdr.Cost
		cdre.totalCost = utils.Round(cdre.totalCost, cdre.roundingDecimals, utils.ROUNDING_MIDDLE)
	}
	if cdre.firstExpOrderId > cdr.OrderID || cdre.firstExpOrderId == 0 {
		cdre.firstExpOrderId = cdr.OrderID
	}
	if cdre.lastExpOrderId < cdr.OrderID {
		cdre.lastExpOrderId = cdr.OrderID
	}
	return nil
}

// Builds header, content and trailers
func (cdre *CDRExporter) processCDRs() (err error) {
	var wg sync.WaitGroup
	for _, cdr := range cdre.cdrs {
		if cdr == nil || len(cdr.CGRID) == 0 { // CDR needs to exist and it's CGRID needs to be populated
			continue
		}
		if len(cdre.exportTemplate.Filters) != 0 {
			if pass, err := cdre.filterS.Pass(cdre.exportTemplate.Tenant,
				cdre.exportTemplate.Filters, config.NewNavigableMap(cdr.AsMapStringIface())); err != nil || !pass {
				continue // Not passes filters, ignore this CDR
			}
		}
		if cdre.synchronous ||
			utils.IsSliceMember([]string{utils.MetaFileCSV, utils.MetaFileFWV}, cdre.exportFormat) {
			wg.Add(1) // wait for synchronous or file ones since these need to be done before continuing
		}
		go func(cdre *CDRExporter, cdr *CDR) {
			if err := cdre.processCDR(cdr); err != nil {
				cdre.Lock()
				cdre.negativeExports[cdr.CGRID] = err.Error()
				cdre.Unlock()
			} else {
				cdre.Lock()
				cdre.positiveExports = append(cdre.positiveExports, cdr.CGRID)
				cdre.Unlock()
			}
			if cdre.synchronous ||
				utils.IsSliceMember([]string{utils.MetaFileCSV, utils.MetaFileFWV}, cdre.exportFormat) {
				wg.Done()
			}
		}(cdre, cdr)
	}
	wg.Wait()
	// Process header and trailer after processing cdrs since the metatag functions can access stats out of built cdrs
	if cdre.exportTemplate.HeaderFields != nil {
		if err = cdre.composeHeader(); err != nil {
			return
		}
	}
	if cdre.exportTemplate.TrailerFields != nil {
		if err = cdre.composeTrailer(); err != nil {
			return
		}
	}
	return
}

// Simple write method
func (cdre *CDRExporter) writeOut(ioWriter io.Writer) error {
	cdre.Lock()
	defer cdre.Unlock()
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
func (cdre *CDRExporter) writeCsv(csvWriter *csv.Writer) error {
	csvWriter.Comma = cdre.fieldSeparator
	cdre.RLock()
	defer cdre.RUnlock()
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

func (cdre *CDRExporter) ExportCDRs() (err error) {
	if err = cdre.processCDRs(); err != nil {
		return
	}
	if utils.IsSliceMember([]string{utils.MetaFileCSV, utils.MetaFileFWV}, cdre.exportFormat) { // files are written after processing all CDRs
		cdre.RLock()
		contLen := len(cdre.content)
		cdre.RUnlock()
		if contLen == 0 {
			return
		}
		var expFormat string
		switch cdre.exportFormat {
		case utils.MetaFileFWV:
			expFormat = "fwv"
		case utils.MetaFileCSV:
			expFormat = "csv"
		default:
			expFormat = cdre.exportFormat
		}
		expPath := cdre.exportPath
		if len(filepath.Ext(expPath)) == 0 { // verify extension from exportPath (if have extension is file else is directory)
			fileName := fmt.Sprintf("cdre_%s.%s", utils.UUIDSha1Prefix(), expFormat)
			expPath = path.Join(expPath, fileName)
		}
		fileOut, err := os.Create(expPath)
		if err != nil {
			return err
		}
		defer fileOut.Close()
		if cdre.exportFormat == utils.MetaFileCSV {
			return cdre.writeCsv(csv.NewWriter(fileOut))
		}
		return cdre.writeOut(fileOut)
	}
	return
}

// Return the first exported Cdr OrderId
func (cdre *CDRExporter) FirstOrderId() int64 {
	return cdre.firstExpOrderId
}

// Return the last exported Cdr OrderId
func (cdre *CDRExporter) LastOrderId() int64 {
	return cdre.lastExpOrderId
}

// Return total cost in the exported cdrs
func (cdre *CDRExporter) TotalCost() float64 {
	return cdre.totalCost
}

func (cdre *CDRExporter) TotalExportedCdrs() int {
	return cdre.numberOfRecords
}

// Return successfully exported CGRIDs
func (cdre *CDRExporter) PositiveExports() []string {
	cdre.RLock()
	defer cdre.RUnlock()
	return cdre.positiveExports
}

// Return failed exported CGRIDs together with the reason
func (cdre *CDRExporter) NegativeExports() map[string]string {
	cdre.RLock()
	defer cdre.RUnlock()
	return cdre.negativeExports
}
