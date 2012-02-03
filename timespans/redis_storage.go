package timespans

import (
	"github.com/simonz05/godis"
	"log"
)

type RedisStorage struct {
	db *godis.Client
}

func NewRedisStorage(address string) (*RedisStorage, error) {
	ndb := godis.New(address, 10, "")
	log.Print("Starting redis storage")
	return &RedisStorage{db: ndb}, nil
}

func (rs *RedisStorage) Close() {
	log.Print("Closing redis storage")
	rs.db.Quit()
}

func (rs *RedisStorage) Get(key string) (string, error) {
	elem, err := rs.db.Get(key)
	return elem.String(), err
}
