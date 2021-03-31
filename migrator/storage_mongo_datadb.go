/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

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

package migrator

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	v2AccountsCol          = "accounts"
	v1ActionTriggersCol    = "action_triggers"
	v1AttributeProfilesCol = "attribute_profiles"
	v2ThresholdProfileCol  = "threshold_profiles"
	v1AliasCol             = "aliases"
	v1UserCol              = "users"
	v1DerivedChargersCol   = "derived_chargers"
	v2StatsCol             = "statqueues"
)

type mongoMigrator struct {
	dm     *engine.DataManager
	mgoDB  *engine.MongoStorage
	cursor *mongo.Cursor
}

type AcKeyValue struct {
	Key   string
	Value v1Actions
}
type AtKeyValue struct {
	Key   string
	Value v1ActionPlans
}

func newMongoMigrator(dm *engine.DataManager) (mgoMig *mongoMigrator) {
	return &mongoMigrator{
		dm:     dm,
		mgoDB:  dm.DataDB().(*engine.MongoStorage),
		cursor: nil,
	}
}

func (v1ms *mongoMigrator) close() {
	v1ms.mgoDB.Close()
}

func (v1ms *mongoMigrator) DataManager() *engine.DataManager {
	return v1ms.dm
}

//Account methods
//V1
//get
func (v1ms *mongoMigrator) getv1Account() (v1Acnt *v1Account, err error) {
	if v1ms.cursor == nil {
		v1ms.cursor, err = v1ms.mgoDB.DB().Collection(v1AccountDBPrefix).Find(v1ms.mgoDB.GetContext(), bson.D{})
		if err != nil {
			return nil, err
		}
	}
	if !(*v1ms.cursor).Next(v1ms.mgoDB.GetContext()) {
		(*v1ms.cursor).Close(v1ms.mgoDB.GetContext())
		v1ms.cursor = nil
		return nil, utils.ErrNoMoreData
	}
	v1Acnt = new(v1Account)
	if err := (*v1ms.cursor).Decode(v1Acnt); err != nil {
		return nil, err
	}
	return v1Acnt, nil
}

//set
func (v1ms *mongoMigrator) setV1Account(x *v1Account) (err error) {
	if x != nil {
		_, err = v1ms.mgoDB.DB().Collection(v1AccountDBPrefix).InsertOne(v1ms.mgoDB.GetContext(), x)
	}
	return
}

//rem
func (v1ms *mongoMigrator) remV1Account(id string) (err error) {
	_, err = v1ms.mgoDB.DB().Collection(v1AccountDBPrefix).DeleteOne(v1ms.mgoDB.GetContext(), bson.M{"id": id})
	return
}

//V2
//get
func (v1ms *mongoMigrator) getv2Account() (v2Acnt *v2Account, err error) {
	if v1ms.cursor == nil {
		v1ms.cursor, err = v1ms.mgoDB.DB().Collection(v2AccountsCol).Find(v1ms.mgoDB.GetContext(), bson.D{})
		if err != nil {
			return nil, err
		}
	}
	if !(*v1ms.cursor).Next(v1ms.mgoDB.GetContext()) {
		(*v1ms.cursor).Close(v1ms.mgoDB.GetContext())
		v1ms.cursor = nil
		return nil, utils.ErrNoMoreData
	}
	v2Acnt = new(v2Account)
	if err := (*v1ms.cursor).Decode(v2Acnt); err != nil {
		return nil, err
	}
	return v2Acnt, nil
}

//set
func (v1ms *mongoMigrator) setV2Account(x *v2Account) (err error) {
	if x != nil {
		_, err = v1ms.mgoDB.DB().Collection(v2AccountsCol).InsertOne(v1ms.mgoDB.GetContext(), x)
	}
	return
}

//rem
func (v1ms *mongoMigrator) remV2Account(id string) (err error) {
	_, err = v1ms.mgoDB.DB().Collection(v2AccountsCol).DeleteOne(v1ms.mgoDB.GetContext(), bson.M{"id": id})
	return
}

