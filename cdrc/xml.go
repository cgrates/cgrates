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
	"encoding/xml"
	"io"
	"path"

	"github.com/ChrisTrenkamp/goxpath"
	"github.com/ChrisTrenkamp/goxpath/tree"
	"github.com/ChrisTrenkamp/goxpath/tree/xmltree"
	"github.com/cgrates/cgrates/engine"
)

func NewXMLRecordsProcessor(recordsReader io.Reader) (*XMLRecordsProcessor, error) {
	xp, err := goxpath.Parse(path.Join("/broadWorksCDR/cdrData/"))
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
	xmlProc := new(XMLRecordsProcessor)
	xmlProc.cdrXmlElmts = goxpath.MustExec(xp, xmlNode, nil)
	return nil, nil
}

type XMLRecordsProcessor struct {
	cdrXmlElmts []tree.Res // result of splitting the XML doc into CDR elements
	procItems   int        // current number of processed records from file
}

func (xmlProc *XMLRecordsProcessor) ProcessedRecordsNr() int64 {
	return int64(xmlProc.procItems)
}

func (xmlProc *XMLRecordsProcessor) ProcessNextRecord() (cdrs []*engine.CDR, err error) {
	if len(xmlProc.cdrXmlElmts) <= xmlProc.procItems {
		return nil, io.EOF // have processed all items
	}
	cdrXml := xmlProc.cdrXmlElmts[xmlProc.procItems]
	xmlProc.procItems += 1
	cdrBuf := bytes.NewBufferString(xml.Header)
	if err := goxpath.Marshal(cdrXml.(tree.Node), cdrBuf); err != nil {
		return nil, err
	}
	return nil, nil
}
