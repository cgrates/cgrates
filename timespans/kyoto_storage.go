package timespans

import (
	//"log"
	"github.com/fsouza/gokabinet/kc"
)

type KyotoStorage struct {
	db *kc.DB
}

func NewKyotoStorage(filaName string) (*KyotoStorage, error) {
	ndb, err := kc.Open(filaName, kc.READ)
	//log.Print("Starting kyoto storage")
	return &KyotoStorage{db: ndb}, err	
}


func (ks *KyotoStorage) Close() {
	//log.Print("Closing kyoto storage")
	ks.db.Close()
}

func (ks *KyotoStorage) Get(key string) (value string, err error) {
	return ks.db.Get(key)
}

