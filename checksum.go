// Copyright 2015 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package flagz

import (
	"hash/fnv"

	"github.com/spf13/pflag"
)

// ChecksumFlagSet will generate a FNV of the *set* values in a FlagSet.
func ChecksumFlagSet(flagSet *pflag.FlagSet, flagFilter func(flag *pflag.Flag) bool) []byte {
	h := fnv.New32a()
	flagSet.VisitAll(func(flag *pflag.Flag) {
		if flagFilter != nil && !flagFilter(flag) {
			return
		}
		h.Write([]byte(flag.Name))
		h.Write([]byte(flag.Value.String()))
	})
	return h.Sum(nil)
}
