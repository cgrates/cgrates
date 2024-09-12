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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
	"gorm.io/gorm"
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

func TestMongoGetContext(t *testing.T) {
	testCtx := context.Background()
	ms := &MongoStorage{
		ctx: testCtx,
	}
	gotCtx := ms.GetContext()
	if gotCtx != testCtx {
		t.Errorf("GetContext() = %v; want %v", gotCtx, testCtx)
	}
}
func TestMongoSelectDatabase(t *testing.T) {
	initialDB := "mongo"
	ms := &MongoStorage{
		db: initialDB,
	}
	newDB := "db"
	if err := ms.SelectDatabase(newDB); err != nil {
		t.Errorf("SelectDatabase() returned an error: %v", err)
	}
	if got := ms.db; got != newDB {
		t.Errorf("SelectDatabase() updated db to %v, want %v", got, newDB)
	}
}

func TestMongoGetStorageType(t *testing.T) {
	ms := &MongoStorage{}
	storageType := ms.GetStorageType()
	expectedStorageType := utils.MetaMongo
	if storageType != expectedStorageType {
		t.Errorf("Expected storage type: %s, got: %s", expectedStorageType, storageType)
	}
}

func TestRemoveKeysForPrefix(t *testing.T) {
	sqlStorage := SQLStorage{}
	testPrefix := "1"
	err := sqlStorage.RemoveKeysForPrefix(testPrefix)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error: %v, got: %v", utils.ErrNotImplemented, err)
	}
}

func TestGetKeysForPrefix(t *testing.T) {
	sqlStorage := SQLStorage{}
	testPrefix := "1"
	keys, err := sqlStorage.GetKeysForPrefix(testPrefix)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error: %v, got: %v", utils.ErrNotImplemented, err)
	}
	if keys != nil {
		t.Errorf("Expected keys to be nil, got: %v", keys)
	}
}

func TestMysqlSelectDatabase(t *testing.T) {
	sqlStorage := SQLStorage{}
	testDBName := "mySql"
	err := sqlStorage.SelectDatabase(testDBName)
	if err != nil {
		t.Errorf("Expected nil error, got: %v", err)
	}
}

func TestExportGormDB(t *testing.T) {
	mockDB := &gorm.DB{}
	sqlStorage := &SQLStorage{
		db: mockDB,
	}
	resultDB := sqlStorage.ExportGormDB()
	if resultDB != mockDB {
		t.Errorf("ExportGormDB() = %v; want %v", resultDB, mockDB)
	}
}
