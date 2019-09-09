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
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
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
	Post(body []byte, fallbackName, key string) error
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

func (pc *PosterCache) GetAMQPv1Poster(dialURL string, attempts int, fallbackFileDir string) (Poster, error) {
	pc.Lock()
	defer pc.Unlock()
	if _, hasIt := pc.amqpv1Cache[dialURL]; !hasIt {
		if pstr, err := NewAMQPv1Poster(dialURL, attempts, fallbackFileDir); err != nil {
			return nil, err
		} else {
			pc.amqpv1Cache[dialURL] = pstr
		}
	}
	return pc.amqpv1Cache[dialURL], nil
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

func (pc *PosterCache) GetKafkaPoster(dialURL string, attempts int, fallbackFileDir string) (Poster, error) {
	pc.Lock()
	defer pc.Unlock()
	if _, hasIt := pc.kafkaCache[dialURL]; !hasIt {
		if pstr, err := NewKafkaPoster(dialURL, attempts, fallbackFileDir); err != nil {
			return nil, err
		} else {
			pc.kafkaCache[dialURL] = pstr
		}
	}
	return pc.kafkaCache[dialURL], nil
}

func (pc *PosterCache) GetS3Poster(dialURL string, attempts int, fallbackFileDir string) (Poster, error) {
	pc.Lock()
	defer pc.Unlock()
	if _, hasIt := pc.s3Cache[dialURL]; !hasIt {
		if pstr, err := NewS3Poster(dialURL, attempts, fallbackFileDir); err != nil {
			return nil, err
		} else {
			pc.s3Cache[dialURL] = pstr
		}
	}
	return pc.s3Cache[dialURL], nil
}

func (pc *PosterCache) PostAMQP(dialURL string, attempts int,
	content []byte, contentType, fallbackFileDir, fallbackFileName string) error {
	amqpPoster, err := pc.GetAMQPPoster(dialURL, attempts, fallbackFileDir)
	if err != nil {
		return err
	}
	return amqpPoster.Post(content, fallbackFileName, "")
}

func (pc *PosterCache) PostAMQPv1(dialURL string, attempts int,
	content []byte, fallbackFileDir, fallbackFileName string) error {
	AMQPv1Poster, err := pc.GetAMQPv1Poster(dialURL, attempts, fallbackFileDir)
	if err != nil {
		return err
	}
	return AMQPv1Poster.Post(content, fallbackFileName, "")
}

func (pc *PosterCache) PostSQS(dialURL string, attempts int,
	content []byte, fallbackFileDir, fallbackFileName string) error {
	sqsPoster, err := pc.GetSQSPoster(dialURL, attempts, fallbackFileDir)
	if err != nil {
		return err
	}
	return sqsPoster.Post(content, fallbackFileName, "")
}

func (pc *PosterCache) PostKafka(dialURL string, attempts int,
	content []byte, fallbackFileDir, fallbackFileName, key string) error {
	kafkaPoster, err := pc.GetKafkaPoster(dialURL, attempts, fallbackFileDir)
	if err != nil {
		return err
	}
	return kafkaPoster.Post(content, fallbackFileName, key)
}

func (pc *PosterCache) PostS3(dialURL string, attempts int,
	content []byte, fallbackFileDir, fallbackFileName, key string) error {
	sqsPoster, err := pc.GetS3Poster(dialURL, attempts, fallbackFileDir)
	if err != nil {
		return err
	}
	return sqsPoster.Post(content, fallbackFileName, key)
}
