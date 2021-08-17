/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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

package ees

/*
func TestPosterJsonMapID(t *testing.T) {
	pstrEE := &PosterJSONMapEE{
		id: "3",
	}
	if rcv := pstrEE.ID(); !reflect.DeepEqual(rcv, "3") {
		t.Errorf("Expected %+v but got %+v", "3", rcv)
	}
}

func TestPosterJsonMapGetMetrics(t *testing.T) {
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	if err != nil {
		t.Error(err)
	}
	pstrEE := &PosterJSONMapEE{
		dc: dc,
	}

	if rcv := pstrEE.GetMetrics(); !reflect.DeepEqual(rcv, pstrEE.dc) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(rcv), utils.ToJSON(pstrEE.dc))
	}
}

func TestPosterJsonMapNewPosterJSONMapEECase2(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaAMQPV1jsonMap
	cgrCfg.EEsCfg().Exporters[0].ExportPath = utils.EmptyString
	filterS := engine.NewFilterS(cgrCfg, nil, nil)
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	if err != nil {
		t.Error(err)
	}
	pstrJSON, err := NewPosterJSONMapEE(cgrCfg, 0, filterS, dc)
	if err != nil {
		t.Error(err)
	}
	pstrJSONExpect := engine.NewAMQPv1Poster(cgrCfg.EEsCfg().Exporters[0].ExportPath,
		cgrCfg.EEsCfg().Exporters[0].Attempts, cgrCfg.EEsCfg().Exporters[0].Opts)
	if !reflect.DeepEqual(pstrJSON.poster, pstrJSONExpect) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(pstrJSONExpect), utils.ToJSON(pstrJSON.poster))
	}
}

func TestPosterJsonMapNewPosterJSONMapEECase3(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaSQSjsonMap
	cgrCfg.EEsCfg().Exporters[0].ExportPath = utils.EmptyString
	filterS := engine.NewFilterS(cgrCfg, nil, nil)
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	if err != nil {
		t.Error(err)
	}
	pstrJSON, err := NewPosterJSONMapEE(cgrCfg, 0, filterS, dc)
	if err != nil {
		t.Error(err)
	}

	if _, canCast := pstrJSON.poster.(*engine.SQSPoster); !canCast {
		t.Error("Can't cast")
	}
}

func TestPosterJsonMapNewPosterJSONMapEECase4(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaKafkajsonMap
	filterS := engine.NewFilterS(cgrCfg, nil, nil)
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	if err != nil {
		t.Error(err)
	}
	pstrJSON, err := NewPosterJSONMapEE(cgrCfg, 0, filterS, dc)
	if err != nil {
		t.Error(err)
	}
	pstrJSONExpect := engine.NewKafkaPoster(cgrCfg.EEsCfg().Exporters[0].ExportPath,
		cgrCfg.EEsCfg().Exporters[0].Attempts, cgrCfg.EEsCfg().Exporters[0].Opts)
	if !reflect.DeepEqual(pstrJSON.poster, pstrJSONExpect) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(pstrJSONExpect), utils.ToJSON(pstrJSON.poster))
	}
}

func TestPosterJsonMapNewPosterJSONMapEECase5(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaS3jsonMap
	filterS := engine.NewFilterS(cgrCfg, nil, nil)
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	if err != nil {
		t.Error(err)
	}
	pstrJSON, err := NewPosterJSONMapEE(cgrCfg, 0, filterS, dc)
	if err != nil {
		t.Error(err)
	}
	pstrJSONExpect := engine.NewS3Poster(cgrCfg.EEsCfg().Exporters[0].ExportPath,
		cgrCfg.EEsCfg().Exporters[0].Attempts, cgrCfg.EEsCfg().Exporters[0].Opts)
	if !reflect.DeepEqual(pstrJSON.poster, pstrJSONExpect) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(pstrJSONExpect), utils.ToJSON(pstrJSON.poster))
	}
}

func TestPosterJsonMapExportEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaSQSjsonMap
	cgrEv := new(utils.CGREvent)
	newIDb := engine.NewInternalDB(nil, nil, true)
	newDM := engine.NewDataManager(newIDb, cgrCfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cgrCfg, nil, newDM)
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	if err != nil {
		t.Error(err)
	}

	pstrEE, err := NewPosterJSONMapEE(cgrCfg, 0, filterS, dc)
	if err != nil {
		t.Error(err)
	}
	cgrEv.Event = map[string]interface{}{
		"test": "string",
	}
	cgrCfg.EEsCfg().Exporters[pstrEE.cfgIdx].Fields = []*config.FCTemplate{
		{
			Path: "*exp.1", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*req.field1", utils.InfieldSep),
		},
		{
			Path: "*exp.2", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("*req.field2", utils.InfieldSep),
		},
	}
	for _, field := range cgrCfg.EEsCfg().Exporters[pstrEE.cfgIdx].Fields {
		field.ComputePath()
	}
	errExpect := "MissingRegion: could not find region configuration"
	if err := pstrEE.ExportEvent(cgrEv); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %q but received %q", errExpect, err)
	}
	dcExpect := int64(1)
	if !reflect.DeepEqual(dcExpect, pstrEE.dc.MapStorage[utils.NumberOfEvents]) {
		t.Errorf("Expected %q but received %q", dcExpect, pstrEE.dc.MapStorage[utils.NumberOfEvents])
	}
	cgrCfg.EEsCfg().Exporters[pstrEE.cfgIdx].ComputeFields()
	if err := pstrEE.ExportEvent(cgrEv); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %q but received %q", errExpect, err)
	}
	dcExpect = int64(2)
	if !reflect.DeepEqual(dcExpect, pstrEE.dc.MapStorage[utils.NumberOfEvents]) {
		t.Errorf("Expected %q but received %q", dcExpect, pstrEE.dc.MapStorage[utils.NumberOfEvents])
	}
}

type testPoster struct {
	body []byte
}

func (pstr *testPoster) Close() {}
func (pstr *testPoster) Post(body []byte, key string) error {
	pstr.body = body
	return nil
}
func TestPosterJsonMapExportEvent1(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaAMQPjsonMap
	cgrEv := new(utils.CGREvent)
	newIDb := engine.NewInternalDB(nil, nil, true)
	newDM := engine.NewDataManager(newIDb, cgrCfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cgrCfg, nil, newDM)
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	if err != nil {
		t.Error(err)
	}
	////
	////
	tstPstr := &testPoster{}
	pstrEE := &PosterJSONMapEE{
		id:      cgrCfg.EEsCfg().Exporters[0].ID,
		cgrCfg:  cgrCfg,
		cfgIdx:  0,
		filterS: filterS,
		dc:      dc,
		poster:  tstPstr,
		reqs:    newConcReq(0),
	}
	// pstrEE.poster = tstPstr
	cgrEv.Event = map[string]interface{}{
		"test": "string",
	}
	cgrCfg.EEsCfg().Exporters[0].Fields = []*config.FCTemplate{
		{
			Path: "*exp.1", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*req.field1", utils.InfieldSep),
		},
		{
			Path: "*exp.2", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("*req.field2", utils.InfieldSep),
		},
	}
	for _, field := range cgrCfg.EEsCfg().Exporters[0].Fields {
		field.ComputePath()
	}
	cgrCfg.EEsCfg().Exporters[0].ComputeFields()
	if err := pstrEE.ExportEvent(cgrEv); err != nil {
		t.Error(err)
	}
	dcExpect := int64(1)
	if !reflect.DeepEqual(dcExpect, pstrEE.dc.MapStorage[utils.NumberOfEvents]) {
		t.Errorf("Expected %q but received %q", dcExpect, pstrEE.dc.MapStorage[utils.NumberOfEvents])
	}
	bodyExpect := map[string]interface{}{
		"2": "*req.field2",
	}
	var rcv map[string]interface{}
	if err := json.Unmarshal(tstPstr.body, &rcv); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, bodyExpect) {
		t.Errorf("Expected %s but received %s", utils.ToJSON(bodyExpect), utils.ToJSON(rcv))
	}
}

func TestPosterJsonMapExportEvent2(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaSQSjsonMap
	cgrEv := new(utils.CGREvent)
	newIDb := engine.NewInternalDB(nil, nil, true)
	newDM := engine.NewDataManager(newIDb, cgrCfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cgrCfg, nil, newDM)
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	if err != nil {
		t.Error(err)
	}

	pstrEE, err := NewPosterJSONMapEE(cgrCfg, 0, filterS, dc)
	if err != nil {
		t.Error(err)
	}
	cgrEv.Event = map[string]interface{}{
		"test": "string",
	}
	cgrCfg.EEsCfg().Exporters[0].Fields = []*config.FCTemplate{
		{
			Path: "*exp.1", Type: utils.MetaVariable,
			Value:   config.NewRSRParsersMustCompile("~*req.field1", utils.InfieldSep),
			Filters: []string{"*wrong-type"},
		},
		{
			Path: "*exp.1", Type: utils.MetaVariable,
			Value:   config.NewRSRParsersMustCompile("~*req.field1", utils.InfieldSep),
			Filters: []string{"*wrong-type"},
		},
	}
	for _, field := range cgrCfg.EEsCfg().Exporters[0].Fields {
		field.ComputePath()
	}
	cgrCfg.EEsCfg().Exporters[0].ComputeFields()
	errExpect := "inline parse error for string: <*wrong-type>"
	if err := pstrEE.ExportEvent(cgrEv); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %q but received %q", errExpect, err)
	}
	dcExpect := int64(1)
	if !reflect.DeepEqual(dcExpect, pstrEE.dc.MapStorage[utils.NumberOfEvents]) {
		t.Errorf("Expected %q but received %q", dcExpect, pstrEE.dc.MapStorage[utils.NumberOfEvents])
	}
}

func TestPosterJsonMapExportEvent3(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaSQSjsonMap
	cgrEv := new(utils.CGREvent)
	newIDb := engine.NewInternalDB(nil, nil, true)
	newDM := engine.NewDataManager(newIDb, cgrCfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cgrCfg, nil, newDM)
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	if err != nil {
		t.Error(err)
	}

	pstrEE, err := NewPosterJSONMapEE(cgrCfg, 0, filterS, dc)
	if err != nil {
		t.Error(err)
	}
	cgrEv.Event = map[string]interface{}{
		"test": "string",
	}
	cgrEv.Event = map[string]interface{}{
		"test": make(chan int),
	}
	cgrCfg.EEsCfg().Exporters[0].Fields = []*config.FCTemplate{{}}
	for _, field := range cgrCfg.EEsCfg().Exporters[0].Fields {
		field.ComputePath()
	}
	cgrCfg.EEsCfg().Exporters[0].ComputeFields()
	errExpect := "json: unsupported type: chan int"
	if err := pstrEE.ExportEvent(cgrEv); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %q but received %q", errExpect, err)
	}
	dcExpect := int64(1)
	if !reflect.DeepEqual(dcExpect, pstrEE.dc.MapStorage[utils.NumberOfEvents]) {
		t.Errorf("Expected %q but received %q", dcExpect, pstrEE.dc.MapStorage[utils.NumberOfEvents])
	}
	pstrEE.OnEvicted("test", "test")
}

type mockPoster struct {
	wg *sync.WaitGroup
}

func (mp mockPoster) Post(body []byte, key string) error {
	// resp, err := http.Get(mp.url)
	// if err != nil {
	// 	return err
	// }
	// defer resp.Body.Close()
	time.Sleep(3 * time.Second)
	mp.wg.Done()
	return nil
}

func (mockPoster) Close() {
	return
}

func TestPosterJsonMapSync(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	var cfgIdx int
	cfgIdx = 0

	cgrCfg.EEsCfg().Exporters[cfgIdx].Type = "*http_json_map"
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		cgrCfg.EEsCfg().Exporters[cfgIdx].Timezone,
		cgrCfg.GeneralCfg().DefaultTimezone))
	if err != nil {
		t.Error(err)
	}

	//Create an event
	cgrEvent := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			"Account":     "1001",
			"Destination": "1002",
		},
	}

	var wg1 = &sync.WaitGroup{}

	wg1.Add(3)

	test := make(chan struct{})
	go func() {
		wg1.Wait()
		close(test)
	}()

	mckPoster := mockPoster{
		wg: wg1,
	}
	exp := &PosterJSONMapEE{
		id:      cgrCfg.EEsCfg().Exporters[cfgIdx].ID,
		cgrCfg:  cgrCfg,
		cfgIdx:  cfgIdx,
		filterS: new(engine.FilterS),
		poster:  mckPoster,
		dc:      dc,
		reqs:    newConcReq(cgrCfg.EEsCfg().Exporters[cfgIdx].ConcurrentRequests),
	}

	for i := 0; i < 3; i++ {
		go exp.ExportEvent(cgrEvent)
	}

	select {
	case <-test:
		return
	case <-time.After(4 * time.Second):
		t.Error("Can't asynchronously export events")
	}
}

func TestPosterJsonMapSyncLimit(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	var cfgIdx int
	cfgIdx = 0

	cgrCfg.EEsCfg().Exporters[cfgIdx].Type = "*http_json_map"
	cgrCfg.EEsCfg().Exporters[cfgIdx].ConcurrentRequests = 1
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		cgrCfg.EEsCfg().Exporters[cfgIdx].Timezone,
		cgrCfg.GeneralCfg().DefaultTimezone))
	if err != nil {
		t.Error(err)
	}

	//Create an event
	cgrEvent := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			"Account":     "1001",
			"Destination": "1002",
		},
	}

	var wg1 = &sync.WaitGroup{}

	wg1.Add(3)

	test := make(chan struct{})
	go func() {
		wg1.Wait()
		close(test)
	}()

	mckPoster := mockPoster{
		wg: wg1,
	}
	exp := &PosterJSONMapEE{
		id:      cgrCfg.EEsCfg().Exporters[cfgIdx].ID,
		cgrCfg:  cgrCfg,
		cfgIdx:  cfgIdx,
		filterS: new(engine.FilterS),
		poster:  mckPoster,
		dc:      dc,
		reqs:    newConcReq(cgrCfg.EEsCfg().Exporters[cfgIdx].ConcurrentRequests),
	}

	for i := 0; i < 3; i++ {
		go exp.ExportEvent(cgrEvent)
	}

	select {
	case <-test:
		t.Error("Should not have been possible to asynchronously export events")
	case <-time.After(4 * time.Second):
		return
	}
}
*/
