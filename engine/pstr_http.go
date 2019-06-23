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
	"os"
	"time"

	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
)

// Post without automatic failover
func HttpJsonPost(url string, skipTlsVerify bool, content []byte) ([]byte, error) {
	tr := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: skipTlsVerify},
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(content))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode > 299 {
		return respBody, fmt.Errorf("Unexpected status code received: %d", resp.StatusCode)
	}
	return respBody, nil
}

func NewHTTPPoster(skipTLSVerify bool, replyTimeout time.Duration) *HTTPPoster {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipTLSVerify},
	}
	return &HTTPPoster{httpClient: &http.Client{Transport: tr, Timeout: replyTimeout}}
}

type HTTPPoster struct {
	httpClient *http.Client
}

// Post with built-in failover
// Returns also reference towards client so we can close it's connections when done
func (poster *HTTPPoster) Post(addr string, contentType string, content interface{}, attempts int, fallbackFilePath string) (respBody []byte, err error) {
	if !utils.IsSliceMember([]string{utils.CONTENT_JSON, utils.CONTENT_FORM, utils.CONTENT_TEXT}, contentType) {
		return nil, fmt.Errorf("unsupported ContentType: %s", contentType)
	}
	var body []byte        // Used to write in file and send over http
	var urlVals url.Values // Used when posting form
	if utils.IsSliceMember([]string{utils.CONTENT_JSON, utils.CONTENT_TEXT}, contentType) {
		body = content.([]byte)
	} else if contentType == utils.CONTENT_FORM {
		urlVals = content.(url.Values)
		body = []byte(urlVals.Encode())
	}
	fib := utils.Fib()
	bodyType := "application/x-www-form-urlencoded"
	if contentType == utils.CONTENT_JSON {
		bodyType = "application/json"
	}
	for i := 0; i < attempts; i++ {
		var resp *http.Response
		if utils.IsSliceMember([]string{utils.CONTENT_JSON, utils.CONTENT_TEXT}, contentType) {
			resp, err = poster.httpClient.Post(addr, bodyType, bytes.NewBuffer(body))
		} else if contentType == utils.CONTENT_FORM {
			resp, err = poster.httpClient.PostForm(addr, urlVals)
		}
		if err != nil {
			utils.Logger.Warning(fmt.Sprintf("<HTTPPoster> Posting to : <%s>, error: <%s>", addr, err.Error()))
			time.Sleep(time.Duration(fib()) * time.Second)
			continue
		}
		defer resp.Body.Close()
		respBody, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			utils.Logger.Warning(fmt.Sprintf("<HTTPPoster> Posting to : <%s>, error: <%s>", addr, err.Error()))
			time.Sleep(time.Duration(fib()) * time.Second)
			continue
		}
		if resp.StatusCode > 299 {
			utils.Logger.Warning(fmt.Sprintf("<HTTPPoster> Posting to : <%s>, unexpected status code received: <%d>", addr, resp.StatusCode))
			time.Sleep(time.Duration(fib()) * time.Second)
			continue
		}
		return respBody, nil
	}
	if fallbackFilePath != utils.META_NONE {
		// If we got that far, post was not possible, write it on disk
		_, err = guardian.Guardian.Guard(func() (interface{}, error) {
			fileOut, err := os.Create(fallbackFilePath)
			if err != nil {
				return nil, err
			}
			_, err = fileOut.Write(body)
			fileOut.Close()
			return nil, err
		}, time.Duration(2*time.Second), utils.FileLockPrefix+fallbackFilePath)
	}
	return
}
