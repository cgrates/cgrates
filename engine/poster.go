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
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
	"github.com/streadway/amqp"
	amqpv1 "github.com/vcabbage/amqp"
)

var AMQPQuery = []string{"cacertfile", "certfile", "keyfile", "verify", "server_name_indication", "auth_mechanism", "heartbeat", "connection_timeout", "channel_max"}

const (
	defaultQueueID      = "cgrates_cdrs"
	defaultExchangeType = "direct"
	queueID             = "queue_id"
	exchange            = "exchange"
	exchangeType        = "exchange_type"
	routingKey          = "routing_key"

	awsRegion    = "aws_region"
	awsID        = "aws_key"
	awsSecret    = "aws_secret"
	awsToken     = "aws_token"
	awsAccountID = "aws_account_id"
)

func init() {
	PostersCache = &PosterCache{
		amqpCache: make(map[string]Poster),
		awsCache:  make(map[string]Poster),
		sqsCache:  make(map[string]Poster),
	} // Initialize the cache for amqpPosters
}

var PostersCache *PosterCache

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

type Poster interface {
	Post([]byte, string) error
	Close()
}

func writeToFile(fileDir, fileName string, content []byte) (err error) {
	fallbackFilePath := path.Join(fileDir, fileName)
	_, err = guardian.Guardian.Guard(func() (interface{}, error) {
		fileOut, err := os.Create(fallbackFilePath)
		if err != nil {
			return nil, err
		}
		_, err = fileOut.Write(content)
		fileOut.Close()
		return nil, err
	}, time.Duration(2*time.Second), utils.FileLockPrefix+fallbackFilePath)
	return
}

func parseURL(dialURL string) (URL string, qID string, err error) {
	u, err := url.Parse(dialURL)
	if err != nil {
		return "", "", err
	}
	qry := u.Query()
	URL = strings.Split(dialURL, "?")[0]
	qID = defaultQueueID
	if vals, has := qry[queueID]; has && len(vals) != 0 {
		qID = vals[0]
	}
	return
}

type PosterCache struct {
	sync.Mutex
	amqpCache map[string]Poster
	awsCache  map[string]Poster
	sqsCache  map[string]Poster
}

func (pc *PosterCache) Close() {
	for _, v := range pc.amqpCache {
		v.Close()
	}
	for _, v := range pc.awsCache {
		v.Close()
	}
	for _, v := range pc.sqsCache {
		v.Close()
	}
}

// GetAMQPPoster creates a new poster only if not already cached
// uses dialURL as cache key
func (pc *PosterCache) GetAMQPPoster(dialURL string, attempts int, fallbackFileDir string) (Poster, error) {
	pc.Lock()
	defer pc.Unlock()
	if _, hasIt := pc.amqpCache[dialURL]; !hasIt {
		if pstr, err := NewAMQPPoster(dialURL, attempts, fallbackFileDir); err != nil {
			return nil, err
		} else {
			pc.amqpCache[dialURL] = pstr
		}
	}
	return pc.amqpCache[dialURL], nil
}

func (pc *PosterCache) GetAWSPoster(dialURL string, attempts int, fallbackFileDir string) (Poster, error) {
	pc.Lock()
	defer pc.Unlock()
	if _, hasIt := pc.awsCache[dialURL]; !hasIt {
		if pstr, err := NewAWSPoster(dialURL, attempts, fallbackFileDir); err != nil {
			return nil, err
		} else {
			pc.awsCache[dialURL] = pstr
		}
	}
	return pc.awsCache[dialURL], nil
}

func (pc *PosterCache) GetSQSPoster(dialURL string, attempts int, fallbackFileDir string) (Poster, error) {
	pc.Lock()
	defer pc.Unlock()
	if _, hasIt := pc.sqsCache[dialURL]; !hasIt {
		if pstr, err := NewSQSPoster(dialURL, attempts, fallbackFileDir); err != nil {
			return nil, err
		} else {
			pc.sqsCache[dialURL] = pstr
		}
	}
	return pc.sqsCache[dialURL], nil
}

