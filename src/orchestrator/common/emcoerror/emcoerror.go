// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package emcoerror

type ErrorType int

const (
	BadRequest ErrorType = iota
	Conflict
	DataError
	NotFound
	Unknown
)

// Error defines the emco errors
type Error struct {
	error
	Message string
	Type    ErrorType
	Cause   error
	// Add any additional data

}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Cause != nil {
		return e.Message + e.Cause.Error()
	}

	return e.Message
}
