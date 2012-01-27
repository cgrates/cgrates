package main

import (
    "github.com/simonz05/godis"    
)

type RedisStorage struct {
	db *godis.Client
}

func NewRedisStorage(address string) (*RedisStorage, error) {
	ndb:= godis.New(address, 10, "")
	return &RedisStorage{db: ndb}, nil
}


func (rs *RedisStorage) Close() {
	rs.db.Quit()
}

func (rs *RedisStorage) Get(key string) (string, error) {
	elem, err := rs.db.Get(key)
	return elem.String(), err
}

