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
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cgrates/cgrates/utils"
)

type CDR struct {
	Tenant    string
	Opts      map[string]interface{}
	Event     map[string]interface{}
	CreatedAt time.Time  `json:",omitempty"`
	UpdatedAt time.Time  `json:",omitempty"`
	DeletedAt *time.Time `json:",omitempty"`
}

type CDRSQLTable struct {
	ID        int64 // this is used for incrementing while seting
	Tenant    string
	Opts      JSON       `gorm:"type:jsonb"` //string
	Event     JSON       `gorm:"type:jsonb"` //string
	CreatedAt time.Time  `json:",omitempty"`
	UpdatedAt time.Time  `json:",omitempty"`
	DeletedAt *time.Time `json:",omitempty"`
}

func (CDRSQLTable) TableName() string {
	return utils.CDRsTBL
}

// JSON type for storing maps of events and opts into gorm columns as jsob type
type JSON map[string]interface{}

func (j JSON) GormDataType() string {
	return "JSONB"
}

// Scan scan value into Jsonb, implements sql.Scanner interface
func (j *JSON) Scan(value interface{}) (err error) {
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
func (j JSON) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func GetUniqueCDRID(cgrEv *utils.CGREvent) string {
	if chargeId, ok := cgrEv.APIOpts[utils.MetaChargeID]; ok {
		return utils.IfaceAsString(chargeId)
	} else if originID, ok := cgrEv.APIOpts[utils.MetaOriginID]; ok {
		return utils.IfaceAsString(originID)
	}
	return utils.UUIDSha1Prefix()
}

func (cdr *CDR) CGREvent() *utils.CGREvent {
	return &utils.CGREvent{
		Tenant:  cdr.Tenant,
		ID:      utils.Sha1(),
		Event:   cdr.Event,
		APIOpts: cdr.Opts,
	}
}

// CDRsToCGREvents converts a slice of *CDR to a slice of *utils.CGREvent.
func CDRsToCGREvents(cdrs []*CDR) []*utils.CGREvent {
	cgrEvs := make([]*utils.CGREvent, 0, len(cdrs))
	for _, cdr := range cdrs {
		cgrEvs = append(cgrEvs, cdr.CGREvent())
	}
	return cgrEvs
}

// checkNestedFields checks if there are elements or values nested (e.g *opts.*rateSCost.Cost)
func checkNestedFields(elem string, values []string) bool {
	if len(strings.Split(elem, utils.NestingSep)) > 2 {
		return true
	}
	for _, val := range values {
		if len(strings.Split(val, utils.NestingSep)) > 2 {
			return true
		}
	}
	return false

}

type CDRFilters struct {
	Tenant    string
	ID        string
	FilterIDs []string
	APIOpts   map[string]interface{}
}
