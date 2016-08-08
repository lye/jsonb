package jsonb

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
)

type List struct {
	ty      *Type
	raw     json.RawMessage
	decoded []interface{}
}

var _ sql.Scanner = &List{}
var _ driver.Valuer = &List{}

func NewList(ty *Type) *List {
	return &List{
		ty:      ty,
		decoded: []interface{}{},
	}
}

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
		// XXX: See note in Table.Scan.
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

func (l *List) MarshalJSON() ([]byte, error) {
	return l.encode()
}

func (l *List) UnmarshalJSON(bs []byte) error {
	var val []interface{}

	if er := json.Unmarshal(bs, &val); er != nil {
		return er
	}

	if !l.ty.IsValid(val) {
		return ErrInvalidJsonType
	}

	// NOTE: We could set .raw here but since .decoded is set, the current
	// implementation would just overwrite .raw. This could be made more
	// efficient (by storing both until the next .decode is called) but
	// realistically I doubt it'll matter.
	l.decoded = val
	return nil
}

func (l *List) Append(val interface{}) error {
	// NOTE: It could be an optimization here to serializing val and appending
	// it directly to .raw if .decoded is nil. Seems like overkill.
	if !l.ty.ListType.IsValid(val) {
		return ErrSchema
	}

	_, er := l.decode()
	if er != nil {
		return er
	}

	if len(l.decoded) == l.ty.MaxLen {
		return ErrSchema
	}

	l.decoded = append(l.decoded, val)
	return nil
}
