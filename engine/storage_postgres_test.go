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

func TestPostgressGetStorageType(t *testing.T) {
	poS := &PostgresStorage{}
	storageType := poS.GetStorageType()
	if storageType != utils.MetaPostgres {
		t.Errorf("expected %s, got %s", utils.MetaPostgres, storageType)
	}
}

func TestExtraFieldsQueries(t *testing.T) {
	poS := &PostgresStorage{}
	field := "Subject"
	value := "1001"
	expectedExistsQuery := fmt.Sprintf(" extra_fields ?'%s'", field)
	expectedValueQuery := fmt.Sprintf(" (extra_fields ->> '%s') = '%s'", field, value)
	existsQuery := poS.extraFieldsExistsQry(field)
	valueQuery := poS.extraFieldsValueQry(field, value)
	if existsQuery != expectedExistsQuery {
		t.Errorf("extraFieldsExistsQry: expected query to be %s, but got %s", expectedExistsQuery, existsQuery)
	}
	if valueQuery != expectedValueQuery {
		t.Errorf("extraFieldsValueQry: expected query to be %s, but got %s", expectedValueQuery, valueQuery)
	}
}

func TestPostgresNotExtraFieldsValueQry(t *testing.T) {
	poS := &PostgresStorage{}
	field := "Tor"
	value := "voice"
	expectedQuery := fmt.Sprintf(" NOT (extra_fields ?'%s' AND (extra_fields ->> '%s') = '%s')", field, field, value)
	query := poS.notExtraFieldsValueQry(field, value)
	if query != expectedQuery {
		t.Errorf("expected query to be %s, but got %s", expectedQuery, query)
	}
}

func TestPostgresNotExtraFieldsExistsQry(t *testing.T) {
	poS := &PostgresStorage{}
	field := "tor"
	expectedQuery := fmt.Sprintf(" NOT extra_fields ?'%s'", field)
	query := poS.notExtraFieldsExistsQry(field)
	if query != expectedQuery {
		t.Errorf("expected query to be %s, but got %s", expectedQuery, query)
	}
}
