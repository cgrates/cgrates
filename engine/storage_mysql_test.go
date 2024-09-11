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
	"fmt"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestGetStorageTypes(t *testing.T) {
	msqlStorage := &MySQLStorage{}
	result := msqlStorage.GetStorageType()
	expected := utils.MetaMySQL
	if result != expected {
		t.Errorf("GetStorageType() = %s; want %s", result, expected)
	}
}

func TestNotExtraFieldsValueQry(t *testing.T) {
	msqlStorage := &MySQLStorage{}
	field := "Tenant"
	value := "cgrates.org"
	result := msqlStorage.notExtraFieldsValueQry(field, value)
	expected := fmt.Sprintf(" extra_fields NOT LIKE '%%\"%s\":\"%s\"%%'", field, value)
	if result != expected {
		t.Errorf("notExtraFieldsValueQry() = %s; want %s", result, expected)
	}
	field = "fieldWith\"SpecialChars"
	value = "valueWith'SpecialChars"
	result = msqlStorage.notExtraFieldsValueQry(field, value)
	expected = fmt.Sprintf(" extra_fields NOT LIKE '%%\"%s\":\"%s\"%%'", field, value)
	if result != expected {
		t.Errorf("notExtraFieldsValueQry() with special chars = %s; want %s", result, expected)
	}
}

func TestNotExtraFieldsExistsQry(t *testing.T) {
	msqlStorage := &MySQLStorage{}
	field := "Tenant"
	result := msqlStorage.notExtraFieldsExistsQry(field)
	expected := fmt.Sprintf(" extra_fields NOT LIKE '%%\"%s\":%%'", field)
	if result != expected {
		t.Errorf("notExtraFieldsExistsQry() = %s; want %s", result, expected)
	}
	field = "fieldWith\"SpecialChars"
	result = msqlStorage.notExtraFieldsExistsQry(field)
	expected = fmt.Sprintf(" extra_fields NOT LIKE '%%\"%s\":%%'", field)
	if result != expected {
		t.Errorf("notExtraFieldsExistsQry() with special chars = %s; want %s", result, expected)
	}
}

func TestExtraFieldsValueQry(t *testing.T) {
	msqlStorage := &MySQLStorage{}
	field := "Tenant"
	value := "cgrates.org"
	result := msqlStorage.extraFieldsValueQry(field, value)
	expected := fmt.Sprintf(" extra_fields LIKE '%%\"%s\":\"%s\"%%'", field, value)
	if result != expected {
		t.Errorf("extraFieldsValueQry() = %s; want %s", result, expected)
	}
	field = "fieldWith\"SpecialChars"
	value = "valueWith'SpecialChars"
	result = msqlStorage.extraFieldsValueQry(field, value)
	expected = fmt.Sprintf(" extra_fields LIKE '%%\"%s\":\"%s\"%%'", field, value)
	if result != expected {
		t.Errorf("extraFieldsValueQry() with special chars = %s; want %s", result, expected)
	}
}

func TestExtraFieldsExistsQry(t *testing.T) {
	msqlStorage := &MySQLStorage{}
	field := "Tenant"
	result := msqlStorage.extraFieldsExistsQry(field)
	expected := fmt.Sprintf(" extra_fields LIKE '%%\"%s\":%%'", field)
	if result != expected {
		t.Errorf("extraFieldsExistsQry() = %s; want %s", result, expected)
	}
	field = "fieldWith\"SpecialChars"
	result = msqlStorage.extraFieldsExistsQry(field)
	expected = fmt.Sprintf(" extra_fields LIKE '%%\"%s\":%%'", field)
	if result != expected {
		t.Errorf("extraFieldsExistsQry() with special chars = %s; want %s", result, expected)
	}
}
func TestAppendToMysqlDSNOptsBasic(t *testing.T) {
	opts := map[string]string{
		"user": "root",
	}
	result := AppendToMysqlDSNOpts(opts)
	expected := "&user=root"
	if result != expected {
		t.Errorf("AppendToMysqlDSNOpts() = %s; want %s", result, expected)
	}
	result = AppendToMysqlDSNOpts(nil)
	if result != utils.EmptyString {
		t.Errorf("AppendToMysqlDSNOpts(nil) = %s; want %s", result, utils.EmptyString)
	}
}
