package engine

import (
	"github.com/cgrates/cgrates/utils"
	"gopkg.in/mgo.v2/bson"
)

const (
	colTpTmg = "tp_timings"
	colTpDst = "tp_destinations"
	colTpRts = "tp_rates"
	colTpDrs = "tp_destinationrates"
	colTpRpl = "tp_ratingplans"
	colTpRpf = "tp_ratingprofiles"
	colTpAct = "tp_actions"
	colTpApl = "tp_actionplans"
	colTpAtr = "tp_actiontriggers"
	colTpAcc = "tp_accounts"
	colTpShg = "tp_sharedgroups"
	colTpLcr = "tp_lcrrules"
	colTpDcs = "tp_derivedchargers"
	colTpAls = "tp_aliases"
	colTpStq = "tp_statsqeues"
	colTpPbs = "tp_pubsub"
	colTpUsr = "tp_users"
	colTpCrs = "tp_cdrstats"
)

func (ms *MongoStorage) GetTpIds() ([]string, error) { return nil, nil }
func (ms *MongoStorage) GetTpTableIds(string, string, utils.TPDistinctIds, map[string]string, *utils.Paginator) ([]string, error) {
	return nil, nil
}
func (ms *MongoStorage) GetTpTimings(tpid, tag string) ([]TpTiming, error) {
	filter := bson.M{
		"tpid": tpid,
	}
	if tag != "" {
		filter["tag"] = tag
	}
	var results []TpTiming
	err := ms.db.C(colTpTmg).Find(filter).All(&results)
	return results, err
}
func (ms *MongoStorage) GetTpDestinations(tpid, tag string) ([]TpDestination, error) {
	filter := bson.M{
		"tpid": tpid,
	}
	if tag != "" {
		filter["tag"] = tag
	}
	var results []TpDestination
	err := ms.db.C(colTpDst).Find(filter).All(&results)
	return results, err
}
func (ms *MongoStorage) GetTpRates(tpid, tag string) ([]TpRate, error) {
	filter := bson.M{
		"tpid": tpid,
	}
	if tag != "" {
		filter["tag"] = tag
	}
	var results []TpRate
	err := ms.db.C(colTpRts).Find(filter).All(&results)
	return results, err
}
func (ms *MongoStorage) GetTpDestinationRates(tpid, tag string, pag *utils.Paginator) ([]TpDestinationRate, error) {
	filter := bson.M{
		"tpid": tpid,
	}
	if tag != "" {
		filter["tag"] = tag
	}
	var results []TpDestinationRate
	q := ms.db.C(colTpDrs).Find(filter)
	if pag != nil {
		if pag.Limit != nil {
			q = q.Limit(*pag.Limit)
		}
		if pag.Offset != nil {
			q = q.Skip(*pag.Offset)
		}
	}
	err := q.All(&results)
	return results, err
}

func (ms *MongoStorage) GetTpRatingPlans(tpid, tag string, pag *utils.Paginator) ([]TpRatingPlan, error) {
	filter := bson.M{
		"tpid": tpid,
	}
	if tag != "" {
		filter["tag"] = tag
	}
	var results []TpRatingPlan
	q := ms.db.C(colTpRpl).Find(filter)
	if pag != nil {
		if pag.Limit != nil {
			q = q.Limit(*pag.Limit)
		}
		if pag.Offset != nil {
			q = q.Skip(*pag.Offset)
		}
	}
	err := q.All(&results)
	return results, err
}

func (ms *MongoStorage) GetTpRatingProfiles(tp *TpRatingProfile) ([]TpRatingProfile, error) {
	filter := bson.M{"tpid": tp.Tpid}
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
	if tp.Loadid != "" {
		filter["loadid"] = tp.Loadid
	}
	var results []TpRatingProfile
	err := ms.db.C(colTpRpf).Find(filter).All(&results)
	return results, err
}

func (ms *MongoStorage) GetTpSharedGroups(tpid, tag string) ([]TpSharedGroup, error) {
	filter := bson.M{
		"tpid": tpid,
	}
	if tag != "" {
		filter["tag"] = tag
	}
	var results []TpSharedGroup
	err := ms.db.C(colTpShg).Find(filter).All(&results)
	return results, err
}

