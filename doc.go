// package jsonb provides wrappers for using PostgreSQL JSONB types.
//
// Using unstructured JSON in a statically-typed language is a bit of a mess;
// the goal of the jsonb package is to provide runtime-defined structure to
// being order to chaos.
//
// Types are defined with the Type struct which can composite primitives from
// the Kind enumeration. The TypeNumberList, for example, specifies a list
// that can only contain numeric values. This allows invalid types coming out
// of the database to be detected and corrected -- List.As(TypeNumberList) will
// return an error.
//
// I should put some example code in here.
package jsonb
