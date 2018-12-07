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

package cdrc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/antchfx/xmlquery"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// getElementText will process the node to extract the elementName's text out of it (only first one found)
// returns utils.ErrNotFound if the element is not found in the node
func elementText(xmlElement *xmlquery.Node, elmntPath string) (string, error) {
	elmnt := xmlquery.FindOne(xmlElement, elmntPath)
	if elmnt == nil {
		return "", utils.ErrNotFound
	}
	return elmnt.InnerText(), nil

}

// handlerUsageDiff will calculate the usage as difference between timeEnd and timeStart
// Expects the 2 arguments in template separated by |
func handlerSubstractUsage(xmlElement *xmlquery.Node, argsTpl config.RSRParsers, cdrPath utils.HierarchyPath, timezone string) (time.Duration, error) {
	var argsStr string
	for _, rsrArg := range argsTpl {
		if rsrArg.Rules == utils.HandlerArgSep {
			argsStr += rsrArg.Rules
			continue
		}
		absolutePath := utils.ParseHierarchyPath(rsrArg.Rules, "")
		relPath := utils.HierarchyPath(absolutePath[len(cdrPath):]) // Need relative path to the xmlElmnt
		argStr, _ := elementText(xmlElement, relPath.AsString("/", false))
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
	if tEnd.IsZero() {
		return time.Duration(0), fmt.Errorf("EndTime is 0")
	}
	tStart, err := utils.ParseTimeDetectLayout(handlerArgs[1], timezone)
	if err != nil {
		return time.Duration(0), err
	}
	if tStart.IsZero() {
		return time.Duration(0), fmt.Errorf("StartTime is 0")
	}
	return tEnd.Sub(tStart), nil
}

func NewXMLRecordsProcessor(recordsReader io.Reader, cdrPath utils.HierarchyPath, timezone string,
	httpSkipTlsCheck bool, cdrcCfgs []*config.CdrcCfg, filterS *engine.FilterS) (*XMLRecordsProcessor, error) {
	//create doc
	doc, err := xmlquery.Parse(recordsReader)
	if err != nil {
		return nil, err
	}
	xmlProc := &XMLRecordsProcessor{cdrPath: cdrPath, timezone: timezone,
		httpSkipTlsCheck: httpSkipTlsCheck, cdrcCfgs: cdrcCfgs, filterS: filterS}

	xmlProc.cdrXmlElmts = xmlquery.Find(doc, cdrPath.AsString("/", true))
	return xmlProc, nil
}

type XMLRecordsProcessor struct {
	cdrXmlElmts      []*xmlquery.Node    // result of splitting the XML doc into CDR elements
	procItems        int                 // current number of processed records from file
	cdrPath          utils.HierarchyPath // path towards one CDR element
	timezone         string
	httpSkipTlsCheck bool
	cdrcCfgs         []*config.CdrcCfg // individual configs for the folder CDRC is monitoring
	filterS          *engine.FilterS
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
	xmlProvider := newXmlProvider(cdrXML, xmlProc.cdrPath)
	for _, cdrcCfg := range xmlProc.cdrcCfgs {
		tenant, err := cdrcCfg.Tenant.ParseDataProvider(xmlProvider, utils.NestingSep)
		if err != nil {
			return nil, err
		}
		if len(cdrcCfg.Filters) != 0 {
			if pass, err := xmlProc.filterS.Pass(tenant,
				cdrcCfg.Filters, xmlProvider); err != nil || !pass {
				continue // Not passes filters, ignore this CDR
			}
		}
		if cdr, err := xmlProc.recordToCDR(cdrXML, cdrcCfg, tenant); err != nil {
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

func (xmlProc *XMLRecordsProcessor) recordToCDR(xmlEntity *xmlquery.Node, cdrcCfg *config.CdrcCfg, tenant string) (*engine.CDR, error) {
	cdr := &engine.CDR{OriginHost: "0.0.0.0", Source: cdrcCfg.CdrSourceId, ExtraFields: make(map[string]string), Cost: -1}
	var lazyHttpFields []*config.FCTemplate
	var err error
	fldVals := make(map[string]string)
	xmlProvider := newXmlProvider(xmlEntity, xmlProc.cdrPath)
	for _, cdrFldCfg := range cdrcCfg.ContentFields {
		if len(cdrFldCfg.Filters) != 0 {
			if pass, err := xmlProc.filterS.Pass(tenant,
				cdrFldCfg.Filters, xmlProvider); err != nil || !pass {
				continue // Not passes filters, ignore this CDR
			}
		}
		if cdrFldCfg.Type == utils.META_COMPOSED {
			out, err := cdrFldCfg.Value.ParseDataProvider(xmlProvider, utils.NestingSep)
			if err != nil {
				return nil, err
			}
			fldVals[cdrFldCfg.FieldId] += out
		} else if cdrFldCfg.Type == utils.META_HTTP_POST {
			lazyHttpFields = append(lazyHttpFields, cdrFldCfg) // Will process later so we can send an estimation of cdr to http server
		} else if cdrFldCfg.Type == utils.META_HANDLER && cdrFldCfg.HandlerId == utils.HandlerSubstractUsage {
			usage, err := handlerSubstractUsage(xmlEntity, cdrFldCfg.Value, xmlProc.cdrPath, xmlProc.timezone)
			if err != nil {
				return nil, fmt.Errorf("Ignoring record: %v - cannot extract field %s, err: %s", xmlEntity, cdrFldCfg.Tag, err.Error())
			}
			fldVals[cdrFldCfg.FieldId] += usage.String()
		} else {
			return nil, fmt.Errorf("Unsupported field type: %s", cdrFldCfg.Type)
		}
		if err := cdr.ParseFieldValue(cdrFldCfg.FieldId, fldVals[cdrFldCfg.FieldId], xmlProc.timezone); err != nil {
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
			if parsed, err := rsrFld.ParseValue(utils.EmptyString); err != nil {
				return nil, fmt.Errorf("Ignoring record: %v - cannot extract http address, err: %s", xmlEntity, err.Error())
			} else {
				httpAddr += parsed
			}
		}
		var jsn []byte
		jsn, err = json.Marshal(cdr)
		if err != nil {
			return nil, err
		}
		if outValByte, err = engine.HttpJsonPost(httpAddr, xmlProc.httpSkipTlsCheck, jsn); err != nil && httpFieldCfg.Mandatory {
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

// newXmlProvider constructs a DataProvider
func newXmlProvider(req *xmlquery.Node, cdrPath utils.HierarchyPath) (dP config.DataProvider) {
	dP = &xmlProvider{req: req, cdrPath: cdrPath, cache: config.NewNavigableMap(nil)}
	return
}

// xmlProvider implements engine.DataProvider so we can pass it to filters
type xmlProvider struct {
	req     *xmlquery.Node
	cdrPath utils.HierarchyPath //used to compute relative path
	cache   *config.NavigableMap
}

// String is part of engine.DataProvider interface
// when called, it will display the already parsed values out of cache
func (xP *xmlProvider) String() string {
	return utils.ToJSON(xP)
}

// FieldAsInterface is part of engine.DataProvider interface
func (xP *xmlProvider) FieldAsInterface(fldPath []string) (data interface{}, err error) {
	if len(fldPath) == 0 {
		return nil, utils.ErrNotFound
	}
	if data, err = xP.cache.FieldAsInterface(fldPath); err == nil ||
		err != utils.ErrNotFound { // item found in cache
		return
	}
	err = nil                                                 // cancel previous err
	relPath := utils.HierarchyPath(fldPath[len(xP.cdrPath):]) // Need relative path to the xmlElmnt
	var slctrStr string
	for i := range relPath {
		if sIdx := strings.Index(relPath[i], "["); sIdx != -1 {
			slctrStr = relPath[i][sIdx:]
			if slctrStr[len(slctrStr)-1:] != "]" {
				return nil, fmt.Errorf("filter rule <%s> needs to end in ]", slctrStr)
			}
			relPath[i] = relPath[i][:sIdx]
			if slctrStr[1:2] != "@" {
				i, err := strconv.Atoi(slctrStr[1 : len(slctrStr)-1])
				if err != nil {
					return nil, err
				}
				slctrStr = "[" + strconv.Itoa(i+1) + "]"
			}
			relPath[i] = relPath[i] + slctrStr
		}
	}
	data, err = elementText(xP.req, relPath.AsString("/", false))
	xP.cache.Set(fldPath, data, false, false)
	return
}

// FieldAsString is part of engine.DataProvider interface
func (xP *xmlProvider) FieldAsString(fldPath []string) (data string, err error) {
	var valIface interface{}
	valIface, err = xP.FieldAsInterface(fldPath)
	if err != nil {
		return
	}
	data, err = utils.IfaceAsString(valIface)
	return
}

// AsNavigableMap is part of engine.DataProvider interface
func (xP *xmlProvider) AsNavigableMap([]*config.FCTemplate) (
	nm *config.NavigableMap, err error) {
	return nil, utils.ErrNotImplemented
}

// RemoteHost is part of engine.DataProvider interface
func (xP *xmlProvider) RemoteHost() net.Addr {
	return new(utils.LocalAddr)
}
