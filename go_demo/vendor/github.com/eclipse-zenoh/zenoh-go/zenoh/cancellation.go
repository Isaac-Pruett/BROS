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

import (
	"runtime"
)

// #include "zenoh.h"
import "C"

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// A Zenoh Cancellation token that can be used to interrupt ongoing queries.
type CancellationToken struct {
	cancellationToken *C.z_owned_cancellation_token_t
}

func cancellationTokenDrop(cancellationToken *C.z_owned_cancellation_token_t) {
	C.z_cancellation_token_drop(C.z_cancellation_token_move(cancellationToken))
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Create a new cancellation token.
func NewCancellationToken() CancellationToken {
	var c C.z_owned_cancellation_token_t
	C.z_cancellation_token_new(&c)
	runtime.SetFinalizer(&c, cancellationTokenDrop)
	return CancellationToken{cancellationToken: &c}
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Interrupt all associated GET queries. If the query callback is being executed,
// the call blocks until execution of callback finishes and its corresponding drop method returns (if any).
// Once token is cancelled, all new associated GET queries will cancel automatically.
func (cancellationToken *CancellationToken) Cancel() error {
	loanedCancellationToken := C.z_cancellation_token_loan_mut(cancellationToken.cancellationToken)
	res := int8(C.z_cancellation_token_cancel(loanedCancellationToken))
	if res != 0 {
		return newZError(res)
	}
	return nil
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Return “true“ if [CancellationToken.Cancel] was called, “false“ otherwise.
func (cancellationToken *CancellationToken) IsCancelled() bool {
	loanedCancellationToken := C.z_cancellation_token_loan(cancellationToken.cancellationToken)
	return bool(C.z_cancellation_token_is_cancelled(loanedCancellationToken))
}

func (cancellationToken CancellationToken) toC(pinner *runtime.Pinner) *C.z_owned_cancellation_token_t {
	pinner.Pin(cancellationToken.cancellationToken)
	return cancellationToken.cancellationToken
}
