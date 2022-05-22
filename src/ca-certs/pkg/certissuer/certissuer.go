// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package certissuer

import (
	"fmt"
	"strings"
)

func (r *ResourceStatus) ResourceName() string {
	return strings.ToLower(fmt.Sprintf("%s+%s", r.Name, r.Kind))
}
