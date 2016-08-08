package jsonb

import (
	"encoding/json"
	"testing"
)

func TestTableUnmarshal(t *testing.T) {
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

	val := Table{}
	if er := rows.Scan(&val); er != nil {
		t.Fatal(er)
	}
}

func TestTableMarshal(t *testing.T) {
	db := testGetDb(t)
	defer db.Close()

	val := Table{
		// NOTE: format here matches pgsql since we're doing raw
		// string compares (nasty).
		raw: json.RawMessage(`{"one": "two"}`),
	}

	rows, er := db.Query(`SELECT $1::jsonb`, &val)
	if er != nil {
		t.Fatal(er)
	}
	defer rows.Close()

	if !rows.Next() {
		t.Fatal("no rows")
	}

	val2 := Table{}
	if er := rows.Scan(&val2); er != nil {
		t.Fatal(er)
	}

	if string(val.raw) != string(val2.raw) {
		t.Errorf("strings differ\n%s\n%s", val.raw, val2.raw)
	}
}

func TestTableWrongType(t *testing.T) {
	db := testGetDb(t)
	defer db.Close()

	rows, er := db.Query(`SELECT '["one","two"]'::jsonb`)
	if er != nil {
		t.Fatal(er)
	}
	defer rows.Close()

	if !rows.Next() {
		t.Fatal("no rows")
	}

	val := Table{}
	er = rows.Scan(&val)
	if er == nil {
		t.Fatal("no error")
	}
}

func TestTableSet(t *testing.T) {
	ty := NewTableType(TableDef{
		"k": &TypeString,
		"v": &TypeNumber,
	})
	tb := NewTable(ty)

	if er := tb.Set("k", "foo"); er != nil {
		t.Error(er)
	}
	if er := tb.Set("k", "bar"); er != nil {
		t.Error(er)
	}
	if er := tb.Set("k", 1); er == nil {
		t.Error("expected error")
	}
	if v, ok := tb.decoded["k"]; !ok {
		t.Error("missing key")
	} else {
		if v1, ok := v.(string); !ok {
			t.Error("invalid type")
		} else if v1 != "bar" {
			t.Error("invalid value")
		}
	}

	if er := tb.Set("v", 1); er != nil {
		t.Error(er)
	}
	if er := tb.Set("v", 2); er != nil {
		t.Error(er)
	}
	if er := tb.Set("v", "f"); er == nil {
		t.Error("expected error")
	}
	if v, ok := tb.decoded["v"]; !ok {
		t.Error("missing key")
	} else {
		if v1, ok := v.(int); !ok {
			t.Error("invalid type")
		} else if v1 != 2 {
			t.Error("invalid value")
		}
	}
}
