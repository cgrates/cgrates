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
along with this program. If not, see <http://www.gnu.org/licenses/>
*/

package engine

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/mgo"
	"github.com/cgrates/mgo/bson"
)

func (ms *MongoStorage) GetTpIds(colName string) ([]string, error) {
	tpidMap := make(map[string]bool)
	session := ms.session.Copy()
	db := session.DB(ms.db)
	defer session.Close()
	var tpids []string
	var err error
	cols := []string{colName}
	if colName == "" {
		cols, err = db.CollectionNames()
		if err != nil {
			return nil, err
		}
	}
	for _, col := range cols {
		if strings.HasPrefix(col, "tp_") {
			tpids := make([]string, 0)
			if err := db.C(col).Find(nil).Select(bson.M{"tpid": 1}).Distinct("tpid", &tpids); err != nil {
				return nil, err
			}
			for _, tpid := range tpids {
				tpidMap[tpid] = true
			}
		}
	}
	for tpid := range tpidMap {
		tpids = append(tpids, tpid)
	}
	return tpids, nil
}

func (ms *MongoStorage) GetTpTableIds(tpid, table string, distinct utils.TPDistinctIds, filter map[string]string, pag *utils.Paginator) ([]string, error) {
	findMap := make(map[string]interface{})
	if tpid != "" {
		findMap["tpid"] = tpid
	}
	for k, v := range filter {
		findMap[k] = v
	}
	for k, v := range distinct { //fix for MongoStorage on TPUsers
		if v == "user_name" {
			distinct[k] = "username"
		}
	}
	if pag != nil && pag.SearchTerm != "" {
		var searchItems []bson.M
		for _, d := range distinct {
			searchItems = append(searchItems, bson.M{d: bson.RegEx{
				Pattern: ".*" + regexp.QuoteMeta(pag.SearchTerm) + ".*",
				Options: ""}})
		}
		findMap["$and"] = []bson.M{bson.M{"$or": searchItems}}
	}

	session, col := ms.conn(table)
	defer session.Close()

	q := col.Find(findMap)
	if pag != nil {
		if pag.Limit != nil {
			q = q.Limit(*pag.Limit)
		}
		if pag.Offset != nil {
			q = q.Skip(*pag.Offset)
		}
	}

	selectors := bson.M{"_id": 0}
	for i, d := range distinct {
		if d == "tag" { // convert the tag used in SQL into id used here
			distinct[i] = "id"
		}
		selectors[distinct[i]] = 1
	}
	iter := q.Select(selectors).Iter()
	distinctIds := make(utils.StringMap)
	item := make(map[string]string)
	for iter.Next(item) {
		var id string
		last := len(distinct) - 1
		for i, d := range distinct {
			if distinctValue, ok := item[d]; ok {
				id += distinctValue
			}
			if i < last {
				id += utils.CONCATENATED_KEY_SEP
			}
		}
		distinctIds[id] = true
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return distinctIds.Slice(), nil
}

func (ms *MongoStorage) GetTPTimings(tpid, id string) ([]*utils.ApierTPTiming, error) {
	filter := bson.M{
		"tpid": tpid,
	}
	if id != "" {
		filter["id"] = id
	}
	var results []*utils.ApierTPTiming
	session, col := ms.conn(utils.TBLTPTimings)
	defer session.Close()
	err := col.Find(filter).All(&results)
	if len(results) == 0 {
		return results, utils.ErrNotFound
	}
	return results, err
}

func (ms *MongoStorage) GetTPDestinations(tpid, id string) ([]*utils.TPDestination, error) {
	filter := bson.M{
		"tpid": tpid,
	}
	if id != "" {
		filter["id"] = id
	}
	var results []*utils.TPDestination
	session, col := ms.conn(utils.TBLTPDestinations)
	defer session.Close()
	err := col.Find(filter).All(&results)
	if len(results) == 0 {
		return results, utils.ErrNotFound
	}
	return results, err
}

func (ms *MongoStorage) GetTPRates(tpid, id string) ([]*utils.TPRate, error) {
	filter := bson.M{
		"tpid": tpid,
	}
	if id != "" {
		filter["id"] = id
	}
	var results []*utils.TPRate
	session, col := ms.conn(utils.TBLTPRates)
	defer session.Close()
	err := col.Find(filter).All(&results)
	if len(results) == 0 {
		return results, utils.ErrNotFound
	}
	for _, r := range results {
		for _, rs := range r.RateSlots {
			rs.SetDurations()
		}
	}
	return results, err
}

func (ms *MongoStorage) GetTPDestinationRates(tpid, id string, pag *utils.Paginator) ([]*utils.TPDestinationRate, error) {
	filter := bson.M{
		"tpid": tpid,
	}
	if id != "" {
		filter["id"] = id
	}
	var results []*utils.TPDestinationRate
	session, col := ms.conn(utils.TBLTPDestinationRates)
	defer session.Close()
	q := col.Find(filter)
	if pag != nil {
		if pag.Limit != nil {
			q = q.Limit(*pag.Limit)
		}
		if pag.Offset != nil {
			q = q.Skip(*pag.Offset)
		}
	}
	err := q.All(&results)
	if len(results) == 0 {
		return results, utils.ErrNotFound
	}
	return results, err
}

func (ms *MongoStorage) GetTPRatingPlans(tpid, id string, pag *utils.Paginator) ([]*utils.TPRatingPlan, error) {
	filter := bson.M{
		"tpid": tpid,
	}
	if id != "" {
		filter["id"] = id
	}
	var results []*utils.TPRatingPlan
	session, col := ms.conn(utils.TBLTPRatingPlans)
	defer session.Close()
	q := col.Find(filter)
	if pag != nil {
		if pag.Limit != nil {
			q = q.Limit(*pag.Limit)
		}
		if pag.Offset != nil {
			q = q.Skip(*pag.Offset)
		}
	}
	err := q.All(&results)
	if len(results) == 0 {
		return results, utils.ErrNotFound
	}
	return results, err
}

func (ms *MongoStorage) GetTPRatingProfiles(tp *utils.TPRatingProfile) ([]*utils.TPRatingProfile, error) {
	filter := bson.M{"tpid": tp.TPid}
	if tp.Direction != "" {
		filter["direction"] = tp.Direction
	}
	if tp.Tenant != "" {
		filter["tenant"] = tp.Tenant
	}
	if tp.Category != "" {
		filter["category"] = tp.Category
	}
	if tp.Subject != "" {
		filter["subject"] = tp.Subject
	}
	if tp.LoadId != "" {
		filter["loadid"] = tp.LoadId
	}
	var results []*utils.TPRatingProfile
	session, col := ms.conn(utils.TBLTPRateProfiles)
	defer session.Close()
	err := col.Find(filter).All(&results)
	if len(results) == 0 {
		return results, utils.ErrNotFound
	}
	return results, err
}

func (ms *MongoStorage) GetTPSharedGroups(tpid, id string) ([]*utils.TPSharedGroups, error) {
	filter := bson.M{
		"tpid": tpid,
	}
	if id != "" {
		filter["id"] = id
	}
	var results []*utils.TPSharedGroups
	session, col := ms.conn(utils.TBLTPSharedGroups)
	defer session.Close()
	err := col.Find(filter).All(&results)
	if len(results) == 0 {
		return results, utils.ErrNotFound
	}
	return results, err
}

func (ms *MongoStorage) GetTPCdrStats(tpid, id string) ([]*utils.TPCdrStats, error) {
	filter := bson.M{
		"tpid": tpid,
	}
	if id != "" {
		filter["id"] = id
	}
	var results []*utils.TPCdrStats
	session, col := ms.conn(utils.TBLTPCdrStats)
	defer session.Close()
	err := col.Find(filter).All(&results)
	if len(results) == 0 {
		return results, utils.ErrNotFound
	}
	return results, err
}

func (ms *MongoStorage) GetTPLCRs(tp *utils.TPLcrRules) ([]*utils.TPLcrRules, error) {
	filter := bson.M{"tpid": tp.TPid}
	if tp.Direction != "" {
		filter["direction"] = tp.Direction
	}
	if tp.Tenant != "" {
		filter["tenant"] = tp.Tenant
	}
	if tp.Category != "" {
		filter["category"] = tp.Category
	}
	if tp.Account != "" {
		filter["account"] = tp.Account
	}
	if tp.Subject != "" {
		filter["subject"] = tp.Subject
	}
	var results []*utils.TPLcrRules
	session, col := ms.conn(utils.TBLTPLcrs)
	defer session.Close()
	err := col.Find(filter).All(&results)
	if len(results) == 0 {
		return results, utils.ErrNotFound
	}
	return results, err
}

func (ms *MongoStorage) GetTPUsers(tp *utils.TPUsers) ([]*utils.TPUsers, error) {
	filter := bson.M{"tpid": tp.TPid}
	if tp.Tenant != "" {
		filter["tenant"] = tp.Tenant
	}
	if tp.UserName != "" {
		filter["username"] = tp.UserName
	}
	var results []*utils.TPUsers
	session, col := ms.conn(utils.TBLTPUsers)
	defer session.Close()
	err := col.Find(filter).All(&results)
	if len(results) == 0 {
		return results, utils.ErrNotFound
	}
	return results, err
}

func (ms *MongoStorage) GetTPAliases(tp *utils.TPAliases) ([]*utils.TPAliases, error) {
	filter := bson.M{"tpid": tp.TPid}
	if tp.Direction != "" {
		filter["direction"] = tp.Direction
	}
	if tp.Tenant != "" {
		filter["tenant"] = tp.Tenant
	}
	if tp.Category != "" {
		filter["category"] = tp.Category
	}
	if tp.Account != "" {
		filter["account"] = tp.Account
	}
	if tp.Subject != "" {
		filter["subject"] = tp.Subject
	}
	if tp.Context != "" {
		filter["context"] = tp.Context
	}
	var results []*utils.TPAliases
	session, col := ms.conn(utils.TBLTPAliases)
	defer session.Close()
	err := col.Find(filter).All(&results)
	if len(results) == 0 {
		return results, utils.ErrNotFound
	}
	return results, err
}

func (ms *MongoStorage) GetTPResources(tpid, id string) ([]*utils.TPResource, error) {
	filter := bson.M{
		"tpid": tpid,
	}
	if id != "" {
		filter["id"] = id
	}
	var results []*utils.TPResource
	session, col := ms.conn(utils.TBLTPResources)
	defer session.Close()
	err := col.Find(filter).All(&results)
	if len(results) == 0 {
		return results, utils.ErrNotFound
	}
	return results, err
}

func (ms *MongoStorage) GetTPStats(tpid, id string) ([]*utils.TPStats, error) {
	filter := bson.M{
		"tpid": tpid,
	}
	if id != "" {
		filter["id"] = id
	}
	var results []*utils.TPStats
	session, col := ms.conn(utils.TBLTPStats)
	defer session.Close()
	err := col.Find(filter).All(&results)
	if len(results) == 0 {
		return results, utils.ErrNotFound
	}
	return results, err
}

func (ms *MongoStorage) GetTPDerivedChargers(tp *utils.TPDerivedChargers) ([]*utils.TPDerivedChargers, error) {
	filter := bson.M{"tpid": tp.TPid}
	if tp.Direction != "" {
		filter["direction"] = tp.Direction
	}
	if tp.Tenant != "" {
		filter["tenant"] = tp.Tenant
	}
	if tp.Category != "" {
		filter["category"] = tp.Category
	}
	if tp.Subject != "" {
		filter["subject"] = tp.Subject
	}
	if tp.Account != "" {
		filter["account"] = tp.Account
	}
	if tp.LoadId != "" {
		filter["loadid"] = tp.LoadId
	}
	var results []*utils.TPDerivedChargers
	session, col := ms.conn(utils.TBLTPDerivedChargers)
	defer session.Close()
	err := col.Find(filter).All(&results)
	if len(results) == 0 {
		return results, utils.ErrNotFound
	}
	return results, err
}

func (ms *MongoStorage) GetTPActions(tpid, id string) ([]*utils.TPActions, error) {
	filter := bson.M{
		"tpid": tpid,
	}
	if id != "" {
		filter["id"] = id
	}
	var results []*utils.TPActions
	session, col := ms.conn(utils.TBLTPActions)
	defer session.Close()
	err := col.Find(filter).All(&results)
	if len(results) == 0 {
		return results, utils.ErrNotFound
	}
	return results, err
}

func (ms *MongoStorage) GetTPActionPlans(tpid, id string) ([]*utils.TPActionPlan, error) {
	filter := bson.M{
		"tpid": tpid,
	}
	if id != "" {
		filter["id"] = id
	}
	var results []*utils.TPActionPlan
	session, col := ms.conn(utils.TBLTPActionPlans)
	defer session.Close()
	err := col.Find(filter).All(&results)
	if len(results) == 0 {
		return results, utils.ErrNotFound
	}
	return results, err
}

func (ms *MongoStorage) GetTPActionTriggers(tpid, id string) ([]*utils.TPActionTriggers, error) {
	filter := bson.M{
		"tpid": tpid,
	}
	if id != "" {
		filter["id"] = id
	}
	var results []*utils.TPActionTriggers
	session, col := ms.conn(utils.TBLTPActionTriggers)
	defer session.Close()
	err := col.Find(filter).All(&results)
	if len(results) == 0 {
		return results, utils.ErrNotFound
	}
	return results, err
}

func (ms *MongoStorage) GetTPAccountActions(tp *utils.TPAccountActions) ([]*utils.TPAccountActions, error) {
	filter := bson.M{"tpid": tp.TPid}
	if tp.Tenant != "" {
		filter["tenant"] = tp.Tenant
	}
	if tp.Account != "" {
		filter["account"] = tp.Account
	}
	if tp.LoadId != "" {
		filter["loadid"] = tp.LoadId
	}
	var results []*utils.TPAccountActions
	session, col := ms.conn(utils.TBLTPAccountActions)
	defer session.Close()
	err := col.Find(filter).All(&results)
	if len(results) == 0 {
		return results, utils.ErrNotFound
	}
	return results, err
}

func (ms *MongoStorage) RemTpData(table, tpid string, args map[string]string) error {
	session := ms.session.Copy()
	db := session.DB(ms.db)
	defer session.Close()
	if len(table) == 0 { // Remove tpid out of all tables
		cols, err := db.CollectionNames()
		if err != nil {
			return err
		}
		for _, col := range cols {
			if strings.HasPrefix(col, "tp_") {
				if _, err := db.C(col).RemoveAll(bson.M{"tpid": tpid}); err != nil {
					return err
				}
			}
		}
		return nil
	}
	// Remove from a single table
	if args == nil {
		args = make(map[string]string)
	}
	for arg, val := range args { //fix for Mongo TPUsers tables
		if arg == "user_name" {
			delete(args, arg)
			args["username"] = val
		}
	}

	if _, has := args["tag"]; has { // API uses tag to be compatible with SQL models, fix it here
		args["id"] = args["tag"]
		delete(args, "tag")
	}
	if tpid != "" {
		args["tpid"] = tpid
	}
	return db.C(table).Remove(args)
}

func (ms *MongoStorage) SetTPTimings(tps []*utils.ApierTPTiming) error {
	if len(tps) == 0 {
		return nil
	}
	session, col := ms.conn(utils.TBLTPTimings)
	defer session.Close()
	tx := col.Bulk()
	for _, tp := range tps {
		tx.Upsert(bson.M{"tpid": tp.TPid, "id": tp.ID}, tp)
	}
	_, err := tx.Run()
	return err
}

func (ms *MongoStorage) SetTPDestinations(tpDsts []*utils.TPDestination) (err error) {
	if len(tpDsts) == 0 {
		return
	}
	session, col := ms.conn(utils.TBLTPDestinations)
	defer session.Close()
	tx := col.Bulk()
	for _, tp := range tpDsts {
		tx.Upsert(bson.M{"tpid": tp.TPid, "id": tp.ID}, tp)
	}
	_, err = tx.Run()
	return
}

func (ms *MongoStorage) SetTPRates(tps []*utils.TPRate) error {
	if len(tps) == 0 {
		return nil
	}
	m := make(map[string]bool)
	session, col := ms.conn(utils.TBLTPRates)
	defer session.Close()
	tx := col.Bulk()
	for _, tp := range tps {
		if found, _ := m[tp.ID]; !found {
			m[tp.ID] = true
			tx.RemoveAll(bson.M{"tpid": tp.TPid, "id": tp.ID})
		}
		tx.Insert(tp)
	}
	_, err := tx.Run()
	return err
}

func (ms *MongoStorage) SetTPDestinationRates(tps []*utils.TPDestinationRate) error {
	if len(tps) == 0 {
		return nil
	}
	m := make(map[string]bool)
	session, col := ms.conn(utils.TBLTPDestinationRates)
	defer session.Close()
	tx := col.Bulk()
	for _, tp := range tps {
		if found, _ := m[tp.ID]; !found {
			m[tp.ID] = true
			tx.RemoveAll(bson.M{"tpid": tp.TPid, "id": tp.ID})
		}
		tx.Insert(tp)
	}
	_, err := tx.Run()
	return err
}

func (ms *MongoStorage) SetTPRatingPlans(tps []*utils.TPRatingPlan) error {
	if len(tps) == 0 {
		return nil
	}
	m := make(map[string]bool)
	session, col := ms.conn(utils.TBLTPRatingPlans)
	defer session.Close()
	tx := col.Bulk()
	for _, tp := range tps {
		if found, _ := m[tp.ID]; !found {
			m[tp.ID] = true
			tx.RemoveAll(bson.M{"tpid": tp.TPid, "id": tp.ID})
		}
		tx.Insert(tp)
	}
	_, err := tx.Run()
	return err
}

func (ms *MongoStorage) SetTPRatingProfiles(tps []*utils.TPRatingProfile) error {
	if len(tps) == 0 {
		return nil
	}
	session, col := ms.conn(utils.TBLTPRateProfiles)
	defer session.Close()
	tx := col.Bulk()
	for _, tp := range tps {
		tx.Upsert(bson.M{
			"tpid":      tp.TPid,
			"loadid":    tp.LoadId,
			"direction": tp.Direction,
			"tenant":    tp.Tenant,
			"category":  tp.Category,
			"subject":   tp.Subject,
		}, tp)
	}
	_, err := tx.Run()
	return err
}

func (ms *MongoStorage) SetTPSharedGroups(tps []*utils.TPSharedGroups) error {
	if len(tps) == 0 {
		return nil
	}
	m := make(map[string]bool)
	session, col := ms.conn(utils.TBLTPSharedGroups)
	defer session.Close()
	tx := col.Bulk()
	for _, tp := range tps {
		if found, _ := m[tp.ID]; !found {
			m[tp.ID] = true
			tx.RemoveAll(bson.M{"tpid": tp.TPid, "id": tp.ID})
		}
		tx.Insert(tp)
	}
	_, err := tx.Run()
	return err
}

func (ms *MongoStorage) SetTPCdrStats(tps []*utils.TPCdrStats) error {
	if len(tps) == 0 {
		return nil
	}
	m := make(map[string]bool)
	session, col := ms.conn(utils.TBLTPCdrStats)
	defer session.Close()
	tx := col.Bulk()
	for _, tp := range tps {
		if found, _ := m[tp.ID]; !found {
			m[tp.ID] = true
			tx.RemoveAll(bson.M{"tpid": tp.TPid, "id": tp.ID})
		}
		tx.Insert(tp)
	}
	_, err := tx.Run()
	return err
}

func (ms *MongoStorage) SetTPUsers(tps []*utils.TPUsers) error {
	if len(tps) == 0 {
		return nil
	}
	m := make(map[string]bool)
	session, col := ms.conn(utils.TBLTPUsers)
	defer session.Close()
	tx := col.Bulk()
	for _, tp := range tps {
		if found, _ := m[tp.GetId()]; !found {
			m[tp.GetId()] = true
			tx.RemoveAll(bson.M{
				"tpid":     tp.TPid,
				"tenant":   tp.Tenant,
				"username": tp.UserName,
			})
		}
		tx.Insert(tp)
	}
	_, err := tx.Run()
	return err
}

func (ms *MongoStorage) SetTPAliases(tps []*utils.TPAliases) error {
	if len(tps) == 0 {
		return nil
	}
	m := make(map[string]bool)
	session, col := ms.conn(utils.TBLTPAliases)
	defer session.Close()
	tx := col.Bulk()
	for _, tp := range tps {
		if found, _ := m[tp.Direction]; !found {
			m[tp.Direction] = true
			tx.RemoveAll(bson.M{
				"tpid":      tp.TPid,
				"direction": tp.Direction,
				"tenant":    tp.Tenant,
				"category":  tp.Category,
				"account":   tp.Account,
				"subject":   tp.Subject,
				"context":   tp.Context})
		}
		tx.Insert(tp)
	}
	_, err := tx.Run()
	return err
}

func (ms *MongoStorage) SetTPDerivedChargers(tps []*utils.TPDerivedChargers) error {
	if len(tps) == 0 {
		return nil
	}
	m := make(map[string]bool)
	session, col := ms.conn(utils.TBLTPDerivedChargers)
	defer session.Close()
	tx := col.Bulk()
	for _, tp := range tps {
		if found, _ := m[tp.Direction]; !found {
			m[tp.Direction] = true
			tx.RemoveAll(bson.M{
				"tpid":      tp.TPid,
				"direction": tp.Direction,
				"tenant":    tp.Tenant,
				"category":  tp.Category,
				"account":   tp.Account,
				"subject":   tp.Subject})
		}
		tx.Insert(tp)
	}
	_, err := tx.Run()
	return err
}

func (ms *MongoStorage) SetTPLCRs(tps []*utils.TPLcrRules) error {
	if len(tps) == 0 {
		return nil
	}
	session, col := ms.conn(utils.TBLTPLcrs)
	defer session.Close()
	tx := col.Bulk()
	for _, tp := range tps {
		tx.Upsert(bson.M{
			"tpid":      tp.TPid,
			"direction": tp.Direction,
			"tenant":    tp.Tenant,
			"category":  tp.Category,
			"account":   tp.Account,
			"subject":   tp.Subject}, tp)

	}
	_, err := tx.Run()
	return err
}

func (ms *MongoStorage) SetTPActions(tps []*utils.TPActions) error {
	if len(tps) == 0 {
		return nil
	}
	m := make(map[string]bool)
	session, col := ms.conn(utils.TBLTPActions)
	defer session.Close()
	tx := col.Bulk()
	for _, tp := range tps {
		if found, _ := m[tp.ID]; !found {
			m[tp.ID] = true
			tx.RemoveAll(bson.M{"tpid": tp.TPid, "id": tp.ID})
		}
		tx.Insert(tp)
	}
	_, err := tx.Run()
	return err
}

func (ms *MongoStorage) SetTPActionPlans(tps []*utils.TPActionPlan) error {
	if len(tps) == 0 {
		return nil
	}
	m := make(map[string]bool)
	session, col := ms.conn(utils.TBLTPActionPlans)
	defer session.Close()
	tx := col.Bulk()
	for _, tp := range tps {
		if found, _ := m[tp.ID]; !found {
			m[tp.ID] = true
			tx.RemoveAll(bson.M{"tpid": tp.TPid, "id": tp.ID})
		}
		tx.Insert(tp)
	}
	_, err := tx.Run()
	return err
}

func (ms *MongoStorage) SetTPActionTriggers(tps []*utils.TPActionTriggers) error {
	if len(tps) == 0 {
		return nil
	}
	m := make(map[string]bool)
	session, col := ms.conn(utils.TBLTPActionTriggers)
	defer session.Close()
	tx := col.Bulk()
	for _, tp := range tps {
		if found, _ := m[tp.ID]; !found {
			m[tp.ID] = true
			tx.RemoveAll(bson.M{"tpid": tp.TPid, "id": tp.ID})
		}
		tx.Insert(tp)
	}
	_, err := tx.Run()
	return err
}

func (ms *MongoStorage) SetTPAccountActions(tps []*utils.TPAccountActions) error {
	if len(tps) == 0 {
		return nil
	}
	session, col := ms.conn(utils.TBLTPAccountActions)
	defer session.Close()
	tx := col.Bulk()
	for _, tp := range tps {
		tx.Upsert(bson.M{
			"tpid":    tp.TPid,
			"loadid":  tp.LoadId,
			"tenant":  tp.Tenant,
			"account": tp.Account}, tp)
	}
	_, err := tx.Run()
	return err
}

func (ms *MongoStorage) SetTPResources(tpRLs []*utils.TPResource) (err error) {
	if len(tpRLs) == 0 {
		return
	}
	session, col := ms.conn(utils.TBLTPResources)
	defer session.Close()
	tx := col.Bulk()
	for _, tp := range tpRLs {
		tx.Upsert(bson.M{"tpid": tp.TPid, "id": tp.ID}, tp)
	}
	_, err = tx.Run()
	return
}

func (ms *MongoStorage) SetTPRStats(tpS []*utils.TPStats) (err error) {
	if len(tpS) == 0 {
		return
	}
	session, col := ms.conn(utils.TBLTPStats)
	defer session.Close()
	tx := col.Bulk()
	for _, tp := range tpS {
		tx.Upsert(bson.M{"tpid": tp.TPid, "id": tp.ID}, tp)
	}
	_, err = tx.Run()
	return
}

func (ms *MongoStorage) SetSMCost(smc *SMCost) error {
	if smc.CostDetails == nil {
		return nil
	}
	session, col := ms.conn(utils.SessionsCostsTBL)
	defer session.Close()
	return col.Insert(smc)
}

func (ms *MongoStorage) RemoveSMCost(smc *SMCost) error {
	session, col := ms.conn(utils.SessionsCostsTBL)
	defer session.Close()
	remParams := bson.M{}
	if smc != nil {
		remParams = bson.M{"cgrid": smc.CGRID, "runid": smc.RunID}
	}
	tx := col.Bulk()
	tx.RemoveAll(remParams)
	_, err := tx.Run()
	return err
}

func (ms *MongoStorage) GetSMCosts(cgrid, runid, originHost, originIDPrefix string) (smcs []*SMCost, err error) {
	filter := bson.M{}
	if cgrid != "" {
		filter[CGRIDLow] = cgrid
	}
	if runid != "" {
		filter[RunIDLow] = runid
	}
	if originHost != "" {
		filter[OriginHostLow] = originHost
	}
	if originIDPrefix != "" {
		filter[OriginIDLow] = bson.M{"$regex": bson.RegEx{Pattern: fmt.Sprintf("^%s", originIDPrefix)}}
	}
	// Execute query
	session, col := ms.conn(utils.SessionsCostsTBL)
	defer session.Close()
	iter := col.Find(filter).Iter()
	var smCost SMCost
	for iter.Next(&smCost) {
		clone := smCost
		smcs = append(smcs, &clone)
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}
	if len(smcs) == 0 {
		return smcs, utils.ErrNotFound
	}
	return smcs, nil
}

func (ms *MongoStorage) SetCDR(cdr *CDR, allowUpdate bool) (err error) {
	if cdr.OrderID == 0 {
		cdr.OrderID = ms.cnter.Next()
	}
	session, col := ms.conn(ColCDRs)
	defer session.Close()
	if allowUpdate {
		_, err = col.Upsert(bson.M{CGRIDLow: cdr.CGRID, RunIDLow: cdr.RunID}, cdr)
	} else {
		err = col.Insert(cdr)
	}
	return err
}

func (ms *MongoStorage) cleanEmptyFilters(filters bson.M) {
	for k, v := range filters {
		switch value := v.(type) {
		case *int64:
			if value == nil {
				delete(filters, k)
			}
		case *float64:
			if value == nil {
				delete(filters, k)
			}
		case *time.Time:
			if value == nil {
				delete(filters, k)
			}
		case *time.Duration:
			if value == nil {
				delete(filters, k)
			}
		case []string:
			if len(value) == 0 {
				delete(filters, k)
			}
		case bson.M:
			ms.cleanEmptyFilters(value)
			if len(value) == 0 {
				delete(filters, k)
			}
		}
	}
}

//  _, err := col(ColCDRs).UpdateAll(bson.M{CGRIDLow: bson.M{"$in": cgrIds}}, bson.M{"$set": bson.M{"deleted_at": time.Now()}})
func (ms *MongoStorage) GetCDRs(qryFltr *utils.CDRsFilter, remove bool) ([]*CDR, int64, error) {
	var minUsage, maxUsage *time.Duration
	if len(qryFltr.MinUsage) != 0 {
		if parsed, err := utils.ParseDurationWithNanosecs(qryFltr.MinUsage); err != nil {
			return nil, 0, err
		} else {
			minUsage = &parsed
		}
	}
	if len(qryFltr.MaxUsage) != 0 {
		if parsed, err := utils.ParseDurationWithNanosecs(qryFltr.MaxUsage); err != nil {
			return nil, 0, err
		} else {
			maxUsage = &parsed
		}
	}
	filters := bson.M{
		CGRIDLow:       bson.M{"$in": qryFltr.CGRIDs, "$nin": qryFltr.NotCGRIDs},
		RunIDLow:       bson.M{"$in": qryFltr.RunIDs, "$nin": qryFltr.NotRunIDs},
		OriginIDLow:    bson.M{"$in": qryFltr.OriginIDs, "$nin": qryFltr.NotOriginIDs},
		OrderIDLow:     bson.M{"$gte": qryFltr.OrderIDStart, "$lt": qryFltr.OrderIDEnd},
		ToRLow:         bson.M{"$in": qryFltr.ToRs, "$nin": qryFltr.NotToRs},
		CDRHostLow:     bson.M{"$in": qryFltr.OriginHosts, "$nin": qryFltr.NotOriginHosts},
		CDRSourceLow:   bson.M{"$in": qryFltr.Sources, "$nin": qryFltr.NotSources},
		RequestTypeLow: bson.M{"$in": qryFltr.RequestTypes, "$nin": qryFltr.NotRequestTypes},
		TenantLow:      bson.M{"$in": qryFltr.Tenants, "$nin": qryFltr.NotTenants},
		CategoryLow:    bson.M{"$in": qryFltr.Categories, "$nin": qryFltr.NotCategories},
		AccountLow:     bson.M{"$in": qryFltr.Accounts, "$nin": qryFltr.NotAccounts},
		SubjectLow:     bson.M{"$in": qryFltr.Subjects, "$nin": qryFltr.NotSubjects},
		SetupTimeLow:   bson.M{"$gte": qryFltr.SetupTimeStart, "$lt": qryFltr.SetupTimeEnd},
		AnswerTimeLow:  bson.M{"$gte": qryFltr.AnswerTimeStart, "$lt": qryFltr.AnswerTimeEnd},
		CreatedAtLow:   bson.M{"$gte": qryFltr.CreatedAtStart, "$lt": qryFltr.CreatedAtEnd},
		UpdatedAtLow:   bson.M{"$gte": qryFltr.UpdatedAtStart, "$lt": qryFltr.UpdatedAtEnd},
		UsageLow:       bson.M{"$gte": minUsage, "$lt": maxUsage},
		//CostDetailsLow + "." + AccountLow: bson.M{"$in": qryFltr.RatedAccounts, "$nin": qryFltr.NotRatedAccounts},
		//CostDetailsLow + "." + SubjectLow: bson.M{"$in": qryFltr.RatedSubjects, "$nin": qryFltr.NotRatedSubjects},
	}
	//file, _ := ioutil.TempFile(os.TempDir(), "debug")
	//file.WriteString(fmt.Sprintf("FILTER: %v\n", utils.ToIJSON(qryFltr)))
	//file.WriteString(fmt.Sprintf("BEFORE: %v\n", utils.ToIJSON(filters)))
	ms.cleanEmptyFilters(filters)
	if len(qryFltr.DestinationPrefixes) != 0 {
		var regexpRule string
		for _, prefix := range qryFltr.DestinationPrefixes {
			if len(prefix) == 0 {
				continue
			}
			if len(regexpRule) != 0 {
				regexpRule += "|"
			}
			regexpRule += "^(" + prefix + ")"
		}
		if _, hasIt := filters["$and"]; !hasIt {
			filters["$and"] = make([]bson.M, 0)
		}
		filters["$and"] = append(filters["$and"].([]bson.M), bson.M{DestinationLow: bson.RegEx{Pattern: regexpRule}}) // $and gathers all rules not fitting top level query
	}
	if len(qryFltr.NotDestinationPrefixes) != 0 {
		if _, hasIt := filters["$and"]; !hasIt {
			filters["$and"] = make([]bson.M, 0)
		}
		for _, prefix := range qryFltr.NotDestinationPrefixes {
			if len(prefix) == 0 {
				continue
			}
			filters["$and"] = append(filters["$and"].([]bson.M), bson.M{DestinationLow: bson.RegEx{Pattern: "^(?!" + prefix + ")"}})
		}
	}

	if len(qryFltr.ExtraFields) != 0 {
		var extrafields []bson.M
		for field, value := range qryFltr.ExtraFields {
			if value == utils.MetaExists {
				extrafields = append(extrafields, bson.M{"extrafields." + field: bson.M{"$exists": true}})
			} else {
				extrafields = append(extrafields, bson.M{"extrafields." + field: value})
			}
		}
		filters["$and"] = extrafields
	}

	if len(qryFltr.NotExtraFields) != 0 {
		var extrafields []bson.M
		for field, value := range qryFltr.NotExtraFields {
			if value == utils.MetaExists {
				extrafields = append(extrafields, bson.M{"extrafields." + field: bson.M{"$exists": false}})
			} else {
				extrafields = append(extrafields, bson.M{"extrafields." + field: bson.M{"$ne": value}})
			}

		}
		filters["$and"] = extrafields
	}

	if qryFltr.MinCost != nil {
		if qryFltr.MaxCost == nil {
			filters[CostLow] = bson.M{"$gte": *qryFltr.MinCost}
		} else if *qryFltr.MinCost == 0.0 && *qryFltr.MaxCost == -1.0 { // Special case when we want to skip errors
			filters["$or"] = []bson.M{
				bson.M{CostLow: bson.M{"$gte": 0.0}},
			}
		} else {
			filters[CostLow] = bson.M{"$gte": *qryFltr.MinCost, "$lt": *qryFltr.MaxCost}
		}
	} else if qryFltr.MaxCost != nil {
		if *qryFltr.MaxCost == -1.0 { // Non-rated CDRs
			filters[CostLow] = 0.0 // Need to include it otherwise all CDRs will be returned
		} else { // Above limited CDRs, since MinCost is empty, make sure we query also NULL cost
			filters[CostLow] = bson.M{"$lt": *qryFltr.MaxCost}
		}
	}
	//file.WriteString(fmt.Sprintf("AFTER: %v\n", utils.ToIJSON(filters)))
	//file.Close()
	session, col := ms.conn(ColCDRs)
	defer session.Close()
	if remove {
		if chgd, err := col.RemoveAll(filters); err != nil {
			return nil, 0, err
		} else {
			return nil, int64(chgd.Removed), nil
		}
	}
	q := col.Find(filters)
	if qryFltr.Paginator.Limit != nil {
		q = q.Limit(*qryFltr.Paginator.Limit)
	}
	if qryFltr.Paginator.Offset != nil {
		q = q.Skip(*qryFltr.Paginator.Offset)
	}
	if qryFltr.OrderBy != "" {
		var orderVal string
		separateVals := strings.Split(qryFltr.OrderBy, utils.INFIELD_SEP)
		if len(separateVals) == 2 && separateVals[1] == "desc" {
			orderVal += "-"
		}
		switch separateVals[0] {
		case utils.OrderID:
			orderVal += "orderid"
		case utils.AnswerTime:
			orderVal += "answertime"
		case utils.SetupTime:
			orderVal += "setuptime"
		case utils.Usage:
			orderVal += "usage"
		case utils.Cost:
			orderVal += "cost"
		default:
			return nil, 0, fmt.Errorf("Invalid value : %s", separateVals[0])
		}
		q = q.Sort(orderVal)
	}
	if qryFltr.Count {
		cnt, err := q.Count()
		if err != nil {
			return nil, 0, err
		}
		return nil, int64(cnt), nil
	}
	// Execute query
	iter := q.Iter()
	var cdrs []*CDR
	cdr := CDR{}
	for iter.Next(&cdr) {
		clone := cdr
		cdrs = append(cdrs, &clone)
	}
	if len(cdrs) == 0 {
		return cdrs, 0, utils.ErrNotFound
	}
	return cdrs, 0, nil
}

func (ms *MongoStorage) GetTPStat(tpid, id string) ([]*utils.TPStats, error) {
	filter := bson.M{
		"tpid": tpid,
	}
	if id != "" {
		filter["id"] = id
	}
	var results []*utils.TPStats
	session, col := ms.conn(utils.TBLTPStats)
	defer session.Close()
	err := col.Find(filter).All(&results)
	if len(results) == 0 {
		return results, utils.ErrNotFound
	}
	return results, err
}

func (ms *MongoStorage) SetTPStats(tpSTs []*utils.TPStats) (err error) {
	if len(tpSTs) == 0 {
		return
	}
	session, col := ms.conn(utils.TBLTPStats)
	defer session.Close()
	tx := col.Bulk()
	for _, tp := range tpSTs {
		tx.Upsert(bson.M{"tpid": tp.TPid, "id": tp.ID}, tp)
	}
	_, err = tx.Run()
	return
}

func (ms *MongoStorage) GetTPThresholds(tpid, id string) ([]*utils.TPThreshold, error) {
	filter := bson.M{
		"tpid": tpid,
	}
	if id != "" {
		filter["id"] = id
	}
	var results []*utils.TPThreshold
	session, col := ms.conn(utils.TBLTPThresholds)
	defer session.Close()
	err := col.Find(filter).All(&results)
	if len(results) == 0 {
		return results, utils.ErrNotFound
	}
	return results, err
}

func (ms *MongoStorage) SetTPThresholds(tpTHs []*utils.TPThreshold) (err error) {
	if len(tpTHs) == 0 {
		return
	}
	session, col := ms.conn(utils.TBLTPThresholds)
	defer session.Close()
	tx := col.Bulk()
	for _, tp := range tpTHs {
		tx.Upsert(bson.M{"tpid": tp.TPid, "id": tp.ID}, tp)
	}
	_, err = tx.Run()
	return
}

func (ms *MongoStorage) GetTPFilters(tpid, id string) ([]*utils.TPFilterProfile, error) {
	filter := bson.M{
		"tpid": tpid,
	}
	if id != "" {
		filter["id"] = id
	}
	var results []*utils.TPFilterProfile
	session, col := ms.conn(utils.TBLTPFilters)
	defer session.Close()
	err := col.Find(filter).All(&results)
	if len(results) == 0 {
		return results, utils.ErrNotFound
	}
	return results, err
}

func (ms *MongoStorage) SetTPFilters(tpTHs []*utils.TPFilterProfile) (err error) {
	if len(tpTHs) == 0 {
		return
	}
	session, col := ms.conn(utils.TBLTPFilters)
	defer session.Close()
	tx := col.Bulk()
	for _, tp := range tpTHs {
		tx.Upsert(bson.M{"tpid": tp.TPid, "id": tp.ID}, tp)
	}
	_, err = tx.Run()
	return
}

func (ms *MongoStorage) GetTPSuppliers(tpid, id string) ([]*utils.TPSupplierProfile, error) {
	filter := bson.M{
		"tpid": tpid,
	}
	if id != "" {
		filter["id"] = id
	}
	var results []*utils.TPSupplierProfile
	session, col := ms.conn(utils.TBLTPSuppliers)
	defer session.Close()
	err := col.Find(filter).All(&results)
	if len(results) == 0 {
		return results, utils.ErrNotFound
	}
	return results, err
}

func (ms *MongoStorage) SetTPSuppliers(tpSPs []*utils.TPSupplierProfile) (err error) {
	if len(tpSPs) == 0 {
		return
	}
	session, col := ms.conn(utils.TBLTPSuppliers)
	defer session.Close()
	tx := col.Bulk()
	for _, tp := range tpSPs {
		tx.Upsert(bson.M{"tpid": tp.TPid, "id": tp.ID}, tp)
	}
	_, err = tx.Run()
	return
}

func (ms *MongoStorage) GetTPAttributes(tpid, id string) ([]*utils.TPAttributeProfile, error) {
	filter := bson.M{
		"tpid": tpid,
	}
	if id != "" {
		filter["id"] = id
	}
	var results []*utils.TPAttributeProfile
	session, col := ms.conn(utils.TBLTPAttributes)
	defer session.Close()
	err := col.Find(filter).All(&results)
	if len(results) == 0 {
		return results, utils.ErrNotFound
	}
	return results, err
}

func (ms *MongoStorage) SetTPAttributes(tpSPs []*utils.TPAttributeProfile) (err error) {
	if len(tpSPs) == 0 {
		return
	}
	session, col := ms.conn(utils.TBLTPAttributes)
	defer session.Close()
	tx := col.Bulk()
	for _, tp := range tpSPs {
		tx.Upsert(bson.M{"tpid": tp.TPid, "id": tp.ID}, tp)
	}
	_, err = tx.Run()
	return
}

func (ms *MongoStorage) GetTPChargers(tpid, id string) ([]*utils.TPChargerProfile, error) {
	filter := bson.M{
		"tpid": tpid,
	}
	if id != "" {
		filter["id"] = id
	}
	var results []*utils.TPChargerProfile
	session, col := ms.conn(utils.TBLTPChargers)
	defer session.Close()
	err := col.Find(filter).All(&results)
	if len(results) == 0 {
		return results, utils.ErrNotFound
	}
	return results, err
}

func (ms *MongoStorage) SetTPChargers(tpCPP []*utils.TPChargerProfile) (err error) {
	if len(tpCPP) == 0 {
		return
	}
	session, col := ms.conn(utils.TBLTPChargers)
	defer session.Close()
	tx := col.Bulk()
	for _, tp := range tpCPP {
		tx.Upsert(bson.M{"tpid": tp.TPid, "id": tp.ID}, tp)
	}
	_, err = tx.Run()
	return
}

func (ms *MongoStorage) GetVersions(itm string) (vrs Versions, err error) {
	session, col := ms.conn(colVer)
	defer session.Close()
	proj := bson.M{} // projection params
	if itm != "" {
		proj[itm] = 1
	}
	if err = col.Find(bson.M{}).Select(proj).One(&vrs); err != nil {
		if err == mgo.ErrNotFound {
			err = utils.ErrNotFound
		}
		return nil, err
	}
	if len(vrs) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MongoStorage) SetVersions(vrs Versions, overwrite bool) (err error) {
	session, col := ms.conn(colVer)
	defer session.Close()
	if overwrite {
		_, err = col.Upsert(bson.M{}, &vrs)
		return
	}
	_, err = col.Upsert(bson.M{}, bson.M{"$set": &vrs})
	return
}

func (ms *MongoStorage) RemoveVersions(vrs Versions) (err error) {
	session, col := ms.conn(colVer)
	defer session.Close()
	if len(vrs) != 0 {
		var pairs []interface{}
		for k := range vrs {
			pairs = append(pairs, bson.M{}) // match first
			pairs = append(pairs, bson.M{"$unset": bson.M{k: 1}})
		}
		bulk := col.Bulk()
		bulk.Unordered()
		bulk.Upsert(pairs...)
		_, err = bulk.Run()
	} else {
		err = col.Remove(bson.M{})
	}
	if err == mgo.ErrNotFound {
		err = utils.ErrNotFound
	}
	return
}

func (ms *MongoStorage) GetStorageType() string {
	return utils.MONGO
}
