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

package loaders

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type CSVReader interface {
	Path() string
	Read() ([]string, error)
	Close() error
}

func NewCSVReader(csvType, dPath, fn string, sep rune, nrFlds int) (CSVReader, error) {
	switch csvType {
	case utils.MetaFileCSV:
		return NewFileCSV(path.Join(dPath, fn), sep, nrFlds)
	case utils.MetaUrl:
		return NewURLCSV(strings.TrimSuffix(dPath, utils.Slash)+utils.Slash+fn, sep, nrFlds)
	case utils.MetaGoogleAPI: // TODO: Implement *gapi
		return nil, nil
	case utils.MetaString:
		return NewStringCSV(fn, sep, nrFlds)
	default:
		return nil, fmt.Errorf("unsupported CSVReader type: <%q>", csvType)
	}
}

func NewCSVFile(rdr io.ReadCloser, path string, sep rune, nrFlds int) CSVReader {
	csvRrdr := csv.NewReader(rdr)
	csvRrdr.Comma = sep
	csvRrdr.Comment = utils.CommentChar
	csvRrdr.FieldsPerRecord = nrFlds
	csvRrdr.TrailingComma = true
	return &CSVFile{
		path:   path,
		cls:    rdr,
		csvRdr: csvRrdr,
	}
}

func NewFileCSV(path string, sep rune, nrFlds int) (_ CSVReader, err error) {
	var file io.ReadCloser
	if file, err = os.Open(path); err != nil {
		return
	}
	return NewCSVFile(file, path, sep, nrFlds), nil
}

func NewStringCSV(data string, sep rune, nrFlds int) (_ CSVReader, err error) {
	return NewCSVFile(io.NopCloser(strings.NewReader(data)), data, sep, nrFlds), nil
}

func NewURLCSV(path string, sep rune, nrFlds int) (_ CSVReader, err error) {
	if _, err = url.ParseRequestURI(path); err != nil {
		return
	}
	var req *http.Response
	if req, err = (&http.Client{
		Transport: engine.GetHTTPPstrTransport(),
		Timeout:   config.CgrConfig().GeneralCfg().ReplyTimeout,
	}).Get(path); err != nil {
		err = utils.ErrPathNotReachable(path)
		return
	}
	if req.StatusCode != http.StatusOK {
		err = utils.ErrNotFound
		return
	}
	return NewCSVFile(req.Body, path, sep, nrFlds), nil
}

type CSVFile struct {
	path   string    // only for logging purposes
	cls    io.Closer // keep reference so we can close it when done
	csvRdr *csv.Reader
}

func (c *CSVFile) Path() string            { return c.path }
func (c *CSVFile) Read() ([]string, error) { return c.csvRdr.Read() }
func (c *CSVFile) Close() error            { return c.cls.Close() }
