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
	"archive/zip"
	"encoding/csv"
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

func NewCSVReader(prv CSVProvider, dPath, fn string, sep rune, nrFlds int) (_ *CSVFile, err error) {
	var file io.ReadCloser
	if file, err = prv.Open(dPath, fn); err != nil {
		return
	}
	return NewCSVFile(file, path.Join(dPath, fn), sep, nrFlds), nil
}

func NewCSVFile(rdr io.ReadCloser, path string, sep rune, nrFlds int) *CSVFile {
	csvRrdr := csv.NewReader(rdr)
	csvRrdr.Comma = sep
	csvRrdr.Comment = utils.CommentChar
	csvRrdr.FieldsPerRecord = nrFlds
	return &CSVFile{
		path:   path,
		cls:    rdr,
		csvRdr: csvRrdr,
	}
}

func NewStringCSV(data string, sep rune, nrFlds int) *CSVFile {
	return NewCSVFile(io.NopCloser(strings.NewReader(data)), data, sep, nrFlds)
}

type CSVFile struct {
	path   string    // only for logging purposes
	cls    io.Closer // keep reference so we can close it when done
	csvRdr *csv.Reader
}

func (c *CSVFile) Path() string            { return c.path }
func (c *CSVFile) Read() ([]string, error) { return c.csvRdr.Read() }
func (c *CSVFile) Close() error            { return c.cls.Close() }

type CSVProvider interface {
	Open(dPath, fn string) (io.ReadCloser, error)
	Type() string
}

type fileProvider struct{}

func (fileProvider) Open(dPath, fn string) (io.ReadCloser, error) {
	return os.Open(path.Join(dPath, fn))
}

func (fileProvider) Type() string { return utils.MetaFileCSV }

type urlProvider struct{}

func (urlProvider) Open(dPath, fn string) (_ io.ReadCloser, err error) {
	path := strings.TrimSuffix(dPath, utils.Slash) + utils.Slash + fn
	if _, err = url.ParseRequestURI(path); err != nil {
		return
	}
	var req *http.Response
	if req, err = (&http.Client{
		Transport: engine.HTTPPstrTransport(),
		Timeout:   config.CgrConfig().GeneralCfg().ReplyTimeout,
	}).Get(path); err != nil {
		err = utils.ErrPathNotReachable(path)
		return
	}
	if req.StatusCode != http.StatusOK {
		err = utils.ErrNotFound
		return
	}
	return req.Body, nil
}
func (urlProvider) Type() string { return utils.MetaUrl }

type zipProvider struct{ *zip.Reader }

func (z zipProvider) Open(_, fn string) (io.ReadCloser, error) { return z.Reader.Open(fn) }
func (zipProvider) Type() string                               { return utils.MetaZip }

type stringProvider struct{}

func (stringProvider) Open(_, fn string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader(fn)), nil
}
func (stringProvider) Type() string { return utils.MetaString }
