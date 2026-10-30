// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package apis

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func validateFilterRules(rules []*engine.FilterRule) error {
	for _, rule := range rules {
		if !rule.IsValid() {
			return fmt.Errorf("there exists at least one filter rule that is not valid")
		}
	}
	return nil
}

// SetFilter add a new Filter
func (adms *AdminSv1) SetFilter(ctx *context.Context, arg *engine.FilterWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(arg.Filter, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if arg.Tenant == utils.EmptyString {
		arg.Tenant = adms.cfg.GeneralCfg().DefaultTenant
	}
	unlock := adms.dm.Lock(utils.ConcatenatedKey(utils.AdminSv1SetFilter, arg.TenantID()))
	defer unlock()
	return adms.setFilter(ctx, arg, reply)
}

func (adms *AdminSv1) setFilter(ctx *context.Context, arg *engine.FilterWithAPIOpts, reply *string) error {
	if len(arg.Rules) == 0 {
		return utils.NewErrMandatoryIeMissing("Filter Rules")
	}
	if err := validateFilterRules(arg.Rules); err != nil {
		return utils.APIErrorHandler(err)
	}
	tntID := arg.TenantID()
	argC := map[string][]string{utils.CacheFilters: {tntID}}
	fltr, err := adms.dm.GetFilter(ctx, arg.Tenant, arg.ID, false, false, utils.NonTransactional)
	if err != nil {
		if err != utils.ErrNotFound {
			return utils.APIErrorHandler(err)
		}
	} else if argC, err = composeCacheArgsForFilter(adms.dm, ctx, fltr, fltr.Tenant, tntID, argC); err != nil {
		return utils.APIErrorHandler(err)
	}
	if err := adms.dm.SetFilter(ctx, arg.Filter, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	if argC, err = composeCacheArgsForFilter(adms.dm, ctx, arg.Filter, arg.Filter.Tenant, tntID, argC); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheFilters and store it in database
	if err := adms.dm.SetLoadIDs(ctx,
		map[string]int64{utils.CacheFilters: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for Filter
	connIDs, err := engine.GetConnIDs(ctx, adms.cfg.AdminSCfg().Conns, utils.MetaCaches, arg.Tenant, arg, nil, adms.fltrS)
	if err != nil {
		return utils.APIErrorHandler(err)
	}
	if err := callCacheForFilter(adms.connMgr, connIDs, ctx,
		utils.IfaceAsString(arg.APIOpts[utils.MetaCache]),
		adms.cfg.GeneralCfg().DefaultCaching,
		arg.Tenant, argC, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// SetFilterRule adds values to a filter rule, or replaces them when Overwrite is set.
// A rule is added if Type and Element do not match an existing rule.
// Multiple matches return an error.
func (adms *AdminSv1) SetFilterRule(ctx *context.Context, arg *engine.SetFilterRuleArgs, reply *string) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID, utils.Type, utils.Element}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if arg.Tenant == "" {
		arg.Tenant = adms.cfg.GeneralCfg().DefaultTenant
	}
	unlock := adms.dm.Lock(utils.ConcatenatedKey(utils.AdminSv1SetFilter, arg.Tenant, arg.ID))
	defer unlock()
	filter, rule, err := adms.getFilterRule(ctx, arg.Tenant, arg.ID, arg.Type, arg.Element)
	if err != nil {
		return err
	}
	if rule == nil {
		rule = &engine.FilterRule{
			Type:    arg.Type,
			Element: arg.Element,
		}
		filter.Rules = append(filter.Rules, rule)
	}
	if arg.Overwrite {
		rule.Values = arg.Values
	} else {
		present := utils.NewStringSet(rule.Values)
		for _, value := range arg.Values {
			if !present.Has(value) {
				rule.Values = append(rule.Values, value)
				present.Add(value)
			}
		}
	}
	return adms.setFilter(ctx, &engine.FilterWithAPIOpts{Filter: filter, APIOpts: arg.APIOpts}, reply)
}

// RemoveFilterRule removes values from a filter rule selected by Type and Element.
// It returns an error unless exactly one rule matches.
func (adms *AdminSv1) RemoveFilterRule(ctx *context.Context, arg *engine.RemoveFilterRuleArgs, reply *string) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID, utils.Type, utils.Element, utils.Values}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if arg.Tenant == "" {
		arg.Tenant = adms.cfg.GeneralCfg().DefaultTenant
	}
	unlock := adms.dm.Lock(utils.ConcatenatedKey(utils.AdminSv1SetFilter, arg.Tenant, arg.ID))
	defer unlock()
	filter, rule, err := adms.getFilterRule(ctx, arg.Tenant, arg.ID, arg.Type, arg.Element)
	if err != nil {
		return err
	}
	if rule == nil {
		return utils.ErrPrefixNotFound("FilterRule")
	}
	remove := utils.NewStringSet(arg.Values)
	rule.Values = slices.DeleteFunc(rule.Values, remove.Has)
	return adms.setFilter(ctx, &engine.FilterWithAPIOpts{Filter: filter, APIOpts: arg.APIOpts}, reply)
}

func (adms *AdminSv1) getFilterRule(ctx *context.Context, tenant, id, ruleType, element string) (*engine.Filter, *engine.FilterRule, error) {
	if strings.HasPrefix(id, utils.Meta) {
		return nil, nil, utils.APIErrorHandler(fmt.Errorf("cannot update values of an inline filter"))
	}
	filter, err := adms.dm.GetFilter(ctx, tenant, id, false, false, utils.NonTransactional)
	if err != nil {
		return nil, nil, utils.APIErrorHandler(err)
	}
	// InternalDB can return the stored filter, so copy it before changing values.
	filterCopy := *filter
	filterCopy.Rules = make([]*engine.FilterRule, len(filter.Rules))
	for i, rule := range filter.Rules {
		filterCopy.Rules[i] = &engine.FilterRule{
			Type:    rule.Type,
			Element: rule.Element,
			Values:  slices.Clone(rule.Values),
		}
	}
	filter = &filterCopy
	var selected *engine.FilterRule
	for _, rule := range filter.Rules {
		if rule.Type != ruleType || rule.Element != element {
			continue
		}
		if selected != nil {
			return nil, nil, utils.APIErrorHandler(fmt.Errorf("multiple filter rules match Type <%s> and Element <%s>", ruleType, element))
		}
		selected = rule
	}
	return filter, selected, nil
}

// GetFilter returns a Filter
func (adms *AdminSv1) GetFilter(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *engine.Filter) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	if fltr, err := adms.dm.GetFilter(ctx, tnt, arg.ID, true, true, utils.NonTransactional); err != nil {
		return utils.APIErrorHandler(err)
	} else {
		*reply = *fltr
	}
	return nil
}

// GetFilters returns a list of filters for a tenant
func (adms *AdminSv1) GetFilters(ctx *context.Context, args *utils.ArgsItemIDs, fltrs *[]*engine.Filter) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	var fltrIDs []string
	if err = adms.GetFilterIDs(ctx, args, &fltrIDs); err != nil {
		return
	}
	*fltrs = make([]*engine.Filter, 0, len(fltrIDs))
	for _, fltrID := range fltrIDs {
		var fltr *engine.Filter
		if fltr, err = adms.dm.GetFilter(ctx, tnt, fltrID, true, true, utils.NonTransactional); err != nil {
			return utils.APIErrorHandler(err)
		}
		*fltrs = append(*fltrs, fltr)
	}
	return
}

