/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package engine

import (
	"slices"
	"strings"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// SetCDR for ManagerDB interface. SetCDR will set a single CDR in internal based on the CGREvent
func (iDB *InternalDB) SetCDR(_ *context.Context, cdr *utils.CGREvent, allowUpdate bool) error {
	cdrID := utils.IfaceAsString(cdr.APIOpts[utils.MetaCDRID])
	if !allowUpdate {
		if _, has := iDB.db.Get(utils.MetaCDRs, cdrID); has {
			return utils.ErrExists
		}
	}
	idx := make(utils.StringSet)
	dp := cdr.AsDataProvider()
	iDB.indexedFieldsMutex.RLock()
	for _, v := range iDB.stringIndexedFields {
		val, err := dp.FieldAsString(strings.Split(v, utils.NestingSep))
		if err != nil {
			if err == utils.ErrNotFound {
				continue
			}
			return err
		}
		idx.Add(utils.ConcatenatedKey(v, val))
	}
	for _, v := range iDB.prefixIndexedFields {
		val, err := dp.FieldAsString(strings.Split(v, utils.NestingSep))
		if err != nil {
			if err == utils.ErrNotFound {
				continue
			}
			return err
		}
		idx.Add(utils.ConcatenatedKey(v, val))
		for i := len(val) - 1; i > 0; i-- {
			idx.Add(utils.ConcatenatedKey(v, val[:i]))
		}
	}
	iDB.indexedFieldsMutex.RUnlock()

	iDB.db.Set(utils.MetaCDRs, cdrID, cdr, idx.AsSlice(), true, utils.NonTransactional)
	return nil
}

func (iDB *InternalDB) GetCDRs(ctx *context.Context, qryFltr []*Filter, opts map[string]any) (cdrs []*utils.CDR, err error) {
	pairFltrs := make(map[string][]string)
	notPairFltrs := make(map[string][]string)
	notIndexed := []*FilterRule{}
	for _, fltr := range qryFltr {
		for _, rule := range fltr.Rules {
			var elem string
			if !slices.Contains(iDB.stringIndexedFields, strings.TrimPrefix(rule.Element, "~")) ||
				rule.Type != utils.MetaString && rule.Type != utils.MetaNotString {
				notIndexed = append(notIndexed, rule)
				continue
			}
			elem = strings.Trim(rule.Element, "~")
			switch rule.Type {
			case utils.MetaString:
				pairFltrs[elem] = rule.Values
			case utils.MetaNotString:
				notPairFltrs[elem] = rule.Values
			}
		}
	}
	// find indexed fields
	var cdrMpIDs utils.StringSet
	// Apply string filter
	for keySlice, fltrSlice := range pairFltrs {
		if len(fltrSlice) == 0 {
			continue
		}
		grpMpIDs := make(utils.StringSet)
		for _, id := range fltrSlice {
			grpMpIDs.AddSlice(iDB.db.GetGroupItemIDs(utils.MetaCDRs, utils.ConcatenatedKey(keySlice, id)))
		}
		if grpMpIDs.Size() == 0 {
			return nil, utils.ErrNotFound
		}
		if cdrMpIDs == nil {
			cdrMpIDs = grpMpIDs
			continue
		}
		cdrMpIDs.Intersect(grpMpIDs)
		if cdrMpIDs.Size() == 0 {
			return nil, utils.ErrNotFound
		}
	}
	if cdrMpIDs == nil {
		cdrMpIDs = utils.NewStringSet(iDB.db.GetItemIDs(utils.MetaCDRs, utils.EmptyString))
	}
	// check for Not filters
	for keySlice, fltrSlice := range notPairFltrs {
		if len(fltrSlice) == 0 {
			continue
		}
		for _, id := range fltrSlice {
			for _, id := range iDB.db.GetGroupItemIDs(utils.MetaCDRs, utils.ConcatenatedKey(keySlice, id)) {
				cdrMpIDs.Remove(id)
				if cdrMpIDs.Size() == 0 {
					return nil, utils.ErrNotFound
				}
			}
		}
	}

	events := []*utils.CGREvent{}
	for key := range cdrMpIDs {
		x, ok := iDB.db.Get(utils.MetaCDRs, key)
		if !ok || x == nil {
			return nil, utils.ErrNotFound
		}
		cgrEv := x.(*utils.CGREvent)
		cgrEvDP := cgrEv.AsDataProvider()
		// checking pass for every filter that cannot be indexed
		var pass bool = true
		for _, fltr := range notIndexed {
			if pass, err = fltr.Pass(ctx, cgrEvDP); err != nil {
				return nil, err
			} else if !pass {
				break
			}
		}
		if !pass {
			continue
		}
		events = append(events, cgrEv)
	}
	if len(events) == 0 {
		return nil, utils.ErrNotFound
	}
	// convert from event into CDRs
	cdrs = make([]*utils.CDR, len(events))
	for i, event := range events {
		cdrs[i] = &utils.CDR{
			Tenant:    event.Tenant,
			Opts:      event.APIOpts,
			Event:     event.Event,
			CreatedAt: time.Now(),
		}
	}
	var limit, offset, maxItems int
	if limit, offset, maxItems, err = utils.GetPaginateOpts(opts); err != nil {
		return
	}
	cdrs, err = utils.Paginate(cdrs, limit, offset, maxItems)
	return
}

func (iDB *InternalDB) RemoveCDRs(ctx *context.Context, qryFltr []*Filter) (err error) {
	pairFltrs := make(map[string][]string)
	notPairFltrs := make(map[string][]string)
	notIndexed := []*FilterRule{}
	for _, fltr := range qryFltr {
		for _, rule := range fltr.Rules {
			var elem string
			if !slices.Contains(iDB.stringIndexedFields, strings.TrimPrefix(rule.Element, "~")) ||
				rule.Type != utils.MetaString && rule.Type != utils.MetaNotString {
				notIndexed = append(notIndexed, rule)
				continue
			}
			elem = strings.Trim(rule.Element, "~")
			switch rule.Type {
			case utils.MetaString:
				pairFltrs[elem] = rule.Values
			case utils.MetaNotString:
				notPairFltrs[elem] = rule.Values
			}
		}
	}
	// find indexed fields
	var cdrMpIDs utils.StringSet
	// Apply string filter
	for keySlice, fltrSlice := range pairFltrs {
		if len(fltrSlice) == 0 {
			continue
		}
		grpMpIDs := make(utils.StringSet)
		for _, id := range fltrSlice {
			grpMpIDs.AddSlice(iDB.db.GetGroupItemIDs(utils.MetaCDRs, utils.ConcatenatedKey(keySlice, id)))
		}
		if grpMpIDs.Size() == 0 {
			return utils.ErrNotFound
		}
		if cdrMpIDs == nil {
			cdrMpIDs = grpMpIDs
			continue
		}
		cdrMpIDs.Intersect(grpMpIDs)
		if cdrMpIDs.Size() == 0 {
			return utils.ErrNotFound
		}
	}
	if cdrMpIDs == nil {
		cdrMpIDs = utils.NewStringSet(iDB.db.GetItemIDs(utils.MetaCDRs, utils.EmptyString))
	}
	// check for Not filters
	for keySlice, fltrSlice := range notPairFltrs {
		if len(fltrSlice) == 0 {
			continue
		}
		for _, id := range fltrSlice {
			for _, id := range iDB.db.GetGroupItemIDs(utils.MetaCDRs, utils.ConcatenatedKey(keySlice, id)) {
				cdrMpIDs.Remove(id)
				if cdrMpIDs.Size() == 0 {
					return utils.ErrNotFound
				}
			}
		}
	}
	// iterrate trough all CDRs found and select only those who match our filters
	for key := range cdrMpIDs {
		x, ok := iDB.db.Get(utils.MetaCDRs, key)
		if !ok || x == nil {
			return utils.ErrNotFound
		}
		cgrEv := x.(*utils.CGREvent)
		cgrEvDP := cgrEv.AsDataProvider()
		// checking pass for every filter that cannot be indexed
		var pass bool = true
		for _, fltr := range notIndexed {
			if pass, err = fltr.Pass(ctx, cgrEvDP); err != nil {
				return err
			} else if !pass {
				// the CDR DID NOT passed, so we will remove it
				cdrMpIDs.Remove(key)
				break
			}
		}
		if !pass {
			continue
		}
	}
	// for every CDRs found, we delete matching by counter(key is a uniqueID)
	for key := range cdrMpIDs {
		iDB.db.Remove(utils.MetaCDRs, key, true, utils.NonTransactional)
	}
	return
}

// Will dump everything inside stordb to files
func (iDB *InternalDB) DumpStorDB() (err error) {
	return iDB.db.DumpAll()
}

// Will rewrite every dump file of StorDB
func (iDB *InternalDB) RewriteStorDB() (err error) {
	return iDB.db.RewriteAll()
}

// BackupStorDB will momentarely stop any dumping and rewriting until all dump folder is backed up in folder path backupFolderPath, making zip true will create a zip file in the path instead
func (iDB *InternalDB) BackupStorDB(backupFolderPath string, zip bool) (err error) {
	return iDB.db.BackupDumpFolder(backupFolderPath, zip)
}
