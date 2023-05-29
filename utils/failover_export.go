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

/*
var failedPostCache *ltcache.Cache

func init() {
	failedPostCache = ltcache.NewCache(-1, 5*time.Second, true, writeFailedPosts)
}

// SetFailedPostCacheTTL recreates the failed cache
func SetFailedPostCacheTTL(ttl time.Duration) {
	failedPostCache = ltcache.NewCache(-1, ttl, true, writeFailedPosts)
}

func writeFailedPosts(_ string, value any) {
	expEv, canConvert := value.(*FailedExportersLogg)
	if !canConvert {
		return
	}
	filePath := expEv.FilePath()
	expEv.lk.RLock()
	if err := WriteToFile(filePath, expEv); err != nil {
		Logger.Warning(fmt.Sprintf("Unable to write failed post to file <%s> because <%s>",
			filePath, err))
		expEv.lk.RUnlock()
		return
	}
	expEv.lk.RUnlock()
}

// FilePath returns the file path it should use for saving the failed events
func (expEv *FailedExportersLogg) FilePath() string {
	return path.Join(expEv.FailedPostsDir, expEv.Module+PipeSep+UUIDSha1Prefix()+GOBSuffix)
}

// WriteToFile writes the events to file
func WriteToFile(filePath string, expEv FailoverPoster) (err error) {
	fileOut, err := os.Create(filePath)
	if err != nil {
		return err
	}
	encd := gob.NewEncoder(fileOut)
	gob.Register(new(CGREvent))
	err = encd.Encode(expEv)
	fileOut.Close()
	return
}

type FailedExportersLogg struct {
	lk             sync.RWMutex
	Path           string
	Opts           map[string]any // THIS WILL BE META
	Format         string
	Events         []any
	FailedPostsDir string
	Module         string
}

func AddFailedMessage(failedPostsDir, expPath, format,
	module string, ev any, opts map[string]any) {
	key := ConcatenatedKey(failedPostsDir, expPath, format, module)
	switch module {
	case EEs:
		// also in case of amqp,amqpv1,s3,sqs and kafka also separe them after queue id
		var amqpQueueID string
		var s3BucketID string
		var sqsQueueID string
		var kafkaTopic string
		if _, has := opts[AMQPQueueID]; has {
			amqpQueueID = IfaceAsString(opts[AMQPQueueID])
		}
		if _, has := opts[S3Bucket]; has {
			s3BucketID = IfaceAsString(opts[S3Bucket])
		}
		if _, has := opts[SQSQueueID]; has {
			sqsQueueID = IfaceAsString(opts[SQSQueueID])
		}
		if _, has := opts[kafkaTopic]; has {
			kafkaTopic = IfaceAsString(opts[KafkaTopic])
		}
		if qID := FirstNonEmpty(amqpQueueID, s3BucketID,
			sqsQueueID, kafkaTopic); len(qID) != 0 {
			key = ConcatenatedKey(key, qID)
		}
	case Kafka:
	}
	var failedPost *FailedExportersLogg
	if x, ok := failedPostCache.Get(key); ok {
		if x != nil {
			failedPost = x.(*FailedExportersLogg)
		}
	}
	if failedPost == nil {
		failedPost = &FailedExportersLogg{
			Path:           expPath,
			Format:         format,
			Opts:           opts,
			Module:         module,
			FailedPostsDir: failedPostsDir,
		}
		failedPostCache.Set(key, failedPost, nil)
	}
	failedPost.AddEvent(ev)
}

// AddEvent adds one event
func (expEv *FailedExportersLogg) AddEvent(ev any) {
	expEv.lk.Lock()
	expEv.Events = append(expEv.Events, ev)
	expEv.lk.Unlock()
}

// NewExportEventsFromFile returns ExportEvents from the file
// used only on replay failed post
func NewExportEventsFromFile(filePath string) (expEv *FailedExportersLogg, err error) {
	var fileContent []byte
	if fileContent, err = os.ReadFile(filePath); err != nil {
		return nil, err
	}
	if err = os.Remove(filePath); err != nil {
		return nil, err
	}
	dec := gob.NewDecoder(bytes.NewBuffer(fileContent))
	// unmarshall it
	expEv = new(FailedExportersLogg)
	err = dec.Decode(&expEv)
	return
}

type FailoverPoster interface {
	ReplayFailedPosts(int, string) error
}

// ReplayFailedPosts tryies to post cdrs again
func (expEv *FailedExportersLogg) ReplayFailedPosts(attempts int, tnt string) (err error) {
	nodeID := IfaceAsString(expEv.Opts[NodeID])
	logLvl, err := IfaceAsInt(expEv.Opts[Level])
	if err != nil {
		return
	}
	expLogger := NewExportLogger(nodeID, tnt, logLvl,
		expEv.Path, expEv.Format, attempts, expEv.FailedPostsDir)
	for _, event := range expEv.Events {
		var content []byte
		if content, err = ToUnescapedJSON(event); err != nil {
			return
		}
		if err = expLogger.Writer.WriteMessages(context.Background(), kafka.Message{
			Key:   []byte(GenUUID()),
			Value: content,
		}); err != nil {
			// if there are any errors in kafka, we will post in FailedPostDirectory
			AddFailedMessage(expLogger.FldPostDir, expLogger.Writer.Addr.String(), MetaKafkaLog, Kafka,
				event, expLogger.GetMeta())
			return nil
		}
	}
	return err
}
*/
