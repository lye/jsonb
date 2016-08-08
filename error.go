package jsonb

import (
	"errors"
)

var (
	// ErrNoSuchKind is returned when attempting to marshal a string
	// to a kind that isn't in the table.
	ErrNoSuchKind = errors.New("jsonb: no such kind")

	// ErrSchema is returned by operations modifying Table/Lists wherein
	// the operation is prohibited by the structure's type.
	ErrSchema = errors.New("jsonb: schema prohibits this operation")

	// ErrInvalidScanType is emitted when something terrible happens while
	// the type is being read from an SQL result. The drivers typically
	// wrap this error, so it's not very useful.
	ErrInvalidScanType = errors.New("jsonb: invalid scan type")

	// ErrInvalidJsonType is emitted when something goes wrong when
	// unmarshaling from JSON. This can happen while coercing from SQL
	// results -- some rudimentary checking is done (currently, XXX that
	// seems like it might cause data to be locked in the db).
	ErrInvalidJsonType = errors.New("jsonb: invalid json type")
)
