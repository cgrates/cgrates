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

package cdrc

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/ChrisTrenkamp/goxpath"
	"github.com/ChrisTrenkamp/goxpath/tree"
	"github.com/ChrisTrenkamp/goxpath/tree/xmltree"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// getElementText will process the node to extract the elementName's text out of it (only first one found)
// returns utils.ErrNotFound if the element is not found in the node
func elementText(xmlRes tree.Res, elmntPath string) (string, error) {
	xp, err := goxpath.Parse(elmntPath)
	if err != nil {
		return "", err
	}
	elmntBuf := bytes.NewBufferString(xml.Header)
	if err := goxpath.Marshal(xmlRes.(tree.Node), elmntBuf); err != nil {
		return "", err
	}
	elmntNode, err := xmltree.ParseXML(elmntBuf)
	if err != nil {
		return "", err
	}
	elmnts, err := goxpath.Exec(xp, elmntNode, nil)
	if err != nil {
		return "", err
	}
	if len(elmnts) == 0 {
		return "", utils.ErrNotFound
	}
	return elmnts[0].String(), nil
}

// handlerUsageDiff will calculate the usage as difference between timeEnd and timeStart
// Expects the 2 arguments in template separated by |
func handlerSubstractUsage(xmlElmnt tree.Res, argsTpl utils.RSRFields, cdrPath utils.HierarchyPath, timezone string) (time.Duration, error) {
	var argsStr string
	for _, rsrArg := range argsTpl {
		if rsrArg.Id == utils.HandlerArgSep {
			argsStr += rsrArg.Id
			continue
		}
		absolutePath := utils.ParseHierarchyPath(rsrArg.Id, "")
		relPath := utils.HierarchyPath(absolutePath[len(cdrPath)-1:]) // Need relative path to the xmlElmnt
		argStr, _ := elementText(xmlElmnt, relPath.AsString("/", true))
		argsStr += argStr
	}
	handlerArgs := strings.Split(argsStr, utils.HandlerArgSep)
	if len(handlerArgs) != 2 {
		return time.Duration(0), errors.New("Unexpected number of arguments")
	}
	tEnd, err := utils.ParseTimeDetectLayout(handlerArgs[0], timezone)
	if err != nil {
		return time.Duration(0), err
	}
	tStart, err := utils.ParseTimeDetectLayout(handlerArgs[1], timezone)
	if err != nil {
		return time.Duration(0), err
	}
	return tEnd.Sub(tStart), nil
}

func NewXMLRecordsProcessor(recordsReader io.Reader, cdrPath utils.HierarchyPath, timezone string, httpSkipTlsCheck bool, cdrcCfgs []*config.CdrcConfig) (*XMLRecordsProcessor, error) {
	xp, err := goxpath.Parse(cdrPath.AsString("/", true))
	if err != nil {
		return nil, err
	}
	optsNotStrict := func(s *xmltree.ParseOptions) {
		s.Strict = false
	}
	xmlNode, err := xmltree.ParseXML(recordsReader, optsNotStrict)
	if err != nil {
		return nil, err
	}
	xmlProc := &XMLRecordsProcessor{cdrPath: cdrPath, timezone: timezone, httpSkipTlsCheck: httpSkipTlsCheck, cdrcCfgs: cdrcCfgs}
	xmlProc.cdrXmlElmts = goxpath.MustExec(xp, xmlNode, nil)
	return xmlProc, nil
}

type XMLRecordsProcessor struct {
	cdrXmlElmts      []tree.Res          // result of splitting the XML doc into CDR elements
	procItems        int                 // current number of processed records from file
	cdrPath          utils.HierarchyPath // path towards one CDR element
	timezone         string
	httpSkipTlsCheck bool
	cdrcCfgs         []*config.CdrcConfig // individual configs for the folder CDRC is monitoring

}

func (xmlProc *XMLRecordsProcessor) ProcessedRecordsNr() int64 {
	return int64(xmlProc.procItems)
}

