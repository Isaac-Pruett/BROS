//
// Copyright (c) 2025 ZettaScale Technology
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0 which is available at
// http://www.eclipse.org/legal/epl-2.0, or the Apache License, Version 2.0
// which is available at https://www.apache.org/licenses/LICENSE-2.0.
//
// SPDX-License-Identifier: EPL-2.0 OR Apache-2.0
//
// Contributors:
//   ZettaScale Zenoh Team, <zenoh@zettascale.tech>
//

package zenoh

// #include "zenoh.h"
import "C"
import (
	"fmt"
	_ "unsafe" // for go:linkname
)

type ZError struct {
	code int8
	msg  string
}

func (e ZError) Error() string { return fmt.Sprintf("%s (Error Code: %d)", e.msg, e.code) }

//go:linkname newZError
func newZError(code int8) ZError {
	var viewString C.z_view_string_t
	C.zc_get_last_error(&viewString)
	loanedString := C.z_view_string_loan(&viewString)
	msg := C.GoStringN(C.z_string_data(loanedString), C.int(C.z_string_len(loanedString)))
	return ZError{code: code, msg: msg}
}
