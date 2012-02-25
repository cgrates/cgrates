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
	"github.com/fsouza/gokabinet/kc"
	//"log"
	"strings"
)

type KyotoStorage struct {
	db *kc.DB
}

func NewKyotoStorage(filaName string) (*KyotoStorage, error) {
	ndb, err := kc.Open(filaName, kc.WRITE)
	return &KyotoStorage{db: ndb}, err
}

func (ks *KyotoStorage) Close() {
	ks.db.Close()
}

func (ks *KyotoStorage) GetActivationPeriods(key string) (aps []*ActivationPeriod, err error) {
	values, err := ks.db.Get(key)

	if err == nil {
		for _, ap_string := range strings.Split(values, "\n") {
			if len(ap_string) > 0 {
				ap := &ActivationPeriod{}
				ap.restore(ap_string)
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
	return ks.db.Set(key, result)
}

func (ks *KyotoStorage) GetDestination(key string) (dest *Destination, err error) {
	if values, err := ks.db.Get(key); err == nil {
		dest = &Destination{Id: key}
		dest.restore(values)
	}
	return
}

func (ks *KyotoStorage) SetDestination(dest *Destination) error {
	return ks.db.Set(dest.Id, dest.store())
}

func (ks *KyotoStorage) GetTariffPlan(key string) (tp *TariffPlan, err error) {
	if values, err := ks.db.Get(key); err == nil {
		tp = &TariffPlan{Id: key}
		tp.restore(values)
	}
	return
}

func (ks *KyotoStorage) SetTariffPlan(tp *TariffPlan) error {
	return ks.db.Set(tp.Id, tp.store())
}

func (ks *KyotoStorage) GetUserBudget(key string) (ub *UserBudget, err error) {
	if values, err := ks.db.Get(key); err == nil {
		ub = &UserBudget{Id: key}
		ub.restore(values)
	}
	return
}

func (ks *KyotoStorage) SetUserBudget(ub *UserBudget) error {
	return ks.db.Set(ub.Id, ub.store())
}
