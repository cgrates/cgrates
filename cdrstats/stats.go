/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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

package cdrstats

type StatsQueue struct {
    cdrs []*QCDR
    config *CdrStatsConfig
    metrics map[string]*Metric
}

// Simplified cdr structure containing only the necessary info
type QCDR struct {
    SetupTime      time.Time
    AnswerTime     time.Time
    Usage          time.Duration
    Cost           float64
}

func NewStatsQueue(config *CdrStatsConfig) *StatsQueue{
    sq := &StatsQueue{
        config:config
        metrics:=make(map[string]*Metric, len(config.Metrics))
    }
    for _, m := range config.Metrics {
        metric := CreateMetric(m)
        if metric != nil{
			sq.metrics[m] = metric
        }
    }
}

func (sq *StatsQueue) AppendCDR(cdr *utils.StoredCdr){
    if sq.AcceptCDR(cdr){
        qcdr := sq.SimplifyCDR(cdr)
        sq.cdrs = append(sq.cdrs, qcdr)
        sq.AddToMetrics(qcdr)
    }
}

func (sq *StatsQueue) AddToMetrics(cdr *QCDR) {
    for _, metric:= range sq.metrics {
        metric.AddCdr(cdr)
    }
}

func (sq *StatsQueue) RemoveFromMetrics(cdr *QCDR) {
    for _, metric:= range sq.metrics {
        metric.RemoveCdr(cdr)
    }
}

func (sq *StatsQueue) SimplifyCDR(cdr *utils.StoredCdr) *QCDR{
    return &QCDR{
        SetupTime: cdr.SetupTime,
        AnswerTime: cdr.AnswerTime,
        Usage: cdr.Usage,
        Cost: cdr.Cost
    }
}

func (sq *StatsQueue) PurgeObsoleteCDRs {
    currentLength := len(sq.cdrs)
    if currentLength>sq.config.QueuedItems{
        for _, cdr := range sq.cdrs[:currentLength-sq.config.QueuedItems] {
            sq.RemoveFromMetrics(cdr)
        }
        sq.cdrs = sq.cdrs[currentLength-sq.config.QueuedItems:]
    }
    for i, cdr := range sq.cdrs {
        if time.Now().Sub(cdr.SetupTime) > sq.config.TimeWindow {
            sq.RemoveFromMetrics(cdr)
            continue
        } else {
            if i > 0 {
                sq.cdrs = sq.cdrs[i:]
            }
            break
        }
    }
}

func (sq *StatsQueue) AcceptCDR(cdr *utils.StoredCdr) bool {
    if len(sq.config.SetupInterval) > 0 {
        if cdr.SetupTime.Before(sq.config.SetupInterval[0]){
            return false
        }
        if len(sq.config.SetupInterval) > 1  && (cdr.SetupTime.Equals(sq.config.SetupInterval[1]) || cdr.SetupTime.After(sq.config.SetupInterval[1])){
            return false
        }
    }
    if len(sq.config.TOR) >0 && !utils.IsSliceMember(sq.config.TOR, cdr.TOR){
        return false
    }
    if len(sq.config.CdrHost) >0 && !utils.IsSliceMember(sq.config.CdrHost, cdr.CdrHost){
        return false
    }
    if len(sq.config.CdrSource) >0 && !utils.IsSliceMember(sq.config.CdrSource, cdr.CdrSource){
        return false
    }
    if len(sq.config.ReqType) >0 && !utils.IsSliceMember(sq.config.ReqType, cdr.ReqType){
        return false
    }
    if len(sq.config.Direction) >0 && !utils.IsSliceMember(sq.config.Direction, cdr.Direction){
        return false
    }
    if len(sq.config.Tenant) >0 && !utils.IsSliceMember(sq.config.Tenant, cdr.Tenant){
        return false
    }
    if len(sq.config.Category) >0 && !utils.IsSliceMember(sq.config.Category, cdr.Category){
        return false
    }
    if len(sq.config.Account) >0 && !utils.IsSliceMember(sq.config.Account, cdr.Account){
        return false
    }
    if len(sq.config.Subject) >0 && !utils.IsSliceMember(sq.config.Subject, cdr.Subject){
        return false
    }
    if len(sq.config.DestinationPrefix)>0 {
        found := false
        for _, prefix := range sq.config.DestinationPrefix {
            if cdr.Destination.HasPrefix(prefix){
                found=true
                break
            }
        }
        if !found{
            return false
        }
    }
    if len(sq.config.UsageInterval) > 0 {
        if cdr.Usage < sq.config.UsageInterval[0]{
            return false
        }
        if len(sq.config.UsageInterval) > 1  && cdr.Usage >= sq.config.UsageInterval[1]{
            return false
        }
    }
    if len(sq.config.MediationRunIds) >0 && !utils.IsSliceMember(sq.config.MediationRunIds, cdr.MediationRunId){
        return false
    }
    if len(sq.config.CostInterval) > 0 {
        if cdr.Cost < sq.config.CostInterval[0]{
            return false
        }
        if len(sq.config.CostInterval) > 1  && cdr.Cost >= sq.config.CostInterval[1]{
            return false
        }
    }
    return true
}
