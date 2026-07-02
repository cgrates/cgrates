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

package utils_test

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

type nestedItem struct {
	Path  string
	Value string
}

type sliceArg struct {
	ID    string
	Items []*nestedItem
}

type recursiveArg struct {
	Name     string
	Children []*recursiveArg
}

func TestDescribeType(t *testing.T) {
	tests := []struct {
		name string
		typ  reflect.Type
		want []utils.FieldDescriptor
	}{
		{
			name: "embedded pointer struct is hoisted",
			typ:  reflect.TypeFor[utils.TenantIDWithAPIOpts](),
			want: []utils.FieldDescriptor{
				{Name: "Tenant", Type: "string"},
				{Name: "ID", Type: "string"},
				{Name: "APIOpts", Type: "map[string]any"},
			},
		},
		{
			name: "slice of struct exposes element fields and qualified name",
			typ:  reflect.TypeFor[sliceArg](),
			want: []utils.FieldDescriptor{
				{Name: "ID", Type: "string"},
				{Name: "Items", Type: "[]utils_test.nestedItem", Fields: []utils.FieldDescriptor{
					{Name: "Path", Type: "string"},
					{Name: "Value", Type: "string"},
				}},
			},
		},
		{
			name: "recursive type keeps its name but stops expanding at the cycle",
			typ:  reflect.TypeFor[recursiveArg](),
			want: []utils.FieldDescriptor{
				{Name: "Name", Type: "string"},
				{Name: "Children", Type: "[]utils_test.recursiveArg"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.DescribeType(tt.typ)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DescribeType(%s) = %+v, want %+v", tt.typ, got, tt.want)
			}
		})
	}
}
