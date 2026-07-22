// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/antchfx/xmlquery"
	"github.com/cgrates/cgrates/utils"
)

// NewXMLProvider constructs a utils.DataProvider
func NewXMLProvider(req *xmlquery.Node, cdrPath utils.HierarchyPath) (dP utils.DataProvider) {
	dP = &XMLProvider{
		req:     req,
		cdrPath: cdrPath,
		cache:   utils.MapStorage{},
	}
	return
}

// XmlProvider implements engine.utils.DataProvider, allowing it to be passed to filters.
// An instance of XmlProvider is created for each element retrieved from xmlquery.QueryAll.
type XMLProvider struct {

	// req is the root node of each individual XML element retrieved.
	// It acts as the starting point for extracting data from the XML element.
	req *xmlquery.Node

	// cdrPath is the root path used to calculate the relative path.
	// The relative path is computed by trimming cdrPath from the full path of the XML element.
	cdrPath utils.HierarchyPath

	// cache stores the inner text of XML elements.
	// The key is the relative path of the XML element.
	cache utils.MapStorage
}

// String is part of engine.utils.DataProvider interface
// when called, it will display the already parsed values out of cache
func (xP *XMLProvider) String() string {
	return xP.cache.String()
}

// FieldAsInterface is part of engine.utils.DataProvider interface
func (xP *XMLProvider) FieldAsInterface(fldPath []string) (data any, err error) {
	if len(fldPath) == 0 {
		return nil, utils.ErrNotFound
	}
	if data, err = xP.cache.FieldAsInterface(fldPath); err == nil ||
		err != utils.ErrNotFound { // item found in cache
		return
	}
	err = nil // cancel previous err
	relPath := utils.HierarchyPath(fldPath[len(xP.cdrPath):])
	var slctrStr string
	for i := range relPath {
		if sIdx := strings.Index(relPath[i], "["); sIdx != -1 {
			slctrStr = relPath[i][sIdx:]
			if slctrStr[len(slctrStr)-1] != ']' {
				return nil, fmt.Errorf("filter rule <%s> needs to end in ]", slctrStr)
			}
			relPath[i] = relPath[i][:sIdx]
			if slctrStr[1] != '@' {
				i, err := strconv.Atoi(slctrStr[1 : len(slctrStr)-1])
				if err != nil {
					return nil, err
				}
				slctrStr = "[" + strconv.Itoa(i+1) + "]"
			}
			relPath[i] = relPath[i] + slctrStr
		}
	}
	data, err = ElementText(xP.req, relPath.AsString("/", false))
	xP.cache.Set(fldPath, data)
	return
}

// FieldAsString is part of engine.utils.DataProvider interface
func (xP *XMLProvider) FieldAsString(fldPath []string) (data string, err error) {
	var valIface any
	valIface, err = xP.FieldAsInterface(fldPath)
	if err != nil {
		return
	}
	return utils.IfaceAsString(valIface), nil
}

// ElementText will process the node to extract the elementName's text out of it (only first one found)
// returns utils.ErrNotFound if the element is not found in the node
// Make the method exportable until we remove the ers
func ElementText(xmlElement *xmlquery.Node, elmntPath string) (string, error) {
	elmnt, err := xmlquery.Query(xmlElement, elmntPath)
	if err != nil {
		return "", err
	}
	if elmnt == nil {
		return "", utils.ErrNotFound
	}
	return elmnt.InnerText(), nil
}
