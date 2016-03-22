package go_flagz

import (
	"testing"
	flag "github.com/spf13/pflag"

	"github.com/stretchr/testify/assert"
)

func TestDynInt64_SetAndGet(t *testing.T) {
	set := flag.NewFlagSet("foobar", flag.ContinueOnError)
	dynFlag := DynInt64(set, "some_int_1", 13371337, "Use it or lose it")
	assert.Equal(t, int64(13371337), dynFlag.Get(), "value must be default after create")
	err := set.Set("some_int_1", "77007700")
	assert.NoError(t, err, "setting value must succeed")
	assert.Equal(t, int64(77007700), dynFlag.Get(), "value must be set after update")
}

func TestDynInt64_IsMarkedDynamic(t *testing.T) {
	set := flag.NewFlagSet("foobar", flag.ContinueOnError)
	DynInt64(set, "some_int_1", 13371337, "Use it or lose it")
	assert.True(t, IsFlagDynamic(set.Lookup("some_int_1")))
}

func TestDynInt64_FiresValidators(t *testing.T) {
	set := flag.NewFlagSet("foobar", flag.ContinueOnError)
	DynInt64(set, "some_int_1", 13371337, "Use it or lose it").WithValidator(ValidateDynInt64Range(0, 2000))

	assert.NoError(t, set.Set("some_int_1", "300"), "no error from validator when in range")
	assert.NoError(t, set.Set("some_int_1", "2001"), "error from validator when value out of range")
}
