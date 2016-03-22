package go_flagz

import (
	"encoding/csv"
	"fmt"
	"strings"
	"sync/atomic"
	"unsafe"

	"github.com/spf13/pflag"
)

func DynStringSlice(flagSet *pflag.FlagSet, name string, value []string, usage string) *DynStringSliceValue {
	dynValue := &DynStringSliceValue{ptr: unsafe.Pointer(&value)}
	flag := flagSet.VarPF(dynValue, name, "", usage)
	setFlagDynamic(flag)
	return dynValue
}

type DynStringSliceValue struct {
	ptr       unsafe.Pointer
	validator func([]string) error
}

func (d *DynStringSliceValue) Set(val string) error {
	v, err := csv.NewReader(strings.NewReader(val)).Read()
	if err != nil {
		return err
	}
	if d.validator != nil {
		if err := d.validator(v); err != nil {
			return err
		}
	}
	atomic.StorePointer(&d.ptr, unsafe.Pointer(&v))
	return nil
}

func (d *DynStringSliceValue) WithValidator(validator func([]string) error) {
	d.validator = validator
}

func (d *DynStringSliceValue) Type() string {
	return "dyn_stringslice"
}

func (d *DynStringSliceValue) Get() []string {
	p := (*[]string)(atomic.LoadPointer(&d.ptr))
	return *p
}

func (d *DynStringSliceValue) String() string {
	return fmt.Sprintf("%v", d.Get())
}

// ValidateDynStringSliceMinElements validates that the given string slice has at least x elements.
func ValidateDynStringSliceMinElements(count int) func([]string) error {
	return func(value []string) error {
		if len(value) < count {
			return fmt.Errorf("value slice %v must have at least %v elements", value, count)
		}
		return nil
	}
}
