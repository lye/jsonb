package jsonb

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
)

type Table struct {
	raw     json.RawMessage
	decoded map[string]interface{}
}

type MutableTable struct {
	*Table
	ty *Type
}

var _ sql.Scanner = &Table{}
var _ driver.Valuer = &Table{}

func NewTable(ty *Type) *MutableTable {
	return &MutableTable{
		ty: ty,
		Table: &Table{
			decoded: make(map[string]interface{}),
		},
	}
}

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

func (t *Table) AsUnsafe(ty *Type) *MutableTable {
	_, er := t.decode()
	if er != nil {
		panic(er)
	}

	return &MutableTable{
		Table: t,
		ty:    ty,
	}
}

func (t *Table) As(ty *Type) (*MutableTable, error) {
	dec, er := t.decode()
	if er != nil {
		return nil, er
	}

	if !ty.IsValid(dec) {
		return nil, ErrSchema
	}

	return t.AsUnsafe(ty), nil
}

func (t *Table) Scan(src interface{}) error {
	bs, ok := src.([]byte)
	if !ok {
		return ErrInvalidScanType
	}

	if len(bs) == 0 || bs[0] != '{' {
		// XXX: This might cause some horrible things to happen (e.g. data
		// gets locked in the database and can't be fixed). It makes sense
		// to blow things up in dev, but is there a reasonable way for
		// applications to migrate data? Or is this really a critical failure
		// (that can cause an outage requiring manual database mucking)?
		return ErrInvalidJsonType
	}

	// NB: The buffer *MUST* be copied since it's re-used by pq
	// and will be filled with garbage.
	newSlice := make([]byte, len(bs))
	copy(newSlice, bs)

	t.raw = json.RawMessage(newSlice)
	t.decoded = nil
	return nil
}

func (t *Table) Value() (driver.Value, error) {
	raw, er := t.encode()
	return []byte(raw), er
}

func (t *Table) MarshalJSON() ([]byte, error) {
	return t.encode()
}

func (t *Table) UnmarshalJSON(bs []byte) error {
	var val map[string]interface{}

	if er := json.Unmarshal(bs, &val); er != nil {
		return er
	}

	// NOTE: See notes in List.UnmarshalJSON.
	t.decoded = val
	return nil
}

func (mt *MutableTable) UnmarshalJSON(bs []byte) error {
	if er := mt.Table.UnmarshalJSON(bs); er != nil {
		return er
	}

	if !mt.ty.IsValid(mt.decoded) {
		mt.decoded = nil
		return ErrSchema
	}

	return nil
}

func (mt *MutableTable) Set(key string, val interface{}) error {
	ty, ok := mt.ty.Fields[key]
	if !ok || !ty.IsValid(val) {
		return ErrSchema
	}

	dec, er := mt.decode()
	if er != nil {
		return er
	}

	dec[key] = val
	return nil
}
