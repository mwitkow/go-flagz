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

func DynString(flagSet *pflag.FlagSet, name string, value string, usage string) *DynStringValue {
	dynValue := &DynStringValue{ptr: unsafe.Pointer(&value)}
	flag := flagSet.VarPF(dynValue, name, "", usage)
	setFlagDynamic(flag)
	return dynValue
}

type DynStringValue struct {
	ptr       unsafe.Pointer
	validator func(string) error
}

func (d *DynStringValue) Set(val string) error {
	if d.validator != nil {
		if err := d.validator(val); err != nil {
			return err
		}
	}
	atomic.StorePointer(&d.ptr, unsafe.Pointer(&val))
	return nil
}

func (d *DynStringValue) WithValidator(validator func(string) error) {
	d.validator = validator
}

func (d *DynStringValue) Type() string {
	return "dyn_string"
}

func (d *DynStringValue) Get() string {
	p := (*string)(atomic.LoadPointer(&d.ptr))
	return *p
}

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
