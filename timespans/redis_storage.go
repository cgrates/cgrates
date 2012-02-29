/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012  Radu Ioan Fericean

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
package timespans

import (
	"bytes"
	"encoding/gob"
	"github.com/simonz05/godis"
	"log"
	"sync"
)

type RedisStorage struct {
	dbNb    int
	db      *godis.Client
	buf     bytes.Buffer
	decAP   *gob.Decoder
	decDest *gob.Decoder
	decTP   *gob.Decoder
	decUB   *gob.Decoder
	mux     sync.Mutex
}

func NewRedisStorage(address string, db int) (*RedisStorage, error) {
	ndb := godis.New(address, db, "")
	rs := &RedisStorage{db: ndb, dbNb: db}

	rs.decAP = gob.NewDecoder(&rs.buf)
	rs.decDest = gob.NewDecoder(&rs.buf)
	rs.decTP = gob.NewDecoder(&rs.buf)
	rs.decUB = gob.NewDecoder(&rs.buf)
	return rs, nil
}

func (rs *RedisStorage) Close() {
	rs.db.Quit()
}

func (rs *RedisStorage) SetActivationPeriods(key string, aps []*ActivationPeriod) error {
	//.db.Select(rs.dbNb)
	rs.mux.Lock()
	defer rs.mux.Unlock()

	var writeBuf bytes.Buffer
	encoder := gob.NewEncoder(&writeBuf)
	encoder.Encode(aps)
	return rs.db.Set(key, writeBuf.Bytes())
}

func (rs *RedisStorage) GetActivationPeriods(key string) (aps []*ActivationPeriod, err error) {
	//rs.db.Select(rs.dbNb)
	rs.mux.Lock()
	defer rs.mux.Unlock()

	elem, err := rs.db.Get(key)
	rs.buf.Reset()
	rs.buf.Write(elem.Bytes())

	rs.decAP.Decode(&aps)
	return
}

func (rs *RedisStorage) SetDestination(dest *Destination) error {
	//rs.db.Select(rs.dbNb + 1)	
	rs.mux.Lock()
	defer rs.mux.Unlock()

	var writeBuf bytes.Buffer
	encoder := gob.NewEncoder(&writeBuf)
	encoder.Encode(dest)
	return rs.db.Set(dest.Id, writeBuf.Bytes())
}

func (rs *RedisStorage) GetDestination(key string) (dest *Destination, err error) {
	//rs.db.Select(rs.dbNb + 1)
	rs.mux.Lock()
	defer rs.mux.Unlock()

	elem, err := rs.db.Get(key)
	rs.buf.Reset()
	rs.buf.Write(elem.Bytes())
	err = rs.decDest.Decode(&dest)
	log.Print(err)
	return
}

func (rs *RedisStorage) SetTariffPlan(tp *TariffPlan) error {
	//rs.db.Select(rs.dbNb + 2)
	rs.mux.Lock()
	defer rs.mux.Unlock()

	var writeBuf bytes.Buffer
	encoder := gob.NewEncoder(&writeBuf)
	encoder.Encode(tp)
	return rs.db.Set(tp.Id, writeBuf.Bytes())
}

func (rs *RedisStorage) GetTariffPlan(key string) (tp *TariffPlan, err error) {
	//rs.db.Select(rs.dbNb + 2)
	rs.mux.Lock()
	defer rs.mux.Unlock()

	elem, err := rs.db.Get(key)
	rs.buf.Reset()
	rs.buf.Write(elem.Bytes())
	rs.decTP.Decode(&tp)
	return
}

func (rs *RedisStorage) SetUserBudget(ub *UserBudget) error {
	//rs.db.Select(rs.dbNb + 3)
	rs.mux.Lock()
	defer rs.mux.Unlock()

	var writeBuf bytes.Buffer
	encoder := gob.NewEncoder(&writeBuf)
	encoder.Encode(ub)
	return rs.db.Set(ub.Id, writeBuf.Bytes())
}

func (rs *RedisStorage) GetUserBudget(key string) (ub *UserBudget, err error) {
	//rs.db.Select(rs.dbNb + 3)
	rs.mux.Lock()
	defer rs.mux.Unlock()

	elem, err := rs.db.Get(key)
	rs.buf.Reset()
	rs.buf.Write(elem.Bytes())
	rs.decUB.Decode(&ub)
	return
}
