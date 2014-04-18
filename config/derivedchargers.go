/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2014 ITsysCOM GmbH

This program is free software: you can Storagetribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITH*out ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package config

import (
	"errors"
	"github.com/cgrates/cgrates/utils"
)

type DerivedCharger struct {
	RunId            string // Unique runId in the chain
	ReqTypeField     string // Field containing request type info, number in case of csv source, '^' as prefix in case of static values
	DirectionField   string // Field containing direction info
	TenantField      string // Field containing tenant info
	TorField         string // Field containing tor info
	AccountField     string // Field containing account information
	SubjectField     string // Field containing subject information
	DestinationField string // Field containing destination information
	SetupTimeField   string // Field containing setup time information
	AnswerTimeField  string // Field containing answer time information
	DurationField    string // Field containing duration information
}

type DerivedChargers []*DerivedCharger

// Precheck that RunId is unique
func (dcs DerivedChargers) Append(dc *DerivedCharger) (DerivedChargers, error) {
	if dc.RunId == utils.DEFAULT_RUNID {
		return nil, errors.New("Reserved RunId")
	}
	for _, dcLocal := range dcs {
		if dcLocal.RunId == dc.RunId {
			return nil, errors.New("Duplicated RunId")
		}
	}
	return append(dcs, dc), nil
}
