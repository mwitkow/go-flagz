// Copyright 2015 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package flagz

import "github.com/spf13/pflag"

const (
	dynamicMarker = "__is_dynamic"
)

func setFlagDynamic(flag *pflag.Flag) {
	if flag.Annotations == nil {
		flag.Annotations = make(map[string][]string)
	}
	flag.Annotations[dynamicMarker] = []string{}
}

// IsFlagDynamic returns whether the given Flag has been created in a Dynamic mode.
func IsFlagDynamic(flag *pflag.Flag) bool {
	_, exists := flag.Annotations[dynamicMarker]
	return exists
}
