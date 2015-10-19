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
				if _, err := ms.db.C(col).RemoveAll(bson.M{"tpid": tpid}); err != nil {
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
	s := &StoredCdr{
		CgrId:          cgrid,
		CdrSource:      source,
		MediationRunId: runid,
		CostDetails:    cc,
	}
	_, err := ms.db.C(colCdrs).Upsert(bson.M{"cgrid": cgrid, "cdrsource": source, "mediationrunid": runid}, s)
	return err
}

func (ms *MongoStorage) GetCallCostLog(cgrid, source, runid string) (cc *CallCost, err error) {
	result := StoredCdr{}
	err = ms.db.C(colCdrs).Find(bson.M{"cgrid": cgrid, "cdrsource": source, "mediationrunid": runid}).One(&result)
	cc = result.CostDetails
	return
}

func (ms *MongoStorage) SetCdr(cdr *StoredCdr) error {
	_, err := ms.db.C(colCdrs).Upsert(bson.M{"cgrid": cdr.CgrId, "mediationrunid": cdr.MediationRunId}, cdr)
	return err
}

func (ms *MongoStorage) SetRatedCdr(storedCdr *StoredCdr) error {
	_, err := ms.db.C(colCdrs).Upsert(bson.M{"cgrid": storedCdr.CgrId, "mediationrunid": storedCdr.MediationRunId}, storedCdr)
	return err
}

// Remove CDR data out of all CDR tables based on their cgrid
func (ms *MongoStorage) RemStoredCdrs(cgrIds []string) error {
	if len(cgrIds) == 0 {
		return nil
	}
	_, err := ms.db.C(colCdrs).UpdateAll(bson.M{"cgrid": bson.M{"$in": cgrIds}}, bson.M{"$set": bson.M{"deleted_at": time.Now()}})
	return err
}

