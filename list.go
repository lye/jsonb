package jsonb

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
)

type List struct {
	ty      Type
	raw     json.RawMessage
	decoded []interface{}
}

var _ sql.Scanner = &List{}
var _ driver.Valuer = &List{}

func (l *List) decode() ([]interface{}, error) {
	if l.decoded == nil {
		if er := json.Unmarshal(l.raw, &l.decoded); er != nil {
			return nil, er
		}
	}

	return l.decoded, nil
}

func (l *List) encode() (_ json.RawMessage, er error) {
	if l.decoded != nil {
		if l.raw, er = json.Marshal(l.decoded); er != nil {
			return l.raw, er
		}

		l.decoded = nil
	}

	return l.raw, nil
}

func (l *List) Scan(src interface{}) error {
	bs, ok := src.([]byte)
	if !ok {
		return ErrInvalidScanType
	}

	if len(bs) == 0 || bs[0] != '[' {
		return ErrInvalidJsonType
	}

	l.raw = json.RawMessage(bs)
	l.decoded = nil
	return nil
}

func (l *List) Value() (driver.Value, error) {
	raw, er := l.encode()
	return []byte(raw), er
}
