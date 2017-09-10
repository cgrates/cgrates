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

package utils

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/cgrates/cgrates/guardian"
	"github.com/streadway/amqp"
)

func init() {
	AMQPPostersCache = &AMQPCachedPosters{cache: make(map[string]*AMQPPoster)} // Initialize the cache for amqpPosters
}

var AMQPPostersCache *AMQPCachedPosters

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

// NewFallbackFileNameFronString will revert the meta information in the fallback file name into original data
func NewFallbackFileNameFronString(fileName string) (ffn *FallbackFileName, err error) {
	ffn = new(FallbackFileName)
	moduleIdx := strings.Index(fileName, HandlerArgSep)
	ffn.Module = fileName[:moduleIdx]
	var supportedModule bool
	for _, prfx := range []string{ActionsPoster, CDRPoster} {
		if strings.HasPrefix(ffn.Module, prfx) {
			supportedModule = true
			break
		}
	}
	if !supportedModule {
		return nil, fmt.Errorf("unsupported module: %s", ffn.Module)
	}
	fileNameWithoutModule := fileName[moduleIdx+1:]
	for _, trspt := range []string{MetaHTTPjsonCDR, MetaHTTPjsonMap, MetaHTTPjson, META_HTTP_POST, MetaAMQPjsonCDR, MetaAMQPjsonMap} {
		if strings.HasPrefix(fileNameWithoutModule, trspt) {
			ffn.Transport = trspt
			break
		}
	}
	if ffn.Transport == "" {
		return nil, fmt.Errorf("unsupported transport in fallback file path: %s", fileName)
	}
	fileNameWithoutTransport := fileNameWithoutModule[len(ffn.Transport)+1:]
	reqIDidx := strings.LastIndex(fileNameWithoutTransport, HandlerArgSep)
	if reqIDidx == -1 {
		return nil, fmt.Errorf("cannot find request ID in fallback file path: %s", fileName)
	}
	if ffn.Address, err = url.QueryUnescape(fileNameWithoutTransport[:reqIDidx]); err != nil {
		return nil, err
	}
	fileNameWithoutAddress := fileNameWithoutTransport[reqIDidx+1:]
	for _, suffix := range []string{TxtSuffix, JSNSuffix, FormSuffix} {
		if strings.HasSuffix(fileNameWithoutAddress, suffix) {
			ffn.FileSuffix = suffix
			break
		}
	}
	if ffn.FileSuffix == "" {
		return nil, fmt.Errorf("unsupported suffix in fallback file path: %s", fileName)
	}
	ffn.RequestID = fileNameWithoutAddress[:len(fileNameWithoutAddress)-len(ffn.FileSuffix)]
	return
}

// FallbackFileName is the struct defining the name of a file where CGRateS will dump data which fails to be sent remotely
type FallbackFileName struct {
	Module     string // name of the module writing the file
	Transport  string // transport used to send data remotely
	Address    string // remote address where data should have been sent
	RequestID  string // unique identifier of the request which should make files unique
	FileSuffix string // informative file termination suffix
}

func (ffn *FallbackFileName) AsString() string {
	if ffn.FileSuffix == "" { // Autopopulate FileSuffix based on the transport used
		ffn.FileSuffix = CDREFileSuffixes[ffn.Transport]
	}
	return fmt.Sprintf("%s%s%s%s%s%s%s%s", ffn.Module, HandlerArgSep, ffn.Transport, HandlerArgSep, url.QueryEscape(ffn.Address), HandlerArgSep, ffn.RequestID, ffn.FileSuffix)
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
	if !IsSliceMember([]string{CONTENT_JSON, CONTENT_FORM, CONTENT_TEXT}, contentType) {
		return nil, fmt.Errorf("unsupported ContentType: %s", contentType)
	}
	var body []byte        // Used to write in file and send over http
	var urlVals url.Values // Used when posting form
	if IsSliceMember([]string{CONTENT_JSON, CONTENT_TEXT}, contentType) {
		body = content.([]byte)
	} else if contentType == CONTENT_FORM {
		urlVals = content.(url.Values)
		body = []byte(urlVals.Encode())
	}
	fib := Fib()
	bodyType := "application/x-www-form-urlencoded"
	if contentType == CONTENT_JSON {
		bodyType = "application/json"
	}
	for i := 0; i < attempts; i++ {
		var resp *http.Response
		if IsSliceMember([]string{CONTENT_JSON, CONTENT_TEXT}, contentType) {
			resp, err = poster.httpClient.Post(addr, bodyType, bytes.NewBuffer(body))
		} else if contentType == CONTENT_FORM {
			resp, err = poster.httpClient.PostForm(addr, urlVals)
		}
		if err != nil {
			Logger.Warning(fmt.Sprintf("<HTTPPoster> Posting to : <%s>, error: <%s>", addr, err.Error()))
			time.Sleep(time.Duration(fib()) * time.Second)
			continue
		}
		defer resp.Body.Close()
		respBody, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			Logger.Warning(fmt.Sprintf("<HTTPPoster> Posting to : <%s>, error: <%s>", addr, err.Error()))
			time.Sleep(time.Duration(fib()) * time.Second)
			continue
		}
		if resp.StatusCode > 299 {
			Logger.Warning(fmt.Sprintf("<HTTPPoster> Posting to : <%s>, unexpected status code received: <%d>", addr, resp.StatusCode))
			time.Sleep(time.Duration(fib()) * time.Second)
			continue
		}
		return respBody, nil
	}
	if fallbackFilePath != META_NONE {
		// If we got that far, post was not possible, write it on disk
		_, err = guardian.Guardian.Guard(func() (interface{}, error) {
			fileOut, err := os.Create(fallbackFilePath)
			if err != nil {
				return nil, err
			}
			defer fileOut.Close()
			if _, err := fileOut.Write(body); err != nil {
				return nil, err
			}
			return nil, nil
		}, time.Duration(2*time.Second), FileLockPrefix+fallbackFilePath)
	}
	return
}

