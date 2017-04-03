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
package engine

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"strings"

	"github.com/cgrates/cgrates/utils"
)

type CSVStorage struct {
	sep        rune
	readerFunc func(string, rune, int) (*csv.Reader, *os.File, error)
	// file names
	destinationsFn, ratesFn, destinationratesFn, timingsFn, destinationratetimingsFn, ratingprofilesFn,
	sharedgroupsFn, lcrFn, actionsFn, actiontimingsFn, actiontriggersFn, accountactionsFn, derivedChargersFn, cdrStatsFn, usersFn, aliasesFn, resLimitsFn string
}

func NewFileCSVStorage(sep rune,
	destinationsFn, timingsFn, ratesFn, destinationratesFn, destinationratetimingsFn, ratingprofilesFn, sharedgroupsFn, lcrFn,
	actionsFn, actiontimingsFn, actiontriggersFn, accountactionsFn, derivedChargersFn, cdrStatsFn, usersFn, aliasesFn, resLimitsFn string) *CSVStorage {
	c := new(CSVStorage)
	c.sep = sep
	c.readerFunc = openFileCSVStorage
	c.destinationsFn, c.timingsFn, c.ratesFn, c.destinationratesFn, c.destinationratetimingsFn, c.ratingprofilesFn,
		c.sharedgroupsFn, c.lcrFn, c.actionsFn, c.actiontimingsFn, c.actiontriggersFn, c.accountactionsFn, c.derivedChargersFn, c.cdrStatsFn, c.usersFn, c.aliasesFn, c.resLimitsFn = destinationsFn, timingsFn,
		ratesFn, destinationratesFn, destinationratetimingsFn, ratingprofilesFn, sharedgroupsFn, lcrFn, actionsFn, actiontimingsFn, actiontriggersFn, accountactionsFn, derivedChargersFn, cdrStatsFn, usersFn, aliasesFn, resLimitsFn
	return c
}

func NewStringCSVStorage(sep rune,
	destinationsFn, timingsFn, ratesFn, destinationratesFn, destinationratetimingsFn, ratingprofilesFn, sharedgroupsFn, lcrFn,
	actionsFn, actiontimingsFn, actiontriggersFn, accountactionsFn, derivedChargersFn, cdrStatsFn, usersFn, aliasesFn, resLimitsFn string) *CSVStorage {
	c := NewFileCSVStorage(sep, destinationsFn, timingsFn, ratesFn, destinationratesFn, destinationratetimingsFn,
		ratingprofilesFn, sharedgroupsFn, lcrFn, actionsFn, actiontimingsFn, actiontriggersFn, accountactionsFn, derivedChargersFn, cdrStatsFn, usersFn, aliasesFn, resLimitsFn)
	c.readerFunc = openStringCSVStorage
	return c
}

func openFileCSVStorage(fn string, comma rune, nrFields int) (csvReader *csv.Reader, fp *os.File, err error) {
	fp, err = os.Open(fn)
	if err != nil {
		return
	}
	csvReader = csv.NewReader(fp)
	csvReader.Comma = comma
	csvReader.Comment = utils.COMMENT_CHAR
	csvReader.FieldsPerRecord = nrFields
	csvReader.TrailingComma = true
	return
}

func openStringCSVStorage(data string, comma rune, nrFields int) (csvReader *csv.Reader, fp *os.File, err error) {
	csvReader = csv.NewReader(strings.NewReader(data))
	csvReader.Comma = comma
	csvReader.Comment = utils.COMMENT_CHAR
	csvReader.FieldsPerRecord = nrFields
	csvReader.TrailingComma = true
	return
}

func (csvs *CSVStorage) GetTPTimings(tpid, id string) ([]*utils.ApierTPTiming, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.timingsFn, csvs.sep, getColumnCount(TpTiming{}))
	if err != nil {
		//log.Print("Could not load timings file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpTimings TpTimings
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Print("bad line in timings csv: ", err)
			return nil, err
		}
		if tpTiming, err := csvLoad(TpTiming{}, record); err != nil {
			log.Print("error loading timing: ", err)
			return nil, err
		} else {
			tm := tpTiming.(TpTiming)
			tm.Tpid = tpid
			tpTimings = append(tpTimings, tm)
		}
	}
	return tpTimings.AsTPTimings(), nil
}

