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

package tpes

import (
	"bytes"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestTPEnewTPFilters(t *testing.T) {
	// dataDB := &engine.DataDBM
	// dm := &engine.NewDataManager()
	cfg := config.NewDefaultCGRConfig()
	connMng := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(&engine.DataDBMock{
		GetFilterDrvF: func(ctx *context.Context, str1, str2 string) (*engine.Filter, error) {
			fltr := &engine.Filter{
				Tenant: utils.CGRateSorg,
				ID:     "fltr_for_prf",
				Rules: []*engine.FilterRule{
					{
						Type:    utils.MetaString,
						Element: "~*req.Subject",
						Values:  []string{"1004", "6774", "22312"},
					},
					{
						Type:    utils.MetaString,
						Element: "~*opts.Subsystems",
						Values:  []string{"*attributes"},
					},
					{
						Type:    utils.MetaPrefix,
						Element: "~*req.Destinations",
						Values:  []string{"+0775", "+442"},
					},
					{
						Type:    utils.MetaExists,
						Element: "~*req.NumberOfEvents",
					},
				},
			}
			return fltr, nil
		},
	}, cfg, connMng)
	exp := &TPFilters{
		dm: dm,
	}
	rcv := newTPFilters(dm)
	if rcv.dm != exp.dm {
		t.Errorf("Expected %v \nbut received %v", exp, rcv)
	}
}

func TestTPEExportItemsFilters(t *testing.T) {
	wrtr := new(bytes.Buffer)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	tpFltr := TPFilters{
		dm: dm,
	}
	fltr := &engine.Filter{
		Tenant: utils.CGRateSorg,
		ID:     "fltr_for_prf",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Subject",
				Values:  []string{"1004", "6774", "22312"},
			},
			{
				Type:    utils.MetaString,
				Element: "~*opts.Subsystems",
				Values:  []string{"*attributes"},
			},
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Destinations",
				Values:  []string{"+0775", "+442"},
			},
			{
				Type:    utils.MetaExists,
				Element: "~*req.NumberOfEvents",
			},
		},
	}
	tpFltr.dm.SetFilter(context.Background(), fltr, false)
	err := tpFltr.exportItems(context.Background(), wrtr, "cgrates.org", []string{"fltr_for_prf"})
	if err != nil {
		t.Errorf("Expected nil\n but received %v", err)
	}
}

func TestTPEExportItemsFiltersNoDbConn(t *testing.T) {
	engine.Cache.Clear(nil)
	wrtr := new(bytes.Buffer)
	tpFltr := TPFilters{
		dm: nil,
	}
	fltr := &engine.Filter{
		Tenant: utils.CGRateSorg,
		ID:     "fltr_for_prf",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Subject",
				Values:  []string{"1004", "6774", "22312"},
			},
			{
				Type:    utils.MetaString,
				Element: "~*opts.Subsystems",
				Values:  []string{"*attributes"},
			},
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Destinations",
				Values:  []string{"+0775", "+442"},
			},
			{
				Type:    utils.MetaExists,
				Element: "~*req.NumberOfEvents",
			},
		},
	}
	tpFltr.dm.SetFilter(context.Background(), fltr, false)
	err := tpFltr.exportItems(context.Background(), wrtr, "cgrates.org", []string{"fltr_for_prf"})
	if err != utils.ErrNoDatabaseConn {
		t.Errorf("Expected %v\n but received %v", utils.ErrNoDatabaseConn, err)
	}
}

func TestTPEExportItemsFiltersIDNotFound(t *testing.T) {
	wrtr := new(bytes.Buffer)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	tpFltr := TPFilters{
		dm: dm,
	}
	fltr := &engine.Filter{
		Tenant: utils.CGRateSorg,
		ID:     "fltr_for_prf",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Subject",
				Values:  []string{"1004", "6774", "22312"},
			},
			{
				Type:    utils.MetaString,
				Element: "~*opts.Subsystems",
				Values:  []string{"*attributes"},
			},
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Destinations",
				Values:  []string{"+0775", "+442"},
			},
			{
				Type:    utils.MetaExists,
				Element: "~*req.NumberOfEvents",
			},
		},
	}
	tpFltr.dm.SetFilter(context.Background(), fltr, false)
	err := tpFltr.exportItems(context.Background(), wrtr, "cgrates.org", []string{"fltr_not_for_prf"})
	errExpect := "<NOT_FOUND> cannot find Filters with id: <fltr_not_for_prf>"
	if err.Error() != errExpect {
		t.Errorf("Expected %v\n but received %v", errExpect, err)
	}
}
