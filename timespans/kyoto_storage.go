/*
Rating system designed to be used in VoIP Carrieks World
Copyright (C) 2012  Radu Ioan Fericean

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either veksion 3 of the License, or
(at your option) any later veksion.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/
package timespans

import (
	"bitbucket.org/ww/cabinet"
	"bytes"
	"encoding/gob"
	//"log"
	"sync"
)

type KyotoStorage struct {
	//db  *kc.DB
	db      *cabinet.KCDB
	buf     bytes.Buffer
	decAP   *gob.Decoder
	encAP   *gob.Encoder
	decDest *gob.Decoder
	encDest *gob.Encoder
	decTP   *gob.Decoder
	encTP   *gob.Encoder
	decUB   *gob.Decoder
	encUB   *gob.Encoder
	mux     sync.Mutex // we need norma lock because we reset the buf variable
}

func NewKyotoStorage(filaName string) (*KyotoStorage, error) {
	ndb := cabinet.New()
	err := ndb.Open(filaName, cabinet.KCOWRITER|cabinet.KCOCREATE)
	ks := &KyotoStorage{db: ndb}

	ks.decAP = gob.NewDecoder(&ks.buf)
	ks.encAP = gob.NewEncoder(&ks.buf)
	ks.decDest = gob.NewDecoder(&ks.buf)
	ks.encDest = gob.NewEncoder(&ks.buf)
	ks.decTP = gob.NewDecoder(&ks.buf)
	ks.encTP = gob.NewEncoder(&ks.buf)
	ks.decUB = gob.NewDecoder(&ks.buf)
	ks.encUB = gob.NewEncoder(&ks.buf)
	ks.trainGobEncodersAndDecoders()
	return ks, err
}

func (ks *KyotoStorage) trainGobEncodersAndDecoders() {
	aps := []*ActivationPeriod{&ActivationPeriod{}}
	ks.encAP.Encode(aps)
	ks.decAP.Decode(&aps)
	ks.buf.Reset()
	dest := &Destination{}
	ks.encDest.Encode(dest)
	ks.decDest.Decode(&dest)
	ks.buf.Reset()
	tp := &TariffPlan{}
	ks.encTP.Encode(tp)
	ks.decTP.Decode(&tp)
	ks.buf.Reset()
	ub := &UserBudget{}
	ks.encUB.Encode(ub)
	ks.decUB.Decode(&ub)
	ks.buf.Reset()
}

func (ks *KyotoStorage) Close() {
	ks.db.Close()
}

func (ks *KyotoStorage) SetActivationPeriods(key string, aps []*ActivationPeriod) error {
	ks.mux.Lock()
	defer ks.mux.Unlock()

	ks.buf.Reset()
	ks.encAP.Encode(aps)
	return ks.db.Set([]byte(key), ks.buf.Bytes())
}

func (ks *KyotoStorage) GetActivationPeriods(key string) (aps []*ActivationPeriod, err error) {
	ks.mux.Lock()
	defer ks.mux.Unlock()

	values, err := ks.db.Get([]byte(key))

	ks.buf.Reset()
	ks.buf.Write(values)
	ks.decAP.Decode(&aps)
	return
}

func (ks *KyotoStorage) SetDestination(dest *Destination) error {
	ks.mux.Lock()
	defer ks.mux.Unlock()

	ks.buf.Reset()
	ks.encDest.Encode(dest)
	return ks.db.Set([]byte(dest.Id), ks.buf.Bytes())
}

func (ks *KyotoStorage) GetDestination(key string) (dest *Destination, err error) {
	ks.mux.Lock()
	defer ks.mux.Unlock()

	values, err := ks.db.Get([]byte(key))

	ks.buf.Reset()
	ks.buf.Write(values)
	ks.decDest.Decode(&dest)
	return
}

func (ks *KyotoStorage) SetTariffPlan(tp *TariffPlan) error {
	ks.mux.Lock()
	defer ks.mux.Unlock()

	ks.buf.Reset()
	ks.encTP.Encode(tp)
	return ks.db.Set([]byte(tp.Id), ks.buf.Bytes())
}

func (ks *KyotoStorage) GetTariffPlan(key string) (tp *TariffPlan, err error) {
	ks.mux.Lock()
	defer ks.mux.Unlock()

	values, err := ks.db.Get([]byte(key))

	ks.buf.Reset()
	ks.buf.Write(values)
	ks.decTP.Decode(&tp)
	return
}

func (ks *KyotoStorage) SetUserBudget(ub *UserBudget) error {
	ks.mux.Lock()
	defer ks.mux.Unlock()

	ks.buf.Reset()
	ks.encUB.Encode(ub)
	return ks.db.Set([]byte(ub.Id), ks.buf.Bytes())
}

func (ks *KyotoStorage) GetUserBudget(key string) (ub *UserBudget, err error) {
	ks.mux.Lock()
	defer ks.mux.Unlock()

	values, err := ks.db.Get([]byte(key))

	ks.buf.Reset()
	ks.buf.Write(values)
	ks.decUB.Decode(&ub)
	return
}
