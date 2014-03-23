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
	"strings"
	"time"
)

const (
	FILLER                = "filler"
	CONSTANT              = "constant"
	CDRFIELD              = "cdrfield"
	CONCATENATED_CDRFIELD = "concatenated_cdrfield"
)

type FixedWidthCdrWriter struct {
	logDb                     engine.LogStorage // Used to extract cost_details if these are requested
	writer                    io.Writer
	exportTemplate            *config.CgrXmlCdreFwCfg
	roundDecimals             int
	header, content, trailer  *bytes.Buffer
	firstCdrTime, lastCdrTime time.Time
	numberOfRecords           int
	totalDuration             time.Duration
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
	case utils.COST_DETAILS: // Special case when we need to further extract cost_details out of logDb
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

// Writes the header into it's buffer
func (fww *FixedWidthCdrWriter) ComposeHeader() error {
	return nil
}

// Writes the trailer into it's buffer
func (fww *FixedWidthCdrWriter) ComposeTrailer() error {
	return nil
}

// Write individual cdr into content buffer, build stats
func (fww *FixedWidthCdrWriter) WriteCdr(cdr *utils.StoredCdr) error {
	var err error
	cdrRow := ""
	for _, cfgFld := range fww.exportTemplate.Content.Fields {
		var outVal string
		switch cfgFld.Type {
		case FILLER, CONSTANT:
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
			engine.Logger.Err(fmt.Sprintf("<CdreFW> Cannot export CDR with cgrid: %s and runid: %s, error: %s", cdr.CgrId, cdr.MediationRunId))
			return err
		}
		if fmtOut, err := FmtFieldWidth(outVal, cfgFld.Width, cfgFld.Strip, cfgFld.Padding); err != nil {
			engine.Logger.Err(fmt.Sprintf("<CdreFW> Cannot export CDR with cgrid: %s and runid: %s, error: %s", cdr.CgrId, cdr.MediationRunId))
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
	return nil
}

func (fww *FixedWidthCdrWriter) Close() {
	fww.ComposeHeader()
	fww.ComposeTrailer()
	for _, buf := range []*bytes.Buffer{fww.header, fww.content, fww.trailer} {
		fww.writer.Write(buf.Bytes())
	}
}
