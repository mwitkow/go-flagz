// Copyright 2015 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package flagz

import (
	"fmt"
	"sync/atomic"
	"unsafe"

	"encoding/json"
	"reflect"

	"github.com/spf13/pflag"
)

func DynJSON(flagSet *pflag.FlagSet, name string, value interface{}, usage string) *DynJSONValue {
	reflectVal := reflect.ValueOf(value)
	if reflectVal.Kind() != reflect.Ptr || reflectVal.Elem().Kind() != reflect.Struct {
		panic("DynJSON value must be a pointer to a struct")
	}
	dynValue := &DynJSONValue{ptr: unsafe.Pointer(reflectVal.Pointer()), structType: reflectVal.Type().Elem()}
	flag := flagSet.VarPF(dynValue, name, "", usage)
	setFlagDynamic(flag)
	return dynValue
}

type DynJSONValue struct {
	structType reflect.Type
	ptr        unsafe.Pointer
	validator  func(interface{}) error
}

func (d *DynJSONValue) Set(val string) error {
	someStruct := reflect.New(d.structType).Interface()
	if err := json.Unmarshal([]byte(val), someStruct); err != nil {
		return err
	}

	if d.validator != nil {
		if err := d.validator(someStruct); err != nil {
			return err
		}
	}
	atomic.StorePointer(&d.ptr, unsafe.Pointer(reflect.ValueOf(someStruct).Pointer()))
	return nil
}

func (d *DynJSONValue) WithValidator(validator func(interface{}) error) {
	d.validator = validator
}

func (d *DynJSONValue) Type() string {
	return "dyn_json"
}

func (d *DynJSONValue) Get() interface{} {
	p := atomic.LoadPointer(&d.ptr)
	n := reflect.NewAt(d.structType, p)
	return n.Interface()
}

func (d *DynJSONValue) PrettyString() string {
	out, err := json.MarshalIndent(d.Get(), "", "  ")
	if err != nil {
		return "ERR"
	}
	return string(out)
}

func (d *DynJSONValue) String() string {
	return fmt.Sprintf("%v", d.Get())
}