func (csvs *CSVStorage) GetTPDestinations(tpid, id string) ([]*utils.TPDestination, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.destinationsFn, csvs.sep, getColumnCount(TpDestination{}))
	if err != nil {
		//log.Print("Could not load destinations file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpDests TpDestinations
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Print("bad line in destinations csv: ", err)
			return nil, err
		}
		if tpDest, err := csvLoad(TpDestination{}, record); err != nil {
			log.Print("error loading destination: ", err)
			return nil, err
		} else {
			d := tpDest.(TpDestination)
			d.Tpid = tpid
			tpDests = append(tpDests, d)
		}
	}
	return tpDests.AsTPDestinations(), nil
}

func (csvs *CSVStorage) GetTPRates(tpid, id string) ([]*utils.TPRate, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.ratesFn, csvs.sep, getColumnCount(TpRate{}))
	if err != nil {
		//log.Print("Could not load rates file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpRates TpRates
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Print("bad line in rates csv: ", err)
			return nil, err
		}
		if tpRate, err := csvLoad(TpRate{}, record); err != nil {
			log.Print("error loading rate: ", err)
			return nil, err
		} else {
			r := tpRate.(TpRate)
			r.Tpid = tpid
			tpRates = append(tpRates, r)
		}
	}
	if rs, err := tpRates.AsTPRates(); err != nil {
		return nil, err
	} else {
		return rs, nil
	}
}

func (csvs *CSVStorage) GetTPDestinationRates(tpid, id string, p *utils.Paginator) ([]*utils.TPDestinationRate, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.destinationratesFn, csvs.sep, getColumnCount(TpDestinationRate{}))
	if err != nil {
		//log.Print("Could not load destination_rates file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpDestinationRates TpDestinationRates
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Print("bad line in destinationrates csv: ", err)
			return nil, err
		}
		if tpRate, err := csvLoad(TpDestinationRate{}, record); err != nil {
			log.Print("error loading destination rate: ", err)
			return nil, err
		} else {
			dr := tpRate.(TpDestinationRate)
			dr.Tpid = tpid
			tpDestinationRates = append(tpDestinationRates, dr)
		}
	}
	if drs, err := tpDestinationRates.AsTPDestinationRates(); err != nil {
		return nil, err
	} else {
		return drs, nil
	}
}

func (csvs *CSVStorage) GetTPRatingPlans(tpid, id string, p *utils.Paginator) ([]*utils.TPRatingPlan, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.destinationratetimingsFn, csvs.sep, getColumnCount(TpRatingPlan{}))
	if err != nil {
		//log.Print("Could not load rate plans file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpRatingPlans TpRatingPlans
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Print("bad line in rating plans csv: ", err)
			return nil, err
		}
		if tpRate, err := csvLoad(TpRatingPlan{}, record); err != nil {
			log.Print("error loading rating plan: ", err)
			return nil, err
		} else {
			rp := tpRate.(TpRatingPlan)
			rp.Tpid = tpid
			tpRatingPlans = append(tpRatingPlans, rp)
		}
	}
	if rps, err := tpRatingPlans.AsTPRatingPlans(); err != nil {
		return nil, err
	} else {
		return rps, nil
	}
}

func (csvs *CSVStorage) GetTPRatingProfiles(filter *utils.TPRatingProfile) ([]*utils.TPRatingProfile, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.ratingprofilesFn, csvs.sep, getColumnCount(TpRatingProfile{}))
	if err != nil {
		//log.Print("Could not load rating profiles file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpRatingProfiles TpRatingProfiles
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Print("bad line rating profiles csv: ", err)
			return nil, err
		}
		if tpRate, err := csvLoad(TpRatingProfile{}, record); err != nil {
			log.Print("error loading rating profile: ", err)
			return nil, err
		} else {
			rpf := tpRate.(TpRatingProfile)
			if filter != nil {
				rpf.Tpid = filter.TPid
				rpf.Loadid = filter.LoadId
			}
			tpRatingProfiles = append(tpRatingProfiles, rpf)
		}
	}
	if rps, err := tpRatingProfiles.AsTPRatingProfiles(); err != nil {
		return nil, err
	} else {
		return rps, nil
	}
}

func (csvs *CSVStorage) GetTPSharedGroups(tpid, id string) ([]*utils.TPSharedGroups, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.sharedgroupsFn, csvs.sep, getColumnCount(TpSharedGroup{}))
	if err != nil {
		//log.Print("Could not load shared groups file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpSharedGroups TpSharedGroups
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Print("bad line in shared groups csv: ", err)
			return nil, err
		}
		if tpRate, err := csvLoad(TpSharedGroup{}, record); err != nil {
			log.Print("error loading shared group: ", err)
			return nil, err
		} else {
			sg := tpRate.(TpSharedGroup)
			sg.Tpid = tpid
			tpSharedGroups = append(tpSharedGroups, sg)
		}
	}
	if sgs, err := tpSharedGroups.AsTPSharedGroups(); err != nil {
		return nil, err
	} else {
		return sgs, nil
	}
}