//Action methods
//get
func (v1ms *mongoMigrator) getV1ActionPlans() (v1aps *v1ActionPlans, err error) {
	strct := new(AtKeyValue)
	if v1ms.cursor == nil {
		v1ms.cursor, err = v1ms.mgoDB.DB().Collection("actiontimings").Find(v1ms.mgoDB.GetContext(), bson.D{})
		if err != nil {
			return nil, err
		}
	}
	if !(*v1ms.cursor).Next(v1ms.mgoDB.GetContext()) {
		(*v1ms.cursor).Close(v1ms.mgoDB.GetContext())
		v1ms.cursor = nil
		return nil, utils.ErrNoMoreData
	}
	if err := (*v1ms.cursor).Decode(strct); err != nil {
		return nil, err
	}
	return &strct.Value, nil
}

//set
func (v1ms *mongoMigrator) setV1ActionPlans(x *v1ActionPlans) (err error) {
	key := utils.ActionPlanPrefix + (*x)[0].Id
	_, err = v1ms.mgoDB.DB().Collection("actiontimings").InsertOne(v1ms.mgoDB.GetContext(), &AtKeyValue{key, *x})
	return
}

//rem
func (v1ms *mongoMigrator) remV1ActionPlans(x *v1ActionPlans) (err error) {
	for _, item := range *x {
		_, err = v1ms.mgoDB.DB().Collection("actiontimings").DeleteOne(v1ms.mgoDB.GetContext(), bson.M{"id": item.Id})
		return
	}
	return
}

//Actions methods
//get
func (v1ms *mongoMigrator) getV1Actions() (v1acs *v1Actions, err error) {
	strct := new(AcKeyValue)
	if v1ms.cursor == nil {
		v1ms.cursor, err = v1ms.mgoDB.DB().Collection("actions").Find(v1ms.mgoDB.GetContext(), bson.D{})
		if err != nil {
			return nil, err
		}
	}
	if !(*v1ms.cursor).Next(v1ms.mgoDB.GetContext()) {
		(*v1ms.cursor).Close(v1ms.mgoDB.GetContext())
		v1ms.cursor = nil
		return nil, utils.ErrNoMoreData
	}
	if err := (*v1ms.cursor).Decode(strct); err != nil {
		return nil, err
	}
	return &strct.Value, nil
}

//set
func (v1ms *mongoMigrator) setV1Actions(x *v1Actions) (err error) {
	key := utils.ActionPrefix + (*x)[0].Id
	_, err = v1ms.mgoDB.DB().Collection("actions").InsertOne(v1ms.mgoDB.GetContext(), &AcKeyValue{key, *x})
	return
}

//rem
func (v1ms *mongoMigrator) remV1Actions(x v1Actions) (err error) {
	_, err = v1ms.mgoDB.DB().Collection("actions").DeleteOne(v1ms.mgoDB.GetContext(), bson.M{"id": x[0].Id})
	return
}

//ActionTriggers methods
//get
func (v1ms *mongoMigrator) getV1ActionTriggers() (v1acts *v1ActionTriggers, err error) {
	if v1ms.cursor == nil {
		v1ms.cursor, err = v1ms.mgoDB.DB().Collection(v1ActionTriggersCol).Find(v1ms.mgoDB.GetContext(), bson.D{})
		if err != nil {
			return nil, err
		}
	}
	if !(*v1ms.cursor).Next(v1ms.mgoDB.GetContext()) {
		(*v1ms.cursor).Close(v1ms.mgoDB.GetContext())
		v1ms.cursor = nil
		return nil, utils.ErrNoMoreData
	}
	v1act := new(v1ActionTrigger)
	if err := (*v1ms.cursor).Decode(v1act); err != nil {
		return nil, err
	}
	return &v1ActionTriggers{v1act}, nil
}

//set
func (v1ms *mongoMigrator) setV1ActionTriggers(act *v1ActionTriggers) (err error) {
	for _, x := range *act {
		_, err = v1ms.mgoDB.DB().Collection(v1ActionTriggersCol).InsertOne(v1ms.mgoDB.GetContext(), *x)
		if err != nil {
			return err
		}
	}
	return
}

//rem
func (v1ms *mongoMigrator) remV1ActionTriggers(x *v1ActionTriggers) (err error) {
	for _, item := range *x {
		_, err = v1ms.mgoDB.DB().Collection(v1ActionTriggersCol).DeleteOne(v1ms.mgoDB.GetContext(), bson.M{"id": item.Id})
		return
	}
	return
}

