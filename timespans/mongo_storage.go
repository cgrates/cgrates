package timespans

import (
	"launchpad.net/mgo"
	"launchpad.net/mgo/bson"
)

type KeyValue struct {
	Key               string
	ActivationPeriods []*ActivationPeriod
}

type MongoStorage struct {
	db      *mgo.Database
	session *mgo.Session
}

func NewMongoStorage(address, db string) (*MongoStorage, error) {
	session, err := mgo.Dial(address)
	if err != nil {
		panic(err)
	}
	session.SetMode(mgo.Monotonic, true)

	index := mgo.Index{Key: []string{"key"}, Unique: true, DropDups: true, Background: true}
	err = session.DB(db).C("ap").EnsureIndex(index)
	index = mgo.Index{Key: []string{"id"}, Unique: true, DropDups: true, Background: true}
	err = session.DB(db).C("dest").EnsureIndex(index)
	index = mgo.Index{Key: []string{"id"}, Unique: true, DropDups: true, Background: true}
	err = session.DB(db).C("tp").EnsureIndex(index)

	return &MongoStorage{db: session.DB(db), session: session}, nil
}

func (ms *MongoStorage) Close() {
	ms.session.Close()
}

func (ms *MongoStorage) GetActivationPeriods(key string) (aps []*ActivationPeriod, err error) {
	ndb := ms.db.C("ap")
	result := KeyValue{}
	err = ndb.Find(bson.M{"key": key}).One(&result)
	return result.ActivationPeriods, err
}

func (ms *MongoStorage) SetActivationPeriods(key string, aps []*ActivationPeriod) {
	ndb := ms.db.C("ap")
	ndb.Insert(&KeyValue{key, aps})
}

func (ms *MongoStorage) GetDestination(key string) (result *Destination, err error) {
	ndb := ms.db.C("dest")
	result = &Destination{}
	err = ndb.Find(bson.M{"id": key}).One(result)
	return
}

func (ms *MongoStorage) SetDestination(dest *Destination) {
	ndb := ms.db.C("dest")
	ndb.Insert(&dest)
}

func (ms *MongoStorage) GetTariffPlan(key string) (result *TariffPlan, err error) {
	ndb := ms.db.C("tp")
	result = &TariffPlan{}
	err = ndb.Find(bson.M{"id": key}).One(result)
	return
}

func (ms *MongoStorage) SetTariffPlan(tp *TariffPlan) {
	ndb := ms.db.C("tp")
	ndb.Insert(&tp)
}
