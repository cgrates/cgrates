package engine

import (
	"strings"
	"time"

	"github.com/cgrates/cgrates/utils"
	"gopkg.in/mgo.v2/bson"
)

func (ms *MongoStorage) GetTpIds() ([]string, error) {
	tpidMap := make(map[string]bool)
	cols, err := ms.db.CollectionNames()
	if err != nil {
		return nil, err
	}
	for _, col := range cols {
		if strings.HasPrefix(col, "tp_") {
			tpids := make([]string, 0)
			if err := ms.db.C(col).Find(nil).Select(bson.M{"tpid": 1}).Distinct("tpid", &tpids); err != nil {
				return nil, err
			}
			for _, tpid := range tpids {
				tpidMap[tpid] = true
			}
		}
	}
	var tpids []string
	for tpid := range tpidMap {
		tpids = append(tpids, tpid)
	}
	return tpids, nil
}
func (ms *MongoStorage) GetTpTableIds(tpid, table string, distinct utils.TPDistinctIds, filter map[string]string, pag *utils.Paginator) ([]string, error) {
	selectors := bson.M{}
	for _, d := range distinct {
		selectors[d] = 1
	}
	findMap := make(map[string]interface{})
	if tpid != "" {
		findMap["tpid"] = tpid
	}
	for k, v := range filter {
		findMap[k] = v
	}

	if pag != nil && pag.SearchTerm != "" {
		var searchItems []bson.M
		for _, d := range distinct {
			searchItems = append(searchItems, bson.M{d: bson.RegEx{
				Pattern: ".*" + pag.SearchTerm + ".*",
				Options: ""}})
		}
		findMap["$and"] = []bson.M{bson.M{"$or": searchItems}}
	}
	q := ms.db.C(table).Find(findMap)
	if pag != nil {
		if pag.Limit != nil {
			q = q.Limit(*pag.Limit)
		}
		if pag.Offset != nil {
			q = q.Skip(*pag.Offset)
		}
	}

	iter := q.Select(selectors).Iter()
	distinctIds := make(map[string]bool)
	item := make(map[string]string)
	for iter.Next(item) {
		id := ""
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
	if err := iter.Close(); err != nil {
		return nil, err
	}
	var results []string
	for id := range distinctIds {
		results = append(results, id)
	}
	return results, nil
}

func (ms *MongoStorage) GetTpTimings(tpid, tag string) ([]TpTiming, error) {
	filter := bson.M{
		"tpid": tpid,
	}
	if tag != "" {
		filter["tag"] = tag
	}
	var results []TpTiming
	err := ms.db.C(utils.TBL_TP_TIMINGS).Find(filter).All(&results)
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
	err := ms.db.C(utils.TBL_TP_DESTINATIONS).Find(filter).All(&results)
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
	err := ms.db.C(utils.TBL_TP_RATES).Find(filter).All(&results)
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
	q := ms.db.C(utils.TBL_TP_DESTINATION_RATES).Find(filter)
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
	q := ms.db.C(utils.TBL_TP_RATING_PLANS).Find(filter)
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
	err := ms.db.C(utils.TBL_TP_RATE_PROFILES).Find(filter).All(&results)
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
	err := ms.db.C(utils.TBL_TP_SHARED_GROUPS).Find(filter).All(&results)
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
	err := ms.db.C(utils.TBL_TP_CDR_STATS).Find(filter).All(&results)
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
	err := ms.db.C(utils.TBL_TP_LCRS).Find(filter).All(&results)
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
	err := ms.db.C(utils.TBL_TP_USERS).Find(filter).All(&results)
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
	err := ms.db.C(utils.TBL_TP_ALIASES).Find(filter).All(&results)
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
	err := ms.db.C(utils.TBL_TP_DERIVED_CHARGERS).Find(filter).All(&results)
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
	err := ms.db.C(utils.TBL_TP_ACTIONS).Find(filter).All(&results)
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
	err := ms.db.C(utils.TBL_TP_ACTION_PLANS).Find(filter).All(&results)
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
	err := ms.db.C(utils.TBL_TP_ACTION_TRIGGERS).Find(filter).All(&results)
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
	err := ms.db.C(utils.TBL_TP_ACCOUNT_ACTIONS).Find(filter).All(&results)
	return results, err
}

func (ms *MongoStorage) RemTpData(table, tpid string, args map[string]string) error {
	if len(table) == 0 { // Remove tpid out of all tables
		cols, err := ms.db.CollectionNames()
		if err != nil {
			return err
		}
		for _, col := range cols {
			if strings.HasPrefix(col, "tp_") {
				if err := ms.db.C(col).Remove(bson.M{"tpid": tpid}); err != nil {
					return err
				}
			}
		}
	}
	// Remove from a single table
	args["tpid"] = tpid
	return ms.db.C(table).Remove(args)
}

func (ms *MongoStorage) SetTpTimings(tps []TpTiming) error {
	if len(tps) == 0 {
		return nil
	}
	m := make(map[string]bool)

	tx := ms.db.C(utils.TBL_TP_TIMINGS).Bulk()
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

	tx := ms.db.C(utils.TBL_TP_DESTINATIONS).Bulk()
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

	tx := ms.db.C(utils.TBL_TP_RATES).Bulk()
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

	tx := ms.db.C(utils.TBL_TP_DESTINATION_RATES).Bulk()
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

	tx := ms.db.C(utils.TBL_TP_RATING_PLANS).Bulk()
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

	tx := ms.db.C(utils.TBL_TP_RATE_PROFILES).Bulk()
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

	tx := ms.db.C(utils.TBL_TP_SHARED_GROUPS).Bulk()
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

	tx := ms.db.C(utils.TBL_TP_CDR_STATS).Bulk()
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

	tx := ms.db.C(utils.TBL_TP_USERS).Bulk()
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

	tx := ms.db.C(utils.TBL_TP_ALIASES).Bulk()
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

	tx := ms.db.C(utils.TBL_TP_DERIVED_CHARGERS).Bulk()
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

	tx := ms.db.C(utils.TBL_TP_LCRS).Bulk()
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

	tx := ms.db.C(utils.TBL_TP_ACTIONS).Bulk()
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

	tx := ms.db.C(utils.TBL_TP_ACTION_PLANS).Bulk()
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

	tx := ms.db.C(utils.TBL_TP_ACTION_TRIGGERS).Bulk()
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

	tx := ms.db.C(utils.TBL_TP_ACCOUNT_ACTIONS).Bulk()
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

func (ms *MongoStorage) LogActionTrigger(ubId, source string, at *ActionTrigger, as Actions) (err error) {
	return ms.db.C(colLogAtr).Insert(&struct {
		ubId          string
		ActionTrigger *ActionTrigger
		Actions       Actions
		LogTime       time.Time
		Source        string
	}{ubId, at, as, time.Now(), source})
}

func (ms *MongoStorage) LogActionPlan(source string, at *ActionPlan, as Actions) (err error) {
	return ms.db.C(colLogApl).Insert(&struct {
		ActionPlan *ActionPlan
		Actions    Actions
		LogTime    time.Time
		Source     string
	}{at, as, time.Now(), source})
}

func (ms *MongoStorage) LogCallCost(cgrid, source, runid string, cc *CallCost) error {
	return ms.db.C(colLogCC).Insert(&struct {
		Id       string `bson:"id,omitempty"`
		Source   string
		Runid    string `bson:"runid,omitempty"`
		CallCost *CallCost
	}{cgrid, source, runid, cc})
}

func (ms *MongoStorage) GetCallCostLog(cgrid, source, runid string) (cc *CallCost, err error) {
	result := &struct {
		Id       string `bson:"id,omitempty"`
		Source   string
		Runid    string `bson:"runid,omitempty"`
		CallCost *CallCost
	}{}
	err = ms.db.C(colLogCC).Find(bson.M{"_id": cgrid, "source": source}).One(result)
	cc = result.CallCost
	return
}

func (ms *MongoStorage) SetCdr(cdr *StoredCdr) error {
	return ms.db.C(colCdrs).Insert(cdr)
}

func (ms *MongoStorage) SetRatedCdr(storedCdr *StoredCdr) error {
	return ms.db.C(colRatedCdrs).Insert(storedCdr)
}

// Remove CDR data out of all CDR tables based on their cgrid
func (ms *MongoStorage) RemStoredCdrs(cgrIds []string) error {
	if len(cgrIds) == 0 {
		return nil
	}

	for _, col := range []string{colCdrs, colRatedCdrs, colLogCC} {
		if err := ms.db.C(col).Update(bson.M{"cgrid": bson.M{"$in": cgrIds}}, map[string]interface{}{"deleted_at": time.Now()}); err != nil {
			return err
		}
	}
	return nil
}

func (ms *MongoStorage) GetStoredCdrs(qryFltr *utils.CdrsFilter) ([]*StoredCdr, int64, error) {
	return nil, 0, utils.ErrNotImplemented
}
