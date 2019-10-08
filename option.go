package portal

import "github.com/pkg/errors"

type Option func(c *Chell) error

// Only specifies the fields to keep.
// Examples:
// ```
// c := New(Only("A")) // keep field A only
// c := New("A[B,C]") // // keep field B and C of the nested struct A
// ```
func Only(fields ...string) Option {
	return func(c *Chell) error {
		filters, err := parseFilters(fields)
		if err != nil {
			return errors.WithStack(err)
		}
		c.onlyFieldFilters = filters
		return nil
	}
}

// Exclude specifies the fields to exclude.
// Examples:
// ```
// c := New(Exclude("A")) // exclude field A
// c := New(Exclude("A[B,C]")) // exclude field B and C of the nested struct A, but other fields of struct A are still selected.
// ```
func Exclude(fields ...string) Option {
	return func(c *Chell) error {
		filters, err := parseFilters(fields)
		if err != nil {
			return errors.WithStack(err)
		}
		c.excludeFieldFilters = filters
		return nil
	}
}

// FieldAliasMapTagName sets the tag name (e.g. `yaml`, `json`) to parse alias of a field name.
// Example:
// ```
// struct Schema {
//     ID   int `json:"id"`
// }
//
// // portal parses the json tag, and maps `id` -> `ID`.
// ```
func FieldAliasMapTagName(tag string) Option {
	return func(c *Chell) error {
		c.fieldAliasMapTagName = tag
		return nil
	}
}
