// Copyright 2015 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package flagz

import (
	"testing"

	flag "github.com/spf13/pflag"

	"regexp"

	"github.com/stretchr/testify/assert"
)

func TestDynString_SetAndGet(t *testing.T) {
	set := flag.NewFlagSet("foobar", flag.ContinueOnError)
	dynFlag := DynString(set, "some_string_1", "something", "Use it or lose it")
	assert.Equal(t, "something", dynFlag.Get(), "value must be default after create")
	err := set.Set("some_string_1", "else")
	assert.NoError(t, err, "setting value must succeed")
	assert.Equal(t, "else", dynFlag.Get(), "value must be set after update")
}

func TestDynString_IsMarkedDynamic(t *testing.T) {
	set := flag.NewFlagSet("foobar", flag.ContinueOnError)
	DynString(set, "some_string_1", "somethign", "Use it or lose it")
	assert.True(t, IsFlagDynamic(set.Lookup("some_string_1")))
}

func TestDynString_FiresValidators(t *testing.T) {
	set := flag.NewFlagSet("foobar", flag.ContinueOnError)
	regex := regexp.MustCompile("^[a-z]{2,8}$")
	DynString(set, "some_string_1", "something", "Use it or lose it").WithValidator(ValidateDynStringMatchesRegex(regex))

	assert.NoError(t, set.Set("some_string_1", "else"), "no error from validator when in range")
	assert.Error(t, set.Set("some_string_1", "else1"), "error from validator when value out of range")
}

func Benchmark_String_Dyn_Get(b *testing.B) {
	set := flag.NewFlagSet("foobar", flag.ContinueOnError)
	value := DynString(set, "some_string_1", "something", "Use it or lose it")
	set.Set("some_string_1", "else")
	for i := 0; i < b.N; i++ {
		x := value.Get()
		x = x + "foo"
	}
}

func Benchmark_String_Normal_get(b *testing.B) {
	set := flag.NewFlagSet("foobar", flag.ContinueOnError)
	valPtr := set.String("some_string_1", "something", "Use it or lose it")
	set.Set("some_string_1", "else")
	for i := 0; i < b.N; i++ {
		x := *valPtr
		x = x + "foo"
	}
}
