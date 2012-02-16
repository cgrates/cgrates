package timespans

import (
        "launchpad.net/mgo/bson"
        "launchpad.net/mgo"
)

type KeyValue struct {
        Key string
        Value string
}

type MongoStorage struct {
	db *mgo.Collection
	session *mgo.Session
}

func NewMongoStorage(address, db string) (*MongoStorage, error) {
	session, err := mgo.Dial(address)
    if err != nil {
            panic(err)
    }
    session.SetMode(mgo.Monotonic, true)

    ndb := session.DB(db).C("ap")
	//log.Print("Starting redis storage")
	return &MongoStorage{db: ndb, session: session}, nil
}

func (ms *MongoStorage) Close() {
	//log.Print("Closing redis storage")
	ms.session.Close()
}

func (ms *MongoStorage) Get(key string) (string, error) {
	result := KeyValue{}
	err := ms.db.Find(bson.M{"key": key}).One(&result)

	return result.Value, err
}
