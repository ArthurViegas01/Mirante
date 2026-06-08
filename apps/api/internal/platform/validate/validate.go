// Package validate wraps go-playground/validator with a single shared instance.
// Domain enums are validated here via `oneof` tags so each domain does not
// re-implement enum checks.
package validate

import "github.com/go-playground/validator/v10"

var v = validator.New(validator.WithRequiredStructEnabled())

// Struct validates a struct's `validate` tags.
func Struct(s any) error { return v.Struct(s) }

// Var validates a single value against a tag (e.g. "oneof=a b c").
func Var(field any, tag string) error { return v.Var(field, tag) }

// Validator exposes the underlying instance for custom registrations.
func Validator() *validator.Validate { return v }
