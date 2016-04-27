package main

import (
	"fmt"
	"log"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type MongoMigrator struct {
	session *mgo.Session
	db      *mgo.Database
}

func NewMongoMigrator(host, port, db, user, pass string) (*MongoMigrator, error) {
	address := fmt.Sprintf("%s:%s", host, port)
	if user != "" && pass != "" {
		address = fmt.Sprintf("%s:%s@%s", user, pass, address)
	}
	session, err := mgo.Dial(address)
	if err != nil {
		return nil, err
	}
	ndb := session.DB(db)
	return &MongoMigrator{session: session, db: ndb}, nil
}

func (mig MongoMigrator) migrateActions() error {
	newAcsMap := make(map[string]engine.Actions)
	iter := mig.db.C("actions").Find(nil).Iter()
	var oldAcs struct {
		Key   string
		Value Actions2
	}
	for iter.Next(&oldAcs) {
		log.Printf("Migrating action: %s...", oldAcs.Key)
		newAcs := make(engine.Actions, len(oldAcs.Value))
		for index, oldAc := range oldAcs.Value {
			a := &engine.Action{
				Id:               oldAc.Id,
				ActionType:       oldAc.ActionType,
				ExtraParameters:  oldAc.ExtraParameters,
				ExpirationString: oldAc.ExpirationString,
				Filter:           oldAc.Filter,
				Weight:           oldAc.Weight,
				Balance: &engine.BalanceFilter{
					Uuid:           oldAc.Balance.Uuid,
					ID:             oldAc.Balance.ID,
					Type:           oldAc.Balance.Type,
					Directions:     oldAc.Balance.Directions,
					ExpirationDate: oldAc.Balance.ExpirationDate,
					Weight:         oldAc.Balance.Weight,
					DestinationIDs: oldAc.Balance.DestinationIDs,
					RatingSubject:  oldAc.Balance.RatingSubject,
					Categories:     oldAc.Balance.Categories,
					SharedGroups:   oldAc.Balance.SharedGroups,
					TimingIDs:      oldAc.Balance.TimingIDs,
					Timings:        oldAc.Balance.Timings,
					Disabled:       oldAc.Balance.Disabled,
					Factor:         oldAc.Balance.Factor,
					Blocker:        oldAc.Balance.Blocker,
				},
			}
			if oldAc.Balance.Value != nil {
				a.Balance.Value = &utils.ValueFormula{Static: *oldAc.Balance.Value}
			}
			newAcs[index] = a
		}
		newAcsMap[oldAcs.Key] = newAcs
	}
	if err := iter.Close(); err != nil {
		return err
	}

	// write data back
	for key, acs := range newAcsMap {
		if _, err := mig.db.C("actions").Upsert(bson.M{"key": key}, &struct {
			Key   string
			Value engine.Actions
		}{Key: key, Value: acs}); err != nil {
			return err
		}
	}
	return nil
}

func (mig MongoMigrator) writeVersion() error {
	_, err := mig.db.C("versions").Upsert(bson.M{"key": utils.VERSION_PREFIX + "struct"}, &struct {
		Key   string
		Value *engine.StructVersion
	}{utils.VERSION_PREFIX + "struct", engine.CurrentVersion})
	return err
}
