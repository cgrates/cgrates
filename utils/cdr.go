// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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

func GetUniqueURID(cgrEv *CGREvent) string {
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
