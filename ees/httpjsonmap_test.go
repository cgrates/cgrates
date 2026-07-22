// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestHttpJsonMapGetMetrics(t *testing.T) {
	em, err := utils.NewExporterMetrics("", "Local")
	if err != nil {
		t.Fatal(err)
	}
	httpEE := &HTTPjsonMapEE{
		em: em,
	}

	if rcv := httpEE.GetMetrics(); !reflect.DeepEqual(rcv, httpEE.em) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(rcv), utils.ToJSON(httpEE.em))
	}
}

func TestHttpJsonMapExportEvent1(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cgrCfg)
	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaSQSjsonMap

	httpEE, err := NewHTTPjsonMapEE(cgrCfg.EEsCfg().Exporters[0], cgrCfg, engine.NewCacheS(cgrCfg, nil, nil, nil, locker), nil, nil)
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
	locker := engine.NewLocker(cgrCfg)
	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaSQSjsonMap

	bodyExpect := map[string]any{
		"2": "*req.field2",
	}
	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var body map[string]any
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
	httpEE, err := NewHTTPjsonMapEE(cgrCfg.EEsCfg().Exporters[0], cgrCfg, engine.NewCacheS(cgrCfg, nil, nil, nil, locker), nil, nil)
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
	locker := engine.NewLocker(cgrCfg)
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

	exp, err := NewHTTPjsonMapEE(cgrCfg.EEsCfg().Exporters[0], cgrCfg, engine.NewCacheS(cgrCfg, nil, nil, nil, locker), nil, nil)
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
	locker := engine.NewLocker(cgrCfg)
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

	exp, err := NewHTTPjsonMapEE(cgrCfg.EEsCfg().Exporters[0], cgrCfg, engine.NewCacheS(cgrCfg, nil, nil, nil, locker), nil, nil)
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

func TestHTTPJsonMapPrepareOrderMap(t *testing.T) {
	httpEE := new(HTTPjsonMapEE)
	onm := utils.NewOrderedNavigableMap()
	fullPath := &utils.FullPath{
		PathSlice: []string{utils.MetaReq, utils.MetaTenant},
		Path:      utils.MetaTenant,
	}
	onm.SetAsSlice(fullPath, []*utils.DataNode{utils.NewLeafNode("value1")})
	rcv, err := httpEE.PrepareOrderMap(onm)
	if err != nil {
		t.Error(err)
	}
	valMp := map[string]any{
		"*req": map[string]any{
			"*tenant": "value1",
		},
	}
	body, err := json.Marshal(valMp)
	exp := &HTTPPosterRequest{
		Header: httpEE.hdr,
		Body:   body,
	}
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", utils.IfaceAsString(exp), utils.IfaceAsString(rcv))
	}
}

func TestHTTPJsonMapPrepareMap(t *testing.T) {
	httpEE := new(HTTPjsonMapEE)
	valMp := map[string]any{
		"*req.*tenant": "value1",
	}
	rcv, err := httpEE.PrepareMap(&utils.CGREvent{
		Event: valMp,
	})
	if err != nil {
		t.Error(err)
	}
	body, err := json.Marshal(valMp)
	if err != nil {
		t.Error(err)
	}
	exp := &HTTPPosterRequest{
		Header: httpEE.hdr,
		Body:   body,
	}
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", utils.IfaceAsString(exp), utils.IfaceAsString(rcv))
	}
}
