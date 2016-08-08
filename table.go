package jsonb

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
)

type Table struct {
	def     *TableDef
	raw     json.RawMessage
	decoded map[string]interface{}
}

var _ sql.Scanner = &Table{}
var _ driver.Valuer = &Table{}

func (t *Table) decode() (map[string]interface{}, error) {
	if t.decoded == nil {
		if er := json.Unmarshal(t.raw, &t.decoded); er != nil {
			return nil, er
		}
	}

	return t.decoded, nil
}

func (t *Table) encode() (_ json.RawMessage, er error) {
	if t.decoded != nil {
		if t.raw, er = json.Marshal(t.decoded); er != nil {
			return t.raw, er
		}

		t.decoded = nil
	}

	return t.raw, nil
}

func (t *Table) Scan(src interface{}) error {
	bs, ok := src.([]byte)
	if !ok {
		return ErrInvalidScanType
	}

	if len(bs) == 0 || bs[0] != '{' {
		return ErrInvalidJsonType
	}

	t.raw = json.RawMessage(bs)
	t.decoded = nil
	return nil
}

func (t *Table) Value() (driver.Value, error) {
	raw, er := t.encode()
	return []byte(raw), er
}