//Actions methods
//get
func (v1ms *mongoMigrator) getV1SharedGroup() (v1sg *v1SharedGroup, err error) {
	if v1ms.cursor == nil {
		v1ms.cursor, err = v1ms.mgoDB.DB().Collection(utils.SharedGroupPrefix).Find(v1ms.mgoDB.GetContext(), bson.D{})
		if err != nil {
			return nil, err
		}
	}
	if !(*v1ms.cursor).Next(v1ms.mgoDB.GetContext()) {
		(*v1ms.cursor).Close(v1ms.mgoDB.GetContext())
		v1ms.cursor = nil
		return nil, utils.ErrNoMoreData
	}
	v1sg = new(v1SharedGroup)
	if err := (*v1ms.cursor).Decode(v1sg); err != nil {
		return nil, err
	}
	return v1sg, nil
}

//set
func (v1ms *mongoMigrator) setV1SharedGroup(x *v1SharedGroup) (err error) {
	_, err = v1ms.mgoDB.DB().Collection(utils.SharedGroupPrefix).InsertOne(v1ms.mgoDB.GetContext(), x)
	return
}

//Stats methods
//get
func (v1ms *mongoMigrator) getV1Stats() (v1st *v1Stat, err error) {
	if v1ms.cursor == nil {
		v1ms.cursor, err = v1ms.mgoDB.DB().Collection(utils.CDRsStatsPrefix).Find(v1ms.mgoDB.GetContext(), bson.D{})
		if err != nil {
			return nil, err
		}
	}
	if !(*v1ms.cursor).Next(v1ms.mgoDB.GetContext()) {
		(*v1ms.cursor).Close(v1ms.mgoDB.GetContext())
		v1ms.cursor = nil
		return nil, utils.ErrNoMoreData
	}
	v1st = new(v1Stat)
	if err := (*v1ms.cursor).Decode(v1st); err != nil {
		return nil, err
	}
	return v1st, nil
}

func (v1ms *mongoMigrator) getV3Stats() (v1st *engine.StatQueueProfile, err error) {
	if v1ms.cursor == nil {
		v1ms.cursor, err = v1ms.mgoDB.DB().Collection(engine.ColSqp).Find(v1ms.mgoDB.GetContext(), bson.D{})
		if err != nil {
			return nil, err
		}
	}
	if !(*v1ms.cursor).Next(v1ms.mgoDB.GetContext()) {
		(*v1ms.cursor).Close(v1ms.mgoDB.GetContext())
		v1ms.cursor = nil
		return nil, utils.ErrNoMoreData
	}
	v1st = new(engine.StatQueueProfile)
	if err := (*v1ms.cursor).Decode(v1st); err != nil {
		return nil, err
	}
	return v1st, nil
}

//set
func (v1ms *mongoMigrator) setV1Stats(x *v1Stat) (err error) {
	_, err = v1ms.mgoDB.DB().Collection(utils.CDRsStatsPrefix).InsertOne(v1ms.mgoDB.GetContext(), x)
	return
}

// get V2
func (v1ms *mongoMigrator) getV2Stats() (v2 *engine.StatQueue, err error) {
	if v1ms.cursor == nil {
		v1ms.cursor, err = v1ms.mgoDB.DB().Collection(v2StatsCol).Find(v1ms.mgoDB.GetContext(), bson.D{})
		if err != nil {
			return nil, err
		}
	}
	if !(*v1ms.cursor).Next(v1ms.mgoDB.GetContext()) {
		(*v1ms.cursor).Close(v1ms.mgoDB.GetContext())
		v1ms.cursor = nil
		return nil, utils.ErrNoMoreData
	}
	v2 = new(engine.StatQueue)
	if err := (*v1ms.cursor).Decode(v2); err != nil {
		return nil, err
	}
	return v2, nil
}

// set v2
func (v1ms *mongoMigrator) setV2Stats(v2 *engine.StatQueue) (err error) {
	_, err = v1ms.mgoDB.DB().Collection(v2StatsCol).InsertOne(v1ms.mgoDB.GetContext(), v2)
	return
}

