// Copyright 2015 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package flagz

import (
	"testing"

	flag "github.com/spf13/pflag"

	"github.com/stretchr/testify/assert"
)

func TestDynFloat64_SetAndGet(t *testing.T) {
	set := flag.NewFlagSet("foobar", flag.ContinueOnError)
	dynFlag := DynFloat64(set, "some_float_1", 13.37, "Use it or lose it")
	assert.Equal(t, float64(13.37), dynFlag.Get(), "value must be default after create")
	err := set.Set("some_float_1", "1.337")
	assert.NoError(t, err, "setting value must succeed")
	assert.Equal(t, float64(1.337), dynFlag.Get(), "value must be set after update")
}

func TestDynFloat64_IsMarkedDynamic(t *testing.T) {
	set := flag.NewFlagSet("foobar", flag.ContinueOnError)
	DynFloat64(set, "some_float_1", 13.37, "Use it or lose it")
	assert.True(t, IsFlagDynamic(set.Lookup("some_float_1")))
}

func TestDynFloat64_FiresValidators(t *testing.T) {
	set := flag.NewFlagSet("foobar", flag.ContinueOnError)
	DynFloat64(set, "some_float_1", 13.37, "Use it or lose it").WithValidator(ValidateDynFloat64Range(10.0, 14.0))

	assert.NoError(t, set.Set("some_float_1", "13.41"), "no error from validator when in range")
	assert.Error(t, set.Set("some_float_1", "14.001"), "error from validator when value out of range")
}

func Benchmark_Float64_Dyn_Get(b *testing.B) {
	set := flag.NewFlagSet("foobar", flag.ContinueOnError)
	value := DynFloat64(set, "some_float_1", 13.37, "Use it or lose it")
	set.Set("some_float_1", "14.00")
	for i := 0; i < b.N; i++ {
		x := value.Get()
		x = x + 1
	}
}

func Benchmark_Float64_Normal_Get(b *testing.B) {
	set := flag.NewFlagSet("foobar", flag.ContinueOnError)
	valPtr := set.Float64("some_float_1", 13.37, "Use it or lose it")
	set.Set("some_float_1", "14.00")
	for i := 0; i < b.N; i++ {
		x := *valPtr
		x = x + 0.01
	}
}
