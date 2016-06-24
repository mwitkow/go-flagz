// Copyright 2015 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package flagz

import (
	"testing"
	"time"

	flag "github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestDynStringSet_SetAndGet(t *testing.T) {
	set := flag.NewFlagSet("foobar", flag.ContinueOnError)
	dynFlag := DynStringSet(set, "some_stringslice_1", []string{"foo", "bar"}, "Use it or lose it")
	assert.Equal(t, map[string]struct{}{"foo": struct{}{}, "bar": struct{}{}}, dynFlag.Get(), "value must be default after create")
	err := set.Set("some_stringslice_1", "car,bar")
	assert.NoError(t, err, "setting value must succeed")
	assert.Equal(t, map[string]struct{}{"car": struct{}{}, "bar": struct{}{}}, dynFlag.Get(), "value must be set after update")
}

func TestDynStringSet_IsMarkedDynamic(t *testing.T) {
	set := flag.NewFlagSet("foobar", flag.ContinueOnError)
	DynStringSet(set, "some_stringslice_1", []string{"foo", "bar"}, "Use it or lose it")
	assert.True(t, IsFlagDynamic(set.Lookup("some_stringslice_1")))
}

func TestDynStringSet_FiresValidators(t *testing.T) {
	set := flag.NewFlagSet("foobar", flag.ContinueOnError)
	DynStringSet(set, "some_stringslice_1", []string{"foo", "bar"}, "Use it or lose it").WithValidator(ValidateDynStringSetMinElements(2))

	assert.NoError(t, set.Set("some_stringslice_1", "car,far"), "no error from validator when in range")
	assert.Error(t, set.Set("some_stringslice_1", "car"), "error from validator when value out of range")
}

func TestDynStringSet_FiresNotifier(t *testing.T) {
	waitCh := make(chan struct{}, 1)
	notifier := func(oldVal map[string]struct{}, newVal map[string]struct{}) {
		assert.EqualValues(t, map[string]struct{}{"foo": struct{}{}, "bar": struct{}{}}, oldVal, "old value in notify must match previous value")
		assert.EqualValues(t, map[string]struct{}{"car": struct{}{}, "far": struct{}{}}, newVal, "new value in notify must match set value")
		waitCh <- struct{}{}
	}

	set := flag.NewFlagSet("foobar", flag.ContinueOnError)
	DynStringSet(set, "some_stringslice_1", []string{"foo", "bar"}, "Use it or lose it").WithNotifier(notifier)
	set.Set("some_stringslice_1", "car,far")
	select {
	case <-time.After(5 * time.Millisecond):
		assert.Fail(t, "failed to trigger notifier")
	case <-waitCh:
	}
}
