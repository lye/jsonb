package jsonb

import (
	"encoding/json"
	"testing"
)

func rtJSON(t *testing.T, val interface{}) (out interface{}) {
	bs, er := json.Marshal(val)
	if er != nil {
		t.Fatal(er)
	}

	if er := json.Unmarshal(bs, &out); er != nil {
		t.Fatal(er)
	}

	return out
}

func TestKindMarshal(t *testing.T) {
	bs, er := json.Marshal(KindNumber)
	if er != nil {
		t.Fatal(er)
	}

	if string(bs) != "\"number\"" {
		t.Errorf("invalid json %s", string(bs))
	}

	var k Kind
	if er := json.Unmarshal(bs, &k); er != nil {
		t.Fatal(er)
	}

	if k != KindNumber {
		t.Errorf("invalid kind %d", k)
	}
}

func TestValidListNumber(t *testing.T) {
	v := rtJSON(t, []int{1, 2, 3, 4})
	ty := NewListType(&TypeNumber, 4)
	if !ty.IsValid(v) {
		t.Error("should be valid")
	}
	if !ty.ListType.IsValid(1) {
		t.Error("should be valid")
	}
}

func TestInvalidListNumber(t *testing.T) {
	v := rtJSON(t, []string{"one"})
	ty := NewListType(&TypeNumber, 5)
	if ty.IsValid(v) {
		t.Error("should be invalid")
	}
	if ty.ListType.IsValid("one") {
		t.Error("should be invalid")
	}
}

func TestListMaxLength(t *testing.T) {
	v := rtJSON(t, []int{1, 2, 3, 4})
	ty := NewListType(&TypeNumber, 3)
	if ty.IsValid(v) {
		t.Error("should be invalid")
	}
}

func TestStringMaxLength(t *testing.T) {
	ty := NewStringType(5)
	if ty.IsValid("123456") {
		t.Error("should be invalid")
	}
	if !ty.IsValid("12345") {
		t.Error("should be valid")
	}
	if !ty.IsValid("") {
		t.Error("should be valid")
	}
}

func TestValidListString(t *testing.T) {
	v := rtJSON(t, []string{"one", "two"})
	ty := NewListType(&TypeString, 2)
	if !ty.IsValid(v) {
		t.Error("should be valid")
	}
}

func TestValidTable1(t *testing.T) {
	v := rtJSON(t, map[string]interface{}{
		"one":   "two",
		"three": 4,
		"five":  []int{1, 2, 3},
	})
	ty := NewTableType(map[string]*Type{
		"one":   &TypeString,
		"three": &TypeNumber,
		"five":  NewListType(&TypeNumber, 3),
	})
	if !ty.IsValid(v) {
		t.Error("should be valid")
	}
}

func TestListTable(t *testing.T) {
	v := rtJSON(t, []interface{}{
		map[string]interface{}{
			"v": 1,
		},
		map[string]interface{}{
			"v": 2,
		},
	})
	ty := NewTableType(map[string]*Type{
		"v": &TypeNumber,
	})
	ty = NewListType(ty, 2)

	if !ty.IsValid(v) {
		t.Error("should be valid")
	}

	v = append(v.([]interface{}), map[string]interface{}{
		"v": "foo",
	})
	if ty.IsValid(v) {
		t.Error("should be invalid")
	}
}
