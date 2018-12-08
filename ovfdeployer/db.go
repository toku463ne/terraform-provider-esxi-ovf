package ovfdeployer

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3" // blank import here because there is no main
	"github.com/pkg/errors"
)

type dbWrapper struct {
	dbname    string
	namespace string
	db        *sql.DB
}

func getDbPath(namespace, dbname string) string {
	dbdir := fmt.Sprintf("./%s/%s", workDir, namespace)
	if _, err := os.Stat(dbdir); os.IsNotExist(err) {
		os.MkdirAll(dbdir, 0755)
	}
	return fmt.Sprintf("./%s/%s.db", dbdir, dbname)
}

func openDb(namespace, dbname string) (*dbWrapper, error) {
	dbpath := getDbPath(namespace, dbname)
	logDebug("openDb(%s, %s) dbpath=%s", namespace, dbname, dbpath)
	db, err := sql.Open("sqlite3", dbpath)
	if err != nil {
		return nil, errors.Wrapf(err, "Error opening DB %s", dbname)
	}
	d := new(dbWrapper)
	d.dbname = dbname
	d.namespace = namespace
	d.db = db
	return d, nil
}

func (dbw dbWrapper) openDb() (*dbWrapper, error) {
	return openDb(dbw.namespace, dbw.dbname)
}

func (dbw dbWrapper) deleteDb() error {
	return deleteDb(dbw.namespace, dbw.dbname)
}

func deleteDb(namespace, dbname string) error {
	logInfo(`Deleting %s%s`, namespace, dbname)
	return os.Remove(getDbPath(namespace, dbname))
}

func (dbw dbWrapper) updateDb(strsql string) error {
	logDebug(`%s/%s sql="%s"`, dbw.namespace, dbw.dbname, strsql)
	db := dbw.db
	tx, err := db.Begin()
	if err != nil {
		return errors.Wrapf(err, "Error begining transaction")
	}
	stmt, err := tx.Prepare(strsql)
	if err != nil {
		return errors.Wrapf(err, "[sqlite] %s", strsql)
	}
	defer stmt.Close()
	if _, err := stmt.Exec(); err != nil {
		return errors.Wrapf(err, "Error executing SQL %s", strsql)
	}
	if err := tx.Commit(); err != nil {
		err = errors.Wrapf(err, "Failed to commit. %s", strsql)
	}
	return err
}

func (dbw dbWrapper) updateDbWithRetry(strsql string) error {
	for i := 0; i < lockRetryCount; i++ {
		err := dbw.updateDb(strsql)
		if err == nil {
			return nil
		}
		if assertError(err, SqliteMsgDbLocked) {
			time.Sleep(lockRetryInterval)
			continue
		} else {
			return err
		}
	}
	return nil
}

func (dbw dbWrapper) closeDb() error {
	db := dbw.db
	return db.Close()
}
