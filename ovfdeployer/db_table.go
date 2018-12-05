package ovfdeployer

import (
	"fmt"

	"github.com/pkg/errors"
)

func (dbw dbWrapper) createKeyTable(tablename string) error {
	s := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS "%s" (
		"KEY" VARCHAR(30) PRIMARY KEY,
		"VAL" VARCHAR(100)
	)`, tablename)
	return dbw.updateDb(s)
}

func (dbw dbWrapper) sync2KeyTable(tablename string, table Table) error {
	logDebug("sync2KeyTable(%s, %+v)", tablename, table)
	err := dbw.createKeyTable(tablename)
	if err != nil {
		return err
	}
	db := dbw.db
	strsql := fmt.Sprintf(`DELETE FROM %s;`, tablename)
	_, err = db.Exec(strsql)
	if err != nil {
		return errors.Wrapf(err, "Error: %s", strsql)
	}
	for k, v := range table {
		if k == "" {
			continue
		}
		strsql = fmt.Sprintf(`INSERT INTO %s ("KEY", "VAL") VALUES ("%s", "%s");`, tablename, k, v)
		_, err := db.Exec(strsql)
		if err != nil {
			return errors.Wrapf(err, "Error: %s", strsql)
		}
	}
	return nil
}

func (dbw dbWrapper) syncFromKeyTable(tablename string) (Table, error) {
	logDebug("syncFromKeyTable(%s)", tablename)
	db := dbw.db
	t := make(Table, 10)
	strsql := fmt.Sprintf("SELECT KEY, VAL FROM %s;", tablename)
	rows, err := db.Query(strsql)
	if err != nil {
		return nil, errors.Wrapf(err, "Error: %s", strsql)
	}
	var k string
	var v string
	for rows.Next() {
		rows.Scan(&k, &v)
		if k == "" {
			continue
		}
		t[k] = v
	}
	return t, nil
}

func (dbw dbWrapper) createListTable(tablename string) error {
	s := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS "%s" (
			"ID" INTEGER PRIMARY KEY AUTOINCREMENT,
			"VAL" VARCHAR(100)
		)`, tablename)
	return dbw.updateDb(s)
}

func (dbw dbWrapper) getValFromKeyTable(tablename, key string) (string, error) {
	db := dbw.db
	strsql := fmt.Sprintf(`SELECT 
		VAL FROM %s 
		WHERE KEY = "%s";`, tablename, key)
	row := db.QueryRow(strsql)
	var val string
	err := row.Scan(&val)
	if err != nil {
		return "", errors.Wrapf(err, "Error: %s", strsql)
	}
	return val, nil
}

func (dbw dbWrapper) getKeysFromKeyTable(tablename, val string) ([]string, error) {
	db := dbw.db
	strsql := fmt.Sprintf(`SELECT KEY FROM %s WHERE VAL="%s";`,
		tablename, val)
	rows, err := db.Query(strsql)
	if err != nil {
		return nil, errors.Wrapf(err, "Error: %s", strsql)
	}
	var k string
	s := make([]string, 0)
	for rows.Next() {
		rows.Scan(&k)
		if k == "" {
			continue
		}
		s = append(s, k)
	}
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (dbw dbWrapper) updateKeyTable(tablename, key, val string) error {
	db := dbw.db
	strSQL := fmt.Sprintf(`UPDATE %s SET VAL="%s" WHERE KEY="%s";`, tablename, val, key)
	_, err := db.Exec(strSQL)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("db=%s sql=%s", dbw.dbname, strSQL))
	}
	return nil
}

func (dbw dbWrapper) createVMsTable() error {
	db := dbw.db
	createsql := `CREATE TABLE IF NOT EXISTS VMS 
	("SEQ" INTEGER PRIMARY KEY, 
		"VMID" INTEGER UNIQUE,
		"NAME"  VARCHAR(30) UNIQUE,
		"HOST_IP" VARCHAR(15),
		"DS_SIZE_MB" INTEGER, 
		"MEM_SIZE_MB" INTEGER,
		"DATASTORE" VARCHAR(100),
		"PORTGROUP" VARCHAR(100),
		"IS_REGISTERED" INTEGER DEFAULT 0);`
	_, err := db.Exec(createsql)
	if err != nil {
		return errors.Wrapf(err, createsql)
	}

	return nil
}
