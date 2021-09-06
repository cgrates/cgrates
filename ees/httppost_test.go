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
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestHttpPostGetMetrics(t *testing.T) {
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	if err != nil {
		t.Error(err)
	}
	httpPost := &HTTPPostEE{
		dc: dc,
	}

	if rcv := httpPost.GetMetrics(); !reflect.DeepEqual(rcv, httpPost.dc) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(rcv), utils.ToJSON(httpPost.dc))
	}
}

func TestHttpPostExportEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaHTTPPost
	cgrEv := new(utils.CGREvent)
	httpPost, err := NewHTTPPostEE(cgrCfg.EEsCfg().Exporters[0], cgrCfg, nil, nil)
	if err != nil {
		t.Error(err)
	}
	cgrEv.Event = map[string]interface{}{
		"Test1": 3,
	}
	errExpect := `Post "/var/spool/cgrates/ees": unsupported protocol scheme ""`
	if err := httpPost.ExportEvent(context.Background(), &HTTPPosterRequest{Body: url.Values{}, Header: make(http.Header)}, ""); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %q but received %q", errExpect, err)
	}
}

func TestHttpPostExportEvent2(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaHTTPPost
	bodyExpect := "2=%2Areq.field2"
	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}
		if strBody := string(body); strBody != bodyExpect {
			t.Errorf("Expected %q but received %q", bodyExpect, strBody)
		}
		rw.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	cgrCfg.EEsCfg().Exporters[0].ExportPath = srv.URL + "/"
	httpPost, err := NewHTTPPostEE(cgrCfg.EEsCfg().Exporters[0], cgrCfg, nil, nil)
	if err != nil {
		t.Error(err)
	}
	vals, err := httpPost.PrepareMap(map[string]interface{}{"2": "*req.field2"})
	if err != nil {
		t.Fatal(err)
	}
	if err := httpPost.ExportEvent(context.Background(), vals, ""); err != nil {
		t.Error(err)
	}
}

func TestHttpPostSync(t *testing.T) {
	//Create new exporter
	cgrCfg := config.NewDefaultCGRConfig()

	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaHTTPPost

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

	exp, err := NewHTTPPostEE(cgrCfg.EEsCfg().Exporters[0], cgrCfg, nil, nil)
	if err != nil {
		t.Error(err)
	}

	vals, err := exp.PrepareMap(map[string]interface{}{
		"Account":     "1001",
		"Destination": "1002",
	})
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 3; i++ {
		go exp.ExportEvent(context.Background(), vals, "")
	}

	select {
	case <-test:
		return
	case <-time.After(50 * time.Millisecond):
		t.Error("Can't asynchronously export events")
	}
}

func TestHttpPostSyncLimit(t *testing.T) {
	//Create new exporter
	cgrCfg := config.NewDefaultCGRConfig()

	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaHTTPPost

	// We set the limit of events to be exported lower than the amount of events we asynchronously want to export
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

	exp, err := NewHTTPPostEE(cgrCfg.EEsCfg().Exporters[0], cgrCfg, nil, nil)
	if err != nil {
		t.Error(err)
	}

	vals, err := exp.PrepareMap(map[string]interface{}{
		"Account":     "1001",
		"Destination": "1002",
	})
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 3; i++ {
		go exp.ExportEvent(context.Background(), vals, "")
	}
	select {
	case <-test:
		t.Error("Should not have been possible to asynchronously export events")
	case <-time.After(50 * time.Millisecond):
		return
	}
}
