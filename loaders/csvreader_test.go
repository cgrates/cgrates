/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package loaders

import (
	"encoding/csv"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestNewCSVStringReader(t *testing.T) {
	data := `cgrates.org,ATTR_VARIABLE,,20,,*req.Category,*variable,~*req.ToR,`

	csvR, err := NewCSVReader(stringProvider{}, utils.EmptyString, data, utils.CSVSep, -1)
	if err != nil {
		t.Fatal(err)
	}
	cls := io.NopCloser(strings.NewReader(data))
	exp := &CSVFile{
		path:   data,
		cls:    cls,
		csvRdr: csv.NewReader(cls),
	}
	exp.csvRdr.Comma = utils.CSVSep
	exp.csvRdr.Comment = utils.CommentChar
	exp.csvRdr.FieldsPerRecord = -1

	if !reflect.DeepEqual(csvR, exp) {
		t.Errorf("Expeceted: %+v, received: %+v", exp, csvR)
	}
	if p := csvR.Path(); p != data {
		t.Errorf("Expeceted: %+v, received: %+v", data, p)
	}

	if p, err := csvR.Read(); err != nil {
		t.Fatal(err)
	} else if exp := strings.Split(data, ","); !reflect.DeepEqual(p, exp) {
		t.Errorf("Expeceted: %+v, received: %+v", exp, p)
	}

	if err := csvR.Close(); err != nil {
		t.Error(err)
	}
	tp := stringProvider{}.Type()
	if tp != utils.MetaString {
		t.Errorf("Expeceted: %q, received: %q", utils.MetaString, tp)
	}
	tp = zipProvider{}.Type()
	if tp != utils.MetaZip {
		t.Errorf("Expeceted: %q, received: %q", utils.MetaZip, tp)
	}
}

func TestNewCSVReaderErrors(t *testing.T) {
	path := "TestNewCSVReaderErrors" + strconv.Itoa(rand.Int()) + utils.CSVSuffix
	expErrMsg := fmt.Sprintf("open %s: no such file or directory", path)
	if _, err := NewCSVReader(fileProvider{}, ".", path, utils.CSVSep, -1); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}

	expErrMsg = fmt.Sprintf("parse %q: invalid URI for request", "./"+path)
	if _, err := NewCSVReader(urlProvider{}, ".", path, utils.CSVSep, -1); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}

	prePath := "http:/localhost:" + strconv.Itoa(rand.Int())
	expErrMsg = fmt.Sprintf(`path:"%s/%s" is not reachable`, prePath, path)
	if _, err := NewCSVReader(urlProvider{}, prePath, path, utils.CSVSep, -1); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %q, received: %q", expErrMsg, err.Error())
	}
}

func TestNewCSVURLReader(t *testing.T) {
	data := `cgrates.org,ATTR_VARIABLE,,20,,*req.Category,*variable,~*req.ToR,`
	mux := http.NewServeMux()
	mux.Handle("/ok/", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) { rw.Write([]byte(data)) }))
	s := httptest.NewServer(mux)
	defer s.Close()
	runtime.Gosched()

	if _, err := NewCSVReader(urlProvider{}, s.URL+"/notFound", utils.AttributesCsv, utils.CSVSep, -1); err != utils.ErrNotFound {
		t.Errorf("Expeceted: %v, received: %v", utils.ErrNotFound, err)
	}

	csvR, err := NewCSVReader(urlProvider{}, s.URL+"/ok", utils.AttributesCsv, utils.CSVSep, -1)
	if err != nil {
		t.Fatal(err)
	}
	expPath := path.Join(s.URL + "/ok/" + utils.AttributesCsv)
	if p := csvR.Path(); p != expPath {
		t.Errorf("Expeceted: %+v, received: %+v", p, expPath)
	}

	if p, err := csvR.Read(); err != nil {
		t.Fatal(err)
	} else if exp := strings.Split(data, ","); !reflect.DeepEqual(p, exp) {
		t.Errorf("Expeceted: %+v, received: %+v", exp, p)
	}

	if err := csvR.Close(); err != nil {
		t.Error(err)
	}
	tp := urlProvider{}.Type()
	if tp != utils.MetaUrl {
		t.Errorf("Expeceted: %q, received: %q", utils.MetaUrl, tp)
	}
}

func TestNewCSVFileReader(t *testing.T) {
	data := `cgrates.org,ATTR_VARIABLE,,20,,*req.Category,*variable,~*req.ToR,`
	dir, err := os.MkdirTemp(utils.EmptyString, "TestNewCSVFileReader")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	fp := path.Join(dir, utils.AttributesCsv)
	f, err := os.Create(fp)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.WriteString(data); err != nil {
		t.Fatal(err)
	}
	if err = f.Sync(); err != nil {
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}

	csvR, err := NewCSVReader(fileProvider{}, dir, utils.AttributesCsv, utils.CSVSep, -1)
	if err != nil {
		t.Fatal(err)
	}
	if p := csvR.Path(); p != fp {
		t.Errorf("Expeceted: %+v, received: %+v", data, fp)
	}

	if p, err := csvR.Read(); err != nil {
		t.Fatal(err)
	} else if exp := strings.Split(data, ","); !reflect.DeepEqual(p, exp) {
		t.Errorf("Expeceted: %+v, received: %+v", exp, p)
	}

	if err := csvR.Close(); err != nil {
		t.Error(err)
	}

}
