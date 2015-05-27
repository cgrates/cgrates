package engine

type TpReader struct {
	tpid              string
	ratingStorage     RatingStorage
	accountingStorage AccountingStorage
	lr                LoadReader
	tp                *TPData
}

func NewTpReader(rs RatingStorage, as AccountingStorage, lr LoadReader, tpid string) *TpReader {
	return &TpReader{
		tpid:              tpid,
		ratingStorage:     rs,
		accountingStorage: as,
		lr:                lr,
		tp:                NewTPData(),
	}
}

func (tpr *TpReader) ShowStatistics() {
	tpr.tp.ShowStatistics()
}

func (tpr *TpReader) IsDataValid() bool {
	return tpr.tp.IsValid()
}

func (tpr *TpReader) WriteToDatabase(flush, verbose bool) (err error) {
	return tpr.tp.WriteToDatabase(tpr.ratingStorage, tpr.accountingStorage, flush, verbose)
}

func (tpr *TpReader) LoadDestinations() (err error) {
	tpDests, err := tpr.lr.GetTpDestinations(tpr.tpid, "")
	if err != nil {
		return err
	}
	return tpr.tp.LoadDestinations(tpDests)
}

func (tpr *TpReader) LoadTimings() (err error) {
	tps, err := tpr.lr.GetTpTimings(tpr.tpid, "")
	if err != nil {
		return err
	}
	return tpr.tp.LoadTimings(tps)
}

func (tpr *TpReader) LoadRates() (err error) {
	tps, err := tpr.lr.GetTpRates(tpr.tpid, "")
	if err != nil {
		return err
	}
	return tpr.tp.LoadRates(tps)
}

func (tpr *TpReader) LoadAll() error {
	var err error
	if err = tpr.LoadDestinations(); err != nil {
		return err
	}
	if err = tpr.LoadTimings(); err != nil {
		return err
	}
	if err = tpr.LoadRates(); err != nil {
		return err
	}
	if err = tpr.LoadDestinationRates(); err != nil {
		return err
	}
	if err = tpr.LoadRatingPlans(); err != nil {
		return err
	}
	if err = tpr.LoadRatingProfiles(); err != nil {
		return err
	}
	if err = tpr.LoadSharedGroups(); err != nil {
		return err
	}
	if err = tpr.LoadLCRs(); err != nil {
		return err
	}
	if err = tpr.LoadActions(); err != nil {
		return err
	}
	if err = tpr.LoadActionTimings(); err != nil {
		return err
	}
	if err = tpr.LoadActionTriggers(); err != nil {
		return err
	}
	if err = tpr.LoadAccountActions(); err != nil {
		return err
	}
	if err = tpr.LoadDerivedChargers(); err != nil {
		return err
	}
	if err = tpr.LoadCdrStats(); err != nil {
		return err
	}
	return nil
}