func (pc *PosterCache) PostAMQP(dialURL string, attempts int,
	content []byte, contentType, fallbackFileDir, fallbackFileName string) error {
	amqpPoster, err := pc.GetAMQPPoster(dialURL, attempts, fallbackFileDir)
	if err != nil {
		return err
	}
	return amqpPoster.Post(content, fallbackFileName)
}

func (pc *PosterCache) PostAWS(dialURL string, attempts int,
	content []byte, fallbackFileDir, fallbackFileName string) error {
	awsPoster, err := pc.GetAWSPoster(dialURL, attempts, fallbackFileDir)
	if err != nil {
		return err
	}
	return awsPoster.Post(content, fallbackFileName)
}

func (pc *PosterCache) PostSQS(dialURL string, attempts int,
	content []byte, fallbackFileDir, fallbackFileName string) error {
	sqsPoster, err := pc.GetSQSPoster(dialURL, attempts, fallbackFileDir)
	if err != nil {
		return err
	}
	return sqsPoster.Post(content, fallbackFileName)
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
func (pstr *AMQPPoster) Post(content []byte, fallbackFileName string) (err error) {
	var chn *amqp.Channel
	fib := utils.Fib()

	for i := 0; i < pstr.attempts; i++ {
		if chn, err = pstr.newPostChannel(); err == nil {
			break
		}
		time.Sleep(time.Duration(fib()) * time.Second)
	}
	if err != nil {
		if fallbackFileName != utils.META_NONE {
			utils.Logger.Warning(fmt.Sprintf("<AMQPPoster> creating new post channel, err: %s", err.Error()))
			err = writeToFile(pstr.fallbackFileDir, fallbackFileName, content)
		}
		return err
	}
	for i := 0; i < pstr.attempts; i++ {
		if err = chn.Publish(
			pstr.exchange,   // exchange
			pstr.routingKey, // routing key
			false,           // mandatory
			false,           // immediate
			amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				ContentType:  utils.CONTENT_JSON,
				Body:         content,
			}); err == nil {
			break
		}
		time.Sleep(time.Duration(fib()) * time.Second)
	}
	if err != nil && fallbackFileName != utils.META_NONE {
		err = writeToFile(pstr.fallbackFileDir, fallbackFileName, content)
		return err
	}
	if chn != nil {
		chn.Close()
	}
	return
}

func (pstr *AMQPPoster) Close() {
	pstr.Lock()
	if pstr.conn != nil {
		pstr.conn.Close()
	}
	pstr.conn = nil
	pstr.Unlock()
}

