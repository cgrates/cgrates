package timespans

import (
	"launchpad.net/mgo"
	"launchpad.net/mgo/bson"
)

type KeyValue struct {
	Key   string
	ActivationPeriods []*ActivationPeriod
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

func (ms *MongoStorage) GetActivationPeriods(key string) (aps []*ActivationPeriod, err error) {
	result := KeyValue{}
	err = ms.db.Find(bson.M{"key": key}).One(&result)
	return result.ActivationPeriods, err
}

func (ms *MongoStorage) SetActivationPeriods(key string, aps []*ActivationPeriod){
	ms.db.Insert(&KeyValue{key, aps})
}
