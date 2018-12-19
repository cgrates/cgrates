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
	"path"
	"strings"
	"sync"
	"time"

	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
	"github.com/streadway/amqp"
)

var AMQPQuery = []string{"cacertfile", "certfile", "keyfile", "verify", "server_name_indication", "auth_mechanism", "heartbeat", "connection_timeout", "channel_max"}

const (
	defaultQueueID      = "cgrates_cdrs"
	defaultExchangeType = "direct"
	queueID             = "queue_id"
	exchange            = "exchange"
	exchangeType        = "exchange_type"
	routingKey          = "routing_key"
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
			defer fileOut.Close()
			if _, err := fileOut.Write(body); err != nil {
				return nil, err
			}
			return nil, nil
		}, time.Duration(2*time.Second), utils.FileLockPrefix+fallbackFilePath)
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
func (pc *AMQPCachedPosters) GetAMQPPoster(dialURL string, attempts int,
	fallbackFileDir string) (amqpPoster *AMQPPoster, err error) {
	pc.Lock()
	defer pc.Unlock()
	if _, hasIt := pc.cache[dialURL]; !hasIt {
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
	amqp := &AMQPPoster{
		attempts:        attempts,
		fallbackFileDir: fallbackFileDir,
	}
	if err := amqp.parseURL(dialURL); err != nil {
		return nil, err
	}
	return amqp, nil
}

type AMQPPoster struct {
	dialURL         string
	queueID         string // identifier of the CDR queue where we publish
	exchange        string
	exchangeType    string
	routingKey      string
	attempts        int
	fallbackFileDir string
	sync.Mutex      // protect connection
	conn            *amqp.Connection
}

func (pstr *AMQPPoster) parseURL(dialURL string) error {
	u, err := url.Parse(dialURL)
	if err != nil {
		return err
	}
	qry := u.Query()
	q := url.Values{}
	for _, key := range AMQPQuery {
		if vals, has := qry[key]; has && len(vals) != 0 {
			q.Add(key, vals[0])
		}
	}
	pstr.dialURL = strings.Split(dialURL, "?")[0] + "?" + q.Encode()
	pstr.queueID = defaultQueueID
	pstr.routingKey = defaultQueueID
	if vals, has := qry[queueID]; has && len(vals) != 0 {
		pstr.queueID = vals[0]
	}
	if vals, has := qry[routingKey]; has && len(vals) != 0 {
		pstr.routingKey = vals[0]
	}
	if vals, has := qry[exchange]; has && len(vals) != 0 {
		pstr.exchange = vals[0]
		pstr.exchangeType = defaultExchangeType
	}
	if vals, has := qry[exchangeType]; has && len(vals) != 0 {
		pstr.exchangeType = vals[0]
	}
	return nil
}

// Post is the method being called when we need to post anything in the queue
// the optional chn will permits channel caching
func (pstr *AMQPPoster) Post(chn *amqp.Channel, contentType string, content []byte,
	fallbackFileName string) (rChn *amqp.Channel, err error) {
	fib := utils.Fib()
	if chn == nil {
		for i := 0; i < pstr.attempts; i++ {
			if chn, err = pstr.NewPostChannel(); err == nil {
				break
			}
			time.Sleep(time.Duration(fib()) * time.Second)
		}
		if err != nil &&
			fallbackFileName != utils.META_NONE {
			utils.Logger.Warning(fmt.Sprintf("<AMQPPoster> creating new post channel, err: %s", err.Error()))
			err = pstr.writeToFile(fallbackFileName, content)
			return nil, err
		}
	}
	if err != nil {
		return
	}
	for i := 0; i < pstr.attempts; i++ {
		if err = chn.Publish(
			pstr.exchange,   // exchange
			pstr.routingKey, // routing key
			false,           // mandatory
			false,           // immediate
			amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				ContentType:  contentType,
				Body:         content,
			}); err == nil {
			break
		}
		time.Sleep(time.Duration(fib()) * time.Second)
	}
	if err != nil && fallbackFileName != utils.META_NONE {
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
					utils.Logger.Err(fmt.Sprintf("Connection error received: %s", err.Error()))
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

	if pstr.exchange != "" {
		if err = postChan.ExchangeDeclare(
			pstr.exchange,     // name
			pstr.exchangeType, // type
			true,              // durable
			false,             // audo-delete
			false,             // internal
			false,             // no-wait
			nil,               // args
		); err != nil {
			return
		}
	}

	if _, err = postChan.QueueDeclare(
		pstr.queueID, // name
		true,         // durable
		false,        // auto-delete
		false,        // exclusive
		false,        // no-wait
		nil,          // args
	); err != nil {
		return
	}

	if pstr.exchange != "" {
		if err = postChan.QueueBind(
			pstr.queueID,    // queue
			pstr.routingKey, // key
			pstr.exchange,   // exchange
			false,           // no-wait
			nil,             // args
		); err != nil {
			return
		}
	}
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
	}, time.Duration(2*time.Second), utils.FileLockPrefix+fallbackFilePath)
	return
}
