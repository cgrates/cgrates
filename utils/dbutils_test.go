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

import "testing"

func TestFilterToSQLQuery(t *testing.T) {
	tests := []struct {
		name      string
		ruleType  string
		beforeSep string
		afterSep  string
		values    []string
		not       bool
		expected  []string
	}{
		{"MetaGreaterThan with values", MetaGreaterThan, "", "answer_time", []string{"NOW() - INTERVAL 7 DAY"}, false, []string{"answer_time > NOW() - INTERVAL 7 DAY"}},
		{"MetaEqual with values", MetaEqual, "cost_details", "Charges[0].RatingID", []string{"RatingID2"}, false, []string{"JSON_VALUE(cost_details, '$.Charges[0].RatingID') = 'RatingID2'"}},

		{"MetaExists with no values", MetaExists, "", "answer_time", nil, false, []string{"answer_time IS NULL"}},
		{"MetaExists with JSON field", MetaExists, "cost_details", "Charges[0].RatingID", nil, false, []string{"JSON_VALUE(cost_details, '$.Charges[0].RatingID') IS NULL"}},
		{"MetaNotExists with no values", MetaNotExists, "", "answer_time", nil, true, []string{"answer_time IS NOT NULL"}},
		{"MetaNotExists with JSON field", MetaNotExists, "cost_details", "Charges[0].RatingID", nil, true, []string{"JSON_VALUE(cost_details, '$.Charges[0].RatingID') IS NOT NULL"}},

		{"MetaString with values", MetaString, "", "answer_time", []string{"value1", "value2"}, false, []string{"answer_time = 'value1'", "answer_time = 'value2'"}},
		{"MetaNotString with values", MetaNotString, "cost_details", "Charges[0].RatingID", []string{"value1"}, true, []string{"JSON_VALUE(cost_details, '$.Charges[0].RatingID') != 'value1'"}},

		{"MetaEmpty with no values", MetaEmpty, "", "answer_time", nil, false, []string{"answer_time == ''"}},
		{"MetaEmpty with JSON field", MetaEmpty, "cost_details", "Charges[0].RatingID", nil, false, []string{"JSON_VALUE(cost_details, '$.Charges[0].RatingID') == ''"}},
		{"MetaNotEmpty with no values", MetaNotEmpty, "", "answer_time", nil, true, []string{"answer_time != ''"}},
		{"MetaNotEmpty with JSON field", MetaNotEmpty, "cost_details", "Charges[0].RatingID", nil, true, []string{"JSON_VALUE(cost_details, '$.Charges[0].RatingID') != ''"}},

		{"MetaGreaterOrEqual with values", MetaGreaterOrEqual, "", "answer_time", []string{"10"}, false, []string{"answer_time >= 10"}},
		{"MetaGreaterThan with values", MetaGreaterThan, "cost_details", "Charges[0].RatingID", []string{"20"}, false, []string{"JSON_VALUE(cost_details, '$.Charges[0].RatingID') > 20"}},

		{"MetaLessThan with values", MetaLessThan, "", "answer_time", []string{"5"}, false, []string{"answer_time < 5"}},
		{"MetaLessOrEqual with values", MetaLessOrEqual, "cost_details", "Charges[0].RatingID", []string{"15"}, false, []string{"JSON_VALUE(cost_details, '$.Charges[0].RatingID') <= 15"}},

		{"MetaPrefix with values", MetaPrefix, "", "answer_time", []string{"pre"}, false, []string{"answer_time LIKE 'pre%'"}},
		{"MetaNotPrefix with values", MetaNotPrefix, "cost_details", "Charges[0].RatingID", []string{"pre"}, true, []string{"JSON_VALUE(cost_details, '$.Charges[0].RatingID') NOT LIKE 'pre%'"}},

		{"MetaSuffix with values", MetaSuffix, "", "answer_time", []string{"suf"}, false, []string{"answer_time LIKE '%suf'"}},
		{"MetaNotSuffix with values", MetaNotSuffix, "cost_details", "Charges[0].RatingID", []string{"suf"}, true, []string{"JSON_VALUE(cost_details, '$.Charges[0].RatingID') NOT LIKE '%suf'"}},

		{"MetaGreaterOrEqual with JSON field", MetaGreaterOrEqual, "cost_details", "Charges[0].RatingID", []string{"100"}, false, []string{"JSON_VALUE(cost_details, '$.Charges[0].RatingID') >= 100"}},

		{"MetaRegex with values", MetaRegex, "", "answer_time", []string{"pattern1", "pattern2"}, false, []string{"answer_time REGEXP 'pattern1'", "answer_time REGEXP 'pattern2'"}},
		{"MetaNotRegex with values", MetaNotRegex, "cost_details", "Charges[0].RatingID", []string{"pattern"}, true, []string{"JSON_VALUE(cost_details, '$.Charges[0].RatingID') NOT REGEXP 'pattern'"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterToSQLQuery(tt.ruleType, tt.beforeSep, tt.afterSep, tt.values, tt.not)
			if len(got) != len(tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, got)
				return
			}
			for i, cond := range got {
				if cond != tt.expected[i] {
					t.Errorf("expected %v, got %v", tt.expected[i], cond)
				}
			}
		})
	}
}

