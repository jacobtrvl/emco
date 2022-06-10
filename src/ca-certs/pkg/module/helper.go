// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package module

import (
	"fmt"
	"strings"
)

// ResourceName
func ResourceName(name, kind string) string {
	return strings.ToLower(fmt.Sprintf("%s+%s", name, kind))
}
