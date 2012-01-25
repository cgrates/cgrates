package main

import (
    "github.com/fsouza/gokabinet/kc"
)

type KyotoStorage struct {
	db *kc.DB
}

func NewKyotoStorage(filaName string) (KyotoStorage, error) {
	ndb, err := kc.Open(filaName, kc.READ)
	return KyotoStorage{db: ndb}, err	
}


func (ks KyotoStorage) Close() {
	ks.db.Close()
}

func (ks KyotoStorage) Get(key string) (value string, err error) {
	return ks.db.Get(key)
}

