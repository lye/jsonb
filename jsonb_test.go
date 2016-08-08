package jsonb

import (
	"encoding/json"
	"testing"
)

func TestJsonbUnmarshal(t *testing.T) {
	db := testGetDb(t)
	defer db.Close()

	rows, er := db.Query("SELECT '[1,2,3]'::jsonb")
	if er != nil {
		t.Fatal(er)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			jsv  json.RawMessage
			vals []int64
		)

		if er := rows.Scan(&jsv); er != nil {
			t.Fatal(er)
		}

		if er := json.Unmarshal(jsv, &vals); er != nil {
			t.Fatal(er)
		}

		if len(vals) != 3 {
			t.Fatalf("bad data %#v", vals)
		}
	}
}