//Stats methods
//get
func (v1ms *mongoMigrator) getV2ActionTrigger() (v2at *v2ActionTrigger, err error) {
	if v1ms.cursor == nil {
		v1ms.cursor, err = v1ms.mgoDB.DB().Collection(v1ActionTriggersCol).Find(v1ms.mgoDB.GetContext(), bson.D{})
		if err != nil {
			return nil, err
		}
	}
	if !(*v1ms.cursor).Next(v1ms.mgoDB.GetContext()) {
		(*v1ms.cursor).Close(v1ms.mgoDB.GetContext())
		v1ms.cursor = nil
		return nil, utils.ErrNoMoreData
	}
	v2at = new(v2ActionTrigger)
	if err := (*v1ms.cursor).Decode(v2at); err != nil {
		return nil, err
	}
	return v2at, nil
}

//set
func (v1ms *mongoMigrator) setV2ActionTrigger(x *v2ActionTrigger) (err error) {
	_, err = v1ms.mgoDB.DB().Collection(v1ActionTriggersCol).InsertOne(v1ms.mgoDB.GetContext(), x)
	return
}

//AttributeProfile methods
//get
func (v1ms *mongoMigrator) getV1AttributeProfile() (v1attrPrf *v1AttributeProfile, err error) {
	if v1ms.cursor == nil {
		v1ms.cursor, err = v1ms.mgoDB.DB().Collection(v1AttributeProfilesCol).Find(v1ms.mgoDB.GetContext(), bson.D{})
		if err != nil {
			return nil, err
		}
	}
	if !(*v1ms.cursor).Next(v1ms.mgoDB.GetContext()) {
		(*v1ms.cursor).Close(v1ms.mgoDB.GetContext())
		v1ms.cursor = nil
		return nil, utils.ErrNoMoreData
	}
	v1attrPrf = new(v1AttributeProfile)
	if err := (*v1ms.cursor).Decode(v1attrPrf); err != nil {
		return nil, err
	}
	return v1attrPrf, nil
}

//set
func (v1ms *mongoMigrator) setV1AttributeProfile(x *v1AttributeProfile) (err error) {
	_, err = v1ms.mgoDB.DB().Collection(v1AttributeProfilesCol).InsertOne(v1ms.mgoDB.GetContext(), x)
	return
}

//ThresholdProfile methods
//get
func (v1ms *mongoMigrator) getV2ThresholdProfile() (v2T *v2Threshold, err error) {
	if v1ms.cursor == nil {
		v1ms.cursor, err = v1ms.mgoDB.DB().Collection(v2ThresholdProfileCol).Find(v1ms.mgoDB.GetContext(), bson.D{})
		if err != nil {
			return nil, err
		}
	}
	if !(*v1ms.cursor).Next(v1ms.mgoDB.GetContext()) {
		(*v1ms.cursor).Close(v1ms.mgoDB.GetContext())
		v1ms.cursor = nil
		return nil, utils.ErrNoMoreData
	}
	v2T = new(v2Threshold)
	if err := (*v1ms.cursor).Decode(v2T); err != nil {
		return nil, err
	}
	return v2T, nil
}

func (v1ms *mongoMigrator) getV3ThresholdProfile() (v2T *engine.ThresholdProfile, err error) {
	if v1ms.cursor == nil {
		v1ms.cursor, err = v1ms.mgoDB.DB().Collection(engine.ColTps).Find(v1ms.mgoDB.GetContext(), bson.D{})
		if err != nil {
			return nil, err
		}
	}
	if !(*v1ms.cursor).Next(v1ms.mgoDB.GetContext()) {
		(*v1ms.cursor).Close(v1ms.mgoDB.GetContext())
		v1ms.cursor = nil
		return nil, utils.ErrNoMoreData
	}
	v2T = new(engine.ThresholdProfile)
	if err := (*v1ms.cursor).Decode(v2T); err != nil {
		return nil, err
	}
	return v2T, nil
}

//set
func (v1ms *mongoMigrator) setV2ThresholdProfile(x *v2Threshold) (err error) {
	_, err = v1ms.mgoDB.DB().Collection(v2ThresholdProfileCol).InsertOne(v1ms.mgoDB.GetContext(), x)
	return
}

