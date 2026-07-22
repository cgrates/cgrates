// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	keys, err := sqlStorage.GetKeysForPrefix(testPrefix, utils.EmptyString)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error: %v, got: %v", utils.ErrNotImplemented, err)
	}
	if keys != nil {
		t.Errorf("Expected keys to be nil, got: %v", keys)
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
