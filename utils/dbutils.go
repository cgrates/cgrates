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

package utils

import "fmt"

// Creates mysql conditions used in WHERE statement out of filters
func FilterToSQLQuery(ruleType, beforeSep, afterSep string, values []string, not bool) (conditions []string) {
	// here are for the filters that their values are empty: *exists, *notexists, *empty, *notempty..
	if len(values) == 0 {
		switch ruleType {
		case MetaExists, MetaNotExists:
			if not {
				if beforeSep == EmptyString {
					conditions = append(conditions, fmt.Sprintf("%s IS NOT NULL", afterSep))
					return
				}
				conditions = append(conditions, fmt.Sprintf("JSON_VALUE(%s, '$.%s') IS NOT NULL", beforeSep, afterSep))
				return
			}
			if beforeSep == EmptyString {
				conditions = append(conditions, fmt.Sprintf("%s IS NULL", afterSep))
				return
			}
			conditions = append(conditions, fmt.Sprintf("JSON_VALUE(%s, '$.%s') IS NULL", beforeSep, afterSep))
		case MetaEmpty, MetaNotEmpty:
			if not {
				if beforeSep == EmptyString {
					conditions = append(conditions, fmt.Sprintf("%s != ''", afterSep))
					return
				}
				conditions = append(conditions, fmt.Sprintf("JSON_VALUE(%s, '$.%s') != ''", beforeSep, afterSep))
				return
			}
			if beforeSep == EmptyString {
				conditions = append(conditions, fmt.Sprintf("%s == ''", afterSep))
				return
			}
			conditions = append(conditions, fmt.Sprintf("JSON_VALUE(%s, '$.%s') == ''", beforeSep, afterSep))
		}
		return
	}
	// here are for the filters that can have more than one value: *string, *prefix, *suffix ..
	for _, value := range values {
		switch value { // in case we have boolean values, it should be queried over 1 or 0
		case "true":
			value = "1"
		case "false":
			value = "0"
		}
		var singleCond string
		switch ruleType {
		case MetaString, MetaNotString, MetaEqual, MetaNotEqual:
			if not {
				if beforeSep == EmptyString {
					conditions = append(conditions, fmt.Sprintf("%s != '%s'", afterSep, value))
					continue
				}
				conditions = append(conditions, fmt.Sprintf("JSON_VALUE(%s, '$.%s') != '%s'",
					beforeSep, afterSep, value))
				continue
			}
			if beforeSep == EmptyString {
				singleCond = fmt.Sprintf("%s = '%s'", afterSep, value)
			} else {
				singleCond = fmt.Sprintf("JSON_VALUE(%s, '$.%s') = '%s'", beforeSep, afterSep, value)
			}
		case MetaLessThan, MetaLessOrEqual, MetaGreaterThan, MetaGreaterOrEqual:
			if ruleType == MetaGreaterOrEqual {
				if beforeSep == EmptyString {
					singleCond = fmt.Sprintf("%s >= %s", afterSep, value)
				} else {
					singleCond = fmt.Sprintf("JSON_VALUE(%s, '$.%s') >= %s", beforeSep, afterSep, value)
				}
			} else if ruleType == MetaGreaterThan {
				if beforeSep == EmptyString {
					singleCond = fmt.Sprintf("%s > %s", afterSep, value)
				} else {
					singleCond = fmt.Sprintf("JSON_VALUE(%s, '$.%s') > %s", beforeSep, afterSep, value)
				}
			} else if ruleType == MetaLessOrEqual {
				if beforeSep == EmptyString {
					singleCond = fmt.Sprintf("%s <= %s", afterSep, value)
				} else {
					singleCond = fmt.Sprintf("JSON_VALUE(%s, '$.%s') <= %s", beforeSep, afterSep, value)
				}
			} else if ruleType == MetaLessThan {
				if beforeSep == EmptyString {
					singleCond = fmt.Sprintf("%s < %s", afterSep, value)
				} else {
					singleCond = fmt.Sprintf("JSON_VALUE(%s, '$.%s') < %s", beforeSep, afterSep, value)
				}
			}
		case MetaPrefix, MetaNotPrefix:
			if not {
				if beforeSep == EmptyString {
					conditions = append(conditions, fmt.Sprintf("%s NOT LIKE '%s%%'", afterSep, value))
					continue
				}
				conditions = append(conditions, fmt.Sprintf("JSON_VALUE(%s, '$.%s') NOT LIKE '%s%%'", beforeSep, afterSep, value))
				continue
			}
			if beforeSep == EmptyString {
				singleCond = fmt.Sprintf("%s LIKE '%s%%'", afterSep, value)
			} else {
				singleCond = fmt.Sprintf("JSON_VALUE(%s, '$.%s') LIKE '%s%%'", beforeSep, afterSep, value)
			}
		case MetaSuffix, MetaNotSuffix:
			if not {
				if beforeSep == EmptyString {
					conditions = append(conditions, fmt.Sprintf("%s NOT LIKE '%%%s'", afterSep, value))
					continue
				}
				conditions = append(conditions, fmt.Sprintf("JSON_VALUE(%s, '$.%s') NOT LIKE '%%%s'", beforeSep, afterSep, value))
				continue
			}
			if beforeSep == EmptyString {
				singleCond = fmt.Sprintf("%s LIKE '%%%s'", afterSep, value)
			} else {
				singleCond = fmt.Sprintf("JSON_VALUE(%s, '$.%s') LIKE '%%%s'", beforeSep, afterSep, value)
			}
		case MetaRegex, MetaNotRegex:
			if not {
				if beforeSep == EmptyString {
					conditions = append(conditions, fmt.Sprintf("%s NOT REGEXP '%s'", afterSep, value))
					continue
				}
				conditions = append(conditions, fmt.Sprintf("JSON_VALUE(%s, '$.%s') NOT REGEXP '%s'", beforeSep, afterSep, value))
				continue
			}
			if beforeSep == EmptyString {
				singleCond = fmt.Sprintf("%s REGEXP '%s'", afterSep, value)
			} else {
				singleCond = fmt.Sprintf("JSON_VALUE(%s, '$.%s') REGEXP '%s'", beforeSep, afterSep, value)
			}
		}
		conditions = append(conditions, singleCond)
	}
	return
}
