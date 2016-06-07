package engine

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/cgrates/cgrates/utils"
	"gopkg.in/mgo.v2/bson"
)

func (ms *MongoStorage) GetTpIds() ([]string, error) {
	tpidMap := make(map[string]bool)
	session := ms.session.Copy()
	db := session.DB(ms.db)
	defer session.Close()
	cols, err := db.CollectionNames()
	if err != nil {
		return nil, err
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
	session, col := ms.conn(utils.TBL_TP_TIMINGS)
	defer session.Close()
	err := col.Find(filter).All(&results)
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
	session, col := ms.conn(utils.TBL_TP_DESTINATIONS)
	defer session.Close()
	err := col.Find(filter).All(&results)
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
	session, col := ms.conn(utils.TBL_TP_RATES)
	defer session.Close()
	err := col.Find(filter).All(&results)
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
	session, col := ms.conn(utils.TBL_TP_DESTINATION_RATES)
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
	session, col := ms.conn(utils.TBL_TP_RATING_PLANS)
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
	session, col := ms.conn(utils.TBL_TP_RATE_PROFILES)
	defer session.Close()
	err := col.Find(filter).All(&results)
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
	session, col := ms.conn(utils.TBL_TP_SHARED_GROUPS)
	defer session.Close()
	err := col.Find(filter).All(&results)
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
	session, col := ms.conn(utils.TBL_TP_CDR_STATS)
	defer session.Close()
	err := col.Find(filter).All(&results)
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
	session, col := ms.conn(utils.TBL_TP_LCRS)
	defer session.Close()
	err := col.Find(filter).All(&results)
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
	session, col := ms.conn(utils.TBL_TP_USERS)
	defer session.Close()
	err := col.Find(filter).All(&results)
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
	session, col := ms.conn(utils.TBL_TP_ALIASES)
	defer session.Close()
	err := col.Find(filter).All(&results)
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
	session, col := ms.conn(utils.TBL_TP_DERIVED_CHARGERS)
	defer session.Close()
	err := col.Find(filter).All(&results)
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
	session, col := ms.conn(utils.TBL_TP_ACTIONS)
	defer session.Close()
	err := col.Find(filter).All(&results)
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
	session, col := ms.conn(utils.TBL_TP_ACTION_PLANS)
	defer session.Close()
	err := col.Find(filter).All(&results)
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
	session, col := ms.conn(utils.TBL_TP_ACTION_TRIGGERS)
	defer session.Close()
	err := col.Find(filter).All(&results)
	return results, err
}

func (ms *MongoStorage) GetTpAccountActions(tp *TpAccountAction) ([]TpAccountAction, error) {
	filter := bson.M{"tpid": tp.Tpid}
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
	session, col := ms.conn(utils.TBL_TP_ACCOUNT_ACTIONS)
	defer session.Close()
	err := col.Find(filter).All(&results)
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
	args["tpid"] = tpid
	return db.C(table).Remove(args)
}

func (ms *MongoStorage) SetTpTimings(tps []TpTiming) error {
	if len(tps) == 0 {
		return nil
	}
	m := make(map[string]bool)
	session, col := ms.conn(utils.TBL_TP_TIMINGS)
	defer session.Close()
	tx := col.Bulk()
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
	session, col := ms.conn(utils.TBL_TP_DESTINATIONS)
	defer session.Close()
	tx := col.Bulk()
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
	session, col := ms.conn(utils.TBL_TP_RATES)
	defer session.Close()
	tx := col.Bulk()
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
	session, col := ms.conn(utils.TBL_TP_DESTINATION_RATES)
	defer session.Close()
	tx := col.Bulk()
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
	session, col := ms.conn(utils.TBL_TP_RATING_PLANS)
	defer session.Close()
	tx := col.Bulk()
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
	session, col := ms.conn(utils.TBL_TP_RATE_PROFILES)
	defer session.Close()
	tx := col.Bulk()
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
	session, col := ms.conn(utils.TBL_TP_SHARED_GROUPS)
	defer session.Close()
	tx := col.Bulk()
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
	session, col := ms.conn(utils.TBL_TP_CDR_STATS)
	defer session.Close()
	tx := col.Bulk()
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
	session, col := ms.conn(utils.TBL_TP_USERS)
	defer session.Close()
	tx := col.Bulk()
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
	session, col := ms.conn(utils.TBL_TP_ALIASES)
	defer session.Close()
	tx := col.Bulk()
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
	session, col := ms.conn(utils.TBL_TP_DERIVED_CHARGERS)
	defer session.Close()
	tx := col.Bulk()
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
	session, col := ms.conn(utils.TBL_TP_LCRS)
	defer session.Close()
	tx := col.Bulk()
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
	session, col := ms.conn(utils.TBL_TP_ACTIONS)
	defer session.Close()
	tx := col.Bulk()
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
	session, col := ms.conn(utils.TBL_TP_ACTION_PLANS)
	defer session.Close()
	tx := col.Bulk()
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
	session, col := ms.conn(utils.TBL_TP_ACTION_TRIGGERS)
	defer session.Close()
	tx := col.Bulk()
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
	session, col := ms.conn(utils.TBL_TP_ACCOUNT_ACTIONS)
	defer session.Close()
	tx := col.Bulk()
	for _, tp := range tps {
		if found, _ := m[tp.GetAccountActionId()]; !found {
			m[tp.GetAccountActionId()] = true
			tx.Upsert(bson.M{
				"tpid":    tp.Tpid,
				"loadid":  tp.Loadid,
				"tenant":  tp.Tenant,
				"account": tp.Account}, tp)
		}
	}
	_, err := tx.Run()
	return err
}

func (ms *MongoStorage) LogActionTrigger(ubId, source string, at *ActionTrigger, as Actions) (err error) {
	session, col := ms.conn(colLogAtr)
	defer session.Close()
	return col.Insert(&struct {
		ubId          string
		ActionTrigger *ActionTrigger
		Actions       Actions
		LogTime       time.Time
		Source        string
	}{ubId, at, as, time.Now(), source})
}

func (ms *MongoStorage) LogActionTiming(source string, at *ActionTiming, as Actions) (err error) {
	session, col := ms.conn(colLogApl)
	defer session.Close()
	return col.Insert(&struct {
		ActionPlan *ActionTiming
		Actions    Actions
		LogTime    time.Time
		Source     string
	}{at, as, time.Now(), source})
}

func (ms *MongoStorage) SetSMCost(smc *SMCost) error {
	session, col := ms.conn(utils.TBLSMCosts)
	defer session.Close()
	return col.Insert(smc)
}

func (ms *MongoStorage) GetSMCosts(cgrid, runid, originHost, originIDPrefix string) (smcs []*SMCost, err error) {
	filter := bson.M{CGRIDLow: cgrid, RunIDLow: runid}
	if originIDPrefix != "" {
		filter = bson.M{OriginIDLow: bson.M{"$regex": bson.RegEx{Pattern: fmt.Sprintf("^%s", originIDPrefix)}}, OriginHostLow: originHost, RunIDLow: runid}
	}
	// Execute query
	session, col := ms.conn(utils.TBLSMCosts)
	defer session.Close()
	iter := col.Find(filter).Iter()
	var smCost SMCost
	for iter.Next(&smCost) {
		smcs = append(smcs, &smCost)
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}
	return smcs, nil
}

func (ms *MongoStorage) SetCDR(cdr *CDR, allowUpdate bool) (err error) {
	if cdr.OrderID == 0 {
		cdr.OrderID = time.Now().UnixNano()
	}
	session, col := ms.conn(utils.TBL_CDRS)
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

//  _, err := col(utils.TBL_CDRS).UpdateAll(bson.M{CGRIDLow: bson.M{"$in": cgrIds}}, bson.M{"$set": bson.M{"deleted_at": time.Now()}})
func (ms *MongoStorage) GetCDRs(qryFltr *utils.CDRsFilter, remove bool) ([]*CDR, int64, error) {
	var minPDD, maxPDD, minUsage, maxUsage *time.Duration
	if len(qryFltr.MinPDD) != 0 {
		if parsed, err := utils.ParseDurationWithSecs(qryFltr.MinPDD); err != nil {
			return nil, 0, err
		} else {
			minPDD = &parsed
		}
	}
	if len(qryFltr.MaxPDD) != 0 {
		if parsed, err := utils.ParseDurationWithSecs(qryFltr.MaxPDD); err != nil {
			return nil, 0, err
		} else {
			maxPDD = &parsed
		}
	}
	if len(qryFltr.MinUsage) != 0 {
		if parsed, err := utils.ParseDurationWithSecs(qryFltr.MinUsage); err != nil {
			return nil, 0, err
		} else {
			minUsage = &parsed
		}
	}
	if len(qryFltr.MaxUsage) != 0 {
		if parsed, err := utils.ParseDurationWithSecs(qryFltr.MaxUsage); err != nil {
			return nil, 0, err
		} else {
			maxUsage = &parsed
		}
	}
	filters := bson.M{
		CGRIDLow:           bson.M{"$in": qryFltr.CGRIDs, "$nin": qryFltr.NotCGRIDs},
		RunIDLow:           bson.M{"$in": qryFltr.RunIDs, "$nin": qryFltr.NotRunIDs},
		OrderIDLow:         bson.M{"$gte": qryFltr.OrderIDStart, "$lt": qryFltr.OrderIDEnd},
		ToRLow:             bson.M{"$in": qryFltr.ToRs, "$nin": qryFltr.NotToRs},
		CDRHostLow:         bson.M{"$in": qryFltr.OriginHosts, "$nin": qryFltr.NotOriginHosts},
		CDRSourceLow:       bson.M{"$in": qryFltr.Sources, "$nin": qryFltr.NotSources},
		RequestTypeLow:     bson.M{"$in": qryFltr.RequestTypes, "$nin": qryFltr.NotRequestTypes},
		DirectionLow:       bson.M{"$in": qryFltr.Directions, "$nin": qryFltr.NotDirections},
		TenantLow:          bson.M{"$in": qryFltr.Tenants, "$nin": qryFltr.NotTenants},
		CategoryLow:        bson.M{"$in": qryFltr.Categories, "$nin": qryFltr.NotCategories},
		AccountLow:         bson.M{"$in": qryFltr.Accounts, "$nin": qryFltr.NotAccounts},
		SubjectLow:         bson.M{"$in": qryFltr.Subjects, "$nin": qryFltr.NotSubjects},
		SupplierLow:        bson.M{"$in": qryFltr.Suppliers, "$nin": qryFltr.NotSuppliers},
		DisconnectCauseLow: bson.M{"$in": qryFltr.DisconnectCauses, "$nin": qryFltr.NotDisconnectCauses},
		SetupTimeLow:       bson.M{"$gte": qryFltr.SetupTimeStart, "$lt": qryFltr.SetupTimeEnd},
		AnswerTimeLow:      bson.M{"$gte": qryFltr.AnswerTimeStart, "$lt": qryFltr.AnswerTimeEnd},
		CreatedAtLow:       bson.M{"$gte": qryFltr.CreatedAtStart, "$lt": qryFltr.CreatedAtEnd},
		UpdatedAtLow:       bson.M{"$gte": qryFltr.UpdatedAtStart, "$lt": qryFltr.UpdatedAtEnd},
		UsageLow:           bson.M{"$gte": minUsage, "$lt": maxUsage},
		PDDLow:             bson.M{"$gte": minPDD, "$lt": maxPDD},
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
	session, col := ms.conn(utils.TBL_CDRS)
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
	return cdrs, 0, nil
}