//rem
func (v1ms *mongoMigrator) remV2ThresholdProfile(tenant, id string) (err error) {
	_, err = v1ms.mgoDB.DB().Collection(v2ThresholdProfileCol).DeleteOne(v1ms.mgoDB.GetContext(), bson.M{"tenant": tenant, "id": id})
	return
}

//Alias methods
//get
func (v1ms *mongoMigrator) getV1Alias() (v1a *v1Alias, err error) {
	if v1ms.cursor == nil {
		v1ms.cursor, err = v1ms.mgoDB.DB().Collection(v1AliasCol).Find(v1ms.mgoDB.GetContext(), bson.D{})
		if err != nil {
			return nil, err
		}
	}
	if !(*v1ms.cursor).Next(v1ms.mgoDB.GetContext()) {
		(*v1ms.cursor).Close(v1ms.mgoDB.GetContext())
		v1ms.cursor = nil
		return nil, utils.ErrNoMoreData
	}
	v1a = new(v1Alias)
	var kv struct {
		Key   string
		Value v1AliasValues
	}
	if err := (*v1ms.cursor).Decode(&kv); err != nil {
		return nil, err
	}
	v1a.Values = kv.Value
	v1a.SetId(kv.Key)
	return v1a, nil
}

//set
func (v1ms *mongoMigrator) setV1Alias(al *v1Alias) (err error) {
	_, err = v1ms.mgoDB.DB().Collection(v1AliasCol).UpdateOne(v1ms.mgoDB.GetContext(), bson.M{"key": al.GetId()},
		bson.M{"$set": struct {
			Key   string
			Value v1AliasValues
		}{Key: al.GetId(), Value: al.Values}},
		options.Update().SetUpsert(true),
	)
	return err
}

//rem
func (v1ms *mongoMigrator) remV1Alias(key string) (err error) {
	al := new(v1Alias)
	al.SetId(key)
	var kv struct {
		Key   string
		Value v1AliasValues
	}
	cur := v1ms.mgoDB.DB().Collection(v1AliasCol).FindOne(v1ms.mgoDB.GetContext(), bson.M{"key": key})
	if err := cur.Decode(&kv); err != nil {
		if err == mongo.ErrNoDocuments {
			return utils.ErrNotFound
		}
		return err
	}
	al.Values = kv.Value
	dr, err := v1ms.mgoDB.DB().Collection(v1AliasCol).DeleteOne(v1ms.mgoDB.GetContext(), bson.M{"key": key})
	if dr.DeletedCount == 0 {
		return utils.ErrNotFound
	}
	if err != nil {
		return err
	}
	for _, value := range al.Values {
		tmpKey := utils.ConcatenatedKey(al.GetId(), value.DestinationId)
		for target, pairs := range value.Pairs {
			for _, alias := range pairs {
				rKey := alias + target + al.Context
				_, err = v1ms.mgoDB.DB().Collection(v1AliasCol).UpdateOne(v1ms.mgoDB.GetContext(), bson.M{"key": rKey},
					bson.M{"$pull": bson.M{"value": tmpKey}})
				if err != nil {
					return err
				}
			}
		}
	}
	return
}

// User methods
//get
func (v1ms *mongoMigrator) getV1User() (v1u *v1UserProfile, err error) {
	if v1ms.cursor == nil {
		v1ms.cursor, err = v1ms.mgoDB.DB().Collection(v1UserCol).Find(v1ms.mgoDB.GetContext(), bson.D{})
		if err != nil {
			return nil, err
		}
	}
	if !(*v1ms.cursor).Next(v1ms.mgoDB.GetContext()) {
		(*v1ms.cursor).Close(v1ms.mgoDB.GetContext())
		v1ms.cursor = nil
		return nil, utils.ErrNoMoreData
	}
	var kv struct {
		Key   string
		Value *v1UserProfile
	}
	if err := (*v1ms.cursor).Decode(&kv); err != nil {
		return nil, err
	}
	return kv.Value, nil
}