func (pstr *AMQPPoster) newPostChannel() (postChan *amqp.Channel, err error) {
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

func NewAWSPoster(dialURL string, attempts int, fallbackFileDir string) (Poster, error) {
	URL, qID, err := parseURL(dialURL)
	if err != nil {
		return nil, err
	}
	return &AWSPoster{
		dialURL:         URL,
		queueID:         "/" + qID,
		attempts:        attempts,
		fallbackFileDir: fallbackFileDir,
	}, nil
}

type AWSPoster struct {
	sync.Mutex
	dialURL         string
	queueID         string // identifier of the CDR queue where we publish
	attempts        int
	fallbackFileDir string
	client          *amqpv1.Client
}

func (pstr *AWSPoster) Close() {
	pstr.Lock()
	if pstr.client != nil {
		pstr.client.Close()
	}
	pstr.client = nil
	pstr.Unlock()
}

func (pstr *AWSPoster) Post(content []byte, fallbackFileName string) (err error) {
	var s *amqpv1.Session
	fib := utils.Fib()

	for i := 0; i < pstr.attempts; i++ {
		if s, err = pstr.newPosterSession(); err == nil {
			break
		}
		// reset client and try again
		// used in case of closed conection because of idle time
		pstr.client = nil
		time.Sleep(time.Duration(fib()) * time.Second)
	}
	if err != nil {
		if fallbackFileName != utils.META_NONE {
			utils.Logger.Warning(fmt.Sprintf("<AWSPoster> creating new post channel, err: %s", err.Error()))
			err = writeToFile(pstr.fallbackFileDir, fallbackFileName, content)
		}
		return err
	}

	ctx := context.Background()
	for i := 0; i < pstr.attempts; i++ {
		sender, err := s.NewSender(
			amqpv1.LinkTargetAddress(pstr.queueID),
		)
		if err != nil {
			time.Sleep(time.Duration(fib()) * time.Second)
			// if pstr.isRecoverableError(err) {
			// 	s.Close(ctx)
			// 	pstr.client.Close()
			// 	pstr.client = nil
			// 	stmp, err := pstr.newPosterSession()
			// 	if err == nil {
			// 		s = stmp
			// 	}
			// }
			continue
		}
		// Send message
		err = sender.Send(ctx, amqpv1.NewMessage(content))
		sender.Close(ctx)
		if err == nil {
			break
		}
		time.Sleep(time.Duration(fib()) * time.Second)
		// if pstr.isRecoverableError(err) {
		// 	s.Close(ctx)
		// 	pstr.client.Close()
		// 	pstr.client = nil
		// 	stmp, err := pstr.newPosterSession()
		// 	if err == nil {
		// 		s = stmp
		// 	}
		// }
	}
	if err != nil && fallbackFileName != utils.META_NONE {
		err = writeToFile(pstr.fallbackFileDir, fallbackFileName, content)
		return err
	}
	if s != nil {
		s.Close(ctx)
	}
	return
}

func (pstr *AWSPoster) newPosterSession() (s *amqpv1.Session, err error) {
	pstr.Lock()
	defer pstr.Unlock()
	if pstr.client == nil {
		var client *amqpv1.Client
		client, err = amqpv1.Dial(pstr.dialURL)
		if err != nil {
			return nil, err
		}
		pstr.client = client
	}
	return pstr.client.NewSession()
}

func isRecoverableCloseError(err error) bool {
	return err == amqpv1.ErrConnClosed ||
		err == amqpv1.ErrLinkClosed ||
		err == amqpv1.ErrSessionClosed
}

func (pstr *AWSPoster) isRecoverableError(err error) bool {
	switch err.(type) {
	case *amqpv1.Error, *amqpv1.DetachError, net.Error:
		if netErr, ok := err.(net.Error); ok {
			if !netErr.Temporary() {
				return false
			}
		}
	default:
		if !isRecoverableCloseError(err) {
			return false
		}
	}
	return true
}

func NewSQSPoster(dialURL string, attempts int, fallbackFileDir string) (Poster, error) {
	pstr := &SQSPoster{
		attempts:        attempts,
		fallbackFileDir: fallbackFileDir,
	}
	err := pstr.parseURL(dialURL)
	if err != nil {
		return nil, err
	}
	return pstr, nil
}

type SQSPoster struct {
	sync.Mutex
	dialURL         string
	awsRegion       string
	awsID           string
	awsKey          string
	awsToken        string
	attempts        int
	fallbackFileDir string
	queueURL        *string
	queueID         string
	// getQueueOnce    sync.Once
	session *session.Session
}

func (pstr *SQSPoster) Close() {}

func (pstr *SQSPoster) parseURL(dialURL string) (err error) {
	u, err := url.Parse(dialURL)
	if err != nil {
		return err
	}
	qry := u.Query()

	pstr.dialURL = strings.Split(dialURL, "?")[0]
	pstr.dialURL = strings.TrimSuffix(pstr.dialURL, "/") // used to remove / to point to correct endpoint
	pstr.queueID = defaultQueueID
	if vals, has := qry[queueID]; has && len(vals) != 0 {
		pstr.queueID = vals[0]
	}
	if vals, has := qry[awsRegion]; has && len(vals) != 0 {
		pstr.awsRegion = url.QueryEscape(vals[0])
	} else {
		utils.Logger.Warning("<SQSPoster> No region present for AWS.")
	}
	if vals, has := qry[awsID]; has && len(vals) != 0 {
		pstr.awsID = url.QueryEscape(vals[0])
	} else {
		utils.Logger.Warning("<SQSPoster> No access key ID present for AWS.")
	}
	if vals, has := qry[awsSecret]; has && len(vals) != 0 {
		pstr.awsKey = url.QueryEscape(vals[0])
	} else {
		utils.Logger.Warning("<SQSPoster> No secret access key present for AWS.")
	}
	if vals, has := qry[awsToken]; has && len(vals) != 0 {
		pstr.awsToken = url.QueryEscape(vals[0])
	} else {
		utils.Logger.Warning("<SQSPoster> No session token present for AWS.")
	}
	pstr.getQueueURL()
	return nil
}

func (pstr *SQSPoster) getQueueURL() (err error) {
	if pstr.queueURL != nil {
		return nil
	}
	// pstr.getQueueOnce.Do(func() {
	var svc *sqs.SQS
	if svc, err = pstr.newPosterSession(); err != nil {
		return
	}
	var result *sqs.GetQueueUrlOutput
	if result, err = svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(pstr.queueID),
	}); err == nil {
		pstr.queueURL = new(string)
		*(pstr.queueURL) = *result.QueueUrl
		return
	}
	if aerr, ok := err.(awserr.Error); ok && aerr.Code() == sqs.ErrCodeQueueDoesNotExist {
		// For CreateQueue
		var createResult *sqs.CreateQueueOutput
		if createResult, err = svc.CreateQueue(&sqs.CreateQueueInput{
			QueueName: aws.String(pstr.queueID),
		}); err == nil {
			pstr.queueURL = new(string)
			*(pstr.queueURL) = *createResult.QueueUrl
			return
		}
	}
	utils.Logger.Warning(fmt.Sprintf("<SQSPoster> can not get url for queue with ID=%s because err: %v", pstr.queueID, err))
	// })
	return err
}

