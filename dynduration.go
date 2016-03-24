// Copyright 2015 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package flagz

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/spf13/pflag"
)


// DynDuration creates a `Flag` that represents `time.Duration` which is safe to change dynamically at runtime.
func DynDuration(flagSet *pflag.FlagSet, name string, value time.Duration, usage string) *DynDurationValue {
	dynValue := &DynDurationValue{ptr: (*int64)(&value)}
	flag := flagSet.VarPF(dynValue, name, "", usage)
	setFlagDynamic(flag)
	return dynValue
}

// DynDurationValue is a flag-related `time.Duration` value wrapper.
type DynDurationValue struct {
	ptr       *int64
	validator func(time.Duration) error
}

// Get retrieves the value in a thread-safe manner.
func (d *DynDurationValue) Get() time.Duration {
	return (time.Duration)(atomic.LoadInt64(d.ptr))
}

// Set updates the value from a string representation in a thread-safe manner.
// This operation may return an error if the provided `input` doesn't parse, or the resulting value doesn't pass an
// optional validator.
// If a notifier is set on the value, it will be invoked in a separate go-routine.
func (d *DynDurationValue) Set(input string) error {
	v, err := time.ParseDuration(input)
	if err != nil {
		return err
	}
	if d.validator != nil {
		if err := d.validator(v); err != nil {
			return err
		}
	}
	atomic.StoreInt64(d.ptr, (int64)(v))
	return nil
}

// WithValidator adds a function that checks values before they're set.
// Any error returned by the validator will lead to the value being rejected.
// Validators are executed on the same go-routine as the call to `Set`.
func (d *DynDurationValue) WithValidator(validator func(time.Duration) error) {
	d.validator = validator
}

// Type is an indicator of what this flag represents.
func (d *DynDurationValue) Type() string {
	return "dyn_duration"
}

// String represents the canonical representation of the type.
func (d *DynDurationValue) String() string {
	return fmt.Sprintf("%v", d.Get())
}
