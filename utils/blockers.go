// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"fmt"
	"strconv"
	"strings"
)

type DynamicBlockers []*DynamicBlocker

// NewDynamicBlockersFromString will build the Blockers that contains slices of FilterIDs and Blocker from a single value. This process is helped by separators
func NewDynamicBlockersFromString(value, blkSep, fltrSep string) (blkrs DynamicBlockers, err error) {
	if len(value) == 0 {
		return DynamicBlockers{{}}, nil
	}
	valSeparated := strings.Split(value, blkSep)
	lenVals := len(valSeparated)
	if lenVals%2 != 0 {
		return nil, fmt.Errorf("invalid DynamicBlocker format for string <%s>", value)
	}
	for idx := 0; idx < lenVals; idx += 2 {
		var fltrIDs []string
		if len(valSeparated[idx]) != 0 {
			fltrIDs = strings.Split(valSeparated[idx], fltrSep)
		}
		var blocker bool
		if len(valSeparated[idx+1]) != 0 {
			if blocker, err = strconv.ParseBool(valSeparated[idx+1]); err != nil {
				return nil, fmt.Errorf("cannot convert bool with value: <%v> into Blocker", valSeparated[idx+1])
			}
		}
		blkrs = append(blkrs, &DynamicBlocker{FilterIDs: fltrIDs, Blocker: blocker})
	}
	return
}

// String will set the Blockers as a string pattern
func (blkrs DynamicBlockers) String(blkSep, fltrSep string) (value string) {
	if len(blkrs) == 0 {
		return
	}
	strBlockers := make([]string, len(blkrs))
	for idx, val := range blkrs {
		strBlockers[idx] = val.String(blkSep, fltrSep)
	}
	return strings.Join(strBlockers, blkSep)
}

type DynamicBlocker struct {
	FilterIDs []string
	Blocker   bool
}

// String will set the DynamicBlocker as a string pattern
func (blckr DynamicBlocker) String(blkSep, fltrSep string) (out string) {
	blocker := "false"
	if blckr.Blocker {
		blocker = "true"
	}
	return strings.Join(blckr.FilterIDs, fltrSep) + blkSep + blocker
}

func (dB *DynamicBlocker) FieldAsInterface(fldPath []string) (any, error) {
	if len(fldPath) != 1 {
		return nil, ErrNotFound
	}
	switch fldPath[0] {
	case FilterIDs:
		return dB.FilterIDs, nil
	case Blocker:
		return dB.Blocker, nil
	default:
		fld, idx := GetPathIndex(fldPath[0])
		if idx != nil &&
			fld == FilterIDs {
			if *idx < len(dB.FilterIDs) {
				return dB.FilterIDs[*idx], nil
			}
		}
	}
	return nil, ErrNotFound
}

// Clone will clone the Blockers
func (blckrs DynamicBlockers) Clone() (clBlkrs DynamicBlockers) {
	if blckrs == nil {
		return
	}
	clBlkrs = make(DynamicBlockers, len(blckrs))
	for i, value := range blckrs {
		clBlkrs[i] = value.Clone()
	}
	return
}

// Clone will clone the a DynamicBlocker
func (blckr *DynamicBlocker) Clone() (cln *DynamicBlocker) {
	if blckr == nil {
		return
	}
	cln = &DynamicBlocker{
		Blocker: blckr.Blocker,
	}
	if blckr.FilterIDs != nil {
		cln.FilterIDs = make([]string, len(blckr.FilterIDs))
		copy(cln.FilterIDs, blckr.FilterIDs)
	}
	return
}