func (pstr *SQSPoster) Post(message []byte, fallbackFileName string) (err error) {
	var svc *sqs.SQS
	fib := utils.Fib()

	for i := 0; i < pstr.attempts; i++ {
		if svc, err = pstr.newPosterSession(); err == nil {
			break
		}
		time.Sleep(time.Duration(fib()) * time.Second)
	}
	if err != nil {
		if fallbackFileName != utils.META_NONE {
			utils.Logger.Warning(fmt.Sprintf("<SQSPoster> creating new session, err: %s", err.Error()))
			err = writeToFile(pstr.fallbackFileDir, fallbackFileName, message)
		}
		return err
	}

	for i := 0; i < pstr.attempts; i++ {
		if _, err = svc.SendMessage(
			&sqs.SendMessageInput{
				MessageBody: aws.String(string(message)),
				QueueUrl:    pstr.queueURL,
			},
		); err == nil {
			break
		}
	}
	if err != nil && fallbackFileName != utils.META_NONE {
		utils.Logger.Warning(fmt.Sprintf("<SQSPoster> posting new message, err: %s", err.Error()))
		err = writeToFile(pstr.fallbackFileDir, fallbackFileName, message)
	}
	return err
}

func (pstr *SQSPoster) newPosterSession() (s *sqs.SQS, err error) {
	pstr.Lock()
	defer pstr.Unlock()
	if pstr.session == nil {
		var ses *session.Session
		cfg := aws.Config{Endpoint: aws.String(pstr.dialURL)}
		if len(pstr.awsRegion) != 0 {
			cfg.Region = aws.String(pstr.awsRegion)
		}
		if len(pstr.awsID) != 0 &&
			len(pstr.awsKey) != 0 {
			cfg.Credentials = credentials.NewStaticCredentials(pstr.awsID, pstr.awsKey, pstr.awsToken)
		}
		ses, err = session.NewSessionWithOptions(
			session.Options{
				Config: cfg,
			},
		)
		if err != nil {
			return nil, err
		}
		pstr.session = ses
	}
	return sqs.New(pstr.session), nil
}
