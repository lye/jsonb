package jsonb

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
)

// List is an abstraction around a JSON array. It provides read-only access
// to a blob of JSON; the blob is unmarshalled to an []interface{} on demand.
// To write values to a List, first cast it to a typed MutableList via As,
// then use the MutableList instead.
//
// Lists are basically for receiving data from PostgreSQL, which may or may
// not have the structure you want. Marshalling a List to JSON is basically
// a no-op, as the JSON is already cached.
type List struct {
	raw     json.RawMessage
	decoded []interface{}
}

// MutableList is a type-checked list that can have values appended to it
// via the Append method. Values are type-checked against the type definition.
type MutableList struct {
	List
	ty *Type
}

var _ sql.Scanner = &List{}
var _ driver.Valuer = &List{}

// NewList returns a newly constructed MutableList with the given type.
func NewList(ty *Type) *MutableList {
	return &MutableList{
		ty: ty,
		List: List{
			decoded: []interface{}{},
		},
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

// AsUnsafe creates a MutableList from a List, but does no type checking.
func (l *List) AsUnsafe(ty *Type) *MutableList {
	return &MutableList{
		List: *l,
		ty:   ty,
	}
}

// As creates a MutableList from a List, and type-checks every value to ensure
// that it's what's expected.
func (l *List) As(ty *Type) (*MutableList, error) {
	dec, er := l.decode()
	if er != nil {
		return nil, er
	}

	if !ty.IsValid(dec) {
		return nil, ErrSchema
	}

	return l.AsUnsafe(ty), nil
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

func (l List) Value() (driver.Value, error) {
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

	// NOTE: We could set .raw here but since .decoded is set, the current
	// implementation would just overwrite .raw. This could be made more
	// efficient (by storing both until the next .decode is called) but
	// realistically I doubt it'll matter.
	l.decoded = val
	return nil
}

func (ml *MutableList) UnmarshalJSON(bs []byte) error {
	if er := ml.List.UnmarshalJSON(bs); er != nil {
		return er
	}

	if !ml.ty.IsValid(ml.decoded) {
		ml.decoded = nil
		return ErrSchema
	}

	return nil
}

// Append inserts val onto the end of the MutableList. It returns an error if
// there's a type issue.
func (ml *MutableList) Append(val interface{}) error {
	// NOTE: It could be an optimization here to serializing val and appending
	// it directly to .raw if .decoded is nil. Seems like overkill.
	if !ml.ty.ListType.IsValid(val) {
		return ErrSchema
	}

	_, er := ml.decode()
	if er != nil {
		return er
	}

	if len(ml.decoded) == ml.ty.MaxLen {
		return ErrSchema
	}

	ml.decoded = append(ml.decoded, val)
	return nil
}

// Values returns the underlying Go values for the list as a []interface{}.
// Note that, if the MutableList is created with AsUnsafe, the values may have
// arbitrary types.
func (ml *MutableList) Values() []interface{} {
	return ml.decoded
}

// Int64Values returns the list as an []int64. The list must only contain
// numeric values. Non-integer numeric values are truncated.
func (ml *MutableList) Int64Values() (out []int64, er error) {
	for _, ival := range ml.decoded {
		switch val := ival.(type) {
		case int:
			out = append(out, int64(val))
		case int32:
			out = append(out, int64(val))
		case int64:
			out = append(out, int64(val))
		case float64:
			out = append(out, int64(val))
			// XXX: Probably need more cases here.
		default:
			return nil, ErrUnexpectedType
		}
	}

	return
}

// StringValues returns the list as a []string.
func (ml *MutableList) StringValues() (out []string, er error) {
	for _, ival := range ml.decoded {
		switch val := ival.(type) {
		case string:
			out = append(out, val)
			// XXX: Can technically convert non-strings to strings.
		default:
			return nil, ErrUnexpectedType
		}
	}

	return
}
