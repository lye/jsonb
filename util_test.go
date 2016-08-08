package jsonb

import (
	"database/sql"
	_ "github.com/lib/pq"
	"testing"
)

func testGetDb(t *testing.T) *sql.DB {
	db, er := sql.Open("postgres", "sslmode=disable")
	if er != nil {
		t.Fatal(er)
	}

	return db
}
