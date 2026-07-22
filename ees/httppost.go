// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ees

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewHTTPPostEE(cfg *config.EventExporterCfg, cgrCfg *config.CGRConfig, filterS *engine.FilterS,
	em *utils.ExporterMetrics) (httpPost *HTTPPostEE, err error) {
	httpPost = &HTTPPostEE{
		cfg:    cfg,
		em:     em,
		client: &http.Client{Transport: engine.GetHTTPPstrTransport(), Timeout: cgrCfg.GeneralCfg().ReplyTimeout},
		reqs:   newConcReq(cfg.ConcurrentRequests),
	}
	httpPost.hdr, err = httpPost.composeHeader(cgrCfg, filterS)
	return
}

// FileCSVee implements EventExporter interface for .csv files
type HTTPPostEE struct {
	cfg    *config.EventExporterCfg
	em     *utils.ExporterMetrics
	client *http.Client
	reqs   *concReq

	hdr http.Header
}

type HTTPPosterRequest struct {
	Header http.Header
	Body   any
}

// Compose and cache the header
func (httpPost *HTTPPostEE) composeHeader(cgrCfg *config.CGRConfig, filterS *engine.FilterS) (hdr http.Header, err error) {
	hdr = make(http.Header)
	if len(httpPost.Cfg().HeaderFields()) == 0 {
		return
	}
	var exp *utils.OrderedNavigableMap
	if exp, err = composeHeaderTrailer(utils.MetaHdr, httpPost.Cfg().HeaderFields(), httpPost.em, cgrCfg, filterS); err != nil {
		return
	}
	for el := exp.GetFirstElement(); el != nil; el = el.Next() {
		path := el.Value
		nmIt, _ := exp.Field(path) //Safe to ignore error, since the path always exists
		path = utils.StripTrailingIndex(path)
		hdr.Set(strings.Join(path, utils.NestingSep), nmIt.String())
	}
	return
}

func (httpPost *HTTPPostEE) Cfg() *config.EventExporterCfg { return httpPost.cfg }

func (httpPost *HTTPPostEE) Connect() (_ error) { return }

func (httpPost *HTTPPostEE) ExportEvent(content any, _ string) (err error) {
	httpPost.reqs.get()
	defer httpPost.reqs.done()
	pReq := content.(*HTTPPosterRequest)
	var req *http.Request
	if req, err = prepareRequest(httpPost.Cfg().ExportPath, utils.ContentForm, pReq.Body, pReq.Header); err != nil {
		return
	}
	_, err = sendHTTPReq(httpPost.client, req)
	return
}

func (httpPost *HTTPPostEE) Close() (_ error) { return }

func (httpPost *HTTPPostEE) GetMetrics() *utils.ExporterMetrics { return httpPost.em }

func (httpPost *HTTPPostEE) PrepareMap(mp *utils.CGREvent) (any, error) {
	urlVals := url.Values{}
	for k, v := range mp.Event {
		urlVals.Set(k, utils.IfaceAsString(v))
	}
	return &HTTPPosterRequest{
		Header: httpPost.hdr.Clone(),
		Body:   urlVals,
	}, nil
}

func (httpPost *HTTPPostEE) PrepareOrderMap(mp *utils.OrderedNavigableMap) (any, error) {
	urlVals := url.Values{}
	for el := mp.GetFirstElement(); el != nil; el = el.Next() {
		path := el.Value
		nmIt, _ := mp.Field(path)
		path = utils.StripTrailingIndex(path)
		urlVals.Set(strings.Join(path, utils.NestingSep), nmIt.String())
	}
	return &HTTPPosterRequest{
		Header: httpPost.hdr.Clone(),
		Body:   urlVals,
	}, nil
}
