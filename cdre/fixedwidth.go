/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	COST_DETAILS          = "cost_details"
	FILLER                = "filler"
	CONSTANT              = "constant"
	CDRFIELD              = "cdrfield"
	METATAG               = "metatag"
	CONCATENATED_CDRFIELD = "concatenated_cdrfield"
	META_EXPORTID         = "export_id"
	META_TIMENOW          = "time_now"
	META_FIRSTCDRATIME    = "first_cdr_atime"
	META_LASTCDRATIME     = "last_cdr_atime"
	META_NRCDRS           = "cdrs_number"
	META_DURCDRS          = "cdrs_duration"
	META_COSTCDRS         = "cdrs_cost"
	META_MASKDESTINATION  = "mask_destination"
)

var err error

func NewFWCdrWriter(logDb engine.LogStorage, outFile *os.File, exportTpl *config.CgrXmlCdreFwCfg, exportId string,
	roundDecimals int, maskDestId string, maskLen int) (*FixedWidthCdrWriter, error) {
	return &FixedWidthCdrWriter{
		logDb:          logDb,
		writer:         outFile,
		exportTemplate: exportTpl,
		exportId:       exportId,
		roundDecimals:  roundDecimals,
		maskDestId:     maskDestId,
		maskLen:        maskLen,
		header:         &bytes.Buffer{},
		content:        &bytes.Buffer{},
		trailer:        &bytes.Buffer{}}, nil
}

type FixedWidthCdrWriter struct {
	logDb                       engine.LogStorage // Used to extract cost_details if these are requested
	writer                      io.Writer
	exportTemplate              *config.CgrXmlCdreFwCfg
	exportId                    string // Unique identifier or this export
	roundDecimals               int
	maskDestId                  string
	maskLen                     int
	header, content, trailer    *bytes.Buffer
	firstCdrATime, lastCdrATime time.Time
	numberOfRecords             int
	totalDuration               time.Duration
	totalCost                   float64
}

// Return Json marshaled callCost attached to
// Keep it separately so we test only this part in local tests
func (fwv *FixedWidthCdrWriter) getCdrCostDetails(cgrId, runId string) (string, error) {
	cc, err := fwv.logDb.GetCallCostLog(cgrId, "", runId)
	if err != nil {
		return "", err
	} else if cc == nil {
		return "", nil
	}
	ccJson, _ := json.Marshal(cc)
	return string(ccJson), nil
}

func (fwv *FixedWidthCdrWriter) maskedDestination(destination string) bool {
	if len(fwv.maskDestId) != 0 && engine.CachedDestHasPrefix(fwv.maskDestId, destination) {
		return true
	}
	return false
}

// Extracts the value specified by cfgHdr out of cdr
func (fwv *FixedWidthCdrWriter) cdrFieldValue(cdr *utils.StoredCdr, cfgHdr, layout string) (string, error) {
	rsrField, err := utils.NewRSRField(cfgHdr)
	if err != nil {
		return "", err
	} else if rsrField == nil {
		return "", nil
	}
	var cdrVal string
	switch rsrField.Id {
	case COST_DETAILS: // Special case when we need to further extract cost_details out of logDb
		if cdrVal, err = fwv.getCdrCostDetails(cdr.CgrId, cdr.MediationRunId); err != nil {
			return "", err
		}
	case utils.COST:
		cdrVal = cdr.FormatCost(fwv.roundDecimals)
	case utils.SETUP_TIME:
		cdrVal = cdr.SetupTime.Format(layout)
	case utils.ANSWER_TIME: // Format time based on layout
		cdrVal = cdr.AnswerTime.Format(layout)
	case utils.DESTINATION:
		cdrVal = cdr.ExportFieldValue(utils.DESTINATION)
		if fwv.maskLen != -1 && fwv.maskedDestination(cdrVal) {
			cdrVal = MaskDestination(cdrVal, fwv.maskLen)
		}
	default:
		cdrVal = cdr.ExportFieldValue(rsrField.Id)
	}
	return rsrField.ParseValue(cdrVal), nil
}

func (fwv *FixedWidthCdrWriter) metaHandler(tag, arg string) (string, error) {
	switch tag {
	case META_EXPORTID:
		return fwv.exportId, nil
	case META_TIMENOW:
		return time.Now().Format(arg), nil
	case META_FIRSTCDRATIME:
		return fwv.firstCdrATime.Format(arg), nil
	case META_LASTCDRATIME:
		return fwv.lastCdrATime.Format(arg), nil
	case META_NRCDRS:
		return strconv.Itoa(fwv.numberOfRecords), nil
	case META_DURCDRS:
		return strconv.FormatFloat(fwv.totalDuration.Seconds(), 'f', -1, 64), nil
	case META_COSTCDRS:
		return strconv.FormatFloat(utils.Round(fwv.totalCost, fwv.roundDecimals, utils.ROUNDING_MIDDLE), 'f', -1, 64), nil
	case META_MASKDESTINATION:
		if fwv.maskedDestination(arg) {
			return "1", nil
		}
		return "0", nil
	default:
		return "", fmt.Errorf("Unsupported METATAG: %s", tag)
	}
	return "", nil
}

