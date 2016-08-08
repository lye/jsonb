package jsonb

import (
	"encoding/json"
	"testing"
)

func TestListUnmarshal(t *testing.T) {
	db := testGetDb(t)
	defer db.Close()

	rows, er := db.Query(`SELECT '["one", "two"]'::jsonb`)
	if er != nil {
		t.Fatal(er)
	}
	defer rows.Close()

	if !rows.Next() {
		t.Fatal("no rows")
	}

	val := List{}
	if er := rows.Scan(&val); er != nil {
		t.Fatal(er)
	}
}

func TestListMarshal(t *testing.T) {
	db := testGetDb(t)
	defer db.Close()

	val := List{
		// NOTE: format here matches pgsql since we're doing raw
		// string compares (nasty).
		raw: json.RawMessage(`["one", "two"]`),
	}

	rows, er := db.Query(`SELECT $1::jsonb`, &val)
	if er != nil {
		t.Fatal(er)
	}
	defer rows.Close()

	if !rows.Next() {
		t.Fatal("no rows")
	}

	val2 := List{}
	if er := rows.Scan(&val2); er != nil {
		t.Fatal(er)
	}

	if string(val.raw) != string(val2.raw) {
		t.Errorf("strings differ\n%s\n%s", val.raw, val2.raw)
	}
}

func TestListWrongType(t *testing.T) {
	db := testGetDb(t)
	defer db.Close()

	rows, er := db.Query(`SELECT '{"one":"two"}'::jsonb`)
	if er != nil {
		t.Fatal(er)
	}
	defer rows.Close()

	if !rows.Next() {
		t.Fatal("no rows")
	}

	val := List{}
	er = rows.Scan(&val)
	if er == nil {
		t.Fatal("no error")
	}
}