// GetFilterIDs returns list of Filter IDs registered for a tenant
func (adms *AdminSv1) GetFilterIDs(ctx *context.Context, args *utils.ArgsItemIDs, fltrIDs *[]string) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.FilterPrefix + tnt + utils.ConcatenatedKeySep
	lenPrfx := len(prfx)
	db, _, err := adms.dm.DBConns().GetConn(utils.MetaFilters)
	if err != nil {
		return err
	}
	var keys []string
	if keys, err = db.GetKeysForPrefix(ctx, prfx, args.ItemsSearch); err != nil {
		return
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	retIDs := make([]string, len(keys))
	for i, key := range keys {
		retIDs[i] = key[lenPrfx:]
	}
	var limit, offset, maxItems int
	if limit, offset, maxItems, err = utils.GetPaginateOpts(args.APIOpts); err != nil {
		return
	}
	*fltrIDs, err = utils.Paginate(retIDs, limit, offset, maxItems)
	return
}

// RemoveFilter  remove a specific filter
func (adms *AdminSv1) RemoveFilter(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	unlock := adms.dm.Lock(utils.ConcatenatedKey(utils.AdminSv1SetFilter, tnt, arg.ID))
	defer unlock()
	if err := adms.dm.RemoveFilter(ctx, tnt, arg.ID, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheFilters and store it in database
	if err := adms.dm.SetLoadIDs(ctx, map[string]int64{utils.CacheFilters: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for Filter
	connIDs, err := engine.GetConnIDs(ctx, adms.cfg.AdminSCfg().Conns, utils.MetaCaches, arg.Tenant, utils.MapStorage{}, nil, adms.fltrS)
	if err != nil {
		return utils.APIErrorHandler(err)
	}
	if err := callCacheForFilter(adms.connMgr, connIDs, ctx,
		utils.IfaceAsString(arg.APIOpts[utils.MetaCache]),
		adms.cfg.GeneralCfg().DefaultCaching,
		arg.Tenant, map[string][]string{utils.CacheFilters: {utils.ConcatenatedKey(tnt, arg.ID)}}, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// GetFiltersCount returns the total number of FilterIDs registered for a tenant
// returns ErrNotFound in case of 0 FilterIDs
func (admS *AdminSv1) GetFiltersCount(ctx *context.Context, args *utils.ArgsItemIDs, reply *int) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.FilterPrefix + tnt + utils.ConcatenatedKeySep
	db, _, err := admS.dm.DBConns().GetConn(utils.MetaFilters)
	if err != nil {
		return err
	}
	var keys []string
	if keys, err = db.GetKeysForPrefix(ctx, prfx, args.ItemsSearch); err != nil {
		return err
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	*reply = len(keys)
	return
}

// FiltersMatch checks whether a set of filter IDs passes for the provided CGREvent
func (admS *AdminSv1) FiltersMatch(ctx *context.Context, args *engine.ArgsFiltersMatch, reply *bool) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	evDP := args.CGREvent.AsDataProvider()
	var pass bool
	if pass, err = admS.fltrS.Pass(ctx, tnt, args.FilterIDs, evDP); err != nil {
		return
	} else if pass {
		*reply = true
	}
	return
}
