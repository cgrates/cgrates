// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"net/url"
	"strings"
	"sync"
)

const (
	defaultQueueID      = "cgrates_cdrs"
	defaultExchangeType = "direct"
	queueID             = "queue_id"
	exchange            = "exchange"
	exchangeType        = "exchange_type"
	routingKey          = "routing_key"

	awsToken   = "aws_token"
	folderPath = "folder_path"
)

func init() {
	PostersCache = &PosterCache{
		amqpCache:   make(map[string]Poster),
		amqpv1Cache: make(map[string]Poster),
		sqsCache:    make(map[string]Poster),
		kafkaCache:  make(map[string]Poster),
		s3Cache:     make(map[string]Poster),
	} // Initialize the cache for amqpPosters
}

var PostersCache *PosterCache

type PosterCache struct {
	sync.Mutex
	amqpCache   map[string]Poster
	amqpv1Cache map[string]Poster
	sqsCache    map[string]Poster
	kafkaCache  map[string]Poster
	s3Cache     map[string]Poster
}

type Poster interface {
	Post(body []byte, key string) error
	Close()
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

// Close closes all cached posters
func (pc *PosterCache) Close() {
	for _, v := range pc.amqpCache {
		v.Close()
	}
	for _, v := range pc.amqpv1Cache {
		v.Close()
	}
	for _, v := range pc.sqsCache {
		v.Close()
	}
	for _, v := range pc.kafkaCache {
		v.Close()
	}
}

// GetAMQPPoster creates a new poster only if not already cached
// uses dialURL as cache key
func (pc *PosterCache) GetAMQPPoster(dialURL string, attempts int) (pstr Poster, err error) {
	pc.Lock()
	defer pc.Unlock()
	if _, hasIt := pc.amqpCache[dialURL]; !hasIt {
		if pstr, err = NewAMQPPoster(dialURL, attempts); err != nil {
			return nil, err
		}
		pc.amqpCache[dialURL] = pstr
	}
	return pc.amqpCache[dialURL], nil
}

// GetAMQPv1Poster creates a new poster only if not already cached
func (pc *PosterCache) GetAMQPv1Poster(dialURL string, attempts int) (pstr Poster, err error) {
	pc.Lock()
	defer pc.Unlock()
	if _, hasIt := pc.amqpv1Cache[dialURL]; !hasIt {
		if pstr, err = NewAMQPv1Poster(dialURL, attempts); err != nil {
			return nil, err
		}
		pc.amqpv1Cache[dialURL] = pstr
	}
	return pc.amqpv1Cache[dialURL], nil
}

// GetSQSPoster creates a new poster only if not already cached
func (pc *PosterCache) GetSQSPoster(dialURL string, attempts int) (pstr Poster, err error) {
	pc.Lock()
	defer pc.Unlock()
	if _, hasIt := pc.sqsCache[dialURL]; !hasIt {
		if pstr, err = NewSQSPoster(dialURL, attempts); err != nil {
			return nil, err
		}
		pc.sqsCache[dialURL] = pstr
	}
	return pc.sqsCache[dialURL], nil
}

// GetKafkaPoster creates a new poster only if not already cached
func (pc *PosterCache) GetKafkaPoster(dialURL string, attempts int) (pstr Poster, err error) {
	pc.Lock()
	defer pc.Unlock()
	if _, hasIt := pc.kafkaCache[dialURL]; !hasIt {
		if pstr, err = NewKafkaPoster(dialURL, attempts); err != nil {
			return nil, err
		}
		pc.kafkaCache[dialURL] = pstr
	}
	return pc.kafkaCache[dialURL], nil
}

// GetS3Poster creates a new poster only if not already cached
func (pc *PosterCache) GetS3Poster(dialURL string, attempts int) (pstr Poster, err error) {
	pc.Lock()
	defer pc.Unlock()
	if _, hasIt := pc.s3Cache[dialURL]; !hasIt {
		if pstr, err = NewS3Poster(dialURL, attempts); err != nil {
			return nil, err
		}
		pc.s3Cache[dialURL] = pstr
	}
	return pc.s3Cache[dialURL], nil
}

func (pc *PosterCache) PostAMQP(dialURL string, attempts int,
	content []byte) error {
	amqpPoster, err := pc.GetAMQPPoster(dialURL, attempts)
	if err != nil {
		return err
	}
	return amqpPoster.Post(content, "")
}

func (pc *PosterCache) PostAMQPv1(dialURL string, attempts int,
	content []byte) error {
	AMQPv1Poster, err := pc.GetAMQPv1Poster(dialURL, attempts)
	if err != nil {
		return err
	}
	return AMQPv1Poster.Post(content, "")
}

func (pc *PosterCache) PostSQS(dialURL string, attempts int,
	content []byte) error {
	sqsPoster, err := pc.GetSQSPoster(dialURL, attempts)
	if err != nil {
		return err
	}
	return sqsPoster.Post(content, "")
}

func (pc *PosterCache) PostKafka(dialURL string, attempts int,
	content []byte, key string) error {
	kafkaPoster, err := pc.GetKafkaPoster(dialURL, attempts)
	if err != nil {
		return err
	}
	return kafkaPoster.Post(content, key)
}

func (pc *PosterCache) PostS3(dialURL string, attempts int,
	content []byte, key string) error {
	sqsPoster, err := pc.GetS3Poster(dialURL, attempts)
	if err != nil {
		return err
	}
	return sqsPoster.Post(content, key)
}
