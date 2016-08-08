package jsonb

type Kind int

const (
	KindTable Kind = iota
	KindList
	KindNumber
	KindString
	KindAny
)

var kindStrings = map[Kind]string{
	KindTable:  `"table"`,
	KindList:   `"list"`,
	KindNumber: `"number"`,
	KindString: `"string"`,
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

type TableDef map[string]*Type

type Type struct {
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

func NewStringType(maxLen int) *Type {
	return &Type{
		Kind:   KindString,
		MaxLen: maxLen,
	}
}

func NewTableType(fields TableDef) *Type {
	return &Type{
		Kind:   KindTable,
		Fields: fields,
	}
}

func NewListType(ty *Type, maxLen int) *Type {
	return &Type{
		Kind:     KindList,
		ListType: ty,
		MaxLen:   maxLen,
		Fields:   ty.Fields,
	}
}

var (
	TypeNumber = Type{Kind: KindNumber}
	TypeString = Type{Kind: KindString}
	TypeAny    = Type{Kind: KindAny}
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
		return false

	case KindString:
		s, ok := val.(string)
		return ok && (ty.MaxLen <= 0 || ty.MaxLen >= len(s))

	case KindAny:
		return true
	}

	panic("unreachable")
}
