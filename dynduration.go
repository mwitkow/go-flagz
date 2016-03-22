package go_flagz

import (
	"fmt"
	"github.com/spf13/pflag"
	"sync/atomic"
	"time"
)

func DynDuration(flagSet *pflag.FlagSet, name string, value time.Duration, usage string) *DynDurationValue {
	dynValue := &DynDurationValue{ptr: (*int64)(&value)}
	flag := flagSet.VarPF(dynValue, name, "", usage)
	setFlagDynamic(flag)
	return dynValue
}

type DynDurationValue struct {
	ptr       *int64
	validator func(time.Duration) error
}

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

func (d *DynDurationValue) WithValidator(validator func(time.Duration) error) {
	d.validator = validator
}

func (d *DynDurationValue) Type() string {
	return "dyn_duration"
}

func (d *DynDurationValue) Get() time.Duration {
	return (time.Duration)(atomic.LoadInt64(d.ptr))
}

func (d *DynDurationValue) String() string {
	return fmt.Sprintf("%v", d.Get())
}
