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
	"errors"
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
	META_FIRSTCDRTIME     = "first_cdr_time"
	META_LASTCDRTIME      = "last_cdr_time"
	META_NRCDRS           = "cdrs_number"
	META_DURCDRS          = "cdrs_duration"
	META_COSTCDRS         = "cdrs_cost"
)

var err error

func NewFWCdrWriter(logDb engine.LogStorage, outFile *os.File, exportTpl *config.CgrXmlCdreFwCfg, exportId string, roundDecimals int) (*FixedWidthCdrWriter, error) {
	return &FixedWidthCdrWriter{
		logDb:          logDb,
		writer:         outFile,
		exportTemplate: exportTpl,
		exportId:       exportId,
		roundDecimals:  roundDecimals,
		header:         &bytes.Buffer{},
		content:        &bytes.Buffer{},
		trailer:        &bytes.Buffer{}}, nil
}

type FixedWidthCdrWriter struct {
	logDb                     engine.LogStorage // Used to extract cost_details if these are requested
	writer                    io.Writer
	exportTemplate            *config.CgrXmlCdreFwCfg
	exportId                  string // Unique identifier or this export
	roundDecimals             int
	header, content, trailer  *bytes.Buffer
	firstCdrTime, lastCdrTime time.Time
	numberOfRecords           int
	totalDuration             time.Duration
	totalCost                 float64
}

// Return Json marshaled callCost attached to
// Keep it separately so we test only this part in local tests
func (fww *FixedWidthCdrWriter) getCdrCostDetails(cgrId, runId string) (string, error) {
	cc, err := fww.logDb.GetCallCostLog(cgrId, "", runId)
	if err != nil {
		return "", err
	} else if cc == nil {
		return "", nil
	}
	ccJson, _ := json.Marshal(cc)
	return string(ccJson), nil
}

// Extracts the value specified by cfgHdr out of cdr
func (fww *FixedWidthCdrWriter) cdrFieldValue(cdr *utils.StoredCdr, cfgHdr, layout string) (string, error) {
	rsrField, err := utils.NewRSRField(cfgHdr)
	if err != nil {
		return "", err
	} else if rsrField == nil {
		return "", nil
	}
	var cdrVal string
	switch rsrField.Id {
	case COST_DETAILS: // Special case when we need to further extract cost_details out of logDb
		if cdrVal, err = fww.getCdrCostDetails(cdr.CgrId, cdr.MediationRunId); err != nil {
			return "", err
		}
	case utils.COST:
		cdrVal = cdr.FormatCost(fww.roundDecimals)
	case utils.SETUP_TIME:
		cdrVal = cdr.SetupTime.Format(layout)
	case utils.ANSWER_TIME: // Format time based on layout
		cdrVal = cdr.AnswerTime.Format(layout)
	default:
		cdrVal = cdr.ExportFieldValue(rsrField.Id)
	}
	return rsrField.ParseValue(cdrVal), nil
}

func (fww *FixedWidthCdrWriter) metaHandler(tag, layout string) (string, error) {
	switch tag {
	case META_EXPORTID:
		return fww.exportId, nil
	case META_TIMENOW:
		return time.Now().Format(layout), nil
	case META_FIRSTCDRTIME:
		return fww.firstCdrTime.Format(layout), nil
	case META_LASTCDRTIME:
		return fww.lastCdrTime.Format(layout), nil
	case META_NRCDRS:
		return strconv.Itoa(fww.numberOfRecords), nil
	case META_DURCDRS:
		return strconv.FormatFloat(fww.totalDuration.Seconds(), 'f', -1, 64), nil
	case META_COSTCDRS:
		return strconv.FormatFloat(utils.Round(fww.totalCost, fww.roundDecimals, utils.ROUNDING_MIDDLE), 'f', -1, 64), nil
	default:
		return "", errors.New("Unsupported METATAG")
	}
	return "", nil
}