//set
func (v1ms *mongoMigrator) setV1User(us *v1UserProfile) (err error) {
	_, err = v1ms.mgoDB.DB().Collection(v1UserCol).UpdateOne(v1ms.mgoDB.GetContext(), bson.M{"key": us.GetId()},
		bson.M{"$set": struct {
			Key   string
			Value *v1UserProfile
		}{Key: us.GetId(), Value: us}},
		options.Update().SetUpsert(true),
	)
	return err
}

//rem
func (v1ms *mongoMigrator) remV1User(key string) (err error) {
	_, err = v1ms.mgoDB.DB().Collection(v1UserCol).DeleteOne(v1ms.mgoDB.GetContext(), bson.M{"key": key})
	return
}

// DerivedChargers methods
//get
func (v1ms *mongoMigrator) getV1DerivedChargers() (v1d *v1DerivedChargersWithKey, err error) {
	if v1ms.cursor == nil {
		v1ms.cursor, err = v1ms.mgoDB.DB().Collection(v1DerivedChargersCol).Find(v1ms.mgoDB.GetContext(), bson.D{})
		if err != nil {
			return nil, err
		}
	}
	if !(*v1ms.cursor).Next(v1ms.mgoDB.GetContext()) {
		(*v1ms.cursor).Close(v1ms.mgoDB.GetContext())
		v1ms.cursor = nil
		return nil, utils.ErrNoMoreData
	}
	v1d = new(v1DerivedChargersWithKey)
	if err := (*v1ms.cursor).Decode(v1d); err != nil {
		return nil, err
	}
	return v1d, nil
}

//set
func (v1ms *mongoMigrator) setV1DerivedChargers(dc *v1DerivedChargersWithKey) (err error) {
	_, err = v1ms.mgoDB.DB().Collection(v1DerivedChargersCol).UpdateOne(v1ms.mgoDB.GetContext(), bson.M{"key": dc.Key},
		bson.M{"$set": dc},
		options.Update().SetUpsert(true),
	)
	return
}

//rem
func (v1ms *mongoMigrator) remV1DerivedChargers(key string) (err error) {
	_, err = v1ms.mgoDB.DB().Collection(v1DerivedChargersCol).DeleteOne(v1ms.mgoDB.GetContext(), bson.M{"key": key})
	return
}

//AttributeProfile methods
//get
func (v1ms *mongoMigrator) getV2AttributeProfile() (v2attrPrf *v2AttributeProfile, err error) {
	if v1ms.cursor == nil {
		v1ms.cursor, err = v1ms.mgoDB.DB().Collection(v1AttributeProfilesCol).Find(v1ms.mgoDB.GetContext(), bson.D{})
		if err != nil {
			return nil, err
		}
	}
	if !(*v1ms.cursor).Next(v1ms.mgoDB.GetContext()) {
		(*v1ms.cursor).Close(v1ms.mgoDB.GetContext())
		v1ms.cursor = nil
		return nil, utils.ErrNoMoreData
	}
	v2attrPrf = new(v2AttributeProfile)
	if err := (*v1ms.cursor).Decode(v2attrPrf); err != nil {
		return nil, err
	}
	return v2attrPrf, nil
}

//set
func (v1ms *mongoMigrator) setV2AttributeProfile(x *v2AttributeProfile) (err error) {
	_, err = v1ms.mgoDB.DB().Collection(v1AttributeProfilesCol).InsertOne(v1ms.mgoDB.GetContext(), x)
	return
}

//rem
func (v1ms *mongoMigrator) remV2AttributeProfile(tenant, id string) (err error) {
	_, err = v1ms.mgoDB.DB().Collection(v1AttributeProfilesCol).DeleteOne(v1ms.mgoDB.GetContext(), bson.M{"tenant": tenant, "id": id})
	return
}

//AttributeProfile methods
//get
func (v1ms *mongoMigrator) getV3AttributeProfile() (v3attrPrf *v3AttributeProfile, err error) {
	if v1ms.cursor == nil {
		v1ms.cursor, err = v1ms.mgoDB.DB().Collection(v1AttributeProfilesCol).Find(v1ms.mgoDB.GetContext(), bson.D{})
		if err != nil {
			return nil, err
		}
	}
	if !(*v1ms.cursor).Next(v1ms.mgoDB.GetContext()) {
		(*v1ms.cursor).Close(v1ms.mgoDB.GetContext())
		v1ms.cursor = nil
		return nil, utils.ErrNoMoreData
	}
	v3attrPrf = new(v3AttributeProfile)
	if err := (*v1ms.cursor).Decode(v3attrPrf); err != nil {
		return nil, err
	}
	return v3attrPrf, nil
}

