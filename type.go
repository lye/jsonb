package jsonb

type Kind int

const (
	KindTable Kind = iota
	KindList
	KindNumber
	KindString
	KindBool
	KindAny
)

var kindStrings = map[Kind]string{
	KindTable:  `"table"`,
	KindList:   `"list"`,
	KindNumber: `"number"`,
	KindString: `"string"`,
	KindBool:   `"bool"`,
	KindAny:    `"any"`,
}

func (k Kind) MarshalJSON() ([]byte, error) {
	s, ok := kindStrings[k]
	if !ok {
		return nil, ErrNoSuchKind
	}

	return []byte(s), nil
}

func (k *Kind) UnmarshalJSON(bs []byte) error {
	s := string(bs)

	for k1, v := range kindStrings {
		if v == s {
			*k = k1
			return nil
		}
	}

	return ErrNoSuchKind
}

// TableDef is a field name-to-value type mapping. It's used for specifying
// an object's field types statically.
type TableDef map[string]*Type

// Type is a static type definition that declares the expected types encoded
// by a JSON blob. Types should be statically constructed and just used via
// pointer. There are a handful of common pre-defined types.
type Type struct {
	// Kind is the JSON primitive kind. For complex types (e.g. lists/tables)
	// the interior type is defined by ListType/Fields, respectively.
	Kind Kind

	// Set only when Kind is KindList. Contains the list subtype.
	ListType *Type

	// Set when Kind is KindList or KindString. Contains the maximum
	// cardinality of the list/string if > 0.
	MaxLen int

	// Set only when Kind or ListKind is KindTable. References the underlying
	// TableDef which is used for object validation.
	Fields TableDef
}

// NewStringType is a helper method that returns a Type for a string with
// a given max length. If you don't have a max length, just use TypeString
// instead.
func NewStringType(maxLen int) *Type {
	return &Type{
		Kind:   KindString,
		MaxLen: maxLen,
	}
}

// NewTableType is a helper method that returns a table Type with the given
// fields.
func NewTableType(fields TableDef) *Type {
	return &Type{
		Kind:   KindTable,
		Fields: fields,
	}
}

// NewListType is a helper method that returns a new list Type with the
// given type and max length. If maxLen is <=0, the list is unbounded. For
// homogenous primitive lists, consider using one of the existing list
// definitions (e.g. TypeNumberList).
func NewListType(ty *Type, maxLen int) *Type {
	return &Type{
		Kind:     KindList,
		ListType: ty,
		MaxLen:   maxLen,
		Fields:   ty.Fields,
	}
}

var (
	TypeNumber     = &Type{Kind: KindNumber}
	TypeString     = &Type{Kind: KindString}
	TypeBool       = &Type{Kind: KindBool}
	TypeAny        = &Type{Kind: KindAny}
	TypeNumberList = NewListType(TypeNumber, -1)
	TypeStringList = NewListType(TypeString, -1)
	TypeBoolList   = NewListType(TypeBool, -1)
	TypeAnyList    = NewListType(TypeAny, -1)
)

func (ty *Type) isValidList(val interface{}) bool {
	l, ok := val.([]interface{})
	if !ok {
		return false
	}

	if ty.MaxLen > 0 && ty.MaxLen < len(l) {
		return false
	}

	for _, v := range l {
		if !ty.ListType.IsValid(v) {
			return false
		}
	}

	return true
}

func (ty *Type) isValidTable(val interface{}) bool {
	t, ok := val.(map[string]interface{})
	if !ok {
		return false
	}

	for k, v := range t {
		sty, ok := ty.Fields[k]
		if !ok {
			return false
		}

		if !sty.IsValid(v) {
			return false
		}
	}

	return true
}

func (ty *Type) IsValid(val interface{}) bool {
	switch ty.Kind {
	case KindTable:
		return ty.isValidTable(val)

	case KindList:
		return ty.isValidList(val)

	case KindNumber:
		if _, ok := val.(float64); ok {
			return true
		}
		// While we're technically marshalling strictly to json, these
		// easements make it less of a pain to interface with List/Table
		// from the Go side. From most->least likely (via guess).
		if _, ok := val.(int); ok {
			return true
		}
		if _, ok := val.(int64); ok {
			return true
		}
		if _, ok := val.(float32); ok {
			return true
		}
		if _, ok := val.(int32); ok {
			return true
		}
		return false

	case KindString:
		s, ok := val.(string)
		return ok && (ty.MaxLen <= 0 || ty.MaxLen >= len(s))

	case KindBool:
		_, ok := val.(bool)
		return ok

	case KindAny:
		return true
	}

	panic("unreachable")
}
