/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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

package ees

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestHttpJsonMapGetMetrics(t *testing.T) {
	dc, err := newEEMetrics("Local")
	if err != nil {
		t.Error(err)
	}
	httpEE := &HTTPjsonMapEE{
		dc: dc,
	}

	if rcv := httpEE.GetMetrics(); !reflect.DeepEqual(rcv, httpEE.dc) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(rcv), utils.ToJSON(httpEE.dc))
	}
}

func TestHttpJsonMapExportEvent1(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaSQSjsonMap

	httpEE, err := NewHTTPjsonMapEE(cgrCfg.EEsCfg().Exporters[0], cgrCfg, nil, nil)
	if err != nil {
		t.Error(err)
	}
	errExpect := `Post "/var/spool/cgrates/ees": unsupported protocol scheme ""`
	if err := httpEE.ExportEvent(context.Background(), &HTTPPosterRequest{Body: []byte{}, Header: make(http.Header)}, ""); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %q but received %q", errExpect, err)
	}
}

func TestHttpJsonMapExportEvent2(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaSQSjsonMap

	bodyExpect := map[string]interface{}{
		"2": "*req.field2",
	}
	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(body, bodyExpect) {
			t.Errorf("Expected %q but received %q", bodyExpect, body)
		}
		rw.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	cgrCfg.EEsCfg().Exporters[0].ExportPath = srv.URL + "/"
	httpEE, err := NewHTTPjsonMapEE(cgrCfg.EEsCfg().Exporters[0], cgrCfg, nil, nil)
	if err != nil {
		t.Error(err)
	}

	if err := httpEE.ExportEvent(context.Background(), &HTTPPosterRequest{Body: []byte(`{"2": "*req.field2"}`), Header: make(http.Header)}, ""); err != nil {
		t.Error(err)
	}
}

func TestHttpJsonMapSync(t *testing.T) {
	//Create new exporter
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaHTTPjsonMap

	var wg1 sync.WaitGroup

	wg1.Add(3)

	test := make(chan struct{})
	go func() {
		wg1.Wait()
		close(test)
	}()

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		time.Sleep(25 * time.Millisecond)
		wg1.Done()
	}))

	defer ts.Close()

	cgrCfg.EEsCfg().Exporters[0].ExportPath = ts.URL

	exp, err := NewHTTPjsonMapEE(cgrCfg.EEsCfg().Exporters[0], cgrCfg, nil, nil)
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < 3; i++ {
		go exp.ExportEvent(context.Background(), &HTTPPosterRequest{Body: []byte(`{"2": "*req.field2"}`), Header: make(http.Header)}, "")
	}

	select {
	case <-test:
		return
	case <-time.After(50 * time.Millisecond):
		t.Error("Can't asynchronously export events")
	}
}

func TestHttpJsonMapSyncLimit(t *testing.T) {
	//Create new exporter
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaHTTPjsonMap
	cgrCfg.EEsCfg().Exporters[0].ConcurrentRequests = 1

	var wg1 sync.WaitGroup
	wg1.Add(3)

	test := make(chan struct{})
	go func() {
		wg1.Wait()
		close(test)
	}()

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		time.Sleep(25 * time.Millisecond)
		wg1.Done()
	}))

	defer ts.Close()

	cgrCfg.EEsCfg().Exporters[0].ExportPath = ts.URL

	exp, err := NewHTTPjsonMapEE(cgrCfg.EEsCfg().Exporters[0], cgrCfg, nil, nil)
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < 3; i++ {
		go exp.ExportEvent(context.Background(), &HTTPPosterRequest{Body: []byte(`{"2": "*req.field2"}`), Header: make(http.Header)}, "")
	}

	select {
	case <-test:
		t.Error("Should not have been possible to asynchronously export events")
	case <-time.After(50 * time.Millisecond):
		return
	}
}
