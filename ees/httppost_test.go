// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestHttpPostGetMetrics(t *testing.T) {
	em, err := utils.NewExporterMetrics("", "Local")
	if err != nil {
		t.Fatal(err)
	}
	httpPost := &HTTPPostEE{
		em: em,
	}

	if rcv := httpPost.GetMetrics(); !reflect.DeepEqual(rcv, httpPost.em) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(rcv), utils.ToJSON(httpPost.em))
	}
}

func TestHttpPostExportEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cgrCfg)
	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaHTTPPost
	cgrEv := new(utils.CGREvent)
	httpPost, err := NewHTTPPostEE(cgrCfg.EEsCfg().Exporters[0], cgrCfg, engine.NewCacheS(cgrCfg, nil, nil, nil, locker), nil, nil)
	if err != nil {
		t.Error(err)
	}
	cgrEv.Event = map[string]any{
		"Test1": 3,
	}
	errExpect := `Post "/var/spool/cgrates/ees": unsupported protocol scheme ""`
	if err := httpPost.ExportEvent(context.Background(), &HTTPPosterRequest{Body: url.Values{}, Header: make(http.Header)}, ""); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %q but received %q", errExpect, err)
	}
}

func TestHttpPostExportEvent2(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cgrCfg)
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
	httpPost, err := NewHTTPPostEE(cgrCfg.EEsCfg().Exporters[0], cgrCfg, engine.NewCacheS(cgrCfg, nil, nil, nil, locker), nil, nil)
	if err != nil {
		t.Error(err)
	}
	vals, err := httpPost.PrepareMap(&utils.CGREvent{Event: map[string]any{"2": "*req.field2"}})
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
	locker := engine.NewLocker(cgrCfg)

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

	exp, err := NewHTTPPostEE(cgrCfg.EEsCfg().Exporters[0], cgrCfg, engine.NewCacheS(cgrCfg, nil, nil, nil, locker), nil, nil)
	if err != nil {
		t.Error(err)
	}

	req1, err := exp.PrepareMap(&utils.CGREvent{
		Event: map[string]any{
			"Account":     "1001",
			"Destination": "1002",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	req2, err := exp.PrepareMap(&utils.CGREvent{
		Event: map[string]any{
			"Account":     "1001",
			"Destination": "1003",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	req3, err := exp.PrepareMap(&utils.CGREvent{
		Event: map[string]any{
			"Account":     "1003",
			"Destination": "1001",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	requests := []any{req1, req2, req3}

	for i := range 3 {
		go exp.ExportEvent(context.Background(), requests[i], "")
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
	locker := engine.NewLocker(cgrCfg)

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

	exp, err := NewHTTPPostEE(cgrCfg.EEsCfg().Exporters[0], cgrCfg, engine.NewCacheS(cgrCfg, nil, nil, nil, locker), nil, nil)
	if err != nil {
		t.Error(err)
	}
	mp := &utils.CGREvent{
		Event: map[string]any{
			"Account":     "1001",
			"Destination": "1002",
		},
	}
	vals, err := exp.PrepareMap(mp)
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
