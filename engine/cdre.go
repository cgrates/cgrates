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
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

const (
	metaExportID      = "*export_id"
	metaTimeNow       = "*time_now"
	metaFirstCDRAtime = "*first_cdr_atime"
	metaLastCDRAtime  = "*last_cdr_atime"
	metaNrCDRs        = "*cdrs_number"
	metaDurCDRs       = "*cdrs_duration"
	metaSMSUsage      = "*sms_usage"
	metaMMSUsage      = "*mms_usage"
	metaGenericUsage  = "*generic_usage"
	metaDataUsage     = "*data_usage"
	metaCostCDRs      = "*cdrs_cost"
)

// NewCDRExporter returns a new CDRExporter
func NewCDRExporter(cdrs []*CDR, exportTemplate *config.CdreCfg, exportFormat, exportPath, fallbackPath, exportID string,
	synchronous bool, attempts int, fieldSeparator rune,
	httpSkipTLSCheck bool, attrsConns []string, filterS *FilterS) (*CDRExporter, error) {
	if len(cdrs) == 0 { // Nothing to export
		return nil, nil
	}
	cdre := &CDRExporter{
		cdrs:             cdrs,
		exportTemplate:   exportTemplate,
		exportFormat:     exportFormat,
		exportPath:       exportPath,
		fallbackPath:     fallbackPath,
		exportID:         exportID,
		synchronous:      synchronous,
		attempts:         attempts,
		fieldSeparator:   fieldSeparator,
		httpSkipTLSCheck: httpSkipTLSCheck,
		negativeExports:  make(map[string]string),
		attrsConns:       attrsConns,
		filterS:          filterS,
	}
	return cdre, nil
}

// CDRExporter used to export the CDRs
type CDRExporter struct {
	sync.RWMutex
	cdrs             []*CDR
	exportTemplate   *config.CdreCfg
	exportFormat     string
	exportPath       string
	fallbackPath     string // folder where we save failed CDRs
	exportID         string // Unique identifier or this export
	synchronous      bool
	attempts         int
	fieldSeparator   rune
	httpSkipTLSCheck bool

	header, trailer []string   // Header and Trailer fields
	content         [][]string // Rows of cdr fields

	firstCdrATime, lastCdrATime time.Time
	numberOfRecords             int
	totalDuration, totalDataUsage, totalSmsUsage,
	totalMmsUsage, totalGenericUsage time.Duration
	totalCost                       float64
	firstExpOrderID, lastExpOrderID int64
	positiveExports                 []string          // CGRIDs of successfully exported CDRs
	negativeExports                 map[string]string // CGRIDs of failed exports

	attrsConns []string
	filterS    *FilterS
}

// Handle various meta functions used in header/trailer
func (cdre *CDRExporter) metaHandler(tag, arg string) (string, error) {
	switch tag {
	case metaExportID:
		return cdre.exportID, nil
	case metaTimeNow:
		return time.Now().Format(arg), nil
	case metaFirstCDRAtime:
		return cdre.firstCdrATime.Format(arg), nil
	case metaLastCDRAtime:
		return cdre.lastCdrATime.Format(arg), nil
	case metaNrCDRs:
		return strconv.Itoa(cdre.numberOfRecords), nil
	case metaDurCDRs:
		cdr := &CDR{ToR: utils.VOICE, Usage: cdre.totalDuration}
		return cdr.FieldAsString(&config.RSRParser{Rules: "~" + utils.Usage, AllFiltersMatch: true})
	case metaSMSUsage:
		cdr := &CDR{ToR: utils.SMS, Usage: cdre.totalDuration}
		return cdr.FieldAsString(&config.RSRParser{Rules: "~" + utils.Usage, AllFiltersMatch: true})
	case metaMMSUsage:
		cdr := &CDR{ToR: utils.MMS, Usage: cdre.totalDuration}
		return cdr.FieldAsString(&config.RSRParser{Rules: "~" + utils.Usage, AllFiltersMatch: true})
	case metaGenericUsage:
		cdr := &CDR{ToR: utils.GENERIC, Usage: cdre.totalDuration}
		return cdr.FieldAsString(&config.RSRParser{Rules: "~" + utils.Usage, AllFiltersMatch: true})
	case metaDataUsage:
		cdr := &CDR{ToR: utils.DATA, Usage: cdre.totalDuration}
		return cdr.FieldAsString(&config.RSRParser{Rules: "~" + utils.Usage, AllFiltersMatch: true})
	case metaCostCDRs:
		return strconv.FormatFloat(utils.Round(cdre.totalCost,
			globalRoundingDecimals, utils.ROUNDING_MIDDLE), 'f', -1, 64), nil
	default:
		return "", fmt.Errorf("Unsupported METATAG: %s", tag)
	}
}

