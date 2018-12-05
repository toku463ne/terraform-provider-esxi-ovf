package ovfdeployer

import (
	"fmt"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func Test_openDb(t *testing.T) {
	err := os.Chdir("/tmp")
	if err != nil {
		t.Error("Failed to change current dir to /tmp")
	}

	dbw, err := openDb("TestOpenDB", "test")
	if err != nil {
		t.Errorf("Error openDb() %+v", err)
	}
	defer dbw.closeDb()
	//dbpath := fmt.Sprintf("/tmp/%s/test.db", workDir)
	strSQL := `
	CREATE TABLE IF NOT EXISTS "TEST" (
		"ID" INTEGER,
		"DESC" VARCHAR(100)
	);`
	err = dbw.updateDb(strSQL)
	if err != nil {
		t.Errorf("%+v", err)
	}
	d1 := "This is test"
	strSQL = fmt.Sprintf(`
	INSERT INTO TEST(ID, DESC) VALUES(1, "%s");
	`, d1)
	err = dbw.updateDbWithRetry(strSQL)
	if err != nil {
		t.Errorf("%+v", err)
	}
	db := dbw.db
	row := db.QueryRow(fmt.Sprintf(`SELECT 
		DESC FROM TEST WHERE ID = %d`, 1))
	if err != nil {
		t.Errorf("%+v", err)
	}
	var d2 string
	err = row.Scan(&d2)
	if err != nil {
		t.Errorf("%+v", err)
	}
	if d1 != d2 {
		t.Errorf("Got unexpected data from TEST table. Got=%s Expected=%s", d2, d1)
	}
	err = dbw.deleteDb()
	if err != nil {
		t.Errorf("%+v", err)
	}
}