func (xmlProc *XMLRecordsProcessor) ProcessNextRecord() (cdrs []*engine.CDR, err error) {
	if len(xmlProc.cdrXmlElmts) <= xmlProc.procItems {
		return nil, io.EOF // have processed all items
	}
	cdrs = make([]*engine.CDR, 0)
	cdrXML := xmlProc.cdrXmlElmts[xmlProc.procItems]
	xmlProc.procItems += 1
	for _, cdrcCfg := range xmlProc.cdrcCfgs {
		filtersPassing := true
		for _, rsrFltr := range cdrcCfg.CdrFilter {
			if rsrFltr == nil {
				continue // Pass
			}
			fieldVal, _ := elementText(cdrXML, rsrFltr.Id)
			if !rsrFltr.FilterPasses(fieldVal) {
				filtersPassing = false
				break
			}
		}
		if !filtersPassing {
			continue
		}
		if cdr, err := xmlProc.recordToCDR(cdrXML, cdrcCfg); err != nil {
			return nil, fmt.Errorf("<CDRC> Failed converting to CDR, error: %s", err.Error())
		} else {
			cdrs = append(cdrs, cdr)
		}
		if !cdrcCfg.ContinueOnSuccess {
			break
		}
	}
	return cdrs, nil
}

func (xmlProc *XMLRecordsProcessor) recordToCDR(xmlEntity tree.Res, cdrcCfg *config.CdrcConfig) (*engine.CDR, error) {
	cdr := &engine.CDR{OriginHost: "0.0.0.0", Source: cdrcCfg.CdrSourceId, ExtraFields: make(map[string]string), Cost: -1}
	var lazyHttpFields []*config.CfgCdrField
	var err error
	for _, cdrFldCfg := range cdrcCfg.ContentFields {
		var fieldVal string
		if cdrFldCfg.Type == utils.META_COMPOSED {
			for _, cfgFieldRSR := range cdrFldCfg.Value {
				if cfgFieldRSR.IsStatic() {
					fieldVal += cfgFieldRSR.ParseValue("")
				} else { // Dynamic value extracted using path
					absolutePath := utils.ParseHierarchyPath(cfgFieldRSR.Id, "")
					relPath := utils.HierarchyPath(absolutePath[len(xmlProc.cdrPath)-1:]) // Need relative path to the xmlElmnt
					if fieldVal, err := elementText(xmlEntity, relPath.AsString("/", true)); err != nil {
						return nil, fmt.Errorf("Ignoring record: %v - cannot extract field %s, err: %s", xmlEntity, cdrFldCfg.Tag, err.Error())
					} else {
						fieldVal += cfgFieldRSR.ParseValue(fieldVal)
					}
				}
			}
		} else if cdrFldCfg.Type == utils.META_HTTP_POST {
			lazyHttpFields = append(lazyHttpFields, cdrFldCfg) // Will process later so we can send an estimation of cdr to http server
		} else if cdrFldCfg.Type == utils.META_HANDLER && cdrFldCfg.HandlerId == utils.HandlerSubstractUsage {
			usage, err := handlerSubstractUsage(xmlEntity, cdrFldCfg.Value, xmlProc.cdrPath, xmlProc.timezone)
			if err != nil {
				return nil, fmt.Errorf("Ignoring record: %v - cannot extract field %s, err: %s", xmlEntity, cdrFldCfg.Tag, err.Error())
			}
			fieldVal += strconv.FormatFloat(usage.Seconds(), 'f', -1, 64)
		} else {
			return nil, fmt.Errorf("Unsupported field type: %s", cdrFldCfg.Type)
		}
		if err := cdr.ParseFieldValue(cdrFldCfg.FieldId, fieldVal, xmlProc.timezone); err != nil {
			return nil, err
		}
	}
	cdr.CGRID = utils.Sha1(cdr.OriginID, cdr.SetupTime.UTC().String())
	if cdr.ToR == utils.DATA && cdrcCfg.DataUsageMultiplyFactor != 0 {
		cdr.Usage = time.Duration(float64(cdr.Usage.Nanoseconds()) * cdrcCfg.DataUsageMultiplyFactor)
	}
	for _, httpFieldCfg := range lazyHttpFields { // Lazy process the http fields
		var outValByte []byte
		var fieldVal, httpAddr string
		for _, rsrFld := range httpFieldCfg.Value {
			httpAddr += rsrFld.ParseValue("")
		}
		var jsn []byte
		jsn, err = json.Marshal(cdr)
		if err != nil {
			return nil, err
		}
		if outValByte, err = utils.HttpJsonPost(httpAddr, xmlProc.httpSkipTlsCheck, jsn); err != nil && httpFieldCfg.Mandatory {
			return nil, err
		} else {
			fieldVal = string(outValByte)
			if len(fieldVal) == 0 && httpFieldCfg.Mandatory {
				return nil, fmt.Errorf("MandatoryIeMissing: Empty result for http_post field: %s", httpFieldCfg.Tag)
			}
			if err := cdr.ParseFieldValue(httpFieldCfg.FieldId, fieldVal, xmlProc.timezone); err != nil {
				return nil, err
			}
		}
	}
	return cdr, nil
}
