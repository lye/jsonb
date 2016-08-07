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

func TestNewList(t *testing.T) {
	ty := NewListType(&TypeNumber, 10)
	l := NewList(ty)

	bs, er := json.Marshal(l)
	if er != nil {
		t.Fatal(er)
	}

	if string(bs) != "[]" {
		t.Errorf("wrong serialization %s", string(bs))
	}
}

func TestNumberListAppend(t *testing.T) {
	ty := NewListType(&TypeNumber, 10)
	l := NewList(ty)

	if er := l.Append(0); er != nil {
		t.Error(er)
	}
	if er := l.Append(1); er != nil {
		t.Error(er)
	}
	if er := l.Append(2); er != nil {
		t.Error(er)
	}
	if er := l.Append("hi"); er == nil {
		t.Error("append should have failed")
	}

	if len(l.decoded) != 3 {
		t.Errorf("internal state broken %#v", l.decoded)
	}

	bs, er := json.Marshal(l)
	if er != nil {
		t.Fatal(er)
	}

	if string(bs) != "[0,1,2]" {
		t.Errorf("wrong serialization %s", string(bs))
	}

	l2 := NewList(ty)
	if er := json.Unmarshal(bs, l2); er != nil {
		t.Fatal(er)
	}

	dec, er := l2.decode()
	if er != nil {
		t.Fatal(er)
	}

	if len(dec) != 3 {
		t.Fatal("len(dec) wrong")
	}

	if dec[0].(float64) != 0 || dec[1].(float64) != 1 || dec[2].(float64) != 2 {
		t.Errorf("failed to unmarshal %#v", dec)
	}
}

func TestListAppendOOB(t *testing.T) {
	ty := NewListType(&TypeNumber, 1)
	l := NewList(ty)

	if er := l.Append(0); er != nil {
		t.Error("first append failed")
	}
	if er := l.Append(1); er == nil {
		t.Error("second append succeeded")
	}

	if len(l.decoded) != 1 {
		t.Error("internal state corrupt")
	}
}

func TestListCoerceGood(t *testing.T) {
	l := List{
		raw: json.RawMessage(`[1,2]`),
	}
	ty := NewListType(&TypeNumber, 2)

	if _, er := l.As(ty); er != nil {
		t.Fatal(er)
	}
}

func TestListCoerceBad(t *testing.T) {
	tb := Table{
		raw: json.RawMessage(`[1,2]`),
	}
	ty := NewListType(&TypeString, 2)

	if _, er := tb.As(ty); er == nil {
		t.Fatal(er)
	}
}
