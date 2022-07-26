// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package emcoerror

import (
	"net/http"
)

// Error implements error interface and glues together everything to do with erroring, including the stack of error causes
type Error struct {
	error
	Kind  ErrorKind // Kind ErrorKind is what glues messages (which I guess can be reused for logging purposes) and HTTP status codes
	Cause *Error    // support tracing of causes with complete detail, including how messages have changed throughout the call stack
}

// ErrorKind describes a unique error in EMCO, with a message and the to-become HTTP status code (if it ever reaches the API layer),
// by itself, this ErrorKind doesn't mean much - see below for more info.
type ErrorKind struct {
	Message        string
	HTTPStatusCode int
}

// Just adapted this function to match the new Error struct...
// Error implements the error interface
func (e *Error) Error() string {
	if e.Cause != nil {
		return e.Kind.Message + e.Cause.Error()
	}

	return e.Kind.Message
}

// This is where ErrorKind gathers meaning, by creating unique instances of it that we (developers) can use to simplify error creation and handling.
// It's basically an enum of structs - this is essentially as lightweight as a typical int enum (much lighter than string comparison),
// while providing the benefits of easily referencing error kinds by name and mapping them to default messages (which can be overriden)
// and to HTTP status codes (if it gets to the API level).

// And we can visually separate the definition of errors, one per controller/service:

// general errors
var (
	InternalServerError = ErrorKind{"EMCO is not feeling well, please try again later", http.StatusInternalServerError}
	UnknownError        = ErrorKind{"Unknown error", http.StatusInternalServerError}
	StateInfoNotFound   = ErrorKind{"State info not found", http.StatusConflict}
	GeneralConflict     = ErrorKind{"General conflict", http.StatusConflict}
)

// caCert errors
var (
	CaCertAlreadyExists             = ErrorKind{"caCert already exists", http.StatusConflict}
	CaCertNotFound                  = ErrorKind{"caCert not found", http.StatusNotFound}
	CaCertClusterGroupAlreadyExists = ErrorKind{"caCert cluster group already exists", http.StatusConflict}
	CaCertClusterGroupNotFound      = ErrorKind{"caCert cluster group not found", http.StatusNotFound}
	CaCertLogicalCloudAlreadyExists = ErrorKind{"caCert logical cloud already exists", http.StatusConflict}
	CaCertLogicalCloudNotFound      = ErrorKind{"caCert logical cloud not found", http.StatusNotFound}
	CaKeyNotFound                   = ErrorKind{"certificate key not found", http.StatusBadRequest}
)

// (We could think of ways of pulling these definitions straight from the controller/service code
// but for now yes this is defining all possible EMCO error codes in a central common place).
// Another benefiting of having it all like this is that this file alone will act as a developer
// guide for error handling in EMCO - great for ramping up people and writing less documentation.
