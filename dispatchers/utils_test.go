// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package dispatchers

import (
	"reflect"
	"strings"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestDParseStringSetMetaZero(t *testing.T) {
	stringTest := utils.MetaZero
	result := ParseStringSet(stringTest)
	expected := make(utils.StringSet)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDParseStringSet(t *testing.T) {
	stringTest := "testString"
	result := ParseStringSet(stringTest)
	expected := utils.NewStringSet(strings.Split(stringTest, utils.ANDSep))
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}
