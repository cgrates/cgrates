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
	db  *cabinet.KCDB
	buf bytes.Buffer
	dec *gob.Decoder
	enc *gob.Encoder
	mux sync.Mutex // we need norma lock because we reset the buf variable
}

func NewKyotoStorage(filaName string) (*KyotoStorage, error) {
	ndb := cabinet.New()
	err := ndb.Open(filaName, cabinet.KCOWRITER|cabinet.KCOCREATE)
	ks := &KyotoStorage{db: ndb}

	ks.dec = gob.NewDecoder(&ks.buf)
	ks.enc = gob.NewEncoder(&ks.buf)
	ks.trainGobEncoderAndDecoder()
	return ks, err
}

func (ks *KyotoStorage) trainGobEncoderAndDecoder() {
	aps := []*ActivationPeriod{&ActivationPeriod{}}
	ks.enc.Encode(aps)
	ks.dec.Decode(&aps)
	ks.buf.Reset()
	dest := &Destination{}
	ks.enc.Encode(dest)
	ks.dec.Decode(&dest)
	ks.buf.Reset()
	tp := &TariffPlan{}
	ks.enc.Encode(tp)
	ks.dec.Decode(&tp)
	ks.buf.Reset()
	ub := &UserBudget{}
	ks.enc.Encode(ub)
	ks.dec.Decode(&ub)
	ks.buf.Reset()
}

func (ks *KyotoStorage) Close() {
	ks.db.Close()
}

func (ks *KyotoStorage) SetActivationPeriods(key string, aps []*ActivationPeriod) error {
	ks.mux.Lock()
	defer ks.mux.Unlock()

	ks.buf.Reset()
	ks.enc.Encode(aps)
	return ks.db.Set([]byte(key), ks.buf.Bytes())
}

func (ks *KyotoStorage) GetActivationPeriods(key string) (aps []*ActivationPeriod, err error) {
	ks.mux.Lock()
	defer ks.mux.Unlock()

	values, err := ks.db.Get([]byte(key))

	ks.buf.Reset()
	ks.buf.Write(values)
	ks.dec.Decode(&aps)
	return
}

func (ks *KyotoStorage) SetDestination(dest *Destination) error {
	ks.mux.Lock()
	defer ks.mux.Unlock()

	ks.buf.Reset()
	ks.enc.Encode(dest)
	return ks.db.Set([]byte(dest.Id), ks.buf.Bytes())
}

func (ks *KyotoStorage) GetDestination(key string) (dest *Destination, err error) {
	ks.mux.Lock()
	defer ks.mux.Unlock()

	values, err := ks.db.Get([]byte(key))

	ks.buf.Reset()
	ks.buf.Write(values)
	ks.dec.Decode(&dest)
	return
}

func (ks *KyotoStorage) SetTariffPlan(tp *TariffPlan) error {
	ks.mux.Lock()
	defer ks.mux.Unlock()

	ks.buf.Reset()
	ks.enc.Encode(tp)
	return ks.db.Set([]byte(tp.Id), ks.buf.Bytes())
}

func (ks *KyotoStorage) GetTariffPlan(key string) (tp *TariffPlan, err error) {
	ks.mux.Lock()
	defer ks.mux.Unlock()

	values, err := ks.db.Get([]byte(key))

	ks.buf.Reset()
	ks.buf.Write(values)
	ks.dec.Decode(&tp)
	return
}

func (ks *KyotoStorage) SetUserBudget(ub *UserBudget) error {
	ks.mux.Lock()
	defer ks.mux.Unlock()

	ks.buf.Reset()
	ks.enc.Encode(ub)
	return ks.db.Set([]byte(ub.Id), ks.buf.Bytes())
}

func (ks *KyotoStorage) GetUserBudget(key string) (ub *UserBudget, err error) {
	ks.mux.Lock()
	defer ks.mux.Unlock()

	values, err := ks.db.Get([]byte(key))

	ks.buf.Reset()
	ks.buf.Write(values)
	ks.dec.Decode(&ub)
	return
}