func (ms *MongoStorage) GetTpCdrStats(tpid, tag string) ([]TpCdrstat, error) {
	filter := bson.M{
		"tpid": tpid,
	}
	if tag != "" {
		filter["tag"] = tag
	}
	var results []TpCdrstat
	err := ms.db.C(colTpCrs).Find(filter).All(&results)
	return results, err
}
func (ms *MongoStorage) GetTpLCRs(tp *TpLcrRule) ([]TpLcrRule, error) {
	filter := bson.M{"tpid": tp.Tpid}
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
	var results []TpLcrRule
	err := ms.db.C(colTpLcr).Find(filter).All(&results)
	return results, err
}

func (ms *MongoStorage) GetTpUsers(tp *TpUser) ([]TpUser, error) {
	filter := bson.M{"tpid": tp.Tpid}
	if tp.Tenant != "" {
		filter["tenant"] = tp.Tenant
	}
	if tp.UserName != "" {
		filter["username"] = tp.UserName
	}
	var results []TpUser
	err := ms.db.C(colTpUsr).Find(filter).All(&results)
	return results, err
}

func (ms *MongoStorage) GetTpAliases(tp *TpAlias) ([]TpAlias, error) {
	filter := bson.M{"tpid": tp.Tpid}
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
	var results []TpAlias
	err := ms.db.C(colTpAls).Find(filter).All(&results)
	return results, err
}

func (ms *MongoStorage) GetTpDerivedChargers(tp *TpDerivedCharger) ([]TpDerivedCharger, error) {
	filter := bson.M{"tpid": tp.Tpid}
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
	if tp.Loadid != "" {
		filter["loadid"] = tp.Loadid
	}
	var results []TpDerivedCharger
	err := ms.db.C(colTpDcs).Find(filter).All(&results)
	return results, err
}

func (ms *MongoStorage) GetTpActions(tpid, tag string) ([]TpAction, error) {
	filter := bson.M{
		"tpid": tpid,
	}
	if tag != "" {
		filter["tag"] = tag
	}
	var results []TpAction
	err := ms.db.C(colTpAct).Find(filter).All(&results)
	return results, err
}

func (ms *MongoStorage) GetTpActionPlans(tpid, tag string) ([]TpActionPlan, error) {
	filter := bson.M{
		"tpid": tpid,
	}
	if tag != "" {
		filter["tag"] = tag
	}
	var results []TpActionPlan
	err := ms.db.C(colTpApl).Find(filter).All(&results)
	return results, err
}

func (ms *MongoStorage) GetTpActionTriggers(tpid, tag string) ([]TpActionTrigger, error) {
	filter := bson.M{
		"tpid": tpid,
	}
	if tag != "" {
		filter["tag"] = tag
	}
	var results []TpActionTrigger
	err := ms.db.C(colTpAtr).Find(filter).All(&results)
	return results, err
}

func (ms *MongoStorage) GetTpAccountActions(tp *TpAccountAction) ([]TpAccountAction, error) {
	filter := bson.M{"tpid": tp.Tpid}
	if tp.Direction != "" {
		filter["direction"] = tp.Direction
	}
	if tp.Tenant != "" {
		filter["tenant"] = tp.Tenant
	}
	if tp.Account != "" {
		filter["account"] = tp.Account
	}
	if tp.Loadid != "" {
		filter["loadid"] = tp.Loadid
	}
	var results []TpAccountAction
	err := ms.db.C(colTpAcc).Find(filter).All(&results)
	return results, err
}

func (ms *MongoStorage) RemTpData(string, string, ...string) error {

	return nil
}

func (ms *MongoStorage) SetTpTimings(tps []TpTiming) error {
	if len(tps) == 0 {
		return nil
	}
	m := make(map[string]bool)

	tx := ms.db.C(colTpTmg).Bulk()
	for _, tp := range tps {
		if found, _ := m[tp.Tag]; !found {
			m[tp.Tag] = true
			tx.Upsert(bson.M{"tpid": tp.Tpid, "tag": tp.Tag}, tp)
		}
	}
	_, err := tx.Run()
	return err
}