//set
func (v1ms *mongoMigrator) setV3AttributeProfile(x *v3AttributeProfile) (err error) {
	_, err = v1ms.mgoDB.DB().Collection(v1AttributeProfilesCol).InsertOne(v1ms.mgoDB.GetContext(), x)
	return
}

//rem
func (v1ms *mongoMigrator) remV3AttributeProfile(tenant, id string) (err error) {
	_, err = v1ms.mgoDB.DB().Collection(v1AttributeProfilesCol).DeleteOne(v1ms.mgoDB.GetContext(), bson.M{"tenant": tenant, "id": id})
	return
}

//AttributeProfile methods
//get
func (v1ms *mongoMigrator) getV4AttributeProfile() (v4attrPrf *v4AttributeProfile, err error) {
	if v1ms.cursor == nil {
		v1ms.cursor, err = v1ms.mgoDB.DB().Collection(v1AttributeProfilesCol).Find(v1ms.mgoDB.GetContext(), bson.D{})
		if err != nil {
			return nil, err
		}
	}
	if !(*v1ms.cursor).Next(v1ms.mgoDB.GetContext()) {
		(*v1ms.cursor).Close(v1ms.mgoDB.GetContext())
		v1ms.cursor = nil
		return nil, utils.ErrNoMoreData
	}
	v4attrPrf = new(v4AttributeProfile)
	if err := (*v1ms.cursor).Decode(v4attrPrf); err != nil {
		return nil, err
	}
	return v4attrPrf, nil
}

func (v1ms *mongoMigrator) getV5AttributeProfile() (v5attrPrf *engine.AttributeProfile, err error) {
	if v1ms.cursor == nil {
		v1ms.cursor, err = v1ms.mgoDB.DB().Collection(v1AttributeProfilesCol).Find(v1ms.mgoDB.GetContext(), bson.D{})
		if err != nil {
			return nil, err
		}
	}
	if !(*v1ms.cursor).Next(v1ms.mgoDB.GetContext()) {
		(*v1ms.cursor).Close(v1ms.mgoDB.GetContext())
		v1ms.cursor = nil
		return nil, utils.ErrNoMoreData
	}
	v5attrPrf = new(engine.AttributeProfile)
	if err := (*v1ms.cursor).Decode(v5attrPrf); err != nil {
		return nil, err
	}
	return v5attrPrf, nil
}

//set
func (v1ms *mongoMigrator) setV4AttributeProfile(x *v4AttributeProfile) (err error) {
	_, err = v1ms.mgoDB.DB().Collection(v1AttributeProfilesCol).InsertOne(v1ms.mgoDB.GetContext(), x)
	return
}

//rem
func (v1ms *mongoMigrator) remV4AttributeProfile(tenant, id string) (err error) {
	_, err = v1ms.mgoDB.DB().Collection(v1AttributeProfilesCol).DeleteOne(v1ms.mgoDB.GetContext(), bson.M{"tenant": tenant, "id": id})
	return
}

// Filter Methods
//get
func (v1ms *mongoMigrator) getV1Filter() (v1Fltr *v1Filter, err error) {
	if v1ms.cursor == nil {
		v1ms.cursor, err = v1ms.mgoDB.DB().Collection(engine.ColFlt).Find(v1ms.mgoDB.GetContext(), bson.D{})
		if err != nil {
			return nil, err
		}
	}
	if !(*v1ms.cursor).Next(v1ms.mgoDB.GetContext()) {
		(*v1ms.cursor).Close(v1ms.mgoDB.GetContext())
		v1ms.cursor = nil
		return nil, utils.ErrNoMoreData
	}
	v1Fltr = new(v1Filter)
	if err := (*v1ms.cursor).Decode(v1Fltr); err != nil {
		return nil, err
	}
	return
}

