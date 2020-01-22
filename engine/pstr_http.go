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

package engine

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/cgrates/cgrates/utils"
)

// keep it global in order to reuse it
var httpPosterTransport *http.Transport

// HttpJsonPost posts without automatic failover
func HttpJsonPost(url string, skipTLSVerify bool, content []byte) (respBody []byte, err error) {
	if httpPosterTransport == nil {
		httpPosterTransport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: skipTLSVerify},
		}
	}
	client := &http.Client{Transport: httpPosterTransport}
	var resp *http.Response
	if resp, err = client.Post(url, "application/json", bytes.NewBuffer(content)); err != nil {
		return
	}
	respBody, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return
	}
	if resp.StatusCode > 299 {
		err = fmt.Errorf("Unexpected status code received: %d", resp.StatusCode)
	}
	return
}

// NewHTTPPoster return a new HTTP poster
func NewHTTPPoster(skipTLSVerify bool, replyTimeout time.Duration,
	addr, contentType string, attempts int) (httposter *HTTPPoster, err error) {
	if !utils.SliceHasMember([]string{utils.CONTENT_FORM, utils.CONTENT_JSON, utils.CONTENT_TEXT}, contentType) {
		return nil, fmt.Errorf("unsupported ContentType: %s", contentType)
	}
	if httpPosterTransport == nil {
		httpPosterTransport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: skipTLSVerify},
		}
	}
	return &HTTPPoster{
		httpClient:  &http.Client{Transport: httpPosterTransport, Timeout: replyTimeout},
		addr:        addr,
		contentType: contentType,
		attempts:    attempts,
	}, nil
}

// HTTPPoster used to post cdrs
type HTTPPoster struct {
	httpClient  *http.Client
	addr        string
	contentType string
	attempts    int
}

// Post will post the event
func (pstr *HTTPPoster) Post(content interface{}, key string) (err error) {
	_, err = pstr.GetResponse(content)
	return
}

// GetResponse will post the event and return the response
func (pstr *HTTPPoster) GetResponse(content interface{}) (respBody []byte, err error) {
	var body []byte        // Used to write in file and send over http
	var urlVals url.Values // Used when posting form
	if pstr.contentType == utils.CONTENT_FORM {
		urlVals = content.(url.Values)
		body = []byte(urlVals.Encode())
	} else {
		body = content.([]byte)
	}
	fib := utils.Fib()
	bodyType := "application/x-www-form-urlencoded"
	if pstr.contentType == utils.CONTENT_JSON {
		bodyType = "application/json"
	}
	for i := 0; i < pstr.attempts; i++ {
		var resp *http.Response
		if pstr.contentType == utils.CONTENT_FORM {
			resp, err = pstr.httpClient.PostForm(pstr.addr, urlVals)
		} else {
			resp, err = pstr.httpClient.Post(pstr.addr, bodyType, bytes.NewBuffer(body))
		}
		if err != nil {
			utils.Logger.Warning(fmt.Sprintf("<HTTPPoster> Posting to : <%s>, error: <%s>", pstr.addr, err.Error()))
			if i+1 < pstr.attempts {
				time.Sleep(time.Duration(fib()) * time.Second)
			}
			continue
		}
		respBody, err = ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			utils.Logger.Warning(fmt.Sprintf("<HTTPPoster> Posting to : <%s>, error: <%s>", pstr.addr, err.Error()))
			if i+1 < pstr.attempts {
				time.Sleep(time.Duration(fib()) * time.Second)
			}
			continue
		}
		if resp.StatusCode > 299 {
			utils.Logger.Warning(fmt.Sprintf("<HTTPPoster> Posting to : <%s>, unexpected status code received: <%d>", pstr.addr, resp.StatusCode))
			if i+1 < pstr.attempts {
				time.Sleep(time.Duration(fib()) * time.Second)
			}
			continue
		}
		return respBody, nil
	}
	return
}
