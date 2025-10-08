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

package utils

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type CDR struct {
	Tenant    string
	Opts      map[string]any
	Event     map[string]any
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `json:",omitempty"`
}

type CDRSQLTable struct {
	ID        int64 // this is used for incrementing while seting
	Tenant    string
	Opts      JSONB `gorm:"type:jsonb"` //string
	Event     JSONB `gorm:"type:jsonb"` //string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `json:",omitempty"`
}

func (CDRSQLTable) TableName() string {
	return CDRsTBL
}

// JSONB type for storing maps of events and opts into gorm columns as jsob type
type JSONB map[string]any

func (j JSONB) GormDataType() string {
	return "JSONB"
}

// Scan scan value into Jsonb, implements sql.Scanner interface
func (j *JSONB) Scan(value any) (err error) {
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, &j)
	case string:
		return json.Unmarshal([]byte(v), &j)
	default:
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}
}

// Value return json value, implement driver.Valuer interface
func (j JSONB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func GetUniqueCDRID(cgrEv *CGREvent) string {
	if chargeId, ok := cgrEv.APIOpts[MetaChargeID]; ok {
		return IfaceAsString(chargeId)
	}
	if originID, ok := cgrEv.APIOpts[MetaOriginID]; ok {
		return IfaceAsString(originID)
	}
	return UUIDSha1Prefix()
}

func (cdr *CDR) CGREvent() *CGREvent {
	return &CGREvent{
		Tenant:  cdr.Tenant,
		ID:      Sha1(),
		Event:   cdr.Event,
		APIOpts: cdr.Opts,
	}
}

// CDRsToCGREvents converts a slice of *CDR to a slice of *utils.CGREvent.
func CDRsToCGREvents(cdrs []*CDR) []*CGREvent {
	cgrEvs := make([]*CGREvent, 0, len(cdrs))
	for _, cdr := range cdrs {
		cgrEvs = append(cgrEvs, cdr.CGREvent())
	}
	return cgrEvs
}

type CDRFilters struct {
	Tenant    string
	ID        string
	FilterIDs []string
	APIOpts   map[string]any
}
