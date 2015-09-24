package engine

import (
	"fmt"

	"github.com/cgrates/cgrates/utils"
	"github.com/gocql/gocql"
)

type CassandraStorage struct {
	keyspace string
	db       *gocql.Session
	ms       Marshaler
}

func NewCassandraStorage(addresses []string, keyspace, mrshlerStr string) (*CassandraStorage, error) {
	cluster := gocql.NewCluster(addresses...)
	cluster.Keyspace = keyspace
	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}

	var mrshler Marshaler
	if mrshlerStr == utils.MSGPACK {
		mrshler = NewCodecMsgpackMarshaler()
	} else if mrshlerStr == utils.JSON {
		mrshler = new(JSONMarshaler)
	} else {
		return nil, fmt.Errorf("Unsupported marshaler: %v", mrshlerStr)
	}
	return &CassandraStorage{db: session, keyspace: keyspace, ms: mrshler}, nil
}

func (cs *CassandraStorage) Close() {
	cs.db.Close()
}

func (cs *CassandraStorage) Flush(ignore string) (err error) {
	return cs.db.Query(fmt.Sprintf("delete * from %s", cs.keyspace)).Exec()
}

func (cs *CassandraStorage) SetRatedCdr(*StoredCdr) error { return nil }
func (cs *CassandraStorage) LogCallCost(cgrid, source, runid string, cc *CallCost) error {
	/*if cc == nil {
		return nil
	}
	tss, err := json.Marshal(cc.Timespans)
	if err != nil {
		Logger.Err(fmt.Sprintf("Error marshalling timespans to json: %v", err))
		return err
	}
	_, err = self.Db.Exec(fmt.Sprintf("INSERT INTO %s (cgrid,runid,tor,direction,tenant,category,account,subject,destination,cost,timespans,cost_source,created_at) VALUES ('%s','%s','%s','%s','%s','%s','%s','%s','%s',%f,'%s','%s','%s') ON DUPLICATE KEY UPDATE tor=values(tor),direction=values(direction),tenant=values(tenant),category=values(category),account=values(account),subject=values(subject),destination=values(destination),cost=values(cost),timespans=values(timespans),cost_source=values(cost_source),updated_at='%s'",
		utils.TBL_COST_DETAILS,
		cgrid,
		runid,
		cc.TOR,
		cc.Direction,
		cc.Tenant,
		cc.Category,
		cc.Account,
		cc.Subject,
		cc.Destination,
		cc.Cost,
		tss,
		source,
		time.Now().Format(time.RFC3339),
		time.Now().Format(time.RFC3339)))
	if err != nil {
		Logger.Err(fmt.Sprintf("failed to execute insert statement: %v", err))
		return err
	}*/
	return nil
}
func (cs *CassandraStorage) GetCallCostLog(cgrid, source, runid string) (*CallCost, error) {
	return nil, nil
}
func (cs *CassandraStorage) GetStoredCdrs(*utils.CdrsFilter) ([]*StoredCdr, int64, error) {
	return nil, 0, nil
}
func (cs *CassandraStorage) RemStoredCdrs([]string) error                      { return nil }
func (cs *CassandraStorage) LogError(uuid, source, runid, errstr string) error { return nil }
func (cs *CassandraStorage) LogActionTrigger(ubId, source string, at *ActionTrigger, as Actions) error {
	return nil
}
func (cs *CassandraStorage) LogActionPlan(source string, at *ActionPlan, as Actions) error { return nil }