func (ms *MongoStorage) SetTpDestinations(tps []TpDestination) error {
	if len(tps) == 0 {
		return nil
	}
	m := make(map[string]bool)

	tx := ms.db.C(colTpDst).Bulk()
	for _, tp := range tps {
		if found, _ := m[tp.Tag]; !found {
			m[tp.Tag] = true
			tx.Upsert(bson.M{"tpid": tp.Tpid, "tag": tp.Tag}, tp)
		}
	}
	_, err := tx.Run()
	return err
}

func (ms *MongoStorage) SetTpRates(tps []TpRate) error {
	if len(tps) == 0 {
		return nil
	}
	m := make(map[string]bool)

	tx := ms.db.C(colTpRts).Bulk()
	for _, tp := range tps {
		if found, _ := m[tp.Tag]; !found {
			m[tp.Tag] = true
			tx.Upsert(bson.M{"tpid": tp.Tpid, "tag": tp.Tag}, tp)
		}
	}
	_, err := tx.Run()
	return err
}

func (ms *MongoStorage) SetTpDestinationRates(tps []TpDestinationRate) error {
	if len(tps) == 0 {
		return nil
	}
	m := make(map[string]bool)

	tx := ms.db.C(colTpDrs).Bulk()
	for _, tp := range tps {
		if found, _ := m[tp.Tag]; !found {
			m[tp.Tag] = true
			tx.Upsert(bson.M{"tpid": tp.Tpid, "tag": tp.Tag}, tp)
		}
	}
	_, err := tx.Run()
	return err
}

func (ms *MongoStorage) SetTpRatingPlans(tps []TpRatingPlan) error {
	if len(tps) == 0 {
		return nil
	}
	m := make(map[string]bool)

	tx := ms.db.C(colTpRpl).Bulk()
	for _, tp := range tps {
		if found, _ := m[tp.Tag]; !found {
			m[tp.Tag] = true
			tx.Upsert(bson.M{"tpid": tp.Tpid, "tag": tp.Tag}, tp)
		}
	}
	_, err := tx.Run()
	return err
}

func (ms *MongoStorage) SetTpRatingProfiles(tps []TpRatingProfile) error {
	if len(tps) == 0 {
		return nil
	}
	m := make(map[string]bool)

	tx := ms.db.C(colTpRpf).Bulk()
	for _, tp := range tps {
		if found, _ := m[tp.GetRatingProfileId()]; !found {
			m[tp.GetRatingProfileId()] = true
			tx.Upsert(bson.M{
				"tpid":      tp.Tpid,
				"loadid":    tp.Loadid,
				"direction": tp.Direction,
				"tenant":    tp.Tenant,
				"category":  tp.Category,
				"subject":   tp.Subject,
			}, tp)
		}
	}
	_, err := tx.Run()
	return err
}

func (ms *MongoStorage) SetTpSharedGroups(tps []TpSharedGroup) error {
	if len(tps) == 0 {
		return nil
	}
	m := make(map[string]bool)

	tx := ms.db.C(colTpShg).Bulk()
	for _, tp := range tps {
		if found, _ := m[tp.Tag]; !found {
			m[tp.Tag] = true
			tx.Upsert(bson.M{"tpid": tp.Tpid, "tag": tp.Tag}, tp)
		}
	}
	_, err := tx.Run()
	return err
}

func (ms *MongoStorage) SetTpCdrStats(tps []TpCdrstat) error {
	if len(tps) == 0 {
		return nil
	}
	m := make(map[string]bool)

	tx := ms.db.C(colTpCrs).Bulk()
	for _, tp := range tps {
		if found, _ := m[tp.Tag]; !found {
			m[tp.Tag] = true
			tx.Upsert(bson.M{"tpid": tp.Tpid, "tag": tp.Tag}, tp)
		}
	}
	_, err := tx.Run()
	return err
}

func (ms *MongoStorage) SetTpUsers(tps []TpUser) error {
	if len(tps) == 0 {
		return nil
	}
	m := make(map[string]bool)

	tx := ms.db.C(colTpUsr).Bulk()
	for _, tp := range tps {
		if found, _ := m[tp.GetId()]; !found {
			m[tp.GetId()] = true
			tx.Upsert(bson.M{
				"tpid":     tp.Tpid,
				"tenant":   tp.Tenant,
				"username": tp.UserName,
			}, tp)
		}
	}
	_, err := tx.Run()
	return err
}