// Compose and cache the header
func (cdre *CDRExporter) composeHeader() (err error) {
	for _, cfgFld := range cdre.exportTemplate.Fields {
		if !strings.HasPrefix(cfgFld.Path, utils.MetaHdr) {
			continue
		}
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
			cfgFld.Padding = utils.MetaRight
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
	for _, cfgFld := range cdre.exportTemplate.Fields {
		if !strings.HasPrefix(cfgFld.Path, utils.MetaTrl) {
			continue
		}
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
			cfgFld.Padding = utils.MetaRight
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
		if body, err = json.Marshal(cdr); err != nil {
			return
		}
	case utils.MetaHTTPjsonMap, utils.MetaAMQPjsonMap, utils.MetaAMQPV1jsonMap, utils.MetaSQSjsonMap, utils.MetaKafkajsonMap, utils.MetaS3jsonMap:
		var expMp map[string]string
		if expMp, err = cdr.AsExportMap(cdre.exportTemplate.Fields, cdre.httpSkipTLSCheck, nil, cdre.filterS); err != nil {
			return
		}
		if body, err = json.Marshal(expMp); err != nil {
			return
		}
	case utils.MetaHTTPPost:
		var expMp map[string]string
		if expMp, err = cdr.AsExportMap(cdre.exportTemplate.Fields, cdre.httpSkipTLSCheck, nil, cdre.filterS); err != nil {
			return
		}
		vals := url.Values{}
		for fld, val := range expMp {
			vals.Set(fld, val)
		}
		body = vals
	default:
		return fmt.Errorf("unsupported exportFormat: <%s>", cdre.exportFormat)
	}
	switch cdre.exportFormat {
	case utils.MetaHTTPjsonCDR, utils.MetaHTTPjsonMap, utils.MetaHTTPjson, utils.MetaHTTPPost:
		var pstr *HTTPPoster
		pstr, err = NewHTTPPoster(config.CgrConfig().GeneralCfg().HttpSkipTlsVerify,
			config.CgrConfig().GeneralCfg().ReplyTimeout, cdre.exportPath,
			utils.PosterTransportContentTypes[cdre.exportFormat], cdre.attempts)
		if err != nil {
			return err
		}
		err = pstr.Post(body, utils.EmptyString)
	case utils.MetaAMQPjsonCDR, utils.MetaAMQPjsonMap:
		err = PostersCache.PostAMQP(cdre.exportPath, cdre.attempts, body.([]byte))
	case utils.MetaAMQPV1jsonMap:
		err = PostersCache.PostAMQPv1(cdre.exportPath, cdre.attempts, body.([]byte))
	case utils.MetaSQSjsonMap:
		err = PostersCache.PostSQS(cdre.exportPath, cdre.attempts, body.([]byte))
	case utils.MetaKafkajsonMap:
		err = PostersCache.PostKafka(cdre.exportPath, cdre.attempts, body.([]byte), utils.ConcatenatedKey(cdr.CGRID, cdr.RunID))
	case utils.MetaS3jsonMap:
		err = PostersCache.PostS3(cdre.exportPath, cdre.attempts, body.([]byte), utils.ConcatenatedKey(cdr.CGRID, cdr.RunID))
	}
	if err != nil && cdre.fallbackPath != utils.META_NONE {
		addFailedPost(cdre.exportPath, cdre.exportFormat, utils.CDRPoster, body)
	}
	return
}

// Write individual cdr into content buffer, build stats
func (cdre *CDRExporter) processCDR(cdr *CDR) (err error) {
	if cdr.ExtraFields == nil { // Avoid assignment in nil map if not initialized
		cdr.ExtraFields = make(map[string]string)
	}
	// send the cdr to be processed by attributeS
	if cdre.exportTemplate.AttributeSContext != utils.EmptyString {
		if len(cdre.attrsConns) == 0 {
			return errors.New("no connection to AttributeS")
		}
		cdrEv := cdr.AsCGREvent()
		args := &AttrArgsProcessEvent{
			Context: utils.StringPointer(utils.FirstNonEmpty(
				utils.IfaceAsString(cdrEv.Event[utils.Context]),
				cdre.exportTemplate.AttributeSContext)),
			CGREvent: cdrEv,
		}
		var evReply AttrSProcessEventReply
		if err = connMgr.Call(cdre.attrsConns, nil,
			utils.AttributeSv1ProcessEvent,
			args, &evReply); err != nil {
			return
		}
		if len(evReply.AlteredFields) != 0 {
			if err = cdr.UpdateFromCGREvent(evReply.CGREvent, evReply.AlteredFields); err != nil {
				return
			}
		}
	}

	switch cdre.exportFormat {
	case utils.MetaFileFWV, utils.MetaFileCSV:
		var cdrRow []string
		cdrRow, err = cdr.AsExportRecord(cdre.exportTemplate.Fields, cdre.httpSkipTLSCheck, cdre.cdrs, cdre.filterS)
		if len(cdrRow) == 0 && err == nil { // No CDR data, most likely no configuration fields defined
			return
		}
		cdre.Lock()
		cdre.content = append(cdre.content, cdrRow)
		cdre.Unlock()
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
	cdre.numberOfRecords++
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
		cdre.totalCost = utils.Round(cdre.totalCost, globalRoundingDecimals, utils.ROUNDING_MIDDLE)
	}
	if cdre.firstExpOrderID > cdr.OrderID || cdre.firstExpOrderID == 0 {
		cdre.firstExpOrderID = cdr.OrderID
	}
	if cdre.lastExpOrderID < cdr.OrderID {
		cdre.lastExpOrderID = cdr.OrderID
	}
	return
}

// processCDRs proccess every cdr
func (cdre *CDRExporter) processCDRs() (err error) {
	var wg sync.WaitGroup
	isSync := cdre.exportTemplate.Synchronous ||
		utils.SliceHasMember([]string{utils.MetaFileCSV, utils.MetaFileFWV}, cdre.exportTemplate.ExportFormat)
	for _, cdr := range cdre.cdrs {
		if cdr == nil || len(cdr.CGRID) == 0 { // CDR needs to exist and it's CGRID needs to be populated
			continue
		}
		if len(cdre.exportTemplate.Filters) != 0 {
			cgrDp := cdr.AsMapStorage()
			if pass, err := cdre.filterS.Pass(cdre.exportTemplate.Tenant,
				cdre.exportTemplate.Filters, cgrDp); err != nil || !pass {
				continue // Not passes filters, ignore this CDR
			}
		}
		if isSync {
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
			if isSync {
				wg.Done()
			}
		}(cdre, cdr)
	}
	wg.Wait()
	return
}

// Simple write method
func (cdre *CDRExporter) writeOut(ioWriter io.Writer) (err error) {
	cdre.Lock()
	defer cdre.Unlock()
	if len(cdre.header) != 0 {
		for _, fld := range append(cdre.header, "\n") {
			if _, err = io.WriteString(ioWriter, fld); err != nil {
				return
			}
		}
	}
	for _, cdrContent := range cdre.content {
		for _, cdrFld := range append(cdrContent, "\n") {
			if _, err = io.WriteString(ioWriter, cdrFld); err != nil {
				return
			}
		}
	}
	if len(cdre.trailer) != 0 {
		for _, fld := range append(cdre.trailer, "\n") {
			if _, err = io.WriteString(ioWriter, fld); err != nil {
				return
			}
		}
	}
	return
}

// csvWriter specific method
func (cdre *CDRExporter) writeCsv(csvWriter *csv.Writer) (err error) {
	csvWriter.Comma = cdre.fieldSeparator
	cdre.RLock()
	defer cdre.RUnlock()
	if len(cdre.header) != 0 {
		if err = csvWriter.Write(cdre.header); err != nil {
			return
		}
	}
	for _, cdrContent := range cdre.content {
		if err = csvWriter.Write(cdrContent); err != nil {
			return
		}
	}
	if len(cdre.trailer) != 0 {
		if err = csvWriter.Write(cdre.trailer); err != nil {
			return
		}
	}
	csvWriter.Flush()
	return
}

// ExportCDRs exports the given CDRs
func (cdre *CDRExporter) ExportCDRs() (err error) {
	if err = cdre.processCDRs(); err != nil {
		return
	}
	if utils.SliceHasMember([]string{utils.MetaFileCSV, utils.MetaFileFWV}, cdre.exportFormat) { // files are written after processing all CDRs
		cdre.RLock()
		contLen := len(cdre.content)
		cdre.RUnlock()
		if contLen == 0 {
			return
		}
		if err = cdre.composeHeader(); err != nil {
			return
		}
		if err = cdre.composeTrailer(); err != nil {
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
		var fileOut *os.File
		if fileOut, err = os.Create(expPath); err != nil {
			return
		}
		defer fileOut.Close()
		if cdre.exportFormat == utils.MetaFileCSV {
			return cdre.writeCsv(csv.NewWriter(fileOut))
		}
		return cdre.writeOut(fileOut)
	}
	return
}

// FirstOrderID returns the first exported Cdr OrderId
func (cdre *CDRExporter) FirstOrderID() int64 {
	return cdre.firstExpOrderID
}

// LastOrderID return the last exported Cdr OrderId
func (cdre *CDRExporter) LastOrderID() int64 {
	return cdre.lastExpOrderID
}

// TotalCost returns the total cost in the exported cdrs
func (cdre *CDRExporter) TotalCost() float64 {
	return cdre.totalCost
}

// TotalExportedCdrs returns the number of exported CDRs
func (cdre *CDRExporter) TotalExportedCdrs() int {
	return cdre.numberOfRecords
}

// PositiveExports returns the successfully exported CGRIDs
func (cdre *CDRExporter) PositiveExports() []string {
	cdre.RLock()
	defer cdre.RUnlock()
	return cdre.positiveExports
}

// NegativeExports returns the failed exported CGRIDs together with the reason
func (cdre *CDRExporter) NegativeExports() map[string]string {
	cdre.RLock()
	defer cdre.RUnlock()
	return cdre.negativeExports
}
