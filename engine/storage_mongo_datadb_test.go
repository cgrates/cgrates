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
	"errors"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestMongoErrIsNotFound(t *testing.T) {

	t.Run("mongo.CommandError with Code 26", func(t *testing.T) {
		err := &mongo.CommandError{Code: 26, Message: "some message"}
		if !isNotFound(err) {
			t.Errorf("expected true, got false for mongo.CommandError with Code 26")
		}
	})

	t.Run("mongo.CommandError with 'ns not found' message", func(t *testing.T) {
		err := &mongo.CommandError{Code: 100, Message: "ns not found"}
		if !isNotFound(err) {
			t.Errorf("expected true, got false for mongo.CommandError with 'ns not found' message")
		}
	})

	t.Run("Non-mongo.CommandError but 'ns not found' in message", func(t *testing.T) {
		err := errors.New("some random error: ns not found")
		if !isNotFound(err) {
			t.Errorf("expected true, got false for error with 'ns not found' in message")
		}
	})

	t.Run("Unrelated error", func(t *testing.T) {
		err := errors.New("some other error")
		if isNotFound(err) {
			t.Errorf("expected false, got true for unrelated error")
		}
	})

}

func TestCleanEmptyFilters(t *testing.T) {
	ms := &MongoStorage{}

	tests := []struct {
		name     string
		input    bson.M
		expected bson.M
	}{
		{
			name: "Remove nil int64",
			input: bson.M{
				"field1": (*int64)(nil),
				"field2": int64(5),
			},
			expected: bson.M{
				"field2": int64(5),
			},
		},
		{
			name: "Remove nil float64",
			input: bson.M{
				"field1": (*float64)(nil),
				"field2": float64(3.14),
			},
			expected: bson.M{
				"field2": float64(3.14),
			},
		},

		{
			name: "Remove nil time.Duration",
			input: bson.M{
				"field1": (*time.Duration)(nil),
				"field2": time.Duration(5),
			},
			expected: bson.M{
				"field2": time.Duration(5),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms.cleanEmptyFilters(tt.input)
			if len(tt.input) != len(tt.expected) {
				t.Errorf("Expected %d keys, got %d", len(tt.expected), len(tt.input))
			}
			for k, v := range tt.expected {
				if inputValue, exists := tt.input[k]; !exists || inputValue != v {
					t.Errorf("Key %s: expected %v, got %v", k, v, inputValue)
				}
			}
		})
	}
}

func TestMongoStoreDBGetStorageType(t *testing.T) {
	ms := &MongoStorage{}

	result := ms.GetStorageType()

	expected := utils.MetaMongo

	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}
