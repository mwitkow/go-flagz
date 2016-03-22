package flagz

import (
	"fmt"
	"strconv"
	"sync/atomic"
	"github.com/spf13/pflag"
	"unsafe"
)


func DynFloat64(flagSet *pflag.FlagSet, name string, value float64, usage string) *DynFloat64Value {
	dynValue := &DynFloat64Value{ptr: unsafe.Pointer(&value)}
	flag := flagSet.VarPF(dynValue, name, "", usage)
	setFlagDynamic(flag)
	return dynValue
}

type DynFloat64Value struct {
	ptr       unsafe.Pointer
	validator func(float64) error
}

func (d *DynFloat64Value) Set(s string) error {
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}
	if d.validator != nil {
		if err := d.validator(val); err != nil {
			return err
		}
	}
	atomic.StorePointer(&d.ptr, unsafe.Pointer(&val))
	return nil
}

func (d *DynFloat64Value) WithValidator(validator func(float64) error) {
	d.validator = validator
}

func (d *DynFloat64Value) Type() string {
	return "dyn_float64"
}

func (d *DynFloat64Value) Get() float64 {
	p := (*float64)(atomic.LoadPointer(&d.ptr))
	return *p
}

func (d *DynFloat64Value) String() string {
	return fmt.Sprintf("%v", d.Get())
}


// ValidateDynFloat64Range returns a validator function that checks if the flag value is in range.
func ValidateDynFloat64Range(fromInclusive float64, toInclusive float64) func (float64) error {
	return func(value float64) error {
		if value > toInclusive || value < fromInclusive {
			return fmt.Errorf("value %v not in [%v, %v] range", value, fromInclusive, toInclusive)
		}
		return nil
	}
}
