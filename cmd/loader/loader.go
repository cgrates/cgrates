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
package main

import (
	"encoding/json"
	"flag"
	"github.com/rif/cgrates/timespans"
	"log"
	"os"
)

var (
	storage     = flag.String("storage", "all", "kyoto|redis|mongo")
	kyotofile   = flag.String("kyotofile", "storage.kch", "kyoto storage file (storage.kch)")
	redisserver = flag.String("redisserver", "tcp:127.0.0.1:6379", "redis server address (tcp:127.0.0.1:6379)")
	redisdb     = flag.Int("rdb", 10, "redis database number (10)")
	mongoserver = flag.String("mongoserver", "127.0.0.1:27017", "mongo server address (127.0.0.1:27017)")
	mongodb     = flag.String("mdb", "test", "mongo database name (test)")
	redispass   = flag.String("pass", "", "redis database password")
	apfile      = flag.String("apfile", "ap.json", "Activation Periods containing intervals file")
	destfile    = flag.String("destfile", "dest.json", "Destinations file")
	tpfile      = flag.String("tpfile", "tp.json", "Tariff plans file")
	ubfile      = flag.String("ubfile", "ub.json", "User budgets file")
)

func writeToStorage(storage timespans.StorageGetter,
	callDescriptors []*timespans.CallDescriptor,
	destinations []*timespans.Destination,
	tariffPlans []*timespans.TariffPlan,
	userBudgets []*timespans.UserBudget) {
	for _, cd := range callDescriptors {
		storage.SetActivationPeriods(cd.GetKey(), cd.ActivationPeriods)
		log.Printf("Storing %q", cd.GetKey())
	}
	for _, d := range destinations {
		storage.SetDestination(d)
		log.Printf("Storing %q", d.Id)
	}
	for _, tp := range tariffPlans {
		storage.SetTariffPlan(tp)
		log.Printf("Storing %q", tp.Id)
	}
	for _, ub := range userBudgets {
		storage.SetUserBudget(ub)
		log.Printf("Storing %q", ub.Id)
	}
}

func main() {
	flag.Parse()

	log.Printf("Reading from %s, %s, %s", *apfile, *destfile, *tpfile)

	// reading activation periods
	fin, err := os.Open(*apfile)

	if err != nil {
		log.Print("Cannot open activation periods input file", err)
	}

	dec := json.NewDecoder(fin)

	var callDescriptors []*timespans.CallDescriptor
	if err := dec.Decode(&callDescriptors); err != nil {
		log.Println(err)
		return
	}
	fin.Close()

	// reading destinations
	fin, err = os.Open(*destfile)

	if err != nil {
		log.Print("Cannot open destinations input file", err)
	}

	dec = json.NewDecoder(fin)

	var destinations []*timespans.Destination
	if err := dec.Decode(&destinations); err != nil {
		log.Println(err)
		return
	}
	fin.Close()

	// reading triff plans
	fin, err = os.Open(*tpfile)

	if err != nil {
		log.Print("Cannot open tariff plans input file", err)
	}

	dec = json.NewDecoder(fin)

	var tariffPlans []*timespans.TariffPlan
	if err := dec.Decode(&tariffPlans); err != nil {
		log.Println(err)
		return
	}
	fin.Close()

	// reading user budgets
	fin, err = os.Open(*ubfile)

	if err != nil {
		log.Print("Cannot open user budgets input file", err)
	}

	dec = json.NewDecoder(fin)

	var userBudgets []*timespans.UserBudget
	if err := dec.Decode(&userBudgets); err != nil {
		log.Println(err)
		return
	}
	fin.Close()

	switch *storage {
	case "kyoto":
		storage, _ := timespans.NewKyotoStorage(*kyotofile)
		defer storage.Close()
		writeToStorage(storage, callDescriptors, destinations, tariffPlans, userBudgets)
	case "mongo":
		storage, _ := timespans.NewMongoStorage(*mongoserver, *mongodb)
		defer storage.Close()
		writeToStorage(storage, callDescriptors, destinations, tariffPlans, userBudgets)
	case "redis":
		storage, _ := timespans.NewRedisStorage(*redisserver, *redisdb)
		defer storage.Close()
		writeToStorage(storage, callDescriptors, destinations, tariffPlans, userBudgets)
	default:
		kyoto, _ := timespans.NewKyotoStorage(*kyotofile)
		writeToStorage(kyoto, callDescriptors, destinations, tariffPlans, userBudgets)
		kyoto.Close()
		mongo, _ := timespans.NewMongoStorage(*mongoserver, *mongodb)
		writeToStorage(mongo, callDescriptors, destinations, tariffPlans, userBudgets)
		mongo.Close()
		redis, _ := timespans.NewRedisStorage(*redisserver, *redisdb)
		writeToStorage(redis, callDescriptors, destinations, tariffPlans, userBudgets)
		redis.Close()
	}
}
