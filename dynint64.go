// Copyright 2015 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package flagz

import (
	"fmt"
	"strconv"
	"sync/atomic"

	"github.com/spf13/pflag"
)

func DynInt64(flagSet *pflag.FlagSet, name string, value int64, usage string) *DynInt64Value {
	dynValue := &DynInt64Value{ptr: &value}
	flag := flagSet.VarPF(dynValue, name, "", usage)
	setFlagDynamic(flag)
	return dynValue
}

type DynInt64Value struct {
	ptr       *int64
	validator func(int64) error
}

func (d *DynInt64Value) Set(s string) error {
	val, err := strconv.ParseInt(s, 0, 64)
	if err != nil {
		return err
	}
	if d.validator != nil {
		if err := d.validator(val); err != nil {
			return err
		}
	}
	atomic.StoreInt64(d.ptr, val)
	return nil
}

func (d *DynInt64Value) WithValidator(validator func(int64) error) {
	d.validator = validator
}

func (d *DynInt64Value) Type() string {
	return "dyn_int64"
}

func (d *DynInt64Value) Get() int64 {
	return atomic.LoadInt64(d.ptr)
}

func (d *DynInt64Value) String() string {
	return fmt.Sprintf("%v", d.Get())
}

// ValidateDynInt64Range returns a validator function that checks if the flag value is in range.
func ValidateDynInt64Range(fromInclusive int64, toInclusive int64) func(int64) error {
	return func(value int64) error {
		if value > toInclusive || value < fromInclusive {
			return fmt.Errorf("value %v not in [%v, %v] range", value, fromInclusive, toInclusive)
		}
		return nil
	}
}
