// Copyright 2015 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package flagz

import flag "github.com/spf13/pflag"

const (
	dynamicMarker = "__is_dynamic"
)

func setFlagDynamic(f *flag.Flag) {
	if f.Annotations == nil {
		f.Annotations = make(map[string][]string)
	}
	f.Annotations[dynamicMarker] = []string{}
}

// IsFlagDynamic returns whether the given Flag has been created in a Dynamic mode.
func IsFlagDynamic(f *flag.Flag) bool {
	_, ok := f.Annotations[dynamicMarker]
	return ok
}
