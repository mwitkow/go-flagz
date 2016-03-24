// Copyright 2015 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package flagz

import (
	"fmt"
	"regexp"
	"sync/atomic"
	"unsafe"

	"github.com/spf13/pflag"
)

// DynString creates a `Flag` that represents `string` which is safe to change dynamically at runtime.
func DynString(flagSet *pflag.FlagSet, name string, value string, usage string) *DynStringValue {
	dynValue := &DynStringValue{ptr: unsafe.Pointer(&value)}
	flag := flagSet.VarPF(dynValue, name, "", usage)
	setFlagDynamic(flag)
	return dynValue
}

// DynStringValue is a flag-related `time.Duration` value wrapper.
type DynStringValue struct {
	ptr       unsafe.Pointer
	validator func(string) error
}

// Get retrieves the value in a thread-safe manner.
func (d *DynStringValue) Get() string {
	p := (*string)(atomic.LoadPointer(&d.ptr))
	return *p
}

// Set updates the value from a string representation in a thread-safe manner.
// This operation may return an error if the provided `input` doesn't parse, or the resulting value doesn't pass an
// optional validator.
// If a notifier is set on the value, it will be invoked in a separate go-routine.
func (d *DynStringValue) Set(val string) error {
	if d.validator != nil {
		if err := d.validator(val); err != nil {
			return err
		}
	}
	atomic.StorePointer(&d.ptr, unsafe.Pointer(&val))
	return nil
}

// WithValidator adds a function that checks values before they're set.
// Any error returned by the validator will lead to the value being rejected.
// Validators are executed on the same go-routine as the call to `Set`.
func (d *DynStringValue) WithValidator(validator func(string) error) {
	d.validator = validator
}

// Type is an indicator of what this flag represents.
func (d *DynStringValue) Type() string {
	return "dyn_string"
}

// String represents the canonical representation of the type.
func (d *DynStringValue) String() string {
	return fmt.Sprintf("%v", d.Get())
}

// ValidateDynStringMatchesRegex returns a validator function that checks all flag's values against regex.
func ValidateDynStringMatchesRegex(matcher *regexp.Regexp) func(string) error {
	return func(value string) error {
		if !matcher.MatchString(value) {
			return fmt.Errorf("value %v must match regex %v", value, matcher)
		}
		return nil
	}
}