func (ms *MongoStorage) SetTpAliases(tps []TpAlias) error {
	if len(tps) == 0 {
		return nil
	}
	m := make(map[string]bool)

	tx := ms.db.C(colTpAls).Bulk()
	for _, tp := range tps {
		if found, _ := m[tp.GetId()]; !found {
			m[tp.GetId()] = true
			tx.Upsert(bson.M{
				"tpid":      tp.Tpid,
				"direction": tp.Direction,
				"tenant":    tp.Tenant,
				"category":  tp.Category,
				"account":   tp.Account,
				"subject":   tp.Subject,
				"context":   tp.Context}, tp)
		}
	}
	_, err := tx.Run()
	return err
}

func (ms *MongoStorage) SetTpDerivedChargers(tps []TpDerivedCharger) error {
	if len(tps) == 0 {
		return nil
	}
	m := make(map[string]bool)

	tx := ms.db.C(colTpRpl).Bulk()
	for _, tp := range tps {
		if found, _ := m[tp.GetDerivedChargersId()]; !found {
			m[tp.GetDerivedChargersId()] = true
			tx.Upsert(bson.M{
				"tpid":      tp.Tpid,
				"direction": tp.Direction,
				"tenant":    tp.Tenant,
				"category":  tp.Category,
				"account":   tp.Account,
				"subject":   tp.Subject}, tp)
		}
	}
	_, err := tx.Run()
	return err
}

func (ms *MongoStorage) SetTpLCRs(tps []TpLcrRule) error {
	if len(tps) == 0 {
		return nil
	}
	m := make(map[string]bool)

	tx := ms.db.C(colTpRpl).Bulk()
	for _, tp := range tps {
		if found, _ := m[tp.GetLcrRuleId()]; !found {
			m[tp.GetLcrRuleId()] = true
			tx.Upsert(bson.M{
				"tpid":      tp.Tpid,
				"direction": tp.Direction,
				"tenant":    tp.Tenant,
				"category":  tp.Category,
				"account":   tp.Account,
				"subject":   tp.Subject}, tp)
		}
	}
	_, err := tx.Run()
	return err
}

func (ms *MongoStorage) SetTpActions(tps []TpAction) error {
	if len(tps) == 0 {
		return nil
	}
	m := make(map[string]bool)

	tx := ms.db.C(colTpAct).Bulk()
	for _, tp := range tps {
		if found, _ := m[tp.Tag]; !found {
			m[tp.Tag] = true
			tx.Upsert(bson.M{"tpid": tp.Tpid, "tag": tp.Tag}, tp)
		}
	}
	_, err := tx.Run()
	return err
}

func (ms *MongoStorage) SetTpActionPlans(tps []TpActionPlan) error {
	if len(tps) == 0 {
		return nil
	}
	m := make(map[string]bool)

	tx := ms.db.C(colTpApl).Bulk()
	for _, tp := range tps {
		if found, _ := m[tp.Tag]; !found {
			m[tp.Tag] = true
			tx.Upsert(bson.M{"tpid": tp.Tpid, "tag": tp.Tag}, tp)
		}
	}
	_, err := tx.Run()
	return err
}

func (ms *MongoStorage) SetTpActionTriggers(tps []TpActionTrigger) error {
	if len(tps) == 0 {
		return nil
	}
	m := make(map[string]bool)

	tx := ms.db.C(colTpAtr).Bulk()
	for _, tp := range tps {
		if found, _ := m[tp.Tag]; !found {
			m[tp.Tag] = true
			tx.Upsert(bson.M{"tpid": tp.Tpid, "tag": tp.Tag}, tp)
		}
	}
	_, err := tx.Run()
	return err
}

func (ms *MongoStorage) SetTpAccountActions(tps []TpAccountAction) error {
	if len(tps) == 0 {
		return nil
	}
	m := make(map[string]bool)

	tx := ms.db.C(colTpAcc).Bulk()
	for _, tp := range tps {
		if found, _ := m[tp.GetAccountActionId()]; !found {
			m[tp.GetAccountActionId()] = true
			tx.Upsert(bson.M{
				"tpid":      tp.Tpid,
				"loadid":    tp.Loadid,
				"direction": tp.Direction,
				"tenant":    tp.Tenant,
				"account":   tp.Account}, tp)
		}
	}
	_, err := tx.Run()
	return err
}
