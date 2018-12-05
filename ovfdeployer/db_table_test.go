package ovfdeployer

import (
	"reflect"
	"testing"
)

func Test_syncKeyTable(t *testing.T) {
	err := init4Test("Test_syncKeyTable")
	if err != nil {
		t.Errorf("%+v", err)
	}

	namespace := "testSyncKeyTable"
	dbname := "test"
	dbw, err := openDb(namespace, dbname)
	if err != nil {
		t.Errorf("Error openDb() %+v", err)
	}
	defer dbw.closeDb()
	tablename := "test"
	tb := Table{
		"k1": "v1",
		"k2": "v2",
		"k3": "v3",
	}
	if err := dbw.sync2KeyTable(tablename, tb); err != nil {
		t.Errorf("%+v", err)
	}
	if err = dbw.closeDb(); err != nil {
		t.Errorf("%+v", err)
	}

	dbw, err = openDb(namespace, dbname)
	if err != nil {
		t.Errorf("Error openDb() %+v", err)
	}
	tb2, err := dbw.syncFromKeyTable(tablename)
	if err != nil {
		t.Errorf("%+v", err)
	}
	if !reflect.DeepEqual(tb, tb2) {
		t.Errorf("Error asserting syncTo and syncFrom. syncTo=%+v syncFrom=%+v", tb, tb2)
	}

	v, err := dbw.getValFromKeyTable(tablename, "k2")
	if err != nil {
		t.Errorf("%+v", err)
	}
	if v != "v2" {
		t.Errorf("Unexpected res from getValFromKeyTable(). Got=%s expected=v2", v)
	}

	tb = Table{
		"k4": "v4",
	}
	if err := dbw.sync2KeyTable(tablename, tb); err != nil {
		t.Errorf("%+v", err)
	}
	v, err = dbw.getValFromKeyTable(tablename, "k4")
	if err != nil {
		t.Errorf("%+v", err)
	}
	if v != "v4" {
		t.Errorf("Unexpected res from getValFromKeyTable(). Got=%s expected=v4", v)
	}
	tb2, err = dbw.syncFromKeyTable(tablename)
	if err != nil {
		t.Errorf("%+v", err)
	}
	if _, ok := tb2["k1"]; ok {
		t.Errorf("Not expected that key k1 is in the table.")
	}
	if v, ok := tb2["k4"]; !ok {
		t.Errorf("Key k4 must be there.")
	} else if v != "v4" {
		t.Errorf("Wrong val for key k4. Got=%s Expected=v4", v)
	}
	dbw.closeDb()
	dbw.deleteDb()
}