func (csvs *CSVStorage) GetTPLCRs(filter *utils.TPLcrRules) ([]*utils.TPLcrRules, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.lcrFn, csvs.sep, getColumnCount(TpLcrRule{}))
	if err != nil {
		//log.Print("Could not load LCR rules file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpLCRs TpLcrRules
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if tpRate, err := csvLoad(TpLcrRule{}, record); err != nil {
			if err != nil {
				log.Print("bad line in lcr rules csv: ", err)
				return nil, err
			}
			return nil, err
		} else {
			lcr := tpRate.(TpLcrRule)
			if filter != nil {
				lcr.Tpid = filter.TPid
			}
			tpLCRs = append(tpLCRs, lcr)
		}
	}
	if lrs, err := tpLCRs.AsTPLcrRules(); err != nil {
		return nil, err
	} else {
		return lrs, nil
	}
}

func (csvs *CSVStorage) GetTPActions(tpid, id string) ([]*utils.TPActions, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.actionsFn, csvs.sep, getColumnCount(TpAction{}))
	if err != nil {
		//log.Print("Could not load action file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpActions TpActions
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Print("bad line in actions csv: ", err)
			return nil, err
		}
		if tpAction, err := csvLoad(TpAction{}, record); err != nil {
			log.Print("error loading action: ", err)
			return nil, err
		} else {
			a := tpAction.(TpAction)
			a.Tpid = tpid
			tpActions = append(tpActions, a)
		}
	}
	if as, err := tpActions.AsTPActions(); err != nil {
		return nil, err
	} else {
		return as, nil
	}
}

func (csvs *CSVStorage) GetTPActionPlans(tpid, id string) ([]*utils.TPActionPlan, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.actiontimingsFn, csvs.sep, getColumnCount(TpActionPlan{}))
	if err != nil {
		//log.Print("Could not load action plans file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpActionPlans TpActionPlans
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if tpRate, err := csvLoad(TpActionPlan{}, record); err != nil {
			log.Print("error loading action plan: ", err)
			return nil, err
		} else {
			ap := tpRate.(TpActionPlan)
			ap.Tpid = tpid
			tpActionPlans = append(tpActionPlans, ap)
		}
	}
	if aps, err := tpActionPlans.AsTPActionPlans(); err != nil {
		return nil, err
	} else {
		return aps, nil
	}
}

func (csvs *CSVStorage) GetTPActionTriggers(tpid, id string) ([]*utils.TPActionTriggers, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.actiontriggersFn, csvs.sep, getColumnCount(TpActionTrigger{}))
	if err != nil {
		//log.Print("Could not load action triggers file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpActionTriggers TpActionTriggers
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Print("bad line in action triggers csv: ", err)
			return nil, err
		}
		if tpAt, err := csvLoad(TpActionTrigger{}, record); err != nil {
			log.Print("error loading action trigger: ", err)
			return nil, err
		} else {
			at := tpAt.(TpActionTrigger)
			at.Tpid = tpid
			tpActionTriggers = append(tpActionTriggers, at)
		}
	}
	if ats, err := tpActionTriggers.AsTPActionTriggers(); err != nil {
		return nil, err
	} else {
		return ats, nil
	}
}

func (csvs *CSVStorage) GetTPAccountActions(filter *utils.TPAccountActions) ([]*utils.TPAccountActions, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.accountactionsFn, csvs.sep, getColumnCount(TpAccountAction{}))
	if err != nil {
		//log.Print("Could not load account actions file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpAccountActions TpAccountActions
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Print("bad line in account actions csv: ", err)
			return nil, err
		}
		if tpAa, err := csvLoad(TpAccountAction{}, record); err != nil {
			log.Print("error loading account action: ", err)
			return nil, err
		} else {
			aa := tpAa.(TpAccountAction)
			if filter != nil {
				aa.Tpid = filter.TPid
				aa.Loadid = filter.LoadId
			}
			tpAccountActions = append(tpAccountActions, aa)
		}
	}
	if ats, err := tpAccountActions.AsTPAccountActions(); err != nil {
		return nil, err
	} else {
		return ats, nil
	}
}

