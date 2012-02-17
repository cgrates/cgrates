package timespans

import (
	"launchpad.net/mgo"
	"launchpad.net/mgo/bson"
)

type KeyValue struct {
	_id   string
	value string
}

type MongoStorage struct {
	db      *mgo.Collection
	session *mgo.Session
}

func NewMongoStorage(address, db string) (*MongoStorage, error) {
	session, err := mgo.Dial(address)
	if err != nil {
		panic(err)
	}
	session.SetMode(mgo.Monotonic, true)

	ndb := session.DB(db).C("ap")
	return &MongoStorage{db: ndb, session: session}, nil
}

func (ms *MongoStorage) Close() {
	ms.session.Close()
}

func (ms *MongoStorage) Get(key string) (string, error) {
	result := KeyValue{}
	err := ms.db.Find(bson.M{"_id": key}).One(&result)
	return result.value, err
}