func (v1ms *mongoMigrator) getV4Filter() (v4Fltr *engine.Filter, err error) {
	if v1ms.cursor == nil {
		v1ms.cursor, err = v1ms.mgoDB.DB().Collection(engine.ColFlt).Find(v1ms.mgoDB.GetContext(), bson.D{})
		if err != nil {
			return nil, err
		}
	}
	if !(*v1ms.cursor).Next(v1ms.mgoDB.GetContext()) {
		(*v1ms.cursor).Close(v1ms.mgoDB.GetContext())
		v1ms.cursor = nil
		return nil, utils.ErrNoMoreData
	}
	v4Fltr = new(engine.Filter)
	if err := (*v1ms.cursor).Decode(v4Fltr); err != nil {
		return nil, err
	}
	return
}

//set
func (v1ms *mongoMigrator) setV1Filter(x *v1Filter) (err error) {
	_, err = v1ms.mgoDB.DB().Collection(engine.ColFlt).InsertOne(v1ms.mgoDB.GetContext(), x)
	return
}

//rem
func (v1ms *mongoMigrator) remV1Filter(tenant, id string) (err error) {
	_, err = v1ms.mgoDB.DB().Collection(engine.ColFlt).DeleteOne(v1ms.mgoDB.GetContext(), bson.M{"tenant": tenant, "id": id})
	return
}

// Supplier Methods
//get
func (v1ms *mongoMigrator) getSupplier() (spl *SupplierProfile, err error) {
	if v1ms.cursor == nil {
		v1ms.cursor, err = v1ms.mgoDB.DB().Collection(ColSpp).Find(v1ms.mgoDB.GetContext(), bson.D{})
		if err != nil {
			return nil, err
		}
	}
	if !(*v1ms.cursor).Next(v1ms.mgoDB.GetContext()) {
		(*v1ms.cursor).Close(v1ms.mgoDB.GetContext())
		v1ms.cursor = nil
		return nil, utils.ErrNoMoreData
	}
	spl = new(SupplierProfile)
	if err := (*v1ms.cursor).Decode(spl); err != nil {
		return nil, err
	}
	return
}

//set
func (v1ms *mongoMigrator) setSupplier(spl *SupplierProfile) (err error) {
	_, err = v1ms.mgoDB.DB().Collection(ColSpp).InsertOne(v1ms.mgoDB.GetContext(), spl)
	return
}

//rem
func (v1ms *mongoMigrator) remSupplier(tenant, id string) (err error) {
	_, err = v1ms.mgoDB.DB().Collection(ColSpp).DeleteOne(v1ms.mgoDB.GetContext(), bson.M{"tenant": tenant, "id": id})
	return
}

func (v1ms *mongoMigrator) getV1ChargerProfile() (v1chrPrf *engine.ChargerProfile, err error) {
	if v1ms.cursor == nil {
		v1ms.cursor, err = v1ms.mgoDB.DB().Collection(engine.ColCpp).Find(v1ms.mgoDB.GetContext(), bson.D{})
		if err != nil {
			return nil, err
		}
	}
	if !(*v1ms.cursor).Next(v1ms.mgoDB.GetContext()) {
		(*v1ms.cursor).Close(v1ms.mgoDB.GetContext())
		v1ms.cursor = nil
		return nil, utils.ErrNoMoreData
	}
	v1chrPrf = new(engine.ChargerProfile)
	if err := (*v1ms.cursor).Decode(v1chrPrf); err != nil {
		return nil, err
	}
	return
}

func (v1ms *mongoMigrator) getV1DispatcherProfile() (v1dppPrf *engine.DispatcherProfile, err error) {
	if v1ms.cursor == nil {
		v1ms.cursor, err = v1ms.mgoDB.DB().Collection(engine.ColDpp).Find(v1ms.mgoDB.GetContext(), bson.D{})
		if err != nil {
			return nil, err
		}
	}
	if !(*v1ms.cursor).Next(v1ms.mgoDB.GetContext()) {
		(*v1ms.cursor).Close(v1ms.mgoDB.GetContext())
		v1ms.cursor = nil
		return nil, utils.ErrNoMoreData
	}
	v1dppPrf = new(engine.DispatcherProfile)
	if err := (*v1ms.cursor).Decode(v1dppPrf); err != nil {
		return nil, err
	}
	return
}
