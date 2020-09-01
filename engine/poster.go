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
	"sync"
)

// General constants for posters
const (
	DefaultQueueID      = "cgrates_cdrs"
	QueueID             = "queue_id"
	DefaultExchangeType = "direct"
	Exchange            = "exchange"
	ExchangeType        = "exchange_type"
	RoutingKey          = "routing_key"

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
func (pc *PosterCache) GetAMQPPoster(dialURL string, attempts int) (pstr Poster) {
	pc.Lock()
	defer pc.Unlock()
	if _, hasIt := pc.amqpCache[dialURL]; !hasIt {
		pstr = NewAMQPPoster(dialURL, attempts, nil)
		pc.amqpCache[dialURL] = pstr
	}
	return pc.amqpCache[dialURL]
}

// GetAMQPv1Poster creates a new poster only if not already cached
func (pc *PosterCache) GetAMQPv1Poster(dialURL string, attempts int) (pstr Poster) {
	pc.Lock()
	defer pc.Unlock()
	if _, hasIt := pc.amqpv1Cache[dialURL]; !hasIt {
		pstr = NewAMQPv1Poster(dialURL, attempts, nil)
		pc.amqpv1Cache[dialURL] = pstr
	}
	return pc.amqpv1Cache[dialURL]
}

// GetSQSPoster creates a new poster only if not already cached
func (pc *PosterCache) GetSQSPoster(dialURL string, attempts int) (pstr Poster) {
	pc.Lock()
	defer pc.Unlock()
	if _, hasIt := pc.sqsCache[dialURL]; !hasIt {
		pstr = NewSQSPoster(dialURL, attempts, nil)
		pc.sqsCache[dialURL] = pstr
	}
	return pc.sqsCache[dialURL]
}

// GetKafkaPoster creates a new poster only if not already cached
func (pc *PosterCache) GetKafkaPoster(dialURL string, attempts int) (pstr Poster) {
	pc.Lock()
	defer pc.Unlock()
	if _, hasIt := pc.kafkaCache[dialURL]; !hasIt {
		pstr = NewKafkaPoster(dialURL, attempts, nil)
		pc.kafkaCache[dialURL] = pstr
	}
	return pc.kafkaCache[dialURL]
}

// GetS3Poster creates a new poster only if not already cached
func (pc *PosterCache) GetS3Poster(dialURL string, attempts int) (pstr Poster) {
	pc.Lock()
	defer pc.Unlock()
	if _, hasIt := pc.s3Cache[dialURL]; !hasIt {
		pstr = NewS3Poster(dialURL, attempts, nil)
		pc.s3Cache[dialURL] = pstr
	}
	return pc.s3Cache[dialURL]
}

func (pc *PosterCache) PostAMQP(dialURL string, attempts int,
	content []byte) error {
	return pc.GetAMQPPoster(dialURL, attempts).Post(content, "")
}

func (pc *PosterCache) PostAMQPv1(dialURL string, attempts int,
	content []byte) error {
	return pc.GetAMQPv1Poster(dialURL, attempts).Post(content, "")
}

func (pc *PosterCache) PostSQS(dialURL string, attempts int,
	content []byte) error {
	return pc.GetSQSPoster(dialURL, attempts).Post(content, "")
}

func (pc *PosterCache) PostKafka(dialURL string, attempts int,
	content []byte, key string) error {
	return pc.GetKafkaPoster(dialURL, attempts).Post(content, key)
}

func (pc *PosterCache) PostS3(dialURL string, attempts int,
	content []byte, key string) error {
	return pc.GetS3Poster(dialURL, attempts).Post(content, key)
}
