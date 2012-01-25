package main

import (
    "github.com/fsouza/gokabinet/kc"
)

type KyotoStorage struct {
	db *kc.DB
}

func (ks KyotoStorage) Open(fileName string) error {
	db, err := kc.Open(fileName, kc.READ)
	ks.db = db
	return err
}

func (ks KyotoStorage) Close() {
	ks.db.Close()
}

func (ks KyotoStorage) Get(key string) (value string, err error) {
	return ks.db.Get(key)
}

