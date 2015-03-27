/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2015 ITsysCOM GmbH

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

package engine

/*
func NewMediator(connector Connector, logDb LogStorage, cdrDb CdrStorage, st StatsInterface, cfg *config.CGRConfig) (m *Mediator, err error) {
	m = &Mediator{
		connector: connector,
		logDb:     logDb,
		cdrDb:     cdrDb,
		stats:     st,
		cgrCfg:    cfg,
	}
	if m.cgrCfg.MediatorStats != "" {
		if m.cgrCfg.MediatorStats != utils.INTERNAL {
			if s, err := NewProxyStats(m.cgrCfg.MediatorStats); err == nil {
				m.stats = s
			} else {
				Logger.Err(fmt.Sprintf("Errors connecting to CDRS stats service (mediator): %s", err.Error()))
			}
		}
	} else {
		// disable stats for mediator
		m.stats = nil
	}
	return m, nil
}
*/
