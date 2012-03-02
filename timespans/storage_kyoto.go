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
	//"github.com/fsouza/gokabinet/kc"
	"bitbucket.org/ww/cabinet"
	//"log"
	"bytes"
)

type KyotoStorage struct {
	db *cabinet.KCDB
}

func NewKyotoStorage(filaName string) (*KyotoStorage, error) {
	// ndb, err := kc.Open(filaName, kc.WRITE)
	ndb := cabinet.New()
	err := ndb.Open(filaName, cabinet.KCOWRITER|cabinet.KCOCREATE)
	return &KyotoStorage{db: ndb}, err
}

func (ks *KyotoStorage) Close() {
	ks.db.Close()
}

func (ks *KyotoStorage) GetActivationPeriods(key string) (aps []*ActivationPeriod, err error) {
	values, err := ks.db.Get([]byte(key))

	if err == nil {
		for _, ap_string := range bytes.Split(values, []byte{'\n'}) {
			if len(ap_string) > 0 {
				ap := &ActivationPeriod{}
				ap.restore(string(ap_string))
				aps = append(aps, ap)
			}
		}
	}
	return aps, err
}

func (ks *KyotoStorage) SetActivationPeriods(key string, aps []*ActivationPeriod) error {
	result := ""
	for _, ap := range aps {
		result += ap.store() + "\n"
	}
	return ks.db.Set([]byte(key), []byte(result))
}

func (ks *KyotoStorage) GetDestination(key string) (dest *Destination, err error) {
	if values, err := ks.db.Get([]byte(key)); err == nil {
		dest = &Destination{Id: key}
		dest.restore(string(values))
	}
	return
}

func (ks *KyotoStorage) SetDestination(dest *Destination) error {
	return ks.db.Set([]byte(dest.Id), []byte(dest.store()))
}

func (ks *KyotoStorage) GetTariffPlan(key string) (tp *TariffPlan, err error) {
	if values, err := ks.db.Get([]byte(key)); err == nil {
		tp = &TariffPlan{Id: key}
		tp.restore(string(values))
	}
	return
}

func (ks *KyotoStorage) SetTariffPlan(tp *TariffPlan) error {
	return ks.db.Set([]byte(tp.Id), []byte(tp.store()))
}

func (ks *KyotoStorage) GetUserBudget(key string) (ub *UserBudget, err error) {
	if values, err := ks.db.Get([]byte(key)); err == nil {
		ub = &UserBudget{Id: key}
		ub.restore(string(values))
	}
	return
}

func (ks *KyotoStorage) SetUserBudget(ub *UserBudget) error {
	return ks.db.Set([]byte(ub.Id), []byte(ub.store()))
}
