CREATE TABLE IF NOT EXISTS ratingprofile (
	id SERIAL PRIMARY KEY,
	fallbackkey VARCHAR(512)
);
CREATE TABLE IF NOT EXISTS destination (
	id SERIAL PRIMARY KEY,
	ratingprofile INTEGER REFERENCES ratingprofile(id) ON DELETE CASCADE,
	name VARCHAR(512),
	prefixes TEXT
);
CREATE TABLE IF NOT EXISTS ratingdestinations (
	id SERIAL PRIMARY KEY,
	ratingprofile INTEGER REFERENCES ratingprofile(id) ON DELETE CASCADE,
	destination INTEGER REFERENCES destination(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS activationprofile  (
	id SERIAL PRIMARY KEY,
	destination INTEGER REFERENCES destination(id) ON DELETE CASCADE,
	activationtime TIMESTAMP
);
CREATE TABLE IF NOT EXISTS interval (
	id SERIAL PRIMARY KEY,
	activationprofile INTEGER REFERENCES activationprofile(id) ON DELETE CASCADE,
	years TEXT,
	months TEXT,
	monthdays TEXT,
	weekdays TEXT,
	starttime TIMESTAMP,
	endtime TIMESTAMP,
	weight FLOAT8,
	connectfee FLOAT8,
	price FLOAT8,
	pricedunits FLOAT8,
	rateincrements FLOAT8
);
CREATE TABLE IF NOT EXISTS minutebucket (
	id SERIAL PRIMARY KEY,
	destination INTEGER REFERENCES destination(id) ON DELETE CASCADE,
	seconds FLOAT8,
	weight FLOAT8,
	price FLOAT8,
	percent FLOAT8
);
CREATE TABLE IF NOT EXISTS unitcounter (
	id SERIAL PRIMARY KEY,
	direction TEXT,
	balance TEXT,
	units FLOAT8
);
CREATE TABLE IF NOT EXISTS unitcounterbucket (
	id SERIAL PRIMARY KEY,
	unitcounter INTEGER REFERENCES unitcounter(id) ON DELETE CASCADE,
	minutebucket INTEGER REFERENCES minutebucket(id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS action (
	id SERIAL PRIMARY KEY,
	minutebucket INTEGER REFERENCES minutebucket(id) ON DELETE CASCADE,
	actiontype TEXT,
	balance TEXT,
	direction TEXT,
	units FLOAT8,
	weight FLOAT8
);
CREATE TABLE IF NOT EXISTS actiontrigger (
	id SERIAL PRIMARY KEY,
	destination INTEGER REFERENCES destination(id) ON DELETE CASCADE,
	actions INTEGER REFERENCES action(id) ON DELETE CASCADE,
	balance TEXT,
	direction TEXT,
	thresholdvalue FLOAT8,
	weight FLOAT8,
	executed BOOL
);
CREATE TABLE IF NOT EXISTS balance (
	id SERIAL PRIMARY KEY,
	name TEXT,
	value FLOAT8
);
CREATE TABLE IF NOT EXISTS userbalance (
	id SERIAL PRIMARY KEY,
	unitcounter INTEGER REFERENCES unitcounter(id) ON DELETE CASCADE,
	minutebucket INTEGER REFERENCES minutebucket(id) ON DELETE CASCADE,
	actiontriggers INTEGER REFERENCES actiontrigger(id) ON DELETE CASCADE,
	balances INTEGER REFERENCES balance(id) ON DELETE CASCADE,
	type TEXT
);
CREATE TABLE IF NOT EXISTS actiontiming (
	id SERIAL PRIMARY KEY,
	tag TEXT,
	userbalances INTEGER REFERENCES userbalance(id) ON DELETE CASCADE,
	timing INTEGER REFERENCES interval(id) ON DELETE CASCADE,
	actions INTEGER REFERENCES action(id) ON DELETE CASCADE,
	weight FLOAT8
);


CREATE TABLE IF NOT EXISTS sharedgroup (
	id SERIAL PRIMARY KEY,
	account TEXT,
	strategy TEXT,
	ratesubject TEXT
);