// Writes the header into it's buffer
func (fwv *FixedWidthCdrWriter) ComposeHeader() error {
	header := ""
	for _, cfgFld := range fwv.exportTemplate.Header.Fields {
		var outVal string
		switch cfgFld.Type {
		case FILLER:
			outVal = cfgFld.Value
			cfgFld.Padding = "right"
		case CONSTANT:
			outVal = cfgFld.Value
		case METATAG:
			outVal, err = fwv.metaHandler(cfgFld.Value, cfgFld.Layout)
		default:
			return fmt.Errorf("Unsupported field type: %s", cfgFld.Type)
		}
		if err != nil {
			engine.Logger.Err(fmt.Sprintf("<CdreFw> Cannot export CDR header, error: %s", err.Error()))
			return err
		}
		if fmtOut, err := FmtFieldWidth(outVal, cfgFld.Width, cfgFld.Strip, cfgFld.Padding, cfgFld.Mandatory); err != nil {
			engine.Logger.Err(fmt.Sprintf("<CdreFw> Cannot export CDR header, error: %s", err.Error()))
			return err
		} else {
			header += fmtOut
		}
	}
	if len(header) == 0 { // No header data, most likely no configuration fields defined
		return nil
	}
	header += "\n" // Done with cdr, postpend new line char
	fwv.header.WriteString(header)
	return nil
}

// Writes the trailer into it's buffer
func (fwv *FixedWidthCdrWriter) ComposeTrailer() error {
	trailer := ""
	for _, cfgFld := range fwv.exportTemplate.Trailer.Fields {
		var outVal string
		switch cfgFld.Type {
		case FILLER:
			outVal = cfgFld.Value
			cfgFld.Padding = "right"
		case CONSTANT:
			outVal = cfgFld.Value
		case METATAG:
			outVal, err = fwv.metaHandler(cfgFld.Value, cfgFld.Layout)
		default:
			return fmt.Errorf("Unsupported field type: %s", cfgFld.Type)
		}
		if err != nil {
			engine.Logger.Err(fmt.Sprintf("<CdreFw> Cannot export CDR trailer, error: %s", err.Error()))
			return err
		}
		if fmtOut, err := FmtFieldWidth(outVal, cfgFld.Width, cfgFld.Strip, cfgFld.Padding, cfgFld.Mandatory); err != nil {
			engine.Logger.Err(fmt.Sprintf("<CdreFw> Cannot export CDR trailer, error: %s", err.Error()))
			return err
		} else {
			trailer += fmtOut
		}
	}
	if len(trailer) == 0 { // No header data, most likely no configuration fields defined
		return nil
	}
	trailer += "\n" // Done with cdr, postpend new line char
	fwv.trailer.WriteString(trailer)
	return nil
}

// Write individual cdr into content buffer, build stats
func (fwv *FixedWidthCdrWriter) WriteCdr(cdr *utils.StoredCdr) error {
	if cdr == nil || len(cdr.CgrId) == 0 { // We do not export empty CDRs
		return nil
	}
	var err error
	cdrRow := ""
	for _, cfgFld := range fwv.exportTemplate.Content.Fields {
		var outVal string
		switch cfgFld.Type {
		case FILLER:
			outVal = cfgFld.Value
			cfgFld.Padding = "right"
		case CONSTANT:
			outVal = cfgFld.Value
		case CDRFIELD:
			outVal, err = fwv.cdrFieldValue(cdr, cfgFld.Value, cfgFld.Layout)
		case CONCATENATED_CDRFIELD:
			for _, fld := range strings.Split(cfgFld.Value, ",") {
				if fldOut, err := fwv.cdrFieldValue(cdr, fld, cfgFld.Layout); err != nil {
					break // The error will be reported bellow
				} else {
					outVal += fldOut
				}
			}
		case METATAG:
			outVal, err = fwv.metaHandler(cfgFld.Value, cfgFld.Layout)
		}
		if err != nil {
			engine.Logger.Err(fmt.Sprintf("<CdreFw> Cannot export CDR with cgrid: %s and runid: %s, error: %s", cdr.CgrId, cdr.MediationRunId, err.Error()))
			return err
		}
		if fmtOut, err := FmtFieldWidth(outVal, cfgFld.Width, cfgFld.Strip, cfgFld.Padding, cfgFld.Mandatory); err != nil {
			engine.Logger.Err(fmt.Sprintf("<CdreFw> Cannot export CDR with cgrid: %s, runid: %s, fieldName: %s, fieldValue: %s, error: %s", cdr.CgrId, cdr.MediationRunId, cfgFld.Name, outVal, err.Error()))
			return err
		} else {
			cdrRow += fmtOut
		}
	}
	if len(cdrRow) == 0 { // No CDR data, most likely no configuration fields defined
		return nil
	}
	cdrRow += "\n" // Done with cdr, postpend new line char
	fwv.content.WriteString(cdrRow)
	// Done with writing content, compute stats here
	if fwv.firstCdrATime.IsZero() || cdr.AnswerTime.Before(fwv.firstCdrATime) {
		fwv.firstCdrATime = cdr.AnswerTime
	}
	if cdr.AnswerTime.After(fwv.lastCdrATime) {
		fwv.lastCdrATime = cdr.AnswerTime
	}
	fwv.numberOfRecords += 1
	fwv.totalDuration += cdr.Duration
	fwv.totalCost += cdr.Cost
	fwv.totalCost = utils.Round(fwv.totalCost, fwv.roundDecimals, utils.ROUNDING_MIDDLE)
	return nil
}

func (fwv *FixedWidthCdrWriter) Close() {
	if fwv.exportTemplate.Header != nil {
		fwv.ComposeHeader()
	}
	if fwv.exportTemplate.Trailer != nil {
		fwv.ComposeTrailer()
	}
	for _, buf := range []*bytes.Buffer{fwv.header, fwv.content, fwv.trailer} {
		fwv.writer.Write(buf.Bytes())
	}
}
