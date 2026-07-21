// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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

type urlProvider struct{ cfg *config.CGRConfig }

func (u urlProvider) Open(dPath, fn string) (_ io.ReadCloser, err error) {
	path := strings.TrimSuffix(dPath, utils.Slash) + utils.Slash + fn
	if _, err = url.ParseRequestURI(path); err != nil {
		return
	}
	var req *http.Response
	if req, err = (&http.Client{
		Transport: u.cfg.HTTPCfg().ClientOpts,
		Timeout:   u.cfg.GeneralCfg().ReplyTimeout,
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