func (ms *MongoStorage) cleanEmptyFilters(filters bson.M) {
	for k, v := range filters {
		switch value := v.(type) {
		case *float64:
			if value == nil {
				delete(filters, k)
			}
		case *time.Time:
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

func (ms *MongoStorage) GetStoredCdrs(qryFltr *utils.CdrsFilter) ([]*StoredCdr, int64, error) {
	filters := bson.M{
		"cgrid":            bson.M{"$in": qryFltr.CgrIds, "$nin": qryFltr.NotCgrIds},
		"mediationrunid":   bson.M{"$in": qryFltr.RunIds, "$nin": qryFltr.NotRunIds},
		"tor":              bson.M{"$in": qryFltr.Tors, "$nin": qryFltr.NotTors},
		"cdrhost":          bson.M{"$in": qryFltr.CdrHosts, "$nin": qryFltr.NotCdrHosts},
		"cdrsource":        bson.M{"$in": qryFltr.CdrSources, "$nin": qryFltr.NotCdrSources},
		"reqtype":          bson.M{"$in": qryFltr.ReqTypes, "$nin": qryFltr.NotReqTypes},
		"direction":        bson.M{"$in": qryFltr.Directions, "$nin": qryFltr.NotDirections},
		"tenant":           bson.M{"$in": qryFltr.Tenants, "$nin": qryFltr.NotTenants},
		"category":         bson.M{"$in": qryFltr.Categories, "$nin": qryFltr.NotCategories},
		"account":          bson.M{"$in": qryFltr.Accounts, "$nin": qryFltr.NotAccounts},
		"subject":          bson.M{"$in": qryFltr.Subjects, "$nin": qryFltr.NotSubjects},
		"supplier":         bson.M{"$in": qryFltr.Suppliers, "$nin": qryFltr.NotSuppliers},
		"disconnect_cause": bson.M{"$in": qryFltr.DisconnectCauses, "$nin": qryFltr.NotDisconnectCauses},
		"setuptime":        bson.M{"$gte": qryFltr.SetupTimeStart, "$lt": qryFltr.SetupTimeEnd},
		"answertime":       bson.M{"$gte": qryFltr.AnswerTimeStart, "$lt": qryFltr.AnswerTimeEnd},
		"created_at":       bson.M{"$gte": qryFltr.CreatedAtStart, "$lt": qryFltr.CreatedAtEnd},
		"updated_at":       bson.M{"$gte": qryFltr.UpdatedAtStart, "$lt": qryFltr.UpdatedAtEnd},
		"usage":            bson.M{"$gte": qryFltr.MinUsage, "$lt": qryFltr.MaxUsage},
		"pdd":              bson.M{"$gte": qryFltr.MinPdd, "$lt": qryFltr.MaxPdd},
		"costdetails.account": bson.M{"$in": qryFltr.RatedAccounts, "$nin": qryFltr.NotRatedAccounts},
		"costdetails.subject": bson.M{"$in": qryFltr.RatedSubjects, "$nin": qryFltr.NotRatedSubjects},
	}
	//file, _ := ioutil.TempFile(os.TempDir(), "debug")
	//file.WriteString(fmt.Sprintf("FILTER: %v\n", utils.ToIJSON(qryFltr)))
	//file.WriteString(fmt.Sprintf("BEFORE: %v\n", utils.ToIJSON(filters)))
	ms.cleanEmptyFilters(filters)

	if qryFltr.OrderIdStart != 0 {
		filters["id"] = bson.M{"$gte": qryFltr.OrderIdStart}
	}
	if qryFltr.OrderIdEnd != 0 {
		if m, ok := filters["id"]; ok {
			m.(bson.M)["$gte"] = qryFltr.OrderIdStart
		} else {
			filters["id"] = bson.M{"$gte": qryFltr.OrderIdStart}
		}
	}

	if len(qryFltr.DestPrefixes) != 0 {
		var regexes []bson.RegEx
		for _, prefix := range qryFltr.DestPrefixes {
			regexes = append(regexes, bson.RegEx{Pattern: prefix + ".*"})
		}
		filters["destination"] = bson.M{"$in": regexes}
	}
	if len(qryFltr.NotDestPrefixes) != 0 {
		var notRegexes []bson.RegEx
		for _, prefix := range qryFltr.DestPrefixes {
			notRegexes = append(notRegexes, bson.RegEx{Pattern: prefix + ".*"})
		}
		if m, ok := filters["destination"]; ok {
			m.(bson.M)["$nin"] = notRegexes
		} else {
			filters["destination"] = bson.M{"$nin": notRegexes}
		}
	}

	if len(qryFltr.ExtraFields) != 0 {
		var extrafields []bson.M
		for field, value := range qryFltr.ExtraFields {
			extrafields = append(extrafields, bson.M{"extrafields." + field: value})
		}
		filters["$or"] = extrafields
	}

	if len(qryFltr.NotExtraFields) != 0 {
		var extrafields []bson.M
		for field, value := range qryFltr.ExtraFields {
			extrafields = append(extrafields, bson.M{"extrafields." + field: value})
		}
		filters["$not"] = bson.M{"$or": extrafields}
	}

	if qryFltr.MinCost != nil {
		if qryFltr.MaxCost == nil {
			filters["cost"] = bson.M{"$gte": *qryFltr.MinCost}
		} else if *qryFltr.MinCost == 0.0 && *qryFltr.MaxCost == -1.0 { // Special case when we want to skip errors
			filters["$or"] = []bson.M{
				bson.M{"cost": bson.M{"$gte": 0.0}},
			}
		} else {
			filters["cost"] = bson.M{"$gte": *qryFltr.MinCost, "$lt": *qryFltr.MaxCost}
		}
	} else if qryFltr.MaxCost != nil {
		if *qryFltr.MaxCost == -1.0 { // Non-rated CDRs
			filters["cost"] = 0.0 // Need to include it otherwise all CDRs will be returned
		} else { // Above limited CDRs, since MinCost is empty, make sure we query also NULL cost
			filters["cost"] = bson.M{"$lt": *qryFltr.MaxCost}
		}
	}
	//file.WriteString(fmt.Sprintf("AFTER: %v\n", utils.ToIJSON(filters)))
	//file.Close()
	q := ms.db.C(colCdrs).Find(filters)
	if qryFltr.Paginator.Limit != nil {
		q = q.Limit(*qryFltr.Paginator.Limit)
	}
	if qryFltr.Paginator.Offset != nil {
		q = q.Skip(*qryFltr.Paginator.Offset)
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
	var cdrs []*StoredCdr
	cdr := StoredCdr{}
	for iter.Next(&cdr) {
		clone := cdr
		cdrs = append(cdrs, &clone)
	}
	return cdrs, 0, nil
}