func TestFilterToSQLQueryValidations(t *testing.T) {
	tests := []struct {
		name      string
		ruleType  string
		beforeSep string
		afterSep  string
		values    []string
		not       bool
		expected  []string
	}{
		{
			name:      "Boolean true value",
			ruleType:  MetaString,
			beforeSep: EmptyString,
			afterSep:  "active",
			values:    []string{"true"},
			not:       false,
			expected:  []string{"active = '1'"},
		},
		{
			name:      "Boolean false value",
			ruleType:  MetaString,
			beforeSep: EmptyString,
			afterSep:  "active",
			values:    []string{"false"},
			not:       false,
			expected:  []string{"active = '0'"},
		},
		{
			name:      "Greater than or equal with empty beforeSep",
			ruleType:  MetaGreaterOrEqual,
			beforeSep: EmptyString,
			afterSep:  "score",
			values:    []string{"10"},
			not:       false,
			expected:  []string{"score >= 10"},
		},
		{
			name:      "Greater than with empty beforeSep",
			ruleType:  MetaGreaterThan,
			beforeSep: EmptyString,
			afterSep:  "score",
			values:    []string{"20"},
			not:       false,
			expected:  []string{"score > 20"},
		},
		{
			name:      "Less than or equal with empty beforeSep",
			ruleType:  MetaLessOrEqual,
			beforeSep: EmptyString,
			afterSep:  "score",
			values:    []string{"30"},
			not:       false,
			expected:  []string{"score <= 30"},
		},
		{
			name:      "Less than with empty beforeSep",
			ruleType:  MetaLessThan,
			beforeSep: EmptyString,
			afterSep:  "score",
			values:    []string{"40"},
			not:       false,
			expected:  []string{"score < 40"},
		},
		{
			name:      "Prefix NOT LIKE with empty beforeSep",
			ruleType:  MetaNotPrefix,
			beforeSep: EmptyString,
			afterSep:  "name",
			values:    []string{"prefix"},
			not:       true,
			expected:  []string{"name NOT LIKE 'prefix%'"},
		},
		{
			name:      "Prefix LIKE with JSON_VALUE",
			ruleType:  MetaPrefix,
			beforeSep: "data",
			afterSep:  "name",
			values:    []string{"prefix"},
			not:       false,
			expected:  []string{"JSON_VALUE(data, '$.name') LIKE 'prefix%'"},
		},
		{
			name:      "Suffix NOT LIKE with empty beforeSep",
			ruleType:  MetaNotSuffix,
			beforeSep: EmptyString,
			afterSep:  "name",
			values:    []string{"suffix"},
			not:       true,
			expected:  []string{"name NOT LIKE '%suffix'"},
		},
		{
			name:      "Suffix LIKE with JSON_VALUE",
			ruleType:  MetaSuffix,
			beforeSep: "data",
			afterSep:  "name",
			values:    []string{"suffix"},
			not:       false,
			expected:  []string{"JSON_VALUE(data, '$.name') LIKE '%suffix'"},
		},
		{
			name:      "Regex NOT REGEXP with empty beforeSep",
			ruleType:  MetaNotRegex,
			beforeSep: EmptyString,
			afterSep:  "pattern",
			values:    []string{"[a-z]+"},
			not:       true,
			expected:  []string{"pattern NOT REGEXP '[a-z]+'"},
		},
		{
			name:      "Regex REGEXP with JSON_VALUE",
			ruleType:  MetaRegex,
			beforeSep: "data",
			afterSep:  "pattern",
			values:    []string{"[0-9]+"},
			not:       false,
			expected:  []string{"JSON_VALUE(data, '$.pattern') REGEXP '[0-9]+'"},
		},

		{
			name:      "Not equal with empty beforeSep",
			ruleType:  MetaString,
			beforeSep: EmptyString,
			afterSep:  "status",
			values:    []string{"inactive"},
			not:       true,
			expected:  []string{"status != 'inactive'"},
		},
		{
			name:      "Equal condition with JSON_VALUE",
			ruleType:  MetaString,
			beforeSep: "data",
			afterSep:  "status",
			values:    []string{"active"},
			not:       false,
			expected:  []string{"JSON_VALUE(data, '$.status') = 'active'"},
		},
		{
			name:      "Greater than condition with JSON_VALUE",
			ruleType:  MetaGreaterThan,
			beforeSep: "data",
			afterSep:  "score",
			values:    []string{"50"},
			not:       false,
			expected:  []string{"JSON_VALUE(data, '$.score') > 50"},
		},
		{
			name:      "Less than or equal condition with JSON_VALUE",
			ruleType:  MetaLessOrEqual,
			beforeSep: "data",
			afterSep:  "score",
			values:    []string{"30"},
			not:       false,
			expected:  []string{"JSON_VALUE(data, '$.score') <= 30"},
		},
		{
			name:      "Less than condition with JSON_VALUE",
			ruleType:  MetaLessThan,
			beforeSep: "data",
			afterSep:  "score",
			values:    []string{"20"},
			not:       false,
			expected:  []string{"JSON_VALUE(data, '$.score') < 20"},
		},

		{
			name:      "MetaExists with no values",
			ruleType:  MetaExists,
			beforeSep: "",
			afterSep:  "column1",
			values:    nil,
			not:       false,
			expected:  []string{"column1 IS NULL"},
		},
		{
			name:      "MetaNotExists with no values",
			ruleType:  MetaNotExists,
			beforeSep: "json_field",
			afterSep:  "key",
			values:    nil,
			not:       true,
			expected:  []string{"JSON_VALUE(json_field, '$.key') IS NOT NULL"},
		},
		{
			name:      "MetaString with values",
			ruleType:  MetaString,
			beforeSep: "",
			afterSep:  "column2",
			values:    []string{"value1", "value2"},
			not:       false,
			expected:  []string{"column2 = 'value1'", "column2 = 'value2'"},
		},
		{
			name:      "MetaPrefix with NOT condition",
			ruleType:  MetaNotPrefix,
			beforeSep: "json_field",
			afterSep:  "key",
			values:    []string{"prefix1"},
			not:       true,
			expected:  []string{"JSON_VALUE(json_field, '$.key') NOT LIKE 'prefix1%'"},
		},
		{
			name:      "MetaRegex with multiple values",
			ruleType:  MetaRegex,
			beforeSep: "",
			afterSep:  "column3",
			values:    []string{"pattern1", "pattern2"},
			not:       false,
			expected:  []string{"column3 REGEXP 'pattern1'", "column3 REGEXP 'pattern2'"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterToSQLQuery(tt.ruleType, tt.beforeSep, tt.afterSep, tt.values, tt.not)
			if len(got) != len(tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, got)
				return
			}
			for i, cond := range got {
				if cond != tt.expected[i] {
					t.Errorf("expected %v, got %v", tt.expected[i], cond)
				}
			}
		})
	}
}