func (csvs *CSVStorage) GetTPDerivedChargers(filter *utils.TPDerivedChargers) ([]*utils.TPDerivedChargers, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.derivedChargersFn, csvs.sep, getColumnCount(TpDerivedCharger{}))
	if err != nil {
		//log.Print("Could not load derivedChargers file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpDerivedChargers TpDerivedChargers
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Print("bad line in derived chargers csv: ", err)
			return nil, err
		}
		if tp, err := csvLoad(TpDerivedCharger{}, record); err != nil {
			log.Print("error loading derived charger: ", err)
			return nil, err
		} else {
			dc := tp.(TpDerivedCharger)
			if filter != nil {
				dc.Tpid = filter.TPid
				dc.Loadid = filter.LoadId
			}
			tpDerivedChargers = append(tpDerivedChargers, dc)
		}
	}
	if dcs, err := tpDerivedChargers.AsTPDerivedChargers(); err != nil {
		return nil, err
	} else {
		return dcs, nil
	}
}

func (csvs *CSVStorage) GetTPCdrStats(tpid, id string) ([]*utils.TPCdrStats, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.cdrStatsFn, csvs.sep, getColumnCount(TpCdrstat{}))
	if err != nil {
		//log.Print("Could not load cdr stats file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpCdrStats TpCdrStats
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Print("bad line in cdr stats csv: ", err)
			return nil, err
		}
		if tpCdrStat, err := csvLoad(TpCdrstat{}, record); err != nil {
			log.Print("error loading cdr stat: ", err)
			return nil, err
		} else {
			cs := tpCdrStat.(TpCdrstat)
			cs.Tpid = tpid
			tpCdrStats = append(tpCdrStats, cs)
		}
	}
	if css, err := tpCdrStats.AsTPCdrStats(); err != nil {
		return nil, err
	} else {
		return css, nil
	}
}

func (csvs *CSVStorage) GetTPUsers(filter *utils.TPUsers) ([]*utils.TPUsers, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.usersFn, csvs.sep, getColumnCount(TpUser{}))
	if err != nil {
		//log.Print("Could not load users file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpUsers TpUsers
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Print("bad line in users csv: ", err)
			return nil, err
		}
		if tpUser, err := csvLoad(TpUser{}, record); err != nil {
			log.Print("error loading user: ", err)
			return nil, err
		} else {
			u := tpUser.(TpUser)
			if filter != nil {
				u.Tpid = filter.TPid
			}
			tpUsers = append(tpUsers, u)
		}
	}
	if us, err := tpUsers.AsTPUsers(); err != nil {
		return nil, err
	} else {
		return us, nil
	}
}

func (csvs *CSVStorage) GetTPAliases(filter *utils.TPAliases) ([]*utils.TPAliases, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.aliasesFn, csvs.sep, getColumnCount(TpAlias{}))
	if err != nil {
		//log.Print("Could not load aliases file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpAliases TpAliases
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Print("bad line in aliases csv: ", err)
			return nil, err
		}
		if tpAlias, err := csvLoad(TpAlias{}, record); err != nil {
			log.Print("error loading alias: ", err)
			return nil, err
		} else {
			u := tpAlias.(TpAlias)
			if filter != nil {
				u.Tpid = filter.TPid
			}
			tpAliases = append(tpAliases, u)
		}
	}
	if as, err := tpAliases.AsTPAliases(); err != nil {
		return nil, err
	} else {
		return as, nil
	}
}

func (csvs *CSVStorage) GetTPResourceLimits(tpid, id string) ([]*utils.TPResourceLimit, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.resLimitsFn, csvs.sep, getColumnCount(TpResourceLimit{}))
	if err != nil {
		//log.Print("Could not load resource limits file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpResLimits TpResourceLimits
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Print("bad line in resourcelimits csv: ", err)
			return nil, err
		}
		if tpResLimit, err := csvLoad(TpResourceLimit{}, record); err != nil {
			log.Print("error loading resourcelimit: ", err)
			return nil, err
		} else {
			tpLimit := tpResLimit.(TpResourceLimit)
			tpLimit.Tpid = tpid
			tpResLimits = append(tpResLimits, &tpLimit)
		}
	}
	return tpResLimits.AsTPResourceLimits(), nil
}

func (csvs *CSVStorage) GetTpIds() ([]string, error) {
	return nil, utils.ErrNotImplemented
}

func (csvs *CSVStorage) GetTpTableIds(tpid, table string, distinct utils.TPDistinctIds, filters map[string]string, p *utils.Paginator) ([]string, error) {
	return nil, utils.ErrNotImplemented
}