// Writes the header into it's buffer
func (fww *FixedWidthCdrWriter) ComposeHeader() error {
	header := ""
	for _, cfgFld := range fww.exportTemplate.Header.Fields {
		var outVal string
		switch cfgFld.Type {
		case FILLER:
			outVal = cfgFld.Value
			cfgFld.Padding = "right"
		case CONSTANT:
			outVal = cfgFld.Value
		case METATAG:
			outVal, err = fww.metaHandler(cfgFld.Value, cfgFld.Layout)
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
	fww.header.WriteString(header)
	return nil
}

// Writes the trailer into it's buffer
func (fww *FixedWidthCdrWriter) ComposeTrailer() error {
	trailer := ""
	for _, cfgFld := range fww.exportTemplate.Trailer.Fields {
		var outVal string
		switch cfgFld.Type {
		case FILLER:
			outVal = cfgFld.Value
			cfgFld.Padding = "right"
		case CONSTANT:
			outVal = cfgFld.Value
		case METATAG:
			outVal, err = fww.metaHandler(cfgFld.Value, cfgFld.Layout)
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
	fww.trailer.WriteString(trailer)
	return nil
}

// Write individual cdr into content buffer, build stats
func (fww *FixedWidthCdrWriter) WriteCdr(cdr *utils.StoredCdr) error {
	if cdr == nil || len(cdr.CgrId) == 0 { // We do not export empty CDRs
		return nil
	}
	var err error
	cdrRow := ""
	for _, cfgFld := range fww.exportTemplate.Content.Fields {
		var outVal string
		switch cfgFld.Type {
		case FILLER:
			outVal = cfgFld.Value
			cfgFld.Padding = "right"
		case CONSTANT:
			outVal = cfgFld.Value
		case CDRFIELD:
			outVal, err = fww.cdrFieldValue(cdr, cfgFld.Value, cfgFld.Layout)
		case CONCATENATED_CDRFIELD:
			for _, fld := range strings.Split(cfgFld.Value, ",") {
				if fldOut, err := fww.cdrFieldValue(cdr, fld, cfgFld.Layout); err != nil {
					break // The error will be reported bellow
				} else {
					outVal += fldOut
				}
			}
		}
		if err != nil {
			engine.Logger.Err(fmt.Sprintf("<CdreFw> Cannot export CDR with cgrid: %s and runid: %s, error: %s", cdr.CgrId, cdr.MediationRunId, err.Error()))
			return err
		}
		if fmtOut, err := FmtFieldWidth(outVal, cfgFld.Width, cfgFld.Strip, cfgFld.Padding, cfgFld.Mandatory); err != nil {
			engine.Logger.Err(fmt.Sprintf("<CdreFw> Cannot export CDR with cgrid: %s and runid: %s, error: %s", cdr.CgrId, cdr.MediationRunId, err.Error()))
			return err
		} else {
			cdrRow += fmtOut
		}
	}
	if len(cdrRow) == 0 { // No CDR data, most likely no configuration fields defined
		return nil
	}
	cdrRow += "\n" // Done with cdr, postpend new line char
	fww.content.WriteString(cdrRow)
	// Done with writing content, compute stats here
	if fww.firstCdrTime.IsZero() || cdr.SetupTime.Before(fww.firstCdrTime) {
		fww.firstCdrTime = cdr.SetupTime
	}
	if cdr.SetupTime.After(fww.lastCdrTime) {
		fww.lastCdrTime = cdr.SetupTime
	}
	fww.numberOfRecords += 1
	fww.totalDuration += cdr.Duration
	fww.totalCost += cdr.Cost
	fww.totalCost = utils.Round(fww.totalCost, fww.roundDecimals, utils.ROUNDING_MIDDLE)
	return nil
}

func (fww *FixedWidthCdrWriter) Close() {
	if fww.exportTemplate.Header != nil {
		fww.ComposeHeader()
	}
	if fww.exportTemplate.Trailer != nil {
		fww.ComposeTrailer()
	}
	for _, buf := range []*bytes.Buffer{fww.header, fww.content, fww.trailer} {
		fww.writer.Write(buf.Bytes())
	}
}