// AMQPPosterCache is used to cache mutliple AMQPPoster connections based on the address
type AMQPCachedPosters struct {
	sync.Mutex
	cache map[string]*AMQPPoster
}

// GetAMQPPoster creates a new poster only if not already cached
// uses dialURL as cache key
func (pc *AMQPCachedPosters) GetAMQPPoster(dialURL string, attempts int, fallbackFileDir string) (amqpPoster *AMQPPoster, err error) {
	pc.Lock()
	defer pc.Unlock()
	var hasIt bool
	if _, hasIt = pc.cache[dialURL]; !hasIt {
		if pstr, err := NewAMQPPoster(dialURL, attempts, fallbackFileDir); err != nil {
			return nil, err
		} else {
			pc.cache[dialURL] = pstr
		}
	}
	return pc.cache[dialURL], nil
}

// "amqp://guest:guest@localhost:5672/?queueID=cgrates_cdrs"
func NewAMQPPoster(dialURL string, attempts int, fallbackFileDir string) (*AMQPPoster, error) {
	u, err := url.Parse(dialURL)
	if err != nil {
		return nil, err
	}
	qry := u.Query()
	posterQueueID := "cgrates_cdrs"
	if vals, has := qry["queue_id"]; has && len(vals) != 0 {
		posterQueueID = vals[0]
	}
	dialURL = strings.Split(dialURL, "?")[0] // Take query params out of dialURL
	return &AMQPPoster{dialURL: dialURL, posterQueueID: posterQueueID,
		attempts: attempts, fallbackFileDir: fallbackFileDir}, nil
}

type AMQPPoster struct {
	dialURL         string
	posterQueueID   string // identifier of the CDR queue where we publish
	attempts        int
	fallbackFileDir string
	sync.Mutex      // protect connection
	conn            *amqp.Connection
}

// Post is the method being called when we need to post anything in the queue
// the optional chn will permits channel caching
func (pstr *AMQPPoster) Post(chn *amqp.Channel, contentType string, content []byte, fallbackFileName string) (*amqp.Channel, error) {
	var err error
	fib := Fib()
	if chn == nil {
		for i := 0; i < pstr.attempts; i++ {
			if chn, err = pstr.NewPostChannel(); err == nil {
				break
			}
			time.Sleep(time.Duration(fib()) * time.Second)
		}
		if err != nil && fallbackFileName != META_NONE {
			err = pstr.writeToFile(fallbackFileName, content)
			return nil, err
		}
	}
	for i := 0; i < pstr.attempts; i++ {
		if err = chn.Publish(
			"",                 // exchange
			pstr.posterQueueID, // routing key
			false,              // mandatory
			false,              // immediate
			amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				ContentType:  contentType,
				Body:         content,
			}); err == nil {
			break
		}
		time.Sleep(time.Duration(fib()) * time.Second)
	}
	if err != nil && fallbackFileName != META_NONE {
		err = pstr.writeToFile(fallbackFileName, content)
		return nil, err
	}
	return chn, nil
}

func (pstr *AMQPPoster) Close() {
	pstr.Lock()
	if pstr.conn != nil {
		pstr.conn.Close()
	}
	pstr.conn = nil
	pstr.Unlock()
}

func (pstr *AMQPPoster) NewPostChannel() (postChan *amqp.Channel, err error) {
	pstr.Lock()
	if pstr.conn == nil {
		var conn *amqp.Connection
		conn, err = amqp.Dial(pstr.dialURL)
		if err == nil {
			pstr.conn = conn
			go func() { // monitor connection errors so we can restart
				if err := <-pstr.conn.NotifyClose(make(chan *amqp.Error)); err != nil {
					Logger.Err(fmt.Sprintf("Connection error received: %s", err.Error()))
					pstr.Close()
				}
			}()
		}
	}
	pstr.Unlock()
	if err != nil {
		return nil, err
	}
	if postChan, err = pstr.conn.Channel(); err != nil {
		return
	}
	_, err = postChan.QueueDeclare(pstr.posterQueueID, true, false, false, false, nil)
	return
}

// writeToFile writes the content in the file with fileName on amqp.fallbackFileDir
func (pstr *AMQPPoster) writeToFile(fileName string, content []byte) (err error) {
	fallbackFilePath := path.Join(pstr.fallbackFileDir, fileName)
	_, err = guardian.Guardian.Guard(func() (interface{}, error) {
		fileOut, err := os.Create(fallbackFilePath)
		if err != nil {
			return nil, err
		}
		defer fileOut.Close()
		if _, err := fileOut.Write(content); err != nil {
			return nil, err
		}
		return nil, nil
	}, time.Duration(2*time.Second), FileLockPrefix+fallbackFilePath)
	return
}
